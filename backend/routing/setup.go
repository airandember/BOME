package routing

import (
	adminHandlers "bome-backend/admin/handlers"
	stripeHandlers "bome-backend/admin/handlers/stripe"
	authHandlers "bome-backend/authentication/handlers"
	authServices "bome-backend/authentication/services"
	creatorHandlers "bome-backend/creator/handlers"
	emailService "bome-backend/services/communication/email"
	subServices "bome-backend/subscription/services"
	videoHandlers "bome-backend/video-streaming/handlers"
	videoModels "bome-backend/video-streaming/models"
	videoServices "bome-backend/video-streaming/services"
	wsPackage "bome-backend/websocket"
	youtubeHandlers "bome-backend/youtube/handlers"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bome-backend/authentication/middleware"
	"bome-backend/infrastructure/config"
	"bome-backend/infrastructure/database"

	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// mapBunnyStatus maps Bunny.net video status codes to readable strings
func mapBunnyStatus(statusCode int) string {
	switch statusCode {
	case 0:
		return "queued"
	case 1:
		return "processing"
	case 2:
		return "encoding"
	case 3:
		return "finished"
	case 4:
		return "failed"
	case 5:
		return "ready"
	default:
		return "unknown"
	}
}

// GetVideosFromBunnyHandler returns videos from Bunny.net
func GetVideosFromBunnyHandler(db *database.DB, bunnyService *videoServices.BunnyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

		// TODO: Implement pagination in GetAllVideos or use GetVideoList
		// For now, return all videos from database
		allVideos, err := videoModels.GetAllVideos(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch videos",
				"code":  "DB_ERROR",
			})
			return
		}

		// Simple pagination
		start := (page - 1) * perPage
		end := start + perPage
		if start > len(allVideos) {
			start = len(allVideos)
		}
		if end > len(allVideos) {
			end = len(allVideos)
		}
		videos := allVideos[start:end]

		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"videos":   videos,
			"page":     page,
			"per_page": perPage,
			"total":    len(allVideos),
		})
	}
}

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

// WebSocketHub interface to avoid circular imports
type WebSocketHub interface {
	BroadcastSubscriberCreated(subscriber interface{})
	BroadcastSubscriberUpdated(subscriber interface{})
	BroadcastKPIUpdate(kpis interface{})
	BroadcastPaymentReceived(payment interface{})
	BroadcastPaymentFailed(payment interface{})
	BroadcastEvent(eventType string, data map[string]interface{}, message string)
}

// SetupRoutes configures all routes for the application
func SetupRoutes(
	router *gin.Engine,
	cfg *config.Config,
	db *database.DB,
	redis *database.Redis,
	bunnyService *videoServices.BunnyService,
	stripeService *subServices.StripeService,
	emailService interface{},
	wsHub WebSocketHub,
) {
	// Debug logging
	fmt.Printf("Setting up routes...\n")

	// Enhanced health check endpoints (with simplified checks for now)
	router.GET("/health", func(c *gin.Context) {
		status := "healthy"
		services := make(map[string]interface{})

		// Check database
		if db != nil && db.DB != nil {
			services["database"] = "healthy"
		} else {
			services["database"] = "unavailable"
			status = "degraded"
		}

		// Check Redis
		if db != nil && db.Redis != nil {
			services["redis"] = "healthy"
		} else {
			services["redis"] = "disabled"
		}

		// Check Stripe
		if stripeService != nil && stripeService.IsEnabled() {
			services["stripe"] = "healthy"
		} else {
			services["stripe"] = "disabled"
		}

		// Check Bunny
		if bunnyService != nil {
			services["bunny"] = "healthy"
		} else {
			services["bunny"] = "unavailable"
		}

		c.JSON(200, gin.H{
			"status":   status,
			"services": services,
			"version":  "1.0.0",
		})
	})
	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "alive"})
	})
	router.GET("/health/ready", func(c *gin.Context) {
		if db != nil && db.DB != nil {
			c.JSON(200, gin.H{"status": "ready"})
		} else {
			c.JSON(503, gin.H{"status": "not ready", "reason": "database unavailable"})
		}
	})
	fmt.Printf("✅ Registered health check endpoints\n")

	// API v1 routes
	v1 := router.Group("/api/v1")
	fmt.Printf("Created v1 route group with base path: %s\n", v1.BasePath())

	// Setup WebSocket routes (if hub is provided)
	if wsHub != nil {
		fmt.Printf("Setting up WebSocket routes...\n")
		// Cast wsHub to *wsPackage.AdminHub for route setup
		if hub, ok := wsHub.(*wsPackage.AdminHub); ok {
			wsPackage.SetupWebSocketRoutes(v1, hub)
			fmt.Printf("✅ WebSocket routes setup complete\n")
		} else {
			fmt.Printf("⚠️  WebSocket hub type assertion failed\n")
		}
	} else {
		fmt.Printf("⚠️  WebSocket hub not provided, real-time updates disabled\n")
	}

	// Initialize public Stripe service for frontend operations (needs to be accessible in both db and non-db scenarios)
	// var stripePublicService *subServices.StripePublicService
	if db != nil {
		// stripePublicService = subServices.NewStripePublicService(db)
	}

	// Admin routes - only setup if database is available
	admin := v1.Group("/admin")
	if db != nil {
		// 🎯 SUBSCRIBER ROUTES (Migrated with WebSocket support!)
		fmt.Printf("Setting up subscriber routes...\n")
		adminHandlers.SetupSubscriberRoutes(admin, db, wsHub)
		fmt.Printf("✅ Subscriber routes setup complete\n")

		// 🎯 SUBSCRIPTION PLANS ROUTES (Migrated with WebSocket support!)
		fmt.Printf("Setting up subscription plans routes...\n")
		adminHandlers.SetupSubscriptionPlanRoutes(admin, db, wsHub)
		fmt.Printf("✅ Subscription plans routes setup complete\n")

		// 🎯 SUBSCRIPTION OFFERS ROUTES (Migrated with WebSocket support!)
		fmt.Printf("Setting up subscription offers routes...\n")
		adminHandlers.SetupSubscriptionOfferRoutes(admin, db, wsHub)
		fmt.Printf("✅ Subscription offers routes setup complete\n")

		// 🔥 STRIPE DATABASE ROUTES (Read cached Stripe data)
		fmt.Printf("Setting up Stripe database routes...\n")
		stripeGroup := admin.Group("/streaming/stripe")
		stripeHandlers.SetupStripeDatabaseRoutes(stripeGroup, db)
		fmt.Printf("✅ Stripe database routes setup complete\n")

		// 🔄 STRIPE CUSTOMER SYNC ROUTES (Bidirectional sync with WebSocket)
		fmt.Printf("Setting up Stripe customer sync routes...\n")
		if stripeService != nil {
			stripeHandlers.SetupStripeCustomerSyncRoutes(stripeGroup, db, stripeService, wsHub)
			fmt.Printf("✅ Stripe customer sync routes setup complete\n")
		} else {
			fmt.Printf("⚠️  Stripe service not available, skipping customer sync routes\n")
		}

		// 🪝 STRIPE WEBHOOK ADMIN ROUTES (Monitor webhooks)
		fmt.Printf("Setting up Stripe webhook admin routes...\n")
		if stripeService != nil {
			stripeHandlers.SetupStripeWebhookAdminRoutes(stripeGroup, db, stripeService, wsHub)
			fmt.Printf("✅ Stripe webhook admin routes setup complete\n")
		} else {
			fmt.Printf("⚠️  Stripe service not available, skipping webhook admin routes\n")
		}

		// 📊 STRIPE ANALYTICS ROUTES (Metrics & KPIs)
		fmt.Printf("Setting up Stripe analytics routes...\n")
		if stripeService != nil {
			stripeHandlers.SetupStripeAnalyticsRoutes(stripeGroup, db, stripeService)
			fmt.Printf("✅ Stripe analytics routes setup complete\n")
		} else {
			fmt.Printf("⚠️  Stripe service not available, skipping analytics routes\n")
		}

		// 🔄 SIMPLE STRIPE SYNC ROUTES (Quick sync operations)
		fmt.Printf("Setting up simple Stripe sync routes...\n")
		if stripeService != nil {
			stripeHandlers.SetupSimpleStripeSyncRoutes(stripeGroup, db, stripeService)
			fmt.Printf("✅ Simple Stripe sync routes setup complete\n")
		} else {
			fmt.Printf("⚠️  Stripe service not available, skipping simple sync routes\n")
		}

		// 🔍 COMPREHENSIVE STRIPE SYNC ROUTES (Full sync with ghost detection)
		fmt.Printf("Setting up comprehensive Stripe sync routes...\n")
		if stripeService != nil {
			stripeHandlers.SetupComprehensiveSyncRoutes(stripeGroup, db, stripeService)
			fmt.Printf("✅ Comprehensive Stripe sync routes setup complete\n")
		} else {
			fmt.Printf("⚠️  Stripe service not available, skipping comprehensive sync routes\n")
		}

		// 👻 GHOST CUSTOMER ROUTES (Data quality management)
		fmt.Printf("Setting up ghost customer management routes...\n")
		stripeHandlers.SetupGhostCustomerRoutes(stripeGroup, db)
		fmt.Printf("✅ Ghost customer routes setup complete\n")

		// 📊 STREAMING ANALYTICS ROUTES (Platform-wide analytics)
		fmt.Printf("Setting up streaming analytics routes...\n")
		streamingGroup := admin.Group("/streaming")
		adminHandlers.SetupStreamingAnalyticsRoutes(streamingGroup, db)
		fmt.Printf("✅ Streaming analytics routes setup complete\n")

		// Admin routes - TEMPORARILY DISABLED
		// The admin routes from backend_original use database methods that don't exist yet
		// TODO: Create proper admin routes compatible with new backend structure
		fmt.Printf("⚠️  Admin routes temporarily disabled (missing database methods)\n")
		// adminHandlers.SetupAdminRoutes(admin, db)

		// Admin streaming routes - DISABLED until dependencies are migrated
		// TODO: Enable after migrating BusinessIntelligenceService, SubscriptionPlanStripeService, etc.
		// adminHandlers.SetupAdminStreamingRoutes(admin, db, stripeService, analyticsService, biService, subscriptionPlanStripeService, subscriptionOffersStripeService, bunnyService)

		// Database monitoring routes (for connection pool health)
		// RegisterDatabaseMonitoringRoutes(admin, db) // TODO: Implement

		// Create plan history service for analytics
		// planHistoryService := subServices.NewPlanHistoryService(db)
		// SetupAnalyticsRoutes(admin, db, planHistoryService) // TODO: Implement

		// Initialize subscription services
		// // subscriptionPlanService := subServices.New// subscriptionPlanService(db)
		// subscriptionPlanStripeService := subServices.NewSubscriptionPlanStripeService(...)     // Add Stripe-integrated service
		// subscriptionOffersStripeService := subServices.NewSubscriptionOffersStripeService(...) // Add Stripe-integrated offers service

		// Create admin cache service
		_ = subServices.NewSubscriptionAnalyticsService(db) // analyticsService - TODO: use this
		// SetupAdminStreamingRoutes(...) // TODO

		// 🎬 Master Video Routes
		fmt.Printf("Setting up master video routes...\n")
		bunnyService := videoServices.NewBunnyService()
		videoHandlers.SetupMasterVideoRoutes(admin, db, bunnyService)
		fmt.Printf("✅ Master video routes registered\n")

		// 🎬 YouTube Routes
		fmt.Printf("Setting up YouTube routes...\n")
		youtubeScheduler := youtubeHandlers.SetupYouTubeRoutes(v1, db)
		fmt.Printf("✅ YouTube routes registered (17 endpoints + scheduler started)\n")
		// Store scheduler reference if needed for graceful shutdown
		_ = youtubeScheduler

		// 💰 CREATOR PAYOUT ROUTES (Admin only)
		fmt.Printf("Setting up creator payout routes...\n")
		creatorHandlers.SetupPresenterRoutes(admin, db)
		creatorHandlers.SetupPayoutFormulaRoutes(admin, db)
		creatorHandlers.SetupPayoutRoutes(admin, db)
		fmt.Printf("✅ Creator payout routes registered (36 endpoints: 14 presenter + 7 formula + 15 payout)\n")

		// 🔗 PUBLIC STRIPE WEBHOOK ENDPOINT (NO AUTH REQUIRED)
		// This must be accessible to Stripe servers without authentication
		webhooks := v1.Group("/webhooks")
		if stripeService != nil {
			fmt.Printf("Setting up public Stripe webhook endpoint...\n")
			stripeHandlers.SetupStripeWebhookPublicRoutes(webhooks, db, stripeService, wsHub)
			fmt.Printf("✅ Public Stripe webhook registered: POST /api/v1/webhooks/stripe\n")
		} else {
			fmt.Printf("⚠️  Stripe service not available, skipping public webhook endpoint\n")
		}

		// 🧪 PUBLIC TEST ENDPOINT FOR SIMPLE STRIPE SYNC (NO AUTH REQUIRED)
		// This is for testing the new simple sync approach
		// simpleStripeSyncService test endpoint - TODO: Implement
		// fmt.Printf("✅ Public test endpoint registered: POST /api/v1/test/simple-stripe-sync\n")

		// Setup tag routes
		// SetupTagRoutes(...) // TODO: Implement

		// Setup advertisement routes - COMMENTED OUT until frontend ad system is ready
		// fmt.Printf("Setting up advertisement routes...\n")
		// fmt.Printf("Creating AdvertisementService...\n")
		// adService := subServices.NewAdvertisementService(db)
		// fmt.Printf("AdvertisementService created successfully\n")
		// fmt.Printf("Calling SetupAdvertisementRoutes...\n")
		// SetupAdvertisementRoutes(v1, adService)
		// fmt.Printf("Advertisement routes setup complete\n")

		// Initialize remaining subscription services
		// // subscriberService := subServices.New// subscriberService(...) // TODO
		// // subscriptionOffersService := subServices.New// subscriptionOffersService(...) // TODO
		// // subscriberHistoryService := subServices.New// subscriberHistoryService(...) // TODO

		// Setup subscription-related routes under admin group
		fmt.Printf("Setting up subscription plan routes...\n")
		// SetupSubscriptionPlanRoutes(...) // TODO: Implement
		fmt.Printf("Setting up subscription plan Stripe integration routes...\n")
		// Note: Stripe routes are now set up within SetupAdminStreamingRoutes
		fmt.Printf("Setting up subscription offers routes...\n")
		// SetupSubscriptionOfferRoutes(...) // TODO: Implement
		fmt.Printf("Setting up subscriber routes...\n")
		// SetupSubscriberRoutes(...) // TODO: Implement

		// Setup enhanced subscriber routes immediately after regular subscriber routes
		fmt.Printf("Setting up enhanced subscriber routes...\n")
		// SetupEnhancedSubscriberRoutes(...) // TODO: Implement
		fmt.Printf("Enhanced subscriber routes setup completed\n")

		fmt.Printf("Setting up subscriber history routes...\n")
		// SetupSubscriberHistoryRoutes(...) // TODO
		// SetupSubscriptionRoutes(...) // TODO

		// Setup email usage routes
		fmt.Printf("Setting up email usage routes...\n")
		// SetupEmailUsageRoutes(...) // TODO

		fmt.Printf("Database-dependent admin routes setup complete\n")
		fmt.Printf("Subscription services setup complete\n")
	} else {
		// Setup fallback admin routes that return service unavailable
		admin.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "degraded",
				"error":  "Database unavailable",
			})
		})

		// Streaming dashboard fallback
		admin.GET("/streaming/dashboard", func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Database unavailable - streaming dashboard temporarily offline",
			})
		})

		// Stripe routes fallback
		streaming := admin.Group("/streaming")
		stripe := streaming.Group("/stripe")
		{
			stripe.GET("/summary", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "Database unavailable - Stripe integration temporarily offline",
				})
			})
			stripe.POST("/secret", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "Database unavailable - cannot save Stripe configuration",
				})
			})
			stripe.GET("/portal-link", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "Database unavailable - portal link temporarily unavailable",
				})
			})
			stripe.POST("/portal-link", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "Database unavailable - cannot save portal link",
				})
			})
			stripe.DELETE("/portal-link", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "Database unavailable - cannot clear portal link",
				})
			})
		}

		// Subscribers fallback
		admin.GET("/subscribers/", func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Database unavailable - subscriber data temporarily offline",
			})
		})

		fmt.Printf("Database unavailable - admin routes setup with fallbacks\n")
		fmt.Printf("Skipping subscription services (database unavailable)\n")
	}

	// Public subscription plan routes - only if database is available
	if db != nil {
		// // subscriptionPlanService := subServices.New// subscriptionPlanService(db)
		// // subscriptionOffersService := subServices.New// subscriptionOffersService(...) // TODO

		publicPlans := v1.Group("/subscription-plans")
		{
			// Get all subscription data (plans + offers) - MUST come before /:id
			publicPlans.GET("/all", func(c *gin.Context) {
				// getAllSubscriptionData(...) // TODO
			})

			// Get active subscription plans
			publicPlans.GET("/active", func(c *gin.Context) {
				// getActiveSubscriptionPlans(...) // TODO
			})

			// Get promoted subscription plans
			publicPlans.GET("/promoted", func(c *gin.Context) {
				// getPromotedSubscriptionPlans(c, subscriptionPlanService) // TODO
				c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented"})
			})

			// Get subscription plan by ID (public) - MUST come last
			publicPlans.GET("/:id", func(c *gin.Context) {
				// getSubscriptionPlanPublic(c, subscriptionPlanService) // TODO
				c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented"})
			})
		}

		// Setup Stripe routes using dedicated file
		// SetupPublicStripeRoutes(v1, stripePublicService) // TODO: Implement
		// SetupAuthenticatedStripeRoutes(v1, stripePublicService, globalUserSubscriptionService) // TODO: Implement
	} else {
		// Provide fallback responses when database is unavailable
		publicPlans := v1.Group("/subscription-plans")
		{
			publicPlans.GET("/all", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "Service temporarily unavailable",
				})
			})
			publicPlans.GET("/active", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "Service temporarily unavailable",
				})
			})
			publicPlans.GET("/promoted", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "Service temporarily unavailable",
				})
			})
			publicPlans.GET("/:id", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "Service temporarily unavailable",
				})
			})
		}

		// Fallback public Stripe routes when database is unavailable
		publicStripe := v1.Group("/stripe")
		publicStripe.Use(middleware.AuthRequired())
		{
			publicStripe.GET("/portal-link", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "Service temporarily unavailable",
				})
			})
			publicStripe.POST("/checkout-session", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "Service temporarily unavailable",
				})
			})
			publicStripe.GET("/config", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "Service temporarily unavailable",
				})
			})
			publicStripe.GET("/session/:session_id", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "Service temporarily unavailable",
				})
			})
			publicStripe.POST("/create-subscription", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "Service temporarily unavailable",
				})
			})
			publicStripe.GET("/subscription-status", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "Service temporarily unavailable",
				})
			})
			publicStripe.POST("/cancel-subscription", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "Service temporarily unavailable",
				})
			})
		}
	}

	fmt.Printf("Admin routes setup complete\n")

	// Only setup YouTube routes if database is available
	if db != nil {
		// SetupYouTubeRoutes(v1, db)
		fmt.Printf("YouTube routes setup complete\n")

		// Search Index routes are now set up within SetupAdminStreamingRoutes
		fmt.Printf("Search Index routes setup complete\n")
	} else {
		fmt.Printf("Skipping YouTube and Search Index routes (database unavailable)\n")
	}

	// Real authentication routes
	fmt.Printf("🔐 Setting up authentication routes...\n")
	auth := v1.Group("/auth")
	{
		auth.POST("/login", authHandlers.LoginHandler(db, nil))
		auth.POST("/register", authHandlers.RegisterHandler(db, nil))
		auth.POST("/logout", authHandlers.LogoutHandler(db))
		auth.POST("/change-password", middleware.AuthRequired(), authHandlers.ChangePasswordHandler(db))
		auth.POST("/setup-password", authHandlers.SetupPasswordHandler(db))
		fmt.Printf("✅ Auth route registered: POST /api/v1/auth/setup-password\n")

		// Email verification routes
		fmt.Printf("🔄 Registering verify-email-link route...\n")
		handler := authHandlers.VerifyEmailLinkHandler(db)
		if handler == nil {
			fmt.Printf("❌ VerifyEmailLinkHandler returned nil!\n")
		} else {
			auth.GET("/verify-email-link", handler) // GET route for email links
			fmt.Printf("✅ Registered verify-email-link route\n")
		}
		auth.POST("/verify-email", authHandlers.VerifyEmailHandler(db))       // POST route for API/mobile
		auth.GET("/verify-email/:token", authHandlers.VerifyEmailHandler(db)) // Legacy route
		auth.POST("/resend-verification", authHandlers.ResendVerificationHandler(db, nil))
		auth.POST("/request-verification", authHandlers.RequestVerificationHandler(db, nil)) // For existing users

		// Password reset routes
		auth.POST("/forgot-password", authHandlers.ForgotPasswordHandler(db, nil))
		auth.POST("/reset-password", authHandlers.ResetPasswordHandler(db))

		// Token refresh
		auth.POST("/refresh", authHandlers.RefreshTokenHandler(db))
	}
	fmt.Printf("✅ Authentication routes setup complete\n")

	// Initialize OAuth2 service
	oauth2Service := authServices.NewOAuth2Service(db)

	// Setup OAuth2 routes
	fmt.Printf("Setting up OAuth2 routes...\n")
	authHandlers.SetupOAuth2Routes(v1, db, oauth2Service)
	fmt.Printf("✅ OAuth2 routes setup complete\n")

	// User profile routes (require email verification)
	users := v1.Group("/users")
	{
		users.GET("/me", middleware.AuthRequired(), authHandlers.GetCurrentUserHandler(db))
		users.PUT("/me", middleware.AuthRequired(), authHandlers.UpdateCurrentUserHandler(db))
		users.GET("/profile", middleware.AuthRequired(), authHandlers.GetCurrentUserHandler(db))    // Alias for /me
		users.PUT("/profile", middleware.AuthRequired(), authHandlers.UpdateCurrentUserHandler(db)) // Alias for /me
	}

	// Video routes using database handlers with bunny.net integration
	videos := v1.Group("/videos")
	{
		fmt.Printf("Setting up video routes...\n")

		// Get all videos with pagination and filtering
		videos.GET("", middleware.AuthRequired(), middleware.SubscriptionValidation(db), func(c *gin.Context) {
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
			videos, err := videoModels.GetVideos(db, limit, offset, category, "")
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
				// var playData *subServices.VideoPlayData
				if video.BunnyVideoID != "" {
					_, _ = bunnyService.GetVideoPlayData(video.BunnyVideoID)
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
				// TODO: Re-implement playData logic

				responseVideos = append(responseVideos, responseVideo)
			}

			// Get total count for pagination
			totalCount, err := videoModels.GetVideoCount(db)
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

		// videos.GET("/categories", GetMockCategoriesHandler) // Must come before /:id // TODO: Implement
		videos.GET("/:id", middleware.AuthRequired(), middleware.SessionActivityTracker(db), middleware.SubscriptionValidation(db), func(c *gin.Context) {
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
				video, err := videoModels.GetVideoByID(db, videoIDInt)
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
					} else {
						// TODO: Re-implement playData logic
						var playDataMap map[string]interface{}
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
			video, err := videoModels.GetVideoByBunnyID(db, videoID)
			if err == nil {
				// Found in database, get fresh play data
				playData, err := bunnyService.GetVideoPlayData(videoID)
				if err != nil {
					fmt.Printf("Failed to get play data: %v\n", err)
					// Continue without play data
				}

				// TODO: Re-implement playData logic
				var playDataMap map[string]interface{}
				playDataBytes, err := json.Marshal(playData)
				if err == nil {
					json.Unmarshal(playDataBytes, &playDataMap)
					video.PlayData = playDataMap
				}
				video.IframeSrc = playData.IframeSrc
				video.DirectPlayURL = playData.DirectPlayURL
				video.PlaybackURL = playData.DirectPlayURL
				video.Resolutions = playData.ResolutionOptions

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
			_, err = bunnyService.GetVideoPlayData(videoID) // playData - TODO: use this
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

			// TODO: Re-implement playData logic

			c.JSON(http.StatusOK, response)
		})

		// TODO: Implement comments endpoint in Communication braid
		// videos.GET("/:id/comments", CommentsHandler)

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
		videos.GET("/:id/blob", middleware.AuthRequired(), middleware.SubscriptionValidation(db), func(c *gin.Context) {
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
		// TODO: Re-implement optimized Bunny service metrics
		// if optimizedService := subServices.GetGlobalOptimizedBunnyService(); optimizedService != nil {
		// 	metrics := optimizedService.GetMetrics()
		// 	c.JSON(http.StatusOK, gin.H{
		// 		"success":   true,
		// 		"metrics":   metrics,
		// 		"timestamp": time.Now().Format(time.RFC3339),
		// 	})
		// } else {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"metrics": gin.H{
				"message":      "Performance metrics coming soon",
				"service_type": "standard",
			},
			"timestamp": time.Now().Format(time.RFC3339),
		})
		// }
	})

	// Bunny.net direct access endpoint (separate from videos to avoid conflicts)
	v1.GET("/bunny-videos", middleware.AuthRequired(), middleware.SubscriptionValidation(db), GetVideosFromBunnyHandler(db, bunnyService))

	// Add single video endpoint
	v1.GET("/bunny-videos/:id", middleware.AuthRequired(), middleware.SubscriptionValidation(db), func(c *gin.Context) {
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
		dbVideo, err := videoModels.GetVideoByBunnyID(db, videoID)
		if err != nil {
			// Video doesn't exist, create it
			dbVideo, err = videoModels.CreateVideo(db,
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
				err = videoModels.UpdateVideo(db, dbVideo.ID, updates)
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
		dbVideo, err := videoModels.GetVideoByBunnyID(db, videoID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found in database"})
			return
		}

		// Update video in database
		if err := videoModels.UpdateVideo(db, dbVideo.ID, updateData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update video"})
			return
		}

		// Log admin action
		// userID := c.GetInt("user_id")
		// TODO: Re-implement admin logging
		// go db.CreateAdminLog(&userID, "bunny_video_updated", "video", &dbVideo.ID, updateData, c.ClientIP(), c.GetHeader("User-Agent"))

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
		dbVideo, err := videoModels.GetVideoByBunnyID(db, videoID)
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
		if err := videoModels.DeleteVideo(db, dbVideo.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete video from database"})
			return
		}

		// Log admin action
		// userID := c.GetInt("user_id")
		// TODO: Re-implement admin logging
		// go db.CreateAdminLog(&userID, "bunny_video_deleted", "video", &dbVideo.ID, map[string]interface{}{"title": dbVideo.Title}, c.ClientIP(), c.GetHeader("User-Agent"))

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
func UploadVideoHandler(db *database.DB, bunnyService *videoServices.BunnyService) gin.HandlerFunc {
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
		video, err := videoModels.CreateVideo(db,
			title,
			description,
			uploadResp.VideoID,
			"", // thumbnail URL will be set later
			category,
			0, // duration will be updated when processing is complete
			header.Size,
			tags,
			userID, // createdBy - use actual user ID
			true,   // vid_status
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save video metadata"})
			return
		}

		// Log the upload action
		// TODO: Re-implement admin logging
		// go db.CreateAdminLog(&userID, "video_uploaded", "video", &video.ID, map[string]interface{}{
		// 	"title":     video.Title,
		// 	"bunny_id":  video.BunnyVideoID,
		// 	"file_size": header.Size,
		// }, c.ClientIP(), c.GetHeader("User-Agent"))

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

// Global services for subscription creation
var (
	// globalUserSubscriptionService *subServices.UserSubscriptionService
	globalEmailService *emailService.EmailService
)

// InitializeSubscriptionServices initializes the global subscription services
func InitializeSubscriptionServices(db *database.DB) {
	globalEmailService = emailService.NewEmailService(db)
	// globalUserSubscriptionService = subServices.NewUserSubscriptionService(db, globalEmailService)
	log.Printf("✅ [SERVICES] Subscription and email services initialized")
}

// createSubscriptionFromSession helper function for creating subscriptions (stubbed)
func createSubscriptionFromSession(userID int, sessionData map[string]interface{}) (interface{}, error) {
	// TODO: Re-implement
	return nil, fmt.Errorf("not implemented")
}

// VerifyEmailHandler handles email verification
func VerifyEmailHandler(emailService *emailService.EmailService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
			return
		}

		if emailService == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Email service not available"})
			return
		}

		err := emailService.VerifyEmail(token)
		if err != nil {
			log.Printf("❌ [EMAIL-VERIFY] Failed to verify email: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Printf("✅ [EMAIL-VERIFY] Email verified successfully")
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Email verified successfully",
		})
	}
}
