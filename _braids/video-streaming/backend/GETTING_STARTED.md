# Getting Started with Video Streaming Braid

**Quick guide to video management and Bunny.net CDN integration**

---

## What is This?

The Video Streaming Braid handles video upload, metadata, CDN delivery via Bunny.net, and playback. It is critical business functionality.

---

## Quick Start

### Key Production Files

- **Handlers**: `backend/video-streaming/handlers/video.go`, `master_video_routes.go`
- **Services**: `backend/video-streaming/services/bunny.go`, `bunny_optimized.go`
- **Models**: `backend/video-streaming/models/video.go`, `master_video.go`
- **Database**: `backend/internal/database/video.go`, `master_video.go`
- **Routes**: `backend/internal/routes/video.go`, `master_video_routes.go`

### Environment Variables

- `BUNNY_API_KEY` - Bunny.net API key
- `BUNNY_LIBRARY_ID` - Video library ID
- `BUNNY_STREAM_URL` - CDN stream base URL

### Common Scenarios

**"Video upload failing"**
1. Check `backend/video-streaming/services/bunny.go`
2. Verify BUNNY_* env vars
3. Check Bunny.net API quotas

**"Video playback not working"**
1. Check stream URL generation in video handlers
2. Verify CDN URLs in `backend/internal/database/video.go`
3. Check access control (subscription requirements)

**"Need to sync video metadata"**
1. See `backend/internal/services/master_video_sync.go`
2. Check `backend/braids/video-analytics/` for view sync

---

## Dependencies

- **Bunny.net**: CDN and stream library
- **Subscription Braid**: Access control for premium content
- **Content Management**: Tags, categories

---

## Documentation

- **BRAID.md**: Full braid overview at `_braids/video-streaming/BRAID.md`
- **Video Analytics**: `backend/braids/video-analytics/` for view count sync

---

**Last Updated**: 2025-03-17
