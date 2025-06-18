package routes

import (
	"net/http"
	"strconv"

	"bome-backend/internal/database"

	"github.com/gin-gonic/gin"
)

// CommentRequest represents a comment creation payload
type CommentRequest struct {
	Content  string `json:"content" binding:"required"`
	ParentID *int   `json:"parent_id"`
}

// AddCommentHandler handles adding a comment to a video
func AddCommentHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
			return
		}

		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req CommentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		comment, err := db.CreateComment(videoID, userID, req.Content, req.ParentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
			return
		}

		// Record activity
		go db.RecordUserActivity(userID, "comment_added", &videoID, map[string]interface{}{"comment_id": comment.ID})

		c.JSON(http.StatusCreated, gin.H{"comment": comment})
	}
}

// GetCommentsHandler handles retrieving comments for a video
func GetCommentsHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
			return
		}

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		if limit > 100 {
			limit = 100
		}

		comments, err := db.GetCommentsByVideoID(videoID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"comments": comments})
	}
}

// LikeVideoHandler handles liking a video
func LikeVideoHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
			return
		}

		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		if err := db.LikeVideo(userID, videoID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to like video"})
			return
		}

		// Record activity
		go db.RecordUserActivity(userID, "video_liked", &videoID, nil)

		c.JSON(http.StatusOK, gin.H{"message": "Video liked successfully"})
	}
}

// UnlikeVideoHandler handles unliking a video
func UnlikeVideoHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
			return
		}

		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		if err := db.UnlikeVideo(userID, videoID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlike video"})
			return
		}

		// Record activity
		go db.RecordUserActivity(userID, "video_unliked", &videoID, nil)

		c.JSON(http.StatusOK, gin.H{"message": "Video unliked successfully"})
	}
}

// FavoriteVideoHandler handles favoriting a video
func FavoriteVideoHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
			return
		}

		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		if err := db.FavoriteVideo(userID, videoID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to favorite video"})
			return
		}

		// Record activity
		go db.RecordUserActivity(userID, "video_favorited", &videoID, nil)

		c.JSON(http.StatusOK, gin.H{"message": "Video favorited successfully"})
	}
}

// UnfavoriteVideoHandler handles unfavoriting a video
func UnfavoriteVideoHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
			return
		}

		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		if err := db.UnfavoriteVideo(userID, videoID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unfavorite video"})
			return
		}

		// Record activity
		go db.RecordUserActivity(userID, "video_unfavorited", &videoID, nil)

		c.JSON(http.StatusOK, gin.H{"message": "Video unfavorited successfully"})
	}
}

// GetFavoritesHandler handles retrieving user favorites
func GetFavoritesHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		if limit > 100 {
			limit = 100
		}

		videos, err := db.GetUserFavorites(userID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch favorites"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"favorites": videos})
	}
}
