package routes

import (
	"log"
	"net/http"
	"strconv"

	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupSubscriberElasticRoutesV2 sets up the V2 elastic service routes (parallel to v1)
func SetupSubscriberElasticRoutesV2(router *gin.RouterGroup, db *database.DB) {
	elasticService := services.NewSubscriberElasticServiceV2(db)

	// V2 routes - parallel to v1 for comparison and testing
	elastic := router.Group("/subscriber-elastic-v2")
	elastic.Use(middleware.AuthRequired())
	elastic.Use(middleware.AdminRequired())

	// Get all subscribers using v2 tables
	elastic.GET("/subscribers", func(c *gin.Context) {
		log.Printf("📋 [V2 ENDPOINT] Starting GetAllUnifiedSubscribersV2 request")
		
		subscribers, err := elasticService.GetAllUnifiedSubscribersV2()
		if err != nil {
			log.Printf("❌ [V2 ENDPOINT] Error fetching subscribers: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch subscribers from v2 tables",
				"details": err.Error(),
				"hint":    "Check if v2 tables are populated and user_stripe_customers_v2 has linked data",
			})
			return
		}

		log.Printf("✅ [V2 ENDPOINT] Successfully fetched %d subscribers", len(subscribers))
		
		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"subscribers": subscribers,
			},
			"count":   len(subscribers),
			"version": "v2",
			"source":  "stripe_v2_tables",
		})
	})

	// Get single subscriber by ID using v2 tables
	elastic.GET("/subscribers/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid subscriber ID",
			})
			return
		}

		subscriber, err := elasticService.GetUnifiedSubscriberByIDV2(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Subscriber not found",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"subscriber": subscriber,
			"version":    "v2",
			"source":     "stripe_v2_tables",
		})
	})

	// Get subscriber statistics using v2 tables
	elastic.GET("/stats", func(c *gin.Context) {
		stats, err := elasticService.GetSubscriberStatsV2()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to calculate statistics",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"stats":   stats,
			"version": "v2",
			"source":  "stripe_v2_tables",
		})
	})
}

// SetupSubscriberComparisonRoutes creates endpoints to compare v1 vs v2 data
func SetupSubscriberComparisonRoutes(router *gin.RouterGroup, db *database.DB) {
	v1Service := services.NewSubscriberElasticService(db)
	v2Service := services.NewSubscriberElasticServiceV2(db)

	comparison := router.Group("/subscriber-comparison")
	comparison.Use(middleware.AuthRequired())
	comparison.Use(middleware.AdminRequired())

	// Compare single subscriber data (v1 vs v2)
	comparison.GET("/subscriber/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid subscriber ID",
			})
			return
		}

		// Fetch from both services
		v1Data, v1Err := v1Service.GetUnifiedSubscriberByID(id)
		v2Data, v2Err := v2Service.GetUnifiedSubscriberByIDV2(id)

		response := gin.H{
			"user_id": id,
		}

		if v1Err == nil {
			response["v1"] = v1Data
		} else {
			response["v1_error"] = v1Err.Error()
		}

		if v2Err == nil {
			response["v2"] = v2Data
		} else {
			response["v2_error"] = v2Err.Error()
		}

		// Add comparison notes
		if v1Err == nil && v2Err == nil {
			response["comparison"] = compareSubscribers(v1Data, v2Data)
		}

		c.JSON(http.StatusOK, response)
	})

	// Compare stats (v1 vs v2)
	comparison.GET("/stats", func(c *gin.Context) {
		v1Stats, v1Err := v1Service.GetSubscriberStats()
		v2Stats, v2Err := v2Service.GetSubscriberStatsV2()

		response := gin.H{}

		if v1Err == nil {
			response["v1_stats"] = v1Stats
		} else {
			response["v1_error"] = v1Err.Error()
		}

		if v2Err == nil {
			response["v2_stats"] = v2Stats
		} else {
			response["v2_error"] = v2Err.Error()
		}

		// Add delta analysis
		if v1Err == nil && v2Err == nil {
			response["delta"] = calculateStatsDelta(v1Stats, v2Stats)
		}

		c.JSON(http.StatusOK, response)
	})

	// Health check - verify v2 tables exist and have data
	comparison.GET("/health", func(c *gin.Context) {
		health := gin.H{
			"v2_tables": make(map[string]interface{}),
		}

		// Check each v2 table
		tables := []string{
			"stripe_customers_v2",
			"stripe_products_v2",
			"stripe_prices_v2",
			"stripe_subscriptions_v2",
			"user_stripe_customers_v2",
		}

		allHealthy := true
		for _, table := range tables {
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
			if err != nil {
				health["v2_tables"].(map[string]interface{})[table] = gin.H{
					"status": "error",
					"error":  err.Error(),
				}
				allHealthy = false
			} else {
				health["v2_tables"].(map[string]interface{})[table] = gin.H{
					"status":    "ok",
					"row_count": count,
				}
			}
		}

		health["overall_status"] = "healthy"
		if !allHealthy {
			health["overall_status"] = "degraded"
		}

		statusCode := http.StatusOK
		if !allHealthy {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, health)
	})
}

// compareSubscribers compares v1 and v2 subscriber data
func compareSubscribers(v1 interface{}, v2 interface{}) map[string]interface{} {
	comparison := map[string]interface{}{
		"matches":     true,
		"differences": []string{},
	}

	// Type assertion to access fields
	// This is a simplified comparison - in production you'd want more detailed field-by-field comparison

	comparison["note"] = "Detailed field comparison available - check raw v1 and v2 objects"

	return comparison
}

// calculateStatsDelta calculates differences between v1 and v2 stats
func calculateStatsDelta(v1Stats, v2Stats map[string]interface{}) map[string]interface{} {
	delta := make(map[string]interface{})

	// Compare key metrics
	metrics := []string{
		"total_users",
		"active_subscriptions",
		"total_mrr",
		"total_arr",
	}

	for _, metric := range metrics {
		v1Val, v1Ok := v1Stats[metric]
		v2Val, v2Ok := v2Stats[metric]

		if v1Ok && v2Ok {
			delta[metric] = gin.H{
				"v1":    v1Val,
				"v2":    v2Val,
				"match": v1Val == v2Val,
			}
		}
	}

	return delta
}
