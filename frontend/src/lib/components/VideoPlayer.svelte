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
            
            if (Hls.isSupported()) {
                console.log('HLS.js is supported');
                
                // Destroy existing instance if any
                if (hls) {
                    hls.destroy();
                }

                hls = new Hls({
                    debug: true,
                    enableWorker: true,
                    maxLoadingDelay: 4000,
                    xhrSetup: function(xhr, url) {
                        if (token) {
                            console.log('Adding auth token to request:', url);
                            xhr.setRequestHeader('Authorization', `Bearer ${token}`);
                            xhr.setRequestHeader('Accept', 'application/x-mpegURL,*/*');
                            xhr.withCredentials = false;
                        }
                    }
                });
                
                hls.on(Hls.Events.ERROR, function(event, data) {
                    console.error('HLS error:', event, data);
                    if (data.fatal) {
                        switch (data.type) {
                            case Hls.ErrorTypes.NETWORK_ERROR:
                                console.error('Network error:', data.details, data.response);
                                
                                if (retryCount < MAX_RETRIES) {
                                    console.log(`Retrying HLS (${retryCount + 1}/${MAX_RETRIES})...`);
                                    retryCount++;
                                    setTimeout(() => {
                                        if (hls) {
                                            console.log('Reloading HLS stream...');
                                            hls.loadSource(proxyUrl);
                                        }
                                    }, 1000 * retryCount);
                                } else {
                                    console.log('Max HLS retries reached, switching to iframe');
                                    switchToIframe();
                                }
                                break;
                            case Hls.ErrorTypes.MEDIA_ERROR:
                                console.error('Media error:', data.details);
                                hls?.recoverMediaError();
                                break;
                            default:
                                console.error('Fatal HLS error:', data.details);
                                switchToIframe();
                                break;
                        }
                    }
                });

                hls.on(Hls.Events.MANIFEST_PARSED, function() {
                    console.log('HLS manifest parsed, attempting playback');
                    videoElement.play().catch(e => {
                        console.error('HLS playback failed:', e);
                        switchToIframe();
                    });
                });

                hls.on(Hls.Events.LEVEL_LOADED, function() {
                    console.log('HLS level loaded successfully');
                    errorMessage = '';
                    retryCount = 0;
                });

                hls.loadSource(proxyUrl);
                hls.attachMedia(videoElement);
            }
            // For Safari that has built-in HLS support
            else if (videoElement.canPlayType('application/vnd.apple.mpegurl')) {
                console.log('Using native HLS support');
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
                        console.error('Native HLS error:', e);
                        switchToIframe();
                    });
                
                videoElement.onerror = function(e) {
                    console.error('Video playback error:', e);
                    switchToIframe();
                };
            } else {
                // Browser doesn't support HLS, use iframe
                console.log('HLS not supported, using iframe');
                switchToIframe();
            }

            if (!useIframe) {
                videoElement.addEventListener('play', () => {
                    errorMessage = '';
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