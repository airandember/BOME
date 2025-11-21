# Auto-Linking Audit & Fix Plan

## 🔍 Issue Discovered

**Date:** 2025-11-20  
**User:** brill@brillhernandezmedia.com (ID: 10467)  
**Customer:** cus_TScmtpJDgGwKhY  
**Subscription:** sub_1SVhX1FpxJJNWdU8HXZwtjtf (active)

### Timeline:
```
23:44:46 - Stripe customer created
23:45:00 - Stripe subscription created (active)
23:45:27 - OAuth2 auth URL generated
23:55:05 - User logged in
           ❌ NO auto-linking logs appeared
           ❌ User has no video access
```

---

## 🐛 Root Cause Analysis

### Current Auto-Linking Trigger Points:

1. ✅ **User Registration** (`auth.go` RegisterHandler) - Working
2. ✅ **Email Verification** (`auth.go` VerifyEmailLinkHandler) - Working
3. ✅ **Password Setup** (`auth.go` SetupPasswordHandler) - Fixed 2025-11-18
4. ⚠️ **OAuth2 New User** (`oauth2.go` CreateOrLinkUser) - Should work
5. ⚠️ **OAuth2 Existing User** (`oauth2.go` CreateOrLinkUser) - **NOT WORKING**

### Why OAuth2 Existing User Failed:

Looking at `oauth2.go` lines 295-304:
```go
// 🔗 AUTO-LINK: Attempt to link any existing Stripe customers with matching email
linkingService := NewCustomerLinkingService(s.db)
linkResult, err := linkingService.LinkUserToCustomers(existingUser.ID)
if err != nil {
    log.Printf("⚠️ [OAUTH2] Failed to auto-link Stripe customers for existing user %d: %v", existingUser.ID, err)
} else if linkResult.CustomersLinked > 0 {
    log.Printf("✅ [OAUTH2] Auto-linked %d Stripe customer(s) to existing user %d (%s)", 
        linkResult.CustomersLinked, existingUser.ID, existingUser.Email)
}
```

**The logging is correct, but we're missing logs!** This means:
- Either the code didn't execute (unreachable?)
- Or `LinkUserToCustomers` returned 0 customers linked

### Possible Issues:

1. **Email Case Mismatch in User Table**
   - Stripe: `Brill@brillhernandezmedia.com`
   - User table might have: Different case
   - Even though we use LOWER(), the user's email needs to match

2. **Timing Issue**
   - Customer was synced at 23:44:46
   - User logged in at 23:55:05
   - 30-minute gap - customer should be in DB

3. **Silent Failure**
   - `LinkUserToCustomers` might be returning error
   - But error log not showing because it's being swallowed

---

## 🔧 Comprehensive Fix Strategy

### Phase 1: Add Debugging & Better Logging

**Problem:** Auto-linking failures are silent - we don't know WHY it didn't link.

**Solution:** Add detailed logging at every step:

```go
// In oauth2.go CreateOrLinkUser (existing user branch)
log.Printf("🔍 [OAUTH2-LINK] Starting auto-link for existing user %d (%s)", existingUser.ID, existingUser.Email)

linkingService := NewCustomerLinkingService(s.db)
linkResult, err := linkingService.LinkUserToCustomers(existingUser.ID)

// Log the full result for debugging
log.Printf("🔍 [OAUTH2-LINK] Link result: CustomersFound=%d, CustomersLinked=%d, Error=%s", 
    linkResult.CustomersFound, linkResult.CustomersLinked, linkResult.Error)

if err != nil {
    log.Printf("❌ [OAUTH2] Failed to auto-link Stripe customers for existing user %d: %v", existingUser.ID, err)
} else if linkResult.CustomersLinked > 0 {
    log.Printf("✅ [OAUTH2] Auto-linked %d Stripe customer(s) to existing user %d (%s)", 
        linkResult.CustomersLinked, existingUser.ID, existingUser.Email)
} else if linkResult.CustomersFound > 0 {
    log.Printf("⚠️ [OAUTH2] Found %d customers for user %d but linked %d - Check skipped: %v",
        linkResult.CustomersFound, existingUser.ID, linkResult.CustomersLinked, linkResult.SkippedCustomers)
} else {
    log.Printf("ℹ️ [OAUTH2] No Stripe customers found for user %d (%s)", existingUser.ID, existingUser.Email)
}
```

---

### Phase 2: Force Re-Link on Every Login

**Problem:** Auto-linking only adds NEW links. If a customer is created AFTER a user logs in once, they won't be linked until next registration event.

**Solution:** Add a login hook that ALWAYS checks for new customers:

```go
// In auth.go LoginHandler (after successful login)
// 🔗 AUTO-LINK CHECK: On every login, check for new unlinked customers
linkingService := services.NewCustomerLinkingService(db)
linkResult, err := linkingService.LinkUserToCustomers(user.ID)
if err != nil {
    log.Printf("⚠️ [LOGIN] Failed to check for unlinked customers: %v", err)
} else if linkResult.CustomersLinked > 0 {
    log.Printf("✅ [LOGIN] Auto-linked %d new Stripe customer(s) on login for user %d", 
        linkResult.CustomersLinked, user.ID)
}
```

---

### Phase 3: Periodic Background Linking Job

**Problem:** If webhooks arrive when user is offline, they might never get linked.

**Solution:** Add a background job that runs every hour:

```go
// New file: backend/internal/services/linking_job.go

func RunPeriodicLinking(db *database.DB) {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    
    for range ticker.C {
        log.Printf("🔄 [LINKING-JOB] Starting periodic linking check...")
        
        // Find users with Stripe customers but no links
        query := `
            SELECT DISTINCT u.id, u.email
            FROM users u
            INNER JOIN stripe_customers_v2 sc ON LOWER(sc.email) = LOWER(u.email)
            LEFT JOIN user_stripe_customers_v2 usc ON usc.user_id = u.id AND usc.stripe_customer_id = sc.id
            WHERE usc.id IS NULL
            LIMIT 100
        `
        
        rows, err := db.Query(query)
        if err != nil {
            log.Printf("❌ [LINKING-JOB] Query failed: %v", err)
            continue
        }
        defer rows.Close()
        
        linkedCount := 0
        linkingService := NewCustomerLinkingService(db)
        
        for rows.Next() {
            var userID int
            var email string
            if err := rows.Scan(&userID, &email); err != nil {
                continue
            }
            
            result, err := linkingService.LinkUserToCustomers(userID)
            if err == nil && result.CustomersLinked > 0 {
                linkedCount++
                log.Printf("✅ [LINKING-JOB] Linked %d customers for user %d (%s)", 
                    result.CustomersLinked, userID, email)
            }
        }
        
        if linkedCount > 0 {
            log.Printf("🎉 [LINKING-JOB] Completed: Linked %d users to their customers", linkedCount)
        }
    }
}
```

---

### Phase 4: Manual Linking Admin Endpoint

**Problem:** For immediate fixes, admins need a way to trigger linking manually.

**Solution:** Add admin endpoint to force linking:

```go
// POST /api/v1/admin/users/:id/link-customers
func LinkCustomersHandler(db *database.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, err := strconv.Atoi(c.Param("id"))
        if err != nil {
            c.JSON(400, gin.H{"error": "Invalid user ID"})
            return
        }
        
        linkingService := services.NewCustomerLinkingService(db)
        result, err := linkingService.LinkUserToCustomers(userID)
        
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        
        c.JSON(200, gin.H{
            "success": true,
            "result": result,
        })
    }
}
```

---

## 📋 Implementation Checklist

### Immediate Fixes (High Priority):
- [ ] Add detailed logging to OAuth2 auto-linking
- [ ] Add detailed logging to all auto-linking trigger points
- [ ] Add auto-linking check on every login
- [ ] Test with the failing user case

### Short-term Improvements:
- [ ] Add periodic background linking job
- [ ] Add admin endpoint for manual linking
- [ ] Add monitoring/alerting for unlinked customers

### Long-term Monitoring:
- [ ] Dashboard showing "unlinked customers with subscriptions"
- [ ] Alert if unlinked count > 5
- [ ] Weekly report of linking statistics

---

## 🧪 Testing Plan

### Test Case 1: Subscribe Before Register (OAuth2)
```
1. Create Stripe customer manually
2. Add subscription
3. Register via Google OAuth2
4. Verify: Customer linked automatically
5. Verify: Video access granted
```

### Test Case 2: Subscribe Before Register (Email/Password)
```
1. Create Stripe customer manually
2. Add subscription
3. Register via email/password
4. Verify email
5. Set password
6. Verify: Customer linked automatically
7. Verify: Video access granted
```

### Test Case 3: Existing User, New Subscription
```
1. User exists and logged in
2. Create Stripe customer (same email)
3. Add subscription
4. User logs in again
5. Verify: Customer linked on login
6. Verify: Video access granted
```

### Test Case 4: Case Sensitivity
```
1. User: brill@example.com
2. Stripe: Brill@example.com (capital B)
3. Verify: Still links correctly (LOWER() working)
```

---

## 🎯 Success Criteria

After fixes are implemented:

1. ✅ **100% auto-linking rate** for subscribe-before-register
2. ✅ **< 1 minute delay** from subscription to access grant
3. ✅ **Zero manual interventions** required
4. ✅ **Clear logs** showing exactly what happened
5. ✅ **Periodic job** catches any missed links within 1 hour

---

## 📊 Monitoring Queries

### Check for Unlinked Customers with Active Subscriptions:
```sql
SELECT 
    sc.stripe_id,
    sc.email,
    ss.status as subscription_status,
    u.id as user_id,
    u.email as user_email
FROM stripe_customers_v2 sc
INNER JOIN stripe_subscriptions_v2 ss ON ss.customer_id = sc.id
LEFT JOIN users u ON LOWER(u.email) = LOWER(sc.email)
LEFT JOIN user_stripe_customers_v2 usc ON usc.user_id = u.id AND usc.stripe_customer_id = sc.id
WHERE ss.status IN ('active', 'trialing')
  AND u.id IS NOT NULL
  AND usc.id IS NULL
ORDER BY ss.created_at DESC;
```

### Check Linking Success Rate (Last 24 Hours):
```sql
SELECT 
    COUNT(*) FILTER (WHERE usc.id IS NOT NULL) as linked_count,
    COUNT(*) FILTER (WHERE usc.id IS NULL) as unlinked_count,
    ROUND(100.0 * COUNT(*) FILTER (WHERE usc.id IS NOT NULL) / COUNT(*), 2) as success_rate
FROM stripe_customers_v2 sc
LEFT JOIN users u ON LOWER(u.email) = LOWER(sc.email)
LEFT JOIN user_stripe_customers_v2 usc ON usc.user_id = u.id AND usc.stripe_customer_id = sc.id
WHERE sc.stripe_created_at > NOW() - INTERVAL '24 hours'
  AND u.id IS NOT NULL;
```

---

Last Updated: 2025-11-21  
Status: ⚠️  NEEDS IMPLEMENTATION

