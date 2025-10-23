# 🔌 Ports (Interfaces)

This directory contains **port interfaces** that define contracts between our application layer and infrastructure layer, following the **Hexagonal Architecture** (Ports & Adapters) pattern.

---

## 📋 **What are Ports?**

**Ports** are interfaces that define *what* operations are available, without specifying *how* they're implemented. They act as contracts between layers:

- **Application Layer** (handlers) depends on **Ports** (interfaces)
- **Infrastructure Layer** (services) implements **Ports**

This decouples business logic from infrastructure concerns, making code:
- ✅ **Testable** - Easy to mock
- ✅ **Flexible** - Swap implementations
- ✅ **Maintainable** - Clear contracts

---

## 🗂️ **Available Ports**

### **CryptoPort** (`crypto_port.go`)
Cryptographic operations, JWT tokens, password hashing, encryption.

**Key Methods:**
- `GenerateJWT()`, `ParseToken()`, `GenerateTokenPair()`
- `HashPassword()`, `CheckPassword()`, `ValidatePassword()`
- `EncryptString()`, `DecryptString()`
- `GetClientIP()`, `GenerateSecureToken()`

**Implementation:** `backend/services/crypto/`

---

### **EmailPort** (`email_port.go`)
Email sending operations.

**Key Methods:**
- `SendVerificationEmail()`, `SendPasswordResetEmail()`
- `SendWelcomeEmail()`, `SendPasswordSetupEmail()`
- `SendTemplatedEmail()`

**Implementation:** `backend/services/email/`

---

### **StripePort** (`stripe_port.go`)
Stripe payment processing.

**Key Methods:**
- Customer: `CreateCustomer()`, `GetCustomer()`, `UpdateCustomer()`
- Subscription: `CreateSubscription()`, `CancelSubscription()`
- Payment: `AttachPaymentMethod()`, `CreateRefund()`
- Products: `ListProducts()`, `ListPrices()`

**Implementation:** `backend/services/stripe/`

---

### **BunnyPort** (`bunny_port.go`)
Bunny.net CDN and video streaming.

**Key Methods:**
- `GetVideoPlayData()`, `UploadVideo()`, `DeleteVideo()`
- `GetVideoInfo()`, `UpdateVideoMetadata()`
- `GetStreamURL()`, `GetEmbedURL()`

**Implementation:** `backend/services/bunny/`

---

### **AnalyticsPort** (`analytics_port.go`)
Subscription analytics and tracking.

**Key Methods:**
- `GetActiveSubscriptionsCount()`, `GetRevenueMetrics()`
- `CalculateMRR()`, `CalculateChurnRate()`
- `TrackSubscriptionEvent()`
- `GenerateSubscriptionReport()`

**Implementation:** `backend/services/analytics/`

---

## 💡 **Usage Example**

### **Before (Direct Dependency):**
```go
import "bome-backend/services/crypto"

func LoginHandler(cryptoService *crypto.CryptoService) gin.HandlerFunc {
    // Tightly coupled to concrete implementation
    token := cryptoService.GenerateJWT(...)
}
```

### **After (Port Dependency):**
```go
import "bome-backend/ports"

func LoginHandler(crypto ports.CryptoPort) gin.HandlerFunc {
    // Depends on interface - flexible!
    token := crypto.GenerateJWT(...)
}
```

### **Benefits:**
```go
// Production
realCrypto := crypto.NewCryptoService()
handler := LoginHandler(realCrypto)

// Testing
mockCrypto := &MockCryptoService{}
testHandler := LoginHandler(mockCrypto)

// Staging/Dev
fakeCrypto := &FakeCryptoService{}
devHandler := LoginHandler(fakeCrypto)
```

---

## 🏗️ **Adding a New Port**

1. **Create the interface file:**
   ```go
   // ports/myservice_port.go
   package ports
   
   type MyServicePort interface {
       DoSomething() error
   }
   ```

2. **Implement in infrastructure:**
   ```go
   // services/myservice/myservice.go
   package myservice
   
   import "bome-backend/ports"
   
   // Ensure conformance
   var _ ports.MyServicePort = (*MyService)(nil)
   
   type MyService struct {}
   
   func (s *MyService) DoSomething() error {
       // Implementation
   }
   ```

3. **Use in handlers:**
   ```go
   import "bome-backend/ports"
   
   func MyHandler(service ports.MyServicePort) gin.HandlerFunc {
       return func(c *gin.Context) {
           service.DoSomething()
       }
   }
   ```

---

## 🎯 **Architecture Diagram**

```
┌─────────────────────────────────────────────┐
│         Application Layer (Handlers)        │
│                                             │
│    Depends on Ports (Interfaces)            │
│    • CryptoPort                             │
│    • EmailPort                              │
│    • StripePort                             │
└─────────────────────────────────────────────┘
                    ↓ (depends on)
┌─────────────────────────────────────────────┐
│          Ports (This Directory)             │
│                                             │
│    Interface Definitions                    │
└─────────────────────────────────────────────┘
                    ↑ (implemented by)
┌─────────────────────────────────────────────┐
│     Infrastructure Layer (Services)         │
│                                             │
│    Concrete Implementations                 │
│    • services/crypto/                       │
│    • services/email/                        │
│    • services/stripe/                       │
└─────────────────────────────────────────────┘
```

---

## ✅ **Best Practices**

1. **Keep interfaces small** - Single Responsibility Principle
2. **Define clear contracts** - Document expected behavior
3. **Use meaningful names** - Port names should end with "Port"
4. **Group related methods** - Logical cohesion
5. **Version interfaces** - Add new methods, don't break existing ones
6. **Test interface compliance** - `var _ Port = (*Implementation)(nil)`

---

**Created:** Phase 1 of Architectural Refinement  
**Last Updated:** [Date]  
**Status:** ✅ Complete

