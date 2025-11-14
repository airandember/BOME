# One-Time Payments (Store) Strand

**Status**: 🚧 PLANNED (Not Yet Implemented)

## Purpose
Handle single purchases for the future BOME store (digital products, merchandise, etc.) as opposed to recurring subscriptions.

## Current State
✅ **Subscriptions are fully implemented** using:
- `customer.subscription.created` webhook → Creates subscription in DB
- `invoice.payment_succeeded` webhook → Grants video access
- Embedded Stripe Checkout for subscription plans

❌ **One-time payments are NOT yet implemented** because:
- We don't currently have a store
- All purchases are subscriptions
- `checkout.session.completed` webhook is not needed yet

## Future Implementation Plan

### When to Implement
When we add a store with products like:
- Individual video purchases
- Digital downloads
- Physical merchandise
- Course access (non-subscription)
- One-time donation options

### Webhook Changes Needed

#### Add checkout.session.completed Handler
```go
// In backend/internal/routes/stripe_webhook_routes.go

case "checkout.session.completed":
    return handleCheckoutSessionCompleted(event, syncService)
```

#### Handler Implementation
```go
func handleCheckoutSessionCompleted(event *stripe.Event, syncService *services.StripeSyncService) error {
    var session stripe.CheckoutSession
    if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
        return err
    }

    log.Printf("🛒 Webhook: Checkout session completed - %s (Mode: %s)", 
        session.ID, session.Mode)

    // Only process one-time payments (subscriptions are handled elsewhere)
    if session.Mode != "payment" {
        log.Printf("ℹ️  Skipping checkout session (mode: %s)", session.Mode)
        return nil
    }

    // Process one-time payment
    return syncService.ProcessOneTimePayment(&session)
}
```

### Database Schema Changes

#### Create `store_purchases` table
```sql
CREATE TABLE store_purchases (
    id SERIAL PRIMARY KEY,
    stripe_checkout_session_id VARCHAR(255) UNIQUE NOT NULL,
    stripe_payment_intent_id VARCHAR(255),
    user_id INTEGER REFERENCES users(id),
    customer_id INTEGER REFERENCES stripe_customers_v2(id),
    
    -- Product info
    product_type VARCHAR(50) NOT NULL, -- 'video', 'merchandise', 'digital', etc.
    product_id VARCHAR(255) NOT NULL,
    product_name VARCHAR(255),
    
    -- Payment info
    amount_paid BIGINT NOT NULL, -- in cents
    currency VARCHAR(3) DEFAULT 'usd',
    status VARCHAR(50) NOT NULL, -- 'complete', 'pending', 'refunded'
    
    -- Fulfillment
    fulfilled BOOLEAN DEFAULT false,
    fulfilled_at TIMESTAMP,
    
    -- Timestamps
    purchased_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    -- Metadata
    metadata JSONB
);

CREATE INDEX idx_store_purchases_user_id ON store_purchases(user_id);
CREATE INDEX idx_store_purchases_session_id ON store_purchases(stripe_checkout_session_id);
CREATE INDEX idx_store_purchases_status ON store_purchases(status);
```

### Frontend Changes

#### Update Checkout Flow
```typescript
// New function in frontend/src/lib/services/store-service.ts
export async function createStoreCheckoutSession(productId: string, productType: 'video' | 'merchandise') {
    const response = await apiRequest('/store/checkout-session', {
        method: 'POST',
        body: JSON.stringify({
            product_id: productId,
            product_type: productType,
            mode: 'payment', // One-time payment, not subscription
            return_url: `${window.location.origin}/store/success?session_id={CHECKOUT_SESSION_ID}`
        })
    });
    
    return response.json();
}
```

#### Create Store Success Page
```svelte
<!-- frontend/src/routes/store/success/+page.svelte -->
<script lang="ts">
    // Similar to /checkout/success but for one-time purchases
    // NO manual processing needed - checkout.session.completed handles it!
</script>
```

### Backend API Changes

#### New Store Routes
```go
// backend/internal/routes/store_routes.go (NEW FILE)

func SetupStoreRoutes(v1 *gin.RouterGroup, db *database.Database) {
    store := v1.Group("/store")
    store.Use(middleware.AuthRequired())
    {
        // Create checkout session for one-time purchase
        store.POST("/checkout-session", createStoreCheckoutSession)
        
        // Get user's purchase history
        store.GET("/purchases", getUserPurchases)
        
        // Get specific purchase details
        store.GET("/purchases/:id", getPurchaseDetails)
    }
}
```

## Key Differences: Subscriptions vs One-Time Payments

| Aspect | Subscriptions | One-Time Payments |
|--------|--------------|-------------------|
| **Checkout Mode** | `mode: 'subscription'` | `mode: 'payment'` |
| **Primary Webhook** | `customer.subscription.created` | `checkout.session.completed` |
| **Payment Webhook** | `invoice.payment_succeeded` | `payment_intent.succeeded` |
| **Recurring** | Yes (monthly/yearly) | No (single charge) |
| **Access Grant** | Ongoing (while active) | Immediate (permanent) |
| **Cancellation** | Can cancel subscription | N/A (already paid) |

## Why checkout.session.completed Matters for Store

### For Subscriptions (Current):
```
checkout.session.completed
    ↓
  (Ignored - subscription webhooks handle everything)
```

### For One-Time Payments (Future):
```
checkout.session.completed
    ↓
  (CRITICAL - this is the ONLY webhook that tells us about the purchase!)
    ↓
  Create record in store_purchases
    ↓
  Grant access to purchased content
    ↓
  Send confirmation email
```

## Testing Checklist (When Implemented)

- [ ] Create one-time payment checkout session
- [ ] Complete payment successfully
- [ ] Verify `checkout.session.completed` webhook received
- [ ] Confirm purchase recorded in `store_purchases` table
- [ ] Verify user can access purchased content
- [ ] Test refund flow
- [ ] Test failed payment handling
- [ ] Ensure subscription checkouts still work (don't break existing flow!)

## Related Files to Update

### Backend
- [ ] `backend/internal/routes/store_routes.go` (NEW)
- [ ] `backend/internal/services/store_service.go` (NEW)
- [ ] `backend/internal/routes/stripe_webhook_routes.go` (add checkout.session.completed)
- [ ] `backend/internal/database/migrations/XXX_create_store_purchases.sql` (NEW)

### Frontend
- [ ] `frontend/src/routes/store/+page.svelte` (NEW)
- [ ] `frontend/src/routes/store/success/+page.svelte` (NEW)
- [ ] `frontend/src/lib/services/store-service.ts` (NEW)

## Notes

- Keep subscription flow unchanged - it works perfectly!
- `checkout.session.completed` is ONLY for mode='payment'
- Always check `session.Mode` in webhook handler to route correctly
- Consider using Stripe's `metadata` field to store product info
- May want to support guest checkout (purchases without account)

## References

- Stripe Docs: [Checkout Sessions](https://stripe.com/docs/api/checkout/sessions)
- Stripe Docs: [One-Time Payments](https://stripe.com/docs/payments/checkout/how-checkout-works#one-time)
- Stripe Webhook: [checkout.session.completed](https://stripe.com/docs/api/events/types#event_types-checkout.session.completed)

---

**Last Updated**: November 14, 2025  
**Status**: Planning Phase  
**Owner**: To be assigned when store development begins

