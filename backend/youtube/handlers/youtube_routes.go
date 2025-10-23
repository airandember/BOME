package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"bome-backend/authentication/middleware"
	"bome-backend/infrastructure/database"
	"bome-backend/youtube/models"
	"bome-backend/youtube/services"

	"github.com/gin-gonic/gin"
)

// SetupYouTubeRoutes registers all YouTube routes and starts the scheduler
func SetupYouTubeRoutes(router *gin.RouterGroup, db *database.DB) *services.YouTubeScheduler {
	log.Printf("🎬 [YOUTUBE-ROUTES] Setting up YouTube routes...")

	youtubeService := services.NewYouTubeService(db)

	// Start the YouTube RSS scheduler for daily syncing at 2 PM MST
	youtubeScheduler := services.NewYouTubeScheduler(db)
	youtubeScheduler.Start()

	// API endpoints for frontend
	youtube := router.Group("/youtube")
	{
		// ================================================
		// PUBLIC ENDPOINTS (No Auth Required)
		// ================================================
		youtube.GET("/videos", getYouTubeVideos(youtubeService))
		youtube.GET("/videos/latest", getLatestYouTubeVideos(youtubeService))
		youtube.GET("/videos/search", searchYouTubeVideos(youtubeService))
		youtube.GET("/videos/category/:category", getYouTubeVideosByCategory(youtubeService))
		youtube.GET("/videos/:id", getYouTubeVideoByID(youtubeService))
		youtube.GET("/status", getYouTubeStatus(youtubeService))
		youtube.GET("/channel", getYouTubeChannelInfo(youtubeService))
		youtube.GET("/channel/stats", getYouTubeChannelStats(youtubeService))
		youtube.GET("/categories", getYouTubeCategories(youtubeService))
		youtube.GET("/tags", getYouTubeTags(youtubeService))

		// ================================================
		// ADMIN SYNC ENDPOINTS (Auth + Admin Required)
		// ================================================
		youtube.POST("/sync/rss", middleware.AuthRequired(), middleware.AdminRequired(), syncFromRSS(youtubeService))
		youtube.GET("/sync/status", middleware.AuthRequired(), middleware.AdminRequired(), getSyncStatus(youtubeService))

		// ================================================
		// SCHEDULER ENDPOINTS (Auth + Admin Required)
		// ================================================
		youtube.GET("/scheduler/status", middleware.AuthRequired(), middleware.AdminRequired(), getSchedulerStatus(youtubeScheduler))
		youtube.POST("/scheduler/trigger", middleware.AuthRequired(), middleware.AdminRequired(), triggerYouTubeSync(youtubeScheduler))

		// ================================================
		// CONFIGURATION ENDPOINTS (Auth + Admin Required)
		// ================================================
		youtube.GET("/config", middleware.AuthRequired(), middleware.AdminRequired(), getYouTubeConfig(youtubeService))
		youtube.POST("/config", middleware.AuthRequired(), middleware.AdminRequired(), updateYouTubeConfig(youtubeService, youtubeScheduler))
	}

	log.Printf("✅ [YOUTUBE-ROUTES] YouTube routes registered successfully (17 endpoints)")
	return youtubeScheduler
}

// ================================================
// PUBLIC ENDPOINTS
// ================================================

// getYouTubeVideos returns all YouTube videos with optional pagination
func getYouTubeVideos(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters
		limitStr := c.Query("limit")
		offsetStr := c.Query("offset")

		limit := 20 // default limit
		offset := 0

		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		if offsetStr != "" {
			if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
				offset = parsedOffset
			}
		}

		// Get videos from service
		response, err := youtubeService.GetAllVideos(limit, offset)
		if err != nil {
			log.Printf("❌ [YOUTUBE-ROUTES] Failed to get videos: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get YouTube videos"})
			return
		}

		// Return JSON response
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
			log.Printf("❌ [YOUTUBE-ROUTES] Failed to get latest videos: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get latest YouTube videos"})
			return
		}

		// Return JSON response
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
			log.Printf("❌ [YOUTUBE-ROUTES] Search failed for query '%s': %v", query, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search YouTube videos"})
			return
		}

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
			log.Printf("❌ [YOUTUBE-ROUTES] Failed to get videos by category '%s': %v", category, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get YouTube videos by category"})
			return
		}

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
			log.Printf("❌ [YOUTUBE-ROUTES] Failed to get video '%s': %v", videoID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get YouTube video"})
			return
		}

		c.JSON(http.StatusOK, video)
	}
}

// getYouTubeChannelInfo returns YouTube channel information
func getYouTubeChannelInfo(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		channelInfo, err := youtubeService.GetChannelInfo()
		if err != nil {
			log.Printf("❌ [YOUTUBE-ROUTES] Failed to get channel info: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get YouTube channel info"})
			return
		}

		c.JSON(http.StatusOK, channelInfo)
	}
}

// getYouTubeChannelStats returns channel statistics (subscribers, videos, views)
func getYouTubeChannelStats(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := youtubeService.GetChannelStats()
		if err != nil {
			log.Printf("❌ [YOUTUBE-ROUTES] Failed to get channel stats: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get channel stats"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"stats":   stats,
		})
	}
}

// getYouTubeStatus returns the current status of the YouTube integration
func getYouTubeStatus(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		status, err := youtubeService.GetStatus()
		if err != nil {
			log.Printf("❌ [YOUTUBE-ROUTES] Failed to get status: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get YouTube status"})
			return
		}

		c.JSON(http.StatusOK, status)
	}
}

// getYouTubeCategories returns all available video categories
func getYouTubeCategories(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		categories, err := youtubeService.GetCategories()
		if err != nil {
			log.Printf("❌ [YOUTUBE-ROUTES] Failed to get categories: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get YouTube categories"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"categories": categories,
			"count":      len(categories),
		})
	}
}

// getYouTubeTags returns all available video tags
func getYouTubeTags(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tags, err := youtubeService.GetTags()
		if err != nil {
			log.Printf("❌ [YOUTUBE-ROUTES] Failed to get tags: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get YouTube tags"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"tags":  tags,
			"count": len(tags),
		})
	}
}

// ================================================
// ADMIN SYNC ENDPOINTS
// ================================================

// syncFromRSS manually triggers a sync from the YouTube RSS feed
func syncFromRSS(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("🔧 [YOUTUBE-ROUTES] Manual RSS sync triggered by admin")

		result, err := youtubeService.SyncFromRSS()
		if err != nil {
			log.Printf("❌ [YOUTUBE-ROUTES] RSS sync failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to sync from RSS feed",
				"details": err.Error(),
			})
			return
		}

		log.Printf("✅ [YOUTUBE-ROUTES] RSS sync completed: %d new, %d updated, %d skipped",
			result.NewVideos, result.UpdatedVideos, result.SkippedVideos)

		c.JSON(http.StatusOK, gin.H{
			"success":        true,
			"message":        "RSS sync completed successfully",
			"total_fetched":  result.TotalFetched,
			"new_videos":     result.NewVideos,
			"updated_videos": result.UpdatedVideos,
			"skipped_videos": result.SkippedVideos,
			"errors":         result.Errors,
			"sync_time":      result.SyncTime,
			"duration":       result.Duration,
		})
	}
}

// getSyncStatus returns information about the current sync status
func getSyncStatus(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		status, err := youtubeService.GetSyncStatus()
		if err != nil {
			log.Printf("❌ [YOUTUBE-ROUTES] Failed to get sync status: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to get sync status",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, status)
	}
}

// ================================================
// SCHEDULER ENDPOINTS
// ================================================

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

// triggerYouTubeSync manually triggers an RSS sync via the scheduler
func triggerYouTubeSync(scheduler *services.YouTubeScheduler) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("🔧 [YOUTUBE-ROUTES] Manual sync triggered via scheduler")

		result, err := scheduler.TriggerManualSync()
		if err != nil {
			log.Printf("❌ [YOUTUBE-ROUTES] Manual sync failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to trigger manual sync",
				"details": err.Error(),
			})
			return
		}

		log.Printf("✅ [YOUTUBE-ROUTES] Manual sync completed: %d new, %d updated, %d skipped",
			result.NewVideos, result.UpdatedVideos, result.SkippedVideos)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Manual sync completed",
			"result":  result,
		})
	}
}

// ================================================
// CONFIGURATION ENDPOINTS
// ================================================

// getYouTubeConfig returns the current YouTube configuration
func getYouTubeConfig(youtubeService *services.YouTubeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		config, err := youtubeService.GetConfig()
		if err != nil {
			log.Printf("❌ [YOUTUBE-ROUTES] Failed to get config: %v", err)
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
		var config models.YouTubeConfig

		if err := c.ShouldBindJSON(&config); err != nil {
			log.Printf("❌ [YOUTUBE-ROUTES] Invalid config data: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid configuration data",
				"details": err.Error(),
			})
			return
		}

		// Validate required fields
		if config.ChannelID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "channel_id is required",
			})
			return
		}

		// Update configuration in service
		if err := youtubeService.UpdateConfig(&config); err != nil {
			log.Printf("❌ [YOUTUBE-ROUTES] Failed to update config: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to update configuration",
				"details": err.Error(),
			})
			return
		}

		// Update scheduler configuration
		if err := scheduler.UpdateConfiguration(&config); err != nil {
			log.Printf("⚠️  [YOUTUBE-ROUTES] Failed to update scheduler config: %v", err)
		}

		log.Printf("✅ [YOUTUBE-ROUTES] Configuration updated: Channel=%s, Enabled=%v", config.ChannelID, config.SyncEnabled)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Configuration updated successfully",
			"config":  config,
		})
	}
}
