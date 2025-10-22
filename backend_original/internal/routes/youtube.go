package routes

import (
	"net/http"
	"strconv"
	"strings"

	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupYouTubeRoutes registers all YouTube routes and starts the scheduler
func SetupYouTubeRoutes(router *gin.RouterGroup, db *database.DB) {
	youtubeService := services.NewYouTubeService(db)

	// Start the YouTube RSS scheduler for daily syncing at 2 PM MST
	youtubeScheduler := services.NewYouTubeScheduler(db)
	youtubeScheduler.Start()

	// API endpoints for frontend
	youtube := router.Group("/youtube")
	{
		youtube.GET("/videos", getYouTubeVideos(youtubeService))
		youtube.GET("/videos/latest", getLatestYouTubeVideos(youtubeService))
		youtube.GET("/videos/rss-latest", getRSSLatestVideos(youtubeService)) // NEW: Fresh RSS data
		youtube.GET("/videos/search", searchYouTubeVideos(youtubeService))
		youtube.GET("/videos/category/:category", getYouTubeVideosByCategory(youtubeService))
		youtube.GET("/videos/:id", getYouTubeVideoByID(youtubeService))
		youtube.GET("/status", getYouTubeStatus(youtubeService))
		youtube.GET("/channel", getYouTubeChannelInfo(youtubeService))
		youtube.GET("/channel/stats", getYouTubeChannelStats(youtubeService))
		youtube.GET("/categories", getYouTubeCategories(youtubeService))
		youtube.GET("/tags", getYouTubeTags(youtubeService))

		// Admin sync endpoints
		youtube.POST("/sync/rss", middleware.AuthRequired(), middleware.AdminRequired(), syncFromRSS(youtubeService))
		youtube.POST("/sync/seed", middleware.AuthRequired(), middleware.AdminRequired(), seedFromMockData(youtubeService))
		youtube.GET("/sync/status", middleware.AuthRequired(), middleware.AdminRequired(), getSyncStatus(youtubeService))

		// Scheduler endpoints
		youtube.GET("/scheduler/status", middleware.AuthRequired(), middleware.AdminRequired(), getSchedulerStatus(youtubeScheduler))
		youtube.POST("/scheduler/trigger", middleware.AuthRequired(), middleware.AdminRequired(), triggerYouTubeSync(youtubeScheduler))

		// Configuration endpoints
		youtube.GET("/config", middleware.AuthRequired(), middleware.AdminRequired(), getYouTubeConfig(youtubeService))
		youtube.POST("/config", middleware.AuthRequired(), middleware.AdminRequired(), updateYouTubeConfig(youtubeService, youtubeScheduler))
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

// getRSSLatestVideos returns fresh videos directly from RSS feed (no database)
func getRSSLatestVideos(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse limit parameter
		limitStr := c.Query("limit")
		limit := 15 // default limit for RSS latest

		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		// Get fresh RSS data (bypass database)
		response, err := youtubeService.GetRSSLatestVideos(limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get RSS latest videos"})
			return
		}

		// Return JSON response
		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, response)
	}
}

// getYouTubeChannelStats returns channel statistics (subscribers, videos, views)
func getYouTubeChannelStats(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := youtubeService.GetChannelStats()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get channel stats"})
			return
		}

		// Return JSON response
		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"stats":   stats,
		})
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

// syncFromRSS manually triggers a sync from the YouTube RSS feed
func syncFromRSS(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := youtubeService.SyncFromRSS()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to sync from RSS feed",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":        "RSS sync completed successfully",
			"total_fetched":  result.TotalFetched,
			"new_videos":     result.NewVideos,
			"updated_videos": result.UpdatedVideos,
			"errors":         result.Errors,
			"sync_time":      result.SyncTime,
		})
	}
}

// seedFromMockData seeds the database with mock data
func seedFromMockData(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := youtubeService.SeedDatabaseFromMockData()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to seed database from mock data",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":        "Database seeded successfully",
			"total_fetched":  result.TotalFetched,
			"new_videos":     result.NewVideos,
			"updated_videos": result.UpdatedVideos,
			"errors":         result.Errors,
			"sync_time":      result.SyncTime,
		})
	}
}

// getSyncStatus returns information about the current sync status
func getSyncStatus(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		status, err := youtubeService.GetSyncStatus()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get sync status",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, status)
	}
}

// getSchedulerStatus returns the current scheduler status
func getSchedulerStatus(scheduler *services.YouTubeScheduler) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := scheduler.GetStatus()
		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"scheduler": status,
		})
	}
}

// triggerYouTubeSync manually triggers an RSS sync
func triggerYouTubeSync(scheduler *services.YouTubeScheduler) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := scheduler.TriggerManualSync()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to trigger manual sync",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Manual sync completed",
			"result":  result,
		})
	}
}

// getYouTubeConfig returns the current YouTube configuration
func getYouTubeConfig(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		config, err := youtubeService.GetConfiguration()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to get configuration",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"config":  config,
		})
	}
}

// updateYouTubeConfig updates the YouTube configuration
func updateYouTubeConfig(youtubeService *services.YouTubeService, scheduler *services.YouTubeScheduler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var configData struct {
			ChannelID       string `json:"channel_id" binding:"required"`
			SyncHour        int    `json:"sync_hour" binding:"min=0,max=23"`
			SyncMinute      int    `json:"sync_minute" binding:"min=0,max=59"`
			Timezone        string `json:"timezone"`
			AutoSyncEnabled bool   `json:"auto_sync_enabled"`
		}

		if err := c.ShouldBindJSON(&configData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid configuration data",
				"details": err.Error(),
			})
			return
		}

		// Update configuration
		err := youtubeService.UpdateConfiguration(configData.ChannelID, configData.SyncHour, configData.SyncMinute, configData.Timezone, configData.AutoSyncEnabled)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to update configuration",
				"details": err.Error(),
			})
			return
		}

		// Restart scheduler with new configuration if needed
		// This would require scheduler restart functionality

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Configuration updated successfully",
		})
	}
}
