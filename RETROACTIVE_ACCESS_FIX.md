# Retroactive Video Access Fix

## 🐛 **Problem**

When a user subscribes to BOME **before** registering an account, their subscription is created but video access is not automatically granted when they later register. This happened because:

1. **Stripe checkout creates customer & subscription** → User doesn't exist yet
2. **Webhooks fire** → Customer and subscription are synced to database
3. **Webhook video access logic** → Skipped because no user found
4. **User registers** → Customer is linked to new user
5. **User logs in** → ❌ No video access despite having active subscription!

## 📋 **Example Case: brill@brillhernandezmedia.com**

### Timeline:
```
Nov 18 21:58:13 - customer.created (cus_TRqb9HajAUnoNZ)
                  ℹ️  No user found - customer synced but not linked

Nov 18 22:05:05 - customer.subscription.created (sub_1SUx1IFpxJJNWdU8Q9OW49w4)
                  Status: active
                  ℹ️  No user found - subscription synced but video access NOT granted

Nov 18 22:09:28 - User registered (ID: 10467)
                  ✅ Linked customer to user
                  ❌ But video access was never granted!

Nov 18 22:19:01 - User logs in
                  ❌ has_video_access = false
```

The user's subscription shows as active in Enhanced Subscribers table, but `has_video_access` remains `false`.

---

## ✅ **Solution**

Added **retroactive video access grant** in `customer_linking_service.go`. Now when a customer is linked to a user during registration or verification, the system:

1. ✅ Checks if the user already has video access (skip if yes)
2. ✅ Queries if any linked customers have active subscriptions
3. ✅ Grants video access automatically if active subscription found
4. ✅ Logs the retroactive grant with source: `retroactive_linking`

### Code Changes:

**File: `backend/internal/services/customer_linking_service.go`**

#### 1. Updated `LinkUserToCustomers()` function:
```go
// Check if any of the linked customers have active subscriptions and grant video access if needed
if result.CustomersLinked > 0 {
    s.checkAndGrantVideoAccessAfterLinking(userID, email)
}
```

#### 2. Added new helper function:
```go
// checkAndGrantVideoAccessAfterLinking checks if a newly linked user has active subscriptions
// and grants video access if they do (retroactive access grant)
func (s *CustomerLinkingService) checkAndGrantVideoAccessAfterLinking(userID int, email string) {
    // Check if user already has video access
    var hasAccess bool
    err := s.db.QueryRow(`
        SELECT COALESCE(has_video_access, false) 
        FROM users 
        WHERE id = $1
    `, userID).Scan(&hasAccess)
    
    if err != nil {
        log.Printf("⚠️  [Customer Linking] Failed to check video access for user %d: %v", userID, err)
        return
    }

    // If user already has access, no need to check further
    if hasAccess {
        return
    }

    // Check if any of their linked customers have active subscriptions
    query := `
        SELECT EXISTS(
            SELECT 1 
            FROM user_stripe_customers_v2 usc
            JOIN stripe_customers_v2 sc ON sc.id = usc.stripe_customer_id
            JOIN stripe_subscriptions_v2 ss ON ss.stripe_customer_id = sc.id
            WHERE usc.user_id = $1
            AND ss.status IN ('active', 'trialing')
            AND (ss.cancel_at_period_end = false OR ss.cancel_at_period_end IS NULL)
        )
    `

    var hasActiveSubscription bool
    err = s.db.QueryRow(query, userID).Scan(&hasActiveSubscription)
    if err != nil {
        log.Printf("⚠️  [Customer Linking] Failed to check active subscriptions for user %d: %v", userID, err)
        return
    }

    // If they have an active subscription, grant video access
    if hasActiveSubscription {
        _, err = s.db.Exec(`
            UPDATE users 
            SET has_video_access = true, 
                video_access_granted_at = NOW(),
                video_access_source = 'retroactive_linking'
            WHERE id = $1
        `, userID)

        if err != nil {
            log.Printf("❌ [Customer Linking] Failed to grant video access to user %d: %v", userID, err)
            return
        }

        log.Printf("✅ [Customer Linking] Granted retroactive video access to user %d (%s) - active subscription found after linking", userID, email)
    }
}
```

---

## 🔄 **When This Runs**

The retroactive access grant runs automatically in these scenarios:

1. **User registration** → `LinkUserToCustomers()` called
2. **Email verification** → `LinkUserToCustomers()` called
3. **OAuth2 login** → `LinkUserToCustomers()` called
4. **Admin manual linking** → `LinkUserToCustomers()` called

Any time a customer is linked to a user, the system now checks for active subscriptions and grants access if needed.

---

## 🧪 **Expected Behavior After Deploy**

### New User Flow (Subscribe → Register):
```
1. User subscribes (no account yet)
   → Stripe creates customer + subscription
   → Webhooks sync to database
   → No user to grant access to (expected)

2. User registers
   → Email matches Stripe customer
   → ✅ Customer linked to user
   → ✅ NEW: System detects active subscription
   → ✅ NEW: Video access granted automatically!
   → Log: "Granted retroactive video access to user X"

3. User logs in
   → ✅ Has immediate video access
   → ✅ Can watch premium content
```

### New User Flow (Register → Subscribe):
```
1. User registers
   → Customer not yet created in Stripe

2. User subscribes
   → Stripe creates customer + subscription
   → Webhooks fire
   → ✅ User found and linked
   → ✅ Video access granted by webhook (existing logic)

3. User redirected to /videos
   → ✅ Has immediate access
```

---

## 📊 **Logs to Watch For**

After deploying, you'll see new log entries like:

```
✅ [Customer Linking] Granted retroactive video access to user 10467 (brill@brillhernandezmedia.com) - active subscription found after linking
```

This confirms the system is detecting and fixing the "subscribe before register" edge case.

---

## 🔧 **Manual Fix for brill@brillhernandezmedia.com**

Since this user (ID: 10467) already exists with the issue, you can manually fix them with:

```sql
UPDATE users 
SET has_video_access = true, 
    video_access_granted_at = NOW(),
    video_access_source = 'manual_retroactive_fix'
WHERE id = 10467;
```

Or use the admin "Grant Access" button in the Enhanced Subscribers dashboard.

After deploying this fix, all **future** users who subscribe before registering will automatically get access!

---

## ✅ **Benefits**

1. ✅ Handles "subscribe before register" flow automatically
2. ✅ No manual intervention needed
3. ✅ Works for all linking scenarios (registration, verification, OAuth2)
4. ✅ Safe: Only grants access if active subscription exists
5. ✅ Logged: Clear audit trail with `video_access_source = 'retroactive_linking'`
6. ✅ No duplicate grants: Checks if user already has access first

---

## 🎯 **Related Fixes**

This completes the subscription automation trilogy:

1. ✅ **SQL Bug Fix** - Fixed `GetUserByStripeCustomerID()` to properly join tables
2. ✅ **Array Type Fix** - Fixed `pq.Array()` for subscription status checks
3. ✅ **Retroactive Access** - This fix (handles subscribe-before-register flow)

All three fixes ensure 100% automatic video access for all subscription flows! 🎉

