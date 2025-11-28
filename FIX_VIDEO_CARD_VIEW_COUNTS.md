# Fix Video Card View Counts 🎬

## Problem
Video cards showing old/wrong view counts because `master_video_list.views` isn't synced with `watch_history`.

## Solution
Create a database trigger to automatically update `master_video_list.views` whenever `watch_history` changes.

---

## Deploy Steps

### Step 1: Run Migration
```bash
cd S:\AirEmber\BOME\BOME\backend
psql -d bome_db -f migrations/068_sync_master_video_views_from_watch_history.sql
```

**Expected Output**:
```
DROP TRIGGER
DROP FUNCTION
CREATE FUNCTION
CREATE TRIGGER
UPDATE X  (backfill existing data)
SELECT 10 (verification query)
```

### Step 2: Verify Sync
```sql
-- Check that views are now synced
SELECT 
    mvl.id,
    mvl.title,
    mvl.views AS card_views,
    COALESCE(SUM(wh.view_count), 0) AS actual_views
FROM master_video_list mvl
LEFT JOIN watch_history wh ON wh.video_id = mvl.id
WHERE wh.video_id IS NOT NULL
GROUP BY mvl.id, mvl.title, mvl.views
ORDER BY mvl.views DESC
LIMIT 10;
```

**Expected**: `card_views` = `actual_views` ✅

### Step 3: Test Real-Time Update
```bash
# 1. Note current view count for video 11042
psql -d bome_db -c "SELECT id, title, views FROM master_video_list WHERE id = 11042;"

# 2. Manually insert a view (simulate analytics)
psql -d bome_db -c "
INSERT INTO watch_history (user_id, video_id, view_count, last_position, progress_percentage)
VALUES (7342, 11042, 1, 10, 0.5)
ON CONFLICT (user_id, video_id) WHERE user_id IS NOT NULL
DO UPDATE SET view_count = watch_history.view_count + 1;
"

# 3. Check view count updated automatically
psql -d bome_db -c "SELECT id, title, views FROM master_video_list WHERE id = 11042;"
-- Should be +1! ✅
```

### Step 4: Refresh Frontend
```
Hard refresh: Ctrl+Shift+R
Check video cards - view counts should update!
```

---

## How It Works

### Trigger Flow
```
1. watch_history INSERT/UPDATE
   ↓
2. Trigger fires: update_master_video_views_trigger
   ↓
3. Function runs: update_master_video_views()
   ↓
4. master_video_list.views = SUM(watch_history.view_count)
   ↓
5. Frontend video cards show updated count ✅
```

### Code
```sql
-- Trigger function
CREATE OR REPLACE FUNCTION update_master_video_views()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE master_video_list
    SET views = (
        SELECT COALESCE(SUM(view_count), 0)
        FROM watch_history
        WHERE video_id = NEW.video_id
    )
    WHERE id = NEW.video_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger
CREATE TRIGGER update_master_video_views_trigger
AFTER INSERT OR UPDATE OF view_count ON watch_history
FOR EACH ROW
EXECUTE FUNCTION update_master_video_views();
```

---

## What Gets Updated

### When Analytics Tracks a View
```
User watches video 11042:

watch_history:
  view_count: 1 → 2 (incremented after 30+ min)
  ↓
master_video_list:
  views: 1 → 2 (automatically updated by trigger) ✅
  ↓
Frontend video card:
  Shows "2 views" ✅
```

### Backfill Existing Data
The migration also backfills:
```sql
-- Update ALL videos with current watch_history totals
UPDATE master_video_list mvl
SET views = (
    SELECT COALESCE(SUM(view_count), 0)
    FROM watch_history
    WHERE video_id = mvl.id
);
```

So videos you've already watched will show correct counts immediately!

---

## Testing Scenarios

### Scenario 1: Fresh Video
```
1. Video 11058 has 0 views in master_video_list
2. User watches video (first time)
3. watch_history: view_count = 1
4. Trigger fires
5. master_video_list.views = 1 ✅
6. Video card shows "1 view" ✅
```

### Scenario 2: Multiple Users
```
User A watches video 11042:
  - watch_history: user_id=7342, view_count=2
  
User B watches video 11042:
  - watch_history: user_id=9999, view_count=1
  
Trigger calculates:
  - master_video_list.views = SUM(2 + 1) = 3 ✅
  
Video card shows: "3 views" ✅
```

### Scenario 3: Same User, New Session
```
User watches video at 8:00 AM:
  - view_count = 1
  - master_video_list.views = 1
  
User watches again at 9:00 AM (30+ min later):
  - view_count = 2 (incremented by session logic)
  - Trigger fires
  - master_video_list.views = 2 ✅
  
Video card updates: "2 views" ✅
```

---

## Verification Queries

### Check Sync Status
```sql
-- Are views synced?
SELECT 
    'master_video_list' AS source,
    SUM(views) AS total_views
FROM master_video_list
WHERE EXISTS (SELECT 1 FROM watch_history WHERE video_id = master_video_list.id)

UNION ALL

SELECT 
    'watch_history' AS source,
    SUM(view_count) AS total_views
FROM watch_history;

-- Both should match! ✅
```

### Check Individual Videos
```sql
-- Your watched videos
SELECT 
    mvl.id,
    mvl.title,
    mvl.views AS card_shows,
    wh.view_count AS you_watched,
    (SELECT SUM(view_count) FROM watch_history WHERE video_id = mvl.id) AS total_views
FROM master_video_list mvl
JOIN watch_history wh ON wh.video_id = mvl.id
WHERE wh.user_id = 7342
ORDER BY wh.last_watched_at DESC;
```

### Check Trending Videos
```sql
-- Should match trending section
SELECT 
    id,
    title,
    views,
    created_at
FROM master_video_list
WHERE views > 0
ORDER BY views DESC
LIMIT 10;
```

---

## Backend Protection

The backend already protects against Bunny overwriting our data (line 986-992 in routes.go):

```go
// Only sync view count from Bunny if we don't have detailed analytics
var hasAnalytics bool
err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM watch_history WHERE video_id = $1 LIMIT 1)", dbVideo.ID).Scan(&hasAnalytics)
if err == nil && !hasAnalytics && dbVideo.ViewCount != bunnyVideo.Views {
    // No analytics data yet, safe to sync from Bunny
    updates["view_count"] = bunnyVideo.Views
}
```

**Result**: 
- ✅ New videos: Sync from Bunny
- ✅ Tracked videos: Use watch_history (our data)
- ✅ Bunny can't overwrite our analytics

---

## Benefits

### Before (Broken)
```
master_video_list.views = Bunny.net static data
Video cards show: 0 views (never updates)
```

### After (Fixed)
```
master_video_list.views = SUM(watch_history.view_count)
Video cards show: Real-time view counts ✅
Automatically updates: Every time someone watches ✅
```

---

## Rollback (If Needed)

```sql
-- Remove trigger
DROP TRIGGER IF EXISTS update_master_video_views_trigger ON watch_history;
DROP FUNCTION IF EXISTS update_master_video_views();
```

---

## Summary

**Problem**: Video cards not showing real view counts
**Root Cause**: `master_video_list.views` not synced with `watch_history`
**Solution**: Database trigger for automatic sync
**Deploy**: Run migration 068
**Result**: Video cards show real-time view counts! 🎉

---

## Quick Deploy

```bash
# One command
cd S:\AirEmber\BOME\BOME\backend && psql -d bome_db -f migrations/068_sync_master_video_views_from_watch_history.sql

# Then refresh frontend
# Ctrl+Shift+R
```

**Done!** Video cards will now show accurate, real-time view counts based on actual user viewing behavior. 🚀

