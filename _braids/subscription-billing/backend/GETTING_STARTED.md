# Getting Started with Subscription Billing Braid

**Quick guide to Stripe integration and subscription management**

---

## What is This?

The Subscription Billing Braid handles subscription plans, Stripe integration, payment processing, webhooks, and billing operations. It is revenue-critical.

---

## Quick Start

### Key Production Files

- **Handlers**: `backend/subscription/handlers/subscription.go`
- **Services**: `backend/subscription/services/stripe.go`, `stripe_sync.go`
- **Models**: `backend/subscription/models/stripe_entities.go`, `subscription.go`
- **Webhooks**: `backend/internal/routes/stripe_webhook_routes.go`
- **Routes**: `backend/internal/routes/subscription.go`, `subscription_plans.go`

### Environment Variables

- `STRIPE_SECRET_KEY` - Stripe API secret
- `STRIPE_WEBHOOK_SECRET` - Webhook signature verification
- `STRIPE_PUBLISHABLE_KEY` - Frontend (optional)

### Common Scenarios

**"Webhook not processing"**
1. Check `backend/internal/routes/stripe_webhook_routes.go`
2. Verify STRIPE_WEBHOOK_SECRET matches Stripe dashboard
3. Ensure webhook URL is correct (e.g., /api/v1/stripe/webhook)

**"Subscription sync issues"**
1. See `backend/subscription/services/stripe_sync.go`
2. Check `backend/services/stripe/` for sync services
3. Run sync command: `backend/cmd/stripe-sync/`

**"Plan not appearing"**
1. Check subscription_plans table
2. Verify Stripe product/price IDs in sync
3. See `backend/internal/services/subscription_plans.go`

---

## Dependencies

- **Stripe API**: Payment processing, webhooks
- **Authentication Braid**: User identity for subscriptions
- **User Management**: Customer profile data

---

## Related Documentation

- **BRAID.md**: Full braid overview at `_braids/subscription-billing/BRAID.md`
- **Subscription Checkout**: `backend/braids/subscription-checkout/` for checkout flow

---

**Last Updated**: 2025-03-17
