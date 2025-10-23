package stripe

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"bome-backend/infrastructure/database"
)

// StripeDatabaseService handles reading cached Stripe data from our database
type StripeDatabaseService struct {
	db *database.DB
}

// Customer represents a cached Stripe customer
type Customer struct {
	StripeID      string                 `json:"stripe_id"`
	Name          string                 `json:"name"`
	Email         string                 `json:"email"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Subscriptions []Subscription         `json:"subscriptions,omitempty"`
}

// Subscription represents a cached Stripe subscription
type Subscription struct {
	StripeID           string    `json:"stripe_id"`
	CustomerID         string    `json:"customer_id"` // Stripe customer ID
	Status             string    `json:"status"`
	CurrentPeriodStart time.Time `json:"current_period_start"`
	CurrentPeriodEnd   time.Time `json:"current_period_end"`
	CreatedAt          time.Time `json:"created_at"`
	PriceID            string    `json:"price_id,omitempty"`
	StripePriceID      string    `json:"stripe_price_id,omitempty"`
	UnitAmount         int64     `json:"unit_amount,omitempty"`
	Currency           string    `json:"currency,omitempty"`
	ProductID          string    `json:"stripe_product_id,omitempty"`
	ProductName        string    `json:"product_name"`
}

// DatabaseStats represents statistics about cached Stripe data
type DatabaseStats struct {
	TotalCustomers      int       `json:"total_customers"`
	TotalSubscriptions  int       `json:"total_subscriptions"`
	ActiveSubscriptions int       `json:"active_subscriptions"`
	TotalProducts       int       `json:"total_products"`
	TotalPrices         int       `json:"total_prices"`
	LastSyncAt          time.Time `json:"last_sync_at,omitempty"`
}

// NewStripeDatabaseService creates a new Stripe database service
func NewStripeDatabaseService(db *database.DB) *StripeDatabaseService {
	return &StripeDatabaseService{
		db: db,
	}
}

// GetCustomers retrieves customers from stripe_customers table
func (s *StripeDatabaseService) GetCustomers(limit, offset int, includeSubscriptions bool) ([]Customer, int, error) {
	// Get total count
	var totalCount int
	err := s.db.DB.QueryRow("SELECT COUNT(*) FROM stripe_customers").Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get customer count: %w", err)
	}

	// Query customers
	query := `
		SELECT 
			stripe_id, name, email, created_at, updated_at, metadata
		FROM stripe_customers 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.DB.Query(query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query customers: %w", err)
	}
	defer rows.Close()

	var customers []Customer
	for rows.Next() {
		var customer Customer
		var stripeID, name, email, metadataJSON sql.NullString
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(&stripeID, &name, &email, &createdAt, &updatedAt, &metadataJSON)
		if err != nil {
			log.Printf("Error scanning customer row: %v", err)
			continue
		}

		customer.StripeID = stripeID.String
		customer.Name = name.String
		customer.Email = email.String
		customer.CreatedAt = createdAt.Time
		customer.UpdatedAt = updatedAt.Time

		// Parse metadata JSON
		if metadataJSON.Valid && metadataJSON.String != "" {
			var metadata map[string]interface{}
			if err := json.Unmarshal([]byte(metadataJSON.String), &metadata); err == nil {
				customer.Metadata = metadata
			}
		}

		// Get subscriptions if requested
		if includeSubscriptions {
			log.Printf("🔍 DEBUG: Getting subscriptions for customer %s", stripeID.String)
			subs, err := s.getCustomerSubscriptions(stripeID.String)
			if err != nil {
				log.Printf("Error getting subscriptions for customer %s: %v", stripeID.String, err)
			} else {
				log.Printf("🔍 DEBUG: Found %d subscriptions for customer %s", len(subs), stripeID.String)
				customer.Subscriptions = subs
			}
		}

		customers = append(customers, customer)
	}

	log.Printf("📊 Retrieved %d customers from database (offset: %d, limit: %d, total: %d)",
		len(customers), offset, limit, totalCount)

	return customers, totalCount, nil
}

// GetSubscriptions retrieves subscriptions from stripe_subscriptions table
func (s *StripeDatabaseService) GetSubscriptions(limit, offset int, status string) ([]Subscription, int, error) {
	// Build query with optional status filter
	countQuery := "SELECT COUNT(*) FROM stripe_subscriptions"
	baseQuery := `
		SELECT 
			stripe_id, customer_id, status, current_period_start, current_period_end,
			created_at, stripe_price_id, unit_amount, currency, stripe_product_id, product_name
		FROM stripe_subscriptions
	`

	var args []interface{}
	argIndex := 1

	if status != "" {
		countQuery += " WHERE status = $1"
		baseQuery += " WHERE status = $1"
		args = append(args, status)
		argIndex++
	}

	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	// Get total count
	var totalCount int
	var countArgs []interface{}
	if status != "" {
		countArgs = []interface{}{status}
	}
	err := s.db.DB.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get subscription count: %w", err)
	}

	// Query subscriptions
	rows, err := s.db.DB.Query(baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []Subscription
	for rows.Next() {
		var sub Subscription
		var customerID sql.NullInt64
		var stripePriceID, currency, stripeProductID, productName sql.NullString
		var unitAmount sql.NullInt64

		err := rows.Scan(
			&sub.StripeID, &customerID, &sub.Status,
			&sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &sub.CreatedAt,
			&stripePriceID, &unitAmount, &currency, &stripeProductID, &productName,
		)
		if err != nil {
			log.Printf("Error scanning subscription row: %v", err)
			continue
		}

		// Get customer Stripe ID from customer_id (which is a foreign key to stripe_customers.id)
		if customerID.Valid {
			sub.CustomerID = s.getCustomerStripeID(int(customerID.Int64))
		}

		if stripePriceID.Valid {
			sub.StripePriceID = stripePriceID.String
		}
		if unitAmount.Valid {
			sub.UnitAmount = unitAmount.Int64
		}
		if currency.Valid {
			sub.Currency = currency.String
		}
		if stripeProductID.Valid {
			sub.ProductID = stripeProductID.String
		}
		if productName.Valid {
			sub.ProductName = productName.String
		}

		subscriptions = append(subscriptions, sub)
	}

	log.Printf("📊 Retrieved %d subscriptions from database (offset: %d, limit: %d, total: %d, status: %s)",
		len(subscriptions), offset, limit, totalCount, status)

	return subscriptions, totalCount, nil
}

// GetCustomerByStripeID retrieves a single customer by Stripe ID
func (s *StripeDatabaseService) GetCustomerByStripeID(stripeID string) (*Customer, error) {
	query := `
		SELECT 
			stripe_id, name, email, created_at, updated_at, metadata
		FROM stripe_customers 
		WHERE stripe_id = $1
	`

	var customer Customer
	var name, email, metadataJSON sql.NullString
	var createdAt, updatedAt sql.NullTime

	err := s.db.DB.QueryRow(query, stripeID).Scan(
		&customer.StripeID, &name, &email, &createdAt, &updatedAt, &metadataJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	customer.Name = name.String
	customer.Email = email.String
	customer.CreatedAt = createdAt.Time
	customer.UpdatedAt = updatedAt.Time

	// Parse metadata JSON
	if metadataJSON.Valid && metadataJSON.String != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataJSON.String), &metadata); err == nil {
			customer.Metadata = metadata
		}
	}

	// Get subscriptions
	subs, err := s.getCustomerSubscriptions(stripeID)
	if err != nil {
		log.Printf("Error getting subscriptions for customer %s: %v", stripeID, err)
	} else {
		customer.Subscriptions = subs
	}

	return &customer, nil
}

// GetStats returns statistics about cached Stripe data
func (s *StripeDatabaseService) GetStats() (*DatabaseStats, error) {
	stats := &DatabaseStats{}

	// Get customer count
	err := s.db.DB.QueryRow("SELECT COUNT(*) FROM stripe_customers").Scan(&stats.TotalCustomers)
	if err != nil {
		log.Printf("Error getting customer count: %v", err)
	}

	// Get subscription counts
	err = s.db.DB.QueryRow("SELECT COUNT(*) FROM stripe_subscriptions").Scan(&stats.TotalSubscriptions)
	if err != nil {
		log.Printf("Error getting subscription count: %v", err)
	}

	err = s.db.DB.QueryRow("SELECT COUNT(*) FROM stripe_subscriptions WHERE status = 'active'").Scan(&stats.ActiveSubscriptions)
	if err != nil {
		log.Printf("Error getting active subscription count: %v", err)
	}

	// Get product count
	err = s.db.DB.QueryRow("SELECT COUNT(*) FROM stripe_products").Scan(&stats.TotalProducts)
	if err != nil {
		log.Printf("Error getting product count: %v", err)
	}

	// Get price count
	err = s.db.DB.QueryRow("SELECT COUNT(*) FROM stripe_prices").Scan(&stats.TotalPrices)
	if err != nil {
		log.Printf("Error getting price count: %v", err)
	}

	// Get last sync time from stripe_entities
	var lastSync sql.NullTime
	err = s.db.DB.QueryRow("SELECT MAX(last_synced_at) FROM stripe_entities").Scan(&lastSync)
	if err != nil {
		log.Printf("Error getting last sync time: %v", err)
	}
	if lastSync.Valid {
		stats.LastSyncAt = lastSync.Time
	}

	return stats, nil
}

// getCustomerSubscriptions retrieves subscriptions for a customer
func (s *StripeDatabaseService) getCustomerSubscriptions(customerStripeID string) ([]Subscription, error) {
	query := `
		SELECT 
			ss.stripe_id, ss.status, ss.current_period_start, ss.current_period_end,
			ss.created_at, ss.stripe_price_id, ss.unit_amount, ss.currency,
			ss.stripe_product_id, ss.product_name
		FROM stripe_subscriptions ss
		JOIN stripe_customers sc ON ss.customer_id = sc.id
		WHERE sc.stripe_id = $1
		ORDER BY ss.created_at DESC
	`

	rows, err := s.db.DB.Query(query, customerStripeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []Subscription
	for rows.Next() {
		var sub Subscription
		var stripePriceID, currency, stripeProductID, productName sql.NullString
		var unitAmount sql.NullInt64

		err := rows.Scan(
			&sub.StripeID, &sub.Status,
			&sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &sub.CreatedAt,
			&stripePriceID, &unitAmount, &currency, &stripeProductID, &productName,
		)
		if err != nil {
			log.Printf("Error scanning subscription: %v", err)
			continue
		}

		// DEBUG: Log what we scanned
		log.Printf("🔍 DEBUG: Scanned subscription for %s: stripePriceID=%s, unitAmount=%d, currency=%s, stripeProductID=%s, productName=%s",
			customerStripeID, stripePriceID.String, unitAmount.Int64, currency.String, stripeProductID.String, productName.String)

		sub.CustomerID = customerStripeID

		if stripePriceID.Valid {
			sub.StripePriceID = stripePriceID.String
		}
		if unitAmount.Valid {
			sub.UnitAmount = unitAmount.Int64
		}
		if currency.Valid {
			sub.Currency = currency.String
		}
		if stripeProductID.Valid {
			sub.ProductID = stripeProductID.String
		}
		if productName.Valid {
			sub.ProductName = productName.String
		}

		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

// getCustomerStripeID converts customer database ID to Stripe ID
func (s *StripeDatabaseService) getCustomerStripeID(customerID int) string {
	var stripeID string
	err := s.db.DB.QueryRow("SELECT stripe_id FROM stripe_customers WHERE id = $1", customerID).Scan(&stripeID)
	if err != nil {
		log.Printf("Error getting customer Stripe ID: %v", err)
		return ""
	}
	return stripeID
}
