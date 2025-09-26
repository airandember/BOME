package main

import (
	"fmt"
	"log"

	"bome-backend/internal/config"
	"bome-backend/internal/database"
	"bome-backend/internal/services"
)

func main() {
	fmt.Println("🎥 BOME YouTube RSS Sync")
	fmt.Println("========================")
	fmt.Println("Syncing YouTube RSS feed to database...")

	// Load configuration
	cfg := config.New()

	// Connect to database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create YouTube service
	youtubeService := services.NewYouTubeService(db)

	// Check current status
	fmt.Println("\n📊 Current status:")
	status, err := youtubeService.GetSyncStatus()
	if err != nil {
		log.Printf("❌ Failed to get sync status: %v", err)
	} else {
		fmt.Printf("Sync enabled: %v\n", status["sync_enabled"])
		if lastSync, ok := status["last_sync"]; ok {
			fmt.Printf("Last sync: %v\n", lastSync)
		}
		if totalVideos, ok := status["total_videos"]; ok {
			fmt.Printf("Total videos in database: %v\n", totalVideos)
		}
	}

	// Trigger RSS sync
	fmt.Println("\n🔄 Syncing from RSS feed...")
	result, err := youtubeService.SyncFromRSS()
	if err != nil {
		log.Fatalf("❌ RSS sync failed: %v", err)
	}

	// Display results
	fmt.Printf("\n✅ RSS sync completed!\n")
	fmt.Printf("📈 Results:\n")
	fmt.Printf("  - Total fetched: %d\n", result.TotalFetched)
	fmt.Printf("  - New videos: %d\n", result.NewVideos)
	fmt.Printf("  - Updated videos: %d\n", result.UpdatedVideos)
	fmt.Printf("  - Sync time: %s\n", result.SyncTime.Format("2006-01-02 15:04:05"))

	if len(result.Errors) > 0 {
		fmt.Printf("⚠️ Errors encountered:\n")
		for _, errMsg := range result.Errors {
			fmt.Printf("  - %s\n", errMsg)
		}
	}

	// Check final status
	fmt.Println("\n📊 Final status:")
	videos, err := youtubeService.GetLatestVideos(5)
	if err != nil {
		log.Printf("❌ Failed to get latest videos: %v", err)
	} else {
		fmt.Printf("Latest %d videos in database:\n", len(videos.Videos))
		for i, video := range videos.Videos {
			fmt.Printf("  %d. %s (ID: %s)\n", i+1, video.Title, video.ID)
		}
	}

	fmt.Println("\n🎉 YouTube RSS sync completed successfully!")
}
