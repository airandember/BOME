# 👻 Ghost Subscriptions - Plan Codes Added

**Date:** November 26, 2025  
**Status:** ✅ COMPLETE

---

## 🎯 Enhancement Made

### Original Implementation:
- Only checked ghost **product IDs** (e.g., `prod_FvNAeI348dup9w`)

### Updated Implementation:
- Now checks both ghost **product IDs** AND **plan codes** (e.g., "Combo", "SYearPlus")

---

## 📋 Ghost Identifiers Now Supported

### Ghost Product IDs:
```go
ghostProductIDs := []string{
    "prod_FvNAeI348dup9w", // Combo product ID
    "prod_HEmcX1PE8TO2CO", // Combo product ID
    "prod_HF5YzcBH5Rwr0d", // Combo product ID
    "prod_FvNAJgnw48hwpZ", // SYearPlus product ID
    "prod_GVV5efccnh13h9", // SYearPlus product ID
}
```

### Ghost Plan Codes (NEW!):
```go
ghostPlanCodes := []string{
    "Combo",     // Plan code
    "SYearPlus", // Plan code
}
```

---

## 🔧 Implementation Details

### Updated: `backend/internal/services/subscriber_elastic_service_v2.go`

#### 1. Separated Arrays
Instead of mixing product IDs and plan codes in one array, we now have two:
- `ghostProductIDs` - For checking `product_id` column
- `ghostPlanCodes` - For checking `product_name` column

#### 2. Updated SQL Query (Both Functions)

**Added Check for Plan Codes:**
```sql
user_access AS (
    SELECT 
        u.id as user_id,
        CASE 
            WHEN u.manual_video_access = true THEN true
            WHEN us.subscription_status IN ('active', 'trialing') AND us.video_approved = true THEN true
            WHEN us.subscription_status IN ('active', 'trialing') AND us.product_id = ANY($2::text[]) THEN true
            WHEN us.subscription_status IN ('active', 'trialing') AND us.product_name = ANY($3::text[]) THEN true  -- NEW!
            ELSE false
        END as has_video_access,
        ...
```

#### 3. Updated Query Execution

**GetUnifiedSubscriberByIDV2:**
```go
err := s.db.QueryRow(query, userID, pq.Array(ghostProductIDs), pq.Array(ghostPlanCodes)).Scan(...)
```

**GetAllUnifiedSubscribersV2:**
```go
rows, err := s.db.Query(query, pq.Array(ghostProductIDs), pq.Array(ghostPlanCodes))
```

---

## 🔍 Access Decision Tree (Updated)

```
User Requests Video Access
    ↓
Check 1: Manual Video Access?
    YES → ✅ GRANT ACCESS
    NO  → Continue to Check 2
    ↓
Check 2: Active/Trialing Subscription with video_approved = true?
    YES → ✅ GRANT ACCESS
    NO  → Continue to Check 3
    ↓
Check 3: Active/Trialing Subscription with Product ID in Ghost Product IDs?
    YES → ✅ GRANT ACCESS
    NO  → Continue to Check 4
    ↓
Check 4: Active/Trialing Subscription with Product Name in Ghost Plan Codes? (NEW!)
    YES → ✅ GRANT ACCESS
    NO  → ❌ DENY ACCESS
```

---

## 💡 Why Two Arrays?

### Technical Reason:
- **`product_id`** comes from `stripe_products_v2.stripe_id` (Stripe's product identifier)
- **`product_name`** comes from `stripe_products_v2.name` (Human-readable plan code)

### Use Cases:
1. **Ghost Product IDs** - For subscriptions where the Stripe product was deleted but we have the ID
2. **Ghost Plan Codes** - For subscriptions identified by internal plan codes (Combo, SYearPlus)

This dual approach catches **all variants** of ghost subscriptions!

---

## 🧪 Testing

### Test Case 1: User with Ghost Product ID
```sql
SELECT u.id, u.email, sp.stripe_id as product_id, sp.name as product_name
FROM users u
JOIN user_stripe_customers_v2 usc ON usc.user_id = u.id AND usc.is_primary = true
JOIN stripe_customers_v2 sc ON sc.id = usc.stripe_customer_id
JOIN stripe_subscriptions_v2 ss ON ss.customer_id = sc.id
LEFT JOIN stripe_prices_v2 spr ON ss.price_id = spr.id
LEFT JOIN stripe_products_v2 sp ON spr.product_id = sp.id
WHERE sp.stripe_id IN (
    'prod_FvNAeI348dup9w',
    'prod_HEmcX1PE8TO2CO',
    'prod_HF5YzcBH5Rwr0d',
    'prod_FvNAJgnw48hwpZ',
    'prod_GVV5efccnh13h9'
)
AND ss.status IN ('active', 'trialing')
LIMIT 5;
```

### Test Case 2: User with Ghost Plan Code
```sql
SELECT u.id, u.email, sp.stripe_id as product_id, sp.name as product_name
FROM users u
JOIN user_stripe_customers_v2 usc ON usc.user_id = u.id AND usc.is_primary = true
JOIN stripe_customers_v2 sc ON sc.id = usc.stripe_customer_id
JOIN stripe_subscriptions_v2 ss ON ss.customer_id = sc.id
LEFT JOIN stripe_prices_v2 spr ON ss.price_id = spr.id
LEFT JOIN stripe_products_v2 sp ON spr.product_id = sp.id
WHERE sp.name IN ('Combo', 'SYearPlus')
AND ss.status IN ('active', 'trialing')
LIMIT 5;
```

**Expected:** Both query types should return users with `has_video_access = true`

---

## 📊 Enhanced Coverage

### Before (Product IDs Only):
- ✅ Users with `prod_FvNAeI348dup9w` → Access
- ✅ Users with `prod_HEmcX1PE8TO2CO` → Access
- ❌ Users with plan name "Combo" (no product ID) → **DENIED**

### After (Product IDs + Plan Codes):
- ✅ Users with `prod_FvNAeI348dup9w` → Access
- ✅ Users with `prod_HEmcX1PE8TO2CO` → Access
- ✅ Users with plan name "Combo" → **ACCESS GRANTED!** 🎉
- ✅ Users with plan name "SYearPlus" → **ACCESS GRANTED!** 🎉

**More comprehensive coverage = Fewer false negatives!**

---

## ✅ Compilation Status

### Build Result:
```bash
go build -o bome-backend.exe main.go
```
**Status:** ✅ SUCCESS (No errors)

### Linter Result:
```
No linter errors found.
```
**Status:** ✅ CLEAN

---

## 🚀 Deployment

### Changes:
- **1 file modified:** `backend/internal/services/subscriber_elastic_service_v2.go`
- **0 database changes**
- **0 breaking changes**

### Deploy:
```powershell
cd S:\AirEmber\BOME\BOME\backend
.\bome-backend.exe
```

### Monitor Logs For:
```
✅ [SubscriptionValidation] User X has video access via ELASTIC SERVICE - Plan: Combo
✅ [SubscriptionValidation] User X has video access via ELASTIC SERVICE - Plan: SYearPlus
```

---

## 🎉 Complete!

Ghost subscriptions now supported via:
- ✅ Product IDs (`prod_*`)
- ✅ Plan Codes (`Combo`, `SYearPlus`)

**Maximum coverage for legitimate subscribers!** 👻➡️🎬

