package routes

import (
	"bome-backend/internal/database"
	"bome-backend/internal/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterAnalyticsExportRoutes registers all analytics export routes
func RegisterAnalyticsExportRoutes(router *gin.RouterGroup, db *database.DB) {
	service := services.NewAnalyticsExportService(db)

	// Export routes (admin only)
	exports := router.Group("/exports")
	{
		// Export video analytics
		exports.GET("/video-analytics", func(c *gin.Context) {
			// Parse date range
			startDate, err := parseDate(c.Query("start_date"), time.Now().AddDate(0, 0, -30))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date"})
				return
			}

			endDate, err := parseDate(c.Query("end_date"), time.Now())
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date"})
				return
			}

			req := services.ExportRequest{
				ExportType: "video_analytics",
				Format:     "csv",
				StartDate:  startDate,
				EndDate:    endDate,
			}

			result, err := service.ExportVideoAnalytics(req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export data"})
				return
			}

			// Set headers for file download
			c.Header("Content-Disposition", "attachment; filename="+result.Filename)
			c.Header("Content-Type", result.ContentType)
			c.String(http.StatusOK, result.Content)
		})

		// Export trending videos
		exports.GET("/trending-videos", func(c *gin.Context) {
			limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
			if err != nil || limit <= 0 {
				limit = 50
			}

			result, err := service.ExportTrendingVideos(limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export data"})
				return
			}

			c.Header("Content-Disposition", "attachment; filename="+result.Filename)
			c.Header("Content-Type", result.ContentType)
			c.String(http.StatusOK, result.Content)
		})

		// Export revenue attribution
		exports.GET("/revenue-attribution", func(c *gin.Context) {
			formulaID, err := strconv.Atoi(c.DefaultQuery("formula_id", "1"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid formula_id"})
				return
			}

			periodDays, err := strconv.Atoi(c.DefaultQuery("period_days", "30"))
			if err != nil || periodDays <= 0 {
				periodDays = 30
			}

			result, err := service.ExportRevenueAttribution(formulaID, periodDays)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export data"})
				return
			}

			c.Header("Content-Disposition", "attachment; filename="+result.Filename)
			c.Header("Content-Type", result.ContentType)
			c.String(http.StatusOK, result.Content)
		})

		// Export top converting videos
		exports.GET("/top-converting-videos", func(c *gin.Context) {
			formulaID, err := strconv.Atoi(c.DefaultQuery("formula_id", "1"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid formula_id"})
				return
			}

			limit, err := strconv.Atoi(c.DefaultQuery("limit", "100"))
			if err != nil || limit <= 0 {
				limit = 100
			}

			result, err := service.ExportTopConvertingVideos(formulaID, limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export data"})
				return
			}

			c.Header("Content-Disposition", "attachment; filename="+result.Filename)
			c.Header("Content-Type", result.ContentType)
			c.String(http.StatusOK, result.Content)
		})

		// Export user watch stats
		exports.GET("/user-watch-stats", func(c *gin.Context) {
			result, err := service.ExportUserWatchStats()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export data"})
				return
			}

			c.Header("Content-Disposition", "attachment; filename="+result.Filename)
			c.Header("Content-Type", result.ContentType)
			c.String(http.StatusOK, result.Content)
		})

		// Export daily report
		exports.GET("/daily-report", func(c *gin.Context) {
			dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
			date, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format (use YYYY-MM-DD)"})
				return
			}

			result, err := service.ExportDailyReport(date)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export data"})
				return
			}

			c.Header("Content-Disposition", "attachment; filename="+result.Filename)
			c.Header("Content-Type", result.ContentType)
			c.String(http.StatusOK, result.Content)
		})
	}
}

// parseDate parses a date string or returns default
func parseDate(dateStr string, defaultDate time.Time) (time.Time, error) {
	if dateStr == "" {
		return defaultDate, nil
	}
	return time.Parse("2006-01-02", dateStr)
}

