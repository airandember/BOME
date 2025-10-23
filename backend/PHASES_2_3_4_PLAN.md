# 🏗️ PHASES 2-4 IMPLEMENTATION PLAN

**Status:** IN PROGRESS  
**Goal:** Complete architectural refinement with domain organization, use cases, and static analysis  
**Estimated Time:** 2-3 hours  

---

## 📋 **PHASE 2: ORGANIZE SHARED SERVICES BY DOMAIN** (45 mins)

### **Objective:**
Reorganize `backend/services/` into domain-specific subdirectories for better organization and discoverability.

### **Current Structure:**
```
services/
├── crypto/
│   ├── service.go
│   └── helpers.go
├── email/
│   └── email.go
├── stripe/
│   ├── stripe.go
│   └── stripe_sync.go
├── bunny/
│   ├── bunny.go
│   └── bunny_optimized.go
└── analytics/
    └── subscription_analytics.go
```

### **Target Structure:**
```
services/
├── security/              # Security domain
│   ├── crypto/
│   │   ├── service.go
│   │   └── helpers.go
│   └── ports.go          # Security port interfaces
├── payment/               # Payment domain
│   ├── stripe/
│   │   ├── service.go
│   │   └── sync.go
│   └── ports.go          # Payment port interfaces
├── media/                 # Media domain
│   ├── bunny/
│   │   ├── service.go
│   │   └── optimized.go
│   └── ports.go          # Media port interfaces
├── communication/         # Communication domain
│   ├── email/
│   │   └── service.go
│   └── ports.go          # Communication port interfaces
└── analytics/             # Analytics domain
    ├── subscription/
    │   └── service.go
    └── ports.go          # Analytics port interfaces
```

### **Implementation Steps:**
1. ✅ Create new domain directories
2. ⏳ Move services to domain folders
3. ⏳ Create domain-specific port files
4. ⏳ Update all imports across the codebase
5. ⏳ Test compilation

### **Benefits:**
- 🎯 Clear domain boundaries
- 📦 Better discoverability
- 🔍 Easier to find related services
- 📚 Domain-driven design

---

## 📋 **PHASE 3: EXTRACT USE CASES FROM HANDLERS** (60 mins)

### **Objective:**
Separate business logic from HTTP concerns by creating use case objects that handlers can call.

### **Current Structure:**
```go
// Handler with embedded business logic
func RegisterHandler(db *database.DB, emailService *email.EmailService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Parse HTTP request
        var req RegisterRequest
        c.ShouldBindJSON(&req)
        
        // Business logic mixed with HTTP
        if err := crypto.ValidateEmail(req.Email); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        
        // More business logic...
        user := &User{...}
        db.CreateUser(user)
        
        // HTTP response
        c.JSON(200, user)
    }
}
```

### **Target Structure:**
```go
// Pure use case (no HTTP)
package usecases

type RegisterUser struct {
    userRepo  ports.UserRepository
    emailSvc  ports.EmailPort
    cryptoSvc ports.CryptoPort
}

func (uc *RegisterUser) Execute(input RegisterInput) (*User, error) {
    // Pure business logic
    if err := uc.cryptoSvc.ValidateEmail(input.Email); err != nil {
        return nil, err
    }
    
    // Create user
    user := &User{...}
    if err := uc.userRepo.Create(user); err != nil {
        return nil, err
    }
    
    // Send email
    uc.emailSvc.SendVerificationEmail(...)
    
    return user, nil
}

// Thin handler (HTTP adapter)
func RegisterHandler(registerUC *usecases.RegisterUser) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req RegisterRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        
        // Call use case
        user, err := registerUC.Execute(usecases.RegisterInput{
            Email:     req.Email,
            FirstName: req.FirstName,
            LastName:  req.LastName,
        })
        
        if err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        
        c.JSON(200, user)
    }
}
```

### **Implementation Steps:**
1. ⏳ Create `authentication/usecases/` directory
2. ⏳ Extract `RegisterUser` use case
3. ⏳ Extract `LoginUser` use case
4. ⏳ Extract `VerifyEmail` use case
5. ⏳ Refactor handlers to use use cases
6. ⏳ Test compilation
7. ⏳ Repeat for subscription braid (as example)

### **Benefits:**
- ✅ **Testability**: Use cases can be unit tested without HTTP
- ✅ **Reusability**: Same use case can be called from HTTP, gRPC, CLI
- ✅ **Clarity**: Business logic is isolated and clear
- ✅ **Maintainability**: Changes to business logic don't affect HTTP layer

---

## 📋 **PHASE 4: ADD STATIC ANALYSIS (LINTING)** (30 mins)

### **Objective:**
Enforce architectural rules automatically with a custom linter that prevents violations.

### **Implementation:**

#### **1. Create architecture-rules.json**
```json
{
  "rules": [
    {
      "name": "No cross-braid imports in handlers",
      "from": "*/handlers/*",
      "cannot_import": [
        "authentication/handlers",
        "subscription/handlers",
        "video-streaming/handlers"
      ],
      "reason": "Handlers cannot import from other braids"
    },
    {
      "name": "Handlers must use ports for services",
      "from": "*/handlers/*",
      "must_import_through": "ports/*",
      "for": ["services/*"],
      "reason": "Handlers must depend on interfaces, not implementations"
    },
    {
      "name": "Use cases cannot import handlers",
      "from": "*/usecases/*",
      "cannot_import": ["*/handlers/*"],
      "reason": "Use cases are pure business logic"
    },
    {
      "name": "Services cannot import braids",
      "from": "services/*",
      "cannot_import": ["authentication/*", "subscription/*", "video-streaming/*"],
      "reason": "Shared services must be independent"
    }
  ]
}
```

#### **2. Create Simple Linter Tool**
```go
// tools/arch-lint/main.go
package main

import (
    "encoding/json"
    "go/parser"
    "go/token"
    "os"
    "path/filepath"
)

type Rule struct {
    Name         string   `json:"name"`
    From         string   `json:"from"`
    CannotImport []string `json:"cannot_import"`
    Reason       string   `json:"reason"`
}

func main() {
    // Parse rules
    // Walk Go files
    // Check imports
    // Report violations
}
```

#### **3. Add to CI**
```yaml
# .github/workflows/arch-lint.yml
name: Architecture Lint

on: [push, pull_request]

jobs:
  arch-lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: go run tools/arch-lint/main.go
```

### **Benefits:**
- 🛡️ **Prevents** architectural violations before merge
- 📚 **Documents** architectural decisions
- 🤖 **Automates** code review
- 🎯 **Enforces** module boundaries

---

## 🎯 **SUCCESS METRICS**

### **Phase 2 Complete:**
- ✅ All services organized by domain
- ✅ Port interfaces created per domain
- ✅ All imports updated
- ✅ Zero compilation errors

### **Phase 3 Complete:**
- ✅ Use cases extracted for auth braid
- ✅ Handlers are thin HTTP adapters
- ✅ Business logic is testable
- ✅ Clear separation of concerns

### **Phase 4 Complete:**
- ✅ Architecture rules documented
- ✅ Linter tool working
- ✅ CI integration complete
- ✅ No architectural violations

---

## ⏱️ **TIMELINE**

- **Phase 2:** 45 minutes
- **Phase 3:** 60 minutes  
- **Phase 4:** 30 minutes
- **Total:** 2 hours 15 minutes

**With breaks and testing:** ~3 hours total

---

**Let's build world-class architecture!** 🚀


