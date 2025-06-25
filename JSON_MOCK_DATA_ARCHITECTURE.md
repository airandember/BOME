# JSON-Based Mock Data Architecture

## Overview

We've successfully implemented a flexible JSON-based mock data architecture that allows seamless switching between development (mock data) and production (real APIs/database) modes.

## Architecture Benefits

### ✅ **Advantages Over Previous Hardcoded Approach:**

1. **Flexible Data Management**: Modify JSON files without recompiling Go code
2. **Clean Separation**: Clear abstraction between data sources and business logic
3. **Version Control Friendly**: Track data changes easily in Git
4. **Team Collaboration**: Non-developers can modify test data
5. **Production Ready**: Simple environment variable switch
6. **Realistic Testing**: JSON structure matches real API responses
7. **Extensible**: Easy to add new data types and endpoints

## File Structure

```
backend/
├── mock_data/                    # JSON mock data files
│   ├── youtube_videos.json      # YouTube video data
│   ├── users.json               # User accounts and profiles
│   └── ...                     # Additional data types
├── internal/
│   ├── data/                    # Data access layer
│   │   ├── interfaces.go        # DataSource interface definitions
│   │   ├── mock_data_loader.go  # JSON file loader
│   │   └── ...                 # Real data source implementations
│   └── services/
│       └── youtube.go          # Updated service using DataSource
└── ...
```

## Implementation Details

### 1. **Data Source Interface**

```go
type DataSource interface {
    GetYouTubeVideos(limit int) ([]database.YouTubeVideo, error)
    GetYouTubeVideoByID(id string) (*database.YouTubeVideo, error)
    GetYouTubeChannelInfo() (*ChannelInfo, error)
    SearchYouTubeVideos(query string, limit int) ([]database.YouTubeVideo, error)
    GetYouTubeVideosByCategory(category string, limit int) ([]database.YouTubeVideo, error)
}
```

### 2. **Mock Data Source**

- **File**: `backend/internal/data/mock_data_loader.go`
- **Purpose**: Loads JSON files and converts to Go structs
- **Features**: 
  - JSON parsing with error handling
  - Type conversion between JSON and database models
  - Search functionality
  - Category filtering

### 3. **Real Data Source**

- **File**: `backend/internal/data/interfaces.go`
- **Purpose**: Production implementation using real APIs/database
- **Status**: Placeholder for future implementation

## Configuration

### Environment Variables

```bash
# Development mode (uses JSON files)
YOUTUBE_USE_MOCK_DATA=true

# Production mode (uses real APIs)
YOUTUBE_USE_MOCK_DATA=false
YOUTUBE_API_KEY=your_real_api_key
```

### Service Initialization

```go
// Automatic mode selection based on environment
youtubeService := services.NewYouTubeService(db)

// The service automatically chooses:
// - MockDataSource if YOUTUBE_USE_MOCK_DATA=true
// - RealDataSource if YOUTUBE_USE_MOCK_DATA=false
```

## Mock Data Files

### 1. **YouTube Videos** (`mock_data/youtube_videos.json`)

```json
{
  "videos": [
    {
      "id": "dQw4w9WgXcQ",
      "title": "Book of Mormon Archaeological Evidence - Ancient Civilizations",
      "description": "Exploring archaeological evidence...",
      "published_at": "2024-12-15T10:00:00Z",
      "updated_at": "2024-12-15T10:00:00Z",
      "created_at": "2024-12-15T10:00:00Z",
      "thumbnail_url": "https://img.youtube.com/vi/dQw4w9WgXcQ/maxresdefault.jpg",
      "video_url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
      "embed_url": "https://www.youtube.com/embed/dQw4w9WgXcQ",
      "duration": "15:42",
      "view_count": 12500,
      "channel_id": "UCHp1EBgpKytZt_-j72EZ83Q",
      "tags": ["archaeology", "book-of-mormon", "ancient-civilizations"],
      "category": "Education",
      "language": "en",
      "status": "published"
    }
  ],
  "metadata": {
    "total_count": 10,
    "last_updated": "2024-12-18T09:00:00Z",
    "channel_info": {
      "id": "UCHp1EBgpKytZt_-j72EZ83Q",
      "name": "Book of Mormon Evidence",
      "subscriber_count": 10800,
      "video_count": 285,
      "view_count": 879012
    }
  }
}
```

### 2. **Users** (`mock_data/users.json`)

```json
{
  "users": [
    {
      "id": 1,
      "email": "admin@bookofmormonevidence.org",
      "full_name": "System Administrator",
      "role": "super_admin",
      "subscription_status": "active",
      "subscription_plan": "lifetime",
      "created_at": "2023-01-15T00:00:00Z",
      "profile": {
        "bio": "Administrator of the Book of Mormon Evidence platform",
        "location": "Utah, USA"
      }
    }
  ],
  "metadata": {
    "total_count": 4,
    "active_subscriptions": 3,
    "premium_users": 2
  }
}
```

## API Endpoints

### YouTube API

| Method | Endpoint | Description | Response |
|--------|----------|-------------|----------|
| GET | `/api/v1/youtube/videos` | Get all videos with pagination | `YouTubeVideosResponse` |
| GET | `/api/v1/youtube/videos/latest` | Get latest videos (cached) | `YouTubeVideosResponse` |
| GET | `/api/v1/youtube/videos/search?q=query` | Search videos | `YouTubeVideosResponse` |
| GET | `/api/v1/youtube/videos/category/:category` | Filter by category | `YouTubeVideosResponse` |
| GET | `/api/v1/youtube/videos/:id` | Get specific video | `YouTubeVideo` |
| GET | `/api/v1/youtube/status` | Integration status | `StatusResponse` |
| GET | `/api/v1/youtube/channel` | Channel information | `ChannelInfo` |

### Example API Calls

```bash
# Get latest videos
curl http://localhost:8080/api/v1/youtube/videos/latest

# Search videos
curl "http://localhost:8080/api/v1/youtube/videos/search?q=archaeology&limit=5"

# Get specific video
curl http://localhost:8080/api/v1/youtube/videos/dQw4w9WgXcQ

# Get channel info
curl http://localhost:8080/api/v1/youtube/channel
```

## Development Workflow

### 1. **Modifying Mock Data**

```bash
# Edit JSON files directly
nano backend/mock_data/youtube_videos.json

# No compilation needed - changes are immediate
# Restart server to reload data
```

### 2. **Adding New Video**

```json
{
  "id": "newVideoID",
  "title": "New Book of Mormon Evidence Video",
  "description": "Latest archaeological discoveries...",
  "published_at": "2024-12-20T10:00:00Z",
  "updated_at": "2024-12-20T10:00:00Z",
  "created_at": "2024-12-20T10:00:00Z",
  "thumbnail_url": "https://img.youtube.com/vi/newVideoID/maxresdefault.jpg",
  "video_url": "https://www.youtube.com/watch?v=newVideoID",
  "embed_url": "https://www.youtube.com/embed/newVideoID",
  "duration": "18:30",
  "view_count": 1500,
  "channel_id": "UCHp1EBgpKytZt_-j72EZ83Q",
  "tags": ["new", "evidence", "archaeology"],
  "category": "Education",
  "language": "en",
  "status": "published"
}
```

### 3. **Testing Different Scenarios**

```bash
# Test with different data sets
cp mock_data/youtube_videos.json mock_data/youtube_videos_backup.json
# Edit youtube_videos.json for testing
# Restore when done
```

## Production Deployment

### 1. **Environment Configuration**

```bash
# Production environment variables
YOUTUBE_USE_MOCK_DATA=false
YOUTUBE_API_KEY=your_production_api_key
YOUTUBE_CHANNEL_ID=UCHp1EBgpKytZt_-j72EZ83Q
```

### 2. **Real Data Source Implementation**

The `RealDataSource` in `backend/internal/data/interfaces.go` needs to be implemented with:

- YouTube Data API v3 integration
- Database operations
- Caching strategies
- Error handling

### 3. **Migration Path**

1. Develop and test with mock data
2. Implement real data source methods
3. Switch environment variable
4. Deploy to production

## Error Handling

### Mock Data Errors

```go
// File not found
if err := loader.LoadYouTubeVideos(); err != nil {
    log.Printf("Failed to load mock data: %v", err)
    // Fallback to empty data or default values
}

// JSON parsing errors
if err := json.Unmarshal(data, &videos); err != nil {
    return fmt.Errorf("invalid JSON format: %w", err)
}
```

### Production Fallbacks

```go
// API failures in production
if !useMockData && apiErr != nil {
    log.Printf("API failed, falling back to cache: %v", apiErr)
    return loadFromCache()
}
```

## Testing

### 1. **Unit Tests**

```go
func TestMockDataLoader(t *testing.T) {
    loader := data.NewMockDataLoader("./testdata")
    videos, err := loader.LoadYouTubeVideos()
    
    assert.NoError(t, err)
    assert.Greater(t, len(videos), 0)
    assert.Equal(t, "dQw4w9WgXcQ", videos[0].ID)
}
```

### 2. **Integration Tests**

```bash
# Test API endpoints
curl http://localhost:8080/api/v1/youtube/videos/latest
curl "http://localhost:8080/api/v1/youtube/videos/search?q=archaeology"
```

## Future Enhancements

### 1. **Additional Data Types**

- Articles (`mock_data/articles.json`)
- Subscriptions (`mock_data/subscriptions.json`)
- Analytics (`mock_data/analytics.json`)
- Comments (`mock_data/comments.json`)

### 2. **Advanced Features**

- Data relationships between JSON files
- Dynamic data generation
- Performance caching
- Data validation schemas

### 3. **Developer Tools**

- Mock data generator scripts
- JSON schema validation
- Data consistency checks
- API response mocking

## Best Practices

### 1. **JSON Structure**

- Use consistent naming conventions (snake_case)
- Include metadata for context
- Maintain realistic data relationships
- Add timestamps for temporal data

### 2. **Data Management**

- Keep files under version control
- Use meaningful commit messages for data changes
- Backup production-like data sets
- Document data relationships

### 3. **Performance**

- Keep JSON files reasonably sized
- Use pagination for large datasets
- Cache parsed data when possible
- Monitor file loading performance

## Conclusion

This JSON-based mock data architecture provides:

- **Flexibility**: Easy data modification without code changes
- **Scalability**: Ready for production deployment
- **Maintainability**: Clear separation of concerns
- **Collaboration**: Non-developers can contribute to test data
- **Testing**: Comprehensive testing scenarios

The architecture successfully bridges the gap between development and production, providing a robust foundation for the BOME streaming platform. 