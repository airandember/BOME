<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';
	import CustomerSyncPanel from '../components/CustomerSyncPanel.svelte';

	let summary: any = null;
	let loading = true;
	let error = '';

	export let data: any = null;

	$: customers = data?.customers || [];
	$: customersCount = data?.customers_count || 0;

	// Debug logging
	$: {
		console.log('=== CUSTOMERS DEBUG ===');
		console.log('Data received:', data);
		console.log('Customers array:', customers);
		console.log('Customers count:', customersCount);
		console.log('Data type:', typeof data);
		console.log('Data keys:', data ? Object.keys(data) : 'No data');
		console.log('Customers key exists:', data?.customers ? 'Yes' : 'No');
		console.log('Customers key type:', data?.customers ? typeof data.customers : 'N/A');
		console.log('Customers is array:', Array.isArray(data?.customers));
		console.log('========================');
	}

	onMount(async () => {
		if (data) {
			summary = data;
			loading = false;
		} else {
			await fetchSummary();
		}
	});

	async function fetchSummary() {
		try {
			loading = true;
			error = '';
			const res = await apiRequest('/admin/streaming/stripe/summary');
			if (res.ok) {
				const data = await res.json();
				summary = data.summary;
			} else {
				error = 'Failed to load customers';
			}
		} catch (err) {
			error = 'Failed to load customers';
			console.error(err);
		} finally {
			loading = false;
		}
	}

	function formatDate(date: string): string {
		return new Date(date).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function formatCurrency(amount: number, currency: string = 'usd') {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: currency.toUpperCase(),
		}).format(amount / 100); // Convert from cents
	}

	function formatPercentage(percent: number) {
		return `${percent}%`;
	}

	function getDiscountDisplay(coupon: any) {
		if (coupon.PercentOff) {
			return formatPercentage(coupon.PercentOff);
		} else if (coupon.AmountOff) {
			return formatCurrency(coupon.AmountOff, coupon.Currency);
		}
		return 'N/A';
	}

	function getDurationDisplay(duration: string) {
		switch (duration) {
			case 'once':
				return 'Single Use';
			case 'repeating':
				return 'Repeating';
			case 'forever':
				return 'Forever';
			default:
				return duration;
		}
	}

	function getStatusColor(valid: boolean) {
		return valid ? 'var(--success)' : 'var(--error)';
	}

	function getStatusText(valid: boolean) {
		return valid ? 'Active' : 'Inactive';
	}

	function getCustomerSubscriptions(customerId: string) {
		if (!summary.subscriptions) return [];
		return summary.subscriptions.filter((sub: any) => sub.CustomerID === customerId);
	}

	function getCustomerPayments(customerId: string) {
		if (!summary.payment_intents) return [];
		return summary.payment_intents.filter((pi: any) => pi.CustomerID === customerId);
	}
</script>

<div class="customers-page">
	<div class="page-header">
		<div class="header-content">
			<h1>👥 Customers</h1>
			<p>Manage customer data and synchronization with Stripe</p>
		</div>
		<div class="header-stats">
			<div class="stat-card">
				<span class="stat-value">{customersCount}</span>
				<span class="stat-label">Total Customers</span>
			</div>
			<div class="stat-card">
				<span class="stat-value">{customers.filter((c: any) => c.Metadata && Object.keys(c.Metadata).length > 0).length}</span>
				<span class="stat-label">With Metadata</span>
			</div>
			<div class="stat-card">
				<span class="stat-value">{customers.filter((c: any) => c.CreatedAt).length}</span>
				<span class="stat-label">Active</span>
			</div>
		</div>
		
		<div class="header-actions">
			<button class="btn btn-secondary" on:click={() => window.location.reload()}>
				🔄 Refresh Page
			</button>
			<button class="btn btn-primary">
				➕ Create Customer
			</button>
		</div>
	</div>

	<!-- Customer Sync Panel -->
	<CustomerSyncPanel customerId={null} customerEmail="" />

	{#if customers.length === 0}
		<div class="empty-state">
			<div class="empty-icon">👥</div>
			<h3>No Customers Found</h3>
			<p>You haven't created any customers yet. Create your first customer to start managing customer data.</p>
		</div>
	{:else}
		<div class="customers-table-container">
			<table class="customers-table">
				<thead>
					<tr>
						<th>Customer</th>
						<th>Email</th>
						<th>Created</th>
						<th>Local ID</th>
						<th>Role</th>
						<th>Metadata</th>
						<th>Actions</th>
					</tr>
				</thead>
				<tbody>
					{#each customers as customer}
						<tr>
							<td>
								<div class="customer-info">
									<div class="customer-name">
										<h4>{customer.Name || 'Unnamed Customer'}</h4>
										<span class="customer-id">#{customer.ID.slice(-8)}</span>
									</div>
								</div>
							</td>
							<td>
								<span class="customer-email">{customer.Email}</span>
							</td>
							<td>
								<span class="customer-created">
									{new Date(customer.CreatedAt).toLocaleDateString()}
								</span>
							</td>
							<td>
								<span class="local-id">
									{customer.Metadata?.local_customer_id || 'N/A'}
								</span>
							</td>
							<td>
								<span class="customer-role">
									{customer.Metadata?.role || 'N/A'}
								</span>
							</td>
							<td>
								{#if customer.Metadata && Object.keys(customer.Metadata).length > 0}
									<details class="metadata-details">
										<summary class="metadata-summary">📋 View</summary>
										<div class="metadata-content">
											{#each Object.entries(customer.Metadata || {}) as [key, value]}
												<div class="metadata-item">
													<span class="metadata-key">{key}:</span>
													<span class="metadata-value">{value}</span>
												</div>
											{/each}
										</div>
									</details>
								{:else}
									<span class="no-metadata">No metadata</span>
								{/if}
							</td>
							<td>
								<div class="customer-actions">
									<button class="btn btn-sm btn-secondary">
										Edit
									</button>
									<button class="btn btn-sm btn-primary">
										Sync
									</button>
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
		padding: var(--space-lg);
	}

	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: var(--space-xl);
		flex-wrap: wrap;
		gap: var(--space-lg);
	}

	.header-content h1 {
		margin: 0 0 var(--space-xs) 0;
		color: var(--text);
		font-size: 2rem;
		font-weight: 700;
	}

	.header-content p {
		margin: 0;
		color: var(--text-muted);
		font-size: 1.1rem;
	}

	.header-stats {
		display: flex;
		gap: var(--space-md);
		flex-wrap: wrap;
	}

	.stat-card {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: var(--space-md);
		background: var(--surface);
		border-radius: var(--radius-lg);
		border: 1px solid var(--border);
		min-width: 100px;
	}

	.stat-value {
		font-size: 1.5rem;
		font-weight: bold;
		color: var(--primary);
		margin-bottom: var(--space-xs);
	}

	.stat-label {
		font-size: 0.875rem;
		color: var(--text-muted);
		text-align: center;
	}

	.header-actions {
		display: flex;
		justify-content: flex-end;
		margin-top: var(--space-md);
	}

	.btn {
		padding: var(--space-sm) var(--space-lg);
		border: none;
		border-radius: var(--radius-md);
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		gap: var(--space-xs);
	}

	.btn-primary {
		background: var(--primary);
		color: white;
	}

	.btn-primary:hover {
		background: var(--primary-hover);
		transform: translateY(-1px);
	}

	.btn-secondary {
		background: var(--secondary);
		color: white;
	}

	.btn-secondary:hover {
		background: var(--secondary-hover);
		transform: translateY(-1px);
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--space-xl);
		text-align: center;
		min-height: 400px;
	}

	.empty-icon {
		font-size: 4rem;
		margin-bottom: var(--space-lg);
		opacity: 0.5;
	}

	.empty-state h3 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
		font-size: 1.5rem;
	}

	.empty-state p {
		margin: 0;
		color: var(--text-muted);
		font-size: 1.1rem;
		max-width: 500px;
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
		padding: var(--space-sm, 0.5rem) var(--space-md, 1rem);
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

	.customers-table tbody tr:nth-child(even) {
		background-color: var(--bg-secondary, #f9fafb);
	}

	.customers-table tbody tr:nth-child(even):hover {
		background-color: var(--bg-hover, #f3f4f6);
	}

	.customer-info {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
	}

	.customer-name h4 {
		margin: 0 0 var(--space-xs, 0.25rem) 0;
		color: var(--text, #111827);
		font-size: 1rem;
		font-weight: 600;
	}

	.customer-id {
		font-size: 0.75rem;
		color: var(--text-muted, #6b7280);
		font-family: var(--font-mono, monospace);
	}

	.customer-email {
		color: var(--text-muted, #6b7280);
		font-size: 0.875rem;
	}

	.customer-created {
		color: var(--text-muted, #6b7280);
		font-size: 0.875rem;
	}

	.local-id {
		color: var(--text-muted, #6b7280);
		font-size: 0.875rem;
	}

	.customer-role {
		color: var(--text-muted, #6b7280);
		font-size: 0.875rem;
	}

	.no-metadata {
		color: var(--text-muted, #6b7280);
		font-size: 0.875rem;
	}

	.customer-actions {
		display: flex;
		gap: var(--space-xs, 0.25rem);
	}

	.btn-sm {
		padding: var(--space-xs, 0.25rem) var(--space-md, 1rem);
		font-size: 0.75rem;
	}

	.metadata-details {
		border: 1px solid var(--border, #e5e7eb);
		border-radius: var(--radius-md, 0.375rem);
		padding: var(--space-sm, 0.5rem);
		background-color: var(--bg-secondary, #f9fafb);
	}

	.metadata-summary {
		cursor: pointer;
		font-weight: 600;
		color: var(--text-muted, #6b7280);
		user-select: none;
		font-size: 0.875rem;
	}

	.metadata-summary:hover {
		color: var(--text, #111827);
	}

	.metadata-content {
		margin-top: var(--space-sm, 0.5rem);
		padding: var(--space-sm, 0.5rem);
		background: var(--bg-primary, #ffffff);
		border-radius: var(--radius-sm, 0.25rem);
	}

	.metadata-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-xs, 0.25rem);
		font-size: 0.8125rem;
	}

	.metadata-key {
		color: var(--text-muted, #6b7280);
		font-weight: 500;
	}

	.metadata-value {
		color: var(--text, #111827);
		font-family: var(--font-mono, monospace);
		word-break: break-all;
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
			display: block;
			overflow-x: auto;
		}

		.customers-table th,
		.customers-table td {
			display: block;
			width: 100%;
			text-align: right;
			padding-left: 50%;
			position: relative;
		}

		.customers-table th:before,
		.customers-table td:before {
			content: attr(data-label);
			position: absolute;
			left: 0;
			width: 50%;
			padding-left: var(--space-md);
			font-weight: 600;
			text-align: left;
			color: var(--text-muted);
		}

		.customer-info {
			flex-direction: column;
			align-items: flex-start;
			text-align: left;
		}

		.customer-name h4 {
			margin-bottom: var(--space-xs);
		}

		.customer-actions {
			justify-content: flex-start;
			flex-wrap: wrap;
			gap: var(--space-xs);
		}

		.btn-sm {
			flex: 1;
			min-width: 100px;
		}
	}
</style> 