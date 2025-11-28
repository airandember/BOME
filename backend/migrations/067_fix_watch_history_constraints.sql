-- Fix watch_history constraints for ON CONFLICT to work properly
-- Issue: Complex COALESCE constraint doesn't work with ON CONFLICT

-- Drop old complex constraints
DROP INDEX IF EXISTS idx_watch_history_user_video;
DROP INDEX IF EXISTS idx_watch_history_session_video;

-- Add simple unique constraints that work with ON CONFLICT
-- For authenticated users: unique on (user_id, video_id) where user_id is not null
CREATE UNIQUE INDEX idx_watch_history_user_video_simple
ON watch_history (user_id, video_id)
WHERE user_id IS NOT NULL;

-- For anonymous users: unique on (session_id, video_id) where session_id is not null
CREATE UNIQUE INDEX idx_watch_history_session_video_simple
ON watch_history (session_id, video_id)
WHERE session_id IS NOT NULL AND session_id != '';

-- Verify constraints
SELECT 
    indexname, 
    indexdef 
FROM pg_indexes 
WHERE tablename = 'watch_history' 
  AND indexname LIKE 'idx_watch_history%';

