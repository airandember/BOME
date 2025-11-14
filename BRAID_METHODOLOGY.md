# BRAID Methodology: A Universal Guide for AI-Assisted Development

## 🧬 What is BRAID?

**BRAID** is a structured, modular approach to software development that organizes code into logical, maintainable strands. Like actual hair braids, each strand (feature/domain) is woven together with others while maintaining its own integrity and separation of concerns.

### Core Philosophy

- **Modularity**: Each feature/domain is a separate "strand" that can be developed, tested, and maintained independently
- **Layered Architecture**: Code is organized in clear layers (persistence, data access, business logic, application, presentation)
- **Documentation-First**: Every braid has documentation that defines its purpose, structure, and current status
- **Gradual Integration**: Strands are woven together incrementally, not all at once
- **AI-Friendly**: Clear structure and documentation make it easy for AI assistants to understand and extend the codebase

---

## 📚 Terminology

### BRAID
A complete feature domain or system component (e.g., Authentication, User Management, Subscription Billing)

### STRAND
An individual feature within a braid (e.g., User Login, Password Reset, Session Management)

### ELASTIC BAND
A layer within the architecture that connects components:
- **ELASTIC-BAND-UP**: Backend layers (Persistence → Data Access → Business Logic → Application)
- **ELASTIC-BAND-DOWN**: Frontend layers (Application → Presentation)

### PILOT
The initial implementation of a braid or strand, marking it as complete and production-ready

---

## 🏗️ Architecture Layers

### Backend (ELASTIC-BAND-UP)

```
┌─────────────────────────────────┐
│   APPLICATION LAYER             │  ← API Routes, Handlers, Middleware
├─────────────────────────────────┤
│   BUSINESS LOGIC LAYER          │  ← Services, Domain Logic, Validation
├─────────────────────────────────┤
│   DATA ACCESS LAYER             │  ← Repositories, Queries, Commands
├─────────────────────────────────┤
│   PERSISTENCE LAYER             │  ← Database Schema, Migrations, Models
└─────────────────────────────────┘
```

### Frontend (ELASTIC-BAND-DOWN)

```
┌─────────────────────────────────┐
│   PRESENTATION LAYER            │  ← UI Components, Pages, Styling
├─────────────────────────────────┤
│   APPLICATION LAYER             │  ← State Management, API Calls, Utils
└─────────────────────────────────┘
```

---

## 📁 Directory Structure

### Standard BRAID Structure

```
_backend/
└── braids/
    └── {braid-name}/                    # e.g., authentication
        ├── BRAID.md                     # Overview of entire braid
        ├── GETTING_STARTED.md           # Setup instructions
        ├── PILOT_COMPLETE.md            # Marks initial completion
        ├── COMPLETION_STATUS.md         # Current status
        │
        ├── layers/
        │   ├── persistence/
        │   │   ├── ELASTIC-BAND-UP.md
        │   │   └── schema/
        │   │       ├── {table-name}-table.md
        │   │       └── ...
        │   │
        │   ├── data-access/
        │   │   └── ELASTIC-BAND-UP.md
        │   │
        │   ├── business-logic/
        │   │   └── ELASTIC-BAND-UP.md
        │   │
        │   └── application/
        │       └── ELASTIC-BAND-UP.md
        │
        └── strands/
            ├── {strand-name}/           # e.g., user-login
            │   └── STRAND.md
            └── ...

_frontend/
└── braids/
    └── {braid-name}/
        ├── BRAID.md
        └── layers/
            ├── application/
            │   └── ELASTIC-BAND-DOWN.md
            └── presentation/
                └── ELASTIC-BAND-DOWN.md
```

---

## 📝 Documentation Templates

### BRAID.md Template

```markdown
# {BRAID NAME} Braid

## Purpose
{Brief description of what this braid handles}

## Scope
- Feature 1
- Feature 2
- Feature 3

## Dependencies
- External: {External services/APIs}
- Internal: {Other braids this depends on}

## Current Status
- [ ] Persistence Layer
- [ ] Data Access Layer
- [ ] Business Logic Layer
- [ ] Application Layer
- [ ] Frontend Integration
- [ ] Testing
- [ ] Documentation

## Related Strands
1. {strand-name} - {brief description}
2. {strand-name} - {brief description}

## Notes
{Important context, decisions, or gotchas}
```

### STRAND.md Template

```markdown
# {STRAND NAME} Strand

## Purpose
{What this specific feature does}

## Implementation Details

### Backend
- **Files**: {List of files}
- **Services**: {Service classes involved}
- **Routes**: {API endpoints}
- **Database**: {Tables used}

### Frontend
- **Pages**: {Routes/pages}
- **Components**: {UI components}
- **State**: {State management approach}

## Flow
1. {Step 1}
2. {Step 2}
3. {Step 3}

## Status
- [x] Backend Complete
- [x] Frontend Complete
- [x] Tested
- [ ] Edge Cases Handled

## Testing
{How to test this strand}

## Known Issues
{Any current bugs or limitations}
```

### ELASTIC-BAND-UP.md Template (Backend Layer)

```markdown
# {LAYER NAME} - Elastic Band Up

## Responsibility
{What this layer does}

## Files
- `path/to/file.go` - {Description}
- `path/to/service.go` - {Description}

## Key Functions/Methods
- `FunctionName()` - {What it does}
- `AnotherFunction()` - {What it does}

## Dependencies
- {What this layer depends on}

## Used By
- {What layers/components use this}

## Notes
{Important context}
```

---

## 🚀 How to Start a New BRAID

### Step 1: Define the Braid
```markdown
1. What domain/feature does it cover?
2. What are the main capabilities (strands)?
3. What external systems does it interact with?
4. What other braids does it depend on?
```

### Step 2: Create Directory Structure
```bash
mkdir -p _backend/braids/{braid-name}/{layers/{persistence/schema,data-access,business-logic,application},strands}
mkdir -p _frontend/braids/{braid-name}/layers/{application,presentation}
```

### Step 3: Write Initial Documentation
- Create `BRAID.md` with scope and purpose
- Create `GETTING_STARTED.md` with setup instructions
- List planned strands

### Step 4: Implement Bottom-Up (Backend)
1. **Persistence Layer**: Define database schema
2. **Data Access Layer**: Create repositories/queries
3. **Business Logic Layer**: Implement domain logic
4. **Application Layer**: Build API routes/handlers

### Step 5: Implement Top-Down (Frontend)
1. **Presentation Layer**: Build UI components
2. **Application Layer**: Integrate with backend APIs

### Step 6: Mark Complete
- Create `PILOT_COMPLETE.md` when fully functional
- Update `COMPLETION_STATUS.md` with current state

---

## 🧵 When to Create a New BRAID vs. STRAND

### Create a NEW BRAID when:
- ✅ It's a completely separate domain (e.g., Authentication vs. Billing)
- ✅ It has its own database tables
- ✅ It could theoretically be a microservice
- ✅ It has minimal overlap with existing braids

### Create a NEW STRAND when:
- ✅ It's a feature within an existing domain
- ✅ It shares database tables with other strands
- ✅ It's a logical extension of an existing braid
- ✅ It's a sub-feature or workflow

### Examples:
```
✅ BRAID: Authentication
   └── STRAND: User Login
   └── STRAND: Password Reset
   └── STRAND: Email Verification
   └── STRAND: Session Management

✅ BRAID: Subscription Billing
   └── STRAND: Plan Management
   └── STRAND: Payment Processing
   └── STRAND: Invoice Generation
   └── STRAND: Subscription Lifecycle
```

---

## 🤖 AI Assistant Instructions

When working with a BRAID-based codebase, follow these guidelines:

### 1. **Always Check for Existing Braids**
```markdown
Before implementing a new feature:
1. Check `_backend/braids/` and `_frontend/braids/` directories
2. Read relevant `BRAID.md` files
3. Check `COMPLETION_STATUS.md` for current state
4. Look for related strands
```

### 2. **Maintain Structure**
```markdown
When adding code:
- Place files in the correct layer
- Update relevant documentation
- Follow existing patterns in the braid
- Don't mix concerns between layers
```

### 3. **Document as You Go**
```markdown
When creating/modifying features:
- Update BRAID.md if scope changes
- Create/update STRAND.md for new features
- Update ELASTIC-BAND-*.md when touching layers
- Keep COMPLETION_STATUS.md current
```

### 4. **Ask Before Restructuring**
```markdown
Before making major changes:
- Explain the proposed structure
- Show how it fits with existing braids
- Discuss potential impacts
- Get user approval
```

### 5. **Code Organization Priorities**
```markdown
1. Separation of Concerns (layer independence)
2. Single Responsibility (one braid = one domain)
3. DRY (Don't Repeat Yourself) - but within reason
4. Documentation alongside code
5. Testability (easy to test each layer)
```

---

## 🎯 Benefits of BRAID Methodology

### For Developers
- ✅ Clear mental model of system architecture
- ✅ Easy to onboard new team members
- ✅ Reduced cognitive load (focus on one braid at a time)
- ✅ Clear ownership and responsibility

### For AI Assistants
- ✅ Structured context makes it easy to understand the system
- ✅ Documentation provides clear guidance
- ✅ Modularity prevents accidental breaking changes
- ✅ Easy to validate changes against architecture

### For the Codebase
- ✅ Maintainable and scalable
- ✅ Testable (each layer can be tested independently)
- ✅ Flexible (easy to refactor one braid without affecting others)
- ✅ Self-documenting (structure reflects architecture)

---

## 🔍 Real-World Example: Authentication Braid

### Structure
```
_backend/braids/authentication/
├── BRAID.md                                    # "Handles all user auth"
├── GETTING_STARTED.md                          # "Run migrations, set env vars"
├── PILOT_COMPLETE.md                           # "✅ Completed Nov 2024"
│
├── layers/
│   ├── persistence/
│   │   ├── ELASTIC-BAND-UP.md                  # "Tables: users, sessions, oauth2_*"
│   │   └── schema/
│   │       ├── users-table.md                  # "Stores user credentials"
│   │       ├── sessions-table.md               # "Active user sessions"
│   │       └── oauth2-tables.md                # "OAuth2 tokens"
│   │
│   ├── data-access/
│   │   └── ELASTIC-BAND-UP.md                  # "UserRepository, SessionRepository"
│   │
│   ├── business-logic/
│   │   └── ELASTIC-BAND-UP.md                  # "AuthService, HashingService"
│   │
│   └── application/
│       └── ELASTIC-BAND-UP.md                  # "POST /auth/login, /auth/logout"
│
└── strands/
    ├── user-registration/STRAND.md             # "Sign up flow"
    ├── user-login/STRAND.md                    # "Login with email/password"
    ├── password-reset/STRAND.md                # "Forgot password flow"
    └── session-management/STRAND.md            # "Token refresh, logout"
```

### Implementation Flow
1. **Persistence**: Created `users`, `sessions`, `oauth2_*` tables
2. **Data Access**: Built `UserRepository.FindByEmail()`, `SessionRepository.Create()`
3. **Business Logic**: Created `AuthService.Login()`, `AuthService.HashPassword()`
4. **Application**: Exposed `POST /auth/login`, `POST /auth/register` routes
5. **Frontend**: Built login page, registration form, session management

---

## 📋 Quick Reference Checklist

### Starting a New Project with BRAID
- [ ] Create `_backend/braids/` and `_frontend/braids/` directories
- [ ] Identify core domains (auth, billing, content, etc.)
- [ ] Create initial BRAID.md files for each domain
- [ ] Plan strands for MVP functionality
- [ ] Implement bottom-up (persistence → application)
- [ ] Document as you build

### Adding a Feature to Existing Braid
- [ ] Read the braid's `BRAID.md`
- [ ] Check if it's a new strand or enhancement
- [ ] Identify which layers need changes
- [ ] Update relevant ELASTIC-BAND-*.md files
- [ ] Create/update STRAND.md
- [ ] Update COMPLETION_STATUS.md

### Refactoring Code
- [ ] Identify which braid(s) are affected
- [ ] Check for dependencies between braids
- [ ] Update documentation to reflect changes
- [ ] Ensure layer boundaries are maintained
- [ ] Test each layer independently

---

## 🎓 Summary

The BRAID methodology is about **organized, documented, modular development**. It helps humans and AI work together effectively by providing:

1. **Clear Structure**: Know where everything goes
2. **Living Documentation**: Context lives alongside code
3. **Incremental Development**: Build and test piece by piece
4. **Easy Maintenance**: Change one thing without breaking everything
5. **AI Collaboration**: Clear patterns for AI assistants to follow

When in doubt, ask: "Which braid does this belong to?" and "Which layer handles this concern?"

---

**Remember**: The goal isn't perfect adherence to structure—it's **clarity, maintainability, and collaboration**. Use BRAID as a guide, not a prison. 🧬✨

//Prompts

//For New Projects:
//When starting a conversation with an AI assistant:"I'm using the BRAID methodology for this project. Please read BRAID_METHODOLOGY.md to understand the structure. 
//We're building [project description] and need to create the [braid name] braid for [feature description]."
//For Existing Projects:
//"This project uses BRAID methodology (see BRAID_METHODOLOGY.md).Check _backend/braids/ and _frontend/braids/ for existing code.I need to add [feature] to the [braid-name] 
//braid."
//For MCP Assistants:
//The document is self-contained and explains everything an AI needs to know to work effectively with your architecture!
//The guide emphasizes that BRAID is about clarity and collaboration, not rigid rules—perfect for flexible AI-assisted development! 🧬✨