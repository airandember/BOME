package routes

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupSubscriptionOfferRoutes sets up all subscription offer routes
func SetupSubscriptionOfferRoutes(router *gin.Engine, db *database.DB, offersService *services.SubscriptionOffersService) {
	log.Printf("Setting up subscription offer routes...")

	// Create admin group with authentication
	admin := router.Group("/api/v1/admin/subscription-offers")
	admin.Use(middleware.AuthRequired())
	admin.Use(middleware.AdminRequired())

	// GET /api/v1/admin/subscription-offers/ - Get all offers
	admin.GET("/", func(c *gin.Context) {
		log.Printf("Route: GET /api/v1/admin/subscription-offers/ - Getting all offers")

		offers, err := offersService.GetAllSubscriptionOffers(c.Request.Context())
		if err != nil {
			log.Printf("Route: Error getting all offers: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve subscription offers"})
			return
		}

		// Ensure we always return an array, never null
		if offers == nil {
			offers = []*services.SubscriptionOfferResponse{}
		}

		log.Printf("Route: Successfully retrieved %d offers", len(offers))
		c.JSON(http.StatusOK, offers)
	})

	// GET /api/v1/admin/subscription-offers/:id - Get offer by ID
	admin.GET("/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Printf("Route: Invalid offer ID: %s", idStr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offer ID"})
			return
		}

		log.Printf("Route: GET /api/v1/admin/subscription-offers/%d - Getting offer by ID", id)

		offer, err := offersService.GetSubscriptionOfferByID(c.Request.Context(), id)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("Route: Offer not found: %d", id)
				c.JSON(http.StatusNotFound, gin.H{"error": "Offer not found"})
				return
			}
			log.Printf("Route: Error getting offer by ID: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve subscription offer"})
			return
		}

		log.Printf("Route: Successfully retrieved offer: %d", id)
		c.JSON(http.StatusOK, offer)
	})

	// Test endpoint to check database schema
	admin.GET("/test-schema", func(c *gin.Context) {
		log.Printf("Route: Testing database schema")

		// Test simple query to check table structure
		query := "SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'subscription_offers' ORDER BY ordinal_position"

		rows, err := db.Query(query)
		if err != nil {
			log.Printf("Route: Error querying schema: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query schema", "details": err.Error()})
			return
		}
		defer rows.Close()

		var columns []map[string]string
		for rows.Next() {
			var columnName, dataType string
			if err := rows.Scan(&columnName, &dataType); err != nil {
				log.Printf("Route: Error scanning schema row: %v", err)
				continue
			}
			columns = append(columns, map[string]string{
				"column_name": columnName,
				"data_type":   dataType,
			})
		}

		log.Printf("Route: Schema columns: %+v", columns)
		c.JSON(http.StatusOK, gin.H{"schema": columns})
	})

	// Test endpoint to check history system
	admin.GET("/test-history/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Printf("Route: Invalid offer ID: %s", idStr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offer ID"})
			return
		}

		log.Printf("Route: Testing history for offer ID: %d", id)

		// Get offer history
		history, err := offersService.GetOfferHistory(c.Request.Context(), id)
		if err != nil {
			log.Printf("Route: Error getting offer history: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get offer history", "details": err.Error()})
			return
		}

		log.Printf("Route: Retrieved %d history events for offer %d", len(history), id)
		c.JSON(http.StatusOK, gin.H{
			"offer_id":      id,
			"history_count": len(history),
			"history":       history,
		})
	})

	// POST /api/v1/admin/subscription-offers/ - Create new offer
	admin.POST("/", func(c *gin.Context) {
		log.Printf("Route: POST /api/v1/admin/subscription-offers/ - Creating new offer")

		var req services.CreateSubscriptionOfferRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Route: Invalid request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
			return
		}

		// Get user data from header and create context
		userDataHeader := c.GetHeader("X-User-Data")
		var ctx context.Context = c.Request.Context()

		if userDataHeader != "" {
			log.Printf("Route: Found user data in header: %s", userDataHeader)
			// Create a new context with user data
			ctx = context.WithValue(c.Request.Context(), "frontend_user_data", userDataHeader)
		} else {
			log.Printf("Route: No user data in header, using context from middleware")
		}

		offer, err := offersService.CreateSubscriptionOffer(ctx, &req)
		if err != nil {
			log.Printf("Route: Error creating offer: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription offer", "details": err.Error()})
			return
		}

		log.Printf("Route: Successfully created offer with ID: %d", offer.ID)
		c.JSON(http.StatusCreated, offer)
	})

	// PUT /api/v1/admin/subscription-offers/:id - Update offer
	admin.PUT("/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Printf("Route: Invalid offer ID: %s", idStr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offer ID"})
			return
		}

		log.Printf("Route: PUT /api/v1/admin/subscription-offers/%d - Updating offer", id)

		var req services.UpdateSubscriptionOfferRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Route: Invalid request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
			return
		}

		req.ID = id

		// Get user data from header and create context
		userDataHeader := c.GetHeader("X-User-Data")
		var ctx context.Context = c.Request.Context()

		if userDataHeader != "" {
			log.Printf("Route: Found user data in header: %s", userDataHeader)
			// Create a new context with user data
			ctx = context.WithValue(c.Request.Context(), "frontend_user_data", userDataHeader)
		} else {
			log.Printf("Route: No user data in header, using context from middleware")
		}

		offer, err := offersService.UpdateSubscriptionOffer(ctx, &req)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("Route: Offer not found for update: %d", id)
				c.JSON(http.StatusNotFound, gin.H{"error": "Offer not found"})
				return
			}
			log.Printf("Route: Error updating offer: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription offer", "details": err.Error()})
			return
		}

		log.Printf("Route: Successfully updated offer: %d", id)
		c.JSON(http.StatusOK, offer)
	})

	// DELETE /api/v1/admin/subscription-offers/:id - Delete offer
	admin.DELETE("/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Printf("Route: Invalid offer ID: %s", idStr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offer ID"})
			return
		}

		log.Printf("Route: DELETE /api/v1/admin/subscription-offers/%d - Deleting offer", id)

		// Get user data from header and create context
		userDataHeader := c.GetHeader("X-User-Data")
		var ctx context.Context = c.Request.Context()

		if userDataHeader != "" {
			log.Printf("Route: Found user data in header: %s", userDataHeader)
			// Create a new context with user data
			ctx = context.WithValue(c.Request.Context(), "frontend_user_data", userDataHeader)
		} else {
			log.Printf("Route: No user data in header, using context from middleware")
		}

		err = offersService.DeleteSubscriptionOffer(ctx, id)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("Route: Offer not found for deletion: %d", id)
				c.JSON(http.StatusNotFound, gin.H{"error": "Offer not found"})
				return
			}
			log.Printf("Route: Error deleting offer: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subscription offer", "details": err.Error()})
			return
		}

		log.Printf("Route: Successfully deleted offer: %d", id)
		c.JSON(http.StatusOK, gin.H{"message": "Subscription offer deleted successfully"})
	})

	// POST /api/v1/admin/subscription-offers/:id/toggle - Toggle offer status
	admin.POST("/:id/toggle", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Printf("Route: Invalid offer ID: %s", idStr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offer ID"})
			return
		}

		log.Printf("Route: POST /api/v1/admin/subscription-offers/%d/toggle - Toggling offer status", id)

		// Get user data from header and create context
		userDataHeader := c.GetHeader("X-User-Data")
		var ctx context.Context = c.Request.Context()

		if userDataHeader != "" {
			log.Printf("Route: Found user data in header: %s", userDataHeader)
			// Create a new context with user data
			ctx = context.WithValue(c.Request.Context(), "frontend_user_data", userDataHeader)
		} else {
			log.Printf("Route: No user data in header, using context from middleware")
		}

		// Get current offer to toggle status
		currentOffer, err := offersService.GetSubscriptionOfferByID(ctx, id)
		if err != nil {
			log.Printf("Route: Error getting current offer: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Offer not found"})
			return
		}

		// Toggle the status
		newStatus := !currentOffer.IsActive
		req := services.UpdateSubscriptionOfferRequest{
			ID:       id,
			IsActive: &newStatus,
		}

		offer, err := offersService.UpdateSubscriptionOffer(ctx, &req)
		if err != nil {
			log.Printf("Route: Error toggling offer status: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle offer status", "details": err.Error()})
			return
		}

		log.Printf("Route: Successfully toggled offer status: %d (new status: %t)", id, offer.IsActive)
		c.JSON(http.StatusOK, offer)
	})

	// GET /api/v1/admin/subscription-offers/:id/history - Get offer history
	admin.GET("/:id/history", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Printf("Route: Invalid offer ID: %s", idStr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offer ID"})
			return
		}

		log.Printf("Route: GET /api/v1/admin/subscription-offers/%d/history - Getting offer history", id)

		history, err := offersService.GetOfferHistory(c.Request.Context(), id)
		if err != nil {
			log.Printf("Route: Error getting offer history: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve offer history"})
			return
		}

		log.Printf("Route: Successfully retrieved history for offer: %d (%d events)", id, len(history))
		c.JSON(http.StatusOK, history)
	})

	log.Printf("Subscription offer routes setup complete")
}
