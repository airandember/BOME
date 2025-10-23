package services

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"bome-backend/infrastructure/database"
	"bome-backend/youtube/models"
)

// YouTubeService handles YouTube operations
type YouTubeService struct {
	db         *database.DB
	rssService *YouTubeRSSService
	channelID  string
}

// YouTubeVideosResponse represents the API response format
type YouTubeVideosResponse struct {
	Videos      []*models.YouTubeVideo `json:"videos"`
	LastUpdated time.Time              `json:"last_updated"`
	TotalCount  int                    `json:"total_count"`
	Channel     *models.ChannelInfo    `json:"channel,omitempty"`
}

// NewYouTubeService creates a new YouTube service
func NewYouTubeService(db *database.DB) *YouTubeService {
	// Get channel ID from config or use default
	config, err := models.GetYouTubeConfig(db)
	var channelID string
	if err != nil || config == nil {
		// Default to Book of Mormon Evidence channel if not configured
		channelID = "UCHp1EBgpKytZt_-j72EZ83Q"
		log.Printf("🎬 [YOUTUBE] Channel ID not configured, using default: %s", channelID)
	} else {
		channelID = config.ChannelID
		log.Printf("🎬 [YOUTUBE] Service using configured channel ID: %s", channelID)
	}

	// Initialize RSS service
	rssService := NewYouTubeRSSService(db, channelID)

	// Check if we have videos in database
	videoCount, err := models.GetYouTubeVideoCount(db)
	if err == nil && videoCount > 0 {
		log.Printf("🎬 [YOUTUBE] Service initialized with database (%d videos)", videoCount)
	} else {
		log.Printf("🎬 [YOUTUBE] Service initialized (database empty, ready for sync)")
	}

	return &YouTubeService{
		db:         db,
		rssService: rssService,
		channelID:  channelID,
	}
}

// ================================================
// VIDEO RETRIEVAL METHODS
// ================================================

// GetLatestVideos returns the latest YouTube videos sorted by newest first
func (y *YouTubeService) GetLatestVideos(limit int) (*YouTubeVideosResponse, error) {
	videos, err := models.GetYouTubeVideos(y.db, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest videos: %w", err)
	}

	return &YouTubeVideosResponse{
		Videos:      videos,
		LastUpdated: time.Now(),
		TotalCount:  len(videos),
	}, nil
}

// GetAllVideos returns all YouTube videos with pagination
func (y *YouTubeService) GetAllVideos(limit, offset int) (*YouTubeVideosResponse, error) {
	// Get total count
	totalCount, err := models.GetYouTubeVideoCount(y.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get video count: %w", err)
	}

	// Get videos with limit (offset would need to be added to model if needed)
	videos, err := models.GetYouTubeVideos(y.db, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get videos: %w", err)
	}

	return &YouTubeVideosResponse{
		Videos:      videos,
		LastUpdated: time.Now(),
		TotalCount:  totalCount,
	}, nil
}

// GetVideoByID returns a specific video by ID
func (y *YouTubeService) GetVideoByID(id string) (*models.YouTubeVideo, error) {
	video, err := models.GetYouTubeVideoByID(y.db, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	return video, nil
}

// GetVideosByCategory retrieves videos by category
func (y *YouTubeService) GetVideosByCategory(category string, limit int) (*YouTubeVideosResponse, error) {
	videos, err := models.GetYouTubeVideosByCategory(y.db, category, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get videos by category: %w", err)
	}

	return &YouTubeVideosResponse{
		Videos:      videos,
		LastUpdated: time.Now(),
		TotalCount:  len(videos),
	}, nil
}

// ================================================
// SEARCH METHODS
// ================================================

// SearchVideos searches for videos by query
func (y *YouTubeService) SearchVideos(query string, limit int) (*YouTubeVideosResponse, error) {
	videos, err := models.SearchYouTubeVideos(y.db, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search videos: %w", err)
	}

	return &YouTubeVideosResponse{
		Videos:      videos,
		LastUpdated: time.Now(),
		TotalCount:  len(videos),
	}, nil
}

// ================================================
// CATEGORY & TAG METHODS
// ================================================

// GetCategories returns all unique categories
func (y *YouTubeService) GetCategories() ([]string, error) {
	categories, err := models.GetAllCategories(y.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return categories, nil
}

// GetTags returns all unique tags
func (y *YouTubeService) GetTags() ([]string, error) {
	tags, err := models.GetAllTags(y.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	return tags, nil
}

// ================================================
// SYNC METHODS
// ================================================

// SyncFromRSS triggers an RSS feed sync
func (y *YouTubeService) SyncFromRSS() (*YouTubeSyncResult, error) {
	log.Printf("🔄 [YOUTUBE] Starting RSS sync...")

	// Create sync log entry
	syncLog := &models.YouTubeSyncLog{
		SyncType:  "rss",
		StartedAt: time.Now(),
		Status:    "running",
	}

	if err := models.CreateSyncLog(y.db, syncLog); err != nil {
		log.Printf("⚠️  [YOUTUBE] Failed to create sync log: %v", err)
	}

	// Perform the sync
	result, err := y.rssService.SyncVideosFromRSS()

	// Update sync log
	if result != nil {
		now := time.Now()
		syncLog.CompletedAt = &now
		syncLog.VideosFound = result.TotalFetched
		syncLog.VideosNew = result.NewVideos
		syncLog.VideosUpdated = result.UpdatedVideos
		syncLog.VideosSkipped = result.SkippedVideos

		if err != nil {
			syncLog.Status = "failed"
			syncLog.ErrorMessage = err.Error()
		} else if len(result.Errors) > 0 {
			syncLog.Status = "partial"
			syncLog.ErrorMessage = strings.Join(result.Errors, "; ")
		} else {
			syncLog.Status = "success"
		}

		if updateErr := models.UpdateSyncLog(y.db, syncLog); updateErr != nil {
			log.Printf("⚠️  [YOUTUBE] Failed to update sync log: %v", updateErr)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("RSS sync failed: %w", err)
	}

	log.Printf("✅ [YOUTUBE] RSS sync completed successfully")
	return result, nil
}

// GetSyncStatus returns information about the current sync status
func (y *YouTubeService) GetSyncStatus() (map[string]interface{}, error) {
	status, err := y.rssService.GetSyncStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to get sync status: %w", err)
	}

	// Add recent sync logs
	logs, err := models.GetRecentSyncLogs(y.db, 10)
	if err != nil {
		log.Printf("⚠️  [YOUTUBE] Failed to get recent sync logs: %v", err)
	} else {
		status["recent_syncs"] = logs
	}

	return status, nil
}

// ================================================
// CHANNEL INFO METHODS
// ================================================

// GetChannelInfo returns YouTube channel information
func (y *YouTubeService) GetChannelInfo() (*models.ChannelInfo, error) {
	config, err := models.GetYouTubeConfig(y.db)
	if err != nil || config == nil {
		// Return basic info from config
		return &models.ChannelInfo{
			ID:    y.channelID,
			Title: "Book of Mormon Evidence",
		}, nil
	}

	videoCount, _ := models.GetYouTubeVideoCount(y.db)

	return &models.ChannelInfo{
		ID:         config.ChannelID,
		Title:      config.ChannelTitle,
		VideoCount: videoCount,
	}, nil
}

// GetChannelStats returns channel statistics
func (y *YouTubeService) GetChannelStats() (map[string]interface{}, error) {
	totalVideos, err := models.GetYouTubeVideoCount(y.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get video count: %w", err)
	}

	// Get latest videos
	videos, err := models.GetYouTubeVideos(y.db, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest videos: %w", err)
	}

	// Calculate total views
	var totalViews int64
	var latestVideo *models.YouTubeVideo
	if len(videos) > 0 {
		latestVideo = videos[0]
		for _, video := range videos {
			totalViews += video.ViewCount
		}
	}

	// Get categories and tags
	categories, _ := models.GetAllCategories(y.db)
	tags, _ := models.GetAllTags(y.db)

	stats := map[string]interface{}{
		"channel_id":     y.channelID,
		"total_videos":   totalVideos,
		"total_views":    totalViews,
		"category_count": len(categories),
		"tag_count":      len(tags),
		"categories":     categories,
		"latest_video":   latestVideo,
	}

	return stats, nil
}

// ================================================
// STATUS METHODS
// ================================================

// GetStatus returns the service status
func (y *YouTubeService) GetStatus() (map[string]interface{}, error) {
	videoCount, err := models.GetYouTubeVideoCount(y.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get video count: %w", err)
	}

	config, _ := models.GetYouTubeConfig(y.db)
	latestLog, _ := models.GetLatestSyncLog(y.db)

	status := map[string]interface{}{
		"service":     "youtube",
		"status":      "online",
		"channel_id":  y.channelID,
		"video_count": videoCount,
		"data_source": "youtube_rss",
	}

	if config != nil {
		status["sync_enabled"] = config.SyncEnabled
		status["last_sync_at"] = config.LastSyncAt
	}

	if latestLog != nil {
		status["latest_sync"] = map[string]interface{}{
			"type":           latestLog.SyncType,
			"status":         latestLog.Status,
			"started_at":     latestLog.StartedAt,
			"completed_at":   latestLog.CompletedAt,
			"videos_new":     latestLog.VideosNew,
			"videos_updated": latestLog.VideosUpdated,
		}
	}

	return status, nil
}

// ================================================
// CONFIGURATION METHODS
// ================================================

// GetConfig returns the YouTube configuration
func (y *YouTubeService) GetConfig() (*models.YouTubeConfig, error) {
	config, err := models.GetYouTubeConfig(y.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	if config == nil {
		// Return default config
		return &models.YouTubeConfig{
			ChannelID:        y.channelID,
			ChannelTitle:     "Book of Mormon Evidence",
			RSSURL:           fmt.Sprintf("https://www.youtube.com/feeds/videos.xml?channel_id=%s", y.channelID),
			SyncEnabled:      true,
			SyncSchedule:     "0 14 * * *", // 2 PM daily
			AutoSyncToMaster: false,
		}, nil
	}

	return config, nil
}

// UpdateConfig updates the YouTube configuration
func (y *YouTubeService) UpdateConfig(config *models.YouTubeConfig) error {
	if err := models.UpdateYouTubeConfig(y.db, config); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	// Update service channel ID and RSS service
	y.channelID = config.ChannelID
	y.rssService.SetChannelID(config.ChannelID)

	log.Printf("✅ [YOUTUBE] Configuration updated: Channel=%s", config.ChannelID)
	return nil
}

// ================================================
// HELPER METHODS
// ================================================

// containsTag checks if a tag exists in a tag list (case-insensitive)
func containsTag(tags []string, query string) bool {
	query = strings.ToLower(query)
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

// sortVideosByDate sorts videos by published date (newest first)
func sortVideosByDate(videos []*models.YouTubeVideo) {
	sort.Slice(videos, func(i, j int) bool {
		return videos[i].PublishedAt.After(videos[j].PublishedAt)
	})
}
