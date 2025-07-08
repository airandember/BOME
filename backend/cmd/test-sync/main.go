package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"bome-backend/internal/config"
	"bome-backend/internal/database"
	"bome-backend/internal/services"
)

// BunnyVideo represents a video from Bunny.net API
type BunnyVideo struct {
	GUID                 string  `json:"guid"`
	Title                string  `json:"title"`
	DateUploaded         string  `json:"dateUploaded"`
	Views                int     `json:"views"`
	IsPublic             bool    `json:"isPublic"`
	Length               int     `json:"length"`
	Status               int     `json:"status"`
	Framerate            float64 `json:"framerate"`
	Rotation             int     `json:"rotation"`
	Width                int     `json:"width"`
	Height               int     `json:"height"`
	AvailableResolutions string  `json:"availableResolutions"`
	ThumbnailCount       int     `json:"thumbnailCount"`
	EncodeProgress       int     `json:"encodeProgress"`
	StorageSize          int64   `json:"storageSize"`
	HasMP4Fallback       bool    `json:"hasMP4Fallback"`
	CollectionId         string  `json:"collectionId"`
	ThumbnailFileName    string  `json:"thumbnailFileName"`
	AverageWatchTime     int     `json:"averageWatchTime"`
	TotalWatchTime       int     `json:"totalWatchTime"`
	Category             string  `json:"category"`
	Chapters             []struct {
		Title string `json:"title"`
		Start int    `json:"start"`
		End   int    `json:"end"`
	} `json:"chapters"`
	Moments []struct {
		Label     string `json:"label"`
		Timestamp int    `json:"timestamp"`
	} `json:"moments"`
	MetaTags []struct {
		Property string `json:"property"`
		Value    string `json:"value"`
	} `json:"metaTags"`
	TranscodingMessages []struct {
		TimeStamp int    `json:"timeStamp"`
		Level     int    `json:"level"`
		IssueCode int    `json:"issueCode"`
		Message   string `json:"message"`
	} `json:"transcodingMessages"`
}

// BunnyVideosResponse represents the API response from Bunny.net
type BunnyVideosResponse struct {
	TotalItems   int          `json:"totalItems"`
	CurrentPage  int          `json:"currentPage"`
	ItemsPerPage int          `json:"itemsPerPage"`
	Items        []BunnyVideo `json:"items"`
}

func main() {
	fmt.Println("🚀 Starting Bunny.net sync test...")

	// Load configuration
	cfg := config.New()

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Bunny service
	bunnyService := services.NewBunnyService()

	fmt.Println("✅ Configuration loaded successfully")
	fmt.Printf("📚 Bunny Stream Library ID: %s\n", cfg.BunnyStreamLibrary)
	fmt.Printf("🔑 API Key configured: %t\n", cfg.BunnyStreamAPIKey != "")

	// Fetch videos from Bunny.net
	fmt.Println("\n📡 Fetching videos from Bunny.net...")
	videos, err := fetchBunnyVideos(cfg.BunnyStreamLibrary, cfg.BunnyStreamAPIKey)
	if err != nil {
		log.Fatalf("Failed to fetch videos: %v", err)
	}

	fmt.Printf("📹 Found %d videos in Bunny.net library\n", len(videos))

	// Display videos found
	for i, video := range videos {
		fmt.Printf("  %d. %s (ID: %s, Status: %d, Length: %d seconds)\n",
			i+1, video.Title, video.GUID, video.Status, video.Length)
	}

	// Sync videos to database
	fmt.Println("\n💾 Syncing videos to database...")
	syncedCount := 0
	skippedCount := 0
	errorCount := 0
	var errors []string

	for _, bunnyVideo := range videos {
		err := syncVideoToDatabase(db, bunnyService, bunnyVideo)
		if err != nil {
			if contains(err.Error(), "already exists") {
				fmt.Printf("⏭️  Skipped: %s (already exists)\n", bunnyVideo.Title)
				skippedCount++
			} else {
				fmt.Printf("❌ Error: %s - %v\n", bunnyVideo.Title, err)
				errorCount++
				errors = append(errors, fmt.Sprintf("%s: %v", bunnyVideo.Title, err))
			}
			continue
		}
		fmt.Printf("✅ Synced: %s\n", bunnyVideo.Title)
		syncedCount++
	}

	// Summary
	fmt.Println("\n📊 Sync Summary:")
	fmt.Printf("  Total videos found: %d\n", len(videos))
	fmt.Printf("  Successfully synced: %d\n", syncedCount)
	fmt.Printf("  Skipped (already exists): %d\n", skippedCount)
	fmt.Printf("  Errors: %d\n", errorCount)

	if len(errors) > 0 {
		fmt.Println("\n❌ Error Details:")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
	}

	// Check final video count
	fmt.Println("\n🔍 Checking final video count in database...")
	finalVideos, err := db.GetVideos(100, 0, "", "")
	if err != nil {
		fmt.Printf("⚠️  Could not get final video count: %v\n", err)
	} else {
		fmt.Printf("📹 Total videos in database: %d\n", len(finalVideos))
	}

	fmt.Println("\n✨ Sync test completed!")
}

// fetchBunnyVideos fetches all videos from Bunny.net library
func fetchBunnyVideos(libraryID, apiKey string) ([]BunnyVideo, error) {
	url := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos", libraryID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("AccessKey", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var response BunnyVideosResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Items, nil
}

// syncVideoToDatabase syncs a Bunny video to the database
func syncVideoToDatabase(db *database.DB, bunnyService *services.BunnyService, bunnyVideo BunnyVideo) error {
	// Check if video already exists
	existingVideo, err := db.GetVideoByBunnyID(bunnyVideo.GUID)
	if err == nil && existingVideo != nil {
		return fmt.Errorf("video %s already exists in database", bunnyVideo.GUID)
	}

	// Generate thumbnail URL
	thumbnailURL := bunnyService.GetThumbnailURL(bunnyVideo.GUID)

	// Determine video status based on Bunny status
	var status string
	switch bunnyVideo.Status {
	case 0:
		status = "queued"
	case 1:
		status = "processing"
	case 2:
		status = "encoding"
	case 3:
		status = "ready"
	case 4:
		status = "error"
	default:
		status = "unknown"
	}

	// Parse category or set default
	category := bunnyVideo.Category
	if category == "" {
		category = "General"
	}

	// Create tags from meta tags
	var tags []string
	for _, metaTag := range bunnyVideo.MetaTags {
		if metaTag.Property == "tag" {
			tags = append(tags, metaTag.Value)
		}
	}

	// If no tags from meta, create some based on the video properties
	if len(tags) == 0 {
		tags = []string{"bunny", "streaming"}
		if bunnyVideo.HasMP4Fallback {
			tags = append(tags, "mp4")
		}
		if bunnyVideo.IsPublic {
			tags = append(tags, "public")
		}
	}

	// Create video description
	description := fmt.Sprintf("Video from Bunny.net library. Duration: %d seconds, Resolution: %dx%d",
		bunnyVideo.Length, bunnyVideo.Width, bunnyVideo.Height)

	if len(bunnyVideo.TranscodingMessages) > 0 {
		description += "\n\nTranscoding notes:"
		for _, msg := range bunnyVideo.TranscodingMessages {
			description += fmt.Sprintf("\n- %s", msg.Message)
		}
	}

	// Create video in database
	video, err := db.CreateVideo(
		bunnyVideo.Title,
		description,
		bunnyVideo.GUID,
		thumbnailURL,
		category,
		bunnyVideo.Length,
		bunnyVideo.StorageSize,
		tags,
		1, // Default to admin user ID
	)
	if err != nil {
		return fmt.Errorf("failed to create video in database: %w", err)
	}

	// Update video status
	err = db.UpdateVideoStatus(video.ID, status)
	if err != nil {
		return fmt.Errorf("failed to update video status: %w", err)
	}

	// Update view count if available
	if bunnyVideo.Views > 0 {
		err = db.UpdateVideoViews(video.ID, bunnyVideo.Views)
		if err != nil {
			return fmt.Errorf("failed to update video views: %w", err)
		}
	}

	return nil
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
