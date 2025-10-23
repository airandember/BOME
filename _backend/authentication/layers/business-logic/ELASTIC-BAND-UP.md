# 🔗 ELASTIC BAND: Business Logic → Application
**Interface Contract Between Go Backend and Frontend API**

---

## 📍 **Connection Points**

**From**: Go Backend Services & HTTP Handlers (Layer 3 - Business Logic)  
**To**: HTTP API Layer consumed by Frontend (Layer 2 - Application)  
**Purpose**: Define REST API contracts, request/response formats, and error handling

---

## 🎯 **API Endpoints**

### **POST /api/v1/auth/register**

**Purpose**: Register new user with email verification flow

**Request**:
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Response - Success** (201 Created):
```json
{
  "success": true,
  "message": "Verification email sent. Please check your inbox.",
  "user_id": "123",
  "email": "user@example.com"
}
```

**Response - Errors**:
```json
// 409 Conflict - Email exists
{
  "error": "Email already exists",
  "code": "DUPLICATE_EMAIL"
}

// 400 Bad Request - Invalid input
{
  "error": "Invalid email format",
  "code": "INVALID_EMAIL",
  "details": {
    "field": "email",
    "value": "not-an-email"
  }
}

// 500 Internal Server Error
{
  "error": "An error occurred. Please try again later.",
  "code": "INTERNAL_ERROR"
}
```

**Handler**: `auth.go:RegisterHandler()`  
**Auth Required**: ❌ No  
**Rate Limit**: 5 requests/minute per IP

---

### **POST /api/v1/auth/login**

**Purpose**: Authenticate user and issue JWT tokens

**Request**:
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Response - Success** (200 OK):
```json
{
  "success": true,
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 14400,
  "user": {
    "id": "123",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "email_verified": true,
    "role": "user",
    "profile_picture_url": null
  }
}
```

**Response - Errors**:
```json
// 401 Unauthorized - Invalid credentials
{
  "error": "Invalid email or password",
  "code": "INVALID_CREDENTIALS"
}

// 403 Forbidden - Email not verified
{
  "error": "Please verify your email before logging in",
  "code": "EMAIL_NOT_VERIFIED",
  "user_id": "123",
  "email": "user@example.com"
}

// 403 Forbidden - Account suspended
{
  "error": "Your account has been suspended",
  "code": "ACCOUNT_SUSPENDED"
}
```

**Handler**: `auth.go:LoginHandler()`  
**Auth Required**: ❌ No  
**Rate Limit**: 10 requests/minute per IP

---

### **POST /api/v1/auth/refresh**

**Purpose**: Refresh access token using refresh token

**Request**:
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response - Success** (200 OK):
```json
{
  "success": true,
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 14400
}
```

**Response - Errors**:
```json
// 401 Unauthorized - Invalid or expired token
{
  "error": "Invalid or expired refresh token",
  "code": "INVALID_REFRESH_TOKEN"
}

// 401 Unauthorized - Session revoked
{
  "error": "Session has been revoked",
  "code": "SESSION_REVOKED"
}
```

**Handler**: `auth.go:RefreshTokenHandler()`  
**Auth Required**: ❌ No (uses refresh token)  
**Rate Limit**: 20 requests/minute per user

---

### **POST /api/v1/auth/logout**

**Purpose**: Invalidate current session

**Request**:
```http
POST /api/v1/auth/logout
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response - Success** (200 OK):
```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

**Handler**: `auth.go:LogoutHandler()`  
**Auth Required**: ✅ Yes  
**Rate Limit**: None

---

### **GET /api/v1/auth/verify-email-link**

**Purpose**: Verify email from link in email

**Request**:
```http
GET /api/v1/auth/verify-email-link?token=abc123&user_id=123
```

**Response - Success** (302 Redirect):
```
Location: /auth/setup-password?token=xyz&user_id=123
```

**Response - Errors** (302 Redirect):
```
Location: /auth/verify-email?error=expired&email=user@example.com
Location: /auth/verify-email?error=invalid
```

**Handler**: `auth.go:VerifyEmailLinkHandler()`  
**Auth Required**: ❌ No (uses token)  
**Rate Limit**: 10 requests/minute per IP

---

### **POST /api/v1/auth/setup-password**

**Purpose**: Set password after email verification

**Request**:
```http
POST /api/v1/auth/setup-password
Content-Type: application/json

{
  "token": "password-setup-token",
  "user_id": "123",
  "password": "SecurePassword123!",
  "confirm_password": "SecurePassword123!"
}
```

**Response - Success** (200 OK):
```json
{
  "success": true,
  "message": "Password set successfully",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 14400,
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

**Response - Errors**:
```json
// 400 Bad Request - Passwords don't match
{
  "error": "Passwords do not match",
  "code": "PASSWORD_MISMATCH"
}

// 400 Bad Request - Weak password
{
  "error": "Password must be at least 8 characters",
  "code": "WEAK_PASSWORD"
}

// 401 Unauthorized - Invalid or expired token
{
  "error": "Password setup token has expired",
  "code": "TOKEN_EXPIRED"
}
```

**Handler**: `auth.go:SetupPasswordHandler()`  
**Auth Required**: ❌ No (uses password setup token)  
**Rate Limit**: 5 requests/minute per IP

---

### **POST /api/v1/auth/resend-verification**

**Purpose**: Resend email verification link

**Request**:
```http
POST /api/v1/auth/resend-verification
Content-Type: application/json

{
  "email": "user@example.com",
  "user_id": "123"
}
```

**Response - Success** (200 OK):
```json
{
  "success": true,
  "message": "Verification email sent"
}
```

**Handler**: `auth.go:ResendVerificationHandler()`  
**Auth Required**: ❌ No  
**Rate Limit**: 3 requests/5 minutes per email

---

### **POST /api/v1/auth/forgot-password**

**Purpose**: Request password reset link

**Request**:
```http
POST /api/v1/auth/forgot-password
Content-Type: application/json

{
  "email": "user@example.com"
}
```

**Response - Success** (200 OK):
```json
{
  "success": true,
  "message": "If that email exists, a password reset link has been sent"
}
```

**Note**: Always returns success to prevent email enumeration

**Handler**: `auth.go:ForgotPasswordHandler()`  
**Auth Required**: ❌ No  
**Rate Limit**: 3 requests/hour per IP

---

### **POST /api/v1/auth/reset-password**

**Purpose**: Reset password with token from email

**Request**:
```http
POST /api/v1/auth/reset-password
Content-Type: application/json

{
  "token": "reset-token",
  "password": "NewSecurePassword123!",
  "confirm_password": "NewSecurePassword123!"
}
```

**Response - Success** (200 OK):
```json
{
  "success": true,
  "message": "Password reset successfully. Please log in."
}
```

**Handler**: `auth.go:ResetPasswordHandler()`  
**Auth Required**: ❌ No (uses reset token)  
**Rate Limit**: 5 requests/hour per token

---

### **GET /api/v1/auth/me**

**Purpose**: Get current authenticated user

**Request**:
```http
GET /api/v1/auth/me
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response - Success** (200 OK):
```json
{
  "id": "123",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "email_verified": true,
  "role": "user",
  "is_active": true,
  "profile_picture_url": "https://...",
  "created_at": "2025-01-15T10:30:00Z",
  "last_login": "2025-10-14T09:00:00Z"
}
```

**Handler**: `auth.go:GetCurrentUserHandler()`  
**Auth Required**: ✅ Yes  
**Rate Limit**: 100 requests/minute per user

---

## 🌐 **OAuth2 Endpoints**

### **GET /api/v1/auth/oauth2/{provider}/login**

**Purpose**: Initiate OAuth2 login flow

**Request**:
```http
GET /api/v1/auth/oauth2/google/login?return_url=/dashboard
```

**Response** (302 Redirect):
```
Location: https://accounts.google.com/o/oauth2/auth?client_id=...&state=...
```

**Handler**: `oauth2_routes.go:InitiateOAuth2Login()`  
**Auth Required**: ❌ No

---

### **GET /api/v1/auth/oauth2/{provider}/callback**

**Purpose**: Handle OAuth2 provider callback

**Request**:
```http
GET /api/v1/auth/oauth2/google/callback?code=4/...&state=abc123
```

**Response - Success** (302 Redirect):
```
Location: /dashboard?auth_success=true
Set-Cookie: access_token=...; HttpOnly; Secure
Set-Cookie: refresh_token=...; HttpOnly; Secure
```

**Response - Error** (302 Redirect):
```
Location: /login?error=oauth_failed
```

**Handler**: `oauth2_routes.go:OAuth2CallbackHandler()`  
**Auth Required**: ❌ No (OAuth2 code exchange)

---

## 🔒 **Authentication & Authorization**

### **JWT Token Structure**:
```json
// Access Token (4 hours)
{
  "user_id": "123",
  "email": "user@example.com",
  "role": "user",
  "token_id": "uuid-token-id",
  "exp": 1728901234,
  "iat": 1728886834,
  "jti": "uuid-jti"
}

// Refresh Token (7 days)
{
  "user_id": "123",
  "token_id": "uuid-token-id",
  "type": "refresh",
  "exp": 1729491234,
  "iat": 1728886834,
  "jti": "uuid-jti-refresh"
}
```

### **Authorization Header**:
```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### **Token Storage** (Frontend):
```typescript
// Stored in SecureTokenStorage (localStorage with encryption)
{
  accessToken: "jwt...",
  refreshToken: "jwt...",
  expiresAt: 1728901234,
  user: {...}
}
```

---

## ⚠️ **Error Response Format**

### **Standard Error Structure**:
```json
{
  "error": "Human-readable error message",
  "code": "MACHINE_READABLE_CODE",
  "details": {
    // Optional additional context
  }
}
```

### **HTTP Status Codes**:
| Code | Meaning | When Used |
|------|---------|-----------|
| 200 | OK | Successful operation |
| 201 | Created | Resource created (registration) |
| 400 | Bad Request | Invalid input, validation failed |
| 401 | Unauthorized | Invalid credentials, expired token |
| 403 | Forbidden | Email not verified, account suspended |
| 404 | Not Found | User not found (rare, usually 401) |
| 409 | Conflict | Duplicate email, resource exists |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server-side error |

### **Error Codes**:
```typescript
// Authentication Errors
INVALID_CREDENTIALS      // Wrong email/password
EMAIL_NOT_VERIFIED       // Login before verification
ACCOUNT_SUSPENDED        // Account disabled
INVALID_TOKEN            // JWT token invalid/expired
SESSION_REVOKED          // Session invalidated

// Registration Errors
DUPLICATE_EMAIL          // Email already registered
INVALID_EMAIL            // Email format invalid
WEAK_PASSWORD            // Password doesn't meet requirements
PASSWORD_MISMATCH        // Passwords don't match

// Token Errors
TOKEN_EXPIRED            // Verification/reset token expired
INVALID_REFRESH_TOKEN    // Refresh token invalid
REFRESH_TOKEN_EXPIRED    // Refresh token expired

// Rate Limiting
RATE_LIMIT_EXCEEDED      // Too many requests

// Server Errors
INTERNAL_ERROR           // Generic server error
DATABASE_ERROR           // Database connection failed
EMAIL_SEND_FAILED        // Email service failed
```

---

## ⏱️ **Response Time SLAs**

| Endpoint | Expected Time | P95 | P99 |
|----------|---------------|-----|-----|
| POST /login | 100-200ms | 300ms | 500ms |
| POST /register | 200-400ms | 600ms | 1s |
| POST /refresh | 50-100ms | 150ms | 300ms |
| POST /logout | 50-100ms | 150ms | 200ms |
| GET /me | 20-50ms | 100ms | 150ms |
| OAuth2 flows | 500-1000ms | 1.5s | 3s |

---

## 🔄 **State Management**

### **Frontend Auth State**:
```typescript
interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  loading: boolean;
  error: string | null;
  accessToken: string | null;
  refreshToken: string | null;
}
```

### **State Transitions**:
```
LOGGED_OUT → [login] → AUTHENTICATED
AUTHENTICATED → [logout] → LOGGED_OUT
AUTHENTICATED → [token expired] → REFRESHING → AUTHENTICATED
REFRESHING → [refresh failed] → LOGGED_OUT
```

---

## 🔐 **Security Headers**

### **Required Response Headers**:
```http
Content-Type: application/json
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
```

### **CORS Configuration**:
```go
Access-Control-Allow-Origin: https://yourdomain.com
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Allow-Credentials: true
Access-Control-Max-Age: 86400
```

---

## 📊 **Rate Limiting**

### **Rate Limit Headers**:
```http
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 7
X-RateLimit-Reset: 1728886900
```

### **Rate Limit Response** (429):
```json
{
  "error": "Too many requests. Please try again later.",
  "code": "RATE_LIMIT_EXCEEDED",
  "retry_after": 60
}
```

---

## 📝 **Logging & Monitoring**

### **Request Logging**:
```
[INFO] 2025-10-14 09:00:00 | POST /api/v1/auth/login | 200 | 150ms | ip=192.168.1.100
[WARN] 2025-10-14 09:01:00 | POST /api/v1/auth/login | 401 | 120ms | ip=192.168.1.100 | error=invalid_credentials
[ERROR] 2025-10-14 09:02:00 | POST /api/v1/auth/register | 500 | 50ms | error=database_connection_failed
```

### **Security Event Logging**:
```
[SECURITY] Failed login attempt: user_id=123, ip=192.168.1.100, attempts=3
[SECURITY] Account locked: user_id=123, reason=too_many_failed_attempts
[SECURITY] Password reset requested: user_id=123, ip=192.168.1.100
```

---

**Last Updated**: October 14, 2025  
**API Version**: v1  
**Maintained By**: Backend Team

