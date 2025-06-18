package services

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/paymentintent"
	"github.com/stripe/stripe-go/v74/subscription"
	"github.com/stripe/stripe-go/v74/webhook"
)

// StripeService handles all Stripe operations
type StripeService struct {
	secretKey         string
	publishableKey    string
	webhookSecret     string
	priceIDMonthly    string
	priceIDYearly     string
	customerPortalURL string
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

// NewStripeService creates a new Stripe service instance
func NewStripeService() *StripeService {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	return &StripeService{
		secretKey:         os.Getenv("STRIPE_SECRET_KEY"),
		publishableKey:    os.Getenv("STRIPE_PUBLISHABLE_KEY"),
		webhookSecret:     os.Getenv("STRIPE_WEBHOOK_SECRET"),
		priceIDMonthly:    os.Getenv("STRIPE_PRICE_ID_MONTHLY"),
		priceIDYearly:     os.Getenv("STRIPE_PRICE_ID_YEARLY"),
		customerPortalURL: os.Getenv("STRIPE_CUSTOMER_PORTAL_URL"),
	}
}

// CreateCustomer creates a new Stripe customer
func (s *StripeService) CreateCustomer(email, name string) (*Customer, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
	}

	customer, err := customer.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return &Customer{
		ID:        customer.ID,
		Email:     customer.Email,
		Name:      customer.Name,
		CreatedAt: time.Unix(customer.Created, 0),
	}, nil
}

// GetCustomer retrieves a customer by ID
func (s *StripeService) GetCustomer(customerID string) (*Customer, error) {
	customer, err := customer.Get(customerID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &Customer{
		ID:        customer.ID,
		Email:     customer.Email,
		Name:      customer.Name,
		CreatedAt: time.Unix(customer.Created, 0),
	}, nil
}

// CreateSubscription creates a new subscription for a customer
func (s *StripeService) CreateSubscription(customerID, priceID string) (*Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(priceID),
			},
		},
		PaymentSettings: &stripe.SubscriptionPaymentSettingsParams{
			PaymentMethodTypes: []*string{
				stripe.String("card"),
			},
		},
	}

	sub, err := subscription.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return s.convertSubscription(sub), nil
}

// CancelSubscription cancels a subscription
func (s *StripeService) CancelSubscription(subscriptionID string, atPeriodEnd bool) error {
	params := &stripe.SubscriptionParams{}

	if atPeriodEnd {
		params.CancelAtPeriodEnd = stripe.Bool(true)
	} else {
		params.CancelAtPeriodEnd = stripe.Bool(false)
	}

	_, err := subscription.Update(subscriptionID, params)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	return nil
}

// ReactivateSubscription reactivates a cancelled subscription
func (s *StripeService) ReactivateSubscription(subscriptionID string) error {
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(false),
	}

	_, err := subscription.Update(subscriptionID, params)
	if err != nil {
		return fmt.Errorf("failed to reactivate subscription: %w", err)
	}

	return nil
}

// CreatePaymentIntent creates a payment intent for one-time payments
func (s *StripeService) CreatePaymentIntent(amount int64, currency, customerID string) (*PaymentIntent, error) {
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
	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return s.convertSubscription(sub), nil
}

// GetCustomerSubscriptions retrieves all subscriptions for a customer
func (s *StripeService) GetCustomerSubscriptions(customerID string) ([]*Subscription, error) {
	params := &stripe.SubscriptionListParams{
		Customer: stripe.String(customerID),
	}

	iter := subscription.List(params)
	var subscriptions []*Subscription

	for iter.Next() {
		subscriptions = append(subscriptions, s.convertSubscription(iter.Subscription()))
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	return subscriptions, nil
}

// CreateCustomerPortalSession creates a customer portal session
func (s *StripeService) CreateCustomerPortalSession(customerID, returnURL string) (string, error) {
	// This would typically use the Stripe Customer Portal API
	// For now, return the configured portal URL
	return s.customerPortalURL, nil
}

// ValidateWebhookSignature validates the webhook signature
func (s *StripeService) ValidateWebhookSignature(payload []byte, signature string) (*stripe.Event, error) {
	event, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to validate webhook signature: %w", err)
	}

	return &event, nil
}

// ProcessWebhook processes webhook events
func (s *StripeService) ProcessWebhook(event *stripe.Event) error {
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
		return fmt.Errorf("unhandled event type: %s", event.Type)
	}
}

// GetSubscriptionPlans returns available subscription plans
func (s *StripeService) GetSubscriptionPlans() []*SubscriptionPlan {
	return []*SubscriptionPlan{
		{
			ID:          s.priceIDMonthly,
			Name:        "Basic Monthly",
			Price:       9.99,
			Currency:    "usd",
			Interval:    "month",
			Description: "Access to basic content",
			Features:    []string{"Basic video access", "Standard quality", "Email support"},
		},
		{
			ID:          s.priceIDYearly,
			Name:        "Premium Yearly",
			Price:       19.99,
			Currency:    "usd",
			Interval:    "month",
			Description: "Full access with exclusive content",
			Features:    []string{"All video content", "HD quality", "Exclusive content", "Priority support"},
		},
	}
}

// Helper methods
func (s *StripeService) convertSubscription(sub *stripe.Subscription) *Subscription {
	return &Subscription{
		ID:                sub.ID,
		Status:            string(sub.Status),
		CurrentPeriodEnd:  time.Unix(sub.CurrentPeriodEnd, 0),
		CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
		Plan: &SubscriptionPlan{
			ID:       sub.Items.Data[0].Price.ID,
			Price:    float64(sub.Items.Data[0].Price.UnitAmount) / 100,
			Currency: string(sub.Items.Data[0].Price.Currency),
			Interval: string(sub.Items.Data[0].Price.Recurring.Interval),
		},
	}
}

func (s *StripeService) handleSubscriptionCreated(event *stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	// Update database with new subscription
	// Send welcome email
	// Update user access permissions

	return nil
}

func (s *StripeService) handleSubscriptionUpdated(event *stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	// Update database with subscription changes
	// Update user access permissions
	// Send notification if needed

	return nil
}

func (s *StripeService) handleSubscriptionDeleted(event *stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	// Update database with subscription cancellation
	// Revoke user access permissions
	// Send cancellation email

	return nil
}

func (s *StripeService) handlePaymentSucceeded(event *stripe.Event) error {
	var inv stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	// Update payment status in database
	// Send payment confirmation email
	// Update subscription status if needed

	return nil
}

func (s *StripeService) handlePaymentFailed(event *stripe.Event) error {
	var inv stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	// Update payment status in database
	// Send payment failure notification
	// Handle retry logic

	return nil
}
