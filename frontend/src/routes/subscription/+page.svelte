<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, apiRequest } from '$lib/auth';
	import { publicPlansService, type PublicSubscriptionPlan, type PublicSubscriptionOffer } from '$lib/services/public-plans';
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
	});

	const loadSubscriptionData = async () => {
		try {
			loading = true;
			
			// Load all subscription data in one call
			const subscriptionData = await publicPlansService.getAllSubscriptionData();
			
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
			showToast('Please sign in to continue', 'warning');
			goto('/login');
			return;
		}

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
		// Start closing animation
		checkoutClosing = true;
		
		// Wait for animation to complete before hiding
		setTimeout(() => {
			showEmbeddedCheckout = false;
			checkoutClosing = false;
			checkoutPlan = null;
			checkoutError = '';
			if (checkout) {
				checkout.destroy();
				checkout = null;
			}
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

	const getItemName = (itemId: number | string | undefined) => {
		if (!itemId) return 'Unknown Item';
		
		const itemNames: Record<number, string> = {
			1: 'ebook',
			2: 'DVD',
			3: 'Expo Ticket'
		};
		return itemNames[Number(itemId)] || `Item ${itemId}`;
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
			<p>💡 <strong>New here?</strong> You can browse plans without signing up. Create an account when you're ready to subscribe!</p>
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
									<div class="plan-card promotional">
										<div class="plan-header">
											<h3>{plan.name}</h3>
										</div>
										<div class="plan-price">
											{formatPrice(plan.price, plan.currency)}
											<span class="interval">/{plan.interval}</span>
										</div>
										<div class="promo-badge">Limited Time</div>
										<br>
										
										<div class="plan-description">
											<p>{plan.description}</p>
										</div>

										<div class="plan-features">
											<ul>
												{#each plan.features.slice(0, 5) as feature}
													<li>
														<svg class="check-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
															<polyline points="20,6 9,17 4,12"></polyline>
														</svg>
														{feature}
													</li>
												{/each}
											</ul>
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

																		<button 
									class="btn btn-primary btn-full btn-bottom" 
									on:click={() => handleSelectPlan(plan)}
								>
									{#if !isAuthenticated}
										Sign In to Subscribe
									{:else}
										Subscribe to {plan.name}
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
								<div class="plan-card">
									<div class="plan-header">
										<h3>{plan.name}</h3>
									</div>
									<div class="plan-price">
										{formatPrice(plan.price, plan.currency)}
										<span class="interval">/{plan.interval}</span>
									</div>
									{#if plan.popular}
										<div class="popular-badge">Most Popular</div>
									{/if}
									
									<br>
									<div class="plan-description">
										<p>{plan.description}</p>
									</div>

									<div class="plan-features">
										<ul>
											{#each plan.features.slice(0, 5) as feature}
												<li>
													<svg class="check-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
														<polyline points="20,6 9,17 4,12"></polyline>
													</svg>
													{feature}
												</li>
											{/each}
										</ul>
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

									<button 
										class="btn btn-primary btn-full btn-bottom" 
										on:click={() => handleSelectPlan(plan)}
									>
										{#if !isAuthenticated}
											Sign In to Subscribe
										{:else}
											Subscribe to {plan.name}
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
				<button class="btn btn-outline" on:click={handleDeclineOffer}>
					No Thanks
				</button>
				<button class="btn btn-primary" on:click={handleAcceptOffer}>
					Add Offer to Cart
				</button>
			</div>
		</div>
	</div>
{/if}
<Footer />
<style>

	.btn-bottom {
		position: absolute;
		bottom: 1rem;
		left: 10%;
		width: 80% !important;
		height: 3rem;
		background: var(--secondary-gradient);
		color: var(--text-primary);
	}

	.subscription-page {
		min-height: 100vh;
		max-width: 100vw;
		padding: 0;
		background: var(--secondary-gradient);
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
		margin-bottom: 5rem;
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
		background: var(--bg-tertiary);
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
		display: flex;
		flex-direction: row;
		gap: 2rem;
		justify-content: center;
		align-items: center;
	}

	.plan-card {
		border-radius: 15px;
		padding: 2rem;
		border: 1px solid var(--border-color);
		transition: all 0.3s ease;
		position: relative;
		overflow: hidden;
		max-width: 350px;
		min-height: 450px;
		background: linear-gradient(145deg, var(--primary-bom), var(--primary-bom-dark));
		box-shadow:  7px 7px 14px #bebebe,
             -7px -7px 14px #ffffff;
	}

	.plan-card:hover {
		transform: translateY(-4px);
		box-shadow: var(--neumorphic-shadow-hover);
	}

	.plan-card.promotional {
		border: 2px solid var(--primary-color);
	}

	.plan-header {
		text-align: center;
		margin-bottom: 1.5rem;
		position: relative;
		text-align: center;
		display: flex;
		justify-content: center;
		align-items: center;
	}

	.plan-header h3 {
		font-size: 2rem;
		font-weight: 700;
		color: var(--primary-gold);
		margin-bottom: 0.5rem;
	}

	.plan-price {
		font-size: 2rem;
		font-weight: 900;
		color: var(--primary-gold);
		margin-bottom: 0.5rem;
		text-align: center;
	}

	.interval {
		font-size: 1rem;
		color: var(--primary-gold-dark);
		font-weight: 400;
	}

	.popular-badge,
	.promo-badge {
		position: absolute;
		top: -0.25rem;
		right: 1rem;
		color: white;
		padding: 0.5rem 1rem;
		border-radius: 0 0 10px 10px;
		font-size: 0.875rem;
		font-weight: 600;
	}

	.promo-badge {
		background: #ba8927;
	}

	.plan-description {
		text-align: center;
		margin-bottom: 1.5rem;
	}

	.plan-description p {
		color: var(--primary-gold-light);
		font-size: 1rem;
		line-height: 1.6;
	}

	.plan-features {
		margin-bottom: 1.5rem;
	}

	.plan-features ul {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.plan-features li {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		color: var(--text-primary);
		font-size: 0.95rem;
	}

	.check-icon {
		width: 18px;
		height: 18px;
		color: var(--success-color);
		flex-shrink: 0;
	}

	.offers-section {
		margin-bottom: 1.5rem;
		padding: 1rem;
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
	}
</style> 
