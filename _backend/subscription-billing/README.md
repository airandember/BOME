# 🧬 BRAID 04: Subscription & Billing
**Status**: ⚪ Not Started | **Priority**: 🔴 Critical | **Complexity**: High

---

## 📋 **Braid Overview**

**Purpose**: Stripe integration, subscription management, billing, and payment processing  
**Estimated Time**: 6-7 days  
**Dependencies**: Authentication (user identity)

---

## 🎯 **What This Braid Will Cover**

### **Stripe Integration**
- Payment processing
- Customer management
- Webhook handling
- Subscription synchronization

### **Subscription Management**
- Subscription plans
- Plan upgrades/downgrades
- Trial periods
- Cancellations

### **Billing**
- Invoice generation
- Payment history
- Failed payment handling
- Refunds

### **Coupons & Discounts**
- Coupon creation
- Discount application
- Promotional codes
- Revenue tracking

---

## 📁 **Key Files to Document**

### **Backend**:
- `backend/internal/services/stripe.go` (⚠️ 2,809 lines!)
- `backend/internal/routes/stripe_webhook_routes.go` (862 lines)
- `backend/internal/routes/subscription.go`
- `backend/internal/routes/subscription_plans.go`
- `backend/internal/services/stripe_sync.go`
- `backend/internal/services/stripe_customers.go`

### **Frontend**:
- `frontend/src/routes/checkout/+page.svelte`
- `frontend/src/routes/subscription/+page.svelte`
- `frontend/src/routes/account/billing/+page.svelte`
- `frontend/src/routes/admin/streaming/stripe/webhooks/+page.svelte`

---

## 🧬 **Planned Strands**

1. **Subscription Creation** - User subscribes to a plan
2. **Payment Processing** - Stripe payment flow
3. **Webhook Handling** - Stripe webhook events
4. **Billing & Invoices** - Invoice generation and management
5. **Coupons & Discounts** - Promotional code system

---

## ⚠️ **Known Issues & TODOs**

From assessment:
- Stripe service is 2,809 lines (needs organization)
- Missing price IDs need investigation
- Webhook handling could be more robust
- Analytics integration gaps

---

## 🚀 **Next Steps**

1. Create BRAID.md overview document
2. Document persistence layer (subscription schemas, Stripe entities)
3. **Focus on** `stripe.go` - break down into logical sections
4. Document webhook handling in detail
5. Document application layer (Stripe API contracts)
6. Document presentation layer (checkout, billing UI)
7. Create strand documents for subscription workflows

---

**See Also**: [Conversion Plan](../../Braid Conversion Plans/BRAID_04_SUBSCRIPTION_BILLING.md) | [Master Index](../../BRAIDS_INDEX.md)

