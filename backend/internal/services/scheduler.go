package services

import (
	"log"
	"time"

	"bome-backend/internal/database"
)

// SchedulerService handles automated scheduling tasks
type SchedulerService struct {
	db     *database.DB
	ticker *time.Ticker
	done   chan bool
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService(db *database.DB) *SchedulerService {
	return &SchedulerService{
		db:   db,
		done: make(chan bool),
	}
}

// Start begins the scheduler with the specified interval
func (s *SchedulerService) Start(interval time.Duration) {
	s.ticker = time.NewTicker(interval)
	go s.run()
	log.Printf("Scheduler service started with %v interval", interval)
}

// Stop stops the scheduler
func (s *SchedulerService) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.done <- true
	log.Println("Scheduler service stopped")
}

// run is the main scheduler loop
func (s *SchedulerService) run() {
	for {
		select {
		case <-s.ticker.C:
			s.processScheduledVideos()
		case <-s.done:
			return
		}
	}
}

// processScheduledVideos checks for and publishes videos that are scheduled to be published
func (s *SchedulerService) processScheduledVideos() {
	now := time.Now()

	// Get videos scheduled to be published before now
	videos, err := s.db.GetScheduledVideos(now)
	if err != nil {
		log.Printf("Error getting scheduled videos: %v", err)
		return
	}

	if len(videos) == 0 {
		return
	}

	log.Printf("Processing %d scheduled videos", len(videos))

	for _, video := range videos {
		// Update video status to published
		if err := s.db.UpdateVideoStatus(video.ID, "published"); err != nil {
			log.Printf("Error publishing scheduled video %d: %v", video.ID, err)
			continue
		}

		// Clear the scheduled publish date
		if err := s.db.UnscheduleVideo(video.ID); err != nil {
			log.Printf("Error clearing schedule for video %d: %v", video.ID, err)
		}

		log.Printf("Successfully published scheduled video: %s (ID: %d)", video.Title, video.ID)
	}
}

// PublishScheduledVideo manually publishes a specific scheduled video
func (s *SchedulerService) PublishScheduledVideo(videoID int) error {
	// Update video status to published
	if err := s.db.UpdateVideoStatus(videoID, "published"); err != nil {
		return err
	}

	// Clear the scheduled publish date
	return s.db.UnscheduleVideo(videoID)
}
