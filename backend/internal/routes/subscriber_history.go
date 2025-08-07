package routes

import (
	"log"
	"net/http"
	"strconv"

	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupSubscriberHistoryRoutes sets up subscriber history routes
func SetupSubscriberHistoryRoutes(router *gin.RouterGroup, db *database.DB, subscriberHistoryService *services.SubscriberHistoryService) {
	// Admin routes for subscriber history management
	admin := router.Group("/subscriber-history")
	admin.Use(middleware.AuthRequired())
	admin.Use(middleware.AdminRequired())

	{
		// Get subscriber history
		admin.GET("/:userID", func(c *gin.Context) {
			getSubscriberHistoryDetails(c, subscriberHistoryService)
		})

		// Get subscriber history summary
		admin.GET("/:userID/summary", func(c *gin.Context) {
			getSubscriberHistorySummary(c, subscriberHistoryService)
		})

		// Add admin note
		admin.POST("/:userID/notes", func(c *gin.Context) {
			addAdminNote(c, subscriberHistoryService)
		})

		// Add system note
		admin.POST("/:userID/system-notes", func(c *gin.Context) {
			addSystemNote(c, subscriberHistoryService)
		})

		// Add user note
		admin.POST("/:userID/user-notes", func(c *gin.Context) {
			addUserNote(c, subscriberHistoryService)
		})

		// Export subscriber history
		admin.GET("/:userID/export", func(c *gin.Context) {
			exportSubscriberHistory(c, subscriberHistoryService)
		})
	}
}

// getSubscriberHistoryDetails handles GET /api/admin/subscriber-history/:userID
func getSubscriberHistoryDetails(c *gin.Context, service *services.SubscriberHistoryService) {
	userIDStr := c.Param("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Printf("Route: Invalid user ID: %s", userIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	log.Printf("Route: Getting subscriber history details for user %d", userID)

	history, err := service.GetSubscriberHistory(userID)
	if err != nil {
		log.Printf("Route: Error getting subscriber history: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscriber history", "details": err.Error()})
		return
	}

	log.Printf("Route: Successfully retrieved subscriber history details for user %d", userID)
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"history": history,
	})
}

// getSubscriberHistorySummary handles GET /api/admin/subscriber-history/:userID/summary
func getSubscriberHistorySummary(c *gin.Context, service *services.SubscriberHistoryService) {
	userIDStr := c.Param("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Printf("Route: Invalid user ID: %s", userIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	log.Printf("Route: Getting subscriber history summary for user %d", userID)

	summary, err := service.GetSubscriberHistorySummary(userID)
	if err != nil {
		log.Printf("Route: Error getting subscriber history summary: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscriber history summary", "details": err.Error()})
		return
	}

	log.Printf("Route: Successfully retrieved subscriber history summary for user %d", userID)
	c.JSON(http.StatusOK, summary)
}

// addAdminNote handles POST /api/admin/subscriber-history/:userID/notes
func addAdminNote(c *gin.Context, service *services.SubscriberHistoryService) {
	userIDStr := c.Param("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Printf("Route: Invalid user ID: %s", userIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		AdminID   int    `json:"admin_id" binding:"required"`
		AdminName string `json:"admin_name" binding:"required"`
		Note      string `json:"note" binding:"required"`
		Category  string `json:"category" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Route: Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	log.Printf("Route: Adding admin note for user %d by admin %d", userID, req.AdminID)

	err = service.AddAdminNote(userID, req.AdminID, req.AdminName, req.Note, req.Category)
	if err != nil {
		log.Printf("Route: Error adding admin note: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add admin note", "details": err.Error()})
		return
	}

	log.Printf("Route: Successfully added admin note for user %d", userID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Admin note added successfully",
		"user_id": userID,
	})
}

// addSystemNote handles POST /api/admin/subscriber-history/:userID/system-notes
func addSystemNote(c *gin.Context, service *services.SubscriberHistoryService) {
	userIDStr := c.Param("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Printf("Route: Invalid user ID: %s", userIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Note     string `json:"note" binding:"required"`
		Category string `json:"category" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Route: Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	log.Printf("Route: Adding system note for user %d", userID)

	err = service.AddSystemNote(userID, req.Note, req.Category)
	if err != nil {
		log.Printf("Route: Error adding system note: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add system note", "details": err.Error()})
		return
	}

	log.Printf("Route: Successfully added system note for user %d", userID)
	c.JSON(http.StatusOK, gin.H{
		"message": "System note added successfully",
		"user_id": userID,
	})
}

// addUserNote handles POST /api/admin/subscriber-history/:userID/user-notes
func addUserNote(c *gin.Context, service *services.SubscriberHistoryService) {
	userIDStr := c.Param("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Printf("Route: Invalid user ID: %s", userIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Note     string `json:"note" binding:"required"`
		Category string `json:"category" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Route: Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	log.Printf("Route: Adding user note for user %d", userID)

	err = service.AddUserNote(userID, req.Note, req.Category)
	if err != nil {
		log.Printf("Route: Error adding user note: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user note", "details": err.Error()})
		return
	}

	log.Printf("Route: Successfully added user note for user %d", userID)
	c.JSON(http.StatusOK, gin.H{
		"message": "User note added successfully",
		"user_id": userID,
	})
}

// exportSubscriberHistory handles GET /api/admin/subscriber-history/:userID/export
func exportSubscriberHistory(c *gin.Context, service *services.SubscriberHistoryService) {
	userIDStr := c.Param("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Printf("Route: Invalid user ID: %s", userIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	log.Printf("Route: Exporting subscriber history for user %d", userID)

	exportData, err := service.ExportSubscriberHistory(userID)
	if err != nil {
		log.Printf("Route: Error exporting subscriber history: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export subscriber history", "details": err.Error()})
		return
	}

	log.Printf("Route: Successfully exported subscriber history for user %d", userID)

	// Set headers for file download
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=subscriber_history_"+userIDStr+".json")
	c.Data(http.StatusOK, "application/json", exportData)
}
