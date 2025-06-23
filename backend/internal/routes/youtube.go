package routes

import (
	"net/http"
	"strconv"
	"strings"

	"bome-backend/internal/database"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupYouTubeRoutes registers all YouTube routes
func SetupYouTubeRoutes(router *gin.RouterGroup, db *database.DB) {
	youtubeService := services.NewYouTubeService(db)

	// API endpoints for frontend
	youtube := router.Group("/youtube")
	{
		youtube.GET("/videos", getYouTubeVideos(youtubeService))
		youtube.GET("/videos/latest", getLatestYouTubeVideos(youtubeService))
		youtube.GET("/videos/search", searchYouTubeVideos(youtubeService))
		youtube.GET("/videos/category/:category", getYouTubeVideosByCategory(youtubeService))
		youtube.GET("/videos/:id", getYouTubeVideoByID(youtubeService))
		youtube.GET("/status", getYouTubeStatus(youtubeService))
		youtube.GET("/channel", getYouTubeChannelInfo(youtubeService))
		youtube.GET("/categories", getYouTubeCategories(youtubeService))
		youtube.GET("/tags", getYouTubeTags(youtubeService))
	}
}

// getYouTubeVideos returns all YouTube videos with optional pagination
func getYouTubeVideos(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters
		limitStr := c.Query("limit")
		limit := 20 // default limit

		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		// Get videos from service
		response, err := youtubeService.GetLatestVideos(limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get YouTube videos"})
			return
		}

		// Return JSON response
		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, response)
	}
}

// getLatestYouTubeVideos returns the latest YouTube videos
func getLatestYouTubeVideos(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse limit parameter
		limitStr := c.Query("limit")
		limit := 10 // default limit for latest

		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		response, err := youtubeService.GetLatestVideos(limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get latest YouTube videos"})
			return
		}

		// Return JSON response
		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, response)
	}
}

// searchYouTubeVideos handles video search requests
func searchYouTubeVideos(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Search query parameter 'q' is required"})
			return
		}

		limitStr := c.Query("limit")
		limit := 20 // default limit

		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		response, err := youtubeService.SearchVideos(query, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search YouTube videos"})
			return
		}

		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, response)
	}
}

// getYouTubeVideosByCategory returns videos filtered by category
func getYouTubeVideosByCategory(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		category := c.Param("category")
		if category == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Category parameter is required"})
			return
		}

		limitStr := c.Query("limit")
		limit := 20 // default limit

		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		response, err := youtubeService.GetVideosByCategory(category, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get YouTube videos by category"})
			return
		}

		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, response)
	}
}

// getYouTubeVideoByID returns a specific video by ID
func getYouTubeVideoByID(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID := c.Param("id")
		if videoID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Video ID parameter is required"})
			return
		}

		video, err := youtubeService.GetVideoByID(videoID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get YouTube video"})
			return
		}

		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, video)
	}
}

// getYouTubeChannelInfo returns YouTube channel information
func getYouTubeChannelInfo(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		channelInfo, err := youtubeService.GetChannelInfo()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get YouTube channel info"})
			return
		}

		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, channelInfo)
	}
}

// getYouTubeStatus returns the current status of the YouTube integration
func getYouTubeStatus(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		status, err := youtubeService.GetStatus()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get YouTube status"})
			return
		}

		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, status)
	}
}

// getYouTubeCategories returns all available video categories
func getYouTubeCategories(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		categories, err := youtubeService.GetAllCategories()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get YouTube categories"})
			return
		}

		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, gin.H{
			"categories": categories,
			"count":      len(categories),
		})
	}
}

// getYouTubeTags returns all available video tags
func getYouTubeTags(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tags, err := youtubeService.GetAllTags()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get YouTube tags"})
			return
		}

		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, gin.H{
			"tags":  tags,
			"count": len(tags),
		})
	}
}
