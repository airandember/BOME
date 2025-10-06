package routes

import (
	"context"
	"net/http"
	"time"

	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// ComprehensiveSyncHandler handles comprehensive Stripe synchronization
type ComprehensiveSyncHandler struct {
	syncService *services.ComprehensiveStripeSyncService
}

// NewComprehensiveSyncHandler creates a new comprehensive sync handler
func NewComprehensiveSyncHandler(syncService *services.ComprehensiveStripeSyncService) *ComprehensiveSyncHandler {
	return &ComprehensiveSyncHandler{
		syncService: syncService,
	}
}

// RunComprehensiveSync triggers a comprehensive Stripe synchronization
func (h *ComprehensiveSyncHandler) RunComprehensiveSync(c *gin.Context) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Minute)
	defer cancel()

	// Run the comprehensive sync
	result, err := h.syncService.RunComprehensiveSync(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to run comprehensive sync",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Comprehensive sync completed",
		"result":  result,
	})
}

// GetSyncStatus returns the current synchronization status
func (h *ComprehensiveSyncHandler) GetSyncStatus(c *gin.Context) {
	// This could be expanded to show real-time sync progress
	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"message": "Comprehensive sync service is ready",
	})
}

// RegisterComprehensiveSyncRoutes registers comprehensive sync routes
func RegisterComprehensiveSyncRoutes(router *gin.RouterGroup, syncService *services.ComprehensiveStripeSyncService) {
	handler := NewComprehensiveSyncHandler(syncService)

	sync := router.Group("/comprehensive-sync")
	{
		// Run comprehensive sync
		sync.POST("/run", handler.RunComprehensiveSync)

		// Get sync status
		sync.GET("/status", handler.GetSyncStatus)
	}
}
