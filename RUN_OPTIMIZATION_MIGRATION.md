# Running the Analytics Optimization Migration

## Step 1: Run the Migration

Connect to your PostgreSQL database and run:

```sql
\i backend/migrations/064_optimize_watch_history_for_analytics.sql
```

Or using psql command line:
```bash
psql -d bome_db -f backend/migrations/064_optimize_watch_history_for_analytics.sql
```

## Step 2: Verify Migration Success

Check that the new columns exist:
```sql
\d watch_history

-- Should show new columns:
-- - total_watch_time
-- - view_count  
-- - first_watched_at
-- - session_id
```

Check the new indexes:
```sql
SELECT indexname FROM pg_indexes WHERE tablename = 'watch_history';

-- Should include:
-- - idx_watch_history_user_video
-- - idx_watch_history_session_video
```

## Step 3: Rebuild and Restart Backend

```powershell
# Navigate to backend directory
cd S:\AirEmber\BOME\BOME\backend

# Build
go build -o bome-backend.exe main.go

# Kill old process
$proc = netstat -ano | findstr ":8080" | findstr "LISTENING" | ForEach-Object { $_ -split '\s+' | Select-Object -Last 1 }
if ($proc) { Stop-Process -Id $proc -Force }

# Start new backend
go run main.go
```

## Step 4: Test the Optimization

Watch a video and check the database:

```sql
-- Should show ONE row per user+video (updated in place)
SELECT * FROM watch_history ORDER BY last_watched_at DESC LIMIT 10;

-- Should NOT be creating new video_views rows anymore
SELECT COUNT(*) FROM video_views WHERE created_at > NOW() - INTERVAL '1 minute';
-- (This count should stay at 0 or very low)
```

## Expected Behavior

### Before Optimization ❌
```
User watches video for 5 minutes:
- 30 new rows in video_views (1 every 10 seconds)
```

### After Optimization ✅
```
User watches video for 5 minutes:
- 1 row in watch_history (updated 30 times)
- Storage: 30x reduction!
```

## What Changed

1. **watch_history**: Now handles all tracking with UPSERT
2. **video_views**: No longer being written to (can be archived)
3. **Queries**: Updated to use watch_history instead of video_views
4. **Storage**: Fixed size (users × videos) instead of infinite growth

## Cleanup (Optional - Later)

Once you're confident the new system works:

```sql
-- Rename video_views for archival
ALTER TABLE video_views RENAME TO video_views_archive;

-- Or drop it entirely if you don't need historical data
-- DROP TABLE video_views;
```

---
**Status**: Ready to run!  
**Priority**: 🔴 CRITICAL - Stops database explosion

