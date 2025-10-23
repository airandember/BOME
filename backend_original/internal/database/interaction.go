package database

import (
	"time"
)

// Comment represents a comment on a video
type Comment struct {
	ID        int
	VideoID   int
	UserID    int
	Content   string
	ParentID  *int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateComment inserts a new comment
func (db *DB) CreateComment(videoID, userID int, content string, parentID *int) (*Comment, error) {
	var id int
	err := db.QueryRow(
		`INSERT INTO comments (video_id, user_id, content, parent_id, created_at, updated_at) VALUES ($1, $2, $3, $4, NOW(), NOW()) RETURNING id`,
		videoID, userID, content, parentID,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return db.GetCommentByID(id)
}

// GetCommentByID retrieves a comment by ID
func (db *DB) GetCommentByID(id int) (*Comment, error) {
	comment := &Comment{}
	err := db.QueryRow(
		`SELECT id, video_id, user_id, content, parent_id, created_at, updated_at FROM comments WHERE id = $1`,
		id,
	).Scan(&comment.ID, &comment.VideoID, &comment.UserID, &comment.Content, &comment.ParentID, &comment.CreatedAt, &comment.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return comment, nil
}

// GetCommentsByVideoID retrieves comments for a video
func (db *DB) GetCommentsByVideoID(videoID, limit, offset int) ([]*Comment, error) {
	rows, err := db.Query(
		`SELECT id, video_id, user_id, content, parent_id, created_at, updated_at FROM comments WHERE video_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		videoID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		comment := &Comment{}
		err := rows.Scan(&comment.ID, &comment.VideoID, &comment.UserID, &comment.Content, &comment.ParentID, &comment.CreatedAt, &comment.UpdatedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

// LikeVideo adds a like to a video
func (db *DB) LikeVideo(userID, videoID int) error {
	_, err := db.Exec(
		`INSERT INTO likes (user_id, video_id, created_at) VALUES ($1, $2, NOW()) ON CONFLICT (user_id, video_id) DO NOTHING`,
		userID, videoID,
	)
	if err != nil {
		return err
	}

	// Update video like count
	_, err = db.Exec(`UPDATE videos SET like_count = (SELECT COUNT(*) FROM likes WHERE video_id = $1), updated_at = NOW() WHERE id = $1`, videoID)
	return err
}

// UnlikeVideo removes a like from a video
func (db *DB) UnlikeVideo(userID, videoID int) error {
	_, err := db.Exec(`DELETE FROM likes WHERE user_id = $1 AND video_id = $2`, userID, videoID)
	if err != nil {
		return err
	}

	// Update video like count
	_, err = db.Exec(`UPDATE videos SET like_count = (SELECT COUNT(*) FROM likes WHERE video_id = $1), updated_at = NOW() WHERE id = $1`, videoID)
	return err
}

// IsVideoLiked checks if a user has liked a video
func (db *DB) IsVideoLiked(userID, videoID int) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM likes WHERE user_id = $1 AND video_id = $2`, userID, videoID).Scan(&count)
	return count > 0, err
}

// FavoriteVideo adds a video to favorites
func (db *DB) FavoriteVideo(userID, videoID int) error {
	_, err := db.Exec(
		`INSERT INTO favorites (user_id, video_id, created_at) VALUES ($1, $2, NOW()) ON CONFLICT (user_id, video_id) DO NOTHING`,
		userID, videoID,
	)
	return err
}

// UnfavoriteVideo removes a video from favorites
func (db *DB) UnfavoriteVideo(userID, videoID int) error {
	_, err := db.Exec(`DELETE FROM favorites WHERE user_id = $1 AND video_id = $2`, userID, videoID)
	return err
}

// IsVideoFavorited checks if a user has favorited a video
func (db *DB) IsVideoFavorited(userID, videoID int) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM favorites WHERE user_id = $1 AND video_id = $2`, userID, videoID).Scan(&count)
	return count > 0, err
}

// GetUserFavorites retrieves a user's favorite videos
func (db *DB) GetUserFavorites(userID, limit, offset int) ([]*Video, error) {
	rows, err := db.Query(
		`SELECT v.id, v.title, v.description, v.bunny_video_id, v.thumbnail_url, v.duration, v.file_size, v.status, v.category, v.tags, v.view_count, v.like_count, v.created_by, v.created_at, v.updated_at 
		 FROM videos v 
		 JOIN favorites f ON v.id = f.video_id 
		 WHERE f.user_id = $1 AND v.status = 'ready' 
		 ORDER BY f.created_at DESC 
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
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

// RecordUserActivity records user activity
func (db *DB) RecordUserActivity(userID int, activityType string, videoID *int, metadata map[string]interface{}) error {
	_, err := db.Exec(
		`INSERT INTO user_activity (user_id, activity_type, video_id, metadata, created_at) VALUES ($1, $2, $3, $4, NOW())`,
		userID, activityType, videoID, metadata,
	)
	return err
}
