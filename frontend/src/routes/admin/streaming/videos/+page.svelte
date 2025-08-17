<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { masterVideoService, type MasterVideo } from '$lib/master-video';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

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

	// Computed properties for video stats
	$: syncedCount = analytics?.video_stats?.synced_videos || 0;
	$: needsAttentionCount = analytics?.video_stats?.needs_attention || 0;
	$: conflictCount = analytics?.video_stats?.videos_by_sync_status?.conflict || 0;

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

	// Handle file selection
	function handleFileSelect(event: Event) {
		const target = event.target as HTMLInputElement;
		if (target.files && target.files[0]) {
			uploadForm.videoFile = target.files[0];
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

	// Delete video
	// async function deleteVideo(video: MasterVideo) {
	// 	if (!confirm(`Are you sure you want to delete "${video.Title}"?`)) {
	// 		return;
	// 	}

	// 	try {
	// 		const response = await masterVideoService.deleteMasterVideo(video.ID);
			
	// 		if (response.success) {
	// 			showToast('Video deleted successfully', 'success');
	// 			await loadVideos();
	// 		} else {
	// 			throw new Error(response.message || 'Delete failed');
	// 		}
	// 	} catch (err: unknown) {
	// 		console.error('Error deleting video:', err);
	// 		showToast(err instanceof Error ? err.message : 'Delete failed', 'error');
	// 	}
	// }

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

	// Format file size
	function formatFileSize(bytes: number): string {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	}

	// Format duration
	function formatDuration(seconds: number): string {
		const hours = Math.floor(seconds / 3600);
		const minutes = Math.floor((seconds % 3600) / 60);
		const secs = seconds % 60;
		
		if (hours > 0) {
			return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
		}
		return `${minutes}:${secs.toString().padStart(2, '0')}`;
	}

	// Get status badge
	function getStatusBadge(status: string) {
		switch (status) {
			case 'active':
				return { text: 'Active', class: 'success' };
			case 'processing':
				return { text: 'Processing', class: 'warning' };
			case 'draft':
				return { text: 'Draft', class: 'info' };
			case 'archived':
				return { text: 'Archived', class: 'error' };
			default:
				return { text: status, class: 'info' };
		}
	}

	// Get sync status badge
	function getSyncStatusBadge(syncStatus: string) {
		switch (syncStatus) {
			case 'synced':
				return { text: 'Synced', class: 'success' };
			case 'needs_attention':
				return { text: 'Needs Attention', class: 'warning' };
			case 'conflict':
				return { text: 'Conflict', class: 'error' };
			default:
				return { text: syncStatus, class: 'info' };
		}
	}

	// Get status badge
	function getVideoStatusBadge(vidStatus: boolean) {
		if (vidStatus) {
			return { text: 'Active', class: 'Vid_Active' };
		} else {
			return { text: 'Inactive', class: 'Vid_Inactive' };
		}
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

	// Get sort icon
	function getSortIcon(field: string) {
		if (sortField !== field) return '';
		return sortDirection === 'asc' ? '↑' : '↓';
	}

	// Format date
	function formatDate(dateString: string) {
		return new Date(dateString).toLocaleDateString();
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

	// Handle file selection for create modal
	function handleCreateFileSelect(event: Event) {
		const target = event.target as HTMLInputElement;
		if (target.files && target.files[0]) {
			createForm.videoFile = target.files[0];
		}
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
		<!-- Header -->
		<!--<div class="page-header">
			<div class="header-content">
				<div class="header-left">
					<h1>Master Video Management</h1>
					<p>Manage video content and Bunny.net synchronization</p>
				</div>
				<div class="header-actions">
					<a
						href="/admin/streaming/upload"
						class="btn bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors flex items-center space-x-2"

					>
						<svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
							<polyline points="7,10 12,15 17,10"></polyline>
							<line x1="12" y1="15" x2="12" y2="3"></line>
						</svg>
						<span class='color: white !important'>Batch Upload</span>
					</a>
					<button
						class="btn bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors flex items-center space-x-2"
						on:click={openCreateModal}
					>
						<svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
							<polyline points="7,10 12,15 17,10"></polyline>
							<line x1="12" y1="15" x2="12" y2="3"></line>
						</svg>
						<span>Create Video</span>
					</button>
				</div>
			</div>
		</div>-->

		<!-- Stats Summary -->
		<div class="stats-grid">
			<div class="stat-card primary">
				<div class="stat-icon">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<polygon points="23,7 16,12 23,17 23,7"></polygon>
						<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
					</svg>
				</div>
				<div class="stat-content">
					<p class="stat-label">Total Videos</p>
					<p class="stat-value">{AbsoluteTotalVideos}</p>
					<!--<p class="stat-change">View all</p>-->
				</div>
			</div>
			<div class="stat-card success">
				<div class="stat-icon">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M9 12l2 2 4-4"/>
						<path d="M21 12c0 4.97-4.03 9-9 9s-9-4.03-9-9 4.03-9 9-9 9 4.03 9 9z"/>
					</svg>
				</div>
				<div class="stat-content">
					<p class="stat-label">Synced</p>
					<p class="stat-value">{syncedCount}</p>
					<!--<p class="stat-change positive">+2.5%</p>-->
				</div>
			</div>
			<div class="stat-card warning">
				<div class="stat-icon">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
						<line x1="12" y1="9" x2="12" y2="13"/>
						<line x1="12" y1="17" x2="12.01" y2="17"/>
					</svg>
				</div>
				<div class="stat-content">
					<p class="stat-label">Found Conflict</p>
					<p class="stat-value">{conflictCount}</p>
					<!--<p class="stat-change positive">+0.8%</p>-->
				</div>
			</div>
			<div class="vid_buttons" style="gap: 1rem; display: flex; flex-direction: column;">
				<div class="rabbitButtons" style="display: flex; flex-direction: row; gap: 1rem;">
					<button 
						class="btn btn-secondary" 
						style="background: linear-gradient(135deg, #f59e0b, #d97706); "
						on:click={checkConflicts} 
						disabled={checkingConflicts}
					>
						{#if checkingConflicts}
							<LoadingSpinner size="small" />
							Checking...
						{:else}
							
							<span style="color: white !important; font-size: clamp(12px, 1vw, 2.5rem);">Run Check<br> 🐇</span>
						{/if}
					</button>
					
					<button class="bt btn-secondary"
						style="background: linear-gradient(135deg, #f59e0b, #d97706); ;"
						
					>
						<span style="color: white !important; font-size: clamp(12px, 1vw, 2.5rem);">🏷️Tags </span>
					</button>
				</div>
				<a
					href="/admin/streaming/upload"
					class="btn bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors flex items-center space-x-2"

				>
					<svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
						<polyline points="7,10 12,15 17,10"></polyline>
						<line x1="12" y1="15" x2="12" y2="3"></line>
					</svg>
					<span style="color: white !important; font-size: clamp(12px, 1.5vw, 4rem">Upload</span>
				</a>
			</div>
		</div>

		<!-- Filters -->
		<div class="filters-section">
			<div class="filters-grid">
				<div class="form-group">
					<label for="search">Search</label>
					<div class="dashSearch">
						<input
							id="search"
							class="dashSearch_input"
							type="text"
							bind:value={searchTerm}
							placeholder="Search videos..."
							on:keydown={(e) => e.key === 'Enter' && handleSearchInput()}
						/>
						<button
							class="dashSearch_glass"
							on:click={handleSearchInput}
						>
							🔍
						</button>
					</div>
					
				</div>
				<div class="form-group">
					<label for="status-filter">Status</label>
					<select
						id="status-filter"
						bind:value={statusFilter}
						on:change={handleFilterChange}
					>
						<option value="all">All Status</option>
						<option value="active">Active</option>
						<option value="processing">Processing</option>
						<option value="draft">Draft</option>
						<option value="archived">Archived</option>
					</select>
				</div>
				<!--<div class="form-group">
					<label for="status-filter">Status</label>
					<select
						id="status-filter"
						bind:value={statusFilter}
						on:change={handleFilterChange}
					>
						<option value="all">All Status</option>
						<option value="active">Active</option>
						<option value="processing">Processing</option>
						<option value="draft">Draft</option>
						<option value="archived">Archived</option>
					</select>
				</div>-->
				<div class="form-group">
					<label for="category-filter">Category</label>
					<select
						id="category-filter"
						bind:value={categoryFilter}
						on:change={handleFilterChange}
					>
						<option value="all">All Categories</option>
						{#each categories as category}
							<option value={category}>{category}</option>
						{/each}
					</select>
				</div>
				<div class="form-group">
					<label for="sync-status-filter">Sync Status</label>
					<select
						id="sync-status-filter"
						bind:value={syncStatusFilter}
						on:change={handleFilterChange}
					>
						<option value="all">All Sync Status</option>
						<option value="synced">Synced</option>
						<option value="needs_attention">Needs Attention</option>
						<option value="conflict">Conflict</option>
					</select>
				</div>
				<div class="form-group">
					<label for="vid-status-filter">Video Status</label>
					<select
						id="vid-status-filter"
						bind:value={vidStatusFilter}
						on:change={handleFilterChange}
					>
						<option value="all">All Status</option>
						<option value="true">Active</option>
						<option value="false">Inactive</option>
					</select>
				</div>
				<div class="form-group" style="justify-content:center">
					<label for="refresh-button">Count: {totalVideos}</label>
					<button
						id="refresh-button"
						class="btn w-full bg-gray-100 text-gray-700 px-4 py-2 rounded-md hover:bg-gray-200 transition-colors"
						style="height: 53.24px"
						on:click={loadVideos}
					>
						Refresh
					</button>
				</div>
			</div>
		</div>

		<!-- Videos Table (Replaced Videos Grid) -->
		<div class="videos-table-container">
			<div class="overflow-x-auto">
				<table class="videos-table">
					<thead>
						<tr>
							<!--<th class="checkbox-header">
								Select all checkbox placeholder
							</th>-->
							<th class="sortable" on:click={() => handleSort('id')}>
								ID {getSortIcon('id')}
							</th>
							<th>Thumbnail</th>
							<th class="sortable" on:click={() => handleSort('title')}>
								Title {getSortIcon('title')}
							</th>
							<th class="sortable" on:click={() => handleSort('category')}>
								Category {getSortIcon('category')}
							</th>
								<!--<th class="sortable" on:click={() => handleSort('status')}>
								Status {getSortIcon('status')}
							</th>-->
							<th class="sortable" on:click={() => handleSort('sync_status')}>
								Synced {getSortIcon('sync_status')}
							</th>
							<th class="sortable" on:click={() => handleSort('views')}>
								Views {getSortIcon('views')}
							</th>
							<th class="sortable" on:click={() => handleSort('duration')}>
								Duration {getSortIcon('duration')}
							</th>
							<th class="sortable" on:click={() => handleSort('created_at')}>
								Created {getSortIcon('created_at')}
							</th>
							<th class="sortable" on:click={() => handleSort('vid_status')}>
								Status {getSortIcon('vid_status')}
							</th>
							<th>Actions</th>
						</tr>
					</thead>
					<tbody>
						{#each videos as video (video.ID)}
							<tr class="hover:bg-gray-50" in:fly={{ y: 20, duration: 300 }}>
								<!--<td class="checkbox-cell">
									Checkbox placeholder
								</td>-->
								<td>{video.ID}</td>
								<td class="thumbnail-cell">
									{#if video.ThumbnailURL}
										<img 
											src={video.ThumbnailURL} 
											alt={video.Title} 
											class="video-thumbnail" 
											loading="lazy"
											referrerpolicy="no-referrer"
											on:error={(e) => {
												console.warn('Thumbnail failed to load:', video.ThumbnailURL);
											}}
										/>
									{:else}
										<div class="no-thumbnail">No Image</div>
									{/if}
								</td>
								<td class="title-cell">
									<div class="video-title">{video.Title}</div>
									<!--<div class="video-description">{video.Description}</div>-->
								</td>
								<td>{video.Category || 'Uncategorized'}</td>
								<!--<td class="td_cell">
									<span class="status-badge {getStatusBadge(video.Status).class}">
										{getStatusBadge(video.Status).text}
									</span>
								</td>-->
								<td class="td_cell">
									<span class="status-badge {getSyncStatusBadge(video.SyncStatus).class}">
										{getSyncStatusBadge(video.SyncStatus).text}
									</span>
								</td>
								<td>{video.Views?.toLocaleString() || 0}</td>
								<td>{formatDuration(video.Duration)}</td>
								<td>{formatDate(video.CreatedAt)}</td>
								<td>
									<span class="status-badge {getVideoStatusBadge(video.Vid_Status).class}">
										{getVideoStatusBadge(video.Vid_Status).text}
									</span>
								</td>
								<td class="actions-cell">
									<button aria-label="Edit video"
										class="btn btn-sm btn-secondary"
										on:click={() => editVideo(video)}
										title="Edit video"
									>
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
											<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
										</svg>
									</button>
									
										<button 
										aria-label="Toggle video status"
										class="btn btn-sm {video.Vid_Status ? 'Vid_Inactive' : 'Vid_Active'}"
										class:loading={togglingVideos.has(video.ID)}
										disabled={togglingVideos.has(video.ID)}
										on:click={() => toggleVideoStatus(video)}
										title="{video.Vid_Status ? 'Deactivate video' : 'Activate video'}"
									>
										{#if togglingVideos.has(video.ID)}
											<!-- Loading spinner -->
											<svg class="animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
												<circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-dasharray="31.416" stroke-dashoffset="31.416">
													<animate attributeName="stroke-dasharray" dur="2s" values="0 31.416;15.708 15.708;0 31.416" repeatCount="indefinite"/>
													<animate attributeName="stroke-dashoffset" dur="2s" values="0;-15.708;-31.416" repeatCount="indefinite"/>
												</circle>
											</svg>
										{:else}
											<!-- Original icon -->
											<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
												{#if video.Vid_Status}
													<!-- Power on icon -->
													<path d="M12 2v10M18.4 6.6a9 9 0 1 1-12.77.04"/>
												{:else}
													<!-- Power off icon -->
													<path d="M12 2v10M18.4 6.6a9 9 0 1 1-12.77.04"/>
													<line x1="2" y1="2" x2="22" y2="22"/>
												{/if}
											</svg>
										{/if}
									</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>

			<!-- Pagination -->
			{#if totalPages > 1}
				<div class="pagination">
					<button 
						class="btn btn-secondary" 
						disabled={currentPage === 1}
						on:click={() => changePage(currentPage - 1)}
					>
						Previous
					</button>
					<span class="page-info">
						Page {currentPage} of {totalPages} ({totalVideos} total videos)
					</span>
					<button 
						class="btn btn-secondary" 
						disabled={currentPage === totalPages}
						on:click={() => changePage(currentPage + 1)}
					>
						Next
					</button>
				</div>
			{/if}
		</div>

		{#if videos.length === 0}
			<div class="empty-state">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<polygon points="23,7 16,12 23,17 23,7"></polygon>
					<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
				</svg>
				<h3>No videos found</h3>
				<p>
					{#if searchTerm || statusFilter !== 'all' || categoryFilter !== 'all' || syncStatusFilter !== 'all' || vidStatusFilter !== 'all'}
						No videos match your current filters. Try adjusting your search criteria.
					{:else}
						Get started by creating your first video or using batch upload for multiple videos.
					{/if}
				</p>
				{#if !searchTerm && statusFilter === 'all' && categoryFilter === 'all' && syncStatusFilter === 'all' && vidStatusFilter === 'all'}
					<div class="flex justify-center space-x-3">
						<button
							class="btn btn-primary"
							on:click={openCreateModal}
						>
							Create Video
						</button>
						<a
							href="/admin/streaming/upload"
							class="btn btn-secondary"
						>
							Batch Upload
						</a>
					</div>
				{/if}
			</div>
		{/if}
	</div>
{/if}

<!-- Upload Modal -->
{#if showUploadModal}
	<div class="modal-overlay" in:fade={{ duration: 200 }}>
		<div class="modal-content" in:fly={{ y: 20, duration: 200 }}>
			<div class="modal-header">
				<h2>Upload Video</h2>
				<button aria-label="Close"
					class="modal-close"
					on:click={() => showUploadModal = false}
				>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"/>
						<line x1="6" y1="6" x2="18" y2="18"/>
					</svg>
				</button>
			</div>

			<form on:submit|preventDefault={uploadVideo} class="form-grid">
				<div class="form-section">
					<h3 class="section-title">Video Details</h3>
					<div class="form-row">
						<div class="form-group">
							<label for="video-file">Video File</label>
							<input
								id="video-file"
								type="file"
								accept="video/*"
								on:change={handleFileSelect}
								class="form-control"
								required
							/>
							<p class="text-sm text-gray-500 mt-1">Supported formats: MP4, AVI, MOV, WMV, FLV, WebM, MKV (max 500MB)</p>
						</div>
						<div class="form-group">
							<label for="title">Title</label>
							<input
								id="title"
								type="text"
								bind:value={uploadForm.title}
								class="form-control"
								required
							/>
						</div>
					</div>
					<div class="form-row">
						<div class="form-group">
							<label for="description">Description</label>
							<textarea
								id="description"
								bind:value={uploadForm.description}
								rows="3"
								class="form-control"
							></textarea>
						</div>
						<div class="form-group">
							<label for="category">Category</label>
							<select
								id="category"
								bind:value={uploadForm.category}
								class="form-control"
							>
								<option value="">Select category</option>
								{#each categories as category}
									<option value={category}>{category}</option>
								{/each}
							</select>
						</div>
					</div>
					<div class="form-row">
						<div class="form-group">
							<label for="tags">Tags</label>
							<input
								id="tags"
								type="text"
								bind:value={uploadForm.tags}
								placeholder="Enter tags separated by commas"
								class="form-control"
							/>
						</div>
					</div>
					{#if isUploading}
						<div class="form-group">
							<label class="block text-sm font-medium text-gray-700 mb-2">Upload Progress</label>
							<div class="w-full bg-gray-200 rounded-full h-2">
								<div class="bg-blue-600 h-2 rounded-full transition-all duration-300" style="width: {uploadProgress}%"></div>
							</div>
							<p class="text-sm text-gray-600 mt-1">{uploadProgress}% complete</p>
						</div>
					{/if}
					<div class="form-actions">
						<button
							type="button"
							class="btn btn-secondary"
							on:click={() => showUploadModal = false}
							disabled={isUploading}
						>
							Cancel
						</button>
						<button
							type="submit"
							disabled={isUploading}
							class="btn btn-primary"
						>
							{isUploading ? 'Uploading...' : 'Upload Video'}
						</button>
					</div>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Edit Modal -->
{#if showEditModal}
	<div class="modal-overlay" in:fade={{ duration: 200 }}>
		<div class="modal-content" in:fly={{ y: 20, duration: 200 }}>
			<div class="modal-header">
				<h2>Edit Video</h2>
				<button aria-label="Close"
					class="modal-close"
					on:click={() => showEditModal = false}
				>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"/>
						<line x1="6" y1="6" x2="18" y2="18"/>
					</svg>
				</button>
			</div>

			<form on:submit|preventDefault={saveEdit} class="form-grid">
				<div class="form-section">
					<h3 class="section-title">Video Details</h3>
					<div class="form-row">
						<div class="form-group">
							<label for="edit-title">Title</label>
							<input
								id="edit-title"
								type="text"
								bind:value={editForm.title}
								class="form-control"
								required
							/>
						</div>
					</div>
					<div class="form-row">
						<div class="form-group">
							<label for="edit-description">Description</label>
							<textarea
								id="edit-description"
								bind:value={editForm.description}
								rows="3"
								class="form-control"
							></textarea>
						</div>
						<div class="form-group">
							<label for="edit-category">Category</label>
							<select
								id="edit-category"
								bind:value={editForm.category}
								class="form-control"
							>
								<option value="">Select category</option>
								{#each categories as category}
									<option value={category}>{category}</option>
								{/each}
							</select>
						</div>
					</div>
					<div class="form-row">
						<div class="form-group">
							<label for="edit-tags">Tags</label>
							<input
								id="edit-tags"
								type="text"
								bind:value={editForm.tags}
								placeholder="Enter tags separated by commas"
								class="form-control"
							/>
						</div>
						<div class="form-group">
							<label for="edit-status">Status</label>
							<select
								id="edit-status"
								bind:value={editForm.status}
								class="form-control"
							>
								<option value="active">Active</option>
								<option value="draft">Draft</option>
								<option value="archived">Archived</option>
							</select>
						</div>
					</div>
					<div class="form-actions">
						<button
							type="button"
							class="btn btn-secondary"
							on:click={() => showEditModal = false}
						>
							Cancel
						</button>
						<button
							type="submit"
							class="btn btn-primary"
						>
							Save Changes
						</button>
					</div>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Create Modal -->
{#if showCreateModal}
	<div class="modal-overlay" in:fade={{ duration: 200 }}>
		<div class="modal-content" in:fly={{ y: 20, duration: 200 }}>
			<div class="modal-header">
				<h2>Create New Video</h2>
				<button aria-label="Close"
					class="modal-close"
					on:click={closeCreateModal}
				>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"/>
						<line x1="6" y1="6" x2="18" y2="18"/>
					</svg>
				</button>
			</div>

			<form on:submit|preventDefault={createVideo} class="form-grid">
				<div class="form-section">
					<h3 class="section-title">Video Details</h3>
					<div class="form-row">
						<div class="form-group">
							<label for="create-video-file">Video File</label>
							<input
								id="create-video-file"
								type="file"
								accept="video/*"
								on:change={handleCreateFileSelect}
								class="form-control"
								required
							/>
							<p class="text-sm text-gray-500 mt-1">Supported formats: MP4, AVI, MOV, WMV, FLV, WebM, MKV (max 500MB)</p>
						</div>
						<div class="form-group">
							<label for="create-title">Title</label>
							<input
								id="create-title"
								type="text"
								bind:value={createForm.title}
								class="form-control"
								required
							/>
						</div>
					</div>
					<div class="form-row">
						<div class="form-group">
							<label for="create-description">Description</label>
							<textarea
								id="create-description"
								bind:value={createForm.description}
								rows="3"
								class="form-control"
							></textarea>
						</div>
						<div class="form-group">
							<label for="create-category">Category</label>
							<select
								id="create-category"
								bind:value={createForm.category}
								class="form-control"
							>
								<option value="">Select category</option>
								{#each categories as category}
									<option value={category}>{category}</option>
								{/each}
							</select>
						</div>
					</div>
					<div class="form-row">
						<div class="form-group">
							<label for="create-tags">Tags</label>
							<input
								id="create-tags"
								type="text"
								bind:value={createForm.tags}
								placeholder="Enter tags separated by commas"
								class="form-control"
							/>
						</div>
					</div>
					{#if isUploading}
						<div class="form-group">
							<label class="block text-sm font-medium text-gray-700 mb-2">Upload Progress</label>
							<div class="w-full bg-gray-200 rounded-full h-2">
								<div class="bg-blue-600 h-2 rounded-full transition-all duration-300" style="width: {uploadProgress}%"></div>
							</div>
							<p class="text-sm text-gray-600 mt-1">{uploadProgress}% complete</p>
						</div>
					{/if}
					<div class="form-actions">
						<button
							type="button"
							class="btn btn-secondary"
							on:click={closeCreateModal}
							disabled={isUploading}
						>
							Cancel
						</button>
						<button
							type="submit"
							disabled={isUploading}
							class="btn btn-primary"
						>
							{isUploading ? 'Creating...' : 'Create Video'}
						</button>
					</div>
				</div>
			</form>
		</div>
	</div>
{/if}

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

	/* Header Styling */
	.page-header {
		background: var(--bg-glass, rgba(255, 255, 255, 0.1));
		backdrop-filter: blur(20px);
		border-radius: 20px;
		padding: 2rem;
		margin-bottom: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.header-content {
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: 1rem;
	}

	.header-left h1 {
		font-size: 2rem;
		font-weight: 700;
		margin: 0 0 0.5rem 0;
		color: var(--text-primary);
	}

	.header-left p {
		color: var(--text-secondary);
		margin: 0;
	}

	.header-actions {
		display: flex;
		gap: 1rem;
		flex-wrap: wrap;
	}

	/* Stats Grid */
	.stats-grid {
		display: flex;
		width: 100%;
		justify-content: space-between;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 1.5rem;
		margin-bottom: 2rem;
	}

	.stat-card {
		background: var(--bg-glass, rgba(255, 255, 255, 0.1));
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: 0.75rem 1.5rem 0 1.5rem ;
		border: 1px solid rgba(255, 255, 255, 0.1);
		display: flex;
		align-items: center;
		gap: 1.5rem;
		transition: all 0.3s ease;

	}

	.stat-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
	}

	.stat-card.primary {
		border-left: 4px solid var(--primary-color, #3b82f6);
	}

	.stat-card.success {
		border-left: 4px solid #10b981;
	}

	.stat-card.warning {
		border-left: 4px solid #f59e0b;
	}

	.stat-card.info {
		border-left: 4px solid #3b82f6;
	}

	.stat-icon {
		width: 60px;
		height: 60px;
		border-radius: 15px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		background: var(--primary-gradient, linear-gradient(135deg, #3b82f6, #1d4ed8));
	}

	.stat-card.success .stat-icon {
		background: linear-gradient(135deg, #10b981, #059669);
	}

	.stat-card.warning .stat-icon {
		background: linear-gradient(135deg, #f59e0b, #d97706);
	}

	.stat-card.info .stat-icon {
		background: linear-gradient(135deg, #3b82f6, #1d4ed8);
	}

	.stat-icon svg {
		width: 28px;
		height: 28px;
		color: white;
	}

	.stat-content {
		flex: 1;
	}

	.rabbitButtons {
		display: flex;
		flex-direction: row;
	}

	.stat-value {
		font-size: 2rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 0.25rem;
	}

	.stat-label {
		font-size: 0.9rem;
		color: var(--text-secondary);
		margin-bottom: 0.5rem;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-weight: 600;
	}

	.stat-change {
		font-size: 0.8rem;
		font-weight: 600;
		color: #10b981;
	}

	/* Filters Section */
	.filters-section {
		background: var(--bg-glass, rgba(255, 255, 255, 0.1));
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: 1.5rem;
		margin-bottom: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.filters-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
	}

	.filter-group {
		display: flex;
		flex-direction: column;
	}

	.filter-group label {
		font-size: 0.9rem;
		color: var(--text-secondary);
		margin-bottom: 0.5rem;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.filter-group input,
	.filter-group select {
		padding: 0.75rem 1rem;
		border: 1px solid rgba(255, 255, 255, 0.2);
		border-radius: 8px;
		background: var(--bg-primary, rgba(255, 255, 255, 0.05));
		color: var(--text-primary, #ffffff);
		font-size: 0.9rem;
		transition: all 0.3s ease;
	}

	.filter-group input:focus,
	.filter-group select:focus {
		outline: none;
		border-color: var(--primary-color, #3b82f6);
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	/* Video Table Styles */
	.tb_center {
		text-align: center;
	}

	.videos-table-container {
		background: var(--bg-glass, rgba(255, 255, 255, 0.1));
		backdrop-filter: blur(20px);
		border-radius: 15px;
		border: 1px solid rgba(255, 255, 255, 0.1);
		overflow: hidden;
	}

	.videos-table {
		width: 100%;
		border-collapse: collapse;
	}

	.videos-table th,
	.videos-table td {
		padding: 1rem;
		text-align: left;
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
	}

	.videos-table th {
		background: rgba(255, 255, 255, 0.05);
		font-weight: 600;
		color: var(--text-primary);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-size: 0.8rem;
		text-align: center;
	}

	.videos-table th.sortable {
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.videos-table th.sortable:hover {
		background: rgba(255, 255, 255, 0.1);
	}

	.videos-table tr:hover {
		background: rgba(255, 255, 255, 0.05);
	}

	.thumbnail-cell {
		width: 80px;
	}

	.Vid_Active {
		color: #ffffe6;
		background: #10b981;
	}

	.Vid_Inactive {
		color: white;
		background: #FF4800;
	}

	.td_cell {
		font-size: clamp(0.3rem, 0.6vw, 0.6rem);
		padding: 0.5rem !important;
	}

	.video-thumbnail {
		width: 260px;
		height: auto;
		object-fit: cover;
		border-radius: 6px;
	}

	.no-thumbnail {
		width: 60px;
		height: 40px;
		background: var(--bg-secondary, rgba(255, 255, 255, 0.05));
		border-radius: 6px;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.7rem;
		color: var(--text-secondary);
	}

	.title-cell {
		max-width: 300px;
	}

	.video-title {
		font-weight: 600;
		margin-bottom: 0.25rem;
		color: var(--text-primary);
	}

	.video-description {
		font-size: 0.8rem;
		color: var(--text-secondary);
		line-height: 1.4;
	}

	.status-badge {
		display: flex;
		flex-wrap: nowrap;
		padding: 0.25rem 0.75rem;
		border-radius: 20px;
		font-size: clamp(0.3rem, 0.6vw, 0.6rem);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.status-badge.success {
		background: rgba(16, 185, 129, 0.2);
		color: #10b981;
	}

	.status-badge.warning {
		background: rgba(245, 158, 11, 0.2);
		color: #f59e0b;
	}

	.status-badge.error {
		background: rgba(239, 68, 68, 0.2);
		color: #ef4444;
	}

	.status-badge.info {
		background: rgba(59, 130, 246, 0.2);
		color: #3b82f6;
	}

	.actions-cell {
		display: flex;
		flex-direction: column;
		height: 185.25px;
		gap: 3rem;
		margin-top: 1rem;
		vertical-align: top;
	}

	/* Button Styles */
	.btn {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 8px;
		font-weight: 600;
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.3s ease;
		text-decoration: none;
	}

	.btn-primary {
		background: var(--primary-gradient, linear-gradient(135deg, #3b82f6, #1d4ed8));
		color: white;
	}

	.btn-primary:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 20px rgba(99, 102, 241, 0.3);
	}

	.btn-secondary {
		background: var(--bg-glass, rgba(255, 255, 255, 0.1));
		color: var(--text-primary);
		border: 1px solid rgba(255, 255, 255, 0.2);
		backdrop-filter: blur(20px);
	}

	.btn-secondary:hover {
		background: rgba(255, 255, 255, 0.2);
		transform: translateY(-2px);
	}

	.btn-danger {
		background: var(--danger-gradient, linear-gradient(135deg, #ef4444, #dc2626));
		color: white;
	}

	.btn-danger:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 20px rgba(239, 68, 68, 0.3);
	}

	.btn-sm {
		padding: 0.5rem;
		font-size: 0.8rem;
	}

	.btn-sm svg {
		width: 16px;
		height: 16px;
		color: inherit;
	}

	/* Header button styling */
	.bg-blue-600 {
		background: var(--primary-gradient, linear-gradient(135deg, #3b82f6, #1d4ed8)) !important;
	}

	.bg-green-600 {
		background: var(--success-gradient, linear-gradient(135deg, #10b981, #047857)) !important;
	}

	.bg-blue-600:hover,
	.bg-green-600:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 20px rgba(59, 130, 246, 0.3);
	}

	/* SVG styling for header buttons */
	.bg-blue-600 svg,
	.bg-green-600 svg {
		color: white;
		width: 20px;
		height: 20px;
	}

	/* Pagination */
	.pagination {
		display: flex;
		justify-content: center;
		align-items: center;
		gap: 1rem;
		padding: 2rem;
	}

	.page-info {
		font-size: 0.9rem;
		color: var(--text-secondary);
	}

	/* Empty State */
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 300px;
		gap: 1rem;
		text-align: center;
	}

	.empty-state svg {
		width: 64px;
		height: 64px;
		color: var(--text-secondary);
	}

	.empty-state h3 {
		font-size: 1.5rem;
		font-weight: 600;
		margin: 0;
		color: var(--text-primary);
	}

	.empty-state p {
		color: var(--text-secondary);
		max-width: 400px;
		line-height: 1.5;
	}

	/* Modal Styles */
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.8);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
		padding: 2rem;
	}

	.modal-content {
		background: var(--bg-primary, rgba(0, 0, 0, 0.9));
		border: 1px solid var(--border-color, rgba(255, 255, 255, 0.1));
		border-radius: 12px;
		padding: 2.5rem;
		width: 95%;
		max-width: 1800px;
		max-height: 90vh;
		overflow-y: auto;
		display: flex;
		flex-direction: column;
		gap: 2rem;
		box-shadow: 0 25px 50px rgba(0, 0, 0, 0.25);
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
	}

	.modal-header h2 {
		margin: 0;
		color: var(--text-primary, #ffffff);
	}

	.modal-close {
		background: none;
		border: none;
		color: var(--text-secondary);
		font-size: 2rem;
		cursor: pointer;
		padding: 0;
		width: 40px;
		height: 40px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 50%;
		transition: all 0.3s ease;
	}

	.modal-close:hover {
		background: rgba(255, 255, 255, 0.1);
		color: var(--text-primary);
	}

	/* Form Styles */
	.form-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
		gap: 2rem;
		margin-bottom: 2rem;
	}

	.form-section {
		background: var(--bg-secondary, rgba(255, 255, 255, 0.05));
		border: 1px solid var(--border-color, rgba(255, 255, 255, 0.1));
		border-radius: 10px;
		padding: 2rem;
		transition: all 0.2s ease;
	}

	.form-section:hover {
		border-color: var(--primary-color, #3b82f6);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}

	.section-title {
		margin-top: 0;
		margin-bottom: 1.5rem;
		color: var(--text-primary);
		font-size: 1.25rem;
		font-weight: 600;
		padding-bottom: 0.75rem;
		border-bottom: 2px solid var(--border-color, rgba(255, 255, 255, 0.1));
		position: relative;
	}

	.section-title::after {
		content: '';
		position: absolute;
		bottom: -2px;
		left: 0;
		width: 60px;
		height: 2px;
		background: var(--primary-color, #3b82f6);
		border-radius: 1px;
	}

	.form-row {
		display: flex;
		gap: 1.5rem;
		margin-bottom: 1.5rem;
	}

	.dashSearch {
		display: flex;
		flex-direction: row !important;
		flex-wrap: nowrap;
		align-items: center;
		justify-content: center;
		
	}

	.dashSearch_input {
		width: 85%;
		border-radius: 0.5rem 0 0 0.5rem !important;
	}

	.dashSearch_glass {
		width: 15%;
		border-radius: 0 0.5rem 0.5rem 0 !important;
		height: 97%;
		background: var(--bg-primary, rgba(255, 255, 255, 0.05));
		border: 2px solid var(--border-color, rgba(255, 255, 255, 0.2));

	}

	.dashSearch_glass:hover {
		background: var(--bg-tertiary, rgba(255, 255, 255, 0.05));
		border: 2px solid var(--border-color, rgba(255, 255, 255, 0.2));
	}

	.dashSearch_glass:active {
		background: var(--bg-secondary, rgba(255, 255, 255, 0.05));
		border: 2px solid var(--border-color, rgba(255, 255, 255, 0.2));
	}

	.form-group {
		flex: 1;
		display: flex;
		flex-direction: column;
	}

	.form-group label {
		font-size: 0.9rem;
		color: var(--text-secondary);
		margin-bottom: 0.75rem;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.form-group input[type="text"],
	.form-group input[type="number"],
	.form-group input[type="url"],
	.form-group input[type="file"],
	.form-group select,
	.form-group textarea {
		padding: 1rem 1.25rem;
		border: 2px solid var(--border-color, rgba(255, 255, 255, 0.2));
		border-radius: 8px;
		background: var(--bg-primary, rgba(255, 255, 255, 0.05));
		color: var(--text-primary, #ffffff);
		font-size: 1rem;
		transition: all 0.3s ease;
		width: 100%;
		box-sizing: border-box;
		font-weight: 500;
	}

	.form-group input[type="text"]:focus,
	.form-group input[type="number"]:focus,
	.form-group input[type="url"]:focus,
	.form-group input[type="file"]:focus,
	.form-group select:focus,
	.form-group textarea:focus {
		outline: none;
		border-color: var(--primary-color, #3b82f6);
		box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.15);
		background: var(--bg-primary, rgba(255, 255, 255, 0.05));
		transform: translateY(-1px);
	}

	.form-group textarea {
		resize: vertical;
		min-height: 80px;
		line-height: 1.5;
	}

	/* Loading button styles */
	.btn.loading {
		opacity: 0.7;
		cursor: not-allowed;
		pointer-events: none;
	}

	.btn.loading svg {
		width: 16px;
		height: 16px;
	}

	.animate-spin {
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		from {
			transform: rotate(0deg);
		}
		to {
			transform: rotate(360deg);
		}
	}

	/* Disable button when loading */
	.btn:disabled {
		opacity: 0.7;
		cursor: not-allowed;
	}

	/* Conflict check button styling */
	.btn-secondary {
		background: var(--bg-glass, rgba(255, 255, 255, 0.1));
		backdrop-filter: blur(20px);
		border: 1px solid rgba(255, 255, 255, 0.1);
		color: var(--text-primary);
		padding: 0.5rem 1rem;
		border-radius: 0.5rem;
		font-weight: 500;
		transition: all 0.3s ease;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		cursor: pointer;
	}

	.btn-secondary:hover:not(:disabled) {
		background: var(--bg-glass-dark, rgba(255, 255, 255, 0.15));
		transform: translateY(-1px);
	}

	.btn-secondary:disabled {
		opacity: 0.6;
		cursor: not-allowed;
		transform: none;
	}

	.btn-secondary svg {
		width: 1.25rem;
		height: 1.25rem;
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.streaming-page {
			padding: 0;
		}

		.header-content {
			flex-direction: column;
			align-items: stretch;
		}

		.header-actions {
			justify-content: center;
		}

		.stats-grid {
			grid-template-columns: 1fr;
		}

		.stat-card {
			flex-direction: column;
			text-align: center;
		}

		.filters-grid {
			grid-template-columns: 1fr;
		}

		.videos-table {
			font-size: 0.8rem;
		}

		.videos-table th,
		.videos-table td {
			padding: 0.5rem;
		}

		.thumbnail-cell {
			width: 60px;
		}

		.video-thumbnail,
		.no-thumbnail {
			width: 40px;
			height: 30px;
		}

		.title-cell {
			max-width: 150px;
		}

		.actions-cell {
			height: auto;
			gap: 1rem;
		}

		.form-grid {
			grid-template-columns: 1fr;
			gap: 1.5rem;
		}

		.form-section {
			padding: 1.5rem;
		}

		.form-row {
			flex-direction: column;
			gap: 1rem;
		}

		.modal-content {
			width: 98%;
			max-width: none;
			padding: 1.5rem;
			margin: 0.5rem;
		}
	}

	@media (max-width: 480px) {
		.stat-icon {
			width: 50px;
			height: 50px;
		}

		.stat-icon svg {
			width: 24px;
			height: 24px;
		}

		.stat-value {
			font-size: 1.5rem;
		}
	}
</style> 
