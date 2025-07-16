<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { videoService, type Video } from '$lib/video';
	import VideoPlayer from '$lib/components/VideoPlayer.svelte';
	import VideoCard from '$lib/components/VideoCard.svelte';
	import SubscriptionCheck from '$lib/components/SubscriptionCheck.svelte';
	import { toastStore } from '$lib/stores/toast';

	let video: Video | null = null;
	let relatedVideos: Video[] = [];
	let loading = true;
	let error = '';
	let showOverlay = false;
	let showSuggestedVideos = false;
	let overlayTimeout: NodeJS.Timeout;
	let mouseMovementTimeout: NodeJS.Timeout;
	let authChecking = true;

	onMount(() => {
		loadVideo();
		
		// Handle mouse movement for overlay
		const handleMouseMove = () => {
			showOverlay = true;
			clearTimeout(overlayTimeout);
			clearTimeout(mouseMovementTimeout);
			
			// Hide overlay after 3 seconds of no mouse movement
			mouseMovementTimeout = setTimeout(() => {
				showOverlay = false;
			}, 3000);
		};

		// Handle mouse leave
		const handleMouseLeave = () => {
			clearTimeout(mouseMovementTimeout);
			overlayTimeout = setTimeout(() => {
				showOverlay = false;
			}, 1000);
		};

		// Add event listeners
		document.addEventListener('mousemove', handleMouseMove);
		document.addEventListener('mouseleave', handleMouseLeave);

		// Cleanup
		return () => {
			document.removeEventListener('mousemove', handleMouseMove);
			document.removeEventListener('mouseleave', handleMouseLeave);
			clearTimeout(overlayTimeout);
			clearTimeout(mouseMovementTimeout);
		};
	});

	async function loadVideo() {
		try {
			const videoId = $page.params.id;
			const loadedVideo = await videoService.getVideo(videoId);
			video = loadedVideo;
			
			// Load related videos
			const relatedResponse = await videoService.getVideos(1, 6, loadedVideo.category);
			relatedVideos = relatedResponse.videos.filter(v => v.id !== loadedVideo.id);
		} catch (err) {
			error = 'Failed to load video. Please try again later.';
			toastStore.error('Error loading video');
		} finally {
			loading = false;
		}
	}

	function goBack() {
		goto('/videos');
	}

	function handleVideoEnd() {
		showSuggestedVideos = true;
	}

	function selectSuggestedVideo(selectedVideo: Video) {
		showSuggestedVideos = false;
		goto(`/videos/${selectedVideo.id}`);
	}

	function handleAuthLoadingChange(event: CustomEvent<{loading: boolean}>) {
		authChecking = event.detail.loading;
	}

	function handleAccessGranted() {
		// Load video data when access is granted
		if (!video && !loading) {
			loadVideo();
		}
	}
</script>

<SubscriptionCheck 
	redirectTo="/login" 
	requireSubscription={true}
	on:loadingChange={handleAuthLoadingChange}
	on:accessGranted={handleAccessGranted}
>
	<div class="fullscreen-video-page">
		{#if loading}
			<div class="loading-state">
				<div class="spinner"></div>
				<p>Loading video...</p>
			</div>
		{:else if error}
			<div class="error-state">
				<p>{error}</p>
				<button on:click={goBack} class="back-button">Back to Videos</button>
			</div>
		{:else if video}
			<!-- Fullscreen Video Player -->
			<div class="video-container">
				<VideoPlayer 
					videoId={video.bunnyVideoId} 
					title={video.title}
					poster={video.thumbnailUrl}
					playbackUrl={video.playbackUrl}
					iframeSrc={video.iframeSrc}
					on:ended={handleVideoEnd}
				/>
			</div>

			<!-- Overlay Controls -->
			<div class="overlay-controls" class:visible={showOverlay}>
				<!-- Back Button -->
				<button class="back-button-overlay" on:click={goBack}>
					<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M19 12H5M12 19l-7-7 7-7"/>
					</svg>
					Back
				</button>

				<!-- Video Title -->
				<div class="video-title-overlay">
					<h1 class='text-white'>{video.title}</h1>
					<div class="video-meta">
						<span class="views">{video.viewCount ? video.viewCount.toLocaleString() : '0'} views</span>
						<span class="date">{new Date(video.createdAt).toLocaleDateString()}</span>
						{#if video.duration}
							<span class="duration">{video.duration} min</span>
						{/if}
					</div>
				</div>
			</div>

			<!-- Suggested Videos Overlay (appears when video ends) -->
			{#if showSuggestedVideos && relatedVideos.length > 0}
				<div class="suggested-videos-overlay">
					<div class="suggested-videos-container">
						<h2>Watch Next</h2>
						<div class="suggested-videos-grid">
							{#each relatedVideos.slice(0, 4) as suggestedVideo}
								<div class="suggested-video-card" on:click={() => selectSuggestedVideo(suggestedVideo)}>
									<img src={suggestedVideo.thumbnailUrl} alt={suggestedVideo.title} />
									<div class="suggested-video-info">
										<h3>{suggestedVideo.title}</h3>
										<p class="suggested-video-meta">
											{suggestedVideo.viewCount ? suggestedVideo.viewCount.toLocaleString() : '0'} views
										</p>
									</div>
								</div>
							{/each}
						</div>
						<button class="close-suggestions" on:click={() => showSuggestedVideos = false}>
							<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<line x1="18" y1="6" x2="6" y2="18"></line>
								<line x1="6" y1="6" x2="18" y2="18"></line>
							</svg>
						</button>
					</div>
				</div>
			{/if}
		{/if}
	</div>
</SubscriptionCheck>

<style lang="postcss">
	.fullscreen-video-page {
		position: fixed;
		top: 0;
		left: 0;
		width: 100vw;
		height: 100vh;
		background: #000;
		color: #fff;
		overflow: hidden;
		z-index: 1000;
	}

	.video-container {
		width: 100%;
		height: 100%;
		position: relative;
	}

	.overlay-controls {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		pointer-events: none;
		opacity: 0;
		transition: opacity 0.3s ease;
		z-index: 10;
	}

	.overlay-controls.visible {
		opacity: 1;
	}

	.back-button-overlay {
		position: absolute;
		top: 2rem;
		left: 2rem;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		background: rgba(0, 0, 0, 0.7);
		color: #fff;
		border: none;
		padding: 0.75rem 1rem;
		border-radius: 8px;
		cursor: pointer;
		font-size: 1rem;
		font-weight: 500;
		transition: background 0.2s;
		pointer-events: auto;
		backdrop-filter: blur(10px);
	}

	.back-button-overlay:hover {
		background: rgba(0, 0, 0, 0.9);
	}

	.back-button-overlay svg {
		width: 20px;
		height: 20px;
	}

	.video-title-overlay {
		position: absolute;
		bottom: 6rem;
		padding: 2rem;
		width: 100vw;
		background: linear-gradient(transparent, rgba(0, 0, 0, 0.8));
		padding: 2rem 2rem 1rem 2rem;
		pointer-events: none;
	}

	.video-title-overlay h1 {
		color: white;
		font-size: 2.5rem;
		font-weight: 600;
		margin: 0 0 0.5rem 0;
		text-shadow: 0 2px 4px rgba(0, 0, 0, 0.8);
	}

	.video-meta {
		display: flex;
		gap: 1rem;
		color: #a3a3a3;
		font-size: 0.9rem;
		text-shadow: 0 1px 2px rgba(0, 0, 0, 0.8);
	}

	.suggested-videos-overlay {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.9);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 20;
		backdrop-filter: blur(10px);
	}

	.suggested-videos-container {
		max-width: 1200px;
		width: 90%;
		position: relative;
	}

	.suggested-videos-container h2 {
		font-size: 2rem;
		font-weight: 600;
		margin-bottom: 2rem;
		text-align: center;
	}

	.suggested-videos-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 1.5rem;
		margin-bottom: 2rem;
	}

	.suggested-video-card {
		background: rgba(255, 255, 255, 0.1);
		border-radius: 12px;
		overflow: hidden;
		cursor: pointer;
		transition: transform 0.2s, background 0.2s;
		backdrop-filter: blur(10px);
	}

	.suggested-video-card:hover {
		transform: translateY(-4px);
		background: rgba(255, 255, 255, 0.15);
	}

	.suggested-video-card img {
		width: 100%;
		height: 140px;
		object-fit: cover;
	}

	.suggested-video-info {
		padding: 1rem;
	}

	.suggested-video-info h3 {
		font-size: 1.1rem;
		font-weight: 500;
		margin: 0 0 0.5rem 0;
		line-height: 1.3;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	.suggested-video-meta {
		color: #a3a3a3;
		font-size: 0.85rem;
		margin: 0;
	}

	.close-suggestions {
		position: absolute;
		top: 1rem;
		right: 1rem;
		background: rgba(255, 255, 255, 0.1);
		border: none;
		color: #fff;
		width: 40px;
		height: 40px;
		border-radius: 50%;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: background 0.2s;
		backdrop-filter: blur(10px);
	}

	.close-suggestions:hover {
		background: rgba(255, 255, 255, 0.2);
	}

	.loading-state,
	.error-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100%;
		text-align: center;
		color: #a3a3a3;
	}

	.spinner {
		width: 40px;
		height: 40px;
		border: 4px solid #333;
		border-top-color: #e50914;
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-bottom: 1rem;
	}

	.back-button {
		display: inline-block;
		margin-top: 1rem;
		padding: 0.75rem 1.5rem;
		background: #e50914;
		color: #fff;
		border: none;
		border-radius: 4px;
		cursor: pointer;
		transition: background 0.2s;
	}

	.back-button:hover {
		background: #f40612;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	@media (max-width: 768px) {
		.video-title-overlay {
			bottom: 6rem;
			left: 1rem;
			right: 1rem;
		}

		.video-title-overlay h1 {
			font-size: 1.8rem;
		}

		.back-button-overlay {
			top: 1rem;
			left: 1rem;
			padding: 0.5rem 0.75rem;
			font-size: 0.9rem;
		}

		.suggested-videos-grid {
			grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
			gap: 1rem;
		}

		.suggested-videos-container h2 {
			font-size: 1.5rem;
		}
	}
</style> 