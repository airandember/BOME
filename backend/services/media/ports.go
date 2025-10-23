package media

// This file defines the ports (interfaces) for the media domain
// Implementations are in the bunny/ subdirectory

import "time"

// Video represents a video object
type Video struct {
	VideoLibraryID int       `json:"videoLibraryId"`
	GUID           string    `json:"guid"`
	Title          string    `json:"title"`
	CollectionID   string    `json:"collectionId"`
	ThumbnailURL   string    `json:"thumbnailFileName"`
	IsPublic       bool      `json:"isPublic"`
	DateUploaded   time.Time `json:"dateUploaded"`
	Views          int       `json:"views"`
	Length         float64   `json:"length"`
	Status         int       `json:"status"`
}

// Collection represents a video collection
type Collection struct {
	GUID           string    `json:"guid"`
	Name           string    `json:"name"`
	VideoLibraryID int       `json:"videoLibraryId"`
	DateCreated    time.Time `json:"dateCreated"`
}

// VideoPlayData represents video playback data
type VideoPlayData struct {
	IframeSrc         string                   `json:"iframeSrc"`
	DirectPlayURL     string                   `json:"directPlayUrl"`
	ThumbnailURL      string                   `json:"thumbnailUrl"`
	ResolutionOptions []map[string]interface{} `json:"resolutionOptions"`
}

// BunnyPort defines the interface for Bunny.net video streaming operations
// Implementation: services/media/bunny/bunny.go
type BunnyPort interface {
	// Video Management
	FetchVideos(libraryID int, page, itemsPerPage int, search string) ([]Video, error)
	FetchVideo(libraryID int, videoID string) (*Video, error)
	CreateVideo(libraryID int, title string) (*Video, error)
	UpdateVideo(libraryID int, videoID string, title string) error
	DeleteVideo(libraryID int, videoID string) error
	SetThumbnail(libraryID int, videoID string, thumbnailURL string) error
	UploadVideo(libraryID int, videoID string, filePath string) error
	GetVideoPlayData(videoID string) (*VideoPlayData, error)

	// Collection Management
	FetchCollections(libraryID int, page, itemsPerPage int, search string) ([]Collection, error)
	FetchCollection(libraryID int, collectionID string) (*Collection, error)
	CreateCollection(libraryID int, name string) (*Collection, error)
	UpdateCollection(libraryID int, collectionID string, name string) error
	DeleteCollection(libraryID int, collectionID string) error
}
