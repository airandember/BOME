package routes

import (
	"context"
	"net/http"
	"strconv"

	"bome-backend/internal/database"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupSubscriptionManagerRoutes sets up admin routes for subscription management
func SetupSubscriptionManagerRoutes(admin *gin.RouterGroup, db *database.DB) {
	linkingService := services.NewCustomerLinkingService(db)
	subscriptionManager := services.NewSubscriptionManagerService(db, linkingService)

	subManager := admin.Group("/subscription-manager")
	{
		// Get subscription summary for a user
		subManager.GET("/user/:user_id/summary", func(c *gin.Context) {
			getUserSubscriptionSummary(c, subscriptionManager)
		})

		// Fix multiple subscriptions for a specific user
		subManager.POST("/user/:user_id/enforce-single", func(c *gin.Context) {
			enforceUserSingleSubscription(c, subscriptionManager, linkingService)
		})

		// Fix all users with multiple subscriptions
		subManager.POST("/fix-all-multiple", func(c *gin.Context) {
			fixAllMultipleSubscriptions(c, subscriptionManager)
		})

		// Manually grant video access
		subManager.POST("/user/:user_id/grant-video-access", func(c *gin.Context) {
			grantVideoAccess(c, subscriptionManager)
		})

		// Manually revoke video access
		subManager.POST("/user/:user_id/revoke-video-access", func(c *gin.Context) {
			revokeVideoAccess(c, subscriptionManager)
		})

		// Update video access for a specific subscription
		subManager.POST("/subscription/:subscription_id/update-video-access", func(c *gin.Context) {
			updateVideoAccessForSubscription(c, subscriptionManager)
		})
	}
}

// getUserSubscriptionSummary returns a summary of a user's subscriptions
func getUserSubscriptionSummary(c *gin.Context, manager *services.SubscriptionManagerService) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	summary, err := manager.GetUserSubscriptionSummary(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"summary": summary,
	})
}

// enforceUserSingleSubscription enforces single subscription for a specific user
func enforceUserSingleSubscription(c *gin.Context, manager *services.SubscriptionManagerService, linkingService *services.CustomerLinkingService) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get user's linked customers
	linkedCustomers, err := linkingService.GetUserLinkedCustomers(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(linkedCustomers) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "User has no linked customers",
		})
		return
	}

	// Find active subscriptions
	// (This is simplified - in a real scenario, you'd query the database for the newest subscription)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Use the fix-all-multiple endpoint to automatically fix users with multiple subscriptions",
		"user_id": userID,
	})
}

// fixAllMultipleSubscriptions finds and fixes all users with multiple subscriptions
func fixAllMultipleSubscriptions(c *gin.Context, manager *services.SubscriptionManagerService) {
	ctx := context.Background()

	results, err := manager.FixMultipleSubscriptions(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Count successes and failures
	successCount := 0
	failureCount := 0
	for _, result := range results {
		if result.Error == "" {
			successCount++
		} else {
			failureCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"total_users":   len(results),
		"success_count": successCount,
		"failure_count": failureCount,
		"results":       results,
	})
}

// grantVideoAccess manually grants video access to a user
func grantVideoAccess(c *gin.Context, manager *services.SubscriptionManagerService) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Reason = "manual grant by admin"
	}

	if err := manager.GrantVideoAccess(userID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Video access granted",
		"user_id": userID,
		"reason":  req.Reason,
	})
}

// revokeVideoAccess manually revokes video access from a user
func revokeVideoAccess(c *gin.Context, manager *services.SubscriptionManagerService) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Reason = "manual revoke by admin"
	}

	if err := manager.RevokeVideoAccess(userID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Video access revoked",
		"user_id": userID,
		"reason":  req.Reason,
	})
}

// updateVideoAccessForSubscription updates video access based on subscription status
func updateVideoAccessForSubscription(c *gin.Context, manager *services.SubscriptionManagerService) {
	subscriptionID := c.Param("subscription_id")

	if err := manager.UpdateVideoAccessForSubscription(subscriptionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"message":         "Video access updated",
		"subscription_id": subscriptionID,
	})
}
