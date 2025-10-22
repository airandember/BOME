package handlers

import (
	"log"
	"net/http"
	"strconv"

	"bome-backend/infrastructure/database"
	"bome-backend/services/subscription"

	"github.com/gin-gonic/gin"
)

// SetupSubscriptionOfferRoutes sets up subscription offer routes
func SetupSubscriptionOfferRoutes(router *gin.RouterGroup, db *database.DB, wsHub WebSocketHub) {
	service := subscription.NewSubscriptionOffersService(db, wsHub)

	// Admin routes for subscription offer management
	offers := router.Group("/subscription-offers")
	{
		// GET /admin/subscription-offers - List all offers
		offers.GET("", func(c *gin.Context) {
			GetAllSubscriptionOffersHandler(c, service)
		})

		// GET /admin/subscription-offers/:id - Get offer by ID
		offers.GET("/:id", func(c *gin.Context) {
			GetSubscriptionOfferByIDHandler(c, service)
		})

		// POST /admin/subscription-offers - Create new offer
		offers.POST("", func(c *gin.Context) {
			CreateSubscriptionOfferHandler(c, service)
		})

		// PUT /admin/subscription-offers/:id - Update offer
		offers.PUT("/:id", func(c *gin.Context) {
			UpdateSubscriptionOfferHandler(c, service)
		})

		// DELETE /admin/subscription-offers/:id - Delete offer
		offers.DELETE("/:id", func(c *gin.Context) {
			DeleteSubscriptionOfferHandler(c, service)
		})

		// GET /admin/subscription-offers/plan/:plan_id/active - Get active offers for a plan
		offers.GET("/plan/:plan_id/active", func(c *gin.Context) {
			GetActiveOffersByPlanHandler(c, service)
		})
	}
}

// GetAllSubscriptionOffersHandler handles GET /admin/subscription-offers
func GetAllSubscriptionOffersHandler(c *gin.Context, service *subscription.SubscriptionOffersService) {
	log.Println("📋 GetAllSubscriptionOffersHandler called")

	offers, err := service.GetAllOffers()
	if err != nil {
		log.Printf("❌ Error getting offers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get subscription offers",
			"details": err.Error(),
		})
		return
	}

	// Always return an array, never null
	if offers == nil {
		offers = []*subscription.SubscriptionOffer{}
	}

	log.Printf("✅ Retrieved %d subscription offers", len(offers))
	c.JSON(http.StatusOK, offers)
}

// GetSubscriptionOfferByIDHandler handles GET /admin/subscription-offers/:id
func GetSubscriptionOfferByIDHandler(c *gin.Context, service *subscription.SubscriptionOffersService) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid offer ID",
		})
		return
	}

	offer, err := service.GetOfferByID(id)
	if err != nil {
		log.Printf("❌ Error getting offer %d: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Subscription offer not found",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Retrieved offer: %s (ID: %d)", offer.OffName, offer.ID)
	c.JSON(http.StatusOK, offer)
}

// CreateSubscriptionOfferHandler handles POST /admin/subscription-offers
func CreateSubscriptionOfferHandler(c *gin.Context, service *subscription.SubscriptionOffersService) {
	log.Println("📝 CreateSubscriptionOfferHandler called")

	var req subscription.CreateSubscriptionOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Bad request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	offer, err := service.CreateOffer(c.Request.Context(), &req)
	if err != nil {
		log.Printf("❌ Error creating offer: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create subscription offer",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Created offer: %s (ID: %d)", offer.OffName, offer.ID)
	c.JSON(http.StatusCreated, offer)
}

// UpdateSubscriptionOfferHandler handles PUT /admin/subscription-offers/:id
func UpdateSubscriptionOfferHandler(c *gin.Context, service *subscription.SubscriptionOffersService) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid offer ID",
		})
		return
	}

	var req subscription.UpdateSubscriptionOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Bad request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	offer, err := service.UpdateOffer(c.Request.Context(), id, &req)
	if err != nil {
		log.Printf("❌ Error updating offer %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update subscription offer",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Updated offer: %s (ID: %d)", offer.OffName, offer.ID)
	c.JSON(http.StatusOK, offer)
}

// DeleteSubscriptionOfferHandler handles DELETE /admin/subscription-offers/:id
func DeleteSubscriptionOfferHandler(c *gin.Context, service *subscription.SubscriptionOffersService) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid offer ID",
		})
		return
	}

	err = service.DeleteOffer(c.Request.Context(), id)
	if err != nil {
		log.Printf("❌ Error deleting offer %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete subscription offer",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Deleted offer ID: %d", id)
	c.JSON(http.StatusOK, gin.H{
		"message": "Subscription offer deleted successfully",
	})
}

// GetActiveOffersByPlanHandler handles GET /admin/subscription-offers/plan/:plan_id/active
func GetActiveOffersByPlanHandler(c *gin.Context, service *subscription.SubscriptionOffersService) {
	planIDStr := c.Param("plan_id")
	planID, err := strconv.Atoi(planIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid plan ID",
		})
		return
	}

	offers, err := service.GetActiveOffersByPlan(planID)
	if err != nil {
		log.Printf("❌ Error getting active offers for plan %d: %v", planID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get active offers",
			"details": err.Error(),
		})
		return
	}

	// Always return an array, never null
	if offers == nil {
		offers = []*subscription.SubscriptionOffer{}
	}

	log.Printf("✅ Retrieved %d active offers for plan %d", len(offers), planID)
	c.JSON(http.StatusOK, offers)
}
