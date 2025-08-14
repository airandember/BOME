<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';
	import CustomerSyncPanel from '../components/CustomerSyncPanel.svelte';

	let summary: any = null;
	let loading = true;
	let error = '';

	export let data: any = null;

	onMount(async () => {
		if (data) {
			summary = data;
			loading = false;
		} else {
			// Only fetch if no data is passed (standalone mode)
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
				error = 'Failed to load summary';
			}
		} catch (err) {
			error = 'Failed to load summary';
			console.error(err);
		} finally {
			loading = false;
		}
	}

	async function refreshData() {
		await fetchSummary();
	}

	// Currency formatting utility
		function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(amount);
	}

	function getCategoryIcon(category: string): string {
		switch (category.toLowerCase()) {
			case 'products':
				return '📦';
			case 'prices':
				return '💰';
			case 'customers':
				return '👥';
			case 'subscriptions':
				return '🔄';
			case 'payment_intents':
				return '💳';
			case 'invoices':
				return '📄';
			case 'webhooks':
				return '🔗';
			case 'coupons':
				return '🎟️';
			default:
				return '⚙️';
		}
	}

	function formatOperationName(operation: string): string {
		return operation.charAt(0).toUpperCase() + operation.slice(1).replace(/_/g, ' ');
	}
</script>

{#if loading}
	<div class="loading">
		<div class="spinner"></div>
		<p>Loading overview...</p>
	</div>
{:else if error}
	<div class="error-state">
		<h3>Error Loading Overview</h3>
		<p>{error}</p>
		<button class="btn btn-primary" on:click={refreshData}>Retry</button>
	</div>
{:else if summary}
	<div class="overview-container">
		<!-- Account Overview Cards -->
		<div class="overview-section">
			<h2>Account Overview</h2>
			<div class="overview-grid">
				<div class="overview-card">
					<div class="card-icon">📊</div>
					<div class="card-content">
						<h3>Products</h3>
						<div class="card-value">{summary.products_count || 0}</div>
						<div class="card-label">Total Products</div>
					</div>
				</div>
				
				<div class="overview-card">
					<div class="card-icon">💰</div>
					<div class="card-content">
						<h3>Prices</h3>
						<div class="card-value">{summary.prices_count || 0}</div>
						<div class="card-label">Active Prices</div>
					</div>
				</div>
				
				<div class="overview-card">
					<div class="card-icon">👥</div>
					<div class="card-content">
						<h3>Customers</h3>
						<div class="card-value">{summary.customers_count || 0}</div>
						<div class="card-label">Total Customers</div>
					</div>
				</div>
				
				<div class="overview-card">
					<div class="card-icon">🔄</div>
					<div class="card-content">
						<h3>Subscriptions</h3>
						<div class="card-value">{summary.subscriptions_count || 0}</div>
						<div class="card-label">Active Subscriptions</div>
					</div>
				</div>
				
				<div class="overview-card">
					<div class="card-icon">💳</div>
					<div class="card-content">
						<h3>Payments</h3>
						<div class="card-value">{summary.payment_intents_count || 0}</div>
						<div class="card-label">Recent Payments</div>
					</div>
				</div>
				
				<div class="overview-card">
					<div class="card-icon">📄</div>
					<div class="card-content">
						<h3>Invoices</h3>
						<div class="card-value">{summary.invoices_count || 0}</div>
						<div class="card-label">Recent Invoices</div>
					</div>
				</div>
				
				<div class="overview-card">
					<div class="card-icon">🎟️</div>
					<div class="card-content">
						<h3>Coupons</h3>
						<div class="card-value">{summary.coupons_count || 0}</div>
						<div class="card-label">Active Coupons</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Environment Info -->
		<div class="environment-banner">
			<div class="environment-indicator {summary.environment === 'live' ? 'live' : 'test'}">
				{summary.environment === 'live' ? '🔴 LIVE' : '🟡 TEST'} Environment
			</div>
			<button class="btn btn-secondary" on:click={refreshData}>
				🔄 Refresh Data
			</button>
		</div>

		<!-- Quick Stats -->
		<div class="quick-stats">
			<h2>Quick Statistics</h2>
			<div class="stats-grid">
				<div class="stat-item">
					<div class="stat-label">Revenue This Month</div>
					<div class="stat-value">
						{#if summary.invoices}
							{formatCurrency(summary.invoices.filter((inv: any) => inv.Status === 'paid').reduce((sum: any, inv: any) => sum + inv.Amount, 0) / 100)}
						{:else}
							{formatCurrency(0)}
						{/if}
					</div>
				</div>
				
				<div class="stat-item">
					<div class="stat-label">Active Subscriptions</div>
					<div class="stat-value">
						{summary.subscriptions?.filter((sub: any) => sub.Status === 'active').length || 0}
					</div>
				</div>
				
				<div class="stat-item">
					<div class="stat-label">Successful Payments</div>
					<div class="stat-value">
						{summary.payment_intents?.filter((pi: any) => pi.Status === 'succeeded').length || 0}
					</div>
				</div>
				
				<div class="stat-item">
					<div class="stat-label">Total Customers</div>
					<div class="stat-value">{summary.customers_count || 0}</div>
				</div>
				
				<div class="stat-item">
					<div class="stat-label">Active Coupons</div>
					<div class="stat-value">
						{summary.coupons?.filter((c: any) => c.Valid).length || 0}
					</div>
				</div>
			</div>
		</div>

		<!-- System Capabilities -->
		<section class="overview-section">
			<h2>🔧 System Capabilities</h2>
			<div class="capabilities-container">
				{#each Object.entries(summary.capabilities || {}) as [category, capabilities]}
					<div class="capability-card">
						<div class="capability-header">
							<h3>{getCategoryIcon(category)} {category.charAt(0).toUpperCase() + category.slice(1)}</h3>
							<div class="capability-status">
								{Object.values(capabilities as any).filter(Boolean).length}/{Object.keys(capabilities as any).length} enabled
							</div>
						</div>
						<div class="capability-items">
							{#each Object.entries(capabilities as any) as [operation, enabled]}
								<div class="capability-item {enabled ? 'enabled' : 'disabled'}">
									<span class="capability-icon">{enabled ? '✅' : '❌'}</span>
									<span class="capability-name">{formatOperationName(operation)}</span>
									<span class="capability-badge {enabled ? 'success' : 'error'}">
										{enabled ? 'Enabled' : 'Disabled'}
									</span>
								</div>
							{/each}
						</div>
					</div>
				{/each}
			</div>
		</section>

		<!-- Customer Sync Panel 
		<section class="customer-sync-section">
			<CustomerSyncPanel customerId={null} customerEmail="" />
		</section>-->
	</div>
{/if}

<style>
	.overview-container {
		padding: var(--space-lg);
	}

	.overview-section {
		margin-bottom: var(--space-xl);
	}

	.overview-section h2 {
		margin-bottom: var(--space-lg);
		color: var(--text);
		font-size: 1.5rem;
	}

	.overview-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: var(--space-md);
	}

	.overview-card {
		background: var(--surface);
		padding: var(--space-lg);
		border-radius: var(--radius-lg);
		border: 1px solid var(--border);
		display: flex;
		align-items: center;
		gap: var(--space-md);
		transition: all 0.2s ease;
	}

	.overview-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
	}

	.card-icon {
		font-size: 2rem;
		opacity: 0.8;
	}

	.card-content h3 {
		margin: 0 0 var(--space-xs) 0;
		font-size: 0.9rem;
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.card-value {
		font-size: 2rem;
		font-weight: bold;
		color: var(--primary);
		margin-bottom: var(--space-xs);
	}

	.card-label {
		font-size: 0.8rem;
		color: var(--text-muted);
	}

	.environment-banner {
		display: flex;
		justify-content: space-between;
		align-items: center;
		background: var(--surface);
		padding: var(--space-md) var(--space-lg);
		border-radius: var(--radius-lg);
		margin-bottom: var(--space-xl);
		border: 1px solid var(--border);
	}

	.environment-indicator {
		padding: var(--space-xs) var(--space-md);
		border-radius: var(--radius-md);
		font-weight: bold;
		font-size: 0.9rem;
	}

	.environment-indicator.test {
		background: var(--warning);
		color: white;
	}

	.environment-indicator.live {
		background: var(--success);
		color: white;
	}

	.btn {
		padding: var(--space-md) var(--space-lg);
		border: none;
		border-radius: var(--radius-md);
		font-size: 1rem;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.btn-secondary {
		background: var(--surface);
		color: var(--text);
		border: 1px solid var(--border);
	}

	.btn-secondary:hover {
		background: var(--surface-hover);
	}

	.btn-primary {
		background: var(--primary);
		color: white;
	}

	.btn-primary:hover {
		background: var(--primary-dark);
	}

	.quick-stats {
		margin-bottom: var(--space-xl);
	}

	.quick-stats h2 {
		margin-bottom: var(--space-lg);
		color: var(--text);
		font-size: 1.5rem;
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: var(--space-md);
	}

	.stat-item {
		background: var(--surface);
		padding: var(--space-lg);
		border-radius: var(--radius-lg);
		border: 1px solid var(--border);
		text-align: center;
	}

	.stat-label {
		font-size: 0.9rem;
		color: var(--text-muted);
		margin-bottom: var(--space-sm);
	}

	.stat-value {
		font-size: 1.8rem;
		font-weight: bold;
		color: var(--primary);
	}

	.capabilities-container {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
		gap: var(--space-lg);
	}

	.capability-card {
		background: var(--surface);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
		border: 1px solid var(--border);
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
		transition: all 0.2s ease;
	}

	.capability-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
	}

	.capability-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-lg);
		padding-bottom: var(--space-md);
		border-bottom: 1px solid var(--border);
	}

	.capability-header h3 {
		margin: 0;
		color: var(--text);
		font-size: 1.1rem;
		font-weight: 600;
		display: flex;
		align-items: center;
		gap: var(--space-sm);
	}

	.capability-status {
		font-size: 0.875rem;
		color: var(--text-muted);
		font-weight: 500;
		padding: var(--space-xs) var(--space-sm);
		background: var(--bg-secondary);
		border-radius: var(--radius-md);
	}

	.capability-items {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}

	.capability-item {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: var(--space-md);
		border-radius: var(--radius-md);
		background: var(--bg-secondary);
		border: 1px solid var(--border);
		transition: all 0.2s ease;
	}

	.capability-item:hover {
		background: var(--surface);
		transform: translateX(2px);
	}

	.capability-item.enabled {
		border-left: 4px solid var(--success);
	}

	.capability-item.disabled {
		border-left: 4px solid var(--error);
		opacity: 0.7;
	}

	.capability-icon {
		font-size: 1.1rem;
		margin-right: var(--space-sm);
	}

	.capability-name {
		flex: 1;
		font-weight: 500;
		color: var(--text);
		font-size: 0.875rem;
	}

	.capability-badge {
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-sm);
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.capability-badge.success {
		background: var(--success-light);
		color: var(--success-dark);
	}

	.capability-badge.error {
		background: var(--error-light);
		color: var(--error-dark);
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
		.overview-grid {
			grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		}

		.stats-grid {
			grid-template-columns: 1fr;
		}

		.capabilities-container {
			grid-template-columns: 1fr;
		}

		.capability-header {
			flex-direction: column;
			align-items: flex-start;
			gap: var(--space-sm);
		}

		.capability-item {
			flex-direction: column;
			align-items: flex-start;
			gap: var(--space-sm);
		}

		.capability-name {
			margin-bottom: var(--space-xs);
		}

		.environment-banner {
			flex-direction: column;
			gap: var(--space-md);
		}
	}
</style> 