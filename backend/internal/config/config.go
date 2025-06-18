package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for the application
type Config struct {
	// Server Configuration
	ServerPort  string
	ServerHost  string
	Environment string
	Debug       bool

	// Database Configuration
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	// Redis Configuration
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// JWT Configuration
	JWTSecret        string
	JWTExpiry        string
	JWTRefreshExpiry string

	// CORS Configuration
	CORSAllowedOrigins []string
	CORSAllowedMethods []string
	CORSAllowedHeaders []string

	// Rate Limiting Configuration
	RateLimitRequests int
	RateLimitWindow   string
	RateLimitBurst    int

	// File Upload Configuration
	MaxFileSize         string
	AllowedVideoFormats []string
	AllowedImageFormats []string
	UploadPath          string

	// Logging Configuration
	LogLevel  string
	LogFormat string
	LogFile   string

	// Monitoring Configuration
	EnableMetrics       bool
	MetricsPort         string
	HealthCheckEndpoint string

	// Security Configuration
	BCryptCost    int
	SessionSecret string
	CSRFSecret    string

	// Admin Configuration
	AdminEmail     string
	AdminPassword  string
	AdminSecretKey string

	// Third-Party Service Configuration
	BunnyStorageZone   string
	BunnyAPIKey        string
	BunnyPullZone      string
	BunnyStreamLibrary string
	BunnyStreamAPIKey  string
	BunnyRegion        string
	BunnyWebhookSecret string

	StripeSecretKey         string
	StripePublishableKey    string
	StripeWebhookSecret     string
	StripePriceIDMonthly    string
	StripePriceIDYearly     string
	StripeCustomerPortalURL string

	DoSpacesKey         string
	DoSpacesSecret      string
	DoSpacesEndpoint    string
	DoSpacesBucket      string
	DoSpacesRegion      string
	DoSpacesCDNEndpoint string

	SendGridAPIKey                      string
	SendGridFromEmail                   string
	SendGridFromName                    string
	SendGridTemplateIDWelcome           string
	SendGridTemplateIDPasswordReset     string
	SendGridTemplateIDSubscription      string
	SendGridTemplateIDEmailVerification string

	// Roku App Configuration
	RokuAPIKey    string
	RokuSecretKey string
	RokuAppID     string
}

// New creates a new Config instance with values from environment variables
func New() *Config {
	return &Config{
		// Server Configuration
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		ServerHost:  getEnv("SERVER_HOST", "0.0.0.0"),
		Environment: getEnv("ENVIRONMENT", "development"),
		Debug:       getEnvBool("DEBUG", true),

		// Database Configuration
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "bome_streaming"),
		DBUser:     getEnv("DB_USER", "bome_user"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),

		// Redis Configuration
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		// JWT Configuration
		JWTSecret:        getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
		JWTExpiry:        getEnv("JWT_EXPIRY", "24h"),
		JWTRefreshExpiry: getEnv("JWT_REFRESH_EXPIRY", "168h"),

		// CORS Configuration
		CORSAllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173", "http://localhost:4173"}),
		CORSAllowedMethods: getEnvSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		CORSAllowedHeaders: getEnvSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization", "X-Requested-With"}),

		// Rate Limiting Configuration
		RateLimitRequests: getEnvInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   getEnv("RATE_LIMIT_WINDOW", "1m"),
		RateLimitBurst:    getEnvInt("RATE_LIMIT_BURST", 200),

		// File Upload Configuration
		MaxFileSize:         getEnv("MAX_FILE_SIZE", "500MB"),
		AllowedVideoFormats: getEnvSlice("ALLOWED_VIDEO_FORMATS", []string{"mp4", "avi", "mov", "wmv", "flv", "webm"}),
		AllowedImageFormats: getEnvSlice("ALLOWED_IMAGE_FORMATS", []string{"jpg", "jpeg", "png", "gif", "webp"}),
		UploadPath:          getEnv("UPLOAD_PATH", "./uploads"),

		// Logging Configuration
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),
		LogFile:   getEnv("LOG_FILE", "./logs/app.log"),

		// Monitoring Configuration
		EnableMetrics:       getEnvBool("ENABLE_METRICS", true),
		MetricsPort:         getEnv("METRICS_PORT", "9090"),
		HealthCheckEndpoint: getEnv("HEALTH_CHECK_ENDPOINT", "/health"),

		// Security Configuration
		BCryptCost:    getEnvInt("BCRYPT_COST", 12),
		SessionSecret: getEnv("SESSION_SECRET", "your-session-secret-key"),
		CSRFSecret:    getEnv("CSRF_SECRET", "your-csrf-secret-key"),

		// Admin Configuration
		AdminEmail:     getEnv("ADMIN_EMAIL", "admin@bookofmormonevidence.org"),
		AdminPassword:  getEnv("ADMIN_PASSWORD", "change_this_in_production"),
		AdminSecretKey: getEnv("ADMIN_SECRET_KEY", "your-admin-secret-key"),

		// Third-Party Service Configuration
		BunnyStorageZone:   getEnv("BUNNY_STORAGE_ZONE", ""),
		BunnyAPIKey:        getEnv("BUNNY_API_KEY", ""),
		BunnyPullZone:      getEnv("BUNNY_PULL_ZONE", ""),
		BunnyStreamLibrary: getEnv("BUNNY_STREAM_LIBRARY_ID", ""),
		BunnyStreamAPIKey:  getEnv("BUNNY_STREAM_API_KEY", ""),
		BunnyRegion:        getEnv("BUNNY_REGION", "de"),
		BunnyWebhookSecret: getEnv("BUNNY_WEBHOOK_SECRET", ""),

		StripeSecretKey:         getEnv("STRIPE_SECRET_KEY", ""),
		StripePublishableKey:    getEnv("STRIPE_PUBLISHABLE_KEY", ""),
		StripeWebhookSecret:     getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripePriceIDMonthly:    getEnv("STRIPE_PRICE_ID_MONTHLY", ""),
		StripePriceIDYearly:     getEnv("STRIPE_PRICE_ID_YEARLY", ""),
		StripeCustomerPortalURL: getEnv("STRIPE_CUSTOMER_PORTAL_URL", ""),

		DoSpacesKey:         getEnv("DO_SPACES_KEY", ""),
		DoSpacesSecret:      getEnv("DO_SPACES_SECRET", ""),
		DoSpacesEndpoint:    getEnv("DO_SPACES_ENDPOINT", "nyc3.digitaloceanspaces.com"),
		DoSpacesBucket:      getEnv("DO_SPACES_BUCKET", ""),
		DoSpacesRegion:      getEnv("DO_SPACES_REGION", "nyc3"),
		DoSpacesCDNEndpoint: getEnv("DO_SPACES_CDN_ENDPOINT", ""),

		SendGridAPIKey:                      getEnv("SENDGRID_API_KEY", ""),
		SendGridFromEmail:                   getEnv("SENDGRID_FROM_EMAIL", "noreply@bookofmormonevidence.org"),
		SendGridFromName:                    getEnv("SENDGRID_FROM_NAME", "Book of Mormon Evidences"),
		SendGridTemplateIDWelcome:           getEnv("SENDGRID_TEMPLATE_ID_WELCOME", ""),
		SendGridTemplateIDPasswordReset:     getEnv("SENDGRID_TEMPLATE_ID_PASSWORD_RESET", ""),
		SendGridTemplateIDSubscription:      getEnv("SENDGRID_TEMPLATE_ID_SUBSCRIPTION", ""),
		SendGridTemplateIDEmailVerification: getEnv("SENDGRID_TEMPLATE_ID_EMAIL_VERIFICATION", ""),

		// Roku App Configuration
		RokuAPIKey:    getEnv("ROKU_API_KEY", ""),
		RokuSecretKey: getEnv("ROKU_SECRET_KEY", ""),
		RokuAppID:     getEnv("ROKU_APP_ID", ""),
	}
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
