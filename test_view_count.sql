-- Test View Count Session Logic
-- Run this in your PostgreSQL client to test view_count behavior

-- ============================================
-- TEST 1: Check Current State
-- ============================================
SELECT 
    v.title,
    wh.video_id,
    wh.view_count,
    wh.total_watch_time,
    wh.last_watched_at,
    EXTRACT(EPOCH FROM (NOW() - wh.last_watched_at))/60 AS minutes_since_last_watch,
    CASE 
        WHEN wh.last_watched_at < NOW() - INTERVAL '30 minutes' 
        THEN '🆕 NEW SESSION (will increment view_count)'
        ELSE '⏳ SAME SESSION (will NOT increment)'
    END AS next_view_status
FROM watch_history wh
JOIN master_video_list v ON wh.video_id = v.id
WHERE wh.user_id = 7342
ORDER BY wh.last_watched_at DESC;

-- ============================================
-- TEST 2: Simulate Old Session (For Quick Testing)
-- ============================================
-- Use this to test without waiting 30 minutes
-- Uncomment and run, then watch the video again

-- UPDATE watch_history 
-- SET last_watched_at = NOW() - INTERVAL '31 minutes'
-- WHERE user_id = 7342 AND video_id = 11042;

-- SELECT 
--     video_id, 
--     view_count, 
--     last_watched_at,
--     '✅ Ready for new session test!' AS status
-- FROM watch_history 
-- WHERE user_id = 7342 AND video_id = 11042;

-- ============================================
-- TEST 3: View Count Verification
-- ============================================
SELECT 
    v.title,
    wh.view_count AS current_view_count,
    wh.total_watch_time AS total_seconds_watched,
    wh.total_watch_time / 60 AS total_minutes_watched,
    wh.first_watched_at,
    wh.last_watched_at,
    AGE(wh.last_watched_at, wh.first_watched_at) AS time_span
FROM watch_history wh
JOIN master_video_list v ON wh.video_id = v.id
WHERE wh.user_id = 7342
ORDER BY wh.view_count DESC, wh.last_watched_at DESC;

-- ============================================
-- TEST 4: Most Watched Videos (Should Work Now!)
-- ============================================
SELECT 
    v.title,
    SUM(wh.view_count) AS total_views,
    COUNT(DISTINCT wh.user_id) AS unique_viewers,
    SUM(wh.total_watch_time) AS total_watch_time_seconds,
    SUM(wh.total_watch_time) / 60 AS total_watch_time_minutes
FROM watch_history wh
JOIN master_video_list v ON wh.video_id = v.id
GROUP BY v.id, v.title
ORDER BY total_views DESC
LIMIT 10;

-- ============================================
-- TEST 5: Your Specific Videos
-- ============================================
SELECT 
    CASE 
        WHEN wh.video_id = 11042 THEN '🎥 Video 11042'
        WHEN wh.video_id = 11058 THEN '🎥 Video 11058'
        ELSE '🎥 Other'
    END AS video,
    wh.view_count,
    wh.total_watch_time,
    wh.last_watched_at,
    EXTRACT(EPOCH FROM (NOW() - wh.last_watched_at))/60 AS minutes_ago,
    CASE 
        WHEN wh.last_watched_at < NOW() - INTERVAL '30 minutes' 
        THEN '✅ Next view = NEW SESSION'
        ELSE '⏳ Next view = SAME SESSION'
    END AS next_view
FROM watch_history wh
WHERE wh.user_id = 7342 
  AND wh.video_id IN (11042, 11058)
ORDER BY wh.video_id;

-- ============================================
-- EXPECTED RESULTS
-- ============================================
-- After first watch:
--   view_count = 1
--   total_watch_time = actual seconds watched
--
-- After second watch (within 30 min):
--   view_count = 1 (unchanged)
--   total_watch_time = increased
--
-- After third watch (30+ min later):
--   view_count = 2 (incremented!)
--   total_watch_time = increased
-- ============================================

