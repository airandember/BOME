package routes

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"bome-backend/internal/database"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SearchIndexRoutes sets up search index management routes
func SearchIndexRoutes(router *gin.RouterGroup, db *database.DB, bunnyService *services.BunnyService) {
	fmt.Printf("🔧 [SEARCH-INDEX] Setting up search index routes...\n")

	// Initialize search index scheduler
	fmt.Printf("🔧 [SEARCH-INDEX] Initializing scheduler...\n")
	searchIndexScheduler := services.NewSearchIndexScheduler(db, bunnyService)

	// Start the scheduler
	fmt.Printf("🔧 [SEARCH-INDEX] Starting scheduler...\n")
	err := searchIndexScheduler.Start()
	if err != nil {
		fmt.Printf("⚠️ [SEARCH-INDEX] Failed to start search index scheduler: %v\n", err)
	} else {
		fmt.Printf("✅ [SEARCH-INDEX] Scheduler started successfully\n")
	}

	searchIndex := router.Group("/search-index")
	{
		// Add middleware to log all requests to search-index routes
		searchIndex.Use(func(c *gin.Context) {
			fmt.Printf("🌐 [SEARCH-INDEX] Incoming request: %s %s\n", c.Request.Method, c.Request.URL.Path)
			c.Next()
			fmt.Printf("🌐 [SEARCH-INDEX] Request completed: %s %s - Status: %d\n", c.Request.Method, c.Request.URL.Path, c.Writer.Status())
		})

		// Get scheduler status
		searchIndex.GET("/scheduler/status", getSearchIndexSchedulerStatus(searchIndexScheduler))
		fmt.Printf("🔧 [SEARCH-INDEX] Registered: GET /api/v1/search-index/scheduler/status\n")

		// Trigger manual generation
		searchIndex.POST("/scheduler/trigger", triggerSearchIndexGeneration(searchIndexScheduler))
		fmt.Printf("🔧 [SEARCH-INDEX] Registered: POST /api/v1/search-index/scheduler/trigger\n")

		// Get configuration
		searchIndex.GET("/config", getSearchIndexConfig(db))
		fmt.Printf("🔧 [SEARCH-INDEX] Registered: GET /api/v1/search-index/config\n")

		// Update configuration
		searchIndex.POST("/config", updateSearchIndexConfig(db, searchIndexScheduler))
		fmt.Printf("🔧 [SEARCH-INDEX] Registered: POST /api/v1/search-index/config\n")

		// Get generation history/stats
		searchIndex.GET("/stats", getSearchIndexStats(db))
		fmt.Printf("🔧 [SEARCH-INDEX] Registered: GET /api/v1/search-index/stats\n")

		// Download current search index
		searchIndex.GET("/download", downloadSearchIndex())
		fmt.Printf("🔧 [SEARCH-INDEX] Registered: GET /api/v1/search-index/download\n")

		// Test endpoint to verify routing works
		searchIndex.GET("/test", func(c *gin.Context) {
			fmt.Printf("🧪 [SEARCH-INDEX] Test endpoint called successfully!\n")
			c.JSON(http.StatusOK, gin.H{
				"success":   true,
				"message":   "Search index routes are working!",
				"timestamp": time.Now().Format("2006-01-02 15:04:05"),
			})
		})
		fmt.Printf("🔧 [SEARCH-INDEX] Registered: GET /api/v1/search-index/test\n")
	}

	fmt.Printf("✅ [SEARCH-INDEX] All routes registered successfully\n")
}

// getSearchIndexSchedulerStatus returns the current scheduler status
func getSearchIndexSchedulerStatus(scheduler *services.SearchIndexScheduler) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("🔍 [SEARCH-INDEX] GET /scheduler/status called\n")

		status := scheduler.GetStatus()
		fmt.Printf("🔍 [SEARCH-INDEX] Scheduler status: %+v\n", status)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"status":  status,
		})

		fmt.Printf("✅ [SEARCH-INDEX] Status response sent successfully\n")
	}
}

// triggerSearchIndexGeneration manually triggers search index generation
func triggerSearchIndexGeneration(scheduler *services.SearchIndexScheduler) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := scheduler.TriggerManualGeneration()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   fmt.Sprintf("Failed to trigger generation: %v", err),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Search index generation triggered successfully",
		})
	}
}

// getSearchIndexConfig returns the current search index configuration
func getSearchIndexConfig(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("🔍 [SEARCH-INDEX] GET /config called\n")

		// Get configuration from public_settings
		schedule, _ := db.GetPublicSetting("search_index_schedule")
		if schedule == "" {
			schedule = "0 0 * * *" // Default: midnight daily
		}

		autoSync, _ := db.GetPublicSetting("search_index_auto_sync")
		autoSyncBool, _ := strconv.ParseBool(autoSync)

		backupSchedule, _ := db.GetPublicSetting("search_index_backup_schedule")
		if backupSchedule == "" {
			backupSchedule = "0 6 * * *" // Default: 6 AM backup
		}

		enableBackup, _ := db.GetPublicSetting("search_index_enable_backup")
		enableBackupBool, _ := strconv.ParseBool(enableBackup)
		if enableBackup == "" {
			enableBackupBool = true // Default: enabled
		}

		fmt.Printf("🔍 [SEARCH-INDEX] Config loaded: schedule=%s, autoSync=%t, backupSchedule=%s, enableBackup=%t\n",
			schedule, autoSyncBool, backupSchedule, enableBackupBool)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"config": gin.H{
				"schedule":       schedule,
				"autoSync":       autoSyncBool,
				"backupSchedule": backupSchedule,
				"enableBackup":   enableBackupBool,
				"timezone":       "America/Denver", // MST
			},
		})

		fmt.Printf("✅ [SEARCH-INDEX] Config response sent successfully\n")
	}
}

// updateSearchIndexConfig updates the search index configuration
func updateSearchIndexConfig(db *database.DB, scheduler *services.SearchIndexScheduler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var config struct {
			Schedule       string `json:"schedule"`
			AutoSync       bool   `json:"autoSync"`
			BackupSchedule string `json:"backupSchedule"`
			EnableBackup   bool   `json:"enableBackup"`
		}

		if err := c.ShouldBindJSON(&config); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid configuration data",
			})
			return
		}

		// Validate cron schedule format (basic validation)
		if config.Schedule == "" {
			config.Schedule = "0 0 * * *"
		}
		if config.BackupSchedule == "" {
			config.BackupSchedule = "0 6 * * *"
		}

		// Save configuration to database
		err := db.SetPublicSetting("search_index_schedule", config.Schedule)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to save schedule configuration",
			})
			return
		}

		err = db.SetPublicSetting("search_index_auto_sync", strconv.FormatBool(config.AutoSync))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to save auto sync configuration",
			})
			return
		}

		err = db.SetPublicSetting("search_index_backup_schedule", config.BackupSchedule)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to save backup schedule configuration",
			})
			return
		}

		err = db.SetPublicSetting("search_index_enable_backup", strconv.FormatBool(config.EnableBackup))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to save backup enable configuration",
			})
			return
		}

		// Restart scheduler with new configuration
		scheduler.Stop()
		err = scheduler.UpdateConfiguration(config.Schedule, config.BackupSchedule, config.EnableBackup)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   fmt.Sprintf("Failed to update scheduler: %v", err),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Search index configuration updated successfully",
		})
	}
}

// getSearchIndexStats returns statistics about search index generation
func getSearchIndexStats(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("🔍 [SEARCH-INDEX] GET /stats called\n")

		// Get video count from database
		videos, err := db.GetAllVideos()
		videoCount := 0
		if err == nil {
			videoCount = len(videos)
			fmt.Printf("🔍 [SEARCH-INDEX] Found %d videos in database\n", videoCount)
		} else {
			fmt.Printf("⚠️ [SEARCH-INDEX] Error getting videos: %v\n", err)
		}

		// Check if search index file exists and get its info
		var fileInfo map[string]interface{}
		var lastGenerated interface{}
		indexPath := getSearchIndexPath()

		if stat, err := os.Stat(indexPath); err == nil {
			modTime := stat.ModTime()
			fileInfo = map[string]interface{}{
				"exists":       true,
				"size":         stat.Size(),
				"sizeKB":       stat.Size() / 1024,
				"sizeMB":       fmt.Sprintf("%.2f", float64(stat.Size())/(1024*1024)),
				"lastModified": modTime.Format("2006-01-02 15:04:05"),
				"age":          time.Since(modTime).String(),
			}
			// Use file modification time as lastGenerated
			lastGenerated = modTime.Format("2006-01-02 15:04:05")
		} else {
			fileInfo = map[string]interface{}{
				"exists": false,
			}
			lastGenerated = nil
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"stats": gin.H{
				"totalVideos":   videoCount,
				"searchIndex":   fileInfo,
				"indexPath":     indexPath,
				"lastGenerated": lastGenerated,
			},
		})
	}
}

// downloadSearchIndex allows downloading the current search index file
func downloadSearchIndex() gin.HandlerFunc {
	return func(c *gin.Context) {
		indexPath := getSearchIndexPath()

		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Search index file not found",
			})
			return
		}

		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", "attachment; filename=search-index.json")
		c.Header("Content-Type", "application/json")

		c.File(indexPath)
	}
}

// SetupPublicSearchIndexRoutes sets up PUBLIC search index routes (NO AUTH REQUIRED)
// This allows the frontend to fetch the search index directly from the backend API
func SetupPublicSearchIndexRoutes(router *gin.RouterGroup) {
	fmt.Printf("🔧 [SEARCH-INDEX] Setting up PUBLIC search index routes...\n")

	// Public endpoint to serve search index - NO AUTH
	router.GET("/search-index.json", servePublicSearchIndex())
	fmt.Printf("✅ [SEARCH-INDEX] Registered PUBLIC: GET /api/v1/search-index.json\n")
}

// servePublicSearchIndex serves the search index file for frontend consumption
func servePublicSearchIndex() gin.HandlerFunc {
	return func(c *gin.Context) {
		indexPath := getSearchIndexPath()

		// Check if file exists
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			fmt.Printf("⚠️ [SEARCH-INDEX] Public request failed - file not found at: %s\n", indexPath)
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Search index not yet generated",
			})
			return
		}

		// Serve the file with proper caching headers
		c.Header("Content-Type", "application/json")
		c.Header("Cache-Control", "public, max-age=300") // Cache for 5 minutes
		c.Header("Access-Control-Allow-Origin", "*")     // Allow CORS for frontend

		c.File(indexPath)
	}
}

// getSearchIndexPath returns the path where the search index should be stored
func getSearchIndexPath() string {
	// Check for custom path from environment variable (PREFERRED METHOD)
	if customPath := os.Getenv("SEARCH_INDEX_PATH"); customPath != "" {
		fmt.Printf("📁 [SEARCH-INDEX] Using SEARCH_INDEX_PATH from environment: %s\n", customPath)
		return customPath
	}

	fmt.Printf("⚠️ [SEARCH-INDEX] SEARCH_INDEX_PATH not set, trying fallback paths...\n")

	// Default paths to try - MUST MATCH search_index_scheduler.go
	paths := []string{
		"../frontend/static/search-index.json",    // Local development
		"../../frontend/static/search-index.json", // Alternative local path
		"/app/frontend/static/search-index.json",  // Docker/container path
		"./static/search-index.json",              // Same directory static folder
		"./search-index.json",                     // Fallback to current directory
	}

	for _, path := range paths {
		// Check if file exists (not just directory)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("📁 [SEARCH-INDEX] Found existing index at: %s\n", path)
			return path
		}
		// Or check if directory exists for writing
		if dir := filepath.Dir(path); dirExists(dir) {
			fmt.Printf("📁 [SEARCH-INDEX] Using path (directory exists): %s\n", path)
			return path
		}
	}

	// Fallback
	return "./search-index.json"
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
