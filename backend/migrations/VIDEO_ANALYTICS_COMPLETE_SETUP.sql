-- ============================================
-- Video Analytics Complete Setup
-- ============================================
-- This is a complete, all-in-one migration file
-- Copy and paste this entire file into pgAdmin or your SQL tool
-- ============================================

-- STEP 1: Create video_views table
-- ============================================
CREATE TABLE IF NOT EXISTS video_views (
    id SERIAL PRIMARY KEY,
    video_id INTEGER NOT NULL REFERENCES master_video_list(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    session_id VARCHAR(255),
    ip_address INET,
    watched_duration INTEGER NOT NULL DEFAULT 0,
    watched_percentage DECIMAL(5,2) NOT NULL DEFAULT 0.0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT video_views_check_percentage CHECK (watched_percentage >= 0 AND watched_percentage <= 100),
    CONSTRAINT video_views_check_duration CHECK (watched_duration >= 0),
    CONSTRAINT video_views_check_user_or_session CHECK (user_id IS NOT NULL OR session_id IS NOT NULL)
);

-- STEP 2: Create watch_history table
-- ============================================
CREATE TABLE IF NOT EXISTS watch_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    video_id INTEGER NOT NULL REFERENCES master_video_list(id) ON DELETE CASCADE,
    last_position INTEGER NOT NULL DEFAULT 0,
    progress_percentage DECIMAL(5,2) NOT NULL DEFAULT 0.0,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    last_watched_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(user_id, video_id),
    
    CONSTRAINT watch_history_check_percentage CHECK (progress_percentage >= 0 AND progress_percentage <= 100),
    CONSTRAINT watch_history_check_position CHECK (last_position >= 0)
);

-- STEP 3: Create indexes
-- ============================================
CREATE INDEX IF NOT EXISTS idx_video_views_video_id ON video_views(video_id);
CREATE INDEX IF NOT EXISTS idx_video_views_user_id ON video_views(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_video_views_session_id ON video_views(session_id) WHERE session_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_video_views_created_at ON video_views(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_video_views_video_created ON video_views(video_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_video_views_video_user_session ON video_views(video_id, user_id, session_id);

CREATE INDEX IF NOT EXISTS idx_watch_history_user_id ON watch_history(user_id);
CREATE INDEX IF NOT EXISTS idx_watch_history_video_id ON watch_history(video_id);
CREATE INDEX IF NOT EXISTS idx_watch_history_last_watched ON watch_history(last_watched_at DESC);
CREATE INDEX IF NOT EXISTS idx_watch_history_user_incomplete ON watch_history(user_id, completed) WHERE completed = FALSE;

-- STEP 4: Create trigger function for view count sync
-- ============================================
CREATE OR REPLACE FUNCTION update_master_video_views()
RETURNS TRIGGER 
LANGUAGE plpgsql
AS $$
DECLARE
    v_video_id INTEGER;
BEGIN
    v_video_id := NEW.video_id;
    
    UPDATE master_video_list
    SET 
        views = (
            SELECT COUNT(DISTINCT COALESCE(user_id::text, session_id))
            FROM video_views
            WHERE video_id = v_video_id
        ),
        total_watch_time = (
            SELECT COALESCE(SUM(watched_duration), 0)
            FROM video_views
            WHERE video_id = v_video_id
        ),
        average_watch_time = (
            SELECT COALESCE(AVG(watched_duration), 0)::integer
            FROM video_views
            WHERE video_id = v_video_id
        ),
        updated_at = CURRENT_TIMESTAMP
    WHERE id = v_video_id;
    
    RETURN NEW;
END;
$$;

-- STEP 5: Create trigger
-- ============================================
DROP TRIGGER IF EXISTS trigger_sync_master_video_views ON video_views;
CREATE TRIGGER trigger_sync_master_video_views
    AFTER INSERT ON video_views
    FOR EACH ROW
    EXECUTE FUNCTION update_master_video_views();

-- STEP 6: Backfill existing data (if any)
-- ============================================
UPDATE master_video_list mvl
SET 
    views = COALESCE(view_counts.unique_views, 0),
    total_watch_time = COALESCE(view_counts.total_watch_time, 0),
    average_watch_time = COALESCE(view_counts.avg_watch_time, 0),
    updated_at = CURRENT_TIMESTAMP
FROM (
    SELECT 
        video_id,
        COUNT(DISTINCT COALESCE(user_id::text, session_id)) as unique_views,
        SUM(watched_duration) as total_watch_time,
        AVG(watched_duration)::integer as avg_watch_time
    FROM video_views
    GROUP BY video_id
) view_counts
WHERE mvl.id = view_counts.video_id;

-- STEP 7: Add comments
-- ============================================
COMMENT ON TABLE video_views IS 'Individual video view events for analytics';
COMMENT ON TABLE watch_history IS 'User watch progress for resume functionality';
COMMENT ON TRIGGER trigger_sync_master_video_views ON video_views IS 'Automatically syncs master_video_list.views, total_watch_time, and average_watch_time from video_views table';
COMMENT ON FUNCTION update_master_video_views() IS 'Updates master_video_list aggregated view metrics from video_views table. Counts unique viewers by user_id or session_id.';

-- ============================================
-- Setup Complete!
-- ============================================
-- Run these queries to verify:
--
-- 1. Check tables exist:
--    SELECT table_name FROM information_schema.tables 
--    WHERE table_name IN ('video_views', 'watch_history');
--
-- 2. Check trigger exists:
--    SELECT trigger_name FROM information_schema.triggers 
--    WHERE trigger_name = 'trigger_sync_master_video_views';
--
-- 3. Test the sync (replace video_id with a real one):
--    INSERT INTO video_views (video_id, session_id, watched_duration, watched_percentage)
--    VALUES (1, 'test-' || gen_random_uuid(), 120, 75.5);
--    
--    SELECT id, views, total_watch_time, average_watch_time 
--    FROM master_video_list WHERE id = 1;
-- ============================================

