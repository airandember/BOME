<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { subscriberStore, filteredSubscribers, computedKPIs, subscriberStoreActions, type SubscriberFilters } from '$lib/stores/subscribers-store';
	import type { EnhancedSubscriber } from '$lib/types/enhanced-subscriber';
	import DataTable from '$lib/components/DataTable.svelte';
	import KPICard from '$lib/components/KPICard.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import EnhancedSubscriberViewModal from './EnhancedSubscriberViewModal.svelte';
	import EnhancedSubscriberEditModal from './EnhancedSubscriberEditModal.svelte';
	import { StripeWebhookAutoSync } from '$lib/services/webhook-auto-sync';
	
	// State
	let showViewModal = $state(false);
	let showEditModal = $state(false);
	let selectedSubscriber: EnhancedSubscriber | null = $state(null);
	let webhookAutoSync: StripeWebhookAutoSync | null = $state(null);
	let connectionStatus = $state({ isConnected: false, reconnectAttempts: 0, maxReconnectAttempts: 5 });
	let statusInterval: NodeJS.Timeout | null = $state(null);
	let webhookStatus = $state<any>(null);
	let webhookLoading = $state(false);
	
	// Filter state
	let searchTerm = $state('');
	let hasActivePlanFilter: boolean | undefined = $state(undefined);
	let hasVideoAccessFilter: boolean | undefined = $state(undefined);
	let planTypeFilter = $state('');
	let emailVerifiedFilter: boolean | undefined = $state(undefined);
	let roleFilter = $state('');
	let isExpiringSoonFilter: boolean | undefined = $state(undefined);
	let minMRRFilter: number | undefined = $state(undefined);
	let maxMRRFilter: number | undefined = $state(undefined);
	let minARRFilter: number | undefined = $state(undefined);
	let maxARRFilter: number | undefined = $state(undefined);
	
	// Pagination state
	let currentPage = $state(1);
	let itemsPerPage = $state(50);
	
	// Reactive store subscriptions (use $derived instead of $effect to avoid loops)
	let storeData = $derived($subscriberStore);
	let filteredData = $derived($filteredSubscribers);
	
	// Paginated data
	let paginatedData = $derived(() => {
		const startIndex = (currentPage - 1) * itemsPerPage;
		const endIndex = startIndex + itemsPerPage;
		return filteredData.slice(startIndex, endIndex);
	});
	
	// Pagination info
	let totalPages = $derived(Math.ceil(filteredData.length / itemsPerPage));
	let hasNextPage = $derived(currentPage < totalPages);
	let hasPrevPage = $derived(currentPage > 1);
	
	// Update filters when form values change
	$effect(() => {
		const filters: SubscriberFilters = {};
		
		if (searchTerm.trim()) filters.search = searchTerm;
		if (hasActivePlanFilter !== undefined) filters.hasActivePlan = hasActivePlanFilter;
		if (hasVideoAccessFilter !== undefined) filters.hasVideoAccess = hasVideoAccessFilter;
		if (planTypeFilter) filters.planType = planTypeFilter as any;
		if (emailVerifiedFilter !== undefined) filters.emailVerified = emailVerifiedFilter;
		if (roleFilter) filters.role = roleFilter;
		if (isExpiringSoonFilter !== undefined) filters.isExpiringSoon = isExpiringSoonFilter;
		if (minMRRFilter !== undefined) filters.minMRR = minMRRFilter;
		if (maxMRRFilter !== undefined) filters.maxMRR = maxMRRFilter;
		if (minARRFilter !== undefined) filters.minARR = minARRFilter;
		if (maxARRFilter !== undefined) filters.maxARR = maxARRFilter;
		
		subscriberStoreActions.setFilters(filters);
		
		// Reset to first page when filters change
		currentPage = 1;
	});
	
	// Table columns
	const columns = [
		{ key: 'email', label: 'Email', sortable: true, width: '200px' },
		{ key: 'first_name', label: 'First Name', sortable: true, width: '120px' },
		{ key: 'last_name', label: 'Last Name', sortable: true, width: '120px' },
		{ key: 'plan_name', label: 'Plan', sortable: true, width: '150px' },
		{ key: 'plan_legacy_status', label: 'Type', sortable: true, width: '80px' },
		{ key: 'has_video_access', label: 'Video Access', type: 'boolean' as const, sortable: true, width: '100px' },
		{ key: 'has_active_plan', label: 'Active Plan', type: 'boolean' as const, sortable: true, width: '100px' },
		{ key: 'email_verified', label: 'Verified', type: 'boolean' as const, sortable: true, width: '80px' },
		{ key: 'plan_status', label: 'Status', sortable: true, width: '100px' },
		{ key: 'billing_period_end', label: 'Expires', type: 'date' as const, sortable: true, width: '100px' },
		{ key: 'mrr_contribution', label: 'MRR', type: 'currency' as const, sortable: true, width: '80px' },
		{ key: 'arr_contribution', label: 'ARR', type: 'currency' as const, sortable: true, width: '80px' },
		{ key: 'days_until_expiry', label: 'Days Left', type: 'days_left' as const, sortable: true, width: '80px' },
		{ key: 'actions', label: 'Actions', type: 'actions' as const, width: '100px' }
	];
	
	// Bulk actions
	const bulkActions = [
		{ id: 'grant_access', label: 'Grant Video Access', icon: '🎬', variant: 'primary' as const },
		{ id: 'revoke_access', label: 'Revoke Video Access', icon: '🚫', variant: 'danger' as const },
		{ id: 'verify_email', label: 'Mark Email Verified', icon: '✅', variant: 'secondary' as const },
		{ id: 'send_reminder', label: 'Send Renewal Reminder', icon: '📧', variant: 'secondary' as const },
		{ id: 'export_selected', label: 'Export Selected', icon: '📊', variant: 'secondary' as const }
	];
	
	// Load data on mount
	onMount(async () => {
		console.log('🚀 Enhanced Subscribers Page: Loading data...');
		console.log('🔍 Initial storeData:', storeData);
		console.log('🔍 storeData.subscribers.length:', storeData.subscribers.length);
		
		await subscriberStoreActions.loadSubscribers();
		
		console.log('🔍 After loadSubscribers - storeData:', storeData);
		console.log('🔍 After loadSubscribers - subscribers.length:', storeData.subscribers.length);
		
		// Load webhook status
		await loadWebhookStatus();
		
		// TODO: Webhook auto-sync disabled - SSE endpoint not implemented in backend
		// webhookAutoSync = StripeWebhookAutoSync.getInstance();
		// webhookAutoSync.startListening();
		
		// Update connection status periodically (disabled)
		// statusInterval = setInterval(() => {
		// 	if (webhookAutoSync) {
		// 		connectionStatus = webhookAutoSync.getConnectionStatus();
		// 	}
		// }, 1000);
	});

	// Load webhook status
	async function loadWebhookStatus() {
		webhookLoading = true;
		try {
			const res = await fetch('/api/v1/admin/streaming/stripe/webhooks/status', {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`,
					'Content-Type': 'application/json'
				}
			});
			
			if (res.ok) {
				const data = await res.json();
				webhookStatus = data.webhook || data;
			}
		} catch (err) {
			console.error('Failed to load webhook status:', err);
			webhookStatus = { error: 'Failed to load webhook status' };
		} finally {
			webhookLoading = false;
		}
	}
	
	// Cleanup on destroy
	onDestroy(() => {
		// TODO: Re-enable when webhook auto-sync is implemented
		// if (webhookAutoSync) {
		// 	webhookAutoSync.stopListening();
		// }
		if (statusInterval) {
			clearInterval(statusInterval);
		}
	});
	
	// Event handlers
	function handleRowAction(event: { item: EnhancedSubscriber; action: string }) {
		selectedSubscriber = event.item;
		
		if (event.action === 'view') {
			showViewModal = true;
		} else if (event.action === 'edit') {
			showEditModal = true;
		}
	}
	
	function handleBulkAction(event: { action: string; items: any[]; requiresConfirmation?: boolean }) {
		console.log('🔧 Bulk action:', event.action, 'on', event.items.length, 'items');
		
		// TODO: Implement bulk actions
		switch (event.action) {
			case 'grant_access':
				// Grant video access to selected users
				break;
			case 'revoke_access':
				// Revoke video access from selected users
				break;
			case 'verify_email':
				// Mark emails as verified
				break;
			case 'send_reminder':
				// Send renewal reminders
				break;
			case 'export_selected':
				// Export selected users
				break;
		}
	}
	
	function handleExport(event: { data: any[]; format: string }) {
		console.log('📊 Exporting', event.data.length, 'records in', event.format, 'format');
		// TODO: Implement export functionality
	}
	
	function handleSubscriberUpdate(updatedSubscriber: EnhancedSubscriber) {
		subscriberStoreActions.updateSubscriber(updatedSubscriber);
		showEditModal = false;
		selectedSubscriber = null;
	}
	
	function clearAllFilters() {
		searchTerm = '';
		hasActivePlanFilter = undefined;
		hasVideoAccessFilter = undefined;
		planTypeFilter = '';
		emailVerifiedFilter = undefined;
		roleFilter = '';
		isExpiringSoonFilter = undefined;
		minMRRFilter = undefined;
		maxMRRFilter = undefined;
		minARRFilter = undefined;
		maxARRFilter = undefined;
		subscriberStoreActions.clearFilters();
		currentPage = 1; // Reset pagination
	}
	
	// Pagination functions
	function goToPage(page: number) {
		if (page >= 1 && page <= totalPages) {
			currentPage = page;
		}
	}
	
	function nextPage() {
		if (hasNextPage) {
			currentPage++;
		}
	}
	
	function prevPage() {
		if (hasPrevPage) {
			currentPage--;
		}
	}
	
	function changeItemsPerPage(newItemsPerPage: number) {
		itemsPerPage = newItemsPerPage;
		currentPage = 1; // Reset to first page
	}
	
	async function refreshData() {
		await subscriberStoreActions.refresh();
		await loadWebhookStatus(); // Also refresh webhook status
	}
</script>

<div class="enhanced-subscribers-page">
	<!-- Header -->
	<div class="page-header">
		<div class="header-content">
			<h1 class="page-title">Enhanced Subscribers</h1>
			<p class="page-description">
				Client-side filtered and searchable subscriber management with real-time updates
			</p>
		</div>
		
		<div class="header-actions">
			<!-- Real-time sync status -->
			<div class="sync-status">
				{#if webhookLoading}
					<span class="status-indicator reconnecting">⏳ Loading...</span>
				{:else if webhookStatus?.active}
					<span class="status-indicator connected">✅ Webhooks Active</span>
				{:else if webhookStatus?.error}
					<span class="status-indicator disconnected">❌ Webhook Error</span>
				{:else}
					<span class="status-indicator disconnected">🔴 Webhooks Inactive</span>
				{/if}
			</div>
			
			<button type="button" class="refresh-btn" onclick={refreshData} disabled={storeData.loading}>
				{#if storeData.loading}
					<span class="spinner"></span>
				{:else}
					🔄
				{/if}
				Refresh
			</button>
		</div>
	</div>
	
	<!-- KPIs -->
	{#if storeData.subscribers.length > 0}
		<div class="kpi-grid">
			<KPICard
				title="Total Customers"
				value={storeData.subscribers.length}
				icon="👥"
			/>
			<KPICard
				title="Active Plans"
				value={storeData.subscribers.filter(sub => sub.has_active_plan).length.toLocaleString()}
				icon="✅"
			/>
			<KPICard
				title="Video Access"
				value={storeData.subscribers.filter(sub => sub.has_video_access).length.toLocaleString()}
				icon="🎬"
			/>
			<!--<KPICard
				title="Total MRR"
				value="${Math.round(storeData.subscribers.reduce((sum, sub) => {
					const contribution = Number(sub.mrr_contribution) || 0;
					// MRR contribution should already be normalized to monthly
					return sum + contribution;
				}, 0)).toLocaleString()}"
				icon="💰"
			/>
			<KPICard
				title="Total ARR"
				value="${Math.round(storeData.subscribers.reduce((sum, sub) => {
					const contribution = Number(sub.mrr_contribution) || 0;
					// Convert MRR to ARR (monthly * 12)
					return sum + (contribution * 12);
				}, 0)).toLocaleString()}"
				icon="📈"
			/>-->
			<KPICard
				title="Churn Risk"
				value={storeData.subscribers.filter(sub => sub.is_expiring_soon).length.toLocaleString()}
				icon="⚠️"
			/>
		</div>
	{:else}
		<div class="loading-container">
			<p>Loading subscribers...</p>
		</div>
	{/if}
	
	<!-- Advanced Filters -->
	<div class="filters-section">
		<h3>Filters</h3>
		
		<!-- Search and MRR/ARR Range (Input Fields) -->
		<div class="input-filters">
			<div class="filter-group">
				<label for="search">Search</label>
				<input
					id="search"
					type="text"
					bind:value={searchTerm}
					placeholder="Email, name, or plan..."
					class="filter-input"
				/>
			</div>
			
			<!-- <div class="filter-group">
				<label for="minMRR">Min MRR</label>
				<input
					id="minMRR"
					type="number"
					bind:value={minMRRFilter}
					placeholder="0"
					min="0"
					step="0.01"
					class="filter-input"
				/>
			</div>
			
			<div class="filter-group">
				<label for="maxMRR">Max MRR</label>
				<input
					id="maxMRR"
					type="number"
					bind:value={maxMRRFilter}
					placeholder="100"
					min="0"
					step="0.01"
					class="filter-input"
				/>
			</div>
			
			<div class="filter-group">
				<label for="minARR">Min ARR</label>
				<input
					id="minARR"
					type="number"
					bind:value={minARRFilter}
					placeholder="0"
					min="0"
					step="0.01"
					class="filter-input"
				/>
			</div>
			
			<div class="filter-group">
				<label for="maxARR">Max ARR</label>
				<input
					id="maxARR"
					type="number"
					bind:value={maxARRFilter}
					placeholder="1200"
					min="0"
					step="0.01"
					class="filter-input"
				/>
			</div>-->
		</div>
		
		<!-- Button Filters -->
		<div class="button-filters">
			<!-- Active Plan Filter -->
			<div class="filter-button-group">
				<label>Active Plan</label>
				<div class="button-group">
					<button 
						type="button" 
						class="filter-btn {hasActivePlanFilter === undefined ? 'active' : ''}"
						onclick={() => hasActivePlanFilter = undefined}
					>
						All
					</button>
					<button 
						type="button" 
						class="filter-btn {hasActivePlanFilter === true ? 'active' : ''}"
						onclick={() => hasActivePlanFilter = true}
					>
						✅ Has Plan
					</button>
					<button 
						type="button" 
						class="filter-btn {hasActivePlanFilter === false ? 'active' : ''}"
						onclick={() => hasActivePlanFilter = false}
					>
						❌ No Plan
					</button>
				</div>
			</div>
			
			<!-- Video Access Filter -->
			<div class="filter-button-group">
				<label>Video Access</label>
				<div class="button-group">
					<button 
						type="button" 
						class="filter-btn {hasVideoAccessFilter === undefined ? 'active' : ''}"
						onclick={() => hasVideoAccessFilter = undefined}
					>
						All
					</button>
					<button 
						type="button" 
						class="filter-btn {hasVideoAccessFilter === true ? 'active' : ''}"
						onclick={() => hasVideoAccessFilter = true}
					>
						🎬 Has Access
					</button>
					<button 
						type="button" 
						class="filter-btn {hasVideoAccessFilter === false ? 'active' : ''}"
						onclick={() => hasVideoAccessFilter = false}
					>
						🚫 No Access
					</button>
				</div>
			</div>
			
			<!-- Plan Type Filter 
			<div class="filter-button-group">
				<label>Plan Type</label>
				<div class="button-group">
					<button 
						type="button" 
						class="filter-btn {planTypeFilter === '' ? 'active' : ''}"
						onclick={() => planTypeFilter = ''}
					>
						All
					</button>
					<button 
						type="button" 
						class="filter-btn {planTypeFilter === 'premium' ? 'active' : ''}"
						onclick={() => planTypeFilter = 'premium'}
					>
						⭐ Premium
					</button>
					<button 
						type="button" 
						class="filter-btn {planTypeFilter === 'basic' ? 'active' : ''}"
						onclick={() => planTypeFilter = 'basic'}
					>
						📦 Basic
					</button>
					<button 
						type="button" 
						class="filter-btn {planTypeFilter === 'none' ? 'active' : ''}"
						onclick={() => planTypeFilter = 'none'}
					>
						➖ None
					</button>
				</div>
			</div>-->
			
			<!-- Email Verified Filter -->
			<div class="filter-button-group">
				<label>Email Status</label>
				<div class="button-group">
					<button 
						type="button" 
						class="filter-btn {emailVerifiedFilter === undefined ? 'active' : ''}"
						onclick={() => emailVerifiedFilter = undefined}
					>
						All
					</button>
					<button 
						type="button" 
						class="filter-btn {emailVerifiedFilter === true ? 'active' : ''}"
						onclick={() => emailVerifiedFilter = true}
					>
						✅ Verified
					</button>
					<button 
						type="button" 
						class="filter-btn {emailVerifiedFilter === false ? 'active' : ''}"
						onclick={() => emailVerifiedFilter = false}
					>
						⚠️ Unverified
					</button>
				</div>
			</div>
			
			<!-- Expiring Soon Filter -->
			<div class="filter-button-group">
				<label>Expiry Status</label>
				<div class="button-group">
					<button 
						type="button" 
						class="filter-btn {isExpiringSoonFilter === undefined ? 'active' : ''}"
						onclick={() => isExpiringSoonFilter = undefined}
					>
						All
					</button>
					<button 
						type="button" 
						class="filter-btn {isExpiringSoonFilter === true ? 'active' : ''}"
						onclick={() => isExpiringSoonFilter = true}
					>
						⏰ Expiring Soon
					</button>
					<button 
						type="button" 
						class="filter-btn {isExpiringSoonFilter === false ? 'active' : ''}"
						onclick={() => isExpiringSoonFilter = false}
					>
						📅 Not Expiring
					</button>
				</div>
			</div>
		</div>
		
		<div class="filter-actions">
			<button type="button" class="clear-filters-btn" onclick={clearAllFilters}>
				🗑️ Clear All Filters
			</button>
			<span class="results-count">
				Showing {((currentPage - 1) * itemsPerPage) + 1}-{Math.min(currentPage * itemsPerPage, filteredData.length)} of {filteredData.length} subscribers
			</span>
		</div>
	</div>
	
	<!-- Data Table -->
	{#if storeData.loading}
		<div class="loading-container">
			<LoadingSpinner />
			<p>Loading subscriber data...</p>
		</div>
	{:else if storeData.error}
		<div class="error-container">
			<h3>❌ Error Loading Data</h3>
			<p>{storeData.error}</p>
			<button type="button" class="retry-btn" onclick={refreshData}>
				🔄 Retry
			</button>
		</div>
	{:else}
		<DataTable
			data={paginatedData()}
			{columns}
			{bulkActions}
			loading={storeData.loading}
			searchable={false}
			exportable={true}
			selectable={true}
			emptyMessage="No subscribers found matching your filters"
			searchPlaceholder="Search is handled by filters above..."
			onBulkAction={handleBulkAction}
			onRowAction={handleRowAction}
			onExport={handleExport}
		/>
		
		<!-- Pagination Controls -->
		{#if totalPages > 1}
			<div class="pagination-section">
				<div class="pagination-info">
					<span>Showing {((currentPage - 1) * itemsPerPage) + 1}-{Math.min(currentPage * itemsPerPage, filteredData.length)} of {filteredData.length} subscribers</span>
					
					<div class="items-per-page">
						<label for="itemsPerPage">Items per page:</label>
						<select id="itemsPerPage" bind:value={itemsPerPage} onchange={() => changeItemsPerPage(itemsPerPage)}>
							<option value={25}>25</option>
							<option value={50}>50</option>
							<option value={100}>100</option>
							<option value={200}>200</option>
						</select>
					</div>
				</div>
				
				<div class="pagination-controls">
					<button type="button" class="page-btn" disabled={!hasPrevPage} onclick={prevPage}>
						← Previous
					</button>
					
					{#each Array.from({ length: Math.min(7, totalPages) }, (_, i) => {
						const startPage = Math.max(1, Math.min(currentPage - 3, totalPages - 6));
						return startPage + i;
					}) as page}
						{#if page <= totalPages}
							<button
								type="button"
								class="page-btn {page === currentPage ? 'active' : ''}"
								onclick={() => goToPage(page)}
							>
								{page}
							</button>
						{/if}
					{/each}
					
					{#if totalPages > 7 && currentPage < totalPages - 3}
						<span class="page-ellipsis">...</span>
						<button type="button" class="page-btn" onclick={() => goToPage(totalPages)}>
							{totalPages}
						</button>
					{/if}
					
					<button type="button" class="page-btn" disabled={!hasNextPage} onclick={nextPage}>
						Next →
					</button>
				</div>
			</div>
		{/if}
	{/if}
	
	<!-- Last Updated Info -->
	{#if storeData.lastUpdated}
		<div class="last-updated">
			<small>
				Last updated: {storeData.lastUpdated.toLocaleString()}
				{#if storeData.loading}
					<span class="updating-indicator">Updating...</span>
				{/if}
			</small>
		</div>
	{/if}
</div>

<!-- Modals -->
{#if showViewModal && selectedSubscriber}
	<EnhancedSubscriberViewModal
		bind:isOpen={showViewModal}
		subscriber={selectedSubscriber}
		onClose={() => { showViewModal = false; selectedSubscriber = null; }}
	/>
{/if}

{#if showEditModal && selectedSubscriber}
	<EnhancedSubscriberEditModal
		bind:isOpen={showEditModal}
		subscriber={selectedSubscriber}
		onSave={handleSubscriberUpdate}
		onCancel={() => { showEditModal = false; selectedSubscriber = null; }}
	/>
{/if}

<style>
	.enhanced-subscribers-page {
		padding: 1.5rem;
		max-width: 100%;
		margin: 0 auto;
	}
	
	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 2rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid #e5e7eb;
	}
	
	.header-content h1 {
		font-size: 1.875rem;
		font-weight: 700;
		color: #111827;
		margin: 0 0 0.5rem 0;
	}
	
	.header-content p {
		color: #6b7280;
		font-size: 0.875rem;
		margin: 0;
	}
	
	.header-actions {
		display: flex;
		align-items: center;
		gap: 1rem;
	}
	
	.sync-status {
		display: flex;
		align-items: center;
	}
	
	.status-indicator {
		padding: 0.25rem 0.75rem;
		border-radius: 9999px;
		font-size: 0.75rem;
		font-weight: 500;
		border: 1px solid transparent;
	}
	
	.status-indicator.connected {
		background: #dcfce7;
		color: #166534;
		border-color: #bbf7d0;
	}
	
	.status-indicator.reconnecting {
		background: #fef3c7;
		color: #d97706;
		border-color: #fde68a;
	}
	
	.status-indicator.disconnected {
		background: #fef2f2;
		color: #dc2626;
		border-color: #fecaca;
	}
	
	.refresh-btn {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		background: #3b82f6;
		color: white;
		border: none;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: background-color 0.2s ease;
	}
	
	.refresh-btn:hover:not(:disabled) {
		background: #2563eb;
	}
	
	.refresh-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}
	
	.spinner {
		display: inline-block;
		width: 12px;
		height: 12px;
		border: 2px solid transparent;
		border-top: 2px solid currentColor;
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}
	
	@keyframes spin {
		to { transform: rotate(360deg); }
	}
	
	.kpi-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
		margin-bottom: 2rem;
	}
	
	.filters-section {
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		padding: 1.5rem;
		margin-bottom: 2rem;
	}
	
	.filters-section h3 {
		font-size: 1.125rem;
		font-weight: 600;
		color: #111827;
		margin: 0 0 1.5rem 0;
	}
	
	.input-filters {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
		margin-bottom: 2rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid #e5e7eb;
	}
	
	.button-filters {
		display: flex;
		flex-direction: row;
		flex-wrap: wrap;
		gap: 1.5rem;
		margin-bottom: 1.5rem;
		justify-content: center;
	}
	
	.filter-button-group {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		border-radius: 10px;
		padding: 1rem;
	}
	
	.filter-button-group label {
		font-size: 0.875rem;
		font-weight: 600;
		color: #374151;
		margin-bottom: 0.25rem;
		border-bottom: 1px solid #414b5f;
	}
	
	.button-group {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}
	
	.filter-btn {
		padding: 0.5rem 1rem;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		background: white;
		color: #374151;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s ease;
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}
	
	.filter-btn:hover {
		border-color: #3b82f6;
		background: #f8fafc;
	}
	
	.filter-btn.active {
		background: #3b82f6;
		color: white;
		border-color: #3b82f6;
		box-shadow: 0 1px 3px rgba(59, 130, 246, 0.3);
	}
	
	.filter-btn.active:hover {
		background: #2563eb;
		border-color: #2563eb;
	}
	
	.filter-group {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}
	
	.filter-group label {
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
	}
	
	.filter-input {
		padding: 0.5rem;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		transition: border-color 0.2s ease;
	}
	
	.filter-input:focus {
		outline: 2px solid #3b82f6;
		outline-offset: -2px;
		border-color: #3b82f6;
	}
	
	.filter-actions {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding-top: 1rem;
		border-top: 1px solid #e5e7eb;
	}
	
	.clear-filters-btn {
		padding: 0.5rem 1rem;
		background: #f3f4f6;
		color: #374151;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s ease;
	}
	
	.clear-filters-btn:hover {
		background: #e5e7eb;
		border-color: #9ca3af;
	}
	
	.results-count {
		font-size: 0.875rem;
		color: #6b7280;
		font-weight: 500;
	}
	
	.loading-container, .error-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 4rem 2rem;
		text-align: center;
	}
	
	.error-container h3 {
		color: #dc2626;
		margin: 0 0 0.5rem 0;
	}
	
	.error-container p {
		color: #6b7280;
		margin: 0 0 1.5rem 0;
	}
	
	.retry-btn {
		padding: 0.5rem 1rem;
		background: #dc2626;
		color: white;
		border: none;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: background-color 0.2s ease;
	}
	
	.retry-btn:hover {
		background: #b91c1c;
	}
	
	.last-updated {
		text-align: center;
		padding: 1rem;
		color: #6b7280;
		font-size: 0.75rem;
	}
	
	.updating-indicator {
		color: #3b82f6;
		font-weight: 500;
		margin-left: 0.5rem;
	}
	
	/* Pagination Styles */
	.pagination-section {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1.5rem;
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		margin-top: 1rem;
	}
	
	.pagination-info {
		display: flex;
		align-items: center;
		gap: 2rem;
		color: #6b7280;
		font-size: 0.875rem;
	}
	
	.items-per-page {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	
	.items-per-page label {
		font-weight: 500;
	}
	
	.items-per-page select {
		padding: 0.25rem 0.5rem;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		font-size: 0.875rem;
	}
	
	.pagination-controls {
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}
	
	.page-btn {
		padding: 0.5rem 0.75rem;
		border: 1px solid #d1d5db;
		background: white;
		color: #374151;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
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
		background: #f9fafb;
	}
	
	.page-ellipsis {
		padding: 0.5rem 0.25rem;
		color: #9ca3af;
		font-weight: 500;
	}
	
	/* Responsive design */
	@media (max-width: 768px) {
		.enhanced-subscribers-page {
			padding: 1rem;
		}
		
		.page-header {
			flex-direction: column;
			gap: 1rem;
		}
		
		.input-filters {
			grid-template-columns: 1fr;
		}
		
		.button-filters {
			gap: 1rem;
		}
		
		.filter-button-group {
			gap: 0.25rem;
		}
		
		.button-group {
			gap: 0.25rem;
		}
		
		.filter-btn {
			padding: 0.375rem 0.75rem;
			font-size: 0.75rem;
		}
		
		.filter-actions {
			flex-direction: column;
			gap: 1rem;
			align-items: stretch;
		}
		
		.kpi-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
