# Production Endpoint Testing Guide

## 🔍 Diagnosing the 504 Gateway Timeout Issue

### Test Endpoints (in order):

1. **Simple Ping Test** (no Stripe dependency):
   ```
   GET https://bometest-f2nod.ondigitalocean.app/bome-backend/api/v1/admin/streaming/stripe/ping
   ```
   - **Expected**: Immediate 200 OK response
   - **If 504**: Load balancer/reverse proxy issue
   - **If 200**: Backend is reachable

2. **Stripe Health Check** (quick Stripe test):
   ```
   GET https://bometest-f2nod.ondigitalocean.app/bome-backend/api/v1/admin/streaming/stripe/health
   ```
   - **Expected**: 503 Service Unavailable (if Stripe disabled) or 200 OK (if enabled)
   - **If 504**: Stripe service initialization issue

3. **Dashboard Endpoint** (full analytics):
   ```
   GET https://bometest-f2nod.ondigitalocean.app/bome-backend/api/v1/admin/streaming/stripe/dash
   ```
   - **Expected**: 503 Service Unavailable (if Stripe disabled) or analytics data
   - **If 504**: Analytics timeout issue

### Expected Responses:

#### When Stripe is Disabled (cleared configuration):
```json
{
  "error": "Stripe service is not enabled",
  "enabled": false,
  "debug": "service_disabled"
}
```

#### When Stripe is Enabled but times out:
```json
{
  "enabled": true,
  "error": "Analytics request timed out - using fallback data",
  "timeout": true,
  "timeout_duration": "30s",
  "method": "fallback_timeout_protection"
}
```

### Logging to Check:

Look for these log patterns in your backend logs:

1. **Request received**:
   ```
   🚀 [DASH-START] Dashboard request initiated at 2025-08-30...
   ```

2. **Service check**:
   ```
   ✅ [DASH-SERVICE] Stripe service object exists
   ⏱️ [DASH-ENABLED-CHECK] IsEnabled() took 1.2ms, result: false
   ```

3. **Response sent**:
   ```
   ❌ [DASH-ERROR] Stripe service is not enabled
   ```

### Possible Issues:

1. **No logs at all** → Load balancer timeout (check DigitalOcean settings)
2. **Logs stop at [DASH-START]** → Backend hanging on service check
3. **Logs show service disabled** → Should return 503, not 504
4. **Code not deployed** → Still using old version without timeout protection

### Quick Fixes to Try:

1. **Test the ping endpoint first** to confirm backend connectivity
2. **Check DigitalOcean load balancer timeout settings** (should be > 30 seconds)
3. **Verify the updated code is deployed** to production
4. **Check backend logs** for the detailed logging we added
