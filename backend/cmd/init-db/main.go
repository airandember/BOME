// backend/cmd/init-db/main.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"bome-backend/internal/services"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Database connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "25060")
	dbName := getEnv("DB_NAME", "bomedb")
	dbUser := getEnv("DB_USER", "bomedb")
	dbPassword := getEnv("DB_PASSWORD", "")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL database")

	// Initialize all tables
	if err := initializeSchema(db); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Create super admin user
	if err := createSuperAdmin(db); err != nil {
		log.Fatalf("Failed to create super admin: %v", err)
	}

	log.Println("🎉 Database initialization completed successfully!")
}

func initializeSchema(db *sql.DB) error {
	log.Println("Initializing database schema...")

	schemas := []string{
		// 1. Core Users Table
		`CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role VARCHAR(50) DEFAULT 'user',
    role_id VARCHAR(100),
    email_verified BOOLEAN DEFAULT FALSE,
    stripe_customer_id VARCHAR(255),
    reset_token VARCHAR(255),
    reset_token_expiry TIMESTAMP,
    verification_token VARCHAR(255),
    password_changed BOOLEAN DEFAULT FALSE,
    bio TEXT,
    location VARCHAR(255),
    website VARCHAR(500),
    phone VARCHAR(50),
    avatar_url VARCHAR(500),
    preferences JSONB DEFAULT '{}',
    last_login TIMESTAMP,
    last_logout TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    sub_id VARCHAR(255),
    has_subbed BOOLEAN DEFAULT FALSE,
    max_sessions INTEGER DEFAULT 5,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`,

		// 2. Roles System
		`CREATE TABLE IF NOT EXISTS roles (
            id SERIAL PRIMARY KEY,
            role_id VARCHAR(100) UNIQUE NOT NULL,
            name VARCHAR(255) NOT NULL,
            slug VARCHAR(100) UNIQUE NOT NULL,
            description TEXT,
            category VARCHAR(50) NOT NULL,
            level INTEGER NOT NULL DEFAULT 1,
            permissions JSONB DEFAULT '[]',
            is_system_role BOOLEAN DEFAULT FALSE,
            color VARCHAR(7) DEFAULT '#6b7280',
            icon VARCHAR(50),
            subsystem_access JSONB DEFAULT '[]',
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );`,

		`CREATE TABLE IF NOT EXISTS user_roles (
            id SERIAL PRIMARY KEY,
            user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
            role_id VARCHAR(100) REFERENCES roles(role_id) ON DELETE CASCADE,
            assigned_by INTEGER REFERENCES users(id),
            assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            expires_at TIMESTAMP,
            UNIQUE(user_id, role_id)
        );`,

		// 3. Departments
		`CREATE TABLE IF NOT EXISTS departments (
            id SERIAL PRIMARY KEY,
            dept_id VARCHAR(100) UNIQUE NOT NULL,
            name VARCHAR(255) NOT NULL,
            slug VARCHAR(100) UNIQUE NOT NULL,
            description TEXT,
            color VARCHAR(7) DEFAULT '#6b7280',
            icon VARCHAR(50),
            is_active BOOLEAN DEFAULT TRUE,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );`,

		// 4. YouTube Videos
		`CREATE TABLE IF NOT EXISTS youtube_videos (
            id VARCHAR(255) PRIMARY KEY,
            title TEXT NOT NULL,
            description TEXT,
            published_at TIMESTAMP NOT NULL,
            updated_at TIMESTAMP NOT NULL,
            thumbnail_url TEXT,
            video_url TEXT NOT NULL,
            embed_url TEXT NOT NULL,
            duration VARCHAR(50),
            view_count BIGINT DEFAULT 0,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );`,

		// 5. Analytics Tables
		`CREATE TABLE IF NOT EXISTS analytics_events (
            id SERIAL PRIMARY KEY,
            event_type VARCHAR(100) NOT NULL,
            user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
            session_id VARCHAR(255),
            subsite VARCHAR(50) DEFAULT 'streaming',
            event_data TEXT,
            ip_address INET,
            user_agent TEXT,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        );`,

		// 6. Subscription Plans
		`CREATE TABLE IF NOT EXISTS subscription_plans (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            description TEXT,
            price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
            currency VARCHAR(3) DEFAULT 'USD',
            interval VARCHAR(20) NOT NULL CHECK (interval IN ('monthly', 'annual', 'weekly', 'daily')),
            interval_count INTEGER DEFAULT 1 CHECK (interval_count > 0),
            stripe_price_id VARCHAR(255) UNIQUE,
            features JSONB,
            is_active BOOLEAN DEFAULT true,
            is_promoted BOOLEAN DEFAULT false,
            promotion_end_date TIMESTAMP WITH TIME ZONE,
            sort_order INTEGER DEFAULT 0,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            deleted_at TIMESTAMP WITH TIME ZONE NULL
        );`,

		// 7. User Sessions
		`CREATE TABLE IF NOT EXISTS user_sessions (
            id SERIAL PRIMARY KEY,
            user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
            session_token VARCHAR(255) UNIQUE NOT NULL,
            ip_address INET,
            user_agent TEXT,
            is_active BOOLEAN DEFAULT TRUE,
            last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            expires_at TIMESTAMP NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );`,

		// 8. Audit Logs
		`CREATE TABLE IF NOT EXISTS audit_logs (
            id SERIAL PRIMARY KEY,
            user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
            action VARCHAR(100) NOT NULL,
            resource_type VARCHAR(100),
            resource_id VARCHAR(100),
            details JSONB,
            ip_address INET,
            user_agent TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );`,

		// 9. Master Video List and Tagging System
		`CREATE TABLE IF NOT EXISTS master_video_list (
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
            last_bunny_sync TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            last_master_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            sync_status VARCHAR(50) DEFAULT 'synced',
            sync_notes TEXT,
            metadata_version INTEGER DEFAULT 1,
            created_by INTEGER REFERENCES users(id),
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );`,

		`CREATE TABLE IF NOT EXISTS video_tags (
            id SERIAL PRIMARY KEY,
            word VARCHAR(100) NOT NULL UNIQUE,
            frequency INTEGER DEFAULT 1,
            category_id INTEGER,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );`,

		`CREATE TABLE IF NOT EXISTS tag_categories (
            id SERIAL PRIMARY KEY,
            name VARCHAR(50) NOT NULL UNIQUE,
            description TEXT,
            color VARCHAR(7) DEFAULT '#6b7280',
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );`,

		`CREATE TABLE IF NOT EXISTS video_sync_conflicts (
            id SERIAL PRIMARY KEY,
            master_video_id INTEGER REFERENCES master_video_list(id) ON DELETE CASCADE,
            bunny_video_id VARCHAR(255) NOT NULL,
            conflict_type VARCHAR(50) NOT NULL,
            field_name VARCHAR(100),
            master_value TEXT,
            bunny_value TEXT,
            proposed_action VARCHAR(50) NOT NULL,
            admin_notes TEXT,
            resolved BOOLEAN DEFAULT false,
            resolved_by INTEGER REFERENCES users(id),
            resolved_at TIMESTAMP,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );`,
	}

	// Execute each schema
	for i, schema := range schemas {
		log.Printf("Creating schema %d/%d...", i+1, len(schemas))
		if _, err := db.Exec(schema); err != nil {
			return fmt.Errorf("failed to create schema %d: %v", i+1, err)
		}
	}

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);",
		"CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);",
		"CREATE INDEX IF NOT EXISTS idx_analytics_events_created_at ON analytics_events(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_subscription_plans_active ON subscription_plans(is_active);",
		"CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(session_token);",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);",
		
		// Master video list indexes
		"CREATE INDEX IF NOT EXISTS idx_master_video_bunny_id ON master_video_list(bunny_video_id);",
		"CREATE INDEX IF NOT EXISTS idx_master_video_status ON master_video_list(status);",
		"CREATE INDEX IF NOT EXISTS idx_master_video_category ON master_video_list(category);",
		"CREATE INDEX IF NOT EXISTS idx_master_video_sync_status ON master_video_list(sync_status);",
		"CREATE INDEX IF NOT EXISTS idx_master_video_created_at ON master_video_list(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_master_video_views ON master_video_list(views DESC);",
		"CREATE INDEX IF NOT EXISTS idx_master_video_collection ON master_video_list(collection_id);",
		"CREATE INDEX IF NOT EXISTS idx_master_video_tagged ON master_video_list(tagged);",
		
		// Video tags indexes
		"CREATE INDEX IF NOT EXISTS idx_video_tags_word ON video_tags(word);",
		"CREATE INDEX IF NOT EXISTS idx_video_tags_frequency ON video_tags(frequency DESC);",
		"CREATE INDEX IF NOT EXISTS idx_video_tags_category ON video_tags(category_id);",
		
		// Tag categories indexes
		"CREATE INDEX IF NOT EXISTS idx_tag_categories_name ON tag_categories(name);",
	}

	for _, index := range indexes {
		if _, err := db.Exec(index); err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
		}
	}

	// Insert essential roles
	if err := insertEssentialRoles(db); err != nil {
		return fmt.Errorf("failed to insert essential roles: %v", err)
	}

	// Insert default tag categories
	if err := insertDefaultTagCategories(db); err != nil {
		return fmt.Errorf("failed to insert default tag categories: %v", err)
	}

	log.Println("✅ Schema initialization completed")
	return nil
}

func insertEssentialRoles(db *sql.DB) error {
	log.Println("Inserting essential roles...")

	roles := []string{
		`INSERT INTO roles (role_id, name, slug, description, category, level, is_system_role, color, icon, subsystem_access) 
         VALUES ('super_admin', 'Super Administrator', 'super-administrator', 'Full system access', 'system', 10, true, '#dc2626', 'crown', '["hub", "articles", "youtube", "streaming", "events"]')
         ON CONFLICT (role_id) DO NOTHING;`,

		`INSERT INTO roles (role_id, name, slug, description, category, level, is_system_role, color, icon, subsystem_access) 
         VALUES ('admin', 'Administrator', 'administrator', 'System administration', 'system', 9, true, '#7c3aed', 'server', '["hub", "articles", "youtube", "streaming"]')
         ON CONFLICT (role_id) DO NOTHING;`,

		`INSERT INTO roles (role_id, name, slug, description, category, level, is_system_role, color, icon, subsystem_access) 
         VALUES ('user', 'User', 'user', 'Standard user', 'general', 1, true, '#6b7280', 'user', '["streaming"]')
         ON CONFLICT (role_id) DO NOTHING;`,
	}

	for _, role := range roles {
		if _, err := db.Exec(role); err != nil {
			return fmt.Errorf("failed to insert role: %v", err)
		}
	}

	return nil
}

func insertDefaultTagCategories(db *sql.DB) error {
	log.Println("Inserting default tag categories...")

	categories := []string{
		`INSERT INTO tag_categories (name, description, color) VALUES
			('Archaeology', 'Archaeological terms and concepts', '#8b5cf6')
			ON CONFLICT (name) DO NOTHING;`,

		`INSERT INTO tag_categories (name, description, color) VALUES
			('Geography', 'Geographic locations and features', '#06b6d4')
			ON CONFLICT (name) DO NOTHING;`,

		`INSERT INTO tag_categories (name, description, color) VALUES
			('DNA Research', 'Genetic and DNA-related terms', '#10b981')
			ON CONFLICT (name) DO NOTHING;`,

		`INSERT INTO tag_categories (name, description, color) VALUES
			('Linguistics', 'Language and linguistic terms', '#f59e0b')
			ON CONFLICT (name) DO NOTHING;`,

		`INSERT INTO tag_categories (name, description, color) VALUES
			('Historical Evidence', 'Historical documentation and evidence', '#ef4444')
			ON CONFLICT (name) DO NOTHING;`,

		`INSERT INTO tag_categories (name, description, color) VALUES
			('Cultural Studies', 'Cultural and anthropological terms', '#ec4899')
			ON CONFLICT (name) DO NOTHING;`,

		`INSERT INTO tag_categories (name, description, color) VALUES
			('Religious Studies', 'Religious and theological terms', '#6366f1')
			ON CONFLICT (name) DO NOTHING;`,

		`INSERT INTO tag_categories (name, description, color) VALUES
			('Documentary', 'Documentary and media terms', '#84cc16')
			ON CONFLICT (name) DO NOTHING;`,

		`INSERT INTO tag_categories (name, description, color) VALUES
			('Lecture', 'Educational and lecture terms', '#f97316')
			ON CONFLICT (name) DO NOTHING;`,

		`INSERT INTO tag_categories (name, description, color) VALUES
			('Interview', 'Interview and discussion terms', '#06b6d4')
			ON CONFLICT (name) DO NOTHING;`,

		`INSERT INTO tag_categories (name, description, color) VALUES
			('Presentation', 'Presentation and presentation terms', '#8b5cf6')
			ON CONFLICT (name) DO NOTHING;`,

		`INSERT INTO tag_categories (name, description, color) VALUES
			('Virtual Tour', 'Tour and exploration terms', '#10b981')
			ON CONFLICT (name) DO NOTHING;`,
	}

	for _, category := range categories {
		if _, err := db.Exec(category); err != nil {
			return fmt.Errorf("failed to insert tag category: %v", err)
		}
	}

	log.Println("✅ Default tag categories inserted successfully")
	return nil
}

func createSuperAdmin(db *sql.DB) error {
	log.Println("Creating super admin user...")

	adminEmail := "admin@bome.test"
	adminPassword := "Admin123!"

	// Check if admin exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", adminEmail).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing admin: %v", err)
	}

	if count > 0 {
		log.Printf("Super admin already exists: %s", adminEmail)
		return nil
	}

	// Hash password
	passwordHash, err := services.HashPassword(adminPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// Create admin user
	var userID int
	err = db.QueryRow(`
        INSERT INTO users (email, password_hash, first_name, last_name, role, email_verified, password_changed, created_at, updated_at) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW()) RETURNING id`,
		adminEmail, passwordHash, "Super", "Admin", "super_admin", true, false,
	).Scan(&userID)

	if err != nil {
		return fmt.Errorf("failed to create admin user: %v", err)
	}

	// Assign super_admin role
	_, err = db.Exec(`
        INSERT INTO user_roles (user_id, role_id, assigned_at) 
        VALUES ($1, $2, NOW())`,
		userID, "super_admin",
	)

	if err != nil {
		return fmt.Errorf("failed to assign admin role: %v", err)
	}

	log.Printf("✅ Super admin created successfully!")
	log.Printf("📧 Email: %s", adminEmail)
	log.Printf("🔑 Password: %s", adminPassword)
	log.Printf("🔒 Password change required on first login")

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
