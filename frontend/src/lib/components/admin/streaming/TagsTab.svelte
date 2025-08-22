<script lang="ts">
	import { toastStore } from '$lib/stores/toast';
	
	// Props
	export let tags: any[] = [];
	export let categories: any[] = [];
	export let loading: boolean = false;
	export let newTag: string = '';
	
	// Events
	import { createEventDispatcher } from 'svelte';
	const dispatch = createEventDispatcher();
	
	// Local state
	let tagsSortField: 'status' | 'word' | 'frequency' | null = null;
	let tagsSortDirection: 'asc' | 'desc' = 'asc';
	
	// Event handlers
	function handleAddTag() {
		dispatch('addTag');
	}
	
	function handleDeleteTag(tag: any) {
		dispatch('deleteTag', tag);
	}
	
	function handleCategoryChange(event: Event, tagId: any) {
		const target = event.target as HTMLSelectElement;
		if (target) {
			dispatch('categoryChange', { tagId, categoryId: target.value });
		}
	}
	
	function handleToggleStatus(tag: any) {
		dispatch('toggleStatus', tag);
	}
	
	function handleAddToExclusions(tag: any) {
		dispatch('addToExclusions', tag);
	}
	
	// Sorting function
	function sortTags(field: 'status' | 'word' | 'frequency') {
		if (tagsSortField === field) {
			tagsSortDirection = tagsSortDirection === 'asc' ? 'desc' : 'asc';
		} else {
			tagsSortField = field;
			tagsSortDirection = 'asc';
		}
		
		// Apply sorting
		tags = [...tags].sort((a, b) => {
			let aVal: any, bVal: any;
			
			switch (field) {
				case 'status':
					aVal = a.active_tag ? 1 : 0;
					bVal = b.active_tag ? 1 : 0;
					break;
				case 'word':
					aVal = a.word.toLowerCase();
					bVal = b.word.toLowerCase();
					break;
				case 'frequency':
					aVal = a.frequency || 0;
					bVal = b.frequency || 0;
					break;
				default:
					return 0;
			}
			
			if (tagsSortDirection === 'asc') {
				return aVal > bVal ? 1 : -1;
			} else {
				return aVal < bVal ? 1 : -1;
			}
		});
	}
</script>

<div class="tab-content">
	<div class="section-header">
		<h2>Streaming Video Tags</h2>
		<div class="add-tag-form">
			<input 
				type="text" 
				bind:value={newTag} 
				placeholder="Add new streaming tag..."
				on:keydown={(e) => e.key === 'Enter' && handleAddTag()}
			/>
			<button on:click={handleAddTag} class="add-button">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<line x1="12" y1="5" x2="12" y2="19"></line>
					<line x1="5" y1="12" x2="19" y2="12"></line>
				</svg>
				Add to Streaming
			</button>
		</div>
	</div>

	<div class="tags-info">
		<p><strong>Tag Actions:</strong></p>
		<ul>
			<li><span class="info-icon">🟢/🔴</span> <strong>Checkbox:</strong> Toggle tag active status (controls if tag is used for new videos)</li>
			<li><span class="info-icon">🚫</span> <strong>Exclude:</strong> Add word to exclusions and remove tag completely</li>
			<li><span class="info-icon">🗑️</span> <strong>Delete:</strong> Remove tag from streaming subsite</li>
		</ul>
	</div>

	{#if loading}
		<div class="loading">Loading streaming tags...</div>
	{:else if tags.length === 0}
		<div class="empty-state">
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"></path>
				<line x1="7" y1="7" x2="7.01" y2="7"></line>
			</svg>
			<p>No streaming-specific tags found</p>
			<p>Add tags to get started</p>
		</div>
	{:else}
		<div class="exclusions-table-container">
			<table class="exclusions-table">
				<thead>
					<tr>
						<th class="exclusion-head sortable" on:click={() => sortTags('status')}>
							Status
							<span class="sort-indicator">
								{#if tagsSortField === 'status'}
									{tagsSortDirection === 'asc' ? '↑' : '↓'}
								{:else}
									↕
								{/if}
							</span>
						</th>
						<th class="exclusion-head sortable" on:click={() => sortTags('word')}>
							Tag Word
							<span class="sort-indicator">
								{#if tagsSortField === 'word'}
									{tagsSortDirection === 'asc' ? '↑' : '↓'}
								{:else}
									↕
								{/if}
							</span>
						</th>
						<th class="exclusion-head sortable" on:click={() => sortTags('frequency')}>
							Frequency
							<span class="sort-indicator">
								{#if tagsSortField === 'frequency'}
									{tagsSortDirection === 'asc' ? '↑' : '↓'}
								{:else}
									↕
								{/if}
							</span>
						</th>
						<th class="exclusion-head">Category</th>
						<th class="exclusion-head">Actions</th>
					</tr>
				</thead>
				<tbody>
					{#each tags as tag}
						<tr class="exclusion-row">
							<td class="exclusion-status">
								<label class="status-toggle">
									<input 
										type="checkbox" 
										checked={tag.active_tag} 
										on:change={() => handleToggleStatus(tag)}
										class="status-checkbox"
										aria-label="Toggle tag active status"
									/>
									<span class="status-indicator" class:active={tag.active_tag}>
										{tag.active_tag ? '🟢 Active' : '🔴 Inactive'}
									</span>
								</label>
							</td>
							<td class="exclusion-word">
								<span class="word-text">{tag.word}</span>
							</td>
							<td class="exclusion-status">
								<span class="tag-frequency-display">
									{tag.frequency || 0} uses
								</span>
							</td>
							<td class="exclusion-status">
								{#if categories.length > 0}
									<select 
										value={tag.category_id || ''} 
										on:change={(e) => handleCategoryChange(e, tag.id)}
										class="category-select"
									>
										<option value="">No Category</option>
										{#each categories as category}
											<option value={category.id}>{category.name}</option>
										{/each}
									</select>
								{/if}
							</td>
							<td class="exclusion-actions">
								<button 
									on:click={() => handleAddToExclusions(tag)} 
									class="remove-button"
									title="Add to exclusions and remove tag"
								>
									🚫
								</button>
								<button 
									on:click={() => handleDeleteTag(tag)} 
									class="remove-button"
									title="Delete tag"
								>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<polyline points="3,6 5,6 21,6"></polyline>
										<path d="M19,6v14a2,2,0,0,1-2,2H7a2,2,0,0,1-2-2V6m3,0V4a2,2,0,0,1,2-2h4a2,2,0,0,1,2,2V6"></path>
									</svg>
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>

<style>
	/* Copy relevant styles from parent component */
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
	
	.add-tag-form {
		display: flex;
		gap: 0.5rem;
		align-items: center;
		flex-wrap: wrap;
	}
	
	.add-tag-form input {
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
	
	.tags-info {
		background: var(--bg-hover);
		border: 1px solid var(--border-color);
		border-radius: 8px;
		padding: 1rem;
		margin-bottom: 1.5rem;
		font-size: 0.9rem;
	}
	
	.tags-info p {
		margin: 0 0 0.75rem 0;
		color: var(--text-primary);
	}
	
	.tags-info ul {
		margin: 0;
		padding-left: 1.5rem;
		color: var(--text-secondary);
	}
	
	.tags-info li {
		margin-bottom: 0.5rem;
	}
	
	.info-icon {
		margin-right: 0.5rem;
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
	
	/* Table styles */
	.exclusions-table-container {
		overflow-x: auto;
		border-radius: 8px;
		border: 1px solid var(--border-color);
	}
	
	.exclusions-table {
		width: 100%;
		border-collapse: collapse;
		background: var(--bg-tertiary);
	}
	
	.exclusions-table th {
		background: var(--bg-hover);
		color: var(--text-primary);
		font-weight: 600;
		padding: 1rem;
		text-align: left;
		border-bottom: 2px solid var(--border-color);
		font-size: 0.9rem;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}
	
	.exclusions-table td {
		padding: 1rem;
		border-bottom: 1px solid var(--border-color);
		vertical-align: middle;
	}
	
	thead {
		background-color: var(--bg-quinary);
	}
	
	.exclusion-head {
		color: var(--text-inverse) !important;
		font-weight: 600;
	}
	
	.exclusion-row {
		border-bottom: 5px solid var(--bg-primary);
		color: var(--text-tertiary);
	}
	
	.exclusion-row:hover {
		background: var(--bg-hover);
		transition: background-color 0.2s ease;
	}
	
	.exclusion-word .word-text {
		font-weight: 600;
		color: var(--text-primary);
		font-family: 'Courier New', monospace;
		background: rgba(59, 130, 246, 0.1);
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		border: 1px solid rgba(59, 130, 246, 0.2);
	}
	
	.status-toggle {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		cursor: pointer;
	}
	
	.status-checkbox {
		width: 18px;
		height: 18px;
		cursor: pointer;
		accent-color: var(--primary-color);
	}
	
	.status-indicator {
		padding: 0.25rem 0.75rem;
		border-radius: 12px;
		font-size: 1.25rem;
		font-weight: 500;
		transition: all 0.2s ease;
	}
	
	.status-indicator.active {
		background: var(--error-bg);
		color: var(--error-text);
	}
	
	.status-indicator:not(.active) {
		background: var(--success-bg);
		color: var(--success-text);
	}
	
	.remove-button {
		background: var(--error-bg);
		color: var(--text-secondary);
		border: none;
		padding: 0.5rem;
		border-radius: 6px;
		cursor: pointer;
		transition: all 0.2s ease;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	
	.remove-button:hover {
		background: var(--error-text);
		color: red;
		transform: scale(1.05);
	}
	
	.remove-button svg {
		width: 16px;
		height: 16px;
	}
	
	.tag-frequency-display {
		font-size: 0.9rem;
		color: var(--text-secondary);
		background: rgba(59, 130, 246, 0.1);
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		border: 1px solid rgba(59, 130, 246, 0.2);
		font-weight: 500;
	}
	
	.category-select {
		padding: 0.25rem 0.5rem;
		border: 1px solid var(--border-color);
		border-radius: 4px;
		background: var(--bg-primary);
		color: var(--text-primary);
		font-size: 0.9rem;
		min-width: 120px;
	}
	
	/* Sortable styles */
	.sortable {
		cursor: pointer;
		user-select: none;
		transition: background-color 0.2s ease;
	}
	
	.sortable:hover {
		background: var(--bg-hover);
	}
	
	.sort-indicator {
		margin-left: 0.5rem;
		font-size: 0.8rem;
		opacity: 0.7;
	}
	
	.sortable:hover .sort-indicator {
		opacity: 1;
	}
</style>
