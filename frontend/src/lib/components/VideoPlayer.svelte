<script lang="ts">
    import { onMount, createEventDispatcher } from 'svelte';
    import { analytics } from '$lib/services/analytics';
    import { auth } from '$lib/auth';
    import Hls from 'hls.js';

    export let videoId: string = '';
    export let title: string = '';
    export let poster: string = '';
    export let playbackUrl: string = '';
    export let iframeSrc: string = '';

    let token: string | null = null;
    let hls: Hls | null = null;
    let retryCount = 0;
    const MAX_RETRIES = 3;
    let useIframe = false;
    let isLoading = false;
    let loadStartTime = 0;

    // Performance monitoring
    let performanceMetrics = {
        loadTime: 0,
        bufferHealth: 0,
        errorCount: 0
    };

    // Subscribe to auth store
    auth.subscribe(state => {
        const newToken = state.token;
        console.log('Auth state updated:', { isAuthenticated: state.isAuthenticated, hasToken: !!state.token });
        
        // If token changed and we have an active HLS instance, reload the stream
        if (newToken !== token && hls && !useIframe) {
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
    let iframeElement: HTMLIFrameElement;
    let errorMessage = '';

    function switchToIframe() {
        console.log('Switching to iframe playback:', iframeSrc);
        useIframe = true;
        errorMessage = '';
        isLoading = false;
        
        if (hls) {
            hls.destroy();
            hls = null;
        }
    }

    function initHls() {
        if (useIframe) {
            return; // Don't initialize HLS if using iframe
        }

        if (!token) {
            console.error('No auth token available, switching to iframe');
            switchToIframe();
            return;
        }

        if (videoElement && proxyUrl) {
            console.log('Setting up HLS video with URL:', proxyUrl);
            isLoading = true;
            loadStartTime = performance.now();
            
            if (Hls.isSupported()) {
                console.log('HLS.js is supported');
                
                // Destroy existing instance if any
                if (hls) {
                    hls.destroy();
                }

                // Optimized HLS configuration
                hls = new Hls({
                    debug: false, // Disable debug in production
                    enableWorker: true,
                    lowLatencyMode: true,
                    backBufferLength: 90,
                    maxBufferLength: 30,
                    maxMaxBufferLength: 600,
                    maxLoadingDelay: 4000,
                    maxBufferSize: 60 * 1000 * 1000, // 60MB
                    maxBufferHole: 0.5,
                    highBufferWatchdogPeriod: 2,
                    nudgeOffset: 0.1,
                    nudgeMaxRetry: 3,
                    maxFragLookUpTolerance: 0.25,
                    liveSyncDurationCount: 3,
                    liveMaxLatencyDurationCount: 10,
                    liveDurationInfinity: false,
                    xhrSetup: function(xhr, url) {
                        if (token) {
                            xhr.setRequestHeader('Authorization', `Bearer ${token}`);
                            xhr.setRequestHeader('Accept', 'application/x-mpegURL,*/*');
                            xhr.withCredentials = false;
                        }
                    }
                });

                // Performance monitoring events
                hls.on(Hls.Events.MANIFEST_LOADED, () => {
                    console.log('HLS manifest loaded');
                    performanceMetrics.loadTime = performance.now() - loadStartTime;
                    isLoading = false;
                });

                hls.on(Hls.Events.LEVEL_LOADED, (event, data) => {
                    performanceMetrics.bufferHealth = data.details.totalduration;
                });

                hls.on(Hls.Events.ERROR, (event, data) => {
                    console.error('HLS error:', data);
                    performanceMetrics.errorCount++;
                    
                    if (data.fatal) {
                        switch (data.type) {
                            case Hls.ErrorTypes.NETWORK_ERROR:
                                console.log('Network error, attempting to recover...');
                                if (retryCount < MAX_RETRIES) {
                                    retryCount++;
                                    setTimeout(() => {
                                        hls?.startLoad();
                                    }, 1000 * retryCount);
                                } else {
                                    console.log('Max retries reached, switching to iframe');
                                    switchToIframe();
                                }
                                break;
                            case Hls.ErrorTypes.MEDIA_ERROR:
                                console.log('Media error, attempting to recover...');
                                if (retryCount < MAX_RETRIES) {
                                    retryCount++;
                                    hls?.recoverMediaError();
                                } else {
                                    switchToIframe();
                                }
                                break;
                            default:
                                console.log('Fatal error, switching to iframe');
                                switchToIframe();
                                break;
                        }
                    }
                });

                hls.on(Hls.Events.MEDIA_ATTACHED, () => {
                    console.log('HLS media attached');
                    hls?.loadSource(proxyUrl);
                });

                hls.attachMedia(videoElement);
            } else if (videoElement.canPlayType('application/vnd.apple.mpegurl')) {
                // Native HLS support (Safari)
                console.log('Using native HLS support');
                videoElement.src = proxyUrl;
                isLoading = false;
            } else {
                console.log('HLS not supported, switching to iframe');
                switchToIframe();
            }
        }
    }

    onMount(() => {
        // Try HLS first, fallback to iframe if it fails
        if (playbackUrl && playbackUrl.includes('playlist.m3u8')) {
            initHls();
        } else if (iframeSrc) {
            switchToIframe();
        } else {
            errorMessage = 'No valid video URL provided';
        }
        
        return () => {
            if (hls) {
                hls.destroy();
                hls = null;
            }
        };
    });
</script>

<div class="video-player">
    {#if useIframe && iframeSrc}
        <iframe
            bind:this={iframeElement}
            src={iframeSrc}
            {title}
            frameborder="0"
            allowfullscreen
            allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
            class="iframe-element"
        ></iframe>
    {:else if videoId && proxyUrl}
        <video
            bind:this={videoElement}
            controls
            {poster}
            preload="auto"
            class="video-element"
        >
            Your browser does not support HTML video.
        </video>
    {/if}
    
    {#if errorMessage}
        <div class="error-message">
            {errorMessage}
            {#if iframeSrc && !useIframe}
                <button on:click={switchToIframe} class="fallback-button">
                    Try alternative player
                </button>
            {/if}
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

    .iframe-element {
        width: 100%;
        height: 100%;
        min-height: 300px;
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

    .fallback-button {
        margin-top: 1rem;
        padding: 0.5rem 1rem;
        background-color: #007bff;
        color: white;
        border: none;
        border-radius: 5px;
        cursor: pointer;
        font-size: 1rem;
    }

    .fallback-button:hover {
        background-color: #0056b3;
    }
</style> 