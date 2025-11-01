package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"bome-backend/internal/database"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/price"
	"github.com/stripe/stripe-go/v74/product"
	"github.com/stripe/stripe-go/v74/subscription"
)

// StripeSyncV2Service handles syncing Stripe data to v2 tables
type StripeSyncV2Service struct {
	db           *database.DB
	ghostTracker *GhostTrackingService
}

// SyncProgress tracks sync progress for reporting
type SyncProgress struct {
	ProductsTotal       int
	ProductsSynced      int
	PricesTotal         int
	PricesSynced        int
	CustomersTotal      int
	CustomersSynced     int
	SubscriptionsTotal  int
	SubscriptionsSynced int
	Errors              []string
	StartedAt           time.Time
	CompletedAt         *time.Time
}

// NewStripeSyncV2Service creates a new Stripe sync service for v2 tables
func NewStripeSyncV2Service(db *database.DB) *StripeSyncV2Service {
	return &StripeSyncV2Service{
		db:           db,
		ghostTracker: NewGhostTrackingService(db),
	}
}

// ================================================================
// FULL SYNC - Syncs all Stripe data to v2 tables
// ================================================================

// SyncAll syncs all Stripe data (products, prices, customers, subscriptions)
func (s *StripeSyncV2Service) SyncAll(ctx context.Context) (*SyncProgress, error) {
	progress := &SyncProgress{
		StartedAt: time.Now(),
		Errors:    []string{},
	}

	log.Printf("🚀 [Stripe Sync v2] Starting full sync...")

	// Step 1: Sync products first (no dependencies)
	log.Printf("📦 [Stripe Sync v2] Step 1/4: Syncing products...")
	if err := s.SyncProducts(ctx, progress); err != nil {
		progress.Errors = append(progress.Errors, fmt.Sprintf("Products: %v", err))
		log.Printf("❌ [Stripe Sync v2] Product sync failed: %v", err)
	}

	// Step 2: Sync prices (depends on products)
	log.Printf("💰 [Stripe Sync v2] Step 2/4: Syncing prices...")
	if err := s.SyncPrices(ctx, progress); err != nil {
		progress.Errors = append(progress.Errors, fmt.Sprintf("Prices: %v", err))
		log.Printf("❌ [Stripe Sync v2] Price sync failed: %v", err)
	}

	// Step 3: Sync customers (no dependencies)
	log.Printf("👥 [Stripe Sync v2] Step 3/4: Syncing customers...")
	if err := s.SyncCustomers(ctx, progress); err != nil {
		progress.Errors = append(progress.Errors, fmt.Sprintf("Customers: %v", err))
		log.Printf("❌ [Stripe Sync v2] Customer sync failed: %v", err)
	}

	// Step 4: Sync subscriptions (depends on customers and prices)
	log.Printf("📋 [Stripe Sync v2] Step 4/4: Syncing subscriptions...")
	if err := s.SyncSubscriptions(ctx, progress); err != nil {
		progress.Errors = append(progress.Errors, fmt.Sprintf("Subscriptions: %v", err))
		log.Printf("❌ [Stripe Sync v2] Subscription sync failed: %v", err)
	}

	completedAt := time.Now()
	progress.CompletedAt = &completedAt

	log.Printf("✅ [Stripe Sync v2] Full sync complete!")
	log.Printf("📊 [Stripe Sync v2] Summary:")
	log.Printf("   Products: %d/%d synced", progress.ProductsSynced, progress.ProductsTotal)
	log.Printf("   Prices: %d/%d synced", progress.PricesSynced, progress.PricesTotal)
	log.Printf("   Customers: %d/%d synced", progress.CustomersSynced, progress.CustomersTotal)
	log.Printf("   Subscriptions: %d/%d synced", progress.SubscriptionsSynced, progress.SubscriptionsTotal)
	if len(progress.Errors) > 0 {
		log.Printf("   ⚠️ Errors: %d", len(progress.Errors))
		for _, err := range progress.Errors {
			log.Printf("      - %s", err)
		}
	}

	return progress, nil
}

// ================================================================
// PRODUCTS SYNC
// ================================================================

// SyncProducts syncs all products from Stripe to stripe_products_v2
func (s *StripeSyncV2Service) SyncProducts(ctx context.Context, progress *SyncProgress) error {
	params := &stripe.ProductListParams{}
	params.Limit = stripe.Int64(100)

	iter := product.List(params)
	count := 0

	for iter.Next() {
		prod := iter.Product()
		count++

		if err := s.upsertProduct(prod); err != nil {
			progress.Errors = append(progress.Errors, fmt.Sprintf("Product %s: %v", prod.ID, err))
			log.Printf("⚠️ [Stripe Sync v2] Failed to sync product %s: %v", prod.ID, err)
			continue
		}

		progress.ProductsSynced++
		if count%10 == 0 {
			log.Printf("📦 [Stripe Sync v2] Synced %d products...", count)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to list products: %w", err)
	}

	progress.ProductsTotal = count
	log.Printf("✅ [Stripe Sync v2] Synced %d products", count)
	return nil
}

func (s *StripeSyncV2Service) upsertProduct(prod *stripe.Product) error {
	// 🛡️ GHOST DETECTION: Block known ghost product IDs
	ghostProducts := map[string]bool{
		"prod_HEmcX1PE8TO2CO": true,
		"prod_FvNAeI348dup9w": true,
		"prod_HF5YzcBH5Rwr0d": true,
		"prod_GVV5efccnh13h9": true,
		"prod_FvNAJgnw48hwpZ": true,
	}

	if ghostProducts[prod.ID] {
		log.Printf("👻 [V2] GHOST BLOCKED: Product %s - LOGGING TO GHOST TABLE", prod.ID)

		if s.ghostTracker != nil {
			metadata := map[string]interface{}{
				"name":        prod.Name,
				"description": prod.Description,
				"active":      prod.Active,
				"created":     prod.Created,
			}
			s.ghostTracker.LogGhostProduct(context.Background(), prod.ID, "Product deleted from Stripe or known ghost", metadata)
		}

		return nil // Skip sync but logged for admin visibility
	}

	query := `
		INSERT INTO stripe_products_v2 (
			stripe_id, name, description, active, metadata,
			stripe_created_at, stripe_updated_at,
			video_approved, is_legacy,
			first_synced_at, last_synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		ON CONFLICT (stripe_id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			active = EXCLUDED.active,
			metadata = EXCLUDED.metadata,
			stripe_updated_at = EXCLUDED.stripe_updated_at,
			last_synced_at = NOW()
	`

	// Convert metadata to JSONB
	metadataJSON, err := json.Marshal(prod.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Check if product should grant video access (based on metadata or name)
	videoApproved := false
	if val, ok := prod.Metadata["video_approved"]; ok && val == "true" {
		videoApproved = true
	}

	// Check if this is a legacy product
	isLegacy := false
	if val, ok := prod.Metadata["legacy"]; ok && val == "true" {
		isLegacy = true
	}

	_, err = s.db.Exec(query,
		prod.ID,                    // stripe_id
		prod.Name,                  // name
		prod.Description,           // description
		prod.Active,                // active
		string(metadataJSON),       // metadata
		time.Unix(prod.Created, 0), // stripe_created_at
		time.Unix(prod.Updated, 0), // stripe_updated_at
		videoApproved,              // video_approved
		isLegacy,                   // is_legacy
	)

	if err != nil {
		return err
	}

	// Note: We don't auto-remove ghost products from the table anymore
	// They stay in the ghost table for admin visibility until manually removed
	// This allows admins to track historical issues and decide when to clean them up

	return nil
}

// ================================================================
// PRICES SYNC
// ================================================================

// SyncPrices syncs all prices from Stripe to stripe_prices_v2
func (s *StripeSyncV2Service) SyncPrices(ctx context.Context, progress *SyncProgress) error {
	params := &stripe.PriceListParams{}
	params.Limit = stripe.Int64(100)

	iter := price.List(params)
	count := 0

	for iter.Next() {
		pr := iter.Price()
		count++

		if err := s.upsertPrice(pr); err != nil {
			progress.Errors = append(progress.Errors, fmt.Sprintf("Price %s: %v", pr.ID, err))
			log.Printf("⚠️ [Stripe Sync v2] Failed to sync price %s: %v", pr.ID, err)
			continue
		}

		progress.PricesSynced++
		if count%10 == 0 {
			log.Printf("💰 [Stripe Sync v2] Synced %d prices...", count)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to list prices: %w", err)
	}

	progress.PricesTotal = count
	log.Printf("✅ [Stripe Sync v2] Synced %d prices", count)
	return nil
}

func (s *StripeSyncV2Service) upsertPrice(pr *stripe.Price) error {
	// 🛡️ GHOST DETECTION: Block prices for known ghost product IDs
	ghostProducts := map[string]bool{
		"prod_HEmcX1PE8TO2CO": true,
		"prod_FvNAeI348dup9w": true,
		"prod_HF5YzcBH5Rwr0d": true,
		"prod_GVV5efccnh13h9": true,
		"prod_FvNAJgnw48hwpZ": true,
	}

	if ghostProducts[pr.Product.ID] {
		log.Printf("👻 [V2] GHOST BLOCKED: Price %s references ghost product %s - LOGGING TO GHOST TABLE", pr.ID, pr.Product.ID)

		if s.ghostTracker != nil {
			metadata := map[string]interface{}{
				"currency":    pr.Currency,
				"unit_amount": pr.UnitAmount,
				"active":      pr.Active,
				"created":     pr.Created,
			}
			if pr.Recurring != nil {
				metadata["recurring_interval"] = pr.Recurring.Interval
			}
			s.ghostTracker.LogGhostPrice(context.Background(), pr.ID, pr.Product.ID, metadata)
		}

		return nil // Skip sync but logged for admin visibility
	}

	// First, get the product_id from stripe_products_v2
	var productID int
	err := s.db.QueryRow(`
		SELECT id FROM stripe_products_v2 WHERE stripe_id = $1
	`, pr.Product.ID).Scan(&productID)

	if err == sql.ErrNoRows {
		// Product doesn't exist yet - this shouldn't happen if we synced products first
		// But we can sync the product now
		log.Printf("⚠️ [Stripe Sync v2] Product %s not found for price %s, fetching...", pr.Product.ID, pr.ID)
		prod, err := product.Get(pr.Product.ID, nil)
		if err != nil {
			return fmt.Errorf("failed to fetch product %s: %w", pr.Product.ID, err)
		}
		if err := s.upsertProduct(prod); err != nil {
			return fmt.Errorf("failed to upsert product %s: %w", pr.Product.ID, err)
		}
		// Try again
		err = s.db.QueryRow(`
			SELECT id FROM stripe_products_v2 WHERE stripe_id = $1
		`, pr.Product.ID).Scan(&productID)
		if err != nil {
			return fmt.Errorf("failed to get product id after upsert: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to query product: %w", err)
	}

	query := `
		INSERT INTO stripe_prices_v2 (
			stripe_id, product_id, unit_amount, currency, active,
			recurring_interval, recurring_interval_count, metadata,
			stripe_created_at, first_synced_at, last_synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		ON CONFLICT (stripe_id) DO UPDATE SET
			product_id = EXCLUDED.product_id,
			unit_amount = EXCLUDED.unit_amount,
			currency = EXCLUDED.currency,
			active = EXCLUDED.active,
			recurring_interval = EXCLUDED.recurring_interval,
			recurring_interval_count = EXCLUDED.recurring_interval_count,
			metadata = EXCLUDED.metadata,
			last_synced_at = NOW()
	`

	// Convert metadata to JSONB
	metadataJSON, err := json.Marshal(pr.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Get recurring interval details
	var recurringInterval *string
	var recurringIntervalCount *int
	if pr.Recurring != nil {
		interval := string(pr.Recurring.Interval)
		recurringInterval = &interval
		count := int(pr.Recurring.IntervalCount)
		recurringIntervalCount = &count
	}

	_, err = s.db.Exec(query,
		pr.ID,                    // stripe_id
		productID,                // product_id (FK to stripe_products_v2.id)
		pr.UnitAmount,            // unit_amount (cents)
		pr.Currency,              // currency
		pr.Active,                // active
		recurringInterval,        // recurring_interval
		recurringIntervalCount,   // recurring_interval_count
		string(metadataJSON),     // metadata
		time.Unix(pr.Created, 0), // stripe_created_at
	)

	if err != nil {
		return err
	}

	// Note: We don't auto-remove ghost prices from the table anymore
	// They stay visible for admin review until manually removed

	return nil
}

// ================================================================
// CUSTOMERS SYNC
// ================================================================

// SyncCustomers syncs all customers from Stripe to stripe_customers_v2
func (s *StripeSyncV2Service) SyncCustomers(ctx context.Context, progress *SyncProgress) error {
	params := &stripe.CustomerListParams{}
	params.Limit = stripe.Int64(100)

	iter := customer.List(params)
	count := 0

	for iter.Next() {
		cust := iter.Customer()
		count++

		if err := s.upsertCustomer(cust); err != nil {
			progress.Errors = append(progress.Errors, fmt.Sprintf("Customer %s: %v", cust.ID, err))
			log.Printf("⚠️ [Stripe Sync v2] Failed to sync customer %s: %v", cust.ID, err)
			continue
		}

		progress.CustomersSynced++
		if count%100 == 0 {
			log.Printf("👥 [Stripe Sync v2] Synced %d customers...", count)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to list customers: %w", err)
	}

	progress.CustomersTotal = count
	log.Printf("✅ [Stripe Sync v2] Synced %d customers", count)
	return nil
}

func (s *StripeSyncV2Service) upsertCustomer(cust *stripe.Customer) error {
	query := `
		INSERT INTO stripe_customers_v2 (
			stripe_id, email, name, phone, address, metadata,
			balance, currency, delinquent,
			stripe_created_at, first_synced_at, last_synced_at, sync_source
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW(), 'manual_sync')
		ON CONFLICT (stripe_id) DO UPDATE SET
			email = EXCLUDED.email,
			name = EXCLUDED.name,
			phone = EXCLUDED.phone,
			address = EXCLUDED.address,
			metadata = EXCLUDED.metadata,
			balance = EXCLUDED.balance,
			currency = EXCLUDED.currency,
			delinquent = EXCLUDED.delinquent,
			last_synced_at = NOW()
	`

	// Convert address to JSONB
	var addressJSON *string
	if cust.Address != nil {
		addrBytes, err := json.Marshal(cust.Address)
		if err == nil {
			addrStr := string(addrBytes)
			addressJSON = &addrStr
		}
	}

	// Convert metadata to JSONB
	metadataJSON, err := json.Marshal(cust.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = s.db.Exec(query,
		cust.ID,                    // stripe_id
		cust.Email,                 // email
		cust.Name,                  // name
		cust.Phone,                 // phone
		addressJSON,                // address
		string(metadataJSON),       // metadata
		cust.Balance,               // balance (cents)
		cust.Currency,              // currency
		cust.Delinquent,            // delinquent
		time.Unix(cust.Created, 0), // stripe_created_at
	)

	return err
}

// ================================================================
// SUBSCRIPTIONS SYNC
// ================================================================

// SyncSubscriptions syncs all subscriptions from Stripe to stripe_subscriptions_v2
func (s *StripeSyncV2Service) SyncSubscriptions(ctx context.Context, progress *SyncProgress) error {
	params := &stripe.SubscriptionListParams{}
	params.Limit = stripe.Int64(100)

	iter := subscription.List(params)
	count := 0

	for iter.Next() {
		sub := iter.Subscription()
		count++

		if err := s.upsertSubscription(sub); err != nil {
			progress.Errors = append(progress.Errors, fmt.Sprintf("Subscription %s: %v", sub.ID, err))
			log.Printf("⚠️ [Stripe Sync v2] Failed to sync subscription %s: %v", sub.ID, err)
			continue
		}

		progress.SubscriptionsSynced++
		if count%100 == 0 {
			log.Printf("📋 [Stripe Sync v2] Synced %d subscriptions...", count)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to list subscriptions: %w", err)
	}

	progress.SubscriptionsTotal = count
	log.Printf("✅ [Stripe Sync v2] Synced %d subscriptions", count)
	return nil
}

func (s *StripeSyncV2Service) upsertSubscription(sub *stripe.Subscription) error {
	// 🛡️ GHOST DETECTION: Block subscriptions referencing known ghost products
	ghostProducts := map[string]bool{
		"prod_HEmcX1PE8TO2CO": true,
		"prod_FvNAeI348dup9w": true,
		"prod_HF5YzcBH5Rwr0d": true,
		"prod_GVV5efccnh13h9": true,
		"prod_FvNAJgnw48hwpZ": true,
	}

	// Check if this subscription references a ghost product
	if len(sub.Items.Data) > 0 {
		firstItem := sub.Items.Data[0]
		if firstItem.Price != nil && firstItem.Price.Product != nil {
			productID := firstItem.Price.Product.ID
			if ghostProducts[productID] {
				log.Printf("👻 [V2] GHOST BLOCKED: Subscription %s references ghost product %s - LOGGING TO GHOST TABLE", sub.ID, productID)

				// Get customer email
				var customerEmail string
				if sub.Customer != nil {
					customerEmail = sub.Customer.Email
				}

				// Log to ghost tracking table
				if s.ghostTracker != nil {
					metadata := map[string]interface{}{
						"status":               sub.Status,
						"current_period_end":   sub.CurrentPeriodEnd,
						"current_period_start": sub.CurrentPeriodStart,
						"created":              sub.Created,
					}
					if firstItem.Price != nil {
						metadata["price_id"] = firstItem.Price.ID
						metadata["unit_amount"] = firstItem.Price.UnitAmount
						metadata["currency"] = firstItem.Price.Currency
					}
					s.ghostTracker.LogGhostSubscription(context.Background(), sub.ID, productID, sub.Customer.ID, customerEmail, metadata)
				}

				return nil // Skip sync but logged for admin visibility
			}
		}
	}

	// Get customer_id from stripe_customers_v2
	var customerID int
	err := s.db.QueryRow(`
		SELECT id FROM stripe_customers_v2 WHERE stripe_id = $1
	`, sub.Customer.ID).Scan(&customerID)

	if err == sql.ErrNoRows {
		// Customer doesn't exist yet - fetch and sync
		log.Printf("⚠️ [Stripe Sync v2] Customer %s not found for subscription %s, fetching...", sub.Customer.ID, sub.ID)
		cust, err := customer.Get(sub.Customer.ID, nil)
		if err != nil {
			return fmt.Errorf("failed to fetch customer %s: %w", sub.Customer.ID, err)
		}
		if err := s.upsertCustomer(cust); err != nil {
			return fmt.Errorf("failed to upsert customer %s: %w", sub.Customer.ID, err)
		}
		// Try again
		err = s.db.QueryRow(`
			SELECT id FROM stripe_customers_v2 WHERE stripe_id = $1
		`, sub.Customer.ID).Scan(&customerID)
		if err != nil {
			return fmt.Errorf("failed to get customer id after upsert: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to query customer: %w", err)
	}

	// Get price_id from stripe_prices_v2 (use first item's price)
	if len(sub.Items.Data) == 0 {
		return fmt.Errorf("subscription %s has no items", sub.ID)
	}

	var priceID int
	priceStripeID := sub.Items.Data[0].Price.ID
	err = s.db.QueryRow(`
		SELECT id FROM stripe_prices_v2 WHERE stripe_id = $1
	`, priceStripeID).Scan(&priceID)

	if err == sql.ErrNoRows {
		// Price doesn't exist yet - fetch and sync
		log.Printf("⚠️ [Stripe Sync v2] Price %s not found for subscription %s, fetching...", priceStripeID, sub.ID)
		pr, err := price.Get(priceStripeID, nil)
		if err != nil {
			// Price fetch failed - this is a ghost price!
			log.Printf("👻 [V2] GHOST DETECTED: Price %s cannot be fetched from Stripe (subscription %s) - LOGGING TO GHOST TABLE", priceStripeID, sub.ID)

			// Get customer email
			var customerEmail string
			if sub.Customer != nil {
				customerEmail = sub.Customer.Email
			}

			// Log subscription as ghost because it references a ghost price
			if s.ghostTracker != nil {
				metadata := map[string]interface{}{
					"status":               sub.Status,
					"current_period_end":   sub.CurrentPeriodEnd,
					"current_period_start": sub.CurrentPeriodStart,
					"created":              sub.Created,
					"ghost_price_id":       priceStripeID,
					"error":                err.Error(),
				}
				if len(sub.Items.Data) > 0 && sub.Items.Data[0].Price != nil {
					metadata["unit_amount"] = sub.Items.Data[0].Price.UnitAmount
					metadata["currency"] = sub.Items.Data[0].Price.Currency
					if sub.Items.Data[0].Price.Product != nil {
						metadata["product_id"] = sub.Items.Data[0].Price.Product.ID
					}
				}
				s.ghostTracker.LogGhostSubscription(context.Background(), sub.ID, priceStripeID, sub.Customer.ID, customerEmail, metadata)

				// Also log the ghost price itself
				priceMetadata := map[string]interface{}{
					"referenced_by_subscription": sub.ID,
					"error":                      err.Error(),
				}
				s.ghostTracker.LogGhostPrice(context.Background(), priceStripeID, "unknown_product", priceMetadata)
			}

			return nil // Skip this subscription but logged for admin visibility
		}
		if err := s.upsertPrice(pr); err != nil {
			return fmt.Errorf("failed to upsert price %s: %w", priceStripeID, err)
		}
		// Try again
		err = s.db.QueryRow(`
			SELECT id FROM stripe_prices_v2 WHERE stripe_id = $1
		`, priceStripeID).Scan(&priceID)
		if err != nil {
			return fmt.Errorf("failed to get price id after upsert: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to query price: %w", err)
	}

	query := `
		INSERT INTO stripe_subscriptions_v2 (
			stripe_id, customer_id, price_id, status,
			current_period_start, current_period_end,
			cancel_at_period_end, canceled_at,
			stripe_created_at, first_synced_at, last_synced_at, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW(), $10)
		ON CONFLICT (stripe_id) DO UPDATE SET
			customer_id = EXCLUDED.customer_id,
			price_id = EXCLUDED.price_id,
			status = EXCLUDED.status,
			current_period_start = EXCLUDED.current_period_start,
			current_period_end = EXCLUDED.current_period_end,
			cancel_at_period_end = EXCLUDED.cancel_at_period_end,
			canceled_at = EXCLUDED.canceled_at,
			last_synced_at = NOW(),
			metadata = EXCLUDED.metadata
	`

	// Convert metadata to JSONB
	metadataJSON, err := json.Marshal(sub.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Handle canceled_at
	var canceledAt *time.Time
	if sub.CanceledAt > 0 {
		t := time.Unix(sub.CanceledAt, 0)
		canceledAt = &t
	}

	_, err = s.db.Exec(query,
		sub.ID,                               // stripe_id
		customerID,                           // customer_id (FK)
		priceID,                              // price_id (FK)
		string(sub.Status),                   // status
		time.Unix(sub.CurrentPeriodStart, 0), // current_period_start
		time.Unix(sub.CurrentPeriodEnd, 0),   // current_period_end
		sub.CancelAtPeriodEnd,                // cancel_at_period_end
		canceledAt,                           // canceled_at
		time.Unix(sub.Created, 0),            // stripe_created_at
		string(metadataJSON),                 // metadata
	)

	if err != nil {
		return err
	}

	// Auto-remove from ghost table if it was previously a ghost
	if s.ghostTracker != nil {
		s.ghostTracker.CheckAndRemoveGhostSubscription(context.Background(), sub.ID)
	}

	return nil
}

// ================================================================
// SINGLE-ENTITY SYNC - For webhook use (Phase 5)
// ================================================================

// SyncSingleProduct syncs a single product by ID from Stripe to v2 tables
func (s *StripeSyncV2Service) SyncSingleProduct(ctx context.Context, productID string) error {
	log.Printf("📦 [Stripe Sync v2] Syncing single product: %s", productID)

	// Fetch the product from Stripe
	prod, err := product.Get(productID, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch product %s: %w", productID, err)
	}

	// Upsert to database
	if err := s.upsertProduct(prod); err != nil {
		return fmt.Errorf("failed to upsert product %s: %w", productID, err)
	}

	log.Printf("✅ [Stripe Sync v2] Product %s synced successfully", productID)
	return nil
}

// SyncSinglePrice syncs a single price by ID from Stripe to v2 tables
func (s *StripeSyncV2Service) SyncSinglePrice(ctx context.Context, priceID string) error {
	log.Printf("💰 [Stripe Sync v2] Syncing single price: %s", priceID)

	// Fetch the price from Stripe
	p, err := price.Get(priceID, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch price %s: %w", priceID, err)
	}

	// Upsert to database
	if err := s.upsertPrice(p); err != nil {
		return fmt.Errorf("failed to upsert price %s: %w", priceID, err)
	}

	log.Printf("✅ [Stripe Sync v2] Price %s synced successfully", priceID)
	return nil
}

// SyncSingleCustomer syncs a single customer by ID from Stripe to v2 tables
func (s *StripeSyncV2Service) SyncSingleCustomer(ctx context.Context, customerID string) error {
	log.Printf("👥 [Stripe Sync v2] Syncing single customer: %s", customerID)

	// Fetch the customer from Stripe
	cust, err := customer.Get(customerID, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch customer %s: %w", customerID, err)
	}

	// Upsert to database
	if err := s.upsertCustomer(cust); err != nil {
		return fmt.Errorf("failed to upsert customer %s: %w", customerID, err)
	}

	log.Printf("✅ [Stripe Sync v2] Customer %s synced successfully", customerID)
	return nil
}

// SyncSingleSubscription syncs a single subscription by ID from Stripe to v2 tables
func (s *StripeSyncV2Service) SyncSingleSubscription(ctx context.Context, subscriptionID string) error {
	log.Printf("📋 [Stripe Sync v2] Syncing single subscription: %s", subscriptionID)

	// Fetch the subscription from Stripe
	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch subscription %s: %w", subscriptionID, err)
	}

	// Upsert to database
	if err := s.upsertSubscription(sub); err != nil {
		return fmt.Errorf("failed to upsert subscription %s: %w", subscriptionID, err)
	}

	log.Printf("✅ [Stripe Sync v2] Subscription %s synced successfully", subscriptionID)
	return nil
}

// GetDB returns the database connection (for webhook service)
func (s *StripeSyncV2Service) GetDB() *database.DB {
	return s.db
}
