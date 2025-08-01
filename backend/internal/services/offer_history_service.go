package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"bome-backend/internal/database"
)

// OfferHistoryEvent represents an offer history event
type OfferHistoryEvent struct {
	ID          string                 `json:"id"`
	OfferID     int                    `json:"offer_id"`
	EventType   string                 `json:"event_type"`
	Timestamp   time.Time              `json:"timestamp"`
	UserID      string                 `json:"user_id"`
	Description string                 `json:"description"`
	OldValues   map[string]interface{} `json:"old_values,omitempty"`
	NewValues   map[string]interface{} `json:"new_values,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// OfferHistoryService handles offer history operations
type OfferHistoryService struct {
	db *database.DB
}

// NewOfferHistoryService creates a new offer history service
func NewOfferHistoryService(db *database.DB) *OfferHistoryService {
	return &OfferHistoryService{
		db: db,
	}
}

// AddHistoryEvent adds a new history event for an offer
func (s *OfferHistoryService) AddHistoryEvent(offerID int, eventType, description, userDataJSON string, oldOffer *database.SubscriptionOffer, metadata map[string]interface{}) error {
	log.Printf("AddHistoryEvent: Starting to add event for offer %d", offerID)
	log.Printf("AddHistoryEvent: Event details - Type=%s, Description=%s, UserID=%s", eventType, description, userDataJSON)

	// Create history event
	event := &database.SubscriptionOfferHistory{
		OfferID:   offerID,
		EventType: eventType,
		Accepted:  false, // Default to false, will be set based on event type
	}

	// Handle user data
	if userDataJSON != "" {
		event.UserID = sql.NullString{String: userDataJSON, Valid: true}
	}

	// Set accepted based on event type
	switch eventType {
	case "offer_accepted", "offer_applied":
		event.Accepted = true
	case "offer_created", "offer_updated", "offer_deleted", "offer_viewed", "offer_expired":
		event.Accepted = false
	}

	// Add metadata
	if metadata != nil {
		metadataJSON, _ := json.Marshal(metadata)
		event.DeviceInfo = sql.NullString{String: string(metadataJSON), Valid: true}
	}

	// Add to database
	err := s.db.AddOfferHistoryEvent(event)
	if err != nil {
		log.Printf("AddHistoryEvent: Error adding history event: %v", err)
		return err
	}

	log.Printf("AddHistoryEvent: Successfully added history event for offer %d", offerID)
	return nil
}

// GetOfferHistory retrieves history events for a specific offer
func (s *OfferHistoryService) GetOfferHistory(offerID int) ([]*database.SubscriptionOfferHistory, error) {
	log.Printf("GetOfferHistory: Getting history for offer %d", offerID)

	events, err := s.db.GetOfferHistory(offerID)
	if err != nil {
		log.Printf("GetOfferHistory: Error getting offer history: %v", err)
		return nil, err
	}

	log.Printf("GetOfferHistory: Retrieved %d history events for offer %d", len(events), offerID)
	return events, nil
}

// CreateOfferCreatedEvent creates a history event for offer creation
func (s *OfferHistoryService) CreateOfferCreatedEvent(offer *database.SubscriptionOffer, userDataJSON string) OfferHistoryEvent {
	eventID := fmt.Sprintf("evt_%d", time.Now().Unix())

	event := OfferHistoryEvent{
		ID:          eventID,
		OfferID:     offer.ID,
		EventType:   "offer_created",
		Timestamp:   time.Now(),
		UserID:      userDataJSON,
		Description: fmt.Sprintf("Offer '%s' was created", offer.OffName),
		NewValues: map[string]interface{}{
			"id":                 offer.ID,
			"off_name":           offer.OffName,
			"off_discount_type":  offer.OffDiscountType,
			"off_discount_value": offer.OffDiscountValue,
			"is_active":          offer.IsActive,
			"off_priority":       offer.OffPriority,
			"off_auto_apply":     offer.OffAutoApply,
		},
		Metadata: map[string]interface{}{
			"action":     "created",
			"offer_name": offer.OffName,
		},
	}
	return event
}

// CreateOfferUpdatedEvent creates a history event for offer updates
func (s *OfferHistoryService) CreateOfferUpdatedEvent(offer *database.SubscriptionOffer, oldOffer *database.SubscriptionOffer, userDataJSON string) OfferHistoryEvent {
	eventID := fmt.Sprintf("evt_%d", time.Now().Unix())

	event := OfferHistoryEvent{
		ID:          eventID,
		OfferID:     offer.ID,
		EventType:   "offer_updated",
		Timestamp:   time.Now(),
		UserID:      userDataJSON,
		Description: fmt.Sprintf("Offer '%s' was updated", offer.OffName),
		NewValues: map[string]interface{}{
			"id":                 offer.ID,
			"off_name":           offer.OffName,
			"off_discount_type":  offer.OffDiscountType,
			"off_discount_value": offer.OffDiscountValue,
			"is_active":          offer.IsActive,
			"off_priority":       offer.OffPriority,
			"off_auto_apply":     offer.OffAutoApply,
		},
		Metadata: map[string]interface{}{
			"action":     "updated",
			"offer_name": offer.OffName,
		},
	}

	// Add old values if available
	if oldOffer != nil {
		event.OldValues = map[string]interface{}{
			"id":                 oldOffer.ID,
			"off_name":           oldOffer.OffName,
			"off_discount_type":  oldOffer.OffDiscountType,
			"off_discount_value": oldOffer.OffDiscountValue,
			"is_active":          oldOffer.IsActive,
			"off_priority":       oldOffer.OffPriority,
			"off_auto_apply":     oldOffer.OffAutoApply,
		}
	}

	return event
}

// CreateOfferDeletedEvent creates a history event for offer deletion
func (s *OfferHistoryService) CreateOfferDeletedEvent(offer *database.SubscriptionOffer, userDataJSON string) OfferHistoryEvent {
	eventID := fmt.Sprintf("evt_%d", time.Now().Unix())

	event := OfferHistoryEvent{
		ID:          eventID,
		OfferID:     offer.ID,
		EventType:   "offer_deleted",
		Timestamp:   time.Now(),
		UserID:      userDataJSON,
		Description: fmt.Sprintf("Offer '%s' was deleted", offer.OffName),
		OldValues: map[string]interface{}{
			"id":                 offer.ID,
			"off_name":           offer.OffName,
			"off_discount_type":  offer.OffDiscountType,
			"off_discount_value": offer.OffDiscountValue,
			"is_active":          offer.IsActive,
			"off_priority":       offer.OffPriority,
			"off_auto_apply":     offer.OffAutoApply,
		},
		Metadata: map[string]interface{}{
			"action":     "deleted",
			"offer_name": offer.OffName,
		},
	}
	return event
}

// CreateOfferAcceptedEvent creates a history event for offer acceptance
func (s *OfferHistoryService) CreateOfferAcceptedEvent(offer *database.SubscriptionOffer, userDataJSON string, discountAmount, originalPrice, finalPrice float64) OfferHistoryEvent {
	eventID := fmt.Sprintf("evt_%d", time.Now().Unix())

	event := OfferHistoryEvent{
		ID:          eventID,
		OfferID:     offer.ID,
		EventType:   "offer_accepted",
		Timestamp:   time.Now(),
		UserID:      userDataJSON,
		Description: fmt.Sprintf("Offer '%s' was accepted by user", offer.OffName),
		NewValues: map[string]interface{}{
			"discount_amount": discountAmount,
			"original_price":  originalPrice,
			"final_price":     finalPrice,
		},
		Metadata: map[string]interface{}{
			"action":        "accepted",
			"offer_name":    offer.OffName,
			"discount_type": offer.OffDiscountType,
		},
	}
	return event
}

// CreateOfferViewedEvent creates a history event for offer viewing
func (s *OfferHistoryService) CreateOfferViewedEvent(offer *database.SubscriptionOffer, userDataJSON string, sessionID, referrerURL, userAgent string) OfferHistoryEvent {
	eventID := fmt.Sprintf("evt_%d", time.Now().Unix())

	event := OfferHistoryEvent{
		ID:          eventID,
		OfferID:     offer.ID,
		EventType:   "offer_viewed",
		Timestamp:   time.Now(),
		UserID:      userDataJSON,
		Description: fmt.Sprintf("Offer '%s' was viewed by user", offer.OffName),
		Metadata: map[string]interface{}{
			"action":       "viewed",
			"offer_name":   offer.OffName,
			"session_id":   sessionID,
			"referrer_url": referrerURL,
			"user_agent":   userAgent,
		},
	}
	return event
}
