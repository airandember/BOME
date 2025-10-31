# Phase 7: User-Controlled Subscription Management Dashboard

**Goal**: Give users full visibility and control over their subscriptions

---

## 🎯 **User Story**

**As a user**, I want to:
1. See **all my active subscriptions** in one place
2. See **my subscription history** (active, canceled, expired)
3. **Choose which subscription to keep** if I have multiple active ones
4. **Cancel unwanted subscriptions** with clear confirmation
5. See when my subscription renews and how much I'm paying

---

## 📊 **Dashboard Design**

### **View 1: Active Subscriptions** (if multiple)

```
┌──────────────────────────────────────────────────────────┐
│ ⚠️  Multiple Active Subscriptions Detected               │
│                                                           │
│ You have 3 active subscriptions. To avoid being charged  │
│ multiple times, please choose ONE subscription to keep   │
│ and cancel the others.                                   │
└──────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│ 📋 Subscription 1 - Yearly Plan                         │
│                                                          │
│ Status: ✅ Active                                        │
│ Price: $95.64/year                                       │
│ Started: Sep 26, 2023                                    │
│ Renews: Sep 26, 2025                                     │
│ Days Left: 336 days                                      │
│                                                          │
│ [ ⭐ Keep This One ]  [ ❌ Cancel ]                      │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│ 📋 Subscription 2 - Monthly Plan                        │
│                                                          │
│ Status: ✅ Active                                        │
│ Price: $7.97/month                                       │
│ Started: Jan 15, 2024                                    │
│ Renews: Nov 15, 2025                                     │
│ Days Left: 16 days                                       │
│                                                          │
│ [ ⭐ Keep This One ]  [ ❌ Cancel ]                      │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│ 📋 Subscription 3 - Lifetime Access                     │
│                                                          │
│ Status: ✅ Active                                        │
│ Price: $297.00 (one-time)                                │
│ Started: Mar 10, 2024                                    │
│ Expires: Never                                           │
│                                                          │
│ [ ⭐ Keep This One ]  [ Cannot Cancel - Lifetime ]       │
└─────────────────────────────────────────────────────────┘

[ Review & Confirm Cancellations ]
```

### **View 2: Single Active Subscription** (ideal state)

```
┌─────────────────────────────────────────────────────────┐
│ ✅ Your Active Subscription                              │
│                                                          │
│ Plan: Yearly Membership                                  │
│ Status: Active                                           │
│ Price: $95.64/year                                       │
│ Started: Sep 26, 2023                                    │
│ Next Billing: Sep 26, 2025                               │
│ Days Until Renewal: 336 days                             │
│                                                          │
│ [ 📝 Update Payment Method ]  [ ❌ Cancel Subscription ] │
└─────────────────────────────────────────────────────────┘
```

### **View 3: Subscription History**

```
┌─────────────────────────────────────────────────────────┐
│ 📜 Subscription History                                  │
│                                                          │
│ ✅ Yearly Plan - Active (current)                       │
│    Started: Sep 26, 2023 | Renews: Sep 26, 2025        │
│                                                          │
│ ❌ Monthly Plan - Canceled                              │
│    Started: Jan 15, 2024 | Ended: Oct 30, 2025         │
│    Reason: User chose different plan                     │
│                                                          │
│ ⏱️  Monthly Plan - Expired                              │
│    Started: Jun 10, 2023 | Ended: Sep 10, 2023         │
│    Reason: Payment failed                                │
│                                                          │
│ ✅ 30-Day Trial - Completed                             │
│    Started: May 11, 2023 | Ended: Jun 10, 2023         │
│    Converted to: Monthly Plan                            │
└─────────────────────────────────────────────────────────┘
```

---

## 🛠️ **Technical Implementation**

### **Backend API Endpoints** (to create)

```go
// Get user's subscription details (all subscriptions, not just active)
GET /api/v1/user/subscriptions

Response:
{
  "success": true,
  "active_subscriptions": [
    {
      "id": "sub_123",
      "stripe_customer_id": "cus_ABC",
      "plan_name": "Yearly",
      "status": "active",
      "price": 9564,  // cents
      "currency": "usd",
      "interval": "year",
      "current_period_start": "2023-09-26T23:20:45Z",
      "current_period_end": "2025-09-26T23:20:45Z",
      "days_until_renewal": 336,
      "cancel_at_period_end": false,
      "created_at": "2023-09-26T23:20:45Z"
    },
    {
      "id": "sub_456",
      "stripe_customer_id": "cus_XYZ",
      "plan_name": "Monthly",
      "status": "active",
      "price": 797,
      "currency": "usd",
      "interval": "month",
      "current_period_start": "2025-10-15T00:00:00Z",
      "current_period_end": "2025-11-15T00:00:00Z",
      "days_until_renewal": 16,
      "cancel_at_period_end": false,
      "created_at": "2024-01-15T00:00:00Z"
    }
  ],
  "canceled_subscriptions": [
    {
      "id": "sub_789",
      "plan_name": "Monthly",
      "status": "canceled",
      "canceled_at": "2025-10-30T00:00:00Z",
      "ended_at": "2025-10-30T00:00:00Z"
    }
  ],
  "subscription_count": {
    "active": 2,
    "trialing": 0,
    "canceled": 1,
    "past_due": 0,
    "total": 3
  }
}

// Cancel specific subscriptions (bulk)
POST /api/v1/user/subscriptions/cancel-multiple

Request:
{
  "subscription_ids": ["sub_456", "sub_789"],
  "keep_subscription_id": "sub_123"  // Optional: the one they want to keep
}

Response:
{
  "success": true,
  "message": "2 subscriptions will be canceled at the end of their billing periods",
  "canceled_subscriptions": [
    {
      "id": "sub_456",
      "status": "active",
      "cancel_at_period_end": true,
      "canceled_at": "2025-10-30T00:00:00Z",
      "ends_on": "2025-11-15T00:00:00Z"
    },
    {
      "id": "sub_789",
      "status": "canceled",
      "canceled_at": "2025-10-30T00:00:00Z"
    }
  ],
  "kept_subscription": {
    "id": "sub_123",
    "status": "active",
    "plan_name": "Yearly"
  }
}

// Get subscription history (all time)
GET /api/v1/user/subscriptions/history

Response:
{
  "success": true,
  "history": [
    {
      "id": "sub_123",
      "plan_name": "Yearly",
      "status": "active",
      "created_at": "2023-09-26T23:20:45Z",
      "current_period_end": "2025-09-26T23:20:45Z",
      "is_current": true
    },
    {
      "id": "sub_456",
      "plan_name": "Monthly",
      "status": "canceled",
      "created_at": "2024-01-15T00:00:00Z",
      "canceled_at": "2025-10-30T00:00:00Z",
      "ended_at": "2025-10-30T00:00:00Z",
      "cancellation_reason": "User chose different plan"
    }
  ]
}

// Cancel a single subscription
POST /api/v1/user/subscriptions/:subscription_id/cancel

Request:
{
  "reason": "Switching to yearly plan"  // Optional
}

Response:
{
  "success": true,
  "message": "Subscription will be canceled at the end of the billing period",
  "subscription": {
    "id": "sub_456",
    "cancel_at_period_end": true,
    "ends_on": "2025-11-15T00:00:00Z"
  }
}
```

---

## 🎨 **Frontend Components** (Svelte)

### **1. Subscription Dashboard Page**
**File**: `frontend/src/routes/user/subscriptions/+page.svelte`

**Features**:
- Fetch user's subscriptions on load
- Display warning banner if multiple active subscriptions
- Show subscription cards with details
- "Keep This One" + "Cancel" buttons
- Confirmation modal before canceling
- Loading states and error handling

### **2. Subscription Card Component**
**File**: `frontend/src/lib/components/SubscriptionCard.svelte`

**Props**:
```typescript
export let subscription: Subscription;
export let canSelect: boolean = false;  // true if multiple active
export let onKeep: (id: string) => void;
export let onCancel: (id: string) => void;
```

### **3. Cancellation Confirmation Modal**
**File**: `frontend/src/lib/components/SubscriptionCancelModal.svelte`

**Features**:
- Shows selected subscriptions to cancel
- Shows subscription to keep
- Warns about end dates ("You'll keep access until Nov 15, 2025")
- Requires confirmation checkbox: "I understand I'm canceling X subscriptions"

### **4. Subscription History Component**
**File**: `frontend/src/lib/components/SubscriptionHistory.svelte`

**Features**:
- Timeline view of all subscriptions
- Status badges (Active, Canceled, Expired, Trialing)
- Dates and prices
- Expandable details

---

## 🔐 **Security & Business Logic**

### **Backend Validation**
```go
// UserSubscriptionService.CancelMultipleSubscriptions()

1. Verify user owns ALL subscription IDs
   ├─ Query: SELECT user_id FROM user_stripe_customers_v2 WHERE stripe_customer_id IN (...)
   └─ If any don't belong to user → Return 403 Forbidden

2. Verify "keep" subscription is not in cancel list
   └─ If keep_subscription_id is in subscription_ids → Return 400 Bad Request

3. For each subscription to cancel:
   ├─ Call Stripe API: subscription.Update(id, {cancel_at_period_end: true})
   ├─ Update stripe_subscriptions_v2: canceled_at = NOW()
   └─ Log cancellation reason

4. If only 1 subscription will remain after cancellation:
   └─ Set as primary in user_stripe_customers_v2

5. Return success with details
```

### **User Permissions**
- ✅ Users can only see/cancel their own subscriptions
- ✅ Users cannot cancel lifetime subscriptions (if implemented)
- ✅ Users cannot cancel subscriptions that are already canceled
- ✅ Admins can see/cancel any subscription (with reason logging)

---

## 📋 **Implementation Phases**

### **Phase 7.1: Backend API** ✅
1. Create `UserSubscriptionService` (user-facing, not admin)
2. Implement `GetUserSubscriptions(userID)`
3. Implement `CancelMultipleSubscriptions(userID, subscriptionIDs, keepID)`
4. Implement `GetSubscriptionHistory(userID)`
5. Create routes in `backend/internal/routes/user_subscription_routes.go`
6. Add middleware: `AuthRequired()` (users can only access their own)

### **Phase 7.2: Frontend Components** ✅
1. Create `SubscriptionCard.svelte`
2. Create `SubscriptionCancelModal.svelte`
3. Create `SubscriptionHistory.svelte`
4. Create main page: `routes/user/subscriptions/+page.svelte`
5. Add navigation link to user dashboard

### **Phase 7.3: Testing** ✅
1. Test with user with 1 subscription (happy path)
2. Test with user with 3 subscriptions (multiple active)
3. Test cancellation flow (confirm modal, API call, success message)
4. Test history view (shows all statuses correctly)

---

## 🎯 **User Experience Flow**

### **Happy Path: User with Multiple Subscriptions**

```
1. User logs in
   ↓
2. Dashboard shows: "You have 3 active subscriptions"
   ↓
3. User sees all 3 subscriptions with details
   ↓
4. User clicks "Keep This One" on Yearly subscription
   ↓
5. Other subscriptions show "Will be canceled" state
   ↓
6. User clicks "Review & Confirm Cancellations"
   ↓
7. Modal shows:
   "You are canceling:
    - Monthly Plan (ends Nov 15, 2025)
    - Weekly Trial (ends Nov 5, 2025)
    
    You are keeping:
    - Yearly Plan (renews Sep 26, 2026)
    
    [x] I understand these subscriptions will be canceled
    
    [ Cancel Anyway ]  [ Go Back ]"
   ↓
8. User confirms
   ↓
9. API call to /user/subscriptions/cancel-multiple
   ↓
10. Success message: "2 subscriptions will be canceled. You'll keep access until their end dates."
   ↓
11. Dashboard refreshes → Now shows 1 active subscription ✅
```

---

## 💡 **Key Benefits of This Approach**

### **1. User Transparency** 🔍
- Users see exactly what they're paying for
- No surprises or confusion
- Clear dates and amounts

### **2. User Control** 🎮
- Users choose which subscription to keep
- Users see exactly when access ends
- Users can keep multiple if they want (their choice!)

### **3. Fair Billing** 💰
- Cancel at period end (users keep what they paid for)
- No immediate loss of access
- Clear communication about end dates

### **4. Trust Building** 🤝
- Proactive problem detection (multiple subs warning)
- Clear action steps
- Honest communication

### **5. Reduced Support Tickets** 📧
- Self-service subscription management
- Clear history and status
- Obvious next steps

---

## 🚨 **Edge Cases to Handle**

### **1. User Has No Subscriptions**
```
┌─────────────────────────────────────────┐
│ 📭 No Active Subscription              │
│                                         │
│ You don't have an active subscription. │
│                                         │
│ [ 🎬 Subscribe Now ]                   │
└─────────────────────────────────────────┘
```

### **2. User Has 1 Canceled Subscription (ending soon)**
```
┌─────────────────────────────────────────┐
│ ⚠️  Subscription Ending Soon            │
│                                         │
│ Your subscription ends on Nov 15, 2025 │
│ (16 days left)                          │
│                                         │
│ [ 🔄 Reactivate Subscription ]         │
└─────────────────────────────────────────┘
```

### **3. Subscription in Trialing Status**
```
Status: 🆓 Free Trial (14 days left)
Price: $0.00 (then $7.97/month)
Trial Ends: Nov 14, 2025
```

### **4. Past Due Subscription**
```
┌─────────────────────────────────────────┐
│ ⚠️  Payment Failed                      │
│                                         │
│ Your last payment failed. Please       │
│ update your payment method to keep     │
│ access to your subscription.           │
│                                         │
│ [ 💳 Update Payment Method ]           │
└─────────────────────────────────────────┘
```

---

## 📊 **Success Metrics**

After Phase 7:
- ✅ Users can see all their subscriptions
- ✅ Users can choose which subscription to keep
- ✅ Users can cancel multiple subscriptions at once
- ✅ Users can see subscription history
- ✅ Clear warning if multiple active subscriptions
- ✅ Cancel at period end (fair billing)
- ✅ Self-service (reduces support tickets)

---

## 🚀 **Ready to Build?**

This is a **much better approach** than auto-canceling! Users will appreciate:
1. Transparency (seeing all subscriptions)
2. Control (choosing which to keep)
3. Fairness (cancel at period end)
4. Trust (no surprises)

**Shall we start with Phase 7.1 (Backend API)?** 🎯

