package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"bome-backend/creator/models"
	"bome-backend/infrastructure/database"
)

// PayoutService handles business logic for payout calculation and management
type PayoutService struct {
	db *database.DB
}

// NewPayoutService creates a new PayoutService
func NewPayoutService(db *database.DB) *PayoutService {
	return &PayoutService{
		db: db,
	}
}

// GenerateMonthlyPayouts generates payouts for all presenters for a specific month
func (s *PayoutService) GenerateMonthlyPayouts(month time.Time, calculatedBy int) (map[string]interface{}, error) {
	// Ensure month is the first day of the month
	month = time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)

	log.Printf("[PAYOUT-SERVICE] Generating monthly payouts for: %s", month.Format("2006-01-02"))

	// Call the database function
	query := `SELECT * FROM generate_monthly_payouts($1, $2)`

	var generatedCount int
	var totalAmount float64
	var detailsJSON []byte

	err := s.db.QueryRow(query, month, calculatedBy).Scan(&generatedCount, &totalAmount, &detailsJSON)
	if err != nil {
		log.Printf("[PAYOUT-SERVICE] Error generating monthly payouts: %v", err)
		return nil, fmt.Errorf("failed to generate monthly payouts: %w", err)
	}

	// Parse the JSON details
	var details map[string]interface{}
	err = json.Unmarshal(detailsJSON, &details)
	if err != nil {
		log.Printf("[PAYOUT-SERVICE] Error parsing payout details: %v", err)
		return nil, fmt.Errorf("failed to parse payout details: %w", err)
	}

	log.Printf("[PAYOUT-SERVICE] Generated %d payouts, total amount: $%.2f", generatedCount, totalAmount)

	return map[string]interface{}{
		"generated_count": generatedCount,
		"total_amount":    totalAmount,
		"details":         details,
	}, nil
}

// CalculatePresenterPayout calculates payout for a specific presenter for a month
func (s *PayoutService) CalculatePresenterPayout(presenterID int, month time.Time, formulaID int) (map[string]interface{}, error) {
	// Ensure month is the first day of the month
	month = time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)

	log.Printf("[PAYOUT-SERVICE] Calculating payout for presenter %d for %s using formula %d", presenterID, month.Format("2006-01-02"), formulaID)

	// Call the database function
	query := `SELECT * FROM calculate_presenter_payout($1, $2, $3)`

	var presenterName string
	var totalVideos int
	var totalViews int64
	var totalWatchMinutes int64
	var baseAmount, bonusAmount, finalAmount float64
	var calculationJSON []byte

	err := s.db.QueryRow(query, presenterID, month, formulaID).Scan(
		&presenterID, &presenterName, &totalVideos, &totalViews, &totalWatchMinutes,
		&baseAmount, &bonusAmount, &finalAmount, &calculationJSON,
	)

	if err != nil {
		log.Printf("[PAYOUT-SERVICE] Error calculating payout: %v", err)
		return nil, fmt.Errorf("failed to calculate payout: %w", err)
	}

	// Parse the calculation JSON
	var calculationDetails map[string]interface{}
	err = json.Unmarshal(calculationJSON, &calculationDetails)
	if err != nil {
		log.Printf("[PAYOUT-SERVICE] Error parsing calculation details: %v", err)
		return nil, fmt.Errorf("failed to parse calculation details: %w", err)
	}

	log.Printf("[PAYOUT-SERVICE] Calculated payout: $%.2f for presenter %d", finalAmount, presenterID)

	return map[string]interface{}{
		"presenter_id":        presenterID,
		"presenter_name":      presenterName,
		"total_videos":        totalVideos,
		"total_views":         totalViews,
		"total_watch_minutes": totalWatchMinutes,
		"base_amount":         baseAmount,
		"bonus_amount":        bonusAmount,
		"final_amount":        finalAmount,
		"calculation_details": calculationDetails,
	}, nil
}

// GetPayoutByID retrieves a payout by ID
func (s *PayoutService) GetPayoutByID(id int) (*models.PresenterPayout, error) {
	log.Printf("[PAYOUT-SERVICE] Getting payout by ID: %d", id)

	payout, err := models.GetPresenterPayoutByID(s.db, id)
	if err != nil {
		log.Printf("[PAYOUT-SERVICE] Error getting payout: %v", err)
		return nil, fmt.Errorf("payout not found: %w", err)
	}

	return payout, nil
}

// GetPresenterPayouts retrieves all payouts for a presenter
func (s *PayoutService) GetPresenterPayouts(presenterID int) ([]*models.PresenterPayout, error) {
	log.Printf("[PAYOUT-SERVICE] Getting payouts for presenter: %d", presenterID)

	payouts, err := models.GetPresenterPayoutsByPresenterID(s.db, presenterID)
	if err != nil {
		log.Printf("[PAYOUT-SERVICE] Error getting presenter payouts: %v", err)
		return nil, fmt.Errorf("failed to get presenter payouts: %w", err)
	}

	log.Printf("[PAYOUT-SERVICE] Retrieved %d payouts for presenter %d", len(payouts), presenterID)
	return payouts, nil
}

// GetPayoutsByMonth retrieves all payouts for a specific month
func (s *PayoutService) GetPayoutsByMonth(month time.Time, status string) ([]*models.PresenterPayoutWithDetails, error) {
	// Ensure month is the first day of the month
	month = time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)

	log.Printf("[PAYOUT-SERVICE] Getting payouts for month: %s, status: %s", month.Format("2006-01-02"), status)

	payouts, err := models.GetPresenterPayoutsByMonth(s.db, month, status)
	if err != nil {
		log.Printf("[PAYOUT-SERVICE] Error getting payouts by month: %v", err)
		return nil, fmt.Errorf("failed to get payouts by month: %w", err)
	}

	log.Printf("[PAYOUT-SERVICE] Retrieved %d payouts for month %s", len(payouts), month.Format("2006-01-02"))
	return payouts, nil
}

// GetPayoutSummary retrieves summary statistics for a month
func (s *PayoutService) GetPayoutSummary(month time.Time) (*models.PayoutSummary, error) {
	// Ensure month is the first day of the month
	month = time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)

	log.Printf("[PAYOUT-SERVICE] Getting payout summary for: %s", month.Format("2006-01-02"))

	// Call the database function
	query := `SELECT * FROM get_payout_summary($1)`

	summary := &models.PayoutSummary{}
	var statusBreakdownJSON []byte

	err := s.db.QueryRow(query, month).Scan(
		&summary.PayoutMonth,
		&summary.TotalPresenters,
		&summary.TotalVideos,
		&summary.TotalViews,
		&summary.TotalAmount,
		&summary.PendingAmount,
		&summary.ApprovedAmount,
		&summary.PaidAmount,
		&statusBreakdownJSON,
	)

	if err != nil {
		log.Printf("[PAYOUT-SERVICE] Error getting payout summary: %v", err)
		return nil, fmt.Errorf("failed to get payout summary: %w", err)
	}

	summary.StatusBreakdown = json.RawMessage(statusBreakdownJSON)

	return summary, nil
}

// ApprovePayouts approves multiple payouts for payment
func (s *PayoutService) ApprovePayouts(payoutIDs []int, approvedBy int) error {
	log.Printf("[PAYOUT-SERVICE] Approving %d payouts", len(payoutIDs))

	successCount := 0
	errorCount := 0

	for _, id := range payoutIDs {
		input := &models.UpdatePayoutStatusInput{
			Status:    "approved",
			UpdatedBy: approvedBy,
		}

		_, err := models.UpdatePayoutStatus(s.db, id, input)
		if err != nil {
			log.Printf("[PAYOUT-SERVICE] Error approving payout %d: %v", id, err)
			errorCount++
		} else {
			successCount++
		}
	}

	log.Printf("[PAYOUT-SERVICE] Approved %d payouts, %d errors", successCount, errorCount)

	if errorCount > 0 && successCount == 0 {
		return fmt.Errorf("failed to approve any payouts")
	}

	return nil
}

// UpdatePayoutStatus updates the status of a payout
func (s *PayoutService) UpdatePayoutStatus(id int, input *models.UpdatePayoutStatusInput) (*models.PresenterPayout, error) {
	log.Printf("[PAYOUT-SERVICE] Updating payout status: %d to %s", id, input.Status)

	// Validate status
	validStatuses := map[string]bool{
		"pending":    true,
		"approved":   true,
		"processing": true,
		"paid":       true,
		"failed":     true,
		"cancelled":  true,
		"on_hold":    true,
	}

	if !validStatuses[input.Status] {
		return nil, fmt.Errorf("invalid status: %s", input.Status)
	}

	// Check if payout exists
	payout, err := models.GetPresenterPayoutByID(s.db, id)
	if err != nil {
		return nil, fmt.Errorf("payout not found: %w", err)
	}

	// Update status
	updatedPayout, err := models.UpdatePayoutStatus(s.db, id, input)
	if err != nil {
		log.Printf("[PAYOUT-SERVICE] Error updating payout status: %v", err)
		return nil, fmt.Errorf("failed to update payout status: %w", err)
	}

	// If marking as paid, create a transaction record
	if input.Status == "paid" {
		transactionInput := &models.CreatePayoutTransactionInput{
			PayoutID:        &id,
			PresenterID:     payout.PresenterID,
			TransactionType: "payment",
			Amount:          payout.FinalAmount,
			Currency:        payout.Currency,
			PaymentMethod:   input.PaymentMethod,
			Description:     "Monthly payout",
			ProcessedBy:     input.UpdatedBy,
		}

		_, err = models.CreatePayoutTransaction(s.db, transactionInput)
		if err != nil {
			log.Printf("[PAYOUT-SERVICE] Warning: failed to create transaction record: %v", err)
		}
	}

	log.Printf("[PAYOUT-SERVICE] Payout status updated successfully: %d", id)
	return updatedPayout, nil
}

// UpdatePayoutAmounts updates adjustment amounts for a payout
func (s *PayoutService) UpdatePayoutAmounts(id int, input *models.UpdatePayoutAmountsInput) (*models.PresenterPayout, error) {
	log.Printf("[PAYOUT-SERVICE] Updating payout amounts: %d", id)

	// Check if payout exists
	_, err := models.GetPresenterPayoutByID(s.db, id)
	if err != nil {
		return nil, fmt.Errorf("payout not found: %w", err)
	}

	payout, err := models.UpdatePayoutAmounts(s.db, id, input)
	if err != nil {
		log.Printf("[PAYOUT-SERVICE] Error updating payout amounts: %v", err)
		return nil, fmt.Errorf("failed to update payout amounts: %w", err)
	}

	log.Printf("[PAYOUT-SERVICE] Payout amounts updated successfully: %d", id)
	return payout, nil
}

// DeletePayout deletes a payout (use with caution!)
func (s *PayoutService) DeletePayout(id int) error {
	log.Printf("[PAYOUT-SERVICE] Deleting payout: %d", id)

	// Check if payout exists and is not paid
	payout, err := models.GetPresenterPayoutByID(s.db, id)
	if err != nil {
		return fmt.Errorf("payout not found: %w", err)
	}

	if payout.Status == "paid" {
		return fmt.Errorf("cannot delete paid payouts")
	}

	err = models.DeletePresenterPayout(s.db, id)
	if err != nil {
		log.Printf("[PAYOUT-SERVICE] Error deleting payout: %v", err)
		return fmt.Errorf("failed to delete payout: %w", err)
	}

	log.Printf("[PAYOUT-SERVICE] Payout deleted successfully: %d", id)
	return nil
}

// CreateTransaction creates a payout transaction
func (s *PayoutService) CreateTransaction(input *models.CreatePayoutTransactionInput) (*models.PayoutTransaction, error) {
	log.Printf("[PAYOUT-SERVICE] Creating transaction for presenter: %d", input.PresenterID)

	// Validate transaction type
	validTypes := map[string]bool{
		"payment":    true,
		"adjustment": true,
		"refund":     true,
		"chargeback": true,
		"bonus":      true,
		"correction": true,
	}

	if !validTypes[input.TransactionType] {
		return nil, fmt.Errorf("invalid transaction type: %s", input.TransactionType)
	}

	transaction, err := models.CreatePayoutTransaction(s.db, input)
	if err != nil {
		log.Printf("[PAYOUT-SERVICE] Error creating transaction: %v", err)
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	log.Printf("[PAYOUT-SERVICE] Transaction created successfully: ID=%d", transaction.ID)
	return transaction, nil
}

// GetTransactionsByPayoutID retrieves all transactions for a payout
func (s *PayoutService) GetTransactionsByPayoutID(payoutID int) ([]*models.PayoutTransactionWithDetails, error) {
	log.Printf("[PAYOUT-SERVICE] Getting transactions for payout: %d", payoutID)

	transactions, err := models.GetPayoutTransactionsByPayoutID(s.db, payoutID)
	if err != nil {
		log.Printf("[PAYOUT-SERVICE] Error getting transactions: %v", err)
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	log.Printf("[PAYOUT-SERVICE] Retrieved %d transactions for payout %d", len(transactions), payoutID)
	return transactions, nil
}

// GetTransactionsByPresenterID retrieves all transactions for a presenter
func (s *PayoutService) GetTransactionsByPresenterID(presenterID int) ([]*models.PayoutTransactionWithDetails, error) {
	log.Printf("[PAYOUT-SERVICE] Getting transactions for presenter: %d", presenterID)

	transactions, err := models.GetPayoutTransactionsByPresenterID(s.db, presenterID)
	if err != nil {
		log.Printf("[PAYOUT-SERVICE] Error getting transactions: %v", err)
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	log.Printf("[PAYOUT-SERVICE] Retrieved %d transactions for presenter %d", len(transactions), presenterID)
	return transactions, nil
}

// GetRecentTransactions retrieves recent transactions
func (s *PayoutService) GetRecentTransactions(limit int, status string) ([]*models.PayoutTransactionWithDetails, error) {
	log.Printf("[PAYOUT-SERVICE] Getting recent transactions (limit=%d, status=%s)", limit, status)

	transactions, err := models.GetRecentPayoutTransactions(s.db, limit, status)
	if err != nil {
		log.Printf("[PAYOUT-SERVICE] Error getting recent transactions: %v", err)
		return nil, fmt.Errorf("failed to get recent transactions: %w", err)
	}

	log.Printf("[PAYOUT-SERVICE] Retrieved %d recent transactions", len(transactions))
	return transactions, nil
}

// UpdateTransactionStatus updates the status of a transaction
func (s *PayoutService) UpdateTransactionStatus(id int, input *models.UpdateTransactionStatusInput) (*models.PayoutTransaction, error) {
	log.Printf("[PAYOUT-SERVICE] Updating transaction status: %d to %s", id, input.Status)

	// Validate status
	validStatuses := map[string]bool{
		"pending":    true,
		"processing": true,
		"completed":  true,
		"failed":     true,
		"reversed":   true,
	}

	if !validStatuses[input.Status] {
		return nil, fmt.Errorf("invalid transaction status: %s", input.Status)
	}

	transaction, err := models.UpdatePayoutTransactionStatus(s.db, id, input)
	if err != nil {
		log.Printf("[PAYOUT-SERVICE] Error updating transaction status: %v", err)
		return nil, fmt.Errorf("failed to update transaction status: %w", err)
	}

	log.Printf("[PAYOUT-SERVICE] Transaction status updated successfully: %d", id)
	return transaction, nil
}
