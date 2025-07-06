package services

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte
var jwtRefreshSecret []byte
var secretsInitialized bool

// TokenBlacklist manages blacklisted tokens
type TokenBlacklist struct {
	tokens map[string]time.Time
	mutex  sync.RWMutex
}

// Global token blacklist instance
var tokenBlacklist = &TokenBlacklist{
	tokens: make(map[string]time.Time),
}

// BlacklistToken adds a token to the blacklist
func (tb *TokenBlacklist) BlacklistToken(tokenID string, expiry time.Time) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	tb.tokens[tokenID] = expiry
}

// IsBlacklisted checks if a token is blacklisted
func (tb *TokenBlacklist) IsBlacklisted(tokenID string) bool {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	if expiry, exists := tb.tokens[tokenID]; exists {
		if time.Now().Before(expiry) {
			return true
		}
		// Clean up expired token
		tb.mutex.RUnlock()
		tb.mutex.Lock()
		delete(tb.tokens, tokenID)
		tb.mutex.Unlock()
		tb.mutex.RLock()
	}
	return false
}

// CleanupExpiredTokens removes expired tokens from blacklist
func (tb *TokenBlacklist) CleanupExpiredTokens() {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	now := time.Now()
	for tokenID, expiry := range tb.tokens {
		if now.After(expiry) {
			delete(tb.tokens, tokenID)
		}
	}
}

// GetBlacklistSize returns the number of blacklisted tokens
func (tb *TokenBlacklist) GetBlacklistSize() int {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return len(tb.tokens)
}

// initializeSecrets ensures JWT secrets are loaded from environment variables
func initializeSecrets() error {
	if secretsInitialized {
		return nil
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return errors.New("JWT_SECRET environment variable is required")
	}
	jwtSecret = []byte(secret)

	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if refreshSecret == "" {
		refreshSecret = secret + "_refresh"
	}
	jwtRefreshSecret = []byte(refreshSecret)

	secretsInitialized = true
	return nil
}

// Claims represents the JWT claims
type Claims struct {
	UserID        int    `json:"user_id"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"email_verified"`
	TokenType     string `json:"token_type"` // "access" or "refresh"
	TokenID       string `json:"token_id"`   // Unique token identifier for blacklisting
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// GenerateTokenPair generates both access and refresh tokens
func GenerateTokenPair(userID int, email, role string, emailVerified bool) (*TokenPair, error) {
	if err := initializeSecrets(); err != nil {
		return nil, fmt.Errorf("failed to initialize JWT secrets: %w", err)
	}

	// Generate unique token IDs
	accessTokenID := fmt.Sprintf("access_%d_%s", userID, time.Now().Format("20060102150405"))
	refreshTokenID := fmt.Sprintf("refresh_%d_%s", userID, time.Now().Format("20060102150405"))

	// Generate access token (short-lived: 15 minutes)
	accessToken, err := generateToken(userID, email, role, emailVerified, "access", 15*time.Minute, jwtSecret, accessTokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token (long-lived: 7 days)
	refreshToken, err := generateToken(userID, email, role, emailVerified, "refresh", 7*24*time.Hour, jwtRefreshSecret, refreshTokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64((15 * time.Minute).Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// GenerateToken generates a JWT for a user (backward compatibility)
func GenerateToken(userID int, email, role string, expiry time.Duration) (string, error) {
	if err := initializeSecrets(); err != nil {
		return "", fmt.Errorf("failed to initialize JWT secrets: %w", err)
	}
	tokenID := fmt.Sprintf("legacy_%d_%s", userID, time.Now().Format("20060102150405"))
	return generateToken(userID, email, role, true, "access", expiry, jwtSecret, tokenID)
}

// generateToken internal helper to generate tokens
func generateToken(userID int, email, role string, emailVerified bool, tokenType string, expiry time.Duration, secret []byte, tokenID string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:        userID,
		Email:         email,
		Role:          role,
		EmailVerified: emailVerified,
		TokenType:     tokenType,
		TokenID:       tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "bome-backend",
			Subject:   fmt.Sprintf("user:%d", userID),
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ParseToken parses and validates a JWT
func ParseToken(tokenString string) (*Claims, error) {
	if err := initializeSecrets(); err != nil {
		return nil, fmt.Errorf("failed to initialize JWT secrets: %w", err)
	}

	if tokenString == "" {
		return nil, errors.New("token is required")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Validate token type
	if claims.TokenType != "access" {
		return nil, errors.New("invalid token type for access")
	}

	// Validate issuer
	if claims.Issuer != "bome-backend" {
		return nil, errors.New("invalid token issuer")
	}

	// Check if token is blacklisted
	if tokenBlacklist.IsBlacklisted(claims.TokenID) {
		return nil, errors.New("token has been revoked")
	}

	return claims, nil
}

// ParseRefreshToken parses and validates a refresh token
func ParseRefreshToken(tokenString string) (*Claims, error) {
	if err := initializeSecrets(); err != nil {
		return nil, fmt.Errorf("failed to initialize JWT secrets: %w", err)
	}

	if tokenString == "" {
		return nil, errors.New("refresh token is required")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtRefreshSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	// Validate token type
	if claims.TokenType != "refresh" {
		return nil, errors.New("invalid token type for refresh")
	}

	// Validate issuer
	if claims.Issuer != "bome-backend" {
		return nil, errors.New("invalid token issuer")
	}

	// Check if token is blacklisted
	if tokenBlacklist.IsBlacklisted(claims.TokenID) {
		return nil, errors.New("refresh token has been revoked")
	}

	return claims, nil
}

// RefreshTokenPair generates new tokens from a valid refresh token
func RefreshTokenPair(refreshToken string) (*TokenPair, error) {
	claims, err := ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Blacklist the old refresh token
	if claims.TokenID != "" {
		tokenBlacklist.BlacklistToken(claims.TokenID, claims.ExpiresAt.Time)
	}

	// Generate new token pair
	return GenerateTokenPair(claims.UserID, claims.Email, claims.Role, claims.EmailVerified)
}

// BlacklistToken adds a token to the blacklist
func BlacklistToken(tokenString string) error {
	// Try to parse as access token first
	claims, err := ParseToken(tokenString)
	if err != nil {
		// Try as refresh token
		claims, err = ParseRefreshToken(tokenString)
		if err != nil {
			return fmt.Errorf("invalid token: %w", err)
		}
	}

	if claims.TokenID != "" {
		tokenBlacklist.BlacklistToken(claims.TokenID, claims.ExpiresAt.Time)
	}

	return nil
}

// ValidateTokenClaims performs additional validation on token claims
func ValidateTokenClaims(claims *Claims) error {
	if claims.UserID <= 0 {
		return errors.New("invalid user ID in token")
	}

	if claims.Email == "" {
		return errors.New("invalid email in token")
	}

	if claims.Role == "" {
		return errors.New("invalid role in token")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return errors.New("token has expired")
	}

	// Check if token is not yet valid
	if claims.NotBefore != nil && claims.NotBefore.After(time.Now()) {
		return errors.New("token not yet valid")
	}

	return nil
}

// StartTokenBlacklistCleanup starts a background goroutine to clean up expired tokens
func StartTokenBlacklistCleanup() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour) // Clean up every hour
		defer ticker.Stop()

		for range ticker.C {
			tokenBlacklist.CleanupExpiredTokens()
		}
	}()
}
