<script lang="ts">
	import { onMount } from 'svelte';
	import { videoService, type Video, type VideoCategory } from '$lib/video';
	import VideoCard from '$lib/components/VideoCard.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import Navigation from '$lib/components/Navigation.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import { toastStore } from '$lib/stores/toast';
	import AdDisplay from '$lib/components/AdDisplay.svelte';
	import SubscriptionCheck from '$lib/components/SubscriptionCheck.svelte';
	import { isAdmin } from '$lib/auth';
	import { goto } from '$app/navigation';

	let videos: Video[] = [];
	let categories: VideoCategory[] = [];
	let loading = false;
	let error = '';
	let searchQuery = '';
	let selectedCategory = '';
	let currentPage = 1;
	let hasMore = true;
	let loadingMore = false;
	let authChecking = true;

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
			toastStore.error('Failed to load videos. Please try again.');
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
			toastStore.error('Failed to load videos. Please try again.');
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

	function handleAuthLoadingChange(event: CustomEvent<{loading: boolean}>) {
		authChecking = event.detail.loading;
	}

	function handleAccessGranted() {
		loadInitialData();
	}
</script>

<svelte:head>
	<title>Videos - Book of Mormon Evidences</title>
	<meta name="description" content="Browse our collection of Book of Mormon evidence videos." />
</svelte:head>

<Navigation />

<div class="page-wrapper">
	<main class="main-content-wrapper">
		<SubscriptionCheck 
			redirectTo="/login" 
			requireSubscription={true}
			on:loadingChange={handleAuthLoadingChange}
			on:accessGranted={handleAccessGranted}
		>
			{#if loading}
				<div class="loading-container">
					<LoadingSpinner size="large" />
					<p>Loading videos...</p>
				</div>
			{:else}
				<div class="videos-page">
					<div class="container">
						<!-- Ad Placement: Header Banner -->
						<div class="ad-placement">
							<AdDisplay placement="videos-header" />
						</div>

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

						<!-- Ad Placement: Mid-page -->
						<div class="ad-placement">
							<AdDisplay placement="videos-mid" />
						</div>

						{#if error}
							<div class="error-message">
								{error}
							</div>
						{/if}

						{#if videos.length === 0}
							<div class="empty-state">
								<div class="empty-icon">📹</div>
								<h3>No videos found</h3>
								<p>Try adjusting your search or filters</p>
								<button class="btn-primary" on:click={clearFilters}>
									Clear All Filters
								</button>
							</div>
						{:else}
							<div class="content-with-sidebar">
								<div class="main-content">
									<div class="videos-grid">
										{#each videos as video, index (video.id)}
											<VideoCard {video} />
											<!-- Ad between videos every 6 videos -->
											{#if (index + 1) % 6 === 0 && index < videos.length - 1}
												<div class="ad-placement between-videos">
													<AdDisplay placement="videos-between" />
												</div>
											{/if}
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
								</div>

								<!-- Sidebar with ads -->
								<aside class="sidebar">
									<div class="sidebar-content">
										<h3>Sponsored</h3>
										<div class="ad-placement">
											<AdDisplay placement="videos-sidebar" />
										</div>
										
										<!-- Additional sidebar content can go here -->
										<div class="sidebar-section">
											<h4>Popular Categories</h4>
											<div class="category-list">
												{#each categories.slice(0, 5) as category}
													<button 
														class="category-item"
														on:click={() => {
															selectedCategory = category.name;
															handleCategoryChange();
														}}
													>
														{category.name} ({category.videoCount})
													</button>
												{/each}
											</div>
										</div>
									</div>
								</aside>
							</div>
						{/if}

						<!-- Ad Placement: Footer -->
						<div class="ad-placement">
							<AdDisplay placement="videos-footer" />
						</div>
					</div>
				</div>
			{/if}
		</SubscriptionCheck>
	</main>
	<Footer />
</div>

<style>
	.page-wrapper {
		display: flex;
		flex-direction: column;
		min-height: 100vh;
	}

	.main-content-wrapper {
		flex: 1 0 auto;
		width: 100%;
	}

	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 50vh;
		gap: 1rem;
	}

	.loading-container p {
		color: var(--text-color);
		font-size: 1.1rem;
	}

	.videos-page {
		background: var(--bg-color);
		padding: 2rem 0;
		width: 100%;
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

	.content-with-sidebar {
		display: grid;
		grid-template-columns: 1fr 300px;
		gap: 2rem;
		margin-bottom: 2rem;
	}

	.main-content {
		min-width: 0; /* Prevent overflow */
	}

	.videos-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: 2rem;
		margin-bottom: 2rem;
	}

	.sidebar {
		position: sticky;
		top: 2rem;
		height: fit-content;
	}

	.sidebar-content {
		background: var(--card-bg);
		border-radius: 20px;
		padding: 1.5rem;
		box-shadow: 
			8px 8px 16px var(--shadow-dark),
			-8px -8px 16px var(--shadow-light);
	}

	.sidebar h3 {
		font-size: 1.2rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.sidebar-section {
		margin-top: 2rem;
		padding-top: 2rem;
		border-top: 1px solid var(--border-color);
	}

	.sidebar-section h4 {
		font-size: 1rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.category-list {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.category-item {
		background: var(--bg-secondary);
		border: none;
		padding: 0.75rem 1rem;
		border-radius: 10px;
		color: var(--text-primary);
		cursor: pointer;
		transition: all 0.3s ease;
		text-align: left;
		font-size: 0.9rem;
		box-shadow: 
			2px 2px 4px var(--shadow-dark),
			-2px -2px 4px var(--shadow-light);
	}

	.category-item:hover {
		background: var(--primary-color);
		color: white;
		transform: translateY(-1px);
	}

	/* Ad Placements */
	.ad-placement {
		margin: 2rem 0;
		display: flex;
		justify-content: center;
	}

	.ad-placement.between-videos {
		grid-column: 1 / -1; /* Span full width of grid */
		margin: 1rem 0;
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
	@media (max-width: 1024px) {
		.content-with-sidebar {
			grid-template-columns: 1fr 250px;
		}
	}

	@media (max-width: 768px) {
		.content-with-sidebar {
			grid-template-columns: 1fr;
			gap: 1rem;
		}

		.sidebar {
			position: static;
			order: -1; /* Show sidebar above content on mobile */
		}

		.sidebar-content {
			padding: 1rem;
		}

		.videos-grid {
			grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
			gap: 1rem;
		}

		.filters-section {
			padding: 1rem;
		}

		.search-bar {
			flex-direction: column;
			gap: 0.5rem;
		}

		.category-filters {
			flex-direction: column;
			gap: 0.5rem;
		}
	}

	@media (max-width: 480px) {
		.videos-grid {
			grid-template-columns: 1fr;
		}

		.container {
			padding: 0 1rem;
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