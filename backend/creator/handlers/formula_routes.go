package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"bome-backend/creator/models"
	"bome-backend/creator/services"
	"bome-backend/infrastructure/database"
)

// SetupPayoutFormulaRoutes registers all payout formula-related routes
func SetupPayoutFormulaRoutes(router *gin.RouterGroup, db *database.DB) {
	service := services.NewPayoutFormulaService(db)
	
	// Formula CRUD
	router.GET("/payout-formulas", GetFormulasHandler(service))
	router.POST("/payout-formulas", CreateFormulaHandler(service))
	router.GET("/payout-formulas/default", GetDefaultFormulaHandler(service))
	router.GET("/payout-formulas/:id", GetFormulaByIDHandler(service))
	router.PUT("/payout-formulas/:id", UpdateFormulaHandler(service))
	router.DELETE("/payout-formulas/:id", DeleteFormulaHandler(service))
	router.POST("/payout-formulas/:id/set-default", SetDefaultFormulaHandler(service))
	
	log.Println("[FORMULA-ROUTES] ✅ Registered 7 payout formula endpoints")
}

// GetFormulasHandler retrieves all payout formulas
func GetFormulasHandler(service *services.PayoutFormulaService) gin.HandlerFunc {
	return func(c *gin.Context) {
		activeOnly := c.Query("active") == "true"
		
		formulas, err := service.GetFormulas(activeOnly)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"formulas": formulas,
			"count":    len(formulas),
		})
	}
}

// CreateFormulaHandler creates a new payout formula
func CreateFormulaHandler(service *services.PayoutFormulaService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.CreatePayoutFormulaInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Get admin user ID from context
		adminUserID, exists := c.Get("user_id")
		if exists {
			input.CreatedBy = adminUserID.(int)
		}
		
		formula, err := service.CreateFormula(&input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusCreated, formula)
	}
}

// GetDefaultFormulaHandler retrieves the default payout formula
func GetDefaultFormulaHandler(service *services.PayoutFormulaService) gin.HandlerFunc {
	return func(c *gin.Context) {
		formula, err := service.GetDefaultFormula()
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no default formula configured"})
			return
		}
		
		c.JSON(http.StatusOK, formula)
	}
}

// GetFormulaByIDHandler retrieves a payout formula by ID
func GetFormulaByIDHandler(service *services.PayoutFormulaService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid formula ID"})
			return
		}
		
		formula, err := service.GetFormulaByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "formula not found"})
			return
		}
		
		c.JSON(http.StatusOK, formula)
	}
}

// UpdateFormulaHandler updates a payout formula
func UpdateFormulaHandler(service *services.PayoutFormulaService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid formula ID"})
			return
		}
		
		var input models.UpdatePayoutFormulaInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		formula, err := service.UpdateFormula(id, &input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, formula)
	}
}

// DeleteFormulaHandler soft-deletes a payout formula
func DeleteFormulaHandler(service *services.PayoutFormulaService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid formula ID"})
			return
		}
		
		err = service.DeleteFormula(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{"message": "Formula deleted successfully"})
	}
}

// SetDefaultFormulaHandler sets a formula as the default
func SetDefaultFormulaHandler(service *services.PayoutFormulaService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid formula ID"})
			return
		}
		
		err = service.SetDefaultFormula(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{"message": "Default formula set successfully"})
	}
}

