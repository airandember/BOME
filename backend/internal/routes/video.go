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

// GetVideosFromBunnyHandler fetches videos directly from Bunny.net library
func GetVideosFromBunnyHandler(db *database.DB, bunnyService *services.BunnyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		sync := c.DefaultQuery("sync", "false") == "true"

		// Validate and limit parameters
		if limit > 100 {
			limit = 100
		}
		if limit < 1 {
			limit = 20
		}

		// Fetch videos directly from Bunny.net with pagination
		videos, totalItems, err := fetchBunnyVideos(bunnyService.GetStreamLibrary(), bunnyService.GetStreamAPIKey(), page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch videos from Bunny.net",
				"details": err.Error(),
			})
			return
		}

		// Transform Bunny.net videos to API response format
		var responseVideos []gin.H
		var totalDuration int64
		var totalSize int64

		for _, bunnyVideo := range videos {
			// Get streaming URL from bunny.net
			streamURL := bunnyService.GetStreamURL(bunnyVideo.GUID)
			thumbnailURL := bunnyService.GetThumbnailURLWithFilename(bunnyVideo.GUID, bunnyVideo.ThumbnailFileName)
			iframeURL := bunnyService.GetIframeURL(bunnyVideo.GUID)

			// Enhanced response with Bunny.net data
			description := fmt.Sprintf("Video from Bunny.net library. Duration: %d seconds, Resolution: %dx%d",
				bunnyVideo.Length, bunnyVideo.Width, bunnyVideo.Height)
			if bunnyVideo.Description != nil {
				description = *bunnyVideo.Description
			}

			responseVideo := gin.H{
				"id":           bunnyVideo.GUID,
				"title":        bunnyVideo.Title,
				"description":  description,
				"thumbnailUrl": thumbnailURL,
				"videoUrl":     streamURL,
				"iframeSrc":    iframeURL,
				"playbackUrl":  streamURL,
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
					"collection_id":         bunnyVideo.CollectionID,
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
		totalPages := (totalItems + limit - 1) / limit
		hasMore := currentPage < totalPages

		// Sync to database if requested
		if sync {
			go func() {
				for _, bunnyVideo := range videos {
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
				"total":        totalItems,
				"total_pages":  totalPages,
				"has_more":     hasMore,
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

// GetBunnyVideoHandler fetches a single video from Bunny.net
func GetBunnyVideoHandler(db *database.DB, bunnyService *services.BunnyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID := c.Param("id")

		// Get video details from Bunny.net
		bunnyVideo, err := bunnyService.GetVideo(videoID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error":    "Video not found",
				"bunny_id": videoID,
				"details":  err.Error(),
				"code":     "VIDEO_NOT_FOUND",
			})
			return
		}

		// Get streaming URL and thumbnail URL
		streamURL := bunnyService.GetStreamURL(videoID)
		thumbnailURL := bunnyService.GetThumbnailURLWithFilename(videoID, bunnyVideo.ThumbnailFileName)

		// Create response
		response := gin.H{
			"id":           bunnyVideo.GUID,
			"title":        bunnyVideo.Title,
			"description":  bunnyVideo.Description,
			"thumbnailUrl": thumbnailURL,
			"playbackUrl":  streamURL, // Add this
			"duration":     bunnyVideo.Length,
			"status":       mapBunnyStatus(bunnyVideo.Status),
			"created_at":   bunnyVideo.DateUploaded,
			"updated_at":   bunnyVideo.DateUploaded,
			"bunny_id":     bunnyVideo.GUID,
			// ... other fields ...
		}

		c.JSON(http.StatusOK, response)
	}
}

// Helper functions for Bunny.net integration

// extractTagsFromBunnyVideo extracts tags from Bunny.net video metadata
func extractTagsFromBunnyVideo(bunnyVideo services.BunnyVideo) []string {
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
		return "created"
	case 1:
		return "uploaded"
	case 2:
		return "processing"
	case 3:
		return "transcoding"
	case 4:
		return "ready" // Finished = Ready for playback
	case 5:
		return "error"
	case 6:
		return "upload_failed"
	case 7:
		return "jit_segmenting"
	case 8:
		return "jit_playlists_created"
	default:
		return "unknown"
	}
}

// Update fetchBunnyVideos to support pagination and date sorting
func fetchBunnyVideos(libraryID, apiKey string, page int, itemsPerPage int) ([]services.BunnyVideo, int, error) {
	// Default values
	if page < 1 {
		page = 1
	}
	if itemsPerPage < 1 || itemsPerPage > 100 {
		itemsPerPage = 100
	}

	url := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos?page=%d&itemsPerPage=%d&orderBy=date",
		libraryID, page, itemsPerPage)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("AccessKey", apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response struct {
		Items        []services.BunnyVideo `json:"items"`
		TotalItems   int                   `json:"totalItems"`
		Page         int                   `json:"currentPage"`
		ItemsPerPage int                   `json:"itemsPerPage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Items, response.TotalItems, nil
}

// Update syncVideoToDatabase
func syncVideoToDatabase(db *database.DB, bunnyService *services.BunnyService, bunnyVideo services.BunnyVideo) error {
	description := fmt.Sprintf("Video from Bunny.net library. Duration: %d seconds, Resolution: %dx%d",
		bunnyVideo.Length, bunnyVideo.Width, bunnyVideo.Height)
	if bunnyVideo.Description != nil {
		description = *bunnyVideo.Description
	}

	_, err := db.CreateVideo(
		bunnyVideo.Title,
		description,
		bunnyVideo.GUID,
		bunnyService.GetThumbnailURLWithFilename(bunnyVideo.GUID, bunnyVideo.ThumbnailFileName),
		bunnyVideo.Category,
		bunnyVideo.Length,
		bunnyVideo.StorageSize,
		extractTagsFromBunnyVideo(bunnyVideo),
		0, // createdBy - system
	)
	return err
}
