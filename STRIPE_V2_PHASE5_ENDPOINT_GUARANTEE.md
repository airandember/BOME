# Phase 5 - Stripe Webhook Endpoint Guarantee

**Created**: October 30, 2025  
**Critical**: This endpoint MUST remain unchanged

---

## 🚨 **DO NOT CHANGE THIS URL**

```
POST https://watch.bookofmormonevidence.org/bome-backend/api/v1/webhooks/stripe
```

**Why**: This is configured in your Stripe Dashboard. If we change it, Stripe will fail to deliver webhooks!

---

## ✅ **What We're Changing (Internal Only)**

### **Before Phase 5** (v1 Handler)
```go
webhooks.POST("/stripe", func(c *gin.Context) {
    HandleStripeWebhook(c, stripeService, syncService)  // v1 only
})
```

### **After Phase 5** (v2 Dual-Write)
```go
webhooks.POST("/stripe", func(c *gin.Context) {
    HandleStripeWebhook(c, stripeService, syncServiceV1, syncServiceV2, webhookServiceV2, linkingService)  // v1 + v2
})
```

**The URL stays `/webhooks/stripe`** - only the handler parameters change!

---

## 🔒 **Stripe Dashboard Configuration (Unchanged)**

```
Endpoint ID:      we_1S85P2FpxJJNWdU8kvvk2K7v
Endpoint URL:     https://watch.bookofmormonevidence.org/bome-backend/api/v1/webhooks/stripe
Description:      The MAAAAAIN Event
API version:      2025-07-30.basil
Events:           56 configured
Payload style:    Snapshot
Signing secret:   whsec_···
Status:           Active ✅
```

**Phase 5 will NOT change any of this!**

---

## 🔄 **What Happens Inside the Handler**

### **Current Flow (v1)**
```
1. Stripe sends POST to /webhooks/stripe
2. Validate signature
3. Parse event
4. Route to handler (customer.created, subscription.created, etc.)
5. Write to v1 tables (stripe_customers, stripe_subscriptions)
6. Return 200 OK
```

### **Phase 5 Flow (v1 + v2 Dual-Write)**
```
1. Stripe sends POST to /webhooks/stripe (SAME URL!)
2. Validate signature (SAME LOGIC!)
3. Parse event (SAME LOGIC!)
4. Route to handler (SAME EVENTS!)
5. Write to v1 tables (KEEP AS FALLBACK!)
6. Write to v2 tables (NEW!)
7. Auto-link customer to user (NEW!)
8. Return 200 OK (SAME!)
```

**Only steps 6-7 are new - everything else stays the same!**

---

## 📊 **Implementation Strategy**

### **Step 1: Add V2 Parameters**
```go
func HandleStripeWebhook(
    c *gin.Context,
    stripeService *services.StripeService,           // EXISTING
    syncServiceV1 *services.StripeSyncService,       // EXISTING (rename for clarity)
    syncServiceV2 *services.StripeSyncV2Service,     // NEW
    webhookServiceV2 *services.StripeWebhookServiceV2, // NEW
    linkingService *services.CustomerLinkingService,  // NEW
) {
    // Implementation stays mostly the same
    // Just add v2 handlers alongside v1
}
```

### **Step 2: Dual-Write in Event Handlers**
```go
func handleCustomerCreated(event *stripe.Event, v1, v2, linking) error {
    var customer stripe.Customer
    json.Unmarshal(event.Data.Raw, &customer)
    
    // V1 (existing - keep as fallback)
    if err := v1.UpsertCustomerFromWebhook(&customer); err != nil {
        log.Printf("⚠️  V1 write failed: %v", err)
        // Don't fail webhook
    }
    
    // V2 (new - main path)
    if err := v2.HandleCustomerCreated(&customer); err != nil {
        log.Printf("❌ V2 write failed: %v", err)
        return err  // Fail webhook if v2 fails
    }
    
    return nil
}
```

### **Step 3: Update routes.go**
```go
// Initialize v2 services
stripeSyncV2Service := services.NewStripeSyncV2Service(db)
customerLinkingService := services.NewCustomerLinkingService(db)
webhookServiceV2 := services.NewStripeWebhookServiceV2(
    stripeSyncV2Service,
    customerLinkingService,
    db,
)

// Register webhook with v2 services
webhooks := v1.Group("/webhooks")
{
    webhooks.POST("/stripe", func(c *gin.Context) {
        HandleStripeWebhook(
            c,
            stripeService,
            syncService,             // v1 (existing)
            stripeSyncV2Service,     // v2 (new)
            webhookServiceV2,        // v2 (new)
            customerLinkingService,  // v2 (new)
        )
    })
}
```

---

## ✅ **Verification Checklist**

After Phase 5 implementation:

- [ ] Webhook URL is still `POST /api/v1/webhooks/stripe`
- [ ] Stripe signature validation still works
- [ ] Webhook status endpoint still works (`GET /webhooks/stripe/status`)
- [ ] Webhook logs still work (`GET /webhooks/stripe/logs`)
- [ ] Admin dashboard still displays webhook health
- [ ] Test webhook from Stripe Dashboard sends to correct URL
- [ ] Events are written to v1 tables (fallback)
- [ ] Events are written to v2 tables (primary)
- [ ] Customers are auto-linked to users

---

## 🚨 **If Webhook URL Ever Needs to Change**

**Steps to update Stripe Dashboard**:
1. Go to https://dashboard.stripe.com
2. Developers → Webhooks
3. Click on endpoint: `we_1S85P2FpxJJNWdU8kvvk2K7v`
4. Update "Endpoint URL"
5. Click "Update endpoint"
6. Wait for Stripe to verify (sends test event)

**But for Phase 5, this is NOT needed!** ✅

---

## 📝 **Summary**

| Component | Before Phase 5 | After Phase 5 | Changed? |
|-----------|----------------|---------------|----------|
| **Webhook URL** | `/webhooks/stripe` | `/webhooks/stripe` | ❌ No |
| **Stripe Config** | Active | Active | ❌ No |
| **Signature Validation** | Yes | Yes | ❌ No |
| **Event Routing** | customer.*, subscription.*, etc. | Same | ❌ No |
| **V1 Tables** | Written | Written (fallback) | ✅ Still works |
| **V2 Tables** | Not written | Written (primary) | ✅ New |
| **Customer Linking** | Manual only | Automatic | ✅ New |

**Result**: Stripe keeps sending to the same URL, but now we write to v2 tables AND auto-link customers! 🎉

---

**Phase 5 is 100% backward compatible with your Stripe Dashboard configuration!**

