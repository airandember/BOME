<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { videoService, type Video } from '$lib/video';
	import VideoPlayer from '$lib/components/VideoPlayer.svelte';
	import VideoCard from '$lib/components/VideoCard.svelte';
	import { toastStore } from '$lib/stores/toast';

	let video: Video | null = null;
	let relatedVideos: Video[] = [];
	let loading = true;
	let error = '';

	onMount(() => {
		loadVideo();
	});

	async function loadVideo() {
		try {
			const videoId = parseInt($page.params.id);
			const loadedVideo = await videoService.getVideo(videoId);
			video = loadedVideo;
			
			// Load related videos
			const relatedResponse = await videoService.getVideos(1, 5, loadedVideo.category);
			relatedVideos = relatedResponse.videos.filter(v => v.id !== loadedVideo.id);
		} catch (err) {
			error = 'Failed to load video. Please try again later.';
			toastStore.error('Error loading video');
		} finally {
			loading = false;
		}
	}
</script>

<div class="video-page">
	{#if loading}
		<div class="loading-state">
			<div class="spinner"></div>
			<p>Loading video...</p>
		</div>
	{:else if error}
		<div class="error-state">
			<p>{error}</p>
			<a href="/videos" class="back-link">Back to Videos</a>
		</div>
	{:else if video}
		<div class="video-section">
			<VideoPlayer 
				videoId={video.bunnyVideoId} 
				title={video.title}
				poster={video.thumbnailUrl}
				playbackUrl={video.playData?.playbackUrl}
				autoplay={true}
			/>
		</div>

		<div class="content-section">
			<div class="video-info">
				<h1>{video.title}</h1>
				{#if video.description}
					<p class="description">{video.description}</p>
				{/if}
				<div class="metadata">
					<span class="views">{video.viewCount ? video.viewCount.toLocaleString() : '0'} views</span>
					<span class="date">
						{new Date(video.createdAt).toLocaleDateString()}
					</span>
					{#if video.duration}
						<span class="duration">{video.duration} min</span>
					{/if}
				</div>
			</div>

			{#if relatedVideos.length > 0}
				<div class="related-videos">
					<h2>More Like This</h2>
					<div class="video-grid">
						{#each relatedVideos as relatedVideo}
							<VideoCard video={relatedVideo} />
						{/each}
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>

<style lang="postcss">
	.video-page {
		margin-top: 50px;
		min-height: calc(100vh - 50px);
		background: #141414;
		color: #fff;
	}

	.video-section {
		width: 100%;
		background: #000;
		position: relative;
	}

	.content-section {
		max-width: 1200px;
		margin: 0 auto;
		padding: 2rem;
	}

	.video-info {
		margin-bottom: 3rem;
	}

	h1 {
		font-size: 2.5rem;
		font-weight: 600;
		margin-bottom: 1rem;
		color: #fff;
	}

	.description {
		font-size: 1.1rem;
		line-height: 1.6;
		color: #a3a3a3;
		margin-bottom: 1.5rem;
		max-width: 800px;
	}

	.metadata {
		display: flex;
		gap: 1.5rem;
		color: #737373;
		font-size: 0.9rem;
	}

	.related-videos {
		margin-top: 4rem;
	}

	.related-videos h2 {
		font-size: 1.8rem;
		font-weight: 500;
		margin-bottom: 1.5rem;
		color: #fff;
	}

	.video-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
		gap: 1.5rem;
	}

	.loading-state,
	.error-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 50vh;
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

	.back-link {
		display: inline-block;
		margin-top: 1rem;
		padding: 0.75rem 1.5rem;
		background: #e50914;
		color: #fff;
		text-decoration: none;
		border-radius: 4px;
		transition: background 0.2s;
	}

	.back-link:hover {
		background: #f40612;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	@media (max-width: 768px) {
		.content-section {
			padding: 1rem;
		}

		h1 {
			font-size: 1.8rem;
		}

		.video-grid {
			grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		}
	}
</style> 