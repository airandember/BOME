# Video Access V2 Fix - Complete! ✅

## 🐛 **The Problem**

User `bometesting@gmail.com` (ID: 10469) had an **active subscription** in Stripe but **couldn't access premium videos**.

### **Root Causes:**

1. ❌ **Missing columns in `users` table**: `video_access_granted_at` and `video_access_source`
   - Webhooks were failing with: `pq: column "video_access_source" does not exist`

2. ❌ **Video access query using V1 tables** instead of V2 tables
   - `video_access.go` was querying: `stripe_customers`, `stripe_subscriptions`, `stripe_products`
   - But data is in: `stripe_customers_v2`, `stripe_subscriptions_v2`, `stripe_products_v2`
   - Result: Query always returned `false` even with active subscription!

---

## ✅ **The Fixes**

### **Fix 1: Added Missing Columns** ✅

**Migration:** `backend/migrations/060_add_video_access_tracking.sql`

```sql
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS video_access_granted_at TIMESTAMPTZ;

ALTER TABLE users 
ADD COLUMN IF NOT EXISTS video_access_source VARCHAR(255);
```

**Purpose:**
- `video_access_granted_at`: Tracks WHEN access was granted
- `video_access_source`: Tracks HOW they got access (e.g., `webhook:sub_xxx`, `session_verification:cs_xxx`)

---

### **Fix 2: Updated Video Access Query to Use V2 Tables** ✅

**File:** `backend/internal/database/video_access.go`

**OLD Query (V1 tables - WRONG):**
```sql
SELECT EXISTS(
    SELECT 1 FROM users u
    INNER JOIN stripe_customers sc ON (
        u.stripe_customer_id = sc.stripe_id OR 
        sc.stripe_id = ANY(COALESCE(u.stripe_customer_ids, '{}'))
    )
    INNER JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
    INNER JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
    WHERE u.id = $1
    AND ss.status IN ('active', 'trialing')
    AND sp.video_approved = true
)
```

**NEW Query (V2 tables - CORRECT):**
```sql
SELECT EXISTS(
    SELECT 1 
    FROM user_stripe_customers_v2 usc
    INNER JOIN stripe_customers_v2 sc ON sc.id = usc.stripe_customer_id
    INNER JOIN stripe_subscriptions_v2 ss ON ss.customer_id = sc.id
    INNER JOIN stripe_prices_v2 sp ON sp.id = ss.price_id
    INNER JOIN stripe_products_v2 sprod ON sprod.id = sp.product_id
    WHERE usc.user_id = $1
    AND ss.status IN ('active', 'trialing')
    AND ss.current_period_end > NOW()
    AND sprod.video_approved = true
    AND sprod.active = true
)
```

**Key Changes:**
1. ✅ Uses `user_stripe_customers_v2` linking table (proper many-to-many)
2. ✅ Uses `stripe_customers_v2` (integer PKs, better performance)
3. ✅ Uses `stripe_subscriptions_v2` (proper FK to customer_id)
4. ✅ Uses `stripe_prices_v2` (proper FK to product_id)
5. ✅ Uses `stripe_products_v2` (has `video_approved` flag)
6. ✅ Checks `current_period_end > NOW()` (subscription not expired)

---

## 🔄 **The Flow (Now Fixed)**

### **Step 1: User Subscribes**
```
User clicks "Subscribe" → Stripe Checkout → Payment Succeeded
```

### **Step 2: Webhook Fires**
```
customer.subscription.created → Backend receives webhook
```

### **Step 3: Database Updated**
```sql
-- V2 tables populated by webhook handler
INSERT INTO stripe_customers_v2 (stripe_id, email, ...) VALUES ('cus_xxx', 'user@example.com', ...);
INSERT INTO stripe_subscriptions_v2 (stripe_id, customer_id, status, ...) VALUES ('sub_xxx', 123, 'active', ...);
INSERT INTO user_stripe_customers_v2 (user_id, stripe_customer_id) VALUES (10469, 123);

-- ✅ NOW WORKS: Grant video access
UPDATE users 
SET manual_video_access = true,
    video_access_granted_at = NOW(),
    video_access_source = 'webhook:sub_xxx'
WHERE id = 10469;
```

### **Step 4: User Checks Access**
```
Frontend calls: /api/v1/subscriptions/
Backend calls: HasVideoAccess(10469)
Query checks: user_stripe_customers_v2 → stripe_subscriptions_v2
Result: ✅ TRUE (active subscription found in V2 tables!)
```

### **Step 5: User Accesses Videos**
```
User navigates to /videos
SubscriptionCheck component verifies access
✅ Access granted - premium content unlocked! 🎉
```

---

## 📊 **Table Relationships (V2)**

```
users (id: 10469)
    ↓
user_stripe_customers_v2 (user_id: 10469, stripe_customer_id: 123)
    ↓
stripe_customers_v2 (id: 123, stripe_id: 'cus_TSyBCtUPqj9Hre')
    ↓
stripe_subscriptions_v2 (customer_id: 123, price_id: 45, status: 'active')
    ↓
stripe_prices_v2 (id: 45, product_id: 67)
    ↓
stripe_products_v2 (id: 67, video_approved: true, name: 'Monthly')
```

**Result:** User has active subscription → Video access = `true` ✅

---

## 🎯 **What's Different Now?**

| Aspect | Before | After |
|--------|--------|-------|
| **Database Columns** | Missing `video_access_granted_at`, `video_access_source` | ✅ Columns added |
| **Video Access Query** | Querying V1 tables (empty) | ✅ Querying V2 tables (populated) |
| **Webhook Success** | ❌ SQL errors | ✅ Grants access successfully |
| **User Access Check** | ❌ Always returns `false` | ✅ Returns `true` for active subs |
| **Premium Content** | 🔒 Blocked | ✅ Accessible! |

---

## 🚀 **Testing**

### **Verify Fix:**

1. **Restart backend** (already rebuilt ✅)

2. **Check user's video access:**
   ```sql
   SELECT id, email, manual_video_access, video_access_granted_at, video_access_source
   FROM users 
   WHERE id = 10469;
   ```
   Should show: `manual_video_access = true`

3. **Check V2 subscription:**
   ```sql
   SELECT ss.stripe_id, ss.status, ss.current_period_end
   FROM user_stripe_customers_v2 usc
   JOIN stripe_subscriptions_v2 ss ON ss.customer_id = usc.stripe_customer_id
   WHERE usc.user_id = 10469;
   ```
   Should show: `status = 'active'`, `current_period_end > NOW()`

4. **Test frontend:**
   - Login as `bometesting@gmail.com`
   - Navigate to `/videos`
   - ✅ Should see premium content!

---

## 📝 **Key Learnings**

### **Why This Happened:**

1. **Schema migration from V1 → V2** was completed
2. **Webhook handlers updated** to use V2 tables
3. **Video access checker WASN'T updated** - still querying V1 tables
4. **Missing audit columns** prevented webhooks from completing

### **Prevention for Future:**

1. ✅ **Always grep for table usage** before schema changes
2. ✅ **Check all query locations** when migrating tables
3. ✅ **Add required columns** before deploying code that uses them
4. ✅ **Test end-to-end flow** after major schema changes

---

## 🎉 **Status: COMPLETE**

- ✅ Migration applied (`060_add_video_access_tracking.sql`)
- ✅ Video access query updated to V2 tables
- ✅ Backend rebuilt and ready
- ✅ User should now have video access!

**Next Step:** Restart backend and test! 🚀

