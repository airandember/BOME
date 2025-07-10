<script lang="ts">
	import { videoUtils } from '$lib/video';
	import type { Video } from '$lib/video';
	import LazyImage from './LazyImage.svelte';

	export let video: Video;
	export let showCategory: boolean = true;
	export let showStats: boolean = true;

	// Use thumbnail URL directly from backend response
	$: thumbnailSrc = video.thumbnailUrl || '/16X10_Placeholder_IMG.png';

	function handleClick() {
		// Navigate to video page
		window.location.href = `/videos/${video.id}`;
	}

	function handleKeyDown(event: KeyboardEvent) {
		if (event.key === 'Enter' || event.key === ' ') {
			event.preventDefault();
			handleClick();
		}
	}

	function getStatusColor(status: string): string {
		switch (status) {
			case 'ready':
				return '#4CAF50'; // Green
			case 'created':
			case 'uploaded':
				return '#2196F3'; // Blue
			case 'processing':
			case 'transcoding':
			case 'jit_segmenting':
				return '#FF9800'; // Orange
			case 'jit_playlists_created':
				return '#9C27B0'; // Purple
			case 'error':
			case 'upload_failed':
				return '#F44336'; // Red
			default:
				return '#9E9E9E'; // Gray
		}
	}

	function getStatusText(status: string): string {
		switch (status) {
			case 'ready':
				return 'Ready';
			case 'created':
				return 'Created';
			case 'uploaded':
				return 'Uploaded';
			case 'processing':
				return 'Processing';
			case 'transcoding':
				return 'Transcoding';
			case 'jit_segmenting':
				return 'Segmenting';
			case 'jit_playlists_created':
				return 'Finalizing';
			case 'error':
				return 'Error';
			case 'upload_failed':
				return 'Upload Failed';
			default:
				return 'Unknown';
		}
	}
</script>

<div 
	class="video-card" 
	on:click={handleClick} 
	on:keydown={handleKeyDown}
	role="button"
	tabindex="0"
	aria-label="Play video: {video.title}"
>
	<div class="thumbnail-container">
		<LazyImage
			src={thumbnailSrc}
			alt={video.title}
			className="thumbnail"
			width="100%"
			height="100%"
			loading="lazy"
			fallback="/16X10_Placeholder_IMG.png"
			placeholder="/16X10_Placeholder_IMG.png"
		/>
		<div class="duration-badge">
			{videoUtils.formatDuration(video.duration)}
		</div>
		{#if video.status && video.status !== 'ready'}
			<div class="status-badge" style="background-color: {getStatusColor(video.status)}">
				{getStatusText(video.status)}
			</div>
		{/if}
		<div class="play-overlay">
			<div class="play-icon">▶️</div>
		</div>
	</div>

	<div class="video-info">
		<h3 class="video-title" title={video.title}>
			{video.title}
		</h3>
		
		{#if showCategory && video.category}
			<div class="video-category">
				{video.category}
			</div>
		{/if}

		{#if showStats}
			<div class="video-stats">
				<span class="stat">
					👁️ {videoUtils.formatViewCount(video.viewCount)} views
				</span>
				<span class="stat">
					❤️ {videoUtils.formatViewCount(video.likeCount)} likes
				</span>
			</div>
		{/if}

		<div class="video-meta">
			<span class="upload-date">
				{new Date(video.createdAt).toLocaleDateString()}
			</span>
		</div>
	</div>
</div>

<style>
	.video-card {
		background: var(--card-bg);
		border-radius: 16px;
		overflow: hidden;
		cursor: pointer;
		transition: all 0.3s ease;
		box-shadow: 
			4px 4px 8px var(--shadow-dark),
			-2px -2px 4px var(--shadow-light);
	}

	.video-card:hover {
		transform: translateY(-4px);
		box-shadow: 
			8px 8px 16px var(--shadow-dark),
			-4px -4px 8px var(--shadow-light);
	}

	.thumbnail-container {
		position: relative;
		width: 100%;
		aspect-ratio: 16 / 9;
		overflow: hidden;
	}

	.duration-badge {
		position: absolute;
		bottom: 8px;
		right: 8px;
		background: rgba(0, 0, 0, 0.8);
		color: white;
		padding: 2px 6px;
		border-radius: 4px;
		font-size: 0.8rem;
		font-weight: 500;
	}

	.status-badge {
		position: absolute;
		top: 8px;
		right: 8px;
		color: white;
		padding: 2px 6px;
		border-radius: 4px;
		font-size: 0.7rem;
		font-weight: 500;
		text-transform: uppercase;
	}

	.play-overlay {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.3);
		display: flex;
		align-items: center;
		justify-content: center;
		opacity: 0;
		transition: opacity 0.3s ease;
	}

	.video-card:hover .play-overlay {
		opacity: 1;
	}

	.play-icon {
		font-size: 3rem;
		color: white;
		background: rgba(0, 0, 0, 0.7);
		border-radius: 50%;
		width: 60px;
		height: 60px;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.video-info {
		padding: 1rem;
	}

	.video-title {
		font-size: 1rem;
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 0.5rem 0;
		line-height: 1.4;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	.video-category {
		display: inline-block;
		background: var(--accent-color);
		color: white;
		padding: 0.25rem 0.5rem;
		border-radius: 12px;
		font-size: 0.75rem;
		font-weight: 500;
		margin-bottom: 0.5rem;
	}

	.video-stats {
		display: flex;
		gap: 1rem;
		margin-bottom: 0.5rem;
	}

	.stat {
		font-size: 0.8rem;
		color: var(--text-secondary);
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}

	.video-meta {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.upload-date {
		font-size: 0.8rem;
		color: var(--text-secondary);
	}

	@media (max-width: 768px) {
		.video-info {
			padding: 0.75rem;
		}

		.video-title {
			font-size: 0.9rem;
		}

		.video-stats {
			flex-direction: column;
			gap: 0.25rem;
		}
	}
</style> 