package routes

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

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

	// Video routes using existing handlers
	videos := v1.Group("/videos")
	{
		videos.GET("", GetMockVideosHandler)
		videos.GET("/:id", GetMockVideoHandler)
		videos.GET("/:id/comments", GetMockCommentsHandler)
		videos.GET("/categories", GetMockCategoriesHandler)
	}

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
