package handlers

import (
	"fmt"
	"net/http"
	"time"

	"bome-backend/authentication/models"
	"bome-backend/authentication/services"

	"database/sql"
	"encoding/json"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	db                 *models.DB
	planHistoryService *services.PlanHistoryService
}

func NewAnalyticsHandler(db *models.DB, planHistoryService *services.PlanHistoryService) *AnalyticsHandler {
	return &AnalyticsHandler{
		db:                 db,
		planHistoryService: planHistoryService,
	}
}

// SetupAnalyticsRoutes sets up analytics-related routes
func SetupAnalyticsRoutes(router *gin.RouterGroup, db *models.DB, planHistoryService *services.PlanHistoryService) {
	handler := NewAnalyticsHandler(db, planHistoryService)

	// Analytics overview endpoint
	router.GET("/analytics/overview", handler.GetAnalyticsOverview)

	// Promotion analytics endpoint
	router.GET("/analytics/promotions", handler.GetPromotionAnalytics)

	// Timeline events endpoint
	router.GET("/analytics/timeline", handler.GetTimelineEvents)

	// Audit logs endpoint
	router.GET("/analytics/audit-logs", handler.GetAuditLogs)

	// Export analytics data
	router.GET("/analytics/export", handler.ExportAnalyticsData)
}

// GetAnalyticsOverview returns general analytics overview
func (h *AnalyticsHandler) GetAnalyticsOverview(c *gin.Context) {
	period := c.Query("period")
	if period == "" {
		period = "30d"
	}

	// Mock analytics data for now
	analyticsData := map[string]interface{}{
		"total_revenue":       125000.00,
		"total_subscriptions": 1250,
		"revenue_trend": []map[string]interface{}{
			{"date": "2025-07-01", "amount": 4200.00},
			{"date": "2025-07-02", "amount": 4500.00},
			{"date": "2025-07-03", "amount": 4800.00},
			{"date": "2025-07-04", "amount": 5200.00},
			{"date": "2025-07-05", "amount": 4900.00},
		},
		"subscription_trend": []map[string]interface{}{
			{"date": "2025-07-01", "count": 42},
			{"date": "2025-07-02", "count": 45},
			{"date": "2025-07-03", "count": 48},
			{"date": "2025-07-04", "count": 52},
			{"date": "2025-07-05", "count": 49},
		},
		"churn_trend": []map[string]interface{}{
			{"date": "2025-07-01", "rate": 2.1},
			{"date": "2025-07-02", "rate": 1.9},
			{"date": "2025-07-03", "rate": 2.3},
			{"date": "2025-07-04", "rate": 1.8},
			{"date": "2025-07-05", "rate": 2.0},
		},
		"subscription_status_distribution": map[string]int{
			"active":    1200,
			"cancelled": 30,
			"past_due":  20,
		},
		"plan_distribution": map[string]int{
			"monthly": 800,
			"annual":  400,
			"premium": 50,
		},
	}

	c.JSON(http.StatusOK, analyticsData)
}

// GetPromotionAnalytics returns promotion-specific analytics
func (h *AnalyticsHandler) GetPromotionAnalytics(c *gin.Context) {
	// Get all subscription plans
	plans, err := h.db.GetAllSubscriptionPlans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription plans"})
		return
	}

	// Filter promotional plans and process metadata
	promotionAnalytics := []map[string]interface{}{}

	for _, plan := range plans {
		if plan.SubType == "prmo" {
			analytics := map[string]interface{}{
				"id":         plan.ID,
				"name":       plan.Name,
				"price":      plan.Price,
				"currency":   plan.Currency,
				"is_active":  plan.IsActive,
				"sub_type":   plan.SubType,
				"created_at": plan.CreatedAt,
				"updated_at": plan.UpdatedAt,
			}

			// Add promotion dates if available
			if plan.PromotionStartDate.Valid {
				analytics["promotion_start_date"] = plan.PromotionStartDate.Time
			}
			if plan.PromotionEndDate.Valid {
				analytics["promotion_end_date"] = plan.PromotionEndDate.Time
			}

			// Process promotion metadata
			if plan.PromotionMetadata.Valid {
				var metadata map[string]interface{}
				if err := json.Unmarshal([]byte(plan.PromotionMetadata.String), &metadata); err == nil {
					analytics["promotion_metadata"] = metadata
				}
			}

			promotionAnalytics = append(promotionAnalytics, analytics)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"promotions":        promotionAnalytics,
		"total_promotions":  len(promotionAnalytics),
		"active_promotions": len(promotionAnalytics), // Simplified for now
	})
}

// GetTimelineEvents returns timeline events from plan change history
func (h *AnalyticsHandler) GetTimelineEvents(c *gin.Context) {
	// Get all subscription plans
	plans, err := h.db.GetAllSubscriptionPlans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription plans"})
		return
	}

	timelineEvents := []map[string]interface{}{}

	for _, plan := range plans {
		if plan.PlanChangeHistory.Valid && plan.PlanChangeHistory.String != "" {
			var events []map[string]interface{}
			if err := json.Unmarshal([]byte(plan.PlanChangeHistory.String), &events); err == nil {
				for _, event := range events {
					event["plan_id"] = plan.ID
					event["plan_name"] = plan.Name
					timelineEvents = append(timelineEvents, event)
				}
			}
		}
	}

	// Sort events by timestamp (newest first)
	// This is a simplified sort - in production you'd want proper timestamp sorting
	c.JSON(http.StatusOK, gin.H{
		"events":       timelineEvents,
		"total_events": len(timelineEvents),
	})
}

// GetAuditLogs returns audit logs for system activities
func (h *AnalyticsHandler) GetAuditLogs(c *gin.Context) {
	// Mock audit logs for now
	auditLogs := []map[string]interface{}{
		{
			"id":            "audit_001",
			"timestamp":     time.Now().Add(-24 * time.Hour),
			"action":        "plan_created",
			"description":   "New subscription plan 'Premium Monthly' was created",
			"user_id":       "admin_123",
			"resource_type": "subscription_plan",
			"resource_id":   "1",
			"metadata": map[string]interface{}{
				"plan_name": "Premium Monthly",
				"price":     29.99,
				"currency":  "USD",
			},
		},
		{
			"id":            "audit_002",
			"timestamp":     time.Now().Add(-12 * time.Hour),
			"action":        "promotion_started",
			"description":   "Promotion started for 'Premium Monthly' plan",
			"user_id":       "admin_123",
			"resource_type": "subscription_plan",
			"resource_id":   "1",
			"metadata": map[string]interface{}{
				"plan_name":       "Premium Monthly",
				"promotion_start": "2025-07-28T00:00:00Z",
				"promotion_end":   "2025-08-15T00:00:00Z",
			},
		},
		{
			"id":            "audit_003",
			"timestamp":     time.Now().Add(-6 * time.Hour),
			"action":        "plan_updated",
			"description":   "Plan 'Standard Annual' price updated",
			"user_id":       "admin_456",
			"resource_type": "subscription_plan",
			"resource_id":   "2",
			"metadata": map[string]interface{}{
				"plan_name": "Standard Annual",
				"old_price": 95.64,
				"new_price": 99.99,
			},
		},
		{
			"id":            "audit_004",
			"timestamp":     time.Now().Add(-2 * time.Hour),
			"action":        "status_toggled",
			"description":   "Plan 'Plan Share!' status changed to active",
			"user_id":       "admin_123",
			"resource_type": "subscription_plan",
			"resource_id":   "4",
			"metadata": map[string]interface{}{
				"plan_name":  "Plan Share!",
				"old_status": false,
				"new_status": true,
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"audit_logs": auditLogs,
		"total_logs": len(auditLogs),
	})
}

// ExportAnalyticsData exports analytics data in various formats
func (h *AnalyticsHandler) ExportAnalyticsData(c *gin.Context) {
	format := c.Query("format")
	if format == "" {
		format = "json"
	}

	// Get analytics data
	analyticsData := map[string]interface{}{
		"export_timestamp": time.Now(),
		"period":           c.Query("period"),
		"format":           format,
	}

	switch format {
	case "json":
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", "attachment; filename=analytics_export.json")
		c.JSON(http.StatusOK, analyticsData)

	case "csv":
		// In production, you'd implement CSV export
		c.JSON(http.StatusNotImplemented, gin.H{"error": "CSV export not yet implemented"})

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported export format"})
	}
}

// parseFlexibleDate handles multiple date formats
func parseFlexibleDate(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}

	// List of supported date formats
	formats := []string{
		"2006-01-02T15:04:05Z07:00",   // RFC3339
		"2006-01-02T15:04:05",         // ISO without timezone
		"2006-01-02 15:04:05",         // Space separated
		"2006-01-02",                  // Date only
		"01/02/2006",                  // MM/DD/YYYY
		"02/01/2006",                  // DD/MM/YYYY
		"2006-01-02T15:04:05.000Z",    // ISO with milliseconds
		"2006-01-02T15:04:05.000000Z", // ISO with microseconds
	}

	for _, format := range formats {
		if parsed, err := time.Parse(format, dateStr); err == nil {
			return &parsed, nil
		}
	}

	return nil, fmt.Errorf("unable to parse date '%s' with any supported format", dateStr)
}

// formatDateForDatabase ensures consistent date formatting for database storage
func formatDateForDatabase(t *time.Time, isEndDate bool) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}

	var formattedTime time.Time
	if isEndDate {
		// Round up to end of day for end dates (23:59:59.999999999)
		formattedTime = time.Date(
			t.Year(), t.Month(), t.Day(),
			23, 59, 59, 999999999,
			t.Location(),
		)
	} else {
		// Round down to start of day for start dates (00:00:00)
		formattedTime = time.Date(
			t.Year(), t.Month(), t.Day(),
			0, 0, 0, 0,
			t.Location(),
		)
	}

	return sql.NullTime{Time: formattedTime, Valid: true}
}
