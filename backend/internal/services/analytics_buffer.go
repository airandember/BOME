package services

import (
	"bome-backend/internal/database"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// AnalyticsBuffer handles async buffering of video analytics events
// Events are buffered in Redis and flushed to PostgreSQL in batches
type AnalyticsBuffer struct {
	redis         *database.Redis
	db            *database.DB
	bufferKey     string
	batchSize     int
	flushInterval time.Duration
	ticker        *time.Ticker
	done          chan bool
	mu            sync.Mutex
	isRunning     bool
	stats         BufferStats
}

// BufferStats tracks buffer performance metrics
type BufferStats struct {
	EventsReceived   int64
	EventsFlushed    int64
	EventsDropped    int64
	FlushCount       int64
	LastFlushTime    time.Time
	LastFlushSize    int
	AvgFlushDuration time.Duration
}

// NewAnalyticsBuffer creates a new analytics buffer
func NewAnalyticsBuffer(db *database.DB, redis *database.Redis) *AnalyticsBuffer {
	return &AnalyticsBuffer{
		redis:         redis,
		db:            db,
		bufferKey:     "analytics:video_tracking_buffer",
		batchSize:     100,             // Flush when 100 events accumulated
		flushInterval: 5 * time.Second, // Or flush every 5 seconds
		done:          make(chan bool),
		stats:         BufferStats{},
	}
}

// AddEvent adds a tracking event to the Redis buffer (non-blocking)
func (b *AnalyticsBuffer) AddEvent(event VideoTrackingRequest) error {
	log.Printf("📦 [BUFFER] ======================================")
	log.Printf("📦 [BUFFER] AddEvent called: video=%d, user=%v", event.VideoID, event.UserID)

	// Quick validation
	if event.VideoID == 0 {
		log.Printf("❌ [BUFFER] Validation failed: invalid video_id")
		return fmt.Errorf("invalid video_id")
	}
	if event.UserID == nil && event.SessionID == "" {
		log.Printf("❌ [BUFFER] Validation failed: missing user_id and session_id")
		return fmt.Errorf("either user_id or session_id required")
	}
	log.Printf("✅ [BUFFER] Validation passed")

	// Serialize event
	log.Printf("🔄 [BUFFER] Serializing event to JSON")
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("❌ [BUFFER] JSON marshal failed: %v", err)
		b.incrementDropped()
		return err
	}
	log.Printf("✅ [BUFFER] Event serialized (%d bytes)", len(data))

	// Check if Redis is available
	if b.redis == nil {
		log.Printf("❌ [BUFFER] Redis client is nil - buffer unavailable")
		b.incrementDropped()
		return fmt.Errorf("redis not available")
	}

	// Push to Redis list (non-blocking, returns immediately)
	log.Printf("📤 [BUFFER→REDIS] Pushing to Redis list: %s", b.bufferKey)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = b.redis.Client.RPush(ctx, b.bufferKey, data).Err()
	if err != nil {
		log.Printf("❌ [BUFFER] Redis RPUSH failed: %v", err)
		b.incrementDropped()
		return err
	}
	log.Printf("✅ [BUFFER←REDIS] Event pushed to Redis successfully")

	b.incrementReceived()

	// Check buffer size and flush if needed
	size, _ := b.redis.Client.LLen(context.Background(), b.bufferKey).Result()
	log.Printf("📊 [BUFFER] Current buffer size: %d/%d", size, b.batchSize)

	if size >= int64(b.batchSize) {
		log.Printf("🚀 [BUFFER] Buffer full (%d >= %d) - triggering async flush", size, b.batchSize)
		// Buffer is full, trigger immediate flush (async)
		go b.FlushBatch()
	} else {
		log.Printf("⏳ [BUFFER] Buffer not full yet - waiting for timer or batch size")
	}

	log.Printf("📦 [BUFFER] ======================================")
	return nil
}

// StartFlusher starts the background flusher goroutine
func (b *AnalyticsBuffer) StartFlusher() {
	b.mu.Lock()
	if b.isRunning {
		b.mu.Unlock()
		log.Printf("⚠️ [Analytics Buffer] Flusher already running")
		return
	}
	b.isRunning = true
	b.mu.Unlock()

	log.Printf("🚀 [Analytics Buffer] Starting flusher (interval: %v, batch size: %d)", b.flushInterval, b.batchSize)

	b.ticker = time.NewTicker(b.flushInterval)
	go func() {
		for {
			select {
			case <-b.ticker.C:
				b.FlushBatch()
			case <-b.done:
				log.Printf("🛑 [Analytics Buffer] Flusher stopped")
				return
			}
		}
	}()
}

// StopFlusher stops the background flusher gracefully
func (b *AnalyticsBuffer) StopFlusher() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.isRunning {
		return
	}

	log.Printf("⏸️ [Analytics Buffer] Stopping flusher...")
	b.ticker.Stop()
	b.done <- true
	b.isRunning = false

	// Flush any remaining events
	b.FlushBatch()
	log.Printf("✅ [Analytics Buffer] Flusher stopped gracefully")
}

// FlushBatch processes a batch of events from Redis to PostgreSQL
func (b *AnalyticsBuffer) FlushBatch() error {
	log.Printf("🔥 [BUFFER-FLUSH] ====================================")
	log.Printf("🔥 [BUFFER-FLUSH] FlushBatch triggered")
	startTime := time.Now()

	if b.redis == nil {
		log.Printf("❌ [BUFFER-FLUSH] Redis not available - aborting")
		return fmt.Errorf("redis not available")
	}

	ctx := context.Background()

	// Pop up to batchSize events from Redis
	log.Printf("📥 [BUFFER-FLUSH←REDIS] Reading batch from Redis (max %d events)", b.batchSize)
	events, err := b.redis.Client.LRange(ctx, b.bufferKey, 0, int64(b.batchSize-1)).Result()
	if err != nil {
		log.Printf("❌ [BUFFER-FLUSH] Redis LRange failed: %v", err)
		return err
	}

	if len(events) == 0 {
		log.Printf("ℹ️  [BUFFER-FLUSH] No events to flush")
		return nil
	}

	log.Printf("📦 [BUFFER-FLUSH] Retrieved %d events from Redis", len(events))

	// Remove events from Redis buffer (we've got them in memory now)
	log.Printf("✂️  [BUFFER-FLUSH→REDIS] Trimming processed events from Redis")
	b.redis.Client.LTrim(ctx, b.bufferKey, int64(len(events)), -1)

	// Process each event with UPSERT
	log.Printf("💾 [BUFFER-FLUSH→DB] Starting batch UPSERT to watch_history")
	successCount := 0
	errorCount := 0

	for i, eventJSON := range events {
		log.Printf("💾 [BUFFER-FLUSH] Processing event %d/%d", i+1, len(events))
		var event VideoTrackingRequest
		if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
			log.Printf("⚠️ [Analytics Buffer] Failed to unmarshal event: %v", err)
			errorCount++
			continue
		}

		// Perform UPSERT
		if err := b.upsertWatchHistory(event); err != nil {
			log.Printf("⚠️ [Analytics Buffer] Failed to upsert event: %v", err)
			errorCount++
			continue
		}

		successCount++
	}

	// Update stats
	duration := time.Since(startTime)

	b.mu.Lock()
	b.stats.EventsFlushed += int64(successCount)
	b.stats.EventsDropped += int64(errorCount)
	b.stats.FlushCount++
	b.stats.LastFlushTime = time.Now()
	b.stats.LastFlushSize = len(events)

	// Calculate rolling average flush duration
	if b.stats.AvgFlushDuration == 0 {
		b.stats.AvgFlushDuration = duration
	} else {
		b.stats.AvgFlushDuration = (b.stats.AvgFlushDuration + duration) / 2
	}
	b.mu.Unlock()

	log.Printf("✅ [BUFFER-FLUSH] Batch complete in %v", duration)
	log.Printf("📊 [BUFFER-FLUSH] Results: %d success, %d errors (total: %d)", successCount, errorCount, len(events))
	log.Printf("📈 [BUFFER-FLUSH] Stats: total_flushed=%d, total_dropped=%d, flush_count=%d",
		b.stats.EventsFlushed, b.stats.EventsDropped, b.stats.FlushCount)
	log.Printf("🔥 [BUFFER-FLUSH] ====================================")

	return nil
}

// upsertWatchHistory performs the UPSERT operation for a single event
func (b *AnalyticsBuffer) upsertWatchHistory(req VideoTrackingRequest) error {
	completed := req.WatchedPercentage >= 95.0

	var query string
	var args []interface{}

	if req.UserID != nil {
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

	_, err := b.db.Exec(query, args...)
	return err
}

// GetStats returns buffer statistics
func (b *AnalyticsBuffer) GetStats() BufferStats {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Get current buffer size from Redis
	if b.redis != nil {
		size, err := b.redis.Client.LLen(context.Background(), b.bufferKey).Result()
		if err == nil {
			log.Printf("📊 [Analytics Buffer] Current buffer size: %d events", size)
		}
	}

	return b.stats
}

// incrementReceived safely increments received counter
func (b *AnalyticsBuffer) incrementReceived() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.stats.EventsReceived++
}

// incrementDropped safely increments dropped counter
func (b *AnalyticsBuffer) incrementDropped() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.stats.EventsDropped++
}

// ClearBuffer clears all pending events (use with caution)
func (b *AnalyticsBuffer) ClearBuffer() error {
	if b.redis == nil {
		return fmt.Errorf("redis not available")
	}

	ctx := context.Background()
	return b.redis.Client.Del(ctx, b.bufferKey).Err()
}
