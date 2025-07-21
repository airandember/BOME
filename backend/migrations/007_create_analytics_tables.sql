-- Migration: Create Analytics Tables
-- Description: Creates tables for analytics, webhooks, system metrics, and request logging

-- Analytics Events Table
CREATE TABLE IF NOT EXISTS analytics_events (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    session_id VARCHAR(255),
    subsite VARCHAR(50) DEFAULT 'streaming',
    event_data TEXT, -- JSON string
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for analytics events
CREATE INDEX IF NOT EXISTS idx_analytics_events_created_at ON analytics_events(created_at);
CREATE INDEX IF NOT EXISTS idx_analytics_events_event_type ON analytics_events(event_type);
CREATE INDEX IF NOT EXISTS idx_analytics_events_user_id ON analytics_events(user_id);
CREATE INDEX IF NOT EXISTS idx_analytics_events_subsite ON analytics_events(subsite);

-- Webhook Events Table
CREATE TABLE IF NOT EXISTS webhook_events (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    subsite VARCHAR(50) DEFAULT 'streaming',
    endpoint VARCHAR(500) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- success, failed, pending
    response_time INTEGER, -- milliseconds
    payload_size INTEGER, -- bytes
    status_code INTEGER,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for webhook events
CREATE INDEX IF NOT EXISTS idx_webhook_events_created_at ON webhook_events(created_at);
CREATE INDEX IF NOT EXISTS idx_webhook_events_status ON webhook_events(status);
CREATE INDEX IF NOT EXISTS idx_webhook_events_subsite ON webhook_events(subsite);
CREATE INDEX IF NOT EXISTS idx_webhook_events_event_type ON webhook_events(event_type);

-- System Metrics Table
CREATE TABLE IF NOT EXISTS system_metrics (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    cpu_usage DECIMAL(5,2), -- percentage
    memory_usage DECIMAL(5,2), -- percentage
    disk_usage DECIMAL(5,2), -- percentage
    network_in BIGINT, -- bytes
    network_out BIGINT, -- bytes
    active_sessions INTEGER,
    error_rate DECIMAL(5,4), -- percentage as decimal
    response_time INTEGER, -- milliseconds
    database_size BIGINT, -- bytes
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for system metrics
CREATE INDEX IF NOT EXISTS idx_system_metrics_timestamp ON system_metrics(timestamp);
CREATE INDEX IF NOT EXISTS idx_system_metrics_created_at ON system_metrics(created_at);

-- Request Logs Table (for error rate and response time calculations)
CREATE TABLE IF NOT EXISTS request_logs (
    id SERIAL PRIMARY KEY,
    method VARCHAR(10) NOT NULL,
    path VARCHAR(500) NOT NULL,
    status_code INTEGER NOT NULL,
    response_time INTEGER, -- milliseconds
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for request logs
CREATE INDEX IF NOT EXISTS idx_request_logs_created_at ON request_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_request_logs_status_code ON request_logs(status_code);
CREATE INDEX IF NOT EXISTS idx_request_logs_path ON request_logs(path);

-- Video Ratings Table (for average rating calculations)
CREATE TABLE IF NOT EXISTS video_ratings (
    id SERIAL PRIMARY KEY,
    video_id INTEGER REFERENCES videos(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    review TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(video_id, user_id)
);

-- Create index for video ratings
CREATE INDEX IF NOT EXISTS idx_video_ratings_video_id ON video_ratings(video_id);
CREATE INDEX IF NOT EXISTS idx_video_ratings_rating ON video_ratings(rating);

-- Cross-Subsite Stats Table
CREATE TABLE IF NOT EXISTS cross_subsite_stats (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL,
    subsite VARCHAR(50) NOT NULL,
    users INTEGER DEFAULT 0,
    content INTEGER DEFAULT 0,
    views INTEGER DEFAULT 0,
    revenue DECIMAL(10,2) DEFAULT 0.00,
    engagement_rate DECIMAL(5,4) DEFAULT 0.0000,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(date, subsite)
);

-- Create index for cross-subsite stats
CREATE INDEX IF NOT EXISTS idx_cross_subsite_stats_date ON cross_subsite_stats(date);
CREATE INDEX IF NOT EXISTS idx_cross_subsite_stats_subsite ON cross_subsite_stats(subsite);

-- Alerts Table
CREATE TABLE IF NOT EXISTS alerts (
    id SERIAL PRIMARY KEY,
    severity VARCHAR(20) NOT NULL DEFAULT 'info', -- info, warning, critical
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    subsite VARCHAR(50),
    acknowledged BOOLEAN DEFAULT FALSE,
    acknowledged_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for alerts
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);
CREATE INDEX IF NOT EXISTS idx_alerts_acknowledged ON alerts(acknowledged);

-- Insert some sample data for testing
INSERT INTO cross_subsite_stats (date, subsite, users, content, views, revenue, engagement_rate) VALUES
    (CURRENT_DATE, 'streaming', 1250, 89, 4567, 1250.00, 0.0234),
    (CURRENT_DATE, 'articles', 890, 45, 2340, 890.00, 0.0189),
    (CURRENT_DATE, 'expo', 567, 12, 890, 567.00, 0.0156)
ON CONFLICT (date, subsite) DO NOTHING;

-- Insert sample system metrics
INSERT INTO system_metrics (cpu_usage, memory_usage, disk_usage, network_in, network_out, active_sessions, error_rate, response_time, database_size) VALUES
    (25.5, 45.2, 67.8, 1024000, 2048000, 45, 0.0234, 150, 1073741824)
ON CONFLICT DO NOTHING;

-- Insert sample webhook events
INSERT INTO webhook_events (event_type, subsite, endpoint, status, response_time, payload_size, status_code) VALUES
    ('user.signup', 'streaming', 'https://webhook.site/abc123', 'success', 245, 1024, 200),
    ('video.upload', 'streaming', 'https://webhook.site/abc123', 'success', 189, 2048, 200),
    ('subscription.created', 'streaming', 'https://webhook.site/abc123', 'failed', 5000, 512, 500)
ON CONFLICT DO NOTHING;

-- Insert sample analytics events
INSERT INTO analytics_events (event_type, subsite, event_data, ip_address) VALUES
    ('page_view', 'streaming', '{"page": "/videos", "referrer": "google.com"}', '192.168.1.1'),
    ('video_view', 'streaming', '{"video_id": 123, "duration": 45}', '192.168.1.2'),
    ('user_signup', 'streaming', '{"source": "organic"}', '192.168.1.3')
ON CONFLICT DO NOTHING; 