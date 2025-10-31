<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import { UserSubscriptionService } from '$lib/services/user-subscription-service';
	import type {
		UserSubscriptionsResponse,
		UserSubscription
	} from '$lib/services/user-subscription-service';
	import { SupportSettingsService, type SupportSettings } from '$lib/services/support-settings-service';
	import SubscriptionCard from '$lib/components/SubscriptionCard.svelte';
	// SubscriptionCancelModal removed - users contact support for cancellations

	export let embedded = false; // Whether this is embedded in dashboard or standalone page

	let loading = true;
	let error: string | null = null;
	let subscriptions: UserSubscriptionsResponse | null = null;
	let selectedSubscriptionId: string | null = null;
	let supportSettings: SupportSettings | null = null;

	$: hasMultipleActive = subscriptions?.has_multiple_active || false;
	$: activeSubscriptions = subscriptions?.active_subscriptions || [];
	$: canceledSubscriptions = subscriptions?.canceled_subscriptions || [];

	onMount(async () => {
		// Check auth
		if (!$auth.isAuthenticated) {
			goto('/auth/login');
			return;
		}

		await Promise.all([loadSubscriptions(), loadSupportSettings()]);
	});

	async function loadSubscriptions() {
		try {
			loading = true;
			error = null;
			const response = await UserSubscriptionService.getSubscriptions();
			subscriptions = response;
		} catch (err) {
			console.error('Failed to load subscriptions:', err);
			error = err instanceof Error ? err.message : 'Failed to load subscriptions';
		} finally {
			loading = false;
		}
	}

	async function loadSupportSettings() {
		try {
			supportSettings = await SupportSettingsService.getSupportSettings();
		} catch (err) {
			console.error('Failed to load support settings:', err);
		}
	}

	// For multiple active subscriptions, track which one user wants to keep
	function handleSelectSubscription(subscription: UserSubscription) {
		if (hasMultipleActive) {
			selectedSubscriptionId = subscription.id;
		}
	}
</script>

<div class="subscription-management {embedded ? 'embedded' : ''}">
	{#if loading}
		<div class="loading">
			<div class="spinner"></div>
			<p>Loading your subscriptions...</p>
		</div>
	{:else if error}
		<div class="error-state">
			<svg class="error-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<circle cx="12" cy="12" r="10"/>
				<line x1="15" y1="9" x2="9" y2="15"/>
				<line x1="9" y1="9" x2="15" y2="15"/>
			</svg>
			<p>{error}</p>
			<button class="btn-primary" on:click={loadSubscriptions}>Try Again</button>
		</div>
	{:else if subscriptions}
		<!-- Multiple Active Subscriptions Warning -->
		{#if hasMultipleActive && supportSettings}
			<div class="warning-banner">
				<div class="warning-icon">⚠️</div>
				<div class="warning-content">
					<p class="warning-main-text">
						You have {activeSubscriptions.length} active subscriptions. Please contact support to consolidate your subscriptions.
					</p>

					{#if supportSettings.email || supportSettings.phone || supportSettings.url}
						<div class="support-contact">
							<p class="support-message">{supportSettings.message || 'Please contact our support team for assistance.'}</p>
							<div class="support-methods">
								{#if supportSettings.email}
									<a
										href="mailto:{supportSettings.email}?subject=Multiple%20Active%20Subscriptions&body=Hi%2C%20I%20have%20multiple%20active%20subscriptions%20and%20would%20like%20to%20keep%20only%20one.%0A%0AMy%20email%3A%20{encodeURIComponent($auth.user?.email || '')}"
										class="support-link"
									>
										<svg class="support-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/>
											<polyline points="22,6 12,13 2,6"/>
										</svg>
										Email: {supportSettings.email}
									</a>
								{/if}

								{#if supportSettings.phone}
									<a href="tel:{supportSettings.phone}" class="support-link">
										<svg class="support-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"/>
										</svg>
										Phone: {supportSettings.phone}
									</a>
								{/if}

								{#if supportSettings.url}
									<a href={supportSettings.url} target="_blank" rel="noopener noreferrer" class="support-link">
										<svg class="support-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<circle cx="12" cy="12" r="10"/>
											<path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/>
											<line x1="12" y1="17" x2="12.01" y2="17"/>
										</svg>
										Support Portal
									</a>
								{/if}
							</div>
							{#if supportSettings.hours}
								<p class="support-hours">Hours: {supportSettings.hours}</p>
							{/if}
						</div>
					{:else}
						<p class="warning-fallback">
							Please contact our support team to consolidate your subscriptions.
						</p>
					{/if}

					<div class="support-instructions">
						<strong>What to include in your message:</strong>
						<ul>
							<li>Your email address: {$auth.user?.email}</li>
							<li>Which subscription you'd like to keep</li>
						</ul>
					</div>
				</div>
			</div>
		{/if}

		<!-- Active Subscriptions -->
		{#if activeSubscriptions.length > 0}
			<div class="subscriptions-section">
				<h2 class="section-title">
					{#if hasMultipleActive}
						⚠️ Your Active Subscriptions ({activeSubscriptions.length})
					{:else}
						✅ Current Plan
					{/if}
				</h2>
				{#if hasMultipleActive}
					<p class="section-description">You have multiple active subscriptions. Please contact support to consolidate them.</p>
				{/if}
				<div class="subscriptions-grid">
					{#each activeSubscriptions as subscription}
						<SubscriptionCard {subscription} />
					{/each}
				</div>
			</div>
		{:else}
			<div class="no-subscription">
				<p>You don't have any active subscriptions.</p>
				<button class="btn-primary" on:click={() => goto('/subscription')}>View Plans</button>
			</div>
		{/if}

		<!-- Canceled/Expired Subscriptions -->
		{#if canceledSubscriptions.length > 0}
			<div class="subscriptions-section">
				<h2 class="section-title">Subscription History</h2>
				<div class="subscriptions-grid">
					{#each canceledSubscriptions as subscription}
						<SubscriptionCard {subscription} />
					{/each}
				</div>
			</div>
		{/if}
	{/if}
</div>

<!-- Note: Cancel modal removed - users should contact support for subscription changes -->

<style>
	.subscription-management {
		width: 100%;
		max-width: 1200px;
		margin: 0 auto;
		padding: 2rem 1rem;
	}

	.subscription-management.embedded {
		padding: 0;
	}

	.loading,
	.error-state,
	.no-subscription {
		text-align: center;
		padding: 3rem 1rem;
	}

	.spinner {
		border: 4px solid rgba(255, 255, 255, 0.1);
		border-left-color: var(--primary-color, #3b82f6);
		border-radius: 50%;
		width: 48px;
		height: 48px;
		animation: spin 1s linear infinite;
		margin: 0 auto 1rem;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	.error-icon {
		width: 64px;
		height: 64px;
		stroke: #ef4444;
		margin: 0 auto 1rem;
	}

	.warning-banner {
		background: rgba(251, 191, 36, 0.1);
		border: 2px solid #f59e0b;
		border-radius: 12px;
		padding: 1.5rem;
		margin-bottom: 2rem;
		display: flex;
		gap: 1rem;
	}

	.warning-icon {
		font-size: 2rem;
		flex-shrink: 0;
	}

	.warning-content {
		flex: 1;
	}

	.warning-main-text {
		font-weight: 600;
		color: #f59e0b;
		margin-bottom: 1rem;
	}

	.support-contact {
		margin-top: 1rem;
	}

	.support-methods {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		margin: 1rem 0;
	}

	.support-link {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1rem;
		background: rgba(255, 255, 255, 0.05);
		border-radius: 8px;
		text-decoration: none;
		color: inherit;
		transition: background 0.2s;
	}

	.support-link:hover {
		background: rgba(255, 255, 255, 0.1);
	}

	.support-icon {
		width: 20px;
		height: 20px;
	}

	.subscriptions-section {
		margin-bottom: 3rem;
	}

	.section-title {
		font-size: 1.5rem;
		font-weight: 600;
		margin-bottom: 0.5rem;
	}

	.section-description {
		color: var(--text-secondary, rgba(255, 255, 255, 0.6));
		margin-bottom: 1.5rem;
	}

	.subscriptions-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: 1.5rem;
	}

	.btn-primary {
		background: var(--primary-color, #3b82f6);
		color: white;
		border: none;
		padding: 0.75rem 2rem;
		border-radius: 8px;
		font-weight: 500;
		cursor: pointer;
		transition: opacity 0.2s;
	}

	.btn-primary:hover {
		opacity: 0.9;
	}

	@media (max-width: 768px) {
		.subscriptions-grid {
			grid-template-columns: 1fr;
		}
	}
</style>

