# 🧬 Communication Braid - Frontend
**Svelte5 UI for email preferences and notifications**

---

## 🔗 **Cross-Repository Braid**

> **⚠️ IMPORTANT**: This is the **frontend portion** of the Communication Braid.  
> **Backend portion**: See `_braids/communication/backend/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## 📋 **Frontend Overview**

**Purpose**: User interface for notification preferences and email management  
**Technology**: Svelte 5, TypeScript, TailwindCSS  
**Entry Points**: `/account/notifications`, `/admin/streaming/email`  
**State Management**: Svelte stores for notification state

---

## 🎯 **Key Features**

### **1. User Notification Preferences**:
- Email notification toggles
- Notification frequency (immediate, daily, weekly)
- Category preferences (security, marketing, updates)
- Unsubscribe management

### **2. In-App Notifications** (Future):
- Toast notifications
- Notification center
- Read/unread status
- Notification history

### **3. Admin Email Management**:
- View email logs
- Send manual emails
- Template management
- Delivery analytics

---

## 📄 **Frontend Pages**

### **1. Email Preferences Page** (`/account/settings#notifications`)
**File**: Integrated in `frontend/src/routes/account/settings/+page.svelte`

**Features**:
- Toggle email notifications
- Select notification types
- Set frequency preferences
- Unsubscribe options
- Save preferences button

**Example UI**:
```svelte
<div class="email-preferences">
  <h2>Email Notifications</h2>
  
  <!-- Master toggle -->
  <label>
    <input type="checkbox" bind:checked={emailEnabled} />
    Enable email notifications
  </label>
  
  {#if emailEnabled}
    <!-- Notification categories -->
    <div class="categories">
      <label>
        <input type="checkbox" bind:checked={preferences.security} />
        Security alerts (recommended)
      </label>
      
      <label>
        <input type="checkbox" bind:checked={preferences.account} />
        Account updates
      </label>
      
      <label>
        <input type="checkbox" bind:checked={preferences.marketing} />
        Product updates and news
      </label>
      
      <label>
        <input type="checkbox" bind:checked={preferences.billing} />
        Billing and subscription
      </label>
    </div>
    
    <!-- Frequency -->
    <div class="frequency">
      <label for="frequency">Email frequency:</label>
      <select id="frequency" bind:value={preferences.frequency}>
        <option value="immediate">Immediate</option>
        <option value="daily">Daily digest</option>
        <option value="weekly">Weekly digest</option>
      </select>
    </div>
  {/if}
  
  <button on:click={savePreferences}>
    Save Preferences
  </button>
</div>
```

---

### **2. Admin Email Management** (`/admin/streaming/email`)
**File**: `frontend/src/routes/admin/streaming/email/+page.svelte`

**Features**:
- Email delivery dashboard
- Recent email logs
- Failed delivery alerts
- Send manual email form
- Template preview

**UI Sections**:
- **Stats Cards**: Sent, delivered, bounced, failed
- **Recent Emails**: Table with status
- **Quick Actions**: Send email, view templates
- **Charts**: Email volume over time

---

## 🧩 **Frontend Components**

### **NotificationToast Component** (Future)
**Purpose**: Display in-app notifications

**Features**:
- Auto-dismiss after 5 seconds
- Multiple toast types (success, error, info, warning)
- Action buttons
- Stacking notifications

**Usage**:
```svelte
<script>
  import { notifications } from '$lib/stores/notifications';
</script>

{#each $notifications as notification}
  <Toast
    type={notification.type}
    message={notification.message}
    onDismiss={() => notifications.dismiss(notification.id)}
  />
{/each}
```

---

### **EmailPreferencePanel Component**
**Purpose**: Reusable preference settings

**Props**:
```typescript
interface EmailPreferencePanelProps {
  preferences: EmailPreferences;
  onChange: (prefs: EmailPreferences) => void;
  disabled?: boolean;
}
```

---

## 🗃️ **Frontend Stores**

### **Email Preferences Store** (`$lib/stores/emailPreferences.ts`)
**Purpose**: Manage user email preferences

**State**:
```typescript
interface EmailPreferences {
  enabled: boolean;
  security: boolean;
  account: boolean;
  marketing: boolean;
  billing: boolean;
  frequency: 'immediate' | 'daily' | 'weekly';
}

interface EmailPreferencesState {
  preferences: EmailPreferences;
  loading: boolean;
  error: string | null;
}
```

**Methods**:
```typescript
export const emailPreferences = {
  async load() {
    // GET /api/v1/users/email-preferences
  },
  
  async update(prefs: EmailPreferences) {
    // PUT /api/v1/users/email-preferences
  },
  
  async unsubscribeAll() {
    // POST /api/v1/users/unsubscribe
  }
};
```

---

### **Notification Store** (`$lib/stores/notifications.ts`) (Future)
**Purpose**: In-app notification system

**State**:
```typescript
interface Notification {
  id: string;
  type: 'success' | 'error' | 'info' | 'warning';
  message: string;
  action?: {
    label: string;
    onClick: () => void;
  };
  duration?: number; // ms, default 5000
}

interface NotificationsState {
  items: Notification[];
}
```

**Methods**:
```typescript
export const notifications = {
  success(message: string) {
    this.add({ type: 'success', message });
  },
  
  error(message: string) {
    this.add({ type: 'error', message });
  },
  
  info(message: string) {
    this.add({ type: 'info', message });
  },
  
  dismiss(id: string) {
    // Remove notification
  }
};
```

**Usage**:
```typescript
import { notifications } from '$lib/stores/notifications';

// Show success
notifications.success('Profile updated!');

// Show error
notifications.error('Failed to save changes');

// Show with action
notifications.info('New message received', {
  label: 'View',
  onClick: () => goto('/messages')
});
```

---

## 🎨 **UI Patterns**

### **Notification Toast Example**:
```svelte
<script>
  import { notifications } from '$lib/stores/notifications';
  import { fade, fly } from 'svelte/transition';
  
  $: toasts = $notifications.items;
</script>

<div class="toast-container">
  {#each toasts as toast (toast.id)}
    <div
      class="toast toast-{toast.type}"
      transition:fly="{{ y: -20, duration: 300 }}"
    >
      <div class="toast-icon">
        {#if toast.type === 'success'}✓{/if}
        {#if toast.type === 'error'}✗{/if}
        {#if toast.type === 'info'}ℹ{/if}
        {#if toast.type === 'warning'}⚠{/if}
      </div>
      
      <div class="toast-content">
        <p>{toast.message}</p>
        {#if toast.action}
          <button on:click={toast.action.onClick}>
            {toast.action.label}
          </button>
        {/if}
      </div>
      
      <button
        class="toast-close"
        on:click={() => notifications.dismiss(toast.id)}
      >
        ×
      </button>
    </div>
  {/each}
</div>

<style>
  .toast-container {
    position: fixed;
    top: 1rem;
    right: 1rem;
    z-index: 9999;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  
  .toast {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1rem;
    background: white;
    border-radius: 0.5rem;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    min-width: 300px;
    max-width: 500px;
  }
  
  .toast-success { border-left: 4px solid #10b981; }
  .toast-error { border-left: 4px solid #ef4444; }
  .toast-info { border-left: 4px solid #3b82f6; }
  .toast-warning { border-left: 4px solid #f59e0b; }
</style>
```

---

## 🔄 **Data Flow Examples**

### **Update Email Preferences**:
```
1. User toggles preference
2. Store.update() called
3. API: PUT /api/v1/users/email-preferences
4. Backend saves to database
5. Success response
6. Store updates state
7. Toast notification shown
```

### **Show In-App Notification**:
```
1. Event triggers notification
2. notifications.success(message)
3. Notification added to store
4. Toast appears with animation
5. Auto-dismiss after 5 seconds
6. User can manually dismiss
```

---

## 🔒 **Security & Privacy**

### **User Control**:
- ✅ Easy unsubscribe
- ✅ Granular preferences
- ✅ Clear opt-in/opt-out
- ✅ Frequency control

### **Data Protection**:
- ✅ Preferences encrypted in transit
- ✅ No email sharing
- ✅ GDPR compliance
- ✅ CAN-SPAM compliance

---

## 📊 **Analytics Integration** (Future)

**Track Engagement**:
```typescript
import { analytics } from '$lib/integrations/analytics';

// Track preference change
analytics.track('email_preferences_updated', {
  categories_enabled: ['security', 'account'],
  frequency: 'daily'
});

// Track unsubscribe
analytics.track('email_unsubscribed', {
  reason: reason || 'not_specified'
});
```

---

## 🔗 **Related Documentation**

### **Backend Portion**:
- [`_braids/communication/backend/BRAID.md`](../../_braids/communication/backend/BRAID.md)

### **API Contracts** (Future):
- Email preferences API
- Notification API

### **Strands**:
- Email delivery flow
- Template management
- Preference management

---

## 📝 **Known Issues**

### **To Implement**:
1. In-app notification system
2. Notification center
3. Email preference API endpoint
4. Unsubscribe functionality
5. Admin email dashboard

---

## 🚀 **Quick Links**

**Actual Files**:
- Settings Page: `frontend/src/routes/account/settings/+page.svelte`
- Admin Email: `frontend/src/routes/admin/streaming/email/+page.svelte`
- Notification Store: `frontend/src/lib/stores/notifications.ts` (future)

---

**Last Updated**: October 14, 2025  
**Status**: Core structure defined, implementation in progress  
**Technology**: Svelte 5 + TypeScript  
**Backend Counterpart**: `_braids/communication/backend/`

---

**Navigate**:  
[🏠 Master Index](../../../BRAIDS_INDEX.md) | [⬅️ Backend Braid](../../_braids/communication/backend/BRAID.md) | [🔗 User Mgmt Braid](../user-management/BRAID.md)

