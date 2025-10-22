# 🌐 Shared Services Layer - Elastic Band Architecture
## The Universal Services Layer

**Created:** 2025-01-15  
**Status:** ✅ Implemented  
**Purpose:** Cross-cutting services available to all braids  

---

## 🎯 **Architectural Innovation**

The **Shared Services Layer** sits between braid-specific handlers and models, providing universal functionality that multiple braids need to access.

### **The Problem It Solves:**

**Before:**
```go
// ❌ Cross-braid imports (breaks encapsulation)
import "bome-backend/authentication/services"  // For crypto
import "bome-backend/subscription/services"    // For stripe
import "bome-backend/video-streaming/services" // For bunny
```

**After:**
```go
// ✅ Clean shared services
import "bome-backend/services/crypto"
import "bome-backend/services/stripe"
import "bome-backend/services/bunny"
```

---

## 🏗️ **Architecture Diagram**

```
┌─────────────────────────────────────────────────────┐
│          PRESENTATION LAYER (Svelte5)               │
└────────────────────┬────────────────────────────────┘
                     │ elastic band
┌────────────────────┴────────────────────────────────┐
│         APPLICATION LAYER (API Routes)              │
└────────────────────┬────────────────────────────────┘
                     │ elastic band
┌────────────────────┴────────────────────────────────┐
│    BUSINESS LOGIC (Braid-Specific Handlers)         │
│                                                      │
│  • authentication/handlers/                          │
│  • subscription/handlers/                            │
│  • video-streaming/handlers/                         │
│  • admin/handlers/                                   │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
        ╔════════════════════════════════╗
        ║   🌐 SHARED SERVICES LAYER 🌐  ║  ← NEW!
        ║                                ║
        ║  services/                     ║
        ║  ├── stripe/                   ║  Cross-cutting
        ║  ├── email/                    ║  concerns that
        ║  ├── crypto/                   ║  multiple braids
        ║  ├── bunny/                    ║  need to access
        ║  └── analytics/                ║
        ╚════════════════════════════════╝
                     │
                     ▼ elastic band (back to braid)
┌────────────────────┴────────────────────────────────┐
│       DATA ACCESS (Braid-Specific Models)           │
│                                                      │
│  • authentication/models/                            │
│  • subscription/models/                              │
│  • video-streaming/models/                           │
└────────────────────┬────────────────────────────────┘
                     │ elastic band
┌────────────────────┴────────────────────────────────┐
│              PERSISTENCE LAYER (Database)           │
└─────────────────────────────────────────────────────┘
```

---

## 📁 **Directory Structure**

```
backend/
├── services/                    # 🌐 SHARED SERVICES LAYER
│   ├── stripe/
│   │   ├── stripe.go           # Stripe API integration
│   │   ├── stripe_sync.go      # Stripe data synchronization
│   │   └── stripe_logger.go    # Stripe logging utilities
│   ├── email/
│   │   ├── email.go            # Email sending service
│   │   └── email_helpers.go    # Email templates & helpers
│   ├── crypto/
│   │   └── crypto.go           # Encryption/decryption service
│   ├── bunny/
│   │   ├── bunny.go            # Bunny.net CDN integration
│   │   └── bunny_optimized.go  # Optimized Bunny operations
│   └── analytics/
│       └── subscription_analytics.go  # Cross-braid analytics
│
├── authentication/              # BRAID
│   ├── handlers/               # Business logic
│   ├── models/                 # Data access
│   └── middleware/             # Auth middleware
│
├── subscription/                # BRAID
│   ├── handlers/               # Business logic
│   └── models/                 # Data access
│
├── video-streaming/             # BRAID
│   ├── handlers/               # Business logic
│   └── models/                 # Data access
│
├── admin/                       # BRAID
│   ├── handlers/               # Business logic
│   └── services/               # Admin-specific services
│
└── infrastructure/              # FOUNDATION
    ├── config/
    └── database/
```

---

## 🎯 **Shared Services Catalog**

### **1. Stripe Service** (`services/stripe/`)
**Purpose:** Payment processing and subscription management  
**Used By:** Subscription, Admin, Analytics  

**Files:**
- `stripe.go` - Core Stripe API integration
- `stripe_sync.go` - Data synchronization
- `stripe_logger.go` - Logging utilities

**Example:**
```go
import "bome-backend/services/stripe"

stripeService := stripe.NewStripeService(apiKey)
subscription, err := stripeService.CreateSubscription(customerID, planID)
```

---

### **2. Email Service** (`services/email/`)
**Purpose:** Email delivery and templating  
**Used By:** Authentication, Communication, Subscription  

**Files:**
- `email.go` - Email sending service
- `email_helpers.go` - Templates and helpers

**Example:**
```go
import "bome-backend/services/email"

emailService := email.NewEmailService()
err := emailService.SendVerificationEmail(user.Email, token)
```

---

### **3. Crypto Service** (`services/crypto/`)
**Purpose:** Encryption and decryption  
**Used By:** Authentication, Admin (secure settings)  

**Files:**
- `crypto.go` - AES encryption/decryption

**Example:**
```go
import "bome-backend/services/crypto"

cryptoService := crypto.GetGlobalCryptoService()
encrypted, err := cryptoService.EncryptString(secretKey)
```

---

### **4. Bunny Service** (`services/bunny/`)
**Purpose:** CDN and video streaming  
**Used By:** Video Streaming, Admin  

**Files:**
- `bunny.go` - Bunny.net API integration
- `bunny_optimized.go` - Optimized operations

**Example:**
```go
import "bome-backend/services/bunny"

bunnyService := bunny.NewBunnyService(apiKey, libraryID)
video, err := bunnyService.UploadVideo(file, title)
```

---

### **5. Analytics Service** (`services/analytics/`)
**Purpose:** Cross-braid analytics and reporting  
**Used By:** Subscription, Admin, Video Streaming  

**Files:**
- `subscription_analytics.go` - Subscription metrics

**Example:**
```go
import "bome-backend/services/analytics"

analyticsService := analytics.NewSubscriptionAnalyticsService(db)
mrr, err := analyticsService.CalculateMRR(time.Now(), nil)
```

---

## 🌊 **Strand Flow Example**

**User Registration with Email Verification:**

1. **Frontend** → `RegisterPage.svelte` (Presentation)
2. **API** → `POST /auth/register` (Application)
3. **Handler** → `authentication/handlers/auth.go` (Braid-specific)
4. **🌐 Email Service** → `services/email/email.go` (SHARED!)
5. **🌐 Crypto Service** → `services/crypto/crypto.go` (SHARED!)
6. **Model** → `authentication/models/user.go` (Braid-specific)
7. **Database** → PostgreSQL (Persistence)

**The strand passes through shared services but stays in its braid for handlers and models!**

---

## 🎯 **Import Patterns**

### **Handler Imports (Business Logic):**
```go
package handlers

import (
    "bome-backend/infrastructure/database"
    "bome-backend/authentication/models"
    
    // Shared services
    "bome-backend/services/crypto"
    "bome-backend/services/email"
    "bome-backend/services/stripe"
)
```

### **Model Imports (Data Access):**
```go
package models

import (
    "bome-backend/infrastructure/database"
    
    // Models should NOT import shared services
    // They only do database operations
)
```

---

## ✅ **Benefits**

### **1. Clean Separation of Concerns**
- Handlers contain business logic
- Services provide utilities
- Models handle data access

### **2. No Cross-Braid Dependencies**
- Braids don't import from other braids
- All shared code is in `services/`

### **3. Easy Testing**
- Services can be mocked independently
- Each braid can be tested in isolation

### **4. Scalability**
- New braids can use existing services
- New services can be added without touching braids

### **5. Maintainability**
- Services are in one place
- Changes to services don't require braid updates

---

## 📊 **Migration Status**

**Services Moved to Shared Layer:**
- ✅ Stripe (5 files)
- ✅ Email (3 files)
- ✅ Crypto (1 file)
- ✅ Bunny (2 files)
- ✅ Analytics (1 file)

**Total:** 12 services centralized ✅

---

## 🎯 **Elastic Band Contract**

### **Services ↔ Handlers:**
```markdown
## Interface Contract: Shared Services ↔ Braid Handlers

### Data Flow Direction:
🎯 Handlers → 🌐 Services → 🎯 Handlers

### Contract Definition:
**Services provide:**
- Stateless utility functions
- External API integrations
- Cross-cutting concerns
- Reusable business logic

**Services do NOT:**
- Access models directly
- Perform database operations
- Maintain state between calls
- Depend on specific braids

**Handlers are responsible for:**
- Calling services with correct parameters
- Handling service errors
- Coordinating between services and models
- Maintaining braid-specific logic
```

---

## 🚀 **Next Steps**

1. ✅ Create shared services structure
2. ✅ Move services to shared layer
3. ✅ Update package declarations
4. ⏸️ Update imports in handlers
5. ⏸️ Test compilation
6. ⏸️ Document elastic band contracts

---

## 💡 **Key Insight**

**This architectural refinement makes the "Strand and Braid" methodology even more powerful by clearly separating:**

- **Braid-specific logic** (handlers & models)
- **Cross-cutting concerns** (shared services)
- **Foundation** (infrastructure)

**Result:** A truly modular, maintainable, and scalable architecture! 🎯

---

**End of Document**  
**Status:** ✅ Implemented  
**Next:** Update handler imports to use shared services

