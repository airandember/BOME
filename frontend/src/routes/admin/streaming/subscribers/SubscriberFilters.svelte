<script lang="ts">
	import type { Subscriber, NonSubscriber } from '$lib/services/streaming-subscribers';
	import DateRangePicker from './DateRangePicker.svelte';

	export let searchTerm = '';
	export let emailVerifiedFilter: boolean | undefined = undefined;
	export let roleFilter = '';
	export let planFilter = '';
	export let lastLoginFilter = '';
	export let createdDateFilter = '';
	export let hasSubbedFilter: boolean | undefined = undefined; // Updated filter
	export let subscribers: Subscriber[] = [];
	export let nonSubscribers: NonSubscriber[] = [];
	export let activeTab: 'subscribers' | 'non-subscribers' = 'subscribers';
	export let roles: any[] = [];

	// Callback props for Svelte 5
	export let onSearch: (term: string) => void = () => {};
	export let onFilterChange: (type: 'emailVerified' | 'role' | 'plan' | 'lastLogin' | 'createdDate' | 'hasSubbed', value: any) => void = () => {};
	export let onClearAll: () => void = () => {};

	// Debounced search for better performance
	let searchTimeout: ReturnType<typeof setTimeout>;

	// Date range states - synced with filter values
	$: lastLoginStartDate = (() => {
		if (lastLoginFilter && lastLoginFilter.includes('|')) {
			const [startDate] = lastLoginFilter.split('|');
			return startDate || '';
		}
		return '';
	})();

	$: lastLoginEndDate = (() => {
		if (lastLoginFilter && lastLoginFilter.includes('|')) {
			const [, endDate] = lastLoginFilter.split('|');
			return endDate || '';
		}
		return '';
	})();

	$: createdDateStartDate = (() => {
		if (createdDateFilter && createdDateFilter.includes('|')) {
			const [startDate] = createdDateFilter.split('|');
			return startDate || '';
		}
		return '';
	})();

	$: createdDateEndDate = (() => {
		if (createdDateFilter && createdDateFilter.includes('|')) {
			const [, endDate] = createdDateFilter.split('|');
			return endDate || '';
		}
		return '';
	})();

	// Get unique roles from current data with normalized names
	$: uniqueRoles = (() => {
		const allUsers = activeTab === 'subscribers' ? subscribers : nonSubscribers;
		const roleMap = new Map<string, string>();
		
		allUsers.forEach(user => {
			if (user.role) {
				// Find the normalized name from the roles array
				const normalizedRole = roles.find(r => r.id === user.role);
				const displayName = normalizedRole ? normalizedRole.name : user.role;
				// Use the display name as both the key and value
				roleMap.set(displayName, displayName);
			}
		});
		
		const result = Array.from(roleMap.entries()).sort((a, b) => a[1].localeCompare(b[1]));
		return result;
	})();

	// Get unique plans from subscribers data
	$: uniquePlans = (() => {
		if (activeTab !== 'subscribers') return [];
		
		const planMap = new Map<string, string>();
		
		subscribers.forEach(subscriber => {
			if (subscriber.plan_name) {
				planMap.set(subscriber.plan_name, subscriber.plan_name);
			}
		});
		
		const result = Array.from(planMap.entries()).sort((a, b) => a[1].localeCompare(b[1]));
		return result;
	})();

	function handleSearch() {
		// Clear existing timeout
		clearTimeout(searchTimeout);
		
		// Debounce search for 150ms
		searchTimeout = setTimeout(() => {
			onSearch(searchTerm);
		}, 150);
	}

	function handleEmailVerifiedChange() {
		onFilterChange('emailVerified', emailVerifiedFilter);
	}

	function handleRoleChange() {
		onFilterChange('role', roleFilter);
	}

	function handlePlanChange() {
		onFilterChange('plan', planFilter);
	}

	function handleSubscriptionHistoryChange() {
		onFilterChange('hasSubbed', hasSubbedFilter);
	}

	function handleClearAll() {
		searchTerm = '';
		emailVerifiedFilter = undefined;
		roleFilter = '';
		planFilter = '';
		lastLoginFilter = '';
		createdDateFilter = '';
		hasSubbedFilter = undefined;
		onClearAll();
	}

	// Check if any filters are active
	$: hasActiveFilters = searchTerm || 
		emailVerifiedFilter !== undefined || 
		roleFilter || 
		planFilter ||
		lastLoginFilter || 
		createdDateFilter ||
		hasSubbedFilter !== undefined;
</script>

<div class="filters-container">
	<!-- Search Filter -->
	<div class="filter-group">
		<label for="search" class="filter-label">Search</label>
		<input
			id="search"
			type="text"
			bind:value={searchTerm}
			on:input={handleSearch}
			placeholder="Search by name or email..."
			class="filter-input"
		/>
	</div>

	<!-- Email Verified Filter -->
	<div class="filter-group">
		<label for="email-verified" class="filter-label">Email Verified</label>
		<select
			id="email-verified"
			bind:value={emailVerifiedFilter}
			on:change={handleEmailVerifiedChange}
			class="filter-select"
		>
			<option value={undefined}>All</option>
			<option value={true}>Verified</option>
			<option value={false}>Not Verified</option>
		</select>
	</div>

	<!-- Role Filter -->
	<div class="filter-group">
		<label for="role" class="filter-label">Role</label>
		<select
			id="role"
			bind:value={roleFilter}
			on:change={handleRoleChange}
			class="filter-select"
		>
			<option value="">All Roles</option>
			{#each uniqueRoles as [value, label]}
				<option value={value}>{label}</option>
			{/each}
		</select>
	</div>

	<!-- Plan Filter (Subscribers only) -->
	{#if activeTab === 'subscribers'}
		<div class="filter-group">
			<label for="plan" class="filter-label">Plan</label>
			<select
				id="plan"
				bind:value={planFilter}
				on:change={handlePlanChange}
				class="filter-select"
			>
				<option value="">All Plans</option>
				{#each uniquePlans as [value, label]}
					<option value={value}>{label}</option>
				{/each}
			</select>
		</div>
	{/if}

	<!-- Has Subbed Filter (Non-subscribers only) -->
	{#if activeTab === 'non-subscribers'}
		<div class="filter-group">
			<label for="has-subbed" class="filter-label">Has Subbed</label>
			<select
				id="has-subbed"
				bind:value={hasSubbedFilter}
				on:change={handleSubscriptionHistoryChange}
				class="filter-select"
			>
				<option value={undefined}>All Users</option>
				<option value={false}>Never Subscribed</option>
				<option value={true}>Previously Subscribed</option>
			</select>
		</div>
	{/if}

	<!-- Last Login Filter -->
	<div class="filter-group">
		<label for="last-login" class="filter-label">Last Login</label>
		<DateRangePicker
			startDate={lastLoginStartDate}
			endDate={lastLoginEndDate}
			onDateRangeChange={(startDate, endDate) => {
				console.log('📅 SubscriberFilters: Last login date range changed:', { startDate, endDate });
				// Create filter value if either date is provided
				const filterValue = (startDate && startDate.trim() !== '') || (endDate && endDate.trim() !== '') 
					? `${startDate || ''}|${endDate || ''}` 
					: '';
				console.log('📅 SubscriberFilters: Calling onFilterChange with:', { type: 'lastLogin', value: filterValue });
				onFilterChange('lastLogin', filterValue);
			}}
		/>
	</div>

	<!-- Created Date Filter -->
	<div class="filter-group">
		<label for="created-date" class="filter-label">Created Date</label>
		<DateRangePicker
			startDate={createdDateStartDate}
			endDate={createdDateEndDate}
			onDateRangeChange={(startDate, endDate) => {
				console.log('📅 SubscriberFilters: Created date range changed:', { startDate, endDate });
				// Create filter value if either date is provided
				const filterValue = (startDate && startDate.trim() !== '') || (endDate && endDate.trim() !== '') 
					? `${startDate || ''}|${endDate || ''}` 
					: '';
				console.log('📅 SubscriberFilters: Calling onFilterChange with:', { type: 'createdDate', value: filterValue });
				onFilterChange('createdDate', filterValue);
			}}
		/>
	</div>

	<!-- Clear Filters Button -->
	<div class="filter-group">
		<button
			type="button"
			on:click={handleClearAll}
			class="clear-filters-btn"
		>
			Clear Filters
		</button>
	</div>
</div>

<style>
	.filters-container {
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
		margin-bottom: 1.5rem;
		padding: 1rem;
		background: #f9fafb;
		border-radius: 0.5rem;
		border: 1px solid #e5e7eb;
	}

	.filter-group {
		display: flex;
		flex-direction: column;
		min-width: 200px;
	}

	.filter-label {
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
		margin-bottom: 0.5rem;
	}

	.filter-input,
	.filter-select {
		padding: 0.5rem;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		background-color: white;
	}

	.filter-input:focus,
	.filter-select:focus {
		outline: none;
		border-color: #2563eb;
		box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
	}

	.clear-filters-btn {
		padding: 0.5rem 1rem;
		background: #ef4444;
		color: white;
		border: none;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: background-color 0.2s;
	}

	.clear-filters-btn:hover {
		background: #dc2626;
	}
</style> 
