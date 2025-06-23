package database

import (
	"database/sql"
	"fmt"
	"time"
)

// YouTubeVideo represents a YouTube video in the database
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
	Tags         []string  `json:"tags"`
	Category     string    `json:"category"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// CreateYouTubeVideo creates a new YouTube video record
func (db *DB) CreateYouTubeVideo(video YouTubeVideo) error {
	query := `
		INSERT INTO youtube_videos (
			id, title, description, published_at, updated_at,
			thumbnail_url, video_url, embed_url, duration, view_count, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

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
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create YouTube video: %w", err)
	}

	return nil
}

// GetYouTubeVideoByID retrieves a YouTube video by its ID
func (db *DB) GetYouTubeVideoByID(id string) (*YouTubeVideo, error) {
	query := `
		SELECT id, title, description, published_at, updated_at,
		       thumbnail_url, video_url, embed_url, duration, view_count, created_at
		FROM youtube_videos 
		WHERE id = ?`

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
		&video.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Video not found
		}
		return nil, fmt.Errorf("failed to get YouTube video: %w", err)
	}

	return &video, nil
}

// UpdateYouTubeVideo updates an existing YouTube video record
func (db *DB) UpdateYouTubeVideo(video YouTubeVideo) error {
	query := `
		UPDATE youtube_videos 
		SET title = ?, description = ?, published_at = ?, updated_at = ?,
		    thumbnail_url = ?, video_url = ?, embed_url = ?, duration = ?, view_count = ?
		WHERE id = ?`

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
		video.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update YouTube video: %w", err)
	}

	return nil
}

// GetYouTubeVideos retrieves YouTube videos with optional limit
func (db *DB) GetYouTubeVideos(limit int) ([]YouTubeVideo, error) {
	query := `
		SELECT id, title, description, published_at, updated_at,
		       thumbnail_url, video_url, embed_url, duration, view_count, created_at
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

	var videos []YouTubeVideo
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
			&video.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan YouTube video: %w", err)
		}
		videos = append(videos, video)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate YouTube videos: %w", err)
	}

	return videos, nil
}

// DeleteYouTubeVideo deletes a YouTube video by ID
func (db *DB) DeleteYouTubeVideo(id string) error {
	query := `DELETE FROM youtube_videos WHERE id = ?`

	_, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete YouTube video: %w", err)
	}

	return nil
}

// GetYouTubeVideoCount returns the total count of YouTube videos
func (db *DB) GetYouTubeVideoCount() (int, error) {
	query := `SELECT COUNT(*) FROM youtube_videos`

	var count int
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get YouTube video count: %w", err)
	}

	return count, nil
}
