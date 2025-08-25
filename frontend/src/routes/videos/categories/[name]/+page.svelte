<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { videoService, type Video, type VideoCategory } from '$lib/video';
	import VideoCard from '$lib/components/VideoCard.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import Navigation from '$lib/components/Navigation.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import { toastStore } from '$lib/stores/toast';

	let category: VideoCategory | null = null;
	let videos: Video[] = [];
	let loading = true;
	let error = '';
	let currentPage = 1;
	let hasMore = true;
	let loadingMore = false;

	onMount(async () => {
		const categoryParam = $page.params.name;
		if (!categoryParam) return;

		try {
			// Check if the param is a number (category ID) or a name
			const isNumeric = /^\d+$/.test(categoryParam);
			
			if (isNumeric) {
				// It's a category ID, load directly
				const categoryId = parseInt(categoryParam);
				
				// Load category details from tag categories
				const categoriesResponse = await videoService.getTagCategories();
				category = categoriesResponse.categories.find(c => c.id === categoryId) || null;
			} else {
				// It's a category name, find by name (fallback for old URLs)
				const categoriesResponse = await videoService.getTagCategories();
				category = categoriesResponse.categories.find(c => c.name === categoryParam) || null;
			}
			
			if (!category) {
				error = 'Category not found';
				loading = false;
				return;
			}

			// Load videos for this category using tag-based filtering
			await loadVideos();
		} catch (err: any) {
			console.error('Error loading category:', err);
			error = err.message || 'Failed to load category';
			toastStore.error(error);
		} finally {
			loading = false;
		}
	});

	async function loadVideos(reset = false) {
		if (!category) return;
		
		if (reset) {
			currentPage = 1;
			videos = [];
		}

		try {
			loadingMore = true;
			const response = await videoService.getVideosByTagCategory(category.id, currentPage, 20);
			
			if (reset) {
				videos = response.videos;
			} else {
				videos = [...videos, ...response.videos];
			}
			
			hasMore = response.pagination.hasMore;
			currentPage++;
		} catch (err: any) {
			console.error('Error loading more videos:', err);
			toastStore.error('Failed to load more videos');
		} finally {
			loadingMore = false;
		}
	}

	function loadMore() {
		if (!loadingMore && hasMore) {
			loadVideos();
		}
	}
</script>

<svelte:head>
	<title>{category?.name || 'Category'} Videos - Book of Mormon Evidences</title>
	<meta name="description" content="Browse {category?.name} videos about Book of Mormon evidences" />
</svelte:head>

<Navigation />

<div class="page-wrapper">
	<main class="main-content-wrapper">
		{#if loading}
			<div class="loading-container">
				<LoadingSpinner size="large" />
				<p>Loading category videos...</p>
			</div>
		{:else if error}
			<div class="error-message">
				<p>{error}</p>
			</div>
		{:else}
			<div class="category-page">
				<div class="container">
					<header class="category-header">
						<div class="header-content">
							<a href="/videos" class="back-link" aria-label="Back to categories">
								<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
									<path d="M19 12H5"/>
									<path d="M12 19l-7-7 7-7"/>
								</svg>
							</a>
							<h1>{category?.name}</h1>
						</div>
						{#if category?.description}
							<p class="category-description">{category.description}</p>
						{/if}
						<div class="category-stats">
							<span>{category?.videoCount} videos</span>
						</div>
					</header>

					<div class="video-grid">
						{#each videos as video (video.id)}
							<VideoCard {video} />
						{/each}
					</div>

					{#if videos.length === 0}
						<div class="empty-state">
							<p>No videos found in this category</p>
						</div>
					{:else if hasMore}
						<div class="load-more">
							<button class="btn-secondary" on:click={loadMore} disabled={loadingMore}>
								{loadingMore ? 'Loading...' : 'Load More'}
							</button>
						</div>
					{/if}
				</div>
			</div>
		{/if}
	</main>
</div>

<Footer />

<style lang="postcss">
	.main-content-wrapper {
		margin-top: 50px;
	}

	.category-page {
		padding: 2rem 0;
	}

	.category-header {
		text-align: center;
		margin-bottom: 2rem;
	}

	.header-content {
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		margin-bottom: 0.5rem;
	}

	.back-link {
		position: absolute;
		left: 0;
		color: var(--color-text-muted);
		transition: color 0.2s;
		padding: 0.5rem;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;

		&:hover {
			color: var(--color-primary);
			background: var(--color-bg-hover);
		}
	}

	h1 {
		font-size: 2.5rem;
		color: var(--color-primary);
		margin: 0;
	}

	.category-description {
		max-width: 800px;
		margin: 0 auto 1rem;
		color: var(--color-text-muted);
		font-size: 1.1rem;
		line-height: 1.6;
	}

	.category-stats {
		display: flex;
		justify-content: center;
		gap: 1rem;
		color: var(--color-text-muted);
		font-size: 1.1rem;
	}

	.video-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: 2rem;
		margin-bottom: 2rem;
	}

	.empty-state {
		text-align: center;
		padding: 4rem 0;
		color: var(--color-text-muted);
	}

	.error-message {
		text-align: center;
		color: var(--color-error);
		padding: 1rem;
		margin: 1rem 0;
		background: var(--color-error-bg);
		border-radius: 0.5rem;
	}

	.load-more {
		text-align: center;
		margin-top: 2rem;
	}

	@media (max-width: 768px) {
		.video-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
