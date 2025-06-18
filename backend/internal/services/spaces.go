package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// SpacesService handles all Digital Ocean Spaces operations
type SpacesService struct {
	client   *s3.Client
	bucket   string
	endpoint string
	region   string
	cdnURL   string
}

// FileInfo represents file information
type FileInfo struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	ETag         string    `json:"etag"`
	ContentType  string    `json:"content_type"`
	URL          string    `json:"url"`
}

// UploadResult represents the result of a file upload
type UploadResult struct {
	Key         string `json:"key"`
	URL         string `json:"url"`
	CDNURL      string `json:"cdn_url"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}

// NewSpacesService creates a new Digital Ocean Spaces service instance
func NewSpacesService() (*SpacesService, error) {
	accessKey := os.Getenv("DO_SPACES_KEY")
	secretKey := os.Getenv("DO_SPACES_SECRET")
	endpoint := os.Getenv("DO_SPACES_ENDPOINT")
	bucket := os.Getenv("DO_SPACES_BUCKET")
	region := os.Getenv("DO_SPACES_REGION")
	cdnURL := os.Getenv("DO_SPACES_CDN_ENDPOINT")

	// Create custom endpoint resolver
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s", endpoint),
		}, nil
	})

	// Load configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg)

	return &SpacesService{
		client:   client,
		bucket:   bucket,
		endpoint: endpoint,
		region:   region,
		cdnURL:   cdnURL,
	}, nil
}

// UploadFile uploads a file to Digital Ocean Spaces
func (s *SpacesService) UploadFile(file *multipart.FileHeader, key string) (*UploadResult, error) {
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Read file content
	content, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Determine content type
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload to Spaces
	_, err = s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(content),
		ContentType: aws.String(contentType),
		ACL:         "public-read", // Make file publicly accessible
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Generate URLs
	url := fmt.Sprintf("https://%s.%s/%s", s.bucket, s.endpoint, key)
	cdnURL := fmt.Sprintf("%s/%s", s.cdnURL, key)

	return &UploadResult{
		Key:         key,
		URL:         url,
		CDNURL:      cdnURL,
		Size:        int64(len(content)),
		ContentType: contentType,
	}, nil
}

// UploadFileFromPath uploads a file from a local path
func (s *SpacesService) UploadFileFromPath(localPath, key string) (*UploadResult, error) {
	// Open local file
	file, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Determine content type based on file extension
	contentType := s.getContentType(filepath.Ext(localPath))

	// Upload to Spaces
	_, err = s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(content),
		ContentType: aws.String(contentType),
		ACL:         "public-read",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Generate URLs
	url := fmt.Sprintf("https://%s.%s/%s", s.bucket, s.endpoint, key)
	cdnURL := fmt.Sprintf("%s/%s", s.cdnURL, key)

	return &UploadResult{
		Key:         key,
		URL:         url,
		CDNURL:      cdnURL,
		Size:        fileInfo.Size(),
		ContentType: contentType,
	}, nil
}

// GetFile retrieves file information
func (s *SpacesService) GetFile(key string) (*FileInfo, error) {
	result, err := s.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	cdnURL := fmt.Sprintf("%s/%s", s.cdnURL, key)

	return &FileInfo{
		Key:          key,
		Size:         *result.ContentLength,
		LastModified: *result.LastModified,
		ETag:         strings.Trim(*result.ETag, "\""),
		ContentType:  *result.ContentType,
		URL:          cdnURL, // Use CDN URL for better performance
	}, nil
}

// ListFiles lists files in a directory
func (s *SpacesService) ListFiles(prefix string) ([]*FileInfo, error) {
	result, err := s.client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	var files []*FileInfo
	for _, obj := range result.Contents {
		cdnURL := fmt.Sprintf("%s/%s", s.cdnURL, *obj.Key)

		files = append(files, &FileInfo{
			Key:          *obj.Key,
			Size:         *obj.Size,
			LastModified: *obj.LastModified,
			ETag:         strings.Trim(*obj.ETag, "\""),
			URL:          cdnURL,
		})
	}

	return files, nil
}

// DeleteFile deletes a file from Spaces
func (s *SpacesService) DeleteFile(key string) error {
	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// CreateBackup creates a backup of a file or directory
func (s *SpacesService) CreateBackup(localPath, backupKey string) (*UploadResult, error) {
	// Add timestamp to backup key
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	backupKey = fmt.Sprintf("%s-%s", backupKey, timestamp)

	return s.UploadFileFromPath(localPath, backupKey)
}

// DownloadFile downloads a file from Spaces to local path
func (s *SpacesService) DownloadFile(key, localPath string) error {
	result, err := s.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer result.Body.Close()

	// Create local file
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Copy content to local file
	_, err = io.Copy(file, result.Body)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// GetSignedURL generates a signed URL for temporary access
func (s *SpacesService) GetSignedURL(key string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiration))
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return request.URL, nil
}

// UploadBackup uploads a database backup
func (s *SpacesService) UploadBackup(backupPath, backupName string) (*UploadResult, error) {
	// Create backup key with timestamp
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	backupKey := fmt.Sprintf("backups/%s-%s.sql", backupName, timestamp)

	return s.UploadFileFromPath(backupPath, backupKey)
}

// ListBackups lists all backup files
func (s *SpacesService) ListBackups() ([]*FileInfo, error) {
	return s.ListFiles("backups/")
}

// CleanupOldBackups removes backup files older than specified days
func (s *SpacesService) CleanupOldBackups(daysToKeep int) error {
	backups, err := s.ListBackups()
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	cutoffTime := time.Now().AddDate(0, 0, -daysToKeep)

	for _, backup := range backups {
		if backup.LastModified.Before(cutoffTime) {
			if err := s.DeleteFile(backup.Key); err != nil {
				return fmt.Errorf("failed to delete old backup %s: %w", backup.Key, err)
			}
		}
	}

	return nil
}

// Helper method to determine content type
func (s *SpacesService) getContentType(extension string) string {
	switch strings.ToLower(extension) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".mp4":
		return "video/mp4"
	case ".avi":
		return "video/x-msvideo"
	case ".mov":
		return "video/quicktime"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".sql":
		return "application/sql"
	default:
		return "application/octet-stream"
	}
}
