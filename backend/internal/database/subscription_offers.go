package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// SubscriptionOffer represents a subscription offer in the database
type SubscriptionOffer struct {
	ID                    int            `json:"id"`
	PlanID                int            `json:"plan_id"`
	ItemID                sql.NullString `json:"item_id"`
	OffDiscountType       string         `json:"off_discount_type"`
	OffDiscountValue      float64        `json:"off_discount_value"`
	OfferStartDate        sql.NullTime   `json:"offer_start_date"`
	OffEndDate            sql.NullTime   `json:"off_end_date"`
	IsActive              bool           `json:"is_active"`
	OffDescription        sql.NullString `json:"off_description"`
	OffCreatedAt          time.Time      `json:"off_created_at"`
	OffUpdatedAt          time.Time      `json:"off_updated_at"`
	OffName               string         `json:"off_name"`
	OffCode               sql.NullString `json:"off_code"`
	OffMaxUses            sql.NullInt32  `json:"off_max_uses"`
	OffCurrentUses        int            `json:"off_current_uses"`
	OffTermsConditions    sql.NullString `json:"off_terms_conditions"`
	OffTarget             sql.NullString `json:"off_target"`
	OffPriority           int            `json:"off_priority"`
	OffAutoApply          bool           `json:"off_auto_apply"`
	StripeCouponID        sql.NullString `json:"stripe_coupon_id"`
	StripePromotionCodeID sql.NullString `json:"stripe_promotion_code_id"`
}

// SubscriptionOfferHistory represents a history event for subscription offers
type SubscriptionOfferHistory struct {
	ID             int             `json:"id"`
	OfferID        int             `json:"offer_id"`
	UserID         sql.NullString  `json:"user_id"`
	SubPlanID      sql.NullInt32   `json:"sub_plan_id"`
	Accepted       bool            `json:"accepted"`
	OffUserIP      sql.NullString  `json:"off_user_ip"`
	DeviceInfo     sql.NullString  `json:"device_info"`
	EventType      string          `json:"event_type"`
	DiscountAmount sql.NullFloat64 `json:"discount_amount"`
	OriginalPrice  sql.NullFloat64 `json:"original_price"`
	FinalPrice     sql.NullFloat64 `json:"final_price"`
	SessionID      sql.NullString  `json:"session_id"`
	ReferrerURL    sql.NullString  `json:"referrer_url"`
	UserAgent      sql.NullString  `json:"user_agent"`
	CreatedAt      time.Time       `json:"created_at"`
	Metadata       sql.NullString  `json:"metadata"`
	OldValues      sql.NullString  `json:"old_values"`
	NewValues      sql.NullString  `json:"new_values"`
	Description    sql.NullString  `json:"description"`
}

// CreateSubscriptionOffer creates a new subscription offer
func (db *DB) CreateSubscriptionOffer(offer *SubscriptionOffer) error {
	log.Printf("Database: Creating subscription offer: %s", offer.OffName)

	// First, let's check the actual table schema
	schemaQuery := "SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'subscription_offers' ORDER BY ordinal_position"
	rows, err := db.DB.Query(schemaQuery)
	if err != nil {
		log.Printf("Database: Error checking schema: %v", err)
		return err
	}
	defer rows.Close()

	log.Printf("Database: Current table schema:")
	for rows.Next() {
		var columnName, dataType string
		if err := rows.Scan(&columnName, &dataType); err != nil {
			log.Printf("Database: Error scanning schema row: %v", err)
			continue
		}
		log.Printf("  %s: %s", columnName, dataType)
	}

	query := `
		INSERT INTO subscription_offers (
			plan_id, item_id, off_discount_type, off_discount_value, 
			offer_start_date, off_end_date, is_active, off_description,
			off_name, off_code, off_max_uses, off_current_uses,
			off_terms_conditions, off_target, off_priority, off_auto_apply,
			off_created_at, off_updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW()
		) RETURNING id, off_created_at, off_updated_at
	`

	// Convert sql.NullString to interface{} for proper handling
	var itemID interface{}
	if offer.ItemID.Valid {
		itemID = offer.ItemID.String
	} else {
		itemID = nil
	}

	var offerStartDate interface{}
	if offer.OfferStartDate.Valid {
		offerStartDate = offer.OfferStartDate.Time
	} else {
		offerStartDate = nil
	}

	var offEndDate interface{}
	if offer.OffEndDate.Valid {
		offEndDate = offer.OffEndDate.Time
	} else {
		offEndDate = nil
	}

	var offDescription interface{}
	if offer.OffDescription.Valid {
		offDescription = offer.OffDescription.String
	} else {
		offDescription = nil
	}

	var offCode interface{}
	if offer.OffCode.Valid {
		offCode = offer.OffCode.String
	} else {
		offCode = nil
	}

	var offMaxUses interface{}
	if offer.OffMaxUses.Valid {
		offMaxUses = offer.OffMaxUses.Int32
	} else {
		offMaxUses = nil
	}

	var offTermsConditions interface{}
	if offer.OffTermsConditions.Valid {
		offTermsConditions = offer.OffTermsConditions.String
	} else {
		offTermsConditions = nil
	}

	var offTarget interface{}
	if offer.OffTarget.Valid {
		offTarget = offer.OffTarget.String
	} else {
		offTarget = nil
	}

	// Debug logging
	log.Printf("Database: Query parameters:")
	log.Printf("  plan_id: %v (type: %T)", offer.PlanID, offer.PlanID)
	log.Printf("  item_id: %v (type: %T)", itemID, itemID)
	log.Printf("  off_discount_type: %v (type: %T)", offer.OffDiscountType, offer.OffDiscountType)
	log.Printf("  off_discount_value: %v (type: %T)", offer.OffDiscountValue, offer.OffDiscountValue)
	log.Printf("  offer_start_date: %v (type: %T)", offerStartDate, offerStartDate)
	log.Printf("  off_end_date: %v (type: %T)", offEndDate, offEndDate)
	log.Printf("  is_active: %v (type: %T)", offer.IsActive, offer.IsActive)
	log.Printf("  off_description: %v (type: %T)", offDescription, offDescription)
	log.Printf("  off_name: %v (type: %T)", offer.OffName, offer.OffName)
	log.Printf("  off_code: %v (type: %T)", offCode, offCode)
	log.Printf("  off_max_uses: %v (type: %T)", offMaxUses, offMaxUses)
	log.Printf("  off_current_uses: %v (type: %T)", offer.OffCurrentUses, offer.OffCurrentUses)
	log.Printf("  off_terms_conditions: %v (type: %T)", offTermsConditions, offTermsConditions)
	log.Printf("  off_target: %v (type: %T)", offTarget, offTarget)
	log.Printf("  off_priority: %v (type: %T)", offer.OffPriority, offer.OffPriority)
	log.Printf("  off_auto_apply: %v (type: %T)", offer.OffAutoApply, offer.OffAutoApply)

	err = db.DB.QueryRow(
		query,
		offer.PlanID,
		itemID,
		offer.OffDiscountType,
		offer.OffDiscountValue,
		offerStartDate,
		offEndDate,
		offer.IsActive,
		offDescription,
		offer.OffName,
		offCode,
		offMaxUses,
		offer.OffCurrentUses,
		offTermsConditions,
		offTarget,
		offer.OffPriority,
		offer.OffAutoApply,
	).Scan(&offer.ID, &offer.OffCreatedAt, &offer.OffUpdatedAt)

	if err != nil {
		log.Printf("Database: Error creating subscription offer: %v", err)
		return err
	}

	log.Printf("Database: Successfully created subscription offer with ID: %d", offer.ID)
	return nil
}

// GetAllSubscriptionOffers retrieves all active subscription offers
func (db *DB) GetAllSubscriptionOffers() ([]*SubscriptionOffer, error) {
	log.Printf("Database: Getting all subscription offers")

	if db == nil {
		log.Printf("Database: Connection is nil")
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
		SELECT id, plan_id, item_id, off_discount_type, off_discount_value,
			   offer_start_date, off_end_date, is_active, off_description,
			   off_created_at, off_updated_at, off_name, off_code, off_max_uses,
			   off_current_uses, off_terms_conditions, off_target, off_priority, off_auto_apply
		FROM subscription_offers
		ORDER BY off_priority DESC, off_created_at DESC
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		log.Printf("Database: Error querying subscription offers: %v", err)
		return nil, err
	}
	defer rows.Close()

	var offers []*SubscriptionOffer
	for rows.Next() {
		offer := &SubscriptionOffer{}
		err := rows.Scan(
			&offer.ID,
			&offer.PlanID,
			&offer.ItemID,
			&offer.OffDiscountType,
			&offer.OffDiscountValue,
			&offer.OfferStartDate,
			&offer.OffEndDate,
			&offer.IsActive,
			&offer.OffDescription,
			&offer.OffCreatedAt,
			&offer.OffUpdatedAt,
			&offer.OffName,
			&offer.OffCode,
			&offer.OffMaxUses,
			&offer.OffCurrentUses,
			&offer.OffTermsConditions,
			&offer.OffTarget,
			&offer.OffPriority,
			&offer.OffAutoApply,
		)
		if err != nil {
			log.Printf("Database: Error scanning offer row: %v", err)
			return nil, err
		}
		offers = append(offers, offer)
	}

	log.Printf("Database: Retrieved %d subscription offers", len(offers))
	return offers, nil
}

// GetSubscriptionOfferByID retrieves a specific subscription offer by ID
func (db *DB) GetSubscriptionOfferByID(id int) (*SubscriptionOffer, error) {
	log.Printf("Database: Getting subscription offer by ID: %d", id)

	if db == nil {
		log.Printf("Database: Connection is nil")
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
		SELECT id, plan_id, item_id, off_discount_type, off_discount_value,
			   offer_start_date, off_end_date, is_active, off_description,
			   off_created_at, off_updated_at, off_name, off_code, off_max_uses,
			   off_current_uses, off_terms_conditions, off_target, off_priority, off_auto_apply,
			   stripe_coupon_id, stripe_promotion_code_id
		FROM subscription_offers
		WHERE id = $1
	`

	offer := &SubscriptionOffer{}
	err := db.DB.QueryRow(query, id).Scan(
		&offer.ID,
		&offer.PlanID,
		&offer.ItemID,
		&offer.OffDiscountType,
		&offer.OffDiscountValue,
		&offer.OfferStartDate,
		&offer.OffEndDate,
		&offer.IsActive,
		&offer.OffDescription,
		&offer.OffCreatedAt,
		&offer.OffUpdatedAt,
		&offer.OffName,
		&offer.OffCode,
		&offer.OffMaxUses,
		&offer.OffCurrentUses,
		&offer.OffTermsConditions,
		&offer.OffTarget,
		&offer.OffPriority,
		&offer.OffAutoApply,
		&offer.StripeCouponID,
		&offer.StripePromotionCodeID,
	)

	if err != nil {
		log.Printf("Database: Error getting subscription offer by ID: %v", err)
		return nil, err
	}

	log.Printf("Database: Successfully retrieved subscription offer: %s", offer.OffName)
	return offer, nil
}

// UpdateSubscriptionOffer updates an existing subscription offer
func (db *DB) UpdateSubscriptionOffer(id int, updates map[string]interface{}) (*SubscriptionOffer, error) {
	log.Printf("Database: Updating subscription offer ID: %d", id)

	if db == nil {
		log.Printf("Database: Connection is nil")
		return nil, fmt.Errorf("database connection is nil")
	}

	// Build dynamic query
	query := "UPDATE subscription_offers SET "
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		if argIndex > 1 {
			query += ", "
		}
		query += fmt.Sprintf("%s = $%d", key, argIndex)
		args = append(args, value)
		argIndex++
	}

	query += ", off_updated_at = NOW() WHERE id = $" + fmt.Sprintf("%d", argIndex)
	args = append(args, id)

	log.Printf("Database: Final query: %s", query)
	log.Printf("Database: Final args: %v", args)

	_, err := db.Exec(query, args...)
	if err != nil {
		log.Printf("Database: Error updating subscription offer: %v", err)
		return nil, err
	}

	// Return updated offer
	return db.GetSubscriptionOfferByID(id)
}

// DeleteSubscriptionOffer soft deletes a subscription offer
func (db *DB) DeleteSubscriptionOffer(id int) error {
	log.Printf("Database: Deleting subscription offer ID: %d", id)

	if db == nil {
		log.Printf("Database: Connection is nil")
		return fmt.Errorf("database connection is nil")
	}

	query := "UPDATE subscription_offers SET is_active = false, off_updated_at = NOW() WHERE id = $1"
	_, err := db.Exec(query, id)
	if err != nil {
		log.Printf("Database: Error deleting subscription offer: %v", err)
		return err
	}

	log.Printf("Database: Successfully deleted subscription offer ID: %d", id)
	return nil
}

// AddOfferHistoryEvent adds a new history event for an offer
func (db *DB) AddOfferHistoryEvent(event *SubscriptionOfferHistory) error {
	log.Printf("Database: Adding history event for offer %d, type: %s", event.OfferID, event.EventType)

	if db == nil {
		log.Printf("Database: Connection is nil")
		return fmt.Errorf("database connection is nil")
	}

	query := `
		INSERT INTO subscription_offers_history (
			offer_id, user_id, sub_plan_id, accepted, off_user_ip, device_info,
			event_type, discount_amount, original_price, final_price, session_id,
			referrer_url, user_agent, created_at, metadata, old_values, new_values, description
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), $14, $15, $16, $17)
	`

	_, err := db.Exec(
		query,
		event.OfferID,
		event.UserID,
		event.SubPlanID,
		event.Accepted,
		event.OffUserIP,
		event.DeviceInfo,
		event.EventType,
		event.DiscountAmount,
		event.OriginalPrice,
		event.FinalPrice,
		event.SessionID,
		event.ReferrerURL,
		event.UserAgent,
		event.Metadata,
		event.OldValues,
		event.NewValues,
		event.Description,
	)

	if err != nil {
		log.Printf("Database: Error adding offer history event: %v", err)
		return err
	}

	log.Printf("Database: Successfully added history event for offer %d", event.OfferID)
	return nil
}

// GetOfferHistory retrieves history events for a specific offer
func (db *DB) GetOfferHistory(offerID int) ([]*SubscriptionOfferHistory, error) {
	log.Printf("Database: Getting history for offer %d", offerID)

	if db == nil {
		log.Printf("Database: Connection is nil")
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
		SELECT id, offer_id, user_id, sub_plan_id, accepted, off_user_ip, device_info,
			   event_type, discount_amount, original_price, final_price, session_id,
			   referrer_url, user_agent, created_at, metadata, old_values, new_values, description
		FROM subscription_offers_history
		WHERE offer_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.DB.Query(query, offerID)
	if err != nil {
		log.Printf("Database: Error querying offer history: %v", err)
		return nil, err
	}
	defer rows.Close()

	var events []*SubscriptionOfferHistory
	for rows.Next() {
		event := &SubscriptionOfferHistory{}
		err := rows.Scan(
			&event.ID,
			&event.OfferID,
			&event.UserID,
			&event.SubPlanID,
			&event.Accepted,
			&event.OffUserIP,
			&event.DeviceInfo,
			&event.EventType,
			&event.DiscountAmount,
			&event.OriginalPrice,
			&event.FinalPrice,
			&event.SessionID,
			&event.ReferrerURL,
			&event.UserAgent,
			&event.CreatedAt,
			&event.Metadata,
			&event.OldValues,
			&event.NewValues,
			&event.Description,
		)
		if err != nil {
			log.Printf("Database: Error scanning offer history row: %v", err)
			return nil, err
		}
		events = append(events, event)
	}

	log.Printf("Database: Retrieved %d history events for offer %d", len(events), offerID)
	return events, nil
}
