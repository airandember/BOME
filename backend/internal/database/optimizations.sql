-- BOME Database Optimization Script
-- This file contains performance optimizations for the BOME streaming platform

-- =====================================================
-- COMPOSITE INDEXES FOR COMMON QUERY PATTERNS
-- =====================================================

-- Video-related indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_videos_status_category_created 
ON videos(status, category, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_videos_ready_views 
ON videos(status, view_count DESC) WHERE status = 'ready';

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_videos_search_text 
ON videos USING gin(to_tsvector('english', title || ' ' || description));

-- User session optimization
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_sessions_active_expires 
ON user_sessions(is_active, expires_at) WHERE is_active = TRUE;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_sessions_user_activity 
ON user_sessions(user_id, last_activity DESC) WHERE is_active = TRUE;

-- Analytics and audit logs
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_composite 
ON audit_logs(user_id, action, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_activity_type_created 
ON user_activity(activity_type, created_at DESC);

-- Advertisement system indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ad_campaigns_active_dates 
ON ad_campaigns(status, start_date, end_date) WHERE status = 'active';

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ad_schedules_active_time 
ON ad_schedules(is_active, start_date, end_date) WHERE is_active = TRUE;

-- =====================================================
-- OPTIMIZED VIEWS FOR COMMON QUERIES
-- =====================================================

-- Video list with user information
CREATE OR REPLACE VIEW video_list_view AS
SELECT 
    v.id,
    v.title,
    v.description,
    v.bunny_video_id,
    v.thumbnail_url,
    v.duration,
    v.file_size,
    v.status,
    v.category,
    v.tags,
    v.view_count,
    v.like_count,
    v.created_at,
    v.updated_at,
    u.first_name as creator_first_name,
    u.last_name as creator_last_name,
    u.email as creator_email
FROM videos v
JOIN users u ON v.created_by = u.id
WHERE v.status = 'ready';

-- User dashboard view
CREATE OR REPLACE VIEW user_dashboard_view AS
SELECT 
    u.id,
    u.email,
    u.first_name,
    u.last_name,
    u.role,
    u.created_at,
    u.last_login,
    s.status as subscription_status,
    s.current_period_end as subscription_expires,
    COUNT(DISTINCT v.id) as video_count,
    COUNT(DISTINCT l.id) as like_count,
    COUNT(DISTINCT f.id) as favorite_count
FROM users u
LEFT JOIN subscriptions s ON u.id = s.user_id AND s.status IN ('active', 'trialing')
LEFT JOIN videos v ON u.id = v.created_by
LEFT JOIN likes l ON u.id = l.user_id
LEFT JOIN favorites f ON u.id = f.user_id
GROUP BY u.id, u.email, u.first_name, u.last_name, u.role, u.created_at, u.last_login, s.status, s.current_period_end;

-- =====================================================
-- STORED PROCEDURES FOR COMMON OPERATIONS
-- =====================================================

-- Increment video view count with analytics
CREATE OR REPLACE FUNCTION increment_video_views(video_id_param INTEGER, user_id_param INTEGER DEFAULT NULL)
RETURNS VOID AS $$
BEGIN
    -- Update video view count
    UPDATE videos 
    SET view_count = view_count + 1, updated_at = NOW() 
    WHERE id = video_id_param;
    
    -- Record user activity if user is logged in
    IF user_id_param IS NOT NULL THEN
        INSERT INTO user_activity (user_id, activity_type, video_id, created_at)
        VALUES (user_id_param, 'video_view', video_id_param, NOW());
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Cleanup expired sessions and tokens
CREATE OR REPLACE FUNCTION cleanup_expired_data()
RETURNS INTEGER AS $$
DECLARE
    cleaned_count INTEGER := 0;
BEGIN
    -- Clean expired sessions
    DELETE FROM user_sessions WHERE expires_at < NOW();
    GET DIAGNOSTICS cleaned_count = ROW_COUNT;
    
    -- Clean expired tokens
    UPDATE users 
    SET verification_token = NULL 
    WHERE updated_at < NOW() - INTERVAL '24 hours' AND verification_token IS NOT NULL;
    
    UPDATE users 
    SET reset_token = NULL, reset_token_expiry = NULL 
    WHERE reset_token_expiry < NOW() AND reset_token IS NOT NULL;
    
    RETURN cleaned_count;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- MATERIALIZED VIEWS FOR ANALYTICS
-- =====================================================

-- Daily video statistics
CREATE MATERIALIZED VIEW IF NOT EXISTS daily_video_stats AS
SELECT 
    DATE(created_at) as date,
    COUNT(*) as videos_uploaded,
    SUM(view_count) as total_views,
    SUM(like_count) as total_likes,
    AVG(duration) as avg_duration
FROM videos
WHERE status = 'ready'
GROUP BY DATE(created_at)
ORDER BY date DESC;

-- Create unique index for concurrent refresh
CREATE UNIQUE INDEX IF NOT EXISTS idx_daily_video_stats_date 
ON daily_video_stats(date);

-- User engagement statistics
CREATE MATERIALIZED VIEW IF NOT EXISTS user_engagement_stats AS
SELECT 
    u.id as user_id,
    u.email,
    u.role,
    COUNT(DISTINCT v.id) as videos_created,
    COUNT(DISTINCT l.id) as likes_given,
    COUNT(DISTINCT f.id) as favorites_count,
    COUNT(DISTINCT c.id) as comments_count,
    MAX(ua.created_at) as last_activity
FROM users u
LEFT JOIN videos v ON u.id = v.created_by
LEFT JOIN likes l ON u.id = l.user_id
LEFT JOIN favorites f ON u.id = f.user_id
LEFT JOIN comments c ON u.id = c.user_id
LEFT JOIN user_activity ua ON u.id = ua.user_id
GROUP BY u.id, u.email, u.role;

-- Create unique index for concurrent refresh
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_engagement_stats_user_id 
ON user_engagement_stats(user_id);

-- =====================================================
-- PERFORMANCE MONITORING QUERIES
-- =====================================================

-- Query to identify slow queries
CREATE OR REPLACE VIEW slow_queries AS
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    rows
FROM pg_stat_statements
WHERE mean_time > 100  -- Queries taking more than 100ms on average
ORDER BY mean_time DESC;

-- Table size monitoring
CREATE OR REPLACE VIEW table_sizes AS
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size,
    pg_total_relation_size(schemaname||'.'||tablename) as size_bytes
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- =====================================================
-- AUTOMATIC MAINTENANCE JOBS
-- =====================================================

-- Schedule automatic cleanup (requires pg_cron extension)
-- SELECT cron.schedule('cleanup-expired-data', '0 2 * * *', 'SELECT cleanup_expired_data();');

-- Refresh materialized views daily
-- SELECT cron.schedule('refresh-daily-stats', '0 1 * * *', 'REFRESH MATERIALIZED VIEW CONCURRENTLY daily_video_stats;');
-- SELECT cron.schedule('refresh-user-stats', '0 3 * * *', 'REFRESH MATERIALIZED VIEW CONCURRENTLY user_engagement_stats;');

-- =====================================================
-- QUERY OPTIMIZATION HINTS
-- =====================================================

-- Example of optimized video search query
/*
-- Instead of:
SELECT * FROM videos WHERE title ILIKE '%search%' OR description ILIKE '%search%';

-- Use:
SELECT * FROM videos WHERE to_tsvector('english', title || ' ' || description) @@ plainto_tsquery('english', 'search');
*/

-- Example of optimized pagination
/*
-- Instead of:
SELECT * FROM videos ORDER BY created_at DESC LIMIT 20 OFFSET 1000;

-- Use cursor-based pagination:
SELECT * FROM videos WHERE created_at < $cursor ORDER BY created_at DESC LIMIT 20;
*/ 