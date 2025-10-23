package main

import (
	"fmt"
	"log"
	"os"

	"bome-backend/internal/config"
	"bome-backend/internal/database"
	"bome-backend/internal/services"
)

func main() {
	fmt.Println("🚀 BOME Resend Email Setup")
	fmt.Println("=========================")

	// Load configuration
	cfg := config.New()

	// Connect to database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Get Resend API key from environment
	resendAPIKey := os.Getenv("RESEND_API_KEY")
	if resendAPIKey == "" {
		log.Fatal("❌ RESEND_API_KEY environment variable not set")
	}

	// Validate API key format (Resend keys start with 're_')
	if len(resendAPIKey) < 10 || resendAPIKey[:3] != "re_" {
		log.Fatal("❌ Invalid Resend API key format. Keys should start with 're_'")
	}

	// Get other settings from environment with fallbacks
	fromEmail := os.Getenv("RESEND_FROM_EMAIL")
	if fromEmail == "" {
		fromEmail = "noreply@bookofmormonevidence.org"
		fmt.Printf("⚠️  RESEND_FROM_EMAIL not set, using default: %s\n", fromEmail)
	}

	fromName := os.Getenv("RESEND_FROM_NAME")
	if fromName == "" {
		fromName = "BOME Support"
		fmt.Printf("⚠️  RESEND_FROM_NAME not set, using default: %s\n", fromName)
	}

	baseURL := os.Getenv("PUBLIC_APP_URL")
	if baseURL == "" {
		baseURL = "https://bookofmormonevidence.org"
		fmt.Printf("⚠️  PUBLIC_APP_URL not set, using default: %s\n", baseURL)
	}

	// Initialize crypto service for encryption
	cryptoService := services.GetGlobalCryptoService()
	if cryptoService == nil {
		log.Fatal("❌ Failed to initialize crypto service")
	}

	// Encrypt the API key
	encryptedAPIKey, err := cryptoService.EncryptString(resendAPIKey)
	if err != nil {
		log.Fatalf("❌ Failed to encrypt API key: %v", err)
	}

	fmt.Println("\n📧 Setting up Resend email configuration...")

	// Store settings in database
	settings := map[string]struct {
		value       string
		description string
	}{
		"resend_api_key": {
			value:       encryptedAPIKey,
			description: "Resend API key (encrypted)",
		},
		"smtp_from_email": {
			value:       fromEmail,
			description: "Default sender email address",
		},
		"smtp_from_name": {
			value:       fromName,
			description: "Default sender name",
		},
		"base_url": {
			value:       baseURL,
			description: "Base URL for email links",
		},
		"support_email": {
			value:       fromEmail,
			description: "Support contact email",
		},
		"email_enabled": {
			value:       "true",
			description: "Enable email sending",
		},
		"resend_daily_limit": {
			value:       "100",
			description: "Daily email limit for Resend",
		},
		"resend_monthly_limit": {
			value:       "3000",
			description: "Monthly email limit for Resend",
		},
	}

	for key, setting := range settings {
		err := db.SetEmailSetting(key, setting.value)
		if err != nil {
			log.Printf("⚠️  Failed to set %s: %v", key, err)
		} else {
			fmt.Printf("✅ Set %s\n", setting.description)
		}
	}

	fmt.Println("\n🧪 Testing Resend configuration...")

	// Test email sending
	emailService := services.NewEmailService(db)

	// Test that we can get the API key back
	storedKey, err := db.GetEmailSetting("resend_api_key")
	if err != nil {
		log.Printf("⚠️  Failed to retrieve API key: %v", err)
	} else {
		// Try to decrypt it
		decryptedKey, err := cryptoService.DecryptString(storedKey)
		if err != nil {
			log.Printf("⚠️  Failed to decrypt API key: %v", err)
		} else if decryptedKey == resendAPIKey {
			fmt.Println("✅ API key encryption/decryption working")
		} else {
			fmt.Println("❌ API key encryption/decryption failed")
		}
	}

	// Test email service initialization
	if emailService != nil {
		fmt.Println("✅ Email service initialized successfully")
	} else {
		fmt.Println("❌ Failed to initialize email service")
	}

	fmt.Println("\n🎉 Resend setup completed!")
	fmt.Println("\n📋 Configuration Summary:")
	fmt.Printf("   API Key: %s...%s (encrypted)\n", resendAPIKey[:6], resendAPIKey[len(resendAPIKey)-4:])
	fmt.Printf("   From Email: %s\n", fromEmail)
	fmt.Printf("   From Name: %s\n", fromName)
	fmt.Printf("   Base URL: %s\n", baseURL)
	fmt.Printf("   Daily Limit: 100 emails\n")
	fmt.Printf("   Monthly Limit: 3000 emails\n")

	fmt.Println("\n💡 Next Steps:")
	fmt.Println("   1. Verify your domain in Resend dashboard")
	fmt.Println("   2. Test email verification by registering a new user")
	fmt.Println("   3. Monitor email usage in admin dashboard")
	fmt.Println("\n📚 Resend Free Tier Limits:")
	fmt.Println("   • 3,000 emails/month")
	fmt.Println("   • 100 emails/day")
	fmt.Println("   • No contact approval required!")
}
