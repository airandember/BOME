package routes

import (
	"net/http"
	"time"

	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"log"

	"database/sql"
	"strings"

	"github.com/gin-gonic/gin"
)

// SubscriptionPlanHandler handles subscription plan routes
type SubscriptionPlanHandler struct {
	service *services.SubscriptionPlanService
}

// NewSubscriptionPlanHandler creates a new subscription plan handler
func NewSubscriptionPlanHandler(service *services.SubscriptionPlanService) *SubscriptionPlanHandler {
	return &SubscriptionPlanHandler{service: service}
}

// SetupSubscriptionPlanRoutes sets up subscription plan routes
func SetupSubscriptionPlanRoutes(router *gin.RouterGroup, db *database.DB, subscriptionPlanService *services.SubscriptionPlanService) {
	handler := NewSubscriptionPlanHandler(subscriptionPlanService)

	// Admin routes for subscription plan management
	admin := router.Group("/subscription-plans")
	admin.Use(middleware.AuthRequired())
	admin.Use(middleware.AdminRequired())

	{
		// Create subscription plan
		admin.POST("/", handler.CreateSubscriptionPlanHandler)

		// Get all subscription plans
		admin.GET("/", handler.GetAllSubscriptionPlansHandler)

		// Get subscription plan by ID
		admin.GET("/:id", handler.GetSubscriptionPlanByIDHandler)

		// Update subscription plan
		admin.PUT("/:id", handler.UpdateSubscriptionPlanHandler)

		// Soft delete subscription plan
		admin.DELETE("/:id", handler.SoftDeleteSubscriptionPlanHandler)

		// Toggle subscription plan status
		admin.PUT("/:id/status", handler.ToggleSubscriptionPlanStatus)

		// Update promotion status
		admin.PUT("/:id/promotion", handler.UpdatePromotionStatusHandler)
	}

	// Public routes for listing available plans (need to be on main router)
	// These will be set up separately in the main routes file
}

// CreateSubscriptionPlanHandler handles POST /api/subscription_plans
// Creates a new subscription plan and returns the created object
func (h *SubscriptionPlanHandler) CreateSubscriptionPlanHandler(c *gin.Context) {
	log.Println("CreateSubscriptionPlanHandler called")
	var req services.CreateSubscriptionPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("CreateSubscriptionPlanHandler: bad request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert request to database model
	plan := &database.SubscriptionPlan{
		Name:             req.Name,
		Description:      req.Description,
		ShortDesc:        sql.NullString{String: req.ShortDesc, Valid: req.ShortDesc != ""},
		Price:            req.Price,
		Currency:         req.Currency,
		Interval:         req.Interval,
		IntervalCount:    req.IntervalCount,
		StripePriceID:    sql.NullString{String: req.StripePriceID, Valid: req.StripePriceID != ""},
		Features:         sql.NullString{String: strings.Join(req.Features, ","), Valid: len(req.Features) > 0},
		IsActive:         req.IsActive,
		IsPromoted:       req.IsPromoted,
		PromotionEndDate: sql.NullTime{Time: *req.PromotionEndDate, Valid: req.PromotionEndDate != nil},
		SubType:          100, // Default to standard plan
		SortOrder:        req.SortOrder,
		IsDeleted:        sql.NullBool{Bool: false, Valid: true},
	}

	createdPlan, err := h.service.CreateSubscriptionPlan(c.Request.Context(), plan)
	if err != nil {
		log.Printf("CreateSubscriptionPlanHandler: error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, createdPlan)
}

// GetAllSubscriptionPlansHandler handles GET /api/subscription_plans
// @route GET /api/subscription_plans
func (h *SubscriptionPlanHandler) GetAllSubscriptionPlansHandler(c *gin.Context) {
	log.Println("GetAllSubscriptionPlansHandler called")
	plans, err := h.service.GetAllSubscriptionPlans(c.Request.Context())
	if err != nil {
		log.Printf("GetAllSubscriptionPlansHandler: error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get subscription plans",
			"details": err.Error(),
			"status":  "error",
		})
		return
	}

	log.Printf("Retrieved %d subscription plans\n", len(plans))
	c.JSON(http.StatusOK, plans)
}

// GetSubscriptionPlanByIDHandler handles GET /api/subscription_plans/:id
// @route GET /api/subscription_plans/:id
func (h *SubscriptionPlanHandler) GetSubscriptionPlanByIDHandler(c *gin.Context) {
	log.Println("GetSubscriptionPlanByIDHandler called")
	id := c.Param("id")

	plan, err := h.service.GetSubscriptionPlan(c.Request.Context(), id)
	if err != nil {
		log.Printf("GetSubscriptionPlanByIDHandler: error: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Subscription plan not found",
			"details": err.Error(),
			"status":  "error",
		})
		return
	}

	log.Printf("Retrieved subscription plan by ID: %s\n", id)
	c.JSON(http.StatusOK, plan)
}

// UpdateSubscriptionPlanHandler handles PUT /api/subscription_plans/:id
// Updates a subscription plan and returns the updated object
func (h *SubscriptionPlanHandler) UpdateSubscriptionPlanHandler(c *gin.Context) {
	log.Println("UpdateSubscriptionPlanHandler called")
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		log.Printf("UpdateSubscriptionPlanHandler: bad request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	plan, err := h.service.UpdateSubscriptionPlan(c.Request.Context(), id, updates)
	if err != nil {
		log.Printf("UpdateSubscriptionPlanHandler: error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, plan)
}

// SoftDeleteSubscriptionPlanHandler handles DELETE /api/subscription_plans/:id
// Soft deletes a plan and returns success
func (h *SubscriptionPlanHandler) SoftDeleteSubscriptionPlanHandler(c *gin.Context) {
	log.Println("SoftDeleteSubscriptionPlanHandler called")
	id := c.Param("id")
	if err := h.service.SoftDeleteSubscriptionPlan(c.Request.Context(), id); err != nil {
		log.Printf("SoftDeleteSubscriptionPlanHandler: error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ToggleSubscriptionPlanStatus handles PUT /api/admin/subscription-plans/:id/status
func (h *SubscriptionPlanHandler) ToggleSubscriptionPlanStatus(c *gin.Context) {
	log.Println("ToggleSubscriptionPlanStatus called")
	id := c.Param("id")

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ToggleSubscriptionPlanStatus: bad request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
			"status":  "error",
		})
		return
	}

	plan, err := h.service.ToggleSubscriptionPlanStatus(c.Request.Context(), id, req.IsActive)
	if err != nil {
		log.Printf("ToggleSubscriptionPlanStatus: error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to toggle subscription plan status",
			"details": err.Error(),
			"status":  "error",
		})
		return
	}

	log.Printf("Subscription plan status toggled successfully: ID=%s, IsActive=%v\n", id, req.IsActive)
	c.JSON(http.StatusOK, plan)
}

// UpdatePromotionStatusHandler handles PUT /api/subscription_plans/:id/promotion
// Atomically updates promotion status and all relevant fields, returns updated plan
func (h *SubscriptionPlanHandler) UpdatePromotionStatusHandler(c *gin.Context) {
	log.Println("[ROUTE] UpdatePromotionStatusHandler called")
	id := c.Param("id")

	log.Printf("UpdatePromotionStatus route called for plan ID: %s\n", id)

	var req struct {
		IsPromoted       bool   `json:"is_promoted"`
		PromotionEndDate string `json:"promotion_end_date,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("UpdatePromotionStatusHandler: bad request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
			"status":  "error",
		})
		return
	}

	log.Printf("Request received: IsPromoted=%v, PromotionEndDate=%s\n", req.IsPromoted, req.PromotionEndDate)

	var promotionEndDate *time.Time
	if req.PromotionEndDate != "" {
		parsed, err := time.Parse(time.RFC3339, req.PromotionEndDate)
		if err != nil {
			log.Printf("UpdatePromotionStatusHandler: error parsing promotion end date: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid promotion end date format",
				"details": err.Error(),
				"status":  "error",
			})
			return
		}
		promotionEndDate = &parsed
	}

	plan, err := h.service.UpdatePromotionStatus(c.Request.Context(), id, req.IsPromoted, promotionEndDate)
	if err != nil {
		log.Printf("UpdatePromotionStatusHandler: error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update promotion status",
			"details": err.Error(),
			"status":  "error",
		})
		return
	}

	log.Printf("Plan updated successfully: ID=%s, IsPromoted=%v, IsActive=%v\n", plan.ID, plan.IsPromoted, plan.IsActive)
	c.JSON(http.StatusOK, plan)
}

// Public route handlers (for non-admin access)
func getActiveSubscriptionPlans(c *gin.Context, service *services.SubscriptionPlanService) {
	log.Println("getActiveSubscriptionPlans called")
	plans, err := service.GetAllSubscriptionPlans(c.Request.Context())
	if err != nil {
		log.Printf("getActiveSubscriptionPlans: error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get active subscription plans",
			"details": err.Error(),
			"status":  "error",
		})
		return
	}

	// Filter for active plans only
	var activePlans []*services.SubscriptionPlanResponse
	for _, plan := range plans {
		if plan.IsActive && !plan.IsPromoted {
			activePlans = append(activePlans, plan)
		}
	}

	log.Printf("Retrieved %d active subscription plans\n", len(activePlans))
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"subscription_plans": activePlans,
		},
		"message": "Active subscription plans retrieved successfully",
		"status":  "success",
	})
}

func getPromotedSubscriptionPlans(c *gin.Context, service *services.SubscriptionPlanService) {
	log.Println("getPromotedSubscriptionPlans called")
	plans, err := service.GetAllSubscriptionPlans(c.Request.Context())
	if err != nil {
		log.Printf("getPromotedSubscriptionPlans: error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get promoted subscription plans",
			"details": err.Error(),
			"status":  "error",
		})
		return
	}

	// Filter for promoted plans only
	var promotedPlans []*services.SubscriptionPlanResponse
	for _, plan := range plans {
		if plan.IsPromoted {
			promotedPlans = append(promotedPlans, plan)
		}
	}

	log.Printf("Retrieved %d promoted subscription plans\n", len(promotedPlans))
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"subscription_plans": promotedPlans,
		},
		"message": "Promoted subscription plans retrieved successfully",
		"status":  "success",
	})
}

func getSubscriptionPlanPublic(c *gin.Context, service *services.SubscriptionPlanService) {
	log.Println("getSubscriptionPlanPublic called")
	id := c.Param("id")

	plan, err := service.GetSubscriptionPlan(c.Request.Context(), id)
	if err != nil {
		log.Printf("getSubscriptionPlanPublic: error: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Subscription plan not found",
			"details": err.Error(),
			"status":  "error",
		})
		return
	}

	log.Printf("Retrieved subscription plan by ID: %s\n", id)
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"subscription_plan": plan,
		},
		"message": "Subscription plan retrieved successfully",
		"status":  "success",
	})
}
