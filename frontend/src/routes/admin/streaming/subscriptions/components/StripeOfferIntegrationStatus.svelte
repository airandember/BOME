<script lang="ts">
	import { onMount } from 'svelte';
	import { StripeOfferIntegrationService, type StripeOfferIntegrationStatus } from '$lib/services/stripe-offers-integration';
	import { showToast } from '$lib/toast';

	export let offerId: string;
	export let offerName: string;

	let status: StripeOfferIntegrationStatus | null = null;
	let loading = true;
	let syncing = false;
	let lastError: string | null = null;

	onMount(async () => {
		await loadStatus();
	});

	async function loadStatus() {
		try {
			loading = true;
			const statusData = await StripeOfferIntegrationService.getStripeStatus(offerId);
			status = statusData;
			// Clear any previous errors when status loads successfully
			lastError = null;
		} catch (error: any) {
			console.error('Failed to load Stripe status:', error);
			showToast('Failed to load Stripe status', 'error');
		} finally {
			loading = false;
		}
	}

	function clearError() {
		lastError = null;
	}

	async function syncWithStripe() {
		try {
			syncing = true;
			await StripeOfferIntegrationService.syncOfferWithStripe(offerId);
			showToast('Offer synced with Stripe successfully', 'success');
			await loadStatus(); // Reload status
		} catch (error: any) {
			console.error('Failed to sync with Stripe:', error);
			
			// Extract more detailed error information
			let errorMessage = 'Unknown error occurred';
			let stripeErrorDetails = '';
			
			if (error.message) {
				errorMessage = error.message;
			} else if (error.error) {
				errorMessage = error.error;
			} else if (typeof error === 'string') {
				errorMessage = error;
			}
			
			// Extract Stripe-specific error details if available
			if (error.stripeError) {
				if (error.stripeError.detail) {
					stripeErrorDetails = error.stripeError.detail;
				} else if (error.stripeError.message) {
					stripeErrorDetails = error.stripeError.message;
				}
			}
			
			// Show detailed error in toast
			showToast(`Failed to sync with Stripe: ${errorMessage}`, 'error');
			
			// Store error details for debugging
			lastError = stripeErrorDetails || errorMessage;
		} finally {
			syncing = false;
		}
	}

	$: statusColor = status?.sync_status === 'synced' ? 'var(--success)' : 
	                 status?.sync_status === 'partial' ? 'var(--warning)' : 'var(--error)';
	$: statusText = status?.sync_status === 'synced' ? 'Fully Synced' :
	                status?.sync_status === 'partial' ? 'Partially Synced' : 'Not Synced';
</script>

<div class="stripe-status">
	<div class="status-header">
		<h4>🎟️ Stripe Integration</h4>
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
			<div class="detail-row">
				<span class="detail-label">Coupon:</span>
				<span class="detail-value">
					{#if status.has_stripe_coupon}
						✅ Created ({status.stripe_coupon_id?.slice(-8) || 'Unknown'})
					{:else}
						❌ Not created
					{/if}
				</span>
			</div>
			
			<div class="detail-row">
				<span class="detail-label">Promo Code:</span>
				<span class="detail-value">
					{#if status.has_stripe_promotion_code}
						✅ Created ({status.stripe_promotion_code_id?.slice(-8) || 'Unknown'})
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
		</div>

		{#if status.sync_status !== 'synced'}
			<div class="sync-actions">
				<button 
					class="btn btn-sm btn-primary" 
					on:click={syncWithStripe}
					disabled={syncing}
				>
					{syncing ? 'Syncing...' : 'Sync with Stripe'}
				</button>
			</div>
			
			{#if lastError}
				<div class="error-details">
					<details>
						<summary class="error-summary">⚠️ View Error Details</summary>
						<div class="error-content">
							<strong>Last Error:</strong> {lastError}
						</div>
						<div class="error-actions">
							<button 
								class="btn btn-sm btn-outline-secondary" 
								on:click={clearError}
							>
								Clear Error
							</button>
						</div>
					</details>
				</div>
			{/if}
		{:else}
			<div class="sync-success">
				<span class="success-message">✅ Offer is fully synced with Stripe</span>
			</div>
		{/if}
	{/if}
</div>

<style>
	.stripe-status {
		background: var(--card-bg);
		border: 1px solid var(--border);
		border-radius: 8px;
		padding: 1rem;
		margin-top: 1rem;
	}

	.status-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.75rem;
	}

	.status-header h4 {
		margin: 0;
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--text);
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

	.status-indicator {
		font-size: 0.875rem;
		font-weight: 500;
		display: flex;
		align-items: center;
	}

	.status-details {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		margin-bottom: 0.75rem;
	}

	.detail-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		font-size: 0.8125rem;
	}

	.detail-label {
		color: var(--text-muted);
		font-weight: 500;
	}

	.detail-value {
		color: var(--text);
		font-family: var(--font-mono);
		font-size: 0.75rem;
	}

	.sync-actions {
		display: flex;
		justify-content: flex-end;
	}

	.sync-success {
		display: flex;
		justify-content: center;
		padding: 0.5rem;
		background: var(--success-bg);
		border-radius: 4px;
	}

	.success-message {
		font-size: 0.8125rem;
		color: var(--success);
		font-weight: 500;
	}

	.btn {
		padding: 0.375rem 0.75rem;
		font-size: 0.8125rem;
		font-weight: 500;
		border: none;
		border-radius: 4px;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		white-space: nowrap;
	}

	.btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-primary {
		background: var(--primary);
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: var(--primary-hover);
		transform: translateY(-1px);
	}

	.btn-sm {
		padding: 0.25rem 0.5rem;
		font-size: 0.75rem;
	}

	.error-details {
		margin-top: 1rem;
		padding: 0.75rem;
		background: var(--error-bg, #fef2f2);
		border: 1px solid var(--error-border, #fecaca);
		border-radius: 6px;
	}

	.error-summary {
		cursor: pointer;
		font-weight: 600;
		color: var(--error, #dc2626);
		user-select: none;
	}

	.error-summary:hover {
		text-decoration: underline;
	}

	.error-content {
		margin-top: 0.5rem;
		padding: 0.5rem;
		background: var(--card-bg);
		border-radius: 4px;
		font-family: monospace;
		font-size: 0.875rem;
		white-space: pre-wrap;
		word-break: break-word;
	}

	.error-actions {
		margin-top: 0.75rem;
		text-align: right;
	}

	.btn-outline-secondary {
		background: transparent;
		border: 1px solid var(--border);
		color: var(--text-secondary);
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-size: 0.75rem;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.btn-outline-secondary:hover {
		background: var(--border);
		color: var(--text);
	}
</style> 