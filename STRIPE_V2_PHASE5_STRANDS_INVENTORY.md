# Stripe V2 Phase 5 - Strands Inventory

**Created**: October 30, 2025  
**Purpose**: Document all Stripe-related strands to ensure Phase 5 doesn't break anything

---

## 🎯 **User's Concern**

> "Well also, the /streaming/stripe dashboard is where we set the whsec and webhook health. So there may be strands we should account for so we don't through them out"

**Translation**: Don't break the admin dashboard that manages Stripe configuration!

---

## 📦 **Critical Strands to PRESERVE**

### **1. Stripe Configuration Strand** (KEEP!)

**Location**: `backend/internal/routes/admin_streaming.go`

**Endpoints**:
```go
POST /admin/streaming/stripe/secret          // Set Stripe API secret key (sk_ or rk_)
POST /admin/streaming/stripe/webhook-secret  // Set webhook secret (whsec_)
GET  /admin/streaming/stripe/summary         // Get Stripe account summary
```

**Purpose**: 
- Admins use the frontend dashboard to configure Stripe keys
- Keys are encrypted via `CryptoService` and stored in `secure_settings`
- This is the **WRITE-ONLY FROM FRONTEND** security pattern we documented

**Frontend Pages**:
- `frontend/src/routes/admin/streaming/stripe/setup/+page.svelte` - Key setup form
- `frontend/src/routes/admin/streaming/stripe/webhooks/+page.svelte` - Webhook health

**Status**: ✅ **MUST KEEP** - This is how admins configure Stripe!

---

### **2. Webhook Status & Health Strand** (KEEP!)

**Location**: `backend/internal/routes/stripe_webhook_routes.go`

**Endpoints**:
```go
GET  /webhooks/stripe/status   // Webhook health status (for admin dashboard)
POST /webhooks/stripe/ping     // Test webhook endpoint
GET  /webhooks/stripe/logs     // Webhook event logs (paginated)
POST /webhooks/stripe/retry/:id // Retry failed webhook event
```

**Purpose**:
- Admin dashboard displays webhook health (active/inactive/degraded)
- Shows last event time, success rate, event types
- Allows admins to retry failed events
- Shows real-time webhook activity

**Frontend Pages**:
- `frontend/src/routes/admin/streaming/stripe/webhooks/+page.svelte` - Full webhook dashboard
- `frontend/src/routes/admin/streaming/subscribers/EnhancedSubscribersPageNew.svelte` - Uses webhook auto-sync

**Database Tables Used**:
- `webhook_events` - Stores all webhook events for monitoring

**Status**: ✅ **MUST KEEP** - Critical for monitoring webhook health!

---

### **3. Webhook Event Handler Strand** (MIGRATE TO V2!)

**Location**: `backend/internal/routes/stripe_webhook_routes.go`

**Endpoint**:
```go
POST /webhooks/stripe  // Main webhook handler (public, no auth)
```

**Current Behavior** (v1):
- Validates Stripe signature
- Routes events to handlers
- Writes to v1 tables (`stripe_customers`, `stripe_subscriptions`, etc.)
- **DOES NOT** link customers to users

**Phase 5 Changes** (v2):
- Keep same endpoint URL (don't break Stripe webhook!)
- Keep signature validation
- Keep event routing
- **ADD**: Write to v2 tables
- **ADD**: Auto-link customers to users
- **ADD**: Enforce single subscription logic (Phase 6)
- Dual-write to v1 + v2 during migration

**Status**: ⚠️ **MIGRATE TO V2** - This is the core of Phase 5!

---

### **4. Stripe Service Strand** (KEEP & EXTEND!)

**Location**: `backend/internal/services/stripe_service.go`

**Key Methods**:
```go
UpdateSecretKey(key string)        // Runtime key update
UpdateWebhookSecret(secret string) // Runtime webhook secret update
ValidateWebhookSignature(...)      // Verify Stripe sent the webhook
GetAccountSummary(...)             // Get Stripe account info
```

**Purpose**:
- Manages Stripe API client
- Handles webhook signature validation
- Provides account summary for admin dashboard

**Status**: ✅ **KEEP** - Core infrastructure!

---

### **5. Stripe Sync Service V1** (KEEP FOR NOW!)

**Location**: `backend/internal/services/stripe_sync_service.go`

**Methods**:
```go
UpsertCustomerFromWebhook(customer *stripe.Customer)
UpsertSubscriptionFromWebhook(sub *stripe.Subscription)
UpsertProductFromWebhook(product *stripe.Product)
UpsertPriceFromWebhook(price *stripe.Price)
// ... etc
```

**Purpose**:
- Writes webhook data to v1 tables
- Used by current webhook handler

**Phase 5 Status**: ⏳ **KEEP AS FALLBACK** - Dual-write during migration

---

### **6. Stripe Sync Service V2** (ALREADY BUILT!)

**Location**: `backend/internal/services/stripe_sync_v2.go`

**Methods**:
```go
SyncCustomers()      // Sync all customers to v2
SyncSubscriptions()  // Sync all subscriptions to v2
SyncProducts()       // Sync all products to v2
SyncPrices()         // Sync all prices to v2
```

**Purpose**:
- Syncs Stripe data to v2 tables
- Used by `cmd/stripe-sync` CLI tool

**Phase 5 Enhancement**:
- Add `SyncSingleCustomer(id string)` for webhook use
- Add `SyncSingleSubscription(id string)` for webhook use
- Add `SyncSingleProduct(id string)` for webhook use
- Add `SyncSinglePrice(id string)` for webhook use

**Status**: ✅ **EXTEND** - Add single-entity sync methods

---

### **7. Customer Linking Service** (ALREADY BUILT!)

**Location**: `backend/internal/services/customer_linking_service.go`

**Methods**:
```go
LinkUserToStripeCustomers(email string)  // Link by email
SetPrimaryStripeCustomer(userID, cusID)  // Set primary
GetUserLinkedCustomers(userID)           // Get all linked
```

**Purpose**:
- Links Stripe customers to users via email
- Manages primary customer designation

**Phase 5 Usage**:
- Call `LinkUserToStripeCustomers()` when `customer.created` webhook arrives

**Status**: ✅ **USE IN WEBHOOKS** - Auto-link on customer creation

---

## 🗺️ **Frontend Strands (Do NOT Break!)**

### **Stripe Dashboard Pages**

```
/admin/streaming/stripe
├── /setup              - Set API key & webhook secret ✅
├── /webhooks           - Webhook health & logs ✅
├── /overview           - Stripe account summary ✅
├── /products           - Product management ✅
├── /subscriptions      - Subscription list ✅
├── /invoices           - Invoice list ✅
├── /payments           - Payment history ✅
├── /coupons            - Coupon management ✅
└── /metadata           - Metadata health ✅
```

**All these pages use**:
- `/admin/streaming/stripe/secret` (POST) - Set API key
- `/admin/streaming/stripe/webhook-secret` (POST) - Set webhook secret
- `/admin/streaming/stripe/summary` (GET) - Get account info
- `/webhooks/stripe/status` (GET) - Get webhook health

**Phase 5 Impact**: ✅ **ZERO** - These endpoints remain unchanged!

---

## 🔄 **Phase 5 Migration Strategy**

### **What We're Changing**

| Component | v1 (Old) | Phase 5 (New) | Impact |
|-----------|----------|---------------|--------|
| **Webhook Handler** | Writes to v1 tables | Dual-write (v1 + v2) | ✅ No breaking changes |
| **Customer Sync** | Manual only | Auto-link on webhook | ✅ New feature |
| **Subscription Sync** | No user linking | Links to user | ✅ New feature |
| **Data Tables** | v1 tables only | v1 + v2 (parallel) | ✅ Additive |

### **What We're NOT Changing**

| Component | Status | Why |
|-----------|--------|-----|
| **Webhook URL** | `/webhooks/stripe` | ✅ Stripe config would break |
| **Admin Endpoints** | `/admin/streaming/stripe/*` | ✅ Dashboard would break |
| **Signature Validation** | `StripeService.ValidateWebhookSignature()` | ✅ Security critical |
| **Webhook Status** | `/webhooks/stripe/status` | ✅ Monitoring depends on it |
| **Webhook Logs** | `webhook_events` table | ✅ Audit trail |

---

## 🎯 **Phase 5 Implementation Checklist**

### **Step 1: Extend Services** (No Breaking Changes)

- [ ] Add `SyncSingleCustomer(id)` to `StripeSyncV2Service`
- [ ] Add `SyncSingleSubscription(id)` to `StripeSyncV2Service`
- [ ] Add `SyncSingleProduct(id)` to `StripeSyncV2Service`
- [ ] Add `SyncSinglePrice(id)` to `StripeSyncV2Service`

### **Step 2: Create Webhook Service V2** (New Service)

- [ ] Create `backend/internal/services/stripe_webhook_service_v2.go`
- [ ] Implement `HandleCustomerCreated()` - sync + auto-link
- [ ] Implement `HandleCustomerUpdated()` - sync
- [ ] Implement `HandleSubscriptionCreated()` - sync + link
- [ ] Implement `HandleSubscriptionUpdated()` - sync
- [ ] Implement `HandleSubscriptionDeleted()` - sync + remove access
- [ ] Implement `HandleProductCreated/Updated()` - sync
- [ ] Implement `HandlePriceCreated/Updated()` - sync

### **Step 3: Update Webhook Routes** (Dual-Write)

- [ ] Update `HandleStripeWebhook()` to accept v2 services
- [ ] Implement dual-write logic (v1 + v2)
- [ ] Keep v1 as fallback
- [ ] Log which system succeeded/failed
- [ ] Update `routes.go` to pass v2 services

### **Step 4: Test Webhook Flow** (No Downtime)

- [ ] Test `customer.created` → v2 sync + auto-link
- [ ] Test `customer.subscription.created` → v2 sync + link to user
- [ ] Test `customer.subscription.updated` → v2 sync
- [ ] Test `customer.subscription.deleted` → v2 sync + video access removal
- [ ] Verify admin dashboard still works
- [ ] Verify webhook health page still works
- [ ] Verify webhook logs still work

### **Step 5: Monitor Dual-Write** (1 Week)

- [ ] Compare v1 vs v2 data after each webhook
- [ ] Log discrepancies
- [ ] Fix any edge cases
- [ ] Verify customer linking works

---

## 🚨 **What NOT to Delete**

**DO NOT DELETE** (even after Phase 5):

```
✅ backend/internal/routes/admin_streaming.go
   └── Stripe configuration endpoints (secret, webhook-secret, summary)

✅ backend/internal/routes/stripe_webhook_routes.go
   └── Webhook status, logs, ping, retry endpoints

✅ backend/internal/services/stripe_service.go
   └── Core Stripe client and signature validation

✅ frontend/src/routes/admin/streaming/stripe/**
   └── All admin dashboard pages

✅ webhook_events table
   └── Audit trail for webhook monitoring
```

**CAN DELETE** (after Phase 10 cutover):

```
⏳ backend/internal/services/stripe_sync_service.go (v1)
   └── Old sync service (only after v2 proven stable)

⏳ stripe_customers, stripe_subscriptions (v1 tables)
   └── Old tables (only after data migrated to v2)
```

**DELETE NOW** (fragmentation):

```
❌ backend/subscription/handlers/stripe_webhook_routes.go
❌ _backend/subscription-billing/.../stripe_webhook_routes.go
❌ _braids/subscription-billing/.../stripe_webhook_routes.go
❌ _BRAIDS/subscription-billing/.../stripe_webhook_routes.go
```

---

## 📊 **Summary**

### **Strands Inventory**

| Strand | Location | Phase 5 Action | Frontend Impact |
|--------|----------|----------------|-----------------|
| **Stripe Config** | `admin_streaming.go` | ✅ Keep unchanged | ✅ Zero |
| **Webhook Status** | `stripe_webhook_routes.go` | ✅ Keep unchanged | ✅ Zero |
| **Webhook Handler** | `stripe_webhook_routes.go` | ⚠️ Migrate to v2 | ✅ Zero |
| **Stripe Service** | `stripe_service.go` | ✅ Keep unchanged | ✅ Zero |
| **Sync Service V1** | `stripe_sync_service.go` | ⏳ Keep as fallback | ✅ Zero |
| **Sync Service V2** | `stripe_sync_v2.go` | ✅ Extend with single-sync | ✅ Zero |
| **Customer Linking** | `customer_linking_service.go` | ✅ Use in webhooks | ✅ Zero |

**Key Principle**: Phase 5 is **ADDITIVE**, not destructive. We're adding v2 functionality alongside v1, not replacing it yet!

---

**Ready to proceed with Phase 5?** 🚀

All critical strands are documented and will be preserved!

