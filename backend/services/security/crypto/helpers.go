package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"sync"
	"time"
)

// =============================================================================
// TOKEN BLACKLIST
// =============================================================================

// TokenBlacklist manages blacklisted tokens
type TokenBlacklist struct {
	tokens map[string]time.Time
	mutex  sync.RWMutex
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
	}
	return false
}

// CleanupExpiredTokens removes expired tokens
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

// =============================================================================
// RATE LIMITERS
// =============================================================================

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	limit      int
	window     time.Duration
	requests   map[string][]time.Time
	mutex      sync.RWMutex
	cleanupTTL time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		limit:      limit,
		window:     window,
		requests:   make(map[string][]time.Time),
		cleanupTTL: window * 2,
	}
	go rl.cleanup()
	return rl
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Get existing requests for this key
	times, exists := rl.requests[key]
	if !exists {
		rl.requests[key] = []time.Time{now}
		return true
	}

	// Filter out old requests
	var validTimes []time.Time
	for _, t := range times {
		if t.After(windowStart) {
			validTimes = append(validTimes, t)
		}
	}

	// Check if limit exceeded
	if len(validTimes) >= rl.limit {
		rl.requests[key] = validTimes
		return false
	}

	// Add new request
	validTimes = append(validTimes, now)
	rl.requests[key] = validTimes
	return true
}

// cleanup periodically removes old entries
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupTTL)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.cleanupTTL)
		for key, times := range rl.requests {
			if len(times) == 0 || times[len(times)-1].Before(cutoff) {
				delete(rl.requests, key)
			}
		}
		rl.mutex.Unlock()
	}
}

// =============================================================================
// ENHANCED RATE LIMITER (for login attempts)
// =============================================================================

// LoginAttempt tracks login attempts per user
type LoginAttempt struct {
	attempts     int
	firstAttempt time.Time
	lockUntil    time.Time
}

// EnhancedRateLimiter implements advanced rate limiting with lockout
type EnhancedRateLimiter struct {
	maxAttempts int
	window      time.Duration
	lockoutTime time.Duration
	attempts    map[string]*LoginAttempt
	mutex       sync.RWMutex
}

// NewEnhancedRateLimiter creates a new enhanced rate limiter
func NewEnhancedRateLimiter(maxAttempts int, window, lockoutTime time.Duration) *EnhancedRateLimiter {
	rl := &EnhancedRateLimiter{
		maxAttempts: maxAttempts,
		window:      window,
		lockoutTime: lockoutTime,
		attempts:    make(map[string]*LoginAttempt),
	}
	go rl.cleanup()
	return rl
}

// CheckLoginAttempt checks if a login attempt should be allowed
func (rl *EnhancedRateLimiter) CheckLoginAttempt(email, ip string) bool {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	key := email + "|" + ip
	attempt, exists := rl.attempts[key]
	if !exists {
		return true
	}

	// Check if currently locked out
	if time.Now().Before(attempt.lockUntil) {
		return false
	}

	// Check if window has expired
	if time.Now().After(attempt.firstAttempt.Add(rl.window)) {
		return true
	}

	// Check if max attempts exceeded
	return attempt.attempts < rl.maxAttempts
}

// RecordFailedAttempt records a failed login attempt
func (rl *EnhancedRateLimiter) RecordFailedAttempt(email, ip string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	key := email + "|" + ip
	now := time.Now()

	attempt, exists := rl.attempts[key]
	if !exists {
		rl.attempts[key] = &LoginAttempt{
			attempts:     1,
			firstAttempt: now,
			lockUntil:    time.Time{},
		}
		return
	}

	// Reset if window expired
	if now.After(attempt.firstAttempt.Add(rl.window)) {
		attempt.attempts = 1
		attempt.firstAttempt = now
		attempt.lockUntil = time.Time{}
		return
	}

	// Increment attempts
	attempt.attempts++

	// Lock if max attempts exceeded
	if attempt.attempts >= rl.maxAttempts {
		attempt.lockUntil = now.Add(rl.lockoutTime)
	}
}

// RecordSuccessfulAttempt clears failed attempts on successful login
func (rl *EnhancedRateLimiter) RecordSuccessfulAttempt(email string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	// Clear all attempts for this email
	for key := range rl.attempts {
		if strings.HasPrefix(key, email+"|") {
			delete(rl.attempts, key)
		}
	}
}

// GetRemainingLockoutTime returns the remaining lockout time
func (rl *EnhancedRateLimiter) GetRemainingLockoutTime(email string) time.Duration {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	for key, attempt := range rl.attempts {
		if strings.HasPrefix(key, email+"|") {
			if time.Now().Before(attempt.lockUntil) {
				return time.Until(attempt.lockUntil)
			}
		}
	}

	return 0
}

// cleanup periodically removes old entries
func (rl *EnhancedRateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window * 2)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()
		for key, attempt := range rl.attempts {
			// Remove if window expired and not locked
			if now.After(attempt.firstAttempt.Add(rl.window)) && now.After(attempt.lockUntil) {
				delete(rl.attempts, key)
			}
		}
		rl.mutex.Unlock()
	}
}

// =============================================================================
// GLOBAL RATE LIMITER INSTANCES (for backward compatibility)
// =============================================================================

var (
	RegisterRateLimiter      = NewRateLimiter(5, 15*time.Minute)
	LoginRateLimiter         = NewRateLimiter(10, 15*time.Minute)
	PasswordRateLimiter      = NewRateLimiter(3, 15*time.Minute)
	EnhancedLoginRateLimiter = NewEnhancedRateLimiter(5, 15*time.Minute, 1*time.Hour)
)

// =============================================================================
// PASSWORD VALIDATION HELPERS
// =============================================================================

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
	}

	lowerPassword := strings.ToLower(password)
	for _, common := range commonPasswords {
		if lowerPassword == common {
			return true
		}
	}

	return false
}

// =============================================================================
// TOKEN GENERATION HELPERS
// =============================================================================

// GenerateRandomToken generates a random token of specified length
func GenerateRandomToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}

// GenerateSecureToken generates a secure random token
func GenerateSecureToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(bytes)
}
