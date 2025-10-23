package services

import (
	"fmt"
	"log"

	"bome-backend/creator/models"
	"bome-backend/infrastructure/database"
)

// PresenterService handles business logic for presenter management
type PresenterService struct {
	db *database.DB
}

// NewPresenterService creates a new PresenterService
func NewPresenterService(db *database.DB) *PresenterService {
	return &PresenterService{
		db: db,
	}
}

// CreatePresenter creates a new presenter
func (s *PresenterService) CreatePresenter(input *models.CreatePresenterInput) (*models.Presenter, error) {
	log.Printf("[PRESENTER-SERVICE] Creating presenter: %s", input.Name)
	
	// Validate required fields
	if input.Name == "" {
		return nil, fmt.Errorf("presenter name is required")
	}
	
	// Create presenter
	presenter, err := models.CreatePresenter(s.db, input)
	if err != nil {
		log.Printf("[PRESENTER-SERVICE] Error creating presenter: %v", err)
		return nil, fmt.Errorf("failed to create presenter: %w", err)
	}
	
	log.Printf("[PRESENTER-SERVICE] Presenter created successfully: ID=%d, Name=%s", presenter.ID, presenter.Name)
	return presenter, nil
}

// GetPresenterByID retrieves a presenter by ID
func (s *PresenterService) GetPresenterByID(id int) (*models.Presenter, error) {
	log.Printf("[PRESENTER-SERVICE] Getting presenter by ID: %d", id)
	
	presenter, err := models.GetPresenterByID(s.db, id)
	if err != nil {
		log.Printf("[PRESENTER-SERVICE] Error getting presenter: %v", err)
		return nil, fmt.Errorf("presenter not found: %w", err)
	}
	
	return presenter, nil
}

// GetPresenters retrieves all presenters with optional filtering
func (s *PresenterService) GetPresenters(activeOnly bool, verifiedOnly bool) ([]*models.Presenter, error) {
	log.Printf("[PRESENTER-SERVICE] Getting presenters (active=%v, verified=%v)", activeOnly, verifiedOnly)
	
	presenters, err := models.GetPresenters(s.db, activeOnly, verifiedOnly)
	if err != nil {
		log.Printf("[PRESENTER-SERVICE] Error getting presenters: %v", err)
		return nil, fmt.Errorf("failed to get presenters: %w", err)
	}
	
	log.Printf("[PRESENTER-SERVICE] Retrieved %d presenters", len(presenters))
	return presenters, nil
}

// UpdatePresenter updates a presenter's information
func (s *PresenterService) UpdatePresenter(id int, input *models.UpdatePresenterInput) (*models.Presenter, error) {
	log.Printf("[PRESENTER-SERVICE] Updating presenter: %d", id)
	
	// Check if presenter exists
	_, err := models.GetPresenterByID(s.db, id)
	if err != nil {
		return nil, fmt.Errorf("presenter not found: %w", err)
	}
	
	// Update presenter
	presenter, err := models.UpdatePresenter(s.db, id, input)
	if err != nil {
		log.Printf("[PRESENTER-SERVICE] Error updating presenter: %v", err)
		return nil, fmt.Errorf("failed to update presenter: %w", err)
	}
	
	log.Printf("[PRESENTER-SERVICE] Presenter updated successfully: %d", id)
	return presenter, nil
}

// DeletePresenter soft-deletes a presenter
func (s *PresenterService) DeletePresenter(id int) error {
	log.Printf("[PRESENTER-SERVICE] Deleting presenter: %d", id)
	
	// Check if presenter exists
	_, err := models.GetPresenterByID(s.db, id)
	if err != nil {
		return fmt.Errorf("presenter not found: %w", err)
	}
	
	// Soft delete
	err = models.DeletePresenter(s.db, id)
	if err != nil {
		log.Printf("[PRESENTER-SERVICE] Error deleting presenter: %v", err)
		return fmt.Errorf("failed to delete presenter: %w", err)
	}
	
	log.Printf("[PRESENTER-SERVICE] Presenter deleted successfully: %d", id)
	return nil
}

// VerifyPresenter marks a presenter as verified
func (s *PresenterService) VerifyPresenter(presenterID int, verifiedBy int) error {
	log.Printf("[PRESENTER-SERVICE] Verifying presenter: %d by user %d", presenterID, verifiedBy)
	
	// Check if presenter exists
	_, err := models.GetPresenterByID(s.db, presenterID)
	if err != nil {
		return fmt.Errorf("presenter not found: %w", err)
	}
	
	err = models.VerifyPresenter(s.db, presenterID, verifiedBy)
	if err != nil {
		log.Printf("[PRESENTER-SERVICE] Error verifying presenter: %v", err)
		return fmt.Errorf("failed to verify presenter: %w", err)
	}
	
	log.Printf("[PRESENTER-SERVICE] Presenter verified successfully: %d", presenterID)
	return nil
}

// GetPresenterStats retrieves aggregated presenter statistics
func (s *PresenterService) GetPresenterStats() (*models.PresenterStats, error) {
	log.Printf("[PRESENTER-SERVICE] Getting presenter stats")
	
	stats, err := models.GetPresenterStats(s.db)
	if err != nil {
		log.Printf("[PRESENTER-SERVICE] Error getting presenter stats: %v", err)
		return nil, fmt.Errorf("failed to get presenter stats: %w", err)
	}
	
	return stats, nil
}

// LinkPresenterToVideo links a presenter to a video
func (s *PresenterService) LinkPresenterToVideo(input *models.CreateVideoPresenterInput) (*models.VideoPresenter, error) {
	log.Printf("[PRESENTER-SERVICE] Linking presenter %d to video %d", input.PresenterID, input.VideoID)
	
	// Validate presenter exists
	_, err := models.GetPresenterByID(s.db, input.PresenterID)
	if err != nil {
		return nil, fmt.Errorf("presenter not found: %w", err)
	}
	
	// Create link
	videoPresenter, err := models.CreateVideoPresenter(s.db, input)
	if err != nil {
		log.Printf("[PRESENTER-SERVICE] Error linking presenter to video: %v", err)
		return nil, fmt.Errorf("failed to link presenter to video: %w", err)
	}
	
	// Update presenter statistics
	err = s.UpdatePresenterStatistics(input.PresenterID)
	if err != nil {
		log.Printf("[PRESENTER-SERVICE] Warning: failed to update presenter stats: %v", err)
	}
	
	log.Printf("[PRESENTER-SERVICE] Presenter linked to video successfully")
	return videoPresenter, nil
}

// GetPresenterVideos retrieves all videos for a presenter
func (s *PresenterService) GetPresenterVideos(presenterID int) ([]*models.VideoPresenterWithDetails, error) {
	log.Printf("[PRESENTER-SERVICE] Getting videos for presenter: %d", presenterID)
	
	videos, err := models.GetVideoPresentersByPresenterID(s.db, presenterID)
	if err != nil {
		log.Printf("[PRESENTER-SERVICE] Error getting presenter videos: %v", err)
		return nil, fmt.Errorf("failed to get presenter videos: %w", err)
	}
	
	log.Printf("[PRESENTER-SERVICE] Retrieved %d videos for presenter %d", len(videos), presenterID)
	return videos, nil
}

// GetVideoPresenters retrieves all presenters for a video
func (s *PresenterService) GetVideoPresenters(videoID int) ([]*models.VideoPresenterWithDetails, error) {
	log.Printf("[PRESENTER-SERVICE] Getting presenters for video: %d", videoID)
	
	presenters, err := models.GetVideoPresentersByVideoID(s.db, videoID)
	if err != nil {
		log.Printf("[PRESENTER-SERVICE] Error getting video presenters: %v", err)
		return nil, fmt.Errorf("failed to get video presenters: %w", err)
	}
	
	log.Printf("[PRESENTER-SERVICE] Retrieved %d presenters for video %d", len(presenters), videoID)
	return presenters, nil
}

// UpdateVideoPresenter updates a video-presenter link
func (s *PresenterService) UpdateVideoPresenter(id int, input *models.UpdateVideoPresenterInput) (*models.VideoPresenter, error) {
	log.Printf("[PRESENTER-SERVICE] Updating video-presenter link: %d", id)
	
	videoPresenter, err := models.UpdateVideoPresenter(s.db, id, input)
	if err != nil {
		log.Printf("[PRESENTER-SERVICE] Error updating video-presenter link: %v", err)
		return nil, fmt.Errorf("failed to update video-presenter link: %w", err)
	}
	
	log.Printf("[PRESENTER-SERVICE] Video-presenter link updated successfully")
	return videoPresenter, nil
}

// UnlinkPresenterFromVideo removes a presenter from a video
func (s *PresenterService) UnlinkPresenterFromVideo(id int) error {
	log.Printf("[PRESENTER-SERVICE] Unlinking video-presenter: %d", id)
	
	// Get the link first to get presenter ID for stats update
	videoPresenter, err := models.GetVideoPresenterByID(s.db, id)
	if err != nil {
		return fmt.Errorf("video-presenter link not found: %w", err)
	}
	
	err = models.DeleteVideoPresenter(s.db, id)
	if err != nil {
		log.Printf("[PRESENTER-SERVICE] Error unlinking presenter from video: %v", err)
		return fmt.Errorf("failed to unlink presenter from video: %w", err)
	}
	
	// Update presenter statistics
	err = s.UpdatePresenterStatistics(videoPresenter.PresenterID)
	if err != nil {
		log.Printf("[PRESENTER-SERVICE] Warning: failed to update presenter stats: %v", err)
	}
	
	log.Printf("[PRESENTER-SERVICE] Presenter unlinked from video successfully")
	return nil
}

// UpdatePresenterStatistics calls the database function to update cached stats
func (s *PresenterService) UpdatePresenterStatistics(presenterID int) error {
	log.Printf("[PRESENTER-SERVICE] Updating statistics for presenter: %d", presenterID)
	
	query := `SELECT update_presenter_statistics($1)`
	_, err := s.db.Exec(query, presenterID)
	if err != nil {
		log.Printf("[PRESENTER-SERVICE] Error updating presenter statistics: %v", err)
		return fmt.Errorf("failed to update presenter statistics: %w", err)
	}
	
	log.Printf("[PRESENTER-SERVICE] Presenter statistics updated successfully: %d", presenterID)
	return nil
}

// UpdateAllPresenterStatistics updates statistics for all presenters
func (s *PresenterService) UpdateAllPresenterStatistics() error {
	log.Printf("[PRESENTER-SERVICE] Updating statistics for all presenters")
	
	// Get all presenters
	presenters, err := models.GetPresenters(s.db, false, false)
	if err != nil {
		return fmt.Errorf("failed to get presenters: %w", err)
	}
	
	// Update each presenter's stats
	successCount := 0
	errorCount := 0
	
	for _, presenter := range presenters {
		err := s.UpdatePresenterStatistics(presenter.ID)
		if err != nil {
			log.Printf("[PRESENTER-SERVICE] Error updating stats for presenter %d: %v", presenter.ID, err)
			errorCount++
		} else {
			successCount++
		}
	}
	
	log.Printf("[PRESENTER-SERVICE] Updated statistics: %d successful, %d errors", successCount, errorCount)
	
	if errorCount > 0 && successCount == 0 {
		return fmt.Errorf("failed to update any presenter statistics")
	}
	
	return nil
}

