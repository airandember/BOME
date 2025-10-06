package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/stripe/stripe-go/v74"
)

// ComprehensiveStripeSyncService handles complete Stripe data synchronization
type ComprehensiveStripeSyncService struct {
	db            *sql.DB
	stripeService *StripeService
}

// NewComprehensiveStripeSyncService creates a new comprehensive sync service
func NewComprehensiveStripeSyncService(db *sql.DB, stripeService *StripeService) *ComprehensiveStripeSyncService {
	return &ComprehensiveStripeSyncService{
		db:            db,
		stripeService: stripeService,
	}
}

// ComprehensiveCustomerSyncResult represents the result of customer synchronization
type ComprehensiveCustomerSyncResult struct {
	UserID         int    `json:"user_id"`
	Email          string `json:"email"`
	OldStripeID    string `json:"old_stripe_id"`
	NewStripeID    string `json:"new_stripe_id"`
	SyncStatus     string `json:"sync_status"`
	PlanName       string `json:"plan_name"`
	SubscriptionID string `json:"subscription_id"`
	Error          string `json:"error,omitempty"`
}

// ComprehensiveSyncResult represents the overall sync results
type ComprehensiveSyncResult struct {
	TotalUsers          int                               `json:"total_users"`
	GhostCustomers      int                               `json:"ghost_customers"`
	RealStripeCustomers int                               `json:"real_stripe_customers"`
	NewlyLinked         int                               `json:"newly_linked"`
	FixedPlans          int                               `json:"fixed_plans"`
	Errors              int                               `json:"errors"`
	ProcessingTimeMs    int64                             `json:"processing_time_ms"`
	CustomerResults     []ComprehensiveCustomerSyncResult `json:"customer_results"`
}

// RunComprehensiveSync performs a complete Stripe synchronization
func (s *ComprehensiveStripeSyncService) RunComprehensiveSync(ctx context.Context) (*ComprehensiveSyncResult, error) {
	startTime := time.Now()
	log.Println("🚀 Starting Comprehensive Stripe Sync...")

	result := &ComprehensiveSyncResult{
		CustomerResults: make([]ComprehensiveCustomerSyncResult, 0),
	}

	// Step 1: Analyze current contamination
	contamination, err := s.analyzeContamination()
	if err != nil {
		return nil, fmt.Errorf("failed to analyze contamination: %v", err)
	}

	log.Printf("📊 Contamination Analysis: %d total users, %d ghosts, %d real Stripe",
		contamination.TotalUsers, contamination.GhostCustomers, contamination.RealStripeCustomers)

	result.TotalUsers = contamination.TotalUsers
	result.GhostCustomers = contamination.GhostCustomers
	result.RealStripeCustomers = contamination.RealStripeCustomers

	// Step 2: Get all users that need sync
	users, err := s.getUsersForSync()
	if err != nil {
		return nil, fmt.Errorf("failed to get users for sync: %v", err)
	}

	log.Printf("👥 Found %d users to process", len(users))

	// Step 3: Process each user
	for _, user := range users {
		customerResult := s.processUser(ctx, user)
		result.CustomerResults = append(result.CustomerResults, customerResult)

		switch customerResult.SyncStatus {
		case "newly_linked":
			result.NewlyLinked++
		case "plan_fixed":
			result.FixedPlans++
		case "error":
			result.Errors++
		}
	}

	// Step 4: Update statistics
	result.ProcessingTimeMs = time.Since(startTime).Milliseconds()

	log.Printf("✅ Comprehensive Sync Complete: %d linked, %d plans fixed, %d errors in %dms",
		result.NewlyLinked, result.FixedPlans, result.Errors, result.ProcessingTimeMs)

	return result, nil
}

// ContaminationAnalysis represents the current state of customer data
type ContaminationAnalysis struct {
	TotalUsers          int `json:"total_users"`
	RealStripeCustomers int `json:"real_stripe_customers"`
	GhostCustomers      int `json:"ghost_customers"`
	NoStripeID          int `json:"no_stripe_id"`
}

// analyzeContamination analyzes the current customer data contamination
func (s *ComprehensiveStripeSyncService) analyzeContamination() (*ContaminationAnalysis, error) {
	query := `
		SELECT 
			COUNT(*) as total_users,
			COUNT(*) FILTER (WHERE stripe_customer_id LIKE 'cus_%') as real_stripe_customers,
			COUNT(*) FILTER (WHERE stripe_customer_id LIKE '#%') as ghost_customers,
			COUNT(*) FILTER (WHERE stripe_customer_id IS NULL) as no_stripe_id
		FROM users
	`

	var analysis ContaminationAnalysis
	err := s.db.QueryRow(query).Scan(
		&analysis.TotalUsers,
		&analysis.RealStripeCustomers,
		&analysis.GhostCustomers,
		&analysis.NoStripeID,
	)

	return &analysis, err
}

// UserForSync represents a user that needs Stripe synchronization
type UserForSync struct {
	ID               int            `json:"id"`
	Email            string         `json:"email"`
	FirstName        sql.NullString `json:"first_name"`
	LastName         sql.NullString `json:"last_name"`
	StripeCustomerID sql.NullString `json:"stripe_customer_id"`
	SubID            sql.NullInt64  `json:"sub_id"`
	Role             string         `json:"role"`
	CreatedAt        time.Time      `json:"created_at"`
}

// getUsersForSync gets all users that need Stripe synchronization
func (s *ComprehensiveStripeSyncService) getUsersForSync() ([]UserForSync, error) {
	query := `
		SELECT 
			u.id, u.email, u.first_name, u.last_name, 
			u.stripe_customer_id, u.sub_id, u.role, u.created_at
		FROM users u
		WHERE u.is_active = true
		  AND u.email IS NOT NULL
		  AND u.email != ''
		ORDER BY 
			CASE 
				WHEN u.stripe_customer_id LIKE 'cus_%' THEN 1  -- Real Stripe customers first
				WHEN u.stripe_customer_id LIKE '#%' THEN 2     -- Ghost customers second
				ELSE 3                                          -- No Stripe ID last
			END,
			u.created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UserForSync
	for rows.Next() {
		var user UserForSync
		err := rows.Scan(
			&user.ID, &user.Email, &user.FirstName, &user.LastName,
			&user.StripeCustomerID, &user.SubID, &user.Role, &user.CreatedAt,
		)
		if err != nil {
			log.Printf("⚠️ Error scanning user: %v", err)
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

// processUser processes a single user for Stripe synchronization
func (s *ComprehensiveStripeSyncService) processUser(ctx context.Context, user UserForSync) ComprehensiveCustomerSyncResult {
	result := ComprehensiveCustomerSyncResult{
		UserID:      user.ID,
		Email:       user.Email,
		OldStripeID: user.StripeCustomerID.String,
		SyncStatus:  "no_change",
	}

	// Determine user type and processing strategy
	if user.StripeCustomerID.Valid {
		if strings.HasPrefix(user.StripeCustomerID.String, "cus_") {
			// Real Stripe customer - verify and sync
			result = s.processRealStripeCustomer(ctx, user, result)
		} else if strings.HasPrefix(user.StripeCustomerID.String, "#") {
			// Ghost customer - try to find real Stripe customer or mark as legacy
			result = s.processGhostCustomer(ctx, user, result)
		} else {
			// Invalid format - try to find or create
			result = s.processInvalidFormatCustomer(ctx, user, result)
		}
	} else {
		// No Stripe ID - try to find existing or create new
		result = s.processNoStripeIDCustomer(ctx, user, result)
	}

	return result
}

// processRealStripeCustomer processes a user with a real Stripe customer ID
func (s *ComprehensiveStripeSyncService) processRealStripeCustomer(ctx context.Context, user UserForSync, result ComprehensiveCustomerSyncResult) ComprehensiveCustomerSyncResult {
	// Verify the customer exists in Stripe
	stripeCustomer, err := s.stripeService.GetCustomer(user.StripeCustomerID.String)
	if err != nil {
		log.Printf("⚠️ Stripe customer %s not found for user %d: %v", user.StripeCustomerID.String, user.ID, err)
		result.SyncStatus = "stripe_customer_not_found"
		result.Error = fmt.Sprintf("Stripe customer not found: %v", err)
		return result
	}

	// Update local customer data
	err = s.updateLocalCustomerData(user.ID, stripeCustomer.ID)
	if err != nil {
		result.SyncStatus = "error"
		result.Error = fmt.Sprintf("Failed to update local data: %v", err)
		return result
	}

	// Get and update subscription/plan information
	planInfo, err := s.getCustomerPlanInfo(user.StripeCustomerID.String)
	if err != nil {
		log.Printf("⚠️ Failed to get plan info for customer %s: %v", user.StripeCustomerID.String, err)
	} else {
		result.PlanName = planInfo.PlanName
		result.SubscriptionID = planInfo.SubscriptionID
	}

	result.NewStripeID = user.StripeCustomerID.String
	result.SyncStatus = "verified"
	return result
}

// processGhostCustomer processes a user with a ghost Stripe customer ID
func (s *ComprehensiveStripeSyncService) processGhostCustomer(ctx context.Context, user UserForSync, result ComprehensiveCustomerSyncResult) ComprehensiveCustomerSyncResult {
	log.Printf("👻 Processing ghost customer: %s (%s)", user.StripeCustomerID.String, user.Email)

	// Try to find existing Stripe customer by email
	stripeCustomer, err := s.findStripeCustomerByEmail(user.Email)
	if err == nil && stripeCustomer != nil {
		// Found existing Stripe customer - link them
		log.Printf("🔗 Found existing Stripe customer for %s: %s", user.Email, stripeCustomer.ID)

		err = s.linkUserToStripeCustomer(user.ID, stripeCustomer.ID)
		if err != nil {
			result.SyncStatus = "error"
			result.Error = fmt.Sprintf("Failed to link to existing customer: %v", err)
			return result
		}

		result.NewStripeID = stripeCustomer.ID
		result.SyncStatus = "newly_linked"

		// Get plan info for newly linked customer
		planInfo, _ := s.getCustomerPlanInfo(stripeCustomer.ID)
		result.PlanName = planInfo.PlanName
		result.SubscriptionID = planInfo.SubscriptionID

		return result
	}

	// No existing Stripe customer found - check if they have legacy subscription
	if user.SubID.Valid {
		// User has legacy subscription - keep as local-only customer
		legacyPlan, err := s.getLegacyPlanName(user.SubID.Int64)
		if err == nil {
			result.PlanName = legacyPlan
		}
		result.SyncStatus = "legacy_local_only"
		result.NewStripeID = "" // Clear the ghost ID

		// Clear the ghost Stripe ID from database
		s.clearGhostStripeID(user.ID)

		return result
	}

	// No legacy subscription either - mark as ghost for cleanup
	result.SyncStatus = "ghost_for_cleanup"
	return result
}

// processInvalidFormatCustomer processes a user with invalid format Stripe ID
func (s *ComprehensiveStripeSyncService) processInvalidFormatCustomer(ctx context.Context, user UserForSync, result ComprehensiveCustomerSyncResult) ComprehensiveCustomerSyncResult {
	log.Printf("⚠️ Invalid format Stripe ID for user %d: %s", user.ID, user.StripeCustomerID.String)

	// Try to find by email
	stripeCustomer, err := s.findStripeCustomerByEmail(user.Email)
	if err == nil && stripeCustomer != nil {
		err = s.linkUserToStripeCustomer(user.ID, stripeCustomer.ID)
		if err != nil {
			result.SyncStatus = "error"
			result.Error = fmt.Sprintf("Failed to link: %v", err)
			return result
		}

		result.NewStripeID = stripeCustomer.ID
		result.SyncStatus = "newly_linked"
		return result
	}

	// Clear invalid ID and treat as no Stripe ID
	s.clearGhostStripeID(user.ID)
	result.SyncStatus = "invalid_id_cleared"
	return result
}

// processNoStripeIDCustomer processes a user with no Stripe customer ID
func (s *ComprehensiveStripeSyncService) processNoStripeIDCustomer(ctx context.Context, user UserForSync, result ComprehensiveCustomerSyncResult) ComprehensiveCustomerSyncResult {
	// Try to find existing Stripe customer by email
	stripeCustomer, err := s.findStripeCustomerByEmail(user.Email)
	if err == nil && stripeCustomer != nil {
		err = s.linkUserToStripeCustomer(user.ID, stripeCustomer.ID)
		if err != nil {
			result.SyncStatus = "error"
			result.Error = fmt.Sprintf("Failed to link: %v", err)
			return result
		}

		result.NewStripeID = stripeCustomer.ID
		result.SyncStatus = "newly_linked"
		return result
	}

	// No existing customer - check if they have legacy subscription
	if user.SubID.Valid {
		legacyPlan, _ := s.getLegacyPlanName(user.SubID.Int64)
		result.PlanName = legacyPlan
		result.SyncStatus = "legacy_only"
		return result
	}

	result.SyncStatus = "no_stripe_data"
	return result
}

// Helper functions

func (s *ComprehensiveStripeSyncService) findStripeCustomerByEmail(email string) (*stripe.Customer, error) {
	// Use the existing StripeService to search customers
	// This is a simplified version - you may need to implement customer search in StripeService
	return nil, fmt.Errorf("customer search not implemented - email: %s", email)
}

func (s *ComprehensiveStripeSyncService) linkUserToStripeCustomer(userID int, stripeCustomerID string) error {
	query := `UPDATE users SET stripe_customer_id = $1 WHERE id = $2`
	_, err := s.db.Exec(query, stripeCustomerID, userID)
	return err
}

func (s *ComprehensiveStripeSyncService) clearGhostStripeID(userID int) error {
	query := `UPDATE users SET stripe_customer_id = NULL WHERE id = $1`
	_, err := s.db.Exec(query, userID)
	return err
}

func (s *ComprehensiveStripeSyncService) updateLocalCustomerData(userID int, stripeCustomerID string) error {
	// Update user record with latest Stripe data
	query := `
		UPDATE users 
		SET 
			stripe_customer_id = $1,
			updated_at = NOW()
		WHERE id = $2
	`
	_, err := s.db.Exec(query, stripeCustomerID, userID)
	return err
}

type CustomerPlanInfo struct {
	PlanName       string
	SubscriptionID string
	Status         string
}

func (s *ComprehensiveStripeSyncService) getCustomerPlanInfo(stripeCustomerID string) (*CustomerPlanInfo, error) {
	// Get plan info from local stripe_subscriptions table
	query := `
		SELECT 
			ss.stripe_id,
			ss.status,
			COALESCE(ss.product_name, 'Unknown Plan') as plan_name
		FROM stripe_subscriptions ss
		JOIN stripe_customers sc ON ss.customer_id = sc.id
		WHERE sc.stripe_id = $1
		  AND ss.status IN ('active', 'trialing')
		ORDER BY ss.created_at DESC
		LIMIT 1
	`

	planInfo := &CustomerPlanInfo{}
	err := s.db.QueryRow(query, stripeCustomerID).Scan(
		&planInfo.SubscriptionID,
		&planInfo.Status,
		&planInfo.PlanName,
	)

	if err != nil && err != sql.ErrNoRows {
		return planInfo, err
	}

	return planInfo, nil
}

func (s *ComprehensiveStripeSyncService) getLegacyPlanName(subID int64) (string, error) {
	query := `SELECT name FROM subscription_plans WHERE id = $1 AND is_active = true`
	var planName string
	err := s.db.QueryRow(query, subID).Scan(&planName)
	return planName, err
}
