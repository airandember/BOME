package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"bome-backend/internal/config"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB wraps the database connection
type DB struct {
	*sql.DB
	GormDB *gorm.DB // Add GORM support for design system features
}

// New creates a new database connection
func New(cfg *config.Config) (*DB, error) {
	// Build PostgreSQL connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	// Open database connection with sql package
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for production
	// These settings are optimized for production load
	db.SetMaxOpenConns(50)                  // Maximum number of open connections (increased for production)
	db.SetMaxIdleConns(10)                  // Maximum number of idle connections (increased for production)
	db.SetConnMaxLifetime(10 * time.Minute) // Maximum connection lifetime (increased for production)
	db.SetConnMaxIdleTime(5 * time.Minute)  // Maximum idle time for connections

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize GORM with PostgreSQL
	gormDB, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GORM: %w", err)
	}

	log.Printf("PostgreSQL database connection established: %s:%s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)
	log.Printf("GORM database connection established successfully")
	log.Printf("Database pool configured: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v", 50, 10, 10*time.Minute)

	return &DB{DB: db, GormDB: gormDB}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// RunMigrations runs database migrations
func (db *DB) RunMigrations() error {
	log.Println("Running database migrations...")

	// Create migrations table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Define migrations in order
	migrations := []string{
		createUsersTable,
		createVideosTable,
		createSubscriptionsTable,
		createYouTubeVideosTable,
		createUserSessionsTable,
		createAuditLogsTable,
		addUserProfileFields,
		addSessionManagementFields,
		createSessionIndexes,
		createCommentsTable,
		createLikesTable,
		createFavoritesTable,
		createUserActivityTable,
		createAdminLogsTable,
		createAdvertiserAccountsTable,
		createAdCampaignsTable,
		createAdvertisementsTable,
		createAdPlacementsTable,
		createAdSchedulesTable,
		createAdAnalyticsTable,
		createAdClicksTable,
		createAdImpressionsTable,
		createAdBillingTable,
		createAdAuditLogTable,
		createIndexes,
		applyPerformanceOptimizations, // Add the new optimization migration
	}

	for i, migration := range migrations {
		migrationName := fmt.Sprintf("migration_%d", i+1)

		// Check if migration already applied
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM migrations WHERE name = $1", migrationName).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if count == 0 {
			// Run migration
			if _, err := db.Exec(migration); err != nil {
				return fmt.Errorf("failed to run migration %s: %w", migrationName, err)
			}

			// Record migration
			if _, err := db.Exec("INSERT INTO migrations (name) VALUES ($1)", migrationName); err != nil {
				return fmt.Errorf("failed to record migration: %w", err)
			}

			log.Printf("Applied migration: %s", migrationName)
		}
	}

	// Seed default ad placements after migrations
	if err := db.SeedAdPlacements(); err != nil {
		log.Printf("Warning: Failed to seed ad placements: %v", err)
	}

	log.Println("Database migrations completed")
	return nil
}

// Migration SQL statements - PostgreSQL compatible
const createUsersTable = `
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
    last_logout TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createVideosTable = `
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
`

const createSubscriptionsTable = `
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
`

const createAdvertisementsTable = `
CREATE TABLE IF NOT EXISTS advertisements (
    id SERIAL PRIMARY KEY,
    campaign_id INTEGER REFERENCES ad_campaigns(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    image_url VARCHAR(500),
    click_url VARCHAR(500) NOT NULL,
    ad_type VARCHAR(50) DEFAULT 'banner',
    width INTEGER DEFAULT 728,
    height INTEGER DEFAULT 90,
    priority INTEGER DEFAULT 1,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createYouTubeVideosTable = `
CREATE TABLE IF NOT EXISTS youtube_videos (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    video_id VARCHAR(255) UNIQUE NOT NULL,
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
`

const createUserSessionsTable = `
CREATE TABLE IF NOT EXISTS user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    session_id VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createAuditLogsTable = `
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    target_type VARCHAR(50),
    target_id INTEGER,
    details JSONB,
    ip_address INET,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const addUserProfileFields = `
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'bio'
    ) THEN
        ALTER TABLE users ADD COLUMN bio TEXT;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'location'
    ) THEN
        ALTER TABLE users ADD COLUMN location VARCHAR(255);
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'website'
    ) THEN
        ALTER TABLE users ADD COLUMN website VARCHAR(500);
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'phone'
    ) THEN
        ALTER TABLE users ADD COLUMN phone VARCHAR(50);
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'avatar_url'
    ) THEN
        ALTER TABLE users ADD COLUMN avatar_url VARCHAR(500);
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'preferences'
    ) THEN
        ALTER TABLE users ADD COLUMN preferences JSONB DEFAULT '{}';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'last_login'
    ) THEN
        ALTER TABLE users ADD COLUMN last_login TIMESTAMP;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'last_logout'
    ) THEN
        ALTER TABLE users ADD COLUMN last_logout TIMESTAMP;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'max_sessions'
    ) THEN
        ALTER TABLE users ADD COLUMN max_sessions INTEGER DEFAULT 5;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'is_active'
    ) THEN
        ALTER TABLE users ADD COLUMN is_active BOOLEAN DEFAULT TRUE;
    END IF;
END $$;
`

const addSessionManagementFields = `
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'user_sessions' AND column_name = 'token_id'
    ) THEN
        ALTER TABLE user_sessions ADD COLUMN token_id VARCHAR(255);
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'user_sessions' AND column_name = 'device_info'
    ) THEN
        ALTER TABLE user_sessions ADD COLUMN device_info TEXT;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'user_sessions' AND column_name = 'ip_address'
    ) THEN
        ALTER TABLE user_sessions ADD COLUMN ip_address INET;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'user_sessions' AND column_name = 'user_agent'
    ) THEN
        ALTER TABLE user_sessions ADD COLUMN user_agent TEXT;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'user_sessions' AND column_name = 'last_activity'
    ) THEN
        ALTER TABLE user_sessions ADD COLUMN last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'user_sessions' AND column_name = 'is_active'
    ) THEN
        ALTER TABLE user_sessions ADD COLUMN is_active BOOLEAN DEFAULT TRUE;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'user_sessions' AND column_name = 'expires_at'
    ) THEN
        ALTER TABLE user_sessions ADD COLUMN expires_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP + INTERVAL '24 hours');
    END IF;
END $$;
`

const createSessionIndexes = `
-- User sessions indexes (basic columns that exist in createUserSessionsTable)
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);

-- User sessions indexes (for columns added by addSessionManagementFields)
CREATE INDEX IF NOT EXISTS idx_user_sessions_token_id ON user_sessions(token_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_active ON user_sessions(is_active);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_activity ON user_sessions(last_activity);

-- Audit logs indexes
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);
`

const createCommentsTable = `
CREATE TABLE IF NOT EXISTS comments (
    id SERIAL PRIMARY KEY,
    video_id INTEGER REFERENCES videos(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    parent_id INTEGER REFERENCES comments(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createLikesTable = `
CREATE TABLE IF NOT EXISTS likes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    video_id INTEGER REFERENCES videos(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, video_id)
);
`

const createFavoritesTable = `
CREATE TABLE IF NOT EXISTS favorites (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    video_id INTEGER REFERENCES videos(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, video_id)
);
`

const createUserActivityTable = `
CREATE TABLE IF NOT EXISTS user_activity (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    activity_type VARCHAR(50) NOT NULL,
    activity_data JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createAdminLogsTable = `
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
`

const createAdvertiserAccountsTable = `
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
`

const createAdCampaignsTable = `
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
`

const createAdPlacementsTable = `
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
`

const createAdSchedulesTable = `
CREATE TABLE IF NOT EXISTS ad_schedules (
    id SERIAL PRIMARY KEY,
    ad_id INTEGER REFERENCES advertisements(id) ON DELETE CASCADE,
    placement_id INTEGER REFERENCES ad_placements(id) ON DELETE CASCADE,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    priority INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createAdAnalyticsTable = `
CREATE TABLE IF NOT EXISTS ad_analytics (
    id SERIAL PRIMARY KEY,
    ad_id INTEGER REFERENCES advertisements(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    impressions INTEGER DEFAULT 0,
    clicks INTEGER DEFAULT 0,
    revenue DECIMAL(10,2) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createAdClicksTable = `
CREATE TABLE IF NOT EXISTS ad_clicks (
    id SERIAL PRIMARY KEY,
    ad_id INTEGER REFERENCES advertisements(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id),
    ip_address INET,
    user_agent TEXT,
    clicked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createAdImpressionsTable = `
CREATE TABLE IF NOT EXISTS ad_impressions (
    id SERIAL PRIMARY KEY,
    ad_id INTEGER REFERENCES advertisements(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id),
    ip_address INET,
    user_agent TEXT,
    viewed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createAdBillingTable = `
CREATE TABLE IF NOT EXISTS ad_billing (
    id SERIAL PRIMARY KEY,
    campaign_id INTEGER REFERENCES ad_campaigns(id) ON DELETE CASCADE,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    impressions INTEGER DEFAULT 0,
    clicks INTEGER DEFAULT 0,
    amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    stripe_invoice_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createAdAuditLogTable = `
CREATE TABLE IF NOT EXISTS ad_audit_log (
    id SERIAL PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,
    entity_id INTEGER NOT NULL,
    actor_id INTEGER REFERENCES users(id),
    actor_type VARCHAR(50) DEFAULT 'user',
    action VARCHAR(100) NOT NULL,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createIndexes = `
-- User indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_email_verified ON users(email_verified);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

-- Video indexes
CREATE INDEX IF NOT EXISTS idx_videos_status ON videos(status);
CREATE INDEX IF NOT EXISTS idx_videos_category ON videos(category);
CREATE INDEX IF NOT EXISTS idx_videos_created_at ON videos(created_at);
CREATE INDEX IF NOT EXISTS idx_videos_created_by ON videos(created_by);
CREATE INDEX IF NOT EXISTS idx_videos_bunny_id ON videos(bunny_video_id);

-- Subscription indexes
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_stripe_id ON subscriptions(stripe_subscription_id);

-- Comment indexes
CREATE INDEX IF NOT EXISTS idx_comments_video_id ON comments(video_id);
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);
CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments(parent_id);
CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments(created_at);

-- Like indexes
CREATE INDEX IF NOT EXISTS idx_likes_user_id ON likes(user_id);
CREATE INDEX IF NOT EXISTS idx_likes_video_id ON likes(video_id);

-- Favorite indexes
CREATE INDEX IF NOT EXISTS idx_favorites_user_id ON favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_favorites_video_id ON favorites(video_id);

-- User activity indexes
CREATE INDEX IF NOT EXISTS idx_user_activity_user_id ON user_activity(user_id);
CREATE INDEX IF NOT EXISTS idx_user_activity_type ON user_activity(activity_type);
CREATE INDEX IF NOT EXISTS idx_user_activity_created_at ON user_activity(created_at);

-- Admin logs indexes
CREATE INDEX IF NOT EXISTS idx_admin_logs_admin_user_id ON admin_logs(admin_user_id);
CREATE INDEX IF NOT EXISTS idx_admin_logs_created_at ON admin_logs(created_at);

-- User sessions indexes (basic columns that exist in createUserSessionsTable)
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);

-- Audit logs indexes
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- Advertiser indexes
CREATE INDEX IF NOT EXISTS idx_advertiser_accounts_user_id ON advertiser_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_advertiser_accounts_status ON advertiser_accounts(status);
CREATE INDEX IF NOT EXISTS idx_advertiser_accounts_business_email ON advertiser_accounts(business_email);

-- Ad campaign indexes
CREATE INDEX IF NOT EXISTS idx_ad_campaigns_advertiser_id ON ad_campaigns(advertiser_id);
CREATE INDEX IF NOT EXISTS idx_ad_campaigns_status ON ad_campaigns(status);
CREATE INDEX IF NOT EXISTS idx_ad_campaigns_dates ON ad_campaigns(start_date, end_date);

-- Advertisement indexes
CREATE INDEX IF NOT EXISTS idx_advertisements_campaign_id ON advertisements(campaign_id);
CREATE INDEX IF NOT EXISTS idx_advertisements_status ON advertisements(status);
CREATE INDEX IF NOT EXISTS idx_advertisements_type ON advertisements(ad_type);

-- Ad placement indexes
CREATE INDEX IF NOT EXISTS idx_ad_placements_location ON ad_placements(location);
CREATE INDEX IF NOT EXISTS idx_ad_placements_active ON ad_placements(is_active);

-- Ad schedule indexes
CREATE INDEX IF NOT EXISTS idx_ad_schedules_ad_id ON ad_schedules(ad_id);
CREATE INDEX IF NOT EXISTS idx_ad_schedules_dates ON ad_schedules(start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_ad_schedules_active ON ad_schedules(is_active);

-- Ad analytics indexes
CREATE INDEX IF NOT EXISTS idx_ad_analytics_ad_id ON ad_analytics(ad_id);
CREATE INDEX IF NOT EXISTS idx_ad_analytics_date ON ad_analytics(date);

-- Ad click indexes
CREATE INDEX IF NOT EXISTS idx_ad_clicks_ad_id ON ad_clicks(ad_id);
CREATE INDEX IF NOT EXISTS idx_ad_clicks_clicked_at ON ad_clicks(clicked_at);

-- Ad impression indexes
CREATE INDEX IF NOT EXISTS idx_ad_impressions_ad_id ON ad_impressions(ad_id);
CREATE INDEX IF NOT EXISTS idx_ad_impressions_viewed_at ON ad_impressions(viewed_at);

-- Ad billing indexes
CREATE INDEX IF NOT EXISTS idx_ad_billing_campaign_id ON ad_billing(campaign_id);
CREATE INDEX IF NOT EXISTS idx_ad_billing_status ON ad_billing(status);
CREATE INDEX IF NOT EXISTS idx_ad_billing_period ON ad_billing(period_start, period_end);

-- Ad audit log indexes
CREATE INDEX IF NOT EXISTS idx_ad_audit_log_entity ON ad_audit_log(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_ad_audit_log_actor ON ad_audit_log(actor_id, actor_type);
CREATE INDEX IF NOT EXISTS idx_ad_audit_log_created_at ON ad_audit_log(created_at);
`

// Performance optimizations migration
const applyPerformanceOptimizations = `
-- =====================================================
-- COMPOSITE INDEXES FOR COMMON QUERY PATTERNS
-- =====================================================

-- Video-related indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_videos_status_category_created 
ON videos(status, category, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_videos_ready_views 
ON videos(status, view_count DESC) WHERE status = 'ready';

-- Enable pg_trgm extension for better text search if not exists
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Full-text search index for videos
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_videos_search_text 
ON videos USING gin(to_tsvector('english', title || ' ' || COALESCE(description, '')));

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
        INSERT INTO user_activity (user_id, activity_type, activity_data, created_at)
        VALUES (user_id_param, 'video_view', 
                json_build_object('video_id', video_id_param), NOW());
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
    COALESCE(SUM(view_count), 0) as total_views,
    COALESCE(SUM(like_count), 0) as total_likes,
    COALESCE(AVG(duration), 0) as avg_duration
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
-- PERFORMANCE MONITORING VIEWS
-- =====================================================

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
`
