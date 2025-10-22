package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RequireEmailVerification middleware ensures users have verified their email
// before accessing protected routes
func RequireEmailVerification() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip verification for certain routes
		path := c.Request.URL.Path

		// Allow these routes without email verification
		skipRoutes := []string{
			"/api/v1/auth/verify-email",
			"/api/v1/auth/resend-verification",
			"/api/v1/auth/logout",
			"/api/v1/oauth2/",
			"/health",
		}

		for _, route := range skipRoutes {
			if strings.HasPrefix(path, route) {
				c.Next()
				return
			}
		}

		// Get user from context (set by JWT middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// Get email verification status
		emailVerified, exists := c.Get("email_verified")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Unable to verify email status",
			})
			c.Abort()
			return
		}

		// Check if email is verified
		if verified, ok := emailVerified.(bool); !ok || !verified {
			c.JSON(http.StatusForbidden, gin.H{
				"error":                 "Email verification required",
				"message":               "Please verify your email address before accessing this resource",
				"verification_required": true,
				"user_id":               userID,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireEmailVerificationForDashboard is a stricter version for dashboard access
func RequireEmailVerificationForDashboard() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// Get email verification status
		emailVerified, exists := c.Get("email_verified")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Unable to verify email status",
			})
			c.Abort()
			return
		}

		// Check if email is verified
		if verified, ok := emailVerified.(bool); !ok || !verified {
			c.JSON(http.StatusForbidden, gin.H{
				"error":                 "Email verification required for dashboard access",
				"message":               "You must verify your email address before accessing the dashboard. Please check your email for a verification link.",
				"verification_required": true,
				"user_id":               userID,
				"redirect_to":           "/auth/verify-email-required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
