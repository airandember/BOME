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
				error = 'Failed to load products';
			}
		} catch (err) {
			error = 'Failed to load products';
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

	// Currency formatting utility
	function formatCurrency(amount: number, currency: string = 'USD'): string {
		return new Intl.NumberFormat('en-US', { 
			style: 'currency', 
			currency: currency.toUpperCase() 
		}).format(amount);
	}
</script>

{#if loading}
	<div class="loading">
		<div class="spinner"></div>
		<p>Loading products...</p>
	</div>
{:else if error}
	<div class="error-state">
		<h3>Error Loading Products</h3>
		<p>{error}</p>
		<button class="btn btn-primary" on:click={fetchSummary}>Retry</button>
	</div>
{:else if summary}
	<div class="products-container">
		<div class="products-header">
			<h1>Products</h1>
			<p>Manage your Stripe products and offerings</p>
			<div class="stats-summary">
				<div class="stat">
					<span class="stat-value">{summary.products_count || 0}</span>
					<span class="stat-label">Total Products</span>
				</div>
				<div class="stat">
					<span class="stat-value">{summary.products?.filter((p: any) => p.Active).length || 0}</span>
					<span class="stat-label">Active</span>
				</div>
				<div class="stat">
					<span class="stat-value">{summary.products?.filter((p: any) => !p.Active).length || 0}</span>
					<span class="stat-label">Inactive</span>
				</div>
			</div>
		</div>

		{#if summary.products && summary.products.length > 0}
			<div class="products-grid">
				{#each summary.products as product}
					<div class="product-card">
						<div class="product-header">
							<div class="product-info">
								<h3 class="product-name">{product.Name}</h3>
								<span class="product-id">ID: {product.ID}</span>
							</div>
							<span class="status-badge {product.Active ? 'active' : 'inactive'}">
								{product.Active ? 'Active' : 'Inactive'}
							</span>
						</div>

						{#if product.Description}
							<div class="product-description">
								<p>{product.Description}</p>
							</div>
						{/if}

						<div class="product-metadata">
							<div class="metadata-item">
								<span class="metadata-label">Created:</span>
								<span class="metadata-value">{formatDate(product.CreatedAt)}</span>
							</div>
							
							{#if product.Metadata && Object.keys(product.Metadata).length > 0}
								<div class="metadata-item">
									<span class="metadata-label">Custom Data:</span>
									<div class="metadata-tags">
										{#each Object.entries(product.Metadata) as [key, value]}
											<span class="metadata-tag">{key}: {value}</span>
										{/each}
									</div>
								</div>
							{/if}
						</div>

						<!-- Related Prices -->
						{#if summary.prices}
							{@const relatedPrices = summary.prices.filter((p: any) => p.ProductID === product.ID)}
							{#if relatedPrices.length > 0}
								<div class="related-prices">
									<h4>Pricing Options ({relatedPrices.length})</h4>
									<div class="prices-list">
										{#each relatedPrices as price}
											<div class="price-item">
												<div class="price-amount">
													{formatCurrency(price.UnitAmount / 100, price.Currency)}
												</div>
												{#if price.Recurring}
													<div class="price-interval">
														per {price.Recurring.Interval}
														{#if price.Recurring.IntervalCount > 1}
															({price.Recurring.IntervalCount})
														{/if}
													</div>
												{:else}
													<div class="price-interval">One-time</div>
												{/if}
												<span class="price-status {price.Active ? 'active' : 'inactive'}">
													{price.Active ? 'Active' : 'Inactive'}
												</span>
											</div>
										{/each}
									</div>
								</div>
							{/if}
						{/if}

						<div class="product-actions">
							<button class="btn btn-outline btn-sm">Edit Product</button>
							<button class="btn btn-outline btn-sm">View Prices</button>
							{#if product.Active}
								<button class="btn btn-outline btn-sm btn-warning">Deactivate</button>
							{:else}
								<button class="btn btn-outline btn-sm btn-success">Activate</button>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		{:else}
			<div class="empty-state">
				<div class="empty-icon">📦</div>
				<h3>No Products Found</h3>
				<p>You haven't created any products yet. Products represent your goods or services.</p>
				<button class="btn btn-primary">Create First Product</button>
			</div>
		{/if}
	</div>
{/if}

<style>
	.products-container {
		padding: var(--space-lg);
	}

	.products-header {
		margin-bottom: var(--space-xl);
	}

	.products-header h1 {
		margin: 0 0 var(--space-sm) 0;
		color: var(--text);
		font-size: 2rem;
	}

	.products-header p {
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

	.products-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
		gap: var(--space-lg);
	}

	.product-card {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
		transition: all 0.2s ease;
	}

	.product-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
	}

	.product-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: var(--space-md);
	}

	.product-info {
		flex: 1;
	}

	.product-name {
		margin: 0 0 var(--space-xs) 0;
		color: var(--text);
		font-size: 1.3rem;
		font-weight: 600;
	}

	.product-id {
		font-family: monospace;
		font-size: 0.9rem;
		color: var(--text-muted);
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

	.product-description {
		margin-bottom: var(--space-md);
		padding: var(--space-md);
		background: var(--bg-secondary);
		border-radius: var(--radius-md);
		border-left: 4px solid var(--primary);
	}

	.product-description p {
		margin: 0;
		color: var(--text);
		line-height: 1.5;
	}

	.product-metadata {
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

	.related-prices {
		margin-bottom: var(--space-md);
		padding: var(--space-md);
		background: var(--bg-secondary);
		border-radius: var(--radius-md);
	}

	.related-prices h4 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
		font-size: 1rem;
	}

	.prices-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}

	.price-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-sm);
		background: var(--surface);
		border-radius: var(--radius-md);
		border: 1px solid var(--border);
	}

	.price-amount {
		font-weight: 600;
		color: var(--primary);
		font-size: 1.1rem;
	}

	.price-interval {
		color: var(--text-muted);
		font-size: 0.9rem;
	}

	.price-status {
		padding: 2px 8px;
		border-radius: var(--radius-sm);
		font-size: 0.7rem;
		font-weight: bold;
		text-transform: uppercase;
	}

	.price-status.active {
		background: var(--success);
		color: white;
	}

	.price-status.inactive {
		background: var(--text-muted);
		color: white;
	}

	.product-actions {
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

	.btn-warning {
		color: var(--warning);
		border-color: var(--warning);
	}

	.btn-warning:hover {
		background: var(--warning);
		color: white;
	}

	.btn-success {
		color: var(--success);
		border-color: var(--success);
	}

	.btn-success:hover {
		background: var(--success);
		color: white;
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
		.products-grid {
			grid-template-columns: 1fr;
		}

		.stats-summary {
			justify-content: center;
		}

		.product-actions {
			justify-content: center;
		}
	}
</style> 
