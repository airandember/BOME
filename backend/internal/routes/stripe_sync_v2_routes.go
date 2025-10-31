package routes

import (
	"context"
	"log"
	"net/http"
	"time"

	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupStripeSyncV2Routes configures routes for Stripe v2 sync
func SetupStripeSyncV2Routes(router *gin.RouterGroup, db *database.DB) {
	syncV2 := router.Group("/stripe-v2")
	syncV2.Use(middleware.AuthRequired())
	syncV2.Use(middleware.AdminRequired())

	// GET /admin/stripe-v2/status - Get sync status
	syncV2.GET("/status", func(c *gin.Context) {
		getSyncStatusV2(c, db)
	})

	// POST /admin/stripe-v2/sync - Trigger full sync
	syncV2.POST("/sync", func(c *gin.Context) {
		triggerSyncV2(c, db)
	})

	// POST /admin/stripe-v2/sync-products - Sync products only
	syncV2.POST("/sync-products", func(c *gin.Context) {
		syncProductsV2(c, db)
	})

	// POST /admin/stripe-v2/sync-prices - Sync prices only
	syncV2.POST("/sync-prices", func(c *gin.Context) {
		syncPricesV2(c, db)
	})

	// POST /admin/stripe-v2/sync-customers - Sync customers only
	syncV2.POST("/sync-customers", func(c *gin.Context) {
		syncCustomersV2(c, db)
	})

	// POST /admin/stripe-v2/sync-subscriptions - Sync subscriptions only
	syncV2.POST("/sync-subscriptions", func(c *gin.Context) {
		syncSubscriptionsV2(c, db)
	})
}

// getSyncStatusV2 returns the current status of v2 tables
func getSyncStatusV2(c *gin.Context, db *database.DB) {
	type TableStats struct {
		ProductsCount      int        `json:"products_count"`
		PricesCount        int        `json:"prices_count"`
		CustomersCount     int        `json:"customers_count"`
		SubscriptionsCount int        `json:"subscriptions_count"`
		LastSyncAt         *time.Time `json:"last_sync_at"`
	}

	var stats TableStats

	// Get counts from v2 tables
	db.QueryRow("SELECT COUNT(*) FROM stripe_products_v2").Scan(&stats.ProductsCount)
	db.QueryRow("SELECT COUNT(*) FROM stripe_prices_v2").Scan(&stats.PricesCount)
	db.QueryRow("SELECT COUNT(*) FROM stripe_customers_v2").Scan(&stats.CustomersCount)
	db.QueryRow("SELECT COUNT(*) FROM stripe_subscriptions_v2").Scan(&stats.SubscriptionsCount)

	// Get last sync time (most recent last_synced_at across all tables)
	db.QueryRow(`
		SELECT MAX(last_synced_at) FROM (
			SELECT MAX(last_synced_at) as last_synced_at FROM stripe_products_v2
			UNION ALL
			SELECT MAX(last_synced_at) FROM stripe_prices_v2
			UNION ALL
			SELECT MAX(last_synced_at) FROM stripe_customers_v2
			UNION ALL
			SELECT MAX(last_synced_at) FROM stripe_subscriptions_v2
		) as all_syncs
	`).Scan(&stats.LastSyncAt)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// triggerSyncV2 triggers a full Stripe sync
func triggerSyncV2(c *gin.Context, db *database.DB) {
	log.Printf("🚀 [Admin] Full Stripe v2 sync triggered by user")

	syncService := services.NewStripeSyncV2Service(db)
	ctx := context.Background()

	// Run sync in background (with timeout)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	progress, err := syncService.SyncAll(ctx)
	if err != nil {
		log.Printf("❌ [Admin] Sync failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Sync failed",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Sync completed",
		"data":    progress,
	})
}

// syncProductsV2 syncs products only
func syncProductsV2(c *gin.Context, db *database.DB) {
	log.Printf("📦 [Admin] Products sync triggered by user")

	syncService := services.NewStripeSyncV2Service(db)
	ctx := context.Background()

	progress := &services.SyncProgress{
		StartedAt: time.Now(),
		Errors:    []string{},
	}

	if err := syncService.SyncProducts(ctx, progress); err != nil {
		log.Printf("❌ [Admin] Products sync failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Products sync failed",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Products synced",
		"data":    progress,
	})
}

// syncPricesV2 syncs prices only
func syncPricesV2(c *gin.Context, db *database.DB) {
	log.Printf("💰 [Admin] Prices sync triggered by user")

	syncService := services.NewStripeSyncV2Service(db)
	ctx := context.Background()

	progress := &services.SyncProgress{
		StartedAt: time.Now(),
		Errors:    []string{},
	}

	if err := syncService.SyncPrices(ctx, progress); err != nil {
		log.Printf("❌ [Admin] Prices sync failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Prices sync failed",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Prices synced",
		"data":    progress,
	})
}

// syncCustomersV2 syncs customers only
func syncCustomersV2(c *gin.Context, db *database.DB) {
	log.Printf("👥 [Admin] Customers sync triggered by user")

	syncService := services.NewStripeSyncV2Service(db)
	ctx := context.Background()

	progress := &services.SyncProgress{
		StartedAt: time.Now(),
		Errors:    []string{},
	}

	if err := syncService.SyncCustomers(ctx, progress); err != nil {
		log.Printf("❌ [Admin] Customers sync failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Customers sync failed",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Customers synced",
		"data":    progress,
	})
}

// syncSubscriptionsV2 syncs subscriptions only
func syncSubscriptionsV2(c *gin.Context, db *database.DB) {
	log.Printf("📋 [Admin] Subscriptions sync triggered by user")

	syncService := services.NewStripeSyncV2Service(db)
	ctx := context.Background()

	progress := &services.SyncProgress{
		StartedAt: time.Now(),
		Errors:    []string{},
	}

	if err := syncService.SyncSubscriptions(ctx, progress); err != nil {
		log.Printf("❌ [Admin] Subscriptions sync failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Subscriptions sync failed",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Subscriptions synced",
		"data":    progress,
	})
}
