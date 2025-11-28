# 🎬 Video Analytics BRAID - Architecture

## 🏗️ **System Architecture**

```
┌──────────────────────────────────────────────────────────────────┐
│                     VIDEO ANALYTICS SYSTEM                        │
│                                                                    │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐     │
│  │ Event Tracking │  │ State Tracking │  │  Aggregation   │     │
│  │  (Real-time)   │  │   (Per User)   │  │   (Batch)      │     │
│  └────────┬───────┘  └────────┬───────┘  └────────┬───────┘     │
│           │                    │                    │              │
│           ↓                    ↓                    ↓              │
│  ┌────────────────────────────────────────────────────────┐      │
│  │              DATABASE LAYER                             │      │
│  │  • video_views (raw events)                            │      │
│  │  • video_watch_history (user state)                    │      │
│  │  • video_metrics (daily rollups)                       │      │
│  │  • user_metrics (user engagement)                      │      │
│  └────────────────────────────────────────────────────────┘      │
└──────────────────────────────────────────────────────────────────┘
```

---

## 📊 **Three-Tier Analytics Model**

### **Tier 1: Raw Events (video_views)**
**Purpose**: Capture every video interaction  
**Retention**: 90 days  
**Volume**: High (millions of rows)

```sql
CREATE TABLE video_views (
    id SERIAL PRIMARY KEY,
    video_id INTEGER REFERENCES master_video_list(id),
    user_id INTEGER REFERENCES users(id),      -- NULL for anonymous
    session_id VARCHAR(255),                    -- Track anonymous users
    ip_address INET,                            -- Geolocation
    watched_duration INTEGER DEFAULT 0,         -- Seconds watched
    watched_percentage DECIMAL(5,2) DEFAULT 0,  -- % of video
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Use Cases:**
- Real-time "X people watching" counter
- Individual user behavior analysis
- Fraud detection (bot views)
- Geographic distribution

---

### **Tier 2: User State (video_watch_history)**
**Purpose**: Track where each user left off  
**Retention**: Permanent (or until user deletes)  
**Volume**: Medium (one row per user per video)

```sql
CREATE TABLE video_watch_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    video_id INTEGER REFERENCES master_video_list(id),
    last_position INTEGER DEFAULT 0,           -- Resume from here
    completed BOOLEAN DEFAULT FALSE,            -- Finished?
    first_watched_at TIMESTAMP DEFAULT NOW(),   -- Discovery date
    last_watched_at TIMESTAMP DEFAULT NOW(),    -- Most recent view
    UNIQUE(user_id, video_id)
);
```

**Use Cases:**
- "Continue Watching" feature
- "You completed this video" badge
- Binge-watching patterns
- Content recommendations

---

### **Tier 3: Aggregated Metrics (video_metrics)**
**Purpose**: Pre-calculated daily statistics  
**Retention**: Permanent  
**Volume**: Low (one row per video per day)

```sql
CREATE TABLE video_metrics (
    id SERIAL PRIMARY KEY,
    video_id INTEGER REFERENCES master_video_list(id),
    date DATE NOT NULL,
    views INTEGER DEFAULT 0,                    -- Total plays
    unique_views INTEGER DEFAULT 0,             -- Distinct users
    watch_time INTEGER DEFAULT 0,               -- Total seconds
    completion_rate DECIMAL(5,2) DEFAULT 0.00,  -- % who finished
    likes INTEGER DEFAULT 0,
    comments INTEGER DEFAULT 0,
    shares INTEGER DEFAULT 0,
    bounce_rate DECIMAL(5,2) DEFAULT 0.00,      -- % left <10s
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(video_id, date)
);
```

**Use Cases:**
- Admin dashboards
- Trending videos algorithm
- Performance reports
- Year-over-year comparisons

---

## 🔄 **Data Flow Architecture**

### **1. Event Capture (Real-time)**

```typescript
// Frontend Video Player
player.on('timeupdate', async () => {
  if (Math.floor(player.currentTime) % 10 === 0) {
    // Report every 10 seconds
    await trackVideoProgress({
      video_id: currentVideo.id,
      watched_duration: Math.floor(player.currentTime),
      watched_percentage: (player.currentTime / player.duration) * 100,
      session_id: getSessionId()
    });
  }
});

player.on('ended', async () => {
  await markVideoComplete({
    video_id: currentVideo.id,
    watched_duration: Math.floor(player.duration)
  });
});
```

**Backend Route:**
```go
// POST /api/v1/analytics/video/track
func TrackVideoView(c *gin.Context) {
    var req VideoTrackingRequest
    c.BindJSON(&req)
    
    // Insert into video_views (raw event)
    videoAnalyticsService.RecordView(req)
    
    // Update video_watch_history (user state)
    watchHistoryService.UpdateProgress(req)
    
    c.JSON(200, gin.H{"status": "tracked"})
}
```

---

### **2. State Management (On-demand)**

```go
// GET /api/v1/videos/:id/watch-history
func GetWatchHistory(c *gin.Context) {
    videoID := c.Param("id")
    userID := c.Get("user_id")
    
    history := watchHistoryService.GetHistory(userID, videoID)
    
    c.JSON(200, gin.H{
        "last_position": history.LastPosition,
        "completed": history.Completed,
        "percentage": history.WatchedPercentage
    })
}
```

**Use Case:**
```typescript
// Frontend: Resume playback
const history = await fetch(`/api/v1/videos/${videoId}/watch-history`);
if (history.last_position > 0) {
  player.seekTo(history.last_position);
  showMessage(`Resuming from ${formatTime(history.last_position)}`);
}
```

---

### **3. Aggregation (Batch Jobs)**

```go
// Cron Job: Every night at 2 AM
func AggregateVideoMetrics() {
    yesterday := time.Now().AddDate(0, 0, -1)
    
    // For each video
    videos := db.GetAllVideos()
    for _, video := range videos {
        metrics := CalculateDailyMetrics(video.ID, yesterday)
        
        db.UpsertVideoMetrics(metrics)
        
        // Cleanup old raw events (>90 days)
        db.DeleteOldVideoViews(video.ID, 90)
    }
}

func CalculateDailyMetrics(videoID int, date time.Time) VideoMetrics {
    return VideoMetrics{
        VideoID:        videoID,
        Date:           date,
        Views:          CountViews(videoID, date),
        UniqueViews:    CountUniqueViewers(videoID, date),
        WatchTime:      SumWatchTime(videoID, date),
        CompletionRate: CalculateCompletionRate(videoID, date),
        BounceRate:     CalculateBounceRate(videoID, date),
    }
}
```

---

## 🎯 **Service Layer Architecture**

### **VideoAnalyticsService**

```go
type VideoAnalyticsService struct {
    db *database.DB
}

// Core Methods
func (s *VideoAnalyticsService) RecordView(req VideoTrackingRequest) error
func (s *VideoAnalyticsService) GetVideoStats(videoID int, period string) (*VideoStats, error)
func (s *VideoAnalyticsService) GetTrendingVideos(limit int) ([]TrendingVideo, error)
func (s *VideoAnalyticsService) GetUserEngagement(userID int) (*UserEngagement, error)
```

### **WatchHistoryService**

```go
type WatchHistoryService struct {
    db *database.DB
}

// Core Methods
func (s *WatchHistoryService) UpdateProgress(userID, videoID, position int) error
func (s *WatchHistoryService) MarkComplete(userID, videoID int) error
func (s *WatchHistoryService) GetHistory(userID, videoID int) (*WatchHistory, error)
func (s *WatchHistoryService) GetContinueWatching(userID int) ([]Video, error)
```

### **MetricsAggregationService**

```go
type MetricsAggregationService struct {
    db *database.DB
}

// Core Methods
func (s *MetricsAggregationService) AggregateDaily(date time.Time) error
func (s *MetricsAggregationService) CleanupOldEvents(retentionDays int) error
func (s *MetricsAggregationService) RecalculateMetrics(videoID int) error
```

---

## 📊 **Key Metrics Calculations**

### **1. Completion Rate**
```sql
SELECT 
    COUNT(DISTINCT CASE WHEN watched_percentage >= 95 THEN user_id END)::FLOAT /
    COUNT(DISTINCT user_id)::FLOAT * 100 AS completion_rate
FROM video_views
WHERE video_id = $1 AND date = $2;
```

### **2. Average Watch Time**
```sql
SELECT 
    AVG(watched_duration) AS avg_watch_time,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY watched_duration) AS median_watch_time
FROM video_views
WHERE video_id = $1 AND date = $2;
```

### **3. Bounce Rate** (left within 10 seconds)
```sql
SELECT 
    COUNT(CASE WHEN watched_duration <= 10 THEN 1 END)::FLOAT /
    COUNT(*)::FLOAT * 100 AS bounce_rate
FROM video_views
WHERE video_id = $1 AND date = $2;
```

### **4. Engagement Score** (0-100)
```sql
SELECT 
    (
        (completion_rate * 0.4) +               -- 40% weight
        (avg_watch_percentage * 0.3) +          -- 30% weight
        ((100 - bounce_rate) * 0.2) +           -- 20% weight
        (LEAST(likes_per_view * 100, 100) * 0.1)  -- 10% weight
    ) AS engagement_score
FROM video_metrics
WHERE video_id = $1;
```

---

## 🔐 **Privacy & Compliance**

### **Anonymous Tracking:**
```go
// For non-logged-in users
type AnonymousView struct {
    SessionID  string  // Browser fingerprint
    IPAddress  string  // For geolocation only
    UserAgent  string  // Device/browser stats
}

// Do NOT store:
// - Personal identifiable information
// - Exact user location
// - Browsing history outside platform
```

### **GDPR Compliance:**
```go
// User deletion request
func DeleteUserAnalyticsData(userID int) error {
    // Anonymize video_views
    db.Exec(`
        UPDATE video_views 
        SET user_id = NULL, 
            ip_address = NULL,
            session_id = 'deleted'
        WHERE user_id = $1
    `, userID)
    
    // Delete watch history
    db.Exec(`DELETE FROM video_watch_history WHERE user_id = $1`, userID)
    
    return nil
}
```

---

## 🚀 **Performance Optimizations**

### **1. Batching Inserts**
```go
// Buffer events for 5 seconds before batch insert
type EventBuffer struct {
    events []VideoView
    mu     sync.Mutex
    ticker *time.Ticker
}

func (b *EventBuffer) Flush() {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    if len(b.events) == 0 {
        return
    }
    
    // Bulk insert
    db.BulkInsertVideoViews(b.events)
    b.events = nil
}
```

### **2. Caching Hot Data**
```go
// Cache trending videos for 10 minutes
var trendingCache = cache.New(10*time.Minute, 15*time.Minute)

func GetTrendingVideos() ([]Video, error) {
    if cached, found := trendingCache.Get("trending"); found {
        return cached.([]Video), nil
    }
    
    trending := db.CalculateTrending()
    trendingCache.Set("trending", trending, cache.DefaultExpiration)
    return trending, nil
}
```

### **3. Read Replicas**
```go
// Heavy analytics queries go to read replica
analyticsDB := database.NewReadReplica()
stats := analyticsDB.GetVideoStats(videoID)
```

---

## 📈 **Trending Algorithm**

```go
func CalculateTrendingScore(video VideoMetrics) float64 {
    // Time decay: recent views matter more
    daysSinceLastView := time.Since(video.LastViewedAt).Hours() / 24
    timeDecay := math.Exp(-daysSinceLastView / 3) // Decay over 3 days
    
    // Velocity: views per hour
    velocity := float64(video.Last24HViews) / 24.0
    
    // Engagement: completion rate + likes
    engagement := (video.CompletionRate + float64(video.Likes)*10) / 2
    
    // Combined score
    return (velocity * 0.5 + engagement * 0.3) * timeDecay
}
```

---

## 🎊 **Integration with Subscription System**

### **Revenue Attribution:**
```sql
-- Which videos drive subscriptions?
SELECT 
    v.title,
    COUNT(DISTINCT vv.user_id) AS viewers,
    COUNT(DISTINCT CASE WHEN u.created_at > vv.created_at 
                         AND u.created_at < vv.created_at + INTERVAL '7 days'
                         THEN u.id END) AS subscriptions_within_7days,
    COUNT(DISTINCT CASE ... END)::FLOAT / 
    COUNT(DISTINCT vv.user_id)::FLOAT * 100 AS conversion_rate
FROM video_views vv
JOIN master_video_list v ON v.id = vv.video_id
LEFT JOIN user_stripe_customers_v2 usc ON usc.user_id = vv.user_id
LEFT JOIN users u ON u.id = vv.user_id
GROUP BY v.id, v.title
ORDER BY conversion_rate DESC;
```

---

## 🎯 **Success Criteria**

✅ Track 100% of video views  
✅ Resume playback within 2 seconds  
✅ Aggregate metrics in <5 minutes  
✅ Dashboard loads in <1 second  
✅ Handle 1000+ concurrent viewers  
✅ GDPR compliant data deletion  
✅ Revenue attribution accuracy >95%  

---

**Built on V2 Foundation - Ready for Production! 🚀**

