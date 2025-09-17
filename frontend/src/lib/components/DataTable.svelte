<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import LoadingSpinner from './LoadingSpinner.svelte';
	
	interface Column {
		key: string;
		label: string;
		sortable?: boolean;
		filterable?: boolean;
		type?: 'text' | 'number' | 'currency' | 'date' | 'boolean' | 'actions';
		width?: string;
		align?: 'left' | 'center' | 'right';
		format?: (value: any) => string;
	}
	
	interface BulkAction {
		id: string;
		label: string;
		icon: string;
		variant?: 'primary' | 'secondary' | 'danger';
		requiresConfirmation?: boolean;
	}
	
	// Props
	export let data: any[] = [];
	export let columns: Column[] = [];
	export let loading = false;
	export let searchable = true;
	export let exportable = true;
	export let selectable = true;
	export let bulkActions: BulkAction[] = [];
	export let emptyMessage = 'No data available';
	export let searchPlaceholder = 'Search...';
	
	// State
	let searchTerm = '';
	let sortColumn = '';
	let sortDirection: 'asc' | 'desc' = 'asc';
	let selectedItems = new Set<any>();
	let selectAll = false;
	
	// Event dispatcher
	const dispatch = createEventDispatcher();
	
	// Computed values
	$: filteredData = filterData(data, searchTerm);
	$: sortedData = sortData(filteredData, sortColumn, sortDirection);
	$: selectedCount = selectedItems.size;
	$: allSelected = selectedCount > 0 && selectedCount === sortedData.length;
	$: someSelected = selectedCount > 0 && selectedCount < sortedData.length;
	
	function filterData(items: any[], search: string): any[] {
		if (!search.trim()) return items;
		
		const searchLower = search.toLowerCase();
		return items.filter(item => 
			Object.values(item).some(value => 
				String(value).toLowerCase().includes(searchLower)
			)
		);
	}
	
	function sortData(items: any[], column: string, direction: 'asc' | 'desc'): any[] {
		if (!column) return items;
		
		return [...items].sort((a, b) => {
			const aVal = a[column];
			const bVal = b[column];
			
			// Handle null/undefined values
			if (aVal == null && bVal == null) return 0;
			if (aVal == null) return direction === 'asc' ? 1 : -1;
			if (bVal == null) return direction === 'asc' ? -1 : 1;
			
			// Compare values
			if (aVal < bVal) return direction === 'asc' ? -1 : 1;
			if (aVal > bVal) return direction === 'asc' ? 1 : -1;
			return 0;
		});
	}
	
	function handleSort(column: Column) {
		if (!column.sortable) return;
		
		if (sortColumn === column.key) {
			sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
		} else {
			sortColumn = column.key;
			sortDirection = 'asc';
		}
	}
	
	function handleSelectAll() {
		if (allSelected || someSelected) {
			selectedItems = new Set();
			selectAll = false;
		} else {
			selectedItems = new Set(sortedData.map(item => item.id));
			selectAll = true;
		}
	}
	
	function handleSelectItem(item: any) {
		const newSelection = new Set(selectedItems);
		if (newSelection.has(item.id)) {
			newSelection.delete(item.id);
		} else {
			newSelection.add(item.id);
		}
		selectedItems = newSelection;
		selectAll = selectedItems.size === sortedData.length;
	}
	
	function handleBulkAction(action: BulkAction) {
		if (selectedItems.size === 0) return;
		
		dispatch('bulkAction', {
			action: action.id,
			items: Array.from(selectedItems),
			requiresConfirmation: action.requiresConfirmation
		});
	}
	
	function handleRowAction(item: any, action: string) {
		dispatch('rowAction', { item, action });
	}
	
	function handleExport() {
		dispatch('export', { data: sortedData, format: 'csv' });
	}
	
	function formatValue(value: any, column: Column): string {
		if (value == null) return '-';
		
		if (column.format) {
			return column.format(value);
		}
		
		switch (column.type) {
			case 'currency':
				return new Intl.NumberFormat('en-US', {
					style: 'currency',
					currency: 'USD'
				}).format(Number(value));
			
			case 'date':
				return new Date(value).toLocaleDateString();
			
			case 'boolean':
				return value ? '✅' : '❌';
			
			case 'number':
				return Number(value).toLocaleString();
			
			default:
				return String(value);
		}
	}
	
	function getSortIcon(column: Column): string {
		if (!column.sortable) return '';
		if (sortColumn !== column.key) return '↕️';
		return sortDirection === 'asc' ? '↑' : '↓';
	}
</script>

<div class="data-table">
	<!-- Table Header Controls -->
	<div class="table-controls">
		<div class="left-controls">
			{#if searchable}
				<div class="search-box">
					<input
						type="text"
						bind:value={searchTerm}
						placeholder={searchPlaceholder}
						class="search-input"
					/>
					<span class="search-icon">🔍</span>
				</div>
			{/if}
			
			{#if selectedCount > 0}
				<div class="selection-info">
					<span class="selected-count">{selectedCount} selected</span>
				</div>
			{/if}
		</div>
		
		<div class="right-controls">
			{#if selectedCount > 0 && bulkActions.length > 0}
				<div class="bulk-actions">
					{#each bulkActions as action}
						<button
							type="button"
							class="bulk-action-btn {action.variant || 'secondary'}"
							onclick={() => handleBulkAction(action)}
						>
							<span class="action-icon">{action.icon}</span>
							{action.label}
						</button>
					{/each}
				</div>
			{/if}
			
			{#if exportable}
				<button
					type="button"
					class="export-btn"
					onclick={handleExport}
				>
					📊 Export
				</button>
			{/if}
		</div>
	</div>
	
	<!-- Data Table -->
	<div class="table-container">
		{#if loading}
			<div class="loading-overlay">
				<LoadingSpinner />
				<p>Loading data...</p>
			</div>
		{:else if sortedData.length === 0}
			<div class="empty-state">
				<div class="empty-icon">📋</div>
				<h3>No Data Found</h3>
				<p>{emptyMessage}</p>
				{#if searchTerm}
					<button
						type="button"
						class="clear-search-btn"
						onclick={() => searchTerm = ''}
					>
						Clear Search
					</button>
				{/if}
			</div>
		{:else}
			<table class="table">
				<thead>
					<tr>
						{#if selectable}
							<th class="select-column">
								<input
									type="checkbox"
									checked={allSelected}
									indeterminate={someSelected}
									onchange={handleSelectAll}
									aria-label="Select all"
								/>
							</th>
						{/if}
						
						{#each columns as column}
							<th
								class="column-header {column.align || 'left'}"
								class:sortable={column.sortable}
								style:width={column.width}
								onclick={() => handleSort(column)}
							>
								<div class="header-content">
									<span class="header-label">{column.label}</span>
									{#if column.sortable}
										<span class="sort-icon">{getSortIcon(column)}</span>
									{/if}
								</div>
							</th>
						{/each}
					</tr>
				</thead>
				
				<tbody>
					{#each sortedData as item, index (`${item.id}-${index}`)}
						<tr class="table-row">
							{#if selectable}
								<td class="select-cell">
									<input
										type="checkbox"
										checked={selectedItems.has(item.id)}
										onchange={() => handleSelectItem(item)}
										aria-label="Select row"
									/>
								</td>
							{/if}
							
							{#each columns as column}
								<td class="table-cell {column.align || 'left'}">
									{#if column.type === 'actions'}
										<div class="action-buttons">
											<button
												type="button"
												class="action-btn"
												onclick={() => handleRowAction(item, 'edit')}
												title="Edit"
											>
												✏️
											</button>
											<button
												type="button"
												class="action-btn"
												onclick={() => handleRowAction(item, 'view')}
												title="View Details"
											>
												👁️
											</button>
										</div>
									{:else}
										{formatValue(item[column.key], column)}
									{/if}
								</td>
							{/each}
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>
	
	<!-- Table Footer -->
	<div class="table-footer">
		<div class="footer-info">
			Showing {sortedData.length} of {data.length} entries
			{#if searchTerm}
				(filtered from {data.length} total)
			{/if}
		</div>
	</div>
</div>

<style>
	.data-table {
		background: white;
		border-radius: 0.75rem;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
		overflow: hidden;
	}
	
	.table-controls {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem 1.5rem;
		border-bottom: 1px solid #e5e7eb;
		background: #f9fafb;
	}
	
	.left-controls, .right-controls {
		display: flex;
		align-items: center;
		gap: 1rem;
	}
	
	.search-box {
		position: relative;
	}
	
	.search-input {
		padding: 0.5rem 2.5rem 0.5rem 1rem;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		min-width: 250px;
	}
	
	.search-input:focus {
		outline: 2px solid #3b82f6;
		outline-offset: -2px;
		border-color: #3b82f6;
	}
	
	.search-icon {
		position: absolute;
		right: 0.75rem;
		top: 50%;
		transform: translateY(-50%);
		color: #6b7280;
		pointer-events: none;
	}
	
	.selected-count {
		font-size: 0.875rem;
		color: #3b82f6;
		font-weight: 500;
		background: #eff6ff;
		padding: 0.25rem 0.75rem;
		border-radius: 9999px;
	}
	
	.bulk-actions {
		display: flex;
		gap: 0.5rem;
	}
	
	.bulk-action-btn {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		border: none;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s ease;
	}
	
	.bulk-action-btn.primary {
		background: #3b82f6;
		color: white;
	}
	
	.bulk-action-btn.primary:hover {
		background: #2563eb;
	}
	
	.bulk-action-btn.secondary {
		background: #f3f4f6;
		color: #374151;
	}
	
	.bulk-action-btn.secondary:hover {
		background: #e5e7eb;
	}
	
	.bulk-action-btn.danger {
		background: #ef4444;
		color: white;
	}
	
	.bulk-action-btn.danger:hover {
		background: #dc2626;
	}
	
	.export-btn {
		padding: 0.5rem 1rem;
		background: #10b981;
		color: white;
		border: none;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: background-color 0.2s ease;
	}
	
	.export-btn:hover {
		background: #059669;
	}
	
	.table-container {
		position: relative;
		overflow-x: auto;
	}
	
	.loading-overlay {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 4rem;
		gap: 1rem;
	}
	
	.loading-overlay p {
		color: #6b7280;
		font-size: 0.875rem;
	}
	
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 4rem 2rem;
		text-align: center;
	}
	
	.empty-icon {
		font-size: 3rem;
		margin-bottom: 1rem;
		opacity: 0.6;
	}
	
	.empty-state h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: #374151;
		margin: 0 0 0.5rem 0;
	}
	
	.empty-state p {
		color: #6b7280;
		font-size: 0.875rem;
		margin: 0 0 1.5rem 0;
	}
	
	.clear-search-btn {
		background: #3b82f6;
		color: white;
		border: none;
		padding: 0.5rem 1rem;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: background-color 0.2s ease;
	}
	
	.clear-search-btn:hover {
		background: #2563eb;
	}
	
	.table {
		width: 100%;
		border-collapse: collapse;
	}
	
	.column-header {
		padding: 0.75rem 1rem;
		text-align: left;
		font-weight: 600;
		color: #374151;
		background: #f9fafb;
		border-bottom: 1px solid #e5e7eb;
		white-space: nowrap;
	}
	
	.column-header.sortable {
		cursor: pointer;
		user-select: none;
	}
	
	.column-header.sortable:hover {
		background: #f3f4f6;
	}
	
	.column-header.center {
		text-align: center;
	}
	
	.column-header.right {
		text-align: right;
	}
	
	.header-content {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	
	.sort-icon {
		font-size: 0.75rem;
		color: #6b7280;
	}
	
	.select-column {
		width: 40px;
		padding: 0.75rem 1rem;
	}
	
	.table-row:hover {
		background: #f9fafb;
	}
	
	.table-cell {
		padding: 0.75rem 1rem;
		border-bottom: 1px solid #f3f4f6;
		vertical-align: middle;
	}
	
	.table-cell.center {
		text-align: center;
	}
	
	.table-cell.right {
		text-align: right;
	}
	
	.select-cell {
		width: 40px;
		text-align: center;
	}
	
	.action-buttons {
		display: flex;
		gap: 0.5rem;
		justify-content: center;
	}
	
	.action-btn {
		padding: 0.25rem 0.5rem;
		background: none;
		border: 1px solid #d1d5db;
		border-radius: 0.25rem;
		cursor: pointer;
		transition: all 0.2s ease;
		font-size: 0.875rem;
	}
	
	.action-btn:hover {
		background: #f3f4f6;
		border-color: #9ca3af;
	}
	
	.table-footer {
		padding: 1rem 1.5rem;
		background: #f9fafb;
		border-top: 1px solid #e5e7eb;
	}
	
	.footer-info {
		font-size: 0.875rem;
		color: #6b7280;
	}
	
	/* Responsive design */
	@media (max-width: 768px) {
		.table-controls {
			flex-direction: column;
			gap: 1rem;
			align-items: stretch;
		}
		
		.left-controls, .right-controls {
			justify-content: center;
		}
		
		.search-input {
			min-width: 200px;
		}
		
		.bulk-actions {
			flex-wrap: wrap;
			justify-content: center;
		}
	}
</style>
