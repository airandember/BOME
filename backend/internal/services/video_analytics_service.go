package services

import (
	"bome-backend/internal/database"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// VideoAnalyticsService handles video viewing analytics
type VideoAnalyticsService struct {
	db     *database.DB
	buffer *AnalyticsBuffer // Async event buffer
	redis  *database.Redis  // For query caching
}

// NewVideoAnalyticsService creates a new video analytics service
func NewVideoAnalyticsService(db *database.DB, redis *database.Redis) *VideoAnalyticsService {
	var buffer *AnalyticsBuffer
	if redis != nil {
		buffer = NewAnalyticsBuffer(db, redis)
		buffer.StartFlusher() // Start background flusher
		log.Printf("✅ [Video Analytics] Async buffer enabled with Redis")
	} else {
		log.Printf("⚠️ [Video Analytics] Redis not available, using direct DB writes")
	}

	return &VideoAnalyticsService{
		db:     db,
		buffer: buffer,
		redis:  redis,
	}
}

// VideoTrackingRequest represents a video tracking event from the frontend
type VideoTrackingRequest struct {
	VideoID           int     `json:"video_id" binding:"required"`
	UserID            *int    `json:"user_id"`                               // NULL for anonymous
	SessionID         string  `json:"session_id"`                            // Required for anonymous
	WatchedDuration   int     `json:"watched_duration" binding:"required"`   // Seconds
	WatchedPercentage float64 `json:"watched_percentage" binding:"required"` // 0-100
	IPAddress         string  `json:"ip_address"`                            // Optional, can be set server-side
	UserAgent         string  `json:"user_agent"`                            // Optional, can be set server-side
}

// VideoStats represents aggregated statistics for a video
type VideoStats struct {
	VideoID         int       `json:"video_id"`
	TotalViews      int       `json:"total_views"`
	UniqueViewers   int       `json:"unique_viewers"`
	TotalWatchTime  int       `json:"total_watch_time_seconds"` // Total seconds
	AvgWatchTime    float64   `json:"avg_watch_time_seconds"`
	AvgPercentage   float64   `json:"avg_percentage_watched"`
	CompletionRate  float64   `json:"completion_rate"`  // % who watched >= 95%
	BounceRate      float64   `json:"bounce_rate"`      // % who left < 10 seconds
	EngagementScore float64   `json:"engagement_score"` // 0-100
	Period          string    `json:"period"`           // e.g., "7d", "30d"
	LastViewedAt    time.Time `json:"last_viewed_at"`
}

// TrendingVideo represents a video in the trending list
type TrendingVideo struct {
	VideoID       int     `json:"video_id"`
	BunnyVideoID  string  `json:"bunny_video_id"`
	Title         string  `json:"title"`
	ThumbnailURL  string  `json:"thumbnail_url"` // Deprecated: kept for backward compatibility
	Last24HViews  int     `json:"last_24h_views"`
	TrendingScore float64 `json:"trending_score"`
}

// RecordView records a video view event using async Redis buffer
// Returns immediately without blocking on database writes
func (s *VideoAnalyticsService) RecordView(req VideoTrackingRequest) error {
	log.Printf("🎯 [SERVICE] ========================================")
	log.Printf("🎯 [SERVICE] RecordView called")
	log.Printf("🎯 [SERVICE] Video: %d, User: %v, Duration: %ds, Percentage: %.2f%%",
		req.VideoID, req.UserID, req.WatchedDuration, req.WatchedPercentage)

	// Use async buffer if available (production mode)
	if s.buffer != nil {
		log.Printf("📤 [SERVICE→BUFFER] Buffer available, adding event to Redis")
		err := s.buffer.AddEvent(req)
		if err != nil {
			// Buffer failed, fall back to direct DB write
			log.Printf("⚠️ [SERVICE] Buffer.AddEvent failed: %v", err)
			log.Printf("🔄 [SERVICE] Falling back to direct DB write")
			return s.recordViewDirect(req)
		}
		log.Printf("✅ [SERVICE←BUFFER] Event buffered successfully (async)")
		log.Printf("⚡ [SERVICE] Returning immediately (non-blocking)")
		log.Printf("🎯 [SERVICE] ========================================")
		return nil
	}

	// No buffer available, write directly (fallback mode)
	log.Printf("⚠️ [SERVICE] No buffer available (Redis not configured)")
	log.Printf("📤 [SERVICE→DB] Using direct DB write (fallback)")
	result := s.recordViewDirect(req)
	log.Printf("🎯 [SERVICE] ========================================")
	return result
}

// recordViewDirect performs direct synchronous DB write (fallback)
func (s *VideoAnalyticsService) recordViewDirect(req VideoTrackingRequest) error {
	log.Printf("💾 [SERVICE-DB] Direct DB write started")
	log.Printf("💾 [SERVICE-DB] Completed status: %.2f%% >= 95%% = %v", req.WatchedPercentage, req.WatchedPercentage >= 95.0)

	completed := req.WatchedPercentage >= 95.0

	var query string
	var args []interface{}

	if req.UserID != nil {
		log.Printf("💾 [SERVICE-DB] Using authenticated user path (user_id=%d)", *req.UserID)
		// Authenticated user - UPSERT on (user_id, video_id)
		query = `
			INSERT INTO watch_history (
				user_id, video_id, last_position, progress_percentage, 
				total_watch_time, view_count, completed,
				first_watched_at, last_watched_at, created_at
			) VALUES ($1, $2, $3, $4, $3, 1, $5, NOW(), NOW(), NOW())
			ON CONFLICT (user_id, video_id) 
			WHERE user_id IS NOT NULL
			DO UPDATE SET
				last_position = EXCLUDED.last_position,
				progress_percentage = EXCLUDED.progress_percentage,
				total_watch_time = GREATEST(watch_history.total_watch_time, EXCLUDED.last_position),
				view_count = CASE 
					WHEN watch_history.last_watched_at < NOW() - INTERVAL '30 minutes' 
					THEN watch_history.view_count + 1 
					ELSE watch_history.view_count 
				END,
				completed = watch_history.completed OR EXCLUDED.completed,
				last_watched_at = NOW()
		`
		args = []interface{}{req.UserID, req.VideoID, req.WatchedDuration, req.WatchedPercentage, completed}
	} else {
		log.Printf("💾 [SERVICE-DB] Using anonymous user path (session_id=%s)", req.SessionID)
		// Anonymous user - UPSERT on (session_id, video_id)
		query = `
			INSERT INTO watch_history (
				session_id, video_id, last_position, progress_percentage, 
				total_watch_time, view_count, completed,
				first_watched_at, last_watched_at, created_at
			) VALUES ($1, $2, $3, $4, $3, 1, $5, NOW(), NOW(), NOW())
			ON CONFLICT (session_id, video_id) 
			WHERE session_id IS NOT NULL
			DO UPDATE SET
				last_position = EXCLUDED.last_position,
				progress_percentage = EXCLUDED.progress_percentage,
				total_watch_time = GREATEST(watch_history.total_watch_time, EXCLUDED.last_position),
				view_count = CASE 
					WHEN watch_history.last_watched_at < NOW() - INTERVAL '30 minutes' 
					THEN watch_history.view_count + 1 
					ELSE watch_history.view_count 
				END,
				completed = watch_history.completed OR EXCLUDED.completed,
				last_watched_at = NOW()
		`
		args = []interface{}{req.SessionID, req.VideoID, req.WatchedDuration, req.WatchedPercentage, completed}
	}

	log.Printf("💾 [SERVICE-DB] Executing UPSERT to watch_history table")
	_, err := s.db.Exec(query, args...)
	if err != nil {
		log.Printf("❌ [SERVICE-DB] Database UPSERT failed: %v", err)
		return fmt.Errorf("failed to record view: %w", err)
	}

	log.Printf("✅ [SERVICE-DB] UPSERT successful - watch_history updated")
	return nil
}

// GetVideoStats retrieves aggregated statistics for a video
func (s *VideoAnalyticsService) GetVideoStats(videoID int, period string) (*VideoStats, error) {
	log.Printf("📊 [Video Analytics] Getting stats for video %d (period: %s)", videoID, period)

	// Parse period into time range
	var startTime time.Time
	switch period {
	case "24h":
		startTime = time.Now().AddDate(0, 0, -1)
	case "7d":
		startTime = time.Now().AddDate(0, 0, -7)
	case "30d":
		startTime = time.Now().AddDate(0, 0, -30)
	case "90d":
		startTime = time.Now().AddDate(0, 0, -90)
	case "all":
		startTime = time.Time{} // Zero time = all time
	default:
		startTime = time.Now().AddDate(0, 0, -7) // Default to 7 days
		period = "7d"
	}

	query := `
		SELECT 
			SUM(view_count) AS total_views,
			COUNT(DISTINCT COALESCE(user_id::text, session_id)) AS unique_viewers,
			SUM(total_watch_time) AS total_watch_time,
			AVG(total_watch_time) AS avg_watch_time,
			AVG(progress_percentage) AS avg_percentage,
			COUNT(CASE WHEN completed = true THEN 1 END)::FLOAT / 
				NULLIF(COUNT(*), 0)::FLOAT * 100 AS completion_rate,
			COUNT(CASE WHEN total_watch_time <= 10 THEN 1 END)::FLOAT / 
				NULLIF(COUNT(*), 0)::FLOAT * 100 AS bounce_rate,
			MAX(last_watched_at) AS last_viewed_at
		FROM watch_history
		WHERE video_id = $1
		AND ($2::timestamp IS NULL OR last_watched_at >= $2)
	`

	var stats VideoStats
	var totalWatchTime, avgWatchTime sql.NullFloat64
	var lastViewedAt sql.NullTime

	err := s.db.QueryRow(query, videoID, startTime).Scan(
		&stats.TotalViews,
		&stats.UniqueViewers,
		&totalWatchTime,
		&avgWatchTime,
		&stats.AvgPercentage,
		&stats.CompletionRate,
		&stats.BounceRate,
		&lastViewedAt,
	)

	if err != nil {
		log.Printf("❌ [Video Analytics] Failed to get stats: %v", err)
		return nil, fmt.Errorf("failed to get video stats: %w", err)
	}

	stats.VideoID = videoID
	stats.Period = period
	stats.TotalWatchTime = int(totalWatchTime.Float64)
	stats.AvgWatchTime = avgWatchTime.Float64
	if lastViewedAt.Valid {
		stats.LastViewedAt = lastViewedAt.Time
	}

	// Calculate engagement score (0-100)
	// Formula: (completion_rate * 0.4) + (avg_percentage * 0.3) + ((100 - bounce_rate) * 0.3)
	stats.EngagementScore = (stats.CompletionRate * 0.4) +
		(stats.AvgPercentage * 0.3) +
		((100 - stats.BounceRate) * 0.3)

	log.Printf("✅ [Video Analytics] Stats retrieved: views=%d, unique=%d, engagement=%.2f",
		stats.TotalViews, stats.UniqueViewers, stats.EngagementScore)

	return &stats, nil
}

// GetTrendingVideos retrieves trending videos based on recent activity
// HYBRID APPROACH: Uses detailed analytics if available, falls back to master_video_list.views
// CACHED: Results cached in Redis for 5 minutes
func (s *VideoAnalyticsService) GetTrendingVideos(limit int) ([]TrendingVideo, error) {
	startTime := time.Now()
	log.Printf("📊 [Video Analytics] ========================================")
	log.Printf("📊 [Video Analytics] GetTrendingVideos called with limit=%d", limit)

	// Try Redis cache first
	cacheKey := fmt.Sprintf("analytics:trending:%d", limit)
	if s.redis != nil {
		log.Printf("📊 [Video Analytics] Checking Redis cache: %s", cacheKey)
		if cached, err := s.getFromCache(cacheKey); err == nil {
			duration := time.Since(startTime)
			log.Printf("✅ [Video Analytics] Cache HIT! Returning cached data in %v", duration)
			log.Printf("📊 [Video Analytics] ========================================")
			return cached.([]TrendingVideo), nil
		}
		log.Printf("📊 [Video Analytics] Cache MISS - fetching from database")
	} else {
		log.Printf("⚠️  [Video Analytics] Redis not available - no caching")
	}

	if limit <= 0 || limit > 100 {
		log.Printf("📊 [Video Analytics] Adjusting limit from %d to 100 (max)", limit)
		limit = 100 // Default limit - top 100 trending
	}

	// Use optimized fallback query directly
	// master_video_list.views is auto-synced from watch_history via trigger
	// This is fast (<100ms) and accurate
	log.Printf("⚡ [Video Analytics] Using optimized query (master_video_list.views)")

	videos, err := s.getFallbackTrendingVideos(limit)
	if err != nil {
		duration := time.Since(startTime)
		log.Printf("❌ [Video Analytics] Failed after %v: %v", duration, err)
		log.Printf("📊 [Video Analytics] ========================================")
		return nil, err
	}

	// Cache results for 5 minutes
	if s.redis != nil && len(videos) > 0 {
		log.Printf("📊 [Video Analytics] Caching %d videos in Redis for 5 minutes", len(videos))
		s.setCache(cacheKey, videos, 5*time.Minute)
	}

	duration := time.Since(startTime)
	log.Printf("✅ [Video Analytics] Successfully returned %d videos in %v", len(videos), duration)
	log.Printf("📊 [Video Analytics] ========================================")
	return videos, nil
}

// getFallbackTrendingVideos returns trending videos based purely on master_video_list.views
// Used as a fallback when the complex analytics query times out
func (s *VideoAnalyticsService) getFallbackTrendingVideos(limit int) ([]TrendingVideo, error) {
	startTime := time.Now()
	log.Printf("⚡ [Video Analytics] ========================================")
	log.Printf("⚡ [Video Analytics] getFallbackTrendingVideos started")
	log.Printf("⚡ [Video Analytics] Requested limit: %d", limit)

	query := `
		SELECT 
			v.id AS video_id,
			v.title,
			v.bunny_video_id,
			GREATEST(v.views / 10, 1) AS last_24h_views,
			v.updated_at AS last_view_at,
			0 AS completion_rate,
			v.likes
		FROM master_video_list v
		WHERE v.status = 'ready' AND v.views > 0
		ORDER BY v.views DESC, v.updated_at DESC
		LIMIT $1
	`

	log.Printf("⚡ [Video Analytics] Creating context with 5s timeout")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Printf("⚡ [Video Analytics] Executing query against master_video_list...")
	queryStart := time.Now()
	rows, err := s.db.QueryContext(ctx, query, limit)
	queryDuration := time.Since(queryStart)

	if err != nil {
		log.Printf("❌ [Video Analytics] Query failed after %v: %v", queryDuration, err)
		log.Printf("❌ [Video Analytics] Context error: %v", ctx.Err())
		log.Printf("⚡ [Video Analytics] ========================================")
		return []TrendingVideo{}, nil // Return empty array rather than error
	}
	defer rows.Close()
	log.Printf("✅ [Video Analytics] Query completed in %v", queryDuration)

	log.Printf("⚡ [Video Analytics] Scanning rows...")
	var trending []TrendingVideo
	rowCount := 0
	for rows.Next() {
		rowCount++
		var video TrendingVideo
		var lastViewAt time.Time
		var completionRate float64
		var likes int

		err := rows.Scan(
			&video.VideoID,
			&video.Title,
			&video.BunnyVideoID,
			&video.Last24HViews,
			&lastViewAt,
			&completionRate,
			&likes,
		)
		if err != nil {
			log.Printf("⚠️  [Video Analytics] Error scanning row %d: %v", rowCount, err)
			continue
		}

		// Generate thumbnail URL (frontend will construct full path using VITE_API_BASE_URL)
		video.ThumbnailURL = fmt.Sprintf("videos/%s/thumbnail", video.BunnyVideoID)

		// Simple score based on views and recency
		daysSinceUpdate := time.Since(lastViewAt).Hours() / 24.0
		recencyBonus := 100.0 / (1.0 + daysSinceUpdate/30.0) // Decay over 30 days
		video.TrendingScore = (float64(video.Last24HViews) * 10) + recencyBonus

		trending = append(trending, video)

		if rowCount <= 3 {
			log.Printf("⚡ [Video Analytics] Row %d: ID=%d, Title=%s, Views=%d, Score=%.2f",
				rowCount, video.VideoID, video.Title, video.Last24HViews, video.TrendingScore)
		}
	}

	totalDuration := time.Since(startTime)
	log.Printf("✅ [Video Analytics] Scanned %d rows successfully", rowCount)
	log.Printf("✅ [Video Analytics] Returning %d trending videos", len(trending))
	log.Printf("✅ [Video Analytics] Total duration: %v", totalDuration)
	log.Printf("⚡ [Video Analytics] ========================================")
	return trending, nil
}

// GetUserEngagement retrieves engagement metrics for a specific user
func (s *VideoAnalyticsService) GetUserEngagement(userID int, days int) (map[string]interface{}, error) {
	log.Printf("📊 [Video Analytics] Getting engagement for user %d (last %d days)", userID, days)

	if days <= 0 {
		days = 30 // Default to 30 days
	}

	query := `
		SELECT 
			COUNT(DISTINCT video_id) AS videos_watched,
			SUM(view_count) AS total_views,
			SUM(total_watch_time) AS total_watch_time_seconds,
			AVG(progress_percentage) AS avg_completion,
			COUNT(CASE WHEN completed = true THEN 1 END) AS completed_videos,
			MIN(first_watched_at) AS first_view,
			MAX(last_watched_at) AS last_view
		FROM watch_history
		WHERE user_id = $1
		AND last_watched_at > NOW() - INTERVAL '1 day' * $2
	`

	var videosWatched, totalViews, completedVideos int
	var totalWatchTime sql.NullFloat64
	var avgCompletion sql.NullFloat64
	var firstView, lastView sql.NullTime

	err := s.db.QueryRow(query, userID, days).Scan(
		&videosWatched,
		&totalViews,
		&totalWatchTime,
		&avgCompletion,
		&completedVideos,
		&firstView,
		&lastView,
	)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("❌ [Video Analytics] Failed to get user engagement: %v", err)
		return nil, fmt.Errorf("failed to get user engagement: %w", err)
	}

	engagement := map[string]interface{}{
		"user_id":                   userID,
		"period_days":               days,
		"videos_watched":            videosWatched,
		"total_views":               totalViews,
		"total_watch_time_seconds":  int(totalWatchTime.Float64),
		"total_watch_time_hours":    totalWatchTime.Float64 / 3600,
		"avg_completion_percentage": avgCompletion.Float64,
		"completed_videos":          completedVideos,
		"completion_rate":           0.0,
	}

	if videosWatched > 0 {
		engagement["completion_rate"] = float64(completedVideos) / float64(videosWatched) * 100
	}

	if firstView.Valid {
		engagement["first_view"] = firstView.Time
	}
	if lastView.Valid {
		engagement["last_view"] = lastView.Time
	}

	log.Printf("✅ [Video Analytics] User engagement: %d videos, %.2f hours watched",
		videosWatched, totalWatchTime.Float64/3600)

	return engagement, nil
}

// GetTopVideos retrieves the most viewed videos in a time period
// HYBRID APPROACH: Uses detailed analytics if available, falls back to master_video_list.views
// CACHED: Results cached in Redis for 10 minutes
func (s *VideoAnalyticsService) GetTopVideos(limit int, days int) ([]map[string]interface{}, error) {
	log.Printf("📊 [Video Analytics] Getting top %d videos (last %d days)", limit, days)

	// Try Redis cache first
	cacheKey := fmt.Sprintf("analytics:top_videos:%d:%d", limit, days)
	if s.redis != nil {
		if cached, err := s.getFromCache(cacheKey); err == nil {
			log.Printf("✅ [Video Analytics] Returning cached top videos")
			return cached.([]map[string]interface{}), nil
		}
	}

	if limit <= 0 || limit > 100 {
		limit = 100 // Default to top 100
	}
	if days <= 0 {
		days = 30
	}

	// HYBRID QUERY: Use watch_history (optimized) + master_video_list.views fallback
	query := `
		WITH analytics_views AS (
			SELECT 
				wh.video_id,
				COUNT(DISTINCT COALESCE(wh.user_id::text, wh.session_id)) AS unique_viewers,
				AVG(wh.progress_percentage) AS avg_completion,
				SUM(wh.total_watch_time) AS total_watch_time,
				COUNT(*) FILTER (WHERE wh.last_watched_at > NOW() - INTERVAL '1 day' * $2) AS recent_views
			FROM watch_history wh
			GROUP BY wh.video_id
	)
	SELECT 
		v.id,
		v.title,
		v.bunny_video_id,
		v.duration,
		-- Use analytics data if available, otherwise use master_video_list.views
		COALESCE(av.unique_viewers, v.views, 0) AS total_views,
		COALESCE(av.unique_viewers, v.views, 0) AS unique_viewers,
		COALESCE(av.avg_completion, 0.0) AS avg_completion,
		COALESCE(av.total_watch_time, 0) AS total_watch_time
	FROM master_video_list v
		LEFT JOIN analytics_views av ON av.video_id = v.id
		WHERE v.status = 'ready' AND (av.unique_viewers > 0 OR v.views > 0)
		ORDER BY total_views DESC
		LIMIT $1
	`

	rows, err := s.db.Query(query, limit, days)
	if err != nil {
		log.Printf("❌ [Video Analytics] Failed to get top videos: %v", err)
		return nil, fmt.Errorf("failed to get top videos: %w", err)
	}
	defer rows.Close()

	var videos []map[string]interface{}
	for rows.Next() {
		var id, duration, totalViews, uniqueViewers, totalWatchTime int
		var title, bunnyVideoID string
		var avgCompletion float64

		err := rows.Scan(&id, &title, &bunnyVideoID, &duration, &totalViews, &uniqueViewers, &avgCompletion, &totalWatchTime)
		if err != nil {
			log.Printf("⚠️  [Video Analytics] Error scanning video: %v", err)
			continue
		}

		// Generate relative thumbnail URL (frontend constructs full path)
		thumbnailURL := fmt.Sprintf("videos/%s/thumbnail", bunnyVideoID)

		videos = append(videos, map[string]interface{}{
			"video_id":         id,
			"bunny_video_id":   bunnyVideoID,
			"title":            title,
			"thumbnail_url":    thumbnailURL,
			"duration":         duration,
			"total_views":      totalViews,
			"unique_viewers":   uniqueViewers,
			"avg_completion":   avgCompletion,
			"total_watch_time": totalWatchTime,
		})
	}

	// Cache results for 10 minutes
	if s.redis != nil {
		s.setCache(cacheKey, videos, 10*time.Minute)
	}

	log.Printf("✅ [Video Analytics] Retrieved %d top videos (using hybrid query)", len(videos))
	return videos, nil
}

// getFromCache retrieves data from Redis cache
func (s *VideoAnalyticsService) getFromCache(key string) (interface{}, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("redis not available")
	}

	ctx := context.Background()
	data, err := s.redis.Client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var result interface{}
	err = json.Unmarshal([]byte(data), &result)
	return result, err
}

// setCache stores data in Redis cache
func (s *VideoAnalyticsService) setCache(key string, value interface{}, expiration time.Duration) {
	if s.redis == nil {
		return
	}

	ctx := context.Background()
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("⚠️ [Video Analytics] Failed to marshal cache data: %v", err)
		return
	}

	err = s.redis.Client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		log.Printf("⚠️ [Video Analytics] Failed to set cache: %v", err)
	}
}

// InvalidateCache removes analytics cache entries
func (s *VideoAnalyticsService) InvalidateCache() {
	if s.redis == nil {
		return
	}

	ctx := context.Background()
	pattern := "analytics:*"
	keys, err := s.redis.Client.Keys(ctx, pattern).Result()
	if err != nil {
		log.Printf("⚠️ [Video Analytics] Failed to get cache keys: %v", err)
		return
	}

	if len(keys) > 0 {
		err = s.redis.Client.Del(ctx, keys...).Err()
		if err != nil {
			log.Printf("⚠️ [Video Analytics] Failed to invalidate cache: %v", err)
		} else {
			log.Printf("✅ [Video Analytics] Invalidated %d cache entries", len(keys))
		}
	}
}

// Stop gracefully stops the analytics service
func (s *VideoAnalyticsService) Stop() {
	if s.buffer != nil {
		s.buffer.StopFlusher()
	}
}
