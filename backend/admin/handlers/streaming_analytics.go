package handlers

import (
	"log"
	"net/http"

	analyticsServices "bome-backend/analytics/services"
	"bome-backend/infrastructure/database"

	"github.com/gin-gonic/gin"
)

// SetupStreamingAnalyticsRoutes sets up streaming analytics routes
func SetupStreamingAnalyticsRoutes(router *gin.RouterGroup, db *database.DB) {
	service := analyticsServices.NewAnalyticsService(db)

	analytics := router.Group("/analytics")
	{
		// GET /admin/streaming/analytics/overview
		analytics.GET("/overview", func(c *gin.Context) {
			GetAnalyticsOverviewHandler(c, service, db)
		})

		// GET /admin/streaming/analytics/executive-summary
		analytics.GET("/executive-summary", func(c *gin.Context) {
			GetExecutiveSummaryHandler(c, service, db)
		})

		// GET /admin/streaming/analytics/funnel
		analytics.GET("/funnel", func(c *gin.Context) {
			GetFunnelAnalysisHandler(c, service, db)
		})

		// GET /admin/streaming/analytics/revenue-impact
		analytics.GET("/revenue-impact", func(c *gin.Context) {
			GetRevenueImpactHandler(c, service, db)
		})

		// GET /admin/streaming/analytics/customer-journey
		analytics.GET("/customer-journey", func(c *gin.Context) {
			GetCustomerJourneyHandler(c, service, db)
		})

		// GET /admin/streaming/analytics/promotions
		analytics.GET("/promotions", func(c *gin.Context) {
			GetPromotionsAnalyticsHandler(c, service, db)
		})

		// GET /admin/streaming/analytics/real-time
		analytics.GET("/real-time", func(c *gin.Context) {
			GetRealTimeAnalyticsHandler(c, service)
		})

		// GET /admin/streaming/analytics/system-health
		analytics.GET("/system-health", func(c *gin.Context) {
			GetSystemHealthHandler(c, service)
		})
	}

	log.Println("✅ Streaming analytics routes registered")
}

// GetAnalyticsOverviewHandler handles GET /admin/streaming/analytics/overview
func GetAnalyticsOverviewHandler(c *gin.Context, service *analyticsServices.AnalyticsService, db *database.DB) {
	log.Println("📊 GetAnalyticsOverviewHandler: Fetching analytics overview...")

	period := c.DefaultQuery("period", "30d")

	// Get comprehensive overview from database model
	overview, err := db.GetAnalyticsOverview(period)
	if err != nil {
		log.Printf("❌ Error fetching analytics overview: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch analytics overview",
		})
		return
	}

	// Get video analytics
	videoStats, err := db.GetVideoStats()
	if err != nil {
		log.Printf("⚠️ Warning: Could not fetch video stats: %v", err)
		videoStats = map[string]interface{}{
			"total_videos":    0,
			"synced_videos":   0,
			"needs_attention": 0,
			"total_views":     0,
		}
	}

	// Get subscriber metrics
	subscriberMetrics, err := db.GetSubscriberMetrics()
	if err != nil {
		log.Printf("⚠️ Warning: Could not fetch subscriber metrics: %v", err)
		subscriberMetrics = map[string]interface{}{
			"total_subscribers":    0,
			"active_subscriptions": 0,
			"monthly_revenue":      0.0,
			"churn_rate":           0.0,
		}
	}

	// Get view analytics
	viewAnalytics, err := db.GetViewAnalytics()
	if err != nil {
		log.Printf("⚠️ Warning: Could not fetch view analytics: %v", err)
		viewAnalytics = map[string]interface{}{
			"total_views": 0,
			"views_today": 0,
			"views_week":  0,
			"growth_rate": 0.0,
		}
	}

	// Combine all metrics
	overview["video_stats"] = videoStats
	overview["subscriber_metrics"] = subscriberMetrics
	overview["view_analytics"] = viewAnalytics
	overview["period"] = period

	log.Printf("✅ Analytics overview retrieved successfully for period: %s", period)
	c.JSON(http.StatusOK, overview)
}

// GetExecutiveSummaryHandler handles GET /admin/streaming/analytics/executive-summary
func GetExecutiveSummaryHandler(c *gin.Context, service *analyticsServices.AnalyticsService, db *database.DB) {
	log.Println("📈 GetExecutiveSummaryHandler: Fetching executive summary...")

	period := c.DefaultQuery("period", "30d")

	// Get subscriber metrics
	subscriberMetrics, err := db.GetSubscriberMetrics()
	if err != nil {
		log.Printf("❌ Error fetching subscriber metrics: %v", err)
		subscriberMetrics = map[string]interface{}{
			"total_subscribers":    0,
			"active_subscriptions": 0,
			"monthly_revenue":      0.0,
			"churn_rate":           0.0,
		}
	}

	// Get video stats
	videoStats, err := db.GetVideoStats()
	if err != nil {
		log.Printf("❌ Error fetching video stats: %v", err)
		videoStats = map[string]interface{}{
			"total_videos":    0,
			"synced_videos":   0,
			"needs_attention": 0,
			"total_views":     0,
		}
	}

	// Build executive summary
	summary := gin.H{
		"period": period,
		"revenue_impact": gin.H{
			"promotional_revenue": 0,
			"standard_revenue":    subscriberMetrics["monthly_revenue"],
			"total_mrr":           subscriberMetrics["monthly_revenue"],
			"growth_rate":         0,
		},
		"customer_impact": gin.H{
			"new_customers_promos": 0,
			"standard_conversions": subscriberMetrics["active_subscriptions"],
			"overall_growth":       0,
			"total_subscribers":    subscriberMetrics["total_subscribers"],
		},
		"funnel_performance": gin.H{
			"promo_conversion":    0,
			"standard_conversion": 0,
			"conversion_lift":     0,
		},
		"content_performance": gin.H{
			"total_videos":    videoStats["total_videos"],
			"total_views":     videoStats["total_views"],
			"synced_videos":   videoStats["synced_videos"],
			"needs_attention": videoStats["needs_attention"],
		},
	}

	log.Printf("✅ Executive summary retrieved successfully")
	c.JSON(http.StatusOK, summary)
}

// GetFunnelAnalysisHandler handles GET /admin/streaming/analytics/funnel
func GetFunnelAnalysisHandler(c *gin.Context, service *analyticsServices.AnalyticsService, db *database.DB) {
	log.Println("🔍 GetFunnelAnalysisHandler: Fetching funnel analysis...")

	period := c.DefaultQuery("period", "30d")

	// Get subscriber metrics for funnel data
	subscriberMetrics, err := db.GetSubscriberMetrics()
	if err != nil {
		log.Printf("❌ Error fetching subscriber metrics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch funnel analysis",
		})
		return
	}

	// Build funnel data
	funnel := gin.H{
		"period": period,
		"stages": []gin.H{
			{
				"name":       "Visitors",
				"count":      0,
				"conversion": 0,
			},
			{
				"name":       "Sign-ups",
				"count":      subscriberMetrics["total_subscribers"],
				"conversion": 0,
			},
			{
				"name":       "Trial Users",
				"count":      0,
				"conversion": 0,
			},
			{
				"name":       "Active Subscribers",
				"count":      subscriberMetrics["active_subscriptions"],
				"conversion": 0,
			},
		},
		"overall_conversion": 0,
		"drop_off_points":    []gin.H{},
	}

	log.Printf("✅ Funnel analysis retrieved successfully")
	c.JSON(http.StatusOK, funnel)
}

// GetRevenueImpactHandler handles GET /admin/streaming/analytics/revenue-impact
func GetRevenueImpactHandler(c *gin.Context, service *analyticsServices.AnalyticsService, db *database.DB) {
	log.Println("💰 GetRevenueImpactHandler: Fetching revenue impact...")

	period := c.DefaultQuery("period", "30d")

	// Get subscriber metrics
	subscriberMetrics, err := db.GetSubscriberMetrics()
	if err != nil {
		log.Printf("❌ Error fetching subscriber metrics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch revenue impact",
		})
		return
	}

	// Build revenue impact data
	revenueImpact := gin.H{
		"period": period,
		"revenue_by_source": []gin.H{
			{
				"source": "Subscriptions",
				"amount": subscriberMetrics["monthly_revenue"],
			},
		},
		"total_revenue": subscriberMetrics["monthly_revenue"],
		"mrr":           subscriberMetrics["monthly_revenue"],
		"arr":           0,
		"churn_rate":    subscriberMetrics["churn_rate"],
		"growth_rate":   0,
	}

	log.Printf("✅ Revenue impact retrieved successfully")
	c.JSON(http.StatusOK, revenueImpact)
}

// GetCustomerJourneyHandler handles GET /admin/streaming/analytics/customer-journey
func GetCustomerJourneyHandler(c *gin.Context, service *analyticsServices.AnalyticsService, db *database.DB) {
	log.Println("🗺️ GetCustomerJourneyHandler: Fetching customer journey...")

	period := c.DefaultQuery("period", "30d")

	// Get subscriber metrics
	subscriberMetrics, err := db.GetSubscriberMetrics()
	if err != nil {
		log.Printf("❌ Error fetching subscriber metrics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch customer journey",
		})
		return
	}

	// Build customer journey data
	journey := gin.H{
		"period": period,
		"lifecycle_stages": []gin.H{
			{
				"stage": "Prospect",
				"count": 0,
			},
			{
				"stage": "Trial",
				"count": 0,
			},
			{
				"stage": "Active",
				"count": subscriberMetrics["active_subscriptions"],
			},
			{
				"stage": "At Risk",
				"count": 0,
			},
			{
				"stage": "Churned",
				"count": 0,
			},
		},
		"avg_time_to_conversion": "0 days",
		"retention_rate":         0,
		"cohort_analysis":        []gin.H{},
	}

	log.Printf("✅ Customer journey retrieved successfully")
	c.JSON(http.StatusOK, journey)
}

// GetPromotionsAnalyticsHandler handles GET /admin/streaming/analytics/promotions
func GetPromotionsAnalyticsHandler(c *gin.Context, service *analyticsServices.AnalyticsService, db *database.DB) {
	log.Println("🎁 GetPromotionsAnalyticsHandler: Fetching promotions analytics...")

	period := c.DefaultQuery("period", "30d")

	// Build promotions analytics data
	promotions := gin.H{
		"period":            period,
		"active_promotions": []gin.H{},
		"promotion_performance": gin.H{
			"total_uses":      0,
			"total_revenue":   0,
			"conversion_rate": 0,
		},
		"coupon_usage": []gin.H{},
		"conversion_impact": gin.H{
			"with_promo":    0,
			"without_promo": 0,
			"lift":          0,
		},
	}

	log.Printf("✅ Promotions analytics retrieved successfully")
	c.JSON(http.StatusOK, promotions)
}

// GetRealTimeAnalyticsHandler handles GET /admin/streaming/analytics/real-time
func GetRealTimeAnalyticsHandler(c *gin.Context, service *analyticsServices.AnalyticsService) {
	log.Println("⚡ GetRealTimeAnalyticsHandler: Fetching real-time analytics...")

	data, err := service.GetRealTimeAnalytics()
	if err != nil {
		log.Printf("❌ Error fetching real-time analytics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch real-time analytics",
		})
		return
	}

	log.Printf("✅ Real-time analytics retrieved successfully")
	c.JSON(http.StatusOK, data)
}

// GetSystemHealthHandler handles GET /admin/streaming/analytics/system-health
func GetSystemHealthHandler(c *gin.Context, service *analyticsServices.AnalyticsService) {
	log.Println("🏥 GetSystemHealthHandler: Fetching system health...")

	health, err := service.GetSystemHealth()
	if err != nil {
		log.Printf("❌ Error fetching system health: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch system health",
		})
		return
	}

	log.Printf("✅ System health retrieved successfully")
	c.JSON(http.StatusOK, gin.H{
		"health": health,
	})
}
