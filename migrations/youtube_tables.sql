-- =====================================================
-- YOUTUBE STRAND DATABASE MIGRATION
-- Run this in your PostgreSQL database
-- =====================================================

-- 1. ADD MISSING COLUMNS TO youtube_videos TABLE
-- =====================================================
ALTER TABLE youtube_videos 
ADD COLUMN IF NOT EXISTS tags TEXT[];

ALTER TABLE youtube_videos 
ADD COLUMN IF NOT EXISTS category VARCHAR(100);

-- Add indexes for new columns
CREATE INDEX IF NOT EXISTS idx_youtube_videos_published_at ON youtube_videos(published_at DESC);
CREATE INDEX IF NOT EXISTS idx_youtube_videos_category ON youtube_videos(category);
CREATE INDEX IF NOT EXISTS idx_youtube_videos_tags ON youtube_videos USING GIN(tags);

-- 2. CREATE youtube_config TABLE
-- =====================================================
CREATE TABLE IF NOT EXISTS youtube_config (
    id SERIAL PRIMARY KEY,
    channel_id VARCHAR(255) NOT NULL,
    channel_title VARCHAR(255),
    rss_url TEXT NOT NULL,
    sync_enabled BOOLEAN DEFAULT true,
    sync_schedule VARCHAR(100) DEFAULT '0 14 * * *',
    auto_sync_to_master BOOLEAN DEFAULT false,
    last_sync_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. CREATE youtube_sync_log TABLE
-- =====================================================
CREATE TABLE IF NOT EXISTS youtube_sync_log (
    id SERIAL PRIMARY KEY,
    sync_type VARCHAR(50),  -- 'rss', 'manual', 'scheduled', 'initial'
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    videos_found INTEGER DEFAULT 0,
    videos_new INTEGER DEFAULT 0,
    videos_updated INTEGER DEFAULT 0,
    videos_skipped INTEGER DEFAULT 0,
    status VARCHAR(50),  -- 'running', 'success', 'partial', 'failed'
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_youtube_sync_log_started_at ON youtube_sync_log(started_at DESC);

-- 4. MIGRATE DATA FROM public_settings TO youtube_config
-- =====================================================
-- Only insert if youtube_config is empty
INSERT INTO youtube_config (
    channel_id, 
    channel_title, 
    rss_url, 
    sync_enabled, 
    sync_schedule,
    created_at,
    updated_at
)
SELECT 
    (SELECT value FROM public_settings WHERE key = 'youtube_channel_id'),
    'Book of Mormon Evidence',
    'https://www.youtube.com/feeds/videos.xml?channel_id=' || (SELECT value FROM public_settings WHERE key = 'youtube_channel_id'),
    CASE 
        WHEN (SELECT value FROM public_settings WHERE key = 'youtube_auto_sync_enabled') = 'true' THEN true
        ELSE false
    END,
    '0 ' || COALESCE((SELECT value FROM public_settings WHERE key = 'youtube_sync_hour'), '14') || ' * * *',
    NOW(),
    NOW()
WHERE NOT EXISTS (SELECT 1 FROM youtube_config)
AND EXISTS (SELECT 1 FROM public_settings WHERE key = 'youtube_channel_id');

-- 5. VERIFY THE MIGRATION
-- =====================================================
-- Check youtube_videos columns
SELECT 
    column_name, 
    data_type, 
    is_nullable
FROM information_schema.columns
WHERE table_name = 'youtube_videos'
ORDER BY ordinal_position;

-- Check youtube_config data
SELECT * FROM youtube_config;

-- Check youtube_sync_log (should be empty initially)
SELECT COUNT(*) as log_count FROM youtube_sync_log;

-- =====================================================
-- MIGRATION COMPLETE! 🎉
-- =====================================================

