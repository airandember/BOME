package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"bome-backend/authentication/middleware"
	"bome-backend/infrastructure/database"
	subServices "bome-backend/services/subscription"

	"github.com/gin-gonic/gin"
)

// WebSocketHub interface to avoid circular imports
type WebSocketHub interface {
	BroadcastSubscriberCreated(subscriber interface{})
	BroadcastSubscriberUpdated(subscriber interface{})
	BroadcastKPIUpdate(kpis interface{})
	BroadcastEvent(eventType string, data map[string]interface{}, message string)
}

// SetupSubscriberRoutes sets up subscriber management routes
func SetupSubscriberRoutes(router *gin.RouterGroup, db *database.DB, hub WebSocketHub) {
	fmt.Printf("🔧 SUBSCRIBER ROUTES: Starting setup...\n")
	log.Printf("🔧 SUBSCRIBER ROUTES: Function called with router=%v, db=%v, hub=%v", router != nil, db != nil, hub != nil)

	if router == nil {
		log.Printf("❌ SUBSCRIBER ROUTES: router is nil!")
		return
	}
	if db == nil {
		log.Printf("❌ SUBSCRIBER ROUTES: db is nil!")
		return
	}

	// Initialize services (with WebSocket hub for real-time updates)
	subscriberService := subServices.NewSubscriberService(db, hub)
	enhancedSubscriberService := subServices.NewEnhancedSubscriberService(db)

	log.Printf("✅ SUBSCRIBER ROUTES: Services created successfully")

	// Create subscriber group with admin middleware
	subscribers := router.Group("/subscribers")
	subscribers.Use(middleware.AuthRequired())
	subscribers.Use(middleware.AdminRequired())

	{
		// ========================================
		// ENHANCED SUBSCRIBER ENDPOINTS (Priority)
		// ========================================

		// Enhanced subscribers endpoint for unified dashboard
		subscribers.GET("/enhanced", func(c *gin.Context) {
			log.Printf("🎯 Enhanced subscribers endpoint called")
			getEnhancedSubscribers(c, enhancedSubscriberService)
		})
		fmt.Printf("[GIN-debug] GET    %s/subscribers/enhanced --> Enhanced Subscribers Handler\n", router.BasePath())

		// Enhanced subscribers ALL endpoint (no pagination for client-side processing)
		subscribers.GET("/enhanced/all", func(c *gin.Context) {
			log.Printf("🎯 Enhanced subscribers ALL endpoint called (no pagination)")
			getEnhancedSubscribersAll(c, enhancedSubscriberService)
		})
		fmt.Printf("[GIN-debug] GET    %s/subscribers/enhanced/all --> Enhanced Subscribers ALL Handler\n", router.BasePath())

		// KPIs endpoint for dashboard summary cards
		subscribers.GET("/kpis", func(c *gin.Context) {
			log.Printf("🎯 Enhanced subscriber KPIs endpoint called")
			getSubscriberKPIs(c, enhancedSubscriberService)
		})
		fmt.Printf("[GIN-debug] GET    %s/subscribers/kpis --> Enhanced Subscriber KPIs Handler\n", router.BasePath())

		// ========================================
		// STANDARD SUBSCRIBER ENDPOINTS
		// ========================================

		// Get all subscribers with filters
		subscribers.GET("/", func(c *gin.Context) {
			log.Printf("🎯 Get all subscribers endpoint called")
			getSubscribers(c, subscriberService)
		})

		// Get subscriber count
		subscribers.GET("/count", func(c *gin.Context) {
			log.Printf("🎯 Get subscriber count endpoint called")
			getSubscriberCount(c, subscriberService)
		})

		// Get subscriber statistics
		subscribers.GET("/stats", func(c *gin.Context) {
			log.Printf("🎯 Get subscriber stats endpoint called")
			getSubscriberStats(c, subscriberService)
		})

		// Get subscribers by email verification status
		subscribers.GET("/verified", func(c *gin.Context) {
			log.Printf("🎯 Get verified subscribers endpoint called")
			getSubscribersByEmailVerification(c, subscriberService, true)
		})

		subscribers.GET("/unverified", func(c *gin.Context) {
			log.Printf("🎯 Get unverified subscribers endpoint called")
			getSubscribersByEmailVerification(c, subscriberService, false)
		})

		// Get subscriber counts by email verification status
		subscribers.GET("/verified/count", func(c *gin.Context) {
			getSubscriberCountByEmailVerification(c, subscriberService, true)
		})

		subscribers.GET("/unverified/count", func(c *gin.Context) {
			getSubscriberCountByEmailVerification(c, subscriberService, false)
		})

		// Export subscribers (must come before parameterized routes)
		subscribers.GET("/export", func(c *gin.Context) {
			log.Printf("🎯 Export subscribers endpoint called")
			exportSubscribers(c, subscriberService)
		})

		// Get subscribers by plan
		subscribers.GET("/plan/:planId", func(c *gin.Context) {
			getSubscribersByPlan(c, subscriberService)
		})

		// Get subscribers by status
		subscribers.GET("/status/:status", func(c *gin.Context) {
			getSubscribersByStatus(c, subscriberService)
		})

		// Search subscribers
		subscribers.GET("/search", func(c *gin.Context) {
			log.Printf("🎯 Search subscribers endpoint called")
			searchSubscribers(c, subscriberService)
		})

		// Get all non-subscribers with filters
		subscribers.GET("/non-subscribers", func(c *gin.Context) {
			getNonSubscribers(c, subscriberService)
		})

		// Get non-subscriber count
		subscribers.GET("/non-subscribers/count", func(c *gin.Context) {
			getNonSubscriberCount(c, subscriberService)
		})

		// Export non-subscribers (must come before parameterized routes)
		subscribers.GET("/non-subscribers/export", func(c *gin.Context) {
			exportNonSubscribers(c, subscriberService)
		})

		// Get subscriber by ID (must come after more specific routes)
		subscribers.GET("/:id", func(c *gin.Context) {
			log.Printf("🎯 Get subscriber by ID endpoint called")
			getSubscriberByID(c, subscriberService)
		})

		// Update subscriber
		subscribers.PUT("/:id", func(c *gin.Context) {
			log.Printf("🎯 Update subscriber endpoint called")
			updateSubscriber(c, subscriberService)
		})

		// Delete subscriber (soft delete)
		subscribers.DELETE("/:id", func(c *gin.Context) {
			deleteSubscriber(c, subscriberService)
		})

		// Grant video access
		subscribers.POST("/:id/grant-access", func(c *gin.Context) {
			log.Printf("🎯 Grant video access endpoint called")
			grantVideoAccess(c, subscriberService)
		})

		// Get subscriber subscription history
		subscribers.GET("/:id/history", func(c *gin.Context) {
			getSubscriberHistory(c, subscriberService)
		})
	}

	fmt.Printf("✅ SUBSCRIBER ROUTES: Setup complete - %d routes registered\n", 24)
	log.Printf("✅ SUBSCRIBER ROUTES: All routes registered and middleware applied")
}

// ============================================================================
// ENHANCED SUBSCRIBER HANDLERS (New Dashboard)
// ============================================================================

// getEnhancedSubscribers handles GET /admin/subscribers/enhanced
func getEnhancedSubscribers(c *gin.Context, service *subServices.EnhancedSubscriberService) {
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
	filters := &subServices.EnhancedSubscriberFilters{}

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

	// Get enhanced subscribers
	response, err := service.GetEnhancedSubscribers(page, limit, filters)
	if err != nil {
		log.Printf("❌ Error getting enhanced subscribers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve enhanced subscribers",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Successfully retrieved %d enhanced subscribers (page %d, limit %d)", len(response.Subscribers), page, limit)
	c.JSON(http.StatusOK, response)
}

// getEnhancedSubscribersAll handles GET /admin/subscribers/enhanced/all (no pagination)
func getEnhancedSubscribersAll(c *gin.Context, service *subServices.EnhancedSubscriberService) {
	log.Printf("🎯 getEnhancedSubscribersAll handler called - Request: %s %s", c.Request.Method, c.Request.URL.Path)

	// Parse filters (same as regular endpoint)
	filters := &subServices.EnhancedSubscriberFilters{}

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
		log.Printf("❌ Error getting all enhanced subscribers: %v", err)
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
func getSubscriberKPIs(c *gin.Context, service *subServices.EnhancedSubscriberService) {
	log.Printf("🎯 getSubscriberKPIs handler called - Request: %s %s", c.Request.Method, c.Request.URL.Path)

	kpis, err := service.GetKPIs()
	if err != nil {
		log.Printf("❌ Error getting subscriber KPIs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve subscriber KPIs",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Successfully retrieved subscriber KPIs")
	c.JSON(http.StatusOK, kpis)
}

// ============================================================================
// STANDARD SUBSCRIBER HANDLERS
// ============================================================================

// getSubscribers handles GET /admin/subscribers
func getSubscribers(c *gin.Context, service *subServices.SubscriberService) {
	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 1000 {
		limit = 50
	}

	// Parse filters
	filters := &subServices.SubscriberFilters{}
	if search := c.Query("search"); search != "" {
		filters.Search = search
	}
	if role := c.Query("role"); role != "" {
		filters.Role = &role
	}
	if emailVerified := c.Query("email_verified"); emailVerified != "" {
		if val, err := strconv.ParseBool(emailVerified); err == nil {
			filters.EmailVerified = &val
		}
	}

	subscribers, total, err := service.GetSubscribers(page, limit, filters)
	if err != nil {
		log.Printf("❌ Error getting subscribers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve subscribers",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscribers": subscribers,
		"total":       total,
		"page":        page,
		"limit":       limit,
	})
}

// getSubscriberCount handles GET /admin/subscribers/count
func getSubscriberCount(c *gin.Context, service *subServices.SubscriberService) {
	count, err := service.GetSubscriberCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

// getSubscriberStats handles GET /admin/subscribers/stats
func getSubscriberStats(c *gin.Context, service *subServices.SubscriberService) {
	stats, err := service.GetSubscriberStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// getSubscribersByEmailVerification handles GET /admin/subscribers/verified and /unverified
func getSubscribersByEmailVerification(c *gin.Context, service *subServices.SubscriberService, verified bool) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	subscribers, total, err := service.GetSubscribersByEmailVerification(verified, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscribers": subscribers,
		"total":       total,
		"page":        page,
		"limit":       limit,
	})
}

// getSubscriberCountByEmailVerification handles GET /admin/subscribers/verified/count and /unverified/count
func getSubscriberCountByEmailVerification(c *gin.Context, service *subServices.SubscriberService, verified bool) {
	count, err := service.GetSubscriberCountByEmailVerification(verified)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

// getSubscribersByPlan handles GET /admin/subscribers/plan/:planId
func getSubscribersByPlan(c *gin.Context, service *subServices.SubscriberService) {
	planID, err := strconv.Atoi(c.Param("planId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	subscribers, total, err := service.GetSubscribersByPlan(planID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscribers": subscribers,
		"total":       total,
		"page":        page,
		"limit":       limit,
	})
}

// getSubscribersByStatus handles GET /admin/subscribers/status/:status
func getSubscribersByStatus(c *gin.Context, service *subServices.SubscriberService) {
	status := c.Param("status")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	subscribers, total, err := service.GetSubscribersByStatus(status, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscribers": subscribers,
		"total":       total,
		"page":        page,
		"limit":       limit,
	})
}

// searchSubscribers handles GET /admin/subscribers/search
func searchSubscribers(c *gin.Context, service *subServices.SubscriberService) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	subscribers, total, err := service.SearchSubscribers(query, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscribers": subscribers,
		"total":       total,
		"page":        page,
		"limit":       limit,
	})
}

// getNonSubscribers handles GET /admin/subscribers/non-subscribers
func getNonSubscribers(c *gin.Context, service *subServices.SubscriberService) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	filters := &subServices.NonSubscriberFilters{}
	if search := c.Query("search"); search != "" {
		filters.Search = search
	}

	nonSubscribers, total, err := service.GetNonSubscribers(page, limit, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"non_subscribers": nonSubscribers,
		"total":           total,
		"page":            page,
		"limit":           limit,
	})
}

// getNonSubscriberCount handles GET /admin/subscribers/non-subscribers/count
func getNonSubscriberCount(c *gin.Context, service *subServices.SubscriberService) {
	count, err := service.GetNonSubscriberCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

// exportSubscribers handles GET /admin/subscribers/export
func exportSubscribers(c *gin.Context, service *subServices.SubscriberService) {
	// Export to CSV
	csvData, err := service.ExportSubscribersToCSV()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=subscribers.csv")
	c.String(http.StatusOK, csvData)
}

// exportNonSubscribers handles GET /admin/subscribers/non-subscribers/export
func exportNonSubscribers(c *gin.Context, service *subServices.SubscriberService) {
	csvData, err := service.ExportNonSubscribersToCSV()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=non-subscribers.csv")
	c.String(http.StatusOK, csvData)
}

// getSubscriberByID handles GET /admin/subscribers/:id
func getSubscriberByID(c *gin.Context, service *subServices.SubscriberService) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscriber ID"})
		return
	}

	subscriber, err := service.GetSubscriberByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if subscriber == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscriber not found"})
		return
	}

	c.JSON(http.StatusOK, subscriber)
}

// updateSubscriber handles PUT /admin/subscribers/:id
func updateSubscriber(c *gin.Context, service *subServices.SubscriberService) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscriber ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.UpdateSubscriber(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get updated subscriber
	subscriber, err := service.GetSubscriberByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("✅ Successfully updated subscriber %d", id)
	c.JSON(http.StatusOK, subscriber)
}

// deleteSubscriber handles DELETE /admin/subscribers/:id
func deleteSubscriber(c *gin.Context, service *subServices.SubscriberService) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscriber ID"})
		return
	}

	if err := service.DeleteSubscriber(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscriber deleted successfully"})
}

// grantVideoAccess handles POST /admin/subscribers/:id/grant-access
func grantVideoAccess(c *gin.Context, service *subServices.SubscriberService) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscriber ID"})
		return
	}

	var req struct {
		ExpiryDate *time.Time `json:"expiry_date"`
		Reason     string     `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.GrantVideoAccess(id, req.ExpiryDate, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("✅ Granted video access to subscriber %d", id)
	c.JSON(http.StatusOK, gin.H{"message": "Video access granted successfully"})
}

// getSubscriberHistory handles GET /admin/subscribers/:id/history
func getSubscriberHistory(c *gin.Context, service *subServices.SubscriberService) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscriber ID"})
		return
	}

	history, err := service.GetSubscriberHistory(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"history": history})
}
