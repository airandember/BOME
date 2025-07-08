package main

import (
	"database/sql"
	"log"
	"os"

	"bome-backend/internal/database"
	"bome-backend/internal/services"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// For development, use SQLite instead of PostgreSQL
	log.Println("Using SQLite for development...")

	// Create SQLite database file
	dbPath := "./bome_dev.db"

	// Open SQLite database
	sqliteDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer sqliteDB.Close()

	// Create a simple database wrapper for SQLite
	db := &database.DB{DB: sqliteDB}

	// Create users table if it doesn't exist
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		first_name TEXT,
		last_name TEXT,
		role TEXT DEFAULT 'user',
		email_verified INTEGER DEFAULT 0,
		stripe_customer_id TEXT,
		reset_token TEXT,
		reset_token_expiry TEXT,
		verification_token TEXT,
		created_at TEXT DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := db.Exec(createUsersTable); err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}

	// Test admin credentials
	adminEmail := "admin@bome.test"
	adminPassword := "Admin123!"
	adminFirstName := "Test"
	adminLastName := "Administrator"

	// Check if admin already exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", adminEmail).Scan(&count)
	if err != nil {
		log.Fatalf("Failed to check if user exists: %v", err)
	}

	if count > 0 {
		log.Printf("Admin user %s already exists", adminEmail)
		os.Exit(0)
	}

	// Hash password
	passwordHash, err := services.HashPassword(adminPassword)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create admin user
	result, err := db.Exec(
		`INSERT INTO users (email, password_hash, first_name, last_name, role, email_verified, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		adminEmail, passwordHash, adminFirstName, adminLastName, "admin", 1,
	)
	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		log.Fatalf("Failed to get user ID: %v", err)
	}

	log.Printf("✅ Test admin account created successfully!")
	log.Printf("📧 Email: %s", adminEmail)
	log.Printf("🔑 Password: %s", adminPassword)
	log.Printf("👤 Name: %s %s", adminFirstName, adminLastName)
	log.Printf("🆔 User ID: %d", userID)
	log.Printf("🔐 Role: admin")
	log.Printf("✅ Email Verified: true")
	log.Printf("🗄️  Database: %s", dbPath)
	log.Printf("")
	log.Printf("You can now log in to the admin dashboard at:")
	log.Printf("http://localhost:5174/admin")
	log.Printf("")
	log.Printf("⚠️  IMPORTANT: Change these credentials in production!")
	log.Printf("📁 SQLite database file: %s", dbPath)
}
