package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bome-backend/internal/database"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// Logger returns a gin.HandlerFunc for logging
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// SecurityHeaders adds comprehensive security headers to all responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Basic security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=()")

		// Strict Transport Security (HSTS)
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// Cross-Origin Embedder Policy
		c.Header("Cross-Origin-Embedder-Policy", "require-corp")

		// Cross-Origin Opener Policy
		c.Header("Cross-Origin-Opener-Policy", "same-origin")

		// Cross-Origin Resource Policy
		c.Header("Cross-Origin-Resource-Policy", "same-origin")

		// Enhanced Content Security Policy
		nonce := generateNonce()
		c.Set("csp_nonce", nonce)

		csp := "default-src 'self'; " +
			"script-src 'self' 'nonce-" + nonce + "' https://js.stripe.com; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https: blob:; " +
			"font-src 'self' data:; " +
			"connect-src 'self' http://localhost:3000 http://localhost:4173 http://localhost:5173 http://localhost:8080 https://api.stripe.com https://bunnycdn.com; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self';"

		c.Header("Content-Security-Policy", csp)

		c.Next()
	}
}

// generateNonce generates a cryptographically secure nonce for CSP
func generateNonce() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to time-based nonce if crypto/rand fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.StdEncoding.EncodeToString(bytes)
}

// CORS configures Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// In production, implement proper origin validation
		// For now, allow specific origins or all during development
		allowedOrigins := []string{
			"http://localhost:3000", // Production frontend
			"http://localhost:4173", // SvelteKit preview
			"http://localhost:5173", // SvelteKit dev server
		}

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		// Also allow DigitalOcean App Platform URLs
		if !allowed && strings.HasSuffix(origin, ".ondigitalocean.app") && strings.HasPrefix(origin, "https://") {
			allowed = true
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With, X-User-Data, X-Frontend-Request-ID, X-Frontend-Timestamp, X-Frontend-User-Agent")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Recovery returns a gin.HandlerFunc for recovery
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.String(http.StatusInternalServerError, fmt.Sprintf("error: %s", err))
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

// RateLimiting implements basic rate limiting per IP
func RateLimiting() gin.HandlerFunc {
	// Increased rate limiter: 200 requests per minute per IP to accommodate video preloading
	rateLimiter := services.NewRateLimiter(200, time.Minute)

	return func(c *gin.Context) {
		clientIP := services.GetClientIP(c.Request.RemoteAddr, c.GetHeader("X-Forwarded-For"), c.GetHeader("X-Real-IP"))

		if !rateLimiter.Allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded. Please slow down.",
				"retry_after": "60s",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthRequired middleware that requires a valid JWT token
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Check Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Use 'Bearer <token>'",
			})
			c.Abort()
			return
		}

		token := parts[1]
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token is required",
			})
			c.Abort()
			return
		}

		// Parse and validate token
		claims, err := services.ParseToken(token)
		if err != nil {
			log.Printf("Token validation failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Additional token validation
		if err := services.ValidateTokenClaims(claims); err != nil {
			log.Printf("Token claims validation failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
			})
			c.Abort()
			return
		}

		// Store user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("email_verified", claims.EmailVerified)
		c.Set("token_id", claims.TokenID) // Store token ID for session tracking

		// Log successful authentication
		log.Printf("Authenticated user: %s (ID: %d, Role: %s)", claims.Email, claims.UserID, claims.Role)

		c.Next()
	}
}

// AdminRequired middleware that requires admin role
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context (set by AuthRequired)
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// Admin roles include all roles with level 7+ (subsystem managers and above)
		adminRoles := []string{
			"super_admin",           // Level 10: Super Administrator
			"system_admin",          // Level 9: System Administrator
			"content_manager",       // Level 8: Content Manager
			"articles_manager",      // Level 7: Articles Manager
			"youtube_manager",       // Level 7: YouTube Manager
			"streaming_manager",     // Level 7: Video Streaming Manager
			"events_manager",        // Level 7: Events Manager
			"advertisement_manager", // Level 7: Advertisement Manager
			"user_manager",          // Level 7: User Account Manager
			"analytics_manager",     // Level 7: Analytics Manager
			"financial_admin",       // Level 7: Financial Administrator
			"admin",                 // Legacy admin role
		}

		roleStr := role.(string)
		isAdmin := false
		for _, adminRole := range adminRoles {
			if roleStr == adminRole {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			userEmail, _ := c.Get("user_email")
			log.Printf("Admin access denied for user: %v (role: %s)", userEmail, roleStr)
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// StreamingAdminRequired middleware that requires streaming manager role or higher
func StreamingAdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context (set by AuthRequired)
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// Streaming admin roles include streaming manager and higher roles
		streamingAdminRoles := []string{
			"super_admin",       // Level 10: Super Administrator
			"system_admin",      // Level 9: System Administrator
			"content_manager",   // Level 8: Content Manager
			"streaming_manager", // Level 7: Video Streaming Manager
		}

		roleStr := role.(string)
		isStreamingAdmin := false
		for _, adminRole := range streamingAdminRoles {
			if roleStr == adminRole {
				isStreamingAdmin = true
				break
			}
		}

		if !isStreamingAdmin {
			userEmail, _ := c.Get("user_email")
			log.Printf("Streaming admin access denied for user: %v (role: %s)", userEmail, roleStr)
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Streaming admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SubscriptionAccessRequired middleware that requires active subscription for protected content
func SubscriptionAccessRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// Get user role from context
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User role not found",
			})
			c.Abort()
			return
		}

		// Admin roles get automatic access
		adminRoles := []string{
			"super_admin",           // Level 10: Super Administrator
			"system_admin",          // Level 9: System Administrator
			"content_manager",       // Level 8: Content Manager
			"articles_manager",      // Level 7: Articles Manager
			"youtube_manager",       // Level 7: YouTube Manager
			"streaming_manager",     // Level 7: Video Streaming Manager
			"events_manager",        // Level 7: Events Manager
			"advertisement_manager", // Level 7: Advertisement Manager
			"user_manager",          // Level 7: User Account Manager
			"analytics_manager",     // Level 7: Analytics Manager
			"financial_admin",       // Level 7: Financial Administrator
			"admin",                 // Legacy admin role
		}

		roleStr := role.(string)
		isAdmin := false
		for _, adminRole := range adminRoles {
			if roleStr == adminRole {
				isAdmin = true
				break
			}
		}

		if isAdmin {
			c.Next()
			return
		}

		// For non-admin users, check subscription status
		// TODO: Implement subscription status check from database
		// For now, allow access (this should be implemented with actual subscription checking)
		c.Next()
	}
}

// SubscriptionPlanRequired middleware that requires specific subscription plan
func SubscriptionPlanRequired(requiredPlan string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// Get user role from context
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User role not found",
			})
			c.Abort()
			return
		}

		// Admin roles get automatic access
		adminRoles := []string{
			"super_admin",       // Level 10: Super Administrator
			"system_admin",      // Level 9: System Administrator
			"content_manager",   // Level 8: Content Manager
			"streaming_manager", // Level 7: Video Streaming Manager
		}

		roleStr := role.(string)
		isAdmin := false
		for _, adminRole := range adminRoles {
			if roleStr == adminRole {
				isAdmin = true
				break
			}
		}

		if isAdmin {
			c.Next()
			return
		}

		// For non-admin users, check subscription plan
		// TODO: Implement subscription plan check from database
		// For now, allow access (this should be implemented with actual plan checking)
		c.Next()
	}
}

// SubscriptionAuditLog middleware that logs subscription-related operations
func SubscriptionAuditLog(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer for operation duration
		start := time.Now()

		// Process request
		c.Next()

		// Log subscription-related operations
		userID := c.GetInt("user_id")
		userEmail, _ := c.Get("user_email")
		role, _ := c.Get("user_role")
		method := c.Request.Method
		path := c.Request.URL.Path
		statusCode := c.Writer.Status()
		duration := time.Since(start)

		// Check if this is a subscription-related operation
		subscriptionPaths := []string{
			"/api/subscriptions",
			"/api/admin/subscriptions",
			"/api/admin/streaming",
			"/api/subscription",
		}

		isSubscriptionOperation := false
		for _, subscriptionPath := range subscriptionPaths {
			if strings.HasPrefix(path, subscriptionPath) {
				isSubscriptionOperation = true
				break
			}
		}

		if isSubscriptionOperation {
			log.Printf("SUBSCRIPTION_AUDIT: user_id=%d, email=%v, role=%v, method=%s, path=%s, status=%d, duration=%v, ip=%s",
				userID, userEmail, role, method, path, statusCode, duration, c.ClientIP())

			// TODO: Store audit log in database for compliance and analytics
			// This would integrate with the subscription_events table
		}
	}
}

// SubscriptionRateLimit middleware that applies rate limiting to subscription operations
func SubscriptionRateLimit() gin.HandlerFunc {
	// Create a rate limiter for subscription operations
	// This is a simplified implementation - in production, use a proper rate limiting library
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.Next()
			return
		}

		// Check if this is a subscription-related operation
		path := c.Request.URL.Path
		subscriptionPaths := []string{
			"/api/subscriptions",
			"/api/admin/subscriptions",
			"/api/admin/streaming",
			"/api/subscription",
		}

		isSubscriptionOperation := false
		for _, subscriptionPath := range subscriptionPaths {
			if strings.HasPrefix(path, subscriptionPath) {
				isSubscriptionOperation = true
				break
			}
		}

		if isSubscriptionOperation {
			// TODO: Implement proper rate limiting
			// For now, just continue
			c.Next()
			return
		}

		c.Next()
	}
}

// SubscriptionValidation middleware that validates subscription access
func SubscriptionValidation(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// Get user role from context
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User role not found",
			})
			c.Abort()
			return
		}

		// Admin roles get automatic access
		adminRoles := []string{
			"super_admin",       // Level 10: Super Administrator
			"system_admin",      // Level 9: System Administrator
			"content_manager",   // Level 8: Content Manager
			"streaming_manager", // Level 7: Video Streaming Manager
		}

		roleStr := role.(string)
		isAdmin := false
		for _, adminRole := range adminRoles {
			if roleStr == adminRole {
				isAdmin = true
				break
			}
		}

		if isAdmin {
			c.Next()
			return
		}

	// For non-admin users, validate subscription using Stripe sync data
	if db != nil {
		hasActiveSub, subInfo, err := db.HasActiveStripeSubscription(userID)
		if err != nil {
			log.Printf("Error checking subscription for user %d: %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Subscription check failed",
				"code":  "SUBSCRIPTION_CHECK_ERROR",
			})
			c.Abort()
			return
		}
		
		if !hasActiveSub {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Active subscription required",
				"code":  "SUBSCRIPTION_REQUIRED",
				"message": "You need an active subscription to access this content",
			})
			c.Abort()
			return
		}
		
		// Store subscription info in context for later use
		c.Set("stripe_subscription", subInfo)
		log.Printf("✅ User %d has active subscription: %s", userID, subInfo.SubscriptionID)
	}

		c.Next()
	}
}

// SubscriptionPlanValidation middleware that validates specific subscription plan access
func SubscriptionPlanValidation(db *database.DB, requiredPlan string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// Get user role from context
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User role not found",
			})
			c.Abort()
			return
		}

		// Admin roles get automatic access
		adminRoles := []string{
			"super_admin",       // Level 10: Super Administrator
			"system_admin",      // Level 9: System Administrator
			"content_manager",   // Level 8: Content Manager
			"streaming_manager", // Level 7: Video Streaming Manager
		}

		roleStr := role.(string)
		isAdmin := false
		for _, adminRole := range adminRoles {
			if roleStr == adminRole {
				isAdmin = true
				break
			}
		}

		if isAdmin {
			c.Next()
			return
		}

		// For non-admin users, validate subscription plan
		if db != nil {
			subscription, err := db.GetSubscriptionByUserID(userID)
			if err != nil || subscription == nil {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Active subscription required",
					"code":  "SUBSCRIPTION_REQUIRED",
				})
				c.Abort()
				return
			}

			// Check if subscription is active
			if subscription.Status != "active" {
				c.JSON(http.StatusForbidden, gin.H{
					"error":  "Subscription is not active",
					"code":   "SUBSCRIPTION_INACTIVE",
					"status": subscription.Status,
				})
				c.Abort()
				return
			}

			// Check if subscription has expired
			if subscription.CurrentPeriodEnd != nil && time.Now().After(*subscription.CurrentPeriodEnd) {
				c.JSON(http.StatusForbidden, gin.H{
					"error":      "Subscription has expired",
					"code":       "SUBSCRIPTION_EXPIRED",
					"expired_at": subscription.CurrentPeriodEnd,
				})
				c.Abort()
				return
			}

			// Check if user has the required plan
			if subscription.PlanID.Valid {
				plan, err := db.GetSubscriptionPlanByID(int(subscription.PlanID.Int32))
				if err == nil && plan != nil {
					// TODO: Implement plan hierarchy checking
					// For now, just check if the plan name matches
					if plan.Name != requiredPlan {
						c.JSON(http.StatusForbidden, gin.H{
							"error":         "Higher subscription plan required",
							"code":          "PLAN_UPGRADE_REQUIRED",
							"required_plan": requiredPlan,
							"current_plan":  plan.Name,
						})
						c.Abort()
						return
					}
				}
			}

			// Store subscription info in context for later use
			c.Set("user_subscription", subscription)
		}

		c.Next()
	}
}

// SubscriptionExpirationWarning middleware that adds expiration warnings to responses
func SubscriptionExpirationWarning(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			c.Next()
			return
		}

		// Get user role from context
		role, exists := c.Get("user_role")
		if !exists {
			c.Next()
			return
		}

		// Admin roles don't need expiration warnings
		adminRoles := []string{
			"super_admin",       // Level 10: Super Administrator
			"system_admin",      // Level 9: System Administrator
			"content_manager",   // Level 8: Content Manager
			"streaming_manager", // Level 7: Video Streaming Manager
		}

		roleStr := role.(string)
		isAdmin := false
		for _, adminRole := range adminRoles {
			if roleStr == adminRole {
				isAdmin = true
				break
			}
		}

		if isAdmin {
			c.Next()
			return
		}

		// For non-admin users, check subscription expiration
		if db != nil {
			subscription, err := db.GetSubscriptionByUserID(userID)
			if err == nil && subscription != nil && subscription.CurrentPeriodEnd != nil {
				daysUntilExpiration := int(time.Until(*subscription.CurrentPeriodEnd).Hours() / 24)

				// Add warning headers for expiring subscriptions
				if daysUntilExpiration <= 7 && daysUntilExpiration > 0 {
					c.Header("X-Subscription-Warning", "expiring_soon")
					c.Header("X-Subscription-Expires-In", strconv.Itoa(daysUntilExpiration)+" days")
				} else if daysUntilExpiration <= 0 {
					c.Header("X-Subscription-Warning", "expired")
					c.Header("X-Subscription-Expired", subscription.CurrentPeriodEnd.Format(time.RFC3339))
				}

				// Store expiration info in context
				c.Set("subscription_expiration_days", daysUntilExpiration)
			}
		}

		c.Next()
	}
}

// EmailVerificationRequired middleware that requires verified email
func EmailVerificationRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get email verification status from context
		emailVerified, exists := c.Get("email_verified")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		if !emailVerified.(bool) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Email verification required. Please check your email and verify your account.",
				"code":  "EMAIL_NOT_VERIFIED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth middleware that extracts user info if token is present but doesn't require it
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format, continue without authentication
			c.Next()
			return
		}

		token := parts[1]
		claims, err := services.ParseToken(token)
		if err != nil {
			// Invalid token, continue without authentication
			log.Printf("Optional auth - token validation failed: %v", err)
			c.Next()
			return
		}

		// Valid token, store user info
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("email_verified", claims.EmailVerified)

		c.Next()
	}
}

// UserOwnershipRequired middleware that ensures user can only access their own resources
func UserOwnershipRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authenticated user ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// Get requested user ID from URL parameter
		requestedUserID := c.Param("user_id")
		if requestedUserID == "" {
			requestedUserID = c.Param("id")
		}

		// Allow if user is accessing their own resource or is admin
		userRole, _ := c.Get("user_role")

		// Admin roles include all roles with level 7+ (subsystem managers and above)
		adminRoles := []string{
			"super_admin",           // Level 10: Super Administrator
			"system_admin",          // Level 9: System Administrator
			"content_manager",       // Level 8: Content Manager
			"articles_manager",      // Level 7: Articles Manager
			"youtube_manager",       // Level 7: YouTube Manager
			"streaming_manager",     // Level 7: Video Streaming Manager
			"events_manager",        // Level 7: Events Manager
			"advertisement_manager", // Level 7: Advertisement Manager
			"user_manager",          // Level 7: User Account Manager
			"analytics_manager",     // Level 7: Analytics Manager
			"financial_admin",       // Level 7: Financial Administrator
			"admin",                 // Legacy admin role
		}

		// Check if user has admin role
		isAdmin := false
		if userRoleStr, ok := userRole.(string); ok {
			for _, adminRole := range adminRoles {
				if userRoleStr == adminRole {
					isAdmin = true
					break
				}
			}
		}

		if isAdmin || fmt.Sprintf("%d", userID) == requestedUserID {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied. You can only access your own resources.",
		})
		c.Abort()
	}
}

// ContentSecurityPolicy applies stricter CSP for specific routes
func ContentSecurityPolicy(policy string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Security-Policy", policy)
		c.Next()
	}
}

// RecoveryWithLogging provides panic recovery with logging
func RecoveryWithLogging() gin.HandlerFunc {
	return gin.RecoveryWithWriter(gin.DefaultWriter, func(c *gin.Context, recovered interface{}) {
		if recovered != nil {
			log.Printf("Panic recovered: %v", recovered)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		}
	})
}

// SessionActivityTracker updates session activity for authenticated requests
func SessionActivityTracker(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only track activity for authenticated requests
		if tokenID, exists := c.Get("token_id"); exists && db != nil {
			if tokenIDStr, ok := tokenID.(string); ok && tokenIDStr != "" {
				// Update session activity asynchronously to not block the request
				go func() {
					if err := db.UpdateSessionActivityByTokenID(tokenIDStr); err != nil {
						log.Printf("Failed to update session activity: %v", err)
					}
				}()
			}
		}
		c.Next()
	}
}

// ClientInfo represents rate limiting information for a client
type ClientInfo struct {
	Requests int
	LastSeen time.Time
}

// CSRFProtection middleware that validates CSRF tokens
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF check for GET requests and OPTIONS
		if c.Request.Method == "GET" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Skip CSRF check for API endpoints that use token-based auth
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/auth/") {
			c.Next()
			return
		}

		token := c.GetHeader("X-CSRF-Token")
		if token == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token required"})
			c.Abort()
			return
		}

		// Validate token against session
		if !validateCSRFToken(c, token) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid CSRF token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// validateCSRFToken validates the CSRF token
func validateCSRFToken(c *gin.Context, token string) bool {
	// For now, implement basic validation
	// In production, this should validate against a session-based token

	// Check if token is at least 32 characters (basic security)
	if len(token) < 32 {
		return false
	}

	// Check if token contains only valid characters
	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for _, char := range token {
		if !strings.ContainsRune(validChars, char) {
			return false
		}
	}

	// TODO: Implement proper session-based CSRF validation
	// This would involve:
	// 1. Storing CSRF tokens in user sessions
	// 2. Validating tokens against stored session tokens
	// 3. Regenerating tokens on login/logout

	return true
}

// VideoUploadRequired middleware that requires video upload permissions
func VideoUploadRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context (set by AuthRequired)
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// Video upload roles include admin and content management roles
		uploadRoles := []string{
			"super_admin",       // Level 10: Super Administrator
			"system_admin",      // Level 9: System Administrator
			"content_manager",   // Level 8: Content Manager
			"youtube_manager",   // Level 7: YouTube Manager
			"streaming_manager", // Level 7: Video Streaming Manager
		}

		roleStr := role.(string)
		canUpload := false
		for _, uploadRole := range uploadRoles {
			if roleStr == uploadRole {
				canUpload = true
				break
			}
		}

		if !canUpload {
			userEmail, _ := c.Get("user_email")
			log.Printf("Video upload denied for user: %v (role: %s)", userEmail, roleStr)
			c.JSON(http.StatusForbidden, gin.H{
				"error":          "Video upload access required. Only administrators and content managers can upload videos.",
				"required_roles": uploadRoles,
				"user_role":      roleStr,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
