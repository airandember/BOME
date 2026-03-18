package models

import (
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

// Note: Database methods for analytics are defined in internal/database/analytics.go
// This file only contains model type definitions.
