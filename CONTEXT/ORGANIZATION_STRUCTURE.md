# 📂 CONTEXT Folder Organization Structure

**Last Updated**: October 22, 2025  
**Purpose**: Quick reference for where to find and place documentation

---

## 📁 Folder Structure

```
CONTEXT/
├── 0-RECOVERY/              → Recovery logs, cleanup reports, session summaries
├── 1-ARCHITECTURE/          → System design, braids, naming conventions, standards
├── 2-DATABASE/              → Schema, migrations, RBAC, PostgreSQL docs
├── 3-TESTING/               → Testing standards, braid combing, checklists
├── 4-FRONTEND/              → Frontend guides, fixes, style guides, reactivity
├── 5-DEPLOYMENT/            → Deployment guides, Docker, Git workflow
├── 6-MIGRATIONS/            → Feature migrations organized by topic
│   ├── admin/
│   ├── braids/
│   ├── creator-payouts/
│   ├── stripe/
│   ├── subscriptions/
│   ├── videos/
│   └── youtube/
├── 7-PHASES/                → Phase completion reports, milestones, status updates
├── 8-STATUS/                → Project status, bug fixes, production readiness
├── 9-BRAIDS/                → Braid-specific documentation, split-end repairs
│   ├── admin/
│   ├── advertisement/
│   ├── analytics/
│   ├── authentication/
│   ├── communication/
│   ├── content/
│   ├── infrastructure/
│   ├── subscription/
│   ├── user-management/
│   └── video-streaming/
├── 10-FEATURES/             → Feature-specific guides and implementations
├── 11-GUIDES/               → User guides, test accounts, task files
└── scripts/                 → Automation scripts
    └── recovery/
```

---

## 📋 Documentation Placement Guide

### When creating NEW documentation, place it in:

#### Frontend Changes
- **Style guides, UI patterns** → `4-FRONTEND/`
- **Bug fixes, reactivity issues** → `4-FRONTEND/`
- **Component guides** → `4-FRONTEND/`

#### Backend Changes
- **Architecture decisions** → `1-ARCHITECTURE/`
- **Database changes** → `2-DATABASE/`
- **Braid-specific work** → `9-BRAIDS/[braid-name]/`

#### Testing & Quality
- **Testing standards** → `3-TESTING/`
- **Braid combing reports** → `3-TESTING/`
- **Bug fix summaries** → `8-STATUS/`

#### Features & Migrations
- **Migration reports** → `6-MIGRATIONS/[feature-name]/`
- **Feature completions** → `10-FEATURES/`
- **Phase completions** → `7-PHASES/`

#### Project Status
- **Milestone reports** → `7-PHASES/`
- **Production readiness** → `8-STATUS/`
- **Status updates** → `8-STATUS/`

---

## 🎯 Today's Documentation (October 22, 2025)

### Created & Organized:

#### Frontend Documentation
- ✅ `CONTEXT/4-FRONTEND/ADMIN_SIDEBAR_STYLE_GUIDE.md`
  - **Purpose**: User-approved sidebar styling standards
  - **Content**: Icon sizing, active states, glassmorphic effects, collapse behavior

- ✅ `CONTEXT/4-FRONTEND/VIDEO_FILTERS_CLEANUP.md`
  - **Purpose**: Documentation of statusFilter removal
  - **Content**: Redundancy explanation, cleanup actions, testing checklist

- ✅ `CONTEXT/4-FRONTEND/HMR_AUTH_FLICKER_FIX.md`
  - **Purpose**: Fix for Hot Module Reload authentication flicker
  - **Content**: Root cause analysis, solution implementation, testing results

#### Phase Documentation
- ✅ `CONTEXT/7-PHASES/PHASE_7C_ANALYTICS_PLAN.md`
  - **Purpose**: Complete plan for analytics implementation
  - **Content**: Implementation checklist, database queries, API formats

- ✅ `CONTEXT/7-PHASES/PHASE_7C_COMPLETE.md`
  - **Purpose**: Phase 7C completion report
  - **Content**: Implementation summary, testing plan, success criteria

---

## 📝 Naming Conventions

### File Naming Patterns:

#### Phase Completions
- Format: `PHASE_[X]_COMPLETE.md`
- Example: `PHASE_6_COMPLETE.md`, `PHASE_7C_COMPLETE.md`

#### Feature Migrations
- Format: `[FEATURE]_MIGRATION_COMPLETE.md`
- Example: `YOUTUBE_MIGRATION_COMPLETE.md`

#### Status Reports
- Format: `[TOPIC]_STATUS.md` or `[TOPIC]_REPORT.md`
- Example: `PRODUCTION_READINESS_REPORT.md`

#### Guides & Standards
- Format: `[TOPIC]_GUIDE.md` or `[TOPIC]_STANDARD.md`
- Example: `BRAID_COMBING_STANDARD.md`

#### Fixes & Solutions
- Format: `[PROBLEM]_FIX.md` or `[PROBLEM]_SOLUTION.md`
- Example: `HMR_AUTH_FLICKER_FIX.md`

---

## 🗂️ Quick Reference by Use Case

### "I need to find..."

| What | Where |
|------|-------|
| **Braid architecture** | `1-ARCHITECTURE/BRAIDS_INDEX.md` |
| **Database schema** | `2-DATABASE/DATABASE_SCHEMA.md` |
| **Testing standards** | `3-TESTING/BRAID_COMBING_STANDARD.md` |
| **Frontend style guide** | `4-FRONTEND/ADMIN_SIDEBAR_STYLE_GUIDE.md` |
| **Deployment steps** | `5-DEPLOYMENT/DEPLOYMENT_GUIDE.md` |
| **Phase status** | `7-PHASES/PHASE_[X]_COMPLETE.md` |
| **Production readiness** | `8-STATUS/PRODUCTION_READINESS_REPORT.md` |
| **Split-end repairs** | `9-BRAIDS/[braid]/split-ends/` |
| **Test accounts** | `11-GUIDES/TEST_ACCOUNTS.md` |

### "I need to document..."

| What | Place It In |
|------|-------------|
| **New frontend pattern** | `4-FRONTEND/[NAME]_GUIDE.md` |
| **Bug fix** | `8-STATUS/[BUG]_FIX.md` or `4-FRONTEND/[BUG]_FIX.md` |
| **Feature migration** | `6-MIGRATIONS/[feature]/[FEATURE]_COMPLETE.md` |
| **Phase completion** | `7-PHASES/PHASE_[X]_COMPLETE.md` |
| **Braid update** | `9-BRAIDS/[braid]/[UPDATE].md` |
| **Database change** | `2-DATABASE/[CHANGE].md` |
| **Testing result** | `3-TESTING/[TEST]_RESULTS.md` |

---

## ✅ Organization Checklist

When creating documentation:

- [ ] Choose correct top-level folder (0-11)
- [ ] Use consistent naming convention
- [ ] Include creation date in document
- [ ] Add clear purpose statement
- [ ] Update this structure doc if needed
- [ ] Remove any root-level duplicates
- [ ] Verify file is in `.gitignore` exceptions if needed

---

## 🎯 Folder Purpose Summary

| Folder | Files | Purpose |
|--------|-------|---------|
| **0-RECOVERY** | 9 | Recovery logs, cleanup reports |
| **1-ARCHITECTURE** | 10 | System design, braids, standards |
| **2-DATABASE** | 6 | Schema, migrations, RBAC |
| **3-TESTING** | 6 | Testing standards, combing |
| **4-FRONTEND** | 15 | Frontend guides, fixes |
| **5-DEPLOYMENT** | 6 | Deployment docs |
| **6-MIGRATIONS** | 52 | Feature migrations |
| **7-PHASES** | 13 | Phase completions |
| **8-STATUS** | 7 | Project status |
| **9-BRAIDS** | 28 | Braid-specific docs |
| **10-FEATURES** | 8 | Feature implementations |
| **11-GUIDES** | 3 | User guides |

**Total**: ~160+ organized documentation files

---

## 🚀 Best Practices

1. **One Topic Per File**: Each file should cover a single topic or feature
2. **Clear Titles**: Use descriptive, searchable file names
3. **Date Everything**: Include creation/update dates
4. **Cross-Reference**: Link to related docs
5. **Keep Updated**: Update docs when features change
6. **Avoid Duplication**: Check for existing docs first
7. **Use Templates**: Follow existing doc structures
8. **Be Specific**: Include code examples and screenshots
9. **Status Indicators**: Use ✅, ⏳, ❌ for clarity
10. **Index Updates**: Update indexes when adding new categories

---

**Maintained By**: AI Assistant & Development Team  
**Review Frequency**: Weekly  
**Last Audit**: October 22, 2025

