# Most Watched Videos Feature

## Overview
Added a "Most Watched" button to the Trending tab that displays the top 25 most viewed videos from the last 30 days.

## Implementation

### Backend (Already Exists)
- **Endpoint**: `GET /api/v1/analytics/top`
- **Query Parameters**:
  - `limit` (default: 10, max: 100)
  - `days` (default: 30)
- **Service**: `VideoAnalyticsService.GetTopVideos()`
- **Returns**: Top videos with metrics including:
  - Total views
  - Unique viewers
  - Average completion rate
  - Total watch time
  - Video duration

### Frontend Updates

#### 1. Video Analytics Service (`frontend/src/lib/services/videoAnalytics.ts`)
Added new method:
```typescript
async getTopVideos(limit: number = 25, days: number = 30): Promise<any[]>
```
- Calls `/api/v1/analytics/top` endpoint
- Returns top videos by total view count
- Public endpoint (no auth required)

#### 2. TrendingVideos Component (`frontend/src/lib/components/TrendingVideos.svelte`)
Enhanced to support both "Trending" and "Most Watched" view modes:

**New Features**:
- View mode toggle buttons (Trending 🔥 / Most Watched 📊)
- Dynamic title and subtitle based on view mode
- Different data display for each mode:
  - **Trending Mode**: Shows trending score, 24h views, trending badges
  - **Most Watched Mode**: Shows total views, unique viewers, completion rate, video duration
- Unified video grid layout for both modes
- Top 25 videos when in "Most Watched" mode

**Component Props** (unchanged):
- `limit`: Number of videos to display (default: 10 for trending, 25 for most watched)
- `showTitle`: Whether to show section header (default: true)
- `autoRefresh`: Enable auto-refresh (default: false)
- `refreshInterval`: Refresh interval in ms (default: 60000)

**State Management**:
- `viewMode`: 'trending' | 'most_watched'
- `timePeriod`: 'week' | 'month' | 'all-time' (for Most Watched mode)
- Switches data loading logic based on view mode and time period
- Maintains separate display logic for each mode
- Time period selection dynamically updates the days parameter in API call

## Visual Design

### Toggle Buttons
- Pill-style toggle with two options (Trending / Most Watched)
- Active state highlighted with white background and brand color text
- Smooth transitions between states
- Located below the section header

### Time Period Selector (Most Watched Only)
- Three options: **This Week** | **This Month** | **All-Time**
- Appears with slide-down animation when Most Watched is active
- Border-style buttons with hover effects
- Active period has filled background with brand color
- Time periods:
  - **This Week**: Last 7 days
  - **This Month**: Last 30 days (default)
  - **All-Time**: All available history (100 years lookback)

### Video Cards (Most Watched Mode)
- **Rank Badge**: Position number (#1-#25) with gold/silver/bronze for top 3
- **Duration Badge**: Video length displayed in bottom-left
- **Views Badge**: Total views displayed in bottom-right (e.g., "1.2K total")
- **Stats Row**: Shows unique viewers and completion percentage below title

### Video Cards (Trending Mode)
- **Rank Badge**: Position number with special styling for top 3
- **Trending Badge**: "ON FIRE", "HOT", "TRENDING", or "RISING" based on score
- **Views Badge**: 24h views (e.g., "1.2K (24h)")
- **Score Indicator**: Progress bar showing trending score out of 100

## Usage

The component is used in the videos page trending tab:

```svelte
<TrendingVideos 
  limit={20} 
  showTitle={true} 
  autoRefresh={true} 
  refreshInterval={60000} 
/>
```

Users can now:
1. View trending videos (what's hot right now)
2. Click "Most Watched" to see the top 25 most viewed videos
3. Select time period: This Week, This Month, or All-Time
4. Switch between modes and time periods seamlessly
5. See relevant metrics for each view mode

## Technical Notes

### SSR Compatibility
The VideoAnalytics service now includes browser environment checks:
- Checks for `window`, `sessionStorage`, and `localStorage` before accessing them
- Returns fallback values during SSR
- Ensures the component can be server-side rendered without errors

### Data Loading
- Separate API calls for trending vs most watched
- Loading states maintained during view switches
- Error handling for both modes
- Auto-refresh respects current view mode

### Responsive Design
- Grid layout adapts to screen size
- Toggle buttons stack on mobile
- Cards resize appropriately
- All badges and overlays remain visible on small screens

## Future Enhancements

Potential improvements:
- Add category filter for both modes
- Add "Save to Playlist" quick action on cards
- Add share functionality
- Add sorting options (views, completion rate, recency)
- Add pagination for Most Watched (beyond top 25)

## Testing Checklist

- [x] Frontend service method added
- [x] Component view mode toggle implemented
- [x] Different display logic for each mode
- [x] SSR compatibility fixed
- [x] No TypeScript errors
- [x] No linting errors
- [x] Responsive design maintained
- [ ] Manual testing: Toggle between modes
- [ ] Manual testing: Verify data loads correctly
- [ ] Manual testing: Check mobile responsive
- [ ] Manual testing: Verify auto-refresh works in both modes

