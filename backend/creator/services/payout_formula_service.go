package services

import (
	"fmt"
	"log"

	"bome-backend/creator/models"
	"bome-backend/infrastructure/database"
)

// PayoutFormulaService handles business logic for payout formula management
type PayoutFormulaService struct {
	db *database.DB
}

// NewPayoutFormulaService creates a new PayoutFormulaService
func NewPayoutFormulaService(db *database.DB) *PayoutFormulaService {
	return &PayoutFormulaService{
		db: db,
	}
}

// CreateFormula creates a new payout formula
func (s *PayoutFormulaService) CreateFormula(input *models.CreatePayoutFormulaInput) (*models.PayoutFormula, error) {
	log.Printf("[FORMULA-SERVICE] Creating payout formula: %s", input.Name)
	
	// Validate required fields
	if input.Name == "" {
		return nil, fmt.Errorf("formula name is required")
	}
	
	if input.FormulaType == "" {
		return nil, fmt.Errorf("formula type is required")
	}
	
	// Validate formula type
	validTypes := map[string]bool{
		"per_view":         true,
		"per_watch_minute": true,
		"flat_rate":        true,
		"tier_based":       true,
		"hybrid":           true,
	}
	
	if !validTypes[input.FormulaType] {
		return nil, fmt.Errorf("invalid formula type: %s", input.FormulaType)
	}
	
	// Create formula
	formula, err := models.CreatePayoutFormula(s.db, input)
	if err != nil {
		log.Printf("[FORMULA-SERVICE] Error creating formula: %v", err)
		return nil, fmt.Errorf("failed to create formula: %w", err)
	}
	
	log.Printf("[FORMULA-SERVICE] Formula created successfully: ID=%d, Name=%s", formula.ID, formula.Name)
	return formula, nil
}

// GetFormulaByID retrieves a payout formula by ID
func (s *PayoutFormulaService) GetFormulaByID(id int) (*models.PayoutFormula, error) {
	log.Printf("[FORMULA-SERVICE] Getting formula by ID: %d", id)
	
	formula, err := models.GetPayoutFormulaByID(s.db, id)
	if err != nil {
		log.Printf("[FORMULA-SERVICE] Error getting formula: %v", err)
		return nil, fmt.Errorf("formula not found: %w", err)
	}
	
	return formula, nil
}

// GetFormulas retrieves all payout formulas with optional filtering
func (s *PayoutFormulaService) GetFormulas(activeOnly bool) ([]*models.PayoutFormula, error) {
	log.Printf("[FORMULA-SERVICE] Getting formulas (active=%v)", activeOnly)
	
	formulas, err := models.GetPayoutFormulas(s.db, activeOnly)
	if err != nil {
		log.Printf("[FORMULA-SERVICE] Error getting formulas: %v", err)
		return nil, fmt.Errorf("failed to get formulas: %w", err)
	}
	
	log.Printf("[FORMULA-SERVICE] Retrieved %d formulas", len(formulas))
	return formulas, nil
}

// GetDefaultFormula retrieves the default payout formula
func (s *PayoutFormulaService) GetDefaultFormula() (*models.PayoutFormula, error) {
	log.Printf("[FORMULA-SERVICE] Getting default formula")
	
	formula, err := models.GetDefaultPayoutFormula(s.db)
	if err != nil {
		log.Printf("[FORMULA-SERVICE] Error getting default formula: %v", err)
		return nil, fmt.Errorf("no default formula configured: %w", err)
	}
	
	return formula, nil
}

// UpdateFormula updates a payout formula
func (s *PayoutFormulaService) UpdateFormula(id int, input *models.UpdatePayoutFormulaInput) (*models.PayoutFormula, error) {
	log.Printf("[FORMULA-SERVICE] Updating formula: %d", id)
	
	// Check if formula exists
	_, err := models.GetPayoutFormulaByID(s.db, id)
	if err != nil {
		return nil, fmt.Errorf("formula not found: %w", err)
	}
	
	// Update formula
	formula, err := models.UpdatePayoutFormula(s.db, id, input)
	if err != nil {
		log.Printf("[FORMULA-SERVICE] Error updating formula: %v", err)
		return nil, fmt.Errorf("failed to update formula: %w", err)
	}
	
	log.Printf("[FORMULA-SERVICE] Formula updated successfully: %d", id)
	return formula, nil
}

// DeleteFormula soft-deletes a payout formula
func (s *PayoutFormulaService) DeleteFormula(id int) error {
	log.Printf("[FORMULA-SERVICE] Deleting formula: %d", id)
	
	// Check if formula exists
	formula, err := models.GetPayoutFormulaByID(s.db, id)
	if err != nil {
		return fmt.Errorf("formula not found: %w", err)
	}
	
	// Don't allow deleting the default formula
	if formula.IsDefault {
		return fmt.Errorf("cannot delete the default formula")
	}
	
	// Soft delete
	err = models.DeletePayoutFormula(s.db, id)
	if err != nil {
		log.Printf("[FORMULA-SERVICE] Error deleting formula: %v", err)
		return fmt.Errorf("failed to delete formula: %w", err)
	}
	
	log.Printf("[FORMULA-SERVICE] Formula deleted successfully: %d", id)
	return nil
}

// SetDefaultFormula sets a formula as the default
func (s *PayoutFormulaService) SetDefaultFormula(id int) error {
	log.Printf("[FORMULA-SERVICE] Setting default formula: %d", id)
	
	// Check if formula exists and is active
	formula, err := models.GetPayoutFormulaByID(s.db, id)
	if err != nil {
		return fmt.Errorf("formula not found: %w", err)
	}
	
	if !formula.IsActive {
		return fmt.Errorf("cannot set inactive formula as default")
	}
	
	err = models.SetDefaultFormula(s.db, id)
	if err != nil {
		log.Printf("[FORMULA-SERVICE] Error setting default formula: %v", err)
		return fmt.Errorf("failed to set default formula: %w", err)
	}
	
	log.Printf("[FORMULA-SERVICE] Default formula set successfully: %d", id)
	return nil
}

