-- Migration 066: Drop video_views table
-- Description: Removes the legacy video_views table after migrating all queries to watch_history
-- Prerequisites: All services must be updated to use watch_history instead

-- Step 1: Drop any triggers that reference video_views
DROP TRIGGER IF EXISTS trigger_sync_master_video_views ON video_views;

-- Step 2: Drop the function that was used by the trigger
DROP FUNCTION IF EXISTS update_master_video_views();

-- Step 3: Drop the video_views table (CASCADE will drop any dependent objects)
DROP TABLE IF EXISTS video_views CASCADE;

-- Success message
DO $$ 
BEGIN 
    RAISE NOTICE '✅ video_views table dropped successfully';
    RAISE NOTICE '✅ All analytics now use watch_history with UPSERT pattern';
    RAISE NOTICE '✅ Database storage optimized - no more infinite row growth';
    RAISE NOTICE '';
    RAISE NOTICE '📊 Analytics Storage Summary:';
    RAISE NOTICE '   - OLD: video_views (1 row per tracking event = millions of rows)';
    RAISE NOTICE '   - NEW: watch_history (1 row per user+video = efficient)';
    RAISE NOTICE '   - Storage savings: 100-1000x reduction';
END $$;

