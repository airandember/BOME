<script lang="ts">
	// Props with Svelte 5 syntax
	interface Props {
		currentPage?: number;
		totalPages?: number;
		totalItems?: number;
		itemsPerPage?: number;
		position?: 'top' | 'bottom';
		showItemsPerPageSelector?: boolean;
		showTotalInfo?: boolean;
		backgroundLoading?: boolean;
		onPageChange?: (page: number) => void;
		onItemsPerPageChange?: (itemsPerPage: number) => void;
	}

	let {
		currentPage = 1,
		totalPages = 1,
		totalItems = 0,
		itemsPerPage = 50,
		position = 'bottom',
		showItemsPerPageSelector = true,
		showTotalInfo = true,
		backgroundLoading = false,
		onPageChange,
		onItemsPerPageChange
	}: Props = $props();

	// Calculate displayed range - using derived state
	let startItem = $derived((currentPage - 1) * itemsPerPage + 1);
	let endItem = $derived(Math.min(currentPage * itemsPerPage, totalItems));

	// Generate page numbers to show - using derived state
	let pageNumbers = $derived(generatePageNumbers(currentPage, totalPages));

	function generatePageNumbers(current: number, total: number): (number | string)[] {
		if (total <= 7) {
			return Array.from({ length: total }, (_, i) => i + 1);
		}

		const pages: (number | string)[] = [];
		
		// Always show first page
		pages.push(1);
		
		if (current > 4) {
			pages.push('...');
		}
		
		// Show pages around current
		const start = Math.max(2, current - 1);
		const end = Math.min(total - 1, current + 1);
		
		for (let i = start; i <= end; i++) {
			if (i !== 1 && i !== total) {
				pages.push(i);
			}
		}
		
		if (current < total - 3) {
			pages.push('...');
		}
		
		// Always show last page
		if (total > 1) {
			pages.push(total);
		}
		
		return pages;
	}

	function handlePageClick(page: number) {
		if (page !== currentPage && page >= 1 && page <= totalPages) {
			onPageChange?.(page);
		}
	}

	function handleItemsPerPageChange(event: Event) {
		const target = event.target as HTMLSelectElement;
		const newItemsPerPage = parseInt(target.value);
		onItemsPerPageChange?.(newItemsPerPage);
	}
</script>

{#if totalItems > 0}
<div class="pagination-container {position}" class:background-loading={backgroundLoading}>
	{#if showTotalInfo}
		<div class="pagination-info">
			<span class="items-info">
				Showing {startItem.toLocaleString()}-{endItem.toLocaleString()} of {totalItems.toLocaleString()} items
			</span>
			{#if backgroundLoading}
				<span class="background-loading-indicator">
					<svg class="loading-spinner" viewBox="0 0 24 24">
						<circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-dasharray="31.416" stroke-dashoffset="31.416">
							<animate attributeName="stroke-dasharray" dur="2s" values="0 31.416;15.708 15.708;0 31.416" repeatCount="indefinite"/>
							<animate attributeName="stroke-dashoffset" dur="2s" values="0;-15.708;-31.416" repeatCount="indefinite"/>
						</circle>
					</svg>
					Loading...
				</span>
			{/if}
		</div>
	{/if}

	{#if totalPages > 1}
		<div class="pagination-controls">
			<!-- Previous button -->
			<button 
				type="button"
				class="pagination-btn prev-btn"
				class:disabled={currentPage === 1}
				disabled={currentPage === 1}
				onclick={() => handlePageClick(currentPage - 1)}
			>
				<span class="btn-icon">←</span>
				<span class="btn-text">Previous</span>
			</button>

			<!-- Page numbers -->
			<div class="page-numbers">
				{#each pageNumbers as pageNum}
					{#if pageNum === '...'}
						<span class="ellipsis">...</span>
					{:else}
						<button 
							type="button"
							class="page-btn"
							class:active={pageNum === currentPage}
							onclick={() => handlePageClick(pageNum as number)}
						>
							{pageNum}
						</button>
					{/if}
				{/each}
			</div>

			<!-- Next button -->
			<button 
				type="button"
				class="pagination-btn next-btn"
				class:disabled={currentPage === totalPages}
				disabled={currentPage === totalPages}
				onclick={() => handlePageClick(currentPage + 1)}
			>
				<span class="btn-text">Next</span>
				<span class="btn-icon">→</span>
			</button>
		</div>
	{/if}

	{#if showItemsPerPageSelector}
		<div class="items-per-page">
			<label for="items-per-page-{position}">Items per page:</label>
			<select 
				id="items-per-page-{position}"
				value={itemsPerPage} 
				onchange={handleItemsPerPageChange}
			>
				<option value="25">25</option>
				<option value="50">50</option>
				<option value="100">100</option>
				<option value="200">200</option>
			</select>
		</div>
	{/if}
</div>
{/if}

<style>
	.background-loading-indicator {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		margin-left: 1rem;
		color: #6b7280;
		font-size: 0.875rem;
		opacity: 0.8;
	}

	.loading-spinner {
		width: 16px;
		height: 16px;
		color: #3b82f6;
	}

	.pagination-container.background-loading {
		position: relative;
	}

	.pagination-container.background-loading::after {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: linear-gradient(90deg, transparent 0%, rgba(59, 130, 246, 0.1) 50%, transparent 100%);
		animation: shimmer 2s infinite;
		pointer-events: none;
		border-radius: 0.5rem;
	}

	@keyframes shimmer {
		0% {
			transform: translateX(-100%);
		}
		100% {
			transform: translateX(100%);
		}
	}

	.pagination-container {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		padding: 1rem;
		background: white;
		border-radius: 0.5rem;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
		flex-wrap: wrap;
	}

	.pagination-container.top {
		margin-bottom: 1rem;
	}

	.pagination-container.bottom {
		margin-top: 1rem;
	}

	.pagination-info {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.items-info {
		font-size: 0.875rem;
		color: #6b7280;
		font-weight: 500;
	}

	.pagination-controls {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.pagination-btn {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		background: white;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		color: #374151;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.pagination-btn:hover:not(.disabled) {
		background: #f9fafb;
		border-color: #9ca3af;
	}

	.pagination-btn.disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.btn-icon {
		font-size: 1rem;
	}

	.page-numbers {
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}

	.page-btn {
		min-width: 2.5rem;
		height: 2.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
		background: white;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		color: #374151;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.page-btn:hover {
		background: #f9fafb;
		border-color: #9ca3af;
	}

	.page-btn.active {
		background: #3b82f6;
		border-color: #3b82f6;
		color: white;
	}

	.ellipsis {
		padding: 0.5rem;
		color: #6b7280;
		font-weight: 500;
	}

	.items-per-page {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.items-per-page label {
		font-size: 0.875rem;
		color: #6b7280;
		font-weight: 500;
	}

	.items-per-page select {
		padding: 0.375rem 0.75rem;
		background: white;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		color: #374151;
		font-size: 0.875rem;
		cursor: pointer;
	}

	.items-per-page select:focus {
		outline: none;
		border-color: #3b82f6;
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	/* Responsive design */
	@media (max-width: 768px) {
		.pagination-container {
			flex-direction: column;
			gap: 0.75rem;
		}

		.pagination-controls {
			order: 1;
		}

		.pagination-info {
			order: 2;
		}

		.items-per-page {
			order: 3;
		}

		.btn-text {
			display: none;
		}

		.page-numbers {
			gap: 0.125rem;
		}

		.page-btn {
			min-width: 2rem;
			height: 2rem;
			font-size: 0.75rem;
		}
	}
</style>
