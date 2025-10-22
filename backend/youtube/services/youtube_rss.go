package services

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"bome-backend/infrastructure/database"
	"bome-backend/youtube/models"
)

// YouTubeRSSService handles fetching and parsing YouTube RSS feeds
type YouTubeRSSService struct {
	db        *database.DB
	channelID string
	client    *http.Client
}

// YouTubeRSSFeed represents the structure of YouTube RSS feed
type YouTubeRSSFeed struct {
	XMLName xml.Name `xml:"feed"`
	Title   string   `xml:"title"`
	Updated string   `xml:"updated"`
	Author  struct {
		Name string `xml:"name"`
		URI  string `xml:"uri"`
	} `xml:"author"`
	Entries []YouTubeRSSEntry `xml:"entry"`
}

// YouTubeRSSEntry represents a single video entry in the RSS feed
type YouTubeRSSEntry struct {
	ID        string `xml:"id"`
	VideoID   string `xml:"videoId"`
	Title     string `xml:"title"`
	Published string `xml:"published"`
	Updated   string `xml:"updated"`
	Author    struct {
		Name string `xml:"name"`
		URI  string `xml:"uri"`
	} `xml:"author"`
	Link struct {
		Href string `xml:"href,attr"`
	} `xml:"link"`
	Thumbnail struct {
		URL    string `xml:"url,attr"`
		Width  string `xml:"width,attr"`
		Height string `xml:"height,attr"`
	} `xml:"thumbnail"`
	Description string `xml:"description"`
	Views       string `xml:"views"`
	Rating      string `xml:"rating"`
}

// YouTubeSyncResult represents the result of a sync operation
type YouTubeSyncResult struct {
	TotalFetched  int       `json:"total_fetched"`
	NewVideos     int       `json:"new_videos"`
	UpdatedVideos int       `json:"updated_videos"`
	SkippedVideos int       `json:"skipped_videos"`
	Errors        []string  `json:"errors"`
	SyncTime      time.Time `json:"sync_time"`
	Duration      string    `json:"duration"`
}

// NewYouTubeRSSService creates a new YouTube RSS service
func NewYouTubeRSSService(db *database.DB, channelID string) *YouTubeRSSService {
	return &YouTubeRSSService{
		db:        db,
		channelID: channelID,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchRSSFeed fetches and parses the YouTube RSS feed
func (y *YouTubeRSSService) FetchRSSFeed() (*YouTubeRSSFeed, error) {
	url := fmt.Sprintf("https://www.youtube.com/feeds/videos.xml?channel_id=%s", y.channelID)
	log.Printf("🌐 [YOUTUBE-RSS] Fetching RSS feed from: %s", url)

	resp, err := y.client.Get(url)
	if err != nil {
		log.Printf("❌ [YOUTUBE-RSS] HTTP request failed: %v", err)
		return nil, fmt.Errorf("failed to fetch RSS feed: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("📡 [YOUTUBE-RSS] HTTP response status: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ [YOUTUBE-RSS] Bad HTTP status: %d", resp.StatusCode)
		return nil, fmt.Errorf("RSS feed request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ [YOUTUBE-RSS] Failed to read response body: %v", err)
		return nil, fmt.Errorf("failed to read RSS feed response: %w", err)
	}

	log.Printf("📄 [YOUTUBE-RSS] Response body length: %d bytes", len(body))

	var feed YouTubeRSSFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		log.Printf("❌ [YOUTUBE-RSS] Failed to parse XML: %v", err)
		maxLen := 500
		if len(body) < maxLen {
			maxLen = len(body)
		}
		log.Printf("📄 [YOUTUBE-RSS] Raw XML (first %d chars): %s", maxLen, string(body[:maxLen]))
		return nil, fmt.Errorf("failed to parse RSS feed XML: %w", err)
	}

	log.Printf("✅ [YOUTUBE-RSS] Successfully parsed RSS feed: %d entries", len(feed.Entries))
	return &feed, nil
}

// ConvertRSSEntryToVideo converts an RSS entry to a database video record
func (y *YouTubeRSSService) ConvertRSSEntryToVideo(entry YouTubeRSSEntry) (*models.YouTubeVideo, error) {
	// Extract video ID from the entry ID (format: yt:video:VIDEO_ID)
	videoID := strings.TrimPrefix(entry.ID, "yt:video:")
	if videoID == entry.ID {
		// Fallback: try to extract from VideoID field
		videoID = entry.VideoID
	}

	if videoID == "" {
		return nil, fmt.Errorf("no video ID found in entry")
	}

	// Parse timestamps
	published, err := time.Parse(time.RFC3339, entry.Published)
	if err != nil {
		log.Printf("⚠️  [YOUTUBE-RSS] Failed to parse published date: %v, using now", err)
		published = time.Now()
	}

	updated, err := time.Parse(time.RFC3339, entry.Updated)
	if err != nil {
		log.Printf("⚠️  [YOUTUBE-RSS] Failed to parse updated date: %v, using now", err)
		updated = time.Now()
	}

	// Generate URLs
	videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
	embedURL := fmt.Sprintf("https://www.youtube.com/embed/%s", videoID)
	thumbnailURL := entry.Thumbnail.URL
	if thumbnailURL == "" {
		thumbnailURL = fmt.Sprintf("https://i.ytimg.com/vi/%s/hqdefault.jpg", videoID)
	}

	return &models.YouTubeVideo{
		ID:           videoID,
		Title:        entry.Title,
		Description:  entry.Description,
		PublishedAt:  published,
		UpdatedAt:    updated,
		ThumbnailURL: thumbnailURL,
		VideoURL:     videoURL,
		EmbedURL:     embedURL,
		Duration:     "",         // RSS feed doesn't provide duration
		ViewCount:    0,          // RSS feed doesn't provide view count
		Tags:         []string{}, // RSS feed doesn't provide tags
		Category:     "General",  // Default category
		CreatedAt:    time.Now(),
	}, nil
}

// SyncVideosFromRSS fetches RSS feed and saves new videos to database
func (y *YouTubeRSSService) SyncVideosFromRSS() (*YouTubeSyncResult, error) {
	startTime := time.Now()
	log.Printf("🔄 [YOUTUBE-RSS] Starting RSS sync for channel: %s", y.channelID)

	feed, err := y.FetchRSSFeed()
	if err != nil {
		log.Printf("❌ [YOUTUBE-RSS] Failed to fetch RSS feed: %v", err)
		return nil, fmt.Errorf("failed to fetch RSS feed: %w", err)
	}

	log.Printf("📺 [YOUTUBE-RSS] Fetched RSS feed with %d entries", len(feed.Entries))

	result := &YouTubeSyncResult{
		TotalFetched:  len(feed.Entries),
		NewVideos:     0,
		UpdatedVideos: 0,
		SkippedVideos: 0,
		Errors:        []string{},
		SyncTime:      startTime,
	}

	for i, entry := range feed.Entries {
		log.Printf("🎬 [YOUTUBE-RSS] Processing entry %d/%d: %s", i+1, len(feed.Entries), entry.Title)

		video, err := y.ConvertRSSEntryToVideo(entry)
		if err != nil {
			log.Printf("❌ [YOUTUBE-RSS] Failed to convert entry %s: %v", entry.ID, err)
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to convert entry %s: %v", entry.ID, err))
			continue
		}

		log.Printf("📝 [YOUTUBE-RSS] Converted video: ID=%s, Title=%s", video.ID, video.Title)

		// Check if video already exists
		existingVideo, err := models.GetYouTubeVideoByID(y.db, video.ID)
		if err != nil && err.Error() != "video not found" {
			log.Printf("❌ [YOUTUBE-RSS] Failed to check existing video %s: %v", video.ID, err)
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to check existing video %s: %v", video.ID, err))
			continue
		}

		if existingVideo == nil {
			// Video doesn't exist, create it
			log.Printf("➕ [YOUTUBE-RSS] Creating new video: %s", video.ID)
			if err := models.CreateYouTubeVideo(y.db, video); err != nil {
				log.Printf("❌ [YOUTUBE-RSS] Failed to create video %s: %v", video.ID, err)
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to create video %s: %v", video.ID, err))
				continue
			}
			result.NewVideos++
			log.Printf("✅ [YOUTUBE-RSS] Successfully created video: %s", video.ID)
		} else {
			// Video exists, update it if needed
			if existingVideo.UpdatedAt.Before(video.UpdatedAt) {
				log.Printf("🔄 [YOUTUBE-RSS] Updating existing video: %s", video.ID)
				// Preserve existing fields
				video.CreatedAt = existingVideo.CreatedAt
				if video.Duration == "" {
					video.Duration = existingVideo.Duration
				}
				if video.ViewCount == 0 {
					video.ViewCount = existingVideo.ViewCount
				}
				if len(video.Tags) == 0 {
					video.Tags = existingVideo.Tags
				}
				if video.Category == "General" && existingVideo.Category != "" {
					video.Category = existingVideo.Category
				}

				if err := models.UpdateYouTubeVideo(y.db, video); err != nil {
					log.Printf("❌ [YOUTUBE-RSS] Failed to update video %s: %v", video.ID, err)
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to update video %s: %v", video.ID, err))
					continue
				}
				result.UpdatedVideos++
				log.Printf("✅ [YOUTUBE-RSS] Successfully updated video: %s", video.ID)
			} else {
				log.Printf("⏭️  [YOUTUBE-RSS] Video %s is up to date, skipping", video.ID)
				result.SkippedVideos++
			}
		}
	}

	duration := time.Since(startTime)
	result.Duration = duration.String()

	log.Printf("📊 [YOUTUBE-RSS] Sync completed in %s: %d fetched, %d new, %d updated, %d skipped, %d errors",
		result.Duration, result.TotalFetched, result.NewVideos, result.UpdatedVideos, result.SkippedVideos, len(result.Errors))

	if len(result.Errors) > 0 {
		log.Printf("⚠️  [YOUTUBE-RSS] Errors encountered:")
		for _, errMsg := range result.Errors {
			log.Printf("   - %s", errMsg)
		}
	}

	// Update last sync time in config
	if err := models.UpdateLastSyncTime(y.db); err != nil {
		log.Printf("⚠️  [YOUTUBE-RSS] Failed to update last sync time: %v", err)
	}

	return result, nil
}

// GetSyncStatus returns information about the current sync status
func (y *YouTubeRSSService) GetSyncStatus() (map[string]interface{}, error) {
	totalVideos, err := models.GetYouTubeVideoCount(y.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get video count: %w", err)
	}

	// Get latest video to determine last sync
	videos, err := models.GetYouTubeVideos(y.db, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest video: %w", err)
	}

	var lastVideoPublished *time.Time
	if len(videos) > 0 {
		lastVideoPublished = &videos[0].PublishedAt
	}

	// Get config for last sync time
	config, err := models.GetYouTubeConfig(y.db)
	if err != nil {
		log.Printf("⚠️  [YOUTUBE-RSS] Failed to get config: %v", err)
	}

	// Get latest sync log
	latestLog, err := models.GetLatestSyncLog(y.db)
	if err != nil {
		log.Printf("⚠️  [YOUTUBE-RSS] Failed to get latest sync log: %v", err)
	}

	status := map[string]interface{}{
		"channel_id":           y.channelID,
		"total_videos":         totalVideos,
		"last_video_published": lastVideoPublished,
		"sync_enabled":         true,
		"data_source":          "youtube_rss",
		"rss_feed_url":         fmt.Sprintf("https://www.youtube.com/feeds/videos.xml?channel_id=%s", y.channelID),
	}

	if config != nil {
		status["last_sync_at"] = config.LastSyncAt
		status["sync_schedule"] = config.SyncSchedule
	}

	if latestLog != nil {
		status["latest_sync"] = map[string]interface{}{
			"sync_type":      latestLog.SyncType,
			"started_at":     latestLog.StartedAt,
			"completed_at":   latestLog.CompletedAt,
			"videos_found":   latestLog.VideosFound,
			"videos_new":     latestLog.VideosNew,
			"videos_updated": latestLog.VideosUpdated,
			"status":         latestLog.Status,
		}
	}

	return status, nil
}

// GetChannelID returns the configured channel ID
func (y *YouTubeRSSService) GetChannelID() string {
	return y.channelID
}

// SetChannelID updates the channel ID
func (y *YouTubeRSSService) SetChannelID(channelID string) {
	y.channelID = channelID
}
