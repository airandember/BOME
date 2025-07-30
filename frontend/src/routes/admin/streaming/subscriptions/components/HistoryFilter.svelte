<script lang="ts">
	export let events: Record<string, any>[] = [];
	export let onFilterChange: (filteredEvents: Record<string, any>[]) => void = () => {};

	// Filter state
	let selectedEventType = 'all';
	let selectedDateRange = 'all';
	let searchTerm = '';

	// Available filter options
	const eventTypes = [
		{ value: 'all', label: 'All Events' },
		{ value: 'plan_created', label: 'Plan Created' },
		{ value: 'plan_updated', label: 'Plan Updated' },
		{ value: 'status_activated', label: 'Status Activated' },
		{ value: 'status_deactivated', label: 'Status Deactivated' },
		{ value: 'promotion_started', label: 'Promotion Started' },
		{ value: 'promotion_ended', label: 'Promotion Ended' },
		{ value: 'promotion_expired', label: 'Promotion Expired' },
		{ value: 'price_changed', label: 'Price Changed' },
		{ value: 'type_changed', label: 'Type Changed' }
	];

	const dateRanges = [
		{ value: 'all', label: 'All Time' },
		{ value: 'today', label: 'Today' },
		{ value: 'week', label: 'This Week' },
		{ value: 'month', label: 'This Month' },
		{ value: 'quarter', label: 'This Quarter' },
		{ value: 'year', label: 'This Year' }
	];

	// Computed filtered events
	$: filteredEvents = events.filter(event => {
		// Event type filter
		if (selectedEventType !== 'all' && event.event_type !== selectedEventType) {
			return false;
		}

		// Date range filter
		if (selectedDateRange !== 'all') {
			const eventDate = new Date(event.timestamp);
			const now = new Date();
			
			switch (selectedDateRange) {
				case 'today':
					const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
					if (eventDate < today) return false;
					break;
				case 'week':
					const weekAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
					if (eventDate < weekAgo) return false;
					break;
				case 'month':
					const monthAgo = new Date(now.getFullYear(), now.getMonth() - 1, now.getDate());
					if (eventDate < monthAgo) return false;
					break;
				case 'quarter':
					const quarterAgo = new Date(now.getFullYear(), now.getMonth() - 3, now.getDate());
					if (eventDate < quarterAgo) return false;
					break;
				case 'year':
					const yearAgo = new Date(now.getFullYear() - 1, now.getMonth(), now.getDate());
					if (eventDate < yearAgo) return false;
					break;
			}
		}

		// Search term filter
		if (searchTerm) {
			const searchLower = searchTerm.toLowerCase();
			const description = event.description?.toLowerCase() || '';
			const eventType = event.event_type?.toLowerCase() || '';
			const user = event.user_id?.toLowerCase() || '';
			
			if (!description.includes(searchLower) && 
				!eventType.includes(searchLower) && 
				!user.includes(searchLower)) {
				return false;
			}
		}

		return true;
	});

	// Handle filter changes
	function handleFilterChange() {
		onFilterChange(filteredEvents);
	}

	// Clear all filters
	function clearFilters() {
		selectedEventType = 'all';
		selectedDateRange = 'all';
		searchTerm = '';
		handleFilterChange();
	}

	// Export filtered events
	$: {
		handleFilterChange();
	}
</script>

<div class="history-filter">
	<div class="filter-header">
		<h4>Filter History</h4>
		<button class="clear-filters" on:click={clearFilters}>
			Clear Filters
		</button>
	</div>
	
	<div class="filter-controls">
		<div class="filter-group">
			<label for="event-type">Event Type:</label>
			<select id="event-type" bind:value={selectedEventType}>
				{#each eventTypes as type}
					<option value={type.value}>{type.label}</option>
				{/each}
			</select>
		</div>
		
		<div class="filter-group">
			<label for="date-range">Date Range:</label>
			<select id="date-range" bind:value={selectedDateRange}>
				{#each dateRanges as range}
					<option value={range.value}>{range.label}</option>
				{/each}
			</select>
		</div>
		
		<div class="filter-group">
			<label for="search">Search:</label>
			<input 
				id="search" 
				type="text" 
				placeholder="Search events..." 
				bind:value={searchTerm}
			/>
		</div>
	</div>
	
	<div class="filter-stats">
		<span class="stats-text">
			Showing {filteredEvents.length} of {events.length} events
		</span>
	</div>
</div>

<style>
	.history-filter {
		margin-bottom: 1.5rem;
		padding: 1rem;
		background: #f8fafc;
		border: 1px solid #e2e8f0;
		border-radius: 0.5rem;
	}

	.filter-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
	}

	.filter-header h4 {
		margin: 0;
		font-size: 1rem;
		font-weight: 600;
		color: #374151;
	}

	.clear-filters {
		background: #ef4444;
		color: white;
		border: none;
		padding: 0.25rem 0.75rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		font-weight: 500;
		cursor: pointer;
		transition: background-color 0.2s;
	}

	.clear-filters:hover {
		background: #dc2626;
	}

	.filter-controls {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.filter-group {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.filter-group label {
		font-size: 0.875rem;
		font-weight: 500;
		color: #6b7280;
	}

	.filter-group select,
	.filter-group input {
		padding: 0.5rem;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		background: white;
	}

	.filter-group select:focus,
	.filter-group input:focus {
		outline: none;
		border-color: #3b82f6;
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	.filter-stats {
		text-align: center;
	}

	.stats-text {
		font-size: 0.875rem;
		color: #6b7280;
		font-weight: 500;
	}

	@media (max-width: 768px) {
		.filter-controls {
			grid-template-columns: 1fr;
		}
		
		.filter-header {
			flex-direction: column;
			align-items: flex-start;
			gap: 0.5rem;
		}
	}
</style> 