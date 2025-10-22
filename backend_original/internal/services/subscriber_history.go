package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"bome-backend/internal/database"
)

// SubscriberHistoryService handles business logic for subscriber history
type SubscriberHistoryService struct {
	db *database.DB
}

// NewSubscriberHistoryService creates a new subscriber history service
func NewSubscriberHistoryService(db *database.DB) *SubscriberHistoryService {
	return &SubscriberHistoryService{db: db}
}

// GetSubscriberHistory retrieves complete subscriber history for a user
func (s *SubscriberHistoryService) GetSubscriberHistory(userID int) (map[string]interface{}, error) {
	log.Printf("Service: Getting subscriber history for user %d", userID)

	history, err := s.db.GetSubscriberHistoryEntries(userID)
	if err != nil {
		log.Printf("Service: Error getting subscriber history: %v", err)
		return nil, fmt.Errorf("failed to get subscriber history: %w", err)
	}

	log.Printf("Service: Successfully retrieved subscriber history for user %d", userID)
	return history, nil
}

// AddSubscriptionHistoryEntry adds a subscription history entry
func (s *SubscriberHistoryService) AddSubscriptionHistoryEntry(userID int, action, description string, metadata map[string]interface{}) error {
	log.Printf("Service: Adding subscription history entry for user %d: %s", userID, action)

	entry := &database.SubscriberHistoryEntry{
		Action:      action,
		Timestamp:   time.Now(),
		Description: description,
		Metadata:    metadata,
	}

	err := s.db.AddSubscriberHistoryEntry(userID, "subscription", entry)
	if err != nil {
		log.Printf("Service: Error adding subscription history entry: %v", err)
		return fmt.Errorf("failed to add subscription history entry: %w", err)
	}

	log.Printf("Service: Successfully added subscription history entry for user %d", userID)
	return nil
}

// AddOfferHistoryEntry adds an offer history entry
func (s *SubscriberHistoryService) AddOfferHistoryEntry(userID int, action, description string, metadata map[string]interface{}) error {
	log.Printf("Service: Adding offer history entry for user %d: %s", userID, action)

	entry := &database.SubscriberHistoryEntry{
		Action:      action,
		Timestamp:   time.Now(),
		Description: description,
		Metadata:    metadata,
	}

	err := s.db.AddSubscriberHistoryEntry(userID, "offer", entry)
	if err != nil {
		log.Printf("Service: Error adding offer history entry: %v", err)
		return fmt.Errorf("failed to add offer history entry: %w", err)
	}

	log.Printf("Service: Successfully added offer history entry for user %d", userID)
	return nil
}

// AddSuspensionHistoryEntry adds a suspension history entry
func (s *SubscriberHistoryService) AddSuspensionHistoryEntry(userID int, action, reason string, previousStatus, newStatus string) error {
	log.Printf("Service: Adding suspension history entry for user %d: %s", userID, action)

	metadata := map[string]interface{}{
		"reason":          reason,
		"previous_status": previousStatus,
		"new_status":      newStatus,
	}

	entry := &database.SubscriberHistoryEntry{
		Action:      action,
		Timestamp:   time.Now(),
		Description: fmt.Sprintf("Account %s: %s", action, reason),
		Metadata:    metadata,
	}

	err := s.db.AddSubscriberHistoryEntry(userID, "subscription", entry)
	if err != nil {
		log.Printf("Service: Error adding suspension history entry: %v", err)
		return fmt.Errorf("failed to add suspension history entry: %w", err)
	}

	// Also add to updates
	update := &database.SubscriberHistoryUpdate{
		Timestamp: time.Now(),
		Action:    action,
		Category:  "suspension",
		Details:   metadata,
	}

	err = s.db.AddSubscriberHistoryUpdate(userID, update)
	if err != nil {
		log.Printf("Service: Error adding suspension update: %v", err)
		return fmt.Errorf("failed to add suspension update: %w", err)
	}

	log.Printf("Service: Successfully added suspension history entry for user %d", userID)
	return nil
}

// AddAdminNote adds an admin note to subscriber history
func (s *SubscriberHistoryService) AddAdminNote(userID, adminID int, adminName, note, category string) error {
	log.Printf("Service: Adding admin note for user %d by admin %d", userID, adminID)

	noteEntry := &database.SubscriberHistoryNote{
		Timestamp:  time.Now(),
		AdminID:    adminID,
		AdminName:  adminName,
		Note:       note,
		Category:   category,
		Visibility: "admin_only",
	}

	err := s.db.AddSubscriberHistoryNote(userID, noteEntry)
	if err != nil {
		log.Printf("Service: Error adding admin note: %v", err)
		return fmt.Errorf("failed to add admin note: %w", err)
	}

	log.Printf("Service: Successfully added admin note for user %d", userID)
	return nil
}

// AddSystemNote adds a system note to subscriber history
func (s *SubscriberHistoryService) AddSystemNote(userID int, note, category string) error {
	log.Printf("Service: Adding system note for user %d", userID)

	noteEntry := &database.SubscriberHistoryNote{
		Timestamp:  time.Now(),
		Note:       note,
		Category:   category,
		Visibility: "system",
	}

	err := s.db.AddSubscriberHistoryNote(userID, noteEntry)
	if err != nil {
		log.Printf("Service: Error adding system note: %v", err)
		return fmt.Errorf("failed to add system note: %w", err)
	}

	log.Printf("Service: Successfully added system note for user %d", userID)
	return nil
}

// AddUserNote adds a user note to subscriber history
func (s *SubscriberHistoryService) AddUserNote(userID int, note, category string) error {
	log.Printf("Service: Adding user note for user %d", userID)

	noteEntry := &database.SubscriberHistoryNote{
		Timestamp:  time.Now(),
		Note:       note,
		Category:   category,
		Visibility: "user_visible",
	}

	err := s.db.AddSubscriberHistoryNote(userID, noteEntry)
	if err != nil {
		log.Printf("Service: Error adding user note: %v", err)
		return fmt.Errorf("failed to add user note: %w", err)
	}

	log.Printf("Service: Successfully added user note for user %d", userID)
	return nil
}

// GetSubscriberHistorySummary returns a summary of subscriber history
func (s *SubscriberHistoryService) GetSubscriberHistorySummary(userID int) (map[string]interface{}, error) {
	log.Printf("Service: Getting subscriber history summary for user %d", userID)

	history, err := s.db.GetSubscriberHistoryEntries(userID)
	if err != nil {
		log.Printf("Service: Error getting subscriber history: %v", err)
		return nil, fmt.Errorf("failed to get subscriber history: %w", err)
	}

	// Create summary
	summary := map[string]interface{}{
		"user_id": userID,
		"summary": map[string]interface{}{
			"total_subscription_events": 0,
			"total_offer_events":        0,
			"total_notes":               0,
			"total_updates":             0,
			"last_activity":             nil,
		},
		"history": history,
	}

	// Count events
	if subHistory, ok := history["subscription_history"].(map[string]interface{}); ok {
		if entries, ok := subHistory["entries"].([]interface{}); ok {
			summary["summary"].(map[string]interface{})["total_subscription_events"] = len(entries)
		}
	}

	if offHistory, ok := history["offer_history"].(map[string]interface{}); ok {
		if entries, ok := offHistory["entries"].([]interface{}); ok {
			summary["summary"].(map[string]interface{})["total_offer_events"] = len(entries)
		}
	}

	if notes, ok := history["notes"].(map[string]interface{}); ok {
		if noteEntries, ok := notes["notes"].([]interface{}); ok {
			summary["summary"].(map[string]interface{})["total_notes"] = len(noteEntries)
		}
	}

	if updates, ok := history["updates"].(map[string]interface{}); ok {
		if updateEntries, ok := updates["updates"].([]interface{}); ok {
			summary["summary"].(map[string]interface{})["total_updates"] = len(updateEntries)
		}
	}

	log.Printf("Service: Successfully created subscriber history summary for user %d", userID)
	return summary, nil
}

// ExportSubscriberHistory exports subscriber history as JSON
func (s *SubscriberHistoryService) ExportSubscriberHistory(userID int) ([]byte, error) {
	log.Printf("Service: Exporting subscriber history for user %d", userID)

	history, err := s.db.GetSubscriberHistoryEntries(userID)
	if err != nil {
		log.Printf("Service: Error getting subscriber history for export: %v", err)
		return nil, fmt.Errorf("failed to get subscriber history for export: %w", err)
	}

	// Add export metadata
	exportData := map[string]interface{}{
		"export_timestamp": time.Now().Format(time.RFC3339),
		"user_id":          userID,
		"history":          history,
	}

	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		log.Printf("Service: Error marshaling export data: %v", err)
		return nil, fmt.Errorf("failed to marshal export data: %w", err)
	}

	log.Printf("Service: Successfully exported subscriber history for user %d", userID)
	return jsonData, nil
}
