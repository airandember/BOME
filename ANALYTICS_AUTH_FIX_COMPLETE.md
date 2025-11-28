# Analytics Authentication Fix ✅

## Summary
Fixed **401 Unauthorized** error preventing video analytics from reaching the backend.

---

## Problem Identified

### Error
```
POST http://localhost:5173/api/v1/analytics/video/track 401 (Unauthorized)
{"error":"Authorization header required"}
```

### Root Cause
The `videoAnalytics.ts` service was looking for tokens in the **wrong localStorage key**:

```typescript
// WRONG - videoAnalytics.ts was using
let token = localStorage.getItem('auth_token');
```

But the **auth system** (`auth.ts`) stores tokens as:

```typescript
// CORRECT - auth.ts stores tokens as
localStorage.setItem('bome_auth_data', JSON.stringify({
    access_token: '...',
    refresh_token: '...',
    expires_in: 14400,
    expires_at: 1764262018305
}));
```

---

## Solution Applied

### File Changed
**`frontend/src/lib/services/videoAnalytics.ts`**

### Fix
Updated `getAuthToken()` method to correctly parse `bome_auth_data`:

```typescript
private getAuthToken(): string | null {
    // Check if we're in a browser environment
    if (typeof window === 'undefined' || typeof localStorage === 'undefined') {
        return null; // SSR - no auth token available
    }
    
    // Get tokens from BOME auth storage (matches auth.ts structure)
    try {
        const authDataStr = localStorage.getItem('bome_auth_data');
        if (authDataStr) {
            const authData = JSON.parse(authDataStr);
            return authData.access_token || null;
        }
    } catch (error) {
        console.error('Failed to parse auth data:', error);
    }
    
    return null;
}
```

---

## Expected Behavior After Fix

### Console Logs (Frontend)
```
🎬 [FRONTEND] trackProgress called: video=11042, time=9s, duration=2798s
✅ [FRONTEND] 10s since last report - tracking now!
📤 [FRONTEND] Sending tracking event: video=11042, time=9s, %=0.3%
🌐 [FRONTEND→BACKEND] Preparing request to /api/v1/analytics/video/track
🔑 [FRONTEND] Using JWT authentication  ← Now shows JWT!
📥 [FRONTEND←BACKEND] Response received in 4ms: 200  ← Now 200!
✅ [FRONTEND] Successfully tracked
```

### Backend Logs (Should now appear!)
```
🌐 [ROUTE] ============================================
🌐 [ROUTE] Received POST /analytics/video/track
🌐 [ROUTE] Tracking for user: 42, video: 11042
🌐 [ROUTE→SERVICE] Calling analyticsService.RecordView
🎯 [SERVICE] RecordView called for video 11042, user 42
🎯 [SERVICE→BUFFER] Adding event to Redis buffer
📦 [BUFFER] AddEvent called: video=11042, user=42
✅ [BUFFER←REDIS] Event pushed to Redis successfully
```

### Database (After ~5 seconds when buffer flushes)
```sql
SELECT * FROM watch_history WHERE video_id = 11042;

-- Should show data! 🎉
video_id | user_id | last_position | watch_percentage | total_watch_time | view_count
---------|---------|---------------|------------------|------------------|------------
11042    | 42      | 49            | 1.75             | 49               | 1
```

---

## Why This Happened

The `videoAnalytics.ts` service was originally written with a generic `auth_token` key assumption, but your authentication system uses a more structured approach with:

1. **Key**: `bome_auth_data` (not `auth_token`)
2. **Format**: JSON object with multiple fields (not just a string)
3. **Field**: `access_token` property within the object

This is actually a **better** pattern because it allows storing multiple auth-related fields together (access token, refresh token, expiration, etc.).

---

## Related Fixes

This fix completes the **Video Analytics BRAID** authentication integration:

1. ✅ **Throttling Fix**: Changed from modulo logic to time-elapsed logic
2. ✅ **Auth Integration**: Now correctly retrieves JWT from `bome_auth_data`
3. ✅ **BRAID Integrity**: Analytics service now properly integrates with auth BRAID

---

## Testing Instructions

1. **Hard refresh browser** (Ctrl+Shift+R or Cmd+Shift+R)
2. **Watch a video** for 20+ seconds
3. **Check console** - Should see:
   - `🔑 [FRONTEND] Using JWT authentication` (not session ID)
   - `📥 [FRONTEND←BACKEND] Response received in Xms: 200` (not 401)
   - `✅ [FRONTEND] Successfully tracked`

4. **Check backend logs** - Should see full BRAID flow:
   - Route receives request
   - Service processes event
   - Buffer adds to Redis

5. **Check database** after 10+ seconds:
   ```sql
   SELECT * FROM watch_history 
   WHERE user_id = 42  -- Your user ID
   ORDER BY last_watched_at DESC 
   LIMIT 5;
   ```

---

## Files Modified

1. **`frontend/src/lib/services/videoAnalytics.ts`**
   - Updated `getAuthToken()` method (lines 366-387)
   - Now correctly parses `bome_auth_data` JSON

---

## ✅ Status

- [x] Bug identified (wrong localStorage key)
- [x] Root cause found (auth system structure mismatch)
- [x] Solution applied (parse JSON from correct key)
- [x] No linter errors
- [x] Ready for testing

**The authentication bug is squashed!** 🐛→✅

Combined with the throttling fix, analytics should now work end-to-end! 🎉

