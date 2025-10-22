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

	// Populate default tag categories
	if err := populateTagCategories(db); err != nil {
		log.Fatalf("Failed to populate tag categories: %v", err)
	}

	log.Println("🎉 Tag categories population completed successfully!")
}

func populateTagCategories(db *sql.DB) error {
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

	for i, category := range categories {
		log.Printf("Inserting category %d/%d...", i+1, len(categories))
		if _, err := db.Exec(category); err != nil {
			return fmt.Errorf("failed to insert tag category: %v", err)
		}
	}

	// Verify the insertion
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM tag_categories").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to count tag categories: %v", err)
	}

	log.Printf("✅ Successfully inserted %d tag categories", count)

	// Show what was inserted
	rows, err := db.Query("SELECT name, description, color FROM tag_categories ORDER BY name")
	if err != nil {
		return fmt.Errorf("failed to query tag categories: %v", err)
	}
	defer rows.Close()

	log.Println("📋 Inserted categories:")
	for rows.Next() {
		var name, description, color string
		if err := rows.Scan(&name, &description, &color); err != nil {
			return fmt.Errorf("failed to scan tag category: %v", err)
		}
		log.Printf("  • %s (%s) - %s", name, color, description)
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
