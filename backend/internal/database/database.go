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

	// Configure connection pool for PostgreSQL
	db.SetMaxOpenConns(25)                 // Maximum number of open connections
	db.SetMaxIdleConns(5)                  // Maximum number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum connection lifetime

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
	createMigrationsTable := `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`

	if _, err := db.Exec(createMigrationsTable); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Run migrations in order
	migrations := []string{
		createUsersTable,
		createVideosTable,
		createSubscriptionsTable,
		createCommentsTable,
		createLikesTable,
		createFavoritesTable,
		createUserActivityTable,
		createAdminLogsTable,
		createDatabaseIndexes,
		// Add profile fields migration
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS bio TEXT;
		 ALTER TABLE users ADD COLUMN IF NOT EXISTS location VARCHAR(255);
		 ALTER TABLE users ADD COLUMN IF NOT EXISTS website VARCHAR(500);
		 ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(50);
		 ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500);
		 ALTER TABLE users ADD COLUMN IF NOT EXISTS preferences JSONB DEFAULT '{}';
		 ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login TIMESTAMP;
		 CREATE INDEX IF NOT EXISTS idx_users_last_login ON users(last_login);
		 CREATE INDEX IF NOT EXISTS idx_users_location ON users(location);
		 UPDATE users SET preferences = '{}' WHERE preferences IS NULL;`,
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

const createDatabaseIndexes = `
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_videos_status ON videos(status);
CREATE INDEX IF NOT EXISTS idx_videos_category ON videos(category);
CREATE INDEX IF NOT EXISTS idx_videos_created_at ON videos(created_at);
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_comments_video_id ON comments(video_id);
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);
CREATE INDEX IF NOT EXISTS idx_likes_user_id ON likes(user_id);
CREATE INDEX IF NOT EXISTS idx_likes_video_id ON likes(video_id);
CREATE INDEX IF NOT EXISTS idx_favorites_user_id ON favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_favorites_video_id ON favorites(video_id);
CREATE INDEX IF NOT EXISTS idx_user_activity_user_id ON user_activity(user_id);
CREATE INDEX IF NOT EXISTS idx_user_activity_type ON user_activity(activity_type);
CREATE INDEX IF NOT EXISTS idx_user_activity_created_at ON user_activity(created_at);
CREATE INDEX IF NOT EXISTS idx_admin_logs_admin_user_id ON admin_logs(admin_user_id);
CREATE INDEX IF NOT EXISTS idx_admin_logs_created_at ON admin_logs(created_at);
`
