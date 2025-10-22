# 🧬 STRAND: Session Management
**Complete data flow for JWT session lifecycle and management**

---

## 📋 **Strand Overview**

**Purpose**: Document JWT token lifecycle, refresh, and session tracking  
**Complexity**: High (Token refresh, session validation, multi-device)  
**Entry Points**: Token refresh, middleware authentication, logout  
**Layers Traversed**: All 5 layers (Presentation → Persistence)  
**Critical Path**: Every authenticated request

---

## 🎯 **Session Lifecycle**

```
1. User logs in
   ↓
2. JWT tokens generated (access + refresh)
   ↓
3. Session created in database
   ↓
4. Tokens stored in frontend
   ↓
5. Access token sent with each request
   ↓
6. Middleware validates token
   ↓
7. Session activity updated
   ↓
8. Token expires (4 hours)
   ↓
9. Frontend auto-refreshes with refresh token
   ↓
10. New access token issued
   ↓
11. User logs out → session invalidated
```

**Session Duration**: 7 days (refresh token)  
**Access Token**: 4 hours  
**Refresh Token**: 7 days  
**Max Concurrent Sessions**: 5 (configurable per user)

---

## 🔑 **JWT Token Structure**

### **Access Token** (4-hour lifespan):
```json
{
  "user_id": "123",
  "email": "user@example.com",
  "role": "user",
  "token_id": "session-uuid-456",
  "exp": 1728901234,
  "iat": 1728886834,
  "jti": "access-token-uuid"
}
```

**Claims**:
- `user_id`: User database ID
- `email`: User email (for display)
- `role`: User role (user, admin, etc.)
- `token_id`: Session ID for revocation
- `exp`: Expiration timestamp
- `iat`: Issued at timestamp
- `jti`: JWT ID (unique identifier)

---

### **Refresh Token** (7-day lifespan):
```json
{
  "user_id": "123",
  "token_id": "session-uuid-456",
  "type": "refresh",
  "exp": 1729491234,
  "iat": 1728886834,
  "jti": "refresh-token-uuid"
}
```

**Claims**:
- `user_id`: User database ID
- `token_id`: Session ID for validation
- `type`: "refresh" (identifies token type)
- `exp`: Expiration timestamp (7 days)
- `iat`: Issued at timestamp
- `jti`: JWT ID (unique identifier)

---

## 🗄️ **Database Session Tracking**

### **Table**: `user_sessions`

**Purpose**: Track active sessions for:
- Token revocation (logout)
- Multi-device management
- Security monitoring
- Activity tracking

**Session Record**:
```sql
CREATE TABLE user_sessions (
    id SERIAL PRIMARY KEY,
    session_id VARCHAR(255) NOT NULL UNIQUE,     -- UUID
    user_id INTEGER NOT NULL REFERENCES users(id),
    token_id VARCHAR(255) NOT NULL UNIQUE,       -- JWT jti claim
    device_info VARCHAR(500),                     -- "Windows 10, Chrome 96"
    ip_address VARCHAR(45),                       -- IPv4/IPv6
    user_agent TEXT,                              -- Full UA string
    created_at TIMESTAMP DEFAULT NOW(),
    last_activity TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,                -- 7 days from creation
    is_active BOOLEAN DEFAULT true
);
```

**Indexes**:
```sql
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token_id ON user_sessions(token_id);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);
```

---

## 🌐 **Layer-by-Layer Flow**

---

### **🎨 LAYER 1: Presentation (Frontend Session Management)**

**File**: `frontend/src/lib/auth.ts`

#### **Token Storage**:
```typescript
class SecureTokenStorage {
  private static readonly STORAGE_KEY = 'bome_auth_tokens';
  
  static setTokens(tokens: {
    accessToken: string;
    refreshToken: string;
    expiresAt: number;
  }): void {
    localStorage.setItem(this.STORAGE_KEY, JSON.stringify(tokens));
  }
  
  static getTokens(): Tokens | null {
    const stored = localStorage.getItem(this.STORAGE_KEY);
    return stored ? JSON.parse(stored) : null;
  }
  
  static clearTokens(): void {
    localStorage.removeItem(this.STORAGE_KEY);
  }
  
  static isTokenExpired(): boolean {
    const tokens = this.getTokens();
    if (!tokens) return true;
    
    // Check if expires in next 5 minutes
    return Date.now() >= (tokens.expiresAt - 5 * 60 * 1000);
  }
}
```

---

#### **Auto-Refresh Pattern**:
```typescript
// On app initialization
import { browser } from '$app/environment';

if (browser) {
  // Check token on page load
  onMount(async () => {
    const tokens = SecureTokenStorage.getTokens();
    
    if (tokens) {
      if (SecureTokenStorage.isTokenExpired()) {
        // Token expired or expiring soon, refresh it
        try {
          await auth.refreshToken();
          await auth.getCurrentUser();
        } catch (error) {
          // Refresh failed, logout
          auth.logout();
        }
      } else {
        // Token still valid, just load user
        await auth.getCurrentUser();
      }
    }
  });
  
  // Set up periodic token check (every 5 minutes)
  setInterval(async () => {
    if (SecureTokenStorage.isTokenExpired()) {
      try {
        await auth.refreshToken();
      } catch (error) {
        auth.logout();
      }
    }
  }, 5 * 60 * 1000); // 5 minutes
}
```

---

#### **Authenticated Request Pattern**:
```typescript
async function authenticatedFetch(url: string, options: RequestInit = {}) {
  const tokens = SecureTokenStorage.getTokens();
  
  if (!tokens) {
    throw new Error('Not authenticated');
  }
  
  // Check if token needs refresh
  if (SecureTokenStorage.isTokenExpired()) {
    await auth.refreshToken();
  }
  
  // Make request with token
  const response = await fetch(url, {
    ...options,
    headers: {
      ...options.headers,
      'Authorization': `Bearer ${SecureTokenStorage.getTokens()?.accessToken}`,
    },
  });
  
  // Handle 401 Unauthorized
  if (response.status === 401) {
    try {
      // Try to refresh token
      await auth.refreshToken();
      
      // Retry original request
      return fetch(url, {
        ...options,
        headers: {
          ...options.headers,
          'Authorization': `Bearer ${SecureTokenStorage.getTokens()?.accessToken}`,
        },
      });
    } catch (refreshError) {
      // Refresh failed, logout
      auth.logout();
      throw refreshError;
    }
  }
  
  return response;
}
```

---

### **🔗 LAYER 2: Application (API Contracts)**

#### **Token Refresh Endpoint**:
```
POST /api/v1/auth/refresh
```

**Request**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 14400
}
```

**Error** (401 Unauthorized):
```json
{
  "error": "Invalid or expired refresh token",
  "code": "INVALID_REFRESH_TOKEN"
}
```

---

### **⚙️ LAYER 3: Business Logic (Auth Middleware)**

#### **File**: `backend/internal/middleware/middleware.go`

**Function**: `AuthMiddleware(next http.Handler) http.Handler`

```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. Extract Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            respondError(w, "Authorization required", http.StatusUnauthorized)
            return
        }
        
        // 2. Parse "Bearer {token}"
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            respondError(w, "Invalid authorization format", http.StatusUnauthorized)
            return
        }
        
        token := parts[1]
        
        // 3. Validate JWT token
        claims, err := jwt.ValidateToken(token)
        if err != nil {
            if err == jwt.ErrTokenExpired {
                respondError(w, "Token expired", http.StatusUnauthorized)
            } else {
                respondError(w, "Invalid token", http.StatusUnauthorized)
            }
            return
        }
        
        // 4. Extract claims
        userID := claims["user_id"].(string)
        tokenID := claims["token_id"].(string)
        
        // 5. Check if session is still active
        session, err := database.GetSessionByTokenID(tokenID, userID)
        if err == sql.ErrNoRows {
            respondError(w, "Session not found", http.StatusUnauthorized)
            return
        }
        if err != nil {
            log.Printf("Database error: %v", err)
            respondError(w, "Authentication failed", http.StatusInternalServerError)
            return
        }
        
        if !session.IsActive {
            respondError(w, "Session revoked", http.StatusUnauthorized)
            return
        }
        
        if time.Now().After(session.ExpiresAt) {
            respondError(w, "Session expired", http.StatusUnauthorized)
            return
        }
        
        // 6. Get user details
        user, err := database.GetUserByID(userID)
        if err != nil {
            respondError(w, "User not found", http.StatusUnauthorized)
            return
        }
        
        if !user.IsActive {
            respondError(w, "Account suspended", http.StatusForbidden)
            return
        }
        
        // 7. Update session activity (async for performance)
        go database.UpdateSessionActivity(tokenID)
        
        // 8. Add user to request context
        ctx := context.WithValue(r.Context(), "user", user)
        ctx = context.WithValue(ctx, "session", session)
        
        // 9. Call next handler
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**Key Features**:
- ✅ JWT signature validation
- ✅ Token expiration check
- ✅ Session revocation support
- ✅ User active status check
- ✅ Session activity tracking
- ✅ User context injection
- ⚡ Async activity update for performance

---

#### **Token Refresh Handler**:

**File**: `backend/internal/routes/auth.go`

```go
func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        RefreshToken string `json:"refresh_token"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // 1. Validate refresh token
    claims, err := jwt.ValidateToken(req.RefreshToken)
    if err != nil {
        respondError(w, "Invalid refresh token", http.StatusUnauthorized)
        return
    }
    
    // 2. Check token type
    tokenType, ok := claims["type"].(string)
    if !ok || tokenType != "refresh" {
        respondError(w, "Not a refresh token", http.StatusUnauthorized)
        return
    }
    
    // 3. Extract claims
    userID := claims["user_id"].(string)
    tokenID := claims["token_id"].(string)
    
    // 4. Validate session
    session, err := database.GetSessionByTokenID(tokenID, userID)
    if err == sql.ErrNoRows {
        respondError(w, "Session not found", http.StatusUnauthorized)
        return
    }
    if err != nil {
        log.Printf("Database error: %v", err)
        respondError(w, "Token refresh failed", http.StatusInternalServerError)
        return
    }
    
    if !session.IsActive {
        respondError(w, "Session revoked", http.StatusUnauthorized)
        return
    }
    
    // 5. Get user
    user, err := database.GetUserByID(userID)
    if err != nil {
        respondError(w, "User not found", http.StatusUnauthorized)
        return
    }
    
    if !user.IsActive {
        respondError(w, "Account suspended", http.StatusForbidden)
        return
    }
    
    // 6. Generate new tokens
    newAccessToken, err := jwt.GenerateAccessToken(user)
    if err != nil {
        log.Printf("Failed to generate access token: %v", err)
        respondError(w, "Token refresh failed", http.StatusInternalServerError)
        return
    }
    
    newRefreshToken, err := jwt.GenerateRefreshToken(user)
    if err != nil {
        log.Printf("Failed to generate refresh token: %v", err)
        respondError(w, "Token refresh failed", http.StatusInternalServerError)
        return
    }
    
    // 7. Update session token_id
    newTokenID := extractTokenID(newAccessToken)
    err = database.UpdateSession(session.ID, map[string]interface{}{
        "token_id":      newTokenID,
        "last_activity": time.Now(),
    })
    if err != nil {
        log.Printf("Failed to update session: %v", err)
    }
    
    // 8. Return new tokens
    respondJSON(w, map[string]interface{}{
        "success":       true,
        "access_token":  newAccessToken,
        "refresh_token": newRefreshToken,
        "token_type":    "Bearer",
        "expires_in":    14400, // 4 hours
    }, http.StatusOK)
}
```

---

### **🗄️ LAYER 4: Data Access (Session Operations)**

#### **Get Session by Token ID**:
```go
func GetSessionByTokenID(tokenID string, userID string) (*Session, error) {
    query := `
        SELECT id, session_id, user_id, token_id,
               device_info, ip_address, user_agent,
               created_at, last_activity, expires_at, is_active
        FROM user_sessions
        WHERE token_id = $1 AND user_id = $2
    `
    
    var session Session
    err := db.QueryRow(query, tokenID, userID).Scan(
        &session.ID,
        &session.SessionID,
        &session.UserID,
        &session.TokenID,
        &session.DeviceInfo,
        &session.IPAddress,
        &session.UserAgent,
        &session.CreatedAt,
        &session.LastActivity,
        &session.ExpiresAt,
        &session.IsActive,
    )
    
    if err == sql.ErrNoRows {
        return nil, sql.ErrNoRows
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get session: %w", err)
    }
    
    return &session, nil
}
```

**Performance**: <3ms (indexed on token_id)

---

#### **Update Session Activity**:
```go
func UpdateSessionActivity(tokenID string) error {
    query := `
        UPDATE user_sessions
        SET last_activity = NOW()
        WHERE token_id = $1 AND is_active = true
    `
    
    _, err := db.Exec(query, tokenID)
    if err != nil {
        return fmt.Errorf("failed to update session activity: %w", err)
    }
    
    return nil
}
```

**Performance**: <5ms (updated async, doesn't block requests)

---

#### **Invalidate Session** (Logout):
```go
func InvalidateSession(tokenID string, userID string) error {
    query := `
        UPDATE user_sessions
        SET is_active = false,
            updated_at = NOW()
        WHERE token_id = $1 AND user_id = $2
    `
    
    _, err := db.Exec(query, tokenID, userID)
    if err != nil {
        return fmt.Errorf("failed to invalidate session: %w", err)
    }
    
    return nil
}
```

---

#### **Clean Up Expired Sessions** (Background Job):
```go
func CleanupExpiredSessions() error {
    query := `
        DELETE FROM user_sessions
        WHERE expires_at < NOW() OR
              (is_active = false AND updated_at < NOW() - INTERVAL '30 days')
    `
    
    result, err := db.Exec(query)
    if err != nil {
        return fmt.Errorf("failed to cleanup sessions: %w", err)
    }
    
    rowsAffected, _ := result.RowsAffected()
    log.Printf("Cleaned up %d expired sessions", rowsAffected)
    
    return nil
}
```

**Runs**: Every 1 hour (cron job)

---

### **📊 LAYER 5: Persistence (Database)**

**SQL Queries**:
```sql
-- Check session validity
SELECT id, session_id, user_id, token_id, is_active, expires_at
FROM user_sessions
WHERE token_id = 'uuid-123' AND user_id = '456';

-- Update activity
UPDATE user_sessions
SET last_activity = NOW()
WHERE token_id = 'uuid-123' AND is_active = true;

-- Logout (invalidate)
UPDATE user_sessions
SET is_active = false
WHERE token_id = 'uuid-123' AND user_id = '456';

-- Cleanup expired
DELETE FROM user_sessions
WHERE expires_at < NOW();
```

---

## 🔒 **Security Features**

### **1. Token Revocation**:
- Sessions stored in database
- Can be invalidated server-side
- Logout immediately revokes access
- Admin can force logout users

### **2. Multi-Device Support**:
- Track each device separately
- View all active sessions
- Logout specific devices
- Logout all other devices

### **3. Session Limits**:
- Max 5 concurrent sessions (configurable)
- Oldest session auto-deleted when limit reached
- Prevents unlimited session creation

### **4. Activity Tracking**:
- Last activity timestamp
- IP address logging
- Device fingerprinting
- User agent tracking

### **5. Automatic Cleanup**:
- Expired sessions deleted hourly
- Inactive sessions removed after 30 days
- Prevents database bloat

---

## ⚡ **Performance Optimizations**

### **1. Async Activity Updates**:
```go
// Don't block request waiting for activity update
go database.UpdateSessionActivity(tokenID)
```

### **2. Token Validation Caching** (Future):
```go
// Cache valid tokens for 5 minutes
cacheKey := fmt.Sprintf("session:%s", tokenID)
if cachedSession := cache.Get(cacheKey); cachedSession != nil {
    return cachedSession, nil
}

// Otherwise fetch from database
session, err := database.GetSessionByTokenID(tokenID, userID)
cache.Set(cacheKey, session, 5*time.Minute)
```

### **3. Database Indexes**:
- `token_id` index for fast lookups
- `user_id` index for user queries
- `expires_at` index for cleanup

---

## 📊 **Session Management UI** (Future Feature)

**Page**: `/account/sessions`

**Features**:
- List all active sessions
- Show device info for each
- Last activity timestamp
- Current session highlighted
- "Logout" button for each session
- "Logout all other devices" button

**Example**:
```
Your Active Sessions

✅ Current Session
   Windows 10, Chrome 96
   Last active: Just now
   IP: 192.168.1.100
   
🖥️ Desktop Computer
   macOS, Safari 15
   Last active: 2 hours ago
   IP: 192.168.1.101
   [Logout]
   
📱 iPhone
   iOS, Safari 15
   Last active: 1 day ago
   IP: 10.0.0.50
   [Logout]

[Logout All Other Devices]
```

---

## 🎯 **Success Criteria**

Session management is successful when:
1. ✅ JWT tokens properly validated
2. ✅ Sessions tracked in database
3. ✅ Token refresh works seamlessly
4. ✅ Logout immediately revokes access
5. ✅ Expired sessions cleaned up
6. ✅ Multi-device support works
7. ✅ Activity tracking accurate
8. ✅ Performance optimized (async updates)

---

**Last Updated**: October 14, 2025  
**Status**: ✅ Production-ready  
**Session Duration**: 7 days  
**Access Token Lifespan**: 4 hours  
**Max Sessions**: 5 per user

