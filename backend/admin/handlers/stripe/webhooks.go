package stripe

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"bome-backend/infrastructure/database"
	stripeServices "bome-backend/services/stripe"
	subServices "bome-backend/subscription/services"

	"github.com/gin-gonic/gin"
)

// SetupStripeWebhookPublicRoutes sets up PUBLIC webhook routes (no auth required!)
// This MUST be accessible to Stripe servers
func SetupStripeWebhookPublicRoutes(router *gin.RouterGroup, db *database.DB, stripeService *subServices.StripeService, wsHub WebSocketHub) {
	service := stripeServices.NewStripeWebhookService(db, stripeService, wsHub)

	// PUBLIC endpoint - NO AUTHENTICATION
	// Stripe will POST to this endpoint
	router.POST("/stripe", func(c *gin.Context) {
		HandleStripeWebhookEndpoint(c, service)
	})

	log.Println("✅ Public Stripe webhook endpoint registered: POST /webhooks/stripe")
}

// SetupStripeWebhookAdminRoutes sets up ADMIN webhook routes (auth required)
func SetupStripeWebhookAdminRoutes(router *gin.RouterGroup, db *database.DB, stripeService *subServices.StripeService, wsHub WebSocketHub) {
	service := stripeServices.NewStripeWebhookService(db, stripeService, wsHub)

	webhooks := router.Group("/webhooks")
	{
		// GET /admin/streaming/stripe/webhooks/status
		webhooks.GET("/status", func(c *gin.Context) {
			GetWebhookStatusHandler(c, service)
		})

		// GET /admin/streaming/stripe/webhooks/events
		webhooks.GET("/events", func(c *gin.Context) {
			GetWebhookEventsHandler(c, service)
		})

		// POST /admin/streaming/stripe/webhooks/test
		webhooks.POST("/test", func(c *gin.Context) {
			TestWebhookHandler(c, service)
		})
	}
}

// HandleStripeWebhookEndpoint handles the PUBLIC Stripe webhook endpoint
func HandleStripeWebhookEndpoint(c *gin.Context, service *stripeServices.StripeWebhookService) {
	log.Printf("🔔 Webhook received from Stripe")

	// Read the request body
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("❌ Webhook: Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read request body",
		})
		return
	}

	// Get the Stripe signature header
	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		log.Printf("❌ Webhook: Missing Stripe-Signature header")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing Stripe-Signature header",
		})
		return
	}

	// Process the webhook
	err = service.ProcessWebhook(payload, signature)
	if err != nil {
		log.Printf("❌ Webhook processing failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Webhook processing failed",
		})
		return
	}

	log.Printf("✅ Webhook processed successfully")
	c.JSON(http.StatusOK, gin.H{
		"received": true,
	})
}

// GetWebhookStatusHandler handles GET /admin/streaming/stripe/webhooks/status
func GetWebhookStatusHandler(c *gin.Context, service *stripeServices.StripeWebhookService) {
	log.Println("📊 GetWebhookStatusHandler called")

	stats := service.GetStats()

	// Calculate uptime and health
	health := map[string]interface{}{
		"status": "healthy",
		"stats":  stats,
	}

	// Determine health status
	if stats.FailureCount > 0 && stats.SuccessCount > 0 {
		failureRate := float64(stats.FailureCount) / float64(stats.TotalEvents)
		if failureRate > 0.1 { // More than 10% failure rate
			health["status"] = "degraded"
		}
	}

	if stats.TotalEvents == 0 {
		health["status"] = "no_events"
	}

	log.Printf("✅ Webhook stats: %d total events, %d success, %d failures",
		stats.TotalEvents, stats.SuccessCount, stats.FailureCount)

	c.JSON(http.StatusOK, health)
}

// GetWebhookEventsHandler handles GET /admin/streaming/stripe/webhooks/events
func GetWebhookEventsHandler(c *gin.Context, service *stripeServices.StripeWebhookService) {
	log.Println("📊 GetWebhookEventsHandler called")

	// Default to last 100 events
	limit := 100
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := parseIntParam(limitParam); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	events, err := service.GetRecentEvents(limit)
	if err != nil {
		log.Printf("❌ Error getting webhook events: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get webhook events",
			"details": err.Error(),
		})
		return
	}

	// Ensure we return an array, not null
	if events == nil {
		events = []stripeServices.WebhookEvent{}
	}

	log.Printf("✅ Retrieved %d webhook events", len(events))
	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"count":  len(events),
	})
}

// TestWebhookHandler handles POST /admin/streaming/stripe/webhooks/test
func TestWebhookHandler(c *gin.Context, service *stripeServices.StripeWebhookService) {
	log.Println("🧪 TestWebhookHandler called")

	// This is just a test endpoint to verify webhook connectivity
	stats := service.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook endpoint is reachable",
		"stats":   stats,
		"status":  "ok",
	})
}

// Helper function to parse int parameters
func parseIntParam(param string) (int, error) {
	var result int
	_, err := fmt.Sscanf(param, "%d", &result)
	return result, err
}
