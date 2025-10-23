# 🧪 Phase 7C: Analytics Testing Guide

**Created:** October 22, 2025  
**Status:** Ready for Testing  
**Prerequisites:** Backend must be recompiled with new analytics routes

---

## ⚠️ IMPORTANT: RECOMPILE REQUIRED

The existing `bome-backend-phase7.exe` was compiled **before** we added the analytics routes.  
**You must recompile the backend** to include the new analytics endpoints.

---

## 🔧 STEP 1: REBUILD BACKEND

### Option A: Standard Build
```bash
cd backend
go build -o bin/bome-backend-analytics.exe .
```

### Option B: Clean Build (if module cache issues)
```bash
cd backend
go clean -modcache
go mod tidy
go build -o bin/bome-backend-analytics.exe .
```

### Option C: Quick Test Build
```bash
cd backend
go run main.go
```

---

## 🚀 STEP 2: START BACKEND

### Start the Server
```bash
cd backend
./bin/bome-backend-analytics.exe
```

### Expected Output
```
🚀 Starting BOME Backend...
📊 Database connected
Setting up Stripe analytics routes...
✅ Stripe analytics routes setup complete
📊 STREAMING ANALYTICS ROUTES
✅ Streaming analytics routes setup complete
...
🌐 Server running on :8080
```

**Look for:** "✅ Streaming analytics routes setup complete"

---

## 🧪 STEP 3: TEST ANALYTICS ENDPOINTS

### Test 1: Analytics Overview
```bash
curl http://localhost:8080/api/v1/admin/streaming/analytics/overview?period=30d
```

**Expected Response:**
```json
{
  "users": {
    "total": 1250,
    "new_today": 12,
    "active_today": 487,
    "growth_rate": 13.2
  },
  "videos": {
    "total": 342,
    "published": 298,
    "total_views": 45230
  },
  "subscriptions": {
    "revenue_month": 12450.00,
    "mrr": 12450.00
  },
  "video_stats": {
    "total_videos": 342,
    "synced_videos": 298,
    "needs_attention": 5,
    "total_views": 45230
  },
  "subscriber_metrics": {
    "total_subscribers": 1250,
    "active_subscriptions": 412,
    "monthly_revenue": 12450.00,
    "churn_rate": 2.9
  },
  "view_analytics": {
    "total_views": 45230,
    "views_today": 3420,
    "views_week": 15230,
    "growth_rate": 8.5
  },
  "period": "30d"
}
```

**Status:** ✅ Should return 200 OK with real data

---

### Test 2: Executive Summary
```bash
curl http://localhost:8080/api/v1/admin/streaming/analytics/executive-summary?period=30d
```

**Expected Response:**
```json
{
  "period": "30d",
  "revenue_impact": {
    "promotional_revenue": 0,
    "standard_revenue": 12450,
    "total_mrr": 12450,
    "growth_rate": 0
  },
  "customer_impact": {
    "new_customers_promos": 0,
    "standard_conversions": 412,
    "overall_growth": 0,
    "total_subscribers": 1250
  },
  "funnel_performance": {
    "promo_conversion": 0,
    "standard_conversion": 0,
    "conversion_lift": 0
  },
  "content_performance": {
    "total_videos": 342,
    "total_views": 45230,
    "synced_videos": 298,
    "needs_attention": 5
  }
}
```

**Status:** ✅ Should return 200 OK

---

### Test 3: Funnel Analysis
```bash
curl http://localhost:8080/api/v1/admin/streaming/analytics/funnel?period=30d
```

**Expected Response:**
```json
{
  "period": "30d",
  "stages": [
    {
      "name": "Visitors",
      "count": 0,
      "conversion": 0
    },
    {
      "name": "Sign-ups",
      "count": 1250,
      "conversion": 0
    },
    {
      "name": "Trial Users",
      "count": 0,
      "conversion": 0
    },
    {
      "name": "Active Subscribers",
      "count": 412,
      "conversion": 0
    }
  ],
  "overall_conversion": 0,
  "drop_off_points": []
}
```

**Status:** ✅ Should return 200 OK

---

### Test 4: Revenue Impact
```bash
curl http://localhost:8080/api/v1/admin/streaming/analytics/revenue-impact?period=30d
```

**Expected Response:**
```json
{
  "period": "30d",
  "revenue_by_source": [
    {
      "source": "Subscriptions",
      "amount": 12450
    }
  ],
  "total_revenue": 12450,
  "mrr": 12450,
  "arr": 0,
  "churn_rate": 2.9,
  "growth_rate": 0
}
```

**Status:** ✅ Should return 200 OK

---

### Test 5: Customer Journey
```bash
curl http://localhost:8080/api/v1/admin/streaming/analytics/customer-journey?period=30d
```

**Expected Response:**
```json
{
  "period": "30d",
  "lifecycle_stages": [
    {
      "stage": "Prospect",
      "count": 0
    },
    {
      "stage": "Trial",
      "count": 0
    },
    {
      "stage": "Active",
      "count": 412
    },
    {
      "stage": "At Risk",
      "count": 0
    },
    {
      "stage": "Churned",
      "count": 0
    }
  ],
  "avg_time_to_conversion": "0 days",
  "retention_rate": 0,
  "cohort_analysis": []
}
```

**Status:** ✅ Should return 200 OK

---

### Test 6: Promotions Analytics
```bash
curl http://localhost:8080/api/v1/admin/streaming/analytics/promotions?period=30d
```

**Expected Response:**
```json
{
  "period": "30d",
  "active_promotions": [],
  "promotion_performance": {
    "total_uses": 0,
    "total_revenue": 0,
    "conversion_rate": 0
  },
  "coupon_usage": [],
  "conversion_impact": {
    "with_promo": 0,
    "without_promo": 0,
    "lift": 0
  }
}
```

**Status:** ✅ Should return 200 OK

---

### Test 7: Real-Time Analytics
```bash
curl http://localhost:8080/api/v1/admin/streaming/analytics/real-time
```

**Expected Response:**
```json
{
  "active_users": 5,
  "current_streams": 2,
  "server_load": 0.25,
  "bandwidth_usage": "45.2 MB/s",
  "recent_signups": 3,
  "recent_subscriptions": 2,
  "error_rate": 0.01,
  "response_time": 150,
  "live_events": [
    {
      "time": "2025-10-22T14:35:00Z",
      "event": "video_view",
      "details": "{\"video_id\": 123}"
    }
  ],
  "top_content_now": [
    {
      "title": "Getting Started with BOME",
      "viewers": 45
    }
  ]
}
```

**Status:** ✅ Should return 200 OK

---

### Test 8: System Health
```bash
curl http://localhost:8080/api/v1/admin/streaming/analytics/system-health
```

**Expected Response:**
```json
{
  "health": {
    "uptime": "5 days 12 hours",
    "response_time": "150ms",
    "error_rate": "0.10%",
    "storage_used": "50.0 GB",
    "bandwidth_used": "20.5 MB/s",
    "cdn_hits": "145,230",
    "database_size": "50.0 GB",
    "active_sessions": 127,
    "last_write": "2025-10-22T14:35:00Z",
    "total_events_tracked": 1245230
  }
}
```

**Status:** ✅ Should return 200 OK

---

## 🌐 STEP 4: TEST FRONTEND

### Start Frontend Dev Server
```bash
cd frontend
npm run dev
```

### Navigate to Analytics Dashboard
Open browser to: `http://localhost:5173/admin/streaming/analytics`

### Test Checklist
- [ ] Page loads without errors
- [ ] No console errors
- [ ] "Overview" tab displays data
- [ ] "Executive Summary" tab displays data
- [ ] "Funnel Analysis" tab displays data
- [ ] "Revenue Impact" tab displays data
- [ ] "Customer Journey" tab displays data
- [ ] "Promotions" tab displays data
- [ ] Period selector works (7d, 30d, 90d, 1y)
- [ ] Charts render correctly
- [ ] Data updates when period changes
- [ ] No mock data fallbacks (if backend is running)

---

## 🔍 DEBUGGING

### If Endpoints Return 404
**Cause:** Backend was not recompiled with new routes  
**Solution:** Rebuild backend (see Step 1)

### If Endpoints Return 500
**Cause:** Database query errors or missing tables  
**Check:** Backend logs for SQL errors  
**Solution:** Verify database schema and run any missing migrations

### If Frontend Shows Mock Data
**Cause:** Backend is not running or endpoints are failing  
**Check:** Browser console for API errors  
**Solution:** Verify backend is running and responding

### If Go Build Fails
**Error:** "The system cannot find the file specified"  
**Cause:** Go module cache corruption  
**Solution:** Run `go clean -modcache && go mod tidy`

---

## 📊 DATA VALIDATION

### Check Users Table
```sql
SELECT COUNT(*) FROM users;
```
**Expected:** Should match `users.total` in analytics

### Check Subscriptions Table
```sql
SELECT COUNT(*) FROM subscriptions WHERE status = 'active';
```
**Expected:** Should match `subscriber_metrics.active_subscriptions`

### Check Videos Table
```sql
SELECT COUNT(*) FROM master_video_list;
SELECT SUM(views) FROM master_video_list;
```
**Expected:** Should match `video_stats.total_videos` and `total_views`

### Check Real-Time Metrics
```sql
SELECT COUNT(DISTINCT user_id) 
FROM user_sessions 
WHERE last_activity > NOW() - INTERVAL '5 minutes' 
AND is_active = true;
```
**Expected:** Should match `active_users` in real-time analytics

---

## ✅ SUCCESS CRITERIA

### Backend Tests Pass
- ✅ All 8 endpoints return 200 OK
- ✅ Responses contain valid JSON
- ✅ Data values are realistic (not all zeros)
- ✅ Response time < 2 seconds
- ✅ No errors in backend logs

### Frontend Tests Pass
- ✅ Dashboard loads successfully
- ✅ All tabs display data
- ✅ Charts render without errors
- ✅ Period selector updates data
- ✅ No console errors
- ✅ Real data (not mock data) is displayed

### Integration Tests Pass
- ✅ Frontend calls backend successfully
- ✅ Data flows from database → models → service → handler → frontend
- ✅ Caching works (subsequent requests are faster)
- ✅ Error handling works (graceful degradation if data missing)

---

## 🎉 COMPLETION CHECKLIST

Once all tests pass:

- [ ] All 8 analytics endpoints tested and working
- [ ] Frontend analytics dashboard displays real data
- [ ] Period selector works across all tabs
- [ ] Charts render correctly
- [ ] Performance is acceptable (< 2s per request)
- [ ] No errors in backend or frontend logs
- [ ] Data accuracy verified against database
- [ ] Update `PHASE_7C_IMPLEMENTATION_COMPLETE.md` with test results
- [ ] Mark analytics braid as 100% complete
- [ ] Update platform completion to 100%
- [ ] **CELEBRATE! 🎊**

---

## 🐛 KNOWN ISSUES

### Issue 1: Audit Logs Endpoint Missing
- **Frontend:** Calls `/admin/streaming/analytics/audit-logs`
- **Backend:** Endpoint not implemented
- **Impact:** Low (graceful failure)
- **Fix:** Add endpoint or handle in frontend

### Issue 2: Placeholder Values
- **Issue:** Some metrics return 0 or placeholder values
- **Cause:** Tables don't exist or have no data
- **Impact:** Low (expected for new installations)
- **Fix:** Populate tables with real data

### Issue 3: Mock Data Fallbacks
- **Issue:** Frontend shows mock data when API fails
- **Cause:** Intentional fallback for development
- **Impact:** None (helps during development)
- **Fix:** Remove fallbacks after confirming real data works

---

## 🚀 NEXT STEPS AFTER TESTING

1. **Document Test Results**
   - Update `PHASE_7C_IMPLEMENTATION_COMPLETE.md`
   - Note any bugs or issues found
   - Record performance metrics

2. **Fix Any Issues**
   - Address 404 errors (rebuild)
   - Fix 500 errors (query issues)
   - Optimize slow queries

3. **Deploy to Production**
   - Build production backend
   - Update environment variables
   - Run migrations if needed
   - Monitor performance

4. **Announce 100% Completion**
   - Update platform status
   - Celebrate achievement
   - Plan next features

---

*Phase 7C Testing Guide*  
*October 22, 2025*  
*Ready for comprehensive testing after backend recompile*

