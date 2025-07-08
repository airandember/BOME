<script lang="ts">
	import { onMount } from 'svelte';
	import { videoService, type Video, type VideoCategory, type VideosResponse, type BunnyCollection } from '$lib/video';
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
	let latestVideos: Video[] = [];
	let collections: BunnyCollection[] = [];
	let categories: VideoCategory[] = [];
	let loading = false;
	let error = '';
	let searchQuery = '';
	let selectedCategory = '';
	let currentPage = 1;
	let hasMore = true;
	let loadingMore = false;
	let authChecking = true;
	let initialDataLoaded = false;
	let activeTab: 'latest' | 'collections' | 'topics' = 'latest';

	onMount(async () => {
		await loadInitialData();
	});

	async function loadInitialData() {
		if (initialDataLoaded || loading) return;

		try {
			loading = true;
			error = '';

			// Load collections
			try {
				const collectionsResponse = await videoService.getCollections();
				collections = collectionsResponse.items || [];
			} catch (err) {
				console.warn('Failed to load collections:', err);
				collections = [];
			}

			// Load categories
			try {
				const categoriesResponse = await videoService.getCategories();
				categories = categoriesResponse.categories || [];
			} catch (err) {
				console.warn('Failed to load categories:', err);
				categories = [];
			}

			// Load latest videos
			try {
				const response = await videoService.getVideos(1, 6); // Get latest 6 videos
				latestVideos = response.videos || [];
			} catch (err) {
				console.warn('Failed to load latest videos:', err);
				latestVideos = [];
			}

			// Load regular videos
			await loadVideos();
			initialDataLoaded = true;
		} catch (err: any) {
			handleError(err);
		} finally {
			loading = false;
		}
	}

	function handleError(err: any) {
		console.error('Error:', err);
		if (err.error_type === 'authentication_error') {
			error = 'Authentication required. Please log in.';
			toastStore.error('Please log in to view videos.');
		} else if (err.error_type === 'network_error') {
			error = 'Network error. Please check your connection.';
			toastStore.error('Network error. Please check your connection and try again.');
		} else {
			error = err.message || 'An error occurred';
			toastStore.error(error);
		}
	}

	async function loadVideos(reset = false) {
		try {
			if (reset) {
				currentPage = 1;
				videos = [];
			}

			console.log('Loading videos - page:', currentPage, 'category:', selectedCategory, 'search:', searchQuery);

			const response: VideosResponse = await videoService.getVideos(
				currentPage,
				20,
				selectedCategory || undefined,
				searchQuery || undefined
			);

			console.log('Videos response:', response);

			const newVideos = response.videos || [];
			
			if (reset) {
				videos = newVideos;
			} else {
				// Prevent duplicate videos by checking IDs
				const existingIds = new Set(videos.map(v => v.id));
				const uniqueNewVideos = newVideos.filter(video => !existingIds.has(video.id));
				videos = [...videos, ...uniqueNewVideos];
			}

			// Update pagination info
			hasMore = response.pagination?.has_more || false;
			console.log('Videos loaded:', newVideos.length, 'hasMore:', hasMore);
			
			// Clear any previous errors on success
			error = '';
		} catch (err: any) {
			console.error('Error loading videos:', err);
			
			// Provide more specific error messages
			if (err.error_type === 'authentication_error') {
				error = 'Authentication required. Please log in.';
				toastStore.error('Please log in to view videos.');
			} else if (err.error_type === 'network_error') {
				error = 'Network error. Please check your connection.';
				toastStore.error('Network error. Please check your connection and try again.');
			} else if (err.status === 429) {
				error = 'Too many requests. Please wait a moment.';
				toastStore.error('Too many requests. Please wait a moment before trying again.');
			} else {
				error = err.message || 'Failed to load videos';
				toastStore.error('Failed to load videos. Please try again.');
			}
		}
	}

	async function handleSearch() {
		console.log('Searching videos with query:', searchQuery);
		// Reset the initial data flag when searching
		initialDataLoaded = false;
		await loadVideos(true);
	}

	async function handleCategoryChange() {
		console.log('Filtering by category:', selectedCategory);
		// Reset the initial data flag when changing categories
		initialDataLoaded = false;
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
		// Reset the initial data flag when clearing filters
		initialDataLoaded = false;
		loadVideos(true);
	}

	function handleAuthLoadingChange(event: CustomEvent<{loading: boolean}>) {
		authChecking = event.detail.loading;
	}

	function handleAccessGranted() {
		// Only reload if we haven't already loaded data
		if (!initialDataLoaded && !loading) {
			loadInitialData();
		}
	}

	function switchTab(tab: typeof activeTab) {
		activeTab = tab;
		if (tab === 'latest') {
			searchQuery = '';
			selectedCategory = '';
			loadVideos(true);
		}
	}
</script>

<svelte:head>
	<title>Video Hub - Book of Mormon Evidences</title>
	<meta name="description" content="Explore our comprehensive collection of Book of Mormon evidence videos, organized by collections, topics, and latest uploads." />
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
					<p>Loading video hub...</p>
				</div>
			{:else}
				<div class="video-hub">
					<div class="container">
						<header class="hub-header">
							<h1>Video Hub</h1>
							<p>Discover our extensive collection of Book of Mormon evidence videos</p>
						</header>

						<!-- Navigation Tabs -->
						<div class="hub-tabs">
							<button 
								class="tab-button {activeTab === 'latest' ? 'active' : ''}" 
								on:click={() => switchTab('latest')}
							>
								Latest Videos
							</button>
							<button 
								class="tab-button {activeTab === 'collections' ? 'active' : ''}" 
								on:click={() => switchTab('collections')}
							>
								Collections
							</button>
							<button 
								class="tab-button {activeTab === 'topics' ? 'active' : ''}" 
								on:click={() => switchTab('topics')}
							>
								Topics
							</button>
						</div>

						<!-- Latest Videos Section -->
						{#if activeTab === 'latest'}
							<section class="latest-videos">
								<h2>Latest Uploads</h2>
								<div class="video-grid">
									{#each latestVideos as video (video.id)}
										<VideoCard {video} />
									{/each}
								</div>

								<div class="filters-section">
									<div class="search-bar">
										<input
											type="text"
											placeholder="Search videos..."
											bind:value={searchQuery}
											on:keydown={(e) => e.key === 'Enter' && handleSearch()}
										/>
										<button class="btn-primary" on:click={handleSearch}>
											🔍 Search
										</button>
									</div>
								</div>

								<div class="video-grid">
									{#each videos as video (video.id)}
										<VideoCard {video} />
									{/each}
								</div>

								{#if hasMore}
									<div class="load-more">
										<button class="btn-secondary" on:click={loadMore} disabled={loadingMore}>
											{loadingMore ? 'Loading...' : 'Load More'}
										</button>
									</div>
								{/if}
							</section>
						{/if}

						<!-- Collections Section -->
						{#if activeTab === 'collections'}
							<section class="collections">
								<h2>Video Collections</h2>
								<div class="collections-grid">
									{#each collections as collection (collection.id)}
										<div class="collection-card">
											<h3>{collection.name}</h3>
											<p>{collection.videoCount} videos</p>
											<button class="btn-primary" on:click={() => goto(`/videos/collections/${collection.id}`)}>
												View Collection
											</button>
										</div>
									{/each}
								</div>
							</section>
						{/if}

						<!-- Topics Section -->
						{#if activeTab === 'topics'}
							<section class="topics">
								<h2>Browse by Topic</h2>
								<div class="topics-grid">
									{#each categories as category (category.id)}
										<div class="topic-card">
											<h3>{category.name}</h3>
											<p>{category.videoCount} videos</p>
											<p class="description">{category.description}</p>
											<button class="btn-primary" on:click={() => goto(`/videos/topics/${category.name}`)}>
												Explore Topic
											</button>
										</div>
									{/each}
								</div>
							</section>
						{/if}

						{#if error}
							<div class="error-message">
								<p>{error}</p>
							</div>
						{/if}

						<!-- Ad Placement -->
						<div class="ad-placement">
							<AdDisplay placement="videos-footer" />
						</div>
					</div>
				</div>
			{/if}
		</SubscriptionCheck>
	</main>
</div>

<Footer />

<style lang="postcss">
	.main-content-wrapper {
		margin-top: 50px;
	}

	.video-hub {
		padding: 2rem 0;
	}

	.hub-header {
		text-align: center;
		margin-bottom: 2rem;

		h1 {
			font-size: 2.5rem;
			color: var(--color-primary);
			margin-bottom: 0.5rem;
		}

		p {
			font-size: 1.2rem;
			color: var(--color-text-muted);
		}
	}

	.hub-tabs {
		display: flex;
		justify-content: center;
		gap: 1rem;
		margin-bottom: 2rem;
		border-bottom: 1px solid var(--color-border);
		padding-bottom: 1rem;
	}

	.tab-button {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 0.5rem;
		background: transparent;
		color: var(--color-text);
		font-size: 1.1rem;
		cursor: pointer;
		transition: all 0.2s ease;

		&:hover {
			background: var(--color-bg-hover);
		}

		&.active {
			background: var(--color-primary);
			color: white;
		}
	}

	.video-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: 2rem;
		margin-bottom: 2rem;
	}

	.collections-grid, .topics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
		gap: 2rem;
		margin-bottom: 2rem;
	}

	.collection-card, .topic-card {
		background: var(--color-bg-card);
		border-radius: 1rem;
		padding: 1.5rem;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
		transition: transform 0.2s ease;

		&:hover {
			transform: translateY(-4px);
		}

		h3 {
			font-size: 1.25rem;
			margin-bottom: 0.5rem;
			color: var(--color-primary);
		}

		p {
			color: var(--color-text-muted);
			margin-bottom: 1rem;

			&.description {
				font-size: 0.9rem;
				line-height: 1.4;
				margin-bottom: 1.5rem;
			}
		}
	}

	.filters-section {
		margin: 2rem 0;
	}

	.search-bar {
		display: flex;
		gap: 1rem;
		max-width: 600px;
		margin: 0 auto;

		input {
			flex: 1;
			padding: 0.75rem 1rem;
			border: 1px solid var(--color-border);
			border-radius: 0.5rem;
			font-size: 1rem;
		}
	}

	.load-more {
		text-align: center;
		margin-top: 2rem;
	}

	.error-message {
		text-align: center;
		color: var(--color-error);
		padding: 1rem;
		margin: 1rem 0;
		background: var(--color-error-bg);
		border-radius: 0.5rem;
	}

	.ad-placement {
		margin: 2rem 0;
	}

	@media (max-width: 768px) {
		.hub-tabs {
			flex-direction: column;
			gap: 0.5rem;
		}

		.video-grid, .collections-grid, .topics-grid {
			grid-template-columns: 1fr;
		}

		.search-bar {
			flex-direction: column;
		}
	}
</style> 