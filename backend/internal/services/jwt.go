package services

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte
var jwtRefreshSecret []byte
var secretsInitialized bool

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

	// Generate access token (short-lived: 15 minutes)
	accessToken, err := generateToken(userID, email, role, emailVerified, "access", 15*time.Minute, jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token (long-lived: 7 days)
	refreshToken, err := generateToken(userID, email, role, emailVerified, "refresh", 7*24*time.Hour, jwtRefreshSecret)
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
	return generateToken(userID, email, role, true, "access", expiry, jwtSecret)
}

// generateToken internal helper to generate tokens
func generateToken(userID int, email, role string, emailVerified bool, tokenType string, expiry time.Duration, secret []byte) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:        userID,
		Email:         email,
		Role:          role,
		EmailVerified: emailVerified,
		TokenType:     tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "bome-backend",
			Subject:   fmt.Sprintf("user:%d", userID),
			ID:        fmt.Sprintf("%d_%s_%d", userID, tokenType, now.Unix()),
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

	return claims, nil
}

// RefreshTokenPair generates new tokens from a valid refresh token
func RefreshTokenPair(refreshToken string) (*TokenPair, error) {
	claims, err := ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Generate new token pair
	return GenerateTokenPair(claims.UserID, claims.Email, claims.Role, claims.EmailVerified)
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
