package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"bome-backend/infrastructure/database"
)

// PresenterPayout represents a monthly payout record for a presenter
type PresenterPayout struct {
	ID                  int             `json:"id"`
	PresenterID         int             `json:"presenter_id"`
	FormulaID           sql.NullInt64   `json:"formula_id,omitempty"`
	PayoutMonth         time.Time       `json:"payout_month"`
	TotalVideos         int             `json:"total_videos"`
	TotalViews          int64           `json:"total_views"`
	TotalWatchMinutes   int64           `json:"total_watch_minutes"`
	UniqueViewers       int             `json:"unique_viewers"`
	SubscriberViews     int64           `json:"subscriber_views"`
	AvgCompletionRate   float64         `json:"avg_completion_rate"`
	TotalEngagement     int             `json:"total_engagement"`
	BaseAmount          float64         `json:"base_amount"`
	BonusAmount         float64         `json:"bonus_amount"`
	AdjustmentAmount    float64         `json:"adjustment_amount"`
	Deductions          float64         `json:"deductions"`
	FinalAmount         float64         `json:"final_amount"`
	Currency            string          `json:"currency"`
	Status              string          `json:"status"`
	PaymentMethod       sql.NullString  `json:"payment_method,omitempty"`
	PaymentReference    sql.NullString  `json:"payment_reference,omitempty"`
	PaymentFee          float64         `json:"payment_fee"`
	PaidAt              sql.NullTime    `json:"paid_at,omitempty"`
	CalculationData     json.RawMessage `json:"calculation_data,omitempty"`
	Notes               sql.NullString  `json:"notes,omitempty"`
	AdminNotes          sql.NullString  `json:"admin_notes,omitempty"`
	CalculatedBy        sql.NullInt64   `json:"calculated_by,omitempty"`
	CalculatedAt        sql.NullTime    `json:"calculated_at,omitempty"`
	ApprovedBy          sql.NullInt64   `json:"approved_by,omitempty"`
	ApprovedAt          sql.NullTime    `json:"approved_at,omitempty"`
	PaidBy              sql.NullInt64   `json:"paid_by,omitempty"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

// PresenterPayoutWithDetails includes presenter details
type PresenterPayoutWithDetails struct {
	ID                  int             `json:"id"`
	PresenterID         int             `json:"presenter_id"`
	PresenterName       string          `json:"presenter_name"`
	PresenterEmail      string          `json:"presenter_email"`
	FormulaID           sql.NullInt64   `json:"formula_id,omitempty"`
	FormulaName         string          `json:"formula_name"`
	PayoutMonth         time.Time       `json:"payout_month"`
	TotalVideos         int             `json:"total_videos"`
	TotalViews          int64           `json:"total_views"`
	TotalWatchMinutes   int64           `json:"total_watch_minutes"`
	BaseAmount          float64         `json:"base_amount"`
	BonusAmount         float64         `json:"bonus_amount"`
	AdjustmentAmount    float64         `json:"adjustment_amount"`
	Deductions          float64         `json:"deductions"`
	FinalAmount         float64         `json:"final_amount"`
	Currency            string          `json:"currency"`
	Status              string          `json:"status"`
	PaymentMethod       sql.NullString  `json:"payment_method,omitempty"`
	PaymentReference    sql.NullString  `json:"payment_reference,omitempty"`
	PaidAt              sql.NullTime    `json:"paid_at,omitempty"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

// UpdatePayoutStatusInput represents input for updating payout status
type UpdatePayoutStatusInput struct {
	Status           string  `json:"status" binding:"required"`
	PaymentMethod    string  `json:"payment_method"`
	PaymentReference string  `json:"payment_reference"`
	PaymentFee       float64 `json:"payment_fee"`
	Notes            string  `json:"notes"`
	AdminNotes       string  `json:"admin_notes"`
	UpdatedBy        int     `json:"updated_by"`
}

// UpdatePayoutAmountsInput represents input for manual adjustments
type UpdatePayoutAmountsInput struct {
	AdjustmentAmount float64 `json:"adjustment_amount"`
	Deductions       float64 `json:"deductions"`
	AdminNotes       string  `json:"admin_notes"`
	UpdatedBy        int     `json:"updated_by"`
}

// PayoutSummary represents aggregated payout statistics for a month
type PayoutSummary struct {
	PayoutMonth         time.Time       `json:"payout_month"`
	TotalPresenters     int             `json:"total_presenters"`
	TotalVideos         int             `json:"total_videos"`
	TotalViews          int64           `json:"total_views"`
	TotalAmount         float64         `json:"total_amount"`
	PendingAmount       float64         `json:"pending_amount"`
	ApprovedAmount      float64         `json:"approved_amount"`
	PaidAmount          float64         `json:"paid_amount"`
	StatusBreakdown     json.RawMessage `json:"status_breakdown"`
}

// GetPresenterPayoutByID retrieves a payout by ID
func GetPresenterPayoutByID(db *database.DB, id int) (*PresenterPayout, error) {
	payout := &PresenterPayout{}
	
	query := `
		SELECT id, presenter_id, formula_id, payout_month, total_videos, total_views,
			total_watch_minutes, unique_viewers, subscriber_views, avg_completion_rate,
			total_engagement, base_amount, bonus_amount, adjustment_amount, deductions,
			final_amount, currency, status, payment_method, payment_reference, payment_fee,
			paid_at, calculation_data, notes, admin_notes, calculated_by, calculated_at,
			approved_by, approved_at, paid_by, created_at, updated_at
		FROM presenter_payouts
		WHERE id = $1
	`
	
	err := db.QueryRow(query, id).Scan(
		&payout.ID, &payout.PresenterID, &payout.FormulaID, &payout.PayoutMonth,
		&payout.TotalVideos, &payout.TotalViews, &payout.TotalWatchMinutes,
		&payout.UniqueViewers, &payout.SubscriberViews, &payout.AvgCompletionRate,
		&payout.TotalEngagement, &payout.BaseAmount, &payout.BonusAmount,
		&payout.AdjustmentAmount, &payout.Deductions, &payout.FinalAmount,
		&payout.Currency, &payout.Status, &payout.PaymentMethod,
		&payout.PaymentReference, &payout.PaymentFee, &payout.PaidAt,
		&payout.CalculationData, &payout.Notes, &payout.AdminNotes,
		&payout.CalculatedBy, &payout.CalculatedAt, &payout.ApprovedBy,
		&payout.ApprovedAt, &payout.PaidBy, &payout.CreatedAt, &payout.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return payout, nil
}

// GetPresenterPayoutsByPresenterID retrieves all payouts for a presenter
func GetPresenterPayoutsByPresenterID(db *database.DB, presenterID int) ([]*PresenterPayout, error) {
	query := `
		SELECT id, presenter_id, formula_id, payout_month, total_videos, total_views,
			total_watch_minutes, unique_viewers, subscriber_views, avg_completion_rate,
			total_engagement, base_amount, bonus_amount, adjustment_amount, deductions,
			final_amount, currency, status, payment_method, payment_reference, payment_fee,
			paid_at, calculation_data, notes, admin_notes, calculated_by, calculated_at,
			approved_by, approved_at, paid_by, created_at, updated_at
		FROM presenter_payouts
		WHERE presenter_id = $1
		ORDER BY payout_month DESC
	`
	
	rows, err := db.Query(query, presenterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	payouts := []*PresenterPayout{}
	
	for rows.Next() {
		payout := &PresenterPayout{}
		err := rows.Scan(
			&payout.ID, &payout.PresenterID, &payout.FormulaID, &payout.PayoutMonth,
			&payout.TotalVideos, &payout.TotalViews, &payout.TotalWatchMinutes,
			&payout.UniqueViewers, &payout.SubscriberViews, &payout.AvgCompletionRate,
			&payout.TotalEngagement, &payout.BaseAmount, &payout.BonusAmount,
			&payout.AdjustmentAmount, &payout.Deductions, &payout.FinalAmount,
			&payout.Currency, &payout.Status, &payout.PaymentMethod,
			&payout.PaymentReference, &payout.PaymentFee, &payout.PaidAt,
			&payout.CalculationData, &payout.Notes, &payout.AdminNotes,
			&payout.CalculatedBy, &payout.CalculatedAt, &payout.ApprovedBy,
			&payout.ApprovedAt, &payout.PaidBy, &payout.CreatedAt, &payout.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		payouts = append(payouts, payout)
	}
	
	return payouts, nil
}

// GetPresenterPayoutsByMonth retrieves all payouts for a specific month with presenter details
func GetPresenterPayoutsByMonth(db *database.DB, month time.Time, status string) ([]*PresenterPayoutWithDetails, error) {
	query := `
		SELECT 
			pp.id, pp.presenter_id, p.name as presenter_name, p.email as presenter_email,
			pp.formula_id, pf.name as formula_name, pp.payout_month,
			pp.total_videos, pp.total_views, pp.total_watch_minutes,
			pp.base_amount, pp.bonus_amount, pp.adjustment_amount, pp.deductions, pp.final_amount,
			pp.currency, pp.status, pp.payment_method, pp.payment_reference, pp.paid_at,
			pp.created_at, pp.updated_at
		FROM presenter_payouts pp
		INNER JOIN presenters p ON pp.presenter_id = p.id
		LEFT JOIN payout_formulas pf ON pp.formula_id = pf.id
		WHERE pp.payout_month = $1
	`
	
	args := []interface{}{month}
	
	if status != "" {
		query += ` AND pp.status = $2`
		args = append(args, status)
	}
	
	query += ` ORDER BY pp.final_amount DESC`
	
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	payouts := []*PresenterPayoutWithDetails{}
	
	for rows.Next() {
		payout := &PresenterPayoutWithDetails{}
		err := rows.Scan(
			&payout.ID, &payout.PresenterID, &payout.PresenterName, &payout.PresenterEmail,
			&payout.FormulaID, &payout.FormulaName, &payout.PayoutMonth,
			&payout.TotalVideos, &payout.TotalViews, &payout.TotalWatchMinutes,
			&payout.BaseAmount, &payout.BonusAmount, &payout.AdjustmentAmount,
			&payout.Deductions, &payout.FinalAmount, &payout.Currency, &payout.Status,
			&payout.PaymentMethod, &payout.PaymentReference, &payout.PaidAt,
			&payout.CreatedAt, &payout.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		payouts = append(payouts, payout)
	}
	
	return payouts, nil
}

// UpdatePayoutStatus updates the status of a payout
func UpdatePayoutStatus(db *database.DB, id int, input *UpdatePayoutStatusInput) (*PresenterPayout, error) {
	query := `
		UPDATE presenter_payouts
		SET status = $1, updated_at = NOW()
	`
	
	args := []interface{}{input.Status}
	argIndex := 2
	
	if input.PaymentMethod != "" {
		query += `, payment_method = $` + string(rune('0'+argIndex))
		args = append(args, input.PaymentMethod)
		argIndex++
	}
	
	if input.PaymentReference != "" {
		query += `, payment_reference = $` + string(rune('0'+argIndex))
		args = append(args, input.PaymentReference)
		argIndex++
	}
	
	if input.PaymentFee > 0 {
		query += `, payment_fee = $` + string(rune('0'+argIndex))
		args = append(args, input.PaymentFee)
		argIndex++
	}
	
	if input.Notes != "" {
		query += `, notes = $` + string(rune('0'+argIndex))
		args = append(args, input.Notes)
		argIndex++
	}
	
	if input.AdminNotes != "" {
		query += `, admin_notes = $` + string(rune('0'+argIndex))
		args = append(args, input.AdminNotes)
		argIndex++
	}
	
	// Update approved_by and approved_at if status is "approved"
	if input.Status == "approved" {
		query += `, approved_by = $` + string(rune('0'+argIndex)) + `, approved_at = NOW()`
		args = append(args, input.UpdatedBy)
		argIndex++
	}
	
	// Update paid_by and paid_at if status is "paid"
	if input.Status == "paid" {
		query += `, paid_by = $` + string(rune('0'+argIndex)) + `, paid_at = NOW()`
		args = append(args, input.UpdatedBy)
		argIndex++
	}
	
	query += ` WHERE id = $` + string(rune('0'+argIndex))
	args = append(args, id)
	
	_, err := db.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	
	return GetPresenterPayoutByID(db, id)
}

// UpdatePayoutAmounts updates adjustment amounts for a payout
func UpdatePayoutAmounts(db *database.DB, id int, input *UpdatePayoutAmountsInput) (*PresenterPayout, error) {
	query := `
		UPDATE presenter_payouts
		SET adjustment_amount = $1, deductions = $2, 
		    final_amount = base_amount + bonus_amount + $1 - $2,
		    admin_notes = $3, updated_at = NOW()
		WHERE id = $4
	`
	
	_, err := db.Exec(query, input.AdjustmentAmount, input.Deductions, input.AdminNotes, id)
	if err != nil {
		return nil, err
	}
	
	return GetPresenterPayoutByID(db, id)
}

// DeletePresenterPayout deletes a payout (use with caution!)
func DeletePresenterPayout(db *database.DB, id int) error {
	query := `DELETE FROM presenter_payouts WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}

// GetPayoutSummaryForMonth retrieves summary statistics for a month
func GetPayoutSummaryForMonth(db *database.DB, month time.Time) (*PayoutSummary, error) {
	summary := &PayoutSummary{
		PayoutMonth: month,
	}
	
	query := `
		SELECT 
			COUNT(*)::INTEGER as total_presenters,
			COALESCE(SUM(total_videos), 0)::INTEGER as total_videos,
			COALESCE(SUM(total_views), 0) as total_views,
			COALESCE(SUM(final_amount), 0.00) as total_amount,
			COALESCE(SUM(final_amount) FILTER (WHERE status = 'pending'), 0.00) as pending_amount,
			COALESCE(SUM(final_amount) FILTER (WHERE status = 'approved'), 0.00) as approved_amount,
			COALESCE(SUM(final_amount) FILTER (WHERE status = 'paid'), 0.00) as paid_amount,
			jsonb_build_object(
				'pending', COUNT(*) FILTER (WHERE status = 'pending'),
				'approved', COUNT(*) FILTER (WHERE status = 'approved'),
				'processing', COUNT(*) FILTER (WHERE status = 'processing'),
				'paid', COUNT(*) FILTER (WHERE status = 'paid'),
				'failed', COUNT(*) FILTER (WHERE status = 'failed'),
				'cancelled', COUNT(*) FILTER (WHERE status = 'cancelled')
			) as status_breakdown
		FROM presenter_payouts
		WHERE payout_month = $1
	`
	
	var statusBreakdown []byte
	
	err := db.QueryRow(query, month).Scan(
		&summary.TotalPresenters,
		&summary.TotalVideos,
		&summary.TotalViews,
		&summary.TotalAmount,
		&summary.PendingAmount,
		&summary.ApprovedAmount,
		&summary.PaidAmount,
		&statusBreakdown,
	)
	
	if err != nil {
		return nil, err
	}
	
	summary.StatusBreakdown = json.RawMessage(statusBreakdown)
	
	return summary, nil
}

