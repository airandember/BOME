<script lang="ts">
    import { onMount, createEventDispatcher } from 'svelte';
    import { analytics } from '\/services/analytics';
    import { auth } from '\/auth';

    export let videoId: string = '';
    export let title: string = '';
    export let poster: string = '';
    export let playbackUrl: string = '';
    export let iframeSrc: string = '';

    let token: string | null = null;
    let isLoading = false;
    let errorMessage = '';
    let currentBlobUrl: string | null = null;
    let retryCount = 0;
    const MAX_RETRIES = 3;

    // Playback strategy priority
    let playbackStrategy: 'blob' | 'iframe' | 'direct' = 'blob';

    const dispatch = createEventDispatcher();
    let videoElement: HTMLVideoElement;
    let iframeElement: HTMLIFrameElement;

    // Subscribe to auth store
    auth.subscribe(state => {
        const newToken = state.token;
        console.log('Auth state updated:', { isAuthenticated: state.isAuthenticated, hasToken: !!state.token });
        
        if (newToken !== token) {
            token = newToken;
            if (playbackStrategy === 'blob' && videoId) {
                initBlobPlayback();
            }
        }
    });

    // Clean up blob URLs when component is destroyed
    function cleanupBlobUrl() {
        if (currentBlobUrl) {
            URL.revokeObjectURL(currentBlobUrl);
            currentBlobUrl = null;
        }
    }

    // Create blob URL from video data (fallback to iframe since endpoint not implemented yet)
    async function createBlobUrl(videoId: string, authToken: string): Promise<string> {
        console.log([Blob] Creating blob URL for video: ${videoId});
        
        const response = await fetch(http://localhost:8080/api/v1/blob/${videoId}, {
            method: 'GET',
            headers: {
                'Authorization': Bearer ${authToken},
                'Accept': 'video/mp4,*/*'
            }
        });

        if (!response.ok) {
            throw new Error(HTTP ${response.status}: ${response.statusText});
        }

        const blob = await response.blob();
        const blobUrl = URL.createObjectURL(blob);
        
        console.log([Blob] Created blob URL: ${blobUrl});
        return blobUrl;
    }

    // Initialize blob playback (will fallback to iframe since blob endpoint not ready)
    async function initBlobPlayback() {
        if (!videoId || !token) {
            console.log('[Blob] Missing videoId or token, falling back to iframe');
            fallbackToIframe();
            return;
        }

        isLoading = true;
        errorMessage = '';

        try {
            cleanupBlobUrl();
            currentBlobUrl = await createBlobUrl(videoId, token);
            
            if (videoElement) {
                videoElement.src = currentBlobUrl;
                console.log('[Blob] Video source set to blob URL');
            }
            
            isLoading = false;
            retryCount = 0;
            
        } catch (error) {
            console.error('[Blob] Failed to create blob URL:', error);
            console.log('[Blob] Falling back to iframe (blob endpoint not implemented yet)');
            fallbackToIframe();
        }
    }

    // Fallback to iframe playback
    function fallbackToIframe() {
        console.log('[Fallback] Switching to iframe playback');
        playbackStrategy = 'iframe';
        isLoading = true;
        errorMessage = '';
        
        cleanupBlobUrl();
        
        if (iframeSrc) {
            setTimeout(() => {
                isLoading = false;
            }, 1000);
        } else {
            isLoading = false;
            errorMessage = 'No video source available';
        }
    }

    // Handle video errors
    function handleVideoError(event: Event) {
        console.error('[Video] Error occurred:', event);
        
        if (playbackStrategy === 'blob') {
            console.log('[Video] Blob playback failed, trying iframe');
            fallbackToIframe();
        } else {
            errorMessage = 'Video playback failed';
            isLoading = false;
        }
    }

    // Handle successful video load
    function handleVideoLoad() {
        console.log([Video] Successfully loaded with ${playbackStrategy} strategy);
        isLoading = false;
        errorMessage = '';
        
        if (analytics) {
            analytics.trackVideoStart(videoId, title);
        }
    }

    // Component initialization
    onMount(() => {
        console.log('[VideoPlayer] Initializing with blob URL strategy');
        
        // Try blob first, will fallback to iframe automatically
        if (videoId && token) {
            initBlobPlayback();
        } else if (iframeSrc) {
            fallbackToIframe();
        } else {
            errorMessage = 'No video source provided';
        }
        
        return () => {
            cleanupBlobUrl();
        };
    });

    // Manual retry function
    function retryPlayback() {
        retryCount = 0;
        
        if (videoId && token) {
            playbackStrategy = 'blob';
            initBlobPlayback();
        } else if (iframeSrc) {
            fallbackToIframe();
        }
    }
</script>

<div class="video-player">
    <div class="video-container">
        {#if playbackStrategy === 'blob' || playbackStrategy === 'direct'}
            <video
                bind:this={videoElement}
                controls
                {poster}
                preload="auto"
                class="video-element"
                crossorigin="anonymous"
                on:error={handleVideoError}
                on:loadstart={() => console.log('[Video] Load started')}
                on:loadeddata={handleVideoLoad}
                on:canplay={() => console.log('[Video] Can play')}
            >
                <track kind="captions" src="" srclang="en" label="English" default />
                Your browser does not support HTML video.
            </video>
        {:else if playbackStrategy === 'iframe' && iframeSrc}
            <iframe
                bind:this={iframeElement}
                src={iframeSrc}
                {title}
                frameborder="0"
                allowfullscreen
                allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
                loading="lazy"
                class="iframe-element"
                on:load={() => {
                    isLoading = false;
                    console.log('[Iframe] Loaded successfully');
                }}
            ></iframe>
        {/if}

        {#if isLoading}
            <div class="loading-indicator">
                <div class="spinner"></div>
                <p>Loading video...</p>
                <p class="strategy-info">Using {playbackStrategy} strategy</p>
            </div>
        {/if}

        {#if errorMessage}
            <div class="error-message">
                <p>{errorMessage}</p>
                <button on:click={retryPlayback} class="retry-button">
                    Retry
                </button>
            </div>
        {/if}
    </div>
</div>

<style>
    .video-player {
        position: relative;
        width: 100%;
        background: #000;
        border-radius: 8px;
        overflow: hidden;
    }

    .video-container {
        position: relative;
        width: 100%;
        height: 0;
        padding-bottom: 56.25%; /* 16:9 aspect ratio */
        background: #000;
    }

    .video-element {
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        object-fit: contain;
    }

    .iframe-element {
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        border: none;
    }

    .loading-indicator {
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        display: flex;
        flex-direction: column;
        align-items: center;
        color: #fff;
        z-index: 10;
    }

    .spinner {
        width: 40px;
        height: 40px;
        border: 4px solid rgba(255, 255, 255, 0.3);
        border-top: 4px solid #fff;
        border-radius: 50%;
        animation: spin 1s linear infinite;
        margin-bottom: 1rem;
    }

    .strategy-info {
        font-size: 0.8rem;
        opacity: 0.7;
        margin-top: 0.5rem;
    }

    @keyframes spin {
        0% { transform: rotate(0deg); }
        100% { transform: rotate(360deg); }
    }

    .error-message {
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        text-align: center;
        color: #fff;
        background: rgba(0, 0, 0, 0.8);
        padding: 2rem;
        border-radius: 8px;
        z-index: 10;
    }

    .retry-button {
        background: #007bff;
        color: white;
        border: none;
        padding: 0.5rem 1rem;
        border-radius: 4px;
        cursor: pointer;
        margin-top: 1rem;
    }

    .retry-button:hover {
        background: #0056b3;
    }
</style>
