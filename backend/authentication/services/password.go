package services

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a plain password
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword compares a plain password with a hash
func CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
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
