package stripe

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"bome-backend/infrastructure/database"
	subServices "bome-backend/subscription/services"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/balance"
)

// StripeAnalyticsService provides Stripe analytics and metrics
type StripeAnalyticsService struct {
	db            *database.DB
	stripeService *subServices.StripeService
}

// DashboardMetrics represents comprehensive dashboard data
type DashboardMetrics struct {
	Balance          *BalanceInfo          `json:"balance"`
	Customers        *CustomerMetrics      `json:"customers"`
	Subscriptions    *SubscriptionMetrics  `json:"subscriptions"`
	Revenue          *RevenueMetrics       `json:"revenue"`
	DatabaseMetrics  *DatabaseMetrics      `json:"database_metrics"`
	LastUpdated      time.Time             `json:"last_updated"`
}

// BalanceInfo represents Stripe account balance
type BalanceInfo struct {
	Available        []BalanceAmount `json:"available"`
	Pending          []BalanceAmount `json:"pending"`
	InstantAvailable []BalanceAmount `json:"instant_available"`
	TotalUSD         float64         `json:"total_usd"`
}

// BalanceAmount represents a balance in a specific currency
type BalanceAmount struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	USD      float64 `json:"usd"` // Converted amount
}

// CustomerMetrics represents customer statistics
type CustomerMetrics struct {
	TotalCustomers      int     `json:"total_customers"`
	ActiveSubscribers   int     `json:"active_subscribers"`
	TrialingSubscribers int     `json:"trialing_subscribers"`
	ChurnRate           float64 `json:"churn_rate"`
	GrowthRate          float64 `json:"growth_rate"`
}

// SubscriptionMetrics represents subscription statistics
type SubscriptionMetrics struct {
	TotalSubscriptions  int                `json:"total_subscriptions"`
	ActiveSubscriptions int                `json:"active_subscriptions"`
	PausedSubscriptions int                `json:"paused_subscriptions"`
	CanceledToday       int                `json:"canceled_today"`
	ByStatus            map[string]int     `json:"by_status"`
}

// RevenueMetrics represents revenue statistics
type RevenueMetrics struct {
	MRR              float64 `json:"mrr"`              // Monthly Recurring Revenue
	ARR              float64 `json:"arr"`              // Annual Recurring Revenue
	TotalRevenue     float64 `json:"total_revenue"`     // All-time
	RevenueToday     float64 `json:"revenue_today"`
	RevenueThisMonth float64 `json:"revenue_this_month"`
	AverageRevenue   float64 `json:"average_revenue"`   // Per customer
}

// DatabaseMetrics represents local database statistics  
type DatabaseMetrics struct {
	CustomersInDB     int       `json:"customers_in_db"`
	SubscriptionsInDB int       `json:"subscriptions_in_db"`
	ProductsInDB      int       `json:"products_in_db"`
	InvoicesInDB      int       `json:"invoices_in_db"`
	LastSyncAt        time.Time `json:"last_sync_at"`
}

// NewStripeAnalyticsService creates a new analytics service
func NewStripeAnalyticsService(db *database.DB, stripeService *subServices.StripeService) *StripeAnalyticsService {
	return &StripeAnalyticsService{
		db:            db,
		stripeService: stripeService,
	}
}

// GetDashboardMetrics returns comprehensive dashboard data
func (s *StripeAnalyticsService) GetDashboardMetrics() (*DashboardMetrics, error) {
	log.Println("📊 Fetching dashboard metrics...")

	metrics := &DashboardMetrics{
		LastUpdated: time.Now(),
	}

	// Fetch all metrics in parallel
	var (
		balanceErr       error
		customersErr     error
		subscriptionsErr error
		revenueErr       error
		dbStatsErr       error
	)

	// Use goroutines for parallel fetching
	done := make(chan bool)

	go func() {
		metrics.Balance, balanceErr = s.GetBalance()
		done <- true
	}()

	go func() {
		metrics.Customers, customersErr = s.GetCustomerMetrics()
		done <- true
	}()

	go func() {
		metrics.Subscriptions, subscriptionsErr = s.GetSubscriptionMetrics()
		done <- true
	}()

	go func() {
		metrics.Revenue, revenueErr = s.GetRevenueMetrics()
		done <- true
	}()

	go func() {
		metrics.DatabaseMetrics, dbStatsErr = s.GetDatabaseMetrics()
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Log any errors but don't fail the entire request
	if balanceErr != nil {
		log.Printf("⚠️  Balance fetch error: %v", balanceErr)
	}
	if customersErr != nil {
		log.Printf("⚠️  Customers fetch error: %v", customersErr)
	}
	if subscriptionsErr != nil {
		log.Printf("⚠️  Subscriptions fetch error: %v", subscriptionsErr)
	}
	if revenueErr != nil {
		log.Printf("⚠️  Revenue fetch error: %v", revenueErr)
	}
	if dbStatsErr != nil {
		log.Printf("⚠️  Database stats fetch error: %v", dbStatsErr)
	}

	log.Println("✅ Dashboard metrics fetched successfully")
	return metrics, nil
}

// GetBalance returns Stripe account balance
func (s *StripeAnalyticsService) GetBalance() (*BalanceInfo, error) {
	if !s.stripeService.IsEnabled() {
		return nil, fmt.Errorf("stripe service not enabled")
	}

	// Use context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	bal, err := balance.Get(&stripe.BalanceParams{
		Params: stripe.Params{Context: ctx},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	info := &BalanceInfo{
		Available:        make([]BalanceAmount, 0),
		Pending:          make([]BalanceAmount, 0),
		InstantAvailable: make([]BalanceAmount, 0),
	}

	// Convert available balances
	for _, a := range bal.Available {
		amount := BalanceAmount{
			Amount:   a.Amount,
			Currency: string(a.Currency),
			USD:      float64(a.Amount) / 100,
		}
		info.Available = append(info.Available, amount)
		if string(a.Currency) == "usd" {
			info.TotalUSD += amount.USD
		}
	}

	// Convert pending balances
	for _, p := range bal.Pending {
		info.Pending = append(info.Pending, BalanceAmount{
			Amount:   p.Amount,
			Currency: string(p.Currency),
			USD:      float64(p.Amount) / 100,
		})
	}

	// Instant available
	for _, ia := range bal.InstantAvailable {
		info.InstantAvailable = append(info.InstantAvailable, BalanceAmount{
			Amount:   ia.Amount,
			Currency: string(ia.Currency),
			USD:      float64(ia.Amount) / 100,
		})
	}

	return info, nil
}

// GetCustomerMetrics returns customer statistics
func (s *StripeAnalyticsService) GetCustomerMetrics() (*CustomerMetrics, error) {
	metrics := &CustomerMetrics{}

	// Get total customers from database
	err := s.db.DB.QueryRow("SELECT COUNT(*) FROM stripe_customers").Scan(&metrics.TotalCustomers)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to count customers: %w", err)
	}

	// Get active subscribers
	err = s.db.DB.QueryRow(`
		SELECT COUNT(DISTINCT customer_id) 
		FROM stripe_subscriptions 
		WHERE status IN ('active', 'trialing')
	`).Scan(&metrics.ActiveSubscribers)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Warning: Failed to count active subscribers: %v", err)
	}

	// Get trialing subscribers
	err = s.db.DB.QueryRow(`
		SELECT COUNT(DISTINCT customer_id) 
		FROM stripe_subscriptions 
		WHERE status = 'trialing'
	`).Scan(&metrics.TrialingSubscribers)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Warning: Failed to count trialing subscribers: %v", err)
	}

	// Calculate growth rate (simple month-over-month)
	var currentMonthCount, lastMonthCount int
	currentMonth := time.Now().Format("2006-01")
	lastMonth := time.Now().AddDate(0, -1, 0).Format("2006-01")

	s.db.DB.QueryRow("SELECT COUNT(*) FROM stripe_customers WHERE created_at >= $1", currentMonth+"-01").Scan(&currentMonthCount)
	s.db.DB.QueryRow("SELECT COUNT(*) FROM stripe_customers WHERE created_at >= $1 AND created_at < $2", lastMonth+"-01", currentMonth+"-01").Scan(&lastMonthCount)

	if lastMonthCount > 0 {
		metrics.GrowthRate = float64(currentMonthCount-lastMonthCount) / float64(lastMonthCount) * 100
	}

	return metrics, nil
}

// GetSubscriptionMetrics returns subscription statistics
func (s *StripeAnalyticsService) GetSubscriptionMetrics() (*SubscriptionMetrics, error) {
	metrics := &SubscriptionMetrics{
		ByStatus: make(map[string]int),
	}

	// Get total subscriptions
	err := s.db.DB.QueryRow("SELECT COUNT(*) FROM stripe_subscriptions").Scan(&metrics.TotalSubscriptions)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to count subscriptions: %w", err)
	}

	// Get active subscriptions
	err = s.db.DB.QueryRow("SELECT COUNT(*) FROM stripe_subscriptions WHERE status = 'active'").Scan(&metrics.ActiveSubscriptions)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Warning: Failed to count active subscriptions: %v", err)
	}

	// Get paused subscriptions
	err = s.db.DB.QueryRow("SELECT COUNT(*) FROM stripe_subscriptions WHERE status = 'paused'").Scan(&metrics.PausedSubscriptions)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Warning: Failed to count paused subscriptions: %v", err)
	}

	// Get canceled today
	today := time.Now().Format("2006-01-02")
	err = s.db.DB.QueryRow("SELECT COUNT(*) FROM stripe_subscriptions WHERE status = 'canceled' AND updated_at >= $1", today).Scan(&metrics.CanceledToday)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Warning: Failed to count canceled today: %v", err)
	}

	// Get subscription counts by status
	rows, err := s.db.DB.Query("SELECT status, COUNT(*) FROM stripe_subscriptions GROUP BY status")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var status string
			var count int
			if err := rows.Scan(&status, &count); err == nil {
				metrics.ByStatus[status] = count
			}
		}
	}

	return metrics, nil
}

// GetRevenueMetrics returns revenue statistics
func (s *StripeAnalyticsService) GetRevenueMetrics() (*RevenueMetrics, error) {
	metrics := &RevenueMetrics{}

	// Calculate MRR from active subscriptions
	var totalMRR sql.NullInt64
	err := s.db.DB.QueryRow(`
		SELECT SUM(unit_amount) 
		FROM stripe_subscriptions 
		WHERE status IN ('active', 'trialing')
	`).Scan(&totalMRR)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Warning: Failed to calculate MRR: %v", err)
	}

	if totalMRR.Valid {
		metrics.MRR = float64(totalMRR.Int64) / 100 // Convert cents to dollars
		metrics.ARR = metrics.MRR * 12
	}

	// Get today's revenue from stripe_daily_revenue table (if exists)
	today := time.Now().Format("2006-01-02")
	var revenueToday sql.NullInt64
	err = s.db.DB.QueryRow("SELECT total_revenue FROM stripe_daily_revenue WHERE date = $1", today).Scan(&revenueToday)
	if err == nil && revenueToday.Valid {
		metrics.RevenueToday = float64(revenueToday.Int64) / 100
	}

	// Get this month's revenue
	currentMonth := time.Now().Format("2006-01")
	var revenueThisMonth sql.NullInt64
	err = s.db.DB.QueryRow(`
		SELECT SUM(total_revenue) 
		FROM stripe_daily_revenue 
		WHERE date >= $1
	`, currentMonth+"-01").Scan(&revenueThisMonth)
	if err == nil && revenueThisMonth.Valid {
		metrics.RevenueThisMonth = float64(revenueThisMonth.Int64) / 100
	}

	// Calculate average revenue per customer
	if metrics.MRR > 0 {
		var customerCount int
		s.db.DB.QueryRow("SELECT COUNT(DISTINCT customer_id) FROM stripe_subscriptions WHERE status IN ('active', 'trialing')").Scan(&customerCount)
		if customerCount > 0 {
			metrics.AverageRevenue = metrics.MRR / float64(customerCount)
		}
	}

	return metrics, nil
}

// GetDatabaseMetrics returns local database statistics
func (s *StripeAnalyticsService) GetDatabaseMetrics() (*DatabaseMetrics, error) {
	stats := &DatabaseMetrics{}

	// Count customers
	s.db.DB.QueryRow("SELECT COUNT(*) FROM stripe_customers").Scan(&stats.CustomersInDB)

	// Count subscriptions
	s.db.DB.QueryRow("SELECT COUNT(*) FROM stripe_subscriptions").Scan(&stats.SubscriptionsInDB)

	// Count products
	s.db.DB.QueryRow("SELECT COUNT(*) FROM stripe_products").Scan(&stats.ProductsInDB)

	// Count invoices
	s.db.DB.QueryRow("SELECT COUNT(*) FROM stripe_invoices").Scan(&stats.InvoicesInDB)

	// Get last sync time (from most recently updated customer)
	var lastSync sql.NullTime
	s.db.DB.QueryRow("SELECT MAX(updated_at) FROM stripe_customers").Scan(&lastSync)
	if lastSync.Valid {
		stats.LastSyncAt = lastSync.Time
	}

	return stats, nil
}

// CheckHealth performs a quick health check on Stripe
func (s *StripeAnalyticsService) CheckHealth() (bool, error) {
	if !s.stripeService.IsEnabled() {
		return false, fmt.Errorf("stripe service not enabled")
	}

	// Try to get balance with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := balance.Get(&stripe.BalanceParams{
		Params: stripe.Params{Context: ctx},
	})

	return err == nil, err
}

