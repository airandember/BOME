package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v74"
)

// Webhook activity tracking for live status monitoring
var (
	webhookActivity = struct {
		sync.RWMutex
		lastEvent     time.Time
		eventsToday   int
		totalEvents   int
		successCount  int
		failureCount  int
		eventTypes    map[string]int
		lastResetDate string
	}{
		eventTypes: make(map[string]int),
	}
)

// updateWebhookActivity tracks webhook events for status monitoring
func updateWebhookActivity(eventType string) {
	webhookActivity.Lock()
	defer webhookActivity.Unlock()

	now := time.Now()
	today := now.Format("2006-01-02")

	// Reset daily counters if it's a new day
	if webhookActivity.lastResetDate != today {
		webhookActivity.eventsToday = 0
		webhookActivity.lastResetDate = today
	}

	webhookActivity.lastEvent = now
	webhookActivity.eventsToday++
	webhookActivity.totalEvents++
	webhookActivity.successCount++

	if webhookActivity.eventTypes == nil {
		webhookActivity.eventTypes = make(map[string]int)
	}
	webhookActivity.eventTypes[eventType]++
}

// recordWebhookFailure tracks webhook failures
func recordWebhookFailure() {
	webhookActivity.Lock()
	defer webhookActivity.Unlock()
	webhookActivity.failureCount++
}

// RegisterStripeWebhookRoutes registers webhook endpoints for Stripe events
func RegisterStripeWebhookRoutes(router *gin.RouterGroup, stripeService *services.StripeService, syncService *services.StripeSyncService) {
	webhooks := router.Group("/stripe/webhooks")
	{
		// Main webhook endpoint for all Stripe events
		webhooks.POST("/", func(c *gin.Context) { HandleStripeWebhook(c, stripeService, syncService) })

		// Webhook status endpoint for admin dashboard
		webhooks.GET("/status", func(c *gin.Context) { getWebhookStatus(c, syncService) })
	}
}

// HandleStripeWebhook processes incoming Stripe webhook events dynamically (exported for public endpoint)
func HandleStripeWebhook(c *gin.Context, stripeService *services.StripeService, syncService *services.StripeSyncService) {
	// Read the request body
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("❌ Webhook: Failed to read request body: %v", err)
		recordWebhookFailure()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Get the Stripe signature header
	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		log.Printf("❌ Webhook: Missing Stripe-Signature header")
		recordWebhookFailure()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing Stripe-Signature header"})
		return
	}

	// Parse the raw event to determine version and type
	var rawEvent map[string]interface{}
	if err := json.Unmarshal(payload, &rawEvent); err != nil {
		log.Printf("❌ Webhook: Failed to parse JSON: %v", err)
		recordWebhookFailure()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Extract event type safely
	eventType, typeOk := rawEvent["type"].(string)
	if !typeOk {
		log.Printf("❌ Webhook: Missing or invalid event type")
		recordWebhookFailure()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing event type"})
		return
	}

	log.Printf("📨 Webhook received: %s", eventType)

	// Determine if this is a v1 or v2 event
	isV2Event := strings.HasPrefix(eventType, "v2.")

	if isV2Event {
		// Handle v2 events with raw signature validation
		if err := stripeService.ValidateWebhookSignatureRaw(payload, signature); err != nil {
			log.Printf("❌ Webhook: Invalid v2 signature: %v", err)
			recordWebhookFailure()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
			return
		}

		log.Printf("✅ Webhook: v2 event signature validated successfully")
		updateWebhookActivity(eventType)

		// Process v2 events
		switch eventType {
		case "v2.core.event_destination.ping":
			log.Printf("📍 Webhook: v2 ping event - endpoint is healthy")
			c.JSON(http.StatusOK, gin.H{
				"received":  true,
				"type":      "v2_ping",
				"status":    "healthy",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
			return

		default:
			log.Printf("📋 Webhook: v2 event %s acknowledged but not processed", eventType)
			c.JSON(http.StatusOK, gin.H{
				"received":   true,
				"ignored":    true,
				"reason":     "v2_event_not_implemented",
				"event_type": eventType,
			})
			return
		}
	} else {
		// Handle v1 events with full parsing
		event, err := stripeService.ValidateWebhookSignature(payload, signature)
		if err != nil {
			log.Printf("❌ Webhook: Invalid v1 signature: %v", err)
			recordWebhookFailure()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
			return
		}

		log.Printf("✅ Webhook: v1 event signature validated successfully")
		updateWebhookActivity(event.Type)

		// Process v1 events (your existing logic)
		err = processV1Event(event, syncService)
		if err != nil {
			log.Printf("❌ Webhook: Failed to process v1 event %s: %v", event.Type, err)
			recordWebhookFailure()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process webhook"})
			return
		}

		log.Printf("✅ Webhook: Successfully processed v1 event %s", event.Type)
		c.JSON(http.StatusOK, gin.H{"received": true, "processed": true, "type": "v1_event"})
	}
}

// processV1Event handles all v1 Stripe events
func processV1Event(event *stripe.Event, syncService *services.StripeSyncService) error {
	// Process only the events we care about based on your requirements
	switch event.Type {
	// Customer events - YES
	case "customer.created":
		return handleCustomerCreated(event, syncService)
	case "customer.updated":
		return handleCustomerUpdated(event, syncService)
	case "customer.deleted":
		return handleCustomerDeleted(event, syncService)

	// Subscription events - YES (updated per your request)
	case "customer.subscription.created":
		return handleSubscriptionCreated(event, syncService)
	case "customer.subscription.updated":
		return handleSubscriptionUpdated(event, syncService)
	case "customer.subscription.deleted":
		return handleSubscriptionDeleted(event, syncService)

	// Invoice payment events - YES
	case "invoice.payment_succeeded":
		return handleInvoicePaymentSucceeded(event, syncService)
	case "invoice.payment_failed":
		return handleInvoicePaymentFailed(event, syncService)

	// Product events - YES
	case "product.created":
		return handleProductCreated(event, syncService)
	case "product.updated":
		return handleProductUpdated(event, syncService)

	// Price events - YES
	case "price.created":
		return handlePriceCreated(event, syncService)
	case "price.updated":
		return handlePriceUpdated(event, syncService)

	// Events we explicitly DON'T handle based on your requirements:
	// - payment_intent.* (NO - left out for now)

	default:
		log.Printf("📋 Webhook: Ignoring v1 event type %s (not in our priority list)", event.Type)
		return nil // Not an error - just ignored
	}
}

// Customer webhook handlers
func handleCustomerCreated(event *stripe.Event, syncService *services.StripeSyncService) error {
	var customer stripe.Customer
	if err := json.Unmarshal(event.Data.Raw, &customer); err != nil {
		return err
	}

	log.Printf("👥 Webhook: Customer created - %s (%s)", customer.ID, customer.Email)
	return syncService.UpsertCustomerFromWebhook(&customer)
}

func handleCustomerUpdated(event *stripe.Event, syncService *services.StripeSyncService) error {
	var customer stripe.Customer
	if err := json.Unmarshal(event.Data.Raw, &customer); err != nil {
		return err
	}

	log.Printf("👥 Webhook: Customer updated - %s (%s)", customer.ID, customer.Email)
	return syncService.UpsertCustomerFromWebhook(&customer)
}

func handleCustomerDeleted(event *stripe.Event, syncService *services.StripeSyncService) error {
	var customer stripe.Customer
	if err := json.Unmarshal(event.Data.Raw, &customer); err != nil {
		return err
	}

	log.Printf("👥 Webhook: Customer deleted - %s", customer.ID)
	return syncService.MarkCustomerDeleted(customer.ID)
}

// Invoice payment webhook handlers
func handleInvoicePaymentSucceeded(event *stripe.Event, syncService *services.StripeSyncService) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return err
	}

	log.Printf("🧾 Webhook: Invoice payment succeeded - %s (Amount: %d)", invoice.ID, invoice.AmountPaid)

	// First, sync the invoice data
	if err := syncService.UpsertInvoiceFromWebhook(&invoice); err != nil {
		log.Printf("❌ Failed to sync invoice data: %v", err)
		return err
	}

	// 🎯 CRITICAL: Grant video access when payment succeeds
	if invoice.Customer != nil && invoice.Subscription != nil {
		log.Printf("🎥 Processing video access for customer %s (subscription: %s)",
			invoice.Customer.ID, invoice.Subscription.ID)

		// This will trigger the sync service to update user video access
		// based on their active subscription and product video_approved status
		if err := syncService.ProcessVideoAccessForCustomer(invoice.Customer.ID); err != nil {
			log.Printf("⚠️ Failed to process video access for customer %s: %v",
				invoice.Customer.ID, err)
			// Don't return error - invoice sync succeeded, video access is secondary
		} else {
			log.Printf("✅ Video access processed for customer %s", invoice.Customer.ID)
		}
	}

	return nil
}

func handleInvoicePaymentFailed(event *stripe.Event, syncService *services.StripeSyncService) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return err
	}

	log.Printf("🧾 Webhook: Invoice payment failed - %s", invoice.ID)
	return syncService.UpsertInvoiceFromWebhook(&invoice)
}

// Product webhook handlers
func handleProductCreated(event *stripe.Event, syncService *services.StripeSyncService) error {
	var product stripe.Product
	if err := json.Unmarshal(event.Data.Raw, &product); err != nil {
		return err
	}

	log.Printf("📦 Webhook: Product created - %s (%s)", product.ID, product.Name)
	return syncService.UpsertProductFromWebhook(&product)
}

func handleProductUpdated(event *stripe.Event, syncService *services.StripeSyncService) error {
	var product stripe.Product
	if err := json.Unmarshal(event.Data.Raw, &product); err != nil {
		return err
	}

	log.Printf("📦 Webhook: Product updated - %s (%s)", product.ID, product.Name)
	return syncService.UpsertProductFromWebhook(&product)
}

// Price webhook handlers
func handlePriceCreated(event *stripe.Event, syncService *services.StripeSyncService) error {
	var price stripe.Price
	if err := json.Unmarshal(event.Data.Raw, &price); err != nil {
		return err
	}

	log.Printf("💰 Webhook: Price created - %s (Amount: %d)", price.ID, price.UnitAmount)
	return syncService.UpsertPriceFromWebhook(&price)
}

func handlePriceUpdated(event *stripe.Event, syncService *services.StripeSyncService) error {
	var price stripe.Price
	if err := json.Unmarshal(event.Data.Raw, &price); err != nil {
		return err
	}

	log.Printf("💰 Webhook: Price updated - %s (Amount: %d)", price.ID, price.UnitAmount)
	return syncService.UpsertPriceFromWebhook(&price)
}

// Subscription webhook handlers
func handleSubscriptionCreated(event *stripe.Event, syncService *services.StripeSyncService) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return err
	}

	log.Printf("📋 Webhook: Subscription created - %s (Customer: %s, Status: %s)",
		subscription.ID, subscription.Customer.ID, subscription.Status)
	return syncService.UpsertSubscriptionFromWebhook(&subscription)
}

func handleSubscriptionUpdated(event *stripe.Event, syncService *services.StripeSyncService) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return err
	}

	log.Printf("📋 Webhook: Subscription updated - %s (Customer: %s, Status: %s)",
		subscription.ID, subscription.Customer.ID, subscription.Status)
	return syncService.UpsertSubscriptionFromWebhook(&subscription)
}

func handleSubscriptionDeleted(event *stripe.Event, syncService *services.StripeSyncService) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return err
	}

	log.Printf("📋 Webhook: Subscription deleted - %s (Customer: %s)",
		subscription.ID, subscription.Customer.ID)
	return syncService.MarkSubscriptionDeleted(subscription.ID)
}

// getWebhookStatus provides real-time webhook status information for the admin dashboard
func getWebhookStatus(c *gin.Context, syncService *services.StripeSyncService) {
	webhookActivity.RLock()
	defer webhookActivity.RUnlock()

	// Calculate if webhook is "active" based on recent activity
	now := time.Now()
	isActive := false
	lastEventTime := "Never"

	if !webhookActivity.lastEvent.IsZero() {
		// Consider active if we've received an event in the last 24 hours
		// or if we've received any events today
		timeSinceLastEvent := now.Sub(webhookActivity.lastEvent)
		isActive = timeSinceLastEvent < 24*time.Hour || webhookActivity.eventsToday > 0

		// Format last event time
		if timeSinceLastEvent < time.Minute {
			lastEventTime = "Just now"
		} else if timeSinceLastEvent < time.Hour {
			minutes := int(timeSinceLastEvent.Minutes())
			lastEventTime = fmt.Sprintf("%d minute%s ago", minutes, pluralize(minutes))
		} else if timeSinceLastEvent < 24*time.Hour {
			hours := int(timeSinceLastEvent.Hours())
			lastEventTime = fmt.Sprintf("%d hour%s ago", hours, pluralize(hours))
		} else {
			days := int(timeSinceLastEvent.Hours() / 24)
			lastEventTime = fmt.Sprintf("%d day%s ago", days, pluralize(days))
		}
	}

	// Calculate success rate
	successRate := 100.0
	if webhookActivity.totalEvents > 0 {
		successRate = float64(webhookActivity.successCount) / float64(webhookActivity.totalEvents) * 100
	}

	// Build the public webhook endpoint URL
	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}
	webhookEndpoint := fmt.Sprintf("%s://%s/api/v1/webhooks/stripe", scheme, c.Request.Host)

	status := gin.H{
		"active":      isActive,
		"lastEvent":   lastEventTime,
		"eventsToday": webhookActivity.eventsToday,
		"totalEvents": webhookActivity.totalEvents,
		"successRate": int(successRate),
		"endpoint":    webhookEndpoint,
		"eventTypes":  webhookActivity.eventTypes,
		"failures":    webhookActivity.failureCount,
	}

	log.Printf("📊 Webhook status requested from admin dashboard - Active: %v, Events Today: %d", isActive, webhookActivity.eventsToday)
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"webhook": status,
	})
}

// pluralize helper function
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
