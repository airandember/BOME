# 🎯 BOME CODEBASE STANDARDS & GLOSSARY

**Last Updated**: October 26, 2025  
**Purpose**: Single source of truth for naming conventions, architecture patterns, and development standards

---

## 📚 **TABLE OF CONTENTS**

1. [Architecture Patterns](#architecture-patterns)
2. [Package Structure](#package-structure)
3. [Naming Conventions](#naming-conventions)
4. [Database Standards](#database-standards)
5. [Service Layer Standards](#service-layer-standards)
6. [API Route Standards](#api-route-standards)
7. [Frontend Standards](#frontend-standards)
8. [Common Pitfalls & Solutions](#common-pitfalls--solutions)

---

## 🏗️ **ARCHITECTURE PATTERNS**

### **BRAIDS Architecture**

```
BRAID (Business Requirement And Integrated Domain)
├── Elastic Services (Backend - Single Source of Truth)
│   ├── Data access layer
│   ├── Business logic
│   └── External API integration
│
└── Strands (Frontend - UI Components)
    ├── Components consume elastic services
    ├── Client-side caching
    └── State management
```

**Key Principles**:
- ✅ **Single Responsibility**: Each service handles ONE domain
- ✅ **Elastic Services**: Backend services are the source of truth
- ✅ **Strands**: Frontend components consume elastic services
- ✅ **No Duplication**: Consolidate overlapping services

---

## 📁 **PACKAGE STRUCTURE**

### **Backend Directory Structure**

```
backend/
├── cmd/                          # CLI tools and main applications
│   ├── stripe-sync/             # Stripe sync CLI tool
│   └── create-test-user/        # Test user creation
│
├── infrastructure/               # ✅ USE THIS FOR CORE INFRASTRUCTURE
│   ├── config/                  # Configuration management
│   └── database/                # ✅ DATABASE CONNECTION (PRIMARY)
│       ├── database.go          # Main DB type and connection
│       └── redis.go             # Redis connection
│
├── internal/                    # ⚠️ INTERNAL PACKAGES (some legacy)
│   ├── config/                  # ⚠️ LEGACY - Use infrastructure/config
│   ├── database/                # ⚠️ DUPLICATE - Use infrastructure/database
│   ├── middleware/              # ✅ Auth, RBAC, session middleware
│   ├── routes/                  # ✅ API route handlers
│   └── services/                # ✅ Business logic services
│
├── migrations/                  # Database migrations (numbered)
└── services/                    # ⚠️ LEGACY - Moving to internal/services
```

### **🚨 CRITICAL: Database Package**

**ALWAYS USE**: `bome-backend/internal/database`

**❌ DO NOT USE**: `bome-backend/infrastructure/database` (duplicate package, causes type mismatches)

```go
// ✅ CORRECT (for ALL code - CLI, routes, services)
import "bome-backend/internal/config"
import "bome-backend/internal/database"

cfg := config.New()
db, err := database.New(cfg)
```

```go
// ❌ WRONG (will cause type mismatch errors)
import "bome-backend/infrastructure/database"
import "bome-backend/infrastructure/config"

// These create DIFFERENT types that Go won't accept as compatible
```

**Why?** Go treats `internal/database.DB` and `infrastructure/database.DB` as **completely different types**, even if they have identical structures. This causes compilation errors when passing DB instances between functions.

**🚨 KNOWN ISSUE**: `infrastructure/database` still exists in the codebase but should NOT be used.

**FUTURE TASK**: Remove `infrastructure/database` and `infrastructure/config` entirely to prevent confusion.

---

## 🏷️ **NAMING CONVENTIONS**

### **Services**

| Pattern | Example | When to Use |
|---------|---------|-------------|
| `{Domain}Service` | `StripeService` | Core domain service |
| `{Domain}ElasticService` | `SubscriberElasticService` | Unified elastic service (BRAIDS) |
| `{Domain}Service_v2` | `StripeSyncV2Service` | New version alongside old |
| `{Domain}PublicService` | `StripePublicService` | Public-facing API service |

### **Database Tables**

| Pattern | Example | When to Use |
|---------|---------|-------------|
| `{entity}` | `users`, `subscriptions` | Primary entities |
| `{entity}_v2` | `stripe_customers_v2` | New schema version |
| `user_{relation}` | `user_stripe_customers` | Link/junction tables |
| `{entity}_history` | `subscriber_history` | Audit trail tables |

### **API Routes**

| Pattern | Example | When to Use |
|---------|---------|-------------|
| `/admin/{resource}` | `/admin/users` | Admin-only resources |
| `/admin/{area}/{resource}` | `/admin/streaming/subscribers` | Grouped admin resources |
| `/api/v1/{resource}` | `/api/v1/auth/login` | Public API endpoints |
| `/webhooks/{provider}` | `/webhooks/stripe` | External webhooks (no auth) |

### **Function Names**

| Pattern | Example | When to Use |
|---------|---------|-------------|
| `Get{Entity}` | `GetUser(id)` | Fetch single entity |
| `GetAll{Entities}` | `GetAllSubscribers()` | Fetch all entities |
| `List{Entities}` | `ListUsers(limit, offset)` | Paginated list |
| `Create{Entity}` | `CreateUser(data)` | Create new entity |
| `Update{Entity}` | `UpdateUser(id, data)` | Update existing entity |
| `Delete{Entity}` | `DeleteUser(id)` | Delete entity |
| `Sync{Entity}` | `SyncCustomers()` | Sync from external source |
| `{Action}Handler` | `LoginHandler()` | HTTP handler function |

### **Variables**

```go
// Database connection
db *database.DB  // ✅ Always "db"

// Gin context
c *gin.Context   // ✅ Always "c"

// Configuration
cfg *config.Config  // ✅ Always "cfg"

// Services
stripeService *services.StripeService  // ✅ {domain}Service
elasticService *services.SubscriberElasticService  // ✅ elasticService
```

---

## 💾 **DATABASE STANDARDS**

### **Connection Initialization**

```go
// ✅ STANDARD WAY (use everywhere)
import (
    "bome-backend/infrastructure/config"
    "bome-backend/infrastructure/database"
)

func main() {
    cfg := config.New()
    db, err := database.New(cfg)
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer db.Close()
    
    // Use db...
}
```

### **Foreign Keys**

```sql
-- ✅ ALWAYS use foreign keys for referential integrity
CREATE TABLE stripe_subscriptions_v2 (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL REFERENCES stripe_customers_v2(id) ON DELETE CASCADE,
    price_id INTEGER NOT NULL REFERENCES stripe_prices_v2(id) ON DELETE RESTRICT
);
```

**Cascade Rules**:
- `ON DELETE CASCADE` - Delete children when parent deleted (e.g., subscriptions when customer deleted)
- `ON DELETE RESTRICT` - Prevent deletion if children exist (e.g., can't delete price if subscriptions exist)
- `ON DELETE SET NULL` - Null out FK when parent deleted (rare, be careful)

### **Indexes**

```sql
-- ✅ ALWAYS index foreign keys
CREATE INDEX idx_subscriptions_customer ON stripe_subscriptions_v2(customer_id);

-- ✅ Index frequently queried columns
CREATE INDEX idx_subscriptions_status ON stripe_subscriptions_v2(status);

-- ✅ Partial indexes for common queries
CREATE INDEX idx_subscriptions_active 
    ON stripe_subscriptions_v2(customer_id, status) 
    WHERE status IN ('active', 'trialing');
```

### **Naming Conventions**

```sql
-- Tables: lowercase, underscores
users, stripe_customers_v2, user_stripe_customers

-- Columns: lowercase, underscores
user_id, stripe_customer_id, created_at

-- Indexes: idx_{table}_{column(s)}
idx_users_email, idx_subscriptions_customer_status

-- Foreign Keys: fk_{table}_{referenced_table}
fk_subscriptions_customer, fk_prices_product

-- Unique Constraints: {table}_{column}_unique
stripe_customers_v2_stripe_id_unique
```

---

## 🔐 **SECURITY STANDARDS**

### **Sensitive Configuration Storage**

**🚨 CRITICAL**: Sensitive keys (API keys, secrets) are stored in the `secure_settings` table, NOT in environment variables or `.env` files.

```sql
-- secure_settings table structure
CREATE TABLE secure_settings (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

---

### **🔒 Secret Key Security Model: WRITE-ONLY FROM FRONTEND**

**CRITICAL RULE**: Secret keys are **WRITE-ONLY** from the frontend. They are **NEVER** read or returned to the frontend.

#### **Frontend: WRITE Only (Admin UI)**

```typescript
// ✅ CORRECT - Frontend can UPDATE secrets (admin only)
async function updateStripeKey(newKey: string) {
    await apiRequest('/admin/settings/stripe-key', {
        method: 'POST',
        body: { value: newKey }
    });
    // Backend stores it, frontend never sees it again
}
```

```typescript
// ❌ WRONG - Frontend should NEVER retrieve secrets
const key = await apiRequest('/admin/settings/stripe-key');  // This endpoint should NOT exist!
```

#### **Backend: READ for Internal Use Only**

```go
// ✅ CORRECT - Backend reads from database for internal use
var stripeKey string
err := db.QueryRow(`
    SELECT value FROM secure_settings 
    WHERE key = 'stripe_secret_key'
`).Scan(&stripeKey)

if err != nil {
    // Fallback to environment variable (dev only)
    stripeKey = os.Getenv("STRIPE_SECRET_KEY")
}

// Use the key for Stripe API calls
stripe.Key = stripeKey
```

```go
// ❌ WRONG - Backend should NEVER return secrets in API responses
func GetSettings(c *gin.Context) {
    var key string
    db.QueryRow("SELECT value FROM secure_settings WHERE key = 'stripe_secret_key'").Scan(&key)
    
    c.JSON(200, gin.H{
        "stripe_key": key  // ❌ NEVER DO THIS!
    })
}
```

#### **Admin Settings Update Endpoint Pattern**

```go
// ✅ CORRECT - Admin can UPDATE, but never READ back
func UpdateSecureSetting(c *gin.Context) {
    var req struct {
        Key   string `json:"key" binding:"required"`
        Value string `json:"value" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    // Validate admin permission
    userRole := c.GetString("user_role")
    if userRole != "super_admin" {
        c.JSON(403, gin.H{"error": "Only super_admins can update secrets"})
        return
    }
    
    // Store in database (WRITE)
    _, err := db.Exec(`
        INSERT INTO secure_settings (key, value, updated_at)
        VALUES ($1, $2, NOW())
        ON CONFLICT (key) DO UPDATE SET
            value = EXCLUDED.value,
            updated_at = NOW()
    `, req.Key, req.Value)
    
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to update setting"})
        return
    }
    
    // ✅ CORRECT - Confirm success WITHOUT returning the value
    c.JSON(200, gin.H{
        "status": "success",
        "message": fmt.Sprintf("Setting '%s' updated successfully", req.Key),
        // ❌ DO NOT INCLUDE: "value": req.Value
    })
}
```

#### **Admin Settings READ Endpoint (Safe)**

```go
// ✅ CORRECT - Return metadata ONLY, not values
func GetSecureSettingsList(c *gin.Context) {
    rows, err := db.Query(`
        SELECT key, created_at, updated_at
        FROM secure_settings
        ORDER BY key
    `)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to get settings"})
        return
    }
    defer rows.Close()
    
    var settings []gin.H
    for rows.Next() {
        var key string
        var createdAt, updatedAt time.Time
        rows.Scan(&key, &createdAt, &updatedAt)
        
        settings = append(settings, gin.H{
            "key": key,
            "created_at": createdAt,
            "updated_at": updatedAt,
            // ✅ CORRECT - Show that it EXISTS, but not the value
            "is_configured": true,
            // ❌ DO NOT INCLUDE: "value": actualValue
        })
    }
    
    c.JSON(200, gin.H{
        "status": "success",
        "data": settings,
    })
}
```

---

### **Common Secure Settings Keys**

**Secret Keys (NEVER returned to frontend)**:
- `stripe_secret_key` - Stripe API secret key (sk_live_xxx or sk_test_xxx)
- `stripe_webhook_secret` - Stripe webhook signing secret
- `jwt_secret` - JWT signing secret
- `encryption_key` - Data encryption key
- `database_password` - Database credentials (if stored)

**Public Keys (CAN be returned to frontend)**:
- `stripe_publishable_key` - Stripe publishable key (pk_live_xxx or pk_test_xxx)
- `bunny_stream_library_id` - Public library ID
- `recaptcha_site_key` - Public reCAPTCHA site key

---

### **Security Checklist**

When implementing settings management:

- [ ] **Frontend WRITE only** - Admins can update, never read back
- [ ] **Backend READ for internal use** - Services fetch from DB
- [ ] **API responses NEVER include secret values** - Only confirmation/metadata
- [ ] **Super admin only** - Only highest privilege can update secrets
- [ ] **Audit logging** - Log who updated what and when
- [ ] **No console.log** - Never log secret values
- [ ] **No error messages with values** - Don't leak secrets in error messages

---

## 🔧 **SERVICE LAYER STANDARDS**

### **Service Structure**

```go
// ✅ STANDARD SERVICE PATTERN
type {Domain}Service struct {
    db *database.DB
    // other dependencies
}

func New{Domain}Service(db *database.DB) *{Domain}Service {
    return &{Domain}Service{db: db}
}

// Public methods
func (s *{Domain}Service) Get{Entity}(id int) (*{Entity}, error) {
    // implementation
}
```

### **Error Handling**

```go
// ✅ RETURN wrapped errors with context
func (s *Service) GetUser(id int) (*User, error) {
    var user User
    err := s.db.QueryRow("SELECT ...").Scan(...)
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("user not found: %d", id)
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get user %d: %w", id, err)
    }
    return &user, nil
}
```

### **Logging Standards**

```go
// ✅ STRUCTURED LOGGING with context
log.Printf("🚀 [Service] Starting operation: userID=%d", userID)
log.Printf("✅ [Service] Operation successful: count=%d", count)
log.Printf("❌ [Service] Operation failed: %v", err)
log.Printf("⚠️ [Service] Warning: potential issue detected")
```

**Emoji Guide**:
- 🚀 Starting/Initializing
- ✅ Success
- ❌ Error/Failure
- ⚠️ Warning
- 🔍 Debug/Investigating
- 📊 Statistics/Metrics
- 👥 Users
- 📦 Products
- 💰 Prices/Money
- 📋 Subscriptions
- 🔗 Linking/Relationships

---

## 🛣️ **API ROUTE STANDARDS**

### **Route Setup Pattern**

```go
// ✅ STANDARD ROUTE SETUP
func Setup{Domain}Routes(router *gin.RouterGroup, db *database.DB) {
    group := router.Group("/{path}")
    
    // Apply middleware
    group.Use(middleware.AuthRequired())
    group.Use(middleware.AdminRequired())
    
    // Register routes
    group.GET("/", handler)
    group.POST("/", handler)
    group.GET("/:id", handler)
    group.PUT("/:id", handler)
    group.DELETE("/:id", handler)
}
```

### **Handler Pattern**

```go
// ✅ STANDARD HANDLER PATTERN
func {Action}Handler(db *database.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Parse input
        var req RequestType
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        
        // 2. Validate
        if err := validateRequest(req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        
        // 3. Execute business logic
        result, err := service.DoSomething(req)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        
        // 4. Return response
        c.JSON(http.StatusOK, gin.H{
            "status": "success",
            "data": result,
        })
    }
}
```

### **Response Format**

```go
// ✅ SUCCESS RESPONSE
c.JSON(http.StatusOK, gin.H{
    "status": "success",
    "data": result,
    "message": "Operation completed successfully",  // optional
})

// ✅ ERROR RESPONSE
c.JSON(http.StatusBadRequest, gin.H{
    "status": "error",
    "error": "Invalid input",
    "details": validationErrors,  // optional
})

// ✅ LIST RESPONSE WITH PAGINATION
c.JSON(http.StatusOK, gin.H{
    "status": "success",
    "data": items,
    "pagination": gin.H{
        "limit": 50,
        "offset": 0,
        "total": 1000,
    },
})
```

---

## 🎨 **FRONTEND STANDARDS**

### **Service Files**

```typescript
// ✅ STANDARD SERVICE PATTERN
// frontend/src/lib/services/{domain}-service.ts

export class {Domain}Service {
    static async get{Entity}(id: number): Promise<{Entity}> {
        const response = await apiRequest(`/{domain}/${id}`);
        if (response.error) {
            throw new Error(response.error);
        }
        return response.data;
    }
}
```

### **Store Files**

```typescript
// ✅ STANDARD STORE PATTERN
// frontend/src/lib/stores/{domain}-store.ts

import { writable } from 'svelte/store';

interface {Domain}Store {
    items: {Entity}[];
    loading: boolean;
    error: string | null;
}

export const {domain}Store = writable<{Domain}Store>({
    items: [],
    loading: false,
    error: null,
});
```

---

## ⚠️ **COMMON PITFALLS & SOLUTIONS**

### **1. Database Package Confusion**

**Problem**: Using wrong database package
```go
// ❌ WRONG
import "bome-backend/internal/database"
```

**Solution**: Always use infrastructure
```go
// ✅ CORRECT
import "bome-backend/infrastructure/database"
```

---

### **2. Missing Foreign Keys**

**Problem**: No referential integrity
```sql
-- ❌ WRONG
CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER  -- No FK constraint!
);
```

**Solution**: Always add FK constraints
```sql
-- ✅ CORRECT
CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE CASCADE
);
```

---

### **3. Array Field Performance**

**Problem**: Slow queries with array fields
```sql
-- ❌ SLOW
SELECT * FROM users u
WHERE 'cus_xxx' = ANY(u.stripe_customer_ids);  -- O(n*m) complexity
```

**Solution**: Use link tables
```sql
-- ✅ FAST
SELECT * FROM users u
INNER JOIN user_stripe_customers usc ON usc.user_id = u.id
WHERE usc.stripe_customer_id = 'cus_xxx';  -- O(1) with index
```

---

### **4. Service Duplication**

**Problem**: Multiple services doing the same thing
```
subscriber-service.ts
streaming-subscribers.ts
admin-subscriber-service.ts  -- All do similar things!
```

**Solution**: Create unified elastic service
```
subscriber-elastic-service.ts  -- Single source of truth
```

---

### **5. Naming Conflicts**

**Problem**: Function name already exists
```go
// ❌ ERROR: redeclared in this block
func getSyncStatus(c *gin.Context, db *database.DB) {
    // ...
}
```

**Solution**: Add version or domain suffix
```go
// ✅ CORRECT
func getSyncStatusV2(c *gin.Context, db *database.DB) {
    // ...
}

// OR
func getStripeSyncStatus(c *gin.Context, db *database.DB) {
    // ...
}
```

---

## 📝 **MIGRATION STANDARDS**

### **File Naming**

```
migrations/
├── 001_initial_schema.sql
├── 002_add_users_table.sql
├── 050_create_stripe_v2_schema.sql  # Major version bump
└── 051_add_indexes_to_stripe_v2.sql
```

**Rules**:
- 3-digit numbering: `001`, `002`, etc.
- Descriptive names: `add_`, `create_`, `fix_`
- Major versions: Jump to next milestone (050, 100, 150)

### **Migration Template**

```sql
-- ================================================================
-- Migration: {number}_{description}
-- ================================================================
-- Description: What this migration does
-- Date: YYYY-MM-DD
-- ================================================================

-- Your SQL here

-- ================================================================
-- Verification
-- ================================================================

-- Add verification queries (commented out)
-- SELECT COUNT(*) FROM new_table;
```

---

## 🎯 **DECISION MATRIX**

### **When to Create a New Service?**

| Scenario | Action |
|----------|--------|
| New domain/feature | ✅ Create new service |
| Existing service > 500 lines | ✅ Split into multiple services |
| Similar functionality exists | ❌ Extend existing service |
| Need v2 of existing service | ✅ Create `{Service}V2` |

### **When to Create a New Table?**

| Scenario | Action |
|----------|--------|
| New entity type | ✅ Create new table |
| Many-to-many relationship | ✅ Create link table |
| Current schema can't support new fields | ✅ Create `{table}_v2` |
| Just adding a field | ❌ Add column with migration |

---

## 🚀 **GETTING STARTED CHECKLIST**

When starting a new feature:

- [ ] Check this document for naming conventions
- [ ] Use `infrastructure/database` (not `internal/database`)
- [ ] Follow service naming: `{Domain}Service` or `{Domain}ElasticService`
- [ ] Add foreign keys to all new tables
- [ ] Index all foreign key columns
- [ ] Log with structured format and emojis
- [ ] Create migration file with proper numbering
- [ ] Add to BRAIDS documentation if new domain

---

## 📞 **QUESTIONS?**

If you're unsure about:
- Naming a service/table/function
- Which database package to use
- How to structure a new feature
- Migration numbering

**Refer to this document first!**

If still unclear, ask: "What's the standard way to {X} in the BOME codebase?"

---

**Last Updated**: October 26, 2025  
**Maintained By**: Development Team  
**Next Review**: Every major feature addition

