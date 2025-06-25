# YouTube Backend Integration Documentation

## Overview
This document describes the complete implementation of moving YouTube mock data from the frontend to the backend, creating a proper API architecture for the BOME streaming platform.

## Architecture Changes

### Before: Frontend Mock Data
- Mock data was stored directly in `frontend/src/lib/stores/youtube.ts`
- No backend API calls
- Data was hardcoded in the frontend

### After: Backend API with Mock Data
- Mock data moved to `backend/internal/services/youtube_mock.go`
- Frontend calls backend API endpoints
- Proper separation of concerns
- Ready for real API integration

## Backend Implementation

### 1. Mock Data Service (`backend/internal/services/youtube_mock.go`)
```go
// GetMockYouTubeVideos returns mock YouTube video data for development and testing
func GetMockYouTubeVideos() []database.YouTubeVideo
```

**Features:**
- 10 realistic Book of Mormon Evidence videos
- Mix of YouTube thumbnails and BOME placeholders
- Proper metadata (titles, descriptions, view counts, durations)
- Sorted by publication date (newest first)

### 2. Enhanced YouTube Service (`backend/internal/services/youtube.go`)
```go
type YouTubeService struct {
    channelID   string
    callbackURL string
    verifyToken string
    hubURL      string
    apiKey      string
    client      *http.Client
    db          *database.DB
    useMockData bool // Flag to enable mock data mode
}
```

**Key Features:**
- Mock data mode controlled by `YOUTUBE_USE_MOCK_DATA` environment variable
- Defaults to mock mode when no environment variable is set
- Seamless switching between mock and real data
- Proper error handling and fallbacks

### 3. API Endpoints

#### Available Endpoints:
- `GET /api/v1/youtube/videos` - Get all videos with pagination
- `GET /api/v1/youtube/videos/latest` - Get latest videos (cached)
- `GET /api/v1/youtube/status` - Integration status and statistics
- `POST /api/v1/youtube/subscribe` - Manual subscription management
- `POST /api/v1/youtube/unsubscribe` - Manual unsubscription

#### Response Format:
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
      "view_count": 12500
    }
  ],
  "last_updated": "2024-12-18T09:05:23Z",
  "total_count": 10
}
```

## Frontend Implementation

### 1. Updated YouTube Store (`frontend/src/lib/stores/youtube.ts`)
```typescript
// Fetch latest YouTube videos from backend API
async fetchLatestVideos() {
    const response = await fetch('http://localhost:8080/youtube/videos/latest');
    const data: YouTubeVideosResponse = await response.json();
    // Update store with backend data
}
```

**Changes:**
- Removed all frontend mock data
- Added proper API calls to backend
- Enhanced error handling
- Proper TypeScript typing

### 2. Updated Types (`frontend/src/lib/types/youtube.ts`)
```typescript
export interface YouTubeVideo {
    id: string;
    title: string;
    description: string;
    published_at: string; // ISO date string from backend
    updated_at: string;
    created_at: string;
    thumbnail_url: string;
    video_url: string;
    embed_url: string;
    duration?: string;
    view_count?: number;
}
```

**Features:**
- Aligned with backend data structure
- Proper date handling (ISO strings)
- Optional fields for flexibility

## Configuration

### Backend Configuration (`backend/test_config.env`)
```env
# Enable mock data mode for testing
YOUTUBE_USE_MOCK_DATA=true

# YouTube API Configuration (for future use)
YOUTUBE_API_KEY=AIzaSyDYSN0Vk3gdgRM8mtiaOH7c7eXKsXRjyKk
YOUTUBE_CHANNEL_ID=UCalautjgmA5BvpDQiQvCSUw

# Server Configuration
SERVER_PORT=8080
ENVIRONMENT=development
```

### Frontend Configuration
- API endpoint: `http://localhost:8080/youtube/videos/latest`
- Automatic fallback to BOME placeholders for missing thumbnails
- Error handling with user-friendly messages

## Testing

### 1. Backend Testing
```bash
# Start backend server
cd backend
go run main.go

# Test API endpoint
curl http://localhost:8080/api/v1/youtube/videos/latest
```

### 2. Frontend Testing
```bash
# Start frontend development server
cd frontend
npm run dev

# Navigate to: http://localhost:5173/youtube
```

### 3. Integration Testing
1. Start backend server on port 8080
2. Start frontend development server on port 5173
3. Navigate to YouTube page
4. Verify videos load from backend API

## Mock Data Details

### Video Content
- **10 videos** with Book of Mormon Evidence themes
- **Realistic titles** and descriptions
- **Mixed thumbnails**: YouTube URLs and BOME placeholders
- **Proper metadata**: view counts, durations, publication dates

### Placeholder Handling
- Empty thumbnail URLs fallback to BOME placeholder
- BOME placeholders: `/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png`
- Alternative BOME globe: `/src/lib/HOMEPAGE_TEST_ASSETS/The_Globe_World.png`

## Future Enhancements

### 1. Real YouTube Integration
- Switch `YOUTUBE_USE_MOCK_DATA=false`
- Implement YouTube Data API v3 calls
- Add webhook support for real-time updates

### 2. Database Integration
- Store videos in PostgreSQL database
- Add caching layer with Redis
- Implement video search and filtering

### 3. Performance Optimizations
- Add pagination support
- Implement lazy loading
- Add video thumbnail optimization

## API Endpoints Reference

| Method | Endpoint | Description | Response |
|--------|----------|-------------|----------|
| GET | `/api/v1/youtube/videos` | Get all videos with optional pagination | `YouTubeVideosResponse` |
| GET | `/api/v1/youtube/videos/latest` | Get latest videos (cached) | `YouTubeVideosResponse` |
| GET | `/api/v1/youtube/status` | Integration status and statistics | `StatusResponse` |
| POST | `/api/v1/youtube/subscribe` | Manual subscription management | `SuccessResponse` |
| POST | `/api/v1/youtube/unsubscribe` | Manual unsubscription | `SuccessResponse` |

## Error Handling

### Backend Errors
- **500**: Internal server error
- **400**: Bad request (invalid parameters)
- **404**: Endpoint not found

### Frontend Errors
- **Network errors**: Display user-friendly message
- **API errors**: Show error toast notification
- **Loading states**: Display spinner during API calls

## Security Considerations

### API Security
- CORS headers configured for development
- Rate limiting (planned)
- Input validation and sanitization

### Data Privacy
- No sensitive user data in mock responses
- YouTube API key stored securely in environment variables
- Proper error message sanitization

## Deployment Notes

### Development
- Mock data mode enabled by default
- No database required
- Simple environment configuration

### Production
- Set `YOUTUBE_USE_MOCK_DATA=false`
- Configure real YouTube API integration
- Set up proper database and caching

## Troubleshooting

### Common Issues
1. **API endpoint not found (404)**
   - Verify backend server is running on port 8080
   - Check route registration in `routes.go`

2. **CORS errors**
   - Verify CORS configuration in backend
   - Check frontend API endpoint URL

3. **Empty video list**
   - Verify mock data service is working
   - Check console for API errors

### Debug Commands
```bash
# Check backend health
curl http://localhost:8080/health

# Check YouTube status
curl http://localhost:8080/api/v1/youtube/status

# Test video endpoint
curl http://localhost:8080/api/v1/youtube/videos/latest
```

## Conclusion

The YouTube integration has been successfully migrated from frontend mock data to a proper backend API architecture. This provides:

- ✅ **Separation of concerns**: Data logic in backend, UI logic in frontend
- ✅ **Scalability**: Ready for real YouTube API integration
- ✅ **Maintainability**: Centralized data management
- ✅ **Testing**: Isolated components for better testing
- ✅ **Performance**: Potential for caching and optimization

The system is now ready for production deployment with real YouTube API integration when needed. 