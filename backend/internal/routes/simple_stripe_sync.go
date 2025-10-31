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

// SimpleStripeSyncHandler handles simple Stripe sync operations (v2 only)
type SimpleStripeSyncHandler struct {
	syncServiceV2          *services.StripeSyncV2Service
	customerLinkingService *services.CustomerLinkingService
	syncStatus             struct {
		mu        sync.RWMutex
		isRunning bool
		lastRun   time.Time
		lastError string
	}
}

// NewSimpleStripeSyncHandler creates a new simple sync handler (v2 only)
func NewSimpleStripeSyncHandler(syncServiceV2 *services.StripeSyncV2Service, customerLinkingService *services.CustomerLinkingService) *SimpleStripeSyncHandler {
	return &SimpleStripeSyncHandler{
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

		log.Println("🚀 Starting Stripe v2 sync (new architecture)...")

		var err error

		// Sync v2 tables only (v1 is deprecated)
		log.Println("📦 Step 1/2: Syncing Stripe data to v2 tables...")
		syncProgress, errV2 := h.syncServiceV2.SyncAll(ctx)
		if errV2 != nil {
			log.Printf("❌ V2 sync failed: %v", errV2)
			err = errV2 // Set error for status reporting
		} else {
			log.Println("✅ V2 tables synced successfully!")
			if syncProgress != nil {
				log.Printf("   📦 Products: %d synced", syncProgress.ProductsSynced)
				log.Printf("   💰 Prices: %d synced", syncProgress.PricesSynced)
				log.Printf("   👥 Customers: %d synced", syncProgress.CustomersSynced)
				log.Printf("   📋 Subscriptions: %d synced", syncProgress.SubscriptionsSynced)
			}
		}

		// Link customers to users (v2 only)
		log.Println("📦 Step 2/2: Linking customers to users...")
		linkResults, errLink := h.customerLinkingService.LinkAllUsers()
		if errLink != nil {
			log.Printf("❌ Customer linking failed: %v", errLink)
			if err == nil {
				err = errLink // Only set if no previous error
			}
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
			log.Printf("❌ Stripe v2 sync completed with errors: %v", err)
			log.Println("💡 Check the errors above and try running the sync again")
		} else {
			log.Println("")
			log.Println("═══════════════════════════════════════════════════")
			log.Println("🎉 SUCCESS: Stripe v2 Sync Complete!")
			log.Println("═══════════════════════════════════════════════════")
			log.Println("✅ All Stripe data synced to v2 tables")
			log.Println("✅ Customers linked to users")
			log.Println("")
			log.Println("📊 Data available at:")
			log.Println("   - User Dashboard: /user/subscriptions")
			log.Println("   - Admin API: /api/v1/admin/subscriber-elastic-v2")
			log.Println("═══════════════════════════════════════════════════")
			log.Println("")
		}
	}()
}

// LinkCustomersHandler handles linking Stripe customers to local users (v2 only)
func (h *SimpleStripeSyncHandler) LinkCustomersHandler(c *gin.Context) {
	// Run v2 customer linking
	linkResults, err := h.customerLinkingService.LinkAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to link customers to users",
			"details": err.Error(),
		})
		return
	}

	// Count successful links
	successCount := 0
	totalCustomers := 0
	for _, result := range linkResults {
		if result.Error == "" && result.CustomersLinked > 0 {
			successCount++
			totalCustomers += result.CustomersLinked
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "✅ Customer linking completed successfully",
		"status":           "success",
		"users_linked":     successCount,
		"customers_linked": totalCustomers,
	})
}

// RegisterSimpleStripeSyncRoutes registers the simple sync routes (v2 only)
func RegisterSimpleStripeSyncRoutes(router *gin.RouterGroup, syncServiceV2 *services.StripeSyncV2Service, customerLinkingService *services.CustomerLinkingService) {
	handler := NewSimpleStripeSyncHandler(syncServiceV2, customerLinkingService)

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
