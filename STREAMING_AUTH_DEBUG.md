# 🔍 Streaming Dashboard Auth Debug

**Issue**: Super admin user redirected to `/auth/login?expired=true` when clicking "Streaming" dashboard  
**User Role**: `super_admin` (verified in backend logs)  
**Status**: ⏳ **DEBUGGING IN PROGRESS**

---

## 🐛 **SYMPTOMS**

1. User logs in via OAuth2 → successful (`role=super_admin`)
2. Navigate to `/admin/dashboard` → ✅ works
3. Click "Streaming" → ❌ **IMMEDIATE REDIRECT** to `/auth/login?expired=true`
4. Backend logs show: `AUTH_SUCCESS: user=aarongusa@gmail.com, id=7342, role=super_admin`
5. Frontend logs show: `✅ User authenticated as admin - access granted`
6. **BUT** still getting kicked out of streaming dashboard

---

## 🔍 **ROOT CAUSE ANALYSIS**

### **Theory 1: API Call Returning 401** ✅ **CONFIRMED**

The streaming layout's `onMount` calls `/admin/streaming/dashboard` immediately:

```typescript
// frontend/src/routes/admin/streaming/+layout.svelte
async function loadQuickStats() {
    const response = await api.get('/admin/streaming/dashboard');
    // If this returns 401, api/client.ts redirects to login
}
```

The `/admin/streaming/dashboard` endpoint has **`StreamingAdminRequired()` middleware**:

```go
// backend/internal/routes/admin_streaming.go
streaming := admin.Group("/streaming")
streaming.Use(middleware.AuthRequired())
streaming.Use(middleware.StreamingAdminRequired()) // ← RESTRICTIVE CHECK
{
    streaming.GET("/dashboard", GetStreamingDashboardHandler(db, analyticsService))
}
```

**`StreamingAdminRequired()` only allows**:
- `super_admin` ✅ (user has this)
- `system_admin`
- `content_manager`
- `streaming_manager`

### **Theory 2: Context Not Persisting Between Middlewares** ⏳ **INVESTIGATING**

Possible scenarios:
1. `AuthRequired()` sets `c.Set("user_role", claims.Role)` ✅
2. `StreamingAdminRequired()` tries to get `c.Get("user_role")` ❓
3. Context might not be persisting → role not found → returns 401 → frontend redirects

**Evidence**:
- Backend logs show `AUTH_SUCCESS` with `role=super_admin` ✅
- BUT no logs from `StreamingAdminRequired()` showing the role check
- This suggests the middleware might not even be reached, OR the role isn't in context

### **Theory 3: Middleware Order Issue** ⏳ **INVESTIGATING**

The `/admin/streaming` group applies middlewares in this order:
1. `AuthRequired()` - validates JWT, sets `user_role` in context
2. `StreamingAdminRequired()` - checks `user_role` from context

If there's a middleware execution issue, `user_role` might not be available.

---

## 🔧 **DEBUG CHANGES APPLIED**

### **1. Backend Logging** (`backend/internal/middleware/middleware.go`)

Added comprehensive logging to `StreamingAdminRequired()`:

```go
func StreamingAdminRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        log.Printf("🔍 [STREAMING-AUTH] Checking streaming admin permissions for path: %s", c.Request.URL.Path)
        
        role, exists := c.Get("user_role")
        log.Printf("🔍 [STREAMING-AUTH] Role from context - exists: %v, role: %v", exists, role)
        
        if !exists {
            log.Printf("❌ [STREAMING-AUTH] No user_role in context!")
            // Returns 401
        }
        
        roleStr := role.(string)
        log.Printf("🔍 [STREAMING-AUTH] Checking if role '%s' is in streaming admin list", roleStr)
        
        if !isStreamingAdmin {
            log.Printf("❌ [STREAMING-AUTH] Streaming admin access DENIED for user: %v (role: %s)", userEmail, roleStr)
            // Returns 403
        }
        
        log.Printf("✅ [STREAMING-AUTH] Streaming admin access GRANTED for role: %s", roleStr)
        c.Next()
    }
}
```

### **2. Frontend Logging** (`frontend/src/routes/admin/streaming/+layout.svelte`)

Added logging before API call:

```typescript
async function loadQuickStats() {
    console.log('🔍 [STREAMING] About to call /admin/streaming/dashboard');
    console.log('🔍 [STREAMING] Current user:', currentUser);
    console.log('🔍 [STREAMING] User role:', currentUser?.role);
    
    const response = await api.get('/admin/streaming/dashboard');
    console.log('🔍 [STREAMING] Dashboard API response:', response);
    // ...
}
```

### **3. API Client Logging** (`frontend/src/lib/api/client.ts`)

Added logging for 401 responses:

```typescript
if (response.status === 401 && token) {
    console.warn('🔴 [API] 401 Unauthorized for endpoint:', endpoint);
    console.warn('🔴 [API] Response status:', response.status);
    console.warn('🔴 [API] Redirecting to login due to 401');
    goto('/auth/login?expired=true');
}
```

---

## 🧪 **NEXT TESTING STEPS**

### **Restart Backend and Test**

1. **Restart the backend** (it's already rebuilt with new logging)
   ```bash
   cd backend
   ./bome-backend.exe
   ```

2. **Clear browser cache/storage** (to get fresh token)
   - Open DevTools → Application → Storage → Clear site data

3. **Re-login via OAuth2**

4. **Click "Streaming" in the sidebar**

5. **Check BOTH console logs**:
   
   **Frontend Console** (Chrome DevTools):
   ```
   🔍 [STREAMING] About to call /admin/streaming/dashboard
   🔍 [STREAMING] Current user: { email: "...", role: "super_admin" }
   🔍 [STREAMING] User role: super_admin
   [IF 401] 🔴 [API] 401 Unauthorized for endpoint: /admin/streaming/dashboard
   [IF 401] 🔴 [API] Redirecting to login due to 401
   ```
   
   **Backend Terminal**:
   ```
   🔍 [STREAMING-AUTH] Checking streaming admin permissions for path: /api/v1/admin/streaming/dashboard
   🔍 [STREAMING-AUTH] Role from context - exists: true, role: super_admin
   🔍 [STREAMING-AUTH] Checking if role 'super_admin' is in streaming admin list
   ✅ [STREAMING-AUTH] Role 'super_admin' matched admin role 'super_admin'
   ✅ [STREAMING-AUTH] Streaming admin access GRANTED for role: super_admin
   ```

---

## 🎯 **EXPECTED OUTCOMES**

### **Scenario A: Role Not in Context** (Most Likely)
**Backend logs will show**:
```
🔍 [STREAMING-AUTH] Role from context - exists: false, role: <nil>
❌ [STREAMING-AUTH] No user_role in context!
```

**Fix**: The `AuthRequired()` middleware isn't being called first, OR the context isn't persisting. Need to check middleware chain.

### **Scenario B: Role String Mismatch**
**Backend logs will show**:
```
🔍 [STREAMING-AUTH] Role from context - exists: true, role: SUPER_ADMIN
🔍 [STREAMING-AUTH] Checking if role 'SUPER_ADMIN' is in streaming admin list
❌ [STREAMING-AUTH] Streaming admin access DENIED (role: SUPER_ADMIN)
```

**Fix**: Role string case sensitivity issue. Need to normalize to lowercase.

### **Scenario C: Middleware Not Reached**
**Backend logs will show**:
```
AUTH_SUCCESS: user=aarongusa@gmail.com, id=7342, role=super_admin
[NO STREAMING-AUTH LOGS]
```

**Fix**: Request isn't reaching the middleware, possibly a routing issue.

### **Scenario D: Everything Works** (Hoped For!)
**Backend logs will show**:
```
🔍 [STREAMING-AUTH] Role from context - exists: true, role: super_admin
✅ [STREAMING-AUTH] Streaming admin access GRANTED for role: super_admin
```

**Frontend**: No redirect, streaming dashboard loads successfully!

---

## 📋 **FILES MODIFIED**

1. **`backend/internal/middleware/middleware.go`**
   - Added debug logging to `StreamingAdminRequired()`
   - Lines 463-510

2. **`frontend/src/routes/admin/streaming/+layout.svelte`**
   - Added logging before `/admin/streaming/dashboard` API call
   - Lines 124-129

3. **`frontend/src/lib/api/client.ts`**
   - Added logging for 401 responses
   - Lines 79-88

---

## 🔐 **SECURITY NOTES**

- **Frontend auth checks** = UX only (can be bypassed with DevTools)
- **Backend middleware** = REAL security (cannot be bypassed)
- The redirect is happening because backend is returning 401/403
- Even if we "fix" the frontend, the backend must allow access for it to work

---

## 📞 **NEXT ACTIONS**

1. ✅ Backend rebuilt with debug logging
2. ⏳ **AWAITING USER**: Restart backend server
3. ⏳ **AWAITING USER**: Re-login and click "Streaming"
4. ⏳ **AWAITING USER**: Share backend terminal logs showing `[STREAMING-AUTH]` messages
5. ⏳ Based on logs, apply targeted fix

---

**Status**: ⏳ **READY FOR TESTING**  
**Blocker**: Need backend server restart and fresh test with new logging

**Expected Fix Time**: < 5 minutes once we see the logs

