package services

import (
	"bome-backend/internal/database"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// WatchHistoryService handles user video watch history and resume functionality
type WatchHistoryService struct {
	db *database.DB
}

// NewWatchHistoryService creates a new watch history service
func NewWatchHistoryService(db *database.DB) *WatchHistoryService {
	return &WatchHistoryService{db: db}
}

// WatchHistory represents a user's watch history for a video
type WatchHistory struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	VideoID         int       `json:"video_id"`
	LastPosition    int       `json:"last_position"`     // Seconds
	Completed       bool      `json:"completed"`
	FirstWatchedAt  time.Time `json:"first_watched_at"`
	LastWatchedAt   time.Time `json:"last_watched_at"`
	WatchPercentage float64   `json:"watch_percentage"` // Calculated field
}

// ContinueWatchingVideo represents a video in the "Continue Watching" list
type ContinueWatchingVideo struct {
	VideoID       int       `json:"video_id"`
	Title         string    `json:"title"`
	ThumbnailURL  string    `json:"thumbnail_url"`
	Duration      int       `json:"duration"`
	LastPosition  int       `json:"last_position"`
	Percentage    float64   `json:"percentage"`
	Completed     bool      `json:"completed"`
	LastWatchedAt time.Time `json:"last_watched_at"`
}

// UpdateProgress updates or creates a user's watch history for a video
func (s *WatchHistoryService) UpdateProgress(userID, videoID, position int) error {
	if userID == 0 || videoID == 0 {
		return fmt.Errorf("invalid userID or videoID")
	}

	log.Printf("📝 [Watch History] Updating progress: user=%d, video=%d, position=%ds", userID, videoID, position)

	query := `
		INSERT INTO watch_history (
			user_id, video_id, last_position, completed,
			first_watched_at, last_watched_at
		) VALUES ($1, $2, $3, false, NOW(), NOW())
		ON CONFLICT (user_id, video_id)
		DO UPDATE SET
			last_position = EXCLUDED.last_position,
			last_watched_at = NOW()
		WHERE watch_history.completed = false
	`

	_, err := s.db.Exec(query, userID, videoID, position)
	if err != nil {
		log.Printf("❌ [Watch History] Failed to update progress: %v", err)
		return fmt.Errorf("failed to update watch history: %w", err)
	}

	log.Printf("✅ [Watch History] Progress updated successfully")
	return nil
}

// MarkComplete marks a video as completed for a user
func (s *WatchHistoryService) MarkComplete(userID, videoID int) error {
	if userID == 0 || videoID == 0 {
		return fmt.Errorf("invalid userID or videoID")
	}

	log.Printf("✓ [Watch History] Marking complete: user=%d, video=%d", userID, videoID)

	query := `
		INSERT INTO watch_history (
			user_id, video_id, last_position, completed,
			first_watched_at, last_watched_at
		) VALUES ($1, $2, 0, true, NOW(), NOW())
		ON CONFLICT (user_id, video_id)
		DO UPDATE SET
			completed = true,
			last_watched_at = NOW()
	`

	_, err := s.db.Exec(query, userID, videoID)
	if err != nil {
		log.Printf("❌ [Watch History] Failed to mark complete: %v", err)
		return fmt.Errorf("failed to mark video complete: %w", err)
	}

	log.Printf("✅ [Watch History] Video marked as complete")
	return nil
}

// GetHistory retrieves a user's watch history for a specific video
func (s *WatchHistoryService) GetHistory(userID, videoID int) (*WatchHistory, error) {
	log.Printf("🔍 [Watch History] Getting history: user=%d, video=%d", userID, videoID)

	query := `
		SELECT 
			id, user_id, video_id, last_position, completed,
			first_watched_at, last_watched_at
		FROM watch_history
		WHERE user_id = $1 AND video_id = $2
	`

	var history WatchHistory
	err := s.db.QueryRow(query, userID, videoID).Scan(
		&history.ID,
		&history.UserID,
		&history.VideoID,
		&history.LastPosition,
		&history.Completed,
		&history.FirstWatchedAt,
		&history.LastWatchedAt,
	)

	if err == sql.ErrNoRows {
		log.Printf("ℹ️  [Watch History] No history found")
		return nil, nil // No history found (not an error)
	}

	if err != nil {
		log.Printf("❌ [Watch History] Failed to get history: %v", err)
		return nil, fmt.Errorf("failed to get watch history: %w", err)
	}

	// Calculate percentage (need video duration)
	var duration int
	durationQuery := `SELECT duration FROM master_video_list WHERE id = $1`
	if err := s.db.QueryRow(durationQuery, videoID).Scan(&duration); err == nil && duration > 0 {
		history.WatchPercentage = float64(history.LastPosition) / float64(duration) * 100
	}

	log.Printf("✅ [Watch History] History retrieved: position=%ds, completed=%v", history.LastPosition, history.Completed)
	return &history, nil
}

// GetContinueWatching retrieves a user's "Continue Watching" list
func (s *WatchHistoryService) GetContinueWatching(userID int, limit int) ([]ContinueWatchingVideo, error) {
	log.Printf("📺 [Watch History] Getting continue watching list for user %d", userID)

	if limit <= 0 || limit > 50 {
		limit = 20 // Default limit
	}

	query := `
		SELECT 
			v.id AS video_id,
			v.title,
			v.thumbnail_url,
			v.duration,
			wh.last_position,
			(wh.last_position::FLOAT / NULLIF(v.duration, 0)::FLOAT * 100) AS percentage,
			wh.last_watched_at
		FROM watch_history wh
		JOIN master_video_list v ON v.id = wh.video_id
		WHERE wh.user_id = $1
		AND wh.completed = false
		AND wh.last_position > 10
		AND wh.last_position < (v.duration - 30)
		AND v.status = 'active'
		ORDER BY wh.last_watched_at DESC
		LIMIT $2
	`

	rows, err := s.db.Query(query, userID, limit)
	if err != nil {
		log.Printf("❌ [Watch History] Failed to get continue watching: %v", err)
		return nil, fmt.Errorf("failed to get continue watching list: %w", err)
	}
	defer rows.Close()

	var videos []ContinueWatchingVideo
	for rows.Next() {
		var video ContinueWatchingVideo
		err := rows.Scan(
			&video.VideoID,
			&video.Title,
			&video.ThumbnailURL,
			&video.Duration,
			&video.LastPosition,
			&video.Percentage,
			&video.LastWatchedAt,
		)
		if err != nil {
			log.Printf("⚠️  [Watch History] Error scanning video: %v", err)
			continue
		}

		videos = append(videos, video)
	}

	log.Printf("✅ [Watch History] Found %d videos in continue watching", len(videos))
	return videos, nil
}

// GetCompletedVideos retrieves a user's completed videos
func (s *WatchHistoryService) GetCompletedVideos(userID int, limit int, offset int) ([]ContinueWatchingVideo, error) {
	log.Printf("✓ [Watch History] Getting completed videos for user %d", userID)

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `
		SELECT 
			v.id AS video_id,
			v.title,
			v.thumbnail_url,
			v.duration,
			wh.last_watched_at
		FROM watch_history wh
		JOIN master_video_list v ON v.id = wh.video_id
		WHERE wh.user_id = $1
		AND wh.completed = true
		AND v.status = 'active'
		ORDER BY wh.last_watched_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.Query(query, userID, limit, offset)
	if err != nil {
		log.Printf("❌ [Watch History] Failed to get completed videos: %v", err)
		return nil, fmt.Errorf("failed to get completed videos: %w", err)
	}
	defer rows.Close()

	var videos []ContinueWatchingVideo
	for rows.Next() {
		var video ContinueWatchingVideo
		video.Percentage = 100 // Completed videos are 100%
		video.Completed = true

		err := rows.Scan(
			&video.VideoID,
			&video.Title,
			&video.ThumbnailURL,
			&video.Duration,
			&video.LastWatchedAt,
		)
		if err != nil {
			log.Printf("⚠️  [Watch History] Error scanning video: %v", err)
			continue
		}

		videos = append(videos, video)
	}

	log.Printf("✅ [Watch History] Found %d completed videos", len(videos))
	return videos, nil
}

// GetWatchStats retrieves overall watch statistics for a user
func (s *WatchHistoryService) GetWatchStats(userID int) (map[string]interface{}, error) {
	log.Printf("📊 [Watch History] Getting watch stats for user %d", userID)

	query := `
		SELECT 
			COUNT(*) AS total_videos,
			COUNT(CASE WHEN completed = true THEN 1 END) AS completed_videos,
			COUNT(CASE WHEN completed = false THEN 1 END) AS in_progress_videos,
			MIN(first_watched_at) AS first_watch,
			MAX(last_watched_at) AS last_watch
		FROM watch_history
		WHERE user_id = $1
	`

	var totalVideos, completedVideos, inProgressVideos int
	var firstWatch, lastWatch sql.NullTime

	err := s.db.QueryRow(query, userID).Scan(
		&totalVideos,
		&completedVideos,
		&inProgressVideos,
		&firstWatch,
		&lastWatch,
	)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("❌ [Watch History] Failed to get watch stats: %v", err)
		return nil, fmt.Errorf("failed to get watch stats: %w", err)
	}

	stats := map[string]interface{}{
		"user_id":            userID,
		"total_videos":       totalVideos,
		"completed_videos":   completedVideos,
		"in_progress_videos": inProgressVideos,
		"completion_rate":    0.0,
	}

	if totalVideos > 0 {
		stats["completion_rate"] = float64(completedVideos) / float64(totalVideos) * 100
	}

	if firstWatch.Valid {
		stats["first_watch"] = firstWatch.Time
	}
	if lastWatch.Valid {
		stats["last_watch"] = lastWatch.Time
	}

	log.Printf("✅ [Watch History] Stats: %d total, %d completed", totalVideos, completedVideos)
	return stats, nil
}

// ClearHistory removes a specific video from a user's watch history
func (s *WatchHistoryService) ClearHistory(userID, videoID int) error {
	log.Printf("🗑️  [Watch History] Clearing history: user=%d, video=%d", userID, videoID)

	query := `DELETE FROM watch_history WHERE user_id = $1 AND video_id = $2`

	result, err := s.db.Exec(query, userID, videoID)
	if err != nil {
		log.Printf("❌ [Watch History] Failed to clear history: %v", err)
		return fmt.Errorf("failed to clear watch history: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		log.Printf("ℹ️  [Watch History] No history found to clear")
		return nil
	}

	log.Printf("✅ [Watch History] History cleared successfully")
	return nil
}

