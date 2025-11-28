package routes

import (
	"bome-backend/internal/database"
	"bome-backend/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RegisterRevenueAttributionRoutes registers all revenue attribution routes
func RegisterRevenueAttributionRoutes(router *gin.RouterGroup, db *database.DB) {
	service := services.NewRevenueAttributionService(db)

	// Formula management routes (admin only)
	formulas := router.Group("/attribution/formulas")
	{
		// Get all formulas
		formulas.GET("", func(c *gin.Context) {
			activeOnly := c.DefaultQuery("active_only", "false") == "true"
			formulas, err := service.GetAllFormulas(activeOnly)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch formulas"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"formulas": formulas})
		})

		// Get single formula
		formulas.GET("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid formula ID"})
				return
			}

			formula, err := service.GetFormula(id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Formula not found"})
				return
			}
			c.JSON(http.StatusOK, formula)
		})

		// Create new formula (admin only)
		formulas.POST("", func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
				return
			}

			var req services.CreateFormulaRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			formula, err := service.CreateFormula(req, userID.(int))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, formula)
		})

		// Update formula (admin only)
		formulas.PATCH("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid formula ID"})
				return
			}

			var req services.UpdateFormulaRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if err := service.UpdateFormula(id, req); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Formula updated successfully"})
		})

		// Delete formula (admin only)
		formulas.DELETE("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid formula ID"})
				return
			}

			if err := service.DeleteFormula(id); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Formula deleted successfully"})
		})
	}

	// Attribution calculation routes
	attribution := router.Group("/attribution")
	{
		// Calculate attribution for a subscription (internal/webhook use)
		attribution.POST("/calculate", func(c *gin.Context) {
			var req struct {
				UserID            int     `json:"user_id" binding:"required"`
				SubscriptionID    string  `json:"subscription_id" binding:"required"`
				SubscriptionValue float64 `json:"subscription_value" binding:"required"`
				FormulaID         *int    `json:"formula_id,omitempty"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			err := service.CalculateAttribution(req.UserID, req.SubscriptionID, req.SubscriptionValue, req.FormulaID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Attribution calculated successfully"})
		})

		// Get video conversion metrics
		attribution.GET("/video/:videoId/metrics", func(c *gin.Context) {
			videoID, err := strconv.Atoi(c.Param("videoId"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
				return
			}

			formulaID, err := strconv.Atoi(c.DefaultQuery("formula_id", "1"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid formula ID"})
				return
			}

			metrics, err := service.GetVideoConversionMetrics(videoID, formulaID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Metrics not found"})
				return
			}

			c.JSON(http.StatusOK, metrics)
		})

		// Get top converting videos
		attribution.GET("/top-videos", func(c *gin.Context) {
			formulaID, err := strconv.Atoi(c.DefaultQuery("formula_id", "1"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid formula ID"})
				return
			}

			limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
			if err != nil || limit <= 0 {
				limit = 10
			}

			sortBy := c.DefaultQuery("sort_by", "revenue")

			videos, err := service.GetTopConvertingVideos(formulaID, limit, sortBy)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch top videos"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"videos": videos, "count": len(videos)})
		})

		// Get comprehensive attribution report
		attribution.GET("/report", func(c *gin.Context) {
			formulaID, err := strconv.Atoi(c.DefaultQuery("formula_id", "1"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid formula ID"})
				return
			}

			periodDays, err := strconv.Atoi(c.DefaultQuery("period_days", "30"))
			if err != nil || periodDays <= 0 {
				periodDays = 30
			}

			report, err := service.GetAttributionReport(formulaID, periodDays)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate report"})
				return
			}

			c.JSON(http.StatusOK, report)
		})
	}

	// Preview/test formula (admin only)
	router.POST("/attribution/preview", func(c *gin.Context) {
		var req struct {
			FormulaID int `json:"formula_id" binding:"required"`
			UserID    int `json:"user_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get formula
		formula, err := service.GetFormula(req.FormulaID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Formula not found"})
			return
		}

		// This would show what attribution would look like for a user
		// For now, return formula details
		c.JSON(http.StatusOK, gin.H{
			"formula": formula,
			"message": "Preview functionality - shows how attribution would be calculated",
		})
	})
}

