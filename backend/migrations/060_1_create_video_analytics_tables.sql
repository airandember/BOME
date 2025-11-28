-- Migration 060.1: Create Video Analytics Tables
-- Description: Creates tables for detailed video tracking and analytics
-- Must run BEFORE 062_sync_master_video_views.sql

-- Table 1: video_views - Track individual view events
CREATE TABLE IF NOT EXISTS video_views (
    id SERIAL PRIMARY KEY,
    video_id INTEGER NOT NULL REFERENCES master_video_list(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    session_id VARCHAR(255),
    ip_address INET,
    watched_duration INTEGER NOT NULL DEFAULT 0, -- Seconds watched
    watched_percentage DECIMAL(5,2) NOT NULL DEFAULT 0.0, -- 0-100
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT video_views_check_percentage CHECK (watched_percentage >= 0 AND watched_percentage <= 100),
    CONSTRAINT video_views_check_duration CHECK (watched_duration >= 0),
    CONSTRAINT video_views_check_user_or_session CHECK (user_id IS NOT NULL OR session_id IS NOT NULL)
);

-- Table 2: watch_history - Track user watch progress
CREATE TABLE IF NOT EXISTS watch_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    video_id INTEGER NOT NULL REFERENCES master_video_list(id) ON DELETE CASCADE,
    last_position INTEGER NOT NULL DEFAULT 0, -- Seconds
    progress_percentage DECIMAL(5,2) NOT NULL DEFAULT 0.0, -- 0-100
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    last_watched_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- One record per user-video combination
    UNIQUE(user_id, video_id),
    
    -- Constraints
    CONSTRAINT watch_history_check_percentage CHECK (progress_percentage >= 0 AND progress_percentage <= 100),
    CONSTRAINT watch_history_check_position CHECK (last_position >= 0)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_video_views_video_id ON video_views(video_id);
CREATE INDEX IF NOT EXISTS idx_video_views_user_id ON video_views(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_video_views_session_id ON video_views(session_id) WHERE session_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_video_views_created_at ON video_views(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_video_views_video_created ON video_views(video_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_watch_history_user_id ON watch_history(user_id);
CREATE INDEX IF NOT EXISTS idx_watch_history_video_id ON watch_history(video_id);
CREATE INDEX IF NOT EXISTS idx_watch_history_last_watched ON watch_history(last_watched_at DESC);
CREATE INDEX IF NOT EXISTS idx_watch_history_user_incomplete ON watch_history(user_id, completed) WHERE completed = FALSE;

-- Comments
COMMENT ON TABLE video_views IS 'Individual video view events for analytics';
COMMENT ON TABLE watch_history IS 'User watch progress for resume functionality';

COMMENT ON COLUMN video_views.watched_duration IS 'Total seconds watched in this viewing session';
COMMENT ON COLUMN video_views.watched_percentage IS 'Percentage of video watched (0-100)';
COMMENT ON COLUMN video_views.session_id IS 'Anonymous session ID for non-authenticated users';

COMMENT ON COLUMN watch_history.last_position IS 'Last playback position in seconds';
COMMENT ON COLUMN watch_history.progress_percentage IS 'Overall progress percentage (0-100)';
COMMENT ON COLUMN watch_history.completed IS 'Whether user has completed watching (>95%)';

