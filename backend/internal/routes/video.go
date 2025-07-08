package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bome-backend/internal/database"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// Add at the top with other imports:
type BunnyVideoDEF struct {
	GUID                 string   `json:"guid"`
	Title                string   `json:"title"`
	Length               int      `json:"length"`
	Width                int      `json:"width"`
	Height               int      `json:"height"`
	Status               int      `json:"status"`
	Views                int      `json:"views"`
	StorageSize          int64    `json:"storageSize"`
	Framerate            float64  `json:"framerate"`
	IsPublic             bool     `json:"isPublic"`
	EncodeProgress       int      `json:"encodeProgress"`
	AvailableResolutions []string `json:"resolutions"`
	ThumbnailCount       int      `json:"thumbnailCount"`
	HasMP4Fallback       bool     `json:"hasMP4Fallback"`
	CollectionId         string   `json:"collectionId"`
	AverageWatchTime     int      `json:"averageWatchTime"`
	TotalWatchTime       int64    `json:"totalWatchTime"`
	Category             string   `json:"category"`
	DateUploaded         string   `json:"dateUploaded"` // Changed to string to match BunnyVideo struct
}

// GetVideosFromBunnyHandler fetches videos directly from Bunny.net library
func GetVideosFromBunnyHandler(db *database.DB, bunnyService *services.BunnyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		sync := c.DefaultQuery("sync", "false") == "true"

		// Validate and limit parameters
		if limit > 100 {
			limit = 100
		}
		if limit < 1 {
			limit = 20
		}

		// Fetch videos directly from Bunny.net
		videos, err := fetchBunnyVideos(bunnyService.GetStreamLibrary(), bunnyService.GetStreamAPIKey())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch videos from Bunny.net",
				"details": err.Error(),
			})
			return
		}

		// Apply pagination
		start := offset
		end := start + limit
		if start >= len(videos) {
			start = len(videos)
		}
		if end > len(videos) {
			end = len(videos)
		}

		paginatedVideos := videos[start:end]

		// Transform Bunny.net videos to API response format
		var responseVideos []gin.H
		var totalDuration int64
		var totalSize int64

		for _, bunnyVideo := range paginatedVideos {
			// Get streaming URL from bunny.net
			streamURL := bunnyService.GetStreamURL(bunnyVideo.GUID)
			thumbnailURL := bunnyService.GetThumbnailURL(bunnyVideo.GUID)

			// Enhanced response with Bunny.net data
			responseVideo := gin.H{
				"id":           bunnyVideo.GUID,
				"title":        bunnyVideo.Title,
				"description":  fmt.Sprintf("Video from Bunny.net library. Duration: %d seconds, Resolution: %dx%d", bunnyVideo.Length, bunnyVideo.Width, bunnyVideo.Height),
				"thumbnailUrl": thumbnailURL,
				"videoUrl":     streamURL,
				"duration":     bunnyVideo.Length,
				"viewCount":    bunnyVideo.Views,
				"likeCount":    0, // Bunny.net doesn't provide like counts
				"category":     bunnyVideo.Category,
				"tags":         extractTagsFromBunnyVideo(bunnyVideo),
				"status":       mapBunnyStatus(bunnyVideo.Status),
				"createdAt":    bunnyVideo.DateUploaded,
				"updatedAt":    bunnyVideo.DateUploaded,
				"bunny": gin.H{
					"bunny_id":              bunnyVideo.GUID,
					"bunny_status":          bunnyVideo.Status,
					"bunny_duration":        bunnyVideo.Length,
					"bunny_size":            bunnyVideo.StorageSize,
					"width":                 bunnyVideo.Width,
					"height":                bunnyVideo.Height,
					"framerate":             bunnyVideo.Framerate,
					"views":                 bunnyVideo.Views,
					"is_public":             bunnyVideo.IsPublic,
					"encode_progress":       bunnyVideo.EncodeProgress,
					"available_resolutions": bunnyVideo.AvailableResolutions,
					"thumbnail_count":       bunnyVideo.ThumbnailCount,
					"has_mp4_fallback":      bunnyVideo.HasMP4Fallback,
					"collection_id":         bunnyVideo.CollectionId,
					"average_watch_time":    bunnyVideo.AverageWatchTime,
					"total_watch_time":      bunnyVideo.TotalWatchTime,
				},
				"metadata": gin.H{
					"fileSize":   bunnyVideo.StorageSize,
					"resolution": fmt.Sprintf("%dx%d", bunnyVideo.Width, bunnyVideo.Height),
					"framerate":  bunnyVideo.Framerate,
				},
			}
			responseVideos = append(responseVideos, responseVideo)
			totalDuration += int64(bunnyVideo.Length)
			totalSize += bunnyVideo.StorageSize
		}

		// Calculate pagination info
		currentPage := page
		totalPages := (len(videos) + limit - 1) / limit
		hasMore := currentPage < totalPages

		// Sync to database if requested
		if sync {
			go func() {
				for _, bunnyVideo := range paginatedVideos {
					syncVideoToDatabase(db, bunnyService, bunnyVideo)
				}
			}()
		}

		// Enhanced response with bunny.net integration info
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"source":  "bunny_net_library",
			"videos":  responseVideos,
			"pagination": gin.H{
				"current_page": currentPage,
				"per_page":     limit,
				"total":        len(videos),
				"total_pages":  totalPages,
				"has_more":     hasMore,
				"offset":       offset,
			},
			"summary": gin.H{
				"total_videos":   len(responseVideos),
				"total_duration": totalDuration,
				"total_size":     totalSize,
				"average_duration": func() float64 {
					if len(responseVideos) > 0 {
						return float64(totalDuration) / float64(len(responseVideos))
					}
					return 0
				}(),
			},
			"bunny_integration": gin.H{
				"library_id":   bunnyService.GetStreamLibrary(),
				"region":       bunnyService.GetRegion(),
				"cdn_domain":   "iframe.mediadelivery.net",
				"sync_enabled": sync,
			},
			"timestamp": time.Now().Format("2006-01-02T15:04:05Z"),
		})
	}
}

// Helper functions for Bunny.net integration

// extractTagsFromBunnyVideo extracts tags from Bunny.net video metadata
func extractTagsFromBunnyVideo(bunnyVideo BunnyVideoDEF) []string {
	tags := []string{"bunny", "streaming"}

	if bunnyVideo.Title != "" {
		tags = append(tags, strings.ToLower(bunnyVideo.Title))
	}

	if bunnyVideo.Category != "" {
		tags = append(tags, strings.ToLower(bunnyVideo.Category))
	}

	return tags
}

// mapBunnyStatus maps Bunny.net status codes to readable status strings
func mapBunnyStatus(status int) string {
	switch status {
	case 0:
		return "queued"
	case 1:
		return "processing"
	case 2:
		return "encoding"
	case 3:
		return "ready"
	case 4:
		return "error"
	default:
		return "unknown"
	}
}

// Update fetchBunnyVideos to handle the paginated response
func fetchBunnyVideos(libraryID, apiKey string) ([]BunnyVideoDEF, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET",
		fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos", libraryID),
		nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("AccessKey", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check status code first
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	// Decode into paginated response structure
	var response struct {
		Items []BunnyVideoDEF `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Items, nil
}

// Update syncVideoToDatabase to use the correct bunnyVideo parameter for tag extraction
func syncVideoToDatabase(db *database.DB, bunnyService *services.BunnyService, bunnyVideo BunnyVideoDEF) error {
	// Convert to services.BunnyVideo format for database
	uploadTime, _ := time.Parse(time.RFC3339, bunnyVideo.DateUploaded)
	serviceVideo := services.BunnyVideo{
		ID:    bunnyVideo.GUID,
		Title: bunnyVideo.Title,
		Description: fmt.Sprintf("Video from Bunny.net library. Duration: %d seconds, Resolution: %dx%d",
			bunnyVideo.Length, bunnyVideo.Width, bunnyVideo.Height),
		Status:    mapBunnyStatus(bunnyVideo.Status),
		CreatedAt: uploadTime,
		UpdatedAt: uploadTime,
		Duration:  float64(bunnyVideo.Length),
		Size:      bunnyVideo.StorageSize,
		Thumbnail: bunnyService.GetThumbnailURL(bunnyVideo.GUID),
		Preview:   bunnyService.GetStreamURL(bunnyVideo.GUID),
		LibraryID: bunnyService.GetStreamLibrary(),
	}

	_, err := db.CreateVideo(
		serviceVideo.Title,
		serviceVideo.Description,
		serviceVideo.ID,
		serviceVideo.Thumbnail,
		bunnyVideo.Category,
		int(serviceVideo.Duration),
		serviceVideo.Size,
		extractTagsFromBunnyVideo(bunnyVideo), // Fixed: Use original bunnyVideo instead of serviceVideo
		0,                                     // createdBy - system
	)
	return err
}
