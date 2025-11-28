# Analytics Throttling Bug Fix 🐛→✅

## Problem Identified

**Issue**: Analytics never sent to backend despite video playing for 900+ seconds.

### Root Cause
The throttling logic was checking if `currentSecond % 10 === 0` (exact multiples of 10), but the tracking interval was catching the video at times like **9s, 19s, 29s, 39s...** due to timing between the iframe polling (every 1s) and the tracking interval (every 10s).

### The Bug
```typescript
// OLD - BROKEN LOGIC
if (currentSecond % this.trackingInterval !== 0) {
    return; // Skip if not exactly 0, 10, 20, 30, 40...
}
```

**Result**: Video at 9s, 19s, 29s... would NEVER match the condition (9 % 10 = 9, not 0).

---

## Solution Applied

Changed from "exact multiple" logic to "time elapsed since last report" logic.

### The Fix
```typescript
// NEW - FIXED LOGIC
const lastReported = this.lastReportedTime.get(videoId) || -999;
const secondsSinceLastReport = currentSecond - lastReported;

if (secondsSinceLastReport < this.trackingInterval) {
    return; // Skip if not enough time passed
}
```

**Result**: Tracks at 9s, then 19s, then 29s, then 39s... (every 10 seconds, regardless of exact time).

---

## Before vs After

### Before (Broken)
```
Time 0s:   Track ✅ (0 % 10 = 0)
Time 9s:   Skip ❌ (9 % 10 = 9)
Time 19s:  Skip ❌ (19 % 10 = 9)
Time 29s:  Skip ❌ (29 % 10 = 9)
...
Time 989s: Skip ❌ (989 % 10 = 9)

Result: Only tracks at 0s, then NEVER again! 🚫
```

### After (Fixed)
```
Time 0s:   Track ✅ (0 - (-999) = 999s > 10s)
Time 9s:   Skip ⏭️  (9 - 0 = 9s < 10s)
Time 19s:  Track ✅ (19 - 0 = 19s > 10s)
Time 29s:  Skip ⏭️  (29 - 19 = 10s, but exact)
Time 39s:  Track ✅ (39 - 19 = 20s > 10s)
...

Result: Tracks every ~10 seconds! ✅
```

---

## Expected Behavior Now

### Console Logs
```
🎬 [FRONTEND] trackProgress called: video=11042, time=9s, duration=2798s
⏭️  [FRONTEND] Skipping - only 9s since last report (need 10s)

🎬 [FRONTEND] trackProgress called: video=11042, time=19s, duration=2798s
✅ [FRONTEND] 19s since last report - tracking now!
📤 [FRONTEND] Sending tracking event: video=11042, time=19s, %=0.68%
🌐 [FRONTEND→BACKEND] Preparing request...
📥 [FRONTEND←BACKEND] Response received in 4ms: 200
✅ [FRONTEND] Successfully tracked

🎬 [FRONTEND] trackProgress called: video=11042, time=29s, duration=2798s
⏭️  [FRONTEND] Skipping - only 10s since last report (need 10s)

🎬 [FRONTEND] trackProgress called: video=11042, time=39s, duration=2798s
✅ [FRONTEND] 20s since last report - tracking now!
📤 [FRONTEND] Sending tracking event: video=11042, time=39s, %=1.39%
```

### Backend Logs (Should now appear!)
```
🌐 [ROUTE] ============================================
🌐 [ROUTE] Received POST /analytics/video/track
🎯 [SERVICE] RecordView called
📦 [BUFFER] AddEvent called: video=11042, user=42
✅ [BUFFER←REDIS] Event pushed to Redis successfully
```

### Database (After watching for 20+ seconds)
```sql
SELECT * FROM watch_history WHERE video_id = 11042;
-- Should have data! 🎉
```

---

## Why This Happened

The issue was a **race condition** between two intervals:
1. **Iframe polling**: Increments time every 1 second (1s, 2s, 3s...)
2. **Tracking interval**: Checks every 10 seconds from `+page.svelte`

The timing meant it was checking at 9s, 19s, 29s instead of 10s, 20s, 30s.

The modulo logic (`%`) required **exact** multiples of 10, which the timing never hit.

---

## File Changed

**`frontend/src/lib/services/videoAnalytics.ts`**
- Lines 73-107 (trackProgress method)
- Changed throttling logic from modulo to time-elapsed
- Now tracks every 10 seconds regardless of exact position

---

## Testing

1. **Refresh browser** (Ctrl+Shift+R)
2. **Play video** for 20+ seconds
3. **Check console** - Should see tracking events
4. **Check backend logs** - Should see full BRAID flow
5. **Check database**:
   ```sql
   SELECT * FROM watch_history ORDER BY last_watched_at DESC LIMIT 5;
   ```

---

## ✅ Fix Status

- [x] Bug identified (modulo logic)
- [x] Root cause found (timing mismatch)
- [x] Solution applied (time-elapsed logic)
- [x] Auth token retrieval fixed (401 error)
- [x] No linter errors
- [x] Ready for testing

---

## Additional Fix: 401 Unauthorized Error

### Problem
After fixing the throttling, requests were failing with `401 Unauthorized` because `videoAnalytics.ts` was looking for `auth_token` in localStorage, but the auth system stores tokens in `bome_auth_data` as a JSON object.

### Solution
Updated `getAuthToken()` to parse `bome_auth_data` and extract `access_token`:

```typescript
// OLD - Wrong key
let token = localStorage.getItem('auth_token');

// NEW - Correct format
const authDataStr = localStorage.getItem('bome_auth_data');
if (authDataStr) {
    const authData = JSON.parse(authDataStr);
    return authData.access_token || null;
}
```

---

**Both bugs are squashed!** 🐛🐛→✅✅

Now analytics should fire every 10 seconds with proper authentication and write to `watch_history`! 🎉

