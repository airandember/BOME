# 📊 Database Schema: `users` Table
**Primary authentication table for user accounts**

---

## 📍 **Source Files**

**Migrations**:
- `backend/migrations/005_add_user_profile_fields.sql`
- `backend/migrations/034_create_user_subscriptions_and_email_verification.sql`
- Various enhancement migrations

**Database Models**:
- `backend/internal/database/user.go`

---

## 🗄️ **Table Structure**

```sql
CREATE TABLE users (
    -- Primary Identity
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),              -- NULL for OAuth-only accounts
    
    -- Profile Information
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    profile_picture_url TEXT,
    notes TEXT,                               -- Admin notes
    
    -- Email Verification
    email_verified BOOLEAN DEFAULT FALSE,
    email_verified_at TIMESTAMP,
    
    -- Session Management
    last_login TIMESTAMP,
    last_logout TIMESTAMP,
    max_sessions INTEGER DEFAULT 5,
    
    -- Account Status
    role VARCHAR(50) DEFAULT 'user',          -- user, admin, advertiser, etc.
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- OAuth Integration (from oauth2_accounts table linkage)
    -- Note: OAuth data now primarily in oauth2_accounts table
);
```

---

## 🔑 **Primary Key**

**Column**: `id`  
**Type**: SERIAL (auto-incrementing integer)  
**Purpose**: Unique identifier for each user  
**Usage**: Referenced by all user-related foreign keys

---

## 📋 **Column Definitions**

### **Identity Columns**

#### `email` VARCHAR(255) UNIQUE NOT NULL
- **Purpose**: User's email address (primary login identifier)
- **Constraints**: 
  - UNIQUE (enforced at database level)
  - NOT NULL
  - Case-sensitive (normalized to lowercase in application)
- **Validation**: Must be valid email format
- **Index**: Automatically indexed (UNIQUE constraint)
- **Usage**: Used for login, password reset, notifications

#### `password_hash` VARCHAR(255)
- **Purpose**: Bcrypt hash of user's password
- **Can be NULL**: Yes (for OAuth-only accounts or during email verification flow)
- **Algorithm**: Bcrypt with cost factor ≥ 10
- **Security**: 
  - ❌ NEVER log this field
  - ❌ NEVER return in API responses
  - ✅ Always use json:"-" tag in Go struct
- **Validation**: Set only after email verification for new signups

---

### **Profile Columns**

#### `first_name` VARCHAR(100)
- **Purpose**: User's first name
- **Required**: Collected during registration
- **Validation**: 1-100 characters
- **Display**: Used in UI greetings, email personalization

#### `last_name` VARCHAR(100)
- **Purpose**: User's last name
- **Required**: Collected during registration
- **Validation**: 1-100 characters
- **Display**: Combined with first_name for full name display

#### `profile_picture_url` TEXT
- **Purpose**: URL to user's profile image
- **Sources**: 
  - OAuth provider (Google profile pic)
  - User upload to Digital Ocean Spaces
  - Default avatar if not set
- **Validation**: Valid URL format
- **Storage**: Actual images stored in Digital Ocean Spaces/CDN

#### `notes` TEXT
- **Purpose**: Admin-only notes about the user
- **Visibility**: Only visible to admin users
- **Usage**: Support notes, account status, special considerations
- **Added**: Migration 038_add_notes_to_users.sql

---

### **Email Verification Columns**

#### `email_verified` BOOLEAN DEFAULT FALSE
- **Purpose**: Whether user has verified their email
- **Default**: FALSE (unverified)
- **Flow**: 
  1. User registers → email_verified = FALSE
  2. Email sent with verification token
  3. User clicks link → email_verified = TRUE
- **Impact**: Used in middleware to block unverified users from certain actions
- **Added**: Migration 034

#### `email_verified_at` TIMESTAMP
- **Purpose**: Timestamp when email was verified
- **Can be NULL**: Yes (NULL if not yet verified)
- **Usage**: Audit trail, compliance reporting
- **Set**: When user completes email verification flow
- **Added**: Migration 034

---

### **Session Management Columns**

#### `last_login` TIMESTAMP
- **Purpose**: Most recent successful login timestamp
- **Updated**: Every successful login via `auth.go:LoginHandler`
- **Usage**: 
  - Security monitoring (detect suspicious activity)
  - Analytics (user engagement tracking)
  - Compliance (audit trails)

#### `last_logout` TIMESTAMP
- **Purpose**: Most recent logout timestamp
- **Updated**: When user explicitly logs out
- **Usage**: Session cleanup, security auditing
- **Added**: Migration 005

#### `max_sessions` INTEGER DEFAULT 5
- **Purpose**: Maximum concurrent sessions allowed for this user
- **Default**: 5 sessions
- **Enforcement**: Checked in `middleware.go` before creating new session
- **Admin Override**: Can be increased for specific users
- **Usage**: Prevent account sharing, security enforcement
- **Added**: Migration 005

---

### **Account Status Columns**

#### `role` VARCHAR(50) DEFAULT 'user'
- **Purpose**: User's role for RBAC (Role-Based Access Control)
- **Default**: 'user'
- **Possible Values**:
  - `'user'` - Standard subscriber
  - `'admin'` - Full admin access
  - `'advertiser'` - Advertiser account
  - `'super_admin'` - Super admin (all permissions)
  - Additional roles in `roles` table
- **Enforcement**: Checked in `middleware.go:RequireRole()`
- **Standardization**: Migration 006_standardize_roles.sql
- **Related Tables**: May join with `roles` table for granular permissions

#### `is_active` BOOLEAN DEFAULT TRUE
- **Purpose**: Whether account is active (soft delete)
- **Default**: TRUE (active)
- **Usage**: 
  - FALSE = Account suspended/disabled
  - Checked in authentication middleware
  - Prevents login for disabled accounts
- **Recovery**: Can be re-activated by admin

---

### **Timestamp Columns**

#### `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
- **Purpose**: When user account was created
- **Set**: Automatically on INSERT
- **Immutable**: Never updated after creation
- **Usage**: 
  - Analytics (user acquisition tracking)
  - Compliance (data retention policies)
  - Sorting (newest users)

#### `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
- **Purpose**: Last time any user data was modified
- **Updated**: Via trigger on UPDATE operations
- **Trigger**: `update_updated_at_column()` function
- **Usage**: Cache invalidation, audit trails
- **Maintenance**: Automatically maintained by database trigger

---

## 🔍 **Indexes**

### **Existing Indexes**

```sql
-- Primary key index (automatic)
CREATE UNIQUE INDEX users_pkey ON users(id);

-- Email uniqueness (automatic with UNIQUE constraint)
CREATE UNIQUE INDEX users_email_key ON users(email);

-- Role-based queries (added via standardization)
CREATE INDEX idx_users_role ON users(role);

-- Active users (for filtering)
CREATE INDEX idx_users_active ON users(is_active);

-- Email verification status (common filter)
CREATE INDEX idx_users_email_verified ON users(email_verified);
```

### **Recommended Additional Indexes** ⚠️
```sql
-- For login timestamp queries
CREATE INDEX idx_users_last_login ON users(last_login) WHERE is_active = TRUE;

-- For recently created users
CREATE INDEX idx_users_created_at ON users(created_at DESC);

-- Composite for active verified users (common query)
CREATE INDEX idx_users_active_verified ON users(is_active, email_verified);
```

---

## 🔗 **Foreign Key Relationships**

### **Referenced By (Other Tables)**
```sql
-- User sessions
user_sessions.user_id → users.id (ON DELETE CASCADE)

-- OAuth accounts
oauth2_accounts.user_id → users.id (ON DELETE CASCADE)

-- Email verification tokens
email_verification_tokens.user_id → users.id (ON DELETE CASCADE)

-- User subscriptions
user_subscriptions.user_id → users.id (ON DELETE CASCADE)

-- Analytics events
analytics_events.user_id → users.id (ON DELETE SET NULL)

-- And many more...
```

### **Cascade Behavior**
⚠️ **ON DELETE CASCADE**: Deleting a user will automatically delete:
- All sessions
- All OAuth accounts
- All verification tokens
- All subscriptions
- All user-generated content

---

## ⚙️ **Triggers**

### **updated_at Trigger**
```sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

**Purpose**: Automatically updates `updated_at` on any UPDATE operation  
**Behavior**: Runs before UPDATE, sets current timestamp  
**Note**: Cannot be disabled without dropping trigger

---

## 🛡️ **Security Considerations**

### **Sensitive Data Protection**

**Password Hash**:
- ❌ NEVER return in API responses
- ❌ NEVER log or display
- ✅ Use `json:"-"` tag in Go struct
- ✅ Hash with bcrypt cost ≥ 10

**Email Privacy**:
- Emails are PII (Personally Identifiable Information)
- Consider masking in logs
- Apply data retention policies

**Notes Field**:
- Admin-only visibility
- May contain sensitive information
- Audit access to this field

---

## 📊 **Data Patterns**

### **New User Registration Flow**
```sql
INSERT INTO users (
    email, 
    first_name, 
    last_name, 
    email_verified, 
    password_hash
) VALUES (
    'user@example.com',
    'John',
    'Doe',
    FALSE,              -- Unverified initially
    NULL                -- Password set after verification
);
```

### **Email Verification Completion**
```sql
UPDATE users 
SET 
    email_verified = TRUE,
    email_verified_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND email_verified = FALSE;
```

### **Password Setup (Post-Verification)**
```sql
UPDATE users 
SET 
    password_hash = $1,  -- Bcrypt hash
    updated_at = CURRENT_TIMESTAMP
WHERE id = $2 AND email_verified = TRUE;
```

### **Login Success Update**
```sql
UPDATE users 
SET 
    last_login = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;
```

---

## 📈 **Performance Considerations**

### **Query Optimization**

**Most Frequent Queries**:
1. `SELECT * FROM users WHERE email = $1` (login)
2. `SELECT * FROM users WHERE id = $1` (auth middleware)
3. `SELECT * FROM users WHERE email_verified = FALSE` (admin dashboard)
4. `SELECT * FROM users WHERE role = $1` (RBAC checks)

**Performance Expectations**:
- Email lookup: <5ms (indexed)
- ID lookup: <2ms (primary key)
- Role queries: <10ms (indexed)

### **Optimization Strategies**:
✅ Email column is UNIQUE (automatic index)  
✅ Role column should be indexed  
⚠️ Consider partitioning if user count > 10M  
⚠️ Monitor query performance for complex JOINs

---

## 🔧 **Maintenance**

### **Data Cleanup**
```sql
-- Find inactive users (no login in 1 year)
SELECT id, email, last_login 
FROM users 
WHERE last_login < NOW() - INTERVAL '1 year'
  AND is_active = TRUE;

-- Find unverified accounts older than 30 days
SELECT id, email, created_at 
FROM users 
WHERE email_verified = FALSE 
  AND created_at < NOW() - INTERVAL '30 days';
```

### **Health Checks**
```sql
-- Check for duplicate emails (should be 0)
SELECT email, COUNT(*) 
FROM users 
GROUP BY email 
HAVING COUNT(*) > 1;

-- Check for users without password or OAuth
SELECT id, email 
FROM users 
WHERE password_hash IS NULL 
  AND id NOT IN (SELECT DISTINCT user_id FROM oauth2_accounts);
```

---

## 📝 **Migration History**

| Migration | Description |
|-----------|-------------|
| `initial` | Created users table |
| `005_add_user_profile_fields.sql` | Added profile fields, session management |
| `006_standardize_roles.sql` | Standardized role values |
| `034_create_user_subscriptions_and_email_verification.sql` | Added email_verified, email_verified_at |
| `038_add_notes_to_users.sql` | Added notes field for admin |

---

## 🚨 **Known Issues / Technical Debt**

### **🟡 Medium Priority**
1. **Role System Evolution**
   - Currently using VARCHAR(50) for roles
   - Future: May need `user_roles` join table for multiple roles per user
   - **Action**: Document plans for role system refactor

2. **Email Case Sensitivity**
   - Database stores case-sensitive emails
   - Application normalizes to lowercase
   - **Risk**: Case variations could cause issues
   - **Action**: Consider LOWER() index or database constraint

3. **Soft Delete Pattern**
   - Using `is_active = FALSE` for soft deletes
   - Deleted users still occupy space
   - **Action**: Consider adding `deleted_at` timestamp

### **🟢 Low Priority**
1. Profile picture URL could be in separate media table
2. `max_sessions` rarely used, could be in settings table
3. Notes field could be in separate admin_notes table

---

**Last Updated**: October 14, 2025  
**Schema Version**: 1.38 (migration 038)  
**Status**: ✅ Production-ready, well-tested

