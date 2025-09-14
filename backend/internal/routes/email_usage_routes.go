package routes

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// EmailUsageStats represents email usage statistics for a provider
type EmailUsageStats struct {
	Provider     string `json:"provider"`
	EmailsSent   int    `json:"emails_sent"`
	EmailsFailed int    `json:"emails_failed"`
	DailyLimit   int    `json:"daily_limit"`
	Remaining    int    `json:"remaining"`
	UsagePercent int    `json:"usage_percent"`
}

// EmailUsageResponse represents the complete email usage response
type EmailUsageResponse struct {
	Date           string            `json:"date"`
	Providers      []EmailUsageStats `json:"providers"`
	TotalSent      int               `json:"total_sent"`
	TotalFailed    int               `json:"total_failed"`
	TotalRemaining int               `json:"total_remaining"`
	TotalLimit     int               `json:"total_limit"`
	OverallPercent int               `json:"overall_percent"`
	FailoverCount  int               `json:"failover_count"`
}

// SetupEmailUsageRoutes sets up email usage tracking routes
func SetupEmailUsageRoutes(router *gin.RouterGroup, db *database.DB) {
	email := router.Group("/email")
	// Add authentication, email verification, and admin requirements for all email routes
	email.Use(middleware.AuthRequired())
	email.Use(middleware.RequireEmailVerificationForDashboard())
	email.Use(middleware.AdminRequired())
	// Add database middleware
	email.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})
	{
		// Get today's email usage statistics
		email.GET("/usage", func(c *gin.Context) {
			getEmailUsageStats(c)
		})

		// Get email usage for a specific date
		email.GET("/usage/:date", func(c *gin.Context) {
			getEmailUsageStatsForDate(c, db)
		})

		// Get email usage history (last 7 days)
		email.GET("/usage/history", func(c *gin.Context) {
			getEmailUsageHistory(c)
		})

		// Get email provider settings
		email.GET("/settings", func(c *gin.Context) {
			getEmailSettings(c)
		})

		// Update email provider settings
		email.PUT("/settings", func(c *gin.Context) {
			updateEmailSettings(c)
		})

		// Test email sending (admin only)
		email.POST("/test", func(c *gin.Context) {
			testEmailSending(c)
		})
	}
}

// getEmailUsageStats gets today's email usage statistics
func getEmailUsageStats(c *gin.Context) {
	today := time.Now().Format("2006-01-02")
	getEmailUsageStatsForDateInternal(c, today)
}

// getEmailUsageStatsForDate gets email usage statistics for a specific date
func getEmailUsageStatsForDate(c *gin.Context, db *database.DB) {
	date := c.Param("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Date parameter is required"})
		return
	}
	getEmailUsageStatsForDateInternal(c, date)
}

// getEmailUsageStatsForDateInternal internal function to get usage stats for a date
func getEmailUsageStatsForDateInternal(c *gin.Context, date string) {
	db := c.MustGet("db").(*database.DB)

	// Get usage data for all providers
	query := `
		SELECT provider, 
		       COALESCE(emails_sent, 0) as emails_sent,
		       COALESCE(emails_failed, 0) as emails_failed
		FROM daily_email_usage 
		WHERE date = $1
		ORDER BY provider`

	rows, err := db.DB.Query(query, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get email usage stats"})
		return
	}
	defer rows.Close()

	var providers []EmailUsageStats
	totalSent := 0
	totalFailed := 0
	totalLimit := 0

	// Process existing data
	existingProviders := make(map[string]EmailUsageStats)
	for rows.Next() {
		var provider string
		var sent, failed int

		if err := rows.Scan(&provider, &sent, &failed); err != nil {
			continue
		}

		// Get limit for this provider (default 100)
		limit := getProviderLimitFromDB(db, provider, 100)
		remaining := limit - sent
		if remaining < 0 {
			remaining = 0
		}

		usagePercent := 0
		if limit > 0 {
			usagePercent = (sent * 100) / limit
		}

		stat := EmailUsageStats{
			Provider:     provider,
			EmailsSent:   sent,
			EmailsFailed: failed,
			DailyLimit:   limit,
			Remaining:    remaining,
			UsagePercent: usagePercent,
		}

		existingProviders[provider] = stat
		totalSent += sent
		totalFailed += failed
		totalLimit += limit
	}

	// Ensure we have entries for both resend and mailgun (even if 0 usage)
	defaultProviders := []string{"resend", "mailgun"}
	for _, provider := range defaultProviders {
		if stat, exists := existingProviders[provider]; exists {
			providers = append(providers, stat)
		} else {
			// Create empty entry for provider with no usage yet
			limit := getProviderLimitFromDB(db, provider, 100)
			stat := EmailUsageStats{
				Provider:     provider,
				EmailsSent:   0,
				EmailsFailed: 0,
				DailyLimit:   limit,
				Remaining:    limit,
				UsagePercent: 0,
			}
			providers = append(providers, stat)
			totalLimit += limit
		}
	}

	totalRemaining := totalLimit - totalSent
	if totalRemaining < 0 {
		totalRemaining = 0
	}

	overallPercent := 0
	if totalLimit > 0 {
		overallPercent = (totalSent * 100) / totalLimit
	}

	// Calculate failover count (when resend is at/near limit but mailgun has usage)
	failoverCount := 0
	if len(providers) >= 2 {
		resendUsage := providers[0].UsagePercent
		mailgunSent := providers[1].EmailsSent
		if resendUsage >= 90 && mailgunSent > 0 {
			failoverCount = mailgunSent
		}
	}

	response := EmailUsageResponse{
		Date:           date,
		Providers:      providers,
		TotalSent:      totalSent,
		TotalFailed:    totalFailed,
		TotalRemaining: totalRemaining,
		TotalLimit:     totalLimit,
		OverallPercent: overallPercent,
		FailoverCount:  failoverCount,
	}

	c.JSON(http.StatusOK, response)
}

// getProviderLimitFromDB gets the daily limit for a provider from database settings
func getProviderLimitFromDB(db *database.DB, provider string, defaultLimit int) int {
	settingKey := "daily_email_limit_" + provider
	limitStr, err := db.GetEmailSetting(settingKey)
	if err != nil || limitStr == "" {
		return defaultLimit
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return defaultLimit
	}

	return limit
}

// getEmailUsageHistory gets email usage for the last 7 days
func getEmailUsageHistory(c *gin.Context) {
	db := c.MustGet("db").(*database.DB)
	// Get last 7 days of data
	query := `
		SELECT date, provider, 
		       COALESCE(emails_sent, 0) as emails_sent,
		       COALESCE(emails_failed, 0) as emails_failed
		FROM daily_email_usage 
		WHERE date >= CURRENT_DATE - INTERVAL '7 days'
		ORDER BY date DESC, provider`

	rows, err := db.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get email usage history"})
		return
	}
	defer rows.Close()

	history := make(map[string][]EmailUsageStats)

	for rows.Next() {
		var date, provider string
		var sent, failed int

		if err := rows.Scan(&date, &provider, &sent, &failed); err != nil {
			continue
		}

		limit := getProviderLimitFromDB(db, provider, 100)
		remaining := limit - sent
		if remaining < 0 {
			remaining = 0
		}

		usagePercent := 0
		if limit > 0 {
			usagePercent = (sent * 100) / limit
		}

		stat := EmailUsageStats{
			Provider:     provider,
			EmailsSent:   sent,
			EmailsFailed: failed,
			DailyLimit:   limit,
			Remaining:    remaining,
			UsagePercent: usagePercent,
		}

		history[date] = append(history[date], stat)
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
		"days":    len(history),
	})
}

// getEmailSettings gets current email provider settings
func getEmailSettings(c *gin.Context) {
	db := c.MustGet("db").(*database.DB)
	settings := make(map[string]string)

	// Get all email settings (non-encrypted ones for display)
	query := `
		SELECT setting_key, setting_value 
		FROM email_settings 
		WHERE setting_key IN (
			'email_provider_primary', 'email_provider_secondary',
			'daily_email_limit_resend', 'daily_email_limit_mailgun',
			'auto_failover_enabled', 'email_enabled'
		)`

	rows, err := db.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get email settings"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		settings[key] = value
	}

	c.JSON(http.StatusOK, gin.H{
		"settings": settings,
	})
}

// updateEmailSettings updates email provider settings
func updateEmailSettings(c *gin.Context) {
	db := c.MustGet("db").(*database.DB)
	var request map[string]string
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Update allowed settings
	allowedSettings := []string{
		"daily_email_limit_resend", "daily_email_limit_mailgun",
		"email_provider_primary", "auto_failover_enabled", "email_enabled",
	}

	for _, key := range allowedSettings {
		if value, exists := request[key]; exists {
			err := db.SetEmailSetting(key, value)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to update setting: " + key,
				})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Email settings updated successfully",
	})
}

// testEmailSending sends a test email to verify the system is working
func testEmailSending(c *gin.Context) {
	db := c.MustGet("db").(*database.DB)
	var request struct {
		Email   string `json:"email" binding:"required"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email address is required"})
		return
	}

	// Set defaults if not provided
	if request.Subject == "" {
		request.Subject = "Test Email from BOME Admin"
	}
	if request.Body == "" {
		request.Body = "This is a test email to verify your email configuration is working correctly."
	}

	// Check if email is enabled
	enabled, err := db.GetEmailSetting("email_enabled")
	if err != nil || enabled != "true" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Email service is not enabled",
		})
		return
	}

	// Get global email service from routes context
	emailService := getGlobalEmailService()
	if emailService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Email service not available",
		})
		return
	}

	// Get current user ID for notification tracking
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Send test email
	err = emailService.SendTestEmail(request.Email, request.Subject, request.Body, userIDInt)
	if err != nil {
		log.Printf("❌ [EMAIL-TEST] Failed to send test email: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to send test email: " + err.Error(),
		})
		return
	}

	log.Printf("✅ [EMAIL-TEST] Test email sent successfully to %s", request.Email)
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Test email sent successfully!",
		"email":   request.Email,
	})
}

// getGlobalEmailService returns the global email service instance
func getGlobalEmailService() *services.EmailService {
	return globalEmailService
}
