<script lang="ts">
	import type { SubscriptionPlan } from '$lib/services/streaming-subscriptions';

	export let subscriptionPlans: SubscriptionPlan[];
	export let onCreateClick: () => void;

	// Reactive statistics
	$: totalPlans = subscriptionPlans?.length || 0;
	$: activePlansCount = subscriptionPlans?.filter(p => p?.is_active)?.length || 0;
	$: promotedPlansCount = subscriptionPlans?.filter(p => p?.is_promoted)?.length || 0;
	$: inactivePlansCount = subscriptionPlans?.filter(p => !p?.is_active)?.length || 0;
</script>

<div class="subscription-header">
	<div class="header-content">
		<div class="header-title">
			<h1>Subscription Plans</h1>
			<p>Manage your streaming subscription plans and promotions</p>
		</div>
		
		<div class="subscription-stats">
			<div class="stat-item">
				<span class="stat-number">{totalPlans}</span>
				<span class="stat-label">Total Plans</span>
			</div>
			<div class="stat-item">
				<span class="stat-number">{activePlansCount}</span>
				<span class="stat-label">Active</span>
			</div>
			<div class="stat-item">
				<span class="stat-number">{promotedPlansCount}</span>
				<span class="stat-label">Promoted</span>
			</div>
			<div class="stat-item">
				<span class="stat-number">{inactivePlansCount}</span>
				<span class="stat-label">Inactive</span>
			</div>
		</div>
	</div>
	
	<button class="btn btn-primary" on:click={onCreateClick}>
		<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
		</svg>
		Create Plan
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

	.header-title h1 {
		color: #111827;
		font-size: 2rem;
		font-weight: 700;
		margin: 0 0 0.5rem 0;
	}

	.header-title p {
		color: #6b7280;
		margin: 0;
		font-size: 1rem;
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
		margin-bottom: 0.25rem;
	}

	.stat-label {
		display: block;
		color: #6b7280;
		font-size: 0.875rem;
		font-weight: 500;
	}

	.btn {
		padding: 0.75rem 1.5rem;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		border: none;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
	}

	.btn:hover {
		transform: translateY(-1px);
	}

	.btn:focus {
		outline: 2px solid #2563eb;
		outline-offset: 2px;
	}

	.btn-primary {
		background: #2563eb;
		color: white;
	}

	.btn-primary:hover {
		background: #1d4ed8;
		box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.subscription-header {
			padding: 1rem;
			flex-direction: column;
			align-items: flex-start;
			gap: 1rem;
		}

		.header-content {
			flex-direction: column;
			align-items: flex-start;
			gap: 1rem;
			width: 100%;
		}

		.subscription-stats {
			margin-right: 0;
			align-self: flex-end;
			gap: 1rem;
		}

		.stat-item {
			min-width: 60px;
			padding: 0.75rem;
		}

		.stat-number {
			font-size: 1.25rem;
		}

		.stat-label {
			font-size: 0.75rem;
		}
	}
</style> 