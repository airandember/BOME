package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
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
			"connect-src 'self' https://api.stripe.com https://bunnycdn.com; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'; " +
			"upgrade-insecure-requests; " +
			"block-all-mixed-content;"

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
			"http://localhost:3000",
			"http://localhost:5173",
			"https://bome.example.com", // Replace with actual domain
		}

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
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
	// Global rate limiter: 100 requests per minute per IP
	rateLimiter := services.NewRateLimiter(100, time.Minute)

	return func(c *gin.Context) {
		clientIP := services.GetClientIP(c.Request.RemoteAddr, c.GetHeader("X-Forwarded-For"), c.GetHeader("X-Real-IP"))

		if !rateLimiter.Allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please slow down.",
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
