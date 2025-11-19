# Subscription Checkout BRAID

## Overview

The Subscription Checkout BRAID handles the complete lifecycle of user subscriptions, from initial checkout through payment confirmation and video access provisioning. This BRAID implements a **dual-confirmation pattern** for reliability and instant user feedback.

## Problem Statement

Users need to:
1. Subscribe to premium plans to access video content
2. Receive **immediate** confirmation and access after payment
3. Have reliable access even if webhooks are delayed or fail
4. Experience seamless flows whether subscribing before or after account creation

## Architecture Pattern: Dual-Confirmation

```
┌─────────────────────────────────────────────────────────────┐
│                    USER COMPLETES PAYMENT                    │
└────────────────────┬────────────────────────────────────────┘
                     │
        ┌────────────┴────────────┐
        │                         │
   PRIMARY (Immediate)       SECONDARY (Backup)
   Session Verification      Webhook Confirmation
        │                         │
        ├─ Verify with Stripe     ├─ customer.subscription.created
        ├─ Link customer          ├─ invoice.payment_succeeded
        ├─ Grant access           ├─ Verify & confirm access
        │  (0-2 seconds)          │  (1-30 seconds)
        │                         │
        └────────────┬────────────┘
                     │
            ✅ USER HAS ACCESS
```

### Why Dual-Confirmation?

- **Primary (Session Verification):** Instant UX, user gets immediate feedback
- **Secondary (Webhooks):** Ensures no missed grants, handles edge cases
- **Idempotent:** Both can run safely, no duplicate grants

---

## Components

### Entry Points (Frontend to Backend)

1. **Plan Selection** → `/subscription`
2. **Checkout Initiation** → `POST /api/v1/stripe/checkout-session`
3. **Payment Return** → `/checkout/success?session_id={CHECKOUT_SESSION_ID}`
4. **Session Verification** → `GET /api/v1/stripe/session/:session_id`
5. **Webhook Reception** → `POST /api/v1/admin/streaming/stripe/webhooks/`

---

## Key Files by Layer

### Frontend (Presentation)
- `frontend/src/routes/subscription/+page.svelte` - Plan selection & embedded checkout
- `frontend/src/routes/checkout/success/+page.svelte` - Success page & verification
- `frontend/src/lib/services/public-plans-service.ts` - API client

### Backend Routes (Application)
- `backend/internal/routes/stripe_public_routes.go` - Session & checkout endpoints
- `backend/internal/routes/stripe_webhook_routes.go` - Webhook handlers
- `backend/internal/routes/auth.go` - Registration (auto-linking integration)

### Backend Services (Business Logic)
- `backend/internal/services/stripe_public.go` - `VerifyAndGrantAccess()`
- `backend/internal/services/customer_linking_service.go` - Customer-user linking
- `backend/internal/services/subscription_manager_service.go` - Access grants
- `backend/internal/services/oauth2.go` - OAuth2 integration (auto-linking)

### Database
- `users` - `has_video_access`, `video_access_source`, `video_access_granted_at`
- `stripe_customers_v2` - Stripe customer records
- `user_stripe_customers_v2` - Customer-user relationship
- `stripe_subscriptions_v2` - Subscription records
- `stripe_invoices` - Invoice records

---

## Flow Documentation

See individual strand files for detailed flows:
- `strands/checkout-flow/` - Embedded checkout process
- `strands/session-verification/` - Immediate access grant
- `strands/customer-linking/` - Auto-linking mechanisms
- `strands/webhook-confirmation/` - Backup confirmation
- `strands/access-management/` - Grant/revoke logic

---

## Edge Cases Handled

1. **Subscribe Before Register** - Retroactive access grant via customer linking
2. **Incomplete Subscription** - Checks invoice payment status
3. **Webhook Delays** - Session verification provides immediate access
4. **OAuth2 vs Email/Password** - Parity achieved with multiple safety nets

---

## Related Documentation

- `EMAIL_PASSWORD_BUG_FIX.md` - Password setup missing auto-linking
- `RETROACTIVE_ACCESS_FIX.md` - Subscribe-before-register solution
- `DUAL_CONFIRMATION_SUBSCRIPTION_FLOW.md` - Architecture overview

Last Updated: 2025-11-18

