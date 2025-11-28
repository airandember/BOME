# Fix Trending Video Thumbnails 🖼️

## Problem
Trending section not loading thumbnails because query selects wrong column name.

## Root Cause
The `GetTrendingVideos` query selects `v.thumbnail_url` from `master_video_list`, but the actual column might be named differently (e.g., `thumbnail_path`, `poster_url`, etc.).

---

## Quick Fix

### Step 1: Check Column Name
```sql
-- What's the actual column name?
SELECT column_name 
FROM information_schema.columns 
WHERE table_name = 'master_video_list' 
  AND column_name LIKE '%thumb%' OR column_name LIKE '%poster%' OR column_name LIKE '%image%';
```

Common possibilities:
- `thumbnail_url` ✅ (if this, no fix needed!)
- `thumbnail_path`
- `poster_url`
- `poster_path`
- `image_url`

### Step 2: Check Sample Data
```sql
SELECT 
    id, 
    title,
    thumbnail_url,  -- Try this first
    poster_url,     -- Or this
    poster_path     -- Or this
FROM master_video_list 
LIMIT 3;
```

### Step 3: Update Query (if needed)
If the column is named differently, update line 296 in `video_analytics_service.go`:

```go
// FROM:
v.thumbnail_url,

// TO (example if it's poster_url):
v.poster_url AS thumbnail_url,
```

---

## Alternative: Use Bunny Thumbnail URL

Since you're syncing from Bunny anyway, construct the thumbnail URL:

```sql
-- Instead of v.thumbnail_url, use:
'https://vz-f75053f7-465.b-cdn.net/' || v.bunny_video_id || '/thumbnail.jpg' AS thumbnail_url
```

This works if videos are stored in Bunny.net!

---

## Test Query
```sql
-- Test the trending query with thumbnail
SELECT 
    v.id,
    v.title,
    v.bunny_video_id,
    'https://vz-f75053f7-465.b-cdn.net/' || v.bunny_video_id || '/thumbnail.jpg' AS constructed_thumbnail,
    v.views
FROM master_video_list v
WHERE v.status = 'ready' AND v.views > 0
ORDER BY v.views DESC
LIMIT 5;
```

Check if `constructed_thumbnail` URLs work!

---

## Quick Deploy

### Option A: If Column Name is Different
```go
// In video_analytics_service.go line 296
// Change from:
v.thumbnail_url,

// To (example):
v.poster_url AS thumbnail_url,  -- Or whatever the real column is
```

### Option B: Use Bunny CDN (Recommended for now)
```go
// In video_analytics_service.go line 296
// Change from:
v.thumbnail_url,

// To:
CONCAT('https://vz-f75053f7-465.b-cdn.net/', v.bunny_video_id, '/thumbnail.jpg') AS thumbnail_url,
```

Then restart backend:
```bash
go run main.go
```

---

## Summary

**Problem**: Thumbnails not loading in trending section
**Likely Cause**: Column name mismatch or empty data
**Quick Fix**: Use Bunny CDN URL construction
**Better Fix**: Check actual column name and update query

Let me know what the column is called and I'll give you the exact fix! 🎯

