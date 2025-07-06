package routes

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"bome-backend/internal/database"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// RegisterRequest represents the registration payload
type RegisterRequest struct {
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

// LoginRequest represents the login payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ForgotPasswordRequest represents the forgot password payload
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required"`
}

// ResetPasswordRequest represents the reset password payload
type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest represents the refresh token payload
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// VerifyEmailRequest represents the email verification payload
type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

// ChangePasswordRequest represents the change password payload
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

// LogoutRequest represents a logout request
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	AllDevices   bool   `json:"all_devices"` // Optional: logout from all devices
}

// RegisterHandler handles user registration
func RegisterHandler(db *database.DB, emailService *services.EmailService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Rate limiting
		clientIP := services.GetClientIP(c.Request.RemoteAddr, c.GetHeader("X-Forwarded-For"), c.GetHeader("X-Real-IP"))
		if !services.RegisterRateLimiter.Allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many registration attempts. Please try again later.",
			})
			return
		}

		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// Validate and sanitize input
		req.Email = strings.ToLower(services.SanitizeString(req.Email))
		req.FirstName = services.SanitizeString(req.FirstName)
		req.LastName = services.SanitizeString(req.LastName)

		// Validate email
		if err := services.ValidateEmail(req.Email); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate password strength
		if err := services.ValidatePassword(req.Password); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate names
		if err := services.ValidateName(req.FirstName); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid first name: " + err.Error()})
			return
		}
		if err := services.ValidateName(req.LastName); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid last name: " + err.Error()})
			return
		}

		// Check if database is available
		if db == nil {
			log.Printf("Database not available for registration request from %s", clientIP)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Service temporarily unavailable. Please try again later.",
			})
			return
		}

		// Check if user already exists
		exists, err := db.CheckUserExists(req.Email)
		if err != nil {
			log.Printf("Database error checking user existence: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Service temporarily unavailable. Please try again later.",
			})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "An account with this email already exists"})
			return
		}

		// Hash password
		hash, err := services.HashPassword(req.Password)
		if err != nil {
			log.Printf("Failed to hash password: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Service temporarily unavailable. Please try again later.",
			})
			return
		}

		// Create user
		user, err := db.CreateUser(req.Email, hash, req.FirstName, req.LastName, "user")
		if err != nil {
			log.Printf("Failed to create user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Service temporarily unavailable. Please try again later.",
			})
			return
		}

		// Generate email verification token
		verificationToken := services.GenerateSecureToken()
		if err := db.SetVerificationToken(user.ID, verificationToken); err != nil {
			log.Printf("Failed to set verification token: %v", err)
			// Continue anyway - user can request new verification
		}

		// Send verification email
		if emailService != nil {
			if err := emailService.SendEmailVerification(user.FirstName, user.Email, verificationToken); err != nil {
				log.Printf("Failed to send verification email: %v", err)
				// Continue anyway - user can request new verification
			}
		}

		// Log successful registration
		log.Printf("User registered successfully: %s (ID: %d) from %s", user.Email, user.ID, clientIP)

		c.JSON(http.StatusCreated, gin.H{
			"message": "Registration successful. Please check your email to verify your account.",
			"user_id": user.ID,
		})
	}
}

// LoginHandler handles user login
func LoginHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Enhanced rate limiting
		clientIP := services.GetClientIP(c.Request.RemoteAddr, c.GetHeader("X-Forwarded-For"), c.GetHeader("X-Real-IP"))

		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// Validate and sanitize input
		req.Email = strings.ToLower(services.SanitizeString(req.Email))

		// Validate email format
		if err := services.ValidateEmail(req.Email); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
			return
		}

		// Check enhanced rate limiting
		if !services.EnhancedLoginRateLimiter.CheckLoginAttempt(req.Email, clientIP) {
			remainingTime := services.EnhancedLoginRateLimiter.GetRemainingLockoutTime(req.Email)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":             fmt.Sprintf("Account temporarily locked. Please try again in %v", remainingTime),
				"lockout_remaining": remainingTime.String(),
			})
			return
		}

		// Check if database is available
		if db == nil {
			log.Printf("Database not available for login request from %s", clientIP)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Service temporarily unavailable. Please try again later.",
			})
			return
		}

		// Get user by email
		user, err := db.GetUserByEmail(req.Email)
		if err != nil {
			if err == sql.ErrNoRows {
				// Record failed attempt
				services.EnhancedLoginRateLimiter.RecordFailedAttempt(req.Email, clientIP)

				// Log failed login attempt
				if db != nil {
					auditLog := &database.AuditLog{
						UserEmail: &req.Email,
						Action:    "login",
						Resource:  "authentication",
						IPAddress: clientIP,
						UserAgent: c.GetHeader("User-Agent"),
						Status:    "failed",
						Details:   &[]string{"Invalid email or password"}[0],
						Severity:  "medium",
					}
					db.CreateAuditLog(auditLog)
				}

				// Don't reveal if email exists or not
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
				return
			}
			log.Printf("Database error during login: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Service temporarily unavailable. Please try again later.",
			})
			return
		}

		// Verify password
		if err := services.CheckPassword(user.PasswordHash, req.Password); err != nil {
			// Record failed attempt
			services.EnhancedLoginRateLimiter.RecordFailedAttempt(req.Email, clientIP)

			// Log failed login attempt
			if db != nil {
				auditLog := &database.AuditLog{
					UserID:    &user.ID,
					UserEmail: &user.Email,
					Action:    "login",
					Resource:  "authentication",
					IPAddress: clientIP,
					UserAgent: c.GetHeader("User-Agent"),
					Status:    "failed",
					Details:   &[]string{"Invalid password"}[0],
					Severity:  "medium",
				}
				db.CreateAuditLog(auditLog)
			}

			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		// Record successful attempt
		services.EnhancedLoginRateLimiter.RecordSuccessfulAttempt(req.Email)

		// Check session limit
		maxSessions := 5 // Default max sessions
		if user.MaxSessions > 0 {
			maxSessions = user.MaxSessions
		}

		canCreateSession, err := db.CheckSessionLimit(user.ID, maxSessions)
		if err != nil {
			log.Printf("Error checking session limit: %v", err)
			// Continue anyway - don't block login for session limit errors
		} else if !canCreateSession {
			// Log session limit exceeded
			if db != nil {
				auditLog := &database.AuditLog{
					UserID:    &user.ID,
					UserEmail: &user.Email,
					Action:    "login",
					Resource:  "authentication",
					IPAddress: clientIP,
					UserAgent: c.GetHeader("User-Agent"),
					Status:    "warning",
					Details:   &[]string{"Session limit exceeded"}[0],
					Severity:  "medium",
				}
				db.CreateAuditLog(auditLog)
			}
		}

		// Generate token pair
		tokenPair, err := services.GenerateTokenPair(user.ID, user.Email, user.Role, user.EmailVerified)
		if err != nil {
			log.Printf("Failed to generate tokens: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Service temporarily unavailable. Please try again later.",
			})
			return
		}

		// Create session record
		var session *database.Session
		if db != nil {
			deviceInfo := services.GenerateDeviceFingerprint(c.Request)
			// Extract token ID from refresh token for session tracking
			refreshClaims, _ := services.ParseRefreshToken(tokenPair.RefreshToken)
			tokenID := ""
			if refreshClaims != nil {
				tokenID = refreshClaims.TokenID
			}

			session, err = db.CreateSession(
				user.ID,
				tokenID,
				deviceInfo,
				clientIP,
				c.GetHeader("User-Agent"),
				time.Now().Add(7*24*time.Hour), // Session expires with refresh token
			)
			if err != nil {
				log.Printf("Failed to create session: %v", err)
				// Continue anyway - don't block login for session creation errors
			}
		}

		// Update last login timestamp
		if err := db.UpdateLastLogin(user.ID); err != nil {
			log.Printf("Failed to update last login: %v", err)
			// Continue anyway - this is not critical
		}

		// Log successful login
		if db != nil {
			auditLog := &database.AuditLog{
				UserID:    &user.ID,
				UserEmail: &user.Email,
				Action:    "login",
				Resource:  "authentication",
				IPAddress: clientIP,
				UserAgent: c.GetHeader("User-Agent"),
				Status:    "success",
				Details:   &[]string{"Login successful"}[0],
				Severity:  "low",
				Metadata: map[string]interface{}{
					"session_id":  session.ID,
					"device_info": session.DeviceInfo,
				},
			}
			db.CreateAuditLog(auditLog)
		}

		log.Printf("User logged in successfully: %s (ID: %d) from %s", user.Email, user.ID, clientIP)

		c.JSON(http.StatusOK, gin.H{
			"access_token":  tokenPair.AccessToken,
			"refresh_token": tokenPair.RefreshToken,
			"expires_in":    tokenPair.ExpiresIn,
			"token_type":    tokenPair.TokenType,
			"session_id":    session.ID,
			"user": gin.H{
				"id":             user.ID,
				"email":          user.Email,
				"role":           user.Role,
				"first_name":     user.FirstName,
				"last_name":      user.LastName,
				"email_verified": user.EmailVerified,
			},
		})
	}
}

// RefreshTokenHandler handles token refresh
func RefreshTokenHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// Parse and validate refresh token
		tokenPair, err := services.RefreshTokenPair(req.RefreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
			return
		}

		// Verify user still exists and is active
		claims, err := services.ParseRefreshToken(req.RefreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			return
		}

		if db != nil {
			_, err = db.GetUserByID(claims.UserID)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User account not found"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token":  tokenPair.AccessToken,
			"refresh_token": tokenPair.RefreshToken,
			"expires_in":    tokenPair.ExpiresIn,
			"token_type":    tokenPair.TokenType,
		})
	}
}

// ForgotPasswordHandler handles password reset requests
func ForgotPasswordHandler(db *database.DB, emailService *services.EmailService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Rate limiting
		clientIP := services.GetClientIP(c.Request.RemoteAddr, c.GetHeader("X-Forwarded-For"), c.GetHeader("X-Real-IP"))
		if !services.PasswordRateLimiter.Allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many password reset attempts. Please try again later.",
			})
			return
		}

		var req ForgotPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// Validate and sanitize email
		req.Email = strings.ToLower(services.SanitizeString(req.Email))
		if err := services.ValidateEmail(req.Email); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
			return
		}

		// Always return success to prevent email enumeration
		response := gin.H{"message": "If an account with this email exists, a password reset link has been sent."}

		// Check if database is available
		if db == nil {
			c.JSON(http.StatusOK, response)
			return
		}

		// Check if user exists
		user, err := db.GetUserByEmail(req.Email)
		if err != nil {
			// Don't reveal if email exists or not
			c.JSON(http.StatusOK, response)
			return
		}

		// Generate reset token
		resetToken := services.GenerateSecureToken()
		expiry := time.Now().Add(1 * time.Hour)

		if err := db.SetPasswordResetToken(user.ID, resetToken, expiry); err != nil {
			log.Printf("Failed to set reset token: %v", err)
			c.JSON(http.StatusOK, response)
			return
		}

		// Send reset email
		if emailService != nil {
			if err := emailService.SendPasswordResetEmail(user.FirstName, user.Email, resetToken); err != nil {
				log.Printf("Failed to send reset email: %v", err)
			}
		}

		log.Printf("Password reset requested for: %s from %s", user.Email, clientIP)
		c.JSON(http.StatusOK, response)
	}
}

// ResetPasswordHandler handles password reset with token
func ResetPasswordHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ResetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// Validate password
		if err := services.ValidatePassword(req.Password); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if database is available
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Service temporarily unavailable. Please try again later.",
			})
			return
		}

		// Get user by reset token
		user, err := db.GetUserByResetToken(req.Token)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired reset token"})
			return
		}

		// Hash new password
		hash, err := services.HashPassword(req.Password)
		if err != nil {
			log.Printf("Failed to hash password: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Service temporarily unavailable. Please try again later.",
			})
			return
		}

		// Update password
		if err := db.UpdateUserPassword(user.ID, hash); err != nil {
			log.Printf("Failed to update password: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Service temporarily unavailable. Please try again later.",
			})
			return
		}

		// Clear reset token
		if err := db.ClearPasswordResetToken(user.ID); err != nil {
			log.Printf("Failed to clear reset token: %v", err)
		}

		log.Printf("Password reset completed for: %s (ID: %d)", user.Email, user.ID)
		c.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
	}
}

// VerifyEmailHandler handles email verification
func VerifyEmailHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req VerifyEmailRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// Check if database is available
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Service temporarily unavailable. Please try again later.",
			})
			return
		}

		// Get user by verification token
		user, err := db.GetUserByVerificationToken(req.Token)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification token"})
			return
		}

		// Set email as verified
		if err := db.SetUserEmailVerified(user.ID); err != nil {
			log.Printf("Failed to verify email: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Service temporarily unavailable. Please try again later.",
			})
			return
		}

		// Clear verification token
		if err := db.ClearVerificationToken(user.ID); err != nil {
			log.Printf("Failed to clear verification token: %v", err)
		}

		log.Printf("Email verified for: %s (ID: %d)", user.Email, user.ID)
		c.JSON(http.StatusOK, gin.H{"message": "Email verification successful"})
	}
}

// ChangePasswordHandler handles password changes for authenticated users
func ChangePasswordHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context (set by auth middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		var req ChangePasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// Validate new password
		if err := services.ValidatePassword(req.NewPassword); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if database is available
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Service temporarily unavailable. Please try again later.",
			})
			return
		}

		// Get current user
		user, err := db.GetUserByID(userID.(int))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Service temporarily unavailable. Please try again later.",
			})
			return
		}

		// Verify current password
		if err := services.CheckPassword(user.PasswordHash, req.CurrentPassword); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is incorrect"})
			return
		}

		// Hash new password
		hash, err := services.HashPassword(req.NewPassword)
		if err != nil {
			log.Printf("Failed to hash password: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Service temporarily unavailable. Please try again later.",
			})
			return
		}

		// Update password
		if err := db.UpdateUserPassword(user.ID, hash); err != nil {
			log.Printf("Failed to update password: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Service temporarily unavailable. Please try again later.",
			})
			return
		}

		log.Printf("Password changed for: %s (ID: %d)", user.Email, user.ID)
		c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
	}
}

// LogoutHandler handles user logout with token blacklisting
func LogoutHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LogoutRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// Get user from context if available
		userID, userExists := c.Get("user_id")
		userEmail, emailExists := c.Get("user_email")
		clientIP := services.GetClientIP(c.Request.RemoteAddr, c.GetHeader("X-Forwarded-For"), c.GetHeader("X-Real-IP"))

		// Blacklist the refresh token
		if err := services.BlacklistToken(req.RefreshToken); err != nil {
			log.Printf("Failed to blacklist token: %v", err)
			// Continue with logout even if blacklisting fails
		}

		// If user is authenticated, log the logout and manage sessions
		if userExists && emailExists {
			userIDInt := userID.(int)
			userEmailStr := userEmail.(string)

			log.Printf("User logged out: %s (ID: %v) from %s", userEmailStr, userIDInt, clientIP)

			// Update last logout timestamp in database if available
			if db != nil {
				if err := db.UpdateLastLogout(userIDInt); err != nil {
					log.Printf("Failed to update last logout: %v", err)
				}

				// Log the logout event
				auditLog := &database.AuditLog{
					UserID:    &userIDInt,
					UserEmail: &userEmailStr,
					Action:    "logout",
					Resource:  "authentication",
					IPAddress: clientIP,
					UserAgent: c.GetHeader("User-Agent"),
					Status:    "success",
					Details:   &[]string{"User logged out successfully"}[0],
					Severity:  "low",
				}
				db.CreateAuditLog(auditLog)

				// If all_devices is true, deactivate all user's sessions
				if req.AllDevices {
					if err := db.DeactivateAllUserSessions(userIDInt); err != nil {
						log.Printf("Failed to deactivate all user sessions: %v", err)
					} else {
						// Log the all-devices logout
						auditLog := &database.AuditLog{
							UserID:    &userIDInt,
							UserEmail: &userEmailStr,
							Action:    "logout_all_devices",
							Resource:  "authentication",
							IPAddress: clientIP,
							UserAgent: c.GetHeader("User-Agent"),
							Status:    "success",
							Details:   &[]string{"All devices logged out"}[0],
							Severity:  "medium",
						}
						db.CreateAuditLog(auditLog)
					}
				} else {
					// Deactivate specific session if we have the session ID
					// This would require storing session ID in the token or request
					// For now, we'll just log the single device logout
				}
			}
		} else {
			log.Printf("Anonymous logout from %s", clientIP)

			// Log anonymous logout attempt
			if db != nil {
				auditLog := &database.AuditLog{
					Action:    "logout",
					Resource:  "authentication",
					IPAddress: clientIP,
					UserAgent: c.GetHeader("User-Agent"),
					Status:    "warning",
					Details:   &[]string{"Anonymous logout attempt"}[0],
					Severity:  "low",
				}
				db.CreateAuditLog(auditLog)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message":                "Logout successful",
			"all_devices_logged_out": req.AllDevices,
		})
	}
}
