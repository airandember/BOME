package routes

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v74"
)

// RegisterStripeWebhookRoutes registers webhook endpoints for Stripe events
func RegisterStripeWebhookRoutes(router *gin.RouterGroup, stripeService *services.StripeService, syncService *services.StripeSyncService) {
	webhooks := router.Group("/stripe/webhooks")
	{
		// Main webhook endpoint for all Stripe events
		webhooks.POST("/", func(c *gin.Context) { handleStripeWebhook(c, stripeService, syncService) })
	}
}

// handleStripeWebhook processes incoming Stripe webhook events
func handleStripeWebhook(c *gin.Context, stripeService *services.StripeService, syncService *services.StripeSyncService) {
	// Read the request body
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("❌ Webhook: Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Get the Stripe signature header
	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		log.Printf("❌ Webhook: Missing Stripe-Signature header")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing Stripe-Signature header"})
		return
	}

	// Validate the webhook signature
	event, err := stripeService.ValidateWebhookSignature(payload, signature)
	if err != nil {
		log.Printf("❌ Webhook: Invalid signature: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
		return
	}

	log.Printf("📨 Webhook received: %s", event.Type)

	// Process only the events we care about based on your requirements
	switch event.Type {
	// Customer events - YES
	case "customer.created":
		err = handleCustomerCreated(event, syncService)
	case "customer.updated":
		err = handleCustomerUpdated(event, syncService)
	case "customer.deleted":
		err = handleCustomerDeleted(event, syncService)

	// Subscription events - YES (updated per your request)
	case "customer.subscription.created":
		err = handleSubscriptionCreated(event, syncService)
	case "customer.subscription.updated":
		err = handleSubscriptionUpdated(event, syncService)
	case "customer.subscription.deleted":
		err = handleSubscriptionDeleted(event, syncService)

	// Invoice payment events - YES
	case "invoice.payment_succeeded":
		err = handleInvoicePaymentSucceeded(event, syncService)
	case "invoice.payment_failed":
		err = handleInvoicePaymentFailed(event, syncService)

	// Product events - YES
	case "product.created":
		err = handleProductCreated(event, syncService)
	case "product.updated":
		err = handleProductUpdated(event, syncService)

	// Price events - YES
	case "price.created":
		err = handlePriceCreated(event, syncService)
	case "price.updated":
		err = handlePriceUpdated(event, syncService)

	// Events we explicitly DON'T handle based on your requirements:
	// - payment_intent.* (NO - left out for now)

	default:
		log.Printf("📋 Webhook: Ignoring event type %s (not in our priority list)", event.Type)
		c.JSON(http.StatusOK, gin.H{"received": true, "ignored": true})
		return
	}

	if err != nil {
		log.Printf("❌ Webhook: Failed to process %s: %v", event.Type, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process webhook"})
		return
	}

	log.Printf("✅ Webhook: Successfully processed %s", event.Type)
	c.JSON(http.StatusOK, gin.H{"received": true})
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
	return syncService.UpsertInvoiceFromWebhook(&invoice)
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
