<script lang="ts">
	import { onMount } from 'svelte';
	import { masterVideoService, type MasterVideo } from '$lib/master-video';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import SmartTaggingModal from '$lib/components/SmartTaggingModal.svelte';
	import TagAnalyticsModal from '$lib/components/TagAnalyticsModal.svelte';
	import VideoTable from '$lib/components/VideoTable.svelte';
	import VideoStats from '$lib/components/VideoStats.svelte';
	import VideoFilters from '$lib/components/VideoFilters.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import VideoModals from '$lib/components/VideoModals.svelte';

	// Reactive variables
	let isLoading = true;
	let videos: MasterVideo[] = [];
	let error: string | null = null;
	let showUploadModal = false;
	let showEditModal = false;
	let showCreateModal = false;
	let selectedVideo: MasterVideo | null = null;
	let isUploading = false;
	let uploadProgress = 0;
	let checkingConflicts = false;

	// Conflict check results
	let lastConflictCheck: any = null;

	// Analytics data
	let analytics: any = null;
	let analyticsLoading = false;

	// Loading state for individual video toggles
	let togglingVideos = new Set<number>();

	// Smart tagging state
	let showTaggingModal = false;
	let selectedVideoForTagging: MasterVideo | null = null;
	let showTagAnalytics = false;

	// Pagination
	let currentPage = 1;
	let pageSize = 20;
	let totalVideos = 0;
	let totalPages = 0;
	let AbsoluteTotalVideos = 0;

	// Upload form data
	let uploadForm = {
		title: '',
		description: '',
		category: '',
		tags: '',
		videoFile: null as File | null
	};

	// Create form data (for new video creation)
	let createForm = {
		title: '',
		description: '',
		category: '',
		tags: '',
		videoFile: null as File | null
	};

	// Edit form data
	let editForm = {
		title: '',
		description: '',
		category: '',
		tags: '',
		status: 'active'
	};

	// Filter and search
	let searchTerm = '';
	let statusFilter = 'all';
	let categoryFilter = 'all';
	let syncStatusFilter = 'all';
	let vidStatusFilter = 'all';

	// Sorting
	let sortField: string | null = null;
	let sortDirection: 'asc' | 'desc' = 'asc';

	// Debounced search
	let searchTimeout: ReturnType<typeof setTimeout> | null = null;

	// Available categories
	const categories = [
		'Archaeology',
		'Geography',
		'DNA Research',
		'Linguistics',
		'Historical Evidence',
		'Cultural Studies',
		'Religious Studies',
		'Documentary',
		'Lecture',
		'Interview',
		'Presentation',
		'Virtual Tour'
	];

	// Load analytics data
	async function loadAnalytics() {
		try {
			analyticsLoading = true;
			const response = await masterVideoService.getDashboardAnalytics();
			if (response.success) {
				analytics = response.data;
			} else {
				// Fallback to basic stats
				const statsResponse = await masterVideoService.getStats();
				if (statsResponse.success) {
					analytics = {
						video_stats: {
							total_videos: statsResponse.stats.total_videos || 0,
							synced_videos: statsResponse.stats.videos_by_sync_status?.synced || 0,
							needs_attention: statsResponse.stats.videos_by_sync_status?.needs_attention || 0,
							videos_by_sync_status: statsResponse.stats.videos_by_sync_status || {}
						}
					};
				}
			}
		} catch (error) {
			console.error('Error loading analytics:', error);
		} finally {
			analyticsLoading = false;
		}
	}

	// Load videos using MasterVideoService
	async function loadVideos() {
		try {
			isLoading = true;
			error = null;

			const response = await masterVideoService.getMasterVideos({
				page: currentPage,
				limit: pageSize,
				search: searchTerm || undefined,
				status: statusFilter !== 'all' ? statusFilter : undefined,
				category: categoryFilter !== 'all' ? categoryFilter : undefined,
				sync_status: syncStatusFilter !== 'all' ? syncStatusFilter : undefined,
				vid_status: vidStatusFilter !== 'all' ? vidStatusFilter : undefined,
				sort_field: sortField || undefined,
				sort_direction: sortDirection
			});

			if (response.success) {
				videos = response.videos || [];
				totalVideos = response.pagination.total;
				totalPages = response.pagination.total_pages;
				console.log(response);
				if (AbsoluteTotalVideos != 0) {
					AbsoluteTotalVideos = AbsoluteTotalVideos;
				} else {
					AbsoluteTotalVideos = totalVideos;
				};
				
				// Also refresh analytics when videos are loaded
				await loadAnalytics();
			} else {
				error = 'Failed to load videos';
			}
		} catch (err: unknown) {
			console.error('Error loading videos:', err);
			error = err instanceof Error ? err.message : 'Failed to load videos';
		} finally {
			isLoading = false;
		}
	}

	// Check Conflicts between Bunny and Local, and sync if master list is empty
	async function checkConflicts() {
		try {
			checkingConflicts = true;
			
			// If master list is empty, sync from Bunny.net first
			if (videos.length === 0 && totalVideos === 0) {
				showToast('Master list is empty. Syncing videos from Bunny.net...', 'info');
				
				const syncResponse = await masterVideoService.syncFromBunny();
				if (syncResponse.success) {
					showToast(`Successfully synced ${syncResponse.result.synced || 0} videos from Bunny.net`, 'success');
					
					// Reload videos after sync
					await loadVideos();
					await loadAnalytics();
					
					// Now check for conflicts after sync
					showToast('Checking for conflicts...', 'info');
				} else {
					showToast('Failed to sync from Bunny.net', 'error');
					return;
				}
			} else {
				showToast('Checking for conflicts...', 'info');
			}
			
			// Check for conflicts
			const response = await masterVideoService.checkConflicts();
			if (response.success) {
				lastConflictCheck = response.result;
				const message = response.result.conflict_count > 0 
					? `Found ${response.result.conflict_count} conflicts that need attention`
					: 'No conflicts found - all videos are in sync';
				const type = response.result.conflict_count > 0 ? 'warning' : 'success';
				
				showToast(message, type);
				
				// Reload data to show any updates
				await loadVideos();
				await loadAnalytics();
			} else {
				showToast('Conflict check failed', 'error');
			}
		} catch (error) {
			showToast('Failed to check conflicts', 'error');
			console.error('Error checking conflicts:', error);
		} finally {
			checkingConflicts = false;
		}
	}

	// Upload video
	async function uploadVideo() {
		if (!uploadForm.videoFile) {
			showToast('Please select a video file', 'error');
			return;
		}

		if (!uploadForm.title.trim()) {
			showToast('Please enter a title', 'error');
			return;
		}

		try {
			isUploading = true;
			uploadProgress = 0;

			const formData = new FormData();
			formData.append('video', uploadForm.videoFile);
			formData.append('title', uploadForm.title);
			formData.append('description', uploadForm.description);
			formData.append('category', uploadForm.category);
			formData.append('tags', uploadForm.tags);

			const response = await masterVideoService.uploadVideo(formData);

			if (response.success) {
				showToast('Video uploaded successfully', 'success');
				showUploadModal = false;
				resetUploadForm();
				await loadVideos();
			} else {
				throw new Error(response.error || 'Upload failed');
			}
		} catch (err: unknown) {
			console.error('Error uploading video:', err);
			showToast(err instanceof Error ? err.message : 'Upload failed', 'error');
		} finally {
			isUploading = false;
			uploadProgress = 0;
		}
	}

	// Edit video
	function editVideo(video: MasterVideo) {
		selectedVideo = video;
		editForm = {
			title: video.Title,
			description: video.Description,
			category: video.Category,
			tags: video.Tags?.join(', ') || '',
			status: video.Status
		};
		showEditModal = true;
	}

	// Save edit
	async function saveEdit() {
		if (!selectedVideo) return;

		try {
			const response = await masterVideoService.updateMasterVideo(selectedVideo.ID, {
				Title: editForm.title,
				Description: editForm.description,
				Category: editForm.category,
				Tags: editForm.tags.split(',').map(tag => tag.trim()).filter(tag => tag.length > 0),
				Status: editForm.status
			});

			if (response.success) {
				showToast('Video updated successfully', 'success');
				showEditModal = false;
				await loadVideos();
			} else {
				throw new Error(response.message || 'Update failed');
			}
		} catch (err: unknown) {
			console.error('Error updating video:', err);
			showToast(err instanceof Error ? err.message : 'Update failed', 'error');
		}
	}

	// Toggle video status (vid_status) - Immutable update approach
	async function toggleVideoStatus(video: MasterVideo) {
		const newStatus = !video.Vid_Status;
		const statusText = newStatus ? 'activate' : 'deactivate';

		// Add to loading set
		togglingVideos.add(video.ID);

		try {
			const response = await masterVideoService.toggleVideoStatus(video.ID, newStatus);
			
			if (response.success) {
				// IMMUTABLE UPDATE - This prevents page jumping!
				videos = videos.map(v => 
					v.ID === video.ID 
						? { ...v, Vid_Status: newStatus }
						: v
				);
				showToast(`Video ${statusText}d successfully`, 'success');
			} else {
				throw new Error(response.message || `${statusText} failed`);
			}
		} catch (err: unknown) {
			console.error(`Error ${statusText}ing video:`, err);
			showToast(err instanceof Error ? err.message : `${statusText} failed`, 'error');
		} finally {
			// Remove from loading set
			togglingVideos.delete(video.ID);
		}
	}

	// Reset upload form
	function resetUploadForm() {
		uploadForm = {
			title: '',
			description: '',
			category: '',
			tags: '',
			videoFile: null
		};
	}

	// Handle page change
	async function changePage(page: number) {
		if (page >= 1 && page <= totalPages) {
			currentPage = page;
			await loadVideos();
		}
	}

	// Handle search with debouncing
	function handleSearchInput() {
		// Clear existing timeout
		if (searchTimeout) {
			clearTimeout(searchTimeout);
		}
		
		// Set new timeout for debounced search
		searchTimeout = setTimeout(() => {
			currentPage = 1;
			loadVideos();
		}, 300); // 300ms delay
	}

	// Handle search (for immediate search when needed)
	async function handleSearch() {
		if (searchTimeout) {
			clearTimeout(searchTimeout);
		}
		currentPage = 1;
		await loadVideos();
	}

	// Handle filter changes
	async function handleFilterChange() {
		currentPage = 1;
		await loadVideos();
	}

	// Handle sorting
	function handleSort(field: string) {
		if (sortField === field) {
			sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
		} else {
			sortField = field;
			sortDirection = 'asc';
		}
		loadVideos();
	}

	// Open create modal
	function openCreateModal() {
		showCreateModal = true;
	}

	// Close create modal
	function closeCreateModal() {
		showCreateModal = false;
		createForm = {
			title: '',
			description: '',
			category: '',
			tags: '',
			videoFile: null
		};
	}

	// Create new video
	async function createVideo() {
		if (!createForm.videoFile) {
			showToast('Please select a video file', 'error');
			return;
		}

		if (!createForm.title.trim()) {
			showToast('Please enter a title', 'error');
			return;
		}

		try {
			isUploading = true;
			uploadProgress = 0;

			const formData = new FormData();
			formData.append('video', createForm.videoFile);
			formData.append('title', createForm.title);
			formData.append('description', createForm.description);
			formData.append('category', createForm.category);
			formData.append('tags', createForm.tags);

			const response = await masterVideoService.uploadVideo(formData);

			if (response.success) {
				showToast('Video created successfully', 'success');
				closeCreateModal();
				await loadVideos();
			} else {
				throw new Error(response.error || 'Creation failed');
			}
		} catch (err: unknown) {
			console.error('Error creating video:', err);
			showToast(err instanceof Error ? err.message : 'Creation failed', 'error');
		} finally {
			isUploading = false;
			uploadProgress = 0;
		}
	}

	onMount(() => {
		loadAnalytics();
		loadVideos();
	});

	// Smart Tagging Functions
	function openTaggingModal(video: MasterVideo) {
		selectedVideoForTagging = video;
		showTaggingModal = true;
	}

	function closeTaggingModal() {
		showTaggingModal = false;
		selectedVideoForTagging = null;
	}

	function showTagAnalyticsModal() {
		showTagAnalytics = true;
	}

	function closeTagAnalyticsModal() {
		showTagAnalytics = false;
	}

	function handleVideoUpdated(event: CustomEvent<{ video: MasterVideo; tags: string[] }>) {
		const { video: updatedVideo, tags } = event.detail;
		
		// Update the video in the list
		const videoIndex = videos.findIndex(v => v.ID === updatedVideo.ID);
		if (videoIndex !== -1) {
			videos[videoIndex] = updatedVideo;
			videos = [...videos]; // Trigger reactivity
		}
		
		showToast(`✅ Video "${updatedVideo.Title}" tagged with ${tags.length} tags`, 'success');
	}
</script>

<svelte:head>
	<title>Video Management - Streaming Admin</title>
</svelte:head>

{#if isLoading}
	<div class="flex items-center justify-center py-12">
		<LoadingSpinner />
	</div>
{:else if error}
	<div class="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
		<div class="flex items-center">
			<svg class="h-5 w-5 text-red-400 mr-2" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
				<line x1="12" y1="9" x2="12" y2="13"/>
				<line x1="12" y1="17" x2="12.01" y2="17"/>
			</svg>
			<p class="text-red-800">{error}</p>
		</div>
	</div>
{:else}
	<div class="streaming-page">
		<!-- Video Stats -->
		<VideoStats
			{analytics}
			{AbsoluteTotalVideos}
			{checkingConflicts}
			on:checkConflicts={checkConflicts}
			on:showTagAnalytics={showTagAnalyticsModal}
		/>

		<!-- Video Filters -->
		<VideoFilters
			bind:searchTerm
			bind:statusFilter
			bind:categoryFilter
			bind:syncStatusFilter
			bind:vidStatusFilter
			{totalVideos}
			{categories}
			on:searchInput={handleSearchInput}
			on:filterChange={handleFilterChange}
			on:refresh={loadVideos}
		/>

		<!-- Videos Table -->
		<VideoTable
			{videos}
			{currentPage}
			{totalPages}
			{totalVideos}
			{sortField}
			{sortDirection}
			{togglingVideos}
			on:editVideo={(e) => editVideo(e.detail)}
			on:toggleVideoStatus={(e) => toggleVideoStatus(e.detail)}
			on:changePage={(e) => changePage(e.detail)}
			on:handleSort={(e) => handleSort(e.detail)}
			on:tagVideo={(e) => openTaggingModal(e.detail)}
		/>

		<!-- Empty State -->
		<EmptyState
			{videos}
			{searchTerm}
			{statusFilter}
			{categoryFilter}
			{syncStatusFilter}
			{vidStatusFilter}
			on:openCreateModal={openCreateModal}
		/>
	</div>
{/if}

<!-- Smart Tagging Modal -->
<SmartTaggingModal
	video={selectedVideoForTagging}
	isOpen={showTaggingModal}
	on:close={closeTaggingModal}
	on:videoUpdated={handleVideoUpdated}
/>

<!-- Tag Analytics Modal -->
<TagAnalyticsModal
	isOpen={showTagAnalytics}
	on:close={closeTagAnalyticsModal}
/>

<!-- Video Modals -->
<VideoModals
	bind:showUploadModal
	bind:showEditModal
	bind:showCreateModal
	{selectedVideo}
	{isUploading}
	{uploadProgress}
	{uploadForm}
	{editForm}
	{createForm}
	{categories}
	on:uploadVideo={uploadVideo}
	on:saveEdit={saveEdit}
	on:createVideo={createVideo}
	on:closeUploadModal={() => showUploadModal = false}
	on:closeEditModal={() => showEditModal = false}
	on:closeCreateModal={closeCreateModal}
/>

<style>
	/* Page Layout */
	.streaming-page {
		min-height: 100vh;
		background: var(--bg-primary);
		color: var(--text-primary);
		padding: 0;
	}

	/* Loading and Error States */
	.loading-container {
		display: flex;
		justify-content: center;
		align-items: center;
		min-height: 400px;
	}

	.error-container {
		display: flex;
		justify-content: center;
		align-items: center;
		min-height: 400px;
	}

	.error-content {
		text-align: center;
		padding: var(--space-xl);
	}

	.error-content h2 {
		color: var(--error);
		margin-bottom: var(--space-md);
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.streaming-page {
			padding: 0;
		}
	}
</style> 
