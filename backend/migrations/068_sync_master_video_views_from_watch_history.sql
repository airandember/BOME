-- Sync master_video_list.views from watch_history
-- Automatically update master_video_list.views when watch_history changes

-- Drop existing trigger if any
DROP TRIGGER IF EXISTS update_master_video_views_trigger ON watch_history;
DROP FUNCTION IF EXISTS update_master_video_views();

-- Create function to sync views
CREATE OR REPLACE FUNCTION update_master_video_views()
RETURNS TRIGGER AS $$
BEGIN
    -- Update master_video_list.views with sum of view_count from watch_history
    UPDATE master_video_list
    SET views = (
        SELECT COALESCE(SUM(view_count), 0)
        FROM watch_history
        WHERE video_id = NEW.video_id
    )
    WHERE id = NEW.video_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger on INSERT and UPDATE
CREATE TRIGGER update_master_video_views_trigger
AFTER INSERT OR UPDATE OF view_count ON watch_history
FOR EACH ROW
EXECUTE FUNCTION update_master_video_views();

-- Backfill existing data
-- Update all videos with current watch_history counts
UPDATE master_video_list mvl
SET views = COALESCE((
    SELECT SUM(wh.view_count)
    FROM watch_history wh
    WHERE wh.video_id = mvl.id
), 0)
WHERE EXISTS (
    SELECT 1 FROM watch_history WHERE video_id = mvl.id
);

-- Verify the sync
SELECT 
    mvl.id,
    mvl.title,
    mvl.views AS master_video_list_views,
    COALESCE(SUM(wh.view_count), 0) AS watch_history_total_views
FROM master_video_list mvl
LEFT JOIN watch_history wh ON wh.video_id = mvl.id
GROUP BY mvl.id, mvl.title, mvl.views
ORDER BY mvl.views DESC
LIMIT 10;

