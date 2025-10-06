package routes

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GhostCustomer represents a customer that exists locally but not in Stripe
type GhostCustomer struct {
	ID                int       `json:"id"`
	LocalCustomerID   *int      `json:"local_customer_id"`
	StripeCustomerID  string    `json:"stripe_customer_id"`
	CustomerEmail     string    `json:"customer_email"`
	CustomerName      string    `json:"customer_name"`
	GhostType         string    `json:"ghost_type"`
	GhostReason       string    `json:"ghost_reason"`
	PurgeStatus       string    `json:"purge_status"`
	DetectionDate     time.Time `json:"detection_date"`
	Notes             *string   `json:"notes"`
	CurrentStatus     string    `json:"current_status"`
	SubscriptionCount int       `json:"subscription_count"`
	InvoiceCount      int       `json:"invoice_count"`
}

// GhostDetectionSummary provides overview of ghost detection results
type GhostDetectionSummary struct {
	TotalGhosts         int `json:"total_ghosts"`
	HashIDGhosts        int `json:"hash_id_ghosts"`
	InvalidFormatGhosts int `json:"invalid_format_ghosts"`
	MarkedForPurge      int `json:"marked_for_purge"`
	AlreadyPurged       int `json:"already_purged"`
}

// GhostCustomerHandler handles ghost customer management
type GhostCustomerHandler struct {
	db *sql.DB
}

// NewGhostCustomerHandler creates a new ghost customer handler
func NewGhostCustomerHandler(db *sql.DB) *GhostCustomerHandler {
	return &GhostCustomerHandler{db: db}
}

// GetGhostCustomers returns all detected ghost customers
func (h *GhostCustomerHandler) GetGhostCustomers(c *gin.Context) {
	query := `
		SELECT 
			sg.id, sg.local_customer_id, sg.stripe_customer_id, sg.customer_email,
			sg.customer_name, sg.ghost_type, sg.ghost_reason, sg.purge_status,
			sg.detection_date, sg.notes,
			CASE WHEN sc.id IS NOT NULL THEN 'exists' ELSE 'already_deleted' END as current_status,
			COALESCE(sub_count.count, 0) as subscription_count,
			COALESCE(inv_count.count, 0) as invoice_count
		FROM stripe_ghosts sg
		LEFT JOIN stripe_customers sc ON sg.stripe_customer_id = sc.stripe_id
		LEFT JOIN (
			SELECT customer_id, COUNT(*) as count 
			FROM stripe_subscriptions 
			GROUP BY customer_id
		) sub_count ON sc.id = sub_count.customer_id
		LEFT JOIN (
			SELECT customer_id, COUNT(*) as count 
			FROM stripe_invoices 
			GROUP BY customer_id
		) inv_count ON sc.id = inv_count.customer_id
		ORDER BY sg.detection_date DESC
	`

	rows, err := h.db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query ghost customers"})
		return
	}
	defer rows.Close()

	var ghosts []GhostCustomer
	for rows.Next() {
		var ghost GhostCustomer
		err := rows.Scan(
			&ghost.ID, &ghost.LocalCustomerID, &ghost.StripeCustomerID,
			&ghost.CustomerEmail, &ghost.CustomerName, &ghost.GhostType,
			&ghost.GhostReason, &ghost.PurgeStatus, &ghost.DetectionDate,
			&ghost.Notes, &ghost.CurrentStatus, &ghost.SubscriptionCount,
			&ghost.InvoiceCount,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan ghost customer"})
			return
		}
		ghosts = append(ghosts, ghost)
	}

	c.JSON(http.StatusOK, gin.H{
		"ghosts": ghosts,
		"total":  len(ghosts),
	})
}

// GetGhostSummary returns summary statistics of ghost detection
func (h *GhostCustomerHandler) GetGhostSummary(c *gin.Context) {
	query := `
		SELECT 
			COUNT(*) as total_ghosts,
			COUNT(*) FILTER (WHERE ghost_reason LIKE '%hash_id%') as hash_id_ghosts,
			COUNT(*) FILTER (WHERE ghost_reason LIKE '%invalid%') as invalid_format_ghosts,
			COUNT(*) FILTER (WHERE purge_status = 'marked_for_purge') as marked_for_purge,
			COUNT(*) FILTER (WHERE purge_status = 'purged') as already_purged
		FROM stripe_ghosts
	`

	var summary GhostDetectionSummary
	err := h.db.QueryRow(query).Scan(
		&summary.TotalGhosts, &summary.HashIDGhosts, &summary.InvalidFormatGhosts,
		&summary.MarkedForPurge, &summary.AlreadyPurged,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get ghost summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// RunGhostDetection manually triggers ghost detection
func (h *GhostCustomerHandler) RunGhostDetection(c *gin.Context) {
	// Call the ghost detection function
	query := `SELECT * FROM detect_ghost_customers()`

	rows, err := h.db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to run ghost detection"})
		return
	}
	defer rows.Close()

	var newGhosts []GhostCustomer
	for rows.Next() {
		var customerID int
		var stripeID, email, name, reason string

		err := rows.Scan(&customerID, &stripeID, &email, &name, &reason)
		if err != nil {
			continue
		}

		// Insert into ghosts table if not already exists
		insertQuery := `
			INSERT INTO stripe_ghosts (
				local_customer_id, stripe_customer_id, customer_email, 
				customer_name, ghost_type, ghost_reason, purge_status
			) VALUES ($1, $2, $3, $4, 'customer', $5, 'detected')
			ON CONFLICT (stripe_customer_id) DO NOTHING
		`

		h.db.Exec(insertQuery, customerID, stripeID, email, name, reason)

		newGhosts = append(newGhosts, GhostCustomer{
			LocalCustomerID:  &customerID,
			StripeCustomerID: stripeID,
			CustomerEmail:    email,
			CustomerName:     name,
			GhostReason:      reason,
			PurgeStatus:      "detected",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Ghost detection completed",
		"new_ghosts": len(newGhosts),
		"ghosts":     newGhosts,
	})
}

// MarkGhostForPurge marks a ghost customer for purging
func (h *GhostCustomerHandler) MarkGhostForPurge(c *gin.Context) {
	stripeCustomerID := c.Param("stripe_customer_id")

	type PurgeRequest struct {
		Reason string `json:"reason"`
	}

	var req PurgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Reason == "" {
		req.Reason = "admin_decision"
	}

	// Call the mark_customer_for_purge function
	query := `SELECT mark_customer_for_purge($1, $2)`

	var success bool
	err := h.db.QueryRow(query, stripeCustomerID, req.Reason).Scan(&success)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark customer for purge"})
		return
	}

	if !success {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "Customer marked for purge",
		"stripe_customer_id": stripeCustomerID,
		"reason":             req.Reason,
	})
}

// PurgeGhostCustomer permanently deletes a ghost customer and all related data
func (h *GhostCustomerHandler) PurgeGhostCustomer(c *gin.Context) {
	stripeCustomerID := c.Param("stripe_customer_id")

	type PurgeRequest struct {
		AdminUser string `json:"admin_user"`
		Confirm   bool   `json:"confirm"`
	}

	var req PurgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if !req.Confirm {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Confirmation required for purging"})
		return
	}

	if req.AdminUser == "" {
		req.AdminUser = "unknown_admin"
	}

	// Call the purge_ghost_customer function
	query := `SELECT purge_ghost_customer($1, $2)`

	var success bool
	err := h.db.QueryRow(query, stripeCustomerID, req.AdminUser).Scan(&success)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to purge customer"})
		return
	}

	if !success {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found or purge failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "Customer successfully purged",
		"stripe_customer_id": stripeCustomerID,
		"admin_user":         req.AdminUser,
	})
}

// BulkPurgeGhosts purges multiple ghost customers at once
func (h *GhostCustomerHandler) BulkPurgeGhosts(c *gin.Context) {
	type BulkPurgeRequest struct {
		CustomerIDs []string `json:"customer_ids"`
		AdminUser   string   `json:"admin_user"`
		Confirm     bool     `json:"confirm"`
	}

	var req BulkPurgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if !req.Confirm {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Confirmation required for bulk purging"})
		return
	}

	if req.AdminUser == "" {
		req.AdminUser = "unknown_admin"
	}

	var results []map[string]interface{}
	successCount := 0

	for _, customerID := range req.CustomerIDs {
		query := `SELECT purge_ghost_customer($1, $2)`

		var success bool
		err := h.db.QueryRow(query, customerID, req.AdminUser).Scan(&success)

		result := map[string]interface{}{
			"stripe_customer_id": customerID,
			"success":            success && err == nil,
		}

		if err != nil {
			result["error"] = err.Error()
		}

		if success && err == nil {
			successCount++
		}

		results = append(results, result)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Bulk purge completed",
		"total":      len(req.CustomerIDs),
		"successful": successCount,
		"failed":     len(req.CustomerIDs) - successCount,
		"results":    results,
	})
}

// RegisterGhostCustomerRoutes registers all ghost customer management routes
func RegisterGhostCustomerRoutes(router *gin.RouterGroup, handler *GhostCustomerHandler) {
	ghosts := router.Group("/ghosts")
	{
		// Get all ghost customers
		ghosts.GET("/customers", handler.GetGhostCustomers)

		// Get ghost detection summary
		ghosts.GET("/summary", handler.GetGhostSummary)

		// Run ghost detection manually
		ghosts.POST("/detect", handler.RunGhostDetection)

		// Mark specific customer for purge
		ghosts.PUT("/customers/:stripe_customer_id/mark-purge", handler.MarkGhostForPurge)

		// Purge specific customer
		ghosts.DELETE("/customers/:stripe_customer_id", handler.PurgeGhostCustomer)

		// Bulk purge multiple customers
		ghosts.POST("/bulk-purge", handler.BulkPurgeGhosts)
	}
}
