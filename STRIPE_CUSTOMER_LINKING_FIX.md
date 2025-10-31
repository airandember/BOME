# 🔗 Stripe Customer Linking - Critical Fix

**Issue**: Users creating new subscriptions get NEW Stripe customer IDs that aren't linked to their existing user accounts.

**Impact**: 
- User shows as "No Plan" despite active Stripe subscription
- Multiple customer IDs per user email
- Data fragmentation between `users.stripe_customer_id` and actual Stripe data

---

## 🚨 **AFFECTED USER: Eric Gessel**

### **Current State (BROKEN)**:
- **User ID**: 9797
- **Email**: ericgessel@gmail.com
- **Primary Customer ID** (in `users` table): `cus_HJsNLfuaMqxZ5m` (OLD, canceled subscription)
- **Actual Customer ID** (in Stripe): `cus_TGAcxsB1BicDbY` (NEW, active subscription)
- **Result**: Shows as "No Plan" despite $95.64/year active subscription

### **Immediate Fix (SQL)**:
```sql
-- Fix Eric Gessel's account
UPDATE users 
SET stripe_customer_id = 'cus_TGAcxsB1BicDbY',
    stripe_customer_ids = ARRAY['cus_HJsNLfuaMqxZ5m', 'cus_TGAcxsB1BicDbY']::text[],
    updated_at = NOW()
WHERE id = 9797;
```

---

## 🔍 **ROOT CAUSE**

### **Webhook Handlers Not Implemented**

File: `backend/internal/services/stripe.go`  
Lines: 801-831

```go
// handleSubscriptionCreated handles subscription created events
func (s *StripeService) handleSubscriptionCreated(event *stripe.Event) error {
    var subscription stripe.Subscription
    err := json.Unmarshal(event.Data.Raw, &subscription)
    if err != nil {
        return fmt.Errorf("failed to unmarshal subscription: %w", err)
    }

    log.Printf("📊 [ANALYTICS] Subscription created: %s", subscription.ID)
    s.trackSubscriptionEvent("subscription_created", &subscription, event)

    // ❌ TODO: Update local database with subscription information
    return nil  // ← DOES NOTHING!
}
```

**What SHOULD Happen**:
1. Get subscription's customer ID from webhook
2. Get customer email from Stripe
3. Find user by email in database
4. Link new customer ID to user account using `AddStripeCustomerID()`

---

## ✅ **SOLUTION**

### **Step 1: Implement Webhook Customer Linking**

Add this helper function to `stripe.go`:

```go
// linkCustomerToUser links a Stripe customer ID to an existing user account
func (s *StripeService) linkCustomerToUser(customerID string) error {
    if !s.isEnabled {
        return fmt.Errorf("stripe service is disabled")
    }

    // Get customer from Stripe
    customer, err := customer.Get(customerID, nil)
    if err != nil {
        return fmt.Errorf("failed to fetch customer from Stripe: %w", err)
    }

    if customer.Email == "" {
        return fmt.Errorf("customer has no email")
    }

    // Find user by email
    user, err := s.db.GetUserByEmail(customer.Email)
    if err != nil {
        if err == sql.ErrNoRows {
            log.Printf("⚠️ No user found for email %s - customer %s not linked", customer.Email, customerID)
            return nil // Not an error - user might not exist yet
        }
        return fmt.Errorf("failed to find user: %w", err)
    }

    // Check if customer ID is already linked
    if user.StripeCustomerID == customerID {
        log.Printf("✅ Customer %s already linked to user %d (%s)", customerID, user.ID, user.Email)
        return nil
    }

    // Check if ID exists in array
    for _, existingID := range user.StripeCustomerIDs {
        if existingID == customerID {
            // ID in array but not primary - update primary
            log.Printf("🔄 Setting customer %s as primary for user %d (%s)", customerID, user.ID, user.Email)
            _, err = s.db.Exec(`
                UPDATE users 
                SET stripe_customer_id = $1, updated_at = NOW() 
                WHERE id = $2
            `, customerID, user.ID)
            return err
        }
    }

    // Add new customer ID
    log.Printf("➕ Adding NEW customer %s to user %d (%s)", customerID, user.ID, user.Email)
    return s.db.AddStripeCustomerID(user.ID, customerID)
}
```

### **Step 2: Update Subscription Created Handler**

```go
func (s *StripeService) handleSubscriptionCreated(event *stripe.Event) error {
    var subscription stripe.Subscription
    err := json.Unmarshal(event.Data.Raw, &subscription)
    if err != nil {
        return fmt.Errorf("failed to unmarshal subscription: %w", err)
    }

    log.Printf("📊 [WEBHOOK] Subscription created: %s for customer %s", subscription.ID, subscription.Customer.ID)

    // Track subscription creation analytics
    s.trackSubscriptionEvent("subscription_created", &subscription, event)

    // ✅ Link customer to user account
    if err := s.linkCustomerToUser(subscription.Customer.ID); err != nil {
        log.Printf("❌ Failed to link customer %s: %v", subscription.Customer.ID, err)
        // Don't fail the webhook - just log the error
    }

    return nil
}
```

### **Step 3: Update Subscription Updated Handler**

```go
func (s *StripeService) handleSubscriptionUpdated(event *stripe.Event) error {
    var subscription stripe.Subscription
    err := json.Unmarshal(event.Data.Raw, &subscription)
    if err != nil {
        return fmt.Errorf("failed to unmarshal subscription: %w", err)
    }

    log.Printf("📊 [WEBHOOK] Subscription updated: %s (status: %s)", subscription.ID, subscription.Status)

    // Track subscription update analytics
    s.trackSubscriptionEvent("subscription_updated", &subscription, event)

    // ✅ Ensure customer is still linked (in case it changed)
    if err := s.linkCustomerToUser(subscription.Customer.ID); err != nil {
        log.Printf("❌ Failed to link customer %s: %v", subscription.Customer.ID, err)
    }

    return nil
}
```

---

## 🔍 **FIND ALL AFFECTED USERS**

### **Query to Identify Users with Multiple Customer IDs**

```sql
-- Find users where their primary customer ID doesn't match their active subscription
WITH user_customers AS (
    SELECT 
        u.id as user_id,
        u.email,
        u.stripe_customer_id as primary_customer_id,
        sc.stripe_id as actual_customer_id,
        sc.id as customer_table_id
    FROM users u
    INNER JOIN stripe_customers sc ON sc.email = u.email
    WHERE u.stripe_customer_id != sc.stripe_id
),
active_subscriptions AS (
    SELECT 
        ss.customer_id,
        ss.stripe_id as subscription_id,
        ss.status,
        ss.current_period_end
    FROM stripe_subscriptions ss
    WHERE ss.status IN ('active', 'trialing')
)
SELECT 
    uc.user_id,
    uc.email,
    uc.primary_customer_id as old_customer_id,
    uc.actual_customer_id as new_customer_id,
    asub.subscription_id,
    asub.status,
    asub.current_period_end
FROM user_customers uc
INNER JOIN active_subscriptions asub ON asub.customer_id = uc.customer_table_id
ORDER BY uc.user_id;
```

### **Fix All Affected Users**

```sql
-- Update all users to use their MOST RECENT customer ID with active subscription
UPDATE users u
SET stripe_customer_id = sc.stripe_id,
    stripe_customer_ids = array_append(COALESCE(u.stripe_customer_ids, '{}'), sc.stripe_id),
    updated_at = NOW()
FROM stripe_customers sc
INNER JOIN stripe_subscriptions ss ON ss.customer_id = sc.id
WHERE sc.email = u.email
  AND ss.status IN ('active', 'trialing')
  AND u.stripe_customer_id != sc.stripe_id
  AND NOT (sc.stripe_id = ANY(COALESCE(u.stripe_customer_ids, '{}')));
```

---

## 📋 **TESTING CHECKLIST**

### **Test Eric Gessel's Account**
1. ✅ Run SQL fix to update his customer ID
2. ✅ Verify his subscription shows as "Active" in admin dashboard
3. ✅ Verify he has video access
4. ✅ Check his account shows correct expiry date (Oct 18, 2026)

### **Test Webhook Linking**
1. ✅ Create test subscription in Stripe
2. ✅ Trigger `customer.subscription.created` webhook
3. ✅ Verify new customer ID is added to `users.stripe_customer_ids`
4. ✅ Verify new customer ID becomes primary `users.stripe_customer_id`
5. ✅ Verify user shows as having active plan in elastic service

### **Test Elastic Service**
1. ✅ Query elastic service for Eric Gessel: `/admin/subscriber-elastic/subscribers?email=ericgessel@gmail.com`
2. ✅ Should show: `has_active_plan: true, plan_status: 'active', plan_name: 'Yearly'`
3. ✅ Should show: `days_until_expiry: ~356` (days until Oct 18, 2026)

---

## 🎯 **IMPLEMENTATION PRIORITY**

### **Phase 1: Immediate (Do Now)**
1. ✅ Run SQL fix for Eric Gessel
2. ✅ Run diagnostic query to find all affected users
3. ✅ Run mass update SQL for all affected users

### **Phase 2: Short-term (This Week)**
1. ✅ Implement `linkCustomerToUser()` function
2. ✅ Update `handleSubscriptionCreated()` webhook
3. ✅ Update `handleSubscriptionUpdated()` webhook
4. ✅ Deploy and test in production

### **Phase 3: Long-term (Next Sprint)**
1. ✅ Add automated job to detect mismatched customer IDs
2. ✅ Add admin dashboard alert for users with multiple customer IDs
3. ✅ Add Stripe checkout integration to pass `client_reference_id` (user ID)
4. ✅ Prevent duplicate customer creation in Stripe Checkout

---

## 🔐 **SECURITY CONSIDERATIONS**

### **Webhook Verification**
- ✅ All webhook handlers already verify signature
- ✅ Customer linking only happens for verified webhooks

### **Email Matching**
- ⚠️ **Risk**: User changes email in Stripe but not in your system
- ✅ **Mitigation**: Always use Stripe customer ID as source of truth
- ✅ **Mitigation**: Add job to sync email changes from Stripe

### **Multi-Customer Scenarios**
- ✅ Array `stripe_customer_ids` preserves history
- ✅ Most recent active customer becomes primary
- ✅ Old customers remain in array for auditing

---

## 📊 **EXPECTED OUTCOMES**

### **Before Fix**:
- ❌ Eric shows "No Plan" despite $95.64/year subscription
- ❌ Other users might have same issue
- ❌ Revenue reporting incomplete (missing active subscriptions)

### **After Fix**:
- ✅ Eric shows "Active - Yearly - $95.64/year"
- ✅ All users with active subscriptions properly linked
- ✅ Elastic service returns accurate subscription data
- ✅ Future subscriptions automatically linked via webhooks

---

## 🔧 **FILES TO MODIFY**

1. **`backend/internal/services/stripe.go`**
   - Add `linkCustomerToUser()` function
   - Update `handleSubscriptionCreated()`
   - Update `handleSubscriptionUpdated()`

2. **`database/migrations/`** (if needed)
   - Ensure `users.stripe_customer_ids` column exists
   - Ensure array type is `text[]`

3. **`backend/internal/services/subscriber_elastic_service.go`**
   - Already uses `stripe_customer_ids` array ✅
   - Should automatically work after user records fixed ✅

---

## ⚠️ **COMMON PITFALLS**

1. **Don't delete old customer IDs** - keep in array for history
2. **Don't assume one customer per user** - users can resubscribe
3. **Don't trust `users.stripe_customer_id` alone** - check array too
4. **Don't fail webhooks on linking errors** - log and continue

---

**Status**: 🚨 **CRITICAL - REQUIRES IMMEDIATE ATTENTION**  
**Affected Users**: At least 1 confirmed (Eric Gessel), likely more  
**Revenue Impact**: Active subscriptions not being tracked  
**User Impact**: Paying users showing as "No Plan"

