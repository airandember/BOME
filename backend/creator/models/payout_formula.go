package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"bome-backend/infrastructure/database"
)

// PayoutFormula represents a configurable payout calculation formula
type PayoutFormula struct {
	ID                     int             `json:"id"`
	Name                   string          `json:"name"`
	Description            sql.NullString  `json:"description,omitempty"`
	FormulaType            string          `json:"formula_type"`
	BaseRate               float64         `json:"base_rate"`
	TierConfig             json.RawMessage `json:"tier_config,omitempty"`
	SubscriberMultiplier   float64         `json:"subscriber_multiplier"`
	CompletionMultiplier   float64         `json:"completion_multiplier"`
	EngagementMultiplier   float64         `json:"engagement_multiplier"`
	CompletionThreshold    float64         `json:"completion_threshold"`
	EngagementThreshold    int             `json:"engagement_threshold"`
	MinPayout              float64         `json:"min_payout"`
	MaxPayout              sql.NullFloat64 `json:"max_payout,omitempty"`
	IsActive               bool            `json:"is_active"`
	IsDefault              bool            `json:"is_default"`
	EffectiveDate          sql.NullTime    `json:"effective_date,omitempty"`
	ExpirationDate         sql.NullTime    `json:"expiration_date,omitempty"`
	CreatedBy              sql.NullInt64   `json:"created_by,omitempty"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
}

// CreatePayoutFormulaInput represents input for creating a payout formula
type CreatePayoutFormulaInput struct {
	Name                   string          `json:"name" binding:"required"`
	Description            string          `json:"description"`
	FormulaType            string          `json:"formula_type" binding:"required"`
	BaseRate               float64         `json:"base_rate"`
	TierConfig             json.RawMessage `json:"tier_config"`
	SubscriberMultiplier   float64         `json:"subscriber_multiplier"`
	CompletionMultiplier   float64         `json:"completion_multiplier"`
	EngagementMultiplier   float64         `json:"engagement_multiplier"`
	CompletionThreshold    float64         `json:"completion_threshold"`
	EngagementThreshold    int             `json:"engagement_threshold"`
	MinPayout              float64         `json:"min_payout"`
	MaxPayout              *float64        `json:"max_payout"`
	IsActive               bool            `json:"is_active"`
	IsDefault              bool            `json:"is_default"`
	EffectiveDate          *time.Time      `json:"effective_date"`
	ExpirationDate         *time.Time      `json:"expiration_date"`
	CreatedBy              int             `json:"created_by"`
}

// UpdatePayoutFormulaInput represents input for updating a payout formula
type UpdatePayoutFormulaInput struct {
	Name                   *string          `json:"name"`
	Description            *string          `json:"description"`
	BaseRate               *float64         `json:"base_rate"`
	TierConfig             *json.RawMessage `json:"tier_config"`
	SubscriberMultiplier   *float64         `json:"subscriber_multiplier"`
	CompletionMultiplier   *float64         `json:"completion_multiplier"`
	EngagementMultiplier   *float64         `json:"engagement_multiplier"`
	CompletionThreshold    *float64         `json:"completion_threshold"`
	EngagementThreshold    *int             `json:"engagement_threshold"`
	MinPayout              *float64         `json:"min_payout"`
	MaxPayout              *float64         `json:"max_payout"`
	IsActive               *bool            `json:"is_active"`
	IsDefault              *bool            `json:"is_default"`
	EffectiveDate          *time.Time       `json:"effective_date"`
	ExpirationDate         *time.Time       `json:"expiration_date"`
}

// CreatePayoutFormula inserts a new payout formula into the database
func CreatePayoutFormula(db *database.DB, input *CreatePayoutFormulaInput) (*PayoutFormula, error) {
	formula := &PayoutFormula{}
	
	// Set defaults
	subscriberMult := 1.00
	if input.SubscriberMultiplier > 0 {
		subscriberMult = input.SubscriberMultiplier
	}
	
	completionMult := 1.00
	if input.CompletionMultiplier > 0 {
		completionMult = input.CompletionMultiplier
	}
	
	engagementMult := 1.00
	if input.EngagementMultiplier > 0 {
		engagementMult = input.EngagementMultiplier
	}
	
	completionThresh := 80.00
	if input.CompletionThreshold > 0 {
		completionThresh = input.CompletionThreshold
	}
	
	engagementThresh := 10
	if input.EngagementThreshold > 0 {
		engagementThresh = input.EngagementThreshold
	}
	
	query := `
		INSERT INTO payout_formulas (
			name, description, formula_type, base_rate, tier_config,
			subscriber_multiplier, completion_multiplier, engagement_multiplier,
			completion_threshold, engagement_threshold, min_payout, max_payout,
			is_active, is_default, effective_date, expiration_date, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, name, description, formula_type, base_rate, tier_config,
			subscriber_multiplier, completion_multiplier, engagement_multiplier,
			completion_threshold, engagement_threshold, min_payout, max_payout,
			is_active, is_default, effective_date, expiration_date, created_by,
			created_at, updated_at
	`
	
	err := db.QueryRow(
		query,
		input.Name, input.Description, input.FormulaType, input.BaseRate, input.TierConfig,
		subscriberMult, completionMult, engagementMult,
		completionThresh, engagementThresh, input.MinPayout, input.MaxPayout,
		input.IsActive, input.IsDefault, input.EffectiveDate, input.ExpirationDate, input.CreatedBy,
	).Scan(
		&formula.ID, &formula.Name, &formula.Description, &formula.FormulaType,
		&formula.BaseRate, &formula.TierConfig, &formula.SubscriberMultiplier,
		&formula.CompletionMultiplier, &formula.EngagementMultiplier,
		&formula.CompletionThreshold, &formula.EngagementThreshold,
		&formula.MinPayout, &formula.MaxPayout, &formula.IsActive, &formula.IsDefault,
		&formula.EffectiveDate, &formula.ExpirationDate, &formula.CreatedBy,
		&formula.CreatedAt, &formula.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return formula, nil
}

// GetPayoutFormulaByID retrieves a payout formula by ID
func GetPayoutFormulaByID(db *database.DB, id int) (*PayoutFormula, error) {
	formula := &PayoutFormula{}
	
	query := `
		SELECT id, name, description, formula_type, base_rate, tier_config,
			subscriber_multiplier, completion_multiplier, engagement_multiplier,
			completion_threshold, engagement_threshold, min_payout, max_payout,
			is_active, is_default, effective_date, expiration_date, created_by,
			created_at, updated_at
		FROM payout_formulas
		WHERE id = $1
	`
	
	err := db.QueryRow(query, id).Scan(
		&formula.ID, &formula.Name, &formula.Description, &formula.FormulaType,
		&formula.BaseRate, &formula.TierConfig, &formula.SubscriberMultiplier,
		&formula.CompletionMultiplier, &formula.EngagementMultiplier,
		&formula.CompletionThreshold, &formula.EngagementThreshold,
		&formula.MinPayout, &formula.MaxPayout, &formula.IsActive, &formula.IsDefault,
		&formula.EffectiveDate, &formula.ExpirationDate, &formula.CreatedBy,
		&formula.CreatedAt, &formula.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return formula, nil
}

// GetPayoutFormulas retrieves all payout formulas with optional filtering
func GetPayoutFormulas(db *database.DB, activeOnly bool) ([]*PayoutFormula, error) {
	query := `
		SELECT id, name, description, formula_type, base_rate, tier_config,
			subscriber_multiplier, completion_multiplier, engagement_multiplier,
			completion_threshold, engagement_threshold, min_payout, max_payout,
			is_active, is_default, effective_date, expiration_date, created_by,
			created_at, updated_at
		FROM payout_formulas
		WHERE 1=1
	`
	
	if activeOnly {
		query += ` AND is_active = true`
	}
	
	query += ` ORDER BY is_default DESC, name ASC`
	
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	formulas := []*PayoutFormula{}
	
	for rows.Next() {
		formula := &PayoutFormula{}
		err := rows.Scan(
			&formula.ID, &formula.Name, &formula.Description, &formula.FormulaType,
			&formula.BaseRate, &formula.TierConfig, &formula.SubscriberMultiplier,
			&formula.CompletionMultiplier, &formula.EngagementMultiplier,
			&formula.CompletionThreshold, &formula.EngagementThreshold,
			&formula.MinPayout, &formula.MaxPayout, &formula.IsActive, &formula.IsDefault,
			&formula.EffectiveDate, &formula.ExpirationDate, &formula.CreatedBy,
			&formula.CreatedAt, &formula.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		formulas = append(formulas, formula)
	}
	
	return formulas, nil
}

// GetDefaultPayoutFormula retrieves the default payout formula
func GetDefaultPayoutFormula(db *database.DB) (*PayoutFormula, error) {
	formula := &PayoutFormula{}
	
	query := `
		SELECT id, name, description, formula_type, base_rate, tier_config,
			subscriber_multiplier, completion_multiplier, engagement_multiplier,
			completion_threshold, engagement_threshold, min_payout, max_payout,
			is_active, is_default, effective_date, expiration_date, created_by,
			created_at, updated_at
		FROM payout_formulas
		WHERE is_default = true AND is_active = true
		LIMIT 1
	`
	
	err := db.QueryRow(query).Scan(
		&formula.ID, &formula.Name, &formula.Description, &formula.FormulaType,
		&formula.BaseRate, &formula.TierConfig, &formula.SubscriberMultiplier,
		&formula.CompletionMultiplier, &formula.EngagementMultiplier,
		&formula.CompletionThreshold, &formula.EngagementThreshold,
		&formula.MinPayout, &formula.MaxPayout, &formula.IsActive, &formula.IsDefault,
		&formula.EffectiveDate, &formula.ExpirationDate, &formula.CreatedBy,
		&formula.CreatedAt, &formula.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return formula, nil
}

// UpdatePayoutFormula updates a payout formula
func UpdatePayoutFormula(db *database.DB, id int, input *UpdatePayoutFormulaInput) (*PayoutFormula, error) {
	// Build dynamic UPDATE query
	query := `UPDATE payout_formulas SET updated_at = NOW()`
	args := []interface{}{}
	argIndex := 1
	
	if input.Name != nil {
		query += `, name = $` + string(rune('0'+argIndex))
		args = append(args, *input.Name)
		argIndex++
	}
	if input.Description != nil {
		query += `, description = $` + string(rune('0'+argIndex))
		args = append(args, *input.Description)
		argIndex++
	}
	if input.BaseRate != nil {
		query += `, base_rate = $` + string(rune('0'+argIndex))
		args = append(args, *input.BaseRate)
		argIndex++
	}
	if input.TierConfig != nil {
		query += `, tier_config = $` + string(rune('0'+argIndex))
		args = append(args, *input.TierConfig)
		argIndex++
	}
	if input.SubscriberMultiplier != nil {
		query += `, subscriber_multiplier = $` + string(rune('0'+argIndex))
		args = append(args, *input.SubscriberMultiplier)
		argIndex++
	}
	if input.CompletionMultiplier != nil {
		query += `, completion_multiplier = $` + string(rune('0'+argIndex))
		args = append(args, *input.CompletionMultiplier)
		argIndex++
	}
	if input.EngagementMultiplier != nil {
		query += `, engagement_multiplier = $` + string(rune('0'+argIndex))
		args = append(args, *input.EngagementMultiplier)
		argIndex++
	}
	if input.CompletionThreshold != nil {
		query += `, completion_threshold = $` + string(rune('0'+argIndex))
		args = append(args, *input.CompletionThreshold)
		argIndex++
	}
	if input.EngagementThreshold != nil {
		query += `, engagement_threshold = $` + string(rune('0'+argIndex))
		args = append(args, *input.EngagementThreshold)
		argIndex++
	}
	if input.MinPayout != nil {
		query += `, min_payout = $` + string(rune('0'+argIndex))
		args = append(args, *input.MinPayout)
		argIndex++
	}
	if input.MaxPayout != nil {
		query += `, max_payout = $` + string(rune('0'+argIndex))
		args = append(args, *input.MaxPayout)
		argIndex++
	}
	if input.IsActive != nil {
		query += `, is_active = $` + string(rune('0'+argIndex))
		args = append(args, *input.IsActive)
		argIndex++
	}
	if input.IsDefault != nil {
		// If setting as default, unset all other defaults first
		if *input.IsDefault {
			_, err := db.Exec(`UPDATE payout_formulas SET is_default = false WHERE id != $1`, id)
			if err != nil {
				return nil, err
			}
		}
		query += `, is_default = $` + string(rune('0'+argIndex))
		args = append(args, *input.IsDefault)
		argIndex++
	}
	if input.EffectiveDate != nil {
		query += `, effective_date = $` + string(rune('0'+argIndex))
		args = append(args, *input.EffectiveDate)
		argIndex++
	}
	if input.ExpirationDate != nil {
		query += `, expiration_date = $` + string(rune('0'+argIndex))
		args = append(args, *input.ExpirationDate)
		argIndex++
	}
	
	query += ` WHERE id = $` + string(rune('0'+argIndex))
	args = append(args, id)
	
	_, err := db.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	
	// Return updated formula
	return GetPayoutFormulaByID(db, id)
}

// DeletePayoutFormula soft-deletes a payout formula (sets is_active = false)
func DeletePayoutFormula(db *database.DB, id int) error {
	query := `UPDATE payout_formulas SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}

// SetDefaultFormula sets a formula as the default (and unsets all others)
func SetDefaultFormula(db *database.DB, id int) error {
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// Unset all defaults
	_, err = tx.Exec(`UPDATE payout_formulas SET is_default = false`)
	if err != nil {
		return err
	}
	
	// Set new default
	_, err = tx.Exec(`UPDATE payout_formulas SET is_default = true, updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return err
	}
	
	return tx.Commit()
}

