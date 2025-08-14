package routes

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupSubscriptionPlanStripeRoutes sets up Stripe-integrated subscription plan routes
func SetupSubscriptionPlanStripeRoutes(router *gin.RouterGroup, stripeService *services.StripeService, planService *services.SubscriptionPlanStripeService) {
	fmt.Printf("Setting up Stripe subscription plan routes...\n")
	stripe := router.Group("/subscription-plans/stripe")
	stripe.Use(middleware.AuthRequired())
	stripe.Use(middleware.AdminRequired())

	{
		// Create plan with automatic Stripe integration
		stripe.POST("/", createPlanWithStripeHandler(planService))
		fmt.Printf("Registered: POST /subscription-plans/stripe/\n")

		// Sync existing plan with Stripe
		stripe.POST("/:id/sync", syncPlanWithStripeHandler(planService))
		fmt.Printf("Registered: POST /subscription-plans/stripe/:id/sync\n")

		// Get Stripe status for a plan
		stripe.GET("/:id/stripe-status", getStripeStatusHandler(planService))
		fmt.Printf("Registered: GET /subscription-plans/stripe/:id/stripe-status\n")
	}
	fmt.Printf("Stripe subscription plan routes setup complete.\n")
}

func createPlanWithStripeHandler(service *services.SubscriptionPlanStripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name               string   `json:"name" binding:"required"`
			Description        string   `json:"description"`
			ShortDesc          string   `json:"short_desc"`
			Price              float64  `json:"price" binding:"required"`
			Currency           string   `json:"currency" binding:"required"`
			Interval           string   `json:"interval" binding:"required"`
			IntervalCount      int      `json:"interval_count" binding:"required"`
			Features           []string `json:"features"`
			IsActive           bool     `json:"is_active"`
			PromotionStartDate string   `json:"promotion_start_date"`
			PromotionEndDate   string   `json:"promotion_end_date"`
			SubType            string   `json:"sub_type"`
			AutoCreateStripe   bool     `json:"auto_create_stripe"`
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
		plan := &database.SubscriptionPlan{
			Name:          req.Name,
			Description:   req.Description,
			ShortDesc:     sql.NullString{String: req.ShortDesc, Valid: req.ShortDesc != ""},
			Price:         req.Price,
			Currency:      req.Currency,
			Interval:      req.Interval,
			IntervalCount: req.IntervalCount,
			Features:      sql.NullString{String: "[]", Valid: true}, // Will be updated below
			IsActive:      req.IsActive,
			SubType:       req.SubType,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		// Handle features JSON marshaling
		if len(req.Features) > 0 {
			featuresJSON, err := json.Marshal(req.Features)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process features"})
				return
			}
			plan.Features = sql.NullString{String: string(featuresJSON), Valid: true}
		}

		// Handle promotion dates
		if req.PromotionStartDate != "" {
			if startDate, err := services.ParseFlexibleDate(req.PromotionStartDate); err == nil {
				plan.PromotionStartDate = services.FormatDateForDatabase(startDate, false)
			}
		}

		if req.PromotionEndDate != "" {
			if endDate, err := services.ParseFlexibleDate(req.PromotionEndDate); err == nil {
				plan.PromotionEndDate = services.FormatDateForDatabase(endDate, true)
			}
		}

		// Create with Stripe integration
		response, err := service.CreateSubscriptionPlanWithStripe(ctx, plan, req.AutoCreateStripe)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, response)
	}
}

func syncPlanWithStripeHandler(service *services.SubscriptionPlanStripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		planID := c.Param("id")

		response, err := service.SyncWithStripe(c.Request.Context(), planID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Plan synced with Stripe successfully",
			"plan":    response,
		})
	}
}

func getStripeStatusHandler(service *services.SubscriptionPlanStripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		planID := c.Param("id")

		status, err := service.GetStripeIntegrationStatus(c.Request.Context(), planID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, status)
	}
}
