<script lang="ts">
	import type { SubscriptionOffer } from '$lib/services/subscription-offers';

	export let subscriptionOffers: SubscriptionOffer[] = [];
	export let onCreateClick: () => void = () => {};

	// Calculate statistics
	$: totalOffers = subscriptionOffers.length;
	$: activeOffers = subscriptionOffers.filter(offer => offer.is_active).length;
	$: inactiveOffers = subscriptionOffers.filter(offer => !offer.is_active).length;
	$: totalUses = subscriptionOffers.reduce((sum, offer) => sum + offer.off_current_uses, 0);
	$: averagePriority = subscriptionOffers.length > 0 
		? Math.round(subscriptionOffers.reduce((sum, offer) => sum + offer.off_priority, 0) / subscriptionOffers.length)
		: 0;
</script>

<div class="subscription-header p-0">
	<div class="header-content p-0">
		<div class="subscription-stats">
			<div class="stat-item">
				<span class="stat-number">{activeOffers}</span>
				<span class="stat-label">Active</span>
			</div>
			<div class="stat-item">
				<span class="stat-number">{inactiveOffers}</span>
				<span class="stat-label">Inactive</span>
			</div>
			<div class="stat-item">
				<span class="stat-number">{totalUses}</span>
				<span class="stat-label">Total Uses</span>
			</div>
			<div class="stat-item">
				<span class="stat-number">{totalOffers}</span>
				<span class="stat-label">Total Offers</span>
			</div>
		</div>
	</div>
	
	<button class="btn btn-primary" on:click={onCreateClick}>
		<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
		</svg>
		Create Offer
	</button>
</div>

<style>
	.subscription-header {
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		padding: 2rem;
		margin-bottom: 2rem;
		display: flex;
		align-items: center;
		justify-content: space-between;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
	}

	.header-content {
		display: flex;
		align-items: center;
		gap: 3rem;
	}

	.subscription-stats {
		display: flex;
		gap: 2rem;
	}

	.stat-item {
		text-align: center;
		padding: 1rem;
		background: #f9fafb;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		min-width: 80px;
	}

	.stat-number {
		display: block;
		color: #111827;
		font-size: 1.5rem;
		font-weight: 700;
		line-height: 1;
	}

	.stat-label {
		display: block;
		color: #6b7280;
		font-size: 0.875rem;
		font-weight: 500;
		margin-top: 0.25rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.btn {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1.5rem;
		border-radius: 0.375rem;
		font-weight: 600;
		text-decoration: none;
		transition: all 0.2s ease;
		border: none;
		cursor: pointer;
		font-size: 0.875rem;
	}

	.btn-primary {
		background: #2563eb;
		color: white;
	}

	.btn-primary:hover {
		background: #1d4ed8;
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
	}

	.btn:focus {
		outline: 2px solid #2563eb;
		outline-offset: 2px;
	}

	@media (max-width: 768px) {
		.subscription-header {
			flex-direction: column;
			gap: 1rem;
			text-align: center;
		}

		.subscription-stats {
			justify-content: center;
			flex-wrap: wrap;
		}

		.stat-item {
			min-width: 70px;
		}
	}
</style> 
