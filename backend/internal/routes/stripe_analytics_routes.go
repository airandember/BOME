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
	log.Printf("🚀 [DASH-START] Dashboard request initiated at %v", startTime)

	// Check if Stripe is enabled first
	if !stripeService.IsEnabled() {
		log.Printf("❌ [DASH-ERROR] Stripe service is not enabled")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Stripe service is not enabled",
			"enabled": false,
		})
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
	err := db.QueryRow("SELECT COUNT(*) FROM stripe_customers").Scan(&customerCount)
	if err != nil {
		log.Printf("Error getting customer count: %v", err)
		customerCount = 0
	}

	// Get subscription count from stripe_subscriptions table
	var subscriptionCount int
	err = db.QueryRow("SELECT COUNT(*) FROM stripe_subscriptions").Scan(&subscriptionCount)
	if err != nil {
		log.Printf("Error getting subscription count: %v", err)
		subscriptionCount = 0
	}

	// Get product count from stripe_products table
	var productCount int
	err = db.QueryRow("SELECT COUNT(*) FROM stripe_products").Scan(&productCount)
	if err != nil {
		log.Printf("Error getting product count: %v", err)
		productCount = 0
	}

	// Get invoice count from stripe_invoices table
	var invoiceCount int
	err = db.QueryRow("SELECT COUNT(*) FROM stripe_invoices").Scan(&invoiceCount)
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

	log.Printf("📊 Database stats: %d customers, %d subscriptions, %d products, %d invoices",
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
	err = db.QueryRow("SELECT COUNT(*) FROM stripe_customers").Scan(&totalCount)
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

	rows, err := db.Query(baseQuery, limit, offset)
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
	err = db.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		log.Printf("Error getting subscription count: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription count"})
		return
	}

	rows, err := db.Query(baseQuery, args...)
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

	rows, err := db.Query(query, stripeCustomerID)
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
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":           "Invalid sync type. Use 'customers', 'initial', 'coupons', or 'monthly_metrics'",
			"available_types": []string{"customers", "initial", "coupons", "monthly_metrics"},
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
