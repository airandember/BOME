package payment

// This file defines the ports (interfaces) for the payment domain
// Implementations are in the stripe/ subdirectory

import (
	"time"

	"github.com/stripe/stripe-go/v74"
)

// StripePort defines the interface for Stripe payment operations
// Implementation: services/payment/stripe/stripe.go
type StripePort interface {
	// Configuration and Status
	UpdateSecretKey(key string)
	UpdateWebhookSecret(secret string)
	IsEnabled() bool
	GetAccountSummaryWithOptions(section string, limit int) (map[string]interface{}, error)

	// Customer Management
	CreateCustomer(email, name string) (*stripe.Customer, error)
	GetCustomer(customerID string) (*stripe.Customer, error)
	UpdateCustomer(customerID string, params *stripe.CustomerParams) (*stripe.Customer, error)
	DeleteCustomer(customerID string) error
	GetCustomerByEmail(email string) (*stripe.Customer, error)

	// Product & Price Management
	CreateProduct(name, description string) (*stripe.Product, error)
	GetProduct(productID string) (*stripe.Product, error)
	UpdateProduct(productID string, params *stripe.ProductParams) (*stripe.Product, error)
	ListProducts(params *stripe.ProductListParams) *stripe.ProductIter
	CreatePrice(productID string, unitAmount int64, currency string, recurring *stripe.PriceRecurringParams) (*stripe.Price, error)
	GetPrice(priceID string) (*stripe.Price, error)
	ListPrices(params *stripe.PriceListParams) *stripe.PriceIter

	// Subscription Management
	CreateSubscription(customerID, priceID string, params *stripe.SubscriptionParams) (*stripe.Subscription, error)
	GetSubscription(subscriptionID string) (*stripe.Subscription, error)
	UpdateSubscription(subscriptionID string, params *stripe.SubscriptionParams) (*stripe.Subscription, error)
	CancelSubscription(subscriptionID string, atPeriodEnd bool) error
	ListSubscriptions(params *stripe.SubscriptionListParams) *stripe.SubscriptionIter
	GetUpcomingInvoice(customerID string) (*stripe.Invoice, error)

	// Checkout & Billing Portal
	CreateCheckoutSession(customerID, priceID, successURL, cancelURL string) (*stripe.CheckoutSession, error)
	CreateBillingPortalSession(customerID, returnURL string) (*stripe.BillingPortalSession, error)

	// Payments & Refunds
	CreatePaymentIntent(amount int64, currency, customerID string) (*stripe.PaymentIntent, error)
	ConfirmPaymentIntent(paymentIntentID string) (*stripe.PaymentIntent, error)

	// Webhooks
	ConstructWebhookEvent(payload []byte, signature string) (stripe.Event, error)
}

// Refund represents a payment refund
type Refund struct {
	ID            string    `json:"id"`
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	Reason        string    `json:"reason"`
	Status        string    `json:"status"`
	ChargeID      string    `json:"charge_id"`
	PaymentIntent string    `json:"payment_intent"`
	CreatedAt     time.Time `json:"created_at"`
}
