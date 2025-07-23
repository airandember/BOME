<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import { subscriptionService, subscriptionUtils, type SubscriptionPlan } from '$lib/subscription';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import Navigation from '$lib/components/Navigation.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import { StarIcon, CheckIcon, XMarkIcon, ExclamationTriangleIcon } from '@heroicons/svelte/24/outline';

	let plans: SubscriptionPlan[] = [];
	let loading = true;
	let error = '';
	let currentSubscription: any = null;
	let showPromotionBanner = false;
	let promotionEndDate: Date | null = null;

	onMount(async () => {
		try {
			// Load subscription plans
			const response = await subscriptionService.getPlans();
			plans = response.plans || [];
			
			// Check for promoted plans
			const promotedPlans = plans.filter(plan => plan.is_promoted);
			if (promotedPlans.length > 0) {
				showPromotionBanner = true;
				// Get the earliest promotion end date
				const endDates = promotedPlans
					.map(plan => plan.promotion_end_date ? new Date(plan.promotion_end_date) : null)
					.filter(date => date !== null);
				if (endDates.length > 0) {
					promotionEndDate = new Date(Math.min(...endDates.map(d => d!.getTime())));
				}
			}

			// Load current subscription if user is authenticated
			if ($auth.isAuthenticated) {
				try {
					const subResponse = await subscriptionService.getCurrentSubscription();
					currentSubscription = subResponse.subscription;
				} catch (err) {
					// User might not have a subscription, which is fine
					console.log('No current subscription found');
				}
			}
		} catch (err) {
			error = 'Failed to load subscription plans';
			console.error('Error loading plans:', err);
		} finally {
			loading = false;
		}
	});

	const handleSubscribe = async (planId: string) => {
		if (!$auth.isAuthenticated) {
			showToast('Please log in to subscribe', 'warning');
			goto('/login');
			return;
		}

		// Check if user already has an active subscription
		if (currentSubscription && currentSubscription.status === 'active') {
			showToast('You already have an active subscription', 'warning');
			goto('/subscription/manage');
			return;
		}

		try {
			const successUrl = `${window.location.origin}/subscription/success`;
			const cancelUrl = `${window.location.origin}/subscription`;
			
			const response = await subscriptionService.createCheckoutSession(planId, successUrl, cancelUrl);
			
			if (response.url) {
				window.location.href = response.url;
			} else {
				showToast('Failed to create checkout session', 'error');
			}
		} catch (err) {
			showToast('Failed to start subscription process', 'error');
			console.error('Error creating checkout session:', err);
		}
	};

	const getPopularBadge = (plan: SubscriptionPlan) => {
		if (plan.popular) {
			return `
				<div class="popular-badge">
					<span>Most Popular</span>
				</div>
			`;
		}
		return '';
	};

	const getPromotionBadge = (plan: SubscriptionPlan) => {
		if (plan.is_promoted) {
			return `
				<div class="promotion-badge">
					<svg class="star-icon" viewBox="0 0 24 24" fill="currentColor">
						<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
					</svg>
					<span>Limited Time Deal</span>
				</div>
			`;
		}
		return '';
	};

	const formatPromotionEndDate = (date: Date) => {
		const now = new Date();
		const diffTime = date.getTime() - now.getTime();
		const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
		
		if (diffDays <= 0) return 'Ends today';
		if (diffDays === 1) return 'Ends tomorrow';
		if (diffDays <= 7) return `Ends in ${diffDays} days`;
		return `Ends ${date.toLocaleDateString()}`;
	};

	const getSubscriptionStatusText = () => {
		if (!currentSubscription) return '';
		
		switch (currentSubscription.status) {
			case 'active':
				return 'You have an active subscription';
			case 'cancelled':
				return 'Your subscription has been cancelled';
			case 'past_due':
				return 'Your subscription payment is past due';
			case 'unpaid':
				return 'Your subscription payment failed';
			default:
				return 'Subscription status: ' + currentSubscription.status;
		}
	};

	const getSubscriptionStatusColor = () => {
		if (!currentSubscription) return 'text-gray-600';
		
		switch (currentSubscription.status) {
			case 'active':
				return 'text-green-600';
			case 'cancelled':
				return 'text-red-600';
			case 'past_due':
				return 'text-yellow-600';
			case 'unpaid':
				return 'text-red-600';
			default:
				return 'text-gray-600';
		}
	};
</script>

<svelte:head>
	<title>Subscription Plans - BOME</title>
	<meta name="description" content="Choose your subscription plan to access exclusive Book of Mormon evidence content" />
</svelte:head>

<Navigation />

<div class="subscription-page">
	<div class="container">
		<!-- Promotion Banner -->
		{#if showPromotionBanner}
			<div class="promotion-banner" in:fly={{ y: -20, duration: 300 }}>
				<div class="promotion-content">
					<StarIcon class="star-icon" />
					<div class="promotion-text">
						<h3>Limited Time Offers!</h3>
						<p>Special deals on premium plans - {promotionEndDate ? formatPromotionEndDate(promotionEndDate) : 'Limited time only'}</p>
					</div>
				</div>
			</div>
		{/if}

		<!-- Current Subscription Status -->
		{#if currentSubscription}
			<div class="subscription-status" in:fly={{ y: 20, duration: 300 }}>
				<div class="status-content">
					<div class="status-info">
						<h3>Current Subscription</h3>
						<p class={getSubscriptionStatusColor()}>{getSubscriptionStatusText()}</p>
						{#if currentSubscription.plan_name}
							<p class="plan-name">{currentSubscription.plan_name}</p>
						{/if}
					</div>
					<div class="status-actions">
						<button class="btn btn-secondary" on:click={() => goto('/subscription/manage')}>
							Manage Subscription
						</button>
					</div>
				</div>
			</div>
		{/if}

		<header class="page-header">
			<h1>Choose Your Plan</h1>
			<p>Unlock exclusive content and support our mission to share Book of Mormon evidence</p>
		</header>

		{#if loading}
			<div class="loading-container">
				<LoadingSpinner />
				<p>Loading subscription plans...</p>
			</div>
		{:else if error}
			<div class="error-container">
				<p class="error-message">{error}</p>
				<button class="btn btn-primary" on:click={() => window.location.reload()}>
					Try Again
				</button>
			</div>
		{:else}
			<div class="plans-grid">
				{#each plans as plan}
					<div class="plan-card" class:popular={plan.popular} class:promoted={plan.is_promoted}>
						{@html getPopularBadge(plan)}
						{@html getPromotionBadge(plan)}
						
						<div class="plan-header">
							<h3 class="plan-name">{plan.name}</h3>
							<div class="plan-price">
								<span class="price-amount">
									{subscriptionUtils.formatPrice(plan.price, plan.currency)}
								</span>
								<span class="price-interval">/{plan.interval}</span>
							</div>
							{#if plan.interval === 'year'}
								<div class="savings-badge">
									Save {Math.round((1 - plan.price / (subscriptionUtils.getMonthlyPrice(plan) * 12)) * 100)}%
								</div>
							{/if}
							{#if plan.is_promoted && plan.promotion_end_date}
								<div class="promotion-timer">
									{formatPromotionEndDate(new Date(plan.promotion_end_date))}
								</div>
							{/if}
						</div>

						<div class="plan-description">
							<p>{plan.description}</p>
						</div>

						<div class="plan-features">
							<ul>
								{#each plan.features as feature}
									<li>
										<CheckIcon class="check-icon" />
										{feature}
									</li>
								{/each}
							</ul>
						</div>

						<div class="plan-action">
							{#if currentSubscription && currentSubscription.status === 'active'}
								<button class="btn btn-disabled btn-full" disabled>
									Current Plan
								</button>
							{:else}
								<button 
									class="btn btn-primary btn-full" 
									on:click={() => handleSubscribe(plan.id)}
								>
									{plan.popular ? 'Get Started' : plan.is_promoted ? 'Get Deal' : 'Subscribe'}
								</button>
							{/if}
						</div>
					</div>
				{/each}
			</div>

			<div class="subscription-info">
				<div class="info-card">
					<h3>What's Included</h3>
					<ul>
						<li>Exclusive video content about Book of Mormon evidence</li>
						<li>Early access to new research and discoveries</li>
						<li>Ad-free viewing experience</li>
						<li>Download videos for offline viewing</li>
						<li>Priority customer support</li>
						<li>Access to our community forum</li>
					</ul>
				</div>

				<div class="info-card">
					<h3>Subscription Details</h3>
					<ul>
						<li>Cancel anytime - no long-term commitment</li>
						<li>Secure payment processing with Stripe</li>
						<li>Automatic renewal unless canceled</li>
						<li>Access to all content during subscription period</li>
						<li>30-day money-back guarantee</li>
					</ul>
				</div>
			</div>
		{/if}
	</div>
</div>

<Footer />

<style>
	.subscription-page {
		min-height: 100vh;
		padding: 2rem 0;
		background: var(--bg-gradient);
	}

	.container {
		max-width: 1200px;
		margin: 0 auto;
		padding: 0 1rem;
	}

	/* Promotion Banner */
	.promotion-banner {
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
		color: white;
		border-radius: 16px;
		padding: 1.5rem;
		margin-bottom: 2rem;
		box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
	}

	.promotion-content {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.promotion-content .star-icon {
		width: 24px;
		height: 24px;
		flex-shrink: 0;
	}

	.promotion-text h3 {
		font-size: 1.25rem;
		font-weight: 600;
		margin-bottom: 0.25rem;
	}

	.promotion-text p {
		margin: 0;
		opacity: 0.9;
	}

	/* Subscription Status */
	.subscription-status {
		background: var(--card-bg);
		border-radius: 16px;
		padding: 1.5rem;
		margin-bottom: 2rem;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
	}

	.status-content {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 1rem;
	}

	.status-info h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.status-info p {
		margin: 0.25rem 0;
	}

	.plan-name {
		font-weight: 600;
		color: var(--primary-color);
	}

	.status-actions {
		flex-shrink: 0;
	}

	.page-header {
		text-align: center;
		margin-bottom: 3rem;
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
		max-width: 600px;
		margin: 0 auto;
	}

	.loading-container,
	.error-container {
		text-align: center;
		padding: 3rem 0;
	}

	.error-message {
		color: var(--error-text);
		margin-bottom: 1rem;
	}

	.plans-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: 2rem;
		margin-bottom: 3rem;
	}

	.plan-card {
		position: relative;
		background: var(--card-bg);
		border-radius: 20px;
		padding: 2rem;
		box-shadow: var(--neumorphic-shadow);
		transition: all 0.3s ease;
		border: 1px solid var(--border-color);
	}

	.plan-card:hover {
		transform: translateY(-5px);
		box-shadow: var(--neumorphic-shadow-hover);
	}

	.plan-card.popular {
		border: 2px solid var(--primary-color);
		transform: scale(1.05);
	}

	.plan-card.promoted {
		border: 2px solid #667eea;
		background: linear-gradient(135deg, var(--card-bg) 0%, rgba(102, 126, 234, 0.05) 100%);
	}

	.popular-badge {
		position: absolute;
		top: -12px;
		left: 50%;
		transform: translateX(-50%);
		background: var(--primary-color);
		color: white;
		padding: 0.5rem 1rem;
		border-radius: 20px;
		font-size: 0.875rem;
		font-weight: 600;
	}

	.promotion-badge {
		position: absolute;
		top: -12px;
		right: 1rem;
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
		color: white;
		padding: 0.5rem 0.75rem;
		border-radius: 20px;
		font-size: 0.75rem;
		font-weight: 600;
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}

	.promotion-badge .star-icon {
		width: 14px;
		height: 14px;
	}

	.plan-header {
		text-align: center;
		margin-bottom: 1.5rem;
	}

	.plan-name {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.plan-price {
		margin-bottom: 0.5rem;
	}

	.price-amount {
		font-size: 2.5rem;
		font-weight: 700;
		color: var(--primary-color);
	}

	.price-interval {
		font-size: 1rem;
		color: var(--text-secondary);
	}

	.savings-badge {
		display: inline-block;
		background: var(--success-bg);
		color: var(--success-text);
		padding: 0.25rem 0.75rem;
		border-radius: 12px;
		font-size: 0.875rem;
		font-weight: 600;
		margin-bottom: 0.5rem;
	}

	.promotion-timer {
		display: inline-block;
		background: rgba(102, 126, 234, 0.1);
		color: #667eea;
		padding: 0.25rem 0.75rem;
		border-radius: 12px;
		font-size: 0.875rem;
		font-weight: 600;
	}

	.plan-description {
		text-align: center;
		margin-bottom: 1.5rem;
	}

	.plan-description p {
		color: var(--text-secondary);
		line-height: 1.6;
	}

	.plan-features {
		margin-bottom: 2rem;
	}

	.plan-features ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.plan-features li {
		display: flex;
		align-items: center;
		padding: 0.5rem 0;
		color: var(--text-primary);
	}

	.check-icon {
		width: 20px;
		height: 20px;
		color: var(--success-color);
		margin-right: 0.75rem;
		flex-shrink: 0;
	}

	.plan-action {
		text-align: center;
	}

	.btn-full {
		width: 100%;
	}

	.btn-disabled {
		background: var(--gray-300);
		color: var(--gray-600);
		cursor: not-allowed;
	}

	.subscription-info {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: 2rem;
		margin-top: 3rem;
	}

	.info-card {
		background: var(--card-bg);
		border-radius: 16px;
		padding: 1.5rem;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
	}

	.info-card h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.info-card ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.info-card li {
		padding: 0.5rem 0;
		color: var(--text-secondary);
		position: relative;
		padding-left: 1.5rem;
	}

	.info-card li::before {
		content: '•';
		position: absolute;
		left: 0;
		color: var(--primary-color);
		font-weight: bold;
	}

	@media (max-width: 768px) {
		.page-header h1 {
			font-size: 2rem;
		}

		.plans-grid {
			grid-template-columns: 1fr;
			gap: 1.5rem;
		}

		.plan-card {
			padding: 1.5rem;
		}

		.plan-card.popular {
			transform: none;
		}

		.price-amount {
			font-size: 2rem;
		}

		.subscription-info {
			grid-template-columns: 1fr;
		}

		.status-content {
			flex-direction: column;
			text-align: center;
		}

		.promotion-content {
			flex-direction: column;
			text-align: center;
		}
	}
</style> 