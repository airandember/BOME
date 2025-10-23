# Braid: communication

**Architecture:** Full-Stack Braid (Frontend to Backend)
**Last Updated:** 2025-10-17

---

## Backend Architecture

**Email notifications, messaging, and communication management**

---

## ðŸ”— **Cross-Repository Braid**

> **âš ï¸ IMPORTANT**: This is the **backend portion** of the Communication Braid.  
> **Frontend portion**: See `_frontend/braids/communication/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## ðŸ“‹ **Backend Overview**

**Purpose**: Server-side email and notification system  
**Technology**: Go, Resend (Email Service), PostgreSQL  
**Complexity**: Medium (Email Integration, Templates, Queuing)  
**Dependencies**: Auth Braid (user identity), User Mgmt Braid (preferences)

---

## ðŸŽ¯ **Key Features**

### **1. Email System**:
- Transactional emails (verification, password reset)
- Email templates with variables
- HTML email rendering
- Development mode (console logging)
- Production mode (Resend/SMTP)

### **2. Email Templates**:
- Verification email
- Password reset email
- Welcome email
- Subscription notifications
- Payment receipts
- Account updates

### **3. User Preferences**:
- Email notification toggles
- Frequency settings (immediate, daily digest)
- Unsubscribe management
- Communication channels

### **4. Delivery Tracking**:
- Email delivery status
- Bounce handling
- Open/click tracking (optional)
- Delivery analytics

---

## ðŸ“§ **Email Service Integration**

### **Current Implementation**: `backend/internal/services/email.go`

**Service**: Resend (primary), with fallback options

**Configuration**:
```go
type EmailConfig struct {
    Provider    string // "resend", "smtp", "mock"
    APIKey      string // Resend API key
    FromEmail   string // "noreply@yourdomain.com"
    FromName    string // "BOME Platform"
    Environment string // "development", "production"
}
```

**Development Mode**:
```go
if config.Environment == "development" {
    // Log email to console instead of sending
    log.Printf("MOCK EMAIL to %s: %s", email, subject)
    log.Printf("Body: %s", body)
    return nil
}
```

**Production Mode**:
```go
// Send via Resend API
client := resend.NewClient(config.APIKey)
params := &resend.SendEmailRequest{
    From:    config.FromEmail,
    To:      []string{email},
    Subject: subject,
    Html:    htmlBody,
}
_, err := client.Emails.Send(params)
```

---

## ðŸ“„ **Email Templates** (In Code)

### **File**: `backend/internal/services/email.go`

**Embedded HTML Templates** (lines 118-390):

**1. Verification Email**:
```go
func SendVerificationEmail(email, firstName, verificationLink string) error {
    subject := "Verify your BOME email address"
    
    htmlBody := fmt.Sprintf(`
        <!DOCTYPE html>
        <html>
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
        </head>
        <body style="font-family: Arial, sans-serif; ...">
            <div style="max-width: 600px; margin: 0 auto; ...">
                <h1>Welcome to BOME!</h1>
                <p>Hi %s,</p>
                <p>Please verify your email address...</p>
                <a href="%s" style="...">Verify Email</a>
                <p>This link expires in 3 hours.</p>
            </div>
        </body>
        </html>
    `, firstName, verificationLink)
    
    return sendEmail(email, subject, htmlBody)
}
```

**2. Password Reset Email**:
```go
func SendPasswordResetEmail(email, firstName, resetLink string) error {
    subject := "Reset your BOME password"
    // Similar HTML template with reset link
    // Expires in 1 hour
}
```

**3. Welcome Email**:
```go
func SendWelcomeEmail(email, firstName string) error {
    subject := "Welcome to BOME!"
    // Onboarding information
    // Getting started tips
}
```

**4. Subscription Notification**:
```go
func SendSubscriptionEmail(email, firstName, plan string) error {
    subject := "Your BOME subscription is active"
    // Plan details
    // Billing information
    // Next billing date
}
```

---

## ðŸ—„ï¸ **Database Schema** (Optional - Not Yet Implemented)

### **Future Enhancement**: `email_templates` table

```sql
CREATE TABLE email_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    subject VARCHAR(255) NOT NULL,
    html_body TEXT NOT NULL,
    text_body TEXT,
    variables JSONB, -- Template variables
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### **Future Enhancement**: `email_logs` table

```sql
CREATE TABLE email_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    template_name VARCHAR(100),
    recipient VARCHAR(255) NOT NULL,
    subject VARCHAR(255),
    status VARCHAR(50), -- 'sent', 'delivered', 'bounced', 'failed'
    provider_message_id VARCHAR(255),
    sent_at TIMESTAMP DEFAULT NOW(),
    delivered_at TIMESTAMP,
    opened_at TIMESTAMP,
    clicked_at TIMESTAMP,
    error_message TEXT
);

CREATE INDEX idx_email_logs_user_id ON email_logs(user_id);
CREATE INDEX idx_email_logs_status ON email_logs(status);
CREATE INDEX idx_email_logs_sent_at ON email_logs(sent_at);
```

### **Current Implementation**: User preferences in `users` table

```sql
-- In users table:
email_verified BOOLEAN DEFAULT false
-- Future: email_preferences JSONB column
```

---

## ðŸŒ **API Endpoints** (Future Enhancement)

### **Email Management** (Admin):
```
GET    /api/v1/admin/emails          # List email logs
GET    /api/v1/admin/emails/:id      # Get email details
POST   /api/v1/admin/emails/send     # Send manual email
GET    /api/v1/admin/templates       # List templates
POST   /api/v1/admin/templates       # Create template
PUT    /api/v1/admin/templates/:id   # Update template
```

### **User Preferences**:
```
GET    /api/v1/users/email-preferences     # Get preferences
PUT    /api/v1/users/email-preferences     # Update preferences
POST   /api/v1/users/unsubscribe           # Unsubscribe
```

---

## ðŸ”§ **Email Service Functions**

### **Current Implementation** (`email.go`):

**1. SendVerificationEmail()** - Lines 458-508
- Sends email verification link
- 3-hour expiration
- Embedded HTML template
- Mock mode for development

**2. SendPasswordResetEmail()** - Similar pattern
- Password reset link
- 1-hour expiration
- Security notice

**3. SendEmail()** - Base function
- Handles provider selection
- Mock vs production mode
- Error handling
- Retry logic (optional)

---

## ðŸ“Š **Email Analytics** (Future)

### **Metrics to Track**:
- Emails sent
- Delivery rate
- Open rate
- Click rate
- Bounce rate
- Unsubscribe rate

### **Reports**:
- Daily email volume
- Template performance
- User engagement
- Delivery issues

---

## ðŸ”’ **Security & Compliance**

### **Security Measures**:
- âœ… API keys in environment variables
- âœ… Rate limiting on email sending
- âœ… SPF/DKIM/DMARC records
- âœ… Unsubscribe links required
- âœ… Personal data encryption

### **Compliance**:
- âœ… **CAN-SPAM**: Unsubscribe links, physical address
- âœ… **GDPR**: Data protection, right to be forgotten
- âœ… **Privacy**: User consent, opt-in/opt-out

---

## âš¡ **Performance Optimization**

### **Current Implementation**:
- Synchronous email sending
- Blocking on email service
- ~500-1000ms per email

### **Future Improvements**:
1. **Message Queue**: Background job processing
2. **Batch Sending**: Multiple recipients
3. **Async Processing**: Non-blocking
4. **Caching**: Template caching
5. **Connection Pooling**: SMTP connections

---

## ðŸŽ¯ **Email Templates Best Practices**

### **HTML Email Design**:
- âœ… Inline CSS (no external stylesheets)
- âœ… Responsive design (mobile-friendly)
- âœ… Plain text fallback
- âœ… Alt text for images
- âœ… Clear call-to-action
- âœ… Unsubscribe link

### **Content Guidelines**:
- Clear, concise subject lines
- Personalization (first name)
- Action-oriented CTAs
- Brand consistency
- Professional tone

---

## ðŸ”— **Integration with Other Braids**

### **Authentication Braid**:
- âœ… Email verification
- âœ… Password reset
- âœ… Welcome emails

### **Subscription Braid**:
- âœ… Payment confirmations
- âœ… Subscription updates
- âœ… Invoice emails

### **User Management Braid**:
- âœ… Profile update notifications
- âœ… Security alerts
- âœ… Account changes

### **Admin Dashboard Braid**:
- âœ… Admin notifications
- âœ… System alerts
- âœ… Report delivery

---

## ðŸ“ **Known Technical Debt**

### **Current Limitations**:
1. **Templates in code**: Should be in database
2. **No email logs**: No delivery tracking
3. **Synchronous sending**: Blocks requests
4. **No retry logic**: Failed emails lost
5. **No queue system**: Not scalable
6. **No analytics**: Can't track performance

### **Planned Improvements**:
1. âœ… Move templates to database
2. âœ… Implement email logging
3. âœ… Add message queue (Redis/RabbitMQ)
4. âœ… Add retry mechanism
5. âœ… Implement delivery tracking
6. âœ… Add analytics dashboard

---

## ðŸ§¬ **Strands (Complete Flows)**

### **1. Email Delivery Strand**:
Complete flow from email send request to delivery

### **2. Template Management Strand**:
Template creation, editing, and rendering

### **3. Notification System Strand**:
User notification preferences and delivery

---

## ðŸš€ **Quick Start**

### **Understanding Communication System** (15 min):
1. Read this BRAID.md (5 min)
2. Check `backend/internal/services/email.go` (5 min)
3. Review email verification strand (5 min)

### **Sending Emails**:
```go
import "backend/internal/services"

// Send verification email
email.SendVerificationEmail(
    user.Email,
    user.FirstName,
    verificationLink,
)

// Send password reset
email.SendPasswordResetEmail(
    user.Email,
    user.FirstName,
    resetLink,
)
```

---

## ðŸ“ **File Locations**

### **Backend Files**:
- Email Service: `backend/internal/services/email.go`
- Email Helpers: `backend/internal/services/email_helpers.go`
- Config: `backend/internal/config/config.go` (email settings)

### **Environment Variables**:
```env
EMAIL_PROVIDER=resend
RESEND_API_KEY=re_xxx
EMAIL_FROM=noreply@yourdomain.com
EMAIL_FROM_NAME=BOME Platform
ENVIRONMENT=development
```

---

**Last Updated**: October 14, 2025  
**Status**: Core implementation complete, enhancements planned  
**Technology**: Go + Resend  
**Frontend Counterpart**: `_frontend/braids/communication/`

---

**Navigate**:  
[ðŸ  Master Index](../../BRAIDS_INDEX.md) | [ðŸŽ¨ Frontend Braid](../../_frontend/braids/communication/BRAID.md) | [ðŸ”— Auth Braid](../authentication/BRAID.md)



---

## Frontend Architecture

**Svelte5 UI for email preferences and notifications**

---

## ðŸ”— **Cross-Repository Braid**

> **âš ï¸ IMPORTANT**: This is the **frontend portion** of the Communication Braid.  
> **Backend portion**: See `_backend/braids/communication/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## ðŸ“‹ **Frontend Overview**

**Purpose**: User interface for notification preferences and email management  
**Technology**: Svelte 5, TypeScript, TailwindCSS  
**Entry Points**: `/account/notifications`, `/admin/streaming/email`  
**State Management**: Svelte stores for notification state

---

## ðŸŽ¯ **Key Features**

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

## ðŸ“„ **Frontend Pages**

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

## ðŸ§© **Frontend Components**

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

## ðŸ—ƒï¸ **Frontend Stores**

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

## ðŸŽ¨ **UI Patterns**

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
        {#if toast.type === 'success'}âœ“{/if}
        {#if toast.type === 'error'}âœ—{/if}
        {#if toast.type === 'info'}â„¹{/if}
        {#if toast.type === 'warning'}âš {/if}
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
        Ã—
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

## ðŸ”„ **Data Flow Examples**

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

## ðŸ”’ **Security & Privacy**

### **User Control**:
- âœ… Easy unsubscribe
- âœ… Granular preferences
- âœ… Clear opt-in/opt-out
- âœ… Frequency control

### **Data Protection**:
- âœ… Preferences encrypted in transit
- âœ… No email sharing
- âœ… GDPR compliance
- âœ… CAN-SPAM compliance

---

## ðŸ“Š **Analytics Integration** (Future)

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

## ðŸ”— **Related Documentation**

### **Backend Portion**:
- [`_backend/braids/communication/BRAID.md`](../../_backend/braids/communication/BRAID.md)

### **API Contracts** (Future):
- Email preferences API
- Notification API

### **Strands**:
- Email delivery flow
- Template management
- Preference management

---

## ðŸ“ **Known Issues**

### **To Implement**:
1. In-app notification system
2. Notification center
3. Email preference API endpoint
4. Unsubscribe functionality
5. Admin email dashboard

---

## ðŸš€ **Quick Links**

**Actual Files**:
- Settings Page: `frontend/src/routes/account/settings/+page.svelte`
- Admin Email: `frontend/src/routes/admin/streaming/email/+page.svelte`
- Notification Store: `frontend/src/lib/stores/notifications.ts` (future)

---

**Last Updated**: October 14, 2025  
**Status**: Core structure defined, implementation in progress  
**Technology**: Svelte 5 + TypeScript  
**Backend Counterpart**: `_backend/braids/communication/`

---

**Navigate**:  
[ðŸ  Master Index](../../../BRAIDS_INDEX.md) | [â¬…ï¸ Backend Braid](../../_backend/braids/communication/BRAID.md) | [ðŸ”— User Mgmt Braid](../user-management/BRAID.md)



---

## Integration Notes

- Frontend: `_braids/communication/frontend/`
- Backend: `_braids/communication/backend/`

This braid represents a complete vertical slice of functionality.

