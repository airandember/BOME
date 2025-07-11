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

    // Player.js instance for iframe control
    let playerJsInstance: any = null;
    let playerJsReady = false;

    // Load Player.js library dynamically
    function loadPlayerJs(): Promise<void> {
        return new Promise((resolve, reject) => {
            if ((window as any).playerjs) {
                resolve();
                return;
            }
            
            const script = document.createElement('script');
            script.src = 'https://assets.mediadelivery.net/playerjs/player-0.1.0.min.js';
            script.onload = () => resolve();
            script.onerror = () => reject(new Error('Failed to load Player.js'));
            document.head.appendChild(script);
        });
    }

    // Initialize Player.js for iframe control
    async function initPlayerJs() {
        if (!iframeElement) return;
        
        try {
            await loadPlayerJs();
            playerJsInstance = new (window as any).playerjs.Player(iframeElement);
            
            playerJsInstance.on('ready', () => {
                console.log('Player.js ready');
                playerJsReady = true;
                isLoading = false;
                
                // You can now control the player programmatically
                // playerJsInstance.play();
                // playerJsInstance.pause();
                // playerJsInstance.setCurrentTime(30);
            });
            
            playerJsInstance.on('play', () => {
                console.log('Video started playing');
            });
            
            playerJsInstance.on('pause', () => {
                console.log('Video paused');
            });
            
            playerJsInstance.on('ended', () => {
                console.log('Video ended');
            });
            
            playerJsInstance.on('error', (error: any) => {
                console.error('Player.js error:', error);
            });
            
        } catch (error) {
            console.error('Failed to initialize Player.js:', error);
        }
    }

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

    // Convert Bunny.net URL to our proxy URL - FIX: Use backend port 8080
    $: proxyUrl = playbackUrl ? playbackUrl.replace(
        /https:\/\/vz-[^\/]+\.b-cdn\.net\/([^\/]+)(\/.*)?/,
        `http://localhost:8080/api/v1/stream/$1$2`
    ) : '';

    // Also try to extract direct video URL from iframe if available
    $: directVideoUrl = iframeSrc ? extractDirectVideoUrl(iframeSrc) : '';

    function extractDirectVideoUrl(iframeSrc: string): string {
        // Extract video ID from iframe URL: https://iframe.mediadelivery.net/embed/347378/VIDEO_ID
        const match = iframeSrc.match(/\/embed\/\d+\/([^\/\?]+)/);
        if (match) {
            const videoId = match[1];
            // Return direct play URL - use /play/ for direct video playback
            return `https://iframe.mediadelivery.net/play/347378/${videoId}`;
        }
        return '';
    }

    const dispatch = createEventDispatcher();
    let videoElement: HTMLVideoElement;
    let iframeElement: HTMLIFrameElement;
    let errorMessage = '';

    function switchToIframe() {
        console.log('Switching to iframe playback:', iframeSrc);
        useIframe = true;
        errorMessage = '';
        isLoading = true;
        
        if (hls) {
            hls.destroy();
            hls = null;
        }
        
        // Initialize Player.js after iframe is mounted
        setTimeout(() => {
            initPlayerJs();
        }, 100);
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
                                    console.log('Max retries reached, trying direct video...');
                                    if (directVideoUrl) {
                                        useDirectVideo();
                                    } else {
                                        switchToIframe();
                                    }
                                }
                                break;
                            case Hls.ErrorTypes.MEDIA_ERROR:
                                console.log('Media error, attempting to recover...');
                                if (retryCount < MAX_RETRIES) {
                                    retryCount++;
                                    hls?.recoverMediaError();
                                } else {
                                    console.log('Max retries reached, trying direct video...');
                                    if (directVideoUrl) {
                                        useDirectVideo();
                                    } else {
                                        switchToIframe();
                                    }
                                }
                                break;
                            default:
                                console.log('Fatal error, trying direct video...');
                                if (directVideoUrl) {
                                    useDirectVideo();
                                } else {
                                    switchToIframe();
                                }
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
        // Priority order: Iframe -> HLS -> Direct MP4 (since videos are private)
        if (iframeSrc) {
            console.log('Using iframe playback (recommended for private videos)');
            switchToIframe();
        } else if (playbackUrl && playbackUrl.includes('playlist.m3u8')) {
            console.log('Trying HLS playback');
            initHls();
        } else if (directVideoUrl) {
            console.log('Trying direct MP4 playback:', directVideoUrl);
            useDirectVideo();
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

    function useDirectVideo() {
        if (videoElement && directVideoUrl) {
            console.log('Setting direct video source:', directVideoUrl);
            videoElement.src = directVideoUrl;
            isLoading = false;
            useIframe = false;
        }
    }
</script>

<div class="video-player">
    <div class="video-container">
        {#if useIframe && iframeSrc}
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
                    console.log('Iframe loaded successfully');
                }}
            ></iframe>
        {:else if videoId && (proxyUrl || directVideoUrl)}
            <video
                bind:this={videoElement}
                controls
                {poster}
                preload="auto"
                class="video-element"
                crossorigin="anonymous"
            >
                <track kind="captions" src="" srclang="en" label="English" default />
                Your browser does not support HTML video.
            </video>
        {/if}
        
        {#if isLoading}
            <div class="loading-indicator">
                <div class="spinner"></div>
                <p>Loading video...</p>
            </div>
        {/if}
        
        {#if errorMessage}
            <div class="error-message">
                {errorMessage}
                <div class="error-actions">
                    {#if directVideoUrl && !useIframe}
                        <button on:click={useDirectVideo} class="fallback-button">
                            Try Direct Video
                        </button>
                    {/if}
                    {#if iframeSrc && !useIframe}
                        <button on:click={switchToIframe} class="fallback-button">
                            Try Iframe Player
                        </button>
                    {/if}
                </div>
            </div>
        {/if}
    </div>
</div>

<style>
    .video-player {
        width: 100%;
        max-width: 100vw;
        max-height: 80vh;
        margin: 0 auto;
        background: #000;
        position: relative;
        padding: 0;
    }

    .video-container {
        width: 100%;
        height: 80vh;
        
        
    }

    .wrapper {
        border: 5px red solid !important;
    }

    .video-element {
        width: 100%;
        height: 100%;
        position: absolute;
        top: 0;
        left: 0;
        object-fit: contain; /* Ensure video fits within container */
    }

    .iframe-element {
        width: 100%;
        height: 100%;
        position: absolute;
        top: 0;
        left: 0;
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

    @keyframes spin {
        0% { transform: rotate(0deg); }
        100% { transform: rotate(360deg); }
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

    .error-actions {
        margin-top: 1rem;
        display: flex;
        gap: 1rem;
        justify-content: center;
        flex-wrap: wrap;
    }

    .fallback-button {
        padding: 0.5rem 1rem;
        background-color: #007bff;
        color: white;
        border: none;
        border-radius: 5px;
        cursor: pointer;
        font-size: 1rem;
        transition: background-color 0.2s;
    }

    .fallback-button:hover {
        background-color: #0056b3;
    }
</style> 