<script lang="ts">
	import { toastStore } from '$lib/stores/toast';
	
	// Props using Svelte 5 $props rune with callback props
	let { 
		categories = [], 
		tags = [], 
		loading = false, 
		newCategory = $bindable(), 
		newCategoryColor = $bindable('#3B82F6'),
		// Callback props instead of events
		onAddCategory,
		onDeleteCategory,
		onUpdateCategory,
		onBatchTagChanges
	} = $props();
	
	// Modal state using $state rune for reactivity
	let showModal = $state(false);
	let selectedCategory = $state<any>(null);
	let searchTerm = $state('');
	let editingCategory = $state(false);
	let editCategoryName = $state('');
	let editCategoryColor = $state('');
	let editCategoryDescription = $state('');

	// Batch tag changes tracking
	let pendingTagChanges = $state<Array<{tagId: number, categoryId: number | null, action: 'add' | 'remove'}>>([]);
	let hasUnsavedChanges = $state(false);
	
	// Replace the complex tagLookup and affiliatedTags with this simple version:
	const affiliatedTags = $derived(() => {
		if (!selectedCategory?.tag_ids) return [];
		
		// Simple filter - find tags where the tag.id is in the category's tag_ids array
		return tags.filter(tag => selectedCategory.tag_ids.includes(tag.id));
	});

	const filteredTags = $derived(() => {
		if (!selectedCategory) return [];
		// Show tags that are NOT in the category's tag_ids
		return tags.filter(tag => !selectedCategory.tag_ids || !selectedCategory.tag_ids.includes(tag.id));
	});
	
	const searchFilteredTags = $derived(() => {
    if (!searchTerm.trim()) return filteredTags();
    return filteredTags().filter((tag: any) => 
        tag.word.toLowerCase().includes(searchTerm.toLowerCase())
    );
});
	
	// Event handlers
	function handleAddCategory() {
		onAddCategory?.();
	}
	
	function handleDeleteCategory(category: any) {
		onDeleteCategory?.(category);
	}
	
	function handleCategoryClick(category: any) {
		selectedCategory = category;
		searchTerm = '';
		editCategoryName = category.name;
		editCategoryColor = category.color;
		editCategoryDescription = category.description || '';
		pendingTagChanges = []; // Reset pending changes
		hasUnsavedChanges = false;
		showModal = true;
	}
	
	function closeModal() {
		if (hasUnsavedChanges) {
			if (confirm('You have unsaved tag changes. Are you sure you want to close?')) {
				resetModal();
			}
		} else {
			resetModal();
		}
	}

	function resetModal() {
		showModal = false;
		selectedCategory = null;
		searchTerm = '';
		editingCategory = false;
		editCategoryName = '';
		editCategoryColor = '';
		editCategoryDescription = '';
		pendingTagChanges = [];
		hasUnsavedChanges = false;
	}
	
	function toggleEditMode() {
		editingCategory = !editingCategory;
		if (editingCategory) {
			editCategoryName = selectedCategory.name;
			editCategoryColor = selectedCategory.color;
			editCategoryDescription = selectedCategory.description || '';
		}
	}
	
	function saveCategoryChanges() {
		onUpdateCategory?.({
			id: selectedCategory.id,
			name: editCategoryName,
			color: editCategoryColor,
			description: editCategoryDescription
		});
		editingCategory = false;
	}
	
	function handleTagSelect(tag: any) {
		// Add to pending changes instead of immediate dispatch
		const change = {
			tagId: tag.id,
			categoryId: selectedCategory.id,
			action: 'add' as const
		};
		
		// Check if this change is already pending
		const existingIndex = pendingTagChanges.findIndex(c => c.tagId === tag.id);
		if (existingIndex >= 0) {
			pendingTagChanges[existingIndex] = change;
		} else {
			pendingTagChanges = [...pendingTagChanges, change];
		}
		
		hasUnsavedChanges = true;
		toastStore.success(`"${tag.word}" added to pending changes`);
	}
	
	// Improve error handling for tag removal
	async function removeTagFromCategory(tag: any) {
		try {
			// Add to pending changes instead of immediate dispatch
			const change = {
				tagId: tag.id,
				categoryId: null,
				action: 'remove' as const
			};
			
			// Check if this change is already pending
			const existingIndex = pendingTagChanges.findIndex(c => c.tagId === tag.id);
			if (existingIndex >= 0) {
				pendingTagChanges[existingIndex] = change;
			} else {
				pendingTagChanges = [...pendingTagChanges, change];
			}
			
			hasUnsavedChanges = true;
			toastStore.success(`"${tag.word}" marked for removal`);
		} catch (error) {
			console.error('Error marking tag for removal:', error);
			toastStore.error(`Failed to mark "${tag.word}" for removal`);
		}
	}

	// New batch save function
	async function saveTagChanges() {
		try {
			// Create the changes array for the parent component
			const changes = pendingTagChanges.map(change => ({
				tagId: change.tagId,
				categoryId: change.categoryId,
				action: change.action
			}));
			
			// Dispatch the batch update event
			onBatchTagChanges?.(changes);
			
			// Clear pending changes
			pendingTagChanges = [];
			hasUnsavedChanges = false;
			
			toastStore.success(`Successfully saved ${changes.length} tag changes`);
		} catch (error) {
			toastStore.error('Failed to save tag changes');
			console.error('Error saving tag changes:', error);
		}
	}

	function cancelTagChanges() {
		pendingTagChanges = [];
		hasUnsavedChanges = false;
		toastStore.info('Tag changes cancelled');
	}

	// Add helper function to check pending state
	function getTagPendingState(tagId: number) {
		const pendingChange = pendingTagChanges.find(change => change.tagId === tagId);
		if (!pendingChange) return null;
		return pendingChange.action; // 'add' or 'remove'
	}
</script>

<div class="tab-content">
	<div class="section-header">
		<h2>Streaming Tag Categories</h2>
		<div class="add-category-form">
			<input 
				type="text" 
				bind:value={newCategory} 
				placeholder="Streaming category name..."
				on:keydown={(e) => e.key === 'Enter' && handleAddCategory()}
			/>
			<input 
				type="color" 
				bind:value={newCategoryColor}
				class="color-picker"
			/>
			<button on:click={handleAddCategory} class="add-button">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<line x1="12" y1="5" x2="12" y2="19"></line>
					<line x1="5" y1="12" x2="19" y2="12"></line>
				</svg>
				Add to Streaming
			</button>
		</div>
	</div>

	{#if loading}
		<div class="loading">Loading streaming categories...</div>
	{:else if categories.length === 0}
		<div class="empty-state">
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M3 3h18v18H3zM8 12h8M12 8v8"></path>
			</svg>
			<p>No streaming-specific categories found</p>
			<p>Create categories or import from the global system</p>
		</div>
	{:else}
		<div class="categories-grid">
			{#each categories as category}
				<div class="category-card" style="--category-color: {category.color}" on:click={() => handleCategoryClick(category)}>
					<div class="category-header">
						<div class="category-color" style="background-color: {category.color}"></div>
						<span class="category-name">{category.name}</span>
						<span class="category-count">
							{category.tag_ids.length} tags
						</span>
					</div>
					<div class="category-description">
						{category.description || 'Streaming-specific category'}
					</div>
					<div class="category-actions">
						<button on:click|stopPropagation={() => handleDeleteCategory(category)} class="delete-button">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polyline points="3,6 5,6 21,6"></polyline>
								<path d="M19,6v14a2,2,0,0,1-2,2H7a2,2,0,0,1-2-2V6m3,0V4a2,2,0,0,1,2-2h4a2,2,0,0,1,2,2V6"></path>
							</svg>
						</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Modal for adding tags to category -->
{#if showModal && selectedCategory}
	<div class="modal-overlay" on:click={closeModal}>
		<div class="modal-content" on:click|stopPropagation>
			<div class="modal-header">
				<h3>
					{#if editingCategory}
						Edit Category: "{editCategoryName}"
					{:else}
						Manage "{selectedCategory.name}" Category
					{/if}
				</h3>
				<button class="close-button" on:click={closeModal}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"></line>
						<line x1="6" y1="6" x2="18" y2="18"></line>
					</svg>
				</button>
			</div>
			
			

			<div class="modal-body">
				<!-- Category Details Section -->
				<div class="category-details-section">
					<div class="section-header">
						<h4>Category Details</h4>
						{#if !editingCategory}
							<button class="edit-button" on:click={toggleEditMode}>
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
									<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
								</svg>
								Edit
							</button>
						{:else}
							<div class="edit-actions">
								<button class="save-button" on:click={saveCategoryChanges}>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<polyline points="20,6 9,17 4,12"></polyline>
									</svg>
									Save
								</button>
								<button class="cancel-button" on:click={toggleEditMode}>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<line x1="18" y1="6" x2="6" y2="18"></line>
										<line x1="6" y1="6" x2="18" y2="18"></line>
									</svg>
									Cancel
								</button>
							</div>
						{/if}
					</div>
					
					{#if editingCategory}
						<div class="edit-fields">
							<div class="field-group">
								<label for="category-name">Name:</label>
								<input
									id="category-name"
									type="text"
									bind:value={editCategoryName}
									placeholder="Category name..."
									class="edit-input"
								/>
							</div>
							<div class="field-group">
								<label for="category-color">Color:</label>
								<input
									id="category-color"
									type="color"
									bind:value={editCategoryColor}
									class="color-picker-edit"
								/>
							</div>
							<div class="field-group">
								<label for="category-description">Description:</label>
								<textarea
									id="category-description"
									bind:value={editCategoryDescription}
									placeholder="Category description..."
									class="edit-textarea"
									rows="3"
								></textarea>
							</div>
						</div>
					{:else}
						<div class="category-info">
							<div class="info-row">
								<span class="info-label">Name:</span>
								<span class="info-value">{selectedCategory.name}</span>
							</div>
							<div class="info-row">
								<span class="info-label">Color:</span>
								<div class="color-display" style="background-color: {selectedCategory.color}"></div>
							</div>
							<div class="info-row">
								<span class="info-label">Description:</span>
								<span class="info-value">{selectedCategory.description || 'No description'}</span>
							</div>
						</div>
					{/if}
				</div>
				
				<!-- Affiliated Tags Section -->
				<div class="affiliated-tags-section">
					<div class="section-header">
						<h4>Affiliated Tags ({affiliatedTags().length})</h4>
					</div>
					
					{#if affiliatedTags().length === 0}
						<div class="no-affiliated-tags">
							<p>No tags are currently assigned to this category</p>
						</div>
					{:else}
						<div class="affiliated-tags-list">
							{#each affiliatedTags() as tag}
								<div class="affiliated-tag-item">
									<span class="tag-word">{tag.word}</span>
									<span class="tag-frequency">({tag.frequency || 0})</span>
									<button class="remove-tag-button" on:click={() => removeTagFromCategory(tag)}>
										❌
									</button>
								</div>
							{/each}
						</div>
					{/if}
				</div>

				<!-- Pending Changes Section - Only shows when there are changes -->
				{#if hasUnsavedChanges && pendingTagChanges.length > 0}
					<div class="pending-changes-section">
						<div class="section-header">
							<h4>Pending Changes ({pendingTagChanges.length})</h4>
						</div>
						
						<div class="pending-changes-list">
							{#each pendingTagChanges as change}
								{@const tag = tags.find(t => t.id === change.tagId)}
								{#if tag}
									<div class="pending-change-item" class:adding={change.action === 'add'} class:removing={change.action === 'remove'}>
										<span class="change-action">{change.action === 'add' ? '➕' : '➖'}</span>
										<span class="tag-word">{tag.word}</span>
										<span class="tag-frequency">({tag.frequency || 0})</span>
										<!--<span class="change-description">
											{change.action === 'add' ? 'Will be added to category' : 'Will be removed from category'}
										</span>-->
									</div>
								{/if}
							{/each}
						</div>
					</div>
				{/if}

				<div class="modal-actions" style="display: flex; gap: 1rem; justify-content: flex-end; padding: 1rem; border-bottom: 1px solid var(--border-color);">
					{#if hasUnsavedChanges}
						<button class="cancel-button" on:click={cancelTagChanges}>
							Cancel Changes
						</button>
						<button class="save-button" on:click={saveTagChanges}>
							Save Tag Changes ({pendingTagChanges.length})
						</button>
					{/if}
				</div>
				
				<!-- Add Tags Section -->
				<div class="add-tags-section">
					<div class="section-header">
						<h4>Add New Tags</h4>
					</div>
					
					<div class="search-container">
						<input
							type="text"
							bind:value={searchTerm}
							placeholder="Search available tags..."
							class="search-input"
						/>
					</div>
					
					<div class="tags-list">
						{#if searchFilteredTags().length === 0}
							<div class="no-tags">
								{#if searchTerm.trim()}
									<p>No tags found matching "{searchTerm}"</p>
								{:else}
									<p>No available tags to add</p>
								{/if}
							</div>
						{:else}
							{#each searchFilteredTags() as tag}
								<div class="tag-item" on:click={() => handleTagSelect(tag)}>
									<span class="tag-word">{tag.word}</span>
									<span class="tag-frequency">({tag.frequency || 0})</span>
								</div>
							{/each}
						{/if}
					</div>
				</div>

				
			</div>
		</div>
	</div>
{/if}

<style>
	.tab-content {
		background: var(--bg-glass);
		border-radius: 12px;
		padding: 2rem;
		backdrop-filter: blur(80px);
		border: 1px solid var(--border-color);
	}
	
	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 2rem;
		flex-wrap: wrap;
		gap: 1rem;
	}
	
	.section-header h2 {
		margin: 0;
		font-size: 1.5rem;
		color: var(--text-primary);
	}
	
	.add-category-form {
		display: flex;
		gap: 0.5rem;
		align-items: center;
		flex-wrap: wrap;
	}
	
	.add-category-form input[type="text"] {
		font-size: 1.1rem;
		padding: 0.5rem 1rem;
		border: 1px solid var(--border-color);
		border-radius: 11px;
		min-height: 65px;
		color: var(--text-primary);
		min-width: 200px;
		background: var(--bg-primary);
		box-shadow: inset 5px 5px 10px var(--bg-senary),
					inset -5px -5px 10px var(--bg-secondary);
	}
	
	.color-picker {
		width: 50px;
		height: 40px;
		border: none;
		border-radius: 6px;
		cursor: pointer;
	}
	
	.add-button {
		background: var(--bg-glass-dark);
		color: var(--text-secondary);
		border: none;
		padding: 0.5rem 1rem;
		border-radius: 6px;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		transition: all 0.3s ease;
		white-space: nowrap;
		min-height: 65px;
		box-shadow: 0 0 0 var(--bg-senary),
					0 0 0 var(--bg-secondary);
	}
	
	.add-button:hover {
		background: var(--bg-glass);
		transform: translateY(-1px);
		box-shadow: 5px 5px 10px var(--bg-senary),
					-5px -5px 10px var(--bg-secondary);
	}
	
	.add-button:active {
		transform: translateY(1px);
		border-radius: 11px;
		background: var(--bg-glass-dark);
		box-shadow: inset 5px 5px 10px var(--bg-senary),
					inset -5px -5px 10px var(--bg-secondary);
	}
	
	.loading {
		text-align: center;
		padding: 2rem;
		color: var(--text-secondary);
	}
	
	.empty-state {
		text-align: center;
		padding: 3rem;
		color: var(--text-secondary);
	}
	
	.empty-state svg {
		width: 64px;
		height: 64px;
		margin-bottom: 1rem;
		opacity: 0.5;
	}
	
	.categories-grid {
		display: grid;
		gap: 1rem;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
	}
	
	.category-card {
		background: var(--bg-primary);
		border: 1px solid var(--border-color);
		border-radius: 8px;
		padding: 1rem;
		transition: all 0.3s ease;
		cursor: pointer;
	}
	
	.category-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
	}
	
	.category-header {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 1rem;
	}
	
	.category-color {
		width: 16px;
		height: 16px;
		border-radius: 50%;
		border: 2px solid var(--border-color);
	}
	
	.category-name {
		font-weight: 600;
		color: var(--text-primary);
		flex: 1;
	}
	
	.category-count {
		font-size: 0.9rem;
		color: var(--text-secondary);
	}
	
	.category-description {
		color: var(--text-secondary);
		font-size: 0.9rem;
		margin-bottom: 1rem;
		font-style: italic;
	}
	
	.category-actions {
		display: flex;
		justify-content: flex-end;
	}
	
	.delete-button {
		background: var(--error-bg);
		color: var(--error-text);
		border: none;
		padding: 0.25rem;
		border-radius: 4px;
		cursor: pointer;
		transition: all 0.3s ease;
	}
	
	.delete-button:hover {
		background: var(--error-text);
		color: white;
	}

	/* Modal Styles */
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
		backdrop-filter: blur(4px);
	}

	.modal-content {
		background: var(--bg-primary);
		border: 1px solid var(--border-color);
		border-radius: 12px;
		width: 90%;
		max-width: 1500px;
		max-height: 82vh;
		overflow: scroll;
		box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);

	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1.5rem;
		border-bottom: 1px solid var(--border-color);
	}

	.modal-header h3 {
		margin: 0;
		color: var(--text-primary);
		font-size: 1.25rem;
	}

	.close-button {
		background: none;
		border: none;
		color: var(--text-secondary);
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 4px;
		transition: all 0.2s ease;
	}

	.close-button:hover {
		background: var(--bg-secondary);
		color: var(--text-primary);
	}

	.close-button svg {
		width: 20px;
		height: 20px;
	}

	.modal-body {
		padding: 1.5rem;
		height: 100%;
		box-sizing: border-box;
	}

	.search-container {
		margin-bottom: 1.5rem;
	}

	.search-input {
		width: 100%;
		padding: 0.75rem 1rem;
		border: 1px solid var(--border-color);
		border-radius: 8px;
		font-size: 1rem;
		background: var(--bg-secondary);
		color: var(--text-primary);
		box-sizing: border-box;
	}

	.search-input:focus {
		outline: none;
		border-color: var(--accent-color);
		box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
	}

	.tags-list {
		max-height: 300px;
		overflow-y: auto;
	}

	.tag-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.75rem 1rem;
		border: 1px solid var(--border-color);
		border-radius: 6px;
		margin-bottom: 0.5rem;
		cursor: pointer;
		transition: all 0.2s ease;
		color: var(--text-primary);
	}

	.tag-item:hover {
		background: var(--bg-secondary);
		border-color: var(--accent-color);
		transform: translateX(4px);
	}

	.tag-word {
		font-weight: 500;
		color: var(--text-primary);
		font-size: 1.2rem;
		font-weight: 600;
		
	}

	.tag-frequency {
		font-size: 0.9rem;
		color: var(--text-secondary);
	}

	.no-tags {
		text-align: center;
		padding: 2rem;
		color: var(--text-secondary);
	}

	.no-tags p {
		margin: 0;
	}

	/* New styles for category details and affiliated tags */
    .add-tags-section, .tags-list {
		overflow-x: hidden;
		scroll-behavior: smooth;
	}

	.category-details-section, .affiliated-tags-section, .add-tags-section {
		background: var(--bg-secondary);
		border: 1px solid var(--border-color);
		border-radius: 8px;
		padding: 1.5rem;
		margin-bottom: 1.5rem;
		box-shadow: inset 0 0 10px rgba(0, 0, 0, 0.05);
	}

	.category-details-section .section-header, .affiliated-tags-section .section-header, .add-tags-section .section-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1.5rem;
	}

	.category-details-section .section-header h4, .affiliated-tags-section .section-header h4, .add-tags-section .section-header h4 {
		margin: 0;
		color: var(--text-primary);
		font-size: 1.1rem;
	}

	.edit-button, .save-button, .cancel-button {
		background: var(--bg-glass-dark);
		color: var(--text-secondary);
		border: none;
		padding: 0.5rem 1rem;
		border-radius: 6px;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		transition: all 0.3s ease;
		white-space: nowrap;
		box-shadow: 0 0 0 var(--bg-senary),
					0 0 0 var(--bg-secondary);
	}

	.edit-button:hover, .save-button:hover, .cancel-button:hover {
		background: var(--bg-glass);
		transform: translateY(-1px);
		box-shadow: 5px 5px 10px var(--bg-senary),
					-5px -5px 10px var(--bg-secondary);
	}

	.edit-button:active, .save-button:active, .cancel-button:active {
		transform: translateY(1px);
		border-radius: 11px;
		background: var(--bg-glass-dark);
		box-shadow: inset 5px 5px 10px var(--bg-senary),
					inset -5px -5px 10px var(--bg-secondary);
	}

	.edit-fields {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.field-group {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.field-group label {
		font-size: 0.9rem;
		color: var(--text-secondary);
		min-width: 80px;
	}

	.edit-input, .edit-textarea {
		flex: 1;
		padding: 0.75rem 1rem;
		border: 1px solid var(--border-color);
		border-radius: 8px;
		font-size: 1rem;
		background: var(--bg-primary);
		color: var(--text-primary);
		box-sizing: border-box;
		box-shadow: inset 5px 5px 10px var(--bg-senary),
					inset -5px -5px 10px var(--bg-secondary);
	}

	.edit-input:focus, .edit-textarea:focus {
		outline: none;
		border-color: var(--accent-color);
		box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
	}

	.edit-textarea {
		resize: vertical;
		min-height: 80px;
	}

	.category-info {
		display: flex;
		flex-direction: row;
	    flex-wrap: wrap;
		justify-content: space-around;
		align-items: center;
		margin-bottom: 1.5rem;
	}

	.info-row {
		display: flex;
		flex-direction: row;
		justify-content: center;
		align-items: center;
		margin-bottom: 0.75rem;
	}

	.info-label {
		font-size: 0.9rem;
		color: var(--text-secondary);
		font-weight: 500;
	}

	.info-value {
		font-weight: 600;
		color: var(--text-primary);
		flex: 1;
		text-align: right;
		padding-left: 1rem;
	}

	.color-display {
		width: 120px;
		height: 20px;
		border: 2px solid var(--border-color);
		margin-left: 1rem;
	}

	.affiliated-tags-list {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		margin-top: 0.5rem;
	}

	.affiliated-tag-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		border: 1px solid var(--border-color);
		padding: 0.4rem 0.8rem;
		font-size: 0.9rem;
		color: var(--text-primary);
		border-radius: 15px;
background: linear-gradient(145deg, var(--bg-primary), var(--bg-septenary));
box-shadow:  5px 5px 7px var(--bg-senary),
             -5px -5px 7px var(--bg-secondary);
		transition: all 0.2s ease;
	}

	.affiliated-tag-item:hover {
		transform: translateY(-1px);
	}

	.affiliated-tag-item.adding {
		background: linear-gradient(145deg, var(--bg-adding), var(--bg-septenary));
		border-color: var(--accent-color);
	}

	.affiliated-tag-item.removing {
		background: linear-gradient(145deg, var(--bg-removing), var(--bg-septenary));
		border-color: var(--error-color);
	}

	.remove-tag-button {
		background: none;
		border: none;
		color: var(--text-secondary);
		cursor: pointer;
		padding: 0.25rem;
		border-radius: 4px;
		transition: all 0.2s ease;
	}

	.remove-tag-button:hover {
		background: var(--bg-secondary);
		color: var(--text-primary);
	}

	.remove-tag-button svg {
		width: 16px;
		height: 16px;
	}

	.no-affiliated-tags {
		text-align: center;
		padding: 1rem;
		color: var(--text-secondary);
		font-style: italic;
	}

	.pending-changes-badge {
		background-color: var(--warning-bg);
		color: var(--warning-text);
		padding: 0.4rem 0.8rem;
		border-radius: 6px;
		font-size: 0.9rem;
		font-weight: 600;
		margin-left: 1rem;
	}

	/* Pending Changes Section */
	.pending-changes-section {
		background: var(--bg-tertiary);
		border: 1px solid var(--border-color);
		border-radius: 8px;
		padding: 1.5rem;
		margin-bottom: 1.5rem;
		box-shadow: inset 0 0 10px rgba(0, 0, 0, 0.05);
		border-left: 4px solid var(--warning-color);
	}

	.pending-changes-section .section-header h4 {
		margin: 0;
		color: var(--warning-text);
		font-size: 1.1rem;
	}

	.pending-changes-list {
		display: flex;
		flex-direction: row;
		gap: 0.75rem;
		margin-top: 1rem;
	}

	.pending-change-item {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem 1rem;
		border: 1px solid var(--border-color);
		border-radius: 8px;
		background: var(--bg-primary);
		transition: all 0.2s ease;
	}

	.pending-change-item.adding {
		border-left: 4px solid var(--bg-adding);
		background: linear-gradient(145deg, var(--bg-adding), var(--bg-primary));
	}

	.pending-change-item.removing {
		border-left: 4px solid var(--bg-removing);
		background: linear-gradient(145deg, var(--bg-removing), var(--bg-primary));
	}

	.change-action {
		font-size: 1.2rem;
		font-weight: bold;
	}

	.change-description {
		font-size: 0.9rem;
		color: var(--text-secondary);
		font-style: italic;
		margin-left: auto;
	}
</style>
