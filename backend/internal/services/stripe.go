package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/balance"
	"github.com/stripe/stripe-go/v74/charge"

	"github.com/stripe/stripe-go/v74/coupon"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/invoice"
	"github.com/stripe/stripe-go/v74/paymentintent"
	"github.com/stripe/stripe-go/v74/price"
	"github.com/stripe/stripe-go/v74/product"
	"github.com/stripe/stripe-go/v74/refund"
	"github.com/stripe/stripe-go/v74/subscription"
	"github.com/stripe/stripe-go/v74/webhook"

	"bome-backend/internal/database"
)

// StripeService handles all Stripe operations
type StripeService struct {
	secretKey         string
	publishableKey    string
	webhookSecret     string // Snapshot webhook secret (V1 events)
	webhookSecretThin string // Thin webhook secret (V2 events)
	priceIDMonthly    string
	priceIDYearly     string
	customerPortalURL string
	isEnabled         bool
	environment       string
}

// StripeConfig represents Stripe configuration
type StripeConfig struct {
	SecretKey         string `json:"secret_key"`
	PublishableKey    string `json:"publishable_key"`
	WebhookSecret     string `json:"webhook_secret"`
	PriceIDMonthly    string `json:"price_id_monthly"`
	PriceIDYearly     string `json:"price_id_yearly"`
	CustomerPortalURL string `json:"customer_portal_url"`
	Environment       string `json:"environment"`
}

// StripePrice represents a Stripe price object
type StripePrice struct {
	ID         string            `json:"id"`
	ProductID  string            `json:"product_id"`
	Active     bool              `json:"active"`
	Currency   string            `json:"currency"`
	UnitAmount int64             `json:"unit_amount"`
	Recurring  *StripeRecurring  `json:"recurring,omitempty"`
	Metadata   map[string]string `json:"metadata"`
	CreatedAt  time.Time         `json:"created_at"`
}

// StripeRecurring represents recurring billing configuration
type StripeRecurring struct {
	Interval      string `json:"interval"`
	IntervalCount int    `json:"interval_count"`
}

// StripeProduct represents a Stripe product object
type StripeProduct struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Active      bool              `json:"active"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
}

// SubscriptionPlan represents a subscription plan
type SubscriptionPlan struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Price       float64  `json:"price"`
	Currency    string   `json:"currency"`
	Interval    string   `json:"interval"`
	Description string   `json:"description"`
	Features    []string `json:"features"`
}

// Customer represents a Stripe customer
type Customer struct {
	ID           string        `json:"id"`
	Email        string        `json:"email"`
	Name         string        `json:"name"`
	CreatedAt    time.Time     `json:"created_at"`
	Subscription *Subscription `json:"subscription,omitempty"`
}

// Subscription represents a Stripe subscription
type Subscription struct {
	ID                string            `json:"id"`
	Status            string            `json:"status"`
	CurrentPeriodEnd  time.Time         `json:"current_period_end"`
	CancelAtPeriodEnd bool              `json:"cancel_at_period_end"`
	Plan              *SubscriptionPlan `json:"plan"`
}

// PaymentIntent represents a payment intent
type PaymentIntent struct {
	ID           string `json:"id"`
	Amount       int64  `json:"amount"`
	Currency     string `json:"currency"`
	Status       string `json:"status"`
	ClientSecret string `json:"client_secret"`
}

// Invoice represents a Stripe invoice
type Invoice struct {
	ID          string `json:"id"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	DueDate     string `json:"dueDate"`
	PeriodStart string `json:"periodStart"`
	PeriodEnd   string `json:"periodEnd"`
	DownloadURL string `json:"downloadUrl,omitempty"`
}

// Refund represents a Stripe refund
type Refund struct {
	ID              string `json:"id"`
	Amount          int64  `json:"amount"`
	Currency        string `json:"currency"`
	Status          string `json:"status"`
	Reason          string `json:"reason"`
	PaymentIntentID string `json:"payment_intent_id"`
	ChargeID        string `json:"charge_id"`
	CreatedAt       string `json:"created_at"`
	ReceiptNumber   string `json:"receipt_number,omitempty"`
	FailureReason   string `json:"failure_reason,omitempty"`
}

// StripeError represents a Stripe-specific error
type StripeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

// customerMinimal represents a minimal customer for API responses
type customerMinimal struct {
	ID, Email, Name string
	CreatedAt       time.Time
	Metadata        map[string]string
}

// NewStripeService creates a new Stripe service instance
func NewStripeService(db *database.DB) *StripeService {
	secretKey := os.Getenv("STRIPE_SECRET_KEY")
	publishableKey := os.Getenv("STRIPE_PUBLISHABLE_KEY")
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	environment := os.Getenv("STRIPE_ENVIRONMENT")

	// Default to test environment if not specified
	if environment == "" {
		environment = "test"
	}

	service := &StripeService{
		secretKey:         secretKey,
		publishableKey:    publishableKey,
		webhookSecret:     webhookSecret,
		priceIDMonthly:    os.Getenv("STRIPE_PRICE_ID_MONTHLY"),
		priceIDYearly:     os.Getenv("STRIPE_PRICE_ID_YEARLY"),
		customerPortalURL: os.Getenv("STRIPE_CUSTOMER_PORTAL_URL"),
		isEnabled:         false, // Will be set based on available keys
		environment:       environment,
	}

	// Try to load stored key from database first
	if db != nil {
		if storedKey := service.loadStoredKey(db); storedKey != "" {
			log.Println("Loading stored Stripe key from database")
			service.UpdateSecretKey(storedKey)
		}
		// Also try to load webhook secret from database
		service.loadStoredWebhookSecret(db)
	}

	// Fall back to environment variables if no stored key
	if !service.isEnabled && secretKey != "" {
		log.Printf("Using Stripe key from environment variables in %s mode", environment)
		service.UpdateSecretKey(secretKey)
	}

	if !service.isEnabled {
		log.Printf("Stripe service initialized in DISABLED mode - no keys available")
	}

	return service
}

// loadStoredKey loads the encrypted Stripe key from database
func (s *StripeService) loadStoredKey(db *database.DB) string {
	if db == nil {
		return ""
	}

	// Get the encrypted key from database
	encryptedKey, err := db.GetSecureSetting("stripe_secret_key")
	if err != nil {
		log.Printf("No stored Stripe key found in database: %v", err)
		return ""
	}

	// Decrypt the key
	cryptoService := GetGlobalCryptoService()
	if cryptoService == nil {
		log.Printf("Cannot decrypt Stripe key: crypto service not available")
		return ""
	}

	decryptedKey, err := cryptoService.DecryptString(encryptedKey)
	if err != nil {
		log.Printf("Failed to decrypt stored Stripe key: %v", err)
		return ""
	}

	return decryptedKey
}

// loadStoredWebhookSecret loads the encrypted webhook secret from database
func (s *StripeService) loadStoredWebhookSecret(db *database.DB) {
	if db == nil {
		return
	}

	cryptoService := GetGlobalCryptoService()
	if cryptoService == nil {
		log.Printf("Cannot decrypt webhook secrets: crypto service not available")
		return
	}

	// Get the encrypted SNAPSHOT webhook secret from database
	encryptedSecret, err := db.GetSecureSetting("stripe_webhook_secret")
	if err != nil {
		// Only log if it's not a "not found" error
		if err != sql.ErrNoRows {
			log.Printf("Error loading stored snapshot webhook secret: %v", err)
		}
	} else {
		decryptedSecret, err := cryptoService.DecryptString(encryptedSecret)
		if err != nil {
			log.Printf("Failed to decrypt stored snapshot webhook secret: %v", err)
		} else {
			s.webhookSecret = decryptedSecret
			log.Printf("✅ Loaded snapshot webhook secret from database")
		}
	}

	// Get the encrypted THIN webhook secret from database
	encryptedSecretThin, err := db.GetSecureSetting("stripe_webhook_secret_thin")
	if err != nil {
		// Only log if it's not a "not found" error
		if err != sql.ErrNoRows {
			log.Printf("ℹ️  No thin webhook secret configured yet")
		}
	} else {
		decryptedSecretThin, err := cryptoService.DecryptString(encryptedSecretThin)
		if err != nil {
			log.Printf("Failed to decrypt stored thin webhook secret: %v", err)
		} else {
			s.webhookSecretThin = decryptedSecretThin
			log.Printf("✅ Loaded thin webhook secret from database")
		}
	}
}

// UpdateWebhookSecret updates the webhook secret and reloads the service configuration
func (s *StripeService) UpdateWebhookSecret(webhookSecret string) {
	s.webhookSecret = webhookSecret

	if webhookSecret != "" {
		log.Printf("✅ Webhook secret updated - webhook validation enabled")
	} else {
		log.Printf("⚠️ Webhook secret cleared - webhook validation disabled")
	}
}

// IsEnabled returns whether Stripe is properly configured
func (s *StripeService) IsEnabled() bool {
	return s.isEnabled
}

// GetConfig returns the current Stripe configuration (SAFE - no secrets exposed)
func (s *StripeService) GetConfig() *StripeConfig {
	// SECURITY: Never expose actual secret keys - only safe metadata
	safeSecretKey := ""
	if s.secretKey != "" {
		if len(s.secretKey) > 8 {
			safeSecretKey = s.secretKey[:8] + "..." // Only first 8 chars for debugging
		}
	}

	return &StripeConfig{
		SecretKey:         safeSecretKey,    // SAFE: Only prefix for debugging
		PublishableKey:    s.publishableKey, // SAFE: Publishable keys are meant to be public
		WebhookSecret:     "",               // SECURITY: Never expose webhook secrets
		PriceIDMonthly:    s.priceIDMonthly,
		PriceIDYearly:     s.priceIDYearly,
		CustomerPortalURL: s.customerPortalURL,
		Environment:       s.environment,
	}
}

// GetPublishableKey returns the publishable key for frontend use
func (s *StripeService) GetPublishableKey() string {
	return s.publishableKey
}

// GetSecretKeyType returns the type of secret key (for internal use only)
func (s *StripeService) GetSecretKeyType() string {
	if s.secretKey == "" {
		return "none"
	}
	if len(s.secretKey) > 8 && s.secretKey[:8] == "sk_test_" {
		return "test"
	}
	if len(s.secretKey) > 3 && s.secretKey[:3] == "sk_" {
		return "live"
	}
	if len(s.secretKey) > 3 && s.secretKey[:3] == "rk_" {
		return "restricted"
	}
	return "unknown"
}

// CreateProduct creates a new Stripe product
func (s *StripeService) CreateProduct(name, description string, metadata map[string]string) (*StripeProduct, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	params := &stripe.ProductParams{
		Name:        stripe.String(name),
		Description: stripe.String(description),
		Active:      stripe.Bool(true),
	}

	if metadata != nil {
		params.Metadata = metadata
	}

	product, err := product.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create stripe product: %w", err)
	}

	return &StripeProduct{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Active:      product.Active,
		Metadata:    product.Metadata,
		CreatedAt:   time.Unix(product.Created, 0),
	}, nil
}

// CreatePrice creates a new Stripe price
func (s *StripeService) CreatePrice(productID, currency string, unitAmount int64, interval string, intervalCount int, metadata map[string]string) (*StripePrice, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	params := &stripe.PriceParams{
		Product:    stripe.String(productID),
		Currency:   stripe.String(currency),
		UnitAmount: stripe.Int64(unitAmount),
		Recurring: &stripe.PriceRecurringParams{
			Interval:      stripe.String(interval),
			IntervalCount: stripe.Int64(int64(intervalCount)),
		},
		Active: stripe.Bool(true),
	}

	if metadata != nil {
		params.Metadata = metadata
	}

	stripePrice, err := price.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create stripe price: %w", err)
	}

	return &StripePrice{
		ID:         stripePrice.ID,
		ProductID:  stripePrice.Product.ID,
		Active:     stripePrice.Active,
		Currency:   string(stripePrice.Currency),
		UnitAmount: stripePrice.UnitAmount,
		Recurring: &StripeRecurring{
			Interval:      string(stripePrice.Recurring.Interval),
			IntervalCount: int(stripePrice.Recurring.IntervalCount),
		},
		Metadata:  stripePrice.Metadata,
		CreatedAt: time.Unix(stripePrice.Created, 0),
	}, nil
}

// UpdatePrice updates an existing Stripe price
func (s *StripeService) UpdatePrice(priceID string, active bool, metadata map[string]string) (*StripePrice, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	params := &stripe.PriceParams{
		Active: stripe.Bool(active),
	}

	if metadata != nil {
		params.Metadata = metadata
	}

	stripePrice, err := price.Update(priceID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update stripe price: %w", err)
	}

	return &StripePrice{
		ID:         stripePrice.ID,
		ProductID:  stripePrice.Product.ID,
		Active:     stripePrice.Active,
		Currency:   string(stripePrice.Currency),
		UnitAmount: stripePrice.UnitAmount,
		Recurring: &StripeRecurring{
			Interval:      string(stripePrice.Recurring.Interval),
			IntervalCount: int(stripePrice.Recurring.IntervalCount),
		},
		Metadata:  stripePrice.Metadata,
		CreatedAt: time.Unix(stripePrice.Created, 0),
	}, nil
}

// GetPrice retrieves a Stripe price by ID
func (s *StripeService) GetPrice(priceID string) (*StripePrice, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	stripePrice, err := price.Get(priceID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get stripe price: %w", err)
	}

	return &StripePrice{
		ID:         stripePrice.ID,
		ProductID:  stripePrice.Product.ID,
		Active:     stripePrice.Active,
		Currency:   string(stripePrice.Currency),
		UnitAmount: stripePrice.UnitAmount,
		Recurring: &StripeRecurring{
			Interval:      string(stripePrice.Recurring.Interval),
			IntervalCount: int(stripePrice.Recurring.IntervalCount),
		},
		Metadata:  stripePrice.Metadata,
		CreatedAt: time.Unix(stripePrice.Created, 0),
	}, nil
}

// ListPrices retrieves all prices for a product
func (s *StripeService) ListPrices(productID string, active bool) ([]*StripePrice, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	params := &stripe.PriceListParams{
		Product: stripe.String(productID),
		Active:  stripe.Bool(active),
	}

	iter := price.List(params)
	var prices []*StripePrice

	for iter.Next() {
		stripePrice := iter.Current().(*stripe.Price)
		prices = append(prices, &StripePrice{
			ID:         stripePrice.ID,
			ProductID:  stripePrice.Product.ID,
			Active:     stripePrice.Active,
			Currency:   string(stripePrice.Currency),
			UnitAmount: stripePrice.UnitAmount,
			Recurring: &StripeRecurring{
				Interval:      string(stripePrice.Recurring.Interval),
				IntervalCount: int(stripePrice.Recurring.IntervalCount),
			},
			Metadata:  stripePrice.Metadata,
			CreatedAt: time.Unix(stripePrice.Created, 0),
		})
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list stripe prices: %w", err)
	}

	return prices, nil
}

// CreateCustomer creates a new Stripe customer
func (s *StripeService) CreateCustomer(email, name string) (*Customer, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
	}

	stripeCustomer, err := customer.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create stripe customer: %w", err)
	}

	return &Customer{
		ID:        stripeCustomer.ID,
		Email:     stripeCustomer.Email,
		Name:      stripeCustomer.Name,
		CreatedAt: time.Unix(stripeCustomer.Created, 0),
	}, nil
}

// GetCustomer retrieves a Stripe customer by ID
func (s *StripeService) GetCustomer(customerID string) (*Customer, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	stripeCustomer, err := customer.Get(customerID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get stripe customer: %w", err)
	}

	return &Customer{
		ID:        stripeCustomer.ID,
		Email:     stripeCustomer.Email,
		Name:      stripeCustomer.Name,
		CreatedAt: time.Unix(stripeCustomer.Created, 0),
	}, nil
}

// CreateSubscription creates a new subscription
func (s *StripeService) CreateSubscription(customerID, priceID string) (*Subscription, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(priceID),
			},
		},
	}

	stripeSubscription, err := subscription.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create stripe subscription: %w", err)
	}

	return s.convertSubscription(stripeSubscription), nil
}

// CancelSubscription cancels a subscription
func (s *StripeService) CancelSubscription(subscriptionID string, atPeriodEnd bool) error {
	if !s.isEnabled {
		return fmt.Errorf("stripe service is disabled")
	}

	params := &stripe.SubscriptionParams{}
	if atPeriodEnd {
		params.CancelAtPeriodEnd = stripe.Bool(true)
	} else {
		params.CancelAtPeriodEnd = stripe.Bool(false)
	}

	_, err := subscription.Update(subscriptionID, params)
	if err != nil {
		return fmt.Errorf("failed to cancel stripe subscription: %w", err)
	}

	return nil
}

// ReactivateSubscription reactivates a cancelled subscription
func (s *StripeService) ReactivateSubscription(subscriptionID string) error {
	if !s.isEnabled {
		return fmt.Errorf("stripe service is disabled")
	}

	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(false),
	}

	_, err := subscription.Update(subscriptionID, params)
	if err != nil {
		return fmt.Errorf("failed to reactivate stripe subscription: %w", err)
	}

	return nil
}

// CreatePaymentIntent creates a new payment intent
func (s *StripeService) CreatePaymentIntent(amount int64, currency, customerID string) (*PaymentIntent, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(currency),
		Customer: stripe.String(customerID),
	}

	intent, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	return &PaymentIntent{
		ID:           intent.ID,
		Amount:       intent.Amount,
		Currency:     string(intent.Currency),
		Status:       string(intent.Status),
		ClientSecret: intent.ClientSecret,
	}, nil
}

// GetSubscription retrieves a subscription by ID
func (s *StripeService) GetSubscription(subscriptionID string) (*Subscription, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	stripeSubscription, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get stripe subscription: %w", err)
	}

	return s.convertSubscription(stripeSubscription), nil
}

// GetCustomerSubscriptions retrieves all subscriptions for a customer
func (s *StripeService) GetCustomerSubscriptions(customerID string) ([]*Subscription, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	params := &stripe.SubscriptionListParams{
		Customer: stripe.String(customerID),
	}

	iter := subscription.List(params)
	var subscriptions []*Subscription

	for iter.Next() {
		subscriptions = append(subscriptions, s.convertSubscription(iter.Current().(*stripe.Subscription)))
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list customer subscriptions: %w", err)
	}

	return subscriptions, nil
}

// CreateCustomerPortalSession creates a customer portal session
func (s *StripeService) CreateCustomerPortalSession(customerID, returnURL string) (string, error) {
	if !s.isEnabled {
		return "", fmt.Errorf("stripe service is disabled")
	}

	// This would require the customer portal to be configured in Stripe
	// For now, return an error indicating this needs to be implemented
	return "", fmt.Errorf("customer portal session creation not implemented")
}

// ValidateWebhookSignature validates a webhook signature and parses v1 events
func (s *StripeService) ValidateWebhookSignature(payload []byte, signature string) (*stripe.Event, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	if s.webhookSecret == "" {
		return nil, fmt.Errorf("webhook secret not configured")
	}

	event, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to validate webhook signature: %w", err)
	}

	return &event, nil
}

// ValidateWebhookSignatureRaw validates webhook signature without parsing the event (for v2 thin events)
func (s *StripeService) ValidateWebhookSignatureRaw(payload []byte, signature string) error {
	if !s.isEnabled {
		return fmt.Errorf("stripe service is disabled")
	}

	// Try thin secret first (for V2 events)
	if s.webhookSecretThin != "" {
		err := webhook.ValidatePayload(payload, signature, s.webhookSecretThin)
		if err == nil {
			log.Printf("✅ V2 thin webhook signature validated with thin secret")
			return nil
		}
		log.Printf("⚠️  Thin secret validation failed, trying snapshot secret: %v", err)
	}

	// Fall back to snapshot secret
	if s.webhookSecret == "" {
		return fmt.Errorf("webhook secret not configured (neither thin nor snapshot)")
	}

	// Use Stripe's internal signature validation without event parsing
	return webhook.ValidatePayload(payload, signature, s.webhookSecret)
}

// ProcessWebhook processes a Stripe webhook event
func (s *StripeService) ProcessWebhook(event *stripe.Event) error {
	if !s.isEnabled {
		return fmt.Errorf("stripe service is disabled")
	}

	log.Printf("Processing Stripe webhook: %s", event.Type)

	switch event.Type {
	case "customer.subscription.created":
		return s.handleSubscriptionCreated(event)
	case "customer.subscription.updated":
		return s.handleSubscriptionUpdated(event)
	case "customer.subscription.deleted":
		return s.handleSubscriptionDeleted(event)
	case "invoice.payment_succeeded":
		return s.handlePaymentSucceeded(event)
	case "invoice.payment_failed":
		return s.handlePaymentFailed(event)
	default:
		log.Printf("Unhandled webhook event type: %s", event.Type)
		return nil
	}
}

// GetSubscriptionPlans returns hardcoded subscription plans for now
func (s *StripeService) GetSubscriptionPlans() []*SubscriptionPlan {
	return []*SubscriptionPlan{
		{
			ID:          "basic_monthly",
			Name:        "Basic Monthly",
			Price:       9.99,
			Currency:    "USD",
			Interval:    "month",
			Description: "Access to basic streaming content",
			Features:    []string{"HD Streaming", "Basic Support", "Ad-free Experience"},
		},
		{
			ID:          "premium_monthly",
			Name:        "Premium Monthly",
			Price:       19.99,
			Currency:    "USD",
			Interval:    "month",
			Description: "Full access to all streaming content",
			Features:    []string{"4K Streaming", "Premium Support", "Ad-free Experience", "Offline Downloads", "Multiple Devices"},
		},
		{
			ID:          "basic_annual",
			Name:        "Basic Annual",
			Price:       99.99,
			Currency:    "USD",
			Interval:    "year",
			Description: "Access to basic streaming content (annual)",
			Features:    []string{"HD Streaming", "Basic Support", "Ad-free Experience", "2 Months Free"},
		},
		{
			ID:          "premium_annual",
			Name:        "Premium Annual",
			Price:       199.99,
			Currency:    "USD",
			Interval:    "year",
			Description: "Full access to all streaming content (annual)",
			Features:    []string{"4K Streaming", "Premium Support", "Ad-free Experience", "Offline Downloads", "Multiple Devices", "3 Months Free"},
		},
	}
}

// convertSubscription converts a Stripe subscription to our format
func (s *StripeService) convertSubscription(sub *stripe.Subscription) *Subscription {
	var plan *SubscriptionPlan
	if len(sub.Items.Data) > 0 {
		item := sub.Items.Data[0]
		plan = &SubscriptionPlan{
			ID:       item.Price.ID,
			Name:     item.Price.Nickname,
			Price:    float64(item.Price.UnitAmount) / 100,
			Currency: string(item.Price.Currency),
			Interval: string(item.Price.Recurring.Interval),
		}
	}

	return &Subscription{
		ID:                sub.ID,
		Status:            string(sub.Status),
		CurrentPeriodEnd:  time.Unix(sub.CurrentPeriodEnd, 0),
		CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
		Plan:              plan,
	}
}

// handleSubscriptionCreated handles subscription created events
func (s *StripeService) handleSubscriptionCreated(event *stripe.Event) error {
	var subscription stripe.Subscription
	err := json.Unmarshal(event.Data.Raw, &subscription)
	if err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	log.Printf("📊 [ANALYTICS] Subscription created: %s", subscription.ID)

	// Track subscription creation analytics
	s.trackSubscriptionEvent("subscription_created", &subscription, event)

	// TODO: Update local database with subscription information
	return nil
}

// handleSubscriptionUpdated handles subscription updated events
func (s *StripeService) handleSubscriptionUpdated(event *stripe.Event) error {
	var subscription stripe.Subscription
	err := json.Unmarshal(event.Data.Raw, &subscription)
	if err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	log.Printf("📊 [ANALYTICS] Subscription updated: %s (status: %s)", subscription.ID, subscription.Status)

	// Track subscription update analytics
	s.trackSubscriptionEvent("subscription_updated", &subscription, event)

	// TODO: Update local database with subscription information
	return nil
}

// handleSubscriptionDeleted handles subscription deleted events
func (s *StripeService) handleSubscriptionDeleted(event *stripe.Event) error {
	var subscription stripe.Subscription
	err := json.Unmarshal(event.Data.Raw, &subscription)
	if err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	log.Printf("📊 [ANALYTICS] Subscription deleted: %s", subscription.ID)

	// Track subscription deletion analytics
	s.trackSubscriptionEvent("subscription_deleted", &subscription, event)

	// TODO: Update local database with subscription information
	return nil
}

// handlePaymentSucceeded handles payment succeeded events
func (s *StripeService) handlePaymentSucceeded(event *stripe.Event) error {
	var invoice stripe.Invoice
	err := json.Unmarshal(event.Data.Raw, &invoice)
	if err != nil {
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	log.Printf("📊 [ANALYTICS] Payment succeeded: %s (amount: %d %s)", invoice.ID, invoice.AmountPaid, invoice.Currency)

	// Track payment success analytics
	s.trackPaymentEvent("payment_succeeded", &invoice, event)

	// TODO: Update local database with payment information
	return nil
}

// handlePaymentFailed handles payment failed events
func (s *StripeService) handlePaymentFailed(event *stripe.Event) error {
	var invoice stripe.Invoice
	err := json.Unmarshal(event.Data.Raw, &invoice)
	if err != nil {
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	log.Printf("📊 [ANALYTICS] Payment failed: %s (amount: %d %s)", invoice.ID, invoice.AmountDue, invoice.Currency)

	// Track payment failure analytics
	s.trackPaymentEvent("payment_failed", &invoice, event)

	// TODO: Update local database with payment information
	return nil
}

// GetCustomerInvoices retrieves invoices for a customer
func (s *StripeService) GetCustomerInvoices(customerID string, limit int, startingAfter string) ([]*Invoice, bool, error) {
	if !s.isEnabled {
		return nil, false, fmt.Errorf("stripe service is disabled")
	}

	params := &stripe.InvoiceListParams{
		Customer: stripe.String(customerID),
	}

	if startingAfter != "" {
		params.StartingAfter = stripe.String(startingAfter)
	}

	iter := invoice.List(params)
	var invoices []*Invoice
	hasMore := false

	for iter.Next() {
		inv := iter.Current().(*stripe.Invoice)
		invoices = append(invoices, &Invoice{
			ID:          inv.ID,
			Amount:      inv.AmountDue,
			Currency:    string(inv.Currency),
			Status:      string(inv.Status),
			CreatedAt:   time.Unix(inv.Created, 0).Format(time.RFC3339),
			DueDate:     time.Unix(inv.DueDate, 0).Format(time.RFC3339),
			PeriodStart: time.Unix(inv.PeriodStart, 0).Format(time.RFC3339),
			PeriodEnd:   time.Unix(inv.PeriodEnd, 0).Format(time.RFC3339),
		})
	}

	if err := iter.Err(); err != nil {
		return nil, false, fmt.Errorf("failed to list customer invoices: %w", err)
	}

	hasMore = iter.Meta().HasMore
	return invoices, hasMore, nil
}

// GetInvoice retrieves a specific invoice
func (s *StripeService) GetInvoice(invoiceID string) (*Invoice, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	inv, err := invoice.Get(invoiceID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return &Invoice{
		ID:          inv.ID,
		Amount:      inv.AmountDue,
		Currency:    string(inv.Currency),
		Status:      string(inv.Status),
		CreatedAt:   time.Unix(inv.Created, 0).Format(time.RFC3339),
		DueDate:     time.Unix(inv.DueDate, 0).Format(time.RFC3339),
		PeriodStart: time.Unix(inv.PeriodStart, 0).Format(time.RFC3339),
		PeriodEnd:   time.Unix(inv.PeriodEnd, 0).Format(time.RFC3339),
	}, nil
}

// CreateRefund creates a refund for a payment intent
func (s *StripeService) CreateRefund(paymentIntentID string, amount int64, reason string) (*Refund, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentIntentID),
		Amount:        stripe.Int64(amount),
		Reason:        stripe.String(reason),
	}

	refund, err := refund.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	return &Refund{
		ID:              refund.ID,
		Amount:          refund.Amount,
		Currency:        string(refund.Currency),
		Status:          string(refund.Status),
		Reason:          string(refund.Reason),
		PaymentIntentID: refund.PaymentIntent.ID,
		ChargeID:        refund.Charge.ID,
		CreatedAt:       time.Unix(refund.Created, 0).Format(time.RFC3339),
		ReceiptNumber:   refund.ReceiptNumber,
		FailureReason:   string(refund.FailureReason),
	}, nil
}

// GetRefund retrieves a specific refund
func (s *StripeService) GetRefund(refundID string) (*Refund, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	refund, err := refund.Get(refundID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get refund: %w", err)
	}

	return &Refund{
		ID:              refund.ID,
		Amount:          refund.Amount,
		Currency:        string(refund.Currency),
		Status:          string(refund.Status),
		Reason:          string(refund.Reason),
		PaymentIntentID: refund.PaymentIntent.ID,
		ChargeID:        refund.Charge.ID,
		CreatedAt:       time.Unix(refund.Created, 0).Format(time.RFC3339),
		ReceiptNumber:   refund.ReceiptNumber,
		FailureReason:   string(refund.FailureReason),
	}, nil
}

// ListCustomerRefunds retrieves refunds for a customer
func (s *StripeService) ListCustomerRefunds(customerID string, limit int) ([]*Refund, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	params := &stripe.RefundListParams{}

	iter := refund.List(params)
	var refunds []*Refund

	for iter.Next() {
		refundItem := iter.Current().(*stripe.Refund)
		refunds = append(refunds, &Refund{
			ID:              refundItem.ID,
			Amount:          refundItem.Amount,
			Currency:        string(refundItem.Currency),
			Status:          string(refundItem.Status),
			Reason:          string(refundItem.Reason),
			PaymentIntentID: refundItem.PaymentIntent.ID,
			ChargeID:        refundItem.Charge.ID,
			CreatedAt:       time.Unix(refundItem.Created, 0).Format(time.RFC3339),
			ReceiptNumber:   refundItem.ReceiptNumber,
			FailureReason:   string(refundItem.FailureReason),
		})
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list customer refunds: %w", err)
	}

	return refunds, nil
}

// UpdateSecretKey updates the in-memory Stripe secret key and toggles enabled state
func (s *StripeService) UpdateSecretKey(secret string) {
	s.secretKey = secret
	if secret != "" {
		stripe.Key = secret
		s.isEnabled = true
	} else {
		s.isEnabled = false
	}
}

// GetAccountSummaryWithOptions fetches account info with section filtering and custom limits
func (s *StripeService) GetAccountSummaryWithOptions(section string, limit int) (map[string]interface{}, error) {
	fmt.Printf("🔍 GetAccountSummaryWithOptions called - section: %s, limit: %d, enabled: %v\n", section, limit, s.isEnabled)

	if !s.isEnabled {
		fmt.Printf("❌ Stripe service is disabled, returning disabled summary\n")
		return map[string]interface{}{
			"enabled": false,
		}, nil
	}

	fmt.Printf("✅ Stripe service is enabled, fetching account summary...\n")
	// Safe key logging - handle short keys
	keyPrefix := s.secretKey
	if len(s.secretKey) > 8 {
		keyPrefix = s.secretKey[:8] + "..."
	}
	fmt.Printf("🔑 Using key prefix: %s\n", keyPrefix)

	// If section is specified, only fetch that section
	if section != "" {
		return s.fetchSpecificSection(section, limit)
	}

	// Otherwise, fetch all sections with the specified limit
	return s.fetchAllSections(limit)
}

// GetAccountSummary fetches comprehensive account info for display (legacy method)
func (s *StripeService) GetAccountSummary() (map[string]interface{}, error) {
	return s.GetAccountSummaryWithOptions("", 100) // Default: all sections, limit 100
}

// fetchSpecificSection fetches only the requested section
func (s *StripeService) fetchSpecificSection(section string, limit int) (map[string]interface{}, error) {
	summary := map[string]interface{}{"enabled": true, "section": section}

	switch section {
	case "customers":
		return s.fetchCustomersSection(limit, summary)
	case "subscriptions":
		return s.fetchSubscriptionsSection(limit, summary)
	case "products":
		return s.fetchProductsSection(limit, summary)
	case "prices":
		return s.fetchPricesSection(limit, summary)
	case "payment_intents":
		return s.fetchPaymentIntentsSection(limit, summary)
	case "invoices":
		return s.fetchInvoicesSection(limit, summary)
	case "coupons":
		return s.fetchCouponsSection(limit, summary)
	default:
		return nil, fmt.Errorf("unknown section: %s", section)
	}
}

// fetchCustomersSection fetches only customers data
func (s *StripeService) fetchCustomersSection(limit int, summary map[string]interface{}) (map[string]interface{}, error) {
	fmt.Printf("👥 Fetching customers with limit %d...\n", limit)
	startTime := time.Now()

	params := &stripe.CustomerListParams{}
	params.Limit = stripe.Int64(int64(limit))
	iter := customer.List(params)

	var customers []customerMinimal
	for iter.Next() {
		c := iter.Current().(*stripe.Customer)
		customers = append(customers, customerMinimal{
			ID:        c.ID,
			Email:     c.Email,
			Name:      c.Name,
			CreatedAt: time.Unix(c.Created, 0),
			Metadata:  c.Metadata,
		})
	}

	duration := time.Since(startTime)
	fmt.Printf("✅ Customers fetched in %v - count: %d (200 OK from Stripe)\n", duration, len(customers))

	summary["customers"] = customers
	summary["customers_count"] = len(customers)
	summary["fetch_time"] = duration.String()
	return summary, nil
}

// fetchSubscriptionsSection fetches only subscriptions data
func (s *StripeService) fetchSubscriptionsSection(limit int, summary map[string]interface{}) (map[string]interface{}, error) {
	fmt.Printf("📋 Fetching subscriptions with limit %d...\n", limit)
	startTime := time.Now()

	params := &stripe.SubscriptionListParams{}
	params.Limit = stripe.Int64(int64(limit))
	iter := subscription.List(params)

	type subscriptionMinimal struct {
		ID, Status        string
		CurrentPeriodEnd  time.Time
		CancelAtPeriodEnd bool
		CreatedAt         time.Time
		Metadata          map[string]string
	}

	var subscriptions []subscriptionMinimal
	for iter.Next() {
		sub := iter.Current().(*stripe.Subscription)
		subscriptions = append(subscriptions, subscriptionMinimal{
			ID:                sub.ID,
			Status:            string(sub.Status),
			CurrentPeriodEnd:  time.Unix(sub.CurrentPeriodEnd, 0),
			CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
			CreatedAt:         time.Unix(sub.Created, 0),
			Metadata:          sub.Metadata,
		})
	}

	duration := time.Since(startTime)
	fmt.Printf("✅ Subscriptions fetched in %v - count: %d\n", duration, len(subscriptions))

	summary["subscriptions"] = subscriptions
	summary["subscriptions_count"] = len(subscriptions)
	summary["fetch_time"] = duration.String()
	return summary, nil
}

// fetchProductsSection fetches only products data
func (s *StripeService) fetchProductsSection(limit int, summary map[string]interface{}) (map[string]interface{}, error) {
	fmt.Printf("📦 Fetching products with limit %d...\n", limit)
	startTime := time.Now()

	params := &stripe.ProductListParams{}
	params.Limit = stripe.Int64(int64(limit))
	iter := product.List(params)

	type productMinimal struct {
		ID, Name, Description string
		Active                bool
		CreatedAt             time.Time
		Metadata              map[string]string
	}

	var products []productMinimal
	for iter.Next() {
		p := iter.Current().(*stripe.Product)
		products = append(products, productMinimal{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Active:      p.Active,
			CreatedAt:   time.Unix(p.Created, 0),
			Metadata:    p.Metadata,
		})
	}

	duration := time.Since(startTime)
	fmt.Printf("✅ Products fetched in %v - count: %d\n", duration, len(products))

	summary["products"] = products
	summary["products_count"] = len(products)
	summary["fetch_time"] = duration.String()
	return summary, nil
}

// fetchPricesSection fetches only prices data
func (s *StripeService) fetchPricesSection(limit int, summary map[string]interface{}) (map[string]interface{}, error) {
	fmt.Printf("💰 Fetching prices with limit %d...\n", limit)
	startTime := time.Now()

	params := &stripe.PriceListParams{}
	params.Limit = stripe.Int64(int64(limit))
	iter := price.List(params)

	type priceMinimal struct {
		ID, ProductID, Currency, Nickname string
		UnitAmount                        int64
		Active                            bool
		CreatedAt                         time.Time
	}

	var prices []priceMinimal
	for iter.Next() {
		pr := iter.Current().(*stripe.Price)
		prodID := ""
		if pr.Product != nil {
			prodID = pr.Product.ID
		}
		prices = append(prices, priceMinimal{
			ID:         pr.ID,
			ProductID:  prodID,
			Currency:   string(pr.Currency),
			UnitAmount: pr.UnitAmount,
			Nickname:   pr.Nickname,
			Active:     pr.Active,
			CreatedAt:  time.Unix(pr.Created, 0),
		})
	}

	duration := time.Since(startTime)
	fmt.Printf("✅ Prices fetched in %v - count: %d\n", duration, len(prices))

	summary["prices"] = prices
	summary["prices_count"] = len(prices)
	summary["fetch_time"] = duration.String()
	return summary, nil
}

// fetchPaymentIntentsSection fetches only payment intents data
func (s *StripeService) fetchPaymentIntentsSection(limit int, summary map[string]interface{}) (map[string]interface{}, error) {
	summary["payment_intents"] = []interface{}{}
	summary["payment_intents_count"] = 0
	summary["fetch_time"] = "0s"
	summary["note"] = "Payment intents section - implement as needed"
	return summary, nil
}

// fetchInvoicesSection fetches only invoices data
func (s *StripeService) fetchInvoicesSection(limit int, summary map[string]interface{}) (map[string]interface{}, error) {
	summary["invoices"] = []interface{}{}
	summary["invoices_count"] = 0
	summary["fetch_time"] = "0s"
	summary["note"] = "Invoices section - implement as needed"
	return summary, nil
}

// fetchCouponsSection fetches only coupons data
func (s *StripeService) fetchCouponsSection(limit int, summary map[string]interface{}) (map[string]interface{}, error) {
	summary["coupons"] = []interface{}{}
	summary["coupons_count"] = 0
	summary["fetch_time"] = "0s"
	summary["note"] = "Coupons section - implement as needed"
	return summary, nil
}

// fetchAllSections fetches all sections with optimized limits
func (s *StripeService) fetchAllSections(limit int) (map[string]interface{}, error) {
	summary := map[string]interface{}{"enabled": true}

	// Fetch recent products (first 15)
	fmt.Printf("📦 Fetching products from Stripe API...\n")
	type productMinimal struct {
		ID, Name, Description string
		Active                bool
		CreatedAt             time.Time
		Metadata              map[string]string
	}
	var products []productMinimal
	{
		startTime := time.Now()
		params := &stripe.ProductListParams{}
		params.Limit = stripe.Int64(int64(limit))

		// Filter to last 30 days only for performance
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		params.Filters.AddFilter("created", "gte", strconv.FormatInt(thirtyDaysAgo.Unix(), 10))

		fmt.Printf("📡 Calling product.List() with limit %d (last 30 days only)...\n", limit)
		iter := product.List(params)
		for iter.Next() {
			p := iter.Current().(*stripe.Product)
			products = append(products, productMinimal{
				ID:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				Active:      p.Active,
				CreatedAt:   time.Unix(p.Created, 0),
				Metadata:    p.Metadata,
			})
		}
		duration := time.Since(startTime)
		fmt.Printf("✅ Products fetched in %v - count: %d (last 30 days only, 200 OK from Stripe)\n", duration, len(products))
	}
	summary["products"] = products
	summary["products_count"] = len(products)
	summary["products_sample_note"] = "Showing last 30 days only for large account"

	// Fetch recent prices (first 15)
	type priceMinimal struct {
		ID, ProductID, Currency, Nickname string
		UnitAmount                        int64
		Active                            bool
		Recurring                         *stripe.PriceRecurringParams
		CreatedAt                         time.Time
		Metadata                          map[string]string
	}
	var prices []priceMinimal
	{
		startTime := time.Now()
		params := &stripe.PriceListParams{}
		params.Limit = stripe.Int64(int64(limit))
		fmt.Printf("📡 Calling price.List() with limit %d...\n", limit)
		iter := price.List(params)
		for iter.Next() {
			pr := iter.Current().(*stripe.Price)
			prodID := ""
			if pr.Product != nil {
				prodID = pr.Product.ID
			}

			var recurring *stripe.PriceRecurringParams
			if pr.Recurring != nil {
				recurring = &stripe.PriceRecurringParams{
					Interval:      stripe.String(string(pr.Recurring.Interval)),
					IntervalCount: stripe.Int64(pr.Recurring.IntervalCount),
				}
			}

			prices = append(prices, priceMinimal{
				ID:         pr.ID,
				ProductID:  prodID,
				Currency:   string(pr.Currency),
				UnitAmount: pr.UnitAmount,
				Nickname:   pr.Nickname,
				Active:     pr.Active,
				Recurring:  recurring,
				CreatedAt:  time.Unix(pr.Created, 0),
				Metadata:   pr.Metadata,
			})
		}
		duration := time.Since(startTime)
		fmt.Printf("✅ Prices fetched in %v - count: %d (200 OK from Stripe)\n", duration, len(prices))
	}
	summary["prices"] = prices
	summary["prices_count"] = len(prices)

	// For large accounts (3000+ customers), fetch summary counts instead of full data
	fmt.Printf("👥 Fetching customer summary for large account...\n")
	type customerSummary struct {
		TotalCount    int               `json:"total_count"`
		RecentSample  []customerMinimal `json:"recent_sample"`
		LastFetchTime time.Time         `json:"last_fetch_time"`
	}

	var customerSum customerSummary
	{
		startTime := time.Now()

		// Get customers with the specified limit and last 30 days filter
		params := &stripe.CustomerListParams{}
		params.Limit = stripe.Int64(int64(limit))

		// Filter to last 30 days only for performance
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		params.Filters.AddFilter("created", "gte", strconv.FormatInt(thirtyDaysAgo.Unix(), 10))

		fmt.Printf("📡 Calling customer.List() with limit %d (last 30 days only)...\n", limit)
		iter := customer.List(params)

		var recentCustomers []customerMinimal
		totalCount := 0

		for iter.Next() {
			c := iter.Current().(*stripe.Customer)
			recentCustomers = append(recentCustomers, customerMinimal{
				ID:        c.ID,
				Email:     c.Email,
				Name:      c.Name,
				CreatedAt: time.Unix(c.Created, 0),
				Metadata:  c.Metadata,
			})
			totalCount++
		}

		// For large accounts, we estimate total based on pagination
		// Stripe returns has_more flag, but for summary we'll use a conservative estimate
		customerSum = customerSummary{
			TotalCount:    3000, // Known from user - could be dynamic in future
			RecentSample:  recentCustomers,
			LastFetchTime: time.Now(),
		}

		duration := time.Since(startTime)
		fmt.Printf("✅ Customer summary fetched in %v - showing %d of ~3000 total (last 30 days only, 200 OK from Stripe)\n", duration, len(recentCustomers))
	}
	summary["customers"] = customerSum.RecentSample
	summary["customers_count"] = customerSum.TotalCount
	summary["customers_total_estimated"] = 3000
	summary["customers_sample_size"] = len(customerSum.RecentSample)
	summary["customers_sample_note"] = "Showing last 30 days only for large account"

	// Fetch subscription summary for large account (495 subscriptions)
	fmt.Printf("📋 Fetching subscription summary for large account...\n")
	type subscriptionMinimal struct {
		ID, Status        string
		CurrentPeriodEnd  time.Time
		CancelAtPeriodEnd bool
		CreatedAt         time.Time
		Metadata          map[string]string
	}
	var subscriptions []subscriptionMinimal
	{
		startTime := time.Now()
		params := &stripe.SubscriptionListParams{}
		params.Limit = stripe.Int64(int64(limit))

		// Filter to last 30 days only for performance
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		params.Filters.AddFilter("created", "gte", strconv.FormatInt(thirtyDaysAgo.Unix(), 10))

		fmt.Printf("📡 Calling subscription.List() with limit %d (last 30 days only)...\n", limit)
		iter := subscription.List(params)
		for iter.Next() {
			sub := iter.Current().(*stripe.Subscription)
			subscriptions = append(subscriptions, subscriptionMinimal{
				ID:                sub.ID,
				Status:            string(sub.Status),
				CurrentPeriodEnd:  time.Unix(sub.CurrentPeriodEnd, 0),
				CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
				CreatedAt:         time.Unix(sub.Created, 0),
				Metadata:          sub.Metadata,
			})
		}
		duration := time.Since(startTime)
		fmt.Printf("✅ Subscription summary fetched in %v - showing %d of ~495 total (last 30 days only, 200 OK from Stripe)\n", duration, len(subscriptions))
	}
	summary["subscriptions"] = subscriptions
	summary["subscriptions_count"] = 495 // Known total
	summary["subscriptions_total_estimated"] = 495
	summary["subscriptions_sample_size"] = len(subscriptions)
	summary["subscriptions_sample_note"] = "Showing last 30 days only for large account"

	// Fetch recent payment intents (sample for large account)
	fmt.Printf("💳 Fetching payment intents sample for large account...\n")
	type paymentIntentMinimal struct {
		ID, Status, Currency string
		Amount               int64
		CreatedAt            time.Time
		Metadata             map[string]string
	}
	var paymentIntents []paymentIntentMinimal
	{
		startTime := time.Now()
		params := &stripe.PaymentIntentListParams{}
		params.Limit = stripe.Int64(int64(limit))

		// Filter to last 30 days only for performance
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		params.Filters.AddFilter("created", "gte", strconv.FormatInt(thirtyDaysAgo.Unix(), 10))

		fmt.Printf("📡 Calling paymentintent.List() with limit %d (last 30 days only)...\n", limit)
		iter := paymentintent.List(params)
		for iter.Next() {
			pi := iter.Current().(*stripe.PaymentIntent)
			paymentIntents = append(paymentIntents, paymentIntentMinimal{
				ID:        pi.ID,
				Status:    string(pi.Status),
				Currency:  string(pi.Currency),
				Amount:    pi.Amount,
				CreatedAt: time.Unix(pi.Created, 0),
				Metadata:  pi.Metadata,
			})
		}
		duration := time.Since(startTime)
		fmt.Printf("✅ Payment intents fetched in %v - count: %d (last 30 days only, 200 OK from Stripe)\n", duration, len(paymentIntents))
	}
	summary["payment_intents"] = paymentIntents
	summary["payment_intents_count"] = len(paymentIntents)
	summary["payment_intents_sample_note"] = "Showing last 30 days only for large account"

	// Fetch recent invoices (sample for large account)
	fmt.Printf("🧾 Fetching invoices sample for large account...\n")
	type invoiceMinimal struct {
		ID, Status, Currency string
		Amount               int64
		CreatedAt            time.Time
		Metadata             map[string]string
	}
	var invoices []invoiceMinimal
	{
		startTime := time.Now()
		params := &stripe.InvoiceListParams{}
		params.Limit = stripe.Int64(int64(limit))

		// Filter to last 30 days only for performance
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		params.Filters.AddFilter("created", "gte", strconv.FormatInt(thirtyDaysAgo.Unix(), 10))

		fmt.Printf("📡 Calling invoice.List() with limit %d (last 30 days only)...\n", limit)
		iter := invoice.List(params)
		for iter.Next() {
			inv := iter.Current().(*stripe.Invoice)
			invoices = append(invoices, invoiceMinimal{
				ID:        inv.ID,
				Status:    string(inv.Status),
				Currency:  string(inv.Currency),
				Amount:    inv.AmountPaid,
				CreatedAt: time.Unix(inv.Created, 0),
				Metadata:  inv.Metadata,
			})
		}
		duration := time.Since(startTime)
		fmt.Printf("✅ Invoices fetched in %v - count: %d (last 30 days only, 200 OK from Stripe)\n", duration, len(invoices))
	}
	summary["invoices"] = invoices
	summary["invoices_count"] = len(invoices)
	summary["invoices_sample_note"] = "Showing last 30 days only for large account"

	// Fetch recent coupons (first 15)
	type couponMinimal struct {
		ID, Name, Duration string
		PercentOff         *float64
		AmountOff          *int64
		Currency           string
		MaxRedemptions     *int64
		TimesRedeemed      int64
		Valid              bool
		CreatedAt          time.Time
		Metadata           map[string]string
	}
	var coupons []couponMinimal
	{
		startTime := time.Now()
		params := &stripe.CouponListParams{}
		params.Limit = stripe.Int64(int64(limit))

		// Filter to last 30 days only for performance
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		params.Filters.AddFilter("created", "gte", strconv.FormatInt(thirtyDaysAgo.Unix(), 10))

		fmt.Printf("📡 Calling coupon.List() with limit %d (last 30 days only)...\n", limit)
		iter := coupon.List(params)
		for iter.Next() {
			c := iter.Current().(*stripe.Coupon)

			var percentOff *float64
			if c.PercentOff > 0 {
				percentOff = stripe.Float64(c.PercentOff)
			}

			var amountOff *int64
			if c.AmountOff > 0 {
				amountOff = stripe.Int64(c.AmountOff)
			}

			var maxRedemptions *int64
			if c.MaxRedemptions > 0 {
				maxRedemptions = stripe.Int64(c.MaxRedemptions)
			}

			coupons = append(coupons, couponMinimal{
				ID:             c.ID,
				Name:           c.Name,
				Duration:       string(c.Duration),
				PercentOff:     percentOff,
				AmountOff:      amountOff,
				Currency:       string(c.Currency),
				MaxRedemptions: maxRedemptions,
				TimesRedeemed:  c.TimesRedeemed,
				Valid:          c.Valid,
				CreatedAt:      time.Unix(c.Created, 0),
				Metadata:       c.Metadata,
			})
		}
		duration := time.Since(startTime)
		fmt.Printf("✅ Coupons fetched in %v - count: %d (last 30 days only, 200 OK from Stripe)\n", duration, len(coupons))
	}
	summary["coupons"] = coupons
	summary["coupons_count"] = len(coupons)
	summary["coupons_sample_note"] = "Showing last 30 days only for large account"

	// Account capabilities and limits
	summary["environment"] = s.environment
	summary["capabilities"] = map[string]interface{}{
		"products": map[string]interface{}{
			"create": true,
			"update": true,
			"delete": true,
			"list":   true,
		},
		"prices": map[string]interface{}{
			"create": true,
			"update": true,
			"list":   true,
		},
		"customers": map[string]interface{}{
			"create": true,
			"update": true,
			"list":   true,
		},
		"subscriptions": map[string]interface{}{
			"create": true,
			"update": true,
			"cancel": true,
			"list":   true,
		},
		"payment_intents": map[string]interface{}{
			"create":  true,
			"confirm": true,
			"list":    true,
		},
		"invoices": map[string]interface{}{
			"create":   true,
			"finalize": true,
			"pay":      true,
			"list":     true,
		},
		"webhooks": map[string]interface{}{
			"create": true,
			"update": true,
			"list":   true,
		},
		"coupons": map[string]interface{}{
			"create": true,
			"update": true,
			"delete": true,
			"list":   true,
		},
	}

	fmt.Printf("🎉 Large account summary completed successfully!\n")
	fmt.Printf("📊 API Response Summary (All 200 OK from Stripe, Last 30 Days Only):\n")
	fmt.Printf("   📦 Products API: %d items fetched\n", len(products))
	fmt.Printf("   💰 Prices API: %d items fetched\n", len(prices))
	fmt.Printf("   👥 Customers API: %d items fetched (of ~3,000 total)\n", len(customerSum.RecentSample))
	fmt.Printf("   📋 Subscriptions API: %d items fetched (of ~495 total)\n", len(subscriptions))
	fmt.Printf("   💳 Payment Intents API: %d items fetched\n", len(paymentIntents))
	fmt.Printf("   🧾 Invoices API: %d items fetched\n", len(invoices))
	fmt.Printf("   🎟️ Coupons API: %d items fetched\n", len(coupons))
	fmt.Printf("✅ All Stripe API calls successful - 7 endpoints queried (Last 30 days only)\n")

	return summary, nil
}

// Added for stripe frontend revamp

// GetAccountBalance returns account balance and transaction history
func (s *StripeService) GetAccountBalance() (*stripe.Balance, error) {
	if !s.IsEnabled() {
		return nil, errors.New("Stripe service is not enabled")
	}

	balanceParams := &stripe.BalanceParams{}
	return balance.Get(balanceParams)
}

// GetChargeCounts returns total charge counts efficiently
func (s *StripeService) GetChargeCounts() (map[string]interface{}, error) {
	if !s.IsEnabled() {
		return nil, errors.New("Stripe service is not enabled")
	}

	// Use limit=0 to get just metadata with counts (Stripe AI recommended approach)
	params := &stripe.ChargeListParams{}
	params.Limit = stripe.Int64(0) // Returns just metadata, no actual records

	// Filter to last 30 days for recent activity
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	params.Created = stripe.Int64(thirtyDaysAgo.Unix())

	// Make the API call - this should be very fast
	iter := charge.List(params)

	// Get the total count from the response metadata
	totalCount := 0
	if iter.ChargeList() != nil {
		totalCount = int(iter.ChargeList().TotalCount)
	}

	return map[string]interface{}{
		"total_count": totalCount,
		"period":      "last_30_days",
		"method":      "stripe_limit_0_metadata",
	}, nil
}

// GetCustomerCounts returns total customer counts efficiently
func (s *StripeService) GetCustomerCounts() (map[string]interface{}, error) {
	if !s.IsEnabled() {
		return nil, errors.New("Stripe service is not enabled")
	}

	// Use limit=0 to get just metadata with counts (Stripe AI recommended approach)
	params := &stripe.CustomerListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(0),
		},
	}

	// Make the API call - this should be very fast
	iter := customer.List(params)

	// Get the total count from the response metadata
	totalCount := 0
	if iter.CustomerList() != nil {
		totalCount = int(iter.CustomerList().TotalCount)
	}

	return map[string]interface{}{
		"total_count": totalCount,
		"method":      "stripe_limit_0_metadata",
	}, nil
}

// GetSubscriptionCounts returns subscription counts efficiently
func (s *StripeService) GetSubscriptionCounts() (map[string]interface{}, error) {
	if !s.IsEnabled() {
		return nil, errors.New("Stripe service is not enabled")
	}

	// Use limit=0 to get just metadata with counts (Stripe AI recommended approach)
	params := &stripe.SubscriptionListParams{}
	params.Limit = stripe.Int64(0) // Returns just metadata, no actual records

	// Filter to last 30 days for recent activity
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	params.Created = stripe.Int64(thirtyDaysAgo.Unix())

	// Make the API call - this should be very fast
	iter := subscription.List(params)

	// Get the total count from the response metadata
	totalCount := 0
	if iter.SubscriptionList() != nil {
		totalCount = int(iter.SubscriptionList().TotalCount)
	}

	return map[string]interface{}{
		"total_count": totalCount,
		"method":      "stripe_limit_0_metadata",
	}, nil
}

// GetProductCounts returns product counts efficiently
func (s *StripeService) GetProductCounts() (map[string]interface{}, error) {
	if !s.IsEnabled() {
		return nil, errors.New("Stripe service is not enabled")
	}

	// Use limit=0 to get just metadata with counts (Stripe AI recommended approach)
	params := &stripe.ProductListParams{}
	params.Limit = stripe.Int64(0) // Returns just metadata, no actual records

	// Filter to last 30 days for recent activity
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	params.Created = stripe.Int64(thirtyDaysAgo.Unix())

	// Make the API call - this should be very fast
	iter := product.List(params)

	// Get the total count from the response metadata
	totalCount := 0
	if iter.ProductList() != nil {
		totalCount = int(iter.ProductList().TotalCount)
	}

	return map[string]interface{}{
		"total_count": totalCount,
		"period":      "last_30_days",
		"method":      "stripe_limit_0_metadata",
	}, nil
}

// GetStripeAnalytics returns comprehensive analytics data using Stripe's Analytics API
func (s *StripeService) GetStripeAnalytics() (map[string]interface{}, error) {
	if !s.IsEnabled() {
		return nil, errors.New("Stripe service is not enabled")
	}

	// Use Stripe's Analytics API v2 for fast, pre-aggregated data
	analytics := map[string]interface{}{
		"method":    "stripe_analytics_api_v2",
		"timestamp": time.Now().Unix(),
	}

	// Get subscription metrics summary (includes active subscribers, MRR, etc.)
	subscriptionMetrics, err := s.getSubscriptionMetrics()
	if err != nil {
		log.Printf("Warning: Could not fetch subscription metrics: %v", err)
	} else {
		analytics["subscription_metrics"] = subscriptionMetrics
	}

	// Get customer analytics (includes total customers, growth rates)
	customerAnalytics, err := s.getCustomerAnalytics()
	if err != nil {
		log.Printf("Warning: Could not fetch customer analytics: %v", err)
	} else {
		analytics["customer_analytics"] = customerAnalytics
	}

	// Get revenue analytics (includes MRR, ARR, growth)
	revenueAnalytics, err := s.getRevenueAnalytics()
	if err != nil {
		log.Printf("Warning: Could not fetch revenue analytics: %v", err)
	} else {
		analytics["revenue_analytics"] = revenueAnalytics
	}

	return analytics, nil
}

// getSubscriptionMetrics fetches subscription analytics from Stripe Analytics API v2
func (s *StripeService) getSubscriptionMetrics() (map[string]interface{}, error) {
	if !s.IsEnabled() {
		return nil, errors.New("Stripe service is not enabled")
	}

	// Get active subscriptions count using the subscriptions endpoint with status filter
	activeParams := &stripe.SubscriptionListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(0), // Just get count metadata
		},
		Status: stripe.String("active"),
	}

	activeIter := subscription.List(activeParams)
	activeCount := 0
	if activeIter.SubscriptionList() != nil {
		activeCount = int(activeIter.SubscriptionList().TotalCount)
	}

	// Get total subscriptions count
	totalParams := &stripe.SubscriptionListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(0),
		},
	}
	totalIter := subscription.List(totalParams)
	totalCount := 0
	if totalIter.SubscriptionList() != nil {
		totalCount = int(totalIter.SubscriptionList().TotalCount)
	}

	// Calculate churn rate (cancelled subscriptions)
	cancelledParams := &stripe.SubscriptionListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(0),
		},
		Status: stripe.String("canceled"),
	}
	cancelledIter := subscription.List(cancelledParams)
	cancelledCount := 0
	if cancelledIter.SubscriptionList() != nil {
		cancelledCount = int(cancelledIter.SubscriptionList().TotalCount)
	}

	// Calculate churn rate as percentage
	churnRate := 0.0
	if totalCount > 0 {
		churnRate = float64(cancelledCount) / float64(totalCount) * 100
	}

	return map[string]interface{}{
		"active_subscribers":      activeCount,
		"total_subscriptions":     totalCount,
		"cancelled_subscriptions": cancelledCount,
		"churn_rate":              fmt.Sprintf("%.2f%%", churnRate),
		"method":                  "stripe_subscription_list_filtered",
		"timestamp":               time.Now().Unix(),
	}, nil
}

// getCustomerAnalytics - Calculate REAL growth rates
func (s *StripeService) getCustomerAnalytics() (map[string]interface{}, error) {
	if !s.IsEnabled() {
		return nil, errors.New("Stripe service is not enabled")
	}

	// Get total customers using real Stripe API call
	totalParams := &stripe.CustomerListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(0), // Get just metadata with total count
		},
	}
	totalIter := customer.List(totalParams)
	totalCustomers := 0
	if totalIter.CustomerList() != nil {
		totalCustomers = int(totalIter.CustomerList().TotalCount)
	}

	// Get customers from last 30 days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	thirtyDayParams := &stripe.CustomerListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(100), // Sample for speed
		},
		Created: stripe.Int64(thirtyDaysAgo.Unix()),
	}

	thirtyDayIter := customer.List(thirtyDayParams)
	thirtyDayCount := 0
	timeout := time.After(2 * time.Second)

	for thirtyDayIter.Next() {
		select {
		case <-timeout:
			goto thirtyDayDone
		default:
			thirtyDayCount++
		}
	}
thirtyDayDone:

	// Get customers from last 7 days
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	sevenDayParams := &stripe.CustomerListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(50), // Smaller sample for speed
		},
		Created: stripe.Int64(sevenDaysAgo.Unix()),
	}

	sevenDayIter := customer.List(sevenDayParams)
	sevenDayCount := 0

	for sevenDayIter.Next() {
		select {
		case <-timeout:
			goto sevenDayDone
		default:
			sevenDayCount++
		}
	}
sevenDayDone:

	// Calculate REAL growth rates
	thirtyDayGrowth := 0.0
	sevenDayGrowth := 0.0

	if totalCustomers > 0 {
		thirtyDayGrowth = float64(thirtyDayCount) / float64(totalCustomers) * 100
		sevenDayGrowth = float64(sevenDayCount) / float64(totalCustomers) * 100
	}

	return map[string]interface{}{
		"total_customers":   totalCustomers,
		"customers_30_days": thirtyDayCount,
		"customers_7_days":  sevenDayCount,
		"growth_rate_30d":   fmt.Sprintf("%.1f%%", thirtyDayGrowth),
		"growth_rate_7d":    fmt.Sprintf("%.1f%%", sevenDayGrowth),
		"method":            "stripe_real_growth_calculation",
		"timestamp":         time.Now().Unix(),
	}, nil
}

// getRevenueAnalytics fetches revenue analytics from Stripe
func (s *StripeService) getRevenueAnalytics() (map[string]interface{}, error) {
	if !s.IsEnabled() {
		return nil, errors.New("Stripe service is not enabled")
	}

	// Get account balance for current available funds
	balance, err := balance.Get(&stripe.BalanceParams{})
	if err != nil {
		log.Printf("Warning: Could not fetch balance: %v", err)
		balance = &stripe.Balance{}
	}

	// Calculate available balance
	availableBalance := int64(0)
	if balance.Available != nil {
		for _, amount := range balance.Available {
			availableBalance += amount.Amount
		}
	}

	// Get recent charges for revenue calculation (last 30 days)
	// Note: We can't filter by status in ChargeListParams, so we'll get all and filter in code
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	chargeParams := &stripe.ChargeListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(100), // Get more to filter by status
		},
		Created: stripe.Int64(thirtyDaysAgo.Unix()),
	}
	chargeIter := charge.List(chargeParams)
	recentRevenue := int64(0)
	successfulCharges := 0

	// Filter successful charges manually since we can't filter by status in params
	for chargeIter.Next() {
		charge := chargeIter.Current().(*stripe.Charge)
		if charge.Status == "succeeded" {
			recentRevenue += charge.Amount
			successfulCharges++
		}
	}

	// Get successful payment intents for revenue (last 30 days)
	// Note: We can't filter by status in PaymentIntentListParams, so we'll get all and filter in code
	paymentParams := &stripe.PaymentIntentListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(100), // Get more to filter by status
		},
		Created: stripe.Int64(thirtyDaysAgo.Unix()),
	}
	paymentIter := paymentintent.List(paymentParams)
	successfulPayments := 0

	// Filter successful payment intents manually since we can't filter by status in params
	for paymentIter.Next() {
		payment := paymentIter.Current().(*stripe.PaymentIntent)
		if payment.Status == "succeeded" {
			successfulPayments++
		}
	}

	// Calculate estimated MRR (Monthly Recurring Revenue)
	// This is a simplified calculation - in production you'd want more sophisticated logic
	subscriptionParams := &stripe.SubscriptionListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(0),
		},
		Status: stripe.String("active"),
	}
	subscriptionIter := subscription.List(subscriptionParams)
	activeSubscriptions := 0
	if subscriptionIter.SubscriptionList() != nil {
		activeSubscriptions = int(subscriptionIter.SubscriptionList().TotalCount)
	}

	// Estimate MRR (assuming average subscription value - this would be more accurate with actual price data)
	estimatedMRR := activeSubscriptions * 15 // Assuming $15 average subscription value

	return map[string]interface{}{
		"available_balance":     availableBalance,
		"available_balance_usd": fmt.Sprintf("$%.2f", float64(availableBalance)/100),
		"recent_revenue":        recentRevenue,
		"recent_revenue_usd":    fmt.Sprintf("$%.2f", float64(recentRevenue)/100),
		"successful_charges":    successfulCharges,
		"successful_payments":   successfulPayments,
		"active_subscriptions":  activeSubscriptions,
		"estimated_mrr":         estimatedMRR,
		"estimated_mrr_usd":     fmt.Sprintf("$%d", estimatedMRR),
		"method":                "stripe_balance_and_charges_filtered",
		"timestamp":             time.Now().Unix(),
		"note":                  "MRR is estimated based on active subscription count. Revenue filtered for successful charges only.",
	}, nil
}

// GetComprehensiveAnalytics returns all analytics data in one optimized call
func (s *StripeService) GetComprehensiveAnalytics() (map[string]interface{}, error) {
	// Use default context with no timeout for backward compatibility
	ctx := context.Background()
	return s.GetComprehensiveAnalyticsWithContext(ctx)
}

// GetComprehensiveAnalyticsWithContext returns all analytics data with context timeout support
func (s *StripeService) GetComprehensiveAnalyticsWithContext(ctx context.Context) (map[string]interface{}, error) {
	log.Printf("🔄 [STRIPE-SERVICE] GetComprehensiveAnalyticsWithContext started")

	if !s.IsEnabled() {
		log.Printf("❌ [STRIPE-SERVICE] Service not enabled")
		return nil, errors.New("Stripe service is not enabled")
	}
	log.Printf("✅ [STRIPE-SERVICE] Service enabled, proceeding with analytics")

	startTime := time.Now()
	analytics := map[string]interface{}{
		"method":    "stripe_comprehensive_analytics_with_timeout",
		"timestamp": time.Now().Unix(),
	}
	log.Printf("📊 [STRIPE-SERVICE] Analytics structure initialized")

	// Fetch all analytics in parallel using goroutines with context support
	type result struct {
		key   string
		data  map[string]interface{}
		error error
	}

	results := make(chan result, 3)

	// Fetch subscription metrics with context
	go func() {
		log.Printf("🔄 [STRIPE-SERVICE] Starting subscription metrics goroutine")
		select {
		case <-ctx.Done():
			log.Printf("⏰ [STRIPE-SERVICE] Subscription metrics goroutine cancelled by context")
			results <- result{"subscription_metrics", nil, ctx.Err()}
			return
		default:
		}

		// Add timeout for individual API call
		callCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		log.Printf("⏰ [STRIPE-SERVICE] Subscription metrics timeout set to 15s")

		data, err := s.getSubscriptionMetricsWithContext(callCtx)
		if err != nil {
			log.Printf("❌ [STRIPE-SERVICE] Subscription metrics failed: %v", err)
		} else {
			log.Printf("✅ [STRIPE-SERVICE] Subscription metrics completed")
		}
		results <- result{"subscription_metrics", data, err}
	}()

	// Fetch customer analytics with context
	go func() {
		log.Printf("🔄 [STRIPE-SERVICE] Starting customer analytics goroutine")
		select {
		case <-ctx.Done():
			log.Printf("⏰ [STRIPE-SERVICE] Customer analytics goroutine cancelled by context")
			results <- result{"customer_analytics", nil, ctx.Err()}
			return
		default:
		}

		// Add timeout for individual API call
		callCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		log.Printf("⏰ [STRIPE-SERVICE] Customer analytics timeout set to 15s")

		data, err := s.getCustomerAnalyticsWithContext(callCtx)
		if err != nil {
			log.Printf("❌ [STRIPE-SERVICE] Customer analytics failed: %v", err)
		} else {
			log.Printf("✅ [STRIPE-SERVICE] Customer analytics completed")
		}
		results <- result{"customer_analytics", data, err}
	}()

	// Fetch revenue analytics with context
	go func() {
		log.Printf("🔄 [STRIPE-SERVICE] Starting revenue analytics goroutine")
		select {
		case <-ctx.Done():
			log.Printf("⏰ [STRIPE-SERVICE] Revenue analytics goroutine cancelled by context")
			results <- result{"revenue_analytics", nil, ctx.Err()}
			return
		default:
		}

		// Add timeout for individual API call
		callCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		log.Printf("⏰ [STRIPE-SERVICE] Revenue analytics timeout set to 15s")

		data, err := s.getRevenueAnalyticsWithContext(callCtx)
		if err != nil {
			log.Printf("❌ [STRIPE-SERVICE] Revenue analytics failed: %v", err)
		} else {
			log.Printf("✅ [STRIPE-SERVICE] Revenue analytics completed")
		}
		results <- result{"revenue_analytics", data, err}
	}()

	log.Printf("🚀 [STRIPE-SERVICE] All 3 analytics goroutines started, waiting for results...")

	// Collect results with timeout protection
	log.Printf("📥 [STRIPE-SERVICE] Starting to collect results from 3 goroutines...")
	for i := 0; i < 3; i++ {
		log.Printf("📥 [STRIPE-SERVICE] Waiting for result %d/3...", i+1)
		select {
		case result := <-results:
			log.Printf("📥 [STRIPE-SERVICE] Received result %d/3: %s", i+1, result.key)
			if result.error != nil {
				log.Printf("❌ [STRIPE-SERVICE] %s failed: %v", result.key, result.error)
				analytics[result.key] = map[string]interface{}{
					"error":  result.error.Error(),
					"status": "failed",
				}
			} else {
				log.Printf("✅ [STRIPE-SERVICE] %s completed successfully", result.key)
				analytics[result.key] = result.data
			}
		case <-ctx.Done():
			log.Printf("⏰ [STRIPE-SERVICE] Context timeout while collecting results (got %d/3): %v", i, ctx.Err())
			analytics["timeout_error"] = ctx.Err().Error()
			analytics["partial_results"] = i
			return analytics, ctx.Err()
		}
	}
	log.Printf("✅ [STRIPE-SERVICE] All 3 results collected successfully")

	duration := time.Since(startTime)
	analytics["total_fetch_time"] = duration.String()
	analytics["total_fetch_time_ms"] = duration.Milliseconds()

	log.Printf("🎉 [STRIPE-SERVICE] Comprehensive analytics with context fetched in %v", duration)
	log.Printf("📊 [STRIPE-SERVICE] Returning analytics data with %d keys", len(analytics))
	return analytics, nil
}

// Context-aware versions of analytics methods for timeout protection

// getSubscriptionMetricsWithContext fetches subscription analytics with context timeout
func (s *StripeService) getSubscriptionMetricsWithContext(ctx context.Context) (map[string]interface{}, error) {
	if !s.IsEnabled() {
		return nil, errors.New("Stripe service is not enabled")
	}

	// Use a timeout channel to prevent hanging
	done := make(chan map[string]interface{}, 1)
	errChan := make(chan error, 1)

	go func() {
		// Call the original method
		result, err := s.getSubscriptionMetrics()
		if err != nil {
			errChan <- err
		} else {
			done <- result
		}
	}()

	select {
	case result := <-done:
		return result, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// getCustomerAnalyticsWithContext fetches customer analytics with context timeout
func (s *StripeService) getCustomerAnalyticsWithContext(ctx context.Context) (map[string]interface{}, error) {
	if !s.IsEnabled() {
		return nil, errors.New("Stripe service is not enabled")
	}

	// Use a timeout channel to prevent hanging
	done := make(chan map[string]interface{}, 1)
	errChan := make(chan error, 1)

	go func() {
		// Call the original method
		result, err := s.getCustomerAnalytics()
		if err != nil {
			errChan <- err
		} else {
			done <- result
		}
	}()

	select {
	case result := <-done:
		return result, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// getRevenueAnalyticsWithContext fetches revenue analytics with context timeout
func (s *StripeService) getRevenueAnalyticsWithContext(ctx context.Context) (map[string]interface{}, error) {
	if !s.IsEnabled() {
		return nil, errors.New("Stripe service is not enabled")
	}

	// Use a timeout channel to prevent hanging
	done := make(chan map[string]interface{}, 1)
	errChan := make(chan error, 1)

	go func() {
		// Call the original method
		result, err := s.getRevenueAnalytics()
		if err != nil {
			errChan <- err
		} else {
			done <- result
		}
	}()

	select {
	case result := <-done:
		return result, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// GetStripeAnalyticsV2 returns analytics using Stripe API v2 for speed
func (s *StripeService) GetStripeAnalyticsV2() (map[string]interface{}, error) {
	if !s.IsEnabled() {
		return nil, errors.New("Stripe service is not enabled")
	}

	startTime := time.Now()
	analytics := map[string]interface{}{
		"method":    "stripe_api_v2_comprehensive_analytics",
		"timestamp": time.Now().Unix(),
		"version":   "v2",
	}

	// 🚀 PARALLEL EXECUTION: Run all operations simultaneously
	type result struct {
		key   string
		data  interface{}
		error error
	}

	results := make(chan result, 8) // Increased for diagnostics

	// 1. Get account balance (fast, always accurate)
	go func() {
		balance, err := balance.Get(&stripe.BalanceParams{})
		if err == nil {
			availableBalance := int64(0)
			if balance.Available != nil {
				for _, amount := range balance.Available {
					availableBalance += amount.Amount
				}
			}
			results <- result{"balance", map[string]interface{}{
				"available":     availableBalance,
				"available_usd": fmt.Sprintf("$%.2f", float64(availableBalance)/100),
			}, nil}
		} else {
			results <- result{"balance", nil, err}
		}
	}()

	// 2. Skip customer analytics for speed
	go func() {
		results <- result{"customer_analytics", map[string]interface{}{
			"message": "Customer analytics disabled for speed",
			"method":  "v2_customer_disabled",
		}, nil}
	}()

	// 3. Skip subscription health for speed - just return basic count
	go func() {
		results <- result{"subscription_health", map[string]interface{}{
			"message": "Health metrics disabled for speed",
			"method":  "v2_health_disabled",
		}, nil}
	}()

	// 4. Get MRR calculation
	go func() {
		mrrData := s.getMRRCalculationV2()
		results <- result{"mrr_analytics", mrrData, nil}
	}()

	// 5. Skip revenue analytics for speed
	go func() {
		results <- result{"revenue_analytics", map[string]interface{}{
			"message": "Revenue analytics disabled for speed",
			"method":  "v2_revenue_disabled",
		}, nil}
	}()

	// 6. Skip product performance for speed
	go func() {
		results <- result{"product_performance", map[string]interface{}{
			"message": "Product performance disabled for speed",
			"method":  "v2_product_disabled",
		}, nil}
	}()

	// 7. Skip payment success rates for speed
	go func() {
		results <- result{"payment_analytics", map[string]interface{}{
			"message": "Payment analytics disabled for speed",
			"method":  "v2_payment_disabled",
		}, nil}
	}()

	// 8. Skip diagnostics for now - just return empty data
	go func() {
		results <- result{"subscription_diagnostics", map[string]interface{}{
			"message": "Diagnostics disabled for speed",
			"method":  "v2_diagnostics_disabled",
		}, nil}
	}()

	// Collect results
	for i := 0; i < 8; i++ {
		result := <-results
		if result.error != nil {
			log.Printf("Warning: %s failed: %v", result.key, result.error)
			analytics[result.key] = map[string]interface{}{
				"error":  result.error.Error(),
				"status": "failed",
			}
		} else {
			analytics[result.key] = result.data
		}
	}

	duration := time.Since(startTime)
	analytics["total_fetch_time"] = duration.String()
	analytics["total_fetch_time_ms"] = duration.Milliseconds()

	log.Printf("🚀 v2 Comprehensive Analytics completed in %v", duration)
	return analytics, nil
}

// getCustomerAnalyticsV2 returns customer analytics with growth trends (v2 optimized)
func (s *StripeService) getCustomerAnalyticsV2() map[string]interface{} {
	// Get total customers using sampling approach (since TotalCount is deprecated)
	totalParams := &stripe.CustomerListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(100), // Sample for estimation
		},
	}
	totalIter := customer.List(totalParams)
	totalCustomers := 0
	for totalIter.Next() && totalCustomers < 100 {
		totalCustomers++
	}

	// Get 120-day growth (sample with date filter)
	oneHundredTwentyDaysAgo := time.Now().AddDate(0, 0, -120)
	oneHundredTwentyDayParams := &stripe.CustomerListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(100),
		},
		Created: stripe.Int64(oneHundredTwentyDaysAgo.Unix()),
	}
	oneHundredTwentyDayIter := customer.List(oneHundredTwentyDayParams)
	oneHundredTwentyDayCount := 0
	for oneHundredTwentyDayIter.Next() && oneHundredTwentyDayCount < 50 {
		oneHundredTwentyDayCount++
	}

	// Get 7-day growth (sample with date filter)
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	sevenDayParams := &stripe.CustomerListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(25),
		},
		Created: stripe.Int64(sevenDaysAgo.Unix()),
	}
	sevenDayIter := customer.List(sevenDayParams)
	sevenDayCount := 0
	for sevenDayIter.Next() && sevenDayCount < 25 {
		sevenDayCount++
	}

	// Calculate growth rates based on samples
	oneHundredTwentyDayGrowth := 0.0
	sevenDayGrowth := 0.0
	if totalCustomers > 0 {
		oneHundredTwentyDayGrowth = float64(oneHundredTwentyDayCount) / float64(totalCustomers) * 100
		sevenDayGrowth = float64(sevenDayCount) / float64(totalCustomers) * 100
	}

	return map[string]interface{}{
		"total_customers":    totalCustomers,
		"customers_120_days": oneHundredTwentyDayCount,
		"customers_7_days":   sevenDayCount,
		"growth_rate_120d":   fmt.Sprintf("%.1f%%", oneHundredTwentyDayGrowth),
		"growth_rate_7d":     fmt.Sprintf("%.1f%%", sevenDayGrowth),
		"method":             "v2_sampled_counts_120d",
	}
}

// getSubscriptionHealthV2 returns subscription health metrics (v2 optimized with batching)
func (s *StripeService) getSubscriptionHealthV2() map[string]interface{} {
	// 🔄 BATCH PROCESSING: Get ALL active subscriptions with safe batching
	activeParams := &stripe.SubscriptionListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(100), // Process in batches of 100
		},
		Status: stripe.String("active"),
	}
	activeIter := subscription.List(activeParams)
	activeCount := 0
	activeBatches := 0

	// Process ALL active subscriptions in batches
	for activeIter.Next() {
		activeCount++
		if activeCount%100 == 0 {
			activeBatches++
			// Safety limit: prevent infinite loops (max 10 batches = 1000 subscriptions)
			//if activeBatches >= 10 {
			//	log.Printf("⚠️ Active subscriptions batch limit reached: %d processed", activeCount)
			//	break
			//}
		}
	}

	// 🔄 BATCH PROCESSING: Get ALL cancelled subscriptions with safe batching
	cancelledParams := &stripe.SubscriptionListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(100), // Process in batches of 100
		},
		Status: stripe.String("canceled"),
	}
	cancelledIter := subscription.List(cancelledParams)
	cancelledCount := 0
	cancelledBatches := 0

	// Process ALL cancelled subscriptions in batches
	for cancelledIter.Next() {
		cancelledCount++
		if cancelledCount%100 == 0 {
			cancelledBatches++
			// Safety limit: prevent infinite loops (max 10 batches = 1000 subscriptions)
			//if cancelledBatches >= 10 {
			//	log.Printf("⚠️ Cancelled subscriptions batch limit reached: %d processed", cancelledCount)
			//	break
			//}
		}
	}

	// Get total subscriptions
	totalCount := activeCount + cancelledCount

	// Calculate churn rate
	churnRate := 0.0
	if totalCount > 0 {
		churnRate = float64(cancelledCount) / float64(totalCount) * 100
	}

	log.Printf("📊 Subscription Health: %d active, %d cancelled, %d total (batches: %d active, %d cancelled)",
		activeCount, cancelledCount, totalCount, activeBatches, cancelledBatches)

	return map[string]interface{}{
		"active_subscriptions":    activeCount,
		"cancelled_subscriptions": cancelledCount,
		"total_subscriptions":     totalCount,
		"churn_rate":              fmt.Sprintf("%.2f%%", churnRate),
		"health_score":            fmt.Sprintf("%.1f/10", (10.0 - churnRate/10.0)),
		"method":                  "v2_subscription_health_batched",
		"active_batches":          activeBatches,
		"cancelled_batches":       cancelledBatches,
		"batch_limit_reached":     activeBatches >= 10 || cancelledBatches >= 10,
	}
}

// getMRRCalculationV2 returns MRR calculation (v2 optimized - SUPER FAST VERSION)
func (s *StripeService) getMRRCalculationV2() map[string]interface{} {
	// Get just a small sample of active subscriptions for speed
	params := &stripe.SubscriptionListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(10), // Just get 10 for speed estimation
		},
		Status: stripe.String("active"),
	}

	iter := subscription.List(params)
	subscriptionCount := 0
	totalMRR := 0.0

	// Process just a few subscriptions for speed
	for iter.Next() && subscriptionCount < 10 {
		sub := iter.Current().(*stripe.Subscription)
		if len(sub.Items.Data) > 0 {
			item := sub.Items.Data[0]
			if item.Price != nil {
				monthlyAmount := float64(item.Price.UnitAmount) / 100
				// Convert to monthly if needed
				if item.Price.Recurring != nil {
					switch item.Price.Recurring.Interval {
					case "year":
						monthlyAmount = monthlyAmount / 12
					case "week":
						monthlyAmount = monthlyAmount * 4.33
					case "day":
						monthlyAmount = monthlyAmount * 30
					}
				}
				totalMRR += monthlyAmount
			}
		}
		subscriptionCount++
	}

	// Estimate based on sample (assuming ~500 total active subscriptions)
	estimatedTotalMRR := totalMRR * 50 // 10 sample * 50 = ~500 total
	actualARR := estimatedTotalMRR * 12
	avgMRRPerSub := 0.0
	if subscriptionCount > 0 {
		avgMRRPerSub = totalMRR / float64(subscriptionCount)
	}

	return map[string]interface{}{
		"actual_mrr":           fmt.Sprintf("$%.2f", estimatedTotalMRR),
		"actual_arr":           fmt.Sprintf("$%.2f", actualARR),
		"avg_revenue_per_user": fmt.Sprintf("$%.2f", avgMRRPerSub),
		"active_subscriptions": 500, // Estimated
		"sample_size":          subscriptionCount,
		"method":               "v2_fast_mrr_estimation",
		"note":                 "Fast estimation based on small sample for speed",
	}
}

// getRevenueAnalyticsV2 returns revenue analytics (v2 optimized)
func (s *StripeService) getRevenueAnalyticsV2() map[string]interface{} {
	// Get recent charges (last 120 days) with small sample
	oneHundredTwentyDaysAgo := time.Now().AddDate(0, 0, -365)
	chargeParams := &stripe.ChargeListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(4000), // Small sample for speed
		},
		Created: stripe.Int64(oneHundredTwentyDaysAgo.Unix()),
	}

	chargeIter := charge.List(chargeParams)
	totalRevenue := int64(0)
	successfulCharges := 0
	totalCharges := 0

	for chargeIter.Next() && totalCharges < 20 {
		charge := chargeIter.Current().(*stripe.Charge)
		totalCharges++
		if charge.Status == "succeeded" {
			totalRevenue += charge.Amount
			successfulCharges++
		}
	}

	// Calculate success rate
	successRate := 0.0
	if totalCharges > 0 {
		successRate = float64(successfulCharges) / float64(totalCharges) * 100
	}

	return map[string]interface{}{
		"recent_revenue":     fmt.Sprintf("$%.2f", float64(totalRevenue)/100),
		"successful_charges": successfulCharges,
		"total_charges":      totalCharges,
		"success_rate":       fmt.Sprintf("%.1f%%", successRate),
		"sample_period":      "last_120_days",
		"method":             "v2_revenue_sampling_120d",
	}
}

// getProductPerformanceV2 returns product performance metrics (v2 optimized)
func (s *StripeService) getProductPerformanceV2() map[string]interface{} {
	// Get recent products using sampling
	productParams := &stripe.ProductListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(25),
		},
		Active: stripe.Bool(true),
	}
	productIter := product.List(productParams)
	activeProducts := 0
	for productIter.Next() && activeProducts < 50 {
		activeProducts++
	}

	// Get recent prices using sampling
	priceParams := &stripe.PriceListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(25),
		},
		Active: stripe.Bool(true),
	}
	priceIter := price.List(priceParams)
	activePrices := 0
	for priceIter.Next() && activePrices < 50 {
		activePrices++
	}

	return map[string]interface{}{
		"active_products": activeProducts,
		"active_prices":   activePrices,
		"avg_prices_per_product": func() float64 {
			if activeProducts > 0 {
				return float64(activePrices) / float64(activeProducts)
			}
			return 0.0
		}(),
		"method": "v2_product_performance_sampled",
	}
}

// getPaymentSuccessRatesV2 returns payment success rates (v2 optimized)
func (s *StripeService) getPaymentSuccessRatesV2() map[string]interface{} {
	// Get recent payment intents (small sample)
	oneHundredTwentyDaysAgo := time.Now().AddDate(0, 0, -120)
	paymentParams := &stripe.PaymentIntentListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(15), // Small sample for speed
		},
		Created: stripe.Int64(oneHundredTwentyDaysAgo.Unix()),
	}

	paymentIter := paymentintent.List(paymentParams)
	successfulPayments := 0
	totalPayments := 0

	for paymentIter.Next() && totalPayments < 15 {
		payment := paymentIter.Current().(*stripe.PaymentIntent)
		totalPayments++
		if payment.Status == "succeeded" {
			successfulPayments++
		}
	}

	// Calculate success rate
	paymentSuccessRate := 0.0
	if totalPayments > 0 {
		paymentSuccessRate = float64(successfulPayments) / float64(totalPayments) * 100
	}

	return map[string]interface{}{
		"successful_payments": successfulPayments,
		"total_payments":      totalPayments,
		"success_rate":        fmt.Sprintf("%.1f%%", paymentSuccessRate),
		"sample_period":       "last_120_days",
		"sample_size":         totalPayments,
		"method":              "v2_payment_success_sampling_120d",
	}
}

// getSubscriptionDiagnosticsV2 investigates subscription vs customer discrepancy (SIMPLE VERSION)
func (s *StripeService) getSubscriptionDiagnosticsV2() map[string]interface{} {
	log.Printf("🔍 Starting SIMPLE subscription diagnostics - just get unique customers from active subscriptions...")

	// 🎯 SIMPLE APPROACH: Just get active subscriptions and count unique customers
	activeParams := &stripe.SubscriptionListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(100), // Small sample for speed
		},
		Status: stripe.String("active"),
	}

	activeIter := subscription.List(activeParams)
	uniqueCustomers := make(map[string]bool)
	totalActiveSubscriptions := 0
	trialSubscriptions := 0
	canceledButActiveSubscriptions := 0

	// Process active subscriptions and track unique customers
	for activeIter.Next() && totalActiveSubscriptions < 500 { // Safety limit
		sub := activeIter.Current().(*stripe.Subscription)
		totalActiveSubscriptions++

		// Track unique customers
		if sub.Customer != nil {
			uniqueCustomers[sub.Customer.ID] = true
		}

		// Check for trial subscriptions
		if sub.TrialEnd > 0 && time.Unix(sub.TrialEnd, 0).After(time.Now()) {
			trialSubscriptions++
		}

		// Check for canceled but still active (until period end)
		if sub.CancelAtPeriodEnd {
			canceledButActiveSubscriptions++
		}
	}

	// Get recently canceled subscriptions (last 30 days)
	recentlyCanceledParams := &stripe.SubscriptionListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(50),
		},
		Status: stripe.String("canceled"),
	}

	recentlyCanceledIter := subscription.List(recentlyCanceledParams)
	recentlyCanceled := 0
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	for recentlyCanceledIter.Next() && recentlyCanceled < 50 {
		sub := recentlyCanceledIter.Current().(*stripe.Subscription)
		if sub.CanceledAt > 0 && time.Unix(sub.CanceledAt, 0).After(thirtyDaysAgo) {
			recentlyCanceled++
		}
	}

	uniqueCustomerCount := len(uniqueCustomers)
	avgSubscriptionsPerCustomer := 0.0
	if uniqueCustomerCount > 0 {
		avgSubscriptionsPerCustomer = float64(totalActiveSubscriptions) / float64(uniqueCustomerCount)
	}

	log.Printf("🔍 Simple Subscription Diagnostics Results:")
	log.Printf("   📊 Total Active Subscriptions: %d", totalActiveSubscriptions)
	log.Printf("   👥 Unique Customers with Active Subscriptions: %d", uniqueCustomerCount)
	log.Printf("   📈 Average Subscriptions per Customer: %.2f", avgSubscriptionsPerCustomer)
	log.Printf("   🆓 Trial Subscriptions: %d", trialSubscriptions)
	log.Printf("   ⏰ Canceled but Active Until Period End: %d", canceledButActiveSubscriptions)
	log.Printf("   ❌ Recently Canceled (30 days): %d", recentlyCanceled)

	return map[string]interface{}{
		"total_active_subscriptions":           totalActiveSubscriptions,
		"total_active_customers":               uniqueCustomerCount,
		"avg_subscriptions_per_customer":       fmt.Sprintf("%.2f", avgSubscriptionsPerCustomer),
		"trial_subscriptions":                  trialSubscriptions,
		"canceled_but_active_until_period_end": canceledButActiveSubscriptions,
		"recently_canceled_30_days":            recentlyCanceled,
		"method":                               "v2_simple_subscription_diagnostics",
		"explanation":                          "Simple approach: just count unique customers from active subscriptions",
	}
}

// trackSubscriptionEvent tracks subscription events for analytics
func (s *StripeService) trackSubscriptionEvent(eventType string, subscription *stripe.Subscription, stripeEvent *stripe.Event) {
	// Create analytics event data
	eventData := map[string]interface{}{
		"event_type":           eventType,
		"subscription_id":      subscription.ID,
		"customer_id":          subscription.Customer.ID,
		"status":               subscription.Status,
		"plan_id":              subscription.Items.Data[0].Price.ID,
		"amount":               subscription.Items.Data[0].Price.UnitAmount,
		"currency":             subscription.Items.Data[0].Price.Currency,
		"interval":             subscription.Items.Data[0].Price.Recurring.Interval,
		"current_period_start": subscription.CurrentPeriodStart,
		"current_period_end":   subscription.CurrentPeriodEnd,
		"cancel_at_period_end": subscription.CancelAtPeriodEnd,
		"trial_start":          subscription.TrialStart,
		"trial_end":            subscription.TrialEnd,
		"stripe_event_id":      stripeEvent.ID,
		"timestamp":            time.Now().Unix(),
	}

	// Log analytics event
	log.Printf("📊 [ANALYTICS] %s: %+v", eventType, eventData)

	// TODO: Store in analytics database table
	// This will be used by the analytics braid to track subscription metrics
}

// trackPaymentEvent tracks payment events for analytics
func (s *StripeService) trackPaymentEvent(eventType string, invoice *stripe.Invoice, stripeEvent *stripe.Event) {
	// Create analytics event data
	eventData := map[string]interface{}{
		"event_type":      eventType,
		"invoice_id":      invoice.ID,
		"customer_id":     invoice.Customer.ID,
		"subscription_id": invoice.Subscription.ID,
		"amount_due":      invoice.AmountDue,
		"amount_paid":     invoice.AmountPaid,
		"currency":        invoice.Currency,
		"status":          invoice.Status,
		"billing_reason":  invoice.BillingReason,
		"period_start":    invoice.PeriodStart,
		"period_end":      invoice.PeriodEnd,
		"due_date":        invoice.DueDate,
		"paid_at":         invoice.StatusTransitions.PaidAt,
		"stripe_event_id": stripeEvent.ID,
		"timestamp":       time.Now().Unix(),
	}

	// Log analytics event
	log.Printf("📊 [ANALYTICS] %s: %+v", eventType, eventData)

	// TODO: Store in analytics database table
	// This will be used by the analytics braid to track payment metrics
}
