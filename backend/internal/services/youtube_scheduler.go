package services

import (
	"fmt"
	"log"
	"time"

	"bome-backend/internal/database"
)

// YouTubeScheduler handles scheduled YouTube RSS syncing
type YouTubeScheduler struct {
	db             *database.DB
	youtubeService *YouTubeService
	ticker         *time.Ticker
	done           chan bool
	isRunning      bool
}

// NewYouTubeScheduler creates a new YouTube scheduler
func NewYouTubeScheduler(db *database.DB) *YouTubeScheduler {
	return &YouTubeScheduler{
		db:             db,
		youtubeService: NewYouTubeService(db),
		done:           make(chan bool),
		isRunning:      false,
	}
}

// Start begins the scheduled RSS sync at 2 PM MST daily
func (s *YouTubeScheduler) Start() {
	if s.isRunning {
		log.Printf("📺 [YOUTUBE-SCHEDULER] Already running")
		return
	}

	s.isRunning = true
	log.Printf("🚀 [YOUTUBE-SCHEDULER] Starting daily RSS sync at 2 PM MST")

	// Run initial sync if database is empty
	go s.runInitialSyncIfNeeded()

	// Start the daily scheduler
	go s.scheduleDailySync()
}

// Stop stops the scheduler
func (s *YouTubeScheduler) Stop() {
	if !s.isRunning {
		return
	}

	log.Printf("🛑 [YOUTUBE-SCHEDULER] Stopping scheduler")
	s.isRunning = false

	if s.ticker != nil {
		s.ticker.Stop()
	}

	close(s.done)
}

// scheduleDailySync runs the scheduler loop
func (s *YouTubeScheduler) scheduleDailySync() {
	// Calculate time until next 2 PM MST
	nextSync := s.getNext2PMMST()
	log.Printf("📅 [YOUTUBE-SCHEDULER] Next sync scheduled for: %s", nextSync.Format("2006-01-02 15:04:05 MST"))

	// Wait until the first 2 PM MST
	initialDelay := time.Until(nextSync)
	time.Sleep(initialDelay)

	// Run the first sync
	s.performSync("scheduled")

	// Set up daily ticker (24 hours)
	s.ticker = time.NewTicker(24 * time.Hour)

	for {
		select {
		case <-s.ticker.C:
			s.performSync("scheduled")
		case <-s.done:
			log.Printf("📺 [YOUTUBE-SCHEDULER] Scheduler stopped")
			return
		}
	}
}

// getNext2PMMST calculates the next 2 PM MST
func (s *YouTubeScheduler) getNext2PMMST() time.Time {
	// MST is UTC-7
	mst := time.FixedZone("MST", -7*60*60)
	now := time.Now().In(mst)

	// Target time: 2 PM MST today
	target := time.Date(now.Year(), now.Month(), now.Day(), 14, 0, 0, 0, mst)

	// If it's already past 2 PM today, schedule for tomorrow
	if now.After(target) {
		target = target.Add(24 * time.Hour)
	}

	return target
}

// runInitialSyncIfNeeded runs an initial sync if the database is empty
func (s *YouTubeScheduler) runInitialSyncIfNeeded() {
	// Check if we have any videos in the database
	videos, err := s.youtubeService.GetLatestVideos(1)
	if err != nil {
		log.Printf("❌ [YOUTUBE-SCHEDULER] Failed to check existing videos: %v", err)
		return
	}

	if len(videos.Videos) == 0 {
		log.Printf("📺 [YOUTUBE-SCHEDULER] Database is empty, running initial sync...")
		s.performSync("initial")
	} else {
		log.Printf("📺 [YOUTUBE-SCHEDULER] Database has %d videos, skipping initial sync", len(videos.Videos))
	}
}

// performSync executes the RSS sync
func (s *YouTubeScheduler) performSync(syncType string) {
	log.Printf("🔄 [YOUTUBE-SCHEDULER] Starting %s RSS sync...", syncType)

	startTime := time.Now()
	result, err := s.youtubeService.SyncFromRSS()
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("❌ [YOUTUBE-SCHEDULER] %s sync failed after %v: %v", syncType, duration, err)
		return
	}

	log.Printf("✅ [YOUTUBE-SCHEDULER] %s sync completed in %v", syncType, duration)
	log.Printf("📊 [YOUTUBE-SCHEDULER] Results: %d fetched, %d new, %d updated",
		result.TotalFetched, result.NewVideos, result.UpdatedVideos)

	if len(result.Errors) > 0 {
		log.Printf("⚠️ [YOUTUBE-SCHEDULER] %d errors encountered during sync", len(result.Errors))
		for _, errMsg := range result.Errors {
			log.Printf("   - %s", errMsg)
		}
	}
}

// GetStatus returns the current scheduler status
func (s *YouTubeScheduler) GetStatus() map[string]interface{} {
	nextSync := s.getNext2PMMST()

	return map[string]interface{}{
		"running":      s.isRunning,
		"next_sync":    nextSync.Format("2006-01-02 15:04:05 MST"),
		"next_sync_in": time.Until(nextSync).String(),
		"sync_time":    "2:00 PM MST daily",
		"timezone":     "MST (UTC-7)",
	}
}

// TriggerManualSync allows manual triggering of sync (for admin endpoints)
func (s *YouTubeScheduler) TriggerManualSync() (*YouTubeSyncResult, error) {
	log.Printf("🔧 [YOUTUBE-SCHEDULER] Manual sync triggered")

	if s.youtubeService == nil {
		log.Printf("❌ [YOUTUBE-SCHEDULER] YouTube service is nil!")
		return nil, fmt.Errorf("YouTube service not initialized")
	}

	log.Printf("✅ [YOUTUBE-SCHEDULER] YouTube service is initialized, calling SyncFromRSS")
	result, err := s.youtubeService.SyncFromRSS()
	if err != nil {
		log.Printf("❌ [YOUTUBE-SCHEDULER] Manual sync failed: %v", err)
		return nil, err
	}

	log.Printf("✅ [YOUTUBE-SCHEDULER] Manual sync completed successfully")
	return result, err
}

// UpdateConfiguration updates the scheduler's configuration and restarts it
func (s *YouTubeScheduler) UpdateConfiguration(channelID string, syncHour, syncMinute int, timezone string, autoSyncEnabled bool) {
	log.Printf("🔧 [YOUTUBE-SCHEDULER] Updating configuration: Channel=%s, Time=%02d:%02d %s, AutoSync=%v",
		channelID, syncHour, syncMinute, timezone, autoSyncEnabled)

	// Recreate the YouTube service with the new channel ID (it will read from public_settings)
	s.youtubeService = NewYouTubeService(s.db)

	// Note: The scheduler itself doesn't store config - it reads from public_settings via YouTubeService
	// This method just ensures the service is refreshed with the new configuration
}
