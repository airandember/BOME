# 🚀 Stripe V2 Sync Tool

CLI tool for syncing all Stripe data to v2 tables.

## 📋 What It Does

Syncs all data from Stripe API to your local v2 tables:
- ✅ Products (with metadata)
- ✅ Prices (with product FKs)
- ✅ Customers (with all details)
- ✅ Subscriptions (with customer & price FKs)

## 🏃 How to Run

### Prerequisites
1. Stripe v2 tables created (`050_create_stripe_v2_schema.sql`)
2. `STRIPE_SECRET_KEY` in your `.env` file
3. Database connection configured

### Run the Sync

```bash
cd backend
go run cmd/stripe-sync/main.go
```

## 📊 Expected Output

```
================================================================
🚀 STRIPE V2 SYNC - Syncing all data from Stripe API
================================================================

🚀 [Stripe Sync v2] Starting full sync...
📦 [Stripe Sync v2] Step 1/4: Syncing products...
📦 [Stripe Sync v2] Synced 10 products...
✅ [Stripe Sync v2] Synced 15 products
💰 [Stripe Sync v2] Step 2/4: Syncing prices...
💰 [Stripe Sync v2] Synced 10 prices...
✅ [Stripe Sync v2] Synced 20 prices
👥 [Stripe Sync v2] Step 3/4: Syncing customers...
👥 [Stripe Sync v2] Synced 100 customers...
👥 [Stripe Sync v2] Synced 200 customers...
✅ [Stripe Sync v2] Synced 300 customers
📋 [Stripe Sync v2] Step 4/4: Syncing subscriptions...
📋 [Stripe Sync v2] Synced 100 subscriptions...
✅ [Stripe Sync v2] Synced 150 subscriptions
✅ [Stripe Sync v2] Full sync complete!

================================================================
📊 SYNC SUMMARY
================================================================

⏱️  Duration: 45.2s

📦 Products:       15/15 (100.0%)
💰 Prices:         20/20 (100.0%)
👥 Customers:      300/300 (100.0%)
📋 Subscriptions:  150/150 (100.0%)
✅ Errors:         0

✅ Sync completed successfully!
```

## ⚠️ If Errors Occur

The sync will:
1. Continue even if some records fail
2. Report all errors at the end
3. Show which records failed
4. Exit with code 1 if any errors

Example error output:
```
⚠️  ERRORS ENCOUNTERED:
   1. Product prod_xxx: failed to sync - network timeout
   2. Customer cus_yyy: failed to sync - invalid email
```

## 🔄 Running Manually

You can also sync individual entities:

```bash
# Sync only products
curl -X POST http://localhost:8080/api/v1/admin/stripe-v2/sync-products \
  -H "Authorization: Bearer YOUR_TOKEN"

# Sync only prices
curl -X POST http://localhost:8080/api/v1/admin/stripe-v2/sync-prices \
  -H "Authorization: Bearer YOUR_TOKEN"

# Sync only customers
curl -X POST http://localhost:8080/api/v1/admin/stripe-v2/sync-customers \
  -H "Authorization: Bearer YOUR_TOKEN"

# Sync only subscriptions
curl -X POST http://localhost:8080/api/v1/admin/stripe-v2/sync-subscriptions \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 🎯 Next Steps

After running the sync:

1. **Verify data:**
   ```sql
   SELECT COUNT(*) FROM stripe_products_v2;
   SELECT COUNT(*) FROM stripe_prices_v2;
   SELECT COUNT(*) FROM stripe_customers_v2;
   SELECT COUNT(*) FROM stripe_subscriptions_v2;
   ```

2. **Check sync status:**
   ```bash
   curl http://localhost:8080/api/v1/admin/stripe-v2/status \
     -H "Authorization: Bearer YOUR_TOKEN"
   ```

3. **Link users to customers** (Phase 3 - coming next!)

## 🔍 Troubleshooting

### "STRIPE_SECRET_KEY not set"
- Add to `.env`: `STRIPE_SECRET_KEY=sk_test_...`

### "Failed to connect to database"
- Check database is running
- Verify connection string in `.env`

### "Sync times out"
- Stripe API rate limits (100 requests/sec)
- Large datasets may take several minutes
- This is normal for initial sync

### "Foreign key violation"
- Products must sync before prices
- Customers must sync before subscriptions
- The tool handles this automatically

