<script lang="ts">
	import { StripeProductsService, type StripeProduct } from '$lib/services/stripe-products';

	// Props
	export let product: StripeProduct;
	export let isUpdating: boolean = false;
	export let onToggleVideoApproval: () => void;

	// Reactive statements
	$: statusBadge = StripeProductsService.getStatusBadge(product);
	$: formattedPrice = StripeProductsService.formatPrice(product);
	$: truncatedDescription = StripeProductsService.truncateDescription(product.description, 200);
</script>

<div class="product-card" class:updating={isUpdating}>
	<div class="product-header">
		<div class="product-info">
			<h3 class="product-name">{product.name}</h3>
			<div class="product-meta">
				<span class="product-id">ID: {product.id}</span>
				<span class="product-stripe-id">Stripe: {product.stripe_id}</span>
			</div>
		</div>
		<div class="product-actions">
			<div class="status-badge {statusBadge.class}">
				<span class="badge-icon">{statusBadge.icon}</span>
				<span class="badge-text">{statusBadge.text}</span>
			</div>
		</div>
	</div>

	{#if product.description}
		<div class="product-description">
			<p>{truncatedDescription}</p>
		</div>
	{/if}

	<div class="product-details">
		<div class="detail-grid">
			<div class="detail-item">
				<span class="detail-label">Price:</span>
				<span class="detail-value">{formattedPrice}</span>
			</div>
			<div class="detail-item">
				<span class="detail-label">Status:</span>
				<span class="detail-value {product.active ? 'active' : 'inactive'}">
					{product.active ? '🟢 Active' : '⚪ Inactive'}
				</span>
			</div>
			<div class="detail-item">
				<span class="detail-label">Created:</span>
				<span class="detail-value">{new Date(product.created_at).toLocaleDateString()}</span>
			</div>
			<div class="detail-item">
				<span class="detail-label">Updated:</span>
				<span class="detail-value">{new Date(product.updated_at).toLocaleDateString()}</span>
			</div>
		</div>
	</div>

	<div class="product-footer" class:has-legacy={product.legacy}>
		<div class="video-approval-section">
			<div class="approval-info">
				<span class="approval-label">Video Access:</span>
				<span class="approval-status {product.video_approved ? 'approved' : 'not-approved'}">
					{product.video_approved ? '✅ Approved' : '❌ Not Approved'}
				</span>
			</div>
			<button 
				class="toggle-btn {product.video_approved ? 'btn-danger' : 'btn-success'}"
				disabled={isUpdating}
				on:click={onToggleVideoApproval}
			>
				{#if isUpdating}
					<div class="loading-spinner small"></div>
					Updating...
				{:else}
					{product.video_approved ? '🚫 Remove Access' : '✅ Grant Access'}
				{/if}
			</button>
		</div>
		
		{#if product.legacy}
			<div class="legacy-tag">
				Legacy Plan
			</div>
		{/if}
	</div>
</div>

<style>
	.product-card {
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 0.75rem;
		padding: 1.5rem;
		margin-bottom: 1rem;
		transition: all 0.2s ease;
		position: relative;
	}

	.product-card:hover {
		box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
		border-color: #d1d5db;
	}

	.product-card.updating {
		opacity: 0.7;
		pointer-events: none;
	}

	.product-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 1rem;
	}

	.product-info {
		flex: 1;
	}

	.product-name {
		font-size: 1.125rem;
		font-weight: 600;
		color: #111827;
		margin: 0 0 0.5rem 0;
		line-height: 1.4;
	}

	.product-meta {
		display: flex;
		gap: 1rem;
		font-size: 0.75rem;
		color: #6b7280;
	}

	.product-actions {
		margin-left: 1rem;
	}

	.status-badge {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		padding: 0.25rem 0.75rem;
		border-radius: 9999px;
		font-size: 0.75rem;
		font-weight: 500;
	}

	.status-badge.badge-success {
		background: #d1fae5;
		color: #065f46;
	}

	.status-badge.badge-warning {
		background: #fef3c7;
		color: #92400e;
	}

	.status-badge.badge-secondary {
		background: #f3f4f6;
		color: #374151;
	}

	.product-description {
		margin-bottom: 1rem;
		padding: 1rem;
		background: #f9fafb;
		border-radius: 0.5rem;
		border-left: 3px solid #3b82f6;
	}

	.product-description p {
		margin: 0;
		color: #374151;
		font-size: 0.875rem;
		line-height: 1.5;
	}

	.product-details {
		margin-bottom: 1rem;
	}

	.detail-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 0.75rem;
	}

	.detail-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.5rem;
		background: #f9fafb;
		border-radius: 0.375rem;
	}

	.detail-label {
		font-size: 0.75rem;
		font-weight: 500;
		color: #6b7280;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.detail-value {
		font-size: 0.875rem;
		font-weight: 500;
		color: #111827;
	}

	.detail-value.active {
		color: #059669;
		font-weight: 600;
	}

	.detail-value.inactive {
		color: #6b7280;
		font-weight: 500;
	}

	.product-footer {
		border-top: 1px solid #e5e7eb;
		padding-top: 1rem;
	}

	.video-approval-section {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 1rem;
	}

	.approval-info {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.approval-label {
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
	}

	.approval-status {
		font-size: 0.875rem;
		font-weight: 600;
	}

	.approval-status.approved {
		color: #059669;
	}

	.approval-status.not-approved {
		color: #dc2626;
	}

	.toggle-btn {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		border: none;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.toggle-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-success {
		background: #10b981;
		color: white;
	}

	.btn-success:hover:not(:disabled) {
		background: #059669;
	}

	.btn-danger {
		background: #ef4444;
		color: white;
	}

	.btn-danger:hover:not(:disabled) {
		background: #dc2626;
	}

	.loading-spinner {
		width: 16px;
		height: 16px;
		border: 2px solid transparent;
		border-top: 2px solid currentColor;
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	@media (max-width: 768px) {
		.product-header {
			flex-direction: column;
			gap: 1rem;
		}

		.product-actions {
			margin-left: 0;
		}

		.video-approval-section {
			flex-direction: column;
			align-items: stretch;
			gap: 0.75rem;
		}

		.detail-grid {
			grid-template-columns: 1fr;
		}
	}

	.legacy-tag {
		position: absolute;
		bottom: 0;
		left: 0;
		right: 0;
		background: linear-gradient(135deg, #f59e0b, #d97706);
		color: white;
		text-align: center;
		padding: 0.5rem;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		border-bottom-left-radius: 0.75rem;
		border-bottom-right-radius: 0.75rem;
		box-shadow: 0 -2px 4px rgba(0, 0, 0, 0.1);
		border-top: 1px solid rgba(255, 255, 255, 0.2);
	}

	.product-footer.has-legacy {
		margin-bottom: 2.5rem;
	}

</style>
