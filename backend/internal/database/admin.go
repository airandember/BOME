package database

import (
	"time"
)

// AdminLog represents an admin action log
type AdminLog struct {
	ID           int
	AdminID      *int
	Action       string
	ResourceType string
	ResourceID   *int
	Details      map[string]interface{}
	IPAddress    string
	UserAgent    string
	CreatedAt    time.Time
}

// CreateAdminLog inserts a new admin log entry
func (db *DB) CreateAdminLog(adminID *int, action, resourceType string, resourceID *int, details map[string]interface{}, ipAddress, userAgent string) error {
	_, err := db.Exec(
		`INSERT INTO admin_logs (admin_id, action, resource_type, resource_id, details, ip_address, user_agent, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`,
		adminID, action, resourceType, resourceID, details, ipAddress, userAgent,
	)
	return err
}

// GetUsers retrieves users with pagination and filtering
func (db *DB) GetUsers(limit, offset int, role, search string) ([]*User, error) {
	query := `SELECT id, email, password_hash, first_name, last_name, role, email_verified, stripe_customer_id, reset_token, reset_token_expiry, verification_token, created_at, updated_at FROM users WHERE 1=1`
	args := []interface{}{}
	argCount := 0

	if role != "" {
		argCount++
		query += ` AND role = $` + string(rune(argCount+'0'))
		args = append(args, role)
	}

	if search != "" {
		argCount++
		query += ` AND (email ILIKE $` + string(rune(argCount+'0')) + ` OR first_name ILIKE $` + string(rune(argCount+'0')) + ` OR last_name ILIKE $` + string(rune(argCount+'0')) + `)`
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm)
	}

	query += ` ORDER BY created_at DESC LIMIT $` + string(rune(argCount+1+'0')) + ` OFFSET $` + string(rune(argCount+2+'0'))
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Role, &user.EmailVerified, &user.StripeCustomerID, &user.ResetToken, &user.ResetTokenExpiry, &user.VerificationToken, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// UpdateUserRole updates a user's role
func (db *DB) UpdateUserRole(userID int, role string) error {
	_, err := db.Exec(`UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2`, role, userID)
	return err
}

// DeleteUser deletes a user and all associated data
func (db *DB) DeleteUser(userID int) error {
	_, err := db.Exec(`DELETE FROM users WHERE id = $1`, userID)
	return err
}

// GetUserCount returns the total number of users
func (db *DB) GetUserCount() (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

// GetVideoCount returns the total number of videos
func (db *DB) GetVideoCount() (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM videos`).Scan(&count)
	return count, err
}

// GetTotalViews returns the total number of video views
func (db *DB) GetTotalViews() (int, error) {
	var total int
	err := db.QueryRow(`SELECT COALESCE(SUM(view_count), 0) FROM videos`).Scan(&total)
	return total, err
}

// GetTotalLikes returns the total number of video likes
func (db *DB) GetTotalLikes() (int, error) {
	var total int
	err := db.QueryRow(`SELECT COALESCE(SUM(like_count), 0) FROM videos`).Scan(&total)
	return total, err
}

// GetRecentActivity returns recent user activity
func (db *DB) GetRecentActivity(limit int) ([]map[string]interface{}, error) {
	rows, err := db.Query(
		`SELECT ua.id, ua.user_id, ua.activity_type, ua.video_id, ua.metadata, ua.created_at, u.email, u.first_name, u.last_name, v.title as video_title
		 FROM user_activity ua
		 JOIN users u ON ua.user_id = u.id
		 LEFT JOIN videos v ON ua.video_id = v.id
		 ORDER BY ua.created_at DESC
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []map[string]interface{}
	for rows.Next() {
		var id, userID int
		var activityType, email, firstName, lastName string
		var videoID *int
		var metadata map[string]interface{}
		var createdAt time.Time
		var videoTitle *string

		err := rows.Scan(&id, &userID, &activityType, &videoID, &metadata, &createdAt, &email, &firstName, &lastName, &videoTitle)
		if err != nil {
			return nil, err
		}

		activity := map[string]interface{}{
			"id":            id,
			"user_id":       userID,
			"activity_type": activityType,
			"video_id":      videoID,
			"metadata":      metadata,
			"created_at":    createdAt,
			"user_email":    email,
			"user_name":     firstName + " " + lastName,
			"video_title":   videoTitle,
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

// GetSystemHealth returns system health information
func (db *DB) GetSystemHealth() (map[string]interface{}, error) {
	health := make(map[string]interface{})

	// Check database connection
	if err := db.Ping(); err != nil {
		health["database"] = "unhealthy"
		health["database_error"] = err.Error()
	} else {
		health["database"] = "healthy"
	}

	// Get basic stats
	userCount, _ := db.GetUserCount()
	videoCount, _ := db.GetVideoCount()
	totalViews, _ := db.GetTotalViews()
	totalLikes, _ := db.GetTotalLikes()

	health["stats"] = map[string]interface{}{
		"users":       userCount,
		"videos":      videoCount,
		"total_views": totalViews,
		"total_likes": totalLikes,
	}

	return health, nil
}
