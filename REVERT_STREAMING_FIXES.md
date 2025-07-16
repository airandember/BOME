# Revert Guide for Streaming Fixes

If the streaming fixes cause issues, here's how to revert them:

## Backend Changes (routes.go)

**File**: `backend/internal/routes/routes.go`

**Lines to revert**:
1. **Line ~1003**: Change back to:
   ```go
   req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bunnyToken))
   ```

2. **Remove lines ~1005-1008**: Remove the new authentication attempt code

3. **Line ~1010**: Change back to:
   ```go
   if k != "Authorization" && // Skip Authorization as we set it above
       k != "Host" && // Skip Host as it will be set by the client
       k != "Connection" { // Skip Connection as it will be managed by the client
   ```

## Frontend Changes (VideoPlayer.svelte)

**File**: `frontend/src/lib/components/VideoPlayer.svelte`

**Template section**: Remove the `<div class="video-container">` wrapper

**CSS section**: Remove these styles:
```css
.video-container {
    position: relative;
    width: 100%;
    padding-bottom: 56.25%; /* 16:9 Aspect Ratio */
    height: 0;
}
```

**Video element CSS**: Change back to:
```css
.video-element {
    width: 100%;
    height: auto;
    display: block;
}
```

**Iframe element CSS**: Change back to:
```css
.iframe-element {
    width: 100%;
    height: 100%;
    min-height: 300px;
}
```

## Quick Revert Commands

If you need to revert quickly:

1. **Backend**: 
   ```bash
   git checkout backend/internal/routes/routes.go
   ```

2. **Frontend**:
   ```bash
   git checkout frontend/src/lib/components/VideoPlayer.svelte
   ```

3. **Restart backend**:
   ```bash
   cd backend && go run main.go
   ```

## Test the Changes

After applying the fixes:
1. Try loading a video page
2. Check browser console for errors
3. Check backend logs for streaming requests
4. If 403 errors persist, revert the backend changes
5. If video sizing is wrong, revert the frontend changes 