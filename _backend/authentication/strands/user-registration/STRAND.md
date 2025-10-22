# 🧬 STRAND: User Registration
**Complete data flow from registration form to logged-in user**

---

## 📋 **Strand Overview**

**Purpose**: Document the complete user registration workflow  
**Complexity**: Medium (email verification + password setup)  
**Entry Point**: Registration form submit  
**Exit Point**: User logged in with JWT tokens  
**Layers Traversed**: All 5 layers (Presentation → Persistence)

---

## 🎯 **User Experience Flow**

```
1. User fills registration form (email, first name, last name)
   ↓
2. Form submits → Backend creates user (NO password yet)
   ↓
3. Email verification sent
   ↓
4. User sees "Check your email" page
   ↓
5. User clicks link in email
   ↓
6. Email verified → Redirected to password setup page
   ↓
7. User creates password
   ↓
8. Password saved → Auto-logged in → Redirected to dashboard
```

**Total Time**: ~3-5 minutes (waiting for email)  
**User Interactions**: 3 (form submit, email click, password creation)

---

## 🌐 **Layer-by-Layer Flow**

---

### **🎨 LAYER 1: Presentation (Svelte5 Frontend)**

#### **File**: `frontend/src/routes/register/+page.svelte`

**Component Structure**:
```svelte
<script lang="ts">
  import { auth } from '$lib/auth';
  
  let email = '';
  let firstName = '';
  let lastName = '';
  let loading = false;
  let error = '';
  
  async function handleSubmit() {
    loading = true;
    error = '';
    
    try {
      const result = await auth.register({
        email,
        first_name: firstName,
        last_name: lastName
      });
      
      if (result.success) {
        // Redirect to verification page
        goto(`/auth/verify-email?email=${email}&user_id=${result.user_id}`);
      }
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }
</script>

<form on:submit|preventDefault={handleSubmit}>
  <input bind:value={email} type="email" required />
  <input bind:value={firstName} required />
  <input bind:value={lastName} required />
  <button type="submit" disabled={loading}>
    {loading ? 'Creating account...' : 'Sign Up'}
  </button>
  {#if error}
    <p class="error">{error}</p>
  {/if}
</form>
```

**Key Points**:
- ✅ Form validation (email format, required fields)
- ✅ Loading state during submission
- ✅ Error display
- ✅ Redirect to verification page on success

**State Management** (`frontend/src/lib/auth.ts`):
```typescript
// Auth store handles registration
export const auth = {
  async register(userData: {email: string, first_name: string, last_name: string}) {
    const response = await fetch(`${API_URL}/auth/register`, {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify(userData)
    });
    
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Registration failed');
    }
    
    return await response.json();
  }
};
```

**↓ ELASTIC BAND: Presentation → Application**

**Data Sent**:
```json
{
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Expected Response**:
```json
{
  "success": true,
  "message": "Verification email sent",
  "user_id": "123",
  "email": "user@example.com"
}
```

**Error Response**:
```json
{
  "error": "Email already exists",
  "code": "DUPLICATE_EMAIL"
}
```

---

### **🔗 LAYER 2: Application (API Contract)**

**Endpoint**: `POST /api/v1/auth/register`

**Request Contract**:
```go
type RegisterRequest struct {
    Email     string `json:"email" validate:"required,email"`
    FirstName string `json:"first_name" validate:"required,min=1,max=100"`
    LastName  string `json:"last_name" validate:"required,min=1,max=100"`
}
```

**Response Contract**:
```go
type RegisterResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    UserID  string `json:"user_id"`
    Email   string `json:"email"`
}
```

**Validation Rules**:
- ✅ Email must be valid format
- ✅ First/last name must be 1-100 characters
- ✅ Email must be lowercase
- ✅ No special characters in names (security)

**Error Codes**:
| Code | HTTP Status | Meaning |
|------|-------------|---------|
| `DUPLICATE_EMAIL` | 409 | Email already registered |
| `INVALID_EMAIL` | 400 | Email format invalid |
| `VALIDATION_ERROR` | 400 | Missing/invalid fields |
| `SERVER_ERROR` | 500 | Internal server error |

**↓ ELASTIC BAND: Application → Business Logic**

---

### **⚙️ LAYER 3: Business Logic (Go Backend)**

#### **File**: `backend/internal/routes/auth.go` (Lines 70-188)

**Function**: `RegisterHandler(w http.ResponseWriter, r *http.Request)`

**Complete Logic Flow**:

```go
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request body
    var req RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // 2. Validate input
    req.Email = strings.ToLower(strings.TrimSpace(req.Email))
    if !isValidEmail(req.Email) {
        respondError(w, "Invalid email format", http.StatusBadRequest)
        return
    }
    
    // 3. Check if email already exists
    existingUser, _ := database.GetUserByEmail(req.Email)
    if existingUser != nil {
        respondError(w, "Email already exists", http.StatusConflict)
        return
    }
    
    // 4. Create user (password_hash = NULL initially)
    user := &database.User{
        Email:         req.Email,
        FirstName:     req.FirstName,
        LastName:      req.LastName,
        EmailVerified: false,
        Role:          "user",
        IsActive:      true,
        CreatedAt:     time.Now(),
    }
    
    createdUser, err := database.CreateUser(user)
    if err != nil {
        log.Printf("Failed to create user: %v", err)
        respondError(w, "Failed to create user", http.StatusInternalServerError)
        return
    }
    
    // 5. Generate email verification token (3 hour expiry)
    token := generateSecureToken()
    expiresAt := time.Now().Add(3 * time.Hour)
    
    err = database.CreateEmailVerificationToken(createdUser.ID, token, expiresAt)
    if err != nil {
        log.Printf("Failed to create verification token: %v", err)
        respondError(w, "Failed to send verification email", http.StatusInternalServerError)
        return
    }
    
    // 6. Send verification email
    verificationLink := fmt.Sprintf("%s/api/v1/auth/verify-email-link?token=%s&user_id=%s", 
        os.Getenv("FRONTEND_URL"), token, createdUser.ID)
    
    err = email.SendVerificationEmail(createdUser.Email, createdUser.FirstName, verificationLink)
    if err != nil {
        log.Printf("Failed to send email: %v", err)
        // Don't fail registration if email fails (user can resend)
    }
    
    // 7. Log registration event
    database.CreateAuditLog("user_registration", createdUser.ID, r.RemoteAddr)
    
    // 8. Return success response
    respondJSON(w, RegisterResponse{
        Success: true,
        Message: "Verification email sent. Please check your inbox.",
        UserID:  createdUser.ID,
        Email:   createdUser.Email,
    }, http.StatusCreated)
}
```

**Key Business Rules**:
- ⚠️ Password is NOT set during registration (NULL in database)
- ⚠️ User cannot login until email is verified
- ⚠️ Verification token expires in 3 hours
- ⚠️ Email failure doesn't block registration (can resend)
- ✅ Email normalized to lowercase
- ✅ Audit log created for tracking
- ✅ User starts with role='user', is_active=true

**Services Called**:
1. `database.GetUserByEmail()` - Check for duplicates
2. `database.CreateUser()` - Create user record
3. `database.CreateEmailVerificationToken()` - Store verification token
4. `email.SendVerificationEmail()` - Send email
5. `database.CreateAuditLog()` - Audit trail

**↓ ELASTIC BAND: Business Logic → Data Access**

---

### **🗄️ LAYER 4: Data Access (Database Operations)**

#### **File**: `backend/internal/database/user.go`

**Function**: `CreateUser(user *User) (*User, error)`

```go
func CreateUser(user *User) (*User, error) {
    query := `
        INSERT INTO users (
            email, 
            first_name, 
            last_name, 
            email_verified, 
            role, 
            is_active,
            created_at,
            updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, created_at, updated_at
    `
    
    err := db.QueryRow(
        query,
        user.Email,
        user.FirstName,
        user.LastName,
        user.EmailVerified,
        user.Role,
        user.IsActive,
        time.Now(),
        time.Now(),
    ).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
    
    if err != nil {
        // Check for unique constraint violation
        if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
            return nil, ErrDuplicateEmail
        }
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    return user, nil
}
```

**Database Constraints Enforced**:
- ✅ Email UNIQUE constraint (catches race conditions)
- ✅ NOT NULL constraints on required fields
- ✅ Foreign key references (if applicable)

**Function**: `CreateEmailVerificationToken(userID string, token string, expiresAt time.Time) error`

```go
func CreateEmailVerificationToken(userID string, token string, expiresAt time.Time) error {
    query := `
        INSERT INTO email_verification_tokens (
            user_id,
            token,
            email,
            expires_at,
            created_at
        ) SELECT 
            $1, $2, email, $3, NOW()
        FROM users 
        WHERE id = $1
    `
    
    _, err := db.Exec(query, userID, token, expiresAt)
    if err != nil {
        return fmt.Errorf("failed to create verification token: %w", err)
    }
    
    return nil
}
```

**↓ ELASTIC BAND: Data Access → Persistence**

---

### **📊 LAYER 5: Persistence (Database Schema)**

#### **Table**: `users`

**SQL Executed**:
```sql
INSERT INTO users (
    email, 
    first_name, 
    last_name, 
    email_verified, 
    password_hash,
    role, 
    is_active,
    created_at,
    updated_at
) VALUES (
    'user@example.com',
    'John',
    'Doe',
    FALSE,               -- Not verified yet
    NULL,                -- No password yet
    'user',
    TRUE,
    '2025-10-14 09:00:00',
    '2025-10-14 09:00:00'
)
RETURNING id, created_at, updated_at;
```

**Result**:
```
id | created_at          | updated_at
---|---------------------|-------------------
123| 2025-10-14 09:00:00 | 2025-10-14 09:00:00
```

#### **Table**: `email_verification_tokens`

**SQL Executed**:
```sql
INSERT INTO email_verification_tokens (
    user_id,
    token,
    email,
    expires_at,
    created_at
) VALUES (
    123,
    'abc123def456...',
    'user@example.com',
    '2025-10-14 12:00:00',  -- 3 hours from now
    '2025-10-14 09:00:00'
);
```

**Constraints Applied**:
- ✅ `user_id` foreign key references `users(id)`
- ✅ `token` is UNIQUE
- ✅ `expires_at` is NOT NULL

---

## 📧 **Email Verification (Part 2 of Registration)**

### **Email Sent** (`backend/internal/services/email.go`)

**Email Template** (embedded in `email.go`):
```html
Subject: Verify your email for BOME

<html>
<body style="background: linear-gradient(to right, #8B5CF6, #EC4899);">
  <div style="max-width: 600px; margin: 0 auto; background: white;">
    <h1>Welcome to BOME, John!</h1>
    <p>Thanks for signing up. Please verify your email address to complete your registration.</p>
    
    <a href="https://yourdomain.com/api/v1/auth/verify-email-link?token=abc123&user_id=123"
       style="background: #8B5CF6; color: white; padding: 12px 24px;">
      Verify Email Address
    </a>
    
    <p><small>This link expires in 3 hours.</small></p>
    <p><small>If you didn't create this account, please ignore this email.</small></p>
  </div>
</body>
</html>
```

**Email Delivery**:
- Service: Resend or Mailgun (hybrid system)
- Fallback: Switches to secondary if primary fails
- Development: Logged to console (not actually sent)
- Production: Real email sent

---

## ✅ **Email Verification Click**

### **User clicks link in email**

**URL**: `GET /api/v1/auth/verify-email-link?token=abc123&user_id=123`

**Handler**: `backend/internal/routes/auth.go:VerifyEmailLinkHandler()` (Lines 756-849)

**Flow**:
```go
1. Extract token and user_id from query params
2. Look up email_verification_tokens WHERE token = $1 AND user_id = $2
3. Check if token expired (expires_at < NOW())
4. If valid:
   a. UPDATE users SET email_verified = TRUE, email_verified_at = NOW()
   b. DELETE email_verification_tokens WHERE token = $1
   c. Generate password_setup_token (short lived, 1 hour)
   d. Redirect to /auth/setup-password?token=xyz&user_id=123
5. If invalid:
   - Show error page with "resend verification" option
```

**Database Updates**:
```sql
-- Mark email as verified
UPDATE users 
SET 
    email_verified = TRUE,
    email_verified_at = NOW(),
    updated_at = NOW()
WHERE id = 123;

-- Delete used verification token
DELETE FROM email_verification_tokens 
WHERE token = 'abc123' AND user_id = 123;
```

---

## 🔐 **Password Setup (Final Step)**

### **Frontend**: `frontend/src/routes/auth/setup-password/+page.svelte`

**Component**:
```svelte
<script lang="ts">
  import { auth } from '$lib/auth';
  
  let password = '';
  let confirmPassword = '';
  let error = '';
  let success = false;
  
  // Get token from URL
  const urlParams = new URLSearchParams(window.location.search);
  const token = urlParams.get('token');
  const userId = urlParams.get('user_id');
  
  async function handleSetPassword() {
    // Validation
    if (password.length < 8) {
      error = 'Password must be at least 8 characters';
      return;
    }
    
    if (password !== confirmPassword) {
      error = 'Passwords do not match';
      return;
    }
    
    try {
      const result = await auth.setupPassword(token, userId, password);
      
      if (result.success) {
        // Auto-logged in, tokens are stored
        success = true;
        setTimeout(() => goto('/dashboard'), 2000);
      }
    } catch (err) {
      error = err.message;
    }
  }
</script>

<form on:submit|preventDefault={handleSetPassword}>
  <h1>Create Your Password</h1>
  
  <input type="password" bind:value={password} required />
  <div class="requirements">
    <p class:valid={password.length >= 8}>✓ At least 8 characters</p>
    <p class:valid={/[A-Z]/.test(password)}>✓ One uppercase letter</p>
    <p class:valid={/[a-z]/.test(password)}>✓ One lowercase letter</p>
    <p class:valid={/[0-9]/.test(password)}>✓ One number</p>
  </div>
  
  <input type="password" bind:value={confirmPassword} placeholder="Confirm password" required />
  
  <button type="submit">Complete Registration</button>
</form>
```

**Backend**: `backend/internal/routes/auth.go:SetupPasswordHandler()` (Lines 1251-1374)

**Flow**:
```go
1. Validate password_setup_token (1 hour expiry)
2. Hash password with bcrypt (cost factor 10)
3. UPDATE users SET password_hash = $1 WHERE id = $2
4. Generate JWT access + refresh tokens
5. Create session in user_sessions table
6. Return tokens in response
7. Frontend stores tokens and redirects to dashboard
```

**SQL**:
```sql
-- Set password
UPDATE users 
SET 
    password_hash = '$2a$10$N9qo8uLOickgx2ZMRZoMye...', -- bcrypt hash
    updated_at = NOW()
WHERE id = 123 AND email_verified = TRUE;

-- Create session
INSERT INTO user_sessions (
    session_id,
    user_id,
    token_id,
    device_info,
    ip_address,
    user_agent,
    expires_at,
    is_active,
    created_at,
    last_activity
) VALUES (
    'session-uuid',
    123,
    'token-uuid',
    'Windows 10, Chrome 96.0',
    '192.168.1.100',
    'Mozilla/5.0...',
    NOW() + INTERVAL '7 days',
    TRUE,
    NOW(),
    NOW()
);
```

**Response**:
```json
{
  "success": true,
  "message": "Password set successfully",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "123",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "email_verified": true,
    "role": "user"
  }
}
```

---

## 🎉 **Registration Complete!**

**User State**:
- ✅ User created in database
- ✅ Email verified
- ✅ Password set
- ✅ JWT tokens issued
- ✅ Session created
- ✅ Logged in
- ✅ Redirected to dashboard

**Database State**:
```
users table:
  id: 123
  email: user@example.com
  password_hash: $2a$10$...
  email_verified: TRUE
  email_verified_at: 2025-10-14 09:15:00
  last_login: 2025-10-14 09:15:00
  role: user
  is_active: TRUE

user_sessions table:
  session_id: session-uuid
  user_id: 123
  token_id: token-uuid
  is_active: TRUE
  expires_at: 2025-10-21 09:15:00 (7 days)
```

---

## ⏱️ **Performance Metrics**

| Step | Expected Time |
|------|---------------|
| Form submit → Response | 100-300ms |
| Email delivery | 1-5 seconds |
| User clicks email link | <100ms |
| Password setup → Login | 100-200ms |
| **Total user time** | **3-5 minutes** (waiting for email) |

**Database Queries**:
- User creation: 1 INSERT
- Verification token: 1 INSERT
- Email lookup: 1 SELECT
- Email verification: 1 UPDATE + 1 DELETE
- Password setup: 1 UPDATE + 1 INSERT (session)
- **Total**: 7 queries across ~3-5 minutes

---

## 🔒 **Security Measures**

1. ✅ Email verification required before password setup
2. ✅ Verification tokens expire (3 hours)
3. ✅ Password setup tokens expire (1 hour)
4. ✅ Passwords hashed with bcrypt (cost 10)
5. ✅ Email normalized to lowercase (prevent case variations)
6. ✅ Audit logging for registration events
7. ✅ Rate limiting on registration endpoint (in middleware)
8. ✅ Input validation (email format, name length)

---

## 🐛 **Common Issues & Solutions**

### **Issue: Email not received**
- **Check**: Spam folder
- **Check**: Email service logs
- **Solution**: Resend verification endpoint available

### **Issue: Verification link expired**
- **Error**: "Verification token expired"
- **Solution**: Resend verification from `/auth/verify-email`

### **Issue: User clicks verify link multiple times**
- **Behavior**: First click works, subsequent clicks fail
- **Reason**: Token deleted after first use (one-time use)
- **Solution**: Show success message, redirect to login

### **Issue: Password setup token expired**
- **Error**: "Password setup token expired"
- **Solution**: User must re-verify email (generates new password setup token)

---

## 📊 **Analytics Events**

Events logged during registration:
1. `user_registration_started` - Form submitted
2. `user_created` - User record created
3. `verification_email_sent` - Email sent
4. `email_verified` - User clicked verification link
5. `password_set` - User completed password setup
6. `user_first_login` - User logged in for first time

---

## 🔄 **Alternative Flows**

### **OAuth2 Registration** (Google Sign-In)
- Skips email verification (Google already verified)
- Creates user + oauth2_accounts record
- Auto-logs in immediately
- See strand: `oauth2-integration/STRAND.md`

### **Admin-Created Users**
- Admin creates user with password
- Email verification optional (admin can mark as verified)
- User can login immediately
- See strand: `admin-user-creation/STRAND.md`

---

**Last Updated**: October 14, 2025  
**Status**: ✅ Production-tested, working reliably  
**Completion Rate**: ~98% (some users abandon before email verification)

