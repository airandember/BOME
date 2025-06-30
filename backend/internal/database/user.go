package database

import (
	"database/sql"
	"encoding/json"
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
	// Extended profile fields
	Bio         sql.NullString
	Location    sql.NullString
	Website     sql.NullString
	Phone       sql.NullString
	AvatarURL   sql.NullString
	Preferences sql.NullString
	LastLogin   sql.NullTime
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserProfile represents the public profile data for a user
type UserProfile struct {
	ID            int                    `json:"id"`
	Email         string                 `json:"email"`
	FirstName     string                 `json:"first_name"`
	LastName      string                 `json:"last_name"`
	Role          string                 `json:"role"`
	EmailVerified bool                   `json:"email_verified"`
	Bio           *string                `json:"bio,omitempty"`
	Location      *string                `json:"location,omitempty"`
	Website       *string                `json:"website,omitempty"`
	Phone         *string                `json:"phone,omitempty"`
	AvatarURL     *string                `json:"avatar_url,omitempty"`
	Preferences   map[string]interface{} `json:"preferences,omitempty"`
	LastLogin     *time.Time             `json:"last_login,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
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
		`SELECT id, email, password_hash, first_name, last_name, role, email_verified, stripe_customer_id, reset_token, reset_token_expiry, verification_token, bio, location, website, phone, avatar_url, preferences, last_login, created_at, updated_at FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Role, &user.EmailVerified, &user.StripeCustomerID, &user.ResetToken, &user.ResetTokenExpiry, &user.VerificationToken, &user.Bio, &user.Location, &user.Website, &user.Phone, &user.AvatarURL, &user.Preferences, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (db *DB) GetUserByEmail(email string) (*User, error) {
	user := &User{}
	err := db.QueryRow(
		`SELECT id, email, password_hash, first_name, last_name, role, email_verified, stripe_customer_id, reset_token, reset_token_expiry, verification_token, bio, location, website, phone, avatar_url, preferences, last_login, created_at, updated_at FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Role, &user.EmailVerified, &user.StripeCustomerID, &user.ResetToken, &user.ResetTokenExpiry, &user.VerificationToken, &user.Bio, &user.Location, &user.Website, &user.Phone, &user.AvatarURL, &user.Preferences, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserProfile retrieves a user's public profile data
func (db *DB) GetUserProfile(userID int) (*UserProfile, error) {
	user, err := db.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	profile := &UserProfile{
		ID:            user.ID,
		Email:         user.Email,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Role:          user.Role,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}

	// Handle nullable fields
	if user.Bio.Valid {
		profile.Bio = &user.Bio.String
	}
	if user.Location.Valid {
		profile.Location = &user.Location.String
	}
	if user.Website.Valid {
		profile.Website = &user.Website.String
	}
	if user.Phone.Valid {
		profile.Phone = &user.Phone.String
	}
	if user.AvatarURL.Valid {
		profile.AvatarURL = &user.AvatarURL.String
	}
	if user.LastLogin.Valid {
		profile.LastLogin = &user.LastLogin.Time
	}

	// Parse preferences JSON
	if user.Preferences.Valid {
		var prefs map[string]interface{}
		if err := json.Unmarshal([]byte(user.Preferences.String), &prefs); err == nil {
			profile.Preferences = prefs
		}
	}

	return profile, nil
}

// UpdateUserProfile updates a user's profile information
func (db *DB) UpdateUserProfile(userID int, updates map[string]interface{}) error {
	// Build dynamic query based on provided fields
	query := "UPDATE users SET updated_at = NOW()"
	args := []interface{}{}
	argCount := 0

	// Map of allowed fields and their database column names
	allowedFields := map[string]string{
		"first_name": "first_name",
		"last_name":  "last_name",
		"bio":        "bio",
		"location":   "location",
		"website":    "website",
		"phone":      "phone",
		"avatar_url": "avatar_url",
	}

	for field, value := range updates {
		if dbField, exists := allowedFields[field]; exists {
			argCount++
			query += ", " + dbField + " = $" + string(rune(argCount+'0'))
			args = append(args, value)
		}
	}

	// Handle preferences separately (JSON field)
	if prefs, exists := updates["preferences"]; exists {
		if prefsMap, ok := prefs.(map[string]interface{}); ok {
			if prefsJSON, err := json.Marshal(prefsMap); err == nil {
				argCount++
				query += ", preferences = $" + string(rune(argCount+'0'))
				args = append(args, string(prefsJSON))
			}
		}
	}

	query += " WHERE id = $" + string(rune(argCount+1+'0'))
	args = append(args, userID)

	_, err := db.Exec(query, args...)
	return err
}

// UpdateLastLogin updates a user's last login timestamp
func (db *DB) UpdateLastLogin(userID int) error {
	_, err := db.Exec(`UPDATE users SET last_login = NOW(), updated_at = NOW() WHERE id = $1`, userID)
	return err
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
