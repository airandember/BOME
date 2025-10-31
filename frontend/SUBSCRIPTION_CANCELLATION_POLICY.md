# Subscription Cancellation Policy

## Overview
Users **cannot self-cancel subscriptions** through the frontend UI. All subscription changes (cancellations, plan changes, etc.) must be handled through **support contact**.

## Design Decision
This policy ensures:
1. ✅ **Human touch** - Users speak to support before canceling
2. ✅ **Retention opportunity** - Support can address concerns and offer alternatives
3. ✅ **Proper handling** - Complex scenarios (multiple subscriptions, refunds, etc.) are handled correctly
4. ✅ **Security** - Prevents accidental or unauthorized cancellations

## User Experience

### Single Active Subscription
- User sees their subscription details
- **No cancel button** is shown
- User is directed to contact support (via `public_settings` table) for any changes

### Multiple Active Subscriptions
- User sees a **warning banner** with support contact information
- User can select which subscription they want to **keep** (for support reference)
- Support contact info is pulled from `public_settings` table:
  - Email (with pre-filled subject line)
  - Phone
  - URL (support portal link)
  - Hours
  - Custom message

### Support Contact Information
Configured in the admin panel at `/admin/system/support` and stored in the `public_settings` table:
- `support_email` - Support email address
- `support_phone` - Support phone number
- `support_url` - Support portal URL
- `support_hours` - Support operating hours
- `support_message` - Custom message to display

## Implementation Details

### Frontend Components Modified

#### `SubscriptionCard.svelte`
- **Removed**: `onCancel` prop and `handleCancel()` function
- **Removed**: All cancel button UI (lines 105-123)
- **Kept**: Status badges, renewal info, and subscription details
- **Result**: Clean, read-only subscription display

#### `SubscriptionManagement.svelte`
- **Removed**: `SubscriptionCancelModal` import
- **Removed**: `showCancelModal`, `isProcessing`, `subscriptionsToCancel`, `subscriptionToKeep` state
- **Removed**: `handleCancelClick()`, `closeCancelModal()`, `handleConfirmCancel()` functions
- **Removed**: `onCancel` props passed to `SubscriptionCard`
- **Removed**: Cancel modal at bottom of component
- **Kept**: Support settings integration and display
- **Result**: Read-only subscription management with support referral

### Backend Routes (Not Exposed to Frontend)
The backend still has cancellation routes for **admin/support use only**:
- `DELETE /api/v1/user/subscriptions/:subscription_id/cancel` - Admin only
- `POST /api/v1/admin/user/:user_id/subscriptions/cancel-multiple` - Admin only

### Future Enhancements (If Needed)
If self-service cancellation is ever required:
1. Add a "Request Cancellation" button that opens a support ticket
2. Require confirmation (email/SMS) for cancellation requests
3. Implement a "cooling off" period before cancellation takes effect
4. Add exit survey to understand cancellation reasons

## Testing Checklist
- [ ] Single subscription displays without cancel button
- [ ] Multiple subscriptions show support contact info
- [ ] Support email link pre-fills subject and body correctly
- [ ] Support contact info pulls from `public_settings` correctly
- [ ] Admin can still cancel subscriptions via admin panel
- [ ] Canceled subscriptions show in "History" section
- [ ] No console errors related to missing `onCancel` props

## Related Files
- `frontend/src/lib/components/SubscriptionCard.svelte`
- `frontend/src/lib/components/SubscriptionManagement.svelte`
- `frontend/src/lib/services/support-settings-service.ts`
- `backend/internal/services/support_settings_service.go`
- `backend/internal/routes/support_settings_routes.go`
- `backend/migrations/051_add_support_settings.sql`

---

**Last Updated**: October 31, 2024
**Policy Status**: ✅ Active
**Next Review**: Quarterly (or as needed based on user feedback)

