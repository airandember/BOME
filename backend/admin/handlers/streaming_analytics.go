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
			GetAnalyticsOverviewHandler(c, service)
		})

		// GET /admin/streaming/analytics/executive-summary
		analytics.GET("/executive-summary", func(c *gin.Context) {
			GetExecutiveSummaryHandler(c, service)
		})

		// GET /admin/streaming/analytics/funnel
		analytics.GET("/funnel", func(c *gin.Context) {
			GetFunnelAnalysisHandler(c, service)
		})

		// GET /admin/streaming/analytics/revenue-impact
		analytics.GET("/revenue-impact", func(c *gin.Context) {
			GetRevenueImpactHandler(c, service)
		})

		// GET /admin/streaming/analytics/customer-journey
		analytics.GET("/customer-journey", func(c *gin.Context) {
			GetCustomerJourneyHandler(c, service)
		})

		// GET /admin/streaming/analytics/promotions
		analytics.GET("/promotions", func(c *gin.Context) {
			GetPromotionsAnalyticsHandler(c, service)
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

// metricsFromAnalytics builds video_stats, subscriber_metrics, view_analytics from braid GetAnalytics result
func metricsFromAnalytics(overview map[string]interface{}) (videoStats, subscriberMetrics, viewAnalytics map[string]interface{}) {
	v, _ := overview["videos"].(map[string]interface{})
	s, _ := overview["subscriptions"].(map[string]interface{})
	videoStats = map[string]interface{}{
		"total_videos":    numOrZero(v, "total"),
		"synced_videos":  numOrZero(v, "published"),
		"needs_attention": numOrZero(v, "pending"),
		"total_views":    numOrZero(v, "total_views"),
	}
	subscriberMetrics = map[string]interface{}{
		"total_subscribers":    numOrZero(s, "active"),
		"active_subscriptions": numOrZero(s, "active"),
		"monthly_revenue":      floatOrZero(s, "mrr"),
		"churn_rate":           0.0,
	}
	viewAnalytics = map[string]interface{}{
		"total_views": numOrZero(v, "total_views"),
		"views_today": 0,
		"views_week":  0,
		"growth_rate": 0.0,
	}
	return videoStats, subscriberMetrics, viewAnalytics
}

func numOrZero(m map[string]interface{}, key string) int {
	if m == nil {
		return 0
	}
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}

func floatOrZero(m map[string]interface{}, key string) float64 {
	if m == nil {
		return 0
	}
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	default:
		return 0
	}
}

// GetAnalyticsOverviewHandler handles GET /admin/streaming/analytics/overview
func GetAnalyticsOverviewHandler(c *gin.Context, service *analyticsServices.AnalyticsService) {
	log.Println("📊 GetAnalyticsOverviewHandler: Fetching analytics overview...")

	period := c.DefaultQuery("period", "30d")

	overview, err := service.GetAnalytics(period)
	if err != nil {
		log.Printf("❌ Error fetching analytics overview: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch analytics overview",
		})
		return
	}

	videoStats, subscriberMetrics, viewAnalytics := metricsFromAnalytics(overview)
	overview["video_stats"] = videoStats
	overview["subscriber_metrics"] = subscriberMetrics
	overview["view_analytics"] = viewAnalytics
	overview["period"] = period

	log.Printf("✅ Analytics overview retrieved successfully for period: %s", period)
	c.JSON(http.StatusOK, overview)
}

// GetExecutiveSummaryHandler handles GET /admin/streaming/analytics/executive-summary
func GetExecutiveSummaryHandler(c *gin.Context, service *analyticsServices.AnalyticsService) {
	log.Println("📈 GetExecutiveSummaryHandler: Fetching executive summary...")

	period := c.DefaultQuery("period", "30d")

	overview, err := service.GetAnalytics(period)
	if err != nil {
		log.Printf("❌ Error fetching analytics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch executive summary"})
		return
	}
	videoStats, subscriberMetrics, _ := metricsFromAnalytics(overview)

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
func GetFunnelAnalysisHandler(c *gin.Context, service *analyticsServices.AnalyticsService) {
	log.Println("🔍 GetFunnelAnalysisHandler: Fetching funnel analysis...")

	period := c.DefaultQuery("period", "30d")

	overview, err := service.GetAnalytics(period)
	if err != nil {
		log.Printf("❌ Error fetching analytics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch funnel analysis"})
		return
	}
	_, subscriberMetrics, _ := metricsFromAnalytics(overview)

	funnel := gin.H{
		"period": period,
		"stages": []gin.H{
			{"name": "Visitors", "count": 0, "conversion": 0},
			{"name": "Sign-ups", "count": subscriberMetrics["total_subscribers"], "conversion": 0},
			{"name": "Trial Users", "count": 0, "conversion": 0},
			{"name": "Active Subscribers", "count": subscriberMetrics["active_subscriptions"], "conversion": 0},
		},
		"overall_conversion": 0,
		"drop_off_points":    []gin.H{},
	}

	log.Printf("✅ Funnel analysis retrieved successfully")
	c.JSON(http.StatusOK, funnel)
}

// GetRevenueImpactHandler handles GET /admin/streaming/analytics/revenue-impact
func GetRevenueImpactHandler(c *gin.Context, service *analyticsServices.AnalyticsService) {
	log.Println("💰 GetRevenueImpactHandler: Fetching revenue impact...")

	period := c.DefaultQuery("period", "30d")

	overview, err := service.GetAnalytics(period)
	if err != nil {
		log.Printf("❌ Error fetching analytics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch revenue impact"})
		return
	}
	_, subscriberMetrics, _ := metricsFromAnalytics(overview)

	revenueImpact := gin.H{
		"period": period,
		"revenue_by_source": []gin.H{
			{"source": "Subscriptions", "amount": subscriberMetrics["monthly_revenue"]},
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
func GetCustomerJourneyHandler(c *gin.Context, service *analyticsServices.AnalyticsService) {
	log.Println("🗺️ GetCustomerJourneyHandler: Fetching customer journey...")

	period := c.DefaultQuery("period", "30d")

	overview, err := service.GetAnalytics(period)
	if err != nil {
		log.Printf("❌ Error fetching analytics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customer journey"})
		return
	}
	_, subscriberMetrics, _ := metricsFromAnalytics(overview)

	journey := gin.H{
		"period": period,
		"lifecycle_stages": []gin.H{
			{"stage": "Prospect", "count": 0},
			{"stage": "Trial", "count": 0},
			{"stage": "Active", "count": subscriberMetrics["active_subscriptions"]},
			{"stage": "At Risk", "count": 0},
			{"stage": "Churned", "count": 0},
		},
		"avg_time_to_conversion": "0 days",
		"retention_rate":         0,
		"cohort_analysis":        []gin.H{},
	}

	log.Printf("✅ Customer journey retrieved successfully")
	c.JSON(http.StatusOK, journey)
}

// GetPromotionsAnalyticsHandler handles GET /admin/streaming/analytics/promotions
func GetPromotionsAnalyticsHandler(c *gin.Context, service *analyticsServices.AnalyticsService) {
	log.Println("🎁 GetPromotionsAnalyticsHandler: Fetching promotions analytics...")

	period := c.DefaultQuery("period", "30d")

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
	_ = service

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
