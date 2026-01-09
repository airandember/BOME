# Search Index Strand

## Overview

The Search Index strand provides lightning-fast video search functionality by pre-generating a static JSON index file that the frontend loads and searches client-side using Fuse.js fuzzy search.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                    SEARCH INDEX ARCHITECTURE                        │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Backend (Go)                                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  SearchIndexScheduler                                        │   │
│  │  - Cron: Midnight + 6 AM backup (MST)                       │   │
│  │  - Manual trigger via Admin API                              │   │
│  │  - Reads from: master_video_list table                       │   │
│  │  - Writes to: SEARCH_INDEX_PATH env var                      │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              ↓                                      │
│                 /app/frontend/static/search-index.json              │
│                              ↓                                      │
│  Frontend (Svelte)                                                  │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  /videos page                                                │   │
│  │  - Loads /search-index.json on mount                         │   │
│  │  - Builds Fuse.js index (once)                               │   │
│  │  - Fuzzy search with weighted fields                         │   │
│  │  - ~10ms search time for 1800+ videos                        │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

## Data Flow

1. **Generation (Backend)**
   - Scheduler runs at midnight and 6 AM (MST)
   - Admin can trigger manual generation via UI
   - Reads active videos from `master_video_list` where `vid_status = true`
   - Generates compact JSON (~2MB for 1800 videos)
   - Writes to path specified by `SEARCH_INDEX_PATH` env var

2. **Consumption (Frontend)**
   - On `/videos` page mount, fetches `/search-index.json`
   - Builds Fuse.js index once (never rebuilds)
   - Search executes client-side in ~10ms
   - Results sorted by relevance then date

## Search Configuration

### Fuse.js Options
```javascript
{
  keys: [
    { name: 'title', weight: 0.7 },       // 70% weight
    { name: 'description', weight: 0.2 }, // 20% weight
    { name: 'category', weight: 0.05 },   // 5% weight
    { name: 'tags', weight: 0.05 }        // 5% weight
  ],
  threshold: 0.4,        // 0.0 = exact, 1.0 = match anything
  includeScore: true,
  minMatchCharLength: 2,
  ignoreLocation: true   // Match anywhere in string
}
```

## Index Format (v2.0)

```json
{
  "version": "2.0",
  "generatedAt": "2026-01-09T00:00:00Z",
  "totalVideos": 1805,
  "metadata": {
    "generationTimeMs": 150,
    "source": "master_video_list",
    "indexedFields": ["title", "description", "category", "tags"]
  },
  "videos": [
    {
      "id": "bunny-video-guid",
      "title": "Video Title",
      "description": "Video description for search",
      "category": "Category Name",
      "tags": ["tag1", "tag2"],
      "duration": 3600,
      "createdAt": "2026-01-01T00:00:00Z",
      "thumbnail": "https://cdn.example.com/thumb.jpg",
      "thumbnailUrl": "https://cdn.example.com/thumb.jpg",
      "views": 1234,
      "status": "active",
      "videoUrl": "https://cdn.example.com/video.m3u8",
      "iframeSrc": "https://iframe.example.com/embed/guid",
      "bunny": {
        "guid": "bunny-video-guid",
        "previewImageUrl": "https://cdn.example.com/preview.jpg",
        "length": 3600
      }
    }
  ]
}
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SEARCH_INDEX_PATH` | Full path to write search-index.json | `./search-index.json` |

## API Endpoints

### Admin Routes (`/api/v1/admin/streaming/search-index/`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/scheduler/status` | Get scheduler status |
| POST | `/scheduler/trigger` | Trigger manual generation |
| GET | `/config` | Get scheduler configuration |
| POST | `/config` | Update scheduler configuration |
| GET | `/stats` | Get search index statistics |
| GET | `/download` | Download current search index file |

## Files

### Backend
- `backend/internal/services/search_index_scheduler.go` - Main scheduler service
- `backend/internal/routes/search_index.go` - Admin API routes

### Frontend
- `frontend/src/routes/videos/+page.svelte` - Search consumer (Fuse.js)
- `frontend/src/routes/admin/streaming/search-index/+page.svelte` - Admin UI

## Performance Characteristics

- **Generation time**: ~150ms for 1800 videos
- **File size**: ~2MB (compact JSON, no indentation)
- **Frontend load time**: ~200ms to fetch and parse
- **Search time**: ~10ms for any query
- **Memory usage**: ~5MB for Fuse.js index

## Troubleshooting

### Video not appearing in search
1. Check if video has `vid_status = true` in `master_video_list`
2. Check if video is in `master_video_list` table
3. Trigger manual search index generation
4. Verify `SEARCH_INDEX_PATH` is correctly set
5. Check if frontend can access `/search-index.json`

### Search index file not found
1. Verify `SEARCH_INDEX_PATH` environment variable
2. Check backend logs for write errors
3. Ensure directory exists and is writable
4. Trigger manual generation via admin panel

## Related Strands
- `bunny-views-sync` - Syncs view counts from Bunny.net
- `video-playback` - Video player and streaming
