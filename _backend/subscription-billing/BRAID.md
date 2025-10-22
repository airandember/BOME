# 🧬 Subscription & Billing Braid - Backend
**Stripe-powered revenue engine for subscription management**

---

## 🔗 **Cross-Repository Braid**

> **⚠️ IMPORTANT**: This is the **backend portion** of the Subscription & Billing Braid.  
> **Frontend portion**: See `_frontend/braids/subscription-billing/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## 📋 **Backend Overview**

**Purpose**: Server-side subscription management, payment processing, and Stripe integration  
**Technology**: Go, PostgreSQL, Stripe API  
**Complexity**: **Very High** (Third-party integration, webhooks, billing logic)  
**Priority**: **CRITICAL** - Revenue generation depends on this!

**Critical Files**:
- `backend/internal/routes/stripe_webhook_routes.go` (862 lines!)
- `backend/internal/services/stripe.go` (Massive service file)
- `backend/internal/database/subscription.go`
- `backend/internal/database/stripe_entities.go`

---

## 🎯 **Key Features**

### **1. Subscription Plans**:
- Multiple tier system (Free, Basic, Premium, Enterprise)
- Monthly/Annual billing intervals
- Feature-based access control
- Plan upgrades/downgrades
- Proration handling
- Trial periods

### **2. Stripe Integration**:
- Customer creation/sync
- Payment method management
- Subscription lifecycle management
- Invoice generation
- Webhook event processing
- Price ID synchronization
- Product catalog sync

### **3. Payment Processing**:
- Secure payment handling (PCI compliant)
- Multiple payment methods
- Automatic retry logic
- Failed payment recovery
- Refund processing
- Payment history tracking

### **4. Webhook Management**:
- Real-time subscription updates
- Payment success/failure notifications
- Invoice events
- Customer updates
- Signature verification
- Idempotency handling

### **5. Billing Operations**:
- Invoice generation and delivery
- Payment receipts
- Billing history
- Subscription cancellation
- Refund processing
- Billing portal access

---

## 🗄️ **Database Schema**

### **Subscription Plans Table**:
**File**: `backend/internal/database/subscription_plans.go`

```sql
CREATE TABLE subscription_plans (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    stripe_product_id VARCHAR(255) UNIQUE,
    stripe_price_id VARCHAR(255) UNIQUE,
    price DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    billing_interval VARCHAR(20) NOT NULL, -- 'month', 'year'
    trial_days INTEGER DEFAULT 0,
    features JSONB,
    is_active BOOLEAN DEFAULT true,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_subscription_plans_stripe_product ON subscription_plans(stripe_product_id);
CREATE INDEX idx_subscription_plans_stripe_price ON subscription_plans(stripe_price_id);
CREATE INDEX idx_subscription_plans_active ON subscription_plans(is_active);

COMMENT ON TABLE subscription_plans IS 'Subscription plan definitions and Stripe product/price mappings';
COMMENT ON COLUMN subscription_plans.features IS 'JSON array of feature flags for this plan';
```

**Example Plans**:
```json
{
  "free": {
    "name": "free",
    "price": 0,
    "features": ["limited_videos", "sd_quality"]
  },
  "basic": {
    "name": "basic",
    "price": 9.99,
    "features": ["unlimited_videos", "hd_quality", "no_ads"]
  },
  "premium": {
    "name": "premium",
    "price": 19.99,
    "features": ["unlimited_videos", "4k_quality", "no_ads", "offline_downloads", "priority_support"]
  }
}
```

---

### **User Subscriptions Table**:
**File**: `backend/internal/database/subscription.go`

```sql
CREATE TABLE user_subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    subscription_plan_id INTEGER REFERENCES subscription_plans(id),
    stripe_customer_id VARCHAR(255),
    stripe_subscription_id VARCHAR(255) UNIQUE,
    status VARCHAR(50) NOT NULL, -- 'active', 'canceled', 'past_due', 'trialing', 'incomplete', 'incomplete_expired'
    current_period_start TIMESTAMP,
    current_period_end TIMESTAMP,
    cancel_at_period_end BOOLEAN DEFAULT false,
    canceled_at TIMESTAMP,
    trial_start TIMESTAMP,
    trial_end TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, subscription_plan_id)
);

CREATE INDEX idx_user_subscriptions_user_id ON user_subscriptions(user_id);
CREATE INDEX idx_user_subscriptions_stripe_customer ON user_subscriptions(stripe_customer_id);
CREATE INDEX idx_user_subscriptions_stripe_subscription ON user_subscriptions(stripe_subscription_id);
CREATE INDEX idx_user_subscriptions_status ON user_subscriptions(status);
CREATE INDEX idx_user_subscriptions_period_end ON user_subscriptions(current_period_end);

COMMENT ON TABLE user_subscriptions IS 'Active user subscriptions linked to Stripe';
COMMENT ON COLUMN user_subscriptions.status IS 'Stripe subscription status: active, canceled, past_due, trialing, etc.';
```

---

### **Stripe Entities Table**:
**File**: `backend/internal/database/stripe_entities.go`

```sql
CREATE TABLE stripe_entities (
    id SERIAL PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL, -- 'customer', 'subscription', 'invoice', 'payment_intent', 'price'
    stripe_id VARCHAR(255) NOT NULL UNIQUE,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    data JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_stripe_entities_type ON stripe_entities(entity_type);
CREATE INDEX idx_stripe_entities_stripe_id ON stripe_entities(stripe_id);
CREATE INDEX idx_stripe_entities_user_id ON stripe_entities(user_id);

COMMENT ON TABLE stripe_entities IS 'Cached Stripe API responses for quick lookups';
COMMENT ON COLUMN stripe_entities.data IS 'Full Stripe API response JSON';
```

**Purpose**: Cache Stripe API data to reduce API calls and improve performance.

---

### **Invoices Table**:
```sql
CREATE TABLE invoices (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    subscription_id INTEGER REFERENCES user_subscriptions(id) ON DELETE SET NULL,
    stripe_invoice_id VARCHAR(255) UNIQUE,
    amount_due DECIMAL(10, 2),
    amount_paid DECIMAL(10, 2),
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(50), -- 'draft', 'open', 'paid', 'void', 'uncollectible'
    invoice_pdf_url VARCHAR(500),
    hosted_invoice_url VARCHAR(500),
    due_date TIMESTAMP,
    paid_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_invoices_user_id ON invoices(user_id);
CREATE INDEX idx_invoices_stripe_invoice_id ON invoices(stripe_invoice_id);
CREATE INDEX idx_invoices_status ON invoices(status);
```

---

### **Payment History Table**:
```sql
CREATE TABLE payment_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    subscription_id INTEGER REFERENCES user_subscriptions(id) ON DELETE SET NULL,
    stripe_payment_intent_id VARCHAR(255),
    stripe_charge_id VARCHAR(255),
    amount DECIMAL(10, 2),
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(50), -- 'succeeded', 'pending', 'failed'
    payment_method_type VARCHAR(50), -- 'card', 'bank_transfer', etc.
    last_four VARCHAR(4),
    failure_code VARCHAR(100),
    failure_message TEXT,
    receipt_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_payment_history_user_id ON payment_history(user_id);
CREATE INDEX idx_payment_history_status ON payment_history(status);
CREATE INDEX idx_payment_history_created_at ON payment_history(created_at);
```

---

### **Stripe Webhooks Log Table**:
```sql
CREATE TABLE stripe_webhook_events (
    id SERIAL PRIMARY KEY,
    stripe_event_id VARCHAR(255) UNIQUE NOT NULL,
    event_type VARCHAR(100) NOT NULL, -- 'customer.subscription.updated', 'invoice.paid', etc.
    data JSONB NOT NULL,
    processed BOOLEAN DEFAULT false,
    processed_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_stripe_webhook_events_type ON stripe_webhook_events(event_type);
CREATE INDEX idx_stripe_webhook_events_processed ON stripe_webhook_events(processed);
CREATE INDEX idx_stripe_webhook_events_stripe_id ON stripe_webhook_events(stripe_event_id);

COMMENT ON TABLE stripe_webhook_events IS 'Stripe webhook event log for debugging and idempotency';
```

**Purpose**: Prevent duplicate webhook processing and provide audit trail.

---

## 🌐 **API Endpoints**

### **Subscription Management**:
**File**: `backend/internal/routes/subscription.go`

```
GET    /api/v1/subscriptions                    # Get user's current subscription
POST   /api/v1/subscriptions                    # Create new subscription
PUT    /api/v1/subscriptions/:id                # Update subscription
DELETE /api/v1/subscriptions/:id                # Cancel subscription
POST   /api/v1/subscriptions/:id/resume         # Resume canceled subscription
POST   /api/v1/subscriptions/:id/upgrade        # Upgrade subscription plan
POST   /api/v1/subscriptions/:id/downgrade      # Downgrade subscription plan
```

---

### **Subscription Plans**:
**File**: `backend/internal/routes/subscription_plans.go`

```
GET    /api/v1/subscription-plans               # List all available plans
GET    /api/v1/subscription-plans/:id           # Get specific plan details
POST   /api/v1/admin/subscription-plans         # Create plan (admin)
PUT    /api/v1/admin/subscription-plans/:id     # Update plan (admin)
DELETE /api/v1/admin/subscription-plans/:id     # Delete plan (admin)
```

---

### **Stripe Webhooks** (⚠️ CRITICAL):
**File**: `backend/internal/routes/stripe_webhook_routes.go` (862 lines!)

```
POST   /api/v1/stripe/webhooks                  # Stripe webhook endpoint
```

**Webhook Events Handled**:
```go
// Customer events
- customer.created
- customer.updated
- customer.deleted

// Subscription events
- customer.subscription.created
- customer.subscription.updated
- customer.subscription.deleted
- customer.subscription.trial_will_end

// Payment events
- invoice.created
- invoice.finalized
- invoice.paid
- invoice.payment_failed
- invoice.payment_action_required

// Payment Intent events
- payment_intent.succeeded
- payment_intent.payment_failed
- payment_intent.canceled

// Charge events
- charge.succeeded
- charge.failed
- charge.refunded
```

---

### **Billing & Invoices**:
```
GET    /api/v1/billing/invoices                 # List user invoices
GET    /api/v1/billing/invoices/:id             # Get invoice details
GET    /api/v1/billing/payment-methods          # List payment methods
POST   /api/v1/billing/payment-methods          # Add payment method
DELETE /api/v1/billing/payment-methods/:id      # Remove payment method
POST   /api/v1/billing/portal                   # Get Stripe customer portal URL
```

---

### **Admin Subscription Management**:
```
GET    /api/v1/admin/subscriptions              # List all subscriptions
GET    /api/v1/admin/subscriptions/:id          # Get subscription details
POST   /api/v1/admin/subscriptions/:id/cancel   # Force cancel (admin)
POST   /api/v1/admin/subscriptions/:id/refund   # Issue refund (admin)
GET    /api/v1/admin/revenue-analytics          # Revenue reporting
```

---

## 🔧 **Backend Services**

### **Stripe Service** (`backend/internal/services/stripe.go`):

**⚠️ WARNING**: This is a **massive file** with complex Stripe integration!

**Key Functions**:
```go
// Customer management
func CreateStripeCustomer(user *User) (*stripe.Customer, error)
func GetOrCreateStripeCustomer(userID int) (*stripe.Customer, error)
func UpdateStripeCustomer(customerID string, params *stripe.CustomerParams) error

// Subscription management
func CreateSubscription(userID int, planID int, paymentMethodID string) (*stripe.Subscription, error)
func UpdateSubscription(subscriptionID string, params *stripe.SubscriptionParams) (*stripe.Subscription, error)
func CancelSubscription(subscriptionID string, cancelImmediately bool) (*stripe.Subscription, error)
func ReactivateSubscription(subscriptionID string) (*stripe.Subscription, error)

// Plan management
func UpgradeSubscription(subscriptionID string, newPriceID string) error
func DowngradeSubscription(subscriptionID string, newPriceID string) error

// Payment methods
func AttachPaymentMethod(customerID string, paymentMethodID string) error
func SetDefaultPaymentMethod(customerID string, paymentMethodID string) error
func DetachPaymentMethod(paymentMethodID string) error

// Invoice handling
func FetchInvoices(customerID string) ([]*stripe.Invoice, error)
func VoidInvoice(invoiceID string) error
func PayInvoice(invoiceID string) error

// Portal access
func CreateBillingPortalSession(customerID string, returnURL string) (*stripe.BillingPortalSession, error)
```

---

### **Stripe Sync Service** (`backend/internal/services/stripe_sync.go`):

**Purpose**: Keep local database in sync with Stripe

**Key Functions**:
```go
// Sync from Stripe to database
func SyncStripeCustomer(customerID string) error
func SyncStripeSubscription(subscriptionID string) error
func SyncStripePrices() error
func SyncStripeProducts() error

// Full sync operations
func SyncAllCustomers() error
func SyncAllSubscriptions() error

// Verification
func VerifyDatabaseSync() ([]string, error) // Returns list of discrepancies
```

---

### **Stripe Customers Service** (`backend/internal/services/stripe_customers.go`):

**Purpose**: Customer-specific operations

**Key Functions**:
```go
func GetCustomerSubscriptions(customerID string) ([]*stripe.Subscription, error)
func GetCustomerInvoices(customerID string) ([]*stripe.Invoice, error)
func GetCustomerPaymentMethods(customerID string) ([]*stripe.PaymentMethod, error)
func UpdateCustomerEmail(customerID string, email string) error
```

---

### **Stripe Coupons Service** (`backend/internal/services/stripe_coupons.go`):

**Purpose**: Discount and coupon management

**Key Functions**:
```go
func CreateCoupon(code string, percentOff int, duration string) (*stripe.Coupon, error)
func ApplyCoupon(subscriptionID string, couponCode string) error
func RemoveCoupon(subscriptionID string) error
func ValidateCoupon(code string) (*stripe.Coupon, error)
```

---

## 🎣 **Webhook Processing**

### **Webhook Handler** (`backend/internal/routes/stripe_webhook_routes.go`):

**Critical Flow**:
```go
func HandleStripeWebhook(c *fiber.Ctx) error {
    // 1. Verify Stripe signature
    signature := c.Get("Stripe-Signature")
    event, err := webhook.ConstructEvent(payload, signature, webhookSecret)
    if err != nil {
        return c.Status(400).SendString("Invalid signature")
    }
    
    // 2. Check for duplicate events (idempotency)
    exists := CheckEventProcessed(event.ID)
    if exists {
        return c.Status(200).SendString("Already processed")
    }
    
    // 3. Log event
    LogWebhookEvent(event)
    
    // 4. Route to specific handler
    switch event.Type {
        case "customer.subscription.updated":
            return handleSubscriptionUpdated(event)
        case "invoice.paid":
            return handleInvoicePaid(event)
        case "invoice.payment_failed":
            return handleInvoicePaymentFailed(event)
        // ... many more cases
    }
    
    // 5. Mark as processed
    MarkEventProcessed(event.ID)
    
    return c.Status(200).SendString("Webhook processed")
}
```

### **Key Webhook Handlers**:

**Subscription Updated**:
```go
func handleSubscriptionUpdated(event stripe.Event) error {
    var subscription stripe.Subscription
    json.Unmarshal(event.Data.Raw, &subscription)
    
    // Update database
    UpdateSubscriptionInDB(&subscription)
    
    // Update user access
    UpdateUserAccess(subscription.Customer.ID, subscription.Status)
    
    // Send notification
    if subscription.Status == "canceled" {
        SendCancellationEmail(subscription.Customer.Email)
    }
    
    return nil
}
```

**Invoice Payment Failed**:
```go
func handleInvoicePaymentFailed(event stripe.Event) error {
    var invoice stripe.Invoice
    json.Unmarshal(event.Data.Raw, &invoice)
    
    // Log failed payment
    LogFailedPayment(&invoice)
    
    // Update subscription status
    UpdateSubscriptionStatus(invoice.Subscription, "past_due")
    
    // Send payment failed email
    SendPaymentFailedEmail(invoice.CustomerEmail, invoice.AmountDue)
    
    // Trigger retry logic
    SchedulePaymentRetry(invoice.ID)
    
    return nil
}
```

---

## 💰 **Subscription Lifecycle**

### **Complete Flow**:

```
1. User Registration
   └─> Create Stripe Customer
       └─> Store customer_id in database

2. Select Plan
   └─> User chooses plan from /subscription-plans

3. Add Payment Method
   └─> Stripe Elements collects card
       └─> Create PaymentMethod on Stripe
           └─> Attach to Customer

4. Create Subscription
   └─> POST /api/v1/subscriptions
       └─> Stripe creates subscription
           └─> Webhook: customer.subscription.created
               └─> Update database
                   └─> Grant user access

5. Recurring Billing
   └─> Stripe automatically charges
       └─> Webhook: invoice.paid
           └─> Extend subscription period
               └─> Send receipt email

6. Payment Failure
   └─> Webhook: invoice.payment_failed
       └─> Mark subscription past_due
           └─> Send failed payment email
               └─> Retry payment (Stripe Smart Retries)

7. Cancellation
   └─> POST /api/v1/subscriptions/:id (cancel)
       └─> Cancel at period end
           └─> Webhook: customer.subscription.updated
               └─> Update database
                   └─> Send confirmation email

8. Subscription Ends
   └─> Webhook: customer.subscription.deleted
       └─> Revoke user access
           └─> Downgrade to free plan
```

---

## 🔒 **Security & Compliance**

### **PCI Compliance**:
- ✅ Never store card numbers directly
- ✅ Use Stripe Elements for card collection
- ✅ Token-based payment method handling
- ✅ HTTPS only for all payment endpoints
- ✅ Webhook signature verification

### **Webhook Security**:
```go
func VerifyWebhookSignature(payload []byte, signature string) (*stripe.Event, error) {
    webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
    
    event, err := webhook.ConstructEvent(payload, signature, webhookSecret)
    if err != nil {
        return nil, fmt.Errorf("webhook signature verification failed: %v", err)
    }
    
    return &event, nil
}
```

### **Idempotency**:
```go
func CheckEventProcessed(eventID string) bool {
    var count int
    db.QueryRow("SELECT COUNT(*) FROM stripe_webhook_events WHERE stripe_event_id = $1 AND processed = true", eventID).Scan(&count)
    return count > 0
}
```

---

## ⚡ **Performance Optimizations**

### **Caching**:
- **Subscription plans**: Cached for 30 minutes
- **Stripe customer**: Cached in stripe_entities table
- **Active subscriptions**: Cached per user for 5 minutes

### **Webhook Processing**:
- Async processing for non-critical events
- Idempotency checks prevent duplicates
- Queue-based retry for failures

### **Database Optimization**:
- Indexed foreign keys
- Compound index on (user_id, status)
- Partial indexes for active subscriptions

---

## 📝 **Known Technical Debt**

### **Current Issues**:
1. ⚠️ **Massive stripe.go file** - needs refactoring
2. ⚠️ **Limited webhook error handling** - can silently fail
3. ⚠️ **No webhook retry queue** - relies on Stripe retries
4. ⚠️ **Subscription downgrade logic** - proration issues
5. ⚠️ **Failed payment recovery** - basic implementation

### **Future Enhancements**:
1. ✅ Split stripe.go into smaller services
2. ✅ Add webhook event queue (Redis/SQS)
3. ✅ Implement advanced retry logic
4. ✅ Add subscription analytics
5. ✅ Support multiple payment methods per customer
6. ✅ Add subscription gifting
7. ✅ Implement usage-based billing

---

## 🚀 **Quick Start**

### **Understanding Subscriptions** (20 min):
1. Read this BRAID.md (10 min)
2. Review stripe_webhook_routes.go (5 min)
3. Check database schema (5 min)

### **Debugging Subscriptions**:
1. Check stripe_webhook_events table for event processing
2. Verify stripe_entities has customer synced
3. Check user_subscriptions status field
4. Review payment_history for failed payments
5. Use Stripe Dashboard for event logs

---

**Last Updated**: October 14, 2025  
**Status**: Critical revenue system  
**Technology**: Go + Stripe API  
**Frontend Counterpart**: `_frontend/braids/subscription-billing/`

---

**Navigate**:  
[🏠 Master Index](../../BRAIDS_INDEX.md) | [🎨 Frontend Braid](../../_frontend/braids/subscription-billing/BRAID.md)

