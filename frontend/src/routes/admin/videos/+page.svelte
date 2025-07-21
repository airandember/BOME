<script lang="ts">
	import { onMount } from 'svelte';
	import { MasterVideoService, type MasterVideo } from '$lib/master-video';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	const masterVideoService = new MasterVideoService();

	let videos: MasterVideo[] = [];
	let loading = true;
	let currentPage = 1;
	let pageSize = 20;
	let totalVideos = 0;
	let searchQuery = '';

	onMount(() => {
		loadVideos();
	});

	async function loadVideos() {
		try {
			loading = true;
			const response = await masterVideoService.getMasterVideos({
				page: currentPage,
				limit: pageSize,
				search: searchQuery || undefined
			});
			
			if (response.success) {
				videos = response.videos || [];
				totalVideos = response.pagination.total;
			}
		} catch (error) {
			showToast('Failed to load videos', 'error');
			console.error('Error loading videos:', error);
		} finally {
			loading = false;
		}
	}

	async function deleteVideo(videoId: number) {
		if (!confirm('Are you sure you want to delete this video? This action cannot be undone.')) {
			return;
		}

		try {
			const response = await masterVideoService.deleteMasterVideo(videoId);
			if (response.success) {
				showToast('Video deleted successfully', 'success');
				loadVideos();
			}
		} catch (error) {
			showToast('Failed to delete video', 'error');
			console.error('Error deleting video:', error);
		}
	}

	function formatDuration(seconds: number): string {
		if (!seconds) return '0:00';
		const hours = Math.floor(seconds / 3600);
		const minutes = Math.floor((seconds % 3600) / 60);
		const secs = seconds % 60;
		
		if (hours > 0) {
			return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
		}
		return `${minutes}:${secs.toString().padStart(2, '0')}`;
	}

	function formatFileSize(bytes: number): string {
		if (!bytes) return '0 B';
		const sizes = ['B', 'KB', 'MB', 'GB'];
		let size = bytes;
		let unitIndex = 0;
		
		while (size >= 1024 && unitIndex < sizes.length - 1) {
			size /= 1024;
			unitIndex++;
		}
		
		return `${size.toFixed(1)} ${sizes[unitIndex]}`;
	}

	function formatDate(dateString: string): string {
		if (!dateString) return 'N/A';
		return new Date(dateString).toLocaleDateString();
	}

	function getStatusColor(status: string): string {
		switch (status) {
			case 'ready': return 'success';
			case 'processing': case 'transcoding': return 'warning';
			case 'error': case 'upload_failed': return 'error';
			default: return 'info';
		}
	}

	function getSyncStatusColor(syncStatus: string): string {
		switch (syncStatus) {
			case 'synced': return 'success';
			case 'needs_attention': return 'warning';
			case 'conflict': return 'error';
			default: return 'info';
		}
	}

	$: totalPages = Math.ceil(totalVideos / pageSize);
</script>

<svelte:head>
	<title>Video Management - Admin Dashboard</title>
</svelte:head>

<div class="video-management">
	<div class="page-header">
		<div class="header-content">
			<div class="header-text">
				<h1>Video Management</h1>
				<p>Manage master video list</p>
			</div>
		</div>
	</div>

	<!-- Search -->
	<div class="search-section">
		<input
			type="text"
			placeholder="Search videos..."
			bind:value={searchQuery}
			on:input={() => { currentPage = 1; loadVideos(); }}
			class="search-input"
		/>
	</div>

	<!-- Videos List -->
	{#if loading}
		<div class="loading-container">
			<LoadingSpinner size="large" color="primary" />
			<p>Loading videos...</p>
		</div>
	{:else if videos.length === 0}
		<div class="empty-state">
			<h3>No videos found</h3>
			<p>No videos match your current search criteria.</p>
		</div>
	{:else}
		<div class="videos-container">
			{#each videos as video}
				<div class="video-row">
					<div class="video-thumbnail">
						{#if video.ThumbnailURL}
							<img src={video.ThumbnailURL} alt={video.Title} />
						{:else}
							<div class="no-thumbnail">No Image</div>
						{/if}
					</div>
					
					<div class="video-details">
						<h4 class="video-title">{video.Title}</h4>
						<p class="video-description">{video.Description}</p>
						<div class="video-meta">
							<span class="meta-item">Category: {video.Category}</span>
							<span class="meta-item">Views: {(video.Views || 0).toLocaleString()}</span>
							<span class="meta-item">Duration: {formatDuration(video.Duration)}</span>
							<span class="meta-item">Size: {formatFileSize(video.FileSize)}</span>
						</div>
					</div>
					
					<div class="video-status">
						<span class="status-badge {getStatusColor(video.Status)}">
							{video.Status}
						</span>
						<span class="status-badge {getSyncStatusColor(video.SyncStatus)}">
							{video.SyncStatus}
						</span>
					</div>
					
					<div class="video-actions">
						<button class="btn btn-secondary btn-sm" on:click={() => deleteVideo(video.ID)}>
							Delete
						</button>
					</div>
				</div>
			{/each}
		</div>

		<!-- Pagination -->
		{#if totalPages > 1}
			<div class="pagination">
				<button 
					class="btn btn-secondary btn-sm"
					disabled={currentPage === 1}
					on:click={() => {currentPage--; loadVideos();}}
				>
					Previous
				</button>
				
				<span class="page-info">
					Page {currentPage} of {totalPages} ({totalVideos} total)
				</span>
				
				<button 
					class="btn btn-secondary btn-sm"
					disabled={currentPage === totalPages}
					on:click={() => {currentPage++; loadVideos();}}
				>
					Next
				</button>
			</div>
		{/if}
	{/if}
</div>

<style>
	.video-management {
		padding: 2rem;
		max-width: 1200px;
		margin: 0 auto;
	}

	.page-header {
		margin-bottom: 2rem;
	}

	.header-text h1 {
		font-size: 2rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.header-text p {
		color: var(--text-secondary);
		margin: 0;
	}

	.search-section {
		margin-bottom: 2rem;
	}

	.search-input {
		width: 100%;
		max-width: 400px;
		padding: 0.75rem 1rem;
		border: 1px solid var(--border-color);
		border-radius: 6px;
		background: var(--bg-primary);
		color: var(--text-primary);
	}

	.loading-container,
	.empty-state {
		text-align: center;
		padding: 4rem 2rem;
	}

	.videos-container {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.video-row {
		display: grid;
		grid-template-columns: 120px 1fr 150px 100px;
		gap: 1rem;
		padding: 1rem;
		background: var(--bg-secondary);
		border: 1px solid var(--border-color);
		border-radius: 8px;
		align-items: center;
	}

	.video-thumbnail {
		width: 100px;
		height: 60px;
		border-radius: 4px;
		overflow: hidden;
	}

	.video-thumbnail img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.no-thumbnail {
		width: 100%;
		height: 100%;
		background: var(--bg-primary);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.75rem;
		color: var(--text-secondary);
	}

	.video-title {
		font-size: 1rem;
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 0.5rem 0;
	}

	.video-description {
		font-size: 0.875rem;
		color: var(--text-secondary);
		margin: 0 0 0.5rem 0;
	}

	.video-meta {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.meta-item {
		font-size: 0.75rem;
		color: var(--text-secondary);
	}

	.status-badge {
		display: block;
		padding: 0.25rem 0.75rem;
		border-radius: 12px;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		margin-bottom: 0.25rem;
	}

	.status-badge.success {
		background: var(--success-bg);
		color: var(--success-color);
	}

	.status-badge.warning {
		background: var(--warning-bg);
		color: var(--warning-color);
	}

	.status-badge.error {
		background: var(--error-bg);
		color: var(--error-color);
	}

	.status-badge.info {
		background: var(--info-bg);
		color: var(--info-color);
	}

	.pagination {
		display: flex;
		justify-content: center;
		align-items: center;
		gap: 1rem;
		margin-top: 2rem;
	}

	.page-info {
		color: var(--text-secondary);
		font-size: 0.875rem;
	}

	@media (max-width: 768px) {
		.video-row {
			grid-template-columns: 1fr;
			gap: 1rem;
		}

		.video-thumbnail {
			width: 100%;
			height: 200px;
		}
	}
</style> 