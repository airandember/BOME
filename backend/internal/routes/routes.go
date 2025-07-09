package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"bome-backend/internal/config"
	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

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
	SetupAnalyticsRoutes(admin)
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

		videos.GET("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Videos endpoint - use /bunny for Bunny.net videos"})
		})

		// Test endpoint to verify route registration
		videos.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Video test endpoint working"})
		})

		videos.GET("/categories", GetMockCategoriesHandler) // Must come before /:id
		videos.GET("/:id", func(c *gin.Context) {
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

			// Create response
			response := gin.H{
				"id":            videoID,
				"title":         bunnyVideo.Title,
				"description":   bunnyVideo.Description,
				"status":        bunnyVideo.Status,
				"created_at":    bunnyVideo.CreatedAt,
				"updated_at":    bunnyVideo.UpdatedAt,
				"bunny_id":      videoID,
				"thumbnail_url": bunnyService.GetThumbnailURL(videoID),
				"duration":      bunnyVideo.Duration,
				"size":          bunnyVideo.Size,
				"preview":       bunnyVideo.Preview,
				"library_id":    bunnyVideo.LibraryID,
			}

			if playData != nil {
				response["play_data"] = playData
				response["iframe_src"] = playData.IframeSrc
				response["direct_play_url"] = playData.DirectPlayURL
				response["resolutions"] = playData.ResolutionOptions
				response["playback_url"] = playData.DirectPlayURL // Use HLS stream URL for playback
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

		fmt.Printf("Video routes setup complete\n")
	}

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

		// First try to get from database
		video, err := db.GetVideoByBunnyID(videoID)
		if err != nil {
			fmt.Printf("Database lookup failed for video %s: %v\n", videoID, err)

			// If not in database, try to fetch from Bunny.net
			fmt.Printf("Attempting to fetch video %s from Bunny.net\n", videoID)
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

			fmt.Printf("Successfully fetched video from Bunny.net: %+v\n", bunnyVideo)

			// Create video in database
			video, err = db.CreateVideo(
				bunnyVideo.Title,
				bunnyVideo.Description,
				bunnyVideo.ID,
				bunnyService.GetThumbnailURL(bunnyVideo.ID),
				"", // Category not available in BunnyVideo
				int(bunnyVideo.Duration),
				bunnyVideo.Size,
				[]string{}, // No tags initially
				1,          // Default admin user
			)
			if err != nil {
				fmt.Printf("Failed to save video %s to database: %v\n", videoID, err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":    "Failed to save video metadata",
					"code":     "DATABASE_ERROR",
					"details":  err.Error(),
					"bunny_id": videoID,
				})
				return
			}

			fmt.Printf("Successfully created database entry for video %s\n", videoID)
		} else {
			fmt.Printf("Found existing video in database: %+v\n", video)
		}

		// Get video play data
		playData, err := bunnyService.GetVideoPlayData(videoID)
		if err != nil {
			fmt.Printf("Failed to get video play data: %v\n", err)
			// Don't return error, just continue without play data
		}

		// Combine video data with play data
		response := gin.H{
			"id":          video.ID,
			"title":       video.Title,
			"description": video.Description,
			"bunny_id":    video.BunnyVideoID,
			"status":      video.Status,
			"created_at":  video.CreatedAt,
			"updated_at":  video.UpdatedAt,
		}

		if playData != nil {
			response["play_data"] = playData
			response["iframe_src"] = playData.IframeSrc
			response["direct_play_url"] = playData.DirectPlayURL
			response["thumbnail_url"] = playData.ThumbnailURL
			response["resolutions"] = playData.ResolutionOptions
		}

		c.JSON(http.StatusOK, response)
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

	// Subscription routes
	subscriptions := v1.Group("/subscriptions")
	{
		subscriptions.GET("/plans", GetSubscriptionPlansHandler(stripeService))
		subscriptions.GET("/current", middleware.AuthRequired(), middleware.SessionActivityTracker(db), GetSubscriptionHandler(db))
		subscriptions.POST("", middleware.AuthRequired(), middleware.SessionActivityTracker(db), CreateSubscriptionHandler(db))
		subscriptions.POST("/:id/cancel", middleware.AuthRequired(), middleware.SessionActivityTracker(db), CancelSubscriptionHandler(db))
		subscriptions.POST("/checkout", CreateCheckoutSessionHandler(stripeService))
	}

	// User profile routes
	users := v1.Group("/users")
	{
		users.GET("/profile", middleware.AuthRequired(), middleware.SessionActivityTracker(db), GetProfileHandler(db))
		users.PUT("/profile", middleware.AuthRequired(), middleware.SessionActivityTracker(db), UpdateProfileHandler(db))
	}

	// User dashboard
	v1.GET("/dashboard", GetDashboardDataHandler)

	// Advertisement routes for public ad serving
	ads := v1.Group("/ads")
	{
		ads.GET("/serve/:placement", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Ad serving endpoint (mock)"})
		})
	}

	// Bunny.net test endpoint
	v1.GET("/test/bunny", func(c *gin.Context) {
		// Enhanced configuration check with more details
		configStatus := gin.H{
			"stream_library": gin.H{
				"value": cfg.BunnyStreamLibrary,
				"set":   cfg.BunnyStreamLibrary != "",
				"valid": len(cfg.BunnyStreamLibrary) > 0,
			},
			"stream_api_key": gin.H{
				"set":     cfg.BunnyStreamAPIKey != "",
				"length":  len(cfg.BunnyStreamAPIKey),
				"valid":   len(cfg.BunnyStreamAPIKey) >= 40, // Bunny.net API keys are typically long
				"preview": cfg.BunnyStreamAPIKey[:min(8, len(cfg.BunnyStreamAPIKey))] + "...",
			},
			"storage_zone": gin.H{
				"value": cfg.BunnyStorageZone,
				"set":   cfg.BunnyStorageZone != "",
				"valid": len(cfg.BunnyStorageZone) > 0,
			},
			"storage_api_key": gin.H{
				"set":     cfg.BunnyAPIKey != "",
				"length":  len(cfg.BunnyAPIKey),
				"valid":   len(cfg.BunnyAPIKey) >= 40,
				"preview": cfg.BunnyAPIKey[:min(8, len(cfg.BunnyAPIKey))] + "...",
			},
			"pull_zone": gin.H{
				"value": cfg.BunnyPullZone,
				"set":   cfg.BunnyPullZone != "",
				"valid": len(cfg.BunnyPullZone) > 0,
			},
			"region": gin.H{
				"value": cfg.BunnyRegion,
				"set":   cfg.BunnyRegion != "",
				"valid": len(cfg.BunnyRegion) > 0,
			},
			"webhook_secret": gin.H{
				"set":    cfg.BunnyWebhookSecret != "",
				"length": len(cfg.BunnyWebhookSecret),
				"valid":  len(cfg.BunnyWebhookSecret) >= 10,
			},
		}

		// Calculate overall configuration status
		requiredFields := []bool{
			cfg.BunnyStreamLibrary != "",
			cfg.BunnyStreamAPIKey != "",
			cfg.BunnyStorageZone != "",
			cfg.BunnyAPIKey != "",
			cfg.BunnyPullZone != "",
		}

		allConfigured := true
		for _, field := range requiredFields {
			if !field {
				allConfigured = false
				break
			}
		}

		// Test API connectivity if configured
		var connectivityTest gin.H
		if allConfigured {
			// Test Stream API
			streamURL := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos", cfg.BunnyStreamLibrary)
			streamReq, streamErr := http.NewRequest("GET", streamURL, nil)
			if streamErr == nil {
				streamReq.Header.Set("AccessKey", cfg.BunnyStreamAPIKey)
				streamReq.Header.Set("Content-Type", "application/json")

				client := &http.Client{Timeout: 5 * time.Second}
				streamResp, streamRespErr := client.Do(streamReq)

				if streamRespErr == nil {
					defer streamResp.Body.Close()
					connectivityTest = gin.H{
						"stream_api": gin.H{
							"status":        "tested",
							"url":           streamURL,
							"response_code": streamResp.StatusCode,
							"success":       streamResp.StatusCode == 200,
							"message":       getStatusMessage(streamResp.StatusCode),
						},
					}
				} else {
					connectivityTest = gin.H{
						"stream_api": gin.H{
							"status":  "error",
							"url":     streamURL,
							"error":   streamRespErr.Error(),
							"success": false,
						},
					}
				}
			}
		} else {
			connectivityTest = gin.H{
				"stream_api": gin.H{
					"status":  "skipped",
					"message": "Configuration incomplete - cannot test connectivity",
					"success": false,
				},
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"status":       "configuration_check",
			"timestamp":    time.Now().Format(time.RFC3339),
			"config":       configStatus,
			"connectivity": connectivityTest,
			"summary": gin.H{
				"all_configured":    allConfigured,
				"ready_for_testing": allConfigured,
				"missing_fields":    getMissingFields(cfg),
			},
		})
	})

	// Bunny.net connection test endpoint
	v1.GET("/test/bunny/connect", func(c *gin.Context) {
		if bunnyService == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Bunny service not configured",
				"config": gin.H{
					"streamLibrary": cfg.BunnyStreamLibrary,
					"streamAPIKey":  cfg.BunnyStreamAPIKey != "",
					"storageZone":   cfg.BunnyStorageZone,
					"apiKey":        cfg.BunnyAPIKey != "",
					"pullZone":      cfg.BunnyPullZone,
				},
			})
			return
		}

		// Test Bunny Stream library access
		libraryID := cfg.BunnyStreamLibrary
		apiKey := cfg.BunnyStreamAPIKey

		if libraryID == "" || apiKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Bunny Stream library ID or API key not configured",
				"config": gin.H{
					"streamLibrary": libraryID,
					"streamAPIKey":  apiKey != "",
				},
			})
			return
		}

		// Make a test request to Bunny Stream API
		url := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos", libraryID)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create request",
				"details": err.Error(),
			})
			return
		}

		req.Header.Set("AccessKey", apiKey)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to connect to Bunny Stream API",
				"details": err.Error(),
				"url":     url,
			})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			c.JSON(http.StatusOK, gin.H{
				"status":        "success",
				"message":       "Successfully connected to Bunny Stream library",
				"library_id":    libraryID,
				"response_code": resp.StatusCode,
				"config": gin.H{
					"streamLibrary": libraryID,
					"streamAPIKey":  apiKey != "",
					"storageZone":   cfg.BunnyStorageZone,
					"apiKey":        cfg.BunnyAPIKey != "",
					"pullZone":      cfg.BunnyPullZone,
				},
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":         "Failed to access Bunny Stream library",
				"response_code": resp.StatusCode,
				"library_id":    libraryID,
				"url":           url,
			})
		}
	})

	// Manual sync endpoint for Bunny.net videos
	v1.POST("/admin/sync-bunny-videos", middleware.AuthRequired(), func(c *gin.Context) {
		// Check if user is admin
		userRole := c.GetString("user_role")
		if userRole != "admin" && userRole != "super_admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			return
		}

		// Fetch videos from Bunny.net
		videos, err := fetchBunnyVideos(cfg.BunnyStreamLibrary, cfg.BunnyStreamAPIKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch videos from Bunny.net",
				"details": err.Error(),
			})
			return
		}

		// Sync videos to database
		syncedCount := 0
		skippedCount := 0
		errorCount := 0
		var errors []string

		for _, bunnyVideo := range videos {
			err := syncVideoToDatabase(db, bunnyService, bunnyVideo)
			if err != nil {
				if strings.Contains(err.Error(), "already exists") {
					skippedCount++
				} else {
					errorCount++
					errors = append(errors, fmt.Sprintf("%s: %v", bunnyVideo.Title, err))
				}
				continue
			}
			syncedCount++
		}

		c.JSON(http.StatusOK, gin.H{
			"success":       true,
			"message":       "Sync completed",
			"total_videos":  len(videos),
			"synced":        syncedCount,
			"skipped":       skippedCount,
			"errors":        errorCount,
			"error_details": errors,
		})
	})

	// Test sync endpoint (no auth required for testing)
	v1.POST("/test/sync-bunny-videos", func(c *gin.Context) {
		// Fetch videos from Bunny.net
		videos, err := fetchBunnyVideos(cfg.BunnyStreamLibrary, cfg.BunnyStreamAPIKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch videos from Bunny.net",
				"details": err.Error(),
			})
			return
		}

		// Sync videos to database
		syncedCount := 0
		skippedCount := 0
		errorCount := 0
		var errors []string

		for _, bunnyVideo := range videos {
			err := syncVideoToDatabase(db, bunnyService, bunnyVideo)
			if err != nil {
				if strings.Contains(err.Error(), "already exists") {
					skippedCount++
				} else {
					errorCount++
					errors = append(errors, fmt.Sprintf("%s: %v", bunnyVideo.Title, err))
				}
				continue
			}
			syncedCount++
		}

		c.JSON(http.StatusOK, gin.H{
			"success":       true,
			"message":       "Test sync completed",
			"total_videos":  len(videos),
			"synced":        syncedCount,
			"skipped":       skippedCount,
			"errors":        errorCount,
			"error_details": errors,
		})
	})

	// Simple sync endpoint (no auth or webhook secret required for testing)
	v1.POST("/sync-bunny-videos", func(c *gin.Context) {
		// Validate configuration
		if cfg.BunnyStreamLibrary == "" || cfg.BunnyStreamAPIKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Bunny.net configuration missing",
				"details": gin.H{
					"library_id_set": cfg.BunnyStreamLibrary != "",
					"api_key_set":    cfg.BunnyStreamAPIKey != "",
				},
			})
			return
		}

		// Fetch videos from Bunny.net
		videos, err := fetchBunnyVideos(cfg.BunnyStreamLibrary, cfg.BunnyStreamAPIKey)
		if err != nil {
			// Categorize errors for better client handling
			var statusCode int
			var errorType string

			switch {
			case strings.Contains(err.Error(), "unauthorized"):
				statusCode = http.StatusUnauthorized
				errorType = "authentication_error"
			case strings.Contains(err.Error(), "forbidden"):
				statusCode = http.StatusForbidden
				errorType = "permission_error"
			case strings.Contains(err.Error(), "not found"):
				statusCode = http.StatusNotFound
				errorType = "resource_not_found"
			case strings.Contains(err.Error(), "rate limited"):
				statusCode = http.StatusTooManyRequests
				errorType = "rate_limit_error"
			case strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "connection"):
				statusCode = http.StatusServiceUnavailable
				errorType = "network_error"
			default:
				statusCode = http.StatusInternalServerError
				errorType = "api_error"
			}

			c.JSON(statusCode, gin.H{
				"error":      "Failed to fetch videos from Bunny.net",
				"error_type": errorType,
				"details":    err.Error(),
				"timestamp":  time.Now().Format(time.RFC3339),
			})
			return
		}

		// Sync videos to database with detailed tracking
		syncedCount := 0
		skippedCount := 0
		errorCount := 0
		var errors []gin.H
		var skipped []gin.H

		for i, bunnyVideo := range videos {
			err := syncVideoToDatabase(db, bunnyService, bunnyVideo)
			if err != nil {
				if strings.Contains(err.Error(), "already exists") {
					skippedCount++
					skipped = append(skipped, gin.H{
						"title":  bunnyVideo.Title,
						"guid":   bunnyVideo.GUID,
						"reason": "already_exists",
					})
				} else {
					errorCount++
					errors = append(errors, gin.H{
						"title": bunnyVideo.Title,
						"guid":  bunnyVideo.GUID,
						"error": err.Error(),
						"index": i,
					})
				}
				continue
			}
			syncedCount++
		}

		// Determine overall success status
		overallSuccess := errorCount == 0 || syncedCount > 0
		statusCode := http.StatusOK
		if errorCount > 0 && syncedCount == 0 {
			statusCode = http.StatusPartialContent
		}

		c.JSON(statusCode, gin.H{
			"success":         overallSuccess,
			"message":         fmt.Sprintf("Sync completed: %d synced, %d skipped, %d errors", syncedCount, skippedCount, errorCount),
			"total_videos":    len(videos),
			"synced":          syncedCount,
			"skipped":         skippedCount,
			"errors":          errorCount,
			"error_details":   errors,
			"skipped_details": skipped,
			"timestamp":       time.Now().Format(time.RFC3339),
		})
	})

	// Webhook endpoint for Bunny.net sync (no auth required)
	v1.POST("/webhook/bunny-sync", func(c *gin.Context) {
		// Simple webhook secret validation - allow if no secret is configured
		webhookSecret := c.GetHeader("X-Webhook-Secret")
		if cfg.BunnyWebhookSecret != "" && webhookSecret != cfg.BunnyWebhookSecret {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid webhook secret"})
			return
		}

		// Fetch videos from Bunny.net
		videos, err := fetchBunnyVideos(cfg.BunnyStreamLibrary, cfg.BunnyStreamAPIKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch videos from Bunny.net",
				"details": err.Error(),
			})
			return
		}

		// Sync videos to database
		syncedCount := 0
		skippedCount := 0
		errorCount := 0
		var errors []string

		for _, bunnyVideo := range videos {
			err := syncVideoToDatabase(db, bunnyService, bunnyVideo)
			if err != nil {
				if strings.Contains(err.Error(), "already exists") {
					skippedCount++
				} else {
					errorCount++
					errors = append(errors, fmt.Sprintf("%s: %v", bunnyVideo.Title, err))
				}
				continue
			}
			syncedCount++
		}

		c.JSON(http.StatusOK, gin.H{
			"success":       true,
			"message":       "Webhook sync completed",
			"total_videos":  len(videos),
			"synced":        syncedCount,
			"skipped":       skippedCount,
			"errors":        errorCount,
			"error_details": errors,
		})
	})
}

// Placeholder handler functions - these will be implemented in separate files
func handleRegister(db *database.DB, emailService *services.EmailService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Register endpoint - TODO"})
	}
}

func handleLogin(db *database.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Login endpoint - TODO"})
	}
}

func handleLogout() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Logout endpoint - TODO"})
	}
}

func handleRefreshToken(db *database.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Refresh token endpoint - TODO"})
	}
}

func handleForgotPassword(db *database.DB, emailService *services.EmailService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Forgot password endpoint - TODO"})
	}
}

func handleResetPassword(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Reset password endpoint - TODO"})
	}
}

func handleVerifyEmail(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Verify email endpoint - TODO"})
	}
}

func handleGetVideos(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Get videos endpoint - TODO"})
	}
}

func handleGetVideo(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Get video endpoint - TODO"})
	}
}

func handleStreamVideo(db *database.DB, bunnyService *services.BunnyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Stream video endpoint - TODO"})
	}
}

func handleGetCategories(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Get categories endpoint - TODO"})
	}
}

func handleSearchVideos(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Search videos endpoint - TODO"})
	}
}

func handleGetSubscriptionPlans(stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Get subscription plans endpoint - TODO"})
	}
}

func handleCreateSubscription(db *database.DB, stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Create subscription endpoint - TODO"})
	}
}

func handleStripeWebhook(db *database.DB, stripeService *services.StripeService, emailService *services.EmailService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Stripe webhook endpoint - TODO"})
	}
}

// GetProfileHandler handles retrieving user profile
func GetProfileHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		profile, err := db.GetUserProfile(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get profile"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"user": profile})
	}
}

// UpdateProfileHandler handles updating user profile
func UpdateProfileHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var updates map[string]interface{}
		if err := c.ShouldBindJSON(&updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// Validate required fields
		if firstName, exists := updates["first_name"]; exists {
			if firstNameStr, ok := firstName.(string); !ok || firstNameStr == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "First name is required"})
				return
			}
		}

		if email, exists := updates["email"]; exists {
			if emailStr, ok := email.(string); !ok || emailStr == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
				return
			}
		}

		// Update profile
		if err := db.UpdateUserProfile(userID, updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}

		// Get updated profile
		profile, err := db.GetUserProfile(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated profile"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Profile updated successfully",
			"user":    profile,
		})
	}
}

func handleGetUserActivity(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Get user activity endpoint - TODO"})
	}
}

func handleGetFavorites(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Get favorites endpoint - TODO"})
	}
}

func handleLikeVideo(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Like video endpoint - TODO"})
	}
}

func handleUnlikeVideo(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Unlike video endpoint - TODO"})
	}
}

func handleFavoriteVideo(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Favorite video endpoint - TODO"})
	}
}

func handleUnfavoriteVideo(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Unfavorite video endpoint - TODO"})
	}
}

func handleAddComment(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Add comment endpoint - TODO"})
	}
}

func handleGetComments(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Get comments endpoint - TODO"})
	}
}

func handleGetCurrentSubscription(db *database.DB, stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Get current subscription endpoint - TODO"})
	}
}

func handleCancelSubscription(db *database.DB, stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Cancel subscription endpoint - TODO"})
	}
}

func handleReactivateSubscription(db *database.DB, stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Reactivate subscription endpoint - TODO"})
	}
}

func handleGetBillingHistory(db *database.DB, stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Get pagination parameters
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 20
		}

		// Get user to find their Stripe customer ID
		user, err := db.GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			return
		}

		if !user.StripeCustomerID.Valid || user.StripeCustomerID.String == "" {
			// User has no Stripe customer ID, return empty billing history
			c.JSON(http.StatusOK, gin.H{
				"invoices": []interface{}{},
				"total":    0,
				"page":     page,
				"limit":    limit,
			})
			return
		}

		// Get starting after parameter for pagination
		startingAfter := c.Query("starting_after")

		// Get invoices from Stripe
		invoices, hasMore, err := stripeService.GetCustomerInvoices(user.StripeCustomerID.String, limit, startingAfter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get billing history"})
			return
		}

		// Calculate total (approximation since Stripe doesn't provide exact counts)
		total := len(invoices)
		if hasMore {
			total = page*limit + 1 // Indicate there are more items
		}

		c.JSON(http.StatusOK, gin.H{
			"invoices": invoices,
			"total":    total,
			"page":     page,
			"limit":    limit,
			"has_more": hasMore,
		})
	}
}

func handleGetInvoice(db *database.DB, stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		invoiceID := c.Param("id")
		if invoiceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invoice ID is required"})
			return
		}

		// Get user to verify ownership
		user, err := db.GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			return
		}

		if !user.StripeCustomerID.Valid || user.StripeCustomerID.String == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
			return
		}

		// Get invoice from Stripe
		invoice, err := stripeService.GetInvoice(invoiceID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"invoice": invoice})
	}
}

func handleDownloadInvoice(db *database.DB, stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		invoiceID := c.Param("id")
		if invoiceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invoice ID is required"})
			return
		}

		// Get user to verify ownership
		user, err := db.GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			return
		}

		if !user.StripeCustomerID.Valid || user.StripeCustomerID.String == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
			return
		}

		// Get invoice from Stripe
		invoice, err := stripeService.GetInvoice(invoiceID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
			return
		}

		if invoice.DownloadURL == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invoice download not available"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"downloadUrl": invoice.DownloadURL})
	}
}

// Admin analytics handlers
func handleAdminGetAnalytics(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Admin analytics endpoint - TODO"})
	}
}

func handleAdminGetUserAnalytics(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Admin user analytics endpoint - TODO"})
	}
}

func handleAdminGetVideoAnalytics(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Admin video analytics endpoint - TODO"})
	}
}

func handleAdminGetRevenueAnalytics(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Admin revenue analytics endpoint - TODO"})
	}
}

func handleAdminGetSystemHealth(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Admin system health endpoint - TODO"})
	}
}

func handleAdminCreateBackup(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Admin create backup endpoint - TODO"})
	}
}

func handleAdminGetLogs(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Admin get logs endpoint - TODO"})
	}
}

func handleGetRefunds(db *database.DB, stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Get pagination parameters
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		if limit < 1 || limit > 100 {
			limit = 20
		}

		// Get user to find their Stripe customer ID
		user, err := db.GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			return
		}

		if !user.StripeCustomerID.Valid || user.StripeCustomerID.String == "" {
			// User has no Stripe customer ID, return empty refunds
			c.JSON(http.StatusOK, gin.H{
				"refunds": []interface{}{},
				"total":   0,
			})
			return
		}

		// Get refunds from Stripe
		refunds, err := stripeService.ListCustomerRefunds(user.StripeCustomerID.String, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get refunds"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"refunds": refunds,
			"total":   len(refunds),
		})
	}
}

func handleGetRefund(db *database.DB, stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		refundID := c.Param("id")
		if refundID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Refund ID is required"})
			return
		}

		// Get user to verify ownership
		user, err := db.GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			return
		}

		if !user.StripeCustomerID.Valid || user.StripeCustomerID.String == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Refund not found"})
			return
		}

		// Get refund from Stripe
		refund, err := stripeService.GetRefund(refundID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Refund not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"refund": refund})
	}
}

func handleCreateRefund(db *database.DB, stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var refundReq struct {
			PaymentIntentID string `json:"payment_intent_id" binding:"required"`
			Amount          int64  `json:"amount"`
			Reason          string `json:"reason"`
		}

		if err := c.ShouldBindJSON(&refundReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate refund reason
		validReasons := []string{"duplicate", "fraudulent", "requested_by_customer"}
		if refundReq.Reason == "" {
			refundReq.Reason = "requested_by_customer"
		}

		reasonValid := false
		for _, validReason := range validReasons {
			if refundReq.Reason == validReason {
				reasonValid = true
				break
			}
		}

		if !reasonValid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid refund reason"})
			return
		}

		// Get user to verify ownership
		user, err := db.GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			return
		}

		if !user.StripeCustomerID.Valid || user.StripeCustomerID.String == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No payment methods found"})
			return
		}

		// Create refund through Stripe
		refund, err := stripeService.CreateRefund(refundReq.PaymentIntentID, refundReq.Amount, refundReq.Reason)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create refund"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"refund": refund})
	}
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

// HealthHandler returns the health status of the API
func HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"service":   "bome-backend",
			"version":   "1.0.0",
			"timestamp": time.Now(),
		})
	}
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
			userID,
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

// This StreamVideoHandler is now handled in video.go

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
}
