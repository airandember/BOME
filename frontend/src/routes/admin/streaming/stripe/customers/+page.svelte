<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';

	let summary: any = null;
	let loading = true;
	let error = '';

	export let data: any = null;

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

	function getCustomerSubscriptions(customerId: string) {
		if (!summary.subscriptions) return [];
		return summary.subscriptions.filter((sub: any) => sub.CustomerID === customerId);
	}

	function getCustomerPayments(customerId: string) {
		if (!summary.payment_intents) return [];
		return summary.payment_intents.filter((pi: any) => pi.CustomerID === customerId);
	}
</script>

{#if loading}
	<div class="loading">
		<div class="spinner"></div>
		<p>Loading customers...</p>
	</div>
{:else if error}
	<div class="error-state">
		<h3>Error Loading Customers</h3>
		<p>{error}</p>
		<button class="btn btn-primary" on:click={fetchSummary}>Retry</button>
	</div>
{:else if summary}
	<div class="customers-container">
		<div class="customers-header">
			<h1>Customers</h1>
			<p>Manage your Stripe customers and their information</p>
			<div class="stats-summary">
				<div class="stat">
					<span class="stat-value">{summary.customers_count || 0}</span>
					<span class="stat-label">Total Customers</span>
				</div>
				<div class="stat">
					<span class="stat-value">{summary.subscriptions?.length || 0}</span>
					<span class="stat-label">With Subscriptions</span>
				</div>
				<div class="stat">
					<span class="stat-value">{summary.payment_intents?.filter((pi: any) => pi.Status === 'succeeded').length || 0}</span>
					<span class="stat-label">Successful Payments</span>
				</div>
			</div>
		</div>

		{#if summary.customers && summary.customers.length > 0}
			<div class="customers-grid">
				{#each summary.customers as customer}
					<div class="customer-card">
						<div class="customer-header">
							<div class="customer-info">
								<h3 class="customer-name">{customer.Name || 'No Name'}</h3>
								<div class="customer-email">{customer.Email}</div>
								<div class="customer-id">ID: {customer.ID}</div>
							</div>
							<div class="customer-avatar">
								{customer.Name ? customer.Name.charAt(0).toUpperCase() : '?'}
							</div>
						</div>

						<div class="customer-metadata">
							<div class="metadata-item">
								<span class="metadata-label">Created:</span>
								<span class="metadata-value">{formatDate(customer.CreatedAt)}</span>
							</div>
							
							{#if customer.Metadata && Object.keys(customer.Metadata).length > 0}
								<div class="metadata-item">
									<span class="metadata-label">Custom Data:</span>
									<div class="metadata-tags">
										{#each Object.entries(customer.Metadata) as [key, value]}
											<span class="metadata-tag">{key}: {value}</span>
										{/each}
									</div>
								</div>
							{/if}
						</div>

						<!-- Customer Subscriptions -->
						{#if summary.subscriptions}
							{@const customerSubs = summary.subscriptions.filter((sub: any) => sub.CustomerID === customer.ID)}
							{#if customerSubs.length > 0}
								<div class="customer-subscriptions">
									<h4>Subscriptions ({customerSubs.length})</h4>
									<div class="subscriptions-list">
										{#each customerSubs as subscription}
											<div class="subscription-item">
												<div class="subscription-status">
													<span class="status-indicator {subscription.Status}"></span>
													<span class="status-text">{subscription.Status}</span>
												</div>
												<div class="subscription-details">
													<div class="subscription-id">{subscription.ID.slice(-8)}</div>
													<div class="subscription-period">
														Next: {formatDate(subscription.CurrentPeriodEnd)}
													</div>
													{#if subscription.CancelAtPeriodEnd}
														<div class="cancellation-notice">Cancels at period end</div>
													{/if}
												</div>
											</div>
										{/each}
									</div>
								</div>
							{/if}
						{/if}

						<!-- Recent Payments -->
						{#if summary.payment_intents}
							{@const customerPayments = summary.payment_intents.filter((pi: any) => pi.CustomerID === customer.ID).slice(0, 3)}
							{#if customerPayments.length > 0}
								<div class="customer-payments">
									<h4>Recent Payments ({customerPayments.length})</h4>
									<div class="payments-list">
										{#each customerPayments as payment}
											<div class="payment-item">
												<div class="payment-amount">
													${(payment.Amount / 100).toFixed(2)} {payment.Currency.toUpperCase()}
												</div>
												<div class="payment-details">
													<span class="payment-status {payment.Status}">{payment.Status}</span>
													<span class="payment-date">{formatDate(payment.CreatedAt)}</span>
												</div>
											</div>
										{/each}
									</div>
								</div>
							{/if}
						{/if}

						<div class="customer-actions">
							<button class="btn btn-outline btn-sm">View Details</button>
							<button class="btn btn-outline btn-sm">Edit Customer</button>
							<button class="btn btn-outline btn-sm">View Invoices</button>
						</div>
					</div>
				{/each}
			</div>
		{:else}
			<div class="empty-state">
				<div class="empty-icon">👥</div>
				<h3>No Customers Found</h3>
				<p>You haven't acquired any customers yet. Customers are created when they make their first purchase.</p>
				<button class="btn btn-primary">Create Test Customer</button>
			</div>
		{/if}
	</div>
{/if}

<style>
	.customers-container {
		padding: var(--space-lg);
	}

	.customers-header {
		margin-bottom: var(--space-xl);
	}

	.customers-header h1 {
		margin: 0 0 var(--space-sm) 0;
		color: var(--text);
		font-size: 2rem;
	}

	.customers-header p {
		margin: 0 0 var(--space-lg) 0;
		color: var(--text-muted);
		font-size: 1.1rem;
	}

	.stats-summary {
		display: flex;
		gap: var(--space-lg);
		flex-wrap: wrap;
	}

	.stat {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: var(--space-md);
		background: var(--surface);
		border-radius: var(--radius-lg);
		border: 1px solid var(--border);
		min-width: 120px;
	}

	.stat-value {
		font-size: 1.8rem;
		font-weight: bold;
		color: var(--primary);
		margin-bottom: var(--space-xs);
	}

	.stat-label {
		font-size: 0.9rem;
		color: var(--text-muted);
	}

	.customers-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
		gap: var(--space-lg);
	}

	.customer-card {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
		transition: all 0.2s ease;
	}

	.customer-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
	}

	.customer-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: var(--space-md);
	}

	.customer-info {
		flex: 1;
	}

	.customer-name {
		margin: 0 0 var(--space-xs) 0;
		color: var(--text);
		font-size: 1.3rem;
		font-weight: 600;
	}

	.customer-email {
		color: var(--primary);
		font-size: 1rem;
		margin-bottom: var(--space-xs);
	}

	.customer-id {
		font-family: monospace;
		font-size: 0.9rem;
		color: var(--text-muted);
	}

	.customer-avatar {
		width: 50px;
		height: 50px;
		border-radius: 50%;
		background: var(--primary);
		color: white;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 1.5rem;
		font-weight: bold;
	}

	.customer-metadata {
		margin-bottom: var(--space-md);
	}

	.metadata-item {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
		margin-bottom: var(--space-md);
	}

	.metadata-label {
		font-size: 0.9rem;
		font-weight: 600;
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.metadata-value {
		color: var(--text);
	}

	.metadata-tags {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-xs);
	}

	.metadata-tag {
		background: var(--bg-secondary);
		color: var(--text-muted);
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-sm);
		font-size: 0.8rem;
		font-family: monospace;
	}

	.customer-subscriptions {
		margin-bottom: var(--space-md);
		padding: var(--space-md);
		background: var(--bg-secondary);
		border-radius: var(--radius-md);
	}

	.customer-subscriptions h4 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
		font-size: 1rem;
	}

	.subscriptions-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}

	.subscription-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-sm);
		background: var(--surface);
		border-radius: var(--radius-md);
		border: 1px solid var(--border);
	}

	.subscription-status {
		display: flex;
		align-items: center;
		gap: var(--space-xs);
	}

	.status-indicator {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--text-muted);
	}

	.status-indicator.active {
		background: var(--success);
	}

	.status-indicator.past_due {
		background: var(--warning);
	}

	.status-indicator.canceled {
		background: var(--error);
	}

	.status-text {
		font-weight: 600;
		text-transform: capitalize;
	}

	.subscription-details {
		text-align: right;
		font-size: 0.9rem;
	}

	.subscription-id {
		font-family: monospace;
		color: var(--text-muted);
	}

	.subscription-period {
		color: var(--text);
	}

	.cancellation-notice {
		color: var(--warning);
		font-size: 0.8rem;
		font-weight: bold;
	}

	.customer-payments {
		margin-bottom: var(--space-md);
		padding: var(--space-md);
		background: var(--bg-secondary);
		border-radius: var(--radius-md);
	}

	.customer-payments h4 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
		font-size: 1rem;
	}

	.payments-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}

	.payment-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-sm);
		background: var(--surface);
		border-radius: var(--radius-md);
		border: 1px solid var(--border);
	}

	.payment-amount {
		font-weight: 600;
		color: var(--primary);
		font-size: 1.1rem;
	}

	.payment-details {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: var(--space-xs);
		font-size: 0.9rem;
	}

	.payment-status {
		padding: 2px 8px;
		border-radius: var(--radius-sm);
		font-size: 0.7rem;
		font-weight: bold;
		text-transform: uppercase;
	}

	.payment-status.succeeded {
		background: var(--success);
		color: white;
	}

	.payment-status.canceled {
		background: var(--error);
		color: white;
	}

	.payment-status.processing {
		background: var(--warning);
		color: white;
	}

	.payment-date {
		color: var(--text-muted);
	}

	.customer-actions {
		display: flex;
		gap: var(--space-sm);
		flex-wrap: wrap;
	}

	.btn {
		padding: var(--space-sm) var(--space-md);
		border: none;
		border-radius: var(--radius-md);
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		display: inline-block;
		text-align: center;
	}

	.btn-sm {
		padding: var(--space-xs) var(--space-sm);
		font-size: 0.8rem;
	}

	.btn-primary {
		background: var(--primary);
		color: white;
	}

	.btn-primary:hover {
		background: var(--primary-dark);
	}

	.btn-outline {
		background: transparent;
		color: var(--text);
		border: 1px solid var(--border);
	}

	.btn-outline:hover {
		background: var(--surface-hover);
	}

	.empty-state {
		text-align: center;
		padding: var(--space-xl) var(--space-lg);
		background: var(--surface);
		border: 2px dashed var(--border);
		border-radius: var(--radius-lg);
	}

	.empty-icon {
		font-size: 4rem;
		margin-bottom: var(--space-lg);
		opacity: 0.5;
	}

	.empty-state h3 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
	}

	.empty-state p {
		margin: 0 0 var(--space-lg) 0;
		color: var(--text-muted);
	}

	.loading {
		text-align: center;
		padding: var(--space-xl);
	}

	.spinner {
		width: 40px;
		height: 40px;
		border: 4px solid var(--border);
		border-top: 4px solid var(--primary);
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin: 0 auto var(--space-md);
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.error-state {
		text-align: center;
		padding: var(--space-xl);
	}

	.error-state h3 {
		color: var(--error);
		margin-bottom: var(--space-md);
	}

	@media (max-width: 768px) {
		.customers-grid {
			grid-template-columns: 1fr;
		}

		.stats-summary {
			justify-content: center;
		}

		.customer-actions {
			justify-content: center;
		}

		.customer-header {
			flex-direction: column;
			align-items: center;
			text-align: center;
			gap: var(--space-md);
		}
	}
</style> 