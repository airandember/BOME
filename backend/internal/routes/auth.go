package routes

import (
	"database/sql"
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
		// Rate limiting
		clientIP := services.GetClientIP(c.Request.RemoteAddr, c.GetHeader("X-Forwarded-For"), c.GetHeader("X-Real-IP"))
		if !services.LoginRateLimiter.Allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many login attempts. Please try again in 15 minutes.",
			})
			return
		}

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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
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

		// Update last login timestamp
		if err := db.UpdateLastLogin(user.ID); err != nil {
			log.Printf("Failed to update last login: %v", err)
			// Continue anyway - this is not critical
		}

		// Log successful login
		log.Printf("User logged in successfully: %s (ID: %d) from %s", user.Email, user.ID, clientIP)

		c.JSON(http.StatusOK, gin.H{
			"access_token":  tokenPair.AccessToken,
			"refresh_token": tokenPair.RefreshToken,
			"expires_in":    tokenPair.ExpiresIn,
			"token_type":    tokenPair.TokenType,
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

// LogoutHandler handles user logout (mainly for logging purposes)
func LogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context if available
		if userID, exists := c.Get("user_id"); exists {
			if email, exists := c.Get("user_email"); exists {
				log.Printf("User logged out: %s (ID: %v)", email, userID)
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
	}
}
