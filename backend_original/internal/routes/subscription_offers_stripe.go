package routes

import (
	"context"
	"database/sql"
	"net/http"

	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupSubscriptionOffersStripeRoutes sets up Stripe-integrated subscription offer routes
func SetupSubscriptionOffersStripeRoutes(router *gin.RouterGroup, stripeService *services.StripeService, offerService *services.SubscriptionOffersStripeService) {
	stripe := router.Group("/subscription-offers/stripe")
	stripe.Use(middleware.AuthRequired())
	stripe.Use(middleware.AdminRequired())

	{
		// Create offer with automatic Stripe integration
		stripe.POST("/", createOfferWithStripeHandler(offerService))

		// Sync existing offer with Stripe
		stripe.POST("/:id/sync", syncOfferWithStripeHandler(offerService))

		// Get Stripe status for an offer
		stripe.GET("/:id/stripe-status", getOfferStripeStatusHandler(offerService))
	}
}

func createOfferWithStripeHandler(service *services.SubscriptionOffersStripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			PlanID             int     `json:"plan_id" binding:"required"`
			ItemID             *string `json:"item_id"`
			OffDiscountType    string  `json:"off_discount_type" binding:"required"`
			OffDiscountValue   float64 `json:"off_discount_value" binding:"required"`
			OfferStartDate     *string `json:"offer_start_date"`
			OffEndDate         *string `json:"off_end_date"`
			IsActive           bool    `json:"is_active"`
			OffDescription     *string `json:"off_description"`
			OffName            string  `json:"off_name" binding:"required"`
			OffCode            *string `json:"off_code"`
			OffMaxUses         *int    `json:"off_max_uses"`
			OffCurrentUses     int     `json:"off_current_uses"`
			OffTermsConditions *string `json:"off_terms_conditions"`
			OffTarget          *string `json:"off_target"`
			OffPriority        int     `json:"off_priority"`
			OffAutoApply       bool    `json:"off_auto_apply"`
			AutoCreateStripe   bool    `json:"auto_create_stripe"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get user data from header if available
		userDataHeader := c.GetHeader("X-User-Data")
		var ctx context.Context = c.Request.Context()

		if userDataHeader != "" {
			ctx = context.WithValue(c.Request.Context(), "frontend_user_data", userDataHeader)
		}

		// Convert to database model
		offer := &database.SubscriptionOffer{
			PlanID:           req.PlanID,
			OffDiscountType:  req.OffDiscountType,
			OffDiscountValue: req.OffDiscountValue,
			IsActive:         req.IsActive,
			OffName:          req.OffName,
			OffCurrentUses:   req.OffCurrentUses,
			OffPriority:      req.OffPriority,
			OffAutoApply:     req.OffAutoApply,
		}

		// Handle optional fields
		if req.ItemID != nil {
			offer.ItemID = sql.NullString{String: *req.ItemID, Valid: true}
		}
		if req.OfferStartDate != nil {
			if parsedDate, err := services.ParseFlexibleDate(*req.OfferStartDate); err == nil {
				offer.OfferStartDate = sql.NullTime{Time: *parsedDate, Valid: true}
			}
		}
		if req.OffEndDate != nil {
			if parsedDate, err := services.ParseFlexibleDate(*req.OffEndDate); err == nil {
				offer.OffEndDate = sql.NullTime{Time: *parsedDate, Valid: true}
			}
		}
		if req.OffDescription != nil {
			offer.OffDescription = sql.NullString{String: *req.OffDescription, Valid: true}
		}
		if req.OffCode != nil {
			offer.OffCode = sql.NullString{String: *req.OffCode, Valid: true}
		}
		if req.OffMaxUses != nil {
			offer.OffMaxUses = sql.NullInt32{Int32: int32(*req.OffMaxUses), Valid: true}
		}
		if req.OffTermsConditions != nil {
			offer.OffTermsConditions = sql.NullString{String: *req.OffTermsConditions, Valid: true}
		}
		if req.OffTarget != nil {
			offer.OffTarget = sql.NullString{String: *req.OffTarget, Valid: true}
		}

		// Create with Stripe integration
		response, err := service.CreateOfferWithStripe(ctx, offer, req.AutoCreateStripe)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, response)
	}
}

func syncOfferWithStripeHandler(service *services.SubscriptionOffersStripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		offerID := c.Param("id")

		response, err := service.SyncOfferWithStripe(c.Request.Context(), offerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Offer synced with Stripe successfully",
			"offer":   response,
		})
	}
}

func getOfferStripeStatusHandler(service *services.SubscriptionOffersStripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		offerID := c.Param("id")

		status, err := service.GetStripeIntegrationStatus(c.Request.Context(), offerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, status)
	}
}
