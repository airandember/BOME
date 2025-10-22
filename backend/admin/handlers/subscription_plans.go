package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"bome-backend/infrastructure/database"
	"bome-backend/services/subscription"

	"github.com/gin-gonic/gin"
)

// SetupSubscriptionPlanRoutes sets up subscription plan routes
func SetupSubscriptionPlanRoutes(router *gin.RouterGroup, db *database.DB, wsHub WebSocketHub) {
	service := subscription.NewSubscriptionPlanService(db, wsHub)

	// Admin routes for subscription plan management
	plans := router.Group("/subscription-plans")
	{
		// GET /admin/subscription-plans - List all plans
		plans.GET("", func(c *gin.Context) {
			GetAllSubscriptionPlansHandler(c, service)
		})

		// GET /admin/subscription-plans/:id - Get plan by ID
		plans.GET("/:id", func(c *gin.Context) {
			GetSubscriptionPlanByIDHandler(c, service)
		})

		// POST /admin/subscription-plans - Create new plan
		plans.POST("", func(c *gin.Context) {
			CreateSubscriptionPlanHandler(c, service)
		})

		// PUT /admin/subscription-plans/:id - Update plan
		plans.PUT("/:id", func(c *gin.Context) {
			UpdateSubscriptionPlanHandler(c, service)
		})

		// DELETE /admin/subscription-plans/:id - Soft delete plan
		plans.DELETE("/:id", func(c *gin.Context) {
			DeleteSubscriptionPlanHandler(c, service)
		})

		// PUT /admin/subscription-plans/:id/status - Toggle active status
		plans.PUT("/:id/status", func(c *gin.Context) {
			ToggleSubscriptionPlanStatusHandler(c, service)
		})

		// PUT /admin/subscription-plans/:id/promotion - Update promotion dates
		plans.PUT("/:id/promotion", func(c *gin.Context) {
			UpdatePromotionStatusHandler(c, service)
		})
	}
}

// GetAllSubscriptionPlansHandler handles GET /admin/subscription-plans
func GetAllSubscriptionPlansHandler(c *gin.Context, service *subscription.SubscriptionPlanService) {
	log.Println("📋 GetAllSubscriptionPlansHandler called")

	plans, err := service.GetAllPlans()
	if err != nil {
		log.Printf("❌ Error getting plans: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get subscription plans",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Retrieved %d subscription plans", len(plans))
	c.JSON(http.StatusOK, plans)
}

// GetSubscriptionPlanByIDHandler handles GET /admin/subscription-plans/:id
func GetSubscriptionPlanByIDHandler(c *gin.Context, service *subscription.SubscriptionPlanService) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid plan ID",
		})
		return
	}

	plan, err := service.GetPlanByID(id)
	if err != nil {
		log.Printf("❌ Error getting plan %d: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Subscription plan not found",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Retrieved plan: %s (ID: %d)", plan.Name, plan.ID)
	c.JSON(http.StatusOK, plan)
}

// CreateSubscriptionPlanHandler handles POST /admin/subscription-plans
func CreateSubscriptionPlanHandler(c *gin.Context, service *subscription.SubscriptionPlanService) {
	log.Println("📝 CreateSubscriptionPlanHandler called")

	var req subscription.CreateSubscriptionPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Bad request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	plan, err := service.CreatePlan(c.Request.Context(), &req)
	if err != nil {
		log.Printf("❌ Error creating plan: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create subscription plan",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Created plan: %s (ID: %d)", plan.Name, plan.ID)
	c.JSON(http.StatusCreated, plan)
}

// UpdateSubscriptionPlanHandler handles PUT /admin/subscription-plans/:id
func UpdateSubscriptionPlanHandler(c *gin.Context, service *subscription.SubscriptionPlanService) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid plan ID",
		})
		return
	}

	var req subscription.UpdateSubscriptionPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Bad request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	plan, err := service.UpdatePlan(c.Request.Context(), id, &req)
	if err != nil {
		log.Printf("❌ Error updating plan %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update subscription plan",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Updated plan: %s (ID: %d)", plan.Name, plan.ID)
	c.JSON(http.StatusOK, plan)
}

// DeleteSubscriptionPlanHandler handles DELETE /admin/subscription-plans/:id
func DeleteSubscriptionPlanHandler(c *gin.Context, service *subscription.SubscriptionPlanService) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid plan ID",
		})
		return
	}

	err = service.DeletePlan(c.Request.Context(), id)
	if err != nil {
		log.Printf("❌ Error deleting plan %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete subscription plan",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Deleted plan ID: %d", id)
	c.JSON(http.StatusOK, gin.H{
		"message": "Subscription plan deleted successfully",
	})
}

// ToggleSubscriptionPlanStatusHandler handles PUT /admin/subscription-plans/:id/status
func ToggleSubscriptionPlanStatusHandler(c *gin.Context, service *subscription.SubscriptionPlanService) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid plan ID",
		})
		return
	}

	plan, err := service.TogglePlanStatus(c.Request.Context(), id)
	if err != nil {
		log.Printf("❌ Error toggling plan status %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to toggle plan status",
			"details": err.Error(),
		})
		return
	}

	status := "inactive"
	if plan.IsActive {
		status = "active"
	}
	log.Printf("✅ Toggled plan status: %s (ID: %d) -> %s", plan.Name, plan.ID, status)
	c.JSON(http.StatusOK, plan)
}

// UpdatePromotionStatusHandler handles PUT /admin/subscription-plans/:id/promotion
func UpdatePromotionStatusHandler(c *gin.Context, service *subscription.SubscriptionPlanService) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid plan ID",
		})
		return
	}

	var req struct {
		PromotionStartDate string `json:"promotion_start_date"`
		PromotionEndDate   string `json:"promotion_end_date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Bad request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Parse dates
	var startDate, endDate *time.Time
	if req.PromotionStartDate != "" {
		if date, err := time.Parse("2006-01-02", req.PromotionStartDate); err == nil {
			startDate = &date
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid promotion_start_date format. Use YYYY-MM-DD",
			})
			return
		}
	}
	if req.PromotionEndDate != "" {
		if date, err := time.Parse("2006-01-02", req.PromotionEndDate); err == nil {
			endDate = &date
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid promotion_end_date format. Use YYYY-MM-DD",
			})
			return
		}
	}

	plan, err := service.UpdatePromotionStatus(c.Request.Context(), id, startDate, endDate)
	if err != nil {
		log.Printf("❌ Error updating promotion status %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update promotion status",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Updated promotion status: %s (ID: %d)", plan.Name, plan.ID)
	c.JSON(http.StatusOK, plan)
}
