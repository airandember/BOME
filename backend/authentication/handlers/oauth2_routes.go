package handlers

import (
	"net/http"
	"strings"
	"time"

	authModels "bome-backend/authentication/models"
	authServices "bome-backend/authentication/services"
	"bome-backend/infrastructure/database"
	"bome-backend/services/security/crypto"

	"github.com/gin-gonic/gin"
)

// OAuth2LoginRequest represents an OAuth2 login initiation request
type OAuth2LoginRequest struct {
	Provider  string `json:"provider" binding:"required"`
	ReturnURL string `json:"return_url"`
}

// OAuth2CallbackRequest represents an OAuth2 callback request
type OAuth2CallbackRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}

// OAuth2ProvidersResponse represents available OAuth2 providers
type OAuth2ProvidersResponse struct {
	Providers []OAuth2ProviderInfo `json:"providers"`
}

// OAuth2ProviderInfo represents information about an OAuth2 provider
type OAuth2ProviderInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Icon        string `json:"icon"`
	Enabled     bool   `json:"enabled"`
}

// SetupOAuth2Routes sets up OAuth2 authentication routes
func SetupOAuth2Routes(router *gin.RouterGroup, db *database.DB, oauth2Service *authServices.OAuth2Service) {
	oauth := router.Group("/oauth2")
	{
		// Get available OAuth2 providers
		oauth.GET("/providers", func(c *gin.Context) {
			getOAuth2Providers(c, oauth2Service)
		})

		// Initiate OAuth2 login
		oauth.POST("/login", func(c *gin.Context) {
			initiateOAuth2Login(c, oauth2Service)
		})

		// Handle OAuth2 callback
		oauth.POST("/callback", func(c *gin.Context) {
			handleOAuth2Callback(c, db, oauth2Service)
		})

		// Get OAuth2 account links for authenticated user
		oauth.GET("/accounts", func(c *gin.Context) {
			getLinkedOAuth2Accounts(c, db)
		})

		// Unlink OAuth2 account
		oauth.DELETE("/accounts/:provider", func(c *gin.Context) {
			unlinkOAuth2Account(c, db)
		})
	}

	// Admin routes for OAuth2 configuration
	adminOAuth := router.Group("/admin/oauth2")
	{
		// Get OAuth2 configuration
		adminOAuth.GET("/config", func(c *gin.Context) {
			getOAuth2Config(c, db)
		})

		// Update OAuth2 configuration
		adminOAuth.PUT("/config", func(c *gin.Context) {
			updateOAuth2Config(c, db, oauth2Service)
		})

		// Test OAuth2 provider configuration
		adminOAuth.POST("/test/:provider", func(c *gin.Context) {
			testOAuth2Provider(c, oauth2Service)
		})
	}
}

// getOAuth2Providers returns available OAuth2 providers
func getOAuth2Providers(c *gin.Context, oauth2Service *authServices.OAuth2Service) {
	providers := []OAuth2ProviderInfo{
		{
			Name:        "google",
			DisplayName: "Google",
			Icon:        "🔍",
			Enabled:     oauth2Service.IsConfigured(authServices.ProviderGoogle),
		},
	}

	// Add configured generic providers
	configuredProviders := oauth2Service.GetConfiguredProviders()
	for _, provider := range configuredProviders {
		if provider != "google" {
			providers = append(providers, OAuth2ProviderInfo{
				Name:        provider,
				DisplayName: strings.Title(provider),
				Icon:        "🔗",
				Enabled:     true,
			})
		}
	}

	c.JSON(http.StatusOK, OAuth2ProvidersResponse{
		Providers: providers,
	})
}

// initiateOAuth2Login starts the OAuth2 login flow
func initiateOAuth2Login(c *gin.Context, oauth2Service *authServices.OAuth2Service) {
	var req OAuth2LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate provider
	provider := authServices.OAuth2Provider(req.Provider)
	if !oauth2Service.IsConfigured(provider) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "OAuth2 provider not configured: " + req.Provider,
		})
		return
	}

	// Generate authorization URL
	authURL, err := oauth2Service.GenerateAuthURL(provider, req.ReturnURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate authorization URL",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"provider": req.Provider,
	})
}

// handleOAuth2Callback processes OAuth2 callback
func handleOAuth2Callback(c *gin.Context, db *database.DB, oauth2Service *authServices.OAuth2Service) {
	var req OAuth2CallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Handle OAuth2 callback
	userInfo, err := oauth2Service.HandleCallback(req.Code, req.State)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "OAuth2 callback failed: " + err.Error(),
		})
		return
	}

	// Create or link user account
	user, isNewUser, err := oauth2Service.CreateOrLinkUser(userInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create or link user account",
		})
		return
	}

	// Generate JWT tokens for the user
	tokenPair, err := crypto.GetGlobalCryptoService().GenerateTokenPair(user.ID, user.Email, user.Role, user.EmailVerified)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate authentication tokens",
		})
		return
	}

	// Create session record
	clientIP := crypto.GetGlobalCryptoService().GetClientIP(c.Request.RemoteAddr, c.GetHeader("X-Forwarded-For"), c.GetHeader("X-Real-IP"))
	deviceInfo := crypto.GetGlobalCryptoService().GenerateDeviceFingerprint(c.Request)

	// Extract token ID from refresh token for session tracking
	refreshClaims, _ := crypto.GetGlobalCryptoService().ParseRefreshToken(tokenPair.RefreshToken)
	tokenID := ""
	if refreshClaims != nil {
		tokenID = refreshClaims.TokenID
	}

	_, err = authModels.CreateSession(db,
		user.ID,
		tokenID,
		deviceInfo,
		clientIP,
		c.GetHeader("User-Agent"),
		time.Now().Add(7*24*time.Hour), // Session expires with refresh token (7 days)
	)
	if err != nil {
		// Log error but don't fail the login
		c.Header("X-Session-Warning", "Failed to create session record")
	}

	// Update last login
	if err := authModels.UpdateLastLogin(db, user.ID); err != nil {
		// Log error but don't fail the login
		c.Header("X-Login-Warning", "Failed to update last login")
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_in":    tokenPair.ExpiresIn,
		"token_type":    tokenPair.TokenType,
		"user": gin.H{
			"id":             user.ID,
			"email":          user.Email,
			"first_name":     user.FirstName,
			"last_name":      user.LastName,
			"role":           user.Role,
			"email_verified": user.EmailVerified,
		},
		"is_new_user": isNewUser,
		"provider":    userInfo.Provider,
	})
}

// getLinkedOAuth2Accounts returns OAuth2 accounts linked to the current user
func getLinkedOAuth2Accounts(c *gin.Context, db *database.DB) {
	// This would require authentication middleware
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	query := `
		SELECT provider, email, name, picture, created_at
		FROM oauth2_accounts 
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := db.DB.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get linked accounts"})
		return
	}
	defer rows.Close()

	var accounts []gin.H
	for rows.Next() {
		var provider, email, name, picture string
		var createdAt string

		err := rows.Scan(&provider, &email, &name, &picture, &createdAt)
		if err != nil {
			continue
		}

		accounts = append(accounts, gin.H{
			"provider":   provider,
			"email":      email,
			"name":       name,
			"picture":    picture,
			"created_at": createdAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts": accounts,
	})
}

// unlinkOAuth2Account removes OAuth2 account link
func unlinkOAuth2Account(c *gin.Context, db *database.DB) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider parameter required"})
		return
	}

	query := `DELETE FROM oauth2_accounts WHERE user_id = $1 AND provider = $2`
	result, err := db.DB.Exec(query, userID, provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlink account"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "OAuth2 account not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "OAuth2 account unlinked successfully",
		"provider": provider,
	})
}

// getOAuth2Config returns OAuth2 configuration (admin only)
func getOAuth2Config(c *gin.Context, db *database.DB) {
	// Get OAuth2 settings
	settings := make(map[string]string)

	query := `
		SELECT setting_key, setting_value 
		FROM email_settings 
		WHERE setting_key LIKE 'google_oauth_%' OR setting_key LIKE 'oauth2_%'`

	rows, err := db.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get OAuth2 configuration"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}

		// Don't return encrypted secrets
		if strings.Contains(key, "secret") {
			settings[key] = "[ENCRYPTED]"
		} else {
			settings[key] = value
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"settings": settings,
	})
}

// updateOAuth2Config updates OAuth2 configuration (admin only)
func updateOAuth2Config(c *gin.Context, db *database.DB, oauth2Service *authServices.OAuth2Service) {
	var request map[string]string
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Update allowed OAuth2 settings
	allowedSettings := []string{
		"google_oauth_enabled", "google_oauth_client_id", "google_oauth_client_secret",
		"google_oauth_redirect_url", "oauth2_state_ttl_minutes", "oauth2_auto_link_accounts",
		"oauth2_auto_verify_email",
	}

	for _, key := range allowedSettings {
		if value, exists := request[key]; exists {
			err := db.SetEmailSetting(key, value)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to update setting: " + key,
				})
				return
			}
		}
	}

	// Reinitialize OAuth2 service with new settings
	// In a real implementation, you might want to reload the service

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "OAuth2 configuration updated successfully",
	})
}

// testOAuth2Provider tests OAuth2 provider configuration
func testOAuth2Provider(c *gin.Context, oauth2Service *authServices.OAuth2Service) {
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider parameter required"})
		return
	}

	oauth2Provider := authServices.OAuth2Provider(provider)
	isConfigured := oauth2Service.IsConfigured(oauth2Provider)

	c.JSON(http.StatusOK, gin.H{
		"provider":   provider,
		"configured": isConfigured,
		"status": map[string]interface{}{
			"google": oauth2Service.IsConfigured(authServices.ProviderGoogle),
		},
	})
}
