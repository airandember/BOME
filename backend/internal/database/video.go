package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Video represents a video in the system
type Video struct {
	ID                   int        `json:"id"`
	Title                string     `json:"title"`
	Description          string     `json:"description"`
	BunnyVideoID         string     `json:"bunnyVideoId"`
	ThumbnailURL         string     `json:"thumbnailUrl"`
	ThumbnailFileName    string     `json:"thumbnailFileName"`
	Duration             int        `json:"duration"`
	FileSize             int64      `json:"fileSize"`
	Status               string     `json:"status"`
	Category             string     `json:"category"`
	Tags                 []string   `json:"tags"`
	ViewCount            int        `json:"viewCount"`
	LikeCount            int        `json:"likeCount"`
	CreatedBy            int        `json:"createdBy"`
	ScheduledPublishDate *time.Time `json:"scheduledPublishDate,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
	Vid_Status           bool       `json:"vidStatus"`

	// Bunny.net play data
	PlayData      map[string]interface{} `json:"playData,omitempty"`
	IframeSrc     string                 `json:"iframeSrc,omitempty"`
	DirectPlayURL string                 `json:"directPlayUrl,omitempty"`
	PlaybackURL   string                 `json:"playbackUrl,omitempty"`
	Resolutions   []string               `json:"resolutions,omitempty"`
}

// CreateVideo inserts a new video into the database
func (db *DB) CreateVideo(title, description, bunnyVideoID, thumbnailURL, category string, duration int, fileSize int64, tags []string, createdBy int, vid_status bool) (*Video, error) {
	// Convert tags to JSON string
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %v", err)
	}

	var id int
	err = db.QueryRow(
		`INSERT INTO master_video_list (title, description, bunny_video_id, thumbnail_url, duration, file_size, status, category, tags, created_by, vid_status, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW()) RETURNING id`,
		title, description, bunnyVideoID, thumbnailURL, duration, fileSize, "processing", category, string(tagsJSON), createdBy, vid_status,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return db.GetVideoByID(id)
}

// GetVideoByID retrieves a video by ID
func (db *DB) GetVideoByID(id int) (*Video, error) {
	video := &Video{}
	var tagsStr sql.NullString  // Handle NULL tags for older videos
	var createdBy sql.NullInt64 // Handle NULL created_by for older videos
	
	err := db.QueryRow(
		`SELECT id, title, description, bunny_video_id, thumbnail_url, duration, file_size, status, category, tags, views, likes, created_by, created_at, updated_at, vid_status FROM master_video_list WHERE id = $1`,
		id,
	).Scan(&video.ID, &video.Title, &video.Description, &video.BunnyVideoID, &video.ThumbnailURL, &video.Duration, &video.FileSize, &video.Status, &video.Category, &tagsStr, &video.ViewCount, &video.LikeCount, &createdBy, &video.CreatedAt, &video.UpdatedAt, &video.Vid_Status)
	if err != nil {
		return nil, fmt.Errorf("GetVideoByID failed for id %d: %v", id, err)
	}

	// Handle NULL created_by
	if createdBy.Valid {
		video.CreatedBy = int(createdBy.Int64)
	} else {
		video.CreatedBy = 0 // Default to 0 if NULL
	}

	// Parse tags from JSON string (master_video_list uses JSONB for tags)
	// Handle both NULL, "null" string, and empty string
	if tagsStr.Valid && tagsStr.String != "" && tagsStr.String != "null" {
		if err := json.Unmarshal([]byte(tagsStr.String), &video.Tags); err != nil {
			// Don't fail the entire query if tags are malformed, just log it
			fmt.Printf("⚠️ Failed to unmarshal tags for video ID %d: %v (tags: %s)\n", id, err, tagsStr.String)
			video.Tags = []string{} // Set empty array as fallback
		}
	} else {
		video.Tags = []string{} // Default to empty array
	}

	return video, nil
}

// GetVideoByBunnyID retrieves a video by Bunny video ID
func (db *DB) GetVideoByBunnyID(bunnyVideoID string) (*Video, error) {
	video := &Video{}
	var tagsStr sql.NullString  // Handle NULL tags for older videos
	var createdBy sql.NullInt64 // Handle NULL created_by for older videos
	
	err := db.QueryRow(
		`SELECT id, title, description, bunny_video_id, thumbnail_url, duration, file_size, status, category, tags, views, likes, created_by, created_at, updated_at, vid_status FROM master_video_list WHERE bunny_video_id = $1`,
		bunnyVideoID,
	).Scan(&video.ID, &video.Title, &video.Description, &video.BunnyVideoID, &video.ThumbnailURL, &video.Duration, &video.FileSize, &video.Status, &video.Category, &tagsStr, &video.ViewCount, &video.LikeCount, &createdBy, &video.CreatedAt, &video.UpdatedAt, &video.Vid_Status)
	if err != nil {
		return nil, fmt.Errorf("GetVideoByBunnyID failed for %s: %v", bunnyVideoID, err)
	}

	// Handle NULL created_by
	if createdBy.Valid {
		video.CreatedBy = int(createdBy.Int64)
	} else {
		video.CreatedBy = 0 // Default to 0 if NULL
	}

	// Parse tags from JSON string
	// Handle both NULL, "null" string, and empty string
	if tagsStr.Valid && tagsStr.String != "" && tagsStr.String != "null" {
		if err := json.Unmarshal([]byte(tagsStr.String), &video.Tags); err != nil {
			// Don't fail the entire query if tags are malformed, just log it
			fmt.Printf("⚠️ Failed to unmarshal tags for video %s: %v (tags: %s)\n", bunnyVideoID, err, tagsStr.String)
			video.Tags = []string{} // Set empty array as fallback
		}
	} else {
		video.Tags = []string{} // Default to empty array
	}

	return video, nil
}

// GetVideos retrieves videos with pagination and filtering
func (db *DB) GetVideos(limit, offset int, category, status string) ([]*Video, error) {
	query := `SELECT id, title, description, bunny_video_id, thumbnail_url, duration, file_size, status, category, tags, views, likes, created_by, created_at, updated_at, vid_status FROM master_video_list WHERE 1=1`
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
		var tagsStr sql.NullString  // Handle NULL tags for older videos
		var createdBy sql.NullInt64 // Handle NULL created_by for older videos
		
		err := rows.Scan(&video.ID, &video.Title, &video.Description, &video.BunnyVideoID, &video.ThumbnailURL, &video.Duration, &video.FileSize, &video.Status, &video.Category, &tagsStr, &video.ViewCount, &video.LikeCount, &createdBy, &video.CreatedAt, &video.UpdatedAt, &video.Vid_Status)
		if err != nil {
			return nil, err
		}
		
		// Handle NULL created_by
		if createdBy.Valid {
			video.CreatedBy = int(createdBy.Int64)
		} else {
			video.CreatedBy = 0
		}
		
		// Parse tags from JSONB (master_video_list uses JSONB for tags)
		// Handle both NULL, "null" string, and empty string
		if tagsStr.Valid && tagsStr.String != "" && tagsStr.String != "null" {
			if err := json.Unmarshal([]byte(tagsStr.String), &video.Tags); err != nil {
				fmt.Printf("⚠️ Failed to unmarshal tags for video ID %d: %v (tags: %s)\n", video.ID, err, tagsStr.String)
				video.Tags = []string{}
			}
		} else {
			video.Tags = []string{}
		}
		videos = append(videos, video)
	}

	return videos, nil
}

// GetAllVideos retrieves all videos from the database for search index generation
func (db *DB) GetAllVideos() ([]Video, error) {
	query := `SELECT id, title, description, bunny_video_id, thumbnail_url, duration, file_size, status, category, tags, views, likes, created_by, created_at, updated_at FROM master_video_list ORDER BY created_at DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []Video
	for rows.Next() {
		video := Video{}
		var tagsStr sql.NullString // Use sql.NullString to handle NULL tags
		err := rows.Scan(&video.ID, &video.Title, &video.Description, &video.BunnyVideoID, &video.ThumbnailURL, &video.Duration, &video.FileSize, &video.Status, &video.Category, &tagsStr, &video.ViewCount, &video.LikeCount, &video.CreatedBy, &video.CreatedAt, &video.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// Set ThumbnailFileName to empty string since it's not in the database yet
		video.ThumbnailFileName = ""

		// Parse tags from JSONB (master_video_list uses JSONB for tags)
		// Handle NULL tags (videos that haven't been through the tagging system yet)
		if tagsStr.Valid && tagsStr.String != "" {
			if err := json.Unmarshal([]byte(tagsStr.String), &video.Tags); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tags: %v", err)
			}
		} else {
			// Videos that haven't been tagged yet get empty tags array
			video.Tags = []string{}
		}

		videos = append(videos, video)
	}

	return videos, nil
}

// UpdateVideoStatus updates a video's status
func (db *DB) UpdateVideoStatus(videoID int, status string) error {
	_, err := db.Exec(`UPDATE master_video_list SET status = $1, updated_at = NOW() WHERE id = $2`, status, videoID)
	return err
}

// UpdateVideoViews updates a video's view count
func (db *DB) UpdateVideoViews(videoID int, views int) error {
	_, err := db.Exec(`UPDATE master_video_list SET views = $1, updated_at = NOW() WHERE id = $2`, views, videoID)
	return err
}

// IncrementViewCount increments a video's view count
func (db *DB) IncrementViewCount(videoID int) error {
	_, err := db.Exec(`UPDATE master_video_list SET views = views + 1, updated_at = NOW() WHERE id = $1`, videoID)
	return err
}

// GetVideoCategories retrieves all video categories
func (db *DB) GetVideoCategories() ([]string, error) {
	rows, err := db.Query(`SELECT DISTINCT category FROM master_video_list WHERE category IS NOT NULL AND category != '' ORDER BY category`)
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
		`SELECT id, title, description, bunny_video_id, thumbnail_url, duration, file_size, status, category, tags, views, likes, created_by, created_at, updated_at, vid_status FROM master_video_list WHERE (title ILIKE $1 OR description ILIKE $1) AND status = 'ready' ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		searchQuery, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []*Video
	for rows.Next() {
		video := &Video{}
		var tagsStr sql.NullString  // Handle NULL tags for older videos
		var createdBy sql.NullInt64 // Handle NULL created_by for older videos
		
		err := rows.Scan(&video.ID, &video.Title, &video.Description, &video.BunnyVideoID, &video.ThumbnailURL, &video.Duration, &video.FileSize, &video.Status, &video.Category, &tagsStr, &video.ViewCount, &video.LikeCount, &createdBy, &video.CreatedAt, &video.UpdatedAt, &video.Vid_Status)
		if err != nil {
			return nil, err
		}
		
		// Handle NULL created_by
		if createdBy.Valid {
			video.CreatedBy = int(createdBy.Int64)
		} else {
			video.CreatedBy = 0
		}
		
		// Parse tags from JSONB
		// Handle both NULL, "null" string, and empty string
		if tagsStr.Valid && tagsStr.String != "" && tagsStr.String != "null" {
			if err := json.Unmarshal([]byte(tagsStr.String), &video.Tags); err != nil {
				fmt.Printf("⚠️ Failed to unmarshal tags for video ID %d: %v (tags: %s)\n", video.ID, err, tagsStr.String)
				video.Tags = []string{}
			}
		} else {
			video.Tags = []string{}
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
	setParts = append(setParts, "updated_at = NOW()")

	query := fmt.Sprintf("UPDATE master_video_list SET %s WHERE id = $%d", strings.Join(setParts, ", "), argCount)
	args = append(args, videoID)

	_, err := db.Exec(query, args...)
	return err
}

// DeleteVideo deletes a video from the database
func (db *DB) DeleteVideo(videoID int) error {
	_, err := db.Exec(`DELETE FROM master_video_list WHERE id = $1`, videoID)
	return err
}

// ScheduleVideo schedules a video to be published at a specific time
func (db *DB) ScheduleVideo(videoID int, publishDate time.Time) error {
	_, err := db.Exec(`UPDATE master_video_list SET scheduled_publish_date = $1, status = 'scheduled', updated_at = NOW() WHERE id = $2`, publishDate, videoID)
	return err
}

// GetScheduledVideos retrieves videos scheduled to be published before the given time
func (db *DB) GetScheduledVideos(beforeTime time.Time) ([]*Video, error) {
	query := `SELECT id, title, description, bunny_video_id, thumbnail_url, duration, file_size, status, category, tags, views, likes, created_by, scheduled_publish_date, created_at, updated_at, vid_status FROM master_video_list WHERE status = 'scheduled' AND scheduled_publish_date <= $1`

	rows, err := db.Query(query, beforeTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []*Video
	for rows.Next() {
		video := &Video{}
		err := rows.Scan(&video.ID, &video.Title, &video.Description, &video.BunnyVideoID, &video.ThumbnailURL, &video.Duration, &video.FileSize, &video.Status, &video.Category, &video.Tags, &video.ViewCount, &video.LikeCount, &video.CreatedBy, &video.ScheduledPublishDate, &video.CreatedAt, &video.UpdatedAt, &video.Vid_Status)
		if err != nil {
			return nil, err
		}
		videos = append(videos, video)
	}

	return videos, nil
}

// UnscheduleVideo removes the scheduled publish date and sets status back to draft
func (db *DB) UnscheduleVideo(videoID int) error {
	_, err := db.Exec(`UPDATE master_video_list SET scheduled_publish_date = NULL, status = 'draft', updated_at = NOW() WHERE id = $1`, videoID)
	return err
}
