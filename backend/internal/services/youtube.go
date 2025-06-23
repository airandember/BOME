package services

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"bome-backend/internal/database"
)

// YouTubeService handles YouTube PubSubHubbub webhooks and video management
type YouTubeService struct {
	channelID   string
	callbackURL string
	verifyToken string
	hubURL      string
	client      *http.Client
	db          *database.DB
}

// YouTubeFeed represents the XML feed from YouTube PubSubHubbub
type YouTubeFeed struct {
	XMLName xml.Name       `xml:"feed"`
	Entries []YouTubeEntry `xml:"entry"`
}

// YouTubeEntry represents an individual video entry in the feed
type YouTubeEntry struct {
	ID        string    `xml:"id"`
	VideoID   string    `xml:"videoId"`
	ChannelID string    `xml:"channelId"`
	Title     string    `xml:"title"`
	Published time.Time `xml:"published"`
	Updated   time.Time `xml:"updated"`
	Author    struct {
		Name string `xml:"name"`
		URI  string `xml:"uri"`
	} `xml:"author"`
	Link struct {
		Href string `xml:"href,attr"`
	} `xml:"link"`
}

// YouTubeVideosResponse represents the JSON response for frontend
type YouTubeVideosResponse struct {
	Videos      []database.YouTubeVideo `json:"videos"`
	LastUpdated time.Time               `json:"last_updated"`
	TotalCount  int                     `json:"total_count"`
}

// NewYouTubeService creates a new YouTube webhook service
func NewYouTubeService(db *database.DB) *YouTubeService {
	return &YouTubeService{
		channelID:   "UCHp1EBgpKytZt_-j72EZ83Q", // Book of Mormon Evidence channel
		callbackURL: os.Getenv("YOUTUBE_WEBHOOK_URL"),
		verifyToken: os.Getenv("YOUTUBE_VERIFY_TOKEN"),
		hubURL:      "https://pubsubhubbub.appspot.com/subscribe",
		client:      &http.Client{Timeout: 30 * time.Second},
		db:          db,
	}
}

// Subscribe subscribes to YouTube channel updates via PubSubHubbub
func (y *YouTubeService) Subscribe() error {
	topicURL := fmt.Sprintf("https://www.youtube.com/xml/feeds/videos.xml?channel_id=%s", y.channelID)

	data := url.Values{}
	data.Set("hub.callback", y.callbackURL)
	data.Set("hub.topic", topicURL)
	data.Set("hub.verify", "async")
	data.Set("hub.mode", "subscribe")
	data.Set("hub.verify_token", y.verifyToken)
	data.Set("hub.lease_seconds", "864000") // 10 days

	req, err := http.NewRequest("POST", y.hubURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create subscription request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := y.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send subscription request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("subscription failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Successfully subscribed to YouTube channel %s", y.channelID)
	return nil
}

// Unsubscribe unsubscribes from YouTube channel updates
func (y *YouTubeService) Unsubscribe() error {
	topicURL := fmt.Sprintf("https://www.youtube.com/xml/feeds/videos.xml?channel_id=%s", y.channelID)

	data := url.Values{}
	data.Set("hub.callback", y.callbackURL)
	data.Set("hub.topic", topicURL)
	data.Set("hub.verify", "async")
	data.Set("hub.mode", "unsubscribe")
	data.Set("hub.verify_token", y.verifyToken)

	req, err := http.NewRequest("POST", y.hubURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create unsubscription request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := y.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send unsubscription request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("Successfully unsubscribed from YouTube channel %s", y.channelID)
	return nil
}

// HandleVerification handles the webhook verification challenge
func (y *YouTubeService) HandleVerification(w http.ResponseWriter, r *http.Request) {
	challenge := r.URL.Query().Get("hub.challenge")
	mode := r.URL.Query().Get("hub.mode")
	topic := r.URL.Query().Get("hub.topic")
	verifyToken := r.URL.Query().Get("hub.verify_token")

	log.Printf("YouTube webhook verification: mode=%s, topic=%s, token=%s", mode, topic, verifyToken)

	// Verify the request
	expectedTopic := fmt.Sprintf("https://www.youtube.com/xml/feeds/videos.xml?channel_id=%s", y.channelID)
	if topic != expectedTopic {
		log.Printf("Invalid topic: expected %s, got %s", expectedTopic, topic)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if verifyToken != y.verifyToken {
		log.Printf("Invalid verify token: expected %s, got %s", y.verifyToken, verifyToken)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if mode == "subscribe" || mode == "unsubscribe" {
		log.Printf("YouTube webhook %s verification successful", mode)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
		return
	}

	w.WriteHeader(http.StatusForbidden)
}

// HandleNotification handles new video notifications from YouTube
func (y *YouTubeService) HandleNotification(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read notification body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("Received YouTube notification: %s", string(body))

	var feed YouTubeFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		log.Printf("Failed to parse XML feed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Process each video entry
	for _, entry := range feed.Entries {
		if err := y.processVideoEntry(entry); err != nil {
			log.Printf("Failed to process video entry %s: %v", entry.VideoID, err)
		}
	}

	// Update the cached JSON file
	if err := y.updateVideoCache(); err != nil {
		log.Printf("Failed to update video cache: %v", err)
	}

	w.WriteHeader(http.StatusOK)
}

// processVideoEntry processes a single video entry from the feed
func (y *YouTubeService) processVideoEntry(entry YouTubeEntry) error {
	// Extract video ID from the entry ID
	videoID := entry.VideoID
	if videoID == "" {
		// Fallback: extract from entry.ID if videoId is not present
		parts := strings.Split(entry.ID, ":")
		if len(parts) > 0 {
			videoID = parts[len(parts)-1]
		}
	}

	if videoID == "" {
		return fmt.Errorf("could not extract video ID from entry")
	}

	video := database.YouTubeVideo{
		ID:           videoID,
		Title:        entry.Title,
		Description:  "", // Will be fetched from YouTube API if needed
		PublishedAt:  entry.Published,
		UpdatedAt:    entry.Updated,
		ThumbnailURL: fmt.Sprintf("https://img.youtube.com/vi/%s/maxresdefault.jpg", videoID),
		VideoURL:     fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
		EmbedURL:     fmt.Sprintf("https://www.youtube.com/embed/%s", videoID),
	}

	// Store or update video in database
	return y.storeVideo(video)
}

// storeVideo stores a YouTube video in the database
func (y *YouTubeService) storeVideo(video database.YouTubeVideo) error {
	// Check if video already exists
	existingVideo, err := y.db.GetYouTubeVideoByID(video.ID)
	if err == nil && existingVideo != nil {
		// Update existing video
		return y.db.UpdateYouTubeVideo(video)
	}

	// Create new video record
	return y.db.CreateYouTubeVideo(video)
}

// GetLatestVideos returns the latest YouTube videos sorted by newest first
func (y *YouTubeService) GetLatestVideos(limit int) (*YouTubeVideosResponse, error) {
	videos, err := y.db.GetYouTubeVideos(limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get YouTube videos: %w", err)
	}

	// Sort videos by published date (newest first)
	sort.Slice(videos, func(i, j int) bool {
		return videos[i].PublishedAt.After(videos[j].PublishedAt)
	})

	return &YouTubeVideosResponse{
		Videos:      videos,
		LastUpdated: time.Now(),
		TotalCount:  len(videos),
	}, nil
}

// updateVideoCache updates the cached JSON file for frontend consumption
func (y *YouTubeService) updateVideoCache() error {
	videos, err := y.GetLatestVideos(50) // Cache last 50 videos
	if err != nil {
		return err
	}

	// Create cache directory if it doesn't exist
	cacheDir := "./cache"
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Write to cache file
	cacheFile := "./cache/youtube_videos.json"
	data, err := json.MarshalIndent(videos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal videos: %w", err)
	}

	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	log.Printf("Updated YouTube video cache with %d videos", len(videos.Videos))
	return nil
}

// LoadCachedVideos loads videos from the cache file
func (y *YouTubeService) LoadCachedVideos() (*YouTubeVideosResponse, error) {
	cacheFile := "./cache/youtube_videos.json"

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		// If cache doesn't exist, return empty response
		return &YouTubeVideosResponse{
			Videos:      []database.YouTubeVideo{},
			LastUpdated: time.Now(),
			TotalCount:  0,
		}, nil
	}

	var response YouTubeVideosResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached videos: %w", err)
	}

	return &response, nil
}

// InitialFetch performs an initial fetch of videos (for setup)
func (y *YouTubeService) InitialFetch() error {
	// This would typically use YouTube Data API v3 to fetch initial videos
	// For now, we'll create a placeholder that gets populated by webhooks
	log.Printf("YouTube service initialized for channel: %s", y.channelID)
	log.Printf("Webhook URL: %s", y.callbackURL)

	// Create initial empty cache
	return y.updateVideoCache()
}

// ValidateSignature validates webhook signature (if YouTube sends one)
func (y *YouTubeService) ValidateSignature(payload []byte, signature string) bool {
	// YouTube PubSubHubbub doesn't typically use HMAC signatures
	// but we can implement this for additional security if needed
	if signature == "" {
		return true // No signature to validate
	}

	secret := os.Getenv("YOUTUBE_WEBHOOK_SECRET")
	if secret == "" {
		return true // No secret configured
	}

	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := "sha1=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
