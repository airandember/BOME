package routes

import (
	"context"
	"fmt"
	"log"
	"net/http"
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

		// Manual UI-triggered sync endpoints (for frontend users)
		stripe.POST("/sync/trigger", func(c *gin.Context) { triggerManualSync(c, syncService) })
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
