package routes

import (
	"net/http"

	"bome-backend/internal/database"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupPublicSupportSettingsRoutes sets up public support settings routes (no auth)
func SetupPublicSupportSettingsRoutes(router *gin.RouterGroup, db *database.DB) {
	settingsService := services.NewSupportSettingsService(db)

	system := router.Group("/system")
	{
		// Get support settings (for displaying support contact to users)
		system.GET("/support", func(c *gin.Context) {
			getSupportSettings(c, settingsService)
		})
	}
}

// SetupAdminSupportSettingsRoutes sets up admin support settings routes
func SetupAdminSupportSettingsRoutes(adminGroup *gin.RouterGroup, db *database.DB) {
	settingsService := services.NewSupportSettingsService(db)

	supportSettings := adminGroup.Group("/support-settings")
	{
		// Get support settings
		supportSettings.GET("/", func(c *gin.Context) {
			getSupportSettingsAdmin(c, settingsService)
		})

		// Update support settings (batch)
		supportSettings.PUT("/", func(c *gin.Context) {
			updateSupportSettings(c, settingsService)
		})
	}
}

// ================================================================
// PUBLIC HANDLERS (No Auth)
// ================================================================

// getSupportSettings returns support contact settings
func getSupportSettings(c *gin.Context, service *services.SupportSettingsService) {
	support, err := service.GetSupportSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"support": support,
	})
}

// ================================================================
// ADMIN HANDLERS (Auth Required)
// ================================================================

// getSupportSettingsAdmin returns support settings (admin only)
func getSupportSettingsAdmin(c *gin.Context, service *services.SupportSettingsService) {
	support, err := service.GetSupportSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"support": support,
	})
}

// updateSupportSettings updates support settings (admin only)
func updateSupportSettings(c *gin.Context, service *services.SupportSettingsService) {
	var req struct {
		SupportEmail   string `json:"support_email"`
		SupportPhone   string `json:"support_phone"`
		SupportURL     string `json:"support_url"`
		SupportHours   string `json:"support_hours"`
		SupportMessage string `json:"support_message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	settings := map[string]string{
		"support_email":   req.SupportEmail,
		"support_phone":   req.SupportPhone,
		"support_url":     req.SupportURL,
		"support_hours":   req.SupportHours,
		"support_message": req.SupportMessage,
	}

	if err := service.UpdateSupportSettings(settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Support settings updated successfully",
	})
}
