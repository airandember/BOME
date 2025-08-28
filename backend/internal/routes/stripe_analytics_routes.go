package routes

import (
	"fmt"
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

	// Fetch all analytics data in parallel for maximum speed
	type result struct {
		name string
		data interface{}
		err  error
	}

	results := make(chan result, 5)

	// Launch all analytics calls in parallel
	go func() {
		fmt.Printf("🔄 Starting balance fetch...\n")
		if balance, err := stripeService.GetAccountBalance(); err == nil {
			fmt.Printf("✅ Balance fetch completed\n")
			results <- result{"balance", balance, nil}
		} else {
			fmt.Printf("❌ Balance fetch failed: %v\n", err)
			results <- result{"balance", nil, err}
		}
	}()

	go func() {
		fmt.Printf("🔄 Starting charges fetch...\n")
		if charges, err := stripeService.GetChargeCounts(); err == nil {
			fmt.Printf("✅ Charges fetch completed\n")
			results <- result{"charges", charges, nil}
		} else {
			fmt.Printf("❌ Charges fetch failed: %v\n", err)
			results <- result{"charges", nil, err}
		}
	}()

	go func() {
		fmt.Printf("🔄 Starting customers fetch...\n")
		if customers, err := stripeService.GetCustomerCounts(); err == nil {
			fmt.Printf("✅ Customers fetch completed\n")
			results <- result{"customers", customers, nil}
		} else {
			fmt.Printf("❌ Customers fetch failed: %v\n", err)
			results <- result{"customers", nil, err}
		}
	}()

	go func() {
		fmt.Printf("🔄 Starting subscriptions fetch...\n")
		if subscriptions, err := stripeService.GetSubscriptionCounts(); err == nil {
			fmt.Printf("✅ Subscriptions fetch completed\n")
			results <- result{"subscriptions", subscriptions, nil}
		} else {
			fmt.Printf("❌ Subscriptions fetch failed: %v\n", err)
			results <- result{"subscriptions", nil, err}
		}
	}()

	go func() {
		fmt.Printf("🔄 Starting products fetch...\n")
		if products, err := stripeService.GetProductCounts(); err == nil {
			fmt.Printf("✅ Products fetch completed\n")
			results <- result{"products", products, nil}
		} else {
			fmt.Printf("❌ Products fetch failed: %v\n", err)
			results <- result{"products", nil, err}
		}
	}()

	// Collect all results
	dashboardData := gin.H{
		"enabled":   true,
		"timestamp": time.Now().Unix(),
		"performance": gin.H{
			"start_time": startTime.Unix(),
		},
	}

	// Wait for all results (with timeout)
	timeout := time.After(5 * time.Second)
	received := 0

	for received < 5 {
		select {
		case result := <-results:
			received++
			if result.err == nil {
				dashboardData[result.name] = result.data
			} else {
				dashboardData[result.name] = gin.H{"error": result.err.Error()}
			}
		case <-timeout:
			// Timeout reached, return partial data
			dashboardData["timeout_warning"] = "Some data may be incomplete due to timeout"
			fmt.Printf("⚠️ Timeout reached after %v - only %d/%d endpoints completed\n", time.Since(startTime), received, 5)
			goto timeout_exit
		}
	}

timeout_exit:

	duration := time.Since(startTime)
	dashboardData["performance"].(gin.H)["duration_ms"] = duration.Milliseconds()
	dashboardData["performance"].(gin.H)["end_time"] = time.Now().Unix()

	// Log performance metrics
	fmt.Printf("🚀 /stripe/dash completed in %v - %d endpoints processed\n", duration, received)

	c.JSON(http.StatusOK, dashboardData)
}
