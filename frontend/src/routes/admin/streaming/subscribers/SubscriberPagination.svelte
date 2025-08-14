<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	export let currentPage = 1;
	export let totalPages = 1;

	const dispatch = createEventDispatcher<{
		pageChange: { page: number };
	}>();

	function handlePageChange(page: number) {
		dispatch('pageChange', { page });
	}

    
</script>

{#if totalPages > 1}
	<div class="pagination">
		<button 
			class="pagination-btn" 
			disabled={currentPage === 1}
			on:click={() => handlePageChange(currentPage - 1)}
		>
			Previous
		</button>
		
		{#each Array.from({ length: totalPages }, (_, i) => i + 1) as page}
			<button 
				class="pagination-btn {currentPage === page ? 'active' : ''}"
				on:click={() => handlePageChange(page)}
			>
				{page}
			</button>
		{/each}
		
		<button 
			class="pagination-btn" 
			disabled={currentPage === totalPages}
			on:click={() => handlePageChange(currentPage + 1)}
		>
			Next
		</button>
	</div>
{/if}

<style>
	.pagination {
		display: flex;
		justify-content: center;
		gap: 0.5rem;
		margin-top: 2rem;
	}

	.pagination-btn {
		padding: 0.5rem 1rem;
		border: 1px solid var(--border-color);
		background: var(--bg-secondary);
		color: var(--text-primary);
		border-radius: 6px;
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.pagination-btn:hover:not(:disabled) {
		background: var(--bg-tertiary);
		border-color: var(--primary-color);
	}

	.pagination-btn.active {
		background: var(--primary-color);
		color: white;
		border-color: var(--primary-color);
	}

	.pagination-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
</style> 
