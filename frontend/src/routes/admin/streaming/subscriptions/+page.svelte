<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { StreamingSubscriptionService, type SubscriptionPlan, type CreateSubscriptionPlanData } from '$lib/services/streaming-subscriptions';
	import { showToast } from '$lib/toast';
	import { auth } from '$lib/auth';
	
	// Components
	import SubscriptionHeader from './components/SubscriptionHeader.svelte';
	import SubscriptionAccordion from './components/SubscriptionAccordion.svelte';
	import PlanCard from './components/PlanCard.svelte';
	import PlanModal from './components/PlanModal.svelte';

	// State
	let isLoading = true;
	let subscriptionPlans: SubscriptionPlan[] = [];
	let showCreateModal = false;
	let showEditModal = false;
	let showDeleteModal = false;
	let selectedPlan: SubscriptionPlan | null = null;
	let isSubmitting = false;
	let activeAccordion: string | null = 'active';

	// Form data
	let formData: CreateSubscriptionPlanData = {
		name: '',
		description: '',
		short_desc: '',
		price: 0,
		currency: 'USD',
		interval: 'month',
		interval_count: 1,
		features: [],
		is_active: true,
		is_promoted: false,
		promotion_end_date: null,
		sort_order: 0
	};

	// Optimistic updates
	let optimisticUpdates = new Map();

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

	// --- CRUD: READ ---
	async function loadPlans() {
		try {
			isLoading = true;
			console.log('loadPlans: Starting to load plans');
			console.log('loadPlans: Auth state:', $auth);
			console.log('loadPlans: Is authenticated:', $auth.isAuthenticated);
			console.log('loadPlans: User role:', $auth.user?.role);
			
			const allPlans = await StreamingSubscriptionService.getAll();
			console.log('loadPlans: Received plans:', allPlans);
			subscriptionPlans = [...allPlans]; // Immutable update
		} catch (err) {
			console.error('Load plans error', err);
			showToast('Failed to load plans', 'error');
		} finally {
			isLoading = false;
		}
	}

	// --- CRUD: UPDATE ---
	async function editPlan(id: string, updates: Partial<SubscriptionPlan>) {
		try {
			const updatedPlan = await StreamingSubscriptionService.update({ id, ...updates });
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
		promoted: subscriptionPlans.filter(p => p.sub_type === 300 && p.is_active),
		active: subscriptionPlans.filter(p => p.sub_type === 100 && p.is_active),
		inactive: subscriptionPlans.filter(p => !p.is_active),
	};

	// --- DOCUMENTATION ---
	// addPlan: Handles creation of new plans (standard or promo)
	// loadPlans: Loads all plans from backend
	// editPlan: Edits a plan, updates state immutably
	// deletePlan: Soft deletes a plan
	// togglePromotionStatus: Moves plan between promo/standard/inactive
	// groupedPlans: Reactive grouping for UI accordions
	// All state updates are immutable. All errors are logged and surfaced to user.

	// Toggle accordion
	function toggleAccordion(section: string) {
		activeAccordion = activeAccordion === section ? null : section;
	}

	// Optimistic update helpers
	function addOptimisticUpdate(planId: string, updates: Partial<SubscriptionPlan>) {
		optimisticUpdates.set(planId, updates);
	}

	function removeOptimisticUpdate(planId: string) {
		optimisticUpdates.delete(planId);
	}

	$: isOptimisticallyUpdating = (planId: string) => optimisticUpdates.has(planId);

	// Plan actions
	async function createSubscriptionPlan() {
		try {
			isSubmitting = true;
			const newPlan = await StreamingSubscriptionService.create(formData);
			showCreateModal = false;
			resetForm();
			subscriptionPlans = [...subscriptionPlans, newPlan];
			showToast('Subscription plan created successfully', 'success');
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
		const newPromotionStatus = !plan.is_promoted;
		
		console.log(`=== Toggle Promotion Status ===`);
		console.log(`Plan ID: ${planId}`);
		console.log(`Current is_promoted: ${plan.is_promoted}`);
		console.log(`New is_promoted: ${newPromotionStatus}`);
		console.log(`Current sub_type: ${plan.sub_type}`);

		// Optimistic update
		const optimisticUpdate = {
			...plan,
			is_promoted: newPromotionStatus,
			sub_type: newPromotionStatus ? 300 : 100, // Set sub_type based on promotion status
			is_active: newPromotionStatus ? true : false // When ending promotion, set to inactive
		};

		optimisticUpdates.set(planId, optimisticUpdate);
		subscriptionPlans = subscriptionPlans.map(p => 
			p.id === planId ? optimisticUpdate : p
		);

		try {
			const updatedPlan = await StreamingSubscriptionService.togglePromotion(planId, newPromotionStatus);
			
			// Check if we got a valid response
			if (!updatedPlan) {
				throw new Error('No response received from server');
			}
			
			// Update with real data from server
			subscriptionPlans = subscriptionPlans.map(p => 
				p.id === planId ? updatedPlan : p
			);
			
			console.log(`Promotion status updated successfully for plan ${planId}`);
			console.log(`Updated plan:`, updatedPlan);
			
			showToast(`Plan ${newPromotionStatus ? 'promoted' : 'promotion ended'} successfully`, 'success');
		} catch (error) {
			console.error('Error toggling promotion status:', error);
			
			// Revert optimistic update - restore the original plan
			subscriptionPlans = subscriptionPlans.map(p => 
				p.id === planId ? plan : p
			);
			
			showToast(`Failed to ${newPromotionStatus ? 'promote' : 'end promotion for'} plan`, 'error');
		} finally {
			optimisticUpdates.delete(planId);
		}
	}

	// Event handlers
	function handleEdit(event: CustomEvent) {
		selectedPlan = event.detail.plan;
		if (!selectedPlan) return;
		
		formData = {
			name: selectedPlan.name,
			description: selectedPlan.description,
			short_desc: selectedPlan.short_desc,
			price: selectedPlan.price,
			currency: selectedPlan.currency,
			interval: selectedPlan.interval,
			interval_count: selectedPlan.interval_count,
			features: [...selectedPlan.features],
			is_active: selectedPlan.is_active,
			is_promoted: selectedPlan.is_promoted,
			promotion_end_date: selectedPlan.promotion_end_date,
			sort_order: selectedPlan.sort_order
		};
		showEditModal = true;
	}

	function handleDelete(event: CustomEvent) {
		selectedPlan = event.detail.plan;
		showDeleteModal = true;
	}

	function handleToggleStatus(event: CustomEvent) {
		togglePlanStatus(event.detail.plan);
	}

	function handleTogglePromotion(event: CustomEvent) {
		console.log("Here we GO!")
		togglePromotionStatus(event.detail.plan);
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
			features: [],
			is_active: true,
			is_promoted: false,
			promotion_end_date: null,
			sort_order: 0
		};
	}

	// Load data on mount
	onMount(() => {
		loadPlans();
	});
</script>

<div class="subscription-content" transition:fade>
	{#if isLoading}
		<div class="loading-state">
			<div class="loading-spinner"></div>
			<p>Loading subscription plans...</p>
		</div>
	{:else}
		<SubscriptionHeader 
			{subscriptionPlans} 
			onCreateClick={() => showCreateModal = true} 
		/>

		<div class="subscription-accordions">
			<!-- Promoted Plans -->
			<SubscriptionAccordion
				title="Promoted Plans"
				icon="<svg class='w-6 h-6' fill='none' stroke='currentColor' viewBox='0 0 24 24'><path stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z'></path></svg>"
				count={groupedPlans.promoted.length}
				isActive={activeAccordion === 'promoted'}
				plans={groupedPlans.promoted}
				on:toggle={() => toggleAccordion('promoted')}
			>
				{#each groupedPlans.promoted as plan (plan.id)}
					<PlanCard 
						{plan} 
						{isOptimisticallyUpdating}
						on:edit={handleEdit}
						on:toggleStatus={handleToggleStatus}
						on:togglePromotion={handleTogglePromotion}
					/>
				{/each}
			</SubscriptionAccordion>

			<!-- Active Plans -->
			<SubscriptionAccordion
				title="Active Plans"
				icon="<svg class='w-6 h-6' fill='none' stroke='currentColor' viewBox='0 0 24 24'><path stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z'></path></svg>"
				count={groupedPlans.active.length}
				isActive={activeAccordion === 'active'}
				plans={groupedPlans.active}
				on:toggle={() => toggleAccordion('active')}
			>
				{#each groupedPlans.active as plan (plan.id)}
					<PlanCard 
						{plan} 
						{isOptimisticallyUpdating}
						on:edit={handleEdit}
						on:toggleStatus={handleToggleStatus}
						on:togglePromotion={handleTogglePromotion}
					/>
				{/each}
			</SubscriptionAccordion>

			<!-- Inactive Plans -->
			<SubscriptionAccordion
				title="Inactive Plans"
				icon="<svg class='w-6 h-6' fill='none' stroke='currentColor' viewBox='0 0 24 24'><path stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z'></path></svg>"
				count={groupedPlans.inactive.length}
				isActive={activeAccordion === 'inactive'}
				plans={groupedPlans.inactive}
				on:toggle={() => toggleAccordion('inactive')}
			>
				{#each groupedPlans.inactive as plan (plan.id)}
					<PlanCard 
						{plan} 
						{isOptimisticallyUpdating}
						on:edit={handleEdit}
						on:delete={handleDelete}
						on:toggleStatus={handleToggleStatus}
						on:togglePromotion={handleTogglePromotion}
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
	{formData}
	{isSubmitting}
	mode="create"
	on:submit={createSubscriptionPlan}
	on:cancel={() => showCreateModal = false}
/>

<PlanModal
	isOpen={showEditModal}
	title="Edit Subscription Plan"
	{formData}
	{isSubmitting}
	mode="edit"
	on:submit={updateSubscriptionPlan}
	on:cancel={() => showEditModal = false}
/>

<!-- Delete Confirmation Modal -->
{#if showDeleteModal}
	<div class="modal-backdrop" transition:fade={{ duration: 200 }} on:click={() => showDeleteModal = false}>
		<div class="modal-content delete-modal" transition:fly={{ y: 20, duration: 200 }} on:click|stopPropagation>
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
