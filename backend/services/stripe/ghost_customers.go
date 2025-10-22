package stripe

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"bome-backend/infrastructure/database"
)

// GhostCustomersService handles ghost customer detection and management
// ⚠️ NO MOCK DATA - Only identifies and manages existing problematic records
type GhostCustomersService struct {
	db *database.DB
}

// NewGhostCustomersService creates a new ghost customers service
func NewGhostCustomersService(db *database.DB) *GhostCustomersService {
	return &GhostCustomersService{
		db: db,
	}
}

// GhostCustomer represents a customer that exists locally but not in Stripe
type GhostCustomer struct {
	ID                int        `json:"id"`
	LocalCustomerID   *int       `json:"local_customer_id"`
	StripeCustomerID  string     `json:"stripe_customer_id"`
	CustomerEmail     string     `json:"customer_email"`
	CustomerName      string     `json:"customer_name"`
	GhostType         string     `json:"ghost_type"`
	GhostReason       string     `json:"ghost_reason"`
	PurgeStatus       string     `json:"purge_status"`
	DetectionDate     time.Time  `json:"detection_date"`
	Notes             *string    `json:"notes"`
	CurrentStatus     string     `json:"current_status"`
	SubscriptionCount int        `json:"subscription_count"`
	InvoiceCount      int        `json:"invoice_count"`
	PurgedAt          *time.Time `json:"purged_at,omitempty"`
	PurgedBy          *string    `json:"purged_by,omitempty"`
}

// GhostDetectionSummary provides overview of ghost detection results
type GhostDetectionSummary struct {
	TotalGhosts         int `json:"total_ghosts"`
	HashIDGhosts        int `json:"hash_id_ghosts"`
	InvalidFormatGhosts int `json:"invalid_format_ghosts"`
	MarkedForPurge      int `json:"marked_for_purge"`
	AlreadyPurged       int `json:"already_purged"`
}

// GetAllGhosts returns all detected ghost customers
func (s *GhostCustomersService) GetAllGhosts() ([]GhostCustomer, error) {
	log.Println("🔍 [GHOST-SERVICE] Fetching all ghost customers...")

	query := `
		SELECT 
			sg.id, sg.local_customer_id, sg.stripe_customer_id, sg.customer_email,
			sg.customer_name, sg.ghost_type, sg.ghost_reason, sg.purge_status,
			sg.detection_date, sg.notes, sg.purged_at, sg.purged_by,
			CASE WHEN sc.id IS NOT NULL THEN 'exists' ELSE 'already_deleted' END as current_status,
			COALESCE(sub_count.count, 0) as subscription_count,
			COALESCE(inv_count.count, 0) as invoice_count
		FROM stripe_ghosts sg
		LEFT JOIN stripe_customers sc ON sg.stripe_customer_id = sc.stripe_id
		LEFT JOIN (
			SELECT customer_id, COUNT(*) as count 
			FROM stripe_subscriptions 
			GROUP BY customer_id
		) sub_count ON sc.id = sub_count.customer_id
		LEFT JOIN (
			SELECT customer_id, COUNT(*) as count 
			FROM stripe_invoices 
			GROUP BY customer_id
		) inv_count ON sc.id = inv_count.customer_id
		ORDER BY sg.detection_date DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query ghost customers: %w", err)
	}
	defer rows.Close()

	var ghosts []GhostCustomer
	for rows.Next() {
		var ghost GhostCustomer
		err := rows.Scan(
			&ghost.ID, &ghost.LocalCustomerID, &ghost.StripeCustomerID,
			&ghost.CustomerEmail, &ghost.CustomerName, &ghost.GhostType,
			&ghost.GhostReason, &ghost.PurgeStatus, &ghost.DetectionDate,
			&ghost.Notes, &ghost.PurgedAt, &ghost.PurgedBy,
			&ghost.CurrentStatus, &ghost.SubscriptionCount,
			&ghost.InvoiceCount,
		)
		if err != nil {
			log.Printf("⚠️ [GHOST-SERVICE] Error scanning ghost customer: %v", err)
			continue
		}
		ghosts = append(ghosts, ghost)
	}

	log.Printf("✅ [GHOST-SERVICE] Found %d ghost customers", len(ghosts))
	return ghosts, nil
}

// GetGhostSummary returns summary statistics of ghost detection
func (s *GhostCustomersService) GetGhostSummary() (*GhostDetectionSummary, error) {
	log.Println("📊 [GHOST-SERVICE] Fetching ghost detection summary...")

	query := `
		SELECT 
			COUNT(*) as total_ghosts,
			COUNT(*) FILTER (WHERE ghost_reason LIKE '%hash_id%') as hash_id_ghosts,
			COUNT(*) FILTER (WHERE ghost_reason LIKE '%invalid%') as invalid_format_ghosts,
			COUNT(*) FILTER (WHERE purge_status = 'marked_for_purge') as marked_for_purge,
			COUNT(*) FILTER (WHERE purge_status = 'purged') as already_purged
		FROM stripe_ghosts
	`

	var summary GhostDetectionSummary
	err := s.db.QueryRow(query).Scan(
		&summary.TotalGhosts, &summary.HashIDGhosts, &summary.InvalidFormatGhosts,
		&summary.MarkedForPurge, &summary.AlreadyPurged,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get ghost summary: %w", err)
	}

	log.Printf("✅ [GHOST-SERVICE] Summary: %d total ghosts, %d marked for purge",
		summary.TotalGhosts, summary.MarkedForPurge)
	return &summary, nil
}

// DetectedGhostResult represents the result of ghost detection
type DetectedGhostResult struct {
	CustomerID int    `json:"customer_id"`
	StripeID   string `json:"stripe_id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Reason     string `json:"reason"`
}

// RunGhostDetection manually triggers ghost detection using database function
func (s *GhostCustomersService) RunGhostDetection() ([]DetectedGhostResult, error) {
	log.Println("🔍 [GHOST-SERVICE] Running ghost detection...")

	// Call the database function
	query := `SELECT * FROM detect_ghost_customers()`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to run ghost detection: %w", err)
	}
	defer rows.Close()

	var newGhosts []DetectedGhostResult
	for rows.Next() {
		var ghost DetectedGhostResult
		err := rows.Scan(&ghost.CustomerID, &ghost.StripeID, &ghost.Email, &ghost.Name, &ghost.Reason)
		if err != nil {
			log.Printf("⚠️ [GHOST-SERVICE] Error scanning detected ghost: %v", err)
			continue
		}

		// Insert into ghosts table if not already exists
		insertQuery := `
			INSERT INTO stripe_ghosts (
				local_customer_id, stripe_customer_id, customer_email, 
				customer_name, ghost_type, ghost_reason, purge_status
			) VALUES ($1, $2, $3, $4, 'customer', $5, 'detected')
			ON CONFLICT (stripe_customer_id) DO NOTHING
		`

		_, err = s.db.Exec(insertQuery, ghost.CustomerID, ghost.StripeID, ghost.Email, ghost.Name, ghost.Reason)
		if err != nil {
			log.Printf("⚠️ [GHOST-SERVICE] Failed to insert ghost %s: %v", ghost.StripeID, err)
		}

		newGhosts = append(newGhosts, ghost)
	}

	log.Printf("✅ [GHOST-SERVICE] Detected %d ghost customers", len(newGhosts))
	return newGhosts, nil
}

// MarkGhostForPurge marks a ghost customer for purging using database function
func (s *GhostCustomersService) MarkGhostForPurge(stripeCustomerID string, reason string) error {
	log.Printf("⚠️ [GHOST-SERVICE] Marking ghost for purge: %s (reason: %s)", stripeCustomerID, reason)

	if reason == "" {
		reason = "admin_decision"
	}

	// Call the database function
	query := `SELECT mark_customer_for_purge($1, $2)`

	var success bool
	err := s.db.QueryRow(query, stripeCustomerID, reason).Scan(&success)
	if err != nil {
		return fmt.Errorf("failed to mark customer for purge: %w", err)
	}

	if !success {
		return fmt.Errorf("customer not found: %s", stripeCustomerID)
	}

	log.Printf("✅ [GHOST-SERVICE] Marked ghost for purge: %s", stripeCustomerID)
	return nil
}

// PurgeGhostCustomer permanently deletes a ghost customer using database function
// ⚠️ DESTRUCTIVE OPERATION - Cannot be undone!
func (s *GhostCustomersService) PurgeGhostCustomer(stripeCustomerID string, adminUser string) error {
	log.Printf("🗑️ [GHOST-SERVICE] PURGING ghost customer: %s (by: %s)", stripeCustomerID, adminUser)

	if adminUser == "" {
		adminUser = "unknown_admin"
	}

	// Call the database function
	query := `SELECT purge_ghost_customer($1, $2)`

	var success bool
	err := s.db.QueryRow(query, stripeCustomerID, adminUser).Scan(&success)
	if err != nil {
		return fmt.Errorf("failed to purge customer: %w", err)
	}

	if !success {
		return fmt.Errorf("customer not found or purge failed: %s", stripeCustomerID)
	}

	log.Printf("✅ [GHOST-SERVICE] Successfully purged ghost customer: %s", stripeCustomerID)
	return nil
}

// BulkPurgeResult represents the result of a single purge operation in bulk
type BulkPurgeResult struct {
	StripeCustomerID string `json:"stripe_customer_id"`
	Success          bool   `json:"success"`
	Error            string `json:"error,omitempty"`
}

// BulkPurgeGhosts purges multiple ghost customers at once
// ⚠️ DESTRUCTIVE OPERATION - Cannot be undone!
func (s *GhostCustomersService) BulkPurgeGhosts(customerIDs []string, adminUser string) ([]BulkPurgeResult, error) {
	log.Printf("🗑️ [GHOST-SERVICE] Bulk purging %d ghost customers (by: %s)", len(customerIDs), adminUser)

	if adminUser == "" {
		adminUser = "unknown_admin"
	}

	var results []BulkPurgeResult
	successCount := 0

	for _, customerID := range customerIDs {
		result := BulkPurgeResult{
			StripeCustomerID: customerID,
		}

		// Call the database function for each customer
		query := `SELECT purge_ghost_customer($1, $2)`

		var success bool
		err := s.db.QueryRow(query, customerID, adminUser).Scan(&success)

		if err != nil {
			result.Success = false
			result.Error = err.Error()
			log.Printf("❌ [GHOST-SERVICE] Failed to purge %s: %v", customerID, err)
		} else if success {
			result.Success = true
			successCount++
			log.Printf("✅ [GHOST-SERVICE] Purged %s", customerID)
		} else {
			result.Success = false
			result.Error = "Purge returned false - customer may not exist"
			log.Printf("⚠️ [GHOST-SERVICE] Purge failed for %s: not found", customerID)
		}

		results = append(results, result)
	}

	log.Printf("✅ [GHOST-SERVICE] Bulk purge complete: %d/%d successful", successCount, len(customerIDs))
	return results, nil
}

// GetGhostByStripeID gets a specific ghost customer by Stripe ID
func (s *GhostCustomersService) GetGhostByStripeID(stripeCustomerID string) (*GhostCustomer, error) {
	query := `
		SELECT 
			sg.id, sg.local_customer_id, sg.stripe_customer_id, sg.customer_email,
			sg.customer_name, sg.ghost_type, sg.ghost_reason, sg.purge_status,
			sg.detection_date, sg.notes, sg.purged_at, sg.purged_by,
			CASE WHEN sc.id IS NOT NULL THEN 'exists' ELSE 'already_deleted' END as current_status,
			COALESCE(sub_count.count, 0) as subscription_count,
			COALESCE(inv_count.count, 0) as invoice_count
		FROM stripe_ghosts sg
		LEFT JOIN stripe_customers sc ON sg.stripe_customer_id = sc.stripe_id
		LEFT JOIN (
			SELECT customer_id, COUNT(*) as count 
			FROM stripe_subscriptions 
			GROUP BY customer_id
		) sub_count ON sc.id = sub_count.customer_id
		LEFT JOIN (
			SELECT customer_id, COUNT(*) as count 
			FROM stripe_invoices 
			GROUP BY customer_id
		) inv_count ON sc.id = inv_count.customer_id
		WHERE sg.stripe_customer_id = $1
	`

	var ghost GhostCustomer
	err := s.db.QueryRow(query, stripeCustomerID).Scan(
		&ghost.ID, &ghost.LocalCustomerID, &ghost.StripeCustomerID,
		&ghost.CustomerEmail, &ghost.CustomerName, &ghost.GhostType,
		&ghost.GhostReason, &ghost.PurgeStatus, &ghost.DetectionDate,
		&ghost.Notes, &ghost.PurgedAt, &ghost.PurgedBy,
		&ghost.CurrentStatus, &ghost.SubscriptionCount,
		&ghost.InvoiceCount,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("ghost customer not found: %s", stripeCustomerID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get ghost customer: %w", err)
	}

	return &ghost, nil
}

