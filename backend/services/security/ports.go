package security

// This file defines the ports (interfaces) for the security domain
// Implementations are in the crypto/ subdirectory

import (
	"time"
)

// CryptoPort defines the interface for cryptographic and security operations
// Implementation: services/security/crypto/service.go
type CryptoPort interface {
	// Encryption/Decryption
	EncryptString(plaintext string) (string, error)
	DecryptString(encryptedB64 string) (string, error)

	// Password Hashing/Verification
	HashPassword(password string) (string, error)
	CheckPassword(hash, password string) error
	ValidatePassword(password string) error

	// JWT Operations
	GenerateTokenPair(userID int, email, role string, emailVerified bool) (*TokenPair, error)
	ParseToken(tokenString string) (*Claims, error)
	ParseRefreshToken(tokenString string) (*Claims, error)
	RefreshTokenPair(refreshToken string) (*TokenPair, error)
	BlacklistToken(tokenString string) error
	ValidateTokenClaims(claims *Claims) error

	// Validation & Sanitization
	ValidateEmail(email string) error
	ValidateName(name string) error
	SanitizeString(input string) string

	// Security Utilities
	GetClientIP(remoteAddr, xForwardedFor, xRealIP string) string
	GenerateSecureToken() string
	GenerateDeviceFingerprint(r interface{}) string
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
type RateLimiterPort interface {
	Allow(key string) bool
	CheckLoginAttempt(email, ip string) bool
	RecordFailedAttempt(email, ip string)
	RecordSuccessfulAttempt(email string)
	GetRemainingLockoutTime(email string) time.Duration
}
