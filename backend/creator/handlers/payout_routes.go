package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"bome-backend/creator/models"
	"bome-backend/creator/services"
	"bome-backend/infrastructure/database"
)

// SetupPayoutRoutes registers all payout-related routes
func SetupPayoutRoutes(router *gin.RouterGroup, db *database.DB) {
	service := services.NewPayoutService(db)

	// Payout Generation & Calculation
	router.POST("/payouts/generate", GenerateMonthlyPayoutsHandler(service))
	router.POST("/payouts/calculate", CalculatePresenterPayoutHandler(service))

	// Payout Management
	router.GET("/payouts/:id", GetPayoutByIDHandler(service))
	router.GET("/payouts/presenter/:presenter_id", GetPresenterPayoutsHandler(service))
	router.GET("/payouts/month/:month", GetPayoutsByMonthHandler(service))
	router.GET("/payouts/month/:month/summary", GetPayoutSummaryHandler(service))

	// Payout Actions
	router.PUT("/payouts/:id/status", UpdatePayoutStatusHandler(service))
	router.PUT("/payouts/:id/amounts", UpdatePayoutAmountsHandler(service))
	router.POST("/payouts/approve", ApprovePayoutsHandler(service))
	router.DELETE("/payouts/:id", DeletePayoutHandler(service))

	// Transaction Management
	router.POST("/payout-transactions", CreateTransactionHandler(service))
	router.GET("/payout-transactions/payout/:payout_id", GetTransactionsByPayoutHandler(service))
	router.GET("/payout-transactions/presenter/:presenter_id", GetTransactionsByPresenterHandler(service))
	router.GET("/payout-transactions/recent", GetRecentTransactionsHandler(service))
	router.PUT("/payout-transactions/:id/status", UpdateTransactionStatusHandler(service))

	log.Println("[PAYOUT-ROUTES] ✅ Registered 15 payout management endpoints")
}

// GenerateMonthlyPayoutsHandler generates payouts for all presenters for a month
func GenerateMonthlyPayoutsHandler(service *services.PayoutService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Month string `json:"month" binding:"required"` // Format: "2006-01-02" or "2006-01"
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Parse month
		month, err := time.Parse("2006-01-02", input.Month)
		if err != nil {
			// Try alternate format
			month, err = time.Parse("2006-01", input.Month)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month format, use YYYY-MM-DD or YYYY-MM"})
				return
			}
		}

		// Get admin user ID from context
		adminUserID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		result, err := service.GenerateMonthlyPayouts(month, adminUserID.(int))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// CalculatePresenterPayoutHandler calculates payout for a specific presenter
func CalculatePresenterPayoutHandler(service *services.PayoutService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			PresenterID int    `json:"presenter_id" binding:"required"`
			Month       string `json:"month" binding:"required"`
			FormulaID   int    `json:"formula_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Parse month
		month, err := time.Parse("2006-01-02", input.Month)
		if err != nil {
			month, err = time.Parse("2006-01", input.Month)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month format"})
				return
			}
		}

		result, err := service.CalculatePresenterPayout(input.PresenterID, month, input.FormulaID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// GetPayoutByIDHandler retrieves a payout by ID
func GetPayoutByIDHandler(service *services.PayoutService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payout ID"})
			return
		}

		payout, err := service.GetPayoutByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "payout not found"})
			return
		}

		c.JSON(http.StatusOK, payout)
	}
}

// GetPresenterPayoutsHandler retrieves all payouts for a presenter
func GetPresenterPayoutsHandler(service *services.PayoutService) gin.HandlerFunc {
	return func(c *gin.Context) {
		presenterIDParam := c.Param("presenter_id")
		presenterID, err := strconv.Atoi(presenterIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid presenter ID"})
			return
		}

		payouts, err := service.GetPresenterPayouts(presenterID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"payouts": payouts,
			"count":   len(payouts),
		})
	}
}

// GetPayoutsByMonthHandler retrieves all payouts for a month
func GetPayoutsByMonthHandler(service *services.PayoutService) gin.HandlerFunc {
	return func(c *gin.Context) {
		monthParam := c.Param("month")

		// Parse month (YYYY-MM-DD or YYYY-MM)
		month, err := time.Parse("2006-01-02", monthParam)
		if err != nil {
			month, err = time.Parse("2006-01", monthParam)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month format"})
				return
			}
		}

		status := c.Query("status") // Optional filter

		payouts, err := service.GetPayoutsByMonth(month, status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"payouts": payouts,
			"count":   len(payouts),
			"month":   month.Format("2006-01"),
		})
	}
}

// GetPayoutSummaryHandler retrieves summary statistics for a month
func GetPayoutSummaryHandler(service *services.PayoutService) gin.HandlerFunc {
	return func(c *gin.Context) {
		monthParam := c.Param("month")

		// Parse month
		month, err := time.Parse("2006-01-02", monthParam)
		if err != nil {
			month, err = time.Parse("2006-01", monthParam)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month format"})
				return
			}
		}

		summary, err := service.GetPayoutSummary(month)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, summary)
	}
}

// UpdatePayoutStatusHandler updates the status of a payout
func UpdatePayoutStatusHandler(service *services.PayoutService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payout ID"})
			return
		}

		var input models.UpdatePayoutStatusInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get admin user ID from context
		adminUserID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}
		input.UpdatedBy = adminUserID.(int)

		payout, err := service.UpdatePayoutStatus(id, &input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, payout)
	}
}

// UpdatePayoutAmountsHandler updates adjustment amounts for a payout
func UpdatePayoutAmountsHandler(service *services.PayoutService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payout ID"})
			return
		}

		var input models.UpdatePayoutAmountsInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get admin user ID from context
		adminUserID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}
		input.UpdatedBy = adminUserID.(int)

		payout, err := service.UpdatePayoutAmounts(id, &input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, payout)
	}
}

// ApprovePayoutsHandler approves multiple payouts
func ApprovePayoutsHandler(service *services.PayoutService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			PayoutIDs []int `json:"payout_ids" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get admin user ID from context
		adminUserID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		err := service.ApprovePayouts(input.PayoutIDs, adminUserID.(int))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":        "Payouts approved successfully",
			"approved_count": len(input.PayoutIDs),
		})
	}
}

// DeletePayoutHandler deletes a payout
func DeletePayoutHandler(service *services.PayoutService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payout ID"})
			return
		}

		err = service.DeletePayout(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Payout deleted successfully"})
	}
}

// CreateTransactionHandler creates a payout transaction
func CreateTransactionHandler(service *services.PayoutService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.CreatePayoutTransactionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get admin user ID from context
		adminUserID, exists := c.Get("user_id")
		if exists {
			input.ProcessedBy = adminUserID.(int)
		}

		transaction, err := service.CreateTransaction(&input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, transaction)
	}
}

// GetTransactionsByPayoutHandler retrieves all transactions for a payout
func GetTransactionsByPayoutHandler(service *services.PayoutService) gin.HandlerFunc {
	return func(c *gin.Context) {
		payoutIDParam := c.Param("payout_id")
		payoutID, err := strconv.Atoi(payoutIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payout ID"})
			return
		}

		transactions, err := service.GetTransactionsByPayoutID(payoutID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"transactions": transactions,
			"count":        len(transactions),
		})
	}
}

// GetTransactionsByPresenterHandler retrieves all transactions for a presenter
func GetTransactionsByPresenterHandler(service *services.PayoutService) gin.HandlerFunc {
	return func(c *gin.Context) {
		presenterIDParam := c.Param("presenter_id")
		presenterID, err := strconv.Atoi(presenterIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid presenter ID"})
			return
		}

		transactions, err := service.GetTransactionsByPresenterID(presenterID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"transactions": transactions,
			"count":        len(transactions),
		})
	}
}

// GetRecentTransactionsHandler retrieves recent transactions
func GetRecentTransactionsHandler(service *services.PayoutService) gin.HandlerFunc {
	return func(c *gin.Context) {
		limitParam := c.DefaultQuery("limit", "50")
		limit, err := strconv.Atoi(limitParam)
		if err != nil {
			limit = 50
		}

		status := c.Query("status") // Optional filter

		transactions, err := service.GetRecentTransactions(limit, status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"transactions": transactions,
			"count":        len(transactions),
		})
	}
}

// UpdateTransactionStatusHandler updates the status of a transaction
func UpdateTransactionStatusHandler(service *services.PayoutService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction ID"})
			return
		}

		var input models.UpdateTransactionStatusInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get admin user ID from context
		adminUserID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}
		input.ProcessedBy = adminUserID.(int)

		transaction, err := service.UpdateTransactionStatus(id, &input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, transaction)
	}
}
