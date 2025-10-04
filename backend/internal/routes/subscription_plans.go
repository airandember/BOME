package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"log"

	"database/sql"

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

		// Debug endpoint to check raw database data
		admin.GET("/:id/debug", handler.DebugSubscriptionPlanHandler)

		// Update subscription plan
		admin.PUT("/:id", handler.UpdateSubscriptionPlanHandler)

		// Soft delete subscription plan
		admin.DELETE("/:id", handler.SoftDeleteSubscriptionPlanHandler)

		// Toggle subscription plan status
		admin.PUT("/:id/status", handler.ToggleSubscriptionPlanStatus)

		// Update promotion status
		admin.PUT("/:id/promotion", handler.UpdatePromotionStatusHandler)

		// Analytics endpoints for history
		admin.GET("/:id/history", handler.GetPlanHistoryHandler)
		admin.GET("/analytics/history-stats", handler.GetHistoryStatsHandler)
		admin.GET("/analytics/history-by-type/:eventType", handler.GetHistoryByTypeHandler)
		admin.GET("/analytics/history-by-user/:userID", handler.GetHistoryByUserHandler)
		admin.GET("/analytics/history-by-date-range", handler.GetHistoryByDateRangeHandler)
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

	// Get user data from header if available
	userDataHeader := c.GetHeader("X-User-Data")
	var ctx context.Context = c.Request.Context()

	if userDataHeader != "" {
		log.Printf("CreateSubscriptionPlanHandler: Found user data in header: %s", userDataHeader)
		// Create a new context with user data
		ctx = context.WithValue(c.Request.Context(), "frontend_user_data", userDataHeader)
	} else {
		log.Printf("CreateSubscriptionPlanHandler: No user data in header, using context from middleware")
	}

	// Convert request to database model
	plan := &database.SubscriptionPlan{
		Name:          req.Name,
		Description:   req.Description,
		ShortDesc:     sql.NullString{String: req.ShortDesc, Valid: req.ShortDesc != ""},
		Price:         req.Price,
		Currency:      req.Currency,
		Interval:      req.Interval,
		IntervalCount: req.IntervalCount,
		StripePriceID: sql.NullString{String: req.StripePriceID, Valid: req.StripePriceID != ""},
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
			log.Printf("CreateSubscriptionPlanHandler: failed to marshal features: %v", err)
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

	createdPlan, err := h.service.CreateSubscriptionPlan(ctx, plan)
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
	log.Println("🎯 BEEP BOOP BEEP - GetAllSubscriptionPlansHandler called")
	log.Printf("🎯 BEEP BOOP BEEP - Time: %s", time.Now().Format("2006-01-02 15:04:05"))

	plans, err := h.service.GetAllSubscriptionPlans(c.Request.Context())
	if err != nil {
		log.Printf("🎯 BEEP BOOP BEEP - ERROR in GetAllSubscriptionPlansHandler: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get subscription plans",
			"details": err.Error(),
			"status":  "error",
			"debug":   "BEEP BOOP BEEP - Backend is running new code!",
		})
		return
	}

	log.Printf("🎯 BEEP BOOP BEEP - Retrieved %d subscription plans successfully", len(plans))
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

// DebugSubscriptionPlanHandler handles GET /api/subscription_plans/:id/debug
// Returns raw database data for debugging
func (h *SubscriptionPlanHandler) DebugSubscriptionPlanHandler(c *gin.Context) {
	log.Println("DebugSubscriptionPlanHandler called")
	id := c.Param("id")

	// Get raw database data through service
	plan, err := h.service.GetSubscriptionPlan(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
		return
	}

	// Return processed data for debugging
	c.JSON(http.StatusOK, gin.H{
		"id":                  plan.ID,
		"name":                plan.Name,
		"plan_change_history": plan.PlanChangeHistory,
		"promotion_metadata":  plan.PromotionMetadata,
	})
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

	// Debug: Log all headers
	log.Printf("UpdateSubscriptionPlanHandler: All headers: %+v", c.Request.Header)

	// Get user data from header if available
	userDataHeader := c.GetHeader("X-User-Data")
	var ctx context.Context = c.Request.Context()

	if userDataHeader != "" {
		log.Printf("UpdateSubscriptionPlanHandler: Found user data in header: %s", userDataHeader)
		// Create a new context with user data
		ctx = context.WithValue(c.Request.Context(), "frontend_user_data", userDataHeader)
		log.Printf("UpdateSubscriptionPlanHandler: Created context with frontend_user_data")
	} else {
		log.Printf("UpdateSubscriptionPlanHandler: No user data in header, using context from middleware")
	}

	plan, err := h.service.UpdateSubscriptionPlan(ctx, id, updates)
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

	// Get user data from header if available
	userDataHeader := c.GetHeader("X-User-Data")
	var ctx context.Context = c.Request.Context()

	if userDataHeader != "" {
		log.Printf("ToggleSubscriptionPlanStatus: Found user data in header: %s", userDataHeader)
		// Create a new context with user data
		ctx = context.WithValue(c.Request.Context(), "frontend_user_data", userDataHeader)
		log.Printf("ToggleSubscriptionPlanStatus: Created context with frontend_user_data")
	} else {
		log.Printf("ToggleSubscriptionPlanStatus: No user data in header, using context from middleware")
	}

	plan, err := h.service.ToggleSubscriptionPlanStatus(ctx, id, req.IsActive)
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

// UpdatePromotionStatusRequest represents a request to update promotion status
type UpdatePromotionStatusRequest struct {
	IsPromoted       bool   `json:"is_promoted"`
	PromotionEndDate string `json:"promotion_end_date,omitempty"`
}

// UpdatePromotionStatusHandler handles PUT /api/subscription_plans/:id/promotion
func (h *SubscriptionPlanHandler) UpdatePromotionStatusHandler(c *gin.Context) {
	log.Println("UpdatePromotionStatusHandler called")
	id := c.Param("id")

	var req UpdatePromotionStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("UpdatePromotionStatusHandler: bad request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Request received: IsPromoted=%v, PromotionEndDate=%s\n", req.IsPromoted, req.PromotionEndDate)

	// Pass the original context with user information from middleware
	ctx := c.Request.Context()

	// Parse promotion end date if provided
	var promotionEndDate *time.Time
	if req.PromotionEndDate != "" {
		parsedDate, err := time.Parse("2006-01-02", req.PromotionEndDate)
		if err != nil {
			log.Printf("UpdatePromotionStatusHandler: invalid date format: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
			return
		}
		promotionEndDate = &parsedDate
	}

	plan, err := h.service.UpdatePromotionStatus(ctx, id, req.IsPromoted, promotionEndDate)
	if err != nil {
		log.Printf("UpdatePromotionStatusHandler: error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Plan updated successfully: ID=%s, SubType=%s, IsActive=%v\n", plan.ID, plan.SubType, plan.IsActive)
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
		if plan.IsActive && plan.SubType == "stnd" {
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
		if plan.IsActive && plan.SubType == "prmo" {
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

// New comprehensive endpoint for all subscription data
func getAllSubscriptionData(c *gin.Context, service *services.SubscriptionPlanService, offersService *services.SubscriptionOffersService) {
	log.Println("getAllSubscriptionData called")

	// Get all plans
	plans, err := service.GetAllSubscriptionPlans(c.Request.Context())
	if err != nil {
		log.Printf("getAllSubscriptionData: error getting plans: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get subscription data",
			"details": err.Error(),
			"status":  "error",
		})
		return
	}

	// Get all offers
	offers, err := offersService.GetAllSubscriptionOffers(c.Request.Context())
	if err != nil {
		log.Printf("getAllSubscriptionData: error getting offers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get subscription data",
			"details": err.Error(),
			"status":  "error",
		})
		return
	}

	// Filter plans by type
	var standardPlans []*services.SubscriptionPlanResponse
	var promotionalPlans []*services.SubscriptionPlanResponse

	// Ensure plans is not nil
	if plans != nil {
		for _, plan := range plans {
			if plan.IsActive {
				if plan.SubType == "stnd" {
					standardPlans = append(standardPlans, plan)
				} else if plan.SubType == "prmo" {
					promotionalPlans = append(promotionalPlans, plan)
				}
			}
		}
	}

	// Filter active offers (exclude plan_id = 0 offers)
	var activeOffers []*services.SubscriptionOfferResponse
	// Ensure offers is not nil
	if offers != nil {
		for _, offer := range offers {
			if offer.IsActive && offer.PlanID != 0 {
				activeOffers = append(activeOffers, offer)
			}
		}
	}

	// Ensure arrays are never nil
	if standardPlans == nil {
		standardPlans = []*services.SubscriptionPlanResponse{}
	}
	if promotionalPlans == nil {
		promotionalPlans = []*services.SubscriptionPlanResponse{}
	}
	if activeOffers == nil {
		activeOffers = []*services.SubscriptionOfferResponse{}
	}

	log.Printf("Retrieved %d standard plans, %d promotional plans, and %d active offers\n",
		len(standardPlans), len(promotionalPlans), len(activeOffers))

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"standard_plans":    standardPlans,
			"promotional_plans": promotionalPlans,
			"offers":            activeOffers,
		},
		"message": "Subscription data retrieved successfully",
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

// Analytics handlers for history system

// GetPlanHistoryHandler handles GET /api/admin/subscription-plans/:id/history
func (h *SubscriptionPlanHandler) GetPlanHistoryHandler(c *gin.Context) {
	log.Println("GetPlanHistoryHandler called")
	id := c.Param("id")

	planID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan ID"})
		return
	}

	historyEvents, err := h.service.GetHistoryService().GetPlanHistory(c.Request.Context(), planID)
	if err != nil {
		log.Printf("GetPlanHistoryHandler: error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plan_id": id,
		"history": historyEvents,
		"count":   len(historyEvents),
	})
}

// GetHistoryStatsHandler handles GET /api/admin/subscription-plans/analytics/history-stats
func (h *SubscriptionPlanHandler) GetHistoryStatsHandler(c *gin.Context) {
	log.Println("GetHistoryStatsHandler called")

	stats, err := h.service.GetDatabase().GetHistoryStats()
	if err != nil {
		log.Printf("GetHistoryStatsHandler: error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// GetHistoryByTypeHandler handles GET /api/admin/subscription-plans/analytics/history-by-type/:eventType
func (h *SubscriptionPlanHandler) GetHistoryByTypeHandler(c *gin.Context) {
	log.Println("GetHistoryByTypeHandler called")
	eventType := c.Param("eventType")

	// Get limit from query params, default to 100
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	events, err := h.service.GetDatabase().GetHistoryEventsByType(eventType, limit)
	if err != nil {
		log.Printf("GetHistoryByTypeHandler: error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"event_type": eventType,
		"events":     events,
		"count":      len(events),
	})
}

// GetHistoryByUserHandler handles GET /api/admin/subscription-plans/analytics/history-by-user/:userID
func (h *SubscriptionPlanHandler) GetHistoryByUserHandler(c *gin.Context) {
	log.Println("GetHistoryByUserHandler called")
	userID := c.Param("userID")

	// Get limit from query params, default to 100
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	events, err := h.service.GetDatabase().GetHistoryEventsByUser(userID, limit)
	if err != nil {
		log.Printf("GetHistoryByUserHandler: error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"events":  events,
		"count":   len(events),
	})
}

// GetHistoryByDateRangeHandler handles GET /api/admin/subscription-plans/analytics/history-by-date-range
func (h *SubscriptionPlanHandler) GetHistoryByDateRangeHandler(c *gin.Context) {
	log.Println("GetHistoryByDateRangeHandler called")

	// Get date range from query params
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format (use YYYY-MM-DD)"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format (use YYYY-MM-DD)"})
		return
	}

	// Get limit from query params, default to 100
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	events, err := h.service.GetDatabase().GetHistoryEventsByDateRange(startDate, endDate, limit)
	if err != nil {
		log.Printf("GetHistoryByDateRangeHandler: error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"start_date": startDate,
		"end_date":   endDate,
		"events":     events,
		"count":      len(events),
	})
}
