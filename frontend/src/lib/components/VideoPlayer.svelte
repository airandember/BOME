<script lang="ts">
    import { onMount, createEventDispatcher } from 'svelte';
    import { analytics } from '$lib/services/analytics';
    import { auth } from '$lib/auth';
    import Hls from 'hls.js';

    export let videoId: string = '';
    export let title: string = '';
    export let poster: string = '';
    export let playbackUrl: string = '';

    let token: string | null = null;
    let hls: Hls | null = null;
    let retryCount = 0;
    const MAX_RETRIES = 3;

    // Subscribe to auth store
    auth.subscribe(state => {
        const newToken = state.token;
        console.log('Auth state updated:', { isAuthenticated: state.isAuthenticated, hasToken: !!state.token });
        
        // If token changed and we have an active HLS instance, reload the stream
        if (newToken !== token && hls) {
            token = newToken;
            console.log('Token changed, reloading stream');
            retryCount = 0; // Reset retry count on token change
            hls.loadSource(proxyUrl);
        } else {
            token = newToken;
        }
    });

    // Convert Bunny.net URL to our proxy URL
    $: proxyUrl = playbackUrl ? playbackUrl.replace(
        /https:\/\/vz-[^\/]+\.b-cdn\.net\/([^\/]+)(\/.*)?/,
        `${window.location.origin}/api/v1/stream/$1$2`
    ) : '';

    const dispatch = createEventDispatcher();
    let videoElement: HTMLVideoElement;
    let errorMessage = '';

    function initHls() {
        if (!token) {
            console.error('No auth token available');
            errorMessage = 'Authentication required';
            return;
        }

        if (videoElement && proxyUrl) {
            console.log('Setting up video with URL:', proxyUrl);
            
            if (Hls.isSupported()) {
                console.log('HLS.js is supported');
                
                // Destroy existing instance if any
                if (hls) {
                    hls.destroy();
                }

                hls = new Hls({
                    debug: true, // Enable debug logs
                    enableWorker: true,
                    maxLoadingDelay: 4000,
                    xhrSetup: function(xhr, url) {
                        // Add auth token to all HLS requests
                        if (token) {
                            console.log('Adding auth token to request:', url);
                            xhr.setRequestHeader('Authorization', `Bearer ${token}`);
                            // Add additional headers that might be needed
                            xhr.setRequestHeader('Accept', 'application/x-mpegURL,*/*');
                            xhr.withCredentials = false; // Important for CORS
                        } else {
                            console.error('No auth token available for request:', url);
                        }
                    }
                });
                
                hls.on(Hls.Events.ERROR, function(event, data) {
                    console.error('HLS error:', event, data);
                    if (data.fatal) {
                        switch (data.type) {
                            case Hls.ErrorTypes.NETWORK_ERROR:
                                errorMessage = 'Network error while loading video';
                                console.error('Network error:', data.details, data.response);
                                
                                // Retry logic for network errors
                                if (retryCount < MAX_RETRIES) {
                                    console.log(`Retrying (${retryCount + 1}/${MAX_RETRIES})...`);
                                    retryCount++;
                                    setTimeout(() => {
                                        if (hls) {
                                            console.log('Reloading stream...');
                                            hls.loadSource(proxyUrl);
                                        }
                                    }, 1000 * retryCount); // Exponential backoff
                                } else {
                                    console.error('Max retries reached');
                                }
                                break;
                            case Hls.ErrorTypes.MEDIA_ERROR:
                                errorMessage = 'Media error while playing video';
                                console.error('Media error:', data.details);
                                hls?.recoverMediaError();
                                break;
                            default:
                                errorMessage = 'Error playing video';
                                console.error('Fatal error:', data.details);
                                break;
                        }
                    }
                });

                hls.on(Hls.Events.MANIFEST_PARSED, function() {
                    console.log('HLS manifest parsed, attempting playback');
                    videoElement.play().catch(e => {
                        console.error('Playback failed:', e);
                        errorMessage = 'Failed to start video playback';
                    });
                });

                hls.on(Hls.Events.LEVEL_LOADED, function() {
                    console.log('HLS level loaded successfully');
                    errorMessage = ''; // Clear any previous errors
                    retryCount = 0; // Reset retry counter on successful load
                });

                hls.loadSource(proxyUrl);
                hls.attachMedia(videoElement);
            }
            // For Safari that has built-in HLS support
            else if (videoElement.canPlayType('application/vnd.apple.mpegurl')) {
                console.log('Using native HLS support');
                // For Safari, we need to handle auth manually
                const headers = new Headers();
                headers.append('Authorization', `Bearer ${token}`);
                
                fetch(proxyUrl, { headers })
                    .then(response => {
                        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
                        return response.blob();
                    })
                    .then(blob => {
                        videoElement.src = URL.createObjectURL(blob);
                    })
                    .catch(e => {
                        console.error('Video playback error:', e);
                        errorMessage = 'Error playing video';
                    });
                
                videoElement.onerror = function(e) {
                    console.error('Video playback error:', e);
                    errorMessage = 'Error playing video';
                };
            }

            videoElement.addEventListener('play', () => {
                errorMessage = ''; // Clear any error message on successful play
                analytics.trackVideoEvent(videoId, 'play', {
                    currentTime: videoElement.currentTime || 0,
                    duration: videoElement.duration || 0
                });
            });

            videoElement.addEventListener('ended', () => {
                analytics.trackVideoEvent(videoId, 'complete', {
                    duration: videoElement.duration || 0
                });
            });
        }
    }

    onMount(() => {
        initHls();
        
        return () => {
            if (hls) {
                hls.destroy();
                hls = null;
            }
        };
    });
</script>

<div class="video-player">
    {#if videoId && proxyUrl}
        <video
            bind:this={videoElement}
            controls
            {poster}
            preload="auto"
            class="video-element"
        >
            Your browser does not support HTML video.
        </video>
        {#if errorMessage}
            <div class="error-message">
                {errorMessage}
            </div>
        {/if}
    {:else}
        <div class="video-placeholder">
            <p>Video not available</p>
        </div>
    {/if}
</div>

<style>
    .video-player {
        width: 100%;
        max-width: 1280px;
        margin: 0 auto;
        background: #000;
        position: relative;
    }

    .video-element {
        width: 100%;
        height: auto;
        display: block;
    }

    .video-placeholder {
        width: 100%;
        height: 100%;
        min-height: 300px;
        display: flex;
        align-items: center;
        justify-content: center;
        background: #000;
        color: #fff;
    }

    .error-message {
        position: absolute;
        bottom: 0;
        left: 0;
        right: 0;
        padding: 1rem;
        background: rgba(0, 0, 0, 0.8);
        color: #fff;
        text-align: center;
    }
</style> 