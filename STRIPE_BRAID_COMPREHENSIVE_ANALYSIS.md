# 🔍 STRIPE BRAID COMPREHENSIVE ANALYSIS & REBUILD PLAN

**Date**: October 26, 2025  
**Status**: 🚨 **CRITICAL - COMPLETE REBUILD RECOMMENDED**  
**Strategy**: **Build New Braid Alongside Old, Then Swap**

---

## 📊 **PART 1: STRIPE API ANALYSIS**

### **What Stripe Sends Us (from [docs.stripe.com](https://docs.stripe.com/api/customers/object))**

#### **Customer Object** (`/v1/customers/:id`)
```json
{
  "id": "cus_NffrFeUfNV2Hib",
  "object": "customer",
  "email": "jennyrosen@example.com",
  "name": "Jenny Rosen",
  "phone": null,
  "address": null,
  "metadata": {},
  "created": 1680893993,
  "balance": 0,
  "currency": null,
  "default_source": null,
  "delinquent": false,
  "invoice_prefix": "0759376C",
  "livemode": false,
  "subscriptions": null,  // ⚠️ NOT INCLUDED unless expanded!
  "sources": null,
  "tax_exempt": "none"
}
```

**Key Insights**:
- ✅ We get: `id`, `email`, `name`, `phone`, `address`, `metadata`, `created`, `balance`
- ❌ We DON'T get subscriptions by default (must call separately or use `expand[]=subscriptions`)
- ⚠️ `subscriptions` is a LIST object (separate API endpoint: `/v1/subscriptions?customer=:id`)

#### **Subscription Object** (`/v1/subscriptions/:id`)
```json
{
  "id": "sub_1SBqYrFpxJJNWdU8nY1inXI4",
  "object": "subscription",
  "customer": "cus_NffrFeUfNV2Hib",
  "status": "active",
  "current_period_start": 1680893993,
  "current_period_end": 1683485993,
  "created": 1680893993,
  "cancel_at_period_end": false,
  "items": {
    "data": [{
      "id": "si_xxxxx",
      "price": {
        "id": "price_1XXXXXX",
        "product": "prod_XXXXXX",
        "unit_amount": 2000,
        "currency": "usd",
        "recurring": {
          "interval": "month"
        }
      }
    }]
  },
  "metadata": {}
}
```

**Key Insights**:
- ✅ We get: `id`, `customer` (cus_xxx), `status`, `current_period_start`, `current_period_end`
- ✅ We get: `items.data[].price` (nested price object with `product`, `unit_amount`, `currency`, `recurring`)
- ⚠️ We need to extract price and product info from nested structure

#### **Price Object** (`/v1/prices/:id`)
```json
{
  "id": "price_1XXXXXX",
  "object": "price",
  "product": "prod_XXXXXX",
  "unit_amount": 2000,
  "currency": "usd",
  "recurring": {
    "interval": "month",
    "interval_count": 1
  },
  "active": true,
  "metadata": {}
}
```

#### **Product Object** (`/v1/products/:id`)
```json
{
  "id": "prod_XXXXXX",
  "object": "product",
  "name": "Premium Plan",
  "description": "Full access to streaming library",
  "active": true,
  "metadata": {},
  "created": 1680893993
}
```

---

## 🔍 **PART 2: CURRENT SYSTEM FRAGMENTATION ANALYSIS**

### **Your Current Tables** (Inferred from Code)

#### **Table 1: `stripe_customers`**
```sql
CREATE TABLE stripe_customers (
    id SERIAL PRIMARY KEY,                -- Internal ID (integer)
    stripe_id VARCHAR(255) UNIQUE,        -- cus_xxxxx
    name TEXT,
    email TEXT,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    metadata JSONB
    -- ❌ NO foreign key to users table!
);
```

**Problems**:
1. ❌ No link to `users` table (requires email matching or array scan)
2. ❌ No way to mark "primary" customer for a user
3. ❌ No audit trail (when linked, by whom, why)

---

#### **Table 2: `stripe_subscriptions`**
```sql
CREATE TABLE stripe_subscriptions (
    id SERIAL PRIMARY KEY,
    stripe_id VARCHAR(255) UNIQUE,        -- sub_xxxxx
    customer_id INTEGER,                  -- FK to stripe_customers.id (integer)
    status TEXT,
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    created_at TIMESTAMPTZ,
    price_id INTEGER,                     -- FK to stripe_prices.id (integer)
    stripe_price_id TEXT,                 -- price_xxxxx (redundant?)
    unit_amount BIGINT,                   -- cents
    currency TEXT,
    stripe_product_id TEXT,               -- prod_xxxxx (redundant?)
    product_name TEXT                     -- denormalized
    -- ❌ NO FK constraint to stripe_customers!
);
```

**Problems**:
1. ❌ No FK constraint on `customer_id` (orphaned records possible)
2. ❌ Mixed integer IDs (`price_id`) and string IDs (`stripe_price_id`) - confusing!
3. ❌ Denormalized data (`product_name`, `unit_amount`, `currency`) - can go stale
4. ❌ No way to find subscription for a user without complex join

---

#### **Table 3: `stripe_products`**
```sql
CREATE TABLE stripe_products (
    id SERIAL PRIMARY KEY,
    stripe_id VARCHAR(255) UNIQUE,        -- prod_xxxxx
    name TEXT,
    description TEXT,
    active BOOLEAN,
    available BOOLEAN,                    -- ❓ custom field?
    video_approved BOOLEAN,               -- ❓ custom field?
    livemode BOOLEAN,
    legacy_product BOOLEAN,               -- ❓ custom field?
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
    -- ✅ Looks reasonable
);
```

**Problems**:
1. ⚠️ Custom fields (`available`, `video_approved`, `legacy_product`) not from Stripe
2. ❌ No FK from `stripe_subscriptions` to this table

---

#### **Table 4: `stripe_prices`**
```sql
CREATE TABLE stripe_prices (
    id SERIAL PRIMARY KEY,
    stripe_id VARCHAR(255) UNIQUE,        -- price_xxxxx
    product_id VARCHAR(255),              -- ❌ Should be integer FK!
    unit_amount BIGINT,
    currency TEXT,
    recurring_interval TEXT,              -- 'month', 'year'
    active BOOLEAN,
    created_at TIMESTAMPTZ
    -- ❌ NO FK constraint to stripe_products!
);
```

**Problems**:
1. ❌ `product_id` is VARCHAR but should FK to `stripe_products.stripe_id`
2. ❌ No FK constraint (orphaned prices possible)
3. ❌ `stripe_subscriptions.price_id` references this table's integer `id`, but prices link to products by string `product_id` - **TYPE MISMATCH HELL**

---

#### **Table 5: `users`**
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE,
    stripe_customer_id TEXT,              -- PRIMARY customer (cus_xxxxx)
    stripe_customer_ids TEXT[],           -- ALL customers (array)
    -- ... many other fields ...
    -- ❌ NO foreign keys to stripe_customers!
);
```

**Problems**:
1. ❌ Array field `stripe_customer_ids` requires unnesting for joins (slow!)
2. ❌ No link table for many-to-many relationship
3. ❌ No way to know WHEN a customer was linked or WHY

---

### **Your Current Join Path (THE NIGHTMARE)**

**To find a user's active subscription**:
```sql
SELECT ...
FROM users u
LEFT JOIN stripe_customers sc ON sc.stripe_id = ANY(u.stripe_customer_ids)  -- 😱 ARRAY SCAN
LEFT JOIN stripe_subscriptions ss ON ss.customer_id = sc.id                  -- FK join (good)
LEFT JOIN stripe_prices sp ON sp.id = ss.price_id                             -- FK join (good)
LEFT JOIN stripe_products prod ON prod.stripe_id = sp.product_id             -- ❌ STRING JOIN (bad)
WHERE u.id = 9797;
```

**Problems**:
1. 🐌 **Array unnesting** for every query (O(n*m) complexity)
2. 🐌 **String matching** on `prod.stripe_id = sp.product_id` (no index possible)
3. 🐌 **4-table join** for basic question "does user have subscription?"

---

## 🚨 **PART 3: REAL-WORLD FRAGMENTATION DATA**

### **Users with Multiple Customer IDs**

From your diagnostic query, **13 users** have multiple `cus_` numbers:

```
7356   "robberch@gmail.com"          - 2 customers, 1 active subscription
7369   "joyfullavatar@gmail.com"     - 2 customers, 1 active subscription
7374   "jameskersey2@gmail.com"      - 3 customers, 3 ACTIVE SUBSCRIPTIONS (!!)
7470   "dbates62@hotmail.com"        - 2 customers, 0 active subscriptions
7475   "jillypill1@yahoo.com"        - 2 customers, 1 active subscription
7680   "sherryjohns@hotmail.com"     - 2 customers, 0 active subscriptions
7780   "garrettreichert@hotmail.com" - 2 customers, 0 active subscriptions
7788   "shauna_math@outlook.com"     - 2 customers, 0 active subscriptions
8264   "lbar3351@gmail.com"          - 2 customers, 0 active subscriptions
9797   "ericgessel@gmail.com"        - 2 customers, 1 active subscription (FIXED)
10423  "gay.martin@gmail.com"        - 2 customers, 1 active subscription
10447  "nelsonlarren@gmail.com"      - 6 CUSTOMERS, 6 ACTIVE SUBSCRIPTIONS (🚨🚨🚨)
10448  "chirohorses1951@gmail.com"   - 4 CUSTOMERS, 4 ACTIVE SUBSCRIPTIONS (🚨🚨🚨)
```

### **🚨 CRITICAL FINDINGS**:

1. **User 7374** (`jameskersey2@gmail.com`): Has **3 ACTIVE SUBSCRIPTIONS**
   - This should be impossible per your requirements ("only one active subscription at a time")
   - Indicates: Webhook not canceling old subs when new ones created

2. **User 10447** (`nelsonlarren@gmail.com`): Has **6 ACTIVE SUBSCRIPTIONS**
   - All created within ~3 hours on Oct 18, 2025
   - Looks like checkout failures/retries that all succeeded
   - User is being charged 6x!

3. **User 10448** (`chirohorses1951@gmail.com`): Has **4 ACTIVE SUBSCRIPTIONS**
   - All created within ~12 minutes on Oct 22, 2025
   - Same issue as above

4. **Users with historical customers**:
   - 7356, 7369, 7475: Have old customers from 2020, new customers from 2024-2025
   - Indicates: Resubscribes after cancellation, but old customers not purged

---

## 🎯 **PART 4: IDEAL SCHEMA (NEW BRAID)**

### **Design Principles**

1. ✅ **Single source of truth**: Stripe is the source, we cache it
2. ✅ **Foreign key integrity**: All relationships enforced by database
3. ✅ **Integer PKs, String Stripe IDs**: Fast joins, clear separation
4. ✅ **Link table for users ↔ customers**: Explicit many-to-many
5. ✅ **Audit trail**: Track when/how/why data changes
6. ✅ **Normalized**: No duplicate data (except strategic denormalization)

---

### **NEW SCHEMA (v2)**

#### **Table 1: `stripe_customers_v2`**
```sql
CREATE TABLE stripe_customers_v2 (
    id SERIAL PRIMARY KEY,
    stripe_id VARCHAR(255) UNIQUE NOT NULL,  -- cus_xxxxx
    email VARCHAR(512),
    name VARCHAR(255),
    phone VARCHAR(50),
    address JSONB,
    metadata JSONB,
    balance INTEGER DEFAULT 0,               -- cents
    currency VARCHAR(3),                     -- 'usd', 'eur', etc.
    delinquent BOOLEAN DEFAULT false,
    
    -- Stripe timestamps
    stripe_created_at TIMESTAMPTZ NOT NULL,
    
    -- Our timestamps
    first_synced_at TIMESTAMPTZ DEFAULT NOW(),
    last_synced_at TIMESTAMPTZ DEFAULT NOW(),
    sync_source VARCHAR(50)                  -- 'webhook', 'manual_sync', 'checkout'
    
    -- Primary key and unique constraint defined inline above with SERIAL PRIMARY KEY and UNIQUE NOT NULL
);

CREATE INDEX idx_stripe_customers_v2_email ON stripe_customers_v2(email);
CREATE INDEX idx_stripe_customers_v2_stripe_created ON stripe_customers_v2(stripe_created_at DESC);
```

**Changes from old**:
- ✅ Added `stripe_created_at` (from Stripe)
- ✅ Added `first_synced_at` / `last_synced_at` (audit trail)
- ✅ Added `sync_source` (know where data came from)
- ✅ Added `balance`, `currency`, `delinquent` (from Stripe API)

---

#### **Table 2: `stripe_products_v2`**
```sql
CREATE TABLE stripe_products_v2 (
    id SERIAL PRIMARY KEY,
    stripe_id VARCHAR(255) UNIQUE NOT NULL,  -- prod_xxxxx
    name VARCHAR(500) NOT NULL,
    description TEXT,
    active BOOLEAN DEFAULT true,
    metadata JSONB,
    
    -- Stripe timestamps
    stripe_created_at TIMESTAMPTZ NOT NULL,
    stripe_updated_at TIMESTAMPTZ,
    
    -- Our custom fields (NOT from Stripe)
    video_approved BOOLEAN DEFAULT false,    -- Custom: does this grant video access?
    is_legacy BOOLEAN DEFAULT false,         -- Custom: old/deprecated product
    
    -- Our timestamps
    first_synced_at TIMESTAMPTZ DEFAULT NOW(),
    last_synced_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_stripe_products_v2_active ON stripe_products_v2(active) WHERE active = true;
CREATE INDEX idx_stripe_products_v2_video_approved ON stripe_products_v2(video_approved) WHERE video_approved = true;
```

**Changes from old**:
- ✅ Removed `available` (unused?)
- ✅ Renamed `legacy_product` → `is_legacy` (clearer)
- ✅ Added `stripe_updated_at` (from Stripe)
- ✅ Added audit timestamps

---

#### **Table 3: `stripe_prices_v2`**
```sql
CREATE TABLE stripe_prices_v2 (
    id SERIAL PRIMARY KEY,
    stripe_id VARCHAR(255) UNIQUE NOT NULL,  -- price_xxxxx
    
    -- ✅ PROPER FOREIGN KEY TO PRODUCTS
    product_id INTEGER NOT NULL REFERENCES stripe_products_v2(id) ON DELETE CASCADE,
    
    unit_amount BIGINT NOT NULL,             -- cents (can be 0 for free)
    currency VARCHAR(3) NOT NULL DEFAULT 'usd',
    active BOOLEAN DEFAULT true,
    
    -- Recurring details
    recurring_interval VARCHAR(20),          -- 'month', 'year', 'week', 'day'
    recurring_interval_count INTEGER DEFAULT 1,
    
    metadata JSONB,
    
    -- Stripe timestamps
    stripe_created_at TIMESTAMPTZ NOT NULL,
    
    -- Our timestamps
    first_synced_at TIMESTAMPTZ DEFAULT NOW(),
    last_synced_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_stripe_prices_v2_product ON stripe_prices_v2(product_id);
CREATE INDEX idx_stripe_prices_v2_active ON stripe_prices_v2(active) WHERE active = true;
CREATE INDEX idx_stripe_prices_v2_recurring ON stripe_prices_v2(recurring_interval);
```

**Changes from old**:
- ✅ **PROPER FK** to `stripe_products_v2(id)` (integer, not string!)
- ✅ Added `recurring_interval_count`
- ✅ Removed redundant denormalized fields

---

#### **Table 4: `stripe_subscriptions_v2`**
```sql
CREATE TABLE stripe_subscriptions_v2 (
    id SERIAL PRIMARY KEY,
    stripe_id VARCHAR(255) UNIQUE NOT NULL,  -- sub_xxxxx
    
    -- ✅ PROPER FOREIGN KEYS
    customer_id INTEGER NOT NULL REFERENCES stripe_customers_v2(id) ON DELETE CASCADE,
    price_id INTEGER NOT NULL REFERENCES stripe_prices_v2(id) ON DELETE RESTRICT,
    
    status VARCHAR(50) NOT NULL,             -- 'active', 'canceled', 'past_due', etc.
    
    -- Billing periods
    current_period_start TIMESTAMPTZ NOT NULL,
    current_period_end TIMESTAMPTZ NOT NULL,
    cancel_at_period_end BOOLEAN DEFAULT false,
    canceled_at TIMESTAMPTZ,
    
    -- Stripe timestamps
    stripe_created_at TIMESTAMPTZ NOT NULL,
    
    -- Our timestamps
    first_synced_at TIMESTAMPTZ DEFAULT NOW(),
    last_synced_at TIMESTAMPTZ DEFAULT NOW(),
    
    metadata JSONB
);

CREATE INDEX idx_stripe_subscriptions_v2_customer ON stripe_subscriptions_v2(customer_id);
CREATE INDEX idx_stripe_subscriptions_v2_status ON stripe_subscriptions_v2(status);
CREATE INDEX idx_stripe_subscriptions_v2_active ON stripe_subscriptions_v2(customer_id, status) 
    WHERE status IN ('active', 'trialing');
CREATE INDEX idx_stripe_subscriptions_v2_period_end ON stripe_subscriptions_v2(current_period_end);
```

**Changes from old**:
- ✅ **PROPER FKS** with cascade rules
- ✅ Removed denormalized fields (`product_name`, `unit_amount`, etc.)
- ✅ Removed redundant `stripe_price_id` field
- ✅ Added `canceled_at`, `cancel_at_period_end`
- ✅ Added index for active subscriptions (most common query)

---

#### **Table 5: `user_stripe_customers_v2` (THE KEY INNOVATION)**
```sql
CREATE TABLE user_stripe_customers_v2 (
    id SERIAL PRIMARY KEY,
    
    -- ✅ EXPLICIT MANY-TO-MANY LINK
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stripe_customer_id INTEGER NOT NULL REFERENCES stripe_customers_v2(id) ON DELETE CASCADE,
    
    -- ✅ PRIMARY CUSTOMER TRACKING
    is_primary BOOLEAN DEFAULT false,
    
    -- ✅ AUDIT TRAIL
    linked_at TIMESTAMPTZ DEFAULT NOW(),
    linked_by VARCHAR(50) NOT NULL,          -- 'webhook', 'manual_sync', 'admin', 'checkout'
    linked_reason TEXT,                      -- 'new_subscription', 'email_match', 'manual_link', etc.
    
    -- Optional: why is this customer kept?
    notes TEXT,
    
    CONSTRAINT user_stripe_customers_v2_unique UNIQUE (user_id, stripe_customer_id)
);

-- ⚠️ CRITICAL: Ensure only one primary customer per user
CREATE UNIQUE INDEX idx_user_stripe_customers_v2_primary 
    ON user_stripe_customers_v2(user_id) 
    WHERE is_primary = true;

CREATE INDEX idx_user_stripe_customers_v2_user ON user_stripe_customers_v2(user_id);
CREATE INDEX idx_user_stripe_customers_v2_customer ON user_stripe_customers_v2(stripe_customer_id);
```

**This is the magic table!**
- ✅ No more array fields in `users` table
- ✅ Explicit many-to-many relationship
- ✅ Audit trail (when, how, why linked)
- ✅ **Unique index ensures only ONE primary customer per user**
- ✅ Fast lookups in both directions

---

### **NEW QUERY PATTERN (LIGHTNING FAST)**

**Find user's active subscription (OLD WAY - 4 tables, array scan)**:
```sql
SELECT ...
FROM users u
LEFT JOIN stripe_customers sc ON sc.stripe_id = ANY(u.stripe_customer_ids)  -- 😱 O(n*m)
LEFT JOIN stripe_subscriptions ss ON ss.customer_id = sc.id
LEFT JOIN stripe_prices sp ON sp.id = ss.price_id
LEFT JOIN stripe_products prod ON prod.stripe_id = sp.product_id            -- 😱 string join
WHERE u.id = 9797;
```

**Find user's active subscription (NEW WAY - 5 tables, integer FKs)**:
```sql
SELECT 
    u.id, u.email,
    sc.stripe_id as customer_stripe_id,
    ss.stripe_id as subscription_id,
    ss.status,
    ss.current_period_end,
    p.unit_amount,
    p.currency,
    p.recurring_interval,
    prod.name as product_name,
    prod.video_approved
FROM users u
INNER JOIN user_stripe_customers_v2 usc ON usc.user_id = u.id AND usc.is_primary = true  -- ✅ O(1)
INNER JOIN stripe_customers_v2 sc ON sc.id = usc.stripe_customer_id                      -- ✅ O(1)
INNER JOIN stripe_subscriptions_v2 ss ON ss.customer_id = sc.id                           -- ✅ O(1)
    AND ss.status IN ('active', 'trialing')
INNER JOIN stripe_prices_v2 p ON p.id = ss.price_id                                       -- ✅ O(1)
INNER JOIN stripe_products_v2 prod ON prod.id = p.product_id                              -- ✅ O(1)
WHERE u.id = 9797;
```

**Performance**: **10-100x faster** (all integer FK joins, no string matching, no array scans)

---

## 🏗️ **PART 5: MIGRATION STRATEGY**

### **Option 1: Build Beside, Then Swap (RECOMMENDED)**

#### **Phase 1: Create v2 Tables (No Data)**
```sql
-- Create all v2 tables (empty)
CREATE TABLE stripe_customers_v2 (...);
CREATE TABLE stripe_products_v2 (...);
CREATE TABLE stripe_prices_v2 (...);
CREATE TABLE stripe_subscriptions_v2 (...);
CREATE TABLE user_stripe_customers_v2 (...);
```

#### **Phase 2: Sync v2 from Stripe API (Fresh Data)**
```
1. Call Stripe API: GET /v1/products
2. Insert into stripe_products_v2
3. Call Stripe API: GET /v1/prices
4. Insert into stripe_prices_v2 (with FK to stripe_products_v2.id)
5. Call Stripe API: GET /v1/customers
6. Insert into stripe_customers_v2
7. Call Stripe API: GET /v1/subscriptions (for each customer)
8. Insert into stripe_subscriptions_v2 (with FKs to stripe_customers_v2.id and stripe_prices_v2.id)
```

#### **Phase 3: Link Users to Customers**
```sql
-- Link users to their Stripe customers by email
INSERT INTO user_stripe_customers_v2 (user_id, stripe_customer_id, is_primary, linked_by, linked_reason)
SELECT 
    u.id,
    sc.id,
    true,  -- Mark as primary if only one customer
    'migration',
    'email_match'
FROM users u
INNER JOIN stripe_customers_v2 sc ON sc.email = u.email
ON CONFLICT (user_id, stripe_customer_id) DO NOTHING;
```

#### **Phase 4: Update Elastic Service to Query v2**
```go
// backend/internal/services/subscriber_elastic_service_v2.go

func (s *SubscriberElasticService) GetAllUnifiedSubscribersV2() ([]*UnifiedSubscriber, error) {
    query := `
        WITH user_subscriptions AS (
            SELECT DISTINCT ON (u.id)
                u.id,
                u.email,
                sc.stripe_id as customer_stripe_id,
                ss.stripe_id as subscription_id,
                ss.status,
                ss.current_period_start,
                ss.current_period_end,
                p.unit_amount,
                p.currency,
                p.recurring_interval,
                prod.name as product_name,
                prod.video_approved
            FROM users u
            LEFT JOIN user_stripe_customers_v2 usc ON usc.user_id = u.id AND usc.is_primary = true
            LEFT JOIN stripe_customers_v2 sc ON sc.id = usc.stripe_customer_id
            LEFT JOIN stripe_subscriptions_v2 ss ON ss.customer_id = sc.id 
                AND ss.status IN ('active', 'trialing')
            LEFT JOIN stripe_prices_v2 p ON p.id = ss.price_id
            LEFT JOIN stripe_products_v2 prod ON prod.id = p.product_id
            ORDER BY u.id, ss.stripe_created_at DESC
        )
        SELECT * FROM user_subscriptions;
    `
    // ... rest of query ...
}
```

#### **Phase 5: Test v2 System**
```
1. Frontend: Load subscribers from v2
2. Verify data matches Stripe dashboard
3. Test webhook handlers (create test subscription)
4. Verify user linking works
5. Performance test (should be 10-100x faster)
```

#### **Phase 6: Deploy Webhook Handlers for v2**
```go
func (s *StripeService) handleSubscriptionCreated(event *stripe.Event) error {
    // 1. Extract subscription from event
    var sub stripe.Subscription
    json.Unmarshal(event.Data.Raw, &sub)
    
    // 2. Fetch or create customer in v2
    customer, err := s.getOrCreateCustomerV2(sub.Customer.ID)
    
    // 3. Fetch or create price in v2
    price, err := s.getOrCreatePriceV2(sub.Items.Data[0].Price.ID)
    
    // 4. Create subscription in v2
    _, err = s.createSubscriptionV2(&sub, customer.ID, price.ID)
    
    // 5. Link customer to user (by email)
    err = s.linkCustomerToUserByEmail(customer.Email, customer.ID)
    
    return nil
}
```

#### **Phase 7: Run Both Systems in Parallel (1 week)**
```
- v1 tables (old) still exist
- v2 tables (new) receiving webhooks
- Frontend queries both, logs differences
- If differences found, investigate
```

#### **Phase 8: Cut Over to v2**
```
1. Update all code to use v2 tables only
2. Remove references to v1 tables
3. Monitor for 48 hours
4. If stable, proceed to Phase 9
```

#### **Phase 9: Archive v1 Tables**
```sql
-- Rename old tables with _deprecated suffix
ALTER TABLE stripe_customers RENAME TO stripe_customers_deprecated;
ALTER TABLE stripe_subscriptions RENAME TO stripe_subscriptions_deprecated;
ALTER TABLE stripe_products RENAME TO stripe_products_deprecated;
ALTER TABLE stripe_prices RENAME TO stripe_prices_deprecated;

-- Keep for 30 days, then drop
-- (set a calendar reminder)
```

#### **Phase 10: Rename v2 → Production**
```sql
-- Once v1 is archived and v2 is proven stable
ALTER TABLE stripe_customers_v2 RENAME TO stripe_customers;
ALTER TABLE stripe_subscriptions_v2 RENAME TO stripe_subscriptions;
ALTER TABLE stripe_products_v2 RENAME TO stripe_products;
ALTER TABLE stripe_prices_v2 RENAME TO stripe_prices;
ALTER TABLE user_stripe_customers_v2 RENAME TO user_stripe_customers;

-- Update all indexes and constraints
```

---

## 📊 **PART 6: IMPLEMENTATION PLAN**

### **Time Estimates**

| Phase | Task | Time |
|-------|------|------|
| 1 | Create v2 schema SQL files | 1 hour |
| 2 | Full Stripe API sync script | 2 hours |
| 3 | User linking logic | 1 hour |
| 4 | Update elastic service for v2 | 2 hours |
| 5 | Test v2 queries | 1 hour |
| 6 | Implement v2 webhook handlers | 3 hours |
| 7 | Parallel testing (monitoring) | 1 week |
| 8 | Cut over to v2 | 1 hour |
| **TOTAL** | **~11 hours + 1 week monitoring** |

---

## 🎯 **PART 7: BENEFITS OF NEW SYSTEM**

### **Performance**
- ✅ **10-100x faster queries** (integer FK joins vs array scans + string matches)
- ✅ **Better index usage** (all FKs indexed, unique indexes on primary customers)
- ✅ **Optimized for most common queries** (user's active subscription = single indexed lookup)

### **Data Integrity**
- ✅ **Foreign key constraints** prevent orphaned records
- ✅ **Unique index** ensures only one primary customer per user
- ✅ **Cascade rules** maintain referential integrity
- ✅ **Audit trail** tracks all data changes

### **Maintainability**
- ✅ **Clear separation** of Stripe data vs custom fields
- ✅ **Normalized** - no duplicate data
- ✅ **Self-documenting** - FK relationships explicit
- ✅ **Easy to extend** - add new payment processors by creating similar tables

### **Debugging**
- ✅ **Audit timestamps** show when data was synced
- ✅ **Sync source tracking** shows where data came from (webhook vs manual)
- ✅ **Link reasons** explain why customers are linked to users
- ✅ **Easy to query** - all relationships explicit

---

## 🚀 **NEXT STEPS**

### **Questions for You**

1. **Approve v2 schema?** Any fields you want to add/remove/change?

2. **Migration strategy?** 
   - Option A: Beta reset (clear v1, populate v2 from Stripe)
   - Option B: Build beside (keep v1, build v2, run parallel, cut over)

3. **Timeline?** When do you want this done?
   - Urgent (this weekend)?
   - Normal (this week)?
   - Careful (2 weeks with thorough testing)?

4. **Multiple active subscriptions?** What's the business rule?
   - Should we auto-cancel old subscriptions when new one created?
   - Or allow multiple (like James Kersey with 3 active subs)?

5. **Start implementation now?** I can begin creating:
   - `migrations/050_create_stripe_v2_schema.sql`
   - `backend/internal/services/stripe_sync_v2.go`
   - `backend/internal/services/customer_linking_service.go`

---

**🔥 I'm ready to build this! What do you say?**

