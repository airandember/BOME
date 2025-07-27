<script lang="ts">
	import { fade, fly } from 'svelte/transition';
	import { createEventDispatcher } from 'svelte';
	import type { CreateSubscriptionPlanData } from '$lib/services/streaming-subscriptions';

	export let isOpen: boolean = false;
	export let title: string = '';
	export let formData: CreateSubscriptionPlanData;
	export let isSubmitting: boolean = false;
	export let mode: 'create' | 'edit' = 'create';

	const dispatch = createEventDispatcher();

	let newFeature = '';

	function handleSubmit() {
		dispatch('submit', { formData });
	}

	function handleCancel() {
		dispatch('cancel');
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

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			event.preventDefault();
			addFeature();
		}
	}
</script>

{#if isOpen}
	<div class="modal-backdrop" transition:fade={{ duration: 200 }} on:click={handleCancel}>
		<div class="modal-content" transition:fly={{ y: 20, duration: 200 }} on:click|stopPropagation>
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
					<div class="checkbox-group">
						<input
							id="is_promoted"
							type="checkbox"
							bind:checked={formData.is_promoted}
							class="form-checkbox"
						/>
						<label for="is_promoted" class="checkbox-label">Promoted</label>
					</div>
				</div>

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
		max-width: 600px;
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
	}
</style> 