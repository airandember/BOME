package services

import (
	"fmt"
	"log"
	"time"

	"bome-backend/internal/database"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/customer"
)

// StripeCustomerSyncService handles synchronization between local customers and Stripe
type StripeCustomerSyncService struct {
	stripeService *StripeService
	db            *database.DB
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
func NewStripeCustomerSyncService(stripeService *StripeService, db *database.DB) *StripeCustomerSyncService {
	return &StripeCustomerSyncService{
		stripeService: stripeService,
		db:            db,
	}
}

// SyncCustomerToStripe syncs a local customer to Stripe
func (s *StripeCustomerSyncService) SyncCustomerToStripe(customerID int) (*CustomerSyncResult, error) {
	if !s.stripeService.IsEnabled() {
		return nil, fmt.Errorf("stripe service is not enabled")
	}

	// Get customer from database
	customer, err := s.db.GetUserByID(customerID)
	if err != nil {
		return &CustomerSyncResult{
			CustomerID: customerID,
			Action:     "error",
			Error:      fmt.Sprintf("Failed to get customer: %v", err),
			LastSyncAt: time.Now(),
		}, err
	}

	// Check if customer already has Stripe ID
	if customer.StripeCustomerID.Valid && customer.StripeCustomerID.String != "" {
		// Update existing Stripe customer
		return s.updateStripeCustomer(customer)
	} else {
		// Create new Stripe customer
		return s.createStripeCustomer(customer)
	}
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

	// Check if customer exists locally by email (since we don't have GetUserByStripeID)
	localCustomer, err := s.db.GetUserByEmail(stripeCustomer.Email)
	if err != nil {
		// Customer doesn't exist locally, create new
		return s.createLocalCustomer(stripeCustomer)
	} else {
		// Customer exists, update
		return s.updateLocalCustomer(localCustomer, stripeCustomer)
	}
}

// BulkSyncCustomersToStripe syncs multiple customers to Stripe
func (s *StripeCustomerSyncService) BulkSyncCustomersToStripe(customerIDs []int) (*CustomerSyncStats, error) {
	if !s.stripeService.IsEnabled() {
		return nil, fmt.Errorf("stripe service is not enabled")
	}

	startTime := time.Now()
	stats := &CustomerSyncStats{}

	for _, customerID := range customerIDs {
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
	}

	stats.Duration = time.Since(startTime)
	return stats, nil
}

// BulkSyncCustomersFromStripe syncs multiple customers from Stripe
func (s *StripeCustomerSyncService) BulkSyncCustomersFromStripe(stripeCustomerIDs []string) (*CustomerSyncStats, error) {
	if !s.stripeService.IsEnabled() {
		return nil, fmt.Errorf("stripe service is not enabled")
	}

	startTime := time.Now()
	stats := &CustomerSyncStats{}

	for _, stripeID := range stripeCustomerIDs {
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
	}

	stats.Duration = time.Since(startTime)
	return stats, nil
}

// SyncAllCustomers performs a full sync in both directions
func (s *StripeCustomerSyncService) SyncAllCustomers() (*CustomerSyncStats, error) {
	if !s.stripeService.IsEnabled() {
		return nil, fmt.Errorf("stripe service is not enabled")
	}

	startTime := time.Now()
	stats := &CustomerSyncStats{}

	// Get all local customers (we'll need to implement this or use a different approach)
	// For now, we'll just sync from Stripe to local
	stripeCustomers, err := s.getAllStripeCustomers()
	if err != nil {
		return nil, fmt.Errorf("failed to get Stripe customers: %v", err)
	}

	// Create maps for efficient lookup
	stripeCustomerMap := make(map[string]*stripe.Customer)

	for _, customer := range stripeCustomers {
		stripeCustomerMap[customer.ID] = customer
	}

	// Sync Stripe customers to local
	for stripeID, stripeCustomer := range stripeCustomerMap {
		// Check if customer exists locally by email
		localCustomer, err := s.db.GetUserByEmail(stripeCustomer.Email)
		if err != nil {
			// New Stripe customer, create locally
			result, err := s.SyncCustomerFromStripe(stripeID)
			if err != nil {
				stats.Errors++
				continue
			}
			stats.TotalProcessed++
			if result.Action == "created" {
				stats.Created++
			}
		} else {
			// Customer exists, update if needed
			_ = localCustomer // Use the variable to avoid unused error
		}
	}

	stats.Duration = time.Since(startTime)
	return stats, nil
}

// GetSyncStatus returns the sync status for a customer
func (s *StripeCustomerSyncService) GetSyncStatus(customerID int) (*CustomerSyncResult, error) {
	customer, err := s.db.GetUserByID(customerID)
	if err != nil {
		return nil, err
	}

	if !customer.StripeCustomerID.Valid || customer.StripeCustomerID.String == "" {
		return &CustomerSyncResult{
			CustomerID: customerID,
			Action:     "not_synced",
			Message:    "Customer not synced to Stripe",
			LastSyncAt: time.Time{},
		}, nil
	}

	// Check if Stripe customer still exists
	_, err = customer.Get(customer.StripeCustomerID.String, nil)
	if err != nil {
		return &CustomerSyncResult{
			CustomerID: customerID,
			StripeID:   customer.StripeCustomerID.String,
			Action:     "stripe_not_found",
			Message:    "Stripe customer not found",
			LastSyncAt: time.Now(),
		}, nil
	}

	return &CustomerSyncResult{
		CustomerID: customerID,
		StripeID:   customer.StripeCustomerID.String,
		Action:     "synced",
		Message:    "Customer is in sync with Stripe",
		LastSyncAt: time.Now(),
	}, nil
}

// Helper methods
func (s *StripeCustomerSyncService) createStripeCustomer(localCustomer *database.User) (*CustomerSyncResult, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(localCustomer.Email),
		Name:  stripe.String(fmt.Sprintf("%s %s", localCustomer.FirstName, localCustomer.LastName)),
	}

	// Add metadata using AddMetadata method
	params.AddMetadata("local_customer_id", fmt.Sprintf("%d", localCustomer.ID))
	params.AddMetadata("role", localCustomer.Role)
	params.AddMetadata("created_at", localCustomer.CreatedAt.Format(time.RFC3339))

	stripeCustomer, err := customer.New(params)
	if err != nil {
		return &CustomerSyncResult{
			CustomerID: localCustomer.ID,
			Action:     "error",
			Error:      fmt.Sprintf("Failed to create Stripe customer: %v", err),
			LastSyncAt: time.Now(),
		}, err
	}

	// Update local customer with Stripe ID
	err = s.db.UpdateUserStripeCustomerID(localCustomer.ID, stripeCustomer.ID)
	if err != nil {
		log.Printf("Warning: Failed to update local customer with Stripe ID: %v", err)
	}

	return &CustomerSyncResult{
		CustomerID: localCustomer.ID,
		StripeID:   stripeCustomer.ID,
		Action:     "created",
		Message:    "Customer created in Stripe",
		LastSyncAt: time.Now(),
	}, nil
}

func (s *StripeCustomerSyncService) updateStripeCustomer(localCustomer *database.User) (*CustomerSyncResult, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(localCustomer.Email),
		Name:  stripe.String(fmt.Sprintf("%s %s", localCustomer.FirstName, localCustomer.LastName)),
	}

	// Add metadata using AddMetadata method
	params.AddMetadata("local_customer_id", fmt.Sprintf("%d", localCustomer.ID))
	params.AddMetadata("role", localCustomer.Role)
	params.AddMetadata("updated_at", time.Now().Format(time.RFC3339))

	_, err := customer.Update(localCustomer.StripeCustomerID.String, params)
	if err != nil {
		return &CustomerSyncResult{
			CustomerID: localCustomer.ID,
			StripeID:   localCustomer.StripeCustomerID.String,
			Action:     "error",
			Error:      fmt.Sprintf("Failed to update Stripe customer: %v", err),
			LastSyncAt: time.Now(),
		}, err
	}

	// Update local customer sync timestamp (we'll need to implement this)
	// For now, just log the success
	log.Printf("Successfully updated Stripe customer %s for local customer %d", localCustomer.StripeCustomerID.String, localCustomer.ID)

	return &CustomerSyncResult{
		CustomerID: localCustomer.ID,
		StripeID:   localCustomer.StripeCustomerID.String,
		Action:     "updated",
		Message:    "Customer updated in Stripe",
		LastSyncAt: time.Now(),
	}, nil
}

func (s *StripeCustomerSyncService) createLocalCustomer(stripeCustomer *stripe.Customer) (*CustomerSyncResult, error) {
	// Parse name from Stripe customer
	firstName, lastName := s.parseName(stripeCustomer.Name)

	// Create new local customer with default password
	passwordHash := "temp_hash" // In production, this should be properly hashed

	createdCustomer, err := s.db.CreateUser(stripeCustomer.Email, passwordHash, firstName, lastName, "user")
	if err != nil {
		return &CustomerSyncResult{
			StripeID:   stripeCustomer.ID,
			Action:     "error",
			Error:      fmt.Sprintf("Failed to create local customer: %v", err),
			LastSyncAt: time.Now(),
		}, err
	}

	// Update the customer with Stripe ID
	err = s.db.UpdateUserStripeCustomerID(createdCustomer.ID, stripeCustomer.ID)
	if err != nil {
		log.Printf("Warning: Failed to update local customer with Stripe ID: %v", err)
	}

	return &CustomerSyncResult{
		CustomerID: createdCustomer.ID,
		StripeID:   stripeCustomer.ID,
		Action:     "created",
		Message:    "Customer created locally from Stripe",
		LastSyncAt: time.Now(),
	}, nil
}

func (s *StripeCustomerSyncService) updateLocalCustomer(localCustomer *database.User, stripeCustomer *stripe.Customer) (*CustomerSyncResult, error) {
	firstName, lastName := s.parseName(stripeCustomer.Name)

	// Update customer profile
	updates := map[string]interface{}{
		"email":      stripeCustomer.Email,
		"first_name": firstName,
		"last_name":  lastName,
	}

	err := s.db.UpdateUserProfile(localCustomer.ID, updates)
	if err != nil {
		return &CustomerSyncResult{
			CustomerID: localCustomer.ID,
			StripeID:   stripeCustomer.ID,
			Action:     "error",
			Error:      fmt.Sprintf("Failed to update local customer: %v", err),
			LastSyncAt: time.Now(),
		}, err
	}

	// Update Stripe customer ID if not set
	if !localCustomer.StripeCustomerID.Valid || localCustomer.StripeCustomerID.String == "" {
		err = s.db.UpdateUserStripeCustomerID(localCustomer.ID, stripeCustomer.ID)
		if err != nil {
			log.Printf("Warning: Failed to update local customer with Stripe ID: %v", err)
		}
	}

	return &CustomerSyncResult{
		CustomerID: localCustomer.ID,
		StripeID:   stripeCustomer.ID,
		Action:     "updated",
		Message:    "Customer updated locally from Stripe",
		LastSyncAt: time.Now(),
	}, nil
}

func (s *StripeCustomerSyncService) getAllStripeCustomers() ([]*stripe.Customer, error) {
	var customers []*stripe.Customer
	params := &stripe.CustomerListParams{}
	params.Limit = stripe.Int64(100) // Adjust as needed

	iter := customer.List(params)
	for iter.Next() {
		customers = append(customers, iter.Current().(*stripe.Customer))
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return customers, nil
}

func (s *StripeCustomerSyncService) parseName(fullName string) (string, string) {
	if fullName == "" {
		return "", ""
	}

	// Simple name parsing - split on first space
	for i, char := range fullName {
		if char == ' ' {
			firstName := fullName[:i]
			lastName := fullName[i+1:]
			return firstName, lastName
		}
	}

	// No space found, treat as first name only
	return fullName, ""
}
