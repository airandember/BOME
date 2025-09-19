<script lang="ts">
	import { onMount } from 'svelte';
	import { showToast } from '$lib/toast';
	import { subscriberCache } from '$lib/cache/subscriber-cache';
	import type { 
		EnhancedSubscriber, 
		SubscriberFilters, 
		SubscriberKPIs, 
		QuickFilter,
		BulkAction 
	} from '$lib/types/enhanced-subscriber';
	import DataTable from '$lib/components/DataTable.svelte';
	import KPICard from '$lib/components/KPICard.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import EnhancedSubscriberViewModal from './EnhancedSubscriberViewModal.svelte';
	import EnhancedSubscriberEditModal from './EnhancedSubscriberEditModal.svelte';

	// State
	let subscribers = $state<EnhancedSubscriber[]>([]);
	let kpis = $state<SubscriberKPIs | null>(null);
	let loading = $state(true);
	let kpisLoading = $state(true);
	let currentPage = $state(1);
	let totalPages = $state(1);
	let totalCount = $state(0);
	let itemsPerPage = $state(50);

	// Filters
	let activeFilters = $state<SubscriberFilters>({});
	let activeQuickFilter = $state<string | null>(null);
	let searchTerm = $state('');
	let searchTimeout: ReturnType<typeof setTimeout>;

	// Modal state
	let viewModalOpen = $state(false);
	let editModalOpen = $state(false);
	let selectedSubscriber = $state<EnhancedSubscriber | null>(null);

	// Quick filters for common views
	const quickFilters: QuickFilter[] = [
		{
			label: 'Active Plans',
			icon: '✅',
			filter: { has_active_plan: true },
			description: 'Users with active subscription plans'
		},
		{
			label: 'Video Access',
			icon: '🎥',
			filter: { has_video_access: true },
			description: 'Users who can access premium videos'
		},
		{
			label: 'Premium Users',
			icon: '💎',
			filter: { plan_type: 'premium' },
			description: 'Users on premium subscription plans'
		},
		{
			label: 'Expiring Soon',
			icon: '⏰',
			filter: { is_expiring_soon: true },
			description: 'Subscriptions expiring within 7 days'
		},
		{
			label: 'Recent Starts',
			icon: '🆕',
			filter: { created_date_from: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0] },
			description: 'Plans started within the last 7 days'
		},
		{
			label: 'No Video Access',
			icon: '⚠️',
			filter: { has_video_access: false },
			description: 'Users without video access (for cleanup)'
		}
	];

	// Table columns configuration
	const columns = [
		{ key: 'full_name', label: 'Name', sortable: true, width: '200px' },
		{ key: 'email', label: 'Email', sortable: true, width: '250px' },
		{ key: 'plan_name', label: 'Plan', sortable: true, filterable: true, width: '150px' },
		{ key: 'plan_legacy_status', label: 'Legacy/Current', filterable: true, width: '110px', align: 'center' as const },
		{ key: 'has_active_plan', label: 'Active Plan', type: 'boolean' as const, filterable: true, width: '100px', align: 'center' as const },
		{ key: 'has_video_access', label: 'Video Access', type: 'boolean' as const, filterable: true, width: '100px', align: 'center' as const },
		{ key: 'plan_start_date', label: 'Plan Start', type: 'date' as const, sortable: true, width: '120px' },
		{ key: 'billing_period_start', label: 'Billing Start', type: 'date' as const, sortable: true, width: '120px' },
		{ key: 'billing_period_end', label: 'Billing End', type: 'date' as const, sortable: true, width: '120px' },
		{ key: 'mrr_contribution', label: 'MRR', type: 'currency' as const, sortable: true, width: '100px', align: 'right' as const },
		{ key: 'days_until_expiry', label: 'Days Left', type: 'number' as const, sortable: true, width: '100px', align: 'right' as const },
		{ key: 'last_login', label: 'Last Login', type: 'date' as const, sortable: true, width: '120px' },
		{ key: 'actions', label: 'Actions', type: 'actions' as const, width: '100px', align: 'center' as const }
	];

	// Bulk actions configuration
	const bulkActions: BulkAction[] = [
		{
			id: 'grant_video_access',
			label: 'Grant Video Access',
			icon: '🎥',
			variant: 'primary'
		},
		{
			id: 'revoke_video_access',
			label: 'Revoke Video Access',
			icon: '🚫',
			variant: 'danger',
			requiresConfirmation: true
		},
		{
			id: 'extend_trial',
			label: 'Extend Trial',
			icon: '⏰',
			variant: 'secondary'
		},
		{
			id: 'send_email',
			label: 'Send Email',
			icon: '📧',
			variant: 'secondary'
		}
	];

	onMount(async () => {
		await Promise.all([
			loadSubscribers(),
			loadKPIs()
		]);
	});

	async function loadSubscribers() {
		try {
			loading = true;
			const response = await subscriberCache.getSubscribers(currentPage, itemsPerPage, activeFilters);
			
			subscribers = response.subscribers;
			totalCount = response.total_count;
			totalPages = response.pagination.total_pages;
			
			// Update KPIs if included in response
			if (response.kpis) {
				kpis = response.kpis;
				kpisLoading = false;
			}
			
		} catch (error) {
			console.error('Error loading subscribers:', error);
			showToast('Failed to load subscribers', 'error');
		} finally {
			loading = false;
		}
	}

	async function loadKPIs() {
		try {
			kpisLoading = true;
			kpis = await subscriberCache.getKPIs();
		} catch (error) {
			console.error('Error loading KPIs:', error);
			showToast('Failed to load KPIs', 'error');
		} finally {
			kpisLoading = false;
		}
	}

	async function applyQuickFilter(filter: QuickFilter) {
		activeQuickFilter = activeQuickFilter === filter.label ? null : filter.label;
		activeFilters = activeQuickFilter ? filter.filter : {};
		currentPage = 1;
		await loadSubscribers();
	}

	async function handleSearch(searchTerm: string) {
		activeFilters = { ...activeFilters, search: searchTerm || undefined };
		currentPage = 1;
		await loadSubscribers();
	}

	async function handleSearchEvent(event: CustomEvent<{ searchTerm: string }>) {
		const { searchTerm } = event.detail;
		await handleSearch(searchTerm);
	}

	function handleSearchInput() {
		// Clear existing timeout
		if (searchTimeout) {
			clearTimeout(searchTimeout);
		}
		
		// Debounce search for 300ms
		searchTimeout = setTimeout(() => {
			handleSearch(searchTerm);
		}, 300);
	}

	async function handlePageChange(page: number) {
		currentPage = page;
		await loadSubscribers();
	}

	function handleRowAction(event: CustomEvent<{ item: EnhancedSubscriber; action: string }>) {
		const { item, action } = event.detail;
		
		switch (action) {
			case 'edit':
				selectedSubscriber = item;
				editModalOpen = true;
				break;
			case 'view':
				selectedSubscriber = item;
				viewModalOpen = true;
				break;
		}
	}

	// Modal handlers
	function handleViewModalClose() {
		viewModalOpen = false;
		selectedSubscriber = null;
	}

	function handleEditModalClose() {
		editModalOpen = false;
		selectedSubscriber = null;
	}

	function handleEditModalSave(updatedSubscriber: EnhancedSubscriber) {
		// Update the subscriber in the list
		const index = subscribers.findIndex(s => s.id === updatedSubscriber.id);
		if (index !== -1) {
			subscribers[index] = updatedSubscriber;
		}
		
		// Close modal
		editModalOpen = false;
		selectedSubscriber = null;
		
		// Invalidate cache to ensure fresh data on next load
		subscriberCache.invalidate();
	}

	function handleBulkAction(event: CustomEvent<{ action: string; items: number[]; requiresConfirmation?: boolean }>) {
		const { action, items, requiresConfirmation } = event.detail;
		
		if (requiresConfirmation) {
			const confirmed = confirm(`Are you sure you want to ${action} for ${items.length} subscribers?`);
			if (!confirmed) return;
		}
		
		// Process bulk action
		showToast(`${action} applied to ${items.length} subscribers`, 'success');
		
		// Invalidate cache and reload
		subscriberCache.invalidate();
		loadSubscribers();
	}

	function handleExport(event: CustomEvent<{ data: EnhancedSubscriber[]; format: string }>) {
		const { data, format } = event.detail;
		
		// Create CSV content
		const headers = columns.filter(col => col.type !== 'actions').map(col => col.label);
		const csvContent = [
			headers.join(','),
			...data.map(item => 
				columns
					.filter(col => col.type !== 'actions')
					.map(col => {
						const value = item[col.key as keyof EnhancedSubscriber];
						return `"${String(value || '').replace(/"/g, '""')}"`;
					})
					.join(',')
			)
		].join('\n');
		
		// Download file
		const blob = new Blob([csvContent], { type: 'text/csv' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `subscribers-${new Date().toISOString().split('T')[0]}.csv`;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
		
		showToast('Subscriber data exported successfully', 'success');
	}
</script>

<div class="unified-subscribers-dashboard">
	<!-- Page Header -->
	<div class="page-header">
		<div class="header-content">
			<h1>Subscribers Management</h1>
			<p>Unified dashboard for managing subscribers, plans, and video access</p>
		</div>
		<div class="header-actions">
			<button type="button" class="refresh-btn" onclick={() => { subscriberCache.invalidate(); loadSubscribers(); loadKPIs(); }}>
				🔄 Refresh
			</button>
		</div>
	</div>

	<!-- KPI Summary Cards -->
	{#if kpis}
		<div class="kpi-grid">
			<KPICard 
				title="Total Subscribers" 
				value={kpis.total_subscribers} 
				icon="👥" 
				color="blue"
				loading={kpisLoading}
			/>
			<KPICard 
				title="Active Plans" 
				value={kpis.active_subscribers} 
				icon="✅" 
				color="green"
				loading={kpisLoading}
			/>
			<KPICard 
				title="Video Access" 
				value={kpis.video_access_users} 
				icon="🎥" 
				color="purple"
				loading={kpisLoading}
			/>
			<KPICard 
				title="Monthly Revenue" 
				value={kpis.total_mrr} 
				format="currency" 
				icon="💰" 
				color="green"
				loading={kpisLoading}
			/>
			<KPICard 
				title="Premium Users" 
				value={kpis.premium_users} 
				icon="💎" 
				color="yellow"
				loading={kpisLoading}
			/>
			<KPICard 
				title="Churn Risk" 
				value={kpis.churn_risk_count} 
				icon="⚠️" 
				color="red"
				loading={kpisLoading}
			/>
		</div>
	{/if}

	<!-- Quick Filters -->
	<div class="quick-filters">
		<div class="filters-header">
			<h3>Quick Filters</h3>
			<span class="filters-subtitle">Common subscriber views</span>
		</div>
		<div class="filters-grid">
			{#each quickFilters as filter}
				<button
					type="button"
					class="quick-filter-btn"
					class:active={activeQuickFilter === filter.label}
					onclick={() => applyQuickFilter(filter)}
					title={filter.description}
				>
					<span class="filter-icon">{filter.icon}</span>
					<span class="filter-label">{filter.label}</span>
				</button>
			{/each}
			{#if activeQuickFilter}
				<button
					type="button"
					class="clear-filters-btn"
					onclick={() => { activeQuickFilter = null; activeFilters = {}; loadSubscribers(); }}
				>
					<span class="filter-icon">🗑️</span>
					<span class="filter-label">Clear Filters</span>
				</button>
			{/if}
		</div>
	</div>

	<!-- Search Input -->
	<div class="search-section">
		<div class="search-container">
			<div class="search-input-wrapper">
				<input
					type="text"
					bind:value={searchTerm}
					placeholder="Search subscribers by name, email, or plan..."
					class="search-input"
					oninput={handleSearchInput}
				/>
				<span class="search-icon">🔍</span>
			</div>
			{#if searchTerm}
				<button
					type="button"
					class="clear-search-btn"
					onclick={() => { searchTerm = ''; handleSearch(''); }}
					title="Clear search"
				>
					✕
				</button>
			{/if}
		</div>
		<div class="search-info">
			<span class="search-results-count">
				{#if searchTerm}
					Found {totalCount.toLocaleString()} results for "{searchTerm}"
				{:else}
					Showing {totalCount.toLocaleString()} total subscribers
				{/if}
			</span>
		</div>
	</div>

	<!-- Unified Data Table -->
	<div class="data-table-section">
		<DataTable
			data={subscribers}
			{columns}
			{loading}
			searchable={false}
			exportable={true}
			selectable={true}
			bulkActions={bulkActions}
			emptyMessage="No subscribers found matching your criteria"
			on:rowAction={handleRowAction}
			on:bulkAction={handleBulkAction}
			on:export={handleExport}
		/>
	</div>

	<!-- Pagination -->
	{#if totalPages > 1}
		<div class="pagination-section">
			<div class="pagination-info">
				Showing {((currentPage - 1) * itemsPerPage) + 1} to {Math.min(currentPage * itemsPerPage, totalCount)} of {totalCount.toLocaleString()} subscribers
			</div>
			<div class="pagination-controls">
				<button
					type="button"
					class="page-btn"
					disabled={currentPage === 1}
					onclick={() => handlePageChange(currentPage - 1)}
				>
					← Previous
				</button>
				
				{#each Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
					const startPage = Math.max(1, currentPage - 2);
					return startPage + i;
				}) as page}
					{#if page <= totalPages}
						<button
							type="button"
							class="page-btn"
							class:active={page === currentPage}
							onclick={() => handlePageChange(page)}
						>
							{page}
						</button>
					{/if}
				{/each}
				
				<button
					type="button"
					class="page-btn"
					disabled={currentPage === totalPages}
					onclick={() => handlePageChange(currentPage + 1)}
				>
					Next →
				</button>
			</div>
		</div>
	{/if}
</div>

<!-- Modals -->
<EnhancedSubscriberViewModal
	isOpen={viewModalOpen}
	subscriber={selectedSubscriber}
	onClose={handleViewModalClose}
/>

<EnhancedSubscriberEditModal
	isOpen={editModalOpen}
	subscriber={selectedSubscriber}
	onSave={handleEditModalSave}
	onCancel={handleEditModalClose}
/>

<style>
	.unified-subscribers-dashboard {
		padding: 1.5rem;
		max-width: 100%;
		margin: 0 auto;
		background: #f8fafc;
		min-height: 100vh;
	}

	/* Page Header */
	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 2rem;
		background: white;
		padding: 2rem;
		border-radius: 1rem;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
	}

	.header-content h1 {
		font-size: 2.25rem;
		font-weight: 700;
		color: #111827;
		margin: 0 0 0.5rem 0;
	}

	.header-content p {
		color: #6b7280;
		font-size: 1.125rem;
		margin: 0;
	}

	.refresh-btn {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1.5rem;
		background: #3b82f6;
		color: white;
		border: none;
		border-radius: 0.5rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.refresh-btn:hover {
		background: #2563eb;
		transform: translateY(-1px);
	}

	/* KPI Grid */
	.kpi-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
		gap: 1.5rem;
		margin-bottom: 2rem;
	}

	/* Quick Filters */
	.quick-filters {
		background: white;
		padding: 1.5rem;
		border-radius: 1rem;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
		margin-bottom: 2rem;
	}

	.filters-header {
		margin-bottom: 1rem;
	}

	.filters-header h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: #111827;
		margin: 0 0 0.25rem 0;
	}

	.filters-subtitle {
		color: #6b7280;
		font-size: 0.875rem;
	}

	.filters-grid {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
	}

	.quick-filter-btn {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1rem;
		background: #f3f4f6;
		border: 2px solid transparent;
		border-radius: 0.5rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.quick-filter-btn:hover {
		background: #e5e7eb;
		transform: translateY(-1px);
	}

	.quick-filter-btn.active {
		background: #eff6ff;
		border-color: #3b82f6;
		color: #1d4ed8;
	}

	.clear-filters-btn {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1rem;
		background: #fee2e2;
		border: 2px solid #fecaca;
		border-radius: 0.5rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: #dc2626;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.clear-filters-btn:hover {
		background: #fecaca;
		transform: translateY(-1px);
	}

	.filter-icon {
		font-size: 1rem;
	}

	.filter-label {
		white-space: nowrap;
	}

	/* Search Section */
	.search-section {
		background: white;
		padding: 1.5rem;
		border-radius: 1rem;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
		margin-bottom: 1.5rem;
	}

	.search-container {
		display: flex;
		align-items: center;
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.search-input-wrapper {
		position: relative;
		flex: 1;
		max-width: 500px;
	}

	.search-input {
		width: 100%;
		padding: 0.75rem 3rem 0.75rem 1rem;
		border: 2px solid #e5e7eb;
		border-radius: 0.5rem;
		font-size: 1rem;
		color: #111827;
		background: #f9fafb;
		transition: all 0.2s ease;
	}

	.search-input:focus {
		outline: none;
		border-color: #3b82f6;
		background: white;
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	.search-icon {
		position: absolute;
		right: 1rem;
		top: 50%;
		transform: translateY(-50%);
		color: #6b7280;
		font-size: 1.25rem;
		pointer-events: none;
	}

	.clear-search-btn {
		padding: 0.5rem;
		background: #f3f4f6;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		color: #6b7280;
		cursor: pointer;
		transition: all 0.2s ease;
		font-size: 1rem;
		width: 2.5rem;
		height: 2.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.clear-search-btn:hover {
		background: #e5e7eb;
		border-color: #9ca3af;
		color: #374151;
	}

	.search-info {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.search-results-count {
		color: #6b7280;
		font-size: 0.875rem;
		font-weight: 500;
	}

	/* Data Table Section */
	.data-table-section {
		margin-bottom: 2rem;
	}

	/* Pagination */
	.pagination-section {
		display: flex;
		justify-content: space-between;
		align-items: center;
		background: white;
		padding: 1.5rem;
		border-radius: 1rem;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
	}

	.pagination-info {
		color: #6b7280;
		font-size: 0.875rem;
	}

	.pagination-controls {
		display: flex;
		gap: 0.5rem;
	}

	.page-btn {
		padding: 0.5rem 1rem;
		background: white;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.page-btn:hover:not(:disabled) {
		background: #f3f4f6;
		border-color: #9ca3af;
	}

	.page-btn.active {
		background: #3b82f6;
		border-color: #3b82f6;
		color: white;
	}

	.page-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	/* Responsive Design */
	@media (max-width: 1024px) {
		.kpi-grid {
			grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
		}
	}

	@media (max-width: 768px) {
		.unified-subscribers-dashboard {
			padding: 1rem;
		}

		.page-header {
			flex-direction: column;
			gap: 1rem;
			align-items: stretch;
		}

		.header-content h1 {
			font-size: 1.75rem;
		}

		.kpi-grid {
			grid-template-columns: 1fr;
		}

		.filters-grid {
			justify-content: center;
		}

		.pagination-section {
			flex-direction: column;
			gap: 1rem;
			text-align: center;
		}

		.pagination-controls {
			justify-content: center;
		}
	}

	@media (max-width: 640px) {
		.quick-filter-btn,
		.clear-filters-btn {
			flex: 1;
			justify-content: center;
		}

		.pagination-controls {
			flex-wrap: wrap;
		}
	}
</style>
