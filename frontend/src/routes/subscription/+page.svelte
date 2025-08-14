<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import { publicPlansService, type PublicSubscriptionPlan, type PublicSubscriptionOffer } from '$lib/services/public-plans';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let user: any = null;
	let availablePlans: PublicSubscriptionPlan[] = [];
	let availableOffers: PublicSubscriptionOffer[] = [];
	let loading = true;
	let isAuthenticated = false;
	let selectedPlan: PublicSubscriptionPlan | null = null;
	let selectedOffer: PublicSubscriptionOffer | null = null;
	let showOfferModal = false;
	let cart: { plan: PublicSubscriptionPlan; offer?: PublicSubscriptionOffer } | null = null;

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

	const handleSelectPlan = (plan: PublicSubscriptionPlan) => {
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
			// No offers, add plan directly to cart
			addToCart(plan);
		}
	};

	const addToCart = (plan: PublicSubscriptionPlan, offer?: PublicSubscriptionOffer) => {
		cart = { plan, offer };
		showToast('Plan added to cart!', 'success');
		// Here you would typically navigate to checkout or show a cart modal
		console.log('Cart updated:', cart);
	};

	const handleAcceptOffer = () => {
		if (selectedPlan && selectedOffer) {
			addToCart(selectedPlan, selectedOffer);
			showOfferModal = false;
			selectedOffer = null;
		}
	};

	const handleDeclineOffer = () => {
		if (selectedPlan) {
			addToCart(selectedPlan);
			showOfferModal = false;
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

<div class="subscription-page">
	<div class="container"  class:expandedLeft={cart !== null}>
		<header class="page-header">
			<button class="back-button" on:click={() => goto('/')}>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M19 12H5"></path>
					<path d="M12 19l-7-7 7-7"></path>
				</svg>
				Back to Home
			</button>
			<h1>Choose Your Subscription</h1>
			<p>Browse our premium subscription plans and exclusive offers</p>
			{#if !isAuthenticated}
				<div class="auth-notice">
					<p>💡 <strong>New here?</strong> You can browse plans without signing up. Create an account when you're ready to subscribe!</p>
				</div>
			{/if}
		</header>

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
											Select {plan.name}
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
										Select {plan.name}
									</button>
								</div>
							{/each}
						</div>
					</div>
				</div>

				<!-- Cart Section (Right Side) -->
				<div class="cart-container" class:expanded={cart !== null}>
					{#if cart}
						<div class="cart-content">
							<div class="cart-header">
								<h2>Your Selection</h2>
								<button class="close-cart" on:click={removeFromCart}>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<line x1="18" y1="6" x2="6" y2="18"></line>
										<line x1="6" y1="6" x2="18" y2="18"></line>
									</svg>
								</button>
							</div>
							<div class="cart-summary">
								<div class="cart-item">
									<div class="cart-plan-cont">
										
										<h3>{cart.plan.name}</h3>
										<p>. . .</p>
										<p>{formatPrice(cart.plan.price, cart.plan.currency)}/{cart.plan.interval}</p>
										<button class="close-cart-item" on:click={removeFromCart}>
											<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
												<line x1="18" y1="6" x2="6" y2="18"></line>
												<line x1="6" y1="6" x2="18" y2="18"></line>
											</svg>
										</button>
									</div>
									<p>
										{cart.plan.description}
									</p>
								
								</div>
								{#if cart.offer}
								<h4>Additional Items</h4>
									<div class="cart-offer">
										<div class="cart-plan-cont">
										
											<h3>{cart.offer.off_name}</h3>
										
											<p>{formatDiscount(cart.offer)} - {getItemName(cart.offer.item_id)}</p>
											<button class="close-cart-offer" on:click={removeOfferFromCart}>
												<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
													<line x1="18" y1="6" x2="6" y2="18"></line>
													<line x1="6" y1="6" x2="18" y2="18"></line>
												</svg>
											</button>
										</div>
									</div>
								{/if}
								{#if isAuthenticated}
									<button class="btn btn-primary btn-full" on:click={() => goto('/checkout')}>
										Proceed to Checkout
									</button>
								{:else}
									<div class="auth-required">
										<p>Please sign in to continue with your subscription</p>
										<div class="auth-buttons">
											<button class="btn btn-outline" on:click={() => goto('/login')}>
												Sign In
											</button>
											<button class="btn btn-primary" on:click={() => goto('/register')}>
												Sign Up
											</button>
										</div>
									</div>
								{/if}
							</div>
						</div>
					{:else}
						<div class="cart-placeholder">
							<svg class="cart-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<circle cx="9" cy="21" r="1"></circle>
								<circle cx="20" cy="21" r="1"></circle>
								<path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"></path>
							</svg>
							<p>Select a plan to see your cart</p>
						</div>
					{/if}
				</div>
			</div>
		{/if}
	</div>
</div>

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

<style>

	.btn-bottom {
		position: absolute;
		bottom: 1rem;
		left: 10%;
		width: 80% !important;
		height: 3rem;
	}

	.subscription-page {
		min-height: 100vh;
		max-width: 100vw;
		padding: 0;
		background: var(--bg-dark);
	}

	.container {
		max-width: 1200px;
		margin: 0 auto;
		padding: 0 1rem;
		transition: max-width 1.75s 500ms ease-out;
	}

	.container.expandedLeft {
		max-width: 1800px !important;
	}

	.page-header {
		text-align: center;
		margin-bottom: 5rem;
		position: absolute;
		top: 0;
		left: 0;
		height: 5vh;
		width:  100vw;
		background: var(--bg-secondary);
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
		margin-top: 10vh;
		padding: 2rem;
		background: var(--bg-dark);
		min-width: 1300px;
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
		border-radius: 20px;
		padding: 2rem;
		border: 1px solid var(--border-color);
		transition: all 0.3s ease;
		position: relative;
		overflow: hidden;
		max-width: 350px;
		min-height: 450px;
		border-radius: 50px;
		background: linear-gradient(145deg, #cacaca, #f0f0f0);
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
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.plan-price {
		font-size: 2rem;
		font-weight: 900;
		color: var(--primary-color);
		margin-bottom: 0.5rem;
		text-align: center;
	}

	.interval {
		font-size: 1rem;
		color: var(--text-secondary);
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
		color: var(--text-secondary);
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

	.auth-notice {
		background: var(--info-color);
		color: var(--text-primary);
		padding: 1rem;
		border-radius: 12px;
		margin-top: 2rem;
		font-size: 0.9rem;
		font-weight: 500;
		line-height: 1.5;
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
