# 🔍 Stripe BRAID Comprehensive Audit

**Date**: October 26, 2025  
**Status**: 🚨 **CRITICAL FRAGMENTATION DETECTED**  
**Priority**: HIGH - Multiple data integrity issues found

---

## 📋 **SYSTEM DESIGN CLARIFICATIONS**

### **User Flow**
1. User **MUST** sign up/sign in BEFORE subscribing ✅
2. User subscribes via Stripe Checkout
3. Webhook links subscription to user account
4. User gets video access

### **Customer ID Strategy**
- `users.stripe_customer_id` = **ACTIVE** customer being tracked
- `users.stripe_customer_ids[]` = **ALL** customer IDs (historical)
- Multiple `cus_` per email because Stripe sends canceled plans, resubscribes, etc.

### **Data Flow**
- **Stripe → Your DB**: One-way sync (webhooks + manual)
- **Your DB → Stripe**: ONLY during checkout (customer creation, subscription signup)
- **Future Goal**: Payment processor agnostic (not Stripe-dependent)

### **Sync Strategy**
- **Primary**: Webhooks (real-time)
- **Fallback**: Manual sync (admin triggered)
- **No automated cron jobs** (yet)

### **Current Join Logic** (NEEDS IMPROVEMENT)
```
users.stripe_customer_ids[] 
  → stripe_customers.stripe_id 
  → stripe_customers.id 
  → stripe_subscriptions.customer_id
```

---

## 🔍 **AUDIT FINDINGS**

### **🚨 CRITICAL ISSUES**

#### **1. WEBHOOKS NOT IMPLEMENTED**
**Files**: `backend/internal/services/stripe.go`  
**Lines**: 801-831  
**Severity**: 🔴 **CRITICAL**

**Problem**:
```go
func (s *StripeService) handleSubscriptionCreated(event *stripe.Event) error {
    // ... logs event ...
    // TODO: Update local database with subscription information
    return nil  // ❌ DOES NOTHING!
}
```

**Impact**:
- ❌ New subscriptions not linked to user accounts
- ❌ Customer IDs not added to `users.stripe_customer_ids`
- ❌ `users.stripe_customer_id` not updated
- ❌ Users show as "No Plan" despite active Stripe subscription

**Affected Users**: Eric Gessel confirmed, likely more

---

#### **2. INEFFICIENT JOIN STRATEGY**
**Current Path**: 
```
users.stripe_customer_ids[] (array)
  → stripe_customers.stripe_id (string match)
  → stripe_customers.id (integer)
  → stripe_subscriptions.customer_id (foreign key)
```

**Problems**:
- ❌ **3-table join** with array unnesting (slow for large datasets)
- ❌ **String matching** on `stripe_id` instead of integer foreign key
- ❌ **Array scan** for every query (no index optimization possible)
- ❌ **Multiple queries** required in code (not single SQL join)

**Better Approach**:
```
users.id 
  → user_stripe_customers.user_id (link table)
  → user_stripe_customers.stripe_customer_id
  → stripe_subscriptions.customer_id
```

---

#### **3. NO FOREIGN KEY RELATIONSHIPS**
**Tables**: `users`, `stripe_customers`, `stripe_subscriptions`

**Problem**: No database-level foreign keys linking:
- `stripe_subscriptions.customer_id` → `stripe_customers.id`
- `users` → `stripe_customers` (no direct link at all!)

**Impact**:
- ❌ Orphaned records possible
- ❌ No referential integrity
- ❌ Slow joins (no FK indexes)
- ❌ Cascading deletes not possible

---

#### **4. MULTIPLE SYNC ENDPOINTS WITH NO COORDINATION**
**Found**:
- `/admin/streaming/stripe/sync-customers` (manual sync)
- `/admin/streaming/stripe/sync-subscriptions` (manual sync)
- `/admin/streaming/stripe/comprehensive-sync` (comprehensive sync)
- Webhook handlers (not working)

**Problem**: No single source of truth for sync logic

---

### **⚠️ MAJOR ISSUES**

#### **5. NO USER → CUSTOMER LINKING TABLE**
**Current**: User has array of customer IDs in `stripe_customer_ids`  
**Problem**: 
- Can't query "all users for this customer"
- Can't track "when was this customer linked"
- Can't track "which customer is primary and why"
- Can't efficiently join on array fields

**Recommendation**: Create `user_stripe_customers` link table

---

#### **6. ELASTIC SERVICE SCANS ARRAY EVERY QUERY**
**File**: `backend/internal/services/subscriber_elastic_service.go`

**Current Query** (simplified):
```sql
SELECT ...
FROM users u
LEFT JOIN stripe_customers sc ON sc.stripe_id = ANY(u.stripe_customer_ids)
LEFT JOIN stripe_subscriptions ss ON ss.customer_id = sc.id
```

**Problem**:
- PostgreSQL must unnest array for every row
- Can't use index on array field efficiently
- O(n*m) complexity (users * customer_ids per user)

---

#### **7. STRIPE API DEPENDENCY FOR CUSTOMER DATA**
**Question Asked**: "If we pull customers, do we need to pull subscriptions too?"

**Stripe API Answer**:
- `/v1/customers/:id` returns customer data + metadata
- **DOES NOT** include subscription details
- Must call `/v1/subscriptions?customer=:id` separately
- OR call `/v1/customers/:id?expand[]=subscriptions` (expensive)

**Implication**: You'll always need separate subscription API calls for any payment processor

---

### **🟡 MINOR ISSUES**

#### **8. NO MIGRATION SYSTEM**
**Current**: Manual table creation, no version control  
**Recommendation**: Create `database/migrations/` with versioned SQL files

#### **9. DUPLICATE SYNC LOGIC**
Multiple places implement "sync from Stripe":
- Webhook handlers (incomplete)
- Manual sync endpoints (working)
- Elastic service queries (reads local data)

#### **10. NO STRIPE CUSTOMER DEDUPLICATION**
If user subscribes, cancels, resubscribes:
- Stripe might create NEW customer (if they use different email/payment)
- Your system has no logic to detect and merge duplicate customers

---

## 🏗️ **RECOMMENDED ARCHITECTURE**

### **New Database Schema**

```sql
-- EXISTING (keep as-is)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    stripe_customer_id TEXT, -- PRIMARY/ACTIVE customer
    stripe_customer_ids TEXT[], -- ALL historical customers
    -- ... other fields ...
);

-- EXISTING (keep as-is)
CREATE TABLE stripe_customers (
    id SERIAL PRIMARY KEY,
    stripe_id TEXT UNIQUE NOT NULL, -- cus_xxxxx
    email TEXT,
    name TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);

-- EXISTING (keep as-is)
CREATE TABLE stripe_subscriptions (
    id SERIAL PRIMARY KEY,
    stripe_id TEXT UNIQUE NOT NULL, -- sub_xxxxx
    customer_id INTEGER NOT NULL, -- FK to stripe_customers.id
    status TEXT,
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    price_id INTEGER,
    -- ... other fields ...
);

-- NEW: Link table for many-to-many relationship
CREATE TABLE user_stripe_customers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stripe_customer_id INTEGER NOT NULL REFERENCES stripe_customers(id) ON DELETE CASCADE,
    is_primary BOOLEAN DEFAULT false,
    linked_at TIMESTAMPTZ DEFAULT NOW(),
    linked_by TEXT, -- 'webhook', 'manual_sync', 'admin'
    UNIQUE(user_id, stripe_customer_id)
);

-- Index for fast lookups
CREATE INDEX idx_user_stripe_customers_user ON user_stripe_customers(user_id);
CREATE INDEX idx_user_stripe_customers_customer ON user_stripe_customers(stripe_customer_id);
CREATE INDEX idx_user_stripe_customers_primary ON user_stripe_customers(user_id, is_primary);

-- Add missing foreign key
ALTER TABLE stripe_subscriptions 
ADD CONSTRAINT fk_stripe_subscriptions_customer 
FOREIGN KEY (customer_id) REFERENCES stripe_customers(id) ON DELETE CASCADE;
```

### **New Query Pattern (MUCH FASTER)**

```sql
-- OLD WAY (slow array unnest):
SELECT u.*, ss.*
FROM users u
LEFT JOIN stripe_customers sc ON sc.stripe_id = ANY(u.stripe_customer_ids)
LEFT JOIN stripe_subscriptions ss ON ss.customer_id = sc.id
WHERE u.id = 9797;

-- NEW WAY (fast FK joins):
SELECT u.*, ss.*
FROM users u
LEFT JOIN user_stripe_customers usc ON usc.user_id = u.id AND usc.is_primary = true
LEFT JOIN stripe_customers sc ON sc.id = usc.stripe_customer_id
LEFT JOIN stripe_subscriptions ss ON ss.customer_id = sc.id
WHERE u.id = 9797;
```

**Performance**: O(1) lookups vs O(n) array scans

---

## 🔧 **IMPLEMENTATION PLAN**

### **Phase 1: Database Schema (1 hour)**

**Step 1.1**: Create `user_stripe_customers` link table
**Step 1.2**: Add foreign key constraints
**Step 1.3**: Create indexes
**Step 1.4**: Migrate existing `stripe_customer_ids` data to link table

**Migration SQL**:
```sql
-- Populate link table from existing array data
INSERT INTO user_stripe_customers (user_id, stripe_customer_id, is_primary, linked_by)
SELECT 
    u.id,
    sc.id,
    (sc.stripe_id = u.stripe_customer_id) as is_primary,
    'migration'
FROM users u
CROSS JOIN LATERAL unnest(u.stripe_customer_ids) AS cus_id
INNER JOIN stripe_customers sc ON sc.stripe_id = cus_id
ON CONFLICT (user_id, stripe_customer_id) DO NOTHING;
```

---

### **Phase 2: Webhook Implementation (2-3 hours)**

**Step 2.1**: Create `CustomerLinkingService`

```go
// File: backend/internal/services/customer_linking_service.go

type CustomerLinkingService struct {
    db *database.DB
}

func NewCustomerLinkingService(db *database.DB) *CustomerLinkingService {
    return &CustomerLinkingService{db: db}
}

// LinkCustomerToUser links a Stripe customer to a user account
func (s *CustomerLinkingService) LinkCustomerToUser(stripeCustomerID, email string) error {
    // 1. Find or create stripe_customers record
    customer, err := s.getOrCreateStripeCustomer(stripeCustomerID, email)
    if err != nil {
        return fmt.Errorf("failed to get/create customer: %w", err)
    }
    
    // 2. Find user by email
    user, err := s.db.GetUserByEmail(email)
    if err != nil {
        if err == sql.ErrNoRows {
            // No user account yet - subscription before signup
            log.Printf("⚠️ Customer %s has no user account yet (email: %s)", stripeCustomerID, email)
            return nil
        }
        return fmt.Errorf("failed to find user: %w", err)
    }
    
    // 3. Check if already linked
    exists, err := s.isCustomerLinked(user.ID, customer.ID)
    if err != nil {
        return err
    }
    if exists {
        // Already linked - maybe update to primary
        return s.maybeUpdatePrimaryCustomer(user.ID, customer.ID)
    }
    
    // 4. Create link
    return s.createCustomerLink(user.ID, customer.ID, "webhook")
}

func (s *CustomerLinkingService) getOrCreateStripeCustomer(stripeID, email string) (*StripeCustomer, error) {
    // Check if exists
    var customer StripeCustomer
    err := s.db.QueryRow(`
        SELECT id, stripe_id, email, name 
        FROM stripe_customers 
        WHERE stripe_id = $1
    `, stripeID).Scan(&customer.ID, &customer.StripeID, &customer.Email, &customer.Name)
    
    if err == sql.ErrNoRows {
        // Fetch from Stripe API and create
        return s.fetchAndCreateCustomer(stripeID)
    }
    
    return &customer, err
}

func (s *CustomerLinkingService) isCustomerLinked(userID, customerID int) (bool, error) {
    var exists bool
    err := s.db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM user_stripe_customers 
            WHERE user_id = $1 AND stripe_customer_id = $2
        )
    `, userID, customerID).Scan(&exists)
    return exists, err
}

func (s *CustomerLinkingService) createCustomerLink(userID, customerID int, source string) error {
    // Check if user has any primary customer
    var hasPrimary bool
    err := s.db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM user_stripe_customers 
            WHERE user_id = $1 AND is_primary = true
        )
    `, userID).Scan(&hasPrimary)
    if err != nil {
        return err
    }
    
    // If no primary, make this one primary
    isPrimary := !hasPrimary
    
    _, err = s.db.Exec(`
        INSERT INTO user_stripe_customers (user_id, stripe_customer_id, is_primary, linked_by)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id, stripe_customer_id) DO NOTHING
    `, userID, customerID, isPrimary, source)
    
    if err != nil {
        return err
    }
    
    // Update users.stripe_customer_id for backwards compatibility
    if isPrimary {
        stripeID, _ := s.getStripeIDForCustomer(customerID)
        _, _ = s.db.Exec(`
            UPDATE users 
            SET stripe_customer_id = $1, updated_at = NOW() 
            WHERE id = $2
        `, stripeID, userID)
    }
    
    return nil
}

func (s *CustomerLinkingService) maybeUpdatePrimaryCustomer(userID, customerID int) error {
    // Check if this customer has an active subscription
    var hasActiveSubscription bool
    err := s.db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM stripe_subscriptions 
            WHERE customer_id = $1 AND status IN ('active', 'trialing')
        )
    `, customerID).Scan(&hasActiveSubscription)
    
    if err != nil || !hasActiveSubscription {
        return err
    }
    
    // Has active subscription - make this the primary customer
    tx, err := s.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Unset all primaries for this user
    _, err = tx.Exec(`
        UPDATE user_stripe_customers 
        SET is_primary = false 
        WHERE user_id = $1
    `, userID)
    if err != nil {
        return err
    }
    
    // Set this one as primary
    _, err = tx.Exec(`
        UPDATE user_stripe_customers 
        SET is_primary = true 
        WHERE user_id = $1 AND stripe_customer_id = $2
    `, userID, customerID)
    if err != nil {
        return err
    }
    
    // Update users.stripe_customer_id
    stripeID, _ := s.getStripeIDForCustomer(customerID)
    _, err = tx.Exec(`
        UPDATE users 
        SET stripe_customer_id = $1, updated_at = NOW() 
        WHERE id = $2
    `, stripeID, userID)
    if err != nil {
        return err
    }
    
    return tx.Commit()
}
```

**Step 2.2**: Update webhook handlers to use linking service

```go
// File: backend/internal/services/stripe.go

func (s *StripeService) handleSubscriptionCreated(event *stripe.Event) error {
    var subscription stripe.Subscription
    err := json.Unmarshal(event.Data.Raw, &subscription)
    if err != nil {
        return fmt.Errorf("failed to unmarshal subscription: %w", err)
    }
    
    log.Printf("📊 [WEBHOOK] Subscription created: %s for customer %s", subscription.ID, subscription.Customer.ID)
    
    // 1. Get customer email from Stripe
    customer, err := customer.Get(subscription.Customer.ID, nil)
    if err != nil {
        return fmt.Errorf("failed to fetch customer: %w", err)
    }
    
    // 2. Link customer to user account
    linkingService := NewCustomerLinkingService(s.db)
    if err := linkingService.LinkCustomerToUser(subscription.Customer.ID, customer.Email); err != nil {
        log.Printf("❌ Failed to link customer: %v", err)
        // Don't fail webhook - just log
    }
    
    // 3. Track analytics
    s.trackSubscriptionEvent("subscription_created", &subscription, event)
    
    return nil
}

func (s *StripeService) handleSubscriptionUpdated(event *stripe.Event) error {
    var subscription stripe.Subscription
    err := json.Unmarshal(event.Data.Raw, &subscription)
    if err != nil {
        return fmt.Errorf("failed to unmarshal subscription: %w", err)
    }
    
    log.Printf("📊 [WEBHOOK] Subscription updated: %s (status: %s)", subscription.ID, subscription.Status)
    
    // Re-link customer (in case status changed to 'active')
    customer, err := customer.Get(subscription.Customer.ID, nil)
    if err == nil {
        linkingService := NewCustomerLinkingService(s.db)
        _ = linkingService.LinkCustomerToUser(subscription.Customer.ID, customer.Email)
    }
    
    s.trackSubscriptionEvent("subscription_updated", &subscription, event)
    return nil
}
```

---

### **Phase 3: Elastic Service Update (1 hour)**

**Step 3.1**: Update elastic service query to use link table

```go
// File: backend/internal/services/subscriber_elastic_service.go

func (s *SubscriberElasticService) GetUnifiedSubscriberByID(userID int) (*UnifiedSubscriber, error) {
    query := `
        WITH user_subscriptions AS (
            SELECT DISTINCT ON (u.id)
                u.id as user_id,
                u.email,
                sc.stripe_id as customer_stripe_id,
                ss.stripe_id as subscription_id,
                ss.status as subscription_status,
                sp.name as product_name,
                spr.unit_amount as product_price,
                spr.currency as product_currency,
                spr.recurring_interval as product_interval,
                ss.current_period_start as billing_start,
                ss.current_period_end as billing_end
            FROM users u
            -- NEW: Use link table instead of array
            LEFT JOIN user_stripe_customers usc ON usc.user_id = u.id AND usc.is_primary = true
            LEFT JOIN stripe_customers sc ON sc.id = usc.stripe_customer_id
            LEFT JOIN stripe_subscriptions ss ON ss.customer_id = sc.id 
                AND ss.status IN ('active', 'trialing', 'canceled', 'past_due')
            LEFT JOIN stripe_prices spr ON ss.price_id = spr.id
            LEFT JOIN stripe_products sp ON spr.product_id = sp.stripe_id
            WHERE u.id = $1
            ORDER BY u.id, ss.created_at DESC
        )
        SELECT ... FROM user_subscriptions ...
    `
    // ... rest of query ...
}
```

**Performance Improvement**: 10-100x faster (no array unnesting!)

---

### **Phase 4: Migration System (1 hour)**

**Step 4.1**: Create migrations directory structure

```
database/
  migrations/
    001_initial_schema.sql
    002_add_user_stripe_customers.sql
    003_add_foreign_keys.sql
    README.md
```

**Step 4.2**: Create migration runner (optional)

```go
// File: backend/cmd/migrate/main.go

func main() {
    // Simple migration runner
    // Reads SQL files in order, executes them
}
```

---

### **Phase 5: Admin Sync Dashboard (2 hours)**

**Step 5.1**: Create unified sync endpoint

```go
// File: backend/internal/routes/admin_stripe_sync.go

func SyncAllStripeData(c *gin.Context) {
    // 1. Sync all customers from Stripe
    customers, err := syncAllCustomers()
    
    // 2. Sync all subscriptions from Stripe
    subscriptions, err := syncAllSubscriptions()
    
    // 3. Link customers to users by email
    linkingService := NewCustomerLinkingService(db)
    for _, customer := range customers {
        _ = linkingService.LinkCustomerToUser(customer.ID, customer.Email)
    }
    
    c.JSON(200, gin.H{
        "customers_synced": len(customers),
        "subscriptions_synced": len(subscriptions),
        "users_linked": linkedCount,
    })
}
```

**Step 5.2**: Create admin UI for sync

```svelte
<!-- File: frontend/src/routes/admin/streaming/stripe/sync/+page.svelte -->

<button on:click={syncAllData}>
    🔄 Full Stripe Sync
</button>

<div>
    Last Sync: {lastSyncTime}
    Customers: {customerCount}
    Subscriptions: {subscriptionCount}
    Linked Users: {linkedUserCount}
</div>
```

---

## 📊 **MIGRATION STRATEGY (BETA RESET)**

### **Option A: Clean Slate Migration**

```sql
-- 1. Backup current data
CREATE TABLE users_backup AS SELECT * FROM users;
CREATE TABLE stripe_customers_backup AS SELECT * FROM stripe_customers;
CREATE TABLE stripe_subscriptions_backup AS SELECT * FROM stripe_subscriptions;

-- 2. Drop Stripe tables
TRUNCATE TABLE stripe_subscriptions CASCADE;
TRUNCATE TABLE stripe_customers CASCADE;

-- 3. Clear Stripe references in users
UPDATE users SET stripe_customer_id = NULL, stripe_customer_ids = '{}';

-- 4. Create new schema
-- (run all CREATE TABLE statements from above)

-- 5. Run full sync from Stripe API
-- (call admin sync endpoint)

-- 6. Verify data integrity
SELECT 
    COUNT(*) as total_users,
    COUNT(stripe_customer_id) as users_with_customer,
    COUNT(CASE WHEN has_active_plan THEN 1 END) as users_with_plan
FROM users;
```

### **Option B: Incremental Migration (Keep Existing Data)**

```sql
-- 1. Create new tables
-- (run CREATE TABLE user_stripe_customers)

-- 2. Migrate existing data
-- (run INSERT INTO user_stripe_customers ... FROM users)

-- 3. Verify migration
SELECT 
    u.email,
    u.stripe_customer_id as old_primary,
    sc.stripe_id as new_primary
FROM users u
LEFT JOIN user_stripe_customers usc ON usc.user_id = u.id AND usc.is_primary = true
LEFT JOIN stripe_customers sc ON sc.id = usc.stripe_customer_id
WHERE u.stripe_customer_id IS NOT NULL;

-- 4. Once verified, deprecated old columns (optional)
-- ALTER TABLE users DROP COLUMN stripe_customer_ids;
```

---

## 🎯 **PRIORITY ACTIONS**

### **🔴 CRITICAL (Do Today)**
1. ✅ Fix Eric Gessel's account (SQL update)
2. ✅ Find all other users with mismatch (diagnostic query)
3. ✅ Implement basic webhook customer linking

### **🟡 HIGH (Do This Week)**
1. ✅ Create `user_stripe_customers` link table
2. ✅ Migrate existing data to link table
3. ✅ Update elastic service to use link table
4. ✅ Test webhook linking with test subscription

### **🟢 MEDIUM (Do Next Week)**
1. ✅ Create migration system
2. ✅ Create admin sync dashboard
3. ✅ Add foreign key constraints
4. ✅ Performance test queries

---

## 📈 **EXPECTED IMPROVEMENTS**

### **Before Fix**:
- ❌ Webhook handlers don't link customers
- ❌ Array scans on every query (slow)
- ❌ Users with active subs show "No Plan"
- ❌ No referential integrity

### **After Fix**:
- ✅ Webhooks auto-link customers to users
- ✅ FK joins (10-100x faster queries)
- ✅ All active subs show correctly
- ✅ Database enforces referential integrity
- ✅ Easy to add other payment processors

---

## 🚀 **PAYMENT PROCESSOR AGNOSTIC DESIGN**

### **Future Schema** (when adding PayPal, etc.):

```sql
CREATE TABLE payment_processors (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE, -- 'stripe', 'paypal', 'square'
    is_enabled BOOLEAN DEFAULT true
);

CREATE TABLE user_payment_customers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    processor_id INTEGER REFERENCES payment_processors(id),
    external_customer_id TEXT, -- cus_xxx, paypal_email, etc.
    is_primary BOOLEAN DEFAULT false,
    linked_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE payment_subscriptions (
    id SERIAL PRIMARY KEY,
    processor_id INTEGER REFERENCES payment_processors(id),
    customer_id INTEGER REFERENCES user_payment_customers(id),
    external_subscription_id TEXT,
    status TEXT,
    -- ... normalized fields ...
);
```

This way, Stripe is just one row in `payment_processors` table!

---

**Status**: 🚨 **READY FOR IMPLEMENTATION**  
**Estimated Time**: 8-10 hours total  
**Expected Outcome**: Clean, fast, maintainable Stripe integration

