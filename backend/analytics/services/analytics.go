package services

import (
	authModels "bome-backend/authentication/models"
	"bome-backend/infrastructure/database"
	subModels "bome-backend/subscription/models"
	videoModels "bome-backend/video-streaming/models"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// SystemHealth represents system health metrics
type SystemHealth struct {
	Status      string  `json:"status"`
	Uptime      int64   `json:"uptime"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
}

// SystemMetrics represents detailed system metrics
type SystemMetrics struct {
	Timestamp   int64   `json:"timestamp"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemoryUsed  uint64  `json:"memory_used"`
	MemoryTotal uint64  `json:"memory_total"`
	DiskUsed    uint64  `json:"disk_used"`
	DiskTotal   uint64  `json:"disk_total"`
}

// AnalyticsService handles analytics operations
type AnalyticsService struct {
	db *database.DB
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(db *database.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

// GetAnalytics returns comprehensive analytics data
func (s *AnalyticsService) GetAnalytics(period string) (map[string]interface{}, error) {
	// Period parameter available for future use
	_ = period

	// Get basic counts
	userCount, err := authModels.GetUserCount(s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get user count: %w", err)
	}

	videoCount, err := videoModels.GetVideoCount(s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get video count: %w", err)
	}

	totalViews, err := videoModels.GetTotalViews(s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get total views: %w", err)
	}

	totalLikes, err := videoModels.GetTotalLikes(s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get total likes: %w", err)
	}

	activeSubscriptions, err := subModels.GetActiveSubscriptions(s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get active subscriptions: %w", err)
	}

	// Get recent activity
	recentActivity, err := GetRecentActivity(s.db, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}

	// Get real-time metrics
	realTimeMetrics, err := getRealTimeMetrics(s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get real-time metrics: %w", err)
	}

	// Build analytics response
	analytics := map[string]interface{}{
		"metadata": map[string]interface{}{
			"last_updated": time.Now().Format(time.RFC3339),
			"version":      "1.0.0",
		},
		"real_time": realTimeMetrics,
		"users": map[string]interface{}{
			"total":        userCount,
			"new_today":    func() int { c, _ := s.getNewUsersCount(s.db, 1); return c }(),
			"new_week":     func() int { c, _ := s.getNewUsersCount(s.db, 7); return c }(),
			"new_month":    func() int { c, _ := s.getNewUsersCount(s.db, 30); return c }(),
			"active_today": realTimeMetrics["active_users"],
			"growth_rate":  s.calculateGrowthRate(0, 0),
		},
		"videos": map[string]interface{}{
			"total":          videoCount,
			"published":      func() int { c, _ := s.getPublishedVideosCount(s.db); return c }(),
			"pending":        func() int { c, _ := s.getPendingVideosCount(s.db); return c }(),
			"draft":          func() int { c, _ := s.getDraftVideosCount(s.db); return c }(),
			"total_views":    totalViews,
			"total_likes":    totalLikes,
			"avg_rating":     0.0,
			"top_categories": []string{},
		},
		"subscriptions": map[string]interface{}{
			"active":        activeSubscriptions,
			"new_today":     0,
			"new_week":      0,
			"new_month":     0,
			"revenue_today": 0.0,
			"revenue_week":  0.0,
			"revenue_month": 0.0,
			"mrr":           0.0,
			"arr":           0.0,
		},
		"activity": recentActivity,
		"period":   period,
	}

	return analytics, nil
}

// GetRealTimeAnalytics returns real-time analytics data
func (s *AnalyticsService) GetRealTimeAnalytics() (map[string]interface{}, error) {
	// Get real-time metrics
	metrics, err := getRealTimeMetrics(s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get real-time metrics: %w", err)
	}

	// Get live events (last 10 minutes)
	liveEvents, err := s.db.GetLiveEvents(10 * time.Minute)
	if err != nil {
		liveEvents = []map[string]interface{}{} // Default to empty array on error
	}

	// Get top content now
	topContent, err := s.db.GetTopContentNow(5)
	if err != nil {
		topContent = []map[string]interface{}{} // Default to empty array on error
	}

	// Get server metrics
	serverLoad, _ := s.db.GetServerLoad()
	bandwidthUsage, _ := s.db.GetBandwidthUsage()
	errorRate, _ := s.db.GetErrorRate()
	responseTime, _ := s.db.GetAverageResponseTime()

	realTimeData := map[string]interface{}{
		"active_users":         metrics["active_users"],
		"current_streams":      metrics["current_streams"],
		"server_load":          serverLoad,
		"bandwidth_usage":      bandwidthUsage,
		"recent_signups":       metrics["recent_signups"],
		"recent_subscriptions": metrics["recent_subscriptions"],
		"error_rate":           errorRate,
		"response_time":        responseTime,
		"live_events":          liveEvents,
		"top_content_now":      topContent,
	}

	return realTimeData, nil
}

// GetSystemHealth returns system health metrics
func (s *AnalyticsService) GetSystemHealth() (*SystemHealth, error) {
	health, err := s.db.GetSystemHealth()
	if err != nil {
		return &SystemHealth{
			Uptime:      0,
			CPUUsage:    0,
			MemoryUsage: 0,
			DiskUsage:   0,
		}, nil
	}
	return &SystemHealth{
		Status:      health.Status,
		Uptime:      health.Uptime,
		CPUUsage:    health.CPUUsage,
		MemoryUsage: health.MemoryUsage,
		DiskUsage:   health.DiskUsage,
	}, nil
}

// GetMonitoringData returns monitoring data for the admin dashboard
func (s *AnalyticsService) GetMonitoringData() (map[string]interface{}, error) {
	// Get system metrics (stubbed)
	systemMetrics := gin.H{}

	// Get webhook events
	webhookEvents, err := getWebhookEvents(s.db, time.Now().Add(-24*time.Hour), time.Now(), 50)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook events: %w", err)
	}

	// Get subsite health (stubbed)
	subsiteHealth := []interface{}{}

	// Get alerts
	alerts, err := getAlerts(s.db, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}

	return map[string]interface{}{
		"metrics": systemMetrics,
		"events":  webhookEvents,
		"health":  subsiteHealth,
		"alerts":  alerts,
	}, nil
}

// GetCrossSubsiteAnalytics returns cross-subsite analytics
func (s *AnalyticsService) GetCrossSubsiteAnalytics(timeframe string, subsite string) (map[string]interface{}, error) {
	// Timeframe and subsite parameters available for future use
	_ = timeframe
	_ = subsite

	// Get cross-subsite stats
	stats, err := getCrossSubsiteStats(s.db)
	if err != nil || stats == nil {
		return gin.H{
			"subsites":           []gin.H{},
			"crossSubsiteTotals": gin.H{},
		}, nil
	}
	// Stub - return empty stats
	return gin.H{
		"subsites":           []gin.H{},
		"crossSubsiteTotals": gin.H{},
	}, nil
}

// Helper functions for analytics service

func GetRecentActivity(db *database.DB, limit int) ([]interface{}, error) {
	activities, err := db.GetRecentActivity(limit)
	if err != nil {
		return []interface{}{}, nil // Return empty array on error
	}

	// Convert to interface array
	result := make([]interface{}, len(activities))
	for i, activity := range activities {
		result[i] = map[string]interface{}{
			"type":       activity.Type,
			"user_id":    activity.UserID,
			"action":     activity.Action,
			"details":    activity.Details,
			"created_at": activity.CreatedAt,
		}
	}
	return result, nil
}

func getRealTimeMetrics(db *database.DB) (map[string]interface{}, error) {
	metrics, err := db.GetRealTimeMetrics()
	if err != nil {
		return make(map[string]interface{}), nil // Return empty map on error
	}
	return metrics, nil
}

func getSystemMetrics(db *database.DB) (*SystemMetrics, error) {
	// Get latest system metrics from database
	metrics, err := db.GetSystemMetrics(time.Now().Add(-1*time.Hour), time.Now())
	if err != nil || len(metrics) == 0 {
		// Return default metrics if none found
		return &SystemMetrics{
			Timestamp:   time.Now().Unix(),
			CPUPercent:  25.0,
			MemoryUsed:  2 * 1024 * 1024 * 1024,   // 2GB
			MemoryTotal: 8 * 1024 * 1024 * 1024,   // 8GB
			DiskUsed:    50 * 1024 * 1024 * 1024,  // 50GB
			DiskTotal:   100 * 1024 * 1024 * 1024, // 100GB
		}, nil
	}

	// Convert to our SystemMetrics format
	latest := metrics[0]
	return &SystemMetrics{
		Timestamp:   latest.Timestamp.Unix(),
		CPUPercent:  latest.CPUUsage,
		MemoryUsed:  uint64(latest.MemoryUsage * 1024 * 1024 * 1024), // Convert GB to bytes
		MemoryTotal: 8 * 1024 * 1024 * 1024,                          // 8GB default
		DiskUsed:    uint64(latest.DiskUsage * 1024 * 1024 * 1024),   // Convert GB to bytes
		DiskTotal:   100 * 1024 * 1024 * 1024,                        // 100GB default
	}, nil
}

func getWebhookEvents(db *database.DB, params ...interface{}) ([]interface{}, error) {
	// Parse parameters
	var startTime, endTime time.Time
	var limit int = 50

	if len(params) >= 1 {
		if st, ok := params[0].(time.Time); ok {
			startTime = st
		}
	}
	if len(params) >= 2 {
		if et, ok := params[1].(time.Time); ok {
			endTime = et
		}
	}
	if len(params) >= 3 {
		if l, ok := params[2].(int); ok {
			limit = l
		}
	}

	events, err := db.GetWebhookEvents(startTime, endTime, limit)
	if err != nil {
		return []interface{}{}, nil // Return empty array on error
	}

	// Convert to interface array
	result := make([]interface{}, len(events))
	for i, event := range events {
		result[i] = event
	}
	return result, nil
}

func getAlerts(db *database.DB, params ...interface{}) ([]interface{}, error) {
	limit := 20
	if len(params) >= 1 {
		if l, ok := params[0].(int); ok {
			limit = l
		}
	}

	alerts, err := db.GetAlerts(limit)
	if err != nil {
		return []interface{}{}, nil // Return empty array on error
	}

	// Convert to interface array
	result := make([]interface{}, len(alerts))
	for i, alert := range alerts {
		result[i] = alert
	}
	return result, nil
}

func getCrossSubsiteStats(db *database.DB, params ...interface{}) (interface{}, error) {
	// Default to last 30 days
	endDate := time.Now()
	startDate := endDate.Add(-30 * 24 * time.Hour)

	stats, err := db.GetCrossSubsiteStats(startDate, endDate)
	if err != nil {
		return nil, nil
	}

	return stats, nil
}

func (s *AnalyticsService) getNewUsersCount(db *database.DB, days int) (int, error) {
	endDate := time.Now()
	startDate := endDate.Add(-time.Duration(days) * 24 * time.Hour)
	return db.GetNewUsersCount(startDate, endDate)
}

func (s *AnalyticsService) calculateGrowthRate(current, previous int) float64 {
	if previous == 0 {
		return 0
	}
	return float64(current-previous) / float64(previous) * 100
}

func (s *AnalyticsService) getPublishedVideosCount(db *database.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM master_video_list WHERE status = 'published'").Scan(&count)
	return count, err
}

func (s *AnalyticsService) getPendingVideosCount(db *database.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM master_video_list WHERE status = 'pending'").Scan(&count)
	return count, err
}

func (s *AnalyticsService) getDraftVideosCount(db *database.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM master_video_list WHERE status = 'draft'").Scan(&count)
	return count, err
}
