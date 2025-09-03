package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"bome-backend/internal/database"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
)

// StripePublicService handles public Stripe operations using publishable keys
type StripePublicService struct {
	db             *database.DB
	publishableKey string
	isEnabled      bool
}

// NewStripePublicService creates a new public Stripe service instance
func NewStripePublicService(db *database.DB) *StripePublicService {
	service := &StripePublicService{
		db:        db,
		isEnabled: false,
	}

	// Load publishable key from public_settings
	if err := service.loadPublicSettings(); err != nil {
		log.Printf("⚠️ [STRIPE-PUBLIC] Failed to load public settings: %v", err)
	}

	return service
}

// loadPublicSettings loads the publishable key from public_settings table
func (s *StripePublicService) loadPublicSettings() error {
	publishableKey, err := s.db.GetPublicSetting("stripe_publishable_key")
	if err != nil {
		return fmt.Errorf("failed to get publishable key: %w", err)
	}

	if publishableKey == "" {
		return errors.New("stripe publishable key not configured")
	}

	s.publishableKey = publishableKey
	s.isEnabled = true

	log.Printf("✅ [STRIPE-PUBLIC] Service initialized with publishable key: %s...", publishableKey[:12])
	return nil
}

// GetPublishableKey returns the Stripe publishable key for frontend use
func (s *StripePublicService) GetPublishableKey() (string, error) {
	if !s.isEnabled {
		if err := s.loadPublicSettings(); err != nil {
			return "", fmt.Errorf("stripe public service not available: %w", err)
		}
	}

	return s.publishableKey, nil
}

// GetStripeConfig returns the public Stripe configuration for frontend
func (s *StripePublicService) GetStripeConfig() (map[string]interface{}, error) {
	publishableKey, err := s.GetPublishableKey()
	if err != nil {
		return nil, err
	}

	// Get portal URL from public settings (optional)
	portalURL, _ := s.db.GetPublicSetting("stripe_portal_url")

	config := map[string]interface{}{
		"publishable_key": publishableKey,
	}

	if portalURL != "" {
		config["portal_url"] = portalURL
	}

	return config, nil
}

// CreateEmbeddedCheckoutSession creates a Stripe embedded checkout session for public use
func (s *StripePublicService) CreateEmbeddedCheckoutSession(planID, returnURL, userID string) (string, error) {
	if !s.isEnabled {
		return "", errors.New("Stripe public service is not enabled")
	}

	log.Printf("🔍 [STRIPE-PUBLIC] Creating embedded checkout session for plan %s, user %s", planID, userID)

	// Get the encrypted secret key from secure settings for creating sessions
	// (We need the secret key to create sessions, but we'll return the client secret for public use)
	encryptedKey, err := s.db.GetSecureSetting("stripe_secret_key")
	if err != nil || encryptedKey == "" {
		return "", errors.New("stripe secret key not configured")
	}

	// Decrypt the key
	log.Printf("🔍 [STRIPE-PUBLIC] Encrypted key length: %d", len(encryptedKey))
	cryptoService := GetGlobalCryptoService()
	if cryptoService == nil {
		log.Printf("❌ [STRIPE-PUBLIC] Crypto service not available")
		return "", errors.New("crypto service not available")
	}

	secretKey, err := cryptoService.DecryptString(encryptedKey)
	if err != nil {
		log.Printf("❌ [STRIPE-PUBLIC] Failed to decrypt key: %v", err)
		return "", fmt.Errorf("failed to decrypt stripe key: %w", err)
	}

	log.Printf("✅ [STRIPE-PUBLIC] Successfully decrypted key, length: %d", len(secretKey))
	log.Printf("🔍 [STRIPE-PUBLIC] Decrypted key starts with: %s", secretKey[:12])

	// Set the Stripe API key for this operation
	stripe.Key = secretKey

	// Look up actual plan details from database
	log.Printf("🔍 [STRIPE-PUBLIC] Looking up plan details for plan ID: %s", planID)

	// Get plan from database
	planQuery := `SELECT stripe_price_id, name, price, currency FROM subscription_plans WHERE id = $1 AND is_active = true`
	var stripePriceID, planName, currency string
	var price float64

	err = s.db.DB.QueryRow(planQuery, planID).Scan(&stripePriceID, &planName, &price, &currency)
	if err != nil {
		log.Printf("❌ [STRIPE-PUBLIC] Failed to get plan details: %v", err)
		return "", fmt.Errorf("plan not found or inactive: %w", err)
	}

	log.Printf("🔍 [STRIPE-PUBLIC] Found plan: %s, Price ID: %s, Price: %.2f %s", planName, stripePriceID, price, currency)

	// Validate stripe_price_id exists
	if stripePriceID == "" {
		log.Printf("❌ [STRIPE-PUBLIC] Plan %s has no Stripe price ID configured", planName)
		return "", errors.New("plan is not configured with Stripe - please contact support")
	}

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(stripePriceID),
				Quantity: stripe.Int64(1),
			},
		},
		CustomerEmail: stripe.String(fmt.Sprintf("user%s@bome.test", userID)), // TODO: Get actual user email from database
	}

	// Set UI mode for embedded checkout
	params.AddExtra("ui_mode", "embedded")
	params.AddExtra("return_url", returnURL)

	// Add metadata
	params.AddMetadata("user_id", userID)
	params.AddMetadata("plan_id", planID)

	sess, err := session.New(params)
	if err != nil {
		log.Printf("❌ [STRIPE-PUBLIC] Failed to create checkout session: %v", err)
		return "", fmt.Errorf("failed to create checkout session: %w", err)
	}

	log.Printf("✅ [STRIPE-PUBLIC] Checkout session created: %s", sess.ID)

	// For embedded checkout, we need to return the client secret
	// The client secret is what Stripe.js needs to mount the embedded checkout
	// In stripe-go v74, we need to access it through the LastResponse.RawJSON
	if sess.LastResponse == nil || sess.LastResponse.RawJSON == nil {
		log.Printf("❌ [STRIPE-PUBLIC] No raw JSON in session response")
		return "", errors.New("checkout session missing raw response data")
	}

	var sessionData map[string]interface{}
	if err := json.Unmarshal(sess.LastResponse.RawJSON, &sessionData); err != nil {
		log.Printf("❌ [STRIPE-PUBLIC] Failed to parse session JSON: %v", err)
		return "", fmt.Errorf("failed to parse session response: %w", err)
	}

	clientSecret, exists := sessionData["client_secret"]
	if !exists {
		log.Printf("❌ [STRIPE-PUBLIC] No client_secret in session data")
		return "", errors.New("checkout session missing client secret for embedded mode")
	}

	clientSecretStr, ok := clientSecret.(string)
	if !ok {
		log.Printf("❌ [STRIPE-PUBLIC] Client secret is not a string")
		return "", errors.New("invalid client secret format")
	}

	log.Printf("✅ [STRIPE-PUBLIC] Returning client secret: %s...", clientSecretStr[:20])
	return clientSecretStr, nil
}

// GetCustomerPortalURL creates a customer portal session and returns the URL
func (s *StripePublicService) GetCustomerPortalURL(customerID, returnURL string) (string, error) {
	if !s.isEnabled {
		return "", errors.New("Stripe public service is not enabled")
	}

	// Get the encrypted secret key for creating portal sessions
	encryptedKey, err := s.db.GetSecureSetting("stripe_secret_key")
	if err != nil || encryptedKey == "" {
		return "", errors.New("stripe secret key not configured")
	}

	// Decrypt the key
	cryptoService := GetGlobalCryptoService()
	if cryptoService == nil {
		return "", errors.New("crypto service not available")
	}

	secretKey, err := cryptoService.DecryptString(encryptedKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt stripe key: %w", err)
	}

	// Set the Stripe API key
	stripe.Key = secretKey

	// TODO: Implement customer portal session creation
	// This would use the Stripe customer portal API

	log.Printf("🔍 [STRIPE-PUBLIC] Creating customer portal session for customer %s", customerID)

	// For now, return the configured portal URL from public settings
	portalURL, err := s.db.GetPublicSetting("stripe_portal_url")
	if err != nil || portalURL == "" {
		return "", errors.New("stripe portal URL not configured")
	}

	return portalURL, nil
}

// IsEnabled returns whether the public Stripe service is properly configured
func (s *StripePublicService) IsEnabled() bool {
	return s.isEnabled
}

// RefreshSettings reloads the public settings from database
func (s *StripePublicService) RefreshSettings() error {
	return s.loadPublicSettings()
}
