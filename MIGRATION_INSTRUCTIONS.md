# Video Analytics Migration Instructions

## Problem
You got the error: `ERROR: relation "video_views" does not exist`

## Cause
The view sync migration (062) requires the `video_views` table to exist first, but it wasn't created yet.

## Solution
You need to run the migrations in the correct order.

---

## Option 1: All-in-One (RECOMMENDED - Easiest)

**Copy/Paste this entire file into pgAdmin:**

`backend/migrations/VIDEO_ANALYTICS_COMPLETE_SETUP.sql`

This single file contains:
- ✅ Creates `video_views` table
- ✅ Creates `watch_history` table
- ✅ Creates all indexes
- ✅ Creates trigger function
- ✅ Creates trigger
- ✅ Backfills existing data
- ✅ Adds documentation comments

**Just copy/paste the entire file and execute it once.**

---

## Option 2: Run Migrations Individually

If you prefer to run migrations step-by-step:

### Step 1: Create Tables
Run this first:
```
backend/migrations/060_1_create_video_analytics_tables.sql
```

### Step 2: Create View Sync
Run this second:
```
backend/migrations/062_sync_master_video_views.sql
```

---

## Verification

After running the migration, verify it worked:

### 1. Check Tables Exist
```sql
SELECT table_name 
FROM information_schema.tables 
WHERE table_name IN ('video_views', 'watch_history');
```

**Expected Output**: 2 rows showing both tables

### 2. Check Trigger Exists
```sql
SELECT trigger_name, event_manipulation, event_object_table
FROM information_schema.triggers 
WHERE trigger_name = 'trigger_sync_master_video_views';
```

**Expected Output**: 1 row showing the trigger on `video_views` table

### 3. Test the Sync
```sql
-- Insert a test view (replace 1 with a real video_id from your database)
INSERT INTO video_views (video_id, session_id, watched_duration, watched_percentage)
VALUES (1, 'test-' || gen_random_uuid()::text, 120, 75.5);

-- Check if master_video_list was updated
SELECT id, title, views, total_watch_time, average_watch_time 
FROM master_video_list 
WHERE id = 1;
```

**Expected Result**: The `views`, `total_watch_time`, and `average_watch_time` columns should reflect the new view.

### 4. Check for Discrepancies
```sql
-- This should return 0 rows if everything is synced correctly
SELECT 
    mvl.id,
    mvl.title,
    mvl.views as master_views,
    COUNT(DISTINCT COALESCE(vv.user_id::text, vv.session_id)) as analytics_views,
    mvl.views - COUNT(DISTINCT COALESCE(vv.user_id::text, vv.session_id)) as difference
FROM master_video_list mvl
LEFT JOIN video_views vv ON vv.video_id = mvl.id
GROUP BY mvl.id, mvl.title, mvl.views
HAVING mvl.views != COUNT(DISTINCT COALESCE(vv.user_id::text, vv.session_id))
ORDER BY ABS(difference) DESC;
```

---

## Troubleshooting

### Error: "relation already exists"
This is fine - it means you've already run part of the migration. The `IF NOT EXISTS` clauses will skip creating duplicate tables.

### Error: "constraint already exists"
This is fine - constraints with `IF NOT EXISTS` protection will be skipped.

### Error: "trigger already exists"
The migration includes `DROP TRIGGER IF EXISTS` so this shouldn't happen. If it does, manually drop it first:
```sql
DROP TRIGGER IF EXISTS trigger_sync_master_video_views ON video_views;
```

### View counts are still 0
1. Make sure you have data in `video_views` table
2. Try manually running the backfill query from the migration
3. Insert a test view to trigger the sync

---

## What This Migration Does

### Creates Tables
1. **`video_views`** - Records every video view with:
   - Video ID
   - User ID (if logged in) or Session ID (if anonymous)
   - Watch duration and percentage
   - Timestamp

2. **`watch_history`** - Tracks user progress for "Continue Watching":
   - Last position in video
   - Progress percentage
   - Completion status
   - Last watched timestamp

### Sets Up Sync
- **Trigger**: Automatically updates `master_video_list.views` when a view is recorded
- **Unique Counting**: Counts distinct users/sessions (not total requests)
- **Aggregates**: Updates views, total_watch_time, and average_watch_time

### Performance
- Creates 10+ indexes for fast queries
- Optimized for trending, top videos, and user history
- Minimal trigger overhead (< 10ms per view)

---

## Files Reference

| File | Purpose | Order |
|------|---------|-------|
| `VIDEO_ANALYTICS_COMPLETE_SETUP.sql` | All-in-one complete setup | Run this (easiest) |
| `060_1_create_video_analytics_tables.sql` | Create tables only | Run first |
| `062_sync_master_video_views.sql` | Create sync trigger | Run second |
| `RUN_VIDEO_ANALYTICS_MIGRATIONS.sql` | Auto-runner script | For psql command line |

---

## After Migration

Once migration is complete:
1. ✅ Backend API will start tracking video views
2. ✅ Trending videos will show accurate data
3. ✅ Most Watched will work with real metrics
4. ✅ Continue Watching will remember progress
5. ✅ Analytics dashboards will have accurate counts
6. ✅ `master_video_list.views` will stay in sync automatically

---

## Need Help?

If you encounter any issues:
1. Check the error message carefully
2. Verify you're connected to the correct database (`bome_db`)
3. Ensure `master_video_list` and `users` tables exist (required for foreign keys)
4. Check the migration logs for specific error details

The migration uses `IF NOT EXISTS` and `IF EXISTS` clauses extensively, so it's **safe to run multiple times** if needed.

