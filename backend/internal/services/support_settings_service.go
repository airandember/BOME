package services

import (
	"database/sql"
	"fmt"
	"log"

	"bome-backend/internal/database"
)

// SupportSettingsService manages support contact configuration
type SupportSettingsService struct {
	db *database.DB
}

// NewSupportSettingsService creates a new SupportSettingsService
func NewSupportSettingsService(db *database.DB) *SupportSettingsService {
	return &SupportSettingsService{db: db}
}

// SupportSettings represents all support-related settings
type SupportSettings struct {
	Email   *string `json:"email"`
	Phone   *string `json:"phone"`
	URL     *string `json:"url"`
	Hours   *string `json:"hours"`
	Message *string `json:"message"`
}

// GetSupportSettings returns all support settings (public, no auth required)
func (s *SupportSettingsService) GetSupportSettings() (*SupportSettings, error) {
	query := `
		SELECT key, value 
		FROM public_settings 
		WHERE key IN ('support_email', 'support_phone', 'support_url', 'support_hours', 'support_message')
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query support settings: %w", err)
	}
	defer rows.Close()

	support := &SupportSettings{}
	for rows.Next() {
		var key string
		var value sql.NullString
		if err := rows.Scan(&key, &value); err != nil {
			log.Printf("⚠️  Failed to scan support setting: %v", err)
			continue
		}

		var valPtr *string
		if value.Valid && value.String != "" {
			valPtr = &value.String
		}

		switch key {
		case "support_email":
			support.Email = valPtr
		case "support_phone":
			support.Phone = valPtr
		case "support_url":
			support.URL = valPtr
		case "support_hours":
			support.Hours = valPtr
		case "support_message":
			support.Message = valPtr
		}
	}

	return support, nil
}

// UpdateSupportSettings updates support settings (admin only)
func (s *SupportSettingsService) UpdateSupportSettings(settings map[string]string) error {
	if len(settings) == 0 {
		return nil
	}

	// Use a transaction for atomic updates
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO public_settings (key, value, updated_at)
		VALUES ($1, NULLIF($2, ''), NOW())
		ON CONFLICT (key) 
		DO UPDATE SET value = NULLIF(EXCLUDED.value, ''), updated_at = NOW()
	`

	for key, value := range settings {
		// Only allow support-related keys
		if key != "support_email" && key != "support_phone" && key != "support_url" && key != "support_hours" && key != "support_message" {
			continue
		}

		_, err := tx.Exec(query, key, value)
		if err != nil {
			return fmt.Errorf("failed to update setting %s: %w", key, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("✅ [Support Settings] Updated %d support settings", len(settings))
	return nil
}

// HasSupportContact checks if any support contact is configured
func (s *SupportSettingsService) HasSupportContact() bool {
	support, err := s.GetSupportSettings()
	if err != nil {
		return false
	}
	return support.Email != nil || support.Phone != nil || support.URL != nil
}
