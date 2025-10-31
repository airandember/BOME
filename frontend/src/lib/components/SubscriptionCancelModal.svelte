<script lang="ts">
	import type { UserSubscription } from '$lib/services/user-subscription-service';
	import { UserSubscriptionService } from '$lib/services/user-subscription-service';

	export let isOpen: boolean = false;
	export let subscriptionsToCancel: UserSubscription[] = [];
	export let subscriptionToKeep: UserSubscription | null = null;
	export let onConfirm: (() => void) | undefined = undefined;
	export let onClose: (() => void) | undefined = undefined;
	export let isProcessing: boolean = false;

	let confirmChecked = false;

	$: totalCount = subscriptionsToCancel.length;
	$: totalMonthlySavings = subscriptionsToCancel.reduce((sum, sub) => {
		const monthlyAmount = sub.interval === 'year' ? sub.price / 12 : sub.price;
		return sum + monthlyAmount;
	}, 0);

	function handleConfirm() {
		if (confirmChecked && onConfirm) {
			onConfirm();
		}
	}

	function handleClose() {
		if (!isProcessing && onClose) {
			confirmChecked = false;
			onClose();
		}
	}

	function handleBackdropClick(event: MouseEvent) {
		if (event.target === event.currentTarget) {
			handleClose();
		}
	}
</script>

{#if isOpen}
	<div class="modal-backdrop" on:click={handleBackdropClick}>
		<div class="modal-content">
			<div class="modal-header">
				<h2>⚠️ Confirm Subscription Cancellation</h2>
				{#if !isProcessing}
					<button class="close-btn" on:click={handleClose}>×</button>
				{/if}
			</div>

			<div class="modal-body">
				{#if subscriptionToKeep}
					<div class="section keep-section">
						<h3>✅ You are keeping:</h3>
						<div class="subscription-summary keep">
							<div class="summary-name">{subscriptionToKeep.plan_name}</div>
							<div class="summary-details">
								{UserSubscriptionService.formatPrice(subscriptionToKeep.price, subscriptionToKeep.currency)}/{subscriptionToKeep.interval}
								· Renews {UserSubscriptionService.formatDate(subscriptionToKeep.current_period_end)}
							</div>
						</div>
					</div>
				{/if}

				<div class="section cancel-section">
					<h3>❌ You are canceling {totalCount === 1 ? 'this subscription' : `${totalCount} subscriptions`}:</h3>
					{#each subscriptionsToCancel as sub}
						<div class="subscription-summary cancel">
							<div class="summary-name">{sub.plan_name}</div>
							<div class="summary-details">
								{UserSubscriptionService.formatPrice(sub.price, sub.currency)}/{sub.interval}
								· Access until {UserSubscriptionService.formatDate(sub.current_period_end)}
							</div>
							<div class="summary-note">
								You'll keep access for {sub.days_until_renewal} more days
							</div>
						</div>
					{/each}
				</div>

				<div class="section savings-section">
					<div class="savings-box">
						<div class="savings-label">You'll save approximately:</div>
						<div class="savings-amount">
							{UserSubscriptionService.formatPrice(totalMonthlySavings, 'USD')}/month
						</div>
						<div class="savings-note">
							(after current billing periods end)
						</div>
					</div>
				</div>

				<div class="section confirmation-section">
					<label class="checkbox-label">
						<input
							type="checkbox"
							bind:checked={confirmChecked}
							disabled={isProcessing}
						/>
						<span>
							I understand that {totalCount === 1 ? 'this subscription will' : 'these subscriptions will'} be canceled at the end of
							{totalCount === 1 ? 'its billing period' : 'their billing periods'}
						</span>
					</label>
				</div>

				<div class="warning-box">
					<strong>⚠️ Important:</strong> You'll keep access until the end of your current billing period.
					No refunds will be issued for time remaining.
				</div>
			</div>

			<div class="modal-footer">
				<button
					class="btn btn-secondary"
					on:click={handleClose}
					disabled={isProcessing}
				>
					Go Back
				</button>
				<button
					class="btn btn-danger"
					on:click={handleConfirm}
					disabled={!confirmChecked || isProcessing}
				>
					{#if isProcessing}
						<span class="spinner"></span>
						Processing...
					{:else}
						Confirm Cancellation{totalCount > 1 ? 's' : ''}
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.modal-backdrop {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.6);
		backdrop-filter: blur(4px);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
		padding: 1rem;
		overflow-y: auto;
	}

	.modal-content {
		background: white;
		border-radius: 16px;
		max-width: 600px;
		width: 100%;
		max-height: 90vh;
		overflow-y: auto;
		box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
	}

	.modal-header {
		padding: 1.5rem;
		border-bottom: 2px solid #f3f4f6;
		display: flex;
		justify-content: space-between;
		align-items: center;
		position: sticky;
		top: 0;
		background: white;
		z-index: 10;
		border-radius: 16px 16px 0 0;
	}

	.modal-header h2 {
		margin: 0;
		font-size: 1.5rem;
		color: #111827;
	}

	.close-btn {
		background: none;
		border: none;
		font-size: 2rem;
		color: #6b7280;
		cursor: pointer;
		padding: 0;
		width: 2rem;
		height: 2rem;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 8px;
		transition: all 0.2s;
	}

	.close-btn:hover {
		background: #f3f4f6;
		color: #111827;
	}

	.modal-body {
		padding: 1.5rem;
	}

	.section {
		margin-bottom: 1.5rem;
	}

	.section h3 {
		margin: 0 0 1rem 0;
		font-size: 1.125rem;
		font-weight: 600;
		color: #374151;
	}

	.subscription-summary {
		padding: 1rem;
		border-radius: 8px;
		margin-bottom: 0.75rem;
	}

	.subscription-summary.keep {
		background: #d1fae5;
		border: 2px solid #10b981;
	}

	.subscription-summary.cancel {
		background: #fee2e2;
		border: 2px solid #ef4444;
	}

	.summary-name {
		font-weight: 700;
		font-size: 1rem;
		color: #111827;
		margin-bottom: 0.25rem;
	}

	.summary-details {
		font-size: 0.875rem;
		color: #6b7280;
		margin-bottom: 0.25rem;
	}

	.summary-note {
		font-size: 0.75rem;
		color: #9ca3af;
		font-style: italic;
	}

	.savings-section {
		background: linear-gradient(135deg, #dbeafe, #bfdbfe);
		border-radius: 12px;
		padding: 1.5rem;
		text-align: center;
	}

	.savings-box {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.savings-label {
		font-size: 0.875rem;
		color: #1e40af;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.savings-amount {
		font-size: 2rem;
		font-weight: 700;
		color: #1e3a8a;
	}

	.savings-note {
		font-size: 0.75rem;
		color: #3b82f6;
		font-style: italic;
	}

	.confirmation-section {
		background: #f9fafb;
		padding: 1rem;
		border-radius: 8px;
		border: 2px solid #e5e7eb;
	}

	.checkbox-label {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		cursor: pointer;
		font-size: 0.875rem;
		color: #374151;
		line-height: 1.5;
	}

	.checkbox-label input[type='checkbox'] {
		margin-top: 0.125rem;
		width: 1.25rem;
		height: 1.25rem;
		cursor: pointer;
		flex-shrink: 0;
	}

	.checkbox-label input[type='checkbox']:disabled {
		cursor: not-allowed;
	}

	.warning-box {
		background: #fef3c7;
		border: 2px solid #fbbf24;
		border-radius: 8px;
		padding: 1rem;
		font-size: 0.875rem;
		color: #78350f;
		line-height: 1.5;
	}

	.warning-box strong {
		display: block;
		margin-bottom: 0.5rem;
	}

	.modal-footer {
		padding: 1.5rem;
		border-top: 2px solid #f3f4f6;
		display: flex;
		gap: 1rem;
		justify-content: flex-end;
		position: sticky;
		bottom: 0;
		background: white;
		border-radius: 0 0 16px 16px;
	}

	.btn {
		padding: 0.75rem 1.5rem;
		border-radius: 8px;
		font-weight: 600;
		font-size: 0.875rem;
		border: none;
		cursor: pointer;
		transition: all 0.2s;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		justify-content: center;
		min-width: 120px;
	}

	.btn:disabled {
		cursor: not-allowed;
		opacity: 0.6;
	}

	.btn-secondary {
		background: #f3f4f6;
		color: #374151;
		border: 2px solid #d1d5db;
	}

	.btn-secondary:hover:not(:disabled) {
		background: #e5e7eb;
		transform: translateY(-2px);
	}

	.btn-danger {
		background: linear-gradient(135deg, #ef4444, #dc2626);
		color: white;
	}

	.btn-danger:hover:not(:disabled) {
		transform: translateY(-2px);
		box-shadow: 0 8px 16px rgba(239, 68, 68, 0.3);
	}

	.spinner {
		width: 1rem;
		height: 1rem;
		border: 2px solid rgba(255, 255, 255, 0.3);
		border-top-color: white;
		border-radius: 50%;
		animation: spin 0.6s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	@media (max-width: 640px) {
		.modal-content {
			margin: 1rem;
		}

		.modal-header h2 {
			font-size: 1.25rem;
		}

		.savings-amount {
			font-size: 1.5rem;
		}

		.modal-footer {
			flex-direction: column-reverse;
		}

		.btn {
			width: 100%;
		}
	}
</style>

