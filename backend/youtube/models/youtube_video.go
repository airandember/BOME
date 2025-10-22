package models

import (
	"database/sql"
	"fmt"
	"time"

	"bome-backend/infrastructure/database"

	"github.com/lib/pq"
)

// YouTubeVideo represents a YouTube video record
type YouTubeVideo struct {
	ID           string    `json:"id" db:"id"`
	Title        string    `json:"title" db:"title"`
	Description  string    `json:"description" db:"description"`
	PublishedAt  time.Time `json:"published_at" db:"published_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	ThumbnailURL string    `json:"thumbnail_url" db:"thumbnail_url"`
	VideoURL     string    `json:"video_url" db:"video_url"`
	EmbedURL     string    `json:"embed_url" db:"embed_url"`
	Duration     string    `json:"duration" db:"duration"`
	ViewCount    int64     `json:"view_count" db:"view_count"`
	Tags         []string  `json:"tags" db:"tags"`
	Category     string    `json:"category" db:"category"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// YouTubeSyncLog represents a sync operation log entry
type YouTubeSyncLog struct {
	ID            int        `json:"id" db:"id"`
	SyncType      string     `json:"sync_type" db:"sync_type"` // 'rss', 'manual', 'scheduled'
	StartedAt     time.Time  `json:"started_at" db:"started_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	VideosFound   int        `json:"videos_found" db:"videos_found"`
	VideosNew     int        `json:"videos_new" db:"videos_new"`
	VideosUpdated int        `json:"videos_updated" db:"videos_updated"`
	VideosSkipped int        `json:"videos_skipped" db:"videos_skipped"`
	Status        string     `json:"status" db:"status"` // 'running', 'success', 'partial', 'failed'
	ErrorMessage  string     `json:"error_message,omitempty" db:"error_message"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// YouTubeConfig represents YouTube configuration
type YouTubeConfig struct {
	ID               int        `json:"id" db:"id"`
	ChannelID        string     `json:"channel_id" db:"channel_id"`
	ChannelTitle     string     `json:"channel_title" db:"channel_title"`
	RSSURL           string     `json:"rss_url" db:"rss_url"`
	SyncEnabled      bool       `json:"sync_enabled" db:"sync_enabled"`
	SyncSchedule     string     `json:"sync_schedule" db:"sync_schedule"` // Cron format
	AutoSyncToMaster bool       `json:"auto_sync_to_master" db:"auto_sync_to_master"`
	LastSyncAt       *time.Time `json:"last_sync_at,omitempty" db:"last_sync_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// ChannelInfo represents YouTube channel information
type ChannelInfo struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	SubscriberCount int       `json:"subscriber_count"`
	VideoCount      int       `json:"video_count"`
	ViewCount       int       `json:"view_count"`
	PublishedAt     time.Time `json:"published_at"`
	Country         string    `json:"country"`
	CustomURL       string    `json:"custom_url"`
	ThumbnailURL    string    `json:"thumbnail_url"`
}

// ================================================
// CRUD OPERATIONS
// ================================================

// CreateYouTubeVideo creates a new YouTube video record
func CreateYouTubeVideo(db *database.DB, video *YouTubeVideo) error {
	query := `
		INSERT INTO youtube_videos (
			id, title, description, published_at, updated_at,
			thumbnail_url, video_url, embed_url, duration, view_count, 
			tags, category, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO NOTHING`

	_, err := db.Exec(query,
		video.ID,
		video.Title,
		video.Description,
		video.PublishedAt,
		video.UpdatedAt,
		video.ThumbnailURL,
		video.VideoURL,
		video.EmbedURL,
		video.Duration,
		video.ViewCount,
		pq.Array(video.Tags),
		video.Category,
		video.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create YouTube video: %w", err)
	}

	return nil
}

// GetYouTubeVideoByID retrieves a YouTube video by ID
func GetYouTubeVideoByID(db *database.DB, id string) (*YouTubeVideo, error) {
	query := `
		SELECT id, title, description, published_at, updated_at,
		       thumbnail_url, video_url, embed_url, duration, view_count, 
		       COALESCE(tags, '{}') as tags, COALESCE(category, '') as category, 
		       created_at
		FROM youtube_videos 
		WHERE id = $1`

	var video YouTubeVideo
	err := db.QueryRow(query, id).Scan(
		&video.ID,
		&video.Title,
		&video.Description,
		&video.PublishedAt,
		&video.UpdatedAt,
		&video.ThumbnailURL,
		&video.VideoURL,
		&video.EmbedURL,
		&video.Duration,
		&video.ViewCount,
		pq.Array(&video.Tags),
		&video.Category,
		&video.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("video not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get YouTube video: %w", err)
	}

	return &video, nil
}

// UpdateYouTubeVideo updates an existing YouTube video
func UpdateYouTubeVideo(db *database.DB, video *YouTubeVideo) error {
	query := `
		UPDATE youtube_videos 
		SET title = $1, description = $2, published_at = $3, updated_at = $4,
		    thumbnail_url = $5, video_url = $6, embed_url = $7, duration = $8, 
		    view_count = $9, tags = $10, category = $11
		WHERE id = $12`

	_, err := db.Exec(query,
		video.Title,
		video.Description,
		video.PublishedAt,
		video.UpdatedAt,
		video.ThumbnailURL,
		video.VideoURL,
		video.EmbedURL,
		video.Duration,
		video.ViewCount,
		pq.Array(video.Tags),
		video.Category,
		video.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update YouTube video: %w", err)
	}

	return nil
}

// DeleteYouTubeVideo deletes a YouTube video by ID
func DeleteYouTubeVideo(db *database.DB, id string) error {
	query := `DELETE FROM youtube_videos WHERE id = $1`

	_, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete YouTube video: %w", err)
	}

	return nil
}

// ================================================
// QUERY OPERATIONS
// ================================================

// GetYouTubeVideos retrieves YouTube videos with optional limit
func GetYouTubeVideos(db *database.DB, limit int) ([]*YouTubeVideo, error) {
	query := `
		SELECT id, title, description, published_at, updated_at,
		       thumbnail_url, video_url, embed_url, duration, view_count,
		       COALESCE(tags, '{}') as tags, COALESCE(category, '') as category,
		       created_at
		FROM youtube_videos 
		ORDER BY published_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query YouTube videos: %w", err)
	}
	defer rows.Close()

	var videos []*YouTubeVideo
	for rows.Next() {
		var video YouTubeVideo
		err := rows.Scan(
			&video.ID,
			&video.Title,
			&video.Description,
			&video.PublishedAt,
			&video.UpdatedAt,
			&video.ThumbnailURL,
			&video.VideoURL,
			&video.EmbedURL,
			&video.Duration,
			&video.ViewCount,
			pq.Array(&video.Tags),
			&video.Category,
			&video.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan YouTube video: %w", err)
		}
		videos = append(videos, &video)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return videos, nil
}

// GetYouTubeVideoCount returns the total count of YouTube videos
func GetYouTubeVideoCount(db *database.DB) (int, error) {
	query := `SELECT COUNT(*) FROM youtube_videos`

	var count int
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get YouTube video count: %w", err)
	}

	return count, nil
}

// SearchYouTubeVideos searches videos by title or description
func SearchYouTubeVideos(db *database.DB, searchTerm string, limit int) ([]*YouTubeVideo, error) {
	query := `
		SELECT id, title, description, published_at, updated_at,
		       thumbnail_url, video_url, embed_url, duration, view_count,
		       COALESCE(tags, '{}') as tags, COALESCE(category, '') as category,
		       created_at
		FROM youtube_videos 
		WHERE title ILIKE $1 OR description ILIKE $1
		ORDER BY published_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	searchPattern := "%" + searchTerm + "%"
	rows, err := db.Query(query, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search YouTube videos: %w", err)
	}
	defer rows.Close()

	var videos []*YouTubeVideo
	for rows.Next() {
		var video YouTubeVideo
		err := rows.Scan(
			&video.ID,
			&video.Title,
			&video.Description,
			&video.PublishedAt,
			&video.UpdatedAt,
			&video.ThumbnailURL,
			&video.VideoURL,
			&video.EmbedURL,
			&video.Duration,
			&video.ViewCount,
			pq.Array(&video.Tags),
			&video.Category,
			&video.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan YouTube video: %w", err)
		}
		videos = append(videos, &video)
	}

	return videos, nil
}

// GetYouTubeVideosByCategory retrieves videos by category
func GetYouTubeVideosByCategory(db *database.DB, category string, limit int) ([]*YouTubeVideo, error) {
	query := `
		SELECT id, title, description, published_at, updated_at,
		       thumbnail_url, video_url, embed_url, duration, view_count,
		       COALESCE(tags, '{}') as tags, COALESCE(category, '') as category,
		       created_at
		FROM youtube_videos 
		WHERE category = $1
		ORDER BY published_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := db.Query(query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to query videos by category: %w", err)
	}
	defer rows.Close()

	var videos []*YouTubeVideo
	for rows.Next() {
		var video YouTubeVideo
		err := rows.Scan(
			&video.ID,
			&video.Title,
			&video.Description,
			&video.PublishedAt,
			&video.UpdatedAt,
			&video.ThumbnailURL,
			&video.VideoURL,
			&video.EmbedURL,
			&video.Duration,
			&video.ViewCount,
			pq.Array(&video.Tags),
			&video.Category,
			&video.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan YouTube video: %w", err)
		}
		videos = append(videos, &video)
	}

	return videos, nil
}

// GetAllCategories returns all unique categories
func GetAllCategories(db *database.DB) ([]string, error) {
	query := `
		SELECT DISTINCT category 
		FROM youtube_videos 
		WHERE category IS NOT NULL AND category != ''
		ORDER BY category`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// GetAllTags returns all unique tags
func GetAllTags(db *database.DB) ([]string, error) {
	query := `
		SELECT DISTINCT unnest(tags) as tag 
		FROM youtube_videos 
		WHERE tags IS NOT NULL
		ORDER BY tag`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// ================================================
// CONFIG OPERATIONS
// ================================================

// GetYouTubeConfig retrieves the YouTube configuration
func GetYouTubeConfig(db *database.DB) (*YouTubeConfig, error) {
	query := `
		SELECT id, channel_id, channel_title, rss_url, sync_enabled, 
		       sync_schedule, auto_sync_to_master, last_sync_at, 
		       created_at, updated_at
		FROM youtube_config 
		ORDER BY id DESC 
		LIMIT 1`

	var config YouTubeConfig
	var lastSyncAt sql.NullTime

	err := db.QueryRow(query).Scan(
		&config.ID,
		&config.ChannelID,
		&config.ChannelTitle,
		&config.RSSURL,
		&config.SyncEnabled,
		&config.SyncSchedule,
		&config.AutoSyncToMaster,
		&lastSyncAt,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No config exists yet
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get YouTube config: %w", err)
	}

	if lastSyncAt.Valid {
		config.LastSyncAt = &lastSyncAt.Time
	}

	return &config, nil
}

// UpdateYouTubeConfig updates the YouTube configuration
func UpdateYouTubeConfig(db *database.DB, config *YouTubeConfig) error {
	// Check if config exists
	existing, err := GetYouTubeConfig(db)
	if err != nil {
		return err
	}

	if existing == nil {
		// Insert new config
		query := `
			INSERT INTO youtube_config (
				channel_id, channel_title, rss_url, sync_enabled, 
				sync_schedule, auto_sync_to_master, last_sync_at,
				created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
			RETURNING id`

		return db.QueryRow(query,
			config.ChannelID,
			config.ChannelTitle,
			config.RSSURL,
			config.SyncEnabled,
			config.SyncSchedule,
			config.AutoSyncToMaster,
			config.LastSyncAt,
		).Scan(&config.ID)
	}

	// Update existing config
	query := `
		UPDATE youtube_config 
		SET channel_id = $1, channel_title = $2, rss_url = $3, 
		    sync_enabled = $4, sync_schedule = $5, auto_sync_to_master = $6,
		    last_sync_at = $7, updated_at = NOW()
		WHERE id = $8`

	_, err = db.Exec(query,
		config.ChannelID,
		config.ChannelTitle,
		config.RSSURL,
		config.SyncEnabled,
		config.SyncSchedule,
		config.AutoSyncToMaster,
		config.LastSyncAt,
		existing.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update YouTube config: %w", err)
	}

	return nil
}

// UpdateLastSyncTime updates the last sync timestamp
func UpdateLastSyncTime(db *database.DB) error {
	query := `
		UPDATE youtube_config 
		SET last_sync_at = NOW(), updated_at = NOW()
		WHERE id = (SELECT id FROM youtube_config ORDER BY id DESC LIMIT 1)`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to update last sync time: %w", err)
	}

	return nil
}

// ================================================
// SYNC LOG OPERATIONS
// ================================================

// CreateSyncLog creates a new sync log entry
func CreateSyncLog(db *database.DB, log *YouTubeSyncLog) error {
	query := `
		INSERT INTO youtube_sync_log (
			sync_type, started_at, completed_at, videos_found, 
			videos_new, videos_updated, videos_skipped, 
			status, error_message, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		RETURNING id`

	return db.QueryRow(query,
		log.SyncType,
		log.StartedAt,
		log.CompletedAt,
		log.VideosFound,
		log.VideosNew,
		log.VideosUpdated,
		log.VideosSkipped,
		log.Status,
		log.ErrorMessage,
	).Scan(&log.ID)
}

// UpdateSyncLog updates an existing sync log entry
func UpdateSyncLog(db *database.DB, log *YouTubeSyncLog) error {
	query := `
		UPDATE youtube_sync_log 
		SET completed_at = $1, videos_found = $2, videos_new = $3, 
		    videos_updated = $4, videos_skipped = $5, status = $6, 
		    error_message = $7
		WHERE id = $8`

	_, err := db.Exec(query,
		log.CompletedAt,
		log.VideosFound,
		log.VideosNew,
		log.VideosUpdated,
		log.VideosSkipped,
		log.Status,
		log.ErrorMessage,
		log.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update sync log: %w", err)
	}

	return nil
}

// GetRecentSyncLogs retrieves the most recent sync logs
func GetRecentSyncLogs(db *database.DB, limit int) ([]*YouTubeSyncLog, error) {
	query := `
		SELECT id, sync_type, started_at, completed_at, videos_found,
		       videos_new, videos_updated, videos_skipped, status, 
		       COALESCE(error_message, '') as error_message, created_at
		FROM youtube_sync_log 
		ORDER BY started_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sync logs: %w", err)
	}
	defer rows.Close()

	var logs []*YouTubeSyncLog
	for rows.Next() {
		var log YouTubeSyncLog
		var completedAt sql.NullTime

		err := rows.Scan(
			&log.ID,
			&log.SyncType,
			&log.StartedAt,
			&completedAt,
			&log.VideosFound,
			&log.VideosNew,
			&log.VideosUpdated,
			&log.VideosSkipped,
			&log.Status,
			&log.ErrorMessage,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sync log: %w", err)
		}

		if completedAt.Valid {
			log.CompletedAt = &completedAt.Time
		}

		logs = append(logs, &log)
	}

	return logs, nil
}

// GetLatestSyncLog retrieves the most recent sync log
func GetLatestSyncLog(db *database.DB) (*YouTubeSyncLog, error) {
	logs, err := GetRecentSyncLogs(db, 1)
	if err != nil {
		return nil, err
	}

	if len(logs) == 0 {
		return nil, nil
	}

	return logs[0], nil
}
