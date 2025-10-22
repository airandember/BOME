package ports

import "time"

// StripePort defines the interface for Stripe payment operations
type StripePort interface {
	// Service Configuration
	IsEnabled() bool
	UpdateSecretKey(key string)
	UpdateWebhookSecret(secret string)

	// Customer Management
	CreateCustomer(email, name string, metadata map[string]string) (string, error)
	GetCustomer(customerID string) (*StripeCustomer, error)
	UpdateCustomer(customerID string, updates map[string]interface{}) error
	DeleteCustomer(customerID string) error

	// Subscription Management
	CreateSubscription(customerID, priceID string) (*StripeSubscription, error)
	GetSubscription(subscriptionID string) (*StripeSubscription, error)
	CancelSubscription(subscriptionID string, atPeriodEnd bool) error
	UpdateSubscription(subscriptionID string, updates map[string]interface{}) error

	// Payment Methods
	AttachPaymentMethod(paymentMethodID, customerID string) error
	DetachPaymentMethod(paymentMethodID string) error
	SetDefaultPaymentMethod(customerID, paymentMethodID string) error

	// Invoices
	GetInvoice(invoiceID string) (*StripeInvoice, error)
	GetUpcomingInvoice(customerID string) (*StripeInvoice, error)

	// Refunds
	CreateRefund(subscriptionID string, amount int64, reason string) (*StripeRefund, error)

	// Products & Prices
	ListProducts() ([]*StripeProduct, error)
	ListPrices(productID string) ([]*StripePrice, error)
	CreatePrice(productID string, amount int64, currency, interval string) (*StripePrice, error)

	// Account Summary
	GetAccountSummary() (map[string]interface{}, error)
	GetAccountSummaryWithOptions(section string, limit int) (map[string]interface{}, error)

	// Webhook Handling
	ConstructEvent(payload []byte, signature string) (interface{}, error)
}

// StripeCustomer represents a Stripe customer
type StripeCustomer struct {
	ID       string
	Email    string
	Name     string
	Metadata map[string]string
	Created  time.Time
}

// StripeSubscription represents a Stripe subscription
type StripeSubscription struct {
	ID                 string
	CustomerID         string
	Status             string
	CurrentPeriodEnd   time.Time
	CurrentPeriodStart time.Time
	CancelAtPeriodEnd  bool
	Items              []StripeSubscriptionItem
}

// StripeSubscriptionItem represents a subscription line item
type StripeSubscriptionItem struct {
	ID       string
	PriceID  string
	Quantity int64
}

// StripeInvoice represents a Stripe invoice
type StripeInvoice struct {
	ID               string
	CustomerID       string
	AmountDue        int64
	AmountPaid       int64
	Status           string
	PeriodStart      time.Time
	PeriodEnd        time.Time
	HostedInvoiceURL string
}

// StripeRefund represents a Stripe refund
type StripeRefund struct {
	ID      string
	Amount  int64
	Status  string
	Reason  string
	Created time.Time
}

// StripeProduct represents a Stripe product
type StripeProduct struct {
	ID          string
	Name        string
	Description string
	Active      bool
	Metadata    map[string]string
}

// StripePrice represents a Stripe price
type StripePrice struct {
	ID         string
	ProductID  string
	UnitAmount int64
	Currency   string
	Recurring  *StripePriceRecurring
	Active     bool
}

// StripePriceRecurring represents recurring price details
type StripePriceRecurring struct {
	Interval      string // day, week, month, year
	IntervalCount int64
}
