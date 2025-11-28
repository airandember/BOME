# Trending Section: Top 100 Videos Update ✅

## Changes Made

### Backend
1. **`video_analytics_service.go`**:
   - **GetTrendingVideos**: Limit increased from 50 → 100
   - **GetTopVideos**: Default changed from 10 → 100
   - Both now support up to 100 videos

### Frontend
2. **`TrendingVideos.svelte`**:
   - Default `limit` prop increased from 10 → 100
   - Will now fetch and display top 100 trending videos

---

## Thumbnail Fix

### Status: Already Correct! ✅
The query already uses `v.thumbnail_url` which matches the `master_video_list` schema (same as GetVideos).

**If thumbnails still don't show**, it's likely because:
1. Data is empty in database
2. URL is broken/invalid
3. CORS issues

### Quick Check
```sql
-- Check if thumbnail_url has data
SELECT 
    id,
    title,
    thumbnail_url,
    views
FROM master_video_list
WHERE status = 'ready'
ORDER BY views DESC
LIMIT 5;
```

If `thumbnail_url` is NULL, you need to sync from Bunny!

---

## Deploy

### Restart Backend
```bash
cd S:\AirEmber\BOME\BOME\backend
go run main.go
```

### Hard Refresh Frontend
```
Ctrl+Shift+R
```

---

## Expected Results

### Trending Tab
- Shows up to **100 videos** (was 10)
- Sorted by trending score (24h activity + engagement)
- Thumbnails from `master_video_list.thumbnail_url`

### Most Watched Tab
- Shows **top 25 videos** (as defined in component line 74)
- Time period filters: Week / Month / All-Time
- Sorted by total view count

---

## Testing

### Test Trending
1. Go to `/videos` page
2. Click "Trending" tab
3. Should show up to 100 videos with thumbnails
4. Videos sorted by recent activity

### Test Most Watched
1. Click "Most Watched" button
2. Should show 25 videos
3. Try time period filters:
   - This Week
   - This Month
   - All-Time

---

## If Thumbnails Still Don't Show

### Option 1: Check Database
```sql
-- Are thumbnails populated?
SELECT COUNT(*) FROM master_video_list WHERE thumbnail_url IS NOT NULL;
SELECT COUNT(*) FROM master_video_list WHERE thumbnail_url IS NULL;
```

### Option 2: Sync from Bunny
If thumbnails are NULL, update them from Bunny.net:
```go
// When syncing videos from Bunny
updates["thumbnail_url"] = bunnyService.GetThumbnailURL(bunnyVideo.GUID)
```

### Option 3: Use CDN Fallback
Update query to construct URL if NULL:
```sql
COALESCE(
    v.thumbnail_url,
    CONCAT('https://vz-f75053f7-465.b-cdn.net/', v.bunny_video_id, '/thumbnail.jpg')
) AS thumbnail_url
```

---

## Summary

**Changed**:
- ✅ Backend supports up to 100 trending videos
- ✅ Frontend requests 100 by default
- ✅ Thumbnail query matches GetVideos (already correct)

**Status**: Ready to test! 🚀

**Next**: 
1. Restart backend
2. Hard refresh browser
3. Check trending section for 100 videos with thumbnails

