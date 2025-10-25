package main

import (
	"encoding/json"
	"fmt"
	"log"

	"bome-backend/internal/config"
	"bome-backend/internal/database"
	"bome-backend/internal/services"
)

func main() {
	fmt.Println("🔍 Testing Subscriber Elastic Service...")
	
	// Initialize database
	cfg := config.New()
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	
	// Create elastic service
	elasticService := services.NewSubscriberElasticService(db)
	
	// Test 1: Get all unified subscribers
	fmt.Println("\n📊 Test 1: Getting all unified subscribers...")
	subscribers, err := elasticService.GetAllUnifiedSubscribers()
	if err != nil {
		log.Printf("❌ Error getting subscribers: %v", err)
	} else {
		fmt.Printf("✅ Retrieved %d unified subscribers\n", len(subscribers))
		
		// Show first few subscribers
		for i, sub := range subscribers {
			if i >= 3 {
				break
			}
			fmt.Printf("  - %s (%s): Plan=%s, Access=%t, StripeIDs=%d\n", 
				sub.FullName, sub.Email, 
				getStringValue(sub.PlanName), 
				sub.HasVideoAccess,
				len(sub.StripeCustomerIDs))
		}
	}
	
	// Test 2: Find subscribers with multiple Stripe customers
	fmt.Println("\n🔍 Test 2: Finding subscribers with multiple Stripe customers...")
	multipleCustomers, err := elasticService.GetSubscribersWithMultipleStripeCustomers()
	if err != nil {
		log.Printf("❌ Error getting multiple customers: %v", err)
	} else {
		fmt.Printf("✅ Found %d subscribers with multiple Stripe customers\n", len(multipleCustomers))
		
		for _, sub := range multipleCustomers {
			fmt.Printf("  - %s (%s): %d Stripe IDs: %v\n", 
				sub.FullName, sub.Email, 
				len(sub.StripeCustomerIDs), 
				sub.StripeCustomerIDs)
		}
	}
	
	// Test 3: Find subscribers with active plans but no access
	fmt.Println("\n⚠️ Test 3: Finding subscribers with active plans but no access...")
	noAccess, err := elasticService.GetSubscribersWithActivePlansButNoAccess()
	if err != nil {
		log.Printf("❌ Error getting no access: %v", err)
	} else {
		fmt.Printf("✅ Found %d subscribers with active plans but no access\n", len(noAccess))
		
		for _, sub := range noAccess {
			fmt.Printf("  - %s (%s): Plan=%s, Status=%s, Access=%t\n", 
				sub.FullName, sub.Email, 
				getStringValue(sub.PlanName),
				sub.PlanStatus,
				sub.HasVideoAccess)
		}
	}
	
	// Test 4: Get subscriber statistics
	fmt.Println("\n📈 Test 4: Getting subscriber statistics...")
	stats, err := elasticService.GetSubscriberStats()
	if err != nil {
		log.Printf("❌ Error getting stats: %v", err)
	} else {
		fmt.Printf("✅ Retrieved subscriber statistics:\n")
		statsJSON, _ := json.MarshalIndent(stats, "  ", "  ")
		fmt.Printf("%s\n", statsJSON)
	}
	
	// Test 5: Test specific user (James M Kersey II)
	fmt.Println("\n🎯 Test 5: Testing specific user (James M Kersey II)...")
	jamesSubscriber, err := elasticService.GetUnifiedSubscriberByEmail("jameskersey2@gmail.com")
	if err != nil {
		log.Printf("❌ Error getting James: %v", err)
	} else {
		fmt.Printf("✅ Found James M Kersey II:\n")
		fmt.Printf("  - ID: %d\n", jamesSubscriber.ID)
		fmt.Printf("  - Email: %s\n", jamesSubscriber.Email)
		fmt.Printf("  - Plan: %s\n", getStringValue(jamesSubscriber.PlanName))
		fmt.Printf("  - Plan Type: %s\n", jamesSubscriber.PlanType)
		fmt.Printf("  - Plan Status: %s\n", jamesSubscriber.PlanStatus)
		fmt.Printf("  - Has Active Plan: %t\n", jamesSubscriber.HasActivePlan)
		fmt.Printf("  - Has Video Access: %t\n", jamesSubscriber.HasVideoAccess)
		fmt.Printf("  - Manual Access: %t\n", jamesSubscriber.ManualAccessGranted)
		fmt.Printf("  - Stripe Customer IDs: %v\n", jamesSubscriber.StripeCustomerIDs)
		fmt.Printf("  - MRR Contribution: $%.2f\n", jamesSubscriber.MRRContribution)
		fmt.Printf("  - ARR Contribution: $%.2f\n", jamesSubscriber.ARRContribution)
	}
	
	fmt.Println("\n🎉 Elastic service testing complete!")
}

// Helper function to safely get string value from pointer
func getStringValue(s *string) string {
	if s == nil {
		return "N/A"
	}
	return *s
}
