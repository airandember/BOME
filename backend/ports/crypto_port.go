package ports

import "time"

// CryptoPort defines the interface for cryptographic operations, JWT, and security utilities
type CryptoPort interface {
	// JWT Operations
	GenerateJWT(userID int, email, role string, verified bool) (string, error)
	GenerateTokenPair(userID int, email, role string, verified bool) (*TokenPair, error)
	ParseToken(tokenString string) (*Claims, error)
	ParseRefreshToken(tokenString string) (*Claims, error)
	RefreshTokenPair(refreshToken string) (*TokenPair, error)
	ValidateTokenClaims(claims *Claims) error
	BlacklistToken(token string) error

	// Password Operations
	HashPassword(password string) (string, error)
	CheckPassword(hash, password string) error
	ValidatePassword(password string) error

	// Encryption/Decryption
	EncryptString(plaintext string) (string, error)
	DecryptString(encrypted string) (string, error)

	// Validation & Sanitization
	ValidateEmail(email string) error
	ValidateName(name string) error
	SanitizeString(input string) string

	// Security Utilities
	GetClientIP(remoteAddr, xForwardedFor, xRealIP string) string
	GenerateSecureToken() string
	GenerateDeviceFingerprint(r interface{}) string

	// Rate Limiting (if exposed through interface)
	// Note: May be better as separate RateLimiterPort
}

// TokenPair represents an access + refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// Claims represents JWT token claims
type Claims struct {
	UserID        int    `json:"user_id"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"email_verified"`
	TokenID       string `json:"token_id,omitempty"`
	StandardClaims
}

// StandardClaims represents standard JWT claims
type StandardClaims struct {
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	Issuer    string `json:"iss"`
	Subject   string `json:"sub"`
}

// RateLimiterPort defines the interface for rate limiting operations
// Separated from CryptoPort for single responsibility
type RateLimiterPort interface {
	Allow(key string) bool
	CheckLoginAttempt(email, ip string) bool
	RecordFailedAttempt(email, ip string)
	RecordSuccessfulAttempt(email string)
	GetRemainingLockoutTime(email string) time.Duration
}
