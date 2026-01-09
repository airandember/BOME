package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"bome-backend/internal/database"

	"github.com/robfig/cron/v3"
)

// SearchIndexScheduler manages automated search index generation
type SearchIndexScheduler struct {
	cron         *cron.Cron
	db           *database.DB
	bunnyService *BunnyService
	isRunning    bool
	lastRun      time.Time
	nextRun      time.Time
}

// NewSearchIndexScheduler creates a new search index scheduler
func NewSearchIndexScheduler(db *database.DB, bunnyService *BunnyService) *SearchIndexScheduler {
	return &SearchIndexScheduler{
		cron:         cron.New(cron.WithSeconds()),
		db:           db,
		bunnyService: bunnyService,
		isRunning:    false,
	}
}

// Start begins the scheduled search index generation
func (s *SearchIndexScheduler) Start() error {
	if s.isRunning {
		return fmt.Errorf("scheduler is already running")
	}

	log.Println("🌙 Starting search index scheduler...")

	// Schedule for midnight every day (0 0 0 * * *)
	_, err := s.cron.AddFunc("0 0 0 * * *", func() {
		log.Println("🕛 Midnight search index generation triggered")
		s.generateSearchIndex()
	})
	if err != nil {
		return fmt.Errorf("failed to schedule search index generation: %w", err)
	}

	// Also schedule for 6 AM as backup (0 0 6 * * *)
	_, err = s.cron.AddFunc("0 0 6 * * *", func() {
		log.Println("🌅 Morning search index generation triggered")
		s.generateSearchIndex()
	})
	if err != nil {
		return fmt.Errorf("failed to schedule backup search index generation: %w", err)
	}

	s.cron.Start()
	s.isRunning = true

	// Set next run time
	entries := s.cron.Entries()
	if len(entries) > 0 {
		s.nextRun = entries[0].Next
	}

	log.Printf("✅ Search index scheduler started - next run: %s", s.nextRun.Format("2006-01-02 15:04:05"))
	return nil
}

// Stop stops the scheduler
func (s *SearchIndexScheduler) Stop() {
	if !s.isRunning {
		return
	}

	log.Println("🛑 Stopping search index scheduler...")
	s.cron.Stop()
	s.isRunning = false
	log.Println("✅ Search index scheduler stopped")
}

// UpdateConfiguration updates the scheduler configuration and restarts it
func (s *SearchIndexScheduler) UpdateConfiguration(schedule, backupSchedule string, enableBackup bool) error {
	// Stop current scheduler
	s.Stop()

	// Create new cron instance
	s.cron = cron.New(cron.WithSeconds())

	log.Printf("🔄 Updating search index scheduler configuration...")
	log.Printf("   Main schedule: %s", schedule)
	log.Printf("   Backup schedule: %s", backupSchedule)
	log.Printf("   Backup enabled: %t", enableBackup)

	// Add main schedule
	_, err := s.cron.AddFunc(schedule, func() {
		log.Println("🕛 Scheduled search index generation triggered")
		s.generateSearchIndex()
	})
	if err != nil {
		return fmt.Errorf("failed to schedule main search index generation: %w", err)
	}

	// Add backup schedule if enabled
	if enableBackup {
		_, err = s.cron.AddFunc(backupSchedule, func() {
			log.Println("🌅 Backup search index generation triggered")
			s.generateSearchIndex()
		})
		if err != nil {
			return fmt.Errorf("failed to schedule backup search index generation: %w", err)
		}
	}

	// Start the scheduler
	return s.Start()
}

// TriggerManualGeneration manually triggers search index generation
func (s *SearchIndexScheduler) TriggerManualGeneration() error {
	log.Println("🔄 Manual search index generation triggered")
	return s.generateSearchIndex()
}

// GetStatus returns the current scheduler status
func (s *SearchIndexScheduler) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"running":  s.isRunning,
		"lastRun":  nil,
		"nextRun":  nil,
		"schedule": "Daily at midnight (00:00) and 6 AM backup",
	}

	if !s.lastRun.IsZero() {
		status["lastRun"] = s.lastRun.Format("2006-01-02 15:04:05")
	}

	if !s.nextRun.IsZero() && s.isRunning {
		status["nextRun"] = s.nextRun.Format("2006-01-02 15:04:05")

		// Update next run from cron entries
		entries := s.cron.Entries()
		if len(entries) > 0 {
			status["nextRun"] = entries[0].Next.Format("2006-01-02 15:04:05")
		}
	}

	return status
}

// generateSearchIndex performs the actual search index generation
func (s *SearchIndexScheduler) generateSearchIndex() error {
	startTime := time.Now()
	log.Println("🚀 Starting search index generation...")

	defer func() {
		s.lastRun = time.Now()
		duration := s.lastRun.Sub(startTime)
		log.Printf("⏱️ Search index generation completed in %v", duration)
	}()

	// Get all videos from database
	videos, err := s.db.GetAllVideos()
	if err != nil {
		log.Printf("❌ Failed to get videos: %v", err)
		return fmt.Errorf("failed to get videos: %w", err)
	}

	log.Printf("📥 Found %d videos in database", len(videos))

	// Convert videos to optimized search format
	searchVideos := s.convertVideosToSearchFormat(videos)

	// Generate search index structure with metadata
	searchIndex := map[string]interface{}{
		"version":     "2.0",
		"generatedAt": time.Now().Format(time.RFC3339),
		"totalVideos": len(videos),
		"videos":      searchVideos,
		"metadata": map[string]interface{}{
			"generationTimeMs": 0, // Will be updated after generation
			"source":           "master_video_list",
			"indexedFields":    []string{"title", "description", "category", "tags"},
		},
	}

	// Write the search index to file
	indexPath := getSearchIndexOutputPath()
	log.Printf("📝 Writing search index to: %s", indexPath)

	// Ensure the directory exists
	dir := filepath.Dir(indexPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("❌ Failed to create directory %s: %v", dir, err)
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Update generation time in metadata
	generationTime := time.Since(startTime).Milliseconds()
	if metadata, ok := searchIndex["metadata"].(map[string]interface{}); ok {
		metadata["generationTimeMs"] = generationTime
	}

	// Marshal to COMPACT JSON for production (smaller file = faster load)
	// Use json.Marshal instead of MarshalIndent for ~30% smaller file size
	jsonData, err := json.Marshal(searchIndex)
	if err != nil {
		log.Printf("❌ Failed to marshal search index: %v", err)
		return fmt.Errorf("failed to marshal search index: %w", err)
	}

	// Write to primary file
	if err := os.WriteFile(indexPath, jsonData, 0644); err != nil {
		log.Printf("❌ Failed to write search index file: %v", err)
		return fmt.Errorf("failed to write search index file: %w", err)
	}

	// Get file size for logging
	fileInfo, _ := os.Stat(indexPath)
	fileSizeKB := int64(0)
	if fileInfo != nil {
		fileSizeKB = fileInfo.Size() / 1024
	}

	log.Printf("✅ Search index written successfully: %d videos, %d KB, %dms generation time", len(videos), fileSizeKB, generationTime)

	// Also write fallback copies to frontend static folders for redundancy
	// This ensures the frontend can always load the index even if the API is down
	writeFallbackCopies(jsonData)

	return nil
}

// writeFallbackCopies writes the search index to multiple fallback locations for redundancy
// NOTE: These fallbacks only work for LOCAL DEVELOPMENT where backend and frontend
// share the same filesystem. In PRODUCTION with separate Docker containers,
// the frontend must fetch from the backend API endpoint (/api/v1/search-index.json)
func writeFallbackCopies(jsonData []byte) {
	// Fallback paths to try (LOCAL DEVELOPMENT ONLY - frontend static folders)
	// In production, these paths won't exist because containers have separate filesystems
	fallbackPaths := []string{
		"../frontend/static/search-index.json",    // Local dev: backend and frontend are siblings
		"../../frontend/static/search-index.json", // Alternative local path
	}

	for _, path := range fallbackPaths {
		dir := filepath.Dir(path)

		// Check if directory exists
		if !dirExists(dir) {
			continue
		}

		// Try to write the fallback
		if err := os.WriteFile(path, jsonData, 0644); err != nil {
			log.Printf("⚠️ Could not write fallback to %s: %v", path, err)
		} else {
			log.Printf("✅ Fallback search index written to: %s", path)
		}
	}
}

// getSearchIndexOutputPath returns the path where the search index should be written
func getSearchIndexOutputPath() string {
	// Check for custom path from environment variable (PREFERRED METHOD)
	if customPath := os.Getenv("SEARCH_INDEX_PATH"); customPath != "" {
		log.Printf("📁 Using SEARCH_INDEX_PATH from environment: %s", customPath)
		return customPath
	}

	// Log warning that env var is not set
	log.Printf("⚠️ SEARCH_INDEX_PATH not set, trying fallback paths...")

	// Try multiple paths in order of preference
	paths := []string{
		"../frontend/static/search-index.json",    // Local development
		"../../frontend/static/search-index.json", // Alternative local path
		"/app/frontend/static/search-index.json",  // Docker/container path
		"./static/search-index.json",              // Same directory static folder
		"./search-index.json",                     // Fallback to current directory
	}

	for _, path := range paths {
		dir := filepath.Dir(path)
		if dirExists(dir) {
			log.Printf("📁 Using fallback path: %s (directory exists)", path)
			return path
		}
	}

	// Default fallback - create in current directory
	log.Printf("⚠️ No suitable directory found, using ./search-index.json")
	return "./search-index.json"
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// convertVideosToSearchFormat converts database videos to optimized search index format
// Pre-allocates slice for better performance with large video libraries
func (s *SearchIndexScheduler) convertVideosToSearchFormat(videos []database.Video) []map[string]interface{} {
	// Pre-allocate slice with exact capacity for performance
	searchVideos := make([]map[string]interface{}, 0, len(videos))

	for _, video := range videos {
		// Skip videos without bunny ID (can't be played)
		if video.BunnyVideoID == "" {
			continue
		}

		// Use thumbnail_url from database (already contains the full correct URL)
		// This preserves the actual thumbnail filename like thumbnail_fbf84c30.jpg
		finalThumbnail := video.ThumbnailURL

		// Fallback only if database thumbnail_url is empty
		if finalThumbnail == "" {
			// Try to generate from ThumbnailFileName if available
			if video.ThumbnailFileName != "" {
				finalThumbnail = s.bunnyService.GetThumbnailURLWithFilename(video.BunnyVideoID, video.ThumbnailFileName)
			} else {
				// Last resort fallback
				finalThumbnail = fmt.Sprintf("https://vz-f75053f7-465.b-cdn.net/%s/thumbnail.jpg", video.BunnyVideoID)
			}
		}

		// Build optimized search video object
		// Only include fields needed for search and display
		searchVideo := map[string]interface{}{
			// Primary search fields (indexed by Fuse.js)
			"id":          video.BunnyVideoID,
			"title":       video.Title,
			"description": video.Description,
			"category":    video.Category,
			"tags":        video.Tags,

			// Display fields
			"duration":     video.Duration,
			"createdAt":    video.CreatedAt.Format(time.RFC3339),
			"thumbnail":    finalThumbnail,
			"thumbnailUrl": finalThumbnail,
			"views":        video.ViewCount,
			"status":       video.Status,

			// Playback URLs
			"videoUrl":  s.bunnyService.GetStreamURL(video.BunnyVideoID),
			"iframeSrc": s.bunnyService.GetIframeURL(video.BunnyVideoID),

			// Bunny metadata (minimal for display)
			"bunny": map[string]interface{}{
				"guid":            video.BunnyVideoID,
				"previewImageUrl": finalThumbnail,
				"length":          video.Duration,
			},
		}

		searchVideos = append(searchVideos, searchVideo)
	}

	return searchVideos
}
