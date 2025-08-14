package database

import (
	"encoding/json"
	"fmt"
	"time"
)

// MasterVideo represents a video in the master list
type MasterVideo struct {
	ID                   int
	BunnyVideoID         string
	Title                string
	Description          string
	Category             string
	Tags                 []string
	Duration             int
	FileSize             int64
	Resolution           string
	Framerate            float64
	ThumbnailURL         string
	VideoURL             string
	IframeSrc            string
	PlaybackURL          string
	Status               string
	Views                int
	Likes                int
	IsPublic             bool
	EncodeProgress       int
	AvailableResolutions []string
	CollectionID         string
	AverageWatchTime     int
	TotalWatchTime       int64

	// Sync tracking
	LastBunnySync    time.Time
	LastMasterUpdate time.Time
	SyncStatus       string // synced, needs_attention, conflict
	SyncNotes        string

	// Metadata
	MetadataVersion int
	CreatedBy       int
	CreatedAt       time.Time
	UpdatedAt       time.Time

	// Vid Status - True or False Video is active or not
	Vid_Status bool
}

// SyncConflict represents a conflict between master list and Bunny.net
type SyncConflict struct {
	ID             int
	MasterVideoID  int
	BunnyVideoID   string
	ConflictType   string // field_mismatch, missing_field, status_mismatch
	FieldName      string
	MasterValue    string
	BunnyValue     string
	ProposedAction string // update_master, update_bunny, update_both, manual_review
	AdminNotes     string
	Resolved       bool
	ResolvedBy     *int
	ResolvedAt     *time.Time
	CreatedAt      time.Time
}

// SyncAuditLog represents an audit log entry for sync operations
type SyncAuditLog struct {
	ID            int
	MasterVideoID int
	BunnyVideoID  string
	SyncAction    string // sync_from_bunny, sync_to_bunny, conflict_resolved, manual_update
	SyncResult    string // success, failed, partial, conflict
	ChangesMade   map[string]interface{}
	ErrorMessage  string
	PerformedBy   *int
	PerformedAt   time.Time
}

// CreateMasterVideo creates a new video in the master list
func (db *DB) CreateMasterVideo(video *MasterVideo) (*MasterVideo, error) {
	tagsJSON, err := json.Marshal(video.Tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %v", err)
	}

	resolutionsJSON, err := json.Marshal(video.AvailableResolutions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resolutions: %v", err)
	}

	var id int
	err = db.QueryRow(`
		INSERT INTO master_video_list (
			bunny_video_id, title, description, category, tags, duration, file_size,
			resolution, framerate, thumbnail_url, video_url, iframe_src, playback_url,
			status, views, likes, is_public, encode_progress, available_resolutions,
			collection_id, average_watch_time, total_watch_time, sync_status,
			sync_notes, metadata_version, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26)
		RETURNING id`,
		video.BunnyVideoID, video.Title, video.Description, video.Category, string(tagsJSON),
		video.Duration, video.FileSize, video.Resolution, video.Framerate, video.ThumbnailURL,
		video.VideoURL, video.IframeSrc, video.PlaybackURL, video.Status, video.Views,
		video.Likes, video.IsPublic, video.EncodeProgress, string(resolutionsJSON),
		video.CollectionID, video.AverageWatchTime, video.TotalWatchTime, video.SyncStatus,
		video.SyncNotes, video.MetadataVersion, video.CreatedBy,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	return db.GetMasterVideoByID(id)
}

// GetMasterVideoByID retrieves a master video by ID
func (db *DB) GetMasterVideoByID(id int) (*MasterVideo, error) {
	video := &MasterVideo{}
	var tagsStr, resolutionsStr string

	err := db.QueryRow(`
		SELECT id, bunny_video_id, title, description, category, tags, duration, file_size,
		       resolution, framerate, thumbnail_url, video_url, iframe_src, playback_url,
		       status, views, likes, is_public, encode_progress, available_resolutions,
		       collection_id, average_watch_time, total_watch_time, last_bunny_sync,
		       last_master_update, sync_status, sync_notes, metadata_version, created_by,
		       created_at, updated_at, vid_status
		FROM master_video_list WHERE id = $1`,
		id,
	).Scan(
		&video.ID, &video.BunnyVideoID, &video.Title, &video.Description, &video.Category,
		&tagsStr, &video.Duration, &video.FileSize, &video.Resolution, &video.Framerate,
		&video.ThumbnailURL, &video.VideoURL, &video.IframeSrc, &video.PlaybackURL,
		&video.Status, &video.Views, &video.Likes, &video.IsPublic, &video.EncodeProgress,
		&resolutionsStr, &video.CollectionID, &video.AverageWatchTime, &video.TotalWatchTime,
		&video.LastBunnySync, &video.LastMasterUpdate, &video.SyncStatus, &video.SyncNotes,
		&video.MetadataVersion, &video.CreatedBy, &video.CreatedAt, &video.UpdatedAt, &video.Vid_Status,
	)
	if err != nil {
		return nil, err
	}

	// Parse tags from JSON
	if tagsStr != "" {
		if err := json.Unmarshal([]byte(tagsStr), &video.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %v", err)
		}
	}

	// Parse resolutions from JSON
	if resolutionsStr != "" {
		if err := json.Unmarshal([]byte(resolutionsStr), &video.AvailableResolutions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal resolutions: %v", err)
		}
	}

	return video, nil
}

// GetMasterVideoByBunnyID retrieves a master video by Bunny.net ID
func (db *DB) GetMasterVideoByBunnyID(bunnyVideoID string) (*MasterVideo, error) {
	video := &MasterVideo{}
	var tagsStr, resolutionsStr string

	err := db.QueryRow(`
		SELECT id, bunny_video_id, title, description, category, tags, duration, file_size,
		       resolution, framerate, thumbnail_url, video_url, iframe_src, playback_url,
		       status, views, likes, is_public, encode_progress, available_resolutions,
		       collection_id, average_watch_time, total_watch_time, last_bunny_sync,
		       last_master_update, sync_status, sync_notes, metadata_version, created_by,
		       created_at, updated_at, vid_status
		FROM master_video_list WHERE bunny_video_id = $1`,
		bunnyVideoID,
	).Scan(
		&video.ID, &video.BunnyVideoID, &video.Title, &video.Description, &video.Category,
		&tagsStr, &video.Duration, &video.FileSize, &video.Resolution, &video.Framerate,
		&video.ThumbnailURL, &video.VideoURL, &video.IframeSrc, &video.PlaybackURL,
		&video.Status, &video.Views, &video.Likes, &video.IsPublic, &video.EncodeProgress,
		&resolutionsStr, &video.CollectionID, &video.AverageWatchTime, &video.TotalWatchTime,
		&video.LastBunnySync, &video.LastMasterUpdate, &video.SyncStatus, &video.SyncNotes,
		&video.MetadataVersion, &video.CreatedBy, &video.CreatedAt, &video.UpdatedAt, &video.Vid_Status,
	)
	if err != nil {
		return nil, err
	}

	// Parse tags from JSON
	if tagsStr != "" {
		if err := json.Unmarshal([]byte(tagsStr), &video.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %v", err)
		}
	}

	// Parse resolutions from JSON
	if resolutionsStr != "" {
		if err := json.Unmarshal([]byte(resolutionsStr), &video.AvailableResolutions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal resolutions: %v", err)
		}
	}

	return video, nil
}

// GetMasterVideos retrieves videos from the master list with filtering and pagination
func (db *DB) GetMasterVideos(limit, offset int, category, status, syncStatus, vidStatus, sortField, sortDirection string) ([]*MasterVideo, error) {
	query := `
		SELECT id, bunny_video_id, title, description, category, tags, duration, file_size,
		       resolution, framerate, thumbnail_url, video_url, iframe_src, playback_url,
		       status, views, likes, is_public, encode_progress, available_resolutions,
		       collection_id, average_watch_time, total_watch_time, last_bunny_sync,
		       last_master_update, sync_status, sync_notes, metadata_version, created_by,
		       created_at, updated_at, vid_status
		FROM master_video_list WHERE 1=1`
	args := []interface{}{}
	argCount := 0

	if category != "" {
		argCount++
		query += fmt.Sprintf(` AND category = $%d`, argCount)
		args = append(args, category)
	}

	if status != "" {
		argCount++
		query += fmt.Sprintf(` AND status = $%d`, argCount)
		args = append(args, status)
	}

	if syncStatus != "" {
		argCount++
		query += fmt.Sprintf(` AND sync_status = $%d`, argCount)
		args = append(args, syncStatus)
	}

	if vidStatus != "" {
		argCount++
		query += fmt.Sprintf(` AND vid_status = $%d`, argCount)
		args = append(args, vidStatus == "true")
	}

	// Add sorting
	validSortFields := map[string]string{
		"id": "id", "title": "title", "category": "category", "status": "status",
		"sync_status": "sync_status", "views": "views", "created_at": "created_at",
		"duration": "duration", "file_size": "file_size", "vid_status": "vid_status",
	}

	sortColumn, valid := validSortFields[sortField]
	if !valid {
		sortColumn = "id"
	}

	if sortDirection != "asc" && sortDirection != "desc" {
		sortDirection = "desc"
	}

	argCount++
	query += fmt.Sprintf(` ORDER BY %s %s LIMIT $%d OFFSET $%d`, sortColumn, sortDirection, argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []*MasterVideo
	for rows.Next() {
		video := &MasterVideo{}
		var tagsStr, resolutionsStr string

		err := rows.Scan(
			&video.ID, &video.BunnyVideoID, &video.Title, &video.Description, &video.Category,
			&tagsStr, &video.Duration, &video.FileSize, &video.Resolution, &video.Framerate,
			&video.ThumbnailURL, &video.VideoURL, &video.IframeSrc, &video.PlaybackURL,
			&video.Status, &video.Views, &video.Likes, &video.IsPublic, &video.EncodeProgress,
			&resolutionsStr, &video.CollectionID, &video.AverageWatchTime, &video.TotalWatchTime,
			&video.LastBunnySync, &video.LastMasterUpdate, &video.SyncStatus, &video.SyncNotes,
			&video.MetadataVersion, &video.CreatedBy, &video.CreatedAt, &video.UpdatedAt, &video.Vid_Status,
		)
		if err != nil {
			return nil, err
		}

		// Parse tags from JSON
		if tagsStr != "" {
			if err := json.Unmarshal([]byte(tagsStr), &video.Tags); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tags: %v", err)
			}
		}

		// Parse resolutions from JSON
		if resolutionsStr != "" {
			if err := json.Unmarshal([]byte(resolutionsStr), &video.AvailableResolutions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal resolutions: %v", err)
			}
		}

		videos = append(videos, video)
	}

	return videos, nil
}

// UpdateMasterVideo updates a video in the master list
func (db *DB) UpdateMasterVideo(video *MasterVideo) error {
	tagsJSON, err := json.Marshal(video.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %v", err)
	}

	resolutionsJSON, err := json.Marshal(video.AvailableResolutions)
	if err != nil {
		return fmt.Errorf("failed to marshal resolutions: %v", err)
	}

	_, err = db.Exec(`
		UPDATE master_video_list SET
			title = $1, description = $2, category = $3, tags = $4, duration = $5,
			file_size = $6, resolution = $7, framerate = $8, thumbnail_url = $9,
			video_url = $10, iframe_src = $11, playback_url = $12, status = $13,
			views = $14, likes = $15, is_public = $16, encode_progress = $17,
			available_resolutions = $18, collection_id = $19, average_watch_time = $20,
			total_watch_time = $21, sync_status = $22, sync_notes = $23, vid_status = $24,
			metadata_version = metadata_version + 1, last_master_update = NOW()
		WHERE id = $25`,
		video.Title, video.Description, video.Category, string(tagsJSON), video.Duration,
		video.FileSize, video.Resolution, video.Framerate, video.ThumbnailURL,
		video.VideoURL, video.IframeSrc, video.PlaybackURL, video.Status, video.Views,
		video.Likes, video.IsPublic, video.EncodeProgress, string(resolutionsJSON),
		video.CollectionID, video.AverageWatchTime, video.TotalWatchTime, video.SyncStatus,
		video.SyncNotes, video.Vid_Status, video.ID,
	)
	return err
}

// DeleteMasterVideo deletes a video from the master list
func (db *DB) DeleteMasterVideo(id int) error {
	_, err := db.Exec(`DELETE FROM master_video_list WHERE id = $1`, id)
	return err
}

// GetSyncConflicts retrieves unresolved sync conflicts
func (db *DB) GetSyncConflicts(masterVideoID *int) ([]*SyncConflict, error) {
	query := `SELECT id, master_video_id, bunny_video_id, conflict_type, field_name,
		       master_value, bunny_value, proposed_action, admin_notes, resolved,
		       resolved_by, resolved_at, created_at
		       FROM video_sync_conflicts WHERE resolved = false`
	args := []interface{}{}

	if masterVideoID != nil {
		query += ` AND master_video_id = $1`
		args = append(args, *masterVideoID)
	}

	query += ` ORDER BY created_at DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conflicts []*SyncConflict
	for rows.Next() {
		conflict := &SyncConflict{}
		err := rows.Scan(
			&conflict.ID, &conflict.MasterVideoID, &conflict.BunnyVideoID, &conflict.ConflictType,
			&conflict.FieldName, &conflict.MasterValue, &conflict.BunnyValue, &conflict.ProposedAction,
			&conflict.AdminNotes, &conflict.Resolved, &conflict.ResolvedBy, &conflict.ResolvedAt,
			&conflict.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		conflicts = append(conflicts, conflict)
	}

	return conflicts, nil
}

// ResolveSyncConflict marks a conflict as resolved
func (db *DB) ResolveSyncConflict(conflictID int, resolvedBy int, adminNotes string) error {
	_, err := db.Exec(`
		UPDATE video_sync_conflicts SET
			resolved = true, resolved_by = $1, resolved_at = NOW(), admin_notes = $2
		WHERE id = $3`,
		resolvedBy, adminNotes, conflictID,
	)
	return err
}

// LogSyncAudit logs a sync operation
func (db *DB) LogSyncAudit(audit *SyncAuditLog) error {
	changesJSON, err := json.Marshal(audit.ChangesMade)
	if err != nil {
		return fmt.Errorf("failed to marshal changes: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO video_sync_audit_log (
			master_video_id, bunny_video_id, sync_action, sync_result,
			changes_made, error_message, performed_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		audit.MasterVideoID, audit.BunnyVideoID, audit.SyncAction, audit.SyncResult,
		string(changesJSON), audit.ErrorMessage, audit.PerformedBy,
	)
	return err
}

// GetMasterVideoStats returns statistics about the master video list
func (db *DB) GetMasterVideoStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total videos
	var totalVideos int
	err := db.QueryRow(`SELECT COUNT(*) FROM master_video_list`).Scan(&totalVideos)
	if err != nil {
		return nil, err
	}
	stats["total_videos"] = totalVideos

	// Videos by status
	statusRows, err := db.Query(`
		SELECT status, COUNT(*) FROM master_video_list 
		GROUP BY status ORDER BY COUNT(*) DESC`)
	if err != nil {
		return nil, err
	}
	defer statusRows.Close()

	statusCounts := make(map[string]int)
	for statusRows.Next() {
		var status string
		var count int
		if err := statusRows.Scan(&status, &count); err != nil {
			return nil, err
		}
		statusCounts[status] = count
	}
	stats["videos_by_status"] = statusCounts

	// Videos by sync status
	syncRows, err := db.Query(`
		SELECT sync_status, COUNT(*) FROM master_video_list 
		GROUP BY sync_status ORDER BY COUNT(*) DESC`)
	if err != nil {
		return nil, err
	}
	defer syncRows.Close()

	syncCounts := make(map[string]int)
	for syncRows.Next() {
		var syncStatus string
		var count int
		if err := syncRows.Scan(&syncStatus, &count); err != nil {
			return nil, err
		}
		syncCounts[syncStatus] = count
	}
	stats["videos_by_sync_status"] = syncCounts

	// Pending conflicts
	var pendingConflicts int
	err = db.QueryRow(`SELECT COUNT(*) FROM video_sync_conflicts WHERE resolved = false`).Scan(&pendingConflicts)
	if err != nil {
		return nil, err
	}
	stats["pending_conflicts"] = pendingConflicts

	// Total views
	var totalViews int
	err = db.QueryRow(`SELECT COALESCE(SUM(views), 0) FROM master_video_list`).Scan(&totalViews)
	if err != nil {
		return nil, err
	}
	stats["total_views"] = totalViews

	// Total duration
	var totalDuration int
	err = db.QueryRow(`SELECT COALESCE(SUM(duration), 0) FROM master_video_list`).Scan(&totalDuration)
	if err != nil {
		return nil, err
	}
	stats["total_duration"] = totalDuration

	// Total file size
	var totalFileSize int64
	err = db.QueryRow(`SELECT COALESCE(SUM(file_size), 0) FROM master_video_list`).Scan(&totalFileSize)
	if err != nil {
		return nil, err
	}
	stats["total_file_size"] = totalFileSize

	return stats, nil
}

// SearchMasterVideos searches videos in the master list
func (db *DB) SearchMasterVideos(query string, limit, offset int, sortField, sortDirection string) ([]*MasterVideo, error) {
	searchQuery := `
		SELECT id, bunny_video_id, title, description, category, tags, duration, file_size,
		       resolution, framerate, thumbnail_url, video_url, iframe_src, playback_url,
		       status, views, likes, is_public, encode_progress, available_resolutions,
		       collection_id, average_watch_time, total_watch_time, last_bunny_sync,
		       last_master_update, sync_status, sync_notes, metadata_version, created_by,
		       created_at, updated_at, vid_status
		FROM master_video_list 
		WHERE title ILIKE $1 OR description ILIKE $1 OR category ILIKE $1`

	// Add sorting
	validSortFields := map[string]string{
		"id": "id", "title": "title", "category": "category", "status": "status",
		"sync_status": "sync_status", "views": "views", "created_at": "created_at",
		"duration": "duration", "file_size": "file_size", "vid_status": "vid_status",
	}

	sortColumn, valid := validSortFields[sortField]
	if !valid {
		sortColumn = "id"
	}

	if sortDirection != "asc" && sortDirection != "desc" {
		sortDirection = "desc"
	}

	searchQuery += fmt.Sprintf(` ORDER BY %s %s LIMIT $2 OFFSET $3`, sortColumn, sortDirection)

	searchTerm := "%" + query + "%"
	rows, err := db.Query(searchQuery, searchTerm, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []*MasterVideo
	for rows.Next() {
		video := &MasterVideo{}
		var tagsStr, resolutionsStr string

		err := rows.Scan(
			&video.ID, &video.BunnyVideoID, &video.Title, &video.Description, &video.Category,
			&tagsStr, &video.Duration, &video.FileSize, &video.Resolution, &video.Framerate,
			&video.ThumbnailURL, &video.VideoURL, &video.IframeSrc, &video.PlaybackURL,
			&video.Status, &video.Views, &video.Likes, &video.IsPublic, &video.EncodeProgress,
			&resolutionsStr, &video.CollectionID, &video.AverageWatchTime, &video.TotalWatchTime,
			&video.LastBunnySync, &video.LastMasterUpdate, &video.SyncStatus, &video.SyncNotes,
			&video.MetadataVersion, &video.CreatedBy, &video.CreatedAt, &video.UpdatedAt, &video.Vid_Status,
		)
		if err != nil {
			return nil, err
		}

		// Parse tags from JSON
		if tagsStr != "" {
			if err := json.Unmarshal([]byte(tagsStr), &video.Tags); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tags: %v", err)
			}
		}

		// Parse resolutions from JSON
		if resolutionsStr != "" {
			if err := json.Unmarshal([]byte(resolutionsStr), &video.AvailableResolutions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal resolutions: %v", err)
			}
		}

		videos = append(videos, video)
	}

	return videos, nil
}

// GetMasterVideoCount returns the total count of master videos with optional filtering
func (db *DB) GetMasterVideoCount(category, status, syncStatus, vidStatus string) (int, error) {
	query := `SELECT COUNT(*) FROM master_video_list WHERE 1=1`
	args := []interface{}{}
	argCount := 0

	if category != "" {
		argCount++
		query += fmt.Sprintf(` AND category = $%d`, argCount)
		args = append(args, category)
	}

	if status != "" {
		argCount++
		query += fmt.Sprintf(` AND status = $%d`, argCount)
		args = append(args, status)
	}

	if syncStatus != "" {
		argCount++
		query += fmt.Sprintf(` AND sync_status = $%d`, argCount)
		args = append(args, syncStatus)
	}

	if vidStatus != "" {
		argCount++
		query += fmt.Sprintf(` AND vid_status = $%d`, argCount)
		args = append(args, vidStatus == "true")
	}

	var count int
	err := db.QueryRow(query, args...).Scan(&count)
	return count, err
}

// GetMasterVideoSearchCount returns the total count of master videos matching a search query
func (db *DB) GetMasterVideoSearchCount(query string) (int, error) {
	searchQuery := `
		SELECT COUNT(*) FROM master_video_list 
		WHERE title ILIKE $1 OR description ILIKE $1 OR category ILIKE $1`

	searchTerm := "%" + query + "%"
	var count int
	err := db.QueryRow(searchQuery, searchTerm).Scan(&count)
	return count, err
}
