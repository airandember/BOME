# 🎯 PHASE 1: PORTS & ADAPTERS (HEXAGONAL ARCHITECTURE)

**Status:** IN PROGRESS  
**Goal:** Decouple handlers from concrete service implementations  
**Timeline:** 2-3 hours  
**Completion:** 0%

---

## 📋 **OVERVIEW**

Phase 1 implements the **Ports & Adapters** pattern (Hexagonal Architecture) by:
1. Creating interface definitions (ports) for all shared services
2. Making handlers depend on interfaces instead of concrete implementations
3. Enabling better testing, mocking, and implementation swapping

This makes our architecture:
- ✅ **Testable** - Easy to mock services in tests
- ✅ **Flexible** - Swap implementations without changing handlers
- ✅ **Maintainable** - Clear contracts between layers
- ✅ **Scalable** - Add new implementations easily

---

## 🏗️ **ARCHITECTURE**

```
┌─────────────────────────────────────────────────────────────┐
│                    PRESENTATION LAYER                       │
│                   (Svelte Frontend)                         │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                   APPLICATION LAYER                         │
│              (Handlers - Thin Adapters)                     │
│                                                             │
│  ┌───────────────────────────────────────────────────┐     │
│  │  Depends on PORTS (Interfaces)                    │     │
│  │  • CryptoPort                                     │     │
│  │  • EmailPort                                      │     │
│  │  • StripePort                                     │     │
│  │  • BunnyPort                                      │     │
│  │  • AnalyticsPort                                  │     │
│  └───────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                  DOMAIN/USE CASES LAYER                     │
│          (Business Logic - Braid Specific)                  │
│                                                             │
│  Uses interfaces from ports, implements business rules     │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                 INFRASTRUCTURE LAYER                        │
│            (Concrete Implementations)                       │
│                                                             │
│  backend/services/                                          │
│  ├── crypto/    → implements CryptoPort                     │
│  ├── email/     → implements EmailPort                      │
│  ├── stripe/    → implements StripePort                     │
│  ├── bunny/     → implements BunnyPort                      │
│  └── analytics/ → implements AnalyticsPort                  │
└─────────────────────────────────────────────────────────────┘
```

---

## 📁 **DIRECTORY STRUCTURE**

```
backend/
├── ports/                          # NEW - Interface definitions
│   ├── crypto_port.go             # Crypto service interface
│   ├── email_port.go              # Email service interface
│   ├── stripe_port.go             # Stripe service interface
│   ├── bunny_port.go              # Bunny.net service interface
│   ├── analytics_port.go          # Analytics service interface
│   └── README.md                  # Ports documentation
│
├── services/                       # Concrete implementations
│   ├── crypto/
│   │   └── crypto.go              # implements ports.CryptoPort
│   ├── email/
│   │   └── email.go               # implements ports.EmailPort
│   ├── stripe/
│   │   └── stripe.go              # implements ports.StripePort
│   ├── bunny/
│   │   └── bunny.go               # implements ports.BunnyPort
│   └── analytics/
│       └── analytics.go           # implements ports.AnalyticsPort
│
└── [braids]/
    └── handlers/
        └── *.go                    # Depend on ports.* interfaces
```

---

## 🎯 **PHASE 1 TASKS**

### **Task 1: Create Ports Directory & Interfaces (30 mins)**
- [ ] Create `backend/ports/` directory
- [ ] Define `CryptoPort` interface
- [ ] Define `EmailPort` interface
- [ ] Define `StripePort` interface
- [ ] Define `BunnyPort` interface
- [ ] Define `AnalyticsPort` interface
- [ ] Create `ports/README.md` documentation

### **Task 2: Update Service Implementations (30 mins)**
- [ ] Add interface conformance check to `services/crypto/crypto.go`
- [ ] Add interface conformance check to `services/email/email.go`
- [ ] Add interface conformance check to `services/stripe/stripe.go`
- [ ] Add interface conformance check to `services/bunny/bunny.go`
- [ ] Add interface conformance check to `services/analytics/analytics.go`

### **Task 3: Update Handlers to Use Ports (45 mins)**
- [ ] Update `authentication/handlers/auth.go` to use ports
- [ ] Update `subscription/handlers/subscription.go` to use ports
- [ ] Update `video-streaming/handlers/` to use ports
- [ ] Update `admin/handlers/` to use ports
- [ ] Update `routing/setup.go` to inject interfaces

### **Task 4: Update Main Initialization (15 mins)**
- [ ] Update `main.go` to initialize services as ports
- [ ] Ensure dependency injection is clean
- [ ] Verify compilation

### **Task 5: Test & Validate (15 mins)**
- [ ] Compile entire project
- [ ] Run basic smoke tests
- [ ] Verify no regressions
- [ ] Document completion

---

## 💡 **EXAMPLE: CryptoPort Interface**

### **Before (Phase 0):**
```go
// Handler depends on concrete implementation
import "bome-backend/services/crypto"

func LoginHandler(db *database.DB, cryptoService *crypto.CryptoService) gin.HandlerFunc {
    return func(c *gin.Context) {
        clientIP := cryptoService.GetClientIP(...)
        // ...
    }
}
```

### **After (Phase 1):**
```go
// Handler depends on interface
import "bome-backend/ports"

func LoginHandler(db *database.DB, crypto ports.CryptoPort) gin.HandlerFunc {
    return func(c *gin.Context) {
        clientIP := crypto.GetClientIP(...)
        // ...
    }
}
```

### **Port Definition:**
```go
// backend/ports/crypto_port.go
package ports

type CryptoPort interface {
    // JWT Operations
    GenerateJWT(userID int, email, role string, verified bool) (string, error)
    ParseToken(tokenString string) (*Claims, error)
    ValidateTokenClaims(claims *Claims) error
    
    // Password Operations
    HashPassword(password string) (string, error)
    CheckPassword(hash, password string) error
    ValidatePassword(password string) error
    
    // Encryption
    EncryptString(plaintext string) (string, error)
    DecryptString(encrypted string) (string, error)
    
    // Utilities
    GetClientIP(remoteAddr, xForwardedFor, xRealIP string) string
    GenerateSecureToken() string
    SanitizeString(input string) string
    ValidateEmail(email string) error
}
```

### **Implementation Conformance:**
```go
// backend/services/crypto/crypto.go
package crypto

import "bome-backend/ports"

// Ensure CryptoService implements ports.CryptoPort
var _ ports.CryptoPort = (*CryptoService)(nil)

type CryptoService struct {
    // ... existing fields
}

// All methods already exist from Phase 0!
// No implementation changes needed!
```

---

## 🎯 **BENEFITS**

### **Before Phase 1:**
```go
// Tightly coupled to concrete implementation
handler := auth.LoginHandler(db, cryptoService)
```

### **After Phase 1:**
```go
// Flexible - can inject ANY implementation
handler := auth.LoginHandler(db, crypto)  // crypto is ports.CryptoPort

// Easy testing with mocks
mockCrypto := &MockCryptoService{}
testHandler := auth.LoginHandler(db, mockCrypto)

// Easy to swap implementations
productionCrypto := crypto.NewCryptoService()
stagingCrypto := crypto.NewFakeCryptoService()  // For staging/dev
```

---

## 📊 **SUCCESS METRICS**

- ✅ All handlers depend on ports, not concrete types
- ✅ All services implement their respective ports
- ✅ Zero compilation errors
- ✅ Dependency injection is clean and explicit
- ✅ Easy to add mock implementations for testing

---

## 🚀 **NEXT: PHASE 2**

After Phase 1, we'll tackle **Phase 2: Organize Shared Services by Domain**:
- Group services into domain-specific subdirectories
- `services/payment/stripe/`
- `services/media/bunny/`
- `services/security/crypto/`
- `services/communication/email/`

---

**Let's get started!** 🎉

