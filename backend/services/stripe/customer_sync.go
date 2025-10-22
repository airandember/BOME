package stripe

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"bome-backend/infrastructure/database"
	stripeServices "bome-backend/subscription/services"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/customer"
)

// WebSocketHub interface for real-time updates
type WebSocketHub interface {
	BroadcastEvent(eventType string, data map[string]interface{}, message string)
}

// StripeCustomerSyncService handles synchronization between local customers and Stripe
type StripeCustomerSyncService struct {
	db            *database.DB
	stripeService *stripeServices.StripeService
	hub           WebSocketHub
}

// CustomerSyncResult represents the result of a sync operation
type CustomerSyncResult struct {
	CustomerID int       `json:"customer_id"`
	StripeID   string    `json:"stripe_id"`
	Action     string    `json:"action"` // "created", "updated", "synced", "error"
	Message    string    `json:"message"`
	Error      string    `json:"error,omitempty"`
	LastSyncAt time.Time `json:"last_sync_at"`
}

// CustomerSyncStats represents sync operation statistics
type CustomerSyncStats struct {
	TotalProcessed int           `json:"total_processed"`
	Created        int           `json:"created"`
	Updated        int           `json:"updated"`
	Synced         int           `json:"synced"`
	Errors         int           `json:"errors"`
	Duration       time.Duration `json:"duration"`
}

// NewStripeCustomerSyncService creates a new customer sync service
func NewStripeCustomerSyncService(db *database.DB, stripeService *stripeServices.StripeService, hub WebSocketHub) *StripeCustomerSyncService {
	return &StripeCustomerSyncService{
		db:            db,
		stripeService: stripeService,
		hub:           hub,
	}
}

// SyncCustomerToStripe syncs a local customer to Stripe
func (s *StripeCustomerSyncService) SyncCustomerToStripe(customerID int) (*CustomerSyncResult, error) {
	if !s.stripeService.IsEnabled() {
		return nil, fmt.Errorf("stripe service is not enabled")
	}

	// Get customer from database (users table)
	query := `
		SELECT id, email, first_name, last_name, stripe_customer_id
		FROM users
		WHERE id = $1
	`

	var user struct {
		ID               int
		Email            string
		FirstName        sql.NullString
		LastName         sql.NullString
		StripeCustomerID sql.NullString
	}

	err := s.db.DB.QueryRow(query, customerID).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.StripeCustomerID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return &CustomerSyncResult{
				CustomerID: customerID,
				Action:     "error",
				Error:      "Customer not found",
				LastSyncAt: time.Now(),
			}, fmt.Errorf("customer not found")
		}
		return &CustomerSyncResult{
			CustomerID: customerID,
			Action:     "error",
			Error:      fmt.Sprintf("Failed to get customer: %v", err),
			LastSyncAt: time.Now(),
		}, err
	}

	// Check if customer already has Stripe ID
	if user.StripeCustomerID.Valid && user.StripeCustomerID.String != "" {
		// Update existing Stripe customer
		return s.updateStripeCustomer(&user)
	}

	// Create new Stripe customer
	return s.createStripeCustomer(&user)
}

// SyncCustomerFromStripe syncs a Stripe customer to local database
func (s *StripeCustomerSyncService) SyncCustomerFromStripe(stripeCustomerID string) (*CustomerSyncResult, error) {
	if !s.stripeService.IsEnabled() {
		return nil, fmt.Errorf("stripe service is not enabled")
	}

	// Get customer from Stripe
	stripeCustomer, err := customer.Get(stripeCustomerID, nil)
	if err != nil {
		return &CustomerSyncResult{
			StripeID:   stripeCustomerID,
			Action:     "error",
			Error:      fmt.Sprintf("Failed to get Stripe customer: %v", err),
			LastSyncAt: time.Now(),
		}, err
	}

	// Check if customer exists locally by email
	var localID int
	var existingStripeID sql.NullString
	err = s.db.DB.QueryRow("SELECT id, stripe_customer_id FROM users WHERE email = $1", stripeCustomer.Email).Scan(&localID, &existingStripeID)

	if err == sql.ErrNoRows {
		// Customer doesn't exist locally, store in stripe_customers table
		return s.storeStripeCustomer(stripeCustomer)
	}

	if err != nil {
		return &CustomerSyncResult{
			StripeID:   stripeCustomerID,
			Action:     "error",
			Error:      fmt.Sprintf("Database error: %v", err),
			LastSyncAt: time.Now(),
		}, err
	}

	// Customer exists, update their Stripe ID
	return s.linkLocalCustomer(localID, stripeCustomer)
}

// BulkSyncCustomersToStripe syncs multiple customers to Stripe
func (s *StripeCustomerSyncService) BulkSyncCustomersToStripe(customerIDs []int) (*CustomerSyncStats, error) {
	if !s.stripeService.IsEnabled() {
		return nil, fmt.Errorf("stripe service is not enabled")
	}

	startTime := time.Now()
	stats := &CustomerSyncStats{}

	// Broadcast start event
	if s.hub != nil {
		s.hub.BroadcastEvent("stripe.bulk_sync.started", map[string]interface{}{
			"total":     len(customerIDs),
			"direction": "to_stripe",
		}, fmt.Sprintf("Starting bulk sync: %d customers to Stripe", len(customerIDs)))
	}

	for i, customerID := range customerIDs {
		result, err := s.SyncCustomerToStripe(customerID)
		if err != nil {
			stats.Errors++
			log.Printf("Error syncing customer %d: %v", customerID, err)
			continue
		}

		stats.TotalProcessed++
		switch result.Action {
		case "created":
			stats.Created++
		case "updated":
			stats.Updated++
		case "synced":
			stats.Synced++
		case "error":
			stats.Errors++
		}

		// Broadcast progress every 10 customers
		if s.hub != nil && (i+1)%10 == 0 {
			s.hub.BroadcastEvent("stripe.bulk_sync.progress", map[string]interface{}{
				"processed": i + 1,
				"total":     len(customerIDs),
				"created":   stats.Created,
				"updated":   stats.Updated,
				"errors":    stats.Errors,
			}, fmt.Sprintf("Synced %d/%d customers", i+1, len(customerIDs)))
		}
	}

	stats.Duration = time.Since(startTime)

	// Broadcast completion
	if s.hub != nil {
		s.hub.BroadcastEvent("stripe.bulk_sync.completed", map[string]interface{}{
			"stats": stats,
		}, fmt.Sprintf("Bulk sync completed: %d created, %d updated, %d errors", stats.Created, stats.Updated, stats.Errors))
	}

	return stats, nil
}

// BulkSyncCustomersFromStripe syncs multiple customers from Stripe
func (s *StripeCustomerSyncService) BulkSyncCustomersFromStripe(stripeCustomerIDs []string) (*CustomerSyncStats, error) {
	if !s.stripeService.IsEnabled() {
		return nil, fmt.Errorf("stripe service is not enabled")
	}

	startTime := time.Now()
	stats := &CustomerSyncStats{}

	// Broadcast start event
	if s.hub != nil {
		s.hub.BroadcastEvent("stripe.bulk_sync.started", map[string]interface{}{
			"total":     len(stripeCustomerIDs),
			"direction": "from_stripe",
		}, fmt.Sprintf("Starting bulk sync: %d customers from Stripe", len(stripeCustomerIDs)))
	}

	for i, stripeID := range stripeCustomerIDs {
		result, err := s.SyncCustomerFromStripe(stripeID)
		if err != nil {
			stats.Errors++
			log.Printf("Error syncing Stripe customer %s: %v", stripeID, err)
			continue
		}

		stats.TotalProcessed++
		switch result.Action {
		case "created":
			stats.Created++
		case "updated":
			stats.Updated++
		case "synced":
			stats.Synced++
		case "error":
			stats.Errors++
		}

		// Broadcast progress every 10 customers
		if s.hub != nil && (i+1)%10 == 0 {
			s.hub.BroadcastEvent("stripe.bulk_sync.progress", map[string]interface{}{
				"processed": i + 1,
				"total":     len(stripeCustomerIDs),
				"created":   stats.Created,
				"updated":   stats.Updated,
				"errors":    stats.Errors,
			}, fmt.Sprintf("Synced %d/%d customers", i+1, len(stripeCustomerIDs)))
		}
	}

	stats.Duration = time.Since(startTime)

	// Broadcast completion
	if s.hub != nil {
		s.hub.BroadcastEvent("stripe.bulk_sync.completed", map[string]interface{}{
			"stats": stats,
		}, fmt.Sprintf("Bulk sync completed: %d created, %d updated, %d errors", stats.Created, stats.Updated, stats.Errors))
	}

	return stats, nil
}

// GetSyncStatus gets the sync status for a customer
func (s *StripeCustomerSyncService) GetSyncStatus(customerID int) (*CustomerSyncResult, error) {
	query := `
		SELECT id, email, stripe_customer_id, updated_at
		FROM users
		WHERE id = $1
	`

	var user struct {
		ID               int
		Email            string
		StripeCustomerID sql.NullString
		UpdatedAt        time.Time
	}

	err := s.db.DB.QueryRow(query, customerID).Scan(&user.ID, &user.Email, &user.StripeCustomerID, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return &CustomerSyncResult{
				CustomerID: customerID,
				Action:     "error",
				Error:      "Customer not found",
				LastSyncAt: time.Now(),
			}, fmt.Errorf("customer not found")
		}
		return nil, err
	}

	result := &CustomerSyncResult{
		CustomerID: user.ID,
		LastSyncAt: user.UpdatedAt,
	}

	if user.StripeCustomerID.Valid && user.StripeCustomerID.String != "" {
		result.StripeID = user.StripeCustomerID.String
		result.Action = "synced"
		result.Message = "Customer is synced with Stripe"
	} else {
		result.Action = "not_synced"
		result.Message = "Customer not synced to Stripe yet"
	}

	return result, nil
}

// Helper functions

func (s *StripeCustomerSyncService) createStripeCustomer(user interface{}) (*CustomerSyncResult, error) {
	u := user.(*struct {
		ID               int
		Email            string
		FirstName        sql.NullString
		LastName         sql.NullString
		StripeCustomerID sql.NullString
	})

	// Build customer name
	name := ""
	if u.FirstName.Valid {
		name = u.FirstName.String
	}
	if u.LastName.Valid {
		if name != "" {
			name += " "
		}
		name += u.LastName.String
	}

	// Create customer in Stripe
	params := &stripe.CustomerParams{
		Email: stripe.String(u.Email),
	}
	if name != "" {
		params.Name = stripe.String(name)
	}

	stripeCustomer, err := customer.New(params)
	if err != nil {
		return &CustomerSyncResult{
			CustomerID: u.ID,
			Action:     "error",
			Error:      fmt.Sprintf("Failed to create Stripe customer: %v", err),
			LastSyncAt: time.Now(),
		}, err
	}

	// Update local user with Stripe ID
	_, err = s.db.DB.Exec("UPDATE users SET stripe_customer_id = $1, updated_at = NOW() WHERE id = $2", stripeCustomer.ID, u.ID)
	if err != nil {
		log.Printf("Warning: Created Stripe customer %s but failed to update local record: %v", stripeCustomer.ID, err)
	}

	// Store in stripe_customers table
	s.storeStripeCustomer(stripeCustomer)

	result := &CustomerSyncResult{
		CustomerID: u.ID,
		StripeID:   stripeCustomer.ID,
		Action:     "created",
		Message:    fmt.Sprintf("Created Stripe customer: %s", stripeCustomer.ID),
		LastSyncAt: time.Now(),
	}

	// Broadcast event
	if s.hub != nil {
		s.hub.BroadcastEvent("stripe.customer.synced", map[string]interface{}{
			"customer_id": u.ID,
			"stripe_id":   stripeCustomer.ID,
			"action":      "created",
		}, fmt.Sprintf("Customer synced to Stripe: %s", u.Email))
	}

	return result, nil
}

func (s *StripeCustomerSyncService) updateStripeCustomer(user interface{}) (*CustomerSyncResult, error) {
	u := user.(*struct {
		ID               int
		Email            string
		FirstName        sql.NullString
		LastName         sql.NullString
		StripeCustomerID sql.NullString
	})

	// Build customer name
	name := ""
	if u.FirstName.Valid {
		name = u.FirstName.String
	}
	if u.LastName.Valid {
		if name != "" {
			name += " "
		}
		name += u.LastName.String
	}

	// Update customer in Stripe
	params := &stripe.CustomerParams{
		Email: stripe.String(u.Email),
	}
	if name != "" {
		params.Name = stripe.String(name)
	}

	stripeCustomer, err := customer.Update(u.StripeCustomerID.String, params)
	if err != nil {
		return &CustomerSyncResult{
			CustomerID: u.ID,
			StripeID:   u.StripeCustomerID.String,
			Action:     "error",
			Error:      fmt.Sprintf("Failed to update Stripe customer: %v", err),
			LastSyncAt: time.Now(),
		}, err
	}

	// Update stripe_customers table
	s.storeStripeCustomer(stripeCustomer)

	result := &CustomerSyncResult{
		CustomerID: u.ID,
		StripeID:   stripeCustomer.ID,
		Action:     "updated",
		Message:    fmt.Sprintf("Updated Stripe customer: %s", stripeCustomer.ID),
		LastSyncAt: time.Now(),
	}

	// Broadcast event
	if s.hub != nil {
		s.hub.BroadcastEvent("stripe.customer.synced", map[string]interface{}{
			"customer_id": u.ID,
			"stripe_id":   stripeCustomer.ID,
			"action":      "updated",
		}, fmt.Sprintf("Customer updated in Stripe: %s", u.Email))
	}

	return result, nil
}

func (s *StripeCustomerSyncService) storeStripeCustomer(stripeCustomer *stripe.Customer) (*CustomerSyncResult, error) {
	// Store/update in stripe_customers table
	query := `
		INSERT INTO stripe_customers (stripe_id, email, name, created_at, updated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (stripe_id) 
		DO UPDATE SET email = $2, name = $3, updated_at = $5, metadata = $6
	`

	metadata := "{}"
	if stripeCustomer.Metadata != nil && len(stripeCustomer.Metadata) > 0 {
		if metadataBytes, err := json.Marshal(stripeCustomer.Metadata); err == nil {
			metadata = string(metadataBytes)
		}
	}

	_, err := s.db.DB.Exec(query,
		stripeCustomer.ID,
		stripeCustomer.Email,
		stripeCustomer.Name,
		time.Unix(stripeCustomer.Created, 0),
		time.Now(),
		metadata,
	)

	if err != nil {
		log.Printf("Warning: Failed to store Stripe customer in database: %v", err)
	}

	return &CustomerSyncResult{
		StripeID:   stripeCustomer.ID,
		Action:     "created",
		Message:    fmt.Sprintf("Stored Stripe customer: %s", stripeCustomer.Email),
		LastSyncAt: time.Now(),
	}, nil
}

func (s *StripeCustomerSyncService) linkLocalCustomer(localID int, stripeCustomer *stripe.Customer) (*CustomerSyncResult, error) {
	// Update local user with Stripe ID
	_, err := s.db.DB.Exec("UPDATE users SET stripe_customer_id = $1, updated_at = NOW() WHERE id = $2", stripeCustomer.ID, localID)
	if err != nil {
		return &CustomerSyncResult{
			CustomerID: localID,
			StripeID:   stripeCustomer.ID,
			Action:     "error",
			Error:      fmt.Sprintf("Failed to link customer: %v", err),
			LastSyncAt: time.Now(),
		}, err
	}

	// Store in stripe_customers table
	s.storeStripeCustomer(stripeCustomer)

	result := &CustomerSyncResult{
		CustomerID: localID,
		StripeID:   stripeCustomer.ID,
		Action:     "updated",
		Message:    fmt.Sprintf("Linked customer to Stripe: %s", stripeCustomer.Email),
		LastSyncAt: time.Now(),
	}

	// Broadcast event
	if s.hub != nil {
		s.hub.BroadcastEvent("stripe.customer.synced", map[string]interface{}{
			"customer_id": localID,
			"stripe_id":   stripeCustomer.ID,
			"action":      "linked",
		}, fmt.Sprintf("Customer linked to Stripe: %s", stripeCustomer.Email))
	}

	return result, nil
}
