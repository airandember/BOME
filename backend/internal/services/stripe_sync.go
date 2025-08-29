package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/invoice"
	"github.com/stripe/stripe-go/v74/price"
	"github.com/stripe/stripe-go/v74/product"
	"github.com/stripe/stripe-go/v74/subscription"

	"bome-backend/internal/database"
)

// StripeSyncService handles all Stripe data synchronization
type StripeSyncService struct {
	db            *database.DB
	stripeService *StripeService
	logger        *StripeLogger
}

// SyncConfig represents sync configuration for an entity type
type SyncConfig struct {
	EntityType          string                 `json:"entity_type"`
	SyncEnabled         bool                   `json:"sync_enabled"`
	SyncIntervalHours   int                    `json:"sync_interval_hours"`
	BatchSize           int                    `json:"batch_size"`
	LastFullSync        *time.Time             `json:"last_full_sync"`
	LastIncrementalSync *time.Time             `json:"last_incremental_sync"`
	ConfigData          map[string]interface{} `json:"config_data"`
}

// SyncJob represents a sync job
type SyncJob struct {
	ID             int        `json:"id"`
	JobType        string     `json:"job_type"`
	EntityType     string     `json:"entity_type"`
	Status         string     `json:"status"`
	TotalItems     int        `json:"total_items"`
	ProcessedItems int        `json:"processed_items"`
	LastCursor     string     `json:"last_cursor"`
	ErrorMessage   string     `json:"error_message"`
	StartedAt      *time.Time `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

// NewStripeSyncService creates a new sync service
func NewStripeSyncService(db *database.DB, stripeService *StripeService) *StripeSyncService {
	return &StripeSyncService{
		db:            db,
		stripeService: stripeService,
		logger:        NewStripeLogger("SYNC"),
	}
}

// GetDB returns the database instance for testing
func (s *StripeSyncService) GetDB() *database.DB {
	return s.db
}

// InitialDataSync performs the initial 1.5-year historical data sync
func (s *StripeSyncService) InitialDataSync(ctx context.Context) error {
	if !s.stripeService.IsEnabled() {
		return fmt.Errorf("stripe service is not enabled")
	}

	s.logger.LogSyncStart("INITIAL_SYNC", "ALL_ENTITIES", 0)

	// Calculate 1.5 years ago
	oneAndHalfYearsAgo := time.Now().AddDate(-1, -6, 0)

	// Define sync order (dependencies matter)
	syncOrder := []string{
		"product",
		"price",
		"customer",
		"subscription",
		"invoice",
		"coupon",
	}

	// Create initial sync job
	jobID, err := s.createSyncJob("initial_sync", "all", len(syncOrder))
	if err != nil {
		return fmt.Errorf("failed to create sync job: %w", err)
	}

	defer func() {
		s.completeSyncJob(jobID, err)
	}()

	// Sync each entity type in order
	for i, entityType := range syncOrder {
		log.Printf("📊 Syncing %s data (step %d/%d)", entityType, i+1, len(syncOrder))

		switch entityType {
		case "product":
			err = s.syncProducts(ctx, oneAndHalfYearsAgo, jobID)
		case "price":
			err = s.syncPrices(ctx, oneAndHalfYearsAgo, jobID)
		case "customer":
			err = s.syncCustomers(ctx, oneAndHalfYearsAgo, jobID)
		case "subscription":
			err = s.syncSubscriptions(ctx, oneAndHalfYearsAgo, jobID)
		case "invoice":
			err = s.syncInvoices(ctx, oneAndHalfYearsAgo, jobID)
		case "coupon":
			err = s.syncCoupons(ctx, oneAndHalfYearsAgo, jobID)
		}

		if err != nil {
			log.Printf("❌ Failed to sync %s: %v", entityType, err)
			return fmt.Errorf("failed to sync %s: %w", entityType, err)
		}

		// Update progress
		s.updateSyncJobProgress(jobID, i+1)
		log.Printf("✅ Completed %s sync", entityType)
	}

	log.Println("🎉 Initial data sync completed successfully!")
	return nil
}

// syncProducts syncs products with batch processing
func (s *StripeSyncService) syncProducts(ctx context.Context, since time.Time, jobID int) error {
	log.Println("📦 Starting product sync with batch processing...")

	config, err := s.getSyncConfig("product")
	if err != nil {
		return err
	}

	params := &stripe.ProductListParams{}
	params.Created = stripe.Int64(since.Unix())
	params.Limit = stripe.Int64(int64(config.BatchSize))

	iter := product.List(params)
	batchCount := 0
	totalProcessed := 0

	for iter.Next() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		prod := iter.Current().(*stripe.Product)

		// Insert/update product with conflict resolution
		err := s.upsertProduct(prod)
		if err != nil {
			log.Printf("⚠️ Failed to upsert product %s: %v", prod.ID, err)
			continue
		}

		totalProcessed++

		// Batch processing - pause every 50 items for low resource usage
		if totalProcessed%50 == 0 {
			batchCount++
			log.Printf("📦 Processed %d products (batch %d)", totalProcessed, batchCount)

			// Small pause to be gentle on resources
			time.Sleep(100 * time.Millisecond)
		}

		// Rate limiting - respect Stripe's limits
		if totalProcessed%100 == 0 {
			log.Printf("⏸️ Rate limiting pause after %d products", totalProcessed)
			time.Sleep(1 * time.Second)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("stripe API error during product sync: %w", err)
	}

	log.Printf("✅ Product sync completed: %d products processed", totalProcessed)
	return nil
}

// syncCustomers syncs customers with batch processing
func (s *StripeSyncService) syncCustomers(ctx context.Context, since time.Time, jobID int) error {
	log.Println("👥 Starting customer sync with batch processing...")

	config, err := s.getSyncConfig("customer")
	if err != nil {
		return err
	}

	params := &stripe.CustomerListParams{}
	params.Created = stripe.Int64(since.Unix())
	params.Limit = stripe.Int64(int64(config.BatchSize))

	iter := customer.List(params)
	batchCount := 0
	totalProcessed := 0

	for iter.Next() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		cust := iter.Current().(*stripe.Customer)

		// Insert/update customer
		err := s.upsertCustomer(cust)
		if err != nil {
			log.Printf("⚠️ Failed to upsert customer %s: %v", cust.ID, err)
			continue
		}

		totalProcessed++

		// Batch processing
		if totalProcessed%50 == 0 {
			batchCount++
			log.Printf("👥 Processed %d customers (batch %d)", totalProcessed, batchCount)
			time.Sleep(100 * time.Millisecond)
		}

		// Rate limiting for large accounts
		if totalProcessed%100 == 0 {
			log.Printf("⏸️ Rate limiting pause after %d customers", totalProcessed)
			time.Sleep(1 * time.Second)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("stripe API error during customer sync: %w", err)
	}

	log.Printf("✅ Customer sync completed: %d customers processed", totalProcessed)
	return nil
}

// syncPrices syncs prices with batch processing
func (s *StripeSyncService) syncPrices(ctx context.Context, since time.Time, jobID int) error {
	log.Println("💰 Starting price sync with batch processing...")

	config, err := s.getSyncConfig("price")
	if err != nil {
		return err
	}

	params := &stripe.PriceListParams{}
	params.Created = stripe.Int64(since.Unix())
	params.Limit = stripe.Int64(int64(config.BatchSize))

	iter := price.List(params)
	totalProcessed := 0

	for iter.Next() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		pr := iter.Current().(*stripe.Price)

		err := s.upsertPrice(pr)
		if err != nil {
			log.Printf("⚠️ Failed to upsert price %s: %v", pr.ID, err)
			continue
		}

		totalProcessed++

		if totalProcessed%50 == 0 {
			log.Printf("💰 Processed %d prices", totalProcessed)
			time.Sleep(100 * time.Millisecond)
		}
	}

	log.Printf("✅ Price sync completed: %d prices processed", totalProcessed)
	return nil
}

// syncSubscriptions syncs subscriptions with batch processing
func (s *StripeSyncService) syncSubscriptions(ctx context.Context, since time.Time, jobID int) error {
	log.Println("📋 Starting subscription sync with batch processing...")

	config, err := s.getSyncConfig("subscription")
	if err != nil {
		return err
	}

	params := &stripe.SubscriptionListParams{}
	params.Created = stripe.Int64(since.Unix())
	params.Limit = stripe.Int64(int64(config.BatchSize))

	iter := subscription.List(params)
	totalProcessed := 0

	for iter.Next() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		sub := iter.Current().(*stripe.Subscription)

		err := s.upsertSubscription(sub)
		if err != nil {
			log.Printf("⚠️ Failed to upsert subscription %s: %v", sub.ID, err)
			continue
		}

		totalProcessed++

		if totalProcessed%25 == 0 {
			log.Printf("📋 Processed %d subscriptions", totalProcessed)
			time.Sleep(100 * time.Millisecond)
		}
	}

	log.Printf("✅ Subscription sync completed: %d subscriptions processed", totalProcessed)
	return nil
}

// syncInvoices syncs invoices with batch processing
func (s *StripeSyncService) syncInvoices(ctx context.Context, since time.Time, jobID int) error {
	log.Println("🧾 Starting invoice sync with batch processing...")

	config, err := s.getSyncConfig("invoice")
	if err != nil {
		return err
	}

	params := &stripe.InvoiceListParams{}
	params.Created = stripe.Int64(since.Unix())
	params.Limit = stripe.Int64(int64(config.BatchSize))

	iter := invoice.List(params)
	totalProcessed := 0

	for iter.Next() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		inv := iter.Current().(*stripe.Invoice)

		err := s.upsertInvoice(inv)
		if err != nil {
			log.Printf("⚠️ Failed to upsert invoice %s: %v", inv.ID, err)
			continue
		}

		totalProcessed++

		if totalProcessed%25 == 0 {
			log.Printf("🧾 Processed %d invoices", totalProcessed)
			time.Sleep(100 * time.Millisecond)
		}
	}

	log.Printf("✅ Invoice sync completed: %d invoices processed", totalProcessed)
	return nil
}

// Placeholder for other sync methods
func (s *StripeSyncService) syncPaymentIntents(ctx context.Context, since time.Time, jobID int) error {
	log.Println("💳 Payment intent sync - implementing...")
	return nil
}

func (s *StripeSyncService) syncCoupons(ctx context.Context, since time.Time, jobID int) error {
	log.Println("🎟️ Coupon sync - implementing...")
	return nil
}

// Database operations with conflict resolution
func (s *StripeSyncService) upsertProduct(prod *stripe.Product) error {
	query := `
		INSERT INTO stripe_products (stripe_id, name, description, active, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (stripe_id) 
		DO UPDATE SET 
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			active = EXCLUDED.active,
			created_at = EXCLUDED.created_at
	`

	_, err := s.db.Exec(query,
		prod.ID,
		prod.Name,
		prod.Description,
		prod.Active,
		time.Unix(prod.Created, 0),
	)
	return err
}

func (s *StripeSyncService) upsertCustomer(cust *stripe.Customer) error {
	metadataJSON, _ := json.Marshal(cust.Metadata)

	query := `
		INSERT INTO stripe_customers (stripe_id, email, name, created_at, updated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (stripe_id) 
		DO UPDATE SET 
			email = EXCLUDED.email,
			name = EXCLUDED.name,
			updated_at = EXCLUDED.updated_at,
			metadata = EXCLUDED.metadata
	`

	_, err := s.db.Exec(query,
		cust.ID,
		cust.Email,
		cust.Name,
		time.Unix(cust.Created, 0),
		time.Now(),
		metadataJSON,
	)
	return err
}

func (s *StripeSyncService) upsertPrice(pr *stripe.Price) error {
	// Get product ID from our database
	var productID sql.NullInt64
	err := s.db.QueryRow("SELECT id FROM stripe_products WHERE stripe_id = $1", pr.Product.ID).Scan(&productID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	var recurringInterval string
	if pr.Recurring != nil {
		recurringInterval = string(pr.Recurring.Interval)
	}

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
	return err
}

func (s *StripeSyncService) upsertSubscription(sub *stripe.Subscription) error {
	// Get customer ID from our database
	var customerID sql.NullInt64
	err := s.db.QueryRow("SELECT id FROM stripe_customers WHERE stripe_id = $1", sub.Customer.ID).Scan(&customerID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	query := `
		INSERT INTO stripe_subscriptions (stripe_id, customer_id, status, current_period_start, current_period_end, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (stripe_id) 
		DO UPDATE SET 
			customer_id = EXCLUDED.customer_id,
			status = EXCLUDED.status,
			current_period_start = EXCLUDED.current_period_start,
			current_period_end = EXCLUDED.current_period_end
	`

	_, err = s.db.Exec(query,
		sub.ID,
		customerID,
		string(sub.Status),
		time.Unix(sub.CurrentPeriodStart, 0),
		time.Unix(sub.CurrentPeriodEnd, 0),
		time.Unix(sub.Created, 0),
	)
	return err
}

func (s *StripeSyncService) upsertInvoice(inv *stripe.Invoice) error {
	// Get customer and subscription IDs
	var customerID, subscriptionID sql.NullInt64

	if inv.Customer != nil {
		s.db.QueryRow("SELECT id FROM stripe_customers WHERE stripe_id = $1", inv.Customer.ID).Scan(&customerID)
	}

	if inv.Subscription != nil {
		s.db.QueryRow("SELECT id FROM stripe_subscriptions WHERE stripe_id = $1", inv.Subscription.ID).Scan(&subscriptionID)
	}

	query := `
		INSERT INTO stripe_invoices (stripe_id, customer_id, subscription_id, amount_paid, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (stripe_id) 
		DO UPDATE SET 
			customer_id = EXCLUDED.customer_id,
			subscription_id = EXCLUDED.subscription_id,
			amount_paid = EXCLUDED.amount_paid,
			status = EXCLUDED.status
	`

	_, err := s.db.Exec(query,
		inv.ID,
		customerID,
		subscriptionID,
		inv.AmountPaid,
		string(inv.Status),
		time.Unix(inv.Created, 0),
	)
	return err
}

// Sync job management
func (s *StripeSyncService) createSyncJob(jobType, entityType string, totalItems int) (int, error) {
	query := `
		INSERT INTO stripe_sync_jobs (job_type, entity_type, status, total_items, started_at)
		VALUES ($1, $2, 'running', $3, NOW())
		RETURNING id
	`

	var jobID int
	err := s.db.QueryRow(query, jobType, entityType, totalItems).Scan(&jobID)
	return jobID, err
}

func (s *StripeSyncService) updateSyncJobProgress(jobID, processedItems int) error {
	query := `UPDATE stripe_sync_jobs SET processed_items = $1, updated_at = NOW() WHERE id = $2`
	_, err := s.db.Exec(query, processedItems, jobID)
	return err
}

func (s *StripeSyncService) completeSyncJob(jobID int, jobError error) error {
	status := "completed"
	var errorMsg *string

	if jobError != nil {
		status = "failed"
		errStr := jobError.Error()
		errorMsg = &errStr
	}

	query := `
		UPDATE stripe_sync_jobs 
		SET status = $1, error_message = $2, completed_at = NOW(), updated_at = NOW() 
		WHERE id = $3
	`

	_, err := s.db.Exec(query, status, errorMsg, jobID)
	return err
}

// getSyncConfig gets sync configuration for an entity type
func (s *StripeSyncService) getSyncConfig(entityType string) (*SyncConfig, error) {
	query := `
		SELECT entity_type, sync_enabled, sync_interval_hours, batch_size, 
		       last_full_sync, last_incremental_sync, config_data
		FROM stripe_sync_config 
		WHERE entity_type = $1
	`

	config := &SyncConfig{}
	var configDataJSON []byte

	err := s.db.QueryRow(query, entityType).Scan(
		&config.EntityType,
		&config.SyncEnabled,
		&config.SyncIntervalHours,
		&config.BatchSize,
		&config.LastFullSync,
		&config.LastIncrementalSync,
		&configDataJSON,
	)

	if err == sql.ErrNoRows {
		// Return default config if not found
		return &SyncConfig{
			EntityType:        entityType,
			SyncEnabled:       true,
			SyncIntervalHours: 6,
			BatchSize:         100,
			ConfigData:        make(map[string]interface{}),
		}, nil
	}

	if err != nil {
		return nil, err
	}

	if configDataJSON != nil {
		json.Unmarshal(configDataJSON, &config.ConfigData)
	}

	return config, nil
}

// Webhook-specific methods for real-time updates
func (s *StripeSyncService) UpsertCustomerFromWebhook(cust *stripe.Customer) error {
	log.Printf("🔄 Webhook sync: Updating customer %s", cust.ID)
	return s.upsertCustomer(cust)
}

func (s *StripeSyncService) MarkCustomerDeleted(customerID string) error {
	log.Printf("🗑️ Webhook sync: Marking customer %s as deleted", customerID)

	// Instead of deleting, we mark as deleted for audit trail
	query := `
		UPDATE stripe_customers 
		SET updated_at = NOW(), metadata = jsonb_set(COALESCE(metadata, '{}'), '{deleted}', 'true')
		WHERE stripe_id = $1
	`

	_, err := s.db.Exec(query, customerID)
	return err
}

func (s *StripeSyncService) UpsertInvoiceFromWebhook(inv *stripe.Invoice) error {
	log.Printf("🔄 Webhook sync: Updating invoice %s", inv.ID)
	return s.upsertInvoice(inv)
}

func (s *StripeSyncService) UpsertProductFromWebhook(prod *stripe.Product) error {
	log.Printf("🔄 Webhook sync: Updating product %s", prod.ID)
	return s.upsertProduct(prod)
}

func (s *StripeSyncService) UpsertPriceFromWebhook(pr *stripe.Price) error {
	log.Printf("🔄 Webhook sync: Updating price %s", pr.ID)
	return s.upsertPrice(pr)
}

func (s *StripeSyncService) UpsertSubscriptionFromWebhook(sub *stripe.Subscription) error {
	log.Printf("🔄 Webhook sync: Updating subscription %s", sub.ID)
	return s.upsertSubscription(sub)
}

func (s *StripeSyncService) MarkSubscriptionDeleted(subscriptionID string) error {
	log.Printf("🗑️ Webhook sync: Marking subscription %s as deleted", subscriptionID)

	// Instead of deleting, we mark as deleted for audit trail
	query := `
		UPDATE stripe_subscriptions 
		SET status = 'canceled'
		WHERE stripe_id = $1
	`

	_, err := s.db.Exec(query, subscriptionID)
	return err
}

// IncrementalSync performs incremental sync from a specific date
func (s *StripeSyncService) IncrementalSync(ctx context.Context, since time.Time) error {
	if !s.stripeService.IsEnabled() {
		return fmt.Errorf("stripe service is not enabled")
	}

	log.Printf("🔄 Starting incremental sync from %s", since.Format("2006-01-02 15:04:05"))

	// Create incremental sync job
	jobID, err := s.createSyncJob("incremental_sync", "all", 7)
	if err != nil {
		return fmt.Errorf("failed to create incremental sync job: %w", err)
	}

	defer func() {
		s.completeSyncJob(jobID, err)
	}()

	// Sync each entity type incrementally
	entityTypes := []string{"customer", "product", "price", "subscription", "invoice", "coupon"}

	for i, entityType := range entityTypes {
		log.Printf("🔄 Incremental sync: %s (step %d/%d)", entityType, i+1, len(entityTypes))

		switch entityType {
		case "customer":
			err = s.syncCustomers(ctx, since, jobID)
		case "product":
			err = s.syncProducts(ctx, since, jobID)
		case "price":
			err = s.syncPrices(ctx, since, jobID)
		case "subscription":
			err = s.syncSubscriptions(ctx, since, jobID)
		case "invoice":
			err = s.syncInvoices(ctx, since, jobID)
		case "coupon":
			err = s.syncCoupons(ctx, since, jobID)
		}

		if err != nil {
			log.Printf("❌ Incremental sync failed for %s: %v", entityType, err)
			return fmt.Errorf("incremental sync failed for %s: %w", entityType, err)
		}

		s.updateSyncJobProgress(jobID, i+1)
	}

	log.Println("✅ Incremental sync completed successfully")
	return nil
}

// CleanupOldSyncJobs removes old sync job records
func (s *StripeSyncService) CleanupOldSyncJobs(ctx context.Context, daysToKeep int) error {
	log.Printf("🧹 Cleaning up sync jobs older than %d days", daysToKeep)

	query := `
		DELETE FROM stripe_sync_jobs 
		WHERE created_at < NOW() - INTERVAL '%d days'
		AND status IN ('completed', 'failed')
	`

	result, err := s.db.Exec(fmt.Sprintf(query, daysToKeep))
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("🧹 Cleaned up %d old sync job records", rowsAffected)

	return nil
}

// UpdateAggregationTables updates the daily and monthly metrics tables
func (s *StripeSyncService) UpdateAggregationTables(ctx context.Context) error {
	log.Println("📊 Updating aggregation tables...")

	// Update daily revenue table
	err := s.updateDailyRevenue()
	if err != nil {
		log.Printf("⚠️ Failed to update daily revenue: %v", err)
	}

	// Update monthly metrics table
	err = s.updateMonthlyMetrics()
	if err != nil {
		log.Printf("⚠️ Failed to update monthly metrics: %v", err)
	}

	log.Println("📊 Aggregation tables updated")
	return nil
}

// updateDailyRevenue updates the stripe_daily_revenue table
func (s *StripeSyncService) updateDailyRevenue() error {
	query := `
		INSERT INTO stripe_daily_revenue (date, total_revenue, customer_count, subscription_count)
		SELECT 
			DATE(si.created_at) as date,
			SUM(si.amount_paid) as total_revenue,
			COUNT(DISTINCT si.customer_id) as customer_count,
			COUNT(DISTINCT si.subscription_id) as subscription_count
		FROM stripe_invoices si
		WHERE si.status = 'paid'
		AND DATE(si.created_at) >= CURRENT_DATE - INTERVAL '7 days'
		GROUP BY DATE(si.created_at)
		ON CONFLICT (date) 
		DO UPDATE SET 
			total_revenue = EXCLUDED.total_revenue,
			customer_count = EXCLUDED.customer_count,
			subscription_count = EXCLUDED.subscription_count
	`

	_, err := s.db.Exec(query)
	return err
}

// updateMonthlyMetrics updates the stripe_monthly_metrics table
func (s *StripeSyncService) updateMonthlyMetrics() error {
	query := `
		INSERT INTO stripe_monthly_metrics (year_month, mrr, arr, churn_rate, growth_rate)
		SELECT 
			TO_CHAR(DATE_TRUNC('month', ss.created_at), 'YYYY-MM') as year_month,
			SUM(CASE WHEN sp.recurring_interval = 'month' THEN sp.unit_amount ELSE sp.unit_amount/12 END) as mrr,
			SUM(CASE WHEN sp.recurring_interval = 'year' THEN sp.unit_amount ELSE sp.unit_amount*12 END) as arr,
			0.00 as churn_rate, -- Calculate churn rate separately
			0.00 as growth_rate -- Calculate growth rate separately
		FROM stripe_subscriptions ss
		JOIN stripe_customers sc ON ss.customer_id = sc.id
		JOIN stripe_prices sp ON sp.id = (
			SELECT sp2.id FROM stripe_prices sp2 
			WHERE sp2.stripe_id IN (
				SELECT jsonb_array_elements_text(ss.metadata->'price_ids')
			) LIMIT 1
		)
		WHERE ss.status = 'active'
		AND DATE_TRUNC('month', ss.created_at) >= DATE_TRUNC('month', CURRENT_DATE - INTERVAL '3 months')
		GROUP BY TO_CHAR(DATE_TRUNC('month', ss.created_at), 'YYYY-MM')
		ON CONFLICT (year_month) 
		DO UPDATE SET 
			mrr = EXCLUDED.mrr,
			arr = EXCLUDED.arr
	`

	_, err := s.db.Exec(query)
	return err
}
