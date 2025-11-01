package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"bome-backend/internal/database"
)

// GhostTrackingService handles tracking and managing ghost Stripe data
type GhostTrackingService struct {
	db *database.DB
}

// NewGhostTrackingService creates a new ghost tracking service
func NewGhostTrackingService(db *database.DB) *GhostTrackingService {
	return &GhostTrackingService{
		db: db,
	}
}

// GhostEntry represents a ghost data entry
type GhostEntry struct {
	ID              int                    `json:"id"`
	GhostType       string                 `json:"ghost_type"`
	StripeID        string                 `json:"stripe_id"`
	GhostReason     string                 `json:"ghost_reason"`
	ReferencedBy    map[string]interface{} `json:"referenced_by"`
	FirstDetectedAt time.Time              `json:"first_detected_at"`
	LastSeenAt      time.Time              `json:"last_seen_at"`
	AttemptedSyncs  int                    `json:"attempted_syncs"`
	Metadata        map[string]interface{} `json:"metadata"`
	Notes           string                 `json:"notes"`
}

// GhostReport contains categorized ghost data
type GhostReport struct {
	TotalGhosts        int          `json:"total_ghosts"`
	GhostProducts      []GhostEntry `json:"ghost_products"`
	GhostPrices        []GhostEntry `json:"ghost_prices"`
	GhostSubscriptions []GhostEntry `json:"ghost_subscriptions"`
	GhostCustomers     []GhostEntry `json:"ghost_customers"`
	LastUpdated        time.Time    `json:"last_updated"`
}

// ================================================================
// LOGGING METHODS
// ================================================================

// LogGhostProduct logs a ghost product that was blocked from syncing
func (s *GhostTrackingService) LogGhostProduct(ctx context.Context, productID string, reason string, metadata map[string]interface{}) error {
	log.Printf("👻 [Ghost Tracking] Logging ghost product: %s", productID)

	metadataJSON, _ := json.Marshal(metadata)

	query := `
		INSERT INTO stripe_ghosts_v2 (ghost_type, stripe_id, ghost_reason, metadata, first_detected_at, last_seen_at, attempted_syncs)
		VALUES ($1, $2, $3, $4, NOW(), NOW(), 1)
		ON CONFLICT (ghost_type, stripe_id) 
		DO UPDATE SET 
			last_seen_at = NOW(),
			attempted_syncs = stripe_ghosts_v2.attempted_syncs + 1,
			metadata = EXCLUDED.metadata,
			ghost_reason = EXCLUDED.ghost_reason
	`

	_, err := s.db.Exec(query, "product", productID, reason, metadataJSON)
	if err != nil {
		log.Printf("❌ Failed to log ghost product %s: %v", productID, err)
		return err
	}

	log.Printf("✅ [Ghost Tracking] Logged ghost product: %s", productID)
	return nil
}

// LogGhostPrice logs a ghost price that was blocked from syncing
func (s *GhostTrackingService) LogGhostPrice(ctx context.Context, priceID string, ghostProductID string, metadata map[string]interface{}) error {
	log.Printf("👻 [Ghost Tracking] Logging ghost price: %s (references ghost product: %s)", priceID, ghostProductID)

	reason := fmt.Sprintf("References deleted product %s", ghostProductID)

	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["ghost_product_id"] = ghostProductID

	metadataJSON, _ := json.Marshal(metadata)
	referencedByJSON, _ := json.Marshal(map[string]interface{}{
		"ghost_product": ghostProductID,
	})

	query := `
		INSERT INTO stripe_ghosts_v2 (ghost_type, stripe_id, ghost_reason, referenced_by, metadata, first_detected_at, last_seen_at, attempted_syncs)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW(), 1)
		ON CONFLICT (ghost_type, stripe_id) 
		DO UPDATE SET 
			last_seen_at = NOW(),
			attempted_syncs = stripe_ghosts_v2.attempted_syncs + 1,
			metadata = EXCLUDED.metadata,
			referenced_by = EXCLUDED.referenced_by,
			ghost_reason = EXCLUDED.ghost_reason
	`

	_, err := s.db.Exec(query, "price", priceID, reason, referencedByJSON, metadataJSON)
	if err != nil {
		log.Printf("❌ Failed to log ghost price %s: %v", priceID, err)
		return err
	}

	log.Printf("✅ [Ghost Tracking] Logged ghost price: %s", priceID)
	return nil
}

// LogGhostSubscription logs a ghost subscription that was blocked from syncing
func (s *GhostTrackingService) LogGhostSubscription(ctx context.Context, subID string, ghostProductID string, customerID string, customerEmail string, metadata map[string]interface{}) error {
	log.Printf("👻 [Ghost Tracking] Logging ghost subscription: %s (customer: %s, ghost product: %s)", subID, customerEmail, ghostProductID)

	reason := fmt.Sprintf("References deleted product %s", ghostProductID)

	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["ghost_product_id"] = ghostProductID

	metadataJSON, _ := json.Marshal(metadata)
	referencedByJSON, _ := json.Marshal(map[string]interface{}{
		"customer_id":    customerID,
		"customer_email": customerEmail,
		"ghost_product":  ghostProductID,
	})

	notes := fmt.Sprintf("Customer %s is being charged for a subscription that references a deleted product", customerEmail)

	query := `
		INSERT INTO stripe_ghosts_v2 (ghost_type, stripe_id, ghost_reason, referenced_by, metadata, notes, first_detected_at, last_seen_at, attempted_syncs)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW(), 1)
		ON CONFLICT (ghost_type, stripe_id) 
		DO UPDATE SET 
			last_seen_at = NOW(),
			attempted_syncs = stripe_ghosts_v2.attempted_syncs + 1,
			metadata = EXCLUDED.metadata,
			referenced_by = EXCLUDED.referenced_by,
			notes = EXCLUDED.notes,
			ghost_reason = EXCLUDED.ghost_reason
	`

	_, err := s.db.Exec(query, "subscription", subID, reason, referencedByJSON, metadataJSON, notes)
	if err != nil {
		log.Printf("❌ Failed to log ghost subscription %s: %v", subID, err)
		return err
	}

	log.Printf("✅ [Ghost Tracking] Logged ghost subscription: %s", subID)
	return nil
}

// LogGhostCustomer logs a ghost customer that was blocked from syncing
func (s *GhostTrackingService) LogGhostCustomer(ctx context.Context, customerID string, reason string, metadata map[string]interface{}) error {
	log.Printf("👻 [Ghost Tracking] Logging ghost customer: %s", customerID)

	metadataJSON, _ := json.Marshal(metadata)

	query := `
		INSERT INTO stripe_ghosts_v2 (ghost_type, stripe_id, ghost_reason, metadata, first_detected_at, last_seen_at, attempted_syncs)
		VALUES ($1, $2, $3, $4, NOW(), NOW(), 1)
		ON CONFLICT (ghost_type, stripe_id) 
		DO UPDATE SET 
			last_seen_at = NOW(),
			attempted_syncs = stripe_ghosts_v2.attempted_syncs + 1,
			metadata = EXCLUDED.metadata,
			ghost_reason = EXCLUDED.ghost_reason
	`

	_, err := s.db.Exec(query, "customer", customerID, reason, metadataJSON)
	if err != nil {
		log.Printf("❌ Failed to log ghost customer %s: %v", customerID, err)
		return err
	}

	log.Printf("✅ [Ghost Tracking] Logged ghost customer: %s", customerID)
	return nil
}

// ================================================================
// RETRIEVAL METHODS
// ================================================================

// GetAllGhosts retrieves all ghost data categorized by type
func (s *GhostTrackingService) GetAllGhosts(ctx context.Context) (*GhostReport, error) {
	log.Printf("📊 [Ghost Tracking] Fetching all ghost data")

	report := &GhostReport{
		GhostProducts:      []GhostEntry{},
		GhostPrices:        []GhostEntry{},
		GhostSubscriptions: []GhostEntry{},
		GhostCustomers:     []GhostEntry{},
		LastUpdated:        time.Now(),
	}

	query := `
		SELECT 
			id, 
			ghost_type, 
			stripe_id, 
			ghost_reason, 
			COALESCE(referenced_by, '{}'), 
			first_detected_at, 
			last_seen_at, 
			attempted_syncs, 
			COALESCE(metadata, '{}'), 
			COALESCE(notes, '')
		FROM stripe_ghosts_v2
		ORDER BY last_seen_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		log.Printf("❌ Failed to fetch ghost data: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var entry GhostEntry
		var referencedByJSON, metadataJSON []byte

		err := rows.Scan(
			&entry.ID,
			&entry.GhostType,
			&entry.StripeID,
			&entry.GhostReason,
			&referencedByJSON,
			&entry.FirstDetectedAt,
			&entry.LastSeenAt,
			&entry.AttemptedSyncs,
			&metadataJSON,
			&entry.Notes,
		)
		if err != nil {
			log.Printf("⚠️  Failed to scan ghost entry: %v", err)
			continue
		}

		// Parse JSON fields
		json.Unmarshal(referencedByJSON, &entry.ReferencedBy)
		json.Unmarshal(metadataJSON, &entry.Metadata)

		// Categorize by type
		switch entry.GhostType {
		case "product":
			report.GhostProducts = append(report.GhostProducts, entry)
		case "price":
			report.GhostPrices = append(report.GhostPrices, entry)
		case "subscription":
			report.GhostSubscriptions = append(report.GhostSubscriptions, entry)
		case "customer":
			report.GhostCustomers = append(report.GhostCustomers, entry)
		}

		report.TotalGhosts++
	}

	log.Printf("✅ [Ghost Tracking] Found %d total ghosts: %d products, %d prices, %d subscriptions, %d customers",
		report.TotalGhosts,
		len(report.GhostProducts),
		len(report.GhostPrices),
		len(report.GhostSubscriptions),
		len(report.GhostCustomers),
	)

	return report, nil
}

// GetGhostCount returns just the count of active ghosts
func (s *GhostTrackingService) GetGhostCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM stripe_ghosts_v2").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// ================================================================
// RESOLUTION METHODS
// ================================================================

// RemoveGhost removes a ghost entry (called when webhook succeeds)
func (s *GhostTrackingService) RemoveGhost(ctx context.Context, ghostType string, stripeID string) error {
	log.Printf("🧹 [Ghost Tracking] Removing resolved ghost: %s %s", ghostType, stripeID)

	query := `DELETE FROM stripe_ghosts_v2 WHERE ghost_type = $1 AND stripe_id = $2`
	result, err := s.db.Exec(query, ghostType, stripeID)
	if err != nil {
		log.Printf("❌ Failed to remove ghost %s %s: %v", ghostType, stripeID, err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("✅ [Ghost Tracking] Removed ghost: %s %s (auto-resolved)", ghostType, stripeID)
	}

	return nil
}

// CheckAndRemoveGhostProduct checks if a product sync should remove it from ghosts
func (s *GhostTrackingService) CheckAndRemoveGhostProduct(ctx context.Context, productID string) error {
	return s.RemoveGhost(ctx, "product", productID)
}

// CheckAndRemoveGhostPrice checks if a price sync should remove it from ghosts
func (s *GhostTrackingService) CheckAndRemoveGhostPrice(ctx context.Context, priceID string) error {
	return s.RemoveGhost(ctx, "price", priceID)
}

// CheckAndRemoveGhostSubscription checks if a subscription sync should remove it from ghosts
func (s *GhostTrackingService) CheckAndRemoveGhostSubscription(ctx context.Context, subscriptionID string) error {
	return s.RemoveGhost(ctx, "subscription", subscriptionID)
}

// CheckAndRemoveGhostCustomer checks if a customer sync should remove it from ghosts
func (s *GhostTrackingService) CheckAndRemoveGhostCustomer(ctx context.Context, customerID string) error {
	return s.RemoveGhost(ctx, "customer", customerID)
}

// ================================================================
// UTILITY METHODS
// ================================================================

// IsGhostProduct checks if a product ID is in the ghost table
func (s *GhostTrackingService) IsGhostProduct(ctx context.Context, productID string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM stripe_ghosts_v2 WHERE ghost_type = 'product' AND stripe_id = $1)",
		productID,
	).Scan(&exists)
	return exists, err
}

// GetGhostsByType retrieves ghosts of a specific type
func (s *GhostTrackingService) GetGhostsByType(ctx context.Context, ghostType string) ([]GhostEntry, error) {
	var entries []GhostEntry

	query := `
		SELECT 
			id, 
			ghost_type, 
			stripe_id, 
			ghost_reason, 
			COALESCE(referenced_by, '{}'), 
			first_detected_at, 
			last_seen_at, 
			attempted_syncs, 
			COALESCE(metadata, '{}'), 
			COALESCE(notes, '')
		FROM stripe_ghosts_v2
		WHERE ghost_type = $1
		ORDER BY last_seen_at DESC
	`

	rows, err := s.db.Query(query, ghostType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var entry GhostEntry
		var referencedByJSON, metadataJSON []byte

		err := rows.Scan(
			&entry.ID,
			&entry.GhostType,
			&entry.StripeID,
			&entry.GhostReason,
			&referencedByJSON,
			&entry.FirstDetectedAt,
			&entry.LastSeenAt,
			&entry.AttemptedSyncs,
			&metadataJSON,
			&entry.Notes,
		)
		if err != nil {
			continue
		}

		json.Unmarshal(referencedByJSON, &entry.ReferencedBy)
		json.Unmarshal(metadataJSON, &entry.Metadata)

		entries = append(entries, entry)
	}

	return entries, nil
}
