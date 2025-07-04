package routes

import (
	"net/http"
	"strconv"

	"bome-backend/internal/database"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// GetVideosHandler handles video listing with pagination and filtering
func GetVideosHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		category := c.Query("category")
		status := c.DefaultQuery("status", "ready")

		if limit > 100 {
			limit = 100
		}

		videos, err := db.GetVideos(limit, offset, category, status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch videos"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"videos": videos})
	}
}

// GetVideoHandler handles single video retrieval
func GetVideoHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
			return
		}

		video, err := db.GetVideoByID(videoID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
			return
		}

		// Increment view count
		go db.IncrementViewCount(videoID)

		c.JSON(http.StatusOK, gin.H{"video": video})
	}
}

// StreamVideoHandler handles video streaming via Bunny.net
func StreamVideoHandler(db *database.DB, bunnyService *services.BunnyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
			return
		}

		video, err := db.GetVideoByID(videoID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
			return
		}

		if video.Status != "ready" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Video not ready for streaming"})
			return
		}

		// Get streaming URL from Bunny.net
		streamURL := bunnyService.GetStreamURL(video.BunnyVideoID)

		c.JSON(http.StatusOK, gin.H{"stream_url": streamURL})
	}
}

// GetCategoriesHandler handles video categories listing
func GetCategoriesHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		categories, err := db.GetVideoCategories()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"categories": categories})
	}
}

// SearchVideosHandler handles video search
func SearchVideosHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Search query required"})
			return
		}

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		if limit > 100 {
			limit = 100
		}

		videos, err := db.SearchVideos(query, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search videos"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"videos": videos, "query": query})
	}
}
