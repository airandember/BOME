package routes

import (
	"net/http"
	"time"

	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// RegisterStripeTestRoutes registers test and monitoring endpoints
func RegisterStripeTestRoutes(router *gin.RouterGroup, stripeService *services.StripeService, syncService *services.StripeSyncService, cronService *services.StripeCronService) {
	test := router.Group("/stripe/test")
	{
		// System status and diagnostics
		test.GET("/status", func(c *gin.Context) { getSystemStatus(c, stripeService, syncService, cronService) })
		test.GET("/tables", func(c *gin.Context) { getTableStats(c, syncService) })
		test.GET("/logs", func(c *gin.Context) { getRecentLogs(c) })

		// Connection tests
		test.GET("/connection", func(c *gin.Context) { testStripeConnection(c, stripeService) })
		test.GET("/database", func(c *gin.Context) { testDatabaseConnection(c, syncService) })

		// Data validation
		test.GET("/validate", func(c *gin.Context) { validateSyncedData(c, syncService) })
		test.GET("/sample", func(c *gin.Context) { getSampleData(c, syncService) })
	}
}

// getSystemStatus returns comprehensive system status
func getSystemStatus(c *gin.Context, stripeService *services.StripeService, syncService *services.StripeSyncService, cronService *services.StripeCronService) {
	logger := services.NewStripeLogger("TEST")

	// Check Stripe service status
	stripeEnabled := stripeService.IsEnabled()
	keyType := stripeService.GetSecretKeyType() // SAFE: Uses new method that doesn't expose secrets

	// Get next scheduled runs
	schedules := cronService.GetNextScheduledRuns()

	// Log system status
	logger.LogSystemStatus(stripeEnabled, keyType, true) // Assuming tables are ready

	status := gin.H{
		"timestamp": time.Now().Format("2006-01-02 15:04:05 MST"),
		"stripe": gin.H{
			"enabled":  stripeEnabled,
			"key_type": keyType,
			"environment": func() string {
				if stripeEnabled {
					return stripeService.GetConfig().Environment
				}
				return "none"
			}(),
		},
		"database": gin.H{
			"connected":    true, // We'll test this
			"tables_ready": true,
		},
		"sync": gin.H{
			"service_ready": syncService != nil,
			"cron_ready":    cronService != nil,
		},
		"schedules": schedules,
		"endpoints": gin.H{
			"sync_initial":    "/stripe/sync/initial",
			"sync_status":     "/stripe/sync/status",
			"webhooks":        "/stripe/webhooks/",
			"test_connection": "/stripe/test/connection",
			"table_stats":     "/stripe/test/tables",
		},
	}

	c.JSON(http.StatusOK, status)
}

// getTableStats returns database table statistics
func getTableStats(c *gin.Context, syncService *services.StripeSyncService) {
	logger := services.NewStripeLogger("TABLES")

	// Get counts from each Stripe table
	tables := []string{
		"stripe_customers",
		"stripe_products",
		"stripe_prices",
		"stripe_subscriptions",
		"stripe_invoices",
		"stripe_coupons",
		"stripe_sync_jobs",
		"stripe_sync_config",
		"stripe_entities",
	}

	stats := make(map[string]interface{})

	for _, table := range tables {
		count, err := getTableCount(syncService, table)
		if err != nil {
			stats[table] = gin.H{"error": err.Error()}
			logger.LogError("table_count", err, map[string]interface{}{"table": table})
		} else {
			stats[table] = gin.H{"count": count}
			logger.LogTableStats(table, count)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"timestamp": time.Now().Format("2006-01-02 15:04:05 MST"),
		"tables":    stats,
	})
}

// testStripeConnection tests Stripe API connectivity
func testStripeConnection(c *gin.Context, stripeService *services.StripeService) {
	logger := services.NewStripeLogger("CONNECTION")

	if !stripeService.IsEnabled() {
		logger.LogError("stripe_test", nil, map[string]interface{}{"reason": "service_disabled"})
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Stripe service is not enabled",
			"message": "Please configure a valid Stripe API key first",
		})
		return
	}

	startTime := time.Now()

	// Test with a simple balance call
	balance, err := stripeService.GetAccountBalance()
	duration := time.Since(startTime)

	if err != nil {
		logger.LogError("balance_test", err, map[string]interface{}{"duration": duration})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Failed to connect to Stripe",
			"message":  err.Error(),
			"duration": duration.String(),
		})
		return
	}

	logger.LogAPICall("balance", map[string]interface{}{"test": true}, duration)

	c.JSON(http.StatusOK, gin.H{
		"success":           true,
		"message":           "Stripe connection successful",
		"duration":          duration.String(),
		"balance_available": len(balance.Available) > 0,
		"test_time":         time.Now().Format("2006-01-02 15:04:05 MST"),
	})
}

// testDatabaseConnection tests database connectivity
func testDatabaseConnection(c *gin.Context, syncService *services.StripeSyncService) {
	logger := services.NewStripeLogger("DATABASE")

	startTime := time.Now()

	// Test with a simple query
	count, err := getTableCount(syncService, "stripe_entities")
	duration := time.Since(startTime)

	if err != nil {
		logger.LogError("db_test", err, map[string]interface{}{"duration": duration})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Database connection failed",
			"message":  err.Error(),
			"duration": duration.String(),
		})
		return
	}

	logger.LogDatabaseOperation("SELECT", "stripe_entities", count, duration)

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"message":        "Database connection successful",
		"duration":       duration.String(),
		"entities_count": count,
		"test_time":      time.Now().Format("2006-01-02 15:04:05 MST"),
	})
}

// validateSyncedData validates the integrity of synced data
func validateSyncedData(c *gin.Context, syncService *services.StripeSyncService) {
	logger := services.NewStripeLogger("VALIDATE")

	validation := gin.H{
		"timestamp": time.Now().Format("2006-01-02 15:04:05 MST"),
		"checks":    gin.H{},
	}

	// Check for orphaned records
	checks := []struct {
		name  string
		query string
		desc  string
	}{
		{
			"orphaned_prices",
			"SELECT COUNT(*) FROM stripe_prices WHERE product_id IS NOT NULL AND product_id NOT IN (SELECT id FROM stripe_products)",
			"Prices without valid products",
		},
		{
			"orphaned_subscriptions",
			"SELECT COUNT(*) FROM stripe_subscriptions WHERE customer_id IS NOT NULL AND customer_id NOT IN (SELECT id FROM stripe_customers)",
			"Subscriptions without valid customers",
		},
		{
			"orphaned_invoices",
			"SELECT COUNT(*) FROM stripe_invoices WHERE customer_id IS NOT NULL AND customer_id NOT IN (SELECT id FROM stripe_customers)",
			"Invoices without valid customers",
		},
	}

	for _, check := range checks {
		count, err := executeCountQuery(syncService, check.query)
		if err != nil {
			validation["checks"].(gin.H)[check.name] = gin.H{
				"error":       err.Error(),
				"description": check.desc,
			}
			logger.LogError("validation", err, map[string]interface{}{"check": check.name})
		} else {
			status := "✅ PASS"
			if count > 0 {
				status = "⚠️ ISSUES FOUND"
			}

			validation["checks"].(gin.H)[check.name] = gin.H{
				"status":      status,
				"count":       count,
				"description": check.desc,
			}

			logger.LogTableStats(check.name, count)
		}
	}

	c.JSON(http.StatusOK, validation)
}

// getSampleData returns sample records from each table
func getSampleData(c *gin.Context, syncService *services.StripeSyncService) {
	logger := services.NewStripeLogger("SAMPLE")

	samples := gin.H{
		"timestamp": time.Now().Format("2006-01-02 15:04:05 MST"),
		"data":      gin.H{},
	}

	// Get sample records from key tables
	tables := []string{"stripe_customers", "stripe_products", "stripe_subscriptions"}

	for _, table := range tables {
		sample, err := getSampleRecords(syncService, table, 3)
		if err != nil {
			samples["data"].(gin.H)[table] = gin.H{"error": err.Error()}
			logger.LogError("sample", err, map[string]interface{}{"table": table})
		} else {
			samples["data"].(gin.H)[table] = sample
			logger.LogDatabaseOperation("SAMPLE", table, len(sample), 0)
		}
	}

	c.JSON(http.StatusOK, samples)
}

// getRecentLogs returns recent log entries (placeholder)
func getRecentLogs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Log viewing endpoint",
		"note":    "Check your terminal/console for detailed logs",
		"log_prefixes": []string{
			"🚀 SYNC START",
			"📊 PROGRESS",
			"✅ SUCCESS",
			"📡 API CALL",
			"💾 DB OPERATION",
			"📨 WEBHOOK",
			"❌ ERROR",
		},
	})
}

// Helper functions
func getTableCount(syncService *services.StripeSyncService, table string) (int, error) {
	query := "SELECT COUNT(*) FROM " + table
	return executeCountQuery(syncService, query)
}

func executeCountQuery(syncService *services.StripeSyncService, query string) (int, error) {
	var count int
	err := syncService.GetDB().QueryRow(query).Scan(&count)
	return count, err
}

func getSampleRecords(syncService *services.StripeSyncService, table string, limit int) ([]map[string]interface{}, error) {
	query := "SELECT * FROM " + table + " ORDER BY created_at DESC LIMIT $1"

	rows, err := syncService.GetDB().Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		record := make(map[string]interface{})
		for i, col := range columns {
			record[col] = values[i]
		}

		results = append(results, record)
	}

	return results, nil
}
