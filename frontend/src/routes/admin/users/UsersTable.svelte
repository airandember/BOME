<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { StandardizedRole } from '$lib/types/standardized_roles';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	export let users: any[] = [];
	export let roles: StandardizedRole[] = [];
	export let loading = false;
	export let error = '';
	export let currentPage = 1;
	export let totalPages = 0;
	export let searchTerm = '';
	export let roleFilter = '';
	export let statusFilter = '';

	const dispatch = createEventDispatcher<{
		pageChange: { page: number };
		search: { term: string };
		filterChange: void;
		editUser: { user: any };
	}>();

	function handleSearch() {
		dispatch('search', { term: searchTerm });
	}

	function handleFilterChange() {
		dispatch('filterChange');
	}

	function handlePageChange(page: number) {
		dispatch('pageChange', { page });
	}

	function handleEditUser(user: any) {
		dispatch('editUser', { user });
	}

	// Helper functions
	function getRoleColor(roleId: string): string {
		const role = roles.find(r => r.id === roleId);
		return role?.color || '#6b7280';
	}

	function getRoleIcon(roleId: string): string {
		const role = roles.find(r => r.id === roleId);
		return role?.icon || 'user';
	}

	function formatDate(dateString: string | null): string {
		if (!dateString) return 'Never';
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function getStatusBadgeClass(status: string): string {
		const statusClasses: { [key: string]: string } = {
			'active': 'bg-green-100 text-green-800',
			'inactive': 'bg-gray-100 text-gray-800',
			'pending': 'bg-yellow-100 text-yellow-800',
			'suspended': 'bg-red-100 text-red-800'
		};
		return statusClasses[status] || 'bg-gray-100 text-gray-800';
	}

	function getSubscriptionBadgeClass(subscription: string): string {
		const subscriptionClasses: { [key: string]: string } = {
			'premium': 'bg-green-600 text-white',
			'standard': 'bg-blue-100 text-blue-800',
			'free': 'warning text-gray-800'
		};
		return subscriptionClasses[subscription] || 'bg-gray-100 text-gray-800';
	}
</script>

<div class="users-section">
	<!-- Search and Filter Bar -->
	<div class="filter-bar">
		<div class="search-container">
			<input
				type="text"
				placeholder="Search users by name or email..."
				bind:value={searchTerm}
				on:input={handleSearch}
				class="search-input"
			/>
			<span class="search-icon">🔍</span>
		</div>

		<div class="filter-container">
			<select bind:value={roleFilter} on:change={handleFilterChange} class="filter-select">
				<option value="">All Roles</option>
				{#each roles as role}
					<option value={role.id}>{role.name}</option>
				{/each}
			</select>

			<select bind:value={statusFilter} on:change={handleFilterChange} class="filter-select">
				<option value="">All Status</option>
				<option value="active">Active</option>
				<option value="pending">Pending</option>
				<option value="inactive">Inactive</option>
				<option value="suspended">Suspended</option>
			</select>
		</div>
	</div>

	<!-- Users Table -->
	{#if loading}
		<div class="loading-container">
			<LoadingSpinner />
			<p>Loading users...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<p class="error-message">{error}</p>
			<button class="btn btn-secondary" on:click={() => dispatch('filterChange')}>Retry</button>
		</div>
	{:else}
		<div class="table-container">
			<table class="users-table">
				<thead>
					<tr>
						<th>User</th>
						<th>LID</th>
						<th>Email</th>
						<th>Verified</th>
						<th>Role</th>
						<th>Status</th>
						<th>Sub</th>
						<th>Stripe Customer ID</th>
						<th>Created</th>
						<th>Last Login</th>
						<th>Actions</th>
					</tr>
				</thead>
				<tbody>
					{#each users as user, index (user.id || `user-${index}`)}
						<tr>
							<td>
								<div class="user-info">
									<div class="user-avatar">
										{(user.FirstName || '').charAt(0)}{(user.LastName || '').charAt(0)}
									</div>
									<div class="user-details">
										<div class="user-name">{user.FirstName || 'Unknown'} {user.LastName || 'User'}</div>
										<div class="user-id">ID: {user.ID}</div>
									</div>
								</div>
							</td>
							<td>
								<span>{user.ID}</span>
							</td>
							<td>
								<div class="email-cell">
									<span class="email">{user.Email}</span>
								</div>
							</td>
							<td>
								<div class="status-cell">
									{#if user.EmailVerified}
										<span class="verified-badge">✓</span>
									{:else} 
										<span class="unverified-badge">🚫</span>
									{/if}
								</div>
							</td>
							<td>
								<div class="role-badge" style="background: {getRoleColor(user.Role)}20; color: {getRoleColor(user.Role)}">
									{getRoleIcon(user.Role)} {roles.find(r => r.id === user.Role)?.name || user.Role}
								</div>
							</td>
							<td>
								<div class="status-cell">
									<span class="status-badge {getStatusBadgeClass(user.IsActive.Bool ? 'active' : 'inactive')}">
										{user.IsActive.Bool ? 'Active' : 'Inactive'}
									</span>
								</div>
							</td>

							<td>
								<span class="subscription-badge {getSubscriptionBadgeClass(user.HasSubbed.Bool ? 'premium' : 'free')}">
									{user.HasSubbed.Bool ? 'Subscribed' : 'Free'}
								</span>
							</td>
							<td>
								<span>{user.StripeCustomerID?.String || ''}</span>
							</td>
							<td>
								<span>{formatDate(user.CreatedAt)}</span>
							</td>
							<td>
								<span class="last-login">{formatDate(user.LastLogin?.Time)}</span>
							</td>
							<td>
								<div class="action-buttons">
									<button class="btn btn-sm btn-secondary" on:click={() => handleEditUser(user)}>
										Edit
									</button>
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>

		<!-- Pagination -->
		{#if totalPages > 1}
			<div class="pagination">
				<button 
					class="btn btn-secondary" 
					disabled={currentPage === 1}
					on:click={() => handlePageChange(currentPage - 1)}
				>
					Previous
				</button>
				
				<div class="page-numbers">
					{#each Array.from({length: totalPages}, (_, i) => i + 1) as pageNum}
						<button 
							class="page-number" 
							class:active={pageNum === currentPage}
							on:click={() => handlePageChange(pageNum)}
						>
							{pageNum}
						</button>
					{/each}
				</div>
				
				<button 
					class="btn btn-secondary" 
					disabled={currentPage === totalPages}
					on:click={() => handlePageChange(currentPage + 1)}
				>
					Next
				</button>
			</div>
		{/if}
	{/if}
</div>

<style>

	.bg-green-100 {
		background: var(--bg-green-100);
		color: var(--bg-green-600);
	}

	.bg-green-600 {
		background: var(--bg-green-600);
		color: var(--text-inverse);
	}

	.bg-gray-100 {
		background: var(--gray-800);
		color: var(--text-inverse);
	}

	.warning {
		background: var(--warning);
	}

	.users-section {
		background: white;
		border-radius: 0.5rem;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
		overflow: hidden;
	}

	.filter-bar {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem 1.5rem;
		border-bottom: 1px solid #e5e7eb;
		background: #f9fafb;
	}

	.search-container {
		position: relative;
		flex: 1;
		max-width: 400px;
	}

	.search-input {
		width: 100%;
		padding: 0.5rem 1rem 0.5rem 2.5rem;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		font-size: 0.875rem;
	}

	.search-icon {
		position: absolute;
		left: 0.75rem;
		top: 50%;
		transform: translateY(-50%);
		color: #6b7280;
	}

	.filter-container {
		display: flex;
		gap: 1rem;
	}

	.filter-select {
		padding: 0.5rem;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		background: white;
	}

	.table-container {
		overflow-x: auto;
	}

	.users-table {
		width: 100%;
		border-collapse: collapse;
	}

	.users-table th,
	.users-table td {
		padding: 1rem;
		text-align: left;
		border-bottom: 1px solid #e5e7eb;
	}

	.users-table th {
		background: #f9fafb;
		font-weight: 600;
		color: #374151;
		font-size: 0.875rem;
	}

	.user-info {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.user-avatar {
		width: 2.5rem;
		height: 2.5rem;
		border-radius: 50%;
		background: #3b82f6;
		color: white;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 600;
		font-size: 0.875rem;
	}

	.user-name {
		font-weight: 500;
		color: #111827;
	}

	.user-id {
		font-size: 0.75rem;
		color: #6b7280;
	}

	.email-cell {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.status-cell {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		text-align: center;
		justify-content: center;
	}

	.verified-badge {
		color: #059669;
		font-weight: bold;
	}

	.role-badge, .department-badge, .status-badge, .subscription-badge {
		padding: 0.25rem 0.5rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		font-weight: 500;
	}

	.last-login {
		font-size: 0.875rem;
		color: #6b7280;
	}

	.action-buttons {
		display: flex;
		gap: 0.5rem;
	}

	/* Pagination */
	.pagination {
		display: flex;
		justify-content: center;
		align-items: center;
		gap: 1rem;
		padding: 1rem;
		border-top: 1px solid #e5e7eb;
	}

	.page-numbers {
		display: flex;
		gap: 0.25rem;
	}

	.page-number {
		padding: 0.5rem 0.75rem;
		border: 1px solid #d1d5db;
		background: white;
		color: #374151;
		border-radius: 0.25rem;
		cursor: pointer;
		transition: all 0.2s;
	}

	.page-number:hover {
		background: #f3f4f6;
	}

	.page-number.active {
		background: #2563eb;
		color: white;
		border-color: #2563eb;
	}

	/* Loading and Error States */
	.loading-container, .error-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 3rem;
		text-align: center;
	}

	.error-message {
		color: #dc2626;
		margin-bottom: 1rem;
	}

	/* Buttons */
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

	.btn-sm {
		padding: 0.25rem 0.5rem;
		font-size: 0.75rem;
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.filter-bar {
			flex-direction: column;
			gap: 1rem;
		}

		.search-container {
			max-width: none;
		}

		.filter-container {
			width: 100%;
		}

		.filter-select {
			flex: 1;
		}
	}
</style> 
