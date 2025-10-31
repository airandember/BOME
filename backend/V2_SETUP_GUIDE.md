# V2 Tables Setup Guide

## 🎯 Goal
Populate the v2 Stripe tables so users can see their subscriptions in `/user/subscriptions`

---

## ✅ Step 1: Run Database Migration (1 minute)

**Copy and execute this file in your database:**
```
backend/migrations/050_1_alter_user_stripe_customers_v2.sql
```

**Expected output:**
```
NOTICE: ✅ Renamed linked_at to first_linked_at (or Added first_linked_at column)
NOTICE: ✅ Added last_synced_at column
NOTICE: ✅ Set default for linked_by
NOTICE: ✅ Migration successful! user_stripe_customers_v2 table updated.
```

---

## 🚀 Step 2: Populate V2 Tables (5-10 minutes)

### Option A: One-Click Script (Recommended)
```powershell
cd S:\AirEmber\BOME\BOME\backend
.\populate-v2-tables.ps1
```

This will:
1. Wait for you to confirm migration is done
2. Sync all Stripe data (products, prices, customers, subscriptions)
3. Link users to their Stripe customers by email
4. Show a summary

### Option B: Manual Steps
If you prefer to run each step manually:

#### 2a. Sync Stripe Data
```powershell
cd cmd/stripe-sync
.\stripe-sync.exe
cd ../..
```

Expected output:
```
📦 Syncing products...
✅ Synced 5 products

💰 Syncing prices...
✅ Synced 12 prices

👥 Syncing customers...
✅ Synced 150 customers

📋 Syncing subscriptions...
✅ Synced 145 subscriptions
```

#### 2b. Link Users to Customers
```powershell
cd cmd/customer-linking
.\customer-linking.exe --link-all --pretty
cd ../..
```

Expected output:
```
🔗 Linking all users to Stripe customers...
✅ Linked 148 users to 150 customers
```

---

## ✅ Step 3: Verify

### Check User 4826
Visit: http://localhost:5173/user/subscriptions (logged in as aaalifesolutions@gmail.com)

**Before**: "ℹ️ User 4826 has no linked customers"  
**After**: Should show their active subscription!

### API Verification (Optional)
```bash
curl http://localhost:8080/api/v1/admin/customer-linking/user/4826 \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

Should return:
```json
{
  "user_id": 4826,
  "email": "aaalifesolutions@gmail.com",
  "linked_customers": [
    {
      "stripe_customer_id": "cus_xxxxx",
      "is_primary": true,
      "subscriptions_count": 1
    }
  ]
}
```

---

## 🎉 Success!

Once complete, you should have:
- ✅ V2 tables populated with Stripe data
- ✅ Users linked to their Stripe customers
- ✅ User subscription dashboard working
- ✅ Ready for Phase 9 (data migration and cleanup)

---

## 🐛 Troubleshooting

### Issue: "User has no linked customers"
- Check that Step 1 (migration) completed successfully
- Check that Step 2b (customer linking) ran without errors
- Verify user email matches their Stripe customer email

### Issue: "Stripe sync failed"
- Check your Stripe API key in `.env` or `secure_settings` table
- Verify internet connection
- Check backend logs for detailed error

### Issue: "Customer linking failed"
- Ensure Step 2a (Stripe sync) completed first
- Check backend logs for detailed error

---

## 📊 What Gets Synced

### Products & Prices
- All active Stripe products
- All active and archived prices
- Stored in `stripe_products_v2` and `stripe_prices_v2`

### Customers
- All Stripe customers (active and deleted)
- Metadata and billing details
- Stored in `stripe_customers_v2`

### Subscriptions
- All subscriptions (active, canceled, trialing, etc.)
- Subscription items and billing periods
- Stored in `stripe_subscriptions_v2`

### User Links
- Email-based matching between users and Stripe customers
- Primary customer designation
- Stored in `user_stripe_customers_v2`

---

## 🔄 Re-running

You can safely re-run the sync and linking commands multiple times:
- Sync will **upsert** (update or insert) data
- Linking will **skip** already-linked users
- No duplicate data will be created

---

## 📝 Next Steps After V2 Works

Once v2 tables are working:
1. **Phase 9**: Migrate existing data and fix multi-subscription users
2. **Phase 10**: Cut over to v2 exclusively, archive v1 tables
3. Monitor for 48 hours
4. Celebrate! 🎉

