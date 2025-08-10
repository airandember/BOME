package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bome-backend/internal/config"
	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// Helper function to get status message
func getStatusMessage(statusCode int) string {
	switch statusCode {
	case 200:
		return "Success - API key and library access confirmed"
	case 401:
		return "Unauthorized - Check API key and permissions"
	case 403:
		return "Forbidden - Insufficient permissions"
	case 404:
		return "Not Found - Check library ID"
	case 429:
		return "Rate Limited - Too many requests"
	default:
		return fmt.Sprintf("HTTP %d - Unexpected response", statusCode)
	}
}

// Helper function to get missing fields
func getMissingFields(cfg *config.Config) []string {
	var missing []string
	if cfg.BunnyStreamLibrary == "" {
		missing = append(missing, "BUNNY_STREAM_LIBRARY_ID")
	}
	if cfg.BunnyStreamAPIKey == "" {
		missing = append(missing, "BUNNY_STREAM_API_KEY")
	}
	if cfg.BunnyStorageZone == "" {
		missing = append(missing, "BUNNY_STORAGE_ZONE")
	}
	if cfg.BunnyAPIKey == "" {
		missing = append(missing, "BUNNY_API_KEY")
	}
	if cfg.BunnyPullZone == "" {
		missing = append(missing, "BUNNY_PULL_ZONE")
	}
	return missing
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SetupRoutes configures all routes for the application
func SetupRoutes(
	router *gin.Engine,
	cfg *config.Config,
	db *database.DB,
	redis *database.Redis,
	bunnyService *services.BunnyService,
	stripeService *services.StripeService,
	spacesService *services.SpacesService,
	emailService *services.EmailService,
	biService *services.BusinessIntelligenceService,
) {
	// Debug logging
	fmt.Printf("Setting up routes...\n")

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "bome-streaming-backend",
		})
	})
	fmt.Printf("Registered health check endpoint\n")

	// API v1 routes
	v1 := router.Group("/api/v1")
	fmt.Printf("Created v1 route group with base path: %s\n", v1.BasePath())

	// Admin routes
	admin := v1.Group("/admin")
	SetupAdminRoutes(admin, db)

	// Create plan history service for analytics
	planHistoryService := services.NewPlanHistoryService(db)
	SetupAnalyticsRoutes(admin, db, planHistoryService)

	// Create admin cache service
	analyticsService := services.NewSubscriptionAnalyticsService(db)
	SetupAdminStreamingRoutes(admin, db, stripeService, analyticsService, biService)
	SetupMasterVideoRoutes(admin, db, bunnyService)

	// Initialize subscription services
	subscriptionPlanService := services.NewSubscriptionPlanService(db)
	subscriberService := services.NewSubscriberService(db)
	subscriptionOffersService := services.NewSubscriptionOffersService(db)
	subscriberHistoryService := services.NewSubscriberHistoryService(db)

	// Setup subscription-related routes under admin group
	fmt.Printf("Setting up subscription plan routes...\n")
	SetupSubscriptionPlanRoutes(admin, db, subscriptionPlanService)
	fmt.Printf("Setting up subscription offers routes...\n")
	SetupSubscriptionOfferRoutes(router, db, subscriptionOffersService)
	fmt.Printf("Setting up subscriber routes...\n")
	SetupSubscriberRoutes(admin, db, subscriberService)
	fmt.Printf("Setting up subscriber history routes...\n")
	SetupSubscriberHistoryRoutes(admin, db, subscriberHistoryService)
	SetupSubscriptionRoutes(router, db, stripeService, analyticsService)

	// Public subscription plan routes using existing functions
	publicPlans := v1.Group("/subscription-plans")
	{
		// Get all subscription data (plans + offers) - MUST come before /:id
		publicPlans.GET("/all", func(c *gin.Context) {
			getAllSubscriptionData(c, subscriptionPlanService, subscriptionOffersService)
		})

		// Get active subscription plans
		publicPlans.GET("/active", func(c *gin.Context) {
			getActiveSubscriptionPlans(c, subscriptionPlanService)
		})

		// Get promoted subscription plans
		publicPlans.GET("/promoted", func(c *gin.Context) {
			getPromotedSubscriptionPlans(c, subscriptionPlanService)
		})

		// Get subscription plan by ID (public) - MUST come last
		publicPlans.GET("/:id", func(c *gin.Context) {
			getSubscriptionPlanPublic(c, subscriptionPlanService)
		})
	}

	fmt.Printf("Admin routes setup complete\n")

	// Setup all mock data routes for development/testing
	fmt.Printf("Setting up mock data routes...\n")
	SetupMockDataRoutes(v1)
	SetupArticlesRoutes(v1)
	SetupRolesRoutes(v1)
	SetupStandardizedRolesRoutes(v1)
	SetupYouTubeRoutes(v1, db)
	fmt.Printf("Mock data routes setup complete\n")

	// Real authentication routes
	auth := v1.Group("/auth")
	{
		auth.POST("/login", LoginHandler(db))
		auth.POST("/register", RegisterHandler(db, emailService))
		auth.POST("/logout", LogoutHandler(db))
	}

	// Video routes using database handlers with bunny.net integration
	videos := v1.Group("/videos")
	{
		fmt.Printf("Setting up video routes...\n")

		// Get all videos with pagination and filtering
		videos.GET("", middleware.AuthRequired(), middleware.SessionActivityTracker(db), func(c *gin.Context) {
			// Parse query parameters
			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
			category := c.DefaultQuery("category", "")

			// Validate parameters
			if limit > 100 {
				limit = 100
			}
			if limit < 1 {
				limit = 20
			}

			// Calculate offset
			offset := (page - 1) * limit

			// Get videos from database
			videos, err := db.GetVideos(limit, offset, category, "")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to fetch videos",
					"details": err.Error(),
				})
				return
			}

			// Transform videos to API response format
			var responseVideos []gin.H
			for _, video := range videos {
				// Get play data from Bunny.net if available
				var playData *services.VideoPlayData
				if video.BunnyVideoID != "" {
					playData, _ = bunnyService.GetVideoPlayData(video.BunnyVideoID)
				}

				responseVideo := gin.H{
					"id":           video.ID,
					"title":        video.Title,
					"description":  video.Description,
					"bunnyVideoId": video.BunnyVideoID,
					"thumbnailUrl": video.ThumbnailURL,
					"duration":     video.Duration,
					"fileSize":     video.FileSize,
					"status":       video.Status,
					"category":     video.Category,
					"tags":         video.Tags,
					"viewCount":    video.ViewCount,
					"likeCount":    video.LikeCount,
					"createdAt":    video.CreatedAt.Format(time.RFC3339),
					"updatedAt":    video.UpdatedAt.Format(time.RFC3339),
				}

				// Add Bunny.net play data if available
				if playData != nil {
					responseVideo["playData"] = playData
					responseVideo["iframeSrc"] = playData.IframeSrc
					responseVideo["directPlayUrl"] = playData.DirectPlayURL
					responseVideo["playbackUrl"] = playData.DirectPlayURL
					responseVideo["resolutions"] = playData.ResolutionOptions
				}

				responseVideos = append(responseVideos, responseVideo)
			}

			// Get total count for pagination
			totalCount, err := db.GetVideoCount()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to get video count",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"videos":  responseVideos,
				"pagination": gin.H{
					"current_page": page,
					"per_page":     limit,
					"total":        totalCount,
					"total_pages":  (totalCount + limit - 1) / limit,
					"has_more":     page*limit < totalCount,
				},
			})
		})

		// Test endpoint to verify route registration
		videos.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Video test endpoint working"})
		})

		videos.GET("/categories", GetMockCategoriesHandler) // Must come before /:id
		videos.GET("/:id", middleware.AuthRequired(), middleware.SessionActivityTracker(db), func(c *gin.Context) {
			videoID := c.Param("id")
			if videoID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Video ID is required"})
				return
			}

			fmt.Printf("Fetching video with ID: %s\n", videoID)

			// First try to get video from database using numeric ID
			videoIDInt, err := strconv.Atoi(videoID)
			if err == nil {
				// It's a numeric ID, get from database
				video, err := db.GetVideoByID(videoIDInt)
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{
						"error":   "Video not found",
						"details": err.Error(),
					})
					return
				}

				// If video has a Bunny.net ID, get the play data
				if video.BunnyVideoID != "" {
					playData, err := bunnyService.GetVideoPlayData(video.BunnyVideoID)
					if err != nil {
						fmt.Printf("Failed to get play data: %v\n", err)
						// Continue without play data
					}

					if playData != nil {
						playDataMap := make(map[string]interface{})
						playDataBytes, err := json.Marshal(playData)
						if err == nil {
							json.Unmarshal(playDataBytes, &playDataMap)
							video.PlayData = playDataMap
						}
						video.IframeSrc = playData.IframeSrc
						video.DirectPlayURL = playData.DirectPlayURL
						video.PlaybackURL = playData.DirectPlayURL // Use HLS stream URL for playback
						video.Resolutions = playData.ResolutionOptions
					}
				}

				c.JSON(http.StatusOK, video)
				return
			}

			// If not a numeric ID, try to get from database by Bunny ID
			video, err := db.GetVideoByBunnyID(videoID)
			if err == nil {
				// Found in database, get fresh play data
				playData, err := bunnyService.GetVideoPlayData(videoID)
				if err != nil {
					fmt.Printf("Failed to get play data: %v\n", err)
					// Continue without play data
				}

				if playData != nil {
					playDataMap := make(map[string]interface{})
					playDataBytes, err := json.Marshal(playData)
					if err == nil {
						json.Unmarshal(playDataBytes, &playDataMap)
						video.PlayData = playDataMap
					}
					video.IframeSrc = playData.IframeSrc
					video.DirectPlayURL = playData.DirectPlayURL
					video.PlaybackURL = playData.DirectPlayURL
					video.Resolutions = playData.ResolutionOptions
				}

				c.JSON(http.StatusOK, video)
				return
			}

			// If not found in database, try to fetch directly from Bunny.net
			bunnyVideo, err := bunnyService.GetVideo(videoID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"error":   "Video not found",
					"details": err.Error(),
				})
				return
			}

			// Get video play data
			playData, err := bunnyService.GetVideoPlayData(videoID)
			if err != nil {
				fmt.Printf("Failed to get play data: %v\n", err)
				// Continue without play data
			}

			// Return the full Bunny.net response
			response := gin.H{
				"videoLibraryId":       bunnyVideo.VideoLibraryID,
				"guid":                 bunnyVideo.GUID,
				"title":                bunnyVideo.Title,
				"description":          bunnyVideo.Description,
				"dateUploaded":         bunnyVideo.DateUploaded,
				"views":                bunnyVideo.Views,
				"isPublic":             bunnyVideo.IsPublic,
				"length":               bunnyVideo.Length,
				"status":               bunnyVideo.Status,
				"framerate":            bunnyVideo.Framerate,
				"width":                bunnyVideo.Width,
				"height":               bunnyVideo.Height,
				"availableResolutions": bunnyVideo.AvailableResolutions,
				"outputCodecs":         "x264", // This seems to be fixed in your example
				"thumbnailCount":       bunnyVideo.ThumbnailCount,
				"encodeProgress":       bunnyVideo.EncodeProgress,
				"storageSize":          bunnyVideo.StorageSize,
				"hasMP4Fallback":       bunnyVideo.HasMP4Fallback,
				"collectionId":         bunnyVideo.CollectionID,
				"thumbnailFileName":    bunnyVideo.ThumbnailFileName,
				"averageWatchTime":     bunnyVideo.AverageWatchTime,
				"totalWatchTime":       bunnyVideo.TotalWatchTime,
				"category":             bunnyVideo.Category,
				"captions":             []interface{}{}, // Empty array as shown in your example
				"chapters":             []interface{}{},
				"moments":              []interface{}{},
				"metaTags":             []interface{}{},
				"jitEncodingEnabled":   false,
			}

			if playData != nil {
				response["playData"] = playData
				response["iframeSrc"] = playData.IframeSrc
				response["directPlayUrl"] = playData.DirectPlayURL
				response["thumbnailUrl"] = playData.ThumbnailURL
				response["resolutions"] = playData.ResolutionOptions
			}

			c.JSON(http.StatusOK, response)
		})

		videos.GET("/:id/comments", GetMockCommentsHandler)

		// Add secure video upload endpoint - RESTRICTED TO ADMINS AND CONTENT MANAGERS
		videos.POST("/upload",
			middleware.AuthRequired(),
			middleware.SessionActivityTracker(db),
			middleware.VideoUploadRequired(),
			UploadVideoHandler(db, bunnyService))

		// Add streaming endpoint for frontend
		videos.GET("/:id/stream", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Video streaming endpoint"})
		})

		// Add blob URL endpoint for direct video data access
		videos.GET("/:id/blob", middleware.AuthRequired(), func(c *gin.Context) {
			videoID := c.Param("id")

			// Get user info from context
			userID := c.GetInt("user_id")
			if userID == 0 {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
				return
			}

			fmt.Printf("[Blob] Request for video: %s by user: %d\n", videoID, userID)

			// Get the direct video URL from Bunny
			directURL := fmt.Sprintf("https://vz-%s-%s.b-cdn.net/%s/play_720p.mp4",
				bunnyService.GetStreamLibrary(),
				bunnyService.GetRegion(),
				videoID)

			fmt.Printf("[Blob] Fetching from: %s\n", directURL)

			// Create the request
			req, err := http.NewRequest("GET", directURL, nil)
			if err != nil {
				fmt.Printf("[Blob] Failed to create request: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
				return
			}

			// Add headers
			req.Header.Set("Accept", "video/mp4,*/*")
			req.Header.Set("User-Agent", "BOME-Backend/1.0")

			// Try without authentication first
			client := &http.Client{Timeout: 60 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("[Blob] Request failed: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch video"})
				return
			}
			defer resp.Body.Close()

			fmt.Printf("[Blob] Response status: %d\n", resp.StatusCode)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				fmt.Printf("[Blob] Error response: %s\n", string(body))
				c.JSON(resp.StatusCode, gin.H{"error": "Video not accessible"})
				return
			}

			// Set response headers for blob creation
			c.Header("Content-Type", "video/mp4")
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
			c.Header("Cache-Control", "public, max-age=3600")

			// Copy content length if available
			if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
				c.Header("Content-Length", contentLength)
			}

			// Stream the video data
			c.Status(http.StatusOK)
			written, err := io.Copy(c.Writer, resp.Body)
			if err != nil {
				fmt.Printf("[Blob] Error streaming: %v\n", err)
			} else {
				fmt.Printf("[Blob] Successfully streamed %d bytes\n", written)
			}
		})

		fmt.Printf("Video routes setup complete\n")
	}

	// Test endpoint to verify route registration
	v1.GET("/test/optimization", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":   "Optimization test endpoint working",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Performance monitoring endpoint
	v1.GET("/performance/metrics", func(c *gin.Context) {
		// Get metrics from the global optimized Bunny service if available
		if optimizedService := services.GetGlobalOptimizedBunnyService(); optimizedService != nil {
			metrics := optimizedService.GetMetrics()
			c.JSON(http.StatusOK, gin.H{
				"success":   true,
				"metrics":   metrics,
				"timestamp": time.Now().Format(time.RFC3339),
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"metrics": gin.H{
					"message":      "Optimized service not available",
					"service_type": "standard",
				},
				"timestamp": time.Now().Format(time.RFC3339),
			})
		}
	})

	// Bunny.net direct access endpoint (separate from videos to avoid conflicts)
	v1.GET("/bunny-videos", GetVideosFromBunnyHandler(db, bunnyService))

	// Add single video endpoint
	v1.GET("/bunny-videos/:id", func(c *gin.Context) {
		videoID := c.Param("id")
		if videoID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Video ID is required",
				"code":  "MISSING_VIDEO_ID",
			})
			return
		}

		// Log the request
		fmt.Printf("Fetching video with Bunny ID: %s\n", videoID)

		// Always fetch fresh data from Bunny.net
		bunnyVideo, err := bunnyService.GetVideo(videoID)
		if err != nil {
			fmt.Printf("Bunny.net fetch failed for video %s: %v\n", videoID, err)
			c.JSON(http.StatusNotFound, gin.H{
				"error":    "Video not found",
				"code":     "VIDEO_NOT_FOUND",
				"details":  err.Error(),
				"bunny_id": videoID,
			})
			return
		}

		// Get video play data
		_, err = bunnyService.GetVideoPlayData(videoID)
		if err != nil {
			fmt.Printf("Failed to get video play data: %v\n", err)
			// Don't return error, just continue without play data
		}

		// Create description string if it's null
		description := ""
		if bunnyVideo.Description != nil {
			description = *bunnyVideo.Description
		}

		// Map Bunny.net status to our status
		status := "processing"
		switch bunnyVideo.Status {
		case 0:
			status = "created"
		case 1:
			status = "uploaded"
		case 2:
			status = "processing"
		case 3:
			status = "transcoding"
		case 4:
			status = "ready" // Finished = Ready for playback
		case 5:
			status = "error"
		case 6:
			status = "upload_failed"
		case 7:
			status = "jit_segmenting"
		case 8:
			status = "jit_playlists_created"
		default:
			status = "unknown"
		}

		// Check if video exists in our database
		dbVideo, err := db.GetVideoByBunnyID(videoID)
		if err != nil {
			// Video doesn't exist, create it
			dbVideo, err = db.CreateVideo(
				bunnyVideo.Title,
				description,
				bunnyVideo.GUID,
				bunnyService.GetThumbnailURL(bunnyVideo.GUID),
				bunnyVideo.Category,
				bunnyVideo.Length,
				bunnyVideo.StorageSize,
				[]string{},
				1,    // createdBy - system //SHOULD WE CONSIDER CHANGING THIS TO USER ID?
				true, // vid_status
			)
			if err != nil {
				fmt.Printf("Failed to create video in database: %v\n", err)
			}
		} else {
			// Video exists, check if it needs updating
			updates := make(map[string]interface{})

			if dbVideo.Title != bunnyVideo.Title {
				updates["title"] = bunnyVideo.Title
			}
			if dbVideo.Description != description {
				updates["description"] = description
			}
			if dbVideo.Category != bunnyVideo.Category {
				updates["category"] = bunnyVideo.Category
			}
			if dbVideo.Status != status {
				updates["status"] = status
			}
			if dbVideo.Duration != bunnyVideo.Length {
				updates["duration"] = bunnyVideo.Length
			}
			if dbVideo.FileSize != bunnyVideo.StorageSize {
				updates["file_size"] = bunnyVideo.StorageSize
			}
			if dbVideo.ViewCount != bunnyVideo.Views {
				updates["view_count"] = bunnyVideo.Views
			}

			// If we have updates, apply them
			if len(updates) > 0 {
				err = db.UpdateVideo(dbVideo.ID, updates)
				if err != nil {
					fmt.Printf("Failed to update video in database: %v\n", err)
				} else {
					fmt.Printf("Updated video %s in database with changes: %+v\n", videoID, updates)
				}
			}
		}

		// Return the full Bunny.net response
		response := gin.H{
			"id":             bunnyVideo.GUID,
			"title":          bunnyVideo.Title,
			"description":    description,
			"status":         status,
			"duration":       bunnyVideo.Length,
			"views":          bunnyVideo.Views,
			"thumbnailUrl":   bunnyService.GetThumbnailURL(bunnyVideo.GUID),
			"videoUrl":       bunnyService.GetStreamURL(bunnyVideo.GUID),
			"iframeSrc":      bunnyService.GetIframeURL(bunnyVideo.GUID),
			"playbackUrl":    bunnyService.GetStreamURL(bunnyVideo.GUID),
			"createdAt":      bunnyVideo.DateUploaded,
			"updatedAt":      bunnyVideo.DateUploaded,
			"fileSize":       bunnyVideo.StorageSize,
			"resolution":     fmt.Sprintf("%dx%d", bunnyVideo.Width, bunnyVideo.Height),
			"category":       bunnyVideo.Category,
			"tags":           []string{}, // Bunny.net doesn't provide tags
			"encodeProgress": 0,          // Bunny.net doesn't provide this
			"storageSize":    bunnyVideo.StorageSize,
		}

		c.JSON(http.StatusOK, response)
	})

	// Add PUT route for updating Bunny.net video metadata
	v1.PUT("/bunny-videos/:id", middleware.AuthRequired(), middleware.SessionActivityTracker(db), func(c *gin.Context) {
		videoID := c.Param("id")
		if videoID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Video ID is required"})
			return
		}

		var updateData map[string]interface{}
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get video from database by Bunny ID
		dbVideo, err := db.GetVideoByBunnyID(videoID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found in database"})
			return
		}

		// Update video in database
		if err := db.UpdateVideo(dbVideo.ID, updateData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update video"})
			return
		}

		// Log admin action
		userID := c.GetInt("user_id")
		go db.CreateAdminLog(&userID, "bunny_video_updated", "video", &dbVideo.ID, updateData, c.ClientIP(), c.GetHeader("User-Agent"))

		c.JSON(http.StatusOK, gin.H{"message": "Video updated successfully"})
	})

	// Add DELETE route for deleting Bunny.net videos
	v1.DELETE("/bunny-videos/:id", middleware.AuthRequired(), middleware.SessionActivityTracker(db), func(c *gin.Context) {
		videoID := c.Param("id")
		if videoID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Video ID is required"})
			return
		}

		// Get video from database by Bunny ID
		dbVideo, err := db.GetVideoByBunnyID(videoID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found in database"})
			return
		}

		// Delete from Bunny.net first
		if err := bunnyService.DeleteVideo(videoID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete video from Bunny.net"})
			return
		}

		// Delete from database
		if err := db.DeleteVideo(dbVideo.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete video from database"})
			return
		}

		// Log admin action
		userID := c.GetInt("user_id")
		go db.CreateAdminLog(&userID, "bunny_video_deleted", "video", &dbVideo.ID, map[string]interface{}{"title": dbVideo.Title}, c.ClientIP(), c.GetHeader("User-Agent"))

		c.JSON(http.StatusOK, gin.H{"message": "Video deleted successfully"})
	})

	// Bunny.net collections endpoints
	v1.GET("/bunny-collections", func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

		collections, err := bunnyService.GetCollections(page, perPage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to fetch collections: %v", err),
			})
			return
		}

		c.JSON(http.StatusOK, collections)
	})

	v1.GET("/bunny-collections/:id", func(c *gin.Context) {
		collectionID := c.Param("id")
		if collectionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Collection ID is required"})
			return
		}

		collection, err := bunnyService.GetCollection(collectionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to fetch collection: %v", err),
			})
			return
		}

		c.JSON(http.StatusOK, collection)
	})

	// Get videos by collection ID - PREMIUM FEATURE
	v1.GET("/bunny-collections/:id/videos", middleware.AuthRequired(), middleware.SessionActivityTracker(db), func(c *gin.Context) {
		collectionID := c.Param("id")
		if collectionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Collection ID is required"})
			return
		}

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

		videos, totalItems, err := bunnyService.GetVideosByCollection(collectionID, page, perPage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to fetch videos for collection: %v", err),
			})
			return
		}

		// Transform videos to API response format
		var responseVideos []gin.H
		for _, bunnyVideo := range videos {
			streamURL := bunnyService.GetStreamURL(bunnyVideo.GUID)
			thumbnailURL := bunnyService.GetThumbnailURLWithFilename(bunnyVideo.GUID, bunnyVideo.ThumbnailFileName)
			iframeURL := bunnyService.GetIframeURL(bunnyVideo.GUID)

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
				"likeCount":    0, // Bunny.net doesn't provide likes, default to 0
				"category":     bunnyVideo.Category,
				"tags":         []string{}, // Bunny.net doesn't provide tags, default to empty array
				"status":       mapBunnyStatus(bunnyVideo.Status),
				"createdAt":    bunnyVideo.DateUploaded,
				"updatedAt":    bunnyVideo.DateUploaded,
				"bunnyVideoId": bunnyVideo.GUID, // Add this field for frontend compatibility
				"collectionId": bunnyVideo.CollectionID,
			}
			responseVideos = append(responseVideos, responseVideo)
		}

		// Calculate pagination info
		totalPages := (totalItems + perPage - 1) / perPage
		hasMore := page < totalPages

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"videos":  responseVideos,
			"pagination": gin.H{
				"current_page": page,
				"per_page":     perPage,
				"total":        totalItems,
				"total_pages":  totalPages,
				"has_more":     hasMore,
			},
			"collection_id": collectionID,
		})
	})
}

// UploadVideoHandler handles secure video uploads via backend - ADMIN/CONTENT MANAGER ONLY
func UploadVideoHandler(db *database.DB, bunnyService *services.BunnyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID and role from context
		userID := c.GetInt("user_id")
		userRole := c.GetString("user_role")

		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Parse multipart form
		if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32MB max
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
			return
		}

		// Get video file
		file, header, err := c.Request.FormFile("video")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No video file provided"})
			return
		}
		defer file.Close()

		// Validate file type
		if !isValidVideoFile(header.Filename) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video file type. Allowed: mp4, avi, mov, wmv, flv, webm, mkv"})
			return
		}

		// Get metadata from form
		title := c.PostForm("title")
		description := c.PostForm("description")
		category := c.PostForm("category")
		tagsStr := c.PostForm("tags")

		if title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
			return
		}

		// Parse tags
		var tags []string
		if tagsStr != "" {
			if err := json.Unmarshal([]byte(tagsStr), &tags); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tags format"})
				return
			}
		}

		// Create a temporary file to pass to Bunny service
		tempFile, err := os.CreateTemp("", "upload-*.tmp")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temporary file"})
			return
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()

		// Copy uploaded file to temp file
		if _, err := io.Copy(tempFile, file); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded file"})
			return
		}

		// Reset file pointer for reading
		tempFile.Seek(0, 0)

		// Create multipart file header for Bunny service
		fileHeader := &multipart.FileHeader{
			Filename: header.Filename,
			Header:   header.Header,
		}

		// Upload to Bunny.net
		uploadResp, err := bunnyService.UploadVideo(fileHeader, title, description)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload video: " + err.Error()})
			return
		}

		// Save video metadata to database
		video, err := db.CreateVideo(
			title,
			description,
			uploadResp.VideoID,
			"", // thumbnail URL will be set later
			category,
			0, // duration will be updated when processing is complete
			header.Size,
			tags,
			1,    // createdBy - system //SHOULD WE CONSIDER CHANGING THIS TO USER ID?
			true, // vid_status
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save video metadata"})
			return
		}

		// Log the upload action
		go db.CreateAdminLog(&userID, "video_uploaded", "video", &video.ID, map[string]interface{}{
			"title":     video.Title,
			"bunny_id":  video.BunnyVideoID,
			"file_size": header.Size,
		}, c.ClientIP(), c.GetHeader("User-Agent"))

		// Return success response
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Video uploaded successfully",
			"video": gin.H{
				"id":          video.ID,
				"title":       video.Title,
				"bunny_id":    video.BunnyVideoID,
				"status":      video.Status,
				"uploaded_at": video.CreatedAt,
				"uploaded_by": userRole,
			},
		})
	}
}

// isValidVideoFile checks if the file is a valid video format
func isValidVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm", ".mkv"}

	for _, validExt := range validExtensions {
		if ext == validExt {
			return true
		}
	}
	return false
}

// RoleRequired middleware that requires specific roles
func RoleRequired(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(401, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		roleStr := userRole.(string)
		for _, role := range allowedRoles {
			if roleStr == role {
				c.Next()
				return
			}
		}

		c.JSON(403, gin.H{"error": "Insufficient permissions"})
		c.Abort()
	}
}
