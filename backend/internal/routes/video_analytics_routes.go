package routes

import (
	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RegisterVideoAnalyticsRoutes registers all video analytics routes
func RegisterVideoAnalyticsRoutes(router *gin.RouterGroup, db *database.DB, redis *database.Redis) {
	log.Println("📊 [Routes] Registering Video Analytics routes...")

	// Initialize services with Redis support
	analyticsService := services.NewVideoAnalyticsService(db, redis)
	watchHistoryService := services.NewWatchHistoryService(db)

	// Initialize resilience middleware
	resilience := middleware.NewAnalyticsResilience()

	// Analytics routes group with resilience middleware
	analytics := router.Group("/analytics")
	analytics.Use(resilience.Middleware())
	{
		// Video tracking endpoint (authenticated users only - video page requires auth)
		analytics.POST("/video/track", middleware.AuthRequired(), func(c *gin.Context) {
			log.Printf("🌐 [ROUTE] ============================================")
			log.Printf("🌐 [ROUTE] Received POST /analytics/video/track")
			log.Printf("🌐 [ROUTE] Client IP: %s, User-Agent: %s", c.ClientIP(), c.GetHeader("User-Agent"))

			var req services.VideoTrackingRequest

			if err := c.ShouldBindJSON(&req); err != nil {
				log.Printf("❌ [ROUTE] Invalid JSON: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			log.Printf("📦 [ROUTE] Request payload: video_id=%d, duration=%ds, percentage=%.2f%%",
				req.VideoID, req.WatchedDuration, req.WatchedPercentage)

			// Get user ID from auth context (always present due to AuthRequired middleware)
			userID, exists := c.Get("user_id")
			if !exists {
				log.Printf("❌ [ROUTE] No user_id in context despite AuthRequired")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
				return
			}

			uid := userID.(int)
			req.UserID = &uid
			log.Printf("🔐 [ROUTE] Authenticated user: %d", uid)

			// Capture IP and User-Agent server-side
			req.IPAddress = c.ClientIP()
			req.UserAgent = c.GetHeader("User-Agent")
			log.Printf("📝 [ROUTE] Enhanced request with IP=%s, UA=%s", req.IPAddress, req.UserAgent[:50])

			// Check circuit breaker
			log.Printf("🛡️  [ROUTE] Checking circuit breaker...")
			if !resilience.ShouldAllowRequest() {
				// Circuit is open, drop this request (graceful degradation)
				log.Printf("⚠️ [ROUTE] Circuit breaker OPEN - throttling request")
				c.JSON(http.StatusOK, gin.H{
					"status":   "throttled",
					"video_id": req.VideoID,
					"message":  "Analytics temporarily throttled",
				})
				return
			}
			log.Printf("✅ [ROUTE] Circuit breaker OK - proceeding")

			// Record the view
			log.Printf("📤 [ROUTE→SERVICE] Calling analyticsService.RecordView()")
			if err := analyticsService.RecordView(req); err != nil {
				log.Printf("❌ [ROUTE] RecordView failed: %v", err)
				resilience.RecordFailure()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track view"})
				return
			}
			log.Printf("✅ [ROUTE←SERVICE] RecordView successful")

			// Record success
			resilience.RecordSuccess()

			// If user is authenticated, also update watch history
			if req.UserID != nil && req.WatchedDuration > 0 {
				log.Printf("📤 [ROUTE→HISTORY] Updating watch history for user %d", *req.UserID)
				if err := watchHistoryService.UpdateProgress(*req.UserID, req.VideoID, req.WatchedDuration); err != nil {
					log.Printf("⚠️  [ROUTE] Watch history update failed (non-fatal): %v", err)
					// Don't fail the request - tracking is more important than history
				} else {
					log.Printf("✅ [ROUTE←HISTORY] Watch history updated")
				}
			}

			log.Printf("📤 [ROUTE→FRONTEND] Sending 200 OK response")
			log.Printf("🌐 [ROUTE] ============================================")
			c.JSON(http.StatusOK, gin.H{
				"status":   "tracked",
				"video_id": req.VideoID,
			})
		})

		// Video statistics endpoint (authenticated only)
		analytics.GET("/video/:id/stats", middleware.AuthRequired(), func(c *gin.Context) {
			videoID, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
				return
			}

			period := c.DefaultQuery("period", "7d") // Default to 7 days

			stats, err := analyticsService.GetVideoStats(videoID, period)
			if err != nil {
				log.Printf("❌ [Video Analytics] Failed to get stats: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get video stats"})
				return
			}

			c.JSON(http.StatusOK, stats)
		})

		// Trending videos endpoint
		analytics.GET("/trending", func(c *gin.Context) {
			log.Printf("🌐 [ROUTE] ============================================")
			log.Printf("🌐 [ROUTE] Received GET /analytics/trending")
			log.Printf("🌐 [ROUTE] Client IP: %s", c.ClientIP())
			log.Printf("🌐 [ROUTE] User-Agent: %s", c.GetHeader("User-Agent"))
			log.Printf("🌐 [ROUTE] Query params: %s", c.Request.URL.RawQuery)

			limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
			if err != nil || limit <= 0 {
				log.Printf("🌐 [ROUTE] Invalid limit parameter, using default: 10")
				limit = 10
			}
			log.Printf("🌐 [ROUTE] Requested limit: %d", limit)

			log.Printf("📤 [ROUTE→SERVICE] Calling GetTrendingVideos()")
			trending, err := analyticsService.GetTrendingVideos(limit)
			if err != nil {
				log.Printf("❌ [ROUTE] GetTrendingVideos failed: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trending videos"})
				log.Printf("🌐 [ROUTE] ============================================")
				return
			}
			log.Printf("✅ [ROUTE←SERVICE] Received %d trending videos", len(trending))

			log.Printf("📤 [ROUTE→CLIENT] Sending JSON response")
			c.JSON(http.StatusOK, gin.H{
				"trending": trending,
				"count":    len(trending),
			})
			log.Printf("✅ [ROUTE] Response sent successfully")
			log.Printf("🌐 [ROUTE] ============================================")
		})

		// User engagement endpoint (authenticated only)
		analytics.GET("/user/engagement", middleware.AuthRequired(), func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
				return
			}

			days, err := strconv.Atoi(c.DefaultQuery("days", "30"))
			if err != nil || days <= 0 {
				days = 30
			}

			engagement, err := analyticsService.GetUserEngagement(userID.(int), days)
			if err != nil {
				log.Printf("❌ [Video Analytics] Failed to get engagement: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user engagement"})
				return
			}

			c.JSON(http.StatusOK, engagement)
		})

		// Top videos endpoint
		analytics.GET("/top", func(c *gin.Context) {
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
			days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

			if limit <= 0 {
				limit = 10
			}
			if days <= 0 {
				days = 30
			}

			videos, err := analyticsService.GetTopVideos(limit, days)
			if err != nil {
				log.Printf("❌ [Video Analytics] Failed to get top videos: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get top videos"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"videos": videos,
				"count":  len(videos),
				"period": days,
			})
		})
	}

	// Watch history routes (authenticated only)
	history := router.Group("/videos")
	history.Use(middleware.AuthRequired())
	{
		// Get watch history for a specific video
		history.GET("/:id/watch-history", func(c *gin.Context) {
			videoID, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
				return
			}

			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
				return
			}

			watchHistory, err := watchHistoryService.GetHistory(userID.(int), videoID)
			if err != nil {
				log.Printf("❌ [Watch History] Failed to get history: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get watch history"})
				return
			}

			if watchHistory == nil {
				c.JSON(http.StatusOK, gin.H{
					"last_position": 0,
					"completed":     false,
					"percentage":    0,
				})
				return
			}

			c.JSON(http.StatusOK, watchHistory)
		})

		// Mark video as complete
		history.POST("/:id/complete", func(c *gin.Context) {
			videoID, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
				return
			}

			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
				return
			}

			if err := watchHistoryService.MarkComplete(userID.(int), videoID); err != nil {
				log.Printf("❌ [Watch History] Failed to mark complete: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark video complete"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":    "completed",
				"video_id":  videoID,
				"completed": true,
			})
		})

		// Get continue watching list
		history.GET("/continue-watching", func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
				return
			}

			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
			if limit <= 0 {
				limit = 20
			}

			videos, err := watchHistoryService.GetContinueWatching(userID.(int), limit)
			if err != nil {
				log.Printf("❌ [Watch History] Failed to get continue watching: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get continue watching"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"videos": videos,
				"count":  len(videos),
			})
		})

		// Get completed videos
		history.GET("/completed", func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
				return
			}

			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
			offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

			if limit <= 0 {
				limit = 20
			}

			videos, err := watchHistoryService.GetCompletedVideos(userID.(int), limit, offset)
			if err != nil {
				log.Printf("❌ [Watch History] Failed to get completed videos: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get completed videos"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"videos": videos,
				"count":  len(videos),
				"limit":  limit,
				"offset": offset,
			})
		})

		// Get watch stats
		history.GET("/watch-stats", func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
				return
			}

			stats, err := watchHistoryService.GetWatchStats(userID.(int))
			if err != nil {
				log.Printf("❌ [Watch History] Failed to get stats: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get watch stats"})
				return
			}

			c.JSON(http.StatusOK, stats)
		})

		// Clear watch history for a video
		history.DELETE("/:id/watch-history", func(c *gin.Context) {
			videoID, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
				return
			}

			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
				return
			}

			if err := watchHistoryService.ClearHistory(userID.(int), videoID); err != nil {
				log.Printf("❌ [Watch History] Failed to clear history: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear watch history"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":   "cleared",
				"video_id": videoID,
			})
		})
	}

	// Health and monitoring endpoints
	analytics.GET("/health", func(c *gin.Context) {
		status := resilience.GetStatus()

		// Get buffer stats if available
		if analyticsService != nil {
			// Buffer stats would be added here if needed
		}

		c.JSON(http.StatusOK, gin.H{
			"status":     "healthy",
			"resilience": status,
		})
	})

	log.Println("✅ [Routes] Video Analytics routes registered successfully")
}
