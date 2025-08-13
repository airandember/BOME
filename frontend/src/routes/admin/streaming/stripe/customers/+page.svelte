<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';
	import CustomerSyncPanel from '../components/CustomerSyncPanel.svelte';
	import { StripeCustomerSyncService } from '$lib/services/stripe-customer-sync';
	import { StreamingSubscriberService } from '$lib/services/streaming-subscribers';
	import { showToast } from '$lib/toast';
	import { goto } from '$app/navigation';

	// State variables
	let loading = true;
	let error = '';
	
	// Data arrays
	let localSubscribers: any[] = [];
	let stripeCustomers: any[] = [];
	let hybridCustomers: any[] = [];
	
	// Sync state
	let syncingCustomers = new Set<string>();
	
	// Stats
	let totalCount = 0;
	let syncedCount = 0;
	let localOnlyCount = 0;
	let stripeOnlyCount = 0;

	onMount(async () => {
		await loadAllData();
	});

	// Load both local subscribers and Stripe customers
	async function loadAllData() {
		try {
			loading = true;
			error = '';

			console.log('🔄 Loading local subscribers and Stripe customers...');

			// Fetch local subscribers with subscription plans
			const subscribersData = await StreamingSubscriberService.getSubscribers({ limit: 1000 });
			localSubscribers = subscribersData.subscribers || [];
			console.log('✅ Loaded local subscribers:', localSubscribers.length);

			// Fetch Stripe customers
			const stripeRes = await apiRequest('/admin/streaming/stripe/summary');
			if (stripeRes.ok) {
				const stripeData = await stripeRes.json();
				stripeCustomers = stripeData.summary?.customers || [];
				console.log('✅ Loaded Stripe customers:', stripeCustomers.length);
			} else {
				console.error('❌ Failed to load Stripe customers');
			}

			// Create hybrid customer list
			createHybridCustomerList();

		} catch (err) {
			error = 'Failed to load data';
			console.error('❌ Error loading data:', err);
		} finally {
			loading = false;
		}
	}

	// Create a hybrid list that compares local subscribers with Stripe customers
	function createHybridCustomerList() {
		const customerMap = new Map();

		console.log('🔄 Creating hybrid customer list...');

		// Add local subscribers first
		localSubscribers.forEach((subscriber: any) => {
			const key = subscriber.email.toLowerCase();
			customerMap.set(key, {
				id: `local_${subscriber.id}`,
				source: 'local',
				name: `${subscriber.first_name} ${subscriber.last_name}`.trim(),
				email: subscriber.email,
				localId: subscriber.id,
				role: subscriber.role,
				planName: subscriber.plan_name,
				stripePriceId: subscriber.stripe_price_id,
				createdAt: subscriber.created_at,
				stripeCustomerId: subscriber.stripe_customer_id,
				syncStatus: 'local_only'
			});
		});

		// Check Stripe customers against local subscribers
		stripeCustomers.forEach((stripeCustomer: any) => {
			const key = stripeCustomer.Email?.toLowerCase();
			if (!key) return;

			const existing = customerMap.get(key);
			
			if (existing) {
				// This customer exists in both systems
				existing.source = 'hybrid';
				existing.stripeId = stripeCustomer.ID;
				existing.stripeCreatedAt = stripeCustomer.CreatedAt;
				existing.stripeMetadata = stripeCustomer.Metadata;
				existing.syncStatus = existing.stripeCustomerId ? 'synced' : 'not_synced';
			} else {
				// This customer only exists in Stripe
				customerMap.set(key, {
					id: `stripe_${stripeCustomer.ID}`,
					source: 'stripe',
					name: stripeCustomer.Name || 'Unnamed Customer',
					email: stripeCustomer.Email,
					stripeId: stripeCustomer.ID,
					stripeCreatedAt: stripeCustomer.CreatedAt,
					stripeMetadata: stripeCustomer.Metadata,
					localId: stripeCustomer.Metadata?.local_customer_id || null,
					syncStatus: 'stripe_only'
				});
			}
		});

		// Convert map to array and sort by creation date
		hybridCustomers = Array.from(customerMap.values()).sort((a, b) => {
			const dateA = new Date(a.createdAt || a.stripeCreatedAt || 0);
			const dateB = new Date(b.createdAt || b.stripeCreatedAt || 0);
			return dateB.getTime() - dateA.getTime();
		});

		// Calculate stats
		totalCount = hybridCustomers.length;
		syncedCount = hybridCustomers.filter(c => c.syncStatus === 'synced').length;
		localOnlyCount = hybridCustomers.filter(c => c.syncStatus === 'local_only').length;
		stripeOnlyCount = hybridCustomers.filter(c => c.syncStatus === 'stripe_only').length;

		console.log('✅ Hybrid customer list created:', {
			total: totalCount,
			synced: syncedCount,
			localOnly: localOnlyCount,
			stripeOnly: stripeOnlyCount
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

	// Sync individual customer to Stripe
	async function syncCustomerToStripe(customer: any) {
		if (!customer.localId) {
			showToast('No local customer ID available for sync', 'warning');
			return;
		}

		if (syncingCustomers.has(customer.id)) return;

		try {
			syncingCustomers.add(customer.id);
			syncingCustomers = new Set(syncingCustomers); // Trigger reactivity

			const result = await StripeCustomerSyncService.syncCustomerToStripe(customer.localId);
			
			// Update the customer status
			customer.syncStatus = 'synced';
			customer.source = 'hybrid'; // Change from 'local' to 'hybrid' after sync
			customer.stripeCustomerId = result.stripe_id;
			customer.stripeId = result.stripe_id; // Also set stripeId for consistency
			hybridCustomers = [...hybridCustomers]; // Trigger reactivity
			
			showToast(`Customer synced: ${result.message}`, 'success');
		} catch (error) {
			console.error('Failed to sync customer:', error);
			showToast('Failed to sync customer', 'error');
		} finally {
			syncingCustomers.delete(customer.id);
			syncingCustomers = new Set(syncingCustomers); // Trigger reactivity
		}
	}

	// Add user from Stripe customer data
	function addUserFromStripe(customer: any) {
		if (!customer.stripeId) {
			showToast('Invalid Stripe customer ID', 'error');
			return;
		}

		// Parse name into first and last name if possible
		let firstName = '';
		let lastName = '';
		if (customer.name) {
			const nameParts = customer.name.trim().split(' ');
			if (nameParts.length === 1) {
				firstName = nameParts[0];
			} else if (nameParts.length >= 2) {
				firstName = nameParts[0];
				lastName = nameParts.slice(1).join(' ');
			}
		}

		// Navigate to users page with pre-filled data
		const queryParams = new URLSearchParams({
			stripe_customer_id: customer.stripeId,
			email: customer.email,
			first_name: firstName,
			last_name: lastName,
			source: 'stripe_import'
		});

		goto(`/admin/users/add?${queryParams.toString()}`);
	}

	// Format date
	function formatDate(dateString: string): string {
		if (!dateString) return 'N/A';
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	// Refresh data
	async function refreshData() {
		await loadAllData();
		showToast('Data refreshed', 'success');
	}
</script>

<div class="customers-page">
	<div class="page-header">
		<div class="header-content">
			<h1>👥 Customer Sync Dashboard</h1>
			<p>Compare local subscribers with Stripe customers and manage synchronization</p>
		</div>
		<div class="header-stats">
			<div class="stat-card">
				<span class="stat-value">{totalCount}</span>
				<span class="stat-label">Total Customers</span>
			</div>
			<div class="stat-card synced">
				<span class="stat-value">{syncedCount}</span>
				<span class="stat-label">Synced</span>
			</div>
			<div class="stat-card local">
				<span class="stat-value">{localOnlyCount}</span>
				<span class="stat-label">Local Only</span>
			</div>
			<div class="stat-card stripe">
				<span class="stat-value">{stripeOnlyCount}</span>
				<span class="stat-label">Stripe Only</span>
			</div>
		</div>
		
		<div class="header-actions">
			<button class="btn btn-secondary" on:click={refreshData}>
				🔄 Refresh Data
			</button>
		</div>
	</div>

	<!-- Customer Sync Panel -->
	<CustomerSyncPanel customerId={null} customerEmail="" />

	{#if loading}
		<div class="loading-container">
			<div class="loading-spinner"></div>
			<p>Loading customer data...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<h3>Error Loading Data</h3>
			<p>{error}</p>
			<button class="btn btn-primary" on:click={refreshData}>Retry</button>
		</div>
	{:else if hybridCustomers.length === 0}
		<div class="empty-state">
			<div class="empty-icon">👥</div>
			<h3>No Customers Found</h3>
			<p>No local subscribers or Stripe customers found.</p>
		</div>
	{:else}
		<div class="customers-table-container">
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
					{#each hybridCustomers as customer}
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
									{#if customer.source === 'local' || (customer.source === 'hybrid' && customer.syncStatus !== 'synced')}
										<button 
											class="btn btn-sm btn-primary"
											on:click={() => syncCustomerToStripe(customer)}
											disabled={syncingCustomers.has(customer.id)}
										>
											{#if syncingCustomers.has(customer.id)}
												<div class="loading-spinner small"></div>
												Syncing...
											{:else}
												🔄 Sync
											{/if}
										</button>
									{:else if customer.source === 'stripe'}
										<button 
											class="btn btn-sm btn-success"
											on:click={() => addUserFromStripe(customer)}
											title="Create new user with Stripe customer data"
										>
											➕ Add User
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
	{/if}
</div>

<style>
	.customers-page {
		padding: var(--space-lg, 1.5rem);
	}

	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: var(--space-xl, 2rem);
		flex-wrap: wrap;
		gap: var(--space-lg, 1.5rem);
	}

	.header-content h1 {
		margin: 0 0 var(--space-xs, 0.5rem) 0;
		color: var(--text, #111827);
		font-size: 2rem;
		font-weight: 700;
	}

	.header-content p {
		margin: 0;
		color: var(--text-muted, #6b7280);
		font-size: 1.1rem;
	}

	.header-stats {
		display: flex;
		gap: var(--space-md, 1rem);
		flex-wrap: wrap;
	}

	.stat-card {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: var(--space-md, 1rem);
		background: var(--surface, white);
		border-radius: var(--radius-lg, 0.5rem);
		border: 1px solid var(--border, #e5e7eb);
		min-width: 100px;
	}

	.stat-card.synced {
		border-color: #059669;
		background: #f0fdf4;
	}

	.stat-card.local {
		border-color: #2563eb;
		background: #eff6ff;
	}

	.stat-card.stripe {
		border-color: #d97706;
		background: #fffbeb;
	}

	.stat-value {
		font-size: 1.5rem;
		font-weight: bold;
		color: var(--primary, #2563eb);
		margin-bottom: var(--space-xs, 0.5rem);
	}

	.stat-label {
		font-size: 0.875rem;
		color: var(--text-muted, #6b7280);
		text-align: center;
	}

	.header-actions {
		display: flex;
		justify-content: flex-end;
		margin-top: var(--space-md, 1rem);
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

	.btn-primary {
		background: #2563eb;
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: #1d4ed8;
		transform: translateY(-1px);
	}

	.btn-success {
		background: #059669;
		color: white;
	}

	.btn-success:hover:not(:disabled) {
		background: #047857;
		transform: translateY(-1px);
	}

	.btn-secondary {
		background: #f3f4f6;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-secondary:hover:not(:disabled) {
		background: #e5e7eb;
	}

	.btn-sm {
		padding: var(--space-xs, 0.5rem) var(--space-md, 1rem);
		font-size: 0.75rem;
	}

	.loading-container, .error-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--space-xl, 2rem);
		text-align: center;
		min-height: 400px;
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--space-xl, 2rem);
		text-align: center;
		min-height: 400px;
	}

	.empty-icon {
		font-size: 4rem;
		margin-bottom: var(--space-lg, 1.5rem);
		opacity: 0.5;
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

	.customers-table-container {
		overflow-x: auto;
		border-radius: var(--radius-lg, 0.5rem);
		border: 1px solid var(--border, #e5e7eb);
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
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

	.sync-complete {
		color: #059669;
		font-weight: 600;
		font-size: 0.875rem;
	}

	@media (max-width: 768px) {
		.page-header {
			flex-direction: column;
			align-items: center;
			text-align: center;
		}

		.header-stats {
			justify-content: center;
		}

		.customers-table {
			font-size: 0.8125rem;
		}
	}
</style> 