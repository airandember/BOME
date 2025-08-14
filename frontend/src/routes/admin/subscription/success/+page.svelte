<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import { subscriptionService, subscriptionUtils, type Subscription } from '$lib/subscription';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let subscription: Subscription | null = null;
	let loading = true;
	let error = '';
	let isAuthenticated = false;

	// Subscribe to auth store
	auth.subscribe(state => {
		isAuthenticated = state.isAuthenticated;
	});

	onMount(async () => {
		// Check if user is authenticated
		if (!isAuthenticated) {
			goto('/login');
			return;
		}

		try {
			const response = await subscriptionService.getCurrentSubscription();
			subscription = response.subscription;
		} catch (err) {
			error = 'Failed to load subscription details';
			console.error('Error loading subscription:', err);
		} finally {
			loading = false;
		}
	});

	const handleContinue = () => {
		goto('/videos');
	};

	const handleManageSubscription = async () => {
		try {
			const returnUrl = `${window.location.origin}/account`;
			const response = await subscriptionService.createCustomerPortalSession(returnUrl);
			
			if (response.url) {
				window.location.href = response.url;
			} else {
				alert('Failed to open customer portal');
			}
		} catch (err) {
			alert('Failed to open customer portal');
			console.error('Error opening customer portal:', err);
		}
	};
</script>

<svelte:head>
	<title>Subscription Successful - BOME</title>
	<meta name="description" content="Your subscription has been successfully activated" />
</svelte:head>

<div class="success-page">
	<div class="container">
		{#if loading}
			<div class="loading-container">
				<LoadingSpinner />
				<p>Loading your subscription details...</p>
			</div>
		{:else if error}
			<div class="error-container">
				<div class="error-card">
					<div class="error-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<circle cx="12" cy="12" r="10"></circle>
							<line x1="15" y1="9" x2="9" y2="15"></line>
							<line x1="9" y1="9" x2="15" y2="15"></line>
						</svg>
					</div>
					<h2>Something went wrong</h2>
					<p class="error-message">{error}</p>
					<button class="btn btn-primary" on:click={() => window.location.reload()}>
						Try Again
					</button>
				</div>
			</div>
		{:else}
			<div class="success-card">
				<div class="success-icon">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
						<polyline points="22,4 12,14.01 9,11.01"></polyline>
					</svg>
				</div>

				<h1>Welcome to BOME!</h1>
				<p class="success-message">
					Your subscription has been successfully activated. You now have access to all our exclusive content.
				</p>

				{#if subscription}
					<div class="subscription-details">
						<div class="detail-item">
							<span class="label">Status:</span>
							<span class="value" style="color: {subscriptionUtils.getStatusColor(subscription.status)}">
								{subscriptionUtils.getStatusText(subscription.status)}
							</span>
						</div>
						<div class="detail-item">
							<span class="label">Next billing:</span>
							<span class="value">{subscriptionUtils.formatDate(subscription.currentPeriodEnd)}</span>
						</div>
					</div>
				{/if}

				<div class="action-buttons">
					<button class="btn btn-primary btn-large" on:click={handleContinue}>
						Start Watching
					</button>
					<button class="btn btn-secondary btn-large" on:click={handleManageSubscription}>
						Manage Subscription
					</button>
				</div>

				<div class="next-steps">
					<h3>What's Next?</h3>
					<div class="steps-grid">
						<div class="step-card">
							<div class="step-number">1</div>
							<h4>Explore Content</h4>
							<p>Browse our extensive library of Book of Mormon evidence videos</p>
						</div>
						<div class="step-card">
							<div class="step-number">2</div>
							<h4>Create Playlists</h4>
							<p>Save your favorite videos and create custom playlists</p>
						</div>
						<div class="step-card">
							<div class="step-number">3</div>
							<h4>Join Community</h4>
							<p>Connect with other members in our community forum</p>
						</div>
					</div>
				</div>

				<div class="support-info">
					<h3>Need Help?</h3>
					<p>
						Our support team is here to help you get the most out of your subscription. 
						Don't hesitate to reach out if you have any questions.
					</p>
					<div class="support-links">
						<a href="/contact" class="btn btn-outline">Contact Support</a>
						<a href="/faq" class="btn btn-outline">FAQ</a>
					</div>
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	.success-page {
		min-height: 100vh;
		padding: 2rem 0;
		background: var(--bg-gradient);
	}

	.container {
		max-width: 800px;
		margin: 0 auto;
		padding: 0 1rem;
	}

	.loading-container,
	.error-container {
		text-align: center;
		padding: 3rem 0;
	}

	.success-card {
		background: var(--card-bg);
		border-radius: 24px;
		padding: 3rem 2rem;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
		text-align: center;
	}

	.success-icon {
		width: 80px;
		height: 80px;
		margin: 0 auto 2rem;
		background: var(--success-bg);
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.success-icon svg {
		width: 40px;
		height: 40px;
		color: var(--success-color);
	}

	.error-card {
		background: var(--card-bg);
		border-radius: 24px;
		padding: 3rem 2rem;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
		text-align: center;
	}

	.error-icon {
		width: 80px;
		height: 80px;
		margin: 0 auto 2rem;
		background: var(--error-bg);
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.error-icon svg {
		width: 40px;
		height: 40px;
		color: var(--error-color);
	}

	h1 {
		font-size: 2.5rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	h2 {
		font-size: 2rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.success-message {
		font-size: 1.1rem;
		color: var(--text-secondary);
		margin-bottom: 2rem;
		line-height: 1.6;
	}

	.error-message {
		color: var(--error-text);
		margin-bottom: 1.5rem;
	}

	.subscription-details {
		background: var(--bg-secondary);
		border-radius: 16px;
		padding: 1.5rem;
		margin-bottom: 2rem;
		text-align: left;
	}

	.detail-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.5rem 0;
	}

	.detail-item:not(:last-child) {
		border-bottom: 1px solid var(--border-color);
	}

	.label {
		font-weight: 600;
		color: var(--text-primary);
	}

	.value {
		color: var(--text-secondary);
	}

	.action-buttons {
		display: flex;
		gap: 1rem;
		justify-content: center;
		margin-bottom: 3rem;
		flex-wrap: wrap;
	}

	.btn-large {
		padding: 1rem 2rem;
		font-size: 1.1rem;
	}

	.next-steps {
		margin-bottom: 3rem;
	}

	.next-steps h3 {
		font-size: 1.5rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1.5rem;
	}

	.steps-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1.5rem;
	}

	.step-card {
		background: var(--bg-secondary);
		border-radius: 16px;
		padding: 1.5rem;
		text-align: center;
		border: 1px solid var(--border-color);
	}

	.step-number {
		width: 40px;
		height: 40px;
		background: var(--primary-color);
		color: white;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 700;
		margin: 0 auto 1rem;
	}

	.step-card h4 {
		font-size: 1.1rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.step-card p {
		color: var(--text-secondary);
		font-size: 0.9rem;
		line-height: 1.5;
	}

	.support-info {
		border-top: 1px solid var(--border-color);
		padding-top: 2rem;
	}

	.support-info h3 {
		font-size: 1.5rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.support-info p {
		color: var(--text-secondary);
		margin-bottom: 1.5rem;
		line-height: 1.6;
	}

	.support-links {
		display: flex;
		gap: 1rem;
		justify-content: center;
		flex-wrap: wrap;
	}

	.btn-outline {
		background: transparent;
		border: 2px solid var(--primary-color);
		color: var(--primary-color);
	}

	.btn-outline:hover {
		background: var(--primary-color);
		color: white;
	}

	@media (max-width: 768px) {
		.success-card {
			padding: 2rem 1.5rem;
		}

		h1 {
			font-size: 2rem;
		}

		.action-buttons {
			flex-direction: column;
		}

		.btn-large {
			width: 100%;
		}

		.steps-grid {
			grid-template-columns: 1fr;
		}

		.support-links {
			flex-direction: column;
		}

		.btn-outline {
			width: 100%;
		}
	}
</style> 