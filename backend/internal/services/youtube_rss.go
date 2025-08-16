package services

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"bome-backend/internal/database"
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

	resp, err := y.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RSS feed request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read RSS feed response: %w", err)
	}

	var feed YouTubeRSSFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("failed to parse RSS feed XML: %w", err)
	}

	return &feed, nil
}

// ConvertRSSEntryToVideo converts an RSS entry to a database video record
func (y *YouTubeRSSService) ConvertRSSEntryToVideo(entry YouTubeRSSEntry) (database.YouTubeVideo, error) {
	// Extract video ID from the entry ID (format: yt:video:VIDEO_ID)
	videoID := strings.TrimPrefix(entry.ID, "yt:video:")
	if videoID == entry.ID {
		// Fallback: try to extract from VideoID field
		videoID = entry.VideoID
	}

	// Parse timestamps
	published, err := time.Parse(time.RFC3339, entry.Published)
	if err != nil {
		published = time.Now()
	}

	updated, err := time.Parse(time.RFC3339, entry.Updated)
	if err != nil {
		updated = time.Now()
	}

	// Generate URLs
	videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
	embedURL := fmt.Sprintf("https://www.youtube.com/embed/%s", videoID)
	thumbnailURL := entry.Thumbnail.URL
	if thumbnailURL == "" {
		thumbnailURL = fmt.Sprintf("https://i.ytimg.com/vi/%s/hqdefault.jpg", videoID)
	}

	return database.YouTubeVideo{
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
	feed, err := y.FetchRSSFeed()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS feed: %w", err)
	}

	result := &YouTubeSyncResult{
		TotalFetched:  len(feed.Entries),
		NewVideos:     0,
		UpdatedVideos: 0,
		Errors:        []string{},
		SyncTime:      time.Now(),
	}

	for _, entry := range feed.Entries {
		video, err := y.ConvertRSSEntryToVideo(entry)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to convert entry %s: %v", entry.ID, err))
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
		} else {
			// Video exists, update it if needed
			if existingVideo.UpdatedAt.Before(video.UpdatedAt) {
				if err := y.db.UpdateYouTubeVideo(video); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to update video %s: %v", video.ID, err))
					continue
				}
				result.UpdatedVideos++
			}
		}
	}

	return result, nil
}

// YouTubeSyncResult represents the result of a sync operation
type YouTubeSyncResult struct {
	TotalFetched  int       `json:"total_fetched"`
	NewVideos     int       `json:"new_videos"`
	UpdatedVideos int       `json:"updated_videos"`
	Errors        []string  `json:"errors"`
	SyncTime      time.Time `json:"sync_time"`
}

// GetSyncStatus returns information about the current sync status
func (y *YouTubeRSSService) GetSyncStatus() (map[string]interface{}, error) {
	totalVideos, err := y.db.GetYouTubeVideoCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get video count: %w", err)
	}

	// Get latest video to determine last sync
	videos, err := y.db.GetYouTubeVideos(1)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest video: %w", err)
	}

	var lastSync *time.Time
	if len(videos) > 0 {
		lastSync = &videos[0].CreatedAt
	}

	return map[string]interface{}{
		"channel_id":   y.channelID,
		"total_videos": totalVideos,
		"last_sync":    lastSync,
		"sync_enabled": true,
		"data_source":  "youtube_rss",
		"rss_feed_url": fmt.Sprintf("https://www.youtube.com/feeds/videos.xml?channel_id=%s", y.channelID),
	}, nil
}
