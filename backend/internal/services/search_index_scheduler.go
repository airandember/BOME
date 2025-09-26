package services

import (
	"fmt"
	"log"
	"time"

	"bome-backend/internal/database"

	"github.com/robfig/cron/v3"
)

// SearchIndexScheduler manages automated search index generation
type SearchIndexScheduler struct {
	cron         *cron.Cron
	db           *database.DB
	bunnyService *BunnyService
	isRunning    bool
	lastRun      time.Time
	nextRun      time.Time
}

// NewSearchIndexScheduler creates a new search index scheduler
func NewSearchIndexScheduler(db *database.DB, bunnyService *BunnyService) *SearchIndexScheduler {
	return &SearchIndexScheduler{
		cron:         cron.New(cron.WithSeconds()),
		db:           db,
		bunnyService: bunnyService,
		isRunning:    false,
	}
}

// Start begins the scheduled search index generation
func (s *SearchIndexScheduler) Start() error {
	if s.isRunning {
		return fmt.Errorf("scheduler is already running")
	}

	log.Println("🌙 Starting search index scheduler...")

	// Schedule for midnight every day (0 0 0 * * *)
	_, err := s.cron.AddFunc("0 0 0 * * *", func() {
		log.Println("🕛 Midnight search index generation triggered")
		s.generateSearchIndex()
	})
	if err != nil {
		return fmt.Errorf("failed to schedule search index generation: %w", err)
	}

	// Also schedule for 6 AM as backup (0 0 6 * * *)
	_, err = s.cron.AddFunc("0 0 6 * * *", func() {
		log.Println("🌅 Morning search index generation triggered")
		s.generateSearchIndex()
	})
	if err != nil {
		return fmt.Errorf("failed to schedule backup search index generation: %w", err)
	}

	s.cron.Start()
	s.isRunning = true

	// Set next run time
	entries := s.cron.Entries()
	if len(entries) > 0 {
		s.nextRun = entries[0].Next
	}

	log.Printf("✅ Search index scheduler started - next run: %s", s.nextRun.Format("2006-01-02 15:04:05"))
	return nil
}

// Stop stops the scheduler
func (s *SearchIndexScheduler) Stop() {
	if !s.isRunning {
		return
	}

	log.Println("🛑 Stopping search index scheduler...")
	s.cron.Stop()
	s.isRunning = false
	log.Println("✅ Search index scheduler stopped")
}

// UpdateConfiguration updates the scheduler configuration and restarts it
func (s *SearchIndexScheduler) UpdateConfiguration(schedule, backupSchedule string, enableBackup bool) error {
	// Stop current scheduler
	s.Stop()

	// Create new cron instance
	s.cron = cron.New(cron.WithSeconds())

	log.Printf("🔄 Updating search index scheduler configuration...")
	log.Printf("   Main schedule: %s", schedule)
	log.Printf("   Backup schedule: %s", backupSchedule)
	log.Printf("   Backup enabled: %t", enableBackup)

	// Add main schedule
	_, err := s.cron.AddFunc(schedule, func() {
		log.Println("🕛 Scheduled search index generation triggered")
		s.generateSearchIndex()
	})
	if err != nil {
		return fmt.Errorf("failed to schedule main search index generation: %w", err)
	}

	// Add backup schedule if enabled
	if enableBackup {
		_, err = s.cron.AddFunc(backupSchedule, func() {
			log.Println("🌅 Backup search index generation triggered")
			s.generateSearchIndex()
		})
		if err != nil {
			return fmt.Errorf("failed to schedule backup search index generation: %w", err)
		}
	}

	// Start the scheduler
	return s.Start()
}

// TriggerManualGeneration manually triggers search index generation
func (s *SearchIndexScheduler) TriggerManualGeneration() error {
	log.Println("🔄 Manual search index generation triggered")
	return s.generateSearchIndex()
}

// GetStatus returns the current scheduler status
func (s *SearchIndexScheduler) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"running":  s.isRunning,
		"lastRun":  nil,
		"nextRun":  nil,
		"schedule": "Daily at midnight (00:00) and 6 AM backup",
	}

	if !s.lastRun.IsZero() {
		status["lastRun"] = s.lastRun.Format("2006-01-02 15:04:05")
	}

	if !s.nextRun.IsZero() && s.isRunning {
		status["nextRun"] = s.nextRun.Format("2006-01-02 15:04:05")

		// Update next run from cron entries
		entries := s.cron.Entries()
		if len(entries) > 0 {
			status["nextRun"] = entries[0].Next.Format("2006-01-02 15:04:05")
		}
	}

	return status
}

// generateSearchIndex performs the actual search index generation
func (s *SearchIndexScheduler) generateSearchIndex() error {
	startTime := time.Now()
	log.Println("🚀 Starting search index generation...")

	defer func() {
		s.lastRun = time.Now()
		duration := s.lastRun.Sub(startTime)
		log.Printf("⏱️ Search index generation completed in %v", duration)
	}()

	// Get all videos from database
	videos, err := s.db.GetAllVideos()
	if err != nil {
		log.Printf("❌ Failed to get videos: %v", err)
		return fmt.Errorf("failed to get videos: %w", err)
	}

	log.Printf("📥 Found %d videos in database", len(videos))

	// Generate search index structure
	searchIndex := map[string]interface{}{
		"version":     "1.0",
		"generatedAt": time.Now().Format(time.RFC3339),
		"totalVideos": len(videos),
		"videos":      s.convertVideosToSearchFormat(videos),
	}

	// Here you would write to the frontend static directory
	// This could be done via:
	// 1. Direct file write (if frontend is in same container)
	// 2. API call to frontend service
	// 3. Shared volume
	// 4. Cloud storage (S3, etc.)

	log.Printf("✅ Search index generated with %d videos (version: %s)", len(videos), searchIndex["version"])
	return nil
}

// convertVideosToSearchFormat converts database videos to search index format
func (s *SearchIndexScheduler) convertVideosToSearchFormat(videos []database.Video) []map[string]interface{} {
	var searchVideos []map[string]interface{}

	for _, video := range videos {
		// Generate thumbnail URLs
		thumbnailURL := ""
		if video.BunnyVideoID != "" {
			if video.ThumbnailFileName != "" {
				thumbnailURL = s.bunnyService.GetThumbnailURLWithFilename(video.BunnyVideoID, video.ThumbnailFileName)
			} else {
				thumbnailURL = s.bunnyService.GetThumbnailURL(video.BunnyVideoID)
			}
		}

		// Fallback thumbnail
		fallbackThumbnail := ""
		if video.BunnyVideoID != "" {
			fallbackThumbnail = fmt.Sprintf("https://vz-f75053f7-465.b-cdn.net/%s/thumbnail.jpg", video.BunnyVideoID)
		}

		searchVideo := map[string]interface{}{
			"id":           video.BunnyVideoID,
			"title":        video.Title,
			"description":  video.Description,
			"category":     video.Category,
			"tags":         video.Tags,
			"duration":     video.Duration,
			"createdAt":    video.CreatedAt.Format(time.RFC3339),
			"thumbnail":    thumbnailURL,
			"thumbnailUrl": thumbnailURL,
			"bunny": map[string]interface{}{
				"guid":              video.BunnyVideoID,
				"videoLibraryId":    "",
				"thumbnailFileName": video.ThumbnailFileName,
				"previewImageUrl":   fallbackThumbnail,
				"width":             0,
				"height":            0,
				"length":            video.Duration,
			},
			"views":     video.ViewCount,
			"status":    video.Status,
			"videoUrl":  s.bunnyService.GetStreamURL(video.BunnyVideoID),
			"iframeSrc": s.bunnyService.GetIframeURL(video.BunnyVideoID),
		}

		// Use fallback if no thumbnail
		if thumbnailURL == "" {
			searchVideo["thumbnail"] = fallbackThumbnail
			searchVideo["thumbnailUrl"] = fallbackThumbnail
		}

		searchVideos = append(searchVideos, searchVideo)
	}

	return searchVideos
}
