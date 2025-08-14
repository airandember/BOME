package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// UserData represents the user information stored in JSONB format
type UserData struct {
	ID            interface{} `json:"id"` // Can be int or string
	Email         string      `json:"email"`
	Role          string      `json:"role"`
	FirstName     string      `json:"first_name"`
	LastName      string      `json:"last_name"`
	EmailVerified bool        `json:"email_verified,omitempty"`
	StoredAt      int64       `json:"stored_at,omitempty"`
}

// PlanHistoryEvent represents a history event in the database
type PlanHistoryEvent struct {
	ID          int
	PlanID      int
	EventType   string
	Timestamp   time.Time
	UserID      sql.NullString // JSONB data as string
	Description sql.NullString
	OldValues   sql.NullString
	NewValues   sql.NullString
	Metadata    sql.NullString
	CreatedAt   time.Time
}

// GetUserDataFromHistory extracts UserData from the JSONB user_id field
func (event *PlanHistoryEvent) GetUserData() (*UserData, error) {
	if !event.UserID.Valid || event.UserID.String == "" {
		return nil, nil
	}

	var userData UserData
	err := json.Unmarshal([]byte(event.UserID.String), &userData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	return &userData, nil
}

// AddHistoryEvent adds a new history event to the subscription_plans_history table
func (db *DB) AddHistoryEvent(planID int, eventType, userID, description string, oldValues, newValues, metadata map[string]interface{}) error {
	log.Printf("Database: Adding history event for plan %d, type: %s", planID, eventType)
	log.Printf("Database: UserID received: %s", userID)

	// Convert maps to JSON strings
	oldValuesJSON, err := json.Marshal(oldValues)
	if err != nil {
		return fmt.Errorf("failed to marshal old values: %w", err)
	}

	newValuesJSON, err := json.Marshal(newValues)
	if err != nil {
		return fmt.Errorf("failed to marshal new values: %w", err)
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Use userID directly (it's already JSON from localStorage)
	var userDataJSON string
	if userID == "system" || userID == "System" {
		log.Printf("Database: Creating system user data")
		userData := UserData{
			ID:        "system",
			Email:     "system",
			Role:      "system",
			FirstName: "System",
			LastName:  "",
		}
		userDataBytes, err := json.Marshal(userData)
		if err != nil {
			return fmt.Errorf("failed to marshal system user data: %w", err)
		}
		userDataJSON = string(userDataBytes)
		log.Printf("Database: System user data created: %s", userDataJSON)
	} else if userID == "System (Auto-Expiration)" {
		log.Printf("Database: Creating auto-expiration user data")
		userData := UserData{
			ID:        "system",
			Email:     "system",
			Role:      "system",
			FirstName: "System",
			LastName:  "(Auto-Expiration)",
		}
		userDataBytes, err := json.Marshal(userData)
		if err != nil {
			return fmt.Errorf("failed to marshal auto-expiration user data: %w", err)
		}
		userDataJSON = string(userDataBytes)
		log.Printf("Database: Auto-expiration user data created: %s", userDataJSON)
	} else {
		// userID is already JSON from localStorage, use it directly
		log.Printf("Database: Using localStorage user data directly: %s", userID)
		userDataJSON = userID
	}

	log.Printf("Database: Final userDataJSON for insertion: %s", userDataJSON)

	query := `
		INSERT INTO subscription_plans_history (
			plan_id, event_type, timestamp, user_id, description, 
			old_values, new_values, metadata, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	_, err = db.Exec(query,
		planID, eventType, time.Now(), userDataJSON, description,
		string(oldValuesJSON), string(newValuesJSON), string(metadataJSON), time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to insert history event: %w", err)
	}

	log.Printf("Database: Successfully added history event for plan %d", planID)
	return nil
}

// GetPlanHistory retrieves all history events for a specific plan
func (db *DB) GetPlanHistory(planID int) ([]*PlanHistoryEvent, error) {
	log.Printf("Database: Getting history for plan %d", planID)

	query := `
		SELECT id, plan_id, event_type, timestamp, user_id, description, 
		       old_values, new_values, metadata, created_at
		FROM subscription_plans_history 
		WHERE plan_id = $1 
		ORDER BY timestamp DESC
	`

	rows, err := db.Query(query, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to query plan history: %w", err)
	}
	defer rows.Close()

	var events []*PlanHistoryEvent
	for rows.Next() {
		event := &PlanHistoryEvent{}
		err := rows.Scan(
			&event.ID, &event.PlanID, &event.EventType, &event.Timestamp,
			&event.UserID, &event.Description, &event.OldValues, &event.NewValues,
			&event.Metadata, &event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history event: %w", err)
		}
		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating history events: %w", err)
	}

	log.Printf("Database: Retrieved %d history events for plan %d", len(events), planID)
	return events, nil
}

// GetHistoryEventsByType retrieves history events filtered by event type
func (db *DB) GetHistoryEventsByType(eventType string, limit int) ([]*PlanHistoryEvent, error) {
	log.Printf("Database: Getting history events by type: %s", eventType)

	query := `
		SELECT id, plan_id, event_type, timestamp, user_id, description, 
		       old_values, new_values, metadata, created_at
		FROM subscription_plans_history 
		WHERE event_type = $1 
		ORDER BY timestamp DESC
		LIMIT $2
	`

	rows, err := db.Query(query, eventType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query history by type: %w", err)
	}
	defer rows.Close()

	var events []*PlanHistoryEvent
	for rows.Next() {
		event := &PlanHistoryEvent{}
		err := rows.Scan(
			&event.ID, &event.PlanID, &event.EventType, &event.Timestamp,
			&event.UserID, &event.Description, &event.OldValues, &event.NewValues,
			&event.Metadata, &event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history event: %w", err)
		}
		events = append(events, event)
	}

	log.Printf("Database: Retrieved %d history events of type %s", len(events), eventType)
	return events, nil
}

// GetHistoryEventsByUser retrieves history events filtered by user
func (db *DB) GetHistoryEventsByUser(userID string, limit int) ([]*PlanHistoryEvent, error) {
	log.Printf("Database: Getting history events by user: %s", userID)

	query := `
		SELECT id, plan_id, event_type, timestamp, user_id, description, 
		       old_values, new_values, metadata, created_at
		FROM subscription_plans_history 
		WHERE user_id = $1 
		ORDER BY timestamp DESC
		LIMIT $2
	`

	rows, err := db.Query(query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query history by user: %w", err)
	}
	defer rows.Close()

	var events []*PlanHistoryEvent
	for rows.Next() {
		event := &PlanHistoryEvent{}
		err := rows.Scan(
			&event.ID, &event.PlanID, &event.EventType, &event.Timestamp,
			&event.UserID, &event.Description, &event.OldValues, &event.NewValues,
			&event.Metadata, &event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history event: %w", err)
		}
		events = append(events, event)
	}

	log.Printf("Database: Retrieved %d history events for user %s", len(events), userID)
	return events, nil
}

// GetHistoryEventsByDateRange retrieves history events within a date range
func (db *DB) GetHistoryEventsByDateRange(startDate, endDate time.Time, limit int) ([]*PlanHistoryEvent, error) {
	log.Printf("Database: Getting history events from %s to %s", startDate, endDate)

	query := `
		SELECT id, plan_id, event_type, timestamp, user_id, description, 
		       old_values, new_values, metadata, created_at
		FROM subscription_plans_history 
		WHERE timestamp BETWEEN $1 AND $2
		ORDER BY timestamp DESC
		LIMIT $3
	`

	rows, err := db.Query(query, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query history by date range: %w", err)
	}
	defer rows.Close()

	var events []*PlanHistoryEvent
	for rows.Next() {
		event := &PlanHistoryEvent{}
		err := rows.Scan(
			&event.ID, &event.PlanID, &event.EventType, &event.Timestamp,
			&event.UserID, &event.Description, &event.OldValues, &event.NewValues,
			&event.Metadata, &event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history event: %w", err)
		}
		events = append(events, event)
	}

	log.Printf("Database: Retrieved %d history events in date range", len(events))
	return events, nil
}

// GetHistoryStats retrieves statistics about history events
func (db *DB) GetHistoryStats() (map[string]interface{}, error) {
	log.Printf("Database: Getting history statistics")

	// Get total count
	var totalCount int
	err := db.QueryRow("SELECT COUNT(*) FROM subscription_plans_history").Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get count by event type
	query := `
		SELECT event_type, COUNT(*) 
		FROM subscription_plans_history 
		GROUP BY event_type
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get event type counts: %w", err)
	}
	defer rows.Close()

	eventTypeCounts := make(map[string]int)
	for rows.Next() {
		var eventType string
		var count int
		err := rows.Scan(&eventType, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event type count: %w", err)
		}
		eventTypeCounts[eventType] = count
	}

	// Get recent activity (last 7 days)
	var recentCount int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM subscription_plans_history 
		WHERE timestamp > NOW() - INTERVAL '7 days'
	`).Scan(&recentCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent count: %w", err)
	}

	stats := map[string]interface{}{
		"total_events":      totalCount,
		"recent_events":     recentCount,
		"event_type_counts": eventTypeCounts,
	}

	log.Printf("Database: Retrieved history statistics: %+v", stats)
	return stats, nil
}
