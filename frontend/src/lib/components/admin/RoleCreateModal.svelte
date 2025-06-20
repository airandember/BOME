<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { Role } from '$lib/types/roles';

	export let isOpen = false;
	export let editingRole: Role | null = null;

	const dispatch = createEventDispatcher();

	// Simple form data
	let formData = {
		name: '',
		description: '',
		category: 'content',
		level: 5,
		color: '#6366f1',
		icon: 'users'
	};

	let errors: { [key: string]: string } = {};
	let isSubmitting = false;

	// Simple categories
	const roleCategories = [
		{ id: 'core', label: 'Core Admin' },
		{ id: 'content', label: 'Content & Editorial' },
		{ id: 'marketing', label: 'Marketing & Advertising' },
		{ id: 'events', label: 'Events & Community' },
		{ id: 'technical', label: 'Technical & Support' },
		{ id: 'academic', label: 'Academic' }
	];

	// Simple icons
	const availableIcons = [
		{ id: 'users', emoji: '👥' },
		{ id: 'document-text', emoji: '📄' },
		{ id: 'video-camera', emoji: '📹' },
		{ id: 'shield-check', emoji: '🛡️' },
		{ id: 'chart-bar', emoji: '📊' }
	];

	// Simple colors
	const availableColors = [
		'#6366f1', '#8b5cf6', '#ec4899', '#ef4444', '#f97316',
		'#22c55e', '#10b981', '#06b6d4', '#3b82f6'
	];

	$: if (editingRole) {
		formData = {
			name: editingRole.name,
			description: editingRole.description,
			category: editingRole.category,
			level: editingRole.level,
			color: editingRole.color,
			icon: editingRole.icon
		};
	}

	function validateForm(): boolean {
		errors = {};

		if (!formData.name.trim()) {
			errors.name = 'Role name is required';
		}

		if (!formData.description.trim()) {
			errors.description = 'Role description is required';
		}

		return Object.keys(errors).length === 0;
	}

	async function handleSubmit() {
		if (!validateForm()) return;

		isSubmitting = true;

		try {
			const roleData = {
				...formData,
				id: editingRole?.id || formData.name.toLowerCase().replace(/\s+/g, '-'),
				slug: formData.name.toLowerCase().replace(/\s+/g, '-'),
				permissions: editingRole?.permissions || [],
				isSystemRole: editingRole?.isSystemRole || false,
				createdAt: editingRole?.createdAt || new Date().toISOString(),
				updatedAt: new Date().toISOString()
			};

			// Simulate API call
			await new Promise(resolve => setTimeout(resolve, 500));

			dispatch(editingRole ? 'roleUpdated' : 'roleCreated', roleData);
			closeModal();
		} catch (error) {
			console.error('Error saving role:', error);
		} finally {
			isSubmitting = false;
		}
	}

	function closeModal() {
		isOpen = false;
		editingRole = null;
		formData = {
			name: '',
			description: '',
			category: 'content',
			level: 5,
			color: '#6366f1',
			icon: 'users'
		};
		errors = {};
		dispatch('close');
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			closeModal();
		}
	}
</script>

<svelte:window on:keydown={handleKeydown} />

{#if isOpen}
	<div class="modal-overlay" on:click={closeModal} role="dialog" aria-modal="true" aria-labelledby="modal-title">
		<div class="modal-container" on:click|stopPropagation role="document">
			<div class="modal-header">
				<h2 id="modal-title" class="modal-title">
					{editingRole ? 'Edit Role' : 'Create New Role'}
				</h2>
				<button class="modal-close" on:click={closeModal} aria-label="Close modal">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"></line>
						<line x1="6" y1="6" x2="18" y2="18"></line>
					</svg>
				</button>
			</div>

			<div class="modal-content">
				<form on:submit|preventDefault={handleSubmit}>
					<div class="form-group">
						<label for="name" class="form-label">Role Name</label>
						<input
							id="name"
							type="text"
							bind:value={formData.name}
							class="form-input"
							class:error={errors.name}
							placeholder="Enter role name"
							required
						/>
						{#if errors.name}
							<span class="error-message">{errors.name}</span>
						{/if}
					</div>

					<div class="form-group">
						<label for="description" class="form-label">Description</label>
						<textarea
							id="description"
							bind:value={formData.description}
							class="form-textarea"
							class:error={errors.description}
							placeholder="Describe the role's responsibilities"
							rows="3"
							required
						></textarea>
						{#if errors.description}
							<span class="error-message">{errors.description}</span>
						{/if}
					</div>

					<div class="form-grid">
						<div class="form-group">
							<label for="category" class="form-label">Category</label>
							<select id="category" bind:value={formData.category} class="form-select">
								{#each roleCategories as category}
									<option value={category.id}>{category.label}</option>
								{/each}
							</select>
						</div>

						<div class="form-group">
							<label for="level" class="form-label">Access Level (1-10)</label>
							<input
								id="level"
								type="number"
								min="1"
								max="10"
								bind:value={formData.level}
								class="form-input"
							/>
						</div>
					</div>

					<div class="form-grid">
						<div class="form-group">
							<fieldset>
								<legend class="form-label">Icon</legend>
								<div class="icon-grid" role="radiogroup" aria-label="Select role icon">
									{#each availableIcons as icon}
										<button
											type="button"
											class="icon-option"
											class:selected={formData.icon === icon.id}
											on:click={() => formData.icon = icon.id}
											role="radio"
											aria-checked={formData.icon === icon.id}
											aria-label="Select {icon.id} icon"
										>
											{icon.emoji}
										</button>
									{/each}
								</div>
							</fieldset>
						</div>

						<div class="form-group">
							<fieldset>
								<legend class="form-label">Color</legend>
								<div class="color-grid" role="radiogroup" aria-label="Select role color">
									{#each availableColors as color}
										<button
											type="button"
											class="color-option"
											class:selected={formData.color === color}
											style="background-color: {color}"
											on:click={() => formData.color = color}
											role="radio"
											aria-checked={formData.color === color}
											aria-label="Select color {color}"
										>
										</button>
									{/each}
								</div>
							</fieldset>
						</div>
					</div>

					<div class="modal-footer">
						<button type="button" class="btn btn-secondary" on:click={closeModal}>
							Cancel
						</button>
						<button type="submit" class="btn btn-primary" disabled={isSubmitting}>
							{#if isSubmitting}
								Saving...
							{:else}
								{editingRole ? 'Update' : 'Create'} Role
							{/if}
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
		background: rgba(0, 0, 0, 0.8);
		backdrop-filter: blur(8px);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
		padding: var(--space-lg);
	}

	.modal-container {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
		width: 100%;
		max-width: 900px;
		max-height: 90vh;
		overflow: hidden;
		box-shadow: var(--shadow-xl);
		backdrop-filter: blur(10px);
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-xl);
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
		background: var(--bg-glass-dark);
	}

	.modal-title {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.modal-close {
		background: var(--bg-glass);
		border: 1px solid rgba(255, 255, 255, 0.1);
		color: var(--text-primary);
		cursor: pointer;
		padding: var(--space-md);
		border-radius: var(--radius-lg);
		transition: all var(--transition-normal);
		width: 40px;
		height: 40px;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.modal-close:hover {
		background: var(--bg-glass-dark);
		color: var(--text-primary);
		transform: scale(1.05);
	}

	.modal-close svg {
		width: 20px;
		height: 20px;
	}

	.modal-content {
		padding: var(--space-xl);
		overflow-y: auto;
		max-height: calc(90vh - 120px);
	}

	.form-group {
		margin-bottom: var(--space-lg);
	}

	.form-label {
		display: block;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-md);
		font-size: var(--text-sm);
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.form-input,
	.form-textarea,
	.form-select {
		width: 100%;
		padding: var(--space-lg);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-xl);
		background: var(--bg-glass);
		color: var(--text-primary);
		font-size: var(--text-sm);
		transition: all var(--transition-normal);
		backdrop-filter: blur(10px);
	}

	.form-input:focus,
	.form-textarea:focus,
	.form-select:focus {
		outline: none;
		border-color: var(--primary);
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
		background: var(--bg-glass-dark);
	}

	.form-input::placeholder,
	.form-textarea::placeholder {
		color: var(--text-secondary);
	}

	.form-input.error,
	.form-textarea.error {
		border-color: var(--error);
		box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1);
	}

	.form-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: var(--space-lg);
	}

	fieldset {
		border: none;
		padding: 0;
		margin: 0;
	}

	legend {
		display: block;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-md);
		font-size: var(--text-sm);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		padding: 0;
	}

	.icon-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(50px, 1fr));
		gap: var(--space-md);
		margin-top: var(--space-md);
	}

	.icon-option {
		width: 50px;
		height: 50px;
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-lg);
		background: var(--bg-glass);
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: var(--text-xl);
		transition: all var(--transition-normal);
		backdrop-filter: blur(10px);
	}

	.icon-option:hover {
		border-color: var(--primary);
		background: var(--bg-glass-dark);
		transform: translateY(-2px);
		box-shadow: var(--shadow-md);
	}

	.icon-option.selected {
		border-color: var(--primary);
		background: var(--primary-gradient);
		color: var(--white);
		box-shadow: var(--shadow-lg);
	}

	.color-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(40px, 1fr));
		gap: var(--space-md);
		margin-top: var(--space-md);
	}

	.color-option {
		width: 40px;
		height: 40px;
		border: 2px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-lg);
		cursor: pointer;
		transition: all var(--transition-normal);
		position: relative;
	}

	.color-option:hover {
		transform: scale(1.1);
		box-shadow: var(--shadow-md);
	}

	.color-option.selected {
		border-color: var(--white);
		box-shadow: 0 0 0 3px rgba(255, 255, 255, 0.3);
		transform: scale(1.1);
	}

	.color-option.selected::after {
		content: '✓';
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		color: var(--white);
		font-weight: bold;
		font-size: var(--text-sm);
		text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
	}

	.error-message {
		color: var(--error);
		font-size: var(--text-sm);
		margin-top: var(--space-sm);
		display: block;
		font-weight: 500;
	}

	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-lg);
		padding-top: var(--space-xl);
		border-top: 1px solid rgba(255, 255, 255, 0.1);
		margin-top: var(--space-xl);
	}

	.btn {
		padding: var(--space-lg) var(--space-2xl);
		border: none;
		border-radius: var(--radius-xl);
		font-size: var(--text-sm);
		font-weight: 600;
		cursor: pointer;
		transition: all var(--transition-normal);
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		text-decoration: none;
		justify-content: center;
		min-width: 120px;
	}

	.btn-primary {
		background: var(--primary-gradient);
		color: var(--white);
		box-shadow: var(--shadow-md);
	}

	.btn-primary:hover:not(:disabled) {
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}

	.btn-primary:disabled {
		opacity: 0.6;
		cursor: not-allowed;
		transform: none;
		box-shadow: var(--shadow-sm);
	}

	.btn-secondary {
		background: var(--bg-glass);
		color: var(--text-primary);
		border: 1px solid rgba(255, 255, 255, 0.1);
		backdrop-filter: blur(10px);
	}

	.btn-secondary:hover {
		background: var(--bg-glass-dark);
		border-color: var(--primary);
		transform: translateY(-1px);
		box-shadow: var(--shadow-md);
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.modal-overlay {
			padding: var(--space-md);
		}

		.modal-container {
			max-height: 95vh;
		}

		.modal-header,
		.modal-content {
			padding: var(--space-lg);
		}

		.form-grid {
			grid-template-columns: 1fr;
		}

		.icon-grid {
			grid-template-columns: repeat(auto-fill, minmax(45px, 1fr));
		}

		.color-grid {
			grid-template-columns: repeat(auto-fill, minmax(35px, 1fr));
		}

		.modal-footer {
			flex-direction: column;
			gap: var(--space-md);
		}

		.btn {
			width: 100%;
		}
	}

	@media (max-width: 480px) {
		.modal-overlay {
			padding: var(--space-sm);
		}

		.modal-header,
		.modal-content {
			padding: var(--space-md);
		}

		.icon-option {
			width: 40px;
			height: 40px;
		}

		.color-option {
			width: 30px;
			height: 30px;
		}
	}
</style> 