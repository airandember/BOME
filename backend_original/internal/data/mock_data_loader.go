package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"bome-backend/internal/database"
)

// MockDataLoader handles loading mock data from JSON files
type MockDataLoader struct {
	dataDir string
}

// YouTubeData represents the structure of the YouTube JSON file
type YouTubeData struct {
	Videos   []YouTubeVideoJSON `json:"videos"`
	Metadata struct {
		TotalCount  int       `json:"total_count"`
		LastUpdated time.Time `json:"last_updated"`
		ChannelInfo struct {
			ID              string    `json:"id"`
			Name            string    `json:"name"`
			Description     string    `json:"description"`
			SubscriberCount int       `json:"subscriber_count"`
			VideoCount      int       `json:"video_count"`
			ViewCount       int       `json:"view_count"`
			CreatedAt       time.Time `json:"created_at"`
			Country         string    `json:"country"`
			CustomURL       string    `json:"custom_url"`
		} `json:"channel_info"`
	} `json:"metadata"`
}

// YouTubeVideoJSON represents the JSON structure of a YouTube video
type YouTubeVideoJSON struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	PublishedAt  time.Time `json:"published_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	CreatedAt    time.Time `json:"created_at"`
	ThumbnailURL string    `json:"thumbnail_url"`
	VideoURL     string    `json:"video_url"`
	EmbedURL     string    `json:"embed_url"`
	Duration     string    `json:"duration"`
	ViewCount    int64     `json:"view_count"`
	ChannelID    string    `json:"channel_id"`
	Tags         []string  `json:"tags"`
	Category     string    `json:"category"`
	Language     string    `json:"language"`
	Status       string    `json:"status"`
}

// NewMockDataLoader creates a new mock data loader
func NewMockDataLoader(dataDir string) *MockDataLoader {
	return &MockDataLoader{dataDir: dataDir}
}

// LoadYouTubeVideos loads YouTube videos from the JSON file
func (m *MockDataLoader) LoadYouTubeVideos() ([]database.YouTubeVideo, error) {
	filePath := filepath.Join(m.dataDir, "youtube_videos.json")

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YouTube videos file: %w", err)
	}

	var youtubeData YouTubeData
	if err := json.Unmarshal(data, &youtubeData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YouTube data: %w", err)
	}

	// Convert JSON structure to database structure
	var videos []database.YouTubeVideo
	for _, videoJSON := range youtubeData.Videos {
		video := database.YouTubeVideo{
			ID:           videoJSON.ID,
			Title:        videoJSON.Title,
			Description:  videoJSON.Description,
			PublishedAt:  videoJSON.PublishedAt,
			UpdatedAt:    videoJSON.UpdatedAt,
			ThumbnailURL: videoJSON.ThumbnailURL,
			VideoURL:     videoJSON.VideoURL,
			EmbedURL:     videoJSON.EmbedURL,
			Duration:     videoJSON.Duration,
			ViewCount:    videoJSON.ViewCount,
			CreatedAt:    videoJSON.CreatedAt,
		}
		videos = append(videos, video)
	}

	return videos, nil
}

// GetYouTubeChannelInfo loads channel information from the JSON file
func (m *MockDataLoader) GetYouTubeChannelInfo() (*ChannelInfo, error) {
	filePath := filepath.Join(m.dataDir, "youtube_videos.json")

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YouTube videos file: %w", err)
	}

	var youtubeData YouTubeData
	if err := json.Unmarshal(data, &youtubeData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YouTube data: %w", err)
	}

	return &ChannelInfo{
		ID:              youtubeData.Metadata.ChannelInfo.ID,
		Name:            youtubeData.Metadata.ChannelInfo.Name,
		Description:     youtubeData.Metadata.ChannelInfo.Description,
		SubscriberCount: youtubeData.Metadata.ChannelInfo.SubscriberCount,
		VideoCount:      youtubeData.Metadata.ChannelInfo.VideoCount,
		ViewCount:       youtubeData.Metadata.ChannelInfo.ViewCount,
		CreatedAt:       youtubeData.Metadata.ChannelInfo.CreatedAt,
		Country:         youtubeData.Metadata.ChannelInfo.Country,
		CustomURL:       youtubeData.Metadata.ChannelInfo.CustomURL,
	}, nil
}

// GetVideoCount returns the total number of videos in the mock data
func (m *MockDataLoader) GetVideoCount() (int, error) {
	videos, err := m.LoadYouTubeVideos()
	if err != nil {
		return 0, err
	}
	return len(videos), nil
}
