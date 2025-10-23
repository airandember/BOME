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

// GetVideoEngagementAnalyticsHandler handles retrieving video engagement analytics
func GetVideoEngagementAnalyticsHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		period := c.DefaultQuery("period", "30d")
		videoID := c.Param("id")
		
		// Track analytics access
		trackVideoAnalyticsAccess(db, c, videoID, "engagement_analytics", map[string]interface{}{
			"period": period,
			"timestamp": time.Now().Unix(),
		})
		
		analytics := map[string]interface{}{
			"video_id": videoID,
			"period": period,
			"engagement_metrics": map[string]interface{}{
				"total_views": getVideoTotalViews(db, videoID),
				"unique_viewers": getVideoUniqueViewers(db, videoID, period),
				"avg_watch_time": getVideoAverageWatchTime(db, videoID, period),
				"completion_rate": getVideoCompletionRate(db, videoID, period),
				"engagement_score": getVideoEngagementScore(db, videoID, period),
			},
			"content_performance": map[string]interface{}{
				"views_by_hour": getVideoViewsByHour(db, videoID, period),
				"views_by_device": getVideoViewsByDevice(db, videoID, period),
				"views_by_location": getVideoViewsByLocation(db, videoID, period),
				"traffic_sources": getVideoTrafficSources(db, videoID, period),
			},
			"user_behavior": map[string]interface{}{
				"drop_off_points": getVideoDropOffPoints(db, videoID, period),
				"rewatch_rate": getVideoRewatchRate(db, videoID, period),
				"share_rate": getVideoShareRate(db, videoID, period),
				"like_rate": getVideoLikeRate(db, videoID, period),
			},
		}
		
		c.JSON(http.StatusOK, gin.H{"data": analytics})
	}
}

// GetVideosFromBunnyHandler fetches videos directly from Bunny.net library
func GetVideosFromBunnyHandler(db *database.DB, bunnyService *services.BunnyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		sync := c.DefaultQuery("sync", "false") == "true"
		search := c.DefaultQuery("search", "")

		// Validate and limit parameters
		if limit > 100 {
			limit = 100
		}
		if limit < 1 {
			limit = 20
		}

		var videos []services.BunnyVideo
		var totalItems int
		var err error

		if search != "" {
			// For search requests, fetch more videos to filter from
			videos, totalItems, err = fetchBunnyVideosWithSearch(bunnyService.GetStreamLibrary(), bunnyService.GetStreamAPIKey(), search, limit)
		} else {
			// Regular pagination
			videos, totalItems, err = fetchBunnyVideos(bunnyService.GetStreamLibrary(), bunnyService.GetStreamAPIKey(), page, limit, "")
		}

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

// fetchBunnyVideosWithSearch fetches videos from Bunny.net and filters them server-side
func fetchBunnyVideosWithSearch(libraryID, apiKey string, search string, limit int) ([]services.BunnyVideo, int, error) {
	if search == "" {
		return fetchBunnyVideos(libraryID, apiKey, 1, limit, "")
	}

	fmt.Printf("🔍 Server-side search for: %s\n", search)

	// Fetch multiple pages to have more videos to search through
	maxPages := 10 // Fetch up to 10 pages (up to 1000 videos) to search through
	var allVideos []services.BunnyVideo

	for page := 1; page <= maxPages; page++ {
		videos, _, err := fetchBunnyVideos(libraryID, apiKey, page, 100, "") // Fetch 100 per page
		if err != nil {
			return nil, 0, err
		}

		if len(videos) == 0 {
			break // No more videos
		}

		allVideos = append(allVideos, videos...)

		// If we have enough videos to search through, break
		if len(allVideos) >= 1000 {
			break
		}
	}

	// Filter videos based on search query (case-insensitive)
	var filteredVideos []services.BunnyVideo
	searchLower := strings.ToLower(search)

	for _, video := range allVideos {
		// Search in title, description, and category
		searchableText := strings.ToLower(video.Title)
		if video.Description != nil {
			searchableText += " " + strings.ToLower(*video.Description)
		}
		if video.Category != "" {
			searchableText += " " + strings.ToLower(video.Category)
		}

		// Check if search term is found in the searchable text
		if strings.Contains(searchableText, searchLower) {
			filteredVideos = append(filteredVideos, video)
		}
	}

	// Limit the results
	if len(filteredVideos) > limit {
		filteredVideos = filteredVideos[:limit]
	}

	fmt.Printf("🔍 Server-side search results: %d videos found for '%s' (searched through %d total videos)\n",
		len(filteredVideos), search, len(allVideos))

	return filteredVideos, len(filteredVideos), nil
}

// Update fetchBunnyVideos to support pagination and date sorting
func fetchBunnyVideos(libraryID, apiKey string, page int, itemsPerPage int, search string) ([]services.BunnyVideo, int, error) {
	// Default values
	if page < 1 {
		page = 1
	}
	if itemsPerPage < 1 || itemsPerPage > 100 {
		itemsPerPage = 100
	}

	apiURL := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos?page=%d&itemsPerPage=%d&orderBy=date",
		libraryID, page, itemsPerPage)

	// Note: Bunny.net API doesn't support search parameter, so we ignore it here
	// Search filtering is handled in fetchBunnyVideosWithSearch function

	req, err := http.NewRequest("GET", apiURL, nil)
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
		0,    // createdBy - system
		true, // vid_status
	)
	return err
}

// Video Engagement Analytics Helper Functions

func getVideoTotalViews(db *database.DB, videoID string) int {
	if db == nil {
		return 0
	}
	var count int
	db.QueryRow("SELECT COALESCE(view_count, 0) FROM videos WHERE id = $1", videoID).Scan(&count)
	return count
}

func getVideoUniqueViewers(db *database.DB, videoID string, period string) int {
	if db == nil {
		return 0
	}
	var count int
	query := "SELECT COUNT(DISTINCT user_id) FROM video_access_log WHERE video_id = $1 AND created_at > NOW() - INTERVAL '" + period + "'"
	db.QueryRow(query, videoID).Scan(&count)
	return count
}

func getVideoAverageWatchTime(db *database.DB, videoID string, period string) float64 {
	if db == nil {
		return 0.0
	}
	var avgTime float64
	query := "SELECT AVG(watch_duration) FROM video_access_log WHERE video_id = $1 AND created_at > NOW() - INTERVAL '" + period + "' AND watch_duration IS NOT NULL"
	db.QueryRow(query, videoID).Scan(&avgTime)
	return avgTime
}

func getVideoCompletionRate(db *database.DB, videoID string, period string) float64 {
	if db == nil {
		return 0.0
	}
	var totalViews, completedViews int
	
	// Get total views
	db.QueryRow("SELECT COUNT(*) FROM video_access_log WHERE video_id = $1 AND created_at > NOW() - INTERVAL '" + period + "'", videoID).Scan(&totalViews)
	
	// Get completed views (assuming completion means watching 90%+ of video)
	db.QueryRow("SELECT COUNT(*) FROM video_access_log WHERE video_id = $1 AND created_at > NOW() - INTERVAL '" + period + "' AND completion_percentage >= 90", videoID).Scan(&completedViews)
	
	if totalViews == 0 {
		return 0.0
	}
	return float64(completedViews) / float64(totalViews) * 100
}

func getVideoEngagementScore(db *database.DB, videoID string, period string) float64 {
	if db == nil {
		return 0.0
	}
	
	// Calculate engagement score based on multiple factors
	completionRate := getVideoCompletionRate(db, videoID, period)
	rewatchRate := getVideoRewatchRate(db, videoID, period)
	likeRate := getVideoLikeRate(db, videoID, period)
	shareRate := getVideoShareRate(db, videoID, period)
	
	// Weighted engagement score
	engagementScore := (completionRate * 0.4) + (rewatchRate * 0.3) + (likeRate * 0.2) + (shareRate * 0.1)
	return engagementScore
}

func getVideoViewsByHour(db *database.DB, videoID string, period string) map[string]int {
	if db == nil {
		return map[string]int{}
	}
	
	hourlyViews := map[string]int{}
	rows, err := db.Query(`
		SELECT EXTRACT(HOUR FROM created_at) as hour, COUNT(*) as views
		FROM video_access_log 
		WHERE video_id = $1 AND created_at > NOW() - INTERVAL '` + period + `'
		GROUP BY EXTRACT(HOUR FROM created_at)
		ORDER BY hour
	`, videoID)
	
	if err != nil {
		return hourlyViews
	}
	defer rows.Close()
	
	for rows.Next() {
		var hour int
		var views int
		rows.Scan(&hour, &views)
		hourlyViews[fmt.Sprintf("%02d:00", hour)] = views
	}
	
	return hourlyViews
}

func getVideoViewsByDevice(db *database.DB, videoID string, period string) map[string]int {
	if db == nil {
		return map[string]int{}
	}
	
	deviceViews := map[string]int{}
	rows, err := db.Query(`
		SELECT device_type, COUNT(*) as views
		FROM video_access_log 
		WHERE video_id = $1 AND created_at > NOW() - INTERVAL '` + period + `'
		GROUP BY device_type
	`, videoID)
	
	if err != nil {
		return deviceViews
	}
	defer rows.Close()
	
	for rows.Next() {
		var deviceType string
		var views int
		rows.Scan(&deviceType, &views)
		deviceViews[deviceType] = views
	}
	
	return deviceViews
}

func getVideoViewsByLocation(db *database.DB, videoID string, period string) map[string]int {
	if db == nil {
		return map[string]int{}
	}
	
	locationViews := map[string]int{}
	rows, err := db.Query(`
		SELECT country, COUNT(*) as views
		FROM video_access_log 
		WHERE video_id = $1 AND created_at > NOW() - INTERVAL '` + period + `'
		GROUP BY country
		ORDER BY views DESC
		LIMIT 10
	`, videoID)
	
	if err != nil {
		return locationViews
	}
	defer rows.Close()
	
	for rows.Next() {
		var country string
		var views int
		rows.Scan(&country, &views)
		locationViews[country] = views
	}
	
	return locationViews
}

func getVideoTrafficSources(db *database.DB, videoID string, period string) map[string]int {
	if db == nil {
		return map[string]int{}
	}
	
	trafficSources := map[string]int{}
	rows, err := db.Query(`
		SELECT referrer, COUNT(*) as views
		FROM video_access_log 
		WHERE video_id = $1 AND created_at > NOW() - INTERVAL '` + period + `'
		GROUP BY referrer
		ORDER BY views DESC
		LIMIT 10
	`, videoID)
	
	if err != nil {
		return trafficSources
	}
	defer rows.Close()
	
	for rows.Next() {
		var referrer string
		var views int
		rows.Scan(&referrer, &views)
		if referrer == "" {
			referrer = "Direct"
		}
		trafficSources[referrer] = views
	}
	
	return trafficSources
}

func getVideoDropOffPoints(db *database.DB, videoID string, period string) map[string]float64 {
	if db == nil {
		return map[string]float64{}
	}
	
	dropOffPoints := map[string]float64{}
	rows, err := db.Query(`
		SELECT 
			CASE 
				WHEN completion_percentage < 25 THEN '0-25%'
				WHEN completion_percentage < 50 THEN '25-50%'
				WHEN completion_percentage < 75 THEN '50-75%'
				WHEN completion_percentage < 90 THEN '75-90%'
				ELSE '90-100%'
			END as segment,
			COUNT(*) as views
		FROM video_access_log 
		WHERE video_id = $1 AND created_at > NOW() - INTERVAL '` + period + `'
		GROUP BY segment
		ORDER BY 
			CASE 
				WHEN completion_percentage < 25 THEN 1
				WHEN completion_percentage < 50 THEN 2
				WHEN completion_percentage < 75 THEN 3
				WHEN completion_percentage < 90 THEN 4
				ELSE 5
			END
	`, videoID)
	
	if err != nil {
		return dropOffPoints
	}
	defer rows.Close()
	
	var totalViews int
	segments := make(map[string]int)
	
	for rows.Next() {
		var segment string
		var views int
		rows.Scan(&segment, &views)
		segments[segment] = views
		totalViews += views
	}
	
	// Calculate percentages
	for segment, views := range segments {
		if totalViews > 0 {
			dropOffPoints[segment] = float64(views) / float64(totalViews) * 100
		}
	}
	
	return dropOffPoints
}

func getVideoRewatchRate(db *database.DB, videoID string, period string) float64 {
	if db == nil {
		return 0.0
	}
	var totalViews, rewatches int
	
	// Get total unique viewers
	db.QueryRow("SELECT COUNT(DISTINCT user_id) FROM video_access_log WHERE video_id = $1 AND created_at > NOW() - INTERVAL '" + period + "'", videoID).Scan(&totalViews)
	
	// Get users who watched multiple times
	db.QueryRow(`
		SELECT COUNT(*) FROM (
			SELECT user_id, COUNT(*) as view_count
			FROM video_access_log 
			WHERE video_id = $1 AND created_at > NOW() - INTERVAL '` + period + `'
			GROUP BY user_id
			HAVING COUNT(*) > 1
		) as rewatchers
	`, videoID).Scan(&rewatches)
	
	if totalViews == 0 {
		return 0.0
	}
	return float64(rewatches) / float64(totalViews) * 100
}

func getVideoShareRate(db *database.DB, videoID string, period string) float64 {
	if db == nil {
		return 0.0
	}
	var totalViews, shares int
	
	// Get total views
	db.QueryRow("SELECT COUNT(*) FROM video_access_log WHERE video_id = $1 AND created_at > NOW() - INTERVAL '" + period + "'", videoID).Scan(&totalViews)
	
	// Get shares (assuming shares are tracked in a separate table or field)
	db.QueryRow("SELECT COUNT(*) FROM video_shares WHERE video_id = $1 AND created_at > NOW() - INTERVAL '" + period + "'", videoID).Scan(&shares)
	
	if totalViews == 0 {
		return 0.0
	}
	return float64(shares) / float64(totalViews) * 100
}

func getVideoLikeRate(db *database.DB, videoID string, period string) float64 {
	if db == nil {
		return 0.0
	}
	var totalViews, likes int
	
	// Get total views
	db.QueryRow("SELECT COUNT(*) FROM video_access_log WHERE video_id = $1 AND created_at > NOW() - INTERVAL '" + period + "'", videoID).Scan(&totalViews)
	
	// Get likes
	db.QueryRow("SELECT COALESCE(like_count, 0) FROM videos WHERE id = $1", videoID).Scan(&likes)
	
	if totalViews == 0 {
		return 0.0
	}
	return float64(likes) / float64(totalViews) * 100
}

func trackVideoAnalyticsAccess(db *database.DB, c *gin.Context, videoID string, activity string, metadata map[string]interface{}) {
	if db == nil {
		return
	}
	
	userID, exists := c.Get("user_id")
	if !exists {
		return
	}
	
	metadataJSON, _ := json.Marshal(metadata)
	
	_, err := db.Exec(`
		INSERT INTO user_activity_log (user_id, activity_type, metadata, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`, userID, activity, string(metadataJSON), c.ClientIP(), c.GetHeader("User-Agent"))
	
	if err != nil {
		fmt.Printf("Failed to track video analytics access: %v\n", err)
	}
}
