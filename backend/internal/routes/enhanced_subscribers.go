package routes

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupEnhancedSubscriberRoutes sets up enhanced subscriber routes for the unified dashboard
func SetupEnhancedSubscriberRoutes(router *gin.RouterGroup, db *database.DB) {
	fmt.Printf("🔧 ENHANCED SUBSCRIBER ROUTES: Starting setup...\n")
	log.Printf("🔧 ENHANCED SUBSCRIBER ROUTES: Function called with router=%v, db=%v", router != nil, db != nil)

	if router == nil {
		log.Printf("❌ ENHANCED SUBSCRIBER ROUTES: router is nil!")
		return
	}
	if db == nil {
		log.Printf("❌ ENHANCED SUBSCRIBER ROUTES: db is nil!")
		return
	}

	enhancedService := services.NewEnhancedSubscriberService(db)
	log.Printf("✅ ENHANCED SUBSCRIBER ROUTES: Service created successfully")

	// Create enhanced subscriber group with middleware - EXACTLY like old subscriber routes
	enhanced := router.Group("/subscribers")
	enhanced.Use(middleware.AuthRequired())
	enhanced.Use(middleware.AdminRequired())

	{
		// Enhanced subscribers endpoint for unified dashboard
		enhanced.GET("/enhanced", func(c *gin.Context) {
			log.Printf("🎯 Enhanced subscribers endpoint called")
			getEnhancedSubscribers(c, enhancedService)
		})
		fmt.Printf("[GIN-debug] GET    %s/subscribers/enhanced --> Enhanced Subscribers Handler\n", router.BasePath())

		// Enhanced subscribers ALL endpoint (no pagination for client-side processing)
		enhanced.GET("/enhanced/all", func(c *gin.Context) {
			log.Printf("🎯 Enhanced subscribers ALL endpoint called (no pagination)")
			getEnhancedSubscribersAll(c, enhancedService)
		})
		fmt.Printf("[GIN-debug] GET    %s/subscribers/enhanced/all --> Enhanced Subscribers ALL Handler\n", router.BasePath())

		// KPIs endpoint for dashboard summary cards
		enhanced.GET("/kpis", func(c *gin.Context) {
			log.Printf("🎯 Enhanced subscriber KPIs endpoint called")
			getSubscriberKPIs(c, enhancedService)
		})
		fmt.Printf("[GIN-debug] GET    %s/subscribers/kpis --> Enhanced Subscriber KPIs Handler\n", router.BasePath())
	}

	fmt.Printf("✅ ENHANCED SUBSCRIBER ROUTES: Setup complete - routes registered successfully\n")
	log.Printf("✅ ENHANCED SUBSCRIBER ROUTES: All routes registered and middleware applied")
}

// getEnhancedSubscribers handles GET /admin/subscribers/enhanced
func getEnhancedSubscribers(c *gin.Context, service *services.EnhancedSubscriberService) {
	log.Printf("🎯 getEnhancedSubscribers handler called - Request: %s %s", c.Request.Method, c.Request.URL.Path)
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 1000 {
		limit = 50
	}

	// Parse filters
	filters := &services.EnhancedSubscriberFilters{}

	if search := c.Query("search"); search != "" {
		filters.Search = search
	}

	if planType := c.Query("plan_type"); planType != "" {
		filters.PlanType = &planType
	}

	if hasActivePlan := c.Query("has_active_plan"); hasActivePlan != "" {
		if val, err := strconv.ParseBool(hasActivePlan); err == nil {
			filters.HasActivePlan = &val
		}
	}

	if hasVideoAccess := c.Query("has_video_access"); hasVideoAccess != "" {
		if val, err := strconv.ParseBool(hasVideoAccess); err == nil {
			filters.HasVideoAccess = &val
		}
	}

	// VideoAccessSource filter removed - plans are now the only source

	if isExpiringSoon := c.Query("is_expiring_soon"); isExpiringSoon != "" {
		if val, err := strconv.ParseBool(isExpiringSoon); err == nil {
			filters.IsExpiringSoon = &val
		}
	}

	if emailVerified := c.Query("email_verified"); emailVerified != "" {
		if val, err := strconv.ParseBool(emailVerified); err == nil {
			filters.EmailVerified = &val
		}
	}

	if role := c.Query("role"); role != "" {
		filters.Role = &role
	}

	if createdDateFrom := c.Query("created_date_from"); createdDateFrom != "" {
		if val, err := time.Parse("2006-01-02", createdDateFrom); err == nil {
			filters.CreatedDateFrom = &val
		}
	}

	if createdDateTo := c.Query("created_date_to"); createdDateTo != "" {
		if val, err := time.Parse("2006-01-02", createdDateTo); err == nil {
			filters.CreatedDateTo = &val
		}
	}

	if lastLoginFrom := c.Query("last_login_from"); lastLoginFrom != "" {
		if val, err := time.Parse("2006-01-02", lastLoginFrom); err == nil {
			filters.LastLoginFrom = &val
		}
	}

	if lastLoginTo := c.Query("last_login_to"); lastLoginTo != "" {
		if val, err := time.Parse("2006-01-02", lastLoginTo); err == nil {
			filters.LastLoginTo = &val
		}
	}

	if minMRR := c.Query("min_mrr"); minMRR != "" {
		if val, err := strconv.ParseFloat(minMRR, 64); err == nil {
			filters.MinMRR = &val
		}
	}

	if maxMRR := c.Query("max_mrr"); maxMRR != "" {
		if val, err := strconv.ParseFloat(maxMRR, 64); err == nil {
			filters.MaxMRR = &val
		}
	}

	// Get enhanced subscribers
	response, err := service.GetEnhancedSubscribers(page, limit, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve enhanced subscribers",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// getEnhancedSubscribersAll handles GET /admin/subscribers/enhanced/all (no pagination)
func getEnhancedSubscribersAll(c *gin.Context, service *services.EnhancedSubscriberService) {
	log.Printf("🎯 getEnhancedSubscribersAll handler called - Request: %s %s", c.Request.Method, c.Request.URL.Path)

	// Parse filters (same as regular endpoint)
	filters := &services.EnhancedSubscriberFilters{}

	if search := c.Query("search"); search != "" {
		filters.Search = search
	}

	if planType := c.Query("plan_type"); planType != "" {
		filters.PlanType = &planType
	}

	if hasActivePlan := c.Query("has_active_plan"); hasActivePlan != "" {
		if val, err := strconv.ParseBool(hasActivePlan); err == nil {
			filters.HasActivePlan = &val
		}
	}

	if hasVideoAccess := c.Query("has_video_access"); hasVideoAccess != "" {
		if val, err := strconv.ParseBool(hasVideoAccess); err == nil {
			filters.HasVideoAccess = &val
		}
	}

	if isExpiringSoon := c.Query("is_expiring_soon"); isExpiringSoon != "" {
		if val, err := strconv.ParseBool(isExpiringSoon); err == nil {
			filters.IsExpiringSoon = &val
		}
	}

	if emailVerified := c.Query("email_verified"); emailVerified != "" {
		if val, err := strconv.ParseBool(emailVerified); err == nil {
			filters.EmailVerified = &val
		}
	}

	if role := c.Query("role"); role != "" {
		filters.Role = &role
	}

	if createdDateFrom := c.Query("created_date_from"); createdDateFrom != "" {
		if val, err := time.Parse("2006-01-02", createdDateFrom); err == nil {
			filters.CreatedDateFrom = &val
		}
	}

	if createdDateTo := c.Query("created_date_to"); createdDateTo != "" {
		if val, err := time.Parse("2006-01-02", createdDateTo); err == nil {
			filters.CreatedDateTo = &val
		}
	}

	if lastLoginFrom := c.Query("last_login_from"); lastLoginFrom != "" {
		if val, err := time.Parse("2006-01-02", lastLoginFrom); err == nil {
			filters.LastLoginFrom = &val
		}
	}

	if lastLoginTo := c.Query("last_login_to"); lastLoginTo != "" {
		if val, err := time.Parse("2006-01-02", lastLoginTo); err == nil {
			filters.LastLoginTo = &val
		}
	}

	if minMRR := c.Query("min_mrr"); minMRR != "" {
		if val, err := strconv.ParseFloat(minMRR, 64); err == nil {
			filters.MinMRR = &val
		}
	}

	if maxMRR := c.Query("max_mrr"); maxMRR != "" {
		if val, err := strconv.ParseFloat(maxMRR, 64); err == nil {
			filters.MaxMRR = &val
		}
	}

	// Use a very high limit to get ALL records (effectively no pagination)
	page := 1
	limit := 100000 // Very high limit to get all records

	log.Printf("🔄 Fetching ALL enhanced subscribers (limit=%d, no effective pagination)", limit)

	// Get enhanced subscribers
	response, err := service.GetEnhancedSubscribers(page, limit, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve all enhanced subscribers",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Successfully fetched %d subscribers (total_count=%d)", len(response.Subscribers), response.TotalCount)

	c.JSON(http.StatusOK, response)
}

// getSubscriberKPIs handles GET /admin/subscribers/kpis
func getSubscriberKPIs(c *gin.Context, service *services.EnhancedSubscriberService) {
	log.Printf("🎯 getSubscriberKPIs handler called - Request: %s %s", c.Request.Method, c.Request.URL.Path)
	kpis, err := service.GetKPIs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve subscriber KPIs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, kpis)
}
