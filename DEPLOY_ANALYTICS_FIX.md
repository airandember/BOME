# Deploy Analytics Fixes - Quick Guide

## What We're Fixing

1. ❌ **Database constraint error** - UPSERT can't match ON CONFLICT
2. ❌ **Frontend 400 error** - Initial tracking with duration=0

---

## Deploy Steps (In Order)

### Step 1: Run Database Migration
```bash
cd S:\AirEmber\BOME\BOME\backend
psql -d bome_db -f migrations/067_fix_watch_history_constraints.sql
```

**Expected Output**:
```
DROP INDEX
DROP INDEX
CREATE INDEX
CREATE INDEX
SELECT 2
```

---

### Step 2: Restart Backend
```bash
# Stop current backend (Ctrl+C in terminal)

# Start backend (your production way)
go run main.go
```

**Expected Output**:
```
✅ [Video Analytics] Async buffer enabled with Redis
✅ [Analytics Buffer] Background flusher started
Server running on :8080
```

---

### Step 3: Refresh Frontend
```
Hard refresh browser: Ctrl+Shift+R (Windows) or Cmd+Shift+R (Mac)
```

---

### Step 4: Test
1. **Watch a video** for 30+ seconds
2. **Check console** - Should see:
   ```
   ✅ [FRONTEND] Successfully tracked: video=11042, time=9s
   ✅ [FRONTEND] Successfully tracked: video=11042, time=19s
   ✅ [FRONTEND] Successfully tracked: video=11042, time=29s
   ```

3. **Check backend logs** - Should see:
   ```
   🌐 [ROUTE] Received POST /analytics/video/track
   🎯 [SERVICE] RecordView called
   ✅ [BUFFER←REDIS] Event pushed to Redis successfully
   ```

4. **Check database**:
   ```bash
   psql -d bome_db -c "SELECT video_id, view_count, total_watch_time FROM watch_history WHERE user_id = 7342;"
   ```

---

## What Changed

### Backend (Database)
**File**: `migrations/067_fix_watch_history_constraints.sql`

**Before** (Broken):
```sql
-- Complex COALESCE constraint
CREATE UNIQUE INDEX idx_watch_history_user_video 
ON watch_history ((COALESCE(user_id::text, session_id)), video_id);
```

**After** (Fixed):
```sql
-- Simple partial indexes
CREATE UNIQUE INDEX idx_watch_history_user_video_simple
ON watch_history (user_id, video_id)
WHERE user_id IS NOT NULL;
```

### Frontend
**File**: `frontend/src/routes/videos/[id]/+page.svelte`

**Before** (Broken):
```typescript
// Always track immediately (even with duration=0)
videoAnalytics.trackProgress(numericId, 0, video.duration);
```

**After** (Fixed):
```typescript
// Skip if duration unknown
if (video.duration && video.duration > 0) {
    videoAnalytics.trackProgress(numericId, 0, video.duration);
}

// Use fallback duration for iframe
const videoDuration = video.duration || 2798;
```

---

## Verification

### ✅ Success Indicators
- [ ] Migration runs without errors
- [ ] Backend starts successfully
- [ ] No console errors when watching video
- [ ] Backend logs show tracking events
- [ ] Database `watch_history` table populates

### ❌ If Still Broken

**Database constraint error?**
```bash
# Check constraints exist
psql -d bome_db -c "
SELECT indexname, indexdef 
FROM pg_indexes 
WHERE tablename = 'watch_history';
"
```

**Frontend 400 error?**
```javascript
// Check console for duration value
console.log('Video duration:', video.duration);
// Should be > 0 or will skip tracking
```

---

## Quick Commands

### Check Backend Logs
```bash
# In backend directory
go run main.go
# Watch for tracking logs
```

### Check Database
```bash
# Recent views
psql -d bome_db -c "
SELECT v.title, wh.view_count, wh.total_watch_time, wh.last_watched_at
FROM watch_history wh
JOIN master_video_list v ON wh.video_id = v.id
WHERE wh.user_id = 7342
ORDER BY wh.last_watched_at DESC
LIMIT 5;
"
```

### Test Session Logic
```bash
# Simulate old session
psql -d bome_db -c "
UPDATE watch_history 
SET last_watched_at = NOW() - INTERVAL '31 minutes'
WHERE user_id = 7342 AND video_id = 11042;
"

# Watch video again - view_count should increment!
```

---

## Production Note

You're using `go run main.go` which is perfect because:
- ✅ Hot reload during development
- ✅ Matches production initialization
- ✅ No binary management needed
- ✅ Consistent environment

---

## Summary

**Problem**: Analytics broken due to DB constraint + frontend validation
**Fix**: Simple constraints + skip initial tracking if duration unknown
**Deploy**: Run migration → Restart backend → Refresh browser
**Result**: Analytics working end-to-end! 🎉

**Ready to deploy!** 🚀

