package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
	LastLogout  sql.NullTime
	MaxSessions int
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

// Session represents an active user session
type Session struct {
	ID           string    `json:"id"`
	UserID       int       `json:"user_id"`
	TokenID      string    `json:"token_id"`
	DeviceInfo   string    `json:"device_info"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	LastActivity time.Time `json:"last_activity"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
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
		`SELECT id, email, password_hash, first_name, last_name, role, email_verified, stripe_customer_id, reset_token, reset_token_expiry, verification_token, bio, location, website, phone, avatar_url, preferences, last_login, last_logout, max_sessions, created_at, updated_at FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Role, &user.EmailVerified, &user.StripeCustomerID, &user.ResetToken, &user.ResetTokenExpiry, &user.VerificationToken, &user.Bio, &user.Location, &user.Website, &user.Phone, &user.AvatarURL, &user.Preferences, &user.LastLogin, &user.LastLogout, &user.MaxSessions, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (db *DB) GetUserByEmail(email string) (*User, error) {
	user := &User{}
	err := db.QueryRow(
		`SELECT id, email, password_hash, first_name, last_name, role, email_verified, stripe_customer_id, reset_token, reset_token_expiry, verification_token, bio, location, website, phone, avatar_url, preferences, last_login, last_logout, max_sessions, created_at, updated_at FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Role, &user.EmailVerified, &user.StripeCustomerID, &user.ResetToken, &user.ResetTokenExpiry, &user.VerificationToken, &user.Bio, &user.Location, &user.Website, &user.Phone, &user.AvatarURL, &user.Preferences, &user.LastLogin, &user.LastLogout, &user.MaxSessions, &user.CreatedAt, &user.UpdatedAt)
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

// UpdateLastLogout updates the last logout timestamp for a user
func (db *DB) UpdateLastLogout(userID int) error {
	_, err := db.Exec(`UPDATE users SET last_logout = NOW(), updated_at = NOW() WHERE id = $1`, userID)
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

// SetVerificationToken sets an email verification token for a user with expiration
func (db *DB) SetVerificationToken(userID int, token string) error {
	_, err := db.Exec(`UPDATE users SET verification_token = $1, updated_at = NOW() WHERE id = $2`, token, userID)
	return err
}

// GetUserByVerificationToken retrieves a user by verification token (with expiration check)
func (db *DB) GetUserByVerificationToken(token string) (*User, error) {
	user := &User{}
	err := db.QueryRow(
		`SELECT id, email, password_hash, first_name, last_name, role, email_verified, stripe_customer_id, reset_token, reset_token_expiry, verification_token, created_at, updated_at FROM users WHERE verification_token = $1 AND updated_at > NOW() - INTERVAL '24 hours'`,
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

// CreateSession creates a new user session
func (db *DB) CreateSession(userID int, tokenID, deviceInfo, ipAddress, userAgent string, expiresAt time.Time) (*Session, error) {
	sessionID := fmt.Sprintf("sess_%d_%s", userID, time.Now().Format("20060102150405"))

	_, err := db.Exec(`
		INSERT INTO user_sessions (session_id, user_id, token_id, device_info, ip_address, user_agent, last_activity, is_active, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), TRUE, NOW(), $7)
	`, sessionID, userID, tokenID, deviceInfo, ipAddress, userAgent, expiresAt)

	if err != nil {
		return nil, err
	}

	return &Session{
		ID:           sessionID,
		UserID:       userID,
		TokenID:      tokenID,
		DeviceInfo:   deviceInfo,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		LastActivity: time.Now(),
		IsActive:     true,
		CreatedAt:    time.Now(),
		ExpiresAt:    expiresAt,
	}, nil
}

// GetActiveSessions retrieves all active sessions for a user
func (db *DB) GetActiveSessions(userID int) ([]*Session, error) {
	rows, err := db.Query(`
		SELECT session_id, user_id, token_id, device_info, ip_address, user_agent, last_activity, is_active, created_at, expires_at
		FROM user_sessions 
		WHERE user_id = $1 AND is_active = TRUE AND expires_at > NOW()
		ORDER BY last_activity DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		session := &Session{}
		err := rows.Scan(&session.ID, &session.UserID, &session.TokenID, &session.DeviceInfo, &session.IPAddress, &session.UserAgent, &session.LastActivity, &session.IsActive, &session.CreatedAt, &session.ExpiresAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// DeactivateSession deactivates a specific session
func (db *DB) DeactivateSession(sessionID string) error {
	_, err := db.Exec(`UPDATE user_sessions SET is_active = FALSE WHERE session_id = $1`, sessionID)
	return err
}

// DeactivateAllUserSessions deactivates all sessions for a user
func (db *DB) DeactivateAllUserSessions(userID int) error {
	_, err := db.Exec(`UPDATE user_sessions SET is_active = FALSE WHERE user_id = $1`, userID)
	return err
}

// UpdateSessionActivity updates the last activity time for a session
func (db *DB) UpdateSessionActivity(sessionID string) error {
	_, err := db.Exec(`UPDATE user_sessions SET last_activity = NOW() WHERE session_id = $1`, sessionID)
	return err
}

// UpdateSessionActivityByTokenID updates the last activity time for a session by token ID
func (db *DB) UpdateSessionActivityByTokenID(tokenID string) error {
	_, err := db.Exec(`UPDATE user_sessions SET last_activity = NOW() WHERE token_id = $1 AND is_active = TRUE AND expires_at > NOW()`, tokenID)
	return err
}

// CleanupExpiredSessions removes expired sessions
func (db *DB) CleanupExpiredSessions() error {
	_, err := db.Exec(`DELETE FROM user_sessions WHERE expires_at < NOW()`)
	return err
}

// GetSessionCount returns the number of active sessions for a user
func (db *DB) GetSessionCount(userID int) (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM user_sessions WHERE user_id = $1 AND is_active = TRUE AND expires_at > NOW()`, userID).Scan(&count)
	return count, err
}

// CheckSessionLimit checks if user has reached maximum allowed sessions
func (db *DB) CheckSessionLimit(userID int, maxSessions int) (bool, error) {
	count, err := db.GetSessionCount(userID)
	if err != nil {
		return false, err
	}
	return count < maxSessions, nil
}

// CleanupExpiredTokens removes expired verification and reset tokens
func (db *DB) CleanupExpiredTokens() error {
	// Clean up expired verification tokens (older than 24 hours)
	_, err := db.Exec(`UPDATE users SET verification_token = NULL WHERE updated_at < NOW() - INTERVAL '24 hours' AND verification_token IS NOT NULL`)
	if err != nil {
		return err
	}

	// Clean up expired reset tokens (older than 1 hour)
	_, err = db.Exec(`UPDATE users SET reset_token = NULL, reset_token_expiry = NULL WHERE reset_token_expiry < NOW() AND reset_token IS NOT NULL`)
	return err
}

// GetUsers retrieves a list of users with pagination and filtering
func (db *DB) GetUsers(limit, offset int, role, search string) ([]*User, error) {
	query := `SELECT id, email, password_hash, first_name, last_name, role, email_verified, 
              stripe_customer_id, reset_token, reset_token_expiry, verification_token, bio, 
              location, website, phone, avatar_url, preferences, last_login, last_logout, 
              max_sessions, created_at, updated_at 
              FROM users WHERE 1=1`

	args := []interface{}{}
	argCount := 0

	if role != "" {
		argCount++
		query += fmt.Sprintf(" AND role = $%d", argCount)
		args = append(args, role)
	}

	if search != "" {
		argCount++
		query += fmt.Sprintf(" AND (email ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d)",
			argCount, argCount, argCount)
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm)
	}

	query += " ORDER BY created_at DESC LIMIT $%d OFFSET $%d"
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName,
			&user.Role, &user.EmailVerified, &user.StripeCustomerID, &user.ResetToken,
			&user.ResetTokenExpiry, &user.VerificationToken, &user.Bio, &user.Location,
			&user.Website, &user.Phone, &user.AvatarURL, &user.Preferences, &user.LastLogin,
			&user.LastLogout, &user.MaxSessions, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// UpdateUserRole updates a user's role
func (db *DB) UpdateUserRole(userID int, newRole string) error {
	query := `UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2`
	_, err := db.Exec(query, newRole, userID)
	return err
}

// DeleteUser deletes a user from the system
func (db *DB) DeleteUser(userID int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := db.Exec(query, userID)
	return err
}

// GetUserCount returns the total number of users
func (db *DB) GetUserCount() (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}
