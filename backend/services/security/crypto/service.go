package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"bome-backend/services/security"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Ensure CryptoService implements security.CryptoPort
var _ security.CryptoPort = (*CryptoService)(nil)

// CryptoService provides all cryptographic and security operations
type CryptoService struct {
	// Encryption
	encryptionKey []byte

	// JWT
	jwtSecret        []byte
	jwtRefreshSecret []byte
	tokenBlacklist   *TokenBlacklist

	// Rate Limiting
	registerLimiter *RateLimiter
	loginLimiter    *EnhancedRateLimiter
	passwordLimiter *RateLimiter
}

// Config for CryptoService initialization
type Config struct {
	EncryptionKey    string
	JWTSecret        string
	JWTRefreshSecret string
}

// NewCryptoService creates a new CryptoService instance
func NewCryptoService(config *Config) (*CryptoService, error) {
	// Derive encryption key
	key := sha256.Sum256([]byte(config.EncryptionKey))

	return &CryptoService{
		encryptionKey:    key[:],
		jwtSecret:        []byte(config.JWTSecret),
		jwtRefreshSecret: []byte(config.JWTRefreshSecret),
		tokenBlacklist: &TokenBlacklist{
			tokens: make(map[string]time.Time),
		},
		registerLimiter: NewRateLimiter(5, 15*time.Minute),
		loginLimiter:    NewEnhancedRateLimiter(5, 15*time.Minute, 1*time.Hour),
		passwordLimiter: NewRateLimiter(3, 15*time.Minute),
	}, nil
}

// NewCryptoServiceFromEnv creates a CryptoService from environment variables
func NewCryptoServiceFromEnv() (*CryptoService, error) {
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		return nil, errors.New("ENCRYPTION_KEY environment variable is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET environment variable is required")
	}

	jwtRefreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if jwtRefreshSecret == "" {
		return nil, errors.New("JWT_REFRESH_SECRET environment variable is required")
	}

	return NewCryptoService(&Config{
		EncryptionKey:    encryptionKey,
		JWTSecret:        jwtSecret,
		JWTRefreshSecret: jwtRefreshSecret,
	})
}

// =============================================================================
// ENCRYPTION METHODS
// =============================================================================

// EncryptString encrypts a plaintext string using AES-GCM
func (c *CryptoService) EncryptString(plaintext string) (string, error) {
	block, err := aes.NewCipher(c.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptString decrypts a base64-encoded encrypted string
func (c *CryptoService) DecryptString(encryptedB64 string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedB64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(c.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// =============================================================================
// JWT METHODS
// =============================================================================

// GenerateJWT generates a JWT token (deprecated - use GenerateTokenPair)
func (c *CryptoService) GenerateJWT(userID int, email, role string, verified bool) (string, error) {
	return c.generateToken(userID, email, role, verified, "access", 15*time.Minute, c.jwtSecret, "")
}

// GenerateTokenPair generates both access and refresh tokens
func (c *CryptoService) GenerateTokenPair(userID int, email, role string, emailVerified bool) (*security.TokenPair, error) {
	tokenID := GenerateSecureToken()

	accessToken, err := c.generateToken(userID, email, role, emailVerified, "access", 15*time.Minute, c.jwtSecret, tokenID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := c.generateToken(userID, email, role, emailVerified, "refresh", 7*24*time.Hour, c.jwtRefreshSecret, tokenID)
	if err != nil {
		return nil, err
	}

	return &security.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 minutes in seconds
		TokenType:    "Bearer",
	}, nil
}

// generateToken is the internal token generation method
func (c *CryptoService) generateToken(userID int, email, role string, emailVerified bool, tokenType string, expiry time.Duration, secret []byte, tokenID string) (string, error) {
	claims := &security.Claims{
		UserID:        userID,
		Email:         email,
		Role:          role,
		EmailVerified: emailVerified,
		TokenID:       tokenID,
		StandardClaims: security.StandardClaims{
			ExpiresAt: time.Now().Add(expiry).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "bome-backend",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":        claims.UserID,
		"email":          claims.Email,
		"role":           claims.Role,
		"email_verified": claims.EmailVerified,
		"token_id":       claims.TokenID,
		"exp":            claims.StandardClaims.ExpiresAt,
		"iat":            claims.StandardClaims.IssuedAt,
		"iss":            claims.StandardClaims.Issuer,
		"sub":            claims.StandardClaims.Subject,
	})

	return token.SignedString(secret)
}

// ParseToken parses and validates an access token
func (c *CryptoService) ParseToken(tokenString string) (*security.Claims, error) {
	return c.parseTokenWithSecret(tokenString, c.jwtSecret)
}

// ParseRefreshToken parses and validates a refresh token
func (c *CryptoService) ParseRefreshToken(tokenString string) (*security.Claims, error) {
	return c.parseTokenWithSecret(tokenString, c.jwtRefreshSecret)
}

// parseTokenWithSecret is the internal token parsing method
func (c *CryptoService) parseTokenWithSecret(tokenString string, secret []byte) (*security.Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Safe type assertions with error checking
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			return nil, errors.New("invalid user_id claim")
		}

		email, ok := claims["email"].(string)
		if !ok {
			return nil, errors.New("invalid email claim")
		}

		role, ok := claims["role"].(string)
		if !ok {
			return nil, errors.New("invalid role claim")
		}

		emailVerified, ok := claims["email_verified"].(bool)
		if !ok {
			return nil, errors.New("invalid email_verified claim")
		}

		tokenID, _ := claims["token_id"].(string) // Optional field

		expFloat, ok := claims["exp"].(float64)
		if !ok {
			return nil, errors.New("invalid exp claim")
		}

		iatFloat, ok := claims["iat"].(float64)
		if !ok {
			return nil, errors.New("invalid iat claim")
		}

		issuer, ok := claims["iss"].(string)
		if !ok {
			return nil, errors.New("invalid iss claim")
		}

		subject, ok := claims["sub"].(string)
		if !ok {
			return nil, errors.New("invalid sub claim")
		}

		result := &security.Claims{
			UserID:        int(userIDFloat),
			Email:         email,
			Role:          role,
			EmailVerified: emailVerified,
			TokenID:       tokenID,
			StandardClaims: security.StandardClaims{
				ExpiresAt: int64(expFloat),
				IssuedAt:  int64(iatFloat),
				Issuer:    issuer,
				Subject:   subject,
			},
		}

		// Check if token is blacklisted
		if tokenID != "" && c.tokenBlacklist.IsBlacklisted(tokenID) {
			return nil, errors.New("token has been revoked")
		}

		return result, nil
	}

	return nil, errors.New("invalid token claims")
}

// RefreshTokenPair creates a new token pair from a valid refresh token
func (c *CryptoService) RefreshTokenPair(refreshToken string) (*security.TokenPair, error) {
	claims, err := c.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Validate token hasn't expired
	if err := c.ValidateTokenClaims(claims); err != nil {
		return nil, err
	}

	// Generate new token pair
	return c.GenerateTokenPair(claims.UserID, claims.Email, claims.Role, claims.EmailVerified)
}

// BlacklistToken adds a token to the blacklist
func (c *CryptoService) BlacklistToken(tokenString string) error {
	claims, err := c.ParseRefreshToken(tokenString)
	if err != nil {
		// Try parsing as access token
		claims, err = c.ParseToken(tokenString)
		if err != nil {
			return err
		}
	}

	if claims.TokenID != "" {
		expiry := time.Unix(claims.StandardClaims.ExpiresAt, 0)
		c.tokenBlacklist.BlacklistToken(claims.TokenID, expiry)
	}

	return nil
}

// ValidateTokenClaims validates token claims
func (c *CryptoService) ValidateTokenClaims(claims *security.Claims) error {
	if claims == nil {
		return errors.New("claims cannot be nil")
	}

	if time.Now().Unix() > claims.StandardClaims.ExpiresAt {
		return errors.New("token has expired")
	}

	return nil
}

// =============================================================================
// PASSWORD METHODS
// =============================================================================

// HashPassword hashes a plain password using bcrypt
func (c *CryptoService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword compares a plain password with a hash
func (c *CryptoService) CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// ValidatePassword validates password strength with enhanced security
func (c *CryptoService) ValidatePassword(password string) error {
	if len(password) < 12 {
		return fmt.Errorf("password must be at least 12 characters long")
	}

	if len(password) > 128 {
		return fmt.Errorf("password too long")
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	uniqueChars := make(map[rune]bool)

	for _, char := range password {
		uniqueChars[char] = true
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	if len(uniqueChars) < 8 {
		return fmt.Errorf("password must contain at least 8 unique characters")
	}

	if hasRepeatingPatterns(password) {
		return fmt.Errorf("password contains repeating patterns")
	}

	if isCommonPassword(password) {
		return fmt.Errorf("password is too common, please choose a stronger password")
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}

	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// =============================================================================
// VALIDATION & SANITIZATION METHODS
// =============================================================================

// ValidateEmail validates email format
func (c *CryptoService) ValidateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	return nil
}

// ValidateName validates name format
func (c *CryptoService) ValidateName(name string) error {
	if len(name) < 1 {
		return errors.New("name cannot be empty")
	}
	if len(name) > 100 {
		return errors.New("name too long")
	}
	return nil
}

// SanitizeString removes potentially dangerous characters
func (c *CryptoService) SanitizeString(input string) string {
	return strings.TrimSpace(input)
}

// =============================================================================
// UTILITY METHODS
// =============================================================================

// GetClientIP extracts the real client IP from request
func (c *CryptoService) GetClientIP(remoteAddr, xForwardedFor, xRealIP string) string {
	if xRealIP != "" {
		return xRealIP
	}

	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
		return remoteAddr[:idx]
	}

	return remoteAddr
}

// GenerateDeviceFingerprint generates a device fingerprint from HTTP request
func (c *CryptoService) GenerateDeviceFingerprint(r interface{}) string {
	req, ok := r.(*http.Request)
	if !ok {
		return "unknown"
	}

	fingerprint := req.UserAgent() + "|" + req.Header.Get("Accept-Language")
	hash := sha256.Sum256([]byte(fingerprint))
	return base64.URLEncoding.EncodeToString(hash[:])[:32]
}

// GenerateSecureToken generates a secure random token
func (c *CryptoService) GenerateSecureToken() string {
	bytes := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

// =============================================================================
// GLOBAL INSTANCE PATTERN (for backward compatibility)
// =============================================================================

var globalCryptoService *CryptoService
var globalMutex sync.RWMutex

// SetGlobalCryptoService sets the global crypto service instance
func SetGlobalCryptoService(c *CryptoService) {
	globalMutex.Lock()
	defer globalMutex.Unlock()
	globalCryptoService = c
}

// GetGlobalCryptoService returns the global crypto service instance
func GetGlobalCryptoService() *CryptoService {
	globalMutex.RLock()
	defer globalMutex.RUnlock()
	return globalCryptoService
}
