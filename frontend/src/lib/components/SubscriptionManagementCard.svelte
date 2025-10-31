<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import LoadingSpinner from './LoadingSpinner.svelte';
	import type { UserSubscription } from '$lib/services/user-subscription-service';
	import { UserSubscriptionService } from '$lib/services/user-subscription-service';

	let loading = true;
	let subscriptions: UserSubscription[] = [];
	let activeSubscription: UserSubscription | null = null;
	let hasMultipleActive = false;

	onMount(async () => {
		await loadSubscriptions();
	});

	async function loadSubscriptions() {
		try {
			loading = true;
			const response = await UserSubscriptionService.getUserSubscriptions();
			subscriptions = response.subscriptions || [];
			
			// Find active subscriptions
			const activeSubs = subscriptions.filter(s => s.status === 'active' || s.status === 'trialing');
			hasMultipleActive = activeSubs.length > 1;
			activeSubscription = activeSubs[0] || null;
		} catch (err) {
			console.error('Failed to load subscriptions:', err);
		} finally {
			loading = false;
		}
	}

	function formatPrice(price: number, currency: string): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: currency.toUpperCase(),
		}).format(price / 100);
	}

	function formatDate(dateStr: string | undefined): string {
		if (!dateStr) return 'N/A';
		return new Date(dateStr).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
		});
	}
</script>

<div class="subscription-card">
	<div class="card-header">
		<h2>Subscription</h2>
		<button class="manage-link" on:click={() => goto('/user/subscriptions')}>
			View Details →
		</button>
	</div>

	{#if loading}
		<div class="loading-state">
			<LoadingSpinner />
			<p>Loading subscription...</p>
		</div>
	{:else if hasMultipleActive}
		<div class="warning-state">
			<svg class="warning-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
				<line x1="12" y1="9" x2="12" y2="13"/>
				<line x1="12" y1="17" x2="12.01" y2="17"/>
			</svg>
			<p class="warning-text">Multiple active subscriptions detected</p>
			<button class="btn-warning" on:click={() => goto('/user/subscriptions')}>
				Manage Subscriptions
			</button>
		</div>
	{:else if activeSubscription}
		<div class="active-subscription">
			<div class="sub-status">
				<span class="status-badge {activeSubscription.status}">
					{activeSubscription.status === 'active' ? '✓ Active' : 
					 activeSubscription.status === 'trialing' ? '🎁 Trial' : activeSubscription.status}
				</span>
			</div>
			
			<div class="sub-info">
				<div class="info-item">
					<label>Plan:</label>
					<span class="value">{activeSubscription.product_name}</span>
				</div>
				<div class="info-item">
					<label>Price:</label>
					<span class="value">
						{formatPrice(activeSubscription.price, activeSubscription.currency)}
						/{activeSubscription.interval}
					</span>
				</div>
				<div class="info-item">
					<label>Next billing:</label>
					<span class="value">{formatDate(activeSubscription.current_period_end)}</span>
				</div>
			</div>

			<div class="sub-actions">
				<button class="btn-secondary" on:click={() => goto('/user/subscriptions')}>
					Change Plan
				</button>
			</div>
		</div>
	{:else}
		<div class="no-subscription">
			<svg class="info-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<circle cx="12" cy="12" r="10"/>
				<line x1="12" y1="16" x2="12" y2="12"/>
				<line x1="12" y1="8" x2="12.01" y2="8"/>
			</svg>
			<p>No active subscription</p>
			<button class="btn-primary" on:click={() => goto('/subscription')}>
				View Plans
			</button>
		</div>
	{/if}
</div>

<style>
	.subscription-card {
		background: var(--glass-bg, rgba(255, 255, 255, 0.05));
		border: 1px solid var(--glass-border, rgba(255, 255, 255, 0.1));
		border-radius: 12px;
		padding: 1.5rem;
		backdrop-filter: blur(10px);
	}

	.card-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1.5rem;
	}

	.card-header h2 {
		margin: 0;
		font-size: 1.25rem;
		font-weight: 600;
	}

	.manage-link {
		background: none;
		border: none;
		color: var(--primary-color, #3b82f6);
		cursor: pointer;
		font-size: 0.875rem;
		font-weight: 500;
		transition: opacity 0.2s;
	}

	.manage-link:hover {
		opacity: 0.8;
	}

	.loading-state,
	.warning-state,
	.no-subscription {
		text-align: center;
		padding: 2rem 1rem;
	}

	.warning-state {
		background: rgba(251, 191, 36, 0.1);
		border-radius: 8px;
	}

	.warning-icon,
	.info-icon {
		width: 48px;
		height: 48px;
		margin: 0 auto 1rem;
		stroke: var(--warning-color, #f59e0b);
	}

	.info-icon {
		stroke: var(--text-secondary, rgba(255, 255, 255, 0.6));
	}

	.warning-text {
		color: var(--warning-color, #f59e0b);
		font-weight: 500;
		margin-bottom: 1rem;
	}

	.btn-warning {
		background: var(--warning-color, #f59e0b);
		color: white;
		border: none;
		padding: 0.75rem 1.5rem;
		border-radius: 8px;
		font-weight: 500;
		cursor: pointer;
		transition: transform 0.2s, opacity 0.2s;
	}

	.btn-warning:hover {
		transform: translateY(-2px);
		opacity: 0.9;
	}

	.active-subscription {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.sub-status {
		display: flex;
		justify-content: center;
	}

	.status-badge {
		display: inline-block;
		padding: 0.5rem 1rem;
		border-radius: 20px;
		font-size: 0.875rem;
		font-weight: 600;
	}

	.status-badge.active {
		background: rgba(34, 197, 94, 0.2);
		color: #22c55e;
	}

	.status-badge.trialing {
		background: rgba(59, 130, 246, 0.2);
		color: #3b82f6;
	}

	.sub-info {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.info-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.info-item label {
		color: var(--text-secondary, rgba(255, 255, 255, 0.6));
		font-size: 0.875rem;
	}

	.info-item .value {
		font-weight: 500;
	}

	.sub-actions {
		display: flex;
		gap: 0.75rem;
		margin-top: 0.5rem;
	}

	.btn-primary,
	.btn-secondary {
		flex: 1;
		padding: 0.75rem 1.5rem;
		border-radius: 8px;
		font-weight: 500;
		cursor: pointer;
		transition: transform 0.2s, opacity 0.2s;
		border: none;
	}

	.btn-primary {
		background: var(--primary-color, #3b82f6);
		color: white;
	}

	.btn-secondary {
		background: rgba(255, 255, 255, 0.1);
		color: var(--text-primary, white);
		border: 1px solid rgba(255, 255, 255, 0.2);
	}

	.btn-primary:hover,
	.btn-secondary:hover {
		transform: translateY(-2px);
		opacity: 0.9;
	}

	.no-subscription p {
		color: var(--text-secondary, rgba(255, 255, 255, 0.6));
		margin-bottom: 1rem;
	}
</style>

