package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"bome-backend/internal/config"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB wraps the database connection
type DB struct {
	*sql.DB
	GormDB *gorm.DB // Add GORM support for design system features
}

// New creates a new database connection
func New(cfg *config.Config) (*DB, error) {
	// Use SQLite for development
	dbPath := "./data/bome.db"

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for SQLite
	db.SetMaxOpenConns(1) // SQLite works best with single connection
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0) // No connection lifetime limit for SQLite

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize GORM with the same database file
	gormDB, err := gorm.Open(sqlite.Open(dbPath+"?_foreign_keys=on&_journal_mode=WAL"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GORM: %w", err)
	}

	log.Printf("SQLite database connection established: %s", dbPath)
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
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
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
		createAdvertisementIndexes,
	}

	for i, migration := range migrations {
		migrationName := fmt.Sprintf("migration_%d", i+1)

		// Check if migration already applied
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM migrations WHERE name = ?", migrationName).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if count == 0 {
			// Run migration
			if _, err := db.Exec(migration); err != nil {
				return fmt.Errorf("failed to run migration %s: %w", migrationName, err)
			}

			// Record migration
			if _, err := db.Exec("INSERT INTO migrations (name) VALUES (?)", migrationName); err != nil {
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

// Migration SQL statements
const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    first_name TEXT,
    last_name TEXT,
    role TEXT DEFAULT 'user',
    email_verified BOOLEAN DEFAULT FALSE,
    stripe_customer_id TEXT,
    reset_token TEXT,
    reset_token_expiry DATETIME,
    verification_token TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

const createVideosTable = `
CREATE TABLE IF NOT EXISTS videos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    bunny_video_id TEXT UNIQUE NOT NULL,
    thumbnail_url TEXT,
    duration INTEGER,
    file_size INTEGER,
    status TEXT DEFAULT 'processing',
    category TEXT,
    tags TEXT,
    view_count INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    created_by INTEGER REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

const createSubscriptionsTable = `
CREATE TABLE IF NOT EXISTS subscriptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    stripe_subscription_id TEXT UNIQUE NOT NULL,
    stripe_price_id TEXT NOT NULL,
    status TEXT NOT NULL,
    current_period_start DATETIME,
    current_period_end DATETIME,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

const createCommentsTable = `
CREATE TABLE IF NOT EXISTS comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    video_id INTEGER REFERENCES videos(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    parent_id INTEGER REFERENCES comments(id) ON DELETE CASCADE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

const createLikesTable = `
CREATE TABLE IF NOT EXISTS likes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    video_id INTEGER REFERENCES videos(id) ON DELETE CASCADE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, video_id)
);
`

const createFavoritesTable = `
CREATE TABLE IF NOT EXISTS favorites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    video_id INTEGER REFERENCES videos(id) ON DELETE CASCADE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, video_id)
);
`

const createUserActivityTable = `
CREATE TABLE IF NOT EXISTS user_activity (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    activity_type TEXT NOT NULL,
    video_id INTEGER REFERENCES videos(id) ON DELETE SET NULL,
    metadata TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

const createAdminLogsTable = `
CREATE TABLE IF NOT EXISTS admin_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    admin_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    resource_type TEXT,
    resource_id INTEGER,
    details TEXT,
    ip_address TEXT,
    user_agent TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`
