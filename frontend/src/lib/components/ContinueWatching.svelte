<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { videoAnalytics } from '$lib/services/videoAnalytics';
	import type { ContinueWatchingVideo } from '$lib/services/videoAnalytics';
	
	// Props
	export let limit: number = 10;
	export let showTitle: boolean = true;
	
	// State
	let videos: ContinueWatchingVideo[] = [];
	let isLoading = true;
	let error: string | null = null;
	
	onMount(async () => {
		await loadVideos();
	});
	
	async function loadVideos() {
		isLoading = true;
		error = null;
		
		try {
			videos = await videoAnalytics.getContinueWatching(limit);
			console.log(`📺 [Continue Watching] Loaded ${videos.length} videos`);
		} catch (err) {
			console.error('❌ [Continue Watching] Failed to load:', err);
			error = 'Failed to load continue watching list';
		} finally {
			isLoading = false;
		}
	}
	
	function handleVideoClick(videoId: number) {
		goto(`/videos/${videoId}`);
	}
	
	function formatTime(seconds: number): string {
		const mins = Math.floor(seconds / 60);
		const secs = Math.floor(seconds % 60);
		return `${mins}:${secs.toString().padStart(2, '0')}`;
	}
	
	function formatDuration(seconds: number): string {
		const hours = Math.floor(seconds / 3600);
		const mins = Math.floor((seconds % 3600) / 60);
		
		if (hours > 0) {
			return `${hours}h ${mins}m`;
		}
		return `${mins}m`;
	}
	
	function getProgressColor(percentage: number): string {
		if (percentage < 25) return '#ef4444'; // Red
		if (percentage < 50) return '#f59e0b'; // Orange
		if (percentage < 75) return '#eab308'; // Yellow
		return '#22c55e'; // Green
	}
	
	function formatRelativeTime(dateString: string): string {
		const date = new Date(dateString);
		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffMins = Math.floor(diffMs / 60000);
		const diffHours = Math.floor(diffMs / 3600000);
		const diffDays = Math.floor(diffMs / 86400000);
		
		if (diffMins < 1) return 'just now';
		if (diffMins < 60) return `${diffMins} min ago`;
		if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
		if (diffDays < 7) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
		if (diffDays < 30) return `${Math.floor(diffDays / 7)} week${Math.floor(diffDays / 7) > 1 ? 's' : ''} ago`;
		return date.toLocaleDateString();
	}
</script>

{#if showTitle}
	<div class="section-header">
		<h2>Continue Watching</h2>
		<p class="subtitle">Pick up where you left off</p>
	</div>
{/if}

<div class="continue-watching-container">
	{#if isLoading}
		<div class="loading">
			<div class="spinner"></div>
			<p>Loading your videos...</p>
		</div>
	{:else if error}
		<div class="error">
			<p>{error}</p>
			<button on:click={loadVideos} class="retry-btn">Retry</button>
		</div>
	{:else if videos.length === 0}
		<div class="empty-state">
			<div class="empty-icon">📺</div>
			<h3>No videos in progress</h3>
			<p>Start watching a video to see it here!</p>
		</div>
	{:else}
		<div class="video-grid">
			{#each videos as video (video.video_id)}
				<div class="video-card" on:click={() => handleVideoClick(video.video_id)} on:keypress={(e) => e.key === 'Enter' && handleVideoClick(video.video_id)} role="button" tabindex="0">
					<div class="thumbnail-container">
						<img 
							src={video.thumbnail_url || '/placeholder-video.png'} 
							alt={video.title}
							class="thumbnail"
						/>
						<div class="progress-overlay">
							<div class="progress-bar">
								<div 
									class="progress-fill" 
									style="width: {video.percentage}%; background-color: {getProgressColor(video.percentage)};"
								></div>
							</div>
						</div>
						<div class="resume-badge">
							<svg width="16" height="16" viewBox="0 0 16 16" fill="none">
								<path d="M3 2L13 8L3 14V2Z" fill="currentColor"/>
							</svg>
							<span>{Math.round(video.percentage)}%</span>
						</div>
						<div class="duration-badge">
							{formatDuration(video.duration)}
						</div>
					</div>
					<div class="video-info">
						<h3 class="video-title">{video.title}</h3>
						<p class="resume-text">
							Resume at {formatTime(video.last_position)}
						</p>
						<p class="watch-time">
							Last watched {formatRelativeTime(video.last_watched_at)}
						</p>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
	.section-header {
		margin-bottom: 1.5rem;
	}
	
	.section-header h2 {
		font-size: 1.75rem;
		font-weight: 700;
		color: #1a1a1a;
		margin: 0 0 0.5rem 0;
	}
	
	.subtitle {
		color: #666;
		margin: 0;
		font-size: 0.95rem;
	}
	
	.continue-watching-container {
		width: 100%;
	}
	
	.loading, .error, .empty-state {
		text-align: center;
		padding: 3rem 1rem;
	}
	
	.spinner {
		width: 40px;
		height: 40px;
		border: 3px solid #f0f0f0;
		border-top-color: #d4a574;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
		margin: 0 auto 1rem;
	}
	
	@keyframes spin {
		to { transform: rotate(360deg); }
	}
	
	.retry-btn {
		margin-top: 1rem;
		padding: 0.5rem 1.5rem;
		background: #d4a574;
		color: white;
		border: none;
		border-radius: 6px;
		cursor: pointer;
		font-weight: 600;
	}
	
	.empty-state {
		background: #f9f9f9;
		border-radius: 12px;
		padding: 3rem 2rem;
	}
	
	.empty-icon {
		font-size: 4rem;
		margin-bottom: 1rem;
	}
	
	.empty-state h3 {
		margin: 0 0 0.5rem 0;
		color: #333;
		font-size: 1.25rem;
	}
	
	.empty-state p {
		color: #666;
		margin: 0;
	}
	
	.video-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 1.5rem;
	}
	
	.video-card {
		background: #fff;
		border-radius: 12px;
		overflow: hidden;
		cursor: pointer;
		transition: all 0.3s ease;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
	}
	
	.video-card:hover {
		transform: translateY(-4px);
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
	}
	
	.thumbnail-container {
		position: relative;
		padding-top: 56.25%; /* 16:9 aspect ratio */
		background: #000;
		overflow: hidden;
	}
	
	.thumbnail {
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	
	.progress-overlay {
		position: absolute;
		bottom: 0;
		left: 0;
		right: 0;
		background: linear-gradient(to top, rgba(0,0,0,0.8), transparent);
		padding: 1rem 0.5rem 0.5rem;
	}
	
	.progress-bar {
		width: 100%;
		height: 4px;
		background: rgba(255, 255, 255, 0.3);
		border-radius: 2px;
		overflow: hidden;
	}
	
	.progress-fill {
		height: 100%;
		transition: width 0.3s ease;
		border-radius: 2px;
	}
	
	.resume-badge {
		position: absolute;
		top: 0.75rem;
		left: 0.75rem;
		background: rgba(212, 165, 116, 0.95);
		color: white;
		padding: 0.25rem 0.75rem;
		border-radius: 20px;
		font-size: 0.85rem;
		font-weight: 600;
		display: flex;
		align-items: center;
		gap: 0.25rem;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
	}
	
	.duration-badge {
		position: absolute;
		bottom: 3rem;
		right: 0.75rem;
		background: rgba(0, 0, 0, 0.8);
		color: white;
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-size: 0.75rem;
		font-weight: 600;
	}
	
	.video-info {
		padding: 1rem;
	}
	
	.video-title {
		margin: 0 0 0.5rem 0;
		font-size: 1rem;
		font-weight: 600;
		color: #1a1a1a;
		line-height: 1.4;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	
	.resume-text {
		margin: 0;
		font-size: 0.9rem;
		color: #d4a574;
		font-weight: 600;
	}
	
	.watch-time {
		margin: 0.25rem 0 0 0;
		font-size: 0.85rem;
		color: #999;
	}
	
	@media (max-width: 768px) {
		.video-grid {
			grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
			gap: 1rem;
		}
		
		.section-header h2 {
			font-size: 1.5rem;
		}
	}
	
	@media (max-width: 480px) {
		.video-grid {
			grid-template-columns: 1fr;
		}
	}
</style>

