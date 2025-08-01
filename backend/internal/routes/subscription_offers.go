package routes

import (
	"net/http"
	"strconv"

	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupSubscriptionOfferRoutes sets up the subscription offer routes
func SetupSubscriptionOfferRoutes(router *gin.RouterGroup, db *database.DB, offersService *services.SubscriptionOffersService) {
	// Admin routes for subscription offer management
	admin := router.Group("/subscription-offers")
	admin.Use(middleware.AuthRequired())
	admin.Use(middleware.AdminRequired())

	{
		// GET /admin/subscription-offers - Get all offers
		admin.GET("/", func(c *gin.Context) {
			offers, err := offersService.GetAllSubscriptionOffers(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription offers"})
				return
			}
			c.JSON(http.StatusOK, offers)
		})

		// GET /admin/subscription-offers/:id - Get specific offer
		admin.GET("/:id", func(c *gin.Context) {
			idStr := c.Param("id")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offer ID"})
				return
			}

			offer, err := offersService.GetSubscriptionOfferByID(c.Request.Context(), id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Offer not found"})
				return
			}
			c.JSON(http.StatusOK, offer)
		})

		// POST /admin/subscription-offers - Create new offer
		admin.POST("/", func(c *gin.Context) {
			var req services.CreateSubscriptionOfferRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
				return
			}

			// Add user data to context
			if userData := c.GetHeader("X-User-Data"); userData != "" {
				c.Set("frontend_user_data", userData)
			}

			offer, err := offersService.CreateSubscriptionOffer(c.Request.Context(), &req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription offer"})
				return
			}
			c.JSON(http.StatusCreated, offer)
		})

		// PUT /admin/subscription-offers/:id - Update offer
		admin.PUT("/:id", func(c *gin.Context) {
			idStr := c.Param("id")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offer ID"})
				return
			}

			var req services.UpdateSubscriptionOfferRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
				return
			}
			req.ID = id

			// Add user data to context
			if userData := c.GetHeader("X-User-Data"); userData != "" {
				c.Set("frontend_user_data", userData)
			}

			offer, err := offersService.UpdateSubscriptionOffer(c.Request.Context(), &req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription offer"})
				return
			}
			c.JSON(http.StatusOK, offer)
		})

		// DELETE /admin/subscription-offers/:id - Delete offer
		admin.DELETE("/:id", func(c *gin.Context) {
			idStr := c.Param("id")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offer ID"})
				return
			}

			// Add user data to context
			if userData := c.GetHeader("X-User-Data"); userData != "" {
				c.Set("frontend_user_data", userData)
			}

			err = offersService.DeleteSubscriptionOffer(c.Request.Context(), id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subscription offer"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Subscription offer deleted successfully"})
		})

		// POST /admin/subscription-offers/:id/toggle - Toggle offer status
		admin.POST("/:id/toggle", func(c *gin.Context) {
			idStr := c.Param("id")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offer ID"})
				return
			}

			// Get current offer
			currentOffer, err := offersService.GetSubscriptionOfferByID(c.Request.Context(), id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Offer not found"})
				return
			}

			// Toggle status
			newStatus := !currentOffer.IsActive
			req := services.UpdateSubscriptionOfferRequest{
				ID:       id,
				IsActive: &newStatus,
			}

			// Add user data to context
			if userData := c.GetHeader("X-User-Data"); userData != "" {
				c.Set("frontend_user_data", userData)
			}

			updatedOffer, err := offersService.UpdateSubscriptionOffer(c.Request.Context(), &req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle offer status"})
				return
			}

			c.JSON(http.StatusOK, updatedOffer)
		})

		// GET /admin/subscription-offers/:id/history - Get offer history
		admin.GET("/:id/history", func(c *gin.Context) {
			idStr := c.Param("id")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offer ID"})
				return
			}

			history, err := offersService.GetOfferHistory(c.Request.Context(), id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get offer history"})
				return
			}

			c.JSON(http.StatusOK, history)
		})
	}
}
