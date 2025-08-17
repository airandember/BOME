<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	export let videos: any[];
	export let searchTerm: string;
	export let statusFilter: string;
	export let categoryFilter: string;
	export let syncStatusFilter: string;
	export let vidStatusFilter: string;

	const dispatch = createEventDispatcher();

	function openCreateModal() {
		dispatch('openCreateModal');
	}
</script>

{#if videos.length === 0}
	<div class="empty-state">
		<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
			<polygon points="23,7 16,12 23,17 23,7"></polygon>
			<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
		</svg>
		<h3>No videos found</h3>
		<p>
			{#if searchTerm || statusFilter !== 'all' || categoryFilter !== 'all' || syncStatusFilter !== 'all' || vidStatusFilter !== 'all'}
				No videos match your current filters. Try adjusting your search criteria.
			{:else}
				Get started by creating your first video or using batch upload for multiple videos.
			{/if}
		</p>
		{#if !searchTerm && statusFilter === 'all' && categoryFilter === 'all' && syncStatusFilter === 'all' && vidStatusFilter === 'all'}
			<div class="flex justify-center space-x-3">
				<button
					class="btn btn-primary"
					on:click={openCreateModal}
				>
					Create Video
				</button>
				<a
					href="/admin/streaming/upload"
					class="btn btn-secondary"
				>
					Batch Upload
				</a>
			</div>
		{/if}
	</div>
{/if}

<style>
	/* Empty State */
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 300px;
		gap: 1rem;
		text-align: center;
	}

	.empty-state svg {
		width: 64px;
		height: 64px;
		color: var(--text-secondary);
	}

	.empty-state h3 {
		font-size: 1.5rem;
		font-weight: 600;
		margin: 0;
		color: var(--text-primary);
	}

	.empty-state p {
		color: var(--text-secondary);
		max-width: 400px;
		line-height: 1.5;
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

	.flex {
		display: flex;
	}

	.justify-center {
		justify-content: center;
	}

	.space-x-3 > * + * {
		margin-left: 0.75rem;
	}
</style>
