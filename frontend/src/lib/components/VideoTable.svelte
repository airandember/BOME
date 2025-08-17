<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { fly } from 'svelte/transition';
	import type { MasterVideo } from '$lib/master-video';

	// Props
	export let videos: MasterVideo[] = [];
	export let currentPage: number = 1;
	export let totalPages: number = 1;
	export let totalVideos: number = 0;
	export let sortField: string | null = null;
	export let sortDirection: 'asc' | 'desc' = 'asc';
	export let togglingVideos: Set<number> = new Set();

	// Events
	const dispatch = createEventDispatcher<{
		editVideo: MasterVideo;
		toggleVideoStatus: MasterVideo;
		changePage: number;
		handleSort: string;
		tagVideo: MasterVideo;
	}>();

	// Available categories
	const categories = [
		'Archaeology', 'Geography', 'DNA Research', 'Linguistics',
		'Historical Evidence', 'Cultural Studies', 'Religious Studies',
		'Documentary', 'Lecture', 'Interview', 'Presentation', 'Virtual Tour'
	];

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

	// Get video status badge
	function getVideoStatusBadge(vidStatus: boolean) {
		if (vidStatus) {
			return { text: 'Active', class: 'Vid_Active' };
		} else {
			return { text: 'Inactive', class: 'Vid_Inactive' };
		}
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

	// Handle edit video
	function handleEditVideo(video: MasterVideo) {
		dispatch('editVideo', video);
	}

	// Handle toggle video status
	function handleToggleVideoStatus(video: MasterVideo) {
		dispatch('toggleVideoStatus', video);
	}

	// Handle page change
	function handlePageChange(page: number) {
		dispatch('changePage', page);
	}

	// Handle sort
	function handleSort(field: string) {
		dispatch('handleSort', field);
	}

	// Handle tag video
	function handleTagVideo(video: MasterVideo) {
		dispatch('tagVideo', video);
	}
</script>

<div class="videos-table-container">
	<div class="overflow-x-auto">
		<table class="videos-table">
			<thead>
				<tr>
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
						</td>
						<td>{video.Category || 'Uncategorized'}</td>
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
								on:click={() => handleEditVideo(video)}
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
								on:click={() => handleToggleVideoStatus(video)}
								title="{video.Vid_Status ? 'Deactivate video' : 'Activate video'}"
							>
								{#if togglingVideos.has(video.ID)}
									<svg class="animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-dasharray="31.416" stroke-dashoffset="31.416">
											<animate attributeName="stroke-dasharray" dur="2s" values="0 31.416;15.708 15.708;0 31.416" repeatCount="indefinite"/>
											<animate attributeName="stroke-dashoffset" dur="2s" values="0;-15.708;-31.416" repeatCount="indefinite"/>
										</circle>
									</svg>
								{:else}
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										{#if video.Vid_Status}
											<path d="M12 2v10M18.4 6.6a9 9 0 1 1-12.77.04"/>
										{:else}
											<path d="M12 2v10M18.4 6.6a9 9 0 1 1-12.77.04"/>
											<line x1="2" y1="2" x2="22" y2="22"/>
										{/if}
									</svg>
								{/if}
							</button>

							<button 
								aria-label="Tag video"
								class="btn btn-sm btn-tag"
								on:click={() => handleTagVideo(video)}
								title="Tag video"
							>
								��️
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
				on:click={() => handlePageChange(currentPage - 1)}
			>
				Previous
			</button>
			<span class="page-info">
				Page {currentPage} of {totalPages} ({totalVideos} total videos)
			</span>
			<button 
				class="btn btn-secondary" 
				disabled={currentPage === totalPages}
				on:click={() => handlePageChange(currentPage + 1)}
			>
				Next
			</button>
		</div>
	{/if}
</div>

<style>
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

	.btn-sm {
		padding: 0.5rem;
		font-size: 0.8rem;
	}

	.btn-sm svg {
		width: 16px;
		height: 16px;
		color: inherit;
	}

	.btn-tag {
		background: linear-gradient(135deg, #f59e0b, #d97706);
		color: white;
		border: none;
	}

	.btn-tag:hover {
		background: linear-gradient(135deg, #d97706, #b45309);
		transform: translateY(-1px);
	}

	.btn.loading {
		opacity: 0.7;
		cursor: not-allowed;
		pointer-events: none;
	}

	.animate-spin {
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
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

	@media (max-width: 768px) {
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
	}
</style>
