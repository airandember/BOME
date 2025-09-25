<script lang="ts">
	import { formatDateForDisplay, dateInputToIso, isoToDateInput, getTodayDateInput, getDateInputFromToday } from '$lib/utils/date';
	import type { CreateSubscriptionPlanData, StripeProduct } from '$lib/services/streaming-subscriptions';
	import StripeProductModal from './StripeProductModal.svelte';

	export let isOpen: boolean = false;
	export let plan: any = null;
	export let onSave: (data: CreateSubscriptionPlanData) => void = () => {};
	export let onCancel: () => void = () => {};

	export let title: string = '';
	// Form data
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
		promotion_start_date: getTodayDateInput(),
		promotion_end_date: getDateInputFromToday(7),
		sub_type: 'stnd'
	};
	export let isSubmitting: boolean = false;
	export let mode: 'create' | 'edit' = 'create';

	// Remove createEventDispatcher usage
	// const dispatch = createEventDispatcher();

	let newFeature = '';
	
	// Stripe product modal state
	let showStripeProductModal = false;
	let selectedStripeProduct: StripeProduct | null = null;

	// Update formData when plan prop changes (for edit mode)
	$: if (plan && mode === 'edit') {
		formData = {
			name: plan.name || '',
			description: plan.description || '',
			short_desc: plan.short_desc || '',
			price: plan.price || 0,
			currency: plan.currency || 'USD',
			interval: plan.interval || 'month',
			interval_count: plan.interval_count || 1,
			stripe_price_id: plan.stripe_price_id || '',
			stripe_product_id: plan.stripe_product_id || null,
			features: plan.features ? [...plan.features] : [],
			is_active: plan.is_active !== undefined ? plan.is_active : true,
			promotion_start_date: plan.promotion_start_date || null,
			promotion_end_date: plan.promotion_end_date || null,
			sub_type: plan.sub_type || 'stnd'
		};
	}

	// Convert ISO dates to HTML date input format for the form
	$: formData.promotion_start_date = isoToDateInput(formData.promotion_start_date);
	$: formData.promotion_end_date = isoToDateInput(formData.promotion_end_date);

	function handleSubmit() {
		// Convert date input values back to ISO format before submitting
		const submitData = {
			...formData,
			promotion_start_date: dateInputToIso(formData.promotion_start_date),
			promotion_end_date: dateInputToIso(formData.promotion_end_date)
		};
		onSave(submitData);
	}

	function handleCancel() {
		onCancel();
	}

	function addFeature() {
		if (newFeature.trim()) {
			formData.features = [...formData.features, newFeature.trim()];
			newFeature = '';
		}
	}

	function removeFeature(index: number) {
		formData.features = formData.features.filter((_, i) => i !== index);
	}

	// Stripe product handlers
	function openStripeProductModal() {
		showStripeProductModal = true;
	}

	function handleStripeProductSelect(product: StripeProduct | null) {
		selectedStripeProduct = product;
		formData.stripe_product_id = product?.stripe_id || null;
		showStripeProductModal = false;
	}

	function clearStripeProduct() {
		selectedStripeProduct = null;
		formData.stripe_product_id = null;
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			event.preventDefault();
			addFeature();
		}
	}

	function setDefaultDates() {
		formData.promotion_start_date = getTodayDateInput();
		formData.promotion_end_date = getDateInputFromToday(7);
	}
</script>

{#if isOpen}
	<div class="modal-backdrop" on:click={handleCancel}>
		<div class="modal-content" on:click|stopPropagation>
			<div class="modal-header">
				<h2 class="modal-title">{title}</h2>
				<button class="modal-close" on:click={handleCancel} aria-label="Close">
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
					</svg>
				</button>
			</div>

			<form on:submit|preventDefault={handleSubmit} class="modal-form">
				<!-- Basic Information -->
				<div class="form-grid">
					<div class="form-group">
						<label for="name" class="form-label">Plan Name</label>
						<input
							id="name"
							type="text"
							bind:value={formData.name}
							class="form-input"
							required
						/>
					</div>
					<div class="form-group">
						<label for="short_desc" class="form-label">Short Description</label>
						<input
							id="short_desc"
							type="text"
							bind:value={formData.short_desc}
							class="form-input"
						/>
					</div>
				</div>

				<div class="form-group">
					<label for="description" class="form-label">Description</label>
					<textarea
						id="description"
						bind:value={formData.description}
						rows="3"
						class="form-textarea"
						required
					></textarea>
				</div>

				<!-- Pricing -->
				<div class="form-grid">
					<div class="form-group">
						<label for="price" class="form-label">Price</label>
						<input
							id="price"
							type="number"
							step="0.01"
							min="0"
							bind:value={formData.price}
							class="form-input"
							required
						/>
					</div>
					<div class="form-group">
						<label for="currency" class="form-label">Currency</label>
						<select
							id="currency"
							bind:value={formData.currency}
							class="form-select"
						>
							<option value="USD">USD</option>
							<option value="EUR">EUR</option>
							<option value="GBP">GBP</option>
						</select>
					</div>
					<div class="form-group">
						<label for="interval" class="form-label">Billing Interval</label>
						<select
							id="interval"
							bind:value={formData.interval}
							class="form-select"
						>
							<option value="month">Monthly</option>
							<option value="year">Annual</option>
							<option value="week">Weekly</option>
							<option value="day">Daily</option>
						</select>
					</div>
					<div class="form-group">
						<label for="interval_count" class="form-label">Interval Count</label>
						<input
							id="interval_count"
							type="number"
							min="1"
							bind:value={formData.interval_count}
							class="form-input"
							required
						/>
					</div>
					<div class="form-group">
						<label for="stripe_price_id" class="form-label">Stripe Price ID</label>
						<input
							id="stripe_price_id"
							type="text"
							bind:value={formData.stripe_price_id}
							class="form-input"
							placeholder="price_..."
						/>
					</div>
					
					<!-- Stripe Product Selection -->
					<div class="form-group">
						<label class="form-label">Stripe Product</label>
						<div class="stripe-product-selector">
							{#if selectedStripeProduct}
								<div class="selected-product">
									<div class="product-info">
										<span class="product-name">{selectedStripeProduct.name}</span>
										<span class="product-id">{selectedStripeProduct.stripe_id}</span>
									</div>
									<div class="product-actions">
										<button 
											type="button" 
											class="btn-change" 
											on:click={openStripeProductModal}
										>
											Change
										</button>
										<button 
											type="button" 
											class="btn-clear" 
											on:click={clearStripeProduct}
										>
											Clear
										</button>
									</div>
								</div>
							{:else}
								<button 
									type="button" 
									class="btn-select-product" 
									on:click={openStripeProductModal}
								>
									Select Stripe Product
								</button>
							{/if}
						</div>
					</div>
				</div>

				<!-- Features -->
				<div class="form-group">
					<label class="form-label">Features</label>
					<div class="features-list">
						{#each formData.features as feature, index}
							<div class="feature-item">
								<span class="feature-text">{feature}</span>
								<button 
									type="button"
									class="feature-remove"
									on:click={() => removeFeature(index)}
									aria-label="Remove feature"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
									</svg>
								</button>
							</div>
						{/each}
					</div>
					<div class="feature-add">
						<input
							type="text"
							bind:value={newFeature}
							placeholder="Add a feature..."
							class="form-input"
							on:keydown={handleKeydown}
						/>
						<button
							type="button"
							class="btn btn-secondary"
							on:click={addFeature}
						>
							Add
						</button>
					</div>
				</div>

				<!-- Settings -->
				<div class="form-group">
					<div class="checkbox-group">
						<input
							id="is_active"
							type="checkbox"
							bind:checked={formData.is_active}
							class="form-checkbox"
						/>
						<label for="is_active" class="checkbox-label">Active</label>
					</div>
					<div class="form-group">
						<label for="sub_type" class="form-label">Plan Type</label>
						<select
							id="sub_type"
							bind:value={formData.sub_type}
							class="form-select"
							required
						>
							<option value="stnd">Standard Plan</option>
							<option value="prmo">Promotional Plan</option>
						</select>
					</div>
				</div>

				<!-- Promotion Dates (only show for promotional plans) -->
				{#if formData.sub_type === 'prmo'}
					<div class="form-group">
						<label class="form-label">Promotion Settings</label>
						<div class="promotion-options">
							<div class="radio-group">
								<input
									id="default_dates"
									type="radio"
									name="promotion_dates"
									value="default"
									checked={!formData.promotion_start_date && !formData.promotion_end_date}
									on:change={() => {
										setDefaultDates();
									}}
								/>
								<label for="default_dates">Default (Today to 7 days from now)</label>
							</div>
							<div class="radio-group">
								<input
									id="custom_dates"
									type="radio"
									name="promotion_dates"
									value="custom"
									checked={!!formData.promotion_start_date && !!formData.promotion_end_date}
									on:change={() => {
										// Keep existing dates if they exist
									}}
								/>
								<label for="custom_dates">Custom dates</label>
							</div>
						</div>
						
						<div class="form-grid">
							<div class="form-group">
								<label for="promotion_start_date" class="form-label">Start Date</label>
								<input
									id="promotion_start_date"
									type="date"
									bind:value={formData.promotion_start_date}
									class="form-input"
									min={getTodayDateInput()}
								/>
							</div>
							<div class="form-group">
								<label for="promotion_end_date" class="form-label">End Date</label>
								<input
									id="promotion_end_date"
									type="date"
									bind:value={formData.promotion_end_date}
									class="form-input"
									min={formData.promotion_start_date || getTodayDateInput()}
								/>
							</div>
						</div>
					</div>
				{/if}

				<!-- Actions -->
				<div class="modal-actions">
					<button
						type="button"
						class="btn btn-secondary"
						on:click={handleCancel}
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={isSubmitting}
						class="btn btn-primary"
					>
						{isSubmitting ? (mode === 'create' ? 'Creating...' : 'Updating...') : (mode === 'create' ? 'Create Plan' : 'Update Plan')}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Stripe Product Selection Modal -->
<StripeProductModal
	bind:show={showStripeProductModal}
	selectedProductId={formData.stripe_product_id}
	onSelect={handleStripeProductSelect}
	onClose={() => showStripeProductModal = false}
/>

<style>
	.modal-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.7);
		backdrop-filter: blur(10px);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 50;
		padding: 1rem;
	}

	.modal-content {
		background: rgba(255, 255, 255, 0.98);
		backdrop-filter: blur(25px);
		border-radius: 16px;
		border: 1px solid rgba(255, 255, 255, 0.3);
		box-shadow: 
			0 25px 50px rgba(0, 0, 0, 0.4),
			0 0 0 1px rgba(255, 255, 255, 0.2),
			0 0 0 4px rgba(59, 130, 246, 0.1);
		width: 100%;
		max-width: 1200px;
		max-height: 90vh;
		overflow-y: auto;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1.5rem 1.5rem 0 1.5rem;
		border-bottom: 1px solid rgba(0, 0, 0, 0.1);
		margin-bottom: 1.5rem;
	}

	.modal-title {
		font-size: 1.25rem;
		font-weight: 600;
		color: #1f2937;
		margin: 0;
	}

	.modal-close {
		background: none;
		border: none;
		color: #6b7280;
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 6px;
		transition: all 0.2s ease;
	}

	.modal-close:hover {
		background: rgba(0, 0, 0, 0.05);
		color: #374151;
	}

	.modal-form {
		padding: 0 1.5rem 1.5rem 1.5rem;
	}

	.form-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.form-group {
		margin-bottom: 1rem;
	}

	.form-label {
		display: block;
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
		margin-bottom: 0.5rem;
	}

	.form-input,
	.form-select,
	.form-textarea {
		width: 100%;
		padding: 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 8px;
		font-size: 0.875rem;
		transition: all 0.2s ease;
		background: white;
		color: #1f2937;
	}

	.form-input:focus,
	.form-select:focus,
	.form-textarea:focus {
		outline: none;
		border-color: #3b82f6;
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	.form-textarea {
		resize: vertical;
		min-height: 80px;
	}

	.checkbox-group {
		display: flex;
		align-items: center;
		margin-bottom: 0.5rem;
	}

	.form-checkbox {
		margin-right: 0.5rem;
		width: 1rem;
		height: 1rem;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		background: white;
	}

	.checkbox-label {
		font-size: 0.875rem;
		color: #374151;
		cursor: pointer;
	}

	.features-list {
		margin-bottom: 0.5rem;
	}

	.feature-item {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.5rem;
		background: #f9fafb;
		border-radius: 6px;
		margin-bottom: 0.25rem;
	}

	.feature-text {
		flex: 1;
		font-size: 0.875rem;
		color: #374151;
	}

	.feature-remove {
		min-width: 25px;
		background: none;
		border: none;
		color: #ef4444;
		cursor: pointer;
		padding: 0.25rem;
		border-radius: 4px;
		transition: all 0.2s ease;
	}

	.feature-remove:hover {
		background: rgba(239, 68, 68, 0.1);
	}

	.feature-add {
		display: flex;
		gap: 0.5rem;
	}

	.feature-add .form-input {
		flex: 1;
	}

	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
		padding-top: 1.5rem;
		border-top: 1px solid rgba(0, 0, 0, 0.1);
		margin-top: 1.5rem;
	}

	.promotion-options {
		margin-bottom: 1rem;
	}

	.radio-group {
		display: flex;
		align-items: center;
		margin-bottom: 0.5rem;
	}

	.radio-group input[type="radio"] {
		margin-right: 0.5rem;
	}

	.radio-group label {
		font-size: 0.875rem;
		color: #374151;
		cursor: pointer;
	}

	/* Stripe Product Selector Styles */
	.stripe-product-selector {
		margin-top: 0.5rem;
	}

	.selected-product {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 6px;
		background: #f9fafb;
	}

	.product-info {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.product-name {
		font-weight: 600;
		color: #111827;
	}

	.product-id {
		font-size: 0.875rem;
		color: #6b7280;
		font-family: monospace;
	}

	.product-actions {
		display: flex;
		gap: 0.5rem;
	}

	.btn-change, .btn-clear {
		padding: 0.375rem 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		font-size: 0.875rem;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.btn-change {
		background: white;
		color: #374151;
	}

	.btn-change:hover {
		background: #f3f4f6;
		border-color: #9ca3af;
	}

	.btn-clear {
		background: #fee2e2;
		color: #dc2626;
		border-color: #fecaca;
	}

	.btn-clear:hover {
		background: #fecaca;
		border-color: #f87171;
	}

	.btn-select-product {
		width: 100%;
		padding: 0.75rem;
		border: 2px dashed #d1d5db;
		border-radius: 6px;
		background: white;
		color: #6b7280;
		font-size: 0.875rem;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.btn-select-product:hover {
		border-color: #3b82f6;
		color: #3b82f6;
		background: #eff6ff;
	}

	@media (max-width: 768px) {
		.modal-content {
			margin: 0.5rem;
			max-height: calc(100vh - 1rem);
		}

		.form-grid {
			grid-template-columns: 1fr;
		}

		.modal-actions {
			flex-direction: column;
		}

		.modal-actions .btn {
			width: 100%;
		}

		.product-actions {
			flex-direction: column;
		}

		.selected-product {
			flex-direction: column;
			align-items: flex-start;
			gap: 0.75rem;
		}
	}
</style> 
