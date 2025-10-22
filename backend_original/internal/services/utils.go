package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"sync"
	"time"
)

// GenerateRandomToken generates a cryptographically secure random token
func GenerateRandomToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to time-based token if crypto/rand fails
		return fmt.Sprintf("%d%d", time.Now().UnixNano(), time.Now().Unix())
	}
	return hex.EncodeToString(bytes)
}

// GenerateSecureToken generates a URL-safe random token
func GenerateSecureToken() string {
	return GenerateRandomToken(32)
}

// ValidateEmail validates email format and common patterns
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	// Parse email address
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email format")
	}

	// Additional validation
	if addr.Address != email {
		return fmt.Errorf("invalid email format")
	}

	// Check for common issues
	if strings.Contains(email, "..") {
		return fmt.Errorf("email contains consecutive dots")
	}

	if len(email) > 254 {
		return fmt.Errorf("email too long")
	}

	return nil
}

// ValidatePassword validates password strength with enhanced security
func ValidatePassword(password string) error {
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

	// Check for minimum unique characters
	if len(uniqueChars) < 8 {
		return fmt.Errorf("password must contain at least 8 unique characters")
	}

	// Check for repeating patterns
	if hasRepeatingPatterns(password) {
		return fmt.Errorf("password contains repeating patterns")
	}

	// Check against common passwords
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

// hasRepeatingPatterns checks for repeating character patterns
func hasRepeatingPatterns(password string) bool {
	if len(password) < 4 {
		return false
	}

	// Check for repeating sequences
	for i := 0; i < len(password)-3; i++ {
		pattern := password[i : i+2]
		if strings.Count(password, pattern) > 2 {
			return true
		}
	}

	// Check for keyboard patterns
	keyboardPatterns := []string{
		"qwerty", "asdfgh", "zxcvbn",
		"123456", "abcdef", "password",
		"admin", "user", "test",
	}

	lowerPassword := strings.ToLower(password)
	for _, pattern := range keyboardPatterns {
		if strings.Contains(lowerPassword, pattern) {
			return true
		}
	}

	return false
}

// isCommonPassword checks against common weak passwords
func isCommonPassword(password string) bool {
	commonPasswords := []string{
		"password", "123456", "123456789", "qwerty", "abc123",
		"password123", "admin", "letmein", "welcome", "monkey",
		"dragon", "master", "hello", "freedom", "whatever",
		"qwerty123", "trustno1", "jordan", "harley", "ranger",
		"iwantu", "jennifer", "joshua", "maggie", "password1",
		"robert", "daniel", "heather", "michelle", "charlie",
		"andrew", "matthew", "abigail", "david", "sophia",
		"james", "elizabeth", "olivia", "emma", "noah",
		"william", "ava", "isabella", "mason", "sophia",
	}

	lowerPassword := strings.ToLower(password)
	for _, common := range commonPasswords {
		if lowerPassword == common {
			return true
		}
	}

	return false
}

// ValidateName validates first/last name
func ValidateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if len(name) > 50 {
		return fmt.Errorf("name too long")
	}

	// Allow letters, spaces, hyphens, and apostrophes
	validName := regexp.MustCompile(`^[a-zA-Z\s\-']+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("name contains invalid characters")
	}

	return nil
}

// SanitizeString removes potentially dangerous characters
func SanitizeString(input string) string {
	// Remove null bytes and control characters
	cleaned := strings.Map(func(r rune) rune {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return -1
		}
		return r
	}, input)

	return strings.TrimSpace(cleaned)
}

// GetClientIP extracts the real client IP from request
func GetClientIP(remoteAddr, xForwardedFor, xRealIP string) string {
	// Check X-Real-IP header first
	if xRealIP != "" {
		if ip := net.ParseIP(xRealIP); ip != nil {
			return ip.String()
		}
	}

	// Check X-Forwarded-For header
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if parsedIP := net.ParseIP(ip); parsedIP != nil {
				return parsedIP.String()
			}
		}
	}

	// Fall back to remote address
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		if ip := net.ParseIP(host); ip != nil {
			return ip.String()
		}
	}

	return remoteAddr
}

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if the request is allowed for the given key
func (rl *RateLimiter) Allow(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Clean old requests
	if requests, exists := rl.requests[key]; exists {
		var validRequests []time.Time
		for _, req := range requests {
			if req.After(cutoff) {
				validRequests = append(validRequests, req)
			}
		}
		rl.requests[key] = validRequests
	}

	// Check if under limit
	if len(rl.requests[key]) >= rl.limit {
		return false
	}

	// Add current request
	rl.requests[key] = append(rl.requests[key], now)
	return true
}

// cleanup removes old entries from the rate limiter
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		cutoff := time.Now().Add(-rl.window * 2) // Keep some buffer

		for key, requests := range rl.requests {
			var validRequests []time.Time
			for _, req := range requests {
				if req.After(cutoff) {
					validRequests = append(validRequests, req)
				}
			}

			if len(validRequests) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = validRequests
			}
		}
		rl.mutex.Unlock()
	}
}

// Global rate limiters for different operations
var (
	LoginRateLimiter    = NewRateLimiter(5, 15*time.Minute) // 5 attempts per 15 minutes
	RegisterRateLimiter = NewRateLimiter(10, 1*time.Hour)   // 10 registrations per hour (increased for development)
	PasswordRateLimiter = NewRateLimiter(3, 1*time.Hour)    // 3 password resets per hour
)

// EnhancedRateLimiter provides advanced rate limiting with account lockout
type EnhancedRateLimiter struct {
	loginAttempts  map[string][]time.Time
	failedAttempts map[string]int
	lockoutUntil   map[string]time.Time
	mutex          sync.RWMutex
	window         time.Duration
	maxAttempts    int
	lockoutTime    time.Duration
}

// NewEnhancedRateLimiter creates a new enhanced rate limiter
func NewEnhancedRateLimiter(maxAttempts int, window, lockoutTime time.Duration) *EnhancedRateLimiter {
	limiter := &EnhancedRateLimiter{
		loginAttempts:  make(map[string][]time.Time),
		failedAttempts: make(map[string]int),
		lockoutUntil:   make(map[string]time.Time),
		window:         window,
		maxAttempts:    maxAttempts,
		lockoutTime:    lockoutTime,
	}

	// Start cleanup goroutine
	go limiter.cleanup()

	return limiter
}

// CheckLoginAttempt checks if a login attempt is allowed
func (rl *EnhancedRateLimiter) CheckLoginAttempt(email, ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	// Check if account is locked out
	if lockoutTime, exists := rl.lockoutUntil[email]; exists && time.Now().Before(lockoutTime) {
		return false
	}

	// Check failed attempts
	if failed := rl.failedAttempts[email]; failed >= rl.maxAttempts {
		rl.lockoutUntil[email] = time.Now().Add(rl.lockoutTime)
		return false
	}

	return true
}

// RecordFailedAttempt records a failed login attempt
func (rl *EnhancedRateLimiter) RecordFailedAttempt(email, ip string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	rl.failedAttempts[email]++

	// Record attempt with timestamp
	if rl.loginAttempts[email] == nil {
		rl.loginAttempts[email] = make([]time.Time, 0)
	}
	rl.loginAttempts[email] = append(rl.loginAttempts[email], now)

	// If max attempts reached, set lockout
	if rl.failedAttempts[email] >= rl.maxAttempts {
		rl.lockoutUntil[email] = now.Add(rl.lockoutTime)
		log.Printf("Account locked out: %s from %s for %v", email, ip, rl.lockoutTime)
	}
}

// RecordSuccessfulAttempt resets failed attempts for successful login
func (rl *EnhancedRateLimiter) RecordSuccessfulAttempt(email string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	rl.failedAttempts[email] = 0
	delete(rl.lockoutUntil, email)

	// Clear old attempts
	if attempts, exists := rl.loginAttempts[email]; exists {
		cutoff := time.Now().Add(-rl.window)
		var validAttempts []time.Time
		for _, attempt := range attempts {
			if attempt.After(cutoff) {
				validAttempts = append(validAttempts, attempt)
			}
		}
		rl.loginAttempts[email] = validAttempts
	}
}

// GetRemainingLockoutTime returns remaining lockout time for an account
func (rl *EnhancedRateLimiter) GetRemainingLockoutTime(email string) time.Duration {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	if lockoutTime, exists := rl.lockoutUntil[email]; exists {
		if remaining := time.Until(lockoutTime); remaining > 0 {
			return remaining
		}
	}
	return 0
}

// GetFailedAttempts returns the number of failed attempts for an account
func (rl *EnhancedRateLimiter) GetFailedAttempts(email string) int {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()
	return rl.failedAttempts[email]
}

// cleanup removes old entries periodically
func (rl *EnhancedRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		cutoff := time.Now().Add(-rl.window * 2)

		// Clean up expired lockouts
		for email, lockoutTime := range rl.lockoutUntil {
			if lockoutTime.Before(cutoff) {
				delete(rl.lockoutUntil, email)
				delete(rl.failedAttempts, email)
				delete(rl.loginAttempts, email)
			}
		}

		// Clean up old attempts
		for email, attempts := range rl.loginAttempts {
			var validAttempts []time.Time
			for _, attempt := range attempts {
				if attempt.After(cutoff) {
					validAttempts = append(validAttempts, attempt)
				}
			}
			if len(validAttempts) == 0 {
				delete(rl.loginAttempts, email)
			} else {
				rl.loginAttempts[email] = validAttempts
			}
		}

		rl.mutex.Unlock()
	}
}

// Global enhanced rate limiters
var (
	EnhancedLoginRateLimiter    = NewEnhancedRateLimiter(5, 15*time.Minute, 15*time.Minute)
	EnhancedRegisterRateLimiter = NewEnhancedRateLimiter(3, 1*time.Hour, 1*time.Hour)
	EnhancedPasswordRateLimiter = NewEnhancedRateLimiter(3, 1*time.Hour, 1*time.Hour)
)

// GenerateDeviceFingerprint creates a device fingerprint from request headers
func GenerateDeviceFingerprint(r *http.Request) string {
	// Create a simple device fingerprint from headers
	fingerprint := fmt.Sprintf("%s|%s|%s|%s",
		r.Header.Get("User-Agent"),
		r.Header.Get("Accept-Language"),
		r.Header.Get("Accept-Encoding"),
		r.Header.Get("Accept"),
	)

	// Hash the fingerprint for consistency
	hash := sha256.Sum256([]byte(fingerprint))
	return hex.EncodeToString(hash[:])[:16] // Return first 16 characters
}
