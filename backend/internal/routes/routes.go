package routes

import (
	"net/http"
	"strconv"

	"bome-backend/internal/config"
	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

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
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "bome-streaming-backend",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		public := v1.Group("")
		{
			// Authentication routes
			auth := public.Group("/auth")
			{
				auth.POST("/register", RegisterHandler(db))
				auth.POST("/login", LoginHandler(db))
				auth.POST("/logout", handleLogout())
				auth.POST("/refresh", handleRefreshToken(db, cfg))
				auth.POST("/forgot-password", ForgotPasswordHandler(db, emailService))
				auth.POST("/reset-password", ResetPasswordHandler(db))
				auth.POST("/verify-email", VerifyEmailHandler(db))
			}

			// Video routes (public)
			videos := public.Group("/videos")
			{
				videos.GET("", GetVideosHandler(db))
				videos.GET("/:id", GetVideoHandler(db))
				videos.GET("/:id/stream", StreamVideoHandler(db, bunnyService))
				videos.GET("/categories", GetCategoriesHandler(db))
				videos.GET("/search", SearchVideosHandler(db))
			}

			// Subscription routes
			subscriptions := public.Group("/subscriptions")
			{
				subscriptions.GET("/plans", handleGetSubscriptionPlans(stripeService))
				subscriptions.POST("/create", handleCreateSubscription(db, stripeService))
			}

			// Stripe webhook
			public.POST("/webhooks/stripe", handleStripeWebhook(db, stripeService, emailService))
		}

		// Protected routes (require authentication)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("/profile", handleGetProfile(db))
				users.PUT("/profile", handleUpdateProfile(db))
				users.GET("/activity", handleGetUserActivity(db))
				users.GET("/favorites", GetFavoritesHandler(db))
			}

			// Video interaction routes
			videoInteractions := protected.Group("/videos")
			{
				videoInteractions.POST("/:id/like", LikeVideoHandler(db))
				videoInteractions.DELETE("/:id/like", UnlikeVideoHandler(db))
				videoInteractions.POST("/:id/favorite", FavoriteVideoHandler(db))
				videoInteractions.DELETE("/:id/favorite", UnfavoriteVideoHandler(db))
				videoInteractions.POST("/:id/comment", AddCommentHandler(db))
				videoInteractions.GET("/:id/comments", GetCommentsHandler(db))
			}

			// Subscription management
			subscriptions := protected.Group("/subscriptions")
			{
				subscriptions.GET("", GetSubscriptionHandler(db))
				subscriptions.POST("", CreateSubscriptionHandler(db))
				subscriptions.DELETE("", CancelSubscriptionHandler(db))
			}
		}

		// Admin routes (require admin authentication)
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
		{
			// User management
			adminUsers := admin.Group("/users")
			{
				adminUsers.GET("", GetUsersHandler(db))
				adminUsers.GET("/:id", GetUserHandler(db))
				adminUsers.PUT("/:id", UpdateUserHandler(db))
				adminUsers.DELETE("/:id", DeleteUserHandler(db))
			}

			// Video management
			adminVideos := admin.Group("/videos")
			{
				adminVideos.GET("", GetAdminVideosHandler(db))
				adminVideos.POST("", handleAdminUploadVideo(db, bunnyService))
				adminVideos.PUT("/:id", UpdateVideoHandler(db))
				adminVideos.DELETE("/:id", DeleteVideoHandler(db))
				adminVideos.GET("/pending", handleAdminGetPendingVideos(db))
				adminVideos.POST("/:id/approve", handleAdminApproveVideo(db))
				adminVideos.POST("/:id/reject", handleAdminRejectVideo(db))
				adminVideos.POST("/bulk", BulkVideoOperationHandler(db))
				adminVideos.GET("/stats", GetVideoStatsHandler(db))
				adminVideos.GET("/categories", GetVideoCategoriesHandler(db))
			}

			// Analytics
			adminAnalytics := admin.Group("/analytics")
			{
				adminAnalytics.GET("/overview", GetAnalyticsHandler(db))
				adminAnalytics.GET("/detailed/:metric", GetDetailedAnalyticsHandler(db))
				adminAnalytics.GET("/realtime", GetRealTimeAnalyticsHandler(db))
				adminAnalytics.GET("/export", ExportAnalyticsHandler(db))
			}

			// System management
			adminSystem := admin.Group("/system")
			{
				adminSystem.GET("/health", GetSystemHealthHandler(db))
				adminSystem.POST("/backup", handleAdminCreateBackup(db, spacesService))
				adminSystem.GET("/logs", handleAdminGetLogs())
			}
		}
	}
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

func handleGetProfile(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Get profile endpoint - TODO"})
	}
}

func handleUpdateProfile(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Update profile endpoint - TODO"})
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
		c.JSON(http.StatusOK, gin.H{"message": "Get billing history endpoint - TODO"})
	}
}

// Admin handlers
func handleAdminGetUsers(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Admin get users endpoint - TODO"})
	}
}

func handleAdminGetUser(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Admin get user endpoint - TODO"})
	}
}

func handleAdminUpdateUser(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Admin update user endpoint - TODO"})
	}
}

func handleAdminDeleteUser(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Admin delete user endpoint - TODO"})
	}
}

func handleAdminUploadVideo(db *database.DB, bunnyService *services.BunnyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse multipart form
		file, err := c.FormFile("video")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No video file provided"})
			return
		}

		title := c.PostForm("title")
		description := c.PostForm("description")
		category := c.PostForm("category")

		if title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
			return
		}

		adminID := c.GetInt("user_id")

		// Upload to Bunny.net
		uploadResp, err := bunnyService.UploadVideo(file, title, description)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload video"})
			return
		}

		// Create video record in database
		video, err := db.CreateVideo(title, description, uploadResp.VideoID, "", category, 0, file.Size, []string{}, adminID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create video record"})
			return
		}

		// Log admin action
		go db.CreateAdminLog(&adminID, "video_uploaded", "video", &video.ID, map[string]interface{}{"title": title}, c.ClientIP(), c.GetHeader("User-Agent"))

		c.JSON(http.StatusOK, gin.H{
			"message": "Video uploaded successfully",
			"video":   video,
		})
	}
}

func handleAdminUpdateVideo(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Admin update video endpoint - TODO"})
	}
}

func handleAdminDeleteVideo(db *database.DB, bunnyService *services.BunnyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Admin delete video endpoint - TODO"})
	}
}

func handleAdminGetPendingVideos(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Mock pending videos for development mode
		if db == nil {
			pendingVideos := []map[string]interface{}{
				{
					"id":          2,
					"title":       "DNA and the Book of Mormon",
					"description": "Scientific perspectives on DNA evidence and Book of Mormon populations",
					"duration":    "22:15",
					"thumbnail":   "https://example.com/thumb2.jpg",
					"status":      "pending",
					"category":    "Science",
					"uploaded_by": map[string]interface{}{
						"id":    3,
						"name":  "Dr. Sarah Johnson",
						"email": "sarah.johnson@byu.edu",
					},
					"upload_date": "2024-01-18T14:20:00Z",
					"file_size":   "298.4 MB",
					"resolution":  "1080p",
					"tags":        []string{"dna", "science", "genetics"},
				},
			}

			c.JSON(http.StatusOK, gin.H{"videos": pendingVideos})
			return
		}

		// Real database implementation
		videos, err := db.GetVideos(100, 0, "", "pending")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pending videos"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"videos": videos})
	}
}

func handleAdminApproveVideo(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
			return
		}

		adminID := c.GetInt("user_id")

		// Update video status to published
		if err := db.UpdateVideoStatus(videoID, "published"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve video"})
			return
		}

		// Log admin action
		go db.CreateAdminLog(&adminID, "video_approved", "video", &videoID, nil, c.ClientIP(), c.GetHeader("User-Agent"))

		c.JSON(http.StatusOK, gin.H{"message": "Video approved successfully"})
	}
}

func handleAdminRejectVideo(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
			return
		}

		adminID := c.GetInt("user_id")

		// Update video status to rejected
		if err := db.UpdateVideoStatus(videoID, "rejected"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject video"})
			return
		}

		// Log admin action
		go db.CreateAdminLog(&adminID, "video_rejected", "video", &videoID, nil, c.ClientIP(), c.GetHeader("User-Agent"))

		c.JSON(http.StatusOK, gin.H{"message": "Video rejected successfully"})
	}
}

func handleAdminGetAnalytics(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Admin get analytics endpoint - TODO"})
	}
}

func handleAdminGetUserAnalytics(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Admin get user analytics endpoint - TODO"})
	}
}

func handleAdminGetVideoAnalytics(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Admin get video analytics endpoint - TODO"})
	}
}

func handleAdminGetRevenueAnalytics(db *database.DB, stripeService *services.StripeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Admin get revenue analytics endpoint - TODO"})
	}
}

func handleAdminGetSystemHealth(db *database.DB, redis *database.Redis) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Admin get system health endpoint - TODO"})
	}
}

func handleAdminCreateBackup(db *database.DB, spacesService *services.SpacesService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Admin create backup endpoint - TODO"})
	}
}

func handleAdminGetLogs() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Admin get logs endpoint - TODO"})
	}
}
