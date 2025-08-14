<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { StandardizedRole } from '$lib/types/standardized_roles';

	export let isOpen = false;
	export let editingUser: any = null;
	export let roles: StandardizedRole[] = [];
	export let userForm = {
		firstName: '',
		lastName: '',
		email: '',
		role: '',
		roleId: '',
		emailVerified: false,
		isActive: true,
		hasSubbed: false,
		stripeCustomerId: ''
	};

	const dispatch = createEventDispatcher<{
		close: void;
		save: { userForm: typeof userForm; editingUser: any };
	}>();

	function closeModal() {
		dispatch('close');
	}

	function handleSave() {
		dispatch('save', { userForm, editingUser });
	}

	function handleSubmit(event: Event) {
		event.preventDefault();
		handleSave();
	}
</script>

{#if isOpen}
	<div class="modal-overlay" on:click={closeModal}>
		<div class="modal-content" on:click|stopPropagation>
			<div class="modal-header">
				<h2>{editingUser ? 'Edit User' : 'Add New User'}</h2>
				<button class="modal-close" on:click={closeModal}>×</button>
			</div>
			
			<div class="modal-body">
				<form on:submit={handleSubmit}>
					<div class="form-group">
						<label for="firstName">First Name *</label>
						<input
							id="firstName"
							type="text"
							bind:value={userForm.firstName}
							required
							class="form-input"
						/>
					</div>
					
					<div class="form-group">
						<label for="lastName">Last Name *</label>
						<input
							id="lastName"
							type="text"
							bind:value={userForm.lastName}
							required
							class="form-input"
						/>
					</div>
					
					<div class="form-group">
						<label for="email">Email *</label>
						<input
							id="email"
							type="email"
							bind:value={userForm.email}
							required
							class="form-input"
						/>
					</div>
					
					<div class="form-group">
						<label for="role">Role *</label>
						<select
							id="role"
							bind:value={userForm.role}
							required
							class="form-select"
						>
							<option value="">Select a role</option>
							{#each roles as role}
								<option value={role.id}>{role.name}</option>
							{/each}
						</select>
					</div>
					
					<div class="form-group">
						<label for="roleId">Role ID</label>
						<input
							id="roleId"
							type="text"
							bind:value={userForm.roleId}
							class="form-input"
							placeholder="Optional role ID"
						/>
					</div>
					
					<div class="form-group">
						<label for="stripeCustomerId">Stripe Customer ID</label>
						<input
							id="stripeCustomerId"
							type="text"
							bind:value={userForm.stripeCustomerId}
							class="form-input"
							placeholder="Optional Stripe customer ID"
						/>
					</div>
					
					<div class="form-group">
						<label class="checkbox-label">
							<input
								type="checkbox"
								bind:checked={userForm.emailVerified}
								class="form-checkbox"
							/>
							Email Verified
						</label>
					</div>
					
					<div class="form-group">
						<label class="checkbox-label">
							<input
								type="checkbox"
								bind:checked={userForm.isActive}
								class="form-checkbox"
							/>
							User Active
						</label>
					</div>
					
					<div class="form-group">
						<label class="checkbox-label">
							<input
								type="checkbox"
								bind:checked={userForm.hasSubbed}
								class="form-checkbox"
							/>
							Has Subscribed
						</label>
					</div>
					
					<div class="modal-actions">
						<button type="button" class="btn btn-secondary" on:click={closeModal}>
							Cancel
						</button>
						<button type="submit" class="btn btn-primary">
							{editingUser ? 'Update User' : 'Create User'}
						</button>
					</div>
				</form>
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
		padding: 1rem;
	}

	.modal-content {
		background: white;
		border-radius: 0.5rem;
		box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
		max-width: 500px;
		width: 100%;
		max-height: 90vh;
		overflow-y: auto;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1.5rem;
		border-bottom: 1px solid #e5e7eb;
	}

	.modal-header h2 {
		margin: 0;
		font-size: 1.25rem;
		font-weight: 600;
		color: #111827;
	}

	.modal-close {
		background: none;
		border: none;
		font-size: 1.5rem;
		color: #6b7280;
		cursor: pointer;
		padding: 0;
		width: 2rem;
		height: 2rem;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 0.25rem;
		transition: all 0.2s;
	}

	.modal-close:hover {
		background: #f3f4f6;
		color: #374151;
	}

	.modal-body {
		padding: 1.5rem;
	}

	.form-group {
		margin-bottom: 1rem;
	}

	.form-group label {
		display: block;
		margin-bottom: 0.5rem;
		font-weight: 500;
		color: #374151;
	}

	.form-input,
	.form-select {
		width: 100%;
		padding: 0.5rem;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		transition: border-color 0.2s;
	}

	.form-input:focus,
	.form-select:focus {
		outline: none;
		border-color: #2563eb;
		box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
	}

	.checkbox-label {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		cursor: pointer;
	}

	.form-checkbox {
		width: 1rem;
		height: 1rem;
		margin: 0;
	}

	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
		margin-top: 1.5rem;
		padding-top: 1rem;
		border-top: 1px solid #e5e7eb;
	}

	.btn {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		border: none;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
		text-decoration: none;
	}

	.btn-primary {
		background: #2563eb;
		color: white;
	}

	.btn-primary:hover {
		background: #1d4ed8;
	}

	.btn-secondary {
		background: #f3f4f6;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-secondary:hover {
		background: #e5e7eb;
	}

	@media (max-width: 768px) {
		.modal-content {
			margin: 1rem;
			max-height: calc(100vh - 2rem);
		}
	}
</style> 
