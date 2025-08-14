<script lang="ts">
	export let events: Record<string, any>[] = [];
	export let onFilterChange: (filteredEvents: Record<string, any>[]) => void = () => {};

	// Filter state
	let selectedEventType = 'all';
	let selectedDateRange = 'all';
	let selectedUser = 'all';
	let searchTerm = '';

	// Detect event types from the events array
	$: detectedEventTypes = (() => {
		const types = new Set<string>();
		events.forEach(event => {
			if (event.event_type) {
				types.add(event.event_type);
			}
		});
		return Array.from(types).sort();
	})();

	// Create event type options based on detected types
	$: eventTypes = [
		{ value: 'all', label: 'All Events' },
		...detectedEventTypes.map(type => ({
			value: type,
			label: type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())
		}))
	];

	const dateRanges = [
		{ value: 'all', label: 'All Time' },
		{ value: 'today', label: 'Today' },
		{ value: 'week', label: 'This Week' },
		{ value: 'month', label: 'This Month' },
		{ value: 'quarter', label: 'This Quarter' },
		{ value: 'year', label: 'This Year' }
	];

	// Get unique users from events
	$: uniqueUsers = (() => {
		const users = new Set<string>();
		events.forEach(event => {
			if (event.user_id) {
				try {
					// Try to parse as JSON to get user details
					const userData = typeof event.user_id === 'string' ? JSON.parse(event.user_id) : event.user_id;
					if (userData && userData.email) {
						users.add(userData.email);
					} else if (typeof event.user_id === 'string') {
						users.add(event.user_id);
					}
				} catch {
					// If not JSON, use as is
					users.add(event.user_id);
				}
			}
		});
		return Array.from(users).sort();
	})();

	// Computed filtered events
	$: filteredEvents = events.filter(event => {
		// Event type filter
		if (selectedEventType !== 'all' && event.event_type !== selectedEventType) {
			return false;
		}

		// Date range filter
		if (selectedDateRange !== 'all') {
			const eventDate = new Date(event.timestamp || event.created_at);
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

		// User filter
		if (selectedUser !== 'all') {
			let eventUser = '';
			try {
				const userData = typeof event.user_id === 'string' ? JSON.parse(event.user_id) : event.user_id;
				eventUser = userData && userData.email ? userData.email : event.user_id;
			} catch {
				eventUser = event.user_id;
			}
			if (eventUser !== selectedUser) {
				return false;
			}
		}

		// Search term filter
		if (searchTerm) {
			const searchLower = searchTerm.toLowerCase();
			const eventText = [
				event.event_type,
				event.description,
				event.user_id,
				JSON.stringify(event.metadata || {}),
				JSON.stringify(event.old_values || {}),
				JSON.stringify(event.new_values || {})
			].join(' ').toLowerCase();
			
			if (!eventText.includes(searchLower)) {
				return false;
			}
		}

		return true;
	});

	// Watch for filter changes and notify parent
	$: {
		onFilterChange(filteredEvents);
	}

	// Reset filters
	function resetFilters() {
		selectedEventType = 'all';
		selectedDateRange = 'all';
		selectedUser = 'all';
		searchTerm = '';
	}
</script>

<div class="history-filter">
	<div class="filter-header">
		<h4 class="filter-title">Filter History</h4>
		<button type="button" class="reset-btn" on:click={resetFilters}>
			Reset Filters
		</button>
	</div>

	<div class="filter-grid">
		<!-- Event Type Filter -->
		<div class="filter-group">
			<label for="event-type" class="filter-label">Event Type</label>
			<select id="event-type" bind:value={selectedEventType} class="filter-select">
				{#each eventTypes as type}
					<option value={type.value}>{type.label}</option>
				{/each}
			</select>
		</div>

		<!-- Date Range Filter -->
		<div class="filter-group">
			<label for="date-range" class="filter-label">Date Range</label>
			<select id="date-range" bind:value={selectedDateRange} class="filter-select">
				{#each dateRanges as range}
					<option value={range.value}>{range.label}</option>
				{/each}
			</select>
		</div>

		<!-- User Filter -->
		<div class="filter-group">
			<label for="user-filter" class="filter-label">User</label>
			<select id="user-filter" bind:value={selectedUser} class="filter-select">
				<option value="all">All Users</option>
				{#each uniqueUsers as user}
					<option value={user}>{user}</option>
				{/each}
			</select>
		</div>

		<!-- Search Filter -->
		<div class="filter-group">
			<label for="search-term" class="filter-label">Search</label>
			<input
				id="search-term"
				type="text"
				bind:value={searchTerm}
				placeholder="Search events..."
				class="filter-input"
			/>
		</div>
	</div>

	<div class="filter-summary">
		<span class="summary-text">
			Showing {filteredEvents.length} of {events.length} events
		</span>
	</div>
</div>

<style>
	.history-filter {
		background: #f9fafb;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		padding: 1rem;
		margin-bottom: 1rem;
	}

	.filter-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
	}

	.filter-title {
		font-size: 1rem;
		font-weight: 600;
		color: #111827;
		margin: 0;
	}

	.reset-btn {
		background: #f3f4f6;
		color: #374151;
		border: 1px solid #d1d5db;
		padding: 0.25rem 0.5rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.reset-btn:hover {
		background: #e5e7eb;
		border-color: #9ca3af;
	}

	.filter-grid {
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

	.filter-label {
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
	}

	.filter-select,
	.filter-input {
		padding: 0.5rem;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		background: white;
		transition: border-color 0.2s ease;
	}

	.filter-select:focus,
	.filter-input:focus {
		outline: none;
		border-color: #2563eb;
		box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
	}

	.filter-summary {
		border-top: 1px solid #e5e7eb;
		padding-top: 0.75rem;
	}

	.summary-text {
		font-size: 0.875rem;
		color: #6b7280;
	}

	@media (max-width: 640px) {
		.filter-grid {
			grid-template-columns: 1fr;
		}

		.filter-header {
			flex-direction: column;
			align-items: flex-start;
			gap: 0.5rem;
		}
	}
</style> 
