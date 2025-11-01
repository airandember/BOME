# ✅ Stripe Webhook Dual Secret Implementation Complete

**Date:** October 31, 2025  
**Status:** ✅ Production Ready

---

## 🎯 **What Was Built:**

### **Problem:**
Stripe requires TWO separate webhook secrets when configuring both Snapshot and Thin payload destinations:
- **Destination 1 (Snapshot):** Full V1 events with complete data → `whsec_xxx...`
- **Destination 2 (Thin):** Minimal V2 events requiring API fetch → `whsec_yyy...`

### **Solution:**
Implemented dual webhook secret storage and validation with proper encryption.

---

## 🛠️ **Backend Changes:**

### **1. Database Storage (`secure_settings` table):**

Both secrets are stored **encrypted** using AES-GCM:

```sql
-- Snapshot secret (V1 events)
"stripe_webhook_secret" → encrypted with master key

-- Thin secret (V2 events)  
"stripe_webhook_secret_thin" → encrypted with master key
```

### **2. Service Layer (`backend/internal/services/stripe.go`):**

**Added fields:**
```go
type StripeService struct {
    webhookSecret         string // Snapshot webhook secret (V1 events)
    webhookSecretThin     string // Thin webhook secret (V2 events)
    // ... other fields
}
```

**Load both secrets on startup:**
```go
func (s *StripeService) loadStoredWebhookSecret(db *database.DB) {
    // Load snapshot secret
    s.webhookSecret = decrypted("stripe_webhook_secret")
    
    // Load thin secret
    s.webhookSecretThin = decrypted("stripe_webhook_secret_thin")
}
```

**Validate with correct secret:**
```go
func (s *StripeService) ValidateWebhookSignatureRaw(payload []byte, signature string) error {
    // Try thin secret first (for V2 events)
    if s.webhookSecretThin != "" {
        err := webhook.ValidatePayload(payload, signature, s.webhookSecretThin)
        if err == nil {
            return nil // ✅ Validated with thin secret
        }
    }
    
    // Fall back to snapshot secret
    return webhook.ValidatePayload(payload, signature, s.webhookSecret)
}
```

### **3. Thin Event Handler (`backend/internal/services/stripe_webhook_thin.go`):**

**NEW SERVICE:** Handles V2 thin events by fetching full objects from Stripe API.

```go
type StripeWebhookThinService struct {
    webhookServiceV2 *StripeWebhookServiceV2
}

func (s *StripeWebhookThinService) ProcessThinEvent(rawPayload []byte) error {
    // 1. Parse thin event
    var thinEvent ThinEvent
    json.Unmarshal(rawPayload, &thinEvent)
    
    // 2. Extract object ID
    subscriptionID := s.extractObjectID(event, "subscription")
    
    // 3. Fetch full object from Stripe API
    sub, err := subscription.Get(subscriptionID, nil)
    
    // 4. Delegate to V2 webhook service
    return s.webhookServiceV2.HandleSubscriptionCreated(sub)
}
```

**Supported V2 events:**
- `v2.billing.subscription.created`
- `v2.billing.subscription.updated`
- `v2.billing.subscription.paused`
- `v2.billing.subscription.resumed`
- `v2.core.event_destination.ping`

### **4. Webhook Routes (`backend/internal/routes/stripe_webhook_routes.go`):**

**Updated handler signature:**
```go
func HandleStripeWebhook(
    c *gin.Context,
    stripeService *services.StripeService,
    syncServiceV1 *services.StripeSyncService,
    syncServiceV2 *services.StripeSyncV2Service,
    webhookServiceV2 *services.StripeWebhookServiceV2,
    thinService *services.StripeWebhookThinService, // NEW
) {
    // ... routing logic
}
```

**Flow:**
```go
if isV2Event {
    // Validate with thin secret
    stripeService.ValidateWebhookSignatureRaw(payload, signature)
    
    // Process via thin service
    thinService.ProcessThinEvent(payload)
} else {
    // Validate with snapshot secret
    event := stripeService.ValidateWebhookSignature(payload, signature)
    
    // Process with dual-write
    processV1EventWithDualWrite(event, syncServiceV1, webhookServiceV2)
}
```

### **5. Admin API (`backend/internal/routes/admin_streaming.go`):**

**Updated endpoint to accept type:**
```go
streaming.POST("/stripe/webhook-secret", func(c *gin.Context) {
    var req struct {
        Secret string `json:"secret" binding:"required"`
        Type   string `json:"type"` // "snapshot" or "thin"
    }
    
    // Determine database key
    var dbKey string
    if req.Type == "thin" {
        dbKey = "stripe_webhook_secret_thin"
    } else {
        dbKey = "stripe_webhook_secret" // snapshot (default)
    }
    
    // Encrypt and store
    db.SetSecureSetting(dbKey, encryptedSecret)
})
```

---

## 🎨 **Frontend Changes:**

### **1. Admin UI (`frontend/src/routes/admin/streaming/stripe/setup/+page.svelte`):**

**Added two separate secret input fields:**

**Destination 1 - Snapshot:**
```svelte
<div class="setting-item">
    <div class="setting-header">
        <h4>Webhook Secret - Snapshot (V1)</h4>
        <span class="setting-type secure">Secure</span>
        <span class="setting-badge primary">Destination 1</span>
    </div>
    <input 
        type="password" 
        placeholder="whsec_... (from Snapshot destination)" 
        bind:value={webhookSecretSnapshot}
    />
    <button onclick={saveWebhookSecretSnapshot}>Save Snapshot Secret</button>
</div>
```

**Destination 2 - Thin:**
```svelte
<div class="setting-item">
    <div class="setting-header">
        <h4>Webhook Secret - Thin (V2)</h4>
        <span class="setting-type secure">Secure</span>
        <span class="setting-badge secondary">Destination 2</span>
    </div>
    <input 
        type="password" 
        placeholder="whsec_... (from Thin destination)" 
        bind:value={webhookSecretThin}
    />
    <button onclick={saveWebhookSecretThin}>Save Thin Secret</button>
</div>
```

**Visual badges:**
- **Destination 1:** Blue gradient badge
- **Destination 2:** Purple gradient badge

### **2. Save Functions:**

```typescript
async function saveWebhookSecretSnapshot() {
    await apiRequest('/admin/streaming/stripe/webhook-secret', {
        method: 'POST',
        body: JSON.stringify({ 
            secret: webhookSecretSnapshot,
            type: 'snapshot'
        })
    });
}

async function saveWebhookSecretThin() {
    await apiRequest('/admin/streaming/stripe/webhook-secret', {
        method: 'POST',
        body: JSON.stringify({ 
            secret: webhookSecretThin,
            type: 'thin'
        })
    });
}
```

---

## 📊 **Webhook Flow:**

### **Snapshot Events (V1):**

```
Stripe sends webhook → V1 event (customer.created, etc.)
    ↓
Your endpoint: /api/v1/webhooks/stripe
    ↓
Validate signature with stripe_webhook_secret
    ↓
Parse full event object (all data included)
    ↓
Dual-write to v1 + v2 tables
    ↓
Return 200 OK
```

### **Thin Events (V2):**

```
Stripe sends webhook → V2 event (v2.billing.subscription.created, etc.)
    ↓
Your endpoint: /api/v1/webhooks/stripe
    ↓
Validate signature with stripe_webhook_secret_thin
    ↓
Parse thin event (minimal data)
    ↓
Fetch full object from Stripe API
    ↓
Delegate to V2 webhook service
    ↓
Write to v2 tables
    ↓
Return 200 OK
```

---

## 🔒 **Security Features:**

1. **✅ Encrypted Storage**
   - Both secrets encrypted with AES-GCM
   - Master key from environment (`MASTER_ENCRYPTION_KEY`)
   - Never exposed via API

2. **✅ Signature Validation**
   - Every webhook validated before processing
   - Invalid signatures = automatic 400 rejection
   - Automatic fallback between secrets

3. **✅ Write-Only Pattern**
   - Frontend can update secrets
   - Frontend can't read secrets back
   - Backend reads for validation only

4. **✅ Format Validation**
   - Must start with `whsec_`
   - Validated before storage

---

## 📋 **Stripe Dashboard Configuration:**

### **Step 1: Create Snapshot Destination**

1. Go to: https://dashboard.stripe.com/webhooks
2. Click "Add endpoint"
3. **URL:** `https://watch.bookofmormonevidence.org/bome-backend/api/v1/webhooks/stripe`
4. **Destination name:** BOME - BETA (Snapshot)
5. **Payload style:** Snapshot
6. **API version:** 2025-10-29.clover
7. Select events: `customer.*`, `customer.subscription.*`, `invoice.*`, `checkout.*`
8. **Copy the signing secret** (starts with `whsec_`)

### **Step 2: Create Thin Destination**

1. Click "Add endpoint" again
2. **URL:** `https://watch.bookofmormonevidence.org/bome-backend/api/v1/webhooks/stripe` (SAME URL!)
3. **Destination name:** BOME - BETA (Thin)
4. **Payload style:** Thin
5. **API version:** Unversioned
6. Select events: `v2.billing.*`, `v2.core.*`
7. **Copy the signing secret** (different from snapshot!)

### **Step 3: Add Secrets to Admin UI**

1. Go to: `/admin/streaming/stripe/setup`
2. Scroll to "Webhook Secret - Snapshot (V1)"
3. Paste the snapshot secret → Save
4. Scroll to "Webhook Secret - Thin (V2)"
5. Paste the thin secret → Save
6. ✅ Both secrets are now encrypted and stored

---

## 🧪 **Testing:**

### **Test Snapshot Webhooks:**

```bash
# In Stripe Dashboard
1. Go to your Snapshot webhook endpoint
2. Click "Send test webhook"
3. Select: customer.created
4. Send

# Expected response:
{
  "received": true,
  "processed": true,
  "type": "v1_event",
  "dual_write": "v1+v2"
}
```

### **Test Thin Webhooks:**

```bash
# In Stripe Dashboard
1. Go to your Thin webhook endpoint
2. Click "Send test webhook"
3. Select: v2.core.event_destination.ping
4. Send

# Expected response:
{
  "received": true,
  "processed": true,
  "type": "v2_thin_event",
  "fetched_full_object": true
}
```

### **Check Backend Logs:**

```
✅ Loaded snapshot webhook secret from database
✅ Loaded thin webhook secret from database
📨 Webhook received: customer.created
✅ Webhook: v1 event signature validated successfully
✅ Webhook: Successfully processed v1 event (dual-write to v1 + v2)
```

```
📨 Webhook received: v2.billing.subscription.created
✅ V2 thin webhook signature validated with thin secret
🔷 [Thin Webhook] Fetching full subscription: sub_xxxxx
✅ Webhook: Successfully processed v2 thin event
```

---

## ✅ **Benefits:**

1. **✅ Future-Proof**: Ready for Stripe's V2 event system
2. **✅ Backward Compatible**: V1 events still work perfectly
3. **✅ Single Endpoint**: Both event types use same URL
4. **✅ Auto-Routing**: Backend automatically detects and routes
5. **✅ Secure**: Both secrets encrypted at rest
6. **✅ User-Friendly**: Clear UI with visual badges
7. **✅ Flexible**: Can use one or both destinations

---

## 📊 **Database Schema:**

```sql
-- secure_settings table
CREATE TABLE secure_settings (
    id SERIAL PRIMARY KEY,
    key VARCHAR(255) UNIQUE NOT NULL,
    value TEXT NOT NULL, -- Encrypted
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Example data (encrypted values)
INSERT INTO secure_settings (key, value) VALUES
('stripe_webhook_secret', 'ek+Q0ZonOZs2NK7b...'), -- Snapshot secret (encrypted)
('stripe_webhook_secret_thin', 'dL9P2XpnQAt3ML8c...'); -- Thin secret (encrypted)
```

---

## 🎯 **What's Next:**

1. ✅ **Configure both webhook destinations in Stripe**
2. ✅ **Add both secrets to admin UI**
3. ✅ **Test both event types**
4. ✅ **Monitor webhook logs in admin dashboard**
5. ✅ **Watch for real events from Stripe**

---

## 🚀 **Status:**

- ✅ Backend implementation complete
- ✅ Frontend UI complete
- ✅ Encryption/security complete
- ✅ Validation logic complete
- ✅ Thin event processing complete
- ✅ Documentation complete

**Ready to configure in production!** 🎉

---

**Same endpoint URL for both:**
```
https://watch.bookofmormonevidence.org/bome-backend/api/v1/webhooks/stripe
```

**Your system automatically knows which secret to use based on the event type!** 🧠✨

