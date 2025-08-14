<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { videoService, type Video, type VideosResponse, type BunnyCollection } from '$lib/video';
	import VideoCard from '$lib/components/VideoCard.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import Navigation from '$lib/components/Navigation.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import { toastStore } from '$lib/stores/toast';
	import SubscriptionCheck from '$lib/components/SubscriptionCheck.svelte';

	let videos: Video[] = [];
	let collection: BunnyCollection | null = null;
	let loading = true;
	let loadingMore = false;
	let error = '';
	let currentPage = 1;
	let itemsPerPage = 20;
	let hasMore = true;
	let authChecking = true;
	let debugInfo = '';
	let scrollThreshold = 800; // pixels from bottom to trigger auto-load (accounts for footer height)

	$: collectionId = $page.params.id;
	$: console.log('Collection ID:', collectionId);

	onMount(() => {
		if (collectionId) {
			console.log('Loading collection data for ID:', collectionId);
			loadCollectionData();
		} else {
			console.error('No collection ID provided');
			error = 'No collection ID provided';
		}

		// Add scroll listener for infinite scroll
		const handleScroll = () => {
			// Only auto-load if we have more content and not already loading
			if (hasMore && !loadingMore && !loading) {
				const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
				const windowHeight = window.innerHeight;
				const documentHeight = document.documentElement.scrollHeight;
				
				// Check if we're within the threshold of the bottom
				if (scrollTop + windowHeight >= documentHeight - scrollThreshold) {
					console.log('🚀 Collection infinite scroll triggered!', {
						scrollTop,
						windowHeight,
						documentHeight,
						threshold: scrollThreshold,
						hasMore,
						loadingMore,
						videosCount: videos.length
					});
					loadMore();
				}
			}
		};

		// Add throttled scroll listener
		let scrollTimer: NodeJS.Timeout | null = null;
		const throttledScroll = () => {
			if (scrollTimer) return;
			scrollTimer = setTimeout(() => {
				handleScroll();
				scrollTimer = null;
			}, 100);
		};

		window.addEventListener('scroll', throttledScroll);
		
		// Cleanup function
		return () => {
			window.removeEventListener('scroll', throttledScroll);
			if (scrollTimer) clearTimeout(scrollTimer);
		};
	});

	async function loadCollectionData() {
		try {
			loading = true;
			error = '';
			debugInfo = '';

			console.log('Fetching collection data...');
			
			// Load collection details and its videos in parallel
			const [collectionResponse, videosResponse] = await Promise.all([
				//@ts-ignore
				videoService.getCollection(collectionId).catch(err => {
					console.error('Collection fetch error:', err);
					throw new Error(`Failed to fetch collection: ${err.message}`);
				}),
				//@ts-ignore
				videoService.getVideosByCollection(collectionId, currentPage, itemsPerPage).catch(err => {
					console.error('Videos fetch error:', err);
					throw new Error(`Failed to fetch videos: ${err.message}`);
				})
			]);

			console.log('Collection response:', collectionResponse);
			console.log('Videos response:', videosResponse);

			collection = collectionResponse;
			videos = videosResponse.videos;
			hasMore = videosResponse.pagination.hasMore;

			// Debug information
			debugInfo = `Found ${videos.length} videos. Has more: ${hasMore}. Current page: ${currentPage}`;
			console.log(debugInfo);

			if (videos.length === 0) {
				console.log('No videos found in collection');
				toastStore.info('No videos found in this collection');
			}

			// Log video IDs for debugging
			console.log('Video IDs:', videos.map(v => ({ id: v.id, bunnyId: v.bunnyVideoId })));

		} catch (err: unknown) {
			console.error('Error in loadCollectionData:', err);
			error = err instanceof Error ? err.message : 'Failed to load collection';
			toastStore.error(error);
			debugInfo = `Error occurred: ${error}. Check console for details.`;
		} finally {
			loading = false;
		}
	}

	async function loadMore() {
		if (loadingMore || !hasMore) return;
		
		try {
			loadingMore = true;
			console.log(`Loading more videos. Page ${currentPage + 1}`);
			//@ts-ignore
			const response = await videoService.getVideosByCollection(collectionId, currentPage + 1, itemsPerPage);
			console.log('Load more response:', response);
			
			videos = [...videos, ...response.videos];
			hasMore = response.pagination.hasMore;
			currentPage++;
			
			console.log(`Loaded ${response.videos.length} more videos. Total: ${videos.length}`);
			
			if (response.videos.length === 0) {
				toastStore.info('No more videos to load');
			}
		} catch (err) {
			console.error('Error loading more videos:', err);
			toastStore.error('Failed to load more videos');
		} finally {
			loadingMore = false;
		}
	}

	function handleAuthLoadingChange(data: {loading: boolean}) {
		authChecking = data.loading;
	}

	function handleAccessGranted() {
		// This function is called when the user is granted access.
		// You can perform actions here, such as re-fetching data or updating UI.
		console.log('Access granted!');
		loadCollectionData(); // Re-fetch data after successful access
	}
</script>

<svelte:head>
	<title>{collection?.name || 'Collection'} - BOME Video Hub</title>
	<meta name="description" content="Watch videos from the {collection?.name || 'collection'} on BOME Video Hub" />
</svelte:head>

<div class="main-content-wrapper">
	<Navigation />
	
	<main class="collection-page">
		<SubscriptionCheck 
			redirectTo="/login"
			requireSubscription={true}
			requiredTier="premium"
			onLoadingChange={handleAuthLoadingChange}
			onAccessGranted={handleAccessGranted}
		>
			{#if loading}
				<div class="loading-container">
					<LoadingSpinner />
					<p>Loading collection...</p>
				</div>
			{:else if error}
				<div class="error-container">
					<h1>Error</h1>
					<p>{error}</p>
					{#if debugInfo}
						<p class="debug-info">{debugInfo}</p>
					{/if}
					<button class="btn-primary" on:click={() => goto('/videos?tab=collections')}>
						← Back to Collections
					</button>
				</div>
			{:else if collection}
				<div class="collection-header">
					<button class="back-button" on:click={() => goto('/videos?tab=collections')}>
						← Back to Collections
					</button>
					<h1>{collection.name}</h1>
					<div class="collection-info">
						<p>{collection.videoCount} videos in this collection</p>
						<p class="collection-date">Created: {new Date(collection.dateCreated).toLocaleDateString()}</p>
						{#if debugInfo}
							<p class="debug-info">{debugInfo}</p>
						{/if}
					</div>
				</div>

				<div class="videos-container">
					{#if videos.length === 0}
						<div class="no-videos">
							<p>No videos found in this collection.</p>
							<p class="debug-info">Collection ID: {collectionId}</p>
						</div>
					{:else}
						<div class="videos-grid">
							{#each videos as video (video.id || video.bunnyVideoId || crypto.randomUUID())}
								<VideoCard 
									{video} 
									on:click={() => goto(`/videos/${video.bunnyVideoId || video.id}`)}
								/>
							{/each}
						</div>

						{#if hasMore}
							<div class="load-more-container">
								<button 
									class="btn-secondary load-more-btn"
									on:click={loadMore}
									disabled={loadingMore}
								>
									{#if loadingMore}
										<LoadingSpinner size="small" />
										Loading...
									{:else}
										Load More Videos (or keep scrolling)
									{/if}
								</button>
							</div>
						{/if}
					{/if}
				</div>
			{:else}
				<div class="error-container">
					<h1>Collection Not Found</h1>
					<p>The requested collection could not be found.</p>
					{#if debugInfo}
						<p class="debug-info">{debugInfo}</p>
					{/if}
					<button class="btn-primary" on:click={() => goto('/videos?tab=collections')}>
						← Back to Collections
					</button>
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

	.collection-page {
		padding: 2rem 0;
		min-height: 60vh;
	}

	.loading-container {
		text-align: center;
		padding: 4rem 0;
		
		p {
			margin-top: 1rem;
			color: var(--color-text-secondary);
		}
	}

	.error-container {
		text-align: center;
		padding: 4rem 0;
		
		h1 {
			color: var(--color-error);
			margin-bottom: 1rem;
		}
		
		p {
			color: var(--color-text-secondary);
			margin-bottom: 2rem;
		}
	}

	.collection-header {
		text-align: center;
		margin-bottom: 3rem;
		
		.back-button {
			background: none;
			border: none;
			color: var(--color-primary);
			cursor: pointer;
			font-size: 1rem;
			margin-bottom: 1rem;
			padding: 0.5rem 1rem;
			border-radius: 4px;
			transition: background-color 0.2s;
			
			&:hover {
				background: var(--color-surface-hover);
			}
		}
		
		h1 {
			font-size: 2.5rem;
			color: var(--color-primary);
			margin-bottom: 1rem;
		}
		
		.collection-info {
			color: var(--color-text-secondary);
			
			p {
				margin: 0.5rem 0;
			}
			
			.collection-date {
				font-size: 0.9rem;
			}
		}
	}

	.videos-container {
		width: 95vw;
		max-width: 1200px;
		margin: 0 auto;
		padding: 0 1rem;
	}

	.videos-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: 1.5rem;
		margin-bottom: 2rem;
	}

	.no-videos {
		text-align: center;
		padding: 4rem 0;
		color: var(--color-text-secondary);
	}

	.load-more-container {
		text-align: center;
		margin-top: 2rem;
	}

	.load-more-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		margin: 0 auto;
		
		&:disabled {
			opacity: 0.7;
			cursor: not-allowed;
		}
	}

	.debug-info {
		font-family: monospace;
		font-size: 0.9rem;
		color: var(--color-text-secondary);
		opacity: 0.8;
		margin-top: 1rem;
		padding: 0.5rem;
		background: rgba(0, 0, 0, 0.05);
		border-radius: 4px;
	}

	@media (max-width: 768px) {
		.collection-header h1 {
			font-size: 2rem;
		}
		
		.videos-grid {
			grid-template-columns: 1fr;
			gap: 1rem;
		}
	}
</style>
