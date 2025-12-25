<script lang="ts">
	// @ts-ignore
	import { onMount } from 'svelte';
	import { StreamingSubscriptionService, type SubscriptionPlan, type CreateSubscriptionPlanData } from '$lib/services/streaming-subscriptions';
	import { SubscriptionOfferService, type SubscriptionOffer, type CreateSubscriptionOfferData, type UpdateSubscriptionOfferData } from '$lib/services/subscription-offers';
	import { showToast } from '$lib/toast';
	import { auth } from '$lib/auth';
	import { isoToDateInput } from '$lib/utils/date';
	
	// Components
	import SubscriptionHeader from './components/SubscriptionHeader.svelte';
	import OffersHeader from './components/OffersHeader.svelte';
	import SubscriptionAccordion from './components/SubscriptionAccordion.svelte';
	import PlanCard from './components/PlanCard.svelte';
	import PlanModal from './components/PlanModal.svelte';
	import PlanModalWithStripe from './components/PlanModalWithStripe.svelte'; // Add Stripe-enabled modal
	import PlanDetailsModal from './components/PlanDetailsModal.svelte';
	import OfferModal from './components/OfferModal.svelte';
	import OfferCard from './components/OfferCard.svelte';
	import OfferDetailsModal from './components/OfferDetailsModal.svelte';
	import StripeIntegrationStatus from './components/StripeIntegrationStatus.svelte'; // Add Stripe status component
	import StripeImportModal from './components/StripeImportModal.svelte';
	import StripeProductsAccordion from './components/StripeProductsAccordion.svelte';

	// State
	let isLoading = true;
	let subscriptionPlans: SubscriptionPlan[] = [];
	let subscriptionOffers: SubscriptionOffer[] = [];
	let showCreateModal = false;
	let showCreateStripeModal = false; // Add Stripe modal state
	let showEditModal = false;
	let showDeleteModal = false;
	let selectedPlan: SubscriptionPlan | null = null;
	let selectedOffer: SubscriptionOffer | null = null;
	let isSubmitting = false;
	let activeAccordion: string | null = 'active';

	// Modal state
	let showDetailsModal = false;
	let detailsPlan: SubscriptionPlan | null = null;
	let detailsOffer: SubscriptionOffer | null = null;

	// Offer modal state
	let showCreateOfferModal = false;
	let showEditOfferModal = false;
	let showOfferDetailsModal = false;
	
	// Stripe import modal state
	let showStripeImportModal = false;

	// Initialize form data
	let formData: CreateSubscriptionPlanData = {
		name: '',
		description: '',
		short_desc: '',
		price: 0,
		currency: 'USD',
		interval: 'month',
		interval_count: 1,
		stripe_price_id: '',
		stripe_product_id: null,
		features: [],
		is_active: true,
		sub_type: 'stnd',
		promotion_start_date: null,
		promotion_end_date: null
	};

	// Initialize offer form data
	let offerFormData: CreateSubscriptionOfferData = {
		plan_id: 1,
		item_id: undefined,
		off_discount_type: 'percentage',
		off_discount_value: 10,
		offer_start_date: undefined,
		off_end_date: undefined,
		is_active: true,
		off_description: undefined,
		off_name: '',
		off_code: undefined,
		off_max_uses: 100,
		off_current_uses: 0,
		off_terms_conditions: undefined,
		off_target: undefined,
		off_priority: 1,
		off_auto_apply: false
	};

	// Optimistic updates
	let optimisticUpdates = new Map<string | number, any>();

	// --- CRUD: CREATE ---
	async function addPlan(formData: CreateSubscriptionPlanData) {
		try {
			const newPlan = await StreamingSubscriptionService.create(formData);
			subscriptionPlans = [...subscriptionPlans, newPlan]; // Immutable update
			showToast('Plan created successfully', 'success');
		} catch (err) {
			console.error('Add plan error', err);
			showToast('Failed to create plan', 'error');
		}
	}

	// --- CRUD: CREATE OFFER ---
	async function addOffer(offerFormData: CreateSubscriptionOfferData) {
		try {
			const newOffer = await SubscriptionOfferService.create(offerFormData);
			subscriptionOffers = [...subscriptionOffers, newOffer]; // Immutable update
			showToast('Offer created successfully', 'success');
		} catch (err) {
			console.error('Add offer error', err);
			showToast('Failed to create offer', 'error');
		}
	}

	// --- CRUD: READ ---
	async function loadPlans() {
		try {
			isLoading = true;
			console.log('loadPlans: Starting to load plans');
			console.log('loadPlans: Auth state:', $auth);
			console.log('loadPlans: Is authenticated:', ($auth as any).isAuthenticated);
			console.log('loadPlans: User role:', ($auth as any).user?.role);
			
			// Load subscription plans (Stripe products are now imported as real plans)
			const allPlans = await StreamingSubscriptionService.getAll();
			console.log('loadPlans: Received subscription plans:', allPlans);
			subscriptionPlans = [...allPlans]; // Immutable update
			
			// Check for expired promotions and deactivate them
			await checkAndDeactivateExpiredPromotions();
		} catch (err) {
			console.error('Load plans error', err);
			showToast('Failed to load plans', 'error');
		} finally {
			isLoading = false;
		}
	}

	// --- CRUD: READ OFFERS ---
	async function loadOffers() {
		try {
			console.log('loadOffers: Starting to load offers');
			
			const allOffers = await SubscriptionOfferService.getAll();
			console.log('loadOffers: Received offers:', allOffers);
			subscriptionOffers = [...allOffers]; // Immutable update
		} catch (err) {
			console.error('Load offers error', err);
			showToast('Failed to load offers', 'error');
		}
	}

	// Check for expired promotions and deactivate them
	async function checkAndDeactivateExpiredPromotions() {
		const now = new Date();
		const expiredPromotions = subscriptionPlans.filter(plan => {
			if (plan.sub_type !== 'prmo' || !plan.is_active) return false;
			
			const endDate = plan.promotion_end_date ? new Date(plan.promotion_end_date) : null;
			return endDate && now > endDate;
		});

		if (expiredPromotions.length > 0) {
			console.log(`Found ${expiredPromotions.length} expired promotions, deactivating...`);
			
			for (const plan of expiredPromotions) {
				try {
					// Optimistic update
					addOptimisticUpdate(plan.id, { is_active: false });
					subscriptionPlans = subscriptionPlans.map(p => 
						p.id === plan.id ? { ...p, is_active: false } : p
					);

					// Call backend to deactivate
					const updatedPlan = await StreamingSubscriptionService.toggleStatus(plan.id, false);
					removeOptimisticUpdate(plan.id);
					
					// Update with server response
					subscriptionPlans = subscriptionPlans.map(p => 
						p.id === plan.id ? updatedPlan : p
					);
					
					console.log(`Deactivated expired promotion: ${plan.name}`);
				} catch (err) {
					console.error(`Failed to deactivate expired promotion ${plan.name}:`, err);
					removeOptimisticUpdate(plan.id);
					// Revert optimistic update
					subscriptionPlans = subscriptionPlans.map(p => 
						p.id === plan.id ? plan : p
					);
				}
			}
			
			if (expiredPromotions.length > 0) {
				showToast(`${expiredPromotions.length} expired promotion(s) deactivated`, 'info');
			}
		}
	}

	// --- CRUD: UPDATE ---
	async function editPlan(id: string, updates: Partial<SubscriptionPlan>) {
		try {
			// Ensure sub_type is properly typed if it exists in updates
			const typedUpdates: any = { ...updates };
			if (typedUpdates.sub_type && typeof typedUpdates.sub_type === 'string') {
				typedUpdates.sub_type = typedUpdates.sub_type as 'stnd' | 'prmo';
			}
			
			const updatedPlan = await StreamingSubscriptionService.update({ id, ...typedUpdates });
			subscriptionPlans = subscriptionPlans.map(p => p.id === id ? updatedPlan : p); // Immutable update
			showToast('Plan updated successfully', 'success');
		} catch (err) {
			console.error('Edit plan error', err);
			showToast('Failed to update plan', 'error');
		}
	}

	// --- CRUD: DELETE (Soft Delete) ---
	async function deletePlan(id: string) {
		try {
			await StreamingSubscriptionService.delete(id);
			subscriptionPlans = subscriptionPlans.filter(p => p.id !== id); // Immutable update
			showToast('Plan deleted successfully', 'success');
		} catch (err) {
			console.error('Delete plan error', err);
			showToast('Failed to delete plan', 'error');
		}
	}

	// --- GROUPING LOGIC ---
	// Group plans by sub_type and is_active for display
	$: groupedPlans = {
		promoted: subscriptionPlans.filter(p => p && p.sub_type === "prmo" && p.is_active),
		active: subscriptionPlans.filter(p => p && p.sub_type === "stnd" && p.is_active),
		inactive: subscriptionPlans.filter(p => p && !p.is_active),
	};

	// Group offers by is_active for display
	$: groupedOffers = {
		active: subscriptionOffers.filter(o => o && o.is_active),
		inactive: subscriptionOffers.filter(o => o && !o.is_active),
	};

	// --- DOCUMENTATION ---
	// addPlan: Handles creation of new plans (standard or promo)
	// loadPlans: Loads all plans from backend
	// editPlan: Edits a plan, updates state immutably
	// deletePlan: Soft deletes a plan
	// togglePromotionStatus: Moves plan between promo/standard/inactive
	// groupedPlans: Reactive grouping for UI accordions
	// addOffer: Handles creation of new offers
	// loadOffers: Loads all offers from backend
	// groupedOffers: Reactive grouping for UI accordions
	// All state updates are immutable. All errors are logged and surfaced to user.

	// Toggle accordion
	function toggleAccordion(section: string) {
		activeAccordion = activeAccordion === section ? null : section;
	}

	// Handle view details
	function handleViewDetails(plan: SubscriptionPlan) {
		detailsPlan = plan;
		showDetailsModal = true;
	}

	// Handle view offer details
	function handleViewOfferDetails(offer: SubscriptionOffer) {
		detailsOffer = offer;
		showOfferDetailsModal = true;
	}

	// Refresh plan data (for reactive updates)
	async function refreshPlanData(planId: string) {
		try {
			const refreshedPlan = await StreamingSubscriptionService.getById(planId);
			subscriptionPlans = subscriptionPlans.map(p => p.id === planId ? refreshedPlan : p);
			
			// Update details modal if it's open for this plan
			if (detailsPlan && detailsPlan.id === planId) {
				detailsPlan = refreshedPlan;
			}
		} catch (err) {
			console.error('Failed to refresh plan data:', err);
		}
	}

	// Refresh offer data (for reactive updates)
	async function refreshOfferData(offerId: number) {
		try {
			const refreshedOffer = await SubscriptionOfferService.getById(offerId.toString());
			subscriptionOffers = subscriptionOffers.map(o => o.id === offerId ? refreshedOffer : o);
			
			// Update details modal if it's open for this offer
			if (detailsOffer && detailsOffer.id === offerId) {
				detailsOffer = refreshedOffer;
			}
		} catch (err) {
			console.error('Failed to refresh offer data:', err);
		}
	}

	// Optimistic update helpers
	function addOptimisticUpdate(id: string | number, updates: any) {
		optimisticUpdates.set(id, updates);
	}

	function removeOptimisticUpdate(id: string | number) {
		optimisticUpdates.delete(id);
	}

	$: isOptimisticallyUpdating = (id: string | number) => optimisticUpdates.has(id);

	// Plan actions
	async function createSubscriptionPlan() {
		try {
			isSubmitting = true;
			const newPlan = await StreamingSubscriptionService.create(formData);
			
			// Only add to array if we got a valid response
			if (newPlan && newPlan.id) {
				showCreateModal = false;
				resetForm();
				subscriptionPlans = [...subscriptionPlans, newPlan];
				showToast('Subscription plan created successfully', 'success');
			} else {
				throw new Error('Invalid response from server');
			}
		} catch (err) {
			console.error('Error creating subscription plan:', err);
			showToast('Failed to create subscription plan', 'error');
		} finally {
			isSubmitting = false;
		}
	}

	async function updateSubscriptionPlan() {
		if (!selectedPlan) return;
		
		try {
			isSubmitting = true;
			const updatedPlan = await StreamingSubscriptionService.update({
				id: selectedPlan.id,
				...formData
			});
			showEditModal = false;
			resetForm();
			subscriptionPlans = subscriptionPlans.map(plan => 
				plan.id === selectedPlan?.id ? updatedPlan : plan
			);
			showToast('Subscription plan updated successfully', 'success');
		} catch (err) {
			console.error('Error updating subscription plan:', err);
			showToast('Failed to update subscription plan', 'error');
		} finally {
			isSubmitting = false;
		}
	}

	async function deleteSubscriptionPlan() {
		if (!selectedPlan) return;
		
		try {
			isSubmitting = true;
			await StreamingSubscriptionService.delete(selectedPlan.id);
			showDeleteModal = false;
			subscriptionPlans = subscriptionPlans.filter(plan => plan.id !== selectedPlan?.id);
			selectedPlan = null;
			showToast('Subscription plan deleted successfully', 'success');
		} catch (err) {
			console.error('Error deleting subscription plan:', err);
			showToast('Failed to delete subscription plan', 'error');
		} finally {
			isSubmitting = false;
		}
	}

	// Toggle plan status with optimistic updates
	async function togglePlanStatus(plan: SubscriptionPlan) {
		const newStatus = !plan.is_active;
		addOptimisticUpdate(plan.id, { is_active: newStatus });
		
		try {
			const updatedPlan = await StreamingSubscriptionService.toggleStatus(plan.id, newStatus);
			removeOptimisticUpdate(plan.id);
			subscriptionPlans = subscriptionPlans.map(p => p?.id === plan.id ? updatedPlan : p);
			
			// Refresh plan data to get updated history
			await refreshPlanData(plan.id);
			
			showToast(`Plan ${newStatus ? 'activated' : 'deactivated'} successfully`, 'success');
		} catch (err) {
			console.error('Error toggling plan status:', err);
			removeOptimisticUpdate(plan.id);
			showToast('Failed to update plan status', 'error');
		}
	}

	// Toggle promotion status with optimistic updates
	async function togglePromotionStatus(plan: SubscriptionPlan) {
		const planId = plan.id;
		const isCurrentlyPromoted = plan.sub_type === "prmo";
		const isCurrentlyActive = plan.is_active;
		
		console.log(`=== Toggle Promotion Status ===`);
		console.log(`Plan ID: ${planId}`);
		console.log(`Current sub_type: ${plan.sub_type}`);
		console.log(`Current is_active: ${isCurrentlyActive}`);
		console.log(`Is currently promoted: ${isCurrentlyPromoted}`);

		// If plan is inactive, just activate it without changing sub_type
		if (!isCurrentlyActive) {
			console.log(`Plan is inactive, activating without changing sub_type`);
			
			// Optimistic update - only change is_active
			const optimisticUpdate = {
				...plan,
				is_active: true
			};

			optimisticUpdates.set(planId, optimisticUpdate);
			subscriptionPlans = subscriptionPlans.map(p => 
				p.id === planId ? optimisticUpdate : p
			);

			try {
				const updatedPlan = await StreamingSubscriptionService.toggleStatus(planId, true);
				
				if (!updatedPlan) {
					throw new Error('No response received from server');
				}
				
				subscriptionPlans = subscriptionPlans.map(p => 
					p.id === planId ? updatedPlan : p
				);
				
				// Refresh plan data to get updated history
				await refreshPlanData(planId);
				
				console.log(`Plan activated successfully:`, updatedPlan);
				showToast('Plan activated successfully', 'success');
			} catch (error) {
				console.error('Error activating plan:', error);
				subscriptionPlans = subscriptionPlans.map(p => 
					p.id === planId ? plan : p
				);
				showToast('Failed to activate plan', 'error');
			} finally {
				optimisticUpdates.delete(planId);
			}
			return;
		}

		// If plan is active, toggle promotion status (change sub_type)
		const newPromotionStatus = !isCurrentlyPromoted;
		console.log(`Plan is active, toggling promotion status to: ${newPromotionStatus}`);

		// Optimistic update - change sub_type and set is_active to false when ending promotion
		const optimisticUpdate = {
			...plan,
			sub_type: newPromotionStatus ? "prmo" : "stnd",
			is_active: newPromotionStatus ? true : false // When ending promotion, set to inactive
		};

		optimisticUpdates.set(planId, optimisticUpdate);
		subscriptionPlans = subscriptionPlans.map(p => 
			p.id === planId ? optimisticUpdate : p
		);

		try {
			const updatedPlan = await StreamingSubscriptionService.togglePromotion(planId, newPromotionStatus);
			
			if (!updatedPlan) {
				throw new Error('No response received from server');
			}
			
			subscriptionPlans = subscriptionPlans.map(p => 
				p.id === planId ? updatedPlan : p
			);
			
			// Refresh plan data to get updated history
			await refreshPlanData(planId);
			
			console.log(`Plan promotion status updated successfully:`, updatedPlan);
			showToast(`Plan ${newPromotionStatus ? 'promoted' : 'unpromoted'} successfully`, 'success');
		} catch (error) {
			console.error('Error updating promotion status:', error);
			subscriptionPlans = subscriptionPlans.map(p => 
				p.id === planId ? plan : p
			);
			showToast('Failed to update promotion status', 'error');
		} finally {
			optimisticUpdates.delete(planId);
		}
	}

	// Event handlers
	function handleEdit(plan: SubscriptionPlan) {
		selectedPlan = plan;
		if (!selectedPlan) return;
		
		// All plans are now real subscription plans (Stripe products are imported)
		formData = {
			name: selectedPlan.name,
			description: selectedPlan.description,
			short_desc: selectedPlan.short_desc,
			price: selectedPlan.price,
			currency: selectedPlan.currency,
			interval: selectedPlan.interval,
			interval_count: selectedPlan.interval_count,
			stripe_price_id: selectedPlan.stripe_price_id || '',
			stripe_product_id: selectedPlan.stripe_product_id,
			features: [...selectedPlan.features],
			is_active: selectedPlan.is_active,
			sub_type: selectedPlan.sub_type as 'stnd' | 'prmo',
			promotion_start_date: selectedPlan.promotion_start_date ? isoToDateInput(selectedPlan.promotion_start_date) : null,
			promotion_end_date: selectedPlan.promotion_end_date ? isoToDateInput(selectedPlan.promotion_end_date) : null
		};
		showEditModal = true;
	}

	function handleDelete(plan: SubscriptionPlan) {
		selectedPlan = plan;
		showDeleteModal = true;
	}

	function handleToggleStatus(plan: SubscriptionPlan) {
		togglePlanStatus(plan);
	}

	function handleTogglePromotion(plan: SubscriptionPlan) {
		console.log("Here we GO!")
		togglePromotionStatus(plan);
	}

	function resetForm() {
		formData = {
			name: '',
			description: '',
			short_desc: '',
			price: 0,
			currency: 'USD',
			interval: 'month',
			interval_count: 1,
			stripe_price_id: '',
			stripe_product_id: null,
			features: [],
			is_active: true,
			sub_type: 'stnd',
			promotion_start_date: null,
			promotion_end_date: null
		};
	}

	// Load data on mount
	onMount(() => {
		loadPlans();
		loadOffers();
	});

	// --- CRUD: UPDATE OFFER ---
	async function editOffer(id: number, updates: Partial<SubscriptionOffer>) {
		try {
			const updatedOffer = await SubscriptionOfferService.update({ id, ...updates });
			subscriptionOffers = subscriptionOffers.map(o => o.id === id ? updatedOffer : o); // Immutable update
			showToast('Offer updated successfully', 'success');
		} catch (err) {
			console.error('Edit offer error', err);
			showToast('Failed to update offer', 'error');
		}
	}

	// --- CRUD: DELETE OFFER ---
	async function deleteOffer(id: number) {
		try {
			await SubscriptionOfferService.delete(id.toString());
			subscriptionOffers = subscriptionOffers.filter(o => o.id !== id); // Immutable update
			showToast('Offer deleted successfully', 'success');
		} catch (err) {
			console.error('Delete offer error', err);
			showToast('Failed to delete offer', 'error');
		}
	}

	// --- OFFER ACTIONS ---
	async function createSubscriptionOffer(formData: CreateSubscriptionOfferData) {
		try {
			isSubmitting = true;
			const newOffer = await SubscriptionOfferService.create(formData);
			
			// Only add to array if we got a valid response
			if (newOffer && newOffer.id) {
				showCreateOfferModal = false;
				subscriptionOffers = [...subscriptionOffers, newOffer];
				showToast('Subscription offer created successfully', 'success');
			} else {
				throw new Error('Invalid response from server');
			}
		} catch (err) {
			console.error('Error creating subscription offer:', err);
			showToast('Failed to create subscription offer', 'error');
		} finally {
			isSubmitting = false;
		}
	}

	async function updateSubscriptionOffer(formData: UpdateSubscriptionOfferData) {
		if (!selectedOffer) return;
		
		try {
			isSubmitting = true;
			const updatedOffer = await SubscriptionOfferService.update(formData);
			showEditOfferModal = false;
			subscriptionOffers = subscriptionOffers.map(offer => 
				offer.id === selectedOffer?.id ? updatedOffer : offer
			);
			showToast('Subscription offer updated successfully', 'success');
		} catch (err) {
			console.error('Error updating subscription offer:', err);
			showToast('Failed to update subscription offer', 'error');
		} finally {
			isSubmitting = false;
		}
	}

	// Handle edit offer
	function handleEditOffer(offer: SubscriptionOffer) {
		selectedOffer = offer;
		showEditOfferModal = true;
	}

	// Handle delete offer
	function handleDeleteOffer(offer: SubscriptionOffer) {
		selectedOffer = offer;
		// You can add a delete confirmation modal here if needed
		deleteOffer(offer.id);
	}

	// Toggle offer status
	async function toggleOfferStatus(offer: SubscriptionOffer) {
		const newStatus = !offer.is_active;
		addOptimisticUpdate(offer.id, { is_active: newStatus });
		
		try {
			const updatedOffer = await SubscriptionOfferService.toggleStatus(offer.id.toString(), newStatus);
			removeOptimisticUpdate(offer.id);
			subscriptionOffers = subscriptionOffers.map(o => o.id === offer.id ? updatedOffer : o);
			showToast(`Offer ${updatedOffer.is_active ? 'activated' : 'deactivated'} successfully`, 'success');
		} catch (err) {
			console.error('Error toggling offer status:', err);
			removeOptimisticUpdate(offer.id);
			showToast('Failed to update offer status', 'error');
		}
	}

	// Wrapper functions for modal compatibility
	function handleOfferSubmit(formData: CreateSubscriptionOfferData | UpdateSubscriptionOfferData) {
		if ('id' in formData) {
			return updateSubscriptionOffer(formData as UpdateSubscriptionOfferData);
		} else {
			return createSubscriptionOffer(formData as CreateSubscriptionOfferData);
		}
	}
</script>

<div class="subscription-content p-0">
	{#if isLoading}
		<div class="loading-state">
			<div class="loading-spinner"></div>
			<p>Loading subscription plans...</p>
		</div>
	{:else}
		

		<div class="subscription-accordions">
			<!-- Plans Section -->
			<div class="section-header">
				<h2 class="section-title">Plans</h2>
				<p class="section-description">Subscription plans for non-subscribed users</p>
			</div>
		<SubscriptionHeader 
			{subscriptionPlans} 
			onCreateClick={() => showCreateModal = true}
			onCreateWithStripeClick={() => showCreateStripeModal = true}
			onSelectStripeClick={() => showStripeImportModal = true}
		/>
			<!-- Promoted Plans -->
			<SubscriptionAccordion
				title="Promoted Plans"
				class_title="promoted-plans"
				icon="<svg class='w-6 h-6' fill='none' stroke='currentColor' viewBox='0 0 24 24'><path stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z'></path></svg>"
				count={groupedPlans.promoted.length}
				isActive={activeAccordion === 'promoted'}
				plans={groupedPlans.promoted}
				onToggle={() => toggleAccordion('promoted')}
			>
				{#each groupedPlans.promoted as plan (plan.id)}
					<PlanCard 
						{plan} 
						{isOptimisticallyUpdating}
						onEdit={handleEdit}
						onToggleStatus={handleToggleStatus}
						onTogglePromotion={handleTogglePromotion}
						onViewDetails={handleViewDetails}
					/>
				{/each}
			</SubscriptionAccordion>

			<!-- Active Plans -->
			<SubscriptionAccordion
				title="Active Plans"
				class_title="active-plans"
				icon="<svg class='w-6 h-6' fill='none' stroke='currentColor' viewBox='0 0 24 24'><path stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z'></path></svg>"
				count={groupedPlans.active.length}
				isActive={activeAccordion === 'active'}
				plans={groupedPlans.active}
				onToggle={() => toggleAccordion('active')}
			>
				{#each groupedPlans.active as plan (plan.id)}
					<PlanCard 
						{plan} 
						{isOptimisticallyUpdating}
						onEdit={handleEdit}
						onToggleStatus={handleToggleStatus}
						onTogglePromotion={handleTogglePromotion}
						onViewDetails={handleViewDetails}
					/>
				{/each}
			</SubscriptionAccordion>

			<!-- Inactive Plans -->
			<SubscriptionAccordion
				title="Inactive Plans"
				class_title="inactive-plans"
				icon="<svg class='w-6 h-6' fill='none' stroke='currentColor' viewBox='0 0 24 24'><path stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z'></path></svg>"
				count={groupedPlans.inactive.length}
				isActive={activeAccordion === 'inactive'}
				plans={groupedPlans.inactive}
				onToggle={() => toggleAccordion('inactive')}
			>
				{#each groupedPlans.inactive as plan (plan.id)}
					<PlanCard 
						{plan} 
						{isOptimisticallyUpdating}
						onEdit={handleEdit}
						onToggleStatus={handleToggleStatus}
						onTogglePromotion={handleTogglePromotion}
						onViewDetails={handleViewDetails}
					/>
				{/each}
			</SubscriptionAccordion>

			<!-- Stripe Products Accordion (nested within Plans section) -->
			<StripeProductsAccordion bind:activeAccordion />

			<!-- Separator -->
			<hr class="section-divider" />

			

			<!-- Offers Section -->
			<div class="section-header">
				<h2 class="section-title">Offers</h2>
				<p class="section-description">Special offers presented to users who have chosen a subscription</p>
			</div>
			
			<!-- Offers Header -->
			<OffersHeader 
				{subscriptionOffers} 
				onCreateClick={() => showCreateOfferModal = true} 
			/>

			<!-- Active Offers -->
			<SubscriptionAccordion
				title="Active Offers"
				class_title="active-offers"
				icon="<svg class='w-6 h-6' fill='none' stroke='currentColor' viewBox='0 0 24 24'><path stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z'></path></svg>"
				count={groupedOffers.active.length}
				isActive={activeAccordion === 'active-offers'}
				plans={groupedOffers.active}
				onToggle={() => toggleAccordion('active-offers')}
			>
				{#each groupedOffers.active as offer (offer.id)}
					<OfferCard 
						{offer}
						{subscriptionPlans}
						{isOptimisticallyUpdating}
						onEdit={handleEditOffer}
						onToggleStatus={toggleOfferStatus}
						onViewDetails={handleViewOfferDetails}
					/>
				{/each}
			</SubscriptionAccordion>

			<!-- Inactive Offers -->
			<SubscriptionAccordion
				title="Inactive Offers"
				class_title="inactive-offers"
				icon="<svg class='w-6 h-6' fill='none' stroke='currentColor' viewBox='0 0 24 24'><path stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z'></path></svg>"
				count={groupedOffers.inactive.length}
				isActive={activeAccordion === 'inactive-offers'}
				plans={groupedOffers.inactive}
				onToggle={() => toggleAccordion('inactive-offers')}
			>
				{#each groupedOffers.inactive as offer (offer.id)}
					<OfferCard 
						{offer}
						{subscriptionPlans}
						{isOptimisticallyUpdating}
						onEdit={handleEditOffer}
						onToggleStatus={toggleOfferStatus}
						onViewDetails={handleViewOfferDetails}
					/>
				{/each}
			</SubscriptionAccordion>
		</div>
	{/if}
</div>

<!-- Modals -->
<PlanModal
	isOpen={showCreateModal}
	title="Create Subscription Plan"
	plan={formData}
	{isSubmitting}
	mode="create"
	onSave={createSubscriptionPlan}
	onCancel={() => showCreateModal = false}
/>

<PlanModal
	isOpen={showEditModal}
	title="Edit Subscription Plan"
	plan={formData}
	{isSubmitting}
	mode="edit"
	onSave={updateSubscriptionPlan}
	onCancel={() => showEditModal = false}
/>

<!-- Stripe-enabled Plan Modal -->
<PlanModalWithStripe
	show={showCreateStripeModal}
	on:planCreated={(event) => {
		subscriptionPlans = [...subscriptionPlans, event.detail];
		showCreateStripeModal = false;
	}}
/>

<!-- Stripe Import Modal -->
<StripeImportModal
	bind:show={showStripeImportModal}
	onClose={() => showStripeImportModal = false}
	onImportComplete={async () => {
		// Reload plans to show newly imported Stripe products
		await loadPlans();
		showToast('Stripe products imported as subscription plans', 'success');
	}}
/>

<PlanDetailsModal
	bind:isOpen={showDetailsModal}
	plan={detailsPlan}
/>

<!-- Offer Modals -->
<OfferModal
	isOpen={showCreateOfferModal}
	mode="create"
	offer={null}
	{subscriptionPlans}
	{isSubmitting}
	onSubmit={handleOfferSubmit}
	onCancel={() => showCreateOfferModal = false}
/>

<OfferModal
	isOpen={showEditOfferModal}
	mode="edit"
	offer={selectedOffer}
	{subscriptionPlans}
	{isSubmitting}
	onSubmit={handleOfferSubmit}
	onCancel={() => showEditOfferModal = false}
/>

<OfferDetailsModal
	bind:isOpen={showOfferDetailsModal}
	offer={detailsOffer}
	{subscriptionPlans}
/>

<!-- Delete Confirmation Modal -->
{#if showDeleteModal}
	<div class="modal-backdrop" on:click={() => showDeleteModal = false}>
		<div class="modal-content delete-modal" on:click|stopPropagation>
			<div class="modal-header">
				<div class="modal-title-with-icon">
					<svg class="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"></path>
					</svg>
					<h2 class="modal-title">Delete Subscription Plan</h2>
				</div>
				<button type="button" class="modal-close" on:click={() => showDeleteModal = false} aria-label="Close">
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
					</svg>
				</button>
			</div>
			
			<div class="modal-body">
				<p class="delete-message">
					Are you sure you want to delete "<strong>{selectedPlan?.name}</strong>"? This action will mark the plan as deleted but preserve all data for historical records.
				</p>
			</div>

			<div class="modal-actions">
				<button
					class="btn btn-secondary"
					on:click={() => showDeleteModal = false}
				>
					Cancel
				</button>
				<button
					class="btn btn-danger"
					disabled={isSubmitting}
					on:click={deleteSubscriptionPlan}
				>
					{isSubmitting ? 'Deleting...' : 'Delete Plan'}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.subscription-content {
		margin: 0 auto;
	}

	.loading-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 400px;
		color: #6b7280;
	}

	.loading-state p {
		margin-top: 1rem;
		font-size: 1.125rem;
	}

	.loading-spinner {
		width: 40px;
		height: 40px;
		border: 3px solid #e5e7eb;
		border-top: 3px solid #2563eb;
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.subscription-accordions {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.section-header {
		margin-bottom: 1rem;
	}

	.section-title {
		font-size: 1.5rem;
		font-weight: 700;
		color: #111827;
		margin: 0 0 0.5rem 0;
	}

	.section-description {
		font-size: 0.875rem;
		color: #6b7280;
		margin: 0;
	}

	.section-divider {
		border: none;
		height: 1px;
		background: #e5e7eb;
		margin: 2rem 0;
	}

	/* Delete Modal Styles */
	.modal-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 50;
		padding: 1rem;
	}

	.modal-content {
		background: white;
		border-radius: 0.5rem;
		border: 1px solid #e5e7eb;
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.25);
		width: 100%;
		max-width: 500px;
		max-height: 90vh;
		overflow-y: auto;
	}

	.delete-modal {
		max-width: 450px;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1.5rem 1.5rem 0 1.5rem;
		border-bottom: 1px solid #e5e7eb;
		margin-bottom: 1.5rem;
	}

	.modal-title-with-icon {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.modal-title {
		font-size: 1.25rem;
		font-weight: 600;
		color: #111827;
		margin: 0;
	}

	.modal-close {
		background: none;
		border: none;
		color: #6b7280;
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 0.375rem;
		transition: all 0.2s ease;
	}

	.modal-close:hover {
		background: #f3f4f6;
		color: #374151;
	}

	.modal-body {
		padding: 0 1.5rem;
	}

	.delete-message {
		color: #374151;
		font-size: 0.875rem;
		line-height: 1.5;
		margin: 0;
	}

	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
		padding: 1.5rem;
		border-top: 1px solid #e5e7eb;
		margin-top: 1.5rem;
	}

	.btn {
		padding: 0.5rem 1rem;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		border: none;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
	}

	.btn:hover {
		transform: translateY(-1px);
	}

	.btn:focus {
		outline: 2px solid #2563eb;
		outline-offset: 2px;
	}

	.btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
		transform: none !important;
	}

	.btn-secondary {
		background: #f3f4f6;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-secondary:hover {
		background: #e5e7eb;
		border-color: #9ca3af;
	}

	.btn-danger {
		background: #dc2626;
		color: white;
	}

	.btn-danger:hover {
		background: #b91c1c;
		box-shadow: 0 4px 12px rgba(220, 38, 38, 0.3);
	}

	@media (max-width: 768px) {
		.modal-content {
			margin: 0.5rem;
			max-height: calc(100vh - 1rem);
		}

		.modal-actions {
			flex-direction: column;
		}

		.modal-actions .btn {
			width: 100%;
		}
	}
</style>
