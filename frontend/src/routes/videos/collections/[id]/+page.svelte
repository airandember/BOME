<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { videoService, type Video, type BunnyCollection } from '$lib/video';
	import VideoCard from '$lib/components/VideoCard.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import Navigation from '$lib/components/Navigation.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import { toastStore } from '$lib/stores/toast';

	let collection: BunnyCollection | null = null;
	let videos: Video[] = [];
	let loading = true;
	let error = '';

	onMount(async () => {
		const collectionId = $page.params.id;
		if (!collectionId) return;

		try {
			// Load collection details
			collection = await videoService.getCollection(collectionId);
			
			// Load videos from this collection
			const response = await videoService.getVideos(1, 50); // Load more videos for collections
			videos = response.videos || [];
			
		} catch (err: any) {
			console.error('Error loading collection:', err);
			error = err.message || 'Failed to load collection';
			toastStore.error(error);
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>{collection?.name || 'Collection'} - Book of Mormon Evidences</title>
	<meta name="description" content="Browse videos in the {collection?.name} collection" />
</svelte:head>

<Navigation />

<div class="page-wrapper">
	<main class="main-content-wrapper">
		{#if loading}
			<div class="loading-container">
				<LoadingSpinner size="large" />
				<p>Loading collection...</p>
			</div>
		{:else if error}
			<div class="error-message">
				<p>{error}</p>
			</div>
		{:else}
			<div class="collection-page">
				<div class="container">
					<header class="collection-header">
						<h1>{collection?.name}</h1>
						<div class="collection-stats">
							<span>{collection?.videoCount} videos</span>
							<span>•</span>
							<span>{Math.round((collection?.totalSize || 0) / 1024 / 1024)}MB</span>
						</div>
					</header>

					<div class="video-grid">
						{#each videos as video (video.id)}
							<VideoCard {video} />
						{/each}
					</div>

					{#if videos.length === 0}
						<div class="empty-state">
							<p>No videos found in this collection</p>
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

	.collection-page {
		padding: 2rem 0;
	}

	.collection-header {
		text-align: center;
		margin-bottom: 2rem;

		h1 {
			font-size: 2.5rem;
			color: var(--color-primary);
			margin-bottom: 0.5rem;
		}
	}

	.collection-stats {
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

	@media (max-width: 768px) {
		.video-grid {
			grid-template-columns: 1fr;
		}
	}
</style> 