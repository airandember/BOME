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
	import { auth, isAdmin } from '$lib/auth';
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
	let activeTab: 'latest' | 'collections' | 'topics' | 'allVideos' = 'allVideos';

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
		if (tab === 'latest' || tab === 'allVideos') {
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
								class="tab-button {activeTab === 'allVideos' ? 'active' : ''}"
								on:click={() => switchTab('allVideos')}
							>
								All Videos
							</button>
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

						<!-- All Videos Section -->
						{#if activeTab === 'allVideos'}
						
							<section class="all-videos">
								<h2>Book of Mormon Evidence Videos</h2>
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


						<!-- Latest Videos Section -->
						{#if activeTab === 'latest'}
							<section class="latest-videos">
								<h2>Latest Uploads</h2>
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
									{#each latestVideos as video (video.id)}
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
			margin-bottom: 1rem;
		}

		p {
			font-size: 1.1rem;
			color: var(--color-text-secondary);
			margin-bottom: 1rem;
		}
	}

	.auth-notice {
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: 8px;
		padding: 1rem;
		margin-top: 1rem;
		max-width: 600px;
		margin-left: auto;
		margin-right: auto;

		p {
			margin: 0;
			font-size: 0.9rem;
			color: var(--color-text-secondary);
		}

		.link-button {
			background: none;
			border: none;
			color: var(--color-primary);
			cursor: pointer;
			text-decoration: underline;
			font-size: inherit;
			padding: 0;
			margin: 0;
			display: inline;
		}

		.link-button:hover {
			color: var(--color-primary-hover);
		}
	}

	.hub-tabs {
		display: flex;
		justify-content: center;
		gap: 1rem;
		margin-bottom: 2rem;
		flex-wrap: wrap;
	}

	.tab-button {
		padding: 0.75rem 1.5rem;
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		color: var(--color-text);
		border-radius: 8px;
		cursor: pointer;
		transition: all 0.2s;
		font-weight: 500;

		&:hover {
			background: var(--color-surface-hover);
			border-color: var(--color-primary);
		}

		&.active {
			background: var(--color-primary);
			color: white;
			border-color: var(--color-primary);
		}
	}

	.all-videos, .latest-videos, .collections, .topics {
		margin-bottom: 2rem;
	}

	.all-videos h2, .latest-videos h2, .collections h2, .topics h2 {
		font-size: 1.8rem;
		color: var(--color-text);
		margin-bottom: 1.5rem;
		text-align: center;
	}

	.filters-section {
		display: flex;
		justify-content: center;
		margin-bottom: 2rem;
	}

	.search-bar {
		display: flex;
		gap: 0.5rem;
		max-width: 400px;
		width: 100%;
	}

	.search-bar input {
		flex: 1;
		padding: 0.75rem;
		border: 1px solid var(--color-border);
		border-radius: 8px;
		font-size: 1rem;
		background: var(--color-surface);
		color: var(--color-text);
	}

	.search-bar input:focus {
		outline: none;
		border-color: var(--color-primary);
	}

	.video-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: 1.5rem;
		margin-bottom: 2rem;
	}

	.collections-grid, .topics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 1.5rem;
		margin-bottom: 2rem;
	}

	.collection-card, .topic-card {
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: 12px;
		padding: 1.5rem;
		text-align: center;
		transition: transform 0.2s, box-shadow 0.2s;
	}

	.collection-card:hover, .topic-card:hover {
		transform: translateY(-4px);
		box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
	}

	.collection-card h3, .topic-card h3 {
		font-size: 1.2rem;
		color: var(--color-text);
		margin-bottom: 0.5rem;
	}

	.collection-card p, .topic-card p {
		color: var(--color-text-secondary);
		margin-bottom: 1rem;
	}

	.topic-card .description {
		font-size: 0.9rem;
		line-height: 1.4;
	}

	.load-more {
		display: flex;
		justify-content: center;
		margin-top: 2rem;
	}

	.error-message {
		background: #fee;
		border: 1px solid #fcc;
		color: #c33;
		padding: 1rem;
		border-radius: 8px;
		margin: 1rem 0;
		text-align: center;
	}

	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 4rem 2rem;
		text-align: center;
	}

	.loading-container p {
		margin-top: 1rem;
		color: var(--color-text-secondary);
	}

	.ad-placement {
		margin-top: 3rem;
		text-align: center;
	}

	.container {
		max-width: 1200px;
		margin: 0 auto;
		padding: 0 1rem;
	}

	.btn-primary {
		background: var(--color-primary);
		color: white;
		border: none;
		padding: 0.75rem 1.5rem;
		border-radius: 8px;
		cursor: pointer;
		font-weight: 500;
		transition: background 0.2s;

		&:hover {
			background: var(--color-primary-hover);
		}

		&:disabled {
			opacity: 0.5;
			cursor: not-allowed;
		}
	}

	.btn-secondary {
		background: var(--color-surface);
		color: var(--color-text);
		border: 1px solid var(--color-border);
		padding: 0.75rem 1.5rem;
		border-radius: 8px;
		cursor: pointer;
		font-weight: 500;
		transition: all 0.2s;

		&:hover {
			background: var(--color-surface-hover);
			border-color: var(--color-primary);
		}

		&:disabled {
			opacity: 0.5;
			cursor: not-allowed;
		}
	}

	@media (max-width: 768px) {
		.hub-header h1 {
			font-size: 2rem;
		}

		.hub-tabs {
			gap: 0.5rem;
		}

		.tab-button {
			padding: 0.5rem 1rem;
			font-size: 0.9rem;
		}

		.video-grid {
			grid-template-columns: 1fr;
			gap: 1rem;
		}

		.collections-grid, .topics-grid {
			grid-template-columns: 1fr;
			gap: 1rem;
		}

		.search-bar {
			flex-direction: column;
			gap: 0.5rem;
		}

		.search-bar input {
			width: 100%;
		}
	}

	@media (max-width: 480px) {
		.hub-header h1 {
			font-size: 1.8rem;
		}

		.hub-tabs {
			flex-direction: column;
			align-items: center;
		}

		.tab-button {
			width: 100%;
			max-width: 200px;
		}
	}
</style> 