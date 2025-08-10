<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';

	// Import child components
	import Overview from './overview/+page.svelte';
	import Products from './products/+page.svelte';
	import Customers from './customers/+page.svelte';
	import Setup from './setup/+page.svelte';

	let summary: any = null;
	let loading = true;
	let error = '';
	let activeTab = 'overview';

	// Tab configuration
	const tabs = [
		{ id: 'overview', name: 'Overview', icon: '📊', component: Overview },
		{ id: 'products', name: 'Products', icon: '📦', component: Products },
		{ id: 'customers', name: 'Customers', icon: '👥', component: Customers },
		{ id: 'invoices', name: 'Invoices', icon: '📄', component: null },
		{ id: 'payments', name: 'Payments', icon: '💳', component: null },
		{ id: 'subscriptions', name: 'Subscriptions', icon: '🔄', component: null },
		{ id: 'setup', name: 'Setup', icon: '⚙️', component: Setup }
	];

	onMount(async () => {
		await fetchSummary();
		
		// If not enabled, default to setup tab
		if (summary && !summary.enabled) {
			activeTab = 'setup';
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
				error = 'Failed to load Stripe data';
			}
		} catch (err) {
			error = 'Failed to load Stripe data';
			console.error(err);
		} finally {
			loading = false;
		}
	}

	function switchTab(tabId: string) {
		activeTab = tabId;
	}

	$: activeTabConfig = tabs.find(tab => tab.id === activeTab);
</script>

{#if loading}
	<div class="loading-container">
		<div class="spinner"></div>
		<p>Loading Stripe Dashboard...</p>
	</div>
{:else if error}
	<div class="error-container">
		<h3>Error Loading Stripe Dashboard</h3>
		<p>{error}</p>
		<button class="btn btn-primary" on:click={fetchSummary}>Retry</button>
	</div>
{:else}
	<div class="stripe-dashboard">
		<!-- Header -->
		<div class="dashboard-header">
			<div class="header-content">
				<h1>Stripe Dashboard</h1>
				<p>Manage payments, subscriptions, and customer data</p>
			</div>
			
			{#if summary?.enabled}
				<div class="header-status">
					<div class="status-indicator connected">
						<div class="status-dot"></div>
						<span>Connected</span>
					</div>
					<div class="environment-badge {summary.environment === 'live' ? 'live' : 'test'}">
						{summary.environment === 'live' ? '🔴 LIVE' : '🟡 TEST'}
					</div>
				</div>
			{:else}
				<div class="header-status">
					<div class="status-indicator disconnected">
						<div class="status-dot"></div>
						<span>Not Connected</span>
					</div>
				</div>
			{/if}
		</div>

		<!-- Tab Navigation -->
		<div class="tab-navigation">
			<div class="tab-list">
				{#each tabs as tab}
					<button 
						class="tab-button {activeTab === tab.id ? 'active' : ''}"
						class:disabled={!summary?.enabled && tab.id !== 'setup'}
						on:click={() => switchTab(tab.id)}
						disabled={!summary?.enabled && tab.id !== 'setup'}
					>
						<span class="tab-icon">{tab.icon}</span>
						<span class="tab-name">{tab.name}</span>
						{#if !summary?.enabled && tab.id !== 'setup'}
							<span class="tab-lock">🔒</span>
						{/if}
					</button>
				{/each}
			</div>
		</div>

		<!-- Tab Content -->
		<div class="tab-content">
			{#if activeTab === 'overview' && activeTabConfig?.component}
				<svelte:component this={activeTabConfig.component} data={summary} />
			{:else if activeTab === 'products' && activeTabConfig?.component}
				<svelte:component this={activeTabConfig.component} data={summary} />
			{:else if activeTab === 'customers' && activeTabConfig?.component}
				<svelte:component this={activeTabConfig.component} data={summary} />
			{:else if activeTab === 'setup' && activeTabConfig?.component}
				<svelte:component this={activeTabConfig.component} data={summary} />
			{:else if activeTab === 'invoices'}
				<div class="coming-soon">
					<div class="coming-soon-icon">📄</div>
					<h3>Invoices Coming Soon</h3>
					<p>Invoice management features are currently in development.</p>
					{#if summary?.invoices && summary.invoices.length > 0}
						<div class="preview-stats">
							<div class="stat">
								<span class="stat-value">{summary.invoices_count}</span>
								<span class="stat-label">Total Invoices</span>
							</div>
							<div class="stat">
								<span class="stat-value">{summary.invoices.filter((inv: any) => inv.Status === 'paid').length}</span>
								<span class="stat-label">Paid</span>
							</div>
						</div>
					{/if}
				</div>
			{:else if activeTab === 'payments'}
				<div class="coming-soon">
					<div class="coming-soon-icon">💳</div>
					<h3>Payments Coming Soon</h3>
					<p>Payment management features are currently in development.</p>
					{#if summary?.payment_intents && summary.payment_intents.length > 0}
						<div class="preview-stats">
							<div class="stat">
								<span class="stat-value">{summary.payment_intents_count}</span>
								<span class="stat-label">Total Payments</span>
							</div>
							<div class="stat">
								<span class="stat-value">{summary.payment_intents.filter((pi: any) => pi.Status === 'succeeded').length}</span>
								<span class="stat-label">Successful</span>
							</div>
						</div>
					{/if}
				</div>
			{:else if activeTab === 'subscriptions'}
				<div class="coming-soon">
					<div class="coming-soon-icon">🔄</div>
					<h3>Subscriptions Coming Soon</h3>
					<p>Subscription management features are currently in development.</p>
					{#if summary?.subscriptions && summary.subscriptions.length > 0}
						<div class="preview-stats">
							<div class="stat">
								<span class="stat-value">{summary.subscriptions_count}</span>
								<span class="stat-label">Total Subscriptions</span>
							</div>
							<div class="stat">
								<span class="stat-value">{summary.subscriptions.filter((sub: any) => sub.Status === 'active').length}</span>
								<span class="stat-label">Active</span>
							</div>
						</div>
					{/if}
				</div>
			{/if}
		</div>
	</div>
{/if}

<style>
	.stripe-dashboard {
		min-height: 100vh;
		background: var(--bg-primary);
	}

	.loading-container,
	.error-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--space-xl);
		min-height: 50vh;
	}

	.spinner {
		width: 40px;
		height: 40px;
		border: 4px solid var(--border);
		border-top: 4px solid var(--primary);
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-bottom: var(--space-md);
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.error-container h3 {
		color: var(--error);
		margin-bottom: var(--space-md);
	}

	.btn {
		padding: var(--space-md) var(--space-lg);
		border: none;
		border-radius: var(--radius-md);
		font-size: 1rem;
		cursor: pointer;
		transition: all 0.2s ease;
		font-weight: 600;
	}

	.btn-primary {
		background: var(--primary);
		color: white;
	}

	.btn-primary:hover {
		background: var(--primary-dark);
		transform: translateY(-1px);
	}

	.dashboard-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-xl) var(--space-lg);
		background: var(--surface);
		border-bottom: 1px solid var(--border);
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

	.header-status {
		display: flex;
		align-items: center;
		gap: var(--space-md);
	}

	.status-indicator {
		display: flex;
		align-items: center;
		gap: var(--space-xs);
		padding: var(--space-xs) var(--space-md);
		border-radius: var(--radius-md);
		font-size: 0.9rem;
		font-weight: 600;
	}

	.status-indicator.connected {
		background: var(--success-light);
		color: var(--success-dark);
	}

	.status-indicator.disconnected {
		background: var(--error-light);
		color: var(--error-dark);
	}

	.status-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: currentColor;
	}

	.environment-badge {
		padding: var(--space-xs) var(--space-md);
		border-radius: var(--radius-md);
		font-size: 0.8rem;
		font-weight: bold;
	}

	.environment-badge.test {
		background: var(--warning);
		color: white;
	}

	.environment-badge.live {
		background: var(--error);
		color: white;
	}

	.tab-navigation {
		background: var(--surface);
		border-bottom: 1px solid var(--border);
		padding: 0 var(--space-lg);
	}

	.tab-list {
		display: flex;
		gap: var(--space-xs);
		overflow-x: auto;
		scrollbar-width: none;
		-ms-overflow-style: none;
	}

	.tab-list::-webkit-scrollbar {
		display: none;
	}

	.tab-button {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		padding: var(--space-md) var(--space-lg);
		border: none;
		background: transparent;
		color: var(--text-muted);
		cursor: pointer;
		border-bottom: 3px solid transparent;
		transition: all 0.2s ease;
		white-space: nowrap;
		font-size: 0.95rem;
		font-weight: 500;
	}

	.tab-button:hover:not(:disabled) {
		color: var(--text);
		background: var(--bg-secondary);
	}

	.tab-button.active {
		color: var(--primary);
		border-bottom-color: var(--primary);
		background: var(--primary-light);
	}

	.tab-button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.tab-icon {
		font-size: 1.1rem;
	}

	.tab-name {
		font-weight: inherit;
	}

	.tab-lock {
		font-size: 0.8rem;
		opacity: 0.7;
	}

	.tab-content {
		flex: 1;
		min-height: calc(100vh - 200px);
		background: var(--bg-primary);
	}

	.coming-soon {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--space-xl);
		text-align: center;
		min-height: 400px;
	}

	.coming-soon-icon {
		font-size: 4rem;
		margin-bottom: var(--space-lg);
		opacity: 0.5;
	}

	.coming-soon h3 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
		font-size: 1.5rem;
	}

	.coming-soon p {
		margin: 0 0 var(--space-lg) 0;
		color: var(--text-muted);
		font-size: 1.1rem;
	}

	.preview-stats {
		display: flex;
		gap: var(--space-lg);
		justify-content: center;
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
		font-size: 2rem;
		font-weight: bold;
		color: var(--primary);
		margin-bottom: var(--space-xs);
	}

	.stat-label {
		font-size: 0.9rem;
		color: var(--text-muted);
	}

	@media (max-width: 768px) {
		.dashboard-header {
			flex-direction: column;
			gap: var(--space-md);
			text-align: center;
		}

		.header-content h1 {
			font-size: 1.5rem;
		}

		.tab-list {
			justify-content: flex-start;
		}

		.tab-button {
			padding: var(--space-sm) var(--space-md);
			font-size: 0.9rem;
		}

		.preview-stats {
			flex-direction: column;
			align-items: center;
		}
	}
</style> 