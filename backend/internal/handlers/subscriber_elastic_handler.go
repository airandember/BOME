package handlers

import (
	"net/http"
	"strconv"

	"bome-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// SubscriberElasticHandler handles requests for unified subscriber data
type SubscriberElasticHandler struct {
	service *services.SubscriberElasticService
}

// NewSubscriberElasticHandler creates a new handler instance
func NewSubscriberElasticHandler(service *services.SubscriberElasticService) *SubscriberElasticHandler {
	return &SubscriberElasticHandler{service: service}
}

// GetAllUnifiedSubscribers returns all subscribers with complete data
func (h *SubscriberElasticHandler) GetAllUnifiedSubscribers(c *gin.Context) {
	subscribers, err := h.service.GetAllUnifiedSubscribers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve unified subscribers",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"subscribers": subscribers,
			"count": len(subscribers),
		},
	})
}

// GetUnifiedSubscriberByEmail returns a specific subscriber by email
func (h *SubscriberElasticHandler) GetUnifiedSubscriberByEmail(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email parameter is required",
		})
		return
	}

	subscriber, err := h.service.GetUnifiedSubscriberByEmail(email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Subscriber not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": subscriber,
	})
}

// GetUnifiedSubscriberByID returns a specific subscriber by ID
func (h *SubscriberElasticHandler) GetUnifiedSubscriberByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid subscriber ID",
		})
		return
	}

	subscriber, err := h.service.GetUnifiedSubscriberByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Subscriber not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": subscriber,
	})
}

// GetSubscribersWithMultipleStripeCustomers returns subscribers with multiple Stripe customer IDs
func (h *SubscriberElasticHandler) GetSubscribersWithMultipleStripeCustomers(c *gin.Context) {
	subscribers, err := h.service.GetSubscribersWithMultipleStripeCustomers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve subscribers with multiple Stripe customers",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"subscribers": subscribers,
			"count": len(subscribers),
			"description": "Subscribers with multiple Stripe customer IDs (potential duplicates)",
		},
	})
}

// GetSubscribersWithActivePlansButNoAccess returns subscribers who should have access but don't
func (h *SubscriberElasticHandler) GetSubscribersWithActivePlansButNoAccess(c *gin.Context) {
	subscribers, err := h.service.GetSubscribersWithActivePlansButNoAccess()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve subscribers with access issues",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"subscribers": subscribers,
			"count": len(subscribers),
			"description": "Subscribers with active plans but no video access (potential bugs)",
		},
	})
}

// GetSubscribersWithVideoAccessButNoPlan returns subscribers with manual video access
func (h *SubscriberElasticHandler) GetSubscribersWithVideoAccessButNoPlan(c *gin.Context) {
	subscribers, err := h.service.GetSubscribersWithVideoAccessButNoPlan()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve subscribers with manual access",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"subscribers": subscribers,
			"count": len(subscribers),
			"description": "Subscribers with manual video access (no active plan)",
		},
	})
}

// UpdateManualVideoAccess updates manual video access for a subscriber
func (h *SubscriberElasticHandler) UpdateManualVideoAccess(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid subscriber ID",
		})
		return
	}

	var request struct {
		HasAccess bool `json:"has_access"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	err = h.service.UpdateManualVideoAccess(id, request.HasAccess)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update manual video access",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Manual video access updated successfully",
		"data": gin.H{
			"user_id": id,
			"has_access": request.HasAccess,
		},
	})
}

// GetSubscriberStats returns comprehensive statistics about subscribers
func (h *SubscriberElasticHandler) GetSubscriberStats(c *gin.Context) {
	stats, err := h.service.GetSubscriberStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve subscriber statistics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": stats,
	})
}

// DiagnoseSubscriberIssues returns a comprehensive diagnosis of subscriber data issues
func (h *SubscriberElasticHandler) DiagnoseSubscriberIssues(c *gin.Context) {
	// Get all diagnostic data
	multipleCustomers, err1 := h.service.GetSubscribersWithMultipleStripeCustomers()
	noAccess, err2 := h.service.GetSubscribersWithActivePlansButNoAccess()
	manualAccess, err3 := h.service.GetSubscribersWithVideoAccessButNoPlan()
	stats, err4 := h.service.GetSubscriberStats()

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve diagnostic data",
			"details": map[string]string{
				"multiple_customers": getErrorString(err1),
				"no_access": getErrorString(err2),
				"manual_access": getErrorString(err3),
				"stats": getErrorString(err4),
			},
		})
		return
	}

	diagnosis := gin.H{
		"summary": gin.H{
			"total_subscribers": stats["total_subscribers"],
			"issues_found": len(multipleCustomers) + len(noAccess),
			"manual_overrides": len(manualAccess),
		},
		"issues": gin.H{
			"multiple_stripe_customers": gin.H{
				"count": len(multipleCustomers),
				"description": "Users with multiple Stripe customer IDs (potential duplicates)",
				"subscribers": multipleCustomers,
			},
			"active_plan_no_access": gin.H{
				"count": len(noAccess),
				"description": "Users with active plans but no video access (potential bugs)",
				"subscribers": noAccess,
			},
		},
		"manual_overrides": gin.H{
			"count": len(manualAccess),
			"description": "Users with manual video access (no active plan)",
			"subscribers": manualAccess,
		},
		"statistics": stats,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": diagnosis,
	})
}

// Helper function to convert error to string
func getErrorString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
