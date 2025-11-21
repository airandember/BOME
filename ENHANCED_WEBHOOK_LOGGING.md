# Enhanced Webhook Event Logging

**Date:** November 21, 2025  
**Status:** ✅ IMPLEMENTED & READY FOR DEPLOYMENT

---

## 🎯 Overview

We've enhanced the `webhook_events` table to capture detailed information from Stripe webhooks, enabling rich analytics, detailed event tracking, and user-specific webhook history.

---

## 📋 New Columns Added

### **Core Identifiers**

| Column | Type | Description | Example |
|--------|------|-------------|---------|
| `stripe_event_id` | VARCHAR(255) | Unique Stripe event ID | `evt_1SVx2aFpxJJNWdU8n67rCYuN` |
| `stripe_object_id` | VARCHAR(255) | ID of affected object | `sub_1Quyv4FpxJJNWdU8HeoWqZIw` |
| `stripe_object_type` | VARCHAR(50) | Type of object | `subscription`, `customer`, `invoice` |

### **User Identification**

| Column | Type | Description | Example |
|--------|------|-------------|---------|
| `user_id` | INTEGER | BOME user ID | `9042` |
| `user_email` | VARCHAR(255) | User's email | `srnavasettles@gmail.com` |
| `customer_id` | VARCHAR(255) | Stripe customer ID | `cus_Roc9Ab6KeG3x4v` |

### **Subscription Data**

| Column | Type | Description | Example |
|--------|------|-------------|---------|
| `subscription_id` | VARCHAR(255) | Stripe subscription ID | `sub_1Quyv4FpxJJNWdU8HeoWqZIw` |
| `subscription_status` | VARCHAR(50) | Subscription status | `active`, `unpaid`, `canceled` |

### **Payment Data**

| Column | Type | Description | Example |
|--------|------|-------------|---------|
| `invoice_id` | VARCHAR(255) | Stripe invoice ID | `in_1SVx2ZFpxJJNWdU8aGaNmOKA` |
| `amount_cents` | INTEGER | Amount in cents | `997` ($9.97) |
| `currency` | VARCHAR(3) | Currency code | `USD` |
| `payment_status` | VARCHAR(50) | Payment status | `paid`, `unpaid`, `pending` |

### **Metadata**

| Column | Type | Description | Example |
|--------|------|-------------|---------|
| `event_data` | JSONB | Full webhook payload | `{...}` |
| `api_version` | VARCHAR(50) | Stripe API version | `2025-10-29.clover` |
| `livemode` | BOOLEAN | Production vs test mode | `true` |
| `description` | TEXT | Human-readable description | `srnavasettles@gmail.com's subscription updated (status: active)` |

---

## 🗄️ Database Migration

**File:** `backend/migrations/039_enhance_webhook_events.sql`

**What it does:**
- ✅ Adds 16 new columns to `webhook_events` table
- ✅ Creates 8 new indexes for fast queries
- ✅ Adds table and column comments for documentation
- ✅ Uses `IF NOT EXISTS` for safe deployment

**To apply:**
```bash
psql -d bome_db -f backend/migrations/039_enhance_webhook_events.sql
```

---

## 💻 Code Changes

### **1. Enhanced Logging Function**

**File:** `backend/internal/routes/stripe_webhook_routes.go`

**New function:** `logWebhookEventToDB()`

**Features:**
- ✅ Parses event data based on event type
- ✅ Extracts relevant fields (customer, subscription, invoice, etc.)
- ✅ Looks up user ID from customer ID automatically
- ✅ Generates human-readable descriptions
- ✅ Stores full JSON payload for detailed inspection
- ✅ Handles all major Stripe event types

**Event types supported:**
- `customer.*` (created, updated, deleted)
- `customer.subscription.*` (created, updated, deleted)
- `invoice.payment_succeeded` / `invoice.payment_failed`
- `product.*` (created, updated)
- `price.*` (created, updated)

---

## 📊 Example Data Captured

### **Subscription Updated Event**

From your provided webhook:

```json
{
  "stripe_event_id": "evt_1SVx2aFpxJJNWdU8n67rCYuN",
  "stripe_object_id": "sub_1Quyv4FpxJJNWdU8HeoWqZIw",
  "stripe_object_type": "subscription",
  "user_id": 9042,
  "user_email": "srnavasettles@gmail.com",
  "customer_id": "cus_Roc9Ab6KeG3x4v",
  "subscription_id": "sub_1Quyv4FpxJJNWdU8HeoWqZIw",
  "subscription_status": "unpaid",
  "amount_cents": null,
  "currency": "usd",
  "payment_status": null,
  "api_version": "2025-10-29.clover",
  "livemode": true,
  "description": "srnavasettles@gmail.com's subscription updated (status: unpaid)",
  "event_data": { /* Full webhook payload */ }
}
```

---

## 🎨 Frontend Dashboard Enhancement

### **Current UI (Before)**

```
Timestamp         | Event Type                      | Status
------------------|---------------------------------|--------
2025-11-21 4:13pm | customer.subscription.updated   | success
2025-11-21 4:13pm | customer.updated                | success
```

**Problems:**
- ❌ No user identification
- ❌ No way to see details
- ❌ Can't filter by user
- ❌ Can't see what changed

---

### **Proposed Enhanced UI (After)**

```
Timestamp         | User                            | Event Type                      | Status  | Actions
------------------|---------------------------------|---------------------------------|---------|--------
2025-11-21 4:13pm | srnavasettles@gmail.com        | customer.subscription.updated   | success | [Details]
                  | (User ID: 9042)                 | → Status changed to unpaid      |         |
                  |                                 |                                 |         |
2025-11-21 4:13pm | srnavasettles@gmail.com        | customer.updated                | success | [Details]
                  | (User ID: 9042)                 | → Customer account updated      |         |
```

---

### **"Details" Modal**

When clicking `[Details]` button:

```
┌─────────────────────────────────────────────────────────┐
│ Webhook Event Details                                   │
├─────────────────────────────────────────────────────────┤
│                                                           │
│ Event ID: evt_1SVx2aFpxJJNWdU8n67rCYuN                  │
│ Timestamp: 2025-11-21 4:13:40 PM                         │
│                                                           │
│ User Information:                                         │
│   Email: srnavasettles@gmail.com                         │
│   User ID: 9042                                          │
│   Customer ID: cus_Roc9Ab6KeG3x4v                        │
│                                                           │
│ Event Details:                                            │
│   Type: customer.subscription.updated                     │
│   Description: srnavasettles@gmail.com's subscription    │
│                updated (status: unpaid)                   │
│                                                           │
│ Subscription:                                             │
│   ID: sub_1Quyv4FpxJJNWdU8HeoWqZIw                       │
│   Status: unpaid → active                                 │
│   Amount: $9.97 USD (monthly)                            │
│                                                           │
│ Technical:                                                │
│   API Version: 2025-10-29.clover                          │
│   Live Mode: Yes                                          │
│   Response Time: 245ms                                    │
│   Payload Size: 4.2 KB                                    │
│                                                           │
│ Full Event Data:                                          │
│ [View JSON] [Download] [Copy]                            │
│                                                           │
└─────────────────────────────────────────────────────────┘
```

---

## 🔍 Useful Queries

### **Find All Events for a User**

```sql
SELECT 
    created_at,
    event_type,
    description,
    subscription_status,
    payment_status,
    amount_cents
FROM webhook_events
WHERE user_id = 9042
ORDER BY created_at DESC;
```

### **Find All Subscription Changes**

```sql
SELECT 
    user_email,
    subscription_id,
    subscription_status,
    description,
    created_at
FROM webhook_events
WHERE event_type LIKE 'customer.subscription.%'
AND stripe_object_type = 'subscription'
ORDER BY created_at DESC;
```

### **Find Failed Payments**

```sql
SELECT 
    user_email,
    invoice_id,
    amount_cents,
    currency,
    description,
    created_at
FROM webhook_events
WHERE event_type = 'invoice.payment_failed'
ORDER BY created_at DESC;
```

### **Find All Events for a Subscription**

```sql
SELECT 
    created_at,
    event_type,
    user_email,
    subscription_status,
    description
FROM webhook_events
WHERE subscription_id = 'sub_1Quyv4FpxJJNWdU8HeoWqZIw'
ORDER BY created_at ASC;
```

### **Revenue by Day**

```sql
SELECT 
    DATE(created_at) as date,
    COUNT(*) as payments,
    SUM(amount_cents) / 100.0 as revenue_usd
FROM webhook_events
WHERE event_type = 'invoice.payment_succeeded'
AND livemode = true
GROUP BY DATE(created_at)
ORDER BY date DESC;
```

---

## 🎯 Admin Dashboard Features to Build

### **1. Webhook Events Tab (Enhanced)**

**Current Table Columns:**
- ❌ Timestamp
- ❌ Event Type  
- ❌ Status

**New Table Columns:**
- ✅ Timestamp
- ✅ **User** (email + user ID link)
- ✅ Event Type
- ✅ **Description** (human-readable)
- ✅ Status
- ✅ **Actions** ([Details] button)

---

### **2. User-Specific Webhook History**

**Location:** Admin → Users → [Select User] → Webhooks Tab

**Features:**
- View all Stripe events for a specific user
- Timeline view showing subscription changes
- Payment history from webhooks
- Quick troubleshooting for user issues

---

### **3. Real-Time Event Monitor**

**Features:**
- Live feed of incoming webhooks
- Color-coded by event type
- Alert on failed events
- Filter by user, event type, status

---

### **4. Webhook Analytics Dashboard**

**Metrics:**
- Events per day/hour
- Success rate
- Most common event types
- Failed event breakdown
- Average response time
- User activity heatmap

---

## 📈 Benefits

### **For Support Team:**
- ✅ Quickly see all Stripe events for a user
- ✅ Debug subscription issues faster
- ✅ Verify payment processing
- ✅ Track customer lifecycle

### **For Development:**
- ✅ Detailed webhook history for debugging
- ✅ Full payload stored for inspection
- ✅ User correlation built-in
- ✅ Easy to query and analyze

### **For Business:**
- ✅ Revenue tracking from webhook data
- ✅ Subscription churn analysis
- ✅ Payment failure trends
- ✅ Customer behavior insights

---

## 🚀 Next Steps

### **1. Apply Migration**
```bash
cd backend
psql -d bome_db -f migrations/039_enhance_webhook_events.sql
```

### **2. Deploy Backend**
- ✅ Code already built successfully
- ⏳ Deploy updated backend
- ⏳ Restart server

### **3. Test with Live Webhook**
- ⏳ Trigger a subscription event
- ⏳ Verify data is captured
- ⏳ Check `user_email` and `description` fields

### **4. Build Frontend UI**
- ⏳ Add "User" column to webhook events table
- ⏳ Implement "Details" modal
- ⏳ Add user filtering
- ⏳ Create user-specific webhook history view

---

## 📝 Example Frontend Code

### **Enhanced Table Column (User)**

```svelte
<td class="user-cell">
    {#if event.user_email}
        <div class="user-info">
            <a href="/admin/users/{event.user_id}" class="user-email">
                {event.user_email}
            </a>
            <span class="user-id">ID: {event.user_id}</span>
        </div>
    {:else}
        <span class="no-user">System Event</span>
    {/if}
</td>
```

### **Details Modal**

```svelte
<script>
    async function showEventDetails(eventId) {
        const response = await fetch(`/api/admin/webhooks/${eventId}`);
        const event = await response.json();
        
        // Show modal with full event details
        openModal({
            title: 'Webhook Event Details',
            event: event,
            jsonData: event.event_data
        });
    }
</script>

<button class="btn-details" on:click={() => showEventDetails(event.id)}>
    Details
</button>
```

---

## ✅ Summary

### **What Was Added:**
- 16 new columns to `webhook_events` table
- Enhanced logging function with automatic data extraction
- User identification from Stripe customer ID
- Human-readable descriptions
- Full JSON payload storage
- 8 new database indexes

### **What You Get:**
- Rich webhook event history
- User-specific event tracking
- Detailed troubleshooting data
- Revenue and payment analytics
- Better support capabilities
- Easier debugging

### **Next Action:**
Apply the migration and restart your backend to start capturing enhanced webhook data!

---

**Status:** ✅ CODE READY - MIGRATION PENDING  
**Risk:** 🟢 LOW (additive changes only, no breaking changes)  
**Impact:** 🟢 HIGH (major improvement in observability)

Last Updated: 2025-11-21

