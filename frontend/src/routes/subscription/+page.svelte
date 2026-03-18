<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, apiRequest } from '$lib/auth';
	import { publicPlansService, type PublicSubscriptionPlan, type PublicSubscriptionOffer } from '$lib/services/public-plans-service';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import Navigation from '$lib/components/Navigation.svelte';
	import Footer from '$lib/components/Footer.svelte';

	let user: any = null;
	let availablePlans: PublicSubscriptionPlan[] = [];
	let availableOffers: PublicSubscriptionOffer[] = [];
	let loading = true;
	let isAuthenticated = false;
	let selectedPlan: PublicSubscriptionPlan | null = null;
	let selectedOffer: PublicSubscriptionOffer | null = null;
	let showOfferModal = false;
	let cart: { plan: PublicSubscriptionPlan; offer?: PublicSubscriptionOffer } | null = null;
	
	// Embedded checkout state
	let showEmbeddedCheckout = false;
	let checkoutClosing = false;
	let checkoutPlan: PublicSubscriptionPlan | null = null;
	let checkoutContainer: HTMLElement;
	let stripe: any = null;
	let checkout: any = null;
	let checkoutLoading = false;
	let checkoutError = '';

	// Subscribe to auth store
	auth.subscribe((state: any) => {
		isAuthenticated = state.isAuthenticated;
		user = state.user;
	});

	onMount(async () => {
		await loadSubscriptionData();
		
		// 🔗 AUTO-CHECKOUT: Check if user just completed registration and should auto-open checkout
		if (isAuthenticated && typeof window !== 'undefined') {
			const urlParams = new URLSearchParams(window.location.search);
			const autoCheckout = urlParams.get('auto_checkout') === 'true';
			const planId = urlParams.get('plan_id');
			
			if (autoCheckout && planId) {
				// Find the plan
				const plan = availablePlans.find(p => p.id === planId);
				
				if (plan) {
					showToast(`Opening checkout for ${plan.name}...`, 'success');
					
					// Wait for UI to settle, then open checkout
					setTimeout(() => {
						startEmbeddedCheckout(plan);
					}, 800);
					
					// Clean up URL (remove query params)
					window.history.replaceState({}, '', '/subscription');
				}
			}
		}
	});

	const loadSubscriptionData = async () => {
		try {
			loading = true;
			
			// Load all subscription data in one call
			const subscriptionData = await publicPlansService.getAllSubscriptionData();
			console.log('Subscription data:', subscriptionData);
			// Combine all plans
			availablePlans = [
				...subscriptionData.promotional_plans,
				...subscriptionData.standard_plans
			];
			
			// Set offers
			availableOffers = subscriptionData.offers;
			
			console.log('Loaded subscription data:', {
				promotionalPlans: subscriptionData.promotional_plans.length,
				standardPlans: subscriptionData.standard_plans.length,
				offers: subscriptionData.offers.length
			});
		} catch (err) {
			showToast('Failed to load subscription data', 'error');
			console.error('Error loading subscription data:', err);
		} finally {
			loading = false;
		}
	};

	const handleSelectPlan = async (plan: PublicSubscriptionPlan) => {
		if (!isAuthenticated) {
			// 🔗 CONTEXT PRESERVATION: Save selected plan for seamless post-registration flow
			sessionStorage.setItem('selected_plan_id', plan.id);
			sessionStorage.setItem('selected_plan_name', plan.name);
			
			showToast(`Please sign in to subscribe to ${plan.name}`, 'info');
			
			// Redirect to login with return context
			goto(`/auth/login?return=${encodeURIComponent('/subscription')}&plan_id=${plan.id}`);
			return;
		}

		// ✅ USER IS AUTHENTICATED: Continue with normal checkout flow
		selectedPlan = plan;
		
		// Check if there are any offers for this plan
		const planOffers = availableOffers.filter(offer => 
			offer.plan_id === parseInt(plan.id) && offer.is_active
		);

		if (planOffers.length > 0) {
			// Show offer modal with the first available offer
			selectedOffer = planOffers[0];
			showOfferModal = true;
		} else {
			// Start embedded checkout directly
			await startEmbeddedCheckout(plan);
		}
	};

	const startEmbeddedCheckout = async (plan: PublicSubscriptionPlan) => {
		checkoutPlan = plan;
		showEmbeddedCheckout = true;
		checkoutLoading = true;
		checkoutError = '';

		try {
			await initializeEmbeddedCheckout(plan);
		} catch (err) {
			console.error('Error initializing checkout:', err);
			checkoutError = 'Failed to initialize checkout';
			checkoutLoading = false;
		}
	};

	const initializeEmbeddedCheckout = async (plan: PublicSubscriptionPlan) => {
		try {
			// Load Stripe.js if not already loaded
			if (!window.Stripe) {
				await loadStripeJS();
			}

			// Get publishable key and initialize Stripe using authenticated API request
			const configResponse = await apiRequest('/stripe/config');
			if (!configResponse.ok) {
				throw new Error('Failed to get Stripe configuration');
			}
			
			const { publishable_key } = await configResponse.json();
			if (!publishable_key) {
				throw new Error('Stripe publishable key not configured');
			}

			stripe = window.Stripe(publishable_key);
			await mountCheckout(plan);
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

	const mountCheckout = async (plan: PublicSubscriptionPlan) => {
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
				const response = await apiRequest('/stripe/checkout-session', {
					method: 'POST',
					body: JSON.stringify({
						plan_id: plan.id,
						return_url: `${window.location.origin}/checkout/success?session_id={CHECKOUT_SESSION_ID}`
					})
				});

				if (!response.ok) {
					const errorData = await response.json();
					
					// 🔒 Handle "already subscribed" case (HTTP 409 Conflict)
					if (response.status === 409 && errorData.action === 'redirect_dashboard') {
						// User already has active subscription - show BETA message and redirect to dashboard
						const supportEmail = errorData.support_email || 'support@bookofmormonevidence.org';
						const betaMessage = `You already have an active subscription! Want to change your subscription while we're in BETA? Contact ${supportEmail}`;
						
						showToast(betaMessage, 'info');
						setTimeout(() => {
							goto('/dashboard?tab=subscription');
						}, 3000); // 3 seconds so they can read the message
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
			
			// Mount checkout - ensure container exists
			if (!checkoutContainer) {
				throw new Error('Checkout container not found - DOM element may not be rendered yet');
			}
			
			console.log('✅ Found checkout container, mounting Stripe checkout...');
			checkout.mount(checkoutContainer);
			checkoutLoading = false;
		} catch (err: any) {
			console.error('Error mounting checkout:', err);
			checkoutError = err.message || 'Failed to load checkout';
			checkoutLoading = false;
		}
	};

	const closeEmbeddedCheckout = () => {
		// Immediately destroy the checkout to prevent multiple checkout objects
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
		
		// Wait for animation to complete before hiding
		setTimeout(() => {
			showEmbeddedCheckout = false;
			checkoutClosing = false;
			checkoutPlan = null;
			checkoutError = '';
		}, 400); // Match the animation duration
	};

	const addToCart = (plan: PublicSubscriptionPlan, offer?: PublicSubscriptionOffer) => {
		cart = { plan, offer };
		showToast('Plan added to cart!', 'success');
		// Here you would typically navigate to checkout or show a cart modal
		console.log('Cart updated:', cart);
	};

	const handleAcceptOffer = async () => {
		if (selectedPlan && selectedOffer) {
			showOfferModal = false;
			// TODO: Handle offers in embedded checkout
			await startEmbeddedCheckout(selectedPlan);
			selectedOffer = null;
		}
	};

	const handleDeclineOffer = async () => {
		if (selectedPlan) {
			showOfferModal = false;
			await startEmbeddedCheckout(selectedPlan);
			selectedOffer = null;
		}
	};

	const removeFromCart = () => {
		cart = null;
		showToast('Cart cleared', 'info');
	};

	const removeOfferFromCart = () => {
		if (cart) {
			cart = { plan: cart.plan };
			showToast('Offer removed from cart', 'info');
		}
	};

	const getPlanOffers = (planId: string) => {
		return availableOffers.filter(offer => 
			offer.plan_id === parseInt(planId) && offer.is_active
		);
	};

	const formatPrice = (price: number, currency: string = 'USD') => {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: currency
		}).format(price);
	};

	const formatDiscount = (offer: PublicSubscriptionOffer) => {
		if (offer.off_discount_type === 'percentage') {
			return `${offer.off_discount_value}% off`;
		} else if (offer.off_discount_type === 'fixed') {
			return `${formatPrice(offer.off_discount_value)} off`;
		}
		return '';
	};

	// Find the monthly equivalent plan for an annual plan (or vice versa)
	const findEquivalentPlan = (plan: PublicSubscriptionPlan, targetInterval: 'month' | 'year') => {
		// First try to match by stripe_product_id (most reliable)
		if (plan.stripe_product_id) {
			const match = availablePlans.find(p => 
				p.stripe_product_id === plan.stripe_product_id &&
				p.interval === targetInterval &&
				p.is_active
			);
			if (match) return match;
		}

		// Fallback: Try to match by name pattern (e.g., "Monthly" vs "Yearly", or similar base names)
		// Remove interval-related words from the name for comparison
		const baseName = plan.name.toLowerCase()
			.replace(/\b(monthly|yearly|annual|month|year)\b/gi, '')
			.trim();
		
		return availablePlans.find(p => {
			const pBaseName = p.name.toLowerCase()
				.replace(/\b(monthly|yearly|annual|month|year)\b/gi, '')
				.trim();
			return pBaseName === baseName && 
				p.interval === targetInterval && 
				p.is_active &&
				p.sub_type === plan.sub_type; // Match same subscription type
		});
	};

	// Calculate monthly breakdown for annual plans
	const getMonthlyBreakdown = (plan: PublicSubscriptionPlan) => {
		if (plan.interval === 'year') {
			return plan.price / 12;
		}
		return null;
	};

	// Calculate savings for annual plans vs monthly
	const calculateAnnualSavings = (plan: PublicSubscriptionPlan) => {
		if (plan.interval !== 'year') return null;

		const monthlyPlan = findEquivalentPlan(plan, 'month');
		if (!monthlyPlan) return null;

		const annualCostIfBilledMonthly = monthlyPlan.price * 12;
		const actualAnnualCost = plan.price;
		const savings = annualCostIfBilledMonthly - actualAnnualCost;
		const savingsPercentage = Math.round((savings / annualCostIfBilledMonthly) * 100);

		return {
			amount: savings,
			percentage: savingsPercentage,
			monthlyEquivalent: monthlyPlan.price
		};
	};

	const getItemName = (itemId: number | string | undefined) => {
		if (!itemId) return 'Unknown Item';
		
		const itemNames: Record<number, string> = {
			1: 'ebook',
			2: 'DVD',
			3: 'Expo Ticket'
		};
		return itemNames[Number(itemId)] || `Item ${itemId}`;
	};

	/** Features to show in pricing card (excludes season: markers) */
	const getDisplayFeatures = (plan: PublicSubscriptionPlan): string[] => {
		if (!plan.features || !Array.isArray(plan.features)) return [];
		return plan.features.filter(f => 
			typeof f === 'string' && !f.toLowerCase().startsWith('season:')
		);
	};

	/** Format price for pricing-card display (matches home page: $9<span>.97/mo</span>) */
	const formatPriceForCard = (price: number, currency: string, suffix: string = '/mo') => {
		const formatted = formatPrice(price, currency);
		// formatPrice returns e.g. "$9.97" (en-US) - split before decimal
		const dotIdx = formatted.lastIndexOf('.');
		if (dotIdx >= 0) {
			return { whole: formatted.slice(0, dotIdx), frac: formatted.slice(dotIdx) + suffix };
		}
		return { whole: formatted, frac: suffix };
	};

	/**
	 * Detects seasonal themes from plan features
	 * Features format: ["season:", "Christmas"] or ["season:", "Easter"], etc.
	 * Returns CSS class name for the season, or empty string if no season
	 */
	const getSeasonClass = (plan: PublicSubscriptionPlan): string => {
		if (!plan.features || !Array.isArray(plan.features)) return '';
		
		// Look for season indicator
		const hasSeasonMarker = plan.features.some(f => 
			f.toLowerCase().includes('season:') || f.toLowerCase() === 'season'
		);
		
		if (!hasSeasonMarker) return '';
		
		// Check for specific seasons
		const features = plan.features.map(f => f.toLowerCase());
		if (features.includes('christmas')) return 'christmas-theme';
		if (features.includes('easter')) return 'easter-theme';
		if (features.includes('halloween')) return 'halloween-theme';
		if (features.includes('valentines') || features.includes("valentine's")) return 'valentines-theme';
		
		return '';
	};
</script>

<svelte:head>
	<title>Choose Your Subscription - BOME</title>
	<meta name="description" content="Choose from our premium subscription plans and exclusive offers" />
</svelte:head>
<Navigation />
<div class="subscription-page">
	<div class="container"  class:expandedLeft={cart !== null}>
		<header class="page-header">
			<!--<button class="back-button" on:click={() => goto('/')}>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M19 12H5"></path>
					<path d="M12 19l-7-7 7-7"></path>
				</svg>
				Back to Home
			</button>
			<h1>Choose Your Subscription</h1>
			<p>Browse our premium subscription plans and exclusive offers</p>-->
			
		</header>
		{#if !isAuthenticated}
		<div class="auth-notice">
			<p>Create an account when you're ready to subscribe! </p>
		</div>
	{/if}
		{#if loading}
			<div class="loading-container">
				<LoadingSpinner />
				<p>Loading subscription plans...</p>
			</div>
		{:else}
			<div class="subscription-layout">
				<!-- Plans Section (Left Side) -->
				<div class="plans-container" >
					<!-- Promotional Plans -->
					{#if availablePlans.filter(plan => plan.sub_type === 'prmo' && plan.is_active).length > 0}
						<div class="plans-section">
							<h2>Limited Time Offers</h2>
							<p class="section-description">Special promotional plans with exclusive benefits</p>
							
							<div class="plans-grid">
								{#each availablePlans.filter(plan => plan.sub_type === 'prmo' && plan.is_active) as plan}
									{@const monthlyBreakdown = plan.interval === 'year' ? getMonthlyBreakdown(plan) : null}
									{@const savings = plan.interval === 'year' ? calculateAnnualSavings(plan) : null}
									{@const displayFeatures = getDisplayFeatures(plan)}
									
								{@const priceParts = plan.interval === 'year' && monthlyBreakdown
									? formatPriceForCard(monthlyBreakdown, plan.currency, '/mo')
									: formatPriceForCard(plan.price, plan.currency, '/' + plan.interval)}
								<div class="pricing-card pricing-card--featured">
									{#if savings && savings.percentage > 0}
										<div class="pricing-badge">{savings.percentage}% OFF</div>
									{:else}
										<div class="pricing-badge">Limited Time</div>
									{/if}
									
									<div class="pricing-card__plan">{plan.name}</div>
									<div class="pricing-card__price">{priceParts.whole}<span>{priceParts.frac}</span></div>
									<div class="pricing-card__billed">
										{#if plan.interval === 'year'}
											Billed annually at {formatPrice(plan.price, plan.currency)}/year
										{:else}
											&nbsp;
										{/if}
									</div>

									{#if getPlanOffers(plan.id).length > 0}
										<div class="offers-section">
											<h4>Special Bonus</h4>
											{#each getPlanOffers(plan.id) as offer}
												<div class="offer-badge">
													<svg class="gift-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
														<polyline points="20,12 20,22 4,22 4,12"></polyline>
														<rect x="2" y="7" width="20" height="5"></rect>
														<line x1="12" y1="22" x2="12" y2="7"></line>
														<polyline points="7,7 12,2 17,7"></polyline>
														<polyline points="7,7 12,12 17,7"></polyline>
													</svg>
													{formatDiscount(offer)} - {getItemName(offer.item_id)}
												</div>
											{/each}
										</div>
									{/if}

									{#if displayFeatures.length > 0}
										<ul class="pricing-card__features">
											{#each displayFeatures as feature}
												<li>{feature}</li>
											{/each}
										</ul>
									{/if}

									<button 
										class="btn-sales btn-sales--gold pricing-btn" 
										on:click={() => handleSelectPlan(plan)}
									>
										{#if !isAuthenticated}
											Sign In to Subscribe
										{:else}
											Subscribe Now
										{/if}
									</button>
								</div>
								{/each}
							</div>
						</div>
					{/if}

					<!-- Standard Plans -->
					<div class="plans-section" >
						<h2>Standard Plans</h2>
						<p class="section-description">Choose from our standard subscription plans</p>
						
						<div class="plans-grid">
							{#each availablePlans.filter(plan => plan.sub_type === 'stnd') as plan}
								{@const monthlyBreakdown = plan.interval === 'year' ? getMonthlyBreakdown(plan) : null}
								{@const savings = plan.interval === 'year' ? calculateAnnualSavings(plan) : null}
								{@const displayFeatures = getDisplayFeatures(plan)}
								{@const priceParts = plan.interval === 'year' && monthlyBreakdown
									? formatPriceForCard(monthlyBreakdown, plan.currency, '/mo')
									: formatPriceForCard(plan.price, plan.currency, '/' + plan.interval)}
								{@const isFeatured = plan.interval === 'year' || plan.popular}
								
							<div class="pricing-card" class:pricing-card--featured={isFeatured}>
								{#if plan.interval === 'year' && savings && savings.percentage > 0}
									<div class="pricing-badge">{savings.percentage}% OFF</div>
								{:else if plan.popular || plan.interval === 'year'}
									<div class="pricing-badge">Best Value</div>
								{/if}
								
								<div class="pricing-card__plan">{plan.name}</div>
								<div class="pricing-card__price">{priceParts.whole}<span>{priceParts.frac}</span></div>
								<div class="pricing-card__billed">
									{#if plan.interval === 'year'}
										Billed annually at {formatPrice(plan.price, plan.currency)}/year
									{:else}
										&nbsp;
									{/if}
								</div>

								{#if getPlanOffers(plan.id).length > 0}
									<div class="offers-section">
										<h4>Special Bonus</h4>
										{#each getPlanOffers(plan.id) as offer}
											<div class="offer-badge">
												<svg class="gift-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
													<polyline points="20,12 20,22 4,22 4,12"></polyline>
													<rect x="2" y="7" width="20" height="5"></rect>
													<line x1="12" y1="22" x2="12" y2="7"></line>
													<polyline points="7,7 12,2 17,7"></polyline>
													<polyline points="7,7 12,12 17,7"></polyline>
												</svg>
												{formatDiscount(offer)} - {getItemName(offer.item_id)}
											</div>
										{/each}
									</div>
								{/if}

								{#if displayFeatures.length > 0}
									<ul class="pricing-card__features">
										{#each displayFeatures as feature}
											<li>{feature}</li>
										{/each}
									</ul>
								{/if}

								<button 
									class="btn-sales pricing-btn" 
									class:btn-sales--gold={isFeatured}
									class:btn-sales--charcoal={!isFeatured}
									on:click={() => handleSelectPlan(plan)}
								>
									{#if !isAuthenticated}
										Sign In to Subscribe
									{:else}
										Subscribe Now
									{/if}
								</button>
							</div>
							{/each}
						</div>
					</div>
				</div>

			</div>
		{/if}
	</div>
</div>

<!-- Embedded Checkout Section (Expands from Bottom) -->
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
				<!-- Always render the container so it's available for mounting -->
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

<!-- Offer Modal -->
{#if showOfferModal && selectedOffer && selectedPlan}
	<div class="modal-overlay" on:click={() => showOfferModal = false}>
		<div class="modal modal-offer" on:click|stopPropagation>
			<div class="modal-header">
				<h3>Special Offer Available!</h3>
				<button class="modal-close" on:click={() => showOfferModal = false}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"></line>
						<line x1="6" y1="6" x2="18" y2="18"></line>
					</svg>
				</button>
			</div>
			
			<div class="modal-content">
				<div class="offer-details">
					<div class="offer-header">
						<svg class="gift-icon-large" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<polyline points="20,12 20,22 4,22 4,12"></polyline>
							<rect x="2" y="7" width="20" height="5"></rect>
							<line x1="12" y1="22" x2="12" y2="7"></line>
							<polyline points="7,7 12,2 17,7"></polyline>
							<polyline points="7,7 12,12 17,7"></polyline>
						</svg>
						<h4>{selectedOffer.off_name}</h4>
					</div>
					
					<p class="offer-description">{selectedOffer.off_description}</p>
					
					<div class="offer-benefits">
						<div class="benefit-item">
							<strong>Discount:</strong> {formatDiscount(selectedOffer)}
						</div>
						<div class="benefit-item">
							<strong>Free Item:</strong> {getItemName(selectedOffer.item_id)}
						</div>
						{#if selectedOffer.off_max_uses && selectedOffer.off_max_uses > 0}
							<div class="benefit-item">
								<strong>Limited:</strong> Only {selectedOffer.off_max_uses - selectedOffer.off_current_uses} remaining
							</div>
						{/if}
					</div>
				</div>
			</div>
			
			<div class="modal-actions">
				<button class="btn-outline" on:click={handleDeclineOffer}>
					No Thanks
				</button>
				<button class="btn-primary" on:click={handleAcceptOffer}>
					Add Offer to Cart
				</button>
			</div>
		</div>
	</div>
{/if}
<Footer />
<style>
	/* Sales page design tokens (from home Choose Your Plan) */
	.subscription-page {
		min-height: 100vh;
		max-width: 100vw;
		padding: 0;
		background: var(--bg-primary);
		--sales-navy: #0D1B3E;
		--sales-gold: #B7953B;
		--sales-gold-light: #D4B255;
		--sales-charcoal: #1E2022;
		--sales-white: #FFFFFF;
		--sales-off-white: #F7F8FA;
		--sales-light-gray: #E8EAF0;
		--sales-text-muted: #9CA3AF;
		--sales-heading-font: 'Playfair Display', Georgia, serif;
		--sales-body-font: 'Poppins', 'Roboto', sans-serif;
	}

	.container {
		max-width: 100vw;
		margin: 0 auto;
		padding: 0 0;
		transition: max-width 1.75s 500ms ease-out;
	}

	.container.expandedLeft {
		max-width: 1800px !important;
	}

	.page-header {
		text-align: center;
		margin-bottom: 1rem;
		height: 165px;
		width:  100vw;
		background: var(--bg-primary);
	}

	.back-button {
		position: absolute;
		left: 2rem;
		top: 3rem;
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
		margin-top: 0.75rem;
	}

	.page-header p {
		font-size: 1.1rem;
		color: var(--text-secondary);
	}

	.loading-container {
		text-align: center;
		padding: 3rem 0;
	}

	.subscription-layout {
		display: flex;
		gap: 2rem;
		padding: 2rem;
		background: rgba(var(--primary-color-rgb), 0.0);
		max-width: 100vw;
	}

	.plans-container {
		flex: 2;
		display: flex;
		flex-direction: column;
		gap: 3rem;
	}


	.cart-container {
		width: 0;
		overflow: hidden;
		transition: width 1.7s 750ms ease-out;
		background: var(--bg-tertiary);
		border-radius: 20px;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
	}

	.cart-container.expanded {
		width: 450px;
	}

	.cart-content {
		padding: 2rem;
		min-width: 350px;
		transform-origin: right;
		animation: slideInFromRight 0.8s ease-out 0.5s both;
	}

	.cart-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1.5rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid var(--border-color);
	}

	.cart-header h2 {
		margin: 0;
		font-size: 1.5rem;
		font-weight: 600;
		color: var(--text-primary);
	}

	.cart-placeholder {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 3rem 0;
		color: var(--text-secondary);
		font-size: 1.1rem;
		min-width: 500px;
		animation: fadeIn 0.8s ease-out 0.5s both;
	}

	.cart-icon {
		width: 60px;
		height: 60px;
		margin-bottom: 1rem;
		color: var(--primary-color);
		animation: bounce 2s ease-in-out infinite;
	}

	@keyframes slideInFromRight {
		from {
			opacity: 0;
			transform: translateX(50px);
		}
		to {
			opacity: 1;
			transform: translateX(0);
		}
	}

	@keyframes fadeIn {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	@keyframes bounce {
		0%, 20%, 50%, 80%, 100% {
			transform: translateY(0);
		}
		40% {
			transform: translateY(-10px);
		}
		60% {
			transform: translateY(-5px);
		}
	}

	.plans-section {
		border-radius: 20px;
		padding: 2rem;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
	}

	.plans-section h2 {
		font-size: 1.75rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.section-description {
		color: var(--text-secondary);
		margin-bottom: 2rem;
		font-size: 1.1rem;
	}

	.plans-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
		gap: 32px;
		justify-content: center;
		align-items: stretch;
		max-width: 900px;
		margin: 0 auto;
	}

	/* Choose Your Plan pricing card styles (from home +page.svelte) */
	.pricing-card {
		border: 1px solid var(--sales-light-gray);
		padding: 40px 32px;
		text-align: center;
		position: relative;
		background: var(--sales-white);
		transition: box-shadow 0.3s ease;
		border-radius: 4px;
	}

	.pricing-card:hover {
		box-shadow: 5px 5px 20px rgba(0, 0, 0, 0.08);
	}

	.pricing-card--featured {
		border: 2px solid var(--sales-gold);
	}

	.pricing-badge {
		position: absolute;
		top: -14px;
		left: 50%;
		transform: translateX(-50%);
		background: var(--sales-gold);
		color: var(--sales-navy);
		font-size: 12px;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 1px;
		padding: 6px 20px;
		border-radius: 4px;
	}

	.pricing-card__plan {
		font-family: var(--sales-heading-font);
		font-size: 20px;
		color: var(--sales-navy);
		margin-bottom: 16px;
	}

	.pricing-card__price {
		font-size: 48px;
		font-weight: 700;
		color: var(--sales-navy);
		line-height: 1;
	}

	.pricing-card__price span {
		font-size: 18px;
		font-weight: 400;
		color: var(--sales-text-muted);
	}

	.pricing-card__billed {
		font-size: 13px;
		color: var(--sales-text-muted);
		margin-top: 4px;
		margin-bottom: 28px;
		min-height: 20px;
	}

	.pricing-card__features {
		list-style: none;
		text-align: left;
		margin-bottom: 32px;
		padding: 0;
	}

	.pricing-card__features li {
		padding: 8px 0;
		font-size: 14px;
		color: #555;
		border-bottom: 1px solid #f0f0f0;
		padding-left: 24px;
		position: relative;
	}

	.pricing-card__features li::before {
		content: "\2713";
		position: absolute;
		left: 0;
		color: var(--sales-gold);
		font-weight: 700;
	}

	.pricing-card__features li:last-child {
		border-bottom: none;
	}

	.pricing-btn {
		width: 100%;
	}

	/* Sales page button styles */
	.subscription-page .btn-sales {
		display: inline-block;
		font-family: var(--sales-body-font);
		font-weight: 600;
		font-size: 14px;
		letter-spacing: 1px;
		text-transform: uppercase;
		padding: 16px 40px;
		border: 2px solid transparent;
		cursor: pointer;
		transition: all 0.3s ease;
		text-decoration: none;
		border-radius: 4px;
		width: 100%;
	}

	.subscription-page .btn-sales--charcoal {
		background: var(--sales-charcoal);
		color: var(--sales-white);
		border-color: var(--sales-charcoal);
	}

	.subscription-page .btn-sales--charcoal:hover {
		background: transparent;
		color: var(--sales-charcoal);
		border-color: var(--sales-charcoal);
	}

	.subscription-page .btn-sales--gold {
		background: var(--sales-gold);
		color: var(--sales-navy);
		border-color: var(--sales-gold);
	}

	.subscription-page .btn-sales--gold:hover {
		background: transparent;
		color: var(--sales-gold);
		border-color: var(--sales-gold);
	}

	/* Offers section inside pricing card */
	.pricing-card .offers-section {
		margin-bottom: 24px;
		padding: 12px 16px;
		background: rgba(183, 149, 59, 0.1);
		border-radius: 8px;
		border: 1px solid rgba(183, 149, 59, 0.3);
		text-align: center;
	}

	.pricing-card .offers-section h4 {
		font-size: 12px;
		font-weight: 700;
		color: var(--sales-navy);
		margin-bottom: 8px;
		text-transform: uppercase;
		letter-spacing: 1px;
	}

	.pricing-card .offer-badge {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		font-size: 14px;
		color: var(--sales-navy);
		font-weight: 600;
	}

	.pricing-card .gift-icon {
		width: 16px;
		height: 16px;
		flex-shrink: 0;
	}

	/* Legacy plan-outer (kept for any remaining refs, can remove if unused) */
	.plan-outer {
		width: 320px;
		min-height: 520px;
		height: 520px; /* Fixed height for uniform cards */
		position: relative;
		padding-top: 20px; /* Space for floating discount tab */
	}

	.plan-outer::before {
		content: '';
		position: absolute;
		inset: 20px -2px -2px -2px; /* Adjusted for padding-top */
		border-radius: 22px;
		background: radial-gradient(circle at center, rgba(212, 175, 55, 0.3), transparent 70%);
		animation: breathe 8s ease-in-out infinite;
		z-index: -1;
		pointer-events: none;
	}

	/* Floating discount tab - looks like a card sticking up */
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
		box-shadow: 
			0 4px 12px rgba(0, 212, 170, 0.4),
			0 2px 4px rgba(0, 0, 0, 0.2);
		transform: translateY(0);
	}

	.discount-tab::after {
		content: '';
		position: absolute;
		bottom: -6px;
		left: 50%;
		transform: translateX(-50%);
		width: 0;
		height: 0;
		border-left: 8px solid transparent;
		border-right: 8px solid transparent;
		border-top: 6px solid #00b894;
	}

	.discount-value {
		letter-spacing: 0.5px;
		text-transform: uppercase;
	}

	.plan-card {
		border-radius: 20px;
		padding: 1.5rem;
		border: 1px solid var(--border-color);
		transition: all 0.3s ease;
		position: relative;
		overflow: hidden;
		width: 100%;
		height: 100%;
		background: var(--bg-primary);
		box-shadow: 7px 7px 14px var(--bg-dark);
		display: flex;
		flex-direction: column;
	}

	.plan-outer:hover {
		transform: translateY(-4px);
	}

	.plan-outer:hover .plan-card {
		box-shadow: var(--neumorphic-shadow-hover);
	}

	/* Plan content - flexbox for proper distribution */
	.plan-content {
		display: flex;
		flex-direction: column;
		height: 100%;
		gap: 0.75rem;
	}

	/* Wrapper for vertically centering main content */
	.plan-main {
		flex: 1;
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		gap: 0.5rem;
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

	/* Pricing section */
	.plan-pricing {
		text-align: left;
	}

	.original-price {
		font-size: 1rem;
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

	/* CTA Button */
	.btn-cta {
		margin-top: auto;
		width: 100%;
		padding: 0.875rem 1.5rem;
		font-size: 1rem;
		font-weight: 600;
		border-radius: 10px;
		transition: all 0.2s ease;
	}

	.btn-cta:hover {
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(212, 175, 55, 0.3);
	}

	/* Billed at footer */
	.billed-at {
		text-align: center;
		font-size: 0.89rem;
		color: var(--text-primary);
		padding-top: 0.75rem;
		margin-top: 0.5rem;
		border-top: 1px solid var(--border-color);
	}

	/* Promotional card variant */
	.plan-outer.promotional .plan-card {
		background: linear-gradient(145deg, #1a1a2e, #16213e);
		border-color: var(--primary-gold);
	}

	.discount-tab.promo {
		background: linear-gradient(135deg, #ffd700 0%, #daa520 100%);
		color: #1a1a2e;
	}

	.discount-tab.promo::after {
		border-top-color: #daa520;
	}

	/* ===============================================
	   🎄 CHRISTMAS THEME STYLES
	   =============================================== */
	
	/* Christmas card outer glow */
	.plan-outer.christmas-theme::before {
		background: radial-gradient(circle at center, rgba(220, 20, 60, 0.3), rgba(34, 139, 34, 0.2), transparent 70%);
	}

	/* Christmas card styling */
	.plan-card.christmas-theme {
		background: linear-gradient(145deg, #1a2e1a, #0f1f0f) !important;
		border-color: #c41e3a !important;
		position: relative;
		overflow: hidden;
	}

	/* Snowflakes animation overlay */
	.plan-card.christmas-theme::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background-image: 
			url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='white' opacity='0.1'%3E%3Cpath d='M12 0L14 8H22L15 13L18 22L12 17L6 22L9 13L2 8H10L12 0Z'/%3E%3C/svg%3E");
		background-size: 30px 30px;
		opacity: 0.05;
		pointer-events: none;
		animation: snowfall 20s linear infinite;
	}

	@keyframes snowfall {
		0% {
			background-position: 0 0, 50px 50px;
		}
		100% {
			background-position: 50px 100px, 100px 150px;
		}
	}

	/* Christmas color accents */
	.plan-outer.christmas-theme .plan-header h1 {
		background: linear-gradient(135deg, #ff6b6b, #ffffff, #4ecdc4);
		background-size: 200% 200%;
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
		animation: christmasGlow 3s ease-in-out infinite;
	}

	@keyframes christmasGlow {
		0%, 100% {
			background-position: 0% 50%;
		}
		50% {
			background-position: 100% 50%;
		}
	}

	/* Christmas pricing highlight */
	.plan-outer.christmas-theme .hero-price {
		color: #ff6b6b;
		text-shadow: 0 0 10px rgba(255, 107, 107, 0.3);
	}

	/* Christmas badge */
	.christmas-badge {
		background: linear-gradient(135deg, #c41e3a, #8b0000) !important;
		animation: christmasPulse 2s ease-in-out infinite;
	}

	@keyframes christmasPulse {
		0%, 100% {
			box-shadow: 0 4px 6px rgba(196, 30, 58, 0.3);
		}
		50% {
			box-shadow: 0 4px 20px rgba(196, 30, 58, 0.6), 0 0 30px rgba(34, 139, 34, 0.3);
		}
	}

	/* Christmas discount tab */
	.plan-outer.christmas-theme .discount-tab {
		background: linear-gradient(135deg, #c41e3a 0%, #228b22 100%);
		color: white;
	}

	.plan-outer.christmas-theme .discount-tab::after {
		border-top-color: #228b22;
	}

	/* Christmas button styling */
	.plan-outer.christmas-theme .btn-cta {
		background: linear-gradient(135deg, #c41e3a, #8b0000);
		border: none;
		position: relative;
		overflow: hidden;
	}

	.plan-outer.christmas-theme .btn-cta::before {
		content: '🎁';
		position: absolute;
		left: 15px;
		font-size: 1rem;
	}

	.plan-outer.christmas-theme .btn-cta:hover {
		background: linear-gradient(135deg, #d42c4a, #a41010);
		box-shadow: 0 4px 15px rgba(196, 30, 58, 0.4);
	}

	/* Christmas offers section */
	.plan-outer.christmas-theme .offers-section {
		background: rgba(196, 30, 58, 0.1);
		border-color: rgba(196, 30, 58, 0.3);
	}

	.plan-outer.christmas-theme .offers-section h4 {
		color: #ff6b6b;
	}

	/* Floating snowflakes decoration (top corners) */
	.plan-outer.christmas-theme::after {
		content: '❄️';
		position: absolute;
		top: 25px;
		right: 20px;
		font-size: 1.5rem;
		opacity: 0.8;
		animation: snowflakeFloat 3s ease-in-out infinite;
	}

	@keyframes snowflakeFloat {
		0%, 100% {
			transform: translateY(0) rotate(0deg);
		}
		50% {
			transform: translateY(-5px) rotate(180deg);
		}
	}

	/* ===============================================
	   END CHRISTMAS THEME STYLES
	   =============================================== */

	.popular-badge,
	.promo-badge {
		position: absolute;
		top: 20px; /* Adjusted for padding-top on plan-outer */
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

	.promo-badge {
		background: linear-gradient(135deg, #ba8927, #8b6914);
	}

	.plan-description {
		text-align: center;
		margin-bottom: 0.5rem;
	}

	.plan-description p {
		color: var(--text-muted);
		font-size: 0.875rem;
		line-height: 1.5;
		margin: 0;
	}

	.plan-features {
		flex: 1; /* Takes available space */
		overflow-y: auto;
		margin-bottom: 0.5rem;
	}

	.plan-features ul {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.plan-features li {
		display: flex;
		align-items: flex-start;
		gap: 0.5rem;
		color: var(--text-primary);
		font-size: 0.85rem;
		line-height: 1.4;
	}

	.check-icon {
		width: 16px;
		height: 16px;
		min-width: 16px;
		color: var(--success-color);
		flex-shrink: 0;
		margin-top: 2px;
	}

	.offers-section {
		margin-bottom: 0.5rem;
		padding: 0.75rem;
		background: rgba(var(--primary-color-rgb), 0.1);
		border-radius: 12px;
		border: 1px solid rgba(var(--primary-color-rgb), 0.2);
		text-align: center;
	}

	.offers-section h4 {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--primary-color);
		margin-bottom: 0.5rem;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.offer-badge {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.875rem;
		color: var(--primary-color);
		font-weight: 600;
		text-align: center;
		justify-content: center;
	}

	.gift-icon {
		width: 16px;
		height: 16px;
	}

	.btn-full {
		width: 100%;
	}

	.cart-summary {
		background: var(--bg-secondary);
		border-radius: 16px;
		padding: 0rem;
	}

	.cart-item {
		border: 1px solid rgb(46, 46, 46);
		margin-bottom: 1rem;
		padding: 0.5rem;
		padding-left: 1rem;
		border-radius: 16px;
		background: linear-gradient(145deg, #cacaca, #f0f0f0);
		box-shadow:  7px 7px 14px #bebebe,
             -7px -7px 14px #ffffff;
	}

	.cart-plan-cont {
		display: flex;
		gap: 1rem;
		align-items: center;
		/*flex-direction: column;
		justify-content: space-between;
		align-items: flex-start;
		
		height: 100%;
		width: 100%;
		border-radius: 0; */
	}

	.cart-plan-cont button {
		justify-content: center;
		align-items: center;
	}

	.cart-item h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.25rem;
		flex: 1;
	}

	.cart-item p {
		color: var(--primary-color);
		font-weight: 600;
		margin: 0;
	}

	.cart-offer {
		margin-bottom: 1.5rem;
		padding: 1rem;
		background: rgba(var(--primary-color-rgb), 0.1);
		border-radius: 12px;
		border: 1px solid rgba(var(--primary-color-rgb), 0.2);
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 1rem;
	}

	.cart-offer h4 {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--primary-color);
		margin-bottom: 0.25rem;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.cart-offer p {
		color: var(--primary-color);
		font-weight: 600;
		margin: 0;
		flex: 1;
	}

	.auth-required {
		text-align: center;
		padding: 1.5rem;
		background: rgba(var(--primary-color-rgb), 0.1);
		border-radius: 12px;
		border: 1px solid rgba(var(--primary-color-rgb), 0.2);
	}

	.auth-required p {
		color: var(--text-secondary);
		font-size: 1rem;
		margin-bottom: 1rem;
	}

	.auth-buttons {
		display: flex;
		justify-content: center;
		gap: 1rem;
	}

	.checkout-options {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	/* Embedded Checkout Styles */
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

	.checkout-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 2rem 2rem 1rem 2rem;
		border-bottom: 1px solid var(--border-color);
		background: var(--bg-secondary);
		border-radius: 20px 20px 0 0;
	}

	.checkout-plan-info h2 {
		font-size: 1.5rem;
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 0.5rem 0;
	}

	.checkout-price {
		font-size: 1.25rem;
		font-weight: 700;
		color: var(--primary-color);
		margin: 0;
	}

	.embedded_spacer {
		position: relative;
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
		transition: all 0.3s ease;
		z-index: 10001;
		backdrop-filter: blur(10px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
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
		font-size: 1rem;
	}

	.stripe-checkout-wrapper {
		min-height: 400px;
		/* Stripe will inject the checkout form here */
	}

	@keyframes fadeIn {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	@keyframes slideUpFromBottom {
		from {
			opacity: 0;
			transform: translateY(100%);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	@keyframes slideDownToBottom {
		from {
			opacity: 1;
			transform: translateY(0);
		}
		to {
			opacity: 0;
			transform: translateY(100%);
		}
	}

	@keyframes fadeOut {
		from {
			opacity: 1;
		}
		to {
			opacity: 0;
		}
	}

	.auth-notice {
		background: var(--info-color);
		color: var(--text-primary);
		padding: 0 1rem;
		border-radius: 12px;
		margin-top: 0;
		font-size: 1.5rem;
		font-weight: 500;
		line-height: 1.5;
		text-align: center;
	}

	.auth-notice p {
		color: var(--text-primary);
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

	.modal-offer {
		background: var(--card-bg);
		border-radius: 20px;
		padding: 0;
		max-width: 500px;
		width: 90%;
		max-height: 90vh;
	}

	.modal {
		background: var(--bg-tertiary);
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

	.offer-details {
		text-align: center;
	}

	.offer-header {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
		margin-bottom: 1.5rem;
	}

	.gift-icon-large {
		width: 48px;
		height: 48px;
		color: var(--primary-color);
	}

	.offer-header h4 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.offer-description {
		color: var(--text-secondary);
		line-height: 1.6;
		margin-bottom: 1.5rem;
	}

	.offer-benefits {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		margin-bottom: 1.5rem;
	}

	.benefit-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.75rem;
		background: var(--bg-secondary);
		border-radius: 8px;
		font-size: 0.95rem;
	}

	.benefit-item strong {
		color: var(--text-primary);
	}

	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 1rem;
		padding: 2rem;
		border-top: 1px solid var(--border-color);
		margin-top: 2rem;
	}

	.close-cart-item {
		background: none;
		border: none;
		color: rgb(107, 53, 53);
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 8px;
		transition: all 0.3s ease;
		
	}

	.close-cart-item:hover {
		background: var(--bg-secondary);
		color: var(--text-primary);
	}

	.close-cart-item svg {
		width: 20px;
		height: 20px;
	}

	.close-cart-offer {
		background: none;
		border: none;
		color: var(--text-secondary);
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 8px;
		transition: all 0.3s ease;
	}

	.close-cart-offer:hover {
		background: var(--bg-secondary);
		color: var(--text-primary);
	}

	.close-cart-offer svg {
		width: 20px;
		height: 20px;
	}

	.close-cart {
		background: none;
		border: none;
		color: var(--text-secondary);
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 8px;
		transition: all 0.3s ease;
	}

	.close-cart:hover {
		background: var(--bg-secondary);
		color: var(--text-primary);
	}

	.close-cart svg {
		width: 20px;
		height: 20px;
	}

	@media (max-width: 768px) {
		.back-button {
			position: static;
			justify-content: center;
			margin-bottom: 1rem;
		}

		.subscription-layout {
			flex-direction: column;
		}

		.plans-container {
			flex: none;
			width: 100%;
		}

		.plans-grid {
			grid-template-columns: 1fr;
			max-width: 400px;
			margin-left: auto;
			margin-right: auto;
		}

		.cart-container {
			flex: none;
			width: 100%;
		}

		.cart-container.expanded {
			width: 100%;
		}

		.modal-actions {
			flex-direction: column;
		}

		.benefit-item {
			flex-direction: column;
			align-items: flex-start;
			gap: 0.25rem;
		}

		.hero-price {
			font-size: 2rem;
		}
	}
</style> 
