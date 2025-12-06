-- Migration 070: DEPRECATED - DO NOT USE
-- This migration incorrectly set all views to 100 as a placeholder.
-- Use migration 071 and the restore_bunny_views Go script instead.

-- ⚠️ REMOVED: The UPDATE statement that was here set ALL videos to 100 views,
-- which destroyed the actual view counts.

-- The correct way to restore views is:
-- 1. Run: go run backend/cmd/restore_bunny_views/main.go
-- This fetches actual view counts from Bunny.net API and updates the database.

-- ORIGINAL PROBLEMATIC CODE (DO NOT UNCOMMENT):
-- UPDATE master_video_list mvl
-- SET views = GREATEST(mvl.views, 100)
-- WHERE mvl.id IN (
--     SELECT DISTINCT video_id 
--     FROM watch_history
-- )
-- AND mvl.views < 100
-- AND mvl.status = 'ready';

-- Verify current state (this is safe to run):
SELECT 
    id,
    title,
    views,
    (SELECT COUNT(*) FROM watch_history WHERE video_id = mvl.id) as watch_history_entries,
    (SELECT SUM(view_count) FROM watch_history WHERE video_id = mvl.id) as watch_history_total
FROM master_video_list mvl
WHERE status = 'ready'
ORDER BY views DESC
LIMIT 20;
