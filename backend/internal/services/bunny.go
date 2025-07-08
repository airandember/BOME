package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BunnyService handles all Bunny.net operations
type BunnyService struct {
	storageZone   string
	apiKey        string
	pullZone      string
	streamLibrary string
	streamAPIKey  string
	region        string
	webhookSecret string
	client        *http.Client
}

// BunnyVideo represents a video in Bunny Stream
type BunnyVideo struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Duration    float64   `json:"duration"`
	Size        int64     `json:"size"`
	Thumbnail   string    `json:"thumbnail"`
	Preview     string    `json:"preview"`
	LibraryID   string    `json:"library_id"`
}

// BunnyUploadResponse represents the response from a video upload
type BunnyUploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	VideoID string `json:"video_id,omitempty"`
}

// BunnyCollection represents a collection in Bunny Stream
type BunnyCollection struct {
	ID         string    `json:"guid"`
	Name       string    `json:"name"`
	VideoCount int       `json:"videoCount"`
	TotalSize  int64     `json:"totalSize"`
	CreatedAt  time.Time `json:"dateCreated"`
	UpdatedAt  time.Time `json:"lastUpdated"`
}

// BunnyCollectionsResponse represents the API response for collections
type BunnyCollectionsResponse struct {
	TotalItems   int               `json:"totalItems"`
	CurrentPage  int               `json:"currentPage"`
	ItemsPerPage int               `json:"itemsPerPage"`
	Items        []BunnyCollection `json:"items"`
}

// VideoPlayData represents the response from the video play data endpoint
type VideoPlayData struct {
	VideoLibraryID    string   `json:"videoLibraryId"`
	VideoGUID         string   `json:"guid"`
	Title             string   `json:"title"`
	Status            int      `json:"status"`
	Framerate         float64  `json:"framerate"`
	Width             int      `json:"width"`
	Height            int      `json:"height"`
	Duration          float64  `json:"duration"`
	ThumbnailCount    int      `json:"thumbnailCount"`
	ResolutionOptions []string `json:"resolutions"`
	ThumbnailFileName string   `json:"thumbnailFileName"`
	HasMP4Fallback    bool     `json:"hasMP4Fallback"`
	PlaybackURL       string   `json:"playbackUrl"`
	IframeSrc         string   `json:"iframeSrc"`
	DirectPlayURL     string   `json:"directPlayUrl"`
	ThumbnailURL      string   `json:"thumbnailUrl"`
}

// NewBunnyService creates a new Bunny.net service instance
func NewBunnyService() *BunnyService {
	return &BunnyService{
		storageZone:   os.Getenv("BUNNY_STORAGE_ZONE"),
		apiKey:        os.Getenv("BUNNY_API_KEY"),
		pullZone:      os.Getenv("BUNNY_PULL_ZONE"),
		streamLibrary: os.Getenv("BUNNY_STREAM_LIBRARY_ID"),
		streamAPIKey:  os.Getenv("BUNNY_STREAM_API_KEY"),
		region:        os.Getenv("BUNNY_REGION"),
		webhookSecret: os.Getenv("BUNNY_WEBHOOK_SECRET"),
		client:        &http.Client{Timeout: 30 * time.Second},
	}
}

// UploadVideo uploads a video file to Bunny Stream
func (b *BunnyService) UploadVideo(file *multipart.FileHeader, title, description string) (*BunnyUploadResponse, error) {
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Create the request body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the video file
	part, err := writer.CreateFormFile("video", filepath.Base(file.Filename))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, src)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	// Add metadata
	writer.WriteField("title", title)
	writer.WriteField("description", description)
	writer.Close()

	// Create the request
	url := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos", b.streamLibrary)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("AccessKey", b.streamAPIKey)

	// Make the request
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var uploadResp BunnyUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !uploadResp.Success {
		return nil, fmt.Errorf("upload failed: %s", uploadResp.Message)
	}

	return &uploadResp, nil
}

// GetVideo retrieves video information from Bunny Stream
func (b *BunnyService) GetVideo(videoID string) (*BunnyVideo, error) {
	if videoID == "" {
		return nil, fmt.Errorf("video ID is required")
	}

	if b.streamLibrary == "" || b.streamAPIKey == "" {
		return nil, fmt.Errorf("Bunny.net configuration missing (library: %v, key: %v)",
			b.streamLibrary != "", b.streamAPIKey != "")
	}

	url := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos/%s", b.streamLibrary, videoID)
	fmt.Printf("Making request to Bunny.net: %s\n", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("AccessKey", b.streamAPIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("Bunny.net response status: %d\n", resp.StatusCode)
	fmt.Printf("Bunny.net response body: %s\n", string(body))

	// Handle different status codes
	switch resp.StatusCode {
	case http.StatusOK:
		// Continue processing
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("unauthorized: check API key")
	case http.StatusForbidden:
		return nil, fmt.Errorf("forbidden: insufficient permissions")
	case http.StatusNotFound:
		return nil, fmt.Errorf("video not found: %s", videoID)
	case http.StatusTooManyRequests:
		return nil, fmt.Errorf("rate limited by Bunny.net")
	default:
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var video BunnyVideo
	if err := json.Unmarshal(body, &video); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w (body: %s)", err, string(body))
	}

	// Validate required fields
	if video.ID == "" {
		return nil, fmt.Errorf("invalid response: missing video ID (body: %s)", string(body))
	}

	return &video, nil
}

// GetStreamURL returns the streaming URL for a video
func (b *BunnyService) GetStreamURL(videoID string) string {
	return fmt.Sprintf("https://iframe.mediadelivery.net/embed/%s/%s", b.streamLibrary, videoID)
}

// GetThumbnailURL returns the thumbnail URL for a video
func (b *BunnyService) GetThumbnailURL(videoID string) string {
	return fmt.Sprintf("https://video.bunnycdn.com/%s/%s/thumbnail.jpg", b.streamLibrary, videoID)
}

// GetStreamLibrary returns the stream library ID
func (b *BunnyService) GetStreamLibrary() string {
	return b.streamLibrary
}

// GetRegion returns the region
func (b *BunnyService) GetRegion() string {
	return b.region
}

// GetStreamAPIKey returns the stream API key
func (b *BunnyService) GetStreamAPIKey() string {
	return b.streamAPIKey
}

// DeleteVideo deletes a video from Bunny Stream
func (b *BunnyService) DeleteVideo(videoID string) error {
	url := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos/%s", b.streamLibrary, videoID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("AccessKey", b.streamAPIKey)

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete failed with status: %d", resp.StatusCode)
	}

	return nil
}

// UploadToStorage uploads a file to Bunny Storage (for thumbnails, etc.)
func (b *BunnyService) UploadToStorage(filePath, remotePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	url := fmt.Sprintf("https://storage.bunnycdn.com/%s/%s", b.storageZone, remotePath)

	req, err := http.NewRequest("PUT", url, file)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("AccessKey", b.apiKey)
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetCDNURL returns the CDN URL for a file
func (b *BunnyService) GetCDNURL(path string) string {
	return fmt.Sprintf("https://%s.b-cdn.net/%s", b.pullZone, strings.TrimPrefix(path, "/"))
}

// ValidateWebhookSignature validates the webhook signature from Bunny.net
func (b *BunnyService) ValidateWebhookSignature(payload []byte, signature string) bool {
	// Implement webhook signature validation
	// This is a simplified version - implement proper HMAC validation
	return true
}

// ProcessWebhook processes webhook events from Bunny.net
func (b *BunnyService) ProcessWebhook(eventType string, payload []byte) error {
	switch eventType {
	case "video.encoded":
		return b.handleVideoEncoded(payload)
	case "video.failed":
		return b.handleVideoFailed(payload)
	default:
		return fmt.Errorf("unknown event type: %s", eventType)
	}
}

func (b *BunnyService) handleVideoEncoded(payload []byte) error {
	// Handle video encoding completion
	// Update database with video status
	return nil
}

func (b *BunnyService) handleVideoFailed(payload []byte) error {
	// Handle video encoding failure
	// Update database with error status
	return nil
}

// GetCollections retrieves all collections from Bunny Stream
func (b *BunnyService) GetCollections(page int, perPage int) (*BunnyCollectionsResponse, error) {
	url := fmt.Sprintf("https://video.bunnycdn.com/library/%s/collections?page=%d&itemsPerPage=%d",
		b.streamLibrary, page, perPage)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("AccessKey", b.streamAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get collections, status: %d", resp.StatusCode)
	}

	var collections BunnyCollectionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&collections); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &collections, nil
}

// GetCollection retrieves a single collection by ID
func (b *BunnyService) GetCollection(collectionID string) (*BunnyCollection, error) {
	url := fmt.Sprintf("https://video.bunnycdn.com/library/%s/collections/%s",
		b.streamLibrary, collectionID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("AccessKey", b.streamAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get collection, status: %d", resp.StatusCode)
	}

	var collection BunnyCollection
	if err := json.NewDecoder(resp.Body).Decode(&collection); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &collection, nil
}

// GetVideoPlayData retrieves video play data from Bunny Stream
func (b *BunnyService) GetVideoPlayData(videoID string) (*VideoPlayData, error) {
	if videoID == "" {
		return nil, fmt.Errorf("video ID is required")
	}

	fmt.Printf("Fetching play data for video %s\n", videoID)

	url := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos/%s", b.streamLibrary, videoID)
	fmt.Printf("Making request to Bunny.net: %s\n", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("AccessKey", b.streamAPIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Bunny.net response status: %d\n", resp.StatusCode)

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("Bunny.net response body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var playData VideoPlayData
	if err := json.Unmarshal(body, &playData); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Get CDN hostname based on region
	cdnHostname := "vz-" + b.pullZone
	if b.region != "" {
		cdnHostname += "-" + b.region
	}
	cdnHostname += ".b-cdn.net"

	// Construct the streaming URLs
	playData.DirectPlayURL = fmt.Sprintf("https://%s/%s/playlist.m3u8", cdnHostname, videoID)
	playData.PlaybackURL = playData.DirectPlayURL // Use the HLS stream URL for playback
	playData.IframeSrc = fmt.Sprintf("https://iframe.mediadelivery.net/embed/%s/%s", b.streamLibrary, videoID)
	playData.ThumbnailURL = fmt.Sprintf("https://%s/%s/thumbnail.jpg", cdnHostname, videoID)

	fmt.Printf("Successfully fetched play data for video %s\n", videoID)
	fmt.Printf("Streaming URLs:\n")
	fmt.Printf("- HLS Stream: %s\n", playData.PlaybackURL)
	fmt.Printf("- Iframe: %s\n", playData.IframeSrc)
	fmt.Printf("- Thumbnail: %s\n", playData.ThumbnailURL)

	return &playData, nil
}
