<script lang="ts">
	import { onMount } from 'svelte';
	import { videoService, type Video, type VideoCategory } from '$lib/video';
	import VideoCard from '$lib/components/VideoCard.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import { toasts } from '$lib/stores/toast';

	let videos: Video[] = [];
	let categories: VideoCategory[] = [];
	let loading = true;
	let error = '';
	let searchQuery = '';
	let selectedCategory = '';
	let currentPage = 1;
	let hasMore = true;
	let loadingMore = false;

	onMount(async () => {
		await loadInitialData();
	});

	async function loadInitialData() {
		try {
			loading = true;
			error = '';

			// Load categories
			const categoriesResponse = await videoService.getCategories();
			categories = categoriesResponse.categories || [];

			// Load videos
			await loadVideos();
		} catch (err) {
			error = 'Failed to load videos';
			toasts.error('Failed to load videos. Please try again.');
			console.error('Error loading videos:', err);
		} finally {
			loading = false;
		}
	}

	async function loadVideos(reset = false) {
		try {
			if (reset) {
				currentPage = 1;
				videos = [];
			}

			const response = await videoService.getVideos(
				currentPage,
				20,
				selectedCategory || undefined,
				searchQuery || undefined
			);

			const newVideos = response.videos || [];
			
			if (reset) {
				videos = newVideos;
			} else {
				videos = [...videos, ...newVideos];
			}

			hasMore = newVideos.length === 20;
		} catch (err) {
			error = 'Failed to load videos';
			toasts.error('Failed to load videos. Please try again.');
			console.error('Error loading videos:', err);
		}
	}

	async function handleSearch() {
		await loadVideos(true);
	}

	async function handleCategoryChange() {
		await loadVideos(true);
	}

	async function loadMore() {
		if (loadingMore || !hasMore) return;

		try {
			loadingMore = true;
			currentPage++;
			await loadVideos();
		} finally {
			loadingMore = false;
		}
	}

	function clearFilters() {
		searchQuery = '';
		selectedCategory = '';
		loadVideos(true);
	}
</script>

<svelte:head>
	<title>Videos - Book of Mormon Evidences</title>
	<meta name="description" content="Browse our collection of Book of Mormon evidence videos." />
</svelte:head>

<div class="videos-page">
	<div class="container">
		<header class="page-header">
			<h1>Videos</h1>
			<p>Explore our collection of Book of Mormon evidence videos</p>
		</header>

		<div class="filters-section">
			<div class="search-bar">
				<input
					type="text"
					placeholder="Search videos..."
					bind:value={searchQuery}
					on:keydown={(e) => e.key === 'Enter' && handleSearch()}
					aria-label="Search videos"
				/>
				<button class="btn-primary" on:click={handleSearch}>
					🔍 Search
				</button>
			</div>

			<div class="category-filters">
				<select 
					bind:value={selectedCategory} 
					on:change={handleCategoryChange}
					aria-label="Filter by category"
				>
					<option value="">All Categories</option>
					{#each categories as category}
						<option value={category.name}>{category.name} ({category.videoCount})</option>
					{/each}
				</select>

				<button class="btn-secondary" on:click={clearFilters}>
					Clear Filters
				</button>
			</div>
		</div>

		{#if error}
			<div class="error-message">
				{error}
			</div>
		{/if}

		{#if loading}
			<LoadingSpinner size="large" text="Loading videos..." />
		{:else if videos.length === 0}
			<div class="empty-state">
				<div class="empty-icon">📹</div>
				<h3>No videos found</h3>
				<p>Try adjusting your search or filters</p>
				<button class="btn-primary" on:click={clearFilters}>
					Clear All Filters
				</button>
			</div>
		{:else}
			<div class="videos-grid">
				{#each videos as video (video.id)}
					<VideoCard {video} />
				{/each}
			</div>

			{#if hasMore}
				<div class="load-more">
					<button 
						class="btn-secondary" 
						on:click={loadMore}
						disabled={loadingMore}
					>
						{loadingMore ? 'Loading...' : 'Load More'}
					</button>
				</div>
			{/if}
		{/if}
	</div>
</div>

<style>
	.videos-page {
		min-height: 100vh;
		background: var(--bg-color);
		padding: 2rem 0;
	}

	.container {
		max-width: 1200px;
		margin: 0 auto;
		padding: 0 2rem;
	}

	.page-header {
		text-align: center;
		margin-bottom: 3rem;
	}

	.page-header h1 {
		font-size: 2.5rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.page-header p {
		font-size: 1.1rem;
		color: var(--text-secondary);
	}

	.filters-section {
		background: var(--card-bg);
		padding: 2rem;
		border-radius: 20px;
		margin-bottom: 2rem;
		box-shadow: 
			8px 8px 16px var(--shadow-dark),
			-8px -8px 16px var(--shadow-light);
	}

	.search-bar {
		display: flex;
		gap: 1rem;
		margin-bottom: 1.5rem;
	}

	.search-bar input {
		flex: 1;
		padding: 0.75rem 1rem;
		border: none;
		border-radius: 12px;
		background: var(--input-bg);
		color: var(--text-primary);
		font-size: 1rem;
		box-shadow: 
			inset 2px 2px 4px var(--shadow-dark),
			inset -2px -2px 4px var(--shadow-light);
	}

	.search-bar input:focus {
		outline: none;
		box-shadow: 
			inset 2px 2px 4px var(--shadow-dark),
			inset -2px -2px 4px var(--shadow-light),
			0 0 0 2px var(--accent-color);
	}

	.category-filters {
		display: flex;
		gap: 1rem;
		align-items: center;
	}

	.category-filters select {
		padding: 0.75rem 1rem;
		border: none;
		border-radius: 12px;
		background: var(--input-bg);
		color: var(--text-primary);
		font-size: 1rem;
		box-shadow: 
			inset 2px 2px 4px var(--shadow-dark),
			inset -2px -2px 4px var(--shadow-light);
		cursor: pointer;
	}

	.category-filters select:focus {
		outline: none;
		box-shadow: 
			inset 2px 2px 4px var(--shadow-dark),
			inset -2px -2px 4px var(--shadow-light),
			0 0 0 2px var(--accent-color);
	}

	.videos-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: 2rem;
		margin-bottom: 2rem;
	}

	.load-more {
		text-align: center;
		margin-top: 2rem;
	}

	.empty-state {
		text-align: center;
		padding: 4rem 0;
	}

	.empty-icon {
		font-size: 4rem;
		margin-bottom: 1rem;
	}

	.empty-state h3 {
		font-size: 1.5rem;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.empty-state p {
		color: var(--text-secondary);
		font-size: 1.1rem;
		margin-bottom: 2rem;
	}

	.error-message {
		background: var(--error-bg);
		color: var(--error-text);
		padding: 1rem;
		border-radius: 12px;
		text-align: center;
		margin-bottom: 2rem;
	}

	.btn-primary,
	.btn-secondary {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 12px;
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.btn-primary {
		background: var(--accent-color);
		color: white;
		box-shadow: 
			4px 4px 8px var(--shadow-dark),
			-2px -2px 4px var(--shadow-light);
	}

	.btn-primary:hover {
		background: var(--accent-hover);
		transform: translateY(-2px);
		box-shadow: 
			6px 6px 12px var(--shadow-dark),
			-3px -3px 6px var(--shadow-light);
	}

	.btn-secondary {
		background: var(--card-bg);
		color: var(--text-primary);
		box-shadow: 
			4px 4px 8px var(--shadow-dark),
			-2px -2px 4px var(--shadow-light);
	}

	.btn-secondary:hover:not(:disabled) {
		transform: translateY(-2px);
		box-shadow: 
			6px 6px 12px var(--shadow-dark),
			-3px -3px 6px var(--shadow-light);
	}

	.btn-secondary:disabled {
		opacity: 0.6;
		cursor: not-allowed;
		transform: none;
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.container {
			padding: 0 1rem;
		}

		.page-header h1 {
			font-size: 2rem;
		}

		.filters-section {
			padding: 1.5rem;
		}

		.search-bar {
			flex-direction: column;
		}

		.category-filters {
			flex-direction: column;
			align-items: stretch;
		}

		.videos-grid {
			grid-template-columns: 1fr;
			gap: 1.5rem;
		}
	}

	@media (min-width: 769px) and (max-width: 1024px) {
		.videos-grid {
			grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		}
	}

	/* Touch-friendly improvements */
	@media (hover: none) and (pointer: coarse) {
		.search-bar input,
		.category-filters select,
		.btn-primary,
		.btn-secondary {
			min-height: 44px;
		}
	}
</style> 