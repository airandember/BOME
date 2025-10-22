package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// SubscriberHistory represents the subscriber_history table
type SubscriberHistory struct {
	ID          int            `json:"id"`
	UserID      int            `json:"user_id"`
	UsrSubHstry sql.NullString `json:"usr_sub_hstry"` // JSONB
	UsrOffHstry sql.NullString `json:"usr_off_hstry"` // JSONB
	UpdatedAt   sql.NullString `json:"updated_at"`    // JSONB
	Notes       sql.NullString `json:"notes"`         // JSONB
	CreatedAt   time.Time      `json:"created_at"`
}

// SubscriberHistoryEntry represents a single history entry
type SubscriberHistoryEntry struct {
	Action      string                 `json:"action"`
	Timestamp   time.Time              `json:"timestamp"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SubscriberHistoryUpdate represents an update entry
type SubscriberHistoryUpdate struct {
	Timestamp time.Time              `json:"timestamp"`
	Action    string                 `json:"action"`
	Category  string                 `json:"category"`
	Details   map[string]interface{} `json:"details"`
}

// SubscriberHistoryNote represents a note entry
type SubscriberHistoryNote struct {
	Timestamp  time.Time `json:"timestamp"`
	AdminID    int       `json:"admin_id,omitempty"`
	AdminName  string    `json:"admin_name,omitempty"`
	Note       string    `json:"note"`
	Category   string    `json:"category"`
	Visibility string    `json:"visibility"`
}

// CreateSubscriberHistory creates a new subscriber history record
func (db *DB) CreateSubscriberHistory(history *SubscriberHistory) error {
	query := `
		INSERT INTO subscriber_history (user_id, usr_sub_hstry, usr_off_hstry, updated_at, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id
	`

	var id int
	err := db.QueryRow(
		query,
		history.UserID,
		history.UsrSubHstry,
		history.UsrOffHstry,
		history.UpdatedAt,
		history.Notes,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to create subscriber history: %w", err)
	}

	history.ID = id
	log.Printf("Database: Created subscriber history record for user %d with ID %d", history.UserID, id)
	return nil
}

// GetSubscriberHistory retrieves subscriber history for a specific user
func (db *DB) GetSubscriberHistory(userID int) (*SubscriberHistory, error) {
	query := `
		SELECT id, user_id, usr_sub_hstry, usr_off_hstry, updated_at, notes, created_at
		FROM subscriber_history
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	history := &SubscriberHistory{}
	err := db.QueryRow(query, userID).Scan(
		&history.ID,
		&history.UserID,
		&history.UsrSubHstry,
		&history.UsrOffHstry,
		&history.UpdatedAt,
		&history.Notes,
		&history.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Create a new history record if none exists
			history = &SubscriberHistory{
				UserID: userID,
			}
			err = db.CreateSubscriberHistory(history)
			if err != nil {
				return nil, fmt.Errorf("failed to create new subscriber history: %w", err)
			}
			return history, nil
		}
		return nil, fmt.Errorf("failed to get subscriber history: %w", err)
	}

	log.Printf("Database: Retrieved subscriber history for user %d", userID)
	return history, nil
}

// UpdateSubscriberHistory updates an existing subscriber history record
func (db *DB) UpdateSubscriberHistory(history *SubscriberHistory) error {
	query := `
		UPDATE subscriber_history
		SET usr_sub_hstry = $2, usr_off_hstry = $3, updated_at = $4, notes = $5
		WHERE user_id = $1
	`

	result, err := db.Exec(
		query,
		history.UserID,
		history.UsrSubHstry,
		history.UsrOffHstry,
		history.UpdatedAt,
		history.Notes,
	)

	if err != nil {
		return fmt.Errorf("failed to update subscriber history: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		// Create new record if none exists
		return db.CreateSubscriberHistory(history)
	}

	log.Printf("Database: Updated subscriber history for user %d", history.UserID)
	return nil
}

// AddSubscriberHistoryEntry adds a new entry to the subscriber history
func (db *DB) AddSubscriberHistoryEntry(userID int, entryType string, entry *SubscriberHistoryEntry) error {
	history, err := db.GetSubscriberHistory(userID)
	if err != nil {
		return fmt.Errorf("failed to get subscriber history: %w", err)
	}

	// Parse existing JSON or create new structure
	var existingData map[string]interface{}
	if entryType == "subscription" {
		if history.UsrSubHstry.Valid && history.UsrSubHstry.String != "" {
			err = json.Unmarshal([]byte(history.UsrSubHstry.String), &existingData)
			if err != nil {
				existingData = make(map[string]interface{})
			}
		} else {
			existingData = make(map[string]interface{})
		}
	} else if entryType == "offer" {
		if history.UsrOffHstry.Valid && history.UsrOffHstry.String != "" {
			err = json.Unmarshal([]byte(history.UsrOffHstry.String), &existingData)
			if err != nil {
				existingData = make(map[string]interface{})
			}
		} else {
			existingData = make(map[string]interface{})
		}
	}

	// Add new entry
	if existingData["entries"] == nil {
		existingData["entries"] = []interface{}{}
	}
	entries := existingData["entries"].([]interface{})
	entries = append(entries, entry)
	existingData["entries"] = entries

	// Convert back to JSON
	jsonData, err := json.Marshal(existingData)
	if err != nil {
		return fmt.Errorf("failed to marshal history data: %w", err)
	}

	// Update the appropriate field
	if entryType == "subscription" {
		history.UsrSubHstry = sql.NullString{String: string(jsonData), Valid: true}
	} else if entryType == "offer" {
		history.UsrOffHstry = sql.NullString{String: string(jsonData), Valid: true}
	}

	return db.UpdateSubscriberHistory(history)
}

// AddSubscriberHistoryUpdate adds a new update entry
func (db *DB) AddSubscriberHistoryUpdate(userID int, update *SubscriberHistoryUpdate) error {
	history, err := db.GetSubscriberHistory(userID)
	if err != nil {
		return fmt.Errorf("failed to get subscriber history: %w", err)
	}

	// Parse existing updates or create new structure
	var existingUpdates map[string]interface{}
	if history.UpdatedAt.Valid && history.UpdatedAt.String != "" {
		err = json.Unmarshal([]byte(history.UpdatedAt.String), &existingUpdates)
		if err != nil {
			existingUpdates = make(map[string]interface{})
		}
	} else {
		existingUpdates = make(map[string]interface{})
	}

	// Add new update
	if existingUpdates["updates"] == nil {
		existingUpdates["updates"] = []interface{}{}
	}
	updates := existingUpdates["updates"].([]interface{})
	updates = append(updates, update)
	existingUpdates["updates"] = updates

	// Convert back to JSON
	jsonData, err := json.Marshal(existingUpdates)
	if err != nil {
		return fmt.Errorf("failed to marshal update data: %w", err)
	}

	history.UpdatedAt = sql.NullString{String: string(jsonData), Valid: true}
	return db.UpdateSubscriberHistory(history)
}

// AddSubscriberHistoryNote adds a new note entry
func (db *DB) AddSubscriberHistoryNote(userID int, note *SubscriberHistoryNote) error {
	history, err := db.GetSubscriberHistory(userID)
	if err != nil {
		return fmt.Errorf("failed to get subscriber history: %w", err)
	}

	// Parse existing notes or create new structure
	var existingNotes map[string]interface{}
	if history.Notes.Valid && history.Notes.String != "" {
		err = json.Unmarshal([]byte(history.Notes.String), &existingNotes)
		if err != nil {
			existingNotes = make(map[string]interface{})
		}
	} else {
		existingNotes = make(map[string]interface{})
	}

	// Add new note
	if existingNotes["notes"] == nil {
		existingNotes["notes"] = []interface{}{}
	}
	notes := existingNotes["notes"].([]interface{})
	notes = append(notes, note)
	existingNotes["notes"] = notes

	// Convert back to JSON
	jsonData, err := json.Marshal(existingNotes)
	if err != nil {
		return fmt.Errorf("failed to marshal notes data: %w", err)
	}

	history.Notes = sql.NullString{String: string(jsonData), Valid: true}
	return db.UpdateSubscriberHistory(history)
}

// GetSubscriberHistoryEntries retrieves all history entries for a user
func (db *DB) GetSubscriberHistoryEntries(userID int) (map[string]interface{}, error) {
	history, err := db.GetSubscriberHistory(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber history: %w", err)
	}

	result := make(map[string]interface{})

	// Parse subscription history
	if history.UsrSubHstry.Valid && history.UsrSubHstry.String != "" {
		var subHistory map[string]interface{}
		err = json.Unmarshal([]byte(history.UsrSubHstry.String), &subHistory)
		if err == nil {
			result["subscription_history"] = subHistory
		}
	}

	// Parse offer history
	if history.UsrOffHstry.Valid && history.UsrOffHstry.String != "" {
		var offHistory map[string]interface{}
		err = json.Unmarshal([]byte(history.UsrOffHstry.String), &offHistory)
		if err == nil {
			result["offer_history"] = offHistory
		}
	}

	// Parse updates
	if history.UpdatedAt.Valid && history.UpdatedAt.String != "" {
		var updates map[string]interface{}
		err = json.Unmarshal([]byte(history.UpdatedAt.String), &updates)
		if err == nil {
			result["updates"] = updates
		}
	}

	// Parse notes
	if history.Notes.Valid && history.Notes.String != "" {
		var notes map[string]interface{}
		err = json.Unmarshal([]byte(history.Notes.String), &notes)
		if err == nil {
			result["notes"] = notes
		}
	}

	return result, nil
}
