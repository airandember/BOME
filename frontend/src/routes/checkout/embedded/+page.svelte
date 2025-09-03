<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { auth, apiRequest } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let loading = true;
	let error = '';
	let checkoutContainer: HTMLElement;
	let stripe: any = null;
	let checkout: any = null;

	// Get plan ID from URL params
	$: planId = $page.url.searchParams.get('plan_id');
	$: returnUrl = `${window.location.origin}/checkout/success?session_id={CHECKOUT_SESSION_ID}`;

	onMount(async () => {
		if (!$auth.isAuthenticated) {
			showToast('Please log in to continue', 'warning');
			goto('/login');
			return;
		}

		if (!planId) {
			showToast('No plan selected', 'error');
			goto('/subscription');
			return;
		}

		try {
			await initializeEmbeddedCheckout();
		} catch (err) {
			console.error('Error initializing checkout:', err);
			error = 'Failed to initialize checkout';
			loading = false;
		}
	});

	async function initializeEmbeddedCheckout() {
		try {
			// Load Stripe.js
			if (!window.Stripe) {
				// Get publishable key from backend
				const configResponse = await apiRequest('/stripe/config');
				const { publishable_key } = await configResponse.json();
				
				if (!publishable_key) {
					throw new Error('Stripe publishable key not configured');
				}

				// Load Stripe.js script
				const script = document.createElement('script');
				script.src = 'https://js.stripe.com/v3/';
				script.onload = () => {
					stripe = window.Stripe(publishable_key);
					mountCheckout();
				};
				script.onerror = () => {
					throw new Error('Failed to load Stripe.js');
				};
				document.head.appendChild(script);
			} else {
				// Get publishable key and initialize Stripe
				const configResponse = await apiRequest('/stripe/config');
				const { publishable_key } = await configResponse.json();
				stripe = window.Stripe(publishable_key);
				await mountCheckout();
			}
		} catch (err) {
			console.error('Error initializing Stripe:', err);
			throw err;
		}
	}

	async function mountCheckout() {
		try {
			// Create fetchClientSecret function
			const fetchClientSecret = async () => {
				const response = await apiRequest('/stripe/checkout-session', {
					method: 'POST',
					body: JSON.stringify({
						plan_id: planId,
						return_url: returnUrl
					})
				});

				if (!response.ok) {
					const errorData = await response.json();
					throw new Error(errorData.error || 'Failed to create checkout session');
				}

				const { client_secret } = await response.json();
				return client_secret;
			};

			// Initialize embedded checkout
			checkout = await stripe.initEmbeddedCheckout({
				fetchClientSecret
			});

			// Mount checkout
			checkout.mount(checkoutContainer);
			loading = false;
		} catch (err) {
			console.error('Error mounting checkout:', err);
			error = err.message || 'Failed to load checkout';
			loading = false;
		}
	}

	function goBack() {
		goto('/subscription');
	}
</script>

<svelte:head>
	<title>Checkout - BOME</title>
	<meta name="description" content="Complete your subscription to BOME" />
</svelte:head>

<div class="checkout-page">
	<div class="container">
		<header class="page-header">
			<button class="back-button" on:click={goBack}>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M19 12H5"></path>
					<path d="M12 19l-7-7 7-7"></path>
				</svg>
				Back to Plans
			</button>
			<h1>Complete Your Subscription</h1>
			<p>Secure checkout powered by Stripe</p>
		</header>

		{#if loading}
			<div class="loading-container">
				<LoadingSpinner />
				<p>Loading secure checkout...</p>
			</div>
		{:else if error}
			<div class="error-container">
				<div class="error-icon">⚠️</div>
				<h2>Checkout Error</h2>
				<p class="error-message">{error}</p>
				<div class="error-actions">
					<button class="btn btn-primary" on:click={goBack}>
						Back to Plans
					</button>
					<button class="btn btn-outline" on:click={() => window.location.reload()}>
						Try Again
					</button>
				</div>
			</div>
		{:else}
			<div class="checkout-container">
				<div class="checkout-wrapper" bind:this={checkoutContainer}>
					<!-- Stripe Embedded Checkout will be mounted here -->
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	.checkout-page {
		min-height: 100vh;
		background: var(--bg-gradient);
		padding: 2rem 0;
	}

	.container {
		max-width: 800px;
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
		padding: 4rem 0;
	}

	.loading-container p {
		margin-top: 1rem;
		color: var(--text-secondary);
		font-size: 1.1rem;
	}

	.error-container {
		text-align: center;
		padding: 4rem 2rem;
		background: var(--card-bg);
		border-radius: 20px;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
	}

	.error-icon {
		font-size: 4rem;
		margin-bottom: 1rem;
	}

	.error-container h2 {
		font-size: 1.5rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.error-message {
		color: var(--error-text);
		margin-bottom: 2rem;
		font-size: 1.1rem;
	}

	.error-actions {
		display: flex;
		justify-content: center;
		gap: 1rem;
		flex-wrap: wrap;
	}

	.checkout-container {
		background: var(--card-bg);
		border-radius: 20px;
		padding: 2rem;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
	}

	.checkout-wrapper {
		min-height: 400px;
		/* Stripe will inject the checkout form here */
	}

	.btn {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
	}

	.btn-primary {
		background: var(--primary-color);
		color: white;
	}

	.btn-primary:hover {
		background: var(--primary-color-hover);
		transform: translateY(-1px);
	}

	.btn-outline {
		background: white;
		color: var(--text-primary);
		border: 1px solid var(--border-color);
	}

	.btn-outline:hover {
		background: var(--bg-secondary);
		border-color: var(--primary-color);
	}

	@media (max-width: 768px) {
		.page-header h1 {
			font-size: 2rem;
		}

		.back-button {
			position: static;
			justify-content: center;
			margin-bottom: 1rem;
		}

		.checkout-container {
			padding: 1rem;
		}

		.error-actions {
			flex-direction: column;
			align-items: center;
		}
	}
</style>
