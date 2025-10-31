# 🚀 Phase 6 Quick Start Guide

**Use This Guide To**: Test Phase 6 subscription management features

---

## 📋 **Available Admin Endpoints**

### **1. Get User Subscription Summary**
```bash
GET /api/v1/admin/subscription-manager/user/{user_id}/summary
```

**Example**:
```bash
curl http://localhost:8080/api/v1/admin/subscription-manager/user/7374/summary \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response**:
```json
{
  "success": true,
  "summary": {
    "user_id": 7374,
    "linked_customers": ["cus_TC503P4Vlw8XrB", "cus_S7VixQutVow4BB"],
    "linked_customer_count": 2,
    "active_subscriptions": 1,
    "canceled_subscriptions": 2,
    "total_subscriptions": 3,
    "has_video_access": true,
    "recommendation": "All good!",
    "action_needed": false
  }
}
```

---

### **2. Fix ALL Users with Multiple Subscriptions** 🔧
```bash
POST /api/v1/admin/subscription-manager/fix-all-multiple
```

**Example**:
```bash
curl -X POST http://localhost:8080/api/v1/admin/subscription-manager/fix-all-multiple \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**What It Does**:
1. Finds all users with 2+ active subscriptions
2. Keeps the newest subscription
3. Cancels all older subscriptions (at period end)
4. Returns detailed results

**Response**:
```json
{
  "success": true,
  "total_users": 10,
  "success_count": 10,
  "failure_count": 0,
  "results": [
    {
      "user_id": 7374,
      "new_subscription_id": "sub_TC503P4Vlw8XrB",
      "canceled_subscription_ids": ["sub_TC4zTVEOZbzRXe", "sub_S7VixQutVow4BB"],
      "video_access_granted": true,
      "error": ""
    }
  ]
}
```

**⚠️ IMPORTANT**: This makes real changes in Stripe! Run once after Phase 6 deployment.

---

### **3. Manually Grant Video Access**
```bash
POST /api/v1/admin/subscription-manager/user/{user_id}/grant-video-access
```

**Example**:
```bash
curl -X POST http://localhost:8080/api/v1/admin/subscription-manager/user/7374/grant-video-access \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"reason": "Promotional access for YouTube subscriber"}'
```

**Response**:
```json
{
  "success": true,
  "message": "Video access granted",
  "user_id": 7374,
  "reason": "Promotional access for YouTube subscriber"
}
```

---

### **4. Manually Revoke Video Access**
```bash
POST /api/v1/admin/subscription-manager/user/{user_id}/revoke-video-access
```

**Example**:
```bash
curl -X POST http://localhost:8080/api/v1/admin/subscription-manager/user/7374/revoke-video-access \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"reason": "User requested refund"}'
```

**⚠️ NOTE**: Will only revoke if user has NO other active subscriptions!

---

### **5. Update Video Access for Subscription**
```bash
POST /api/v1/admin/subscription-manager/subscription/{subscription_id}/update-video-access
```

**Example**:
```bash
curl -X POST http://localhost:8080/api/v1/admin/subscription-manager/subscription/sub_TC503P4Vlw8XrB/update-video-access \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**What It Does**:
- If subscription is `active` or `trialing` → Grant video access
- If subscription is `canceled`, `past_due`, or `unpaid` → Revoke video access (if no other subs)

---

## 🧪 **Testing Scenarios**

### **Scenario 1: Create New Subscription in Stripe**
1. Go to Stripe Dashboard
2. Create a new subscription for an existing customer
3. Check backend logs:
   ```
   📨 Webhook received: customer.subscription.created
   ✅ [Webhook v2] Subscription sub_NEW linked to user 123
   🔒 [Subscription Manager] Enforcing single subscription for user 123
   ⚠️  [Subscription Manager] Found 2 other active subscriptions
   ❌ [Subscription Manager] Canceling old subscription: sub_OLD1
   ✅ [Subscription Manager] Subscription sub_OLD1 will cancel at period end
   🎥 [Subscription Manager] Granting video access to user 123
   ```
4. Verify in Stripe Dashboard: Old subscriptions marked "cancel at period end"
5. Verify in database: `manual_video_access = true` for user

---

### **Scenario 2: Fail a Payment**
1. In Stripe Dashboard, go to a subscription
2. Click "Actions" → "Simulate payment failure"
3. Check backend logs:
   ```
   📨 Webhook received: invoice.payment_failed
   🚫 [Webhook v2] Payment failed for customer cus_XXX
   🔍 [Subscription Manager] Checking if user has other active subscriptions
   ℹ️  [Subscription Manager] User still has active subscription - keeping video access
   ```
4. If user has NO other subscriptions, video access will be revoked

---

### **Scenario 3: Fix Users with Multiple Subscriptions**
1. Run the fix endpoint:
   ```bash
   curl -X POST http://localhost:8080/api/v1/admin/subscription-manager/fix-all-multiple \
     -H "Authorization: Bearer YOUR_TOKEN"
   ```
2. Check response for affected users
3. Verify in Stripe Dashboard: Old subscriptions are canceled
4. Check backend logs for enforcement details

---

## 📊 **Check Phase 6 Is Working**

### **1. Backend Logs**
Start backend and watch for:
```
✅ Subscription Manager routes setup complete
```

### **2. Test Summary Endpoint**
Pick a user ID from your database and check their summary:
```bash
curl http://localhost:8080/api/v1/admin/subscription-manager/user/YOUR_USER_ID/summary \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### **3. Create Test Subscription in Stripe**
1. Go to Stripe Dashboard → Customers
2. Pick a customer with an existing subscription
3. Create a NEW subscription for them
4. Watch backend logs → Should see single sub enforcement!

---

## 🎯 **Success Indicators**

✅ **Backend starts successfully** (no errors about subscription manager)  
✅ **Summary endpoint returns user data** (linked customers, subscription counts)  
✅ **Fix endpoint finds users** (if any have multiple subscriptions)  
✅ **Webhook creates subscription** → Old subs are canceled  
✅ **Video access is granted** → User can watch videos  
✅ **Payment failure** → Video access revoked (if no other subs)

---

## 🚨 **Troubleshooting**

### **Error: "User has no linked customers"**
**Cause**: User not linked to any Stripe customers  
**Fix**: Run customer linking first:
```bash
curl -X POST http://localhost:8080/api/v1/admin/customer-linking/user/USER_ID \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### **Error: "Failed to cancel subscription in Stripe"**
**Cause**: Stripe API key issue or subscription already canceled  
**Check**: 
1. Verify Stripe API key in `secure_settings` table
2. Check subscription status in Stripe Dashboard
3. Check backend logs for Stripe API errors

### **Webhook Not Received**
**Cause**: Webhook not configured in Stripe  
**Check**:
1. Stripe Dashboard → Developers → Webhooks
2. Verify endpoint URL: `https://your-domain.com/api/v1/webhooks/stripe`
3. Verify events are enabled: `customer.subscription.*`, `invoice.payment_*`

---

## 📈 **Next Steps**

After testing Phase 6:
- ✅ Verify single subscription enforcement works
- ✅ Verify video access is granted/revoked correctly
- ✅ Run `/fix-all-multiple` to clean up existing data
- 🚀 Move to **Phase 7**: Frontend dashboard updates
- 🚀 Move to **Phase 8**: Parallel testing (v1 vs v2)

---

**Questions?** Check the full documentation in `STRIPE_V2_PHASE6_COMPLETE.md`

