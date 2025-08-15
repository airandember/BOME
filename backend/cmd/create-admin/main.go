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
	dbUser := getEnv("DB_USER", "doadmin")
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
