package main

import (
	"fmt"
	"log"
	"os"

	"bome-backend/internal/config"
	"bome-backend/internal/database"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.New()

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("🔧 Setting up Mailgun configuration for production...")

	// Read Mailgun settings from environment variables
	mailgunDomain := os.Getenv("EMAIL_MG_SNDBX_DMN")
	if mailgunDomain == "" {
		log.Fatal("❌ EMAIL_MG_SNDBX_DMN environment variable not set")
	}
	err = db.SetEmailSetting("mailgun_domain", mailgunDomain)
	if err != nil {
		log.Fatalf("Failed to set Mailgun domain: %v", err)
	}
	fmt.Printf("✅ Set Mailgun domain: %s\n", mailgunDomain)

	// Set Mailgun API key (from your credentials)
	apiKey := os.Getenv("EMAIL_MG_100")
	if apiKey == "" {
		log.Fatal("❌ EMAIL_MG_100 environment variable not set")
	}
	err = db.SetEmailSetting("mailgun_api_key", apiKey)
	if err != nil {
		log.Fatalf("Failed to set Mailgun API key: %v", err)
	}
	fmt.Println("✅ Set Mailgun API key")

	// Set Mailgun base URL from environment
	baseURL := os.Getenv("EMAIL_MG_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.mailgun.net" // fallback
	}
	err = db.SetEmailSetting("mailgun_base_url", baseURL)
	if err != nil {
		log.Fatalf("Failed to set Mailgun base URL: %v", err)
	}
	fmt.Printf("✅ Set Mailgun base URL: %s\n", baseURL)

	// Set from email from environment
	fromEmail := os.Getenv("EMAIL_MG_FROM_EMAIL")
	if fromEmail == "" {
		fromEmail = "noreply@bookofmormonevidence.org" // fallback
	}
	err = db.SetEmailSetting("mailgun_from_email", fromEmail)
	if err != nil {
		log.Fatalf("Failed to set from email: %v", err)
	}
	fmt.Printf("✅ Set from email: %s\n", fromEmail)

	// Set from name from environment
	fromName := os.Getenv("EMAIL_MG_FROM_NAME")
	if fromName == "" {
		fromName = "Book of Mormon Evidence" // fallback
	}
	err = db.SetEmailSetting("mailgun_from_name", fromName)
	if err != nil {
		log.Fatalf("Failed to set from name: %v", err)
	}
	fmt.Printf("✅ Set from name: %s\n", fromName)

	// Enable email service
	err = db.SetEmailSetting("email_enabled", "true")
	if err != nil {
		log.Fatalf("Failed to enable email service: %v", err)
	}
	fmt.Println("✅ Enabled email service")

	// Set primary provider to mailgun
	err = db.SetEmailSetting("email_provider_primary", "mailgun")
	if err != nil {
		log.Fatalf("Failed to set primary provider: %v", err)
	}
	fmt.Println("✅ Set primary provider to Mailgun")

	// Enable auto failover
	err = db.SetEmailSetting("auto_failover_enabled", "true")
	if err != nil {
		log.Fatalf("Failed to enable auto failover: %v", err)
	}
	fmt.Println("✅ Enabled auto failover")

	// Set daily limits
	err = db.SetEmailSetting("mailgun_daily_limit", "300")
	if err != nil {
		log.Printf("⚠️ Failed to set daily limit: %v", err)
	} else {
		fmt.Println("✅ Set Mailgun daily limit: 300")
	}

	// Set monthly limits
	err = db.SetEmailSetting("mailgun_monthly_limit", "5000")
	if err != nil {
		log.Printf("⚠️ Failed to set monthly limit: %v", err)
	} else {
		fmt.Println("✅ Set Mailgun monthly limit: 5000")
	}

	fmt.Println("\n🎉 Mailgun configuration completed!")
	fmt.Println("📧 Mailgun is now configured and ready to use")
	fmt.Println("🔄 The system will automatically failover from Resend to Mailgun if needed")
	fmt.Printf("⚠️  Remember: This is a sandbox domain, so you can only send to authorized recipients\n")
	fmt.Printf("   Add jake@bookofmormonevidence.org to your Mailgun authorized recipients list\n")

	// Test the configuration
	fmt.Println("\n🧪 Testing Mailgun configuration...")

	// Try to get the settings back to verify they were saved
	domain, err := db.GetEmailSetting("mailgun_domain")
	if err != nil {
		fmt.Printf("❌ Failed to retrieve domain: %v\n", err)
	} else {
		fmt.Printf("✅ Domain retrieved: %s\n", domain)
	}

	apiKeyTest, err := db.GetEmailSetting("mailgun_api_key")
	if err != nil {
		fmt.Printf("❌ Failed to retrieve API key: %v\n", err)
	} else if apiKeyTest != "" {
		fmt.Printf("✅ API key retrieved: %s...\n", apiKeyTest[:10])
	}
}
