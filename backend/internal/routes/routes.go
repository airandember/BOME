package routes

import (
	"net/http"
	"strconv"

	"bome-backend/internal/config"
	"bome-backend/internal/database"
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

	// Setup all mock data routes for development/testing
	SetupMockDataRoutes(v1)
	SetupArticlesRoutes(v1)
	SetupRolesRoutes(v1)
	SetupYouTubeRoutes(v1, db)

	{
		// Mock authentication routes (for development)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", func(c *gin.Context) {
				// Mock login that always returns success with admin token
				c.JSON(http.StatusOK, gin.H{
					"token": "mock-jwt-token-12345",
					"user": gin.H{
						"id":        1,
						"email":     "admin@bome.com",
						"role":      "super_admin",
						"full_name": "System Administrator",
					},
				})
			})
			auth.POST("/register", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Registration successful"})
			})
			auth.POST("/logout", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
			})
		}

		// Video routes using existing handlers
		videos := v1.Group("/videos")
		{
			videos.GET("", GetMockVideosHandler)
			videos.GET("/:id", GetMockVideoHandler)
			videos.GET("/:id/comments", GetMockCommentsHandler)
			videos.GET("/categories", GetMockCategoriesHandler)
		}

		// Admin routes using existing mock handlers
		admin := v1.Group("/admin")
		{
			admin.GET("/analytics", GetAdminAnalyticsHandler)
			admin.GET("/users", GetAdminUsersHandler)
			admin.GET("/videos", GetAdminVideosHandler(db))
			admin.GET("/advertisers", GetAdvertiserAccountsHandler)
			admin.GET("/advertisers/:id", GetAdvertiserAccountHandler)
			admin.GET("/campaigns", GetAdCampaignsHandler)
			admin.GET("/campaigns/:id", GetAdCampaignHandler)
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
