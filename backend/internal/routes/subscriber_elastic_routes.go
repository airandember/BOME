package routes

import (
	"bome-backend/internal/database"
	"bome-backend/internal/handlers"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupSubscriberElasticRoutes sets up routes for the unified subscriber elastic service
// 🔒 SECURITY: All routes are ADMIN-ONLY - protected by AdminRequired() middleware
func SetupSubscriberElasticRoutes(router *gin.RouterGroup, db *database.DB) {
	// Create service and handler
	elasticService := services.NewSubscriberElasticService(db)
	elasticHandler := handlers.NewSubscriberElasticHandler(elasticService)

	// Create routes group with ADMIN PROTECTION
	// 🔒 Only admins and super_admins can access subscriber elastic data
	elastic := router.Group("/subscriber-elastic")
	elastic.Use(middleware.AuthRequired())  // 🔒 STEP 1: Authenticate & set user_role in context
	elastic.Use(middleware.AdminRequired()) // 🔒 STEP 2: Check if role is admin
	{
		// Main data endpoints - ADMIN ONLY
		elastic.GET("/subscribers", elasticHandler.GetAllUnifiedSubscribers)
		elastic.GET("/subscribers/email/:email", elasticHandler.GetUnifiedSubscriberByEmail)
		elastic.GET("/subscribers/id/:id", elasticHandler.GetUnifiedSubscriberByID)

		// Diagnostic endpoints - ADMIN ONLY
		elastic.GET("/diagnose", elasticHandler.DiagnoseSubscriberIssues)
		elastic.GET("/multiple-stripe-customers", elasticHandler.GetSubscribersWithMultipleStripeCustomers)
		elastic.GET("/active-plan-no-access", elasticHandler.GetSubscribersWithActivePlansButNoAccess)
		elastic.GET("/manual-access", elasticHandler.GetSubscribersWithVideoAccessButNoPlan)

		// Statistics - ADMIN ONLY
		elastic.GET("/stats", elasticHandler.GetSubscriberStats)

		// Actions - ADMIN ONLY
		elastic.PUT("/subscribers/:id/manual-access", elasticHandler.UpdateManualVideoAccess)
	}
}
