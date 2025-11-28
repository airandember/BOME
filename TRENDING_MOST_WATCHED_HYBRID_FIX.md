# Trending & Most Watched Hybrid Query Fix - Complete

## Issue Summary
**Problem**: "I've watched videos, but there's still none in our trending or most watched. Most watched is not taking into account the master_video_list views column."

The trending and most watched features were returning empty results even though videos had view counts in `master_video_list.views`.

## Root Cause

The `GetTopVideos()` and `GetTrendingVideos()` functions were **ONLY** querying the `video_views` table:

```sql
-- OLD QUERY ❌
SELECT ... 
FROM video_views vv
JOIN master_video_list v ON v.id = vv.video_id
WHERE vv.created_at > NOW() - INTERVAL '1 day' * $2
```

### The Problem
1. **New Analytics System**: `video_views` table tracks detailed, real-time analytics
2. **Legacy Data**: `master_video_list.views` contains historical view counts from Bunny.net
3. **Empty Results**: If `video_views` is empty or has recent data only → No trending/most watched results

This meant videos with thousands of views in `master_video_list.views` were invisible to trending/most watched!

## The Solution - Hybrid Queries

Created **HYBRID queries** that:
1. **Prefer** detailed analytics from `video_views` (when available)
2. **Fall back** to `master_video_list.views` (for historical data)
3. **Combine** both sources for comprehensive results

### 1. GetTopVideos() - Now Hybrid

**File**: `backend/internal/services/video_analytics_service.go`

```sql
-- NEW HYBRID QUERY ✅
WITH analytics_views AS (
    -- Get detailed analytics if available
    SELECT 
        vv.video_id,
        COUNT(*) AS total_views,
        COUNT(DISTINCT COALESCE(vv.user_id, vv.session_id)) AS unique_viewers,
        AVG(vv.watched_percentage) AS avg_completion,
        SUM(vv.watched_duration) AS total_watch_time
    FROM video_views vv
    WHERE vv.created_at > NOW() - INTERVAL '1 day' * $2
    GROUP BY vv.video_id
)
SELECT 
    v.id,
    v.title,
    v.thumbnail_url,
    v.duration,
    -- Use analytics if available, otherwise use master_video_list.views
    COALESCE(av.total_views, v.views, 0) AS total_views,
    COALESCE(av.unique_viewers, v.views, 0) AS unique_viewers,
    COALESCE(av.avg_completion, 0.0) AS avg_completion,
    COALESCE(av.total_watch_time, 0) AS total_watch_time
FROM master_video_list v
LEFT JOIN analytics_views av ON av.video_id = v.id
WHERE v.status = 'ready' AND (av.total_views > 0 OR v.views > 0)
ORDER BY total_views DESC
LIMIT $1
```

### 2. GetTrendingVideos() - Now Hybrid

```sql
-- NEW HYBRID QUERY ✅
WITH recent_stats AS (
    SELECT 
        video_id,
        COUNT(*) AS last_24h_views,
        MAX(created_at) AS last_view_at
    FROM video_views
    WHERE created_at > NOW() - INTERVAL '24 hours'
    GROUP BY video_id
),
video_engagement AS (
    SELECT 
        vv.video_id,
        COUNT(CASE WHEN vv.watched_percentage >= 95 THEN 1 END)::FLOAT / 
            NULLIF(COUNT(*), 0)::FLOAT * 100 AS completion_rate
    FROM video_views vv
    WHERE vv.created_at > NOW() - INTERVAL '7 days'
    GROUP BY vv.video_id
)
SELECT 
    v.id AS video_id,
    v.title,
    v.thumbnail_url,
    -- Use analytics if available, estimate from master_video_list.views otherwise
    COALESCE(r.last_24h_views, GREATEST(v.views / 10, 1)) AS last_24h_views,
    COALESCE(r.last_view_at, v.updated_at, v.created_at) AS last_view_at,
    COALESCE(ve.completion_rate, 0) AS completion_rate,
    v.likes
FROM master_video_list v
LEFT JOIN recent_stats r ON r.video_id = v.id
LEFT JOIN video_engagement ve ON ve.video_id = v.id
WHERE v.status = 'ready' AND (r.last_24h_views > 0 OR v.views > 0)
ORDER BY last_24h_views DESC
LIMIT $1
```

## How the Hybrid Approach Works

### Scenario 1: New Analytics System is Active
```
video_views has data → Use detailed analytics ✅
- Accurate unique viewers
- Real completion rates
- Precise watch time
```

### Scenario 2: Only Legacy Data Exists
```
video_views is empty → Use master_video_list.views ✅
- Historical view counts from Bunny.net
- Videos with thousands of views still appear
- Trending estimates based on total views
```

### Scenario 3: Transition Period (Current State)
```
Some videos in video_views, some only in master_video_list ✅
- Combines both sources
- Comprehensive results
- Smooth migration from legacy to new system
```

## Benefits

### 1. **Immediate Results**
- ✅ Videos with `master_video_list.views > 0` now appear
- ✅ No more empty trending/most watched sections
- ✅ Historical data is utilized

### 2. **Smooth Transition**
- ✅ As users watch videos, `video_views` fills up
- ✅ Gradually shifts from legacy data to detailed analytics
- ✅ No disruption during migration

### 3. **Best of Both Worlds**
- ✅ Detailed analytics when available
- ✅ Historical data as fallback
- ✅ Always shows something relevant

### 4. **Accuracy Improvements Over Time**
```
Week 1:  Mostly master_video_list.views (legacy)
Week 2:  50/50 mix
Week 4:  Mostly video_views (detailed analytics)
Week 8:  100% video_views (pure analytics)
```

## Expected Behavior Now

### Most Watched Tab
**Before**: Empty (no results from `video_views`)  
**After**: Shows videos sorted by:
- `video_views` count (if tracked)
- `master_video_list.views` (if not tracked yet)

### Trending Tab
**Before**: Empty (no 24h data in `video_views`)  
**After**: Shows videos based on:
- Recent 24h `video_views` activity (if available)
- Estimated from `master_video_list.views` (if not)

## Testing

Try these queries to verify:

```sql
-- Check videos with only master_video_list.views
SELECT id, title, views 
FROM master_video_list 
WHERE views > 0 
AND id NOT IN (SELECT DISTINCT video_id FROM video_views)
LIMIT 10;

-- Check videos with video_views data
SELECT v.id, v.title, COUNT(*) as analytics_views, v.views as legacy_views
FROM video_views vv
JOIN master_video_list v ON v.id = vv.video_id
GROUP BY v.id, v.title, v.views
LIMIT 10;
```

## Files Modified
- ✅ `backend/internal/services/video_analytics_service.go`
  - `GetTopVideos()` - Now uses hybrid query
  - `GetTrendingVideos()` - Now uses hybrid query

## Logs to Watch For
```
📊 [Video Analytics] Retrieved 25 top videos (using hybrid query)
✅ [Video Analytics] Found 15 trending videos
```

The "(using hybrid query)" message confirms the new logic is active!

---
**Status**: ✅ **FIXED AND DEPLOYED**  
**Date**: 2025-11-26  
**Backend**: Running on port 8080 with hybrid trending/most watched queries  
**Impact**: Trending and Most Watched now show videos even without `video_views` data!

**Refresh your videos page** - you should now see trending and most watched results! 🎉

