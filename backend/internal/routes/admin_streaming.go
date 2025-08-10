package routes

import (
	"net/http"
	"strconv"
	"time"

	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupAdminStreamingRoutes sets up streaming admin dashboard routes
func SetupAdminStreamingRoutes(admin *gin.RouterGroup, db *database.DB, stripeService *services.StripeService, analyticsService *services.SubscriptionAnalyticsService, biService *services.BusinessIntelligenceService) {
	// Streaming admin routes - requires streaming manager role or higher
	streaming := admin.Group("/streaming")
	streaming.Use(middleware.AuthRequired())
	streaming.Use(middleware.StreamingAdminRequired())

	{
		// Dashboard overview
		streaming.GET("/dashboard", GetStreamingDashboardHandler(db, analyticsService))

		// Subscription management
		streaming.GET("/subscriptions", GetStreamingSubscriptionsHandler(db))
		streaming.GET("/subscriptions/:id", GetStreamingSubscriptionHandler(db))
		streaming.PUT("/subscriptions/:id", UpdateStreamingSubscriptionHandler(db, stripeService))
		streaming.DELETE("/subscriptions/:id", CancelStreamingSubscriptionHandler(db, stripeService))
		streaming.POST("/subscriptions/:id/refund", ProcessStreamingRefundHandler(db, stripeService))

		// Customer management
		streaming.GET("/customers", GetStreamingCustomersHandler(db))
		streaming.GET("/customers/:id", GetStreamingCustomerHandler(db))
		streaming.GET("/customers/:id/subscriptions", GetCustomerSubscriptionsHandler(db))
		streaming.POST("/customers/:id/communication", SendCustomerCommunicationHandler(db))

		// Analytics and reporting
		streaming.GET("/analytics/overview", GetStreamingAnalyticsOverviewHandler(analyticsService))
		streaming.GET("/analytics/revenue", GetStreamingRevenueAnalyticsHandler(analyticsService))
		streaming.GET("/analytics/subscriptions", GetStreamingSubscriptionAnalyticsHandler(analyticsService))
		streaming.GET("/analytics/customers", GetStreamingCustomerAnalyticsHandler(analyticsService))

		// Business Intelligence Analytics
		streaming.GET("/analytics/executive-summary", GetExecutiveSummaryHandler(biService))
		streaming.GET("/analytics/funnel-analysis", GetFunnelAnalysisHandler(biService))
		streaming.GET("/analytics/revenue-impact", GetRevenueImpactHandler(biService))
		streaming.GET("/analytics/customer-journey", GetCustomerJourneyHandler(biService))

		// Stripe configuration (write-only secret) and summary
		streaming.POST("/stripe/secret", func(c *gin.Context) {
			var req struct {
				Key  string `json:"key" binding:"required"`
				Type string `json:"type"` // optional: "sk" or "rk" or "pk"; we only accept secret keys for backend
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
				return
			}

			// Only allow server secret keys (sk_ or rk_)
			if !(len(req.Key) > 3 && (req.Key[:3] == "sk_" || req.Key[:3] == "rk_")) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "A Stripe secret or restricted secret starting with sk_ or rk_ is required"})
				return
			}

			crypto := services.GetGlobalCryptoService()
			if crypto == nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Crypto service not initialized"})
				return
			}

			encrypted, err := crypto.EncryptString(req.Key)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt key"})
				return
			}

			if db == nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
				return
			}

			if err := db.SetSecureSetting("stripe_secret_key", encrypted); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store key"})
				return
			}

			// Update runtime Stripe service to enable immediate use
			stripeService.UpdateSecretKey(req.Key)

			c.JSON(http.StatusOK, gin.H{"message": "Stripe secret updated"})
		})

		streaming.GET("/stripe/summary", func(c *gin.Context) {
			// never return stored secret; only return runtime capability summary
			summary, err := stripeService.GetAccountSummary()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"summary": summary})
		})

		// Promotions and deals
		streaming.GET("/promotions", GetStreamingPromotionsHandler(db))
		streaming.POST("/promotions", CreateStreamingPromotionHandler(db))
		streaming.PUT("/promotions/:id", UpdateStreamingPromotionHandler(db))
		streaming.DELETE("/promotions/:id", DeleteStreamingPromotionHandler(db))

		// Event-based deals
		streaming.GET("/events", GetStreamingEventsHandler(db))
		streaming.POST("/events", CreateStreamingEventHandler(db))
		streaming.PUT("/events/:id", UpdateStreamingEventHandler(db))
		streaming.DELETE("/events/:id", DeleteStreamingEventHandler(db))
		streaming.POST("/events/:id/subscription-deals", CreateEventSubscriptionDealHandler(db))
	}
}

// GetStreamingDashboardHandler handles streaming admin dashboard overview
func GetStreamingDashboardHandler(db *database.DB, analyticsService *services.SubscriptionAnalyticsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get key metrics for the dashboard
		metrics := map[string]interface{}{
			"timestamp": time.Now(),
		}

		// Get active subscriptions count
		if analyticsService != nil {
			activeCount, err := analyticsService.GetActiveSubscriptionsCount(nil)
			if err == nil {
				metrics["active_subscriptions"] = activeCount
			}

			// Get revenue metrics for last 30 days
			revenueMetrics, err := analyticsService.GetRevenueMetrics(
				time.Now().AddDate(0, 0, -30),
				time.Now(),
				nil,
			)
			if err == nil {
				metrics["revenue_30_days"] = revenueMetrics
			}

			// Get MRR
			mrr, err := analyticsService.CalculateMRR(time.Now(), nil)
			if err == nil {
				metrics["mrr"] = mrr
			}

			// Get churn rate
			churnData, err := analyticsService.CalculateChurnRate(
				time.Now().AddDate(0, 0, -30),
				time.Now(),
				nil,
			)
			if err == nil {
				metrics["churn_rate"] = churnData
			}

			// Get new subscriptions count
			newSubs, err := db.GetNewSubscriptionsCount(
				time.Now().AddDate(0, 0, -30),
				time.Now(),
			)
			if err == nil {
				metrics["new_subscriptions"] = newSubs
			}

			// Get total customers count
			totalCustomers, err := db.GetTotalCustomersCount()
			if err == nil {
				metrics["total_customers"] = totalCustomers
			}

			// Calculate average revenue per user
			if revenueMetrics != nil {
				if avgRev, ok := revenueMetrics["avg_revenue_per_user"]; ok {
					metrics["avg_revenue_per_user"] = avgRev
				}
			}
		}

		// Get recent subscription events
		// TODO: Implement recent events retrieval

		c.JSON(http.StatusOK, gin.H{
			"dashboard": gin.H{
				"metrics":       metrics,
				"recent_events": []interface{}{}, // TODO: Implement
			},
		})
	}
}

// GetStreamingSubscriptionsHandler handles retrieving subscriptions for streaming admin
func GetStreamingSubscriptionsHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters
		limitStr := c.DefaultQuery("limit", "20")
		offsetStr := c.DefaultQuery("offset", "0")
		status := c.Query("status")
		planID := c.Query("plan_id")
		dateFrom := c.Query("date_from")
		dateTo := c.Query("date_to")

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 20
		}

		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			offset = 0
		}

		// TODO: Implement proper filtering in database layer
		subscriptions, err := db.GetUserSubscriptionHistory(0, limit, offset) // Placeholder
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscriptions", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"subscriptions": subscriptions,
			"filters": gin.H{
				"status":    status,
				"plan_id":   planID,
				"date_from": dateFrom,
				"date_to":   dateTo,
			},
			"pagination": gin.H{
				"limit":  limit,
				"offset": offset,
			},
		})
	}
}

// GetStreamingSubscriptionHandler handles retrieving a specific subscription
func GetStreamingSubscriptionHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
			return
		}

		subscription, err := db.GetSubscriptionByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
			return
		}

		// Get subscription plan details
		var plan *database.SubscriptionPlan
		if subscription.PlanID.Valid {
			plan, err = db.GetSubscriptionPlanByID(int(subscription.PlanID.Int32))
			if err != nil {
				// Plan not found, but continue without it
				plan = nil
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"subscription": subscription,
			"plan":         plan,
		})
	}
}

// UpdateStreamingSubscriptionHandler handles updating a subscription
func UpdateStreamingSubscriptionHandler(db *database.DB, stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
			return
		}

		var req struct {
			Status            string `json:"status"`
			PlanID            int    `json:"plan_id"`
			CancelAtPeriodEnd bool   `json:"cancel_at_period_end"`
			Notes             string `json:"notes"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		// Update subscription in database
		updates := map[string]interface{}{
			"status": req.Status,
		}

		if req.PlanID > 0 {
			updates["plan_id"] = req.PlanID
		}

		_, err = db.UpdateSubscriptionPlan(id, updates)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription", "details": err.Error()})
			return
		}

		// Get updated subscription
		subscription, err := db.GetSubscriptionByID(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated subscription"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "Subscription updated successfully",
			"subscription": subscription,
		})
	}
}

// CancelStreamingSubscriptionHandler handles cancelling a subscription
func CancelStreamingSubscriptionHandler(db *database.DB, stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
			return
		}

		var req struct {
			Reason      string `json:"reason"`
			AtPeriodEnd bool   `json:"at_period_end"`
			AdminNotes  string `json:"admin_notes"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		subscription, err := db.GetSubscriptionByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
			return
		}

		// Cancel subscription in Stripe if enabled
		if stripeService != nil && stripeService.IsEnabled() {
			if err := stripeService.CancelSubscription(subscription.StripeSubscriptionID, req.AtPeriodEnd); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel subscription in Stripe", "details": err.Error()})
				return
			}
		}

		// Update subscription in database
		if err := db.CancelSubscription(id, req.Reason); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":                 "Subscription cancelled successfully",
			"cancelled_at_period_end": req.AtPeriodEnd,
		})
	}
}

// ProcessStreamingRefundHandler handles processing refunds
func ProcessStreamingRefundHandler(db *database.DB, stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
			return
		}

		var req struct {
			Amount     int64  `json:"amount" binding:"required,min=1"`
			Reason     string `json:"reason" binding:"required"`
			AdminNotes string `json:"admin_notes"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		subscription, err := db.GetSubscriptionByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
			return
		}

		// Process refund in Stripe if enabled
		var refund *services.Refund
		if stripeService != nil && stripeService.IsEnabled() {
			refund, err = stripeService.CreateRefund(subscription.StripeSubscriptionID, req.Amount, req.Reason)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process refund", "details": err.Error()})
				return
			}
		}

		// Update subscription with refund information
		if err := db.ProcessRefund(id, float64(req.Amount)/100, req.Reason); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription with refund", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Refund processed successfully",
			"refund":  refund,
		})
	}
}

// GetStreamingCustomersHandler handles retrieving customers for streaming admin
func GetStreamingCustomersHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters
		limitStr := c.DefaultQuery("limit", "20")
		offsetStr := c.DefaultQuery("offset", "0")
		hasSubscription := c.Query("has_subscription")
		status := c.Query("status")

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 20
		}

		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			offset = 0
		}

		// TODO: Implement customer retrieval with subscription filtering
		// For now, return placeholder data
		customers := []map[string]interface{}{
			{
				"id":                  1,
				"email":               "customer@example.com",
				"name":                "John Doe",
				"has_subscription":    true,
				"subscription_status": "active",
				"created_at":          time.Now().AddDate(0, -1, 0),
			},
		}

		c.JSON(http.StatusOK, gin.H{
			"customers": customers,
			"filters": gin.H{
				"has_subscription": hasSubscription,
				"status":           status,
			},
			"pagination": gin.H{
				"limit":  limit,
				"offset": offset,
			},
		})
	}
}

// GetStreamingCustomerHandler handles retrieving a specific customer
func GetStreamingCustomerHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
			return
		}

		// TODO: Implement customer retrieval
		customer := map[string]interface{}{
			"id":                 id,
			"email":              "customer@example.com",
			"name":               "John Doe",
			"created_at":         time.Now().AddDate(0, -1, 0),
			"subscription_count": 1,
		}

		c.JSON(http.StatusOK, gin.H{"customer": customer})
	}
}

// GetCustomerSubscriptionsHandler handles retrieving subscriptions for a customer
func GetCustomerSubscriptionsHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
			return
		}

		// Get subscription history for customer
		subscriptions, err := db.GetUserSubscriptionHistory(id, 50, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get customer subscriptions", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"subscriptions": subscriptions})
	}
}

// SendCustomerCommunicationHandler handles sending communications to customers
func SendCustomerCommunicationHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
			return
		}

		var req struct {
			Type    string `json:"type" binding:"required"` // email, sms, notification
			Subject string `json:"subject"`
			Message string `json:"message" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		// TODO: Implement customer communication
		// This would integrate with email/SMS services

		c.JSON(http.StatusOK, gin.H{
			"message":          "Communication sent successfully",
			"communication_id": "comm_" + strconv.Itoa(id),
		})
	}
}

// GetStreamingAnalyticsOverviewHandler handles streaming analytics overview
func GetStreamingAnalyticsOverviewHandler(analyticsService *services.SubscriptionAnalyticsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if analyticsService == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Analytics service not available"})
			return
		}

		// Parse date range
		startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
		endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format"})
			return
		}

		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format"})
			return
		}

		// Generate comprehensive analytics report
		report, err := analyticsService.GenerateSubscriptionReport(startDate, endDate, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate analytics report", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"analytics": report})
	}
}

// GetStreamingRevenueAnalyticsHandler handles revenue analytics
func GetStreamingRevenueAnalyticsHandler(analyticsService *services.SubscriptionAnalyticsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if analyticsService == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Analytics service not available"})
			return
		}

		// Get revenue metrics
		revenueMetrics, err := analyticsService.GetRevenueMetrics(
			time.Now().AddDate(0, -1, 0),
			time.Now(),
			nil,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get revenue metrics"})
			return
		}

		// Get MRR
		mrr, err := analyticsService.CalculateMRR(time.Now(), nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate MRR"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"revenue": revenueMetrics,
			"mrr":     mrr,
		})
	}
}

// GetStreamingSubscriptionAnalyticsHandler handles subscription analytics
func GetStreamingSubscriptionAnalyticsHandler(analyticsService *services.SubscriptionAnalyticsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if analyticsService == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Analytics service not available"})
			return
		}

		// Get active subscriptions count
		activeCount, err := analyticsService.GetActiveSubscriptionsCount(nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active subscriptions count"})
			return
		}

		// Get churn rate
		churnData, err := analyticsService.CalculateChurnRate(
			time.Now().AddDate(0, 0, -30),
			time.Now(),
			nil,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate churn rate"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"active_subscriptions": activeCount,
			"churn_rate":           churnData,
		})
	}
}

// GetStreamingCustomerAnalyticsHandler handles customer analytics
func GetStreamingCustomerAnalyticsHandler(analyticsService *services.SubscriptionAnalyticsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if analyticsService == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Analytics service not available"})
			return
		}

		// TODO: Implement customer-specific analytics
		// This would include customer acquisition, retention, etc.

		c.JSON(http.StatusOK, gin.H{
			"customer_analytics": map[string]interface{}{
				"total_customers": 0,
				"new_customers":   0,
				"retention_rate":  0.0,
			},
		})
	}
}

// GetExecutiveSummaryHandler handles retrieving the executive summary
func GetExecutiveSummaryHandler(biService *services.BusinessIntelligenceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if biService == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Business Intelligence service not available"})
			return
		}

		period := c.Query("period")
		if period == "" {
			period = "30d"
		}

		summary, err := biService.GetExecutiveSummary(period)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get executive summary", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"executive_summary": summary})
	}
}

// GetFunnelAnalysisHandler handles retrieving funnel analysis
func GetFunnelAnalysisHandler(biService *services.BusinessIntelligenceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if biService == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Business Intelligence service not available"})
			return
		}

		period := c.Query("period")
		if period == "" {
			period = "30d"
		}

		analysis, err := biService.GetFunnelAnalysis(period)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get funnel analysis", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"funnel_analysis": analysis})
	}
}

// GetRevenueImpactHandler handles retrieving revenue impact analysis
func GetRevenueImpactHandler(biService *services.BusinessIntelligenceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if biService == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Business Intelligence service not available"})
			return
		}

		period := c.Query("period")
		if period == "" {
			period = "30d"
		}

		impact, err := biService.GetRevenueImpact(period)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get revenue impact", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"revenue_impact": impact})
	}
}

// GetCustomerJourneyHandler handles retrieving customer journey analytics
func GetCustomerJourneyHandler(biService *services.BusinessIntelligenceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if biService == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Business Intelligence service not available"})
			return
		}

		period := c.Query("period")
		if period == "" {
			period = "30d"
		}

		journey, err := biService.GetCustomerJourney(period)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get customer journey", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"customer_journey": journey})
	}
}

// GetStreamingPromotionsHandler handles retrieving promotions
func GetStreamingPromotionsHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement promotions retrieval
		promotions := []map[string]interface{}{
			{
				"id":               1,
				"name":             "Summer Sale",
				"description":      "20% off all plans",
				"discount_percent": 20,
				"start_date":       time.Now(),
				"end_date":         time.Now().AddDate(0, 0, 30),
				"is_active":        true,
			},
		}

		c.JSON(http.StatusOK, gin.H{"promotions": promotions})
	}
}

// CreateStreamingPromotionHandler handles creating promotions
func CreateStreamingPromotionHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name            string    `json:"name" binding:"required"`
			Description     string    `json:"description"`
			DiscountPercent int       `json:"discount_percent" binding:"required,min=1,max=100"`
			StartDate       time.Time `json:"start_date" binding:"required"`
			EndDate         time.Time `json:"end_date" binding:"required"`
			IsActive        bool      `json:"is_active"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		// TODO: Implement promotion creation

		c.JSON(http.StatusCreated, gin.H{
			"message":      "Promotion created successfully",
			"promotion_id": 1,
		})
	}
}

// UpdateStreamingPromotionHandler handles updating promotions
func UpdateStreamingPromotionHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		_, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid promotion ID"})
			return
		}

		var req struct {
			Name            string    `json:"name"`
			Description     string    `json:"description"`
			DiscountPercent int       `json:"discount_percent"`
			StartDate       time.Time `json:"start_date"`
			EndDate         time.Time `json:"end_date"`
			IsActive        bool      `json:"is_active"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		// TODO: Implement promotion update

		c.JSON(http.StatusOK, gin.H{
			"message": "Promotion updated successfully",
		})
	}
}

// DeleteStreamingPromotionHandler handles deleting promotions
func DeleteStreamingPromotionHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		_, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid promotion ID"})
			return
		}

		// TODO: Implement promotion deletion

		c.JSON(http.StatusOK, gin.H{
			"message": "Promotion deleted successfully",
		})
	}
}

// GetStreamingEventsHandler handles retrieving events
func GetStreamingEventsHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement events retrieval
		events := []map[string]interface{}{
			{
				"id":          1,
				"name":        "Summer Festival",
				"description": "Annual summer streaming festival",
				"start_date":  time.Now().AddDate(0, 1, 0),
				"end_date":    time.Now().AddDate(0, 1, 7),
				"is_active":   true,
			},
		}

		c.JSON(http.StatusOK, gin.H{"events": events})
	}
}

// CreateStreamingEventHandler handles creating events
func CreateStreamingEventHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name        string    `json:"name" binding:"required"`
			Description string    `json:"description"`
			StartDate   time.Time `json:"start_date" binding:"required"`
			EndDate     time.Time `json:"end_date" binding:"required"`
			IsActive    bool      `json:"is_active"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		// TODO: Implement event creation

		c.JSON(http.StatusCreated, gin.H{
			"message":  "Event created successfully",
			"event_id": 1,
		})
	}
}

// UpdateStreamingEventHandler handles updating events
func UpdateStreamingEventHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		_, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
			return
		}

		var req struct {
			Name        string    `json:"name"`
			Description string    `json:"description"`
			StartDate   time.Time `json:"start_date"`
			EndDate     time.Time `json:"end_date"`
			IsActive    bool      `json:"is_active"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		// TODO: Implement event update

		c.JSON(http.StatusOK, gin.H{
			"message": "Event updated successfully",
		})
	}
}

// DeleteStreamingEventHandler handles deleting events
func DeleteStreamingEventHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		_, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
			return
		}

		// TODO: Implement event deletion

		c.JSON(http.StatusOK, gin.H{
			"message": "Event deleted successfully",
		})
	}
}

// CreateEventSubscriptionDealHandler handles creating subscription deals for events
func CreateEventSubscriptionDealHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		_, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
			return
		}

		var req struct {
			PlanID          int    `json:"plan_id" binding:"required"`
			DiscountPercent int    `json:"discount_percent" binding:"required,min=1,max=100"`
			Description     string `json:"description"`
			IsActive        bool   `json:"is_active"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		// TODO: Implement event subscription deal creation

		c.JSON(http.StatusCreated, gin.H{
			"message": "Event subscription deal created successfully",
			"deal_id": 1,
		})
	}
}
