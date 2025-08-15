package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RequirePasswordChange middleware that checks if user needs to change password
// This should be used after AuthRequired middleware
// The password_changed status should be set in the route handler or context
func RequirePasswordChange() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get password_changed status from context
		// This should be set by the route handler after checking the database
		passwordChanged, exists := c.Get("password_changed")
		if !exists {
			// If not set, allow access (fallback for backward compatibility)
			c.Next()
			return
		}

		// Check if password needs to be changed
		if !passwordChanged.(bool) {
			// Allow access to password change endpoint
			if strings.HasSuffix(c.Request.URL.Path, "/auth/change-password") {
				c.Next()
				return
			}

			// Block access to all other endpoints
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Password change required",
				"code":  "PASSWORD_CHANGE_REQUIRED",
				"message": "You must change your password before accessing this resource. " +
					"This is required for security reasons.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
