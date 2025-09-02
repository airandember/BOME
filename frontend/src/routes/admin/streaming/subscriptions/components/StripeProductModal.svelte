<script lang="ts">
	import { onMount } from 'svelte';
	import { StreamingSubscriptionService, type StripeProduct } from '$lib/services/streaming-subscriptions';
	import { showToast } from '$lib/toast';

	// Props
	export let show = false;
	export let onSelect: (product: StripeProduct | null) => void = () => {};
	export let onClose: () => void = () => {};
	export let selectedProductId: string | null = null;

	// State
	let products: StripeProduct[] = [];
	let loading = true;
	let searchTerm = '';
	let selectedProduct: StripeProduct | null = null;

	// Filtered products based on search
	$: filteredProducts = products.filter(product => 
		product.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
		product.description?.toLowerCase().includes(searchTerm.toLowerCase()) ||
		product.stripe_id.toLowerCase().includes(searchTerm.toLowerCase())
	);

	// Load products when modal opens
	$: if (show) {
		loadProducts();
	}

	// Set initial selection
	$: if (selectedProductId && products.length > 0) {
		selectedProduct = products.find(p => p.stripe_id === selectedProductId) || null;
	}

	async function loadProducts() {
		try {
			loading = true;
			products = await StreamingSubscriptionService.getAvailableStripeProducts();
		} catch (error) {
			console.error('Failed to load Stripe products:', error);
			showToast('Failed to load Stripe products', 'error');
		} finally {
			loading = false;
		}
	}

	function handleProductSelect(product: StripeProduct) {
		selectedProduct = product;
	}

	function handleConfirm() {
		onSelect(selectedProduct);
		handleClose();
	}

	function handleClose() {
		show = false;
		selectedProduct = null;
		searchTerm = '';
		onClose();
	}

	async function toggleProductAvailability(product: StripeProduct) {
		try {
			const newAvailability = !product.available;
			await StreamingSubscriptionService.updateStripeProductAvailability(product.stripe_id, newAvailability);
			
			// Update local state
			product.available = newAvailability;
			products = [...products]; // Trigger reactivity
			
			showToast(`Product ${newAvailability ? 'enabled' : 'disabled'} successfully`, 'success');
		} catch (error) {
			console.error('Failed to update product availability:', error);
			showToast('Failed to update product availability', 'error');
		}
	}
</script>

{#if show}
	<div class="modal-overlay" on:click={handleClose}>
		<div class="modal-content" on:click|stopPropagation>
			<div class="modal-header">
				<h2>Select Stripe Product</h2>
				<button class="close-button" on:click={handleClose}>×</button>
			</div>

			<div class="modal-body">
				<!-- Search -->
				<div class="search-section">
					<input
						type="text"
						placeholder="Search products by name, description, or ID..."
						bind:value={searchTerm}
						class="search-input"
					/>
				</div>

				{#if loading}
					<div class="loading-state">
						<div class="spinner"></div>
						<p>Loading Stripe products...</p>
					</div>
				{:else if filteredProducts.length === 0}
					<div class="empty-state">
						{#if searchTerm}
							<p>No products found matching "{searchTerm}"</p>
						{:else}
							<p>No available Stripe products found</p>
						{/if}
					</div>
				{:else}
					<div class="products-list">
						{#each filteredProducts as product (product.id)}
							<div 
								class="product-item"
								class:selected={selectedProduct?.id === product.id}
								class:inactive={!product.active}
								on:click={() => handleProductSelect(product)}
							>
								<div class="product-main">
									<!-- Product Image -->
									{#if product.images && product.images.length > 0}
										<div class="product-image">
											<img src={product.images[0]} alt={product.name} />
										</div>
									{:else}
										<div class="product-image-placeholder">
											<svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"></path>
											</svg>
										</div>
									{/if}

									<div class="product-info">
										<div class="product-header">
											<h3 class="product-name">{product.name}</h3>
											{#if product.price}
												<div class="product-price">
													${(product.price / 100).toFixed(2)}
													{#if product.recurring_interval}
														<span class="price-interval">/{product.recurring_interval}</span>
													{/if}
													{#if product.currency && product.currency.toUpperCase() !== 'USD'}
														<span class="price-currency">{product.currency.toUpperCase()}</span>
													{/if}
												</div>
											{/if}
										</div>
										
										<p class="product-id">ID: {product.stripe_id}</p>
										
										{#if product.description}
											<p class="product-description">{product.description}</p>
										{/if}

										<!-- Enhanced Product Details -->
										<div class="product-details">
											{#if product.recurring_interval}
												<span class="detail-badge billing-frequency">
													<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
													</svg>
													Billed {product.recurring_interval}ly
												</span>
											{/if}

											{#if product.unit_label}
												<span class="detail-badge">
													<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"></path>
													</svg>
													{product.unit_label}
												</span>
											{/if}
											
											{#if product.shippable}
												<span class="detail-badge shippable">
													<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"></path>
													</svg>
													Shippable
												</span>
											{/if}

											{#if product.tax_code}
												<span class="detail-badge tax">
													<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z"></path>
													</svg>
													Tax: {product.tax_code}
												</span>
											{/if}

											{#if product.livemode === false}
												<span class="detail-badge test-mode">
													<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
													</svg>
													Test Mode
												</span>
											{/if}
										</div>

										<!-- Metadata Display -->
										{#if product.metadata && Object.keys(product.metadata).length > 0}
											<div class="product-metadata">
												<details class="metadata-details">
													<summary>Metadata ({Object.keys(product.metadata).length})</summary>
													<div class="metadata-content">
														{#each Object.entries(product.metadata) as [key, value]}
															<div class="metadata-item">
																<span class="metadata-key">{key}:</span>
																<span class="metadata-value">{value}</span>
															</div>
														{/each}
													</div>
												</details>
											</div>
										{/if}
									</div>
									
									<div class="product-status">
										<span class="status-badge" class:active={product.active}>
											{product.active ? 'Active' : 'Inactive'}
										</span>
										<button
											class="availability-toggle"
											class:available={product.available}
											on:click|stopPropagation={() => toggleProductAvailability(product)}
											title={product.available ? 'Disable for new plans' : 'Enable for new plans'}
										>
											{product.available ? 'Available' : 'Disabled'}
										</button>
									</div>
								</div>

								{#if selectedProduct?.id === product.id}
									<div class="selection-indicator">✓</div>
								{/if}
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<div class="modal-footer">
				<button class="btn btn-secondary" on:click={handleClose}>
					Cancel
				</button>
				<button 
					class="btn btn-primary" 
					on:click={handleConfirm}
					disabled={!selectedProduct}
				>
					Select Product
				</button>
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
		width: 90%;
		max-width: 800px;
		max-height: 80vh;
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

	.search-section {
		padding: 1rem 1.5rem;
		border-bottom: 1px solid #e5e7eb;
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

	.products-list {
		flex: 1;
		overflow-y: auto;
		padding: 1rem;
	}

	.product-item {
		border: 2px solid #e5e7eb;
		border-radius: 8px;
		padding: 1rem;
		margin-bottom: 0.75rem;
		cursor: pointer;
		transition: all 0.2s ease;
		position: relative;
	}

	.product-item:hover {
		border-color: #d1d5db;
		background: #f9fafb;
	}

	.product-item.selected {
		border-color: #3b82f6;
		background: #eff6ff;
	}

	.product-item.inactive {
		opacity: 0.6;
	}

	.product-main {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
	}

	.product-info {
		flex: 1;
	}

	.product-name {
		margin: 0 0 0.25rem 0;
		font-size: 1.125rem;
		font-weight: 600;
		color: #111827;
	}

	.product-id {
		margin: 0 0 0.5rem 0;
		font-size: 0.875rem;
		color: #6b7280;
		font-family: monospace;
	}

	.product-description {
		margin: 0;
		font-size: 0.875rem;
		color: #4b5563;
		line-height: 1.4;
	}

	.product-status {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 0.5rem;
	}

	.status-badge {
		padding: 0.25rem 0.75rem;
		border-radius: 4px;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		background: #fee2e2;
		color: #dc2626;
	}

	.status-badge.active {
		background: #d1fae5;
		color: #059669;
	}

	.availability-toggle {
		padding: 0.25rem 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		font-size: 0.75rem;
		font-weight: 500;
		cursor: pointer;
		background: #f3f4f6;
		color: #6b7280;
		transition: all 0.2s ease;
	}

	.availability-toggle.available {
		background: #d1fae5;
		color: #059669;
		border-color: #a7f3d0;
	}

	.availability-toggle:hover {
		background: #e5e7eb;
	}

	.availability-toggle.available:hover {
		background: #a7f3d0;
	}

	.selection-indicator {
		position: absolute;
		top: 0.5rem;
		right: 0.5rem;
		width: 24px;
		height: 24px;
		background: #3b82f6;
		color: white;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: bold;
		font-size: 0.875rem;
	}

	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
		padding: 1.5rem;
		border-top: 1px solid #e5e7eb;
	}

	.btn {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 6px;
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
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

	/* Enhanced Product Display Styles */
	.product-image {
		width: 60px;
		height: 60px;
		border-radius: 8px;
		overflow: hidden;
		flex-shrink: 0;
		background: #f9fafb;
		border: 1px solid #e5e7eb;
	}

	.product-image img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.product-image-placeholder {
		width: 60px;
		height: 60px;
		border-radius: 8px;
		background: #f9fafb;
		border: 1px solid #e5e7eb;
		display: flex;
		align-items: center;
		justify-content: center;
		color: #9ca3af;
		flex-shrink: 0;
	}

	.product-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 0.5rem;
	}

	.product-price {
		background: linear-gradient(135deg, #10b981 0%, #059669 100%);
		color: white;
		padding: 0.25rem 0.75rem;
		border-radius: 6px;
		font-size: 0.875rem;
		font-weight: 600;
		white-space: nowrap;
		box-shadow: 0 2px 4px rgba(16, 185, 129, 0.2);
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}

	.price-interval {
		opacity: 0.9;
		font-weight: 500;
		font-size: 0.75rem;
	}

	.price-currency {
		opacity: 0.8;
		font-weight: 500;
		font-size: 0.75rem;
		background: rgba(255, 255, 255, 0.2);
		padding: 0.125rem 0.375rem;
		border-radius: 3px;
	}

	.product-details {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		margin: 0.75rem 0;
	}

	.detail-badge {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		padding: 0.25rem 0.5rem;
		background: #f3f4f6;
		color: #6b7280;
		border-radius: 4px;
		font-size: 0.75rem;
		font-weight: 500;
		border: 1px solid #e5e7eb;
	}

	.detail-badge.billing-frequency {
		background: #e0e7ff;
		color: #3730a3;
		border-color: #c7d2fe;
		font-weight: 600;
	}

	.detail-badge.shippable {
		background: #dbeafe;
		color: #1d4ed8;
		border-color: #bfdbfe;
	}

	.detail-badge.tax {
		background: #fef3c7;
		color: #d97706;
		border-color: #fed7aa;
	}

	.detail-badge.test-mode {
		background: #fee2e2;
		color: #dc2626;
		border-color: #fecaca;
	}

	.product-metadata {
		margin-top: 0.75rem;
	}

	.metadata-details {
		border: 1px solid #e5e7eb;
		border-radius: 4px;
		background: #f9fafb;
	}

	.metadata-details summary {
		padding: 0.5rem 0.75rem;
		cursor: pointer;
		font-size: 0.75rem;
		font-weight: 600;
		color: #6b7280;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		user-select: none;
	}

	.metadata-details summary:hover {
		background: #f3f4f6;
	}

	.metadata-content {
		padding: 0.75rem;
		border-top: 1px solid #e5e7eb;
		background: white;
	}

	.metadata-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.25rem 0;
		font-size: 0.75rem;
	}

	.metadata-item:not(:last-child) {
		border-bottom: 1px solid #f3f4f6;
	}

	.metadata-key {
		font-weight: 600;
		color: #374151;
		margin-right: 0.5rem;
	}

	.metadata-value {
		color: #6b7280;
		font-family: monospace;
		background: #f9fafb;
		padding: 0.125rem 0.375rem;
		border-radius: 3px;
		border: 1px solid #e5e7eb;
	}

	/* Responsive adjustments */
	@media (max-width: 640px) {
		.product-main {
			flex-direction: column;
			gap: 1rem;
		}

		.product-header {
			flex-direction: column;
			gap: 0.5rem;
			align-items: flex-start;
		}

		.product-image,
		.product-image-placeholder {
			width: 80px;
			height: 80px;
		}

		.product-details {
			justify-content: flex-start;
		}
	}
</style>
