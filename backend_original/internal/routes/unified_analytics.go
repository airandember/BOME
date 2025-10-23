package routes

import (
	"net/http"

	"bome-backend/internal/database"

	"github.com/gin-gonic/gin"
)

// UnifiedAnalyticsRequest represents the request for unified analytics
type UnifiedAnalyticsRequest struct {
	Period          string `json:"period"`           // 24h, 7d, 30d, 90d
	IncludeRealtime bool   `json:"include_realtime"` // Whether to include real-time data
	IncludeActivity bool   `json:"include_activity"` // Whether to include recent activity
}

// UnifiedAnalyticsResponse represents the unified analytics response
type UnifiedAnalyticsResponse struct {
	// Core metrics
	TotalUsers          int     `json:"total_users"`
	TotalVideos         int     `json:"total_videos"`
	TotalViews          int64   `json:"total_views"`
	TotalRevenue        float64 `json:"total_revenue"`
	ActiveSubscriptions int     `json:"active_subscriptions"`

	// Video-specific metrics
	VideoStats struct {
		TotalVideos        int            `json:"total_videos"`
		SyncedVideos       int            `json:"synced_videos"`
		NeedsAttention     int            `json:"needs_attention"`
		TotalViews         int64          `json:"total_views"`
		VideosByStatus     map[string]int `json:"videos_by_status"`
		VideosBySyncStatus map[string]int `json:"videos_by_sync_status"`
		PendingConflicts   int            `json:"pending_conflicts"`
	} `json:"video_stats"`

	// View analytics
	ViewAnalytics struct {
		TotalViews int64   `json:"total_views"`
		ViewsToday int64   `json:"views_today"`
		ViewsWeek  int64   `json:"views_week"`
		GrowthRate float64 `json:"growth_rate"`
	} `json:"view_analytics"`

	// Subscriber metrics
	SubscriberMetrics struct {
		TotalSubscribers    int     `json:"total_subscribers"`
		ActiveSubscriptions int     `json:"active_subscriptions"`
		MonthlyRevenue      float64 `json:"monthly_revenue"`
		ChurnRate           float64 `json:"churn_rate"`
	} `json:"subscriber_metrics"`

	// Real-time metrics (optional)
	RealTime *struct {
		ActiveUsers    int                  `json:"active_users"`
		RecentActivity []*database.Activity `json:"recent_activity"`
	} `json:"real_time,omitempty"`
}

// SetupUnifiedAnalyticsRoutes sets up the unified analytics endpoints
func SetupUnifiedAnalyticsRoutes(router *gin.RouterGroup, db *database.DB) {
	analytics := router.Group("/analytics")
	{
		// Unified analytics endpoint - consolidates all analytics in one call
		analytics.POST("/unified", func(c *gin.Context) {
			var req UnifiedAnalyticsRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}

			// Set default period if not provided
			if req.Period == "" {
				req.Period = "7d"
			}

			response, err := getUnifiedAnalytics(db, req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to fetch unified analytics",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status": "success",
				"data":   response,
			})
		})

		// Basic stats endpoint - for simple dashboards
		analytics.GET("/basic", func(c *gin.Context) {
			stats, err := getBasicStats(db)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to fetch basic stats",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status": "success",
				"data":   stats,
			})
		})

		// Video analytics endpoint - for streaming dashboard
		analytics.GET("/video", func(c *gin.Context) {
			stats, err := getVideoAnalytics(db)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to fetch video analytics",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status": "success",
				"data":   stats,
			})
		})
	}
}

// getUnifiedAnalytics consolidates all analytics data in a single efficient query
func getUnifiedAnalytics(db *database.DB, req UnifiedAnalyticsRequest) (*UnifiedAnalyticsResponse, error) {
	response := &UnifiedAnalyticsResponse{}

	// Get core metrics in a single query
	coreQuery := `
		SELECT 
			(SELECT COUNT(*) FROM users) as total_users,
			(SELECT COUNT(*) FROM master_video_list) as total_videos,
			(SELECT COALESCE(SUM(views), 0) FROM master_video_list) as total_views,
			(SELECT COALESCE(SUM(amount), 0) FROM subscriptions WHERE status = 'active') as total_revenue,
			(SELECT COUNT(*) FROM subscriptions WHERE status = 'active') as active_subscriptions
	`

	var totalUsers, totalVideos, activeSubscriptions int
	var totalViews int64
	var totalRevenue float64

	err := db.QueryRow(coreQuery).Scan(
		&totalUsers,
		&totalVideos,
		&totalViews,
		&totalRevenue,
		&activeSubscriptions,
	)
	if err != nil {
		return nil, err
	}

	response.TotalUsers = totalUsers
	response.TotalVideos = totalVideos
	response.TotalViews = totalViews
	response.TotalRevenue = totalRevenue
	response.ActiveSubscriptions = activeSubscriptions

	// Get video stats (consolidated from master video stats)
	videoStats, err := db.GetMasterVideoStats()
	if err != nil {
		return nil, err
	}

	response.VideoStats.TotalVideos = videoStats["total_videos"].(int)
	response.VideoStats.TotalViews = int64(videoStats["total_views"].(int))
	response.VideoStats.VideosByStatus = videoStats["videos_by_status"].(map[string]int)
	response.VideoStats.VideosBySyncStatus = videoStats["videos_by_sync_status"].(map[string]int)
	response.VideoStats.PendingConflicts = videoStats["pending_conflicts"].(int)

	// Calculate synced and needs attention videos
	response.VideoStats.SyncedVideos = response.VideoStats.VideosBySyncStatus["synced"]
	response.VideoStats.NeedsAttention = response.VideoStats.VideosBySyncStatus["needs_attention"]

	// Get view analytics (consolidated from view analytics)
	viewAnalytics, err := db.GetViewAnalytics()
	if err != nil {
		return nil, err
	}

	response.ViewAnalytics.TotalViews = viewAnalytics["total_views"].(int64)
	response.ViewAnalytics.ViewsToday = viewAnalytics["views_today"].(int64)
	response.ViewAnalytics.ViewsWeek = viewAnalytics["views_week"].(int64)
	response.ViewAnalytics.GrowthRate = viewAnalytics["growth_rate"].(float64)

	// Get subscriber metrics
	subscriberMetrics, err := db.GetSubscriberMetrics()
	if err != nil {
		return nil, err
	}

	response.SubscriberMetrics.TotalSubscribers = subscriberMetrics["total_subscribers"].(int)
	response.SubscriberMetrics.ActiveSubscriptions = subscriberMetrics["active_subscriptions"].(int)
	response.SubscriberMetrics.MonthlyRevenue = subscriberMetrics["monthly_revenue"].(float64)
	response.SubscriberMetrics.ChurnRate = subscriberMetrics["churn_rate"].(float64)

	// Include real-time data if requested
	if req.IncludeRealtime {
		activeUsers, err := db.GetActiveUsersCount()
		if err != nil {
			activeUsers = 0 // Don't fail the entire request for real-time data
		}

		response.RealTime = &struct {
			ActiveUsers    int                  `json:"active_users"`
			RecentActivity []*database.Activity `json:"recent_activity"`
		}{
			ActiveUsers: activeUsers,
		}

		// Include recent activity if requested
		if req.IncludeActivity {
			recentActivity, err := db.GetRecentActivity(10)
			if err != nil {
				recentActivity = []*database.Activity{} // Don't fail for activity data
			}
			response.RealTime.RecentActivity = recentActivity
		}
	}

	return response, nil
}

// getBasicStats returns only the essential metrics for simple dashboards
func getBasicStats(db *database.DB) (map[string]interface{}, error) {
	query := `
		SELECT 
			(SELECT COUNT(*) FROM users) as total_users,
			(SELECT COUNT(*) FROM master_video_list) as total_videos,
			(SELECT COALESCE(SUM(views), 0) FROM master_video_list) as total_views,
			(SELECT COALESCE(SUM(amount), 0) FROM subscriptions WHERE status = 'active') as total_revenue
	`

	var totalUsers, totalVideos int
	var totalViews int64
	var totalRevenue float64

	err := db.QueryRow(query).Scan(&totalUsers, &totalVideos, &totalViews, &totalRevenue)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_users":   totalUsers,
		"total_videos":  totalVideos,
		"total_views":   totalViews,
		"total_revenue": totalRevenue,
	}, nil
}

// getVideoAnalytics returns video-specific analytics for streaming dashboard
func getVideoAnalytics(db *database.DB) (map[string]interface{}, error) {
	// Get video stats
	videoStats, err := db.GetVideoStats()
	if err != nil {
		return nil, err
	}

	// Get view analytics
	viewAnalytics, err := db.GetViewAnalytics()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"video_stats":    videoStats,
		"view_analytics": viewAnalytics,
	}, nil
}
