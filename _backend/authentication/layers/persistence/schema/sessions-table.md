# 📊 Database Schema: `user_sessions` Table
**Session tracking and device management for authentication**

---

## 📍 **Source Files**

**Migration**: `backend/migrations/005_create_user_sessions_table.sql`  
**Database Models**: `backend/internal/database/session.go` (if exists)  
**Service Logic**: `backend/internal/services/jwt.go`  
**Middleware**: `backend/internal/middleware/middleware.go`

---

## 🗄️ **Table Structure**

```sql
CREATE TABLE user_sessions (
    id SERIAL PRIMARY KEY,
    session_id VARCHAR(255) UNIQUE NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_id VARCHAR(255) NOT NULL,
    device_info TEXT,
    ip_address INET,
    user_agent TEXT,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);
```

---

## 🎯 **Purpose**

This table serves multiple critical functions:

1. **Session Management**: Track active user sessions across devices
2. **Security**: Enforce max sessions per user (default: 5)
3. **JWT Token Blacklisting**: Store token IDs to invalidate specific JWTs
4. **Device Tracking**: Monitor which devices user is logged in from
5. **Activity Monitoring**: Track last activity for security auditing
6. **Session Expiration**: Automatic cleanup of expired sessions

---

## 📋 **Column Definitions**

### `id` SERIAL PRIMARY KEY
- **Purpose**: Unique identifier for each session
- **Type**: Auto-incrementing integer
- **Usage**: Internal database references only

### `session_id` VARCHAR(255) UNIQUE NOT NULL
- **Purpose**: Unique session identifier (UUID format)
- **Generated**: By application when session is created
- **Format**: UUID v4 (e.g., `550e8400-e29b-41d4-a716-446655440000`)
- **Uniqueness**: Enforced at database level with UNIQUE constraint
- **Usage**: 
  - Session lookup and validation
  - Session revocation
- **Validation**: Must be valid UUID format

### `user_id` INTEGER NOT NULL
- **Purpose**: Link session to specific user
- **Foreign Key**: `REFERENCES users(id) ON DELETE CASCADE`
- **Cascade Behavior**: Deleting user automatically deletes all their sessions
- **Index**: `idx_user_sessions_user_id` for fast user-based queries
- **Usage**:
  - Find all sessions for a user
  - Enforce max_sessions limit
  - Security auditing per user

### `token_id` VARCHAR(255) NOT NULL
- **Purpose**: JWT token identifier (from JWT `jti` claim)
- **Format**: UUID v4 matching JWT token
- **Security**: Used for JWT blacklisting/revocation
- **Not Unique**: Same user can have multiple tokens
- **Constraint**: `UNIQUE (user_id, token_id)` prevents duplicate tokens for same user
- **Usage**:
  - Validate JWT hasn't been revoked
  - Invalidate specific JWT when logging out
  - Track which JWT belongs to which session

### `device_info` TEXT
- **Purpose**: Store device fingerprint information (optional)
- **Format**: Plain text description or JSON
- **Content Example**: 
  ```
  Windows 10, Chrome 96.0
  ```
- **Generated**: By frontend device fingerprinting
- **Usage**:
  - Display "logged in devices" to user
  - Security monitoring (detect unusual devices)
  - User-friendly session management UI

### `ip_address` INET
- **Purpose**: IP address where session was created
- **Type**: PostgreSQL INET type (supports IPv4 and IPv6)
- **Collected**: From HTTP request headers
- **Privacy**: Consider data retention policies
- **Usage**:
  - Security auditing
  - Geo-location based security alerts
  - Fraud detection
- **Example**: `192.168.1.100`, `2001:0db8:85a3::8a2e:0370:7334`

### `user_agent` TEXT
- **Purpose**: Browser/app user agent string
- **Collected**: From HTTP `User-Agent` header
- **Length**: Can be very long (500+ characters)
- **Usage**:
  - Device type detection
  - Browser compatibility tracking
  - Security analysis
- **Example**: 
  ```
  Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36...
  ```

### `last_activity` TIMESTAMP
- **Purpose**: Most recent activity timestamp for this session
- **Default**: CURRENT_TIMESTAMP (set on creation)
- **Updated**: Should be updated on each authenticated request
- **Usage**:
  - Idle session detection
  - Activity-based session expiration
  - Security monitoring (detect dormant accounts)
- **Index**: `idx_user_sessions_activity` for cleanup queries

### `is_active` BOOLEAN
- **Purpose**: Whether session is currently active
- **Default**: TRUE (active)
- **Set to FALSE**: When user logs out or session is revoked
- **Usage**:
  - Filter active sessions only
  - Keep history of past sessions
  - Distinguish between expired and manually logged out
- **Index**: `idx_user_sessions_active` for fast active session queries

### `created_at` TIMESTAMP
- **Purpose**: When session was created
- **Default**: CURRENT_TIMESTAMP
- **Immutable**: Never updated after creation
- **Usage**:
  - Session age tracking
  - Audit trails
  - Analytics (login patterns)

### `expires_at` TIMESTAMP NOT NULL
- **Purpose**: When session should expire
- **Required**: NOT NULL (every session must have expiration)
- **Calculation**: `NOW() + 7 days` (or configured TTL)
- **Enforcement**: Application checks this before validating JWT
- **Index**: `idx_user_sessions_expires` for cleanup jobs
- **Usage**:
  - Automatic session expiration
  - Cleanup expired sessions
  - Refresh token expiration logic

---

## 🔍 **Indexes**

```sql
-- User-based queries (find all sessions for user)
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);

-- Token validation (most frequent query)
CREATE INDEX idx_user_sessions_token_id ON user_sessions(token_id);

-- Active session filtering
CREATE INDEX idx_user_sessions_active ON user_sessions(is_active);

-- Expired session cleanup
CREATE INDEX idx_user_sessions_expires ON user_sessions(expires_at);

-- Activity-based queries
CREATE INDEX idx_user_sessions_activity ON user_sessions(last_activity);
```

### **Query Performance Expectations**:
- Token validation: <3ms ⚡
- Find user sessions: <5ms ⚡
- Cleanup expired sessions: <50ms (batch operation)

---

## 🔒 **Constraints**

```sql
-- Session ID must be unique across all users
ALTER TABLE user_sessions 
ADD CONSTRAINT unique_session_id UNIQUE (session_id);

-- User + Token ID must be unique (no duplicate tokens per user)
ALTER TABLE user_sessions 
ADD CONSTRAINT unique_user_token UNIQUE (user_id, token_id);
```

### **Constraint Violations**:
- **Duplicate session_id**: Application error (UUID collision - extremely rare)
- **Duplicate user_id + token_id**: Token reuse attempt (security issue)

---

## 🔗 **Foreign Key Relationships**

### **References**:
```sql
user_id → users(id) ON DELETE CASCADE
```

### **Cascade Behavior**:
When a user is deleted:
✅ All their sessions are automatically deleted  
✅ No orphaned sessions remain  
⚠️ User cannot be deleted if you need to preserve session history

---

## 📊 **Data Flow**

### **1. Login - Create Session**
```go
// In auth.go LoginHandler
session := Session{
    SessionID:    uuid.New().String(),
    UserID:       user.ID,
    TokenID:      jwtTokenID,  // From JWT "jti" claim
    DeviceInfo:   deviceInfo,   // From request
    IPAddress:    r.RemoteAddr,
    UserAgent:    r.Header.Get("User-Agent"),
    LastActivity: time.Now(),
    IsActive:     true,
    CreatedAt:    time.Now(),
    ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
}

// Check max_sessions limit for user
activeSessions := CountActiveSessions(userID)
if activeSessions >= user.MaxSessions {
    // Delete oldest session or return error
}

// Insert session
db.Exec("INSERT INTO user_sessions (...) VALUES (...)")
```

### **2. Request Authentication - Validate Session**
```go
// In middleware.go AuthMiddleware
token := ExtractJWTFromRequest(r)
claims := ParseJWT(token)

// Check if token is revoked (session inactive)
session, err := db.Query(`
    SELECT is_active, expires_at 
    FROM user_sessions 
    WHERE token_id = $1 AND user_id = $2
`, claims.TokenID, claims.UserID)

if !session.IsActive {
    return errors.New("session revoked")
}

if session.ExpiresAt.Before(time.Now()) {
    return errors.New("session expired")
}

// Update last activity
db.Exec(`
    UPDATE user_sessions 
    SET last_activity = NOW() 
    WHERE token_id = $1
`, claims.TokenID)
```

### **3. Logout - Invalidate Session**
```go
// In auth.go LogoutHandler
db.Exec(`
    UPDATE user_sessions 
    SET is_active = FALSE, last_activity = NOW()
    WHERE token_id = $1 AND user_id = $2
`, tokenID, userID)

// Also update user's last_logout timestamp
db.Exec(`
    UPDATE users 
    SET last_logout = NOW() 
    WHERE id = $1
`, userID)
```

### **4. Cleanup - Remove Expired Sessions**
```go
// Scheduled job (e.g., every 6 hours)
func CleanupExpiredSessions() {
    db.Exec(`
        DELETE FROM user_sessions 
        WHERE expires_at < NOW()
    `)
}
```

---

## 🛡️ **Security Considerations**

### **Session Hijacking Prevention**:
✅ Store IP address and user agent  
✅ Detect changes in device fingerprint  
✅ Expire sessions after inactivity  
✅ Enforce maximum concurrent sessions

### **Token Revocation**:
✅ JWT tokens can be immediately invalidated via `is_active = FALSE`  
✅ Prevents "JWT can't be revoked" problem  
✅ Combines JWT benefits (stateless) with session control (stateful)

### **Data Privacy**:
⚠️ IP addresses are PII - apply retention policies  
⚠️ User agents can fingerprint users - consider privacy implications  
⚠️ Device info may contain sensitive data - sanitize before storage

---

## 🧹 **Maintenance Functions**

### **Cleanup Expired Sessions**
```sql
CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS void AS $$
BEGIN
    DELETE FROM user_sessions WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql;
```

**Usage**: Run via cron job every 6 hours  
**Performance**: Typically fast (<100ms) due to index on `expires_at`

### **Revoke All User Sessions** (Security Event)
```sql
-- Force logout all devices for a user
UPDATE user_sessions 
SET is_active = FALSE 
WHERE user_id = $1 AND is_active = TRUE;
```

**Use Cases**:
- Password change (force re-login)
- Security compromise detected
- Admin action (suspend account)

---

## 📈 **Analytics Queries**

### **Active Sessions by User**
```sql
SELECT 
    u.id,
    u.email,
    COUNT(s.id) as active_sessions
FROM users u
LEFT JOIN user_sessions s ON s.user_id = u.id 
    AND s.is_active = TRUE 
    AND s.expires_at > NOW()
GROUP BY u.id, u.email
ORDER BY active_sessions DESC;
```

### **Session Duration Statistics**
```sql
SELECT 
    AVG(EXTRACT(EPOCH FROM (last_activity - created_at))/3600) as avg_duration_hours,
    MAX(EXTRACT(EPOCH FROM (last_activity - created_at))/3600) as max_duration_hours
FROM user_sessions
WHERE is_active = FALSE;  -- Completed sessions
```

### **Login Frequency by Device**
```sql
SELECT 
    device_info,
    COUNT(*) as login_count
FROM user_sessions
WHERE created_at > NOW() - INTERVAL '30 days'
GROUP BY device_info
ORDER BY login_count DESC
LIMIT 10;
```

---

## ⚠️ **Known Issues / Technical Debt**

### **🟡 Medium Priority**

1. **Last Activity Not Always Updated**
   - Current implementation may not update `last_activity` on every request
   - **Impact**: Inactive session detection may not be accurate
   - **Action**: Consider middleware to update on every authenticated request

2. **No Geographic Tracking**
   - IP addresses stored but not geo-located
   - **Opportunity**: Add country/city columns for better security
   - **Action**: Consider MaxMind GeoIP integration

3. **Device Info Format Inconsistent**
   - Currently free-form TEXT field
   - **Issue**: Difficult to query specific device types
   - **Action**: Consider structured JSONB format

### **🟢 Low Priority**

1. Session history grows indefinitely if using `is_active = FALSE` instead of DELETE
2. No session refresh mechanism documented
3. Could add `logout_reason` column (manual, expired, revoked, etc.)

---

## 🚀 **Performance Optimization**

### **Partitioning Strategy** (For high-traffic systems)
```sql
-- Partition by created_at (monthly partitions)
CREATE TABLE user_sessions_2025_10 PARTITION OF user_sessions
FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');
```

**When to Partition**: 
- > 10M sessions
- Slow cleanup queries
- High write volume

### **Index Optimization**
```sql
-- Composite index for common query pattern
CREATE INDEX idx_sessions_user_active_expires 
ON user_sessions(user_id, is_active, expires_at);
```

---

**Last Updated**: October 14, 2025  
**Migration**: 005_create_user_sessions_table.sql  
**Status**: ✅ Production-ready

