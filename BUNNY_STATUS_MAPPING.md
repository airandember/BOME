# Bunny.net Video Status Mapping

## Overview
This document explains how Bunny.net video status codes are mapped to our internal status system and how they're displayed in the UI.

## Bunny.net Status Codes

| Code | Bunny.net Status | Our Status | Description | UI Color | UI Label |
|------|------------------|------------|-------------|----------|----------|
| 0 | Created | `created` | Video record created but not uploaded | Blue | Created |
| 1 | Uploaded | `uploaded` | Video file uploaded, waiting for processing | Blue | Uploaded |
| 2 | Processing | `processing` | Video is being processed | Orange | Processing |
| 3 | Transcoding | `transcoding` | Video is being transcoded to different formats | Orange | Transcoding |
| 4 | Finished | `ready` | ✅ Video is ready for playback | Green | Ready |
| 5 | Error | `error` | ❌ Error occurred during processing | Red | Error |
| 6 | UploadFailed | `upload_failed` | ❌ Upload failed | Red | Upload Failed |
| 7 | JitSegmenting | `jit_segmenting` | Just-in-time segmenting in progress | Orange | Segmenting |
| 8 | JitPlaylistsCreated | `jit_playlists_created` | JIT playlists created, finalizing | Purple | Finalizing |

## Status Flow

```
Created (0) → Uploaded (1) → Processing (2) → Transcoding (3) → Finished (4)
                    ↓
              UploadFailed (6)
                    ↓
                 Error (5)
```

### Alternative Flow (JIT Processing)
```
Created (0) → Uploaded (1) → JitSegmenting (7) → JitPlaylistsCreated (8) → Finished (4)
```

## Implementation Details

### Backend Mapping
```go
func mapBunnyStatus(status int) string {
    switch status {
    case 0: return "created"
    case 1: return "uploaded" 
    case 2: return "processing"
    case 3: return "transcoding"
    case 4: return "ready"        // ✅ Ready for playback
    case 5: return "error"
    case 6: return "upload_failed"
    case 7: return "jit_segmenting"
    case 8: return "jit_playlists_created"
    default: return "unknown"
    }
}
```

### Frontend Status Display
- **Green (Ready)**: Video is fully processed and ready for streaming
- **Blue (Created/Uploaded)**: Initial states, video being prepared
- **Orange (Processing/Transcoding/Segmenting)**: Video being processed
- **Purple (Finalizing)**: JIT playlists created, almost ready
- **Red (Error/Failed)**: Something went wrong

### UI Behavior
- Only videos with status `!== 'ready'` show status badges
- Status badges appear in the top-right corner of video thumbnails
- Color coding provides quick visual feedback on video state

## Key Fix
**Previously**: Status 4 was incorrectly mapped to "ready" at case 3, causing finished videos to never show as ready.

**Now**: Status 4 correctly maps to "ready", so finished videos will display without status badges and be fully playable.

## Monitoring
Videos stuck in processing states for extended periods may indicate:
- Bunny.net processing issues
- Large file sizes requiring extended processing time
- Network connectivity problems
- Invalid video formats

## Testing
To verify status mapping:
1. Upload a video to Bunny.net
2. Check the API response for status codes
3. Verify the UI displays the correct status badge
4. Confirm videos with status 4 show no badge (ready state) 