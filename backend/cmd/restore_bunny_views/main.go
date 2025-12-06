// restore_bunny_views is a one-time script to restore video view counts from Bunny.net
// Run this after migration 071 to fix the views that were incorrectly set to 100
//
// Usage:
//   go run backend/cmd/restore_bunny_views/main.go
//
// Required environment variables:
//   - DATABASE_URL (or individual DB_* vars)
//   - BUNNY_STREAM_LIBRARY_ID
//   - BUNNY_STREAM_API_KEY
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// BunnyVideo represents a video from Bunny.net API
type BunnyVideo struct {
	GUID  string `json:"guid"`
	Title string `json:"title"`
	Views int    `json:"views"`
}

// BunnyResponse represents the API response
type BunnyResponse struct {
	TotalItems   int          `json:"totalItems"`
	CurrentPage  int          `json:"currentPage"`
	ItemsPerPage int          `json:"itemsPerPage"`
	Items        []BunnyVideo `json:"items"`
}

func main() {
	// Load .env file if present
	_ = godotenv.Load()
	_ = godotenv.Load("../../.env") // Try parent directories too
	_ = godotenv.Load("../../../.env")

	// Get required environment variables
	libraryID := os.Getenv("BUNNY_STREAM_LIBRARY_ID")
	apiKey := os.Getenv("BUNNY_STREAM_API_KEY")
	databaseURL := os.Getenv("DATABASE_URL")

	if libraryID == "" || apiKey == "" {
		log.Fatal("❌ Missing BUNNY_STREAM_LIBRARY_ID or BUNNY_STREAM_API_KEY environment variables")
	}

	// Build database connection string if not provided
	if databaseURL == "" {
		dbHost := getEnvOrDefault("DB_HOST", "localhost")
		dbPort := getEnvOrDefault("DB_PORT", "5432")
		dbUser := getEnvOrDefault("DB_USER", "postgres")
		dbPassword := os.Getenv("DB_PASSWORD")
		dbName := getEnvOrDefault("DB_NAME", "bome_db")
		dbSSLMode := getEnvOrDefault("DB_SSLMODE", "disable")

		databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			dbUser, dbPassword, dbHost, dbPort, dbName, dbSSLMode)
	}

	// Connect to database
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("❌ Failed to ping database: %v", err)
	}
	log.Println("✅ Connected to database")

	// Fetch all videos from Bunny.net
	log.Println("📥 Fetching videos from Bunny.net...")
	bunnyVideos, err := fetchAllBunnyVideos(libraryID, apiKey)
	if err != nil {
		log.Fatalf("❌ Failed to fetch Bunny videos: %v", err)
	}
	log.Printf("✅ Fetched %d videos from Bunny.net", len(bunnyVideos))

	// Create a map for quick lookup
	bunnyViewsMap := make(map[string]int)
	for _, v := range bunnyVideos {
		bunnyViewsMap[v.GUID] = v.Views
	}

	// Get all videos from master_video_list
	rows, err := db.Query(`
		SELECT id, bunny_video_id, title, views 
		FROM master_video_list 
		WHERE bunny_video_id IS NOT NULL AND bunny_video_id != ''
		ORDER BY id
	`)
	if err != nil {
		log.Fatalf("❌ Failed to query master_video_list: %v", err)
	}
	defer rows.Close()

	// Track statistics
	var totalVideos, updated, skipped, notFound int
	var errors []string

	// Prepare update statement
	updateStmt, err := db.Prepare(`UPDATE master_video_list SET views = $1, updated_at = NOW() WHERE id = $2`)
	if err != nil {
		log.Fatalf("❌ Failed to prepare update statement: %v", err)
	}
	defer updateStmt.Close()

	log.Println("🔄 Starting view count restoration...")
	startTime := time.Now()

	for rows.Next() {
		var id int
		var bunnyVideoID, title string
		var currentViews int

		if err := rows.Scan(&id, &bunnyVideoID, &title, &currentViews); err != nil {
			errors = append(errors, fmt.Sprintf("Failed to scan row: %v", err))
			continue
		}

		totalVideos++

		// Look up the Bunny.net view count
		bunnyViews, found := bunnyViewsMap[bunnyVideoID]
		if !found {
			notFound++
			log.Printf("⚠️  Video %d (%s) not found in Bunny.net", id, bunnyVideoID)
			continue
		}

		// Only update if the Bunny.net views are different (and not 100 which was our placeholder)
		if currentViews != bunnyViews {
			_, err := updateStmt.Exec(bunnyViews, id)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to update video %d: %v", id, err))
				continue
			}
			updated++
			log.Printf("✅ Updated video %d: %d → %d views (%s)", id, currentViews, bunnyViews, truncateTitle(title, 40))
		} else {
			skipped++
		}

		// Progress update every 100 videos
		if totalVideos%100 == 0 {
			log.Printf("📊 Progress: %d videos processed, %d updated, %d skipped", totalVideos, updated, skipped)
		}
	}

	duration := time.Since(startTime)

	// Print summary
	log.Println("")
	log.Println("═══════════════════════════════════════════════════")
	log.Println("                  RESTORATION COMPLETE              ")
	log.Println("═══════════════════════════════════════════════════")
	log.Printf("📊 Total videos processed: %d", totalVideos)
	log.Printf("✅ Updated:                %d", updated)
	log.Printf("⏭️  Skipped (no change):    %d", skipped)
	log.Printf("⚠️  Not found in Bunny:     %d", notFound)
	log.Printf("❌ Errors:                 %d", len(errors))
	log.Printf("⏱️  Duration:               %v", duration)
	log.Println("═══════════════════════════════════════════════════")

	if len(errors) > 0 {
		log.Println("")
		log.Println("Errors encountered:")
		for _, e := range errors {
			log.Printf("  - %s", e)
		}
	}

	// Verify the restoration
	log.Println("")
	log.Println("🔍 Verification - Top 10 videos by views:")
	verifyRows, err := db.Query(`
		SELECT id, title, views 
		FROM master_video_list 
		WHERE status = 'ready'
		ORDER BY views DESC 
		LIMIT 10
	`)
	if err == nil {
		defer verifyRows.Close()
		for verifyRows.Next() {
			var id, views int
			var title string
			if verifyRows.Scan(&id, &title, &views) == nil {
				log.Printf("  #%d: %s (%d views)", id, truncateTitle(title, 50), views)
			}
		}
	}
}

func fetchAllBunnyVideos(libraryID, apiKey string) ([]BunnyVideo, error) {
	var allVideos []BunnyVideo
	page := 1
	perPage := 100
	client := &http.Client{Timeout: 30 * time.Second}

	for {
		url := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos?page=%d&itemsPerPage=%d&orderBy=date",
			libraryID, page, perPage)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("AccessKey", apiKey)
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Bunny API returned status %d", resp.StatusCode)
		}

		var response BunnyResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, err
		}

		allVideos = append(allVideos, response.Items...)

		log.Printf("📄 Fetched page %d: %d videos (total so far: %d/%d)",
			page, len(response.Items), len(allVideos), response.TotalItems)

		if len(allVideos) >= response.TotalItems || len(response.Items) == 0 {
			break
		}

		page++
		time.Sleep(100 * time.Millisecond) // Rate limiting
	}

	return allVideos, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func truncateTitle(title string, maxLen int) string {
	if len(title) <= maxLen {
		return title
	}
	return title[:maxLen-3] + "..."
}

