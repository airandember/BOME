<script lang="ts">
	import { onMount } from 'svelte';
	import { youtubeStore, youtubeActions, youtubeUtils } from '$lib/stores/youtube';
	import Navigation from '$lib/components/Navigation.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import LazyImage from '$lib/components/LazyImage.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import type { YouTubeVideo, ChannelInfo } from '$lib/types/youtube';

	let searchTerm = '';
	let selectedCategory = '';
	let filteredVideos: YouTubeVideo[] = [];

	// Subscribe to store
	$: state = $youtubeStore;
	$: videos = state.videos;
	$: searchResults = state.searchResults;
	$: categories = state.categories;
	$: channelInfo = state.channelInfo;
	$: loading = state.loading;
	$: error = state.error;

	// Reactive statement to determine which videos to show
	$: {
		if (searchTerm.trim() !== '') {
			filteredVideos = searchResults;
		} else if (selectedCategory !== '') {
			filteredVideos = videos.filter(video => 
				//video.tags?.some(tag => tag.toLowerCase().includes(selectedCategory.toLowerCase())) ||
				//video.category?.toLowerCase().includes(selectedCategory.toLowerCase()) ||
				video.title.toLowerCase().includes(selectedCategory.toLowerCase()) ||
				video.description.toLowerCase().includes(selectedCategory.toLowerCase())
			);
		} else {
			filteredVideos = videos;
		}
	}

	// Handle search
	async function handleSearch() {
		if (searchTerm.trim() === '') {
			youtubeActions.clearSearch();
			return;
		}
		await youtubeActions.searchVideos(searchTerm.trim(), 20);
	}

	// Handle category filter
	async function handleCategoryFilter(category: string) {
		selectedCategory = category;
		if (category === '') {
			// Show all videos
			await youtubeActions.getAllVideos(20);
		} else {
			// Filter by category
			await youtubeActions.getVideosByCategory(category, 20);
		}
	}

	// Clear all filters
	function clearFilters() {
		searchTerm = '';
		selectedCategory = '';
		youtubeActions.clearSearch();
		youtubeActions.getAllVideos(20);
	}

	// Open video in new tab
	function openVideo(video: YouTubeVideo) {
		window.open(video.video_url, '_blank');
	}

	// Format date for display
	function formatDate(dateString: string): string {
		return youtubeUtils.formatPublishedDate(dateString);
	}

	// Format duration for display
	function formatDuration(duration: string): string {
		return youtubeUtils.formatDuration(duration);
	}

	// Format view count
	function formatViewCount(count: number): string {
		return youtubeUtils.formatViewCount(count);
	}

	// Get optimized thumbnail
	function getThumbnail(video: YouTubeVideo): string {
		return youtubeUtils.getThumbnail(video, 'medium');
	}

	onMount(async () => {
		// Initialize the store with all data
		await youtubeActions.initialize();
	});
</script>

<svelte:head>
	<title>Latest YouTube Videos - Book of Mormon Evidence</title>
	<meta name="description" content="Watch the latest videos from the Book of Mormon Evidence YouTube channel. Explore archaeological evidence, historical insights, and scholarly research." />
</svelte:head>

<Navigation />

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
				
				{#if channelInfo}
					<div class="channel-stats">
						<div class="stat">
							<span class="stat-number">{channelInfo.subscriber_count.toLocaleString()}</span>
							<span class="stat-label">Subscribers</span>
						</div>
						<div class="stat">
							<span class="stat-number">{channelInfo.video_count.toLocaleString()}</span>
							<span class="stat-label">Videos</span>
						</div>
						<div class="stat">
							<span class="stat-number">{channelInfo.view_count.toLocaleString()}</span>
							<span class="stat-label">Total Views</span>
						</div>
					</div>
				{/if}

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
		<!-- Search and Filter Section -->
		<div class="search-section">
			<div class="search-controls">
				<div class="search-box">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="11" cy="11" r="8"></circle>
						<path d="m21 21-4.35-4.35"></path>
					</svg>
					<input
						type="text"
						placeholder="Search videos..."
						bind:value={searchTerm}
						on:input={handleSearch}
						class="search-input"
					/>
					{#if searchTerm}
						<button 
							class="clear-search"
							on:click={() => { searchTerm = ''; youtubeActions.clearSearch(); }}
						>
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<line x1="18" y1="6" x2="6" y2="18"></line>
								<line x1="6" y1="6" x2="18" y2="18"></line>
							</svg>
						</button>
					{/if}
				</div>

				{#if categories.length > 0}
					<div class="category-filters">
						<button 
							class="category-btn {selectedCategory === '' ? 'active' : ''}"
							on:click={() => handleCategoryFilter('')}
						>
							All Videos
						</button>
						{#each categories as category}
							<button 
								class="category-btn {selectedCategory === category ? 'active' : ''}"
								on:click={() => handleCategoryFilter(category)}
							>
								{category}
							</button>
						{/each}
					</div>
				{/if}

				{#if searchTerm || selectedCategory}
					<button class="btn btn-secondary" on:click={clearFilters}>
						Clear All Filters
					</button>
				{/if}
			</div>
		</div>

		<!-- Loading State -->
		{#if loading}
			<div class="loading-section">
				<LoadingSpinner size="large" />
				<p>Loading videos...</p>
			</div>
		{/if}

		<!-- Error State -->
		{#if error}
			<div class="error-section">
				<div class="error-card">
					<h3>Unable to load videos</h3>
					<p>{error}</p>
					<button 
						class="btn btn-primary"
						on:click={() => youtubeActions.getLatestVideos(20)}
					>
						Try Again
					</button>
				</div>
			</div>
		{/if}

		<!-- Videos Grid -->
		{#if !loading && !error && filteredVideos.length > 0}
			<div class="videos-section">
				<div class="section-header">
					<h2>
						{#if searchTerm}
							Search Results for "{searchTerm}" ({filteredVideos.length})
						{:else if selectedCategory}
							{selectedCategory} Videos ({filteredVideos.length})
						{:else}
							Latest Videos ({filteredVideos.length})
						{/if}
					</h2>
				</div>

				<div class="videos-grid">
					{#each filteredVideos as video (video.id)}
						<div class="video-card" on:click={() => openVideo(video)} on:keydown role="button" tabindex="0">
							<div class="video-thumbnail">
								<img
									src={getThumbnail(video)}
									alt={video.title}
									class="thumbnail-image"
									loading="lazy"
									on:error={(e) => { 
										const target = e.target as HTMLImageElement;
										if (!target) return;

										// Get current quality from URL
										const currentUrl = target.src;
										const currentQuality = currentUrl.split('/').pop()?.split('.')[0];

										// Quality fallback chain
										const fallbackMap = {
											maxresdefault: 'sddefault',
											sddefault: 'hqdefault',
											hqdefault: 'mqdefault',
											mqdefault: 'default'
										};

										// Try next quality level or use placeholder
										const nextQuality = fallbackMap[currentQuality as keyof typeof fallbackMap];
										if (nextQuality) {
											target.src = currentUrl.replace(`${currentQuality}.jpg`, `${nextQuality}.jpg`);
										} else {
											target.src = '/16X10_Placeholder_IMG.png';
										}
									}}
								/>
								<div class="play-overlay">
									<svg width="48" height="48" viewBox="0 0 24 24" fill="white">
										<path d="M8 5v14l11-7z"/>
									</svg>
								</div>
								<div class="duration-badge">
									{formatDuration(video.duration)}
								</div>
							</div>
							
							<div class="video-info">
								<h3 class="video-title">{video.title}</h3>
								<div class="video-meta">
									<span class="publish-date">{formatDate(video.published_at)}</span>
									<span class="view-count">{formatViewCount(video.view_count)}</span>
								</div>
								<p class="video-description">
									{video.description.length > 120 
										? video.description.substring(0, 120) + '...' 
										: video.description}
								</p>
							</div>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Empty State -->
		{#if !loading && !error && filteredVideos.length === 0}
			<div class="empty-section">
				<div class="empty-card">
					{#if searchTerm}
						<h3>No videos found</h3>
						<p>No videos match your search term "{searchTerm}"</p>
						<button 
							class="btn btn-primary"
							on:click={() => { searchTerm = ''; youtubeActions.clearSearch(); }}
						>
							Clear Search
						</button>
					{:else if selectedCategory}
						<h3>No videos in this category</h3>
						<p>No videos found for category "{selectedCategory}"</p>
						<button 
							class="btn btn-primary"
							on:click={() => handleCategoryFilter('')}
						>
							Show All Videos
						</button>
					{:else}
						<h3>No videos available</h3>
						<p>We're working on getting the latest videos from our YouTube channel.</p>
						<button 
							class="btn btn-primary"
							on:click={() => youtubeActions.getLatestVideos(20)}
						>
							Refresh
						</button>
					{/if}
				</div>
			</div>
		{/if}
	</div>
</div>

<Footer />

<style>
	.youtube-page {
		min-height: 100vh;
		background: var(--bg-primary);
		padding-top: 80px;
	}

	/* Hero Section - Matching site's design system */
	.hero-section {
		background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
		padding: var(--space-4xl) var(--space-xl);
		text-align: center;
		position: relative;
		overflow: hidden;
	}

	.hero-section::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="20" cy="20" r="1" fill="white" opacity="0.1"/><circle cx="80" cy="30" r="0.5" fill="white" opacity="0.1"/><circle cx="40" cy="60" r="1.5" fill="white" opacity="0.1"/></svg>');
		opacity: 0.3;
	}

	.hero-content {
		position: relative;
		z-index: 2;
		max-width: 800px;
		margin: 0 auto;
	}

	.hero-content h1 {
		font-size: var(--text-5xl);
		font-weight: 900;
		color: #ffd700;
		text-shadow: 4px 4px 8px rgba(0, 0, 0, 0.9);
		filter: drop-shadow(0 0 15px rgba(255, 215, 0, 0.4));
		margin-bottom: var(--space-lg);
		line-height: var(--leading-tight);
	}

	.hero-description {
		font-size: var(--text-xl);
		color: rgba(255, 255, 255, 0.95);
		text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
		line-height: var(--leading-relaxed);
		margin-bottom: var(--space-2xl);
	}

	.channel-stats {
		display: flex;
		justify-content: center;
		gap: var(--space-2xl);
		margin: var(--space-2xl) 0;
		flex-wrap: wrap;
	}

	.stat {
		text-align: center;
	}

	.stat-number {
		display: block;
		font-size: var(--text-2xl);
		font-weight: 700;
		color: #ffd700;
		text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
	}

	.stat-label {
		display: block;
		font-size: var(--text-sm);
		color: rgba(255, 255, 255, 0.8);
		text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.8);
		margin-top: var(--space-xs);
	}

	.channel-link {
		margin-top: var(--space-xl);
	}

	.btn-youtube {
		background: #ff0000;
		color: var(--white);
		border: none;
		padding: var(--space-md) var(--space-xl);
		border-radius: var(--radius-lg);
		font-weight: 600;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		gap: var(--space-sm);
		transition: all var(--transition-normal);
		font-size: var(--text-lg);
		position: relative;
		overflow: hidden;
	}

	.btn-youtube::before {
		content: '';
		position: absolute;
		top: 0;
		left: -100%;
		width: 100%;
		height: 100%;
		background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
		transition: left var(--transition-slow);
	}

	.btn-youtube:hover {
		background: #cc0000;
		transform: translateY(-2px);
		box-shadow: var(--shadow-glow);
	}

	.btn-youtube:hover::before {
		left: 100%;
	}

	.container {
		max-width: 1200px;
		margin: 0 auto;
		padding: 0 var(--space-xl);
	}

	/* Search and Filter Section */
	.search-section {
		padding: var(--space-2xl) 0;
		background: var(--bg-secondary);
	}

	.search-controls {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
		align-items: center;
	}

	.search-box {
		position: relative;
		max-width: 500px;
		width: 100%;
	}

	.search-box svg {
		position: absolute;
		left: var(--space-md);
		top: 50%;
		transform: translateY(-50%);
		color: var(--text-muted);
		z-index: 2;
	}

	.search-input {
		width: 100%;
		padding: var(--space-md) var(--space-md) var(--space-md) 44px;
		border: 2px solid transparent;
		border-radius: var(--radius-xl);
		font-size: var(--text-base);
		background: var(--bg-glass);
		backdrop-filter: blur(10px);
		-webkit-backdrop-filter: blur(10px);
		color: var(--text-primary);
		transition: all var(--transition-normal);
		box-shadow: var(--shadow-md);
	}

	.search-input:focus {
		outline: none;
		border-color: var(--primary);
		box-shadow: var(--shadow-glow);
		background: var(--bg-primary);
	}

	.search-input::placeholder {
		color: var(--text-muted);
	}

	.clear-search {
		position: absolute;
		right: var(--space-md);
		top: 50%;
		transform: translateY(-50%);
		background: none;
		border: none;
		color: var(--text-muted);
		cursor: pointer;
		padding: var(--space-xs);
		border-radius: var(--radius-md);
		transition: all var(--transition-fast);
	}

	.clear-search:hover {
		color: var(--text-primary);
		background: var(--bg-glass);
	}

	.category-filters {
		display: flex;
		gap: var(--space-sm);
		flex-wrap: wrap;
		justify-content: center;
	}

	.category-btn {
		padding: var(--space-sm) var(--space-lg);
		border: 2px solid transparent;
		border-radius: var(--radius-full);
		background: var(--bg-glass);
		backdrop-filter: blur(10px);
		-webkit-backdrop-filter: blur(10px);
		color: var(--text-secondary);
		font-size: var(--text-sm);
		font-weight: 500;
		cursor: pointer;
		transition: all var(--transition-normal);
		white-space: nowrap;
	}

	.category-btn:hover {
		color: var(--text-primary);
		background: var(--bg-glass-dark);
		transform: translateY(-1px);
	}

	.category-btn.active {
		background: var(--primary-gradient);
		color: var(--white);
		border-color: var(--primary);
		box-shadow: var(--shadow-glow);
	}

	/* Loading, Error, and Empty States */
	.loading-section, .error-section, .empty-section {
		text-align: center;
		padding: var(--space-4xl) var(--space-xl);
	}

	.error-card, .empty-card {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		padding: var(--space-2xl);
		border-radius: var(--radius-2xl);
		box-shadow: var(--shadow-lg);
		max-width: 400px;
		margin: 0 auto;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.error-card h3, .empty-card h3 {
		color: var(--text-primary);
		margin-bottom: var(--space-md);
	}

	.error-card p, .empty-card p {
		color: var(--text-secondary);
		margin-bottom: var(--space-lg);
	}

	/* Section Header */
	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-2xl);
		padding: 0 var(--space-xl);
	}

	.section-header h2 {
		font-size: var(--text-4xl);
		font-weight: 700;
		color: var(--text-primary);
		margin: 0;
	}

	/* Videos Section */
	.videos-section {
		padding: var(--space-2xl) 0 var(--space-4xl);
		background: var(--bg-primary);
	}

	.videos-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
		gap: var(--space-2xl);
		padding: 0 var(--space-xl);
	}

	/* Video Cards - Modern glass morphism design */
	.video-card {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		border-radius: var(--radius-2xl);
		overflow: hidden;
		box-shadow: var(--shadow-lg);
		transition: all var(--transition-normal);
		cursor: pointer;
		border: 1px solid rgba(255, 255, 255, 0.1);
		position: relative;
	}

	.video-card::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: var(--glass-gradient);
		opacity: 0;
		transition: opacity var(--transition-normal);
		pointer-events: none;
	}

	.video-card:hover {
		transform: translateY(-8px);
		box-shadow: var(--shadow-2xl);
	}

	.video-card:hover::before {
		opacity: 1;
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
		transition: transform var(--transition-slow);
	}

	.video-card:hover :global(.thumbnail-image) {
		transform: scale(1.05);
	}

	.play-overlay {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		background: rgba(0, 0, 0, 0.8);
		backdrop-filter: blur(10px);
		-webkit-backdrop-filter: blur(10px);
		border-radius: 50%;
		width: 80px;
		height: 80px;
		display: flex;
		align-items: center;
		justify-content: center;
		opacity: 0;
		transition: all var(--transition-normal);
		border: 2px solid rgba(255, 255, 255, 0.3);
	}

	.video-card:hover .play-overlay {
		opacity: 1;
		transform: translate(-50%, -50%) scale(1.1);
	}

	.duration-badge {
		position: absolute;
		bottom: var(--space-sm);
		right: var(--space-sm);
		background: rgba(0, 0, 0, 0.9);
		backdrop-filter: blur(10px);
		-webkit-backdrop-filter: blur(10px);
		color: var(--white);
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-md);
		font-size: var(--text-sm);
		font-weight: 600;
		border: 1px solid rgba(255, 255, 255, 0.2);
	}

	.video-info {
		padding: var(--space-lg);
		position: relative;
		z-index: 2;
	}

	.video-title {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-sm);
		line-height: var(--leading-snug);
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	.video-meta {
		display: flex;
		gap: var(--space-md);
		margin-bottom: var(--space-sm);
		font-size: var(--text-sm);
		color: var(--text-secondary);
		flex-wrap: wrap;
	}

	.video-description {
		color: var(--text-secondary);
		font-size: var(--text-sm);
		line-height: var(--leading-relaxed);
		margin: 0;
	}

	/* Button Styles - Using site's design system */
	.btn {
		padding: var(--space-sm) var(--space-lg);
		border-radius: var(--radius-lg);
		font-weight: 500;
		text-decoration: none;
		border: none;
		cursor: pointer;
		transition: all var(--transition-normal);
		display: inline-flex;
		align-items: center;
		gap: var(--space-sm);
		font-size: var(--text-base);
		position: relative;
		overflow: hidden;
	}

	.btn::before {
		content: '';
		position: absolute;
		top: 0;
		left: -100%;
		width: 100%;
		height: 100%;
		background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
		transition: left var(--transition-slow);
	}

	.btn:hover::before {
		left: 100%;
	}

	.btn-primary {
		background: var(--primary-gradient);
		color: var(--white);
		box-shadow: var(--shadow-md);
	}

	.btn-primary:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-glow);
	}

	.btn-secondary {
		background: var(--bg-glass);
		backdrop-filter: blur(10px);
		-webkit-backdrop-filter: blur(10px);
		color: var(--text-primary);
		border: 1px solid rgba(255, 255, 255, 0.2);
	}

	.btn-secondary:hover {
		background: var(--bg-glass-dark);
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}

	/* Responsive Design */
	@media (max-width: 1024px) {
		.videos-grid {
			grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
			gap: var(--space-lg);
		}
	}

	@media (max-width: 768px) {
		.hero-content h1 {
			font-size: var(--text-4xl);
		}

		.hero-description {
			font-size: var(--text-lg);
		}

		.channel-stats {
			gap: var(--space-lg);
		}

		.videos-grid {
			grid-template-columns: 1fr;
			gap: var(--space-lg);
			padding: 0 var(--space-md);
		}

		.section-header {
			flex-direction: column;
			gap: var(--space-md);
			text-align: center;
			padding: 0 var(--space-md);
		}

		.section-header h2 {
			font-size: var(--text-3xl);
		}

		.container {
			padding: 0 var(--space-md);
		}

		.search-section {
			padding: var(--space-xl) 0;
		}

		.search-controls {
			padding: 0 var(--space-md);
		}

		.category-filters {
			justify-content: flex-start;
			overflow-x: auto;
			padding-bottom: var(--space-sm);
		}
	}

	@media (max-width: 640px) {
		.hero-section {
			padding: var(--space-3xl) var(--space-md);
		}

		.hero-content h1 {
			font-size: var(--text-3xl);
		}

		.btn-youtube {
			padding: var(--space-sm) var(--space-lg);
			font-size: var(--text-base);
		}

		.video-info {
			padding: var(--space-md);
		}

		.video-title {
			font-size: var(--text-lg);
		}

		.stat-number {
			font-size: var(--text-xl);
		}
	}
</style> 