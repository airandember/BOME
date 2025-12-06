-- Migration 071: UNDO the damage from migration 070
-- This migration needs to be followed by running the Go script to restore actual Bunny.net views

-- First, let's see what the damage looks like
-- Run this SELECT to see videos that have been incorrectly set to 100
SELECT 
    id,
    title,
    views,
    bunny_video_id,
    (SELECT COUNT(*) FROM watch_history WHERE video_id = mvl.id) as watch_history_entries
FROM master_video_list mvl
WHERE views = 100
AND status = 'ready'
ORDER BY id DESC
LIMIT 20;

-- The actual restoration will be done via a Go script that:
-- 1. Fetches each video from Bunny.net API
-- 2. Updates master_video_list.views with bunnyVideo.Views
-- 
-- See: backend/cmd/restore_bunny_views/main.go

-- For now, we can at least mark that this issue exists
-- Add a comment to the views column documenting this
COMMENT ON COLUMN master_video_list.views IS 
    'Video view count. NOTE: Views were incorrectly set to 100 by migration 070. 
     Run the restore_bunny_views script to restore actual Bunny.net view counts.';

