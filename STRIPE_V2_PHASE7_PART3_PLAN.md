# Phase 7.3: Subscription Update Flow with Auto-Cancel

**Goal**: Integrate auto-cancel into the subscription creation/update process

---

## 🎯 **User Flow**

### **Scenario: User Already Has Active Subscription**

```
1. User browses subscription plans
   ↓
2. User clicks "Subscribe to Yearly Plan"
   ↓
3. FRONTEND: Detect user has active subscription
   ├─→ Show modal: "You have an active Monthly plan. Switch to Yearly?"
   └─→ User clicks "Yes, Switch Plans"
   ↓
4. BACKEND: Receive subscription change request
   ├─→ Check for existing active subscriptions
   ├─→ Found: Monthly subscription (sub_OLD123)
   ├─→ Call Stripe API: Cancel sub_OLD123
   ├─→ Wait for 200 OK from Stripe ✅
   ├─→ Create new Yearly subscription (sub_NEW456)
   └─→ Return success to frontend
   ↓
5. FRONTEND: Show success message
   "✅ Switched to Yearly plan! Your Monthly plan will end on [date]"
```

---

## 🔧 **Technical Implementation**

### **Backend Flow**

```go
// NewUserSubscriptionService method
func (s *UserSubscriptionService) CreateOrUpdateSubscription(
    userID int, 
    newPlanID string,
    paymentMethodID string,
) (*UserSubscription, error) {
    
    // Step 1: Get user's active subscriptions
    activeSubs, err := s.getActiveSubscriptions(userID)
    if err != nil {
        return nil, err
    }
    
    // Step 2: If user has active subscriptions, cancel them FIRST
    if len(activeSubs) > 0 {
        log.Printf("🔄 [Subscribe] User %d has %d active subscriptions - canceling before creating new one", userID, len(activeSubs))
        
        for _, sub := range activeSubs {
            // Cancel in Stripe (immediate, not at period end)
            params := &stripe.SubscriptionParams{
                CancelAtPeriodEnd: stripe.Bool(false), // Cancel NOW for plan changes
            }
            
            _, err := subscription.Cancel(sub.ID, params)
            if err != nil {
                return nil, fmt.Errorf("failed to cancel old subscription %s: %w", sub.ID, err)
            }
            
            log.Printf("✅ [Subscribe] Canceled old subscription: %s", sub.ID)
        }
    }
    
    // Step 3: Create new subscription in Stripe
    newSub, err := s.createStripeSubscription(userID, newPlanID, paymentMethodID)
    if err != nil {
        return nil, fmt.Errorf("failed to create new subscription: %w", err)
    }
    
    log.Printf("✅ [Subscribe] Created new subscription: %s", newSub.ID)
    
    // Step 4: Sync to v2 tables
    syncService := NewStripeSyncV2Service(s.db)
    if err := syncService.SyncSingleSubscription(context.Background(), newSub.ID); err != nil {
        log.Printf("⚠️  [Subscribe] Failed to sync new subscription to v2: %v", err)
    }
    
    // Step 5: Grant video access
    subscriptionManager := NewSubscriptionManagerService(s.db, s.linkingService)
    if err := subscriptionManager.GrantVideoAccess(userID, "new subscription created"); err != nil {
        log.Printf("⚠️  [Subscribe] Failed to grant video access: %v", err)
    }
    
    return convertToUserSubscription(newSub), nil
}
```

---

## 🎨 **Frontend Flow**

### **Update Subscription Page**

```typescript
// Before creating Stripe checkout session
async function handleSubscribe(planId: string) {
    try {
        // Check if user has active subscriptions
        const subscriptions = await UserSubscriptionService.getSubscriptions();
        
        if (subscriptions.has_multiple_active || subscriptions.active_subscriptions.length > 0) {
            // Show plan change confirmation modal
            showPlanChangeModal = true;
            selectedPlanId = planId;
            return;
        }
        
        // No active subscriptions - proceed normally
        await createCheckoutSession(planId);
    } catch (err) {
        console.error('Failed to check subscriptions:', err);
        // Continue anyway - backend will handle it
        await createCheckoutSession(planId);
    }
}

async function confirmPlanChange() {
    // User confirmed they want to change plans
    // Backend will auto-cancel old subscription before creating new one
    await createCheckoutSession(selectedPlanId);
}
```

---

## 📋 **API Endpoint**

### **New Endpoint: Create/Update Subscription**

```
POST /api/v1/user/subscriptions/create-or-update

Request:
{
  "plan_id": "price_1234567890",
  "payment_method_id": "pm_1234567890"
}

Response (Success):
{
  "success": true,
  "message": "Subscription created successfully",
  "subscription": {
    "id": "sub_NEW456",
    "plan_name": "Yearly",
    "status": "active",
    ...
  },
  "canceled_subscriptions": [
    {
      "id": "sub_OLD123",
      "plan_name": "Monthly",
      "ended_at": "2025-10-30T12:00:00Z"
    }
  ]
}

Response (Error):
{
  "success": false,
  "error": "Failed to cancel old subscription"
}
```

---

## 🚨 **Important Decisions**

### **1. Cancel Immediately vs. At Period End**

**When CHANGING plans** (user choosing a different plan):
- ✅ **Cancel IMMEDIATELY** (don't wait for period end)
- Why? User is upgrading/downgrading and wants the new plan NOW
- Pro-rate if possible

**When CANCELING without replacement** (user's choice via UI):
- ✅ **Cancel AT PERIOD END** (user keeps what they paid for)
- Why? Fair billing - they paid for the full period
- This is what Phase 7.2 does

### **2. Refund Policy**

**Downgrading** (e.g., Yearly → Monthly):
- Calculate pro-rated refund
- Issue credit or refund to payment method

**Upgrading** (e.g., Monthly → Yearly):
- Calculate pro-rated charge
- Charge difference immediately

**Canceling**:
- No refund
- Access until period end

### **3. Error Handling**

**If old subscription cancel FAILS**:
- ❌ **DO NOT** create new subscription
- Return error to user
- Ask them to contact support

**If new subscription create FAILS** (after old was canceled):
- 🚨 **CRITICAL**: Old subscription is canceled, new one failed
- Immediately try to reactivate old subscription
- If that fails, flag for manual intervention
- Send alert to admin

---

## 🔧 **Implementation Steps**

1. ✅ Add `CreateOrUpdateSubscription` method to `UserSubscriptionService`
2. ✅ Add `getActiveSubscriptions` helper method
3. ✅ Create API endpoint `/user/subscriptions/create-or-update`
4. ✅ Add frontend "Plan Change" confirmation modal
5. ✅ Update subscription checkout flow to check for existing subscriptions
6. ✅ Add error recovery logic (reactivate old subscription if new one fails)
7. ✅ Test with real Stripe data

---

## 📊 **Success Criteria**

- ✅ User with active subscription can change plans
- ✅ Old subscription is canceled before new one is created
- ✅ User receives confirmation of plan change
- ✅ No duplicate active subscriptions
- ✅ Error handling prevents orphaned states
- ✅ Admin can see plan change history

---

## 🎯 **Testing Scenarios**

### **Test 1: User Changes from Monthly to Yearly**
```
1. User has active Monthly subscription ($7.97/month)
2. User clicks "Subscribe to Yearly" ($95.64/year)
3. Frontend shows: "Change from Monthly to Yearly?"
4. Backend cancels Monthly subscription
5. Backend creates Yearly subscription
6. User sees: "✅ Switched to Yearly! Your Monthly plan has been canceled."
```

### **Test 2: User Has Multiple Subscriptions**
```
1. User has 2 active subscriptions (Monthly + Yearly)
2. User clicks "Subscribe to Lifetime"
3. Backend cancels BOTH old subscriptions
4. Backend creates Lifetime subscription
5. User sees: "✅ Subscribed to Lifetime! Your 2 previous subscriptions have been canceled."
```

### **Test 3: Cancellation Fails**
```
1. User has active Monthly subscription
2. User clicks "Subscribe to Yearly"
3. Backend tries to cancel Monthly → Stripe returns error
4. Backend DOES NOT create Yearly subscription
5. User sees: "❌ Failed to change plans. Please contact support."
6. User still has their Monthly subscription (no changes made)
```

---

**Ready to implement Phase 7.3?** 🚀

