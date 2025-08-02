<script lang="ts">
	import type { CreateSubscriptionOfferData, UpdateSubscriptionOfferData, SubscriptionOffer } from '$lib/services/subscription-offers';
	import type { SubscriptionPlan } from '$lib/services/streaming-subscriptions';

	export let isOpen = false;
	export let mode: 'create' | 'edit' = 'create';
	export let offer: SubscriptionOffer | null = null;
	export let subscriptionPlans: SubscriptionPlan[] = [];
	export let isSubmitting = false;
	
	// Callback props instead of events
	export let onSubmit: (formData: CreateSubscriptionOfferData | UpdateSubscriptionOfferData) => void = () => {};
	export let onCancel: () => void = () => {};

	// Form data
	let formData: CreateSubscriptionOfferData = {
		plan_id: 1,
		item_id: 0,
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

	// Reactive form validation
	$: isValid = formData.off_name.trim() !== '' && 
				 formData.plan_id > 0 && 
				 formData.off_discount_value > 0 &&
				 formData.off_discount_value <= (formData.off_discount_type === 'percentage' ? 100 : 999999);

	// Initialize form when modal opens or offer changes
	$: if (isOpen && mode === 'edit' && offer) {
		formData = {
			plan_id: offer.plan_id,
			item_id: offer.item_id ? parseInt(offer.item_id) : undefined,
			off_discount_type: offer.off_discount_type,
			off_discount_value: offer.off_discount_value,
			offer_start_date: offer.offer_start_date,
			off_end_date: offer.off_end_date,
			is_active: offer.is_active,
			off_description: offer.off_description,
			off_name: offer.off_name,
			off_code: offer.off_code,
			off_max_uses: offer.off_max_uses,
			off_current_uses: offer.off_current_uses,
			off_terms_conditions: offer.off_terms_conditions,
			off_target: offer.off_target,
			off_priority: offer.off_priority,
			off_auto_apply: offer.off_auto_apply
		};
	} else if (isOpen && mode === 'create') {
		// Reset form for create mode
		formData = {
			plan_id: subscriptionPlans.length > 0 ? parseInt(subscriptionPlans[0].id) : 1,
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
	}

	// Get available plans for dropdown
	$: availablePlans = subscriptionPlans.filter(plan => plan.is_active);

	// Handle form submission
	function handleSubmit(event: Event) {
		event.preventDefault();
		if (!isValid) return;
		
		if (mode === 'create') {
			onSubmit(formData);
		} else {
			onSubmit({ id: Number(offer?.id), ...formData });
		}
	}

	// Handle cancel
	function handleCancel() {
		onCancel();
	}

	// Close modal on backdrop click
	function handleBackdropClick(event: MouseEvent) {
		if (event.target === event.currentTarget) {
			handleCancel();
		}
	}

	// Format date for input
	function formatDateForInput(dateString: string | undefined): string {
		if (!dateString) return '';
		return new Date(dateString).toISOString().split('T')[0];
	}
</script>

{#if isOpen}
	<div class="modal-backdrop" on:click={handleBackdropClick}>
		<div class="modal-content">
			<div class="modal-header">
				<h2 class="modal-title">
					{mode === 'create' ? 'Create New Offer' : 'Edit Offer'}
				</h2>
				<button type="button" class="close-button" on:click={handleCancel} aria-label="Close">
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
					</svg>
				</button>
			</div>

			<form on:submit={handleSubmit} class="modal-body">
				<div class="form-grid">
					<!-- Basic Information -->
					<div class="form-section">
						<h3 class="section-title">Basic Information</h3>
						
						<div class="form-group">
							<label for="off_name" class="form-label">Offer Name *</label>
							<input
								id="off_name"
								type="text"
								bind:value={formData.off_name}
								class="form-input"
								placeholder="Enter offer name"
								required
							/>
						</div>

						<div class="form-group">
							<label for="plan_id" class="form-label">Subscription Plan *</label>
							<select id="plan_id" bind:value={formData.plan_id} class="form-select" required>
								<option value="">Select a plan</option>
								{#each availablePlans as plan}
									<option value={parseInt(plan.id)}>
										{plan.name} - ${plan.price}/{plan.interval}
									</option>
								{/each}
							</select>
						</div>

						<div class="form-group">
							<label for="item_id" class="form-label">Item Type</label>
							<select id="item_id" bind:value={formData.item_id} class="form-select">
								<option value="">No specific item (general offer)</option>
								<option value=1>eBook</option>
								<option value=2>DVD</option>
								<option value=3>Expo Ticket</option>
							</select>
							<p class="form-help">Select a specific item type this offer applies to (optional)</p>
						</div>

						<div class="form-group">
							<label for="off_description" class="form-label">Description</label>
							<textarea
								id="off_description"
								bind:value={formData.off_description}
								class="form-textarea"
								placeholder="Enter offer description"
								rows="3"
							></textarea>
						</div>
					</div>

					<!-- Discount Configuration -->
					<div class="form-section">
						<h3 class="section-title">Discount Configuration</h3>
						
						<div class="form-group">
							<label for="off_discount_type" class="form-label">Discount Type *</label>
							<select id="off_discount_type" bind:value={formData.off_discount_type} class="form-select" required>
								<option value="percentage">Percentage (%)</option>
								<option value="fixed">Fixed Amount ($)</option>
							</select>
						</div>

						<div class="form-group">
							<label for="off_discount_value" class="form-label">
								Discount Value * ({formData.off_discount_type === 'percentage' ? '%' : '$'})
							</label>
							<input
								id="off_discount_value"
								type="number"
								bind:value={formData.off_discount_value}
								class="form-input"
								min="0"
								max={formData.off_discount_type === 'percentage' ? 100 : 999999}
								step={formData.off_discount_type === 'percentage' ? 0.01 : 0.01}
								required
							/>
						</div>

						<div class="form-group">
							<label for="off_code" class="form-label">Promo Code</label>
							<input
								id="off_code"
								type="text"
								bind:value={formData.off_code}
								class="form-input"
								placeholder="e.g., SAVE20"
							/>
						</div>
					</div>

					<!-- Usage Limits -->
					<div class="form-section">
						<h3 class="section-title">Usage Limits</h3>
						
						<div class="form-group">
							<label for="off_max_uses" class="form-label">Maximum Uses</label>
							<input
								id="off_max_uses"
								type="number"
								bind:value={formData.off_max_uses}
								class="form-input"
								min="1"
								placeholder="100"
							/>
						</div>

						<div class="form-group">
							<label for="off_priority" class="form-label">Priority</label>
							<input
								id="off_priority"
								type="number"
								bind:value={formData.off_priority}
								class="form-input"
								min="1"
								max="10"
								placeholder="1"
							/>
							<p class="form-help">Higher priority offers are applied first</p>
						</div>

						<div class="form-group">
							<label class="form-label">Auto-Apply</label>
							<div class="checkbox-group">
								<input
									id="off_auto_apply"
									type="checkbox"
									bind:checked={formData.off_auto_apply}
									class="form-checkbox"
								/>
								<label for="off_auto_apply" class="checkbox-label">
									Automatically apply this offer to eligible users
								</label>
							</div>
						</div>
					</div>

					<!-- Date Range -->
					<div class="form-section">
						<h3 class="section-title">Date Range</h3>
						
						<div class="form-group">
							<label for="offer_start_date" class="form-label">Start Date</label>
							<input
								id="offer_start_date"
								type="date"
								bind:value={formData.offer_start_date}
								class="form-input"
							/>
						</div>

						<div class="form-group">
							<label for="off_end_date" class="form-label">End Date</label>
							<input
								id="off_end_date"
								type="date"
								bind:value={formData.off_end_date}
								class="form-input"
							/>
						</div>
					</div>

					<!-- Additional Settings -->
					<div class="form-section">
						<h3 class="section-title">Additional Settings</h3>
						
						<div class="form-group">
							<label for="off_target" class="form-label">Target Audience</label>
							<select id="off_target" bind:value={formData.off_target} class="form-select">
								<option value="">All Users</option>
								<option value="new_users">New Users Only</option>
								<option value="existing_users">Existing Users Only</option>
								<option value="premium_users">Premium Users Only</option>
							</select>
						</div>

						<div class="form-group">
							<label for="off_terms_conditions" class="form-label">Terms & Conditions</label>
							<textarea
								id="off_terms_conditions"
								bind:value={formData.off_terms_conditions}
								class="form-textarea"
								placeholder="Enter terms and conditions"
								rows="3"
							></textarea>
						</div>

						<div class="form-group">
							<label class="form-label">Status</label>
							<div class="checkbox-group">
								<input
									id="is_active"
									type="checkbox"
									bind:checked={formData.is_active}
									class="form-checkbox"
								/>
								<label for="is_active" class="checkbox-label">
									Active (offer is available for use)
								</label>
							</div>
						</div>
					</div>
				</div>

				<div class="modal-actions">
					<button
						type="button"
						class="btn btn-secondary"
						on:click={handleCancel}
						disabled={isSubmitting}
					>
						Cancel
					</button>
					<button
						type="submit"
						class="btn btn-primary"
						disabled={!isValid || isSubmitting}
					>
						{isSubmitting ? 'Saving...' : (mode === 'create' ? 'Create Offer' : 'Update Offer')}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<style>
	.modal-backdrop {
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
		padding: 1rem;
	}

	.modal-content {
		background: white;
		border-radius: 0.5rem;
		box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
		max-width: 1200px;
		width: 100%;
		max-height: 90vh;
		overflow: hidden;
		display: flex;
		flex-direction: column;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1.5rem;
		border-bottom: 1px solid #e5e7eb;
	}

	.modal-title {
		font-size: 1.25rem;
		font-weight: 600;
		color: #111827;
		margin: 0;
	}

	.close-button {
		background: none;
		border: none;
		color: #6b7280;
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 0.375rem;
		transition: all 0.2s ease;
	}

	.close-button:hover {
		background: #f3f4f6;
		color: #374151;
	}

	.modal-body {
		padding: 1.5rem;
		overflow-y: auto;
		flex: 1;
	}

	.form-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: 2rem;
	}

	.form-section {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.section-title {
		font-size: 1.125rem;
		font-weight: 600;
		color: #111827;
		margin: 0 0 1rem 0;
		border-bottom: 2px solid #e5e7eb;
		padding-bottom: 0.5rem;
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.form-label {
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
	}

	.form-input,
	.form-select,
	.form-textarea {
		padding: 0.5rem 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		transition: border-color 0.2s ease;
	}

	.form-input:focus,
	.form-select:focus,
	.form-textarea:focus {
		outline: none;
		border-color: #2563eb;
		box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
	}

	.form-textarea {
		resize: vertical;
		min-height: 80px;
	}

	.form-help {
		font-size: 0.75rem;
		color: #6b7280;
		margin: 0;
	}

	.checkbox-group {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.form-checkbox {
		width: 1rem;
		height: 1rem;
		border: 1px solid #d1d5db;
		border-radius: 0.25rem;
	}

	.checkbox-label {
		font-size: 0.875rem;
		color: #374151;
		cursor: pointer;
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

	.btn-primary {
		background: #2563eb;
		color: white;
	}

	.btn-primary:hover {
		background: #1d4ed8;
		box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
	}

	.btn-primary:disabled {
		background: #9ca3af;
		box-shadow: none;
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.modal-content {
			margin: 0;
			border-radius: 0;
			max-height: 100vh;
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
	}
</style> 