# Stripe V2 - Remaining Phases - Oct 31, 2025

## ✅ Completed Phases

- ✅ Phase 1: Run migration SQL and verify tables
- ✅ Phase 2: Create StripeSync_v2Service
- ✅ Phase 3: Create CustomerLinkingService
- ✅ Phase 4: Create SubscriberElasticService_v2
- ✅ Phase 5: Update webhook handlers (dual-write)
- ✅ Phase 6: Implement single subscription business logic
- ✅ Phase 7: User-controlled subscription management
- ✅ Phase 8: Run v1 and v2 in parallel, compare results
- ✅ Phase 9.0: Support Settings System
- ✅ **BONUS**: Simple Sync now v2-only (one-click from admin UI!)

---

## 🔄 Pending Phases

### **Phase 9: Data Migration & Multi-Sub Cleanup**
**Status**: Pending (but Simple Sync covers most of this!)

**What's Left**:
1. ✅ V2 tables populated (via Simple Sync)
2. ✅ Customers linked to users (via Simple Sync)
3. ⏳ Identify users with multiple active subscriptions
4. ⏳ Admin tool to consolidate multiple subs (or let webhook auto-cancel)

**Options**:
- **Option A**: Let it happen naturally via webhooks
  - When new subscription created → auto-cancels old ones
  - Gradual migration as users interact with system
  
- **Option B**: Proactive cleanup
  - Build admin report showing multi-sub users
  - Admin manually consolidates via support contact
  - Or build "bulk consolidate" tool

**Recommendation**: **Option A** (passive) - the webhook system will handle it automatically over time.

---

### **Phase 10: Cut Over to V2 Exclusively**
**Status**: In Progress! 🚀

**Already Done**:
- ✅ Simple Sync uses v2 only
- ✅ Webhooks write to v2 (dual-write mode)
- ✅ User dashboard uses v2 (`/user/subscriptions`)
- ✅ V2 elastic service ready

**What's Left**:
1. ⏳ Update v1 elastic service routes to use v2 service instead
2. ⏳ Update admin dashboards to use v2 data
3. ⏳ Monitor for 48 hours
4. ⏳ Archive/rename v1 tables

**Action Items**:
- [ ] Point admin subscriber list to v2 elastic service
- [ ] Point admin analytics to v2 tables
- [ ] Update any remaining v1 API calls to v2
- [ ] Test all admin features work with v2
- [ ] Monitor for 48 hours
- [ ] Rename v1 tables to `_v1_archived` or similar

---

### **Phase 5.6 & 6.5: Testing**
**Status**: Pending

**What to Test**:
1. ✅ Simple Sync (TESTED - working!)
2. ⏳ Stripe webhook flow with real events
   - Create test subscription in Stripe dashboard
   - Verify webhook fires and updates v2 tables
   - Verify customer linking happens automatically
   
3. ⏳ Single subscription enforcement
   - Create subscription for user with existing sub
   - Verify old sub gets auto-canceled
   - Verify video access transfers correctly

**How to Test**:
```bash
# 1. Create test subscription in Stripe dashboard
# 2. Watch backend logs for webhook events
# 3. Check v2 tables:
SELECT * FROM stripe_subscriptions_v2 WHERE stripe_id = 'sub_XXX';
SELECT * FROM user_stripe_customers_v2 WHERE user_id = XXX;

# 4. Verify user dashboard shows subscription
# Visit: /user/subscriptions
```

---

## 🎯 Recommended Next Steps

### **Quick Wins** (Do Now)
1. **Test Webhooks** - Create a test subscription and verify webhooks work
2. **Point Admin to V2** - Update admin routes to use v2 elastic service
3. **Monitor** - Check logs for any v2-related errors

### **Medium Priority** (This Week)
1. **Update All Admin Dashboards** - Ensure all admin features use v2
2. **Build Multi-Sub Report** - Admin view of users with multiple subs
3. **Testing** - Comprehensive webhook and subscription testing

### **Low Priority** (Later)
1. **Archive V1 Tables** - After 48 hours of v2-only operation
2. **Cleanup Old Code** - Remove unused v1 services
3. **Documentation** - Update docs to reflect v2 as primary

---

## 📊 Current System State

### **What's Using V2**
- ✅ Simple Sync (admin UI)
- ✅ Webhooks (dual-write to v1 + v2)
- ✅ User subscription dashboard (`/user/subscriptions`)
- ✅ Customer linking service
- ✅ Subscription manager service (video access)

### **What's Still Using V1**
- ⚠️ Admin subscriber list (elastic v1 service)
- ⚠️ Admin streaming dashboard
- ⚠️ Some analytics endpoints
- ⚠️ Legacy subscription routes

---

## 🚀 Phase 10: Full V2 Cutover Plan

### **Step 1: Point Admin to V2**
Update admin routes to use `SubscriberElasticServiceV2`:

```go
// In admin_streaming.go or similar
elasticServiceV2 := services.NewSubscriberElasticServiceV2(db)
RegisterSubscriberRoutes(admin, elasticServiceV2) // Use v2 instead of v1
```

### **Step 2: Update Frontend Admin Pages**
- Admin subscriber list → call v2 API
- Admin analytics → query v2 tables
- Admin reports → use v2 data

### **Step 3: Testing Checklist**
- [ ] Admin can view all subscribers
- [ ] Admin can filter/search subscribers
- [ ] Admin can view subscription details
- [ ] Analytics show correct MRR/ARR
- [ ] Reports generate successfully
- [ ] No errors in backend logs

### **Step 4: Monitor (48 Hours)**
- Watch backend logs for errors
- Check user reports/complaints
- Verify data accuracy
- Compare v1 vs v2 data (using comparison tool)

### **Step 5: Archive V1 Tables**
```sql
-- After 48 hours of successful v2 operation
ALTER TABLE stripe_customers RENAME TO stripe_customers_v1_archived;
ALTER TABLE stripe_subscriptions RENAME TO stripe_subscriptions_v1_archived;
ALTER TABLE subscription_plans RENAME TO subscription_plans_v1_archived;
-- etc.
```

---

## 🎯 Let's Start Phase 10!

Which would you like to tackle first?

**A. Point Admin Subscriber List to V2** (biggest impact)  
**B. Test Webhooks with Real Stripe Events** (validation)  
**C. Build Multi-Sub User Report** (cleanup)  

Let me know and I'll implement it! 🚀

