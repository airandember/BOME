# 🚀 STRIPE V2 BRAID IMPLEMENTATION PLAN

**Date**: October 26, 2025  
**Status**: 🟢 **IN PROGRESS**  
**Strategy**: Build parallel system, then swap

---

## 📋 **IMPLEMENTATION CHECKLIST**

### **Phase 1: Database Schema** ✅
- [x] Create migration SQL file (`050_create_stripe_v2_schema.sql`)
- [ ] Run migration on development database
- [ ] Verify tables created successfully
- [ ] Test foreign key constraints

### **Phase 2: Stripe Sync Service** 🚧
- [ ] Create `StripeSync_v2Service` (Go)
- [ ] Implement product sync from Stripe API
- [ ] Implement price sync from Stripe API
- [ ] Implement customer sync from Stripe API
- [ ] Implement subscription sync from Stripe API
- [ ] Add error handling and retries

### **Phase 3: Customer Linking Service** 📝
- [ ] Create `CustomerLinkingService` (Go)
- [ ] Implement email-based linking
- [ ] Implement primary customer logic
- [ ] Add duplicate detection
- [ ] Add audit trail logging

### **Phase 4: Elastic Service v2** 📝
- [ ] Create `SubscriberElasticService_v2` (Go)
- [ ] Update queries to use v2 tables
- [ ] Test performance improvements
- [ ] Add v2 endpoints

### **Phase 5: Webhook Handlers v2** 📝
- [ ] Update `subscription.created` handler
- [ ] Update `subscription.updated` handler
- [ ] Update `subscription.deleted` handler
- [ ] Update `customer.created` handler
- [ ] Update `customer.updated` handler
- [ ] Test with Stripe test mode

### **Phase 6: Business Logic - Single Subscription** 📝
- [ ] Create function to cancel old subscriptions
- [ ] Implement in `subscription.created` webhook
- [ ] Add safety checks (don't cancel if same product)
- [ ] Add logging for canceled subscriptions

### **Phase 7: Frontend Updates** 📝
- [ ] Update subscriber dashboard to query v2
- [ ] Create subscription management modal
- [ ] Show active subscription details
- [ ] Add "Cancel Subscription" button
- [ ] Add "Change Plan" flow
- [ ] Update profile page subscription section

### **Phase 8: Parallel Testing** 📝
- [ ] Run v1 and v2 side-by-side
- [ ] Compare results (v1 vs v2)
- [ ] Log any discrepancies
- [ ] Monitor performance
- [ ] Test webhook flow end-to-end

### **Phase 9: Migration** 📝
- [ ] Sync all existing data to v2
- [ ] Link all users to customers
- [ ] Fix users with multiple active subscriptions
- [ ] Verify data integrity
- [ ] Run diagnostic queries

### **Phase 10: Cutover** 📝
- [ ] Update all code to use v2 exclusively
- [ ] Remove v1 references
- [ ] Monitor for 48 hours
- [ ] Archive v1 tables (rename to _deprecated)
- [ ] Schedule v1 table deletion (30 days)

---

## 🎯 **CURRENT TASK: Phase 1 - Database Schema**

### **What You Need to Do:**

1. **Run the migration SQL:**
   ```bash
   # Option A: Using migration tool (if you have one)
   cd backend
   go run cmd/migrate/main.go up
   
   # Option B: Direct SQL execution
   psql -h localhost -U bome_user -d bome_db -f migrations/050_create_stripe_v2_schema.sql
   
   # Option C: Via Docker (if database in container)
   docker exec -i bome-postgres psql -U bome_user -d bome_db < migrations/050_create_stripe_v2_schema.sql
   ```

2. **Verify tables created:**
   ```sql
   -- Check tables exist
   SELECT table_name FROM information_schema.tables 
   WHERE table_name LIKE 'stripe_%_v2' OR table_name = 'user_stripe_customers_v2';
   
   -- Check foreign keys
   SELECT
       tc.table_name, 
       kcu.column_name, 
       ccu.table_name AS foreign_table_name,
       ccu.column_name AS foreign_column_name 
   FROM information_schema.table_constraints AS tc 
   JOIN information_schema.key_column_usage AS kcu
       ON tc.constraint_name = kcu.constraint_name
   JOIN information_schema.constraint_column_usage AS ccu
       ON ccu.constraint_name = tc.constraint_name
   WHERE tc.constraint_type = 'FOREIGN KEY' 
       AND tc.table_name LIKE '%_v2';
   ```

3. **Report back with:**
   - ✅ Tables created successfully
   - ✅ Foreign keys working
   - ❌ Any errors encountered

---

## 🚀 **NEXT PHASE: Phase 2 - Stripe Sync Service**

Once you confirm Phase 1 is complete, I'll create:

1. **`backend/internal/services/stripe_sync_v2.go`**
   - Full Stripe API sync
   - Proper FK linking
   - Error handling
   - Progress reporting

2. **`backend/cmd/stripe-sync/main.go`**
   - CLI tool for manual sync
   - Progress bar
   - Summary report

3. **Admin endpoint for triggering sync**
   - `/admin/stripe/sync-v2`
   - Real-time progress updates
   - Sync status dashboard

---

## 📊 **BUSINESS RULES (CONFIRMED)**

### **Single Subscription Rule:**
- ✅ Users can only have **ONE active subscription** at a time
- ✅ When user subscribes, **cancel all other active subscriptions**
- ✅ Keep old subscriptions in database (for historical data)
- ✅ Exception: Don't cancel if same product (renewal scenario)

### **Subscription Management:**
- ✅ Profile page shows current subscription
- ✅ "Manage Subscription" button opens modal/accordion
- ✅ Modal shows:
  - Plan name
  - Price
  - Billing interval
  - Next billing date
  - Cancel button
  - Change plan button
- ✅ If no active subscription:
  - Show "Subscribe Now" link to `/subscriptions`

---

## 🎯 **SUCCESS CRITERIA**

### **Performance:**
- [ ] Subscriber dashboard loads in < 500ms (currently 2-5s)
- [ ] Single user lookup in < 10ms (currently 50-200ms)
- [ ] Webhook processing in < 500ms

### **Data Integrity:**
- [ ] Zero orphaned records
- [ ] All users linked to correct customers
- [ ] All subscriptions have valid FKs
- [ ] Only ONE primary customer per user

### **Business Logic:**
- [ ] Users with multiple subs reduced to ONE
- [ ] Webhooks create subscriptions correctly
- [ ] Webhooks link customers to users
- [ ] Old subscriptions canceled when new one created

---

## 📞 **READY FOR NEXT STEP?**

**Waiting for your confirmation that Phase 1 is complete!**

Once you run the migration SQL and report back, I'll immediately create:
- Stripe sync service
- Customer linking service
- Admin sync endpoint

Let me know how it goes! 🚀

