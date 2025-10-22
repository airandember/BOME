package stripe

import (
	"log"
	"net/http"
	"strconv"

	"bome-backend/infrastructure/database"
	stripeServices "bome-backend/services/stripe"
	subServices "bome-backend/subscription/services"

	"github.com/gin-gonic/gin"
)

// WebSocketHub interface for real-time updates
type WebSocketHub interface {
	BroadcastEvent(eventType string, data map[string]interface{}, message string)
}

// SetupStripeCustomerSyncRoutes sets up Stripe customer sync routes
func SetupStripeCustomerSyncRoutes(router *gin.RouterGroup, db *database.DB, stripeService *subServices.StripeService, wsHub WebSocketHub) {
	service := stripeServices.NewStripeCustomerSyncService(db, stripeService, wsHub)

	customers := router.Group("/customers")
	{
		// POST /admin/streaming/stripe/customers/:id/sync-to-stripe
		customers.POST("/:id/sync-to-stripe", func(c *gin.Context) {
			SyncCustomerToStripeHandler(c, service)
		})

		// POST /admin/streaming/stripe/customers/sync-from-stripe
		customers.POST("/sync-from-stripe", func(c *gin.Context) {
			SyncCustomerFromStripeHandler(c, service)
		})

		// POST /admin/streaming/stripe/customers/bulk-sync-to-stripe
		customers.POST("/bulk-sync-to-stripe", func(c *gin.Context) {
			BulkSyncCustomersToStripeHandler(c, service)
		})

		// POST /admin/streaming/stripe/customers/bulk-sync-from-stripe
		customers.POST("/bulk-sync-from-stripe", func(c *gin.Context) {
			BulkSyncCustomersFromStripeHandler(c, service)
		})

		// GET /admin/streaming/stripe/customers/:id/sync-status
		customers.GET("/:id/sync-status", func(c *gin.Context) {
			GetSyncStatusHandler(c, service)
		})
	}
}

// SyncCustomerToStripeHandler handles POST /admin/streaming/stripe/customers/:id/sync-to-stripe
func SyncCustomerToStripeHandler(c *gin.Context, service *stripeServices.StripeCustomerSyncService) {
	idStr := c.Param("id")
	customerID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid customer ID",
		})
		return
	}

	log.Printf("🔄 SyncCustomerToStripeHandler: Syncing customer %d to Stripe", customerID)

	result, err := service.SyncCustomerToStripe(customerID)
	if err != nil {
		log.Printf("❌ Error syncing customer to Stripe: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to sync customer to Stripe",
			"result": result,
		})
		return
	}

	log.Printf("✅ Customer synced to Stripe: %s (action: %s)", result.StripeID, result.Action)
	c.JSON(http.StatusOK, gin.H{
		"result": result,
	})
}

// SyncCustomerFromStripeHandler handles POST /admin/streaming/stripe/customers/sync-from-stripe
func SyncCustomerFromStripeHandler(c *gin.Context, service *stripeServices.StripeCustomerSyncService) {
	var req struct {
		StripeCustomerID string `json:"stripe_customer_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body. Required: stripe_customer_id",
		})
		return
	}

	log.Printf("🔄 SyncCustomerFromStripeHandler: Syncing customer %s from Stripe", req.StripeCustomerID)

	result, err := service.SyncCustomerFromStripe(req.StripeCustomerID)
	if err != nil {
		log.Printf("❌ Error syncing customer from Stripe: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to sync customer from Stripe",
			"result": result,
		})
		return
	}

	log.Printf("✅ Customer synced from Stripe: %s (action: %s)", req.StripeCustomerID, result.Action)
	c.JSON(http.StatusOK, gin.H{
		"result": result,
	})
}

// BulkSyncCustomersToStripeHandler handles POST /admin/streaming/stripe/customers/bulk-sync-to-stripe
func BulkSyncCustomersToStripeHandler(c *gin.Context, service *stripeServices.StripeCustomerSyncService) {
	var req struct {
		CustomerIDs []int `json:"customer_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body. Required: customer_ids (array of integers)",
		})
		return
	}

	if len(req.CustomerIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "customer_ids array cannot be empty",
		})
		return
	}

	log.Printf("🔄 BulkSyncCustomersToStripeHandler: Syncing %d customers to Stripe", len(req.CustomerIDs))

	stats, err := service.BulkSyncCustomersToStripe(req.CustomerIDs)
	if err != nil {
		log.Printf("❌ Error during bulk sync to Stripe: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Bulk sync failed",
			"stats": stats,
		})
		return
	}

	log.Printf("✅ Bulk sync to Stripe completed: %d created, %d updated, %d errors", stats.Created, stats.Updated, stats.Errors)
	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// BulkSyncCustomersFromStripeHandler handles POST /admin/streaming/stripe/customers/bulk-sync-from-stripe
func BulkSyncCustomersFromStripeHandler(c *gin.Context, service *stripeServices.StripeCustomerSyncService) {
	var req struct {
		StripeCustomerIDs []string `json:"stripe_customer_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body. Required: stripe_customer_ids (array of strings)",
		})
		return
	}

	if len(req.StripeCustomerIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "stripe_customer_ids array cannot be empty",
		})
		return
	}

	log.Printf("🔄 BulkSyncCustomersFromStripeHandler: Syncing %d customers from Stripe", len(req.StripeCustomerIDs))

	stats, err := service.BulkSyncCustomersFromStripe(req.StripeCustomerIDs)
	if err != nil {
		log.Printf("❌ Error during bulk sync from Stripe: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Bulk sync failed",
			"stats": stats,
		})
		return
	}

	log.Printf("✅ Bulk sync from Stripe completed: %d created, %d updated, %d errors", stats.Created, stats.Updated, stats.Errors)
	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// GetSyncStatusHandler handles GET /admin/streaming/stripe/customers/:id/sync-status
func GetSyncStatusHandler(c *gin.Context, service *stripeServices.StripeCustomerSyncService) {
	idStr := c.Param("id")
	customerID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid customer ID",
		})
		return
	}

	log.Printf("📊 GetSyncStatusHandler: Getting sync status for customer %d", customerID)

	result, err := service.GetSyncStatus(customerID)
	if err != nil {
		log.Printf("❌ Error getting sync status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get sync status",
		})
		return
	}

	log.Printf("✅ Sync status retrieved: %s (action: %s)", result.StripeID, result.Action)
	c.JSON(http.StatusOK, gin.H{
		"result": result,
	})
}
