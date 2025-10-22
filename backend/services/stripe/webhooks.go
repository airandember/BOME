package stripe

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"bome-backend/infrastructure/database"
	subServices "bome-backend/subscription/services"

	"github.com/stripe/stripe-go/v74"
)

// StripeWebhookService handles Stripe webhook events
type StripeWebhookService struct {
	db            *database.DB
	stripeService *subServices.StripeService
	hub           WebSocketHub
	stats         *WebhookStats
}

// WebhookStats tracks webhook activity
type WebhookStats struct {
	sync.RWMutex
	LastEvent     time.Time         `json:"last_event"`
	EventsToday   int               `json:"events_today"`
	TotalEvents   int               `json:"total_events"`
	SuccessCount  int               `json:"success_count"`
	FailureCount  int               `json:"failure_count"`
	EventTypes    map[string]int    `json:"event_types"`
	LastResetDate string            `json:"last_reset_date"`
}

// WebhookEvent represents a stored webhook event
type WebhookEvent struct {
	ID            int       `json:"id"`
	EventType     string    `json:"event_type"`
	StripeEventID string    `json:"stripe_event_id"`
	Status        string    `json:"status"`        // "success", "failed", "pending"
	ResponseTime  int       `json:"response_time"` // milliseconds
	PayloadSize   int       `json:"payload_size"`  // bytes
	ErrorMessage  string    `json:"error_message,omitempty"`
	RetryCount    int       `json:"retry_count"`
	CreatedAt     time.Time `json:"created_at"`
	ProcessedAt   time.Time `json:"processed_at,omitempty"`
}

// NewStripeWebhookService creates a new webhook service
func NewStripeWebhookService(db *database.DB, stripeService *subServices.StripeService, hub WebSocketHub) *StripeWebhookService {
	return &StripeWebhookService{
		db:            db,
		stripeService: stripeService,
		hub:           hub,
		stats: &WebhookStats{
			EventTypes:    make(map[string]int),
			LastResetDate: time.Now().Format("2006-01-02"),
		},
	}
}

// ProcessWebhook processes an incoming Stripe webhook
func (s *StripeWebhookService) ProcessWebhook(payload []byte, signatureHeader string) error {
	startTime := time.Now()

	// Verify the webhook signature using the stripe service
	eventPtr, err := s.stripeService.ValidateWebhookSignature(payload, signatureHeader)
	if err != nil {
		log.Printf("❌ Webhook signature verification failed: %v", err)
		s.recordFailure()
		return fmt.Errorf("invalid signature: %w", err)
	}

	event := *eventPtr

	// Log the event to database
	eventRecord := &WebhookEvent{
		EventType:     event.Type,
		StripeEventID: event.ID,
		Status:        "processing",
		PayloadSize:   len(payload),
		CreatedAt:     time.Now(),
	}

	// Process the event based on type
	err = s.handleEvent(&event)
	
	responseTime := int(time.Since(startTime).Milliseconds())
	eventRecord.ResponseTime = responseTime
	eventRecord.ProcessedAt = time.Now()

	if err != nil {
		log.Printf("❌ Error processing webhook %s: %v", event.Type, err)
		eventRecord.Status = "failed"
		eventRecord.ErrorMessage = err.Error()
		s.recordFailure()
		s.logEventToDB(eventRecord)
		return err
	}

	eventRecord.Status = "success"
	s.recordSuccess(event.Type)
	s.logEventToDB(eventRecord)

	log.Printf("✅ Successfully processed webhook: %s (%dms)", event.Type, responseTime)
	return nil
}

// handleEvent routes events to appropriate handlers
func (s *StripeWebhookService) handleEvent(event *stripe.Event) error {
	log.Printf("🔔 Webhook received: %s (ID: %s)", event.Type, event.ID)

	switch event.Type {
	// Customer events
	case "customer.created":
		return s.handleCustomerCreated(event)
	case "customer.updated":
		return s.handleCustomerUpdated(event)
	case "customer.deleted":
		return s.handleCustomerDeleted(event)

	// Subscription events
	case "customer.subscription.created":
		return s.handleSubscriptionCreated(event)
	case "customer.subscription.updated":
		return s.handleSubscriptionUpdated(event)
	case "customer.subscription.deleted":
		return s.handleSubscriptionDeleted(event)

	// Payment events
	case "payment_intent.succeeded":
		return s.handlePaymentSucceeded(event)
	case "payment_intent.payment_failed":
		return s.handlePaymentFailed(event)

	// Invoice events
	case "invoice.payment_succeeded":
		return s.handleInvoicePaymentSucceeded(event)
	case "invoice.payment_failed":
		return s.handleInvoicePaymentFailed(event)

	default:
		log.Printf("ℹ️  Unhandled webhook event type: %s", event.Type)
		return nil
	}
}

// Customer event handlers
func (s *StripeWebhookService) handleCustomerCreated(event *stripe.Event) error {
	var customer stripe.Customer
	if err := json.Unmarshal(event.Data.Raw, &customer); err != nil {
		return fmt.Errorf("failed to parse customer: %w", err)
	}

	log.Printf("👤 Customer created: %s (%s)", customer.Email, customer.ID)

	// Store in stripe_customers table
	s.storeCustomer(&customer)

	// Broadcast via WebSocket
	if s.hub != nil {
		s.hub.BroadcastEvent("stripe.webhook.customer_created", map[string]interface{}{
			"customer_id":    customer.ID,
			"email":          customer.Email,
			"name":           customer.Name,
		}, fmt.Sprintf("New customer from Stripe: %s", customer.Email))
	}

	return nil
}

func (s *StripeWebhookService) handleCustomerUpdated(event *stripe.Event) error {
	var customer stripe.Customer
	if err := json.Unmarshal(event.Data.Raw, &customer); err != nil {
		return fmt.Errorf("failed to parse customer: %w", err)
	}

	log.Printf("👤 Customer updated: %s (%s)", customer.Email, customer.ID)

	// Update stripe_customers table
	s.storeCustomer(&customer)

	// Broadcast via WebSocket
	if s.hub != nil {
		s.hub.BroadcastEvent("stripe.webhook.customer_updated", map[string]interface{}{
			"customer_id": customer.ID,
			"email":       customer.Email,
		}, fmt.Sprintf("Customer updated in Stripe: %s", customer.Email))
	}

	return nil
}

func (s *StripeWebhookService) handleCustomerDeleted(event *stripe.Event) error {
	var customer stripe.Customer
	if err := json.Unmarshal(event.Data.Raw, &customer); err != nil {
		return fmt.Errorf("failed to parse customer: %w", err)
	}

	log.Printf("👤 Customer deleted: %s", customer.ID)

	// Mark as deleted in database (soft delete)
	query := `UPDATE stripe_customers SET updated_at = NOW() WHERE stripe_id = $1`
	_, err := s.db.DB.Exec(query, customer.ID)
	if err != nil {
		log.Printf("Warning: Failed to update customer deletion: %v", err)
	}

	// Broadcast via WebSocket
	if s.hub != nil {
		s.hub.BroadcastEvent("stripe.webhook.customer_deleted", map[string]interface{}{
			"customer_id": customer.ID,
		}, "Customer deleted in Stripe")
	}

	return nil
}

// Subscription event handlers
func (s *StripeWebhookService) handleSubscriptionCreated(event *stripe.Event) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	log.Printf("📝 Subscription created: %s (customer: %s)", subscription.ID, subscription.Customer.ID)

	// Store in stripe_subscriptions table
	s.storeSubscription(&subscription)

	// Broadcast via WebSocket
	if s.hub != nil {
		s.hub.BroadcastEvent("stripe.webhook.subscription_created", map[string]interface{}{
			"subscription_id": subscription.ID,
			"customer_id":     subscription.Customer.ID,
			"status":          subscription.Status,
		}, "New subscription created in Stripe!")
	}

	return nil
}

func (s *StripeWebhookService) handleSubscriptionUpdated(event *stripe.Event) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	log.Printf("📝 Subscription updated: %s (status: %s)", subscription.ID, subscription.Status)

	// Update stripe_subscriptions table
	s.storeSubscription(&subscription)

	// Broadcast via WebSocket
	if s.hub != nil {
		s.hub.BroadcastEvent("stripe.webhook.subscription_updated", map[string]interface{}{
			"subscription_id": subscription.ID,
			"status":          subscription.Status,
		}, fmt.Sprintf("Subscription %s: %s", subscription.ID, subscription.Status))
	}

	return nil
}

func (s *StripeWebhookService) handleSubscriptionDeleted(event *stripe.Event) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	log.Printf("📝 Subscription deleted: %s", subscription.ID)

	// Update status in database
	query := `UPDATE stripe_subscriptions SET status = 'canceled', updated_at = NOW() WHERE stripe_id = $1`
	_, err := s.db.DB.Exec(query, subscription.ID)
	if err != nil {
		log.Printf("Warning: Failed to update subscription: %v", err)
	}

	// Broadcast via WebSocket
	if s.hub != nil {
		s.hub.BroadcastEvent("stripe.webhook.subscription_canceled", map[string]interface{}{
			"subscription_id": subscription.ID,
		}, "Subscription canceled in Stripe")
	}

	return nil
}

// Payment event handlers
func (s *StripeWebhookService) handlePaymentSucceeded(event *stripe.Event) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		return fmt.Errorf("failed to parse payment intent: %w", err)
	}

	log.Printf("💳 Payment succeeded: %s (amount: %d %s)", paymentIntent.ID, paymentIntent.Amount, paymentIntent.Currency)

	// Broadcast via WebSocket
	if s.hub != nil {
		s.hub.BroadcastEvent("stripe.webhook.payment_received", map[string]interface{}{
			"payment_id": paymentIntent.ID,
			"amount":     float64(paymentIntent.Amount) / 100,
			"currency":   paymentIntent.Currency,
		}, fmt.Sprintf("Payment received: $%.2f", float64(paymentIntent.Amount)/100))
	}

	return nil
}

func (s *StripeWebhookService) handlePaymentFailed(event *stripe.Event) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		return fmt.Errorf("failed to parse payment intent: %w", err)
	}

	log.Printf("❌ Payment failed: %s", paymentIntent.ID)

	// Broadcast via WebSocket
	if s.hub != nil {
		s.hub.BroadcastEvent("stripe.webhook.payment_failed", map[string]interface{}{
			"payment_id": paymentIntent.ID,
			"amount":     float64(paymentIntent.Amount) / 100,
			"currency":   paymentIntent.Currency,
		}, "Payment failed")
	}

	return nil
}

// Invoice event handlers
func (s *StripeWebhookService) handleInvoicePaymentSucceeded(event *stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("failed to parse invoice: %w", err)
	}

	log.Printf("📄 Invoice payment succeeded: %s (amount: %d)", invoice.ID, invoice.AmountPaid)

	// Broadcast via WebSocket
	if s.hub != nil {
		s.hub.BroadcastEvent("stripe.webhook.payment_received", map[string]interface{}{
			"invoice_id": invoice.ID,
			"amount":     float64(invoice.AmountPaid) / 100,
			"currency":   invoice.Currency,
		}, fmt.Sprintf("Invoice paid: $%.2f", float64(invoice.AmountPaid)/100))
	}

	return nil
}

func (s *StripeWebhookService) handleInvoicePaymentFailed(event *stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("failed to parse invoice: %w", err)
	}

	log.Printf("❌ Invoice payment failed: %s", invoice.ID)

	// Broadcast via WebSocket
	if s.hub != nil {
		s.hub.BroadcastEvent("stripe.webhook.payment_failed", map[string]interface{}{
			"invoice_id": invoice.ID,
		}, "Invoice payment failed")
	}

	return nil
}

// Helper functions

func (s *StripeWebhookService) storeCustomer(customer *stripe.Customer) {
	metadata := "{}"
	if customer.Metadata != nil && len(customer.Metadata) > 0 {
		if metadataBytes, err := json.Marshal(customer.Metadata); err == nil {
			metadata = string(metadataBytes)
		}
	}

	query := `
		INSERT INTO stripe_customers (stripe_id, email, name, created_at, updated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (stripe_id) 
		DO UPDATE SET email = $2, name = $3, updated_at = $5, metadata = $6
	`

	_, err := s.db.DB.Exec(query,
		customer.ID,
		customer.Email,
		customer.Name,
		time.Unix(customer.Created, 0),
		time.Now(),
		metadata,
	)

	if err != nil {
		log.Printf("Warning: Failed to store customer: %v", err)
	}
}

func (s *StripeWebhookService) storeSubscription(subscription *stripe.Subscription) {
	// Get customer ID from stripe_customers table
	var customerID int
	err := s.db.DB.QueryRow("SELECT id FROM stripe_customers WHERE stripe_id = $1", subscription.Customer.ID).Scan(&customerID)
	if err != nil {
		log.Printf("Warning: Customer not found for subscription: %v", err)
		return
	}

	query := `
		INSERT INTO stripe_subscriptions (
			stripe_id, customer_id, status, current_period_start, current_period_end, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (stripe_id)
		DO UPDATE SET status = $3, current_period_start = $4, current_period_end = $5
	`

	_, err = s.db.DB.Exec(query,
		subscription.ID,
		customerID,
		subscription.Status,
		time.Unix(subscription.CurrentPeriodStart, 0),
		time.Unix(subscription.CurrentPeriodEnd, 0),
		time.Unix(subscription.Created, 0),
	)

	if err != nil {
		log.Printf("Warning: Failed to store subscription: %v", err)
	}
}

func (s *StripeWebhookService) logEventToDB(event *WebhookEvent) {
	query := `
		INSERT INTO webhook_events (
			event_type, stripe_event_id, status, response_time, payload_size,
			error_message, retry_count, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := s.db.DB.Exec(query,
		event.EventType,
		event.StripeEventID,
		event.Status,
		event.ResponseTime,
		event.PayloadSize,
		event.ErrorMessage,
		event.RetryCount,
		event.CreatedAt,
	)

	if err != nil {
		log.Printf("Warning: Failed to log webhook event: %v", err)
	}
}

func (s *StripeWebhookService) recordSuccess(eventType string) {
	s.stats.Lock()
	defer s.stats.Unlock()

	now := time.Now()
	today := now.Format("2006-01-02")

	// Reset daily counters if it's a new day
	if s.stats.LastResetDate != today {
		s.stats.EventsToday = 0
		s.stats.LastResetDate = today
	}

	s.stats.LastEvent = now
	s.stats.EventsToday++
	s.stats.TotalEvents++
	s.stats.SuccessCount++
	s.stats.EventTypes[eventType]++
}

func (s *StripeWebhookService) recordFailure() {
	s.stats.Lock()
	defer s.stats.Unlock()
	s.stats.FailureCount++
}

// GetStats returns current webhook statistics
func (s *StripeWebhookService) GetStats() *WebhookStats {
	s.stats.RLock()
	defer s.stats.RUnlock()

	// Create a copy to avoid race conditions
	statsCopy := &WebhookStats{
		LastEvent:     s.stats.LastEvent,
		EventsToday:   s.stats.EventsToday,
		TotalEvents:   s.stats.TotalEvents,
		SuccessCount:  s.stats.SuccessCount,
		FailureCount:  s.stats.FailureCount,
		EventTypes:    make(map[string]int),
		LastResetDate: s.stats.LastResetDate,
	}

	for k, v := range s.stats.EventTypes {
		statsCopy.EventTypes[k] = v
	}

	return statsCopy
}

// GetRecentEvents returns recent webhook events from database
func (s *StripeWebhookService) GetRecentEvents(limit int) ([]WebhookEvent, error) {
	query := `
		SELECT id, event_type, stripe_event_id, status, response_time, payload_size,
		       error_message, retry_count, created_at
		FROM webhook_events
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := s.db.DB.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query webhook events: %w", err)
	}
	defer rows.Close()

	var events []WebhookEvent
	for rows.Next() {
		var event WebhookEvent
		err := rows.Scan(
			&event.ID, &event.EventType, &event.StripeEventID, &event.Status,
			&event.ResponseTime, &event.PayloadSize, &event.ErrorMessage,
			&event.RetryCount, &event.CreatedAt,
		)
		if err != nil {
			log.Printf("Error scanning webhook event: %v", err)
			continue
		}
		events = append(events, event)
	}

	return events, nil
}

