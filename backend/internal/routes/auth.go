package routes

import (
	"net/http"
	"strings"
	"time"

	"bome-backend/internal/database"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// RegisterRequest represents the registration payload
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// LoginRequest represents the login payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// ForgotPasswordRequest represents the forgot password payload
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents the reset password payload
type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// VerifyEmailRequest represents the email verification payload
type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

// RegisterHandler handles user registration
func RegisterHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Development mode: mock registration when database is not available
		if db == nil {
			// In development mode, just return success for any registration
			c.JSON(http.StatusCreated, gin.H{
				"message": "Registration successful (development mode)",
				"user_id": 2,
			})
			return
		}

		exists, err := db.CheckUserExists(strings.ToLower(req.Email))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}

		hash, err := services.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		user, err := db.CreateUser(strings.ToLower(req.Email), hash, req.FirstName, req.LastName, "user")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		// TODO: Send verification email

		c.JSON(http.StatusCreated, gin.H{"message": "Registration successful", "user_id": user.ID})
	}
}

// LoginHandler handles user login
func LoginHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Development mode: hardcoded admin credentials when database is not available
		if db == nil {
			if strings.ToLower(req.Email) == "admin@bome.test" && req.Password == "Admin123!" {
				token, err := services.GenerateToken(1, req.Email, "admin", 24*time.Hour)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"token": token,
					"user": gin.H{
						"id":            1,
						"email":         req.Email,
						"role":          "admin",
						"firstName":     "Test",
						"lastName":      "Administrator",
						"emailVerified": true,
					},
				})
				return
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		// Normal database mode
		user, err := db.GetUserByEmail(strings.ToLower(req.Email))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		if err := services.CheckPassword(user.PasswordHash, req.Password); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		token, err := services.GenerateToken(user.ID, user.Email, user.Role, 24*time.Hour)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{
				"id":            user.ID,
				"email":         user.Email,
				"role":          user.Role,
				"firstName":     user.FirstName,
				"lastName":      user.LastName,
				"emailVerified": user.EmailVerified,
			},
		})
	}
}

// ForgotPasswordHandler handles password reset requests
func ForgotPasswordHandler(db *database.DB, emailService *services.EmailService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ForgotPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := db.GetUserByEmail(strings.ToLower(req.Email))
		if err != nil {
			// Don't reveal if email exists or not
			c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link has been sent"})
			return
		}

		// Generate reset token
		token := services.GenerateRandomToken(32)
		expiry := time.Now().Add(1 * time.Hour)

		if err := db.SetPasswordResetToken(user.ID, token, expiry); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate reset token"})
			return
		}

		// Send reset email
		if err := emailService.SendPasswordResetEmail(user.FirstName, user.Email, token); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send reset email"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link has been sent"})
	}
}

// ResetPasswordHandler handles password reset with token
func ResetPasswordHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ResetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := db.GetUserByResetToken(req.Token)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired reset token"})
			return
		}

		// Hash new password
		hash, err := services.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		// Update password and clear reset token
		if err := db.UpdateUserPassword(user.ID, hash); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
			return
		}

		if err := db.ClearPasswordResetToken(user.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear reset token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
	}
}

// VerifyEmailHandler handles email verification
func VerifyEmailHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req VerifyEmailRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := db.GetUserByVerificationToken(req.Token)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification token"})
			return
		}

		// Mark email as verified and clear token
		if err := db.SetUserEmailVerified(user.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email"})
			return
		}

		if err := db.ClearVerificationToken(user.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear verification token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
	}
}
