<script lang="ts">
	import { onMount } from 'svelte';
	import { StreamingSubscriptionService } from '$lib/services/streaming-subscriptions';
	import { showToast } from '$lib/toast';
	import { apiRequest } from '$lib/auth';

	// Props
	export let show = false;
	export let onClose: () => void = () => {};
	export let onImportComplete: () => void = () => {};

	// State
	let products: any[] = [];
	let loading = true;
	let saving = false;
	let searchTerm = '';
	let selectedProducts = new Set<string>(); // Set of stripe_ids

	// Filtered products based on search
	$: filteredProducts = products.filter(product => 
		product.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
		product.stripe_id.toLowerCase().includes(searchTerm.toLowerCase()) ||
		(product.description && product.description.toLowerCase().includes(searchTerm.toLowerCase()))
	);

	// Load products when modal opens
	$: if (show) {
		loadAllStripeProducts();
	}

	async function loadAllStripeProducts() {
		try {
			loading = true;
			// Use apiRequest to get ALL products (not just available ones)
			const response = await apiRequest('/admin/streaming/stripe/products/all');
			
			if (!response.ok) {
				throw new Error(`Failed to fetch products: ${response.status}`);
			}

			const data = await response.json();
			products = data.products || [];
			
			// Initialize selectedProducts with products that already have plans in subscription_plans
			selectedProducts = new Set(
				products
					.filter(p => p.has_plan) // ✅ CHANGED: Use has_plan flag (products already in subscription_plans)
					.map(p => p.stripe_id)
			);
		} catch (error) {
			console.error('Failed to load Stripe products:', error);
			showToast('Failed to load Stripe products', 'error');
		} finally {
			loading = false;
		}
	}

	function toggleProductSelection(stripeId: string) {
		if (selectedProducts.has(stripeId)) {
			selectedProducts.delete(stripeId);
		} else {
			selectedProducts.add(stripeId);
		}
		selectedProducts = new Set(selectedProducts); // Trigger reactivity
	}

	function selectAll() {
		selectedProducts = new Set(filteredProducts.map(p => p.stripe_id));
	}

	function deselectAll() {
		selectedProducts = new Set();
	}

	async function handleSave() {
		saving = true;
		try {
			// Get products that are newly selected (checked but don't already have a plan)
			const newProductsToImport = products
				.filter(p => selectedProducts.has(p.stripe_id) && !p.has_plan)
				.map(p => p.stripe_id);

			if (newProductsToImport.length === 0) {
				showToast('No new products to import - all selected products already have plans', 'info');
				handleClose();
				return;
			}

			// Import only the newly selected products as subscription plans
			const importResult = await StreamingSubscriptionService.importStripeProductsAsPlans(newProductsToImport);
			
			showToast(
				`Successfully imported ${importResult.imported_count} products as subscription plans${importResult.skipped_count > 0 ? ` (${importResult.skipped_count} already existed)` : ''}`, 
				'success'
			);
			
			onImportComplete();
			handleClose();
		} catch (error) {
			console.error('Failed to import products as plans:', error);
			showToast('Failed to import products as subscription plans', 'error');
		} finally {
			saving = false;
		}
	}

	function handleClose() {
		show = false;
		selectedProducts = new Set();
		searchTerm = '';
		onClose();
	}

	// Format price for display
	function formatPrice(price: number | null): string {
		if (price === null || price === undefined) return 'N/A';
		return `$${(price / 100).toFixed(2)}`;
	}
</script>

{#if show}
	<div class="modal-overlay" on:click={handleClose}>
		<div class="modal-content" on:click|stopPropagation>
			<div class="modal-header">
				<h2>Select Stripe Products to Import</h2>
				<button class="close-button" on:click={handleClose}>×</button>
			</div>

			<div class="modal-body">
				{#if loading}
					<div class="loading-state">
						<div class="spinner"></div>
						<p>Loading Stripe products...</p>
					</div>
				{:else}
					<!-- Search and Controls -->
					<div class="controls-section">
						<div class="search-section">
							<input
								type="text"
								placeholder="Search products by name, ID, or description..."
								bind:value={searchTerm}
								class="search-input"
							/>
						</div>
						
						<div class="bulk-actions">
							<button class="btn btn-outline btn-sm" on:click={selectAll}>
								Select All ({filteredProducts.length})
							</button>
							<button class="btn btn-outline btn-sm" on:click={deselectAll}>
								Deselect All
							</button>
							<span class="selection-count">
								{selectedProducts.size} selected
							</span>
						</div>
					</div>

					{#if filteredProducts.length === 0}
						<div class="empty-state">
							{#if searchTerm}
								<p>No products found matching "{searchTerm}"</p>
							{:else}
								<p>No Stripe products found</p>
							{/if}
						</div>
					{:else}
						<!-- Products Table -->
						<div class="table-container">
							<table class="products-table">
								<thead>
									<tr>
										<th class="checkbox-col">Add to Plans</th>
										<th>Title</th>
										<th>Stripe ID</th>
										<th>Price</th>
										<th>Status</th>
									</tr>
								</thead>
								<tbody>
									{#each filteredProducts as product (product.stripe_id)}
										<tr class="product-row" class:selected={selectedProducts.has(product.stripe_id)}>
											<td class="checkbox-col">
												<input
													type="checkbox"
													checked={selectedProducts.has(product.stripe_id)}
													on:change={() => toggleProductSelection(product.stripe_id)}
													class="product-checkbox"
												/>
											</td>
											<td class="product-title">
												<div class="title-info">
													<span class="product-name">{product.name}</span>
													{#if product.description}
														<span class="product-description">{product.description}</span>
													{/if}
												</div>
											</td>
											<td class="stripe-id">
												<code>{product.stripe_id}</code>
											</td>
											<td class="price">
												{formatPrice(product.price)}
											</td>
											<td class="status">
												<div class="status-badges">
													<span class="status-badge" class:active={product.active}>
														{product.active ? 'Active in Stripe' : 'Inactive'}
													</span>
													<span class="availability-badge" class:available={product.has_plan}>
														{product.has_plan ? '✅ Has Plan' : '⏳ No Plan Yet'}
													</span>
												</div>
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					{/if}
				{/if}
			</div>

			<div class="modal-footer">
				<div class="footer-info">
					{#if !loading}
						<span class="product-count">
							{filteredProducts.length} products • {selectedProducts.size} selected
						</span>
					{/if}
				</div>
				<div class="footer-actions">
					<button class="btn btn-secondary" on:click={handleClose} disabled={saving}>
						Cancel
					</button>
					<button 
						class="btn btn-primary" 
						on:click={handleSave}
						disabled={saving || loading}
						title="Add selected products as manageable subscription plans with full Stripe integration"
					>
						{#if saving}
							<div class="btn-spinner"></div>
							Processing...
						{:else}
							✅ Add to Subscription Plans
						{/if}
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}

<style>
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}

	.modal-content {
		background: white;
		border-radius: 8px;
		width: 95%;
		max-width: 1200px;
		max-height: 90vh;
		display: flex;
		flex-direction: column;
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1.5rem;
		border-bottom: 1px solid #e5e7eb;
	}

	.modal-header h2 {
		margin: 0;
		color: #111827;
		font-size: 1.5rem;
		font-weight: 600;
	}

	.close-button {
		background: none;
		border: none;
		font-size: 1.5rem;
		color: #6b7280;
		cursor: pointer;
		padding: 0.25rem;
		line-height: 1;
	}

	.close-button:hover {
		color: #374151;
	}

	.modal-body {
		flex: 1;
		overflow: hidden;
		display: flex;
		flex-direction: column;
	}

	.loading-state, .empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 3rem;
		color: #6b7280;
	}

	.spinner {
		width: 32px;
		height: 32px;
		border: 3px solid #e5e7eb;
		border-top: 3px solid #3b82f6;
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-bottom: 1rem;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.controls-section {
		padding: 1rem 1.5rem;
		border-bottom: 1px solid #e5e7eb;
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 1rem;
		flex-wrap: wrap;
	}

	.search-section {
		flex: 1;
		min-width: 300px;
	}

	.search-input {
		width: 100%;
		padding: 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 6px;
		font-size: 0.875rem;
	}

	.search-input:focus {
		outline: none;
		border-color: #3b82f6;
		box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.25);
	}

	.bulk-actions {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		flex-wrap: wrap;
	}

	.selection-count {
		font-size: 0.875rem;
		color: #6b7280;
		font-weight: 500;
	}

	.table-container {
		flex: 1;
		overflow: auto;
		padding: 0 1.5rem;
	}

	.products-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
	}

	.products-table th {
		background: #f9fafb;
		padding: 0.75rem;
		text-align: left;
		font-weight: 600;
		color: #374151;
		border-bottom: 1px solid #e5e7eb;
		position: sticky;
		top: 0;
		z-index: 10;
	}

	.products-table td {
		padding: 0.75rem;
		border-bottom: 1px solid #f3f4f6;
	}

	.checkbox-col {
		width: 120px;
		text-align: center;
	}

	.product-checkbox {
		width: 16px;
		height: 16px;
		cursor: pointer;
	}

	.product-row {
		transition: background-color 0.2s ease;
	}

	.product-row:hover {
		background: #f9fafb;
	}

	.product-row.selected {
		background: #eff6ff;
	}

	.title-info {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.product-name {
		font-weight: 600;
		color: #111827;
	}

	.product-description {
		font-size: 0.8125rem;
		color: #6b7280;
		line-height: 1.3;
	}

	.stripe-id code {
		background: #f3f4f6;
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-family: 'Monaco', 'Menlo', monospace;
		font-size: 0.8125rem;
		color: #374151;
	}

	.price {
		font-weight: 600;
		color: #059669;
	}

	.status-badges {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.status-badge, .availability-badge {
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		text-align: center;
	}

	.status-badge {
		background: #fee2e2;
		color: #dc2626;
	}

	.status-badge.active {
		background: #d1fae5;
		color: #059669;
	}

	.availability-badge {
		background: #f3f4f6;
		color: #6b7280;
	}

	.availability-badge.available {
		background: #dbeafe;
		color: #1d4ed8;
	}

	.modal-footer {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1.5rem;
		border-top: 1px solid #e5e7eb;
	}

	.footer-info {
		flex: 1;
	}

	.product-count {
		font-size: 0.875rem;
		color: #6b7280;
	}

	.footer-actions {
		display: flex;
		gap: 0.75rem;
	}

	.btn {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 6px;
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-sm {
		padding: 0.5rem 1rem;
		font-size: 0.8125rem;
	}

	.btn-outline {
		background: white;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-outline:hover:not(:disabled) {
		background: #f9fafb;
		border-color: #9ca3af;
	}

	.btn-secondary {
		background: #f3f4f6;
		color: #374151;
	}

	.btn-secondary:hover:not(:disabled) {
		background: #e5e7eb;
	}

	.btn-primary {
		background: #3b82f6;
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: #2563eb;
	}

	.btn-outline {
		background: white;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-outline:hover:not(:disabled) {
		background: #f9fafb;
		border-color: #9ca3af;
		transform: translateY(-1px);
	}

	.btn-spinner {
		width: 16px;
		height: 16px;
		border: 2px solid rgba(255, 255, 255, 0.3);
		border-top: 2px solid white;
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@media (max-width: 768px) {
		.modal-content {
			width: 100%;
			height: 100%;
			max-height: 100vh;
			border-radius: 0;
		}

		.controls-section {
			flex-direction: column;
			align-items: stretch;
		}

		.bulk-actions {
			justify-content: space-between;
		}

		.modal-footer {
			flex-direction: column;
			gap: 1rem;
			align-items: stretch;
		}

		.footer-actions {
			width: 100%;
		}

		.footer-actions .btn {
			flex: 1;
		}
	}
</style>
