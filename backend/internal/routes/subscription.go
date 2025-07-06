package routes

import (
	"net/http"

	"bome-backend/internal/database"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// CreateSubscriptionRequest represents a subscription creation payload
type CreateSubscriptionRequest struct {
	PriceID string `json:"price_id" binding:"required"`
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

		subscription, err := db.GetSubscriptionByUserID(userID)
		if err != nil {
			// User has no subscription
			c.JSON(http.StatusOK, gin.H{"subscription": nil})
			return
		}

		c.JSON(http.StatusOK, gin.H{"subscription": subscription})
	}
}

// CreateSubscriptionHandler handles creating a new subscription
func CreateSubscriptionHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req CreateSubscriptionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// For now, create a simple subscription record
		// In production, this would integrate with Stripe
		subscription, err := db.CreateSubscription(
			userID,
			"sub_"+req.PriceID, // Mock Stripe subscription ID
			req.PriceID,
			"active",
			nil, // currentPeriodStart
			nil, // currentPeriodEnd
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"subscription": subscription})
	}
}

// CancelSubscriptionHandler handles cancelling a subscription
func CancelSubscriptionHandler(db *database.DB) gin.HandlerFunc {
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

		// Update subscription in database
		if err := db.CancelSubscription(subscription.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Subscription cancelled successfully"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Development mode: return mock checkout URL
		if stripeService == nil {
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
