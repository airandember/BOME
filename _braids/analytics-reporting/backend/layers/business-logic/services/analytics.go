package services

import (
	"bome-backend/internal/database"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

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
	// Calculate date range based on period
	endDate := time.Now()
	var startDate time.Time

	switch period {
	case "24h":
		startDate = endDate.Add(-24 * time.Hour)
	case "7d":
		startDate = endDate.Add(-7 * 24 * time.Hour)
	case "30d":
		startDate = endDate.Add(-30 * 24 * time.Hour)
	case "90d":
		startDate = endDate.Add(-90 * 24 * time.Hour)
	default:
		startDate = endDate.Add(-7 * 24 * time.Hour) // Default to 7 days
	}

	// Get basic counts
	userCount, err := s.db.GetUserCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get user count: %w", err)
	}

	videoCount, err := s.db.GetVideoCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get video count: %w", err)
	}

	totalViews, err := s.db.GetTotalViews()
	if err != nil {
		return nil, fmt.Errorf("failed to get total views: %w", err)
	}

	totalLikes, err := s.db.GetTotalLikes()
	if err != nil {
		return nil, fmt.Errorf("failed to get total likes: %w", err)
	}

	activeSubscriptions, err := s.db.GetActiveSubscriptions()
	if err != nil {
		return nil, fmt.Errorf("failed to get active subscriptions: %w", err)
	}

	// Get recent activity
	recentActivity, err := s.db.GetRecentActivity(10)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}

	// Get real-time metrics
	realTimeMetrics, err := s.db.GetRealTimeMetrics()
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
			"new_today":    s.getNewUsersCount(24 * time.Hour),
			"new_week":     s.getNewUsersCount(7 * 24 * time.Hour),
			"new_month":    s.getNewUsersCount(30 * 24 * time.Hour),
			"active_today": realTimeMetrics["active_users"],
			"growth_rate":  s.calculateGrowthRate("users", startDate, endDate),
		},
		"videos": map[string]interface{}{
			"total":          videoCount,
			"published":      s.getPublishedVideosCount(),
			"pending":        s.getPendingVideosCount(),
			"draft":          s.getDraftVideosCount(),
			"total_views":    totalViews,
			"total_likes":    totalLikes,
			"avg_rating":     s.getAverageVideoRating(),
			"top_categories": s.getTopVideoCategories(),
		},
		"subscriptions": map[string]interface{}{
			"active":        activeSubscriptions,
			"new_today":     s.getNewSubscriptionsCount(24 * time.Hour),
			"new_week":      s.getNewSubscriptionsCount(7 * 24 * time.Hour),
			"new_month":     s.getNewSubscriptionsCount(30 * 24 * time.Hour),
			"revenue_today": s.getRevenueForPeriod(24 * time.Hour),
			"revenue_week":  s.getRevenueForPeriod(7 * 24 * time.Hour),
			"revenue_month": s.getRevenueForPeriod(30 * 24 * time.Hour),
			"mrr":           s.calculateMRR(),
			"arr":           s.calculateARR(),
		},
		"activity": recentActivity,
		"period":   period,
	}

	return analytics, nil
}

// GetRealTimeAnalytics returns real-time analytics data
func (s *AnalyticsService) GetRealTimeAnalytics() (map[string]interface{}, error) {
	// Get real-time metrics
	metrics, err := s.db.GetRealTimeMetrics()
	if err != nil {
		return nil, fmt.Errorf("failed to get real-time metrics: %w", err)
	}

	// Get live events (last 10 minutes)
	liveEvents, err := s.getLiveEvents(10 * time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to get live events: %w", err)
	}

	// Get top content now
	topContent, err := s.getTopContentNow()
	if err != nil {
		return nil, fmt.Errorf("failed to get top content: %w", err)
	}

	realTimeData := map[string]interface{}{
		"active_users":         metrics["active_users"],
		"current_streams":      metrics["current_streams"],
		"server_load":          s.getServerLoad(),
		"bandwidth_usage":      s.getBandwidthUsage(),
		"recent_signups":       metrics["recent_signups"],
		"recent_subscriptions": metrics["recent_subscriptions"],
		"error_rate":           s.getErrorRate(),
		"response_time":        s.getResponseTime(),
		"live_events":          liveEvents,
		"top_content_now":      topContent,
	}

	return realTimeData, nil
}

// GetSystemHealth returns system health metrics
func (s *AnalyticsService) GetSystemHealth() (*database.SystemHealth, error) {
	// Get latest system metrics
	metrics, err := s.db.GetSystemMetrics(time.Now().Add(-1*time.Hour), time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get system metrics: %w", err)
	}

	if len(metrics) == 0 {
		// Return default health if no metrics available
		return &database.SystemHealth{
			Uptime:             "0 minutes",
			ResponseTime:       "0ms",
			ErrorRate:          "0%",
			StorageUsed:        "0 GB",
			BandwidthUsed:      "0 MB/s",
			CDNHits:            "0",
			DatabaseSize:       "0 GB",
			ActiveSessions:     0,
			LastWrite:          time.Now().Format(time.RFC3339),
			TotalEventsTracked: 0,
		}, nil
	}

	latest := metrics[0]

	// Calculate uptime (simplified - in production this would come from system monitoring)
	uptime := "5 days 12 hours" // This would be calculated from system start time

	// Format response time
	responseTime := fmt.Sprintf("%dms", latest.ResponseTime)

	// Format error rate
	errorRate := fmt.Sprintf("%.2f%%", latest.ErrorRate*100)

	// Format storage used
	storageUsed := fmt.Sprintf("%.1f GB", float64(latest.DatabaseSize)/(1024*1024*1024))

	// Format bandwidth used
	bandwidthUsed := fmt.Sprintf("%.1f MB/s", float64(latest.NetworkOut)/(1024*1024))

	// Get total events tracked
	totalEvents, err := s.getTotalEventsTracked()
	if err != nil {
		totalEvents = 0
	}

	return &database.SystemHealth{
		Uptime:             uptime,
		ResponseTime:       responseTime,
		ErrorRate:          errorRate,
		StorageUsed:        storageUsed,
		BandwidthUsed:      bandwidthUsed,
		CDNHits:            "145,230", // This would come from CDN API
		DatabaseSize:       storageUsed,
		ActiveSessions:     latest.ActiveSessions,
		LastWrite:          time.Now().Format(time.RFC3339),
		TotalEventsTracked: totalEvents,
	}, nil
}

// GetMonitoringData returns monitoring data for the admin dashboard
func (s *AnalyticsService) GetMonitoringData() (map[string]interface{}, error) {
	// Get system metrics
	systemMetrics, err := s.getSystemMetrics()
	if err != nil {
		return nil, fmt.Errorf("failed to get system metrics: %w", err)
	}

	// Get webhook events
	webhookEvents, err := s.db.GetWebhookEvents(time.Now().Add(-24*time.Hour), time.Now(), 50)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook events: %w", err)
	}

	// Get subsite health
	subsiteHealth, err := s.getSubsiteHealth()
	if err != nil {
		return nil, fmt.Errorf("failed to get subsite health: %w", err)
	}

	// Get alerts
	alerts, err := s.db.GetAlerts(20)
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
	// Calculate date range
	endDate := time.Now()
	var startDate time.Time

	switch timeframe {
	case "24h":
		startDate = endDate.Add(-24 * time.Hour)
	case "7d":
		startDate = endDate.Add(-7 * 24 * time.Hour)
	case "30d":
		startDate = endDate.Add(-30 * 24 * time.Hour)
	case "90d":
		startDate = endDate.Add(-90 * 24 * time.Hour)
	default:
		startDate = endDate.Add(-24 * time.Hour)
	}

	// Get cross-subsite stats
	stats, err := s.db.GetCrossSubsiteStats(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get cross-subsite stats: %w", err)
	}

	// Filter by subsite if specified
	if subsite != "all" {
		filteredStats := make([]*database.CrossSubsiteStats, 0)
		for _, stat := range stats {
			if stat.Subsite == subsite {
				filteredStats = append(filteredStats, stat)
			}
		}
		stats = filteredStats
	}

	// Aggregate stats by subsite
	aggregatedStats := make(map[string]map[string]interface{})

	for _, stat := range stats {
		if _, exists := aggregatedStats[stat.Subsite]; !exists {
			aggregatedStats[stat.Subsite] = map[string]interface{}{
				"users":           0,
				"content":         0,
				"views":           0,
				"revenue":         0.0,
				"engagement_rate": 0.0,
			}
		}

		subsiteData := aggregatedStats[stat.Subsite]
		subsiteData["users"] = subsiteData["users"].(int) + stat.Users
		subsiteData["content"] = subsiteData["content"].(int) + stat.Content
		subsiteData["views"] = subsiteData["views"].(int) + stat.Views
		subsiteData["revenue"] = subsiteData["revenue"].(float64) + stat.Revenue
		// For engagement rate, we'll use the latest value
		subsiteData["engagement_rate"] = stat.EngagementRate
	}

	return map[string]interface{}{
		"stats":     aggregatedStats,
		"timeframe": timeframe,
		"subsite":   subsite,
	}, nil
}

// GetWebhookAnalytics returns webhook analytics
func (s *AnalyticsService) GetWebhookAnalytics(timeframe string) (map[string]interface{}, error) {
	// Calculate date range
	endDate := time.Now()
	var startDate time.Time

	switch timeframe {
	case "24h":
		startDate = endDate.Add(-24 * time.Hour)
	case "7d":
		startDate = endDate.Add(-7 * 24 * time.Hour)
	case "30d":
		startDate = endDate.Add(-30 * 24 * time.Hour)
	case "90d":
		startDate = endDate.Add(-90 * 24 * time.Hour)
	default:
		startDate = endDate.Add(-24 * time.Hour)
	}

	// Get webhook events
	events, err := s.db.GetWebhookEvents(startDate, endDate, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook events: %w", err)
	}

	// Calculate analytics
	totalEvents := len(events)
	successCount := 0
	failedCount := 0
	totalResponseTime := 0
	eventsBySubsite := make(map[string]int)
	eventsByType := make(map[string]int)
	recentFailures := make([]map[string]interface{}, 0)

	for _, event := range events {
		// Count by status
		if event.Status == "success" {
			successCount++
		} else if event.Status == "failed" {
			failedCount++
			// Add to recent failures if within last hour
			if event.CreatedAt.After(time.Now().Add(-1 * time.Hour)) {
				recentFailures = append(recentFailures, map[string]interface{}{
					"timestamp":  event.CreatedAt.Format(time.RFC3339),
					"event_type": event.EventType,
					"subsite":    event.Subsite,
					"error":      event.ErrorMessage,
				})
			}
		}

		// Sum response times
		totalResponseTime += event.ResponseTime

		// Count by subsite
		eventsBySubsite[event.Subsite]++

		// Count by type
		eventsByType[event.EventType]++
	}

	// Calculate success rate
	successRate := 0.0
	if totalEvents > 0 {
		successRate = float64(successCount) / float64(totalEvents)
	}

	// Calculate average response time
	avgResponseTime := 0
	if totalEvents > 0 {
		avgResponseTime = totalResponseTime / totalEvents
	}

	return map[string]interface{}{
		"total_events":      totalEvents,
		"success_rate":      successRate,
		"avg_response_time": avgResponseTime,
		"events_by_subsite": eventsBySubsite,
		"events_by_type":    eventsByType,
		"recent_failures":   recentFailures,
	}, nil
}

// TrackEvent records an analytics event with enhanced error handling and logging
func (s *AnalyticsService) TrackEvent(eventType string, userID *int, sessionID string, subsite string, eventData map[string]interface{}, ipAddress string, userAgent string) error {
	// Log the event attempt
	s.logAnalyticsEvent("track_attempt", map[string]interface{}{
		"event_type": eventType,
		"user_id":    userID,
		"session_id": sessionID,
		"subsite":    subsite,
		"ip_address": ipAddress,
	})

	// Validate the event data
	if err := s.ValidateEvent(eventData); err != nil {
		s.logAnalyticsError("validation_failed", err, map[string]interface{}{
			"event_type": eventType,
			"user_id":    userID,
			"ip_address": ipAddress,
		})
		return fmt.Errorf("event validation failed: %w", err)
	}

	// Sanitize the event data
	sanitizedData := s.SanitizeEvent(eventData)

	// Convert event data to JSON
	eventDataJSON, err := json.Marshal(sanitizedData)
	if err != nil {
		s.logAnalyticsError("json_marshal_failed", err, map[string]interface{}{
			"event_type": eventType,
			"user_id":    userID,
		})
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	event := &database.AnalyticsEvent{
		EventType: eventType,
		UserID:    userID,
		SessionID: sessionID,
		Subsite:   subsite,
		EventData: string(eventDataJSON),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
	}

	// Track the event in database
	if err := s.db.TrackAnalyticsEvent(event); err != nil {
		s.logAnalyticsError("database_track_failed", err, map[string]interface{}{
			"event_type": eventType,
			"user_id":    userID,
			"ip_address": ipAddress,
		})

		// Create alert for critical database failures
		if strings.Contains(err.Error(), "connection") || strings.Contains(err.Error(), "timeout") {
			s.createAlert("critical", "Analytics Database Failure",
				fmt.Sprintf("Failed to track analytics event: %v", err))
		}

		return fmt.Errorf("failed to track analytics event: %w", err)
	}

	// Log successful tracking
	s.logAnalyticsEvent("track_success", map[string]interface{}{
		"event_type": eventType,
		"user_id":    userID,
		"event_id":   event.ID,
	})

	return nil
}

// logAnalyticsEvent logs analytics events for monitoring and debugging
func (s *AnalyticsService) logAnalyticsEvent(eventType string, data map[string]interface{}) {
	logEntry := map[string]interface{}{
		"timestamp":  time.Now().Format(time.RFC3339),
		"event_type": eventType,
		"component":  "analytics_service",
		"data":       data,
	}

	// Log to structured logger (in production, use proper logging framework)
	logJSON, _ := json.Marshal(logEntry)
	fmt.Printf("📊 ANALYTICS: %s\n", string(logJSON))
}

// logAnalyticsError logs analytics errors with detailed context
func (s *AnalyticsService) logAnalyticsError(errorType string, err error, context map[string]interface{}) {
	logEntry := map[string]interface{}{
		"timestamp":  time.Now().Format(time.RFC3339),
		"error_type": errorType,
		"component":  "analytics_service",
		"error":      err.Error(),
		"context":    context,
		"severity":   s.determineErrorSeverity(errorType, err),
	}

	// Log to structured logger
	logJSON, _ := json.Marshal(logEntry)
	fmt.Printf("🚨 ANALYTICS ERROR: %s\n", string(logJSON))

	// Create alert for critical errors
	if s.determineErrorSeverity(errorType, err) == "critical" {
		s.createAlert("critical", "Analytics System Error",
			fmt.Sprintf("Critical analytics error (%s): %v", errorType, err))
	}
}

// determineErrorSeverity determines the severity level of an error
func (s *AnalyticsService) determineErrorSeverity(errorType string, err error) string {
	errorMsg := strings.ToLower(err.Error())

	// Critical errors
	criticalPatterns := []string{
		"connection refused", "connection timeout", "database down",
		"out of memory", "disk full", "permission denied",
		"authentication failed", "authorization failed",
	}

	for _, pattern := range criticalPatterns {
		if strings.Contains(errorMsg, pattern) {
			return "critical"
		}
	}

	// Warning errors
	warningPatterns := []string{
		"validation failed", "rate limit", "invalid data",
		"timeout", "temporary failure", "retry",
	}

	for _, pattern := range warningPatterns {
		if strings.Contains(errorMsg, pattern) {
			return "warning"
		}
	}

	return "info"
}

// Rate limiting map (in production, use Redis)
var rateLimitMap = make(map[string][]time.Time)

// CheckRateLimit implements simple in-memory rate limiting
func (s *AnalyticsService) CheckRateLimit(key string, limit int, window time.Duration) bool {
	now := time.Now()
	windowStart := now.Add(-window)

	// Clean old entries
	if times, exists := rateLimitMap[key]; exists {
		var validTimes []time.Time
		for _, t := range times {
			if t.After(windowStart) {
				validTimes = append(validTimes, t)
			}
		}
		rateLimitMap[key] = validTimes
	} else {
		rateLimitMap[key] = []time.Time{}
	}

	// Check if limit exceeded
	if len(rateLimitMap[key]) >= limit {
		return false
	}

	// Add current request
	rateLimitMap[key] = append(rateLimitMap[key], now)
	return true
}

// ValidateEvent validates an analytics event with comprehensive security checks
func (s *AnalyticsService) ValidateEvent(event map[string]interface{}) error {
	// Check required fields
	eventType, ok := event["event_type"].(string)
	if !ok || eventType == "" {
		return fmt.Errorf("event_type is required and must be a non-empty string")
	}

	// Validate event type (whitelist approach)
	validEventTypes := map[string]bool{
		"page_view":              true,
		"video_view":             true,
		"video_like":             true,
		"video_comment":          true,
		"video_share":            true,
		"user_signup":            true,
		"user_login":             true,
		"subscription_created":   true,
		"subscription_cancelled": true,
		"payment_processed":      true,
		"search_performed":       true,
		"session_start":          true,
		"session_end":            true,
		"error_occurred":         true,
		"video_play":             true,
		"video_pause":            true,
		"video_seek":             true,
		"video_complete":         true,
		"form_submit":            true,
		"button_click":           true,
		"link_click":             true,
		"scroll":                 true,
		"time_on_page":           true,
	}

	if !validEventTypes[eventType] {
		return fmt.Errorf("invalid event_type: %s", eventType)
	}

	// Validate timestamp if present
	if timestamp, ok := event["timestamp"]; ok {
		switch t := timestamp.(type) {
		case string:
			if _, err := time.Parse(time.RFC3339, t); err != nil {
				return fmt.Errorf("invalid timestamp format: %v", err)
			}
		case float64:
			// Unix timestamp
			if t < 0 || t > float64(time.Now().Unix())+86400 { // Allow 1 day in future
				return fmt.Errorf("timestamp out of valid range")
			}
		default:
			return fmt.Errorf("timestamp must be string or number")
		}
	}

	// Validate session_id if present
	if sessionID, ok := event["session_id"].(string); ok {
		if len(sessionID) > 255 {
			return fmt.Errorf("session_id too long (max 255 characters)")
		}
		// Validate session_id format (should be alphanumeric with hyphens/underscores)
		if !s.isValidSessionID(sessionID) {
			return fmt.Errorf("invalid session_id format")
		}
	}

	// Validate subsite if present
	if subsite, ok := event["subsite"].(string); ok {
		validSubsites := map[string]bool{
			"streaming": true,
			"articles":  true,
			"expo":      true,
		}
		if !validSubsites[subsite] {
			return fmt.Errorf("invalid subsite: %s", subsite)
		}
	}

	// Validate user_id if present
	if userID, ok := event["user_id"]; ok {
		switch uid := userID.(type) {
		case float64:
			if uid <= 0 || uid > 999999999 {
				return fmt.Errorf("invalid user_id value")
			}
		case int:
			if uid <= 0 || uid > 999999999 {
				return fmt.Errorf("invalid user_id value")
			}
		case string:
			if uid == "" {
				return fmt.Errorf("user_id cannot be empty string")
			}
		default:
			return fmt.Errorf("user_id must be number or string")
		}
	}

	// Validate event-specific data
	if err := s.validateEventSpecificData(eventType, event); err != nil {
		return fmt.Errorf("event-specific validation failed: %w", err)
	}

	// Check for suspicious patterns
	if err := s.detectSuspiciousActivity(event); err != nil {
		return fmt.Errorf("suspicious activity detected: %w", err)
	}

	return nil
}

// isValidSessionID validates session ID format
func (s *AnalyticsService) isValidSessionID(sessionID string) bool {
	if len(sessionID) == 0 {
		return false
	}

	// Session ID should be alphanumeric with hyphens, underscores, and dots
	for _, char := range sessionID {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_' || char == '.') {
			return false
		}
	}
	return true
}

// validateEventSpecificData validates data specific to each event type
func (s *AnalyticsService) validateEventSpecificData(eventType string, event map[string]interface{}) error {
	switch eventType {
	case "video_view", "video_play", "video_pause", "video_seek", "video_complete":
		if videoID, ok := event["video_id"].(string); ok {
			if len(videoID) > 100 {
				return fmt.Errorf("video_id too long")
			}
		} else {
			return fmt.Errorf("video_id required for video events")
		}

		// Validate duration if present
		if duration, ok := event["duration"].(float64); ok {
			if duration < 0 || duration > 86400 { // Max 24 hours
				return fmt.Errorf("invalid duration value")
			}
		}

	case "page_view":
		if path, ok := event["path"].(string); ok {
			if len(path) > 500 {
				return fmt.Errorf("path too long")
			}
			// Validate path format
			if !strings.HasPrefix(path, "/") {
				return fmt.Errorf("invalid path format")
			}
		}

	case "form_submit":
		if formID, ok := event["form_id"].(string); ok {
			if len(formID) > 100 {
				return fmt.Errorf("form_id too long")
			}
		}

	case "error_occurred":
		if errorMsg, ok := event["error_message"].(string); ok {
			if len(errorMsg) > 1000 {
				return fmt.Errorf("error_message too long")
			}
		}
	}

	return nil
}

// detectSuspiciousActivity checks for suspicious patterns in analytics events
func (s *AnalyticsService) detectSuspiciousActivity(event map[string]interface{}) error {
	// Check for rapid-fire events (rate limiting)
	clientIP := ""
	if ip, ok := event["ip_address"].(string); ok {
		clientIP = ip
	}

	if clientIP != "" {
		rateLimitKey := fmt.Sprintf("analytics_rate:%s", clientIP)
		if !s.CheckRateLimit(rateLimitKey, 1000, time.Minute) { // 1000 events per minute per IP
			return fmt.Errorf("rate limit exceeded for IP: %s", clientIP)
		}
	}

	// Check for suspicious user agent patterns
	if userAgent, ok := event["user_agent"].(string); ok {
		if s.isSuspiciousUserAgent(userAgent) {
			return fmt.Errorf("suspicious user agent detected")
		}
	}

	// Check for data injection attempts
	if err := s.checkForInjectionAttempts(event); err != nil {
		return fmt.Errorf("potential injection attempt detected: %w", err)
	}

	return nil
}

// isSuspiciousUserAgent checks for suspicious user agent patterns
func (s *AnalyticsService) isSuspiciousUserAgent(userAgent string) bool {
	suspiciousPatterns := []string{
		"bot", "crawler", "spider", "scraper",
		"curl", "wget", "python", "java",
		"sqlmap", "nikto", "nmap",
		"<script>", "javascript:", "onload=",
	}

	userAgentLower := strings.ToLower(userAgent)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(userAgentLower, pattern) {
			return true
		}
	}

	return false
}

// checkForInjectionAttempts checks for potential injection attacks
func (s *AnalyticsService) checkForInjectionAttempts(event map[string]interface{}) error {
	dangerousPatterns := []string{
		"<script>", "</script>", "javascript:", "vbscript:",
		"onload=", "onerror=", "onclick=", "onmouseover=",
		"<iframe>", "</iframe>", "<object>", "</object>",
		"<embed>", "</embed>", "<form>", "</form>",
		"union select", "drop table", "delete from",
		"insert into", "update set", "alter table",
		"exec(", "eval(", "system(", "shell_exec(",
	}

	// Recursively check all string values in the event
	return s.checkValueForInjection(event, dangerousPatterns)
}

// checkValueForInjection recursively checks values for injection patterns
func (s *AnalyticsService) checkValueForInjection(value interface{}, patterns []string) error {
	switch v := value.(type) {
	case string:
		valueLower := strings.ToLower(v)
		for _, pattern := range patterns {
			if strings.Contains(valueLower, pattern) {
				return fmt.Errorf("dangerous pattern detected: %s", pattern)
			}
		}
	case map[string]interface{}:
		for _, val := range v {
			if err := s.checkValueForInjection(val, patterns); err != nil {
				return err
			}
		}
	case []interface{}:
		for _, val := range v {
			if err := s.checkValueForInjection(val, patterns); err != nil {
				return err
			}
		}
	}
	return nil
}

// SanitizeEvent sanitizes event data to prevent injection attacks and ensure data integrity
func (s *AnalyticsService) SanitizeEvent(event map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	for key, value := range event {
		// Sanitize the key name
		sanitizedKey := s.sanitizeKey(key)

		// Sanitize the value
		sanitizedValue := s.sanitizeValue(value, 0)

		sanitized[sanitizedKey] = sanitizedValue
	}

	return sanitized
}

// sanitizeKey sanitizes object keys
func (s *AnalyticsService) sanitizeKey(key string) string {
	// Remove potentially dangerous characters from keys
	dangerousChars := []string{"<", ">", "\"", "'", "&", "javascript:", "onload=", "onerror="}
	sanitized := key

	for _, char := range dangerousChars {
		sanitized = strings.ReplaceAll(sanitized, char, "")
	}

	// Limit key length
	if len(sanitized) > 100 {
		sanitized = sanitized[:100]
	}

	return sanitized
}

// sanitizeValue recursively sanitizes nested values with enhanced security
func (s *AnalyticsService) sanitizeValue(value interface{}, depth int) interface{} {
	if depth > 5 { // Prevent deep nesting
		return "[truncated]"
	}

	switch v := value.(type) {
	case string:
		return s.sanitizeString(v)
	case map[string]interface{}:
		result := make(map[string]interface{})
		count := 0
		for key, val := range v {
			if count >= 50 { // Limit object size
				break
			}
			sanitizedKey := s.sanitizeKey(key)
			result[sanitizedKey] = s.sanitizeValue(val, depth+1)
			count++
		}
		return result
	case []interface{}:
		if len(v) > 100 { // Limit array size
			v = v[:100]
		}
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = s.sanitizeValue(val, depth+1)
		}
		return result
	case float64:
		// Validate numeric values
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return 0.0
		}
		// Limit to reasonable range
		if v < -999999999 || v > 999999999 {
			return 0.0
		}
		return v
	case int:
		// Limit to reasonable range
		if v < -999999999 || v > 999999999 {
			return 0
		}
		return v
	default:
		// Convert unknown types to string and sanitize
		return s.sanitizeString(fmt.Sprintf("%v", v))
	}
}

// sanitizeString performs comprehensive string sanitization
func (s *AnalyticsService) sanitizeString(str string) string {
	if len(str) == 0 {
		return ""
	}

	// Remove HTML tags and scripts
	sanitized := s.removeHTMLTags(str)

	// Remove dangerous patterns
	dangerousPatterns := []string{
		"<script>", "</script>", "javascript:", "vbscript:",
		"onload=", "onerror=", "onclick=", "onmouseover=",
		"<iframe>", "</iframe>", "<object>", "</object>",
		"<embed>", "</embed>", "<form>", "</form>",
		"union select", "drop table", "delete from",
		"insert into", "update set", "alter table",
		"exec(", "eval(", "system(", "shell_exec(",
		"document.cookie", "localStorage", "sessionStorage",
		"window.location", "history.pushState",
	}

	for _, pattern := range dangerousPatterns {
		sanitized = strings.ReplaceAll(sanitized, pattern, "")
	}

	// Remove control characters
	sanitized = s.removeControlCharacters(sanitized)

	// Limit string length
	if len(sanitized) > 1000 {
		sanitized = sanitized[:1000]
	}

	return sanitized
}

// removeHTMLTags removes HTML tags from strings
func (s *AnalyticsService) removeHTMLTags(str string) string {
	// Simple HTML tag removal (for production, use a proper HTML sanitizer)
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(str, "")
}

// removeControlCharacters removes control characters from strings
func (s *AnalyticsService) removeControlCharacters(str string) string {
	var result strings.Builder
	for _, char := range str {
		if char >= 32 || char == '\n' || char == '\r' || char == '\t' {
			result.WriteRune(char)
		}
	}
	return result.String()
}

// ProcessBatchEvents processes events in parallel with enhanced error handling
func (s *AnalyticsService) ProcessBatchEvents(events []map[string]interface{}, userID *int, ipAddress, userAgent string) (int, []string) {
	if len(events) == 0 {
		return 0, nil
	}

	// Log batch processing start
	s.logAnalyticsEvent("batch_start", map[string]interface{}{
		"total_events": len(events),
		"user_id":      userID,
		"ip_address":   ipAddress,
	})

	// Use worker pool for parallel processing
	numWorkers := 4
	if len(events) < numWorkers {
		numWorkers = len(events)
	}

	eventChan := make(chan map[string]interface{}, len(events))
	resultChan := make(chan error, len(events))

	// Start workers
	for i := 0; i < numWorkers; i++ {
		go func() {
			for event := range eventChan {
				eventType, _ := event["event_type"].(string)
				sessionID := ""
				if sid, ok := event["session_id"].(string); ok {
					sessionID = sid
				}
				subsite := "streaming"
				if sub, ok := event["subsite"].(string); ok {
					subsite = sub
				}

				err := s.TrackEvent(eventType, userID, sessionID, subsite, event, ipAddress, userAgent)
				resultChan <- err
			}
		}()
	}

	// Send events to workers
	for _, event := range events {
		eventChan <- event
	}
	close(eventChan)

	// Collect results
	processed := 0
	var errors []string
	errorCount := 0

	for i := 0; i < len(events); i++ {
		if err := <-resultChan; err != nil {
			errorCount++
			errors = append(errors, fmt.Sprintf("Event %d: %v", i+1, err))

			// Log individual event errors
			s.logAnalyticsError("batch_event_failed", err, map[string]interface{}{
				"event_index":  i + 1,
				"total_events": len(events),
			})
		} else {
			processed++
		}
	}

	// Log batch processing completion
	s.logAnalyticsEvent("batch_complete", map[string]interface{}{
		"total_events": len(events),
		"processed":    processed,
		"errors":       errorCount,
		"success_rate": float64(processed) / float64(len(events)),
	})

	// Create alert if error rate is high
	errorRate := float64(errorCount) / float64(len(events))
	if errorRate > 0.5 { // More than 50% errors
		s.createAlert("warning", "High Analytics Error Rate",
			fmt.Sprintf("Batch processing error rate: %.1f%% (%d/%d events failed)",
				errorRate*100, errorCount, len(events)))
	}

	return processed, errors
}

// Helper methods for analytics calculations
func (s *AnalyticsService) getNewUsersCount(duration time.Duration) int {
	// Query database for new users in the specified duration
	endDate := time.Now()
	startDate := endDate.Add(-duration)

	count, err := s.db.GetNewUsersCount(startDate, endDate)
	if err != nil {
		// Log error but don't fail the entire request
		fmt.Printf("Error getting new users count: %v\n", err)
		return 0
	}
	return count
}

func (s *AnalyticsService) getPublishedVideosCount() int {
	count, err := s.db.GetPublishedVideosCount()
	if err != nil {
		fmt.Printf("Error getting published videos count: %v\n", err)
		return 0
	}
	return count
}

func (s *AnalyticsService) getPendingVideosCount() int {
	count, err := s.db.GetPendingVideosCount()
	if err != nil {
		fmt.Printf("Error getting pending videos count: %v\n", err)
		return 0
	}
	return count
}

func (s *AnalyticsService) getDraftVideosCount() int {
	count, err := s.db.GetDraftVideosCount()
	if err != nil {
		fmt.Printf("Error getting draft videos count: %v\n", err)
		return 0
	}
	return count
}

func (s *AnalyticsService) getAverageVideoRating() float64 {
	rating, err := s.db.GetAverageVideoRating()
	if err != nil {
		fmt.Printf("Error getting average video rating: %v\n", err)
		return 0.0
	}
	return rating
}

func (s *AnalyticsService) getTopVideoCategories() []map[string]interface{} {
	categories, err := s.db.GetTopVideoCategories(10)
	if err != nil {
		fmt.Printf("Error getting top video categories: %v\n", err)
		return []map[string]interface{}{}
	}
	return categories
}

func (s *AnalyticsService) getNewSubscriptionsCount(duration time.Duration) int {
	endDate := time.Now()
	startDate := endDate.Add(-duration)

	count, err := s.db.GetNewSubscriptionsCount(startDate, endDate)
	if err != nil {
		fmt.Printf("Error getting new subscriptions count: %v\n", err)
		return 0
	}
	return count
}

func (s *AnalyticsService) getRevenueForPeriod(duration time.Duration) float64 {
	endDate := time.Now()
	startDate := endDate.Add(-duration)

	revenue, err := s.db.GetRevenueForPeriod(startDate, endDate)
	if err != nil {
		fmt.Printf("Error getting revenue for period: %v\n", err)
		return 0.0
	}
	return revenue
}

func (s *AnalyticsService) calculateMRR() float64 {
	mrr, err := s.db.CalculateMRR()
	if err != nil {
		fmt.Printf("Error calculating MRR: %v\n", err)
		return 0.0
	}
	return mrr
}

func (s *AnalyticsService) calculateARR() float64 {
	arr, err := s.db.CalculateARR()
	if err != nil {
		fmt.Printf("Error calculating ARR: %v\n", err)
		return 0.0
	}
	return arr
}

func (s *AnalyticsService) calculateGrowthRate(metric string, startDate, endDate time.Time) float64 {
	rate, err := s.db.CalculateGrowthRate(metric, startDate, endDate)
	if err != nil {
		fmt.Printf("Error calculating growth rate: %v\n", err)
		return 0.0
	}
	return rate
}

func (s *AnalyticsService) getLiveEvents(duration time.Duration) ([]map[string]interface{}, error) {
	events, err := s.db.GetLiveEvents(duration)
	if err != nil {
		return nil, fmt.Errorf("failed to get live events: %w", err)
	}
	return events, nil
}

func (s *AnalyticsService) getTopContentNow() ([]map[string]interface{}, error) {
	content, err := s.db.GetTopContentNow(10)
	if err != nil {
		return nil, fmt.Errorf("failed to get top content: %w", err)
	}
	return content, nil
}

func (s *AnalyticsService) getServerLoad() float64 {
	load, err := s.db.GetServerLoad()
	if err != nil {
		fmt.Printf("Error getting server load: %v\n", err)
		return 0.0
	}
	return load
}

func (s *AnalyticsService) getBandwidthUsage() string {
	usage, err := s.db.GetBandwidthUsage()
	if err != nil {
		fmt.Printf("Error getting bandwidth usage: %v\n", err)
		return "0 MB/s"
	}
	return usage
}

func (s *AnalyticsService) getErrorRate() float64 {
	rate, err := s.db.GetErrorRate()
	if err != nil {
		fmt.Printf("Error getting error rate: %v\n", err)
		return 0.0
	}
	return rate
}

func (s *AnalyticsService) getResponseTime() int {
	time, err := s.db.GetAverageResponseTime()
	if err != nil {
		fmt.Printf("Error getting response time: %v\n", err)
		return 0
	}
	return time
}

func (s *AnalyticsService) getTotalEventsTracked() (int, error) {
	// This would count total analytics events
	return 1500000, nil
}

func (s *AnalyticsService) getSystemMetrics() (map[string]interface{}, error) {
	// Get real system metrics from database
	metrics, err := s.db.GetSystemMetrics(time.Now().Add(-1*time.Hour), time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get system metrics: %w", err)
	}

	var latest *database.SystemMetrics
	if len(metrics) > 0 {
		latest = metrics[0]
	}

	// Get database health
	dbHealth, err := s.checkDatabaseHealth()
	if err != nil {
		return nil, fmt.Errorf("failed to check database health: %w", err)
	}

	// Get Redis health
	redisHealth, err := s.checkRedisHealth()
	if err != nil {
		return nil, fmt.Errorf("failed to check Redis health: %w", err)
	}

	// Get external API health
	apiHealth, err := s.checkExternalAPIHealth()
	if err != nil {
		return nil, fmt.Errorf("failed to check external API health: %w", err)
	}

	// Calculate system load and performance metrics
	systemLoad := s.calculateSystemLoad(metrics)

	// Get disk usage
	diskUsage, err := s.getDiskUsage()
	if err != nil {
		diskUsage = map[string]interface{}{
			"total":   "0 GB",
			"used":    "0 GB",
			"free":    "0 GB",
			"percent": 0.0,
		}
	}

	// Get memory usage
	memoryUsage, err := s.getMemoryUsage()
	if err != nil {
		memoryUsage = map[string]interface{}{
			"total":   "0 GB",
			"used":    "0 GB",
			"free":    "0 GB",
			"percent": 0.0,
		}
	}

	// Get network statistics
	networkStats := s.getNetworkStatistics(metrics)

	// Determine overall system status
	overallStatus := "healthy"
	if latest != nil && latest.ErrorRate > 0.05 { // 5% error rate threshold
		overallStatus = "warning"
	}
	if latest != nil && latest.ErrorRate > 0.10 { // 10% error rate threshold
		overallStatus = "critical"
	}

	// Prepare CPU data with nil checks
	cpuData := map[string]interface{}{
		"usage":        0.0, // Default value
		"cores":        runtime.NumCPU(),
		"load_average": systemLoad,
	}
	if latest != nil {
		cpuData["usage"] = latest.CPUUsage
	}

	// Prepare performance data with nil checks
	performanceData := map[string]interface{}{
		"response_time":       0,   // Default value
		"error_rate":          0.0, // Default value
		"active_sessions":     0,   // Default value
		"requests_per_second": s.calculateRequestsPerSecond(metrics),
	}
	if latest != nil {
		performanceData["response_time"] = latest.ResponseTime
		performanceData["error_rate"] = latest.ErrorRate
		performanceData["active_sessions"] = latest.ActiveSessions
	}

	return map[string]interface{}{
		"status":        overallStatus,
		"timestamp":     time.Now().Format(time.RFC3339),
		"uptime":        s.getSystemUptime(),
		"cpu":           cpuData,
		"memory":        memoryUsage,
		"disk":          diskUsage,
		"network":       networkStats,
		"database":      dbHealth,
		"redis":         redisHealth,
		"external_apis": apiHealth,
		"performance":   performanceData,
		"alerts":        s.getActiveAlerts(),
	}, nil
}

func (s *AnalyticsService) getSubsiteHealth() (map[string]interface{}, error) {
	// Check health of each subsite
	subsites := []string{"streaming", "articles", "expo"}
	healthData := make(map[string]interface{})

	for _, subsite := range subsites {
		health, err := s.checkSubsiteHealth(subsite)
		if err != nil {
			health = map[string]interface{}{
				"status":     "unknown",
				"error":      err.Error(),
				"last_check": time.Now().Format(time.RFC3339),
			}
		}
		healthData[subsite] = health
	}

	return healthData, nil
}

// checkDatabaseHealth performs real database health checks
func (s *AnalyticsService) checkDatabaseHealth() (map[string]interface{}, error) {
	start := time.Now()

	// Test database connection
	err := s.db.Ping()
	if err != nil {
		return map[string]interface{}{
			"status":        "unhealthy",
			"error":         err.Error(),
			"response_time": time.Since(start).Milliseconds(),
			"last_check":    time.Now().Format(time.RFC3339),
		}, nil
	}

	// Get database statistics
	dbStats, err := s.getDatabaseStatistics()
	if err != nil {
		return map[string]interface{}{
			"status":        "warning",
			"error":         err.Error(),
			"response_time": time.Since(start).Milliseconds(),
			"last_check":    time.Now().Format(time.RFC3339),
		}, nil
	}

	return map[string]interface{}{
		"status":        "healthy",
		"response_time": time.Since(start).Milliseconds(),
		"connections":   dbStats["connections"],
		"size":          dbStats["size"],
		"tables":        dbStats["tables"],
		"last_check":    time.Now().Format(time.RFC3339),
	}, nil
}

// checkRedisHealth performs Redis health checks
func (s *AnalyticsService) checkRedisHealth() (map[string]interface{}, error) {
	// Try to connect to Redis
	redisClient := s.db.GetRedisClient()
	if redisClient == nil {
		return map[string]interface{}{
			"status":     "disconnected",
			"error":      "Redis client not available",
			"last_check": time.Now().Format(time.RFC3339),
		}, nil
	}

	start := time.Now()

	// Test Redis connection
	ctx := context.Background()
	err := redisClient.Ping(ctx).Err()
	if err != nil {
		return map[string]interface{}{
			"status":        "unhealthy",
			"error":         err.Error(),
			"response_time": time.Since(start).Milliseconds(),
			"last_check":    time.Now().Format(time.RFC3339),
		}, nil
	}

	// Get Redis info
	info, err := redisClient.Info(ctx).Result()
	if err != nil {
		return map[string]interface{}{
			"status":        "warning",
			"error":         err.Error(),
			"response_time": time.Since(start).Milliseconds(),
			"last_check":    time.Now().Format(time.RFC3339),
		}, nil
	}

	// Parse Redis info for key metrics
	redisStats := s.parseRedisInfo(info)

	return map[string]interface{}{
		"status":            "healthy",
		"response_time":     time.Since(start).Milliseconds(),
		"memory_used":       redisStats["memory_used"],
		"memory_total":      redisStats["memory_total"],
		"connected_clients": redisStats["connected_clients"],
		"last_check":        time.Now().Format(time.RFC3339),
	}, nil
}

// checkExternalAPIHealth checks external API dependencies
func (s *AnalyticsService) checkExternalAPIHealth() (map[string]interface{}, error) {
	apis := map[string]string{
		"youtube": "https://www.googleapis.com/youtube/v3",
		"email":   "https://api.sendgrid.com/v3",
		"cdn":     "https://api.cloudflare.com/client/v4",
	}

	healthData := make(map[string]interface{})

	for name, url := range apis {
		health, err := s.checkAPIEndpoint(name, url)
		if err != nil {
			health = map[string]interface{}{
				"status":     "unhealthy",
				"error":      err.Error(),
				"last_check": time.Now().Format(time.RFC3339),
			}
		}
		healthData[name] = health
	}

	return healthData, nil
}

// checkSubsiteHealth checks health of a specific subsite
func (s *AnalyticsService) checkSubsiteHealth(subsite string) (map[string]interface{}, error) {
	// Get recent analytics events for this subsite with pagination
	events, err := s.db.GetAnalyticsEventsBySubsite(subsite, time.Now().Add(-5*time.Minute), time.Now(), 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get subsite events: %w", err)
	}

	// Calculate error rate
	totalEvents := len(events)
	errorCount := 0
	totalResponseTime := 0

	for _, event := range events {
		if strings.Contains(event.EventType, "error") {
			errorCount++
		}
		// Response time would be calculated from event data
	}

	errorRate := 0.0
	if totalEvents > 0 {
		errorRate = float64(errorCount) / float64(totalEvents)
	}

	// Determine status based on error rate
	status := "healthy"
	if errorRate > 0.05 {
		status = "warning"
	}
	if errorRate > 0.10 {
		status = "critical"
	}

	// Calculate average response time safely
	avgResponseTime := 0
	if totalEvents > 0 {
		avgResponseTime = totalResponseTime / totalEvents
	}

	return map[string]interface{}{
		"status":             status,
		"response_time":      avgResponseTime,
		"error_rate":         errorRate,
		"active_connections": totalEvents,
		"last_check":         time.Now().Format(time.RFC3339),
	}, nil
}

// Helper methods for system monitoring
func (s *AnalyticsService) calculateSystemLoad(metrics []*database.SystemMetrics) []float64 {
	if len(metrics) == 0 {
		return []float64{0.0, 0.0, 0.0}
	}

	// Calculate load average from recent metrics
	var load1, load5, load15 float64
	count := 0

	for _, metric := range metrics {
		if metric.Timestamp.After(time.Now().Add(-15 * time.Minute)) {
			load1 += metric.CPUUsage
			count++
		}
	}

	if count > 0 {
		load1 = load1 / float64(count)
		load5 = load1 * 0.8  // Approximate 5-minute load
		load15 = load1 * 0.6 // Approximate 15-minute load
	}

	return []float64{load1, load5, load15}
}

func (s *AnalyticsService) getSystemUptime() string {
	// In production, this would read from /proc/uptime or system API
	return "5 days 12 hours 34 minutes"
}

func (s *AnalyticsService) getDiskUsage() (map[string]interface{}, error) {
	// In production, this would use syscall or system API
	return map[string]interface{}{
		"total":   "100 GB",
		"used":    "45.2 GB",
		"free":    "54.8 GB",
		"percent": 45.2,
	}, nil
}

func (s *AnalyticsService) getMemoryUsage() (map[string]interface{}, error) {
	// In production, this would use syscall or system API
	return map[string]interface{}{
		"total":   "8 GB",
		"used":    "5.4 GB",
		"free":    "2.6 GB",
		"percent": 67.5,
	}, nil
}

func (s *AnalyticsService) getNetworkStatistics(metrics []*database.SystemMetrics) map[string]interface{} {
	if len(metrics) == 0 {
		return map[string]interface{}{
			"bytes_in":    "0 MB/s",
			"bytes_out":   "0 MB/s",
			"packets_in":  0,
			"packets_out": 0,
		}
	}

	latest := metrics[0]

	return map[string]interface{}{
		"bytes_in":    fmt.Sprintf("%.1f MB/s", float64(latest.NetworkIn)/(1024*1024)),
		"bytes_out":   fmt.Sprintf("%.1f MB/s", float64(latest.NetworkOut)/(1024*1024)),
		"packets_in":  latest.NetworkIn / 1500, // Approximate packet count
		"packets_out": latest.NetworkOut / 1500,
	}
}

func (s *AnalyticsService) getDatabaseStatistics() (map[string]interface{}, error) {
	// Get database size
	var size int64
	err := s.db.QueryRow("SELECT pg_database_size(current_database())").Scan(&size)
	if err != nil {
		return nil, err
	}

	// Get connection count
	var connections int
	err = s.db.QueryRow("SELECT count(*) FROM pg_stat_activity").Scan(&connections)
	if err != nil {
		return nil, err
	}

	// Get table count
	var tables int
	err = s.db.QueryRow("SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tables)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"size":        fmt.Sprintf("%.1f GB", float64(size)/(1024*1024*1024)),
		"connections": connections,
		"tables":      tables,
	}, nil
}

func (s *AnalyticsService) parseRedisInfo(info string) map[string]interface{} {
	lines := strings.Split(info, "\n")
	stats := make(map[string]interface{})

	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "used_memory_human":
					stats["memory_used"] = value
				case "maxmemory_human":
					stats["memory_total"] = value
				case "connected_clients":
					if clients, err := strconv.Atoi(value); err == nil {
						stats["connected_clients"] = clients
					}
				}
			}
		}
	}

	return stats
}

func (s *AnalyticsService) checkAPIEndpoint(name, url string) (map[string]interface{}, error) {
	start := time.Now()

	// Make a simple HEAD request to check availability
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Head(url)

	responseTime := time.Since(start).Milliseconds()

	if err != nil {
		return map[string]interface{}{
			"status":        "unhealthy",
			"error":         err.Error(),
			"response_time": responseTime,
			"last_check":    time.Now().Format(time.RFC3339),
		}, nil
	}
	defer resp.Body.Close()

	status := "healthy"
	if resp.StatusCode >= 400 {
		status = "warning"
	}
	if resp.StatusCode >= 500 {
		status = "critical"
	}

	return map[string]interface{}{
		"status":        status,
		"response_time": responseTime,
		"status_code":   resp.StatusCode,
		"last_check":    time.Now().Format(time.RFC3339),
	}, nil
}

func (s *AnalyticsService) calculateRequestsPerSecond(metrics []*database.SystemMetrics) float64 {
	if len(metrics) < 2 {
		return 0.0
	}

	// Calculate RPS from recent metrics
	recentMetrics := metrics[:min(len(metrics), 10)]
	totalRequests := len(recentMetrics)

	if totalRequests == 0 {
		return 0.0
	}

	// Calculate time span
	firstTime := recentMetrics[len(recentMetrics)-1].Timestamp
	lastTime := recentMetrics[0].Timestamp
	duration := lastTime.Sub(firstTime).Seconds()

	if duration == 0 {
		return 0.0
	}

	return float64(totalRequests) / duration
}

// getActiveAlerts returns active alerts
func (s *AnalyticsService) getActiveAlerts() []map[string]interface{} {
	alerts, err := s.db.GetAlerts(10)
	if err != nil {
		return []map[string]interface{}{}
	}

	var activeAlerts []map[string]interface{}
	for _, alert := range alerts {
		if !alert.Acknowledged {
			activeAlerts = append(activeAlerts, map[string]interface{}{
				"id":         alert.ID,
				"severity":   alert.Severity,
				"title":      alert.Title,
				"message":    alert.Message,
				"created_at": alert.CreatedAt.Format(time.RFC3339),
			})
		}
	}

	return activeAlerts
}

// createAlert creates a new system alert
func (s *AnalyticsService) createAlert(severity, title, message string) error {
	alert := &database.Alert{
		Severity:     severity,
		Title:        title,
		Message:      message,
		CreatedAt:    time.Now(),
		Acknowledged: false,
	}

	return s.db.CreateAlert(alert)
}

// checkSystemHealthAndAlert monitors system health and creates alerts for critical issues
func (s *AnalyticsService) checkSystemHealthAndAlert() error {
	// Get current system metrics
	metrics, err := s.db.GetSystemMetrics(time.Now().Add(-5*time.Minute), time.Now())
	if err != nil {
		return fmt.Errorf("failed to get system metrics for health check: %w", err)
	}

	if len(metrics) == 0 {
		return nil // No metrics available
	}

	latest := metrics[0]

	// Check error rate
	if latest.ErrorRate > 0.10 { // 10% error rate threshold
		err := s.createAlert("critical", "High Error Rate",
			fmt.Sprintf("System error rate is %.2f%% which exceeds the 10%% threshold", latest.ErrorRate*100))
		if err != nil {
			return fmt.Errorf("failed to create error rate alert: %w", err)
		}
	} else if latest.ErrorRate > 0.05 { // 5% error rate threshold
		err := s.createAlert("warning", "Elevated Error Rate",
			fmt.Sprintf("System error rate is %.2f%% which exceeds the 5%% threshold", latest.ErrorRate*100))
		if err != nil {
			return fmt.Errorf("failed to create error rate alert: %w", err)
		}
	}

	// Check response time
	if latest.ResponseTime > 2000 { // 2 second threshold
		err := s.createAlert("critical", "High Response Time",
			fmt.Sprintf("Average response time is %dms which exceeds the 2000ms threshold", latest.ResponseTime))
		if err != nil {
			return fmt.Errorf("failed to create response time alert: %w", err)
		}
	} else if latest.ResponseTime > 1000 { // 1 second threshold
		err := s.createAlert("warning", "Elevated Response Time",
			fmt.Sprintf("Average response time is %dms which exceeds the 1000ms threshold", latest.ResponseTime))
		if err != nil {
			return fmt.Errorf("failed to create response time alert: %w", err)
		}
	}

	// Check CPU usage
	if latest.CPUUsage > 90 { // 90% CPU threshold
		err := s.createAlert("critical", "High CPU Usage",
			fmt.Sprintf("CPU usage is %.1f%% which exceeds the 90%% threshold", latest.CPUUsage))
		if err != nil {
			return fmt.Errorf("failed to create CPU alert: %w", err)
		}
	} else if latest.CPUUsage > 80 { // 80% CPU threshold
		err := s.createAlert("warning", "Elevated CPU Usage",
			fmt.Sprintf("CPU usage is %.1f%% which exceeds the 80%% threshold", latest.CPUUsage))
		if err != nil {
			return fmt.Errorf("failed to create CPU alert: %w", err)
		}
	}

	// Check database connections
	dbHealth, err := s.checkDatabaseHealth()
	if err != nil {
		err := s.createAlert("critical", "Database Health Check Failed",
			fmt.Sprintf("Database health check failed: %v", err))
		if err != nil {
			return fmt.Errorf("failed to create database alert: %w", err)
		}
	} else if dbHealth["status"] == "unhealthy" {
		err := s.createAlert("critical", "Database Unhealthy",
			"Database is reporting unhealthy status")
		if err != nil {
			return fmt.Errorf("failed to create database alert: %w", err)
		}
	}

	// Check Redis health
	redisHealth, err := s.checkRedisHealth()
	if err != nil {
		err := s.createAlert("warning", "Redis Health Check Failed",
			fmt.Sprintf("Redis health check failed: %v", err))
		if err != nil {
			return fmt.Errorf("failed to create Redis alert: %w", err)
		}
	} else if redisHealth["status"] == "unhealthy" {
		err := s.createAlert("warning", "Redis Unhealthy",
			"Redis is reporting unhealthy status")
		if err != nil {
			return fmt.Errorf("failed to create Redis alert: %w", err)
		}
	}

	// Check external API health
	apiHealth, err := s.checkExternalAPIHealth()
	if err != nil {
		err := s.createAlert("warning", "External API Health Check Failed",
			fmt.Sprintf("External API health check failed: %v", err))
		if err != nil {
			return fmt.Errorf("failed to create API alert: %w", err)
		}
	} else {
		for apiName, health := range apiHealth {
			if healthMap, ok := health.(map[string]interface{}); ok {
				if status, exists := healthMap["status"]; exists && status == "unhealthy" {
					err := s.createAlert("warning", fmt.Sprintf("%s API Unhealthy", apiName),
						fmt.Sprintf("External API %s is reporting unhealthy status", apiName))
					if err != nil {
						return fmt.Errorf("failed to create API alert: %w", err)
					}
				}
			}
		}
	}

	return nil
}

// StartSystemMonitoring starts the system monitoring loop
func (s *AnalyticsService) StartSystemMonitoring() {
	go func() {
		ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := s.checkSystemHealthAndAlert(); err != nil {
					fmt.Printf("System monitoring error: %v\n", err)
				}
			}
		}
	}()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
