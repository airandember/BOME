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

// SimpleStripeSyncHandler handles simple Stripe sync operations
type SimpleStripeSyncHandler struct {
	syncService *stripeServices.SimpleStripeSyncService
	syncStatus  struct {
		mu        sync.RWMutex
		isRunning bool
		lastRun   time.Time
		lastError string
	}
}

// NewSimpleStripeSyncHandler creates a new simple sync handler
func NewSimpleStripeSyncHandler(syncService *stripeServices.SimpleStripeSyncService) *SimpleStripeSyncHandler {
	return &SimpleStripeSyncHandler{
		syncService: syncService,
	}
}

// SyncAllHandler handles complete Stripe data sync
func (h *SimpleStripeSyncHandler) SyncAllHandler(c *gin.Context) {
	// Check if sync is already running
	h.syncStatus.mu.RLock()
	if h.syncStatus.isRunning {
		h.syncStatus.mu.RUnlock()
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Sync already in progress",
			"message": "A sync operation is already running. Please wait for it to complete.",
		})
		return
	}
	h.syncStatus.mu.RUnlock()

	// Mark sync as running
	h.syncStatus.mu.Lock()
	h.syncStatus.isRunning = true
	h.syncStatus.lastError = ""
	h.syncStatus.mu.Unlock()

	// Return immediately and run sync in background
	c.JSON(http.StatusAccepted, gin.H{
		"message": "🔄 Stripe sync started in background. Check logs for progress or use /status endpoint.",
		"status":  "started",
	})

	// Run sync in background goroutine
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		log.Println("🚀 [SIMPLE-SYNC-HANDLER] Starting background Stripe sync...")
		err := h.syncService.SyncAll(ctx)

		// Update status
		h.syncStatus.mu.Lock()
		h.syncStatus.isRunning = false
		h.syncStatus.lastRun = time.Now()
		if err != nil {
			h.syncStatus.lastError = err.Error()
		} else {
			h.syncStatus.lastError = ""
		}
		h.syncStatus.mu.Unlock()

		if err != nil {
			log.Printf("❌ [SIMPLE-SYNC-HANDLER] Background sync failed: %v", err)
			log.Println("💡 [SIMPLE-SYNC-HANDLER] Check the error above and try running the sync again")
		} else {
			log.Println("✅ [SIMPLE-SYNC-HANDLER] Background sync completed successfully")
			log.Println("🎉 [SIMPLE-SYNC-HANDLER] SUCCESS: Stripe data sync is complete!")
			log.Println("📊 [SIMPLE-SYNC-HANDLER] You can now view the synchronized data at: /admin/streaming/subscribers/")
		}
	}()
}

// LinkCustomersHandler handles linking Stripe customers to local users
func (h *SimpleStripeSyncHandler) LinkCustomersHandler(c *gin.Context) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	log.Println("🔗 [SIMPLE-SYNC-HANDLER] Linking customers to users...")

	// Run linking
	err := h.syncService.LinkCustomersToUsers(ctx)
	if err != nil {
		log.Printf("❌ [SIMPLE-SYNC-HANDLER] Failed to link customers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to link customers to users",
			"details": err.Error(),
		})
		return
	}

	log.Println("✅ [SIMPLE-SYNC-HANDLER] Customer linking completed")

	c.JSON(http.StatusOK, gin.H{
		"message": "✅ Customer linking completed successfully",
		"status":  "success",
	})
}

// SyncStatusHandler returns the current sync status
func (h *SimpleStripeSyncHandler) SyncStatusHandler(c *gin.Context) {
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

	if status == "completed" {
		response["message"] = "✅ Sync completed successfully! Check your dashboard for updated data."
	} else if status == "running" {
		response["message"] = "🔄 Sync is currently running. This may take several minutes."
	} else if status == "failed" {
		response["message"] = "❌ Last sync failed. Check the error message for details."
	}

	c.JSON(http.StatusOK, response)
}

// SetupSimpleStripeSyncRoutes registers the simple sync routes
func SetupSimpleStripeSyncRoutes(stripeGroup *gin.RouterGroup, db *database.DB, stripeService *subscriptionServices.StripeService) {
	// Initialize service
	syncService := stripeServices.NewSimpleStripeSyncService(db, stripeService)
	handler := NewSimpleStripeSyncHandler(syncService)

	// Simple sync routes (under /admin/stripe/simple-sync/)
	syncGroup := stripeGroup.Group("/simple-sync")
	{
		syncGroup.POST("/all", handler.SyncAllHandler)
		syncGroup.GET("/status", handler.SyncStatusHandler)
		syncGroup.POST("/link-customers", handler.LinkCustomersHandler)
	}

	log.Println("✅ [SIMPLE-SYNC-ROUTES] Registered simple sync routes at /admin/stripe/simple-sync/")
}

