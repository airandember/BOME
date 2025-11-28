-- Video Analytics Migration Runner
-- Run this file to set up all video analytics tables and triggers
-- 
-- This will:
-- 1. Create video_views and watch_history tables
-- 2. Create trigger to sync master_video_list.views
-- 3. Backfill existing data (if any)

\echo '============================================'
\echo 'Video Analytics Migration - Starting'
\echo '============================================'
\echo ''

\echo 'Step 1/2: Creating video analytics tables...'
\i 060_1_create_video_analytics_tables.sql
\echo '✓ Tables created'
\echo ''

\echo 'Step 2/2: Setting up view count synchronization...'
\i 062_sync_master_video_views.sql
\echo '✓ Sync trigger installed'
\echo ''

\echo '============================================'
\echo 'Video Analytics Migration - Complete!'
\echo '============================================'
\echo ''
\echo 'Next steps:'
\echo '1. Verify tables exist: \dt video_views watch_history'
\echo '2. Verify trigger: SELECT * FROM information_schema.triggers WHERE trigger_name = ''trigger_sync_master_video_views'';'
\echo '3. Test with: INSERT INTO video_views (video_id, session_id, watched_duration, watched_percentage) VALUES (1, ''test-123'', 60, 50.0);'
\echo ''

