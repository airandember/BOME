# Video Analytics BRAID - Logging Guide 🧵

## Overview
Strategic logging added throughout the Video Analytics BRAID to trace data flow from frontend → backend → Redis → database. This guide explains the logging structure and how to interpret the logs.

---

## 🎯 Logging Layers

### Layer 1: Frontend (Svelte/TypeScript)
**File**: `frontend/src/lib/services/videoAnalytics.ts`

**Log Prefixes**:
- `🎬 [FRONTEND]` - Main tracking method entry
- `⏭️  [FRONTEND]` - Skipped tracking (throttling)
- `📤 [FRONTEND]` - Sending request
- `✅ [FRONTEND]` - Success
- `❌ [FRONTEND]` - Error
- `🌐 [FRONTEND→BACKEND]` - HTTP request preparation
- `📥 [FRONTEND←BACKEND]` - HTTP response received
- `🔑 [FRONTEND]` - Authentication info

### Layer 2: Route Handler (Go/Gin)
**File**: `backend/internal/routes/video_analytics_routes.go`

**Log Prefixes**:
- `🌐 [ROUTE]` - HTTP endpoint entry/exit
- `📦 [ROUTE]` - Request payload
- `🔐 [ROUTE]` - Authentication
- `📝 [ROUTE]` - Request enrichment
- `🛡️  [ROUTE]` - Circuit breaker check
- `📤 [ROUTE→SERVICE]` - Calling analytics service
- `✅ [ROUTE←SERVICE]` - Service response
- `📤 [ROUTE→HISTORY]` - Watch history update
- `✅ [ROUTE←HISTORY]` - History response
- `📤 [ROUTE→FRONTEND]` - HTTP response
- `❌ [ROUTE]` - Error

### Layer 3: Analytics Service (Go)
**File**: `backend/internal/services/video_analytics_service.go`

**Log Prefixes**:
- `🎯 [SERVICE]` - Service method entry/exit
- `📤 [SERVICE→BUFFER]` - Sending to Redis buffer
- `✅ [SERVICE←BUFFER]` - Buffer response
- `⚡ [SERVICE]` - Async operation indicator
- `⚠️ [SERVICE]` - Fallback triggered
- `🔄 [SERVICE]` - Retry or alternate path
- `📤 [SERVICE→DB]` - Direct database write
- `💾 [SERVICE-DB]` - Database operation details

### Layer 4: Analytics Buffer (Go/Redis)
**File**: `backend/internal/services/analytics_buffer.go`

**Log Prefixes**:
- `📦 [BUFFER]` - Buffer operation entry/exit
- `🔄 [BUFFER]` - Processing
- `📤 [BUFFER→REDIS]` - Writing to Redis
- `✅ [BUFFER←REDIS]` - Redis response
- `📊 [BUFFER]` - Buffer status
- `⏳ [BUFFER]` - Waiting for flush
- `🚀 [BUFFER]` - Triggering flush
- `🔥 [BUFFER-FLUSH]` - Flush operation
- `📥 [BUFFER-FLUSH←REDIS]` - Reading from Redis
- `✂️  [BUFFER-FLUSH→REDIS]` - Trimming Redis
- `💾 [BUFFER-FLUSH→DB]` - Writing to database
- `📈 [BUFFER-FLUSH]` - Statistics
- `❌ [BUFFER]` - Error

---

## 📊 Complete Flow Example

Here's what you'll see when a user watches a video:

```
=== FRONTEND ===
🎬 [FRONTEND] trackProgress called: video=12845, time=10s, duration=180s
📤 [FRONTEND] Sending tracking event: video=12845, time=10s, %=5.6%
🌐 [FRONTEND→BACKEND] Preparing request to /api/v1/analytics/video/track
🔑 [FRONTEND] Using JWT authentication
📥 [FRONTEND←BACKEND] Response received in 4.23ms: 200
📋 [FRONTEND] Response data: {status: 'tracked', video_id: 12845}
✅ [FRONTEND] Backend confirmed tracking for video 12845
✅ [FRONTEND] Successfully tracked: video=12845, time=10s, %=5.6%

=== BACKEND ROUTE ===
🌐 [ROUTE] ============================================
🌐 [ROUTE] Received POST /analytics/video/track
🌐 [ROUTE] Client IP: 127.0.0.1, User-Agent: Mozilla/5.0...
📦 [ROUTE] Request payload: video_id=12845, duration=10s, percentage=5.56%
🔐 [ROUTE] Authenticated user: 42
📝 [ROUTE] Enhanced request with IP=127.0.0.1, UA=Mozilla...
🛡️  [ROUTE] Checking circuit breaker...
✅ [ROUTE] Circuit breaker OK - proceeding
📤 [ROUTE→SERVICE] Calling analyticsService.RecordView()

=== ANALYTICS SERVICE ===
🎯 [SERVICE] ========================================
🎯 [SERVICE] RecordView called
🎯 [SERVICE] Video: 12845, User: 42, Duration: 10s, Percentage: 5.56%
📤 [SERVICE→BUFFER] Buffer available, adding event to Redis
  
  === BUFFER ===
  📦 [BUFFER] ======================================
  📦 [BUFFER] AddEvent called: video=12845, user=42
  ✅ [BUFFER] Validation passed
  🔄 [BUFFER] Serializing event to JSON
  ✅ [BUFFER] Event serialized (156 bytes)
  📤 [BUFFER→REDIS] Pushing to Redis list: analytics:video_tracking_buffer
  ✅ [BUFFER←REDIS] Event pushed to Redis successfully
  📊 [BUFFER] Current buffer size: 23/100
  ⏳ [BUFFER] Buffer not full yet - waiting for timer or batch size
  📦 [BUFFER] ======================================

✅ [SERVICE←BUFFER] Event buffered successfully (async)
⚡ [SERVICE] Returning immediately (non-blocking)
🎯 [SERVICE] ========================================

=== BACK TO ROUTE ===
✅ [ROUTE←SERVICE] RecordView successful
📤 [ROUTE→HISTORY] Updating watch history for user 42
✅ [ROUTE←HISTORY] Watch history updated
📤 [ROUTE→FRONTEND] Sending 200 OK response
🌐 [ROUTE] ============================================

=== 5 SECONDS LATER (Background Flush) ===
🔥 [BUFFER-FLUSH] ====================================
🔥 [BUFFER-FLUSH] FlushBatch triggered
📥 [BUFFER-FLUSH←REDIS] Reading batch from Redis (max 100 events)
📦 [BUFFER-FLUSH] Retrieved 23 events from Redis
✂️  [BUFFER-FLUSH→REDIS] Trimming processed events from Redis
💾 [BUFFER-FLUSH→DB] Starting batch UPSERT to watch_history
💾 [BUFFER-FLUSH] Processing event 1/23
💾 [BUFFER-FLUSH] Processing event 2/23
... (21 more events)
💾 [BUFFER-FLUSH] Processing event 23/23
✅ [BUFFER-FLUSH] Batch complete in 145ms
📊 [BUFFER-FLUSH] Results: 23 success, 0 errors (total: 23)
📈 [BUFFER-FLUSH] Stats: total_flushed=156, total_dropped=0, flush_count=7
🔥 [BUFFER-FLUSH] ====================================
```

---

## 🔍 What to Look For

### ✅ Healthy Flow
1. **Fast Response**: `📥 [FRONTEND←BACKEND] Response received in <10ms`
2. **Buffer Working**: `✅ [SERVICE←BUFFER] Event buffered successfully`
3. **Non-Blocking**: `⚡ [SERVICE] Returning immediately`
4. **Regular Flushes**: `🔥 [BUFFER-FLUSH]` every 5 seconds
5. **No Errors**: All checkmarks (✅), no ❌

### ⚠️ Warning Signs
1. **Slow Response**: Response time > 50ms
2. **Buffer Failures**: `⚠️ [SERVICE] Buffer.AddEvent failed`
3. **Fallback Mode**: `🔄 [SERVICE] Falling back to direct DB write`
4. **Circuit Open**: `⚠️ [ROUTE] Circuit breaker OPEN`
5. **Flush Errors**: `❌ [BUFFER-FLUSH]` errors

### ❌ Errors to Watch
1. **Frontend**: `❌ [FRONTEND] Failed to track progress`
2. **Route**: `❌ [ROUTE] RecordView failed`
3. **Buffer**: `❌ [BUFFER] Redis RPUSH failed`
4. **Database**: `❌ [SERVICE-DB] Database UPSERT failed`
5. **Flush**: `📊 [BUFFER-FLUSH] Results: X success, Y errors` (Y > 0)

---

## 🎯 Performance Benchmarks

### Expected Timings
- **Frontend→Backend**: <10ms
- **Route Processing**: <5ms
- **Buffer Add**: <2ms
- **Total User Latency**: <15ms
- **Background Flush**: 100-200ms per batch

### Buffer Metrics
- **Buffer Size**: 0-100 events (should flush at 100)
- **Flush Interval**: Every 5 seconds
- **Batch Size**: Typically 10-50 events per flush
- **Success Rate**: >99% (errors should be rare)

---

## 🧪 Testing the Flow

### Manual Test
1. **Start Backend**: `cd backend && go run main.go`
2. **Watch Logs**: Look for startup messages
3. **Play Video**: Open a video and play for 10+ seconds
4. **Observe Logs**: Watch the complete flow above
5. **Check Database**: After 5 seconds, query `watch_history`

### Test Queries
```sql
-- Check recent analytics
SELECT * FROM watch_history 
ORDER BY last_watched_at DESC 
LIMIT 10;

-- Check buffer stats (if using Redis)
redis-cli LLEN analytics:video_tracking_buffer
```

### Expected Behavior
1. **Every 10 seconds**: You'll see the full flow (frontend → backend → buffer)
2. **Every 5 seconds**: You'll see a background flush
3. **Response time**: Should be < 10ms
4. **Database**: Should update after each flush

---

## 🐛 Troubleshooting

### Issue: No logs appearing
**Check**: 
- Backend running? `go run main.go`
- Frontend dev server running? `npm run dev`
- Video playing (not paused)?

### Issue: "Buffer not available"
**Check**:
- Redis running? `redis-cli ping` should return `PONG`
- Config correct? Check Redis connection in backend config

### Issue: Direct DB writes instead of buffering
**Normal if**: Redis not configured (development mode)
**Problem if**: Redis is configured but failing

### Issue: Slow response times
**Check**:
- Database connection healthy?
- Redis connection healthy?
- Circuit breaker open? (Check `🛡️  [ROUTE]` logs)

### Issue: Events not flushing
**Check**:
- Look for `🔥 [BUFFER-FLUSH]` logs every 5 seconds
- Check buffer size in logs
- Verify background worker started on backend init

---

## 📝 Log Filtering

### Browser Console
```javascript
// Show only analytics logs
console.log = new Proxy(console.log, {
  apply(target, thisArg, args) {
    if (args[0]?.includes('[FRONTEND]')) {
      Reflect.apply(target, thisArg, args);
    }
  }
});
```

### Backend Logs
```bash
# Filter by layer
tail -f backend.log | grep "\[ROUTE\]"
tail -f backend.log | grep "\[SERVICE\]"
tail -f backend.log | grep "\[BUFFER\]"

# Filter by operation
tail -f backend.log | grep "FLUSH"

# Filter errors only
tail -f backend.log | grep "❌"
```

---

## 🎊 Success Criteria

Your analytics BRAID is healthy when you see:

✅ **Fast Responses**: <10ms frontend to backend  
✅ **Async Processing**: Events buffered immediately  
✅ **Regular Flushes**: Every 5 seconds or at 100 events  
✅ **High Success Rate**: >99% events flushed successfully  
✅ **No Errors**: Minimal ❌ in logs  
✅ **Data Persistence**: `watch_history` updating correctly  

---

## 📚 Related Files

- `frontend/src/lib/services/videoAnalytics.ts` - Frontend logging
- `backend/internal/routes/video_analytics_routes.go` - Route logging
- `backend/internal/services/video_analytics_service.go` - Service logging
- `backend/internal/services/analytics_buffer.go` - Buffer logging

---

## 🔄 Removing Logs (Production)

To remove verbose logging for production, search and remove lines containing:
- `log.Printf("🎬 [FRONTEND]`
- `log.Printf("🌐 [ROUTE]`
- `log.Printf("🎯 [SERVICE]`
- `log.Printf("📦 [BUFFER]`

Or use log level configuration to filter debug logs.

---

*Logging added for BRAID testing - November 27, 2025*  
*Temporary strategic logging for development and debugging*  
*Remove or adjust log levels for production deployment*

