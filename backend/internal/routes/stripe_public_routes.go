package routes

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// Note: globalUserSubscriptionService is DEPRECATED in Phase 7

// SetupPublicStripeRoutes sets up public Stripe routes that don't require authentication
func SetupPublicStripeRoutes(v1 *gin.RouterGroup, stripePublicService *services.StripePublicService) {
	if stripePublicService == nil {
		log.Printf("⚠️ [STRIPE-PUBLIC] StripePublicService is nil, setting up fallback routes")
		setupFallbackStripeRoutes(v1)
		return
	}

	log.Printf("🔧 [STRIPE-PUBLIC] Setting up public Stripe routes...")

	// Public Stripe routes (no authentication required)
	publicStripe := v1.Group("/stripe")
	{
		// Get public Stripe configuration (publishable key) - NO AUTH REQUIRED
		publicStripe.GET("/config", func(c *gin.Context) {
			config, err := stripePublicService.GetStripeConfig()
			if err != nil {
				log.Printf("❌ [CONFIG] Failed to get Stripe config: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Stripe configuration"})
				return
			}

			log.Printf("✅ [CONFIG] Stripe config retrieved successfully")
			c.JSON(http.StatusOK, config)
		})
	}

	log.Printf("✅ [STRIPE-PUBLIC] Public Stripe routes setup complete")
}

// SetupAuthenticatedStripeRoutes sets up Stripe routes that require authentication
// DEPRECATED: Phase 7 replaces this with user-controlled subscription management
func SetupAuthenticatedStripeRoutes(v1 *gin.RouterGroup, db *database.DB, stripePublicService *services.StripePublicService) {
	if stripePublicService == nil {
		log.Printf("⚠️ [STRIPE-AUTH] StripePublicService is nil, skipping authenticated routes")
		return
	}

	log.Printf("🔧 [STRIPE-AUTH] Setting up authenticated Stripe routes...")

	// Authenticated Stripe routes
	authenticatedStripe := v1.Group("/stripe")
	authenticatedStripe.Use(middleware.AuthRequired()) // Require authentication but not admin
	{
		// Get customer portal link for authenticated users
		authenticatedStripe.GET("/portal-link", func(c *gin.Context) {
			// Get user from context (set by AuthRequired middleware)
			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
				return
			}

			log.Printf("🔍 [PORTAL] Creating portal link for user %v", userID)

			// TODO: Get actual customer ID from user
			customerID := fmt.Sprintf("cus_%v", userID)
			returnURL := c.Query("return_url")
			if returnURL == "" {
				// Use environment variable for base URL
				baseURL := os.Getenv("PUBLIC_APP_URL")
				if baseURL == "" {
					baseURL = "https://watch.bookofmormonevidence.org" // production fallback
				}
				returnURL = baseURL + "/dashboard"
			}

			portalURL, err := stripePublicService.GetCustomerPortalURL(customerID, returnURL)
			if err != nil {
				log.Printf("❌ [PORTAL] Failed to create portal link: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create portal link", "details": err.Error()})
				return
			}

			log.Printf("✅ [PORTAL] Portal link created successfully")
			c.JSON(http.StatusOK, gin.H{"portal_url": portalURL})
		})

		// Create embedded checkout session for authenticated users
		authenticatedStripe.POST("/checkout-session", func(c *gin.Context) {
			var req struct {
				PlanID    string `json:"plan_id" binding:"required"`
				ReturnURL string `json:"return_url" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
				return
			}

			// Get user from context (set by AuthRequired middleware)
			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
				return
			}

			userIDInt, ok := userID.(int)
			if !ok {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
				return
			}

			log.Printf("🔍 [CHECKOUT] Creating embedded checkout session for user %d, plan %s", userIDInt, req.PlanID)

			// 🔒 CHECK: Prevent users with active subscriptions from creating new ones
			linkingService := services.NewCustomerLinkingService(db)
			userSubService := services.NewUserSubscriptionService(db, linkingService)
			canSubscribe, message, err := userSubService.CanUserSubscribe(userIDInt)
			if err != nil {
				log.Printf("❌ [CHECKOUT] Failed to check subscription eligibility: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify subscription status"})
				return
			}

			if !canSubscribe {
				log.Printf("🚫 [CHECKOUT] User %d blocked from creating new subscription: %s", userIDInt, message)

				// Get support email from public settings
				var supportEmail string
				err = db.DB.QueryRow("SELECT value FROM public_settings WHERE key = 'support_email'").Scan(&supportEmail)
				if err != nil || supportEmail == "" {
					supportEmail = "support@bookofmormonevidence.org" // Fallback
				}

				c.JSON(http.StatusConflict, gin.H{
					"error":         "Cannot create new subscription",
					"message":       "You already have an active subscription",
					"support_email": supportEmail,
					"action":        "redirect_dashboard",
				})
				return
			}

			// Create embedded checkout session using public service
			clientSecret, err := stripePublicService.CreateEmbeddedCheckoutSession(req.PlanID, req.ReturnURL, fmt.Sprintf("%v", userIDInt))
			if err != nil {
				log.Printf("❌ [CHECKOUT] Failed to create checkout session: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create checkout session", "details": err.Error()})
				return
			}

			log.Printf("✅ [CHECKOUT] Embedded checkout session created successfully for user %d", userIDInt)
			c.JSON(http.StatusOK, gin.H{
				"client_secret": clientSecret,
			})
		})

		// 🔐 SECRET PROMO: Create checkout session for hidden promo plans
		// This endpoint is specifically for the /secretsub/[code] page
		// Key differences from /checkout-session:
		// - Does NOT require plan.is_active = true (allows hidden promos)
		// - Uses stripe_price_id directly from DB (exact price, not first active)
		authenticatedStripe.POST("/secret-checkout-session", func(c *gin.Context) {
			var req struct {
				PlanID    string `json:"plan_id" binding:"required"`
				ReturnURL string `json:"return_url" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
				return
			}

			// Get user from context (set by AuthRequired middleware)
			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
				return
			}

			userIDInt, ok := userID.(int)
			if !ok {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
				return
			}

			log.Printf("🔐 [SECRET-PROMO] Creating checkout for user %d, plan %s", userIDInt, req.PlanID)

			// 🔒 CHECK: Prevent users with active subscriptions from creating new ones
			linkingService := services.NewCustomerLinkingService(db)
			userSubService := services.NewUserSubscriptionService(db, linkingService)
			canSubscribe, message, err := userSubService.CanUserSubscribe(userIDInt)
			if err != nil {
				log.Printf("❌ [SECRET-PROMO] Failed to check subscription eligibility: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify subscription status"})
				return
			}

			if !canSubscribe {
				log.Printf("🚫 [SECRET-PROMO] User %d blocked: %s", userIDInt, message)

				var supportEmail string
				err = db.DB.QueryRow("SELECT value FROM public_settings WHERE key = 'support_email'").Scan(&supportEmail)
				if err != nil || supportEmail == "" {
					supportEmail = "support@bookofmormonevidence.org"
				}

				c.JSON(http.StatusConflict, gin.H{
					"error":         "Cannot create new subscription",
					"message":       "You already have an active subscription",
					"support_email": supportEmail,
					"action":        "redirect_dashboard",
				})
				return
			}

			// 🔐 Use the SECRET PROMO checkout method
			clientSecret, err := stripePublicService.CreateSecretPromoCheckoutSession(req.PlanID, req.ReturnURL, fmt.Sprintf("%v", userIDInt))
			if err != nil {
				log.Printf("❌ [SECRET-PROMO] Failed to create checkout session: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create checkout session", "details": err.Error()})
				return
			}

			log.Printf("✅ [SECRET-PROMO] Checkout session created for user %d", userIDInt)
			c.JSON(http.StatusOK, gin.H{
				"client_secret": clientSecret,
			})
		})

		// Verify checkout session status and grant immediate access
		authenticatedStripe.GET("/session/:session_id", func(c *gin.Context) {
			sessionID := c.Param("session_id")
			if sessionID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
				return
			}

			// Get user from context for security
			userIDInterface, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
				return
			}

			userID, ok := userIDInterface.(int)
			if !ok {
				log.Printf("❌ [SESSION-VERIFY] Invalid user_id type: %T", userIDInterface)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context"})
				return
			}

			log.Printf("🔍 [SESSION-VERIFY] User %d verifying session: %s", userID, sessionID)

			// Use new VerifyAndGrantAccess method for immediate access
			sessionData, err := stripePublicService.VerifyAndGrantAccess(sessionID, userID)
			if err != nil {
				log.Printf("❌ [SESSION-VERIFY] Failed to verify session: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			log.Printf("✅ [SESSION-VERIFY] Session verified successfully: %s", sessionID)
			c.JSON(http.StatusOK, gin.H{
				"status": "success",
				"data":   sessionData,
			})
		})

		// Create subscription from successful payment
		authenticatedStripe.POST("/create-subscription", func(c *gin.Context) {
			// Get user from context
			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
				return
			}

			// Parse request
			var request struct {
				SessionID   string                 `json:"session_id"`
				SessionData map[string]interface{} `json:"session_data"`
			}

			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
				return
			}

			log.Printf("🔍 [CREATE-SUB] Creating subscription for user %v, session: %s", userID, request.SessionID)

			// DEPRECATED: Phase 7 replaces this
			log.Printf("⚠️ [CREATE-SUB] Deprecated endpoint - use Phase 7 subscription management")
			c.JSON(http.StatusGone, gin.H{"error": "This endpoint has been replaced. Please use /api/v1/user/subscriptions"})
			return
		})

		// Get user subscription status (including legacy fields)
		authenticatedStripe.GET("/subscription-status", func(c *gin.Context) {
			// DEPRECATED: Phase 7 replaces this
			log.Printf("⚠️ [SUB-STATUS] Deprecated endpoint - use Phase 7 subscription management")
			c.JSON(http.StatusGone, gin.H{"error": "This endpoint has been replaced. Please use /api/v1/user/subscriptions"})
		})

		// Cancel user subscription
		authenticatedStripe.POST("/cancel-subscription", func(c *gin.Context) {
			// DEPRECATED: Phase 7 replaces this
			log.Printf("⚠️ [CANCEL-SUB] Deprecated endpoint - use Phase 7 subscription management")
			c.JSON(http.StatusGone, gin.H{"error": "This endpoint has been replaced. Please use /api/v1/user/subscriptions/:id/cancel"})
		})
	}

	log.Printf("✅ [STRIPE-AUTH] Authenticated Stripe routes setup complete")
}

// setupFallbackStripeRoutes sets up fallback routes when Stripe service is not available
func setupFallbackStripeRoutes(v1 *gin.RouterGroup) {
	log.Printf("⚠️ [STRIPE-FALLBACK] Setting up fallback Stripe routes...")

	// Public Stripe routes (no auth required, even in fallback)
	publicStripe := v1.Group("/stripe")
	{
		publicStripe.GET("/config", func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Stripe service temporarily unavailable",
			})
		})
	}

	// Authenticated Stripe routes (require auth in fallback)
	authenticatedStripe := v1.Group("/stripe")
	authenticatedStripe.Use(middleware.AuthRequired())
	{
		authenticatedStripe.GET("/portal-link", func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Service temporarily unavailable",
			})
		})
		authenticatedStripe.POST("/checkout-session", func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Service temporarily unavailable",
			})
		})
		authenticatedStripe.GET("/session/:session_id", func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Service temporarily unavailable",
			})
		})
		authenticatedStripe.POST("/create-subscription", func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Service temporarily unavailable",
			})
		})
		authenticatedStripe.GET("/subscription-status", func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Service temporarily unavailable",
			})
		})
		authenticatedStripe.POST("/cancel-subscription", func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Service temporarily unavailable",
			})
		})
	}

	log.Printf("✅ [STRIPE-FALLBACK] Fallback Stripe routes setup complete")
}
