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
	import SubscriptionCancelModal from '$lib/components/SubscriptionCancelModal.svelte';

	let loading = true;
	let error: string | null = null;
	let subscriptions: UserSubscriptionsResponse | null = null;
	let selectedSubscriptionId: string | null = null;
	let showCancelModal = false;
	let isProcessing = false;
	let supportSettings: SupportSettings | null = null;

	$: hasMultipleActive = subscriptions?.has_multiple_active || false;
	$: activeSubscriptions = subscriptions?.active_subscriptions || [];
	$: canceledSubscriptions = subscriptions?.canceled_subscriptions || [];
	$: subscriptionsToCancel = activeSubscriptions.filter((sub) => sub.id !== selectedSubscriptionId);
	$: subscriptionToKeep = activeSubscriptions.find((sub) => sub.id === selectedSubscriptionId) || null;

	onMount(async () => {
		// Check auth
		if (!$auth.isAuthenticated) {
			goto('/auth/login');
			return;
		}

		await Promise.all([
			loadSubscriptions(),
			loadSupportSettings()
		]);
	});

	async function loadSubscriptions() {
		try {
			loading = true;
			error = null;
			subscriptions = await UserSubscriptionService.getSubscriptions();
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
			// Don't fail the whole page if support settings fail
		}
	}

	function handleSelectSubscription(id: string) {
		selectedSubscriptionId = id;
	}

	function handleCancelSingle(id: string) {
		selectedSubscriptionId = null; // No "keep" subscription for single cancel
		subscriptionsToCancel = activeSubscriptions.filter((sub) => sub.id === id);
		showCancelModal = true;
	}

	function handleReviewCancellations() {
		if (!selectedSubscriptionId) {
			alert('Please select a subscription to keep');
			return;
		}
		showCancelModal = true;
	}

	async function handleConfirmCancellation() {
		if (subscriptionsToCancel.length === 0) return;

		try {
			isProcessing = true;
			error = null;

			const subscriptionIds = subscriptionsToCancel.map((sub) => sub.id);

			await UserSubscriptionService.cancelMultipleSubscriptions({
				subscription_ids: subscriptionIds,
				keep_subscription_id: selectedSubscriptionId || undefined,
				reason: hasMultipleActive
					? 'User consolidated multiple subscriptions'
					: 'User canceled subscription'
			});

			// Success! Reload subscriptions
			showCancelModal = false;
			selectedSubscriptionId = null;
			await loadSubscriptions();

			// Show success message
			alert(
				`✅ Success! ${subscriptionsToCancel.length} subscription${subscriptionsToCancel.length > 1 ? 's' : ''} will be canceled at the end of ${subscriptionsToCancel.length > 1 ? 'their' : 'the'} billing period${subscriptionsToCancel.length > 1 ? 's' : ''}.`
			);
		} catch (err) {
			console.error('Failed to cancel subscriptions:', err);
			error = err instanceof Error ? err.message : 'Failed to cancel subscriptions';
			isProcessing = false;
		} finally {
			isProcessing = false;
		}
	}

	function handleCloseCancelModal() {
		if (!isProcessing) {
			showCancelModal = false;
		}
	}
</script>

<svelte:head>
	<title>My Subscriptions | BOME</title>
</svelte:head>

<div class="page-container">
	<div class="page-header">
		<h1>📋 My Subscriptions</h1>
		<p class="subtitle">Manage your active subscriptions and billing</p>
	</div>

	{#if loading}
		<div class="loading-state">
			<div class="spinner-large"></div>
			<p>Loading your subscriptions...</p>
		</div>
	{:else if error}
		<div class="error-state">
			<div class="error-icon">⚠️</div>
			<h2>Oops! Something went wrong</h2>
			<p>{error}</p>
			<button class="btn btn-primary" on:click={loadSubscriptions}>Try Again</button>
		</div>
	{:else if subscriptions}
		<!-- Video Access Status -->
		<div class="access-banner" class:has-access={subscriptions.video_access}>
			{#if subscriptions.video_access}
				<span class="access-icon">✅</span>
				<span class="access-text">You have access to premium video content</span>
			{:else}
				<span class="access-icon">🔒</span>
				<span class="access-text">You don't have access to premium video content</span>
			{/if}
		</div>

		<!-- Multiple Active Subscriptions Warning -->
		{#if hasMultipleActive}
			<div class="warning-banner">
				<div class="warning-header">
					<span class="warning-icon">⚠️</span>
					<h2>Multiple Active Subscriptions Detected</h2>
				</div>
				<p class="warning-main-text">
					You have <strong>{activeSubscriptions.length} active subscriptions</strong> and are being charged for each one. 
					To avoid duplicate charges, please <strong>contact support</strong> to let us know which subscription you want to keep.
				</p>
				{#if supportSettings && (supportSettings.email || supportSettings.phone || supportSettings.url)}
					<div class="support-contact">
						<p class="support-message">
							<strong>📧 Contact Support to Consolidate Your Subscriptions:</strong>
						</p>
						<div class="support-methods">
							{#if supportSettings.email}
								<a href="mailto:{supportSettings.email}?subject=Multiple Active Subscriptions - Need Help" class="support-link">
									<span class="support-icon">✉️</span>
									{supportSettings.email}
								</a>
							{/if}
							{#if supportSettings.phone}
								<a href="tel:{supportSettings.phone}" class="support-link">
									<span class="support-icon">📞</span>
									{supportSettings.phone}
								</a>
							{/if}
							{#if supportSettings.url}
								<a href={supportSettings.url} target="_blank" rel="noopener noreferrer" class="support-link">
									<span class="support-icon">🔗</span>
									Help Center
								</a>
							{/if}
						</div>
						{#if supportSettings.hours}
							<p class="support-hours">
								<span class="support-icon">🕐</span>
								<strong>Hours:</strong> {supportSettings.hours}
							</p>
						{/if}
						<p class="support-instructions">
							Please include your email address ({$auth.user?.email || 'your email'}) and specify which subscription you'd like to keep.
						</p>
					</div>
				{:else}
					<p class="warning-fallback">
						Please contact support to consolidate your subscriptions and avoid duplicate charges.
					</p>
				{/if}
			</div>
		{/if}

		<!-- Active Subscriptions -->
		{#if activeSubscriptions.length > 0}
			<section class="subscriptions-section">
				<h2 class="section-title">
					{#if hasMultipleActive}
						⚠️ Your Active Subscriptions ({activeSubscriptions.length})
					{:else}
						✅ Current Plan
					{/if}
				</h2>
				{#if hasMultipleActive}
					<p class="section-description">
						All of these subscriptions are currently active and charging you. Please contact support to choose one.
					</p>
				{/if}

				<div class="subscriptions-grid">
					{#each activeSubscriptions as subscription (subscription.id)}
						<SubscriptionCard
							{subscription}
							canSelect={false}
							isSelected={false}
							onCancel={!hasMultipleActive ? handleCancelSingle : undefined}
						/>
					{/each}
				</div>
			</section>
		{:else}
			<div class="empty-state">
				<div class="empty-icon">📭</div>
				<h2>No Active Subscription</h2>
				<p>You don't have an active subscription.</p>
				<button class="btn btn-primary" on:click={() => goto('/subscribe')}>
					🎬 Subscribe Now
				</button>
			</div>
		{/if}

		<!-- Canceled/Past Subscriptions -->
		{#if canceledSubscriptions.length > 0}
			<section class="subscriptions-section">
				<h2 class="section-title">📜 Subscription History</h2>

				<div class="subscriptions-grid">
					{#each canceledSubscriptions as subscription (subscription.id)}
						<SubscriptionCard {subscription} />
					{/each}
				</div>
			</section>
		{/if}

		<!-- No Subscriptions at All -->
		{#if activeSubscriptions.length === 0 && canceledSubscriptions.length === 0}
			<div class="empty-state">
				<div class="empty-icon">📋</div>
				<h2>No Subscriptions Found</h2>
				<p>You haven't subscribed to any plans yet.</p>
				<button class="btn btn-primary" on:click={() => goto('/subscribe')}>
					Browse Plans
				</button>
			</div>
		{/if}
	{/if}
</div>

<!-- Cancellation Modal -->
<SubscriptionCancelModal
	isOpen={showCancelModal}
	{subscriptionsToCancel}
	{subscriptionToKeep}
	{isProcessing}
	onConfirm={handleConfirmCancellation}
	onClose={handleCloseCancelModal}
/>

<style>
	.page-container {
		max-width: 1200px;
		margin: 0 auto;
		padding: 2rem;
	}

	.page-header {
		margin-bottom: 2rem;
	}

	.page-header h1 {
		font-size: 2.5rem;
		font-weight: 700;
		color: #111827;
		margin: 0 0 0.5rem 0;
	}

	.subtitle {
		font-size: 1.125rem;
		color: #6b7280;
		margin: 0;
	}

	.access-banner {
		padding: 1rem 1.5rem;
		border-radius: 12px;
		margin-bottom: 2rem;
		display: flex;
		align-items: center;
		gap: 1rem;
		font-weight: 600;
		border: 2px solid;
	}

	.access-banner.has-access {
		background: #d1fae5;
		border-color: #10b981;
		color: #065f46;
	}

	.access-banner:not(.has-access) {
		background: #fee2e2;
		border-color: #ef4444;
		color: #991b1b;
	}

	.access-icon {
		font-size: 1.5rem;
	}

	.access-text {
		font-size: 1rem;
	}

	.warning-banner {
		background: linear-gradient(135deg, #fef3c7, #fed7aa);
		border: 3px solid #fbbf24;
		border-radius: 16px;
		padding: 1.5rem;
		margin-bottom: 2rem;
	}

	.warning-header {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 1rem;
	}

	.warning-icon {
		font-size: 2rem;
	}

	.warning-header h2 {
		margin: 0;
		font-size: 1.5rem;
		color: #92400e;
	}

	.warning-banner p {
		margin: 0;
		font-size: 1rem;
		color: #78350f;
		line-height: 1.6;
	}

	.warning-main-text {
		margin-bottom: 1.5rem !important;
		font-size: 1.05rem !important;
	}

	.warning-fallback {
		margin-top: 1rem !important;
		padding: 1rem;
		background: rgba(255, 255, 255, 0.7);
		border-radius: 8px;
		font-style: italic;
	}

	.support-contact {
		margin-top: 1.5rem;
		padding: 1rem;
		background: rgba(255, 255, 255, 0.7);
		border-radius: 8px;
		border: 2px dashed #fbbf24;
	}

	.support-message {
		margin: 0 0 1rem 0;
		font-weight: 600;
		color: #92400e;
		font-size: 0.95rem;
	}

	.support-methods {
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
		margin-bottom: 0.75rem;
	}

	.support-link {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		background: white;
		border: 1px solid #fbbf24;
		border-radius: 6px;
		color: #92400e;
		text-decoration: none;
		font-weight: 600;
		font-size: 0.9rem;
		transition: all 0.2s;
	}

	.support-link:hover {
		background: #fef3c7;
		transform: translateY(-1px);
	}

	.support-icon {
		font-size: 1.1rem;
	}

	.support-hours {
		margin: 0.5rem 0 0 0;
		font-size: 0.875rem;
		color: #78350f;
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.support-instructions {
		margin: 1rem 0 0 0 !important;
		padding: 0.75rem;
		background: rgba(146, 64, 14, 0.1);
		border-left: 3px solid #92400e;
		border-radius: 4px;
		font-size: 0.9rem !important;
		color: #78350f !important;
		font-style: italic;
	}

	.subscriptions-section {
		margin-bottom: 3rem;
	}

	.section-title {
		font-size: 1.75rem;
		font-weight: 700;
		color: #111827;
		margin: 0 0 1rem 0;
	}

	.section-description {
		margin: 0 0 1.5rem 0;
		color: #6b7280;
		font-size: 1rem;
		line-height: 1.6;
	}

	.subscriptions-grid {
		display: grid;
		gap: 1.5rem;
	}

	.action-bar {
		margin-top: 2rem;
		display: flex;
		justify-content: center;
	}

	.loading-state,
	.error-state,
	.empty-state {
		text-align: center;
		padding: 4rem 2rem;
	}

	.spinner-large {
		width: 4rem;
		height: 4rem;
		border: 4px solid #e5e7eb;
		border-top-color: #3b82f6;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
		margin: 0 auto 1.5rem;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.error-icon,
	.empty-icon {
		font-size: 4rem;
		margin-bottom: 1rem;
	}

	.error-state h2,
	.empty-state h2 {
		font-size: 1.5rem;
		color: #111827;
		margin: 0 0 0.5rem 0;
	}

	.error-state p,
	.empty-state p {
		color: #6b7280;
		margin: 0 0 1.5rem 0;
	}

	.btn {
		padding: 0.75rem 1.5rem;
		border-radius: 8px;
		font-weight: 600;
		font-size: 1rem;
		border: none;
		cursor: pointer;
		transition: all 0.2s;
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
	}

	.btn-primary {
		background: linear-gradient(135deg, #3b82f6, #2563eb);
		color: white;
	}

	.btn-primary:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 16px rgba(59, 130, 246, 0.3);
	}

	.btn-large {
		padding: 1rem 2rem;
		font-size: 1.125rem;
	}

	@media (max-width: 768px) {
		.page-container {
			padding: 1rem;
		}

		.page-header h1 {
			font-size: 2rem;
		}

		.subtitle {
			font-size: 1rem;
		}

		.warning-header h2 {
			font-size: 1.25rem;
		}
	}
</style>

