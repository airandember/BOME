-- Test View Count Sync Between watch_history and master_video_list
-- Run this to verify the trigger is working

-- ============================================
-- TEST 1: Check Current State
-- ============================================
SELECT 
    mvl.id,
    mvl.title,
    mvl.views AS "Card Shows",
    COALESCE(SUM(wh.view_count), 0) AS "Actual Views",
    CASE 
        WHEN mvl.views = COALESCE(SUM(wh.view_count), 0) THEN '✅ SYNCED'
        ELSE '❌ OUT OF SYNC'
    END AS "Status"
FROM master_video_list mvl
LEFT JOIN watch_history wh ON wh.video_id = mvl.id
WHERE EXISTS (SELECT 1 FROM watch_history WHERE video_id = mvl.id)
GROUP BY mvl.id, mvl.title, mvl.views
ORDER BY mvl.views DESC
LIMIT 10;

-- ============================================
-- TEST 2: Your Watched Videos
-- ============================================
SELECT 
    mvl.id,
    mvl.title,
    mvl.views AS "Card Shows",
    wh.view_count AS "Your Views",
    wh.last_watched_at AS "Last Watched"
FROM master_video_list mvl
JOIN watch_history wh ON wh.video_id = mvl.id
WHERE wh.user_id = 7342
ORDER BY wh.last_watched_at DESC;

-- ============================================
-- TEST 3: Simulate a View (Manual Test)
-- ============================================
-- Uncomment to test trigger:

-- -- Get current count
-- SELECT id, title, views FROM master_video_list WHERE id = 11042;
-- 
-- -- Add a view
-- INSERT INTO watch_history (user_id, video_id, view_count, last_position, progress_percentage, total_watch_time)
-- VALUES (9999, 11042, 1, 10, 0.5, 10)
-- ON CONFLICT (user_id, video_id) WHERE user_id IS NOT NULL
-- DO UPDATE SET view_count = watch_history.view_count + 1;
-- 
-- -- Check count updated (should be +1)
-- SELECT id, title, views FROM master_video_list WHERE id = 11042;

-- ============================================
-- TEST 4: Total Views Match
-- ============================================
SELECT 
    'master_video_list' AS "Source",
    SUM(views) AS "Total Views"
FROM master_video_list
WHERE EXISTS (SELECT 1 FROM watch_history WHERE video_id = master_video_list.id)

UNION ALL

SELECT 
    'watch_history' AS "Source",
    SUM(view_count) AS "Total Views"
FROM watch_history;

-- Both should match! If not, run backfill:
-- UPDATE master_video_list mvl
-- SET views = COALESCE((
--     SELECT SUM(wh.view_count)
--     FROM watch_history wh
--     WHERE wh.video_id = mvl.id
-- ), 0);

