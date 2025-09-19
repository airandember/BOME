<script lang="ts">
	import type { EnhancedSubscriber } from '$lib/types/enhanced-subscriber';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import { apiRequest } from '$lib/auth';

	export let isOpen = false;
	export let subscriber: EnhancedSubscriber | null = null;
	export let onSave: (subscriber: EnhancedSubscriber) => void = () => {};
	export let onCancel: () => void = () => {};

	// Form state
	let isSubmitting = false;
	let formData: {
		id?: number;
		email?: string;
		first_name?: string;
		last_name?: string;
		role?: string;
		email_verified?: boolean;
		manual_access_granted?: boolean;
		account_status?: string;
	} = {};

	// Initialize form data when subscriber changes
	$: if (subscriber) {
		formData = {
			id: subscriber.id,
			email: subscriber.email,
			first_name: subscriber.first_name,
			last_name: subscriber.last_name,
			role: subscriber.role,
			email_verified: subscriber.email_verified,
			manual_access_granted: subscriber.manual_access_granted || false,
			account_status: subscriber.account_status
		};
	}

	// Available roles
	const roles = [
		{ value: 'user', label: 'User' },
		{ value: 'admin', label: 'Admin' },
		{ value: 'super_admin', label: 'Super Admin' }
	];

	// Available account statuses
	const accountStatuses = [
		{ value: 'active', label: 'Active' },
		{ value: 'suspended', label: 'Suspended' },
		{ value: 'pending', label: 'Pending' },
		{ value: 'inactive', label: 'Inactive' }
	];

	// Handle form submission
	async function handleSubmit(event: Event) {
		event.preventDefault();
		if (!subscriber || !formData.id) return;

		isSubmitting = true;
		try {
			// Prepare update data
			const updateData = {
				first_name: formData.first_name,
				last_name: formData.last_name,
				email: formData.email, // Include email even though it's readonly
				role: formData.role,
				email_verified: formData.email_verified,
				manual_video_access: formData.manual_access_granted,
				status: formData.account_status,
				notes: "" // Add empty notes field
			};

			// Call API to update subscriber
			const response = await apiRequest(`/admin/subscribers/${formData.id}`, {
				method: 'PUT',
				body: JSON.stringify(updateData)
			});

			if (response.ok) {
				const updatedSubscriber = await response.json();
				showToast('Subscriber updated successfully', 'success');
				
				// Create updated subscriber object
				const updated: EnhancedSubscriber = {
					...subscriber,
					first_name: formData.first_name || subscriber.first_name,
					last_name: formData.last_name || subscriber.last_name,
					role: formData.role || subscriber.role,
					email_verified: formData.email_verified ?? subscriber.email_verified,
					manual_access_granted: formData.manual_access_granted ?? subscriber.manual_access_granted,
					account_status: formData.account_status || subscriber.account_status,
					full_name: `${formData.first_name || subscriber.first_name} ${formData.last_name || subscriber.last_name}`.trim()
				};
				
				onSave(updated);
			} else {
				const error = await response.json();
				throw new Error(error.message || 'Failed to update subscriber');
			}
		} catch (error: any) {
			console.error('Error updating subscriber:', error);
			showToast(`Failed to update subscriber: ${error.message}`, 'error');
		} finally {
			isSubmitting = false;
		}
	}

	// Handle manual access toggle
	async function toggleManualAccess() {
		if (!subscriber) return;

		const newValue = !formData.manual_access_granted;
		
		try {
			const response = await apiRequest(`/admin/subscribers/${subscriber.id}/manual-access`, {
				method: 'POST',
				body: JSON.stringify({ manual_access: newValue })
			});

			if (response.ok) {
				formData.manual_access_granted = newValue;
				showToast(`Manual access override ${newValue ? 'enabled' : 'disabled'}`, 'success');
			} else {
				const error = await response.json();
				throw new Error(error.message || 'Failed to update manual access');
			}
		} catch (error: any) {
			console.error('Error updating manual access:', error);
			showToast(`Failed to update manual access: ${error.message}`, 'error');
		}
	}

	// Close modal on escape key
	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' && !isSubmitting) {
			onCancel();
		}
	}

	// Close modal when clicking outside
	function handleBackdropClick(event: MouseEvent) {
		if (event.target === event.currentTarget && !isSubmitting) {
			onCancel();
		}
	}

	// Validate form
	$: isFormValid = formData.first_name?.trim() && formData.last_name?.trim() && formData.email?.trim() && formData.role;
</script>

<svelte:window onkeydown={handleKeydown} />

{#if isOpen && subscriber}
	<div class="modal-overlay" onclick={handleBackdropClick} role="dialog" aria-modal="true" tabindex="-1">
		<div class="modal-content">
			<!-- Header -->
			<div class="modal-header">
				<div class="header-info">
					<h2>✏️ Edit Subscriber</h2>
					<p>Update subscriber information and access permissions</p>
				</div>
				<button 
					type="button" 
					class="close-btn" 
					onclick={onCancel}
					disabled={isSubmitting}
					aria-label="Close modal"
				>
					✕
				</button>
			</div>

			<!-- Content -->
			<form onsubmit={handleSubmit} class="modal-body">
				<!-- Basic Information -->
				<section class="form-section">
					<h3>📋 Basic Information</h3>
					<div class="form-grid">
						<div class="form-group">
							<label for="first_name">First Name *</label>
							<input
								id="first_name"
								type="text"
								bind:value={formData.first_name}
								required
								disabled={isSubmitting}
								class="form-input"
								placeholder="Enter first name"
							/>
						</div>

						<div class="form-group">
							<label for="last_name">Last Name *</label>
							<input
								id="last_name"
								type="text"
								bind:value={formData.last_name}
								required
								disabled={isSubmitting}
								class="form-input"
								placeholder="Enter last name"
							/>
						</div>

						<div class="form-group">
							<label for="email">Email Address</label>
							<input
								id="email"
								type="email"
								value={formData.email}
								disabled
								class="form-input disabled"
								title="Email cannot be changed"
							/>
							<span class="form-help">Email address cannot be modified</span>
						</div>

						<div class="form-group">
							<label for="role">Role *</label>
							<select
								id="role"
								bind:value={formData.role}
								required
								disabled={isSubmitting}
								class="form-select"
							>
								<option value="">Select a role</option>
								{#each roles as role}
									<option value={role.value}>{role.label}</option>
								{/each}
							</select>
						</div>
					</div>
				</section>

				<!-- Account Status -->
				<section class="form-section">
					<h3>🔐 Account Status & Access</h3>
					<div class="form-grid">
						<div class="form-group">
							<label for="account_status">Account Status</label>
							<select
								id="account_status"
								bind:value={formData.account_status}
								disabled={isSubmitting}
								class="form-select"
							>
								{#each accountStatuses as status}
									<option value={status.value}>{status.label}</option>
								{/each}
							</select>
						</div>

						<div class="form-group">
							<label class="checkbox-label">
								<input
									type="checkbox"
									bind:checked={formData.email_verified}
									disabled={isSubmitting}
									class="form-checkbox"
								/>
								<span class="checkbox-text">Email Verified</span>
							</label>
							<span class="form-help">Whether the user's email address has been verified</span>
						</div>
					</div>
				</section>

				<!-- Manual Access Control -->
				<section class="form-section">
					<h3>🎥 Video Access Control</h3>
					<div class="access-control">
						<div class="access-info">
							<h4>Manual Video Access</h4>
							<p>Grant or revoke video access independent of subscription plans</p>
							<div class="current-access">
								<span class="access-label">Manual Override:</span>
								<span class="access-value {formData.manual_access_granted ? 'granted' : 'denied'}">
									{formData.manual_access_granted ? '✅ Active' : '❌ Off'}
								</span>
							</div>
						</div>
						<button
							type="button"
							class="access-toggle-btn {formData.manual_access_granted ? 'revoke' : 'grant'}"
							onclick={toggleManualAccess}
							disabled={isSubmitting}
						>
							{formData.manual_access_granted ? '🚫 Turn Off Override' : '✅ Enable Override'}
						</button>
					</div>
				</section>

				<!-- Read-only Information -->
				<section class="form-section">
					<h3>📊 Current Plan Information</h3>
					<div class="readonly-info">
						<div class="info-grid">
							<div class="info-item">
								<div class="info-label">Plan Name</div>
								<span class="value">{subscriber.plan_name || 'No Plan'}</span>
							</div>
							<div class="info-item">
								<div class="info-label">Plan Type</div>
								<span class="value plan-type-{subscriber.plan_type}">{subscriber.plan_type}</span>
							</div>
							<div class="info-item">
								<div class="info-label">Active Plan</div>
								<span class="value badge {subscriber.has_active_plan ? 'badge-success' : 'badge-error'}">
									{subscriber.has_active_plan ? '✅ Yes' : '❌ No'}
								</span>
							</div>
							<div class="info-item">
								<div class="info-label">Video Access</div>
								<span class="value badge {subscriber.has_video_access ? 'badge-success' : 'badge-error'}">
									{subscriber.has_video_access ? '✅ Yes' : '❌ No'}
								</span>
							</div>
						</div>
						<p class="readonly-note">
							💡 Plan information is managed through Stripe and cannot be edited here. 
							Use the manual access controls above to override subscription-based access.
						</p>
					</div>
				</section>
			</form>

			<!-- Footer -->
			<div class="modal-footer">
				<button 
					type="button" 
					class="btn btn-secondary" 
					onclick={onCancel}
					disabled={isSubmitting}
				>
					Cancel
				</button>
				<button 
					type="button" 
					class="btn btn-primary" 
					onclick={handleSubmit}
					disabled={isSubmitting || !isFormValid}
				>
					{#if isSubmitting}
						<LoadingSpinner size="small" />
						Saving...
					{:else}
						💾 Save Changes
					{/if}
				</button>
			</div>
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
		backdrop-filter: blur(4px);
	}

	.modal-content {
		background: white;
		border-radius: 1rem;
		box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
		max-width: 85vw;
		width: 90%;
		max-height: 95vh;
		overflow-y: auto;
		border: 1px solid #e5e7eb;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		padding: 1.5rem;
		border-bottom: 1px solid #e5e7eb;
		background: #f9fafb;
		border-radius: 1rem 1rem 0 0;
	}

	.header-info h2 {
		margin: 0 0 0.25rem 0;
		color: #111827;
		font-size: 1.5rem;
		font-weight: 600;
	}

	.header-info p {
		margin: 0;
		color: #6b7280;
		font-size: 0.875rem;
	}

	.close-btn {
		background: none;
		border: none;
		font-size: 1.5rem;
		cursor: pointer;
		color: #6b7280;
		padding: 0;
		width: 32px;
		height: 32px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 0.375rem;
		transition: all 0.2s ease;
	}

	.close-btn:hover:not(:disabled) {
		background: #e5e7eb;
		color: #111827;
	}

	.close-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.modal-body {
		padding: 1.5rem;
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.form-section {
		border: 1px solid #e5e7eb;
		border-radius: 0.75rem;
		overflow: hidden;
	}

	.form-section h3 {
		margin: 0;
		padding: 1rem 1.5rem;
		background: #f9fafb;
		border-bottom: 1px solid #e5e7eb;
		font-size: 1.125rem;
		font-weight: 600;
		color: #111827;
	}

	.form-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 1rem;
		padding: 1.5rem;
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.form-group label {
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
	}

	.form-input, .form-select {
		padding: 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 0.5rem;
		font-size: 1rem;
		color: #111827;
		background: white;
		transition: all 0.2s ease;
	}

	.form-input:focus, .form-select:focus {
		outline: none;
		border-color: #3b82f6;
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	.form-input.disabled {
		background: #f9fafb;
		color: #6b7280;
		cursor: not-allowed;
	}

	.form-input:disabled, .form-select:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.checkbox-label {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		cursor: pointer;
	}

	.form-checkbox {
		width: 1.25rem;
		height: 1.25rem;
		cursor: pointer;
	}

	.checkbox-text {
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
	}

	.form-help {
		font-size: 0.75rem;
		color: #6b7280;
		font-style: italic;
	}

	/* Access Control */
	.access-control {
		padding: 1.5rem;
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 1rem;
	}

	.access-info h4 {
		margin: 0 0 0.5rem 0;
		font-size: 1rem;
		font-weight: 600;
		color: #111827;
	}

	.access-info p {
		margin: 0 0 1rem 0;
		font-size: 0.875rem;
		color: #6b7280;
	}

	.current-access {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.access-label {
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
	}

	.access-value {
		font-size: 0.875rem;
		font-weight: 600;
		padding: 0.25rem 0.75rem;
		border-radius: 9999px;
	}

	.access-value.granted {
		background: #dcfce7;
		color: #166534;
	}

	.access-value.denied {
		background: #fee2e2;
		color: #991b1b;
	}

	.access-toggle-btn {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 0.5rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s ease;
		white-space: nowrap;
	}

	.access-toggle-btn.grant {
		background: #dcfce7;
		color: #166534;
		border: 1px solid #bbf7d0;
	}

	.access-toggle-btn.grant:hover:not(:disabled) {
		background: #bbf7d0;
	}

	.access-toggle-btn.revoke {
		background: #fee2e2;
		color: #991b1b;
		border: 1px solid #fecaca;
	}

	.access-toggle-btn.revoke:hover:not(:disabled) {
		background: #fecaca;
	}

	.access-toggle-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	/* Read-only Information */
	.readonly-info {
		padding: 1.5rem;
	}

	.info-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.info-item {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.info-item .info-label {
		font-size: 0.75rem;
		font-weight: 500;
		color: #6b7280;
		text-transform: uppercase;
		letter-spacing: 0.025em;
	}

	.info-item .value {
		font-size: 0.875rem;
		color: #111827;
		font-weight: 500;
	}

	.readonly-note {
		margin: 0;
		padding: 1rem;
		background: #f0f9ff;
		border: 1px solid #bae6fd;
		border-radius: 0.5rem;
		font-size: 0.875rem;
		color: #0c4a6e;
	}

	/* Badge styles */
	.badge {
		display: inline-block;
		padding: 0.25rem 0.75rem;
		border-radius: 9999px;
		font-size: 0.75rem;
		font-weight: 500;
		text-align: center;
	}

	.badge-success {
		background: #dcfce7;
		color: #166534;
	}

	.badge-error {
		background: #fee2e2;
		color: #991b1b;
	}

	/* Plan type styles */
	.plan-type-premium {
		color: #7c3aed;
		font-weight: 600;
		text-transform: capitalize;
	}

	.plan-type-basic {
		color: #059669;
		font-weight: 600;
		text-transform: capitalize;
	}

	.plan-type-none {
		color: #6b7280;
		font-style: italic;
		text-transform: capitalize;
	}

	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: 1rem;
		padding: 1.5rem;
		border-top: 1px solid #e5e7eb;
		background: #f9fafb;
		border-radius: 0 0 1rem 1rem;
	}

	.btn {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 0.5rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s ease;
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.btn-secondary {
		background: #f3f4f6;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-secondary:hover:not(:disabled) {
		background: #e5e7eb;
		border-color: #9ca3af;
	}

	.btn-primary {
		background: #3b82f6;
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: #2563eb;
	}

	.btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	/* Responsive design */
	@media (max-width: 768px) {
		.modal-content {
			width: 95%;
			margin: 1rem;
		}

		.form-grid, .info-grid {
			grid-template-columns: 1fr;
		}

		.access-control {
			flex-direction: column;
			align-items: stretch;
			text-align: center;
		}

		.modal-header {
			flex-direction: column;
			gap: 1rem;
			align-items: stretch;
		}
	}
</style>
