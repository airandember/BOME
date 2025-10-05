package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/coupon"
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
	CreatedAt           *time.Time             `json:"created_at"`
	UpdatedAt           *time.Time             `json:"updated_at"`
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
	UpdatedAt      *time.Time `json:"updated_at"`
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

// TestCustomerSync manually triggers a customer sync for testing (no date filter)
func (s *StripeSyncService) TestCustomerSync(ctx context.Context) error {
	log.Println("🧪 TEST: Manual customer sync triggered...")

	// Use a dummy time since we're not filtering by date anyway
	dummyTime := time.Now().AddDate(-10, 0, 0) // 10 years ago

	return s.syncCustomers(ctx, dummyTime, 0)
}

// TestCustomerSyncUnlimited pulls ALL customers without any limits
func (s *StripeSyncService) TestCustomerSyncUnlimited(ctx context.Context) error {
	log.Println("🚀 TEST: UNLIMITED customer sync - pulling ALL customers...")

	params := &stripe.CustomerListParams{}
	// No date filter, no limits - get everything
	params.Limit = stripe.Int64(100) // Stripe's max per request

	iter := customer.List(params)
	totalProcessed := 0
	batchCount := 0

	log.Printf("📊 Starting unlimited customer sync...")

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

		// Progress logging every 100 customers
		if totalProcessed%100 == 0 {
			batchCount++
			log.Printf("📈 Progress: %d customers processed (batch %d)", totalProcessed, batchCount)

			// Small pause for rate limiting
			time.Sleep(50 * time.Millisecond)
		}

		// Longer pause every 500 customers for large accounts
		if totalProcessed%500 == 0 {
			log.Printf("⏸️ Rate limiting pause after %d customers", totalProcessed)
			time.Sleep(500 * time.Millisecond)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("stripe API error during unlimited customer sync: %w", err)
	}

	log.Printf("🎉 UNLIMITED customer sync completed: %d customers processed", totalProcessed)
	return nil
}

// SyncCouponsManual performs manual coupon sync (all historical coupons)
func (s *StripeSyncService) SyncCouponsManual(ctx context.Context) error {
	if !s.stripeService.IsEnabled() {
		return fmt.Errorf("stripe service is not enabled")
	}

	log.Println("🎟️ Manual coupon sync - pulling ALL coupons...")

	// Create sync job
	jobID, err := s.createSyncJob("manual_sync", "coupon", 1)
	if err != nil {
		return fmt.Errorf("failed to create sync job: %w", err)
	}

	defer func() {
		s.completeSyncJob(jobID, err)
	}()

	// Sync all coupons (no time filter)
	err = s.syncCoupons(ctx, time.Time{}, jobID)
	if err != nil {
		return fmt.Errorf("failed to sync coupons: %w", err)
	}

	log.Println("✅ Manual coupon sync completed successfully")
	return nil
}

// SyncMonthlyMetricsManual performs manual monthly metrics calculation
func (s *StripeSyncService) SyncMonthlyMetricsManual(ctx context.Context) error {
	if !s.stripeService.IsEnabled() {
		return fmt.Errorf("stripe service is not enabled")
	}

	log.Println("📊 Manual monthly metrics sync - calculating last 2 years...")

	// Create sync job
	jobID, err := s.createSyncJob("manual_sync", "monthly_metrics", 24) // 24 months
	if err != nil {
		return fmt.Errorf("failed to create sync job: %w", err)
	}

	defer func() {
		s.completeSyncJob(jobID, err)
	}()

	// Calculate metrics for last 2 years
	twoYearsAgo := time.Now().AddDate(-2, 0, 0)
	err = s.syncMonthlyMetrics(ctx, twoYearsAgo, jobID)
	if err != nil {
		return fmt.Errorf("failed to sync monthly metrics: %w", err)
	}

	log.Println("✅ Manual monthly metrics sync completed successfully")
	return nil
}

// SyncProductsManual performs manual product sync (all products)
func (s *StripeSyncService) SyncProductsManual(ctx context.Context) error {
	if !s.stripeService.IsEnabled() {
		return fmt.Errorf("stripe service is not enabled")
	}

	log.Println("📦 Manual product sync - pulling ALL products...")

	// Create sync job
	jobID, err := s.createSyncJob("manual_sync", "product", 1)
	if err != nil {
		return fmt.Errorf("failed to create sync job: %w", err)
	}

	// Ensure job completion is tracked
	defer func() {
		s.completeSyncJob(jobID, err)
	}()

	// Sync all products (no time limit)
	allTime := time.Time{} // Zero time means all products
	err = s.syncProducts(ctx, allTime, jobID)
	if err != nil {
		return fmt.Errorf("failed to sync products: %w", err)
	}

	log.Println("✅ Manual product sync completed successfully")
	return nil
}

// SyncPricesManual performs manual price sync (all prices)
func (s *StripeSyncService) SyncPricesManual(ctx context.Context) error {
	if !s.stripeService.IsEnabled() {
		return fmt.Errorf("stripe service is not enabled")
	}

	log.Println("💰 Manual price sync - pulling ALL prices...")

	// Create sync job
	jobID, err := s.createSyncJob("manual_sync", "price", 1)
	if err != nil {
		return fmt.Errorf("failed to create sync job: %w", err)
	}

	// Ensure job completion is tracked
	defer func() {
		s.completeSyncJob(jobID, err)
	}()

	// Sync all prices (no time limit)
	allTime := time.Time{} // Zero time means all prices
	err = s.syncPrices(ctx, allTime, jobID)
	if err != nil {
		return fmt.Errorf("failed to sync prices: %w", err)
	}

	log.Println("✅ Manual price sync completed successfully")
	return nil
}

// SyncSubscriptionsManual performs manual subscription sync (all subscriptions)
func (s *StripeSyncService) SyncSubscriptionsManual(ctx context.Context) error {
	if !s.stripeService.IsEnabled() {
		return fmt.Errorf("stripe service is not enabled")
	}

	log.Println("💳 Manual subscription sync - pulling ALL subscriptions...")

	// Create sync job
	jobID, err := s.createSyncJob("manual_sync", "subscription", 1)
	if err != nil {
		return fmt.Errorf("failed to create sync job: %w", err)
	}

	// Ensure job completion is tracked
	defer func() {
		s.completeSyncJob(jobID, err)
	}()

	// Sync all subscriptions (no time limit)
	allTime := time.Time{} // Zero time means all subscriptions
	err = s.syncSubscriptions(ctx, allTime, jobID)
	if err != nil {
		return fmt.Errorf("failed to sync subscriptions: %w", err)
	}

	log.Println("✅ Manual subscription sync completed successfully")
	return nil
}

// CleanupOrphanedSubscriptions cleans up subscriptions with invalid product references
func (s *StripeSyncService) CleanupOrphanedSubscriptions(ctx context.Context) error {
	log.Println("🧹 Starting invalid product subscription cleanup...")

	// Create sync job for tracking
	jobID, err := s.createSyncJob("cleanup", "subscription", 1)
	if err != nil {
		return fmt.Errorf("failed to create cleanup job: %w", err)
	}

	// Ensure job completion is tracked
	defer func() {
		s.completeSyncJob(jobID, err)
	}()

	// First, get statistics about orphaned subscriptions
	stats, err := s.getOrphanedSubscriptionStats()
	if err != nil {
		return fmt.Errorf("failed to get orphaned subscription stats: %w", err)
	}

	log.Printf("📊 Orphaned subscription analysis:")
	log.Printf("   Total active subscriptions: %d", stats.TotalActive)
	log.Printf("   Missing product names: %d", stats.MissingProductNames)
	log.Printf("   Orphaned product IDs: %d", stats.OrphanedProductIDs)
	log.Printf("   Total problematic: %d", stats.TotalProblematic)

	if stats.TotalProblematic == 0 {
		log.Println("✅ No invalid product subscriptions found - database is clean!")
		return nil
	}

	// Perform cleanup - mark problematic subscriptions as having invalid products
	query := `
		UPDATE stripe_subscriptions 
		SET 
			status = 'invalid_product'
		WHERE id IN (
			SELECT ss.id
			FROM stripe_subscriptions ss
			LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
			WHERE ss.status IN ('active', 'trialing') 
				AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
				AND (
					-- Missing product name
					(ss.product_name IS NULL OR ss.product_name = '')
					OR 
					-- Orphaned product ID
					(ss.stripe_product_id IS NOT NULL AND sp.stripe_id IS NULL)
				)
		)
	`

	result, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to cleanup orphaned subscriptions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("✅ Successfully marked %d orphaned subscriptions as 'invalid_product'", rowsAffected)

	// Verify cleanup
	verifyStats, err := s.getOrphanedSubscriptionStats()
	if err != nil {
		log.Printf("⚠️ Failed to verify cleanup: %v", err)
	} else {
		log.Printf("🔍 Post-cleanup verification:")
		log.Printf("   Total active subscriptions: %d", verifyStats.TotalActive)
		log.Printf("   Remaining problematic: %d", verifyStats.TotalProblematic)
		if verifyStats.TotalProblematic == 0 {
			log.Println("✅ Cleanup successful - no more invalid product subscriptions!")
		} else {
			log.Printf("⚠️ Still found %d problematic subscriptions after cleanup", verifyStats.TotalProblematic)
		}
	}

	log.Println("✅ Invalid product subscription cleanup completed")
	return nil
}

// OrphanedSubscriptionStats holds statistics about orphaned subscriptions
type OrphanedSubscriptionStats struct {
	TotalActive         int
	MissingProductNames int
	OrphanedProductIDs  int
	TotalProblematic    int
}

// getOrphanedSubscriptionStats returns statistics about orphaned subscriptions
func (s *StripeSyncService) getOrphanedSubscriptionStats() (*OrphanedSubscriptionStats, error) {
	stats := &OrphanedSubscriptionStats{}

	// Total active subscriptions
	err := s.db.QueryRow(`
		SELECT COUNT(*) 
		FROM stripe_subscriptions 
		WHERE status IN ('active', 'trialing') 
			AND (current_period_end IS NULL OR current_period_end > NOW())
	`).Scan(&stats.TotalActive)
	if err != nil {
		return nil, err
	}

	// Missing product names
	err = s.db.QueryRow(`
		SELECT COUNT(*) 
		FROM stripe_subscriptions 
		WHERE status IN ('active', 'trialing') 
			AND (current_period_end IS NULL OR current_period_end > NOW())
			AND (product_name IS NULL OR product_name = '')
	`).Scan(&stats.MissingProductNames)
	if err != nil {
		return nil, err
	}

	// Orphaned product IDs
	err = s.db.QueryRow(`
		SELECT COUNT(*) 
		FROM stripe_subscriptions ss
		LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
		WHERE ss.status IN ('active', 'trialing') 
			AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
			AND ss.stripe_product_id IS NOT NULL
			AND sp.stripe_id IS NULL
	`).Scan(&stats.OrphanedProductIDs)
	if err != nil {
		return nil, err
	}

	// Total problematic
	stats.TotalProblematic = stats.MissingProductNames + stats.OrphanedProductIDs

	return stats, nil
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
		"monthly_metrics", // Calculate after all data is synced
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
		case "monthly_metrics":
			err = s.syncMonthlyMetrics(ctx, oneAndHalfYearsAgo, jobID)
		}

		if err != nil {
			log.Printf("❌ Failed to sync %s: %v", entityType, err)
			return fmt.Errorf("failed to sync %s: %w", entityType, err)
		}

		// Update progress
		s.updateSyncJobProgress(jobID, i+1, "")
		log.Printf("✅ Completed %s sync", entityType)
	}

	log.Println("🎉 Initial data sync completed successfully!")
	return nil
}

// syncProducts syncs products with batch processing
func (s *StripeSyncService) syncProducts(ctx context.Context, since time.Time, jobID int) error {
	log.Println("📦 Starting product sync with batch processing (ALL products)...")

	config, err := s.getSyncConfig("product")
	if err != nil {
		return err
	}

	params := &stripe.ProductListParams{}
	// Remove date filter - we want ALL products regardless of creation date
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
	// TEMPORARILY REMOVE DATE FILTER TO GET ALL CUSTOMERS
	// params.Created = stripe.Int64(since.Unix())
	log.Printf("🔍 Syncing ALL customers (no date filter) to test data availability...")
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
	log.Println("💰 Starting price sync with batch processing (ALL prices)...")

	config, err := s.getSyncConfig("price")
	if err != nil {
		return err
	}

	params := &stripe.PriceListParams{}
	// Remove date filter - we want ALL prices regardless of creation date
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
	log.Println("📋 Starting subscription sync with batch processing (past 2 years, active/canceled/trialing)...")

	config, err := s.getSyncConfig("subscription")
	if err != nil {
		return err
	}

	// Target statuses we want to sync
	targetStatuses := []string{"active", "canceled", "trialing"}
	totalProcessed := 0

	log.Printf("📋 Syncing ALL subscriptions (no date filter) with statuses: %v", targetStatuses)

	// Sync each status separately for better control
	for _, status := range targetStatuses {
		log.Printf("📋 Syncing %s subscriptions (ALL historical data)...", status)

		params := &stripe.SubscriptionListParams{
			ListParams: stripe.ListParams{
				Limit: stripe.Int64(int64(config.BatchSize)),
			},
			Status: stripe.String(status),
		}

		// NO DATE FILTER - get ALL subscriptions regardless of creation date

		iter := subscription.List(params)
		statusProcessed := 0

		for iter.Next() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			sub := iter.Current().(*stripe.Subscription)

			err := s.upsertSubscription(sub)
			if err != nil {
				log.Printf("⚠️ Failed to upsert subscription %s (status: %s): %v", sub.ID, status, err)
				continue
			}

			statusProcessed++
			totalProcessed++

			if statusProcessed%25 == 0 {
				log.Printf("📋 Processed %d %s subscriptions", statusProcessed, status)
				time.Sleep(100 * time.Millisecond)
			}
		}

		if err := iter.Err(); err != nil {
			log.Printf("❌ Error iterating %s subscriptions: %v", status, err)
			return fmt.Errorf("failed to iterate %s subscriptions: %w", status, err)
		}

		log.Printf("✅ Completed %s subscriptions: %d processed", status, statusProcessed)
	}

	log.Printf("✅ Subscription sync completed: %d total subscriptions processed", totalProcessed)
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
	log.Println("🎟️ Starting coupon sync with batch processing...")

	config, err := s.getSyncConfig("coupon")
	if err != nil {
		return err
	}

	params := &stripe.CouponListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(int64(config.BatchSize)),
		},
	}

	// Add time filter for historical sync (coupons created since the specified date)
	if !since.IsZero() {
		params.Created = stripe.Int64(since.Unix())
	}

	iter := coupon.List(params)
	processed := 0

	for iter.Next() {
		select {
		case <-ctx.Done():
			log.Printf("⚠️ Coupon sync cancelled: %v", ctx.Err())
			return ctx.Err()
		default:
		}

		coup := iter.Current().(*stripe.Coupon)

		err := s.upsertCoupon(coup)
		if err != nil {
			log.Printf("⚠️ Failed to upsert coupon %s: %v", coup.ID, err)
			continue
		}

		processed++

		if processed%25 == 0 {
			log.Printf("   📊 Processed %d coupons...", processed)
		}
	}

	if iter.Err() != nil {
		log.Printf("❌ Error listing coupons: %v", iter.Err())
		return fmt.Errorf("failed to list coupons: %w", iter.Err())
	}

	log.Printf("✅ Completed coupon sync. Total processed: %d", processed)
	return nil
}

func (s *StripeSyncService) syncMonthlyMetrics(ctx context.Context, since time.Time, jobID int) error {
	log.Println("📊 Starting monthly metrics calculation for last 2 years...")

	// Calculate metrics for the last 2 years (24 months)
	now := time.Now()
	twoYearsAgo := now.AddDate(-2, 0, 0)

	// Start from the beginning of the month 2 years ago
	startDate := time.Date(twoYearsAgo.Year(), twoYearsAgo.Month(), 1, 0, 0, 0, 0, time.UTC)

	processed := 0

	// Process each month from 2 years ago to current month
	for currentMonth := startDate; currentMonth.Before(now) || currentMonth.Equal(time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)); currentMonth = currentMonth.AddDate(0, 1, 0) {
		select {
		case <-ctx.Done():
			log.Printf("⚠️ Monthly metrics sync cancelled: %v", ctx.Err())
			return ctx.Err()
		default:
		}

		yearMonth := currentMonth.Format("2006-01")
		log.Printf("   📈 Calculating metrics for %s...", yearMonth)

		metrics, err := s.calculateMonthlyMetrics(currentMonth)
		if err != nil {
			log.Printf("⚠️ Failed to calculate metrics for %s: %v", yearMonth, err)
			continue
		}

		err = s.upsertMonthlyMetrics(yearMonth, metrics)
		if err != nil {
			log.Printf("⚠️ Failed to upsert metrics for %s: %v", yearMonth, err)
			continue
		}

		processed++
	}

	log.Printf("✅ Completed monthly metrics sync. Total processed: %d months", processed)
	return nil
}

// Database operations with conflict resolution
func (s *StripeSyncService) upsertProduct(prod *stripe.Product) error {
	// Convert metadata to JSON
	metadataJSON, err := json.Marshal(prod.Metadata)
	if err != nil {
		log.Printf("⚠️ Failed to marshal metadata for product %s: %v", prod.ID, err)
		metadataJSON = []byte("{}")
	}

	// Convert images array to PostgreSQL array format
	var images []string
	if prod.Images != nil {
		images = prod.Images
	}

	// Convert package dimensions to JSON if present
	var packageDimensionsJSON []byte
	if prod.PackageDimensions != nil {
		packageDimensionsJSON, err = json.Marshal(map[string]interface{}{
			"height": prod.PackageDimensions.Height,
			"length": prod.PackageDimensions.Length,
			"weight": prod.PackageDimensions.Weight,
			"width":  prod.PackageDimensions.Width,
		})
		if err != nil {
			log.Printf("⚠️ Failed to marshal package dimensions for product %s: %v", prod.ID, err)
			packageDimensionsJSON = []byte("{}")
		}
	} else {
		packageDimensionsJSON = []byte("{}")
	}

	query := `
		INSERT INTO stripe_products (
			stripe_id, name, description, active, created_at, updated_at,
			metadata, url, images, package_dimensions, shippable, 
			statement_descriptor, tax_code, unit_label, livemode
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (stripe_id) 
		DO UPDATE SET 
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			active = EXCLUDED.active,
			updated_at = EXCLUDED.updated_at,
			metadata = EXCLUDED.metadata,
			url = EXCLUDED.url,
			images = EXCLUDED.images,
			package_dimensions = EXCLUDED.package_dimensions,
			shippable = EXCLUDED.shippable,
			statement_descriptor = EXCLUDED.statement_descriptor,
			tax_code = EXCLUDED.tax_code,
			unit_label = EXCLUDED.unit_label,
			livemode = EXCLUDED.livemode
	`

	_, err = s.db.Exec(query,
		prod.ID,                       // $1
		prod.Name,                     // $2
		prod.Description,              // $3
		prod.Active,                   // $4
		time.Unix(prod.Created, 0),    // $5
		time.Unix(prod.Updated, 0),    // $6
		string(metadataJSON),          // $7
		prod.URL,                      // $8
		pq.Array(images),              // $9
		string(packageDimensionsJSON), // $10
		prod.Shippable,                // $11
		prod.StatementDescriptor,      // $12
		prod.TaxCode,                  // $13
		prod.UnitLabel,                // $14
		prod.Livemode,                 // $15
	)

	if err != nil {
		log.Printf("❌ Failed to upsert product %s: %v", prod.ID, err)
		return err
	}

	log.Printf("✅ Upserted product: %s (%s)", prod.ID, prod.Name)
	return nil
}

func (s *StripeSyncService) upsertCustomer(cust *stripe.Customer) error {
	// Validate and fix metadata before storing
	validatedMetadata := s.validateAndFixCustomerMetadata(cust)
	metadataJSON, _ := json.Marshal(validatedMetadata)

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

	if err != nil {
		log.Printf("❌ Failed to upsert customer %s: %v", cust.ID, err)
	} else {
		log.Printf("✅ Upserted customer: %s (%s) with validated metadata", cust.ID, cust.Email)
	}

	return err
}

// validateAndFixCustomerMetadata ensures customer metadata has correct local_customer_id
func (s *StripeSyncService) validateAndFixCustomerMetadata(cust *stripe.Customer) map[string]interface{} {
	// Start with existing metadata
	validatedMetadata := make(map[string]interface{})
	for k, v := range cust.Metadata {
		validatedMetadata[k] = v
	}

	// Find the correct user ID for this Stripe customer
	var userID int
	err := s.db.QueryRow(`
		SELECT id FROM users 
		WHERE stripe_customer_id = $1 OR $1 = ANY(COALESCE(stripe_customer_ids, '{}'))
	`, cust.ID).Scan(&userID)

	if err == nil {
		// User found - ensure metadata has correct local_customer_id
		currentLocalCustomerID, exists := validatedMetadata["local_customer_id"]

		if !exists || currentLocalCustomerID != fmt.Sprintf("%d", userID) {
			log.Printf("🔧 Fixing metadata for customer %s: setting local_customer_id to %d", cust.ID, userID)
			validatedMetadata["local_customer_id"] = fmt.Sprintf("%d", userID)
		}
	} else {
		// User not found - this might be a new customer or orphaned record
		log.Printf("⚠️ No user found for Stripe customer %s (%s)", cust.ID, cust.Email)

		// Keep existing metadata but don't add incorrect local_customer_id
		if _, exists := validatedMetadata["local_customer_id"]; exists {
			log.Printf("🔧 Removing potentially incorrect local_customer_id from orphaned customer %s", cust.ID)
			delete(validatedMetadata, "local_customer_id")
		}
	}

	return validatedMetadata
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

// ProcessVideoAccessForCustomer grants video access to a customer based on their active subscription
func (s *StripeSyncService) ProcessVideoAccessForCustomer(customerID string) error {
	log.Printf("🎥 Processing video access for Stripe customer: %s", customerID)

	// Find the user associated with this Stripe customer
	var userID int
	var userEmail string
	err := s.db.QueryRow(`
		SELECT id, email FROM users 
		WHERE stripe_customer_id = $1 OR $1 = ANY(COALESCE(stripe_customer_ids, '{}'))
	`, customerID).Scan(&userID, &userEmail)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("⚠️ No user found for Stripe customer %s - this shouldn't happen in normal flow", customerID)
			log.Printf("🔍 Customer %s may have subscribed without going through our signup process", customerID)

			// This is unusual - typically users sign up first, then subscribe
			// We'll try to find them by email and link them, but won't create new users
			// since that would bypass our normal registration flow
			linkedUserID, err := s.linkExistingUserByEmail(customerID)
			if err != nil {
				log.Printf("❌ Failed to link existing user for Stripe customer %s: %v", customerID, err)
				log.Printf("ℹ️ User should complete normal signup process to access videos")
				return nil // Don't fail the webhook - user can register normally later
			}

			if linkedUserID > 0 {
				userID = linkedUserID
				log.Printf("✅ Linked existing user %d to Stripe customer %s", userID, customerID)
			} else {
				log.Printf("ℹ️ No existing user found with matching email for Stripe customer %s", customerID)
				log.Printf("💡 User should complete normal signup process to access videos")
				return nil // User needs to register through normal flow
			}
		} else {
			return fmt.Errorf("failed to find user for customer %s: %w", customerID, err)
		}
	}

	log.Printf("🔍 Found user %d (%s) for Stripe customer %s", userID, userEmail, customerID)

	// Check if user has an active subscription with video-approved products
	var hasVideoAccess bool
	err = s.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM stripe_subscriptions ss
			JOIN stripe_prices sp ON ss.stripe_price_id = sp.stripe_id
			JOIN stripe_products spr ON sp.product_id = spr.id
			WHERE ss.customer_id = (SELECT id FROM stripe_customers WHERE stripe_id = $1)
			AND ss.status IN ('active', 'trialing')
			AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
			AND spr.video_approved = true
		)
	`, customerID).Scan(&hasVideoAccess)

	if err != nil {
		return fmt.Errorf("failed to check video access for customer %s: %w", customerID, err)
	}

	if hasVideoAccess {
		// Grant video access by updating user's manual_video_access
		_, err = s.db.Exec(`
			UPDATE users 
			SET manual_video_access = true, updated_at = NOW()
			WHERE id = $1
		`, userID)

		if err != nil {
			return fmt.Errorf("failed to grant video access to user %d: %w", userID, err)
		}

		log.Printf("✅ Video access granted to user %d (%s) via Stripe customer %s",
			userID, userEmail, customerID)
	} else {
		log.Printf("ℹ️ No video-approved subscription found for customer %s (user: %s)",
			customerID, userEmail)
	}

	return nil
}

// linkExistingUserByEmail attempts to link an existing user to a Stripe customer by email
func (s *StripeSyncService) linkExistingUserByEmail(customerID string) (int, error) {
	// First, get the Stripe customer data from our database
	var stripeEmail sql.NullString
	err := s.db.QueryRow(`
		SELECT email 
		FROM stripe_customers 
		WHERE stripe_id = $1
	`, customerID).Scan(&stripeEmail)

	if err != nil {
		return 0, fmt.Errorf("failed to find Stripe customer %s in database: %w", customerID, err)
	}

	if !stripeEmail.Valid || stripeEmail.String == "" {
		return 0, fmt.Errorf("stripe customer %s has no email address", customerID)
	}

	email := stripeEmail.String

	// Try to find an existing user with this email
	var existingUserID int
	err = s.db.QueryRow(`
		SELECT id FROM users 
		WHERE email = $1 AND is_active = true
	`, email).Scan(&existingUserID)

	if err == nil {
		// User exists! Link them to this Stripe customer
		log.Printf("🔗 Linking existing user %d (%s) to Stripe customer %s", existingUserID, email, customerID)

		// Update user to include this Stripe customer ID
		_, err = s.db.Exec(`
			UPDATE users 
			SET stripe_customer_id = $1, updated_at = NOW()
			WHERE id = $2 AND (stripe_customer_id IS NULL OR stripe_customer_id = '')
		`, customerID, existingUserID)

		if err != nil {
			return 0, fmt.Errorf("failed to link user %d to Stripe customer %s: %w", existingUserID, customerID, err)
		}

		return existingUserID, nil
	}

	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("error checking for existing user: %w", err)
	}

	// No existing user found
	log.Printf("ℹ️ No existing user found with email %s for Stripe customer %s", email, customerID)
	return 0, nil // Return 0 to indicate no user was found/linked
}

// syncCustomerToUserTable syncs Stripe customer data to the users table
func (s *StripeSyncService) syncCustomerToUserTable(cust *stripe.Customer) error {
	// Find user by stripe_customer_id or email
	var userID int
	var currentEmail string
	err := s.db.QueryRow(`
		SELECT id, email FROM users 
		WHERE (stripe_customer_id = $1 OR $1 = ANY(COALESCE(stripe_customer_ids, '{}'))) 
		   OR (email = $2 AND is_active = true)
	`, cust.ID, cust.Email).Scan(&userID, &currentEmail)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("ℹ️ No user found for Stripe customer %s (%s) - user may not have registered yet", cust.ID, cust.Email)
			return nil // Not an error - user might register later
		}
		return fmt.Errorf("failed to find user for customer %s: %w", cust.ID, err)
	}

	// Parse name from Stripe customer
	firstName, lastName := parseFullName(cust.Name)

	// Update user with Stripe customer data
	log.Printf("🔄 Syncing Stripe customer %s data to user %d", cust.ID, userID)

	_, err = s.db.Exec(`
		UPDATE users 
		SET 
			stripe_customer_id = $1,
			first_name = COALESCE(NULLIF($2, ''), first_name),
			last_name = COALESCE(NULLIF($3, ''), last_name),
			email = COALESCE(NULLIF($4, ''), email),
			updated_at = NOW()
		WHERE id = $5
	`, cust.ID, firstName, lastName, cust.Email, userID)

	if err != nil {
		return fmt.Errorf("failed to update user %d with Stripe customer data: %w", userID, err)
	}

	log.Printf("✅ Successfully synced Stripe customer %s to user %d", cust.ID, userID)
	return nil
}

// parseFullName splits a full name into first and last name
func parseFullName(fullName string) (string, string) {
	if fullName == "" {
		return "", ""
	}

	parts := strings.Fields(strings.TrimSpace(fullName))
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}

	// First name is the first part, last name is everything else joined
	firstName := parts[0]
	lastName := strings.Join(parts[1:], " ")
	return firstName, lastName
}

func (s *StripeSyncService) upsertSubscription(sub *stripe.Subscription) error {
	// Get customer ID from our database
	var customerID sql.NullInt64
	err := s.db.QueryRow("SELECT id FROM stripe_customers WHERE stripe_id = $1", sub.Customer.ID).Scan(&customerID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// Extract price and product information directly from Stripe API
	var priceID sql.NullInt64
	var stripePriceID sql.NullString
	var unitAmount sql.NullInt64
	var currency sql.NullString
	var stripeProductID sql.NullString
	var productName sql.NullString

	if len(sub.Items.Data) > 0 {
		firstItem := sub.Items.Data[0]
		if firstItem.Price != nil {
			// Store Stripe price ID directly
			stripePriceID = sql.NullString{String: firstItem.Price.ID, Valid: true}
			unitAmount = sql.NullInt64{Int64: firstItem.Price.UnitAmount, Valid: true}
			currency = sql.NullString{String: string(firstItem.Price.Currency), Valid: true}

			// Try to get our local price ID (for backward compatibility)
			s.db.QueryRow("SELECT id FROM stripe_prices WHERE stripe_id = $1", firstItem.Price.ID).Scan(&priceID)

			// Get product information
			if firstItem.Price.Product != nil {
				stripeProductID = sql.NullString{String: firstItem.Price.Product.ID, Valid: true}

				// Get product name - try from Stripe API first, fallback to database
				if firstItem.Price.Product.Name != "" {
					productName = sql.NullString{String: firstItem.Price.Product.Name, Valid: true}
					log.Printf("✅ Got product name from Stripe API: %s", firstItem.Price.Product.Name)
				} else {
					// Fallback: Get product name from our database
					var dbProductName string
					err := s.db.QueryRow("SELECT name FROM stripe_products WHERE stripe_id = $1", firstItem.Price.Product.ID).Scan(&dbProductName)
					if err == nil && dbProductName != "" {
						productName = sql.NullString{String: dbProductName, Valid: true}
						log.Printf("✅ Got product name from database: %s", dbProductName)
					} else {
						log.Printf("⚠️ Product name not available in API or database for product: %s", firstItem.Price.Product.ID)
					}
				}
			}
		}
	}

	query := `
		INSERT INTO stripe_subscriptions (
			stripe_id, customer_id, price_id, status, current_period_start, current_period_end, created_at,
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
		log.Printf("❌ Failed to upsert subscription %s: %v", sub.ID, err)
	} else {
		log.Printf("✅ Upserted subscription %s with product: %s ($%.2f)",
			sub.ID,
			productName.String,
			float64(unitAmount.Int64)/100.0)
	}

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

func (s *StripeSyncService) upsertCoupon(coup *stripe.Coupon) error {
	query := `
		INSERT INTO stripe_coupons (stripe_id, name, percent_off, amount_off, duration, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (stripe_id) 
		DO UPDATE SET 
			name = EXCLUDED.name,
			percent_off = EXCLUDED.percent_off,
			amount_off = EXCLUDED.amount_off,
			duration = EXCLUDED.duration,
			created_at = EXCLUDED.created_at
	`

	var percentOff sql.NullFloat64
	if coup.PercentOff > 0 {
		percentOff = sql.NullFloat64{Float64: float64(coup.PercentOff), Valid: true}
	}

	var amountOff sql.NullInt64
	if coup.AmountOff > 0 {
		amountOff = sql.NullInt64{Int64: coup.AmountOff, Valid: true}
	}

	_, err := s.db.Exec(query,
		coup.ID,
		coup.Name,
		percentOff,
		amountOff,
		string(coup.Duration),
		time.Unix(coup.Created, 0),
	)
	return err
}

// MonthlyMetrics represents calculated monthly metrics
type MonthlyMetrics struct {
	MRR        float64
	ARR        float64
	ChurnRate  float64
	GrowthRate float64
}

func (s *StripeSyncService) calculateMonthlyMetrics(month time.Time) (*MonthlyMetrics, error) {
	// Get the start and end of the month
	startOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	// Calculate MRR from active subscriptions at end of month
	// Simple approach: count active subscriptions and use average revenue estimate
	var activeSubCount int
	var mrr float64

	// Count active subscriptions
	countQuery := `
		SELECT COUNT(*)
		FROM stripe_subscriptions s
		WHERE s.status IN ('active', 'trialing')
		AND s.created_at <= $1
		AND (s.current_period_end IS NULL OR s.current_period_end > $1)
	`

	err := s.db.QueryRow(countQuery, endOfMonth).Scan(&activeSubCount)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to count active subscriptions: %w", err)
	}

	// Calculate MRR using average subscription value
	// This is a simplified calculation - in production you'd want actual pricing data
	averageSubscriptionValue := 50.0 // $50 average monthly subscription
	mrr = float64(activeSubCount) * averageSubscriptionValue

	// Calculate ARR
	arr := mrr * 12

	// Calculate churn rate (simplified - subscriptions that ended this month / total active at start)
	var churnRate float64
	churnQuery := `
		WITH month_start_subs AS (
			SELECT COUNT(*) as start_count
			FROM stripe_subscriptions s
			WHERE s.status IN ('active', 'trialing')
			AND s.created_at < $1
		),
		churned_subs AS (
			SELECT COUNT(*) as churned_count
			FROM stripe_subscriptions s
			WHERE s.status = 'canceled'
			AND s.current_period_end >= $1 
			AND s.current_period_end < $2
		)
		SELECT 
			CASE 
				WHEN mss.start_count > 0 THEN (cs.churned_count::float / mss.start_count::float) * 100
				ELSE 0 
			END
		FROM month_start_subs mss, churned_subs cs
	`

	err = s.db.QueryRow(churnQuery, startOfMonth, endOfMonth).Scan(&churnRate)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to calculate churn rate: %w", err)
	}

	// Calculate growth rate (MRR growth compared to previous month)
	var growthRate float64
	prevMonthEnd := startOfMonth.Add(-time.Second)

	// Count active subscriptions in previous month
	var prevActiveSubCount int
	err = s.db.QueryRow(countQuery, prevMonthEnd).Scan(&prevActiveSubCount)
	if err == nil && prevActiveSubCount > 0 {
		prevMRR := float64(prevActiveSubCount) * averageSubscriptionValue
		if prevMRR > 0 {
			growthRate = ((mrr - prevMRR) / prevMRR) * 100
		}
	}

	return &MonthlyMetrics{
		MRR:        mrr,
		ARR:        arr,
		ChurnRate:  churnRate,
		GrowthRate: growthRate,
	}, nil
}

func (s *StripeSyncService) upsertMonthlyMetrics(yearMonth string, metrics *MonthlyMetrics) error {
	query := `
		INSERT INTO stripe_monthly_metrics (year_month, mrr, arr, churn_rate, growth_rate)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (year_month) 
		DO UPDATE SET 
			mrr = EXCLUDED.mrr,
			arr = EXCLUDED.arr,
			churn_rate = EXCLUDED.churn_rate,
			growth_rate = EXCLUDED.growth_rate
	`

	_, err := s.db.Exec(query,
		yearMonth,
		int(metrics.MRR*100), // Store as cents
		int(metrics.ARR*100), // Store as cents
		metrics.ChurnRate,
		metrics.GrowthRate,
	)
	return err
}

// Sync job management functions are now in the System Management section below

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

	// First, sync to stripe_customers table
	err := s.upsertCustomer(cust)
	if err != nil {
		return err
	}

	// Then, sync to users table if a matching user exists
	err = s.syncCustomerToUserTable(cust)
	if err != nil {
		// Don't fail the webhook if user sync fails - log and continue
		log.Printf("⚠️ Failed to sync customer %s to users table: %v", cust.ID, err)
	}

	return nil
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

		s.updateSyncJobProgress(jobID, i+1, "")
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

// ===== SYNC JOB MANAGEMENT =====

// createSyncJob creates a new sync job and returns its ID
func (s *StripeSyncService) createSyncJob(jobType, entityType string, totalItems int) (int, error) {
	query := `
		INSERT INTO stripe_sync_jobs (job_type, entity_type, status, total_items, started_at)
		VALUES ($1, $2, 'running', $3, NOW())
		RETURNING id
	`

	var jobID int
	err := s.db.QueryRow(query, jobType, entityType, totalItems).Scan(&jobID)
	if err != nil {
		return 0, fmt.Errorf("failed to create sync job: %w", err)
	}

	log.Printf("📋 Created sync job #%d: %s/%s (expected: %d items)", jobID, jobType, entityType, totalItems)
	return jobID, nil
}

// completeSyncJob marks a sync job as completed or failed
func (s *StripeSyncService) completeSyncJob(jobID int, syncErr error) error {
	var status, errorMessage string

	if syncErr != nil {
		status = "failed"
		errorMessage = syncErr.Error()
		log.Printf("❌ Sync job #%d failed: %v", jobID, syncErr)
	} else {
		status = "completed"
		log.Printf("✅ Sync job #%d completed successfully", jobID)
	}

	query := `
		UPDATE stripe_sync_jobs 
		SET status = $1, error_message = $2, completed_at = NOW(), updated_at = NOW()
		WHERE id = $3
	`

	_, err := s.db.Exec(query, status, errorMessage, jobID)
	if err != nil {
		log.Printf("⚠️ Failed to update sync job #%d status: %v", jobID, err)
		return err
	}

	return nil
}

// updateSyncJobProgress updates the progress of a sync job
func (s *StripeSyncService) updateSyncJobProgress(jobID int, processedItems int, cursor string) error {
	query := `
		UPDATE stripe_sync_jobs 
		SET processed_items = $1, last_cursor = $2, updated_at = NOW()
		WHERE id = $3
	`

	_, err := s.db.Exec(query, processedItems, cursor, jobID)
	return err
}

// ===== STRIPE ENTITIES TRACKING =====

// trackStripeEntity records or updates a Stripe entity in the universal tracking table
func (s *StripeSyncService) trackStripeEntity(entityType, entityID string, localID int, metadata map[string]interface{}, stripeCreated, stripeUpdated time.Time) error {
	metadataJSON, _ := json.Marshal(metadata)

	query := `
		INSERT INTO stripe_entities (entity_type, entity_id, local_id, metadata, stripe_created_at, stripe_updated_at, last_synced_at, sync_status)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), 'synced')
		ON CONFLICT (entity_type, entity_id) 
		DO UPDATE SET 
			local_id = EXCLUDED.local_id,
			metadata = EXCLUDED.metadata,
			stripe_updated_at = EXCLUDED.stripe_updated_at,
			last_synced_at = NOW(),
			sync_status = 'synced',
			updated_at = NOW()
	`

	_, err := s.db.Exec(query, entityType, entityID, localID, metadataJSON, stripeCreated, stripeUpdated)
	return err
}

// markEntitySyncError marks an entity as having a sync error
func (s *StripeSyncService) markEntitySyncError(entityType, entityID string, syncError error) error {
	query := `
		INSERT INTO stripe_entities (entity_type, entity_id, sync_status, last_synced_at)
		VALUES ($1, $2, 'error', NOW())
		ON CONFLICT (entity_type, entity_id) 
		DO UPDATE SET 
			sync_status = 'error',
			last_synced_at = NOW(),
			updated_at = NOW()
	`

	_, err := s.db.Exec(query, entityType, entityID)
	return err
}

// ===== SYNC CONFIGURATION MANAGEMENT =====

// createOrUpdateSyncConfig creates or updates sync configuration for an entity type
func (s *StripeSyncService) createOrUpdateSyncConfig(entityType string, config *SyncConfig) error {
	configDataJSON, _ := json.Marshal(config.ConfigData)

	query := `
		INSERT INTO stripe_sync_config (entity_type, sync_enabled, sync_interval_hours, batch_size, config_data)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (entity_type) 
		DO UPDATE SET 
			sync_enabled = EXCLUDED.sync_enabled,
			sync_interval_hours = EXCLUDED.sync_interval_hours,
			batch_size = EXCLUDED.batch_size,
			config_data = EXCLUDED.config_data,
			updated_at = NOW()
	`

	_, err := s.db.Exec(query, entityType, config.SyncEnabled, config.SyncIntervalHours, config.BatchSize, configDataJSON)
	return err
}

// updateLastSyncTime updates the last sync time for an entity type
func (s *StripeSyncService) updateLastSyncTime(entityType string, isFullSync bool) error {
	var query string

	if isFullSync {
		query = `
			UPDATE stripe_sync_config 
			SET last_full_sync = NOW(), updated_at = NOW()
			WHERE entity_type = $1
		`
	} else {
		query = `
			UPDATE stripe_sync_config 
			SET last_incremental_sync = NOW(), updated_at = NOW()
			WHERE entity_type = $1
		`
	}

	_, err := s.db.Exec(query, entityType)
	return err
}

// ===== SYSTEM MANAGEMENT API METHODS =====

// GetSyncJobs retrieves sync jobs with optional filtering
func (s *StripeSyncService) GetSyncJobs(status, entityType string, limit int) ([]SyncJob, error) {
	query := `
		SELECT id, job_type, entity_type, status, total_items, processed_items, 
		       last_cursor, error_message, started_at, completed_at, created_at, updated_at
		FROM stripe_sync_jobs
		WHERE ($1 = '' OR status = $1)
		AND ($2 = '' OR entity_type = $2)
		ORDER BY created_at DESC
		LIMIT $3
	`

	rows, err := s.db.Query(query, status, entityType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []SyncJob
	for rows.Next() {
		var job SyncJob
		err := rows.Scan(
			&job.ID, &job.JobType, &job.EntityType, &job.Status,
			&job.TotalItems, &job.ProcessedItems, &job.LastCursor,
			&job.ErrorMessage, &job.StartedAt, &job.CompletedAt,
			&job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			continue
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// GetSyncJobByID retrieves a specific sync job
func (s *StripeSyncService) GetSyncJobByID(jobID int) (*SyncJob, error) {
	query := `
		SELECT id, job_type, entity_type, status, total_items, processed_items, 
		       last_cursor, error_message, started_at, completed_at, created_at, updated_at
		FROM stripe_sync_jobs
		WHERE id = $1
	`

	var job SyncJob
	err := s.db.QueryRow(query, jobID).Scan(
		&job.ID, &job.JobType, &job.EntityType, &job.Status,
		&job.TotalItems, &job.ProcessedItems, &job.LastCursor,
		&job.ErrorMessage, &job.StartedAt, &job.CompletedAt,
		&job.CreatedAt, &job.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &job, nil
}

// CancelSyncJob cancels a running sync job
func (s *StripeSyncService) CancelSyncJob(jobID int) error {
	query := `
		UPDATE stripe_sync_jobs 
		SET status = 'cancelled', completed_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND status = 'running'
	`

	_, err := s.db.Exec(query, jobID)
	return err
}

// GetAllSyncConfigs retrieves all sync configurations
func (s *StripeSyncService) GetAllSyncConfigs() ([]SyncConfig, error) {
	query := `
		SELECT entity_type, sync_enabled, sync_interval_hours, batch_size, 
		       last_full_sync, last_incremental_sync, config_data, created_at, updated_at
		FROM stripe_sync_config
		ORDER BY entity_type
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []SyncConfig
	for rows.Next() {
		var config SyncConfig
		var configDataJSON []byte

		err := rows.Scan(
			&config.EntityType, &config.SyncEnabled, &config.SyncIntervalHours,
			&config.BatchSize, &config.LastFullSync, &config.LastIncrementalSync,
			&configDataJSON, &config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if configDataJSON != nil {
			json.Unmarshal(configDataJSON, &config.ConfigData)
		}

		configs = append(configs, config)
	}

	return configs, nil
}

// GetSyncConfigByType retrieves sync configuration for a specific entity type
func (s *StripeSyncService) GetSyncConfigByType(entityType string) (*SyncConfig, error) {
	return s.getSyncConfig(entityType)
}

// UpdateSyncConfig updates sync configuration for an entity type
func (s *StripeSyncService) UpdateSyncConfig(entityType string, syncEnabled *bool, syncIntervalHours *int, batchSize *int, configData map[string]interface{}) error {
	// Get current config
	currentConfig, err := s.getSyncConfig(entityType)
	if err != nil {
		return err
	}

	// Update only provided fields
	if syncEnabled != nil {
		currentConfig.SyncEnabled = *syncEnabled
	}
	if syncIntervalHours != nil {
		currentConfig.SyncIntervalHours = *syncIntervalHours
	}
	if batchSize != nil {
		currentConfig.BatchSize = *batchSize
	}
	if configData != nil {
		currentConfig.ConfigData = configData
	}

	return s.createOrUpdateSyncConfig(entityType, currentConfig)
}

// StripeEntity represents an entity in the universal tracking table
type StripeEntity struct {
	ID              int                    `json:"id"`
	EntityType      string                 `json:"entity_type"`
	EntityID        string                 `json:"entity_id"`
	LocalID         int                    `json:"local_id"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       *time.Time             `json:"created_at"`
	UpdatedAt       *time.Time             `json:"updated_at"`
	LastSyncedAt    *time.Time             `json:"last_synced_at"`
	StripeCreatedAt *time.Time             `json:"stripe_created_at"`
	StripeUpdatedAt *time.Time             `json:"stripe_updated_at"`
	SyncStatus      string                 `json:"sync_status"`
}

// GetStripeEntities retrieves entities with pagination
func (s *StripeSyncService) GetStripeEntities(limit, offset int) ([]StripeEntity, int, error) {
	// Get total count
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM stripe_entities").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get entities
	query := `
		SELECT id, entity_type, entity_id, local_id, metadata, created_at, updated_at,
		       last_synced_at, stripe_created_at, stripe_updated_at, sync_status
		FROM stripe_entities
		ORDER BY last_synced_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entities []StripeEntity
	for rows.Next() {
		var entity StripeEntity
		var metadataJSON []byte

		err := rows.Scan(
			&entity.ID, &entity.EntityType, &entity.EntityID, &entity.LocalID,
			&metadataJSON, &entity.CreatedAt, &entity.UpdatedAt,
			&entity.LastSyncedAt, &entity.StripeCreatedAt, &entity.StripeUpdatedAt,
			&entity.SyncStatus,
		)
		if err != nil {
			continue
		}

		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &entity.Metadata)
		}

		entities = append(entities, entity)
	}

	return entities, total, nil
}

// GetStripeEntitiesByType retrieves entities by type with pagination
func (s *StripeSyncService) GetStripeEntitiesByType(entityType string, limit, offset int) ([]StripeEntity, int, error) {
	// Get total count
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM stripe_entities WHERE entity_type = $1", entityType).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get entities
	query := `
		SELECT id, entity_type, entity_id, local_id, metadata, created_at, updated_at,
		       last_synced_at, stripe_created_at, stripe_updated_at, sync_status
		FROM stripe_entities
		WHERE entity_type = $1
		ORDER BY last_synced_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.Query(query, entityType, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entities []StripeEntity
	for rows.Next() {
		var entity StripeEntity
		var metadataJSON []byte

		err := rows.Scan(
			&entity.ID, &entity.EntityType, &entity.EntityID, &entity.LocalID,
			&metadataJSON, &entity.CreatedAt, &entity.UpdatedAt,
			&entity.LastSyncedAt, &entity.StripeCreatedAt, &entity.StripeUpdatedAt,
			&entity.SyncStatus,
		)
		if err != nil {
			continue
		}

		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &entity.Metadata)
		}

		entities = append(entities, entity)
	}

	return entities, total, nil
}

// GetStripeEntitiesByStatus retrieves entities by sync status with pagination
func (s *StripeSyncService) GetStripeEntitiesByStatus(status string, limit, offset int) ([]StripeEntity, int, error) {
	// Get total count
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM stripe_entities WHERE sync_status = $1", status).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get entities
	query := `
		SELECT id, entity_type, entity_id, local_id, metadata, created_at, updated_at,
		       last_synced_at, stripe_created_at, stripe_updated_at, sync_status
		FROM stripe_entities
		WHERE sync_status = $1
		ORDER BY last_synced_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.Query(query, status, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entities []StripeEntity
	for rows.Next() {
		var entity StripeEntity
		var metadataJSON []byte

		err := rows.Scan(
			&entity.ID, &entity.EntityType, &entity.EntityID, &entity.LocalID,
			&metadataJSON, &entity.CreatedAt, &entity.UpdatedAt,
			&entity.LastSyncedAt, &entity.StripeCreatedAt, &entity.StripeUpdatedAt,
			&entity.SyncStatus,
		)
		if err != nil {
			continue
		}

		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &entity.Metadata)
		}

		entities = append(entities, entity)
	}

	return entities, total, nil
}

// SystemHealth represents system health status
type SystemHealth struct {
	Status          string                 `json:"status"`
	LastSyncTime    *time.Time             `json:"last_sync_time"`
	ActiveJobs      int                    `json:"active_jobs"`
	FailedJobs      int                    `json:"failed_jobs"`
	ErrorEntities   int                    `json:"error_entities"`
	TotalEntities   int                    `json:"total_entities"`
	DatabaseStatus  string                 `json:"database_status"`
	StripeAPIStatus string                 `json:"stripe_api_status"`
	Details         map[string]interface{} `json:"details"`
}

// GetSystemHealth returns overall system health
func (s *StripeSyncService) GetSystemHealth() (*SystemHealth, error) {
	health := &SystemHealth{
		Details: make(map[string]interface{}),
	}

	// Check database connectivity
	err := s.db.Ping()
	if err != nil {
		health.DatabaseStatus = "error"
		health.Status = "unhealthy"
		health.Details["database_error"] = err.Error()
	} else {
		health.DatabaseStatus = "healthy"
	}

	// Get active jobs count
	err = s.db.QueryRow("SELECT COUNT(*) FROM stripe_sync_jobs WHERE status = 'running'").Scan(&health.ActiveJobs)
	if err != nil {
		health.ActiveJobs = -1
	}

	// Get failed jobs count (last 24 hours)
	err = s.db.QueryRow("SELECT COUNT(*) FROM stripe_sync_jobs WHERE status = 'failed' AND created_at > NOW() - INTERVAL '24 hours'").Scan(&health.FailedJobs)
	if err != nil {
		health.FailedJobs = -1
	}

	// Get error entities count
	err = s.db.QueryRow("SELECT COUNT(*) FROM stripe_entities WHERE sync_status = 'error'").Scan(&health.ErrorEntities)
	if err != nil {
		health.ErrorEntities = -1
	}

	// Get total entities count
	err = s.db.QueryRow("SELECT COUNT(*) FROM stripe_entities").Scan(&health.TotalEntities)
	if err != nil {
		health.TotalEntities = -1
	}

	// Get last sync time
	err = s.db.QueryRow("SELECT MAX(completed_at) FROM stripe_sync_jobs WHERE status = 'completed'").Scan(&health.LastSyncTime)
	if err != nil {
		// No completed syncs yet
	}

	// Check Stripe API status (simple check)
	if s.stripeService.IsEnabled() {
		health.StripeAPIStatus = "enabled"
	} else {
		health.StripeAPIStatus = "disabled"
	}

	// Determine overall status
	if health.DatabaseStatus == "healthy" && health.FailedJobs < 5 && health.ErrorEntities < 100 {
		health.Status = "healthy"
	} else if health.DatabaseStatus == "healthy" {
		health.Status = "degraded"
	} else {
		health.Status = "unhealthy"
	}

	return health, nil
}

// SystemStats represents system statistics
type SystemStats struct {
	TotalSyncJobs    int                   `json:"total_sync_jobs"`
	CompletedJobs    int                   `json:"completed_jobs"`
	FailedJobs       int                   `json:"failed_jobs"`
	RunningJobs      int                   `json:"running_jobs"`
	EntitiesByType   map[string]int        `json:"entities_by_type"`
	EntitiesByStatus map[string]int        `json:"entities_by_status"`
	ConfiguredTypes  []string              `json:"configured_types"`
	LastSyncTimes    map[string]*time.Time `json:"last_sync_times"`
	SyncFrequency    map[string]int        `json:"sync_frequency"`
}

// GetSystemStats returns system statistics
func (s *StripeSyncService) GetSystemStats() (*SystemStats, error) {
	stats := &SystemStats{
		EntitiesByType:   make(map[string]int),
		EntitiesByStatus: make(map[string]int),
		LastSyncTimes:    make(map[string]*time.Time),
		SyncFrequency:    make(map[string]int),
	}

	// Get job statistics
	jobStatsQuery := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed,
			COUNT(CASE WHEN status = 'running' THEN 1 END) as running
		FROM stripe_sync_jobs
	`

	err := s.db.QueryRow(jobStatsQuery).Scan(&stats.TotalSyncJobs, &stats.CompletedJobs, &stats.FailedJobs, &stats.RunningJobs)
	if err != nil {
		return nil, err
	}

	// Get entities by type
	entitiesByTypeQuery := `
		SELECT entity_type, COUNT(*) 
		FROM stripe_entities 
		GROUP BY entity_type
	`

	rows, err := s.db.Query(entitiesByTypeQuery)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var entityType string
			var count int
			if rows.Scan(&entityType, &count) == nil {
				stats.EntitiesByType[entityType] = count
			}
		}
	}

	// Get entities by status
	entitiesByStatusQuery := `
		SELECT sync_status, COUNT(*) 
		FROM stripe_entities 
		GROUP BY sync_status
	`

	rows, err = s.db.Query(entitiesByStatusQuery)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var status string
			var count int
			if rows.Scan(&status, &count) == nil {
				stats.EntitiesByStatus[status] = count
			}
		}
	}

	// Get configured entity types and their last sync times
	configQuery := `
		SELECT entity_type, last_full_sync, sync_interval_hours
		FROM stripe_sync_config
		WHERE sync_enabled = true
	`

	rows, err = s.db.Query(configQuery)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var entityType string
			var lastSync *time.Time
			var intervalHours int
			if rows.Scan(&entityType, &lastSync, &intervalHours) == nil {
				stats.ConfiguredTypes = append(stats.ConfiguredTypes, entityType)
				stats.LastSyncTimes[entityType] = lastSync
				stats.SyncFrequency[entityType] = intervalHours
			}
		}
	}

	return stats, nil
}
