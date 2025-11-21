package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"bome-backend/internal/database"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuth2Provider represents different OAuth2 providers
type OAuth2Provider string

const (
	ProviderGoogle  OAuth2Provider = "google"
	ProviderGeneric OAuth2Provider = "generic"
)

// OAuth2Config holds configuration for OAuth2 providers
type OAuth2Config struct {
	Google  *oauth2.Config
	Generic map[string]*oauth2.Config // For multiple generic providers
}

// OAuth2State represents the state parameter for CSRF protection
type OAuth2State struct {
	State     string    `json:"state"`
	Provider  string    `json:"provider"`
	ReturnURL string    `json:"return_url"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// OAuth2UserInfo represents user information from OAuth2 providers
type OAuth2UserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Picture       string `json:"picture"`
	EmailVerified bool   `json:"email_verified"`
	Provider      string `json:"provider"`
}

// OAuth2Service handles OAuth2 authentication
type OAuth2Service struct {
	db     *database.DB
	config *OAuth2Config
}

// NewOAuth2Service creates a new OAuth2 service
func NewOAuth2Service(db *database.DB) *OAuth2Service {
	service := &OAuth2Service{
		db: db,
		config: &OAuth2Config{
			Generic: make(map[string]*oauth2.Config),
		},
	}

	// Initialize Google OAuth2 config
	service.initializeGoogleConfig()

	return service
}

// initializeGoogleConfig sets up Google OAuth2 configuration
func (s *OAuth2Service) initializeGoogleConfig() {
	// Get Google OAuth2 settings from database
	clientID, err := s.db.GetEmailSetting("google_oauth_client_id")
	if err != nil {
		log.Printf("⚠️ [OAUTH2] Google OAuth2 client ID not configured: %v", err)
		return
	}

	clientSecret, err := s.db.GetEmailSetting("google_oauth_client_secret")
	if err != nil {
		log.Printf("⚠️ [OAUTH2] Google OAuth2 client secret not configured: %v", err)
		return
	}

	// Use environment variable for OAuth2 redirect URL
	redirectURL := os.Getenv("OAUTH2_REDIRECT_URL")
	if redirectURL == "" {
		log.Printf("⚠️ [OAUTH2] OAUTH2_REDIRECT_URL not set! OAuth2 will not work in production!")
		redirectURL = "http://localhost:5173/auth/oauth2/callback" // Development fallback only
	}

	// Decrypt client secret if encrypted
	if s.db != nil {
		// Try to decrypt (assuming it might be encrypted)
		// In a real implementation, you'd check if it's encrypted first
	}

	s.config.Google = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	log.Printf("✅ [OAUTH2] Google OAuth2 configured successfully")
}

// AddGenericProvider adds a generic OAuth2 provider
func (s *OAuth2Service) AddGenericProvider(name, clientID, clientSecret, redirectURL, authURL, tokenURL string, scopes []string) {
	s.config.Generic[name] = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
	}

	log.Printf("✅ [OAUTH2] Generic provider '%s' configured successfully", name)
}

// GenerateAuthURL generates an OAuth2 authorization URL with state parameter
func (s *OAuth2Service) GenerateAuthURL(provider OAuth2Provider, returnURL string) (string, error) {
	// Generate secure state parameter
	state, err := s.generateState()
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}

	// Store state for validation in database
	stateData := OAuth2State{
		State:     state,
		Provider:  string(provider),
		ReturnURL: returnURL,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(10 * time.Minute), // State expires in 10 minutes
	}

	// Store state in database for persistence across server restarts
	if err := s.storeOAuth2State(stateData); err != nil {
		return "", fmt.Errorf("failed to store OAuth2 state: %w", err)
	}

	// Get appropriate OAuth2 config
	var config *oauth2.Config
	switch provider {
	case ProviderGoogle:
		if s.config.Google == nil {
			return "", fmt.Errorf("google OAuth2 not configured")
		}
		config = s.config.Google
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}

	// Generate authorization URL
	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	log.Printf("🔗 [OAUTH2] Generated auth URL for %s provider", provider)
	return authURL, nil
}

// HandleCallback processes OAuth2 callback and exchanges code for tokens
func (s *OAuth2Service) HandleCallback(code, state string) (*OAuth2UserInfo, error) {
	// Validate state parameter from database
	stateData, err := s.getOAuth2State(state)
	if err != nil {
		return nil, fmt.Errorf("invalid state parameter: %w", err)
	}

	// Check if state has expired
	if time.Now().After(stateData.ExpiresAt) {
		s.deleteOAuth2State(state) // Clean up expired state
		return nil, fmt.Errorf("state parameter expired")
	}

	// Remove state from store (one-time use)
	if err := s.deleteOAuth2State(state); err != nil {
		log.Printf("Warning: failed to delete OAuth2 state: %v", err)
	}

	// Get appropriate OAuth2 config
	var config *oauth2.Config
	provider := OAuth2Provider(stateData.Provider)

	switch provider {
	case ProviderGoogle:
		config = s.config.Google
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	// Exchange authorization code for tokens
	ctx := context.Background()
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Get user information
	userInfo, err := s.getUserInfo(provider, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	userInfo.Provider = string(provider)

	log.Printf("✅ [OAUTH2] Successfully processed callback for %s provider", provider)
	return userInfo, nil
}

// getUserInfo fetches user information from the OAuth2 provider
func (s *OAuth2Service) getUserInfo(provider OAuth2Provider, token *oauth2.Token) (*OAuth2UserInfo, error) {
	switch provider {
	case ProviderGoogle:
		return s.getGoogleUserInfo(token)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// getGoogleUserInfo fetches user information from Google
func (s *OAuth2Service) getGoogleUserInfo(token *oauth2.Token) (*OAuth2UserInfo, error) {
	client := s.config.Google.Client(context.Background(), token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info from Google: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google API returned status: %d", resp.StatusCode)
	}

	var googleUser struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
		EmailVerified bool   `json:"verified_email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, fmt.Errorf("failed to decode Google user info: %w", err)
	}

	return &OAuth2UserInfo{
		ID:            googleUser.ID,
		Email:         googleUser.Email,
		Name:          googleUser.Name,
		FirstName:     googleUser.GivenName,
		LastName:      googleUser.FamilyName,
		Picture:       googleUser.Picture,
		EmailVerified: googleUser.EmailVerified,
	}, nil
}

// CreateOrLinkUser creates a new user or links OAuth2 account to existing user
func (s *OAuth2Service) CreateOrLinkUser(userInfo *OAuth2UserInfo) (*database.User, bool, error) {
	// Check if user already exists by email
	existingUser, err := s.db.GetUserByEmail(userInfo.Email)
	if err == nil && existingUser != nil {
		// User exists - link OAuth2 account
		err = s.linkOAuth2Account(existingUser.ID, userInfo)
		if err != nil {
			log.Printf("⚠️ [OAUTH2] Failed to link OAuth2 account: %v", err)
			// Continue anyway - user can still login
		}

		// Update email verification status if OAuth2 provider verified it
		if userInfo.EmailVerified && !existingUser.EmailVerified {
			err = s.db.SetUserEmailVerified(existingUser.ID)
			if err != nil {
				log.Printf("⚠️ [OAUTH2] Failed to update email verification status: %v", err)
			} else {
				existingUser.EmailVerified = true
			}
		}

		// 🔗 AUTO-LINK: Attempt to link any existing Stripe customers with matching email
		// (in case they created Stripe customer before OAuth2 login)
		log.Printf("🔍 [OAUTH2-LINK] Starting auto-link check for existing user %d (%s)", existingUser.ID, existingUser.Email)
		
		linkingService := NewCustomerLinkingService(s.db)
		linkResult, err := linkingService.LinkUserToCustomers(existingUser.ID)
		
		// Detailed logging for debugging
		log.Printf("🔍 [OAUTH2-LINK] Link result: CustomersFound=%d, CustomersLinked=%d, Error=%s", 
			linkResult.CustomersFound, linkResult.CustomersLinked, linkResult.Error)
		
		if err != nil {
			log.Printf("❌ [OAUTH2] Failed to auto-link Stripe customers for existing user %d: %v", existingUser.ID, err)
		} else if linkResult.CustomersLinked > 0 {
			log.Printf("✅ [OAUTH2] Auto-linked %d Stripe customer(s) to existing user %d (%s)", 
				linkResult.CustomersLinked, existingUser.ID, existingUser.Email)
		} else if linkResult.CustomersFound > 0 {
			log.Printf("⚠️  [OAUTH2] Found %d customers for user %d but linked %d - Skipped: %v",
				linkResult.CustomersFound, existingUser.ID, linkResult.CustomersLinked, linkResult.SkippedCustomers)
		} else {
			log.Printf("ℹ️  [OAUTH2] No Stripe customers found for user %d (%s) - may not have subscribed yet", existingUser.ID, existingUser.Email)
		}

		log.Printf("🔗 [OAUTH2] Linked %s account to existing user: %s", userInfo.Provider, userInfo.Email)
		return existingUser, false, nil // false = not a new user
	}

	// User doesn't exist - create new user
	// OAuth2 users don't have passwords initially, so use a placeholder
	passwordHash := "" // Will be empty for OAuth2-only users

	// Create user in database
	createdUser, err := s.db.CreateUser(userInfo.Email, passwordHash, userInfo.FirstName, userInfo.LastName, "user")
	if err != nil {
		return nil, false, fmt.Errorf("failed to create user: %w", err)
	}

	// Set email as verified if OAuth2 provider verified it
	if userInfo.EmailVerified {
		err = s.db.SetUserEmailVerified(createdUser.ID)
		if err != nil {
			log.Printf("⚠️ [OAUTH2] Failed to set email as verified for new user: %v", err)
		} else {
			createdUser.EmailVerified = true
		}
	}

	// Link OAuth2 account to new user
	err = s.linkOAuth2Account(createdUser.ID, userInfo)
	if err != nil {
		log.Printf("⚠️ [OAUTH2] Failed to link OAuth2 account to new user: %v", err)
		// Continue anyway - user is created
	}

	// 🔗 AUTO-LINK: Attempt to link any existing Stripe customers with matching email
	linkingService := NewCustomerLinkingService(s.db)
	linkResult, err := linkingService.LinkUserToCustomers(createdUser.ID)
	if err != nil {
		log.Printf("⚠️ [OAUTH2] Failed to auto-link Stripe customers for new user %d: %v", createdUser.ID, err)
	} else if linkResult.CustomersLinked > 0 {
		log.Printf("✅ [OAUTH2] Auto-linked %d Stripe customer(s) to new user %d (%s)", 
			linkResult.CustomersLinked, createdUser.ID, createdUser.Email)
	}

	log.Printf("👤 [OAUTH2] Created new user from %s: %s", userInfo.Provider, userInfo.Email)
	return createdUser, true, nil // true = new user
}

// linkOAuth2Account links an OAuth2 account to a user
func (s *OAuth2Service) linkOAuth2Account(userID int, userInfo *OAuth2UserInfo) error {
	// Store OAuth2 account linking in database
	// This would typically go in an oauth2_accounts table
	query := `
		INSERT INTO oauth2_accounts (user_id, provider, provider_user_id, email, name, picture, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (user_id, provider) 
		DO UPDATE SET 
			provider_user_id = EXCLUDED.provider_user_id,
			email = EXCLUDED.email,
			name = EXCLUDED.name,
			picture = EXCLUDED.picture,
			updated_at = NOW()`

	_, err := s.db.DB.Exec(query, userID, userInfo.Provider, userInfo.ID, userInfo.Email, userInfo.Name, userInfo.Picture)
	return err
}

// generateState generates a cryptographically secure state parameter
func (s *OAuth2Service) generateState() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// CleanupExpiredStates removes expired state parameters (call periodically)
func (s *OAuth2Service) CleanupExpiredStates() {
	if err := s.cleanupExpiredOAuth2States(); err != nil {
		log.Printf("⚠️ [OAUTH2] Failed to cleanup expired states: %v", err)
	}
}

// IsConfigured checks if OAuth2 providers are properly configured
func (s *OAuth2Service) IsConfigured(provider OAuth2Provider) bool {
	switch provider {
	case ProviderGoogle:
		return s.config.Google != nil && s.config.Google.ClientID != ""
	default:
		return false
	}
}

// GetConfiguredProviders returns a list of configured OAuth2 providers
func (s *OAuth2Service) GetConfiguredProviders() []string {
	var providers []string

	if s.IsConfigured(ProviderGoogle) {
		providers = append(providers, "google")
	}

	for name := range s.config.Generic {
		providers = append(providers, name)
	}

	return providers
}

// storeOAuth2State stores OAuth2 state in database for persistence
func (s *OAuth2Service) storeOAuth2State(state OAuth2State) error {
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	query := `
		INSERT INTO oauth2_states (state, provider, return_url, created_at, expires_at, state_data)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (state) DO UPDATE SET
			provider = EXCLUDED.provider,
			return_url = EXCLUDED.return_url,
			created_at = EXCLUDED.created_at,
			expires_at = EXCLUDED.expires_at,
			state_data = EXCLUDED.state_data
	`

	_, err = s.db.DB.Exec(query, state.State, state.Provider, state.ReturnURL,
		state.CreatedAt, state.ExpiresAt, string(stateJSON))
	if err != nil {
		return fmt.Errorf("failed to store OAuth2 state: %w", err)
	}

	return nil
}

// getOAuth2State retrieves OAuth2 state from database
func (s *OAuth2Service) getOAuth2State(stateParam string) (*OAuth2State, error) {
	var state OAuth2State
	var stateJSON string

	query := `
		SELECT state, provider, return_url, created_at, expires_at, state_data
		FROM oauth2_states 
		WHERE state = $1 AND expires_at > NOW()
	`

	err := s.db.DB.QueryRow(query, stateParam).Scan(
		&state.State, &state.Provider, &state.ReturnURL,
		&state.CreatedAt, &state.ExpiresAt, &stateJSON)
	if err != nil {
		return nil, fmt.Errorf("OAuth2 state not found or expired: %w", err)
	}

	return &state, nil
}

// deleteOAuth2State removes OAuth2 state from database
func (s *OAuth2Service) deleteOAuth2State(stateParam string) error {
	query := `DELETE FROM oauth2_states WHERE state = $1`

	_, err := s.db.DB.Exec(query, stateParam)
	if err != nil {
		return fmt.Errorf("failed to delete OAuth2 state: %w", err)
	}

	return nil
}

// cleanupExpiredOAuth2States removes expired OAuth2 states (should be called periodically)
func (s *OAuth2Service) cleanupExpiredOAuth2States() error {
	query := `DELETE FROM oauth2_states WHERE expires_at <= NOW()`

	result, err := s.db.DB.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired OAuth2 states: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("🧹 [OAUTH2] Cleaned up %d expired OAuth2 states", rowsAffected)
	}

	return nil
}
