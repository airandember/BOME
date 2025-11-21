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

// VerifyEmailLinkRequest represents email verification via URL link
type VerifyEmailLinkRequest struct {
	Token  string `form:"token" binding:"required"`
	UserID int    `form:"user_id"`
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

		// 🔄 UNIFIED FLOW: Check if user already exists
		// If they exist, send them verification email to complete setup
		existingUser, err := db.GetUserByEmail(req.Email)
		if err == nil && existingUser != nil {
			log.Printf("🔄 [UNIFIED-FLOW] Existing user attempting registration: %s (ID: %d)", existingUser.Email, existingUser.ID)

			// Generate new verification token for existing user
			verificationToken := services.GenerateSecureToken()
			if err := db.SetVerificationToken(existingUser.ID, verificationToken); err != nil {
				log.Printf("Failed to set verification token for existing user: %v", err)
			}

			// Send verification email
			if emailService != nil {
				fullName := existingUser.FirstName + " " + existingUser.LastName
				if err := emailService.SendVerificationEmail(existingUser.ID, existingUser.Email, fullName); err != nil {
					log.Printf("Failed to send verification email to existing user: %v", err)
				} else {
					log.Printf("✅ Verification email sent to existing user: %s", existingUser.Email)
				}
			}

			// 🔗 AUTO-LINK: Attempt to link any existing Stripe customers with matching email
			// (in case they created Stripe customer after initial registration)
			linkingService := services.NewCustomerLinkingService(db)
			linkResult, err := linkingService.LinkUserToCustomers(existingUser.ID)
			if err != nil {
				log.Printf("⚠️  Failed to auto-link Stripe customers for existing user %d: %v", existingUser.ID, err)
			} else if linkResult.CustomersLinked > 0 {
				log.Printf("✅ Auto-linked %d Stripe customer(s) to existing user %d (%s)", 
					linkResult.CustomersLinked, existingUser.ID, existingUser.Email)
			}

			// Return success message (same as new user to avoid revealing account existence)
			c.JSON(http.StatusCreated, gin.H{
				"message":               "Registration successful. Please check your email to complete your account setup.",
				"user_id":               existingUser.ID,
				"email":                 existingUser.Email,
				"verification_required": true,
			})
			return
		}

		// Create new user with empty password hash (will trigger password setup after verification)
		passwordHash := "" // Empty password triggers password setup flow
		user, err := db.CreateUser(req.Email, passwordHash, req.FirstName, req.LastName, "user")
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

		// Send verification email using new email service
		if emailService != nil {
			fullName := user.FirstName + " " + user.LastName
			if err := emailService.SendVerificationEmail(user.ID, user.Email, fullName); err != nil {
				log.Printf("Failed to send verification email: %v", err)
				// Continue anyway - user can request new verification
			} else {
				log.Printf("✅ Verification email sent to %s", user.Email)
			}
		}

	// Log successful registration
	log.Printf("User registered successfully: %s (ID: %d) from %s", user.Email, user.ID, clientIP)

	// 🔗 AUTO-LINK: Attempt to link any existing Stripe customers with matching email
	linkingService := services.NewCustomerLinkingService(db)
	linkResult, err := linkingService.LinkUserToCustomers(user.ID)
	if err != nil {
		log.Printf("⚠️  Failed to auto-link Stripe customers for new user %d: %v", user.ID, err)
	} else if linkResult.CustomersLinked > 0 {
		log.Printf("✅ Auto-linked %d Stripe customer(s) to new user %d (%s)", 
			linkResult.CustomersLinked, user.ID, user.Email)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":               "Registration successful. Please check your email to complete your account setup.",
		"user_id":               user.ID,
		"email":                 user.Email,
		"verification_required": true,
	})
	}
}

// RequestVerificationHandler handles verification email requests for existing users
func RequestVerificationHandler(db *database.DB, emailService *services.EmailService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required,email"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// Rate limiting
		clientIP := services.GetClientIP(c.Request.RemoteAddr, c.GetHeader("X-Forwarded-For"), c.GetHeader("X-Real-IP"))
		if !services.RegisterRateLimiter.Allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many verification requests. Please try again later.",
			})
			return
		}

		// Sanitize email
		req.Email = strings.ToLower(services.SanitizeString(req.Email))

		// Get user by email
		user, err := db.GetUserByEmail(req.Email)
		if err != nil {
			// Don't reveal if user exists or not for security
			c.JSON(http.StatusOK, gin.H{
				"message": "If an account with this email exists and requires verification, a verification email has been sent.",
			})
			return
		}

		// Check if already verified
		if user.EmailVerified {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Email address is already verified",
			})
			return
		}

		// Generate new verification token
		verificationToken := services.GenerateSecureToken()
		if err := db.SetVerificationToken(user.ID, verificationToken); err != nil {
			log.Printf("Failed to set verification token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate verification token",
			})
			return
		}

		// Send verification email
		if emailService != nil {
			fullName := user.FirstName + " " + user.LastName
			if err := emailService.SendVerificationEmail(user.ID, user.Email, fullName); err != nil {
				log.Printf("Failed to send verification email: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to send verification email",
				})
				return
			}
		}

		log.Printf("✅ Verification email sent to existing user: %s", user.Email)
		c.JSON(http.StatusOK, gin.H{
			"message": "Verification email sent successfully. Please check your email.",
		})
	}
}

// ResendVerificationHandler handles resending verification emails
func ResendVerificationHandler(db *database.DB, emailService *services.EmailService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required,email"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// Sanitize email
		req.Email = strings.ToLower(services.SanitizeString(req.Email))

		// Get user by email
		user, err := db.GetUserByEmail(req.Email)
		if err != nil {
			// Don't reveal if user exists or not for security
			c.JSON(http.StatusOK, gin.H{
				"message": "If an account with this email exists and is unverified, a verification email has been sent.",
			})
			return
		}

		// Check if already verified
		if user.EmailVerified {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Email address is already verified",
			})
			return
		}

		// Generate new verification token
		verificationToken := services.GenerateSecureToken()
		if err := db.SetVerificationToken(user.ID, verificationToken); err != nil {
			log.Printf("Failed to set verification token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate verification token",
			})
			return
		}

		// Send verification email
		if emailService != nil {
			fullName := user.FirstName + " " + user.LastName
			if err := emailService.SendVerificationEmail(user.ID, user.Email, fullName); err != nil {
				log.Printf("Failed to send verification email: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to send verification email",
				})
				return
			}
		}

		log.Printf("✅ Verification email resent to %s", user.Email)
		c.JSON(http.StatusOK, gin.H{
			"message": "Verification email sent successfully. Please check your email.",
		})
	}
}

// LoginHandler handles user login
func LoginHandler(db *database.DB, emailService *services.EmailService) gin.HandlerFunc {
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

		// 🔐 EMAIL VERIFICATION CHECK: Block login if email not verified AND no previous login
		if !user.EmailVerified && !user.LastLogin.Valid {
			log.Printf("🚫 Login blocked for unverified user: %s (ID: %d) - first-time login requires email verification", user.Email, user.ID)

			// 📧 AUTO-SEND VERIFICATION EMAIL: Send verification email automatically
			if emailService != nil {
				// Generate new verification token
				verificationToken := services.GenerateSecureToken()
				if err := db.SetVerificationToken(user.ID, verificationToken); err != nil {
					log.Printf("Failed to set verification token during login block: %v", err)
				} else {
					// Send verification email
					fullName := user.FirstName + " " + user.LastName
					if err := emailService.SendVerificationEmail(user.ID, user.Email, fullName); err != nil {
						log.Printf("Failed to send auto-verification email during login block: %v", err)
					} else {
						log.Printf("✅ Auto-sent verification email to blocked user: %s", user.Email)
					}
				}
			}

			// Log security event
			if db != nil {
				auditLog := &database.AuditLog{
					UserID:    &user.ID,
					UserEmail: &user.Email,
					Action:    "login_blocked",
					Resource:  "authentication",
					IPAddress: clientIP,
					UserAgent: c.GetHeader("User-Agent"),
					Status:    "warning",
					Details:   &[]string{"Login blocked - email verification required for first-time login"}[0],
					Severity:  "medium",
				}
				db.CreateAuditLog(auditLog)
			}

			c.JSON(http.StatusForbidden, gin.H{
				"error":                 "Email verification required",
				"message":               "Please verify your email address before logging in. A verification email has been sent to your inbox.",
				"verification_required": true,
				"user_id":               user.ID,
			})
			return
		}

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

		// 🔗 AUTO-LINK CHECK: On every login, check for new unlinked Stripe customers
		// This ensures users get linked even if they subscribed after their last login
		linkingService := services.NewCustomerLinkingService(db)
		linkResult, err := linkingService.LinkUserToCustomers(user.ID)
		if err != nil {
			log.Printf("⚠️  [LOGIN-LINK] Failed to check for unlinked customers for user %d: %v", user.ID, err)
		} else if linkResult.CustomersLinked > 0 {
			log.Printf("✅ [LOGIN-LINK] Auto-linked %d new Stripe customer(s) on login for user %d (%s)", 
				linkResult.CustomersLinked, user.ID, user.Email)
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

// getFrontendURL returns the appropriate frontend URL for redirects
func getFrontendURL(c *gin.Context) string {
	if strings.Contains(c.Request.Host, "localhost") {
		return "http://localhost:5173" // Development frontend URL
	}
	// Production - use the same domain as the request
	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s", scheme, c.Request.Host)
}

// VerifyEmailLinkHandler handles email verification via GET link (when user clicks email link)
func VerifyEmailLinkHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("🔍 [VERIFY-LINK] Handler called: %s", c.Request.URL.String())
		var req VerifyEmailLinkRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			log.Printf("❌ [VERIFY-LINK] Failed to bind query parameters: %v", err)
			// Redirect to frontend with error
			errorURL := fmt.Sprintf("%s/auth/verify-email?error=invalid_link", getFrontendURL(c))
			c.Redirect(http.StatusTemporaryRedirect, errorURL)
			return
		}

		log.Printf("🔍 [VERIFY-LINK] Processing token: %s, user_id: %d", req.Token[:8]+"...", req.UserID)

		// Check if database is available
		if db == nil {
			errorURL := fmt.Sprintf("%s/auth/verify-email?error=service_unavailable", getFrontendURL(c))
			c.Redirect(http.StatusTemporaryRedirect, errorURL)
			return
		}

		// Get user by verification token
		user, err := db.GetUserByVerificationToken(req.Token)
		if err != nil {
			log.Printf("❌ [VERIFY-LINK] Token lookup failed: %s (error: %v)", req.Token, err)

			// Let's also check what tokens exist in the database for debugging
			if debugUser, debugErr := db.GetUserByEmail("aarongusa@outlook.com"); debugErr == nil {
				log.Printf("🔍 [DEBUG] User %d current verification_token in DB: %s", debugUser.ID, debugUser.VerificationToken.String)
			}

			errorURL := fmt.Sprintf("%s/auth/verify-email?error=invalid_token", getFrontendURL(c))
			c.Redirect(http.StatusTemporaryRedirect, errorURL)
			return
		}

		// Optional: Verify user ID matches (extra security)
		if req.UserID > 0 && user.ID != req.UserID {
			log.Printf("User ID mismatch in verification: expected %d, got %d", user.ID, req.UserID)
			errorURL := fmt.Sprintf("%s/auth/verify-email?error=invalid_token", getFrontendURL(c))
			c.Redirect(http.StatusTemporaryRedirect, errorURL)
			return
		}

		// Set email as verified
		if err := db.SetUserEmailVerified(user.ID); err != nil {
			log.Printf("Failed to verify email: %v", err)
			errorURL := fmt.Sprintf("%s/auth/verify-email?error=verification_failed", getFrontendURL(c))
			c.Redirect(http.StatusTemporaryRedirect, errorURL)
			return
		}

		// Clear verification token
		if err := db.ClearVerificationToken(user.ID); err != nil {
			log.Printf("Failed to clear verification token: %v", err)
		}

	// Update last login to complete the verification process
	if err := db.UpdateLastLogin(user.ID); err != nil {
		log.Printf("Failed to update last login: %v", err)
	}

	// 🔗 AUTO-LINK: Attempt to link any existing Stripe customers with matching email
	// (safety net in case linking didn't happen at registration)
	linkingService := services.NewCustomerLinkingService(db)
	linkResult, err := linkingService.LinkUserToCustomers(user.ID)
	if err != nil {
		log.Printf("⚠️  Failed to auto-link Stripe customers during verification for user %d: %v", user.ID, err)
	} else if linkResult.CustomersLinked > 0 {
		log.Printf("✅ Auto-linked %d Stripe customer(s) during email verification for user %d (%s)", 
			linkResult.CustomersLinked, user.ID, user.Email)
	}

	log.Printf("✅ Email verified via link for: %s (ID: %d)", user.Email, user.ID)

		// 🔐 CHECK IF USER NEEDS PASSWORD SETUP
		needsPasswordSetup := user.PasswordHash == "" ||
			user.PasswordHash == "temp_hash" ||
			!user.PasswordChanged ||
			!user.LastLogin.Valid

		log.Printf("🔍 Password setup check for user %d: needsSetup=%v (hash_empty=%v, temp_hash=%v, not_changed=%v, no_login=%v)",
			user.ID, needsPasswordSetup,
			user.PasswordHash == "",
			user.PasswordHash == "temp_hash",
			!user.PasswordChanged,
			!user.LastLogin.Valid)

		if needsPasswordSetup {
			// Generate a temporary setup token for password setup
			setupToken := services.GenerateSecureToken()
			if err := db.SetPasswordSetupToken(user.ID, setupToken); err != nil {
				log.Printf("Failed to set password setup token: %v", err)
				// Fall back to normal success page
				successURL := fmt.Sprintf("%s/auth/verify-email?success=true", getFrontendURL(c))
				c.Redirect(http.StatusTemporaryRedirect, successURL)
				return
			}

			// Redirect to password setup page
			setupURL := fmt.Sprintf("%s/auth/setup-password?token=%s&user_id=%d",
				getFrontendURL(c), setupToken, user.ID)
			log.Printf("🔐 [PASSWORD-SETUP] Redirecting user %d to password setup: %s", user.ID, setupURL)
			c.Redirect(http.StatusTemporaryRedirect, setupURL)
			return
		}

		// Generate login tokens for auto-login (existing users with passwords)
		tokenPair, err := services.GenerateTokenPair(user.ID, user.Email, user.Role, true) // email is verified
		if err != nil {
			log.Printf("Failed to generate tokens after verification: %v", err)
			// Fall back to success page without auto-login
			successURL := fmt.Sprintf("%s/auth/verify-email?success=true", getFrontendURL(c))
			c.Redirect(http.StatusTemporaryRedirect, successURL)
			return
		}

		// Create session for tracking
		deviceInfo := services.GenerateDeviceFingerprint(c.Request)
		clientIP := services.GetClientIP(c.Request.RemoteAddr, c.GetHeader("X-Forwarded-For"), c.GetHeader("X-Real-IP"))

		// Extract token ID from refresh token for session tracking
		refreshClaims, _ := services.ParseRefreshToken(tokenPair.RefreshToken)
		tokenID := ""
		if refreshClaims != nil {
			tokenID = refreshClaims.TokenID
		}

		_, err = db.CreateSession(
			user.ID,
			tokenID,
			deviceInfo,
			clientIP,
			c.GetHeader("User-Agent"),
			time.Now().Add(7*24*time.Hour), // Session expires with refresh token
		)
		if err != nil {
			log.Printf("Failed to create session: %v", err)
		}

		// Redirect to frontend with auto-login tokens
		successURL := fmt.Sprintf("%s/auth/verify-email?success=true&access_token=%s&refresh_token=%s&user_id=%d",
			getFrontendURL(c), tokenPair.AccessToken, tokenPair.RefreshToken, user.ID)
		log.Printf("🔄 [VERIFY-LINK] Redirecting to frontend with auto-login: %s", successURL)
		c.Redirect(http.StatusTemporaryRedirect, successURL)
	}
}

// VerifyEmailHandler handles email verification via JSON API (for mobile apps/SPA)
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

// UpdateUserProfileRequest represents the user profile update payload
type UpdateUserProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Bio       string `json:"bio"`
	Location  string `json:"location"`
	Website   string `json:"website"`
	Phone     string `json:"phone"`
}

// GetCurrentUserHandler returns the current user's profile information
func GetCurrentUserHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by AuthRequired middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Get user from database
		user, err := db.GetUserByID(userID.(int))
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			log.Printf("Failed to get user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Service temporarily unavailable. Please try again later.",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user": gin.H{
				"id":             user.ID,
				"email":          user.Email,
				"first_name":     user.FirstName,
				"last_name":      user.LastName,
				"role":           user.Role,
				"email_verified": user.EmailVerified,
				"bio":            user.Bio,
				"location":       user.Location,
				"website":        user.Website,
				"phone":          user.Phone,
				"avatar_url":     user.AvatarURL,
				"created_at":     user.CreatedAt,
				"last_login":     user.LastLogin,
			},
		})
	}
}

// UpdateCurrentUserHandler updates the current user's profile information
func UpdateCurrentUserHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by AuthRequired middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req UpdateUserProfileRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Sanitize input
		req.FirstName = services.SanitizeString(req.FirstName)
		req.LastName = services.SanitizeString(req.LastName)
		req.Bio = services.SanitizeString(req.Bio)
		req.Location = services.SanitizeString(req.Location)
		req.Website = services.SanitizeString(req.Website)
		req.Phone = services.SanitizeString(req.Phone)

		// Validate names if provided
		if req.FirstName != "" {
			if err := services.ValidateName(req.FirstName); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid first name: " + err.Error()})
				return
			}
		}
		if req.LastName != "" {
			if err := services.ValidateName(req.LastName); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid last name: " + err.Error()})
				return
			}
		}

		// Update user profile
		updates := map[string]interface{}{
			"first_name": req.FirstName,
			"last_name":  req.LastName,
			"bio":        req.Bio,
			"location":   req.Location,
			"website":    req.Website,
			"phone":      req.Phone,
		}
		err := db.UpdateUserProfile(userID.(int), updates)
		if err != nil {
			log.Printf("Failed to update user profile: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Service temporarily unavailable. Please try again later.",
			})
			return
		}

		// Get updated user data
		user, err := db.GetUserByID(userID.(int))
		if err != nil {
			log.Printf("Failed to get updated user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Profile updated but failed to retrieve updated data.",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Profile updated successfully",
			"user": gin.H{
				"id":             user.ID,
				"email":          user.Email,
				"first_name":     user.FirstName,
				"last_name":      user.LastName,
				"role":           user.Role,
				"email_verified": user.EmailVerified,
				"bio":            user.Bio,
				"location":       user.Location,
				"website":        user.Website,
				"phone":          user.Phone,
				"avatar_url":     user.AvatarURL,
				"created_at":     user.CreatedAt,
				"last_login":     user.LastLogin,
			},
		})
	}
}

// SetupPasswordRequest represents the password setup request
type SetupPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required"`
	UserID   int    `json:"user_id,omitempty"`
}

// SetupPasswordHandler handles password setup for users without passwords (Stripe customers, etc.)
func SetupPasswordHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("🔐 [SETUP-PASSWORD] Handler called: %s", c.Request.URL.String())

		var req SetupPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("❌ [SETUP-PASSWORD] Failed to bind JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		log.Printf("🔐 [SETUP-PASSWORD] Processing request for token: %s", req.Token[:8]+"...")

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

		// Get user by password setup token
		user, err := db.GetUserByPasswordSetupToken(req.Token)
		if err != nil {
			log.Printf("Invalid password setup token: %s (error: %v)", req.Token, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired setup token"})
			return
		}

		// Optional: Verify user ID matches (extra security)
		if req.UserID > 0 && user.ID != req.UserID {
			log.Printf("User ID mismatch in password setup: expected %d, got %d", user.ID, req.UserID)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid setup token"})
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

		// Update password and mark as changed
		if err := db.UpdateUserPasswordWithChange(user.ID, hash); err != nil {
			log.Printf("Failed to update password: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Service temporarily unavailable. Please try again later.",
			})
			return
		}

		// Clear the setup token
		if err := db.ClearPasswordSetupToken(user.ID); err != nil {
			log.Printf("Failed to clear password setup token: %v", err)
		}

	// Update last login to complete the setup process
	if err := db.UpdateLastLogin(user.ID); err != nil {
		log.Printf("Failed to update last login: %v", err)
	}

	// 🔗 AUTO-LINK: Attempt to link any existing Stripe customers with matching email
	// (safety net in case linking didn't happen earlier in the flow)
	linkingService := services.NewCustomerLinkingService(db)
	linkResult, err := linkingService.LinkUserToCustomers(user.ID)
	if err != nil {
		log.Printf("⚠️  [SETUP-PASSWORD] Failed to auto-link Stripe customers for user %d: %v", user.ID, err)
	} else if linkResult.CustomersLinked > 0 {
		log.Printf("✅ [SETUP-PASSWORD] Auto-linked %d Stripe customer(s) during password setup for user %d (%s)", 
			linkResult.CustomersLinked, user.ID, user.Email)
	}

	log.Printf("✅ Password setup completed for: %s (ID: %d)", user.Email, user.ID)

	// Generate login tokens for auto-login
	tokenPair, err := services.GenerateTokenPair(user.ID, user.Email, user.Role, true) // email is verified
		if err != nil {
			log.Printf("Failed to generate tokens after password setup: %v", err)
			c.JSON(http.StatusOK, gin.H{
				"message": "Password setup successful. Please login with your new password.",
			})
			return
		}

		// Create session for tracking
		deviceInfo := services.GenerateDeviceFingerprint(c.Request)
		clientIP := services.GetClientIP(c.Request.RemoteAddr, c.GetHeader("X-Forwarded-For"), c.GetHeader("X-Real-IP"))

		// Extract token ID from refresh token for session tracking
		refreshClaims, _ := services.ParseRefreshToken(tokenPair.RefreshToken)
		tokenID := ""
		if refreshClaims != nil {
			tokenID = refreshClaims.TokenID
		}

		_, err = db.CreateSession(
			user.ID,
			tokenID,
			deviceInfo,
			clientIP,
			c.GetHeader("User-Agent"),
			time.Now().Add(7*24*time.Hour), // Session expires with refresh token
		)
		if err != nil {
			log.Printf("Failed to create session: %v", err)
		}

		// Return success with auto-login tokens
		c.JSON(http.StatusOK, gin.H{
			"message":       "Password setup successful! You are now logged in.",
			"access_token":  tokenPair.AccessToken,
			"refresh_token": tokenPair.RefreshToken,
			"expires_in":    tokenPair.ExpiresIn,
			"token_type":    "Bearer",
			"user": gin.H{
				"id":             user.ID,
				"email":          user.Email,
				"first_name":     user.FirstName,
				"last_name":      user.LastName,
				"role":           user.Role,
				"email_verified": true, // Always true after setup
			},
		})
	}
}
