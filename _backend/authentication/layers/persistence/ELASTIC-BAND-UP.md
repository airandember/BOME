# 🔗 ELASTIC BAND: Persistence → Data Access
**Interface Contract Between Database Schema and Go Models**

---

## 📍 **Connection Points**

**From**: PostgreSQL Database Schema (Layer 5 - Persistence)  
**To**: Go Database Models (Layer 4 - Data Access)  
**Purpose**: Define how database tables map to Go structs and SQL operations

---

## 📊 **Table → Struct Mapping**

### **users Table → User Struct**

**Database Table** (`backend/migrations/*users*.sql`):
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    email_verified BOOLEAN DEFAULT FALSE,
    email_verification_token VARCHAR(255),
    email_verification_token_expires_at TIMESTAMP,
    profile_picture_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP,
    role VARCHAR(50) DEFAULT 'user',
    is_active BOOLEAN DEFAULT TRUE,
    -- OAuth fields
    oauth_provider VARCHAR(50),
    oauth_provider_id VARCHAR(255),
    -- Additional fields...
);
```

**Go Model** (`backend/internal/database/user.go`):
```go
type User struct {
    ID                              string    `json:"id"`
    Email                           string    `json:"email"`
    PasswordHash                    string    `json:"-"` // Never expose in JSON
    FirstName                       string    `json:"first_name"`
    LastName                        string    `json:"last_name"`
    EmailVerified                   bool      `json:"email_verified"`
    EmailVerificationToken          *string   `json:"-"`
    EmailVerificationTokenExpiresAt *time.Time `json:"-"`
    ProfilePictureURL               *string   `json:"profile_picture_url"`
    CreatedAt                       time.Time `json:"created_at"`
    UpdatedAt                       time.Time `json:"updated_at"`
    LastLoginAt                     *time.Time `json:"last_login_at"`
    Role                            string    `json:"role"`
    IsActive                        bool      `json:"is_active"`
    OAuthProvider                   *string   `json:"oauth_provider"`
    OAuthProviderID                 *string   `json:"oauth_provider_id"`
}
```

**Contract Requirements:**
✅ All database fields must have corresponding struct fields  
✅ Sensitive fields (password_hash, tokens) must be excluded from JSON  
✅ NULL database fields must use pointers in Go (*string, *time.Time)  
✅ Field names must match database columns (snake_case → json tags)  
✅ UUID types in database map to string in Go

---

### **user_sessions Table → Session Struct**

**Database Table** (`backend/migrations/*sessions*.sql`):
```sql
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_id VARCHAR(255) UNIQUE NOT NULL,
    device_info JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_activity_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);
```

**Go Model** (Expected in `backend/internal/database/session.go`):
```go
type Session struct {
    ID             string          `json:"id"`
    UserID         string          `json:"user_id"`
    TokenID        string          `json:"-"` // Don't expose
    DeviceInfo     json.RawMessage `json:"device_info"`
    IPAddress      *string         `json:"ip_address"`
    UserAgent      *string         `json:"user_agent"`
    ExpiresAt      time.Time       `json:"expires_at"`
    CreatedAt      time.Time       `json:"created_at"`
    LastActivityAt time.Time       `json:"last_activity_at"`
    IsActive       bool            `json:"is_active"`
}
```

---

### **oauth2_* Tables**

**Tables**:
- `oauth2_states` - OAuth2 state validation
- `oauth2_tokens` - OAuth2 access/refresh tokens
- `oauth2_users` - OAuth2 user linking

**Mappings**: See `backend/internal/database/*oauth*.go`

---

## 🔍 **Query Pattern Contracts**

### **Standard CRUD Operations**

**Create**:
```go
// Input: User struct with required fields
// Output: Complete User struct with generated ID, timestamps
// Errors: Duplicate email, validation failures
func CreateUser(user *User) (*User, error)
```

**Read**:
```go
// Input: User ID or Email
// Output: Complete User struct or nil
// Errors: Not found, database connection errors
func GetUserByID(id string) (*User, error)
func GetUserByEmail(email string) (*User, error)
```

**Update**:
```go
// Input: User ID, fields to update
// Output: Updated User struct
// Errors: Not found, validation failures
func UpdateUser(id string, updates map[string]interface{}) (*User, error)
```

**Delete**:
```go
// Input: User ID
// Output: Success boolean
// Errors: Not found, cascade failures
func DeleteUser(id string) error
```

---

## 🔒 **Security Contract**

### **Password Handling**
❌ **NEVER** return `password_hash` in any query result  
✅ **ALWAYS** use `json:"-"` tag on PasswordHash field  
✅ **ALWAYS** hash passwords before INSERT/UPDATE  
✅ **ALWAYS** use bcrypt with cost factor ≥ 10

### **Token Handling**
❌ **NEVER** expose verification tokens in JSON  
❌ **NEVER** expose JWT token IDs in responses  
✅ **ALWAYS** check token expiration in queries  
✅ **ALWAYS** invalidate tokens after use

### **Sensitive Data**
❌ **NEVER** log password hashes or tokens  
✅ **ALWAYS** use prepared statements (prevent SQL injection)  
✅ **ALWAYS** validate input before database operations

---

## ⚡ **Performance Contract**

### **Required Indexes**
```sql
-- Email lookups (most common)
CREATE INDEX idx_users_email ON users(email);

-- Session validation (very frequent)
CREATE INDEX idx_sessions_token_id ON user_sessions(token_id);
CREATE INDEX idx_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON user_sessions(expires_at);

-- OAuth lookups
CREATE INDEX idx_users_oauth ON users(oauth_provider, oauth_provider_id);

-- Active user queries
CREATE INDEX idx_users_active ON users(is_active);
```

### **Query Performance Expectations**
- Email lookup: <5ms
- Session validation: <3ms
- User creation: <50ms (includes password hashing)
- OAuth user lookup: <5ms

---

## 🔄 **Transaction Boundaries**

### **User Registration**
```go
// TRANSACTION REQUIRED: User creation + Email token
tx.Begin()
    CreateUser(userData)
    GenerateEmailVerificationToken(userID)
tx.Commit()
```

### **Login**
```go
// TRANSACTION REQUIRED: Session creation + Last login update
tx.Begin()
    CreateSession(userID, deviceInfo)
    UpdateLastLogin(userID)
tx.Commit()
```

### **Email Verification**
```go
// TRANSACTION REQUIRED: Mark verified + Invalidate token
tx.Begin()
    MarkEmailVerified(userID)
    InvalidateVerificationToken(token)
tx.Commit()
```

---

## ❌ **Error Handling Contract**

### **Standard Errors**
```go
// Not Found Errors
ErrUserNotFound = errors.New("user not found")
ErrSessionNotFound = errors.New("session not found")

// Constraint Violations
ErrDuplicateEmail = errors.New("email already exists")
ErrInvalidForeignKey = errors.New("invalid user reference")

// Connection Errors
ErrDatabaseConnection = errors.New("database connection failed")
ErrTimeout = errors.New("query timeout")
```

### **Error Propagation Rules**
✅ Return specific errors for constraint violations  
✅ Return `sql.ErrNoRows` as-is (caller decides semantics)  
✅ Wrap database errors with context  
❌ Never expose raw SQL errors to upper layers

---

## 📝 **Migration Contract**

### **Schema Evolution Rules**
✅ All schema changes via migrations (no manual ALTER TABLE)  
✅ Migrations must be reversible (provide DOWN migration)  
✅ Never delete columns (mark as deprecated, remove later)  
✅ Always provide default values for new columns  
✅ Test migrations on backup before production

### **Breaking Changes**
🚫 Renaming tables/columns requires coordination with code  
🚫 Changing column types requires data migration plan  
🚫 Removing columns requires code deployed first  
✅ Adding columns is safe (with defaults)

---

## 🔗 **Connection Pool Configuration**

### **Expected Settings**
```go
MaxOpenConns: 25      // Maximum connections
MaxIdleConns: 5       // Idle connection pool
ConnMaxLifetime: 5m   // Connection recycling
ConnMaxIdleTime: 2m   // Idle timeout
```

### **Connection Management**
✅ Always use `database/sql` package  
✅ Never hold connections longer than necessary  
✅ Use context for query timeouts  
✅ Close rows/statements in defer blocks

---

## 📊 **Data Type Mapping**

| PostgreSQL Type | Go Type | Notes |
|----------------|---------|-------|
| UUID | string | Use `uuid.Parse()` for validation |
| VARCHAR/TEXT | string | - |
| INTEGER | int | - |
| BIGINT | int64 | - |
| BOOLEAN | bool | - |
| TIMESTAMP | time.Time | Use UTC |
| TIMESTAMPTZ | time.Time | Preferred for timestamps |
| JSONB | json.RawMessage | Unmarshal when needed |
| NUMERIC | float64 | Or use `decimal` library |

---

## ✅ **Validation Contract**

### **Pre-Insert Validation**
✅ Email format validation  
✅ Required fields present  
✅ UUID format validation (if provided)  
✅ Timestamp validity checks  
✅ Foreign key existence checks

### **Post-Query Validation**
✅ Check for nil pointers before dereferencing  
✅ Validate timestamps (not in future)  
✅ Verify boolean flags consistency  
✅ Check JSONB structure validity

---

**Last Updated**: October 14, 2025  
**Maintained By**: Development Team  
**Breaking Changes**: Any changes to this contract require review

