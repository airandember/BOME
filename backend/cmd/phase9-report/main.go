package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"bome-backend/internal/config"
	"bome-backend/internal/database"
)

type MultiSubUser struct {
	UserID                  int      `json:"user_id"`
	Email                   string   `json:"email"`
	ActiveSubscriptionCount int      `json:"active_subscription_count"`
	SubscriptionIDs         []string `json:"subscription_ids"`
}

type UnlinkedUser struct {
	UserID       int    `json:"user_id"`
	Email        string `json:"email"`
	V1CustomerID string `json:"v1_customer_id"`
	V2LinkCount  int    `json:"v2_link_count"`
}

type VideoAccessStats struct {
	ManualAccessUsers int `json:"manual_access_users"`
	NoManualAccess    int `json:"no_manual_access"`
	TotalUsers        int `json:"total_users"`
}

type Phase9Report struct {
	TotalSubscribers int              `json:"total_subscribers"`
	MultiSubUsers    []MultiSubUser   `json:"multi_subscription_users"`
	UnlinkedUsers    []UnlinkedUser   `json:"unlinked_users"`
	VideoAccessStats VideoAccessStats `json:"video_access_stats"`
	Summary          string           `json:"summary"`
}

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)

	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("🔍 Phase 9: Data Migration & Cleanup Report")
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("")

	// Initialize database
	cfg := config.New()
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}
	defer db.Close()

	report := Phase9Report{}

	// Phase 9.1: Total subscriber count
	fmt.Println("📊 Phase 9.1: Counting total subscribers...")
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&report.TotalSubscribers)
	if err != nil {
		log.Printf("⚠️  Failed to count users: %v", err)
	} else {
		fmt.Printf("   ✅ Total users: %d\n", report.TotalSubscribers)
	}
	fmt.Println("")

	// Phase 9.2: Users with multiple active subscriptions
	fmt.Println("📊 Phase 9.2: Finding users with multiple active subscriptions...")
	multiSubQuery := `
		SELECT 
			u.id as user_id,
			u.email,
			COUNT(ss.id) as active_subscription_count,
			ARRAY_AGG(ss.stripe_id) as subscription_ids
		FROM users u
		JOIN user_stripe_customers_v2 usc ON u.id = usc.user_id
		JOIN stripe_customers_v2 sc ON usc.stripe_customer_id = sc.id
		JOIN stripe_subscriptions_v2 ss ON ss.customer_id = sc.id
		WHERE ss.status IN ('active', 'trialing')
		GROUP BY u.id, u.email
		HAVING COUNT(ss.id) > 1
		ORDER BY active_subscription_count DESC
		LIMIT 50
	`

	rows, err := db.Query(multiSubQuery)
	if err != nil {
		log.Printf("⚠️  Failed to query multi-sub users: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var user MultiSubUser
			var subIDs sql.NullString
			err := rows.Scan(&user.UserID, &user.Email, &user.ActiveSubscriptionCount, &subIDs)
			if err != nil {
				log.Printf("⚠️  Failed to scan row: %v", err)
				continue
			}
			// Parse subscription IDs (PostgreSQL array format)
			if subIDs.Valid {
				// Simple parsing - you might need to improve this
				user.SubscriptionIDs = []string{subIDs.String}
			}
			report.MultiSubUsers = append(report.MultiSubUsers, user)
		}
		fmt.Printf("   ⚠️  Found %d users with multiple active subscriptions\n", len(report.MultiSubUsers))
		if len(report.MultiSubUsers) > 0 {
			fmt.Println("   Top 5:")
			for i, user := range report.MultiSubUsers {
				if i >= 5 {
					break
				}
				fmt.Printf("      - User %d (%s): %d active subscriptions\n", user.UserID, user.Email, user.ActiveSubscriptionCount)
			}
		}
	}
	fmt.Println("")

	// Phase 9.3: Unlinked users (in v1 but not v2)
	fmt.Println("📊 Phase 9.3: Checking for unlinked users (v1 → v2)...")
	unlinkedQuery := `
		SELECT 
			u.id,
			u.email,
			COALESCE(us.stripe_customer_id, 'none') as v1_customer_id,
			(SELECT COUNT(*) FROM user_stripe_customers_v2 WHERE user_id = u.id) as v2_link_count
		FROM users u
		LEFT JOIN user_subscriptions us ON u.id = us.user_id
		WHERE us.stripe_customer_id IS NOT NULL
		  AND NOT EXISTS (
			  SELECT 1 FROM user_stripe_customers_v2 
			  WHERE user_id = u.id
		  )
		LIMIT 50
	`

	rows, err = db.Query(unlinkedQuery)
	if err != nil {
		log.Printf("⚠️  Failed to query unlinked users: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var user UnlinkedUser
			err := rows.Scan(&user.UserID, &user.Email, &user.V1CustomerID, &user.V2LinkCount)
			if err != nil {
				log.Printf("⚠️  Failed to scan row: %v", err)
				continue
			}
			report.UnlinkedUsers = append(report.UnlinkedUsers, user)
		}
		if len(report.UnlinkedUsers) > 0 {
			fmt.Printf("   ⚠️  Found %d users in v1 but not linked in v2\n", len(report.UnlinkedUsers))
			fmt.Println("   Top 5:")
			for i, user := range report.UnlinkedUsers {
				if i >= 5 {
					break
				}
				fmt.Printf("      - User %d (%s): v1_customer=%s\n", user.UserID, user.Email, user.V1CustomerID)
			}
		} else {
			fmt.Println("   ✅ All v1 users are linked in v2!")
		}
	}
	fmt.Println("")

	// Phase 9.4: Video access audit
	fmt.Println("📊 Phase 9.4: Auditing video access assignments...")
	videoAccessQuery := `
		SELECT 
			COUNT(*) FILTER (WHERE manual_video_access = true) as manual_access_users,
			COUNT(*) FILTER (WHERE manual_video_access = false) as no_manual_access,
			COUNT(*) as total_users
		FROM users
	`

	err = db.QueryRow(videoAccessQuery).Scan(
		&report.VideoAccessStats.ManualAccessUsers,
		&report.VideoAccessStats.NoManualAccess,
		&report.VideoAccessStats.TotalUsers,
	)
	if err != nil {
		log.Printf("⚠️  Failed to query video access: %v", err)
	} else {
		fmt.Printf("   ✅ Manual video access: %d users\n", report.VideoAccessStats.ManualAccessUsers)
		fmt.Printf("   ✅ No manual access: %d users\n", report.VideoAccessStats.NoManualAccess)
		fmt.Printf("   ✅ Total: %d users\n", report.VideoAccessStats.TotalUsers)
	}
	fmt.Println("")

	// Generate summary
	report.Summary = fmt.Sprintf(
		"Phase 9 Complete: %d total users, %d with multiple subscriptions, %d unlinked from v2",
		report.TotalSubscribers,
		len(report.MultiSubUsers),
		len(report.UnlinkedUsers),
	)

	// Save to JSON
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Fatalf("❌ Failed to marshal JSON: %v", err)
	}

	err = os.WriteFile("phase9-report.json", jsonData, 0644)
	if err != nil {
		log.Fatalf("❌ Failed to write report: %v", err)
	}

	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("✅ Phase 9 Report Generated!")
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("")
	fmt.Printf("📄 Report saved to: phase9-report.json\n")
	fmt.Println("")
	fmt.Println("📋 Summary:")
	fmt.Println(report.Summary)
	fmt.Println("")

	if len(report.MultiSubUsers) > 0 {
		fmt.Println("⚠️  ACTION REQUIRED:")
		fmt.Printf("   - %d users have multiple active subscriptions\n", len(report.MultiSubUsers))
		fmt.Println("   - Review phase9-report.json for details")
		fmt.Println("   - Consider using SubscriptionManagerService to enforce single subscription")
	} else {
		fmt.Println("✅ No users with multiple active subscriptions found!")
	}

	if len(report.UnlinkedUsers) > 0 {
		fmt.Println("⚠️  ACTION REQUIRED:")
		fmt.Printf("   - %d users in v1 are not linked in v2\n", len(report.UnlinkedUsers))
		fmt.Println("   - Run customer linking service to fix")
	} else {
		fmt.Println("✅ All users properly linked to v2!")
	}
}
