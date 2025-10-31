package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"bome-backend/internal/config"
	"bome-backend/internal/database"
	"bome-backend/internal/services"
)

func main() {
	// Command line flags
	userID := flag.Int("user", 0, "Compare specific user ID (0 = compare all)")
	pretty := flag.Bool("pretty", true, "Pretty print JSON output")
	outputFile := flag.String("output", "", "Write results to file (default: stdout)")
	showOnlyDiff := flag.Bool("diff-only", false, "Show only subscribers with differences")
	verbose := flag.Bool("v", false, "Verbose logging")
	flag.Parse()

	// Initialize config and database
	cfg := config.New()
	db := database.New(cfg)
	if db == nil {
		log.Fatal("❌ Failed to initialize database")
	}
	defer db.Close()

	log.Printf("✅ Database connected")

	// Initialize services
	elasticServiceV1 := services.NewSubscriberElasticService(db)
	elasticServiceV2 := services.NewSubscriberElasticServiceV2(db)

	log.Printf("🔍 Starting v1 vs v2 comparison...")

	var results ComparisonResults

	if *userID > 0 {
		// Compare single user
		log.Printf("📋 Comparing user %d", *userID)
		result := compareUser(*userID, elasticServiceV1, elasticServiceV2, *verbose)
		results.Subscribers = []SubscriberComparison{result}
	} else {
		// Compare all users
		log.Printf("📋 Comparing ALL subscribers...")
		results = compareAllUsers(elasticServiceV1, elasticServiceV2, *showOnlyDiff, *verbose)
	}

	// Generate output
	var output []byte
	var err error
	if *pretty {
		output, err = json.MarshalIndent(results, "", "  ")
	} else {
		output, err = json.Marshal(results)
	}
	if err != nil {
		log.Fatalf("❌ Failed to marshal results: %v", err)
	}

	// Write output
	if *outputFile != "" {
		err = os.WriteFile(*outputFile, output, 0644)
		if err != nil {
			log.Fatalf("❌ Failed to write to file: %v", err)
		}
		log.Printf("✅ Results written to %s", *outputFile)
	} else {
		fmt.Println(string(output))
	}

	// Print summary
	printSummary(results)
}

type ComparisonResults struct {
	TotalCompared  int                    `json:"total_compared"`
	IdenticalCount int                    `json:"identical_count"`
	DifferentCount int                    `json:"different_count"`
	V1OnlyCount    int                    `json:"v1_only_count"`
	V2OnlyCount    int                    `json:"v2_only_count"`
	Subscribers    []SubscriberComparison `json:"subscribers"`
	Summary        map[string]int         `json:"summary"`
}

type SubscriberComparison struct {
	UserID      int                         `json:"user_id"`
	Email       string                      `json:"email"`
	Status      string                      `json:"status"` // "identical", "different", "v1_only", "v2_only"
	V1Data      *services.UnifiedSubscriber `json:"v1_data,omitempty"`
	V2Data      *services.UnifiedSubscriber `json:"v2_data,omitempty"`
	Differences []FieldDifference           `json:"differences,omitempty"`
}

type FieldDifference struct {
	Field   string      `json:"field"`
	V1Value interface{} `json:"v1_value"`
	V2Value interface{} `json:"v2_value"`
}

func compareAllUsers(v1Service *services.SubscriberElasticService, v2Service *services.SubscriberElasticServiceV2, diffOnly bool, verbose bool) ComparisonResults {
	results := ComparisonResults{
		Summary: make(map[string]int),
	}

	// Get all v1 subscribers
	v1Subscribers, err := v1Service.GetAllUnifiedSubscribers()
	if err != nil {
		log.Fatalf("❌ Failed to get v1 subscribers: %v", err)
	}

	// Get all v2 subscribers
	v2Subscribers, err := v2Service.GetAllUnifiedSubscribersV2()
	if err != nil {
		log.Fatalf("❌ Failed to get v2 subscribers: %v", err)
	}

	log.Printf("📊 V1: %d subscribers | V2: %d subscribers", len(v1Subscribers), len(v2Subscribers))

	// Create maps for easy lookup
	v1Map := make(map[int]*services.UnifiedSubscriber)
	for i := range v1Subscribers {
		v1Map[v1Subscribers[i].ID] = &v1Subscribers[i]
	}

	v2Map := make(map[int]*services.UnifiedSubscriber)
	for i := range v2Subscribers {
		v2Map[v2Subscribers[i].ID] = &v2Subscribers[i]
	}

	// Compare all users
	allUserIDs := make(map[int]bool)
	for id := range v1Map {
		allUserIDs[id] = true
	}
	for id := range v2Map {
		allUserIDs[id] = true
	}

	for userID := range allUserIDs {
		v1Sub := v1Map[userID]
		v2Sub := v2Map[userID]

		comparison := compareSubscribers(userID, v1Sub, v2Sub, verbose)

		// Skip if diff-only and this is identical
		if diffOnly && comparison.Status == "identical" {
			continue
		}

		results.Subscribers = append(results.Subscribers, comparison)

		// Update counters
		results.TotalCompared++
		switch comparison.Status {
		case "identical":
			results.IdenticalCount++
		case "different":
			results.DifferentCount++
		case "v1_only":
			results.V1OnlyCount++
		case "v2_only":
			results.V2OnlyCount++
		}

		// Track difference types
		for _, diff := range comparison.Differences {
			results.Summary[diff.Field]++
		}
	}

	return results
}

func compareUser(userID int, v1Service *services.SubscriberElasticService, v2Service *services.SubscriberElasticServiceV2, verbose bool) SubscriberComparison {
	// Get v1 data
	v1Sub, err := v1Service.GetUnifiedSubscriberByID(userID)
	if err != nil {
		log.Printf("⚠️  Failed to get v1 data for user %d: %v", userID, err)
	}

	// Get v2 data
	v2Sub, err := v2Service.GetUnifiedSubscriberByIDV2(userID)
	if err != nil {
		log.Printf("⚠️  Failed to get v2 data for user %d: %v", userID, err)
	}

	return compareSubscribers(userID, v1Sub, v2Sub, verbose)
}

func compareSubscribers(userID int, v1Sub, v2Sub *services.UnifiedSubscriber, verbose bool) SubscriberComparison {
	comparison := SubscriberComparison{
		UserID: userID,
	}

	// Handle cases where data exists in only one version
	if v1Sub == nil && v2Sub == nil {
		comparison.Status = "both_missing"
		return comparison
	}

	if v1Sub == nil {
		comparison.Status = "v2_only"
		comparison.V2Data = v2Sub
		comparison.Email = v2Sub.Email
		return comparison
	}

	if v2Sub == nil {
		comparison.Status = "v1_only"
		comparison.V1Data = v1Sub
		comparison.Email = v1Sub.Email
		return comparison
	}

	// Both exist - compare fields
	comparison.V1Data = v1Sub
	comparison.V2Data = v2Sub
	comparison.Email = v1Sub.Email

	// Compare critical fields
	diffs := []FieldDifference{}

	// Email (both are string, not *string)
	if v1Sub.Email != v2Sub.Email {
		diffs = append(diffs, FieldDifference{"email", v1Sub.Email, v2Sub.Email})
	}

	// Full Name (construct from FirstName + LastName for both)
	v1FullName := v1Sub.FirstName + " " + v1Sub.LastName
	v2FullName := v2Sub.FirstName + " " + v2Sub.LastName
	if v1FullName != v2FullName {
		diffs = append(diffs, FieldDifference{"full_name", v1FullName, v2FullName})
	}

	// Video Access
	if v1Sub.HasVideoAccess != v2Sub.HasVideoAccess {
		diffs = append(diffs, FieldDifference{"has_video_access", v1Sub.HasVideoAccess, v2Sub.HasVideoAccess})
	}

	// Active Plan
	if v1Sub.HasActivePlan != v2Sub.HasActivePlan {
		diffs = append(diffs, FieldDifference{"has_active_plan", v1Sub.HasActivePlan, v2Sub.HasActivePlan})
	}

	// Plan Status (string, not *string)
	if v1Sub.PlanStatus != v2Sub.PlanStatus {
		diffs = append(diffs, FieldDifference{"plan_status", v1Sub.PlanStatus, v2Sub.PlanStatus})
	}

	// Plan Name
	if ptrToString(v1Sub.PlanName) != ptrToString(v2Sub.PlanName) {
		diffs = append(diffs, FieldDifference{"plan_name", ptrToString(v1Sub.PlanName), ptrToString(v2Sub.PlanName)})
	}

	// MRR (with tolerance for floating point)
	if !floatEqual(v1Sub.MRRContribution, v2Sub.MRRContribution, 0.01) {
		diffs = append(diffs, FieldDifference{"mrr_contribution", v1Sub.MRRContribution, v2Sub.MRRContribution})
	}

	// ARR (with tolerance for floating point)
	if !floatEqual(v1Sub.ARRContribution, v2Sub.ARRContribution, 0.01) {
		diffs = append(diffs, FieldDifference{"arr_contribution", v1Sub.ARRContribution, v2Sub.ARRContribution})
	}

	// Days Until Expiry (allow 1 day difference for timing)
	if abs(v1Sub.DaysUntilExpiry-v2Sub.DaysUntilExpiry) > 1 {
		diffs = append(diffs, FieldDifference{"days_until_expiry", v1Sub.DaysUntilExpiry, v2Sub.DaysUntilExpiry})
	}

	comparison.Differences = diffs

	if len(diffs) == 0 {
		comparison.Status = "identical"
	} else {
		comparison.Status = "different"
	}

	return comparison
}

func printSummary(results ComparisonResults) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 COMPARISON SUMMARY")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Total Compared:   %d\n", results.TotalCompared)
	fmt.Printf("✅ Identical:     %d (%.1f%%)\n", results.IdenticalCount, percentage(results.IdenticalCount, results.TotalCompared))
	fmt.Printf("⚠️  Different:     %d (%.1f%%)\n", results.DifferentCount, percentage(results.DifferentCount, results.TotalCompared))
	fmt.Printf("❌ V1 Only:       %d\n", results.V1OnlyCount)
	fmt.Printf("❌ V2 Only:       %d\n", results.V2OnlyCount)
	fmt.Println(strings.Repeat("=", 60))

	if len(results.Summary) > 0 {
		fmt.Println("\n📋 Differences by Field:")
		for field, count := range results.Summary {
			fmt.Printf("  - %s: %d\n", field, count)
		}
		fmt.Println(strings.Repeat("=", 60))
	}

	// Recommendation
	if results.DifferentCount == 0 && results.V1OnlyCount == 0 && results.V2OnlyCount == 0 {
		fmt.Println("✅ RESULT: Perfect match! Ready for migration.")
	} else if results.DifferentCount > 0 {
		fmt.Printf("⚠️  RESULT: %d discrepancies found. Review before migration.\n", results.DifferentCount)
	} else {
		fmt.Println("❌ RESULT: Data mismatch. Investigation required.")
	}
	fmt.Println(strings.Repeat("=", 60))
}

// Helper functions
func ptrToString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func floatEqual(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func percentage(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}
