package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// AdminStreamingStats represents comprehensive streaming statistics for admin
type AdminStreamingStats struct {
	TotalVideos      int     `json:"total_videos"`
	ReadyVideos      int     `json:"ready_videos"`
	ProcessingVideos int     `json:"processing_videos"`
	ErrorVideos      int     `json:"error_videos"`
	DraftVideos      int     `json:"draft_videos"`
	ScheduledVideos  int     `json:"scheduled_videos"`
	TotalStorage     int64   `json:"total_storage_bytes"`
	TotalDuration    int     `json:"total_duration_seconds"`
	TotalViews       int     `json:"total_views"`
	AverageFileSize  float64 `json:"average_file_size_mb"`
	ProcessingErrors int     `json:"processing_errors"`
	UploadQueueSize  int     `json:"upload_queue_size"`
	CDNUsage         float64 `json:"cdn_usage_gb"`
	BandwidthUsage   float64 `json:"bandwidth_usage_gb"`
	LastSyncTime     string  `json:"last_sync_time"`
	SyncStatus       string  `json:"sync_status"`
}

// AdminVideoResponse represents enhanced video data for admin management
type AdminVideoResponse struct {
	ID                   int        `json:"id"`
	Title                string     `json:"title"`
	Description          string     `json:"description"`
	BunnyVideoID         string     `json:"bunny_video_id"`
	ThumbnailURL         string     `json:"thumbnail_url"`
	Duration             int        `json:"duration"`
	FileSize             int64      `json:"file_size"`
	Status               string     `json:"status"`
	Category             string     `json:"category"`
	Tags                 []string   `json:"tags"`
	ViewCount            int        `json:"view_count"`
	LikeCount            int        `json:"like_count"`
	CreatedBy            int        `json:"created_by"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	ScheduledPublishDate *time.Time `json:"scheduled_publish_date,omitempty"`

	// Admin-specific fields
	ProcessingProgress int                    `json:"processing_progress"`
	ProcessingErrors   []string               `json:"processing_errors"`
	UploadStatus       string                 `json:"upload_status"`
	UploadProgress     int                    `json:"upload_progress"`
	FileFormat         string                 `json:"file_format"`
	Resolution         string                 `json:"resolution"`
	Bitrate            int                    `json:"bitrate"`
	Framerate          float64                `json:"framerate"`
	EncodingProfile    string                 `json:"encoding_profile"`
	StorageLocation    string                 `json:"storage_location"`
	CDNStatus          string                 `json:"cdn_status"`
	AccessControl      string                 `json:"access_control"`
	Monetization       string                 `json:"monetization"`
	Analytics          map[string]interface{} `json:"analytics"`

	// Bunny.net specific data
	BunnyData *services.BunnyVideo    `json:"bunny_data,omitempty"`
	PlayData  *services.VideoPlayData `json:"play_data,omitempty"`
}

// SetupAdminStreamingRoutes configures admin streaming routes
func SetupAdminStreamingRoutes(router *gin.RouterGroup, db *database.DB, bunnyService *services.BunnyService, adminCache *services.AdminCacheService) {
	fmt.Printf("🔥 ADMIN STREAMING: Starting SetupAdminStreamingRoutes function\n")

	// Test endpoint without authentication for debugging
	router.GET("/streaming/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "success",
			"message":   "Admin streaming test endpoint working",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Create admin streaming group with authentication and admin middleware
	adminStreaming := router.Group("/streaming")
	adminStreaming.Use(middleware.AuthRequired())
	adminStreaming.Use(middleware.AdminRequired())

	// Dashboard analytics endpoint
	adminStreaming.GET("/dashboard", func(c *gin.Context) {
		// Get real-time active users
		activeUsers, err := db.GetActiveUsersCount()
		if err != nil {
			fmt.Printf("Warning: Failed to get active users: %v\n", err)
			activeUsers = 0
		}

		// Get recent activity
		recentActivity, err := db.GetRecentActivity(10)
		if err != nil {
			fmt.Printf("Warning: Failed to get recent activity: %v\n", err)
			recentActivity = []*database.Activity{}
		}

		// Get view analytics
		viewAnalytics, err := db.GetViewAnalytics()
		if err != nil {
			fmt.Printf("Warning: Failed to get view analytics: %v\n", err)
			viewAnalytics = map[string]interface{}{
				"total_views": 0,
				"views_today": 0,
				"views_week":  0,
				"growth_rate": 0.0,
			}
		}

		// Get subscriber metrics
		subscriberMetrics, err := db.GetSubscriberMetrics()
		if err != nil {
			fmt.Printf("Warning: Failed to get subscriber metrics: %v\n", err)
			subscriberMetrics = map[string]interface{}{
				"total_subscribers":    0,
				"active_subscriptions": 0,
				"monthly_revenue":      0.0,
				"churn_rate":           0.0,
			}
		}

		// Get video stats
		videoStats, err := db.GetVideoStats()
		if err != nil {
			fmt.Printf("Warning: Failed to get video stats: %v\n", err)
			videoStats = map[string]interface{}{
				"total_videos":    0,
				"synced_videos":   0,
				"needs_attention": 0,
			}
		}

		response := gin.H{
			"status": "success",
			"data": gin.H{
				"active_users":       activeUsers,
				"recent_activity":    recentActivity,
				"view_analytics":     viewAnalytics,
				"subscriber_metrics": subscriberMetrics,
				"video_stats":        videoStats,
			},
		}

		c.JSON(http.StatusOK, response)
	})

	// Active users endpoint
	adminStreaming.GET("/active-users", func(c *gin.Context) {
		activeUsers, err := db.GetActiveUsersCount()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to get active users: %v", err),
			})
			return
		}

		// Get active users trend
		trend, err := db.GetActiveUsersTrend()
		if err != nil {
			fmt.Printf("Warning: Failed to get active users trend: %v\n", err)
			trend = []map[string]interface{}{}
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": gin.H{
				"active_users": activeUsers,
				"trend":        trend,
			},
		})
	})

	// Recent activity endpoint
	adminStreaming.GET("/recent-activity", func(c *gin.Context) {
		limit := 20
		if limitStr := c.Query("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
				limit = l
			}
		}

		activity, err := db.GetRecentActivity(limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to get recent activity: %v", err),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   activity,
		})
	})

	// View analytics endpoint
	adminStreaming.GET("/view-analytics", func(c *gin.Context) {
		period := c.DefaultQuery("period", "7d")

		analytics, err := db.GetViewAnalyticsByPeriod(period)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to get view analytics: %v", err),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   analytics,
		})
	})

	// Subscriber metrics endpoint
	adminStreaming.GET("/subscriber-metrics", func(c *gin.Context) {
		metrics, err := db.GetSubscriberMetrics()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to get subscriber metrics: %v", err),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   metrics,
		})
	})

	// Enhanced admin video management
	streaming := router.Group("/streaming")
	{
		// Enhanced admin video management
		streaming.GET("/videos", middleware.AuthRequired(), middleware.AdminRequired(), middleware.SessionActivityTracker(db), GetAdminStreamingVideosHandler(db, bunnyService, adminCache))
		streaming.GET("/videos/:id", middleware.AuthRequired(), middleware.AdminRequired(), middleware.SessionActivityTracker(db), GetAdminStreamingVideoHandler(db, bunnyService, adminCache))
		streaming.PUT("/videos/:id", middleware.AuthRequired(), middleware.AdminRequired(), middleware.SessionActivityTracker(db), UpdateAdminStreamingVideoHandler(db, bunnyService, adminCache))
		streaming.DELETE("/videos/:id", middleware.AuthRequired(), middleware.AdminRequired(), middleware.SessionActivityTracker(db), DeleteAdminStreamingVideoHandler(db, bunnyService, adminCache))

		// Bulk operations
		streaming.POST("/videos/bulk", middleware.AuthRequired(), middleware.AdminRequired(), middleware.SessionActivityTracker(db), BulkAdminStreamingOperationHandler(db, bunnyService, adminCache))

		// Enhanced stats and analytics
		streaming.GET("/stats", middleware.AuthRequired(), middleware.AdminRequired(), middleware.SessionActivityTracker(db), GetAdminStreamingStatsHandler(db, bunnyService, adminCache))
		streaming.GET("/videos/:id/analytics", middleware.AuthRequired(), middleware.AdminRequired(), middleware.SessionActivityTracker(db), GetAdminVideoAnalyticsHandler(db, bunnyService, adminCache))

		// Processing and sync management
		streaming.POST("/sync", middleware.AuthRequired(), middleware.AdminRequired(), middleware.SessionActivityTracker(db), SyncAdminStreamingHandler(db, bunnyService, adminCache))
		streaming.GET("/sync/status", middleware.AuthRequired(), middleware.AdminRequired(), middleware.SessionActivityTracker(db), GetSyncStatusHandler(db, bunnyService, adminCache))
		streaming.POST("/videos/:id/retry-processing", middleware.AuthRequired(), middleware.AdminRequired(), middleware.SessionActivityTracker(db), RetryVideoProcessingHandler(db, bunnyService, adminCache))

		// Upload management
		streaming.GET("/uploads", middleware.AuthRequired(), middleware.AdminRequired(), middleware.SessionActivityTracker(db), GetUploadQueueHandler(db, bunnyService, adminCache))
		streaming.POST("/uploads/:id/cancel", middleware.AuthRequired(), middleware.AdminRequired(), middleware.SessionActivityTracker(db), CancelUploadHandler(db, bunnyService, adminCache))

		// CDN and storage management
		streaming.GET("/cdn/usage", middleware.AuthRequired(), middleware.AdminRequired(), middleware.SessionActivityTracker(db), GetCDNUsageHandler(db, bunnyService, adminCache))
		streaming.GET("/storage/usage", middleware.AuthRequired(), middleware.AdminRequired(), middleware.SessionActivityTracker(db), GetStorageUsageHandler(db, bunnyService, adminCache))

		// Quality and encoding management
		streaming.GET("/encoding/profiles", middleware.AuthRequired(), middleware.AdminRequired(), middleware.SessionActivityTracker(db), GetEncodingProfilesHandler(db, bunnyService, adminCache))
		streaming.POST("/videos/:id/re-encode", middleware.AuthRequired(), middleware.AdminRequired(), middleware.SessionActivityTracker(db), ReEncodeVideoHandler(db, bunnyService, adminCache))

		// Cache management
		streaming.GET("/cache/metrics", middleware.AuthRequired(), middleware.AdminRequired(), middleware.SessionActivityTracker(db), GetAdminCacheMetricsHandler(adminCache))
		streaming.POST("/cache/clear", middleware.AuthRequired(), middleware.AdminRequired(), middleware.SessionActivityTracker(db), ClearAdminCacheHandler(adminCache))
	}
}

// GetAdminStreamingVideosHandler returns enhanced video list for admin management
func GetAdminStreamingVideosHandler(db *database.DB, bunnyService *services.BunnyService, adminCache *services.AdminCacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters with admin-specific options
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		status := c.DefaultQuery("status", "") // Can be empty to get all statuses
		category := c.DefaultQuery("category", "")
		search := c.DefaultQuery("search", "")
		sortBy := c.DefaultQuery("sort", "created_at")
		sortOrder := c.DefaultQuery("order", "desc")
		includeProcessing := c.DefaultQuery("include_processing", "true") == "true"

		// Create cache key based on query parameters
		cacheKey := fmt.Sprintf("%s:%d:%d:%s:%s:%s:%s:%s:%t",
			services.AdminCacheKeyVideos, page, limit, status, category, search, sortBy, sortOrder, includeProcessing)

		// Try to get from cache first
		if cached, found := adminCache.Get(cacheKey); found {
			c.JSON(http.StatusOK, cached)
			return
		}

		// Validate parameters
		if limit > 100 {
			limit = 100
		}
		if limit < 1 {
			limit = 20
		}

		// Calculate offset
		offset := (page - 1) * limit

		// Get videos with admin-specific query
		videos, err := getAdminVideos(db, limit, offset, status, category, search, sortBy, sortOrder, includeProcessing)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch videos",
				"details": err.Error(),
			})
			return
		}

		// Enhance videos with Bunny.net data and admin metadata
		enhancedVideos, err := enhanceVideosWithAdminData(videos, bunnyService)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to enhance video data",
				"details": err.Error(),
			})
			return
		}

		// Get total count for pagination
		totalCount, err := getAdminVideoCount(db, status, category, search)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get video count",
				"details": err.Error(),
			})
			return
		}

		response := gin.H{
			"success": true,
			"videos":  enhancedVideos,
			"pagination": gin.H{
				"current_page": page,
				"per_page":     limit,
				"total":        totalCount,
				"total_pages":  (totalCount + limit - 1) / limit,
				"has_more":     page*limit < totalCount,
			},
			"filters": gin.H{
				"status":             status,
				"category":           category,
				"search":             search,
				"sort_by":            sortBy,
				"sort_order":         sortOrder,
				"include_processing": includeProcessing,
			},
		}

		// Cache the response
		adminCache.Set(cacheKey, response, services.AdminCacheTTLs.Videos, "videos")

		c.JSON(http.StatusOK, response)
	}
}

// GetAdminStreamingVideoHandler returns enhanced single video for admin management
func GetAdminStreamingVideoHandler(db *database.DB, bunnyService *services.BunnyService, adminCache *services.AdminCacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID := c.Param("id")
		if videoID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Video ID is required"})
			return
		}

		// Create cache key
		cacheKey := fmt.Sprintf("admin_video:%s", videoID)

		// Try to get from cache first
		if cached, found := adminCache.Get(cacheKey); found {
			c.JSON(http.StatusOK, cached)
			return
		}

		// Try to get video by numeric ID first
		videoIDInt, err := strconv.Atoi(videoID)
		var video *database.Video

		if err == nil {
			// It's a numeric ID, get from database
			video, err = db.GetVideoByID(videoIDInt)
		} else {
			// It's a Bunny GUID, get from database
			video, err = db.GetVideoByBunnyID(videoID)
		}

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Video not found",
				"details": err.Error(),
			})
			return
		}

		// Enhance video with admin data
		enhancedVideo, err := enhanceVideoWithAdminData(video, bunnyService)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to enhance video data",
				"details": err.Error(),
			})
			return
		}

		response := gin.H{
			"success": true,
			"video":   enhancedVideo,
		}

		// Cache the response
		adminCache.Set(cacheKey, response, services.AdminCacheTTLs.Videos, "video")

		c.JSON(http.StatusOK, response)
	}
}

// UpdateAdminStreamingVideoHandler updates video with admin-specific fields
func UpdateAdminStreamingVideoHandler(db *database.DB, bunnyService *services.BunnyService, adminCache *services.AdminCacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID := c.Param("id")
		if videoID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Video ID is required"})
			return
		}

		var updateData map[string]interface{}
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get video from database
		videoIDInt, err := strconv.Atoi(videoID)
		var video *database.Video

		if err == nil {
			video, err = db.GetVideoByID(videoIDInt)
		} else {
			video, err = db.GetVideoByBunnyID(videoID)
		}

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found in database"})
			return
		}

		// Update video in database
		if err := db.UpdateVideo(video.ID, updateData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update video"})
			return
		}

		// Invalidate related cache entries
		adminCache.Delete(fmt.Sprintf("admin_video:%s", videoID))
		adminCache.InvalidateByType("videos")

		// Log admin action
		adminID := c.GetInt("user_id")
		go db.CreateAdminLog(&adminID, "admin_video_updated", "video", &video.ID, updateData, c.ClientIP(), c.GetHeader("User-Agent"))

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Video updated successfully",
		})
	}
}

// DeleteAdminStreamingVideoHandler deletes video with enhanced cleanup
func DeleteAdminStreamingVideoHandler(db *database.DB, bunnyService *services.BunnyService, adminCache *services.AdminCacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID := c.Param("id")
		if videoID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Video ID is required"})
			return
		}

		// Get video from database
		videoIDInt, err := strconv.Atoi(videoID)
		var video *database.Video

		if err == nil {
			video, err = db.GetVideoByID(videoIDInt)
		} else {
			video, err = db.GetVideoByBunnyID(videoID)
		}

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found in database"})
			return
		}

		// Delete from Bunny.net if it has a Bunny ID
		if video.BunnyVideoID != "" {
			if err := bunnyService.DeleteVideo(video.BunnyVideoID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete video from Bunny.net"})
				return
			}
		}

		// Delete from database
		if err := db.DeleteVideo(video.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete video from database"})
			return
		}

		// Invalidate related cache entries
		adminCache.Delete(fmt.Sprintf("admin_video:%s", videoID))
		adminCache.InvalidateByType("videos")

		// Log admin action
		adminID := c.GetInt("user_id")
		go db.CreateAdminLog(&adminID, "admin_video_deleted", "video", &video.ID, map[string]interface{}{"title": video.Title}, c.ClientIP(), c.GetHeader("User-Agent"))

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Video deleted successfully",
		})
	}
}

// GetAdminStreamingStatsHandler returns comprehensive streaming statistics
func GetAdminStreamingStatsHandler(db *database.DB, bunnyService *services.BunnyService, adminCache *services.AdminCacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get from cache first
		if cached, found := adminCache.Get(services.AdminCacheKeyVideoStats); found {
			c.JSON(http.StatusOK, cached)
			return
		}

		stats, err := getAdminStreamingStats(db, bunnyService)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get streaming stats",
				"details": err.Error(),
			})
			return
		}

		response := gin.H{
			"success": true,
			"stats":   stats,
		}

		// Cache the response
		adminCache.Set(services.AdminCacheKeyVideoStats, response, services.AdminCacheTTLs.Stats, "stats")

		c.JSON(http.StatusOK, response)
	}
}

// Helper functions

func getAdminVideos(db *database.DB, limit, offset int, status, category, search, sortBy, sortOrder string, includeProcessing bool) ([]*database.Video, error) {
	// Build query with admin-specific filters
	query := `SELECT id, title, description, bunny_video_id, thumbnail_url, duration, file_size, status, category, tags, view_count, like_count, created_by, created_at, updated_at FROM videos WHERE 1=1`
	args := []interface{}{}
	argCount := 0

	// Add status filter (admin can see all statuses or filter by specific ones)
	if status != "" {
		argCount++
		query += fmt.Sprintf(` AND status = $%d`, argCount)
		args = append(args, status)
	} else if !includeProcessing {
		// Only show ready videos if not including processing
		argCount++
		query += fmt.Sprintf(` AND status = $%d`, argCount)
		args = append(args, "ready")
	}

	if category != "" {
		argCount++
		query += fmt.Sprintf(` AND category = $%d`, argCount)
		args = append(args, category)
	}

	if search != "" {
		argCount++
		query += fmt.Sprintf(` AND (title ILIKE $%d OR description ILIKE $%d)`, argCount, argCount)
		args = append(args, "%"+search+"%")
	}

	// Add sorting
	validSortFields := map[string]string{
		"created_at": "created_at",
		"updated_at": "updated_at",
		"title":      "title",
		"status":     "status",
		"duration":   "duration",
		"file_size":  "file_size",
		"view_count": "view_count",
	}

	sortField := validSortFields[sortBy]
	if sortField == "" {
		sortField = "created_at"
	}

	validOrders := map[string]string{
		"asc":  "ASC",
		"desc": "DESC",
	}

	order := validOrders[sortOrder]
	if order == "" {
		order = "DESC"
	}

	argCount++
	query += fmt.Sprintf(` ORDER BY %s %s LIMIT $%d`, sortField, order, argCount)
	args = append(args, limit)

	argCount++
	query += fmt.Sprintf(` OFFSET $%d`, argCount)
	args = append(args, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []*database.Video
	for rows.Next() {
		video := &database.Video{}
		var tagsStr string
		err := rows.Scan(&video.ID, &video.Title, &video.Description, &video.BunnyVideoID, &video.ThumbnailURL, &video.Duration, &video.FileSize, &video.Status, &video.Category, &tagsStr, &video.ViewCount, &video.LikeCount, &video.CreatedBy, &video.CreatedAt, &video.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// Parse tags from JSON string
		if tagsStr != "" {
			if err := json.Unmarshal([]byte(tagsStr), &video.Tags); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tags: %v", err)
			}
		}

		videos = append(videos, video)
	}

	return videos, nil
}

func getAdminVideoCount(db *database.DB, status, category, search string) (int, error) {
	query := `SELECT COUNT(*) FROM videos WHERE 1=1`
	args := []interface{}{}
	argCount := 0

	if status != "" {
		argCount++
		query += fmt.Sprintf(` AND status = $%d`, argCount)
		args = append(args, status)
	}

	if category != "" {
		argCount++
		query += fmt.Sprintf(` AND category = $%d`, argCount)
		args = append(args, category)
	}

	if search != "" {
		argCount++
		query += fmt.Sprintf(` AND (title ILIKE $%d OR description ILIKE $%d)`, argCount, argCount)
		args = append(args, "%"+search+"%")
	}

	var count int
	err := db.QueryRow(query, args...).Scan(&count)
	return count, err
}

func enhanceVideosWithAdminData(videos []*database.Video, bunnyService *services.BunnyService) ([]*AdminVideoResponse, error) {
	var enhancedVideos []*AdminVideoResponse

	for _, video := range videos {
		enhancedVideo, err := enhanceVideoWithAdminData(video, bunnyService)
		if err != nil {
			// Log error but continue with other videos
			fmt.Printf("Failed to enhance video %d: %v\n", video.ID, err)
			continue
		}
		enhancedVideos = append(enhancedVideos, enhancedVideo)
	}

	return enhancedVideos, nil
}

func enhanceVideoWithAdminData(video *database.Video, bunnyService *services.BunnyService) (*AdminVideoResponse, error) {
	enhanced := &AdminVideoResponse{
		ID:                   video.ID,
		Title:                video.Title,
		Description:          video.Description,
		BunnyVideoID:         video.BunnyVideoID,
		ThumbnailURL:         video.ThumbnailURL,
		Duration:             video.Duration,
		FileSize:             video.FileSize,
		Status:               video.Status,
		Category:             video.Category,
		Tags:                 video.Tags,
		ViewCount:            video.ViewCount,
		LikeCount:            video.LikeCount,
		CreatedBy:            video.CreatedBy,
		CreatedAt:            video.CreatedAt,
		UpdatedAt:            video.UpdatedAt,
		ScheduledPublishDate: video.ScheduledPublishDate,

		// Initialize admin-specific fields
		ProcessingProgress: 0,
		ProcessingErrors:   []string{},
		UploadStatus:       "unknown",
		UploadProgress:     0,
		FileFormat:         "mp4",
		Resolution:         "unknown",
		Bitrate:            0,
		Framerate:          0,
		EncodingProfile:    "default",
		StorageLocation:    "bunny-cdn",
		CDNStatus:          "active",
		AccessControl:      "public",
		Monetization:       "none",
		Analytics:          make(map[string]interface{}),
	}

	// Get Bunny.net data if available
	if video.BunnyVideoID != "" {
		bunnyVideo, err := bunnyService.GetVideo(video.BunnyVideoID)
		if err == nil {
			enhanced.BunnyData = bunnyVideo

			// Extract admin-specific data from Bunny.net
			enhanced.ProcessingProgress = bunnyVideo.EncodeProgress
			enhanced.Resolution = fmt.Sprintf("%dx%d", bunnyVideo.Width, bunnyVideo.Height)
			enhanced.Framerate = bunnyVideo.Framerate
			enhanced.FileFormat = "mp4" // Bunny.net default

			// Determine upload status based on Bunny.net status
			switch bunnyVideo.Status {
			case 0:
				enhanced.UploadStatus = "created"
			case 1:
				enhanced.UploadStatus = "uploaded"
			case 2:
				enhanced.UploadStatus = "processing"
			case 3:
				enhanced.UploadStatus = "transcoding"
			case 4:
				enhanced.UploadStatus = "ready"
			case 5:
				enhanced.UploadStatus = "error"
				enhanced.ProcessingErrors = append(enhanced.ProcessingErrors, "Bunny.net processing error")
			case 6:
				enhanced.UploadStatus = "upload_failed"
				enhanced.ProcessingErrors = append(enhanced.ProcessingErrors, "Upload failed")
			default:
				enhanced.UploadStatus = "unknown"
			}
		}

		// Get play data
		playData, err := bunnyService.GetVideoPlayData(video.BunnyVideoID)
		if err == nil {
			enhanced.PlayData = playData
		}
	}

	return enhanced, nil
}

func getAdminStreamingStats(db *database.DB, bunnyService *services.BunnyService) (*AdminStreamingStats, error) {
	// Get basic stats from database
	query := `
		SELECT 
			COUNT(*) as total_videos,
			COUNT(CASE WHEN status = 'ready' THEN 1 END) as ready_videos,
			COUNT(CASE WHEN status = 'processing' THEN 1 END) as processing_videos,
			COUNT(CASE WHEN status = 'error' THEN 1 END) as error_videos,
			COUNT(CASE WHEN status = 'draft' THEN 1 END) as draft_videos,
			COUNT(CASE WHEN status = 'scheduled' THEN 1 END) as scheduled_videos,
			COALESCE(SUM(file_size), 0) as total_storage,
			COALESCE(SUM(duration), 0) as total_duration,
			COALESCE(SUM(view_count), 0) as total_views,
			COALESCE(AVG(file_size), 0) as avg_file_size
		FROM videos
	`

	var stats AdminStreamingStats
	err := db.QueryRow(query).Scan(
		&stats.TotalVideos,
		&stats.ReadyVideos,
		&stats.ProcessingVideos,
		&stats.ErrorVideos,
		&stats.DraftVideos,
		&stats.ScheduledVideos,
		&stats.TotalStorage,
		&stats.TotalDuration,
		&stats.TotalViews,
		&stats.AverageFileSize,
	)

	if err != nil {
		return nil, err
	}

	// Convert average file size to MB
	stats.AverageFileSize = float64(stats.AverageFileSize) / (1024 * 1024)

	// Add mock CDN and bandwidth usage (in real implementation, get from Bunny.net API)
	stats.CDNUsage = float64(stats.TotalStorage) / (1024 * 1024 * 1024) * 0.1 // Estimate 10% of storage
	stats.BandwidthUsage = float64(stats.TotalViews) * 0.001                  // Estimate 1MB per view

	// Add sync status
	stats.LastSyncTime = time.Now().Format(time.RFC3339)
	stats.SyncStatus = "up_to_date"

	return &stats, nil
}

// Placeholder handlers for additional admin functionality

func BulkAdminStreamingOperationHandler(db *database.DB, bunnyService *services.BunnyService, adminCache *services.AdminCacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Bulk operation endpoint - to be implemented"})
	}
}

func GetAdminVideoAnalyticsHandler(db *database.DB, bunnyService *services.BunnyService, adminCache *services.AdminCacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Video analytics endpoint - to be implemented"})
	}
}

func SyncAdminStreamingHandler(db *database.DB, bunnyService *services.BunnyService, adminCache *services.AdminCacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Sync endpoint - to be implemented"})
	}
}

func GetSyncStatusHandler(db *database.DB, bunnyService *services.BunnyService, adminCache *services.AdminCacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Sync status endpoint - to be implemented"})
	}
}

func RetryVideoProcessingHandler(db *database.DB, bunnyService *services.BunnyService, adminCache *services.AdminCacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Retry processing endpoint - to be implemented"})
	}
}

func GetUploadQueueHandler(db *database.DB, bunnyService *services.BunnyService, adminCache *services.AdminCacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Upload queue endpoint - to be implemented"})
	}
}

func CancelUploadHandler(db *database.DB, bunnyService *services.BunnyService, adminCache *services.AdminCacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Cancel upload endpoint - to be implemented"})
	}
}

func GetCDNUsageHandler(db *database.DB, bunnyService *services.BunnyService, adminCache *services.AdminCacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "CDN usage endpoint - to be implemented"})
	}
}

func GetStorageUsageHandler(db *database.DB, bunnyService *services.BunnyService, adminCache *services.AdminCacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Storage usage endpoint - to be implemented"})
	}
}

func GetEncodingProfilesHandler(db *database.DB, bunnyService *services.BunnyService, adminCache *services.AdminCacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Encoding profiles endpoint - to be implemented"})
	}
}

func ReEncodeVideoHandler(db *database.DB, bunnyService *services.BunnyService, adminCache *services.AdminCacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Re-encode endpoint - to be implemented"})
	}
}

// Cache management handlers

func GetAdminCacheMetricsHandler(adminCache *services.AdminCacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics := adminCache.GetMetrics()
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"metrics": metrics,
		})
	}
}

func ClearAdminCacheHandler(adminCache *services.AdminCacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminCache.Clear()
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Admin cache cleared successfully",
		})
	}
}
