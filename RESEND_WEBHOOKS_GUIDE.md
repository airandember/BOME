# Resend Stripe Webhooks to Populate Enhanced Data

## 🎯 Problem

The migration was applied successfully, but existing webhook logs in the database still show "System Event" because they were created before the new columns existed. We need fresh webhook events to see the enhanced data.

---

## ✅ Solution: Resend Webhooks from Stripe Dashboard

### **Option 1: Resend Individual Webhooks (Recommended)**

1. **Go to Stripe Dashboard**
   - Navigate to: https://dashboard.stripe.com/webhooks

2. **Click on your webhook endpoint**
   - Should be something like: `https://your-domain.com/api/v1/stripe/webhooks`

3. **Find the "Events" section**
   - You'll see a list of recent events that were sent to this endpoint

4. **For each event you want to resend:**
   - Click on the event (e.g., `customer.subscription.updated`)
   - Click the **"Resend"** button in the top right
   - Click **"Resend event"** to confirm

5. **Recommended events to resend:**
   - ✅ `customer.created` - To see customer details
   - ✅ `customer.subscription.created` - To see subscription creation
   - ✅ `customer.subscription.updated` - To see subscription changes
   - ✅ `invoice.payment_succeeded` - To see payment details

---

### **Option 2: Use the "Send test event" Feature**

1. **In Stripe Dashboard → Webhooks**
2. **Click your webhook endpoint**
3. **Click "Send test event"**
4. **Select event type from dropdown:**
   - `customer.created`
   - `customer.subscription.created`
   - `customer.subscription.updated`
   - `invoice.payment_succeeded`
5. **Click "Send test event"**

⚠️ **Note:** Test events use fake IDs and won't link to real users, but they'll show the data structure.

---

### **Option 3: Use the Ping Feature in Your Dashboard**

1. **Go to your admin dashboard:**
   - Navigate to: Webhooks tab
2. **Click "Send Ping"** button
3. This will send a `v2.core.event_destination.ping` event
4. It won't have subscription data, but verifies the system works

---

## 🔍 What to Resend

### **High Priority (Resend These):**

| Event Type | Why | What You'll See |
|------------|-----|-----------------|
| `customer.subscription.updated` | Shows subscription changes | User email, subscription ID, status, amount |
| `invoice.payment_succeeded` | Shows payment details | User email, invoice ID, amount paid |
| `customer.updated` | Shows customer info | User email, customer ID, address |

### **Medium Priority:**

| Event Type | Why | What You'll See |
|------------|-----|-----------------|
| `customer.subscription.created` | Initial subscription | User email, new subscription details |
| `customer.created` | New customer | User email, customer ID, metadata |

---

## 📊 What You'll See After Resending

**Before (old data):**
```
System Event | customer.subscription.updated | success
  Technical Details
    Response Time: 181ms
```

**After (resent webhook):**
```
srnavasettles@gmail.com (ID: 9042) | customer.subscription.updated | success
  
📊 Technical Details
  Response Time: 181ms
  Payload Size: 7.5KB
  Status Code: 200

🔗 Stripe Information
  Event ID: evt_1SVx2aFpxJJNWdU8n67rCYuN
  Object ID: sub_1Quyv4FpxJJNWdU8HeoWqZIw
  Object Type: subscription
  API Version: 2025-10-29.clover

💳 Subscription & Payment
  Subscription: sub_1Quyv4FpxJJNWdU8HeoWqZIw
  Status: unpaid
  Customer: cus_Roc9Ab6KeG3x4v

📄 Full Event Data
{
  "id": "evt_1SVx2aFpxJJNWdU8n67rCYuN",
  "type": "customer.subscription.updated",
  "data": {
    "object": {
      "id": "sub_1Quyv4FpxJJNWdU8HeoWqZIw",
      "customer": "cus_Roc9Ab6KeG3x4v",
      "status": "unpaid",
      "plan": {
        "amount": 997,
        "currency": "usd",
        "interval": "month"
      }
    }
  }
}
```

---

## 🎬 Step-by-Step Example

### **Resending a Subscription Update Event:**

1. **Open Stripe Dashboard:**
   ```
   https://dashboard.stripe.com/webhooks
   ```

2. **Click your webhook endpoint**

3. **Scroll down to "Events" section**
   - Look for recent events like:
     - `evt_1SVx2aFpxJJNWdU8n67rCYuN` (Nov 21, 9:18 AM)
     - `customer.subscription.updated` for `srnavasettles@gmail.com`

4. **Click on the event**

5. **Click "Resend" button** (top right, near the Event ID)

6. **Confirm: "Resend event"**

7. **Wait 1-2 seconds**

8. **Go to your webhook dashboard:**
   - Refresh the page
   - You should see a NEW entry with full details!
   - Click "▶ Details" to see all the rich data

---

## 🔄 Optional: Clear Old Data

If you want to start fresh and only see new enhanced webhooks:

```sql
-- Delete all old webhook events (OPTIONAL - only if you want a clean slate)
DELETE FROM webhook_events 
WHERE user_email IS NULL 
AND created_at < NOW() - INTERVAL '1 hour';
```

⚠️ **Warning:** This will delete old logs. Only do this if you don't need the historical data.

---

## ✅ Verify It's Working

After resending a webhook, check your dashboard:

1. **Webhook logs auto-refresh** (or click Refresh)
2. **Look for the newest entry** (should be at the top)
3. **Check the "User" column** - should show email instead of "System Event"
4. **Click "▶ Details"**
5. **You should see:**
   - ✅ Stripe Information section
   - ✅ Subscription & Payment section (if applicable)
   - ✅ Full Event Data (JSON)

---

## 🎯 Quick Test

**Fastest way to test right now:**

1. Go to: https://dashboard.stripe.com/webhooks
2. Click your webhook endpoint
3. Click "Send test event"
4. Select "customer.subscription.updated"
5. Click "Send test event"
6. Go to your admin dashboard → Webhooks tab
7. Click Details on the newest entry
8. **BAM!** Full details! 🎉

---

## 📝 Summary

- ✅ Migration is applied
- ✅ Backend code is ready
- ✅ Frontend UI is ready
- ⏳ Need to resend webhooks to see the new data
- 🎉 Future webhooks will automatically have all details!

---

**Next Step:** Resend 3-5 webhooks from Stripe to populate your dashboard with rich data!

**Time Required:** 2 minutes  
**Difficulty:** Easy (just clicking "Resend" in Stripe)

