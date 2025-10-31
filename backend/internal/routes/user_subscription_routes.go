package routes

import (
	"context"
	"net/http"
	"time"

	"bome-backend/internal/database"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupUserSubscriptionRoutes sets up user-facing subscription routes
func SetupUserSubscriptionRoutes(userGroup *gin.RouterGroup, db *database.DB) {
	linkingService := services.NewCustomerLinkingService(db)
	userSubService := services.NewUserSubscriptionService(db, linkingService)

	subscriptions := userGroup.Group("/subscriptions")
	{
		// Get all user's subscriptions (active + canceled)
		subscriptions.GET("", func(c *gin.Context) {
			getUserSubscriptions(c, userSubService)
		})

		// Get subscription history
		subscriptions.GET("/history", func(c *gin.Context) {
			getUserSubscriptionHistory(c, userSubService)
		})

		// Cancel multiple subscriptions (bulk)
		subscriptions.POST("/cancel-multiple", func(c *gin.Context) {
			cancelMultipleSubscriptions(c, userSubService)
		})

		// Cancel a single subscription
		subscriptions.POST("/:subscription_id/cancel", func(c *gin.Context) {
			cancelSingleSubscription(c, userSubService)
		})

		// Change subscription plan (Phase 7.3)
		subscriptions.POST("/change-plan", func(c *gin.Context) {
			changeSubscriptionPlan(c, userSubService)
		})

		// Check if user can subscribe (Phase 7.3)
		subscriptions.GET("/can-subscribe", func(c *gin.Context) {
			canUserSubscribe(c, userSubService)
		})
	}
}

// getUserSubscriptions returns all subscriptions for the authenticated user
func getUserSubscriptions(c *gin.Context, service *services.UserSubscriptionService) {
	// Get user ID from auth context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	subscriptions, err := service.GetUserSubscriptions(userIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"subscriptions": subscriptions,
	})
}

// getUserSubscriptionHistory returns subscription history for the authenticated user
func getUserSubscriptionHistory(c *gin.Context, service *services.UserSubscriptionService) {
	// Get user ID from auth context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	history, err := service.GetSubscriptionHistory(userIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"history": history,
	})
}

// cancelMultipleSubscriptions cancels multiple subscriptions for the authenticated user
func cancelMultipleSubscriptions(c *gin.Context, service *services.UserSubscriptionService) {
	// Get user ID from auth context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse request body
	var req services.CancelMultipleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate request
	if len(req.SubscriptionIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No subscription IDs provided"})
		return
	}

	// Cancel subscriptions
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := service.CancelMultipleSubscriptions(ctx, userIDInt, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ================================================================
// PHASE 7.3: CHANGE SUBSCRIPTION PLAN
// ================================================================

// changeSubscriptionPlan changes the user's current subscription to a new plan
func changeSubscriptionPlan(c *gin.Context, service *services.UserSubscriptionService) {
	// Get user ID from auth context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse request body
	var req services.ChangeSubscriptionPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate request
	if req.NewPriceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new_price_id is required"})
		return
	}

	// Change plan
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := service.ChangeSubscriptionPlan(ctx, userIDInt, req)
	if err != nil {
		// Check for specific error messages
		errMsg := err.Error()
		if errMsg == "no active subscription found - user should create a new subscription instead" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "You don't have an active subscription. Please subscribe first.",
				"code":  "NO_ACTIVE_SUBSCRIPTION",
			})
			return
		}
		if errMsg == "user has multiple active subscriptions - please consolidate first using the admin tool" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "You have multiple active subscriptions. Please contact support to consolidate them.",
				"code":  "MULTIPLE_SUBSCRIPTIONS",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// canUserSubscribe checks if a user can create a new subscription
func canUserSubscribe(c *gin.Context, service *services.UserSubscriptionService) {
	// Get user ID from auth context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if user can subscribe
	canSubscribe, message, err := service.CanUserSubscribe(userIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"can_subscribe": canSubscribe,
		"message":       message,
	})
}

// cancelSingleSubscription cancels a single subscription for the authenticated user
func cancelSingleSubscription(c *gin.Context, service *services.UserSubscriptionService) {
	// Get user ID from auth context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get subscription ID from URL
	subscriptionID := c.Param("subscription_id")
	if subscriptionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subscription ID is required"})
		return
	}

	// Parse request body (optional reason)
	var req services.CancelSingleRequest
	_ = c.ShouldBindJSON(&req) // Optional body

	// Cancel subscription
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := service.CancelSingleSubscription(ctx, userIDInt, subscriptionID, req)
	if err != nil {
		// Check for permission error
		if err.Error() == "permission denied: user does not own subscription "+subscriptionID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to cancel this subscription"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
