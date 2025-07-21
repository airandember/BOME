<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { apiRequest } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	interface BunnyVideo {
		id: string;
		title: string;
		description: string;
		status: string;
		duration: number;
		views: number;
		thumbnailUrl: string;
		videoUrl: string;
		iframeSrc: string;
		playbackUrl: string;
		createdAt: string;
		updatedAt: string;
		fileSize: number;
		resolution: string;
		category: string;
		tags: string[];
		encodeProgress: number;
		storageSize: number;
	}

	let loading = true;
	let saving = false;
	let video: BunnyVideo | null = null;
	
	// Form data
	let title = '';
	let description = '';
	let category = '';
	let selectedTags: string[] = [];
	let thumbnailTime = '';

	// Available tags
	const availableTags = [
		'Book of Mormon', 'Archaeology', 'Geography', 'DNA', 'Linguistics',
		'Historical', 'Cultural', 'Religious', 'Evidence', 'Research',
		'Ancient America', 'Mesoamerica', 'North America', 'South America',
		'Documentary', 'Lecture', 'Interview', 'Presentation', 'Tour'
	];

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

	onMount(async () => {
		await loadVideo();
	});

	async function loadVideo() {
		try {
			loading = true;
			const videoId = $page.params.id;
			const response = await apiRequest(`/bunny-videos/${videoId}`);
			
			if (response.ok) {
				video = await response.json();
				
				// Populate form data
				title = video?.title || '';
				description = video?.description || '';
				category = video?.category || '';
				selectedTags = video?.tags || [];
			} else {
				showToast('Failed to load video', 'error');
				goto('/admin/streaming');
			}
		} catch (error) {
			showToast('Failed to load video', 'error');
			console.error('Error loading video:', error);
			goto('/admin/streaming');
		} finally {
			loading = false;
		}
	}

	function toggleTag(tag: string) {
		if (selectedTags.includes(tag)) {
			selectedTags = selectedTags.filter(t => t !== tag);
		} else {
			selectedTags = [...selectedTags, tag];
		}
	}

	async function handleSubmit() {
		if (!video) return;

		if (!title.trim()) {
			showToast('Please enter a title', 'error');
			return;
		}

		try {
			saving = true;

			const updateData = {
				title: title.trim(),
				description: description.trim(),
				category: category,
				tags: selectedTags
			};

			const response = await apiRequest(`/bunny-videos/${video.id}`, {
				method: 'PUT',
				body: JSON.stringify(updateData)
			});

			if (response.ok) {
				showToast('Video updated successfully!', 'success');
				goto('/admin/streaming');
			} else {
				const error = await response.json();
				showToast(error.error || 'Failed to update video', 'error');
			}
		} catch (error) {
			showToast('Failed to update video', 'error');
			console.error('Error updating video:', error);
		} finally {
			saving = false;
		}
	}

	async function handleDelete() {
		if (!video) return;

		if (!confirm('Are you sure you want to delete this video? This action cannot be undone.')) {
			return;
		}

		try {
			const response = await apiRequest(`/bunny-videos/${video.id}`, {
				method: 'DELETE'
			});

			if (response.ok) {
				showToast('Video deleted successfully', 'success');
				goto('/admin/streaming');
			} else {
				const error = await response.json();
				showToast(error.error || 'Failed to delete video', 'error');
			}
		} catch (error) {
			showToast('Failed to delete video', 'error');
			console.error('Error deleting video:', error);
		}
	}

	function handleBack() {
		goto('/admin/streaming');
	}

	function formatDuration(seconds: number): string {
		const hours = Math.floor(seconds / 3600);
		const minutes = Math.floor((seconds % 3600) / 60);
		const secs = seconds % 60;
		
		if (hours > 0) {
			return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
		}
		return `${minutes}:${secs.toString().padStart(2, '0')}`;
	}

	function formatFileSize(bytes: number): string {
		const sizes = ['B', 'KB', 'MB', 'GB'];
		let size = bytes;
		let unitIndex = 0;
		
		while (size >= 1024 && unitIndex < sizes.length - 1) {
			size /= 1024;
			unitIndex++;
		}
		
		return `${size.toFixed(1)} ${sizes[unitIndex]}`;
	}

	function getStatusColor(status: string): string {
		switch (status) {
			case 'ready': return 'success';
			case 'processing': case 'transcoding': return 'warning';
			case 'error': case 'upload_failed': return 'error';
			default: return 'info';
		}
	}
</script>

<svelte:head>
	<title>Edit Video - Streaming Management</title>
	<meta name="description" content="Edit video metadata for Bunny.net streaming platform" />
</svelte:head>

<div class="edit-page">
	<header class="page-header">
		<button class="back-button" on:click={handleBack}>
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M19 12H5"></path>
				<path d="M12 19l-7-7 7-7"></path>
			</svg>
			Back to Streaming
		</button>
		<h1>Edit Video</h1>
		<p>Update video metadata and settings</p>
	</header>

	{#if loading}
		<div class="loading-container">
			<LoadingSpinner size="large" />
			<p>Loading video...</p>
		</div>
	{:else if video}
		<div class="edit-container">
			<!-- Video Preview -->
			<div class="video-preview">
				<div class="video-thumbnail">
					<img src={video.thumbnailUrl} alt={video.title} />
					<div class="video-duration">{formatDuration(video.duration)}</div>
					<div class="video-status {getStatusColor(video.status)}">
						{video.status}
					</div>
				</div>
				
				<div class="video-info">
					<h2>{video.title}</h2>
					<div class="video-stats">
						<div class="stat">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
								<circle cx="12" cy="12" r="3"></circle>
							</svg>
							{video.views.toLocaleString()} views
						</div>
						<div class="stat">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
								<polyline points="14,2 14,8 20,8"></polyline>
							</svg>
							{formatFileSize(video.storageSize)}
						</div>
						<div class="stat">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<rect x="3" y="3" width="7" height="7"></rect>
								<rect x="14" y="3" width="7" height="7"></rect>
								<rect x="14" y="14" width="7" height="7"></rect>
								<rect x="3" y="14" width="7" height="7"></rect>
							</svg>
							{video.category || 'Uncategorized'}
						</div>
					</div>
					<div class="video-actions">
						<button class="btn btn-secondary" on:click={() => window.open(video?.iframeSrc, '_blank')}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polygon points="23,7 16,12 23,17 23,7"></polygon>
								<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
							</svg>
							Preview
						</button>
						<button class="btn btn-danger" on:click={handleDelete}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polyline points="3,6 5,6 21,6"></polyline>
								<path d="m19,6v14a2,2 0 0,1 -2,2H7a2,2 0 0,1 -2,-2V6m3,0V4a2,2 0 0,1 2,-2h4a2,2 0 0,1 2,2v2"></path>
							</svg>
							Delete Video
						</button>
					</div>
				</div>
			</div>

			<!-- Edit Form -->
			<div class="edit-form">
				<h2>Edit Metadata</h2>
				
				<div class="form-group">
					<label for="title">Title *</label>
					<input 
						id="title"
						type="text" 
						bind:value={title} 
						placeholder="Enter video title"
						required
					/>
				</div>

				<div class="form-group">
					<label for="description">Description</label>
					<textarea 
						id="description"
						bind:value={description} 
						placeholder="Enter video description"
						rows="4"
					></textarea>
				</div>

				<div class="form-group">
					<label for="category">Category</label>
					<select id="category" bind:value={category}>
						<option value="">Select a category</option>
						{#each categories as cat}
							<option value={cat}>{cat}</option>
						{/each}
					</select>
				</div>

				<div class="form-group">
					<label>Tags</label>
					<div class="tags-container">
						{#each availableTags as tag}
							<button 
								class="tag-button {selectedTags.includes(tag) ? 'selected' : ''}"
								on:click={() => toggleTag(tag)}
								type="button"
							>
								{tag}
							</button>
						{/each}
					</div>
					{#if selectedTags.length > 0}
						<div class="selected-tags">
							Selected: {selectedTags.join(', ')}
						</div>
					{/if}
				</div>

				<div class="form-group">
					<label for="thumbnailTime">Thumbnail Time (seconds)</label>
					<input 
						id="thumbnailTime"
						type="number" 
						bind:value={thumbnailTime} 
						placeholder="e.g., 30 for 30 seconds"
						min="0"
					/>
					<small>Leave empty to use current thumbnail</small>
				</div>

				<div class="form-actions">
					<button 
						class="btn btn-secondary" 
						on:click={handleBack}
						disabled={saving}
					>
						Cancel
					</button>
					<button 
						class="btn btn-primary" 
						on:click={handleSubmit}
						disabled={!title.trim() || saving}
					>
						{#if saving}
							<LoadingSpinner size="small" />
							Saving...
						{:else}
							Save Changes
						{/if}
					</button>
				</div>
			</div>
		</div>
	{:else}
		<div class="error-container">
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<circle cx="12" cy="12" r="10"></circle>
				<line x1="12" y1="8" x2="12" y2="12"></line>
				<line x1="12" y1="16" x2="12.01" y2="16"></line>
			</svg>
			<h3>Video not found</h3>
			<p>The video you're looking for doesn't exist or has been removed.</p>
			<button class="btn btn-primary" on:click={handleBack}>
				Back to Streaming
			</button>
		</div>
	{/if}
</div>

<style>
	.edit-page {
		padding: 2rem;
		max-width: 1200px;
		margin: 0 auto;
	}

	.page-header {
		margin-bottom: 2rem;
	}

	.back-button {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		background: none;
		border: none;
		color: var(--text-secondary);
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 8px;
		margin-bottom: 1rem;
		transition: all 0.2s ease;
	}

	.back-button:hover {
		background: var(--bg-secondary);
		color: var(--text-primary);
	}

	.back-button svg {
		width: 16px;
		height: 16px;
	}

	.page-header h1 {
		font-size: 2rem;
		font-weight: 700;
		margin: 0 0 0.5rem 0;
		color: var(--text-primary);
	}

	.page-header p {
		margin: 0;
		color: var(--text-secondary);
	}

	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 4rem 2rem;
		color: var(--text-secondary);
	}

	.error-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 4rem 2rem;
		text-align: center;
		color: var(--text-secondary);
	}

	.error-container svg {
		width: 64px;
		height: 64px;
		margin-bottom: 1rem;
		color: var(--error-color);
	}

	.error-container h3 {
		margin: 0 0 0.5rem 0;
		color: var(--text-primary);
	}

	.error-container p {
		margin: 0 0 1.5rem 0;
	}

	.edit-container {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 2rem;
	}

	.video-preview {
		background: var(--bg-card);
		border: 1px solid var(--border-color);
		border-radius: 12px;
		padding: 1.5rem;
	}

	.video-thumbnail {
		position: relative;
		aspect-ratio: 16/9;
		background: var(--bg-secondary);
		border-radius: 8px;
		overflow: hidden;
		margin-bottom: 1rem;
	}

	.video-thumbnail img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.video-duration {
		position: absolute;
		bottom: 0.5rem;
		right: 0.5rem;
		background: rgba(0, 0, 0, 0.8);
		color: white;
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-size: 0.75rem;
		font-weight: 500;
	}

	.video-status {
		position: absolute;
		top: 0.5rem;
		right: 0.5rem;
		background: rgba(0, 0, 0, 0.8);
		color: white;
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-size: 0.75rem;
		font-weight: 500;
	}

	.video-status.success {
		background: var(--success-color);
	}

	.video-status.warning {
		background: var(--warning-color);
	}

	.video-status.error {
		background: var(--error-color);
	}

	.video-info h2 {
		margin: 0 0 1rem 0;
		color: var(--text-primary);
		font-size: 1.25rem;
	}

	.video-stats {
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
		margin-bottom: 1.5rem;
	}

	.stat {
		display: flex;
		align-items: center;
		gap: 0.25rem;
		font-size: 0.875rem;
		color: var(--text-secondary);
	}

	.stat svg {
		width: 16px;
		height: 16px;
	}

	.video-actions {
		display: flex;
		gap: 0.5rem;
	}

	.edit-form {
		background: var(--bg-card);
		border: 1px solid var(--border-color);
		border-radius: 12px;
		padding: 2rem;
	}

	.edit-form h2 {
		margin: 0 0 1.5rem 0;
		color: var(--text-primary);
		font-size: 1.5rem;
	}

	.form-group {
		margin-bottom: 1.5rem;
	}

	.form-group label {
		display: block;
		margin-bottom: 0.5rem;
		font-weight: 500;
		color: var(--text-primary);
	}

	.form-group input,
	.form-group textarea,
	.form-group select {
		width: 100%;
		padding: 0.75rem;
		border: 1px solid var(--border-color);
		border-radius: 8px;
		background: var(--bg-input);
		color: var(--text-primary);
		font-size: 1rem;
		transition: all 0.2s ease;
	}

	.form-group input:focus,
	.form-group textarea:focus,
	.form-group select:focus {
		outline: none;
		border-color: var(--primary-color);
		box-shadow: 0 0 0 3px var(--primary-bg);
	}

	.form-group textarea {
		resize: vertical;
		min-height: 100px;
	}

	.form-group small {
		display: block;
		margin-top: 0.25rem;
		font-size: 0.75rem;
		color: var(--text-tertiary);
	}

	.tags-container {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		margin-bottom: 0.5rem;
	}

	.tag-button {
		padding: 0.5rem 1rem;
		border: 1px solid var(--border-color);
		border-radius: 20px;
		background: var(--bg-input);
		color: var(--text-secondary);
		cursor: pointer;
		font-size: 0.875rem;
		transition: all 0.2s ease;
	}

	.tag-button:hover {
		background: var(--bg-secondary);
		color: var(--text-primary);
	}

	.tag-button.selected {
		background: var(--primary-color);
		color: white;
		border-color: var(--primary-color);
	}

	.selected-tags {
		font-size: 0.875rem;
		color: var(--text-secondary);
		font-style: italic;
	}

	.form-actions {
		display: flex;
		gap: 1rem;
		margin-top: 2rem;
		padding-top: 1.5rem;
		border-top: 1px solid var(--border-color);
	}

	.form-actions .btn {
		flex: 1;
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 8px;
		font-size: 1rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s ease;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
	}

	.btn-primary {
		background: var(--primary-color);
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: var(--primary-hover);
	}

	.btn-primary:disabled {
		background: var(--text-tertiary);
		cursor: not-allowed;
	}

	.btn-secondary {
		background: var(--bg-secondary);
		color: var(--text-primary);
		border: 1px solid var(--border-color);
	}

	.btn-secondary:hover:not(:disabled) {
		background: var(--bg-tertiary);
	}

	.btn-secondary:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.btn-danger {
		background: var(--error-color);
		color: white;
	}

	.btn-danger:hover:not(:disabled) {
		background: var(--error-hover);
	}

	.btn-danger:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.edit-page {
			padding: 1rem;
		}

		.edit-container {
			grid-template-columns: 1fr;
			gap: 1rem;
		}

		.video-preview,
		.edit-form {
			padding: 1.5rem;
		}

		.video-actions {
			flex-direction: column;
		}

		.form-actions {
			flex-direction: column;
		}
	}
</style> 