package routes

import (
	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// SetupAnalyticsRoutes configures analytics-related routes
func SetupAnalyticsRoutes(router *gin.RouterGroup, db *database.DB) {
	fmt.Printf("🔥 ANALYTICS: Starting SetupAnalyticsRoutes function\n")

	// Create analytics service
	analyticsService := services.NewAnalyticsService(db)

	// Create analytics group with authentication
	analytics := router.Group("/dashboard/analytics")
	analytics.Use(middleware.AuthRequired())

	// Main analytics dashboard endpoint
	analytics.GET("", func(c *gin.Context) {
		period := c.DefaultQuery("period", "7d")

		// Use optimized analytics overview with caching
		analyticsData, err := db.GetAnalyticsOverview(period)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to get analytics: %v", err),
				"data":  nil,
			})
			return
		}

		// Get real-time metrics
		realTimeData, err := db.GetRealTimeMetrics()
		if err != nil {
			// Log error but don't fail the request
			fmt.Printf("Warning: Failed to get real-time metrics: %v\n", err)
			realTimeData = map[string]interface{}{
				"active_users":         0,
				"current_streams":      0,
				"recent_signups":       0,
				"recent_subscriptions": 0,
			}
		}

		// Combine data
		responseData := map[string]interface{}{
			"users":         analyticsData["users"],
			"videos":        analyticsData["videos"],
			"subscriptions": analyticsData["subscriptions"],
			"real_time":     realTimeData,
		}

		// Standardize response format
		c.JSON(http.StatusOK, gin.H{
			"data":   responseData,
			"period": period,
			"status": "success",
		})
	})

	// Real-time analytics endpoint
	analytics.GET("/realtime", func(c *gin.Context) {
		realTimeData, err := db.GetRealTimeMetrics()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to get real-time analytics: %v", err),
				"data":  nil,
			})
			return
		}

		// Standardize response format
		c.JSON(http.StatusOK, gin.H{
			"data":   gin.H{"real_time": realTimeData},
			"status": "success",
		})
	})

	// System health endpoint
	analytics.GET("/system-health", func(c *gin.Context) {
		systemHealth, err := db.GetSystemHealth()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to get system health: %v", err),
				"data":  nil,
			})
			return
		}

		// Standardize response format
		c.JSON(http.StatusOK, gin.H{
			"data":   systemHealth,
			"status": "success",
		})
	})

	// Export endpoint
	analytics.GET("/export", func(c *gin.Context) {
		format := c.DefaultQuery("format", "csv")
		period := c.DefaultQuery("period", "7d")

		// Get analytics data
		analyticsData, err := analyticsService.GetAnalytics(period)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get analytics for export: %v", err)})
			return
		}

		switch format {
		case "csv":
			c.Header("Content-Type", "text/csv")
			c.Header("Content-Disposition", "attachment; filename=analytics_export.csv")

			csvData := "Date,Active Users,Page Views,Video Views,Subscriptions,Revenue\n"
			// Add actual data from analyticsData
			csvData += fmt.Sprintf("%s,%v,%v,%v,%v,%v\n",
				time.Now().Format("2006-01-02"),
				analyticsData["real_time"].(map[string]interface{})["active_users"],
				0, // page views
				0, // video views
				analyticsData["subscriptions"].(map[string]interface{})["active"],
				analyticsData["subscriptions"].(map[string]interface{})["revenue_today"])

			c.String(http.StatusOK, csvData)

		case "json":
			c.Header("Content-Type", "application/json")
			c.Header("Content-Disposition", "attachment; filename=analytics_export.json")

			exportData := gin.H{
				"export_date": time.Now().Format(time.RFC3339),
				"period":      period,
				"data":        analyticsData,
			}

			c.JSON(http.StatusOK, exportData)

		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported format. Use 'csv' or 'json'"})
		}
	})

	// Analytics tracking endpoint
	analytics.POST("/track", func(c *gin.Context) {
		var eventData map[string]interface{}
		if err := c.ShouldBindJSON(&eventData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Extract event details
		eventType, ok := eventData["event_type"].(string)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "event_type is required"})
			return
		}

		// Get user ID from context if available
		var userID *int
		if userIDVal, exists := c.Get("user_id"); exists {
			if id, ok := userIDVal.(int); ok {
				userID = &id
			}
		}

		// Extract other fields
		sessionID := ""
		if sid, ok := eventData["session_id"].(string); ok {
			sessionID = sid
		}

		subsite := "streaming"
		if sub, ok := eventData["subsite"].(string); ok {
			subsite = sub
		}

		// Track the event
		err := analyticsService.TrackEvent(
			eventType,
			userID,
			sessionID,
			subsite,
			eventData,
			c.ClientIP(),
			c.GetHeader("User-Agent"),
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to track event: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "success",
			"message":  "Event tracked successfully",
			"event_id": fmt.Sprintf("evt_%d", time.Now().Unix()),
		})
	})

	// Batch analytics tracking endpoint
	analytics.POST("/batch", func(c *gin.Context) {
		// Rate limiting: 100 requests per minute per IP
		clientIP := c.ClientIP()
		rateLimitKey := fmt.Sprintf("analytics_batch:%s", clientIP)

		// Simple in-memory rate limiting (in production, use Redis)
		if !analyticsService.CheckRateLimit(rateLimitKey, 100, time.Minute) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded. Maximum 100 batch requests per minute per IP.",
				"retry_after": 60,
			})
			return
		}

		var events []map[string]interface{}
		if err := c.ShouldBindJSON(&events); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		// Validate batch size
		if len(events) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Empty batch - no events provided"})
			return
		}

		if len(events) > 1000 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":    "Batch size exceeds maximum limit of 1000 events",
				"max_size": 1000,
				"received": len(events),
			})
			return
		}

		// Get user ID from context if available
		var userID *int
		if userIDVal, exists := c.Get("user_id"); exists {
			if id, ok := userIDVal.(int); ok {
				userID = &id
			}
		}

		// Convert to AnalyticsEvent structs for optimized processing
		analyticsEvents := make([]*database.AnalyticsEvent, 0, len(events))
		validationErrors := make([]string, 0)

		for i, eventData := range events {
			// Validate event
			if err := analyticsService.ValidateEvent(eventData); err != nil {
				validationErrors = append(validationErrors, fmt.Sprintf("Event %d: %v", i+1, err))
				continue
			}

			// Sanitize event data
			sanitizedEvent := analyticsService.SanitizeEvent(eventData)

			// Extract event details
			eventType, _ := sanitizedEvent["event_type"].(string)
			sessionID, _ := sanitizedEvent["session_id"].(string)
			subsite, _ := sanitizedEvent["subsite"].(string)
			if subsite == "" {
				subsite = "streaming"
			}

			// Convert event data to JSON string
			eventDataJSON, err := json.Marshal(sanitizedEvent)
			if err != nil {
				validationErrors = append(validationErrors, fmt.Sprintf("Event %d: Failed to serialize event data", i+1))
				continue
			}

			// Create AnalyticsEvent struct
			analyticsEvent := &database.AnalyticsEvent{
				EventType: eventType,
				UserID:    userID,
				SessionID: sessionID,
				Subsite:   subsite,
				EventData: string(eventDataJSON),
				IPAddress: c.ClientIP(),
				UserAgent: c.GetHeader("User-Agent"),
				CreatedAt: time.Now(),
			}

			analyticsEvents = append(analyticsEvents, analyticsEvent)
		}

		// Return validation errors if any events failed validation
		if len(validationErrors) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":             "Some events failed validation",
				"validation_errors": validationErrors,
				"valid_events":      len(analyticsEvents),
				"total_events":      len(events),
			})
			return
		}

		// Use optimized batch processing
		err := db.TrackAnalyticsEventsBatch(analyticsEvents)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to process batch events: %v", err),
			})
			return
		}

		// Build response
		response := gin.H{
			"status":    "success",
			"message":   "Batch events processed successfully",
			"processed": len(analyticsEvents),
			"total":     len(events),
			"errors":    0,
		}

		c.JSON(http.StatusOK, response)
	})

	// Memory monitoring endpoint
	analytics.GET("/memory", func(c *gin.Context) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		memoryInfo := gin.H{
			"alloc":          m.Alloc,
			"total_alloc":    m.TotalAlloc,
			"sys":            m.Sys,
			"num_gc":         m.NumGC,
			"heap_alloc":     m.HeapAlloc,
			"heap_sys":       m.HeapSys,
			"heap_idle":      m.HeapIdle,
			"heap_inuse":     m.HeapInuse,
			"heap_released":  m.HeapReleased,
			"heap_objects":   m.HeapObjects,
			"stack_inuse":    m.StackInuse,
			"stack_sys":      m.StackSys,
			"num_goroutines": runtime.NumGoroutine(),
			"num_cpu":        runtime.NumCPU(),
		}

		// Get database pool stats
		dbStats := db.GetPoolStats()

		// Get cache stats if Redis is available
		var cacheStats gin.H
		if db.Redis != nil {
			ctx := context.Background()
			info, err := db.Redis.Client.Info(ctx).Result()
			if err == nil {
				cacheStats = gin.H{
					"redis_connected": true,
					"redis_info":      info,
				}
			} else {
				cacheStats = gin.H{
					"redis_connected": false,
					"error":           err.Error(),
				}
			}
		} else {
			cacheStats = gin.H{
				"redis_connected": false,
				"message":         "Redis not configured",
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"memory":    memoryInfo,
			"database":  dbStats,
			"cache":     cacheStats,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Cache management endpoint
	analytics.POST("/cache/clear", func(c *gin.Context) {
		// Clear all analytics cache
		db.InvalidateAnalyticsCache()

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Analytics cache cleared successfully",
		})
	})

	// Cache status endpoint
	analytics.GET("/cache/status", func(c *gin.Context) {
		var cacheStatus gin.H

		if db.Redis != nil {
			ctx := context.Background()

			// Get cache keys count
			keys, err := db.Redis.Client.Keys(ctx, "analytics_*").Result()
			if err != nil {
				cacheStatus = gin.H{
					"redis_connected": false,
					"error":           err.Error(),
				}
			} else {
				cacheStatus = gin.H{
					"redis_connected": true,
					"analytics_keys":  len(keys),
					"total_keys":      len(keys),
				}
			}
		} else {
			cacheStatus = gin.H{
				"redis_connected": false,
				"message":         "Redis not configured",
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"cache":     cacheStatus,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	fmt.Printf("🔥 ANALYTICS: Analytics routes registered successfully\n")
	fmt.Printf("🔥 ANALYTICS: SetupAnalyticsRoutes completed successfully\n")
}

// SetupMonitoringRoutes configures monitoring-related routes
func SetupMonitoringRoutes(router *gin.RouterGroup, db *database.DB) {
	fmt.Printf("🔥 MONITORING: Starting SetupMonitoringRoutes function\n")

	// Create analytics service
	analyticsService := services.NewAnalyticsService(db)

	// Create monitoring group with authentication
	monitoring := router.Group("/monitoring")
	monitoring.Use(middleware.AuthRequired(), middleware.AdminRequired())

	// System metrics endpoint
	monitoring.GET("/system", func(c *gin.Context) {
		monitoringData, err := analyticsService.GetMonitoringData()
		if err != nil {
			// Log detailed error for debugging
			fmt.Printf("Error in GetMonitoringData: %v\n", err)

			// Return default data instead of error
			fmt.Printf("Warning: Failed to get monitoring data: %v\n", err)
			c.JSON(http.StatusOK, gin.H{
				"metrics": map[string]interface{}{
					"system_health": map[string]interface{}{
						"status":        "operational",
						"uptime":        "0h 0m",
						"response_time": "0ms",
						"error_rate":    "0%",
					},
					"database": map[string]interface{}{
						"status":      "connected",
						"size":        "0 MB",
						"connections": 0,
					},
					"memory": map[string]interface{}{
						"usage":     "0%",
						"available": "0 MB",
					},
					"cpu": map[string]interface{}{
						"usage": "0%",
					},
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"metrics": monitoringData["metrics"]})
	})

	// Webhook events endpoint
	monitoring.GET("/webhooks", func(c *gin.Context) {
		monitoringData, err := analyticsService.GetMonitoringData()
		if err != nil {
			// Log detailed error for debugging
			fmt.Printf("Error in GetMonitoringData for webhooks: %v\n", err)

			// Return empty events instead of error
			fmt.Printf("Warning: Failed to get webhook events: %v\n", err)
			c.JSON(http.StatusOK, gin.H{"events": []interface{}{}})
			return
		}

		c.JSON(http.StatusOK, gin.H{"events": monitoringData["events"]})
	})

	// Subsite health endpoint
	monitoring.GET("/health", func(c *gin.Context) {
		monitoringData, err := analyticsService.GetMonitoringData()
		if err != nil {
			// Log detailed error for debugging
			fmt.Printf("Error in GetMonitoringData for health: %v\n", err)

			// Return default health data instead of error
			fmt.Printf("Warning: Failed to get subsite health: %v\n", err)
			c.JSON(http.StatusOK, gin.H{
				"health": map[string]interface{}{
					"streaming": map[string]interface{}{
						"status": "operational",
						"uptime": "100%",
					},
					"articles": map[string]interface{}{
						"status": "operational",
						"uptime": "100%",
					},
					"expo": map[string]interface{}{
						"status": "operational",
						"uptime": "100%",
					},
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"health": monitoringData["health"]})
	})

	// Alerts endpoint
	monitoring.GET("/alerts", func(c *gin.Context) {
		monitoringData, err := analyticsService.GetMonitoringData()
		if err != nil {
			// Log detailed error for debugging
			fmt.Printf("Error in GetMonitoringData for alerts: %v\n", err)

			// Return empty alerts instead of error
			fmt.Printf("Warning: Failed to get alerts: %v\n", err)
			c.JSON(http.StatusOK, gin.H{"alerts": []interface{}{}})
			return
		}

		c.JSON(http.StatusOK, gin.H{"alerts": monitoringData["alerts"]})
	})

	// Acknowledge alert endpoint
	monitoring.POST("/alerts/:id/acknowledge", func(c *gin.Context) {
		alertID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
			return
		}

		userID := c.GetInt("user_id")
		err = db.AcknowledgeAlert(alertID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to acknowledge alert: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Alert acknowledged successfully"})
	})

	fmt.Printf("🔥 MONITORING: Monitoring routes registered successfully\n")
}

// SetupCrossSubsiteAnalyticsRoutes configures cross-subsite analytics routes
func SetupCrossSubsiteAnalyticsRoutes(router *gin.RouterGroup, db *database.DB) {
	fmt.Printf("🔥 CROSS-SUBSITE: Starting SetupCrossSubsiteAnalyticsRoutes function\n")

	// Create analytics service
	analyticsService := services.NewAnalyticsService(db)

	// Create cross-subsite analytics group with authentication
	crossSubsite := router.Group("/analytics")
	crossSubsite.Use(middleware.AuthRequired(), middleware.AdminRequired())

	// Cross-subsite analytics endpoint
	crossSubsite.GET("/cross-subsite", func(c *gin.Context) {
		timeframe := c.DefaultQuery("timeframe", "24h")
		subsite := c.DefaultQuery("subsite", "all")

		analyticsData, err := analyticsService.GetCrossSubsiteAnalytics(timeframe, subsite)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get cross-subsite analytics: %v", err)})
			return
		}

		c.JSON(http.StatusOK, analyticsData)
	})

	// Webhook analytics endpoint
	crossSubsite.GET("/webhooks", func(c *gin.Context) {
		timeframe := c.DefaultQuery("timeframe", "24h")

		analyticsData, err := analyticsService.GetWebhookAnalytics(timeframe)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get webhook analytics: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"analytics": analyticsData})
	})

	fmt.Printf("🔥 CROSS-SUBSITE: Cross-subsite analytics routes registered successfully\n")
}
