<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { StreamingSubscriptionService, type SubscriptionPlan, type CreateSubscriptionPlanData } from '$lib/services/streaming-subscriptions';
	import { showToast } from '$lib/toast';

	// Reactive variables
	let isLoading = true;
	let subscriptionPlans: SubscriptionPlan[] = [];
	let error: string | null = null;
	let showCreateModal = false;
	let showEditModal = false;
	let showDeleteModal = false;
	let selectedPlan: SubscriptionPlan | null = null;
	let isSubmitting = false;

	// Form data
	let formData: CreateSubscriptionPlanData = {
		name: '',
		description: '',
		short_desc: '',
		price: 0,
		currency: 'USD',
		interval: 'monthly',
		interval_count: 1,
		features: [],
		is_active: true,
		is_promoted: false,
		promotion_end_date: null,
		sort_order: 0
	};

	// New feature input
	let newFeature = '';

	// Load subscription plans
	async function loadSubscriptionPlans() {
		try {
			isLoading = true;
			error = null;

			const response = await StreamingSubscriptionService.getSubscriptionPlans();
			subscriptionPlans = response.subscription_plans;
		} catch (err: unknown) {
			console.error('Error loading subscription plans:', err);
			error = err instanceof Error ? err.message : 'An unknown error occurred';
		} finally {
			isLoading = false;
		}
	}

	// Create new subscription plan
	async function createSubscriptionPlan() {
		try {
			isSubmitting = true;

			await StreamingSubscriptionService.createSubscriptionPlan(formData);
			
			showCreateModal = false;
			resetForm();
			await loadSubscriptionPlans();
			showToast('Subscription plan created successfully', 'success');
		} catch (err: unknown) {
			console.error('Error creating subscription plan:', err);
			error = err instanceof Error ? err.message : 'An unknown error occurred';
			showToast('Failed to create subscription plan', 'error');
		} finally {
			isSubmitting = false;
		}
	}

	// Update subscription plan
	async function updateSubscriptionPlan() {
		if (!selectedPlan) return;
		
		try {
			isSubmitting = true;

			await StreamingSubscriptionService.updateSubscriptionPlan({
				id: selectedPlan.id,
				...formData
			});
			
			showEditModal = false;
			resetForm();
			await loadSubscriptionPlans();
			showToast('Subscription plan updated successfully', 'success');
		} catch (err: unknown) {
			console.error('Error updating subscription plan:', err);
			error = err instanceof Error ? err.message : 'An unknown error occurred';
			showToast('Failed to update subscription plan', 'error');
		} finally {
			isSubmitting = false;
		}
	}

	// Delete subscription plan
	async function deleteSubscriptionPlan() {
		if (!selectedPlan) return;
		
		try {
			isSubmitting = true;

			await StreamingSubscriptionService.deleteSubscriptionPlan(selectedPlan.id);
			
			showDeleteModal = false;
			selectedPlan = null;
			await loadSubscriptionPlans();
			showToast('Subscription plan deleted successfully', 'success');
		} catch (err: unknown) {
			console.error('Error deleting subscription plan:', err);
			error = err instanceof Error ? err.message : 'An unknown error occurred';
			showToast('Failed to delete subscription plan', 'error');
		} finally {
			isSubmitting = false;
		}
	}

	// Toggle subscription plan status
	async function togglePlanStatus(plan: SubscriptionPlan) {
		try {
			await StreamingSubscriptionService.toggleSubscriptionPlanStatus(plan.id, !plan.is_active);
			await loadSubscriptionPlans();
			showToast(`Plan ${plan.is_active ? 'deactivated' : 'activated'} successfully`, 'success');
		} catch (err: unknown) {
			console.error('Error toggling plan status:', err);
			showToast('Failed to update plan status', 'error');
		}
	}

	// Toggle promotion status
	async function togglePromotionStatus(plan: SubscriptionPlan) {
		try {
			await StreamingSubscriptionService.updatePromotionStatus(plan.id, !plan.is_promoted);
			await loadSubscriptionPlans();
			showToast(`Promotion ${plan.is_promoted ? 'removed' : 'added'} successfully`, 'success');
		} catch (err: unknown) {
			console.error('Error toggling promotion status:', err);
			showToast('Failed to update promotion status', 'error');
		}
	}

	// Edit subscription plan
	function editPlan(plan: SubscriptionPlan) {
		selectedPlan = plan;
		formData = {
			name: plan.name,
			description: plan.description,
			short_desc: plan.short_desc,
			price: plan.price,
			currency: plan.currency,
			interval: plan.interval,
			interval_count: plan.interval_count,
			features: [...plan.features],
			is_active: plan.is_active,
			is_promoted: plan.is_promoted,
			promotion_end_date: plan.promotion_end_date,
			sort_order: plan.sort_order
		};
		showEditModal = true;
	}

	// Delete subscription plan
	function deletePlan(plan: SubscriptionPlan) {
		selectedPlan = plan;
		showDeleteModal = true;
	}

	// Reset form
	function resetForm() {
		formData = {
			name: '',
			description: '',
			short_desc: '',
			price: 0,
			currency: 'USD',
			interval: 'monthly',
			interval_count: 1,
			features: [],
			is_active: true,
			is_promoted: false,
			promotion_end_date: null,
			sort_order: 0
		};
		selectedPlan = null;
		newFeature = '';
	}

	// Add feature
	function addFeature() {
		if (newFeature.trim()) {
			formData.features = [...formData.features, newFeature.trim()];
			newFeature = '';
		}
	}

	// Remove feature
	function removeFeature(index: number) {
		formData.features = formData.features.filter((_, i) => i !== index);
	}

	// Format currency
	function formatCurrency(amount: number, currency: string = 'USD'): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: currency
		}).format(amount);
	}

	// Format interval
	function formatInterval(interval: string, count: number): string {
		if (count === 1) {
			return interval.charAt(0).toUpperCase() + interval.slice(1);
		}
		return `${count} ${interval}s`;
	}

	// Get status badge
	function getStatusBadge(plan: SubscriptionPlan) {
		if (!plan.is_active) {
			return { text: 'Inactive', class: 'bg-gray-100 text-gray-800' };
		}
		if (plan.is_promoted) {
			return { text: 'Promoted', class: 'bg-yellow-100 text-yellow-800' };
		}
		return { text: 'Active', class: 'bg-green-100 text-green-800' };
	}

	onMount(() => {
		loadSubscriptionPlans();
	});
</script>

<svelte:head>
	<title>Subscription Plans - Streaming Admin</title>
</svelte:head>

{#if isLoading}
	<div class="flex items-center justify-center py-12">
		<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
	</div>
{:else if error}
	<div class="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
		<div class="flex items-center">
			<svg class="h-5 w-5 text-red-400 mr-2" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
				<line x1="12" y1="9" x2="12" y2="13"/>
				<line x1="12" y1="17" x2="12.01" y2="17"/>
			</svg>
			<p class="text-red-800">{error}</p>
		</div>
	</div>
{:else}
	<div class="space-y-6" in:fade={{ duration: 300 }}>
		<!-- Header -->
		<div class="flex justify-between items-center">
			<div>
				<h1 class="text-2xl font-bold text-gray-900">Subscription Plans</h1>
				<p class="text-gray-600">Manage subscription plans and promotions</p>
			</div>
			<button
				class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors flex items-center space-x-2"
				on:click={() => showCreateModal = true}
			>
				<svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<line x1="12" y1="5" x2="12" y2="19"/>
					<line x1="5" y1="12" x2="19" y2="12"/>
				</svg>
				<span>Create Plan</span>
			</button>
		</div>

		<!-- Plans Grid -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{#each subscriptionPlans as plan (plan.id)}
				<div class="bg-white border border-gray-200 rounded-lg p-6 hover:shadow-md transition-shadow" in:fly={{ y: 20, duration: 300 }}>
					<!-- Plan Header -->
					<div class="flex justify-between items-start mb-4">
						<div class="flex-1">
							<h3 class="text-lg font-semibold text-gray-900">{plan.name}</h3>
							{#if plan.short_desc}
								<p class="text-sm text-gray-600">{plan.short_desc}</p>
							{/if}
						</div>
						<div class="flex items-center space-x-2">
							{#if plan.is_promoted}
								<svg class="h-5 w-5 text-yellow-500" viewBox="0 0 24 24" fill="currentColor">
									<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
								</svg>
							{/if}
							<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getStatusBadge(plan).class}">
								{getStatusBadge(plan).text}
							</span>
						</div>
					</div>

					<!-- Price -->
					<div class="mb-4">
						<div class="text-2xl font-bold text-gray-900">
							{formatCurrency(plan.price, plan.currency)}
						</div>
						<div class="text-sm text-gray-600">
							per {formatInterval(plan.interval, plan.interval_count)}
						</div>
					</div>

					<!-- Description -->
					<p class="text-gray-700 text-sm mb-4 line-clamp-3">{plan.description}</p>

					<!-- Features -->
					{#if plan.features && plan.features.length > 0}
						<div class="mb-4">
							<ul class="space-y-1">
								{#each plan.features.slice(0, 3) as feature}
									<li class="flex items-center text-sm text-gray-600">
										<svg class="h-4 w-4 text-green-500 mr-2 flex-shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<polyline points="20,6 9,17 4,12"/>
										</svg>
										{feature}
									</li>
								{/each}
								{#if plan.features.length > 3}
									<li class="text-sm text-gray-500">
										+{plan.features.length - 3} more features
									</li>
								{/if}
							</ul>
						</div>
					{/if}

					<!-- Actions -->
					<div class="flex items-center space-x-2 pt-4 border-t border-gray-200">
						<button
							class="flex-1 bg-gray-100 text-gray-700 px-3 py-2 rounded-md hover:bg-gray-200 transition-colors flex items-center justify-center space-x-1"
							on:click={() => editPlan(plan)}
						>
							<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
								<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
							</svg>
							<span>Edit</span>
						</button>
						<button
							class="bg-red-100 text-red-700 px-3 py-2 rounded-md hover:bg-red-200 transition-colors flex items-center justify-center space-x-1"
							on:click={() => deletePlan(plan)}
						>
							<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polyline points="3,6 5,6 21,6"/>
								<path d="M19,6v14a2,2,0,0,1-2,2H7a2,2,0,0,1-2-2V6m3,0V4a2,2,0,0,1,2-2h4a2,2,0,0,1,2,2V6"/>
							</svg>
							<span>Delete</span>
						</button>
					</div>
				</div>
			{/each}
		</div>

		{#if subscriptionPlans.length === 0}
			<div class="text-center py-12">
				<svg class="h-12 w-12 text-gray-400 mx-auto mb-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M12 1v22M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/>
				</svg>
				<h3 class="text-lg font-medium text-gray-900 mb-2">No subscription plans</h3>
				<p class="text-gray-600 mb-4">Get started by creating your first subscription plan.</p>
				<button
					class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
					on:click={() => showCreateModal = true}
				>
					Create Plan
				</button>
			</div>
		{/if}
	</div>
{/if}

<!-- Create Modal -->
{#if showCreateModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" in:fade={{ duration: 200 }}>
		<div class="bg-white rounded-lg p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto" in:fly={{ y: 20, duration: 200 }}>
			<div class="flex justify-between items-center mb-6">
				<h2 class="text-xl font-semibold text-gray-900">Create Subscription Plan</h2>
				<button
					class="text-gray-400 hover:text-gray-600"
					on:click={() => showCreateModal = false}
				>
					<svg class="h-6 w-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"/>
						<line x1="6" y1="6" x2="18" y2="18"/>
					</svg>
				</button>
			</div>

			<form on:submit|preventDefault={createSubscriptionPlan} class="space-y-6">
				<!-- Basic Information -->
				<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
					<div>
						<label for="name" class="block text-sm font-medium text-gray-700 mb-2">Plan Name</label>
						<input
							id="name"
							type="text"
							bind:value={formData.name}
							class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
							required
						/>
					</div>
					<div>
						<label for="short_desc" class="block text-sm font-medium text-gray-700 mb-2">Short Description</label>
						<input
							id="short_desc"
							type="text"
							bind:value={formData.short_desc}
							class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						/>
					</div>
				</div>

				<div>
					<label for="description" class="block text-sm font-medium text-gray-700 mb-2">Description</label>
					<textarea
						id="description"
						bind:value={formData.description}
						rows="3"
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						required
					></textarea>
				</div>

				<!-- Pricing -->
				<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
					<div>
						<label for="price" class="block text-sm font-medium text-gray-700 mb-2">Price</label>
						<input
							id="price"
							type="number"
							step="0.01"
							min="0"
							bind:value={formData.price}
							class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
							required
						/>
					</div>
					<div>
						<label for="currency" class="block text-sm font-medium text-gray-700 mb-2">Currency</label>
						<select
							id="currency"
							bind:value={formData.currency}
							class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						>
							<option value="USD">USD</option>
							<option value="EUR">EUR</option>
							<option value="GBP">GBP</option>
						</select>
					</div>
					<div>
						<label for="interval" class="block text-sm font-medium text-gray-700 mb-2">Billing Interval</label>
						<select
							id="interval"
							bind:value={formData.interval}
							class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						>
							<option value="monthly">Monthly</option>
							<option value="annual">Annual</option>
							<option value="weekly">Weekly</option>
						</select>
					</div>
				</div>

				<!-- Features -->
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-2">Features</label>
					<div class="space-y-2">
						{#each formData.features as feature, index}
							<div class="flex items-center space-x-2">
								<span class="flex-1 text-sm text-gray-700">{feature}</span>
								<button
									type="button"
									class="text-red-600 hover:text-red-800"
									on:click={() => removeFeature(index)}
								>
									<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<line x1="18" y1="6" x2="6" y2="18"/>
										<line x1="6" y1="6" x2="18" y2="18"/>
									</svg>
								</button>
							</div>
						{/each}
						<div class="flex space-x-2">
							<input
								type="text"
								bind:value={newFeature}
								placeholder="Add a feature..."
								class="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
								on:keydown={(e) => e.key === 'Enter' && (e.preventDefault(), addFeature())}
							/>
							<button
								type="button"
								class="bg-blue-600 text-white px-3 py-2 rounded-md hover:bg-blue-700 transition-colors"
								on:click={addFeature}
							>
								Add
							</button>
						</div>
					</div>
				</div>

				<!-- Settings -->
				<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
					<div class="space-y-4">
						<div class="flex items-center">
							<input
								id="is_active"
								type="checkbox"
								bind:checked={formData.is_active}
								class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
							/>
							<label for="is_active" class="ml-2 text-sm text-gray-700">Active</label>
						</div>
						<div class="flex items-center">
							<input
								id="is_promoted"
								type="checkbox"
								bind:checked={formData.is_promoted}
								class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
							/>
							<label for="is_promoted" class="ml-2 text-sm text-gray-700">Promoted</label>
						</div>
					</div>
					<div>
						<label for="sort_order" class="block text-sm font-medium text-gray-700 mb-2">Sort Order</label>
						<input
							id="sort_order"
							type="number"
							bind:value={formData.sort_order}
							class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						/>
					</div>
				</div>

				<!-- Actions -->
				<div class="flex justify-end space-x-3 pt-6 border-t border-gray-200">
					<button
						type="button"
						class="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
						on:click={() => showCreateModal = false}
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={isSubmitting}
						class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors disabled:opacity-50"
					>
						{isSubmitting ? 'Creating...' : 'Create Plan'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Edit Modal -->
{#if showEditModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" in:fade={{ duration: 200 }}>
		<div class="bg-white rounded-lg p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto" in:fly={{ y: 20, duration: 200 }}>
			<div class="flex justify-between items-center mb-6">
				<h2 class="text-xl font-semibold text-gray-900">Edit Subscription Plan</h2>
				<button
					class="text-gray-400 hover:text-gray-600"
					on:click={() => showEditModal = false}
				>
					<svg class="h-6 w-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"/>
						<line x1="6" y1="6" x2="18" y2="18"/>
					</svg>
				</button>
			</div>

			<form on:submit|preventDefault={updateSubscriptionPlan} class="space-y-6">
				<!-- Same form fields as create modal -->
				<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
					<div>
						<label for="edit-name" class="block text-sm font-medium text-gray-700 mb-2">Plan Name</label>
						<input
							id="edit-name"
							type="text"
							bind:value={formData.name}
							class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
							required
						/>
					</div>
					<div>
						<label for="edit-short_desc" class="block text-sm font-medium text-gray-700 mb-2">Short Description</label>
						<input
							id="edit-short_desc"
							type="text"
							bind:value={formData.short_desc}
							class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						/>
					</div>
				</div>

				<div>
					<label for="edit-description" class="block text-sm font-medium text-gray-700 mb-2">Description</label>
					<textarea
						id="edit-description"
						bind:value={formData.description}
						rows="3"
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						required
					></textarea>
				</div>

				<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
					<div>
						<label for="edit-price" class="block text-sm font-medium text-gray-700 mb-2">Price</label>
						<input
							id="edit-price"
							type="number"
							step="0.01"
							min="0"
							bind:value={formData.price}
							class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
							required
						/>
					</div>
					<div>
						<label for="edit-currency" class="block text-sm font-medium text-gray-700 mb-2">Currency</label>
						<select
							id="edit-currency"
							bind:value={formData.currency}
							class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						>
							<option value="USD">USD</option>
							<option value="EUR">EUR</option>
							<option value="GBP">GBP</option>
						</select>
					</div>
					<div>
						<label for="edit-interval" class="block text-sm font-medium text-gray-700 mb-2">Billing Interval</label>
						<select
							id="edit-interval"
							bind:value={formData.interval}
							class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						>
							<option value="monthly">Monthly</option>
							<option value="annual">Annual</option>
							<option value="weekly">Weekly</option>
						</select>
					</div>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-2">Features</label>
					<div class="space-y-2">
						{#each formData.features as feature, index}
							<div class="flex items-center space-x-2">
								<span class="flex-1 text-sm text-gray-700">{feature}</span>
								<button
									type="button"
									class="text-red-600 hover:text-red-800"
									on:click={() => removeFeature(index)}
								>
									<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<line x1="18" y1="6" x2="6" y2="18"/>
										<line x1="6" y1="6" x2="18" y2="18"/>
									</svg>
								</button>
							</div>
						{/each}
						<div class="flex space-x-2">
							<input
								type="text"
								bind:value={newFeature}
								placeholder="Add a feature..."
								class="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
								on:keydown={(e) => e.key === 'Enter' && (e.preventDefault(), addFeature())}
							/>
							<button
								type="button"
								class="bg-blue-600 text-white px-3 py-2 rounded-md hover:bg-blue-700 transition-colors"
								on:click={addFeature}
							>
								Add
							</button>
						</div>
					</div>
				</div>

				<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
					<div class="space-y-4">
						<div class="flex items-center">
							<input
								id="edit-is_active"
								type="checkbox"
								bind:checked={formData.is_active}
								class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
							/>
							<label for="edit-is_active" class="ml-2 text-sm text-gray-700">Active</label>
						</div>
						<div class="flex items-center">
							<input
								id="edit-is_promoted"
								type="checkbox"
								bind:checked={formData.is_promoted}
								class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
							/>
							<label for="edit-is_promoted" class="ml-2 text-sm text-gray-700">Promoted</label>
						</div>
					</div>
					<div>
						<label for="edit-sort_order" class="block text-sm font-medium text-gray-700 mb-2">Sort Order</label>
						<input
							id="edit-sort_order"
							type="number"
							bind:value={formData.sort_order}
							class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						/>
					</div>
				</div>

				<div class="flex justify-end space-x-3 pt-6 border-t border-gray-200">
					<button
						type="button"
						class="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
						on:click={() => showEditModal = false}
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={isSubmitting}
						class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors disabled:opacity-50"
					>
						{isSubmitting ? 'Updating...' : 'Update Plan'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Delete Modal -->
{#if showDeleteModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" in:fade={{ duration: 200 }}>
		<div class="bg-white rounded-lg p-6 w-full max-w-md" in:fly={{ y: 20, duration: 200 }}>
			<div class="flex items-center mb-4">
				<svg class="h-6 w-6 text-red-600 mr-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
					<line x1="12" y1="9" x2="12" y2="13"/>
					<line x1="12" y1="17" x2="12.01" y2="17"/>
				</svg>
				<h2 class="text-xl font-semibold text-gray-900">Delete Subscription Plan</h2>
			</div>
			
			<p class="text-gray-600 mb-6">
				Are you sure you want to delete "{selectedPlan?.name}"? This action will mark the plan as deleted but preserve all data for historical records.
			</p>

			<div class="flex justify-end space-x-3">
				<button
					class="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
					on:click={() => showDeleteModal = false}
				>
					Cancel
				</button>
				<button
					class="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 transition-colors disabled:opacity-50"
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
	/* Glass morphism styling for consistency */
	.bg-white {
		background: var(--bg-glass, rgba(255, 255, 255, 0.1)) !important;
		backdrop-filter: blur(20px);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.bg-gray-50 {
		background: var(--bg-glass-dark, rgba(255, 255, 255, 0.05)) !important;
		backdrop-filter: blur(20px);
	}

	.border-gray-200 {
		border-color: rgba(255, 255, 255, 0.1) !important;
	}

	.border-gray-300 {
		border-color: rgba(255, 255, 255, 0.2) !important;
	}

	/* Button styling */
	.bg-blue-600 {
		background: var(--primary-gradient, linear-gradient(135deg, #3b82f6, #1d4ed8)) !important;
	}

	.bg-blue-600:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 20px rgba(59, 130, 246, 0.3);
	}

	.bg-red-600 {
		background: var(--danger-gradient, linear-gradient(135deg, #ef4444, #dc2626)) !important;
	}

	.bg-red-600:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 20px rgba(239, 68, 68, 0.3);
	}

	.bg-gray-100 {
		background: var(--bg-glass, rgba(255, 255, 255, 0.1)) !important;
		backdrop-filter: blur(20px);
	}

	.bg-gray-100:hover {
		background: rgba(255, 255, 255, 0.2) !important;
		transform: translateY(-2px);
	}

	/* SVG styling */
	svg {
		color: inherit;
	}

	.text-red-600 svg {
		color: #ef4444;
	}

	.text-green-600 svg {
		color: #10b981;
	}

	.text-yellow-600 svg {
		color: #f59e0b;
	}

	.text-blue-600 svg {
		color: #3b82f6;
	}

	/* Status badge styling */
	.status-badge {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		padding: 0.25rem 0.75rem;
		border-radius: 20px;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.status-badge.active {
		background: rgba(16, 185, 129, 0.2);
		color: #10b981;
	}

	.status-badge.inactive {
		background: rgba(107, 114, 128, 0.2);
		color: #6b7280;
	}

	.status-badge.promoted {
		background: rgba(245, 158, 11, 0.2);
		color: #f59e0b;
	}

	/* Card styling */
	.card {
		background: var(--bg-glass, rgba(255, 255, 255, 0.1));
		backdrop-filter: blur(20px);
		border-radius: 15px;
		border: 1px solid rgba(255, 255, 255, 0.1);
		transition: all 0.3s ease;
	}

	.card:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
	}

	/* Form styling */
	input, select, textarea {
		background: var(--bg-primary, rgba(255, 255, 255, 0.05)) !important;
		border-color: rgba(255, 255, 255, 0.2) !important;
		color: var(--text-primary, #ffffff) !important;
	}

	input:focus, select:focus, textarea:focus {
		border-color: var(--primary-color, #3b82f6) !important;
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1) !important;
	}

	/* Text colors */
	.text-gray-700 {
		color: var(--text-primary, #ffffff) !important;
	}

	.text-gray-600 {
		color: var(--text-secondary, #d1d5db) !important;
	}

	.text-gray-900 {
		color: var(--text-primary, #ffffff) !important;
	}

	/* Line clamp utility */
	.line-clamp-3 {
		display: -webkit-box;
		-webkit-line-clamp: 3;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	/* Responsive design */
	@media (max-width: 768px) {
		.card {
			margin: 0.5rem;
		}
	}
</style> 
