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
							${(summary.invoices.filter(inv => inv.Status === 'paid').reduce((sum, inv) => sum + inv.Amount, 0) / 100).toFixed(2)}
						{:else}
							$0.00
						{/if}
					</div>
				</div>
				
				<div class="stat-item">
					<div class="stat-label">Active Subscriptions</div>
					<div class="stat-value">
						{summary.subscriptions?.filter(sub => sub.Status === 'active').length || 0}
					</div>
				</div>
				
				<div class="stat-item">
					<div class="stat-label">Successful Payments</div>
					<div class="stat-value">
						{summary.payment_intents?.filter(pi => pi.Status === 'succeeded').length || 0}
					</div>
				</div>
				
				<div class="stat-item">
					<div class="stat-label">Total Customers</div>
					<div class="stat-value">{summary.customers_count || 0}</div>
				</div>
			</div>
		</div>

		<!-- System Capabilities -->
		<section class="capabilities-section">
			<h2>System Capabilities</h2>
			<div class="capabilities-grid">
				{#each Object.entries(summary.capabilities || {}) as [category, capabilities]}
					<div class="capability-category">
						<h4>{category.charAt(0).toUpperCase() + category.slice(1)}</h4>
						<div class="capability-list">
							{#each Object.entries(capabilities as any) as [operation, enabled]}
								<div class="capability-item {enabled ? 'enabled' : 'disabled'}">
									<span class="capability-icon">{enabled ? '✅' : '❌'}</span>
									<span class="capability-name">{operation}</span>
								</div>
							{/each}
						</div>
					</div>
				{/each}
			</div>
		</section>
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

	.capabilities-section {
		margin-bottom: var(--space-xl);
	}

	.capabilities-section h2 {
		margin-bottom: var(--space-lg);
		color: var(--text);
		font-size: 1.5rem;
	}

	.capabilities-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: var(--space-lg);
	}

	.capability-category h4 {
		margin-bottom: var(--space-md);
		color: var(--text);
		text-transform: capitalize;
	}

	.capability-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}

	.capability-item {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		padding: var(--space-sm);
		border-radius: var(--radius-md);
		background: var(--surface);
		border: 1px solid var(--border);
	}

	.capability-item.enabled {
		border-color: var(--success);
		background: var(--success-light);
	}

	.capability-item.disabled {
		border-color: var(--error);
		background: var(--error-light);
	}

	.capability-icon {
		font-size: 1.2rem;
	}

	.capability-name {
		text-transform: capitalize;
		font-weight: 500;
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

		.environment-banner {
			flex-direction: column;
			gap: var(--space-md);
		}
	}
</style> 