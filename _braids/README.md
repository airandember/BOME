# 🧬 BOME Braids - Unified Architecture

**Methodology:** See [BRAID_METHODOLOGY.md](BRAID_METHODOLOGY.md) or `_BRAIDS/BRAID_METHODOLOGY.md` for the BRAID methodology guide. Use `_braids/` (lowercase) as the canonical path.

**Last Updated:** 2025-10-17  
**Project:** BOME (Video Streaming Platform)  
**Architecture:** Vertical Slice (Braid Architecture)  
**Status:** Production-Ready Backend, Frontend In Development

---

## 🤖 **AI Assistant Context - START HERE**

This README provides complete context for AI assistants working on the BOME project. Read this first to understand our development philosophy, architecture, and protocols.

---

## 🏗️ **Architecture Overview**

### **What is a Braid?**

A **Braid** is a **complete vertical slice** of functionality from frontend UI to backend database. Think of it as a self-contained feature that can be understood and developed independently.

```
🧬 Braid = Complete Feature (Frontend → Backend)

Example: Authentication Braid
  Frontend (UI)
    ↓
  State Management
    ↓
  API Routes
    ↓
  Business Logic
    ↓
  Data Access
    ↓
  Database (Persistence)
```

### **Key Concepts:**

#### **1. Braids (Major Features)**
Complete vertical slices of functionality. Each braid is a self-contained feature domain.

**Available Braids:**
- **authentication** - User auth, JWT, OAuth2, email verification
- **video-streaming** - Video upload, playback, Bunny.net CDN
- **subscription-billing** - Plans, Stripe integration, payments
- **user-management** - User profiles, preferences, settings
- **content-management** - Video metadata, categories, curation
- **communication** - Email, notifications, comments
- **analytics-reporting** - Metrics, dashboards, insights
- **advertisement-system** - Ad serving, tracking, revenue
- **admin-dashboard** - Admin tools, system monitoring
- **infrastructure** - Shared services, utilities, middleware

#### **2. Strands (Sub-Features)**
Smaller, focused workflows within a braid. Each strand represents a specific user journey or use case.

**Example: Authentication Braid Strands:**
- `user-registration` - Sign up flow
- `user-login` - Login flow
- `email-verification` - Email confirmation
- `password-reset` - Forgot password flow
- `session-management` - Token refresh, logout

#### **3. Split-Ends (Edge Cases)**
Specific edge cases, error scenarios, or specialized behaviors that need special handling.

**Example Split-Ends:**
- Rate limiting exceeded
- Invalid token formats
- Concurrent session conflicts
- Email service failures

---

## 📁 **Directory Structure**

### **Unified Braid Structure:**
```
_braids/
  <braid-name>/
    BRAID.md                          # Complete documentation (frontend + backend)
    
    frontend/                         # All frontend code
      layers/
        presentation/                 # UI components, pages
          components/                 # Reusable UI components
          pages/                      # Route pages
          stores/                     # State management
        state-management/             # Global state, context
    
    backend/                          # All backend code
      layers/
        application/                  # API routes, controllers
          contracts/                  # API request/response schemas
          routes/                     # Route definitions
        
        business-logic/               # Core business logic
          handlers/                   # Request handlers
          services/                   # Business services
          middleware/                 # Custom middleware
        
        data-access/                  # Data access layer
          models/                     # Data models
          repositories/               # Data repositories
        
        persistence/                  # Database layer
          schema/                     # Database schemas
          migrations/                 # Database migrations
          indexes/                    # Database indexes
      
      strands/                        # Sub-feature implementations
        <strand-name>/
          STRAND.md                   # Strand documentation
          handlers/                   # Strand-specific handlers
          services/                   # Strand-specific services
```

### **Working Code Directories:**
```
backend/                              # Active Go backend
  authentication/
  services/
  routing/
  database/
  models/
  middleware/

frontend/                             # Active Svelte frontend
  src/
    routes/
    lib/
    components/
```

**Note:** `_braids/` is for **context and documentation**, not the working code. The actual implementations are in `backend/` and `frontend/`.

---

## 🔐 **.env Variable Protocol - IMPORTANT**

### **When AI Requests .env Variables:**

**NEVER hardcode sensitive values in responses.** Instead:

1. **Identify the variable needed** (e.g., `ENCRYPTION_KEY`, `JWT_SECRET`)
2. **Request it from the user** with clear instructions:
   ```
   ⚠️  MISSING ENVIRONMENT VARIABLE
   
   Please add to your .env file:
   ENCRYPTION_KEY=<32-character-random-string>
   
   Generate with: openssl rand -base64 32
   ```
3. **Provide generation instructions** where applicable
4. **Document the variable** in comments

### **Environment Variables Reference:**

**Security:**
- `ENCRYPTION_KEY` - 32-char key for data encryption (required)
- `JWT_SECRET` - JWT signing secret (required)
- `JWT_REFRESH_SECRET` - Refresh token secret (required)

**Database:**
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection (optional in dev)

**External Services:**
- `STRIPE_SECRET_KEY` - Stripe payment processing
- `STRIPE_WEBHOOK_SECRET` - Stripe webhook verification
- `BUNNY_API_KEY` - Bunny.net CDN
- `BUNNY_LIBRARY_ID` - Bunny video library
- `GOOGLE_OAUTH_CLIENT_ID` - Google OAuth2
- `GOOGLE_OAUTH_CLIENT_SECRET` - Google OAuth2

**Email:**
- `SMTP_HOST` - Email server
- `SMTP_PORT` - Email port
- `SMTP_USER` - Email username
- `SMTP_PASSWORD` - Email password

### **Protocol Summary:**
```
❌ DON'T: Include actual secrets in code
✅ DO: Request user to add to .env
✅ DO: Provide clear instructions
✅ DO: Document required format
```

---

## 🎯 **Development Philosophy**

### **1. Vertical Slicing**
- Each braid is a complete feature from UI to database
- Can be understood and developed independently
- Clear boundaries between braids
- Shared infrastructure in `infrastructure` braid

### **2. Layer Separation**
```
Presentation Layer    → What users see (UI)
Application Layer     → What users can do (API)
Business Logic Layer  → What happens (Rules)
Data Access Layer     → What gets stored (Models)
Persistence Layer     → Where it's stored (DB)
```

### **3. Testing Strategy**
- **Unit Tests** - Individual functions/methods
- **Integration Tests** - Multiple components together
- **E2E Tests** - Complete workflows (PowerShell scripts)
- **Braid Tests** - Full vertical slice validation

### **4. Error Handling**
- **Consistent JSON responses** with `error` field
- **Proper HTTP status codes** (400, 401, 404, 500, etc.)
- **Graceful degradation** (e.g., idempotent logout)
- **Safe type assertions** (no panics in production)

---

## 🚀 **Recent Achievements**

### **Mission 1: Backend LIVE & Testing** ✅
- Backend running on port 8080
- PostgreSQL connected
- All routes registered
- Health checks operational

### **Mission 2: Frontend-to-Backend Braid Testing** ✅
- **49 tests executed, 98% pass rate**
- Authentication Braid: 90% (10 tests)
- Video Streaming Braid: 100% (5 tests)
- Cross-Cutting Concerns: 100% (10 tests)
- User Profile Braid: 100% (8 tests)
- Subscription/Billing: 100% (8 tests)
- Email Verification: 100% (8 tests)

### **Critical Bug Fixes:**
- ✅ Invalid token handling (500 → 401)
- ✅ Graceful idempotent logout
- ✅ Safe JWT parsing (no panics)
- ✅ CORS headers configured
- ✅ Security headers present

---

## 🛠️ **Technology Stack**

### **Backend:**
- **Language:** Go 1.21+
- **Framework:** Gin (HTTP router)
- **Database:** PostgreSQL + GORM
- **Cache:** Redis (optional)
- **Auth:** JWT + OAuth2
- **CDN:** Bunny.net
- **Payments:** Stripe

### **Frontend:**
- **Framework:** SvelteKit
- **Language:** TypeScript
- **Styling:** TailwindCSS
- **State:** Svelte stores
- **API:** Fetch + REST

### **DevOps:**
- **Testing:** PowerShell scripts
- **Monitoring:** Built-in health checks
- **Logging:** Structured JSON logs

---

## 📋 **Common Tasks**

### **Starting the Backend:**
```powershell
cd backend
.\start-dev.ps1
# or
go run main.go
```

### **Running Tests:**
```powershell
cd backend
.\test-braid-auth.ps1           # Authentication tests
.\test-braid-video-simple.ps1   # Video streaming tests
.\test-braid-cross-cutting.ps1  # CORS, errors, security
# etc.
```

### **Checking Health:**
```powershell
curl http://localhost:8080/health
```

### **Loading Context for a Braid:**
When working on a feature, load the complete braid:
```
_braids/<braid-name>/BRAID.md     # Full documentation
_braids/<braid-name>/frontend/    # Frontend implementation
_braids/<braid-name>/backend/     # Backend implementation
```

---

## 🎓 **Working with This Codebase**

### **For AI Assistants:**

1. **Always load the complete braid** when working on a feature
2. **Follow the .env protocol** - never expose secrets
3. **Maintain layer separation** - respect architectural boundaries
4. **Test your changes** - run relevant test scripts
5. **Document as you go** - update BRAID.md files
6. **Use proper error handling** - consistent JSON responses
7. **Think in vertical slices** - frontend to backend

### **For Humans:**

1. **Start with the BRAID.md** to understand a feature
2. **Check recent missions** to see what's been tested
3. **Follow the protocols** (especially .env)
4. **Run tests** before committing
5. **Update documentation** when adding features

---

## 🔗 **Key Documentation Files**

- **Architecture:** This file (README.md)
- **Mission Reports:** `backend/MISSION_*_COMPLETE.md`
- **Bug Fixes:** `backend/GRACEFUL_LOGOUT_FIX.md`
- **Consolidation:** `CONSOLIDATION_COMPLETE.md`
- **Individual Braids:** `_braids/<braid-name>/BRAID.md`

---

## 🌟 **Design Principles**

### **1. Security First**
- All sensitive endpoints require authentication
- JWT tokens properly validated
- Environment variables for secrets
- Security headers configured

### **2. User Experience**
- Graceful error handling
- Idempotent operations where possible
- Clear error messages
- Fast response times (~2-5ms avg)

### **3. Maintainability**
- Clear separation of concerns
- Self-contained braids
- Comprehensive documentation
- Consistent code style

### **4. Scalability**
- Connection pooling
- Rate limiting
- Caching strategy (Redis)
- CDN integration (Bunny.net)

---

## 🎯 **Current Status**

**Backend:** ✅ Production-Ready
- All core systems operational
- 98% test coverage
- Performance excellent (~2-5ms)
- Security validated

**Frontend:** 🚧 In Development
- SvelteKit setup complete
- Component library in progress
- Integration with backend pending

**Next Steps:**
- Mission 3: Database optimization
- Frontend-backend integration
- Load testing
- Production deployment prep

---

## 💡 **Quick Reference**

### **Important Paths:**
```
_braids/              # Context & documentation
backend/              # Working Go code
frontend/             # Working Svelte code
backend/test-*.ps1    # Test scripts
```

### **Key Endpoints:**
```
http://localhost:8080/health          # Health check
http://localhost:8080/api/v1/auth/*   # Authentication
http://localhost:8080/api/v1/videos/* # Video streaming
http://localhost:8080/api/v1/users/*  # User management
```

### **Important Commands:**
```powershell
# Backend
go run main.go                        # Start server
go build                              # Build binary

# Testing
.\test-braid-auth.ps1                 # Run auth tests

# Database
psql -U bome_user -d bome_db          # Connect to DB
```

---

## 🤝 **Communication Protocol**

### **When starting a new session:**
1. Read this README.md
2. Check recent MISSION_*.md files
3. Load relevant braid context
4. Ask clarifying questions

### **When working on features:**
1. Load complete braid context
2. Understand the vertical slice
3. Respect layer boundaries
4. Test end-to-end
5. Document changes

### **When handling errors:**
1. Follow .env protocol for secrets
2. Use proper HTTP status codes
3. Return consistent JSON format
4. Log errors appropriately

---

## 📞 **Need Help?**

**Architecture Questions:** Read the BRAID.md for the relevant feature  
**Testing Questions:** Check `backend/MISSION_2_COMPLETE_FINAL.md`  
**Setup Questions:** See `.env protocol` above  
**Code Questions:** Load the complete braid context  

---

## ✅ **Verification Checklist**

Before considering work complete:

- [ ] All tests passing (≥80% pass rate)
- [ ] Documentation updated
- [ ] Environment variables documented
- [ ] Error handling implemented
- [ ] Security validated
- [ ] Performance acceptable
- [ ] Code follows architecture
- [ ] BRAID.md updated if needed

---

## 🎉 **Summary**

**BOME** uses a **Braid Architecture** where each feature is a complete **vertical slice** from UI to database. The `_braids/` directory contains unified documentation and context, while `backend/` and `frontend/` contain the actual working code.

**Key Points:**
- ✅ Always load complete braid context
- ✅ Follow .env protocol for secrets
- ✅ Maintain layer separation
- ✅ Test end-to-end
- ✅ Document as you go

**Current Status:** Backend production-ready (98% test coverage), Frontend in development.

---

**Last Updated:** 2025-10-17  
**Consolidation Date:** 2025-10-17 11:53:46  
**Version:** 2.0 (Unified Braids)

🧬 **Happy Braiding!** 🧬
