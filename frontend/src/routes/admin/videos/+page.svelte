<script lang="ts">
	import { onMount } from 'svelte';
	import { videoService } from '$lib/video';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	interface Video {
		id: number;
		title: string;
		description: string;
		duration: string;
		thumbnail: string;
		status: 'published' | 'pending' | 'rejected' | 'draft';
		category: string;
		uploaded_by: {
			id: number;
			name: string;
			email: string;
		};
		upload_date: string;
		views: number;
		likes: number;
		comments: number;
		file_size: string;
		resolution: string;
		tags: string[];
	}

	interface Category {
		id: number;
		name: string;
		description: string;
		video_count: number;
	}

	let videos: Video[] = [];
	let categories: Category[] = [];
	let loading = true;
	let selectedVideos: Set<number> = new Set();
	let currentView = 'all';
	let searchQuery = '';
	let selectedCategory = '';
	let selectedStatus = '';
	let sortBy = 'upload_date';
	let sortOrder = 'desc';
	let currentPage = 1;
	let pageSize = 20;
	let totalVideos = 0;

	// Upload modal state
	let showUploadModal = false;
	let uploading = false;
	let uploadProgress = 0;

	// Edit modal state
	let showEditModal = false;
	let editingVideo: Video | null = null;

	// Bulk operations
	let showBulkActions = false;
	let bulkOperation = '';

	onMount(() => {
		loadVideos();
		loadCategories();
	});

	async function loadVideos() {
		try {
			loading = true;
			const response = await videoService.admin.getVideos(currentPage, pageSize, {
				search: searchQuery,
				category: selectedCategory,
				status: selectedStatus,
				sort: sortBy,
				order: sortOrder
			});
			
			videos = response.videos || [];
			totalVideos = response.pagination?.total || 0;
		} catch (error) {
			showToast('Failed to load videos', 'error');
			console.error('Error loading videos:', error);
		} finally {
			loading = false;
		}
	}

	async function loadCategories() {
		try {
			const response = await videoService.admin.getCategories();
			categories = response.categories || [];
		} catch (error) {
			console.error('Error loading categories:', error);
		}
	}

	async function updateVideoStatus(videoId: number, status: 'published' | 'rejected') {
		try {
			if (status === 'published') {
				await videoService.admin.approveVideo(videoId);
			} else {
				await videoService.admin.rejectVideo(videoId);
			}
			showToast(`Video ${status} successfully`, 'success');
			loadVideos();
		} catch (error) {
			showToast(`Failed to ${status} video`, 'error');
			console.error(`Error ${status} video:`, error);
		}
	}

	async function deleteVideo(videoId: number) {
		if (!confirm('Are you sure you want to delete this video? This action cannot be undone.')) {
			return;
		}

		try {
			await videoService.admin.deleteVideo(videoId);
			showToast('Video deleted successfully', 'success');
			loadVideos();
		} catch (error) {
			showToast('Failed to delete video', 'error');
			console.error('Error deleting video:', error);
		}
	}

	async function handleBulkOperation() {
		if (selectedVideos.size === 0) {
			showToast('Please select videos first', 'warning');
			return;
		}

		if (!confirm(`Are you sure you want to ${bulkOperation} ${selectedVideos.size} video(s)?`)) {
			return;
		}

		try {
			const videoIds = Array.from(selectedVideos);
			await videoService.admin.bulkOperation(bulkOperation as 'publish' | 'unpublish' | 'delete', videoIds);
			
			showToast(`Bulk ${bulkOperation} completed successfully`, 'success');
			selectedVideos.clear();
			selectedVideos = selectedVideos; // Trigger reactivity
			showBulkActions = false;
			loadVideos();
		} catch (error) {
			showToast(`Failed to perform bulk ${bulkOperation}`, 'error');
			console.error('Error performing bulk operation:', error);
		}
	}

	function toggleVideoSelection(videoId: number) {
		if (selectedVideos.has(videoId)) {
			selectedVideos.delete(videoId);
		} else {
			selectedVideos.add(videoId);
		}
		selectedVideos = selectedVideos; // Trigger reactivity
		showBulkActions = selectedVideos.size > 0;
	}

	function selectAllVideos() {
		if (selectedVideos.size === videos.length) {
			selectedVideos.clear();
		} else {
			videos.forEach(video => selectedVideos.add(video.id));
		}
		selectedVideos = selectedVideos; // Trigger reactivity
		showBulkActions = selectedVideos.size > 0;
	}

	function formatFileSize(bytes: string) {
		const size = parseInt(bytes);
		if (size === 0) return '0 Bytes';
		const k = 1024;
		const sizes = ['Bytes', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(size) / Math.log(k));
		return parseFloat((size / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	}

	function formatDate(dateString: string) {
		return new Date(dateString).toLocaleDateString();
	}

	function getStatusColor(status: string) {
		switch (status) {
			case 'published': return 'success';
			case 'pending': return 'warning';
			case 'rejected': return 'error';
			case 'draft': return 'info';
			default: return 'secondary';
		}
	}

	// Reactive statements
	$: filteredVideos = videos.filter(video => {
		if (currentView !== 'all' && video.status !== currentView) return false;
		return true;
	});

	$: totalPages = Math.ceil(totalVideos / pageSize);

	// Watch for filter changes
	$: {
		if (searchQuery !== undefined || selectedCategory !== undefined || selectedStatus !== undefined) {
			currentPage = 1;
			loadVideos();
		}
	}
</script>

<svelte:head>
	<title>Video Management - Admin Dashboard</title>
</svelte:head>

<div class="video-management">
	<div class="page-header">
		<div class="header-content">
			<div class="header-text">
				<h1>Video Management</h1>
				<p>Manage and moderate video content</p>
			</div>
			
			<div class="header-actions">
				<button class="btn btn-primary" on:click={() => showUploadModal = true}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
						<polyline points="17,8 12,3 7,8"></polyline>
						<line x1="12" y1="3" x2="12" y2="15"></line>
					</svg>
					Upload Video
				</button>
			</div>
		</div>
	</div>

	<!-- Filters and Search -->
	<div class="filters-section glass">
		<div class="search-filters">
			<div class="search-box">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="11" cy="11" r="8"></circle>
					<path d="M21 21l-4.35-4.35"></path>
				</svg>
				<input 
					type="text" 
					placeholder="Search videos..." 
					bind:value={searchQuery}
				/>
			</div>
			
			<select bind:value={selectedCategory} class="filter-select">
				<option value="">All Categories</option>
				{#each categories as category}
					<option value={category.name}>{category.name}</option>
				{/each}
			</select>
			
			<select bind:value={selectedStatus} class="filter-select">
				<option value="">All Statuses</option>
				<option value="published">Published</option>
				<option value="pending">Pending</option>
				<option value="rejected">Rejected</option>
				<option value="draft">Draft</option>
			</select>
			
			<select bind:value={sortBy} class="filter-select">
				<option value="upload_date">Upload Date</option>
				<option value="title">Title</option>
				<option value="views">Views</option>
				<option value="likes">Likes</option>
				<option value="duration">Duration</option>
			</select>
			
			<button 
				class="sort-order-btn"
				class:desc={sortOrder === 'desc'}
				on:click={() => sortOrder = sortOrder === 'asc' ? 'desc' : 'asc'}
			>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					{#if sortOrder === 'asc'}
						<polyline points="6,9 12,15 18,9"></polyline>
					{:else}
						<polyline points="18,15 12,9 6,15"></polyline>
					{/if}
				</svg>
			</button>
		</div>
		
		<!-- View Tabs -->
		<div class="view-tabs">
			<button 
				class="tab-btn" 
				class:active={currentView === 'all'}
				on:click={() => currentView = 'all'}
			>
				All ({totalVideos})
			</button>
			<button 
				class="tab-btn" 
				class:active={currentView === 'published'}
				on:click={() => currentView = 'published'}
			>
				Published
			</button>
			<button 
				class="tab-btn" 
				class:active={currentView === 'pending'}
				on:click={() => currentView = 'pending'}
			>
				Pending
			</button>
			<button 
				class="tab-btn" 
				class:active={currentView === 'rejected'}
				on:click={() => currentView = 'rejected'}
			>
				Rejected
			</button>
			<button 
				class="tab-btn" 
				class:active={currentView === 'draft'}
				on:click={() => currentView = 'draft'}
			>
				Draft
			</button>
		</div>
	</div>

	<!-- Bulk Actions Bar -->
	{#if showBulkActions}
		<div class="bulk-actions glass">
			<div class="bulk-info">
				<span>{selectedVideos.size} video(s) selected</span>
			</div>
			<div class="bulk-controls">
				<select bind:value={bulkOperation} class="bulk-select">
					<option value="">Choose action...</option>
					<option value="publish">Publish</option>
					<option value="reject">Reject</option>
					<option value="delete">Delete</option>
					<option value="archive">Archive</option>
				</select>
				<button 
					class="btn btn-primary btn-sm"
					disabled={!bulkOperation}
					on:click={handleBulkOperation}
				>
					Apply
				</button>
				<button 
					class="btn btn-secondary btn-sm"
					on:click={() => {
						selectedVideos.clear();
						selectedVideos = selectedVideos;
						showBulkActions = false;
					}}
				>
					Cancel
				</button>
			</div>
		</div>
	{/if}

	<!-- Videos List -->
	{#if loading}
		<div class="loading-container">
			<LoadingSpinner size="large" color="primary" />
			<p>Loading videos...</p>
		</div>
	{:else if videos.length === 0}
		<div class="empty-state glass">
			<div class="empty-icon">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<polygon points="23,7 16,12 23,17 23,7"></polygon>
					<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
				</svg>
			</div>
			<h3>No videos found</h3>
			<p>No videos match your current filters.</p>
			<button class="btn btn-primary" on:click={() => showUploadModal = true}>
				Upload First Video
			</button>
		</div>
	{:else}
		<div class="videos-container">
			<!-- Table Header -->
			<div class="videos-header glass">
				<div class="header-checkbox">
					<input 
						type="checkbox" 
						checked={selectedVideos.size === videos.length && videos.length > 0}
						on:change={selectAllVideos}
					/>
				</div>
				<div class="header-thumbnail">Thumbnail</div>
				<div class="header-details">Details</div>
				<div class="header-stats">Stats</div>
				<div class="header-status">Status</div>
				<div class="header-actions">Actions</div>
			</div>
			
			<!-- Video Rows -->
			<div class="videos-list">
				{#each filteredVideos as video}
					<div class="video-row glass">
						<div class="row-checkbox">
							<input 
								type="checkbox" 
								checked={selectedVideos.has(video.id)}
								on:change={() => toggleVideoSelection(video.id)}
							/>
						</div>
						
						<div class="row-thumbnail">
							<img src={video.thumbnail} alt={video.title} />
							<div class="duration-badge">{video.duration}</div>
						</div>
						
						<div class="row-details">
							<h4 class="video-title">{video.title}</h4>
							<p class="video-description">{video.description}</p>
							<div class="video-meta">
								<span class="meta-item">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"></path>
										<rect x="8" y="2" width="8" height="4" rx="1" ry="1"></rect>
									</svg>
									{video.category}
								</span>
								<span class="meta-item">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
										<circle cx="12" cy="7" r="4"></circle>
									</svg>
									{video.uploaded_by.name}
								</span>
								<span class="meta-item">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
										<line x1="16" y1="2" x2="16" y2="6"></line>
										<line x1="8" y1="2" x2="8" y2="6"></line>
										<line x1="3" y1="10" x2="21" y2="10"></line>
									</svg>
									{formatDate(video.upload_date)}
								</span>
							</div>
							<div class="video-tags">
								{#each video.tags as tag}
									<span class="tag">{tag}</span>
								{/each}
							</div>
						</div>
						
						<div class="row-stats">
							<div class="stat-item">
								<span class="stat-value">{video.views.toLocaleString()}</span>
								<span class="stat-label">Views</span>
							</div>
							<div class="stat-item">
								<span class="stat-value">{video.likes.toLocaleString()}</span>
								<span class="stat-label">Likes</span>
							</div>
							<div class="stat-item">
								<span class="stat-value">{video.comments.toLocaleString()}</span>
								<span class="stat-label">Comments</span>
							</div>
							<div class="stat-item">
								<span class="stat-value">{formatFileSize(video.file_size)}</span>
								<span class="stat-label">Size</span>
							</div>
						</div>
						
						<div class="row-status">
							<span class="status-badge {getStatusColor(video.status)}">
								{video.status}
							</span>
						</div>
						
						<div class="row-actions">
							<div class="action-buttons">
								{#if video.status === 'pending'}
									<button 
										class="btn btn-success btn-sm"
										on:click={() => updateVideoStatus(video.id, 'published')}
										title="Approve"
									>
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<polyline points="20,6 9,17 4,12"></polyline>
										</svg>
									</button>
									<button 
										class="btn btn-error btn-sm"
										on:click={() => updateVideoStatus(video.id, 'rejected')}
										title="Reject"
									>
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<line x1="18" y1="6" x2="6" y2="18"></line>
											<line x1="6" y1="6" x2="18" y2="18"></line>
										</svg>
									</button>
								{/if}
								<button 
									class="btn btn-secondary btn-sm"
									on:click={() => {
										editingVideo = video;
										showEditModal = true;
									}}
									title="Edit"
								>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
										<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
									</svg>
								</button>
								<button 
									class="btn btn-error btn-sm"
									on:click={() => deleteVideo(video.id)}
									title="Delete"
								>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<polyline points="3,6 5,6 21,6"></polyline>
										<path d="M19,6v14a2,2 0 0,1 -2,2H7a2,2 0 0,1 -2,-2V6m3,0V4a2,2 0 0,1 2,-2h4a2,2 0 0,1 2,2v2"></path>
									</svg>
								</button>
							</div>
						</div>
					</div>
				{/each}
			</div>
		</div>

		<!-- Pagination -->
		{#if totalPages > 1}
			<div class="pagination glass">
				<button 
					class="btn btn-secondary btn-sm"
					disabled={currentPage === 1}
					on:click={() => {currentPage = 1; loadVideos();}}
				>
					First
				</button>
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
				<button 
					class="btn btn-secondary btn-sm"
					disabled={currentPage === totalPages}
					on:click={() => {currentPage = totalPages; loadVideos();}}
				>
					Last
				</button>
			</div>
		{/if}
	{/if}
</div>

<style>
	.video-management {
		display: flex;
		flex-direction: column;
		gap: var(--space-xl);
	}

	.page-header {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.header-content {
		display: flex;
		align-items: center;
		justify-content: space-between;
		flex-wrap: wrap;
		gap: var(--space-lg);
	}

	.header-text h1 {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: var(--space-sm);
	}

	.header-text p {
		color: var(--text-secondary);
		font-size: var(--text-lg);
		margin: 0;
	}

	.header-actions {
		display: flex;
		gap: var(--space-md);
	}

	.filters-section {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.search-filters {
		display: flex;
		gap: var(--space-md);
		align-items: center;
		flex-wrap: wrap;
	}

	.search-box {
		position: relative;
		flex: 1;
		min-width: 300px;
	}

	.search-box svg {
		position: absolute;
		left: var(--space-md);
		top: 50%;
		transform: translateY(-50%);
		width: 20px;
		height: 20px;
		color: var(--text-secondary);
	}

	.search-box input {
		width: 100%;
		padding: var(--space-md) var(--space-md) var(--space-md) 3rem;
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		background: var(--bg-glass);
		color: var(--text-primary);
		font-size: var(--text-sm);
	}

	.filter-select {
		padding: var(--space-md);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		background: var(--bg-glass);
		color: var(--text-primary);
		font-size: var(--text-sm);
		min-width: 150px;
	}

	.sort-order-btn {
		padding: var(--space-md);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		background: var(--bg-glass);
		color: var(--text-primary);
		cursor: pointer;
		transition: all var(--transition-fast);
	}

	.sort-order-btn:hover {
		background: var(--bg-hover);
	}

	.sort-order-btn svg {
		width: 20px;
		height: 20px;
	}

	.view-tabs {
		display: flex;
		gap: var(--space-sm);
		flex-wrap: wrap;
	}

	.tab-btn {
		padding: var(--space-sm) var(--space-lg);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		background: transparent;
		color: var(--text-secondary);
		font-size: var(--text-sm);
		cursor: pointer;
		transition: all var(--transition-fast);
	}

	.tab-btn.active {
		background: var(--primary);
		color: var(--white);
		border-color: var(--primary);
	}

	.tab-btn:hover:not(.active) {
		background: var(--bg-hover);
		color: var(--text-primary);
	}

	.bulk-actions {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: var(--space-lg);
		border-radius: var(--radius-lg);
		background: var(--warning);
		color: var(--white);
	}

	.bulk-info {
		font-weight: 600;
	}

	.bulk-controls {
		display: flex;
		gap: var(--space-md);
		align-items: center;
	}

	.bulk-select {
		padding: var(--space-sm) var(--space-md);
		border: 1px solid rgba(255, 255, 255, 0.3);
		border-radius: var(--radius-md);
		background: rgba(255, 255, 255, 0.1);
		color: var(--white);
		font-size: var(--text-sm);
	}

	.loading-container,
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-lg);
		padding: var(--space-3xl);
		text-align: center;
	}

	.empty-state {
		border-radius: var(--radius-xl);
	}

	.empty-icon {
		width: 64px;
		height: 64px;
		opacity: 0.5;
	}

	.empty-icon svg {
		width: 100%;
		height: 100%;
		color: var(--text-secondary);
	}

	.videos-container {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.videos-header {
		display: grid;
		grid-template-columns: 40px 120px 1fr 200px 120px 150px;
		gap: var(--space-md);
		padding: var(--space-lg);
		border-radius: var(--radius-lg);
		font-weight: 600;
		color: var(--text-primary);
		font-size: var(--text-sm);
		align-items: center;
	}

	.videos-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.video-row {
		display: grid;
		grid-template-columns: 40px 120px 1fr 200px 120px 150px;
		gap: var(--space-md);
		padding: var(--space-lg);
		border-radius: var(--radius-lg);
		align-items: center;
		transition: all var(--transition-fast);
	}

	.video-row:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}

	.row-thumbnail {
		position: relative;
		width: 100px;
		height: 60px;
		border-radius: var(--radius-md);
		overflow: hidden;
	}

	.row-thumbnail img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.duration-badge {
		position: absolute;
		bottom: 4px;
		right: 4px;
		padding: 2px 6px;
		background: rgba(0, 0, 0, 0.8);
		color: var(--white);
		font-size: var(--text-xs);
		border-radius: var(--radius-sm);
	}

	.row-details {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
		min-width: 0;
	}

	.video-title {
		font-size: var(--text-md);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.video-description {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		margin: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.video-meta {
		display: flex;
		gap: var(--space-md);
		flex-wrap: wrap;
	}

	.meta-item {
		display: flex;
		align-items: center;
		gap: var(--space-xs);
		font-size: var(--text-xs);
		color: var(--text-secondary);
	}

	.meta-item svg {
		width: 14px;
		height: 14px;
	}

	.video-tags {
		display: flex;
		gap: var(--space-xs);
		flex-wrap: wrap;
	}

	.tag {
		padding: 2px 6px;
		background: rgba(102, 126, 234, 0.1);
		color: var(--primary);
		font-size: var(--text-xs);
		border-radius: var(--radius-sm);
	}

	.row-stats {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.stat-item {
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
	}

	.stat-value {
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--text-primary);
	}

	.stat-label {
		font-size: var(--text-xs);
		color: var(--text-secondary);
	}

	.status-badge {
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-full);
		font-size: var(--text-xs);
		font-weight: 600;
		text-transform: uppercase;
	}

	.status-badge.success {
		background: rgba(67, 233, 123, 0.1);
		color: var(--success);
	}

	.status-badge.warning {
		background: rgba(255, 167, 38, 0.1);
		color: var(--warning);
	}

	.status-badge.error {
		background: rgba(255, 107, 107, 0.1);
		color: var(--error);
	}

	.status-badge.info {
		background: rgba(102, 126, 234, 0.1);
		color: var(--primary);
	}

	.action-buttons {
		display: flex;
		gap: var(--space-xs);
	}

	.action-buttons .btn {
		padding: var(--space-xs);
	}

	.action-buttons .btn svg {
		width: 16px;
		height: 16px;
	}

	.pagination {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: var(--space-md);
		padding: var(--space-lg);
		border-radius: var(--radius-lg);
	}

	.page-info {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		margin: 0 var(--space-lg);
	}

	/* Responsive Design */
	@media (max-width: 1200px) {
		.videos-header,
		.video-row {
			grid-template-columns: 40px 100px 1fr 150px 100px 120px;
		}
	}

	@media (max-width: 768px) {
		.header-content {
			flex-direction: column;
			align-items: stretch;
		}

		.search-filters {
			flex-direction: column;
			align-items: stretch;
		}

		.search-box {
			min-width: unset;
		}

		.videos-header {
			display: none;
		}

		.video-row {
			grid-template-columns: 1fr;
			gap: var(--space-lg);
		}

		.row-thumbnail {
			width: 100%;
			height: 200px;
		}

		.action-buttons {
			justify-content: center;
		}
	}
</style> 