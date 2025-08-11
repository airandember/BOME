<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import { StripeSubscriptionIntegrationService, type CreatePlanWithStripeRequest } from '$lib/services/stripe-subscription-integration';
	import { StreamingSubscriptionService } from '$lib/services/streaming-subscriptions';
	import { showToast } from '$lib/toast';

	const dispatch = createEventDispatcher();

	export let show = false;
	export let isEditing = false;
	export let existingPlan: any = null;

	// Form data
	let formData: CreatePlanWithStripeRequest = {
		name: '',
		description: '',
		short_desc: '',
		price: 0,
		currency: 'USD',
		interval: 'month',
		interval_count: 1,
		features: [],
		is_active: true,
		sub_type: 'stnd',
		auto_create_stripe: true, // Default to true
		promotion_start_date: undefined,
		promotion_end_date: undefined
	};

	let isSubmitting = false;
	let stripeEnabled = false;
	let loadingStripeStatus = true;
	let newFeature = '';

	onMount(async () => {
		// Check if Stripe is connected
		try {
			stripeEnabled = await StripeSubscriptionIntegrationService.checkStripeConnection();
		} catch (error) {
			console.error('Failed to check Stripe connection:', error);
			stripeEnabled = false;
		} finally {
			loadingStripeStatus = false;
		}
	});

	async function handleSubmit() {
		if (isSubmitting) return;

		try {
			isSubmitting = true;

			let result;
			if (formData.auto_create_stripe && stripeEnabled) {
				result = await StripeSubscriptionIntegrationService.createPlanWithStripe(formData);
				showToast(
					`Plan created successfully${result.stripe_price_id ? ' with Stripe integration' : ''}`, 
					'success'
				);
			} else {
				// Use existing subscription service
				const legacyFormData = {
					...formData,
					stripe_price_id: '', // Empty for manual entry later
					promotion_start_date: formData.promotion_start_date || null,
					promotion_end_date: formData.promotion_end_date || null,
					sub_type: formData.sub_type as 'stnd' | 'prmo', // Type cast to fix TypeScript error
				};
				result = await StreamingSubscriptionService.create(legacyFormData);
				showToast('Plan created successfully', 'success');
			}

			dispatch('planCreated', result);
			closeModal();
		} catch (error: any) {
			console.error('Failed to create plan:', error);
			showToast(`Failed to create plan: ${error.message || 'Unknown error'}`, 'error');
		} finally {
			isSubmitting = false;
		}
	}

	function addFeature() {
		if (newFeature.trim() && !formData.features.includes(newFeature.trim())) {
			formData.features = [...formData.features, newFeature.trim()];
			newFeature = '';
		}
	}

	function removeFeature(index: number) {
		formData.features = formData.features.filter((_, i) => i !== index);
	}

	function closeModal() {
		show = false;
		// Reset form
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
			sub_type: 'stnd',
			auto_create_stripe: true,
		};
		newFeature = '';
	}

	$: stripePreviewAmount = (formData.price * 100).toFixed(0);
</script>

{#if show}
	<div class="modal-overlay" on:click={closeModal}>
		<div class="modal-content" on:click|stopPropagation>
			<div class="modal-header">
				<h2>{isEditing ? 'Edit' : 'Create'} Subscription Plan</h2>
				<button class="close-btn" on:click={closeModal}>×</button>
			</div>

			<form on:submit|preventDefault={handleSubmit}>
				<!-- Basic Plan Information -->
				<div class="form-section">
					<h3>Plan Details</h3>
					
					<div class="form-group">
						<label for="name">Plan Name *</label>
						<input
							id="name"
							type="text"
							bind:value={formData.name}
							required
							placeholder="e.g., Premium Monthly"
						/>
					</div>

					<div class="form-group">
						<label for="short_desc">Short Description</label>
						<input
							id="short_desc"
							type="text"
							bind:value={formData.short_desc}
							placeholder="Brief plan description"
						/>
					</div>

					<div class="form-group">
						<label for="description">Description</label>
						<textarea
							id="description"
							bind:value={formData.description}
							placeholder="Detailed plan description"
							rows="3"
						></textarea>
					</div>
				</div>

				<!-- Pricing Information -->
				<div class="form-section">
					<h3>Pricing</h3>
					
					<div class="form-row">
						<div class="form-group">
							<label for="price">Price *</label>
							<input
								id="price"
								type="number"
								step="0.01"
								min="0"
								bind:value={formData.price}
								required
							/>
						</div>

						<div class="form-group">
							<label for="currency">Currency</label>
							<select id="currency" bind:value={formData.currency}>
								<option value="USD">USD</option>
								<option value="EUR">EUR</option>
								<option value="GBP">GBP</option>
								<option value="CAD">CAD</option>
							</select>
						</div>
					</div>

					<div class="form-row">
						<div class="form-group">
							<label for="interval">Billing Interval</label>
							<select id="interval" bind:value={formData.interval}>
								<option value="month">Monthly</option>
								<option value="year">Yearly</option>
								<option value="week">Weekly</option>
								<option value="day">Daily</option>
							</select>
						</div>

						<div class="form-group">
							<label for="interval_count">Interval Count</label>
							<input
								id="interval_count"
								type="number"
								min="1"
								bind:value={formData.interval_count}
								required
							/>
						</div>
					</div>
				</div>

				<!-- Features -->
				<div class="form-section">
					<h3>Features</h3>
					
					<div class="feature-input">
						<input
							type="text"
							bind:value={newFeature}
							placeholder="Add a feature..."
							on:keypress={(e) => e.key === 'Enter' && (e.preventDefault(), addFeature())}
						/>
						<button type="button" class="btn btn-sm btn-secondary" on:click={addFeature}>
							Add
						</button>
					</div>

					{#if formData.features.length > 0}
						<div class="features-list">
							{#each formData.features as feature, index}
								<div class="feature-item">
									<span>{feature}</span>
									<button type="button" class="remove-feature" on:click={() => removeFeature(index)}>
										×
									</button>
								</div>
							{/each}
						</div>
					{/if}
				</div>

				<!-- Stripe Integration Section -->
				{#if loadingStripeStatus}
					<div class="form-section">
						<div class="loading-stripe">
							<div class="spinner"></div>
							<span>Checking Stripe connection...</span>
						</div>
					</div>
				{:else if stripeEnabled}
					<div class="stripe-integration-section">
						<h3>🔗 Stripe Integration</h3>
						<div class="integration-option">
							<label class="checkbox-label">
								<input
									type="checkbox"
									bind:checked={formData.auto_create_stripe}
								/>
								<span class="checkmark"></span>
								Automatically create Stripe product and pricing
							</label>
							<p class="help-text">
								When enabled, this will automatically create a corresponding product and price in your Stripe account.
							</p>
						</div>

						{#if formData.auto_create_stripe}
							<div class="stripe-preview">
								<h4>Stripe Preview</h4>
								<div class="preview-card">
									<div class="preview-item">
										<strong>Product:</strong> {formData.name || 'Plan Name'}
									</div>
									<div class="preview-item">
										<strong>Price:</strong> ${formData.price.toFixed(2)} {formData.currency} per {formData.interval}
									</div>
									<div class="preview-item">
										<strong>Stripe Amount:</strong> {stripePreviewAmount} cents
									</div>
									{#if formData.description}
										<div class="preview-item">
											<strong>Description:</strong> {formData.description}
										</div>
									{/if}
								</div>
							</div>
						{/if}
					</div>
				{:else}
					<div class="stripe-disabled-section">
						<h3>🔗 Stripe Integration</h3>
						<div class="disabled-notice">
							<p>⚠️ Stripe is not connected. Plans will be created without payment integration.</p>
							<p>Connect Stripe in the <a href="/admin/streaming/stripe" target="_blank">Stripe Dashboard</a> to enable automatic payment processing.</p>
						</div>
					</div>
				{/if}

				<!-- Plan Settings -->
				<div class="form-section">
					<h3>Settings</h3>
					
					<div class="form-row">
						<div class="form-group">
							<label class="checkbox-label">
								<input type="checkbox" bind:checked={formData.is_active} />
								<span class="checkmark"></span>
								Active Plan
							</label>
						</div>

						<div class="form-group">
							<label for="sub_type">Plan Type</label>
							<select id="sub_type" bind:value={formData.sub_type}>
								<option value="stnd">Standard</option>
								<option value="prmo">Promotional</option>
							</select>
						</div>
					</div>
				</div>

				<div class="modal-actions">
					<button type="button" class="btn btn-secondary" on:click={closeModal}>
						Cancel
					</button>
					<button type="submit" class="btn btn-primary" disabled={isSubmitting}>
						{#if isSubmitting}
							Creating...
						{:else if formData.auto_create_stripe && stripeEnabled}
							Create Plan + Stripe
						{:else}
							Create Plan
						{/if}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<style>
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

	.modal-content {
		background: var(--surface);
		border-radius: var(--radius-lg);
		padding: var(--space-xl);
		max-width: 600px;
		width: 90vw;
		max-height: 90vh;
		overflow-y: auto;
		border: 1px solid var(--border);
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-lg);
	}

	.modal-header h2 {
		margin: 0;
		color: var(--text);
	}

	.close-btn {
		background: none;
		border: none;
		font-size: 1.5rem;
		cursor: pointer;
		color: var(--text-muted);
		padding: var(--space-xs);
	}

	.close-btn:hover {
		color: var(--text);
	}

	.form-section {
		margin-bottom: var(--space-lg);
		padding-bottom: var(--space-lg);
		border-bottom: 1px solid var(--border);
	}

	.form-section:last-of-type {
		border-bottom: none;
	}

	.form-section h3 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
		font-size: 1.1rem;
	}

	.form-group {
		margin-bottom: var(--space-md);
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-md);
	}

	.form-group label {
		display: block;
		margin-bottom: var(--space-xs);
		font-weight: 600;
		color: var(--text);
	}

	.form-group input,
	.form-group select,
	.form-group textarea {
		width: 100%;
		padding: var(--space-sm);
		border: 1px solid var(--border);
		border-radius: var(--radius-md);
		background: var(--bg-primary);
		color: var(--text);
		font-size: 1rem;
	}

	.form-group input:focus,
	.form-group select:focus,
	.form-group textarea:focus {
		outline: none;
		border-color: var(--primary);
		box-shadow: 0 0 0 2px var(--primary-light);
	}

	.feature-input {
		display: flex;
		gap: var(--space-sm);
		margin-bottom: var(--space-md);
	}

	.feature-input input {
		flex: 1;
	}

	.features-list {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-sm);
	}

	.feature-item {
		display: flex;
		align-items: center;
		gap: var(--space-xs);
		background: var(--primary-light);
		color: var(--primary-dark);
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-md);
		font-size: 0.9rem;
	}

	.remove-feature {
		background: none;
		border: none;
		color: var(--primary-dark);
		cursor: pointer;
		font-size: 1.2rem;
		line-height: 1;
		padding: 0;
		margin-left: var(--space-xs);
	}

	.remove-feature:hover {
		color: var(--error);
	}

	.stripe-integration-section {
		margin: var(--space-lg) 0;
		padding: var(--space-md);
		background: var(--surface-secondary);
		border-radius: var(--radius-md);
		border: 1px solid var(--success);
	}

	.stripe-disabled-section {
		margin: var(--space-lg) 0;
		padding: var(--space-md);
		background: var(--surface-secondary);
		border-radius: var(--radius-md);
		border: 1px solid var(--warning);
	}

	.stripe-integration-section h3 {
		margin: 0 0 var(--space-md) 0;
		color: var(--success);
	}

	.stripe-disabled-section h3 {
		margin: 0 0 var(--space-md) 0;
		color: var(--warning);
	}

	.integration-option {
		margin-bottom: var(--space-md);
	}

	.checkbox-label {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		cursor: pointer;
		font-weight: 500;
	}

	.checkbox-label input[type="checkbox"] {
		width: auto;
		margin: 0;
	}

	.help-text {
		margin: var(--space-xs) 0 0 var(--space-xl);
		font-size: 0.9rem;
		color: var(--text-muted);
		line-height: 1.4;
	}

	.stripe-preview {
		margin-top: var(--space-md);
		padding: var(--space-md);
		background: var(--surface);
		border-radius: var(--radius-md);
		border: 1px solid var(--success);
	}

	.stripe-preview h4 {
		margin: 0 0 var(--space-sm) 0;
		color: var(--success);
		font-size: 1rem;
	}

	.preview-card {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.preview-item {
		font-size: 0.9rem;
		color: var(--text);
	}

	.preview-item strong {
		color: var(--text-emphasis);
	}

	.disabled-notice {
		color: var(--text-muted);
	}

	.disabled-notice p {
		margin: var(--space-xs) 0;
	}

	.disabled-notice a {
		color: var(--primary);
		text-decoration: none;
	}

	.disabled-notice a:hover {
		text-decoration: underline;
	}

	.loading-stripe {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		padding: var(--space-md);
		color: var(--text-muted);
	}

	.spinner {
		width: 16px;
		height: 16px;
		border: 2px solid var(--border);
		border-top: 2px solid var(--primary);
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-md);
		margin-top: var(--space-xl);
		padding-top: var(--space-lg);
		border-top: 1px solid var(--border);
	}

	.btn {
		padding: var(--space-md) var(--space-lg);
		border: none;
		border-radius: var(--radius-md);
		font-size: 1rem;
		cursor: pointer;
		transition: all 0.2s ease;
		font-weight: 600;
	}

	.btn-sm {
		padding: var(--space-xs) var(--space-sm);
		font-size: 0.9rem;
	}

	.btn-primary {
		background: var(--primary);
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: var(--primary-dark);
		transform: translateY(-1px);
	}

	.btn-primary:disabled {
		background: var(--border);
		color: var(--text-muted);
		cursor: not-allowed;
		transform: none;
	}

	.btn-secondary {
		background: var(--surface-secondary);
		color: var(--text);
		border: 1px solid var(--border);
	}

	.btn-secondary:hover {
		background: var(--bg-secondary);
		transform: translateY(-1px);
	}

	@media (max-width: 768px) {
		.modal-content {
			width: 95vw;
			padding: var(--space-lg);
		}

		.form-row {
			grid-template-columns: 1fr;
		}

		.modal-actions {
			flex-direction: column;
		}
	}
</style> 