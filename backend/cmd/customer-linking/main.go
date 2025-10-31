package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"

	"bome-backend/internal/config"
	"bome-backend/internal/database"
	"bome-backend/internal/services"
)

func main() {
	// Parse command line flags
	statsOnly := flag.Bool("stats", false, "Only show linking statistics")
	showUnlinked := flag.Bool("unlinked", false, "Show unlinked customers")
	linkAll := flag.Bool("link-all", false, "Link all users to their Stripe customers")
	userID := flag.Int("user", 0, "Link a specific user ID")
	pretty := flag.Bool("pretty", false, "Pretty print JSON output")
	flag.Parse()

	// Initialize config
	cfg := config.New()

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Printf("✅ Connected to database\n")

	// Initialize customer linking service
	linkingService := services.NewCustomerLinkingService(db)

	// Handle different commands
	switch {
	case *statsOnly:
		showStats(linkingService, *pretty)
	case *showUnlinked:
		showUnlinkedCustomers(linkingService, *pretty)
	case *linkAll:
		linkAllUsers(linkingService, *pretty)
	case *userID > 0:
		linkSingleUser(linkingService, *userID, *pretty)
	default:
		// Default: show stats
		showStats(linkingService, *pretty)
	}
}

func showStats(service *services.CustomerLinkingService, pretty bool) {
	log.Printf("📊 Fetching linking statistics...\n")

	stats, err := service.GetLinkingStats()
	if err != nil {
		log.Fatalf("❌ Failed to get stats: %v", err)
	}

	printJSON(stats, pretty)

	// Print human-readable summary
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("CUSTOMER LINKING SUMMARY")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Total Users:                     %v\n", stats["total_users"])
	fmt.Printf("Users with Linked Customers:     %v\n", stats["users_with_linked_customers"])
	fmt.Printf("Linking Percentage:              %.1f%%\n", stats["linking_percentage"])
	fmt.Printf("\n")
	fmt.Printf("Total Stripe Customers:          %v\n", stats["total_stripe_customers"])
	fmt.Printf("Linked Customers:                %v\n", stats["linked_customers"])
	fmt.Printf("Unlinked Customers:              %v\n", stats["unlinked_customers"])
	fmt.Printf("\n")
	fmt.Printf("Users with Multiple Customers:   %v\n", stats["users_with_multiple_customers"])
	fmt.Printf("Users with Orphaned Subs:        %v\n", stats["users_with_orphaned_subscriptions"])
	fmt.Println(strings.Repeat("=", 80))
}

func showUnlinkedCustomers(service *services.CustomerLinkingService, pretty bool) {
	log.Printf("🔍 Fetching unlinked customers...\n")

	unlinked, err := service.GetUnlinkedCustomers()
	if err != nil {
		log.Fatalf("❌ Failed to get unlinked customers: %v", err)
	}

	log.Printf("📋 Found %d unlinked customers\n", len(unlinked))
	printJSON(unlinked, pretty)

	// Print summary
	withUsers := 0
	withActiveSubs := 0
	for _, uc := range unlinked {
		if uc.UserExists {
			withUsers++
		}
		if uc.HasSubscriptions {
			withActiveSubs++
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("Unlinked with matching users:    %d\n", withUsers)
	fmt.Printf("Unlinked with active subs:       %d\n", withActiveSubs)
	fmt.Println(strings.Repeat("=", 80))
}

func linkAllUsers(service *services.CustomerLinkingService, pretty bool) {
	log.Printf("🔗 Linking all users to their Stripe customers...\n")
	log.Printf("⏳ This may take a while...\n\n")

	results, err := service.LinkAllUsers()
	if err != nil {
		log.Fatalf("❌ Failed to link all users: %v", err)
	}

	// Count results
	successCount := 0
	errorCount := 0
	totalLinked := 0

	for _, r := range results {
		if r.Error != "" {
			errorCount++
		} else if r.CustomersLinked > 0 {
			successCount++
			totalLinked += r.CustomersLinked
		}
	}

	log.Printf("\n✅ Linking complete!\n")
	log.Printf("   Users processed:    %d\n", len(results))
	log.Printf("   Successful links:   %d\n", successCount)
	log.Printf("   Errors:             %d\n", errorCount)
	log.Printf("   Total links made:   %d\n", totalLinked)

	if pretty {
		printJSON(results, true)
	}
}

func linkSingleUser(service *services.CustomerLinkingService, userID int, pretty bool) {
	log.Printf("🔗 Linking user %d to their Stripe customers...\n", userID)

	result, err := service.LinkUserToCustomers(userID)
	if err != nil {
		log.Fatalf("❌ Failed to link user: %v", err)
	}

	printJSON(result, pretty)

	if result.Error != "" {
		fmt.Printf("\n❌ Error: %s\n", result.Error)
	} else if result.CustomersLinked > 0 {
		fmt.Printf("\n✅ Successfully linked %d/%d customers for user %d (%s)\n",
			result.CustomersLinked, result.CustomersFound, userID, result.Email)
		fmt.Printf("   Primary customer: %s\n", result.PrimaryCustomer)
	} else {
		fmt.Printf("\n⚠️  No customers found for user %d (%s)\n", userID, result.Email)
	}
}

func printJSON(data interface{}, pretty bool) {
	var bytes []byte
	var err error

	if pretty {
		bytes, err = json.MarshalIndent(data, "", "  ")
	} else {
		bytes, err = json.Marshal(data)
	}

	if err != nil {
		log.Fatalf("❌ Failed to marshal JSON: %v", err)
	}

	fmt.Println(string(bytes))
}
