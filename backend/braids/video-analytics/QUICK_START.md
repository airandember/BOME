# ⚡ Video Analytics Quick Start

## 🚀 **Get Tracking in 15 Minutes!**

> **Note**: This guide shows the basic implementation. For production-ready async buffering with Redis, 
> see `VIDEO_ANALYTICS_OPTIMIZATION_COMPLETE.md` for the optimized version with 10-40x better performance.

---

## Step 1: Add the Service (5 min)

Create `backend/internal/services/video_analytics_service.go`:

```go
package services

import (
	"bome-backend/internal/database"
	"fmt"
)

type VideoAnalyticsService struct {
	db    *database.DB
	redis *database.Redis  // Optional: for caching and buffering
}

// Basic version (synchronous)
func NewVideoAnalyticsService(db *database.DB, redis *database.Redis) *VideoAnalyticsService {
	return &VideoAnalyticsService{
		db:    db,
		redis: redis,
	}
}

type VideoTrackingRequest struct {
	VideoID           int     `json:"video_id" binding:"required"`
	UserID            *int    `json:"user_id"`
	SessionID         string  `json:"session_id"`
	WatchedDuration   int     `json:"watched_duration" binding:"required"`
	WatchedPercentage float64 `json:"watched_percentage" binding:"required"`
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
	
	if err != nil {
		return fmt.Errorf("failed to record view: %w", err)
	}
	
	return nil
}
```

---

## Step 2: Add the Route (5 min)

Add to your existing routes file or create `backend/internal/routes/video_analytics_routes.go`:

```go
package routes

import (
	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterVideoAnalyticsRoutes(router *gin.RouterGroup, db *database.DB, redis *database.Redis) {
	analytics := router.Group("/analytics")
	
	// Optional: require auth (can also work for anonymous users)
	// analytics.Use(middleware.AuthRequired())
	
	service := services.NewVideoAnalyticsService(db, redis)
	
	analytics.POST("/video/track", func(c *gin.Context) {
		var req services.VideoTrackingRequest
		
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Get user ID from auth context if available
		if userID, exists := c.Get("user_id"); exists {
			uid := userID.(int)
			req.UserID = &uid
		}
		
		// Ensure session ID exists
		if req.SessionID == "" {
			req.SessionID = c.GetHeader("X-Session-ID")
		}
		
		if err := service.RecordView(req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to track view",
			})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"status": "tracked",
			"video_id": req.VideoID,
		})
	})
}
```

Register in your main router:

```go
// In backend/internal/routes/routes.go or main.go
RegisterVideoAnalyticsRoutes(apiV1, db, redis)
```

---

## Step 3: Frontend Tracking (5 min)

Create `frontend/src/lib/services/videoAnalytics.ts`:

```typescript
export class VideoAnalyticsService {
  private sessionId: string;
  private lastReportedTime: Map<number, number> = new Map();
  
  constructor() {
    this.sessionId = this.getOrCreateSessionId();
  }
  
  /**
   * Track video progress - call this on timeupdate event
   */
  async trackProgress(videoId: number, currentTime: number, duration: number) {
    // Only report every 10 seconds
    const currentSecond = Math.floor(currentTime);
    if (currentSecond % 10 !== 0) return;
    
    // Don't report same second twice
    const lastReported = this.lastReportedTime.get(videoId) || 0;
    if (currentSecond === lastReported) return;
    
    this.lastReportedTime.set(videoId, currentSecond);
    
    try {
      await fetch('/api/v1/analytics/video/track', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': this.getAuthToken(),
          'X-Session-ID': this.sessionId
        },
        body: JSON.stringify({
          video_id: videoId,
          watched_duration: currentSecond,
          watched_percentage: Math.round((currentTime / duration) * 100 * 100) / 100
        })
      });
    } catch (error) {
      console.warn('Failed to track video progress:', error);
    }
  }
  
  /**
   * Mark video as complete - call on 'ended' event
   */
  async markComplete(videoId: number, duration: number) {
    try {
      await fetch('/api/v1/analytics/video/track', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': this.getAuthToken(),
          'X-Session-ID': this.sessionId
        },
        body: JSON.stringify({
          video_id: videoId,
          watched_duration: Math.floor(duration),
          watched_percentage: 100
        })
      });
    } catch (error) {
      console.warn('Failed to mark video complete:', error);
    }
  }
  
  private getOrCreateSessionId(): string {
    let sessionId = sessionStorage.getItem('video_session_id');
    if (!sessionId) {
      sessionId = `sess_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
      sessionStorage.setItem('video_session_id', sessionId);
    }
    return sessionId;
  }
  
  private getAuthToken(): string {
    // Get from your auth store
    const token = localStorage.getItem('auth_token');
    return token ? `Bearer ${token}` : '';
  }
}

// Singleton instance
export const videoAnalytics = new VideoAnalyticsService();
```

---

## Step 4: Add to Video Player

```svelte
<!-- Example: frontend/src/lib/components/VideoPlayer.svelte -->
<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { videoAnalytics } from '$lib/services/videoAnalytics';
  
  export let videoId: number;
  export let videoUrl: string;
  
  let player: HTMLVideoElement;
  let tracking = true;
  
  function handleTimeUpdate() {
    if (!player || !tracking) return;
    videoAnalytics.trackProgress(videoId, player.currentTime, player.duration);
  }
  
  function handleEnded() {
    if (!tracking) return;
    videoAnalytics.markComplete(videoId, player.duration);
  }
  
  function handlePlay() {
    tracking = true;
  }
  
  function handlePause() {
    // Continue tracking even when paused (records pause duration)
  }
  
  onDestroy(() => {
    tracking = false;
  });
</script>

<video 
  bind:this={player}
  src={videoUrl}
  on:timeupdate={handleTimeUpdate}
  on:ended={handleEnded}
  on:play={handlePlay}
  on:pause={handlePause}
  controls
  class="w-full rounded-lg"
>
  <track kind="captions" />
</video>
```

---

## ✅ **Test It!**

1. **Start your backend**: `go run main.go`
2. **Open your frontend**: Navigate to a video
3. **Check network tab**: Should see POST requests to `/analytics/video/track`
4. **Check database**:

```sql
SELECT * FROM video_views ORDER BY created_at DESC LIMIT 10;
```

You should see:
```
 id | video_id | user_id | session_id | watched_duration | watched_percentage | created_at
----+----------+---------+------------+------------------+--------------------+------------
  1 |      123 |    4567 | sess_abc   |               10 |              15.50 | 2025-11-22...
  2 |      123 |    4567 | sess_abc   |               20 |              31.00 | 2025-11-22...
```

---

## 🎉 **You're Done!**

### **What's Working:**
✅ Video views tracked every 10 seconds  
✅ Anonymous users tracked via session ID  
✅ Logged-in users tracked via user ID  
✅ Duration and percentage captured  

### **Next Steps:**
1. **Add Watch History**: Track user resume points
2. **Build Dashboard**: Show video stats
3. **Add Aggregation**: Daily metrics rollup
4. **Revenue Attribution**: Link to subscriptions

---

## 📊 **Quick Queries**

### **Most Viewed Videos (Last 7 Days)**
```sql
SELECT 
    v.title,
    COUNT(*) AS views,
    COUNT(DISTINCT COALESCE(vv.user_id, vv.session_id)) AS unique_viewers,
    AVG(vv.watched_percentage) AS avg_completion
FROM video_views vv
JOIN master_video_list v ON v.id = vv.video_id
WHERE vv.created_at > NOW() - INTERVAL '7 days'
GROUP BY v.id, v.title
ORDER BY views DESC
LIMIT 10;
```

### **User Watch History**
```sql
SELECT 
    v.title,
    vv.watched_duration,
    vv.watched_percentage,
    vv.created_at
FROM video_views vv
JOIN master_video_list v ON v.id = vv.video_id
WHERE vv.user_id = $1
ORDER BY vv.created_at DESC
LIMIT 20;
```

### **Video Engagement Stats**
```sql
SELECT 
    video_id,
    COUNT(*) AS total_views,
    COUNT(DISTINCT user_id) AS unique_viewers,
    AVG(watched_duration) AS avg_watch_time,
    AVG(watched_percentage) AS avg_completion,
    COUNT(CASE WHEN watched_percentage >= 95 THEN 1 END)::FLOAT / 
        COUNT(*)::FLOAT * 100 AS completion_rate
FROM video_views
WHERE video_id = $1
GROUP BY video_id;
```

---

## 🐛 **Troubleshooting**

### **No data in database?**
- Check browser console for errors
- Verify route is registered in main router
- Check backend logs for SQL errors
- Ensure `video_views` table exists

### **Events not firing?**
- Check video player has `on:timeupdate` handler
- Verify `videoAnalytics` service is imported
- Check network tab for 400/500 errors
- Ensure session ID is being generated

### **Too many requests?**
- Confirm 10-second throttling is working
- Check `lastReportedTime` map is preventing duplicates
- Consider increasing interval to 15 or 30 seconds

---

## 🎬 **You're Tracking Videos!**

Now you have:
- ✅ Real-time view tracking
- ✅ User engagement metrics
- ✅ Anonymous user support
- ✅ Database persistence

**Next**: Build the analytics dashboard! 📊🚀

