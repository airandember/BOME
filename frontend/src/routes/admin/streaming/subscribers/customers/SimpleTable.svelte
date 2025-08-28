<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	
	export let title: string;
	export let customers: any[] = [];
	export let tableType: 'stripe-only' | 'synced' | 'local-only' = 'synced';
	export let syncingCustomers: Set<string> = new Set();
	export let bulkCreatingUsers = false;
	
	const dispatch = createEventDispatcher();
	
	function formatDate(dateString: string): string {
		if (!dateString) return 'N/A';
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	function handleCreateUser(customer: any) {
		dispatch('createUser', customer);
	}

	function handleCreateAllUsers() {
		dispatch('createAllUsers');
	}

	function handleSyncAllToStripe() {
		dispatch('syncAllToStripe');
	}

	function handleSyncToStripe(customer: any) {
		dispatch('syncToStripe', customer);
	}
</script>

{#if customers.length > 0}
	<div class="table-section">
		<div class="table-header">
			<h2>{title} ({customers.length})</h2>
			{#if tableType === 'stripe-only' && customers.length > 0}
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
			{:else if tableType === 'local-only' && customers.length > 0}
				<button 
					class="btn btn-warning"
					disabled={true}
					title="Sync all local users to Stripe (Coming Soon)"
				>
					🔄 Sync All to Stripe
				</button>
			{/if}
		</div>
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
					{#each customers as customer}
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
	</div>
{/if}

<style>
	.table-section {
		margin-bottom: 2rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid #e5e7eb;
	}

	.table-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
		padding: 1rem;
	}

	.table-header h2 {
		margin: 0;
		color: #111827;
		font-size: 1.25rem;
		font-weight: 600;
	}

	.table-container {
		overflow-x: auto;
		border-radius: 0.5rem;
		border: 1px solid #e5e7eb;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
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
