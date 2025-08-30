package routes

import (
	"net/http"
	"strconv"

	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// RegisterStripeSystemRoutes registers system management routes
func RegisterStripeSystemRoutes(router *gin.RouterGroup, syncService *services.StripeSyncService) {
	system := router.Group("/system")
	{
		// Sync Jobs Management
		system.GET("/jobs", func(c *gin.Context) { getSystemSyncJobs(c, syncService) })
		system.GET("/jobs/:id", func(c *gin.Context) { getSystemSyncJob(c, syncService) })
		system.DELETE("/jobs/:id", func(c *gin.Context) { cancelSystemSyncJob(c, syncService) })

		// Sync Configuration Management
		system.GET("/config", func(c *gin.Context) { getSystemSyncConfigs(c, syncService) })
		system.GET("/config/:entity_type", func(c *gin.Context) { getSystemSyncConfig(c, syncService) })
		system.PUT("/config/:entity_type", func(c *gin.Context) { updateSystemSyncConfig(c, syncService) })

		// Entity Tracking
		system.GET("/entities", func(c *gin.Context) { getStripeEntities(c, syncService) })
		system.GET("/entities/:entity_type", func(c *gin.Context) { getStripeEntitiesByType(c, syncService) })
		system.GET("/entities/status/:status", func(c *gin.Context) { getStripeEntitiesByStatus(c, syncService) })

		// System Health & Stats
		system.GET("/health", func(c *gin.Context) { getSystemHealth(c, syncService) })
		system.GET("/stats", func(c *gin.Context) { getSystemStats(c, syncService) })
	}
}

// ===== SYNC JOBS ENDPOINTS =====

func getSystemSyncJobs(c *gin.Context, syncService *services.StripeSyncService) {
	// Get query parameters
	status := c.Query("status")
	entityType := c.Query("entity_type")
	limit := c.DefaultQuery("limit", "50")

	limitInt, _ := strconv.Atoi(limit)

	jobs, err := syncService.GetSyncJobs(status, entityType, limitInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"jobs":  jobs,
		"total": len(jobs),
	})
}

func getSystemSyncJob(c *gin.Context, syncService *services.StripeSyncService) {
	jobIDStr := c.Param("id")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	job, err := syncService.GetSyncJobByID(jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	c.JSON(http.StatusOK, job)
}

func cancelSystemSyncJob(c *gin.Context, syncService *services.StripeSyncService) {
	jobIDStr := c.Param("id")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	err = syncService.CancelSyncJob(jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job cancelled successfully"})
}

// ===== SYNC CONFIGURATION ENDPOINTS =====

func getSystemSyncConfigs(c *gin.Context, syncService *services.StripeSyncService) {
	configs, err := syncService.GetAllSyncConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

func getSystemSyncConfig(c *gin.Context, syncService *services.StripeSyncService) {
	entityType := c.Param("entity_type")

	config, err := syncService.GetSyncConfigByType(entityType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	c.JSON(http.StatusOK, config)
}

func updateSystemSyncConfig(c *gin.Context, syncService *services.StripeSyncService) {
	entityType := c.Param("entity_type")

	var updateReq struct {
		SyncEnabled       *bool                  `json:"sync_enabled"`
		SyncIntervalHours *int                   `json:"sync_interval_hours"`
		BatchSize         *int                   `json:"batch_size"`
		ConfigData        map[string]interface{} `json:"config_data"`
	}

	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := syncService.UpdateSyncConfig(entityType, updateReq.SyncEnabled, updateReq.SyncIntervalHours, updateReq.BatchSize, updateReq.ConfigData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Config updated successfully"})
}

// ===== ENTITY TRACKING ENDPOINTS =====

func getStripeEntities(c *gin.Context, syncService *services.StripeSyncService) {
	limit := c.DefaultQuery("limit", "100")
	offset := c.DefaultQuery("offset", "0")

	limitInt, _ := strconv.Atoi(limit)
	offsetInt, _ := strconv.Atoi(offset)

	entities, total, err := syncService.GetStripeEntities(limitInt, offsetInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entities": entities,
		"total":    total,
		"limit":    limitInt,
		"offset":   offsetInt,
	})
}

func getStripeEntitiesByType(c *gin.Context, syncService *services.StripeSyncService) {
	entityType := c.Param("entity_type")
	limit := c.DefaultQuery("limit", "100")
	offset := c.DefaultQuery("offset", "0")

	limitInt, _ := strconv.Atoi(limit)
	offsetInt, _ := strconv.Atoi(offset)

	entities, total, err := syncService.GetStripeEntitiesByType(entityType, limitInt, offsetInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entities":    entities,
		"entity_type": entityType,
		"total":       total,
		"limit":       limitInt,
		"offset":      offsetInt,
	})
}

func getStripeEntitiesByStatus(c *gin.Context, syncService *services.StripeSyncService) {
	status := c.Param("status")
	limit := c.DefaultQuery("limit", "100")
	offset := c.DefaultQuery("offset", "0")

	limitInt, _ := strconv.Atoi(limit)
	offsetInt, _ := strconv.Atoi(offset)

	entities, total, err := syncService.GetStripeEntitiesByStatus(status, limitInt, offsetInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entities": entities,
		"status":   status,
		"total":    total,
		"limit":    limitInt,
		"offset":   offsetInt,
	})
}

// ===== SYSTEM HEALTH & STATS ENDPOINTS =====

func getSystemHealth(c *gin.Context, syncService *services.StripeSyncService) {
	health, err := syncService.GetSystemHealth()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, health)
}

func getSystemStats(c *gin.Context, syncService *services.StripeSyncService) {
	stats, err := syncService.GetSystemStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
