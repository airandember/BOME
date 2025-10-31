package routes

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SimpleStripeSyncHandler handles simple Stripe sync operations
type SimpleStripeSyncHandler struct {
	syncService            *services.SimpleStripeSyncService
	syncServiceV2          *services.StripeSyncV2Service
	customerLinkingService *services.CustomerLinkingService
	syncStatus             struct {
		mu        sync.RWMutex
		isRunning bool
		lastRun   time.Time
		lastError string
	}
}

// NewSimpleStripeSyncHandler creates a new simple sync handler
func NewSimpleStripeSyncHandler(syncService *services.SimpleStripeSyncService, syncServiceV2 *services.StripeSyncV2Service, customerLinkingService *services.CustomerLinkingService) *SimpleStripeSyncHandler {
	return &SimpleStripeSyncHandler{
		syncService:            syncService,
		syncServiceV2:          syncServiceV2,
		customerLinkingService: customerLinkingService,
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

		log.Println("🚀 Starting background Stripe sync (v1 + v2 tables)...")

		// Sync v1 tables (legacy)
		log.Println("📦 Syncing v1 tables...")
		err := h.syncService.SyncAll(ctx)
		if err != nil {
			log.Printf("⚠️  V1 sync had errors: %v", err)
			// Don't fail completely - continue to v2
		} else {
			log.Println("✅ V1 tables synced successfully")
		}

		// Sync v2 tables (new architecture)
		log.Println("📦 Syncing v2 tables...")
		_, errV2 := h.syncServiceV2.SyncAll(ctx)
		if errV2 != nil {
			log.Printf("⚠️  V2 sync had errors: %v", errV2)
			err = errV2 // Set error for status reporting
		} else {
			log.Println("✅ V2 tables synced successfully")
		}

		// Link customers to users (v2 only)
		log.Println("🔗 Linking customers to users...")
		linkResults, errLink := h.customerLinkingService.LinkAllUsers()
		if errLink != nil {
			log.Printf("⚠️  Customer linking had errors: %v", errLink)
			err = errLink
		} else {
			// Count successful links and total customers
			successCount := 0
			totalCustomers := 0
			for _, result := range linkResults {
				if result.Error == "" && result.CustomersLinked > 0 {
					successCount++
					totalCustomers += result.CustomersLinked
				}
			}
			log.Printf("✅ Successfully linked %d users to %d Stripe customers", successCount, totalCustomers)
		}

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
			log.Printf("❌ Background sync completed with errors: %v", err)
			log.Println("💡 Check the errors above and try running the sync again")
		} else {
			log.Println("✅ Background Stripe sync completed successfully!")
			log.Println("🎉 SUCCESS: V1 + V2 tables synced and customers linked!")
			log.Println("📊 V1 data: /admin/streaming/subscribers/customers/")
			log.Println("📊 V2 data: /user/subscriptions (user dashboard)")
		}
	}()
}

// LinkCustomersHandler handles linking Stripe customers to local users
func (h *SimpleStripeSyncHandler) LinkCustomersHandler(c *gin.Context) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Run linking
	err := h.syncService.LinkCustomersToUsers(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to link customers to users",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "✅ Customer linking completed successfully",
		"status":  "success",
	})
}

// RegisterSimpleStripeSyncRoutes registers the simple sync routes
func RegisterSimpleStripeSyncRoutes(router *gin.RouterGroup, syncService *services.SimpleStripeSyncService, syncServiceV2 *services.StripeSyncV2Service, customerLinkingService *services.CustomerLinkingService) {
	handler := NewSimpleStripeSyncHandler(syncService, syncServiceV2, customerLinkingService)

	// Simple sync routes
	router.POST("/simple-sync/all", handler.SyncAllHandler)
	router.GET("/simple-sync/status", handler.SyncStatusHandler)
	router.POST("/simple-sync/link-customers", handler.LinkCustomersHandler)
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
		response["message"] = "✅ Sync completed successfully! Check your Customer Dashboard for updated plan names."
	}

	c.JSON(http.StatusOK, response)
}
