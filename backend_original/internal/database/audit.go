package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// AuditLog represents a security audit log entry
type AuditLog struct {
	ID         int                    `json:"id"`
	UserID     *int                   `json:"user_id,omitempty"`
	UserEmail  *string                `json:"user_email,omitempty"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	ResourceID *string                `json:"resource_id,omitempty"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	Status     string                 `json:"status"` // success, failed, warning
	Details    *string                `json:"details,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Severity   string                 `json:"severity"` // low, medium, high, critical
	CreatedAt  time.Time              `json:"created_at"`
}

// CreateAuditLog creates a new audit log entry
func (db *DB) CreateAuditLog(log *AuditLog) error {
	metadataJSON := "{}"
	if log.Metadata != nil {
		if metadataBytes, err := json.Marshal(log.Metadata); err == nil {
			metadataJSON = string(metadataBytes)
		}
	}

	query := `
		INSERT INTO audit_logs (user_id, user_email, action, resource, resource_id, ip_address, user_agent, status, details, metadata, severity, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
		RETURNING id
	`

	var id int
	err := db.QueryRow(
		query,
		log.UserID,
		log.UserEmail,
		log.Action,
		log.Resource,
		log.ResourceID,
		log.IPAddress,
		log.UserAgent,
		log.Status,
		log.Details,
		metadataJSON,
		log.Severity,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	log.ID = id
	return nil
}

// GetAuditLogs retrieves audit logs with filtering
func (db *DB) GetAuditLogs(filters map[string]interface{}, limit, offset int) ([]*AuditLog, error) {
	query := `
		SELECT id, user_id, user_email, action, resource, resource_id, ip_address, user_agent, status, details, metadata, severity, created_at
		FROM audit_logs
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	// Add filters
	if userID, ok := filters["user_id"]; ok {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}

	if action, ok := filters["action"]; ok {
		query += fmt.Sprintf(" AND action = $%d", argIndex)
		args = append(args, action)
		argIndex++
	}

	if status, ok := filters["status"]; ok {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	if severity, ok := filters["severity"]; ok {
		query += fmt.Sprintf(" AND severity = $%d", argIndex)
		args = append(args, severity)
		argIndex++
	}

	if startDate, ok := filters["start_date"]; ok {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, startDate)
		argIndex++
	}

	if endDate, ok := filters["end_date"]; ok {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, endDate)
		argIndex++
	}

	query += " ORDER BY created_at DESC LIMIT $%d OFFSET $%d"
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*AuditLog
	for rows.Next() {
		log := &AuditLog{}
		var metadataJSON sql.NullString
		var userID sql.NullInt64
		var userEmail sql.NullString
		var resourceID sql.NullString
		var details sql.NullString

		err := rows.Scan(
			&log.ID,
			&userID,
			&userEmail,
			&log.Action,
			&log.Resource,
			&resourceID,
			&log.IPAddress,
			&log.UserAgent,
			&log.Status,
			&details,
			&metadataJSON,
			&log.Severity,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		// Handle nullable fields
		if userID.Valid {
			userIDInt := int(userID.Int64)
			log.UserID = &userIDInt
		}
		if userEmail.Valid {
			log.UserEmail = &userEmail.String
		}
		if resourceID.Valid {
			log.ResourceID = &resourceID.String
		}
		if details.Valid {
			log.Details = &details.String
		}
		if metadataJSON.Valid {
			json.Unmarshal([]byte(metadataJSON.String), &log.Metadata)
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// GetAuditLogCount returns the total count of audit logs matching filters
func (db *DB) GetAuditLogCount(filters map[string]interface{}) (int, error) {
	query := "SELECT COUNT(*) FROM audit_logs WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	// Add filters (same logic as GetAuditLogs)
	if userID, ok := filters["user_id"]; ok {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}

	if action, ok := filters["action"]; ok {
		query += fmt.Sprintf(" AND action = $%d", argIndex)
		args = append(args, action)
		argIndex++
	}

	if status, ok := filters["status"]; ok {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	if severity, ok := filters["severity"]; ok {
		query += fmt.Sprintf(" AND severity = $%d", argIndex)
		args = append(args, severity)
		argIndex++
	}

	if startDate, ok := filters["start_date"]; ok {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, startDate)
		argIndex++
	}

	if endDate, ok := filters["end_date"]; ok {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, endDate)
		argIndex++
	}

	var count int
	err := db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get audit log count: %w", err)
	}

	return count, nil
}

// CleanupOldAuditLogs removes audit logs older than the specified duration
func (db *DB) CleanupOldAuditLogs(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	_, err := db.Exec("DELETE FROM audit_logs WHERE created_at < $1", cutoff)
	if err != nil {
		return fmt.Errorf("failed to cleanup old audit logs: %w", err)
	}
	return nil
}

// GetSecurityMetrics returns security-related metrics
func (db *DB) GetSecurityMetrics() (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// Failed login attempts in last 24 hours
	var failedLogins int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM audit_logs 
		WHERE action = 'login' AND status = 'failed' 
		AND created_at > NOW() - INTERVAL '24 hours'
	`).Scan(&failedLogins)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed login count: %w", err)
	}
	metrics["failed_logins_24h"] = failedLogins

	// Suspicious activities in last 24 hours
	var suspiciousActivities int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM audit_logs 
		WHERE severity IN ('high', 'critical') 
		AND created_at > NOW() - INTERVAL '24 hours'
	`).Scan(&suspiciousActivities)
	if err != nil {
		return nil, fmt.Errorf("failed to get suspicious activities count: %w", err)
	}
	metrics["suspicious_activities_24h"] = suspiciousActivities

	// Total audit logs in last 7 days
	var totalLogs int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM audit_logs 
		WHERE created_at > NOW() - INTERVAL '7 days'
	`).Scan(&totalLogs)
	if err != nil {
		return nil, fmt.Errorf("failed to get total logs count: %w", err)
	}
	metrics["total_logs_7d"] = totalLogs

	return metrics, nil
}

// CreateAdminLog creates an admin action log with extended information
func (db *DB) CreateAdminLog(adminID *int, action, resourceType string, resourceID *int, metadata map[string]interface{}, ipAddress, userAgent string) error {
	var resourceIDStr *string
	if resourceID != nil {
		str := fmt.Sprintf("%d", *resourceID)
		resourceIDStr = &str
	}

	return db.CreateAuditLog(&AuditLog{
		UserID:     adminID,
		Action:     action,
		Resource:   resourceType,
		ResourceID: resourceIDStr, // Convert *int to *string
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Status:     "success",
		Metadata:   metadata,
		Severity:   "medium",
	})
}
