<script lang="ts">
	import { onMount } from 'svelte';
	import { StripeSubscriptionIntegrationService, type StripeIntegrationStatus } from '$lib/services/stripe-subscription-integration';
	import { showToast } from '$lib/toast';

	export let planId: string;
	export let planName: string;

	let status: StripeIntegrationStatus | null = null;
	let loading = true;
	let syncing = false;

	// Check if this is a Stripe product (not a regular subscription plan)
	$: isStripeProduct = planId.startsWith('stripe_');

	onMount(async () => {
		if (!isStripeProduct) {
			await loadStatus();
		} else {
			// For Stripe products, we don't need to load status since they're already from Stripe
			loading = false;
			status = {
				sync_status: 'synced',
				stripe_product_id: planId.replace('stripe_', ''),
				stripe_price_id: null,
				last_synced: new Date().toISOString(),
				sync_errors: []
			};
		}
	});

	async function loadStatus() {
		try {
			loading = true;
			status = await StripeSubscriptionIntegrationService.getStripeStatus(planId);
		} catch (error) {
			console.error('Failed to load Stripe status:', error);
		} finally {
			loading = false;
		}
	}

	async function syncWithStripe() {
		if (isStripeProduct) {
			showToast('Stripe products are already synced from Stripe', 'info');
			return;
		}

		try {
			syncing = true;
			await StripeSubscriptionIntegrationService.syncPlanWithStripe(planId);
			showToast('Plan synced with Stripe successfully', 'success');
			await loadStatus(); // Reload status
		} catch (error: any) {
			console.error('Failed to sync with Stripe:', error);
			showToast(`Failed to sync with Stripe: ${error.message || 'Unknown error'}`, 'error');
		} finally {
			syncing = false;
		}
	}

	$: statusColor = status?.sync_status === 'synced' ? 'var(--success)' : 
	                 status?.sync_status === 'partial' ? 'var(--warning)' : 'var(--error)';
	$: statusText = isStripeProduct ? 'Native Stripe Product' :
	                status?.sync_status === 'synced' ? 'Fully Synced' :
	                status?.sync_status === 'partial' ? 'Partially Synced' : 'Not Synced';
</script>

<div class="stripe-status">
	<div class="status-header">
		<h4>🔗 Stripe Integration</h4>
		{#if loading}
			<div class="loading-spinner"></div>
		{:else if status}
			<div class="status-indicator" style="color: {statusColor}">
				●&nbsp;{statusText}
			</div>
		{/if}
	</div>

	{#if status && !loading}
		<div class="status-details">
			{#if isStripeProduct}
				<!-- Display for Stripe Products -->
				<div class="detail-row">
					<span class="detail-label">Source:</span>
					<span class="detail-value">
						🔗 Native Stripe Product
					</span>
				</div>
				
				<div class="detail-row">
					<span class="detail-label">Product ID:</span>
					<span class="detail-value">
						{status.stripe_product_id}
					</span>
				</div>

				<div class="detail-row">
					<span class="detail-label">Status:</span>
					<span class="detail-value">
						✅ Automatically synchronized
					</span>
				</div>
			{:else}
				<!-- Display for Regular Subscription Plans -->
				<div class="detail-row">
					<span class="detail-label">Product:</span>
					<span class="detail-value">
						{#if status.has_stripe_product}
							✅ Created ({status.stripe_product_id?.slice(-8) || 'Unknown'})
						{:else}
							❌ Not created
						{/if}
					</span>
				</div>
				
				<div class="detail-row">
					<span class="detail-label">Price:</span>
					<span class="detail-value">
						{#if status.has_stripe_price}
							✅ Created ({status.stripe_price_id?.slice(-8) || 'Unknown'})
						{:else}
							❌ Not created
						{/if}
					</span>
				</div>

				{#if status.last_synced}
					<div class="detail-row">
						<span class="detail-label">Last Synced:</span>
						<span class="detail-value">{new Date(status.last_synced).toLocaleString()}</span>
					</div>
				{/if}
			{/if}
		</div>

		{#if status.sync_status !== 'synced' && !isStripeProduct}
			<div class="sync-actions">
				<button 
					class="btn btn-sm btn-primary" 
					on:click={syncWithStripe}
					disabled={syncing}
				>
					{syncing ? 'Syncing...' : 'Sync with Stripe'}
				</button>
			</div>
		{/if}
	{/if}
</div>

<style>
	.stripe-status {
		margin-top: var(--space-md);
		padding: var(--space-md);
		background: var(--surface-secondary);
		border-radius: var(--radius-md);
		border: 1px solid var(--border);
	}

	.status-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-sm);
	}

	.status-header h4 {
		margin: 0;
		font-size: 1rem;
		color: var(--text);
	}

	.status-indicator {
		font-size: 0.9rem;
		font-weight: 600;
	}

	.status-details {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
		margin-bottom: var(--space-md);
	}

	.detail-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		font-size: 0.9rem;
	}

	.detail-label {
		color: var(--text-muted);
		font-weight: 500;
	}

	.detail-value {
		color: var(--text);
		font-family: monospace;
	}

	.sync-actions {
		text-align: center;
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

	.btn-sm {
		padding: var(--space-xs) var(--space-sm);
		font-size: 0.8rem;
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
		background: var(--border);
		color: var(--text-muted);
		cursor: not-allowed;
		transform: none;
	}

	.loading-spinner {
		width: 16px;
		height: 16px;
		border: 2px solid var(--border);
		border-top: 2px solid var(--primary);
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}
</style> 
