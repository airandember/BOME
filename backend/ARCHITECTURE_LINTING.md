# 🛡️ Architecture Linting Guide

**Purpose:** Automatically enforce clean architecture rules and prevent violations before they reach production.

---

## 📋 **What Is Architecture Linting?**

Architecture linting is like traditional code linting, but for **architectural rules**:
- ✅ Enforces module boundaries
- ✅ Prevents cross-braid dependencies
- ✅ Ensures layer separation
- ✅ Detects circular dependencies
- ✅ Validates naming conventions

**Think of it as:** A guard that prevents architectural technical debt from entering your codebase.

---

## 🎯 **The 11 Architecture Rules**

### **🧩 Braid Rules (BRAID-XXX)**

#### **BRAID-001: No Cross-Braid Imports in Handlers**
**Severity:** Error  
**Rule:** Handlers in one braid cannot import from another braid's packages.

**Bad:** ❌
```go
// authentication/handlers/auth.go
import "bome-backend/subscription/models" // ❌ Cross-braid import!
```

**Good:** ✅
```go
// authentication/handlers/auth.go
import "bome-backend/services/payment/stripe" // ✅ Shared service OK!
```

**Why:** Braids should be independent. Use shared services, not other braids.

---

#### **BRAID-002: Handlers Must Use Services Through Ports**
**Severity:** Error  
**Rule:** Handlers should depend on service interfaces (ports), not concrete implementations.

**Bad:** ❌
```go
func Handler(stripeService *stripe.StripeService) { } // ❌ Concrete type
```

**Good:** ✅
```go
func Handler(stripeService payment.StripePort) { } // ✅ Interface
```

**Why:** Dependency Inversion Principle - depend on abstractions.

---

#### **BRAID-003: Use Cases Cannot Import HTTP Packages**
**Severity:** Error  
**Rule:** Use cases are pure business logic and should not depend on HTTP/Gin.

**Bad:** ❌
```go
// authentication/usecases/register_user.go
import "github.com/gin-gonic/gin" // ❌ HTTP in use case!
```

**Good:** ✅
```go
// authentication/usecases/register_user.go
// No HTTP imports - pure business logic! ✅
```

**Why:** Use cases should be protocol-agnostic and testable without HTTP.

---

#### **BRAID-004: Models Cannot Import Services**
**Severity:** Error  
**Rule:** Models/repositories should not depend on business services.

**Bad:** ❌
```go
// authentication/models/user.go
import "bome-backend/services/email" // ❌ Model importing service
```

**Good:** ✅
```go
// authentication/models/user.go
import "bome-backend/infrastructure/database" // ✅ Infrastructure only
```

**Why:** Data layer should be independent of business logic layer.

---

### **🔌 Shared Services Rules (SHARED-XXX)**

#### **SHARED-001: Shared Services Must Be Domain-Organized**
**Severity:** Warning  
**Rule:** All shared services should be organized under their domain directory.

**Structure:**
```
services/
├── security/     ✅ Domain organized
├── payment/      ✅ Domain organized
├── media/        ✅ Domain organized
├── communication/✅ Domain organized
└── analytics/    ✅ Domain organized
```

**Why:** Discoverability and maintainability - services grouped by domain.

---

#### **SHARED-002: Services Must Have Port Interfaces**
**Severity:** Warning  
**Rule:** Each service domain should have a `ports.go` file defining interfaces.

**Structure:**
```
services/
└── payment/
    ├── stripe/
    │   └── stripe.go
    └── ports.go      ✅ Interface definitions
```

**Why:** Enables dependency injection and testability.

---

### **🏗️ Infrastructure Rules (INFRA-XXX)**

#### **INFRA-001: Infrastructure Only in Infrastructure Package**
**Severity:** Error  
**Rule:** Database, config, and infrastructure code should be in infrastructure package.

**Bad:** ❌
```go
// authentication/usecases/register_user.go
import "database/sql" // ❌ Raw database import
```

**Good:** ✅
```go
// authentication/usecases/register_user.go
import "bome-backend/infrastructure/database" // ✅ Abstracted
```

**Why:** Infrastructure concerns should be abstracted and centralized.

---

### **📚 Layer Rules (LAYER-XXX)**

#### **LAYER-001: Respect Layering Hierarchy**
**Severity:** Error  
**Rule:** Lower layers cannot import from higher layers.

**Hierarchy:**
```
handlers       (Top - Can import everything below)
   ↓
usecases       (Can import services, models, infrastructure)
   ↓
services       (Can import models, infrastructure)
   ↓
models         (Can import infrastructure only)
   ↓
infrastructure (Bottom - Imports nothing internal)
```

**Bad:** ❌
```go
// models/user.go
import "bome-backend/authentication/usecases" // ❌ Lower importing higher!
```

**Good:** ✅
```go
// models/user.go
import "bome-backend/infrastructure/database" // ✅ Same or lower layer
```

**Why:** Prevents circular dependencies and maintains clean architecture.

---

### **🧪 Test Rules (TEST-XXX)**

#### **TEST-001: Use Case Tests Should Be Pure**
**Severity:** Warning  
**Rule:** Use case tests should not require HTTP server or database.

**Bad:** ❌
```go
// usecases/register_user_test.go
import "net/http/httptest" // ❌ HTTP in use case test
```

**Good:** ✅
```go
// usecases/register_user_test.go
// Use mocks - no HTTP! ✅
mockDB := &MockDB{}
useCase := NewRegisterUser(mockDB, mockCrypto, mockEmail)
```

**Why:** Use case tests should be fast and isolated.

---

### **📝 Naming Rules (NAMING-XXX)**

#### **NAMING-001: Use Case Naming Convention**
**Severity:** Info  
**Rule:** Use cases should be named as verbs (snake_case).

**Examples:**
- ✅ `register_user.go`
- ✅ `login_user.go`
- ✅ `verify_email.go`
- ❌ `user_registration.go` (noun, not verb)
- ❌ `RegisterUser.go` (PascalCase, not snake_case)

**Why:** Consistent naming improves readability and discoverability.

---

### **🔄 Circular Dependency Rules (CIRCULAR-XXX)**

#### **CIRCULAR-001: No Circular Dependencies**
**Severity:** Error  
**Rule:** Packages cannot have circular import dependencies.

**Bad:** ❌
```
Package A imports Package B
Package B imports Package A
❌ CIRCULAR DEPENDENCY!
```

**Good:** ✅
```
Package A imports Package C
Package B imports Package C
✅ Both depend on shared package
```

**Why:** Circular dependencies indicate poor architecture and cause compilation issues.

---

## 🔧 **Using the Architecture Linter**

### **1. Build the Linter**
```bash
cd backend/tools/arch-lint
go build -o arch-lint main.go
```

### **2. Run Locally**
```bash
cd backend
./tools/arch-lint/arch-lint -verbose
```

### **3. Run with Output File**
```bash
cd backend
./tools/arch-lint/arch-lint \
  -rules=architecture-rules.json \
  -output=violations.json \
  -verbose
```

### **4. Fail on Warnings**
```bash
./tools/arch-lint/arch-lint -fail-on-warning
```

---

## 📊 **Sample Output**

```
╔══════════════════════════════════════════════════════════════╗
║              ARCHITECTURE LINT REPORT                        ║
╚══════════════════════════════════════════════════════════════╝

📊 Scanned: 142 files
🔍 Total Violations: 3
   ❌ Errors: 2
   ⚠️  Warnings: 1
   ℹ️  Info: 0

📋 Violations:
─────────────────────────────────────────────────────────────

❌ [BRAID-001] authentication/handlers/auth.go
   Handler imports from another braid: bome-backend/subscription/models
   💡 Braids should be independent. Handlers should only use shared services.

❌ [BRAID-003] authentication/usecases/register_user.go
   Use case imports HTTP package: github.com/gin-gonic/gin
   💡 Use cases should be protocol-agnostic and testable without HTTP.

⚠️  [SHARED-002] services/custom/
   Service domain missing ports.go file
   💡 Enables dependency injection and testability.

─────────────────────────────────────────────────────────────
```

---

## 🚀 **CI/CD Integration**

### **GitHub Actions** (Already configured!)
The linter runs automatically on:
- ✅ Pull requests to `main` or `develop`
- ✅ Pushes to `main`
- ✅ Any changes to `.go` files or `architecture-rules.json`

**What happens:**
1. Linter runs on every PR
2. If violations found:
   - ❌ CI fails
   - 📝 Comment added to PR with violations
   - 📊 Report uploaded as artifact
3. If clean:
   - ✅ CI passes
   - 🎉 PR can be merged

### **Pre-commit Hook** (Optional)
Add to `.git/hooks/pre-commit`:
```bash
#!/bin/bash
cd backend
./tools/arch-lint/arch-lint
if [ $? -ne 0 ]; then
  echo "❌ Architecture violations found! Commit aborted."
  exit 1
fi
```

Make executable:
```bash
chmod +x .git/hooks/pre-commit
```

---

## 📝 **Modifying Rules**

### **Add a New Rule**
Edit `backend/architecture-rules.json`:
```json
{
  "id": "CUSTOM-001",
  "name": "My custom rule",
  "severity": "error",
  "description": "Description of the rule",
  "pattern": {
    "from": "*/handlers/*.go",
    "cannot_import": ["some/package"]
  },
  "rationale": "Why this rule exists"
}
```

### **Disable a Rule**
Set severity to `"info"` or remove from rules array.

### **Add Exceptions**
```json
"exceptions": [
  {
    "rule_id": "BRAID-001",
    "file_pattern": "*/handlers/*_integration_test.go",
    "reason": "Integration tests may need cross-braid imports"
  }
]
```

---

## 🎯 **Benefits**

### **1. Early Detection** 🔍
- Catch violations before code review
- Prevent architectural debt
- Faster feedback loop

### **2. Consistent Enforcement** 📏
- No "rule bending" 
- Same standards for everyone
- Automated, not manual

### **3. Living Documentation** 📚
- Rules document architecture decisions
- New developers learn patterns quickly
- Architecture is self-enforcing

### **4. Confidence** 💪
- Know architecture is clean
- Refactor fearlessly
- Scale with confidence

---

## 🏆 **Success Metrics**

**After implementing architecture linting:**
- ✅ 0 cross-braid dependencies
- ✅ 0 circular dependencies
- ✅ 100% use case purity
- ✅ 100% layer separation compliance
- ✅ Architecture quality maintained automatically

---

## 📚 **Next Steps**

1. ✅ Run linter locally to establish baseline
2. ✅ Fix any existing violations
3. ✅ Enable pre-commit hook
4. ✅ Enable CI checks on PRs
5. ✅ Add custom rules as architecture evolves

---

**Built with ❤️ for bulletproof architecture!** 🛡️


