package routes

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"bome-backend/internal/database"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// RegisterStripeAnalyticsRoutes registers the analytics-specific Stripe routes
func RegisterStripeAnalyticsRoutes(router *gin.RouterGroup, stripeService *services.StripeService, db *database.DB, syncService *services.StripeSyncService) {
	stripe := router.Group("/stripe")
	{
		// 🚀 DASH: Lightning-fast dashboard endpoint (double entendre!)
		stripe.GET("/dash", func(c *gin.Context) { getDashboardData(c, stripeService) })

		// 🏥 HEALTH: Quick health check for Stripe connectivity
		stripe.GET("/health", func(c *gin.Context) { getStripeHealth(c, stripeService) })

		// 🔥 PING: Super simple endpoint to test if backend is reachable (no Stripe dependency)
		stripe.GET("/ping", func(c *gin.Context) {
			log.Printf("🏓 [PING] Received ping request from IP: %s", c.ClientIP())
			c.JSON(http.StatusOK, gin.H{
				"status":    "ok",
				"timestamp": time.Now().Unix(),
				"message":   "Backend is reachable",
				"endpoint":  "/stripe/ping",
			})
			log.Printf("🏓 [PING] Ping response sent successfully")
		})

		// 🎯 STATUS: Explicit Stripe configuration status (bulletproof)
		stripe.GET("/status", func(c *gin.Context) {
			log.Printf("🔍 [STATUS] Received status request from IP: %s", c.ClientIP())

			if stripeService == nil {
				log.Printf("❌ [STATUS] Stripe service is nil")
				c.JSON(http.StatusOK, gin.H{
					"configured": false,
					"enabled":    false,
					"reason":     "service_nil",
					"message":    "Stripe service not initialized",
				})
				return
			}

			enabled := stripeService.IsEnabled()
			log.Printf("✅ [STATUS] Stripe enabled check: %v", enabled)

			c.JSON(http.StatusOK, gin.H{
				"configured": enabled,
				"enabled":    enabled,
				"reason":     "checked_service",
				"message":    "Status checked successfully",
			})
			log.Printf("📤 [STATUS] Status response sent: configured=%v", enabled)
		})

		// Individual analytics endpoints (for detailed views)
		stripe.GET("/balance", func(c *gin.Context) { getAccountBalance(c, stripeService) })
		stripe.GET("/charges", func(c *gin.Context) { getChargeCounts(c, stripeService) })
		stripe.GET("/customers", func(c *gin.Context) { getCustomerCounts(c, stripeService) })
		stripe.GET("/subscriptions", func(c *gin.Context) { getSubscriptionCounts(c, stripeService) })
		stripe.GET("/products", func(c *gin.Context) { getProductCounts(c, stripeService) })
		stripe.GET("/analytics", func(c *gin.Context) { getComprehensiveAnalytics(c, stripeService) })
		stripe.GET("/v2/analytics", func(c *gin.Context) { getV2Analytics(c, stripeService) })

		// Database stats endpoints
		stripe.GET("/database/stats", func(c *gin.Context) { getDatabaseStats(c, db) })
		stripe.GET("/database/customers", func(c *gin.Context) { getDatabaseCustomers(c, db) })
		stripe.GET("/database/subscriptions", func(c *gin.Context) { getDatabaseSubscriptions(c, db) })

		// Metadata health endpoints
		stripe.GET("/metadata/health", func(c *gin.Context) { getMetadataHealth(c, db) })
		stripe.POST("/metadata/fix", func(c *gin.Context) { fixMetadataCorruption(c, db) })

		// Stripe products management endpoints
		stripe.GET("/products/available", func(c *gin.Context) { getAvailableStripeProducts(c, db) })
		stripe.GET("/products/all", func(c *gin.Context) { getAllStripeProducts(c, db) })
		stripe.GET("/products/debug", func(c *gin.Context) { debugStripeProductsData(c, db) })
		stripe.PUT("/products/:stripe_id/availability", func(c *gin.Context) { updateStripeProductAvailability(c, db) })
		stripe.PUT("/products/bulk-availability", func(c *gin.Context) { bulkUpdateStripeProductAvailability(c, db) })
		stripe.POST("/products/import-as-plans", func(c *gin.Context) { importStripeProductsAsPlans(c, db) })

		// Manual UI-triggered sync endpoints (for frontend users)
		stripe.POST("/sync/trigger", func(c *gin.Context) { triggerManualSync(c, syncService) })
	}
}

// getStripeHealth returns a quick health check for Stripe connectivity
func getStripeHealth(c *gin.Context, stripeService *services.StripeService) {
	startTime := time.Now()

	// Check if Stripe is enabled first
	if !stripeService.IsEnabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"healthy": false,
			"error":   "Stripe service is not enabled",
			"enabled": false,
		})
		return
	}

	// Quick 5-second timeout for health check
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Use channel for timeout protection
	type healthResult struct {
		healthy bool
		error   error
	}

	resultChan := make(chan healthResult, 1)

	// Run quick health check in goroutine
	go func() {
		// Just try to get account balance (fastest Stripe API call)
		_, err := stripeService.GetAccountBalance()
		resultChan <- healthResult{err == nil, err}
	}()

	// Wait for result or timeout
	select {
	case res := <-resultChan:
		duration := time.Since(startTime)

		if res.healthy {
			log.Printf("✅ Stripe health check passed in %v", duration)
			c.JSON(http.StatusOK, gin.H{
				"healthy":  true,
				"enabled":  true,
				"duration": duration.String(),
				"status":   "ok",
			})
		} else {
			log.Printf("❌ Stripe health check failed in %v: %v", duration, res.error)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"healthy":  false,
				"enabled":  true,
				"duration": duration.String(),
				"error":    res.error.Error(),
				"status":   "error",
			})
		}

	case <-ctx.Done():
		duration := time.Since(startTime)
		log.Printf("⏰ Stripe health check TIMEOUT after %v", duration)

		c.JSON(http.StatusRequestTimeout, gin.H{
			"healthy":  false,
			"enabled":  true,
			"duration": duration.String(),
			"error":    "Health check timed out",
			"status":   "timeout",
		})
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

// 🚀 getDashboardData returns lightning-fast aggregated dashboard data with timeout protection
func getDashboardData(c *gin.Context, stripeService *services.StripeService) {
	startTime := time.Now()
	log.Printf("🚀 [DASH-START] Dashboard request initiated at %v from IP: %s", startTime, c.ClientIP())

	// IMMEDIATE RESPONSE: Check if service is nil first (defensive programming)
	if stripeService == nil {
		log.Printf("❌ [DASH-ERROR] Stripe service is nil")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Stripe service is not initialized",
			"enabled": false,
			"debug":   "service_nil",
		})
		return
	}
	log.Printf("✅ [DASH-SERVICE] Stripe service object exists")

	// IMMEDIATE RESPONSE: Check if Stripe is enabled (with timeout protection)
	enabledCheckStart := time.Now()
	enabled := stripeService.IsEnabled()
	enabledCheckDuration := time.Since(enabledCheckStart)
	log.Printf("⏱️ [DASH-ENABLED-CHECK] IsEnabled() took %v, result: %v", enabledCheckDuration, enabled)

	if !enabled {
		log.Printf("❌ [DASH-ERROR] Stripe service is not enabled")
		responseData := gin.H{
			"error":   "Stripe service is not enabled",
			"enabled": false,
			"debug":   "service_disabled",
		}
		log.Printf("📤 [DASH-RESPONSE] Sending 503 Service Unavailable with data: %+v", responseData)
		c.JSON(http.StatusServiceUnavailable, responseData)
		log.Printf("✅ [DASH-RESPONSE] 503 response sent successfully")
		return
	}
	log.Printf("✅ [DASH-ENABLED] Stripe service is enabled, proceeding with analytics")

	// 🔥 TIMEOUT PROTECTION: Create context with 30-second timeout for production (more aggressive)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	log.Printf("⏰ [DASH-TIMEOUT] Context timeout set to 30 seconds")

	// Use channel to handle timeout gracefully
	type result struct {
		analytics map[string]interface{}
		err       error
	}

	resultChan := make(chan result, 1)

	// Run analytics in goroutine with timeout protection
	go func() {
		log.Printf("🔄 [DASH-GOROUTINE] Starting comprehensive analytics fetch")
		analytics, err := stripeService.GetComprehensiveAnalyticsWithContext(ctx)
		if err != nil {
			log.Printf("❌ [DASH-GOROUTINE] Analytics fetch failed: %v", err)
		} else {
			log.Printf("✅ [DASH-GOROUTINE] Analytics fetch completed successfully")
		}
		resultChan <- result{analytics, err}
		log.Printf("📤 [DASH-GOROUTINE] Result sent to channel")
	}()

	// Wait for result or timeout
	log.Printf("⏳ [DASH-SELECT] Waiting for analytics result or timeout...")
	select {
	case res := <-resultChan:
		log.Printf("📥 [DASH-SELECT] Received result from channel")
		if res.err != nil {
			log.Printf("❌ [DASH-ERROR] Analytics failed: %v", res.err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":              "Failed to get comprehensive analytics: " + res.err.Error(),
				"timeout_protection": true,
			})
			return
		}

		// Add the enabled flag that frontend expects
		res.analytics["enabled"] = true
		log.Printf("✅ [DASH-SUCCESS] Analytics data prepared, adding enabled flag")

		duration := time.Since(startTime)
		log.Printf("🚀 [DASH-COMPLETE] /stripe/dash completed in %v - comprehensive analytics", duration)

		c.JSON(http.StatusOK, res.analytics)
		log.Printf("📤 [DASH-RESPONSE] JSON response sent to client")

	case <-ctx.Done():
		duration := time.Since(startTime)
		log.Printf("⏰ [DASH-TIMEOUT] Context timeout triggered after %v", duration)
		log.Printf("🔄 [DASH-TIMEOUT] Preparing fallback data to avoid 504 error")

		// Return minimal fallback data instead of 504 error
		fallbackData := map[string]interface{}{
			"enabled":          true,
			"error":            "Analytics request timed out - using fallback data",
			"timeout":          true,
			"timeout_duration": duration.String(),
			"method":           "fallback_timeout_protection",
			"timestamp":        time.Now().Unix(),
			"subscription_metrics": map[string]interface{}{
				"error":  "Timed out after " + duration.String(),
				"status": "timeout",
			},
			"customer_analytics": map[string]interface{}{
				"error":  "Timed out after " + duration.String(),
				"status": "timeout",
			},
			"revenue_analytics": map[string]interface{}{
				"error":  "Timed out after " + duration.String(),
				"status": "timeout",
			},
		}

		log.Printf("📤 [DASH-TIMEOUT] Sending fallback response (200 OK) to prevent 504")
		c.JSON(http.StatusOK, fallbackData)
		log.Printf("✅ [DASH-TIMEOUT] Fallback response sent successfully")
	}
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

// getDatabaseStats returns real counts from database tables
func getDatabaseStats(c *gin.Context, db *database.DB) {
	stats := make(map[string]interface{})

	// Get customer count from stripe_customers table
	var customerCount int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM stripe_customers").Scan(&customerCount)
	if err != nil {
		log.Printf("Error getting customer count: %v", err)
		customerCount = 0
	}

	// Get active subscription count from stripe_subscriptions table
	var subscriptionCount int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM stripe_subscriptions WHERE status = 'active'").Scan(&subscriptionCount)
	if err != nil {
		log.Printf("Error getting active subscription count: %v", err)
		subscriptionCount = 0
	}

	// Get product count from stripe_products table
	var productCount int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM stripe_products").Scan(&productCount)
	if err != nil {
		log.Printf("Error getting product count: %v", err)
		productCount = 0
	}

	// Get invoice count from stripe_invoices table
	var invoiceCount int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM stripe_invoices").Scan(&invoiceCount)
	if err != nil {
		log.Printf("Error getting invoice count: %v", err)
		invoiceCount = 0
	}

	stats["customers"] = customerCount
	stats["subscriptions"] = subscriptionCount
	stats["products"] = productCount
	stats["invoices"] = invoiceCount
	stats["last_updated"] = time.Now().Unix()
	stats["source"] = "database"

	log.Printf("📊 Database stats: %d customers, %d active subscriptions, %d products, %d invoices",
		customerCount, subscriptionCount, productCount, invoiceCount)

	c.JSON(http.StatusOK, stats)
}

// getDatabaseCustomers returns customers from stripe_customers table with optional subscriptions
func getDatabaseCustomers(c *gin.Context, db *database.DB) {
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

	var customers []map[string]interface{}
	var totalCount int

	// Get total count
	err = db.DB.QueryRow("SELECT COUNT(*) FROM stripe_customers").Scan(&totalCount)
	if err != nil {
		log.Printf("Error getting customer count: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get customer count"})
		return
	}

	// Base query for customers
	baseQuery := `
		SELECT 
			stripe_id, name, email, created_at, updated_at, metadata
		FROM stripe_customers 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`

	rows, err := db.DB.Query(baseQuery, limit, offset)
	if err != nil {
		log.Printf("Error querying customers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query customers"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var customer map[string]interface{} = make(map[string]interface{})
		var stripeID, name, email, metadata sql.NullString
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(
			&stripeID, &name, &email, &createdAt, &updatedAt, &metadata,
		)
		if err != nil {
			log.Printf("Error scanning customer row: %v", err)
			continue
		}

		customer["stripe_id"] = stripeID.String
		customer["name"] = name.String
		customer["email"] = email.String
		customer["created_at"] = createdAt.Time
		customer["updated_at"] = updatedAt.Time
		customer["metadata"] = metadata.String

		// If requested, get subscriptions for this customer
		if includeSubscriptions {
			subscriptions, err := getCustomerSubscriptions(db, stripeID.String)
			if err != nil {
				log.Printf("Error getting subscriptions for customer %s: %v", stripeID.String, err)
				customer["subscriptions"] = []map[string]interface{}{}
			} else {
				customer["subscriptions"] = subscriptions
			}
		}

		customers = append(customers, customer)
	}

	log.Printf("📊 Retrieved %d customers from database (offset: %d, limit: %d, total: %d)",
		len(customers), offset, limit, totalCount)

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

// getDatabaseSubscriptions returns subscriptions from stripe_subscriptions table
func getDatabaseSubscriptions(c *gin.Context, db *database.DB) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")
	status := c.Query("status") // active, canceled, trialing, etc.

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

	var subscriptions []map[string]interface{}
	var totalCount int

	// Build query with optional status filter
	countQuery := "SELECT COUNT(*) FROM stripe_subscriptions"
	baseQuery := `
		SELECT 
			stripe_id, customer_id, status, current_period_start, current_period_end,
			created_at
		FROM stripe_subscriptions
	`

	var args []interface{}
	argIndex := 1

	if status != "" {
		countQuery += " WHERE status = $" + strconv.Itoa(argIndex)
		baseQuery += " WHERE status = $" + strconv.Itoa(argIndex)
		args = append(args, status)
		argIndex++
	}

	baseQuery += " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(argIndex) + " OFFSET $" + strconv.Itoa(argIndex+1)
	args = append(args, limit, offset)

	// Get total count
	var countArgs []interface{}
	if status != "" {
		countArgs = []interface{}{status}
	}
	err = db.DB.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		log.Printf("Error getting subscription count: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription count"})
		return
	}

	rows, err := db.DB.Query(baseQuery, args...)
	if err != nil {
		log.Printf("Error querying subscriptions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query subscriptions"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var subscription map[string]interface{} = make(map[string]interface{})
		var stripeID, customerID, status sql.NullString
		var currentPeriodStart, currentPeriodEnd, createdAt sql.NullTime

		err := rows.Scan(
			&stripeID, &customerID, &status, &currentPeriodStart, &currentPeriodEnd,
			&createdAt,
		)
		if err != nil {
			log.Printf("Error scanning subscription row: %v", err)
			continue
		}

		subscription["stripe_id"] = stripeID.String
		subscription["customer_id"] = customerID.String
		subscription["status"] = status.String
		subscription["current_period_start"] = currentPeriodStart.Time
		subscription["current_period_end"] = currentPeriodEnd.Time
		subscription["created_at"] = createdAt.Time

		subscriptions = append(subscriptions, subscription)
	}

	log.Printf("📊 Retrieved %d subscriptions from database (offset: %d, limit: %d, total: %d, status: %s)",
		len(subscriptions), offset, limit, totalCount, status)

	c.JSON(http.StatusOK, gin.H{
		"subscriptions": subscriptions,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"total":  totalCount,
		},
		"filters": gin.H{
			"status": status,
		},
		"source": "database",
	})
}

// Helper function to get subscriptions for a specific customer (by Stripe customer ID)
func getCustomerSubscriptions(db *database.DB, stripeCustomerID string) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			s.stripe_id, s.status, s.current_period_start, s.current_period_end,
			s.created_at
		FROM stripe_subscriptions s
		INNER JOIN stripe_customers c ON s.customer_id = c.id
		WHERE c.stripe_id = $1 
		ORDER BY s.created_at DESC
	`

	rows, err := db.DB.Query(query, stripeCustomerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []map[string]interface{}
	for rows.Next() {
		var subscription map[string]interface{} = make(map[string]interface{})
		var stripeID, status sql.NullString
		var currentPeriodStart, currentPeriodEnd, createdAt sql.NullTime

		err := rows.Scan(
			&stripeID, &status, &currentPeriodStart, &currentPeriodEnd,
			&createdAt,
		)
		if err != nil {
			continue
		}

		subscription["stripe_id"] = stripeID.String
		subscription["status"] = status.String
		subscription["current_period_start"] = currentPeriodStart.Time
		subscription["current_period_end"] = currentPeriodEnd.Time
		subscription["created_at"] = createdAt.Time

		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, nil
}

// triggerManualSync triggers a manual sync from the UI/frontend
// This endpoint is specifically for user-initiated syncs with UI feedback
var (
	syncMutexMap   = make(map[string]*sync.Mutex)
	syncMapMutex   = &sync.Mutex{}
	lastSyncTime   = make(map[string]int64)
	requestCounter = make(map[string]int)

	// 🌐 NETWORK DEBUGGING: Track request deduplication
	seenRequests     = make(map[string][]time.Time) // Frontend Request ID -> timestamps
	duplicateTracker = make(map[string]int)         // Frontend Request ID -> count
)

func getSyncMutex(syncType string) *sync.Mutex {
	syncMapMutex.Lock()
	defer syncMapMutex.Unlock()

	if _, exists := syncMutexMap[syncType]; !exists {
		syncMutexMap[syncType] = &sync.Mutex{}
	}
	return syncMutexMap[syncType]
}

func triggerManualSync(c *gin.Context, syncService *services.StripeSyncService) {
	ctx := context.Background()
	//requestStartTime := time.Now()

	// 🔍 COMPREHENSIVE NETWORK DEBUGGING (COMMENTED OUT FOR PRODUCTION)
	// log.Printf("🌐 [NETWORK-DEBUG] ==================== NEW REQUEST ====================")
	// log.Printf("🌐 [NETWORK-DEBUG] Timestamp: %s (nano: %d)", requestStartTime.Format("15:04:05.000000000"), requestStartTime.UnixNano())

	// Get sync type from query parameter (default to customers)
	syncType := c.DefaultQuery("type", "customers")

	// Add unique request ID for debugging
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	// Also capture frontend request ID if provided
	frontendRequestID := c.GetHeader("X-Frontend-Request-ID")
	// Keep essential duplicate tracking but comment out verbose logging
	// if frontendRequestID != "" {
	// 	log.Printf("🔍 [UI-MANUAL] Frontend Request ID: %s", frontendRequestID)

	// 	// 🌐 TRACK REQUEST DUPLICATION
	// 	syncMapMutex.Lock()
	// 	if _, exists := seenRequests[frontendRequestID]; !exists {
	// 		seenRequests[frontendRequestID] = []time.Time{}
	// 		duplicateTracker[frontendRequestID] = 0
	// 	}
	// 	seenRequests[frontendRequestID] = append(seenRequests[frontendRequestID], requestStartTime)
	// 	duplicateTracker[frontendRequestID]++
	// 	duplicateCount := duplicateTracker[frontendRequestID]
	// 	allTimestamps := seenRequests[frontendRequestID]
	// 	syncMapMutex.Unlock()

	// 	log.Printf("🌐 [NETWORK-DEBUG] DUPLICATE ANALYSIS: Frontend ID %s seen %d times", frontendRequestID, duplicateCount)
	// 	if duplicateCount > 1 {
	// 		log.Printf("🚨 [NETWORK-DEBUG] DUPLICATE DETECTED! Frontend ID %s has been seen %d times:", frontendRequestID, duplicateCount)
	// 		for i, timestamp := range allTimestamps {
	// 			log.Printf("🚨 [NETWORK-DEBUG]   Request #%d: %s (nano: %d)", i+1, timestamp.Format("15:04:05.000000000"), timestamp.UnixNano())
	// 			if i > 0 {
	// 				timeDiff := timestamp.Sub(allTimestamps[i-1])
	// 				log.Printf("🚨 [NETWORK-DEBUG]   Time diff from previous: %v", timeDiff)
	// 			}
	// 		}
	// 	}
	// }

	// 🌐 DETAILED NETWORK INFORMATION (COMMENTED OUT FOR PRODUCTION)
	// log.Printf("🌐 [NETWORK-DEBUG] Request Method: %s", c.Request.Method)
	// log.Printf("🌐 [NETWORK-DEBUG] Request URL: %s", c.Request.URL.String())
	// log.Printf("🌐 [NETWORK-DEBUG] Request URI: %s", c.Request.RequestURI)
	// log.Printf("🌐 [NETWORK-DEBUG] Remote Address: %s", c.Request.RemoteAddr)
	// log.Printf("🌐 [NETWORK-DEBUG] Host: %s", c.Request.Host)
	// log.Printf("🌐 [NETWORK-DEBUG] Protocol: %s", c.Request.Proto)
	// log.Printf("🌐 [NETWORK-DEBUG] Content Length: %d", c.Request.ContentLength)
	// log.Printf("🌐 [NETWORK-DEBUG] Transfer Encoding: %v", c.Request.TransferEncoding)

	// // 🌐 ALL REQUEST HEADERS (COMMENTED OUT FOR PRODUCTION)
	// log.Printf("🌐 [NETWORK-DEBUG] === ALL REQUEST HEADERS ===")
	// for headerName, headerValues := range c.Request.Header {
	// 	for _, headerValue := range headerValues {
	// 		log.Printf("🌐 [NETWORK-DEBUG] Header: %s = %s", headerName, headerValue)
	// 	}
	// }

	// // 🌐 CONNECTION DETAILS (COMMENTED OUT FOR PRODUCTION)
	// if c.Request.TLS != nil {
	// 	log.Printf("🌐 [NETWORK-DEBUG] TLS Version: %x", c.Request.TLS.Version)
	// 	log.Printf("🌐 [NETWORK-DEBUG] TLS Cipher Suite: %x", c.Request.TLS.CipherSuite)
	// } else {
	// 	log.Printf("🌐 [NETWORK-DEBUG] TLS: Not used (HTTP)")
	// }

	// Increment request counter for debugging
	syncMapMutex.Lock()
	requestCounter[syncType]++
	currentCount := requestCounter[syncType]
	syncMapMutex.Unlock()

	// Keep essential logging but reduce verbosity
	log.Printf("🔍 [UI-MANUAL] Manual sync request: %s (count: %d)", syncType, currentCount)
	// log.Printf("🔍 [UI-MANUAL] Request ID: %s, Sync Type: %s, Request Count: %d", requestID, syncType, currentCount)
	// log.Printf("🔍 [UI-MANUAL] EVERY REQUEST LOGGED - Previous count was: %d, Now: %d", currentCount-1, currentCount)
	// log.Printf("🔍 [UI-MANUAL] TIMING: Request received at %s (nanoseconds: %d)", requestStartTime.Format("15:04:05.000000000"), requestStartTime.UnixNano())
	// log.Printf("🔍 [UI-MANUAL] CONNECTION: RemoteAddr=%s, Host=%s, Proto=%s", c.Request.RemoteAddr, c.Request.Host, c.Request.Proto)

	// Get the mutex for this sync type
	syncMutex := getSyncMutex(syncType)

	// Try to acquire the lock (non-blocking)
	if !syncMutex.TryLock() {
		log.Printf("🚫 Sync for %s already in progress, skipping...", syncType)
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Sync already in progress",
			"type":    syncType,
			"status":  "in_progress",
			"message": "Another sync of this type is already running",
		})
		return
	}

	log.Printf("🔒 Acquired mutex lock for sync type: %s", syncType)

	// Ensure we unlock when done
	defer func() {
		syncMutex.Unlock()
		log.Printf("🔓 Released mutex lock for sync type: %s", syncType)
	}()

	// Prevent double execution within 10 minutes (600 seconds) - longer than typical sync duration
	now := time.Now().Unix()
	syncMapMutex.Lock()
	if lastTime, exists := lastSyncTime[syncType]; exists && (now-lastTime) < 600 {
		syncMapMutex.Unlock()
		timeSinceLastSync := now - lastTime
		log.Printf("🚫 Sync for %s triggered too recently (%d seconds ago), skipping...", syncType, timeSinceLastSync)
		log.Printf("🚫 Rate limit: Must wait at least 600 seconds between syncs, only %d seconds have passed", timeSinceLastSync)
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":              "Sync triggered too recently",
			"type":               syncType,
			"status":             "rate_limited",
			"message":            "Please wait at least 10 minutes between sync requests",
			"seconds_since_last": timeSinceLastSync,
			"seconds_required":   600,
		})
		return
	}
	lastSyncTime[syncType] = now
	syncMapMutex.Unlock()

	log.Printf("🚀 [UI-MANUAL] Manual sync triggered for: %s (timestamp: %d)", syncType, now)
	log.Printf("🔍 [UI-MANUAL] Request details: Method=%s, URL=%s, RemoteAddr=%s, UserAgent=%s",
		c.Request.Method, c.Request.URL.String(), c.Request.RemoteAddr, c.Request.UserAgent())

	var err error
	switch syncType {
	case "customers":
		err = syncService.TestCustomerSyncUnlimited(ctx)
	case "initial":
		err = syncService.InitialDataSync(ctx)
	case "coupons":
		err = syncService.SyncCouponsManual(ctx)
	case "monthly_metrics":
		err = syncService.SyncMonthlyMetricsManual(ctx)
	case "products":
		err = syncService.SyncProductsManual(ctx)
	case "prices":
		err = syncService.SyncPricesManual(ctx)
	case "subscriptions":
		err = syncService.SyncSubscriptionsManual(ctx)
	case "cleanup_orphaned":
		err = syncService.CleanupOrphanedSubscriptions(ctx)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":           "Invalid sync type. Use 'customers', 'initial', 'coupons', 'monthly_metrics', 'products', 'prices', 'subscriptions', or 'cleanup_orphaned'",
			"available_types": []string{"customers", "initial", "coupons", "monthly_metrics", "products", "prices", "subscriptions", "cleanup_orphaned"},
		})
		return
	}

	if err != nil {
		log.Printf("❌ Manual sync failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Sync failed: " + err.Error(),
			"type":   syncType,
			"status": "failed",
		})
		return
	}

	log.Printf("✅ Manual sync completed successfully: %s", syncType)

	// 🌐 NETWORK DEBUG: Response tracking (COMMENTED OUT FOR PRODUCTION)
	// responseTime := time.Since(requestStartTime)
	// log.Printf("🌐 [NETWORK-DEBUG] ==================== RESPONSE SENT ====================")
	// log.Printf("🌐 [NETWORK-DEBUG] Response Status: 200 OK")
	// log.Printf("🌐 [NETWORK-DEBUG] Total Request Duration: %v", responseTime)
	// log.Printf("🌐 [NETWORK-DEBUG] Frontend Request ID: %s", frontendRequestID)
	// log.Printf("🌐 [NETWORK-DEBUG] Backend Request ID: %s", requestID)

	c.JSON(http.StatusOK, gin.H{
		"message":             "Manual sync completed successfully",
		"type":                syncType,
		"status":              "success",
		"timestamp":           time.Now().Unix(),
		"frontend_request_id": frontendRequestID,
		"backend_request_id":  requestID,
		//"request_duration":    responseTime.String(),
	})
}

// getAvailableStripeProducts returns all Stripe products marked as available with pricing information
func getAvailableStripeProducts(c *gin.Context, db *database.DB) {
	log.Printf("📦 Getting available Stripe products with pricing information...")

	query := `
		SELECT id, stripe_id, name, description, active, available, created_at
		FROM stripe_products
		WHERE available = true
		ORDER BY name ASC
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		log.Printf("❌ Error querying available Stripe products: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch available Stripe products",
		})
		return
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var id int
		var stripeID, name string
		var description *string
		var active, available bool
		var createdAt time.Time

		err := rows.Scan(&id, &stripeID, &name, &description, &active, &available, &createdAt)
		if err != nil {
			log.Printf("❌ Error scanning Stripe product: %v", err)
			continue
		}

		product := map[string]interface{}{
			"id":          id,
			"stripe_id":   stripeID,
			"name":        name,
			"description": description,
			"active":      active,
			"available":   available,
			"created_at":  createdAt.Format(time.RFC3339),
			"price":       nil, // Will be populated below
		}
		products = append(products, product)
	}

	// Now get pricing information for each product
	for i, product := range products {
		productID := product["id"].(int)

		// Get comprehensive price information for this product
		priceQuery := `
			SELECT stripe_id, unit_amount, currency, recurring_interval
			FROM stripe_prices 
			WHERE product_id = $1 
			ORDER BY created_at DESC 
			LIMIT 1
		`

		var priceStripeID, currency, recurringInterval *string
		var unitAmount *int64
		err := db.DB.QueryRow(priceQuery, productID).Scan(&priceStripeID, &unitAmount, &currency, &recurringInterval)
		if err == nil && unitAmount != nil {
			products[i]["price"] = *unitAmount
			if priceStripeID != nil {
				products[i]["price_id"] = *priceStripeID
			}
			if currency != nil {
				products[i]["currency"] = *currency
			}
			if recurringInterval != nil {
				products[i]["recurring_interval"] = *recurringInterval
			}
		} else if err != nil && err.Error() != "sql: no rows in result set" {
			log.Printf("⚠️ Error getting price for available product ID %d: %v", productID, err)
		}
	}

	log.Printf("✅ Found %d available Stripe products with pricing", len(products))

	c.JSON(http.StatusOK, gin.H{
		"products": products,
		"count":    len(products),
	})
}

// updateStripeProductAvailability updates the availability status of a Stripe product
func updateStripeProductAvailability(c *gin.Context, db *database.DB) {
	stripeID := c.Param("stripe_id")
	if stripeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Stripe product ID is required",
		})
		return
	}

	var requestBody struct {
		Available bool `json:"available"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		log.Printf("❌ Error parsing request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	log.Printf("🔄 Updating Stripe product %s availability to: %v", stripeID, requestBody.Available)

	query := `
		UPDATE stripe_products 
		SET available = $1 
		WHERE stripe_id = $2
	`

	result, err := db.Exec(query, requestBody.Available, stripeID)
	if err != nil {
		log.Printf("❌ Error updating Stripe product availability: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update product availability",
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("❌ Error getting rows affected: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to verify update",
		})
		return
	}

	if rowsAffected == 0 {
		log.Printf("❌ No Stripe product found with ID: %s", stripeID)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Stripe product not found",
		})
		return
	}

	log.Printf("✅ Successfully updated Stripe product %s availability to: %v", stripeID, requestBody.Available)

	c.JSON(http.StatusOK, gin.H{
		"message":   "Product availability updated successfully",
		"stripe_id": stripeID,
		"available": requestBody.Available,
		"timestamp": time.Now().Unix(),
	})
}

// getAllStripeProducts returns all Stripe products (both available and unavailable)
func getAllStripeProducts(c *gin.Context, db *database.DB) {
	log.Printf("📦 Getting all Stripe products...")

	// Add detailed debugging
	if db == nil {
		log.Printf("❌ Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database connection is nil",
		})
		return
	}

	if db.DB == nil {
		log.Printf("❌ Database DB field is nil")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database DB field is nil",
		})
		return
	}

	log.Printf("✅ Database connections are valid, executing query...")

	// First check if the stripe_products table exists
	tableCheckQuery := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'stripe_products'
		);
	`

	var tableExists bool
	err := db.DB.QueryRow(tableCheckQuery).Scan(&tableExists)
	if err != nil {
		log.Printf("❌ Error checking if stripe_products table exists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check table existence",
			"details": err.Error(),
		})
		return
	}

	if !tableExists {
		log.Printf("❌ stripe_products table does not exist")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "stripe_products table does not exist",
		})
		return
	}

	log.Printf("✅ stripe_products table exists, proceeding with query...")

	// Start with a simple query to get just the products first
	query := `
		SELECT id, stripe_id, name, description, active, available, created_at
		FROM stripe_products
		ORDER BY name ASC
	`

	log.Printf("🔍 Executing query: %s", query)
	rows, err := db.DB.Query(query)
	if err != nil {
		log.Printf("❌ Error querying all Stripe products: %v", err)
		log.Printf("❌ Error type: %T", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch Stripe products",
			"details": err.Error(),
		})
		return
	}
	defer rows.Close()
	log.Printf("✅ Query executed successfully, processing rows...")

	var products []map[string]interface{}
	productCount := 0

	for rows.Next() {
		var id int
		var stripeID, name string
		var description *string
		var active, available bool
		var createdAt time.Time

		err := rows.Scan(&id, &stripeID, &name, &description, &active, &available, &createdAt)
		if err != nil {
			log.Printf("❌ Error scanning Stripe product: %v", err)
			log.Printf("❌ Scan error type: %T", err)
			continue
		}

		productCount++
		log.Printf("✅ Scanned product %d: ID=%d, StripeID=%s, Name=%s", productCount, id, stripeID, name)

		product := map[string]interface{}{
			"id":          id,
			"stripe_id":   stripeID,
			"name":        name,
			"description": description,
			"active":      active,
			"available":   available,
			"created_at":  createdAt.Format(time.RFC3339),
			"price":       nil, // We'll add price lookup separately
		}
		products = append(products, product)
	}

	log.Printf("✅ Processed %d products from database", len(products))

	// Now try to get prices for each product
	log.Printf("🔍 Starting price lookup for %d products...", len(products))
	for i, product := range products {
		productID := product["id"].(int) // Use the integer ID, not the stripe_id string
		log.Printf("🔍 Looking up price for product ID %d (stripe_id: %s)", productID, product["stripe_id"])

		// Try to get comprehensive price information for this product
		priceQuery := `
			SELECT stripe_id, unit_amount, currency, recurring_interval
			FROM stripe_prices 
			WHERE product_id = $1 
			ORDER BY created_at DESC 
			LIMIT 1
		`

		var priceStripeID, currency, recurringInterval *string
		var unitAmount *int64
		err := db.DB.QueryRow(priceQuery, productID).Scan(&priceStripeID, &unitAmount, &currency, &recurringInterval)
		if err == nil && unitAmount != nil {
			log.Printf("✅ Found price for product ID %d: $%.2f %s %s", productID, float64(*unitAmount)/100, *currency, *recurringInterval)
			products[i]["price"] = *unitAmount
			if priceStripeID != nil {
				products[i]["price_id"] = *priceStripeID
			}
			if currency != nil {
				products[i]["currency"] = *currency
			}
			if recurringInterval != nil {
				products[i]["recurring_interval"] = *recurringInterval
			}
		} else if err != nil && err.Error() != "sql: no rows in result set" {
			log.Printf("❌ Error getting price for product ID %d: %v", productID, err)
			log.Printf("❌ Price query error type: %T", err)
		} else {
			log.Printf("⚠️ No price found for product ID %d", productID)
		}
		// If no price found or error, price remains nil
	}

	log.Printf("✅ Price lookup completed for all products")

	log.Printf("✅ Found %d total Stripe products", len(products))

	c.JSON(http.StatusOK, gin.H{
		"products": products,
		"count":    len(products),
	})
}

// importStripeProductsAsPlans imports selected Stripe products as subscription plans
func importStripeProductsAsPlans(c *gin.Context, db *database.DB) {
	var request struct {
		StripeProductIDs []string `json:"stripe_product_ids"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if len(request.StripeProductIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No products specified"})
		return
	}

	log.Printf("Importing %d Stripe products as subscription plans", len(request.StripeProductIDs))

	tx, err := db.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	importedCount := 0
	skippedCount := 0

	for _, stripeProductID := range request.StripeProductIDs {
		// Check if plan already exists
		var existingPlanID int
		checkQuery := `SELECT id FROM subscription_plans WHERE stripe_product_id = $1`
		err := tx.QueryRow(checkQuery, stripeProductID).Scan(&existingPlanID)
		if err == nil {
			log.Printf("Plan already exists for Stripe product %s, skipping", stripeProductID)
			skippedCount++
			continue
		} else if err != sql.ErrNoRows {
			log.Printf("Error checking existing plan for %s: %v", stripeProductID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Get Stripe product details
		var product struct {
			ID          int    `json:"id"`
			StripeID    string `json:"stripe_id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Active      bool   `json:"active"`
		}

		productQuery := `
			SELECT id, stripe_id, name, COALESCE(description, '') as description, active 
			FROM stripe_products 
			WHERE stripe_id = $1`

		err = tx.QueryRow(productQuery, stripeProductID).Scan(
			&product.ID, &product.StripeID, &product.Name, &product.Description, &product.Active)
		if err != nil {
			log.Printf("Error fetching Stripe product %s: %v", stripeProductID, err)
			continue
		}

		// Get pricing information
		var price struct {
			StripeID          string `json:"stripe_id"`
			UnitAmount        int    `json:"unit_amount"`
			Currency          string `json:"currency"`
			RecurringInterval string `json:"recurring_interval"`
		}

		priceQuery := `
			SELECT stripe_id, unit_amount, currency, COALESCE(recurring_interval, 'month') as recurring_interval
			FROM stripe_prices 
			WHERE product_id = $1 
			ORDER BY created_at DESC 
			LIMIT 1`

		err = tx.QueryRow(priceQuery, product.ID).Scan(
			&price.StripeID, &price.UnitAmount, &price.Currency, &price.RecurringInterval)
		if err != nil {
			log.Printf("No pricing found for Stripe product %s, using defaults", stripeProductID)
			price.UnitAmount = 0
			price.Currency = "usd"
			price.RecurringInterval = "month"
		}

		// Create subscription plan
		insertQuery := `
			INSERT INTO subscription_plans (
				name, description, short_desc, price, currency, interval, interval_count,
				stripe_price_id, stripe_product_id, features, is_active, sub_type,
				created_at, updated_at, is_deleted
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW(), false
			) RETURNING id`

		var newPlanID int
		err = tx.QueryRow(insertQuery,
			product.Name,                    // name
			product.Description,             // description
			product.Description,             // short_desc (same as description)
			float64(price.UnitAmount)/100.0, // price (convert from cents)
			price.Currency,                  // currency
			price.RecurringInterval,         // interval
			1,                               // interval_count
			sql.NullString{String: price.StripeID, Valid: price.StripeID != ""}, // stripe_price_id
			stripeProductID, // stripe_product_id
			"[]",            // features (empty JSON array)
			product.Active,  // is_active
			"stnd",          // sub_type (standard)
		).Scan(&newPlanID)

		if err != nil {
			log.Printf("Error creating subscription plan for %s: %v", stripeProductID, err)
			continue
		}

		log.Printf("Successfully imported Stripe product %s as subscription plan %d", stripeProductID, newPlanID)
		importedCount++
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save changes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"imported_count": importedCount,
		"skipped_count":  skippedCount,
		"total_count":    len(request.StripeProductIDs),
		"message":        fmt.Sprintf("Successfully imported %d products as subscription plans", importedCount),
	})
}

// bulkUpdateStripeProductAvailability updates availability for multiple products
func bulkUpdateStripeProductAvailability(c *gin.Context, db *database.DB) {
	var requestBody struct {
		Updates []struct {
			StripeID  string `json:"stripe_id"`
			Available bool   `json:"available"`
		} `json:"updates"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		log.Printf("❌ Error parsing bulk update request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	if len(requestBody.Updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No updates provided",
		})
		return
	}

	log.Printf("🔄 Bulk updating availability for %d Stripe products", len(requestBody.Updates))

	// Start a transaction for bulk updates
	tx, err := db.DB.Begin()
	if err != nil {
		log.Printf("❌ Error starting transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start transaction",
		})
		return
	}
	defer tx.Rollback() // Will be ignored if tx.Commit() succeeds

	updatedCount := 0
	failedUpdates := []string{}

	for _, update := range requestBody.Updates {
		query := `UPDATE stripe_products SET available = $1 WHERE stripe_id = $2`
		result, err := tx.Exec(query, update.Available, update.StripeID)
		if err != nil {
			log.Printf("❌ Error updating product %s: %v", update.StripeID, err)
			failedUpdates = append(failedUpdates, update.StripeID)
			continue
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Printf("❌ Error getting rows affected for %s: %v", update.StripeID, err)
			failedUpdates = append(failedUpdates, update.StripeID)
			continue
		}

		if rowsAffected > 0 {
			updatedCount++
		} else {
			log.Printf("⚠️ No product found with ID: %s", update.StripeID)
			failedUpdates = append(failedUpdates, update.StripeID)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("❌ Error committing transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to commit updates",
		})
		return
	}

	log.Printf("✅ Successfully updated %d out of %d Stripe products", updatedCount, len(requestBody.Updates))

	response := gin.H{
		"message":       "Bulk update completed",
		"updated_count": updatedCount,
		"total_count":   len(requestBody.Updates),
		"timestamp":     time.Now().Unix(),
	}

	if len(failedUpdates) > 0 {
		response["failed_updates"] = failedUpdates
		response["failed_count"] = len(failedUpdates)
	}

	c.JSON(http.StatusOK, response)
}

// debugStripeProductsData provides detailed information about stripe_products and stripe_prices tables
func debugStripeProductsData(c *gin.Context, db *database.DB) {
	log.Printf("🔍 Debugging Stripe products data...")

	// First, check what's in stripe_products table
	productsQuery := `
		SELECT id, stripe_id, name, description, active, available, created_at
		FROM stripe_products
		ORDER BY created_at DESC
		LIMIT 10
	`

	productsRows, err := db.DB.Query(productsQuery)
	if err != nil {
		log.Printf("❌ Error querying stripe_products: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to query stripe_products table",
		})
		return
	}
	defer productsRows.Close()

	var products []map[string]interface{}
	for productsRows.Next() {
		var id int
		var stripeID, name string
		var description *string
		var active, available bool
		var createdAt time.Time

		err := productsRows.Scan(&id, &stripeID, &name, &description, &active, &available, &createdAt)
		if err != nil {
			log.Printf("❌ Error scanning product: %v", err)
			continue
		}

		product := map[string]interface{}{
			"id":          id,
			"stripe_id":   stripeID,
			"name":        name,
			"description": description,
			"active":      active,
			"available":   available,
			"created_at":  createdAt.Format(time.RFC3339),
		}
		products = append(products, product)
	}

	// Check what's in stripe_prices table
	pricesQuery := `
		SELECT id, stripe_id, product_id, unit_amount, currency, recurring_interval, active, created_at
		FROM stripe_prices
		ORDER BY created_at DESC
		LIMIT 10
	`

	pricesRows, err := db.DB.Query(pricesQuery)
	if err != nil {
		log.Printf("❌ Error querying stripe_prices: %v", err)
		// Continue even if prices query fails
	}

	var prices []map[string]interface{}
	if err == nil {
		defer pricesRows.Close()
		for pricesRows.Next() {
			var id int
			var stripeID, productID string
			var unitAmount *int64
			var currency, recurringInterval *string
			var active bool
			var createdAt time.Time

			err := pricesRows.Scan(&id, &stripeID, &productID, &unitAmount, &currency, &recurringInterval, &active, &createdAt)
			if err != nil {
				log.Printf("❌ Error scanning price: %v", err)
				continue
			}

			price := map[string]interface{}{
				"id":                 id,
				"stripe_id":          stripeID,
				"product_id":         productID,
				"unit_amount":        unitAmount,
				"currency":           currency,
				"recurring_interval": recurringInterval,
				"active":             active,
				"created_at":         createdAt.Format(time.RFC3339),
			}
			prices = append(prices, price)
		}
	}

	// Get table counts
	var productsCount, pricesCount int
	db.DB.QueryRow("SELECT COUNT(*) FROM stripe_products").Scan(&productsCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM stripe_prices").Scan(&pricesCount)

	// Check for products with associated prices
	joinQuery := `
		SELECT 
			p.stripe_id as product_stripe_id,
			p.name as product_name,
			pr.stripe_id as price_stripe_id,
			pr.unit_amount,
			pr.currency
		FROM stripe_products p
		LEFT JOIN stripe_prices pr ON p.stripe_id = pr.product_id
		LIMIT 5
	`

	joinRows, err := db.DB.Query(joinQuery)
	var joinedData []map[string]interface{}
	if err == nil {
		defer joinRows.Close()
		for joinRows.Next() {
			var productStripeID, productName string
			var priceStripeID *string
			var unitAmount *int64
			var currency *string

			err := joinRows.Scan(&productStripeID, &productName, &priceStripeID, &unitAmount, &currency)
			if err != nil {
				log.Printf("❌ Error scanning joined data: %v", err)
				continue
			}

			joined := map[string]interface{}{
				"product_stripe_id": productStripeID,
				"product_name":      productName,
				"price_stripe_id":   priceStripeID,
				"unit_amount":       unitAmount,
				"currency":          currency,
			}
			joinedData = append(joinedData, joined)
		}
	}

	log.Printf("📊 Debug results: %d products, %d prices", productsCount, pricesCount)

	c.JSON(http.StatusOK, gin.H{
		"stripe_products": gin.H{
			"count":   productsCount,
			"samples": products,
		},
		"stripe_prices": gin.H{
			"count":   pricesCount,
			"samples": prices,
		},
		"joined_data": gin.H{
			"description": "Products with their associated prices",
			"samples":     joinedData,
		},
		"analysis": gin.H{
			"products_without_prices": productsCount - len(joinedData),
			"has_prices_table":        pricesCount > 0,
		},
		"timestamp": time.Now().Unix(),
	})
}

// getMetadataHealth returns Stripe customer metadata health statistics
func getMetadataHealth(c *gin.Context, db *database.DB) {
	log.Printf("🏥 [METADATA-HEALTH] Checking metadata health from IP: %s", c.ClientIP())

	health, err := db.GetStripeMetadataHealthCheck()
	if err != nil {
		log.Printf("❌ [METADATA-HEALTH] Failed to get health check: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get metadata health",
			"details": err.Error(),
		})
		return
	}

	// Determine health status
	var status string
	var severity string

	if health.HealthPercentage >= 95 {
		status = "healthy"
		severity = "low"
	} else if health.HealthPercentage >= 80 {
		status = "warning"
		severity = "medium"
	} else {
		status = "critical"
		severity = "high"
	}

	log.Printf("✅ [METADATA-HEALTH] Health check completed: %.1f%% healthy", health.HealthPercentage)

	c.JSON(http.StatusOK, gin.H{
		"status":          status,
		"severity":        severity,
		"health":          health,
		"recommendations": getHealthRecommendations(health),
		"timestamp":       time.Now().Unix(),
	})
}

// fixMetadataCorruption fixes corrupted Stripe customer metadata
func fixMetadataCorruption(c *gin.Context, db *database.DB) {
	log.Printf("🔧 [METADATA-FIX] Fix initiated from IP: %s", c.ClientIP())

	// Check if this is a dry run
	dryRun := c.Query("dry_run") == "true"

	if dryRun {
		// Get health check to show what would be fixed
		health, err := db.GetStripeMetadataHealthCheck()
		if err != nil {
			log.Printf("❌ [METADATA-FIX] Failed to get health check for dry run: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to analyze metadata for dry run",
			})
			return
		}

		recordsToFix := health.MissingMetadata + health.IncorrectMetadata

		c.JSON(http.StatusOK, gin.H{
			"message":        "Dry run completed",
			"records_to_fix": recordsToFix,
			"current_health": health,
			"dry_run":        true,
		})
		return
	}

	// Actually fix the metadata
	err := db.FixStripeCustomerMetadata()
	if err != nil {
		log.Printf("❌ [METADATA-FIX] Failed to fix metadata: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fix metadata corruption",
			"details": err.Error(),
		})
		return
	}

	// Get updated health check
	health, err := db.GetStripeMetadataHealthCheck()
	if err != nil {
		log.Printf("⚠️ [METADATA-FIX] Fix completed but failed to get updated health: %v", err)
		health = nil
	}

	log.Printf("✅ [METADATA-FIX] Metadata fix completed successfully")

	c.JSON(http.StatusOK, gin.H{
		"message":        "Metadata corruption fixed successfully",
		"success":        true,
		"updated_health": health,
		"timestamp":      time.Now().Unix(),
	})
}

// getHealthRecommendations provides recommendations based on health status
func getHealthRecommendations(health *database.StripeMetadataHealth) []string {
	var recommendations []string

	if health.MissingMetadata > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Fix %d customers with missing metadata", health.MissingMetadata))
	}

	if health.IncorrectMetadata > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Fix %d customers with incorrect metadata", health.IncorrectMetadata))
	}

	if health.OrphanedCustomers > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Review %d orphaned customers (no matching users)", health.OrphanedCustomers))
	}

	if health.HealthPercentage < 95 {
		recommendations = append(recommendations, "Run metadata fix to improve system health")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Metadata is healthy - no action needed")
	}

	return recommendations
}
