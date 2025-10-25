package routes

import (
	"bome-backend/internal/database"
	"bome-backend/internal/handlers"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupSubscriberElasticRoutes sets up routes for the unified subscriber elastic service
func SetupSubscriberElasticRoutes(router *gin.RouterGroup, db *database.DB) {
	// Create service and handler
	elasticService := services.NewSubscriberElasticService(db)
	elasticHandler := handlers.NewSubscriberElasticHandler(elasticService)

	// Create routes group
	elastic := router.Group("/subscriber-elastic")
	{
		// Main data endpoints
		elastic.GET("/subscribers", elasticHandler.GetAllUnifiedSubscribers)
		elastic.GET("/subscribers/email/:email", elasticHandler.GetUnifiedSubscriberByEmail)
		elastic.GET("/subscribers/id/:id", elasticHandler.GetUnifiedSubscriberByID)
		
		// Diagnostic endpoints
		elastic.GET("/diagnose", elasticHandler.DiagnoseSubscriberIssues)
		elastic.GET("/multiple-stripe-customers", elasticHandler.GetSubscribersWithMultipleStripeCustomers)
		elastic.GET("/active-plan-no-access", elasticHandler.GetSubscribersWithActivePlansButNoAccess)
		elastic.GET("/manual-access", elasticHandler.GetSubscribersWithVideoAccessButNoPlan)
		
		// Statistics
		elastic.GET("/stats", elasticHandler.GetSubscriberStats)
		
		// Actions
		elastic.PUT("/subscribers/:id/manual-access", elasticHandler.UpdateManualVideoAccess)
	}
}
