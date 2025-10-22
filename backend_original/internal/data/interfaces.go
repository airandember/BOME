package data

import (
	"fmt"
	"strings"
	"time"

	"bome-backend/internal/database"
)

// DataSource defines the interface for data access
type DataSource interface {
	GetYouTubeVideos(limit int) ([]database.YouTubeVideo, error)
	GetYouTubeVideoByID(id string) (*database.YouTubeVideo, error)
	GetYouTubeChannelInfo() (*ChannelInfo, error)
	SearchYouTubeVideos(query string, limit int) ([]database.YouTubeVideo, error)
	GetYouTubeVideosByCategory(category string, limit int) ([]database.YouTubeVideo, error)
}

// ChannelInfo represents YouTube channel information
type ChannelInfo struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	SubscriberCount int       `json:"subscriber_count"`
	VideoCount      int       `json:"video_count"`
	ViewCount       int       `json:"view_count"`
	CreatedAt       time.Time `json:"created_at"`
	Country         string    `json:"country"`
	CustomURL       string    `json:"custom_url"`
}

// MockDataSource implements DataSource using JSON files
type MockDataSource struct {
	loader *MockDataLoader
}

// RealDataSource implements DataSource using APIs/Database
type RealDataSource struct {
	db     *database.DB
	apiKey string
}

// NewMockDataSource creates a new mock data source
func NewMockDataSource(dataDir string) DataSource {
	return &MockDataSource{
		loader: NewMockDataLoader(dataDir),
	}
}

// NewRealDataSource creates a new real data source
func NewRealDataSource(db *database.DB, apiKey string) DataSource {
	return &RealDataSource{
		db:     db,
		apiKey: apiKey,
	}
}

// MockDataSource implementation
func (m *MockDataSource) GetYouTubeVideos(limit int) ([]database.YouTubeVideo, error) {
	videos, err := m.loader.LoadYouTubeVideos()
	if err != nil {
		return nil, err
	}

	if limit > 0 && limit < len(videos) {
		return videos[:limit], nil
	}

	return videos, nil
}

func (m *MockDataSource) GetYouTubeVideoByID(id string) (*database.YouTubeVideo, error) {
	videos, err := m.loader.LoadYouTubeVideos()
	if err != nil {
		return nil, err
	}

	for _, video := range videos {
		if video.ID == id {
			return &video, nil
		}
	}

	return nil, fmt.Errorf("video not found: %s", id)
}

func (m *MockDataSource) GetYouTubeChannelInfo() (*ChannelInfo, error) {
	return m.loader.GetYouTubeChannelInfo()
}

func (m *MockDataSource) SearchYouTubeVideos(query string, limit int) ([]database.YouTubeVideo, error) {
	videos, err := m.loader.LoadYouTubeVideos()
	if err != nil {
		return nil, err
	}

	// Simple search implementation for mock data
	var results []database.YouTubeVideo
	query = strings.ToLower(query)

	for _, video := range videos {
		if strings.Contains(strings.ToLower(video.Title), query) ||
			strings.Contains(strings.ToLower(video.Description), query) {
			results = append(results, video)
			if limit > 0 && len(results) >= limit {
				break
			}
		}
	}

	return results, nil
}

func (m *MockDataSource) GetYouTubeVideosByCategory(category string, limit int) ([]database.YouTubeVideo, error) {
	videos, err := m.loader.LoadYouTubeVideos()
	if err != nil {
		return nil, err
	}

	// Filter by category (this would need to be added to the database model)
	// For now, return all videos as they're all educational content
	if limit > 0 && limit < len(videos) {
		return videos[:limit], nil
	}

	return videos, nil
}

// RealDataSource implementation (for production)
func (r *RealDataSource) GetYouTubeVideos(limit int) ([]database.YouTubeVideo, error) {
	// Implementation for real database calls
	return r.db.GetYouTubeVideos(limit)
}

func (r *RealDataSource) GetYouTubeVideoByID(id string) (*database.YouTubeVideo, error) {
	// Implementation for real database/API calls
	return r.db.GetYouTubeVideoByID(id)
}

func (r *RealDataSource) GetYouTubeChannelInfo() (*ChannelInfo, error) {
	// Implementation for real YouTube API calls
	// This would call YouTube Data API v3
	return nil, fmt.Errorf("real YouTube API integration not implemented yet")
}

func (r *RealDataSource) SearchYouTubeVideos(query string, limit int) ([]database.YouTubeVideo, error) {
	// Implementation for real database search
	return nil, fmt.Errorf("real search not implemented yet")
}

func (r *RealDataSource) GetYouTubeVideosByCategory(category string, limit int) ([]database.YouTubeVideo, error) {
	// Implementation for real database category filtering
	return nil, fmt.Errorf("real category filtering not implemented yet")
}
