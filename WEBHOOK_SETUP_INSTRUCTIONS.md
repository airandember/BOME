# 🚀 Stripe Webhook Setup - Quick Instructions

**Goal:** Configure both Snapshot and Thin webhooks in 10 minutes  
**Endpoint URL:** `https://watch.bookofmormonevidence.org/bome-backend/api/v1/webhooks/stripe`

---

## 📋 **Setup Checklist:**

### **Step 1: Create Snapshot Webhook (5 min)**

1. **Go to:** https://dashboard.stripe.com/webhooks
2. **Click:** "Add endpoint"
3. **Configure:**
   ```
   Name: BOME - BETA (Snapshot)
   URL: https://watch.bookofmormonevidence.org/bome-backend/api/v1/webhooks/stripe
   Payload: Snapshot
   API Version: 2025-10-29.clover (Latest)
   ```
4. **Select Events:**
   - ✅ `customer.*` (all customer events)
   - ✅ `customer.subscription.*` (all subscription events)
   - ✅ `invoice.payment_succeeded`
   - ✅ `invoice.payment_failed`
   - ✅ `checkout.session.*` (all checkout events)
5. **Click:** "Add endpoint"
6. **Copy the signing secret** (`whsec_xxxxx`) → Save it somewhere safe!

---

### **Step 2: Create Thin Webhook (5 min)**

1. **Still in:** https://dashboard.stripe.com/webhooks
2. **Click:** "Add endpoint" again
3. **Configure:**
   ```
   Name: BOME - BETA (Thin)
   URL: https://watch.bookofmormonevidence.org/bome-backend/api/v1/webhooks/stripe
   Payload: Thin
   API Version: Unversioned
   ```
4. **Select Events:**
   - ✅ `v2.billing.subscription.created`
   - ✅ `v2.billing.subscription.updated`
   - ✅ `v2.billing.subscription.paused`
   - ✅ `v2.billing.subscription.resumed`
   - ✅ `v2.core.event_destination.ping`
5. **Click:** "Add endpoint"
6. **Copy the signing secret** (`whsec_yyyyy`) → This will be DIFFERENT from the first one!

---

### **Step 3: Add Secrets to Your Admin UI (2 min)**

1. **Go to:** `http://localhost:5173/admin/streaming/stripe` (or your frontend URL)
2. **Click:** "Setup" tab
3. **Scroll to:** "Webhook Secret - Snapshot (V1)"
   - **Paste:** The first `whsec_xxxxx` secret (from Step 1)
   - **Click:** "Save Snapshot Secret"
   - **See:** ✅ "Snapshot webhook secret saved successfully (encrypted)!"

4. **Scroll to:** "Webhook Secret - Thin (V2)"
   - **Paste:** The second `whsec_yyyyy` secret (from Step 2)
   - **Click:** "Save Thin Secret"
   - **See:** ✅ "Thin webhook secret saved successfully (encrypted)!"

---

### **Step 4: Test Both Webhooks (2 min)**

**Test Snapshot:**
1. Go back to Stripe Dashboard → Webhooks → Your Snapshot endpoint
2. Click "Send test webhook"
3. Select: `customer.created`
4. Click "Send test webhook"
5. ✅ Should see: HTTP 200, "received": true, "dual_write": "v1+v2"

**Test Thin:**
1. Go to Stripe Dashboard → Webhooks → Your Thin endpoint
2. Click "Send test webhook"
3. Select: `v2.core.event_destination.ping`
4. Click "Send test webhook"
5. ✅ Should see: HTTP 200, "received": true, "type": "v2_thin_event"

---

### **Step 5: Verify in Admin Dashboard (1 min)**

1. **Go to:** `/admin/streaming/stripe`
2. **Check "Webhook Health" panel:**
   - ✅ Status: "Healthy" (green)
   - ✅ Events Today: 2+ (from your tests)
   - ✅ Last Event: < 5 minutes ago
   - ✅ Success Rate: 100%

---

## ✅ **You're Done!**

Your webhooks are now:
- ✅ **Configured** in Stripe (both destinations)
- ✅ **Secured** with encrypted secrets
- ✅ **Validated** (both tests passed)
- ✅ **Monitored** (admin dashboard shows health)

---

## 🎯 **What Happens Now:**

### **When a Customer Subscribes:**

```
Customer clicks "Subscribe" on your site
    ↓
Stripe processes payment
    ↓
Stripe sends webhook: customer.subscription.created
    ↓
Your backend receives it (validated with snapshot secret)
    ↓
Dual-write to v1 + v2 tables
    ↓
User is auto-linked to customer by email
    ↓
Video access granted
    ↓
Admin dashboard updated
```

### **When Stripe Sends V2 Events:**

```
Stripe sends webhook: v2.billing.subscription.updated
    ↓
Your backend receives it (validated with thin secret)
    ↓
Fetches full subscription from Stripe API
    ↓
Updates v2 tables
    ↓
Video access updated
    ↓
Admin dashboard updated
```

---

## 🔍 **Troubleshooting:**

### **Problem: "Invalid signature" in logs**
```
✅ Fix: Check you copied the full whsec_ secret
✅ Fix: Make sure you used the RIGHT secret for each destination
      (Snapshot secret → Snapshot field, Thin secret → Thin field)
✅ Fix: Try deleting and recreating the webhook in Stripe
```

### **Problem: "Webhook secret not configured"**
```
✅ Fix: Make sure you saved BOTH secrets in admin UI
✅ Fix: Refresh the page and check they're showing as "configured"
✅ Fix: Restart backend to reload secrets from database
```

### **Problem: No events showing in admin dashboard**
```
✅ Fix: Send test webhooks from Stripe Dashboard
✅ Fix: Check backend logs for errors
✅ Fix: Verify URL is correct (no typos)
✅ Fix: Check firewall allows Stripe IPs
```

---

## 📊 **Quick Reference:**

| Setting | Value |
|---------|-------|
| **Endpoint URL** | `https://watch.bookofmormonevidence.org/bome-backend/api/v1/webhooks/stripe` |
| **Snapshot API Version** | 2025-10-29.clover |
| **Thin API Version** | Unversioned |
| **Snapshot Secret Location** | `secure_settings.stripe_webhook_secret` |
| **Thin Secret Location** | `secure_settings.stripe_webhook_secret_thin` |
| **Admin UI Path** | `/admin/streaming/stripe` (Setup tab) |
| **Encryption** | AES-GCM with master key |
| **Test Endpoint** | Stripe Dashboard → Webhooks → Send test webhook |

---

## 🎉 **Success Indicators:**

You'll know it's working when you see:

1. ✅ **In Stripe Dashboard:**
   - Both webhook endpoints show "Active"
   - Test webhooks return HTTP 200
   - Event history shows successful deliveries

2. ✅ **In Your Admin UI:**
   - Webhook Health: "Healthy" (green)
   - Events Today: > 0
   - Last Event: Recent timestamp
   - Success Rate: 100%

3. ✅ **In Backend Logs:**
   - `✅ Loaded snapshot webhook secret from database`
   - `✅ Loaded thin webhook secret from database`
   - `📨 Webhook received: customer.created`
   - `✅ Webhook: Successfully processed v1 event`

4. ✅ **In Database:**
   - `stripe_customers_v2` table has new rows
   - `stripe_subscriptions_v2` table has new rows
   - `user_stripe_customers_v2` shows user-customer links
   - `webhook_events` table logs all events

---

**That's it! Your webhooks are live and processing events! 🚀**

---

**Questions?**
- Check `WEBHOOK_DUAL_SECRET_COMPLETE.md` for technical details
- Check backend logs for debugging
- Check admin dashboard for real-time status

