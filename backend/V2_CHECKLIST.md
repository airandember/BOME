# V2 Tables Setup Checklist

Use this checklist to track your progress:

---

## 📋 Pre-Flight Checks

- [ ] Backend server is running (http://localhost:8080)
- [ ] Frontend dev server is running (http://localhost:5173)
- [ ] Database is accessible
- [ ] Stripe API keys are configured
- [ ] You have admin access to the database

---

## 🗄️ Database Migration

- [ ] Open database tool (pgAdmin, DBeaver, etc.)
- [ ] Copy contents of `backend/migrations/050_1_alter_user_stripe_customers_v2.sql`
- [ ] Execute the migration script
- [ ] Verify success messages:
  - [ ] `✅ Renamed linked_at to first_linked_at` (or Added)
  - [ ] `✅ Added last_synced_at column`
  - [ ] `✅ Set default for linked_by`
  - [ ] `✅ Migration successful!`

---

## 📦 Data Population

### Option A: One-Click Script

- [ ] Open PowerShell in `S:\AirEmber\BOME\BOME\backend`
- [ ] Run: `.\populate-v2-tables.ps1`
- [ ] Press Enter when prompted (after migration)
- [ ] Wait for Stripe sync to complete (5-10 min)
- [ ] Wait for customer linking to complete (1-2 min)
- [ ] Verify success message: `🎉 V2 Tables Population Complete!`

### Option B: Manual Steps

**Stripe Sync:**
- [ ] Open PowerShell in `S:\AirEmber\BOME\BOME\backend`
- [ ] Run: `cd cmd/stripe-sync`
- [ ] Run: `.\stripe-sync.exe`
- [ ] Wait for completion
- [ ] Verify products synced: `✅ Synced X products`
- [ ] Verify prices synced: `✅ Synced X prices`
- [ ] Verify customers synced: `✅ Synced X customers`
- [ ] Verify subscriptions synced: `✅ Synced X subscriptions`
- [ ] Run: `cd ../..` (back to backend directory)

**Customer Linking:**
- [ ] From backend directory, run: `cd cmd/customer-linking`
- [ ] Run: `.\customer-linking.exe --link-all --pretty`
- [ ] Wait for completion
- [ ] Verify users linked: `✅ Linked X users to X customers`
- [ ] Run: `cd ../..` (back to backend directory)

---

## ✅ Verification

### Frontend Test
- [ ] Open browser to http://localhost:5173
- [ ] Log in as `aaalifesolutions@gmail.com` (user 4826)
- [ ] Navigate to http://localhost:5173/user/subscriptions
- [ ] **VERIFY**: User sees their subscription (not "no linked customers")
- [ ] Check subscription details are displayed correctly

### Backend Logs Test
- [ ] Check backend console logs
- [ ] Look for: `📋 [User Subscriptions] Getting subscriptions for user 4826`
- [ ] **VERIFY**: Should NOT see "User 4826 has no linked customers"
- [ ] **VERIFY**: Should see subscription data being fetched

### API Test (Optional)
- [ ] Get admin auth token from browser DevTools
- [ ] Run curl command:
  ```bash
  curl http://localhost:8080/api/v1/admin/customer-linking/user/4826 \
    -H "Authorization: Bearer YOUR_TOKEN"
  ```
- [ ] **VERIFY**: Response shows linked customers with subscriptions

### Database Test (Optional)
- [ ] Open database tool
- [ ] Run query: `SELECT COUNT(*) FROM stripe_customers_v2;`
- [ ] **VERIFY**: Count > 0
- [ ] Run query: `SELECT COUNT(*) FROM stripe_subscriptions_v2;`
- [ ] **VERIFY**: Count > 0
- [ ] Run query: `SELECT COUNT(*) FROM user_stripe_customers_v2;`
- [ ] **VERIFY**: Count > 0
- [ ] Run query: 
  ```sql
  SELECT u.email, sc.stripe_customer_id, usc.is_primary 
  FROM users u
  JOIN user_stripe_customers_v2 usc ON u.id = usc.user_id
  JOIN stripe_customers_v2 sc ON usc.stripe_customer_id = sc.id
  WHERE u.id = 4826;
  ```
- [ ] **VERIFY**: Shows user 4826 linked to at least one Stripe customer

---

## 🎉 Success Criteria

All of these should be true:

- [x] ✅ Production build error fixed
- [ ] ✅ Migration executed successfully
- [ ] ✅ Stripe data synced to v2 tables
- [ ] ✅ Users linked to Stripe customers
- [ ] ✅ User 4826 sees subscription in dashboard
- [ ] ✅ No "User has no linked customers" error
- [ ] ✅ Subscription details display correctly

---

## 🐛 If Something Goes Wrong

### Migration Fails
- Check PostgreSQL version (need 9.5+)
- Verify you have ALTER TABLE permissions
- Check if table `user_stripe_customers_v2` exists

### Stripe Sync Fails
- Check `.env` file for `STRIPE_SECRET_KEY`
- Verify key starts with `sk_live_` or `sk_test_`
- Check internet connection
- Look at backend logs for detailed error

### Customer Linking Fails
- Ensure Stripe sync completed first
- Check that users table has email addresses
- Verify emails match between users and Stripe customers
- Look at backend logs for detailed error

### User Still Sees "No Linked Customers"
- Restart backend server
- Clear browser cache
- Re-run customer linking: `.\customer-linking.exe --link-all`
- Check database query in "Database Test" section above

---

## 📞 Need Help?

Check these files:
- `backend/V2_SETUP_GUIDE.md` - Detailed instructions
- `backend/QUICK_START_V2.md` - Quick reference
- `backend/PRODUCTION_BUILD_FIX.md` - Build error fix details
- Backend logs in console
- Database error messages

---

**Current Status:** Ready to begin! Start with "Database Migration" section above. ☝️

