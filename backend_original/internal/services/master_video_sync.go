package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"bome-backend/internal/database"
)

// MasterVideoSyncService handles synchronization between master list and Bunny.net
type MasterVideoSyncService struct {
	db           *database.DB
	bunnyService *BunnyService
}

// NewMasterVideoSyncService creates a new sync service
func NewMasterVideoSyncService(db *database.DB, bunnyService *BunnyService) *MasterVideoSyncService {
	return &MasterVideoSyncService{
		db:           db,
		bunnyService: bunnyService,
	}
}

// SyncFromBunny syncs all videos from Bunny.net to the master list
func (s *MasterVideoSyncService) SyncFromBunny(userID int) (*SyncResult, error) {
	log.Println("Starting sync from Bunny.net to master list...")

	result := &SyncResult{
		StartedAt:    time.Now(),
		TotalVideos:  0,
		Synced:       0,
		Updated:      0,
		Conflicts:    0,
		Errors:       0,
		ErrorDetails: []string{},
	}

	// Fetch all videos from Bunny.net
	bunnyVideos, err := s.bunnyService.FetchAllVideos()
	if err != nil {
		result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("Failed to fetch Bunny videos: %v", err))
		return result, err
	}

	result.TotalVideos = len(bunnyVideos)

	// Process each video
	for i, bunnyVideo := range bunnyVideos {
		log.Printf("🔄 [Sync %d/%d] Processing video: %s (%s)", i+1, len(bunnyVideos), bunnyVideo.GUID, bunnyVideo.Title)
		err := s.processBunnyVideo(bunnyVideo, result, userID)
		if err != nil {
			result.Errors++
			errorMsg := fmt.Sprintf("Video %s (%s): %v", bunnyVideo.GUID, bunnyVideo.Title, err)
			result.ErrorDetails = append(result.ErrorDetails, errorMsg)
			log.Printf("❌ [Sync Error %d/%d] %s", result.Errors, len(bunnyVideos), errorMsg)
		} else {
			log.Printf("✅ [Sync Success %d/%d] Video %s processed successfully", i+1, len(bunnyVideos), bunnyVideo.GUID)
		}

		// Log progress every 100 videos
		if (i+1)%100 == 0 {
			log.Printf("📊 [Sync Progress] %d/%d videos processed - %d synced, %d updated, %d errors",
				i+1, len(bunnyVideos), result.Synced, result.Updated, result.Errors)
		}
	}

	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)

	log.Printf("Sync completed: %d total, %d synced, %d updated, %d conflicts, %d errors",
		result.TotalVideos, result.Synced, result.Updated, result.Conflicts, result.Errors)

	return result, nil
}

// SyncToBunny syncs master list changes to Bunny.net
func (s *MasterVideoSyncService) SyncToBunny() (*SyncResult, error) {
	log.Println("Starting sync from master list to Bunny.net...")

	result := &SyncResult{
		StartedAt:    time.Now(),
		TotalVideos:  0,
		Synced:       0,
		Updated:      0,
		Conflicts:    0,
		Errors:       0,
		ErrorDetails: []string{},
	}

	// Get all videos from master list
	videos, err := s.db.GetMasterVideos(1000, 0, "", "", "", "", "id", "desc")
	if err != nil {
		result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("Failed to fetch master videos: %v", err))
		return result, err
	}

	result.TotalVideos = len(videos)

	// Process each video
	for _, video := range videos {
		err := s.processMasterVideo(video, result)
		if err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("Video %s: %v", video.BunnyVideoID, err))
		}
	}

	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)

	log.Printf("Sync to Bunny completed: %d total, %d synced, %d updated, %d conflicts, %d errors",
		result.TotalVideos, result.Synced, result.Updated, result.Conflicts, result.Errors)

	return result, nil
}

// CheckConflicts checks for conflicts between master list and Bunny.net
func (s *MasterVideoSyncService) CheckConflicts() (*ConflictCheckResult, error) {
	log.Println("Checking for conflicts between master list and Bunny.net...")

	result := &ConflictCheckResult{
		StartedAt:     time.Now(),
		TotalVideos:   0,
		ConflictCount: 0,
		Errors:        0,
		ErrorDetails:  []string{},
		Conflicts:     []*VideoConflict{},
	}

	// Get all videos from master list
	videos, err := s.db.GetMasterVideos(1000, 0, "", "", "", "", "id", "desc")
	if err != nil {
		result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("Failed to fetch master videos: %v", err))
		return result, err
	}

	result.TotalVideos = len(videos)

	// Check each video for conflicts
	for _, video := range videos {
		conflicts, err := s.checkVideoConflicts(video)
		if err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("Video %s: %v", video.BunnyVideoID, err))
			continue
		}

		if len(conflicts) > 0 {
			result.Conflicts = append(result.Conflicts, &VideoConflict{
				Video:     video,
				Conflicts: conflicts,
			})
			result.ConflictCount += len(conflicts)
		}
	}

	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)

	log.Printf("Conflict check completed: %d total, %d conflicts found, %d errors",
		result.TotalVideos, result.ConflictCount, result.Errors)

	return result, nil
}

// ResolveConflict resolves a specific conflict
func (s *MasterVideoSyncService) ResolveConflict(conflictID int, resolution *ConflictResolution, userID int) error {
	// Get the conflict
	conflicts, err := s.db.GetSyncConflicts(nil)
	if err != nil {
		return fmt.Errorf("failed to get conflicts: %v", err)
	}

	var targetConflict *database.SyncConflict
	for _, conflict := range conflicts {
		if conflict.ID == conflictID {
			targetConflict = conflict
			break
		}
	}

	if targetConflict == nil {
		return fmt.Errorf("conflict not found: %d", conflictID)
	}

	// Get the master video
	masterVideo, err := s.db.GetMasterVideoByID(targetConflict.MasterVideoID)
	if err != nil {
		return fmt.Errorf("failed to get master video: %v", err)
	}

	// Apply the resolution
	switch resolution.Action {
	case "update_master":
		err = s.updateMasterFromResolution(masterVideo, resolution)
	case "update_bunny":
		err = s.updateBunnyFromResolution(masterVideo, resolution)
	case "update_both":
		err = s.updateBothFromResolution(masterVideo, resolution)
	default:
		return fmt.Errorf("invalid resolution action: %s", resolution.Action)
	}

	if err != nil {
		return fmt.Errorf("failed to apply resolution: %v", err)
	}

	// Mark conflict as resolved
	err = s.db.ResolveSyncConflict(conflictID, userID, resolution.Notes)
	if err != nil {
		return fmt.Errorf("failed to mark conflict as resolved: %v", err)
	}

	// Log the resolution
	audit := &database.SyncAuditLog{
		MasterVideoID: masterVideo.ID,
		BunnyVideoID:  masterVideo.BunnyVideoID,
		SyncAction:    "conflict_resolved",
		SyncResult:    "success",
		ChangesMade: map[string]interface{}{
			"resolution_action": resolution.Action,
			"field":             resolution.Field,
			"old_value":         resolution.OldValue,
			"new_value":         resolution.NewValue,
			"notes":             resolution.Notes,
		},
		PerformedBy: &userID,
	}

	return s.db.LogSyncAudit(audit)
}

// Helper methods

// fetchBunnyVideos fetches videos from Bunny.net API
func fetchBunnyVideos(libraryID, apiKey string, page int, itemsPerPage int, search string) ([]BunnyVideo, int, error) {
	url := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos?page=%d&itemsPerPage=%d", libraryID, page, itemsPerPage)
	if search != "" {
		url += fmt.Sprintf("&search=%s", search)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("AccessKey", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("Bunny API returned status %d", resp.StatusCode)
	}

	var response struct {
		TotalItems   int          `json:"totalItems"`
		CurrentPage  int          `json:"currentPage"`
		ItemsPerPage int          `json:"itemsPerPage"`
		Items        []BunnyVideo `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, 0, err
	}

	return response.Items, response.TotalItems, nil
}

func (s *MasterVideoSyncService) processBunnyVideo(bunnyVideo BunnyVideo, result *SyncResult, userID int) error {
	// Check if video exists in master list
	_, err := s.db.GetMasterVideoByBunnyID(bunnyVideo.GUID)
	if err != nil {
		// Video doesn't exist, create it
		return s.createMasterVideoFromBunny(bunnyVideo, result, userID)
	}

	// Video exists, skip it (or update if needed)
	log.Printf("🔄 Video %s already exists, skipping", bunnyVideo.GUID)
	return nil
}

func (s *MasterVideoSyncService) createMasterVideoFromBunny(bunnyVideo BunnyVideo, result *SyncResult, userID int) error {
	log.Printf("🔄 [CreateMasterVideo] Starting creation for video: %s (%s)", bunnyVideo.GUID, bunnyVideo.Title)

	// Parse available resolutions from string
	var availableResolutions []string
	if bunnyVideo.AvailableResolutions != "" {
		availableResolutions = strings.Split(bunnyVideo.AvailableResolutions, ",")
		log.Printf("🔄 [CreateMasterVideo] Parsed %d resolutions: %v", len(availableResolutions), availableResolutions)
	}

	// Create master video in the master_video_list table
	masterVideo := &database.MasterVideo{
		BunnyVideoID:         bunnyVideo.GUID,
		Title:                s.cleanTitleFromFileExtension(bunnyVideo.Title),
		Description:          s.getBunnyDescription(bunnyVideo),
		Category:             bunnyVideo.Category,
		Tags:                 s.extractTagsFromBunny(bunnyVideo),
		TagIDs:               []int{}, // Default empty
		Tagged:               false,   // Default not tagged
		Duration:             bunnyVideo.Length,
		FileSize:             bunnyVideo.StorageSize,
		Resolution:           fmt.Sprintf("%dx%d", bunnyVideo.Width, bunnyVideo.Height),
		Framerate:            bunnyVideo.Framerate,
		ThumbnailURL:         s.bunnyService.GetThumbnailURLWithFilename(bunnyVideo.GUID, bunnyVideo.ThumbnailFileName),
		VideoURL:             s.bunnyService.GetStreamURL(bunnyVideo.GUID),
		IframeSrc:            s.bunnyService.GetIframeURL(bunnyVideo.GUID),
		PlaybackURL:          s.bunnyService.GetStreamURL(bunnyVideo.GUID),
		Status:               s.mapBunnyStatus(bunnyVideo.Status),
		Views:                bunnyVideo.Views,
		Likes:                0, // Default
		IsPublic:             bunnyVideo.IsPublic,
		EncodeProgress:       bunnyVideo.EncodeProgress,
		AvailableResolutions: availableResolutions,
		CollectionID:         bunnyVideo.CollectionID,
		AverageWatchTime:     bunnyVideo.AverageWatchTime,
		TotalWatchTime:       bunnyVideo.TotalWatchTime,
		SyncStatus:           "synced",
		SyncNotes:            "",
		MetadataVersion:      1,
		CreatedBy:            &userID, // Use authenticated user ID
		Vid_Status:           true,    // Default active
	}

	createdVideo, err := s.db.CreateMasterVideo(masterVideo)
	if err != nil {
		log.Printf("❌ [CreateMasterVideo] Database creation failed for %s: %v", bunnyVideo.GUID, err)
		return fmt.Errorf("failed to create master video: %v", err)
	}

	log.Printf("✅ [CreateMasterVideo] Successfully created video %s with ID %d", bunnyVideo.GUID, createdVideo.ID)
	result.Synced++
	return nil
}

func (s *MasterVideoSyncService) updateMasterVideoFromBunny(masterVideo *database.MasterVideo, bunnyVideo BunnyVideo, result *SyncResult) error {
	// Check for conflicts
	conflicts := s.detectConflicts(masterVideo, bunnyVideo)

	if len(conflicts) > 0 {
		// Log conflicts
		for _, conflict := range conflicts {
			s.logConflict(masterVideo.ID, bunnyVideo.GUID, conflict)
		}

		// Update sync status
		masterVideo.SyncStatus = "needs_attention"
		masterVideo.SyncNotes = fmt.Sprintf("Found %d conflicts with Bunny.net", len(conflicts))
		result.Conflicts++
	} else {
		// No conflicts, update fields that should be synced from Bunny
		masterVideo.Status = s.mapBunnyStatus(bunnyVideo.Status)
		masterVideo.Views = bunnyVideo.Views
		masterVideo.EncodeProgress = bunnyVideo.EncodeProgress
		masterVideo.AverageWatchTime = bunnyVideo.AverageWatchTime
		masterVideo.TotalWatchTime = int64(bunnyVideo.TotalWatchTime)
		masterVideo.SyncStatus = "synced"
		masterVideo.SyncNotes = ""
		result.Updated++
	}

	masterVideo.LastBunnySync = time.Now()

	return s.db.UpdateMasterVideo(masterVideo)
}

func (s *MasterVideoSyncService) processMasterVideo(masterVideo *database.MasterVideo, result *SyncResult) error {
	// Get current Bunny video
	bunnyVideo, err := s.bunnyService.GetVideo(masterVideo.BunnyVideoID)
	if err != nil {
		return fmt.Errorf("failed to get Bunny video: %v", err)
	}

	// Check what needs to be updated in Bunny
	updates := s.determineBunnyUpdates(masterVideo, *bunnyVideo)

	if len(updates) > 0 {
		// Apply updates to Bunny (this would require Bunny API calls)
		// For now, we'll just log what would be updated
		log.Printf("Would update Bunny video %s with: %v", masterVideo.BunnyVideoID, updates)
		result.Updated++
	} else {
		result.Synced++
	}

	// Update sync status
	masterVideo.SyncStatus = "synced"
	masterVideo.SyncNotes = ""
	masterVideo.LastMasterUpdate = time.Now()

	return s.db.UpdateMasterVideo(masterVideo)
}

// cleanTitleFromFileExtension removes common video file extensions from titles
// This handles the pattern where Bunny.net titles include extensions like .mp4, .mov, .wmv, etc.
func (s *MasterVideoSyncService) cleanTitleFromFileExtension(title string) string {
	if title == "" {
		return title
	}

	// Common video file extensions to remove
	extensions := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm", ".mkv", ".mp3"}

	// Check if title ends with any of these extensions
	for _, ext := range extensions {
		if len(title) > len(ext) && strings.HasSuffix(strings.ToLower(title), ext) {
			return title[:len(title)-len(ext)]
		}
	}

	return title
}

func (s *MasterVideoSyncService) checkVideoConflicts(masterVideo *database.MasterVideo) ([]*FieldConflict, error) {
	// Get current Bunny video
	bunnyVideo, err := s.bunnyService.GetVideo(masterVideo.BunnyVideoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Bunny video: %v", err)
	}

	return s.detectConflicts(masterVideo, *bunnyVideo), nil
}

func (s *MasterVideoSyncService) detectConflicts(masterVideo *database.MasterVideo, bunnyVideo BunnyVideo) []*FieldConflict {
	var conflicts []*FieldConflict

	// Check title (clean file extensions from Bunny title for comparison)
	cleanBunnyTitle := s.cleanTitleFromFileExtension(bunnyVideo.Title)
	if masterVideo.Title != cleanBunnyTitle {
		conflicts = append(conflicts, &FieldConflict{
			Field:        "title",
			MasterValue:  masterVideo.Title,
			BunnyValue:   cleanBunnyTitle, // Use cleaned title for conflict reporting
			ConflictType: "field_mismatch",
		})
	}

	// Check description
	bunnyDescription := s.getBunnyDescription(bunnyVideo)
	if masterVideo.Description != bunnyDescription {
		conflicts = append(conflicts, &FieldConflict{
			Field:        "description",
			MasterValue:  masterVideo.Description,
			BunnyValue:   bunnyDescription,
			ConflictType: "field_mismatch",
		})
	}

	// Check category
	if masterVideo.Category != bunnyVideo.Category {
		conflicts = append(conflicts, &FieldConflict{
			Field:        "category",
			MasterValue:  masterVideo.Category,
			BunnyValue:   bunnyVideo.Category,
			ConflictType: "field_mismatch",
		})
	}

	// Check status
	bunnyStatus := s.mapBunnyStatus(bunnyVideo.Status)
	if masterVideo.Status != bunnyStatus {
		conflicts = append(conflicts, &FieldConflict{
			Field:        "status",
			MasterValue:  masterVideo.Status,
			BunnyValue:   bunnyStatus,
			ConflictType: "status_mismatch",
		})
	}

	return conflicts
}

func (s *MasterVideoSyncService) logConflict(masterVideoID int, bunnyVideoID string, conflict *FieldConflict) {
	// This would use the database function to log conflicts
	// For now, we'll just log it
	log.Printf("Conflict detected for video %s: %s field mismatch (Master: %s, Bunny: %s)",
		bunnyVideoID, conflict.Field, conflict.MasterValue, conflict.BunnyValue)
}

func (s *MasterVideoSyncService) determineBunnyUpdates(masterVideo *database.MasterVideo, bunnyVideo BunnyVideo) map[string]interface{} {
	updates := make(map[string]interface{})

	// Check if master has newer metadata that should be pushed to Bunny
	cleanBunnyTitle := s.cleanTitleFromFileExtension(bunnyVideo.Title)
	if masterVideo.Title != cleanBunnyTitle {
		updates["title"] = masterVideo.Title
	}

	bunnyDescription := s.getBunnyDescription(bunnyVideo)
	if masterVideo.Description != bunnyDescription {
		updates["description"] = masterVideo.Description
	}

	if masterVideo.Category != bunnyVideo.Category {
		updates["category"] = masterVideo.Category
	}

	return updates
}

func (s *MasterVideoSyncService) updateMasterFromResolution(masterVideo *database.MasterVideo, resolution *ConflictResolution) error {
	// Update the master video field
	switch resolution.Field {
	case "title":
		masterVideo.Title = resolution.NewValue
	case "description":
		masterVideo.Description = resolution.NewValue
	case "category":
		masterVideo.Category = resolution.NewValue
	case "status":
		masterVideo.Status = resolution.NewValue
	default:
		return fmt.Errorf("unknown field: %s", resolution.Field)
	}

	masterVideo.SyncStatus = "synced"
	masterVideo.SyncNotes = "Updated from conflict resolution"

	return s.db.UpdateMasterVideo(masterVideo)
}

func (s *MasterVideoSyncService) updateBunnyFromResolution(masterVideo *database.MasterVideo, resolution *ConflictResolution) error {
	// This would make API calls to update Bunny.net
	// For now, we'll just log the action
	log.Printf("Would update Bunny video %s field %s to %s",
		masterVideo.BunnyVideoID, resolution.Field, resolution.NewValue)

	// Update master sync status
	masterVideo.SyncStatus = "synced"
	masterVideo.SyncNotes = "Updated Bunny from conflict resolution"

	return s.db.UpdateMasterVideo(masterVideo)
}

func (s *MasterVideoSyncService) updateBothFromResolution(masterVideo *database.MasterVideo, resolution *ConflictResolution) error {
	// Update master
	err := s.updateMasterFromResolution(masterVideo, resolution)
	if err != nil {
		return err
	}

	// Update Bunny
	return s.updateBunnyFromResolution(masterVideo, resolution)
}

// Helper functions

func (s *MasterVideoSyncService) getBunnyDescription(bunnyVideo BunnyVideo) string {
	if bunnyVideo.Description != nil {
		return *bunnyVideo.Description
	}
	return ""
}

func (s *MasterVideoSyncService) extractTagsFromBunny(bunnyVideo BunnyVideo) []string {
	// Since the services.BunnyVideo doesn't have MetaTags, we'll create tags based on other properties
	tags := []string{"bunny", "streaming"}

	if bunnyVideo.Title != "" {
		cleanTitle := s.cleanTitleFromFileExtension(bunnyVideo.Title)
		tags = append(tags, strings.ToLower(cleanTitle))
	}

	if bunnyVideo.Category != "" {
		tags = append(tags, strings.ToLower(bunnyVideo.Category))
	}

	if bunnyVideo.HasMP4Fallback {
		tags = append(tags, "mp4")
	}

	if bunnyVideo.IsPublic {
		tags = append(tags, "public")
	}

	log.Printf("🏷️ [ExtractTags] Video %s generated %d tags: %v", bunnyVideo.GUID, len(tags), tags)
	return tags
}

func (s *MasterVideoSyncService) mapBunnyStatus(bunnyStatus int) string {
	switch bunnyStatus {
	case 0:
		return "created"
	case 1:
		return "uploaded"
	case 2:
		return "processing"
	case 3:
		return "transcoding"
	case 4:
		return "ready"
	case 5:
		return "error"
	case 6:
		return "upload_failed"
	case 7:
		return "jit_segmenting"
	case 8:
		return "jit_playlists_created"
	default:
		return "unknown"
	}
}

// Data structures

type SyncResult struct {
	StartedAt    time.Time
	CompletedAt  time.Time
	Duration     time.Duration
	TotalVideos  int
	Synced       int
	Updated      int
	Conflicts    int
	Errors       int
	ErrorDetails []string
}

type ConflictCheckResult struct {
	StartedAt     time.Time
	CompletedAt   time.Time
	Duration      time.Duration
	TotalVideos   int
	ConflictCount int
	Errors        int
	ErrorDetails  []string
	Conflicts     []*VideoConflict
}

type VideoConflict struct {
	Video     *database.MasterVideo
	Conflicts []*FieldConflict
}

type FieldConflict struct {
	Field        string
	MasterValue  string
	BunnyValue   string
	ConflictType string
}

type ConflictResolution struct {
	Action   string // update_master, update_bunny, update_both
	Field    string
	OldValue string
	NewValue string
	Notes    string
}
