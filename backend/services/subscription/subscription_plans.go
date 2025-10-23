package subscription

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"bome-backend/infrastructure/database"
)

// SubscriptionPlanService handles business logic for subscription plans
type SubscriptionPlanService struct {
	db  *database.DB
	hub WebSocketHub
}

// SubscriptionPlan represents a subscription plan
type SubscriptionPlan struct {
	ID                 int        `json:"id"`
	Name               string     `json:"name"`
	Description        string     `json:"description"`
	ShortDesc          string     `json:"short_desc"`
	Price              float64    `json:"price"`
	Currency           string     `json:"currency"`
	Interval           string     `json:"interval"`
	IntervalCount      int        `json:"interval_count"`
	StripePriceID      *string    `json:"stripe_price_id,omitempty"`
	StripeProductID    *string    `json:"stripe_product_id,omitempty"`
	Features           []string   `json:"features"`
	IsActive           bool       `json:"is_active"`
	SubType            string     `json:"sub_type"` // stnd = standard, prmo = promotional
	PromotionStartDate *time.Time `json:"promotion_start_date,omitempty"`
	PromotionEndDate   *time.Time `json:"promotion_end_date,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
}

// CreateSubscriptionPlanRequest represents a request to create a subscription plan
type CreateSubscriptionPlanRequest struct {
	Name               string   `json:"name" binding:"required"`
	Description        string   `json:"description"`
	ShortDesc          string   `json:"short_desc"`
	Price              float64  `json:"price" binding:"required"`
	Currency           string   `json:"currency"`
	Interval           string   `json:"interval" binding:"required"`
	IntervalCount      int      `json:"interval_count"`
	StripePriceID      string   `json:"stripe_price_id"`
	StripeProductID    string   `json:"stripe_product_id"`
	Features           []string `json:"features"`
	IsActive           bool     `json:"is_active"`
	SubType            string   `json:"sub_type"`
	PromotionStartDate string   `json:"promotion_start_date"`
	PromotionEndDate   string   `json:"promotion_end_date"`
}

// UpdateSubscriptionPlanRequest represents a request to update a subscription plan
type UpdateSubscriptionPlanRequest struct {
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	ShortDesc          string   `json:"short_desc"`
	Price              float64  `json:"price"`
	Currency           string   `json:"currency"`
	Interval           string   `json:"interval"`
	IntervalCount      int      `json:"interval_count"`
	StripePriceID      string   `json:"stripe_price_id"`
	StripeProductID    string   `json:"stripe_product_id"`
	Features           []string `json:"features"`
	IsActive           bool     `json:"is_active"`
	SubType            string   `json:"sub_type"`
	PromotionStartDate string   `json:"promotion_start_date"`
	PromotionEndDate   string   `json:"promotion_end_date"`
}

// NewSubscriptionPlanService creates a new subscription plan service
func NewSubscriptionPlanService(db *database.DB, hub WebSocketHub) *SubscriptionPlanService {
	return &SubscriptionPlanService{
		db:  db,
		hub: hub,
	}
}

// GetAllPlans retrieves all subscription plans
func (s *SubscriptionPlanService) GetAllPlans() ([]*SubscriptionPlan, error) {
	query := `
		SELECT 
			id, name, description, short_desc, price, currency, 
			interval, interval_count, stripe_price_id, stripe_product_id,
			features, is_active, sub_type,
			promotion_start_date, promotion_end_date,
			created_at, updated_at, deleted_at
		FROM subscription_plans
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := s.db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*SubscriptionPlan
	for rows.Next() {
		plan := &SubscriptionPlan{}
		var featuresJSON sql.NullString
		var shortDesc sql.NullString
		var stripePriceID sql.NullString
		var stripeProductID sql.NullString

		err := rows.Scan(
			&plan.ID,
			&plan.Name,
			&plan.Description,
			&shortDesc,
			&plan.Price,
			&plan.Currency,
			&plan.Interval,
			&plan.IntervalCount,
			&stripePriceID,
			&stripeProductID,
			&featuresJSON,
			&plan.IsActive,
			&plan.SubType,
			&plan.PromotionStartDate,
			&plan.PromotionEndDate,
			&plan.CreatedAt,
			&plan.UpdatedAt,
			&plan.DeletedAt,
		)
		if err != nil {
			log.Printf("Error scanning plan: %v", err)
			continue
		}

		// Parse features JSON
		if featuresJSON.Valid && featuresJSON.String != "" {
			if err := json.Unmarshal([]byte(featuresJSON.String), &plan.Features); err != nil {
				log.Printf("Error parsing features: %v", err)
				plan.Features = []string{}
			}
		} else {
			plan.Features = []string{}
		}

		if shortDesc.Valid {
			plan.ShortDesc = shortDesc.String
		}
		if stripePriceID.Valid {
			plan.StripePriceID = &stripePriceID.String
		}
		if stripeProductID.Valid {
			plan.StripeProductID = &stripeProductID.String
		}

		plans = append(plans, plan)
	}

	return plans, nil
}

// GetPlanByID retrieves a subscription plan by ID
func (s *SubscriptionPlanService) GetPlanByID(id int) (*SubscriptionPlan, error) {
	query := `
		SELECT 
			id, name, description, short_desc, price, currency, 
			interval, interval_count, stripe_price_id, stripe_product_id,
			features, is_active, sub_type,
			promotion_start_date, promotion_end_date,
			created_at, updated_at, deleted_at
		FROM subscription_plans
		WHERE id = $1 AND deleted_at IS NULL
	`

	plan := &SubscriptionPlan{}
	var featuresJSON sql.NullString
	var shortDesc sql.NullString
	var stripePriceID sql.NullString
	var stripeProductID sql.NullString

	err := s.db.DB.QueryRow(query, id).Scan(
		&plan.ID,
		&plan.Name,
		&plan.Description,
		&shortDesc,
		&plan.Price,
		&plan.Currency,
		&plan.Interval,
		&plan.IntervalCount,
		&stripePriceID,
		&stripeProductID,
		&featuresJSON,
		&plan.IsActive,
		&plan.SubType,
		&plan.PromotionStartDate,
		&plan.PromotionEndDate,
		&plan.CreatedAt,
		&plan.UpdatedAt,
		&plan.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("plan not found")
		}
		return nil, err
	}

	// Parse features JSON
	if featuresJSON.Valid && featuresJSON.String != "" {
		if err := json.Unmarshal([]byte(featuresJSON.String), &plan.Features); err != nil {
			log.Printf("Error parsing features: %v", err)
			plan.Features = []string{}
		}
	} else {
		plan.Features = []string{}
	}

	if shortDesc.Valid {
		plan.ShortDesc = shortDesc.String
	}
	if stripePriceID.Valid {
		plan.StripePriceID = &stripePriceID.String
	}
	if stripeProductID.Valid {
		plan.StripeProductID = &stripeProductID.String
	}

	return plan, nil
}

// CreatePlan creates a new subscription plan
func (s *SubscriptionPlanService) CreatePlan(ctx context.Context, req *CreateSubscriptionPlanRequest) (*SubscriptionPlan, error) {
	// Marshal features to JSON
	featuresJSON, err := json.Marshal(req.Features)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal features: %w", err)
	}

	// Parse dates if provided
	var promotionStartDate, promotionEndDate *time.Time
	if req.PromotionStartDate != "" {
		if date, err := time.Parse("2006-01-02", req.PromotionStartDate); err == nil {
			promotionStartDate = &date
		}
	}
	if req.PromotionEndDate != "" {
		if date, err := time.Parse("2006-01-02", req.PromotionEndDate); err == nil {
			promotionEndDate = &date
		}
	}

	// Set defaults
	if req.Currency == "" {
		req.Currency = "USD"
	}
	if req.IntervalCount == 0 {
		req.IntervalCount = 1
	}
	if req.SubType == "" {
		req.SubType = "stnd"
	}

	query := `
		INSERT INTO subscription_plans 
		(name, description, short_desc, price, currency, interval, interval_count,
		 stripe_price_id, stripe_product_id, features, is_active, sub_type,
		 promotion_start_date, promotion_end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	plan := &SubscriptionPlan{
		Name:               req.Name,
		Description:        req.Description,
		ShortDesc:          req.ShortDesc,
		Price:              req.Price,
		Currency:           req.Currency,
		Interval:           req.Interval,
		IntervalCount:      req.IntervalCount,
		Features:           req.Features,
		IsActive:           req.IsActive,
		SubType:            req.SubType,
		PromotionStartDate: promotionStartDate,
		PromotionEndDate:   promotionEndDate,
	}

	if req.StripePriceID != "" {
		plan.StripePriceID = &req.StripePriceID
	}
	if req.StripeProductID != "" {
		plan.StripeProductID = &req.StripeProductID
	}

	err = s.db.DB.QueryRowContext(ctx, query,
		plan.Name, plan.Description, plan.ShortDesc, plan.Price, plan.Currency,
		plan.Interval, plan.IntervalCount, plan.StripePriceID, plan.StripeProductID,
		string(featuresJSON), plan.IsActive, plan.SubType,
		plan.PromotionStartDate, plan.PromotionEndDate,
	).Scan(&plan.ID, &plan.CreatedAt, &plan.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create plan: %w", err)
	}

	// Broadcast event via WebSocket
	if s.hub != nil {
		s.hub.BroadcastEvent("plan.created", map[string]interface{}{
			"plan": plan,
		}, fmt.Sprintf("New plan created: %s", plan.Name))
		log.Printf("📡 Broadcasted plan creation: %s", plan.Name)
	}

	return plan, nil
}

// UpdatePlan updates a subscription plan
func (s *SubscriptionPlanService) UpdatePlan(ctx context.Context, id int, req *UpdateSubscriptionPlanRequest) (*SubscriptionPlan, error) {
	// Marshal features to JSON
	featuresJSON, err := json.Marshal(req.Features)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal features: %w", err)
	}

	// Parse dates if provided
	var promotionStartDate, promotionEndDate *time.Time
	if req.PromotionStartDate != "" {
		if date, err := time.Parse("2006-01-02", req.PromotionStartDate); err == nil {
			promotionStartDate = &date
		}
	}
	if req.PromotionEndDate != "" {
		if date, err := time.Parse("2006-01-02", req.PromotionEndDate); err == nil {
			promotionEndDate = &date
		}
	}

	query := `
		UPDATE subscription_plans
		SET name = $1, description = $2, short_desc = $3, price = $4, currency = $5,
		    interval = $6, interval_count = $7, stripe_price_id = $8, stripe_product_id = $9,
		    features = $10, is_active = $11, sub_type = $12,
		    promotion_start_date = $13, promotion_end_date = $14, updated_at = NOW()
		WHERE id = $15 AND deleted_at IS NULL
		RETURNING updated_at
	`

	var updatedAt time.Time
	err = s.db.DB.QueryRowContext(ctx, query,
		req.Name, req.Description, req.ShortDesc, req.Price, req.Currency,
		req.Interval, req.IntervalCount, req.StripePriceID, req.StripeProductID,
		string(featuresJSON), req.IsActive, req.SubType,
		promotionStartDate, promotionEndDate, id,
	).Scan(&updatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("plan not found")
		}
		return nil, fmt.Errorf("failed to update plan: %w", err)
	}

	// Get updated plan
	plan, err := s.GetPlanByID(id)
	if err != nil {
		return nil, err
	}

	// Broadcast event via WebSocket
	if s.hub != nil {
		s.hub.BroadcastEvent("plan.updated", map[string]interface{}{
			"plan": plan,
		}, fmt.Sprintf("Plan updated: %s", plan.Name))
		log.Printf("📡 Broadcasted plan update: %s", plan.Name)
	}

	return plan, nil
}

// DeletePlan soft deletes a subscription plan
func (s *SubscriptionPlanService) DeletePlan(ctx context.Context, id int) error {
	query := `
		UPDATE subscription_plans
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := s.db.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete plan: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("plan not found")
	}

	// Broadcast event via WebSocket
	if s.hub != nil {
		s.hub.BroadcastEvent("plan.deleted", map[string]interface{}{
			"plan_id": id,
		}, "Plan deleted")
		log.Printf("📡 Broadcasted plan deletion: ID %d", id)
	}

	return nil
}

// TogglePlanStatus toggles the is_active status of a plan
func (s *SubscriptionPlanService) TogglePlanStatus(ctx context.Context, id int) (*SubscriptionPlan, error) {
	query := `
		UPDATE subscription_plans
		SET is_active = NOT is_active, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING is_active
	`

	var isActive bool
	err := s.db.DB.QueryRowContext(ctx, query, id).Scan(&isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("plan not found")
		}
		return nil, fmt.Errorf("failed to toggle plan status: %w", err)
	}

	// Get updated plan
	plan, err := s.GetPlanByID(id)
	if err != nil {
		return nil, err
	}

	// Broadcast event via WebSocket
	if s.hub != nil {
		status := "deactivated"
		if isActive {
			status = "activated"
		}
		s.hub.BroadcastEvent("plan.updated", map[string]interface{}{
			"plan": plan,
		}, fmt.Sprintf("Plan %s: %s", status, plan.Name))
		log.Printf("📡 Broadcasted plan status toggle: %s (%s)", plan.Name, status)
	}

	return plan, nil
}

// UpdatePromotionStatus updates the promotion dates for a plan
func (s *SubscriptionPlanService) UpdatePromotionStatus(ctx context.Context, id int, startDate, endDate *time.Time) (*SubscriptionPlan, error) {
	query := `
		UPDATE subscription_plans
		SET promotion_start_date = $1, promotion_end_date = $2, updated_at = NOW()
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := s.db.DB.ExecContext(ctx, query, startDate, endDate, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update promotion status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rows == 0 {
		return nil, fmt.Errorf("plan not found")
	}

	// Get updated plan
	plan, err := s.GetPlanByID(id)
	if err != nil {
		return nil, err
	}

	// Broadcast event via WebSocket
	if s.hub != nil {
		s.hub.BroadcastEvent("plan.updated", map[string]interface{}{
			"plan": plan,
		}, fmt.Sprintf("Promotion updated: %s", plan.Name))
	}

	return plan, nil
}

// Helper function to parse flexible date formats
func parseFlexibleDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}
