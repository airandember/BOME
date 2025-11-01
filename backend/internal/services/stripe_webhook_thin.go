package services

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/subscription"
)

// StripeWebhookThinService handles V2 thin webhook events
// Thin events contain minimal data and require fetching full objects from Stripe API
type StripeWebhookThinService struct {
	webhookServiceV2 *StripeWebhookServiceV2
}

// NewStripeWebhookThinService creates a new thin webhook service
func NewStripeWebhookThinService(webhookServiceV2 *StripeWebhookServiceV2) *StripeWebhookThinService {
	return &StripeWebhookThinService{
		webhookServiceV2: webhookServiceV2,
	}
}

// ThinEvent represents a V2 thin webhook event structure
type ThinEvent struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Created       int64                  `json:"created"`
	Data          map[string]interface{} `json:"data"`
	RelatedObject *ThinRelatedObject     `json:"related_object,omitempty"`
}

// ThinRelatedObject represents a related object reference in thin events
type ThinRelatedObject struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

// ProcessThinEvent processes a V2 thin webhook event
// It parses the thin payload, fetches full objects from Stripe, and delegates to V2 service
func (s *StripeWebhookThinService) ProcessThinEvent(rawPayload []byte) error {
	var thinEvent ThinEvent
	if err := json.Unmarshal(rawPayload, &thinEvent); err != nil {
		return fmt.Errorf("failed to parse thin event: %w", err)
	}

	log.Printf("🔷 [Thin Webhook] Processing: %s (ID: %s)", thinEvent.Type, thinEvent.ID)

	// Route to appropriate handler based on event type
	switch thinEvent.Type {
	// ================================================================
	// BILLING EVENTS (V2 Thin)
	// ================================================================
	case "v2.billing.subscription.created":
		return s.handleSubscriptionCreated(&thinEvent)
	case "v2.billing.subscription.updated":
		return s.handleSubscriptionUpdated(&thinEvent)
	case "v2.billing.subscription.paused":
		return s.handleSubscriptionPaused(&thinEvent)
	case "v2.billing.subscription.resumed":
		return s.handleSubscriptionResumed(&thinEvent)

	// ================================================================
	// CORE EVENTS (V2 Thin)
	// ================================================================
	case "v2.core.event_destination.ping":
		log.Printf("📍 [Thin Webhook] Ping event received - endpoint healthy")
		return nil

	default:
		log.Printf("ℹ️  [Thin Webhook] Unhandled event type: %s", thinEvent.Type)
		return nil // Don't fail on unknown events
	}
}

// ================================================================
// SUBSCRIPTION HANDLERS (Thin → Full Object → V2 Service)
// ================================================================

func (s *StripeWebhookThinService) handleSubscriptionCreated(event *ThinEvent) error {
	subscriptionID := s.extractObjectID(event, "subscription")
	if subscriptionID == "" {
		return fmt.Errorf("no subscription ID in thin event")
	}

	log.Printf("🔷 [Thin Webhook] Fetching full subscription: %s", subscriptionID)

	// Fetch full subscription object from Stripe API
	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch subscription %s: %w", subscriptionID, err)
	}

	// Delegate to V2 webhook service (which expects full objects)
	return s.webhookServiceV2.HandleSubscriptionCreated(sub)
}

func (s *StripeWebhookThinService) handleSubscriptionUpdated(event *ThinEvent) error {
	subscriptionID := s.extractObjectID(event, "subscription")
	if subscriptionID == "" {
		return fmt.Errorf("no subscription ID in thin event")
	}

	log.Printf("🔷 [Thin Webhook] Fetching full subscription: %s", subscriptionID)

	// Fetch full subscription object from Stripe API
	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch subscription %s: %w", subscriptionID, err)
	}

	// Delegate to V2 webhook service
	return s.webhookServiceV2.HandleSubscriptionUpdated(sub)
}

func (s *StripeWebhookThinService) handleSubscriptionPaused(event *ThinEvent) error {
	subscriptionID := s.extractObjectID(event, "subscription")
	if subscriptionID == "" {
		return fmt.Errorf("no subscription ID in thin event")
	}

	log.Printf("🔷 [Thin Webhook] Subscription paused: %s", subscriptionID)

	// Fetch full subscription object
	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch subscription %s: %w", subscriptionID, err)
	}

	// Treat as update (status will be 'paused')
	return s.webhookServiceV2.HandleSubscriptionUpdated(sub)
}

func (s *StripeWebhookThinService) handleSubscriptionResumed(event *ThinEvent) error {
	subscriptionID := s.extractObjectID(event, "subscription")
	if subscriptionID == "" {
		return fmt.Errorf("no subscription ID in thin event")
	}

	log.Printf("🔷 [Thin Webhook] Subscription resumed: %s", subscriptionID)

	// Fetch full subscription object
	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch subscription %s: %w", subscriptionID, err)
	}

	// Treat as update (status will be 'active' again)
	return s.webhookServiceV2.HandleSubscriptionUpdated(sub)
}

// ================================================================
// CUSTOMER HANDLERS (if needed in future)
// ================================================================

func (s *StripeWebhookThinService) handleCustomerCreated(event *ThinEvent) error {
	customerID := s.extractObjectID(event, "customer")
	if customerID == "" {
		return fmt.Errorf("no customer ID in thin event")
	}

	log.Printf("🔷 [Thin Webhook] Fetching full customer: %s", customerID)

	// Fetch full customer object from Stripe API
	cust, err := customer.Get(customerID, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch customer %s: %w", customerID, err)
	}

	// Delegate to V2 webhook service
	return s.webhookServiceV2.HandleCustomerCreated(cust)
}

func (s *StripeWebhookThinService) handleCustomerUpdated(event *ThinEvent) error {
	customerID := s.extractObjectID(event, "customer")
	if customerID == "" {
		return fmt.Errorf("no customer ID in thin event")
	}

	log.Printf("🔷 [Thin Webhook] Fetching full customer: %s", customerID)

	// Fetch full customer object from Stripe API
	cust, err := customer.Get(customerID, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch customer %s: %w", customerID, err)
	}

	// Delegate to V2 webhook service
	return s.webhookServiceV2.HandleCustomerUpdated(cust)
}

// ================================================================
// HELPER FUNCTIONS
// ================================================================

// extractObjectID extracts the object ID from a thin event payload
func (s *StripeWebhookThinService) extractObjectID(event *ThinEvent, objectType string) string {
	// Try data.object.id first (most common)
	if data, ok := event.Data["object"].(map[string]interface{}); ok {
		if id, ok := data["id"].(string); ok {
			return id
		}
	}

	// Try related_object.id (for some events)
	if event.RelatedObject != nil && event.RelatedObject.Type == objectType {
		return event.RelatedObject.ID
	}

	// Try top-level data.id
	if id, ok := event.Data["id"].(string); ok {
		return id
	}

	log.Printf("⚠️  [Thin Webhook] Could not extract %s ID from event", objectType)
	return ""
}

// GetEventMetadata extracts metadata from thin event for logging
func (s *StripeWebhookThinService) GetEventMetadata(event *ThinEvent) map[string]interface{} {
	metadata := map[string]interface{}{
		"event_id":   event.ID,
		"event_type": event.Type,
		"created":    event.Created,
	}

	if event.RelatedObject != nil {
		metadata["related_object_type"] = event.RelatedObject.Type
		metadata["related_object_id"] = event.RelatedObject.ID
	}

	return metadata
}
