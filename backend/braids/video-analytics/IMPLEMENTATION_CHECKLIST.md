# ✅ Video Analytics Implementation Checklist

## 🎯 **Phase 1: Foundation** (Estimated: 2-3 days)

### **Database Layer** ✅ (Already Done!)
- [x] `video_views` table exists
- [x] `video_watch_history` table exists
- [x] `video_metrics` table exists
- [x] `user_metrics` table exists
- [x] Proper indexes created
- [x] Foreign key relationships

### **Service Layer** 📋 (To Do)
- [ ] Create `VideoAnalyticsService`
  - [ ] `RecordView(videoID, userID, duration, percentage)`
  - [ ] `GetVideoStats(videoID, period)`
  - [ ] `GetTrendingVideos(limit)`
- [ ] Create `WatchHistoryService`
  - [ ] `UpdateProgress(userID, videoID, position)`
  - [ ] `GetHistory(userID, videoID)`
  - [ ] `GetContinueWatching(userID)`
  - [ ] `MarkComplete(userID, videoID)`
- [ ] Add service tests

### **API Routes** 📋 (To Do)
- [ ] POST `/api/v1/analytics/video/track` - Track view event
- [ ] POST `/api/v1/analytics/video/:id/complete` - Mark complete
- [ ] GET `/api/v1/videos/:id/watch-history` - Get user progress
- [ ] GET `/api/v1/analytics/video/:id/stats` - Get video stats
- [ ] GET `/api/v1/analytics/trending` - Get trending videos

---

## 🎯 **Phase 2: Frontend Integration** (Estimated: 3-4 days)

### **Video Player Events** 📋 (To Do)
- [ ] Track `play` event
- [ ] Track `pause` event
- [ ] Track `timeupdate` (every 10 seconds)
- [ ] Track `seek` event
- [ ] Track `ended` event
- [ ] Send buffered events (don't spam backend)

### **Watch History UI** 📋 (To Do)
- [ ] "Continue Watching" section on homepage
- [ ] "Resume" button on video page
- [ ] Progress bar on video thumbnails
- [ ] "Completed" badge for finished videos

### **TypeScript Types** 📋 (To Do)
```typescript
interface VideoTrackingEvent {
  video_id: number;
  watched_duration: number;
  watched_percentage: number;
  session_id: string;
  event_type: 'play' | 'pause' | 'progress' | 'complete';
}

interface WatchHistory {
  video_id: number;
  last_position: number;
  completed: boolean;
  percentage: number;
  last_watched_at: string;
}
```

---

## 🎯 **Phase 3: Aggregation Jobs** (Estimated: 2-3 days)

### **Cron Jobs** 📋 (To Do)
- [ ] Create `AggregateVideoMetrics()` - Daily at 2 AM
- [ ] Create `CleanupOldEvents()` - Weekly
- [ ] Create `UpdateTrendingVideos()` - Hourly
- [ ] Add job monitoring/alerting

### **Aggregation Logic** 📋 (To Do)
```go
// Daily aggregation
func AggregateVideoMetrics(date time.Time) {
    // For each video with activity on 'date'
    // 1. Count views, unique_views
    // 2. Calculate watch_time, avg_watch_time
    // 3. Calculate completion_rate
    // 4. Calculate bounce_rate
    // 5. Upsert into video_metrics
}
```

### **Cleanup Jobs** 📋 (To Do)
```go
// Keep raw events for 90 days
func CleanupOldEvents() {
    db.Exec(`
        DELETE FROM video_views 
        WHERE created_at < NOW() - INTERVAL '90 days'
    `)
}
```

---

## 🎯 **Phase 4: Admin Dashboard** (Estimated: 4-5 days)

### **Video Analytics Page** 📋 (To Do)
- [ ] Video stats card
  - [ ] Total views
  - [ ] Unique viewers
  - [ ] Watch time
  - [ ] Engagement score
- [ ] Views over time chart (line graph)
- [ ] Drop-off heatmap
- [ ] Geographic distribution
- [ ] Subscriber vs Free breakdown

### **Trending Videos Widget** 📋 (To Do)
- [ ] Display top 10 trending
- [ ] Trending score badge
- [ ] 24h views indicator
- [ ] Auto-refresh every minute

### **Content Performance Report** 📋 (To Do)
- [ ] Table: All videos sorted by engagement
- [ ] Filters: Date range, category, status
- [ ] Export to CSV
- [ ] Email scheduled reports

---

## 🎯 **Phase 5: Revenue Attribution** (Estimated: 2-3 days)

### **Conversion Tracking** 📋 (To Do)
- [ ] Track which videos led to signups
- [ ] Track which videos led to subscriptions
- [ ] Calculate conversion rates per video
- [ ] Show in admin dashboard

### **Attribution Queries** 📋 (To Do)
```sql
-- Videos that drive subscriptions
SELECT 
    v.title,
    COUNT(DISTINCT vv.user_id) AS viewers,
    COUNT(DISTINCT s.user_id) AS subscribers,
    COUNT(DISTINCT s.user_id)::FLOAT / 
        COUNT(DISTINCT vv.user_id)::FLOAT * 100 AS conversion_rate,
    SUM(sp.price) AS revenue_generated
FROM video_views vv
JOIN master_video_list v ON v.id = vv.video_id
LEFT JOIN user_stripe_customers_v2 usc ON 
    usc.user_id = vv.user_id AND
    usc.created_at > vv.created_at AND
    usc.created_at < vv.created_at + INTERVAL '7 days'
LEFT JOIN stripe_subscriptions_v2 s ON s.customer_id = usc.stripe_customer_id
LEFT JOIN stripe_prices_v2 sp ON sp.id = s.price_id
GROUP BY v.id, v.title
ORDER BY revenue_generated DESC;
```

---

## 🎯 **Phase 6: Optimizations** (Estimated: 2-3 days)

### **Performance Tuning** 📋 (To Do)
- [ ] Add caching for trending videos (10 min TTL)
- [ ] Batch insert video_views (5 second buffer)
- [ ] Use read replica for analytics queries
- [ ] Add database connection pooling
- [ ] Implement query result caching

### **Monitoring** 📋 (To Do)
- [ ] Alert if aggregation job fails
- [ ] Alert if video tracking errors spike
- [ ] Dashboard for tracking system health
- [ ] Log slow queries

---

## 🔧 **Quick Start: Implement Core Tracking**

### **Step 1: Create Service**

```go
// backend/internal/services/video_analytics_service.go
package services

type VideoAnalyticsService struct {
    db    *database.DB
    redis *database.Redis  // Optional: for production optimization
}

// Note: For production with async buffering, see VIDEO_ANALYTICS_OPTIMIZATION_COMPLETE.md
func NewVideoAnalyticsService(db *database.DB, redis *database.Redis) *VideoAnalyticsService {
    return &VideoAnalyticsService{
        db:    db,
        redis: redis,
    }
}

type VideoTrackingRequest struct {
    VideoID           int     `json:"video_id"`
    UserID            *int    `json:"user_id"`
    SessionID         string  `json:"session_id"`
    WatchedDuration   int     `json:"watched_duration"`
    WatchedPercentage float64 `json:"watched_percentage"`
}

func (s *VideoAnalyticsService) RecordView(req VideoTrackingRequest) error {
    query := `
        INSERT INTO video_views (
            video_id, user_id, session_id, watched_duration, 
            watched_percentage, created_at
        ) VALUES ($1, $2, $3, $4, $5, NOW())
    `
    
    _, err := s.db.Exec(query, 
        req.VideoID, 
        req.UserID, 
        req.SessionID, 
        req.WatchedDuration, 
        req.WatchedPercentage,
    )
    
    return err
}
```

### **Step 2: Create Route**

```go
// backend/internal/routes/video_analytics_routes.go
package routes

func RegisterVideoAnalyticsRoutes(router *gin.RouterGroup, db *database.DB, redis *database.Redis) {
    analytics := router.Group("/analytics")
    {
        service := services.NewVideoAnalyticsService(db, redis)
        
        // Track video view
        analytics.POST("/video/track", func(c *gin.Context) {
            var req services.VideoTrackingRequest
            if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(400, gin.H{"error": err.Error()})
                return
            }
            
            // Get user ID from auth context (if logged in)
            if userID, exists := c.Get("user_id"); exists {
                uid := userID.(int)
                req.UserID = &uid
            }
            
            // Generate session ID if not provided
            if req.SessionID == "" {
                req.SessionID = generateSessionID(c)
            }
            
            if err := service.RecordView(req); err != nil {
                c.JSON(500, gin.H{"error": "Failed to track view"})
                return
            }
            
            c.JSON(200, gin.H{"status": "tracked"})
        })
    }
}
```

### **Step 3: Frontend Integration**

```typescript
// frontend/src/lib/services/videoAnalytics.ts
export class VideoAnalyticsService {
  private sessionId: string;
  private lastReportedTime: number = 0;
  
  constructor() {
    this.sessionId = this.getOrCreateSessionId();
  }
  
  async trackProgress(videoId: number, currentTime: number, duration: number) {
    // Only report every 10 seconds
    if (Math.floor(currentTime) % 10 !== 0) return;
    if (currentTime === this.lastReportedTime) return;
    
    this.lastReportedTime = currentTime;
    
    await fetch('/api/v1/analytics/video/track', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${getAuthToken()}`
      },
      body: JSON.stringify({
        video_id: videoId,
        watched_duration: Math.floor(currentTime),
        watched_percentage: (currentTime / duration) * 100,
        session_id: this.sessionId
      })
    });
  }
  
  async markComplete(videoId: number, duration: number) {
    await fetch('/api/v1/analytics/video/track', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${getAuthToken()}`
      },
      body: JSON.stringify({
        video_id: videoId,
        watched_duration: Math.floor(duration),
        watched_percentage: 100,
        session_id: this.sessionId
      })
    });
  }
  
  private getOrCreateSessionId(): string {
    let sessionId = sessionStorage.getItem('video_session_id');
    if (!sessionId) {
      sessionId = `sess_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
      sessionStorage.setItem('video_session_id', sessionId);
    }
    return sessionId;
  }
}
```

```svelte
<!-- frontend/src/lib/components/VideoPlayer.svelte -->
<script lang="ts">
  import { VideoAnalyticsService } from '$lib/services/videoAnalytics';
  
  export let videoId: number;
  export let videoUrl: string;
  
  let player: HTMLVideoElement;
  const analytics = new VideoAnalyticsService();
  
  function handleTimeUpdate() {
    if (!player) return;
    analytics.trackProgress(videoId, player.currentTime, player.duration);
  }
  
  function handleEnded() {
    analytics.markComplete(videoId, player.duration);
  }
</script>

<video 
  bind:this={player}
  src={videoUrl}
  on:timeupdate={handleTimeUpdate}
  on:ended={handleEnded}
  controls
>
  <track kind="captions" />
</video>
```

---

## 🧪 **Testing Checklist**

### **Unit Tests** 📋
- [ ] Test `RecordView()` with valid data
- [ ] Test `RecordView()` with anonymous user
- [ ] Test `GetVideoStats()` calculations
- [ ] Test `GetTrendingVideos()` algorithm
- [ ] Test watch history update logic

### **Integration Tests** 📋
- [ ] Test video tracking API end-to-end
- [ ] Test watch history resume
- [ ] Test metrics aggregation job
- [ ] Test cleanup job

### **Load Tests** 📋
- [ ] 1000 concurrent video views
- [ ] 10,000 tracking events/minute
- [ ] Dashboard load time <1s with 1M views

---

## 📊 **Success Metrics**

| Metric | Target | How to Measure |
|--------|--------|----------------|
| **Tracking Success Rate** | >99.5% | Monitor error logs |
| **API Response Time** | <100ms | APM dashboard |
| **Aggregation Job Time** | <5 min | Cron job logs |
| **Dashboard Load Time** | <1s | Frontend monitoring |
| **Data Accuracy** | >99% | Spot checks vs Stripe |

---

## 🚀 **Ready to Start?**

1. **Start with Phase 1**: Implement `VideoAnalyticsService`
2. **Add API routes**: `/analytics/video/track`
3. **Frontend integration**: Add to video player
4. **Test with real videos**: Watch full flow
5. **Add aggregation**: Daily metrics rollup
6. **Build dashboard**: Show the data!

**Let's track those videos! 🎬📊**

