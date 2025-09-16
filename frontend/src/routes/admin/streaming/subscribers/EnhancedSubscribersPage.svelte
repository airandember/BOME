<script lang="ts">
	import { onMount } from 'svelte';
	import { StreamingSubscriberService, type Subscriber, type SubscriberFilters } from '$lib/services/streaming-subscribers';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import SubscriberFiltersComponent from './SubscriberFilters.svelte';
	import SubscriberTable from './SubscriberTable.svelte';
	import EnhancedPagination from './EnhancedPagination.svelte';

	// State for verified subscribers
	let verifiedSubscribers = $state<Subscriber[]>([]);
	let verifiedLoading = $state(false);
	let verifiedBackgroundLoading = $state(false); // For seamless pagination
	let verifiedCurrentPage = $state(1);
	let verifiedTotalPages = $state(1);
	let verifiedTotalCount = $state(0);
	let verifiedItemsPerPage = $state(50);

	// State for unverified subscribers
	let unverifiedSubscribers = $state<Subscriber[]>([]);
	let unverifiedLoading = $state(false);
	let unverifiedBackgroundLoading = $state(false); // For seamless pagination
	let unverifiedCurrentPage = $state(1);
	let unverifiedTotalPages = $state(1);
	let unverifiedTotalCount = $state(0);
	let unverifiedItemsPerPage = $state(50);

	// Accordion state
	let verifiedExpanded = $state(true);
	let unverifiedExpanded = $state(true);

	// Filter state
	let searchTerm = $state('');
	let roleFilter = $state('');
	let statusFilter = $state('');
	let lastLoginFilter = $state('');
	let createdDateFilter = $state('');

	// Roles data for filters
	let roles: any[] = $state([]);
	let rolesLoading = $state(true);

	// Selection state - using immutable patterns
	let selectedVerifiedSubscribers = $state<Set<number>>(new Set());
	let selectedUnverifiedSubscribers = $state<Set<number>>(new Set());
	let selectAllVerified = $state(false);
	let selectAllUnverified = $state(false);

	onMount(async () => {
		await Promise.all([
			loadRoles(),
			loadVerifiedSubscribers(),
			loadUnverifiedSubscribers()
		]);
	});

	async function loadRoles() {
		try {
			rolesLoading = true;
			// Load roles from your existing API
			const response = await fetch('/api/v1/admin/roles');
			if (response.ok) {
				const data = await response.json();
				roles = data.roles || [];
			}
		} catch (error) {
			console.error('Error loading roles:', error);
		} finally {
			rolesLoading = false;
		}
	}

	async function loadVerifiedSubscribers(seamless = false) {
		try {
			// Use background loading for seamless pagination, regular loading for initial load
			if (seamless && verifiedSubscribers.length > 0) {
				verifiedBackgroundLoading = true;
			} else {
				verifiedLoading = true;
			}
			
			const filters = buildFilters();
			
			const response = await StreamingSubscriberService.getSubscribersByEmailVerification(true, {
				limit: verifiedItemsPerPage,
				offset: (verifiedCurrentPage - 1) * verifiedItemsPerPage,
				filters
			});

			verifiedSubscribers = response.subscribers;
			verifiedTotalCount = response.pagination.total;
			verifiedTotalPages = Math.ceil(verifiedTotalCount / verifiedItemsPerPage);
			
			// Reset selection when data changes (but not during seamless pagination)
			if (!seamless) {
				selectedVerifiedSubscribers = new Set();
				selectAllVerified = false;
			}
		} catch (error) {
			console.error('Error loading verified subscribers:', error);
			showToast('Failed to load verified subscribers', 'error');
		} finally {
			verifiedLoading = false;
			verifiedBackgroundLoading = false;
		}
	}

	async function loadUnverifiedSubscribers(seamless = false) {
		try {
			// Use background loading for seamless pagination, regular loading for initial load
			if (seamless && unverifiedSubscribers.length > 0) {
				unverifiedBackgroundLoading = true;
			} else {
				unverifiedLoading = true;
			}
			
			const filters = buildFilters();
			
			const response = await StreamingSubscriberService.getSubscribersByEmailVerification(false, {
				limit: unverifiedItemsPerPage,
				offset: (unverifiedCurrentPage - 1) * unverifiedItemsPerPage,
				filters
			});

			unverifiedSubscribers = response.subscribers;
			unverifiedTotalCount = response.pagination.total;
			unverifiedTotalPages = Math.ceil(unverifiedTotalCount / unverifiedItemsPerPage);
			
			// Reset selection when data changes (but not during seamless pagination)
			if (!seamless) {
				selectedUnverifiedSubscribers = new Set();
				selectAllUnverified = false;
			}
		} catch (error) {
			console.error('Error loading unverified subscribers:', error);
			showToast('Failed to load unverified subscribers', 'error');
		} finally {
			unverifiedLoading = false;
			unverifiedBackgroundLoading = false;
		}
	}

	function buildFilters(): SubscriberFilters {
		const filters: SubscriberFilters = {};
		
		if (searchTerm.trim()) filters.search = searchTerm.trim();
		if (roleFilter) filters.role = roleFilter;
		if (statusFilter) filters.status = statusFilter;
		if (lastLoginFilter) filters.last_login = lastLoginFilter;
		if (createdDateFilter) filters.created_date = createdDateFilter;
		
		return filters;
	}

	async function handleSearch() {
		// Reset to first page and reload both sections
		verifiedCurrentPage = 1;
		unverifiedCurrentPage = 1;
		await Promise.all([
			loadVerifiedSubscribers(),
			loadUnverifiedSubscribers()
		]);
	}

	async function handleFilterChange() {
		await handleSearch();
	}

	async function handleClearAllFilters() {
		searchTerm = '';
		roleFilter = '';
		statusFilter = '';
		lastLoginFilter = '';
		createdDateFilter = '';
		await handleSearch();
	}

	// Note: Pagination handlers are now inline callbacks for better Svelte 5 compatibility

	// Selection handlers - using immutable state updates
	function handleVerifiedSelectItem(event: CustomEvent<{ itemId: number; checked: boolean }>) {
		const { itemId, checked } = event.detail;
		
		if (itemId === -1) {
			// Select all - create new Set for immutability
			selectedVerifiedSubscribers = checked 
				? new Set(verifiedSubscribers.map(s => s.id))
				: new Set();
			selectAllVerified = checked;
		} else {
			// Individual selection - create new Set for immutability
			const newSelection = new Set(selectedVerifiedSubscribers);
			if (checked) {
				newSelection.add(itemId);
			} else {
				newSelection.delete(itemId);
			}
			selectedVerifiedSubscribers = newSelection;
			selectAllVerified = selectedVerifiedSubscribers.size === verifiedSubscribers.length;
		}
	}

	function handleUnverifiedSelectItem(event: CustomEvent<{ itemId: number; checked: boolean }>) {
		const { itemId, checked } = event.detail;
		
		if (itemId === -1) {
			// Select all - create new Set for immutability
			selectedUnverifiedSubscribers = checked 
				? new Set(unverifiedSubscribers.map(s => s.id))
				: new Set();
			selectAllUnverified = checked;
		} else {
			// Individual selection - create new Set for immutability
			const newSelection = new Set(selectedUnverifiedSubscribers);
			if (checked) {
				newSelection.add(itemId);
			} else {
				newSelection.delete(itemId);
			}
			selectedUnverifiedSubscribers = newSelection;
			selectAllUnverified = selectedUnverifiedSubscribers.size === unverifiedSubscribers.length;
		}
	}

	function toggleAccordion(section: 'verified' | 'unverified') {
		if (section === 'verified') {
			verifiedExpanded = !verifiedExpanded;
		} else {
			unverifiedExpanded = !unverifiedExpanded;
		}
	}
</script>

<div class="enhanced-subscribers-page">
	<div class="page-header">
		<h1>Subscribers Management</h1>
		<p>Manage your subscribers organized by email verification status</p>
	</div>

	<!-- Global Filters -->
	<div class="filters-section">
		<SubscriberFiltersComponent 
			bind:searchTerm={searchTerm}
			bind:roleFilter={roleFilter}
			bind:planFilter={statusFilter}
			bind:lastLoginFilter={lastLoginFilter}
			bind:createdDateFilter={createdDateFilter}
			subscribers={[]}
			nonSubscribers={[]}
			activeTab="subscribers"
			{roles}
			onSearch={handleSearch}
			onFilterChange={handleFilterChange}
			onClearAll={handleClearAllFilters}
		/>
	</div>

	<!-- Verified Subscribers Section -->
	<div class="accordion-section">
		<button 
			type="button"
			class="accordion-header" 
			onclick={() => toggleAccordion('verified')}
			onkeydown={(e) => e.key === 'Enter' && toggleAccordion('verified')}
			aria-expanded={verifiedExpanded}
			aria-controls="verified-content"
		>
			<div class="accordion-title">
				<span class="accordion-icon">{verifiedExpanded ? '▼' : '▶'}</span>
				<span class="accordion-text">
					✅ Verified Email Subscribers ({verifiedTotalCount.toLocaleString()})
				</span>
			</div>
			<div class="accordion-actions">
				{#if selectedVerifiedSubscribers.size > 0}
					<span class="selection-count">
						{selectedVerifiedSubscribers.size} selected
					</span>
				{/if}
			</div>
		</button>

		{#if verifiedExpanded}
			<div class="accordion-content" id="verified-content">
				{#if verifiedLoading}
					<div class="loading-container">
						<LoadingSpinner />
						<p>Loading verified subscribers...</p>
					</div>
				{:else if verifiedTotalCount === 0}
					<div class="empty-state">
						<div class="empty-icon">✅</div>
						<h3>No Verified Subscribers</h3>
						<p>No subscribers with verified email addresses found matching your current filters.</p>
						{#if searchTerm || roleFilter || statusFilter || lastLoginFilter || createdDateFilter}
							<button type="button" class="btn-clear-filters" onclick={handleClearAllFilters}>
								Clear Filters
							</button>
						{/if}
					</div>
				{:else}
					<!-- Top Pagination -->
					<EnhancedPagination 
						currentPage={verifiedCurrentPage}
						totalPages={verifiedTotalPages}
						totalItems={verifiedTotalCount}
						itemsPerPage={verifiedItemsPerPage}
						position="top"
						backgroundLoading={verifiedBackgroundLoading}
						onPageChange={(page) => {
							verifiedCurrentPage = page;
							loadVerifiedSubscribers(true); // Seamless pagination
						}}
						onItemsPerPageChange={(itemsPerPage) => {
							verifiedItemsPerPage = itemsPerPage;
							verifiedCurrentPage = 1;
							loadVerifiedSubscribers(); // Not seamless for items per page change
						}}
					/>

					<!-- Subscribers Table -->
					<div class="table-container" class:transitioning={verifiedBackgroundLoading}>
						<SubscriberTable 
							subscribers={verifiedSubscribers}
							animationDirection="right"
							isAnimating={false}
							isTransitioning={verifiedBackgroundLoading}
							{roles}
							selectedSubscribers={selectedVerifiedSubscribers}
							selectAllSubscribers={selectAllVerified}
							on:selectItem={handleVerifiedSelectItem}
						/>
					</div>

					<!-- Bottom Pagination -->
					<EnhancedPagination 
						currentPage={verifiedCurrentPage}
						totalPages={verifiedTotalPages}
						totalItems={verifiedTotalCount}
						itemsPerPage={verifiedItemsPerPage}
						position="bottom"
						backgroundLoading={verifiedBackgroundLoading}
						onPageChange={(page) => {
							verifiedCurrentPage = page;
							loadVerifiedSubscribers(true); // Seamless pagination
						}}
						onItemsPerPageChange={(itemsPerPage) => {
							verifiedItemsPerPage = itemsPerPage;
							verifiedCurrentPage = 1;
							loadVerifiedSubscribers(); // Not seamless for items per page change
						}}
					/>
				{/if}
			</div>
		{/if}
	</div>

	<!-- Unverified Subscribers Section -->
	<div class="accordion-section">
		<button 
			type="button"
			class="accordion-header" 
			onclick={() => toggleAccordion('unverified')}
			onkeydown={(e) => e.key === 'Enter' && toggleAccordion('unverified')}
			aria-expanded={unverifiedExpanded}
			aria-controls="unverified-content"
		>
			<div class="accordion-title">
				<span class="accordion-icon">{unverifiedExpanded ? '▼' : '▶'}</span>
				<span class="accordion-text">
					🚫 Unverified Email Subscribers ({unverifiedTotalCount.toLocaleString()})
				</span>
			</div>
			<div class="accordion-actions">
				{#if selectedUnverifiedSubscribers.size > 0}
					<span class="selection-count">
						{selectedUnverifiedSubscribers.size} selected
					</span>
				{/if}
			</div>
		</button>

		{#if unverifiedExpanded}
			<div class="accordion-content" id="unverified-content">
				{#if unverifiedLoading}
					<div class="loading-container">
						<LoadingSpinner />
						<p>Loading unverified subscribers...</p>
					</div>
				{:else if unverifiedTotalCount === 0}
					<div class="empty-state">
						<div class="empty-icon">🚫</div>
						<h3>No Unverified Subscribers</h3>
						<p>No subscribers with unverified email addresses found matching your current filters.</p>
						{#if searchTerm || roleFilter || statusFilter || lastLoginFilter || createdDateFilter}
							<button type="button" class="btn-clear-filters" onclick={handleClearAllFilters}>
								Clear Filters
							</button>
						{/if}
					</div>
				{:else}
					<!-- Top Pagination -->
					<EnhancedPagination 
						currentPage={unverifiedCurrentPage}
						totalPages={unverifiedTotalPages}
						totalItems={unverifiedTotalCount}
						itemsPerPage={unverifiedItemsPerPage}
						position="top"
						backgroundLoading={unverifiedBackgroundLoading}
						onPageChange={(page) => {
							unverifiedCurrentPage = page;
							loadUnverifiedSubscribers(true); // Seamless pagination
						}}
						onItemsPerPageChange={(itemsPerPage) => {
							unverifiedItemsPerPage = itemsPerPage;
							unverifiedCurrentPage = 1;
							loadUnverifiedSubscribers(); // Not seamless for items per page change
						}}
					/>

					<!-- Subscribers Table -->
					<div class="table-container" class:transitioning={unverifiedBackgroundLoading}>
						<SubscriberTable 
							subscribers={unverifiedSubscribers}
							animationDirection="right"
							isAnimating={false}
							isTransitioning={unverifiedBackgroundLoading}
							{roles}
							selectedSubscribers={selectedUnverifiedSubscribers}
							selectAllSubscribers={selectAllUnverified}
							on:selectItem={handleUnverifiedSelectItem}
						/>
					</div>

					<!-- Bottom Pagination -->
					<EnhancedPagination 
						currentPage={unverifiedCurrentPage}
						totalPages={unverifiedTotalPages}
						totalItems={unverifiedTotalCount}
						itemsPerPage={unverifiedItemsPerPage}
						position="bottom"
						backgroundLoading={unverifiedBackgroundLoading}
						onPageChange={(page) => {
							unverifiedCurrentPage = page;
							loadUnverifiedSubscribers(true); // Seamless pagination
						}}
						onItemsPerPageChange={(itemsPerPage) => {
							unverifiedItemsPerPage = itemsPerPage;
							unverifiedCurrentPage = 1;
							loadUnverifiedSubscribers(); // Not seamless for items per page change
						}}
					/>
				{/if}
			</div>
		{/if}
	</div>
</div>

<style>
	.table-container {
		transition: opacity 0.2s ease-in-out, transform 0.2s ease-in-out;
	}

	.table-container.transitioning {
		opacity: 0.7;
		transform: translateY(2px);
	}

	.enhanced-subscribers-page {
		padding: 1.5rem;
		max-width: 100%;
		margin: 0 auto;
	}

	.page-header {
		margin-bottom: 2rem;
	}

	.page-header h1 {
		font-size: 2rem;
		font-weight: 700;
		color: #111827;
		margin: 0 0 0.5rem 0;
	}

	.page-header p {
		color: #6b7280;
		font-size: 1rem;
		margin: 0;
	}

	.filters-section {
		margin-bottom: 2rem;
	}

	.accordion-section {
		margin-bottom: 1.5rem;
		background: white;
		border-radius: 0.75rem;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
		overflow: hidden;
	}

	.accordion-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1.25rem 1.5rem;
		background: #f9fafb;
		border: none;
		border-bottom: 1px solid #e5e7eb;
		cursor: pointer;
		transition: background-color 0.2s ease;
		width: 100%;
		text-align: left;
		font-family: inherit;
		font-size: inherit;
	}

	.accordion-header:hover {
		background: #f3f4f6;
	}

	.accordion-header:focus {
		outline: 2px solid #3b82f6;
		outline-offset: -2px;
	}

	.accordion-title {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.accordion-icon {
		font-size: 0.875rem;
		color: #6b7280;
		transition: transform 0.2s ease;
	}

	.accordion-text {
		font-size: 1.125rem;
		font-weight: 600;
		color: #111827;
	}

	.accordion-actions {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.selection-count {
		font-size: 0.875rem;
		color: #3b82f6;
		font-weight: 500;
		background: #eff6ff;
		padding: 0.25rem 0.75rem;
		border-radius: 9999px;
	}

	.accordion-content {
		padding: 0;
	}

	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 3rem;
		gap: 1rem;
	}

	.loading-container p {
		color: #6b7280;
		font-size: 0.875rem;
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 4rem 2rem;
		text-align: center;
		background: #f9fafb;
		border-radius: 0.5rem;
		margin: 1rem;
	}

	.empty-icon {
		font-size: 3rem;
		margin-bottom: 1rem;
		opacity: 0.6;
	}

	.empty-state h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: #374151;
		margin: 0 0 0.5rem 0;
	}

	.empty-state p {
		color: #6b7280;
		font-size: 0.875rem;
		margin: 0 0 1.5rem 0;
		max-width: 400px;
		line-height: 1.5;
	}

	.btn-clear-filters {
		background: #3b82f6;
		color: white;
		border: none;
		padding: 0.5rem 1rem;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: background-color 0.2s ease;
	}

	.btn-clear-filters:hover {
		background: #2563eb;
	}

	.btn-clear-filters:focus {
		outline: 2px solid #3b82f6;
		outline-offset: 2px;
	}

	/* Responsive design */
	@media (max-width: 768px) {
		.enhanced-subscribers-page {
			padding: 1rem;
		}

		.page-header h1 {
			font-size: 1.5rem;
		}

		.accordion-header {
			padding: 1rem;
		}

		.accordion-text {
			font-size: 1rem;
		}
	}
</style>
