<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import LoadingSpinner from './LoadingSpinner.svelte';
	
	interface Column {
		key: string;
		label: string;
		sortable?: boolean;
		type?: 'text' | 'number' | 'currency' | 'date' | 'boolean' | 'actions';
		width?: string;
		minWidth?: string; // For responsive table
		priority?: 'high' | 'medium' | 'low'; // For mobile column priority
		align?: 'left' | 'center' | 'right';
		format?: (value: any) => string;
		sticky?: boolean; // For sticky columns
	}
	
	// Props
	export let data: any[] = [];
	export let columns: Column[] = [];
	export let loading = false;
	export let mobileLayout: 'cards' | 'scroll' | 'stack' = 'cards';
	export let stickyFirstColumn = true;
	export let emptyMessage = 'No data available';
	
	// State
	let selectedItems = new Set<any>();
	
	// Event dispatcher
	const dispatch = createEventDispatcher();
	
	// Get priority columns for mobile
	$: priorityColumns = columns.filter(col => 
		col.priority === 'high' || (!col.priority && columns.indexOf(col) < 3)
	);
	
	// Handle row action
	function handleRowAction(item: any, action: string) {
		dispatch('rowAction', { item, action });
	}
	
	// Format cell value
	function formatValue(value: any, column: Column): string {
		if (value === null || value === undefined) return '-';
		
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
</script>

<div class="responsive-table-container">
	{#if loading}
		<div class="loading-overlay">
			<LoadingSpinner />
			<p>Loading data...</p>
		</div>
	{:else if !data || data.length === 0}
		<div class="empty-state">
			<div class="empty-icon">📋</div>
			<h3>No Data Found</h3>
			<p>{emptyMessage}</p>
		</div>
	{:else}
		<!-- Desktop & Tablet Table -->
		<div class="table-wrapper">
			<table class="responsive-table" class:sticky-first={stickyFirstColumn}>
				<thead>
					<tr>
						{#each columns as column}
							<th
								class="table-header {column.align || 'left'}"
								class:sticky={column.sticky || (stickyFirstColumn && columns.indexOf(column) === 0)}
								class:priority-high={column.priority === 'high'}
								class:priority-medium={column.priority === 'medium'}
								class:priority-low={column.priority === 'low'}
								style:width={column.width}
								style:min-width={column.minWidth}
							>
								{column.label}
							</th>
						{/each}
					</tr>
				</thead>
				
				<tbody>
					{#each data as item, index}
						<tr class="table-row" onclick={() => handleRowAction(item, 'view')}>
							{#each columns as column}
								<td
									class="table-cell {column.align || 'left'}"
									class:sticky={column.sticky || (stickyFirstColumn && columns.indexOf(column) === 0)}
									class:priority-high={column.priority === 'high'}
									class:priority-medium={column.priority === 'medium'}
									class:priority-low={column.priority === 'low'}
								>
									{#if column.type === 'actions'}
										<div class="action-buttons">
											<button
												type="button"
												class="action-btn"
												onclick={(e) => { e.stopPropagation(); handleRowAction(item, 'edit'); }}
												title="Edit"
											>
												✏️
											</button>
											<button
												type="button"
												class="action-btn"
												onclick={(e) => { e.stopPropagation(); handleRowAction(item, 'view'); }}
												title="View"
											>
												👁️
											</button>
										</div>
									{:else}
										<span class="cell-content">
											{formatValue(item[column.key], column)}
										</span>
									{/if}
								</td>
							{/each}
						</tr>
					{/each}
				</tbody>
			</table>
		</div>

		<!-- Mobile Cards (shown only on mobile when mobileLayout is 'cards') -->
		{#if mobileLayout === 'cards'}
			<div class="mobile-cards">
				{#each data as item}
					<div class="mobile-card" onclick={() => handleRowAction(item, 'view')}>
						<div class="card-header">
							<h4 class="card-title">
								{formatValue(item[columns[0]?.key], columns[0])}
							</h4>
							<div class="card-actions">
								<button
									type="button"
									class="action-btn"
									onclick={(e) => { e.stopPropagation(); handleRowAction(item, 'edit'); }}
									title="Edit"
								>
									✏️
								</button>
								<button
									type="button"
									class="action-btn"
									onclick={(e) => { e.stopPropagation(); handleRowAction(item, 'view'); }}
									title="View"
								>
									👁️
								</button>
							</div>
						</div>
						
						<div class="card-body">
							{#each priorityColumns.slice(1) as column}
								{#if column.type !== 'actions' && item[column.key] !== undefined}
									<div class="card-field">
										<span class="field-label">{column.label}:</span>
										<span class="field-value">
											{formatValue(item[column.key], column)}
										</span>
									</div>
								{/if}
							{/each}
						</div>
					</div>
				{/each}
			</div>
		{/if}
	{/if}
</div>

<style>
	.responsive-table-container {
		background: white;
		border-radius: 0.75rem;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
		overflow: hidden;
	}

	.loading-overlay {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 3rem;
		gap: 1rem;
	}

	.empty-state {
		text-align: center;
		padding: 3rem;
	}

	.empty-icon {
		font-size: 3rem;
		margin-bottom: 1rem;
	}

	.empty-state h3 {
		margin: 0 0 0.5rem 0;
		color: #374151;
	}

	.empty-state p {
		color: #6b7280;
		margin: 0;
	}

	/* Table Styles */
	.table-wrapper {
		overflow-x: auto;
		-webkit-overflow-scrolling: touch;
	}

	.responsive-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
	}

	.table-header {
		background: #f9fafb;
		border-bottom: 1px solid #e5e7eb;
		padding: 0.75rem 1rem;
		text-align: left;
		font-weight: 600;
		color: #374151;
		white-space: nowrap;
	}

	.table-header.sticky {
		position: sticky;
		left: 0;
		z-index: 10;
		background: #f9fafb;
		box-shadow: 2px 0 4px rgba(0, 0, 0, 0.1);
	}

	.table-row {
		border-bottom: 1px solid #f3f4f6;
		cursor: pointer;
		transition: background-color 0.15s;
	}

	.table-row:hover {
		background: #f9fafb;
	}

	.table-cell {
		padding: 0.75rem 1rem;
		vertical-align: middle;
		color: #374151;
	}

	.table-cell.sticky {
		position: sticky;
		left: 0;
		background: white;
		z-index: 5;
		box-shadow: 2px 0 4px rgba(0, 0, 0, 0.05);
	}

	.table-row:hover .table-cell.sticky {
		background: #f9fafb;
	}

	.table-cell.center {
		text-align: center;
	}

	.table-cell.right {
		text-align: right;
	}

	.action-buttons {
		display: flex;
		gap: 0.5rem;
		justify-content: center;
	}

	.action-btn {
		background: none;
		border: none;
		cursor: pointer;
		padding: 0.25rem;
		border-radius: 0.25rem;
		font-size: 1rem;
		transition: background-color 0.15s;
	}

	.action-btn:hover {
		background: #f3f4f6;
	}

	/* Mobile Cards */
	.mobile-cards {
		display: none;
		flex-direction: column;
		gap: 1rem;
		padding: 1rem;
	}

	.mobile-card {
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		padding: 1rem;
		cursor: pointer;
		transition: all 0.15s;
	}

	.mobile-card:hover {
		border-color: #d1d5db;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
	}

	.card-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 0.75rem;
		padding-bottom: 0.75rem;
		border-bottom: 1px solid #f3f4f6;
	}

	.card-title {
		font-weight: 600;
		color: #374151;
		margin: 0;
		font-size: 1rem;
	}

	.card-actions {
		display: flex;
		gap: 0.5rem;
	}

	.card-body {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.card-field {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.field-label {
		font-weight: 500;
		color: #6b7280;
		font-size: 0.875rem;
	}

	.field-value {
		color: #374151;
		font-size: 0.875rem;
		text-align: right;
	}

	/* Responsive Behavior */
	@media (max-width: 1024px) {
		/* Hide low priority columns on tablets */
		.priority-low {
			display: none;
		}
	}

	@media (max-width: 768px) {
		/* Hide medium priority columns on mobile */
		.priority-medium {
			display: none;
		}

		/* Show mobile cards instead of table when layout is 'cards' */
		.table-wrapper {
			display: var(--mobile-table-display, block);
		}

		.mobile-cards {
			display: var(--mobile-cards-display, none);
		}

		/* For cards layout */
		.responsive-table-container[data-mobile-layout="cards"] .table-wrapper {
			display: none;
		}

		.responsive-table-container[data-mobile-layout="cards"] .mobile-cards {
			display: flex;
		}

		/* For scroll layout - make table scrollable */
		.responsive-table-container[data-mobile-layout="scroll"] .table-wrapper {
			display: block;
		}

		.responsive-table-container[data-mobile-layout="scroll"] .responsive-table {
			min-width: 600px;
		}

		/* For stack layout - stack cells vertically */
		.responsive-table-container[data-mobile-layout="stack"] .responsive-table,
		.responsive-table-container[data-mobile-layout="stack"] .table-header,
		.responsive-table-container[data-mobile-layout="stack"] .table-row {
			display: block;
		}

		.responsive-table-container[data-mobile-layout="stack"] .table-header {
			display: none;
		}

		.responsive-table-container[data-mobile-layout="stack"] .table-cell {
			display: flex;
			justify-content: space-between;
			padding: 0.5rem 1rem;
			border-bottom: 1px solid #f3f4f6;
		}

		.responsive-table-container[data-mobile-layout="stack"] .table-cell:before {
			content: attr(data-label) ": ";
			font-weight: 600;
			color: #6b7280;
		}
	}

	@media (max-width: 640px) {
		.responsive-table-container {
			margin: 0 -1rem;
			border-radius: 0;
		}

		.mobile-cards {
			padding: 0.5rem;
		}

		.mobile-card {
			margin: 0 0.5rem;
		}
	}
</style>
