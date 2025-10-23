package services

import (
	"fmt"
	"log"
	"time"

	"bome-backend/infrastructure/database"
	"bome-backend/youtube/models"
)

// YouTubeScheduler handles scheduled YouTube RSS syncing
type YouTubeScheduler struct {
	db             *database.DB
	youtubeService *YouTubeService
	ticker         *time.Ticker
	done           chan bool
	isRunning      bool
	config         *models.YouTubeConfig
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

// Start begins the scheduled RSS sync
func (s *YouTubeScheduler) Start() {
	if s.isRunning {
		log.Printf("📺 [YOUTUBE-SCHEDULER] Already running")
		return
	}

	// Load config
	config, err := models.GetYouTubeConfig(s.db)
	if err != nil {
		log.Printf("⚠️  [YOUTUBE-SCHEDULER] Failed to load config: %v", err)
	}
	s.config = config

	s.isRunning = true
	log.Printf("🚀 [YOUTUBE-SCHEDULER] Starting scheduler")

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
	// Calculate time until next sync (default 2 PM MST)
	nextSync := s.getNextScheduledTime()
	log.Printf("📅 [YOUTUBE-SCHEDULER] Next sync scheduled for: %s", nextSync.Format("2006-01-02 15:04:05 MST"))

	// Wait until the first scheduled time
	initialDelay := time.Until(nextSync)
	if initialDelay > 0 {
		time.Sleep(initialDelay)
	}

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

// getNextScheduledTime calculates the next scheduled sync time
func (s *YouTubeScheduler) getNextScheduledTime() time.Time {
	// Default: 2 PM MST (UTC-7)
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
	count, err := models.GetYouTubeVideoCount(s.db)
	if err != nil {
		log.Printf("❌ [YOUTUBE-SCHEDULER] Failed to check existing videos: %v", err)
		return
	}

	if count == 0 {
		log.Printf("📺 [YOUTUBE-SCHEDULER] Database is empty, running initial sync...")
		s.performSync("initial")
	} else {
		log.Printf("📺 [YOUTUBE-SCHEDULER] Database has %d videos, skipping initial sync", count)
	}
}

// performSync executes the RSS sync
func (s *YouTubeScheduler) performSync(syncType string) {
	// Check if sync is enabled
	if s.config != nil && !s.config.SyncEnabled {
		log.Printf("⏸️  [YOUTUBE-SCHEDULER] Sync is disabled in config, skipping %s sync", syncType)
		return
	}

	log.Printf("🔄 [YOUTUBE-SCHEDULER] Starting %s RSS sync...", syncType)

	startTime := time.Now()

	// Create sync log
	syncLog := &models.YouTubeSyncLog{
		SyncType:  syncType,
		StartedAt: startTime,
		Status:    "running",
	}

	if err := models.CreateSyncLog(s.db, syncLog); err != nil {
		log.Printf("⚠️  [YOUTUBE-SCHEDULER] Failed to create sync log: %v", err)
	}

	// Perform sync
	result, err := s.youtubeService.SyncFromRSS()
	duration := time.Since(startTime)

	// Update sync log
	now := time.Now()
	syncLog.CompletedAt = &now

	if err != nil {
		log.Printf("❌ [YOUTUBE-SCHEDULER] %s sync failed after %v: %v", syncType, duration, err)
		syncLog.Status = "failed"
		syncLog.ErrorMessage = err.Error()
	} else {
		log.Printf("✅ [YOUTUBE-SCHEDULER] %s sync completed in %v", syncType, duration)
		log.Printf("📊 [YOUTUBE-SCHEDULER] Results: %d fetched, %d new, %d updated, %d skipped",
			result.TotalFetched, result.NewVideos, result.UpdatedVideos, result.SkippedVideos)

		syncLog.VideosFound = result.TotalFetched
		syncLog.VideosNew = result.NewVideos
		syncLog.VideosUpdated = result.UpdatedVideos
		syncLog.VideosSkipped = result.SkippedVideos

		if len(result.Errors) > 0 {
			log.Printf("⚠️  [YOUTUBE-SCHEDULER] %d errors encountered during sync", len(result.Errors))
			syncLog.Status = "partial"
			syncLog.ErrorMessage = fmt.Sprintf("%d errors encountered", len(result.Errors))
		} else {
			syncLog.Status = "success"
		}
	}

	if updateErr := models.UpdateSyncLog(s.db, syncLog); updateErr != nil {
		log.Printf("⚠️  [YOUTUBE-SCHEDULER] Failed to update sync log: %v", updateErr)
	}
}

// GetStatus returns the current scheduler status
func (s *YouTubeScheduler) GetStatus() map[string]interface{} {
	nextSync := s.getNextScheduledTime()

	status := map[string]interface{}{
		"running":      s.isRunning,
		"next_sync":    nextSync.Format("2006-01-02 15:04:05 MST"),
		"next_sync_in": time.Until(nextSync).String(),
		"sync_time":    "2:00 PM MST daily",
		"timezone":     "MST (UTC-7)",
	}

	if s.config != nil {
		status["sync_enabled"] = s.config.SyncEnabled
		status["channel_id"] = s.config.ChannelID
		status["last_sync_at"] = s.config.LastSyncAt
	}

	// Get latest sync log
	latestLog, err := models.GetLatestSyncLog(s.db)
	if err == nil && latestLog != nil {
		status["latest_sync"] = map[string]interface{}{
			"type":           latestLog.SyncType,
			"status":         latestLog.Status,
			"started_at":     latestLog.StartedAt,
			"completed_at":   latestLog.CompletedAt,
			"videos_found":   latestLog.VideosFound,
			"videos_new":     latestLog.VideosNew,
			"videos_updated": latestLog.VideosUpdated,
		}
	}

	return status
}

// TriggerManualSync allows manual triggering of sync (for admin endpoints)
func (s *YouTubeScheduler) TriggerManualSync() (*YouTubeSyncResult, error) {
	log.Printf("🔧 [YOUTUBE-SCHEDULER] Manual sync triggered")

	if s.youtubeService == nil {
		log.Printf("❌ [YOUTUBE-SCHEDULER] YouTube service is nil!")
		return nil, fmt.Errorf("YouTube service not initialized")
	}

	log.Printf("✅ [YOUTUBE-SCHEDULER] YouTube service is initialized, calling SyncFromRSS")

	// Create sync log
	syncLog := &models.YouTubeSyncLog{
		SyncType:  "manual",
		StartedAt: time.Now(),
		Status:    "running",
	}

	if err := models.CreateSyncLog(s.db, syncLog); err != nil {
		log.Printf("⚠️  [YOUTUBE-SCHEDULER] Failed to create sync log: %v", err)
	}

	// Perform sync
	result, err := s.youtubeService.SyncFromRSS()

	// Update sync log
	now := time.Now()
	syncLog.CompletedAt = &now

	if err != nil {
		log.Printf("❌ [YOUTUBE-SCHEDULER] Manual sync failed: %v", err)
		syncLog.Status = "failed"
		syncLog.ErrorMessage = err.Error()
		models.UpdateSyncLog(s.db, syncLog)
		return nil, err
	}

	log.Printf("✅ [YOUTUBE-SCHEDULER] Manual sync completed successfully")
	syncLog.VideosFound = result.TotalFetched
	syncLog.VideosNew = result.NewVideos
	syncLog.VideosUpdated = result.UpdatedVideos
	syncLog.VideosSkipped = result.SkippedVideos
	syncLog.Status = "success"

	if updateErr := models.UpdateSyncLog(s.db, syncLog); updateErr != nil {
		log.Printf("⚠️  [YOUTUBE-SCHEDULER] Failed to update sync log: %v", updateErr)
	}

	return result, nil
}

// UpdateConfiguration updates the scheduler's configuration and restarts it
func (s *YouTubeScheduler) UpdateConfiguration(config *models.YouTubeConfig) error {
	log.Printf("🔧 [YOUTUBE-SCHEDULER] Updating configuration: Channel=%s, Schedule=%s, Enabled=%v",
		config.ChannelID, config.SyncSchedule, config.SyncEnabled)

	// Update config in database
	if err := models.UpdateYouTubeConfig(s.db, config); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	// Store config
	s.config = config

	// Recreate the YouTube service with the new configuration
	s.youtubeService = NewYouTubeService(s.db)

	log.Printf("✅ [YOUTUBE-SCHEDULER] Configuration updated successfully")
	return nil
}

// IsRunning returns whether the scheduler is currently running
func (s *YouTubeScheduler) IsRunning() bool {
	return s.isRunning
}
