package stripe

import (
	"log"
	"net/http"
	"strconv"

	"bome-backend/infrastructure/database"
	stripeServices "bome-backend/services/stripe"

	"github.com/gin-gonic/gin"
)

// SetupStripeDatabaseRoutes sets up Stripe database query routes
func SetupStripeDatabaseRoutes(router *gin.RouterGroup, db *database.DB) {
	service := stripeServices.NewStripeDatabaseService(db)

	database := router.Group("/database")
	{
		// GET /admin/streaming/stripe/database/customers
		database.GET("/customers", func(c *gin.Context) {
			GetDatabaseCustomersHandler(c, service)
		})

		// GET /admin/streaming/stripe/database/subscriptions
		database.GET("/subscriptions", func(c *gin.Context) {
			GetDatabaseSubscriptionsHandler(c, service)
		})

		// GET /admin/streaming/stripe/database/stats
		database.GET("/stats", func(c *gin.Context) {
			GetDatabaseStatsHandler(c, service)
		})

		// GET /admin/streaming/stripe/database/customers/:stripe_id
		database.GET("/customers/:stripe_id", func(c *gin.Context) {
			GetDatabaseCustomerByIDHandler(c, service)
		})
	}
}

// GetDatabaseCustomersHandler handles GET /admin/streaming/stripe/database/customers
func GetDatabaseCustomersHandler(c *gin.Context, service *stripeServices.StripeDatabaseService) {
	log.Println("📊 GetDatabaseCustomersHandler called")

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")
	includeSubscriptions := c.DefaultQuery("include_subscriptions", "true") == "true"

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000 // Cap at 1000 for performance
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get customers from database
	customers, totalCount, err := service.GetCustomers(limit, offset, includeSubscriptions)
	if err != nil {
		log.Printf("❌ Error getting customers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get customers from database",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Retrieved %d customers (total: %d)", len(customers), totalCount)
	c.JSON(http.StatusOK, gin.H{
		"customers": customers,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"total":  totalCount,
		},
		"source": "database",
	})
}

// GetDatabaseSubscriptionsHandler handles GET /admin/streaming/stripe/database/subscriptions
func GetDatabaseSubscriptionsHandler(c *gin.Context, service *stripeServices.StripeDatabaseService) {
	log.Println("📊 GetDatabaseSubscriptionsHandler called")

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")
	status := c.Query("status") // Optional: active, canceled, trialing, etc.

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get subscriptions from database
	subscriptions, totalCount, err := service.GetSubscriptions(limit, offset, status)
	if err != nil {
		log.Printf("❌ Error getting subscriptions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get subscriptions from database",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Retrieved %d subscriptions (total: %d, status: %s)", len(subscriptions), totalCount, status)
	c.JSON(http.StatusOK, gin.H{
		"subscriptions": subscriptions,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"total":  totalCount,
		},
		"source": "database",
	})
}

// GetDatabaseStatsHandler handles GET /admin/streaming/stripe/database/stats
func GetDatabaseStatsHandler(c *gin.Context, service *stripeServices.StripeDatabaseService) {
	log.Println("📊 GetDatabaseStatsHandler called")

	stats, err := service.GetStats()
	if err != nil {
		log.Printf("❌ Error getting database stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get database statistics",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Retrieved database stats: %d customers, %d subscriptions (%d active)",
		stats.TotalCustomers, stats.TotalSubscriptions, stats.ActiveSubscriptions)

	c.JSON(http.StatusOK, stats)
}

// GetDatabaseCustomerByIDHandler handles GET /admin/streaming/stripe/database/customers/:stripe_id
func GetDatabaseCustomerByIDHandler(c *gin.Context, service *stripeServices.StripeDatabaseService) {
	stripeID := c.Param("stripe_id")
	log.Printf("📊 GetDatabaseCustomerByIDHandler called for: %s", stripeID)

	customer, err := service.GetCustomerByStripeID(stripeID)
	if err != nil {
		log.Printf("❌ Error getting customer: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Customer not found",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Retrieved customer: %s (%s)", customer.Email, customer.StripeID)
	c.JSON(http.StatusOK, customer)
}
