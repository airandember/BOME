# Video Analytics Optimization Proposal

## Current Problem 🚨

The `video_views` table is creating **a new row every 10 seconds** per viewer:

```
Current behavior:
- 10s: INSERT (10s watched)
- 20s: INSERT (20s watched)
- 30s: INSERT (30s watched)
... continues forever

Result with 500 concurrent viewers on 1-hour video:
= 360 rows/viewer × 500 viewers
= 180,000 rows per hour
= 4.3 MILLION rows per day! 💥
```

### Database Impact
- **Storage**: Explodes exponentially
- **Performance**: Queries slow down with millions of rows
- **Cost**: Unnecessary data that provides no extra value
- **Maintenance**: Cleanup/archival becomes critical

## Root Cause

Looking at `video_views` table sample:
```
user_id: NULL (all rows)
session_id: "sess_1763848388208_ohl3yb1gg" (same session)
38 rows for ONE viewing session!
```

The system inserts a new row every 10 seconds instead of **updating** the existing session.

## Proposed Solution: Simplified 2-Tier Architecture

### Current Architecture (3 Tiers)
```
Tier 1: video_views (RAW EVENTS)
  ❌ Problem: Infinite rows per session
  Purpose: Detailed analytics
  Storage: 1 row every 10 seconds

Tier 2: watch_history (USER STATE) 
  ✅ Works: One row per user+video
  Purpose: Resume playback
  Storage: Fixed size

Tier 3: master_video_list.views (AGGREGATED)
  ✅ Works: One number per video
  Purpose: Total view count
  Storage: Minimal
```

### Optimized Architecture (2 Tiers)

```
Tier 1: watch_history (USER STATE + TRACKING)
  ✅ Solution: UPSERT on user+video
  Purpose: Resume + basic tracking
  Storage: Fixed (1 row per user+video)
  
  Schema:
  - user_id (or session_id for anon)
  - video_id
  - last_position
  - progress_percentage
  - total_watch_time (NEW)
  - view_count (NEW - how many times they watched it)
  - completed
  - first_watched_at (NEW)
  - last_watched_at
  
Tier 2: master_video_list (AGGREGATED)
  ✅ Keep: Simple counters
  Purpose: Global stats
  Storage: Minimal
  
  Schema (existing):
  - views (unique viewers)
  - total_watch_time
  - average_watch_time
```

## Implementation Changes

### 1. Update watch_history Schema

```sql
ALTER TABLE watch_history 
  ADD COLUMN total_watch_time INTEGER DEFAULT 0,
  ADD COLUMN view_count INTEGER DEFAULT 1,
  ADD COLUMN first_watched_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  DROP CONSTRAINT watch_history_user_id_fkey,  -- Remove NOT NULL constraint
  ADD COLUMN session_id VARCHAR(255);  -- Allow anonymous tracking

-- Add unique constraint for anonymous users too
CREATE UNIQUE INDEX idx_watch_history_user_video ON watch_history(user_id, video_id) WHERE user_id IS NOT NULL;
CREATE UNIQUE INDEX idx_watch_history_session_video ON watch_history(session_id, video_id) WHERE session_id IS NOT NULL;
```

### 2. Change RecordView to UPSERT

```go
func (s *VideoAnalyticsService) RecordView(req VideoTrackingRequest) error {
    query := `
        INSERT INTO watch_history (
            user_id, session_id, video_id, 
            last_position, progress_percentage, 
            total_watch_time, view_count,
            first_watched_at, last_watched_at
        ) VALUES ($1, $2, $3, $4, $5, $6, 1, NOW(), NOW())
        ON CONFLICT (user_id, video_id) 
        WHERE user_id IS NOT NULL
        DO UPDATE SET
            last_position = EXCLUDED.last_position,
            progress_percentage = EXCLUDED.progress_percentage,
            total_watch_time = watch_history.total_watch_time + (EXCLUDED.last_position - watch_history.last_position),
            last_watched_at = NOW()
    `
    // Execute once per session update
}
```

### 3. Increment master_video_list.views Only Once

```sql
-- On FIRST view only (when watch_history is created)
UPDATE master_video_list 
SET views = views + 1,
    total_watch_time = total_watch_time + watch_time_delta
WHERE id = $video_id;
```

## Metrics We Can Still Track

### Per-Video Metrics (from watch_history aggregation)
- ✅ Total unique viewers: `COUNT(DISTINCT user_id or session_id)`
- ✅ Average completion: `AVG(progress_percentage)`
- ✅ Total watch time: `SUM(total_watch_time)`
- ✅ Completion rate: `COUNT(*) WHERE completed = true / COUNT(*)`
- ✅ Repeat viewers: `COUNT(*) WHERE view_count > 1`

### Per-User Metrics (from watch_history)
- ✅ Videos watched: `COUNT(*)`
- ✅ Total watch time: `SUM(total_watch_time)`
- ✅ Completion rate: Personal stats
- ✅ Continue watching: Where `completed = false`

### What We LOSE (and don't need)
- ❌ Second-by-second tracking (not useful for most analytics)
- ❌ Exact timestamps of every 10-second interval
- ❌ Play/pause event timeline (too granular)

## Storage Comparison

### Current System (video_views)
```
1 viewer watching 1 hour video:
= 360 rows (1 every 10 seconds)

1000 users, each watching 10 videos:
= 360 × 1000 × 10
= 3.6 MILLION rows
```

### Optimized System (watch_history UPSERT)
```
1 viewer watching 1 hour video:
= 1 row (updated in place)

1000 users, each watching 10 videos:
= 1000 × 10
= 10,000 rows (360x reduction!)
```

## Migration Plan

### Phase 1: Stop Creating New video_views (Immediate)
1. Update `RecordView()` to use watch_history UPSERT
2. Deploy backend
3. Stop the bleeding

### Phase 2: Migrate Existing Data (Next)
1. Aggregate existing video_views into watch_history
2. Update master_video_list.views from aggregated data
3. Verify data integrity

### Phase 3: Archive video_views (Final)
1. Rename `video_views` → `video_views_archive`
2. Keep for historical reference if needed
3. Or delete if storage is critical

## Alternative: Event Sampling

If we REALLY want raw events for detailed analytics:

### Option A: Sample Events (1% of updates)
```go
if rand.Float64() < 0.01 {  // 1% sampling
    logDetailedEvent()  // Optional detailed logging
}
// Always update watch_history
updateWatchHistory()
```

### Option B: Periodic Snapshots
```go
// Only log significant events
if percentageChange >= 25 {  // 0%, 25%, 50%, 75%, 100%
    logMilestone()
}
```

## Recommendation

**Implement Optimized 2-Tier System IMMEDIATELY**

Reasons:
1. ✅ **Stops database explosion** (top priority)
2. ✅ **Still provides all meaningful analytics**
3. ✅ **Matches original BRAID design** (watch_history was meant for this)
4. ✅ **Simpler to maintain**
5. ✅ **Better performance**

The current `video_views` approach is unsustainable and will cause production issues within days.

---

**Next Steps**:
1. Approve this optimization
2. Update watch_history schema
3. Modify RecordView() to UPSERT
4. Test with sample data
5. Deploy and monitor

**Priority**: 🔴 **CRITICAL** - Should be fixed before production launch

