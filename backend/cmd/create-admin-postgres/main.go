package main

import (
	"database/sql"
	"log"
	"os"

	"bome-backend/internal/database"
	"bome-backend/internal/services"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Database connection parameters
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbName := getEnv("DB_NAME", "bome_streaming")
	dbUser := getEnv("DB_USER", "bome_user")
	dbPassword := getEnv("DB_PASSWORD", "")

	// Connect to PostgreSQL database
	connStr := "host=" + dbHost + " port=" + dbPort + " user=" + dbUser + " password=" + dbPassword + " dbname=" + dbName + " sslmode=disable"

	postgresDB, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer postgresDB.Close()

	// Test connection
	if err := postgresDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL database successfully")

	// Create database wrapper
	db := &database.DB{DB: postgresDB}

	// Test admin credentials
	adminEmail := "admin@bome.test"
	adminPassword := "Admin123!"
	adminFirstName := "Test"
	adminLastName := "Administrator"

	// Check if admin already exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", adminEmail).Scan(&count)
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
	var id int
	err = db.QueryRow(
		`INSERT INTO users (email, password_hash, first_name, last_name, role, email_verified, created_at, updated_at) 
		 VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) RETURNING id`,
		adminEmail, passwordHash, adminFirstName, adminLastName, "super_admin", true,
	).Scan(&id)
	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	log.Printf("✅ Admin user created successfully with ID: %d", id)
	log.Printf("📧 Email: %s", adminEmail)
	log.Printf("🔑 Password: %s", adminPassword)
	log.Printf("👤 Role: super_admin")
	log.Printf("✅ Email verified: true")
	log.Printf("")
	log.Printf("🌐 You can now log in to the admin dashboard")
	log.Printf("⚠️  Remember to change the password in production!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
