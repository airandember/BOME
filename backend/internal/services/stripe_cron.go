package services

import (
	"context"
	"log"
	"time"
)

// StripeCronService handles scheduled Stripe sync operations
type StripeCronService struct {
	syncService *StripeSyncService
	timezone    *time.Location
	isRunning   bool
	logger      *StripeLogger
}

// NewStripeCronService creates a new cron service for Stripe sync
func NewStripeCronService(syncService *StripeSyncService) *StripeCronService {
	// Load MST timezone (configurable as per your requirements)
	mst, err := time.LoadLocation("America/Denver") // MST timezone
	if err != nil {
		log.Printf("⚠️ Failed to load MST timezone, using UTC: %v", err)
		mst = time.UTC
	}

	return &StripeCronService{
		syncService: syncService,
		timezone:    mst,
		isRunning:   false,
		logger:      NewStripeLogger("CRON"),
	}
}

// StartCronJobs starts all scheduled Stripe sync jobs
func (s *StripeCronService) StartCronJobs() error {
	log.Printf("🕐 Starting Stripe cron jobs in %s timezone", s.timezone.String())

	// For now, we'll implement a simple ticker-based system
	// In production, you would use a proper cron library like github.com/robfig/cron/v3
	s.isRunning = true

	// Start background goroutines for scheduled tasks
	go s.scheduledTaskRunner()

	log.Println("✅ Stripe cron jobs started successfully")
	return nil
}

// StopCronJobs stops all scheduled jobs
func (s *StripeCronService) StopCronJobs() {
	log.Println("🛑 Stopping Stripe cron jobs...")
	s.isRunning = false
	log.Println("✅ Stripe cron jobs stopped")
}

// scheduledTaskRunner runs the scheduled tasks
func (s *StripeCronService) scheduledTaskRunner() {
	ticker := time.NewTicker(1 * time.Hour) // Check every hour
	defer ticker.Stop()

	for s.isRunning {
		<-ticker.C
		now := time.Now().In(s.timezone)

		// Check if it's time for quarterly sync (1st of quarter at midnight)
		if s.isQuarterlyTime(now) {
			log.Println("📅 Quarterly sync triggered at midnight MST")
			go s.runQuarterlySync()
		}

		// Check if it's time for daily sync (2 AM daily)
		if now.Hour() == 2 && now.Minute() == 0 {
			log.Println("📅 Daily incremental sync triggered at 2 AM MST")
			go s.runIncrementalSync()
		}

		// Check if it's time for weekly cleanup (3 AM Sunday)
		if now.Weekday() == time.Sunday && now.Hour() == 3 && now.Minute() == 0 {
			log.Println("📅 Weekly cleanup triggered at 3 AM MST")
			go s.runWeeklyCleanup()
		}
	}
}

// isQuarterlyTime checks if it's the first day of a quarter at midnight
func (s *StripeCronService) isQuarterlyTime(now time.Time) bool {
	// First day of quarter and midnight
	return now.Day() == 1 &&
		(now.Month() == 1 || now.Month() == 4 || now.Month() == 7 || now.Month() == 10) &&
		now.Hour() == 0 && now.Minute() == 0
}

// runQuarterlySync performs the quarterly full sync
func (s *StripeCronService) runQuarterlySync() {
	log.Println("🚀 Starting quarterly Stripe sync...")

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Hour) // 4-hour timeout
	defer cancel()

	// Calculate last quarter date
	now := time.Now().In(s.timezone)
	lastQuarter := now.AddDate(0, -3, 0) // 3 months ago

	log.Printf("📊 Quarterly sync: Fetching data from %s to %s",
		lastQuarter.Format("2006-01-02"),
		now.Format("2006-01-02"))

	// Run incremental sync from last quarter to now
	err := s.syncService.IncrementalSync(ctx, lastQuarter)
	if err != nil {
		log.Printf("❌ Quarterly sync failed: %v", err)
		// TODO: Send alert/notification about failed sync
		return
	}

	// Update sync configuration
	s.updateLastSyncTime("quarterly", now)

	log.Println("✅ Quarterly sync completed successfully")
}

// runIncrementalSync performs daily incremental sync
func (s *StripeCronService) runIncrementalSync() {
	log.Println("🔄 Starting daily incremental sync...")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour) // 1-hour timeout
	defer cancel()

	// Sync data from last 24 hours
	yesterday := time.Now().In(s.timezone).AddDate(0, 0, -1)

	err := s.syncService.IncrementalSync(ctx, yesterday)
	if err != nil {
		log.Printf("❌ Daily incremental sync failed: %v", err)
		return
	}

	s.updateLastSyncTime("incremental", time.Now().In(s.timezone))
	log.Println("✅ Daily incremental sync completed")
}

// runWeeklyCleanup performs maintenance tasks
func (s *StripeCronService) runWeeklyCleanup() {
	log.Println("🧹 Starting weekly cleanup...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Clean up old sync jobs (keep last 30 days)
	err := s.syncService.CleanupOldSyncJobs(ctx, 30)
	if err != nil {
		log.Printf("⚠️ Failed to cleanup old sync jobs: %v", err)
	}

	// Update aggregation tables
	err = s.syncService.UpdateAggregationTables(ctx)
	if err != nil {
		log.Printf("⚠️ Failed to update aggregation tables: %v", err)
	}

	log.Println("✅ Weekly cleanup completed")
}

// updateLastSyncTime updates the last sync time in configuration
func (s *StripeCronService) updateLastSyncTime(syncType string, syncTime time.Time) {
	// Update all entity types with the last sync time
	entityTypes := []string{"customer", "product", "price", "subscription", "invoice", "coupon"}

	for _, entityType := range entityTypes {
		var query string
		if syncType == "quarterly" {
			query = `
				UPDATE stripe_sync_config 
				SET last_full_sync = $1, updated_at = NOW() 
				WHERE entity_type = $2
			`
		} else {
			query = `
				UPDATE stripe_sync_config 
				SET last_incremental_sync = $1, updated_at = NOW() 
				WHERE entity_type = $2
			`
		}

		_, err := s.syncService.db.Exec(query, syncTime, entityType)
		if err != nil {
			log.Printf("⚠️ Failed to update %s sync time for %s: %v", syncType, entityType, err)
		}
	}
}

// ManualQuarterlySync allows manual triggering of quarterly sync
func (s *StripeCronService) ManualQuarterlySync() error {
	log.Println("🔧 Manual quarterly sync triggered")

	go func() {
		s.runQuarterlySync()
	}()

	return nil
}

// GetNextScheduledRuns returns information about next scheduled runs
func (s *StripeCronService) GetNextScheduledRuns() map[string]time.Time {
	now := time.Now().In(s.timezone)
	schedules := make(map[string]time.Time)

	// Calculate next quarterly sync (next 1st of quarter)
	nextQuarter := s.getNextQuarterStart(now)
	schedules["quarterly_sync"] = nextQuarter

	// Calculate next daily sync (next 2 AM)
	nextDaily := now.Truncate(24 * time.Hour).Add(2 * time.Hour)
	if nextDaily.Before(now) {
		nextDaily = nextDaily.Add(24 * time.Hour)
	}
	schedules["daily_incremental"] = nextDaily

	// Calculate next weekly cleanup (next Sunday 3 AM)
	daysUntilSunday := (7 - int(now.Weekday())) % 7
	if daysUntilSunday == 0 && (now.Hour() > 3 || (now.Hour() == 3 && now.Minute() > 0)) {
		daysUntilSunday = 7
	}
	nextSunday := now.AddDate(0, 0, daysUntilSunday).Truncate(24 * time.Hour).Add(3 * time.Hour)
	schedules["weekly_cleanup"] = nextSunday

	return schedules
}

// getNextQuarterStart calculates the next quarter start date
func (s *StripeCronService) getNextQuarterStart(now time.Time) time.Time {
	year := now.Year()
	month := now.Month()

	var nextQuarterMonth time.Month
	switch {
	case month < 4:
		nextQuarterMonth = 4
	case month < 7:
		nextQuarterMonth = 7
	case month < 10:
		nextQuarterMonth = 10
	default:
		nextQuarterMonth = 1
		year++
	}

	return time.Date(year, nextQuarterMonth, 1, 0, 0, 0, 0, s.timezone)
}
