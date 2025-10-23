# 🧬 Video Streaming Braid - Backend
**Bunny.net-powered video delivery and CDN management**

---

## 🔗 **Cross-Repository Braid**

> **⚠️ IMPORTANT**: This is the **backend portion** of the Video Streaming Braid.  
> **Frontend portion**: See `_frontend/braids/video-streaming/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## 📋 **Backend Overview**

**Purpose**: Server-side video management, CDN integration, and streaming delivery  
**Technology**: Go, PostgreSQL, Bunny.net CDN  
**Complexity**: **Very High** (CDN Integration, Video Processing, Adaptive Streaming)  
**Priority**: **CRITICAL** - Core business functionality!

**Critical Files**:
- `backend/internal/services/bunny.go` (Bunny.net integration)
- `backend/internal/services/bunny_optimized.go` (Performance optimized)
- `backend/internal/services/master_video_sync.go` (Video synchronization)
- `backend/internal/database/video.go`
- `backend/internal/database/master_video.go`
- `backend/internal/routes/video.go`

---

## 🎯 **Key Features**

### **1. Video Management**:
- Upload videos to Bunny.net
- Video metadata management
- Thumbnail generation
- Category and tag organization
- Search and discovery
- Video status workflow (draft, processing, published, archived)

### **2. Bunny.net CDN Integration**:
- Stream Library integration
- Video upload to CDN
- Adaptive bitrate streaming
- Global CDN delivery
- Video analytics
- DRM protection (optional)

### **3. Video Playback**:
- HLS/DASH streaming protocols
- Adaptive bitrate selection
- Quality selection (SD, HD, 4K)
- Playback speed control
- Subtitles/Captions support
- Thumbnail preview scrubbing

### **4. Access Control**:
- Subscription-based access
- Free vs. Premium content
- Geo-restrictions
- Time-limited access
- Concurrent stream limiting

### **5. Video Analytics**:
- View counts
- Watch time
- Completion rates
- Geographic distribution
- Device/browser stats
- Engagement metrics

---

## 🗄️ **Database Schema**

### **Videos Table**:
**File**: `backend/internal/database/video.go`

```sql
CREATE TABLE videos (
    id SERIAL PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    slug VARCHAR(500) UNIQUE,
    
    -- Bunny.net Integration
    bunny_video_id VARCHAR(255) UNIQUE,
    bunny_library_id VARCHAR(255),
    bunny_stream_url VARCHAR(1000),
    bunny_thumbnail_url VARCHAR(1000),
    
    -- Video Metadata
    duration_seconds INTEGER,
    file_size_bytes BIGINT,
    resolution VARCHAR(20), -- '1920x1080', '3840x2160', etc.
    fps INTEGER DEFAULT 30,
    bitrate_kbps INTEGER,
    codec VARCHAR(50),
    format VARCHAR(50), -- 'mp4', 'mov', 'avi'
    
    -- Content Organization
    category_id INTEGER REFERENCES categories(id),
    is_featured BOOLEAN DEFAULT false,
    is_trending BOOLEAN DEFAULT false,
    sort_order INTEGER DEFAULT 0,
    
    -- Access Control
    access_level VARCHAR(50) DEFAULT 'free', -- 'free', 'basic', 'premium', 'enterprise'
    requires_subscription BOOLEAN DEFAULT false,
    
    -- Status & Workflow
    status VARCHAR(50) DEFAULT 'draft', -- 'draft', 'processing', 'published', 'archived', 'deleted'
    published_at TIMESTAMP,
    
    -- Analytics
    view_count INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_videos_bunny_video_id ON videos(bunny_video_id);
CREATE INDEX idx_videos_slug ON videos(slug);
CREATE INDEX idx_videos_status ON videos(status);
CREATE INDEX idx_videos_published_at ON videos(published_at);
CREATE INDEX idx_videos_access_level ON videos(access_level);
CREATE INDEX idx_videos_category_id ON videos(category_id);
CREATE INDEX idx_videos_featured ON videos(is_featured) WHERE is_featured = true;

COMMENT ON TABLE videos IS 'Video content with Bunny.net CDN integration';
COMMENT ON COLUMN videos.bunny_video_id IS 'Bunny.net video GUID';
COMMENT ON COLUMN videos.access_level IS 'Minimum subscription tier required';
```

---

### **Master Videos Table** (YouTube Integration):
**File**: `backend/internal/database/master_video.go`

```sql
CREATE TABLE master_videos (
    id SERIAL PRIMARY KEY,
    youtube_video_id VARCHAR(255) UNIQUE,
    youtube_channel_id VARCHAR(255),
    title VARCHAR(500),
    description TEXT,
    published_at TIMESTAMP,
    
    -- Synced Video Reference
    video_id INTEGER REFERENCES videos(id) ON DELETE SET NULL,
    
    -- Sync Status
    sync_status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'syncing', 'synced', 'failed'
    last_sync_at TIMESTAMP,
    sync_error TEXT,
    
    -- YouTube Metadata
    duration VARCHAR(20),
    thumbnail_url VARCHAR(1000),
    view_count BIGINT,
    like_count INTEGER,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_master_videos_youtube_id ON master_videos(youtube_video_id);
CREATE INDEX idx_master_videos_channel_id ON master_videos(youtube_channel_id);
CREATE INDEX idx_master_videos_sync_status ON master_videos(sync_status);

COMMENT ON TABLE master_videos IS 'YouTube videos awaiting sync to Bunny.net';
```

**Purpose**: Track YouTube videos that need to be imported into Bunny.net.

---

### **Video Access Log Table**:
**File**: `backend/internal/database/video_access.go`

```sql
CREATE TABLE video_access_log (
    id SERIAL PRIMARY KEY,
    video_id INTEGER REFERENCES videos(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    
    -- Access Details
    accessed_at TIMESTAMP DEFAULT NOW(),
    ip_address VARCHAR(45),
    user_agent TEXT,
    device_type VARCHAR(50), -- 'desktop', 'mobile', 'tablet', 'tv'
    browser VARCHAR(100),
    operating_system VARCHAR(100),
    
    -- Geographic Data
    country_code VARCHAR(2),
    city VARCHAR(255),
    region VARCHAR(255),
    
    -- Viewing Metrics
    watch_duration_seconds INTEGER,
    completion_percentage DECIMAL(5, 2),
    quality_level VARCHAR(20), -- 'sd', 'hd', '4k'
    
    -- Session Tracking
    session_id VARCHAR(255)
);

CREATE INDEX idx_video_access_video_id ON video_access_log(video_id);
CREATE INDEX idx_video_access_user_id ON video_access_log(user_id);
CREATE INDEX idx_video_access_accessed_at ON video_access_log(accessed_at);
CREATE INDEX idx_video_access_country ON video_access_log(country_code);

-- Partitioning by month for performance
CREATE TABLE video_access_log_2025_10 PARTITION OF video_access_log
FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');

COMMENT ON TABLE video_access_log IS 'Video access and viewing analytics';
```

---

### **Video Tags Table** (Many-to-Many):
```sql
CREATE TABLE video_tags (
    video_id INTEGER REFERENCES videos(id) ON DELETE CASCADE,
    tag_id INTEGER REFERENCES tags(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (video_id, tag_id)
);

CREATE INDEX idx_video_tags_video_id ON video_tags(video_id);
CREATE INDEX idx_video_tags_tag_id ON video_tags(tag_id);
```

---

### **Video Categories Table**:
```sql
CREATE TABLE video_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) UNIQUE,
    description TEXT,
    thumbnail_url VARCHAR(500),
    parent_category_id INTEGER REFERENCES video_categories(id),
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_video_categories_slug ON video_categories(slug);
CREATE INDEX idx_video_categories_parent ON video_categories(parent_category_id);
```

---

## 🌐 **API Endpoints**

### **Video Management**:
**File**: `backend/internal/routes/video.go`

```
GET    /api/v1/videos                      # List videos (public + user access)
GET    /api/v1/videos/:id                  # Get video details
GET    /api/v1/videos/slug/:slug           # Get video by slug
POST   /api/v1/videos                      # Create video (admin)
PUT    /api/v1/videos/:id                  # Update video (admin)
DELETE /api/v1/videos/:id                  # Delete video (admin)
POST   /api/v1/videos/:id/publish          # Publish video (admin)
POST   /api/v1/videos/:id/archive          # Archive video (admin)
```

---

### **Video Streaming**:
```
GET    /api/v1/videos/:id/stream           # Get streaming URL (signed)
POST   /api/v1/videos/:id/view             # Track video view
POST   /api/v1/videos/:id/progress         # Update watch progress
GET    /api/v1/videos/:id/analytics        # Get video analytics
```

---

### **Video Upload**:
```
POST   /api/v1/videos/upload/init          # Initialize upload
POST   /api/v1/videos/upload/chunk         # Upload chunk
POST   /api/v1/videos/upload/complete      # Complete upload
GET    /api/v1/videos/upload/status/:id    # Check upload status
```

---

### **Master Video Sync** (YouTube):
**File**: `backend/internal/routes/master_video_routes.go`

```
GET    /api/v1/admin/master-videos         # List YouTube videos
POST   /api/v1/admin/master-videos/sync    # Sync specific video
POST   /api/v1/admin/master-videos/sync-all # Sync all pending
GET    /api/v1/admin/master-videos/:id/status # Check sync status
```

---

### **Categories & Tags**:
```
GET    /api/v1/videos/categories           # List categories
GET    /api/v1/videos/categories/:id/videos # Videos in category
GET    /api/v1/videos/tags                 # List tags
GET    /api/v1/videos/tags/:id/videos      # Videos with tag
```

---

## 🔧 **Backend Services**

### **Bunny Service** (`backend/internal/services/bunny.go`):

**Key Functions**:
```go
// Video Upload
func UploadVideoToBunny(videoFile io.Reader, filename string) (*BunnyVideo, error)
func CreateVideoInBunnyLibrary(title string, collectionID string) (*BunnyVideo, error)

// Video Management
func GetBunnyVideo(videoID string) (*BunnyVideo, error)
func UpdateBunnyVideo(videoID string, updates *BunnyVideoUpdate) error
func DeleteBunnyVideo(videoID string) error

// Video Processing
func GetVideoStatus(videoID string) (*ProcessingStatus, error)
func GenerateThumbnail(videoID string, timeOffset int) (string, error)
func SetVideoCaption(videoID string, captionFile io.Reader, language string) error

// Streaming URLs
func GetStreamingURL(videoID string, expiresIn time.Duration) (string, error)
func GetSignedStreamURL(videoID string, userID int, expiresIn time.Duration) (string, error)

// Collections (Playlists)
func CreateCollection(name string, description string) (*BunnyCollection, error)
func AddVideoToCollection(collectionID string, videoID string) error
func RemoveVideoFromCollection(collectionID string, videoID string) error

// Analytics
func GetVideoAnalytics(videoID string, dateRange DateRange) (*VideoAnalytics, error)
func GetLibraryAnalytics(dateRange DateRange) (*LibraryAnalytics, error)
```

---

### **Bunny Optimized Service** (`backend/internal/services/bunny_optimized.go`):

**Purpose**: Performance-optimized Bunny.net operations with caching

```go
// Cached Operations
func GetBunnyVideoOptimized(videoID string) (*BunnyVideo, error) {
    // Check cache first
    cached := cache.Get("bunny_video:" + videoID)
    if cached != nil {
        return cached, nil
    }
    
    // Fetch from Bunny.net
    video := GetBunnyVideo(videoID)
    
    // Cache for 5 minutes
    cache.Set("bunny_video:" + videoID, video, 5*time.Minute)
    
    return video, nil
}

// Batch Operations
func GetMultipleVideosOptimized(videoIDs []string) ([]*BunnyVideo, error) {
    // Fetch all videos concurrently
    results := make(chan *BunnyVideo, len(videoIDs))
    
    for _, id := range videoIDs {
        go func(videoID string) {
            video, _ := GetBunnyVideoOptimized(videoID)
            results <- video
        }(id)
    }
    
    // Collect results
    videos := []*BunnyVideo{}
    for i := 0; i < len(videoIDs); i++ {
        videos = append(videos, <-results)
    }
    
    return videos, nil
}
```

---

### **Master Video Sync Service** (`backend/internal/services/master_video_sync.go`):

**Purpose**: Sync YouTube videos to Bunny.net

```go
func SyncMasterVideo(masterVideoID int) error {
    // 1. Get master video record
    masterVideo := GetMasterVideo(masterVideoID)
    
    // 2. Download from YouTube
    videoFile, err := DownloadYouTubeVideo(masterVideo.YouTubeVideoID)
    if err != nil {
        return err
    }
    
    // 3. Upload to Bunny.net
    bunnyVideo, err := UploadVideoToBunny(videoFile, masterVideo.Title)
    if err != nil {
        masterVideo.SyncStatus = "failed"
        masterVideo.SyncError = err.Error()
        UpdateMasterVideo(masterVideo)
        return err
    }
    
    // 4. Create local video record
    video := &Video{
        Title:          masterVideo.Title,
        Description:    masterVideo.Description,
        BunnyVideoID:   bunnyVideo.GUID,
        BunnyStreamURL: bunnyVideo.StreamURL,
        Duration:       bunnyVideo.Duration,
        Status:         "processing",
    }
    CreateVideo(video)
    
    // 5. Update master video sync status
    masterVideo.SyncStatus = "synced"
    masterVideo.VideoID = video.ID
    masterVideo.LastSyncAt = time.Now()
    UpdateMasterVideo(masterVideo)
    
    return nil
}

func SyncAllPendingVideos() error {
    pending := GetPendingMasterVideos()
    
    for _, masterVideo := range pending {
        go SyncMasterVideo(masterVideo.ID) // Async
    }
    
    return nil
}
```

---

## 🎥 **Video Upload Flow**

### **Complete Upload Process**:
```
1. Client initiates upload
   POST /api/v1/videos/upload/init
   └─> Creates video record in database
       └─> Returns upload session ID

2. Client uploads file in chunks
   POST /api/v1/videos/upload/chunk
   └─> Streams directly to Bunny.net
       └─> Bunny.net handles transcoding

3. Client completes upload
   POST /api/v1/videos/upload/complete
   └─> Bunny.net processes video
       └─> Generates multiple quality levels
           └─> Creates thumbnails
               └─> Updates video status to "published"

4. Video ready for streaming
   └─> HLS manifest generated
       └─> CDN cache populated
           └─> Video available globally
```

---

## 🔒 **Access Control**

### **Subscription-Based Access**:
```go
func CanUserAccessVideo(userID int, videoID int) (bool, error) {
    video := GetVideo(videoID)
    
    // Free content
    if video.AccessLevel == "free" {
        return true, nil
    }
    
    // Get user's subscription
    subscription := GetUserSubscription(userID)
    if subscription == nil {
        return false, errors.New("subscription required")
    }
    
    // Check subscription tier
    if !subscription.HasAccess(video.AccessLevel) {
        return false, errors.New("higher subscription tier required")
    }
    
    // Check subscription status
    if subscription.Status != "active" {
        return false, errors.New("subscription inactive")
    }
    
    return true, nil
}
```

---

## 📊 **Video Analytics**

### **Metrics Tracked**:
```go
type VideoAnalytics struct {
    VideoID              int
    TotalViews           int
    UniqueViewers        int
    TotalWatchTime       int64 // seconds
    AverageWatchTime     int   // seconds
    CompletionRate       float64 // percentage
    LikeCount            int
    CommentCount         int
    ShareCount           int
    
    // Geographic
    TopCountries         []CountryStats
    TopCities            []CityStats
    
    // Devices
    DeviceBreakdown      map[string]int // desktop, mobile, tablet
    BrowserBreakdown     map[string]int
    OSBreakdown          map[string]int
    
    // Quality
    QualityBreakdown     map[string]int // sd, hd, 4k
    AverageBitrate       int
    BufferingEvents      int
    ErrorRate            float64
}

func GetVideoAnalytics(videoID int, dateRange DateRange) (*VideoAnalytics, error) {
    // Aggregate from video_access_log table
    analytics := &VideoAnalytics{VideoID: videoID}
    
    // Count views
    db.QueryRow(`
        SELECT COUNT(*), COUNT(DISTINCT user_id), SUM(watch_duration_seconds), AVG(completion_percentage)
        FROM video_access_log
        WHERE video_id = $1 AND accessed_at BETWEEN $2 AND $3
    `, videoID, dateRange.Start, dateRange.End).Scan(
        &analytics.TotalViews,
        &analytics.UniqueViewers,
        &analytics.TotalWatchTime,
        &analytics.CompletionRate,
    )
    
    // Get geographic breakdown
    // Get device breakdown
    // Get quality breakdown
    
    return analytics, nil
}
```

---

## ⚡ **Performance Optimizations**

### **Caching Strategy**:
- **Video metadata**: 10-minute cache
- **Bunny.net video status**: 5-minute cache
- **Streaming URLs**: 1-hour cache (signed)
- **Category listings**: 30-minute cache
- **Thumbnail URLs**: 1-day cache

### **CDN Optimization**:
- Global edge caching
- Adaptive bitrate streaming
- Prefetch next video segments
- Thumbnail sprite sheets

### **Database Optimization**:
- Indexed video lookups
- Partitioned analytics tables
- Read replicas for analytics queries
- Materialized views for popular videos

---

## 📝 **Known Technical Debt**

### **Current Limitations**:
1. ⚠️ **No direct upload to Bunny.net** - goes through server
2. ⚠️ **Limited video processing options** - relies on Bunny defaults
3. ⚠️ **Basic search** - no fuzzy matching or relevance scoring
4. ⚠️ **Manual thumbnail selection** - no AI-powered suggestions
5. ⚠️ **Simple recommendation** - no ML-based suggestions
6. ⚠️ **No live streaming** - only VOD

### **Future Enhancements**:
1. ✅ Direct client-to-Bunny upload (presigned URLs)
2. ✅ Advanced video processing (watermarks, intro/outro)
3. ✅ Elasticsearch integration for search
4. ✅ AI thumbnail generation
5. ✅ ML-based video recommendations
6. ✅ Live streaming support
7. ✅ DVR/Time-shift capabilities
8. ✅ Interactive video features

---

## 🚀 **Quick Start**

### **Understanding Video System** (15 min):
1. Read this BRAID.md (7 min)
2. Review Bunny.net integration (5 min)
3. Check database schema (3 min)

### **Debugging Videos**:
1. Check video status in database
2. Verify Bunny video exists (bunny_video_id)
3. Check streaming URL generation
4. Review access control logic
5. Check video_access_log for viewing issues

---

**Last Updated**: October 14, 2025  
**Status**: Core business system  
**Technology**: Go + Bunny.net CDN  
**Frontend Counterpart**: `_frontend/braids/video-streaming/`

---

**Navigate**:  
[🏠 Master Index](../../BRAIDS_INDEX.md) | [🎨 Frontend Braid](../../_frontend/braids/video-streaming/BRAID.md)

