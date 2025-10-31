# Stripe V2 - Phase 5: Webhook Migration Plan

**Created**: October 30, 2025  
**Status**: 🔄 Planning  
**Goal**: Update webhooks to use v2 tables + customer linking

---

## 🎯 **Phase 5 Objectives**

1. **✅ Maintain BRAIDS Architecture** (you asked for this!)
2. **Update webhook handlers** to use `StripeSyncV2Service`
3. **Auto-link customers** when they're created via webhook
4. **Handle single subscription** business logic
5. **Zero downtime** migration (webhooks keep working)

---

## 🔍 **Current State Assessment**

### **Current Webhook Setup**

| Component | Location | Status |
|-----------|----------|--------|
| **Main Handler** | `backend/internal/routes/stripe_webhook_routes.go` | ✅ Exists, uses v1 |
| **Service** | Uses `StripeSyncService` (v1) | ⚠️ Needs v2 |
| **Tables** | Writes to `stripe_customers`, `stripe_subscriptions`, etc | ⚠️ Needs v2 |
| **Customer Linking** | ❌ None | ❌ Critical gap! |
| **Stripe Config** | 56 events configured | ✅ Active |
| **Endpoint** | `https://watch.bookofmormonevidence.org/bome-backend/api/v1/webhooks/stripe` | ✅ Active |

### **Webhook Events Currently Handled (v1)**

```go
// Customer events
✅ customer.created
✅ customer.updated
✅ customer.deleted

// Subscription events
✅ customer.subscription.created
✅ customer.subscription.updated
✅ customer.subscription.deleted

// Invoice events
✅ invoice.payment_succeeded
✅ invoice.payment_failed

// Product/Price events
✅ product.created
✅ product.updated
✅ price.created
✅ price.updated

// Ignored events (56 total configured, we process ~12)
📋 charge.*, checkout.session.*, etc (logged but not processed)
```

---

## 🚨 **CRITICAL FRAGMENTATION DETECTED!**

Your webhooks are **split across multiple locations**:

```
✅ backend/internal/routes/stripe_webhook_routes.go  (ACTIVE - uses v1)
❌ backend/subscription/handlers/stripe_webhook_routes.go (OLD BRAID - duplicate!)
❌ _backend/subscription-billing/layers/business-logic/handlers/stripe_webhook_routes.go (OLD)
❌ _braids/subscription-billing/backend/layers/business-logic/handlers/stripe_webhook_routes.go (OLD)
❌ _BRAIDS/subscription-billing/backend/layers/business-logic/handlers/stripe_webhook_routes.go (OLD)
```

**The duplicates are NOT being used**, but they exist and cause confusion!

---

## 🏗️ **BRAIDS Architecture for Webhooks**

### **Your Question: "We set up a webhooks braid correct?"**

**Answer**: YES, but it's currently embedded in routes, not a standalone BRAID. Here's how it SHOULD be structured:

```
STRIPE BRAID (Payment Processing)
├── ELASTIC SERVICE: StripeSyncV2Service (Single Source of Truth)
│   ├── Sync from Stripe API (manual/scheduled)
│   └── Sync from Webhooks (real-time)
│
├── STRAND: Webhook Handler (Real-time updates)
│   ├── POST /webhooks/stripe (public, no auth)
│   ├── Signature validation (Stripe-Signature header)
│   ├── Event routing (customer, subscription, product, price)
│   └── Delegates to StripeSyncV2Service
│
├── STRAND: Customer Linking (User <-> Stripe Customer)
│   ├── Auto-link on webhook events (customer.created)
│   ├── Manual linking via admin API
│   └── Primary customer management
│
└── STRAND: Subscription Management (Single Sub Rule)
    ├── Cancel old subscriptions when new one created
    ├── Update user video access
    └── Handle subscription lifecycle
```

**Key Principle**: Webhooks are a **STRAND** (input mechanism) that feeds the **ELASTIC SERVICE** (single source of truth).

---

## 📋 **Phase 5 Implementation Plan**

### **Step 1: Create StripeWebhookServiceV2** (NEW!)

**Why?** Keep webhook logic separate from routes (Single Responsibility)

**File**: `backend/internal/services/stripe_webhook_service_v2.go`

```go
type StripeWebhookServiceV2 struct {
    syncService    *StripeSyncV2Service
    linkingService *CustomerLinkingService
    db             *database.DB
}

// HandleCustomerCreated - sync to v2 tables + auto-link
func (s *StripeWebhookServiceV2) HandleCustomerCreated(customer *stripe.Customer) error {
    // 1. Sync to stripe_customers_v2
    if err := s.syncService.SyncSingleCustomer(customer.ID); err != nil {
        return err
    }
    
    // 2. Auto-link to user by email
    if customer.Email != "" {
        if err := s.linkingService.LinkUserToStripeCustomers(customer.Email); err != nil {
            log.Printf("⚠️  Auto-linking failed for %s: %v", customer.Email, err)
            // Don't fail the webhook, just log
        }
    }
    
    return nil
}

// HandleSubscriptionCreated - sync + enforce single subscription
func (s *StripeWebhookServiceV2) HandleSubscriptionCreated(subscription *stripe.Subscription) error {
    // 1. Sync to stripe_subscriptions_v2
    if err := s.syncService.SyncSingleSubscription(subscription.ID); err != nil {
        return err
    }
    
    // 2. Get the user via customer linking
    user, err := s.linkingService.GetUserByStripeCustomer(subscription.Customer.ID)
    if err != nil {
        log.Printf("⚠️  No user found for customer %s", subscription.Customer.ID)
        return nil // Not fatal - might be a customer without account
    }
    
    // 3. Cancel other active subscriptions for this user (Phase 6 logic)
    if err := s.enforceS ingleSubscription(user.ID, subscription.ID); err != nil {
        log.Printf("⚠️  Failed to enforce single subscription: %v", err)
        // Don't fail the webhook
    }
    
    return nil
}

// More handlers...
```

### **Step 2: Update Webhook Routes** (MODIFY EXISTING)

**File**: `backend/internal/routes/stripe_webhook_routes.go`

**Changes**:
- Add `webhookServiceV2 *services.StripeWebhookServiceV2` parameter
- Add `linkingService *services.CustomerLinkingService` parameter
- Update all `handleXxx` functions to use v2 service
- Keep v1 as fallback during migration (dual-write)

```go
// Updated handler signature
func HandleStripeWebhook(
    c *gin.Context, 
    stripeService *services.StripeService, 
    syncServiceV1 *services.StripeSyncService,      // Keep for fallback
    syncServiceV2 *services.StripeSyncV2Service,     // NEW!
    linkingService *services.CustomerLinkingService, // NEW!
) {
    // ... existing validation ...
    
    // Route to v2 handlers
    err = processV1EventV2(event, syncServiceV2, linkingService)
    
    // ... rest of logic ...
}

// Updated event processor
func processV1EventV2(
    event *stripe.Event, 
    syncService *services.StripeSyncV2Service,
    linkingService *services.CustomerLinkingService,
) error {
    switch event.Type {
    case "customer.created":
        return handleCustomerCreatedV2(event, syncService, linkingService)
    case "customer.subscription.created":
        return handleSubscriptionCreatedV2(event, syncService, linkingService)
    // ... other events ...
    }
}
```

### **Step 3: Add Customer Linking Logic**

**Key Business Rules**:

1. **When `customer.created` webhook arrives**:
   - Sync to `stripe_customers_v2`
   - If `customer.email` matches a `users.email`, create link in `user_stripe_customers_v2`
   - Set `is_primary = true` if user has no other customers

2. **When `customer.subscription.created` webhook arrives**:
   - Sync to `stripe_subscriptions_v2`
   - Find user via `user_stripe_customers_v2`
   - Cancel any OTHER active subscriptions for this user (Phase 6)
   - Update user's video access

3. **When `customer.subscription.deleted` webhook arrives**:
   - Mark subscription as deleted in `stripe_subscriptions_v2`
   - Remove user's video access (if no other active subs)

### **Step 4: Update routes.go Initialization**

**File**: `backend/internal/routes/routes.go`

```go
// Initialize v2 services
stripeSyncV2Service := services.NewStripeSyncV2Service(db)
customerLinkingService := services.NewCustomerLinkingService(db)
webhookServiceV2 := services.NewStripeWebhookServiceV2(
    stripeSyncV2Service,
    customerLinkingService,
    db,
)

// Register webhook routes with v2 services
webhooks := api.Group("/webhooks")
{
    webhooks.POST("/stripe", func(c *gin.Context) {
        HandleStripeWebhook(
            c, 
            stripeService, 
            stripeSyncService,      // v1 (keep for now)
            stripeSyncV2Service,    // v2
            customerLinkingService, // v2
        )
    })
}
```

### **Step 5: Dual-Write Period** (Safety!)

**Strategy**: Write to BOTH v1 and v2 tables during migration

```go
func handleCustomerCreatedDualWrite(event *stripe.Event, v1 *services.StripeSyncService, v2 *services.StripeSyncV2Service, linking *services.CustomerLinkingService) error {
    var customer stripe.Customer
    json.Unmarshal(event.Data.Raw, &customer)
    
    // Write to v1 (existing system)
    if err := v1.UpsertCustomerFromWebhook(&customer); err != nil {
        log.Printf("⚠️  V1 write failed: %v", err)
        // Don't fail the webhook
    }
    
    // Write to v2 (new system)
    if err := v2.SyncSingleCustomer(customer.ID); err != nil {
        log.Printf("❌ V2 write failed: %v", err)
        return err // Fail webhook if v2 fails
    }
    
    // Auto-link (v2 only)
    if customer.Email != "" {
        linking.LinkUserToStripeCustomers(customer.Email)
    }
    
    return nil
}
```

**Duration**: 1 week (monitor for discrepancies)

### **Step 6: Testing Plan**

**Test 1: Customer Creation**
```bash
# Create a test customer in Stripe Dashboard
# Expected:
# 1. Row in stripe_customers_v2
# 2. Row in user_stripe_customers_v2 (if email matches)
# 3. is_primary = true (if first customer for user)
```

**Test 2: Subscription Creation**
```bash
# Create a test subscription in Stripe Dashboard
# Expected:
# 1. Row in stripe_subscriptions_v2
# 2. User's video_access updated (if linked)
```

**Test 3: Webhook Status Endpoint**
```bash
GET /api/v1/webhooks/stripe/status
# Should show recent activity
```

---

## 🎯 **Success Criteria**

✅ Webhooks write to v2 tables  
✅ Customers auto-link to users by email  
✅ Subscription creation updates user video access  
✅ No downtime during migration  
✅ All 56 events acknowledged (12 processed, 44 logged)  
✅ Webhook logs show "success" status  
✅ Zero fragmentation (old files deleted)

---

## 🗑️ **Cleanup (After Phase 5)**

**Delete these duplicate files**:
```
backend/subscription/handlers/stripe_webhook_routes.go
_backend/subscription-billing/layers/business-logic/handlers/stripe_webhook_routes.go
_braids/subscription-billing/backend/layers/business-logic/handlers/stripe_webhook_routes.go
_BRAIDS/subscription-billing/backend/layers/business-logic/handlers/stripe_webhook_routes.go
```

**Keep**:
```
backend/internal/routes/stripe_webhook_routes.go (updated to v2)
backend/internal/services/stripe_webhook_service_v2.go (NEW)
```

---

## 🔗 **Related Phases**

- **Phase 4 (Completed)**: SubscriberElasticServiceV2 (reads v2 data)
- **Phase 5 (This Phase)**: Webhooks (writes v2 data)
- **Phase 6 (Next)**: Single subscription enforcement
- **Phase 7 (After)**: Frontend subscription management

---

**Ready to implement Phase 5?** 🚀

Just say **"Let's build Phase 5!"** and I'll:
1. Create `StripeWebhookServiceV2`
2. Update webhook routes
3. Add customer linking logic
4. Implement dual-write safety
5. Clean up fragmentation

