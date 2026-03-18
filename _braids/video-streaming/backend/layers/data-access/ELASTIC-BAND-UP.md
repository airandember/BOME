# Video Streaming - Data Access Layer (Elastic Band Up)

## Responsibility

Maps database operations for videos, master videos, and CDN metadata to Go models. Sits between persistence and business logic.

## Files (Production)

- `backend/internal/database/video.go` - Video model, CRUD
- `backend/internal/database/master_video.go` - Master video model
- `backend/video-streaming/models/video.go` - Video domain model
- `backend/video-streaming/models/master_video.go` - Master video model

## Key Operations

- Create/Read/Update/Delete videos
- Query videos by category, status, access level
- Bunny.net ID lookups
- Master video sync state

## Dependencies

- Persistence: videos, master_videos, categories tables
- Migrations: `backend/migrations/*video*.sql`

## Used By

- Business logic: `backend/video-streaming/handlers/video.go`, `master_video_routes.go`
- Services: `backend/video-streaming/services/bunny_optimized.go`, `master_video_sync.go`

## Notes

Videos table stores Bunny.net CDN IDs. Master video sync tracks upstream/downstream video state.
