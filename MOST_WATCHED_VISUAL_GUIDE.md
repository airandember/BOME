# Most Watched Feature - Visual Guide

## Component Layout

```
┌─────────────────────────────────────────────────────────────────┐
│  🔥 Trending Now                          Updated 2m ago       │
│  Most watched in the last 24 hours                             │
│                                                                  │
│  ┌────────────────────────┐                                    │
│  │ [🔥 Trending] [📊 Most Watched] ← Toggle Buttons           │
│  └────────────────────────┘                                    │
└─────────────────────────────────────────────────────────────────┘
```

## View Mode: Trending (Default)

```
╔═══════════════════════════════════════╗
║  🔥 Trending Now         Updated 2m  ║
║  Most watched in the last 24 hours   ║
║                                       ║
║  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓   ║
║  ┃ 🔥 Trending  | Most Watched ┃   ║  ← Active: Trending
║  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛   ║
╚═══════════════════════════════════════╝

┌─────────────┬─────────────┬─────────────┐
│    [#1]     │    [#2]     │    [#3]     │  ← Rank badges (gold/silver/bronze)
│  ┌───────┐  │  ┌───────┐  │  ┌───────┐  │
│  │ Thumb │  │  │ Thumb │  │  │ Thumb │  │
│  │  🔥   │  │  │  🔥   │  │  │  HOT  │  │  ← Trending badge
│  │ 🔥    │  │  │ HOT   │  │  │       │  │
│  │ ON FIRE│  │  │       │  │  │       │  │
│  │👁️ 5.2K │  │  │👁️ 3.1K│  │  │👁️ 2.8K│  │  ← 24h views
│  └───────┘  │  └───────┘  │  └───────┘  │
│ Video Title │ Video Title │ Video Title │
│ ████░░ 85  │ ████░░ 78  │ ███░░░ 65  │  ← Trending score
└─────────────┴─────────────┴─────────────┘
```

## View Mode: Most Watched

```
╔═══════════════════════════════════════╗
║  📊 Most Watched         Updated 2m  ║
║  Top 25 most viewed this month       ║
║                                       ║
║  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓   ║
║  ┃ Trending  | 📊 Most Watched ┃   ║  ← Active: Most Watched
║  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛   ║
║                                       ║
║  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓   ║  ← Time Period Selector
║  ┃ This Week | This Month | All-Time ┃  ║
║  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛   ║
╚═══════════════════════════════════════╝

┌─────────────┬─────────────┬─────────────┐
│    [#1]     │    [#2]     │    [#3]     │  ← Rank badges
│  ┌───────┐  │  ┌───────┐  │  ┌───────┐  │
│  │ Thumb │  │  │ Thumb │  │  │ Thumb │  │
│  │       │  │  │       │  │  │       │  │
│  │       │  │  │       │  │  │       │  │
│  │25m  👁️│  │  │18m  👁️│  │  │22m  👁️│  │  ← Duration + Total views
│  │ 45.2K │  │  │ 38.1K │  │  │ 32.8K │  │
│  └───────┘  │  └───────┘  │  └───────┘  │
│ Video Title │ Video Title │ Video Title │
│ 👤 12.3K    │ 👤 10.5K    │ 👤 9.8K     │  ← Unique viewers
│ ✓ 78%       │ ✓ 82%       │ ✓ 75%       │  ← Completion rate
└─────────────┴─────────────┴─────────────┘
```

## Toggle Button States

### Inactive State
```
┌──────────────────────────────────┐
│  [  Trending  ] [  Most Watched  ]  │  ← Gray background, gray text
└──────────────────────────────────┘
```

### Active State (Trending)
```
┌──────────────────────────────────┐
│  [ 🔥 Trending ] [  Most Watched  ]  │  ← White bg, orange text | Gray
└──────────────────────────────────┘
    ^^^^^^^^^^^^
    Active: white background
    Orange text (#ff6b35)
    Subtle shadow
```

### Active State (Most Watched)
```
┌──────────────────────────────────┐
│  [  Trending  ] [ 📊 Most Watched ]  │  ← Gray | White bg, orange text
└──────────────────────────────────┘
                    ^^^^^^^^^^^^^^^
                    Active: white background
                    Orange text (#ff6b35)
                    Subtle shadow
```

## Time Period Buttons (Most Watched Only)

### Inactive State
```
┌───────────────────────────────────────────┐
│  [ This Week ] [ This Month ] [ All-Time ]  │  ← White bg, gray border & text
└───────────────────────────────────────────┘
```

### Hover State
```
┌───────────────────────────────────────────┐
│  [ This Week ] [ This Month ] [ All-Time ]  │
└───────────────────────────────────────────┘
     ^^^^^^^^^
     Hover: Orange border
     Light orange background (#fff5f2)
     Orange text
```

### Active State
```
┌───────────────────────────────────────────┐
│  [ This Week ] [ This Month ] [ All-Time ]  │
└───────────────────────────────────────────┘
                  ^^^^^^^^^^^^
                  Active: Filled orange background
                  White text
                  Orange border
                  Subtle shadow
```

## Badge Styling

### Rank Badges (Top 3)
```
#1: Gold gradient    ┌────┐
    #FFD700          │ #1 │  ← Shining animation
                     └────┘

#2: Silver gradient  ┌────┐
    #C0C0C0          │ #2 │
                     └────┘

#3: Bronze gradient  ┌────┐
    #CD7F32          │ #3 │
                     └────┘

#4+: Dark gradient   ┌────┐
     #1A1A1A         │ #4 │
                     └────┘
```

### Trending Badges (Trending Mode Only)
```
Score 90+:  ╔═══════════╗
ON FIRE     ║  ON FIRE  ║  ← Red-orange gradient (#ff6b35 → #ff5722)
            ╚═══════════╝

Score 75+:  ╔═══════╗
HOT         ║  HOT  ║  ← Orange gradient (#ff9800 → #ff6b35)
            ╚═══════╝

Score 60+:  ╔═══════════╗
TRENDING    ║ TRENDING  ║  ← Yellow-orange gradient (#ffc107 → #ff9800)
            ╚═══════════╝

Score <60:  ╔═══════╗
RISING      ║ RISING ║  ← Green gradient (#4caf50 → #8bc34a)
            ╚═══════╝
```

### Duration Badge (Most Watched Mode Only)
```
┌────────┐
│  25m   │  ← Bottom-left corner
└────────┘   Black background (80% opacity)
             White text
```

### Views Badge
```
Trending Mode:              Most Watched Mode:
┌─────────────┐            ┌─────────────┐
│ 👁️ 1.2K (24h) │            │ 👁️ 45.2K total│
└─────────────┘            └─────────────┘
Bottom-right corner         Bottom-right corner
```

## Responsive Behavior

### Desktop (>768px)
```
┌────────┬────────┬────────┬────────┐
│ Video  │ Video  │ Video  │ Video  │  ← 4 columns
├────────┼────────┼────────┼────────┤
│ Video  │ Video  │ Video  │ Video  │
└────────┴────────┴────────┴────────┘
```

### Tablet (768px)
```
┌────────┬────────┬────────┐
│ Video  │ Video  │ Video  │  ← 3 columns
├────────┼────────┼────────┤
│ Video  │ Video  │ Video  │
└────────┴────────┴────────┘

Toggle buttons stack:
┌─────────────────┐
│  🔥 Trending    │  ← Full width
├─────────────────┤
│ 📊 Most Watched │  ← Full width
└─────────────────┘
```

### Mobile (<480px)
```
┌──────────────┐
│   Video      │  ← Single column
├──────────────┤
│   Video      │
├──────────────┤
│   Video      │
└──────────────┘
```

## Interaction Flow

1. **User lands on Trending tab**
   - Sees "Trending Now" with default trending videos
   - Toggle shows "Trending" as active
   - No time period selector visible

2. **User clicks "Most Watched"**
   - Loading spinner appears briefly
   - Title changes to "📊 Most Watched"
   - Subtitle changes to "Top 25 most viewed this month"
   - Time period selector slides down with animation
   - "This Month" is selected by default
   - Video cards update to show:
     - Duration badges
     - Total views instead of 24h views
     - Unique viewers
     - Completion rate
   - Trending badges removed
   - Toggle shows "Most Watched" as active

3. **User selects "This Week"**
   - Loading spinner appears briefly
   - Subtitle changes to "Top 25 most viewed this week"
   - Videos reload with last 7 days data
   - "This Week" button becomes active

4. **User selects "All-Time"**
   - Loading spinner appears briefly
   - Subtitle changes to "Top 25 most viewed all-time"
   - Videos reload with all-time data
   - "All-Time" button becomes active

5. **User clicks back to "Trending"**
   - Loading spinner appears briefly
   - Title changes to "🔥 Trending Now"
   - Subtitle changes to "Most watched in the last 24 hours"
   - Time period selector disappears
   - Video cards update to show:
     - Trending badges
     - 24h views
     - Trending score bars
   - Duration badges removed
   - Toggle shows "Trending" as active

## Animation Details

### Card Entrance
- Stagger animation: Each card appears 0.05s after previous
- Slide up + fade in effect
- Total time: 1.25s for 25 cards

### Card Hover
```
Default:        Hover:
┌─────────┐    ┌─────────┐
│  Card   │ →  │  Card   │  ← Lifts 8px
└─────────┘    └─────────┘  ← Shadow expands
                              Image scales 105%
```

### Toggle Button Transition
- 0.2s smooth transition
- Background color
- Text color
- Shadow appearance

### Score Bar Animation (Trending)
- 0.5s ease transition when loading
- Fills from left to right
- Color matches trending badge

## Color Palette

### Brand Colors
- Primary Orange: `#ff6b35`
- Hover Orange: `#ff5722`

### Ranking Colors
- Gold: `#FFD700` → `#FFED4E`
- Silver: `#C0C0C0` → `#E8E8E8`
- Bronze: `#CD7F32` → `#E8A87C`
- Default: `#1A1A1A` → `#333333`

### Trending Badge Colors
- Fire (90+): `#ff6b35` → `#ff5722`
- Hot (75+): `#ff9800` → `#ff6b35`
- Trending (60+): `#ffc107` → `#ff9800`
- Rising (<60): `#4caf50` → `#8bc34a`

### UI Colors
- Text Dark: `#1a1a1a`
- Text Gray: `#666666`
- Text Light: `#999999`
- Background: `#ffffff`
- Border: `#e0e0e0`
- Hover BG: `#f5f5f5`

## Accessibility

### Keyboard Navigation
- Tab through toggle buttons
- Enter/Space to activate
- Tab through video cards
- Enter to open video

### Screen Reader
- Toggle buttons announce: "Trending" / "Most Watched"
- Active state announced: "selected"
- Video cards announce: title, views, and rank
- Loading state announced: "Loading videos"

### Focus States
- Toggle buttons: Visible focus ring
- Video cards: Visible focus outline
- Maintains color contrast ratios

## Performance

### Optimization
- Lazy loading images
- Debounced view mode switching
- Efficient re-renders (keyed lists)
- CSS animations (GPU accelerated)

### Loading States
- Skeleton screens for initial load
- Smooth transitions between modes
- Error states with retry buttons

