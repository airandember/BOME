-- Migration 062: Sync master_video_list.views with video_views table
-- This ensures the legacy views column stays in sync with our detailed analytics

-- Step 1: Create a function to update master_video_list.views from video_views
CREATE OR REPLACE FUNCTION update_master_video_views()
RETURNS TRIGGER 
LANGUAGE plpgsql
AS $$
DECLARE
    v_video_id INTEGER;
BEGIN
    -- Get the video_id from the NEW record
    v_video_id := NEW.video_id;
    
    -- Update the views count in master_video_list
    -- Count distinct session_id OR user_id as unique views
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

-- Step 2: Create a trigger to automatically sync on new views
DROP TRIGGER IF EXISTS trigger_sync_master_video_views ON video_views;
CREATE TRIGGER trigger_sync_master_video_views
    AFTER INSERT ON video_views
    FOR EACH ROW
    EXECUTE FUNCTION update_master_video_views();

-- Step 3: Backfill existing data - sync all current video_views to master_video_list
-- This ensures historical data is accurate
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

-- Step 4: Create an index for performance on the sync query
CREATE INDEX IF NOT EXISTS idx_video_views_video_user_session 
    ON video_views(video_id, user_id, session_id);

-- Step 5: Add a comment explaining the sync
COMMENT ON TRIGGER trigger_sync_master_video_views ON video_views IS 
    'Automatically syncs master_video_list.views, total_watch_time, and average_watch_time from video_views table';

COMMENT ON FUNCTION update_master_video_views() IS 
    'Updates master_video_list aggregated view metrics from video_views table. Counts unique viewers by user_id or session_id.';

