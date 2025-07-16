package database

import (
	"fmt"
	"sync"
	"time"
)

// GetVideoCount returns the total number of videos
func (db *DB) GetVideoCount() (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM videos`).Scan(&count)
	return count, err
}

// GetTotalViews returns the total view count across all videos
func (db *DB) GetTotalViews() (int64, error) {
	var total int64
	err := db.QueryRow(`SELECT COALESCE(SUM(view_count), 0) FROM videos`).Scan(&total)
	return total, err
}

// GetTotalLikes returns the total like count across all videos
func (db *DB) GetTotalLikes() (int64, error) {
	var total int64
	err := db.QueryRow(`SELECT COALESCE(SUM(like_count), 0) FROM videos`).Scan(&total)
	return total, err
}

// Activity represents a system activity
type Activity struct {
	Type      string    `json:"type"`
	UserID    int       `json:"user_id"`
	Action    string    `json:"action"`
	Details   string    `json:"details"`
	CreatedAt time.Time `json:"created_at"`
}

// GetRecentActivity returns recent system activity
func (db *DB) GetRecentActivity(limit int) ([]*Activity, error) {
	query := `
        SELECT 'video' as type, user_id, action, details, created_at 
        FROM audit_logs 
        WHERE resource = 'video'
        UNION ALL
        SELECT 'user' as type, user_id, action, details, created_at 
        FROM audit_logs 
        WHERE resource = 'user'
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

// SystemHealth represents the system's health metrics
type SystemHealth struct {
	DBStatus    string                 `json:"db_status"`
	DBLatency   float64                `json:"db_latency"`
	CacheStatus string                 `json:"cache_status"`
	DiskUsage   float64                `json:"disk_usage"`
	MemoryUsage float64                `json:"memory_usage"`
	CPUUsage    float64                `json:"cpu_usage"`
	Metrics     map[string]interface{} `json:"metrics"`
}

// GetSystemHealth returns the current system health metrics
func (db *DB) GetSystemHealth() (*SystemHealth, error) {
	// Check database connectivity
	start := time.Now()
	err := db.QueryRow("SELECT 1").Scan(new(int))
	latency := time.Since(start).Seconds()

	if err != nil {
		return nil, fmt.Errorf("database health check failed: %v", err)
	}

	// Get some basic metrics
	var userCount, videoCount int
	var totalViews, totalLikes int64

	// Run these in parallel using goroutines
	var wg sync.WaitGroup
	var userErr, videoErr, viewsErr, likesErr error

	wg.Add(4)
	go func() {
		defer wg.Done()
		userCount, userErr = db.GetUserCount()
	}()
	go func() {
		defer wg.Done()
		videoCount, videoErr = db.GetVideoCount()
	}()
	go func() {
		defer wg.Done()
		totalViews, viewsErr = db.GetTotalViews()
	}()
	go func() {
		defer wg.Done()
		totalLikes, likesErr = db.GetTotalLikes()
	}()

	wg.Wait()

	// Check for errors
	if userErr != nil || videoErr != nil || viewsErr != nil || likesErr != nil {
		return nil, fmt.Errorf("failed to gather metrics")
	}

	return &SystemHealth{
		DBStatus:    "healthy",
		DBLatency:   latency,
		CacheStatus: "operational",
		DiskUsage:   0.0, // These would come from OS metrics in production
		MemoryUsage: 0.0,
		CPUUsage:    0.0,
		Metrics: map[string]interface{}{
			"total_users":  userCount,
			"total_videos": videoCount,
			"total_views":  totalViews,
			"total_likes":  totalLikes,
			"last_checked": time.Now(),
		},
	}, nil
}
