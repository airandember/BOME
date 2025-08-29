package routes

import (
	"log"
	"net/http"
	"time"

	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// RegisterStripeAnalyticsRoutes registers the analytics-specific Stripe routes
func RegisterStripeAnalyticsRoutes(router *gin.RouterGroup, stripeService *services.StripeService) {
	stripe := router.Group("/stripe")
	{
		// 🚀 DASH: Lightning-fast dashboard endpoint (double entendre!)
		stripe.GET("/dash", func(c *gin.Context) { getDashboardData(c, stripeService) })

		// Individual analytics endpoints (for detailed views)
		stripe.GET("/balance", func(c *gin.Context) { getAccountBalance(c, stripeService) })
		stripe.GET("/charges", func(c *gin.Context) { getChargeCounts(c, stripeService) })
		stripe.GET("/customers", func(c *gin.Context) { getCustomerCounts(c, stripeService) })
		stripe.GET("/subscriptions", func(c *gin.Context) { getSubscriptionCounts(c, stripeService) })
		stripe.GET("/products", func(c *gin.Context) { getProductCounts(c, stripeService) })
		stripe.GET("/analytics", func(c *gin.Context) { getComprehensiveAnalytics(c, stripeService) })
		stripe.GET("/v2/analytics", func(c *gin.Context) { getV2Analytics(c, stripeService) })
	}
}

// getAccountBalance returns account balance and transaction history
func getAccountBalance(c *gin.Context, stripeService *services.StripeService) {
	balanceResult, err := stripeService.GetAccountBalance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get balance: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"available":         balanceResult.Available,
		"pending":           balanceResult.Pending,
		"instant_available": balanceResult.InstantAvailable,
		"last_updated":      time.Now().Unix(),
	})
}

// getChargeCounts returns total charge counts and summaries
func getChargeCounts(c *gin.Context, stripeService *services.StripeService) {
	counts, err := stripeService.GetChargeCounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get charge counts: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, counts)
}

// getCustomerCounts returns total customer counts and metrics
func getCustomerCounts(c *gin.Context, stripeService *services.StripeService) {
	counts, err := stripeService.GetCustomerCounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get customer counts: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, counts)
}

// getSubscriptionCounts returns active/inactive subscription counts
func getSubscriptionCounts(c *gin.Context, stripeService *services.StripeService) {
	counts, err := stripeService.GetSubscriptionCounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription counts: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, counts)
}

// getProductCounts returns product counts and revenue metrics
func getProductCounts(c *gin.Context, stripeService *services.StripeService) {
	counts, err := stripeService.GetProductCounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product counts: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, counts)
}

// 🚀 getDashboardData returns lightning-fast aggregated dashboard data
func getDashboardData(c *gin.Context, stripeService *services.StripeService) {
	startTime := time.Now()

	// Check if Stripe is enabled first
	if !stripeService.IsEnabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Stripe service is not enabled",
			"enabled": false,
		})
		return
	}

	// Use the comprehensive analytics instead of basic counts
	analytics, err := stripeService.GetComprehensiveAnalytics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get comprehensive analytics: " + err.Error()})
		return
	}

	// Add the enabled flag that frontend expects
	analytics["enabled"] = true

	duration := time.Since(startTime)
	log.Printf("🚀 /stripe/dash completed in %v - comprehensive analytics", duration)

	c.JSON(http.StatusOK, analytics)
}

// getComprehensiveAnalytics returns comprehensive analytics data
func getComprehensiveAnalytics(c *gin.Context, stripeService *services.StripeService) {
	analytics, err := stripeService.GetComprehensiveAnalytics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get comprehensive analytics: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, analytics)
}

// getV2Analytics returns analytics using Stripe API v2 principles
func getV2Analytics(c *gin.Context, stripeService *services.StripeService) {
	startTime := time.Now()

	if !stripeService.IsEnabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Stripe service is not enabled",
			"enabled": false,
		})
		return
	}

	// Use the new v2 analytics method
	analytics, err := stripeService.GetStripeAnalyticsV2()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get v2 analytics: " + err.Error()})
		return
	}

	analytics["enabled"] = true

	duration := time.Since(startTime)
	log.Printf("🚀 /stripe/v2/analytics completed in %v - API v2 approach", duration)

	c.JSON(http.StatusOK, analytics)
}
