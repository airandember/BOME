<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	interface ApiKey {
		id: string;
		name: string;
		key: string;
		description?: string;
		permissions: string[];
		created_at: string;
		last_used?: string;
		expires_at?: string;
		status: 'active' | 'inactive' | 'expired';
		usage_count: number;
		rate_limit: number;
		created_by: string;
	}

	let apiKeys: ApiKey[] = [];
	let loading = true;
	let error = '';
	let showCreateModal = false;
	let showEditModal = false;
	let selectedKey: ApiKey | null = null;
	let searchQuery = '';
	let selectedStatus = 'all';

	// Create/Edit form data
	let formData = {
		name: '',
		description: '',
		permissions: [] as string[],
		expires_at: '',
		rate_limit: 1000
	};

	const permissionOptions = [
		{ value: 'videos.read', label: 'Read Videos' },
		{ value: 'videos.write', label: 'Write Videos' },
		{ value: 'users.read', label: 'Read Users' },
		{ value: 'users.write', label: 'Write Users' },
		{ value: 'analytics.read', label: 'Read Analytics' },
		{ value: 'payments.read', label: 'Read Payments' },
		{ value: 'admin.read', label: 'Admin Read' },
		{ value: 'admin.write', label: 'Admin Write' }
	];

	const statusOptions = [
		{ value: 'all', label: 'All Status' },
		{ value: 'active', label: 'Active' },
		{ value: 'inactive', label: 'Inactive' },
		{ value: 'expired', label: 'Expired' }
	];

	onMount(async () => {
		await loadApiKeys();
	});

	const loadApiKeys = async () => {
		try {
			loading = true;
			error = '';
			
			// Mock data - replace with actual API call
			await new Promise(resolve => setTimeout(resolve, 1000));
			
			const mockKeys: ApiKey[] = [
				{
					id: 'key_001',
					name: 'Mobile App API',
					key: 'bome_live_1234567890abcdef',
					description: 'API key for mobile application access',
					permissions: ['videos.read', 'users.read', 'analytics.read'],
					created_at: '2024-06-01T10:00:00Z',
					last_used: '2024-06-18T09:30:00Z',
					expires_at: '2024-12-31T23:59:59Z',
					status: 'active',
					usage_count: 15678,
					rate_limit: 5000,
					created_by: 'admin@bome.com'
				},
				{
					id: 'key_002',
					name: 'Analytics Dashboard',
					key: 'bome_live_abcdef1234567890',
					description: 'API key for analytics dashboard integration',
					permissions: ['analytics.read', 'videos.read'],
					created_at: '2024-05-15T14:30:00Z',
					last_used: '2024-06-18T10:15:00Z',
					status: 'active',
					usage_count: 8234,
					rate_limit: 2000,
					created_by: 'admin@bome.com'
				},
				{
					id: 'key_003',
					name: 'Third Party Integration',
					key: 'bome_live_fedcba0987654321',
					description: 'API key for external service integration',
					permissions: ['videos.read', 'users.read'],
					created_at: '2024-04-20T16:45:00Z',
					last_used: '2024-06-10T08:20:00Z',
					expires_at: '2024-06-30T23:59:59Z',
					status: 'expired',
					usage_count: 3456,
					rate_limit: 1000,
					created_by: 'developer@bome.com'
				},
				{
					id: 'key_004',
					name: 'Testing Environment',
					key: 'bome_test_1122334455667788',
					description: 'API key for testing and development',
					permissions: ['videos.read', 'videos.write', 'users.read'],
					created_at: '2024-06-10T12:00:00Z',
					status: 'inactive',
					usage_count: 234,
					rate_limit: 500,
					created_by: 'developer@bome.com'
				}
			];

			// Apply filters
			let filteredKeys = mockKeys;
			
			if (selectedStatus !== 'all') {
				filteredKeys = filteredKeys.filter(key => key.status === selectedStatus);
			}
			
			if (searchQuery.trim()) {
				const query = searchQuery.toLowerCase();
				filteredKeys = filteredKeys.filter(key => 
					key.name.toLowerCase().includes(query) ||
					key.description?.toLowerCase().includes(query) ||
					key.created_by.toLowerCase().includes(query)
				);
			}

			apiKeys = filteredKeys;
		} catch (err) {
			error = 'Failed to load API keys';
			console.error('Error loading API keys:', err);
		} finally {
			loading = false;
		}
	};

	const handleFilterChange = async () => {
		await loadApiKeys();
	};

	const formatDate = (dateString: string): string => {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	};

	const getStatusBadge = (status: string) => {
		const statusConfig = {
			active: { text: 'Active', class: 'status-active' },
			inactive: { text: 'Inactive', class: 'status-inactive' },
			expired: { text: 'Expired', class: 'status-expired' }
		};

		const config = statusConfig[status as keyof typeof statusConfig] || { text: status, class: 'status-unknown' };
		
		return `<span class="status-badge ${config.class}">${config.text}</span>`;
	};

	const maskApiKey = (key: string): string => {
		if (key.length <= 8) return key;
		return key.substring(0, 8) + '•'.repeat(key.length - 12) + key.substring(key.length - 4);
	};

	const copyApiKey = async (key: string) => {
		try {
			await navigator.clipboard.writeText(key);
			showToast('API key copied to clipboard', 'success');
		} catch (err) {
			showToast('Failed to copy API key', 'error');
		}
	};

	const resetFormData = () => {
		formData = {
			name: '',
			description: '',
			permissions: [],
			expires_at: '',
			rate_limit: 1000
		};
	};

	const openCreateModal = () => {
		resetFormData();
		showCreateModal = true;
	};

	const openEditModal = (key: ApiKey) => {
		selectedKey = key;
		formData = {
			name: key.name,
			description: key.description || '',
			permissions: [...key.permissions],
			expires_at: key.expires_at ? key.expires_at.split('T')[0] : '',
			rate_limit: key.rate_limit
		};
		showEditModal = true;
	};

	const closeModals = () => {
		showCreateModal = false;
		showEditModal = false;
		selectedKey = null;
		resetFormData();
	};

	const handlePermissionChange = (permission: string, checked: boolean) => {
		if (checked) {
			formData.permissions = [...formData.permissions, permission];
		} else {
			formData.permissions = formData.permissions.filter(p => p !== permission);
		}
	};

	const createApiKey = async () => {
		try {
			if (!formData.name.trim()) {
				showToast('Please enter a name for the API key', 'error');
				return;
			}

			if (formData.permissions.length === 0) {
				showToast('Please select at least one permission', 'error');
				return;
			}

			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 1000));
			
			showToast('API key created successfully', 'success');
			closeModals();
			await loadApiKeys();
		} catch (err) {
			showToast('Failed to create API key', 'error');
		}
	};

	const updateApiKey = async () => {
		try {
			if (!formData.name.trim()) {
				showToast('Please enter a name for the API key', 'error');
				return;
			}

			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 1000));
			
			showToast('API key updated successfully', 'success');
			closeModals();
			await loadApiKeys();
		} catch (err) {
			showToast('Failed to update API key', 'error');
		}
	};

	const toggleKeyStatus = async (key: ApiKey) => {
		try {
			const newStatus = key.status === 'active' ? 'inactive' : 'active';
			
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 500));
			
			showToast(`API key ${newStatus === 'active' ? 'activated' : 'deactivated'}`, 'success');
			await loadApiKeys();
		} catch (err) {
			showToast('Failed to update API key status', 'error');
		}
	};

	const deleteApiKey = async (key: ApiKey) => {
		if (!confirm(`Are you sure you want to delete the API key "${key.name}"? This action cannot be undone.`)) {
			return;
		}

		try {
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 500));
			
			showToast('API key deleted successfully', 'success');
			await loadApiKeys();
		} catch (err) {
			showToast('Failed to delete API key', 'error');
		}
	};

	const regenerateApiKey = async (key: ApiKey) => {
		if (!confirm(`Are you sure you want to regenerate the API key "${key.name}"? The old key will no longer work.`)) {
			return;
		}

		try {
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 1000));
			
			showToast('API key regenerated successfully', 'success');
			await loadApiKeys();
		} catch (err) {
			showToast('Failed to regenerate API key', 'error');
		}
	};
</script>

<svelte:head>
	<title>API Keys Management - Admin Dashboard</title>
</svelte:head>

<div class="api-keys-page">
	<div class="page-header">
		<div class="header-content">
			<div class="header-text">
				<h1>API Keys Management</h1>
				<p>Manage API keys and access permissions</p>
			</div>
			
			<div class="header-actions">
				<button class="btn btn-primary" on:click={openCreateModal}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="12" y1="5" x2="12" y2="19"></line>
						<line x1="5" y1="12" x2="19" y2="12"></line>
					</svg>
					Create API Key
				</button>
				<button class="btn btn-outline" on:click={() => goto('/admin/security')}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M19 12H5"></path>
						<path d="M12 19l-7-7 7-7"></path>
					</svg>
					Back to Security
				</button>
			</div>
		</div>
	</div>

	<!-- Filters -->
	<div class="filters-section glass">
		<div class="filters-grid">
			<div class="filter-group">
				<label for="search">Search</label>
				<input
					id="search"
					type="text"
					placeholder="Search by name, description, or creator..."
					bind:value={searchQuery}
					on:input={handleFilterChange}
					class="filter-input"
				/>
			</div>

			<div class="filter-group">
				<label for="status">Status</label>
				<select id="status" bind:value={selectedStatus} on:change={handleFilterChange} class="filter-select">
					{#each statusOptions as option}
						<option value={option.value}>{option.label}</option>
					{/each}
				</select>
			</div>
		</div>
	</div>

	{#if loading}
		<div class="loading-container">
			<LoadingSpinner size="large" color="primary" />
			<p>Loading API keys...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<p class="error-message">{error}</p>
			<button class="btn btn-primary" on:click={loadApiKeys}>
				Try Again
			</button>
		</div>
	{:else}
		<!-- API Keys Table -->
		<div class="keys-table glass">
			<div class="table-header">
				<div class="header-cell">Name</div>
				<div class="header-cell">API Key</div>
				<div class="header-cell">Permissions</div>
				<div class="header-cell">Usage</div>
				<div class="header-cell">Status</div>
				<div class="header-cell">Last Used</div>
				<div class="header-cell">Actions</div>
			</div>

			{#each apiKeys as key}
				<div class="table-row">
					<div class="table-cell">
						<div class="key-info">
							<span class="key-name">{key.name}</span>
							{#if key.description}
								<span class="key-description">{key.description}</span>
							{/if}
							<span class="key-creator">Created by {key.created_by}</span>
						</div>
					</div>
					<div class="table-cell">
						<div class="api-key-display">
							<code class="api-key">{maskApiKey(key.key)}</code>
							<button class="copy-btn" on:click={() => copyApiKey(key.key)} title="Copy API key">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
									<path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
								</svg>
							</button>
						</div>
					</div>
					<div class="table-cell">
						<div class="permissions-list">
							{#each key.permissions.slice(0, 2) as permission}
								<span class="permission-badge">{permission}</span>
							{/each}
							{#if key.permissions.length > 2}
								<span class="permission-more">+{key.permissions.length - 2} more</span>
							{/if}
						</div>
					</div>
					<div class="table-cell">
						<div class="usage-info">
							<span class="usage-count">{key.usage_count.toLocaleString()}</span>
							<span class="rate-limit">Limit: {key.rate_limit}/hr</span>
						</div>
					</div>
					<div class="table-cell">
						{@html getStatusBadge(key.status)}
					</div>
					<div class="table-cell">
						<span class="last-used">
							{key.last_used ? formatDate(key.last_used) : 'Never'}
						</span>
					</div>
					<div class="table-cell">
						<div class="actions-dropdown">
							<button class="btn btn-ghost btn-small">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<circle cx="12" cy="12" r="1"></circle>
									<circle cx="12" cy="5" r="1"></circle>
									<circle cx="12" cy="19" r="1"></circle>
								</svg>
							</button>
							<div class="dropdown-menu">
								<button class="dropdown-item" on:click={() => openEditModal(key)}>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
										<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
									</svg>
									Edit
								</button>
								<button class="dropdown-item" on:click={() => toggleKeyStatus(key)}>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
										<circle cx="12" cy="16" r="1"></circle>
										<path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
									</svg>
									{key.status === 'active' ? 'Deactivate' : 'Activate'}
								</button>
								<button class="dropdown-item" on:click={() => regenerateApiKey(key)}>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<polyline points="23,4 23,10 17,10"></polyline>
										<polyline points="1,20 1,14 7,14"></polyline>
										<path d="M20.49 9A9 9 0 0 0 5.64 5.64l1.27 1.27m4.18 4.18l1.27 1.27A9 9 0 0 0 18.36 18.36"></path>
									</svg>
									Regenerate
								</button>
								<div class="dropdown-divider"></div>
								<button class="dropdown-item danger" on:click={() => deleteApiKey(key)}>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<polyline points="3,6 5,6 21,6"></polyline>
										<path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
									</svg>
									Delete
								</button>
							</div>
						</div>
					</div>
				</div>
			{/each}

			{#if apiKeys.length === 0}
				<div class="empty-state">
					<div class="empty-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"></path>
						</svg>
					</div>
					<h3>No API keys found</h3>
					<p>No API keys match your current filters.</p>
					<button class="btn btn-primary" on:click={openCreateModal}>
						Create Your First API Key
					</button>
				</div>
			{/if}
		</div>
	{/if}
</div>

<!-- Create/Edit Modal -->
{#if showCreateModal || showEditModal}
	<div class="modal-overlay" 
		on:click={closeModals}
		on:keydown={(e) => e.key === 'Escape' && closeModals()}
		role="button"
		aria-modal="true"
		tabindex="-1">
		<div class="modal-content" on:click|stopPropagation>
			<div class="modal-header">
				<h2>{showCreateModal ? 'Create New API Key' : 'Edit API Key'}</h2>
				<button class="modal-close" on:click={closeModals}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"></line>
						<line x1="6" y1="6" x2="18" y2="18"></line>
					</svg>
				</button>
			</div>

			<div class="modal-body">
				<div class="form-group">
					<label for="keyName">Name *</label>
					<input
						id="keyName"
						type="text"
						placeholder="Enter API key name"
						bind:value={formData.name}
						class="form-input"
						required
					/>
				</div>

				<div class="form-group">
					<label for="keyDescription">Description</label>
					<textarea
						id="keyDescription"
						placeholder="Enter API key description"
						bind:value={formData.description}
						class="form-textarea"
						rows="3"
					></textarea>
				</div>

				<div class="form-group">
					<label>Permissions *</label>
					<div class="permissions-grid">
						{#each permissionOptions as permission}
							<label class="permission-checkbox">
								<input
									type="checkbox"
									checked={formData.permissions.includes(permission.value)}
									on:change={(e) => {
										const target = e.target as HTMLInputElement;
										if (target) handlePermissionChange(permission.value, target.checked);
									}}
								/>
								<span class="checkbox-label">{permission.label}</span>
							</label>
						{/each}
					</div>
				</div>

				<div class="form-row">
					<div class="form-group">
						<label for="rateLimit">Rate Limit (requests/hour)</label>
						<input
							id="rateLimit"
							type="number"
							min="1"
							max="10000"
							bind:value={formData.rate_limit}
							class="form-input"
						/>
					</div>

					<div class="form-group">
						<label for="expiresAt">Expires At (optional)</label>
						<input
							id="expiresAt"
							type="date"
							bind:value={formData.expires_at}
							class="form-input"
						/>
					</div>
				</div>
			</div>

			<div class="modal-footer">
				<button class="btn btn-outline" on:click={closeModals}>
					Cancel
				</button>
				<button class="btn btn-primary" on:click={showCreateModal ? createApiKey : updateApiKey}>
					{showCreateModal ? 'Create API Key' : 'Update API Key'}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.api-keys-page {
		padding: var(--space-xl);
		background: var(--bg-secondary);
		min-height: 100vh;
	}

	.page-header {
		margin-bottom: var(--space-2xl);
	}

	.header-content {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: var(--space-xl);
		flex-wrap: wrap;
	}

	.header-text h1 {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: var(--space-sm);
	}

	.header-text p {
		color: var(--text-secondary);
		font-size: var(--text-lg);
		margin: 0;
	}

	.header-actions {
		display: flex;
		gap: var(--space-sm);
	}

	.filters-section {
		padding: var(--space-lg);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
		margin-bottom: var(--space-xl);
	}

	.filters-grid {
		display: grid;
		grid-template-columns: 2fr 1fr;
		gap: var(--space-lg);
		align-items: end;
	}

	.filter-group {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}

	.filter-group label {
		font-weight: 600;
		color: var(--text-primary);
		font-size: var(--text-sm);
	}

	.filter-input,
	.filter-select {
		padding: var(--space-sm) var(--space-md);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		background: var(--bg-glass);
		color: var(--text-primary);
		font-size: var(--text-sm);
	}

	.keys-table {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
	}

	.table-header {
		display: grid;
		grid-template-columns: 1.5fr 1.5fr 1fr 1fr 0.8fr 1fr 0.8fr;
		gap: var(--space-lg);
		padding: var(--space-md) 0;
		border-bottom: 1px solid var(--border-color);
		margin-bottom: var(--space-lg);
	}

	.header-cell {
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.table-row {
		display: grid;
		grid-template-columns: 1.5fr 1.5fr 1fr 1fr 0.8fr 1fr 0.8fr;
		gap: var(--space-lg);
		padding: var(--space-md) 0;
		border-bottom: 1px solid var(--border-color);
		align-items: center;
	}

	.table-row:last-child {
		border-bottom: none;
	}

	.key-info {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.key-name {
		font-weight: 600;
		color: var(--text-primary);
		font-size: var(--text-sm);
	}

	.key-description {
		font-size: var(--text-xs);
		color: var(--text-secondary);
	}

	.key-creator {
		font-size: var(--text-xs);
		color: var(--text-tertiary);
	}

	.api-key-display {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
	}

	.api-key {
		font-family: monospace;
		font-size: var(--text-sm);
		color: var(--text-primary);
		background: var(--bg-glass);
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-sm);
		border: 1px solid var(--border-color);
	}

	.copy-btn {
		padding: var(--space-xs);
		background: none;
		border: none;
		color: var(--text-secondary);
		cursor: pointer;
		transition: color var(--transition-fast);
	}

	.copy-btn:hover {
		color: var(--primary);
	}

	.copy-btn svg {
		width: 16px;
		height: 16px;
	}

	.permissions-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.permission-badge {
		display: inline-block;
		padding: var(--space-xs) var(--space-sm);
		background: var(--primary-bg);
		color: var(--primary-text);
		border-radius: var(--radius-sm);
		font-size: var(--text-xs);
		font-weight: 500;
	}

	.permission-more {
		font-size: var(--text-xs);
		color: var(--text-secondary);
		font-style: italic;
	}

	.usage-info {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.usage-count {
		font-weight: 600;
		color: var(--text-primary);
		font-size: var(--text-sm);
	}

	.rate-limit {
		font-size: var(--text-xs);
		color: var(--text-secondary);
	}

	.status-badge {
		display: inline-block;
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-md);
		font-size: var(--text-xs);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.status-active {
		background: var(--success-bg);
		color: var(--success-text);
	}

	.status-inactive {
		background: var(--warning-bg);
		color: var(--warning-text);
	}

	.status-expired {
		background: var(--error-bg);
		color: var(--error-text);
	}

	.last-used {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.actions-dropdown {
		position: relative;
	}

	.actions-dropdown:hover .dropdown-menu {
		opacity: 1;
		visibility: visible;
		transform: translateY(0);
	}

	.dropdown-menu {
		position: absolute;
		top: 100%;
		right: 0;
		margin-top: var(--space-sm);
		min-width: 150px;
		padding: var(--space-sm);
		background: var(--bg-glass);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-lg);
		opacity: 0;
		visibility: hidden;
		transform: translateY(-10px);
		transition: all var(--transition-normal);
		z-index: 10;
	}

	.dropdown-item {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		padding: var(--space-sm);
		color: var(--text-primary);
		background: none;
		border: none;
		width: 100%;
		text-align: left;
		cursor: pointer;
		font-size: var(--text-sm);
		border-radius: var(--radius-md);
		transition: background var(--transition-fast);
	}

	.dropdown-item:hover {
		background: var(--bg-hover);
	}

	.dropdown-item.danger {
		color: var(--error);
	}

	.dropdown-item.danger:hover {
		background: var(--error-bg);
	}

	.dropdown-item svg {
		width: 16px;
		height: 16px;
	}

	.dropdown-divider {
		height: 1px;
		background: var(--border-color);
		margin: var(--space-sm) 0;
	}

	.btn-small {
		padding: var(--space-xs) var(--space-sm);
		font-size: var(--text-sm);
	}

	.empty-state {
		text-align: center;
		padding: var(--space-3xl) 0;
	}

	.empty-icon {
		width: 80px;
		height: 80px;
		margin: 0 auto var(--space-lg);
		color: var(--text-secondary);
	}

	.empty-icon svg {
		width: 100%;
		height: 100%;
	}

	.empty-state h3 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-sm);
	}

	.empty-state p {
		color: var(--text-secondary);
		margin-bottom: var(--space-lg);
	}

	.loading-container,
	.error-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 400px;
		gap: var(--space-lg);
	}

	.error-message {
		color: var(--error);
		font-size: var(--text-lg);
	}

	/* Modal Styles */
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
		padding: var(--space-lg);
	}

	.modal-content {
		background: var(--bg-primary);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
		box-shadow: var(--shadow-2xl);
		max-width: 600px;
		width: 100%;
		max-height: 90vh;
		overflow-y: auto;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-xl);
		border-bottom: 1px solid var(--border-color);
	}

	.modal-header h2 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.modal-close {
		background: none;
		border: none;
		color: var(--text-secondary);
		cursor: pointer;
		padding: var(--space-sm);
		border-radius: var(--radius-md);
		transition: all var(--transition-fast);
	}

	.modal-close:hover {
		background: var(--bg-hover);
		color: var(--text-primary);
	}

	.modal-close svg {
		width: 20px;
		height: 20px;
	}

	.modal-body {
		padding: var(--space-xl);
	}

	.form-group {
		margin-bottom: var(--space-lg);
	}

	.form-group label {
		display: block;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-sm);
		font-size: var(--text-sm);
	}

	.form-input,
	.form-textarea {
		width: 100%;
		padding: var(--space-sm) var(--space-md);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		background: var(--bg-glass);
		color: var(--text-primary);
		font-size: var(--text-sm);
		transition: border-color var(--transition-fast);
	}

	.form-input:focus,
	.form-textarea:focus {
		outline: none;
		border-color: var(--primary);
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-lg);
	}

	.permissions-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: var(--space-md);
	}

	.permission-checkbox {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		cursor: pointer;
	}

	.permission-checkbox input[type="checkbox"] {
		margin: 0;
	}

	.checkbox-label {
		font-size: var(--text-sm);
		color: var(--text-primary);
	}

	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-sm);
		padding: var(--space-xl);
		border-top: 1px solid var(--border-color);
	}

	@media (max-width: 768px) {
		.api-keys-page {
			padding: var(--space-lg);
		}

		.header-content {
			flex-direction: column;
			align-items: stretch;
		}

		.filters-grid {
			grid-template-columns: 1fr;
		}

		.table-header,
		.table-row {
			grid-template-columns: 1fr;
			gap: var(--space-sm);
		}

		.header-cell {
			display: none;
		}

		.table-cell {
			display: flex;
			justify-content: space-between;
			align-items: center;
			padding: var(--space-sm) 0;
		}

		.table-cell::before {
			content: attr(data-label);
			font-weight: 600;
			color: var(--text-secondary);
		}

		.form-row {
			grid-template-columns: 1fr;
		}

		.permissions-grid {
			grid-template-columns: 1fr;
		}
	}
</style> 