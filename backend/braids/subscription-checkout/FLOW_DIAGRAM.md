# Subscription Checkout - Complete Flow Diagram

## Full User Journey (Frontend → Backend → Database)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         FRONTEND (Svelte)                                │
└─────────────────────────────────────────────────────────────────────────┘

/subscription Page
│
├─ User clicks "Subscribe to Monthly" button
│  └─> handleSelectPlan(plan)
│       └─> startEmbeddedCheckout(plan)
│            │
│            ├─> Load Stripe.js
│            ├─> GET /api/v1/stripe/config (get publishable key)
│            └─> POST /api/v1/stripe/checkout-session
│                 Body: { plan_id: "13", return_url: "/checkout/success?session_id={CHECKOUT_SESSION_ID}" }
│                 │
                  ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    BACKEND (Go - Stripe Public Service)                  │
└─────────────────────────────────────────────────────────────────────────┘

POST /api/v1/stripe/checkout-session Handler
│
├─> Middleware: AuthRequired() ✅ User must be logged in
│
├─> CreateEmbeddedCheckoutSession(planID, userID, returnURL)
│    │
│    ├─> Query database for plan details
│    ├─> Create Stripe Session with metadata:
│    │    { user_id: "10467", plan_id: "13" }
│    └─> Return: { client_secret: "cs_live_xxx_secret_yyy" }
│
                  │
                  ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                     STRIPE HOSTED CHECKOUT                               │
└─────────────────────────────────────────────────────────────────────────┘

User enters payment details
│
├─> Card number, expiry, CVC
├─> Stripe validates payment method
├─> Stripe creates:
│    ├─ Customer (cus_xxx)
│    ├─ PaymentIntent (pi_xxx)
│    ├─ Subscription (sub_xxx)
│    └─ Invoice (in_xxx)
│
└─> Payment succeeds ✅
     │
     ├─────────────────────────────────────────────────────────────────────┐
     │                                                                      │
     ▼                                                                      ▼
REDIRECT USER                                              SEND WEBHOOKS (async)
     │                                                                      │
     └─> https://watch.bome.org/checkout/success?session_id=cs_live_xxx   │
                                                                            │
┌─────────────────────────────────────────────────────────┐               │
│           FRONTEND - Success Page                        │               │
└─────────────────────────────────────────────────────────┘               │
                                                                            │
/checkout/success Page Loads                                               │
│                                                                           │
├─> Extract session_id from URL params                                     │
├─> checkSessionStatus()                                                   │
│    │                                                                      │
│    └─> GET /api/v1/stripe/session/cs_live_xxx  ←── PRIMARY CONFIRMATION │
│         │                                                                 │
          ▼                                                                 │
┌─────────────────────────────────────────────────────────┐               │
│     BACKEND - Session Verification (IMMEDIATE)           │               │
└─────────────────────────────────────────────────────────┘               │
                                                                            │
GET /api/v1/stripe/session/:session_id Handler                             │
│                                                                           │
├─> Middleware: AuthRequired() ✅                                          │
│                                                                           │
├─> VerifyAndGrantAccess(sessionID, userID)                                │
│    │                                                                      │
│    ├─> 1. Call Stripe API: session.Get(session_id)                      │
│    │     └─> Returns: { payment_status: "paid",                         │
│    │                     customer_id: "cus_xxx",                         │
│    │                     subscription_id: "sub_xxx" }                    │
│    │                                                                      │
│    ├─> 2. Check payment status: "paid" ✅                                │
│    │                                                                      │
│    ├─> 3. Link customer to user                                          │
│    │     └─> CustomerLinkingService.LinkUserToCustomers(userID)         │
│    │          │                                                           │
│    │          ├─> Find stripe_customers_v2 WHERE email = user.email     │
│    │          ├─> INSERT INTO user_stripe_customers_v2                  │
│    │          │    (user_id, stripe_customer_id, is_primary)            │
│    │          │                                                           │
│    │          └─> checkAndGrantVideoAccessAfterLinking(userID)          │
│    │               │                                                      │
│    │               ├─> Query for active subscriptions                    │
│    │               │    WHERE user_stripe_customers.user_id = X          │
│    │               │    AND subscription.status IN ('active', 'trialing')│
│    │               │                                                      │
│    │               └─> IF has_subscription AND !has_access:              │
│    │                    └─> UPDATE users                                 │
│    │                         SET has_video_access = true,                │
│    │                             video_access_granted_at = NOW(),        │
│    │                             video_access_source = 'retroactive_linking'│
│    │                                                                      │
│    └─> 4. Grant video access via SubscriptionManager                     │
│         └─> GrantVideoAccess(userID, "session_verification:cs_xxx")     │
│              │                                                            │
│              ├─> Check if user already has access                        │
│              │    IF yes: Update source to include "session_verification"│
│              │    IF no:  Grant new access                               │
│              │                                                            │
│              └─> UPDATE users                                            │
│                   SET has_video_access = true,                           │
│                       video_access_granted_at = NOW(),                   │
│                       video_access_source = 'session_verification',      │
│                       manual_video_access = true                         │
│                   WHERE id = userID                                      │
│                                                                           │
├─> Return to frontend:                                                    │
│    {                                                                      │
│      "status": "success",                                                │
│      "data": {                                                           │
│        "session_id": "cs_live_xxx",                                      │
│        "payment_status": "paid",                                         │
│        "subscription_id": "sub_xxx",                                     │
│        "video_access_granted": true  ← NEW                               │
│      }                                                                    │
│    }                                                                      │
│                                                                           │
          │                                                                 │
          ▼                                                                 │
┌─────────────────────────────────────────────────────────┐               │
│        FRONTEND - Success Page (Continued)               │               │
└─────────────────────────────────────────────────────────┘               │
                                                                            │
Response received (< 500ms elapsed)                                        │
│                                                                           │
├─> Show toast: "🎉 Payment successful! You now have instant access!"      │
├─> Auto-redirect to /videos after 2 seconds                               │
│                                                                           │
User navigates to /videos                                                  │
│                                                                           │
└─> ✅ USER HAS IMMEDIATE ACCESS!                                          │
                                                                            │
                                                                            │
═══════════════════════════════════════════════════════════════════════════│
                                                                            │
                    MEANWHILE (1-30 seconds later)...                      │
                                                                            ▼
┌─────────────────────────────────────────────────────────────────────────┐
│              WEBHOOK CONFIRMATION (BACKUP)                               │
└─────────────────────────────────────────────────────────────────────────┘

Stripe sends webhooks to: /api/v1/admin/streaming/stripe/webhooks/

1. customer.subscription.created
   │
   ├─> Webhook Handler validates signature ✅
   ├─> Dual-write: Sync to v1 + v2 tables
   │    └─> INSERT INTO stripe_subscriptions_v2
   │
   ├─> SubscriptionManager.UpdateVideoAccessForSubscription(sub_id)
   │    │
   │    ├─> Get customer_id from subscription
   │    ├─> CustomerLinkingService.GetUserByStripeCustomerID(cus_xxx)
   │    │     └─> Returns user 10467
   │    │
   │    └─> GrantVideoAccess(10467, "webhook:subscription:sub_xxx")
   │         │
   │         ├─> Check if user already has access
   │         │    └─> YES (from session verification)
   │         │
   │         ├─> Update source to append "webhook"
   │         │    └─> UPDATE users
   │         │         SET video_access_source = 'session_verification,webhook'
   │         │         WHERE id = 10467
   │         │
   │         └─> Log: "ℹ️ User 10467 already has video access, updated source"
   │
   └─> ✅ Webhook confirmed (idempotent, no duplicate grant)

2. invoice.payment_succeeded
   │
   ├─> Webhook Handler validates signature ✅
   ├─> Sync invoice to stripe_invoices table
   │
   └─> (Similar process - idempotent confirmation)

═══════════════════════════════════════════════════════════════════════════

FINAL STATE IN DATABASE:

users table:
┌────────┬────────────────┬──────────────────┬─────────────────────┬───────────────────────────────────┐
│ id     │ email          │ has_video_access │ video_access_source │ video_access_granted_at           │
├────────┼────────────────┼──────────────────┼─────────────────────┼───────────────────────────────────┤
│ 10467  │ user@email.com │ true             │ session,webhook     │ 2025-11-18 22:15:57               │
└────────┴────────────────┴──────────────────┴─────────────────────┴───────────────────────────────────┘

stripe_customers_v2:
┌────┬──────────────┬────────────────┬──────────────────────┐
│ id │ stripe_id    │ email          │ stripe_created_at    │
├────┼──────────────┼────────────────┼──────────────────────┤
│ 42 │ cus_xxx      │ user@email.com │ 2025-11-18 22:05:05  │
└────┴──────────────┴────────────────┴──────────────────────┘

user_stripe_customers_v2 (linking table):
┌────┬─────────┬────────────────────┬────────────┬──────────────────────┐
│ id │ user_id │ stripe_customer_id │ is_primary │ first_linked_at      │
├────┼─────────┼────────────────────┼────────────┼──────────────────────┤
│ 1  │ 10467   │ 42                 │ true       │ 2025-11-18 22:15:57  │
└────┴─────────┴────────────────────┴────────────┴──────────────────────┘

stripe_subscriptions_v2:
┌────┬──────────────┬─────────────┬────────┬─────────────────────────┐
│ id │ stripe_id    │ customer_id │ status │ created_at              │
├────┼──────────────┼─────────────┼────────┼─────────────────────────┤
│ 99 │ sub_xxx      │ 42          │ active │ 2025-11-18 22:05:05     │
└────┴──────────────┴─────────────┴────────┴─────────────────────────┘

═══════════════════════════════════════════════════════════════════════════

TIMING BREAKDOWN:

0.0s  - User clicks "Subscribe"
0.5s  - Checkout session created
5.0s  - User enters payment details
5.5s  - Payment processed by Stripe
5.6s  - User redirected to /checkout/success
6.0s  - Session verification completes
6.0s  - ✅ USER HAS ACCESS (total: 6 seconds from click)
8.0s  - Auto-redirect to /videos
15.0s - First webhook arrives (confirmation)
20.0s - Second webhook arrives (backup)

USER EXPERIENCE: Seamless, instant access in 6 seconds!
```

## Edge Case: Subscribe BEFORE Register

```
1. User not logged in → Goes to /subscription
2. Clicks "Subscribe" → System shows: "Please sign in to continue"
3. User clicks "Sign up" → Goes to /auth/register
4. MEANWHILE: User opens new tab, starts checkout directly from Stripe link
5. Completes payment → Stripe creates customer (cus_xxx) + subscription (sub_xxx)
6. User finishes registration in other tab → Registers email
7. System auto-links customer at:
   a) Registration ✅
   b) Email verification ✅
   c) Password setup ✅
8. checkAndGrantVideoAccessAfterLinking() triggers
   └─> Finds active subscription
   └─> Grants access with source: "retroactive_linking"
9. User logs in → ✅ Has access immediately!
```

## Email/Password vs OAuth2 Parity

### Email/Password Flow
```
Register → [Auto-link #1] → Verify Email → [Auto-link #2] → Setup Password → [Auto-link #3] → Login
           (links customer)    (safety net)                   (FIXED: was missing!)
```

### OAuth2 Flow
```
Click "Sign in with Google" → [Auto-link #1] → Login
                              (links customer)
```

Both now have equivalent reliability with multiple safety nets!

