package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"bome-backend/internal/database"
)

// YouTubeService handles YouTube operations - production ready but using mock data
type YouTubeService struct {
	mockDataPath string
}

// YouTubeMockData represents the structure of our mock JSON file
type YouTubeMockData struct {
	Feed struct {
		Title      string `json:"title"`
		ChannelID  string `json:"channel_id"`
		ChannelURL string `json:"channel_url"`
		Updated    string `json:"updated"`
	} `json:"feed"`
	Videos      []YouTubeVideoMock `json:"videos"`
	ChannelInfo struct {
		ID              string `json:"id"`
		Title           string `json:"title"`
		Description     string `json:"description"`
		SubscriberCount int    `json:"subscriber_count"`
		VideoCount      int    `json:"video_count"`
		ViewCount       int    `json:"view_count"`
		PublishedAt     string `json:"published_at"`
		Country         string `json:"country"`
		CustomURL       string `json:"custom_url"`
		ThumbnailURL    string `json:"thumbnail_url"`
	} `json:"channel_info"`
	Metadata struct {
		TotalVideos int    `json:"total_videos"`
		LastUpdated string `json:"last_updated"`
		APIVersion  string `json:"api_version"`
		MockData    bool   `json:"mock_data"`
	} `json:"metadata"`
}

// YouTubeVideoMock represents a video from our mock data
type YouTubeVideoMock struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Published    string   `json:"published"`
	Updated      string   `json:"updated"`
	ChannelID    string   `json:"channel_id"`
	ChannelTitle string   `json:"channel_title"`
	ThumbnailURL string   `json:"thumbnail_url"`
	VideoURL     string   `json:"video_url"`
	EmbedURL     string   `json:"embed_url"`
	Duration     string   `json:"duration"`
	ViewCount    int64    `json:"view_count"`
	LikeCount    int64    `json:"like_count"`
	CommentCount int64    `json:"comment_count"`
	Tags         []string `json:"tags"`
	CategoryID   string   `json:"category_id"`
	Category     string   `json:"category"`
	Language     string   `json:"language"`
	Status       string   `json:"status"`
}

// YouTubeVideosResponse represents the API response format
type YouTubeVideosResponse struct {
	Videos      []database.YouTubeVideo `json:"videos"`
	LastUpdated time.Time               `json:"last_updated"`
	TotalCount  int                     `json:"total_count"`
	Channel     *ChannelInfo            `json:"channel,omitempty"`
}

// ChannelInfo represents YouTube channel information
type ChannelInfo struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	SubscriberCount int       `json:"subscriber_count"`
	VideoCount      int       `json:"video_count"`
	ViewCount       int       `json:"view_count"`
	PublishedAt     time.Time `json:"published_at"`
	Country         string    `json:"country"`
	CustomURL       string    `json:"custom_url"`
	ThumbnailURL    string    `json:"thumbnail_url"`
}

// NewYouTubeService creates a new YouTube service
func NewYouTubeService(db *database.DB) *YouTubeService {
	// Mock data path - in production this would be replaced with real API calls
	mockDataPath := filepath.Join("internal", "MOCK_DATA", "YOUTUBE_MOCK.json")

	log.Printf("YouTube service initialized with mock data from: %s", mockDataPath)

	return &YouTubeService{
		mockDataPath: mockDataPath,
	}
}

// loadMockData loads the mock data from JSON file
func (y *YouTubeService) loadMockData() (*YouTubeMockData, error) {
	data, err := ioutil.ReadFile(y.mockDataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mock data file: %w", err)
	}

	var mockData YouTubeMockData
	if err := json.Unmarshal(data, &mockData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mock data: %w", err)
	}

	return &mockData, nil
}

// convertMockToDatabase converts mock video data to database format
func (y *YouTubeService) convertMockToDatabase(mockVideo YouTubeVideoMock) (database.YouTubeVideo, error) {
	published, err := time.Parse(time.RFC3339, mockVideo.Published)
	if err != nil {
		published = time.Now()
	}

	updated, err := time.Parse(time.RFC3339, mockVideo.Updated)
	if err != nil {
		updated = time.Now()
	}

	return database.YouTubeVideo{
		ID:           mockVideo.ID,
		Title:        mockVideo.Title,
		Description:  mockVideo.Description,
		PublishedAt:  published,
		UpdatedAt:    updated,
		ThumbnailURL: mockVideo.ThumbnailURL,
		VideoURL:     mockVideo.VideoURL,
		EmbedURL:     mockVideo.EmbedURL,
		Duration:     mockVideo.Duration,
		ViewCount:    mockVideo.ViewCount,
		Tags:         mockVideo.Tags,
		Category:     mockVideo.Category,
		CreatedAt:    published,
	}, nil
}

// GetLatestVideos returns the latest YouTube videos sorted by newest first
func (y *YouTubeService) GetLatestVideos(limit int) (*YouTubeVideosResponse, error) {
	mockData, err := y.loadMockData()
	if err != nil {
		return nil, err
	}

	var videos []database.YouTubeVideo
	for _, mockVideo := range mockData.Videos {
		video, err := y.convertMockToDatabase(mockVideo)
		if err != nil {
			log.Printf("Error converting mock video %s: %v", mockVideo.ID, err)
			continue
		}
		videos = append(videos, video)
	}

	// Sort videos by published date (newest first)
	sort.Slice(videos, func(i, j int) bool {
		return videos[i].PublishedAt.After(videos[j].PublishedAt)
	})

	// Apply limit if specified
	if limit > 0 && limit < len(videos) {
		videos = videos[:limit]
	}

	return &YouTubeVideosResponse{
		Videos:      videos,
		LastUpdated: time.Now(),
		TotalCount:  len(videos),
	}, nil
}

// GetVideoByID returns a specific video by ID
func (y *YouTubeService) GetVideoByID(id string) (*database.YouTubeVideo, error) {
	mockData, err := y.loadMockData()
	if err != nil {
		return nil, err
	}

	for _, mockVideo := range mockData.Videos {
		if mockVideo.ID == id {
			video, err := y.convertMockToDatabase(mockVideo)
			if err != nil {
				return nil, err
			}
			return &video, nil
		}
	}

	return nil, fmt.Errorf("video not found: %s", id)
}

// GetChannelInfo returns YouTube channel information
func (y *YouTubeService) GetChannelInfo() (*ChannelInfo, error) {
	mockData, err := y.loadMockData()
	if err != nil {
		return nil, err
	}

	publishedAt, err := time.Parse(time.RFC3339, mockData.ChannelInfo.PublishedAt)
	if err != nil {
		publishedAt = time.Now()
	}

	return &ChannelInfo{
		ID:              mockData.ChannelInfo.ID,
		Title:           mockData.ChannelInfo.Title,
		Description:     mockData.ChannelInfo.Description,
		SubscriberCount: mockData.ChannelInfo.SubscriberCount,
		VideoCount:      mockData.ChannelInfo.VideoCount,
		ViewCount:       mockData.ChannelInfo.ViewCount,
		PublishedAt:     publishedAt,
		Country:         mockData.ChannelInfo.Country,
		CustomURL:       mockData.ChannelInfo.CustomURL,
		ThumbnailURL:    mockData.ChannelInfo.ThumbnailURL,
	}, nil
}

// SearchVideos searches for videos by query
func (y *YouTubeService) SearchVideos(query string, limit int) (*YouTubeVideosResponse, error) {
	mockData, err := y.loadMockData()
	if err != nil {
		return nil, err
	}

	var matchedVideos []database.YouTubeVideo
	query = strings.ToLower(query)

	for _, mockVideo := range mockData.Videos {
		// Search in title, description, and tags
		if strings.Contains(strings.ToLower(mockVideo.Title), query) ||
			strings.Contains(strings.ToLower(mockVideo.Description), query) ||
			y.containsTag(mockVideo.Tags, query) {

			video, err := y.convertMockToDatabase(mockVideo)
			if err != nil {
				log.Printf("Error converting mock video %s: %v", mockVideo.ID, err)
				continue
			}
			matchedVideos = append(matchedVideos, video)
		}
	}

	// Sort by published date (newest first)
	sort.Slice(matchedVideos, func(i, j int) bool {
		return matchedVideos[i].PublishedAt.After(matchedVideos[j].PublishedAt)
	})

	// Apply limit if specified
	if limit > 0 && limit < len(matchedVideos) {
		matchedVideos = matchedVideos[:limit]
	}

	return &YouTubeVideosResponse{
		Videos:      matchedVideos,
		LastUpdated: time.Now(),
		TotalCount:  len(matchedVideos),
	}, nil
}

// GetVideosByCategory returns videos filtered by category
func (y *YouTubeService) GetVideosByCategory(category string, limit int) (*YouTubeVideosResponse, error) {
	mockData, err := y.loadMockData()
	if err != nil {
		return nil, err
	}

	var categoryVideos []database.YouTubeVideo
	category = strings.ToLower(category)

	for _, mockVideo := range mockData.Videos {
		if strings.ToLower(mockVideo.Category) == category {
			video, err := y.convertMockToDatabase(mockVideo)
			if err != nil {
				log.Printf("Error converting mock video %s: %v", mockVideo.ID, err)
				continue
			}
			categoryVideos = append(categoryVideos, video)
		}
	}

	// Sort by published date (newest first)
	sort.Slice(categoryVideos, func(i, j int) bool {
		return categoryVideos[i].PublishedAt.After(categoryVideos[j].PublishedAt)
	})

	// Apply limit if specified
	if limit > 0 && limit < len(categoryVideos) {
		categoryVideos = categoryVideos[:limit]
	}

	return &YouTubeVideosResponse{
		Videos:      categoryVideos,
		LastUpdated: time.Now(),
		TotalCount:  len(categoryVideos),
	}, nil
}

// GetStatus returns the current status of the YouTube integration
func (y *YouTubeService) GetStatus() (map[string]interface{}, error) {
	mockData, err := y.loadMockData()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"channel_id":    mockData.Feed.ChannelID,
		"channel_title": mockData.ChannelInfo.Title,
		"total_videos":  mockData.Metadata.TotalVideos,
		"last_updated":  mockData.Metadata.LastUpdated,
		"api_version":   mockData.Metadata.APIVersion,
		"mock_mode":     mockData.Metadata.MockData,
		"status":        "active",
		"data_source":   "mock_json",
	}, nil
}

// containsTag checks if any tag contains the search query
func (y *YouTubeService) containsTag(tags []string, query string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

// GetAllCategories returns all unique categories from videos
func (y *YouTubeService) GetAllCategories() ([]string, error) {
	mockData, err := y.loadMockData()
	if err != nil {
		return nil, err
	}

	categorySet := make(map[string]bool)
	for _, video := range mockData.Videos {
		categorySet[video.Category] = true
	}

	var categories []string
	for category := range categorySet {
		categories = append(categories, category)
	}

	sort.Strings(categories)
	return categories, nil
}

// GetAllTags returns all unique tags from videos
func (y *YouTubeService) GetAllTags() ([]string, error) {
	mockData, err := y.loadMockData()
	if err != nil {
		return nil, err
	}

	tagSet := make(map[string]bool)
	for _, video := range mockData.Videos {
		for _, tag := range video.Tags {
			tagSet[tag] = true
		}
	}

	var tags []string
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	sort.Strings(tags)
	return tags, nil
}
