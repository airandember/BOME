<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { auth, apiRequest } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import Navigation from '$lib/components/Navigation.svelte';
	import Footer from '$lib/components/Footer.svelte';

	// ╔══════════════════════════════════════════════════════════════════════════╗
	// ║  🔐 SECRET PROMO CONFIGURATION                                           ║
	// ║  Change these values to update the promo or create new secret pages      ║
	// ║                                                                          ║
	// ║  HOW TO SET UP A NEW SECRET PROMO:                                       ║
	// ║  1. Create the plan in your admin dashboard (Admin → Subscriptions)      ║
	// ║  2. Note the plan's database ID (shown in the plan list)                 ║
	// ║  3. Put that ID in 'planId' below                                        ║
	// ║  4. Set your secret 'activeCode' (this becomes the URL)                  ║
	// ║  5. Update the display details to match                                  ║
	// ║                                                                          ║
	// ║  To change the promo URL: just change 'activeCode'                       ║
	// ║  Old URLs will show "Offer Not Available"                                ║
	// ╚══════════════════════════════════════════════════════════════════════════╝
	const PROMO_CONFIG = {
		// Current active promo code (must match URL path)
		// Example URL: /secretsub/XMAS2024
		activeCode: 'christmas2025',
		
		// Database Plan ID for this promo (NOT the Stripe product ID!)
		// Find this in Admin → Subscriptions → look at the plan's ID column
		// Example: '193' NOT 'prod_Tez86imWlBwdSp'
		// The plan must exist in your subscription_plans table. remember to update to prod id when ready to push
		planId: '15',
		
		// Plan display details (should match what's in the database)
		plan: {
			name: 'VIP Monthly',
			description: 'Enter promocode <br><span style="font-size:2rem; color: lightgreen;">3DOLLAR</span><br> at checkout to get the first month for <span style="font-size:2rem; color: red;">ONLY $3.00!</span>', 
			price: 7.97,
			currency: 'USD',
			interval: 'month' as 'month' | 'year',
			// Optional: seasonal theme - 'christmas', 'easter', 'halloween', 'valentines', or ''
			seasonTheme: 'christmas',
		},
		
		// Page customization
		pageTitle: '🎄 Exclusive Holiday Offer',
		pageSubtitle: 'You\'ve unlocked a special subscription rate!',
		ctaText: 'Claim This Offer',
		
		// Show "was $X" crossed out price (set to 0 to hide)
		originalPrice: 0
	};
	// ╔══════════════════════════════════════════════════════════════════════════╝

	let user: any = null;
	let loading = true;
	let isAuthenticated = false;
	let isValidCode = false;
	
	// Embedded checkout state
	let showEmbeddedCheckout = false;
	let checkoutClosing = false;
	let checkoutContainer: HTMLElement;
	let stripe: any = null;
	let checkout: any = null;
	let checkoutLoading = false;
	let checkoutError = '';

	// Get the code from the URL
	$: urlCode = $page.params.code;

	// Subscribe to auth store
	auth.subscribe((state: any) => {
		isAuthenticated = state.isAuthenticated;
		user = state.user;
	});

	onMount(async () => {
		// Validate the promo code
		if (urlCode?.toUpperCase() === PROMO_CONFIG.activeCode.toUpperCase()) {
			isValidCode = true;
		} else {
			isValidCode = false;
			loading = false;
			return;
		}

		loading = false;

		// 🔗 AUTO-CHECKOUT: Check if user just completed login and should auto-open checkout
		if (isAuthenticated && typeof window !== 'undefined') {
			const urlParams = new URLSearchParams(window.location.search);
			const autoCheckout = urlParams.get('auto_checkout') === 'true';
			
			if (autoCheckout) {
				showToast(`Opening checkout for ${PROMO_CONFIG.plan.name}...`, 'success');
				
				// Wait for UI to settle, then open checkout
				setTimeout(() => {
					startEmbeddedCheckout();
				}, 800);
				
				// Clean up URL (remove query params)
				window.history.replaceState({}, '', `/secretsub/${urlCode}`);
			}
		}
	});

	const handleSubscribe = async () => {
		if (!isAuthenticated) {
			// 🔗 CONTEXT PRESERVATION: Save return info for post-login flow
			sessionStorage.setItem('selected_plan_id', PROMO_CONFIG.planId);
			sessionStorage.setItem('selected_plan_name', PROMO_CONFIG.plan.name);
			
			showToast(`Please sign up to claim your ${PROMO_CONFIG.plan.name} offer`, 'info');
			
			// Redirect to login with return context
			goto(`/auth/register?return=${encodeURIComponent(`/secretsub/${urlCode}`)}&plan_id=secret`);
			return;
		}

		// ✅ USER IS AUTHENTICATED: Start checkout
		await startEmbeddedCheckout();
	};

	const startEmbeddedCheckout = async () => {
		showEmbeddedCheckout = true;
		checkoutLoading = true;
		checkoutError = '';

		try {
			await initializeEmbeddedCheckout();
		} catch (err) {
			console.error('Error initializing checkout:', err);
			checkoutError = 'Failed to initialize checkout';
			checkoutLoading = false;
		}
	};

	const initializeEmbeddedCheckout = async () => {
		try {
			// Load Stripe.js if not already loaded
			if (!window.Stripe) {
				await loadStripeJS();
			}

			// Get publishable key and initialize Stripe
			const configResponse = await apiRequest('/stripe/config');
			if (!configResponse.ok) {
				throw new Error('Failed to get Stripe configuration');
			}
			
			const { publishable_key } = await configResponse.json();
			if (!publishable_key) {
				throw new Error('Stripe publishable key not configured');
			}

			stripe = window.Stripe(publishable_key);
			await mountCheckout();
		} catch (err) {
			console.error('Error initializing Stripe:', err);
			throw err;
		}
	};

	const loadStripeJS = () => {
		return new Promise((resolve, reject) => {
			const script = document.createElement('script');
			script.src = 'https://js.stripe.com/v3/';
			script.onload = resolve;
			script.onerror = reject;
			document.head.appendChild(script);
		});
	};

	const mountCheckout = async () => {
		try {
			// Ensure any existing checkout is properly destroyed first
			if (checkout) {
				try {
					checkout.destroy();
				} catch (e) {
					console.warn('Error destroying previous checkout:', e);
				}
				checkout = null;
			}

			const fetchClientSecret = async () => {
				// 🔐 Use the SECRET PROMO endpoint - allows is_active=false plans and uses exact price from DB
				const response = await apiRequest('/stripe/secret-checkout-session', {
					method: 'POST',
					body: JSON.stringify({
						plan_id: PROMO_CONFIG.planId,
						return_url: `${window.location.origin}/checkout/success?session_id={CHECKOUT_SESSION_ID}`
					})
				});

				if (!response.ok) {
					const errorData = await response.json();
					
					// 🔒 Handle "already subscribed" case (HTTP 409 Conflict)
					if (response.status === 409 && errorData.action === 'redirect_dashboard') {
						const supportEmail = errorData.support_email || 'support@bookofmormonevidence.org';
						const betaMessage = `You already have an active subscription! Want to change? Contact ${supportEmail}`;
						
						showToast(betaMessage, 'info');
						setTimeout(() => {
							goto('/dashboard?tab=subscription');
						}, 3000);
						throw new Error('Redirecting to subscription dashboard...');
					}
					
					throw new Error(errorData.error || 'Failed to create checkout session');
				}

				const { client_secret } = await response.json();
				return client_secret;
			};

			// Initialize embedded checkout
			checkout = await stripe.initEmbeddedCheckout({
				fetchClientSecret
			});

			// Wait a moment for DOM to be ready, then mount checkout
			await new Promise(resolve => setTimeout(resolve, 100));
			
			if (!checkoutContainer) {
				throw new Error('Checkout container not found');
			}
			
			console.log('✅ Mounting Stripe checkout for secret promo...');
			checkout.mount(checkoutContainer);
			checkoutLoading = false;
		} catch (err: any) {
			console.error('Error mounting checkout:', err);
			checkoutError = err.message || 'Failed to load checkout';
			checkoutLoading = false;
		}
	};

	const closeEmbeddedCheckout = () => {
		// Immediately destroy the checkout
		if (checkout) {
			try {
				checkout.destroy();
			} catch (e) {
				console.warn('Error destroying checkout on close:', e);
			}
			checkout = null;
		}

		// Start closing animation
		checkoutClosing = true;
		
		setTimeout(() => {
			showEmbeddedCheckout = false;
			checkoutClosing = false;
			checkoutError = '';
		}, 400);
	};

	const formatPrice = (price: number, currency: string = 'USD') => {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: currency
		}).format(price);
	};

	const getSeasonClass = () => {
		const theme = PROMO_CONFIG.plan.seasonTheme?.toLowerCase();
		if (theme === 'christmas') return 'christmas-theme';
		if (theme === 'easter') return 'easter-theme';
		if (theme === 'halloween') return 'halloween-theme';
		if (theme === 'valentines') return 'valentines-theme';
		return '';
	};

	const getSavingsPercent = () => {
		if (!PROMO_CONFIG.originalPrice || PROMO_CONFIG.originalPrice <= PROMO_CONFIG.plan.price) {
			return 0;
		}
		return Math.round((1 - PROMO_CONFIG.plan.price / PROMO_CONFIG.originalPrice) * 100);
	};
</script>

<svelte:head>
	<title>{PROMO_CONFIG.pageTitle} - BOME</title>
	<meta name="description" content="Exclusive subscription offer - available only through this private link" />
	<meta name="robots" content="noindex, nofollow" />
</svelte:head>

<Navigation />

<div class="secret-sub-page">
	{#if loading}
		<div class="loading-container">
			<LoadingSpinner />
			<p>Loading offer...</p>
		</div>
	{:else if !isValidCode}
		<!-- Invalid or Expired Code -->
		<div class="invalid-code-container">
			<div class="invalid-card">
				<div class="invalid-icon">🔒</div>
				<h1>Offer Not Available</h1>
				<p>This promotional link has expired or is no longer valid.</p>
				<p class="subtext">If you believe this is an error, please contact our support team.</p>
				<div class="invalid-actions">
					<button class="btn btn-primary" on:click={() => goto('/subscription')}>
						View Available Plans
					</button>
					<button class="btn btn-outline" on:click={() => goto('/')}>
						Go Home
					</button>
				</div>
			</div>
		</div>
	{:else}
		<!-- Valid Promo Code - Show Offer -->
		<div class="promo-container">
			<header class="promo-header">
				<h1>{PROMO_CONFIG.pageTitle}</h1>
				<p>{PROMO_CONFIG.pageSubtitle}</p>
			</header>

			{#if !isAuthenticated}
				<div class="auth-notice">
					<p class="text-center">Sign up with the email you recieved the link with to claim this exclusive offer!</p>
				</div>
			{/if}

			<div class="plan-showcase">
				<div class="outerNew plan-outer {getSeasonClass()}">
					<!-- Floating % OFF badge -->
					{#if getSavingsPercent() > 0}
						<div class="discount-tab promo">
							<span class="discount-value">{getSavingsPercent()}% OFF</span>
						</div>
					{/if}
					
					{#if getSeasonClass() === 'christmas-theme'}
						<div class="promo-badge christmas-badge">🎄 Holiday Special</div>
					{:else if getSeasonClass() === 'easter-theme'}
						<div class="promo-badge easter-badge">🐰 Easter Special</div>
					{:else}
						<div class="promo-badge">Exclusive Offer</div>
					{/if}
					
					<div class="cardNew plan-card {getSeasonClass()}">
						<div class="rayNew"></div>
						
						<div class="plan-content">
							<div class="plan-main">
								<div class="plan-header">
									<h1>{PROMO_CONFIG.plan.name}</h1>
								</div>

								<div class="plan-pricing">
									{#if PROMO_CONFIG.originalPrice > 0}
										<div class="original-price">
											{formatPrice(PROMO_CONFIG.originalPrice, PROMO_CONFIG.plan.currency)}
										</div>
									{/if}
									<div class="hero-price">
										{formatPrice(PROMO_CONFIG.plan.price, PROMO_CONFIG.plan.currency)}
										<span class="price-interval">/{PROMO_CONFIG.plan.interval}</span>
									</div>
								</div>

								<div class="plan-description">
									<p>{@html PROMO_CONFIG.plan.description}</p>
								</div>
							</div>

							<button 
								class="btn btn-primary btn-full btn-cta" 
								on:click={handleSubscribe}
							>
								{#if !isAuthenticated}
									Sign Up to {PROMO_CONFIG.ctaText}
								{:else}
									{PROMO_CONFIG.ctaText}
								{/if}
							</button>

							{#if PROMO_CONFIG.plan.interval === 'year'}
								<div class="billed-at">
									Billed annually at {formatPrice(PROMO_CONFIG.plan.price, PROMO_CONFIG.plan.currency)}
								</div>
							{/if}
						</div>
					</div>
				</div>
			</div>

			<div class="trust-badges">
				<div class="badge">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
					</svg>
					<span>Secure Checkout</span>
				</div>
				<div class="badge">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
						<polyline points="22 4 12 14.01 9 11.01"></polyline>
					</svg>
					<span>Cancel Anytime</span>
				</div>
				<div class="badge">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<rect x="1" y="4" width="22" height="16" rx="2" ry="2"></rect>
						<line x1="1" y1="10" x2="23" y2="10"></line>
					</svg>
					<span>Powered by Stripe</span>
				</div>
			</div>
		</div>
	{/if}
</div>

<!-- Embedded Checkout Modal -->
{#if showEmbeddedCheckout}
	<div class="checkout-overlay" class:closing={checkoutClosing} on:click={closeEmbeddedCheckout}>
		<div class="embedded-checkout-container" class:closing={checkoutClosing} on:click|stopPropagation>
			<div class="embedded_spacer"></div>
			<button class="close-checkout" on:click={closeEmbeddedCheckout}>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<line x1="18" y1="6" x2="6" y2="18"></line>
					<line x1="6" y1="6" x2="18" y2="18"></line>
				</svg>
			</button>
			<div class="checkout-content">
				<div class="stripe-checkout-wrapper" bind:this={checkoutContainer}>
					<!-- Stripe Embedded Checkout will be mounted here -->
				</div>
				
				{#if checkoutLoading}
					<div class="checkout-loading">
						<LoadingSpinner />
						<p>Loading secure checkout...</p>
					</div>
				{:else if checkoutError}
					<div class="checkout-error">
						<div class="error-icon">⚠️</div>
						<h3>Checkout Error</h3>
						<p>{checkoutError}</p>
						<button class="btn btn-primary" on:click={closeEmbeddedCheckout}>
							Try Again
						</button>
					</div>
				{/if}
			</div>
		</div>
	</div>
{/if}

<Footer />

<style>

	.text-center {
		color: var(--text-primary);
		text-align: center;
	}

	.secret-sub-page {
		min-height: 100vh;
		background: var(--bg-primary);
		padding-top: 2rem;
	}

	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 60vh;
		gap: 1rem;
	}

	.loading-container p {
		color: var(--text-secondary);
		font-size: 1.1rem;
	}

	/* Invalid Code Styles */
	.invalid-code-container {
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 70vh;
		padding: 2rem;
	}

	.invalid-card {
		background: var(--bg-secondary);
		border-radius: 24px;
		padding: 3rem;
		text-align: center;
		max-width: 500px;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
	}

	.invalid-icon {
		font-size: 4rem;
		margin-bottom: 1.5rem;
	}

	.invalid-card h1 {
		font-size: 1.75rem;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.invalid-card p {
		color: var(--text-secondary);
		font-size: 1.1rem;
		margin-bottom: 0.5rem;
	}

	.invalid-card .subtext {
		font-size: 0.9rem;
		color: var(--text-muted);
		margin-bottom: 2rem;
	}

	.invalid-actions {
		display: flex;
		gap: 1rem;
		justify-content: center;
		flex-wrap: wrap;
	}

	/* Promo Container */
	.promo-container {
		max-width: 600px;
		margin: 0 auto;
		padding: 2rem;
	}

	.promo-header {
		text-align: center;
		margin-bottom: 2rem;
	}

	.promo-header h1 {
		font-size: 2.5rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.promo-header p {
		font-size: 1.2rem;
		color: var(--text-secondary);
	}

	.auth-notice {
		background: linear-gradient(135deg, rgba(212, 175, 55, 0.1), rgba(212, 175, 55, 0.05));
		border: 1px solid rgba(212, 175, 55, 0.3);
		color: var(--text-inverse);
		padding: 1rem 1.5rem;
		border-radius: 12px;
		margin-bottom: 2rem;
		text-align: center;
	}

	.auth-notice p {
		margin: 0;
		font-size: 1rem;
	}

	/* Plan Showcase - Center the single card */
	.plan-showcase {
		display: flex;
		justify-content: center;
		margin-bottom: 3rem;
	}

	/* Plan Card Styles - Copied from subscription page */
	.plan-outer {
		width: 360px;
		min-height: 480px;
		position: relative;
		padding-top: 20px;
	}

	.plan-outer::before {
		content: '';
		position: absolute;
		inset: 20px -2px -2px -2px;
		border-radius: 22px;
		background: radial-gradient(circle at center, rgba(212, 175, 55, 0.3), transparent 70%);
		animation: breathe 8s ease-in-out infinite;
		z-index: -1;
		pointer-events: none;
	}

	@keyframes breathe {
		0%, 100% { opacity: 0.5; transform: scale(1); }
		50% { opacity: 1; transform: scale(1.02); }
	}

	.discount-tab {
		position: absolute;
		top: -15px;
		right: 24px;
		z-index: 20;
		background: linear-gradient(135deg, #00d4aa 0%, #00b894 100%);
		color: #0a3d31;
		padding: 0.5rem 1rem;
		border-radius: 8px;
		font-size: 1.2rem;
		font-weight: 700;
		box-shadow: 0 4px 12px rgba(0, 212, 170, 0.4);
	}

	.discount-tab.promo {
		background: linear-gradient(135deg, #ffd700 0%, #daa520 100%);
		color: #1a1a2e;
	}

	.discount-value {
		letter-spacing: 0.5px;
		text-transform: uppercase;
	}

	.promo-badge {
		position: absolute;
		top: 20px;
		left: 24px;
		color: white;
		background: linear-gradient(135deg, var(--primary-gold), var(--primary-gold-dark));
		padding: 0.5rem 1rem;
		border-radius: 0 0 10px 10px;
		font-size: 0.75rem;
		font-weight: 600;
		z-index: 10;
		box-shadow: 0 4px 6px rgba(0, 0, 0, 0.2);
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.plan-card {
		border-radius: 20px;
		padding: 2rem;
		border: 1px solid var(--border-color);
		position: relative;
		overflow: hidden;
		width: 100%;
		height: 100%;
		background: var(--bg-primary);
		box-shadow: 7px 7px 14px var(--bg-dark);
		display: flex;
		flex-direction: column;
	}

	.plan-content {
		display: flex;
		flex-direction: column;
		height: 100%;
		gap: 1rem;
	}

	.plan-main {
		flex: 1;
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		gap: 1rem;
	}

	.plan-header {
		text-align: center;
	}

	.plan-header h1 {
		font-size: 2rem;
		font-weight: 700;
		color: var(--primary-gold-dark);
		margin: 0;
	}

	.plan-pricing {
		text-align: center;
	}

	.original-price {
		font-size: 1.1rem;
		color: var(--text-muted);
		text-decoration: line-through;
		opacity: 0.7;
		margin-bottom: 0.25rem;
	}

	.hero-price {
		font-size: 3.5rem;
		font-weight: 900;
		color: var(--primary-gold);
		line-height: 1;
	}

	.price-interval {
		font-size: 1rem;
		font-weight: 400;
		color: var(--primary-gold-dark);
	}

	.plan-description {
		text-align: center;
	}

	.plan-description p {
		color: var(--text-muted);
		font-size: 0.95rem;
		line-height: 1.6;
		margin: 0;
	}

	.btn-cta {
		width: 100%;
		padding: 1rem 1.5rem;
		font-size: 1.1rem;
		font-weight: 600;
		border-radius: 12px;
		transition: all 0.2s ease;
	}

	.btn-cta:hover {
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(212, 175, 55, 0.3);
	}

	.billed-at {
		text-align: center;
		font-size: 0.875rem;
		color: var(--text-muted);
		padding-top: 0.75rem;
		border-top: 1px solid var(--border-color);
	}

	/* Trust Badges */
	.trust-badges {
		display: flex;
		justify-content: center;
		gap: 2rem;
		flex-wrap: wrap;
	}

	.badge {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		color: var(--text-secondary);
		font-size: 0.875rem;
	}

	.badge svg {
		width: 20px;
		height: 20px;
		color: var(--primary-gold);
	}

	/* Christmas Theme */
	.christmas-theme.plan-card {
		background: linear-gradient(145deg, #1a2e1a, #0f1f0f) !important;
		border-color: #c41e3a !important;
	}

	.christmas-theme .plan-header h1 {
		background: linear-gradient(135deg, #ff6b6b, #ffffff, #4ecdc4);
		background-size: 200% 200%;
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
		animation: christmasGlow 3s ease-in-out infinite;
	}

	@keyframes christmasGlow {
		0%, 100% { background-position: 0% 50%; }
		50% { background-position: 100% 50%; }
	}

	.christmas-theme .hero-price {
		color: #ff6b6b;
		text-shadow: 0 0 10px rgba(255, 107, 107, 0.3);
	}

	.christmas-badge {
		background: linear-gradient(135deg, #c41e3a, #8b0000) !important;
	}

	.plan-outer.christmas-theme::before {
		background: radial-gradient(circle at center, rgba(220, 20, 60, 0.3), rgba(34, 139, 34, 0.2), transparent 70%);
	}

	.plan-outer.christmas-theme .discount-tab {
		background: linear-gradient(135deg, #c41e3a 0%, #228b22 100%);
		color: white;
	}

	.plan-outer.christmas-theme .btn-cta {
		background: linear-gradient(135deg, #c41e3a, #8b0000);
		border: none;
	}

	.plan-outer.christmas-theme .btn-cta:hover {
		background: linear-gradient(135deg, #d42c4a, #a41010);
		box-shadow: 0 4px 15px rgba(196, 30, 58, 0.4);
	}

	/* Easter Theme */
	.easter-theme.plan-card {
		background: linear-gradient(145deg, #f8f4ff, #e8e0f0) !important;
		border-color: #9b59b6 !important;
	}

	.easter-theme .hero-price {
		color: #9b59b6;
	}

	.easter-badge {
		background: linear-gradient(135deg, #9b59b6, #8e44ad) !important;
	}

	/* Checkout Modal Styles */
	.checkout-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: #fbbe14;
		z-index: 1000;
		display: flex;
		align-items: flex-end;
		justify-content: center;
		animation: fadeIn 0.3s ease-out;
	}

	.checkout-overlay.closing {
		animation: fadeOut 0.4s ease-in forwards;
	}

	.embedded-checkout-container {
		background: var(--card-bg);
		border-radius: 20px 20px 0 0;
		width: 100%;
		max-width: 100vw;
		min-height: 100vh;
		overflow-y: auto;
		box-shadow: 0 -10px 25px rgba(0, 0, 0, 0.2);
		border: 1px solid var(--border-color);
		animation: slideUpFromBottom 0.4s ease-out;
	}

	.embedded-checkout-container.closing {
		animation: slideDownToBottom 0.4s ease-in forwards;
	}

	.embedded_spacer {
		height: 100px;
	}

	.close-checkout {
		position: fixed;
		top: 4rem;
		right: 2rem;
		background: rgba(0, 0, 0, 0.7);
		border: none;
		color: white;
		cursor: pointer;
		padding: 0.75rem;
		border-radius: 50%;
		z-index: 10001;
		backdrop-filter: blur(10px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
		transition: all 0.3s ease;
	}

	.close-checkout:hover {
		background: rgba(0, 0, 0, 0.9);
		transform: scale(1.1);
	}

	.close-checkout svg {
		width: 24px;
		height: 24px;
	}

	.checkout-content {
		padding: 0;
		min-height: 400px;
	}

	.stripe-checkout-wrapper {
		min-height: 400px;
	}

	.checkout-loading {
		text-align: center;
		padding: 3rem 0;
	}

	.checkout-loading p {
		margin-top: 1rem;
		color: var(--text-inverse);
		font-size: 1.1rem;
	}

	.checkout-error {
		text-align: center;
		padding: 3rem 2rem;
	}

	.checkout-error .error-icon {
		font-size: 3rem;
		margin-bottom: 1rem;
	}

	.checkout-error h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text-inverse);
		margin-bottom: 1rem;
	}

	.checkout-error p {
		color: var(--text-inverse);
		margin-bottom: 2rem;
	}

	@keyframes fadeIn {
		from { opacity: 0; }
		to { opacity: 1; }
	}

	@keyframes fadeOut {
		from { opacity: 1; }
		to { opacity: 0; }
	}

	@keyframes slideUpFromBottom {
		from { opacity: 0; transform: translateY(100%); }
		to { opacity: 1; transform: translateY(0); }
	}

	@keyframes slideDownToBottom {
		from { opacity: 1; transform: translateY(0); }
		to { opacity: 0; transform: translateY(100%); }
	}

	/* Button Styles */
	.btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 0.75rem 1.5rem;
		border-radius: 10px;
		font-weight: 600;
		font-size: 1rem;
		cursor: pointer;
		transition: all 0.2s ease;
		border: none;
	}

	.btn-primary {
		background: linear-gradient(135deg, var(--primary-gold), var(--primary-gold-dark));
		color: white;
	}

	.btn-primary:hover {
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(212, 175, 55, 0.4);
	}

	.btn-outline {
		background: transparent;
		border: 2px solid var(--border-color);
		color: var(--text-primary);
	}

	.btn-outline:hover {
		background: var(--bg-secondary);
		border-color: var(--primary-gold);
	}

	/* Responsive */
	@media (max-width: 768px) {
		.promo-header h1 {
			font-size: 1.75rem;
		}

		.plan-outer {
			width: 100%;
			max-width: 360px;
		}

		.hero-price {
			font-size: 2.5rem;
		}

		.trust-badges {
			flex-direction: column;
			align-items: center;
			gap: 1rem;
		}

		.invalid-actions {
			flex-direction: column;
		}
	}
</style>

