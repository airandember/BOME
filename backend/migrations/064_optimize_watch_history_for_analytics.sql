-- Migration 064: Optimize watch_history for analytics tracking
-- Description: Enhance watch_history to handle all analytics (replacing video_views)
-- This eliminates the need for infinite video_views rows

-- Step 1: Add new columns for comprehensive tracking
ALTER TABLE watch_history 
    ADD COLUMN IF NOT EXISTS total_watch_time INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS view_count INTEGER DEFAULT 1,
    ADD COLUMN IF NOT EXISTS first_watched_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN IF NOT EXISTS session_id VARCHAR(255);

-- Step 2: Make user_id nullable to support anonymous sessions
ALTER TABLE watch_history 
    ALTER COLUMN user_id DROP NOT NULL;

-- Step 3: Drop old unique constraint
ALTER TABLE watch_history 
    DROP CONSTRAINT IF EXISTS watch_history_user_id_video_id_key;

-- Step 4: Create new unique indexes for both authenticated and anonymous users
-- For authenticated users: unique on (user_id, video_id)
CREATE UNIQUE INDEX IF NOT EXISTS idx_watch_history_user_video 
    ON watch_history(user_id, video_id) 
    WHERE user_id IS NOT NULL;

-- For anonymous users: unique on (session_id, video_id)
CREATE UNIQUE INDEX IF NOT EXISTS idx_watch_history_session_video 
    ON watch_history(session_id, video_id) 
    WHERE session_id IS NOT NULL;

-- Step 5: Add check constraint to ensure either user_id or session_id exists
ALTER TABLE watch_history 
    ADD CONSTRAINT watch_history_user_or_session 
    CHECK (user_id IS NOT NULL OR session_id IS NOT NULL);

-- Step 6: Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_watch_history_first_watched 
    ON watch_history(first_watched_at DESC);

CREATE INDEX IF NOT EXISTS idx_watch_history_total_watch_time 
    ON watch_history(total_watch_time DESC);

CREATE INDEX IF NOT EXISTS idx_watch_history_view_count 
    ON watch_history(view_count DESC);

-- Step 7: Add comments
COMMENT ON COLUMN watch_history.total_watch_time IS 'Cumulative seconds watched across all sessions';
COMMENT ON COLUMN watch_history.view_count IS 'Number of times user started watching this video';
COMMENT ON COLUMN watch_history.first_watched_at IS 'When user first discovered this video';
COMMENT ON COLUMN watch_history.session_id IS 'Session ID for anonymous users (NULL if authenticated)';

-- Step 8: Backfill existing records with sensible defaults
UPDATE watch_history 
SET 
    total_watch_time = COALESCE(last_position, 0),
    view_count = 1,
    first_watched_at = COALESCE(created_at, NOW())
WHERE total_watch_time IS NULL OR view_count IS NULL OR first_watched_at IS NULL;

-- Step 9: Create a view for easy analytics querying
CREATE OR REPLACE VIEW video_engagement_summary AS
SELECT 
    v.id AS video_id,
    v.title,
    v.duration,
    COUNT(DISTINCT COALESCE(wh.user_id::text, wh.session_id)) AS unique_viewers,
    SUM(wh.total_watch_time) AS total_watch_time,
    AVG(wh.progress_percentage) AS avg_completion_percentage,
    COUNT(*) FILTER (WHERE wh.completed = true) AS completions,
    COUNT(*) FILTER (WHERE wh.view_count > 1) AS repeat_viewers,
    MAX(wh.last_watched_at) AS last_viewed_at
FROM master_video_list v
LEFT JOIN watch_history wh ON wh.video_id = v.id
GROUP BY v.id, v.title, v.duration;

COMMENT ON VIEW video_engagement_summary IS 'Aggregated video engagement metrics from watch_history';

-- Success message
DO $$ 
BEGIN 
    RAISE NOTICE '✅ watch_history optimized for analytics tracking';
    RAISE NOTICE '✅ Ready to replace video_views with UPSERT pattern';
END $$;

