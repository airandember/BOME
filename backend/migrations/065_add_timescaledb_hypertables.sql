-- Migration 065: Add TimescaleDB for Time-Series Analytics
-- Description: Convert watch_history to a TimescaleDB hypertable for optimized time-series queries
-- Prerequisites: TimescaleDB extension must be enabled on PostgreSQL

-- Step 1: Enable TimescaleDB extension
-- Note: This requires PostgreSQL superuser privileges
-- If this fails, your DBA needs to run: CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- Step 2: Convert watch_history to a hypertable
-- This partitions the table by time for efficient time-based queries
SELECT create_hypertable(
    'watch_history', 
    'last_watched_at',
    chunk_time_interval => INTERVAL '7 days',  -- Partition by week
    if_not_exists => TRUE,
    migrate_data => TRUE  -- Migrate existing data into chunks
);

-- Step 3: Add compression policy
-- Compress data older than 7 days to save storage (20-95% compression)
SELECT add_compression_policy('watch_history', INTERVAL '7 days', if_not_exists => TRUE);

-- Step 4: Add retention policy (optional)
-- Automatically drop data older than 2 years
-- Comment out if you want to keep all historical data
-- SELECT add_retention_policy('watch_history', INTERVAL '2 years', if_not_exists => TRUE);

-- Step 5: Create continuous aggregates for fast queries
-- These are materialized views that auto-update
CREATE MATERIALIZED VIEW IF NOT EXISTS watch_history_hourly
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 hour', last_watched_at) AS bucket,
    video_id,
    COUNT(DISTINCT COALESCE(user_id::text, session_id)) AS unique_viewers,
    AVG(progress_percentage) AS avg_completion,
    SUM(total_watch_time) AS total_watch_time,
    COUNT(*) FILTER (WHERE completed = true) AS completions
FROM watch_history
GROUP BY bucket, video_id
WITH NO DATA;

-- Add refresh policy for continuous aggregate (refresh every hour)
SELECT add_continuous_aggregate_policy('watch_history_hourly',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE
);

-- Step 6: Create daily rollup for long-term analytics
CREATE MATERIALIZED VIEW IF NOT EXISTS watch_history_daily
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 day', last_watched_at) AS bucket,
    video_id,
    COUNT(DISTINCT COALESCE(user_id::text, session_id)) AS unique_viewers,
    AVG(progress_percentage) AS avg_completion,
    SUM(total_watch_time) AS total_watch_time,
    COUNT(*) FILTER (WHERE completed = true) AS completions,
    COUNT(*) FILTER (WHERE view_count > 1) AS repeat_viewers
FROM watch_history
GROUP BY bucket, video_id
WITH NO DATA;

-- Add refresh policy for daily aggregate
SELECT add_continuous_aggregate_policy('watch_history_daily',
    start_offset => INTERVAL '7 days',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- Step 7: Add indexes optimized for TimescaleDB
-- TimescaleDB automatically creates time-based indexes, but we need these for other columns
CREATE INDEX IF NOT EXISTS idx_watch_history_video_time ON watch_history(video_id, last_watched_at DESC);
CREATE INDEX IF NOT EXISTS idx_watch_history_user_time ON watch_history(user_id, last_watched_at DESC) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_watch_history_session_time ON watch_history(session_id, last_watched_at DESC) WHERE session_id IS NOT NULL;

-- Step 8: Add comments
COMMENT ON MATERIALIZED VIEW watch_history_hourly IS 'Hourly aggregated video engagement metrics (auto-refreshed every hour)';
COMMENT ON MATERIALIZED VIEW watch_history_daily IS 'Daily aggregated video engagement metrics (auto-refreshed daily)';

-- Step 9: Manually refresh the continuous aggregates for existing data
CALL refresh_continuous_aggregate('watch_history_hourly', NULL, NULL);
CALL refresh_continuous_aggregate('watch_history_daily', NULL, NULL);

-- Success message
DO $$ 
BEGIN 
    RAISE NOTICE '✅ TimescaleDB hypertable created for watch_history';
    RAISE NOTICE '✅ Continuous aggregates created (hourly, daily)';
    RAISE NOTICE '✅ Compression policy added (compress after 7 days)';
    RAISE NOTICE '✅ Analytics queries will now use time-series optimizations';
    RAISE NOTICE '';
    RAISE NOTICE '📊 Performance improvements:';
    RAISE NOTICE '   - 10-100x faster time-range queries';
    RAISE NOTICE '   - 20-95% storage reduction via compression';
    RAISE NOTICE '   - Automatic data management';
END $$;

