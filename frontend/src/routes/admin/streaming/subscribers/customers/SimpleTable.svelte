<script lang="ts">
	interface Props {
		title: string;
		customers?: any[];
		tableType?: 'stripe-only' | 'synced' | 'local-only';
		syncingCustomers?: Set<string>;
		bulkCreatingUsers?: boolean;
		initiallyExpanded?: boolean;
		enableSearch?: boolean;
		enablePagination?: boolean;
		itemsPerPage?: number;
		searchPlaceholder?: string;
		oncreateUser?: (customer: any) => void;
		oncreateAllUsers?: () => void;
		onsyncAllToStripe?: () => void;
		onsyncToStripe?: (customer: any) => void;
	}
	
	let {
		title,
		customers = [],
		tableType = 'synced',
		syncingCustomers = new Set(),
		bulkCreatingUsers = false,
		initiallyExpanded = false,
		enableSearch = false,
		enablePagination = false,
		itemsPerPage = 50,
		searchPlaceholder = "Search by name, email, or Stripe ID...",
		oncreateUser,
		oncreateAllUsers,
		onsyncAllToStripe,
		onsyncToStripe
	}: Props = $props();
	
	let isExpanded = $state(initiallyExpanded);
	
	// Search and pagination state
	let searchTerm = $state('');
	let currentPage = $state(1);
	
	// Filtered and paginated data
	let filteredCustomers = $derived.by(() => {
		if (!enableSearch || !searchTerm.trim()) {
			return customers;
		}
		
		const term = searchTerm.toLowerCase().trim();
		return customers.filter(customer => {
			const name = (customer.name || '').toLowerCase();
			const email = (customer.email || '').toLowerCase();
			const stripeId = (customer.stripeId || customer.stripe_id || '').toLowerCase();
			
			return name.includes(term) || 
				   email.includes(term) || 
				   stripeId.includes(term);
		});
	});
	
	let totalPages = $derived.by(() => {
		if (!enablePagination) return 1;
		return Math.ceil(filteredCustomers.length / itemsPerPage);
	});
	
	let displayedCustomers = $derived.by(() => {
		if (!enablePagination) {
			return filteredCustomers;
		}
		
		const startIndex = (currentPage - 1) * itemsPerPage;
		const endIndex = startIndex + itemsPerPage;
		return filteredCustomers.slice(startIndex, endIndex);
	});
	
	// Reset to page 1 when search changes
	$effect(() => {
		if (enableSearch) {
			currentPage = 1;
		}
	});
	
	function toggleAccordion() {
		isExpanded = !isExpanded;
	}
	
	// Pagination functions
	function goToPage(page: number) {
		if (page >= 1 && page <= totalPages) {
			currentPage = page;
		}
	}
	
	function nextPage() {
		if (currentPage < totalPages) {
			currentPage++;
		}
	}
	
	function prevPage() {
		if (currentPage > 1) {
			currentPage--;
		}
	}
	
	function handleSearch() {
		// Search is reactive via $derived, but we can add any additional logic here if needed
		currentPage = 1; // Reset to first page when searching
	}
	
	function clearSearch() {
		searchTerm = '';
		currentPage = 1;
	}
	
	// Dynamic title with counts
	let dynamicTitle = $derived.by(() => {
		const totalCount = customers.length;
		const filteredCount = filteredCustomers.length;
		const displayedCount = displayedCustomers.length;
		
		if (enableSearch && searchTerm.trim()) {
			return `${title} (${totalCount} total, ${filteredCount} filtered${enablePagination ? `, ${displayedCount} showing` : ''})`;
		} else if (enablePagination && totalPages > 1) {
			return `${title} (${totalCount} total, ${displayedCount} showing)`;
		} else {
			return `${title} (${totalCount} total)`;
		}
	});
	
	function formatDate(dateString: string): string {
		if (!dateString) return 'N/A';
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	function handleCreateUser(customer: any) {
		oncreateUser?.(customer);
	}

	function handleCreateAllUsers() {
		oncreateAllUsers?.();
	}

	function handleSyncAllToStripe() {
		onsyncAllToStripe?.();
	}

	function handleSyncToStripe(customer: any) {
		onsyncToStripe?.(customer);
	}
</script>

{#if customers.length > 0}
	<div class="accordion-section">
		<div class="accordion-header" onclick={toggleAccordion} role="button" tabindex="0" onkeydown={(e) => e.key === 'Enter' && toggleAccordion()}>
			<div class="accordion-title">
				<div class="accordion-icon {isExpanded ? 'expanded' : ''}">
					▶
				</div>
				<h2>{dynamicTitle}</h2>
			</div>
			
			{#if enableSearch}
				<div class="search-controls" onclick={(e) => e.stopPropagation()} role="search" onkeydown={(e) => e.key === 'Enter' && e.stopPropagation()}>
					<input
						type="text"
						placeholder={searchPlaceholder}
						bind:value={searchTerm}
						oninput={handleSearch}
						class="search-input"
					/>
					{#if searchTerm.trim()}
						<button 
							onclick={clearSearch}
							class="btn btn-outline btn-sm clear-search"
							title="Clear search"
						>
							Clear
						</button>
					{/if}
				</div>
			{/if}
			<div class="accordion-actions">
				{#if tableType === 'stripe-only' && customers.length > 0}
					<button 
						class="btn btn-success"
						onclick={(e) => { e.stopPropagation(); handleCreateAllUsers(); }}
						disabled={bulkCreatingUsers}
						title="Create local users for all Stripe-only customers"
					>
						{#if bulkCreatingUsers}
							<div class="loading-spinner small"></div>
							Creating All...
						{:else}
							➕ Add All Users
						{/if}
					</button>
				{:else if tableType === 'local-only' && customers.length > 0}
					<button 
						class="btn btn-warning"
						disabled={true}
						title="Sync all local users to Stripe (Coming Soon)"
						onclick={(e) => e.stopPropagation()}
					>
						🔄 Sync All to Stripe
					</button>
				{/if}
			</div>
		</div>
		
		{#if isExpanded}
			<div class="accordion-content">
				<div class="table-container">
			<table class="customers-table">
				<thead>
					<tr>
						<th>Customer</th>
						<th>Email</th>
						<th>Source</th>
						<th>Local ID</th>
						<th>Stripe ID</th>
						{#if tableType !== 'local-only'}
							<th>Plan</th>
						{/if}
						<th>Role</th>
						<th>Sync Status</th>
						<th>Created</th>
						<th>Actions</th>
					</tr>
				</thead>
				<tbody>
					{#each displayedCustomers as customer}
						<tr>
							<td>
								<div class="customer-info">
									<div class="customer-name">
										<h4>{customer.name || 'Unnamed Customer'}</h4>
										<span class="customer-source">{customer.source}</span>
									</div>
								</div>
							</td>
							<td>
								<span class="customer-email">{customer.email}</span>
							</td>
							<td>
								<span class="source-badge source-{customer.source}">
									{customer.source === 'local' ? '🏠 Local' : 
									 customer.source === 'stripe' ? '💳 Stripe' : '🔗 Hybrid'}
								</span>
							</td>
							<td>
								<span class="local-id">{customer.localId || 'N/A'}</span>
							</td>
							<td>
								<span class="stripe-id">
									{customer.stripeId ? `#${customer.stripeId.slice(-8)}` : 'N/A'}
								</span>
							</td>
							{#if tableType !== 'local-only'}
								<td>
									<span class="plan-name">{customer.planName || 'N/A'}</span>
								</td>
							{/if}
							<td>
								<span class="customer-role">{customer.role || 'N/A'}</span>
							</td>
							<td>
								<span class="sync-status status-{customer.syncStatus}">
									{#if customer.syncStatus === 'synced'}
										✅ Synced
									{:else if customer.syncStatus === 'stripe_only'}
										💳 Stripe Only
									{:else if customer.syncStatus === 'local_only'}
										🏠 Local Only
									{:else}
										❓ Unknown
									{/if}
								</span>
							</td>
							<td>
								<span class="created-date">
									{formatDate(customer.createdAt || customer.stripeCreatedAt)}
								</span>
							</td>
							<td>
								<div class="customer-actions">
									{#if tableType === 'stripe-only'}
										<button 
											class="btn btn-sm btn-success"
											onclick={() => handleCreateUser(customer)}
											disabled={syncingCustomers.has(customer.id) || bulkCreatingUsers}
											title="Create local user from Stripe customer data"
										>
											{#if syncingCustomers.has(customer.id)}
												<div class="loading-spinner small"></div>
												Creating...
											{:else}
												➕ Add User
											{/if}
										</button>
									{:else if tableType === 'local-only'}
										<button 
											class="btn btn-sm btn-warning"
											disabled={true}
											title="Sync to Stripe (Coming Soon)"
										>
											🔄 Sync to Stripe
										</button>
									{:else}
										<span class="sync-complete">✅ Synced</span>
									{/if}
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
		
		{#if enablePagination && totalPages > 1}
			<div class="pagination-controls">
				<div class="pagination">
					<button 
						onclick={prevPage}
						disabled={currentPage === 1}
						class="btn btn-outline btn-sm"
					>
						Previous
					</button>
					
					<span class="page-info">
						Page {currentPage} of {totalPages}
						({displayedCustomers.length} showing)
					</span>
					
					<button 
						onclick={nextPage}
						disabled={currentPage === totalPages}
						class="btn btn-outline btn-sm"
					>
						Next
					</button>
				</div>
			</div>
		{/if}
			</div>
		{/if}
	</div>
{/if}

<style>
	.accordion-section {
		margin-bottom: 1rem;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		overflow: hidden;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
	}

	.accordion-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem 1.5rem;
		background: #f9fafb;
		cursor: pointer;
		transition: background-color 0.2s ease;
		border-bottom: 1px solid #e5e7eb;
	}

	.accordion-header:hover {
		background: #f3f4f6;
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
		user-select: none;
	}

	.accordion-icon.expanded {
		transform: rotate(90deg);
	}

	.accordion-title h2 {
		margin: 0;
		color: #111827;
		font-size: 1.25rem;
		font-weight: 600;
	}

	.accordion-actions {
		display: flex;
		gap: 0.5rem;
		align-items: center;
	}

	.search-controls {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-right: 1rem;
	}

	.search-input {
		padding: 0.5rem;
		border: 1px solid #e5e7eb;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		background: white;
		color: #111827;
		min-width: 250px;
	}

	.search-input:focus {
		outline: none;
		border-color: #3b82f6;
		box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.25);
	}

	.clear-search {
		white-space: nowrap;
	}

	.pagination-controls {
		padding: 1rem 1.5rem;
		border-top: 1px solid #e5e7eb;
		background: #f9fafb;
	}

	.pagination {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 1rem;
	}

	.page-info {
		font-size: 0.875rem;
		color: #6b7280;
		white-space: nowrap;
	}

	.accordion-content {
		animation: slideDown 0.3s ease-out;
	}

	@keyframes slideDown {
		from {
			opacity: 0;
			max-height: 0;
		}
		to {
			opacity: 1;
			max-height: 1000px;
		}
	}

	.table-container {
		overflow-x: auto;
		border: none;
		border-radius: 0;
	}

	.customers-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
		color: #111827;
	}

	.customers-table th,
	.customers-table td {
		padding: 0.75rem 1rem;
		text-align: left;
		border-bottom: 1px solid #e5e7eb;
	}

	.customers-table th {
		background-color: #f9fafb;
		font-weight: 600;
		color: #6b7280;
		text-transform: uppercase;
		font-size: 0.75rem;
		letter-spacing: 0.05em;
	}

	.customers-table tbody tr:hover {
		background-color: #f3f4f6;
	}

	.customer-info {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.customer-name h4 {
		margin: 0 0 0.25rem 0;
		color: #111827;
		font-size: 1rem;
		font-weight: 600;
	}

	.customer-source {
		font-size: 0.75rem;
		color: #6b7280;
		text-transform: uppercase;
		font-weight: 500;
	}

	.source-badge {
		padding: 0.25rem 0.5rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
	}

	.source-local {
		background-color: #dbeafe;
		color: #1e40af;
	}

	.source-stripe {
		background-color: #fef3c7;
		color: #d97706;
	}

	.source-hybrid {
		background-color: #d1fae5;
		color: #059669;
	}

	.sync-status {
		padding: 0.25rem 0.5rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		display: inline-block;
		text-align: center;
		min-width: 80px;
	}

	.sync-status.status-synced {
		background-color: #d1fae5;
		color: #059669;
	}

	.sync-status.status-stripe_only {
		background-color: #fee2e2;
		color: #dc2626;
	}

	.sync-status.status-local_only {
		background-color: #dbeafe;
		color: #2563eb;
	}

	.customer-actions {
		display: flex;
		gap: 0.5rem;
		align-items: center;
	}

	.sync-complete {
		color: #059669;
		font-weight: 600;
		font-size: 0.875rem;
	}

	.btn {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
	}

	.btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-success {
		background: #059669;
		color: white;
	}

	.btn-success:hover:not(:disabled) {
		background: #047857;
		transform: translateY(-1px);
	}

	.btn-warning {
		background: #f59e0b;
		color: white;
	}

	.btn-warning:hover:not(:disabled) {
		background: #d97706;
		transform: translateY(-1px);
	}

	.btn-outline {
		background: white;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-outline:hover:not(:disabled) {
		background: #f9fafb;
		border-color: #9ca3af;
	}

	.btn-sm {
		padding: 0.5rem 1rem;
		font-size: 0.75rem;
	}

	.loading-spinner {
		width: 40px;
		height: 40px;
		border: 3px solid #e5e7eb;
		border-top: 3px solid #2563eb;
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-bottom: 1rem;
	}

	.loading-spinner.small {
		width: 12px;
		height: 12px;
		border-width: 2px;
		margin: 0 4px 0 0;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}
</style> 
