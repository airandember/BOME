# Stripe V2 Table Compliance Audit

## 🎯 **Status Summary**

| Component | Status | Notes |
|-----------|--------|-------|
| **Core Video Access** | ✅ **FIXED** | Now uses V2 tables |
| **Subscription Manager** | ✅ **CLEAN** | Using V2 tables |
| **Customer Linking** | ✅ **CLEAN** | Using V2 tables |
| **Webhook Handlers** | ⚠️ **DUAL-WRITE** | V1 + V2 (intentional) |
| **Enhanced Subscribers** | ❌ **V1** | Needs migration |
| **Elastic Subscriber Service** | ❌ **V1** | Has V2 version available |
| **Simple Subscribers** | ❌ **V1** | Needs migration |
| **Stripe Sync (Legacy)** | ⚠️ **V1** | Kept for backward compat |

---

## ✅ **COMPLIANT (Using V2 Tables)**

### **1. Video Access (JUST FIXED)** ✅
- **File**: `backend/internal/database/video_access.go`
- **Tables Used**: `user_stripe_customers_v2`, `stripe_customers_v2`, `stripe_subscriptions_v2`, `stripe_prices_v2`, `stripe_products_v2`
- **Status**: **FULLY COMPLIANT** - Just updated to V2

### **2. Subscription Manager** ✅
- **File**: `backend/internal/services/subscription_manager_service.go`
- **Tables Used**: V2 tables only
- **Status**: **FULLY COMPLIANT**

### **3. Customer Linking Service** ✅
- **File**: `backend/internal/services/customer_linking_service.go`
- **Tables Used**: `stripe_customers_v2`, `stripe_subscriptions_v2`, `user_stripe_customers_v2`
- **Status**: **FULLY COMPLIANT**

### **4. Stripe Sync V2** ✅
- **File**: `backend/internal/services/stripe_sync_v2.go`
- **Tables Used**: All V2 tables
- **Status**: **FULLY COMPLIANT**

### **5. Subscriber Elastic Service V2** ✅
- **File**: `backend/internal/services/subscriber_elastic_service_v2.go`
- **Tables Used**: V2 tables only
- **Status**: **FULLY COMPLIANT** - This is the V2 version

---

## ⚠️ **INTENTIONAL DUAL-WRITE (V1 + V2)**

### **Webhook Handlers**
- **File**: `backend/internal/routes/stripe_webhook_routes.go`
- **Strategy**: Writes to BOTH V1 and V2 tables during migration period
- **Status**: **INTENTIONAL** - Ensures data integrity during transition
- **Note**: Can be cleaned up after V1 deprecation

---

## ❌ **NON-COMPLIANT (Still Using V1)**

### **1. Enhanced Subscribers Service** ⚠️ **ACTIVELY USED**
- **File**: `backend/internal/services/enhanced_subscribers.go`
- **Lines**: 235, 239-241, 331, 335-337, 584, 588-590
- **Tables Used**: `stripe_customers`, `stripe_subscriptions`, `stripe_products`, `stripe_prices`
- **Impact**: **HIGH** - Used by admin dashboard
- **Action Required**: Migrate to V2 tables

**Problem Queries:**
```go
LEFT JOIN stripe_customers sc ON (
    u.stripe_customer_id = sc.stripe_id OR 
    sc.stripe_id = ANY(COALESCE(u.stripe_customer_ids, '{}'))
)
LEFT JOIN stripe_subscriptions ss ON sc.id = ss.customer_id AND ss.status IN ('active', 'trialing')
LEFT JOIN stripe_prices sp_price ON ss.stripe_price_id = sp_price.stripe_id
LEFT JOIN stripe_products stripe_prod ON sp_price.product_id = stripe_prod.id
```

**Should Be:**
```go
LEFT JOIN user_stripe_customers_v2 usc ON usc.user_id = u.id
LEFT JOIN stripe_customers_v2 sc ON sc.id = usc.stripe_customer_id
LEFT JOIN stripe_subscriptions_v2 ss ON ss.customer_id = sc.id AND ss.status IN ('active', 'trialing')
LEFT JOIN stripe_prices_v2 sp_price ON ss.price_id = sp_price.id
LEFT JOIN stripe_products_v2 stripe_prod ON sp_price.product_id = stripe_prod.id
```

---

### **2. Elastic Subscriber Service (V1)** ⚠️
- **File**: `backend/internal/services/subscriber_elastic_service.go`
- **Lines**: 57, 61-63, 295, 299-301
- **Tables Used**: `stripe_customers`, `stripe_subscriptions`, `stripe_products`, `stripe_prices`
- **Impact**: **MEDIUM** - V2 version exists (`subscriber_elastic_service_v2.go`)
- **Action Required**: **Switch routes to use V2 service**

---

### **3. Simple Subscribers Service** ⚠️
- **File**: `backend/internal/services/subscribers.go`
- **Lines**: 87, 121-124
- **Tables Used**: `stripe_customers`, `stripe_subscriptions`, `stripe_prices`, `stripe_products`
- **Impact**: **MEDIUM** - May be legacy code
- **Action Required**: Migrate to V2 or deprecate

---

### **4. Stripe Sync (V1 - Legacy)** ℹ️
- **File**: `backend/internal/services/stripe_sync.go`
- **Multiple References**: Lines 323-1902
- **Tables Used**: All V1 tables
- **Impact**: **LOW** - Kept for backward compatibility
- **Action Required**: **Document as legacy, ensure V2 is primary**

---

## 🔧 **CRITICAL FIXES NEEDED**

### **Priority 1: Enhanced Subscribers** 🔴
**Why**: Actively used by admin dashboard for subscriber management

**Fix Strategy**:
1. Update queries to use `user_stripe_customers_v2` as the linking table
2. Join to V2 tables instead of V1
3. Update column references (e.g., `ss.price_id` instead of `ss.stripe_price_id`)

**Files to Update**:
- `backend/internal/services/enhanced_subscribers.go` (3 queries)

---

### **Priority 2: Switch to Elastic V2** 🟡
**Why**: V2 version already exists and is ready

**Fix Strategy**:
1. Update routes to use `subscriber_elastic_service_v2.go` instead of V1
2. Deprecate V1 file

**Files to Check**:
- Route handlers that call elastic subscriber service

---

## 📊 **V1 vs V2 Table Mapping**

| V1 Table | V2 Table | Key Difference |
|----------|----------|----------------|
| `stripe_customers` | `stripe_customers_v2` | V2 uses integer PKs for performance |
| `stripe_subscriptions` | `stripe_subscriptions_v2` | V2 uses `customer_id` (int FK) instead of `stripe_customer_id` (string) |
| `stripe_products` | `stripe_products_v2` | V2 has proper `video_approved` flag |
| `stripe_prices` | `stripe_prices_v2` | V2 uses `product_id` (int FK) |
| *N/A* | `user_stripe_customers_v2` | **NEW**: Many-to-many linking table |

---

## 🎯 **V2 Query Pattern**

### **Correct V2 Pattern:**
```sql
SELECT ... 
FROM users u
INNER JOIN user_stripe_customers_v2 usc ON usc.user_id = u.id
INNER JOIN stripe_customers_v2 sc ON sc.id = usc.stripe_customer_id
INNER JOIN stripe_subscriptions_v2 ss ON ss.customer_id = sc.id
INNER JOIN stripe_prices_v2 sp ON sp.id = ss.price_id
INNER JOIN stripe_products_v2 sprod ON sprod.id = sp.product_id
WHERE ...
```

### **Incorrect V1 Pattern:**
```sql
SELECT ... 
FROM users u
INNER JOIN stripe_customers sc ON (
    u.stripe_customer_id = sc.stripe_id OR 
    sc.stripe_id = ANY(u.stripe_customer_ids)
)
INNER JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
INNER JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
WHERE ...
```

---

## ✅ **Next Steps**

1. ✅ **Video Access Fixed** - Using V2 tables
2. 🔲 **Fix Enhanced Subscribers** - Migrate 3 queries to V2
3. 🔲 **Switch to Elastic V2** - Update route handlers
4. 🔲 **Document V1 Sync as Legacy** - Add deprecation notice
5. 🔲 **Test End-to-End** - Ensure admin dashboard works with V2

---

## 📝 **Testing Checklist**

After migrating to V2:
- [ ] Video access works (✅ Already fixed!)
- [ ] Webhooks process correctly
- [ ] Admin dashboard loads subscribers
- [ ] Enhanced subscriber filters work
- [ ] Customer linking functions
- [ ] Subscription status updates

---

## 🎊 **Current Status**

**Video Access: ✅ FIXED**
- User `bometesting@gmail.com` will now have video access
- System correctly queries V2 tables
- Missing columns added

**Admin Features: ⚠️ NEEDS UPDATE**
- Enhanced subscribers still on V1
- Should be migrated for consistency

**Overall V2 Adoption: ~70%** 
- Core functionality: ✅ V2
- Admin/reporting: ⚠️ Mixed V1/V2
- Legacy sync: ℹ️ Intentionally V1

