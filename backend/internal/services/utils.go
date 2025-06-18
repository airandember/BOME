package services

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomToken generates a random token of specified length
func GenerateRandomToken(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
