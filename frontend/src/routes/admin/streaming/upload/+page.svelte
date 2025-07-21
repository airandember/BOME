<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { apiRequest } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	interface VideoUpload {
		id: string;
		file: File;
		title: string;
		description: string;
		category: string;
		tags: string[];
		thumbnailTime: string;
		status: 'pending' | 'uploading' | 'completed' | 'failed' | 'paused';
		progress: number;
		error?: string;
		uploadedAt?: Date;
		videoId?: string;
	}

	let loading = false;
	let uploading = false;
	let dragActive = false;
	let videos: VideoUpload[] = [];
	let currentUploadIndex = -1;
	let uploadQueue: VideoUpload[] = [];
	let isPaused = false;
	
	// Form data for new video
	let title = '';
	let description = '';
	let category = '';
	let tags = ''; // Changed to string for comma-separated input
	let thumbnailTime = '';

	// Bulk edit form
	let bulkEditForm = {
		title: '',
		category: '',
		description: '',
		tags: ''
	};

	// Edit modal state
	let showEditModal = false;
	let editingVideo: VideoUpload | null = null;
	let editForm = {
		title: '',
		category: '',
		description: '',
		tags: ''
	};

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

	onMount(() => {
		// Check admin permissions
		if (!isAdmin()) {
			showToast('Access denied. Admin privileges required.', 'error');
			goto('/admin/streaming');
			return;
		}

		// Load any saved upload state from localStorage
		loadUploadState();
	});

	function isAdmin(): boolean {
		// This would need to be implemented based on your auth system
		return true; // Placeholder
	}

	function loadUploadState() {
		try {
			const savedState = localStorage.getItem('bome_batch_upload_state');
			if (savedState) {
				const state = JSON.parse(savedState);
				// Only restore pending and failed uploads
				videos = state.videos.filter((v: VideoUpload) => 
					v.status === 'pending' || v.status === 'failed'
				);
				if (videos.length > 0) {
					showToast(`Found ${videos.length} pending uploads. You can resume or clear them.`, 'info');
				}
			}
		} catch (error) {
			console.error('Failed to load upload state:', error);
		}
	}

	function saveUploadState() {
		try {
			const state = {
				videos: videos,
				timestamp: new Date().toISOString()
			};
			localStorage.setItem('bome_batch_upload_state', JSON.stringify(state));
		} catch (error) {
			console.error('Failed to save upload state:', error);
		}
	}

	function clearUploadState() {
		localStorage.removeItem('bome_batch_upload_state');
		videos = [];
		showToast('Upload state cleared', 'success');
	}

	function handleDragEnter(e: DragEvent) {
		e.preventDefault();
		e.stopPropagation();
		dragActive = true;
	}

	function handleDragLeave(e: DragEvent) {
		e.preventDefault();
		e.stopPropagation();
		dragActive = false;
	}

	function handleDragOver(e: DragEvent) {
		e.preventDefault();
		e.stopPropagation();
	}

	function handleDrop(e: DragEvent) {
		e.preventDefault();
		e.stopPropagation();
		dragActive = false;

		if (e.dataTransfer?.files) {
			handleFilesSelect(Array.from(e.dataTransfer.files));
		}
	}

	function handleFilesSelect(files: File[]) {
		const videoFiles = files.filter(file => {
			const validTypes = ['video/mp4', 'video/avi', 'video/mov', 'video/wmv', 'video/flv', 'video/webm', 'video/mkv'];
			if (!validTypes.includes(file.type)) {
				showToast(`Invalid file type: ${file.name}`, 'error');
				return false;
			}

			const maxSize = 500 * 1024 * 1024; // 500MB
			if (file.size > maxSize) {
				showToast(`File too large: ${file.name} (max 500MB)`, 'error');
				return false;
			}

			return true;
		});

		videoFiles.forEach(file => {
			const videoUpload: VideoUpload = {
				id: generateId(),
				file: file,
				title: file.name.replace(/\.[^/.]+$/, ''), // Remove extension
				description: '',
				category: '',
				tags: [],
				thumbnailTime: '',
				status: 'pending',
				progress: 0
			};
			videos = [...videos, videoUpload];
		});

		saveUploadState();
		showToast(`Added ${videoFiles.length} video(s) to upload queue`, 'success');
	}

	function generateId(): string {
		return Math.random().toString(36).substr(2, 9);
	}

	function toggleTag(tag: string) {
		// This function is no longer needed since we're using comma-separated input
		// Keeping for backward compatibility but it's not used
	}

	function updateVideo(videoId: string, updates: Partial<VideoUpload>) {
		videos = videos.map(video => 
			video.id === videoId ? { ...video, ...updates } : video
		);
		saveUploadState();
	}

	function applyBulkEdit() {
		if (videos.length === 0) {
			showToast('No videos to edit', 'warning');
			return;
		}

		let updatedCount = 0;

		videos.forEach(video => {
			const updates: Partial<VideoUpload> = {};
			
			if (bulkEditForm.title.trim()) {
				updates.title = bulkEditForm.title.trim();
			}
			if (bulkEditForm.category.trim()) {
				updates.category = bulkEditForm.category.trim();
			}
			if (bulkEditForm.description.trim()) {
				updates.description = bulkEditForm.description.trim();
			}
			if (bulkEditForm.tags.trim()) {
				updates.tags = bulkEditForm.tags.split(',').map(tag => tag.trim()).filter(tag => tag.length > 0);
			}

			if (Object.keys(updates).length > 0) {
				updateVideo(video.id, updates);
				updatedCount++;
			}
		});

		if (updatedCount > 0) {
			showToast(`Applied bulk edit to ${updatedCount} video(s)`, 'success');
			// Clear the bulk edit form
			bulkEditForm = {
				title: '',
				category: '',
				description: '',
				tags: ''
			};
		} else {
			showToast('No changes to apply', 'info');
		}
	}

	function removeVideo(videoId: string) {
		videos = videos.filter(video => video.id !== videoId);
		saveUploadState();
	}

	function applyToAll(updates: Partial<VideoUpload>) {
		videos = videos.map(video => ({ ...video, ...updates }));
		saveUploadState();
	}

	async function startBatchUpload() {
		if (videos.length === 0) {
			showToast('No videos to upload', 'error');
			return;
		}

		const pendingVideos = videos.filter(v => v.status === 'pending' || v.status === 'failed');
		if (pendingVideos.length === 0) {
			showToast('No pending videos to upload', 'info');
			return;
		}

		uploading = true;
		isPaused = false;
		uploadQueue = [...pendingVideos];

		// Reset failed videos
		videos = videos.map(video => 
			video.status === 'failed' ? { ...video, status: 'pending', progress: 0, error: undefined } : video
		);

		await processUploadQueue();
	}

	async function processUploadQueue() {
		for (let i = 0; i < uploadQueue.length; i++) {
			if (isPaused) {
				showToast('Upload paused', 'info');
				break;
			}

			const video = uploadQueue[i];
			currentUploadIndex = videos.findIndex(v => v.id === video.id);
			
			if (currentUploadIndex === -1) continue;

			updateVideo(video.id, { status: 'uploading', progress: 0 });

			try {
				await uploadVideo(video);
				updateVideo(video.id, { 
					status: 'completed', 
					progress: 100, 
					uploadedAt: new Date() 
				});
				showToast(`Uploaded: ${video.title}`, 'success');
			} catch (error) {
				console.error('Upload failed:', error);
				updateVideo(video.id, { 
					status: 'failed', 
					error: error instanceof Error ? error.message : 'Upload failed' 
				});
				showToast(`Failed to upload: ${video.title}`, 'error');
			}
		}

		uploading = false;
		currentUploadIndex = -1;
		
		const completedCount = videos.filter(v => v.status === 'completed').length;
		const failedCount = videos.filter(v => v.status === 'failed').length;
		
		if (completedCount > 0) {
			showToast(`Batch upload completed: ${completedCount} successful, ${failedCount} failed`, 'success');
		}
	}

	async function uploadVideo(video: VideoUpload): Promise<void> {
		return new Promise((resolve, reject) => {
			const formData = new FormData();
			formData.append('video', video.file);
			formData.append('title', video.title);
			formData.append('description', video.description);
			formData.append('category', video.category);
			formData.append('tags', JSON.stringify(video.tags));
			if (video.thumbnailTime) {
				formData.append('thumbnailTime', video.thumbnailTime);
			}

			const xhr = new XMLHttpRequest();
			
			xhr.upload.addEventListener('progress', (e) => {
				if (e.lengthComputable) {
					const progress = (e.loaded / e.total) * 100;
					updateVideo(video.id, { progress });
				}
			});

			xhr.addEventListener('load', () => {
				if (xhr.status >= 200 && xhr.status < 300) {
					try {
						const result = JSON.parse(xhr.responseText);
						updateVideo(video.id, { videoId: result.video_id });
						resolve();
					} catch (error) {
						reject(new Error('Failed to parse response'));
					}
				} else {
					try {
						const error = JSON.parse(xhr.responseText);
						reject(new Error(error.error || 'Upload failed'));
					} catch {
						reject(new Error(`HTTP ${xhr.status}: Upload failed`));
					}
				}
			});

			xhr.addEventListener('error', () => {
				reject(new Error('Network error'));
			});

			xhr.addEventListener('abort', () => {
				reject(new Error('Upload aborted'));
			});

			xhr.open('POST', '/api/v1/videos/upload');
			
			// Add auth header
			const tokens = JSON.parse(localStorage.getItem('bome_auth_tokens') || 'null');
			if (tokens?.access_token) {
				xhr.setRequestHeader('Authorization', `Bearer ${tokens.access_token}`);
			}
			
			xhr.send(formData);
		});
	}

	function pauseUpload() {
		isPaused = true;
		showToast('Upload paused', 'info');
	}

	function resumeUpload() {
		if (isPaused) {
			isPaused = false;
			processUploadQueue();
		}
	}

	function retryFailed() {
		const failedVideos = videos.filter(v => v.status === 'failed');
		if (failedVideos.length === 0) {
			showToast('No failed uploads to retry', 'info');
			return;
		}

		// Reset failed videos to pending
		videos = videos.map(video => 
			video.status === 'failed' ? { ...video, status: 'pending', progress: 0, error: undefined } : video
		);
		saveUploadState();
		showToast(`Reset ${failedVideos.length} failed upload(s)`, 'success');
	}

	function getStatusIcon(status: string) {
		switch (status) {
			case 'pending': return '⏳';
			case 'uploading': return '📤';
			case 'completed': return '✅';
			case 'failed': return '❌';
			case 'paused': return '⏸️';
			default: return '❓';
		}
	}

	function getStatusColor(status: string) {
		switch (status) {
			case 'pending': return '#f59e0b';
			case 'uploading': return '#3b82f6';
			case 'completed': return '#10b981';
			case 'failed': return '#ef4444';
			case 'paused': return '#6b7280';
			default: return '#6b7280';
		}
	}

	function formatFileSize(bytes: number): string {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	}

	function editVideo(video: VideoUpload) {
		editingVideo = video;
		editForm = {
			title: video.title,
			category: video.category,
			description: video.description,
			tags: video.tags.join(', ')
		};
		showEditModal = true;
	}

	function saveEdit() {
		if (!editingVideo) return;

		const updates: Partial<VideoUpload> = {
			title: editForm.title.trim(),
			category: editForm.category.trim(),
			description: editForm.description.trim(),
			tags: editForm.tags.split(',').map(tag => tag.trim()).filter(tag => tag.length > 0)
		};

		updateVideo(editingVideo.id, updates);
		showToast(`Updated: ${editForm.title}`, 'success');
		closeEditModal();
	}

	function closeEditModal() {
		showEditModal = false;
		editingVideo = null;
		editForm = {
			title: '',
			category: '',
			description: '',
			tags: ''
		};
	}

	function handleBack() {
		goto('/admin/streaming');
	}
</script>

<svelte:head>
	<title>Batch Upload Videos - Streaming Management</title>
	<meta name="description" content="Batch upload videos to Bunny.net streaming platform" />
</svelte:head>

<div class="batch-upload-page">
	<header class="page-header">
		<button class="back-button" on:click={handleBack}>
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M19 12H5"></path>
				<path d="M12 19l-7-7 7-7"></path>
			</svg>
			Back to Streaming
		</button>
		<h1>Batch Upload Videos</h1>
		<p>Upload multiple videos with progress tracking and resume capability</p>
	</header>

	<!-- File Drop Zone -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div aria-label="file drop zone"
		class="drop-zone {dragActive ? 'drag-active' : ''}"
		on:dragenter={handleDragEnter}
		on:dragleave={handleDragLeave}
		on:dragover={handleDragOver}
		on:drop={handleDrop}
	>
		<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
			<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
			<polyline points="7,10 12,15 17,10"></polyline>
			<line x1="12" y1="15" x2="12" y2="3"></line>
		</svg>
		<h3>Drop video files here</h3>
		<p>or click to browse multiple files</p>
		<input 
			type="file" 
			accept="video/*" 
			multiple
			on:change={(e) => {
				const target = e.target as HTMLInputElement;
				if (target.files) {
					handleFilesSelect(Array.from(target.files));
				}
			}}
			style="display: none;"
			id="file-input"
		/>
		<label for="file-input" class="browse-button">Browse Files</label>
	</div>

	<!-- Batch Actions -->
	{#if videos.length > 0}
		<div class="batch-actions">
			<div class="batch-stats">
				<span class="stat">
					<span class="stat-label">Total:</span>
					<span class="stat-value">{videos.length}</span>
				</span>
				<span class="stat">
					<span class="stat-label">Pending:</span>
					<span class="stat-value">{videos.filter(v => v.status === 'pending').length}</span>
				</span>
				<span class="stat">
					<span class="stat-label">Completed:</span>
					<span class="stat-value">{videos.filter(v => v.status === 'completed').length}</span>
				</span>
				<span class="stat">
					<span class="stat-label">Failed:</span>
					<span class="stat-value">{videos.filter(v => v.status === 'failed').length}</span>
				</span>
			</div>
			
			<div class="action-buttons">
				{#if !uploading}
					<button class="btn btn-primary" on:click={startBatchUpload} disabled={videos.filter(v => v.status === 'pending' || v.status === 'failed').length === 0}>
						Start Upload
					</button>
				{:else}
					{#if isPaused}
						<button class="btn btn-primary" on:click={resumeUpload}>
							Resume Upload
						</button>
					{:else}
						<button class="btn btn-secondary" on:click={pauseUpload}>
							Pause Upload
						</button>
					{/if}
				{/if}
				
				<button class="btn btn-secondary" on:click={retryFailed} disabled={videos.filter(v => v.status === 'failed').length === 0}>
					Retry Failed
				</button>
				
				<button class="btn btn-danger" on:click={clearUploadState}>
					Clear All
				</button>
			</div>
		</div>

		<!-- Videos Table -->
		<div class="videos-table-container">
			<table class="videos-table">
				<thead>
					<tr>
						<th>Status</th>
						<th>File</th>
						<th>Title</th>
						<th>Category</th>
						<th>Tags</th>
						<th>Progress</th>
						<th>Actions</th>
					</tr>
				</thead>
				<tbody>
					{#each videos as video, index}
						<tr class="video-row {video.status} {index === currentUploadIndex ? 'current-upload' : ''}">
							<td class="status-cell">
								<span class="status-icon" style="color: {getStatusColor(video.status)}">
									{getStatusIcon(video.status)}
								</span>
								<span class="status-text">{video.status}</span>
							</td>
							
							<td class="file-cell">
								<div class="file-info">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<polygon points="23,7 16,12 23,17 23,7"></polygon>
										<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
									</svg>
									<span class="filename">{video.file.name}</span>
								</div>
							</td>
							
							<td class="title-cell">
								<input 
									type="text" 
									bind:value={video.title}
									on:blur={() => updateVideo(video.id, { title: video.title })}
									placeholder="Enter title"
									disabled={video.status === 'uploading'}
								/>
							</td>
							
							<td class="category-cell">
								<select 
									bind:value={video.category}
									on:change={() => updateVideo(video.id, { category: video.category })}
									disabled={video.status === 'uploading'}
								>
									<option value="">Select category</option>
									{#each categories as cat}
										<option value={cat}>{cat}</option>
									{/each}
								</select>
							</td>
							
							<td class="tags-cell title-cell">
								<input 
									type="text" 
									value={video.tags.join(', ')}
									on:blur={(e) => {
										const target = e.target as HTMLInputElement;
										const tags = target.value.split(',').map(tag => tag.trim()).filter(tag => tag.length > 0);
										updateVideo(video.id, { tags });
									}}
									placeholder="Enter tags (comma-separated)"
									disabled={video.status === 'uploading'}
								/>
							</td>
							
							<td class="progress-cell">
								{#if video.status === 'uploading' || video.status === 'completed'}
									<div class="progress-bar">
										<div class="progress-fill" style="width: {video.progress}%"></div>
									</div>
									<span class="progress-text">{video.progress.toFixed(1)}%</span>
								{:else if video.status === 'failed'}
									<span class="error-text">{video.error}</span>
								{:else}
									<span class="pending-text">Pending</span>
								{/if}
							</td>
							
							<td class="actions-cell">
								<button 
									class="action-btn edit-btn"
									on:click={() => editVideo(video)}
									disabled={video.status === 'uploading'}
									title="Edit metadata"
									aria-label="Edit metadata"
								>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
										<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
									</svg>
								</button>
								
								<button aria-label="Remove from queue"
									class="action-btn remove-btn"
									on:click={() => removeVideo(video.id)}
									disabled={video.status === 'uploading'}
									title="Remove from queue"
								>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<line x1="18" y1="6" x2="6" y2="18"></line>
										<line x1="6" y1="6" x2="18" y2="18"></line>
									</svg>
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}

	<!-- Bulk Edit Modal -->
	{#if videos.length > 0}
		<div class="bulk-edit-section">
			<h3>Bulk Edit</h3>
			<div class="bulk-edit-form">
				<div class="form-row">
					<div class="form-group">
						<label for="bulk-title">Title</label>
						<input 
							type="text" 
							id="bulk-title"
							bind:value={bulkEditForm.title}
							placeholder="Leave blank to keep existing titles"
						/>
					</div>
					<div class="form-group">
						<label for="bulk-category">Category</label>
						<select id="bulk-category" bind:value={bulkEditForm.category}>
							<option value="">-- Keep Existing --</option>
							{#each categories as cat}
								<option value={cat}>{cat}</option>
							{/each}
						</select>
					</div>
				</div>
				
				<div class="form-group">
					<label for="bulk-description">Description</label>
					<textarea 
						id="bulk-description"
						bind:value={bulkEditForm.description}
						rows="3"
						placeholder="Leave blank to keep existing descriptions"
					></textarea>
				</div>
				
				<div class="form-group">
					<label for="bulk-tags">Tags (comma-separated)</label>
					<input 
						type="text" 
						id="bulk-tags"
						bind:value={bulkEditForm.tags}
						placeholder="tag1, tag2, tag3 (leave blank to keep existing)"
					/>
				</div>
				
				<div class="bulk-actions">
					<button class="btn btn-primary" on:click={applyBulkEdit}>
						Apply to All Videos ({videos.length})
					</button>
					<button class="btn btn-secondary" on:click={() => {
						bulkEditForm = {
							title: '',
							category: '',
							description: '',
							tags: ''
						};
					}}>
						Clear Form
					</button>
				</div>
			</div>
		</div>
	{/if}
</div>

<!-- Edit Metadata Modal -->

{#if showEditModal && editingVideo} 
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div class="modal-overlay" on:click={closeEditModal}>
		<div class="modal-content" on:click|stopPropagation>
			<div class="modal-header">
				<h2>Edit Video Metadata</h2>
				<button class="modal-close" on:click={closeEditModal} aria-label="Close modal">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"></line>
						<line x1="6" y1="6" x2="18" y2="18"></line>
					</svg>
				</button>
			</div>

			<div class="modal-body">
				<!--<div class="file-preview">
					<div class="file-info">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<polygon points="23,7 16,12 23,17 23,7"></polygon>
							<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
						</svg>
						<span class="filename">{editingVideo.file.name}</span>
					</div>
					<div class="file-size">{formatFileSize(editingVideo.file.size)}</div>
				</div>-->

				<div class="edit-form">
					<div class="form-group">
						<label for="edit-title">Title *</label>
						<input 
							type="text" 
							id="edit-title"
							bind:value={editForm.title}
							placeholder="Enter video title"
							required
						/>
					</div>

					<div class="form-group">
						<label for="edit-category">Category</label>
						<select id="edit-category" bind:value={editForm.category}>
							<option value="">Select a category (optional)</option>
							{#each categories as cat}
								<option value={cat}>{cat}</option>
							{/each}
						</select>
					</div>

					<div class="form-group">
						<label for="edit-description">Description</label>
						<textarea 
							id="edit-description"
							bind:value={editForm.description}
							rows="6"
							placeholder="Enter a detailed description of the video content..."
						></textarea>
						<div class="help-text">Provide a comprehensive description to help viewers understand the video content.</div>
					</div>
					<hr style="width: 100%;">
					<div class="form-group">
						<label for="edit-tags">Tags</label>
						<textarea  
							id="edit-tags"
							rows="3"
							bind:value={editForm.tags}
							placeholder="Enter tags separated by commas (e.g., Book of Mormon, Archaeology, Research)"
						></textarea>
						<div class="help-text">Use tags to help categorize and search for this video. Separate multiple tags with commas.</div>
						
						<div class="suggested-tags">
							<div class="suggested-tags-label">Suggested tags:</div>
							<div class="tag-suggestions">
								{#each availableTags.slice(0, 8) as tag}
									<button 
										type="button"
										class="tag-suggestion"
										on:click={() => {
											const currentTags = editForm.tags ? editForm.tags.split(',').map(t => t.trim()) : [];
											if (!currentTags.includes(tag)) {
												const newTags = [...currentTags, tag];
												editForm.tags = newTags.join(', ');
											}
										}}
									>
										{tag}
									</button>
								{/each}
							</div>
						</div>
					</div>
				</div>
			</div>

			<div class="modal-footer">
				<button class="btn btn-secondary" on:click={closeEditModal}>
					Cancel
				</button>
				<button class="btn btn-primary" on:click={saveEdit}>
					Save Changes
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.batch-upload-page {
		padding: 2rem;
		max-width: 1400px;
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

	.drop-zone {
		border: 2px dashed var(--border-color);
		border-radius: 12px;
		padding: 3rem;
		text-align: center;
		margin-bottom: 2rem;
		transition: all 0.3s ease;
		background: var(--bg-secondary);
	}

	.drop-zone.drag-active {
		border-color: var(--primary);
		background: var(--primary-light);
	}

	.drop-zone svg {
		width: 48px;
		height: 48px;
		color: var(--text-secondary);
		margin-bottom: 1rem;
	}

	.drop-zone h3 {
		margin: 0 0 0.5rem 0;
		color: var(--text-primary);
	}

	.drop-zone p {
		margin: 0 0 1.5rem 0;
		color: var(--text-secondary);
	}

	.browse-button {
		display: inline-block;
		background: var(--primary);
		color: white;
		padding: 0.75rem 1.5rem;
		border-radius: 8px;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
	}

	.browse-button:hover {
		background: var(--primary-dark);
	}

	.batch-actions {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 2rem;
		padding: 1rem;
		background: var(--bg-secondary);
		border-radius: 12px;
		border: 1px solid var(--border-color);
	}

	.batch-stats {
		display: flex;
		gap: 2rem;
	}

	.stat {
		display: flex;
		flex-direction: column;
		align-items: center;
	}

	.stat-label {
		font-size: 0.875rem;
		color: var(--text-secondary);
		margin-bottom: 0.25rem;
	}

	.stat-value {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text-primary);
	}

	.action-buttons {
		display: flex;
		gap: 1rem;
	}

	.btn {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 8px;
		cursor: pointer;
		font-weight: 500;
		transition: all 0.2s ease;
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.btn-primary {
		background: var(--primary);
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: var(--primary-dark);
	}

	.btn-secondary {
		background: var(--bg-primary);
		color: var(--text-primary);
		border: 1px solid var(--border-color);
	}

	.btn-secondary:hover:not(:disabled) {
		background: var(--bg-secondary);
	}

	.btn-danger {
		background: #ef4444;
		color: white;
	}

	.btn-danger:hover:not(:disabled) {
		background: #dc2626;
	}

	.btn-sm {
		padding: 0.5rem 1rem;
		font-size: 0.875rem;
	}

	.videos-table-container {
		background: var(--bg-primary);
		border-radius: 12px;
		border: 1px solid var(--border-color);
		overflow: hidden;
		margin-bottom: 2rem;
	}

	.videos-table {
		width: 100%;
		border-collapse: collapse;
	}

	.videos-table th {
		background: var(--bg-secondary);
		padding: 1rem;
		text-align: left;
		font-weight: 600;
		color: var(--text-primary);
		border-bottom: 1px solid var(--border-color);
	}

	.videos-table td {
		padding: 1rem;
		border-bottom: 1px solid var(--border-color);
		vertical-align: middle;
	}

	.video-row:hover {
		background: var(--bg-secondary);
	}

	.video-row.current-upload {
		background: var(--primary-light);
	}

	.status-cell {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.status-icon {
		font-size: 1.25rem;
	}

	.status-text {
		font-weight: 500;
		text-transform: capitalize;
	}

	.file-cell .file-info {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.file-info svg {
		width: 20px;
		height: 20px;
		color: var(--text-secondary);
	}

	.filename {
		font-weight: 500;
		color: var(--text-primary);
	}

	.title-cell input,
	.category-cell select {
		width: 100%;
		padding: 0.5rem;
		border: 1px solid var(--border-color);
		border-radius: 6px;
		background: var(--bg-primary);
		color: var(--text-primary);
	}

	.title-cell input:disabled,
	.category-cell select:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.tags-display {
		display: flex;
		flex-wrap: wrap;
		gap: 0.25rem;
	}

	.tag {
		background: var(--primary-light);
		color: var(--primary);
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-size: 0.75rem;
		font-weight: 500;
	}

	.no-tags {
		color: var(--text-secondary);
		font-style: italic;
		font-size: 0.875rem;
	}

	.progress-bar {
		width: 100%;
		height: 8px;
		background: var(--bg-secondary);
		border-radius: 4px;
		overflow: hidden;
		margin-bottom: 0.5rem;
	}

	.progress-fill {
		height: 100%;
		background: var(--primary);
		transition: width 0.3s ease;
	}

	.progress-text {
		font-size: 0.875rem;
		color: var(--text-secondary);
	}

	.error-text {
		color: #ef4444;
		font-size: 0.875rem;
	}

	.pending-text {
		color: var(--text-secondary);
		font-size: 0.875rem;
	}

	.actions-cell {
		display: flex;
		gap: 0.5rem;
	}

	.action-btn {
		padding: 0.5rem;
		border: none;
		border-radius: 6px;
		cursor: pointer;
		transition: all 0.2s ease;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.action-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.edit-btn {
		background: var(--bg-secondary);
		color: var(--text-secondary);
	}

	.edit-btn:hover:not(:disabled) {
		background: var(--primary-light);
		color: var(--primary);
	}

	.remove-btn {
		background: var(--bg-secondary);
		color: var(--text-secondary);
	}

	.remove-btn:hover:not(:disabled) {
		background: #fef2f2;
		color: #ef4444;
	}

	.action-btn svg {
		width: 16px;
		height: 16px;
	}

	.bulk-edit-section {
		background: var(--bg-secondary);
		border-radius: 12px;
		padding: 1.5rem;
		border: 1px solid var(--border-color);
	}

	.bulk-edit-section h3 {
		margin: 0 0 1rem 0;
		color: var(--text-primary);
	}

	.bulk-edit-form .form-group {
		margin-bottom: 1rem;
	}

	.bulk-edit-form label {
		display: block;
		margin-bottom: 0.5rem;
		color: var(--text-primary);
		font-weight: 500;
	}

	.bulk-actions {
		display: flex;
		gap: 1rem;
		flex-wrap: wrap;
	}

	/* Responsive design */
	@media (max-width: 768px) {
		.batch-upload-page {
			padding: 1rem;
		}

		.batch-actions {
			flex-direction: column;
			gap: 1rem;
		}

		.batch-stats {
			justify-content: space-around;
		}

		.action-buttons {
			justify-content: center;
		}

		.videos-table {
			font-size: 0.875rem;
		}

		.videos-table th,
		.videos-table td {
			padding: 0.75rem 0.5rem;
		}

		.bulk-actions {
			flex-direction: column;
		}
	}

	/* Modal Styles */
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.7);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
		padding: 1rem;
	}

	.modal-content {
		background: var(--bg-primary);
		border-radius: 16px;
		box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
		max-width: 700px;
		width: 100%;
		max-height: 90vh;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		border: 1px solid var(--border-color);
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1.5rem 2rem;
		border-bottom: 1px solid var(--border-color);
		background: var(--bg-secondary);
	}

	.modal-header h2 {
		margin: 0;
		font-size: 1.5rem;
		font-weight: 600;
		color: var(--text-primary);
	}

	.modal-close {
		background: none;
		border: none;
		color: var(--text-secondary);
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 8px;
		transition: all 0.2s ease;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.modal-close:hover {
		background: var(--bg-primary);
		color: var(--text-primary);
	}

	.modal-close svg {
		width: 20px;
		height: 20px;
	}

	.modal-body {
		padding: 2rem;
		overflow-y: auto;
		flex: 1;
	}

	.file-preview {
		background: var(--bg-secondary);
		border-radius: 12px;
		padding: 1.5rem;
		margin-bottom: 2rem;
		border: 1px solid var(--border-color);
	}

	.file-preview .file-info {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 0.5rem;
	}

	.file-preview .file-info svg {
		width: 24px;
		height: 24px;
		color: var(--primary);
	}

	.file-preview .filename {
		font-weight: 600;
		color: var(--text-primary);
		font-size: 1.1rem;
	}

	.file-preview .file-size {
		color: var(--text-secondary);
		font-size: 0.9rem;
	}

	.edit-form {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.edit-form .form-group {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.edit-form label {
		font-weight: 600;
		color: var(--text-primary);
		font-size: 0.95rem;
	}

	.edit-form input,
	.edit-form select,
	.edit-form textarea {
		padding: 0.75rem;
		border: 2px solid var(--border-color);
		border-radius: 8px;
		background: var(--bg-tertiary);
		color: var(--text-primary);
		font-size: 1rem;
		transition: all 0.2s ease;
	}

	.edit-form input:focus,
	.edit-form select:focus,
	.edit-form textarea:focus {
		outline: none;
		border-color: var(--primary);
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	.edit-form textarea {
		resize: vertical;
		min-height: 120px;
		font-family: inherit;
		line-height: 1.5;
	}

	.help-text {
		font-size: 0.85rem;
		color: var(--text-secondary);
		margin-top: 0.25rem;
		line-height: 1.4;
	}

	.suggested-tags {
		margin-top: 1rem;
	}

	.suggested-tags-label {
		font-size: 0.9rem;
		font-weight: 500;
		color: var(--text-primary);
		margin-bottom: 0.75rem;
	}

	.tag-suggestions {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	.tag-suggestion {
		background: var(--bg-secondary);
		border: 1px solid var(--border-color);
		color: var(--text-primary);
		padding: 0.5rem 0.75rem;
		border-radius: 6px;
		font-size: 0.85rem;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.tag-suggestion:hover {
		background: var(--primary-light);
		border-color: var(--primary);
		color: var(--primary);
	}

	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: 1rem;
		padding: 1.5rem 2rem;
		border-top: 1px solid var(--border-color);
		background: var(--bg-secondary);
	}

	/* Responsive modal */
	@media (max-width: 768px) {
		.modal-overlay {
			padding: 0.5rem;
		}

		.modal-content {
			max-height: 95vh;
		}

		.modal-header,
		.modal-footer {
			padding: 1rem 1.5rem;
		}

		.modal-body {
			padding: 1.5rem;
		}

		.modal-header h2 {
			font-size: 1.25rem;
		}

		.file-preview {
			padding: 1rem;
		}

		.edit-form {
			gap: 1rem;
		}

		.tag-suggestions {
			gap: 0.375rem;
		}

		.tag-suggestion {
			padding: 0.375rem 0.5rem;
			font-size: 0.8rem;
		}
	}
</style> 