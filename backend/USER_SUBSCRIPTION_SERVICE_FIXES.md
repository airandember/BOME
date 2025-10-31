# User Subscription Service - Column Name Fixes

## Issue
The `UserSubscriptionService` was using incorrect column names when querying the v2 Stripe tables, causing 500 errors on the user subscription dashboard.

## Root Cause
The service was written using placeholder/assumed column names that didn't match the actual v2 table schema defined in `050_create_stripe_v2_schema.sql`.

## Fixes Applied

### Fix 1: PostgreSQL Array Parameter
**Error**: `sql: converting argument $1 type: unsupported type []string`

**Location**: `backend/internal/services/user_subscription_service.go:132`

**Fix**:
```go
// Before:
rows, err := s.db.Query(query, linkedCustomers)

// After:
rows, err := s.db.Query(query, pq.Array(linkedCustomers))
```

**Import Added**:
```go
import "github.com/lib/pq"
```

---

### Fix 2: Customer ID Column Name
**Error**: `pq: column ss.stripe_customer_id does not exist`

**Location**: `backend/internal/services/user_subscription_service.go:125`

**Fix**:
```sql
-- Before:
JOIN stripe_customers_v2 sc ON ss.stripe_customer_id = sc.id

-- After:
JOIN stripe_customers_v2 sc ON ss.customer_id = sc.id
```

**Reason**: The `stripe_subscriptions_v2` table uses `customer_id`, not `stripe_customer_id` (see migration line 142).

---

### Fix 3: Recurring Interval Column Name
**Error**: `pq: column spr.interval does not exist`

**Location**: `backend/internal/services/user_subscription_service.go:118`

**Fix**:
```sql
-- Before:
COALESCE(spr.interval, 'month') as interval

-- After:
COALESCE(spr.recurring_interval, 'month') as interval
```

**Reason**: The `stripe_prices_v2` table uses `recurring_interval`, not `interval` (see migration line 105).

---

## Verification
After all three fixes:
1. Backend starts without errors
2. User can access `/dashboard?tab=subscription`
3. API call to `/api/v1/user/subscriptions` returns 200 OK
4. Subscription data loads correctly (or shows "No active subscription" message if none exists)

---

## Related Files
- `backend/internal/services/user_subscription_service.go` - Service with fixes
- `backend/migrations/050_create_stripe_v2_schema.sql` - Table schema definitions
- `frontend/src/routes/dashboard/+page.svelte` - Frontend dashboard
- `frontend/src/lib/components/SubscriptionManagement.svelte` - Subscription UI component

---

## Date
October 31, 2024 05:11 AM MST

