<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { showToast } from '$lib/toast';
	import { StripeOfferIntegrationService, type CreateOfferWithStripeRequest } from '$lib/services/stripe-offers-integration';
	import type { SubscriptionPlan } from '$lib/services/streaming-subscriptions';

	export let show = false;
	export let subscriptionPlans: SubscriptionPlan[] = [];

	const dispatch = createEventDispatcher();

	let formData: CreateOfferWithStripeRequest = {
		plan_id: 0,
		off_discount_type: 'percentage',
		off_discount_value: 0,
		is_active: true,
		off_name: '',
		off_current_uses: 0,
		off_priority: 5,
		off_auto_apply: false,
		auto_create_stripe: true
	};

	let isSubmitting = false;
	let stripeEnabled = false;

	// Check Stripe connection on component mount
	$: if (show) {
		checkStripeConnection();
	}

	async function checkStripeConnection() {
		try {
			const connection = await StripeOfferIntegrationService.checkStripeConnection();
			stripeEnabled = connection.enabled;
		} catch (error) {
			console.error('Failed to check Stripe connection:', error);
			stripeEnabled = false;
		}
	}

	async function handleSubmit() {
		if (!formData.off_name.trim()) {
			showToast('Please enter an offer name', 'error');
			return;
		}

		if (formData.plan_id === 0) {
			showToast('Please select a subscription plan', 'error');
			return;
		}

		if (formData.off_discount_value <= 0) {
			showToast('Please enter a valid discount value', 'error');
			return;
		}

		try {
			isSubmitting = true;
			const newOffer = await StripeOfferIntegrationService.createOfferWithStripe(formData);
			
			showToast(
				`Offer created successfully${formData.auto_create_stripe && stripeEnabled ? ' with Stripe integration' : ''}`, 
				'success'
			);
			
			dispatch('offerCreated', newOffer);
			resetForm();
			show = false;
		} catch (error: any) {
			console.error('Failed to create offer:', error);
			showToast(`Failed to create offer: ${error.message || 'Unknown error'}`, 'error');
		} finally {
			isSubmitting = false;
		}
	}

	function resetForm() {
		formData = {
			plan_id: 0,
			off_discount_type: 'percentage',
			off_discount_value: 0,
			is_active: true,
			off_name: '',
			off_current_uses: 0,
			off_priority: 5,
			off_auto_apply: false,
			auto_create_stripe: true
		};
	}

	function handleCancel() {
		resetForm();
		show = false;
	}

	// Reactive discount preview
	$: discountPreview = formData.off_discount_type === 'percentage' 
		? `${formData.off_discount_value}% OFF`
		: `$${formData.off_discount_value} OFF`;
</script>

{#if show}
	<div class="modal-backdrop" on:click={handleCancel}>
		<div class="modal-content" on:click|stopPropagation>
			<div class="modal-header">
				<div class="modal-title-with-icon">
					<span class="modal-icon">🎟️</span>
					<h2 class="modal-title">Create Offer with Stripe</h2>
				</div>
				<button type="button" class="modal-close" on:click={handleCancel} aria-label="Close">
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
					</svg>
				</button>
			</div>

			<div class="modal-body">
				<form on:submit|preventDefault={handleSubmit}>
					<!-- Basic Information -->
					<div class="form-section">
						<h3 class="section-title">Basic Information</h3>
						
						<div class="form-group">
							<label for="off_name" class="form-label">Offer Name *</label>
							<input
								id="off_name"
								type="text"
								bind:value={formData.off_name}
								placeholder="e.g., Summer Sale, Early Bird Discount"
								class="form-input"
								required
							/>
						</div>

						<div class="form-group">
							<label for="plan_id" class="form-label">Subscription Plan *</label>
							<select
								id="plan_id"
								bind:value={formData.plan_id}
								class="form-select"
								required
							>
								<option value={0}>Select a plan...</option>
								{#each subscriptionPlans as plan}
									<option value={plan.id}>{plan.name} - ${plan.price}/{plan.interval}</option>
								{/each}
							</select>
						</div>

						<div class="form-group">
							<label for="off_description" class="form-label">Description</label>
							<textarea
								id="off_description"
								bind:value={formData.off_description}
								placeholder="Optional description of this offer"
								class="form-textarea"
								rows="3"
							></textarea>
						</div>
					</div>

					<!-- Discount Configuration -->
					<div class="form-section">
						<h3 class="section-title">Discount Configuration</h3>
						
						<div class="form-row">
							<div class="form-group">
								<label for="off_discount_type" class="form-label">Discount Type *</label>
								<select
									id="off_discount_type"
									bind:value={formData.off_discount_type}
									class="form-select"
								>
									<option value="percentage">Percentage</option>
									<option value="amount">Fixed Amount</option>
								</select>
							</div>

							<div class="form-group">
								<label for="off_discount_value" class="form-label">
									{formData.off_discount_type === 'percentage' ? 'Percentage' : 'Amount ($)'} *
								</label>
								<input
									id="off_discount_value"
									type="number"
									step={formData.off_discount_type === 'percentage' ? '0.1' : '0.01'}
									min="0"
									max={formData.off_discount_type === 'percentage' ? '100' : undefined}
									bind:value={formData.off_discount_value}
									class="form-input"
									required
								/>
							</div>
						</div>

						<div class="discount-preview">
							<span class="preview-label">Preview:</span>
							<span class="preview-value">{discountPreview}</span>
						</div>
					</div>

					<!-- Offer Settings -->
					<div class="form-section">
						<h3 class="section-title">Offer Settings</h3>
						
						<div class="form-row">
							<div class="form-group">
								<label for="off_code" class="form-label">Promotion Code</label>
								<input
									id="off_code"
									type="text"
									bind:value={formData.off_code}
									placeholder="e.g., SUMMER25, EARLY50"
									class="form-input"
								/>
							</div>

							<div class="form-group">
								<label for="off_max_uses" class="form-label">Max Uses</label>
								<input
									id="off_max_uses"
									type="number"
									min="1"
									bind:value={formData.off_max_uses}
									placeholder="Leave empty for unlimited"
									class="form-input"
								/>
							</div>
						</div>

						<div class="form-row">
							<div class="form-group">
								<label for="off_priority" class="form-label">Priority (1-10)</label>
								<input
									id="off_priority"
									type="number"
									min="1"
									max="10"
									bind:value={formData.off_priority}
									class="form-input"
								/>
							</div>

							<div class="form-group checkbox-group">
								<label class="checkbox-label">
									<input
										type="checkbox"
										bind:checked={formData.off_auto_apply}
										class="form-checkbox"
									/>
									<span class="checkbox-text">Auto-apply (no code needed)</span>
								</label>
							</div>
						</div>

						<div class="form-group checkbox-group">
							<label class="checkbox-label">
								<input
									type="checkbox"
									bind:checked={formData.is_active}
									class="form-checkbox"
								/>
								<span class="checkbox-text">Active immediately</span>
							</label>
						</div>
					</div>

					<!-- Stripe Integration -->
					<div class="form-section stripe-section">
						<h3 class="section-title">🔗 Stripe Integration</h3>
						
						{#if stripeEnabled}
							<div class="stripe-status connected">
								<span class="status-indicator">✅ Stripe Connected</span>
								<p class="status-description">
									Your offer will be automatically synchronized with Stripe as a coupon
									{#if formData.off_code}and promotion code{/if}.
								</p>
							</div>

							<div class="form-group checkbox-group">
								<label class="checkbox-label">
									<input
										type="checkbox"
										bind:checked={formData.auto_create_stripe}
										class="form-checkbox"
									/>
									<span class="checkbox-text">Create Stripe coupon automatically</span>
								</label>
							</div>

							{#if formData.auto_create_stripe}
								<div class="stripe-preview">
									<h4>Stripe Preview:</h4>
									<ul>
										<li><strong>Coupon:</strong> {discountPreview} discount</li>
										{#if formData.off_code}
											<li><strong>Promotion Code:</strong> {formData.off_code}</li>
										{/if}
										<li><strong>Duration:</strong> Once (single use)</li>
										{#if formData.off_max_uses}
											<li><strong>Max Redemptions:</strong> {formData.off_max_uses}</li>
										{/if}
									</ul>
								</div>
							{/if}
						{:else}
							<div class="stripe-status disconnected">
								<span class="status-indicator">❌ Stripe Not Connected</span>
								<p class="status-description">
									Connect Stripe in the dashboard to enable automatic coupon creation.
									Your offer will still be created normally.
								</p>
							</div>
						{/if}
					</div>
				</form>
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
					type="button"
					class="btn btn-primary"
					on:click={handleSubmit}
					disabled={isSubmitting}
				>
					{isSubmitting ? 'Creating...' : 'Create Offer'}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.modal-backdrop {
		position: fixed;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}

	.modal-content {
		background: var(--card-bg);
		border-radius: 12px;
		box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
		max-width: 600px;
		width: 90%;
		max-height: 90vh;
		overflow-y: auto;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1.5rem 1.5rem 0;
		border-bottom: 1px solid var(--border);
		margin-bottom: 1.5rem;
	}

	.modal-title-with-icon {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.modal-icon {
		font-size: 1.5rem;
	}

	.modal-title {
		margin: 0;
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text);
	}

	.modal-close {
		background: none;
		border: none;
		cursor: pointer;
		color: var(--text-muted);
		padding: 0.5rem;
		border-radius: 6px;
		transition: all 0.2s ease;
	}

	.modal-close:hover {
		background: var(--hover-bg);
		color: var(--text);
	}

	.modal-body {
		padding: 0 1.5rem;
	}

	.form-section {
		margin-bottom: 2rem;
	}

	.section-title {
		margin: 0 0 1rem;
		font-size: 1rem;
		font-weight: 600;
		color: var(--text);
		border-bottom: 1px solid var(--border);
		padding-bottom: 0.5rem;
	}

	.form-group {
		margin-bottom: 1rem;
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
	}

	.form-label {
		display: block;
		margin-bottom: 0.5rem;
		font-weight: 500;
		color: var(--text);
		font-size: 0.875rem;
	}

	.form-input,
	.form-select,
	.form-textarea {
		width: 100%;
		padding: 0.75rem;
		border: 1px solid var(--border);
		border-radius: 6px;
		font-size: 0.875rem;
		background: var(--input-bg);
		color: var(--text);
		transition: border-color 0.2s ease;
	}

	.form-input:focus,
	.form-select:focus,
	.form-textarea:focus {
		outline: none;
		border-color: var(--primary);
		box-shadow: 0 0 0 3px rgba(var(--primary-rgb), 0.1);
	}

	.checkbox-group {
		display: flex;
		align-items: center;
	}

	.checkbox-label {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		cursor: pointer;
		font-size: 0.875rem;
	}

	.form-checkbox {
		width: auto;
		margin: 0;
	}

	.discount-preview {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem;
		background: var(--success-bg);
		border-radius: 6px;
		margin-top: 1rem;
	}

	.preview-label {
		font-weight: 500;
		color: var(--text-muted);
	}

	.preview-value {
		font-weight: 600;
		color: var(--success);
		font-size: 1.1rem;
	}

	.stripe-section {
		border: 1px solid var(--border);
		border-radius: 8px;
		padding: 1rem;
		background: var(--card-bg-secondary);
	}

	.stripe-status {
		margin-bottom: 1rem;
	}

	.stripe-status.connected {
		color: var(--success);
	}

	.stripe-status.disconnected {
		color: var(--warning);
	}

	.status-indicator {
		font-weight: 600;
		display: block;
		margin-bottom: 0.5rem;
	}

	.status-description {
		font-size: 0.875rem;
		color: var(--text-muted);
		margin: 0;
	}

	.stripe-preview {
		margin-top: 1rem;
		padding: 0.75rem;
		background: var(--primary-bg);
		border-radius: 6px;
		border: 1px solid var(--primary);
	}

	.stripe-preview h4 {
		margin: 0 0 0.5rem;
		font-size: 0.875rem;
		color: var(--primary);
	}

	.stripe-preview ul {
		margin: 0;
		padding-left: 1.25rem;
		font-size: 0.8125rem;
		color: var(--text-muted);
	}

	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
		padding: 1.5rem;
		border-top: 1px solid var(--border);
	}

	.btn {
		padding: 0.75rem 1.5rem;
		font-size: 0.875rem;
		font-weight: 500;
		border: none;
		border-radius: 6px;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		white-space: nowrap;
	}

	.btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-primary {
		background: var(--primary);
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: var(--primary-hover);
		transform: translateY(-1px);
	}

	.btn-secondary {
		background: var(--secondary-bg);
		color: var(--text);
		border: 1px solid var(--border);
	}

	.btn-secondary:hover:not(:disabled) {
		background: var(--hover-bg);
	}
</style> 
