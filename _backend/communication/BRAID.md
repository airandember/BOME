# 🧬 Communication Braid - Backend
**Email notifications, messaging, and communication management**

---

## 🔗 **Cross-Repository Braid**

> **⚠️ IMPORTANT**: This is the **backend portion** of the Communication Braid.  
> **Frontend portion**: See `_frontend/braids/communication/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## 📋 **Backend Overview**

**Purpose**: Server-side email and notification system  
**Technology**: Go, Resend (Email Service), PostgreSQL  
**Complexity**: Medium (Email Integration, Templates, Queuing)  
**Dependencies**: Auth Braid (user identity), User Mgmt Braid (preferences)

---

## 🎯 **Key Features**

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

## 📧 **Email Service Integration**

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

## 📄 **Email Templates** (In Code)

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

## 🗄️ **Database Schema** (Optional - Not Yet Implemented)

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

## 🌐 **API Endpoints** (Future Enhancement)

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

## 🔧 **Email Service Functions**

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

## 📊 **Email Analytics** (Future)

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

## 🔒 **Security & Compliance**

### **Security Measures**:
- ✅ API keys in environment variables
- ✅ Rate limiting on email sending
- ✅ SPF/DKIM/DMARC records
- ✅ Unsubscribe links required
- ✅ Personal data encryption

### **Compliance**:
- ✅ **CAN-SPAM**: Unsubscribe links, physical address
- ✅ **GDPR**: Data protection, right to be forgotten
- ✅ **Privacy**: User consent, opt-in/opt-out

---

## ⚡ **Performance Optimization**

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

## 🎯 **Email Templates Best Practices**

### **HTML Email Design**:
- ✅ Inline CSS (no external stylesheets)
- ✅ Responsive design (mobile-friendly)
- ✅ Plain text fallback
- ✅ Alt text for images
- ✅ Clear call-to-action
- ✅ Unsubscribe link

### **Content Guidelines**:
- Clear, concise subject lines
- Personalization (first name)
- Action-oriented CTAs
- Brand consistency
- Professional tone

---

## 🔗 **Integration with Other Braids**

### **Authentication Braid**:
- ✅ Email verification
- ✅ Password reset
- ✅ Welcome emails

### **Subscription Braid**:
- ✅ Payment confirmations
- ✅ Subscription updates
- ✅ Invoice emails

### **User Management Braid**:
- ✅ Profile update notifications
- ✅ Security alerts
- ✅ Account changes

### **Admin Dashboard Braid**:
- ✅ Admin notifications
- ✅ System alerts
- ✅ Report delivery

---

## 📝 **Known Technical Debt**

### **Current Limitations**:
1. **Templates in code**: Should be in database
2. **No email logs**: No delivery tracking
3. **Synchronous sending**: Blocks requests
4. **No retry logic**: Failed emails lost
5. **No queue system**: Not scalable
6. **No analytics**: Can't track performance

### **Planned Improvements**:
1. ✅ Move templates to database
2. ✅ Implement email logging
3. ✅ Add message queue (Redis/RabbitMQ)
4. ✅ Add retry mechanism
5. ✅ Implement delivery tracking
6. ✅ Add analytics dashboard

---

## 🧬 **Strands (Complete Flows)**

### **1. Email Delivery Strand**:
Complete flow from email send request to delivery

### **2. Template Management Strand**:
Template creation, editing, and rendering

### **3. Notification System Strand**:
User notification preferences and delivery

---

## 🚀 **Quick Start**

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

## 📁 **File Locations**

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
[🏠 Master Index](../../BRAIDS_INDEX.md) | [🎨 Frontend Braid](../../_frontend/braids/communication/BRAID.md) | [🔗 Auth Braid](../authentication/BRAID.md)

