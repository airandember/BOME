package routes

import (
	"net/http"
	"strconv"

	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupCustomerLinkingRoutes registers customer linking routes
func SetupCustomerLinkingRoutes(router *gin.RouterGroup, db *database.DB) {
	linkingService := services.NewCustomerLinkingService(db)

	// Admin-only routes
	linking := router.Group("/customer-linking")
	linking.Use(middleware.AuthRequired())
	linking.Use(middleware.AdminRequired())

	// Get linking statistics
	linking.GET("/stats", func(c *gin.Context) {
		stats, err := linkingService.GetLinkingStats()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get linking stats",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"stats": stats,
		})
	})

	// Get unlinked customers
	linking.GET("/unlinked", func(c *gin.Context) {
		unlinked, err := linkingService.GetUnlinkedCustomers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get unlinked customers",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"unlinked_customers": unlinked,
			"count":              len(unlinked),
		})
	})

	// Link a specific user to their customers
	linking.POST("/user/:user_id", func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Param("user_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user ID",
			})
			return
		}

		result, err := linkingService.LinkUserToCustomers(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to link user to customers",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"result": result,
		})
	})

	// Link all users to their customers
	linking.POST("/all", func(c *gin.Context) {
		results, err := linkingService.LinkAllUsers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to link all users",
				"details": err.Error(),
			})
			return
		}

		// Count successes and errors
		successCount := 0
		errorCount := 0
		for _, r := range results {
			if r.Error != "" {
				errorCount++
			} else if r.CustomersLinked > 0 {
				successCount++
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message":         "Linking complete",
			"success_count":   successCount,
			"error_count":     errorCount,
			"total_processed": len(results),
			"results":         results,
		})
	})

	// Get customers for a specific user
	linking.GET("/user/:user_id/customers", func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Param("user_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user ID",
			})
			return
		}

		customers, err := linkingService.GetUserCustomers(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get user customers",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id":   userID,
			"customers": customers,
			"count":     len(customers),
		})
	})

	// Set primary customer for a user
	linking.PUT("/user/:user_id/primary", func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Param("user_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user ID",
			})
			return
		}

		var req struct {
			StripeCustomerID string `json:"stripe_customer_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "stripe_customer_id is required",
			})
			return
		}

		err = linkingService.SetPrimaryCustomer(userID, req.StripeCustomerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to set primary customer",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":          "Primary customer updated",
			"user_id":          userID,
			"primary_customer": req.StripeCustomerID,
		})
	})
}
