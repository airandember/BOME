# 🎯 Stripe Webhook Setup Guide - THIN vs SNAPSHOT

**Date:** October 31, 2025  
**Goal:** Configure Stripe webhooks for perfect real-time sync  
**Decision:** Choose between THIN and SNAPSHOT payloads

---

## 📊 **THIN vs SNAPSHOT: Quick Comparison**

| Factor | THIN Payloads | SNAPSHOT Payloads |
|--------|---------------|-------------------|
| **Size** | Smaller (~1-2KB) | Larger (~5-20KB) |
| **Speed** | Faster delivery | Slightly slower |
| **Reliability** | ⚠️ Requires API call | ✅ All data included |
| **Cost** | More API calls | Fewer API calls |
| **Complexity** | More code | Less code |
| **Best For** | High-volume, simple events | Critical data, reliability |

---

## 🎯 **RECOMMENDATION FOR BOME: SNAPSHOT** ✅

### **Why SNAPSHOT is Better for You:**

1. **✅ Reliability First**
   - All subscription data in one payload
   - No extra API calls needed
   - Less chance of missing data

2. **✅ Your Current Code**
   - Already expects full objects
   - V2 services process complete data
   - Dual-write needs all fields

3. **✅ Lower Volume**
   - ~2,531 subscribers (not millions)
   - Webhook volume is manageable
   - Network overhead is negligible

4. **✅ Simpler Code**
   - No need to fetch missing data
   - Less error handling
   - Faster processing

5. **✅ Better Debugging**
   - Full payload in logs
   - Can replay events easily
   - Complete audit trail

---

## 🚀 **CURRENT WEBHOOK STATUS:**

### **Your Endpoints:**

✅ **Public Endpoint (Live):**
```
https://watch.bookofmormonevidence.org/bome-backend/api/v1/webhooks/stripe
```

✅ **Admin Test Endpoint:**
```
https://watch.bookofmormonevidence.org/bome-backend/api/v1/admin/stripe/webhooks
```

### **What's Working:**

✅ Signature validation  
✅ V1 and V2 event handling  
✅ Dual-write to both v1 and v2 tables  
✅ Customer linking on creation  
✅ Webhook logging to database  
✅ Admin dashboard monitoring  

### **Webhook Secret Storage:**

✅ Stored in `secure_settings` table (encrypted)  
✅ AES-GCM encryption with master key  
✅ Never exposed to frontend  
✅ Configurable via admin UI  

---

## 📋 **STRIPE WEBHOOK CONFIGURATION:**

### **Step 1: Access Stripe Dashboard**

1. Go to: https://dashboard.stripe.com/webhooks
2. Click "Add endpoint"

### **Step 2: Configure Endpoint**

**URL to add:**
```
https://watch.bookofmormonevidence.org/bome-backend/api/v1/webhooks/stripe
```

**API Version:**
- Use: `Latest API version` (or `2024-10-28.acacia` if available)
- Note: V2 events require 2024-06-20 or later

**Description:**
```
BOME Production - Subscription & Customer Sync
```

### **Step 3: Select Events (SNAPSHOT MODE)**

#### **✅ Core Events (Must Have):**

**Customer Events:**
- [x] `customer.created`
- [x] `customer.updated`
- [x] `customer.deleted`

**Subscription Events:**
- [x] `customer.subscription.created`
- [x] `customer.subscription.updated`
- [x] `customer.subscription.deleted`
- [x] `customer.subscription.paused`
- [x] `customer.subscription.resumed`

**Invoice Events:**
- [x] `invoice.payment_succeeded`
- [x] `invoice.payment_failed`
- [x] `invoice.finalized`

**Payment Events:**
- [x] `payment_intent.succeeded`
- [x] `payment_intent.payment_failed`

**Checkout Events:**
- [x] `checkout.session.completed`
- [x] `checkout.session.expired`

#### **🔮 Optional Events (Nice to Have):**

**Product/Price Events:**
- [ ] `product.created`
- [ ] `product.updated`
- [ ] `product.deleted`
- [ ] `price.created`
- [ ] `price.updated`

**Billing Portal:**
- [ ] `billing_portal.session.created`

### **Step 4: API Version Settings**

**Important:** For V2 events (thin payloads), select:
- ✅ **"API version 2024-10-28.acacia"** or later
- This enables: `v2.core.event_destination.*` events

**For V1 events (snapshot - RECOMMENDED):**
- ✅ **"Latest API version"** 
- This gives you full snapshot payloads

### **Step 5: Get Signing Secret**

1. After creating the endpoint, Stripe shows: `whsec_xxxxx`
2. **Copy this secret immediately** (shown only once)
3. Save it securely (you'll need it in Step 6)

---

## 🔧 **CONFIGURE YOUR BACKEND:**

### **Step 6: Add Webhook Secret to Admin UI**

**Option A: Via Admin Dashboard (Recommended)**

1. Go to: `/admin/streaming/stripe/setup`
2. Scroll to "Webhook Configuration"
3. Paste your `whsec_xxxxx` secret
4. Click "Save Webhook Secret"
5. ✅ Secret is encrypted and stored in `secure_settings`

**Option B: Via Database**

```sql
-- Get crypto service encryption key first
-- Then manually insert (not recommended - use admin UI)
```

**Option C: Via Environment Variable (Fallback)**

Add to `.env`:
```bash
STRIPE_WEBHOOK_SECRET=whsec_xxxxx
```

---

## 🧪 **TEST YOUR WEBHOOKS:**

### **Step 7: Test from Stripe Dashboard**

1. In Stripe → Webhooks → Your endpoint
2. Click "Send test webhook"
3. Select: `customer.created`
4. Click "Send test webhook"

**Expected Response:**
```json
{
  "received": true,
  "processed": true,
  "type": "v1_event",
  "dual_write": "v1+v2"
}
```

### **Step 8: Check Admin Dashboard**

1. Go to: `/admin/streaming/stripe`
2. Scroll to "Webhook Health"
3. **Verify:**
   - ✅ Status: "Healthy"
   - ✅ Last Event: < 5 minutes ago
   - ✅ Events Today: > 0
   - ✅ Success Rate: 100%

### **Step 9: Check Database Logs**

```sql
-- Check webhook events table
SELECT 
    event_type,
    status,
    response_time,
    created_at
FROM webhook_events
ORDER BY created_at DESC
LIMIT 10;

-- Check if customer was synced to v2
SELECT * FROM stripe_customers_v2
ORDER BY created_at DESC
LIMIT 5;
```

---

## 🎯 **WEBHOOK EVENTS YOUR SYSTEM HANDLES:**

### **V1 Events (Snapshot Payloads) - ACTIVE:**

```javascript
// Customer Events
✅ customer.created          → Dual-write v1+v2, auto-link user by email
✅ customer.updated          → Dual-write v1+v2
✅ customer.deleted          → Dual-write v1+v2, mark deleted

// Subscription Events
✅ customer.subscription.created   → Dual-write v1+v2, grant video access
✅ customer.subscription.updated   → Dual-write v1+v2, update video access
✅ customer.subscription.deleted   → Dual-write v1+v2, revoke video access

// Invoice Events
✅ invoice.payment_succeeded       → Update video access
✅ invoice.payment_failed          → Revoke video access

// Checkout Events
✅ checkout.session.completed      → Process new subscription
✅ checkout.session.expired        → Clean up abandoned sessions

// Product/Price Events (from manual sync)
✅ product.created/updated         → Via Simple Sync
✅ price.created/updated           → Via Simple Sync
```

### **V2 Events (Thin Payloads) - READY:**

```javascript
// System Events
✅ v2.core.event_destination.ping  → Health check
⏳ v2.billing.subscription.created → Not implemented yet
⏳ v2.billing.subscription.updated → Not implemented yet
```

---

## 🛡️ **SECURITY FEATURES:**

### **Your Webhook Security Stack:**

1. **✅ Signature Verification**
   - Every webhook validated with `stripe.ConstructEvent()`
   - Invalid signatures = automatic 400 rejection

2. **✅ Replay Attack Prevention**
   - Stripe timestamps in signatures
   - 5-minute tolerance window

3. **✅ Encrypted Secret Storage**
   - AES-GCM encryption at rest
   - Master key from environment
   - Never exposed via API

4. **✅ Audit Trail**
   - All events logged to `webhook_events` table
   - Includes: payload size, response time, status
   - Retention: 90 days (configurable)

5. **✅ Error Handling**
   - Failed events logged with error details
   - Manual retry available via admin UI
   - Automatic failure notifications

---

## 📊 **MONITORING & DEBUGGING:**

### **Admin Dashboard Features:**

**Webhook Health Panel:**
- 🟢 Status indicator (Healthy/Warning/Error)
- 📊 Events today / total events
- ✅ Success rate percentage
- 🔔 Last event timestamp
- 📈 Event type breakdown

**Webhook Logs:**
- Last 100 events
- Filter by: event type, status, date
- View full payload
- Retry failed events

**Webhook Testing:**
- Manual ping test
- Send test events
- View live responses

### **Log Files:**

```bash
# Check backend logs for webhook processing
tail -f logs/webhook.log

# Look for these patterns:
"📨 Webhook received: customer.created"
"✅ Webhook: Successfully processed v1 event"
"✅ Dual-write to v1 + v2 complete"
```

---

## 🎯 **SNAPSHOT vs THIN: Technical Details**

### **SNAPSHOT Payload Example:**

```json
{
  "id": "evt_xxxxx",
  "type": "customer.subscription.created",
  "data": {
    "object": {
      "id": "sub_xxxxx",
      "customer": "cus_xxxxx",
      "status": "active",
      "items": {
        "data": [{
          "price": {
            "id": "price_xxxxx",
            "unit_amount": 1000,
            "currency": "usd",
            "recurring": {
              "interval": "month"
            },
            "product": {
              "id": "prod_xxxxx",
              "name": "Premium Plan",
              "metadata": {}
            }
          }
        }]
      },
      "metadata": {},
      "current_period_start": 1698764400,
      "current_period_end": 1701356400
      // ... full subscription object (~5KB)
    }
  }
}
```

**Pros:**
- ✅ All data in one place
- ✅ No API calls needed
- ✅ Your code works as-is

### **THIN Payload Example:**

```json
{
  "id": "evt_xxxxx",
  "type": "v2.billing.subscription.created",
  "data": {
    "object": {
      "id": "sub_xxxxx",
      "customer": "cus_xxxxx",
      "status": "active"
      // ... only basic fields (~1KB)
    }
  },
  "related_object": {
    "id": "cus_xxxxx",
    "type": "customer",
    "url": "/v1/customers/cus_xxxxx"
  }
}
```

**Pros:**
- ✅ Faster delivery
- ✅ Smaller bandwidth

**Cons:**
- ❌ Need to fetch full object via API
- ❌ More code complexity
- ❌ Extra API calls count toward rate limits

---

## 🎯 **FINAL RECOMMENDATION:**

### **Use SNAPSHOT (V1 Events)** ✅

**Configure Stripe to send:**
- ✅ All V1 events (snapshot mode)
- ✅ Latest API version
- ✅ Events listed in Step 3

**Your system is READY for:**
- ✅ Real-time customer sync
- ✅ Real-time subscription sync
- ✅ Automatic user linking
- ✅ Video access management
- ✅ Dual-write to v1+v2

---

## 🚀 **NEXT STEPS:**

1. **Configure Stripe Webhook** (Step 1-5)
2. **Add Signing Secret** (Step 6)
3. **Test with Stripe Dashboard** (Step 7-8)
4. **Verify in Admin UI** (Step 9)
5. **Monitor for 24 hours**
6. **Confirm all events processed**

---

## 💡 **WHY NOT THIN (V2) EVENTS?**

While V2 thin events are newer and more efficient, they're **not necessary** for your use case:

1. **Low Volume**: 2,531 users = minimal bandwidth savings
2. **Current Code**: Built for snapshot payloads
3. **Reliability**: Snapshot = no API dependency
4. **Simplicity**: Less code = fewer bugs
5. **Debugging**: Full payload = easier troubleshooting

**Future:** If you scale to 100K+ subscribers, revisit V2 thin events. For now, snapshot is perfect! ✅

---

## 📞 **SUPPORT:**

**If webhooks fail:**
1. Check admin dashboard: `/admin/streaming/stripe`
2. View webhook logs
3. Verify signing secret is configured
4. Test with Stripe "Send test webhook"
5. Check backend logs for errors

**Common Issues:**
- ❌ Invalid signature → Check `whsec_` secret
- ❌ Endpoint unreachable → Check firewall/DNS
- ❌ Events not processing → Check backend logs
- ❌ Duplicate events → Normal, handle idempotently

---

## ✅ **CHECKLIST:**

- [ ] Stripe webhook endpoint created
- [ ] Webhook secret saved in admin UI
- [ ] All core events selected (snapshot mode)
- [ ] Test webhook sent successfully
- [ ] Admin dashboard shows "Healthy"
- [ ] Database logs show recent events
- [ ] V2 tables receiving data
- [ ] Customer linking working
- [ ] Monitor for 24 hours

---

**Ready to configure? Let me know and I'll help you through each step!** 🎯

