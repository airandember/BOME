# Braid: video-streaming

**Architecture:** Full-Stack Braid (Frontend to Backend)
**Last Updated:** 2025-10-17

---

## Backend Architecture

**Bunny.net-powered video delivery and CDN management**

---

## ðŸ”— **Cross-Repository Braid**

> **âš ï¸ IMPORTANT**: This is the **backend portion** of the Video Streaming Braid.  
> **Frontend portion**: See `_frontend/braids/video-streaming/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## ðŸ“‹ **Backend Overview**

**Purpose**: Server-side video management, CDN integration, and streaming delivery  
**Technology**: Go, PostgreSQL, Bunny.net CDN  
**Complexity**: **Very High** (CDN Integration, Video Processing, Adaptive Streaming)  
**Priority**: **CRITICAL** - Core business functionality!

---

## 📁 **Production File Map**

### **Backend Files (Go)**
```
backend/
├── video-streaming/
│   ├── handlers/video.go               # Video routes
│   ├── handlers/master_video_routes.go # Master video sync routes
│   ├── services/bunny.go               # Bunny.net integration
│   ├── services/bunny_optimized.go     # Performance optimized
│   └── models/video.go, master_video.go
├── internal/
│   ├── routes/video.go                 # Video API routes
│   ├── routes/master_video_routes.go  # Master video routes
│   ├── routes/search_index.go          # Video search
│   ├── database/video.go
│   ├── database/master_video.go
│   └── services/
│       ├── bunny.go, bunny_optimized.go
│       └── master_video_sync.go        # Video synchronization
└── braids/video-analytics/             # Video analytics sub-braid
```

### **Frontend Files (Svelte)**
```
frontend/src/
├── routes/videos/                      # Video listing, playback
└── lib/components/                    # VideoCard, player components
```

---

## ðŸŽ¯ **Key Features**

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

## ðŸ—„ï¸ **Database Schema**

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

## ðŸŒ **API Endpoints**

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

## ðŸ”§ **Backend Services**

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

## ðŸŽ¥ **Video Upload Flow**

### **Complete Upload Process**:
```
1. Client initiates upload
   POST /api/v1/videos/upload/init
   â””â”€> Creates video record in database
       â””â”€> Returns upload session ID

2. Client uploads file in chunks
   POST /api/v1/videos/upload/chunk
   â””â”€> Streams directly to Bunny.net
       â””â”€> Bunny.net handles transcoding

3. Client completes upload
   POST /api/v1/videos/upload/complete
   â””â”€> Bunny.net processes video
       â””â”€> Generates multiple quality levels
           â””â”€> Creates thumbnails
               â””â”€> Updates video status to "published"

4. Video ready for streaming
   â””â”€> HLS manifest generated
       â””â”€> CDN cache populated
           â””â”€> Video available globally
```

---

## ðŸ”’ **Access Control**

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

## ðŸ“Š **Video Analytics**

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

## âš¡ **Performance Optimizations**

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

## ðŸ“ **Known Technical Debt**

### **Current Limitations**:
1. âš ï¸ **No direct upload to Bunny.net** - goes through server
2. âš ï¸ **Limited video processing options** - relies on Bunny defaults
3. âš ï¸ **Basic search** - no fuzzy matching or relevance scoring
4. âš ï¸ **Manual thumbnail selection** - no AI-powered suggestions
5. âš ï¸ **Simple recommendation** - no ML-based suggestions
6. âš ï¸ **No live streaming** - only VOD

### **Future Enhancements**:
1. âœ… Direct client-to-Bunny upload (presigned URLs)
2. âœ… Advanced video processing (watermarks, intro/outro)
3. âœ… Elasticsearch integration for search
4. âœ… AI thumbnail generation
5. âœ… ML-based video recommendations
6. âœ… Live streaming support
7. âœ… DVR/Time-shift capabilities
8. âœ… Interactive video features

---

## ðŸš€ **Quick Start**

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
[ðŸ  Master Index](../../BRAIDS_INDEX.md) | [ðŸŽ¨ Frontend Braid](../../_frontend/braids/video-streaming/BRAID.md)



---

## Frontend Architecture

**Svelte5 video player and content discovery**

---

## ðŸ”— **Cross-Repository Braid**

> **âš ï¸ IMPORTANT**: This is the **frontend portion** of the Video Streaming Braid.  
> **Backend portion**: See `_braids/video-streaming/backend/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## ðŸ“‹ **Frontend Overview**

**Purpose**: User interface for video browsing, playback, and content discovery  
**Technology**: Svelte 5, TypeScript, Video.js/Plyr, TailwindCSS  
**Entry Points**: `/videos`, `/videos/[id]`, `/videos/categories/[name]`  
**State Management**: Svelte stores for video library and player state

---

## ðŸŽ¯ **Key Features**

### **1. Video Library Page**:
- Grid/list view of videos
- Category filtering
- Tag filtering
- Search functionality
- Sort options (newest, popular, trending)
- Pagination or infinite scroll
- Featured videos section

### **2. Video Player Page**:
- Full-featured video player
- Adaptive bitrate streaming (HLS)
- Quality selection (SD, HD, 4K)
- Playback speed control
- Subtitles/captions
- Picture-in-picture
- Fullscreen support
- Keyboard shortcuts
- Thumbnail preview scrubbing

### **3. Video Details**:
- Title and description
- View count and date
- Like/dislike buttons
- Share functionality
- Related videos sidebar
- Comments section
- Playlist add button

### **4. Category Browsing**:
- Category landing pages
- Subcategory navigation
- Category-specific filters
- Category thumbnails

### **5. Video Collections**:
- Curated playlists
- User-created playlists
- Watch later
- History
- Favorites

---

## ðŸ“„ **Frontend Pages**

### **1. Videos Library** (`/videos`)
**File**: `frontend/src/routes/videos/+page.svelte`

**Features**:
- Video grid with thumbnails
- Filter by category/tag
- Search bar
- Sort dropdown
- Load more/infinite scroll

**Example UI**:
```svelte
<script lang="ts">
  import { onMount } from 'svelte';
  import { videoStore } from '$lib/stores/videos';
  import VideoCard from '$lib/components/VideoCard.svelte';
  
  let filter = {
    category: '',
    tag: '',
    search: '',
    sort: 'newest'
  };
  
  onMount(async () => {
    await videoStore.loadVideos(filter);
  });
  
  $: videos = $videoStore.videos;
  $: loading = $videoStore.loading;
  
  async function loadMore() {
    await videoStore.loadMore();
  }
</script>

<div class="videos-page">
  <header>
    <h1>Videos</h1>
    
    <!-- Search Bar -->
    <div class="search-bar">
      <input
        type="text"
        placeholder="Search videos..."
        bind:value={filter.search}
        on:input={() => videoStore.search(filter.search)}
      />
    </div>
    
    <!-- Filters -->
    <div class="filters">
      <select bind:value={filter.category}>
        <option value="">All Categories</option>
        <option value="tutorials">Tutorials</option>
        <option value="reviews">Reviews</option>
        <option value="news">News</option>
      </select>
      
      <select bind:value={filter.sort}>
        <option value="newest">Newest First</option>
        <option value="popular">Most Popular</option>
        <option value="trending">Trending</option>
      </select>
    </div>
  </header>
  
  <!-- Video Grid -->
  <div class="video-grid">
    {#if loading && videos.length === 0}
      <div class="loading">Loading videos...</div>
    {:else}
      {#each videos as video (video.id)}
        <VideoCard {video} />
      {/each}
    {/if}
  </div>
  
  <!-- Load More -->
  {#if !loading && videos.length > 0}
    <button on:click={loadMore} class="load-more">
      Load More
    </button>
  {/if}
</div>

<style>
  .video-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 2rem;
  }
  
  @media (max-width: 768px) {
    .video-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
```

---

### **2. Video Player Page** (`/videos/[id]`)
**File**: `frontend/src/routes/videos/[id]/+page.svelte`

**Features**:
- Video player embed
- Title, description, metadata
- Like/share/save buttons
- Related videos
- Comments section
- Up next autoplay

**Example UI**:
```svelte
<script lang="ts">
  import { page } from '$app/stores';
  import { onMount, onDestroy } from 'svelte';
  import { videoStore } from '$lib/stores/videos';
  import VideoPlayer from '$lib/components/VideoPlayer.svelte';
  
  const videoId = $page.params.id;
  
  let video = null;
  let relatedVideos = [];
  
  onMount(async () => {
    video = await videoStore.getVideo(videoId);
    relatedVideos = await videoStore.getRelatedVideos(videoId);
    
    // Track view
    videoStore.trackView(videoId);
  });
  
  function handleVideoEnded() {
    // Autoplay next video
    if (relatedVideos.length > 0) {
      goto(`/videos/${relatedVideos[0].id}`);
    }
  }
</script>

{#if video}
  <div class="video-page">
    <!-- Video Player -->
    <div class="player-container">
      <VideoPlayer
        src={video.streamUrl}
        poster={video.thumbnailUrl}
        on:ended={handleVideoEnded}
      />
    </div>
    
    <!-- Video Info -->
    <div class="video-info">
      <h1>{video.title}</h1>
      
      <div class="meta">
        <span>{video.viewCount.toLocaleString()} views</span>
        <span>â€¢</span>
        <span>{new Date(video.publishedAt).toLocaleDateString()}</span>
      </div>
      
      <!-- Actions -->
      <div class="actions">
        <button on:click={() => videoStore.likeVideo(video.id)}>
          ðŸ‘ {video.likeCount}
        </button>
        <button on:click={() => videoStore.shareVideo(video.id)}>
          ðŸ“¤ Share
        </button>
        <button on:click={() => videoStore.saveVideo(video.id)}>
          ðŸ’¾ Save
        </button>
      </div>
      
      <!-- Description -->
      <div class="description">
        <p>{video.description}</p>
      </div>
    </div>
    
    <!-- Related Videos Sidebar -->
    <aside class="related-videos">
      <h2>Up Next</h2>
      {#each relatedVideos as relatedVideo}
        <a href="/videos/{relatedVideo.id}" class="related-video-card">
          <img src={relatedVideo.thumbnailUrl} alt={relatedVideo.title} />
          <div>
            <h3>{relatedVideo.title}</h3>
            <p>{relatedVideo.viewCount} views</p>
          </div>
        </a>
      {/each}
    </aside>
  </div>
{/if}

<style>
  .video-page {
    display: grid;
    grid-template-columns: 1fr 350px;
    gap: 2rem;
    max-width: 1600px;
    margin: 0 auto;
  }
  
  .player-container {
    grid-column: 1;
    aspect-ratio: 16 / 9;
    background: black;
  }
  
  .video-info {
    grid-column: 1;
  }
  
  .related-videos {
    grid-column: 2;
    grid-row: 1 / span 2;
  }
  
  @media (max-width: 1024px) {
    .video-page {
      grid-template-columns: 1fr;
    }
    
    .related-videos {
      grid-column: 1;
      grid-row: auto;
    }
  }
</style>
```

---

### **3. Category Page** (`/videos/categories/[name]`)
**File**: `frontend/src/routes/videos/categories/[name]/+page.svelte`

**Features**:
- Category-specific video grid
- Category description
- Subcategory navigation
- Breadcrumb navigation

---

### **4. Collections Page** (`/videos/collections/[id]`)
**File**: `frontend/src/routes/videos/collections/[id]/+page.svelte`

**Features**:
- Playlist of videos
- Play all button
- Reorder videos (if owner)
- Share playlist
- Add to playlist button

---

## ðŸ§© **Frontend Components**

### **VideoCard Component** (`$lib/components/VideoCard.svelte`):
**Purpose**: Display video thumbnail and metadata

```svelte
<script lang="ts">
  export let video: Video;
  
  function formatDuration(seconds: number): string {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  }
</script>

<a href="/videos/{video.id}" class="video-card">
  <div class="thumbnail">
    <img src={video.thumbnailUrl} alt={video.title} loading="lazy" />
    <span class="duration">{formatDuration(video.duration)}</span>
    {#if video.requiresSubscription}
      <span class="badge premium">Premium</span>
    {/if}
  </div>
  
  <div class="info">
    <h3>{video.title}</h3>
    <p class="meta">
      {video.viewCount.toLocaleString()} views â€¢ 
      {new Date(video.publishedAt).toLocaleDateString()}
    </p>
  </div>
</a>

<style>
  .video-card {
    display: block;
    text-decoration: none;
    color: inherit;
    transition: transform 0.2s;
  }
  
  .video-card:hover {
    transform: translateY(-4px);
  }
  
  .thumbnail {
    position: relative;
    aspect-ratio: 16 / 9;
    border-radius: 8px;
    overflow: hidden;
  }
  
  .thumbnail img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .duration {
    position: absolute;
    bottom: 8px;
    right: 8px;
    background: rgba(0, 0, 0, 0.8);
    color: white;
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 0.875rem;
  }
  
  .badge.premium {
    position: absolute;
    top: 8px;
    left: 8px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: bold;
  }
</style>
```

---

### **VideoPlayer Component** (`$lib/components/VideoPlayer.svelte`):
**Purpose**: HTML5 video player with HLS support

```svelte
<script lang="ts">
  import { onMount } from 'svelte';
  import Hls from 'hls.js';
  
  export let src: string;
  export let poster: string = '';
  export let autoplay = false;
  
  let videoElement: HTMLVideoElement;
  let hls: Hls;
  
  onMount(() => {
    if (Hls.isSupported()) {
      hls = new Hls();
      hls.loadSource(src);
      hls.attachMedia(videoElement);
      
      hls.on(Hls.Events.MANIFEST_PARSED, () => {
        if (autoplay) {
          videoElement.play();
        }
      });
    } else if (videoElement.canPlayType('application/vnd.apple.mpegurl')) {
      // Native HLS support (Safari)
      videoElement.src = src;
    }
    
    return () => {
      if (hls) {
        hls.destroy();
      }
    };
  });
  
  function handleTimeUpdate() {
    // Track progress
    const progress = (videoElement.currentTime / videoElement.duration) * 100;
    dispatch('progress', { progress, currentTime: videoElement.currentTime });
  }
</script>

<video
  bind:this={videoElement}
  {poster}
  controls
  playsinline
  on:timeupdate={handleTimeUpdate}
  on:ended
  on:play
  on:pause
>
  <source src={src} type="application/x-mpegURL" />
  Your browser doesn't support video playback.
</video>

<style>
  video {
    width: 100%;
    height: 100%;
    background: black;
  }
</style>
```

---

## ðŸ—ƒï¸ **Frontend Stores**

### **Video Store** (`$lib/stores/videos.ts`):
**Purpose**: Manage video library state

```typescript
interface VideoState {
  videos: Video[];
  currentVideo: Video | null;
  relatedVideos: Video[];
  categories: Category[];
  loading: boolean;
  error: string | null;
  pagination: {
    page: number;
    total: number;
    hasMore: boolean;
  };
}

export const videoStore = {
  async loadVideos(filter: VideoFilter) {
    // GET /api/v1/videos?category=...&sort=...
  },
  
  async getVideo(id: string): Promise<Video> {
    // GET /api/v1/videos/:id
  },
  
  async getRelatedVideos(id: string): Promise<Video[]> {
    // GET /api/v1/videos/:id/related
  },
  
  async searchVideos(query: string): Promise<Video[]> {
    // GET /api/v1/videos?search=...
  },
  
  async trackView(id: string) {
    // POST /api/v1/videos/:id/view
  },
  
  async trackProgress(id: string, currentTime: number, duration: number) {
    // POST /api/v1/videos/:id/progress
  },
  
  async likeVideo(id: string) {
    // POST /api/v1/videos/:id/like
  },
  
  async saveVideo(id: string) {
    // POST /api/v1/videos/:id/save
  }
};
```

---

## ðŸ”„ **Data Flow Examples**

### **Watch Video Flow**:
```
1. User clicks video card
2. Navigate to /videos/:id
3. Load video metadata
4. Request streaming URL from backend
5. Backend generates signed URL
6. Initialize HLS player
7. Video streams from Bunny CDN
8. Track view event
9. Periodically track progress
10. On complete, suggest next video
```

### **Search Videos**:
```
1. User types in search bar
2. Debounce input (300ms)
3. API: GET /api/v1/videos?search=query
4. Backend searches database
5. Return matching videos
6. Update video grid
```

---

## ðŸŽ¨ **Video Player Features**

### **Controls**:
- â–¶ï¸ Play/Pause
- â®ï¸ Rewind 10s
- â­ï¸ Forward 10s
- ðŸ”Š Volume control
- âš™ï¸ Quality selection (SD, HD, 4K)
- ðŸŽ¬ Playback speed (0.5x, 1x, 1.5x, 2x)
- ðŸ“º Picture-in-picture
- â›¶ Fullscreen
- ðŸ’¬ Subtitles/CC

### **Keyboard Shortcuts**:
- `Space`: Play/pause
- `â†/â†’`: Seek 5s
- `â†‘/â†“`: Volume up/down
- `M`: Mute
- `F`: Fullscreen
- `C`: Toggle captions

---

## ðŸ“Š **Video Analytics Display**

### **Public Metrics**:
- View count
- Like/dislike ratio
- Publication date
- Duration

### **Creator Metrics** (if admin/creator):
- Real-time viewers
- Watch time distribution
- Audience retention graph
- Traffic sources
- Geographic distribution

---

## ðŸ”’ **Access Control UI**

### **Free vs. Premium Content**:
```svelte
{#if video.requiresSubscription && !$auth.user?.hasSubscription}
  <div class="locked-content">
    <div class="blur-overlay"></div>
    <div class="unlock-prompt">
      <h2>Premium Content</h2>
      <p>Subscribe to watch this video</p>
      <button on:click={() => goto('/subscription')}>
        Subscribe Now
      </button>
    </div>
  </div>
{:else}
  <VideoPlayer src={video.streamUrl} />
{/if}
```

---

## âš¡ **Performance Optimizations**

### **Lazy Loading**:
- Thumbnail images lazy load
- Defer non-critical video metadata
- Intersection Observer for infinite scroll

### **Caching**:
- Cache video metadata locally
- Cache thumbnails with service worker
- Prefetch next video in playlist

### **Network Optimization**:
- HLS adaptive streaming
- CDN edge caching
- Thumbnail sprite sheets for hover preview

---

## ðŸ“ **Known Issues**

### **To Implement**:
1. Video recommendations algorithm
2. Continue watching feature
3. Watch history
4. Playlist creation
5. Video chapters/timestamps
6. Interactive video elements
7. Live chat for live streams
8. Download for offline viewing (premium)

---

## ðŸš€ **Quick Links**

**Actual Files**:
- Videos Library: `frontend/src/routes/videos/+page.svelte`
- Video Player: `frontend/src/routes/videos/[id]/+page.svelte`
- Categories: `frontend/src/routes/videos/categories/[name]/+page.svelte`
- VideoCard Component: `frontend/src/lib/components/VideoCard.svelte`
- Video Store: `frontend/src/lib/stores/videos.ts`

---

**Last Updated**: October 14, 2025  
**Status**: Core business UI  
**Technology**: Svelte 5 + HLS.js  
**Backend Counterpart**: `_braids/video-streaming/backend/`

---

**Navigate**:  
[ðŸ  Master Index](../../../BRAIDS_INDEX.md) | [â¬…ï¸ Backend Braid](../../_braids/video-streaming/backend/BRAID.md)



---

## Integration Notes

- Frontend: `_braids/video-streaming/frontend/`
- Backend: `_braids/video-streaming/backend/`

This braid represents a complete vertical slice of functionality.

