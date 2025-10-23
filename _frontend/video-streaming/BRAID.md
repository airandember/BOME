# 🧬 Video Streaming Braid - Frontend
**Svelte5 video player and content discovery**

---

## 🔗 **Cross-Repository Braid**

> **⚠️ IMPORTANT**: This is the **frontend portion** of the Video Streaming Braid.  
> **Backend portion**: See `_backend/braids/video-streaming/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## 📋 **Frontend Overview**

**Purpose**: User interface for video browsing, playback, and content discovery  
**Technology**: Svelte 5, TypeScript, Video.js/Plyr, TailwindCSS  
**Entry Points**: `/videos`, `/videos/[id]`, `/videos/categories/[name]`  
**State Management**: Svelte stores for video library and player state

---

## 🎯 **Key Features**

### **1. Video Library Page**:
- Grid/list view of videos
- Category filtering
- Tag filtering
- Search functionality
- Sort options (newest, popular, trending)
- Pagination or infinite scroll
- Featured videos section

### **2. Video Player Page**:
- Full-featured video player
- Adaptive bitrate streaming (HLS)
- Quality selection (SD, HD, 4K)
- Playback speed control
- Subtitles/captions
- Picture-in-picture
- Fullscreen support
- Keyboard shortcuts
- Thumbnail preview scrubbing

### **3. Video Details**:
- Title and description
- View count and date
- Like/dislike buttons
- Share functionality
- Related videos sidebar
- Comments section
- Playlist add button

### **4. Category Browsing**:
- Category landing pages
- Subcategory navigation
- Category-specific filters
- Category thumbnails

### **5. Video Collections**:
- Curated playlists
- User-created playlists
- Watch later
- History
- Favorites

---

## 📄 **Frontend Pages**

### **1. Videos Library** (`/videos`)
**File**: `frontend/src/routes/videos/+page.svelte`

**Features**:
- Video grid with thumbnails
- Filter by category/tag
- Search bar
- Sort dropdown
- Load more/infinite scroll

**Example UI**:
```svelte
<script lang="ts">
  import { onMount } from 'svelte';
  import { videoStore } from '$lib/stores/videos';
  import VideoCard from '$lib/components/VideoCard.svelte';
  
  let filter = {
    category: '',
    tag: '',
    search: '',
    sort: 'newest'
  };
  
  onMount(async () => {
    await videoStore.loadVideos(filter);
  });
  
  $: videos = $videoStore.videos;
  $: loading = $videoStore.loading;
  
  async function loadMore() {
    await videoStore.loadMore();
  }
</script>

<div class="videos-page">
  <header>
    <h1>Videos</h1>
    
    <!-- Search Bar -->
    <div class="search-bar">
      <input
        type="text"
        placeholder="Search videos..."
        bind:value={filter.search}
        on:input={() => videoStore.search(filter.search)}
      />
    </div>
    
    <!-- Filters -->
    <div class="filters">
      <select bind:value={filter.category}>
        <option value="">All Categories</option>
        <option value="tutorials">Tutorials</option>
        <option value="reviews">Reviews</option>
        <option value="news">News</option>
      </select>
      
      <select bind:value={filter.sort}>
        <option value="newest">Newest First</option>
        <option value="popular">Most Popular</option>
        <option value="trending">Trending</option>
      </select>
    </div>
  </header>
  
  <!-- Video Grid -->
  <div class="video-grid">
    {#if loading && videos.length === 0}
      <div class="loading">Loading videos...</div>
    {:else}
      {#each videos as video (video.id)}
        <VideoCard {video} />
      {/each}
    {/if}
  </div>
  
  <!-- Load More -->
  {#if !loading && videos.length > 0}
    <button on:click={loadMore} class="load-more">
      Load More
    </button>
  {/if}
</div>

<style>
  .video-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 2rem;
  }
  
  @media (max-width: 768px) {
    .video-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
```

---

### **2. Video Player Page** (`/videos/[id]`)
**File**: `frontend/src/routes/videos/[id]/+page.svelte`

**Features**:
- Video player embed
- Title, description, metadata
- Like/share/save buttons
- Related videos
- Comments section
- Up next autoplay

**Example UI**:
```svelte
<script lang="ts">
  import { page } from '$app/stores';
  import { onMount, onDestroy } from 'svelte';
  import { videoStore } from '$lib/stores/videos';
  import VideoPlayer from '$lib/components/VideoPlayer.svelte';
  
  const videoId = $page.params.id;
  
  let video = null;
  let relatedVideos = [];
  
  onMount(async () => {
    video = await videoStore.getVideo(videoId);
    relatedVideos = await videoStore.getRelatedVideos(videoId);
    
    // Track view
    videoStore.trackView(videoId);
  });
  
  function handleVideoEnded() {
    // Autoplay next video
    if (relatedVideos.length > 0) {
      goto(`/videos/${relatedVideos[0].id}`);
    }
  }
</script>

{#if video}
  <div class="video-page">
    <!-- Video Player -->
    <div class="player-container">
      <VideoPlayer
        src={video.streamUrl}
        poster={video.thumbnailUrl}
        on:ended={handleVideoEnded}
      />
    </div>
    
    <!-- Video Info -->
    <div class="video-info">
      <h1>{video.title}</h1>
      
      <div class="meta">
        <span>{video.viewCount.toLocaleString()} views</span>
        <span>•</span>
        <span>{new Date(video.publishedAt).toLocaleDateString()}</span>
      </div>
      
      <!-- Actions -->
      <div class="actions">
        <button on:click={() => videoStore.likeVideo(video.id)}>
          👍 {video.likeCount}
        </button>
        <button on:click={() => videoStore.shareVideo(video.id)}>
          📤 Share
        </button>
        <button on:click={() => videoStore.saveVideo(video.id)}>
          💾 Save
        </button>
      </div>
      
      <!-- Description -->
      <div class="description">
        <p>{video.description}</p>
      </div>
    </div>
    
    <!-- Related Videos Sidebar -->
    <aside class="related-videos">
      <h2>Up Next</h2>
      {#each relatedVideos as relatedVideo}
        <a href="/videos/{relatedVideo.id}" class="related-video-card">
          <img src={relatedVideo.thumbnailUrl} alt={relatedVideo.title} />
          <div>
            <h3>{relatedVideo.title}</h3>
            <p>{relatedVideo.viewCount} views</p>
          </div>
        </a>
      {/each}
    </aside>
  </div>
{/if}

<style>
  .video-page {
    display: grid;
    grid-template-columns: 1fr 350px;
    gap: 2rem;
    max-width: 1600px;
    margin: 0 auto;
  }
  
  .player-container {
    grid-column: 1;
    aspect-ratio: 16 / 9;
    background: black;
  }
  
  .video-info {
    grid-column: 1;
  }
  
  .related-videos {
    grid-column: 2;
    grid-row: 1 / span 2;
  }
  
  @media (max-width: 1024px) {
    .video-page {
      grid-template-columns: 1fr;
    }
    
    .related-videos {
      grid-column: 1;
      grid-row: auto;
    }
  }
</style>
```

---

### **3. Category Page** (`/videos/categories/[name]`)
**File**: `frontend/src/routes/videos/categories/[name]/+page.svelte`

**Features**:
- Category-specific video grid
- Category description
- Subcategory navigation
- Breadcrumb navigation

---

### **4. Collections Page** (`/videos/collections/[id]`)
**File**: `frontend/src/routes/videos/collections/[id]/+page.svelte`

**Features**:
- Playlist of videos
- Play all button
- Reorder videos (if owner)
- Share playlist
- Add to playlist button

---

## 🧩 **Frontend Components**

### **VideoCard Component** (`$lib/components/VideoCard.svelte`):
**Purpose**: Display video thumbnail and metadata

```svelte
<script lang="ts">
  export let video: Video;
  
  function formatDuration(seconds: number): string {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  }
</script>

<a href="/videos/{video.id}" class="video-card">
  <div class="thumbnail">
    <img src={video.thumbnailUrl} alt={video.title} loading="lazy" />
    <span class="duration">{formatDuration(video.duration)}</span>
    {#if video.requiresSubscription}
      <span class="badge premium">Premium</span>
    {/if}
  </div>
  
  <div class="info">
    <h3>{video.title}</h3>
    <p class="meta">
      {video.viewCount.toLocaleString()} views • 
      {new Date(video.publishedAt).toLocaleDateString()}
    </p>
  </div>
</a>

<style>
  .video-card {
    display: block;
    text-decoration: none;
    color: inherit;
    transition: transform 0.2s;
  }
  
  .video-card:hover {
    transform: translateY(-4px);
  }
  
  .thumbnail {
    position: relative;
    aspect-ratio: 16 / 9;
    border-radius: 8px;
    overflow: hidden;
  }
  
  .thumbnail img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .duration {
    position: absolute;
    bottom: 8px;
    right: 8px;
    background: rgba(0, 0, 0, 0.8);
    color: white;
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 0.875rem;
  }
  
  .badge.premium {
    position: absolute;
    top: 8px;
    left: 8px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: bold;
  }
</style>
```

---

### **VideoPlayer Component** (`$lib/components/VideoPlayer.svelte`):
**Purpose**: HTML5 video player with HLS support

```svelte
<script lang="ts">
  import { onMount } from 'svelte';
  import Hls from 'hls.js';
  
  export let src: string;
  export let poster: string = '';
  export let autoplay = false;
  
  let videoElement: HTMLVideoElement;
  let hls: Hls;
  
  onMount(() => {
    if (Hls.isSupported()) {
      hls = new Hls();
      hls.loadSource(src);
      hls.attachMedia(videoElement);
      
      hls.on(Hls.Events.MANIFEST_PARSED, () => {
        if (autoplay) {
          videoElement.play();
        }
      });
    } else if (videoElement.canPlayType('application/vnd.apple.mpegurl')) {
      // Native HLS support (Safari)
      videoElement.src = src;
    }
    
    return () => {
      if (hls) {
        hls.destroy();
      }
    };
  });
  
  function handleTimeUpdate() {
    // Track progress
    const progress = (videoElement.currentTime / videoElement.duration) * 100;
    dispatch('progress', { progress, currentTime: videoElement.currentTime });
  }
</script>

<video
  bind:this={videoElement}
  {poster}
  controls
  playsinline
  on:timeupdate={handleTimeUpdate}
  on:ended
  on:play
  on:pause
>
  <source src={src} type="application/x-mpegURL" />
  Your browser doesn't support video playback.
</video>

<style>
  video {
    width: 100%;
    height: 100%;
    background: black;
  }
</style>
```

---

## 🗃️ **Frontend Stores**

### **Video Store** (`$lib/stores/videos.ts`):
**Purpose**: Manage video library state

```typescript
interface VideoState {
  videos: Video[];
  currentVideo: Video | null;
  relatedVideos: Video[];
  categories: Category[];
  loading: boolean;
  error: string | null;
  pagination: {
    page: number;
    total: number;
    hasMore: boolean;
  };
}

export const videoStore = {
  async loadVideos(filter: VideoFilter) {
    // GET /api/v1/videos?category=...&sort=...
  },
  
  async getVideo(id: string): Promise<Video> {
    // GET /api/v1/videos/:id
  },
  
  async getRelatedVideos(id: string): Promise<Video[]> {
    // GET /api/v1/videos/:id/related
  },
  
  async searchVideos(query: string): Promise<Video[]> {
    // GET /api/v1/videos?search=...
  },
  
  async trackView(id: string) {
    // POST /api/v1/videos/:id/view
  },
  
  async trackProgress(id: string, currentTime: number, duration: number) {
    // POST /api/v1/videos/:id/progress
  },
  
  async likeVideo(id: string) {
    // POST /api/v1/videos/:id/like
  },
  
  async saveVideo(id: string) {
    // POST /api/v1/videos/:id/save
  }
};
```

---

## 🔄 **Data Flow Examples**

### **Watch Video Flow**:
```
1. User clicks video card
2. Navigate to /videos/:id
3. Load video metadata
4. Request streaming URL from backend
5. Backend generates signed URL
6. Initialize HLS player
7. Video streams from Bunny CDN
8. Track view event
9. Periodically track progress
10. On complete, suggest next video
```

### **Search Videos**:
```
1. User types in search bar
2. Debounce input (300ms)
3. API: GET /api/v1/videos?search=query
4. Backend searches database
5. Return matching videos
6. Update video grid
```

---

## 🎨 **Video Player Features**

### **Controls**:
- ▶️ Play/Pause
- ⏮️ Rewind 10s
- ⏭️ Forward 10s
- 🔊 Volume control
- ⚙️ Quality selection (SD, HD, 4K)
- 🎬 Playback speed (0.5x, 1x, 1.5x, 2x)
- 📺 Picture-in-picture
- ⛶ Fullscreen
- 💬 Subtitles/CC

### **Keyboard Shortcuts**:
- `Space`: Play/pause
- `←/→`: Seek 5s
- `↑/↓`: Volume up/down
- `M`: Mute
- `F`: Fullscreen
- `C`: Toggle captions

---

## 📊 **Video Analytics Display**

### **Public Metrics**:
- View count
- Like/dislike ratio
- Publication date
- Duration

### **Creator Metrics** (if admin/creator):
- Real-time viewers
- Watch time distribution
- Audience retention graph
- Traffic sources
- Geographic distribution

---

## 🔒 **Access Control UI**

### **Free vs. Premium Content**:
```svelte
{#if video.requiresSubscription && !$auth.user?.hasSubscription}
  <div class="locked-content">
    <div class="blur-overlay"></div>
    <div class="unlock-prompt">
      <h2>Premium Content</h2>
      <p>Subscribe to watch this video</p>
      <button on:click={() => goto('/subscription')}>
        Subscribe Now
      </button>
    </div>
  </div>
{:else}
  <VideoPlayer src={video.streamUrl} />
{/if}
```

---

## ⚡ **Performance Optimizations**

### **Lazy Loading**:
- Thumbnail images lazy load
- Defer non-critical video metadata
- Intersection Observer for infinite scroll

### **Caching**:
- Cache video metadata locally
- Cache thumbnails with service worker
- Prefetch next video in playlist

### **Network Optimization**:
- HLS adaptive streaming
- CDN edge caching
- Thumbnail sprite sheets for hover preview

---

## 📝 **Known Issues**

### **To Implement**:
1. Video recommendations algorithm
2. Continue watching feature
3. Watch history
4. Playlist creation
5. Video chapters/timestamps
6. Interactive video elements
7. Live chat for live streams
8. Download for offline viewing (premium)

---

## 🚀 **Quick Links**

**Actual Files**:
- Videos Library: `frontend/src/routes/videos/+page.svelte`
- Video Player: `frontend/src/routes/videos/[id]/+page.svelte`
- Categories: `frontend/src/routes/videos/categories/[name]/+page.svelte`
- VideoCard Component: `frontend/src/lib/components/VideoCard.svelte`
- Video Store: `frontend/src/lib/stores/videos.ts`

---

**Last Updated**: October 14, 2025  
**Status**: Core business UI  
**Technology**: Svelte 5 + HLS.js  
**Backend Counterpart**: `_backend/braids/video-streaming/`

---

**Navigate**:  
[🏠 Master Index](../../../BRAIDS_INDEX.md) | [⬅️ Backend Braid](../../_backend/braids/video-streaming/BRAID.md)

