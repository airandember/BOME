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
	url := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos/%s", b.streamLibrary, videoID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("AccessKey", b.streamAPIKey)

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var video BunnyVideo
	if err := json.NewDecoder(resp.Body).Decode(&video); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
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
