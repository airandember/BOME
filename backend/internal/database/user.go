package database

import (
	"database/sql"
	"time"
)

// User represents a user in the system
type User struct {
	ID                int
	Email             string
	PasswordHash      string
	FirstName         string
	LastName          string
	Role              string
	EmailVerified     bool
	StripeCustomerID  sql.NullString
	ResetToken        sql.NullString
	ResetTokenExpiry  sql.NullTime
	VerificationToken sql.NullString
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// CreateUser inserts a new user into the database
func (db *DB) CreateUser(email, passwordHash, firstName, lastName, role string) (*User, error) {
	var id int
	err := db.QueryRow(
		`INSERT INTO users (email, password_hash, first_name, last_name, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id`,
		email, passwordHash, firstName, lastName, role,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return db.GetUserByID(id)
}

// GetUserByID retrieves a user by ID
func (db *DB) GetUserByID(id int) (*User, error) {
	user := &User{}
	err := db.QueryRow(
		`SELECT id, email, password_hash, first_name, last_name, role, email_verified, stripe_customer_id, reset_token, reset_token_expiry, verification_token, created_at, updated_at FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Role, &user.EmailVerified, &user.StripeCustomerID, &user.ResetToken, &user.ResetTokenExpiry, &user.VerificationToken, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (db *DB) GetUserByEmail(email string) (*User, error) {
	user := &User{}
	err := db.QueryRow(
		`SELECT id, email, password_hash, first_name, last_name, role, email_verified, stripe_customer_id, reset_token, reset_token_expiry, verification_token, created_at, updated_at FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Role, &user.EmailVerified, &user.StripeCustomerID, &user.ResetToken, &user.ResetTokenExpiry, &user.VerificationToken, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// CheckUserExists checks if a user exists by email
func (db *DB) CheckUserExists(email string) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE email = $1`, email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateUserPassword updates a user's password hash
func (db *DB) UpdateUserPassword(userID int, newPasswordHash string) error {
	_, err := db.Exec(`UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`, newPasswordHash, userID)
	return err
}

// SetUserEmailVerified sets a user's email as verified
func (db *DB) SetUserEmailVerified(userID int) error {
	_, err := db.Exec(`UPDATE users SET email_verified = TRUE, updated_at = NOW() WHERE id = $1`, userID)
	return err
}

// SetPasswordResetToken sets a password reset token for a user
func (db *DB) SetPasswordResetToken(userID int, token string, expiry time.Time) error {
	_, err := db.Exec(`UPDATE users SET reset_token = $1, reset_token_expiry = $2, updated_at = NOW() WHERE id = $3`, token, expiry, userID)
	return err
}

// GetUserByResetToken retrieves a user by reset token
func (db *DB) GetUserByResetToken(token string) (*User, error) {
	user := &User{}
	err := db.QueryRow(
		`SELECT id, email, password_hash, first_name, last_name, role, email_verified, stripe_customer_id, reset_token, reset_token_expiry, verification_token, created_at, updated_at FROM users WHERE reset_token = $1 AND reset_token_expiry > NOW()`,
		token,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Role, &user.EmailVerified, &user.StripeCustomerID, &user.ResetToken, &user.ResetTokenExpiry, &user.VerificationToken, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// ClearPasswordResetToken clears the reset token after use
func (db *DB) ClearPasswordResetToken(userID int) error {
	_, err := db.Exec(`UPDATE users SET reset_token = NULL, reset_token_expiry = NULL, updated_at = NOW() WHERE id = $1`, userID)
	return err
}

// SetVerificationToken sets an email verification token for a user
func (db *DB) SetVerificationToken(userID int, token string) error {
	_, err := db.Exec(`UPDATE users SET verification_token = $1, updated_at = NOW() WHERE id = $2`, token, userID)
	return err
}

// GetUserByVerificationToken retrieves a user by verification token
func (db *DB) GetUserByVerificationToken(token string) (*User, error) {
	user := &User{}
	err := db.QueryRow(
		`SELECT id, email, password_hash, first_name, last_name, role, email_verified, stripe_customer_id, reset_token, reset_token_expiry, verification_token, created_at, updated_at FROM users WHERE verification_token = $1`,
		token,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Role, &user.EmailVerified, &user.StripeCustomerID, &user.ResetToken, &user.ResetTokenExpiry, &user.VerificationToken, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// ClearVerificationToken clears the verification token after use
func (db *DB) ClearVerificationToken(userID int) error {
	_, err := db.Exec(`UPDATE users SET verification_token = NULL, updated_at = NOW() WHERE id = $1`, userID)
	return err
}

// UpdateUserStripeCustomerID updates a user's Stripe customer ID
func (db *DB) UpdateUserStripeCustomerID(userID int, stripeCustomerID string) error {
	_, err := db.Exec(`UPDATE users SET stripe_customer_id = $1, updated_at = NOW() WHERE id = $2`, stripeCustomerID, userID)
	return err
}
