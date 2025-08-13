<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	// Props
	export let customers: any[] = [];
	export let syncingCustomers: Set<string> = new Set();
	export let bulkCreatingUsers = false;

	// Event dispatcher
	const dispatch = createEventDispatcher();

	// Format date
	function formatDate(dateString: string): string {
		if (!dateString) return 'N/A';
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	// Get sync status display
	function getSyncStatusDisplay(customer: any) {
		switch (customer.syncStatus) {
			case 'synced':
				return { text: '✅ Synced', class: 'status-synced' };
			case 'not_synced':
				return { text: '⚠️ Not Synced', class: 'status-not-synced' };
			case 'local_only':
				return { text: '🏠 Local Only', class: 'status-local-only' };
			case 'stripe_only':
				return { text: '💳 Stripe Only', class: 'status-stripe-only' };
			default:
				return { text: '❓ Unknown', class: 'status-unknown' };
		}
	}

	// Event handlers
	function handleCreateUser(customer: any) {
		dispatch('createUser', customer);
	}

	function handleCreateAllUsers() {
		dispatch('createAllUsers');
	}
</script>

{#if customers.length > 0}
	<div class="table-section">
		<div class="table-header">
			<h2>💳 Stripe Only Customers ({customers.length})</h2>
			<button 
				class="btn btn-success"
				on:click={handleCreateAllUsers}
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
		</div>
		<table class="customers-table">
			<thead>
				<tr>
					<th>Customer</th>
					<th>Email</th>
					<th>Source</th>
					<th>Local ID</th>
					<th>Stripe ID</th>
					<th>Plan</th>
					<th>Role</th>
					<th>Sync Status</th>
					<th>Created</th>
					<th>Actions</th>
				</tr>
			</thead>
			<tbody>
				{#each customers as customer}
					{@const statusDisplay = getSyncStatusDisplay(customer)}
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
						<td>
							<span class="plan-name">{customer.planName || 'N/A'}</span>
						</td>
						<td>
							<span class="customer-role">{customer.role || 'N/A'}</span>
						</td>
						<td>
							<span class="sync-status {statusDisplay.class}">
								{statusDisplay.text}
							</span>
						</td>
						<td>
							<span class="created-date">
								{formatDate(customer.createdAt || customer.stripeCreatedAt)}
							</span>
						</td>
						<td>
							<div class="customer-actions">
								<button 
									class="btn btn-sm btn-success"
									on:click={() => handleCreateUser(customer)}
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
							</div>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}

<style>
	/* Component-specific styles - inherits main styles from parent */
	.table-section {
		margin-bottom: var(--space-lg, 1.5rem);
		padding-bottom: var(--space-lg, 1.5rem);
		border-bottom: 1px solid var(--border, #e5e7eb);
	}

	.table-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-md, 1rem);
	}

	.table-header h2 {
		margin: 0;
		color: var(--text, #111827);
		font-size: 1.25rem;
		font-weight: 600;
	}

	.customers-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
		color: var(--text, #111827);
	}

	.customers-table th,
	.customers-table td {
		padding: var(--space-sm, 0.75rem) var(--space-md, 1rem);
		text-align: left;
		border-bottom: 1px solid var(--border, #e5e7eb);
	}

	.customers-table th {
		background-color: var(--bg-secondary, #f9fafb);
		font-weight: 600;
		color: var(--text-muted, #6b7280);
		text-transform: uppercase;
		font-size: 0.75rem;
		letter-spacing: 0.05em;
	}

	.customers-table tbody tr:hover {
		background-color: var(--bg-hover, #f3f4f6);
	}

	.customer-info {
		display: flex;
		align-items: center;
		gap: var(--space-sm, 0.75rem);
	}

	.customer-name h4 {
		margin: 0 0 var(--space-xs, 0.25rem) 0;
		color: var(--text, #111827);
		font-size: 1rem;
		font-weight: 600;
	}

	.customer-source {
		font-size: 0.75rem;
		color: var(--text-muted, #6b7280);
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

	.sync-status.status-not-synced {
		background-color: #fef3c7;
		color: #d97706;
	}

	.sync-status.status-local-only {
		background-color: #dbeafe;
		color: #2563eb;
	}

	.sync-status.status-stripe-only {
		background-color: #fee2e2;
		color: #dc2626;
	}

	.sync-status.status-unknown {
		background-color: #f3f4f6;
		color: #6b7280;
	}

	.customer-actions {
		display: flex;
		gap: var(--space-xs, 0.5rem);
		align-items: center;
	}

	.btn {
		padding: var(--space-sm, 0.75rem) var(--space-lg, 1.5rem);
		border: none;
		border-radius: var(--radius-md, 0.375rem);
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		gap: var(--space-xs, 0.5rem);
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

	.btn-sm {
		padding: var(--space-xs, 0.5rem) var(--space-md, 1rem);
		font-size: 0.75rem;
	}

	.loading-spinner {
		width: 40px;
		height: 40px;
		border: 3px solid #e5e7eb;
		border-top: 3px solid #2563eb;
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-bottom: var(--space-md, 1rem);
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