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
		createAdvertisementsTable,
		createYouTubeVideosTable,
		createUserSessionsTable,
		createAuditLogsTable,
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
