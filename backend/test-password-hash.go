package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "Enterenter12345!!"

	fmt.Printf("🔐 Testing password hashing for: %s\n", password)
	fmt.Printf("📊 Password length: %d characters\n", len(password))

	// Hash the password using bcrypt (same as backend)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Error hashing password:", err)
	}

	fmt.Printf("🔑 bcrypt hash: %s\n", string(hash))
	fmt.Printf("📏 Hash length: %d characters\n", len(hash))

	// Test verification
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		fmt.Printf("❌ Password verification failed: %v\n", err)
	} else {
		fmt.Printf("✅ Password verification successful!\n")
	}

	// Test with wrong password
	wrongPassword := "WrongPassword123!!"
	err = bcrypt.CompareHashAndPassword(hash, []byte(wrongPassword))
	if err != nil {
		fmt.Printf("✅ Wrong password correctly rejected: %v\n", err)
	} else {
		fmt.Printf("❌ Wrong password was accepted (this shouldn't happen!)\n")
	}
}
