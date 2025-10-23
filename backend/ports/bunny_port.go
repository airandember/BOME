package ports

// BunnyPort defines the interface for Bunny.net CDN and video streaming operations
type BunnyPort interface {
	// Video Management
	GetVideoPlayData(videoID string) (*VideoPlayData, error)
	UploadVideo(filename string, file interface{}) (*VideoUploadResult, error)
	DeleteVideo(videoID string) error

	// Video Metadata
	GetVideoInfo(videoID string) (*VideoInfo, error)
	UpdateVideoMetadata(videoID string, metadata map[string]interface{}) error

	// Thumbnail Operations
	GetVideoThumbnail(videoID string) (string, error)
	GenerateThumbnail(videoID string, timeSeconds int) (string, error)

	// Streaming URLs
	GetStreamURL(videoID string) (string, error)
	GetEmbedURL(videoID string) (string, error)

	// Service Status
	IsConfigured() bool
}

// VideoPlayData represents video playback information
type VideoPlayData struct {
	VideoID           string
	IframeSrc         string
	DirectPlayURL     string
	ThumbnailURL      string
	ResolutionOptions []VideoResolution
	Duration          int
	Status            string
}

// VideoResolution represents a video resolution option
type VideoResolution struct {
	Width  int
	Height int
	Label  string // e.g., "1080p", "720p"
	URL    string
}

// VideoUploadResult represents the result of a video upload
type VideoUploadResult struct {
	VideoID   string
	Status    string
	UploadURL string
	Message   string
}

// VideoInfo represents detailed video information
type VideoInfo struct {
	VideoID      string
	Title        string
	Description  string
	Duration     int
	FileSize     int64
	Status       string
	ThumbnailURL string
	CreatedAt    string
	UpdatedAt    string
}
