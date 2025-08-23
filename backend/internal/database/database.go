package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"bome-backend/internal/config"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB wraps the database connection
type DB struct {
	*sql.DB
	GormDB *gorm.DB // Add GORM support for design system features
	Redis  *Redis   // Add Redis client for caching and session management
}

func (db *DB) UpdateTagCategories(tagID int, categoryIDs []int) error {
	// Start a transaction for atomicity
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// First, clear the tag's current category_id
	clearQuery := `
		UPDATE tags 
		SET category_id = NULL, updated_at = NOW()
		WHERE id = $1
	`
	_, err = tx.Exec(clearQuery, tagID)
	if err != nil {
		return fmt.Errorf("failed to clear tag category_id: %v", err)
	}

	// If categoryIDs is provided and not empty, set the first one as the primary category_id
	if len(categoryIDs) > 0 {
		updateQuery := `
			UPDATE tags 
			SET category_id = $2, updated_at = NOW()
			WHERE id = $1
		`
		_, err = tx.Exec(updateQuery, tagID, categoryIDs[0])
		if err != nil {
			return fmt.Errorf("failed to update tag category_id: %v", err)
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
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

	// Configure connection pool for high-traffic production load
	// Optimized for 3,000-5,000 concurrent users
	db.SetMaxOpenConns(200)                 // Increased from 50 to 200 for high concurrency
	db.SetMaxIdleConns(50)                  // Increased from 10 to 50 for better reuse
	db.SetConnMaxLifetime(30 * time.Minute) // Increased from 10 to 30 minutes
	db.SetConnMaxIdleTime(10 * time.Minute) // Increased from 5 to 10 minutes

	// Add connection pool monitoring
	log.Printf("Database pool configured: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v, MaxIdleTime=%v",
		200, 50, 30*time.Minute, 10*time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize GORM with PostgreSQL
	gormDB, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GORM: %w", err)
	}

	// Initialize Redis connection
	redisClient, err := NewRedis(cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize Redis: %v", err)
		redisClient = nil
	}

	log.Printf("PostgreSQL database connection established: %s:%s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)
	log.Printf("GORM database connection established successfully")
	if redisClient != nil {
		log.Printf("Redis connection established successfully")
	}
	log.Printf("Database pool configured: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v", 50, 10, 10*time.Minute)

	return &DB{DB: db, GormDB: gormDB, Redis: redisClient}, nil
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
		createAnalyticsTables, // Add analytics tables migration
		createIndexes,
		applyPerformanceOptimizations,     // Add the new optimization migration
		createMasterVideoList,             // Add master video list migration
		addShortDescToSubscriptionPlans,   // Add short_desc column to existing subscription_plans table
		addMissingSubscriptionPlanColumns, // Add missing columns for subscription plans
		createSubscriberHistoryTable,      // Add subscriber history table
		createSecureSettingsTable,         // Add secure settings table for encrypted config
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

// GetPoolStats returns database connection pool statistics
func (db *DB) GetPoolStats() map[string]interface{} {
	stats := db.DB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

// GetRedisClient returns the Redis client
func (db *DB) GetRedisClient() *redis.Client {
	if db.Redis == nil {
		return nil
	}
	return db.Redis.Client
}

// CreateAlert creates a new system alert
func (db *DB) CreateAlert(alert *Alert) error {
	query := `
		INSERT INTO alerts (severity, title, message, created_at, acknowledged)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := db.QueryRow(
		query,
		alert.Severity,
		alert.Title,
		alert.Message,
		alert.CreatedAt,
		alert.Acknowledged,
	).Scan(&alert.ID)

	if err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	return nil
}

// Secure settings table stores encrypted key-value configuration
const createSecureSettingsTable = `
CREATE TABLE IF NOT EXISTS secure_settings (
    id SERIAL PRIMARY KEY,
    key VARCHAR(255) UNIQUE NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

// SetSecureSetting stores or updates an encrypted secure setting value
func (db *DB) SetSecureSetting(key, encryptedValue string) error {
	if db == nil || db.DB == nil {
		return fmt.Errorf("database not initialized")
	}
	// Upsert behavior
	_, err := db.Exec(`
        INSERT INTO secure_settings (key, value) VALUES ($1, $2)
        ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP
    `, key, encryptedValue)
	return err
}

// GetSecureSetting retrieves an encrypted setting by key
func (db *DB) GetSecureSetting(key string) (string, error) {
	if db == nil || db.DB == nil {
		return "", fmt.Errorf("database not initialized")
	}
	var value string
	err := db.QueryRow(`SELECT value FROM secure_settings WHERE key = $1`, key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return value, nil
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

const createAnalyticsTables = `
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
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User Metrics Table
CREATE TABLE IF NOT EXISTS user_metrics (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    session_count INTEGER DEFAULT 0,
    session_duration INTEGER DEFAULT 0, -- in seconds
    page_views INTEGER DEFAULT 0,
    video_views INTEGER DEFAULT 0,
    video_watch_time INTEGER DEFAULT 0, -- in seconds
    likes_given INTEGER DEFAULT 0,
    comments_made INTEGER DEFAULT 0,
    shares_made INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, date)
);

-- Video Metrics Table
CREATE TABLE IF NOT EXISTS video_metrics (
    id SERIAL PRIMARY KEY,
    video_id INTEGER REFERENCES videos(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    views INTEGER DEFAULT 0,
    unique_views INTEGER DEFAULT 0,
    watch_time INTEGER DEFAULT 0, -- in seconds
    completion_rate DECIMAL(5,2) DEFAULT 0.00,
    likes INTEGER DEFAULT 0,
    comments INTEGER DEFAULT 0,
    shares INTEGER DEFAULT 0,
    bounce_rate DECIMAL(5,2) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(video_id, date)
);

-- System Metrics Table
CREATE TABLE IF NOT EXISTS system_metrics (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL,
    cpu_usage DECIMAL(5,2) DEFAULT 0.00,
    memory_usage DECIMAL(5,2) DEFAULT 0.00,
    disk_usage DECIMAL(5,2) DEFAULT 0.00,
    network_in BIGINT DEFAULT 0, -- bytes
    network_out BIGINT DEFAULT 0, -- bytes
    active_sessions INTEGER DEFAULT 0,
    error_rate DECIMAL(5,4) DEFAULT 0.0000,
    response_time INTEGER DEFAULT 0, -- milliseconds
    database_size BIGINT DEFAULT 0, -- bytes
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Webhook Events Table
CREATE TABLE IF NOT EXISTS webhook_events (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    subsite VARCHAR(50) DEFAULT 'streaming',
    endpoint VARCHAR(500) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('success', 'failed', 'pending')),
    response_time INTEGER DEFAULT 0, -- milliseconds
    payload_size INTEGER DEFAULT 0, -- bytes
    status_code INTEGER,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Alerts Table
CREATE TABLE IF NOT EXISTS alerts (
    id SERIAL PRIMARY KEY,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('info', 'warning', 'critical')),
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    subsite VARCHAR(50),
    acknowledged BOOLEAN DEFAULT FALSE,
    acknowledged_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    acknowledged_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Cross Subsite Stats Table
CREATE TABLE IF NOT EXISTS cross_subsite_stats (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL,
    subsite VARCHAR(50) NOT NULL,
    users INTEGER DEFAULT 0,
    content INTEGER DEFAULT 0,
    views INTEGER DEFAULT 0,
    revenue DECIMAL(10,2) DEFAULT 0.00,
    engagement_rate DECIMAL(5,4) DEFAULT 0.0000,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(date, subsite)
);

-- Analytics Indexes
CREATE INDEX IF NOT EXISTS idx_analytics_events_type ON analytics_events(event_type);
CREATE INDEX IF NOT EXISTS idx_analytics_events_user_id ON analytics_events(user_id);
CREATE INDEX IF NOT EXISTS idx_analytics_events_subsite ON analytics_events(subsite);
CREATE INDEX IF NOT EXISTS idx_analytics_events_created_at ON analytics_events(created_at);

CREATE INDEX IF NOT EXISTS idx_user_metrics_user_id ON user_metrics(user_id);
CREATE INDEX IF NOT EXISTS idx_user_metrics_date ON user_metrics(date);

CREATE INDEX IF NOT EXISTS idx_video_metrics_video_id ON video_metrics(video_id);
CREATE INDEX IF NOT EXISTS idx_video_metrics_date ON video_metrics(date);

CREATE INDEX IF NOT EXISTS idx_system_metrics_timestamp ON system_metrics(timestamp);

CREATE INDEX IF NOT EXISTS idx_webhook_events_type ON webhook_events(event_type);
CREATE INDEX IF NOT EXISTS idx_webhook_events_status ON webhook_events(status);
CREATE INDEX IF NOT EXISTS idx_webhook_events_created_at ON webhook_events(created_at);

CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);
CREATE INDEX IF NOT EXISTS idx_alerts_acknowledged ON alerts(acknowledged);
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at);

CREATE INDEX IF NOT EXISTS idx_cross_subsite_stats_date ON cross_subsite_stats(date);
CREATE INDEX IF NOT EXISTS idx_cross_subsite_stats_subsite ON cross_subsite_stats(subsite);
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
CREATE INDEX IF NOT EXISTS idx_videos_status_category_created 
ON videos(status, category, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_videos_ready_views 
ON videos(status, view_count DESC) WHERE status = 'ready';

-- Enable pg_trgm extension for better text search if not exists
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Full-text search index for videos
CREATE INDEX IF NOT EXISTS idx_videos_search_text 
ON videos USING gin(to_tsvector('english', title || ' ' || COALESCE(description, '')));

-- User session optimization
CREATE INDEX IF NOT EXISTS idx_user_sessions_active_expires 
ON user_sessions(is_active, expires_at) WHERE is_active = TRUE;

CREATE INDEX IF NOT EXISTS idx_user_sessions_user_activity 
ON user_sessions(user_id, last_activity DESC) WHERE is_active = TRUE;

-- Analytics and audit logs
CREATE INDEX IF NOT EXISTS idx_audit_logs_composite 
ON audit_logs(user_id, action, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_user_activity_type_created 
ON user_activity(activity_type, created_at DESC);

-- Advertisement system indexes
CREATE INDEX IF NOT EXISTS idx_ad_campaigns_active_dates 
ON ad_campaigns(status, start_date, end_date) WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_ad_schedules_active_time 
ON ad_schedules(is_active, start_date, end_date) WHERE is_active = TRUE;

-- =====================================================
-- ANALYTICS-SPECIFIC INDEXES FOR PERFORMANCE
-- =====================================================

-- Analytics events indexes for fast querying
CREATE INDEX IF NOT EXISTS idx_analytics_events_subsite_created 
ON analytics_events(subsite, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_analytics_events_type_created 
ON analytics_events(event_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_analytics_events_user_session 
ON analytics_events(user_id, session_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_analytics_events_video_view 
ON analytics_events(event_type, created_at DESC) WHERE event_type = 'video_view';

-- User metrics indexes
CREATE INDEX IF NOT EXISTS idx_user_metrics_user_date 
ON user_metrics(user_id, date DESC);

CREATE INDEX IF NOT EXISTS idx_user_metrics_date_range 
ON user_metrics(date) WHERE date >= CURRENT_DATE - INTERVAL '90 days';

-- Video metrics indexes
CREATE INDEX IF NOT EXISTS idx_video_metrics_video_date 
ON video_metrics(video_id, date DESC);

CREATE INDEX IF NOT EXISTS idx_video_metrics_views_date 
ON video_metrics(views DESC, date DESC);

-- System metrics indexes
CREATE INDEX IF NOT EXISTS idx_system_metrics_timestamp 
ON system_metrics(timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_system_metrics_recent 
ON system_metrics(timestamp DESC) WHERE timestamp >= NOW() - INTERVAL '24 hours';

-- Webhook events indexes
CREATE INDEX IF NOT EXISTS idx_webhook_events_status_created 
ON webhook_events(status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_webhook_events_subsite_type 
ON webhook_events(subsite, event_type, created_at DESC);

-- Alerts indexes
CREATE INDEX IF NOT EXISTS idx_alerts_severity_created 
ON alerts(severity, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_alerts_unacknowledged 
ON alerts(acknowledged, created_at DESC) WHERE acknowledged = FALSE;

-- Cross-subsite stats indexes
CREATE INDEX IF NOT EXISTS idx_cross_subsite_stats_date_subsite 
ON cross_subsite_stats(date DESC, subsite);

-- Users table indexes for analytics queries
CREATE INDEX IF NOT EXISTS idx_users_created_at 
ON users(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_users_role_created 
ON users(role, created_at DESC);

-- Subscriptions indexes for revenue analytics
CREATE INDEX IF NOT EXISTS idx_subscriptions_status_created 
ON subscriptions(status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_subscriptions_billing_cycle 
ON subscriptions(billing_cycle, status) WHERE status = 'active';

-- Payments indexes for revenue tracking
CREATE INDEX IF NOT EXISTS idx_payments_status_created 
ON payments(status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_payments_amount_status 
ON payments(amount DESC, status) WHERE status = 'completed';

-- Video ratings indexes
CREATE INDEX IF NOT EXISTS idx_video_ratings_video_rating 
ON video_ratings(video_id, rating DESC);

CREATE INDEX IF NOT EXISTS idx_video_ratings_user_video 
ON video_ratings(user_id, video_id);

-- =====================================================
-- PARTITIONING FOR LARGE ANALYTICS TABLES
-- =====================================================

-- Partition analytics_events by month for better performance
-- Note: This requires PostgreSQL 10+ and careful planning
-- Uncomment when table size exceeds 10M rows

/*
CREATE TABLE IF NOT EXISTS analytics_events_partitioned (
    LIKE analytics_events INCLUDING ALL
) PARTITION BY RANGE (created_at);

-- Create partitions for current and next 3 months
CREATE TABLE analytics_events_2024_01 PARTITION OF analytics_events_partitioned
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

CREATE TABLE analytics_events_2024_02 PARTITION OF analytics_events_partitioned
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

CREATE TABLE analytics_events_2024_03 PARTITION OF analytics_events_partitioned
    FOR VALUES FROM ('2024-03-01') TO ('2024-04-01');

CREATE TABLE analytics_events_2024_04 PARTITION OF analytics_events_partitioned
    FOR VALUES FROM ('2024-04-01') TO ('2024-05-01');
*/

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

const createMasterVideoList = `
-- Migration: Create master video list table
-- Description: Master list for video metadata with Bunny.net synchronization

CREATE TABLE IF NOT EXISTS master_video_list (
    id SERIAL PRIMARY KEY,
    bunny_video_id VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    tags JSONB DEFAULT '[]',
    tagged BOOLEAN DEFAULT FALSE,
    duration INTEGER DEFAULT 0,
    file_size BIGINT DEFAULT 0,
    resolution VARCHAR(50),
    framerate DECIMAL(5,2),
    thumbnail_url TEXT,
    video_url TEXT,
    iframe_src TEXT,
    playback_url TEXT,
    status VARCHAR(50) DEFAULT 'processing',
    views INTEGER DEFAULT 0,
    likes INTEGER DEFAULT 0,
    is_public BOOLEAN DEFAULT true,
    encode_progress INTEGER DEFAULT 0,
    available_resolutions JSONB DEFAULT '[]',
    collection_id VARCHAR(255),
    average_watch_time INTEGER DEFAULT 0,
    total_watch_time BIGINT DEFAULT 0,
    
    -- Sync tracking fields
    last_bunny_sync TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_master_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    sync_status VARCHAR(50) DEFAULT 'synced', -- synced, needs_attention, conflict
    sync_notes TEXT,
    
    -- Metadata tracking
    metadata_version INTEGER DEFAULT 1,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_master_video_bunny_id ON master_video_list(bunny_video_id);
CREATE INDEX IF NOT EXISTS idx_master_video_status ON master_video_list(status);
CREATE INDEX IF NOT EXISTS idx_master_video_category ON master_video_list(category);
CREATE INDEX IF NOT EXISTS idx_master_video_sync_status ON master_video_list(sync_status);
CREATE INDEX IF NOT EXISTS idx_master_video_created_at ON master_video_list(created_at);
CREATE INDEX IF NOT EXISTS idx_master_video_views ON master_video_list(views DESC);
CREATE INDEX IF NOT EXISTS idx_master_video_collection ON master_video_list(collection_id);
CREATE INDEX IF NOT EXISTS idx_master_video_tagged ON master_video_list(tagged);

-- Create video tags table for word analytics (updated schema)
CREATE TABLE IF NOT EXISTS video_tags (
    id SERIAL PRIMARY KEY,
    word VARCHAR(100) NOT NULL UNIQUE,
    frequency INTEGER DEFAULT 1,
    category_id INTEGER,
    subsite_id INTEGER REFERENCES subsites(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create tag categories table for classification
CREATE TABLE IF NOT EXISTS tag_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    color VARCHAR(7) DEFAULT '#6b7280',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for video tags
CREATE INDEX IF NOT EXISTS idx_video_tags_word ON video_tags(word);
CREATE INDEX IF NOT EXISTS idx_video_tags_frequency ON video_tags(frequency DESC);
CREATE INDEX IF NOT EXISTS idx_video_tags_category ON video_tags(category_id);

-- Create indexes for tag categories
CREATE INDEX IF NOT EXISTS idx_tag_categories_name ON tag_categories(name);

-- Insert default tag categories
INSERT INTO tag_categories (name, description, color) VALUES
    ('Archaeology', 'Archaeological terms and concepts', '#8b5cf6'),
    ('Geography', 'Geographic locations and features', '#06b6d4'),
    ('DNA Research', 'Genetic and DNA-related terms', '#10b981'),
    ('Linguistics', 'Language and linguistic terms', '#f59e0b'),
    ('Historical Evidence', 'Historical documentation and evidence', '#ef4444'),
    ('Cultural Studies', 'Cultural and anthropological terms', '#ec4899'),
    ('Religious Studies', 'Religious and theological terms', '#6366f1'),
    ('Documentary', 'Documentary and media terms', '#84cc16'),
    ('Lecture', 'Educational and lecture terms', '#f97316'),
    ('Interview', 'Interview and discussion terms', '#06b6d4'),
    ('Presentation', 'Presentation and presentation terms', '#8b5cf6'),
    ('Virtual Tour', 'Tour and exploration terms', '#10b981')
ON CONFLICT (name) DO NOTHING;

-- Create sync conflicts table for tracking discrepancies
CREATE TABLE IF NOT EXISTS video_sync_conflicts (
    id SERIAL PRIMARY KEY,
    master_video_id INTEGER REFERENCES master_video_list(id) ON DELETE CASCADE,
    bunny_video_id VARCHAR(255) NOT NULL,
    conflict_type VARCHAR(50) NOT NULL, -- field_mismatch, missing_field, status_mismatch
    field_name VARCHAR(100),
    master_value TEXT,
    bunny_value TEXT,
    proposed_action VARCHAR(50) NOT NULL, -- update_master, update_bunny, update_both, manual_review
    admin_notes TEXT,
    resolved BOOLEAN DEFAULT false,
    resolved_by INTEGER REFERENCES users(id),
    resolved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for sync conflicts
CREATE INDEX IF NOT EXISTS idx_sync_conflicts_master_id ON video_sync_conflicts(master_video_id);
CREATE INDEX IF NOT EXISTS idx_sync_conflicts_bunny_id ON video_sync_conflicts(bunny_video_id);
CREATE INDEX IF NOT EXISTS idx_sync_conflicts_resolved ON video_sync_conflicts(resolved);
CREATE INDEX IF NOT EXISTS idx_sync_conflicts_type ON video_sync_conflicts(conflict_type);

-- Create sync audit log table
CREATE TABLE IF NOT EXISTS video_sync_audit_log (
    id SERIAL PRIMARY KEY,
    master_video_id INTEGER REFERENCES master_video_list(id) ON DELETE CASCADE,
    bunny_video_id VARCHAR(255) NOT NULL,
    sync_action VARCHAR(50) NOT NULL, -- sync_from_bunny, sync_to_bunny, conflict_resolved, manual_update
    sync_result VARCHAR(50) NOT NULL, -- success, failed, partial, conflict
    changes_made JSONB DEFAULT '{}',
    error_message TEXT,
    performed_by INTEGER REFERENCES users(id),
    performed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for audit log
CREATE INDEX IF NOT EXISTS idx_sync_audit_master_id ON video_sync_audit_log(master_video_id);
CREATE INDEX IF NOT EXISTS idx_sync_audit_bunny_id ON video_sync_audit_log(bunny_video_id);
CREATE INDEX IF NOT EXISTS idx_sync_audit_action ON video_sync_audit_log(sync_action);
CREATE INDEX IF NOT EXISTS idx_sync_audit_performed_at ON video_sync_audit_log(performed_at);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_master_video_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for updated_at
DROP TRIGGER IF EXISTS trigger_master_video_updated_at ON master_video_list;
CREATE TRIGGER trigger_master_video_updated_at
    BEFORE UPDATE ON master_video_list
    FOR EACH ROW
    EXECUTE FUNCTION update_master_video_updated_at();

-- Create function to log sync conflicts
CREATE OR REPLACE FUNCTION log_sync_conflict(
    p_master_video_id INTEGER,
    p_bunny_video_id VARCHAR(255),
    p_conflict_type VARCHAR(50),
    p_field_name VARCHAR(100),
    p_master_value TEXT,
    p_bunny_value TEXT,
    p_proposed_action VARCHAR(50)
) RETURNS INTEGER AS $$
DECLARE
    conflict_id INTEGER;
BEGIN
    INSERT INTO video_sync_conflicts (
        master_video_id, bunny_video_id, conflict_type, field_name,
        master_value, bunny_value, proposed_action
    ) VALUES (
        p_master_video_id, p_bunny_video_id, p_conflict_type, p_field_name,
        p_master_value, p_bunny_value, p_proposed_action
    ) RETURNING id INTO conflict_id;
    
    RETURN conflict_id;
END;
$$ LANGUAGE plpgsql;

-- Create function to log sync audit
CREATE OR REPLACE FUNCTION log_sync_audit(
    p_master_video_id INTEGER,
    p_bunny_video_id VARCHAR(255),
    p_sync_action VARCHAR(50),
    p_sync_result VARCHAR(50),
    p_changes_made JSONB,
    p_error_message TEXT,
    p_performed_by INTEGER
) RETURNS INTEGER AS $$
DECLARE
    audit_id INTEGER;
BEGIN
    INSERT INTO video_sync_audit_log (
        master_video_id, bunny_video_id, sync_action, sync_result,
        changes_made, error_message, performed_by
    ) VALUES (
        p_master_video_id, p_bunny_video_id, p_sync_action, p_sync_result,
        p_changes_made, p_error_message, p_performed_by
    ) RETURNING id INTO audit_id;
    
    RETURN audit_id;
END;
$$ LANGUAGE plpgsql;

-- Create view for easy access to sync status
CREATE OR REPLACE VIEW video_sync_status AS
SELECT 
    mvl.id,
    mvl.bunny_video_id,
    mvl.title,
    mvl.status as master_status,
    mvl.sync_status,
    mvl.last_bunny_sync,
    mvl.last_master_update,
    COUNT(vsc.id) as pending_conflicts,
    COUNT(vsc.id) FILTER (WHERE vsc.resolved = false) as unresolved_conflicts
FROM master_video_list mvl
LEFT JOIN video_sync_conflicts vsc ON mvl.id = vsc.master_video_id
GROUP BY mvl.id, mvl.bunny_video_id, mvl.title, mvl.status, mvl.sync_status, 
         mvl.last_bunny_sync, mvl.last_master_update;

-- Create view for admin dashboard
CREATE OR REPLACE VIEW admin_video_dashboard AS
SELECT 
    mvl.id,
    mvl.bunny_video_id,
    mvl.title,
    mvl.category,
    mvl.status,
    mvl.views,
    mvl.duration,
    mvl.file_size,
    mvl.sync_status,
    mvl.last_bunny_sync,
    mvl.last_master_update,
    COUNT(vsc.id) FILTER (WHERE vsc.resolved = false) as needs_attention,
    CASE 
        WHEN mvl.sync_status = 'needs_attention' THEN '⚠️ Needs Review'
        WHEN mvl.sync_status = 'conflict' THEN '🚨 Conflict'
        WHEN mvl.sync_status = 'synced' THEN '✅ Synced'
        ELSE '❓ Unknown'
    END as sync_status_display
FROM master_video_list mvl
LEFT JOIN video_sync_conflicts vsc ON mvl.id = vsc.master_video_id
GROUP BY mvl.id, mvl.bunny_video_id, mvl.title, mvl.category, mvl.status, 
         mvl.views, mvl.duration, mvl.file_size, mvl.sync_status, 
         mvl.last_bunny_sync, mvl.last_master_update;
`

const addShortDescToSubscriptionPlans = `
-- Add short_desc column to existing subscription_plans table
ALTER TABLE subscription_plans 
ADD COLUMN IF NOT EXISTS short_desc VARCHAR(500);

-- Update interval constraint to match actual database values
ALTER TABLE subscription_plans 
DROP CONSTRAINT IF EXISTS subscription_plans_interval_check;

ALTER TABLE subscription_plans 
ADD CONSTRAINT subscription_plans_interval_check 
CHECK (interval IN ('month', 'year', 'week', 'day'));

-- Update existing data to use correct interval values
UPDATE subscription_plans 
SET interval = 'month' 
WHERE interval = 'monthly';

UPDATE subscription_plans 
SET interval = 'year' 
WHERE interval = 'annual';

UPDATE subscription_plans 
SET interval = 'week' 
WHERE interval = 'weekly';

UPDATE subscription_plans 
SET interval = 'day' 
WHERE interval = 'daily';

-- Add default short_desc values for existing plans
UPDATE subscription_plans 
SET short_desc = 'Essential Monthly Access'
WHERE name = 'Essential Monthly' AND short_desc IS NULL;

UPDATE subscription_plans 
SET short_desc = 'Best Value - Save 33%'
WHERE name = 'Premium Annual' AND short_desc IS NULL;

UPDATE subscription_plans 
SET short_desc = 'Complete Library Access'
WHERE name = 'Premium Monthly' AND short_desc IS NULL;

UPDATE subscription_plans 
SET short_desc = 'Maximum Savings - Pro Benefits'
WHERE name = 'Annual Pro' AND short_desc IS NULL;

UPDATE subscription_plans 
SET short_desc = 'Complete Conference Access'
WHERE name = 'Conference + Library Bundle' AND short_desc IS NULL;

UPDATE subscription_plans 
SET short_desc = 'Premium + Conference Benefits'
WHERE name = 'Semi-Annual Premium' AND short_desc IS NULL;

UPDATE subscription_plans 
SET short_desc = 'Ultimate Annual Package'
WHERE name = 'Annual Premium' AND short_desc IS NULL;

UPDATE subscription_plans 
SET short_desc = 'Get Started Today'
WHERE name = 'Starter Monthly' AND short_desc IS NULL;

UPDATE subscription_plans 
SET short_desc = 'Professional Choice'
WHERE name = 'Professional Annual' AND short_desc IS NULL;

-- Set default short_desc for any remaining plans
UPDATE subscription_plans 
SET short_desc = name
WHERE short_desc IS NULL;

-- Add comment for the new column
COMMENT ON COLUMN subscription_plans.short_desc IS 'Short description or tagline for the subscription plan';

-- Update the interval comment
COMMENT ON COLUMN subscription_plans.interval IS 'Billing interval (month, year, week, day)';
`

const addMissingSubscriptionPlanColumns = `
-- Add sub_type column for plan classification (100 = standard, 300 = promotional)
ALTER TABLE subscription_plans 
ADD COLUMN IF NOT EXISTS sub_type INTEGER DEFAULT 100 CHECK (sub_type IN (100, 300));

-- Add promotion_start_date column
ALTER TABLE subscription_plans 
ADD COLUMN IF NOT EXISTS promotion_start_date TIMESTAMP WITH TIME ZONE;

-- Add promotion_history column for tracking promotion changes
ALTER TABLE subscription_plans 
ADD COLUMN IF NOT EXISTS promotion_history JSONB DEFAULT '[]'::jsonb;

-- Add is_deleted column for soft delete status
ALTER TABLE subscription_plans 
ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT false;

-- Update existing plans to have proper sub_type values
UPDATE subscription_plans 
SET sub_type = CASE 
    WHEN is_promoted = true THEN 300 
    ELSE 100 
END
WHERE sub_type IS NULL;

-- Add comments for the new columns
COMMENT ON COLUMN subscription_plans.sub_type IS 'Plan type: 100 = standard plan, 300 = promotional plan';
COMMENT ON COLUMN subscription_plans.promotion_start_date IS 'When the promotion started (NULL if not promoted)';
COMMENT ON COLUMN subscription_plans.promotion_history IS 'JSON array of promotion history events';
COMMENT ON COLUMN subscription_plans.is_deleted IS 'Soft delete flag (true if plan is deleted)';

-- Create index on sub_type for better performance
CREATE INDEX IF NOT EXISTS idx_subscription_plans_sub_type ON subscription_plans(sub_type) WHERE deleted_at IS NULL;
`

const createSubscriberHistoryTable = `
-- Create the subscriber_history table
CREATE TABLE IF NOT EXISTS subscriber_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    usr_sub_hstry JSONB DEFAULT '{}'::jsonb,
    usr_off_hstry JSONB DEFAULT '{}'::jsonb,
    updated_at JSONB DEFAULT '{}'::jsonb,
    notes JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_subscriber_history_user_id ON subscriber_history(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriber_history_created_at ON subscriber_history(created_at);
CREATE INDEX IF NOT EXISTS idx_subscriber_history_usr_sub_hstry ON subscriber_history USING GIN (usr_sub_hstry);
CREATE INDEX IF NOT EXISTS idx_subscriber_history_usr_off_hstry ON subscriber_history USING GIN (usr_off_hstry);
CREATE INDEX IF NOT EXISTS idx_subscriber_history_updated_at ON subscriber_history USING GIN (updated_at);
CREATE INDEX IF NOT EXISTS idx_subscriber_history_notes ON subscriber_history USING GIN (notes);

-- Add table and column comments for documentation
COMMENT ON TABLE subscriber_history IS 'Comprehensive history tracking for subscribers including subscriptions, offers, updates, and notes';
COMMENT ON COLUMN subscriber_history.id IS 'Unique identifier for history record';
COMMENT ON COLUMN subscriber_history.user_id IS 'Foreign key reference to users table';
COMMENT ON COLUMN subscriber_history.usr_sub_hstry IS 'JSONB field storing subscription history entries';
COMMENT ON COLUMN subscriber_history.usr_off_hstry IS 'JSONB field storing offer history entries';
COMMENT ON COLUMN subscriber_history.updated_at IS 'JSONB field storing account update entries';
COMMENT ON COLUMN subscriber_history.notes IS 'JSONB field storing admin, system, and user notes';
COMMENT ON COLUMN subscriber_history.created_at IS 'When this history record was created';

-- Add constraint to ensure user_id is not null
ALTER TABLE subscriber_history 
ADD CONSTRAINT chk_user_id_not_null 
CHECK (user_id IS NOT NULL);

-- Add constraint to ensure JSONB fields are valid JSON when not empty
ALTER TABLE subscriber_history 
ADD CONSTRAINT chk_usr_sub_hstry_jsonb 
CHECK (usr_sub_hstry IS NULL OR jsonb_typeof(usr_sub_hstry) = 'object');

ALTER TABLE subscriber_history 
ADD CONSTRAINT chk_usr_off_hstry_jsonb 
CHECK (usr_off_hstry IS NULL OR jsonb_typeof(usr_off_hstry) = 'object');

ALTER TABLE subscriber_history 
ADD CONSTRAINT chk_updated_at_jsonb 
CHECK (updated_at IS NULL OR jsonb_typeof(updated_at) = 'object');

ALTER TABLE subscriber_history 
ADD CONSTRAINT chk_notes_jsonb 
CHECK (notes IS NULL OR jsonb_typeof(notes) = 'object');
`
