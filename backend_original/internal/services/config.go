package services

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ConfigService handles application configuration
type ConfigService struct {
	// Database configuration
	DatabaseURL      string
	DatabaseHost     string
	DatabasePort     int
	DatabaseName     string
	DatabaseUser     string
	DatabasePassword string

	// Server configuration
	ServerPort  int
	ServerHost  string
	Environment string

	// Stripe configuration
	StripeEnabled           bool
	StripeSecretKey         string
	StripePublishableKey    string
	StripeWebhookSecret     string
	StripeEnvironment       string
	StripePriceIDMonthly    string
	StripePriceIDYearly     string
	StripeCustomerPortalURL string

	// Email configuration
	EmailEnabled  bool
	EmailHost     string
	EmailPort     int
	EmailUsername string
	EmailPassword string
	EmailFrom     string

	// Analytics configuration
	AnalyticsEnabled  bool
	AnalyticsEndpoint string

	// Security configuration
	JWTSecret     string
	SessionSecret string
	CORSOrigins   []string

	// Feature flags
	Features map[string]bool
}

// NewConfigService creates a new configuration service
func NewConfigService() *ConfigService {
	config := &ConfigService{
		Features: make(map[string]bool),
	}

	// Load all configuration
	config.loadDatabaseConfig()
	config.loadServerConfig()
	config.loadStripeConfig()
	config.loadEmailConfig()
	config.loadAnalyticsConfig()
	config.loadSecurityConfig()
	config.loadFeatureFlags()

	return config
}

// loadDatabaseConfig loads database configuration
func (c *ConfigService) loadDatabaseConfig() {
	c.DatabaseURL = c.getEnv("DATABASE_URL", "")
	c.DatabaseHost = c.getEnv("DB_HOST", "localhost")
	c.DatabasePort = c.getEnvAsInt("DB_PORT", 5432)
	c.DatabaseName = c.getEnv("DB_NAME", "bome")
	c.DatabaseUser = c.getEnv("DB_USER", "postgres")
	c.DatabasePassword = c.getEnv("DB_PASSWORD", "")
}

// loadServerConfig loads server configuration
func (c *ConfigService) loadServerConfig() {
	c.ServerPort = c.getEnvAsInt("SERVER_PORT", 8080)
	c.ServerHost = c.getEnv("SERVER_HOST", "0.0.0.0")
	c.Environment = c.getEnv("ENVIRONMENT", "development")
}

// loadStripeConfig loads Stripe configuration
func (c *ConfigService) loadStripeConfig() {
	c.StripeSecretKey = c.getEnv("STRIPE_SECRET_KEY", "")
	c.StripePublishableKey = c.getEnv("STRIPE_PUBLISHABLE_KEY", "")
	c.StripeWebhookSecret = c.getEnv("STRIPE_WEBHOOK_SECRET", "")
	c.StripeEnvironment = c.getEnv("STRIPE_ENVIRONMENT", "test")
	c.StripePriceIDMonthly = c.getEnv("STRIPE_PRICE_ID_MONTHLY", "")
	c.StripePriceIDYearly = c.getEnv("STRIPE_PRICE_ID_YEARLY", "")
	c.StripeCustomerPortalURL = c.getEnv("STRIPE_CUSTOMER_PORTAL_URL", "")

	// Determine if Stripe is enabled
	c.StripeEnabled = c.StripeSecretKey != "" && c.StripePublishableKey != ""
}

// loadEmailConfig loads email configuration
func (c *ConfigService) loadEmailConfig() {
	c.EmailEnabled = c.getEnvAsBool("EMAIL_ENABLED", false)
	c.EmailHost = c.getEnv("EMAIL_HOST", "")
	c.EmailPort = c.getEnvAsInt("EMAIL_PORT", 587)
	c.EmailUsername = c.getEnv("EMAIL_USERNAME", "")
	c.EmailPassword = c.getEnv("EMAIL_PASSWORD", "")
	c.EmailFrom = c.getEnv("EMAIL_FROM", "")
}

// loadAnalyticsConfig loads analytics configuration
func (c *ConfigService) loadAnalyticsConfig() {
	c.AnalyticsEnabled = c.getEnvAsBool("ANALYTICS_ENABLED", true)
	c.AnalyticsEndpoint = c.getEnv("ANALYTICS_ENDPOINT", "")
}

// loadSecurityConfig loads security configuration
func (c *ConfigService) loadSecurityConfig() {
	c.JWTSecret = c.getEnv("JWT_SECRET", "")
	c.SessionSecret = c.getEnv("SESSION_SECRET", "")

	// Load CORS origins
	corsOrigins := c.getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:5173")
	c.CORSOrigins = strings.Split(corsOrigins, ",")
}

// loadFeatureFlags loads feature flags
func (c *ConfigService) loadFeatureFlags() {
	c.Features["subscription_system"] = c.getEnvAsBool("FEATURE_SUBSCRIPTION_SYSTEM", true)
	c.Features["stripe_integration"] = c.getEnvAsBool("FEATURE_STRIPE_INTEGRATION", true)
	c.Features["analytics"] = c.getEnvAsBool("FEATURE_ANALYTICS", true)
	c.Features["email_notifications"] = c.getEnvAsBool("FEATURE_EMAIL_NOTIFICATIONS", false)
	c.Features["admin_dashboard"] = c.getEnvAsBool("FEATURE_ADMIN_DASHBOARD", true)
}

// IsFeatureEnabled checks if a feature is enabled
func (c *ConfigService) IsFeatureEnabled(feature string) bool {
	if enabled, exists := c.Features[feature]; exists {
		return enabled
	}
	return false
}

// IsStripeEnabled returns whether Stripe is properly configured
func (c *ConfigService) IsStripeEnabled() bool {
	return c.StripeEnabled
}

// IsEmailEnabled returns whether email is properly configured
func (c *ConfigService) IsEmailEnabled() bool {
	return c.EmailEnabled
}

// IsAnalyticsEnabled returns whether analytics is properly configured
func (c *ConfigService) IsAnalyticsEnabled() bool {
	return c.AnalyticsEnabled
}

// GetStripeConfig returns Stripe configuration
func (c *ConfigService) GetStripeConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled":             c.StripeEnabled,
		"environment":         c.StripeEnvironment,
		"publishable_key":     c.StripePublishableKey,
		"webhook_secret":      c.StripeWebhookSecret != "",
		"price_id_monthly":    c.StripePriceIDMonthly,
		"price_id_yearly":     c.StripePriceIDYearly,
		"customer_portal_url": c.StripeCustomerPortalURL,
	}
}

// GetDatabaseConfig returns database configuration
func (c *ConfigService) GetDatabaseConfig() map[string]interface{} {
	return map[string]interface{}{
		"host":     c.DatabaseHost,
		"port":     c.DatabasePort,
		"name":     c.DatabaseName,
		"user":     c.DatabaseUser,
		"password": c.DatabasePassword != "",
		"url":      c.DatabaseURL != "",
	}
}

// GetServerConfig returns server configuration
func (c *ConfigService) GetServerConfig() map[string]interface{} {
	return map[string]interface{}{
		"host":        c.ServerHost,
		"port":        c.ServerPort,
		"environment": c.Environment,
	}
}

// ValidateConfig validates the configuration
func (c *ConfigService) ValidateConfig() error {
	var errors []string

	// Validate required database configuration
	if c.DatabaseURL == "" && (c.DatabaseHost == "" || c.DatabaseName == "" || c.DatabaseUser == "") {
		errors = append(errors, "database configuration is incomplete")
	}

	// Validate JWT secret
	if c.JWTSecret == "" {
		errors = append(errors, "JWT_SECRET is required")
	}

	// Validate session secret
	if c.SessionSecret == "" {
		errors = append(errors, "SESSION_SECRET is required")
	}

	// Validate Stripe configuration if enabled
	if c.StripeEnabled {
		if c.StripeSecretKey == "" {
			errors = append(errors, "STRIPE_SECRET_KEY is required when Stripe is enabled")
		}
		if c.StripePublishableKey == "" {
			errors = append(errors, "STRIPE_PUBLISHABLE_KEY is required when Stripe is enabled")
		}
	}

	// Validate email configuration if enabled
	if c.EmailEnabled {
		if c.EmailHost == "" {
			errors = append(errors, "EMAIL_HOST is required when email is enabled")
		}
		if c.EmailUsername == "" {
			errors = append(errors, "EMAIL_USERNAME is required when email is enabled")
		}
		if c.EmailPassword == "" {
			errors = append(errors, "EMAIL_PASSWORD is required when email is enabled")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// GetConfigSummary returns a summary of the configuration
func (c *ConfigService) GetConfigSummary() map[string]interface{} {
	return map[string]interface{}{
		"environment": c.Environment,
		"features":    c.Features,
		"services": map[string]interface{}{
			"stripe":    c.IsStripeEnabled(),
			"email":     c.IsEmailEnabled(),
			"analytics": c.IsAnalyticsEnabled(),
		},
		"database": c.GetDatabaseConfig(),
		"server":   c.GetServerConfig(),
	}
}

// Helper methods for environment variable handling
func (c *ConfigService) getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *ConfigService) getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func (c *ConfigService) getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// GetEnvironment returns the current environment
func (c *ConfigService) GetEnvironment() string {
	return c.Environment
}

// IsDevelopment returns whether the application is running in development mode
func (c *ConfigService) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns whether the application is running in production mode
func (c *ConfigService) IsProduction() bool {
	return c.Environment == "production"
}

// IsTest returns whether the application is running in test mode
func (c *ConfigService) IsTest() bool {
	return c.Environment == "test"
}
