package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
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

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return fmt.Errorf("password too long")
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool

	for _, char := range password {
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
