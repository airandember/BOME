package models

import (
	"database/sql"
	"time"

	"bome-backend/infrastructure/database"
)

// PayoutTransaction represents a payment transaction for a payout
type PayoutTransaction struct {
	ID                     int            `json:"id"`
	PayoutID               sql.NullInt64  `json:"payout_id,omitempty"`
	PresenterID            int            `json:"presenter_id"`
	TransactionType        string         `json:"transaction_type"`
	Amount                 float64        `json:"amount"`
	Currency               string         `json:"currency"`
	PaymentMethod          sql.NullString `json:"payment_method,omitempty"`
	PaymentProvider        sql.NullString `json:"payment_provider,omitempty"`
	ProviderTransactionID  sql.NullString `json:"provider_transaction_id,omitempty"`
	Status                 string         `json:"status"`
	Description            sql.NullString `json:"description,omitempty"`
	Notes                  sql.NullString `json:"notes,omitempty"`
	ErrorMessage           sql.NullString `json:"error_message,omitempty"`
	FeeAmount              float64        `json:"fee_amount"`
	NetAmount              float64        `json:"net_amount"`
	ProcessedBy            sql.NullInt64  `json:"processed_by,omitempty"`
	ProcessedAt            sql.NullTime   `json:"processed_at,omitempty"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
}

// PayoutTransactionWithDetails includes presenter and payout details
type PayoutTransactionWithDetails struct {
	ID                    int            `json:"id"`
	PayoutID              sql.NullInt64  `json:"payout_id,omitempty"`
	PresenterID           int            `json:"presenter_id"`
	PresenterName         string         `json:"presenter_name"`
	TransactionType       string         `json:"transaction_type"`
	Amount                float64        `json:"amount"`
	Currency              string         `json:"currency"`
	PaymentMethod         sql.NullString `json:"payment_method,omitempty"`
	PaymentProvider       sql.NullString `json:"payment_provider,omitempty"`
	ProviderTransactionID sql.NullString `json:"provider_transaction_id,omitempty"`
	Status                string         `json:"status"`
	Description           sql.NullString `json:"description,omitempty"`
	FeeAmount             float64        `json:"fee_amount"`
	NetAmount             float64        `json:"net_amount"`
	ProcessedAt           sql.NullTime   `json:"processed_at,omitempty"`
	CreatedAt             time.Time      `json:"created_at"`
}

// CreatePayoutTransactionInput represents input for creating a transaction
type CreatePayoutTransactionInput struct {
	PayoutID              *int    `json:"payout_id"`
	PresenterID           int     `json:"presenter_id" binding:"required"`
	TransactionType       string  `json:"transaction_type" binding:"required"`
	Amount                float64 `json:"amount" binding:"required"`
	Currency              string  `json:"currency"`
	PaymentMethod         string  `json:"payment_method"`
	PaymentProvider       string  `json:"payment_provider"`
	ProviderTransactionID string  `json:"provider_transaction_id"`
	Description           string  `json:"description"`
	Notes                 string  `json:"notes"`
	FeeAmount             float64 `json:"fee_amount"`
	ProcessedBy           int     `json:"processed_by"`
}

// UpdateTransactionStatusInput represents input for updating transaction status
type UpdateTransactionStatusInput struct {
	Status        string `json:"status" binding:"required"`
	ErrorMessage  string `json:"error_message"`
	ProcessedBy   int    `json:"processed_by"`
}

// CreatePayoutTransaction creates a new transaction record
func CreatePayoutTransaction(db *database.DB, input *CreatePayoutTransactionInput) (*PayoutTransaction, error) {
	transaction := &PayoutTransaction{}
	
	currency := "USD"
	if input.Currency != "" {
		currency = input.Currency
	}
	
	netAmount := input.Amount - input.FeeAmount
	
	query := `
		INSERT INTO payout_transactions (
			payout_id, presenter_id, transaction_type, amount, currency,
			payment_method, payment_provider, provider_transaction_id,
			status, description, notes, fee_amount, net_amount, processed_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, payout_id, presenter_id, transaction_type, amount, currency,
			payment_method, payment_provider, provider_transaction_id, status,
			description, notes, error_message, fee_amount, net_amount,
			processed_by, processed_at, created_at, updated_at
	`
	
	err := db.QueryRow(
		query,
		input.PayoutID, input.PresenterID, input.TransactionType, input.Amount, currency,
		input.PaymentMethod, input.PaymentProvider, input.ProviderTransactionID,
		"pending", input.Description, input.Notes, input.FeeAmount, netAmount, input.ProcessedBy,
	).Scan(
		&transaction.ID, &transaction.PayoutID, &transaction.PresenterID,
		&transaction.TransactionType, &transaction.Amount, &transaction.Currency,
		&transaction.PaymentMethod, &transaction.PaymentProvider,
		&transaction.ProviderTransactionID, &transaction.Status,
		&transaction.Description, &transaction.Notes, &transaction.ErrorMessage,
		&transaction.FeeAmount, &transaction.NetAmount, &transaction.ProcessedBy,
		&transaction.ProcessedAt, &transaction.CreatedAt, &transaction.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return transaction, nil
}

// GetPayoutTransactionByID retrieves a transaction by ID
func GetPayoutTransactionByID(db *database.DB, id int) (*PayoutTransaction, error) {
	transaction := &PayoutTransaction{}
	
	query := `
		SELECT id, payout_id, presenter_id, transaction_type, amount, currency,
			payment_method, payment_provider, provider_transaction_id, status,
			description, notes, error_message, fee_amount, net_amount,
			processed_by, processed_at, created_at, updated_at
		FROM payout_transactions
		WHERE id = $1
	`
	
	err := db.QueryRow(query, id).Scan(
		&transaction.ID, &transaction.PayoutID, &transaction.PresenterID,
		&transaction.TransactionType, &transaction.Amount, &transaction.Currency,
		&transaction.PaymentMethod, &transaction.PaymentProvider,
		&transaction.ProviderTransactionID, &transaction.Status,
		&transaction.Description, &transaction.Notes, &transaction.ErrorMessage,
		&transaction.FeeAmount, &transaction.NetAmount, &transaction.ProcessedBy,
		&transaction.ProcessedAt, &transaction.CreatedAt, &transaction.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return transaction, nil
}

// GetPayoutTransactionsByPayoutID retrieves all transactions for a payout
func GetPayoutTransactionsByPayoutID(db *database.DB, payoutID int) ([]*PayoutTransactionWithDetails, error) {
	query := `
		SELECT 
			pt.id, pt.payout_id, pt.presenter_id, p.name as presenter_name,
			pt.transaction_type, pt.amount, pt.currency, pt.payment_method,
			pt.payment_provider, pt.provider_transaction_id, pt.status,
			pt.description, pt.fee_amount, pt.net_amount, pt.processed_at, pt.created_at
		FROM payout_transactions pt
		INNER JOIN presenters p ON pt.presenter_id = p.id
		WHERE pt.payout_id = $1
		ORDER BY pt.created_at DESC
	`
	
	rows, err := db.Query(query, payoutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	transactions := []*PayoutTransactionWithDetails{}
	
	for rows.Next() {
		transaction := &PayoutTransactionWithDetails{}
		err := rows.Scan(
			&transaction.ID, &transaction.PayoutID, &transaction.PresenterID,
			&transaction.PresenterName, &transaction.TransactionType, &transaction.Amount,
			&transaction.Currency, &transaction.PaymentMethod, &transaction.PaymentProvider,
			&transaction.ProviderTransactionID, &transaction.Status, &transaction.Description,
			&transaction.FeeAmount, &transaction.NetAmount, &transaction.ProcessedAt,
			&transaction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	
	return transactions, nil
}

// GetPayoutTransactionsByPresenterID retrieves all transactions for a presenter
func GetPayoutTransactionsByPresenterID(db *database.DB, presenterID int) ([]*PayoutTransactionWithDetails, error) {
	query := `
		SELECT 
			pt.id, pt.payout_id, pt.presenter_id, p.name as presenter_name,
			pt.transaction_type, pt.amount, pt.currency, pt.payment_method,
			pt.payment_provider, pt.provider_transaction_id, pt.status,
			pt.description, pt.fee_amount, pt.net_amount, pt.processed_at, pt.created_at
		FROM payout_transactions pt
		INNER JOIN presenters p ON pt.presenter_id = p.id
		WHERE pt.presenter_id = $1
		ORDER BY pt.created_at DESC
	`
	
	rows, err := db.Query(query, presenterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	transactions := []*PayoutTransactionWithDetails{}
	
	for rows.Next() {
		transaction := &PayoutTransactionWithDetails{}
		err := rows.Scan(
			&transaction.ID, &transaction.PayoutID, &transaction.PresenterID,
			&transaction.PresenterName, &transaction.TransactionType, &transaction.Amount,
			&transaction.Currency, &transaction.PaymentMethod, &transaction.PaymentProvider,
			&transaction.ProviderTransactionID, &transaction.Status, &transaction.Description,
			&transaction.FeeAmount, &transaction.NetAmount, &transaction.ProcessedAt,
			&transaction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	
	return transactions, nil
}

// GetRecentPayoutTransactions retrieves recent transactions with optional filtering
func GetRecentPayoutTransactions(db *database.DB, limit int, status string) ([]*PayoutTransactionWithDetails, error) {
	query := `
		SELECT 
			pt.id, pt.payout_id, pt.presenter_id, p.name as presenter_name,
			pt.transaction_type, pt.amount, pt.currency, pt.payment_method,
			pt.payment_provider, pt.provider_transaction_id, pt.status,
			pt.description, pt.fee_amount, pt.net_amount, pt.processed_at, pt.created_at
		FROM payout_transactions pt
		INNER JOIN presenters p ON pt.presenter_id = p.id
		WHERE 1=1
	`
	
	args := []interface{}{}
	argIndex := 1
	
	if status != "" {
		query += ` AND pt.status = $` + string(rune('0'+argIndex))
		args = append(args, status)
		argIndex++
	}
	
	query += ` ORDER BY pt.created_at DESC`
	
	if limit > 0 {
		query += ` LIMIT $` + string(rune('0'+argIndex))
		args = append(args, limit)
	}
	
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	transactions := []*PayoutTransactionWithDetails{}
	
	for rows.Next() {
		transaction := &PayoutTransactionWithDetails{}
		err := rows.Scan(
			&transaction.ID, &transaction.PayoutID, &transaction.PresenterID,
			&transaction.PresenterName, &transaction.TransactionType, &transaction.Amount,
			&transaction.Currency, &transaction.PaymentMethod, &transaction.PaymentProvider,
			&transaction.ProviderTransactionID, &transaction.Status, &transaction.Description,
			&transaction.FeeAmount, &transaction.NetAmount, &transaction.ProcessedAt,
			&transaction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	
	return transactions, nil
}

// UpdatePayoutTransactionStatus updates the status of a transaction
func UpdatePayoutTransactionStatus(db *database.DB, id int, input *UpdateTransactionStatusInput) (*PayoutTransaction, error) {
	query := `
		UPDATE payout_transactions
		SET status = $1, updated_at = NOW()
	`
	
	args := []interface{}{input.Status}
	argIndex := 2
	
	if input.ErrorMessage != "" {
		query += `, error_message = $` + string(rune('0'+argIndex))
		args = append(args, input.ErrorMessage)
		argIndex++
	}
	
	if input.Status == "completed" {
		query += `, processed_at = NOW(), processed_by = $` + string(rune('0'+argIndex))
		args = append(args, input.ProcessedBy)
		argIndex++
	}
	
	query += ` WHERE id = $` + string(rune('0'+argIndex))
	args = append(args, id)
	
	_, err := db.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	
	return GetPayoutTransactionByID(db, id)
}

// DeletePayoutTransaction deletes a transaction (use with caution!)
func DeletePayoutTransaction(db *database.DB, id int) error {
	query := `DELETE FROM payout_transactions WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}

