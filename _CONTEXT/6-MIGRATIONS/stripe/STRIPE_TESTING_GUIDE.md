# 🧪 Stripe Embedded Checkout Testing Guide

## Prerequisites
- [ ] Backend running on port 8080
- [ ] Frontend running on dev server
- [ ] Logged in as a test user
- [ ] Using Stripe TEST keys (sk_test_... and pk_test_...)

## Test Scenarios

### 1. 🎯 Successful Checkout Flow
**Steps:**
1. Go to `/subscription` page
2. Click "Subscribe" on any plan
3. Embedded checkout should slide up from bottom
4. Use test card: `4242 4242 4242 4242`
5. Expiry: `12/34`, CVC: `123`
6. Fill in any name/address
7. Click "Subscribe"

**Expected Results:**
- ✅ Payment succeeds
- ✅ Redirected to success page
- ✅ User gets subscription in database
- ✅ Stripe webhook fires (check backend logs)

### 2. ❌ Failed Payment Flow
**Steps:**
1. Same as above but use declined card: `4000 0000 0000 0002`

**Expected Results:**
- ❌ Payment fails with error message
- ✅ User stays on checkout form
- ✅ Can try again with different card

### 3. 🔄 3D Secure Flow
**Steps:**
1. Same as above but use 3DS card: `4000 0025 0000 3155`

**Expected Results:**
- 🔄 Additional authentication popup
- ✅ Can complete or fail authentication
- ✅ Proper handling of both outcomes

### 4. 🚪 Cancel/Close Flow
**Steps:**
1. Open embedded checkout
2. Click the X button (top right)

**Expected Results:**
- ✅ Checkout slides down smoothly
- ✅ Returns to subscription page
- ✅ No payment attempted

## Success Page Testing

### What to Check:
- [ ] Success page displays correctly
- [ ] User subscription status updated
- [ ] Email confirmation sent (if configured)
- [ ] Access to premium content works
- [ ] Stripe Customer created in dashboard

## Backend Logs to Watch:
```
🔍 [STRIPE-PUBLIC] Using customer email: your@email.com
✅ [STRIPE-PUBLIC] Checkout session created: cs_test_...
✅ [STRIPE-PUBLIC] Returning client secret: cs_test_...
```

## Stripe Dashboard (Test Mode):
1. Go to https://dashboard.stripe.com/test/payments
2. Should see test payments appear in real-time
3. Check customer creation
4. Verify subscription details

## Common Issues & Solutions:

### Issue: "Invalid API Key"
- **Solution**: Make sure using `sk_test_...` not `sk_live_...`

### Issue: "No such price"
- **Solution**: Create test products/prices in Stripe test dashboard

### Issue: Checkout doesn't load
- **Solution**: Check browser console for errors
- **Solution**: Verify publishable key is `pk_test_...`

### Issue: Success page not reached
- **Solution**: Check return URL configuration
- **Solution**: Verify webhook endpoint working

## 💰 Cost: $0.00
All test mode transactions are completely FREE! 🎉
