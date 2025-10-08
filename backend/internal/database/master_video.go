package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lib/pq"
)

// MasterVideo represents a video in the master list
type MasterVideo struct {
	ID                   int
	BunnyVideoID         string
	Title                string
	Description          string
	Category             string
	Tags                 []string
	TagIDs               []int `json:"tag_ids"`
	Tagged               bool
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
	CreatedBy       *int
	CreatedAt       time.Time
	UpdatedAt       time.Time

	// Vid Status - True or False Video is active or not
	Vid_Status bool
}

// ArticleExclusion represents a word that should be excluded from tag generation
type ArticleExclusion struct {
	ID        int
	Word      string
	Excluded  bool
	SubsiteID int
	CreatedAt time.Time
	UpdatedAt time.Time
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
	log.Printf("🔄 [DB CreateMasterVideo] Starting creation for video: %s", video.BunnyVideoID)
	log.Printf("🔍 [DB CreateMasterVideo] Input data - Tags: %v (len=%d), Resolutions: %v (len=%d), TagIDs: %v (len=%d)",
		video.Tags, len(video.Tags), video.AvailableResolutions, len(video.AvailableResolutions), video.TagIDs, len(video.TagIDs))

	// For JSONB columns, we need to handle empty slices specially
	// Pass NULL for empty slices to let PostgreSQL use the column default
	var tagsValue, resolutionsValue, tagIDsValue interface{}

	if len(video.Tags) == 0 {
		tagsValue = nil // Let PostgreSQL use default '[]'::jsonb
	} else {
		tagsJSON, err := json.Marshal(video.Tags)
		if err != nil {
			log.Printf("❌ [DB CreateMasterVideo] Failed to marshal tags for %s: %v", video.BunnyVideoID, err)
			return nil, fmt.Errorf("failed to marshal tags: %v", err)
		}
		tagsValue = string(tagsJSON)
	}

	if len(video.AvailableResolutions) == 0 {
		resolutionsValue = nil // Let PostgreSQL use default '[]'::jsonb
	} else {
		resolutionsJSON, err := json.Marshal(video.AvailableResolutions)
		if err != nil {
			log.Printf("❌ [DB CreateMasterVideo] Failed to marshal resolutions for %s: %v", video.BunnyVideoID, err)
			return nil, fmt.Errorf("failed to marshal resolutions: %v", err)
		}
		resolutionsValue = string(resolutionsJSON)
	}

	if len(video.TagIDs) == 0 {
		tagIDsValue = nil // Let PostgreSQL use default '[]'::jsonb
	} else {
		tagIDsJSON, err := json.Marshal(video.TagIDs)
		if err != nil {
			log.Printf("❌ [DB CreateMasterVideo] Failed to marshal tag IDs for %s: %v", video.BunnyVideoID, err)
			return nil, fmt.Errorf("failed to marshal tag IDs: %v", err)
		}
		tagIDsValue = string(tagIDsJSON)
	}

	log.Printf("🔄 [DB CreateMasterVideo] Using NULL for empty slices - Tags: %v, Resolutions: %v, TagIDs: %v",
		tagsValue, resolutionsValue, tagIDsValue)

	var id int
	err := db.QueryRow(`
		INSERT INTO master_video_list (
			bunny_video_id, title, description, category, tags, duration, file_size,
			resolution, framerate, thumbnail_url, video_url, iframe_src, playback_url,
			status, views, likes, is_public, encode_progress, available_resolutions,
			collection_id, average_watch_time, total_watch_time, vid_status, tagged, tag_ids, sync_status,
			sync_notes, metadata_version, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29)
		RETURNING id`,
		video.BunnyVideoID, video.Title, video.Description, video.Category, nil,
		video.Duration, video.FileSize, video.Resolution, video.Framerate, video.ThumbnailURL,
		video.VideoURL, video.IframeSrc, video.PlaybackURL, video.Status, video.Views,
		video.Likes, video.IsPublic, video.EncodeProgress, resolutionsValue,
		video.CollectionID, video.AverageWatchTime, video.TotalWatchTime, video.Vid_Status, video.Tagged, tagIDsValue, video.SyncStatus,
		video.SyncNotes, video.MetadataVersion, video.CreatedBy,
	).Scan(&id)
	if err != nil {
		log.Printf("❌ [DB CreateMasterVideo] INSERT query failed for %s: %v", video.BunnyVideoID, err)
		log.Printf("🔍 [DB CreateMasterVideo] Query values - Title: %s, Category: %s, Status: %s, Vid_Status: %t, Tagged: %t",
			video.Title, video.Category, video.Status, video.Vid_Status, video.Tagged)
		log.Printf("🔍 [DB CreateMasterVideo] Interface values - Tags: %v, Resolutions: %v, TagIDs: %v",
			tagsValue, resolutionsValue, tagIDsValue)
		return nil, err
	}

	log.Printf("✅ [DB CreateMasterVideo] INSERT successful for %s, got ID: %d", video.BunnyVideoID, id)

	result, err := db.GetMasterVideoByID(id)
	if err != nil {
		log.Printf("❌ [DB CreateMasterVideo] Failed to retrieve created video %s (ID: %d): %v", video.BunnyVideoID, id, err)
		return nil, err
	}

	log.Printf("✅ [DB CreateMasterVideo] Successfully created and retrieved video %s (ID: %d)", video.BunnyVideoID, id)
	return result, nil
}

// GetMasterVideoByID retrieves a master video by ID
func (db *DB) GetMasterVideoByID(id int) (*MasterVideo, error) {
	video := &MasterVideo{}
	var tagsStr, resolutionsStr, tagIDsStr sql.NullString
	var createdBy sql.NullInt64

	err := db.QueryRow(`
		SELECT id, bunny_video_id, title, description, category, tags, tagged, duration, file_size,
		       resolution, framerate, thumbnail_url, video_url, iframe_src, playback_url,
		       status, views, likes, is_public, encode_progress, available_resolutions,
		       collection_id, average_watch_time, total_watch_time, last_bunny_sync,
		       last_master_update, sync_status, sync_notes, metadata_version, created_by,
		       created_at, updated_at, vid_status, tag_ids
		FROM master_video_list WHERE id = $1`,
		id,
	).Scan(
		&video.ID, &video.BunnyVideoID, &video.Title, &video.Description, &video.Category,
		&tagsStr, &video.Tagged, &video.Duration, &video.FileSize, &video.Resolution, &video.Framerate,
		&video.ThumbnailURL, &video.VideoURL, &video.IframeSrc, &video.PlaybackURL,
		&video.Status, &video.Views, &video.Likes, &video.IsPublic, &video.EncodeProgress,
		&resolutionsStr, &video.CollectionID, &video.AverageWatchTime, &video.TotalWatchTime,
		&video.LastBunnySync, &video.LastMasterUpdate, &video.SyncStatus, &video.SyncNotes,
		&video.MetadataVersion, &createdBy, &video.CreatedAt, &video.UpdatedAt, &video.Vid_Status, &tagIDsStr,
	)
	if err != nil {
		return nil, err
	}

	// Convert sql.NullInt64 to *int
	if createdBy.Valid {
		createdByInt := int(createdBy.Int64)
		video.CreatedBy = &createdByInt
	} else {
		video.CreatedBy = nil
	}

	// Parse tags from JSON (handle NULL)
	if tagsStr.Valid && tagsStr.String != "" {
		if err := json.Unmarshal([]byte(tagsStr.String), &video.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %v", err)
		}
	}

	// Parse resolutions from JSON (handle NULL)
	if resolutionsStr.Valid && resolutionsStr.String != "" {
		if err := json.Unmarshal([]byte(resolutionsStr.String), &video.AvailableResolutions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal resolutions: %v", err)
		}
	}

	// Parse tag IDs from JSON (handle NULL)
	if tagIDsStr.Valid && tagIDsStr.String != "" {
		if err := json.Unmarshal([]byte(tagIDsStr.String), &video.TagIDs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tag IDs: %v", err)
		}
	}

	return video, nil
}

// GetMasterVideoByBunnyID retrieves a master video by Bunny.net ID
func (db *DB) GetMasterVideoByBunnyID(bunnyVideoID string) (*MasterVideo, error) {
	video := &MasterVideo{}
	var tagsStr, resolutionsStr, tagIDsStr sql.NullString
	var createdBy sql.NullInt64

	err := db.QueryRow(`
		SELECT id, bunny_video_id, title, description, category, tags, tagged, duration, file_size,
		       resolution, framerate, thumbnail_url, video_url, iframe_src, playback_url,
		       status, views, likes, is_public, encode_progress, available_resolutions,
		       collection_id, average_watch_time, total_watch_time, last_bunny_sync,
		       last_master_update, sync_status, sync_notes, metadata_version, created_by,
		       created_at, updated_at, vid_status, tag_ids
		FROM master_video_list WHERE bunny_video_id = $1`,
		bunnyVideoID,
	).Scan(
		&video.ID, &video.BunnyVideoID, &video.Title, &video.Description, &video.Category,
		&tagsStr, &video.Tagged, &video.Duration, &video.FileSize, &video.Resolution, &video.Framerate,
		&video.ThumbnailURL, &video.VideoURL, &video.IframeSrc, &video.PlaybackURL,
		&video.Status, &video.Views, &video.Likes, &video.IsPublic, &video.EncodeProgress,
		&resolutionsStr, &video.CollectionID, &video.AverageWatchTime, &video.TotalWatchTime,
		&video.LastBunnySync, &video.LastMasterUpdate, &video.SyncStatus, &video.SyncNotes,
		&video.MetadataVersion, &createdBy, &video.CreatedAt, &video.UpdatedAt, &video.Vid_Status, &tagIDsStr,
	)
	if err != nil {
		return nil, err
	}

	// Convert sql.NullInt64 to *int
	if createdBy.Valid {
		createdByInt := int(createdBy.Int64)
		video.CreatedBy = &createdByInt
	} else {
		video.CreatedBy = nil
	}

	// Parse tags from JSON (handle NULL)
	if tagsStr.Valid && tagsStr.String != "" {
		if err := json.Unmarshal([]byte(tagsStr.String), &video.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %v", err)
		}
	}

	// Parse resolutions from JSON (handle NULL)
	if resolutionsStr.Valid && resolutionsStr.String != "" {
		if err := json.Unmarshal([]byte(resolutionsStr.String), &video.AvailableResolutions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal resolutions: %v", err)
		}
	}

	// Parse tag IDs from JSON (handle NULL)
	if tagIDsStr.Valid && tagIDsStr.String != "" {
		if err := json.Unmarshal([]byte(tagIDsStr.String), &video.TagIDs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tag IDs: %v", err)
		}
	}

	return video, nil
}

// GetMasterVideos retrieves videos from the master list with filtering and pagination
func (db *DB) GetMasterVideos(limit, offset int, category, status, syncStatus, vidStatus, sortField, sortDirection string) ([]*MasterVideo, error) {
	log.Printf("🎬 [DB-GetMasterVideos] Called with limit=%d, offset=%d, category='%s', status='%s', syncStatus='%s', vidStatus='%s', sortField='%s', sortDirection='%s'",
		limit, offset, category, status, syncStatus, vidStatus, sortField, sortDirection)

	query := `
		SELECT id, bunny_video_id, title, description, category, tags, tagged, duration, file_size,
		       resolution, framerate, thumbnail_url, video_url, iframe_src, playback_url,
		       status, views, likes, is_public, encode_progress, available_resolutions,
		       collection_id, average_watch_time, total_watch_time, last_bunny_sync,
		       last_master_update, sync_status, sync_notes, metadata_version, created_by,
		       created_at, updated_at, vid_status, tag_ids
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

	log.Printf("🎬 [DB-GetMasterVideos] Final query: %s", query)
	log.Printf("🎬 [DB-GetMasterVideos] Query args: %v", args)

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("❌ [DB-GetMasterVideos] Query execution failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	log.Printf("🎬 [DB-GetMasterVideos] Query executed successfully, processing rows...")

	var videos []*MasterVideo
	rowCount := 0
	for rows.Next() {
		rowCount++
		log.Printf("🎬 [DB-GetMasterVideos] Processing row %d", rowCount)
		video := &MasterVideo{}
		var tagsStr, resolutionsStr, tagIDsStr sql.NullString

		// Initialize CreatedBy as a sql.NullInt64 to handle potential NULL values
		var createdBy sql.NullInt64

		err := rows.Scan(
			&video.ID, &video.BunnyVideoID, &video.Title, &video.Description, &video.Category,
			&tagsStr, &video.Tagged, &video.Duration, &video.FileSize, &video.Resolution, &video.Framerate,
			&video.ThumbnailURL, &video.VideoURL, &video.IframeSrc, &video.PlaybackURL,
			&video.Status, &video.Views, &video.Likes, &video.IsPublic, &video.EncodeProgress,
			&resolutionsStr, &video.CollectionID, &video.AverageWatchTime, &video.TotalWatchTime,
			&video.LastBunnySync, &video.LastMasterUpdate, &video.SyncStatus, &video.SyncNotes,
			&video.MetadataVersion, &createdBy, &video.CreatedAt, &video.UpdatedAt, &video.Vid_Status, &tagIDsStr,
		)
		if err != nil {
			return nil, err
		}

		// Convert sql.NullInt64 to *int
		if createdBy.Valid {
			createdByInt := int(createdBy.Int64)
			video.CreatedBy = &createdByInt
		} else {
			video.CreatedBy = nil
		}

		// Parse tags from JSON (handle NULL)
		if tagsStr.Valid && tagsStr.String != "" {
			if err := json.Unmarshal([]byte(tagsStr.String), &video.Tags); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tags: %v", err)
			}
		}

		// Parse resolutions from JSON (handle NULL)
		if resolutionsStr.Valid && resolutionsStr.String != "" {
			if err := json.Unmarshal([]byte(resolutionsStr.String), &video.AvailableResolutions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal resolutions: %v", err)
			}
		}

		// Parse tag IDs from JSON (handle NULL)
		if tagIDsStr.Valid && tagIDsStr.String != "" {
			if err := json.Unmarshal([]byte(tagIDsStr.String), &video.TagIDs); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tag IDs: %v", err)
			}
		}

		videos = append(videos, video)
		log.Printf("🎬 [DB-GetMasterVideos] Video %d added to results", video.ID)
	}

	if err := rows.Err(); err != nil {
		log.Printf("❌ [DB-GetMasterVideos] Rows iteration error: %v", err)
		return nil, err
	}

	log.Printf("🎬 [DB-GetMasterVideos] Completed successfully - returned %d videos", len(videos))
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
		SELECT id, bunny_video_id, title, description, category, tags, tagged, duration, file_size,
		       resolution, framerate, thumbnail_url, video_url, iframe_src, playback_url,
		       status, views, likes, is_public, encode_progress, available_resolutions,
		       collection_id, average_watch_time, total_watch_time, last_bunny_sync,
		       last_master_update, sync_status, sync_notes, metadata_version, created_by,
		       created_at, updated_at, vid_status, tag_ids
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
		var tagsStr, resolutionsStr, tagIDsStr sql.NullString
		var createdBy sql.NullInt64

		err := rows.Scan(
			&video.ID, &video.BunnyVideoID, &video.Title, &video.Description, &video.Category,
			&tagsStr, &video.Tagged, &video.Duration, &video.FileSize, &video.Resolution, &video.Framerate,
			&video.ThumbnailURL, &video.VideoURL, &video.IframeSrc, &video.PlaybackURL,
			&video.Status, &video.Views, &video.Likes, &video.IsPublic, &video.EncodeProgress,
			&resolutionsStr, &video.CollectionID, &video.AverageWatchTime, &video.TotalWatchTime,
			&video.LastBunnySync, &video.LastMasterUpdate, &video.SyncStatus, &video.SyncNotes,
			&video.MetadataVersion, &createdBy, &video.CreatedAt, &video.UpdatedAt, &video.Vid_Status, &tagIDsStr,
		)
		if err != nil {
			return nil, err
		}

		// Convert sql.NullInt64 to *int
		if createdBy.Valid {
			createdByInt := int(createdBy.Int64)
			video.CreatedBy = &createdByInt
		} else {
			video.CreatedBy = nil
		}

		// Parse tags from JSON (handle NULL)
		if tagsStr.Valid && tagsStr.String != "" {
			if err := json.Unmarshal([]byte(tagsStr.String), &video.Tags); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tags: %v", err)
			}
		}

		// Parse resolutions from JSON (handle NULL)
		if resolutionsStr.Valid && resolutionsStr.String != "" {
			if err := json.Unmarshal([]byte(resolutionsStr.String), &video.AvailableResolutions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal resolutions: %v", err)
			}
		}

		// Parse tag IDs from JSON (handle NULL)
		if tagIDsStr.Valid && tagIDsStr.String != "" {
			if err := json.Unmarshal([]byte(tagIDsStr.String), &video.TagIDs); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tag IDs: %v", err)
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

// UpdateVideoTags updates the tags for a video using tag IDs and sets tagged to true
func (db *DB) UpdateVideoTags(videoID int, tags []string) error {
	// Convert tag words to tag IDs
	tagIDs, err := db.convertTagWordsToIDs(tags)
	if err != nil {
		return fmt.Errorf("failed to convert tag words to IDs: %v", err)
	}

	// Update video with tag IDs array
	_, err = db.Exec(`
		UPDATE master_video_list 
		SET tag_ids = $1, tagged = true, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2
	`, pq.Array(tagIDs), videoID)

	if err != nil {
		return fmt.Errorf("failed to update video tag IDs: %v", err)
	}

	// Update tag frequency in tags table
	return db.updateTagFrequency(tags)
}

// ReplaceVideoTags completely replaces tags for a video using tag IDs (used for "Tag All" functionality)
func (db *DB) ReplaceVideoTags(videoID int, tags []string) error {
	// Convert tag words to tag IDs
	tagIDs, err := db.convertTagWordsToIDs(tags)
	if err != nil {
		return fmt.Errorf("failed to convert tag words to IDs: %v", err)
	}

	// First, get the old tag IDs to decrement their frequency
	var oldTagIDs pq.Int64Array
	err = db.QueryRow(`SELECT COALESCE(tag_ids, '{}') FROM master_video_list WHERE id = $1`, videoID).Scan(&oldTagIDs)
	if err != nil {
		return fmt.Errorf("failed to get old tag IDs: %v", err)
	}

	// Convert old tag IDs to words and decrement their frequency
	if len(oldTagIDs) > 0 {
		oldTagWords, err := db.convertTagIDsToWords(oldTagIDs)
		if err == nil && len(oldTagWords) > 0 {
			// Decrement frequency for old tags
			if err := db.decrementTagFrequency(oldTagWords); err != nil {
				log.Printf("⚠️ Warning: Failed to decrement old tag frequencies: %v", err)
				// Continue with the update even if decrement fails
			}
		}
	}

	// Update the video with new tag IDs
	_, err = db.Exec(`
        UPDATE master_video_list 
        SET tag_ids = $1, tagged = true, updated_at = CURRENT_TIMESTAMP 
        WHERE id = $2
    `, pq.Array(tagIDs), videoID)

	if err != nil {
		return fmt.Errorf("failed to update video tag IDs: %v", err)
	}

	// Update tag frequency for new tags
	return db.updateTagFrequency(tags)
}

// decrementTagFrequency decrements the frequency count for tags
func (db *DB) decrementTagFrequency(tags []string) error {
	log.Printf("🔄 Decrementing tag frequency for %d tags: %v", len(tags), tags)

	// Get streaming subsite ID first
	var streamingSubsiteID int
	err := db.QueryRow("SELECT id FROM subsites WHERE subsite_name = 'streaming'").Scan(&streamingSubsiteID)
	if err != nil {
		log.Printf("❌ Failed to get streaming subsite ID: %v", err)
		return fmt.Errorf("failed to get streaming subsite ID: %v", err)
	}

	// Get excluded words for streaming subsite
	excludedWords, err := db.GetExcludedWords(streamingSubsiteID)
	if err != nil {
		log.Printf("❌ Failed to get excluded words: %v", err)
		return fmt.Errorf("failed to get excluded words: %v", err)
	}

	var processedCount, excludedCount int

	for _, tag := range tags {
		cleanTag := strings.ToLower(strings.TrimSpace(tag))
		if cleanTag == "" {
			continue
		}

		// Check if tag is in exclusion list
		if excludedWords[cleanTag] {
			log.Printf("🚫 Skipping excluded tag during decrement: '%s'", cleanTag)
			excludedCount++
			continue
		}

		processedCount++

		// Decrement frequency for existing tag
		result, err := db.Exec(`
            UPDATE tags 
            SET frequency = GREATEST(frequency - 1, 0), updated_at = CURRENT_TIMESTAMP 
            WHERE word = $1 AND (subsite_id_origin = $2 OR $2 = ANY(COALESCE(subsite_ids, '{}')))
        `, cleanTag, streamingSubsiteID)

		if err != nil {
			log.Printf("❌ Failed to decrement tag frequency for '%s': %v", cleanTag, err)
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			log.Printf("✅ Decremented frequency for tag '%s'", cleanTag)
		}
	}

	log.Printf("🎉 Tag frequency decrement completed: %d processed, %d excluded, %d total input tags", processedCount, excludedCount, len(tags))
	return nil
}

// updateTagFrequency updates the frequency count for tags
func (db *DB) updateTagFrequency(tags []string) error {
	log.Printf("🔄 Updating tag frequency for %d tags: %v", len(tags), tags)

	// Get streaming subsite ID first
	var streamingSubsiteID int
	err := db.QueryRow("SELECT id FROM subsites WHERE subsite_name = 'streaming'").Scan(&streamingSubsiteID)
	if err != nil {
		log.Printf("❌ Failed to get streaming subsite ID: %v", err)
		return fmt.Errorf("failed to get streaming subsite ID: %v", err)
	}

	// Get excluded words for streaming subsite
	excludedWords, err := db.GetExcludedWords(streamingSubsiteID)
	if err != nil {
		log.Printf("❌ Failed to get excluded words: %v", err)
		return fmt.Errorf("failed to get excluded words: %v", err)
	}
	log.Printf("📋 Loaded %d excluded words for filtering", len(excludedWords))

	var processedCount, excludedCount int

	for _, tag := range tags {
		// Clean the tag
		cleanTag := strings.ToLower(strings.TrimSpace(tag))
		if cleanTag == "" {
			log.Printf("⚠️ Skipping empty tag: '%s'", tag)
			continue
		}

		// Check if tag is in exclusion list
		if excludedWords[cleanTag] {
			log.Printf("🚫 Skipping excluded tag: '%s' -> '%s'", tag, cleanTag)
			excludedCount++
			continue
		}

		log.Printf("📊 Processing tag: '%s' -> '%s'", tag, cleanTag)
		processedCount++

		// Check if tag already exists in streaming subsite
		var existingID int
		err = db.QueryRow("SELECT id FROM tags WHERE word = $1 AND (subsite_id_origin = $2 OR $2 = ANY(COALESCE(subsite_ids, '{}')))", cleanTag, streamingSubsiteID).Scan(&existingID)

		if err != nil && err != sql.ErrNoRows {
			log.Printf("❌ Error checking existing tag '%s': %v", cleanTag, err)
			return fmt.Errorf("failed to check existing tag '%s': %v", cleanTag, err)
		}

		if err == sql.ErrNoRows {
			// Tag doesn't exist, insert new one in streaming subsite
			log.Printf("➕ Inserting new tag '%s' in streaming subsite", cleanTag)
			_, err = db.Exec(`
				INSERT INTO tags (word, frequency, subsite_id_origin, subsite_ids, active_tag, created_at, updated_at, category_ids) 
				VALUES ($1, 1, $2, ARRAY[$2]::INTEGER[], true, NOW(), NOW(), '{}')
			`, cleanTag, streamingSubsiteID)

			if err != nil {
				log.Printf("❌ Failed to insert new tag '%s': %v", cleanTag, err)
				return fmt.Errorf("failed to insert new tag '%s': %v", cleanTag, err)
			}

			log.Printf("✅ New tag '%s' inserted successfully in streaming", cleanTag)
		} else {
			// Tag exists, increment frequency
			log.Printf("🔄 Updating existing tag '%s' (ID: %d) in streaming", cleanTag, existingID)
			_, err = db.Exec(`
				UPDATE tags 
				SET frequency = frequency + 1, updated_at = CURRENT_TIMESTAMP 
				WHERE id = $1
			`, existingID)

			if err != nil {
				log.Printf("❌ Failed to update tag frequency for '%s' in streaming: %v", cleanTag, err)
				return fmt.Errorf("failed to update tag frequency for '%s': %v", cleanTag, err)
			}

			log.Printf("✅ Tag '%s' frequency updated successfully in streaming", cleanTag)
		}
	}

	log.Printf("🎉 Tag frequency update completed: %d processed, %d excluded, %d total input tags", processedCount, excludedCount, len(tags))
	return nil
}

// GetTagAnalytics returns tag frequency and basic statistics
func (db *DB) GetTagAnalytics() (map[string]interface{}, error) {
	log.Printf("📊 Getting tag analytics...")

	// Get tag frequency from streaming subsite (using new tags table structure)
	rows, err := db.Query(`
		SELECT word, frequency 
		FROM tags
		WHERE active_tag = true 
		  AND (subsite_id_origin = 1 OR 1 = ANY(COALESCE(subsite_ids, '{}')))
		ORDER BY frequency DESC 
		LIMIT 100
	`)
	if err != nil {
		log.Printf("❌ Failed to query tag frequency: %v", err)
		return nil, fmt.Errorf("failed to query tag frequency: %v", err)
	}
	defer rows.Close()

	var tagFrequency []map[string]interface{}
	for rows.Next() {
		var word string
		var frequency int
		if err := rows.Scan(&word, &frequency); err != nil {
			log.Printf("❌ Failed to scan tag frequency row: %v", err)
			return nil, fmt.Errorf("failed to scan tag frequency: %v", err)
		}

		tagFrequency = append(tagFrequency, map[string]interface{}{
			"word":      word,
			"frequency": frequency,
		})
	}

	log.Printf("✅ Retrieved %d tag frequency records", len(tagFrequency))

	// Get tagging statistics
	var totalVideos, taggedVideos int
	err = db.QueryRow(`SELECT COUNT(*) FROM master_video_list`).Scan(&totalVideos)
	if err != nil {
		log.Printf("❌ Failed to get total video count: %v", err)
		return nil, fmt.Errorf("failed to get total video count: %v", err)
	}

	err = db.QueryRow(`SELECT COUNT(*) FROM master_video_list WHERE tagged = true`).Scan(&taggedVideos)
	if err != nil {
		log.Printf("❌ Failed to get tagged video count: %v", err)
		return nil, fmt.Errorf("failed to get tagged video count: %v", err)
	}

	// Get total unique tags in streaming subsite
	var totalUniqueTags int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM tags
		WHERE active_tag = true 
		  AND (subsite_id_origin = 1 OR 1 = ANY(COALESCE(subsite_ids, '{}')))
	`).Scan(&totalUniqueTags)
	if err != nil {
		log.Printf("❌ Failed to get total unique tags count: %v", err)
		return nil, fmt.Errorf("failed to get total unique tags count: %v", err)
	}

	// Get active vs inactive tag counts
	var activeTags, inactiveTags int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM tags
		WHERE active_tag = true 
		  AND (subsite_id_origin = 1 OR 1 = ANY(COALESCE(subsite_ids, '{}')))
	`).Scan(&activeTags)
	if err != nil {
		log.Printf("❌ Failed to get active tags count: %v", err)
		return nil, fmt.Errorf("failed to get active tags count: %v", err)
	}

	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM tags
		WHERE active_tag = false 
		  AND (subsite_id_origin = 1 OR 1 = ANY(COALESCE(subsite_ids, '{}')))
	`).Scan(&inactiveTags)
	if err != nil {
		log.Printf("❌ Failed to get inactive tags count: %v", err)
		return nil, fmt.Errorf("failed to get inactive tags count: %v", err)
	}

	log.Printf("📊 Analytics summary: %d total videos, %d tagged, %d unique tags (%d active, %d inactive)",
		totalVideos, taggedVideos, totalUniqueTags, activeTags, inactiveTags)

	result := map[string]interface{}{
		"tag_frequency":      tagFrequency,
		"total_videos":       totalVideos,
		"tagged_videos":      taggedVideos,
		"untagged_videos":    totalVideos - taggedVideos,
		"total_unique_tags":  totalUniqueTags,
		"active_tags":        activeTags,
		"inactive_tags":      inactiveTags,
		"tagging_percentage": float64(taggedVideos) / float64(totalVideos) * 100,
	}

	log.Printf("✅ Tag analytics retrieved successfully")
	return result, nil
}

// GetUntaggedVideos returns videos that haven't been tagged yet
func (db *DB) GetUntaggedVideos(limit int) ([]*MasterVideo, error) {
	rows, err := db.Query(`
		SELECT id, bunny_video_id, title, description, category, tags, tagged, duration, 
		       file_size, resolution, framerate, thumbnail_url, video_url, iframe_src, 
		       playback_url, status, views, likes, is_public, encode_progress, 
		       available_resolutions, collection_id, average_watch_time, total_watch_time,
		       last_bunny_sync, last_master_update, sync_status, sync_notes, 
		       metadata_version, created_by, created_at, updated_at, vid_status, tag_ids
		FROM master_video_list 
		WHERE tagged = false 
		ORDER BY created_at DESC 
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query untagged videos: %v", err)
	}
	defer rows.Close()

	var videos []*MasterVideo
	for rows.Next() {
		video := &MasterVideo{}
		var tagsStr, resolutionsStr, tagIDsStr sql.NullString
		var createdBy sql.NullInt64

		err := rows.Scan(
			&video.ID, &video.BunnyVideoID, &video.Title, &video.Description, &video.Category,
			&tagsStr, &video.Tagged, &video.Duration, &video.FileSize, &video.Resolution, &video.Framerate,
			&video.ThumbnailURL, &video.VideoURL, &video.IframeSrc, &video.PlaybackURL,
			&video.Status, &video.Views, &video.Likes, &video.IsPublic, &video.EncodeProgress,
			&resolutionsStr, &video.CollectionID, &video.AverageWatchTime, &video.TotalWatchTime,
			&video.LastBunnySync, &video.LastMasterUpdate, &video.SyncStatus, &video.SyncNotes,
			&video.MetadataVersion, &createdBy, &video.CreatedAt, &video.UpdatedAt, &video.Vid_Status, &tagIDsStr,
		)
		if err != nil {
			return nil, err
		}

		// Convert sql.NullInt64 to *int
		if createdBy.Valid {
			createdByInt := int(createdBy.Int64)
			video.CreatedBy = &createdByInt
		} else {
			video.CreatedBy = nil
		}

		// Parse tags from JSON (handle NULL)
		if tagsStr.Valid && tagsStr.String != "" {
			if err := json.Unmarshal([]byte(tagsStr.String), &video.Tags); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tags: %v", err)
			}
		}

		// Parse resolutions from JSON (handle NULL)
		if resolutionsStr.Valid && resolutionsStr.String != "" {
			if err := json.Unmarshal([]byte(resolutionsStr.String), &video.AvailableResolutions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal resolutions: %v", err)
			}
		}

		// Parse tag IDs from JSON (handle NULL)
		if tagIDsStr.Valid && tagIDsStr.String != "" {
			if err := json.Unmarshal([]byte(tagIDsStr.String), &video.TagIDs); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tag IDs: %v", err)
			}
		}

		videos = append(videos, video)
	}

	return videos, nil
}

// GetSubsiteTags returns tags for a specific subsite
func (db *DB) GetSubsiteTags(subsite string) ([]map[string]interface{}, error) {
	log.Printf("📊 Getting tags for subsite: %s", subsite)

	// Get subsite ID
	var subsiteID int
	err := db.QueryRow("SELECT id FROM subsites WHERE subsite_name = $1", subsite).Scan(&subsiteID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("⚠️ Subsite '%s' not found, returning empty tags", subsite)
			return []map[string]interface{}{}, nil
		}
		log.Printf("❌ Failed to get subsite ID for '%s': %v", subsite, err)
		return nil, fmt.Errorf("failed to get subsite ID for '%s': %v", subsite, err)
	}

	// Get tags for this subsite
	rows, err := db.Query(`
		SELECT t.id, t.word, t.frequency, t.active_tag
		FROM tags t
		WHERE t.subsite_id_origin = $1 OR $1 = ANY(COALESCE(t.subsite_ids, '{}'))
		ORDER BY t.frequency DESC, t.word ASC
	`, subsiteID)
	if err != nil {
		log.Printf("❌ Failed to query subsite tags: %v", err)
		return nil, fmt.Errorf("failed to query subsite tags: %v", err)
	}
	defer rows.Close()

	var tags []map[string]interface{}
	for rows.Next() {
		var id int
		var word string
		var frequency int
		var activeTag bool

		if err := rows.Scan(&id, &word, &frequency, &activeTag); err != nil {
			log.Printf("❌ Failed to scan subsite tag row: %v", err)
			return nil, fmt.Errorf("failed to scan subsite tag: %v", err)
		}

		tag := map[string]interface{}{
			"id":         id,
			"word":       word,
			"frequency":  frequency,
			"active_tag": activeTag,
		}

		tags = append(tags, tag)
	}

	log.Printf("✅ Retrieved %d tags for subsite '%s'", len(tags), subsite)
	return tags, nil
}

// GetSubsiteCategories returns categories for a specific subsite
func (db *DB) GetSubsiteCategories(subsite string) ([]map[string]interface{}, error) {
	log.Printf("📁 Getting categories for subsite: %s", subsite)

	// Get subsite ID
	var subsiteID int
	err := db.QueryRow("SELECT id FROM subsites WHERE subsite_name = $1", subsite).Scan(&subsiteID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("⚠️ Subsite '%s' not found, returning empty categories", subsite)
			return []map[string]interface{}{}, nil
		}
		log.Printf("❌ Failed to get subsite ID for '%s': %v", subsite, err)
		return nil, fmt.Errorf("failed to get subsite ID for '%s': %v", subsite, err)
	}

	// Get categories for this subsite
	rows, err := db.Query(`
		SELECT id, name, description, color, created_at, updated_at
		FROM tag_categories
		WHERE subsite_id = $1 OR $1 = ANY(COALESCE(subsite_ids, '{}'))
		ORDER BY name ASC
	`, subsiteID)
	if err != nil {
		log.Printf("❌ Failed to query subsite categories: %v", err)
		return nil, fmt.Errorf("failed to query subsite categories: %v", err)
	}
	defer rows.Close()

	var categories []map[string]interface{}
	for rows.Next() {
		var id int
		var name, description, color string
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &name, &description, &color, &createdAt, &updatedAt); err != nil {
			log.Printf("❌ Failed to scan subsite category row: %v", err)
			return nil, fmt.Errorf("failed to scan subsite category: %v", err)
		}

		category := map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": description,
			"color":       color,
			"created_at":  createdAt,
			"updated_at":  updatedAt,
		}

		categories = append(categories, category)
	}

	log.Printf("✅ Retrieved %d categories for subsite '%s'", len(categories), subsite)
	return categories, nil
}

// AddSubsiteTag adds a new tag to a specific subsite
func (db *DB) AddSubsiteTag(subsite, word string) error {
	log.Printf("➕ Adding tag '%s' to subsite '%s'", word, subsite)

	// Get subsite ID
	var subsiteID int
	err := db.QueryRow("SELECT id FROM subsites WHERE subsite_name = $1", subsite).Scan(&subsiteID)
	if err != nil {
		log.Printf("❌ Failed to get subsite ID for '%s': %v", subsite, err)
		return fmt.Errorf("failed to get subsite ID for '%s': %v", subsite, err)
	}

	// Clean the tag
	cleanWord := strings.ToLower(strings.TrimSpace(word))
	if cleanWord == "" {
		return fmt.Errorf("tag cannot be empty")
	}

	// Check if tag already exists for this subsite
	var existingID int
	err = db.QueryRow("SELECT id FROM tags WHERE word = $1 AND (subsite_id_origin = $2 OR $2 = ANY(COALESCE(subsite_ids, '{}')))", cleanWord, subsiteID).Scan(&existingID)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("❌ Error checking existing tag '%s' in subsite '%s': %v", cleanWord, subsite, err)
		return fmt.Errorf("failed to check existing tag: %v", err)
	}

	if err == sql.ErrNoRows {
		// Tag doesn't exist, insert new one
		log.Printf("➕ Inserting new tag '%s' in subsite '%s'", cleanWord, subsite)
		_, err = db.Exec(`
			INSERT INTO tags (word, frequency, subsite_id_origin, subsite_ids, active_tag, created_at, updated_at, category_ids) 
			VALUES ($1, 1, $2, ARRAY[$2]::INTEGER[], true, NOW(), NOW(), '{}')
		`, cleanWord, subsiteID)

		if err != nil {
			log.Printf("❌ Failed to insert new tag '%s' in subsite '%s': %v", cleanWord, subsite, err)
			return fmt.Errorf("failed to insert new tag: %v", err)
		}

		log.Printf("✅ New tag '%s' inserted successfully in subsite '%s'", cleanWord, subsite)
	} else {
		// Tag exists, increment frequency
		log.Printf("🔄 Updating existing tag '%s' frequency in subsite '%s'", cleanWord, subsite)
		_, err = db.Exec(`
			UPDATE tags 
			SET frequency = frequency + 1, updated_at = CURRENT_TIMESTAMP 
			WHERE id = $1
		`, existingID)

		if err != nil {
			log.Printf("❌ Failed to update tag frequency for '%s' in subsite '%s': %v", cleanWord, subsite, err)
			return fmt.Errorf("failed to update tag frequency: %v", err)
		}

		log.Printf("✅ Tag '%s' frequency updated successfully in subsite '%s'", cleanWord, subsite)
	}

	return nil
}

// DeleteSubsiteTag deletes a tag from a specific subsite
func (db *DB) DeleteSubsiteTag(subsite string, tagID int) error {
	log.Printf("🗑️ Deleting tag ID %d from subsite '%s'", tagID, subsite)

	// Get subsite ID
	var subsiteID int
	err := db.QueryRow("SELECT id FROM subsites WHERE subsite_name = $1", subsite).Scan(&subsiteID)
	if err != nil {
		log.Printf("❌ Failed to get subsite ID for '%s': %v", subsite, err)
		return fmt.Errorf("failed to get subsite ID for '%s': %v", subsite, err)
	}

	// Delete the tag (this will cascade to remove any category assignments)
	result, err := db.Exec(`
		DELETE FROM tags 
		WHERE id = $1 AND (subsite_id_origin = $2 OR $2 = ANY(COALESCE(subsite_ids, '{}')))
	`, tagID, subsiteID)
	if err != nil {
		log.Printf("❌ Failed to delete tag ID %d from subsite '%s': %v", tagID, subsite, err)
		return fmt.Errorf("failed to delete tag: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("⚠️ Could not determine rows affected for tag deletion")
	} else if rowsAffected == 0 {
		log.Printf("⚠️ No tag found with ID %d in subsite '%s'", tagID, subsite)
		return fmt.Errorf("tag not found in subsite")
	}

	log.Printf("✅ Tag ID %d deleted successfully from subsite '%s'", tagID, subsite)
	return nil
}

// AddSubsiteCategory adds a new category to a specific subsite
func (db *DB) AddSubsiteCategory(subsite, name, color, description string) error {
	log.Printf("➕ Adding category '%s' to subsite '%s'", name, subsite)

	// Get subsite ID
	var subsiteID int
	err := db.QueryRow("SELECT id FROM subsites WHERE subsite_name = $1", subsite).Scan(&subsiteID)
	if err != nil {
		log.Printf("❌ Failed to get subsite ID for '%s': %v", subsite, err)
		return fmt.Errorf("failed to get subsite ID for '%s': %v", subsite, err)
	}

	// Clean the name
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return fmt.Errorf("category name cannot be empty")
	}

	// Check if category already exists for this subsite
	var existingID int
	err = db.QueryRow("SELECT id FROM tag_categories WHERE name = $1 AND subsite_id = $2", cleanName, subsiteID).Scan(&existingID)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("❌ Error checking existing category '%s' in subsite '%s': %v", cleanName, subsite, err)
		return fmt.Errorf("failed to check existing category: %v", err)
	}

	if err == sql.ErrNoRows {
		// Category doesn't exist, insert new one
		log.Printf("➕ Inserting new category '%s' in subsite '%s'", cleanName, subsite)
		_, err = db.Exec(`
			INSERT INTO tag_categories (name, color, description, subsite_id) 
			VALUES ($1, $2, $3, $4)
		`, cleanName, color, description, subsiteID)

		if err != nil {
			log.Printf("❌ Failed to insert new category '%s' in subsite '%s': %v", cleanName, subsite, err)
			return fmt.Errorf("failed to insert new category: %v", err)
		}

		log.Printf("✅ New category '%s' inserted successfully in subsite '%s'", cleanName, subsite)
	} else {
		log.Printf("⚠️ Category '%s' already exists in subsite '%s'", cleanName, subsite)
		return fmt.Errorf("category already exists in subsite")
	}

	return nil
}

// DeleteSubsiteCategory deletes a category from a specific subsite
func (db *DB) DeleteSubsiteCategory(subsite string, categoryID int) error {
	log.Printf("🗑️ Deleting category ID %d from subsite '%s'", categoryID, subsite)

	// Get subsite ID
	var subsiteID int
	err := db.QueryRow("SELECT id FROM subsites WHERE subsite_name = $1", subsite).Scan(&subsiteID)
	if err != nil {
		log.Printf("❌ Failed to get subsite ID for '%s': %v", subsite, err)
		return fmt.Errorf("failed to get subsite ID for '%s': %v", subsite, err)
	}

	// Delete the category (this will cascade to remove any tag assignments)
	result, err := db.Exec(`
		DELETE FROM tag_categories 
		WHERE id = $1 AND subsite_id = $2
	`, categoryID, subsiteID)
	if err != nil {
		log.Printf("❌ Failed to delete category ID %d from subsite '%s': %v", categoryID, subsite, err)
		return fmt.Errorf("failed to delete category: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("⚠️ Could not determine rows affected for category deletion")
	} else if rowsAffected == 0 {
		log.Printf("⚠️ No category found with ID %d in subsite '%s'", categoryID, subsite)
		return fmt.Errorf("category not found in subsite")
	}

	log.Printf("✅ Category ID %d deleted successfully from subsite '%s'", categoryID, subsite)
	return nil
}

// RemoveTagFromCategory removes a tag from its category within a specific subsite
func (db *DB) RemoveTagFromCategory(subsite string, tagID int) error {
	log.Printf("🔗 Removing tag ID %d from category in subsite '%s'", tagID, subsite)

	// Get subsite ID
	var subsiteID int
	err := db.QueryRow("SELECT id FROM subsites WHERE subsite_name = $1", subsite).Scan(&subsiteID)
	if err != nil {
		log.Printf("❌ Failed to get subsite ID for '%s': %v", subsite, err)
		return fmt.Errorf("failed to get subsite ID for '%s': %v", subsite, err)
	}

	// Verify tag belongs to this subsite
	var tagSubsiteIDOrigin sql.NullInt64
	err = db.QueryRow("SELECT subsite_id_origin FROM tags WHERE id = $1", tagID).Scan(&tagSubsiteIDOrigin)
	if err != nil {
		log.Printf("❌ Failed to get tag subsite ID: %v", err)
		return fmt.Errorf("failed to get tag subsite ID: %v", err)
	}

	if !tagSubsiteIDOrigin.Valid || int(tagSubsiteIDOrigin.Int64) != subsiteID {
		log.Printf("❌ Tag does not belong to subsite '%s'", subsite)
		return fmt.Errorf("tag does not belong to subsite")
	}

	// Clear the tag's category_id
	_, err = db.Exec(`
		UPDATE tags 
		SET category_id = NULL, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1
	`, tagID)
	if err != nil {
		log.Printf("❌ Failed to remove tag from category: %v", err)
		return fmt.Errorf("failed to remove tag from category: %v", err)
	}

	log.Printf("✅ Tag ID %d successfully removed from category in subsite '%s'", tagID, subsite)
	return nil
}

// AssignSubsiteTagToCategory assigns a tag to a category within a specific subsite
func (db *DB) AssignSubsiteTagToCategory(subsite string, tagID, categoryID int) error {
	log.Printf("🔗 Assigning tag ID %d to category ID %d in subsite '%s'", tagID, categoryID, subsite)

	// Get subsite ID
	var subsiteID int
	err := db.QueryRow("SELECT id FROM subsites WHERE subsite_name = $1", subsite).Scan(&subsiteID)
	if err != nil {
		log.Printf("❌ Failed to get subsite ID for '%s': %v", subsite, err)
		return fmt.Errorf("failed to get subsite ID for '%s': %v", subsite, err)
	}

	// Verify both tag and category belong to this subsite
	var tagSubsiteIDOrigin sql.NullInt64
	var categorySubsiteID int
	err = db.QueryRow("SELECT subsite_id_origin FROM tags WHERE id = $1", tagID).Scan(&tagSubsiteIDOrigin)
	if err != nil {
		log.Printf("❌ Failed to get tag subsite ID: %v", err)
		return fmt.Errorf("failed to get tag subsite ID: %v", err)
	}

	err = db.QueryRow("SELECT subsite_id FROM tag_categories WHERE id = $1", categoryID).Scan(&categorySubsiteID)
	if err != nil {
		log.Printf("❌ Failed to get category subsite ID: %v", err)
		return fmt.Errorf("failed to get category subsite ID: %v", err)
	}

	if !tagSubsiteIDOrigin.Valid || int(tagSubsiteIDOrigin.Int64) != subsiteID || categorySubsiteID != subsiteID {
		log.Printf("❌ Tag or category does not belong to subsite '%s'", subsite)
		return fmt.Errorf("tag or category does not belong to subsite")
	}

	// Update the tag's category_id
	_, err = db.Exec(`
		UPDATE tags 
		SET category_id = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2
	`, categoryID, tagID)
	if err != nil {
		log.Printf("❌ Failed to assign tag to category: %v", err)
		return fmt.Errorf("failed to assign tag to category: %v", err)
	}

	log.Printf("✅ Tag ID %d successfully assigned to category ID %d in subsite '%s'", tagID, categoryID, subsite)
	return nil
}

// ToggleTagActiveStatus toggles the active status of a tag
func (db *DB) ToggleTagActiveStatus(subsite string, tagID int) error {
	log.Printf("🔄 Toggling active status for tag ID %d in subsite '%s'", tagID, subsite)

	// Get subsite ID
	var subsiteID int
	err := db.QueryRow("SELECT id FROM subsites WHERE subsite_name = $1", subsite).Scan(&subsiteID)
	if err != nil {
		log.Printf("❌ Failed to get subsite ID for '%s': %v", subsite, err)
		return fmt.Errorf("failed to get subsite ID for '%s': %v", subsite, err)
	}

	// Toggle the active status
	_, err = db.Exec(`
		UPDATE tags 
		SET active_tag = NOT active_tag, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1 AND (subsite_id_origin = $2 OR $2 = ANY(COALESCE(subsite_ids, '{}')))
	`, tagID, subsiteID)
	if err != nil {
		log.Printf("❌ Failed to toggle tag active status: %v", err)
		return fmt.Errorf("failed to toggle tag active status: %v", err)
	}

	log.Printf("✅ Tag ID %d active status toggled successfully in subsite '%s'", tagID, subsite)
	return nil
}

// GetTagCategories returns all tag categories (for admin dashboard)
func (db *DB) GetTagCategories() ([]map[string]interface{}, error) {
	log.Printf("📁 Getting all tag categories...")

	rows, err := db.Query(`
		SELECT tc.id, tc.name, tc.description, tc.color, tc.created_at, tc.updated_at,
		       s.subsite_name as subsite_name, s.subsite_name as subsite_display_name
		FROM tag_categories tc
		JOIN subsites s ON tc.subsite_id = s.id
		ORDER BY s.subsite_name ASC, tc.name ASC
	`)
	if err != nil {
		log.Printf("❌ Failed to query tag categories: %v", err)
		return nil, fmt.Errorf("failed to query tag categories: %v", err)
	}
	defer rows.Close()

	var categories []map[string]interface{}
	for rows.Next() {
		var id int
		var name, description, color, subsiteName, subsiteDisplayName string
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &name, &description, &color, &createdAt, &updatedAt, &subsiteName, &subsiteDisplayName); err != nil {
			log.Printf("❌ Failed to scan tag category row: %v", err)
			return nil, fmt.Errorf("failed to scan tag category: %v", err)
		}

		category := map[string]interface{}{
			"id":                   id,
			"name":                 name,
			"description":          description,
			"color":                color,
			"created_at":           createdAt,
			"updated_at":           updatedAt,
			"subsite_name":         subsiteName,
			"subsite_display_name": subsiteDisplayName,
		}

		categories = append(categories, category)
	}

	log.Printf("✅ Retrieved %d tag categories", len(categories))
	return categories, nil
}

// Article Exclusions Management
func (db *DB) GetArticleExclusions(subsiteID int) ([]*ArticleExclusion, error) {
	query := `
		SELECT id, word, excluded, subsite_id, created_at, updated_at
		FROM article_exclusions 
		WHERE subsite_id = $1
		ORDER BY word ASC
	`

	rows, err := db.Query(query, subsiteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query article exclusions: %v", err)
	}
	defer rows.Close()

	var exclusions []*ArticleExclusion
	for rows.Next() {
		var exclusion ArticleExclusion
		err := rows.Scan(
			&exclusion.ID,
			&exclusion.Word,
			&exclusion.Excluded,
			&exclusion.SubsiteID,
			&exclusion.CreatedAt,
			&exclusion.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan article exclusion: %v", err)
		}
		exclusions = append(exclusions, &exclusion)
	}

	return exclusions, nil
}

func (db *DB) AddArticleExclusion(subsiteID int, word string) error {
	query := `
		INSERT INTO article_exclusions (word, excluded, subsite_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (word) DO UPDATE SET 
			excluded = EXCLUDED.excluded,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.Exec(query, strings.ToLower(word), true, subsiteID)
	if err != nil {
		return fmt.Errorf("failed to add article exclusion: %v", err)
	}

	return nil
}

func (db *DB) ToggleArticleExclusion(subsiteID int, word string, excluded bool) error {
	query := `
		UPDATE article_exclusions 
		SET excluded = $1, updated_at = CURRENT_TIMESTAMP
		WHERE word = $2 AND subsite_id = $3
	`

	result, err := db.Exec(query, excluded, strings.ToLower(word), subsiteID)
	if err != nil {
		return fmt.Errorf("failed to toggle article exclusion: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("article exclusion not found: %s", word)
	}

	return nil
}

func (db *DB) RemoveArticleExclusion(subsiteID int, word string) error {
	query := `DELETE FROM article_exclusions WHERE word = $1 AND subsite_id = $2`

	result, err := db.Exec(query, strings.ToLower(word), subsiteID)
	if err != nil {
		return fmt.Errorf("failed to remove article exclusion: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("article exclusion not found: %s", word)
	}

	return nil
}

func (db *DB) GetExcludedWords(subsiteID int) (map[string]bool, error) {
	query := `
		SELECT word 
		FROM article_exclusions 
		WHERE subsite_id = $1 AND excluded = true
	`

	rows, err := db.Query(query, subsiteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query excluded words: %v", err)
	}
	defer rows.Close()

	excludedWords := make(map[string]bool)
	for rows.Next() {
		var word string
		err := rows.Scan(&word)
		if err != nil {
			return nil, fmt.Errorf("failed to scan excluded word: %v", err)
		}
		excludedWords[word] = true
	}

	return excludedWords, nil
}

// GetSubsiteID returns the ID for a given subsite name
func (db *DB) GetSubsiteID(subsiteName string) (int, error) {
	var subsiteID int
	err := db.QueryRow("SELECT id FROM subsites WHERE subsite_name = $1", subsiteName).Scan(&subsiteID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("subsite '%s' not found", subsiteName)
		}
		return 0, fmt.Errorf("failed to get subsite ID for '%s': %v", subsiteName, err)
	}
	return subsiteID, nil
}

// ResetAllTagFrequencies resets all tag frequencies to 0 for streaming subsite
func (db *DB) ResetAllTagFrequencies() error {
	var streamingSubsiteID int
	err := db.QueryRow("SELECT id FROM subsites WHERE subsite_name = 'streaming'").Scan(&streamingSubsiteID)
	if err != nil {
		return fmt.Errorf("failed to get streaming subsite ID: %v", err)
	}

	_, err = db.Exec(`UPDATE tags SET frequency = 0, updated_at = CURRENT_TIMESTAMP WHERE subsite_id_origin = $1 OR $1 = ANY(COALESCE(subsite_ids, '{}'))`, streamingSubsiteID)
	if err != nil {
		return fmt.Errorf("failed to reset tag frequencies: %v", err)
	}

	log.Printf("✅ Reset all tag frequencies for streaming subsite")
	return nil
}

// convertTagWordsToIDs converts tag words to their corresponding tag IDs
func (db *DB) convertTagWordsToIDs(tagWords []string) ([]int, error) {
	if len(tagWords) == 0 {
		return []int{}, nil
	}

	// Get streaming subsite ID
	var streamingSubsiteID int
	err := db.QueryRow("SELECT id FROM subsites WHERE subsite_name = 'streaming'").Scan(&streamingSubsiteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get streaming subsite ID: %v", err)
	}

	tagIDs := make([]int, 0, len(tagWords))

	for _, word := range tagWords {
		var tagID int
		err := db.QueryRow(`
			SELECT id FROM tags 
			WHERE word = $1 AND (subsite_id_origin = $2 OR $2 = ANY(COALESCE(subsite_ids, '{}')))
		`, word, streamingSubsiteID).Scan(&tagID)

		if err == sql.ErrNoRows {
			// Tag doesn't exist, create it
			newTag, err := db.CreateTag(word, &streamingSubsiteID)
			if err != nil {
				log.Printf("⚠️ Warning: Failed to create tag '%s': %v", word, err)
				continue
			}
			tagIDs = append(tagIDs, newTag.ID)
			log.Printf("✅ Created new tag '%s' with ID %d", word, newTag.ID)
		} else if err != nil {
			log.Printf("⚠️ Warning: Failed to get tag ID for '%s': %v", word, err)
			continue
		} else {
			tagIDs = append(tagIDs, tagID)
		}
	}

	return tagIDs, nil
}

// convertTagIDsToWords converts tag IDs to their corresponding tag words
func (db *DB) convertTagIDsToWords(tagIDs pq.Int64Array) ([]string, error) {
	if len(tagIDs) == 0 {
		return []string{}, nil
	}

	// Convert pq.Int64Array to []int for the query
	intTagIDs := make([]int, len(tagIDs))
	for i, id := range tagIDs {
		intTagIDs[i] = int(id)
	}

	query := `SELECT word FROM tags WHERE id = ANY($1)`
	rows, err := db.Query(query, pq.Array(intTagIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to query tag words: %v", err)
	}
	defer rows.Close()

	var tagWords []string
	for rows.Next() {
		var word string
		if err := rows.Scan(&word); err != nil {
			log.Printf("⚠️ Warning: Failed to scan tag word: %v", err)
			continue
		}
		tagWords = append(tagWords, word)
	}

	return tagWords, nil
}

// GetVideosByTagCategory gets videos that have tags associated with a specific category
func (db *DB) GetVideosByTagCategory(categoryID int, page, limit int) ([]MasterVideo, int, error) {
	// First, get all tag IDs for this category
	var tagIDs pq.Int64Array
	err := db.QueryRow(`
		SELECT COALESCE(tag_ids, '{}') FROM tag_categories 
		WHERE id = $1
	`, categoryID).Scan(&tagIDs)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get tag IDs for category %d: %v", categoryID, err)
	}

	if len(tagIDs) == 0 {
		// No tags in this category, return empty result
		return []MasterVideo{}, 0, nil
	}

	// Convert to []int for the query
	intTagIDs := make([]int, len(tagIDs))
	for i, id := range tagIDs {
		intTagIDs[i] = int(id)
	}

	// Calculate offset for pagination
	offset := (page - 1) * limit

	// Query videos that have any of these tag IDs
	query := `
		SELECT id, title, description, video_url, thumbnail_url, duration, 
		       created_at, updated_at, tagged, COALESCE(tag_ids, '{}')
		FROM master_video_list 
		WHERE tag_ids && $1::INTEGER[]
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := db.Query(query, pq.Array(intTagIDs), limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query videos by tag category: %v", err)
	}
	defer rows.Close()

	var videos []MasterVideo
	for rows.Next() {
		var video MasterVideo
		var tagIDsArray pq.Int64Array

		err := rows.Scan(
			&video.ID,
			&video.Title,
			&video.Description,
			&video.VideoURL,
			&video.ThumbnailURL,
			&video.Duration,
			&video.CreatedAt,
			&video.UpdatedAt,
			&video.Tagged,
			&tagIDsArray,
		)
		if err != nil {
			log.Printf("⚠️ Warning: Failed to scan video row: %v", err)
			continue
		}

		// Convert tag IDs to integers
		video.TagIDs = make([]int, len(tagIDsArray))
		for i, id := range tagIDsArray {
			video.TagIDs[i] = int(id)
		}

		videos = append(videos, video)
	}

	// Get total count for pagination
	countQuery := `
		SELECT COUNT(*) FROM master_video_list 
		WHERE tag_ids && $1::INTEGER[]
	`
	var totalCount int
	err = db.QueryRow(countQuery, pq.Array(intTagIDs)).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count videos by tag category: %v", err)
	}

	log.Printf("✅ Retrieved %d videos for category %d (page %d, total: %d)", len(videos), categoryID, page, totalCount)
	return videos, totalCount, nil
}
