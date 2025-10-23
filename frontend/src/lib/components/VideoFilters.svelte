<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	export let searchTerm: string;
	// REMOVED: statusFilter - redundant with vidStatusFilter (vid_status boolean)
	export let categoryFilter: string;
	export let syncStatusFilter: string;
	export let vidStatusFilter: string;
	export let totalVideos: number;
	export let categories: string[];

	const dispatch = createEventDispatcher();

	function handleSearchInput() {
		dispatch('searchInput');
	}

	function handleFilterChange() {
		dispatch('filterChange');
	}

	function handleRefresh() {
		dispatch('refresh');
	}
</script>

<!-- Filters -->
<div class="filters-section">
	<div class="filters-grid">
		<div class="form-group">
			<label for="search">Search</label>
			<div class="dashSearch">
				<input
					id="search"
					class="dashSearch_input"
					type="text"
					bind:value={searchTerm}
					placeholder="Search videos..."
					on:keydown={(e) => e.key === 'Enter' && handleSearchInput()}
				/>
				<button
					class="dashSearch_glass"
					on:click={handleSearchInput}
				>
					🔍
				</button>
			</div>
		</div>
		<!-- REMOVED: Status Filter - redundant with Video Status (vid_status boolean) -->
		<div class="form-group">
			<label for="category-filter">Category</label>
			<select
				id="category-filter"
				bind:value={categoryFilter}
				on:change={handleFilterChange}
			>
				<option value="all">All Categories</option>
				{#each categories as category}
					<option value={category}>{category}</option>
				{/each}
			</select>
		</div>
		<div class="form-group">
			<label for="sync-status-filter">Sync Status</label>
			<select
				id="sync-status-filter"
				bind:value={syncStatusFilter}
				on:change={handleFilterChange}
			>
				<option value="all">All Sync Status</option>
				<option value="synced">Synced</option>
				<option value="needs_attention">Needs Attention</option>
				<option value="conflict">Conflict</option>
			</select>
		</div>
		<div class="form-group">
			<label for="vid-status-filter">Video Status</label>
			<select
				id="vid-status-filter"
				bind:value={vidStatusFilter}
				on:change={handleFilterChange}
			>
				<option value="all">All Status</option>
				<option value="true">Active</option>
				<option value="false">Inactive</option>
			</select>
		</div>
		<div class="form-group" style="justify-content:center">
			<label for="refresh-button">Count: {totalVideos}</label>
			<button
				id="refresh-button"
				class="btn w-full bg-gray-100 text-gray-700 px-4 py-2 rounded-md hover:bg-gray-200 transition-colors"
				style="height: 53.24px"
				on:click={handleRefresh}
			>
				Refresh
			</button>
		</div>
	</div>
</div>

<style>
	/* Filters Section */
	.filters-section {
		background: var(--bg-glass, rgba(255, 255, 255, 0.1));
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: 1.5rem;
		margin-bottom: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.filters-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
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
	.form-group select {
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
	.form-group select:focus {
		outline: none;
		border-color: var(--primary-color, #3b82f6);
		box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.15);
		background: var(--bg-primary, rgba(255, 255, 255, 0.05));
		transform: translateY(-1px);
	}

	.dashSearch {
		display: flex;
		flex-direction: row !important;
		flex-wrap: nowrap;
		align-items: center;
		justify-content: center;
	}

	.dashSearch_input {
		width: 85%;
		border-radius: 0.5rem 0 0 0.5rem !important;
	}

	.dashSearch_glass {
		width: 15%;
		border-radius: 0 0.5rem 0.5rem 0 !important;
		height: 97%;
		background: var(--bg-primary, rgba(255, 255, 255, 0.05));
		border: 2px solid var(--border-color, rgba(255, 255, 255, 0.2));
	}

	.dashSearch_glass:hover {
		background: var(--bg-tertiary, rgba(255, 255, 255, 0.05));
		border: 2px solid var(--border-color, rgba(255, 255, 255, 0.2));
	}

	.dashSearch_glass:active {
		background: var(--bg-secondary, rgba(255, 255, 255, 0.05));
		border: 2px solid var(--border-color, rgba(255, 255, 255, 0.2));
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

	.bg-gray-100 {
		background: var(--bg-secondary, rgba(255, 255, 255, 0.05)) !important;
		color: var(--text-primary, #ffffff) !important;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.bg-gray-100:hover {
		background: var(--bg-tertiary, rgba(255, 255, 255, 0.1)) !important;
		transform: translateY(-1px);
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.filters-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
