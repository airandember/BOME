# Bunny.net Views Sync Strand

## Purpose
To provide an admin-accessible mechanism for synchronizing video view counts from Bunny.net's API to the local `master_video_list` database table. This ensures accurate view count data for analytics, trending calculations, and display purposes.

## Background
A bug in migration 068 created a PostgreSQL trigger that **replaced** historical view counts with the sum of `watch_history.view_count`. Since `watch_history` only contains recent analytics data, this destroyed accurate historical view counts from Bunny.net. This strand provides a manual sync button to restore and maintain accurate view counts.

## Implementation Details

### Backend

#### Files
- `backend/internal/services/master_video_sync.go` - Contains `SyncViewsFromBunny()` method
- `backend/internal/routes/master_video_routes.go` - Exposes `POST /admin/master-videos/sync/views-from-bunny`
- `backend/cmd/restore_bunny_views/main.go` - Standalone CLI script for initial restoration

#### Service Method
```go
func (s *MasterVideoSyncService) SyncViewsFromBunny() (*ViewsSyncResult, error)
```

**Logic:**
1. Fetches all videos from Bunny.net API via `bunnyService.FetchAllVideos()`
2. Creates a map: `bunny_video_id` → `views`
3. Queries all videos from `master_video_list` with a `bunny_video_id`
4. For each video, compares current `views` with Bunny.net `views`
5. Updates only if different
6. Returns summary with `updated`, `skipped`, `not_found`, `errors`
7. Includes top 10 videos by views for verification

#### API Endpoint
```
POST /api/v1/admin/master-videos/sync/views-from-bunny
```

**Response:**
```json
{
  "success": true,
  "message": "Views sync completed: 313 updated, 1485 skipped",
  "result": {
    "started_at": "2025-12-06T08:14:59Z",
    "completed_at": "2025-12-06T08:14:59Z",
    "duration": "241.6369ms",
    "total_videos": 1800,
    "updated": 313,
    "skipped": 1485,
    "not_found": 2,
    "errors": 0,
    "top_videos": [
      {"id": 11338, "title": "Hannah Stoddard - Prophets Gone Missing", "views": 125}
    ]
  }
}
```

### Frontend

#### Files
- `frontend/src/lib/master-video.ts` - Contains `syncViewsFromBunny()` service method
- `frontend/src/lib/components/VideoStats.svelte` - UI button component
- `frontend/src/routes/admin/streaming/videos/+page.svelte` - Event handler

#### Components
- **VideoStats.svelte**: Added "Sync Views 📊" button with loading state animation
- **Event**: Dispatches `syncViewsFromBunny` event to parent

#### Service Method
```typescript
async syncViewsFromBunny(): Promise<{
  success: boolean;
  message: string;
  result: ViewsSyncResult;
}>
```

### Database
- **Table**: `master_video_list`
- **Column**: `views` (integer)
- **Source of Truth**: Bunny.net API for historical views, `watch_history` trigger for new views

## Flow
1. Admin navigates to `/admin/streaming/videos`
2. Admin clicks "Sync Views 📊" button in VideoStats component
3. Frontend calls `masterVideoService.syncViewsFromBunny()`
4. Backend fetches all videos from Bunny.net API (~1800 videos, ~25 seconds)
5. Backend updates `master_video_list.views` where different
6. Response shows summary with updated/skipped counts
7. Frontend displays toast notification and logs top 10 videos to console

## Status
- [x] Backend Service Method
- [x] Backend API Route
- [x] Frontend Service Method
- [x] Frontend UI Button
- [x] Event Handling
- [x] Documentation

## Testing
1. Start frontend and backend servers
2. Navigate to `/admin/streaming/videos`
3. Click "Sync Views 📊" button
4. Verify toast shows "Views sync complete: X updated, Y unchanged"
5. Check browser console for top 10 videos by views
6. Verify `master_video_list.views` in database matches Bunny.net values

## CLI Alternative
For initial restoration or production environments without admin access:
```bash
cd backend
go run cmd/restore_bunny_views/main.go
```

Required environment variables:
- `BUNNY_STREAM_LIBRARY_ID`
- `BUNNY_STREAM_API_KEY`
- `DATABASE_URL` or individual `DB_*` variables

## Related Migrations
- `068_sync_master_video_views_from_watch_history.sql` - Original buggy trigger
- `069_fix_views_sync_trigger.sql` - Fixed trigger (increments instead of replaces)
- `070_restore_bunny_views.sql` - Deprecated placeholder
- `071_restore_bunny_views_correct.sql` - Documentation migration

## Known Issues
- Sync takes ~25 seconds for ~1800 videos (rate-limited API calls)
- Videos deleted from Bunny.net will show as "not found" (expected)

## Notes
This strand complements the existing `video-player-analytics` strand. While that strand handles tracking new views via Player.js, this strand ensures the base view count data is accurate from Bunny.net's historical records.

