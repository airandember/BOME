# 🔧 Fix Thin Webhook Invalid Signature - NOW

**Issue:** Thin webhook failing with "Invalid signature"  
**Cause:** Secret saved to database but not loaded into memory  
**Solution:** Hot-reload the secret (no restart needed!)  
**Time:** 30 seconds

---

## ⚡ **Quick Fix (30 seconds):**

### **Step 1: Go to Admin UI**
```
https://watch.bookofmormonevidence.org/admin/streaming/stripe
```
Click **"Setup"** tab

### **Step 2: Re-Save Thin Secret**

1. Scroll to **"Webhook Secret - Thin (V2)"**
2. **Paste the same thin secret** you got from Stripe (starts with `whsec_`)
3. Click **"Save Thin Secret"**
4. ✅ See: **"Thin webhook secret saved and hot-reloaded successfully"**

### **Step 3: Test Immediately**

1. Go to Stripe Dashboard → Webhooks → Your Thin webhook
2. Click **"Send test webhook"**
3. Select: `v2.core.event_destination.ping`
4. Click **"Send test webhook"**
5. ✅ Should now see: **HTTP 200** ✨

---

## 🎯 **What Happens:**

```
You paste secret → Click Save
    ↓
Admin UI: POST /admin/streaming/stripe/webhook-secret
    ↓
Backend: Encrypts secret
    ↓
Backend: Saves to secure_settings table
    ↓
Backend: 🔥 HOT-RELOADS secret into memory (NEW!)
    ↓
Backend: Returns success
    ↓
✅ Next webhook validates with new secret
    ↓
✅ Returns HTTP 200 instead of 400!
```

---

## 📊 **Backend Logs (Check These):**

After saving, you'll see:
```
🔗 Received thin webhook secret request: whsec_...
✅ Stripe thin webhook secret updated successfully in memory
```

When test webhook arrives:
```
📨 Webhook received: v2.core.event_destination.ping
✅ V2 thin webhook signature validated with thin secret
✅ Webhook: Successfully processed v2 thin event
```

---

## ❓ **Why Did This Happen?**

1. You saved the thin secret via admin UI ✅
2. It was encrypted and stored in database ✅
3. But backend didn't reload it into memory ❌
4. Webhooks still used old (empty) secret ❌
5. **Now we have hot-reload!** ✅

---

## 🎉 **After This Fix:**

- ✅ Thin webhooks will validate correctly
- ✅ Snapshot webhooks continue working
- ✅ All future secret updates hot-reload
- ✅ No more restarts needed for secret changes

---

## 🔄 **If It Still Fails:**

1. **Check the secret is correct:**
   - Should start with `whsec_`
   - Copy it directly from Stripe Dashboard
   - No extra spaces or characters

2. **Check backend logs:**
   ```
   Look for: "✅ Stripe thin webhook secret updated successfully in memory"
   If missing: Secret didn't hot-reload
   ```

3. **Restart backend (last resort):**
   - Redeploy on DigitalOcean
   - Backend will load secret on startup

---

## 📞 **Verification:**

After the fix, your thin webhook should show:

**Stripe Dashboard:**
- ✅ Status: Active
- ✅ Last delivery: Success (HTTP 200)
- ✅ Recent events: All successful

**Your Admin Dashboard:**
- ✅ Webhook Health: Healthy
- ✅ Events Today: Includes V2 events
- ✅ Success Rate: 100%

---

**Ready? Go re-save that thin secret now! 🚀**

**No restart needed - the hot-reload will fix it immediately!** 🔥

