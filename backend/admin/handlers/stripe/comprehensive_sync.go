package stripe

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"bome-backend/infrastructure/database"
	stripeServices "bome-backend/services/stripe"
	subscriptionServices "bome-backend/subscription/services"

	"github.com/gin-gonic/gin"
)

// ComprehensiveSyncHandler handles comprehensive Stripe synchronization
type ComprehensiveSyncHandler struct {
	syncService *stripeServices.ComprehensiveStripeSyncService
	syncStatus  struct {
		mu         sync.RWMutex
		isRunning  bool
		lastRun    time.Time
		lastError  string
		lastResult *stripeServices.ComprehensiveSyncResult
	}
}

// NewComprehensiveSyncHandler creates a new comprehensive sync handler
func NewComprehensiveSyncHandler(syncService *stripeServices.ComprehensiveStripeSyncService) *ComprehensiveSyncHandler {
	return &ComprehensiveSyncHandler{
		syncService: syncService,
	}
}

// RunComprehensiveSync triggers a comprehensive Stripe synchronization
func (h *ComprehensiveSyncHandler) RunComprehensiveSync(c *gin.Context) {
	// Check if sync is already running
	h.syncStatus.mu.RLock()
	if h.syncStatus.isRunning {
		h.syncStatus.mu.RUnlock()
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Sync already in progress",
			"message": "A comprehensive sync is already running. Please wait for it to complete.",
		})
		return
	}
	h.syncStatus.mu.RUnlock()

	// Mark sync as running
	h.syncStatus.mu.Lock()
	h.syncStatus.isRunning = true
	h.syncStatus.lastError = ""
	h.syncStatus.lastResult = nil
	h.syncStatus.mu.Unlock()

	// Return immediately and run sync in background
	c.JSON(http.StatusAccepted, gin.H{
		"message": "🔄 Comprehensive sync started in background. Check /status for progress.",
		"status":  "started",
	})

	// Run sync in background goroutine
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		log.Println("🚀 [COMPREHENSIVE-SYNC-HANDLER] Starting comprehensive sync...")

		// Run the comprehensive sync
		result, err := h.syncService.RunComprehensiveSync(ctx)

		// Update status
		h.syncStatus.mu.Lock()
		h.syncStatus.isRunning = false
		h.syncStatus.lastRun = time.Now()
		if err != nil {
			h.syncStatus.lastError = err.Error()
			h.syncStatus.lastResult = nil
		} else {
			h.syncStatus.lastError = ""
			h.syncStatus.lastResult = result
		}
		h.syncStatus.mu.Unlock()

		if err != nil {
			log.Printf("❌ [COMPREHENSIVE-SYNC-HANDLER] Failed: %v", err)
		} else {
			log.Println("✅ [COMPREHENSIVE-SYNC-HANDLER] Completed successfully")
			log.Printf("📊 [COMPREHENSIVE-SYNC-HANDLER] Results: %d linked, %d plans fixed, %d errors",
				result.NewlyLinked, result.FixedPlans, result.Errors)
		}
	}()
}

// GetSyncStatus returns the current synchronization status
func (h *ComprehensiveSyncHandler) GetSyncStatus(c *gin.Context) {
	h.syncStatus.mu.RLock()
	defer h.syncStatus.mu.RUnlock()

	status := "idle"
	if h.syncStatus.isRunning {
		status = "running"
	} else if !h.syncStatus.lastRun.IsZero() {
		if h.syncStatus.lastError != "" {
			status = "failed"
		} else {
			status = "completed"
		}
	}

	response := gin.H{
		"status":    status,
		"isRunning": h.syncStatus.isRunning,
	}

	if !h.syncStatus.lastRun.IsZero() {
		response["lastRun"] = h.syncStatus.lastRun.Format(time.RFC3339)
	}

	if h.syncStatus.lastError != "" {
		response["lastError"] = h.syncStatus.lastError
	}

	if h.syncStatus.lastResult != nil {
		response["lastResult"] = h.syncStatus.lastResult
	}

	if status == "completed" && h.syncStatus.lastResult != nil {
		response["message"] = "✅ Comprehensive sync completed successfully!"
		response["summary"] = gin.H{
			"totalUsers":     h.syncStatus.lastResult.TotalUsers,
			"newlyLinked":    h.syncStatus.lastResult.NewlyLinked,
			"fixedPlans":     h.syncStatus.lastResult.FixedPlans,
			"ghostCustomers": h.syncStatus.lastResult.GhostCustomers,
			"errors":         h.syncStatus.lastResult.Errors,
		}
	} else if status == "running" {
		response["message"] = "🔄 Comprehensive sync is currently running..."
	} else if status == "failed" {
		response["message"] = "❌ Last sync failed. Check the error message for details."
	}

	c.JSON(http.StatusOK, response)
}

// SetupComprehensiveSyncRoutes registers comprehensive sync routes
func SetupComprehensiveSyncRoutes(stripeGroup *gin.RouterGroup, db *database.DB, stripeService *subscriptionServices.StripeService) {
	// Initialize service
	syncService := stripeServices.NewComprehensiveStripeSyncService(db, stripeService)
	handler := NewComprehensiveSyncHandler(syncService)

	// Comprehensive sync routes (under /admin/stripe/comprehensive-sync/)
	syncGroup := stripeGroup.Group("/comprehensive-sync")
	{
		// Run comprehensive sync
		syncGroup.POST("/run", handler.RunComprehensiveSync)

		// Get sync status
		syncGroup.GET("/status", handler.GetSyncStatus)
	}

	log.Println("✅ [COMPREHENSIVE-SYNC-ROUTES] Registered comprehensive sync routes at /admin/stripe/comprehensive-sync/")
}
