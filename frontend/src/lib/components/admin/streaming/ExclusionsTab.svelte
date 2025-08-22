<script lang="ts">
	import { toastStore } from '$lib/stores/toast';
	
	// Props using Svelte 5 $props rune with callback props
	let { 
		articleExclusions = [], 
		newExclusion = $bindable(''), 
		exclusionsLoading = false,
		// Callback props instead of events
		onAddExclusion,
		onToggleExclusion,
		onRemoveExclusion
	} = $props();
	
	// Local state using $state rune
	let exclusionsSortField = $state<'word' | 'status' | null>(null);
	let exclusionsSortDirection = $state<'asc' | 'desc'>('asc');
	
	// Event handlers - now call callback props directly
	function handleAddExclusion() {
		onAddExclusion?.();
	}
	
	function handleToggleExclusion(exclusion: any) {
		onToggleExclusion?.(exclusion);
	}
	
	function handleRemoveExclusion(exclusion: any) {
		onRemoveExclusion?.(exclusion);
	}
	
	// Sorting function
	function sortExclusions(field: 'word' | 'status') {
		if (exclusionsSortField === field) {
			exclusionsSortDirection = exclusionsSortDirection === 'asc' ? 'desc' : 'asc';
		} else {
			exclusionsSortField = field;
			exclusionsSortDirection = 'asc';
		}
		
		// Apply sorting
		articleExclusions = [...articleExclusions].sort((a, b) => {
			let aVal: any, bVal: any;
			
			switch (field) {
				case 'word':
					// Alphanumeric sorting: numbers first, then letters
					aVal = a.Word.toLowerCase();
					bVal = b.Word.toLowerCase();
					
					// Check if both are numbers
					const aNum = parseFloat(aVal);
					const bNum = parseFloat(bVal);
					if (!isNaN(aNum) && !isNaN(bNum)) {
						return exclusionsSortDirection === 'asc' ? aNum - bNum : bNum - aNum;
					}
					// Check if only a is a number
					if (!isNaN(aNum)) return -1;
					// Check if only b is a number
					if (!isNaN(bNum)) return 1;
					// Both are strings, sort alphabetically
					break;
				case 'status':
					aVal = a.Excluded ? 1 : 0;
					bVal = b.Excluded ? 1 : 0;
					break;
				default:
					return 0;
			}
			
			if (exclusionsSortDirection === 'asc') {
				return aVal > bVal ? 1 : -1;
			} else {
				return aVal < bVal ? 1 : -1;
			}
		});
	}
</script>

<div class="tab-content">
	<div class="section-header">
		<h2>Article Exclusion List</h2>
		<div class="add-exclusion-form">
			<input 
				type="text" 
				bind:value={newExclusion} 
				placeholder="Enter article URL or pattern to exclude..."
				on:keydown={(e) => e.key === 'Enter' && handleAddExclusion()}
			/>
			<button on:click={handleAddExclusion} class="add-button">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<line x1="12" y1="5" x2="12" y2="19"></line>
					<line x1="5" y1="12" x2="19" y2="12"></line>
				</svg>
				Add Exclusion
			</button>
		</div>
	</div>

	{#if exclusionsLoading}
		<div class="loading">Loading article exclusions...</div>
	{:else if articleExclusions.length === 0}
		<div class="empty-state">
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M9 12l2 2 4-4"></path>
				<path d="M21 12c-1 0-2-1-2-2s1-2 2-2 2 1 2 2-1 2-2 2z"></path>
				<path d="M3 12c1 0 2-1 2-2s-1-2-2-2-2 1-2 2 1 2 2 2z"></path>
			</svg>
			<p>No article exclusions found</p>
			<p>Add exclusions to prevent certain articles from being processed</p>
		</div>
	{:else}
		<div class="exclusions-table-container">
			<table class="exclusions-table">
				<thead>
					<tr>
						<th class="exclusion-head sortable" on:click={() => sortExclusions('word')}>
							Word/Pattern
							<span class="sort-indicator">
								{#if exclusionsSortField === 'word'}
									{exclusionsSortDirection === 'asc' ? '↑' : '↓'}
								{:else}
									↕
								{/if}
							</span>
						</th>
						<th class="exclusion-head sortable" on:click={() => sortExclusions('status')}>
							Status
							<span class="sort-indicator">
								{#if exclusionsSortField === 'status'}
									{exclusionsSortDirection === 'asc' ? '↑' : '↓'}
								{:else}
									↕
								{/if}
							</span>
						</th>
						<th class="exclusion-head">Actions</th>
					</tr>
				</thead>
				<tbody>
					{#each articleExclusions as exclusion}
						<tr class="exclusion-row">
							<td class="exclusion-word">
								<span class="word-text">{exclusion.Word}</span>
							</td>
							<td class="exclusion-status">
								<label class="status-toggle">
									<input 
										type="checkbox" 
										checked={exclusion.Excluded} 
										on:change={() => handleToggleExclusion(exclusion)}
										class="status-checkbox"
										aria-label="Toggle exclusion status"
									/>
									
									<span class="status-indicator" class:active={exclusion.Excluded}>
										{exclusion.Excluded ? '🚫 Excluded' : '✅ Included'}
									</span>
								</label>
							</td>
							<td class="exclusion-actions">
								<button 
									on:click={() => handleRemoveExclusion(exclusion)} 
									class="remove-button"
									title="Remove exclusion"
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
	
	.add-exclusion-form {
		display: flex;
		gap: 0.5rem;
		align-items: center;
		flex-wrap: wrap;
	}
	
	.add-exclusion-form input {
		font-size: 1.1rem;
		padding: 0.5rem 1rem;
		border: 1px solid var(--border-color);
		border-radius: 11px;
		min-height: 65px;
		color: var(--text-primary);
		min-width: 300px;
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
