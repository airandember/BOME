<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	interface User {
		id: number;
		email: string;
		firstName: string;
		lastName: string;
		role: string;
		emailVerified: boolean;
		createdAt: string;
		lastLogin: string | null;
		status: string;
		subscriptionStatus: string;
	}

	interface Pagination {
		page: number;
		limit: number;
		total: number;
		totalPages: number;
	}

	let users: User[] = [];
	let pagination: Pagination = { page: 1, limit: 10, total: 0, totalPages: 0 };
	let loading = true;
	let searchTerm = '';
	let roleFilter = '';
	let statusFilter = '';
	let selectedUsers: number[] = [];

	onMount(() => {
		loadUsers();
	});

	async function loadUsers() {
		try {
			loading = true;
			const params = new URLSearchParams({
				page: pagination.page.toString(),
				limit: pagination.limit.toString(),
				...(searchTerm && { search: searchTerm }),
				...(roleFilter && { role: roleFilter }),
				...(statusFilter && { status: statusFilter })
			});

			const response = await api.get(`/api/v1/admin/users?${params}`);
			users = response.users || [];
			pagination = response.pagination || pagination;
		} catch (error) {
			showToast('Failed to load users', 'error');
			console.error('Error loading users:', error);
		} finally {
			loading = false;
		}
	}

	function handleSearch() {
		pagination.page = 1;
		loadUsers();
	}

	function handleFilterChange() {
		pagination.page = 1;
		loadUsers();
	}

	function handlePageChange(newPage: number) {
		pagination.page = newPage;
		loadUsers();
	}

	function formatDate(dateString: string | null) {
		if (!dateString) return 'Never';
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function getStatusBadge(status: string) {
		switch (status) {
			case 'active':
				return 'badge-success';
			case 'pending':
				return 'badge-warning';
			case 'suspended':
				return 'badge-error';
			default:
				return 'badge-neutral';
		}
	}

	function getRoleBadge(role: string) {
		switch (role) {
			case 'admin':
				return 'badge-primary';
			case 'moderator':
				return 'badge-info';
			default:
				return 'badge-neutral';
		}
	}

	function toggleUserSelection(userId: number) {
		if (selectedUsers.includes(userId)) {
			selectedUsers = selectedUsers.filter(id => id !== userId);
		} else {
			selectedUsers = [...selectedUsers, userId];
		}
	}

	function selectAllUsers() {
		if (selectedUsers.length === users.length) {
			selectedUsers = [];
		} else {
			selectedUsers = users.map(user => user.id);
		}
	}

	async function updateUserRole(userId: number, newRole: string) {
		try {
			await api.put(`/api/v1/admin/users/${userId}`, { role: newRole });
			showToast('User role updated successfully', 'success');
			loadUsers();
		} catch (error) {
			showToast('Failed to update user role', 'error');
		}
	}

	async function deleteUser(userId: number) {
		if (!confirm('Are you sure you want to delete this user? This action cannot be undone.')) {
			return;
		}

		try {
			await api.delete(`/api/v1/admin/users/${userId}`);
			showToast('User deleted successfully', 'success');
			loadUsers();
		} catch (error) {
			showToast('Failed to delete user', 'error');
		}
	}
</script>

<svelte:head>
	<title>User Management - Admin Dashboard</title>
</svelte:head>

<div class="users-page">
	<div class="page-header">
		<h1>User Management</h1>
		<p>Manage users, roles, and permissions</p>
	</div>

	<!-- Filters and Search -->
	<div class="filters-section glass">
		<div class="search-box">
			<input
				type="text"
				placeholder="Search users by name or email..."
				bind:value={searchTerm}
				on:keydown={(e) => e.key === 'Enter' && handleSearch()}
				class="search-input"
			/>
			<button class="btn btn-primary" on:click={handleSearch}>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="11" cy="11" r="8"></circle>
					<path d="m21 21-4.35-4.35"></path>
				</svg>
				Search
			</button>
		</div>

		<div class="filter-controls">
			<select bind:value={roleFilter} on:change={handleFilterChange} class="filter-select">
				<option value="">All Roles</option>
				<option value="admin">Admin</option>
				<option value="moderator">Moderator</option>
				<option value="user">User</option>
			</select>

			<select bind:value={statusFilter} on:change={handleFilterChange} class="filter-select">
				<option value="">All Status</option>
				<option value="active">Active</option>
				<option value="pending">Pending</option>
				<option value="suspended">Suspended</option>
			</select>
		</div>
	</div>

	<!-- Users Table -->
	{#if loading}
		<div class="loading-container">
			<LoadingSpinner size="large" color="primary" />
			<p>Loading users...</p>
		</div>
	{:else if users.length === 0}
		<div class="empty-state glass">
			<div class="empty-icon">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
					<circle cx="9" cy="7" r="4"></circle>
					<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
					<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
				</svg>
			</div>
			<h3>No users found</h3>
			<p>Try adjusting your search or filter criteria</p>
		</div>
	{:else}
		<div class="users-table-container glass">
			<div class="table-header">
				<div class="bulk-actions">
					<label class="checkbox-container">
						<input
							type="checkbox"
							checked={selectedUsers.length === users.length}
							indeterminate={selectedUsers.length > 0 && selectedUsers.length < users.length}
							on:change={selectAllUsers}
						/>
						<span class="checkmark"></span>
					</label>
					{#if selectedUsers.length > 0}
						<span class="selected-count">{selectedUsers.length} selected</span>
						<button class="btn btn-outline btn-small">Bulk Actions</button>
					{/if}
				</div>
			</div>

			<div class="table-wrapper">
				<table class="users-table">
					<thead>
						<tr>
							<th></th>
							<th>User</th>
							<th>Role</th>
							<th>Status</th>
							<th>Subscription</th>
							<th>Created</th>
							<th>Last Login</th>
							<th>Actions</th>
						</tr>
					</thead>
					<tbody>
						{#each users as user}
							<tr class="user-row">
								<td>
									<label class="checkbox-container">
										<input
											type="checkbox"
											checked={selectedUsers.includes(user.id)}
											on:change={() => toggleUserSelection(user.id)}
										/>
										<span class="checkmark"></span>
									</label>
								</td>
								<td>
									<div class="user-info">
										<div class="user-avatar">
											{user.firstName.charAt(0)}{user.lastName.charAt(0)}
										</div>
										<div class="user-details">
											<div class="user-name">{user.firstName} {user.lastName}</div>
											<div class="user-email">{user.email}</div>
											{#if !user.emailVerified}
												<span class="unverified-badge">Unverified</span>
											{/if}
										</div>
									</div>
								</td>
								<td>
									<span class="badge {getRoleBadge(user.role)}">{user.role}</span>
								</td>
								<td>
									<span class="badge {getStatusBadge(user.status)}">{user.status}</span>
								</td>
								<td>
									<span class="subscription-status">{user.subscriptionStatus}</span>
								</td>
								<td>{formatDate(user.createdAt)}</td>
								<td>{formatDate(user.lastLogin)}</td>
								<td>
									<div class="actions">
										<button class="btn btn-ghost btn-small" title="Edit User">
											<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
												<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
												<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
											</svg>
										</button>
										<button 
											class="btn btn-ghost btn-small" 
											title="Delete User"
											on:click={() => deleteUser(user.id)}
										>
											<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
												<polyline points="3,6 5,6 21,6"></polyline>
												<path d="M19,6v14a2,2,0,0,1-2,2H7a2,2,0,0,1-2-2V6m3,0V4a2,2,0,0,1,2-2h4a2,2,0,0,1,2,2V6"></path>
											</svg>
										</button>
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>

			<!-- Pagination -->
			{#if pagination.totalPages > 1}
				<div class="pagination">
					<button 
						class="btn btn-ghost" 
						disabled={pagination.page === 1}
						on:click={() => handlePageChange(pagination.page - 1)}
					>
						Previous
					</button>
					
					<div class="page-numbers">
						{#each Array(pagination.totalPages) as _, i}
							<button 
								class="btn {pagination.page === i + 1 ? 'btn-primary' : 'btn-ghost'}"
								on:click={() => handlePageChange(i + 1)}
							>
								{i + 1}
							</button>
						{/each}
					</div>
					
					<button 
						class="btn btn-ghost" 
						disabled={pagination.page === pagination.totalPages}
						on:click={() => handlePageChange(pagination.page + 1)}
					>
						Next
					</button>
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	.users-page {
		display: flex;
		flex-direction: column;
		gap: var(--space-2xl);
	}

	.page-header h1 {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: var(--space-sm);
	}

	.page-header p {
		color: var(--text-secondary);
		font-size: var(--text-lg);
	}

	.filters-section {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		display: flex;
		gap: var(--space-xl);
		align-items: center;
		flex-wrap: wrap;
	}

	.search-box {
		display: flex;
		gap: var(--space-md);
		flex: 1;
		min-width: 300px;
	}

	.search-input {
		flex: 1;
		padding: var(--space-md);
		border: none;
		border-radius: var(--radius-lg);
		background: var(--bg-glass);
		color: var(--text-primary);
		font-size: var(--text-base);
	}

	.search-input:focus {
		outline: 2px solid var(--primary);
		outline-offset: 2px;
	}

	.filter-controls {
		display: flex;
		gap: var(--space-md);
	}

	.filter-select {
		padding: var(--space-md);
		border: none;
		border-radius: var(--radius-lg);
		background: var(--bg-glass);
		color: var(--text-primary);
		min-width: 120px;
	}

	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-lg);
		padding: var(--space-3xl);
	}

	.empty-state {
		padding: var(--space-3xl);
		text-align: center;
		border-radius: var(--radius-xl);
	}

	.empty-icon {
		width: 64px;
		height: 64px;
		margin: 0 auto var(--space-lg);
		opacity: 0.5;
	}

	.empty-icon svg {
		width: 100%;
		height: 100%;
		color: var(--text-secondary);
	}

	.users-table-container {
		border-radius: var(--radius-xl);
		overflow: hidden;
	}

	.table-header {
		padding: var(--space-lg);
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
	}

	.bulk-actions {
		display: flex;
		align-items: center;
		gap: var(--space-md);
	}

	.selected-count {
		color: var(--text-secondary);
		font-size: var(--text-sm);
	}

	.table-wrapper {
		overflow-x: auto;
	}

	.users-table {
		width: 100%;
		border-collapse: collapse;
	}

	.users-table th,
	.users-table td {
		padding: var(--space-lg);
		text-align: left;
		border-bottom: 1px solid rgba(255, 255, 255, 0.05);
	}

	.users-table th {
		font-weight: 600;
		color: var(--text-secondary);
		font-size: var(--text-sm);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.user-row:hover {
		background: rgba(255, 255, 255, 0.02);
	}

	.user-info {
		display: flex;
		align-items: center;
		gap: var(--space-md);
	}

	.user-avatar {
		width: 40px;
		height: 40px;
		border-radius: 50%;
		background: var(--primary-gradient);
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 600;
		color: var(--white);
		font-size: var(--text-sm);
	}

	.user-name {
		font-weight: 600;
		color: var(--text-primary);
	}

	.user-email {
		color: var(--text-secondary);
		font-size: var(--text-sm);
	}

	.unverified-badge {
		display: inline-block;
		padding: 2px 6px;
		background: var(--warning);
		color: var(--white);
		border-radius: var(--radius-sm);
		font-size: var(--text-xs);
		margin-top: 2px;
	}

	.badge {
		display: inline-block;
		padding: 4px 8px;
		border-radius: var(--radius-md);
		font-size: var(--text-xs);
		font-weight: 600;
		text-transform: uppercase;
	}

	.badge-success {
		background: var(--success);
		color: var(--white);
	}

	.badge-warning {
		background: var(--warning);
		color: var(--white);
	}

	.badge-error {
		background: var(--error);
		color: var(--white);
	}

	.badge-primary {
		background: var(--primary);
		color: var(--white);
	}

	.badge-info {
		background: var(--info);
		color: var(--white);
	}

	.badge-neutral {
		background: var(--bg-glass);
		color: var(--text-secondary);
	}

	.subscription-status {
		text-transform: capitalize;
		color: var(--text-primary);
	}

	.actions {
		display: flex;
		gap: var(--space-sm);
	}

	.checkbox-container {
		position: relative;
		display: inline-block;
		cursor: pointer;
	}

	.checkbox-container input {
		opacity: 0;
		position: absolute;
		cursor: pointer;
	}

	.checkmark {
		position: relative;
		top: 0;
		left: 0;
		height: 20px;
		width: 20px;
		background: var(--bg-glass);
		border-radius: var(--radius-sm);
		border: 2px solid var(--border-color);
		display: inline-block;
		transition: all var(--transition-normal);
	}

	.checkbox-container input:checked ~ .checkmark {
		background: var(--primary);
		border-color: var(--primary);
	}

	.checkmark:after {
		content: "";
		position: absolute;
		display: none;
		left: 6px;
		top: 2px;
		width: 5px;
		height: 10px;
		border: solid var(--white);
		border-width: 0 2px 2px 0;
		transform: rotate(45deg);
	}

	.checkbox-container input:checked ~ .checkmark:after {
		display: block;
	}

	.pagination {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: var(--space-md);
		padding: var(--space-xl);
		border-top: 1px solid rgba(255, 255, 255, 0.1);
	}

	.page-numbers {
		display: flex;
		gap: var(--space-sm);
	}

	.btn {
		padding: var(--space-sm) var(--space-md);
		border: none;
		border-radius: var(--radius-lg);
		font-weight: 500;
		cursor: pointer;
		transition: all var(--transition-normal);
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		text-decoration: none;
		font-size: var(--text-sm);
	}

	.btn-primary {
		background: var(--primary-gradient);
		color: var(--white);
		box-shadow: var(--shadow-md);
	}

	.btn-primary:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}

	.btn-ghost {
		background: var(--bg-glass);
		color: var(--text-primary);
	}

	.btn-ghost:hover {
		background: var(--bg-glass-dark);
	}

	.btn-outline {
		background: transparent;
		color: var(--text-primary);
		border: 1px solid var(--border-color);
	}

	.btn-outline:hover {
		background: var(--bg-glass);
	}

	.btn-small {
		padding: var(--space-xs) var(--space-sm);
		font-size: var(--text-xs);
	}

	.btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
		transform: none;
	}

	.btn svg {
		width: 16px;
		height: 16px;
	}

	@media (max-width: 768px) {
		.filters-section {
			flex-direction: column;
			align-items: stretch;
		}

		.search-box {
			min-width: auto;
		}

		.filter-controls {
			justify-content: space-between;
		}

		.table-wrapper {
			font-size: var(--text-sm);
		}

		.users-table th,
		.users-table td {
			padding: var(--space-md);
		}

		.user-info {
			flex-direction: column;
			align-items: flex-start;
			gap: var(--space-sm);
		}

		.actions {
			flex-direction: column;
		}
	}
</style> 