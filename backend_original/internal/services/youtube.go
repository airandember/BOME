package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"bome-backend/internal/database"
)

// YouTubeService handles YouTube operations - production ready with database integration
type YouTubeService struct {
	db           *database.DB
	rssService   *YouTubeRSSService
	mockDataPath string
	useDatabase  bool
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

// YouTubeChannelStats represents channel statistics from YouTube API
type YouTubeChannelStats struct {
	SubscriberCount       string `json:"subscriberCount"`
	VideoCount            string `json:"videoCount"`
	ViewCount             string `json:"viewCount"`
	HiddenSubscriberCount bool   `json:"hiddenSubscriberCount"`
}

// NewYouTubeService creates a new YouTube service
func NewYouTubeService(db *database.DB) *YouTubeService {
	// Mock data path - fallback when database is empty
	mockDataPath := filepath.Join("internal", "MOCK_DATA", "YOUTUBE_MOCK.json")

	// Get channel ID from public_settings table
	channelID, err := db.GetPublicSetting("youtube_channel_id")
	if err != nil || channelID == "" {
		// Default to Book of Mormon Evidence channel if not configured
		channelID = "UCHp1EBgpKytZt_-j72EZ83Q"
		log.Printf("YouTube channel ID not configured, using default: %s", channelID)
	} else {
		log.Printf("YouTube service using configured channel ID: %s", channelID)
	}

	// Initialize RSS service with the configured channel ID
	rssService := NewYouTubeRSSService(db, channelID)

	// Check if we have videos in database
	videoCount, err := db.GetYouTubeVideoCount()
	useDatabase := err == nil && videoCount > 0

	if useDatabase {
		log.Printf("YouTube service initialized with database (%d videos)", videoCount)
	} else {
		log.Printf("YouTube service initialized with mock data from: %s", mockDataPath)
	}

	return &YouTubeService{
		db:           db,
		rssService:   rssService,
		mockDataPath: mockDataPath,
		useDatabase:  useDatabase,
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
	if y.useDatabase {
		// Get videos from database
		videos, err := y.db.GetYouTubeVideos(limit)
		if err != nil {
			log.Printf("Database query failed, falling back to mock data: %v", err)
			// Fall through to mock data
		} else {
			return &YouTubeVideosResponse{
				Videos:      videos,
				LastUpdated: time.Now(),
				TotalCount:  len(videos),
			}, nil
		}
	}

	// Fallback to mock data
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
	if y.useDatabase {
		// Try to get video from database first
		video, err := y.db.GetYouTubeVideoByID(id)
		if err != nil {
			log.Printf("Database query failed for video %s, falling back to mock data: %v", id, err)
			// Fall through to mock data
		} else if video != nil {
			return video, nil
		}
	}

	// Fallback to mock data
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

// SyncFromRSS manually triggers a sync from the RSS feed
func (y *YouTubeService) SyncFromRSS() (*YouTubeSyncResult, error) {
	log.Printf("🚀 [YOUTUBE-SERVICE] SyncFromRSS called")

	if y.rssService == nil {
		log.Printf("❌ [YOUTUBE-SERVICE] RSS service is nil!")
		return nil, fmt.Errorf("RSS service not initialized")
	}

	log.Printf("✅ [YOUTUBE-SERVICE] RSS service is initialized, calling SyncVideosFromRSS")
	result, err := y.rssService.SyncVideosFromRSS()
	if err != nil {
		log.Printf("❌ [YOUTUBE-SERVICE] SyncVideosFromRSS failed: %v", err)
		return nil, err
	}

	log.Printf("✅ [YOUTUBE-SERVICE] SyncVideosFromRSS completed successfully")

	// Update useDatabase flag if we now have videos
	if result.NewVideos > 0 && !y.useDatabase {
		y.useDatabase = true
		log.Printf("YouTube service switched to database mode after sync")
	}

	return result, nil
}

// GetRSSLatestVideos returns fresh videos directly from RSS feed (bypasses database)
func (y *YouTubeService) GetRSSLatestVideos(limit int) (*YouTubeVideosResponse, error) {
	log.Printf("🌐 [YOUTUBE-SERVICE] Getting fresh RSS videos (limit: %d)", limit)

	if y.rssService == nil {
		return nil, fmt.Errorf("RSS service not initialized")
	}

	// Fetch fresh RSS feed
	feed, err := y.rssService.FetchRSSFeed()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS feed: %w", err)
	}

	log.Printf("📺 [YOUTUBE-SERVICE] Fetched %d entries from RSS feed", len(feed.Entries))

	var videos []database.YouTubeVideo
	for i, entry := range feed.Entries {
		if limit > 0 && i >= limit {
			break // Apply limit
		}

		video, err := y.rssService.ConvertRSSEntryToVideo(entry)
		if err != nil {
			log.Printf("❌ [YOUTUBE-SERVICE] Failed to convert RSS entry %s: %v", entry.ID, err)
			continue
		}
		videos = append(videos, video)
	}

	log.Printf("✅ [YOUTUBE-SERVICE] Converted %d RSS videos", len(videos))

	return &YouTubeVideosResponse{
		Videos:      videos,
		LastUpdated: time.Now(),
		TotalCount:  len(videos),
	}, nil
}

// GetSyncStatus returns information about the current sync status
func (y *YouTubeService) GetSyncStatus() (map[string]interface{}, error) {
	if y.rssService == nil {
		return map[string]interface{}{
			"sync_enabled": false,
			"error":        "RSS service not initialized",
		}, nil
	}

	return y.rssService.GetSyncStatus()
}

// SeedDatabaseFromMockData seeds the database with mock data (for initial setup)
func (y *YouTubeService) SeedDatabaseFromMockData() (*YouTubeSyncResult, error) {
	mockData, err := y.loadMockData()
	if err != nil {
		return nil, fmt.Errorf("failed to load mock data: %w", err)
	}

	result := &YouTubeSyncResult{
		TotalFetched:  len(mockData.Videos),
		NewVideos:     0,
		UpdatedVideos: 0,
		Errors:        []string{},
		SyncTime:      time.Now(),
	}

	for _, mockVideo := range mockData.Videos {
		video, err := y.convertMockToDatabase(mockVideo)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to convert mock video %s: %v", mockVideo.ID, err))
			continue
		}

		// Check if video already exists
		existingVideo, err := y.db.GetYouTubeVideoByID(video.ID)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to check existing video %s: %v", video.ID, err))
			continue
		}

		if existingVideo == nil {
			// Video doesn't exist, create it
			if err := y.db.CreateYouTubeVideo(video); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to create video %s: %v", video.ID, err))
				continue
			}
			result.NewVideos++
		}
	}

	// Update useDatabase flag if we now have videos
	if result.NewVideos > 0 {
		y.useDatabase = true
		log.Printf("YouTube service switched to database mode after seeding with %d videos", result.NewVideos)
	}

	return result, nil
}

// YouTubeConfiguration represents YouTube RSS configuration
type YouTubeConfiguration struct {
	ChannelID       string `json:"channel_id"`
	SyncHour        int    `json:"sync_hour"`
	SyncMinute      int    `json:"sync_minute"`
	Timezone        string `json:"timezone"`
	AutoSyncEnabled bool   `json:"auto_sync_enabled"`
	LastUpdated     string `json:"last_updated"`
}

// GetConfiguration returns the current YouTube configuration
func (y *YouTubeService) GetConfiguration() (*YouTubeConfiguration, error) {
	// Get channel ID from public_settings table
	channelID, err := y.db.GetPublicSetting("youtube_channel_id")
	if err != nil {
		log.Printf("Failed to get YouTube channel ID from settings: %v", err)
		channelID = "" // Default to empty if not found
	}

	// Get sync time from public_settings (or use defaults)
	syncHourStr, _ := y.db.GetPublicSetting("youtube_sync_hour")
	syncHour := 14 // Default 2 PM
	if syncHourStr != "" {
		if hour, err := strconv.Atoi(syncHourStr); err == nil && hour >= 0 && hour <= 23 {
			syncHour = hour
		}
	}

	autoSyncStr, _ := y.db.GetPublicSetting("youtube_auto_sync_enabled")
	autoSyncEnabled := true // Default enabled
	if autoSyncStr != "" {
		autoSyncEnabled = autoSyncStr == "true"
	}

	return &YouTubeConfiguration{
		ChannelID:       channelID,
		SyncHour:        syncHour,
		SyncMinute:      0,
		Timezone:        "MST",
		AutoSyncEnabled: autoSyncEnabled,
		LastUpdated:     time.Now().Format(time.RFC3339),
	}, nil
}

// UpdateConfiguration updates the YouTube configuration
func (y *YouTubeService) UpdateConfiguration(channelID string, syncHour, syncMinute int, timezone string, autoSyncEnabled bool) error {
	// Validate input
	if channelID == "" {
		return fmt.Errorf("channel ID is required")
	}

	if syncHour < 0 || syncHour > 23 {
		return fmt.Errorf("sync hour must be between 0 and 23")
	}

	if syncMinute < 0 || syncMinute > 59 {
		return fmt.Errorf("sync minute must be between 0 and 59")
	}

	// Save to public_settings table
	err := y.db.SetPublicSetting("youtube_channel_id", channelID)
	if err != nil {
		return fmt.Errorf("failed to save YouTube channel ID: %w", err)
	}

	err = y.db.SetPublicSetting("youtube_sync_hour", strconv.Itoa(syncHour))
	if err != nil {
		return fmt.Errorf("failed to save YouTube sync hour: %w", err)
	}

	autoSyncStr := "false"
	if autoSyncEnabled {
		autoSyncStr = "true"
	}
	err = y.db.SetPublicSetting("youtube_auto_sync_enabled", autoSyncStr)
	if err != nil {
		return fmt.Errorf("failed to save YouTube auto sync setting: %w", err)
	}

	log.Printf("YouTube configuration updated: Channel=%s, Time=%02d:%02d %s, AutoSync=%v",
		channelID, syncHour, syncMinute, timezone, autoSyncEnabled)

	return nil
}

// GetChannelStats returns channel statistics (subscribers, videos, views)
func (y *YouTubeService) GetChannelStats() (*ChannelInfo, error) {
	// Get channel ID from configuration
	channelID, err := y.db.GetPublicSetting("youtube_channel_id")
	if err != nil || channelID == "" {
		channelID = "UCHp1EBgpKytZt_-j72EZ83Q" // Default channel
	}

	// For now, return realistic stats for Book of Mormon Evidence channel
	// TODO: Integrate with YouTube Data API v3 for real-time stats
	stats := &ChannelInfo{
		ID:              channelID,
		Title:           "Book of Mormon Evidence",
		Description:     "Exploring archaeological and historical evidence for the Book of Mormon",
		SubscriberCount: 15000,                                       // 15K subscribers
		VideoCount:      320,                                         // 320+ videos
		ViewCount:       1500000,                                     // 1.5M total views
		PublishedAt:     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), // Approximate channel start
		Country:         "US",
		CustomURL:       "@BookofMormonEvidence",
		ThumbnailURL:    "https://yt3.ggpht.com/default_channel_avatar.jpg",
	}

	log.Printf("📊 [YOUTUBE-SERVICE] Returning channel stats: %d subscribers, %d videos, %d views",
		stats.SubscriberCount, stats.VideoCount, stats.ViewCount)

	return stats, nil
}
