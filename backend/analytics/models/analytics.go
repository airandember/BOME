package models

import (
	"bome-backend/infrastructure/database"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"time"
)

// AnalyticsEvent represents a tracked analytics event
type AnalyticsEvent struct {
	ID        int       `json:"id"`
	EventType string    `json:"event_type"`
	UserID    *int      `json:"user_id"`
	SessionID string    `json:"session_id"`
	Subsite   string    `json:"subsite"`
	EventData string    `json:"event_data"` // JSON string
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

// UserMetrics represents user engagement metrics
type UserMetrics struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	Date            time.Time `json:"date"`
	SessionCount    int       `json:"session_count"`
	SessionDuration int       `json:"session_duration"` // in seconds
	PageViews       int       `json:"page_views"`
	VideoViews      int       `json:"video_views"`
	VideoWatchTime  int       `json:"video_watch_time"` // in seconds
	LikesGiven      int       `json:"likes_given"`
	CommentsMade    int       `json:"comments_made"`
	SharesMade      int       `json:"shares_made"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// VideoMetrics represents video performance metrics
type VideoMetrics struct {
	ID             int       `json:"id"`
	VideoID        int       `json:"video_id"`
	Date           time.Time `json:"date"`
	Views          int       `json:"views"`
	UniqueViews    int       `json:"unique_views"`
	WatchTime      int       `json:"watch_time"` // in seconds
	CompletionRate float64   `json:"completion_rate"`
	Likes          int       `json:"likes"`
	Comments       int       `json:"comments"`
	Shares         int       `json:"shares"`
	BounceRate     float64   `json:"bounce_rate"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// SystemMetrics represents system performance metrics
type SystemMetrics struct {
	ID             int       `json:"id"`
	Timestamp      time.Time `json:"timestamp"`
	CPUUsage       float64   `json:"cpu_usage"`
	MemoryUsage    float64   `json:"memory_usage"`
	DiskUsage      float64   `json:"disk_usage"`
	NetworkIn      int64     `json:"network_in"`  // bytes
	NetworkOut     int64     `json:"network_out"` // bytes
	ActiveSessions int       `json:"active_sessions"`
	ErrorRate      float64   `json:"error_rate"`
	ResponseTime   int       `json:"response_time"` // milliseconds
	DatabaseSize   int64     `json:"database_size"` // bytes
	CreatedAt      time.Time `json:"created_at"`
}

// WebhookEvent represents webhook delivery events
type WebhookEvent struct {
	ID           int       `json:"id"`
	EventType    string    `json:"event_type"`
	Subsite      string    `json:"subsite"`
	Endpoint     string    `json:"endpoint"`
	Status       string    `json:"status"`        // success, failed, pending
	ResponseTime int       `json:"response_time"` // milliseconds
	PayloadSize  int       `json:"payload_size"`  // bytes
	StatusCode   int       `json:"status_code"`
	ErrorMessage string    `json:"error_message"`
	RetryCount   int       `json:"retry_count"`
	CreatedAt    time.Time `json:"created_at"`
}

// Alert represents system alerts
type Alert struct {
	ID             int        `json:"id"`
	Severity       string     `json:"severity"` // info, warning, critical
	Title          string     `json:"title"`
	Message        string     `json:"message"`
	Subsite        string     `json:"subsite"`
	Acknowledged   bool       `json:"acknowledged"`
	AcknowledgedBy *int       `json:"acknowledged_by"`
	AcknowledgedAt *time.Time `json:"acknowledged_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

// CrossSubsiteStats represents cross-subsite analytics
type CrossSubsiteStats struct {
	ID             int       `json:"id"`
	Date           time.Time `json:"date"`
	Subsite        string    `json:"subsite"`
	Users          int       `json:"users"`
	Content        int       `json:"content"`
	Views          int       `json:"views"`
	Revenue        float64   `json:"revenue"`
	EngagementRate float64   `json:"engagement_rate"`
	CreatedAt      time.Time `json:"created_at"`
}

// GetVideoCount returns the total number of videos
func (db *database.DB) GetVideoCount() (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM master_video_list`).Scan(&count)
	return count, err
}

// GetTotalViews returns the total view count across all videos
func (db *database.DB) GetTotalViews() (int64, error) {
	var total int64
	err := db.QueryRow(`SELECT COALESCE(SUM(views), 0) FROM master_video_list`).Scan(&total)
	return total, err
}

// GetTotalLikes returns the total number of likes across all videos
func (db *database.DB) GetTotalLikes() (int64, error) {
	var count int64
	err := db.QueryRow(`SELECT COUNT(*) FROM likes`).Scan(&count)
	return count, err
}

// GetActiveSubscriptions returns the number of active subscriptions
func (db *database.DB) GetActiveSubscriptions() (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM subscriptions WHERE status = 'active'`).Scan(&count)
	return count, err
}

// GetRecentActivity returns recent system activity
func (db *database.DB) GetRecentActivity(limit int) ([]*Activity, error) {
	query := `
        SELECT target_type as type, user_id, action, details, created_at 
        FROM audit_logs 
        WHERE target_type = 'video'
        UNION ALL
        SELECT target_type as type, user_id, action, details, created_at 
        FROM audit_logs 
        WHERE target_type = 'user'
        ORDER BY created_at DESC 
        LIMIT $1
    `

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*Activity
	for rows.Next() {
		activity := &Activity{}
		err := rows.Scan(&activity.Type, &activity.UserID, &activity.Action,
			&activity.Details, &activity.CreatedAt)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

// TrackAnalyticsEvent tracks a single analytics event with memory monitoring
func (db *database.DB) TrackAnalyticsEvent(event *AnalyticsEvent) error {
	// Monitor memory usage before processing
	startMem := getMemoryUsage()
	defer func() {
		endMem := getMemoryUsage()
		if endMem-startMem > 10*1024*1024 { // 10MB threshold
			log.Printf("Warning: High memory usage in TrackAnalyticsEvent: %d bytes", endMem-startMem)
		}
	}()

	query := `
		INSERT INTO analytics_events (event_type, user_id, session_id, subsite, event_data, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := db.Exec(query,
		event.EventType,
		event.UserID,
		event.SessionID,
		event.Subsite,
		event.EventData,
		event.IPAddress,
		event.UserAgent,
		event.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to track analytics event: %w", err)
	}

	// Invalidate relevant cache keys
	db.InvalidateCache("realtime_metrics")
	db.InvalidateCache("system_health")

	return nil
}

// TrackAnalyticsEventsBatch tracks multiple analytics events with optimized batch processing
func (db *database.DB) TrackAnalyticsEventsBatch(events []*AnalyticsEvent) error {
	if len(events) == 0 {
		return nil
	}

	// Monitor memory usage before processing
	startMem := getMemoryUsage()
	defer func() {
		endMem := getMemoryUsage()
		if endMem-startMem > 50*1024*1024 { // 50MB threshold
			log.Printf("Warning: High memory usage in TrackAnalyticsEventsBatch: %d bytes", endMem-startMem)
		}
	}()

	// Use batch size to prevent memory issues
	const batchSize = 1000
	var totalProcessed int

	for i := 0; i < len(events); i += batchSize {
		end := i + batchSize
		if end > len(events) {
			end = len(events)
		}

		batch := events[i:end]
		if err := db.processAnalyticsBatch(batch); err != nil {
			return fmt.Errorf("failed to process batch %d-%d: %w", i, end, err)
		}

		totalProcessed += len(batch)

		// Log progress for large batches
		if len(events) > 10000 {
			log.Printf("Processed %d/%d analytics events", totalProcessed, len(events))
		}
	}

	// Invalidate relevant cache keys
	db.InvalidateCache("realtime_metrics")
	db.InvalidateCache("system_health")

	log.Printf("Successfully processed %d analytics events in batch", totalProcessed)
	return nil
}

// processAnalyticsBatch processes a batch of analytics events efficiently
func (db *database.DB) processAnalyticsBatch(events []*AnalyticsEvent) error {
	if len(events) == 0 {
		return nil
	}

	// Start transaction for batch processing
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare statement for better performance
	stmt, err := tx.Prepare(`
		INSERT INTO analytics_events (event_type, user_id, session_id, subsite, event_data, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Execute batch insert
	for _, event := range events {
		_, err := stmt.Exec(
			event.EventType,
			event.UserID,
			event.SessionID,
			event.Subsite,
			event.EventData,
			event.IPAddress,
			event.UserAgent,
			event.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// getMemoryUsage returns current memory usage in bytes
func getMemoryUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// GetUserMetrics returns user metrics for a specific date range
func (db *database.DB) GetUserMetrics(userID int, startDate, endDate time.Time) ([]*UserMetrics, error) {
	query := `
		SELECT id, user_id, date, session_count, session_duration, page_views, video_views, 
		       video_watch_time, likes_given, comments_made, shares_made, created_at, updated_at
		FROM user_metrics 
		WHERE user_id = $1 AND date BETWEEN $2 AND $3
		ORDER BY date DESC
	`

	rows, err := db.Query(query, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*UserMetrics
	for rows.Next() {
		metric := &UserMetrics{}
		err := rows.Scan(
			&metric.ID, &metric.UserID, &metric.Date, &metric.SessionCount,
			&metric.SessionDuration, &metric.PageViews, &metric.VideoViews,
			&metric.VideoWatchTime, &metric.LikesGiven, &metric.CommentsMade,
			&metric.SharesMade, &metric.CreatedAt, &metric.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// GetVideoMetrics returns video metrics for a specific date range
func (db *database.DB) GetVideoMetrics(videoID int, startDate, endDate time.Time) ([]*VideoMetrics, error) {
	query := `
		SELECT id, video_id, date, views, unique_views, watch_time, completion_rate,
		       likes, comments, shares, bounce_rate, created_at, updated_at
		FROM video_metrics 
		WHERE video_id = $1 AND date BETWEEN $2 AND $3
		ORDER BY date DESC
	`

	rows, err := db.Query(query, videoID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*VideoMetrics
	for rows.Next() {
		metric := &VideoMetrics{}
		err := rows.Scan(
			&metric.ID, &metric.VideoID, &metric.Date, &metric.Views,
			&metric.UniqueViews, &metric.WatchTime, &metric.CompletionRate,
			&metric.Likes, &metric.Comments, &metric.Shares, &metric.BounceRate,
			&metric.CreatedAt, &metric.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// GetSystemMetrics returns system metrics for a specific time range
func (db *database.DB) GetSystemMetrics(startTime, endTime time.Time) ([]*SystemMetrics, error) {
	query := `
		SELECT id, timestamp, cpu_usage, memory_usage, disk_usage, network_in, network_out,
		       active_sessions, error_rate, response_time, database_size, created_at
		FROM system_metrics 
		WHERE timestamp BETWEEN $1 AND $2
		ORDER BY timestamp DESC
	`

	rows, err := db.Query(query, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*SystemMetrics
	for rows.Next() {
		metric := &SystemMetrics{}
		err := rows.Scan(
			&metric.ID, &metric.Timestamp, &metric.CPUUsage, &metric.MemoryUsage,
			&metric.DiskUsage, &metric.NetworkIn, &metric.NetworkOut,
			&metric.ActiveSessions, &metric.ErrorRate, &metric.ResponseTime,
			&metric.DatabaseSize, &metric.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// WebhookEventsResponse represents the response structure for webhook events
type WebhookEventsResponse struct {
	Events     []*WebhookEvent `json:"events"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	HasMore    bool            `json:"has_more"`
	TotalPages int             `json:"total_pages"`
}

// GetWebhookEvents returns webhook events for a specific time range
func (db *database.DB) GetWebhookEvents(startTime, endTime time.Time, limit int) ([]*WebhookEvent, error) {
	// Check if table exists first
	var tableExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'webhook_events'
		)
	`).Scan(&tableExists)

	if err != nil {
		return nil, fmt.Errorf("failed to check if webhook_events table exists: %w", err)
	}

	if !tableExists {
		// Return empty array if table doesn't exist
		return []*WebhookEvent{}, nil
	}

	query := `
		SELECT id, event_type, subsite, endpoint, status, response_time, payload_size,
		       status_code, error_message, retry_count, created_at
		FROM webhook_events 
		WHERE created_at BETWEEN $1 AND $2
		ORDER BY created_at DESC
		LIMIT $3
	`

	rows, err := db.Query(query, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query webhook_events: %w", err)
	}
	defer rows.Close()

	var events []*WebhookEvent
	for rows.Next() {
		event := &WebhookEvent{}
		err := rows.Scan(
			&event.ID, &event.EventType, &event.Subsite, &event.Endpoint,
			&event.Status, &event.ResponseTime, &event.PayloadSize,
			&event.StatusCode, &event.ErrorMessage, &event.RetryCount, &event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook event: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}

// GetWebhookEventsWithPagination retrieves webhook events with pagination and filtering
func (db *database.DB) GetWebhookEventsWithPagination(page, limit int, eventType, status string) (*WebhookEventsResponse, error) {
	// Check if table exists first
	var tableExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'webhook_events'
		)
	`).Scan(&tableExists)

	if err != nil {
		return nil, fmt.Errorf("failed to check if webhook_events table exists: %w", err)
	}

	if !tableExists {
		// Return empty response if table doesn't exist
		return &WebhookEventsResponse{
			Events:     []*WebhookEvent{},
			Total:      0,
			Page:       page,
			Limit:      limit,
			HasMore:    false,
			TotalPages: 0,
		}, nil
	}

	offset := (page - 1) * limit

	// Build the base query
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	// Add filters
	if eventType != "" {
		whereClause += " AND event_type = $" + fmt.Sprintf("%d", argIndex)
		args = append(args, eventType)
		argIndex++
	}

	if status != "" {
		whereClause += " AND status = $" + fmt.Sprintf("%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM webhook_events " + whereClause
	var total int
	err = db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook events count: %w", err)
	}

	// Get paginated events (newest first)
	eventsQuery := `
		SELECT id, event_type, subsite, endpoint, status, response_time, 
		       payload_size, status_code, error_message, retry_count, created_at
		FROM webhook_events 
		` + whereClause + `
		ORDER BY created_at DESC 
		LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	args = append(args, limit, offset)

	rows, err := db.Query(eventsQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query webhook events: %w", err)
	}
	defer rows.Close()

	var events []*WebhookEvent
	for rows.Next() {
		event := &WebhookEvent{}
		var errorMessage sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.Subsite,
			&event.Endpoint,
			&event.Status,
			&event.ResponseTime,
			&event.PayloadSize,
			&event.StatusCode,
			&errorMessage,
			&event.RetryCount,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook event: %w", err)
		}

		// Handle nullable error message
		if errorMessage.Valid {
			event.ErrorMessage = errorMessage.String
		}

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating webhook events: %w", err)
	}

	// Calculate pagination info
	totalPages := (total + limit - 1) / limit
	hasMore := page < totalPages

	return &WebhookEventsResponse{
		Events:     events,
		Total:      total,
		Page:       page,
		Limit:      limit,
		HasMore:    hasMore,
		TotalPages: totalPages,
	}, nil
}

// GetAlerts returns system alerts
func (db *database.DB) GetAlerts(limit int) ([]*Alert, error) {
	// Check if table exists first
	var tableExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'alerts'
		)
	`).Scan(&tableExists)

	if err != nil {
		return nil, fmt.Errorf("failed to check if alerts table exists: %w", err)
	}

	if !tableExists {
		// Return empty array if table doesn't exist
		return []*Alert{}, nil
	}

	query := `
		SELECT id, severity, title, message, subsite, acknowledged, acknowledged_by, 
		       acknowledged_at, created_at
		FROM alerts 
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*Alert
	for rows.Next() {
		alert := &Alert{}
		err := rows.Scan(
			&alert.ID, &alert.Severity, &alert.Title, &alert.Message, &alert.Subsite,
			&alert.Acknowledged, &alert.AcknowledgedBy, &alert.AcknowledgedAt, &alert.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// AcknowledgeAlert marks an alert as acknowledged
func (db *database.DB) AcknowledgeAlert(alertID int, userID int) error {
	query := `
		UPDATE alerts 
		SET acknowledged = true, acknowledged_by = $2, acknowledged_at = NOW()
		WHERE id = $1
	`

	_, err := db.Exec(query, alertID, userID)
	return err
}

// GetCrossSubsiteStats returns cross-subsite statistics
func (db *database.DB) GetCrossSubsiteStats(startDate, endDate time.Time) ([]*CrossSubsiteStats, error) {
	query := `
		SELECT id, date, subsite, users, content, views, revenue, engagement_rate, created_at
		FROM cross_subsite_stats 
		WHERE date BETWEEN $1 AND $2
		ORDER BY date DESC, subsite
	`

	rows, err := db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*CrossSubsiteStats
	for rows.Next() {
		stat := &CrossSubsiteStats{}
		err := rows.Scan(
			&stat.ID, &stat.Date, &stat.Subsite, &stat.Users, &stat.Content,
			&stat.Views, &stat.Revenue, &stat.EngagementRate, &stat.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

// GetRealTimeMetrics returns current real-time metrics with caching
func (db *database.DB) GetRealTimeMetrics() (map[string]interface{}, error) {
	// Try to get from cache first
	cacheKey := "realtime_metrics"
	if cached, err := db.getFromCache(cacheKey); err == nil && cached != nil {
		return cached.(map[string]interface{}), nil
	}

	metrics := make(map[string]interface{})

	// Optimized query for active users with proper indexing
	var activeUsers int
	err := db.QueryRow(`
		SELECT COUNT(DISTINCT user_id) 
		FROM user_sessions 
		WHERE last_activity > NOW() - INTERVAL '5 minutes' 
		AND is_active = true
		AND user_id IS NOT NULL
	`).Scan(&activeUsers)
	if err != nil {
		// Table might not exist, set default value
		activeUsers = 0
	}
	metrics["active_users"] = activeUsers

	// Optimized query for current streams with time-based filtering
	var currentStreams int
	err = db.QueryRow(`
		SELECT COUNT(DISTINCT video_id) 
		FROM analytics_events 
		WHERE event_type = 'video_view' 
		AND created_at > NOW() - INTERVAL '1 minute'
		AND video_id IS NOT NULL
	`).Scan(&currentStreams)
	if err != nil {
		// Table might not exist, set default value
		currentStreams = 0
	}
	metrics["current_streams"] = currentStreams

	// Optimized query for recent signups with index on created_at
	var recentSignups int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM users 
		WHERE created_at > NOW() - INTERVAL '1 hour'
	`).Scan(&recentSignups)
	if err != nil {
		// Table might not exist, set default value
		recentSignups = 0
	}
	metrics["recent_signups"] = recentSignups

	// Optimized query for recent subscriptions with composite index
	var recentSubscriptions int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM subscriptions 
		WHERE created_at > NOW() - INTERVAL '1 hour' 
		AND status = 'active'
	`).Scan(&recentSubscriptions)
	if err != nil {
		// Table might not exist, set default value
		recentSubscriptions = 0
	}
	metrics["recent_subscriptions"] = recentSubscriptions

	// Cache the result for 30 seconds
	db.setCache(cacheKey, metrics, 30*time.Second)

	return metrics, nil
}

// Activity represents a system activity
type Activity struct {
	Type      string    `json:"type"`
	UserID    int       `json:"user_id"`
	Action    string    `json:"action"`
	Details   string    `json:"details"`
	CreatedAt time.Time `json:"created_at"`
}

// SystemHealth represents the system's health metrics
type SystemHealth struct {
	Uptime             string `json:"uptime"`
	ResponseTime       string `json:"response_time"`
	ErrorRate          string `json:"error_rate"`
	StorageUsed        string `json:"storage_used"`
	BandwidthUsed      string `json:"bandwidth_used"`
	CDNHits            string `json:"cdn_hits"`
	DatabaseSize       string `json:"database_size"`
	ActiveSessions     int    `json:"active_sessions"`
	LastWrite          string `json:"last_write"`
	TotalEventsTracked int    `json:"total_events_tracked"`
}

// GetNewUsersCount returns the number of new users in a date range
func (db *database.DB) GetNewUsersCount(startDate, endDate time.Time) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE created_at BETWEEN $1 AND $2`
	err := db.QueryRow(query, startDate, endDate).Scan(&count)
	return count, err
}

// GetPublishedVideosCount returns the number of published videos
func (db *database.DB) GetPublishedVideosCount() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM videos WHERE status = 'published'`
	err := db.QueryRow(query).Scan(&count)
	return count, err
}

// GetPendingVideosCount returns the number of pending videos
func (db *database.DB) GetPendingVideosCount() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM videos WHERE status = 'pending'`
	err := db.QueryRow(query).Scan(&count)
	return count, err
}

// GetDraftVideosCount returns the number of draft videos
func (db *database.DB) GetDraftVideosCount() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM videos WHERE status = 'draft'`
	err := db.QueryRow(query).Scan(&count)
	return count, err
}

// GetAverageVideoRating returns the average rating across all videos
func (db *database.DB) GetAverageVideoRating() (float64, error) {
	var avg float64
	query := `SELECT COALESCE(AVG(rating), 0) FROM video_ratings`
	err := db.QueryRow(query).Scan(&avg)
	return avg, err
}

// GetTopVideoCategories returns the top video categories with counts and views
func (db *database.DB) GetTopVideoCategories(limit int) ([]map[string]interface{}, error) {
	query := `
		SELECT category, COUNT(*) as count, SUM(view_count) as views
		FROM videos 
		WHERE category IS NOT NULL AND category != ''
		GROUP BY category 
		ORDER BY views DESC 
		LIMIT $1
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []map[string]interface{}
	for rows.Next() {
		var category string
		var count int
		var views int64

		err := rows.Scan(&category, &count, &views)
		if err != nil {
			return nil, err
		}

		categories = append(categories, map[string]interface{}{
			"name":  category,
			"count": count,
			"views": views,
		})
	}

	return categories, nil
}

// GetNewSubscriptionsCount returns the number of new subscriptions in a date range
func (db *database.DB) GetNewSubscriptionsCount(startDate, endDate time.Time) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM subscriptions WHERE created_at BETWEEN $1 AND $2`
	err := db.QueryRow(query, startDate, endDate).Scan(&count)
	return count, err
}

// GetRevenueForPeriod returns the total revenue for a date range
func (db *database.DB) GetRevenueForPeriod(startDate, endDate time.Time) (float64, error) {
	var revenue float64
	query := `
		SELECT COALESCE(SUM(amount), 0) 
		FROM payments 
		WHERE status = 'completed' AND created_at BETWEEN $1 AND $2
	`
	err := db.QueryRow(query, startDate, endDate).Scan(&revenue)
	return revenue, err
}

// CalculateMRR returns the Monthly Recurring Revenue
func (db *database.DB) CalculateMRR() (float64, error) {
	var mrr float64
	query := `
		SELECT COALESCE(SUM(
			CASE 
				WHEN sp.interval = 'month' THEN sp.price
				WHEN sp.interval = 'year' THEN sp.price / 12.0
				ELSE sp.price
			END
		), 0) 
		FROM subscriptions s
		JOIN subscription_plans sp ON s.plan_id = sp.id
		WHERE s.status = 'active' AND sp.is_active = true
	`
	err := db.QueryRow(query).Scan(&mrr)
	if err != nil {
		log.Printf("❌ Error calculating MRR: %v", err)
	}
	return mrr, err
}

// CalculateARR returns the Annual Recurring Revenue
func (db *database.DB) CalculateARR() (float64, error) {
	var arr float64
	query := `
		SELECT COALESCE(SUM(
			CASE 
				WHEN sp.interval = 'month' THEN sp.price * 12.0
				WHEN sp.interval = 'year' THEN sp.price
				ELSE sp.price * 12.0
			END
		), 0) 
		FROM subscriptions s
		JOIN subscription_plans sp ON s.plan_id = sp.id
		WHERE s.status = 'active' AND sp.is_active = true
	`
	err := db.QueryRow(query).Scan(&arr)
	if err != nil {
		log.Printf("❌ Error calculating ARR: %v", err)
	}
	return arr, err
}

// CalculateGrowthRate calculates the growth rate for a metric
func (db *database.DB) CalculateGrowthRate(metric string, startDate, endDate time.Time) (float64, error) {
	// This is a simplified calculation - in production you'd want more sophisticated growth rate calculation
	var current, previous float64

	switch metric {
	case "users":
		currentInt, _ := db.GetNewUsersCount(startDate, endDate)
		current = float64(currentInt)
		previousStart := startDate.Add(-(endDate.Sub(startDate)))
		previousInt, _ := db.GetNewUsersCount(previousStart, startDate)
		previous = float64(previousInt)
	case "subscriptions":
		currentInt, _ := db.GetNewSubscriptionsCount(startDate, endDate)
		current = float64(currentInt)
		previousStart := startDate.Add(-(endDate.Sub(startDate)))
		previousInt, _ := db.GetNewSubscriptionsCount(previousStart, startDate)
		previous = float64(previousInt)
	default:
		return 0.0, nil
	}

	if previous == 0 {
		return 0.0, nil
	}

	return ((current - previous) / previous) * 100, nil
}

// GetLiveEvents returns recent analytics events
func (db *database.DB) GetLiveEvents(duration time.Duration) ([]map[string]interface{}, error) {
	query := `
		SELECT event_type, event_data, created_at
		FROM analytics_events
		WHERE created_at >= $1
		ORDER BY created_at DESC
		LIMIT 20
	`

	startTime := time.Now().Add(-duration)
	rows, err := db.Query(query, startTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		var eventType, eventData string
		var createdAt time.Time

		err := rows.Scan(&eventType, &eventData, &createdAt)
		if err != nil {
			return nil, err
		}

		events = append(events, map[string]interface{}{
			"time":    createdAt.Format(time.RFC3339),
			"event":   eventType,
			"details": eventData,
		})
	}

	return events, nil
}

// GetTopContentNow returns current top content
func (db *database.DB) GetTopContentNow(limit int) ([]map[string]interface{}, error) {
	query := `
		SELECT title, view_count as viewers
		FROM videos
		WHERE status = 'published'
		ORDER BY view_count DESC
		LIMIT $1
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var content []map[string]interface{}
	for rows.Next() {
		var title string
		var viewers int

		err := rows.Scan(&title, &viewers)
		if err != nil {
			return nil, err
		}

		content = append(content, map[string]interface{}{
			"title":   title,
			"viewers": viewers,
		})
	}

	return content, nil
}

// GetServerLoad returns current server load (simplified - in production this would come from system monitoring)
func (db *database.DB) GetServerLoad() (float64, error) {
	// For now, return a placeholder. In production, this would query system metrics
	// or integrate with a monitoring service like Prometheus
	return 0.25, nil
}

// GetBandwidthUsage returns current bandwidth usage
func (db *database.DB) GetBandwidthUsage() (string, error) {
	// For now, return a placeholder. In production, this would query system metrics
	return "45.2 MB/s", nil
}

// GetErrorRate returns current error rate
func (db *database.DB) GetErrorRate() (float64, error) {
	// Calculate error rate from recent requests
	query := `
		SELECT 
			COALESCE(
				CAST(COUNT(CASE WHEN status_code >= 400 THEN 1 END) AS FLOAT) / 
				NULLIF(COUNT(*), 0) * 100, 
				0
			) as error_rate
		FROM request_logs 
		WHERE created_at >= $1
	`

	startTime := time.Now().Add(-1 * time.Hour)
	var errorRate float64
	err := db.QueryRow(query, startTime).Scan(&errorRate)
	if err != nil {
		// If table doesn't exist or other error, return 0
		return 0.0, nil
	}

	return errorRate, nil
}

// GetAverageResponseTime returns average response time
func (db *database.DB) GetAverageResponseTime() (int, error) {
	// Calculate average response time from recent requests
	query := `
		SELECT COALESCE(AVG(response_time), 0) as avg_response_time
		FROM request_logs 
		WHERE created_at >= $1
	`

	startTime := time.Now().Add(-1 * time.Hour)
	var avgResponseTime int
	err := db.QueryRow(query, startTime).Scan(&avgResponseTime)
	if err != nil {
		// If table doesn't exist or other error, return 0
		return 0, nil
	}

	return avgResponseTime, nil
}

// GetSystemHealth returns the current system health metrics with caching
func (db *database.DB) GetSystemHealth() (*SystemHealth, error) {
	// Try to get from cache first
	cacheKey := "system_health"
	if cached, err := db.getFromCache(cacheKey); err == nil && cached != nil {
		return cached.(*SystemHealth), nil
	}

	// Get latest system metrics with optimized query
	var metrics SystemMetrics
	err := db.QueryRow(`
		SELECT id, timestamp, cpu_usage, memory_usage, disk_usage, 
		       network_in, network_out, active_sessions, error_rate, 
		       response_time, database_size, created_at
		FROM system_metrics 
		ORDER BY timestamp DESC 
		LIMIT 1
	`).Scan(
		&metrics.ID, &metrics.Timestamp, &metrics.CPUUsage, &metrics.MemoryUsage,
		&metrics.DiskUsage, &metrics.NetworkIn, &metrics.NetworkOut,
		&metrics.ActiveSessions, &metrics.ErrorRate, &metrics.ResponseTime,
		&metrics.DatabaseSize, &metrics.CreatedAt,
	)
	if err != nil {
		// Table might not exist, create default metrics
		metrics = SystemMetrics{
			ID:             1,
			Timestamp:      time.Now(),
			CPUUsage:       25.0,
			MemoryUsage:    45.0,
			DiskUsage:      60.0,
			NetworkIn:      1024 * 1024 * 10, // 10MB
			NetworkOut:     1024 * 1024 * 20, // 20MB
			ActiveSessions: 0,
			ErrorRate:      0.001,
			ResponseTime:   150,
			DatabaseSize:   1024 * 1024 * 100, // 100MB
			CreatedAt:      time.Now(),
		}
	}

	// Get total events tracked with optimized count
	var totalEvents int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM analytics_events
	`).Scan(&totalEvents)
	if err != nil {
		totalEvents = 0
	}

	health := &SystemHealth{
		Uptime:             "5 days 12 hours", // This would be calculated from system start
		ResponseTime:       fmt.Sprintf("%dms", metrics.ResponseTime),
		ErrorRate:          fmt.Sprintf("%.2f%%", metrics.ErrorRate*100),
		StorageUsed:        fmt.Sprintf("%.1f GB", float64(metrics.DatabaseSize)/(1024*1024*1024)),
		BandwidthUsed:      fmt.Sprintf("%.1f MB/s", float64(metrics.NetworkOut)/(1024*1024)),
		CDNHits:            "145,230", // This would come from CDN API
		DatabaseSize:       fmt.Sprintf("%.1f GB", float64(metrics.DatabaseSize)/(1024*1024*1024)),
		ActiveSessions:     metrics.ActiveSessions,
		LastWrite:          time.Now().Format(time.RFC3339),
		TotalEventsTracked: totalEvents,
	}

	// Cache the result for 60 seconds
	db.setCache(cacheKey, health, 60*time.Second)

	return health, nil
}

// getTotalEventsTracked returns the total number of analytics events tracked
func (db *database.DB) getTotalEventsTracked() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM analytics_events`
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		// If table doesn't exist, return 0
		return 0, nil
	}
	return count, nil
}

// GetAnalyticsOverview returns comprehensive analytics data for the dashboard
func (db *database.DB) GetAnalyticsOverview(period string) (map[string]interface{}, error) {
	// Generate cache key
	cacheKey := fmt.Sprintf("analytics_overview_%s", period)

	// Try to get from cache first
	if cached, err := db.getFromCache(cacheKey); err == nil {
		return cached.(map[string]interface{}), nil
	}

	// Calculate date range based on period
	now := time.Now()
	var startDate time.Time

	switch period {
	case "1d":
		startDate = now.AddDate(0, 0, -1)
	case "7d":
		startDate = now.AddDate(0, 0, -7)
	case "30d":
		startDate = now.AddDate(0, 0, -30)
	case "90d":
		startDate = now.AddDate(0, 0, -90)
	default:
		startDate = now.AddDate(0, 0, -7) // Default to 7 days
	}

	// Use optimized queries with proper indexing
	overview := make(map[string]interface{})

	// Get user metrics with optimized aggregation
	var userCount, newToday, activeToday int
	var growthRate float64

	// Total users - this table exists
	err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&userCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get user count: %w", err)
	}

	// New users today - this table exists
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM users 
		WHERE created_at >= CURRENT_DATE
	`).Scan(&newToday)
	if err != nil {
		return nil, fmt.Errorf("failed to get new users count: %w", err)
	}

	// Active users today - check if user_sessions table exists first
	var userSessionsExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'user_sessions'
		)
	`).Scan(&userSessionsExists)

	if err == nil && userSessionsExists {
		err = db.QueryRow(`
			SELECT COUNT(DISTINCT user_id) 
			FROM user_sessions 
			WHERE last_activity >= CURRENT_DATE 
			AND is_active = true
		`).Scan(&activeToday)
		if err != nil {
			activeToday = 0 // Set to 0 if query fails
		}
	} else {
		activeToday = 0 // Table doesn't exist, set to 0
	}

	// Calculate growth rate
	var previousPeriodCount int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM users 
		WHERE created_at BETWEEN $1 AND $2
	`, startDate.Add(-7*24*time.Hour), startDate).Scan(&previousPeriodCount)
	if err != nil {
		growthRate = 0
	} else if previousPeriodCount > 0 {
		currentPeriodCount := userCount - previousPeriodCount
		growthRate = float64(currentPeriodCount) / float64(previousPeriodCount) * 100
	}

	overview["users"] = map[string]interface{}{
		"total":        userCount,
		"new_today":    newToday,
		"active_today": activeToday,
		"growth_rate":  growthRate,
	}

	// Get content metrics with optimized queries
	var totalVideos, totalViews int64
	var avgRating float64

	// Check if videos table exists
	var videosExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'videos'
		)
	`).Scan(&videosExists)

	if err == nil && videosExists {
		err = db.QueryRow(`SELECT COUNT(*) FROM videos WHERE status = 'published'`).Scan(&totalVideos)
		if err != nil {
			totalVideos = 0
		}

		err = db.QueryRow(`SELECT COALESCE(SUM(view_count), 0) FROM videos`).Scan(&totalViews)
		if err != nil {
			totalViews = 0
		}
	} else {
		totalVideos = 0
		totalViews = 0
	}

	// Check if video_ratings table exists before querying
	var videoRatingsExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'video_ratings'
		)
	`).Scan(&videoRatingsExists)

	if err == nil && videoRatingsExists {
		err = db.QueryRow(`SELECT COALESCE(AVG(rating), 0) FROM video_ratings`).Scan(&avgRating)
		if err != nil {
			avgRating = 0
		}
	} else {
		avgRating = 0
	}

	overview["videos"] = map[string]interface{}{
		"total":       totalVideos,
		"total_views": totalViews,
		"avg_rating":  avgRating,
	}

	// Get revenue metrics with optimized aggregation
	var totalMonthly, mrr float64

	// Check if payments table exists
	var paymentsExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'payments'
		)
	`).Scan(&paymentsExists)

	if err == nil && paymentsExists {
		err = db.QueryRow(`
			SELECT COALESCE(SUM(amount), 0) 
			FROM payments 
			WHERE status = 'completed' 
			AND created_at >= DATE_TRUNC('month', CURRENT_DATE)
		`).Scan(&totalMonthly)
		if err != nil {
			totalMonthly = 0
		}
	} else {
		totalMonthly = 0
	}

	// Check if subscription_plans table exists
	var subscriptionPlansExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'subscription_plans'
		)
	`).Scan(&subscriptionPlansExists)

	if err == nil && subscriptionPlansExists {
		// Also check if subscriptions table exists
		var subscriptionsExists bool
		err = db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = 'subscriptions'
			)
		`).Scan(&subscriptionsExists)

		if err == nil && subscriptionsExists {
			err = db.QueryRow(`
				SELECT COALESCE(SUM(sp.amount), 0) 
				FROM subscriptions s
				JOIN subscription_plans sp ON s.plan_id = sp.id
				WHERE s.status = 'active' AND s.billing_cycle = 'monthly'
			`).Scan(&mrr)
			if err != nil {
				mrr = 0
			}
		} else {
			mrr = 0
		}
	} else {
		mrr = 0
	}

	overview["subscriptions"] = map[string]interface{}{
		"revenue_month": totalMonthly,
		"mrr":           mrr,
	}

	// Cache the result for 5 minutes
	db.setCache(cacheKey, overview, 5*time.Minute)

	return overview, nil
}

// getFromCache retrieves data from Redis cache
func (db *database.DB) getFromCache(key string) (interface{}, error) {
	if db.Redis == nil {
		return nil, fmt.Errorf("Redis not available")
	}

	ctx := context.Background()
	data, err := db.Redis.Client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var result interface{}
	err = json.Unmarshal([]byte(data), &result)
	return result, err
}

// setCache stores data in Redis cache
func (db *database.DB) setCache(key string, value interface{}, expiration time.Duration) {
	if db.Redis == nil {
		return
	}

	ctx := context.Background()
	data, err := json.Marshal(value)
	if err != nil {
		return
	}

	db.Redis.Client.Set(ctx, key, data, expiration)
}

// InvalidateCache removes a specific cache key
func (db *database.DB) InvalidateCache(key string) {
	if db.Redis == nil {
		return
	}

	ctx := context.Background()
	db.Redis.Client.Del(ctx, key)
}

// InvalidateAnalyticsCache removes all analytics-related cache keys
func (db *database.DB) InvalidateAnalyticsCache() {
	if db.Redis == nil {
		return
	}

	ctx := context.Background()
	pattern := "analytics_*"
	keys, err := db.Redis.Client.Keys(ctx, pattern).Result()
	if err != nil {
		return
	}

	if len(keys) > 0 {
		db.Redis.Client.Del(ctx, keys...)
	}
}

// GetAnalyticsEventsBySubsite returns analytics events for a specific subsite within a time range with pagination
func (db *database.DB) GetAnalyticsEventsBySubsite(subsite string, startTime, endTime time.Time, limit, offset int) ([]*AnalyticsEvent, error) {
	// Try to get from cache for common queries
	cacheKey := fmt.Sprintf("analytics_events_%s_%s_%s_%d_%d",
		subsite, startTime.Format("2006-01-02"), endTime.Format("2006-01-02"), limit, offset)

	if cached, err := db.getFromCache(cacheKey); err == nil && cached != nil {
		return cached.([]*AnalyticsEvent), nil
	}

	query := `
		SELECT id, event_type, user_id, session_id, subsite, event_data, ip_address, user_agent, created_at
		FROM analytics_events 
		WHERE subsite = $1 AND created_at BETWEEN $2 AND $3
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := db.Query(query, subsite, startTime, endTime, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query analytics events: %w", err)
	}
	defer rows.Close()

	var events []*AnalyticsEvent
	for rows.Next() {
		var event AnalyticsEvent
		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.UserID,
			&event.SessionID,
			&event.Subsite,
			&event.EventData,
			&event.IPAddress,
			&event.UserAgent,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan analytics event: %w", err)
		}
		events = append(events, &event)
	}

	// Cache the result for 5 minutes
	db.setCache(cacheKey, events, 5*time.Minute)

	return events, nil
}

// GetAnalyticsEventsCount returns the total count of analytics events for a subsite
func (db *database.DB) GetAnalyticsEventsCount(subsite string, startTime, endTime time.Time) (int, error) {
	// Try to get from cache
	cacheKey := fmt.Sprintf("analytics_events_count_%s_%s_%s",
		subsite, startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))

	if cached, err := db.getFromCache(cacheKey); err == nil && cached != nil {
		return cached.(int), nil
	}

	query := `
		SELECT COUNT(*) 
		FROM analytics_events 
		WHERE subsite = $1 AND created_at BETWEEN $2 AND $3
	`

	var count int
	err := db.QueryRow(query, subsite, startTime, endTime).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count analytics events: %w", err)
	}

	// Cache the result for 10 minutes
	db.setCache(cacheKey, count, 10*time.Minute)

	return count, nil
}

// GetActiveUsersCount returns the number of currently active users
func (db *database.DB) GetActiveUsersCount() (int, error) {
	var count int
	// Consider users active if they've had activity in the last 15 minutes
	query := `
		SELECT COUNT(DISTINCT user_id) 
		FROM analytics_events 
		WHERE created_at >= NOW() - INTERVAL '15 minutes'
		AND event_type IN ('page_view', 'video_view', 'user_login', 'user_action')
	`
	err := db.QueryRow(query).Scan(&count)
	return count, err
}

// GetActiveUsersTrend returns active users trend over time
func (db *database.DB) GetActiveUsersTrend() ([]map[string]interface{}, error) {
	query := `
		SELECT 
			DATE_TRUNC('hour', created_at) as hour,
			COUNT(DISTINCT user_id) as active_users
		FROM analytics_events 
		WHERE created_at >= NOW() - INTERVAL '24 hours'
		AND event_type IN ('page_view', 'video_view', 'user_login', 'user_action')
		GROUP BY DATE_TRUNC('hour', created_at)
		ORDER BY hour DESC
		LIMIT 24
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trend []map[string]interface{}
	for rows.Next() {
		var hour time.Time
		var activeUsers int
		err := rows.Scan(&hour, &activeUsers)
		if err != nil {
			return nil, err
		}
		trend = append(trend, map[string]interface{}{
			"hour":         hour,
			"active_users": activeUsers,
		})
	}

	return trend, nil
}

// GetViewAnalytics returns comprehensive view analytics
func (db *database.DB) GetViewAnalytics() (map[string]interface{}, error) {
	// Get total views from master_video_list
	var totalViews int64
	err := db.QueryRow(`SELECT COALESCE(SUM(views), 0) FROM master_video_list`).Scan(&totalViews)
	if err != nil {
		return nil, err
	}

	// Get views today (simplified - using updated_at from master_video_list)
	var viewsToday int64
	err = db.QueryRow(`
		SELECT COALESCE(SUM(views), 0) 
		FROM master_video_list 
		WHERE updated_at >= CURRENT_DATE
	`).Scan(&viewsToday)
	if err != nil {
		return nil, err
	}

	// Get views this week
	var viewsWeek int64
	err = db.QueryRow(`
		SELECT COALESCE(SUM(views), 0) 
		FROM master_video_list 
		WHERE updated_at >= CURRENT_DATE - INTERVAL '7 days'
	`).Scan(&viewsWeek)
	if err != nil {
		return nil, err
	}

	// Calculate growth rate (simplified)
	var growthRate float64
	if viewsWeek > 0 {
		// Compare with previous week
		var prevWeekViews int64
		err = db.QueryRow(`
			SELECT COALESCE(SUM(views), 0) 
			FROM master_video_list 
			WHERE updated_at >= CURRENT_DATE - INTERVAL '14 days' 
			AND updated_at < CURRENT_DATE - INTERVAL '7 days'
		`).Scan(&prevWeekViews)
		if err == nil && prevWeekViews > 0 {
			growthRate = float64(viewsWeek-prevWeekViews) / float64(prevWeekViews) * 100
		}
	}

	return map[string]interface{}{
		"total_views": totalViews,
		"views_today": viewsToday,
		"views_week":  viewsWeek,
		"growth_rate": growthRate,
	}, nil
}

// GetViewAnalyticsByPeriod returns view analytics for a specific period
func (db *database.DB) GetViewAnalyticsByPeriod(period string) (map[string]interface{}, error) {
	var interval string
	switch period {
	case "1d":
		interval = "1 day"
	case "7d":
		interval = "7 days"
	case "30d":
		interval = "30 days"
	case "90d":
		interval = "90 days"
	default:
		interval = "7 days"
	}

	query := `
		SELECT 
			DATE_TRUNC('day', created_at) as date,
			COUNT(*) as views
		FROM analytics_events 
		WHERE created_at >= NOW() - INTERVAL '` + interval + `'
		AND event_type = 'video_view'
		GROUP BY DATE_TRUNC('day', created_at)
		ORDER BY date DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dailyViews []map[string]interface{}
	var totalViews int64
	for rows.Next() {
		var date time.Time
		var views int64
		err := rows.Scan(&date, &views)
		if err != nil {
			return nil, err
		}
		dailyViews = append(dailyViews, map[string]interface{}{
			"date":  date,
			"views": views,
		})
		totalViews += views
	}

	return map[string]interface{}{
		"period":      period,
		"total_views": totalViews,
		"daily_views": dailyViews,
	}, nil
}

// GetSubscriberMetrics returns subscriber and financial metrics
func (db *database.DB) GetSubscriberMetrics() (map[string]interface{}, error) {
	// Get total subscribers
	var totalSubscribers int
	err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE role != 'admin'`).Scan(&totalSubscribers)
	if err != nil {
		return nil, err
	}

	// Get active subscriptions
	var activeSubscriptions int
	err = db.QueryRow(`SELECT COUNT(*) FROM subscriptions WHERE status = 'active'`).Scan(&activeSubscriptions)
	if err != nil {
		return nil, err
	}

	// Get REAL monthly revenue from subscription_plans table joined with subscriptions
	var monthlyRevenue float64
	err = db.QueryRow(`
		SELECT COALESCE(SUM(
			CASE 
				-- Convert all to monthly equivalent
				WHEN sp.interval = 'month' THEN sp.price
				WHEN sp.interval = 'year' THEN sp.price / 12.0
				ELSE sp.price  -- Default to monthly if unclear
			END
		), 0)
		FROM subscriptions s
		INNER JOIN subscription_plans sp ON s.plan_id = sp.id
		WHERE s.status = 'active' AND sp.is_active = true
	`).Scan(&monthlyRevenue)
	if err != nil {
		log.Printf("⚠️ Warning: Could not calculate monthly revenue from subscription_plans: %v", err)
		// Fallback: Try to estimate from active subscriptions
		monthlyRevenue = 0
	}

	// Calculate churn rate (simplified)
	var churnRate float64
	var cancelledThisMonth int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM subscriptions 
		WHERE status = 'cancelled' 
		AND updated_at >= CURRENT_DATE - INTERVAL '30 days'
	`).Scan(&cancelledThisMonth)
	if err == nil && activeSubscriptions > 0 {
		churnRate = float64(cancelledThisMonth) / float64(activeSubscriptions) * 100
	}

	return map[string]interface{}{
		"total_subscribers":    totalSubscribers,
		"active_subscriptions": activeSubscriptions,
		"monthly_revenue":      monthlyRevenue,
		"churn_rate":           churnRate,
	}, nil
}

// GetVideoStats returns comprehensive video statistics
func (db *database.DB) GetVideoStats() (map[string]interface{}, error) {
	// Get total videos
	var totalVideos int
	err := db.QueryRow(`SELECT COUNT(*) FROM master_video_list`).Scan(&totalVideos)
	if err != nil {
		return nil, err
	}

	// Get synced videos
	var syncedVideos int
	err = db.QueryRow(`SELECT COUNT(*) FROM master_video_list WHERE sync_status = 'synced'`).Scan(&syncedVideos)
	if err != nil {
		return nil, err
	}

	// Get videos needing attention
	var needsAttention int
	err = db.QueryRow(`SELECT COUNT(*) FROM master_video_list WHERE sync_status = 'needs_attention'`).Scan(&needsAttention)
	if err != nil {
		return nil, err
	}

	// Get total views
	var totalViews int64
	err = db.QueryRow(`SELECT COALESCE(SUM(views), 0) FROM master_video_list`).Scan(&totalViews)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_videos":    totalVideos,
		"synced_videos":   syncedVideos,
		"needs_attention": needsAttention,
		"total_views":     totalViews,
	}, nil
}
