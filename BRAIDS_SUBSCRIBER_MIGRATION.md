# BRAIDS Subscriber Service Migration Documentation

## 🎯 MIGRATION GOAL
Consolidate fragmented subscriber/user services into unified **SubscriberElasticService** following **Single Responsibility Principle** and **BRAIDS Architecture**.

## 📊 CURRENT STATE ASSESSMENT

### 🔴 FRAGMENTED SERVICES (TO BE MIGRATED)

#### 1. **subscriber-cache.ts** 
- **Location**: `frontend/src/lib/cache/subscriber-cache.ts`
- **Function**: Caching layer for subscriber data
- **API Calls**: `/admin/subscribers/enhanced` (OLD FRAGMENTED)
- **Status**: ⚠️ PARTIALLY MIGRATED (updated to use elastic service)
- **Used By**: Multiple streaming components
- **Migration Status**: 🔄 IN PROGRESS

#### 2. **streaming-subscribers.ts**
- **Location**: `frontend/src/lib/services/streaming-subscribers.ts` 
- **Function**: Streaming-specific subscriber operations
- **API Calls**: `/admin/subscribers/enhanced` (OLD FRAGMENTED) → ✅ MIGRATED TO ELASTIC SERVICE
- **Status**: ✅ MIGRATED
- **Used By**: Customer sync panels, streaming layout
- **Migration Status**: ✅ COMPLETE

#### 3. **subscribers-store.ts**
- **Location**: `frontend/src/lib/stores/subscribers-store.ts`
- **Function**: Svelte store for subscriber state management
- **API Calls**: Uses `subscriberElasticService` ✅
- **Status**: ✅ MIGRATED
- **Used By**: Enhanced Subscribers Page
- **Migration Status**: ✅ COMPLETE

### 🟢 UNIFIED SERVICES (TARGET STATE)

#### 1. **SubscriberElasticService** ✅
- **Location**: `frontend/src/lib/services/subscriber-elastic-service.ts`
- **Function**: Single source of truth for subscriber data
- **API Calls**: `/admin/subscriber-elastic/*` (UNIFIED)
- **Status**: ✅ IMPLEMENTED
- **Features**: 
  - `getAllSubscribers()`
  - `getSubscriberByEmail()`
  - `getSubscriberStats()`
  - `getDiagnosticData()`
  - `updateManualVideoAccess()`

#### 2. **UserService** ✅
- **Location**: `frontend/src/lib/services/user-service.ts`
- **Function**: Unified user management (admin + self-service)
- **API Calls**: `/admin/users/*` (admin) + `/users/me` (self-service)
- **Status**: ✅ IMPLEMENTED
- **Features**:
  - **Self-Service**: `getCurrentUser()`, `updateCurrentUser()`
  - **Admin Operations**: `getAllUsers()`, `createUser()`, `updateUser()`, `deleteUser()`
  - **Bulk Operations**: `bulkCreateUsers()`
  - **Analytics**: `getUserStats()`, `getAvailableRoles()`

#### 3. **Backend SubscriberElasticService** ✅
- **Location**: `backend/internal/services/subscriber_elastic_service.go`
- **Function**: Unified backend service with proper DB joins
- **Endpoints**: `/api/v1/admin/subscriber-elastic/*`
- **Status**: ✅ IMPLEMENTED
- **Features**: CTE queries, unified data model

## 🗺️ COMPONENT MAPPING

### STRANDS USING SUBSCRIBER SERVICES

#### **Admin Dashboard Strand**
- **Components**: 
  - `admin/dashboard/+page.svelte` → Uses `subscriberCache`
  - `admin/users/+page.svelte` → Uses `/admin/users` API
- **Migration**: Move to `SubscriberElasticService`

#### **Video Streaming Strand** 
- **Components**:
  - `admin/streaming/subscribers/+page.svelte` → Uses `subscribers-store` ✅
  - `admin/streaming/subscribers/customers/+page.svelte` → Uses `StreamingSubscriberService` ❌
  - `admin/streaming/subscribers/diagnostic/+page.svelte` → Uses `subscriberElasticService` ✅
- **Migration**: Complete streaming-subscribers.ts migration

#### **Authentication Strand**
- **Components**:
  - `auth/*` pages → Use `/auth/*` and `/users/me/*` APIs ✅
- **Migration**: Already using proper RBAC endpoints

#### **Subscription Strand**
- **Components**:
  - `subscription/+page.svelte` → Uses `publicPlansService` ✅
- **Migration**: Already using proper service

## 📋 MIGRATION TASKS

### PHASE 1: Complete Cache Service Migration ✅
- [x] Update `subscriber-cache.ts` to use elastic service
- [x] Add filtering and pagination logic
- [x] Add KPI generation from unified data

### PHASE 2: Migrate Streaming Service ✅
- [x] Update `streaming-subscribers.ts` to use elastic service
- [x] Replace fragmented API calls with unified calls
- [x] Add filtering and pagination logic
- [x] Update `getSubscriberStats()` method

### PHASE 3: Consolidate User Services ✅
- [x] Audit `/admin/users` endpoint usage
- [x] Create unified user profile service
- [x] Implement RBAC-protected user operations

### PHASE 4: Cleanup & Validation ✅
- [x] Test all migrated components
- [x] Verify data consistency across all strands
- [x] Fix backend NULL scanning issues
- [ ] Request permission to delete old services

## 🔍 API ENDPOINT AUDIT

### CURRENT ENDPOINTS IN USE

#### **Admin Endpoints** (RBAC Protected)
- `/api/v1/admin/subscribers/enhanced` → ❌ OLD FRAGMENTED (MIGRATED)
- `/api/v1/admin/subscriber-elastic/subscribers` → ✅ UNIFIED
- `/api/v1/admin/users` → ✅ AUDITED - PROPER RBAC
- `/api/v1/admin/users/stats` → ✅ AUDITED - PROPER RBAC
- `/api/v1/admin/users/roles` → ✅ AUDITED - PROPER RBAC
- `/api/v1/admin/subscriber-elastic/stats` → ✅ UNIFIED
- `/api/v1/admin/subscriber-elastic/diagnostics` → ✅ UNIFIED

#### **User Endpoints** (Self-Service)
- `/api/v1/users/me` → ✅ PROPER RBAC
- `/api/v1/users/profile` → ✅ PROPER RBAC (alias for /me)
- `/api/v1/auth/*` → ✅ PROPER RBAC

#### **Public Endpoints**
- `/api/v1/public/plans` → ✅ PROPER SERVICE

### 📊 AUDIT FINDINGS

#### **Frontend Usage of `/admin/users`:**
1. **`admin/streaming/subscribers/customers/+page.svelte`**:
   - `GET /admin/users?limit=3000` → Load local users for Stripe comparison
   - `POST /admin/users` → Create individual users
   - `POST /admin/users/bulk` → Bulk create users
   - **Status**: ✅ PROPER RBAC (admin required)

2. **`admin/users/+page.svelte`**:
   - `GET /admin/users/stats` → User statistics
   - `GET /admin/users?{filters}` → Paginated user list
   - `PUT /admin/users/{id}` → Update user role
   - `POST /admin/users` → Create new user
   - **Status**: ✅ PROPER RBAC (admin required)

3. **`admin/+layout.svelte`** & **`admin/dashboard/+page.svelte`**:
   - Navigation links to `/admin/users`
   - **Status**: ✅ PROPER RBAC (permission checks)

#### **Backend Implementation:**
- **Admin Routes**: Properly protected with `middleware.AdminRequired()`
- **User Profile Routes**: Properly protected with `middleware.AuthRequired()`
- **RBAC**: ✅ CORRECTLY IMPLEMENTED

### 🧪 TESTING RESULTS

#### **Frontend Testing (2025-10-25):**
- ✅ **679 subscribers loaded** using UNIFIED ELASTIC SERVICE
- ✅ **All services working**: subscriber-cache.ts, streaming-subscribers.ts, subscribers-store.ts
- ✅ **Data consistency achieved**: Frontend tables display correctly
- ✅ **No more data fragmentation**: Single source of truth working

#### **Backend Testing:**
- ✅ **Elastic service responding**: 679 unified subscribers retrieved
- ✅ **NULL scanning fixed**: plan_type and plan_status now handle NULL values properly
- ✅ **Database joins working**: CTE queries returning complete subscriber data

## 🎯 TARGET ARCHITECTURE

```
┌─────────────────────────────────────────────────────────────┐
│                    BRAIDS ARCHITECTURE                      │
├─────────────────────────────────────────────────────────────┤
│  STRANDS (Frontend)                                         │
│  ├── Admin Dashboard Strand → SubscriberElasticService     │
│  ├── Video Streaming Strand → SubscriberElasticService     │
│  ├── Authentication Strand → AuthService + UserService     │
│  └── Subscription Strand → PublicPlansService             │
├─────────────────────────────────────────────────────────────┤
│  ELASTIC SERVICES (Backend)                                │
│  ├── SubscriberElasticService ← Single source of truth     │
│  ├── AuthService (existing)                               │
│  └── PublicPlansService (existing)                        │
├─────────────────────────────────────────────────────────────┤
│  RBAC MIDDLEWARE                                            │
│  ├── Admin operations: /admin/subscriber-elastic/*         │
│  ├── User operations: /users/me/*                         │
│  └── Public operations: /public/*                         │
└─────────────────────────────────────────────────────────────┘
```

## 📝 NOTES FOR IMPLEMENTATION

### Service Consolidation Strategy
1. **Keep old services** until migration is complete and tested
2. **Update components** to use unified service gradually
3. **Test thoroughly** before requesting deletion permission
4. **Document all changes** in this file

### RBAC Implementation
- Admin operations: Require `admin` or `super_admin` role
- User operations: Require authenticated user + own data access
- Public operations: No authentication required

### Data Consistency
- All subscriber data flows through `SubscriberElasticService`
- Unified data model prevents fragmentation
- Single source of truth for all strands

---

**Last Updated**: 2025-10-24  
**Migration Status**: 🔄 IN PROGRESS  
**Next Phase**: Complete streaming-subscribers.ts migration
