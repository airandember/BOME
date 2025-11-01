# 🚀 Webhook Quick Start - 5 Minutes to Live

**Goal:** Get Stripe webhooks running in 5 minutes  
**Recommendation:** SNAPSHOT payloads ✅

---

## ⚡ **SUPER QUICK SETUP:**

### **1️⃣ Create Endpoint in Stripe (2 min)**

Go to: https://dashboard.stripe.com/webhooks

```
URL: https://watch.bookofmormonevidence.org/bome-backend/api/v1/webhooks/stripe
Events: Select "customer.*" and "customer.subscription.*"
Version: Latest
```

### **2️⃣ Copy Secret (30 sec)**

After creating, copy the `whsec_xxxxx` secret

### **3️⃣ Add to Admin UI (1 min)**

Go to: `/admin/streaming/stripe/setup`

Paste secret in "Webhook Configuration" section

Click "Save"

### **4️⃣ Test (1 min)**

In Stripe Dashboard:
- Click "Send test webhook"
- Select `customer.created`
- Send

### **5️⃣ Verify (30 sec)**

Check: `/admin/streaming/stripe`

Should see:
- ✅ Status: Healthy
- ✅ Events Today: 1+
- ✅ Last Event: < 1 minute ago

---

## 🎯 **SNAPSHOT vs THIN - THE ANSWER:**

### **Use SNAPSHOT** ✅

**Why?**
- ✅ All data in one webhook
- ✅ No extra API calls
- ✅ Your code already handles it
- ✅ More reliable
- ✅ Easier to debug

**How?**
- Use V1 events (`customer.created`, not `v2.billing.customer.created`)
- Select "Latest API version" in Stripe
- All events automatically include full objects

### **Don't Use THIN** ❌ (for now)

**Why not?**
- ❌ Need to fetch data via API
- ❌ More code complexity
- ❌ No benefit at your scale
- ❌ Harder to debug

**When to reconsider:** If you scale to 100K+ subscribers

---

## 📋 **EVENTS TO ENABLE:**

### **Must Have:**
```
✅ customer.created
✅ customer.updated
✅ customer.deleted
✅ customer.subscription.created
✅ customer.subscription.updated
✅ customer.subscription.deleted
✅ invoice.payment_succeeded
✅ invoice.payment_failed
✅ checkout.session.completed
```

### **Nice to Have:**
```
⭕ customer.subscription.paused
⭕ customer.subscription.resumed
⭕ checkout.session.expired
⭕ payment_intent.succeeded
⭕ payment_intent.payment_failed
```

---

## 🛠️ **WHAT HAPPENS WHEN WEBHOOKS FIRE:**

```
Stripe Event: customer.subscription.created
    ↓
Your Backend: /api/v1/webhooks/stripe
    ↓
1. Validate signature ✅
2. Parse event
3. Dual-write to v1 + v2 tables
4. Auto-link customer to user by email
5. Grant video access
6. Log to webhook_events table
    ↓
Response: 200 OK
    ↓
Admin Dashboard: Shows "Healthy" ✅
```

---

## 🚨 **TROUBLESHOOTING:**

### **Problem: "Invalid signature"**
```
✅ Fix: Check webhook secret in admin UI
✅ Fix: Make sure you copied full whsec_ string
✅ Fix: Try deleting and recreating in Stripe
```

### **Problem: "Endpoint not found"**
```
✅ Fix: Check URL is exactly:
   https://watch.bookofmormonevidence.org/bome-backend/api/v1/webhooks/stripe
✅ Fix: Make sure backend is running
✅ Fix: Check firewall allows Stripe IPs
```

### **Problem: "Events not showing in dashboard"**
```
✅ Fix: Check backend logs for errors
✅ Fix: Send test webhook from Stripe
✅ Fix: Verify database connection
```

---

## 📊 **CURRENT STATUS:**

### **Your Webhook System:**

✅ **Code:** 100% ready  
✅ **Database:** Tables created  
✅ **Services:** V1 + V2 dual-write  
✅ **Security:** Signature validation  
✅ **Logging:** Full audit trail  
✅ **Monitoring:** Admin dashboard  

**Missing:** Just the Stripe configuration! (5 minutes)

---

## 🎉 **BENEFITS ONCE LIVE:**

1. **Real-time sync** - No manual "Simple Sync" needed
2. **Auto-linking** - New customers instantly linked to users
3. **Video access** - Automatic based on subscription status
4. **Audit trail** - Every event logged
5. **Error recovery** - Failed events can be retried
6. **Status monitoring** - Live health checks

---

## 📞 **READY TO GO?**

**Just need 3 things from you:**

1. Access to Stripe Dashboard
2. 5 minutes of time
3. Confirmation you want SNAPSHOT mode (recommended ✅)

**Then we:**
1. Create the webhook endpoint
2. Copy the secret
3. Add to your admin UI
4. Test it
5. Done! 🎉

---

**Want me to walk you through it step-by-step?** 🚀

