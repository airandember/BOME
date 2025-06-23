<script lang="ts">
	import { onMount } from 'svelte';
	import { youtubeVideos, youtubeLoading, youtubeError, youtubeActions } from '$lib/stores/youtube';
	import LazyImage from '$lib/components/LazyImage.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import type { YouTubeVideo } from '$lib/types/youtube';

	let searchTerm = '';
	let filteredVideos: YouTubeVideo[] = [];

	// Reactive statement to filter videos based on search
	$: {
		if (searchTerm.trim() === '') {
			filteredVideos = $youtubeVideos;
		} else {
			filteredVideos = $youtubeVideos.filter(video =>
				video.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
				video.description.toLowerCase().includes(searchTerm.toLowerCase())
			);
		}
	}

	// Format date for display
	function formatDate(dateString: string): string {
		const date = new Date(dateString);
		return date.toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	// Format duration for display
	function formatDuration(duration?: string): string {
		if (!duration) return '';
		return duration;
	}

	// Format view count
	function formatViewCount(count?: number): string {
		if (!count) return '';
		if (count >= 1000000) {
			return `${(count / 1000000).toFixed(1)}M views`;
		} else if (count >= 1000) {
			return `${(count / 1000).toFixed(1)}K views`;
		}
		return `${count} views`;
	}

	// Open video in new tab
	function openVideo(video: YouTubeVideo) {
		window.open(video.video_url, '_blank');
	}

	onMount(() => {
		// Refresh videos when page loads
		youtubeActions.fetchLatestVideos();
	});
</script>

<svelte:head>
	<title>Latest YouTube Videos - Book of Mormon Evidence</title>
	<meta name="description" content="Watch the latest videos from the Book of Mormon Evidence YouTube channel. Explore archaeological evidence, historical insights, and scholarly research." />
</svelte:head>

<div class="youtube-page">
	<!-- Header Section -->
	<div class="hero-section">
		<div class="container">
			<div class="hero-content">
				<h1>Latest YouTube Videos</h1>
				<p class="hero-description">
					Watch the latest videos from the <strong>Book of Mormon Evidence</strong> YouTube channel. 
					Explore archaeological evidence, historical insights, and scholarly research that supports 
					the Book of Mormon narrative.
				</p>
				<div class="channel-link">
					<a 
						href="https://www.youtube.com/@BookofMormonEvidence" 
						target="_blank" 
						rel="noopener noreferrer"
						class="btn btn-youtube"
					>
						<svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
							<path d="M23.498 6.186a3.016 3.016 0 0 0-2.122-2.136C19.505 3.545 12 3.545 12 3.545s-7.505 0-9.377.505A3.017 3.017 0 0 0 .502 6.186C0 8.07 0 12 0 12s0 3.93.502 5.814a3.016 3.016 0 0 0 2.122 2.136c1.871.505 9.376.505 9.376.505s7.505 0 9.377-.505a3.015 3.015 0 0 0 2.122-2.136C24 15.93 24 12 24 12s0-3.93-.502-5.814zM9.545 15.568V8.432L15.818 12l-6.273 3.568z"/>
						</svg>
						Visit YouTube Channel
					</a>
				</div>
			</div>
		</div>
	</div>

	<div class="container">
		<!-- Search Section -->
		<div class="search-section">
			<div class="search-box">
				<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="11" cy="11" r="8"></circle>
					<path d="m21 21-4.35-4.35"></path>
				</svg>
				<input
					type="text"
					placeholder="Search videos..."
					bind:value={searchTerm}
					class="search-input"
				/>
			</div>
		</div>

		<!-- Loading State -->
		{#if $youtubeLoading}
			<div class="loading-section">
				<LoadingSpinner size="large" />
				<p>Loading latest videos...</p>
			</div>
		{/if}

		<!-- Error State -->
		{#if $youtubeError}
			<div class="error-section">
				<div class="error-card">
					<h3>Unable to load videos</h3>
					<p>{$youtubeError}</p>
					<button 
						class="btn btn-primary"
						on:click={() => youtubeActions.fetchLatestVideos()}
					>
						Try Again
					</button>
				</div>
			</div>
		{/if}

		<!-- Videos Grid -->
		{#if !$youtubeLoading && !$youtubeError && filteredVideos.length > 0}
			<div class="videos-section">
				<div class="section-header">
					<h2>
						{searchTerm ? `Search Results (${filteredVideos.length})` : `Latest Videos (${filteredVideos.length})`}
					</h2>
					{#if searchTerm}
						<button 
							class="btn btn-secondary"
							on:click={() => searchTerm = ''}
						>
							Clear Search
						</button>
					{/if}
				</div>

				<div class="videos-grid">
					{#each filteredVideos as video (video.id)}
						<div class="video-card" on:click={() => openVideo(video)} on:keydown>
							<div class="video-thumbnail">
								<LazyImage
									src={video.thumbnail_url}
									alt={video.title}
									cacheKey={`youtube-thumb-${video.id}`}
									class="thumbnail-image"
								/>
								<div class="play-overlay">
									<svg width="48" height="48" viewBox="0 0 24 24" fill="white">
										<path d="M8 5v14l11-7z"/>
									</svg>
								</div>
								{#if video.duration}
									<div class="duration-badge">
										{formatDuration(video.duration)}
									</div>
								{/if}
							</div>
							
							<div class="video-info">
								<h3 class="video-title">{video.title}</h3>
								<div class="video-meta">
									<span class="publish-date">{formatDate(video.published_at)}</span>
									{#if video.view_count}
										<span class="view-count">{formatViewCount(video.view_count)}</span>
									{/if}
								</div>
								{#if video.description}
									<p class="video-description">
										{video.description.length > 120 
											? video.description.substring(0, 120) + '...' 
											: video.description}
									</p>
								{/if}
							</div>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Empty State -->
		{#if !$youtubeLoading && !$youtubeError && filteredVideos.length === 0}
			<div class="empty-section">
				<div class="empty-card">
					{#if searchTerm}
						<h3>No videos found</h3>
						<p>No videos match your search term "{searchTerm}"</p>
						<button 
							class="btn btn-primary"
							on:click={() => searchTerm = ''}
						>
							Clear Search
						</button>
					{:else}
						<h3>No videos available</h3>
						<p>We're working on getting the latest videos from our YouTube channel.</p>
						<button 
							class="btn btn-primary"
							on:click={() => youtubeActions.fetchLatestVideos()}
						>
							Refresh
						</button>
					{/if}
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	.youtube-page {
		min-height: 100vh;
		background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
	}

	.hero-section {
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
		color: white;
		padding: 4rem 0;
		text-align: center;
	}

	.hero-content h1 {
		font-size: 3rem;
		font-weight: 700;
		margin-bottom: 1rem;
		text-shadow: 0 2px 4px rgba(0,0,0,0.1);
	}

	.hero-description {
		font-size: 1.2rem;
		max-width: 800px;
		margin: 0 auto 2rem;
		line-height: 1.6;
		opacity: 0.95;
	}

	.btn-youtube {
		background: #ff0000;
		color: white;
		border: none;
		padding: 12px 24px;
		border-radius: 8px;
		font-weight: 600;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		gap: 8px;
		transition: all 0.3s ease;
	}

	.btn-youtube:hover {
		background: #cc0000;
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(255,0,0,0.3);
	}

	.container {
		max-width: 1200px;
		margin: 0 auto;
		padding: 0 2rem;
	}

	.search-section {
		padding: 2rem 0;
	}

	.search-box {
		position: relative;
		max-width: 500px;
		margin: 0 auto;
	}

	.search-box svg {
		position: absolute;
		left: 12px;
		top: 50%;
		transform: translateY(-50%);
		color: #666;
	}

	.search-input {
		width: 100%;
		padding: 12px 12px 12px 44px;
		border: 2px solid #e2e8f0;
		border-radius: 12px;
		font-size: 1rem;
		background: white;
		transition: all 0.3s ease;
	}

	.search-input:focus {
		outline: none;
		border-color: #667eea;
		box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
	}

	.loading-section, .error-section, .empty-section {
		text-align: center;
		padding: 4rem 2rem;
	}

	.error-card, .empty-card {
		background: white;
		padding: 2rem;
		border-radius: 12px;
		box-shadow: 0 4px 6px rgba(0,0,0,0.1);
		max-width: 400px;
		margin: 0 auto;
	}

	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 2rem;
	}

	.section-header h2 {
		font-size: 2rem;
		font-weight: 600;
		color: #2d3748;
	}

	.videos-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
		gap: 2rem;
		padding-bottom: 4rem;
	}

	.video-card {
		background: white;
		border-radius: 12px;
		overflow: hidden;
		box-shadow: 0 4px 6px rgba(0,0,0,0.1);
		transition: all 0.3s ease;
		cursor: pointer;
	}

	.video-card:hover {
		transform: translateY(-4px);
		box-shadow: 0 8px 25px rgba(0,0,0,0.15);
	}

	.video-thumbnail {
		position: relative;
		aspect-ratio: 16/9;
		overflow: hidden;
	}

	:global(.video-card .thumbnail-image) {
		width: 100%;
		height: 100%;
		object-fit: cover;
		transition: transform 0.3s ease;
	}

	.video-card:hover :global(.thumbnail-image) {
		transform: scale(1.05);
	}

	.play-overlay {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		background: rgba(0,0,0,0.7);
		border-radius: 50%;
		width: 80px;
		height: 80px;
		display: flex;
		align-items: center;
		justify-content: center;
		opacity: 0;
		transition: opacity 0.3s ease;
	}

	.video-card:hover .play-overlay {
		opacity: 1;
	}

	.duration-badge {
		position: absolute;
		bottom: 8px;
		right: 8px;
		background: rgba(0,0,0,0.8);
		color: white;
		padding: 4px 8px;
		border-radius: 4px;
		font-size: 0.8rem;
		font-weight: 600;
	}

	.video-info {
		padding: 1.5rem;
	}

	.video-title {
		font-size: 1.1rem;
		font-weight: 600;
		color: #2d3748;
		margin-bottom: 0.5rem;
		line-height: 1.4;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	.video-meta {
		display: flex;
		gap: 1rem;
		margin-bottom: 0.5rem;
		font-size: 0.9rem;
		color: #666;
	}

	.video-description {
		color: #666;
		font-size: 0.9rem;
		line-height: 1.5;
		margin: 0;
	}

	.btn {
		padding: 8px 16px;
		border-radius: 6px;
		font-weight: 500;
		text-decoration: none;
		border: none;
		cursor: pointer;
		transition: all 0.3s ease;
		display: inline-block;
	}

	.btn-primary {
		background: #667eea;
		color: white;
	}

	.btn-primary:hover {
		background: #5a67d8;
	}

	.btn-secondary {
		background: #e2e8f0;
		color: #4a5568;
	}

	.btn-secondary:hover {
		background: #cbd5e0;
	}

	@media (max-width: 768px) {
		.hero-content h1 {
			font-size: 2rem;
		}

		.hero-description {
			font-size: 1rem;
		}

		.videos-grid {
			grid-template-columns: 1fr;
			gap: 1.5rem;
		}

		.section-header {
			flex-direction: column;
			gap: 1rem;
			text-align: center;
		}

		.container {
			padding: 0 1rem;
		}
	}
</style> 