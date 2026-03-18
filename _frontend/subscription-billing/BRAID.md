# 🧬 Subscription & Billing Braid - Frontend
**Svelte5 UI for subscription management and payment processing**

---

## 🔗 **Cross-Repository Braid**

> **⚠️ IMPORTANT**: This is the **frontend portion** of the Subscription & Billing Braid.  
> **Backend portion**: See `_braids/subscription-billing/backend/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## 📋 **Frontend Overview**

**Purpose**: User interface for subscription management, payment processing, and billing  
**Technology**: Svelte 5, TypeScript, Stripe.js, TailwindCSS  
**Entry Points**: `/subscription`, `/checkout`, `/account/billing`  
**State Management**: Svelte stores for subscription and billing data

---

## 🎯 **Key Features**

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

## 📄 **Frontend Pages**

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
            <li>✓ {feature}</li>
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
        🔒 Secure payment powered by Stripe
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
          ⚠️ Your subscription will end on {new Date(subscription.currentPeriodEnd).toLocaleDateString()}
        </div>
      {/if}
    </div>
  </div>
  
  <!-- Payment Method -->
  <div class="payment-method-card">
    <h2>Payment Method</h2>
    <div class="card-info">
      <span class="card-brand">Visa</span>
      <span>•••• {subscription.paymentMethod?.last4 || '****'}</span>
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

## 🧩 **Frontend Components**

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
      <li>✓ {feature}</li>
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

## 🗃️ **Frontend Stores**

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

## 🔄 **Data Flow Examples**

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

## 🎨 **Stripe Elements Integration**

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

## 🔒 **Security**

### **Client-Side**:
- ✅ Never send card details to your server
- ✅ Use Stripe.js tokenization
- ✅ Validate on backend before Stripe API calls
- ✅ HTTPS only

### **Data Handling**:
- ✅ Store only Stripe customer/subscription IDs
- ✅ Display last 4 digits of cards only
- ✅ Use Stripe Customer Portal for sensitive operations

---

## 📝 **Known Issues**

### **To Implement**:
1. Subscription upgrade proration preview
2. Trial period handling in UI
3. Multiple payment methods support
4. Billing address collection
5. Tax calculation display
6. Failed payment retry UI
7. Subscription gifting

---

## 🚀 **Quick Links**

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
[🏠 Master Index](../../../BRAIDS_INDEX.md) | [⬅️ Backend Braid](../../_braids/subscription-billing/backend/BRAID.md)

