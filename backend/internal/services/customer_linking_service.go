package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"bome-backend/internal/database"
)

// CustomerLinkingService handles linking users to their Stripe customers
type CustomerLinkingService struct {
	db *database.DB
}

// NewCustomerLinkingService creates a new customer linking service
func NewCustomerLinkingService(db *database.DB) *CustomerLinkingService {
	return &CustomerLinkingService{db: db}
}

// LinkResult represents the result of a linking operation
type LinkResult struct {
	UserID           int       `json:"user_id"`
	Email            string    `json:"email"`
	CustomersFound   int       `json:"customers_found"`
	CustomersLinked  int       `json:"customers_linked"`
	PrimaryCustomer  string    `json:"primary_customer"`
	SkippedCustomers []string  `json:"skipped_customers,omitempty"`
	Error            string    `json:"error,omitempty"`
	LinkedAt         time.Time `json:"linked_at"`
}

// UnlinkedCustomer represents a Stripe customer not yet linked to a user
type UnlinkedCustomer struct {
	StripeCustomerID string `json:"stripe_customer_id"`
	Email            string `json:"email"`
	UserID           *int   `json:"user_id"`
	UserExists       bool   `json:"user_exists"`
	HasSubscriptions bool   `json:"has_subscriptions"`
	CreatedAt        string `json:"created_at"`
}

// LinkUserToCustomers finds all Stripe customers matching the user's email and links them
func (s *CustomerLinkingService) LinkUserToCustomers(userID int) (*LinkResult, error) {
	result := &LinkResult{
		UserID:           userID,
		LinkedAt:         time.Now(),
		SkippedCustomers: []string{},
	}

	// Get user email
	var email string
	err := s.db.QueryRow("SELECT email FROM users WHERE id = $1", userID).Scan(&email)
	if err != nil {
		result.Error = fmt.Sprintf("User not found: %v", err)
		return result, err
	}
	result.Email = email

	// Find all Stripe customers with matching email
	rows, err := s.db.Query(`
		SELECT id, stripe_id, stripe_created_at 
		FROM stripe_customers_v2 
		WHERE LOWER(email) = LOWER($1)
		ORDER BY stripe_created_at DESC
	`, email)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to query customers: %v", err)
		return result, err
	}
	defer rows.Close()

	var primaryCustomerID *int
	var primaryStripeID string
	customersFound := 0

	for rows.Next() {
		var customerID int
		var stripeID string
		var createdAt time.Time

		if err := rows.Scan(&customerID, &stripeID, &createdAt); err != nil {
			log.Printf("⚠️  Failed to scan customer: %v", err)
			result.SkippedCustomers = append(result.SkippedCustomers, stripeID)
			continue
		}

		customersFound++

		// Check if already linked
		var existingLinkID int
		err := s.db.QueryRow(`
			SELECT id FROM user_stripe_customers_v2 
			WHERE user_id = $1 AND stripe_customer_id = $2
		`, userID, customerID).Scan(&existingLinkID)

		if err == nil {
			// Already linked - update last_synced
			_, err = s.db.Exec(`
				UPDATE user_stripe_customers_v2 
				SET last_synced_at = NOW() 
				WHERE id = $1
			`, existingLinkID)
			if err != nil {
				log.Printf("⚠️  Failed to update last_synced for link %d: %v", existingLinkID, err)
			}
			result.CustomersLinked++
			continue
		}

		// Not linked yet - determine if this should be primary
		isPrimary := false
		if primaryCustomerID == nil {
			// First customer (most recent) becomes primary
			isPrimary = true
			primaryCustomerID = &customerID
			primaryStripeID = stripeID
		}

		// Create link
		_, err = s.db.Exec(`
			INSERT INTO user_stripe_customers_v2 
			(user_id, stripe_customer_id, is_primary, first_linked_at, last_synced_at)
			VALUES ($1, $2, $3, NOW(), NOW())
		`, userID, customerID, isPrimary)

		if err != nil {
			log.Printf("⚠️  Failed to link customer %s: %v", stripeID, err)
			result.SkippedCustomers = append(result.SkippedCustomers, stripeID)
			continue
		}

		result.CustomersLinked++
		log.Printf("✅ Linked customer %s to user %d (primary: %v)", stripeID, userID, isPrimary)
	}

	result.CustomersFound = customersFound
	result.PrimaryCustomer = primaryStripeID

	// Update user's stripe_customer_id field if we have a primary
	if primaryStripeID != "" {
		_, err = s.db.Exec(`
			UPDATE users 
			SET stripe_customer_id = $1 
			WHERE id = $2
		`, primaryStripeID, userID)
		if err != nil {
			log.Printf("⚠️  Failed to update user.stripe_customer_id: %v", err)
		}
	}

	// Check if any of the linked customers have active subscriptions and grant video access if needed
	if result.CustomersLinked > 0 {
		s.checkAndGrantVideoAccessAfterLinking(userID, email)
	}

	// Only log if there were errors or multiple customers linked (unusual cases)
	if result.Error != "" {
		log.Printf("❌ Error linking customers for user %d (%s): %s",
			userID, email, result.Error)
	} else if result.CustomersLinked > 1 {
		log.Printf("⚠️  Linked %d customers for user %d (%s) - multiple customers detected",
			result.CustomersLinked, userID, email)
	}
	// Omit success logs for single customer links (normal case)

	return result, nil
}

// LinkAllUsers links all users to their Stripe customers
func (s *CustomerLinkingService) LinkAllUsers() ([]LinkResult, error) {
	// Get all users
	rows, err := s.db.Query(`
		SELECT id 
		FROM users 
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var results []LinkResult
	successCount := 0
	errorCount := 0

	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			log.Printf("⚠️  Failed to scan user: %v", err)
			continue
		}

		result, err := s.LinkUserToCustomers(userID)
		if err != nil {
			errorCount++
			log.Printf("❌ Failed to link user %d: %v", userID, err)
		} else if result.CustomersLinked > 0 {
			successCount++
		}

		// Only include in results if there was activity
		if result.CustomersFound > 0 || result.Error != "" {
			results = append(results, *result)
		}
	}

	log.Printf("📊 Linking complete: %d users linked, %d errors", successCount, errorCount)
	return results, nil
}

// GetUnlinkedCustomers returns Stripe customers not linked to any user
func (s *CustomerLinkingService) GetUnlinkedCustomers() ([]UnlinkedCustomer, error) {
	rows, err := s.db.Query(`
		SELECT 
			sc.stripe_id,
			sc.email,
			u.id as user_id,
			CASE WHEN u.id IS NOT NULL THEN true ELSE false END as user_exists,
			EXISTS(
				SELECT 1 FROM stripe_subscriptions_v2 ss 
				WHERE ss.stripe_customer_id = sc.id 
				AND ss.status IN ('active', 'trialing')
			) as has_active_subscriptions,
			sc.stripe_created_at
		FROM stripe_customers_v2 sc
		LEFT JOIN users u ON LOWER(u.email) = LOWER(sc.email)
		WHERE NOT EXISTS (
			SELECT 1 FROM user_stripe_customers_v2 usc 
			WHERE usc.stripe_customer_id = sc.id
		)
		ORDER BY sc.stripe_created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query unlinked customers: %w", err)
	}
	defer rows.Close()

	var unlinked []UnlinkedCustomer
	for rows.Next() {
		var uc UnlinkedCustomer
		var userID sql.NullInt64
		var createdAt time.Time

		err := rows.Scan(
			&uc.StripeCustomerID,
			&uc.Email,
			&userID,
			&uc.UserExists,
			&uc.HasSubscriptions,
			&createdAt,
		)
		if err != nil {
			log.Printf("⚠️  Failed to scan unlinked customer: %v", err)
			continue
		}

		if userID.Valid {
			uid := int(userID.Int64)
			uc.UserID = &uid
		}
		uc.CreatedAt = createdAt.Format(time.RFC3339)

		unlinked = append(unlinked, uc)
	}

	return unlinked, nil
}

// GetUserCustomers returns all Stripe customers linked to a user
func (s *CustomerLinkingService) GetUserCustomers(userID int) ([]map[string]interface{}, error) {
	rows, err := s.db.Query(`
		SELECT 
			sc.stripe_id,
			sc.email,
			usc.is_primary,
			usc.first_linked_at,
			usc.last_synced_at,
			COUNT(ss.id) as subscription_count,
			COUNT(CASE WHEN ss.status IN ('active', 'trialing') THEN 1 END) as active_subscriptions
		FROM user_stripe_customers_v2 usc
		JOIN stripe_customers_v2 sc ON sc.id = usc.stripe_customer_id
		LEFT JOIN stripe_subscriptions_v2 ss ON ss.stripe_customer_id = sc.id
		WHERE usc.user_id = $1
		GROUP BY sc.stripe_id, sc.email, usc.is_primary, usc.first_linked_at, usc.last_synced_at
		ORDER BY usc.is_primary DESC, usc.first_linked_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user customers: %w", err)
	}
	defer rows.Close()

	var customers []map[string]interface{}
	for rows.Next() {
		var stripeID, email string
		var isPrimary bool
		var firstLinked, lastSynced time.Time
		var subCount, activeSubCount int

		err := rows.Scan(&stripeID, &email, &isPrimary, &firstLinked, &lastSynced, &subCount, &activeSubCount)
		if err != nil {
			log.Printf("⚠️  Failed to scan customer: %v", err)
			continue
		}

		customers = append(customers, map[string]interface{}{
			"stripe_id":            stripeID,
			"email":                email,
			"is_primary":           isPrimary,
			"first_linked_at":      firstLinked.Format(time.RFC3339),
			"last_synced_at":       lastSynced.Format(time.RFC3339),
			"subscription_count":   subCount,
			"active_subscriptions": activeSubCount,
		})
	}

	return customers, nil
}

// SetPrimaryCustomer sets a specific customer as primary for a user
func (s *CustomerLinkingService) SetPrimaryCustomer(userID int, stripeCustomerID string) error {
	// Get the customer's internal ID
	var customerID int
	err := s.db.QueryRow(`
		SELECT id FROM stripe_customers_v2 WHERE stripe_id = $1
	`, stripeCustomerID).Scan(&customerID)
	if err != nil {
		return fmt.Errorf("customer not found: %w", err)
	}

	// Check if link exists
	var linkID int
	err = s.db.QueryRow(`
		SELECT id FROM user_stripe_customers_v2 
		WHERE user_id = $1 AND stripe_customer_id = $2
	`, userID, customerID).Scan(&linkID)
	if err != nil {
		return fmt.Errorf("customer not linked to user: %w", err)
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Remove primary flag from all other customers
	_, err = tx.Exec(`
		UPDATE user_stripe_customers_v2 
		SET is_primary = false 
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to clear primary flags: %w", err)
	}

	// Set new primary
	_, err = tx.Exec(`
		UPDATE user_stripe_customers_v2 
		SET is_primary = true 
		WHERE id = $1
	`, linkID)
	if err != nil {
		return fmt.Errorf("failed to set primary: %w", err)
	}

	// Update user's stripe_customer_id
	_, err = tx.Exec(`
		UPDATE users 
		SET stripe_customer_id = $1 
		WHERE id = $2
	`, stripeCustomerID, userID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("✅ Set %s as primary customer for user %d", stripeCustomerID, userID)
	return nil
}

// GetLinkingStats returns statistics about customer linking
func (s *CustomerLinkingService) GetLinkingStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total users
	var totalUsers int
	s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	stats["total_users"] = totalUsers

	// Users with linked customers
	var usersWithLinkedCustomers int
	s.db.QueryRow(`
		SELECT COUNT(DISTINCT user_id) FROM user_stripe_customers_v2
	`).Scan(&usersWithLinkedCustomers)
	stats["users_with_linked_customers"] = usersWithLinkedCustomers

	// Total Stripe customers
	var totalCustomers int
	s.db.QueryRow("SELECT COUNT(*) FROM stripe_customers_v2").Scan(&totalCustomers)
	stats["total_stripe_customers"] = totalCustomers

	// Linked customers
	var linkedCustomers int
	s.db.QueryRow(`
		SELECT COUNT(DISTINCT stripe_customer_id) FROM user_stripe_customers_v2
	`).Scan(&linkedCustomers)
	stats["linked_customers"] = linkedCustomers

	// Unlinked customers
	stats["unlinked_customers"] = totalCustomers - linkedCustomers

	// Users with multiple customers
	var usersWithMultipleCustomers int
	s.db.QueryRow(`
		SELECT COUNT(*) FROM (
			SELECT user_id 
			FROM user_stripe_customers_v2 
			GROUP BY user_id 
			HAVING COUNT(*) > 1
		) t
	`).Scan(&usersWithMultipleCustomers)
	stats["users_with_multiple_customers"] = usersWithMultipleCustomers

	// Users with active subscriptions but no linked customer
	var usersWithOrphanedSubs int
	s.db.QueryRow(`
		SELECT COUNT(DISTINCT u.id)
		FROM users u
		JOIN stripe_customers_v2 sc ON LOWER(sc.email) = LOWER(u.email)
		JOIN stripe_subscriptions_v2 ss ON ss.stripe_customer_id = sc.id
		WHERE ss.status IN ('active', 'trialing')
		AND NOT EXISTS (
			SELECT 1 FROM user_stripe_customers_v2 usc 
			WHERE usc.user_id = u.id
		)
	`).Scan(&usersWithOrphanedSubs)
	stats["users_with_orphaned_subscriptions"] = usersWithOrphanedSubs

	stats["linking_percentage"] = 0.0
	if totalUsers > 0 {
		stats["linking_percentage"] = float64(usersWithLinkedCustomers) / float64(totalUsers) * 100
	}

	return stats, nil
}

// GetUserByStripeCustomerID gets the user associated with a Stripe customer ID
func (s *CustomerLinkingService) GetUserByStripeCustomerID(stripeCustomerID string) (*database.User, error) {
	// Find the user via the linking table
	// IMPORTANT: stripe_customer_id in user_stripe_customers_v2 is an INTEGER FK to stripe_customers_v2.id
	// We need to join to match on the Stripe ID string
	var userID int
	query := `
		SELECT usc.user_id 
		FROM user_stripe_customers_v2 usc
		JOIN stripe_customers_v2 sc ON sc.id = usc.stripe_customer_id
		WHERE sc.stripe_id = $1
		LIMIT 1
	`

	err := s.db.QueryRow(query, stripeCustomerID).Scan(&userID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no user linked to Stripe customer %s", stripeCustomerID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user by customer ID: %w", err)
	}

	// Fetch the user details
	user, err := s.db.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user %d: %w", userID, err)
	}

	return user, nil
}

// GetUserLinkedCustomers returns just the list of Stripe customer IDs for a user
func (s *CustomerLinkingService) GetUserLinkedCustomers(userID int) ([]string, error) {
	query := `
		SELECT sc.stripe_id
		FROM user_stripe_customers_v2 usc
		JOIN stripe_customers_v2 sc ON sc.id = usc.stripe_customer_id
		WHERE usc.user_id = $1
		ORDER BY usc.is_primary DESC, usc.first_linked_at DESC
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query linked customers: %w", err)
	}
	defer rows.Close()

	var customerIDs []string
	for rows.Next() {
		var stripeID string
		if err := rows.Scan(&stripeID); err != nil {
			log.Printf("⚠️  Failed to scan customer ID: %v", err)
			continue
		}
		customerIDs = append(customerIDs, stripeID)
	}

	return customerIDs, nil
}

// checkAndGrantVideoAccessAfterLinking checks if a newly linked user has active subscriptions
// and grants video access if they do (retroactive access grant)
func (s *CustomerLinkingService) checkAndGrantVideoAccessAfterLinking(userID int, email string) {
	// Check if user already has video access
	var hasAccess bool
	err := s.db.QueryRow(`
		SELECT COALESCE(has_video_access, false) 
		FROM users 
		WHERE id = $1
	`, userID).Scan(&hasAccess)
	
	if err != nil {
		log.Printf("⚠️  [Customer Linking] Failed to check video access for user %d: %v", userID, err)
		return
	}

	// If user already has access, no need to check further
	if hasAccess {
		return
	}

	// Check if any of their linked customers have active subscriptions
	query := `
		SELECT EXISTS(
			SELECT 1 
			FROM user_stripe_customers_v2 usc
			JOIN stripe_customers_v2 sc ON sc.id = usc.stripe_customer_id
			JOIN stripe_subscriptions_v2 ss ON ss.stripe_customer_id = sc.id
			WHERE usc.user_id = $1
			AND ss.status IN ('active', 'trialing')
			AND (ss.cancel_at_period_end = false OR ss.cancel_at_period_end IS NULL)
		)
	`

	var hasActiveSubscription bool
	err = s.db.QueryRow(query, userID).Scan(&hasActiveSubscription)
	if err != nil {
		log.Printf("⚠️  [Customer Linking] Failed to check active subscriptions for user %d: %v", userID, err)
		return
	}

	// If they have an active subscription, grant video access
	if hasActiveSubscription {
		_, err = s.db.Exec(`
			UPDATE users 
			SET has_video_access = true, 
			    video_access_granted_at = NOW(),
			    video_access_source = 'retroactive_linking'
			WHERE id = $1
		`, userID)

		if err != nil {
			log.Printf("❌ [Customer Linking] Failed to grant video access to user %d: %v", userID, err)
			return
		}

		log.Printf("✅ [Customer Linking] Granted retroactive video access to user %d (%s) - active subscription found after linking", userID, email)
	}
}
