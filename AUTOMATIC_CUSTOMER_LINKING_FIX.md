# 🔗 Automatic Customer Linking Fix

## 📋 Problem Statement

**Issue:** When users created Stripe customers **before** registering for a BOME account, the customer would sync to the database but remain unlinked to the user account. This caused:

1. Customers appearing in "Stripe-only" table instead of "Synced" table
2. Users showing "No Plan" even with active subscriptions
3. Manual intervention required via "Add User" or "Add All Users" buttons

### Example Flow That Failed:
```
1. User visits checkout → Stripe creates customer (cus_TR0lAvH1rNi9jR)
2. Stripe webhook → Customer synced to stripe_customers_v2 table
3. Webhook tries to link → ⚠️ No user exists yet, logs warning and continues
4. User registers later → Account created, but NO automatic linking
5. Result: Customer unlinked, user has no subscription access
```

## ✅ Solution

Implemented **automatic customer linking** at **5 critical touchpoints** in the user lifecycle:

### 1. **New User Registration** (Email/Password)
- **Location:** `backend/internal/routes/auth.go` → `RegisterHandler`
- **When:** Immediately after user account creation
- **Action:** Links any existing Stripe customers with matching email

```go
// 🔗 AUTO-LINK: Attempt to link any existing Stripe customers with matching email
linkingService := services.NewCustomerLinkingService(db)
linkResult, err := linkingService.LinkUserToCustomers(user.ID)
if err != nil {
    log.Printf("⚠️  Failed to auto-link Stripe customers for new user %d: %v", user.ID, err)
} else if linkResult.CustomersLinked > 0 {
    log.Printf("✅ Auto-linked %d Stripe customer(s) to new user %d (%s)", 
        linkResult.CustomersLinked, user.ID, user.Email)
}
```

### 2. **Existing User Re-registration** (Unified Flow)
- **Location:** `backend/internal/routes/auth.go` → `RegisterHandler` (existing user path)
- **When:** When user tries to register with an email that already exists
- **Action:** Links any Stripe customers created after initial registration

### 3. **Email Verification** (Safety Net)
- **Location:** `backend/internal/routes/auth.go` → `VerifyEmailLinkHandler`
- **When:** After user verifies their email
- **Action:** Catches any customers that weren't linked during registration
- **Purpose:** Backup layer in case registration linking failed

### 4. **OAuth2 New User** (Google Sign-In)
- **Location:** `backend/internal/services/oauth2.go` → `CreateOrLinkUser` (new user path)
- **When:** After creating new user via OAuth2
- **Action:** Links existing Stripe customers to newly created OAuth2 user

### 5. **OAuth2 Existing User** (Google Sign-In)
- **Location:** `backend/internal/services/oauth2.go` → `CreateOrLinkUser` (existing user path)
- **When:** Existing user signs in with OAuth2
- **Action:** Links customers that may have been created before OAuth2 login

## 🔄 How It Works

### CustomerLinkingService.LinkUserToCustomers()
```sql
-- Finds all Stripe customers with matching email
SELECT id, stripe_id, stripe_created_at 
FROM stripe_customers_v2 
WHERE LOWER(email) = LOWER($1)
ORDER BY stripe_created_at DESC

-- Creates link in junction table
INSERT INTO user_stripe_customers_v2 
(user_id, stripe_customer_id, is_primary, first_linked_at, last_synced_at)
VALUES ($1, $2, $3, NOW(), NOW())

-- Updates user's primary customer
UPDATE users 
SET stripe_customer_id = $1 
WHERE id = $2
```

### Logging Behavior
- **Silent on single customer link** (normal case)
- **Logs when multiple customers found** (unusual)
- **Logs errors** (for troubleshooting)

## 📊 Expected Behavior After Fix

### Scenario 1: Checkout → Register
```
1. User visits /videos → Redirected to checkout
2. Stripe creates customer cus_ABC123 with email user@example.com
3. Webhook syncs customer to stripe_customers_v2
4. User completes registration with email user@example.com
5. ✅ AUTO-LINKED during registration
6. Result: User immediately has subscription access
```

### Scenario 2: Checkout → Google Sign-In
```
1. User visits /videos → Redirected to checkout
2. Stripe creates customer cus_XYZ789 with email user@gmail.com
3. Webhook syncs customer
4. User signs in with Google (user@gmail.com)
5. ✅ AUTO-LINKED during OAuth2 flow
6. Result: User immediately has subscription access
```

### Scenario 3: Register → Checkout → Verify
```
1. User registers but doesn't verify email
2. User visits /videos → Redirected to checkout
3. Stripe creates customer
4. User verifies email
5. ✅ AUTO-LINKED during email verification (safety net)
6. Result: User has subscription access after verification
```

## 🧪 Testing

### Manual Test Cases

**Test 1: Checkout Before Registration**
1. Clear cookies/logout
2. Go to `/videos` → Checkout
3. Complete Stripe checkout with test card
4. Register with same email used in checkout
5. ✅ Verify user appears in "Synced Customers" table
6. ✅ Verify user shows correct plan in "Enhanced Subscribers"

**Test 2: Checkout Before Google Sign-In**
1. Clear cookies/logout
2. Go to `/videos` → Checkout
3. Complete Stripe checkout with Gmail address
4. Sign in with Google OAuth2
5. ✅ Verify immediate subscription access

**Test 3: Multiple Customers**
1. Create 2+ Stripe customers with same email
2. Register with that email
3. ✅ Verify all customers linked
4. ✅ Verify most recent customer is primary

### Production Logs to Watch

Look for these new log entries:

```
✅ Auto-linked 1 Stripe customer(s) to new user 10460 (me_aprilrain@yahoo.com)
✅ [OAUTH2] Auto-linked 1 Stripe customer(s) to new user 10461 (user@gmail.com)
✅ Auto-linked 1 Stripe customer(s) during email verification for user 10462
```

## 🔍 Troubleshooting

### If Linking Still Fails

**Check 1: Email Match**
```sql
-- Verify emails match exactly (case-insensitive)
SELECT 
    u.id, u.email as user_email, 
    sc.stripe_id, sc.email as customer_email
FROM users u
LEFT JOIN stripe_customers_v2 sc ON LOWER(u.email) = LOWER(sc.email)
WHERE u.id = [USER_ID];
```

**Check 2: Customer Synced**
```sql
-- Verify customer exists in v2 table
SELECT * FROM stripe_customers_v2 
WHERE stripe_id = 'cus_ABC123';
```

**Check 3: Link Created**
```sql
-- Verify link in junction table
SELECT * FROM user_stripe_customers_v2 
WHERE user_id = [USER_ID];
```

**Check 4: Webhook Timing**
- If customer webhook arrives AFTER registration completes, linking happens in webhook
- If customer webhook arrives BEFORE registration, linking happens at registration
- Both paths are covered ✅

## 📚 Related Files

### Modified
- `backend/internal/routes/auth.go` (3 locations)
- `backend/internal/services/oauth2.go` (2 locations)

### Dependencies
- `backend/internal/services/customer_linking_service.go`
- `backend/internal/database/db.go`

### Database Tables
- `stripe_customers_v2` (source of customers)
- `user_stripe_customers_v2` (junction table)
- `users` (stores primary customer ID)

## 🎯 Impact

### Before Fix
- **Manual intervention required** for checkout-before-registration flow
- **Poor user experience** (no immediate access after purchase)
- **Support burden** from users asking why they don't have access

### After Fix
- **✅ Automatic linking** in all scenarios
- **✅ Immediate subscription access** after registration/login
- **✅ Reduced support tickets**
- **✅ Better conversion rates** (seamless experience)

## 🚀 Deployment

### Prerequisites
- None (backward compatible)

### Deployment Steps
1. Deploy updated backend
2. Monitor logs for auto-linking messages
3. Verify "Stripe-only" table decreases over time

### Rollback Plan
- Safe to rollback (only adds linking, doesn't remove existing functionality)
- Manual "Add User" buttons still work as backup

## 📈 Metrics to Track

- Decrease in "Stripe-only customers" count
- Increase in successful auto-links (check logs)
- Reduction in support tickets about missing subscriptions
- Improved checkout-to-access conversion rate

## 🐛 Bonus Fix: "Add User" Button

### Problem
The individual "Add User" button in the Stripe-only customers table was not working because of a type mismatch between the parent and child components.

### Root Cause
```typescript
// Parent expected CustomEvent:
function handleCreateUser(event: CustomEvent) {
    const customer = event.detail;
    createUserFromStripe(customer);
}

// But child was calling it directly with customer object:
oncreateUser?.(customer); // Not an event!
```

### Fix
Updated parent handlers to accept customer objects directly instead of `CustomEvent`:

```typescript
// ✅ Now accepts customer directly
function handleCreateUser(customer: any) {
    createUserFromStripe(customer);
}
```

### Impact
- **✅ "Add User" button now works** for individual customer linking
- **✅ "Add All Users" continues to work** as before
- **✅ No changes needed** to child components

## ✅ Completion Checklist

- [x] Auto-link on new user registration
- [x] Auto-link on existing user re-registration
- [x] Auto-link on email verification (safety net)
- [x] Auto-link on OAuth2 new user
- [x] Auto-link on OAuth2 existing user
- [x] Backend builds successfully
- [x] No backend linter errors
- [x] Fixed "Add User" button in Stripe customers UI
- [x] No frontend linter errors
- [x] Documentation created
- [ ] Deployed to production
- [ ] Monitoring confirms auto-linking works
- [ ] User report confirms fix resolved their issue

---

**Date:** November 16, 2025  
**Issues:**  
1. User `me_aprilrain@yahoo.com` (ID: 10460, Customer: `cus_TR0lAvH1rNi9jR`) appeared in Stripe-only table after checkout → register flow  
2. "Add User" button in Stripe-only customers table not functioning

**Root Causes:**  
1. No automatic customer linking at registration touchpoints  
2. Type mismatch between parent/child component event handlers

**Fixes:**  
1. Added 5-point automatic linking system covering all user lifecycle paths  
2. Updated event handlers to accept customer objects directly

**Status:** ✅ Implemented & Built Successfully

