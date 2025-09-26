package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"bome-backend/internal/config"
	"bome-backend/internal/database"
	"bome-backend/internal/services"
)

// SearchIndexVideo represents a video in the search index
type SearchIndexVideo struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	Tags         []string `json:"tags"`
	Duration     int      `json:"duration"`
	CreatedAt    string   `json:"createdAt"`
	Thumbnail    string   `json:"thumbnail"`
	ThumbnailURL string   `json:"thumbnailUrl"`
	Bunny        struct {
		GUID              string `json:"guid"`
		VideoLibraryID    string `json:"videoLibraryId"`
		ThumbnailFileName string `json:"thumbnailFileName"`
		PreviewImageURL   string `json:"previewImageUrl"`
		Width             int    `json:"width"`
		Height            int    `json:"height"`
		Length            int    `json:"length"`
	} `json:"bunny"`
	Views     int    `json:"views"`
	Status    string `json:"status"`
	VideoURL  string `json:"videoUrl"`
	IframeSrc string `json:"iframeSrc"`
}

// SearchIndex represents the complete search index structure
type SearchIndex struct {
	Version     string             `json:"version"`
	GeneratedAt string             `json:"generatedAt"`
	TotalVideos int                `json:"totalVideos"`
	Videos      []SearchIndexVideo `json:"videos"`
}

func main() {
	log.Println("🚀 Starting search index generation...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Bunny service
	bunnyService := services.NewBunnyService(
		cfg.BunnyStreamAPIKey,
		cfg.BunnyStreamLibraryID,
		cfg.BunnyStreamRegion,
		cfg.BunnyCDNAPIKey,
		cfg.BunnyCDNZoneName,
	)

	// Generate search index
	searchIndex, err := generateSearchIndex(db, bunnyService)
	if err != nil {
		log.Fatalf("❌ Failed to generate search index: %v", err)
	}

	// Write to file
	outputPath := getOutputPath()
	err = writeSearchIndex(searchIndex, outputPath)
	if err != nil {
		log.Fatalf("❌ Failed to write search index: %v", err)
	}

	log.Printf("✅ Search index generated successfully!")
	log.Printf("📊 Total videos: %d", searchIndex.TotalVideos)
	log.Printf("📁 Output: %s", outputPath)
}

func generateSearchIndex(db *database.DB, bunnyService *services.BunnyService) (*SearchIndex, error) {
	log.Println("🔍 Fetching all videos from database...")

	// Get all videos from database
	videos, err := db.GetAllVideos()
	if err != nil {
		return nil, fmt.Errorf("failed to get videos: %w", err)
	}

	log.Printf("📥 Found %d videos in database", len(videos))

	var searchVideos []SearchIndexVideo
	for _, video := range videos {
		// Generate thumbnail URLs
		thumbnailURL := ""
		if video.BunnyVideoID != "" {
			if video.ThumbnailFileName != "" {
				thumbnailURL = bunnyService.GetThumbnailURLWithFilename(video.BunnyVideoID, video.ThumbnailFileName)
			} else {
				thumbnailURL = bunnyService.GetThumbnailURL(video.BunnyVideoID)
			}
		}

		// Fallback thumbnail if none provided
		fallbackThumbnail := ""
		if video.BunnyVideoID != "" {
			fallbackThumbnail = fmt.Sprintf("https://vz-f75053f7-465.b-cdn.net/%s/thumbnail.jpg", video.BunnyVideoID)
		}

		searchVideo := SearchIndexVideo{
			ID:           video.BunnyVideoID,
			Title:        video.Title,
			Description:  video.Description,
			Category:     video.Category,
			Tags:         video.Tags,
			Duration:     video.Duration,
			CreatedAt:    video.CreatedAt.Format(time.RFC3339),
			Thumbnail:    thumbnailURL,
			ThumbnailURL: thumbnailURL,
			Views:        video.ViewCount,
			Status:       video.Status,
			VideoURL:     bunnyService.GetStreamURL(video.BunnyVideoID),
			IframeSrc:    bunnyService.GetIframeURL(video.BunnyVideoID),
		}

		// Set bunny metadata
		searchVideo.Bunny.GUID = video.BunnyVideoID
		searchVideo.Bunny.ThumbnailFileName = video.ThumbnailFileName
		searchVideo.Bunny.PreviewImageURL = fallbackThumbnail
		searchVideo.Bunny.Length = video.Duration

		// Use thumbnail or fallback
		if searchVideo.Thumbnail == "" {
			searchVideo.Thumbnail = fallbackThumbnail
			searchVideo.ThumbnailURL = fallbackThumbnail
		}

		searchVideos = append(searchVideos, searchVideo)
	}

	return &SearchIndex{
		Version:     "1.0",
		GeneratedAt: time.Now().Format(time.RFC3339),
		TotalVideos: len(searchVideos),
		Videos:      searchVideos,
	}, nil
}

func getOutputPath() string {
	// Check for custom output path
	if outputPath := os.Getenv("SEARCH_INDEX_OUTPUT"); outputPath != "" {
		return outputPath
	}

	// Default paths to try
	paths := []string{
		"../frontend/static/search-index.json",
		"../../frontend/static/search-index.json",
		"/app/frontend/static/search-index.json",
		"./search-index.json",
	}

	for _, path := range paths {
		if dir := filepath.Dir(path); dirExists(dir) {
			return path
		}
	}

	// Fallback to current directory
	return "./search-index.json"
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func writeSearchIndex(index *SearchIndex, outputPath string) error {
	log.Printf("📝 Writing search index to: %s", outputPath)

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Log file size
	info, _ := os.Stat(outputPath)
	fileSizeKB := info.Size() / 1024
	log.Printf("📊 File size: %d KB", fileSizeKB)

	return nil
}
