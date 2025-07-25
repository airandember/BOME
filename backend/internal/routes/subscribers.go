package routes

import (
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

		// Get subscriber by ID
		admin.GET("/:id", func(c *gin.Context) {
			getSubscriberByID(c, subscriberService)
		})

		// Get subscriber count
		admin.GET("/count", func(c *gin.Context) {
			getSubscriberCount(c, subscriberService)
		})

		// Get subscriber statistics
		admin.GET("/stats", func(c *gin.Context) {
			getSubscriberStats(c, subscriberService)
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
	}
}

// getSubscribers handles GET /api/admin/subscribers/
func getSubscribers(c *gin.Context, service *services.SubscriberService) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	planIDStr := c.Query("plan_id")
	statusStr := c.Query("status")
	searchStr := c.Query("search")
	emailVerifiedStr := c.Query("email_verified")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
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

	subscribers, err := service.GetSubscribers(limit, offset, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscribers", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscribers": subscribers,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// getSubscriberByID handles GET /api/admin/subscribers/:id
func getSubscriberByID(c *gin.Context, service *services.SubscriberService) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscriber ID"})
		return
	}

	subscriber, err := service.GetSubscriberByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscriber not found", "details": err.Error()})
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
