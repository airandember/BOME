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
	Google     *oauth2.Config
	Generic    map[string]*oauth2.Config // For multiple generic providers
	StateStore map[string]OAuth2State    // In-memory state store (use Redis in production)
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
			StateStore: make(map[string]OAuth2State),
			Generic:    make(map[string]*oauth2.Config),
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
		redirectURL = "http://localhost:5173/auth/oauth2/callback" // Default fallback matching frontend route
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

	// Store state for validation
	stateData := OAuth2State{
		State:     state,
		Provider:  string(provider),
		ReturnURL: returnURL,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(10 * time.Minute), // State expires in 10 minutes
	}
	s.config.StateStore[state] = stateData

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
	// Validate state parameter
	stateData, exists := s.config.StateStore[state]
	if !exists {
		return nil, fmt.Errorf("invalid state parameter")
	}

	// Check if state has expired
	if time.Now().After(stateData.ExpiresAt) {
		delete(s.config.StateStore, state)
		return nil, fmt.Errorf("state parameter expired")
	}

	// Remove state from store (one-time use)
	delete(s.config.StateStore, state)

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
	now := time.Now()
	for state, stateData := range s.config.StateStore {
		if now.After(stateData.ExpiresAt) {
			delete(s.config.StateStore, state)
		}
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
