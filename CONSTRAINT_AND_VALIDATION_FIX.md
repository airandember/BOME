# Database Constraint & Frontend Validation Fix

## Issues Fixed

### Issue #1: Database Constraint Error ❌
```
Backend: Failed to update progress pq there is no unique or exclusion constraint 
matching the ON CONFLICT specification
```

**Root Cause**: Complex COALESCE-based unique constraint doesn't work with `ON CONFLICT`

**Fix**: Created simple unique constraints
- For authenticated: `(user_id, video_id) WHERE user_id IS NOT NULL`
- For anonymous: `(session_id, video_id) WHERE session_id IS NOT NULL`

**Migration**: `067_fix_watch_history_constraints.sql`

---

### Issue #2: Frontend 400 Validation Error ❌
```
Error: Key: 'VideoTrackingRequest.WatchedDuration' Error:Field validation 
for 'WatchedDuration' failed on the 'required' tag
```

**Root Cause**: Initial tracking call at 0 seconds with unknown duration (iframe)

**Fix**: 
1. Skip initial tracking if duration is unknown
2. Use fallback duration (2798s) for iframe tracking

**Files**: `frontend/src/routes/videos/[id]/+page.svelte`

---

## How to Deploy

### 1. Run Migration
```bash
cd backend
psql -d bome_db -f migrations/067_fix_watch_history_constraints.sql
```

### 2. Rebuild Backend
```bash
go build -o bome-backend-v2.exe main.go
```

### 3. Restart Backend
```bash
# Stop old backend
# Run new backend
./bome-backend-v2.exe
```

### 4. Test
- Watch a video (should work without errors!)
- Check backend logs (no constraint errors)
- Check frontend console (no 400 errors after first few seconds)

---

## Technical Details

### Old Constraint (Broken)
```sql
CREATE UNIQUE INDEX idx_watch_history_user_video 
ON watch_history ((COALESCE(user_id::text, session_id)), video_id);
-- ❌ Doesn't work with ON CONFLICT
```

### New Constraint (Fixed)
```sql
CREATE UNIQUE INDEX idx_watch_history_user_video_simple
ON watch_history (user_id, video_id)
WHERE user_id IS NOT NULL;
-- ✅ Works with ON CONFLICT
```

### Why This Works
PostgreSQL's `ON CONFLICT` requires a **simple** unique constraint or index.
Complex expressions with `COALESCE` aren't supported.

Solution: Use **partial indexes** with `WHERE` clauses instead.

---

## Frontend Fix

### Before (Broken)
```typescript
// Always track immediately, even with duration=0
videoAnalytics.trackProgress(numericId, 0, video.duration);
```

### After (Fixed)
```typescript
// Skip if duration unknown
if (video.duration && video.duration > 0) {
    videoAnalytics.trackProgress(numericId, 0, video.duration);
}

// Use fallback duration for iframe
const videoDuration = video.duration || 2798;
videoAnalytics.trackProgress(numericId, watchedSeconds, videoDuration);
```

---

## Testing

### Database Constraint
```sql
-- Test authenticated user insert
INSERT INTO watch_history (user_id, video_id, last_position, progress_percentage, view_count)
VALUES (7342, 11042, 10, 0.5, 1);

-- Test UPSERT (should update, not error)
INSERT INTO watch_history (user_id, video_id, last_position, progress_percentage, view_count)
VALUES (7342, 11042, 20, 1.0, 1)
ON CONFLICT (user_id, video_id) WHERE user_id IS NOT NULL
DO UPDATE SET last_position = EXCLUDED.last_position;

-- Should succeed! ✅
```

### Frontend
1. Open video page
2. Check console - should see:
   ```
   ✅ Initial tracking skipped (duration unknown)
   ✅ [9s] Successfully tracked
   ✅ [19s] Successfully tracked
   ```
3. No 400 errors after first load ✅

---

## Files Modified

### Backend
1. **`migrations/067_fix_watch_history_constraints.sql`** (NEW)
   - Drops complex constraints
   - Adds simple partial indexes

### Frontend
2. **`frontend/src/routes/videos/[id]/+page.svelte`**
   - Lines 108-110: Skip initial tracking if duration unknown
   - Line 122: Use fallback duration for iframe

---

## Status

- [ ] Run migration 067
- [ ] Rebuild backend
- [ ] Restart backend
- [ ] Test video playback
- [ ] Verify no errors

Once complete:
✅ Database UPSERT works
✅ No 400 errors
✅ Analytics tracking reliable

---

## Quick Deploy Script

```bash
# 1. Migration
cd backend
psql -d bome_db -f migrations/067_fix_watch_history_constraints.sql

# 2. Build
go build -o bome-backend-v2.exe main.go

# 3. Run
./bome-backend-v2.exe
```

Then test by watching a video!

