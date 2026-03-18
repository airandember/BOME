# Braid: subscription-billing

**Architecture:** Full-Stack Braid (Frontend to Backend)
**Last Updated:** 2025-10-17

---

## Backend Architecture

**Stripe-powered revenue engine for subscription management**

---

## ðŸ”— **Cross-Repository Braid**

> **âš ï¸ IMPORTANT**: This is the **backend portion** of the Subscription & Billing Braid.  
> **Frontend portion**: See `_frontend/braids/subscription-billing/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## ðŸ“‹ **Backend Overview**

**Purpose**: Server-side subscription management, payment processing, and Stripe integration  
**Technology**: Go, PostgreSQL, Stripe API  
**Complexity**: **Very High** (Third-party integration, webhooks, billing logic)  
**Priority**: **CRITICAL** - Revenue generation depends on this!

---

## 📁 **Production File Map**

### **Backend Files (Go)**
```
backend/
├── subscription/
│   ├── handlers/subscription.go       # Subscription routes
│   ├── services/stripe.go             # Stripe integration
│   ├── services/stripe_sync.go        # Stripe sync
│   └── models/subscription.go, stripe_entities.go
├── services/
│   ├── payment/stripe/                # Payment services
│   └── stripe/                        # Stripe sync, webhooks
├── internal/
│   ├── routes/
│   │   ├── stripe_webhook_routes.go   # Webhook handling
│   │   ├── subscription.go
│   │   ├── subscription_plans.go
│   │   └── stripe_*.go
│   └── services/stripe*.go
└── braids/subscription-checkout/       # Checkout flow sub-braid
```

### **Frontend Files (Svelte)**
```
frontend/src/
├── routes/checkout/                    # Checkout flow
└── routes/subscription/                # Subscription management
```

---

## ðŸŽ¯ **Key Features**

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

## ðŸ—„ï¸ **Database Schema**

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

## ðŸŒ **API Endpoints**

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

### **Stripe Webhooks** (âš ï¸ CRITICAL):
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

## ðŸ”§ **Backend Services**

### **Stripe Service** (`backend/internal/services/stripe.go`):

**âš ï¸ WARNING**: This is a **massive file** with complex Stripe integration!

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

## ðŸŽ£ **Webhook Processing**

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

## ðŸ’° **Subscription Lifecycle**

### **Complete Flow**:

```
1. User Registration
   â””â”€> Create Stripe Customer
       â””â”€> Store customer_id in database

2. Select Plan
   â””â”€> User chooses plan from /subscription-plans

3. Add Payment Method
   â””â”€> Stripe Elements collects card
       â””â”€> Create PaymentMethod on Stripe
           â””â”€> Attach to Customer

4. Create Subscription
   â””â”€> POST /api/v1/subscriptions
       â””â”€> Stripe creates subscription
           â””â”€> Webhook: customer.subscription.created
               â””â”€> Update database
                   â””â”€> Grant user access

5. Recurring Billing
   â””â”€> Stripe automatically charges
       â””â”€> Webhook: invoice.paid
           â””â”€> Extend subscription period
               â””â”€> Send receipt email

6. Payment Failure
   â””â”€> Webhook: invoice.payment_failed
       â””â”€> Mark subscription past_due
           â””â”€> Send failed payment email
               â””â”€> Retry payment (Stripe Smart Retries)

7. Cancellation
   â””â”€> POST /api/v1/subscriptions/:id (cancel)
       â””â”€> Cancel at period end
           â””â”€> Webhook: customer.subscription.updated
               â””â”€> Update database
                   â””â”€> Send confirmation email

8. Subscription Ends
   â””â”€> Webhook: customer.subscription.deleted
       â””â”€> Revoke user access
           â””â”€> Downgrade to free plan
```

---

## ðŸ”’ **Security & Compliance**

### **PCI Compliance**:
- âœ… Never store card numbers directly
- âœ… Use Stripe Elements for card collection
- âœ… Token-based payment method handling
- âœ… HTTPS only for all payment endpoints
- âœ… Webhook signature verification

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

## âš¡ **Performance Optimizations**

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

## ðŸ“ **Known Technical Debt**

### **Current Issues**:
1. âš ï¸ **Massive stripe.go file** - needs refactoring
2. âš ï¸ **Limited webhook error handling** - can silently fail
3. âš ï¸ **No webhook retry queue** - relies on Stripe retries
4. âš ï¸ **Subscription downgrade logic** - proration issues
5. âš ï¸ **Failed payment recovery** - basic implementation

### **Future Enhancements**:
1. âœ… Split stripe.go into smaller services
2. âœ… Add webhook event queue (Redis/SQS)
3. âœ… Implement advanced retry logic
4. âœ… Add subscription analytics
5. âœ… Support multiple payment methods per customer
6. âœ… Add subscription gifting
7. âœ… Implement usage-based billing

---

## ðŸš€ **Quick Start**

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
[ðŸ  Master Index](../../BRAIDS_INDEX.md) | [ðŸŽ¨ Frontend Braid](../../_frontend/braids/subscription-billing/BRAID.md)



---

## Frontend Architecture

**Svelte5 UI for subscription management and payment processing**

---

## ðŸ”— **Cross-Repository Braid**

> **âš ï¸ IMPORTANT**: This is the **frontend portion** of the Subscription & Billing Braid.  
> **Backend portion**: See `_braids/subscription-billing/backend/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## ðŸ“‹ **Frontend Overview**

**Purpose**: User interface for subscription management, payment processing, and billing  
**Technology**: Svelte 5, TypeScript, Stripe.js, TailwindCSS  
**Entry Points**: `/subscription`, `/checkout`, `/account/billing`  
**State Management**: Svelte stores for subscription and billing data

---

## ðŸŽ¯ **Key Features**

### **1. Subscription Plans Page**:
- Display available plans
- Feature comparison table
- Pricing toggle (monthly/annual)
- Highlighted "popular" plan
- CTA buttons for each plan
- FAQ section

### **2. Checkout Flow**:
- Plan selection
- Stripe Elements payment form
- Secure card collection
- Processing indicators
- Success/error handling
- 3D Secure support

### **3. Account Billing**:
- Current subscription status
- Next billing date
- Payment method management
- Invoice history
- Download receipts
- Cancel subscription

### **4. Subscription Management**:
- Upgrade/downgrade plans
- Cancel subscription
- Reactivate subscription
- Update payment method
- Apply coupon codes
- Access billing portal

### **5. Admin Subscription Management**:
- View all subscriptions
- User subscription details
- Revenue analytics
- Refund processing
- Webhook event logs
- Stripe sync tools

---

## ðŸ“„ **Frontend Pages**

### **1. Subscription Plans** (`/subscription`)
**File**: `frontend/src/routes/subscription/+page.svelte`

**Features**:
- Pricing cards for each plan
- Monthly/Annual toggle
- Feature comparison
- "Get Started" CTAs
- Responsive grid layout

**Example UI**:
```svelte
<script lang="ts">
  import { goto } from '$app/navigation';
  import { auth } from '$lib/auth';
  
  let billingInterval = 'month'; // or 'year'
  
  const plans = [
    {
      name: 'Free',
      price: 0,
      features: ['Limited content', 'SD quality', 'Ads']
    },
    {
      name: 'Basic',
      price: billingInterval === 'month' ? 9.99 : 99.99,
      popular: false,
      features: ['Unlimited content', 'HD quality', 'No ads']
    },
    {
      name: 'Premium',
      price: billingInterval === 'month' ? 19.99 : 199.99,
      popular: true,
      features: ['Unlimited content', '4K quality', 'No ads', 'Offline downloads', 'Priority support']
    }
  ];
  
  function selectPlan(plan) {
    if (!$auth.user) {
      goto('/register?plan=' + plan.name);
    } else {
      goto('/checkout?plan=' + plan.name + '&interval=' + billingInterval);
    }
  }
</script>

<div class="subscription-plans">
  <header>
    <h1>Choose Your Plan</h1>
    <p>Unlock unlimited access to premium content</p>
    
    <!-- Billing Interval Toggle -->
    <div class="billing-toggle">
      <button
        class:active={billingInterval === 'month'}
        on:click={() => billingInterval = 'month'}
      >
        Monthly
      </button>
      <button
        class:active={billingInterval === 'year'}
        on:click={() => billingInterval = 'year'}
      >
        Annual (Save 17%)
      </button>
    </div>
  </header>
  
  <div class="plans-grid">
    {#each plans as plan}
      <div class="plan-card" class:popular={plan.popular}>
        {#if plan.popular}
          <div class="badge">Most Popular</div>
        {/if}
        
        <h3>{plan.name}</h3>
        <div class="price">
          <span class="amount">${plan.price}</span>
          <span class="period">/{billingInterval}</span>
        </div>
        
        <ul class="features">
          {#each plan.features as feature}
            <li>âœ“ {feature}</li>
          {/each}
        </ul>
        
        <button
          class="cta"
          class:primary={plan.popular}
          on:click={() => selectPlan(plan)}
        >
          {plan.price === 0 ? 'Get Started' : 'Subscribe Now'}
        </button>
      </div>
    {/each}
  </div>
</div>

<style>
  .plans-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 2rem;
    max-width: 1200px;
    margin: 0 auto;
  }
  
  .plan-card {
    border: 2px solid #e0e0e0;
    border-radius: 12px;
    padding: 2rem;
    position: relative;
  }
  
  .plan-card.popular {
    border-color: #6366f1;
    transform: scale(1.05);
  }
  
  .badge {
    position: absolute;
    top: -12px;
    right: 20px;
    background: #6366f1;
    color: white;
    padding: 4px 12px;
    border-radius: 20px;
    font-size: 0.875rem;
  }
</style>
```

---

### **2. Checkout Page** (`/checkout`)
**File**: `frontend/src/routes/checkout/+page.svelte`

**Features**:
- Plan summary
- Stripe Elements payment form
- Coupon code input
- Terms acceptance
- Processing state
- Error handling

**Example UI**:
```svelte
<script lang="ts">
  import { loadStripe } from '@stripe/stripe-js';
  import { Elements, PaymentElement, useStripe, useElements } from '@stripe/stripe-svelte';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  
  let stripe;
  let clientSecret = '';
  let processing = false;
  let error = '';
  
  const planName = $page.url.searchParams.get('plan');
  const interval = $page.url.searchParams.get('interval');
  
  onMount(async () => {
    // Initialize Stripe
    stripe = await loadStripe(import.meta.env.VITE_STRIPE_PUBLIC_KEY);
    
    // Create payment intent on backend
    const res = await fetch('/api/v1/subscriptions/create-payment-intent', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ plan: planName, interval })
    });
    
    const data = await res.json();
    clientSecret = data.clientSecret;
  });
  
  async function handleSubmit() {
    if (!stripe || !clientSecret) return;
    
    processing = true;
    error = '';
    
    try {
      const { error: stripeError } = await stripe.confirmPayment({
        elements,
        confirmParams: {
          return_url: window.location.origin + '/checkout/success'
        }
      });
      
      if (stripeError) {
        error = stripeError.message;
      }
    } catch (e) {
      error = 'Payment failed. Please try again.';
    } finally {
      processing = false;
    }
  }
</script>

<div class="checkout-page">
  <div class="checkout-container">
    <!-- Order Summary -->
    <div class="order-summary">
      <h2>Order Summary</h2>
      <div class="plan-details">
        <h3>{planName} Plan</h3>
        <p>Billed {interval}ly</p>
        <div class="price">
          <span>$19.99</span>
          <span>per {interval}</span>
        </div>
      </div>
      
      <!-- Coupon Code -->
      <div class="coupon-input">
        <input type="text" placeholder="Coupon code" />
        <button>Apply</button>
      </div>
      
      <div class="total">
        <span>Total due today</span>
        <span class="amount">$19.99</span>
      </div>
    </div>
    
    <!-- Payment Form -->
    <div class="payment-form">
      <h2>Payment Information</h2>
      
      {#if clientSecret}
        <Elements {stripe} {clientSecret}>
          <form on:submit|preventDefault={handleSubmit}>
            <PaymentElement />
            
            {#if error}
              <div class="error-message">{error}</div>
            {/if}
            
            <button
              type="submit"
              disabled={processing || !stripe}
              class="submit-button"
            >
              {processing ? 'Processing...' : 'Subscribe Now'}
            </button>
          </form>
        </Elements>
      {:else}
        <div class="loading">Loading payment form...</div>
      {/if}
      
      <div class="security-note">
        ðŸ”’ Secure payment powered by Stripe
      </div>
    </div>
  </div>
</div>
```

---

### **3. Account Billing** (`/account/billing`)
**File**: `frontend/src/routes/account/billing/+page.svelte`

**Features**:
- Current subscription card
- Next billing date countdown
- Payment method on file
- Update payment button
- Cancel subscription button
- Invoice list with download

**Example UI**:
```svelte
<script lang="ts">
  import { onMount } from 'svelte';
  import { subscriptionStore } from '$lib/stores/subscription';
  
  onMount(async () => {
    await subscriptionStore.loadSubscription();
  });
  
  $: subscription = $subscriptionStore.current;
  $: invoices = $subscriptionStore.invoices;
  
  function openBillingPortal() {
    // Opens Stripe Customer Portal
    subscriptionStore.openBillingPortal();
  }
  
  async function cancelSubscription() {
    if (confirm('Are you sure you want to cancel?')) {
      await subscriptionStore.cancelSubscription();
    }
  }
</script>

<div class="billing-page">
  <h1>Billing & Subscription</h1>
  
  <!-- Current Subscription -->
  <div class="subscription-card">
    <div class="header">
      <div>
        <h2>{subscription.plan.name} Plan</h2>
        <p class="status" class:active={subscription.status === 'active'}>
          {subscription.status}
        </p>
      </div>
      <button on:click={() => goto('/subscription')}>
        Change Plan
      </button>
    </div>
    
    <div class="billing-info">
      <div class="info-row">
        <span>Next Billing Date</span>
        <span class="value">{new Date(subscription.currentPeriodEnd).toLocaleDateString()}</span>
      </div>
      <div class="info-row">
        <span>Amount</span>
        <span class="value">${subscription.plan.price}/{subscription.interval}</span>
      </div>
      {#if subscription.cancelAtPeriodEnd}
        <div class="warning">
          âš ï¸ Your subscription will end on {new Date(subscription.currentPeriodEnd).toLocaleDateString()}
        </div>
      {/if}
    </div>
  </div>
  
  <!-- Payment Method -->
  <div class="payment-method-card">
    <h2>Payment Method</h2>
    <div class="card-info">
      <span class="card-brand">Visa</span>
      <span>â€¢â€¢â€¢â€¢ {subscription.paymentMethod?.last4 || '****'}</span>
      <span>Exp: {subscription.paymentMethod?.expMonth}/{subscription.paymentMethod?.expYear}</span>
    </div>
    <button on:click={openBillingPortal}>
      Update Payment Method
    </button>
  </div>
  
  <!-- Invoice History -->
  <div class="invoices-section">
    <h2>Invoice History</h2>
    <table>
      <thead>
        <tr>
          <th>Date</th>
          <th>Amount</th>
          <th>Status</th>
          <th>Invoice</th>
        </tr>
      </thead>
      <tbody>
        {#each invoices as invoice}
          <tr>
            <td>{new Date(invoice.date).toLocaleDateString()}</td>
            <td>${invoice.amount}</td>
            <td>
              <span class="status" class:paid={invoice.status === 'paid'}>
                {invoice.status}
              </span>
            </td>
            <td>
              <a href={invoice.pdfUrl} target="_blank">Download</a>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
  
  <!-- Danger Zone -->
  <div class="danger-zone">
    <h2>Cancel Subscription</h2>
    <p>Your subscription will remain active until the end of the billing period.</p>
    <button class="danger" on:click={cancelSubscription}>
      Cancel Subscription
    </button>
  </div>
</div>
```

---

### **4. Checkout Success** (`/checkout/success`)
**File**: `frontend/src/routes/checkout/success/+page.svelte`

**Features**:
- Success message
- Subscription confirmation
- Next steps
- Redirect to dashboard

---

### **5. Admin Subscriptions** (`/admin/streaming/subscriptions`)
**File**: `frontend/src/routes/admin/streaming/subscriptions/+page.svelte`

**Features**:
- All subscriptions list
- Filter by status
- Search by user
- Revenue metrics
- Export to CSV

---

### **6. Admin Stripe Webhooks** (`/admin/streaming/stripe/webhooks`)
**File**: `frontend/src/routes/admin/streaming/stripe/webhooks/+page.svelte`

**Features**:
- Webhook event log
- Event type filter
- Processing status
- Retry failed webhooks
- Event details view

---

## ðŸ§© **Frontend Components**

### **SubscriptionCard Component**:
**Purpose**: Display subscription status

```svelte
<script>
  export let subscription;
</script>

<div class="subscription-card">
  <h3>{subscription.plan.name}</h3>
  <p class="status">{subscription.status}</p>
  <p>Next billing: {subscription.nextBillingDate}</p>
</div>
```

---

### **PricingCard Component**:
**Purpose**: Display a pricing plan

```svelte
<script>
  export let plan;
  export let interval = 'month';
  export let popular = false;
</script>

<div class="pricing-card" class:popular>
  {#if popular}
    <span class="badge">Most Popular</span>
  {/if}
  <h3>{plan.name}</h3>
  <div class="price">
    ${plan.price[interval]}
    <span>/{interval}</span>
  </div>
  <ul>
    {#each plan.features as feature}
      <li>âœ“ {feature}</li>
    {/each}
  </ul>
  <button>Select Plan</button>
</div>
```

---

### **InvoiceTable Component**:
**Purpose**: Display invoice history

```svelte
<script>
  export let invoices = [];
</script>

<table class="invoice-table">
  <thead>
    <tr>
      <th>Date</th>
      <th>Amount</th>
      <th>Status</th>
      <th>Actions</th>
    </tr>
  </thead>
  <tbody>
    {#each invoices as invoice}
      <tr>
        <td>{invoice.date}</td>
        <td>${invoice.amount}</td>
        <td>{invoice.status}</td>
        <td>
          <a href={invoice.pdfUrl}>Download</a>
        </td>
      </tr>
    {/each}
  </tbody>
</table>
```

---

## ðŸ—ƒï¸ **Frontend Stores**

### **Subscription Store** (`$lib/stores/subscription.ts`):
**Purpose**: Manage subscription state

```typescript
interface SubscriptionState {
  current: Subscription | null;
  plans: SubscriptionPlan[];
  invoices: Invoice[];
  loading: boolean;
  error: string | null;
}

export const subscriptionStore = {
  async loadSubscription() {
    // GET /api/v1/subscriptions
  },
  
  async loadPlans() {
    // GET /api/v1/subscription-plans
  },
  
  async loadInvoices() {
    // GET /api/v1/billing/invoices
  },
  
  async createSubscription(planId: number, paymentMethodId: string) {
    // POST /api/v1/subscriptions
  },
  
  async updateSubscription(planId: number) {
    // PUT /api/v1/subscriptions/:id
  },
  
  async cancelSubscription() {
    // DELETE /api/v1/subscriptions/:id
  },
  
  async openBillingPortal() {
    // POST /api/v1/billing/portal
    // Redirects to Stripe Customer Portal
  }
};
```

---

## ðŸ”„ **Data Flow Examples**

### **Subscribe to Plan**:
```
1. User clicks "Subscribe" on plan
2. Redirect to /checkout?plan=premium&interval=month
3. Load Stripe.js
4. Create payment intent on backend
5. Display Stripe Elements form
6. User enters card details
7. Submit payment
8. Stripe processes payment
9. Webhook: customer.subscription.created
10. Backend updates database
11. Redirect to /checkout/success
12. User has active subscription!
```

### **Cancel Subscription**:
```
1. User clicks "Cancel Subscription"
2. Confirmation dialog
3. POST /api/v1/subscriptions/:id (cancel)
4. Backend cancels with Stripe
5. Webhook: customer.subscription.updated
6. Update UI with "cancels at period end"
7. Send confirmation email
```

---

## ðŸŽ¨ **Stripe Elements Integration**

### **Using Stripe Elements**:
```svelte
<script>
  import { loadStripe } from '@stripe/stripe-js';
  import { Elements, PaymentElement } from '@stripe/stripe-svelte';
  
  let stripe;
  let clientSecret = '';
  
  onMount(async () => {
    stripe = await loadStripe(STRIPE_PUBLIC_KEY);
    
    // Get client secret from backend
    const res = await fetch('/api/create-payment-intent');
    const data = await res.json();
    clientSecret = data.clientSecret;
  });
</script>

{#if stripe && clientSecret}
  <Elements {stripe} {clientSecret}>
    <form on:submit={handleSubmit}>
      <PaymentElement />
      <button type="submit">Pay</button>
    </form>
  </Elements>
{/if}
```

---

## ðŸ”’ **Security**

### **Client-Side**:
- âœ… Never send card details to your server
- âœ… Use Stripe.js tokenization
- âœ… Validate on backend before Stripe API calls
- âœ… HTTPS only

### **Data Handling**:
- âœ… Store only Stripe customer/subscription IDs
- âœ… Display last 4 digits of cards only
- âœ… Use Stripe Customer Portal for sensitive operations

---

## ðŸ“ **Known Issues**

### **To Implement**:
1. Subscription upgrade proration preview
2. Trial period handling in UI
3. Multiple payment methods support
4. Billing address collection
5. Tax calculation display
6. Failed payment retry UI
7. Subscription gifting

---

## ðŸš€ **Quick Links**

**Actual Files**:
- Subscription Page: `frontend/src/routes/subscription/+page.svelte`
- Checkout: `frontend/src/routes/checkout/+page.svelte`
- Billing: `frontend/src/routes/account/billing/+page.svelte`
- Admin Subscriptions: `frontend/src/routes/admin/streaming/subscriptions/+page.svelte`
- Webhook Log: `frontend/src/routes/admin/streaming/stripe/webhooks/+page.svelte`

---

**Last Updated**: October 14, 2025  
**Status**: Critical revenue UI  
**Technology**: Svelte 5 + Stripe.js  
**Backend Counterpart**: `_braids/subscription-billing/backend/`

---

**Navigate**:  
[ðŸ  Master Index](../../../BRAIDS_INDEX.md) | [â¬…ï¸ Backend Braid](../../_braids/subscription-billing/backend/BRAID.md)



---

## Integration Notes

- Frontend: `_braids/subscription-billing/frontend/`
- Backend: `_braids/subscription-billing/backend/`

This braid represents a complete vertical slice of functionality.

