package routes

import (
	"net/http"
	"strconv"

	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupStripeCustomerSyncRoutes sets up the customer sync routes
func SetupStripeCustomerSyncRoutes(router *gin.RouterGroup, customerSyncService *services.StripeCustomerSyncService) {
	customers := router.Group("/customers")
	{
		// Sync individual customer to Stripe
		customers.POST("/:id/sync-to-stripe", func(c *gin.Context) {
			idStr := c.Param("id")
			customerID, err := strconv.Atoi(idStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
				return
			}

			result, err := customerSyncService.SyncCustomerToStripe(customerID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"result": result})
		})

		// Sync individual customer from Stripe
		customers.POST("/sync-from-stripe", func(c *gin.Context) {
			var req struct {
				StripeCustomerID string `json:"stripe_customer_id" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
				return
			}

			result, err := customerSyncService.SyncCustomerFromStripe(req.StripeCustomerID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"result": result})
		})

		// Bulk sync customers to Stripe
		customers.POST("/bulk-sync-to-stripe", func(c *gin.Context) {
			var req struct {
				CustomerIDs []int `json:"customer_ids" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
				return
			}

			stats, err := customerSyncService.BulkSyncCustomersToStripe(req.CustomerIDs)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"stats": stats})
		})

		// Bulk sync customers from Stripe
		customers.POST("/bulk-sync-from-stripe", func(c *gin.Context) {
			var req struct {
				StripeCustomerIDs []string `json:"stripe_customer_ids" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
				return
			}

			stats, err := customerSyncService.BulkSyncCustomersFromStripe(req.StripeCustomerIDs)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"stats": stats})
		})

		// Sync all customers (bidirectional)
		customers.POST("/sync-all", func(c *gin.Context) {
			stats, err := customerSyncService.SyncAllCustomers()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"stats": stats})
		})

		// Get sync status for a customer
		customers.GET("/:id/sync-status", func(c *gin.Context) {
			idStr := c.Param("id")
			customerID, err := strconv.Atoi(idStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
				return
			}

			result, err := customerSyncService.GetSyncStatus(customerID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"result": result})
		})
	}
}
