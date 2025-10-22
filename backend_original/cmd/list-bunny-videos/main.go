package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type BunnyVideo struct {
	VideoID          string `json:"guid"`
	Title            string `json:"title"`
	Status           int    `json:"status"`
	Length           int    `json:"length"`
	Views            int    `json:"views"`
	ThumbnailURL     string `json:"thumbnailFileName"`
	DateUploaded     string `json:"dateUploaded"`
	StorageSize      int64  `json:"storageSize"`
	EncodingProgress int    `json:"encodingProgress"`
}

type BunnyVideosResponse struct {
	Items        []BunnyVideo `json:"items"`
	TotalItems   int          `json:"totalItems"`
	Page         int          `json:"page"`
	ItemsPerPage int          `json:"itemsPerPage"`
}

func main() {
	// Get environment variables
	libraryID := os.Getenv("BUNNY_STREAM_LIBRARY_ID")
	apiKey := os.Getenv("BUNNY_STREAM_API_KEY")

	if libraryID == "" || apiKey == "" {
		log.Fatal("BUNNY_STREAM_LIBRARY_ID and BUNNY_STREAM_API_KEY environment variables are required")
	}

	// Create HTTP client
	client := &http.Client{Timeout: 10 * time.Second}

	// Make request to Bunny Stream API
	url := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos", libraryID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Failed to create request:", err)
	}

	req.Header.Set("AccessKey", apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Failed to execute request:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("API request failed with status %d", resp.StatusCode)
	}

	// Parse response
	var videosResp BunnyVideosResponse
	if err := json.NewDecoder(resp.Body).Decode(&videosResp); err != nil {
		log.Fatal("Failed to decode response:", err)
	}

	// Display results
	fmt.Printf("Found %d videos in Bunny library %s:\n\n", videosResp.TotalItems, libraryID)

	for i, video := range videosResp.Items {
		fmt.Printf("Video %d:\n", i+1)
		fmt.Printf("  ID: %s\n", video.VideoID)
		fmt.Printf("  Title: %s\n", video.Title)
		fmt.Printf("  Status: %d (0=Queued, 1=Processing, 2=Encoding, 3=Finished, 4=Error)\n", video.Status)
		fmt.Printf("  Length: %d seconds\n", video.Length)
		fmt.Printf("  Views: %d\n", video.Views)
		fmt.Printf("  Storage Size: %d bytes\n", video.StorageSize)
		fmt.Printf("  Encoding Progress: %d%%\n", video.EncodingProgress)
		fmt.Printf("  Date Uploaded: %s\n", video.DateUploaded)
		fmt.Printf("  Thumbnail: %s\n", video.ThumbnailURL)
		fmt.Printf("\n")
	}

	if len(videosResp.Items) == 0 {
		fmt.Println("No videos found in your Bunny library.")
		fmt.Println("You can upload videos through the Bunny.net dashboard or use the video upload API.")
	}
}
