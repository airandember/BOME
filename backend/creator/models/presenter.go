package models

import (
	"database/sql"
	"time"

	"bome-backend/infrastructure/database"
)

// Presenter represents a content creator/presenter who appears in videos
type Presenter struct {
	ID               int            `json:"id"`
	UserID           sql.NullInt64  `json:"user_id,omitempty"`
	Name             string         `json:"name"`
	Email            sql.NullString `json:"email,omitempty"`
	Bio              sql.NullString `json:"bio,omitempty"`
	AvatarURL        sql.NullString `json:"avatar_url,omitempty"`
	PaymentMethod    sql.NullString `json:"payment_method,omitempty"`
	StripeConnectID  sql.NullString `json:"stripe_connect_id,omitempty"`
	PaypalEmail      sql.NullString `json:"paypal_email,omitempty"`
	TaxID            sql.NullString `json:"tax_id,omitempty"`
	BankAccountLast4 sql.NullString `json:"bank_account_last4,omitempty"`
	AddressLine1     sql.NullString `json:"address_line1,omitempty"`
	AddressLine2     sql.NullString `json:"address_line2,omitempty"`
	City             sql.NullString `json:"city,omitempty"`
	State            sql.NullString `json:"state,omitempty"`
	PostalCode       sql.NullString `json:"postal_code,omitempty"`
	Country          string         `json:"country"`
	IsActive         bool           `json:"is_active"`
	Verified         bool           `json:"verified"`
	VerifiedAt       sql.NullTime   `json:"verified_at,omitempty"`
	VerifiedBy       sql.NullInt64  `json:"verified_by,omitempty"`
	TotalVideos      int            `json:"total_videos"`
	TotalViews       int64          `json:"total_views"`
	TotalWatchMinutes int64         `json:"total_watch_minutes"`
	TotalEarnings    float64        `json:"total_earnings"`
	LifetimePaid     float64        `json:"lifetime_paid"`
	Notes            sql.NullString `json:"notes,omitempty"`
	InternalNotes    sql.NullString `json:"internal_notes,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// CreatePresenterInput represents input for creating a presenter
type CreatePresenterInput struct {
	UserID          *int   `json:"user_id"`
	Name            string `json:"name" binding:"required"`
	Email           string `json:"email"`
	Bio             string `json:"bio"`
	AvatarURL       string `json:"avatar_url"`
	PaymentMethod   string `json:"payment_method"`
	StripeConnectID string `json:"stripe_connect_id"`
	PaypalEmail     string `json:"paypal_email"`
	TaxID           string `json:"tax_id"`
	AddressLine1    string `json:"address_line1"`
	AddressLine2    string `json:"address_line2"`
	City            string `json:"city"`
	State           string `json:"state"`
	PostalCode      string `json:"postal_code"`
	Country         string `json:"country"`
	Notes           string `json:"notes"`
	InternalNotes   string `json:"internal_notes"`
}

// UpdatePresenterInput represents input for updating a presenter
type UpdatePresenterInput struct {
	Name            *string `json:"name"`
	Email           *string `json:"email"`
	Bio             *string `json:"bio"`
	AvatarURL       *string `json:"avatar_url"`
	PaymentMethod   *string `json:"payment_method"`
	StripeConnectID *string `json:"stripe_connect_id"`
	PaypalEmail     *string `json:"paypal_email"`
	TaxID           *string `json:"tax_id"`
	AddressLine1    *string `json:"address_line1"`
	AddressLine2    *string `json:"address_line2"`
	City            *string `json:"city"`
	State           *string `json:"state"`
	PostalCode      *string `json:"postal_code"`
	Country         *string `json:"country"`
	IsActive        *bool   `json:"is_active"`
	Notes           *string `json:"notes"`
	InternalNotes   *string `json:"internal_notes"`
}

// PresenterStats represents aggregated statistics for a presenter
type PresenterStats struct {
	TotalPresenters     int     `json:"total_presenters"`
	ActivePresenters    int     `json:"active_presenters"`
	VerifiedPresenters  int     `json:"verified_presenters"`
	TotalVideos         int     `json:"total_videos"`
	TotalViews          int64   `json:"total_views"`
	TotalEarnings       float64 `json:"total_earnings"`
	TotalPaid           float64 `json:"total_paid"`
	PendingPayouts      float64 `json:"pending_payouts"`
}

// CreatePresenter inserts a new presenter into the database
func CreatePresenter(db *database.DB, input *CreatePresenterInput) (*Presenter, error) {
	presenter := &Presenter{}
	
	query := `
		INSERT INTO presenters (
			user_id, name, email, bio, avatar_url, payment_method,
			stripe_connect_id, paypal_email, tax_id, address_line1, address_line2,
			city, state, postal_code, country, notes, internal_notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, user_id, name, email, bio, avatar_url, payment_method,
			stripe_connect_id, paypal_email, tax_id, bank_account_last4,
			address_line1, address_line2, city, state, postal_code, country,
			is_active, verified, verified_at, verified_by,
			total_videos, total_views, total_watch_minutes, total_earnings, lifetime_paid,
			notes, internal_notes, created_at, updated_at
	`
	
	country := "US"
	if input.Country != "" {
		country = input.Country
	}
	
	err := db.QueryRow(
		query,
		input.UserID, input.Name, input.Email, input.Bio, input.AvatarURL,
		input.PaymentMethod, input.StripeConnectID, input.PaypalEmail, input.TaxID,
		input.AddressLine1, input.AddressLine2, input.City, input.State,
		input.PostalCode, country, input.Notes, input.InternalNotes,
	).Scan(
		&presenter.ID, &presenter.UserID, &presenter.Name, &presenter.Email,
		&presenter.Bio, &presenter.AvatarURL, &presenter.PaymentMethod,
		&presenter.StripeConnectID, &presenter.PaypalEmail, &presenter.TaxID,
		&presenter.BankAccountLast4, &presenter.AddressLine1, &presenter.AddressLine2,
		&presenter.City, &presenter.State, &presenter.PostalCode, &presenter.Country,
		&presenter.IsActive, &presenter.Verified, &presenter.VerifiedAt, &presenter.VerifiedBy,
		&presenter.TotalVideos, &presenter.TotalViews, &presenter.TotalWatchMinutes,
		&presenter.TotalEarnings, &presenter.LifetimePaid, &presenter.Notes,
		&presenter.InternalNotes, &presenter.CreatedAt, &presenter.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return presenter, nil
}

// GetPresenterByID retrieves a presenter by ID
func GetPresenterByID(db *database.DB, id int) (*Presenter, error) {
	presenter := &Presenter{}
	
	query := `
		SELECT id, user_id, name, email, bio, avatar_url, payment_method,
			stripe_connect_id, paypal_email, tax_id, bank_account_last4,
			address_line1, address_line2, city, state, postal_code, country,
			is_active, verified, verified_at, verified_by,
			total_videos, total_views, total_watch_minutes, total_earnings, lifetime_paid,
			notes, internal_notes, created_at, updated_at
		FROM presenters
		WHERE id = $1
	`
	
	err := db.QueryRow(query, id).Scan(
		&presenter.ID, &presenter.UserID, &presenter.Name, &presenter.Email,
		&presenter.Bio, &presenter.AvatarURL, &presenter.PaymentMethod,
		&presenter.StripeConnectID, &presenter.PaypalEmail, &presenter.TaxID,
		&presenter.BankAccountLast4, &presenter.AddressLine1, &presenter.AddressLine2,
		&presenter.City, &presenter.State, &presenter.PostalCode, &presenter.Country,
		&presenter.IsActive, &presenter.Verified, &presenter.VerifiedAt, &presenter.VerifiedBy,
		&presenter.TotalVideos, &presenter.TotalViews, &presenter.TotalWatchMinutes,
		&presenter.TotalEarnings, &presenter.LifetimePaid, &presenter.Notes,
		&presenter.InternalNotes, &presenter.CreatedAt, &presenter.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return presenter, nil
}

// GetPresenters retrieves all presenters with optional filtering
func GetPresenters(db *database.DB, activeOnly bool, verifiedOnly bool) ([]*Presenter, error) {
	query := `
		SELECT id, user_id, name, email, bio, avatar_url, payment_method,
			stripe_connect_id, paypal_email, tax_id, bank_account_last4,
			address_line1, address_line2, city, state, postal_code, country,
			is_active, verified, verified_at, verified_by,
			total_videos, total_views, total_watch_minutes, total_earnings, lifetime_paid,
			notes, internal_notes, created_at, updated_at
		FROM presenters
		WHERE 1=1
	`
	
	args := []interface{}{}
	argIndex := 1
	
	if activeOnly {
		query += ` AND is_active = $` + string(rune('0'+argIndex))
		args = append(args, true)
		argIndex++
	}
	
	if verifiedOnly {
		query += ` AND verified = $` + string(rune('0'+argIndex))
		args = append(args, true)
		argIndex++
	}
	
	query += ` ORDER BY name ASC`
	
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	presenters := []*Presenter{}
	
	for rows.Next() {
		presenter := &Presenter{}
		err := rows.Scan(
			&presenter.ID, &presenter.UserID, &presenter.Name, &presenter.Email,
			&presenter.Bio, &presenter.AvatarURL, &presenter.PaymentMethod,
			&presenter.StripeConnectID, &presenter.PaypalEmail, &presenter.TaxID,
			&presenter.BankAccountLast4, &presenter.AddressLine1, &presenter.AddressLine2,
			&presenter.City, &presenter.State, &presenter.PostalCode, &presenter.Country,
			&presenter.IsActive, &presenter.Verified, &presenter.VerifiedAt, &presenter.VerifiedBy,
			&presenter.TotalVideos, &presenter.TotalViews, &presenter.TotalWatchMinutes,
			&presenter.TotalEarnings, &presenter.LifetimePaid, &presenter.Notes,
			&presenter.InternalNotes, &presenter.CreatedAt, &presenter.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		presenters = append(presenters, presenter)
	}
	
	return presenters, nil
}

// UpdatePresenter updates a presenter's information
func UpdatePresenter(db *database.DB, id int, input *UpdatePresenterInput) (*Presenter, error) {
	// Build dynamic UPDATE query
	query := `UPDATE presenters SET updated_at = NOW()`
	args := []interface{}{}
	argIndex := 1
	
	if input.Name != nil {
		query += `, name = $` + string(rune('0'+argIndex))
		args = append(args, *input.Name)
		argIndex++
	}
	if input.Email != nil {
		query += `, email = $` + string(rune('0'+argIndex))
		args = append(args, *input.Email)
		argIndex++
	}
	if input.Bio != nil {
		query += `, bio = $` + string(rune('0'+argIndex))
		args = append(args, *input.Bio)
		argIndex++
	}
	if input.AvatarURL != nil {
		query += `, avatar_url = $` + string(rune('0'+argIndex))
		args = append(args, *input.AvatarURL)
		argIndex++
	}
	if input.PaymentMethod != nil {
		query += `, payment_method = $` + string(rune('0'+argIndex))
		args = append(args, *input.PaymentMethod)
		argIndex++
	}
	if input.StripeConnectID != nil {
		query += `, stripe_connect_id = $` + string(rune('0'+argIndex))
		args = append(args, *input.StripeConnectID)
		argIndex++
	}
	if input.PaypalEmail != nil {
		query += `, paypal_email = $` + string(rune('0'+argIndex))
		args = append(args, *input.PaypalEmail)
		argIndex++
	}
	if input.TaxID != nil {
		query += `, tax_id = $` + string(rune('0'+argIndex))
		args = append(args, *input.TaxID)
		argIndex++
	}
	if input.AddressLine1 != nil {
		query += `, address_line1 = $` + string(rune('0'+argIndex))
		args = append(args, *input.AddressLine1)
		argIndex++
	}
	if input.AddressLine2 != nil {
		query += `, address_line2 = $` + string(rune('0'+argIndex))
		args = append(args, *input.AddressLine2)
		argIndex++
	}
	if input.City != nil {
		query += `, city = $` + string(rune('0'+argIndex))
		args = append(args, *input.City)
		argIndex++
	}
	if input.State != nil {
		query += `, state = $` + string(rune('0'+argIndex))
		args = append(args, *input.State)
		argIndex++
	}
	if input.PostalCode != nil {
		query += `, postal_code = $` + string(rune('0'+argIndex))
		args = append(args, *input.PostalCode)
		argIndex++
	}
	if input.Country != nil {
		query += `, country = $` + string(rune('0'+argIndex))
		args = append(args, *input.Country)
		argIndex++
	}
	if input.IsActive != nil {
		query += `, is_active = $` + string(rune('0'+argIndex))
		args = append(args, *input.IsActive)
		argIndex++
	}
	if input.Notes != nil {
		query += `, notes = $` + string(rune('0'+argIndex))
		args = append(args, *input.Notes)
		argIndex++
	}
	if input.InternalNotes != nil {
		query += `, internal_notes = $` + string(rune('0'+argIndex))
		args = append(args, *input.InternalNotes)
		argIndex++
	}
	
	query += ` WHERE id = $` + string(rune('0'+argIndex))
	args = append(args, id)
	
	_, err := db.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	
	// Return updated presenter
	return GetPresenterByID(db, id)
}

// DeletePresenter soft-deletes a presenter (sets is_active = false)
func DeletePresenter(db *database.DB, id int) error {
	query := `UPDATE presenters SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}

// VerifyPresenter marks a presenter as verified
func VerifyPresenter(db *database.DB, presenterID int, verifiedBy int) error {
	query := `
		UPDATE presenters 
		SET verified = true, verified_at = NOW(), verified_by = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := db.Exec(query, verifiedBy, presenterID)
	return err
}

// GetPresenterStats returns aggregated statistics for all presenters
func GetPresenterStats(db *database.DB) (*PresenterStats, error) {
	stats := &PresenterStats{}
	
	query := `
		SELECT 
			COUNT(*) as total_presenters,
			COUNT(*) FILTER (WHERE is_active = true) as active_presenters,
			COUNT(*) FILTER (WHERE verified = true) as verified_presenters,
			COALESCE(SUM(total_videos), 0) as total_videos,
			COALESCE(SUM(total_views), 0) as total_views,
			COALESCE(SUM(total_earnings), 0) as total_earnings,
			COALESCE(SUM(lifetime_paid), 0) as total_paid
		FROM presenters
	`
	
	err := db.QueryRow(query).Scan(
		&stats.TotalPresenters,
		&stats.ActivePresenters,
		&stats.VerifiedPresenters,
		&stats.TotalVideos,
		&stats.TotalViews,
		&stats.TotalEarnings,
		&stats.TotalPaid,
	)
	if err != nil {
		return nil, err
	}
	
	// Calculate pending payouts
	stats.PendingPayouts = stats.TotalEarnings - stats.TotalPaid
	
	return stats, nil
}

