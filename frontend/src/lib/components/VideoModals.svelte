<script lang="ts">
	import { fade, fly } from 'svelte/transition';
	import { createEventDispatcher } from 'svelte';
	import type { MasterVideo } from '$lib/master-video';

	export let showUploadModal: boolean;
	export let showEditModal: boolean;
	export let showCreateModal: boolean;
	export let selectedVideo: MasterVideo | null;
	export let isUploading: boolean;
	export let uploadProgress: number;
	export let uploadForm: any;
	export let editForm: any;
	export let createForm: any;
	export let categories: string[];

	const dispatch = createEventDispatcher();

	function handleFileSelect(event: Event) {
		const target = event.target as HTMLInputElement;
		if (target.files && target.files[0]) {
			uploadForm.videoFile = target.files[0];
		}
	}

	function handleCreateFileSelect(event: Event) {
		const target = event.target as HTMLInputElement;
		if (target.files && target.files[0]) {
			createForm.videoFile = target.files[0];
		}
	}

	function uploadVideo() {
		dispatch('uploadVideo');
	}

	function saveEdit() {
		dispatch('saveEdit');
	}

	function createVideo() {
		dispatch('createVideo');
	}

	function closeUploadModal() {
		dispatch('closeUploadModal');
	}

	function closeEditModal() {
		dispatch('closeEditModal');
	}

	function closeCreateModal() {
		dispatch('closeCreateModal');
	}
</script>

<!-- Upload Modal -->
{#if showUploadModal}
	<div class="modal-overlay" in:fade={{ duration: 200 }}>
		<div class="modal-content" in:fly={{ y: 20, duration: 200 }}>
			<div class="modal-header">
				<h2>Upload Video</h2>
				<button aria-label="Close"
					class="modal-close"
					on:click={closeUploadModal}
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
							on:click={closeUploadModal}
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
					on:click={closeEditModal}
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
							on:click={closeEditModal}
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

	.form-actions {
		display: flex;
		gap: 1rem;
		justify-content: flex-end;
		margin-top: 2rem;
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

	/* Responsive Design */
	@media (max-width: 768px) {
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
</style>
