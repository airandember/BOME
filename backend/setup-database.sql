-- BOME (Book of Mormon Evidences) Database Schema Setup
-- PostgreSQL Database Schema for BOME Streaming Platform

-- =============================================================================
-- DATABASE SETUP
-- =============================================================================

-- Create database and user (run as postgres superuser)
-- CREATE USER bome_admin WITH PASSWORD 'AdminBOME';
-- CREATE DATABASE bome_db OWNER bome_admin;
-- GRANT ALL PRIVILEGES ON DATABASE bome_db TO bome_admin;
-- ALTER USER bome_admin CREATEDB;

-- Connect to bome_db and run the following:

-- =============================================================================
-- CORE TABLES
-- =============================================================================

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role VARCHAR(50) DEFAULT 'user',
    email_verified BOOLEAN DEFAULT FALSE,
    stripe_customer_id VARCHAR(255),
    reset_token VARCHAR(255),
    reset_token_expiry TIMESTAMP,
    verification_token VARCHAR(255),
    -- Extended profile fields
    bio TEXT,
    location VARCHAR(255),
    website VARCHAR(500),
    phone VARCHAR(50),
    avatar_url VARCHAR(500),
    preferences JSONB DEFAULT '{}',
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Videos table
CREATE TABLE IF NOT EXISTS videos (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    bunny_video_id VARCHAR(255) UNIQUE NOT NULL,
    thumbnail_url VARCHAR(500),
    duration INTEGER,
    file_size BIGINT,
    status VARCHAR(50) DEFAULT 'processing',
    category VARCHAR(100),
    tags TEXT,
    view_count INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Subscriptions table
CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    stripe_subscription_id VARCHAR(255) UNIQUE NOT NULL,
    stripe_price_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    current_period_start TIMESTAMP,
    current_period_end TIMESTAMP,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Comments table
CREATE TABLE IF NOT EXISTS comments (
    id SERIAL PRIMARY KEY,
    video_id INTEGER REFERENCES videos(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    parent_id INTEGER REFERENCES comments(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Likes table
CREATE TABLE IF NOT EXISTS likes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    video_id INTEGER REFERENCES videos(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, video_id)
);

-- Favorites table
CREATE TABLE IF NOT EXISTS favorites (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    video_id INTEGER REFERENCES videos(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, video_id)
);

-- User activity table
CREATE TABLE IF NOT EXISTS user_activity (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    activity_type VARCHAR(50) NOT NULL,
    activity_data JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Admin logs table
CREATE TABLE IF NOT EXISTS admin_logs (
    id SERIAL PRIMARY KEY,
    admin_user_id INTEGER REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    target_type VARCHAR(50),
    target_id INTEGER,
    details JSONB,
    ip_address INET,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- ADVERTISING TABLES
-- =============================================================================

-- Advertiser accounts table
CREATE TABLE IF NOT EXISTS advertiser_accounts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    company_name VARCHAR(255) NOT NULL,
    business_email VARCHAR(255) UNIQUE NOT NULL,
    contact_name VARCHAR(255) NOT NULL,
    contact_phone VARCHAR(50),
    business_address TEXT,
    tax_id VARCHAR(100),
    website VARCHAR(500),
    industry VARCHAR(100),
    status VARCHAR(50) DEFAULT 'pending',
    verification_notes TEXT,
    stripe_customer_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Ad campaigns table
CREATE TABLE IF NOT EXISTS ad_campaigns (
    id SERIAL PRIMARY KEY,
    advertiser_id INTEGER REFERENCES advertiser_accounts(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'draft',
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    budget DECIMAL(10,2) NOT NULL,
    spent_amount DECIMAL(10,2) DEFAULT 0.00,
    target_audience TEXT,
    billing_type VARCHAR(50) DEFAULT 'monthly',
    billing_rate DECIMAL(10,2) NOT NULL,
    approval_notes TEXT,
    approved_by INTEGER REFERENCES users(id),
    approved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Advertisements table
CREATE TABLE IF NOT EXISTS advertisements (
    id SERIAL PRIMARY KEY,
    campaign_id INTEGER REFERENCES ad_campaigns(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    image_url VARCHAR(500),
    click_url VARCHAR(500) NOT NULL,
    ad_type VARCHAR(50) NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    priority INTEGER DEFAULT 1,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Ad placements table
CREATE TABLE IF NOT EXISTS ad_placements (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    location VARCHAR(100) NOT NULL,
    ad_type VARCHAR(50) NOT NULL,
    max_width INTEGER NOT NULL,
    max_height INTEGER NOT NULL,
    base_rate DECIMAL(10,2) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Ad schedules table
CREATE TABLE IF NOT EXISTS ad_schedules (
    id SERIAL PRIMARY KEY,
    ad_id INTEGER REFERENCES advertisements(id) ON DELETE CASCADE,
    placement_id INTEGER REFERENCES ad_placements(id) ON DELETE CASCADE,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    days_of_week TEXT,
    start_time VARCHAR(5),
    end_time VARCHAR(5),
    weight INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Ad analytics table
CREATE TABLE IF NOT EXISTS ad_analytics (
    id SERIAL PRIMARY KEY,
    ad_id INTEGER REFERENCES advertisements(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    impressions BIGINT DEFAULT 0,
    clicks BIGINT DEFAULT 0,
    unique_views BIGINT DEFAULT 0,
    view_duration BIGINT DEFAULT 0,
    revenue DECIMAL(10,2) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ad_id, date)
);

-- Ad clicks table
CREATE TABLE IF NOT EXISTS ad_clicks (
    id SERIAL PRIMARY KEY,
    ad_id INTEGER REFERENCES advertisements(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id),
    ip_address INET NOT NULL,
    user_agent TEXT,
    referrer TEXT,
    clicked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Ad impressions table
CREATE TABLE IF NOT EXISTS ad_impressions (
    id SERIAL PRIMARY KEY,
    ad_id INTEGER REFERENCES advertisements(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id),
    ip_address INET NOT NULL,
    user_agent TEXT,
    view_duration INTEGER DEFAULT 0,
    viewed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Ad billing table
CREATE TABLE IF NOT EXISTS ad_billing (
    id SERIAL PRIMARY KEY,
    campaign_id INTEGER REFERENCES ad_campaigns(id) ON DELETE CASCADE,
    advertiser_id INTEGER REFERENCES advertiser_accounts(id) ON DELETE CASCADE,
    billing_period VARCHAR(50) NOT NULL,
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    tax_amount DECIMAL(10,2) DEFAULT 0.00,
    total_amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    stripe_invoice_id VARCHAR(255),
    payment_intent_id VARCHAR(255),
    paid_at TIMESTAMP,
    due_date TIMESTAMP NOT NULL,
    invoice_url TEXT,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Ad audit log table
CREATE TABLE IF NOT EXISTS ad_audit_log (
    id SERIAL PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,
    entity_id INTEGER NOT NULL,
    action VARCHAR(100) NOT NULL,
    actor_id INTEGER REFERENCES users(id),
    actor_type VARCHAR(50) NOT NULL,
    old_values TEXT,
    new_values TEXT,
    ip_address INET,
    user_agent TEXT,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- INDEXES FOR PERFORMANCE
-- =============================================================================

-- Core table indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_email_verified ON users(email_verified);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

CREATE INDEX IF NOT EXISTS idx_videos_status ON videos(status);
CREATE INDEX IF NOT EXISTS idx_videos_category ON videos(category);
CREATE INDEX IF NOT EXISTS idx_videos_created_at ON videos(created_at);
CREATE INDEX IF NOT EXISTS idx_videos_created_by ON videos(created_by);
CREATE INDEX IF NOT EXISTS idx_videos_bunny_id ON videos(bunny_video_id);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_stripe_id ON subscriptions(stripe_subscription_id);

CREATE INDEX IF NOT EXISTS idx_comments_video_id ON comments(video_id);
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);
CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments(parent_id);
CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments(created_at);

CREATE INDEX IF NOT EXISTS idx_likes_user_id ON likes(user_id);
CREATE INDEX IF NOT EXISTS idx_likes_video_id ON likes(video_id);

CREATE INDEX IF NOT EXISTS idx_favorites_user_id ON favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_favorites_video_id ON favorites(video_id);

CREATE INDEX IF NOT EXISTS idx_user_activity_user_id ON user_activity(user_id);
CREATE INDEX IF NOT EXISTS idx_user_activity_type ON user_activity(activity_type);
CREATE INDEX IF NOT EXISTS idx_user_activity_created_at ON user_activity(created_at);

CREATE INDEX IF NOT EXISTS idx_admin_logs_admin_user_id ON admin_logs(admin_user_id);
CREATE INDEX IF NOT EXISTS idx_admin_logs_created_at ON admin_logs(created_at);

-- Advertising table indexes
CREATE INDEX IF NOT EXISTS idx_advertiser_accounts_user_id ON advertiser_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_advertiser_accounts_status ON advertiser_accounts(status);
CREATE INDEX IF NOT EXISTS idx_advertiser_accounts_business_email ON advertiser_accounts(business_email);

CREATE INDEX IF NOT EXISTS idx_ad_campaigns_advertiser_id ON ad_campaigns(advertiser_id);
CREATE INDEX IF NOT EXISTS idx_ad_campaigns_status ON ad_campaigns(status);
CREATE INDEX IF NOT EXISTS idx_ad_campaigns_dates ON ad_campaigns(start_date, end_date);

CREATE INDEX IF NOT EXISTS idx_advertisements_campaign_id ON advertisements(campaign_id);
CREATE INDEX IF NOT EXISTS idx_advertisements_status ON advertisements(status);
CREATE INDEX IF NOT EXISTS idx_advertisements_type ON advertisements(ad_type);

CREATE INDEX IF NOT EXISTS idx_ad_placements_location ON ad_placements(location);
CREATE INDEX IF NOT EXISTS idx_ad_placements_active ON ad_placements(is_active);

CREATE INDEX IF NOT EXISTS idx_ad_schedules_ad_id ON ad_schedules(ad_id);
CREATE INDEX IF NOT EXISTS idx_ad_schedules_dates ON ad_schedules(start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_ad_schedules_active ON ad_schedules(is_active);

CREATE INDEX IF NOT EXISTS idx_ad_analytics_ad_id ON ad_analytics(ad_id);
CREATE INDEX IF NOT EXISTS idx_ad_analytics_date ON ad_analytics(date);

CREATE INDEX IF NOT EXISTS idx_ad_clicks_ad_id ON ad_clicks(ad_id);
CREATE INDEX IF NOT EXISTS idx_ad_clicks_clicked_at ON ad_clicks(clicked_at);

CREATE INDEX IF NOT EXISTS idx_ad_impressions_ad_id ON ad_impressions(ad_id);
CREATE INDEX IF NOT EXISTS idx_ad_impressions_viewed_at ON ad_impressions(viewed_at);

CREATE INDEX IF NOT EXISTS idx_ad_billing_campaign_id ON ad_billing(campaign_id);
CREATE INDEX IF NOT EXISTS idx_ad_billing_status ON ad_billing(status);
CREATE INDEX IF NOT EXISTS idx_ad_billing_period ON ad_billing(period_start, period_end);

CREATE INDEX IF NOT EXISTS idx_ad_audit_log_entity ON ad_audit_log(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_ad_audit_log_actor ON ad_audit_log(actor_id, actor_type);
CREATE INDEX IF NOT EXISTS idx_ad_audit_log_created_at ON ad_audit_log(created_at);

-- =============================================================================
-- SEED DATA
-- =============================================================================

-- Insert default ad placements
INSERT INTO ad_placements (name, description, location, ad_type, max_width, max_height, base_rate, is_active) VALUES
('Header Banner', 'Banner ad displayed in the site header', 'header', 'banner', 728, 90, 50.00, true),
('Sidebar Banner', 'Banner ad displayed in the sidebar', 'sidebar', 'banner', 300, 250, 30.00, true),
('Footer Banner', 'Banner ad displayed in the site footer', 'footer', 'banner', 728, 90, 40.00, true),
('Content Banner', 'Banner ad displayed within content areas', 'content', 'banner', 728, 90, 45.00, true),
('Video Overlay', 'Overlay ad displayed during video playback', 'video_overlay', 'large', 640, 480, 75.00, true)
ON CONFLICT DO NOTHING;

-- =============================================================================
-- GRANTS
-- =============================================================================

-- Grant permissions to bome_admin user
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO bome_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO bome_admin;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO bome_admin;

-- =============================================================================
-- VERIFICATION QUERIES
-- =============================================================================

-- Check if tables were created successfully
SELECT table_name, table_type 
FROM information_schema.tables 
WHERE table_schema = 'public' 
ORDER BY table_name;

-- Check table counts
SELECT 
    'users' as table_name, COUNT(*) as count FROM users
UNION ALL
SELECT 'videos', COUNT(*) FROM videos
UNION ALL
SELECT 'ad_placements', COUNT(*) FROM ad_placements
UNION ALL
SELECT 'advertiser_accounts', COUNT(*) FROM advertiser_accounts; 