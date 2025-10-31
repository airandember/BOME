<script lang="ts">
	import type { UserSubscription } from '$lib/services/user-subscription-service';
	import { UserSubscriptionService } from '$lib/services/user-subscription-service';

	export let subscription: UserSubscription;
	export let canSelect: boolean = false; // Show "Keep This One" button
	export let isSelected: boolean = false; // Is this the selected subscription to keep?
	export let onSelect: ((id: string) => void) | undefined = undefined;
	export let onCancel: ((id: string) => void) | undefined = undefined;

	$: statusColor = UserSubscriptionService.getStatusColor(subscription.status);
	$: daysLeftColor = UserSubscriptionService.getDaysLeftColor(subscription.days_until_renewal);
	$: formattedPrice = UserSubscriptionService.formatPrice(subscription.price, subscription.currency);
	$: formattedInterval = UserSubscriptionService.formatInterval(subscription.interval);
	$: formattedStart = UserSubscriptionService.formatDate(subscription.current_period_start);
	$: formattedEnd = UserSubscriptionService.formatDate(subscription.current_period_end);

	function handleSelect() {
		if (onSelect) {
			onSelect(subscription.id);
		}
	}

	function handleCancel() {
		if (onCancel) {
			onCancel(subscription.id);
		}
	}
</script>

<div class="subscription-card" class:selected={isSelected}>
	{#if subscription.is_primary}
		<div class="primary-badge">Primary</div>
	{/if}

	<div class="card-header">
		<h3 class="plan-name">
			{subscription.plan_name}
			{#if subscription.is_lifetime}
				<span class="lifetime-badge">Lifetime</span>
			{/if}
		</h3>
		<span class="status-badge" data-status={subscription.status} data-color={statusColor}>
			{subscription.status === 'active' ? '✅' : subscription.status === 'trialing' ? '🆓' : subscription.status === 'canceled' ? '❌' : '⚠️'}
			{subscription.status.charAt(0).toUpperCase() + subscription.status.slice(1)}
		</span>
	</div>

	<div class="card-body">
		<div class="info-row">
			<span class="info-label">Price:</span>
			<span class="info-value">{formattedPrice}/{subscription.interval}</span>
		</div>

		<div class="info-row">
			<span class="info-label">Started:</span>
			<span class="info-value">{formattedStart}</span>
		</div>

		{#if subscription.cancel_at_period_end}
			<div class="info-row warning">
				<span class="info-label">⚠️ Ending:</span>
				<span class="info-value">{formattedEnd}</span>
			</div>
			<div class="info-row warning">
				<span class="info-label">Days Left:</span>
				<span class="info-value">{subscription.days_until_renewal} days</span>
			</div>
		{:else if subscription.status === 'active' || subscription.status === 'trialing'}
			<div class="info-row">
				<span class="info-label">Renews:</span>
				<span class="info-value">{formattedEnd}</span>
			</div>
			<div class="info-row">
				<span class="info-label">Days Until Renewal:</span>
				<span class="info-value" data-color={daysLeftColor}>
					{subscription.days_until_renewal} days
				</span>
			</div>
		{:else}
			<div class="info-row">
				<span class="info-label">Ended:</span>
				<span class="info-value">{formattedEnd}</span>
			</div>
		{/if}

		{#if subscription.canceled_at}
			<div class="info-row">
				<span class="info-label">Canceled:</span>
				<span class="info-value">{UserSubscriptionService.formatDate(subscription.canceled_at)}</span>
			</div>
		{/if}
	</div>

	{#if canSelect && !subscription.cancel_at_period_end}
		<div class="card-actions">
			{#if isSelected}
				<button class="btn btn-success" disabled>
					⭐ Keeping This One
				</button>
			{:else}
				<button class="btn btn-outline" on:click={handleSelect}>
					⭐ Keep This One
				</button>
				{#if onCancel}
					<button class="btn btn-danger" on:click={handleCancel}>
						❌ Cancel
					</button>
				{/if}
			{/if}
		</div>
	{:else if subscription.cancel_at_period_end}
		<div class="card-actions">
			<button class="btn btn-warning" disabled>
				⏳ Canceling at Period End
			</button>
		</div>
	{:else if subscription.status === 'active' && onCancel}
		<div class="card-actions">
			<button class="btn btn-danger" on:click={handleCancel}>
				❌ Cancel Subscription
			</button>
		</div>
	{/if}
</div>

<style>
	.subscription-card {
		background: white;
		border: 2px solid #e5e7eb;
		border-radius: 12px;
		padding: 1.5rem;
		margin-bottom: 1rem;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
		transition: all 0.2s;
		position: relative;
	}

	.subscription-card:hover {
		border-color: #3b82f6;
		box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
	}

	.subscription-card.selected {
		border-color: #10b981;
		background: #f0fdf4;
		box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.2);
	}

	.primary-badge {
		position: absolute;
		top: 1rem;
		right: 1rem;
		background: linear-gradient(135deg, #fbbf24, #f59e0b);
		color: white;
		padding: 0.25rem 0.75rem;
		border-radius: 6px;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.lifetime-badge {
		background: linear-gradient(135deg, #8b5cf6, #7c3aed);
		color: white;
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-size: 0.75rem;
		font-weight: 600;
		margin-left: 0.5rem;
	}

	.card-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 1rem;
		gap: 1rem;
	}

	.plan-name {
		font-size: 1.25rem;
		font-weight: 700;
		color: #111827;
		margin: 0;
		display: flex;
		align-items: center;
		flex-wrap: wrap;
	}

	.status-badge {
		padding: 0.375rem 0.75rem;
		border-radius: 6px;
		font-size: 0.875rem;
		font-weight: 600;
		white-space: nowrap;
	}

	.status-badge[data-color='green'] {
		background: #d1fae5;
		color: #065f46;
	}

	.status-badge[data-color='blue'] {
		background: #dbeafe;
		color: #1e40af;
	}

	.status-badge[data-color='gray'] {
		background: #f3f4f6;
		color: #374151;
	}

	.status-badge[data-color='orange'] {
		background: #fed7aa;
		color: #9a3412;
	}

	.status-badge[data-color='red'] {
		background: #fecaca;
		color: #991b1b;
	}

	.status-badge[data-color='yellow'] {
		background: #fef3c7;
		color: #92400e;
	}

	.card-body {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		margin-bottom: 1rem;
	}

	.info-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.5rem 0;
		border-bottom: 1px solid #f3f4f6;
	}

	.info-row:last-child {
		border-bottom: none;
	}

	.info-row.warning {
		background: #fef3c7;
		padding: 0.5rem;
		border-radius: 6px;
		border-bottom: none;
		margin-top: 0.25rem;
	}

	.info-label {
		font-weight: 600;
		color: #6b7280;
		font-size: 0.875rem;
	}

	.info-value {
		font-weight: 500;
		color: #111827;
		font-size: 0.875rem;
	}

	.info-value[data-color='red'] {
		color: #dc2626;
		font-weight: 700;
	}

	.info-value[data-color='orange'] {
		color: #ea580c;
		font-weight: 600;
	}

	.info-value[data-color='yellow'] {
		color: #d97706;
	}

	.info-value[data-color='green'] {
		color: #059669;
	}

	.card-actions {
		display: flex;
		gap: 0.75rem;
		flex-wrap: wrap;
	}

	.btn {
		padding: 0.625rem 1.25rem;
		border-radius: 8px;
		font-weight: 600;
		font-size: 0.875rem;
		border: none;
		cursor: pointer;
		transition: all 0.2s;
		flex: 1;
		min-width: fit-content;
	}

	.btn:disabled {
		cursor: not-allowed;
		opacity: 0.6;
	}

	.btn-success {
		background: linear-gradient(135deg, #10b981, #059669);
		color: white;
	}

	.btn-success:hover:not(:disabled) {
		transform: translateY(-2px);
		box-shadow: 0 4px 8px rgba(16, 185, 129, 0.3);
	}

	.btn-outline {
		background: white;
		border: 2px solid #3b82f6;
		color: #3b82f6;
	}

	.btn-outline:hover:not(:disabled) {
		background: #3b82f6;
		color: white;
		transform: translateY(-2px);
		box-shadow: 0 4px 8px rgba(59, 130, 246, 0.3);
	}

	.btn-danger {
		background: white;
		border: 2px solid #ef4444;
		color: #ef4444;
	}

	.btn-danger:hover:not(:disabled) {
		background: #ef4444;
		color: white;
		transform: translateY(-2px);
		box-shadow: 0 4px 8px rgba(239, 68, 68, 0.3);
	}

	.btn-warning {
		background: #fbbf24;
		color: #78350f;
	}

	@media (max-width: 640px) {
		.subscription-card {
			padding: 1rem;
		}

		.card-header {
			flex-direction: column;
			gap: 0.5rem;
		}

		.plan-name {
			font-size: 1.125rem;
		}

		.card-actions {
			flex-direction: column;
		}

		.btn {
			width: 100%;
		}
	}
</style>

