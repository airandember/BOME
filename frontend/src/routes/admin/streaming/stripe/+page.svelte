<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';

	let secret = '';
	let saving = false;
	let summary: any = null;
	let loading = true;
	let error = '';
	let success = '';

	onMount(async () => {
		await fetchSummary();
	});

	async function fetchSummary() {
		console.log("Fetching summary");
		try {
			error = '';
			const res = await apiRequest('/admin/streaming/stripe/summary');
			console.log(res);
			if (res.ok) {
				loading = false;
				const data = await res.json();
				summary = data.summary;
				console.log("Summary loaded", summary);
			} else {
				error = 'Failed to load summary';
				console.log("Failed to load summary", res);
				loading = false;
			}
		} catch (err) {
			error = 'Failed to load summary';
			console.error(err);
		} finally {
			loading = false;
		}
	}

	async function saveSecret() {
		if (!secret.trim()) return;
		
		saving = true;
		error = '';
		success = '';
		
		try {
			const res = await apiRequest('/admin/streaming/stripe/secret', {
				method: 'POST',
				body: JSON.stringify({ key: secret })
			});
			
			if (res.ok) {
				success = 'Stripe key saved successfully!';
				secret = '';
				await fetchSummary(); // Refresh the summary
			} else {
				const errorData = await res.json();
				error = errorData.error || 'Failed to save key';
			}
		} catch (err) {
			error = 'Failed to save key';
			console.error(err);
		} finally {
			saving = false;
		}
	}

	function formatCurrency(amount: number, currency: string): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: currency.toUpperCase(),
			minimumFractionDigits: 2
		}).format(amount / 100);
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

	function getStatusColor(status: string): string {
		switch (status.toLowerCase()) {
			case 'active':
			case 'succeeded':
			case 'paid':
				return 'var(--success)';
			case 'inactive':
			case 'cancelled':
			case 'failed':
				return 'var(--error)';
			case 'past_due':
			case 'unpaid':
				return 'var(--warning)';
			default:
				return 'var(--text-muted)';
		}
	}
</script>

{#if loading}
	<div class="loading">
		<div class="spinner"></div>
		<p>Loading Stripe data...</p>
	</div>

{:else}
	<div class="stripe-dashboard">
		<header class="dashboard-header">
			<h1>Stripe Dashboard</h1>
			<p>Manage your Stripe integration and view account analytics</p>
		</header>

		{#if !summary?.enabled}
			<div class="setup-section">
				<h2>Setup Stripe Integration</h2>
				<p>Enter your Stripe secret key to connect your account and start managing payments.</p>
				
				<form on:submit|preventDefault={saveSecret}>
					<div class="input-group">
						<input 
							class="input" 
							type="password" 
							placeholder="sk_test_... or sk_live_..." 
							bind:value={secret} 
						/>
						<button type="submit" class="btn btn-primary" disabled={saving || !secret.trim()}>
							{saving ? 'Saving...' : 'Connect Stripe'}
						</button>
					</div>
				</form>
				
				{#if error}
					<div class="error">{error}</div>
				{/if}
				{#if success}
					<div class="success">{success}</div>
				{/if}
			</div>
		{:else}
			<!-- Account Overview -->
			<div class="overview-section">
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
				<div class="banner-actions">
					<button class="btn btn-secondary" on:click={fetchSummary}>
						🔄 Refresh Data
					</button>
					<!-- Add option to update key -->
					<button class="btn btn-outline" on:click={() => { summary = { enabled: false }; }}>
						🔑 Update Key
					</button>
				</div>
			</div>

			<!-- Products Section -->
			{#if summary.products && summary.products.length > 0}
				<section class="data-section">
					<h2>Products</h2>
					<div class="data-grid">
						{#each summary.products as product}
							<div class="data-card">
								<div class="card-header">
									<h4>{product.name}</h4>
									<span class="status-badge {product.active ? 'active' : 'inactive'}">
										{product.active ? 'Active' : 'Inactive'}
									</span>
								</div>
								{#if product.description}
									<p class="description">{product.description}</p>
								{/if}
								<div class="metadata">
									<small>ID: {product.id}</small>
									<small>Created: {formatDate(product.createdAt)}</small>
								</div>
							</div>
						{/each}
					</div>
				</section>
			{/if}

			<!-- Prices Section -->
			{#if summary.prices && summary.prices.length > 0}
				<section class="data-section">
					<h2>Pricing</h2>
					<div class="data-grid">
						{#each summary.prices as price}
							<div class="data-card">
								<div class="card-header">
									<h4>{price.nickname || 'Price'}</h4>
									<span class="status-badge {price.active ? 'active' : 'inactive'}">
										{price.active ? 'Active' : 'Inactive'}
									</span>
								</div>
								<div class="price-info">
									<div class="amount">{formatCurrency(price.unitAmount, price.currency)}</div>
									{#if price.recurring}
										<div class="interval">
											per {price.recurring.interval}
											{#if price.recurring.intervalCount > 1}
												({price.recurring.intervalCount})
											{/if}
										</div>
									{/if}
								</div>
								<div class="metadata">
									<small>ID: {price.id}</small>
									<small>Product: {price.productID}</small>
									<small>Created: {formatDate(price.createdAt)}</small>
								</div>
							</div>
						{/each}
					</div>
				</section>
			{/if}

			<!-- Customers Section -->
			{#if summary.customers && summary.customers.length > 0}
				<section class="data-section">
					<h2>Recent Customers</h2>
					<div class="data-grid">
						{#each summary.customers as customer}
							<div class="data-card">
								<div class="card-header">
									<h4>{customer.name || 'No Name'}</h4>
									<span class="email">{customer.email}</span>
								</div>
								<div class="metadata">
									<small>ID: {customer.id}</small>
									<small>Created: {formatDate(customer.createdAt)}</small>
								</div>
							</div>
						{/each}
					</div>
				</section>
			{/if}

			<!-- Subscriptions Section -->
			{#if summary.subscriptions && summary.subscriptions.length > 0}
				<section class="data-section">
					<h2>Recent Subscriptions</h2>
					<div class="data-grid">
						{#each summary.subscriptions as subscription}
							<div class="data-card">
								<div class="card-header">
									<h4>Subscription {subscription.id.slice(-8)}</h4>
									<span class="status-badge" style="color: {getStatusColor(subscription.status)}">
										{subscription.status}
									</span>
								</div>
								<div class="subscription-info">
									<div class="period">
										Next billing: {formatDate(subscription.currentPeriodEnd)}
									</div>
									{#if subscription.cancelAtPeriodEnd}
										<div class="cancellation">Cancels at period end</div>
									{/if}
								</div>
								<div class="metadata">
									<small>Created: {formatDate(subscription.createdAt)}</small>
								</div>
							</div>
						{/each}
					</div>
				</section>
			{/if}

			<!-- Payment Intents Section -->
			{#if summary.payment_intents && summary.payment_intents.length > 0}
				<section class="data-section">
					<h2>Recent Payments</h2>
					<div class="data-grid">
						{#each summary.payment_intents as payment}
							<div class="data-card">
								<div class="card-header">
									<h4>Payment {payment.id.slice(-8)}</h4>
									<span class="status-badge" style="color: {getStatusColor(payment.status)}">
										{payment.status}
									</span>
								</div>
								<div class="payment-info">
									<div class="amount">{formatCurrency(payment.amount, payment.currency)}</div>
								</div>
								<div class="metadata">
									<small>Created: {formatDate(payment.createdAt)}</small>
								</div>
							</div>
						{/each}
					</div>
				</section>
			{/if}

			<!-- Invoices Section -->
			{#if summary.invoices && summary.invoices.length > 0}
				<section class="data-section">
					<h2>Recent Invoices</h2>
					<div class="data-grid">
						{#each summary.invoices as invoice}
							<div class="data-card">
								<div class="card-header">
									<h4>Invoice {invoice.id.slice(-8)}</h4>
									<span class="status-badge" style="color: {getStatusColor(invoice.status)}">
										{invoice.status}
									</span>
								</div>
								<div class="invoice-info">
									<div class="amount">{formatCurrency(invoice.amount, invoice.currency)}</div>
								</div>
								<div class="metadata">
									<small>Created: {formatDate(invoice.createdAt)}</small>
								</div>
							</div>
						{/each}
					</div>
				</section>
			{/if}

			<!-- Capabilities Section -->
			<section class="data-section">
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
		{/if}
	</div>
{/if}

<style>
	.stripe-dashboard {
		max-width: 1200px;
		margin: 0 auto;
		padding: var(--space-lg);
	}

	.dashboard-header {
		text-align: center;
		margin-bottom: var(--space-xl);
	}

	.dashboard-header h1 {
		font-size: 2.5rem;
		margin-bottom: var(--space-sm);
		color: var(--primary);
	}

	.dashboard-header p {
		font-size: 1.1rem;
		color: var(--text-muted);
	}

	.setup-section {
		background: var(--surface);
		padding: var(--space-xl);
		border-radius: var(--radius-lg);
		text-align: center;
		border: 2px dashed var(--border);
	}

	.setup-section h2 {
		margin-bottom: var(--space-md);
		color: var(--text);
	}

	.setup-section p {
		margin-bottom: var(--space-lg);
		color: var(--text-muted);
	}

	.input-group {
		display: flex;
		gap: var(--space-md);
		max-width: 500px;
		margin: 0 auto;
	}

	.input {
		flex: 1;
		padding: var(--space-md);
		border: 1px solid var(--border);
		border-radius: var(--radius-md);
		font-size: 1rem;
		background: var(--surface);
		color: var(--text);
	}

	.btn {
		padding: var(--space-md) var(--space-lg);
		border: none;
		border-radius: var(--radius-md);
		font-size: 1rem;
		cursor: pointer;
		transition: all 0.2s ease;
		white-space: nowrap;
	}

	.btn-primary {
		background: var(--primary);
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: var(--primary-dark);
		transform: translateY(-1px);
	}

	.btn-primary:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-secondary {
		background: var(--surface);
		color: var(--text);
		border: 1px solid var(--border);
	}

	.btn-secondary:hover {
		background: var(--surface-hover);
	}

	.btn-outline {
		background: transparent;
		color: var(--primary);
		border: 1px solid var(--primary);
	}

	.btn-outline:hover {
		background: var(--primary);
		color: white;
	}

	.overview-section {
		margin-bottom: var(--space-xl);
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

	.banner-actions {
		display: flex;
		gap: var(--space-md);
	}

	.data-section {
		margin-bottom: var(--space-xl);
	}

	.data-section h2 {
		margin-bottom: var(--space-lg);
		color: var(--text);
		font-size: 1.5rem;
		border-bottom: 2px solid var(--border);
		padding-bottom: var(--space-sm);
	}

	.data-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: var(--space-md);
	}

	.data-card {
		background: var(--surface);
		padding: var(--space-lg);
		border-radius: var(--radius-lg);
		border: 1px solid var(--border);
		transition: all 0.2s ease;
	}

	.data-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
	}

	.card-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-md);
	}

	.card-header h4 {
		margin: 0;
		color: var(--text);
		font-size: 1.1rem;
	}

	.status-badge {
		padding: var(--space-xs) var(--space-md);
		border-radius: var(--radius-md);
		font-size: 0.8rem;
		font-weight: bold;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.status-badge.active {
		background: var(--success);
		color: white;
	}

	.status-badge.inactive {
		background: var(--text-muted);
		color: white;
	}

	.description {
		color: var(--text-muted);
		margin-bottom: var(--space-md);
		line-height: 1.5;
	}

	.price-info {
		margin-bottom: var(--space-md);
	}

	.amount {
		font-size: 1.5rem;
		font-weight: bold;
		color: var(--primary);
		margin-bottom: var(--space-xs);
	}

	.interval {
		color: var(--text-muted);
		font-size: 0.9rem;
	}

	.subscription-info {
		margin-bottom: var(--space-md);
	}

	.period {
		color: var(--text);
		margin-bottom: var(--space-xs);
	}

	.cancellation {
		color: var(--warning);
		font-size: 0.9rem;
		font-weight: bold;
	}

	.payment-info {
		margin-bottom: var(--space-md);
	}

	.invoice-info {
		margin-bottom: var(--space-md);
	}

	.email {
		color: var(--text-muted);
		font-size: 0.9rem;
	}

	.metadata {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.metadata small {
		color: var(--text-muted);
		font-family: monospace;
		font-size: 0.8rem;
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

	.error {
		color: var(--error);
		margin: var(--space-sm) 0 0 0;
		padding: var(--space-sm);
		background: var(--error-light);
		border-radius: var(--radius-md);
		border: 1px solid var(--error);
	}

	.success {
		color: var(--success);
		margin: var(--space-sm) 0 0 0;
		padding: var(--space-sm);
		background: var(--success-light);
		border-radius: var(--radius-md);
		border: 1px solid var(--success);
	}

	@media (max-width: 768px) {
		.overview-grid {
			grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		}

		.data-grid {
			grid-template-columns: 1fr;
		}

		.input-group {
			flex-direction: column;
		}

		.environment-banner {
			flex-direction: column;
			gap: var(--space-md);
		}

		.banner-actions {
			width: 100%;
			justify-content: center;
		}
	}
</style> 