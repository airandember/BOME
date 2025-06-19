<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import { subscriptionService, subscriptionUtils, type Subscription, type SubscriptionPlan } from '$lib/subscription';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let user: any = null;
	let subscription: Subscription | null = null;
	let availablePlans: SubscriptionPlan[] = [];
	let loading = true;
	let isAuthenticated = false;
	let showCancelModal = false;
	let showUpgradeModal = false;
	let selectedPlan: SubscriptionPlan | null = null;
	let cancelReason = '';

	// Subscribe to auth store
	auth.subscribe(state => {
		isAuthenticated = state.isAuthenticated;
		user = state.user;
	});

	onMount(async () => {
		if (!isAuthenticated) {
			goto('/login');
			return;
		}

		await loadSubscriptionData();
	});

	const loadSubscriptionData = async () => {
		try {
			loading = true;
			
			// Load current subscription and available plans
			const [subscriptionResponse, plansResponse] = await Promise.allSettled([
				subscriptionService.getCurrentSubscription(),
				subscriptionService.getPlans()
			]);

			if (subscriptionResponse.status === 'fulfilled') {
				subscription = subscriptionResponse.value.subscription;
			}

			if (plansResponse.status === 'fulfilled') {
				availablePlans = plansResponse.value.plans || [];
			}
		} catch (err) {
			showToast('Failed to load subscription data', 'error');
			console.error('Error loading subscription data:', err);
		} finally {
			loading = false;
		}
	};

	const handleManageSubscription = async () => {
		try {
			const returnUrl = `${window.location.origin}/account/subscription`;
			const response = await subscriptionService.createCustomerPortalSession(returnUrl);
			
			if (response.url) {
				window.location.href = response.url;
			} else {
				showToast('Failed to open customer portal', 'error');
			}
		} catch (err) {
			showToast('Failed to open customer portal', 'error');
			console.error('Error opening customer portal:', err);
		}
	};

	const handleUpgrade = () => {
		goto('/subscription');
	};

	const handleChangePlan = (plan: SubscriptionPlan) => {
		selectedPlan = plan;
		showUpgradeModal = true;
	};

	const confirmPlanChange = async () => {
		if (!selectedPlan || !subscription) return;

		try {
			const response = await subscriptionService.updateSubscription(subscription.id, selectedPlan.id);
			
			if (response.success) {
				showToast('Subscription plan updated successfully', 'success');
				await loadSubscriptionData();
			} else {
				showToast('Failed to update subscription plan', 'error');
			}
		} catch (err) {
			showToast('Failed to update subscription plan', 'error');
			console.error('Error updating subscription:', err);
		} finally {
			showUpgradeModal = false;
			selectedPlan = null;
		}
	};

	const handleCancelSubscription = () => {
		showCancelModal = true;
	};

	const confirmCancelSubscription = async () => {
		if (!subscription) return;

		try {
			const response = await subscriptionService.cancelSubscription(subscription.id, true);
			
			if (response.success) {
				showToast('Subscription cancelled successfully', 'success');
				await loadSubscriptionData();
			} else {
				showToast('Failed to cancel subscription', 'error');
			}
		} catch (err) {
			showToast('Failed to cancel subscription', 'error');
			console.error('Error cancelling subscription:', err);
		} finally {
			showCancelModal = false;
			cancelReason = '';
		}
	};

	const handleReactivateSubscription = async () => {
		if (!subscription) return;

		try {
			const response = await subscriptionService.reactivateSubscription(subscription.id);
			
			if (response.success) {
				showToast('Subscription reactivated successfully', 'success');
				await loadSubscriptionData();
			} else {
				showToast('Failed to reactivate subscription', 'error');
			}
		} catch (err) {
			showToast('Failed to reactivate subscription', 'error');
			console.error('Error reactivating subscription:', err);
		}
	};

	const getCurrentPlan = () => {
		if (!subscription) return null;
		return availablePlans.find(plan => plan.id === subscription?.planId);
	};

	const getOtherPlans = () => {
		const currentPlan = getCurrentPlan();
		return availablePlans.filter(plan => plan.id !== currentPlan?.id);
	};
</script>

<svelte:head>
	<title>Manage Subscription - BOME</title>
	<meta name="description" content="Manage your BOME subscription plan and billing" />
</svelte:head>

<div class="subscription-page">
	<div class="container">
		<header class="page-header">
			<button class="back-button" on:click={() => goto('/account')}>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M19 12H5"></path>
					<path d="M12 19l-7-7 7-7"></path>
				</svg>
				Back to Account
			</button>
			<h1>Manage Subscription</h1>
			<p>View and manage your subscription plan and billing</p>
		</header>

		{#if loading}
			<div class="loading-container">
				<LoadingSpinner />
				<p>Loading your subscription...</p>
			</div>
		{:else}
			<div class="subscription-content">
				{#if subscription}
					<!-- Current Subscription -->
					<div class="subscription-section">
						<h2>Current Subscription</h2>
						<div class="current-subscription-card">
							<div class="subscription-header">
								<div class="plan-info">
									<h3>{getCurrentPlan()?.name || 'Unknown Plan'}</h3>
									<div class="plan-price">
										{#if getCurrentPlan()}
											{subscriptionUtils.formatPrice(getCurrentPlan()?.price || 0, getCurrentPlan()?.currency || 'USD')}
											<span class="interval">/{getCurrentPlan()?.interval}</span>
										{/if}
									</div>
								</div>
								<div class="subscription-status">
									<span 
										class="status-badge" 
										style="background-color: {subscriptionUtils.getStatusColor(subscription.status)}"
									>
										{subscriptionUtils.getStatusText(subscription.status)}
									</span>
								</div>
							</div>

							<div class="subscription-details">
								<div class="detail-grid">
									<div class="detail-item">
										<span class="label">Status</span>
										<span class="value">{subscriptionUtils.getStatusText(subscription.status)}</span>
									</div>
									<div class="detail-item">
										<span class="label">Next billing</span>
										<span class="value">{subscriptionUtils.formatDate(subscription.currentPeriodEnd)}</span>
									</div>
									<div class="detail-item">
										<span class="label">Started</span>
										<span class="value">{subscriptionUtils.formatDate(subscription.createdAt)}</span>
									</div>
									{#if subscription.cancelAtPeriodEnd}
										<div class="detail-item">
											<span class="label">Cancellation</span>
											<span class="value warning">Ends at period end</span>
										</div>
									{/if}
								</div>

								{#if getCurrentPlan()}
									<div class="plan-features">
										<h4>Plan Features</h4>
										<ul>
											{#each getCurrentPlan()?.features || [] as feature}
												<li>
													<svg class="check-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
														<polyline points="20,6 9,17 4,12"></polyline>
													</svg>
													{feature}
												</li>
											{/each}
										</ul>
									</div>
								{/if}
							</div>

							<div class="subscription-actions">
								{#if subscription.status === 'active'}
									{#if !subscription.cancelAtPeriodEnd}
										<button class="btn btn-outline" on:click={handleManageSubscription}>
											Manage Billing
										</button>
										<button class="btn btn-danger-outline" on:click={handleCancelSubscription}>
											Cancel Subscription
										</button>
									{:else}
										<button class="btn btn-outline" on:click={handleManageSubscription}>
											Manage Billing
										</button>
										<button class="btn btn-primary" on:click={handleReactivateSubscription}>
											Reactivate Subscription
										</button>
									{/if}
								{:else}
									<button class="btn btn-primary" on:click={handleUpgrade}>
										Resubscribe
									</button>
								{/if}
							</div>
						</div>
					</div>

					<!-- Plan Options -->
					{#if getOtherPlans().length > 0}
						<div class="subscription-section">
							<h2>Change Plan</h2>
							<p class="section-description">Upgrade or downgrade your subscription plan</p>
							
							<div class="plans-grid">
								{#each getOtherPlans() as plan}
									<div class="plan-option-card">
										<div class="plan-header">
											<h3>{plan.name}</h3>
											<div class="plan-price">
												{subscriptionUtils.formatPrice(plan.price, plan.currency)}
												<span class="interval">/{plan.interval}</span>
											</div>
										</div>
										
										<div class="plan-description">
											<p>{plan.description}</p>
										</div>

										<div class="plan-features">
											<ul>
												{#each plan.features.slice(0, 3) as feature}
													<li>
														<svg class="check-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
															<polyline points="20,6 9,17 4,12"></polyline>
														</svg>
														{feature}
													</li>
												{/each}
											</ul>
										</div>

										<button 
											class="btn btn-primary btn-full" 
											on:click={() => handleChangePlan(plan)}
											disabled={subscription?.status !== 'active'}
										>
											{getCurrentPlan() && plan.price > (getCurrentPlan()?.price || 0) ? 'Upgrade' : 'Downgrade'} to {plan.name}
										</button>
									</div>
								{/each}
							</div>
						</div>
					{/if}
				{:else}
					<!-- No Subscription -->
					<div class="no-subscription">
						<div class="no-subscription-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<circle cx="12" cy="12" r="10"></circle>
								<path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"></path>
								<line x1="12" y1="17" x2="12.01" y2="17"></line>
							</svg>
						</div>
						<h2>No Active Subscription</h2>
						<p>You don't have an active subscription. Subscribe now to access exclusive content and features.</p>
						<button class="btn btn-primary" on:click={handleUpgrade}>
							View Subscription Plans
						</button>
					</div>
				{/if}

				<!-- Quick Links -->
				<div class="subscription-section">
					<h2>Quick Links</h2>
					<div class="quick-links">
						<button class="link-card" on:click={() => goto('/account/billing')}>
							<div class="link-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect>
									<line x1="8" y1="21" x2="16" y2="21"></line>
									<line x1="12" y1="17" x2="12" y2="21"></line>
								</svg>
							</div>
							<div class="link-content">
								<h3>Billing History</h3>
								<p>View invoices and payment history</p>
							</div>
						</button>

						<button class="link-card" on:click={() => goto('/subscription')}>
							<div class="link-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<circle cx="12" cy="12" r="3"></circle>
									<path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1 1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
								</svg>
							</div>
							<div class="link-content">
								<h3>All Plans</h3>
								<p>Compare all available plans</p>
							</div>
						</button>
					</div>
				</div>
			</div>
		{/if}
	</div>
</div>

<!-- Cancel Subscription Modal -->
{#if showCancelModal}
	<div class="modal-overlay" on:click={() => showCancelModal = false}>
		<div class="modal" on:click|stopPropagation>
			<div class="modal-header">
				<h3>Cancel Subscription</h3>
				<button class="modal-close" on:click={() => showCancelModal = false}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"></line>
						<line x1="6" y1="6" x2="18" y2="18"></line>
					</svg>
				</button>
			</div>
			
			<div class="modal-content">
				<p>Are you sure you want to cancel your subscription? You'll continue to have access until the end of your current billing period.</p>
				
				<div class="form-group">
					<label for="cancel-reason">Reason for cancellation (optional)</label>
					<textarea
						id="cancel-reason"
						bind:value={cancelReason}
						placeholder="Help us improve by telling us why you're cancelling..."
						rows="3"
					></textarea>
				</div>
			</div>
			
			<div class="modal-actions">
				<button class="btn btn-outline" on:click={() => showCancelModal = false}>
					Keep Subscription
				</button>
				<button class="btn btn-danger" on:click={confirmCancelSubscription}>
					Cancel Subscription
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Plan Change Modal -->
{#if showUpgradeModal && selectedPlan}
	<div class="modal-overlay" on:click={() => showUpgradeModal = false}>
		<div class="modal" on:click|stopPropagation>
			<div class="modal-header">
				<h3>Change Plan</h3>
				<button class="modal-close" on:click={() => showUpgradeModal = false}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"></line>
						<line x1="6" y1="6" x2="18" y2="18"></line>
					</svg>
				</button>
			</div>
			
			<div class="modal-content">
				<p>
					Are you sure you want to change to the <strong>{selectedPlan.name}</strong> plan?
				</p>
				<div class="plan-comparison">
					<div class="current-plan">
						<h4>Current: {getCurrentPlan()?.name || 'Unknown'}</h4>
						<p>{getCurrentPlan() ? subscriptionUtils.formatPrice(getCurrentPlan()?.price || 0, getCurrentPlan()?.currency || 'USD') : 'N/A'}/{getCurrentPlan()?.interval || ''}</p>
					</div>
					<div class="new-plan">
						<h4>New: {selectedPlan.name}</h4>
						<p>{subscriptionUtils.formatPrice(selectedPlan.price, selectedPlan.currency)}/{selectedPlan.interval}</p>
					</div>
				</div>
				<p class="change-note">The change will take effect immediately and you'll be prorated for the difference.</p>
			</div>
			
			<div class="modal-actions">
				<button class="btn btn-outline" on:click={() => showUpgradeModal = false}>
					Cancel
				</button>
				<button class="btn btn-primary" on:click={confirmPlanChange}>
					Change Plan
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.subscription-page {
		min-height: 100vh;
		padding: 2rem 0;
		background: var(--bg-gradient);
	}

	.container {
		max-width: 1000px;
		margin: 0 auto;
		padding: 0 1rem;
	}

	.page-header {
		text-align: center;
		margin-bottom: 3rem;
		position: relative;
	}

	.back-button {
		position: absolute;
		left: 0;
		top: 0;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		background: none;
		border: none;
		color: var(--text-secondary);
		cursor: pointer;
		font-size: 0.875rem;
		transition: color 0.3s ease;
	}

	.back-button:hover {
		color: var(--primary-color);
	}

	.back-button svg {
		width: 16px;
		height: 16px;
	}

	.page-header h1 {
		font-size: 2.5rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.page-header p {
		font-size: 1.1rem;
		color: var(--text-secondary);
	}

	.loading-container {
		text-align: center;
		padding: 3rem 0;
	}

	.subscription-content {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.subscription-section {
		background: var(--card-bg);
		border-radius: 20px;
		padding: 2rem;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
	}

	.subscription-section h2 {
		font-size: 1.5rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.section-description {
		color: var(--text-secondary);
		margin-bottom: 2rem;
	}

	.current-subscription-card {
		background: var(--bg-secondary);
		border-radius: 16px;
		padding: 2rem;
		border: 1px solid var(--border-color);
	}

	.subscription-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 2rem;
	}

	.plan-info h3 {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.plan-price {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--primary-color);
	}

	.interval {
		font-size: 1rem;
		color: var(--text-secondary);
	}

	.status-badge {
		display: inline-block;
		padding: 0.5rem 1rem;
		border-radius: 20px;
		font-size: 0.875rem;
		font-weight: 600;
		color: white;
	}

	.detail-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
		margin-bottom: 2rem;
	}

	.detail-item {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.detail-item .label {
		font-size: 0.875rem;
		color: var(--text-secondary);
	}

	.detail-item .value {
		font-weight: 600;
		color: var(--text-primary);
	}

	.detail-item .value.warning {
		color: var(--warning-text);
	}

	.plan-features {
		margin-bottom: 2rem;
	}

	.plan-features h4 {
		font-size: 1rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.plan-features ul {
		list-style: none;
		padding: 0;
		margin: 0;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 0.5rem;
	}

	.plan-features li {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		color: var(--text-primary);
		font-size: 0.875rem;
	}

	.check-icon {
		width: 16px;
		height: 16px;
		color: var(--success-color);
		flex-shrink: 0;
	}

	.subscription-actions {
		display: flex;
		gap: 1rem;
		flex-wrap: wrap;
	}

	.btn-danger-outline {
		background: transparent;
		color: var(--error-text);
		border: 1px solid var(--error-border);
	}

	.btn-danger-outline:hover {
		background: var(--error-bg);
	}

	.no-subscription {
		text-align: center;
		padding: 3rem 2rem;
		background: var(--card-bg);
		border-radius: 20px;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
	}

	.no-subscription-icon {
		width: 80px;
		height: 80px;
		margin: 0 auto 2rem;
		color: var(--text-secondary);
	}

	.no-subscription-icon svg {
		width: 100%;
		height: 100%;
	}

	.no-subscription h2 {
		font-size: 1.5rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.no-subscription p {
		color: var(--text-secondary);
		margin-bottom: 2rem;
		max-width: 500px;
		margin-left: auto;
		margin-right: auto;
	}

	.plans-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: 1.5rem;
	}

	.plan-option-card {
		background: var(--bg-secondary);
		border-radius: 16px;
		padding: 1.5rem;
		border: 1px solid var(--border-color);
		transition: all 0.3s ease;
	}

	.plan-option-card:hover {
		transform: translateY(-2px);
		box-shadow: var(--neumorphic-shadow-hover);
	}

	.plan-header {
		text-align: center;
		margin-bottom: 1rem;
	}

	.plan-option-card .plan-header h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.plan-description {
		text-align: center;
		margin-bottom: 1.5rem;
	}

	.plan-description p {
		color: var(--text-secondary);
		font-size: 0.875rem;
	}

	.btn-full {
		width: 100%;
	}

	.quick-links {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 1rem;
	}

	.link-card {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1.5rem;
		background: var(--bg-secondary);
		border: 1px solid var(--border-color);
		border-radius: 16px;
		transition: all 0.3s ease;
		cursor: pointer;
		text-align: left;
	}

	.link-card:hover {
		transform: translateY(-2px);
		box-shadow: var(--neumorphic-shadow-hover);
	}

	.link-icon {
		width: 40px;
		height: 40px;
		color: var(--primary-color);
		flex-shrink: 0;
	}

	.link-icon svg {
		width: 100%;
		height: 100%;
	}

	.link-content h3 {
		font-size: 1rem;
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 0.25rem 0;
	}

	.link-content p {
		font-size: 0.875rem;
		color: var(--text-secondary);
		margin: 0;
	}

	/* Modal Styles */
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}

	.modal {
		background: var(--card-bg);
		border-radius: 20px;
		padding: 0;
		max-width: 500px;
		width: 90%;
		max-height: 90vh;
		overflow-y: auto;
		box-shadow: var(--neumorphic-shadow);
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 2rem 2rem 0 2rem;
		border-bottom: 1px solid var(--border-color);
		margin-bottom: 2rem;
	}

	.modal-header h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.modal-close {
		background: none;
		border: none;
		color: var(--text-secondary);
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 8px;
		transition: all 0.3s ease;
	}

	.modal-close:hover {
		background: var(--bg-secondary);
		color: var(--text-primary);
	}

	.modal-close svg {
		width: 20px;
		height: 20px;
	}

	.modal-content {
		padding: 0 2rem;
	}

	.modal-content p {
		color: var(--text-secondary);
		line-height: 1.6;
		margin-bottom: 1.5rem;
	}

	.form-group {
		margin-bottom: 1.5rem;
	}

	.form-group label {
		display: block;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.form-group textarea {
		width: 100%;
		padding: 0.75rem 1rem;
		border: 1px solid var(--border-color);
		border-radius: 12px;
		background: var(--bg-secondary);
		color: var(--text-primary);
		font-size: 1rem;
		resize: vertical;
		min-height: 80px;
	}

	.form-group textarea:focus {
		outline: none;
		border-color: var(--primary-color);
	}

	.plan-comparison {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
		margin: 1.5rem 0;
		padding: 1rem;
		background: var(--bg-secondary);
		border-radius: 12px;
	}

	.current-plan,
	.new-plan {
		text-align: center;
	}

	.current-plan h4,
	.new-plan h4 {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.current-plan p,
	.new-plan p {
		font-size: 1rem;
		font-weight: 600;
		color: var(--primary-color);
		margin: 0;
	}

	.change-note {
		font-size: 0.875rem;
		color: var(--text-tertiary);
		font-style: italic;
	}

	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 1rem;
		padding: 2rem;
		border-top: 1px solid var(--border-color);
		margin-top: 2rem;
	}

	@media (max-width: 768px) {
		.back-button {
			position: static;
			justify-content: center;
			margin-bottom: 1rem;
		}

		.subscription-header {
			flex-direction: column;
			gap: 1rem;
		}

		.detail-grid {
			grid-template-columns: 1fr;
		}

		.subscription-actions {
			flex-direction: column;
		}

		.plan-comparison {
			grid-template-columns: 1fr;
		}

		.modal-actions {
			flex-direction: column;
		}
	}
</style> 