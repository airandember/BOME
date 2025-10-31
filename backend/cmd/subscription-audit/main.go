package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"bome-backend/internal/config"
	"bome-backend/internal/database"
)

type UserSubscriptionAudit struct {
	Email                 string   `json:"email"`
	UserID                int      `json:"user_id"`
	CustomerCount         int      `json:"customer_count"`
	CustomerIDs           []string `json:"customer_ids"`
	ActiveSubCount        int      `json:"active_subscription_count"`
	TotalSubCount         int      `json:"total_subscription_count"`
	ActiveSubscriptionIDs []string `json:"active_subscription_ids"`
	AllSubscriptionIDs    []string `json:"all_subscription_ids"`
	SubscriptionStatuses  []string `json:"subscription_statuses"`
	Issue                 string   `json:"issue"`
}

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)

	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("📊 COMPREHENSIVE SUBSCRIPTION AUDIT REPORT")
	fmt.Println("   Finding: Multiple Customers + Multiple Subscriptions per User")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("")

	// Initialize database
	cfg := config.New()
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Query to find users with issues
	query := `
		WITH user_customers AS (
			SELECT 
				u.id as user_id,
				u.email,
				COUNT(DISTINCT sc.id) as customer_count,
				ARRAY_AGG(DISTINCT sc.stripe_id ORDER BY sc.stripe_id) as customer_ids
			FROM users u
			JOIN user_stripe_customers_v2 usc ON u.id = usc.user_id
			JOIN stripe_customers_v2 sc ON usc.stripe_customer_id = sc.id
			GROUP BY u.id, u.email
		),
		user_subscriptions AS (
			SELECT 
				u.id as user_id,
				u.email,
				COUNT(DISTINCT ss.id) as total_sub_count,
				COUNT(DISTINCT CASE WHEN ss.status IN ('active', 'trialing') THEN ss.id END) as active_sub_count,
				ARRAY_AGG(DISTINCT ss.stripe_id ORDER BY ss.stripe_id) FILTER (WHERE ss.stripe_id IS NOT NULL) as all_sub_ids,
				ARRAY_AGG(DISTINCT ss.stripe_id ORDER BY ss.stripe_id) FILTER (WHERE ss.status IN ('active', 'trialing')) as active_sub_ids,
				ARRAY_AGG(DISTINCT ss.status ORDER BY ss.status) FILTER (WHERE ss.status IS NOT NULL) as statuses
			FROM users u
			JOIN user_stripe_customers_v2 usc ON u.id = usc.user_id
			JOIN stripe_customers_v2 sc ON usc.stripe_customer_id = sc.id
			LEFT JOIN stripe_subscriptions_v2 ss ON ss.customer_id = sc.id
			GROUP BY u.id, u.email
		)
		SELECT 
			uc.user_id,
			uc.email,
			uc.customer_count,
			uc.customer_ids,
			COALESCE(us.active_sub_count, 0) as active_sub_count,
			COALESCE(us.total_sub_count, 0) as total_sub_count,
			us.active_sub_ids,
			us.all_sub_ids,
			us.statuses
		FROM user_customers uc
		LEFT JOIN user_subscriptions us ON uc.user_id = us.user_id
		WHERE uc.customer_count > 1 OR COALESCE(us.active_sub_count, 0) > 1
		ORDER BY 
			COALESCE(us.active_sub_count, 0) DESC,
			uc.customer_count DESC,
			uc.email
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("❌ Failed to query: %v", err)
	}
	defer rows.Close()

	var audits []UserSubscriptionAudit
	var criticalCount int
	var multiCustomerCount int
	var multiSubCount int

	for rows.Next() {
		var audit UserSubscriptionAudit
		var customerIDsStr, activeSubIDsStr, allSubIDsStr, statusesStr sql.NullString

		err := rows.Scan(
			&audit.UserID,
			&audit.Email,
			&audit.CustomerCount,
			&customerIDsStr,
			&audit.ActiveSubCount,
			&audit.TotalSubCount,
			&activeSubIDsStr,
			&allSubIDsStr,
			&statusesStr,
		)
		if err != nil {
			log.Printf("⚠️  Failed to scan row: %v", err)
			continue
		}

		// Parse PostgreSQL arrays
		audit.CustomerIDs = parsePostgresArray(customerIDsStr)
		audit.ActiveSubscriptionIDs = parsePostgresArray(activeSubIDsStr)
		audit.AllSubscriptionIDs = parsePostgresArray(allSubIDsStr)
		audit.SubscriptionStatuses = parsePostgresArray(statusesStr)

		// Determine issue type
		issues := []string{}
		if audit.CustomerCount > 1 {
			issues = append(issues, fmt.Sprintf("%d customers", audit.CustomerCount))
			multiCustomerCount++
		}
		if audit.ActiveSubCount > 1 {
			issues = append(issues, fmt.Sprintf("%d active subs", audit.ActiveSubCount))
			multiSubCount++
		}

		audit.Issue = strings.Join(issues, " + ")

		if audit.CustomerCount > 1 && audit.ActiveSubCount > 1 {
			criticalCount++
		}

		audits = append(audits, audit)
	}

	// Display results
	fmt.Printf("📊 SUMMARY:\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════\n")
	fmt.Printf("   Total Users with Issues:              %d\n", len(audits))
	fmt.Printf("   🔴 Critical (both issues):             %d\n", criticalCount)
	fmt.Printf("   🟡 Multiple Customers Only:            %d\n", multiCustomerCount-criticalCount)
	fmt.Printf("   🟡 Multiple Active Subscriptions Only: %d\n", multiSubCount-criticalCount)
	fmt.Println("")

	if len(audits) > 0 {
		fmt.Println("═══════════════════════════════════════════════════════════════")
		fmt.Println("📋 DETAILED REPORT:")
		fmt.Println("═══════════════════════════════════════════════════════════════")
		fmt.Println("")

		for i, audit := range audits {
			severity := "🟡"
			if audit.CustomerCount > 1 && audit.ActiveSubCount > 1 {
				severity = "🔴 CRITICAL"
			}

			fmt.Printf("%s %d. %s (User ID: %d)\n", severity, i+1, audit.Email, audit.UserID)
			fmt.Printf("   Issue: %s\n", audit.Issue)
			fmt.Println("")

			// Customer details
			fmt.Printf("   Stripe Customers (%d):\n", audit.CustomerCount)
			for _, cusID := range audit.CustomerIDs {
				fmt.Printf("      • %s\n", cusID)
			}
			fmt.Println("")

			// Subscription details
			fmt.Printf("   Subscriptions (Total: %d, Active: %d):\n", audit.TotalSubCount, audit.ActiveSubCount)
			if len(audit.ActiveSubscriptionIDs) > 0 {
				fmt.Printf("      ✅ Active/Trialing (%d):\n", audit.ActiveSubCount)
				for _, subID := range audit.ActiveSubscriptionIDs {
					fmt.Printf("         • %s\n", subID)
				}
			}

			// Show other subscription statuses
			if len(audit.SubscriptionStatuses) > 0 {
				fmt.Printf("      📊 Statuses: %s\n", strings.Join(audit.SubscriptionStatuses, ", "))
			}

			fmt.Println("")
			fmt.Println("   ───────────────────────────────────────────────────────────")
			fmt.Println("")
		}
	}

	// Generate detailed JSON report
	report := map[string]interface{}{
		"generated_at": "2025-10-31",
		"summary": map[string]interface{}{
			"total_users_with_issues":              len(audits),
			"critical_users_both_issues":           criticalCount,
			"users_with_multiple_customers_only":   multiCustomerCount - criticalCount,
			"users_with_multiple_active_subs_only": multiSubCount - criticalCount,
		},
		"users": audits,
	}

	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Fatalf("❌ Failed to marshal JSON: %v", err)
	}

	err = os.WriteFile("subscription-audit-report.json", jsonData, 0644)
	if err != nil {
		log.Fatalf("❌ Failed to write report: %v", err)
	}

	// Generate CSV for easy Excel import
	csvData := generateCSV(audits)
	err = os.WriteFile("subscription-audit-report.csv", []byte(csvData), 0644)
	if err != nil {
		log.Printf("⚠️  Failed to write CSV: %v", err)
	}

	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("✅ REPORTS GENERATED!")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("")
	fmt.Printf("📄 JSON Report: subscription-audit-report.json\n")
	fmt.Printf("📊 CSV Report:  subscription-audit-report.csv\n")
	fmt.Println("")

	if len(audits) > 0 {
		fmt.Println("🔧 RECOMMENDED ACTIONS:")
		fmt.Println("")
		fmt.Printf("   1. 🔴 Critical Users (%d): Multiple customers AND multiple active subs\n", criticalCount)
		fmt.Println("      → Consolidate customers in Stripe Dashboard")
		fmt.Println("      → Cancel duplicate subscriptions (keep only one)")
		fmt.Println("")
		fmt.Printf("   2. 🟡 Multiple Customers Only (%d): One active sub, multiple cus_ IDs\n", multiCustomerCount-criticalCount)
		fmt.Println("      → Archive unused customer records in Stripe")
		fmt.Println("      → Run Simple Sync to update database")
		fmt.Println("")
		fmt.Printf("   3. 🟡 Multiple Active Subs Only (%d): One cus_ ID, multiple active subs\n", multiSubCount-criticalCount)
		fmt.Println("      → Contact user to choose which subscription to keep")
		fmt.Println("      → Cancel duplicate subscriptions")
		fmt.Println("")
		fmt.Println("💡 TIP: Use the CSV report to track cleanup progress!")
	} else {
		fmt.Println("✅ No issues found! All users have single customer + single subscription.")
	}
}

func parsePostgresArray(nullStr sql.NullString) []string {
	if !nullStr.Valid || nullStr.String == "" {
		return []string{}
	}

	str := nullStr.String
	// Remove { and }
	if len(str) > 2 {
		str = str[1 : len(str)-1]
	}

	if str == "" {
		return []string{}
	}

	// Split by comma
	parts := strings.Split(str, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" && trimmed != "NULL" {
			result = append(result, trimmed)
		}
	}

	return result
}

func generateCSV(audits []UserSubscriptionAudit) string {
	csv := "Email,User ID,Customer Count,Active Sub Count,Total Sub Count,Issue,Customer IDs,Active Sub IDs,All Sub IDs\n"

	for _, audit := range audits {
		csv += fmt.Sprintf("%s,%d,%d,%d,%d,\"%s\",\"%s\",\"%s\",\"%s\"\n",
			audit.Email,
			audit.UserID,
			audit.CustomerCount,
			audit.ActiveSubCount,
			audit.TotalSubCount,
			audit.Issue,
			strings.Join(audit.CustomerIDs, "; "),
			strings.Join(audit.ActiveSubscriptionIDs, "; "),
			strings.Join(audit.AllSubscriptionIDs, "; "),
		)
	}

	return csv
}
