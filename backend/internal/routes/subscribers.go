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

// SetupSubscriberRoutes sets up subscriber routes
func SetupSubscriberRoutes(router *gin.RouterGroup, db *database.DB, subscriberService *services.SubscriberService) {
	// Admin routes for subscriber management
	admin := router.Group("/subscribers")
	admin.Use(middleware.AuthRequired())
	admin.Use(middleware.AdminRequired())

	{
		// Get all subscribers with filters
		admin.GET("/", func(c *gin.Context) {
			getSubscribers(c, subscriberService)
		})

		// Get subscriber count
		admin.GET("/count", func(c *gin.Context) {
			getSubscriberCount(c, subscriberService)
		})

		// Get subscriber statistics
		admin.GET("/stats", func(c *gin.Context) {
			getSubscriberStats(c, subscriberService)
		})

		// Export subscribers (must come before parameterized routes)
		admin.GET("/export", func(c *gin.Context) {
			exportSubscribers(c, subscriberService)
		})

		// Get subscribers by plan
		admin.GET("/plan/:planId", func(c *gin.Context) {
			getSubscribersByPlan(c, subscriberService)
		})

		// Get subscribers by status
		admin.GET("/status/:status", func(c *gin.Context) {
			getSubscribersByStatus(c, subscriberService)
		})

		// Search subscribers
		admin.GET("/search", func(c *gin.Context) {
			searchSubscribers(c, subscriberService)
		})

		// Get all non-subscribers with filters
		admin.GET("/non-subscribers", func(c *gin.Context) {
			getNonSubscribers(c, subscriberService)
		})

		// Get non-subscriber count
		admin.GET("/non-subscribers/count", func(c *gin.Context) {
			getNonSubscriberCount(c, subscriberService)
		})

		// Export non-subscribers (must come before parameterized routes)
		admin.GET("/non-subscribers/export", func(c *gin.Context) {
			exportNonSubscribers(c, subscriberService)
		})

		// Get subscriber by ID (must come after more specific routes)
		admin.GET("/:id", func(c *gin.Context) {
			getSubscriberByID(c, subscriberService)
		})

		// Update subscriber
		admin.PUT("/:id", func(c *gin.Context) {
			updateSubscriber(c, subscriberService)
		})

		// Suspend subscriber
		admin.POST("/:id/suspend", func(c *gin.Context) {
			suspendSubscriber(c, subscriberService)
		})

		// Activate subscriber
		admin.POST("/:id/activate", func(c *gin.Context) {
			activateSubscriber(c, subscriberService)
		})

		// Get subscriber history
		admin.GET("/:id/history", func(c *gin.Context) {
			getSubscriberHistory(c, subscriberService)
		})

		// Bulk actions
		admin.POST("/bulk/suspend", func(c *gin.Context) {
			bulkSuspendSubscribers(c, subscriberService)
		})

		admin.POST("/bulk/activate", func(c *gin.Context) {
			bulkActivateSubscribers(c, subscriberService)
		})

		admin.POST("/bulk/change-plan", func(c *gin.Context) {
			bulkChangePlan(c, subscriberService)
		})
	}

	// Debug route
	router.GET("/debug", func(c *gin.Context) {
		subscriberService := services.NewSubscriberService(db)
		err := subscriberService.DebugSubscriptions()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"message": "Debug info logged to console"})
	})
}

// getSubscribers handles GET /api/admin/subscribers/
func getSubscribers(c *gin.Context, service *services.SubscriberService) {
	log.Println("getSubscribers: Starting request")

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	planIDStr := c.Query("plan_id")
	statusStr := c.Query("status")
	searchStr := c.Query("search")
	emailVerifiedStr := c.Query("email_verified")
	roleStr := c.Query("role")
	lastLoginStr := c.Query("last_login")
	createdDateStr := c.Query("created_date")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	log.Printf("getSubscribers: Query params - limit=%s, offset=%s, plan_id=%s, status=%s, search=%s, email_verified=%s, role=%s, last_login=%s, created_date=%s",
		limitStr, offsetStr, planIDStr, statusStr, searchStr, emailVerifiedStr, roleStr, lastLoginStr, createdDateStr)

	// Debug: Print all query parameters
	log.Printf("getSubscribers: All query parameters: %+v", c.Request.URL.Query())

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Build filters
	filters := &services.SubscriberFilters{}

	if planIDStr != "" {
		planID, err := strconv.Atoi(planIDStr)
		if err == nil {
			filters.PlanID = &planID
		}
	}

	if statusStr != "" {
		filters.Status = &statusStr
	}

	if searchStr != "" {
		filters.Search = searchStr
	}

	if emailVerifiedStr != "" {
		emailVerified := emailVerifiedStr == "true"
		filters.EmailVerified = &emailVerified
	}

	if roleStr != "" {
		log.Printf("getSubscribers: Processing role filter: %s", roleStr)
		filters.Role = &roleStr
		log.Printf("getSubscribers: Role filter set to: %s", *filters.Role)
	} else {
		log.Printf("getSubscribers: No role filter provided (roleStr is empty)")
	}

	if lastLoginStr != "" {
		log.Printf("getSubscribers: Processing last login filter: %s", lastLoginStr)
		// Handle different date formats for last login filter
		var lastLogin time.Time
		var err error

		switch lastLoginStr {
		case "today":
			lastLogin = time.Now().Truncate(24 * time.Hour)
		case "week":
			lastLogin = time.Now().AddDate(0, 0, -7)
		case "month":
			lastLogin = time.Now().AddDate(0, -1, 0)
		case "never":
			// Special case for never logged in
			lastLogin = time.Time{}
		default:
			// Try to parse as RFC3339
			lastLogin, err = time.Parse(time.RFC3339, lastLoginStr)
		}

		if err == nil || lastLoginStr == "never" {
			filters.LastLogin = &lastLogin
			log.Printf("getSubscribers: Last login filter set to: %v", *filters.LastLogin)
		} else {
			log.Printf("getSubscribers: Failed to parse last login filter: %v", err)
		}
	} else {
		log.Printf("getSubscribers: No last login filter provided (lastLoginStr is empty)")
	}

	if createdDateStr != "" {
		log.Printf("getSubscribers: Processing created date filter: %s", createdDateStr)
		// Handle different date formats for created date filter
		var createdDate time.Time
		var err error

		switch createdDateStr {
		case "today":
			createdDate = time.Now().Truncate(24 * time.Hour)
		case "week":
			createdDate = time.Now().AddDate(0, 0, -7)
		case "month":
			createdDate = time.Now().AddDate(0, -1, 0)
		case "year":
			createdDate = time.Now().AddDate(-1, 0, 0)
		default:
			// Try to parse as RFC3339
			createdDate, err = time.Parse(time.RFC3339, createdDateStr)
		}

		if err == nil {
			filters.CreatedDate = &createdDate
			log.Printf("getSubscribers: Created date filter set to: %v", *filters.CreatedDate)
		} else {
			log.Printf("getSubscribers: Failed to parse created date filter: %v", err)
		}
	} else {
		log.Printf("getSubscribers: No created date filter provided (createdDateStr is empty)")
	}

	if startDateStr != "" && endDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			endDate, err := time.Parse("2006-01-02", endDateStr)
			if err == nil {
				filters.DateRange = &struct {
					Start time.Time `json:"start"`
					End   time.Time `json:"end"`
				}{
					Start: startDate,
					End:   endDate,
				}
			}
		}
	}

	log.Printf("getSubscribers: Calling service with limit=%d, offset=%d, filters=%+v", limit, offset, filters)

	subscribers, err := service.GetSubscribers(limit, offset, filters)
	if err != nil {
		log.Printf("getSubscribers: Error getting subscribers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscribers", "details": err.Error()})
		return
	}

	log.Printf("getSubscribers: Retrieved %d subscribers", len(subscribers))

	// Get total count for pagination
	total, err := service.GetSubscriberCount(filters)
	if err != nil {
		log.Printf("getSubscribers: Error getting subscriber count: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscriber count", "details": err.Error()})
		return
	}

	log.Printf("getSubscribers: Total subscribers: %d", total)

	response := gin.H{
		"subscribers": subscribers,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	}

	log.Printf("getSubscribers: Sending response with %d subscribers and total=%d", len(subscribers), total)
	c.JSON(http.StatusOK, response)
}

// getSubscriberByID handles GET /api/admin/subscribers/:id
func getSubscriberByID(c *gin.Context, service *services.SubscriberService) {
	idStr := c.Param("id")
	log.Printf("getSubscriberByID: Called with id=%s", idStr)

	// Check if this is actually an export request
	if idStr == "non-subscribers" {
		log.Printf("getSubscriberByID: Detected non-subscribers export request, redirecting")
		// This should not happen if routes are ordered correctly
		c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscriber ID"})
		return
	}

	subscriber, err := service.GetSubscriberByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscriber", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subscriber": subscriber})
}

// getSubscriberCount handles GET /api/admin/subscribers/count
func getSubscriberCount(c *gin.Context, service *services.SubscriberService) {
	// Parse query parameters for filters
	planIDStr := c.Query("plan_id")
	statusStr := c.Query("status")
	searchStr := c.Query("search")
	emailVerifiedStr := c.Query("email_verified")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Build filters
	filters := &services.SubscriberFilters{}

	if planIDStr != "" {
		planID, err := strconv.Atoi(planIDStr)
		if err == nil {
			filters.PlanID = &planID
		}
	}

	if statusStr != "" {
		filters.Status = &statusStr
	}

	if searchStr != "" {
		filters.Search = searchStr
	}

	if emailVerifiedStr != "" {
		emailVerified := emailVerifiedStr == "true"
		filters.EmailVerified = &emailVerified
	}

	if startDateStr != "" && endDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			endDate, err := time.Parse("2006-01-02", endDateStr)
			if err == nil {
				filters.DateRange = &struct {
					Start time.Time `json:"start"`
					End   time.Time `json:"end"`
				}{
					Start: startDate,
					End:   endDate,
				}
			}
		}
	}

	count, err := service.GetSubscriberCount(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscriber count", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// getSubscriberStats handles GET /api/admin/subscribers/stats
func getSubscriberStats(c *gin.Context, service *services.SubscriberService) {
	stats, err := service.GetSubscriberStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscriber stats", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// getSubscribersByPlan handles GET /api/admin/subscribers/plan/:planId
func getSubscribersByPlan(c *gin.Context, service *services.SubscriberService) {
	planIDStr := c.Param("planId")
	planID, err := strconv.Atoi(planIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan ID"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	subscribers, err := service.GetSubscribersByPlan(planID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscribers by plan", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscribers": subscribers,
		"plan_id":     planID,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// getSubscribersByStatus handles GET /api/admin/subscribers/status/:status
func getSubscribersByStatus(c *gin.Context, service *services.SubscriberService) {
	status := c.Param("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status parameter is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	subscribers, err := service.GetSubscribersByStatus(status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscribers by status", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscribers": subscribers,
		"status":      status,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// searchSubscribers handles GET /api/admin/subscribers/search
func searchSubscribers(c *gin.Context, service *services.SubscriberService) {
	searchTerm := c.Query("q")
	if searchTerm == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search term is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	subscribers, err := service.SearchSubscribers(searchTerm, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search subscribers", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscribers": subscribers,
		"search_term": searchTerm,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// getNonSubscribers handles GET /api/admin/subscribers/non-subscribers
func getNonSubscribers(c *gin.Context, service *services.SubscriberService) {
	log.Println("getNonSubscribers: Starting request")

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	searchStr := c.Query("search")
	emailVerifiedStr := c.Query("email_verified")
	roleStr := c.Query("role")
	lastLoginStr := c.Query("last_login")
	createdDateStr := c.Query("created_date")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	subscriptionHistoryStr := c.Query("has_subscription_history") // Updated parameter name

	log.Printf("getNonSubscribers: Query params - limit=%s, offset=%s, search=%s, email_verified=%s, role=%s, last_login=%s, created_date=%s, has_subscription_history=%s",
		limitStr, offsetStr, searchStr, emailVerifiedStr, roleStr, lastLoginStr, createdDateStr, subscriptionHistoryStr)

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Build filters
	filters := &services.NonSubscriberFilters{}

	if searchStr != "" {
		filters.Search = searchStr
	}

	if emailVerifiedStr != "" {
		emailVerified, err := strconv.ParseBool(emailVerifiedStr)
		if err == nil {
			filters.EmailVerified = &emailVerified
		}
	}

	if roleStr != "" {
		filters.Role = &roleStr
	}

	if lastLoginStr != "" {
		// Handle different date formats for last login filter
		var lastLogin time.Time
		var err error

		switch lastLoginStr {
		case "today":
			lastLogin = time.Now().Truncate(24 * time.Hour)
		case "week":
			lastLogin = time.Now().AddDate(0, 0, -7)
		case "month":
			lastLogin = time.Now().AddDate(0, -1, 0)
		case "never":
			// Special case for never logged in
			lastLogin = time.Time{}
		default:
			// Try to parse as RFC3339
			lastLogin, err = time.Parse(time.RFC3339, lastLoginStr)
		}

		if err == nil || lastLoginStr == "never" {
			filters.LastLogin = &lastLogin
		}
	}

	if createdDateStr != "" {
		// Handle different date formats for created date filter
		var createdDate time.Time
		var err error

		switch createdDateStr {
		case "today":
			createdDate = time.Now().Truncate(24 * time.Hour)
		case "week":
			createdDate = time.Now().AddDate(0, 0, -7)
		case "month":
			createdDate = time.Now().AddDate(0, -1, 0)
		case "year":
			createdDate = time.Now().AddDate(-1, 0, 0)
		default:
			// Try to parse as RFC3339
			createdDate, err = time.Parse(time.RFC3339, createdDateStr)
		}

		if err == nil {
			filters.CreatedDate = &createdDate
		}
	}

	if startDateStr != "" && endDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			endDate, err := time.Parse("2006-01-02", endDateStr)
			if err == nil {
				filters.DateRange = &struct {
					Start time.Time `json:"start"`
					End   time.Time `json:"end"`
				}{
					Start: startDate,
					End:   endDate,
				}
			}
		}
	}

	// Handle subscription history filter
	if subscriptionHistoryStr != "" {
		var hasSubscriptionHistory bool
		var shouldSetFilter bool

		switch subscriptionHistoryStr {
		case "true", "previously":
			hasSubscriptionHistory = true
			shouldSetFilter = true
		case "false", "never":
			hasSubscriptionHistory = false
			shouldSetFilter = true
		default:
			// If it's not a recognized value, skip this filter
			shouldSetFilter = false
		}

		if shouldSetFilter {
			filters.HasSubscriptionHistory = &hasSubscriptionHistory
		}
	}

	log.Printf("getNonSubscribers: Calling service with limit=%d, offset=%d, filters=%+v", limit, offset, filters)

	// Get non-subscribers
	nonSubscribers, err := service.GetNonSubscribers(limit, offset, filters)
	if err != nil {
		log.Printf("getNonSubscribers: Error getting non-subscribers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve non-subscribers"})
		return
	}

	log.Printf("getNonSubscribers: Retrieved %d non-subscribers", len(nonSubscribers))

	// Get total count
	total, err := service.GetNonSubscriberCount(filters)
	if err != nil {
		log.Printf("getNonSubscribers: Error getting non-subscriber count: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get non-subscriber count"})
		return
	}

	log.Printf("getNonSubscribers: Total non-subscribers: %d", total)

	response := gin.H{
		"non_subscribers": nonSubscribers,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	}

	log.Printf("getNonSubscribers: Sending response with %d non-subscribers and total=%d", len(nonSubscribers), total)
	c.JSON(http.StatusOK, response)
}

// getNonSubscriberCount handles GET /api/admin/subscribers/non-subscribers/count
func getNonSubscriberCount(c *gin.Context, service *services.SubscriberService) {
	// Parse query parameters
	searchStr := c.Query("search")
	emailVerifiedStr := c.Query("email_verified")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Build filters
	filters := &services.NonSubscriberFilters{}

	if searchStr != "" {
		filters.Search = searchStr
	}

	if emailVerifiedStr != "" {
		emailVerified, err := strconv.ParseBool(emailVerifiedStr)
		if err == nil {
			filters.EmailVerified = &emailVerified
		}
	}

	if startDateStr != "" && endDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			endDate, err := time.Parse("2006-01-02", endDateStr)
			if err == nil {
				filters.DateRange = &struct {
					Start time.Time `json:"start"`
					End   time.Time `json:"end"`
				}{
					Start: startDate,
					End:   endDate,
				}
			}
		}
	}

	// Get count
	count, err := service.GetNonSubscriberCount(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get non-subscriber count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// updateSubscriber handles PUT /api/admin/subscribers/:id
func updateSubscriber(c *gin.Context, service *services.SubscriberService) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscriber ID"})
		return
	}

	var req struct {
		FirstName     string `json:"first_name"`
		LastName      string `json:"last_name"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		PlanID        *int   `json:"plan_id"`
		SubID         *int   `json:"sub_id"`
		Status        string `json:"status"`
		Notes         string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Update subscriber
	updates := map[string]interface{}{
		"first_name":     req.FirstName,
		"last_name":      req.LastName,
		"email":          req.Email,
		"email_verified": req.EmailVerified,
		"notes":          req.Notes,
	}

	// Handle plan_id or sub_id (they both map to the same database field)
	if req.PlanID != nil {
		updates["sub_id"] = *req.PlanID
	} else if req.SubID != nil {
		updates["sub_id"] = *req.SubID
	}

	if req.Status != "" {
		updates["status"] = req.Status
	}

	subscriber, err := service.UpdateSubscriber(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscriber)
}

// suspendSubscriber handles POST /api/admin/subscribers/:id/suspend
func suspendSubscriber(c *gin.Context, service *services.SubscriberService) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscriber ID"})
		return
	}

	subscriber, err := service.SuspendSubscriber(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Subscriber suspended successfully",
		"subscriber": subscriber,
	})
}

// activateSubscriber handles POST /api/admin/subscribers/:id/activate
func activateSubscriber(c *gin.Context, service *services.SubscriberService) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscriber ID"})
		return
	}

	subscriber, err := service.ActivateSubscriber(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Subscriber activated successfully",
		"subscriber": subscriber,
	})
}

// getSubscriberHistory handles GET /api/admin/subscribers/:id/history
func getSubscriberHistory(c *gin.Context, service *services.SubscriberService) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
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

// bulkSuspendSubscribers handles POST /api/admin/subscribers/bulk/suspend
func bulkSuspendSubscribers(c *gin.Context, service *services.SubscriberService) {
	var ids []int
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	updatedSubscribers, err := service.BulkSuspendSubscribers(ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bulk suspend subscribers", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Bulk suspend initiated",
		"subscribers": updatedSubscribers,
	})
}

// bulkActivateSubscribers handles POST /api/admin/subscribers/bulk/activate
func bulkActivateSubscribers(c *gin.Context, service *services.SubscriberService) {
	var ids []int
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	updatedSubscribers, err := service.BulkActivateSubscribers(ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bulk activate subscribers", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Bulk activate initiated",
		"subscribers": updatedSubscribers,
	})
}

// bulkChangePlan handles POST /api/admin/subscribers/bulk/change-plan
func bulkChangePlan(c *gin.Context, service *services.SubscriberService) {
	var req struct {
		PlanID int   `json:"plan_id"`
		IDs    []int `json:"ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	updatedSubscribers, err := service.BulkChangePlan(req.PlanID, req.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bulk change plan", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Bulk change plan initiated",
		"subscribers": updatedSubscribers,
	})
}

// exportSubscribers handles GET /api/admin/subscribers/export
func exportSubscribers(c *gin.Context, service *services.SubscriberService) {
	csvData, err := service.ExportSubscribers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export subscribers", "details": err.Error()})
		return
	}

	// Set headers for CSV download
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=subscribers_%s.csv", time.Now().Format("2006-01-02")))

	c.Data(http.StatusOK, "text/csv", []byte(csvData))
}

// exportNonSubscribers handles GET /api/admin/subscribers/non-subscribers/export
func exportNonSubscribers(c *gin.Context, service *services.SubscriberService) {
	log.Println("exportNonSubscribers: Starting export request")

	csvData, err := service.ExportNonSubscribers()
	if err != nil {
		log.Printf("exportNonSubscribers: Error exporting non-subscribers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export non-subscribers", "details": err.Error()})
		return
	}

	log.Printf("exportNonSubscribers: Generated CSV data with %d bytes", len(csvData))

	// Set headers for CSV download
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=non_subscribers_%s.csv", time.Now().Format("2006-01-02")))

	log.Println("exportNonSubscribers: Sending CSV data response")
	c.Data(http.StatusOK, "text/csv", []byte(csvData))
}
