# 🔗 Stripe-Subscription Integration

This implementation provides seamless integration between the subscription management system and Stripe payment processing, allowing automatic creation of Stripe products and prices when subscription plans are created.

## 🏗️ Architecture Overview

### Backend Components

1. **SubscriptionPlanStripeService** (`backend/internal/services/subscription_plans_stripe.go`)
   - Extends the existing `SubscriptionPlanService` with Stripe integration
   - Automatically creates Stripe products and prices when creating subscription plans
   - Provides sync functionality for existing plans
   - Returns integration status for monitoring

2. **Stripe Integration Routes** (`backend/internal/routes/subscription_plans_stripe.go`)
   - `POST /api/v1/admin/streaming/subscription-plans/stripe` - Create plan with optional Stripe integration
   - `POST /api/v1/admin/streaming/subscription-plans/stripe/:id/sync` - Sync existing plan with Stripe
   - `GET /api/v1/admin/streaming/subscription-plans/stripe/:id/stripe-status` - Get integration status

### Frontend Components

1. **StripeSubscriptionIntegrationService** (`frontend/src/lib/services/stripe-subscription-integration.ts`)
   - Handles API communication for Stripe-integrated plan operations
   - Checks Stripe connection status
   - Manages plan creation with automatic Stripe integration

2. **PlanModalWithStripe** (`frontend/src/routes/admin/streaming/subscriptions/components/PlanModalWithStripe.svelte`)
   - Enhanced plan creation modal with Stripe integration options
   - Real-time preview of Stripe product/price configuration
   - Automatic Stripe connection detection

3. **StripeIntegrationStatus** (`frontend/src/routes/admin/streaming/subscriptions/components/StripeIntegrationStatus.svelte`)
   - Shows integration status for existing plans
   - Provides sync functionality for plans without Stripe integration
   - Visual indicators for sync status

## 🚀 Features

### ✅ Automatic Integration
- **Auto-Create Stripe Entities**: When creating subscription plans, optionally auto-create corresponding Stripe products and prices
- **Metadata Mapping**: Automatically maps plan data to Stripe metadata for tracking
- **Price Conversion**: Automatically converts dollar amounts to cents for Stripe

### 🔄 Sync Capabilities
- **Retroactive Sync**: Sync existing subscription plans with Stripe
- **Status Monitoring**: Real-time status of integration for each plan
- **Error Handling**: Graceful fallback when Stripe is unavailable

### 🎛️ User Experience
- **Smart Detection**: Automatically detects if Stripe is connected
- **Visual Feedback**: Clear indicators of integration status
- **Preview Mode**: Shows exactly what will be created in Stripe
- **Fallback Support**: Works with or without Stripe integration

## 📊 Integration Flow

### Creating a New Plan with Stripe Integration

1. **User Action**: User creates a new subscription plan with "Auto-create Stripe" enabled
2. **Backend Process**:
   ```
   Plan Creation Request
   ↓
   Create Stripe Product (if enabled)
   ↓
   Create Stripe Price (if product creation succeeded)
   ↓
   Store Stripe IDs in subscription plan
   ↓
   Create subscription plan in database
   ↓
   Return plan with Stripe integration status
   ```
3. **Frontend Response**: Shows success message with Stripe integration confirmation

### Syncing Existing Plans

1. **User Action**: User clicks "Sync with Stripe" on an existing plan
2. **Backend Process**:
   ```
   Sync Request
   ↓
   Check if plan already has Stripe entities
   ↓
   Create missing Stripe product/price
   ↓
   Update plan with new Stripe IDs
   ↓
   Return updated plan
   ```
3. **Frontend Response**: Updates integration status display

## 🔧 Configuration

### Backend Setup
The integration is automatically registered in `backend/internal/routes/routes.go`:

```go
// Initialize Stripe-integrated subscription service
subscriptionPlanStripeService := services.NewSubscriptionPlanStripeService(db, stripeService)

// Register routes
SetupSubscriptionPlanStripeRoutes(admin, stripeService, subscriptionPlanStripeService)
```

### Frontend Usage

#### Using the Enhanced Modal
```svelte
<script>
  import PlanModalWithStripe from './components/PlanModalWithStripe.svelte';
  
  let showModal = false;
</script>

<PlanModalWithStripe 
  bind:show={showModal}
  on:planCreated={handlePlanCreated}
/>
```

#### Displaying Integration Status
```svelte
<script>
  import StripeIntegrationStatus from './components/StripeIntegrationStatus.svelte';
</script>

<StripeIntegrationStatus 
  planId={plan.id} 
  planName={plan.name} 
/>
```

## 🛡️ Error Handling

### Backend Resilience
- **Stripe Unavailable**: Plans are created without Stripe integration if service is down
- **Partial Failures**: If Stripe product creation succeeds but price creation fails, the system logs the issue and continues
- **Cleanup Logic**: Future enhancement to clean up orphaned Stripe entities

### Frontend Graceful Degradation
- **Connection Detection**: Automatically detects Stripe availability
- **Fallback UI**: Shows appropriate messaging when Stripe is unavailable
- **Error Messages**: Clear, actionable error messages for users

## 🔍 Monitoring & Status

### Integration Status Levels
- **✅ Fully Synced**: Plan has both Stripe product and price
- **⚠️ Partially Synced**: Plan has some but not all Stripe entities
- **❌ Not Synced**: Plan has no Stripe integration

### Status Information
- Stripe Product ID (last 8 characters shown for security)
- Stripe Price ID (last 8 characters shown for security)
- Last sync timestamp
- Sync actions available

## 🚀 Benefits

1. **Unified Workflow**: Create subscription plans and Stripe entities in one action
2. **Reduced Errors**: Eliminates manual Stripe ID entry
3. **Consistency**: Ensures subscription plans and Stripe entities stay in sync
4. **Flexibility**: Works with or without Stripe integration
5. **Monitoring**: Clear visibility into integration status
6. **Scalability**: Easy to extend with additional Stripe features

## 🔮 Future Enhancements

- **Webhook Integration**: Sync changes from Stripe back to subscription plans
- **Bulk Sync**: Sync multiple plans at once
- **Advanced Metadata**: More sophisticated data mapping
- **Stripe Product Management**: Direct editing of Stripe products from the admin panel
- **Integration Health Dashboard**: System-wide view of Stripe integration status

## 📝 API Reference

### Create Plan with Stripe Integration
```http
POST /api/v1/admin/streaming/subscription-plans/stripe
Content-Type: application/json

{
  "name": "Premium Monthly",
  "description": "Full access to all content",
  "price": 19.99,
  "currency": "USD",
  "interval": "month",
  "interval_count": 1,
  "auto_create_stripe": true,
  "features": ["HD Streaming", "Priority Support"],
  "is_active": true,
  "sub_type": "stnd"
}
```

### Sync Plan with Stripe
```http
POST /api/v1/admin/streaming/subscription-plans/stripe/{planId}/sync
```

### Get Integration Status
```http
GET /api/v1/admin/streaming/subscription-plans/stripe/{planId}/stripe-status
```

Response:
```json
{
  "has_stripe_product": true,
  "has_stripe_price": true,
  "stripe_price_id": "price_1234567890",
  "sync_status": "synced"
}
```

This integration provides a complete, production-ready solution for managing subscription plans with automatic Stripe integration, ensuring your subscription system and payment processing stay perfectly synchronized. 