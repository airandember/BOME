<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { auth } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import { 
		type StandardizedRole,
		type StandardizedPermission,
		type Department
	} from '$lib/types/standardized_roles';
	import { slide } from 'svelte/transition';
	import { 
		FALLBACK_ROLES, 
		FALLBACK_DEPARTMENTS, 
		FALLBACK_USERS,
		getDepartmentsWithRoles,
		getDepartmentUserCount
	} from '$lib/mockData/adminUsers';

	// Tab management
	let activeTab = 'overview';
	
	// Data state
	let users: any[] = [];
	let totalUsers = 0;
	let userStats = {
		total: 0,
		admins: 0,
		verified: 0,
		pending: 0,
		active: 0
	};
	let loading = false;
	let error = '';
	
	// Pagination state
	let currentPage = 1;
	let itemsPerPage = 50;
	let totalPages = 0;
	
	// Filter state
	let searchTerm = '';
	let roleFilter = '';
	let statusFilter = '';
	
	// User management state
	let selectedUsers = new Set<number>();
	let showUserModal = false;
	let editingUser: any = null;
	let userForm = {
		firstName: '',
		lastName: '',
		email: '',
		role: '',
		emailVerified: false
	};

	// State for database-driven roles and departments
	let roles: StandardizedRole[] = [];
	let departments: Department[] = [];
	let rolesLoading = true;
	let rolesError: string | null = null;
	let expandedDepartments = new Set<number>();

	// Fetch roles and departments from backend
	async function fetchRolesAndDepartments() {
		try {
			rolesLoading = true;
			rolesError = null;
			
			const token = $auth.token;
			if (!token) {
				throw new Error('No authentication token available');
			}

			const response = await fetch('/api/v1/admin/rolesAndDepartments', {
				headers: {
					'Authorization': `Bearer ${token}`,
					'Content-Type': 'application/json'
				}
			});

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			const data = await response.json();
			
			if (data.success && data.data) {
				roles = data.data.roles || [];
				departments = data.data.departments || [];
			} else {
				throw new Error('Invalid response format');
			}
			
		} catch (err) {
			console.error('❌ Admin Users: Error fetching roles and departments:', err);
			rolesError = err instanceof Error ? err.message : 'Unknown error occurred';
			// Fallback to mock data if API fails
			//roles = FALLBACK_ROLES;
			//departments = FALLBACK_DEPARTMENTS;
		} finally {
			rolesLoading = false;
		}
	}

	// Load user statistics from backend
	async function loadUserStats() {
		const token = $auth.token;
		if (!token) {
			console.error('❌ Admin Users: No authentication token available');
			return;
		}

		try {
			const response = await fetch('/api/v1/admin/users/stats', {
				headers: {
					'Authorization': `Bearer ${token}`
				}
			});

			if (response.ok) {
				const data = await response.json();
				userStats = data.stats || {
					total: 225,
					admins: 0,
					verified: 0,
					pending: 0,
					active: 0
				};
				console.log(userStats);
			} else {
				console.error('❌ Admin Users: Failed to load user statistics:', response.status);
			}
		} catch (err) {
			console.error('💥 Admin Users: Error loading user statistics:', err);
		}
	}

	// Load users from backend
	async function loadUsers() {
		const token = $auth.token;
		if (!token) {
			console.error('❌ Admin Users: No authentication token available');
			return;
		}

		try {
			loading = true;
			error = '';

			const params = new URLSearchParams({
				page: currentPage.toString(),
				limit: itemsPerPage.toString()
			});

			if (searchTerm) params.append('search', searchTerm);
			if (roleFilter) params.append('role', roleFilter);
			if (statusFilter) params.append('status', statusFilter);

			const response = await fetch(`/api/v1/admin/users?${params}`, {
				headers: {
					'Authorization': `Bearer ${token}`
				}
			});

			if (response.ok) {
				const data = await response.json();
				console.log(data);
				users = data.users || [];
				totalUsers = data.total || 0;
				totalPages = Math.ceil(totalUsers / itemsPerPage);
			} else {
				throw new Error(`HTTP error! status: ${response.status}`);
			}
		} catch (err) {
			console.error('❌ Admin Users: Error loading users:', err);
			error = err instanceof Error ? err.message : 'Failed to load users';
			// Fallback to mock data
			users = FALLBACK_USERS;
			totalUsers = FALLBACK_USERS.length;
			totalPages = 1;
		} finally {
			loading = false;
		}
	}

	// Load data on mount
	onMount(async () => {
		console.log('🚀 Admin Users: Component mounted');
		
		// Check authentication
		if (!$auth.isAuthenticated) {
			console.log('❌ Admin Users: User not authenticated, redirecting to login');
			window.location.href = '/login';
			return;
		}

		// Load all data
		await Promise.all([
			fetchRolesAndDepartments(),
			loadUserStats(),
			loadUsers()
		]);
	});

	// Function to handle page changes
	async function changePage(newPage: number) {
		currentPage = newPage;
		await loadUsers();
	}

	// Function to handle page size changes
	async function changePageSize(newSize: number) {
		itemsPerPage = newSize;
		currentPage = 1; // Reset to first page
		await loadUsers();
	}

	// Function to handle search changes with debouncing
	let searchTimeout: ReturnType<typeof setTimeout>;
	async function handleSearch() {
		clearTimeout(searchTimeout);
		searchTimeout = setTimeout(async () => {
			try {
				if (searchTerm === '' || searchTerm.length >= 3) {
					currentPage = 1; // Reset to first page when searching
					await loadUsers();
				}
			} catch (error) {
				console.error('Search error:', error);
				showToast('Search failed. Please try again.', 'error');
			}
		}, 500);
	}

	// Function to handle filter changes
	async function handleFilterChange() {
		currentPage = 1; // Reset to first page when filtering
		await loadUsers();
	}

	// Accordion functions for roles tab
	function toggleDepartment(deptId: number) {
		if (expandedDepartments.has(deptId)) {
			expandedDepartments.delete(deptId);
		} else {
			expandedDepartments.add(deptId);
		}
		expandedDepartments = expandedDepartments; // Trigger reactivity
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
			'premium': 'bg-purple-100 text-purple-800',
			'standard': 'bg-blue-100 text-blue-800',
			'free': 'bg-gray-100 text-gray-800'
		};
		return subscriptionClasses[subscription] || 'bg-gray-100 text-gray-800';
	}

	// Reactive statements
	$: departmentsWithRoles = getDepartmentsWithRoles(roles, departments);
	$: roleStats = {
		total: roles.length,
		systemRoles: roles.filter(r => r.isSystemRole).length,
		customRoles: roles.filter(r => !r.isSystemRole).length,
		activeRoles: roles.filter(r => users.some(u => u.role === r.id)).length
	};
</script>

<svelte:head>
	<title>User Management - BOME Admin</title>
</svelte:head>

<div class="admin-page">
	<div class="page-header">
		<div class="header-content">
			<h1 class="page-title">User Management</h1>
			<p class="page-description">Manage users, roles, and access control across the platform</p>
		</div>
		<div class="header-actions">
			<button class="btn btn-primary" on:click={() => showUserModal = true}>
				<span class="btn-icon">➕</span>
				Add User
			</button>
		</div>
	</div>

	<!-- Tab Navigation -->
	<div class="tab-navigation">
		<button 
			class="tab-button" 
			class:active={activeTab === 'overview'}
			on:click={() => activeTab = 'overview'}
		>
			<span class="tab-icon">📊</span>
			Overview
		</button>
		<button 
			class="tab-button" 
			class:active={activeTab === 'users'}
			on:click={() => activeTab = 'users'}
		>
			<span class="tab-icon">👥</span>
			Users ({userStats.total})
		</button>
		<button 
			class="tab-button" 
			class:active={activeTab === 'roles'}
			on:click={() => activeTab = 'roles'}
		>
			<span class="tab-icon">🎭</span>
			Roles ({roles.length})
		</button>
	</div>

	<!-- Tab Content -->
	<div class="tab-content">
		{#if activeTab === 'overview'}
			<!-- Overview Tab -->
			<div class="overview-grid">
				<!-- User Statistics Cards -->
				<div class="stats-section">
					<h2 class="section-title">User Statistics</h2>
					<div class="stats-grid">
						<div class="stat-card">
							<div class="stat-icon">👥</div>
							<div class="stat-content">
								<div class="stat-value">{userStats.total}</div>
								<div class="stat-label">Total Users</div>
							</div>
						</div>
						<div class="stat-card">
							<div class="stat-icon">✅</div>
							<div class="stat-content">
								<div class="stat-value">{userStats.verified}</div>
								<div class="stat-label">Verified</div>
							</div>
						</div>
						<div class="stat-card">
							<div class="stat-icon">⏳</div>
							<div class="stat-content">
								<div class="stat-value">{userStats.pending}</div>
								<div class="stat-label">Pending</div>
							</div>
						</div>
						<div class="stat-card">
							<div class="stat-icon">🟢</div>
							<div class="stat-content">
								<div class="stat-value">{userStats.active}</div>
								<div class="stat-label">Active</div>
							</div>
						</div>
					</div>
				</div>

				<!-- Role Statistics -->
				<div class="stats-section">
					<h2 class="section-title">Role Statistics</h2>
					<div class="stats-grid">
						<div class="stat-card">
							<div class="stat-icon">🎭</div>
							<div class="stat-content">
								<div class="stat-value">{roleStats.total}</div>
								<div class="stat-label">Total Roles</div>
							</div>
						</div>
						<div class="stat-card">
							<div class="stat-icon">🔧</div>
							<div class="stat-content">
								<div class="stat-value">{roleStats.systemRoles}</div>
								<div class="stat-label">System Roles</div>
							</div>
						</div>
						<div class="stat-card">
							<div class="stat-icon">⚙️</div>
							<div class="stat-content">
								<div class="stat-value">{roleStats.customRoles}</div>
								<div class="stat-label">Custom Roles</div>
							</div>
						</div>
						<div class="stat-card">
							<div class="stat-icon">🎯</div>
							<div class="stat-content">
								<div class="stat-value">{roleStats.activeRoles}</div>
								<div class="stat-label">Active Roles</div>
							</div>
						</div>
					</div>
				</div>

				<!-- Department Distribution -->
				<div class="stats-section">
					<h2 class="section-title">Department Distribution</h2>
					<div class="department-grid">
						{#each departmentsWithRoles as dept, index (dept.id || `overview-dept-${index}`)}
							<div class="department-card">
								<div class="department-header">
                                    <h3 class="department-name">{dept.name}
										<span class="department-icon">{dept.icon}</span>
	                                </h3>
								</div>	
								<div class="department-info">
										
										<p class="department-description">{dept.description}</p>
								</div>	
								<div class="department-stats">
									<span class="role-count">{dept.roles.length} roles</span>
									<span class="user-count">{getDepartmentUserCount(users, dept.name)} users</span>
								</div>
								
							</div>
						{/each}
					</div>
				</div>
			</div>

		{:else if activeTab === 'users'}
			<!-- Users Tab -->
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
						<button class="btn btn-secondary" on:click={loadUsers}>Retry</button>
					</div>
				{:else}
					<div class="table-container">
						<table class="users-table">
							<thead>
								<tr>
									<th>User</th>
									<th>Email</th>
									<th>Verified</th>
									<th>Role</th>
									<!--<th>Department</th>-->
									<th>Status</th>
									<th>Subscription</th>
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
										<!--<td>
											<span class="department-badge">{user.Department}</span>
										</td>-->
										<td>
											<div class="status-cell">
												<span class="status-badge {getStatusBadgeClass(user.Status)}">
													{user.Status}
												</span>
											</div>
										</td>
										<td>
											<span class="subscription-badge {getSubscriptionBadgeClass(user.subscription)}">
												{user.subscription}
											</span>
										</td>
										<td>
											<span class="last-login">{formatDate(user.lastLogin)}</span>
										</td>
										<td>
											<div class="action-buttons">
												<button class="btn btn-sm btn-secondary" on:click={() => { editingUser = user; showUserModal = true; }}>
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
								on:click={() => changePage(currentPage - 1)}
							>
								Previous
							</button>
							
							<div class="page-numbers">
								{#each Array.from({length: totalPages}, (_, i) => i + 1) as pageNum}
									<button 
										class="page-number" 
										class:active={pageNum === currentPage}
										on:click={() => changePage(pageNum)}
									>
										{pageNum}
									</button>
								{/each}
							</div>
							
							<button 
								class="btn btn-secondary" 
								disabled={currentPage === totalPages}
								on:click={() => changePage(currentPage + 1)}
							>
								Next
							</button>
						</div>
					{/if}
				{/if}
			</div>

		{:else if activeTab === 'roles'}
			<!-- Roles Tab -->
			<div class="roles-section">
				{#if rolesLoading}
					<div class="loading-container">
						<LoadingSpinner />
						<p>Loading roles and departments...</p>
					</div>
				{:else if rolesError}
					<div class="error-container">
						<p class="error-message">Error loading roles: {rolesError}</p>
						<button class="btn btn-secondary" on:click={fetchRolesAndDepartments}>Retry</button>
					</div>
				{:else}
					<div class="departments-accordion">
						{#each departmentsWithRoles as dept, index (dept.id || `dept-${index}`)}
							<div class="department-accordion">
								<button 
									class="department-header"
									on:click={() => toggleDepartment(dept.id)}
									aria-expanded={expandedDepartments.has(dept.id)}
									aria-controls="dept-{dept.id}"
								>
									<div class="department-info">
										<span class="department-icon">{dept.icon}</span>
										<div class="department-details">
											<h3 class="department-name">{dept.name}</h3>
											<p class="department-description">{dept.description}</p>
										</div>
									</div>
									<div class="department-stats">
										<span class="role-count">{dept.roles.length} roles</span>
										<span class="user-count">{getDepartmentUserCount(users, dept.name)} users</span>
									</div>
									<span class="accordion-icon">
										{expandedDepartments.has(dept.id) ? '▼' : '▶'}
									</span>
								</button>
								
								{#if expandedDepartments.has(dept.id)}
									<div 
										class="department-content"
										id="dept-{dept.id}"
										transition:slide={{ duration: 300 }}
									>
										<div class="roles-grid">
											{#each dept.roles as role, roleIndex (role.id || `role-${dept.id}-${roleIndex}`)}
												<div class="role-card">
													<div class="role-header">
														<div class="role-icon" style="color: {role.color}">
															{role.icon}
														</div>
														<div class="role-info">
															<h4 class="role-name">{role.name}</h4>
															<p class="role-description">{role.description}</p>
														</div>
														{#if role.isSystemRole}
															<div class="system-badge">System</div>
														{/if}
													</div>
													
													<div class="role-stats">
														<div class="stat">
															<span class="stat-label">Level</span>
															<span class="stat-value">{role.level}/10</span>
														</div>
														<div class="stat">
															<span class="stat-label">Permissions</span>
															<span class="stat-value">{role.permissions.length}</span>
														</div>
														<div class="stat">
															<span class="stat-label">Category</span>
															<span class="stat-value category-badge" style="background: {role.color}20">
																{role.category}
															</span>
														</div>
													</div>

													<div class="role-footer">
														<span class="role-updated">Updated {formatDate(role.updatedAt)}</span>
														<div class="role-actions">
															<button class="btn btn-sm btn-secondary">View Details</button>
														</div>
													</div>
												</div>
											{/each}
										</div>
									</div>
								{/if}
							</div>
						{/each}
					</div>
				{/if}
			</div>
		{/if}
	</div>
</div>

<style>
	.admin-page {
		max-width: 100%;
		margin: 0 auto;
		padding: 2rem;
	}

	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 2rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid #e5e7eb;
	}

	.page-title {
		font-size: 2rem;
		font-weight: 700;
		color: #111827;
		margin: 0;
	}

	.page-description {
		color: #6b7280;
		margin: 0.5rem 0 0 0;
	}

	.header-actions {
		display: flex;
		gap: 1rem;
	}

	/* Tab Navigation */
	.tab-navigation {
		display: flex;
		gap: 0.5rem;
		margin-bottom: 2rem;
		border-bottom: 1px solid #e5e7eb;
	}

	.tab-button {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1.5rem;
		background: none;
		border: none;
		border-bottom: 2px solid transparent;
		color: #6b7280;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
	}

	.tab-button:hover {
		color: #374151;
		background-color: #f9fafb;
	}

	.tab-button.active {
		color: #2563eb;
		border-bottom-color: #2563eb;
	}

	.tab-icon {
		font-size: 1.1rem;
	}

	/* Overview Tab */
	.overview-grid {
		display: grid;
		gap: 2rem;
	}

	.stats-section {
		background: white;
		border-radius: 0.5rem;
		padding: 1.5rem;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
	}

	.section-title {
		font-size: 1.25rem;
		font-weight: 600;
		color: #111827;
		margin: 0 0 1rem 0;
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
	}

	.stat-card {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1rem;
		background: #f9fafb;
		border-radius: 0.5rem;
		border: 1px solid #e5e7eb;
	}

	.stat-icon {
		font-size: 2rem;
	}

	.stat-value {
		font-size: 1.5rem;
		font-weight: 700;
		color: #111827;
	}

	.stat-label {
		color: #6b7280;
		font-size: 0.875rem;
	}

	.department-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: 1rem;
		text-align: center;
		align-items: center;
	}

	.department-card {
		padding: 1rem;
		background: #f9fafb;
		border-radius: 0.5rem;
		border: 1px solid #e5e7eb;
	}

	/*.department-header {
		align-items: center;
		justify-items: center;
		gap: 1rem;
		flex-wrap: nowrap;
		padding: 0.25rem !important;
		text-align: center !important;
	}*/

	.department-icon {
		font-size: 1.5rem;
	}

	.department-name {
		font-weight: 600;
		color: #111827;
		margin: 0;
	}

	.department-description {
		color: #6b7280;
		font-size: 0.875rem;
		margin: 0.25rem 0 0 0;
	}

	.department-stats {
		display: flex;
		justify-content: center;
		align-items: center;
		height: 4vh;
		gap: 0.5rem;
		margin-left: auto;
	}

	.role-count, .user-count {
		font-size: 0.75rem;
		color: #6b7280;
		background: white;
		padding: 0.25rem 0.5rem;
		border-radius: 0.25rem;
	}

	/* Users Tab */
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

	/* Roles Tab */
	.roles-section {
		background: white;
		border-radius: 0.5rem;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
		overflow: hidden;
	}

	.departments-accordion {
		display: flex;
		flex-direction: column;
	}

	.department-accordion {
		border-bottom: 1px solid #e5e7eb;
	}

	.department-header {
		width: 100%;
		justify-content: space-between;
		align-items: center;
		padding: 1.5rem;
		background: none;
		border: none;
		cursor: pointer;
		transition: background-color 0.2s;
	}

	.department-header:hover {
		background: #f9fafb;
	}

	.department-info {
		display: flex;
		align-items: center;
		gap: 1rem;
		flex: 1;
	}

	.department-icon {
		font-size: 1.5rem;
	}

	.department-name {
		font-size: 1.125rem;
		font-weight: 600;
		color: #111827;
		margin: 0;
	}

	.department-description {
		color: #6b7280;
		font-size: 0.875rem;
		margin: 0.25rem 0 0 0;
	}

	.department-stats {
		display: flex;
		gap: 1rem;
		margin-right: 1rem;
	}

	.accordion-icon {
		font-size: 0.875rem;
		color: #6b7280;
	}

	.department-content {
		padding: 0 1.5rem 1.5rem 1.5rem;
		background: #f9fafb;
	}

	.roles-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: 1rem;
	}

	.role-card {
		background: white;
		border-radius: 0.5rem;
		padding: 1rem;
		border: 1px solid #e5e7eb;
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
	}

	.role-header {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		margin-bottom: 1rem;
	}

	.role-icon {
		font-size: 1.5rem;
		flex-shrink: 0;
	}

	.role-info {
		flex: 1;
	}

	.role-name {
		font-weight: 600;
		color: #111827;
		margin: 0 0 0.25rem 0;
	}

	.role-description {
		color: #6b7280;
		font-size: 0.875rem;
		margin: 0;
		line-height: 1.4;
	}

	.system-badge {
		background: #dc2626;
		color: white;
		padding: 0.25rem 0.5rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		font-weight: 500;
	}

	.role-stats {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 0.75rem;
		margin-bottom: 1rem;
	}

	.stat {
		text-align: center;
	}

	.stat-label {
		display: block;
		font-size: 0.75rem;
		color: #6b7280;
		margin-bottom: 0.25rem;
	}

	.stat-value {
		font-weight: 600;
		color: #111827;
	}

	.category-badge {
		padding: 0.25rem 0.5rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		font-weight: 500;
	}

	.role-footer {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding-top: 1rem;
		border-top: 1px solid #e5e7eb;
	}

	.role-updated {
		font-size: 0.75rem;
		color: #6b7280;
	}

	.role-actions {
		display: flex;
		gap: 0.5rem;
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

	.btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.btn-icon {
		font-size: 1rem;
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.admin-page {
			padding: 1rem;
		}

		.page-header {
			flex-direction: column;
			align-items: flex-start;
			gap: 1rem;
		}

		.tab-navigation {
			flex-wrap: wrap;
		}

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

		.stats-grid {
			grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		}

		.roles-grid {
			grid-template-columns: 1fr;
		}

		.role-stats {
			grid-template-columns: 1fr;
		}
	}
</style> 