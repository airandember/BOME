<script lang="ts">
	import type { SubscriptionOffer } from '$lib/services/subscription-offers';
	import type { SubscriptionPlan } from '$lib/services/streaming-subscriptions';
	import { formatDateForDisplay } from '$lib/utils/date';
	import StripeOfferIntegrationStatus from './StripeOfferIntegrationStatus.svelte';

	export let offer: SubscriptionOffer;
	export let subscriptionPlans: SubscriptionPlan[] = [];
	export let showStripeStatus: boolean = true; // Default to true like PlanCard
	export let isOptimisticallyUpdating: (id: string | number) => boolean = () => false;
	export let onEdit: (offer: SubscriptionOffer) => void = () => {};
	export let onToggleStatus: (offer: SubscriptionOffer) => void = () => {};
	export let onViewDetails: (offer: SubscriptionOffer) => void = () => {};

	// Utility functions with proper plan lookup
	function getPlanName(planId: number): string {
		const plan = subscriptionPlans.find(p => p.id === planId.toString());
		return plan ? plan.name : `Plan ${planId}`;
	}

	function getItemName(itemId: string | null | undefined): string {
		if (!itemId) return 'No specific item';
		
		const itemTypes: Record<string, string> = {
			'ebook': 'eBook',
			'dvd': 'DVD', 
			'expo_ticket': 'Expo Ticket',
			'1': 'eBook',
			'2': 'DVD',
			'3': 'Expo Ticket'
		};
		
		return itemTypes[itemId] || itemId;
	}

	// Format discount display
	$: discountDisplay = offer.off_discount_type === 'percentage' 
		? `${offer.off_discount_value}% OFF`
		: `$${offer.off_discount_value} OFF`;

	// Format usage display
	$: usageDisplay = `${offer.off_current_uses}/${offer.off_max_uses || '∞'}`;

	// Get status color
	$: statusColor = offer.is_active ? 'text-green-600' : 'text-red-600';
	$: statusBg = offer.is_active ? 'bg-green-100' : 'bg-red-100';

	// Get priority color
	$: priorityColor = offer.off_priority <= 3 ? 'text-red-600' : 
					   offer.off_priority <= 6 ? 'text-yellow-600' : 'text-green-600';
</script>

<div class="offer-card" class:updating={isOptimisticallyUpdating(offer.id)}>
	<div class="offer-header">
		<div class="offer-title-section">
			<h3 class="offer-title">{offer.off_name}</h3>
			<div class="offer-badges">
				<span class="badge badge-discount">{discountDisplay}</span>
				<span class="badge badge-priority {priorityColor}">Priority {offer.off_priority}</span>
				{#if offer.off_auto_apply}
					<span class="badge badge-auto">Auto-Apply</span>
				{/if}
				{#if offer.item_id}
					<span class="badge badge-item">{getItemName(offer.item_id)}</span>
				{/if}
			</div>
		</div>
		<div class="offer-status">
			<span class="status-indicator {statusColor} {statusBg}">
				{offer.is_active ? 'Active' : 'Inactive'}
			</span>
		</div>
	</div>

	<div class="offer-content">
		{#if offer.off_description}
			<p class="offer-description">{offer.off_description}</p>
		{/if}

		<div class="offer-details">
			<div class="detail-item">
				<span class="detail-label">Plan:</span>
				<span class="detail-value">{getPlanName(offer.plan_id)}</span>
			</div>
			{#if offer.item_id}
				<div class="detail-item">
					<span class="detail-label">Item Type:</span>
					<span class="detail-value item-type">{getItemName(offer.item_id)}</span>
				</div>
			{/if}
			<div class="detail-item">
				<span class="detail-label">Usage:</span>
				<span class="detail-value">{usageDisplay}</span>
			</div>
			{#if offer.off_code}
				<div class="detail-item">
					<span class="detail-label">Code:</span>
					<span class="detail-value code">{offer.off_code}</span>
				</div>
			{/if}
			{#if offer.offer_start_date || offer.off_end_date}
				<div class="detail-item">
					<span class="detail-label">Valid:</span>
					<span class="detail-value">
						{offer.offer_start_date ? formatDateForDisplay(offer.offer_start_date) : 'No start'}
						{' → '}
						{offer.off_end_date ? formatDateForDisplay(offer.off_end_date) : 'No end'}
					</span>
				</div>
			{/if}
		</div>
	</div>

	<!-- Stripe Integration Status -->
	{#if showStripeStatus}
		<StripeOfferIntegrationStatus 
			offerId={offer.id.toString()} 
			offerName={offer.off_name}
		/>
	{/if}

	<div class="offer-actions">
		<button
			class="action-btn action-view"
			on:click={() => onViewDetails(offer)}
			title="View Details"
		>
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
			</svg>
			Details
		</button>

		<button
			class="action-btn action-edit"
			on:click={() => onEdit(offer)}
			title="Edit Offer"
		>
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
			</svg>
			Edit
		</button>

		<button
			class="action-btn action-toggle"
			class:active={offer.is_active}
			on:click={() => onToggleStatus(offer)}
			title={offer.is_active ? 'Deactivate' : 'Activate'}
		>
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
			</svg>
			{offer.is_active ? 'Active' : 'Inactive'}
		</button>
	</div>
</div>

<style>
	.offer-card {
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		padding: 1rem;
		transition: all 0.2s ease;
		position: relative;
	}

	.offer-card:hover {
		border-color: #d1d5db;
		box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
	}

	.offer-card.updating {
		opacity: 0.6;
		pointer-events: none;
	}

	.offer-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 1rem;
	}

	.offer-title-section {
		flex: 1;
	}

	.offer-title {
		font-size: 1.125rem;
		font-weight: 600;
		color: #111827;
		margin: 0 0 0.5rem 0;
	}

	.offer-badges {
		display: flex;
		gap: 0.5rem;
		flex-wrap: wrap;
	}

	.badge {
		padding: 0.25rem 0.5rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.badge-discount {
		background: #fef3c7;
		color: #92400e;
	}

	.badge-priority {
		background: #f3f4f6;
	}

	.badge-auto {
		background: #dbeafe;
		color: #1e40af;
	}

	.badge-item {
		background: #f0f9ff;
		color: #0369a1;
	}

	.offer-status {
		display: flex;
		align-items: center;
	}

	.status-indicator {
		padding: 0.25rem 0.75rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.offer-content {
		margin-bottom: 1rem;
	}

	.offer-description {
		color: #6b7280;
		font-size: 0.875rem;
		line-height: 1.5;
		margin: 0 0 1rem 0;
	}

	.offer-details {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.detail-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		font-size: 0.875rem;
	}

	.detail-label {
		color: #6b7280;
		font-weight: 500;
	}

	.detail-value {
		color: #111827;
		font-weight: 600;
	}

	.detail-value.code {
		font-family: monospace;
		background: #f3f4f6;
		padding: 0.125rem 0.375rem;
		border-radius: 0.25rem;
	}

	.detail-value.item-type {
		background: #e0f2fe;
		color: #1e40af;
		padding: 0.125rem 0.375rem;
		border-radius: 0.25rem;
		font-weight: 600;
		text-transform: capitalize;
	}

	.offer-actions {
		display: flex;
		gap: 0.5rem;
		justify-content: flex-end;
	}

	.action-btn {
		display: flex;
		align-items: center;
		gap: 0.25rem;
		padding: 0.375rem 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		background: white;
		color: #374151;
		font-size: 0.75rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.action-btn:hover {
		background: #f9fafb;
		border-color: #9ca3af;
	}

	.action-btn.active {
		background: #dcfce7;
		border-color: #16a34a;
		color: #166534;
	}

	.action-view {
		color: #2563eb;
		border-color: #bfdbfe;
	}

	.action-view:hover {
		background: #eff6ff;
		border-color: #93c5fd;
	}

	.action-edit {
		color: #7c3aed;
		border-color: #ddd6fe;
	}

	.action-edit:hover {
		background: #faf5ff;
		border-color: #c4b5fd;
	}

	.action-toggle {
		color: #059669;
		border-color: #a7f3d0;
	}

	.action-toggle:hover {
		background: #f0fdf4;
		border-color: #86efac;
	}

	/* Responsive Design */
	@media (max-width: 640px) {
		.offer-header {
			flex-direction: column;
			align-items: flex-start;
			gap: 0.5rem;
		}

		.offer-actions {
			flex-direction: column;
		}

		.action-btn {
			justify-content: center;
		}
	}
</style> 