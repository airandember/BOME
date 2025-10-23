# 🧪 Stripe Sync System Testing Guide

## 🚀 Quick Start Testing

### 1. **System Status Check**
```bash
curl http://localhost:8080/api/v1/admin/streaming/stripe/test/status
```
**What to expect:** System overview with Stripe connection status, database readiness, and scheduled sync times.

### 2. **Test Stripe Connection**
```bash
curl http://localhost:8080/api/v1/admin/streaming/stripe/test/connection
```
**What to expect:** Confirms your Stripe API key works and can fetch account balance.

### 3. **Check Database Tables**
```bash
curl http://localhost:8080/api/v1/admin/streaming/stripe/test/tables
```
**What to expect:** Shows record counts in all Stripe tables (should be 0 initially).

### 4. **Trigger Initial Sync**
```bash
curl -X POST http://localhost:8080/api/v1/admin/streaming/stripe/sync/initial
```
**What to expect:** Starts 1.5-year historical sync. Check terminal logs for progress.

### 5. **Monitor Sync Progress**
```bash
curl http://localhost:8080/api/v1/admin/streaming/stripe/sync/status
```
**What to expect:** Real-time progress of sync operations.

---

## 📊 What to Watch in Terminal Logs

### **Sync Start Logs**
```
🚀 [SYNC] INITIAL_SYNC START: product | Items: 6 | Time: 14:30:15
📡 [SYNC] API CALL: product.List | Params: map[limit:100] | Duration: 245ms
💾 [SYNC] DB INSERT: stripe_products | Affected: 3 | Duration: 12ms
📊 [SYNC] PROGRESS: product | 3/6 (50.0%) | Duration: 1.2s
✅ [SYNC] PRODUCT COMPLETE | Processed: 6 | Duration: 2.1s | Errors: 0
```

### **Webhook Logs** (when you set them up)
```
📨 [WEBHOOK] WEBHOOK: customer.created | Object: cus_abc123 | Time: 14:35:22
🔄 [WEBHOOK] Webhook sync: Updating customer cus_abc123
💾 [WEBHOOK] DB UPSERT: stripe_customers | Affected: 1 | Duration: 8ms
```

### **Cron Job Logs**
```
⏰ [CRON] CRON: quarterly_sync scheduled for 2024-01-01 00:00:00 MST
⏰ [CRON] CRON: daily_incremental scheduled for 2024-08-29 02:00:00 MST
```

---

## 🔍 Detailed Testing Steps

### **Phase 1: Pre-Sync Validation**

1. **Check System Status**
   ```bash
   curl http://localhost:8080/api/v1/admin/streaming/stripe/test/status
   ```
   ✅ Verify: `stripe.enabled: true`, `key_type: "test"` or `"live"`

2. **Verify Empty Tables**
   ```bash
   curl http://localhost:8080/api/v1/admin/streaming/stripe/test/tables
   ```
   ✅ Verify: All tables show `count: 0`

3. **Test Stripe Connection**
   ```bash
   curl http://localhost:8080/api/v1/admin/streaming/stripe/test/connection
   ```
   ✅ Verify: `success: true`, reasonable duration

### **Phase 2: Initial Sync Testing**

4. **Start Initial Sync**
   ```bash
   curl -X POST http://localhost:8080/api/v1/admin/streaming/stripe/sync/initial
   ```
   ✅ Verify: Returns `202 Accepted` with sync started message

5. **Monitor Progress** (run multiple times)
   ```bash
   curl http://localhost:8080/api/v1/admin/streaming/stripe/sync/status
   ```
   ✅ Watch: `sync_running: true`, increasing `processed_items`

6. **Check Terminal Logs**
   ✅ Look for: Sync start messages, API calls, database operations, progress updates

### **Phase 3: Post-Sync Validation**

7. **Verify Data Populated**
   ```bash
   curl http://localhost:8080/api/v1/admin/streaming/stripe/test/tables
   ```
   ✅ Verify: Tables now show data counts > 0

8. **Check Sample Data**
   ```bash
   curl http://localhost:8080/api/v1/admin/streaming/stripe/test/sample
   ```
   ✅ Verify: Returns actual Stripe data with proper IDs

9. **Validate Data Integrity**
   ```bash
   curl http://localhost:8080/api/v1/admin/streaming/stripe/test/validate
   ```
   ✅ Verify: All checks show "✅ PASS" status

### **Phase 4: Webhook Testing** (Optional)

10. **Set up Stripe Webhook** (in Stripe Dashboard)
    - URL: `https://yourdomain.com/api/v1/admin/streaming/stripe/webhooks/`
    - Events: `customer.*`, `customer.subscription.*`, `invoice.payment_*`, `product.*`, `price.*`

11. **Test Webhook** (create a test customer in Stripe Dashboard)
    ✅ Watch terminal for: `📨 WEBHOOK` messages

---

## 🎯 Expected Results

### **After Initial Sync:**
- **Products**: Your Stripe products synced
- **Prices**: All pricing tiers synced  
- **Customers**: Last 1.5 years of customers
- **Subscriptions**: Active and historical subscriptions
- **Invoices**: Payment history
- **Coupons**: Discount codes

### **Performance Expectations:**
- **Small accounts** (< 100 items): 30-60 seconds
- **Medium accounts** (100-1000 items): 2-5 minutes  
- **Large accounts** (1000+ items): 5-15 minutes

### **Log Volume:**
- Expect 50-200 log lines for small accounts
- Batch processing pauses every 50 items
- Rate limiting pauses every 100 items

---

## 🚨 Troubleshooting

### **Common Issues:**

1. **"Stripe service is not enabled"**
   - Check your `STRIPE_SECRET_KEY` environment variable
   - Verify key starts with `sk_test_` or `sk_live_`

2. **"Database connection failed"**
   - Verify your database is running
   - Check database connection string

3. **"Failed to connect to Stripe"**
   - Check internet connection
   - Verify API key is valid
   - Check Stripe API status

4. **Sync appears stuck**
   - Check terminal logs for error messages
   - Large accounts may take time (this is normal)
   - Rate limiting causes intentional pauses

### **Reset and Retry:**
```sql
-- Clear all Stripe data to start fresh
TRUNCATE stripe_customers, stripe_products, stripe_prices, 
         stripe_subscriptions, stripe_invoices, stripe_coupons, 
         stripe_sync_jobs CASCADE;
```

---

## 📋 Test Checklist

- [ ] System status shows Stripe enabled
- [ ] Database connection test passes  
- [ ] Initial sync starts successfully
- [ ] Terminal shows sync progress logs
- [ ] Tables populate with data
- [ ] Sample data looks correct
- [ ] Data validation passes
- [ ] Webhooks receive events (if configured)
- [ ] Cron schedules are set correctly

**🎉 Success Criteria:** All tables populated, no validation errors, logs show successful operations!
