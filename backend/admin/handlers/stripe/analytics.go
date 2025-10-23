package stripe

import (
	"log"
	"net/http"

	"bome-backend/infrastructure/database"
	stripeServices "bome-backend/services/stripe"
	subServices "bome-backend/subscription/services"

	"github.com/gin-gonic/gin"
)

// SetupStripeAnalyticsRoutes sets up Stripe analytics routes
func SetupStripeAnalyticsRoutes(router *gin.RouterGroup, db *database.DB, stripeService *subServices.StripeService) {
	service := stripeServices.NewStripeAnalyticsService(db, stripeService)

	analytics := router.Group("/analytics")
	{
		// GET /admin/streaming/stripe/analytics/dashboard
		analytics.GET("/dashboard", func(c *gin.Context) {
			GetDashboardHandler(c, service)
		})

		// GET /admin/streaming/stripe/analytics/balance
		analytics.GET("/balance", func(c *gin.Context) {
			GetBalanceHandler(c, service)
		})

		// GET /admin/streaming/stripe/analytics/customers
		analytics.GET("/customers", func(c *gin.Context) {
			GetCustomerMetricsHandler(c, service)
		})

		// GET /admin/streaming/stripe/analytics/subscriptions
		analytics.GET("/subscriptions", func(c *gin.Context) {
			GetSubscriptionMetricsHandler(c, service)
		})

		// GET /admin/streaming/stripe/analytics/revenue
		analytics.GET("/revenue", func(c *gin.Context) {
			GetRevenueMetricsHandler(c, service)
		})

		// GET /admin/streaming/stripe/analytics/health
		analytics.GET("/health", func(c *gin.Context) {
			GetHealthHandler(c, service)
		})
	}

	log.Println("✅ Stripe analytics routes registered")
}

// GetDashboardHandler handles GET /admin/streaming/stripe/analytics/dashboard
func GetDashboardHandler(c *gin.Context, service *stripeServices.StripeAnalyticsService) {
	log.Println("📊 GetDashboardHandler: Fetching comprehensive dashboard metrics...")

	metrics, err := service.GetDashboardMetrics()
	if err != nil {
		log.Printf("❌ Error fetching dashboard metrics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch dashboard metrics",
		})
		return
	}

	log.Printf("✅ Dashboard metrics retrieved successfully")
	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
	})
}

// GetBalanceHandler handles GET /admin/streaming/stripe/analytics/balance
func GetBalanceHandler(c *gin.Context, service *stripeServices.StripeAnalyticsService) {
	log.Println("💰 GetBalanceHandler: Fetching Stripe balance...")

	balance, err := service.GetBalance()
	if err != nil {
		log.Printf("❌ Error fetching balance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch balance",
		})
		return
	}

	log.Printf("✅ Balance retrieved: $%.2f USD", balance.TotalUSD)
	c.JSON(http.StatusOK, gin.H{
		"balance": balance,
	})
}

// GetCustomerMetricsHandler handles GET /admin/streaming/stripe/analytics/customers
func GetCustomerMetricsHandler(c *gin.Context, service *stripeServices.StripeAnalyticsService) {
	log.Println("👥 GetCustomerMetricsHandler: Fetching customer metrics...")

	metrics, err := service.GetCustomerMetrics()
	if err != nil {
		log.Printf("❌ Error fetching customer metrics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch customer metrics",
		})
		return
	}

	log.Printf("✅ Customer metrics retrieved: %d total, %d active", metrics.TotalCustomers, metrics.ActiveSubscribers)
	c.JSON(http.StatusOK, gin.H{
		"customers": metrics,
	})
}

// GetSubscriptionMetricsHandler handles GET /admin/streaming/stripe/analytics/subscriptions
func GetSubscriptionMetricsHandler(c *gin.Context, service *stripeServices.StripeAnalyticsService) {
	log.Println("📋 GetSubscriptionMetricsHandler: Fetching subscription metrics...")

	metrics, err := service.GetSubscriptionMetrics()
	if err != nil {
		log.Printf("❌ Error fetching subscription metrics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch subscription metrics",
		})
		return
	}

	log.Printf("✅ Subscription metrics retrieved: %d total, %d active", metrics.TotalSubscriptions, metrics.ActiveSubscriptions)
	c.JSON(http.StatusOK, gin.H{
		"subscriptions": metrics,
	})
}

// GetRevenueMetricsHandler handles GET /admin/streaming/stripe/analytics/revenue
func GetRevenueMetricsHandler(c *gin.Context, service *stripeServices.StripeAnalyticsService) {
	log.Println("💵 GetRevenueMetricsHandler: Fetching revenue metrics...")

	metrics, err := service.GetRevenueMetrics()
	if err != nil {
		log.Printf("❌ Error fetching revenue metrics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch revenue metrics",
		})
		return
	}

	log.Printf("✅ Revenue metrics retrieved: MRR $%.2f, ARR $%.2f", metrics.MRR, metrics.ARR)
	c.JSON(http.StatusOK, gin.H{
		"revenue": metrics,
	})
}

// GetHealthHandler handles GET /admin/streaming/stripe/analytics/health
func GetHealthHandler(c *gin.Context, service *stripeServices.StripeAnalyticsService) {
	log.Println("🏥 GetHealthHandler: Checking Stripe health...")

	healthy, err := service.CheckHealth()

	if !healthy {
		log.Printf("❌ Stripe health check failed: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"healthy": false,
			"error":   err.Error(),
		})
		return
	}

	log.Println("✅ Stripe is healthy")
	c.JSON(http.StatusOK, gin.H{
		"healthy": true,
		"status":  "ok",
	})
}
