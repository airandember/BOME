# Phase 8: Parallel Testing - Summary

**Status**: Tools Ready, Manual Testing Phase
**Date**: October 30, 2025

---

## ✅ What's Already Built (Phase 4)

We actually already have comparison tools from Phase 4!

### **1. Comparison API Endpoint**
```
GET /api/v1/admin/subscriber-elastic/comparison/:user_id
```

**Returns**:
```json
{
  "user_id": 7374,
  "v1_data": { ... },
  "v2_data": { ... },
  "differences": [...]
}
```

### **2. Health Check Endpoint**
```
GET /api/v1/admin/subscriber-elastic/comparison/health
```

**Returns**:
```json
{
  "v2_tables_exist": true,
  "v2_data_count": {
    "customers": 150,
    "subscriptions": 145,
    "linked_users": 148
  }
}
```

---

## 🧪 Phase 8 Testing Strategy

### **Step 1: Verify v2 Tables Have Data**

```bash
curl http://localhost:8080/api/v1/admin/subscriber-elastic/comparison/health \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Expected**:
- `v2_tables_exist`: true
- `customers`: > 0
- `subscriptions`: > 0  
- `linked_users`: > 0

**If counts are 0**: Run `stripe-sync` and `customer-linking` tools first!

---

### **Step 2: Test Sample Users**

Pick 5-10 random user IDs and compare:

```bash
# User 1
curl http://localhost:8080/api/v1/admin/subscriber-elastic/comparison/7374 \
  -H "Authorization: Bearer YOUR_TOKEN"

# User 2  
curl http://localhost:8080/api/v1/admin/subscriber-elastic/comparison/7113 \
  -H "Authorization: Bearer YOUR_TOKEN"

# User 3
curl http://localhost:8080/api/v1/admin/subscriber-elastic/comparison/7309 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Review**: 
- Are `differences` arrays empty or minimal?
- Do MRR/ARR values match?
- Does `has_video_access` match?

---

### **Step 3: Test Edge Cases**

**Users with multiple subscriptions**:
```sql
SELECT u.id, u.email, COUNT(DISTINCT ss.id) as sub_count
FROM users u
JOIN stripe_customers sc ON u.email = sc.email
JOIN stripe_subscriptions ss ON ss.customer_id = sc.id
WHERE ss.status IN ('active', 'trialing')
GROUP BY u.id, u.email
HAVING COUNT(DISTINCT ss.id) > 1;
```

Test these users - v2 should handle them better!

**Users with no subscriptions**:
```sql
SELECT id, email FROM users 
WHERE id NOT IN (
  SELECT DISTINCT user_id FROM subscriptions WHERE status = 'active'
)
LIMIT 5;
```

**Users with canceled subscriptions**:
```sql
SELECT id, email FROM users u
JOIN subscriptions s ON s.user_id = u.id
WHERE s.status = 'canceled'
LIMIT 5;
```

---

### **Step 4: Performance Testing**

**Test query speed**:

```bash
# v1 endpoint
time curl http://localhost:8080/api/v1/admin/subscriber-elastic/subscribers \
  -H "Authorization: Bearer YOUR_TOKEN"

# v2 endpoint  
time curl http://localhost:8080/api/v1/admin/subscriber-elastic/subscribers-v2 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Expected**: v2 should be equal or faster (better indexes!)

---

### **Step 5: Test New User Subscription API**

**Get user's subscriptions**:
```bash
curl http://localhost:8080/api/v1/user/subscriptions \
  -H "Authorization: Bearer YOUR_USER_TOKEN"
```

**Check if user can subscribe**:
```bash
curl http://localhost:8080/api/v1/user/subscriptions/can-subscribe \
  -H "Authorization: Bearer YOUR_USER_TOKEN"
```

**Response**:
- `can_subscribe`: false (if user has active subscription)
- `message`: "User already has an active subscription (Yearly). Please use the 'Change Plan' feature instead."

---

## 📊 Success Criteria

### ✅ **Ready for Phase 9 (Migration)**

- [ ] Health check shows v2 tables populated
- [ ] Sample user comparisons show < 5% differences
- [ ] MRR/ARR calculations match (±$0.01)
- [ ] Video access flags match
- [ ] v2 API endpoints return valid data
- [ ] Performance is equal or better than v1
- [ ] User subscription API works correctly
- [ ] Can-subscribe check prevents double subscriptions

### ⚠️ **Issues to Fix Before Migration**

- [ ] Ghost subscriptions in Stripe (see Sub_Ghosts_table.txt)
- [ ] Users with multiple active subscriptions
- [ ] Unlinked Stripe customers
- [ ] MRR/ARR calculation discrepancies

---

## 🛠️ Tools Available

### **Backend CLI Tools**

1. **`stripe-sync`** - Sync Stripe data to v2 tables
   ```bash
   cd backend/cmd/stripe-sync
   go run main.go
   ```

2. **`customer-linking`** - Link users to Stripe customers
   ```bash
   cd backend/cmd/customer-linking
   go run main.go --stats
   go run main.go --link-all
   ```

3. **Comparison Tool (attempted Phase 8.1)** - Has type mismatch issues
   - Use API endpoints instead (already working!)

### **API Endpoints** (Admin)

- `GET /admin/subscriber-elastic/subscribers` - v1 data
- `GET /admin/subscriber-elastic/subscribers-v2` - v2 data
- `GET /admin/subscriber-elastic/comparison/:id` - Compare specific user
- `GET /admin/subscriber-elastic/comparison/health` - v2 health check
- `GET /admin/stripe-sync-v2/status` - Sync status
- `POST /admin/stripe-sync-v2/sync` - Trigger full sync
- `GET /admin/customer-linking/stats` - Linking statistics
- `POST /admin/customer-linking/all` - Link all users
- `GET /admin/subscription-manager/user/:id/summary` - User subscription summary

### **API Endpoints** (User)

- `GET /user/subscriptions` - Get user's subscriptions
- `GET /user/subscriptions/can-subscribe` - Check if can subscribe
- `POST /user/subscriptions/change-plan` - Change subscription plan
- `POST /user/subscriptions/cancel-multiple` - Cancel subscriptions

---

## 🎯 Recommended Testing Flow

```
1. Run stripe-sync tool
   ↓
2. Run customer-linking tool  
   ↓
3. Check health endpoint
   ↓
4. Test 10 random users via comparison endpoint
   ↓
5. Test edge cases (multiple subs, no subs, canceled)
   ↓
6. Performance test (v1 vs v2 speed)
   ↓
7. Test user subscription API
   ↓
8. Document any discrepancies
   ↓
9. Fix issues and re-test
   ↓
10. If <5% differences → Proceed to Phase 9!
```

---

## 📝 Phase 8 Completion Checklist

- [x] Phase 8.1: Comparison tool approach (using existing API endpoints)
- [ ] Phase 8.2: Generate discrepancy report (manual testing)
- [ ] Phase 8.3: Performance testing (use `time curl` commands)
- [ ] Phase 8.4: API validation (test all endpoints with real data)

---

## 🚀 Next Steps

**When testing is complete and discrepancies < 5%**:

→ **Phase 9**: Migrate existing data to v2
  - Fix users with multiple active subscriptions
  - Clean up ghost subscriptions in Stripe
  - Verify all users have proper customer links
  - Final data validation

→ **Phase 10**: Production cutover
  - Update frontend to use v2 endpoints exclusively
  - Archive v1 tables
  - Monitor for 48 hours
  - Celebrate! 🎉

---

**Phase 8 Status**: Tools ready, manual testing phase begins!

The comparison CLI tool has type issues, but we have fully functional API endpoints from Phase 4 that do the same job. Use those for testing!

