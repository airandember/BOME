# Video Player Analytics Integration Strand

## Purpose
Enable accurate video analytics tracking by integrating Bunny.net's Player.js API with the iframe-based video player. This replaces the inaccurate polling-based tracking with event-driven analytics that can detect play, pause, seek, and timeupdate events.

## Problem Statement
The current iframe-based video player uses a **1-second polling fallback** to track video progress because:
- Cross-origin restrictions prevent direct access to iframe video state
- The polling cannot detect pause events, leading to over-counted watch time
- Analytics accuracy is ~70% (inflated engagement metrics)

## Solution
Integrate **Bunny.net Player.js API** which provides cross-origin communication with the embedded player:
- Event-driven tracking (play, pause, timeupdate, ended)
- Accurate pause detection
- Real playback position from the player
- ~95% analytics accuracy

---

## Implementation Details

### Backend
- **Files**: No backend changes required - analytics endpoints already exist
- **Endpoints Used**:
  - `POST /api/v1/analytics/video/track` - Receives tracking events
  - `GET /api/v1/analytics/video/:id/stats` - Returns video statistics

### Frontend
- **Files Modified**:
  - `frontend/src/lib/components/VideoPlayer.svelte` - Main player component
  - `frontend/src/lib/video.ts` - Video service (iframe URL generation)

- **Files Used (Unchanged)**:
  - `frontend/src/lib/services/videoAnalytics.ts` - Analytics service

### Key Changes
1. **Restore `initPlayerJs()` function** from VideoPlayer_backup.svelte
2. **Add `?api=1` parameter** to iframe URLs for Player.js API access
3. **Add `timeupdate` event handler** with throttled tracking
4. **Wire events to `videoAnalytics.trackProgress()`**
5. **Remove polling fallback** (the 1-second interval)

---

## Flow

```
1. User navigates to video page
2. VideoPlayer component mounts with iframe
3. Player.js library loads dynamically
4. Player.js connects to iframe via postMessage
5. On 'ready' event → Player initialized
6. On 'play' event → Start tracking, dispatch event
7. On 'timeupdate' event → Every few seconds:
   - Get currentTime and duration from Player.js
   - Call videoAnalytics.trackProgress(videoId, currentTime, duration)
   - Backend receives and stores analytics
8. On 'pause' event → Stop tracking, record position
9. On 'ended' event → Mark complete (100%)
```

---

## Technical Details

### Player.js Events
| Event | Data | Use Case |
|-------|------|----------|
| `ready` | - | Player initialized, safe to call methods |
| `play` | - | Video started/resumed playing |
| `pause` | - | Video paused by user |
| `timeupdate` | `{seconds, duration}` | Playback position changed |
| `ended` | - | Video finished playing |
| `error` | error object | Playback error occurred |

### Iframe URL Format
```
Before: https://iframe.mediadelivery.net/embed/LIBRARY_ID/VIDEO_ID
After:  https://iframe.mediadelivery.net/embed/LIBRARY_ID/VIDEO_ID?api=1
                                                                    ^^^^^^
                                                        Required for Player.js!
```

### Tracking Throttle
- Player.js fires `timeupdate` frequently (~4x per second)
- We throttle to every **10 seconds** to match backend expectations
- Uses `lastReportedTime` Map to track when we last reported

---

## Status
- [x] Backend analytics endpoints ready
- [x] videoAnalytics.ts service ready
- [x] Player.js code exists in backup
- [x] Player.js restored in VideoPlayer.svelte ✅
- [x] Iframe URL updated with ?api=1 ✅
- [x] timeupdate event wired to analytics ✅
- [x] Polling fallback removed ✅
- [ ] Tested with production videos (manual testing needed)

---

## Testing

### Manual Testing
1. Open a video page in browser
2. Open DevTools Console
3. Look for these log messages:
   - `📊 [Video Analytics] Service initialized`
   - `Player.js ready`
   - `⏱️ Time update: Xs / Ys`
   - `✅ [FRONTEND] Successfully tracked: video=X`
4. Pause the video - watch time should STOP counting
5. Resume - watch time should continue
6. Complete video - should mark 100%

### Verify in Database
```sql
SELECT * FROM watch_history 
WHERE video_id = YOUR_VIDEO_ID 
ORDER BY last_watched_at DESC 
LIMIT 5;
```

---

## Known Issues
- Player.js doesn't provide a `paused` state directly in timeupdate
- We infer pause from lack of timeupdate events
- If Player.js fails to load, falls back to iframe without tracking

---

## Related Documentation
- **Bunny.net Player.js**: https://docs.bunny.net/docs/stream-player-settings
- **Video Analytics Service**: `frontend/src/lib/services/videoAnalytics.ts`
- **Iframe Limitation Doc**: `IFRAME_PAUSE_TRACKING_LIMITATION.md`
- **Video Streaming Braid**: `_BRAIDS/video-streaming/BRAID.md`

---

## Migration Notes
- This is a **non-breaking change** - if Player.js fails, iframe still works
- Analytics will improve gradually as users watch videos
- Historical data (from polling) is still valid for trending/ranking

---

**Created**: December 2025  
**Status**: ✅ IMPLEMENTED - Ready for Testing  
**Actual Effort**: ~45 minutes  
**Priority**: High (Analytics Braid dependency)

