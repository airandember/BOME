package main

import (
	"fmt"
	"log"

	"bome-backend/internal/config"
	"bome-backend/internal/database"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
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

	fmt.Println("🔧 Setting up Mailgun configuration...")

	// Set Mailgun domain (sandbox domain from your setup)
	mailgunDomain := "sandbox9424b46275f342fa8b926ec7099c9d55.mailgun.org"
	err = db.SetEmailSetting("mailgun_domain", mailgunDomain)
	if err != nil {
		log.Fatalf("Failed to set Mailgun domain: %v", err)
	}
	fmt.Printf("✅ Set Mailgun domain: %s\n", mailgunDomain)

	// Prompt for API key
	fmt.Print("Enter your Mailgun API key: ")
	var apiKey string
	fmt.Scanln(&apiKey)

	if apiKey == "" {
		log.Fatal("❌ API key cannot be empty")
	}

	// Set Mailgun API key
	err = db.SetEmailSetting("mailgun_api_key", apiKey)
	if err != nil {
		log.Fatalf("Failed to set Mailgun API key: %v", err)
	}
	fmt.Println("✅ Set Mailgun API key")

	// Set Mailgun base URL (US endpoint)
	err = db.SetEmailSetting("mailgun_base_url", "https://api.mailgun.net")
	if err != nil {
		log.Fatalf("Failed to set Mailgun base URL: %v", err)
	}
	fmt.Println("✅ Set Mailgun base URL: https://api.mailgun.net")

	// Enable email service
	err = db.SetEmailSetting("email_enabled", "true")
	if err != nil {
		log.Fatalf("Failed to enable email service: %v", err)
	}
	fmt.Println("✅ Enabled email service")

	// Set primary provider to mailgun for testing
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

	fmt.Println("\n🎉 Mailgun configuration completed!")
	fmt.Println("📧 You can now test email sending from the admin dashboard")
	fmt.Println("⚠️  Remember to add authorized recipients in your Mailgun dashboard for sandbox domain")
}
