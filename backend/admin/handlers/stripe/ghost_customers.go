package stripe

import (
	"log"
	"net/http"

	"bome-backend/infrastructure/database"
	stripeServices "bome-backend/services/stripe"

	"github.com/gin-gonic/gin"
)

// GhostCustomersHandler handles ghost customer management
type GhostCustomersHandler struct {
	service *stripeServices.GhostCustomersService
}

// NewGhostCustomersHandler creates a new ghost customers handler
func NewGhostCustomersHandler(service *stripeServices.GhostCustomersService) *GhostCustomersHandler {
	return &GhostCustomersHandler{
		service: service,
	}
}

// GetGhostCustomers returns all detected ghost customers
func (h *GhostCustomersHandler) GetGhostCustomers(c *gin.Context) {
	log.Println("🔍 [GHOST-HANDLER] Fetching all ghost customers...")

	ghosts, err := h.service.GetAllGhosts()
	if err != nil {
		log.Printf("❌ [GHOST-HANDLER] Failed to get ghosts: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query ghost customers"})
		return
	}

	log.Printf("✅ [GHOST-HANDLER] Returned %d ghost customers", len(ghosts))
	c.JSON(http.StatusOK, gin.H{
		"ghosts": ghosts,
		"total":  len(ghosts),
	})
}

// GetGhostSummary returns summary statistics of ghost detection
func (h *GhostCustomersHandler) GetGhostSummary(c *gin.Context) {
	log.Println("📊 [GHOST-HANDLER] Fetching ghost summary...")

	summary, err := h.service.GetGhostSummary()
	if err != nil {
		log.Printf("❌ [GHOST-HANDLER] Failed to get summary: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get ghost summary"})
		return
	}

	log.Printf("✅ [GHOST-HANDLER] Summary: %d total ghosts", summary.TotalGhosts)
	c.JSON(http.StatusOK, summary)
}

// RunGhostDetection manually triggers ghost detection
func (h *GhostCustomersHandler) RunGhostDetection(c *gin.Context) {
	log.Println("🔍 [GHOST-HANDLER] Running ghost detection...")

	newGhosts, err := h.service.RunGhostDetection()
	if err != nil {
		log.Printf("❌ [GHOST-HANDLER] Failed to run detection: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to run ghost detection"})
		return
	}

	log.Printf("✅ [GHOST-HANDLER] Detected %d new ghosts", len(newGhosts))
	c.JSON(http.StatusOK, gin.H{
		"message":    "Ghost detection completed",
		"new_ghosts": len(newGhosts),
		"ghosts":     newGhosts,
	})
}

// MarkGhostForPurge marks a ghost customer for purging
func (h *GhostCustomersHandler) MarkGhostForPurge(c *gin.Context) {
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

	log.Printf("⚠️ [GHOST-HANDLER] Marking ghost for purge: %s", stripeCustomerID)

	err := h.service.MarkGhostForPurge(stripeCustomerID, req.Reason)
	if err != nil {
		log.Printf("❌ [GHOST-HANDLER] Failed to mark for purge: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark customer for purge"})
		return
	}

	log.Printf("✅ [GHOST-HANDLER] Marked for purge: %s", stripeCustomerID)
	c.JSON(http.StatusOK, gin.H{
		"message":            "Customer marked for purge",
		"stripe_customer_id": stripeCustomerID,
		"reason":             req.Reason,
	})
}

// PurgeGhostCustomer permanently deletes a ghost customer and all related data
func (h *GhostCustomersHandler) PurgeGhostCustomer(c *gin.Context) {
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

	log.Printf("🗑️ [GHOST-HANDLER] PURGING ghost customer: %s (by: %s)", stripeCustomerID, req.AdminUser)

	err := h.service.PurgeGhostCustomer(stripeCustomerID, req.AdminUser)
	if err != nil {
		log.Printf("❌ [GHOST-HANDLER] Purge failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to purge customer"})
		return
	}

	log.Printf("✅ [GHOST-HANDLER] Successfully purged: %s", stripeCustomerID)
	c.JSON(http.StatusOK, gin.H{
		"message":            "Customer successfully purged",
		"stripe_customer_id": stripeCustomerID,
		"admin_user":         req.AdminUser,
	})
}

// BulkPurgeGhosts purges multiple ghost customers at once
func (h *GhostCustomersHandler) BulkPurgeGhosts(c *gin.Context) {
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

	log.Printf("🗑️ [GHOST-HANDLER] Bulk purging %d customers (by: %s)", len(req.CustomerIDs), req.AdminUser)

	results, err := h.service.BulkPurgeGhosts(req.CustomerIDs, req.AdminUser)
	if err != nil {
		log.Printf("❌ [GHOST-HANDLER] Bulk purge failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bulk purge customers"})
		return
	}

	// Count successes
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	log.Printf("✅ [GHOST-HANDLER] Bulk purge complete: %d/%d successful", successCount, len(req.CustomerIDs))
	c.JSON(http.StatusOK, gin.H{
		"message":    "Bulk purge completed",
		"total":      len(req.CustomerIDs),
		"successful": successCount,
		"failed":     len(req.CustomerIDs) - successCount,
		"results":    results,
	})
}

// SetupGhostCustomerRoutes registers all ghost customer management routes
func SetupGhostCustomerRoutes(stripeGroup *gin.RouterGroup, db *database.DB) {
	// Initialize service
	service := stripeServices.NewGhostCustomersService(db)
	handler := NewGhostCustomersHandler(service)

	// Ghost customer routes (under /admin/stripe/ghosts/)
	ghostsGroup := stripeGroup.Group("/ghosts")
	{
		// Get all ghost customers
		ghostsGroup.GET("/customers", handler.GetGhostCustomers)

		// Get ghost detection summary
		ghostsGroup.GET("/summary", handler.GetGhostSummary)

		// Run ghost detection manually
		ghostsGroup.POST("/detect", handler.RunGhostDetection)

		// Mark specific customer for purge
		ghostsGroup.PUT("/customers/:stripe_customer_id/mark-purge", handler.MarkGhostForPurge)

		// Purge specific customer
		ghostsGroup.DELETE("/customers/:stripe_customer_id", handler.PurgeGhostCustomer)

		// Bulk purge multiple customers
		ghostsGroup.POST("/bulk-purge", handler.BulkPurgeGhosts)
	}

	log.Println("✅ [GHOST-ROUTES] Registered ghost customer routes at /admin/stripe/ghosts/")
}
