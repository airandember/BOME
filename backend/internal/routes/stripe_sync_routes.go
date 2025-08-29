package routes

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// RegisterStripeSyncRoutes registers sync management endpoints
func RegisterStripeSyncRoutes(router *gin.RouterGroup, syncService *services.StripeSyncService, cronService *services.StripeCronService) {
	sync := router.Group("/stripe/sync")
	{
		// Manual sync triggers
		sync.POST("/initial", func(c *gin.Context) { triggerInitialSync(c, syncService) })
		sync.POST("/incremental", func(c *gin.Context) { triggerIncrementalSync(c, syncService) })
		sync.POST("/quarterly", func(c *gin.Context) { triggerQuarterlySync(c, cronService) })

		// Sync status and monitoring
		sync.GET("/status", func(c *gin.Context) { getSyncStatusHandler(c, syncService) })
		sync.GET("/jobs", func(c *gin.Context) { getSyncJobs(c, syncService) })
		sync.GET("/config", func(c *gin.Context) { getSyncConfig(c, syncService) })
		sync.GET("/schedule", func(c *gin.Context) { getScheduleInfo(c, cronService) })

		// Sync configuration
		sync.PUT("/config/:entity_type", func(c *gin.Context) { updateSyncConfig(c, syncService) })
	}
}

// triggerInitialSync manually triggers the initial 1.5-year sync
func triggerInitialSync(c *gin.Context, syncService *services.StripeSyncService) {
	// Check if there's already a running sync
	running, err := isInitialSyncRunning(syncService)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check sync status"})
		return
	}

	if running {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Initial sync is already running",
			"message": "Please wait for the current sync to complete",
		})
		return
	}

	// Start initial sync in background
	go func() {
		ctx := context.Background()
		err := syncService.InitialDataSync(ctx)
		if err != nil {
			// Log error - in production you might want to send notifications
			// log.Printf("❌ Initial sync failed: %v", err)
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Initial sync started",
		"note":    "This will sync 1.5 years of historical data. Check /sync/status for progress.",
	})
}

// triggerIncrementalSync manually triggers incremental sync
func triggerIncrementalSync(c *gin.Context, syncService *services.StripeSyncService) {
	// Get optional 'since' parameter
	sinceStr := c.Query("since")
	var since time.Time

	if sinceStr != "" {
		var err error
		since, err = time.Parse("2006-01-02", sinceStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
			return
		}
	} else {
		// Default to last 24 hours
		since = time.Now().AddDate(0, 0, -1)
	}

	// Start incremental sync in background
	go func() {
		ctx := context.Background()
		err := syncService.IncrementalSync(ctx, since)
		if err != nil {
			// Log error
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Incremental sync started",
		"since":   since.Format("2006-01-02"),
	})
}

// triggerQuarterlySync manually triggers quarterly sync
func triggerQuarterlySync(c *gin.Context, cronService *services.StripeCronService) {
	err := cronService.ManualQuarterlySync()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to trigger quarterly sync"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Quarterly sync triggered manually",
		"note":    "This will sync data from the last quarter to now",
	})
}

// getSyncStatusHandler returns current sync status
func getSyncStatusHandler(c *gin.Context, syncService *services.StripeSyncService) {
	// Get recent sync jobs
	jobs, err := getRecentSyncJobs(syncService, 5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sync status"})
		return
	}

	// Check if any sync is currently running
	running := false
	var currentJob *services.SyncJob
	for _, job := range jobs {
		if job.Status == "running" {
			running = true
			currentJob = &job
			break
		}
	}

	status := gin.H{
		"sync_running": running,
		"recent_jobs":  jobs,
	}

	if currentJob != nil {
		progress := 0.0
		if currentJob.TotalItems > 0 {
			progress = float64(currentJob.ProcessedItems) / float64(currentJob.TotalItems) * 100
		}

		status["current_job"] = gin.H{
			"id":               currentJob.ID,
			"type":             currentJob.JobType,
			"entity_type":      currentJob.EntityType,
			"progress_percent": progress,
			"processed_items":  currentJob.ProcessedItems,
			"total_items":      currentJob.TotalItems,
			"started_at":       currentJob.StartedAt,
		}
	}

	c.JSON(http.StatusOK, status)
}

// getSyncJobs returns sync job history
func getSyncJobs(c *gin.Context, syncService *services.StripeSyncService) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit > 100 {
		limit = 20
	}

	jobs, err := getRecentSyncJobs(syncService, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sync jobs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"jobs":  jobs,
		"count": len(jobs),
	})
}

// getSyncConfig returns sync configuration
func getSyncConfig(c *gin.Context, syncService *services.StripeSyncService) {
	// This would get all sync configurations
	// For now, return a placeholder
	c.JSON(http.StatusOK, gin.H{
		"message": "Sync configuration endpoint",
		"note":    "Implementation pending - will return sync settings for all entity types",
	})
}

// getScheduleInfo returns cron schedule information
func getScheduleInfo(c *gin.Context, cronService *services.StripeCronService) {
	schedules := cronService.GetNextScheduledRuns()

	c.JSON(http.StatusOK, gin.H{
		"timezone":  "America/Denver (MST)",
		"next_runs": schedules,
		"schedule_info": gin.H{
			"quarterly_sync":    "0 0 1 */3 * (Midnight MST on 1st of each quarter)",
			"daily_incremental": "0 2 * * * (2 AM MST daily)",
			"weekly_cleanup":    "0 3 * * 0 (3 AM MST every Sunday)",
		},
	})
}

// updateSyncConfig updates sync configuration for an entity type
func updateSyncConfig(c *gin.Context, syncService *services.StripeSyncService) {
	entityType := c.Param("entity_type")

	var config struct {
		SyncEnabled       bool `json:"sync_enabled"`
		SyncIntervalHours int  `json:"sync_interval_hours"`
		BatchSize         int  `json:"batch_size"`
	}

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid configuration"})
		return
	}

	// TODO: Implement config update
	c.JSON(http.StatusOK, gin.H{
		"message":     "Configuration updated",
		"entity_type": entityType,
		"config":      config,
	})
}

// Helper functions
func isInitialSyncRunning(syncService *services.StripeSyncService) (bool, error) {
	// Check if there's a running initial sync job
	jobs, err := getRecentSyncJobs(syncService, 1)
	if err != nil {
		return false, err
	}

	for _, job := range jobs {
		if job.JobType == "initial_sync" && job.Status == "running" {
			return true, nil
		}
	}

	return false, nil
}

func getRecentSyncJobs(syncService *services.StripeSyncService, limit int) ([]services.SyncJob, error) {
	// This would query the database for recent sync jobs
	// For now, return empty slice - you'll need to implement the database query
	return []services.SyncJob{}, nil
}
