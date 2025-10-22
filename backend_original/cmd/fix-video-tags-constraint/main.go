package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

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

	// Fix video_tags table constraints
	if err := fixVideoTagsConstraints(db); err != nil {
		log.Fatalf("Failed to fix video_tags constraints: %v", err)
	}

	log.Println("🎉 Video tags constraints fixed successfully!")
}

func fixVideoTagsConstraints(db *sql.DB) error {
	log.Println("🔧 Fixing video_tags table constraints...")

	// Check if the unique constraint exists
	var constraintExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.table_constraints 
			WHERE table_name = 'video_tags' 
			AND constraint_name = 'video_tags_word_key'
		)`).Scan(&constraintExists)

	if err != nil {
		return fmt.Errorf("failed to check constraint existence: %v", err)
	}

	if constraintExists {
		log.Println("✅ Unique constraint already exists on word column")
		return nil
	}

	log.Println("📝 Adding unique constraint on word column...")

	// Add unique constraint
	_, err = db.Exec(`ALTER TABLE video_tags ADD CONSTRAINT video_tags_word_key UNIQUE (word)`)
	if err != nil {
		return fmt.Errorf("failed to add unique constraint: %v", err)
	}

	log.Println("✅ Unique constraint added successfully")

	// Verify the constraint was added
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.table_constraints 
			WHERE table_name = 'video_tags' 
			AND constraint_name = 'video_tags_word_key'
		)`).Scan(&constraintExists)

	if err != nil {
		return fmt.Errorf("failed to verify constraint: %v", err)
	}

	if constraintExists {
		log.Println("✅ Constraint verification successful")
	} else {
		return fmt.Errorf("constraint was not added properly")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
