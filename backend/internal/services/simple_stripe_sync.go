package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/price"
	"github.com/stripe/stripe-go/v74/product"
	"github.com/stripe/stripe-go/v74/subscription"

	"bome-backend/internal/database"
)

// SimpleStripeSyncService handles straightforward Stripe data synchronization
// No ghost detection, no complex validation - just clean, simple syncing
type SimpleStripeSyncService struct {
	db            *database.DB
	stripeService *StripeService
}

// NewSimpleStripeSyncService creates a new simple sync service
func NewSimpleStripeSyncService(db *database.DB, stripeService *StripeService) *SimpleStripeSyncService {
	return &SimpleStripeSyncService{
		db:            db,
		stripeService: stripeService,
	}
}

// SyncAll performs a complete sync of all Stripe data in the correct order
func (s *SimpleStripeSyncService) SyncAll(ctx context.Context) error {
	log.Println("🚀 Starting simple Stripe sync - all data")

	// Step 1: Sync products first (no dependencies)
	if err := s.syncProducts(ctx); err != nil {
		return fmt.Errorf("failed to sync products: %w", err)
	}

	// Step 2: Sync prices (depends on products)
	if err := s.syncPrices(ctx); err != nil {
		return fmt.Errorf("failed to sync prices: %w", err)
	}

	// Step 3: Sync customers (no dependencies)
	if err := s.syncCustomers(ctx); err != nil {
		return fmt.Errorf("failed to sync customers: %w", err)
	}

	// Step 4: Sync subscriptions (depends on customers and prices)
	if err := s.syncSubscriptions(ctx); err != nil {
		return fmt.Errorf("failed to sync subscriptions: %w", err)
	}

	log.Println("✅ Simple Stripe sync completed successfully")
	return nil
}

// syncProducts syncs all products from Stripe
func (s *SimpleStripeSyncService) syncProducts(ctx context.Context) error {
	log.Println("📦 Syncing products...")

	params := &stripe.ProductListParams{}
	params.Limit = stripe.Int64(100)

	iter := product.List(params)
	count := 0

	for iter.Next() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		prod := iter.Current().(*stripe.Product)

		// Simple upsert - no validation, no ghost detection
		query := `
			INSERT INTO stripe_products (stripe_id, name, description, active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (stripe_id) 
			DO UPDATE SET 
				name = EXCLUDED.name,
				description = EXCLUDED.description,
				active = EXCLUDED.active,
				updated_at = EXCLUDED.updated_at
		`

		_, err := s.db.Exec(query,
			prod.ID,
			prod.Name,
			prod.Description,
			prod.Active,
			time.Unix(prod.Created, 0),
			time.Unix(prod.Updated, 0),
		)

		if err != nil {
			log.Printf("⚠️ Failed to upsert product %s: %v", prod.ID, err)
			continue
		}

		count++
		if count%50 == 0 {
			log.Printf("📦 Processed %d products", count)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("stripe API error: %w", err)
	}

	log.Printf("✅ Synced %d products", count)
	return nil
}

// syncPrices syncs all prices from Stripe
func (s *SimpleStripeSyncService) syncPrices(ctx context.Context) error {
	log.Println("💰 Syncing prices...")

	params := &stripe.PriceListParams{}
	params.Limit = stripe.Int64(100)

	iter := price.List(params)
	count := 0

	for iter.Next() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		pr := iter.Current().(*stripe.Price)

		// Get local product ID - but don't fail if product doesn't exist
		var productID sql.NullInt64
		err := s.db.QueryRow("SELECT id FROM stripe_products WHERE stripe_id = $1", pr.Product.ID).Scan(&productID)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("⚠️ Failed to find product %s for price %s: %v", pr.Product.ID, pr.ID, err)
			continue
		}

		// Skip this price if the product doesn't exist in our database
		if !productID.Valid {
			log.Printf("⚠️ Skipping price %s - product %s not found in database", pr.ID, pr.Product.ID)
			continue
		}

		var recurringInterval string
		if pr.Recurring != nil {
			recurringInterval = string(pr.Recurring.Interval)
		}

		// Simple upsert - using correct product_id column
		query := `
			INSERT INTO stripe_prices (stripe_id, product_id, currency, unit_amount, recurring_interval, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (stripe_id) 
			DO UPDATE SET 
				product_id = EXCLUDED.product_id,
				currency = EXCLUDED.currency,
				unit_amount = EXCLUDED.unit_amount,
				recurring_interval = EXCLUDED.recurring_interval
		`

		_, err = s.db.Exec(query,
			pr.ID,
			productID,
			string(pr.Currency),
			pr.UnitAmount,
			recurringInterval,
			time.Unix(pr.Created, 0),
		)

		if err != nil {
			log.Printf("⚠️ Failed to upsert price %s: %v", pr.ID, err)
			continue
		}

		count++
		if count%50 == 0 {
			log.Printf("💰 Processed %d prices", count)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("stripe API error: %w", err)
	}

	log.Printf("✅ Synced %d prices", count)
	return nil
}

// syncCustomers syncs all customers from Stripe
func (s *SimpleStripeSyncService) syncCustomers(ctx context.Context) error {
	log.Println("👥 Syncing customers...")

	params := &stripe.CustomerListParams{}
	params.Limit = stripe.Int64(100)

	iter := customer.List(params)
	count := 0

	for iter.Next() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		cust := iter.Current().(*stripe.Customer)

		// Simple upsert
		query := `
			INSERT INTO stripe_customers (stripe_id, email, name, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (stripe_id) 
			DO UPDATE SET 
				email = EXCLUDED.email,
				name = EXCLUDED.name,
				updated_at = EXCLUDED.updated_at
		`

		_, err := s.db.Exec(query,
			cust.ID,
			cust.Email,
			cust.Name,
			time.Unix(cust.Created, 0),
			time.Now(),
		)

		if err != nil {
			log.Printf("⚠️ Failed to upsert customer %s: %v", cust.ID, err)
			continue
		}

		count++
		if count%50 == 0 {
			log.Printf("👥 Processed %d customers", count)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("stripe API error: %w", err)
	}

	log.Printf("✅ Synced %d customers", count)
	return nil
}

// syncSubscriptions syncs all subscriptions from Stripe
func (s *SimpleStripeSyncService) syncSubscriptions(ctx context.Context) error {
	log.Println("📋 Syncing subscriptions...")

	// Sync active and canceled subscriptions
	statuses := []string{"active", "canceled", "trialing"}
	totalCount := 0

	for _, status := range statuses {
		log.Printf("📋 Syncing %s subscriptions...", status)

		params := &stripe.SubscriptionListParams{
			ListParams: stripe.ListParams{
				Limit: stripe.Int64(100),
			},
			Status: stripe.String(status),
		}

		iter := subscription.List(params)
		statusCount := 0

		for iter.Next() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			sub := iter.Current().(*stripe.Subscription)

			// Get local customer ID - but don't fail if customer doesn't exist
			var customerID sql.NullInt64
			err := s.db.QueryRow("SELECT id FROM stripe_customers WHERE stripe_id = $1", sub.Customer.ID).Scan(&customerID)
			if err != nil && err != sql.ErrNoRows {
				log.Printf("⚠️ Failed to find customer %s for subscription %s: %v", sub.Customer.ID, sub.ID, err)
				continue
			}

			// Skip this subscription if the customer doesn't exist in our database
			if !customerID.Valid {
				log.Printf("⚠️ Skipping subscription %s - customer %s not found in database", sub.ID, sub.Customer.ID)
				continue
			}

			// Extract price and product info from first item
			var priceID sql.NullInt64
			var stripePriceID, stripeProductID, productName sql.NullString
			var unitAmount sql.NullInt64
			var currency sql.NullString

			if len(sub.Items.Data) > 0 {
				firstItem := sub.Items.Data[0]
				if firstItem.Price != nil {
					stripePriceID = sql.NullString{String: firstItem.Price.ID, Valid: true}
					unitAmount = sql.NullInt64{Int64: firstItem.Price.UnitAmount, Valid: true}
					currency = sql.NullString{String: string(firstItem.Price.Currency), Valid: true}

					// Get local price ID - but don't require it to exist
					s.db.QueryRow("SELECT id FROM stripe_prices WHERE stripe_id = $1", firstItem.Price.ID).Scan(&priceID)

					// Get product info
					if firstItem.Price.Product != nil {
						stripeProductID = sql.NullString{String: firstItem.Price.Product.ID, Valid: true}
						if firstItem.Price.Product.Name != "" {
							productName = sql.NullString{String: firstItem.Price.Product.Name, Valid: true}
						}
					}
				}
			}

			// Simple upsert
			query := `
				INSERT INTO stripe_subscriptions (
					stripe_id, customer_id, price_id, status, 
					current_period_start, current_period_end, created_at,
					stripe_price_id, unit_amount, currency, stripe_product_id, product_name
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
				ON CONFLICT (stripe_id) 
				DO UPDATE SET 
					customer_id = EXCLUDED.customer_id,
					price_id = EXCLUDED.price_id,
					status = EXCLUDED.status,
					current_period_start = EXCLUDED.current_period_start,
					current_period_end = EXCLUDED.current_period_end,
					stripe_price_id = EXCLUDED.stripe_price_id,
					unit_amount = EXCLUDED.unit_amount,
					currency = EXCLUDED.currency,
					stripe_product_id = EXCLUDED.stripe_product_id,
					product_name = EXCLUDED.product_name
			`

			_, err = s.db.Exec(query,
				sub.ID,
				customerID,
				priceID,
				string(sub.Status),
				time.Unix(sub.CurrentPeriodStart, 0),
				time.Unix(sub.CurrentPeriodEnd, 0),
				time.Unix(sub.Created, 0),
				stripePriceID,
				unitAmount,
				currency,
				stripeProductID,
				productName,
			)

			if err != nil {
				log.Printf("⚠️ Failed to upsert subscription %s: %v", sub.ID, err)
				continue
			}

			statusCount++
			totalCount++

			if statusCount%25 == 0 {
				log.Printf("📋 Processed %d %s subscriptions", statusCount, status)
			}
		}

		if err := iter.Err(); err != nil {
			return fmt.Errorf("stripe API error for %s subscriptions: %w", status, err)
		}

		log.Printf("✅ Synced %d %s subscriptions", statusCount, status)
	}

	log.Printf("✅ Synced %d total subscriptions", totalCount)
	return nil
}

// LinkCustomersToUsers links Stripe customers to local users by email
func (s *SimpleStripeSyncService) LinkCustomersToUsers(ctx context.Context) error {
	log.Println("🔗 Linking Stripe customers to local users...")

	// Simple approach: match by email
	query := `
		UPDATE users 
		SET stripe_customer_id = sc.stripe_id
		FROM stripe_customers sc
		WHERE users.email = sc.email 
		AND users.stripe_customer_id IS NULL
		AND sc.email IS NOT NULL
		AND sc.email != ''
	`

	result, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to link customers to users: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("✅ Linked %d customers to users", rowsAffected)

	return nil
}
