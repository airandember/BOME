package routes

import (
	"net/http"

	"bome-backend/internal/database"

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
