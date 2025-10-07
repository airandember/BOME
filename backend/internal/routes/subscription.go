package routes

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// CreateSubscriptionRequest represents a subscription creation payload
type CreateSubscriptionRequest struct {
	PlanID string `json:"plan_id" binding:"required"`
}

// CancelSubscriptionRequest represents a subscription cancellation payload
type CancelSubscriptionRequest struct {
	Reason      string `json:"reason"`
	AtPeriodEnd bool   `json:"at_period_end"`
}

// RefundRequest represents a refund request payload
type RefundRequest struct {
	Amount int64  `json:"amount" binding:"required,min=1"`
	Reason string `json:"reason" binding:"required"`
}

// UpdateSubscriptionRequest represents a subscription update payload
type UpdateSubscriptionRequest struct {
	PlanID string `json:"plan_id" binding:"required"`
}

// SetupSubscriptionRoutes sets up all subscription-related routes
func SetupSubscriptionRoutes(router *gin.Engine, db *database.DB, stripeService *services.StripeService, analyticsService *services.SubscriptionAnalyticsService) {
	// Customer subscription routes
	customer := router.Group("/api/subscriptions")
	customer.Use(middleware.AuthRequired())
	{
		// Get user's subscription
		customer.GET("/", GetSubscriptionHandler(db))

		// Create new subscription
		customer.POST("/", CreateSubscriptionHandler(db, stripeService))

		// Cancel subscription
		customer.DELETE("/", CancelSubscriptionHandler(db, stripeService))

		// Update subscription (change plan)
		customer.PUT("/", UpdateSubscriptionHandler(db, stripeService))

		// Get subscription history
		customer.GET("/history", GetSubscriptionHistoryHandler(db))

		// Get billing information
		customer.GET("/billing", GetBillingInfoHandler(db, stripeService))

		// Request refund
		customer.POST("/refund", RequestRefundHandler(db, stripeService))
	}

	// Admin subscription management routes
	admin := router.Group("/api/admin/subscriptions")
	admin.Use(middleware.AuthRequired())
	admin.Use(middleware.AdminRequired())
	{
		// Get all subscriptions with filters
		admin.GET("/", GetSubscriptionsHandler(db))

		// Get subscription by ID
		admin.GET("/:id", GetSubscriptionByIDHandler(db))

		// Update subscription (admin)
		admin.PUT("/:id", UpdateSubscriptionAdminHandler(db, stripeService))

		// Cancel subscription (admin)
		admin.DELETE("/:id", CancelSubscriptionAdminHandler(db, stripeService))

		// Process refund (admin)
		admin.POST("/:id/refund", ProcessRefundHandler(db, stripeService))

		// Get subscription analytics
		admin.GET("/analytics", GetSubscriptionAnalyticsHandler(analyticsService))

		// Get subscription metrics
		admin.GET("/metrics", GetSubscriptionMetricsHandler(analyticsService))
	}

	// Public routes
	public := router.Group("/api/subscription")
	{
		// Get available subscription plans
		public.GET("/plans", GetSubscriptionPlansHandler(stripeService))

		// Create checkout session
		public.POST("/checkout", CreateCheckoutSessionHandler(stripeService))

		// Webhook endpoint for Stripe events
		public.POST("/webhook", WebhookHandler(stripeService, analyticsService))
	}
}

// GetSubscriptionHandler handles retrieving user subscription
func GetSubscriptionHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Get user role from context
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
			return
		}

		// Admin roles include all roles with level 7+ (subsystem managers and above)
		adminRoles := []string{
			"super_admin",           // Level 10: Super Administrator
			"system_admin",          // Level 9: System Administrator
			"content_manager",       // Level 8: Content Manager
			"articles_manager",      // Level 7: Articles Manager
			"youtube_manager",       // Level 7: YouTube Manager
			"streaming_manager",     // Level 7: Video Streaming Manager
			"events_manager",        // Level 7: Events Manager
			"advertisement_manager", // Level 7: Advertisement Manager
			"user_manager",          // Level 7: User Account Manager
			"analytics_manager",     // Level 7: Analytics Manager
			"financial_admin",       // Level 7: Financial Administrator
			"admin",                 // Legacy admin role
		}

		// Check if user has admin role
		isAdmin := false
		for _, adminRole := range adminRoles {
			if userRole == adminRole {
				isAdmin = true
				break
			}
		}

		// Development mode: return mock subscription data
		if db == nil {
			// Admin users get premium access automatically
			if isAdmin {
				c.JSON(http.StatusOK, gin.H{
					"subscription": map[string]interface{}{
						"id":                 "admin_premium_access",
						"user_id":            userID,
						"plan_id":            "admin_premium",
						"status":             "active",
						"created_at":         "2024-01-01T00:00:00Z",
						"current_period_end": "2099-12-31T23:59:59Z",
					},
				})
				return
			}
			// Regular users have no subscription
			c.JSON(http.StatusOK, gin.H{"subscription": nil})
			return
		}

		// Production mode with database
		// Admin users get premium access automatically
		if isAdmin {
			c.JSON(http.StatusOK, gin.H{
				"subscription": map[string]interface{}{
					"id":                 "admin_premium_access",
					"user_id":            userID,
					"plan_id":            "admin_premium",
					"status":             "active",
					"created_at":         "2024-01-01T00:00:00Z",
					"current_period_end": "2099-12-31T23:59:59Z",
				},
			})
			return
		}

		// First check legacy subscriptions table
		subscription, err := db.GetSubscriptionByUserID(userID)
		if err == nil && subscription != nil {
			// User has legacy subscription
			c.JSON(http.StatusOK, gin.H{"subscription": subscription})
			return
		}

		// Check if user has active Stripe subscription with video access
		log.Printf("🔍 [GetSubscriptionHandler] Checking video access for user %d", userID)
		hasVideoAccess, accessInfo, err := db.HasVideoAccess(userID)

		// Enhanced logging for debugging
		if err != nil {
			log.Printf("❌ [GetSubscriptionHandler] Error checking video access for user %d: %v", userID, err)
		} else {
			log.Printf("🔍 [GetSubscriptionHandler] User %d video access check - HasAccess: %v, AccessInfo: Stripe=%v, Legacy=%v, Manual=%v, Source=%s",
				userID, hasVideoAccess, accessInfo.HasStripeAccess, accessInfo.HasLegacyAccess, accessInfo.HasManualAccess, accessInfo.AccessSource)
		}

		if err == nil && hasVideoAccess {
			// User has video access through Stripe subscription
			if accessInfo.HasStripeAccess {
				log.Printf("✅ [GetSubscriptionHandler] User %d granted Stripe video access", userID)
				// Create a subscription-like response for Stripe users
				c.JSON(http.StatusOK, gin.H{
					"subscription": map[string]interface{}{
						"id":                 "stripe_video_access",
						"user_id":            userID,
						"plan_id":            "stripe_subscription",
						"status":             "active",
						"tier":               "premium", // Grant premium access for Stripe subscription
						"access_source":      "stripe",
						"created_at":         "2024-01-01T00:00:00Z",
						"current_period_end": "2099-12-31T23:59:59Z",
					},
				})
				return
			}

			// User has manual video access
			if accessInfo.HasManualAccess {
				c.JSON(http.StatusOK, gin.H{
					"subscription": map[string]interface{}{
						"id":                 "manual_video_access",
						"user_id":            userID,
						"plan_id":            "manual_access",
						"status":             "active",
						"tier":               "premium", // Grant premium access for manual override
						"created_at":         "2024-01-01T00:00:00Z",
						"current_period_end": "2099-12-31T23:59:59Z",
						"access_source":      "manual",
					},
				})
				return
			}

			// User has legacy access
			if accessInfo.HasLegacyAccess {
				c.JSON(http.StatusOK, gin.H{
					"subscription": map[string]interface{}{
						"id":                 "legacy_video_access",
						"user_id":            userID,
						"plan_id":            "legacy_subscription",
						"status":             "active",
						"tier":               "premium", // Grant premium access for legacy subscription
						"created_at":         "2024-01-01T00:00:00Z",
						"current_period_end": "2099-12-31T23:59:59Z",
						"access_source":      "legacy",
					},
				})
				return
			}
		}

		// User has no subscription and no video access
		log.Printf("❌ [GetSubscriptionHandler] User %d has no video access - returning null subscription", userID)
		c.JSON(http.StatusOK, gin.H{"subscription": nil})
	}
}

// CreateSubscriptionHandler handles creating a new subscription
func CreateSubscriptionHandler(db *database.DB, stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req CreateSubscriptionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		// Check if user already has an active subscription
		existingSubscription, err := db.GetSubscriptionByUserID(userID)
		if err == nil && existingSubscription != nil && existingSubscription.Status == "active" {
			c.JSON(http.StatusConflict, gin.H{"error": "User already has an active subscription"})
			return
		}

		// Get the subscription plan
		plan, err := db.GetSubscriptionPlanByID(0) // TODO: Parse plan ID from req.PlanID
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription plan"})
			return
		}

		// Create subscription in database
		planIDInt := plan.ID
		subscription, err := db.CreateSubscription(
			userID,
			&planIDInt,
			"sub_"+req.PlanID, // Mock Stripe subscription ID
			req.PlanID,
			"active",
			nil, // currentPeriodStart
			nil, // currentPeriodEnd
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription", "details": err.Error()})
			return
		}

		// Track subscription event
		if analyticsService := getAnalyticsService(c); analyticsService != nil {
			analyticsService.TrackSubscriptionEvent(
				subscription.ID,
				userID,
				&plan.ID,
				"subscription_created",
				map[string]interface{}{
					"plan_id":   req.PlanID,
					"plan_name": plan.Name,
				},
				nil,
				c.ClientIP(),
				c.GetHeader("User-Agent"),
			)
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":      "Subscription created successfully",
			"subscription": subscription,
		})
	}
}

// CancelSubscriptionHandler handles cancelling a subscription
func CancelSubscriptionHandler(db *database.DB, stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req CancelSubscriptionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		subscription, err := db.GetSubscriptionByUserID(userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "No active subscription found"})
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
		if err := db.CancelSubscription(subscription.ID, req.Reason); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription", "details": err.Error()})
			return
		}

		// Track cancellation event
		if analyticsService := getAnalyticsService(c); analyticsService != nil {
			var planID *int
			if subscription.PlanID.Valid {
				planIDInt := int(subscription.PlanID.Int32)
				planID = &planIDInt
			}
			analyticsService.TrackSubscriptionEvent(
				subscription.ID,
				userID,
				planID,
				"subscription_cancelled",
				map[string]interface{}{
					"reason":        req.Reason,
					"at_period_end": req.AtPeriodEnd,
				},
				nil,
				c.ClientIP(),
				c.GetHeader("User-Agent"),
			)
		}

		c.JSON(http.StatusOK, gin.H{
			"message":                 "Subscription cancelled successfully",
			"cancelled_at_period_end": req.AtPeriodEnd,
		})
	}
}

// UpdateSubscriptionHandler handles updating a subscription (changing plan)
func UpdateSubscriptionHandler(db *database.DB, stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req UpdateSubscriptionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		subscription, err := db.GetSubscriptionByUserID(userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "No active subscription found"})
			return
		}

		// Get the new subscription plan
		plan, err := db.GetSubscriptionPlanByID(0) // TODO: Parse plan ID from req.PlanID
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription plan"})
			return
		}

		// Update subscription in database
		updates := map[string]interface{}{
			"plan_id": plan.ID,
		}
		_, err = db.UpdateSubscriptionPlan(subscription.ID, updates)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription", "details": err.Error()})
			return
		}

		// Track plan change event
		if analyticsService := getAnalyticsService(c); analyticsService != nil {
			analyticsService.TrackSubscriptionEvent(
				subscription.ID,
				userID,
				&plan.ID,
				"plan_changed",
				map[string]interface{}{
					"new_plan_id":   req.PlanID,
					"new_plan_name": plan.Name,
				},
				nil,
				c.ClientIP(),
				c.GetHeader("User-Agent"),
			)
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "Subscription updated successfully",
			"subscription": subscription,
		})
	}
}

// GetSubscriptionHistoryHandler handles retrieving subscription history
func GetSubscriptionHistoryHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Get query parameters for pagination
		limitStr := c.DefaultQuery("limit", "10")
		offsetStr := c.DefaultQuery("offset", "0")

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 10
		}

		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			offset = 0
		}

		// Get subscription history from database
		subscriptions, err := db.GetUserSubscriptionHistory(userID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription history", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"subscriptions": subscriptions,
			"pagination": gin.H{
				"limit":  limit,
				"offset": offset,
			},
		})
	}
}

// GetBillingInfoHandler handles retrieving billing information
func GetBillingInfoHandler(db *database.DB, stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		subscription, err := db.GetSubscriptionByUserID(userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "No active subscription found"})
			return
		}

		// Get billing information from Stripe if enabled
		var billingInfo map[string]interface{}
		if stripeService != nil && stripeService.IsEnabled() && subscription.StripeSubscriptionID != "" {
			invoices, _, err := stripeService.GetCustomerInvoices(subscription.StripeSubscriptionID, 5, "")
			if err == nil {
				billingInfo = map[string]interface{}{
					"invoices":        invoices,
					"subscription_id": subscription.StripeSubscriptionID,
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"billing_info": billingInfo,
			"subscription": subscription,
		})
	}
}

// RequestRefundHandler handles refund requests
func RequestRefundHandler(db *database.DB, stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req RefundRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		subscription, err := db.GetSubscriptionByUserID(userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "No active subscription found"})
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
		if err := db.ProcessRefund(subscription.ID, float64(req.Amount)/100, req.Reason); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription with refund", "details": err.Error()})
			return
		}

		// Track refund event
		if analyticsService := getAnalyticsService(c); analyticsService != nil {
			var planID *int
			if subscription.PlanID.Valid {
				planIDInt := int(subscription.PlanID.Int32)
				planID = &planIDInt
			}
			analyticsService.TrackSubscriptionEvent(
				subscription.ID,
				userID,
				planID,
				"refund_processed",
				map[string]interface{}{
					"amount": req.Amount,
					"reason": req.Reason,
				},
				nil,
				c.ClientIP(),
				c.GetHeader("User-Agent"),
			)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Refund processed successfully",
			"refund":  refund,
		})
	}
}

// GetSubscriptionsHandler handles retrieving all subscriptions (admin)
func GetSubscriptionsHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters
		limitStr := c.DefaultQuery("limit", "10")
		offsetStr := c.DefaultQuery("offset", "0")
		// TODO: Implement status and planID filtering
		// status := c.Query("status")
		// planID := c.Query("plan_id")

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 10
		}

		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			offset = 0
		}

		// Get subscriptions with filters (simplified for now)
		// TODO: Implement proper filtering in database layer
		subscriptions, err := db.GetUserSubscriptionHistory(0, limit, offset) // Placeholder
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscriptions", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"subscriptions": subscriptions,
			"pagination": gin.H{
				"limit":  limit,
				"offset": offset,
			},
		})
	}
}

// GetSubscriptionByIDHandler handles retrieving a specific subscription (admin)
func GetSubscriptionByIDHandler(db *database.DB) gin.HandlerFunc {
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

		c.JSON(http.StatusOK, gin.H{"subscription": subscription})
	}
}

// UpdateSubscriptionAdminHandler handles updating a subscription (admin)
func UpdateSubscriptionAdminHandler(db *database.DB, stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
			return
		}

		var req UpdateSubscriptionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		// Get the subscription plan
		plan, err := db.GetSubscriptionPlanByID(0) // TODO: Parse plan ID from req.PlanID
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription plan"})
			return
		}

		// Update subscription in database
		updates := map[string]interface{}{
			"plan_id": plan.ID,
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

// CancelSubscriptionAdminHandler handles cancelling a subscription (admin)
func CancelSubscriptionAdminHandler(db *database.DB, stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
			return
		}

		var req CancelSubscriptionRequest
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

// ProcessRefundHandler handles processing refunds (admin)
func ProcessRefundHandler(db *database.DB, stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
			return
		}

		var req RefundRequest
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

// GetSubscriptionAnalyticsHandler handles retrieving subscription analytics (admin)
func GetSubscriptionAnalyticsHandler(analyticsService *services.SubscriptionAnalyticsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if analyticsService == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Analytics service not available"})
			return
		}

		// Parse date range
		startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
		endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))
		planIDStr := c.Query("plan_id")

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

		var planID *int
		if planIDStr != "" {
			if id, err := strconv.Atoi(planIDStr); err == nil {
				planID = &id
			}
		}

		// Generate analytics report
		report, err := analyticsService.GenerateSubscriptionReport(startDate, endDate, planID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate analytics report", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"analytics": report})
	}
}

// GetSubscriptionMetricsHandler handles retrieving subscription metrics (admin)
func GetSubscriptionMetricsHandler(analyticsService *services.SubscriptionAnalyticsService) gin.HandlerFunc {
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

		c.JSON(http.StatusOK, gin.H{
			"metrics": map[string]interface{}{
				"active_subscriptions": activeCount,
				"revenue":              revenueMetrics,
			},
		})
	}
}

// GetSubscriptionPlansHandler handles retrieving available subscription plans
func GetSubscriptionPlansHandler(stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if stripeService == nil {
			// Mock subscription plans for development mode
			plans := []map[string]interface{}{
				{
					"id":          "price_basic_monthly",
					"name":        "Basic Monthly",
					"description": "Access to basic content",
					"price":       9.99,
					"currency":    "usd",
					"interval":    "month",
					"features":    []string{"Basic video access", "Standard quality", "Email support"},
				},
				{
					"id":          "price_premium_yearly",
					"name":        "Premium Yearly",
					"description": "Full access with exclusive content",
					"price":       99.99,
					"currency":    "usd",
					"interval":    "year",
					"features":    []string{"All video content", "HD quality", "Exclusive content", "Priority support"},
					"popular":     true,
				},
			}
			c.JSON(http.StatusOK, gin.H{"plans": plans})
			return
		}

		plans := stripeService.GetSubscriptionPlans()
		c.JSON(http.StatusOK, gin.H{"plans": plans})
	}
}

// CreateCheckoutSessionHandler handles creating Stripe checkout sessions
func CreateCheckoutSessionHandler(stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			PlanID     string `json:"planId" binding:"required"`
			SuccessURL string `json:"successUrl"`
			CancelURL  string `json:"cancelUrl"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		// Development mode: return mock checkout URL
		if stripeService == nil || !stripeService.IsEnabled() {
			c.JSON(http.StatusOK, gin.H{
				"url":        "https://checkout.stripe.com/mock-checkout-session",
				"session_id": "mock_session_" + req.PlanID,
			})
			return
		}

		// TODO: Implement actual Stripe checkout session creation
		c.JSON(http.StatusOK, gin.H{
			"url":        "https://checkout.stripe.com/pay/mock-session",
			"session_id": "mock_session_" + req.PlanID,
		})
	}
}

// WebhookHandler handles Stripe webhook events
func WebhookHandler(stripeService *services.StripeService, analyticsService *services.SubscriptionAnalyticsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if stripeService == nil || !stripeService.IsEnabled() {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Stripe service not available"})
			return
		}

		// Read the request body
		payload, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}

		// Get the signature from headers
		signature := c.GetHeader("Stripe-Signature")
		if signature == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing Stripe signature"})
			return
		}

		// Validate webhook signature
		event, err := stripeService.ValidateWebhookSignature(payload, signature)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook signature"})
			return
		}

		// Track webhook event
		if analyticsService != nil {
			analyticsService.TrackWebhookEvent(
				event.ID,
				event.Type,
				nil, // subscription_id
				nil, // user_id
				"completed",
				nil, // processing_time
				nil, // error_message
				nil, // payload_hash
			)
		}

		// Process the webhook event
		if err := stripeService.ProcessWebhook(event); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process webhook"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

// Helper function to get analytics service from context
func getAnalyticsService(c *gin.Context) *services.SubscriptionAnalyticsService {
	if analyticsService, exists := c.Get("analytics_service"); exists {
		if service, ok := analyticsService.(*services.SubscriptionAnalyticsService); ok {
			return service
		}
	}
	return nil
}
