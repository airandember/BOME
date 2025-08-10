package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/stripe/stripe-go/v74"
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
	webhookSecret     string
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

// IsEnabled returns whether Stripe is properly configured
func (s *StripeService) IsEnabled() bool {
	return s.isEnabled
}

// GetConfig returns the current Stripe configuration
func (s *StripeService) GetConfig() *StripeConfig {
	return &StripeConfig{
		SecretKey:         s.secretKey,
		PublishableKey:    s.publishableKey,
		WebhookSecret:     s.webhookSecret,
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

// ValidateWebhookSignature validates a webhook signature
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

	log.Printf("Subscription created: %s", subscription.ID)
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

	log.Printf("Subscription updated: %s", subscription.ID)
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

	log.Printf("Subscription deleted: %s", subscription.ID)
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

	log.Printf("Payment succeeded for invoice: %s", invoice.ID)
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

	log.Printf("Payment failed for invoice: %s", invoice.ID)
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

// GetAccountSummary fetches comprehensive account info for display
func (s *StripeService) GetAccountSummary() (map[string]interface{}, error) {
	if !s.isEnabled {
		return map[string]interface{}{
			"enabled": false,
		}, nil
	}

	// We avoid persisting or returning any secret. Just probe Stripe resources.
	// Try to list various Stripe resources to verify access and show capabilities
	summary := map[string]interface{}{"enabled": true}

	// Fetch products (first 10)
	type productMinimal struct {
		ID, Name, Description string
		Active                bool
		CreatedAt             time.Time
		Metadata              map[string]string
	}
	var products []productMinimal
	{
		params := &stripe.ProductListParams{}
		params.Limit = stripe.Int64(10)
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
		// ignore iter.Err() to keep summary resilient
	}
	summary["products"] = products
	summary["products_count"] = len(products)

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
		params := &stripe.PriceListParams{}
		params.Limit = stripe.Int64(15)
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
	}
	summary["prices"] = prices
	summary["prices_count"] = len(prices)

	// Fetch recent customers (first 10)
	type customerMinimal struct {
		ID, Email, Name string
		CreatedAt       time.Time
		Metadata        map[string]string
	}
	var customers []customerMinimal
	{
		params := &stripe.CustomerListParams{}
		params.Limit = stripe.Int64(10)
		iter := customer.List(params)
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
	}
	summary["customers"] = customers
	summary["customers_count"] = len(customers)

	// Fetch recent subscriptions (first 10)
	type subscriptionMinimal struct {
		ID, Status        string
		CurrentPeriodEnd  time.Time
		CancelAtPeriodEnd bool
		CreatedAt         time.Time
		Metadata          map[string]string
	}
	var subscriptions []subscriptionMinimal
	{
		params := &stripe.SubscriptionListParams{}
		params.Limit = stripe.Int64(10)
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
	}
	summary["subscriptions"] = subscriptions
	summary["subscriptions_count"] = len(subscriptions)

	// Fetch recent payment intents (first 10)
	type paymentIntentMinimal struct {
		ID, Status, Currency string
		Amount               int64
		CreatedAt            time.Time
		Metadata             map[string]string
	}
	var paymentIntents []paymentIntentMinimal
	{
		params := &stripe.PaymentIntentListParams{}
		params.Limit = stripe.Int64(10)
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
	}
	summary["payment_intents"] = paymentIntents
	summary["payment_intents_count"] = len(paymentIntents)

	// Fetch recent invoices (first 10)
	type invoiceMinimal struct {
		ID, Status, Currency string
		Amount               int64
		CreatedAt            time.Time
		Metadata             map[string]string
	}
	var invoices []invoiceMinimal
	{
		params := &stripe.InvoiceListParams{}
		params.Limit = stripe.Int64(10)
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
	}
	summary["invoices"] = invoices
	summary["invoices_count"] = len(invoices)

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
	}

	return summary, nil
}
