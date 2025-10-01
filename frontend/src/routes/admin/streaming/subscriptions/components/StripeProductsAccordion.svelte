<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import { StripeProductsService, type StripeProduct, type StripeProductsResponse } from '$lib/services/stripe-products';
	import StripeProductCard from './StripeProductCard.svelte';
	import SubscriptionAccordion from './SubscriptionAccordion.svelte';

	// Props
	export let activeAccordion: string | null = 'active-plans';

	// State
	let isLoading = true;
	let error = '';
	let productsData: StripeProductsResponse | null = null;
	let isUpdating = false;

	// Reactive statements
	// Group products by active status and legacy status
	$: activePlansProducts = [
		...(productsData?.products.video_approved || []).filter(p => !p.legacy),
		...(productsData?.products.active || []).filter(p => !p.legacy)
	];
	$: legacyPlansProducts = [
		...(productsData?.products.video_approved || []).filter(p => p.legacy),
		...(productsData?.products.active || []).filter(p => p.legacy)
	];
	$: inactiveProducts = productsData?.products.inactive || [];
	

	onMount(async () => {
		await loadProducts();
	});

	async function loadProducts() {
		try {
			isLoading = true;
			error = '';
			productsData = await StripeProductsService.getProductsForAccordion();
		} catch (err: any) {
			error = err.message || 'Failed to load Stripe products';
			console.error('Error loading Stripe products:', err);
		} finally {
			isLoading = false;
		}
	}

	async function handleVideoApprovalToggle(product: StripeProduct) {
		if (isUpdating) return;

		try {
			isUpdating = true;
			const newStatus = !product.video_approved;
			
			await StripeProductsService.updateVideoApproval(product.id, newStatus);
			
			// Update local state
			if (productsData) {
				// Remove from current group
				if (product.video_approved) {
					productsData.products.video_approved = productsData.products.video_approved.filter(p => p.id !== product.id);
				} else if (product.active) {
					productsData.products.active = productsData.products.active.filter(p => p.id !== product.id);
				} else {
					productsData.products.inactive = productsData.products.inactive.filter(p => p.id !== product.id);
				}

				// Update product and add to new group
				const updatedProduct = { ...product, video_approved: newStatus };
				if (newStatus) {
					productsData.products.video_approved = [...productsData.products.video_approved, updatedProduct];
				} else if (product.active) {
					productsData.products.active = [...productsData.products.active, updatedProduct];
				} else {
					productsData.products.inactive = [...productsData.products.inactive, updatedProduct];
				}

				// Update counts
				productsData.counts.video_approved = productsData.products.video_approved.length;
				productsData.counts.active = productsData.products.active.length;
				productsData.counts.inactive = productsData.products.inactive.length;

				// Trigger reactivity
				productsData = { ...productsData };
			}

			showToast(
				`Product "${product.name}" ${newStatus ? 'approved for' : 'removed from'} video access`,
				'success'
			);
		} catch (err: any) {
			console.error('Error updating video approval:', err);
			showToast(err.message || 'Failed to update video approval', 'error');
		} finally {
			isUpdating = false;
		}
	}

	function toggleAccordion(section: string) {
		activeAccordion = activeAccordion === section ? null : section;
	}

	async function updateLegacyProducts() {
		if (isUpdating) return;

		try {
			isUpdating = true;
			
			const response = await apiRequest('/admin/streaming/stripe/products/update-legacy', {
				method: 'POST'
			});

			if (!response.ok) {
				throw new Error('Failed to update legacy products');
			}

			const result = await response.json();
			showToast(`Updated ${result.updated_count} products to legacy status`, 'success');
			
			// Reload products to show updated legacy tags
			await loadProducts();
		} catch (err: any) {
			console.error('Error updating legacy products:', err);
			showToast(err.message || 'Failed to update legacy products', 'error');
		} finally {
			isUpdating = false;
		}
	}

	// Helper function to get auth token (same as in subscription.ts)
	function getAuthToken(): string {
		const stored = localStorage.getItem('bome_auth_data');
		if (stored) {
			try {
				const tokenData = JSON.parse(stored);
				return tokenData.access_token || '';
			} catch (e) {
				console.error('Failed to parse bome_auth_data:', e);
			}
		}
		// Fallback to old token storage
		return localStorage.getItem('token') || '';
	}
</script>

<div class="stripe-products-section">
	<div class="subsection-header">
		<h3 class="subsection-title">🎯 Stripe Products</h3>
		<p class="subsection-description">Manage video access approval for Stripe products</p>
		{#if productsData}
			<div class="section-stats">
				<span class="stat-badge success">{activePlansProducts.length} Active Plans</span>
				<span class="stat-badge warning">{legacyPlansProducts.length} Legacy Plans</span>
				<span class="stat-badge secondary">{productsData.counts.inactive} Inactive</span>
			</div>
		{/if}
		<div class="admin-actions">
			<button 
				class="btn btn-sm btn-warning" 
				on:click={updateLegacyProducts}
				disabled={isUpdating}
			>
				{#if isUpdating}
					<div class="loading-spinner small"></div>
					Updating...
				{:else}
					🕰️ Update Legacy Products
				{/if}
			</button>
		</div>
	</div>

	{#if isLoading}
		<div class="loading-state">
			<div class="loading-spinner"></div>
			<p>Loading Stripe products...</p>
		</div>
	{:else if error}
		<div class="error-state">
			<div class="error-icon">⚠️</div>
			<h3>Error Loading Products</h3>
			<p>{error}</p>
			<button class="btn btn-primary" on:click={loadProducts}>
				Try Again
			</button>
		</div>
	{:else if productsData}
		<div class="products-accordions">
			<!-- Active Plans (Video Approved + Active Products) -->
			<SubscriptionAccordion
				title="Active Plans"
				class_title="active-plans"
				icon="<svg class='w-6 h-6' fill='none' stroke='currentColor' viewBox='0 0 24 24'><path stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z'></path></svg>"
				count={activePlansProducts.length}
				isActive={activeAccordion === 'active-plans'}
				onToggle={() => toggleAccordion('active-plans')}
				plans={activePlansProducts}
			>
				{#if activePlansProducts.length === 0}
					<div class="empty-state">
						<div class="empty-icon">🟢</div>
						<p>No active plans available</p>
						<small>Products with video approval or active status will appear here</small>
					</div>
				{:else}
					{#each activePlansProducts as product (product.id)}
						<StripeProductCard 
							{product}
							isUpdating={false}
							onToggleVideoApproval={() => handleVideoApprovalToggle(product)}
						/>
					{/each}
				{/if}
			</SubscriptionAccordion>

			<!-- Legacy Plans -->
			<SubscriptionAccordion
				title="Legacy Plans"
				class_title="legacy-plans"
				icon="<svg class='w-6 h-6' fill='none' stroke='currentColor' viewBox='0 0 24 24'><path stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z'></path></svg>"
				count={legacyPlansProducts.length}
				isActive={activeAccordion === 'legacy-plans'}
				onToggle={() => toggleAccordion('legacy-plans')}
				plans={legacyPlansProducts}
			>
				{#if legacyPlansProducts.length === 0}
					<div class="empty-state">
						<div class="empty-icon">🕰️</div>
						<p>No legacy plans available</p>
						<small>Older active products will appear here after running the legacy update</small>
					</div>
				{:else}
					{#each legacyPlansProducts as product (product.id)}
						<StripeProductCard 
							{product}
							isUpdating={false}
							onToggleVideoApproval={() => handleVideoApprovalToggle(product)}
						/>
					{/each}
				{/if}
			</SubscriptionAccordion>

			<!-- Inactive Products -->
			<SubscriptionAccordion
				title="Inactive Products"
				class_title="inactive-products"
				icon="<svg class='w-6 h-6' fill='none' stroke='currentColor' viewBox='0 0 24 24'><path stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728L5.636 5.636m12.728 12.728L18 12M6 6l12 12'></path></svg>"
				count={inactiveProducts.length}
				isActive={activeAccordion === 'inactive-products'}
				onToggle={() => toggleAccordion('inactive-products')}
				plans={inactiveProducts}
			>
				{#if inactiveProducts.length === 0}
					<div class="empty-state">
						<div class="empty-icon">⚫</div>
						<p>No inactive products</p>
					</div>
				{:else}
				{#each inactiveProducts as product (product.id)}
					<StripeProductCard 
						{product}
						isUpdating={false}
						onToggleVideoApproval={() => handleVideoApprovalToggle(product)}
					/>
				{/each}
				{/if}
			</SubscriptionAccordion>
		</div>
	{/if}
</div>

<style>
	.stripe-products-section {
		margin: 1.5rem 0 2rem 0;
		padding: 1.5rem;
		background: #f8fafc;
		border: 1px solid #e2e8f0;
		border-radius: 0.75rem;
	}

	.subsection-header {
		margin-bottom: 1.5rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid #e2e8f0;
	}

	.subsection-title {
		font-size: 1.25rem;
		font-weight: 600;
		color: #1e293b;
		margin: 0 0 0.5rem 0;
	}

	.subsection-description {
		color: #64748b;
		margin: 0 0 1rem 0;
		font-size: 0.875rem;
	}

	.admin-actions {
		margin-top: 1rem;
		display: flex;
		gap: 0.5rem;
		align-items: center;
	}

	.btn-warning {
		background: #f59e0b;
		color: white;
		border: 1px solid #d97706;
	}

	.btn-warning:hover {
		background: #d97706;
		border-color: #b45309;
	}

	.btn-warning:disabled {
		background: #fbbf24;
		border-color: #f59e0b;
		opacity: 0.6;
		cursor: not-allowed;
	}

	.section-stats {
		display: flex;
		gap: 0.75rem;
		flex-wrap: wrap;
	}

	.stat-badge {
		display: inline-flex;
		align-items: center;
		padding: 0.25rem 0.75rem;
		border-radius: 9999px;
		font-size: 0.75rem;
		font-weight: 500;
	}

	.stat-badge.success {
		background: #d1fae5;
		color: #065f46;
	}

	.stat-badge.warning {
		background: #fef3c7;
		color: #92400e;
	}

	.stat-badge.secondary {
		background: #f3f4f6;
		color: #374151;
	}

	.loading-state, .error-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 3rem 1rem;
		text-align: center;
	}

	.loading-spinner {
		width: 40px;
		height: 40px;
		border: 3px solid #e5e7eb;
		border-top: 3px solid #3b82f6;
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-bottom: 1rem;
	}

	.error-icon {
		font-size: 3rem;
		margin-bottom: 1rem;
	}

	.error-state h3 {
		margin: 0 0 0.5rem 0;
		color: #dc2626;
	}

	.error-state p {
		margin: 0 0 1rem 0;
		color: #6b7280;
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 2rem 1rem;
		text-align: center;
		color: #6b7280;
	}

	.empty-icon {
		font-size: 2rem;
		margin-bottom: 0.5rem;
		opacity: 0.7;
	}

	.empty-state p {
		margin: 0 0 0.25rem 0;
		font-weight: 500;
	}

	.empty-state small {
		font-size: 0.75rem;
		opacity: 0.8;
	}

	.products-accordions {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.btn {
		display: inline-flex;
		align-items: center;
		padding: 0.5rem 1rem;
		border: none;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
	}

	.btn-primary {
		background: #3b82f6;
		color: white;
	}

	.btn-primary:hover {
		background: #2563eb;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

</style>
