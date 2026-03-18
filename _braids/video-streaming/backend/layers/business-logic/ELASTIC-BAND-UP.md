# Video Streaming - Business Logic Layer (Elastic Band Up)

## Responsibility

Video upload, CDN integration (Bunny.net), playback metadata, and video lifecycle. Sits between data access and application (API) layer.

## Files (Production)

- `backend/video-streaming/handlers/video.go` - Video CRUD endpoints
- `backend/video-streaming/handlers/master_video_routes.go` - Master video sync
- `backend/video-streaming/services/bunny.go` - Bunny.net API
- `backend/video-streaming/services/bunny_optimized.go` - Optimized CDN ops
- `backend/video-streaming/services/master_video_sync.go` - Video sync logic
- `backend/internal/routes/video.go` - Video API routes
- `backend/internal/services/master_video_sync.go` - Sync service (internal)

## Key Operations

- Upload video to Bunny.net
- Generate stream URLs
- Master video sync (upstream/downstream)
- Video status workflow (draft → processing → published)
- Thumbnail and metadata handling

## Dependencies

- Data Access: Video models
- Infrastructure: Bunny.net API keys (BUNNY_API_KEY, BUNNY_LIBRARY_ID)
- Subscription braid: Access level checks

## Used By

- Application layer: `/api/v1/videos/*`, `/api/v1/master-videos/*`
- Frontend: Video player, upload UI

## Notes

Bunny.net integration is critical. Ensure BUNNY_* env vars are set for production.
