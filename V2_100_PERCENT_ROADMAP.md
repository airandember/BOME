# Roadmap to 100% Stripe V2 Compliance

## 📊 **Current Status: ~85% Compliant**

### ✅ **COMPLETED (V2 Compliant)**
- [x] Video Access (`video_access.go`) ✅
- [x] Subscription Manager (`subscription_manager_service.go`) ✅
- [x] Customer Linking (`customer_linking_service.go`) ✅
- [x] Webhook Handlers (`stripe_webhook_routes.go`) ✅ (dual-write)
- [x] Enhanced Subscribers (`enhanced_subscribers.go`) ✅
- [x] Stripe Sync V2 (`stripe_sync_v2.go`) ✅
- [x] Subscriber Elastic V2 (`subscriber_elastic_service_v2.go`) ✅
- [x] **Middleware** ✅ (No Stripe table references - clean!)
- [x] **Authentication** ✅ (No Stripe table references - clean!)

### ⚠️ **REMAINING V1 REFERENCES (15%)**

---

## 🎯 **Priority 1: Active Production Services** 🔴

### **1. Subscriber Elastic Service (V1)** - 10 matches
- **File**: `backend/internal/services/subscriber_elastic_service.go`
- **Status**: ⚠️ **HAS V2 VERSION AVAILABLE**
- **Action**: Switch routes to use V2 version
- **Impact**: HIGH - Used by admin dashboard
- **Effort**: LOW (V2 already exists!)

**Route Files to Update:**
- `backend/internal/routes/subscriber_elastic_routes.go` (switch to V2)

**Current:**
```go
elastic.GET("/subscribers", elasticHandler.GetAllUnifiedSubscribers)
// Uses: subscriber_elastic_service.go (V1)
```

**Should Be:**
```go
elastic.GET("/subscribers", elasticHandlerV2.GetAllUnifiedSubscribers)
// Uses: subscriber_elastic_service_v2.go (V2)
```

---

### **2. Simple Subscribers Service** - 18 matches
- **File**: `backend/internal/services/subscribers.go`
- **Status**: ⚠️ Actively used by `/subscribers` routes
- **Action**: Migrate to V2 tables
- **Impact**: HIGH - Core subscriber management
- **Effort**: MEDIUM (needs query updates)

**Lines to Fix:**
- Line 121-122: User to Stripe customer join
- Line 169-170: Subquery for active subscriptions
- Line 456-461: Subscriber listing
- Line 703-708: Non-subscriber listing

**Usage in Routes:**
- `backend/internal/routes/subscribers.go` - Main subscriber endpoints

---

## 🎯 **Priority 2: Admin/Reporting Services** 🟡

### **3. Stripe Analytics Routes** - 48 matches
- **File**: `backend/internal/routes/stripe_analytics_routes.go`
- **Status**: ⚠️ Admin analytics dashboard
- **Action**: Audit and migrate queries
- **Impact**: MEDIUM - Admin reporting only
- **Effort**: HIGH (many queries)

---

### **4. Admin Routes** - 3 matches
- **File**: `backend/internal/routes/admin.go`
- **Status**: ⚠️ Admin-specific queries
- **Action**: Migrate to V2
- **Impact**: LOW - Admin only
- **Effort**: LOW

---

## 🎯 **Priority 3: Legacy/Sync Services** 🔵

### **5. Stripe Sync (V1)** - 32 matches
- **File**: `backend/internal/services/stripe_sync.go`
- **Status**: ℹ️ **LEGACY** - Kept for backward compatibility
- **Action**: Document as deprecated, ensure V2 is primary
- **Impact**: LOW - V2 version is primary
- **Effort**: LOW (documentation only)

---

### **6. Simple Stripe Sync** - 12 matches
- **File**: `backend/internal/services/simple_stripe_sync.go`
- **Status**: ℹ️ Legacy sync utility
- **Action**: Deprecate or migrate
- **Impact**: LOW
- **Effort**: LOW

---

### **7. Comprehensive Stripe Sync** - 6 matches
- **File**: `backend/internal/services/comprehensive_stripe_sync.go`
- **Status**: ℹ️ Legacy sync utility
- **Action**: Deprecate or migrate
- **Impact**: LOW
- **Effort**: LOW

---

## 🎯 **Priority 4: Dev/Test Utilities** 🟢

### **8. Stripe Test Routes** - 8 matches
- **File**: `backend/internal/routes/stripe_test_routes.go`
- **Status**: ℹ️ Test/debug endpoints
- **Action**: Update or remove
- **Impact**: NONE (dev only)
- **Effort**: LOW

---

### **9. Ghost Customers** - 2 matches
- **File**: `backend/internal/routes/ghost_customers.go`
- **Status**: ℹ️ Debug utility
- **Action**: Audit and update
- **Impact**: LOW
- **Effort**: LOW

---

## 🎯 **Priority 5: Database Layer** 🟣

### **10. User Database Methods** - 14 matches
- **File**: `backend/internal/database/user.go`
- **Status**: ⚠️ Core database methods
- **Action**: Audit and migrate
- **Impact**: VARIES
- **Effort**: MEDIUM

---

### **11. Database Helpers** - 30 matches
- **File**: `backend/internal/database/database.go`
- **Status**: ⚠️ Utility methods
- **Action**: Audit and migrate
- **Impact**: VARIES
- **Effort**: MEDIUM

---

### **12. Subscriber Elastic Handler** - 1 match
- **File**: `backend/internal/handlers/subscriber_elastic_handler.go`
- **Status**: ⚠️ Should use V2 service
- **Action**: Point to V2 service
- **Impact**: MEDIUM
- **Effort**: LOW

---

### **13. Customer Linking Service** - 1 match (residual)
- **File**: `backend/internal/services/customer_linking_service.go`
- **Status**: ✅ Mostly V2, may have 1 comment or legacy fallback
- **Action**: Audit remaining reference
- **Impact**: LOW
- **Effort**: VERY LOW

---

## 📋 **Action Plan to 100%**

### **Phase 1: Quick Wins (1-2 hours)** ⚡
1. ✅ Switch Elastic routes to use V2 service (already exists!)
2. ✅ Update admin route queries (3 matches)
3. ✅ Fix subscriber elastic handler to use V2
4. ✅ Document V1 sync services as deprecated

**Result: ~92% compliant**

---

### **Phase 2: Core Services (3-4 hours)** 🔧
1. ✅ Migrate `subscribers.go` to V2 tables (18 matches)
2. ✅ Audit and update `user.go` database methods
3. ✅ Audit and update `database.go` utility methods

**Result: ~97% compliant**

---

### **Phase 3: Analytics & Reporting (2-3 hours)** 📊
1. ✅ Migrate `stripe_analytics_routes.go` (48 matches)
2. ✅ Update test/debug routes

**Result: ~99% compliant**

---

### **Phase 4: Cleanup (30 mins)** 🧹
1. ✅ Remove/deprecate legacy sync services
2. ✅ Add V2 compliance badges to all files
3. ✅ Document V1 → V2 migration complete

**Result: 💯 100% V2 COMPLIANT!**

---

## 🚀 **Start with Phase 1 (Quick Wins)**

### **Step 1: Switch Elastic Service to V2**

**File**: `backend/internal/routes/subscriber_elastic_routes.go`

**Find:**
```go
elasticService := services.NewSubscriberElasticService(db)
```

**Replace with:**
```go
elasticService := services.NewSubscriberElasticServiceV2(db)
```

This single change makes the elastic subscriber endpoint use V2 tables! 🎯

---

### **Step 2: Check if V2 Route Already Exists**

You might already have:
- `subscriber_elastic_routes.go` (V1)
- `subscriber_elastic_routes_v2.go` (V2) ← If this exists, just switch your main route to use it!

---

## 📊 **Benefits of 100% Compliance**

1. ✅ **Consistency**: All queries use same tables
2. ✅ **Performance**: Integer joins vs string joins
3. ✅ **Accuracy**: Single source of truth (V2 tables)
4. ✅ **Maintenance**: No confusion about which table to query
5. ✅ **Future-proof**: V1 tables can be archived

---

## 🎯 **Recommended Order**

1. **CRITICAL**: Subscriber Elastic Service → V2 (affects admin dashboard)
2. **HIGH**: Simple Subscribers → V2 (core functionality)
3. **MEDIUM**: Analytics routes → V2 (reporting)
4. **LOW**: Deprecate legacy sync services
5. **CLEANUP**: Test routes and utilities

---

## 📝 **Testing After Each Phase**

### **After Phase 1:**
- [ ] Admin elastic subscriber view works
- [ ] No errors in logs
- [ ] Data matches V2 tables

### **After Phase 2:**
- [ ] Subscriber management works
- [ ] User lookups accurate
- [ ] No V1 table queries in core flows

### **After Phase 3:**
- [ ] Analytics dashboard loads
- [ ] Reports show correct data
- [ ] Export functions work

---

## 🎊 **Current Priorities for Your Issue**

For fixing `bometesting@gmail.com` video access:

1. ✅ **Video Access** - Already fixed!
2. ✅ **Enhanced Subscribers** - Already fixed!
3. ⚠️ **Subscriber Elastic** - Switch to V2 (if admin uses it)
4. ⚠️ **Simple Subscribers** - Migrate to V2 (if admin uses it)

**The first two are done - your user should have access now!** 🎉

The remaining are for **100% compliance** and ensuring the **admin dashboard** shows consistent V2 data everywhere.

---

## 🔍 **Quick Audit Command**

To verify which files still have V1 references:

```bash
grep -r "stripe_customers[^_]" backend/internal --include="*.go" | wc -l
grep -r "stripe_customers_v2" backend/internal --include="*.go" | wc -l
```

Target: V1 matches = 0, V2 matches = all! 💯

