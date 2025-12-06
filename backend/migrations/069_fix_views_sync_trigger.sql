-- Migration 069: Fix views sync trigger to not overwrite historical Bunny.net views
-- The old trigger was replacing master_video_list.views with watch_history sum,
-- which destroyed historical view counts from Bunny.net.

-- Step 1: Drop the problematic trigger
DROP TRIGGER IF EXISTS update_master_video_views_trigger ON watch_history;
DROP FUNCTION IF EXISTS update_master_video_views();

-- Step 2: Create a new function that properly handles view counting
-- Instead of replacing views, we'll track watch_history views separately
-- and use the GREATER of (bunny_views, watch_history_views) or ADD new views

CREATE OR REPLACE FUNCTION sync_watch_history_views()
RETURNS TRIGGER AS $$
DECLARE
    old_view_count INTEGER;
    new_view_count INTEGER;
    view_delta INTEGER;
BEGIN
    -- Get the old view_count for this record (0 if INSERT)
    IF TG_OP = 'INSERT' THEN
        old_view_count := 0;
    ELSE
        old_view_count := COALESCE(OLD.view_count, 0);
    END IF;
    
    -- Get the new view_count
    new_view_count := COALESCE(NEW.view_count, 0);
    
    -- Calculate the delta (how many new views were added)
    view_delta := new_view_count - old_view_count;
    
    -- Only update if there's a positive delta (new views were added)
    IF view_delta > 0 THEN
        UPDATE master_video_list
        SET views = COALESCE(views, 0) + view_delta,
            updated_at = NOW()
        WHERE id = NEW.video_id;
        
        RAISE NOTICE '[Views Sync] Added % view(s) to video %, new total: %', 
            view_delta, NEW.video_id, 
            (SELECT views FROM master_video_list WHERE id = NEW.video_id);
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Step 3: Create the new trigger
CREATE TRIGGER sync_watch_history_views_trigger
AFTER INSERT OR UPDATE OF view_count ON watch_history
FOR EACH ROW
EXECUTE FUNCTION sync_watch_history_views();

-- Step 4: Add a comment explaining the fix
COMMENT ON FUNCTION sync_watch_history_views() IS 
    'Increments master_video_list.views by the DELTA of view_count changes. 
     This preserves historical Bunny.net view counts while adding new views from watch_history.
     Fixed in migration 069 - the old trigger was replacing views instead of adding.';

-- Note: We do NOT backfill here because the views already in master_video_list
-- are likely the correct historical values from Bunny.net. 
-- Going forward, new views will be ADDED to this count.

-- Verification query (run manually to check current state):
-- SELECT id, title, views 
-- FROM master_video_list 
-- WHERE status = 'ready' 
-- ORDER BY views DESC 
-- LIMIT 20;

