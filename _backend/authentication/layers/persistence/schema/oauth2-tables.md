# 📊 Database Schema: OAuth2 Tables
**OAuth2 authentication and account linking system**

---

## 📍 **Source Files**

**Migrations**:
- `backend/migrations/035_create_oauth2_tables.sql`
- `backend/migrations/037_create_oauth2_states_table.sql`
- `backend/migrations/036_cleanup_oauth2_redirect_url.sql`

**Database Models**: `backend/internal/database/*oauth*.go`  
**Service Logic**: `backend/internal/services/oauth2.go`  
**Routes**: `backend/internal/routes/oauth2_routes.go`

---

## 🗄️ **Table Overview**

This OAuth2 system consists of 3 main tables:

1. **`oauth2_accounts`** - Links OAuth2 provider accounts to local users
2. **`oauth2_states`** - CSRF protection and OAuth2 state management
3. **`oauth2_settings`** - OAuth2 provider configurations

---

## 1️⃣ **oauth2_accounts Table**

### **Purpose**
Links external OAuth2 provider accounts (Google, Facebook, etc.) to local user accounts. Supports:
- Social login (OAuth2 as primary auth)
- Account linking (add OAuth2 to existing account)
- Multiple providers per user

### **Table Structure**
```sql
CREATE TABLE oauth2_accounts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,              -- 'google', 'facebook', etc.
    provider_user_id VARCHAR(255) NOT NULL,     -- OAuth2 provider's user ID
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    picture TEXT,                               -- Profile picture URL
    access_token TEXT,                          -- Encrypted OAuth2 access token (optional)
    refresh_token TEXT,                         -- Encrypted OAuth2 refresh token (optional)
    token_expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(user_id, provider),                  -- One account per provider per user
    UNIQUE(provider, provider_user_id)          -- One provider account per user ID
);
```

### **Column Definitions**

#### `user_id` INTEGER NOT NULL
- **Foreign Key**: `REFERENCES users(id) ON DELETE CASCADE`
- **Purpose**: Link OAuth2 account to local user
- **Cascade**: Deleting user removes all OAuth2 account links
- **Index**: `idx_oauth2_accounts_user_id`

#### `provider` VARCHAR(50) NOT NULL
- **Purpose**: Identify OAuth2 provider
- **Values**: `'google'`, `'facebook'`, `'github'`, `'microsoft'`, etc.
- **Case**: Lowercase convention
- **Index**: `idx_oauth2_accounts_provider`
- **Usage**: Filter accounts by provider

#### `provider_user_id` VARCHAR(255) NOT NULL
- **Purpose**: User ID from OAuth2 provider (e.g., Google's user ID)
- **Format**: Provider-specific (often numeric or UUID)
- **Uniqueness**: `UNIQUE(provider, provider_user_id)` - One provider account = One local user
- **Usage**: 
  - Lookup existing OAuth2 account during login
  - Prevent duplicate OAuth2 account linking
- **Example**: Google ID like `"108123456789012345678"`

#### `email` VARCHAR(255) NOT NULL
- **Purpose**: Email address from OAuth2 provider
- **Not Unique**: Same email might exist across providers
- **Usage**: 
  - Account matching (find if user exists by email)
  - Display in UI ("Connected Google: user@gmail.com")
- **Index**: `idx_oauth2_accounts_email`
- **Note**: May differ from `users.email` if user changed email

#### `name` VARCHAR(255)
- **Purpose**: Full name from OAuth2 provider
- **Optional**: Can be NULL
- **Usage**: 
  - Pre-fill registration form
  - Display in account linking UI
- **Example**: "John Doe"

#### `picture` TEXT
- **Purpose**: Profile picture URL from OAuth2 provider
- **Optional**: Can be NULL
- **Usage**: 
  - Set as user's profile picture
  - Display in account management UI
- **Example**: "https://lh3.googleusercontent.com/a/..."

#### `access_token` TEXT
- **Purpose**: OAuth2 access token (optional storage)
- **Security**: ⚠️ Should be encrypted if stored
- **Optional**: Can be NULL (many apps don't store this)
- **Usage**: 
  - Call provider APIs on user's behalf
  - Access user data from provider
- **Expiration**: Check `token_expires_at`
- **Note**: Consider NOT storing tokens unless needed

#### `refresh_token` TEXT
- **Purpose**: OAuth2 refresh token (optional storage)
- **Security**: ⚠️ MUST be encrypted if stored
- **Optional**: Can be NULL
- **Long-lived**: Used to obtain new access tokens
- **Critical**: If compromised, attacker can access provider account
- **Best Practice**: Store only if absolutely necessary

#### `token_expires_at` TIMESTAMP WITH TIME ZONE
- **Purpose**: When the `access_token` expires
- **Optional**: NULL if not storing tokens
- **Usage**: Know when to refresh access token
- **Timezone Aware**: Uses TIMESTAMP WITH TIME ZONE

### **Unique Constraints**

```sql
-- Each user can have only ONE account per provider
UNIQUE(user_id, provider)

-- Each provider account can link to only ONE local user
UNIQUE(provider, provider_user_id)
```

**Implications**:
- User can link Google + Facebook, but not 2 Google accounts
- One Google account cannot be linked to multiple local accounts
- Prevents account confusion and security issues

### **Indexes**
```sql
CREATE INDEX idx_oauth2_accounts_user_id ON oauth2_accounts(user_id);
CREATE INDEX idx_oauth2_accounts_provider ON oauth2_accounts(provider);
CREATE INDEX idx_oauth2_accounts_provider_user_id ON oauth2_accounts(provider, provider_user_id);
CREATE INDEX idx_oauth2_accounts_email ON oauth2_accounts(email);
```

---

## 2️⃣ **oauth2_states Table**

### **Purpose**
Stores OAuth2 state parameters for CSRF protection. Critical for production where in-memory state storage doesn't work across server instances.

**Problem Solved**: OAuth2 callback might hit a different server instance than the one that initiated the auth flow.

### **Table Structure**
```sql
CREATE TABLE oauth2_states (
    id SERIAL PRIMARY KEY,
    state VARCHAR(255) UNIQUE NOT NULL,
    provider VARCHAR(50) NOT NULL,
    return_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    state_data JSONB
);
```

### **Column Definitions**

#### `state` VARCHAR(255) UNIQUE NOT NULL
- **Purpose**: Random state parameter for CSRF protection
- **Format**: UUID or random string
- **Uniqueness**: MUST be unique across all states
- **Generated**: Application generates before OAuth2 redirect
- **Validated**: Checked when OAuth2 callback receives state
- **Security**: ⚠️ Critical for preventing CSRF attacks
- **Index**: `idx_oauth2_states_state`

#### `provider` VARCHAR(50) NOT NULL
- **Purpose**: Which OAuth2 provider this state is for
- **Values**: `'google'`, `'facebook'`, etc.
- **Usage**: Helps organize and debug OAuth2 flows

#### `return_url` TEXT
- **Purpose**: Where to redirect user after OAuth2 completion
- **Example**: `/dashboard`, `/account/settings`
- **Usage**: Maintain user's intended destination through OAuth2 flow
- **Security**: ⚠️ Validate this is a safe internal URL

#### `expires_at` TIMESTAMP WITH TIME ZONE NOT NULL
- **Purpose**: When this state expires
- **Typical TTL**: 10 minutes (OAuth2 flow should be quick)
- **Cleanup**: Expired states should be deleted regularly
- **Security**: Short TTL limits CSRF attack window
- **Index**: `idx_oauth2_states_expires_at`

#### `state_data` JSONB
- **Purpose**: Additional data to pass through OAuth2 flow
- **Format**: Flexible JSON object
- **Examples**:
  ```json
  {
    "action": "link",  // or "login"
    "user_id": 123,    // if linking to existing account
    "referrer": "/premium",
    "metadata": {...}
  }
  ```
- **Usage**: Store context for OAuth2 completion handler

### **Indexes**
```sql
CREATE INDEX idx_oauth2_states_state ON oauth2_states(state);
CREATE INDEX idx_oauth2_states_expires_at ON oauth2_states(expires_at);
```

### **Data Flow**

**1. Initiate OAuth2 Login**:
```go
state := uuid.New().String()
db.Exec(`
    INSERT INTO oauth2_states (state, provider, return_url, expires_at)
    VALUES ($1, $2, $3, NOW() + INTERVAL '10 minutes')
`, state, "google", "/dashboard")

// Redirect to Google with state parameter
redirectURL := fmt.Sprintf("%s&state=%s", googleAuthURL, state)
```

**2. OAuth2 Callback**:
```go
stateParam := r.URL.Query().Get("state")

// Validate state exists and not expired
var returnURL string
err := db.QueryRow(`
    SELECT return_url 
    FROM oauth2_states 
    WHERE state = $1 AND expires_at > NOW()
`, stateParam).Scan(&returnURL)

if err != nil {
    return errors.New("invalid or expired state")
}

// State is valid, delete it (one-time use)
db.Exec("DELETE FROM oauth2_states WHERE state = $1", stateParam)

// Continue OAuth2 flow...
```

---

## 3️⃣ **oauth2_settings Table**

### **Purpose**
Store OAuth2 provider configurations (client IDs, secrets, URLs). Allows runtime configuration without code changes.

### **Table Structure**
```sql
CREATE TABLE oauth2_settings (
    id SERIAL PRIMARY KEY,
    provider VARCHAR(50) NOT NULL UNIQUE,
    client_id VARCHAR(255) NOT NULL,
    client_secret TEXT NOT NULL,                -- Encrypted
    redirect_url VARCHAR(500) NOT NULL,
    auth_url VARCHAR(500),
    token_url VARCHAR(500),
    user_info_url VARCHAR(500),
    scopes TEXT,                                -- JSON array of scopes
    is_enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### **Column Definitions**

#### `provider` VARCHAR(50) UNIQUE NOT NULL
- **Purpose**: OAuth2 provider name
- **Uniqueness**: Only one config per provider
- **Values**: `'google'`, `'facebook'`, `'github'`, etc.

#### `client_id` VARCHAR(255) NOT NULL
- **Purpose**: OAuth2 client ID from provider console
- **Public**: Not sensitive (but don't publish unnecessarily)
- **Example**: `"123456789-abc123.apps.googleusercontent.com"`

#### `client_secret` TEXT NOT NULL
- **Purpose**: OAuth2 client secret from provider console
- **Security**: ⚠️⚠️⚠️ MUST BE ENCRYPTED
- **Critical**: If leaked, attacker can impersonate your app
- **Best Practice**: Encrypt before storing, decrypt only when needed

#### `redirect_url` VARCHAR(500) NOT NULL
- **Purpose**: OAuth2 callback URL (must match provider console)
- **Format**: Full URL with protocol
- **Example**: `"https://yourdomain.com/auth/oauth2/callback/google"`
- **Environment-Specific**: Different for dev/staging/prod

#### `auth_url`, `token_url`, `user_info_url` VARCHAR(500)
- **Purpose**: Provider-specific OAuth2 endpoints
- **Optional**: Can be hardcoded in application
- **Usage**: Dynamic provider addition without code changes

#### `scopes` TEXT
- **Purpose**: OAuth2 scopes to request
- **Format**: JSON array string
- **Example**: `'["email", "profile", "https://www.googleapis.com/auth/userinfo.email"]'`
- **Usage**: Define what data your app can access

#### `is_enabled` BOOLEAN DEFAULT TRUE
- **Purpose**: Enable/disable provider without deleting config
- **Usage**: 
  - Temporarily disable problematic provider
  - A/B testing OAuth2 providers
  - Gradual rollout of new providers

### **Indexes**
```sql
CREATE INDEX idx_oauth2_settings_provider ON oauth2_settings(provider);
CREATE INDEX idx_oauth2_settings_enabled ON oauth2_settings(is_enabled);
```

---

## 🔄 **OAuth2 Authentication Flow**

### **Full Flow Diagram**
```
1. User clicks "Login with Google"
   ↓
2. App generates state, saves to oauth2_states table
   ↓
3. Redirect user to Google with state parameter
   ↓
4. User authorizes on Google
   ↓
5. Google redirects back with code + state
   ↓
6. App validates state from oauth2_states table
   ↓
7. App exchanges code for access token
   ↓
8. App fetches user info from Google
   ↓
9. App checks if Google account exists in oauth2_accounts
   ↓
10a. If exists: Login existing user
10b. If new: Create user + oauth2_accounts entry
   ↓
11. Generate JWT, create session, login user
```

---

## 🛡️ **Security Considerations**

### **Token Storage**
❌ **Avoid storing** access/refresh tokens unless necessary  
⚠️ **If storing**, MUST encrypt:
  - Use AES-256 encryption
  - Store encryption key in environment variable (not database)
  - Decrypt only when needed, never return in API

### **Client Secret Protection**
🔒 **CRITICAL**: `oauth2_settings.client_secret` MUST be encrypted  
⚠️ **Never** log or display client secrets  
✅ **Rotate** secrets periodically  
✅ **Monitor** for unauthorized access

### **CSRF Protection**
✅ Always generate random state parameter  
✅ Validate state on callback  
✅ Use short expiration (10 minutes)  
✅ Delete state after use (one-time use)

### **Account Linking Security**
⚠️ Verify user is authenticated before linking OAuth2 account  
⚠️ Prevent account takeover via OAuth2 email matching  
⚠️ Require password or re-authentication for unlinking

---

## 🧹 **Maintenance**

### **Cleanup Expired States**
```sql
-- Run every hour
DELETE FROM oauth2_states 
WHERE expires_at < NOW();
```

### **Find Orphaned OAuth2 Accounts**
```sql
-- OAuth2 accounts with no corresponding user
SELECT * FROM oauth2_accounts 
WHERE user_id NOT IN (SELECT id FROM users);
-- Should be 0 (CASCADE handles this)
```

### **Token Expiration Monitoring**
```sql
-- Find accounts with expired tokens
SELECT 
    user_id, 
    provider, 
    token_expires_at 
FROM oauth2_accounts 
WHERE token_expires_at < NOW()
  AND access_token IS NOT NULL;
```

---

## ⚠️ **Known Issues / Technical Debt**

### **🟡 Medium Priority**

1. **Token Encryption Not Implemented**
   - `access_token` and `refresh_token` stored as plaintext
   - **Risk**: High if database compromised
   - **Action**: Implement encryption before storing tokens

2. **Provider Settings in Database**
   - `oauth2_settings` table vs environment variables
   - **Trade-off**: Flexibility vs security
   - **Consider**: Move sensitive settings to env vars

3. **No Token Refresh Logic**
   - Storing tokens but not refreshing them
   - **Impact**: Tokens expire, API calls fail
   - **Action**: Implement refresh token flow

### **🟢 Low Priority**

1. OAuth2 provider discovery (auto-configure from .well-known)
2. Support for multiple redirect URLs per provider
3. OAuth2 scope management (dynamic scope requests)

---

## 📚 **Related Documentation**

- **Service**: `backend/internal/services/oauth2.go`
- **Routes**: `backend/internal/routes/oauth2_routes.go`
- **Frontend**: `frontend/src/lib/auth.ts` (OAuth2 buttons)
- **Migrations**: 035, 036, 037

---

**Last Updated**: October 14, 2025  
**Migrations**: 035, 036, 037  
**Status**: ✅ Production-ready (except token encryption)  
**Security Level**: 🟡 Medium (needs token encryption)

