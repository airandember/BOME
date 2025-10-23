# 🏗️ **ARCHITECTURE DOCUMENTATION**

**Purpose:** Platform architecture, design patterns, and braid system documentation

---

## 📚 **DOCUMENTS IN THIS FOLDER**

### **⭐ START HERE**
- **`BOME_CONTEXT_STANDARD.md`** - Complete platform reference guide
  - Technology stack
  - Braid architecture
  - Backend standards (models, services, handlers)
  - Frontend standards (Svelte 5 patterns)
  - Database conventions
  - API design patterns
  - Real-time communication
  - Security standards
  - Testing strategy
  - Braid Combing methodology

### **Platform Overview**
- **`BOME_BRAIDS_SUMMARY.md`** - Overview of all 6 braids
  - Statistics (73+ files, 16,600+ lines, 170+ endpoints)
  - Integration matrix
  - Quality metrics
  - Deployment status

### **Individual Braid Documentation**
- **`braids/authentication-braid.md`** - (Reference: `backend/authentication/BRAID.md`)
- **`braids/creator-payout-braid.md`** - (Reference: `backend/creator/BRAID.md`)
- **`braids/subscription-braid.md`** - (Covered in migration docs)
- **`braids/video-streaming-braid.md`** - (Covered in migration docs)
- **`braids/youtube-braid.md`** - (Covered in migration docs)
- **`braids/admin-braid.md`** - (Covered in migration docs)

---

## 🎯 **WHEN TO USE**

### **Read `BOME_CONTEXT_STANDARD.md` when:**
- Starting a new session (always read first)
- Building a new feature
- Understanding architecture
- Setting up development environment
- Onboarding new developers

### **Read `BOME_BRAIDS_SUMMARY.md` when:**
- Getting platform overview
- Understanding feature domains
- Planning new braids
- Reviewing integration points

### **Read individual braid docs when:**
- Working on specific braid
- Understanding strand structure
- Reviewing API endpoints
- Planning feature additions

---

## 📊 **ARCHITECTURE HIERARCHY**

```
Platform (BOME)
├── Braid (Feature Domain)
│   ├── Strand (Individual Feature)
│   │   ├── Frontend UI (Svelte)
│   │   ├── Frontend Service (TypeScript)
│   │   ├── Backend Handler (Go)
│   │   ├── Backend Service (Go)
│   │   ├── Backend Model (Go)
│   │   └── Database (PostgreSQL)
│   ├── Elastic (Service Layer Interface)
│   └── Split-End (Error Boundary)
```

---

## 🔗 **RELATED DOCUMENTATION**

- **Database:** `../2-DATABASE/DATABASE_SCHEMA.md`
- **Testing:** `../3-TESTING/BRAID_COMBING_STANDARD.md`
- **Frontend:** `../4-FRONTEND/SVELTE5_REACTIVITY_GUIDE.md`

---

**Location:** `CONTEXT/1-ARCHITECTURE/`  
**Files:** 2 major docs + 6 braid references  
**Status:** Complete ✅

