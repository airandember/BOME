<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { masterVideoService, type MasterVideo } from '$lib/master-video';
	import { showToast } from '$lib/toast';

	// Props
	export let video: MasterVideo | null = null;
	export let isOpen: boolean = false;

	// Events
	const dispatch = createEventDispatcher<{
		close: void;
		videoUpdated: { video: MasterVideo; tags: string[] };
	}>();

	// Local state
	let taggingResult: any = null;
	let isTagging = false;
	let newTag = '';
	let isAddingTag = false;
	let isRemovingTag = false;

	// Close modal
	function closeModal() {
		dispatch('close');
	}

	// Reset modal state when video changes
	$: if (video && isOpen) {
		taggingResult = null;
		newTag = '';
	}

	// Add new tag to video
	async function addTag() {
		if (!video || !newTag.trim()) return;
		
		try {
			isAddingTag = true;
			const currentTags = video.Tags || [];
			const updatedTags = [...currentTags, newTag.trim()];
			
			// Update video tags in database
			const response = await masterVideoService.updateMasterVideo(video.ID, { Tags: updatedTags });
			if (response.success) {
				// Update local video object
				const updatedVideo = { ...video, Tags: updatedTags, Tagged: true };
				dispatch('videoUpdated', {
					video: updatedVideo,
					tags: updatedTags
				});
				
				// Reset form
				newTag = '';
				showToast('✅ Tag added successfully', 'success');
			} else {
				showToast('❌ Failed to add tag', 'error');
			}
		} catch (error) {
			console.error('Failed to add tag:', error);
			showToast('❌ Failed to add tag', 'error');
		} finally {
			isAddingTag = false;
		}
	}

	// Remove tag from video
	async function removeTag(tagToRemove: string) {
		if (!video) return;
		
		try {
			isRemovingTag = true;
			const currentTags = video.Tags || [];
			const updatedTags = currentTags.filter(tag => tag !== tagToRemove);
			
			// Update video tags in database
			const response = await masterVideoService.updateMasterVideo(video.ID, { Tags: updatedTags });
			if (response.success) {
				// Update local video object
				const updatedVideo = { ...video, Tags: updatedTags, Tagged: updatedTags.length > 0 };
				dispatch('videoUpdated', {
					video: updatedVideo,
					tags: updatedTags
				});
				
				showToast('✅ Tag removed successfully', 'success');
			} else {
				showToast('❌ Failed to remove tag', 'error');
			}
		} catch (error) {
			console.error('Failed to remove tag:', error);
			showToast('❌ Failed to remove tag', 'error');
		} finally {
			isRemovingTag = false;
		}
	}

	// Auto-tag video
	async function autoTagVideo() {
		if (!video) return;

		try {
			isTagging = true;
			const response = await masterVideoService.autoTagVideo(video.ID);
			
			if (response.success) {
				taggingResult = response.result;
				showToast(`✅ ${response.message}`, 'success');
				
				// Dispatch event to parent
				if (response.result) {
					dispatch('videoUpdated', {
						video: { ...video, Tags: response.result.tags, Tagged: true },
						tags: response.result.tags
					});
				}
			} else {
				showToast(`❌ ${response.message}`, 'error');
			}
		} catch (error) {
			console.error('Failed to tag video:', error);
			showToast('❌ Failed to tag video', 'error');
		} finally {
			isTagging = false;
		}
	}
</script>

{#if isOpen}
	<div class="modal-backdrop" on:click={closeModal} transition:fade={{ duration: 200 }}>
		<div class="modal-content" on:click|stopPropagation transition:fly={{ y: 20, duration: 200 }}>
			<div class="modal-header">
				<h2>🏷️ Smart Tagging</h2>
				<button class="modal-close" on:click={closeModal}>×</button>
			</div>
			
			<div class="modal-body">
				{#if video}
					<div class="video-info">
						<h3>Video: {video.Title}</h3>
						<p><strong>Tagged:</strong> {video.Tagged ? '✅ Yes' : '❌ No'}</p>
					</div>

					<!-- Current Tags Management -->
					<div class="current-tags-section">
						<h4>Current Tags</h4>
						{#if video.Tags && video.Tags.length > 0}
							<div class="tags-display">
								{#each video.Tags as tag}
									<div class="tag-chip">
										<span class="tag-text">{tag}</span>
										<button 
											class="remove-tag-btn" 
											on:click={() => removeTag(tag)}
											disabled={isRemovingTag}
											title="Remove tag"
										>
											❌
										</button>
									</div>
								{/each}
							</div>
						{:else}
							<p class="no-tags">No tags assigned</p>
						{/if}
					</div>

					<!-- Add New Tag -->
					<div class="add-tag-section">
						<h4>Add New Tag</h4>
						<div class="add-tag-form">
							<input 
								type="text" 
								bind:value={newTag} 
								placeholder="Enter new tag..."
								on:keydown={(e) => e.key === 'Enter' && addTag()}
								disabled={isAddingTag}
							/>
							<button 
								on:click={addTag} 
								disabled={!newTag.trim() || isAddingTag}
								class="add-tag-btn"
							>
								{isAddingTag ? 'Adding...' : 'Add Tag'}
							</button>
						</div>
					</div>

					{#if !taggingResult}
						<div class="tagging-actions">
							<button 
								class="btn bg-orange-500 text-white px-6 py-3 rounded-lg hover:bg-orange-600 transition-colors disabled:opacity-50"
								on:click={autoTagVideo}
								disabled={isTagging}
							>
								{#if isTagging}
									<div class="flex items-center space-x-2">
										<div class="animate-spin rounded-full h-5 w-5 border-b-2 border-white"></div>
										<span>Processing...</span>
									</div>
								{:else}
									🚀 Generate Smart Tags
								{/if}
							</button>
						</div>
					{:else}
						<div class="tagging-result">
							<h4>✅ Tagging Complete!</h4>
							<div class="result-details">
								<p><strong>Name:</strong> {taggingResult.name}</p>
								<p><strong>Generated Tags:</strong></p>
								<div class="tags-display">
									{#each taggingResult.tags as tag}
										<span class="tag-chip">{tag}</span>
									{/each}
								</div>
								<p><strong>Original Title:</strong> {taggingResult.original_title}</p>
								<p><strong>Processed:</strong> {taggingResult.processed_title}</p>
							</div>
						</div>
					{/if}
				{/if}
			</div>

			<div class="modal-footer">
				<button class="btn bg-gray-500 text-white px-4 py-2 rounded-lg hover:bg-gray-600 transition-colors" on:click={closeModal}>
					Close
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.modal-backdrop {
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
		max-width: 600px;
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

	.modal-body {
		flex: 1;
	}

	.video-info h3 {
		margin: 0 0 1rem 0;
		color: var(--text-primary);
		font-size: 1.25rem;
	}

	.video-info p {
		margin: 0.5rem 0;
		color: var(--text-secondary);
	}

	.tagging-actions {
		text-align: center;
		margin: 2rem 0;
	}

	.tagging-result {
		background: rgba(16, 185, 129, 0.1);
		border: 1px solid rgba(16, 185, 129, 0.2);
		border-radius: 8px;
		padding: 1.5rem;
	}

	.tagging-result h4 {
		margin: 0 0 1rem 0;
		color: #10b981;
		font-size: 1.1rem;
	}

	.result-details p {
		margin: 0.5rem 0;
		color: var(--text-secondary);
	}

	/* Current Tags Management */
	.current-tags-section,
	.add-tag-section {
		margin: 1.5rem 0;
		padding: 1rem;
		background: rgba(255, 255, 255, 0.05);
		border-radius: 8px;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.current-tags-section h4,
	.add-tag-section h4 {
		margin: 0 0 1rem 0;
		color: var(--text-primary);
		font-size: 1.1rem;
	}

	.tags-display {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		margin-bottom: 1rem;
	}

	.tag-chip {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		background: var(--primary-color);
		color: var(--text-secondary);
		padding: 0.5rem 0.75rem;
		border-radius: 20px;
		font-size: 0.9rem;
	}

	.tag-text {
		font-weight: 500;
	}

	.remove-tag-btn {
		background: none;
		border: none;
		color: inherit;
		cursor: pointer;
		padding: 0;
		font-size: 0.8rem;
		opacity: 0.8;
		transition: opacity 0.2s ease;
	}

	.remove-tag-btn:hover {
		opacity: 1;
	}

	.remove-tag-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.no-tags {
		color: var(--text-secondary);
		font-style: italic;
		margin: 0;
	}

	.add-tag-form {
		display: flex;
		gap: 0.5rem;
		align-items: center;
	}

	.add-tag-form input {
		flex: 1;
		padding: 0.5rem 1rem;
		border: 1px solid var(--border-color);
		border-radius: 6px;
		background: var(--bg-primary);
		color: var(--text-primary);
		font-size: 0.9rem;
	}

	.add-tag-form input:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.add-tag-btn {
		background: var(--primary-color);
		color: var(--text-secondary);
		border: none;
		padding: 0.5rem 1rem;
		border-radius: 6px;
		cursor: pointer;
		font-size: 0.9rem;
		transition: all 0.3s ease;
		white-space: nowrap;
	}

	.add-tag-btn:hover:not(:disabled) {
		background: var(--primary-hover);
		transform: translateY(-1px);
	}

	.add-tag-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
		transform: none;
	}

	.tags-display {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		margin: 0.5rem 0;
	}

	.tag-chip {
		background: rgba(59, 130, 246, 0.2);
		color: #3b82f6;
		padding: 0.25rem 0.75rem;
		border-radius: 20px;
		font-size: 0.8rem;
		font-weight: 500;
	}

	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: 1rem;
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

	.btn:hover:not(:disabled) {
		transform: translateY(-1px);
	}

	.btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.animate-spin {
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
	}

	@media (max-width: 768px) {
		.modal-content {
			width: 98%;
			padding: 1.5rem;
			margin: 0.5rem;
		}
	}
</style>
