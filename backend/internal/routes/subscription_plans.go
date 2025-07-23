package routes

import (
	"net/http"
	"strconv"

	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupSubscriptionPlanRoutes sets up subscription plan routes
func SetupSubscriptionPlanRoutes(router *gin.Engine, db *database.DB, subscriptionPlanService *services.SubscriptionPlanService) {
	// Admin routes for subscription plan management
	admin := router.Group("/api/admin/subscription-plans")
	admin.Use(middleware.AuthRequired())
	admin.Use(middleware.AdminRequired())

	{
		// Create subscription plan
		admin.POST("/", func(c *gin.Context) {
			createSubscriptionPlan(c, subscriptionPlanService)
		})

		// Get all subscription plans with filters
		admin.GET("/", func(c *gin.Context) {
			getSubscriptionPlans(c, subscriptionPlanService)
		})

		// Get subscription plan by ID
		admin.GET("/:id", func(c *gin.Context) {
			getSubscriptionPlan(c, subscriptionPlanService)
		})

		// Update subscription plan
		admin.PUT("/:id", func(c *gin.Context) {
			updateSubscriptionPlan(c, subscriptionPlanService)
		})

		// Soft delete subscription plan
		admin.DELETE("/:id", func(c *gin.Context) {
			deleteSubscriptionPlan(c, subscriptionPlanService)
		})

		// Get subscription plan count
		admin.GET("/count", func(c *gin.Context) {
			getSubscriptionPlanCount(c, subscriptionPlanService)
		})
	}

	// Public routes for listing available plans
	public := router.Group("/api/subscription-plans")
	{
		// Get active subscription plans
		public.GET("/active", func(c *gin.Context) {
			getActiveSubscriptionPlans(c, subscriptionPlanService)
		})

		// Get promoted subscription plans
		public.GET("/promoted", func(c *gin.Context) {
			getPromotedSubscriptionPlans(c, subscriptionPlanService)
		})

		// Get subscription plan by ID (public)
		public.GET("/:id", func(c *gin.Context) {
			getSubscriptionPlanPublic(c, subscriptionPlanService)
		})
	}
}

// createSubscriptionPlan handles POST /api/admin/subscription-plans/
func createSubscriptionPlan(c *gin.Context, service *services.SubscriptionPlanService) {
	var req services.SubscriptionPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	plan, err := service.CreateSubscriptionPlan(&req, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription plan", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Subscription plan created successfully",
		"plan":    plan,
	})
}

// getSubscriptionPlans handles GET /api/admin/subscription-plans/
func getSubscriptionPlans(c *gin.Context, service *services.SubscriptionPlanService) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	isActiveStr := c.Query("is_active")
	isPromotedStr := c.Query("is_promoted")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	var isActive, isPromoted *bool
	if isActiveStr != "" {
		active := isActiveStr == "true"
		isActive = &active
	}
	if isPromotedStr != "" {
		promoted := isPromotedStr == "true"
		isPromoted = &promoted
	}

	plans, err := service.GetSubscriptionPlansWithFilters(limit, offset, isActive, isPromoted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription plans", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plans": plans,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// getSubscriptionPlan handles GET /api/admin/subscription-plans/:id
func getSubscriptionPlan(c *gin.Context, service *services.SubscriptionPlanService) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription plan ID"})
		return
	}

	plan, err := service.GetSubscriptionPlan(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription plan not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plan": plan})
}

// updateSubscriptionPlan handles PUT /api/admin/subscription-plans/:id
func updateSubscriptionPlan(c *gin.Context, service *services.SubscriptionPlanService) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription plan ID"})
		return
	}

	var req services.SubscriptionPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	plan, err := service.UpdateSubscriptionPlan(id, &req, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription plan", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Subscription plan updated successfully",
		"plan":    plan,
	})
}

// deleteSubscriptionPlan handles DELETE /api/admin/subscription-plans/:id
func deleteSubscriptionPlan(c *gin.Context, service *services.SubscriptionPlanService) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription plan ID"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	err = service.SoftDeleteSubscriptionPlan(id, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subscription plan", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription plan deleted successfully"})
}

// getSubscriptionPlanCount handles GET /api/admin/subscription-plans/count
func getSubscriptionPlanCount(c *gin.Context, service *services.SubscriptionPlanService) {
	isActiveStr := c.Query("is_active")
	var isActive *bool
	if isActiveStr != "" {
		active := isActiveStr == "true"
		isActive = &active
	}

	count, err := service.GetSubscriptionPlanCount(isActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription plan count", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// getActiveSubscriptionPlans handles GET /api/subscription-plans/active
func getActiveSubscriptionPlans(c *gin.Context, service *services.SubscriptionPlanService) {
	plans, err := service.GetActiveSubscriptionPlans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active subscription plans", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

// getPromotedSubscriptionPlans handles GET /api/subscription-plans/promoted
func getPromotedSubscriptionPlans(c *gin.Context, service *services.SubscriptionPlanService) {
	plans, err := service.GetPromotedSubscriptionPlans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get promoted subscription plans", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

// getSubscriptionPlanPublic handles GET /api/subscription-plans/:id (public)
func getSubscriptionPlanPublic(c *gin.Context, service *services.SubscriptionPlanService) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription plan ID"})
		return
	}

	plan, err := service.GetSubscriptionPlan(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription plan not found", "details": err.Error()})
		return
	}

	// Only return active plans for public access
	if !plan.IsActive {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription plan not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plan": plan})
}
