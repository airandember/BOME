package routes

import (
	"net/http"
	"strconv"
	"time"

	"bome-backend/internal/database"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupYouTubeRoutes registers all YouTube routes
func SetupYouTubeRoutes(router *gin.RouterGroup, db *database.DB) {
	youtubeService := services.NewYouTubeService(db)

	// Webhook endpoint for YouTube PubSubHubbub
	router.Any("/webhook/youtube", handleYouTubeWebhook(youtubeService))

	// API endpoints for frontend
	youtube := router.Group("/youtube")
	{
		youtube.GET("/videos", getYouTubeVideos(youtubeService))
		youtube.GET("/videos/latest", getLatestYouTubeVideos(youtubeService))
		youtube.POST("/subscribe", subscribeToChannel(youtubeService))
		youtube.POST("/unsubscribe", unsubscribeFromChannel(youtubeService))
		youtube.GET("/status", getYouTubeStatus(youtubeService))
	}
}

// handleYouTubeWebhook handles both verification and notification requests
func handleYouTubeWebhook(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case "GET":
			// Handle verification challenge
			youtubeService.HandleVerification(c.Writer, c.Request)
		case "POST":
			// Handle video notification
			youtubeService.HandleNotification(c.Writer, c.Request)
		default:
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
		}
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

// getLatestYouTubeVideos returns the latest YouTube videos (cached)
func getLatestYouTubeVideos(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to load from cache first
		response, err := youtubeService.LoadCachedVideos()
		if err != nil {
			// If cache fails, try to get from database
			response, err = youtubeService.GetLatestVideos(10)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get latest YouTube videos"})
				return
			}
		}

		// Return JSON response
		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, response)
	}
}

// subscribeToChannel manually triggers subscription to YouTube channel
func subscribeToChannel(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := youtubeService.Subscribe()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to subscribe to YouTube channel: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Successfully subscribed to YouTube channel updates",
		})
	}
}

// unsubscribeFromChannel manually triggers unsubscription from YouTube channel
func unsubscribeFromChannel(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := youtubeService.Unsubscribe()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to unsubscribe from YouTube channel: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Successfully unsubscribed from YouTube channel updates",
		})
	}
}

// getYouTubeStatus returns the current status of the YouTube integration
func getYouTubeStatus(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get video count and last update info
		response, err := youtubeService.LoadCachedVideos()
		if err != nil {
			response = &services.YouTubeVideosResponse{
				Videos:      []database.YouTubeVideo{},
				LastUpdated: time.Time{},
				TotalCount:  0,
			}
		}

		status := gin.H{
			"channel_id":     "UCHp1EBgpKytZt_-j72EZ83Q",
			"channel_name":   "@BookofMormonEvidence",
			"total_videos":   response.TotalCount,
			"last_updated":   response.LastUpdated,
			"webhook_active": true, // This could be determined by checking subscription status
		}

		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, status)
	}
}
