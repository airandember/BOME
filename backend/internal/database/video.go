package database

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Video represents a video in the system
type Video struct {
	ID                   int
	Title                string
	Description          string
	BunnyVideoID         string
	ThumbnailURL         string
	Duration             int
	FileSize             int64
	Status               string
	Category             string
	Tags                 []string
	ViewCount            int
	LikeCount            int
	CreatedBy            int
	ScheduledPublishDate *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time

	// Bunny.net play data
	PlayData      map[string]interface{} `json:"play_data,omitempty"`
	IframeSrc     string                 `json:"iframe_src,omitempty"`
	DirectPlayURL string                 `json:"direct_play_url,omitempty"`
	PlaybackURL   string                 `json:"playback_url,omitempty"`
	Resolutions   []string               `json:"resolutions,omitempty"`
}

// CreateVideo inserts a new video into the database
func (db *DB) CreateVideo(title, description, bunnyVideoID, thumbnailURL, category string, duration int, fileSize int64, tags []string, createdBy int) (*Video, error) {
	// Convert tags to JSON string
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %v", err)
	}

	var id int
	err = db.QueryRow(
		`INSERT INTO videos (title, description, bunny_video_id, thumbnail_url, duration, file_size, status, category, tags, created_by, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW()) RETURNING id`,
		title, description, bunnyVideoID, thumbnailURL, duration, fileSize, "processing", category, string(tagsJSON), createdBy,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return db.GetVideoByID(id)
}

// GetVideoByID retrieves a video by ID
func (db *DB) GetVideoByID(id int) (*Video, error) {
	video := &Video{}
	var tagsStr string
	err := db.QueryRow(
		`SELECT id, title, description, bunny_video_id, thumbnail_url, duration, file_size, status, category, tags, view_count, like_count, created_by, created_at, updated_at FROM videos WHERE id = $1`,
		id,
	).Scan(&video.ID, &video.Title, &video.Description, &video.BunnyVideoID, &video.ThumbnailURL, &video.Duration, &video.FileSize, &video.Status, &video.Category, &tagsStr, &video.ViewCount, &video.LikeCount, &video.CreatedBy, &video.CreatedAt, &video.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Parse tags from JSON string
	if tagsStr != "" {
		if err := json.Unmarshal([]byte(tagsStr), &video.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %v", err)
		}
	}

	return video, nil
}

// GetVideoByBunnyID retrieves a video by Bunny video ID
func (db *DB) GetVideoByBunnyID(bunnyVideoID string) (*Video, error) {
	video := &Video{}
	var tagsStr string
	err := db.QueryRow(
		`SELECT id, title, description, bunny_video_id, thumbnail_url, duration, file_size, status, category, tags, view_count, like_count, created_by, created_at, updated_at FROM videos WHERE bunny_video_id = $1`,
		bunnyVideoID,
	).Scan(&video.ID, &video.Title, &video.Description, &video.BunnyVideoID, &video.ThumbnailURL, &video.Duration, &video.FileSize, &video.Status, &video.Category, &tagsStr, &video.ViewCount, &video.LikeCount, &video.CreatedBy, &video.CreatedAt, &video.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Parse tags from JSON string
	if tagsStr != "" {
		if err := json.Unmarshal([]byte(tagsStr), &video.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %v", err)
		}
	}

	return video, nil
}

// GetVideos retrieves videos with pagination and filtering
func (db *DB) GetVideos(limit, offset int, category, status string) ([]*Video, error) {
	query := `SELECT id, title, description, bunny_video_id, thumbnail_url, duration, file_size, status, category, tags, view_count, like_count, created_by, created_at, updated_at FROM videos WHERE 1=1`
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

	argCount++
	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d`, argCount)
	args = append(args, limit)

	argCount++
	query += fmt.Sprintf(` OFFSET $%d`, argCount)
	args = append(args, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []*Video
	for rows.Next() {
		video := &Video{}
		var tagsStr string
		err := rows.Scan(&video.ID, &video.Title, &video.Description, &video.BunnyVideoID, &video.ThumbnailURL, &video.Duration, &video.FileSize, &video.Status, &video.Category, &tagsStr, &video.ViewCount, &video.LikeCount, &video.CreatedBy, &video.CreatedAt, &video.UpdatedAt)
		if err != nil {
			return nil, err
		}
		// Parse tags from JSON string
		if tagsStr != "" {
			if err := json.Unmarshal([]byte(tagsStr), &video.Tags); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tags: %v", err)
			}
		}
		videos = append(videos, video)
	}

	return videos, nil
}

// UpdateVideoStatus updates a video's status
func (db *DB) UpdateVideoStatus(videoID int, status string) error {
	_, err := db.Exec(`UPDATE videos SET status = $1, updated_at = NOW() WHERE id = $2`, status, videoID)
	return err
}

// UpdateVideoViews updates a video's view count
func (db *DB) UpdateVideoViews(videoID int, views int) error {
	_, err := db.Exec(`UPDATE videos SET view_count = $1, updated_at = NOW() WHERE id = $2`, views, videoID)
	return err
}

// IncrementViewCount increments a video's view count
func (db *DB) IncrementViewCount(videoID int) error {
	_, err := db.Exec(`UPDATE videos SET view_count = view_count + 1, updated_at = NOW() WHERE id = $1`, videoID)
	return err
}

// GetVideoCategories retrieves all video categories
func (db *DB) GetVideoCategories() ([]string, error) {
	rows, err := db.Query(`SELECT DISTINCT category FROM videos WHERE category IS NOT NULL AND category != '' ORDER BY category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// SearchVideos searches videos by title and description
func (db *DB) SearchVideos(query string, limit, offset int) ([]*Video, error) {
	searchQuery := `%` + query + `%`
	rows, err := db.Query(
		`SELECT id, title, description, bunny_video_id, thumbnail_url, duration, file_size, status, category, tags, view_count, like_count, created_by, created_at, updated_at FROM videos WHERE (title ILIKE $1 OR description ILIKE $1) AND status = 'ready' ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		searchQuery, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []*Video
	for rows.Next() {
		video := &Video{}
		err := rows.Scan(&video.ID, &video.Title, &video.Description, &video.BunnyVideoID, &video.ThumbnailURL, &video.Duration, &video.FileSize, &video.Status, &video.Category, &video.Tags, &video.ViewCount, &video.LikeCount, &video.CreatedBy, &video.CreatedAt, &video.UpdatedAt)
		if err != nil {
			return nil, err
		}
		videos = append(videos, video)
	}

	return videos, nil
}

// UpdateVideo updates video details
func (db *DB) UpdateVideo(videoID int, updateData map[string]interface{}) error {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argCount := 0

	for field, value := range updateData {
		switch field {
		case "title", "description", "category", "status":
			argCount++
			setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argCount))
			args = append(args, value)
		case "tags":
			argCount++
			setParts = append(setParts, fmt.Sprintf("tags = $%d", argCount))
			args = append(args, value)
		}
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no valid fields to update")
	}

	argCount++
	setParts = append(setParts, fmt.Sprintf("updated_at = NOW()"))

	query := fmt.Sprintf("UPDATE videos SET %s WHERE id = $%d", strings.Join(setParts, ", "), argCount)
	args = append(args, videoID)

	_, err := db.Exec(query, args...)
	return err
}

// DeleteVideo deletes a video from the database
func (db *DB) DeleteVideo(videoID int) error {
	_, err := db.Exec(`DELETE FROM videos WHERE id = $1`, videoID)
	return err
}

// ScheduleVideo schedules a video to be published at a specific time
func (db *DB) ScheduleVideo(videoID int, publishDate time.Time) error {
	_, err := db.Exec(`UPDATE videos SET scheduled_publish_date = $1, status = 'scheduled', updated_at = NOW() WHERE id = $2`, publishDate, videoID)
	return err
}

// GetScheduledVideos retrieves videos scheduled to be published before the given time
func (db *DB) GetScheduledVideos(beforeTime time.Time) ([]*Video, error) {
	query := `SELECT id, title, description, bunny_video_id, thumbnail_url, duration, file_size, status, category, tags, view_count, like_count, created_by, scheduled_publish_date, created_at, updated_at FROM videos WHERE status = 'scheduled' AND scheduled_publish_date <= $1`

	rows, err := db.Query(query, beforeTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []*Video
	for rows.Next() {
		video := &Video{}
		err := rows.Scan(&video.ID, &video.Title, &video.Description, &video.BunnyVideoID, &video.ThumbnailURL, &video.Duration, &video.FileSize, &video.Status, &video.Category, &video.Tags, &video.ViewCount, &video.LikeCount, &video.CreatedBy, &video.ScheduledPublishDate, &video.CreatedAt, &video.UpdatedAt)
		if err != nil {
			return nil, err
		}
		videos = append(videos, video)
	}

	return videos, nil
}

// UnscheduleVideo removes the scheduled publish date and sets status back to draft
func (db *DB) UnscheduleVideo(videoID int) error {
	_, err := db.Exec(`UPDATE videos SET scheduled_publish_date = NULL, status = 'draft', updated_at = NOW() WHERE id = $1`, videoID)
	return err
}
