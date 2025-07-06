<script lang="ts">
	import { onMount } from 'svelte';
	import { writable } from 'svelte/store';
	import { 
		MOCK_STANDARDIZED_ROLES, 
		MOCK_USERS_WITH_ROLES, 
		MOCK_PERMISSIONS, 
		MOCK_PERMISSION_CATEGORIES,
		MOCK_ROLE_AUDIT_LOGS,
		MOCK_ROLE_ANALYTICS,
		searchRoles,
		searchUsers,
		getRolesByCategory,
		getSystemRoles,
		getCustomRoles,
		getRoleHierarchy,
		getUserByEmail,
		hasUserSuperAdminRole
	} from '$lib/mockData/roles';
	import type { Role, UserWithRoles, Permission, PermissionCategory, RoleAuditLog } from '$lib/types/roles';
	import { auth } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import RoleCreateModal from '$lib/components/admin/RoleCreateModal.svelte';
	// import RoleCreateModal from '$lib/components/admin/RoleCreateModal.svelte';

	// Stores
	const roles = writable<Role[]>([]);
	const users = writable<UserWithRoles[]>([]);
	const permissions = writable<Permission[]>([]);
	const permissionCategories = writable<PermissionCategory[]>([]);
	const auditLogs = writable<RoleAuditLog[]>([]);
	const analytics = writable<any[]>([]);

	// UI State
	let activeTab = 'roles';
	let searchQuery = '';
	let selectedCategory = 'all';
	let selectedRole: Role | null = null;
	let selectedUser: UserWithRoles | null = null;
	let showCreateRoleModal = false;
	let showAssignRoleModal = false;
	let showPermissionModal = false;
	let editingRole: Role | null = null;
	let loading = false;

	// Filtered data
	let filteredRoles: Role[] = [];
	let filteredUsers: UserWithRoles[] = [];

	onMount(async () => {
		loading = true;
		try {
			// Load mock data
			roles.set(MOCK_STANDARDIZED_ROLES);
			users.set(MOCK_USERS_WITH_ROLES);
			permissions.set(MOCK_PERMISSIONS);
			permissionCategories.set(MOCK_PERMISSION_CATEGORIES);
			auditLogs.set(MOCK_ROLE_AUDIT_LOGS);
			analytics.set(MOCK_ROLE_ANALYTICS);

			// Initialize filtered data
			filteredRoles = MOCK_STANDARDIZED_ROLES;
			filteredUsers = MOCK_USERS_WITH_ROLES;
		} catch (error) {
			console.error('Error loading role data:', error);
		} finally {
			loading = false;
		}
	});

	// Filter functions
	function filterRoles() {
		let filtered = MOCK_STANDARDIZED_ROLES;

		if (searchQuery.trim()) {
			filtered = searchRoles(searchQuery);
		}

		if (selectedCategory !== 'all') {
			filtered = filtered.filter(role => role.category === selectedCategory);
		}

		filteredRoles = filtered;
	}

	function filterUsers() {
		let filtered = MOCK_USERS_WITH_ROLES;

		if (searchQuery.trim()) {
			filtered = searchUsers(searchQuery);
		}

		filteredUsers = filtered;
	}

	// Event handlers
	function handleSearch() {
		if (activeTab === 'roles') {
			filterRoles();
		} else if (activeTab === 'users') {
			filterUsers();
		}
	}

	function handleCategoryChange() {
		if (activeTab === 'roles') {
			filterRoles();
		}
	}

	function selectRole(role: Role) {
		selectedRole = role;
		showPermissionModal = true;
	}

	function selectUser(user: UserWithRoles) {
		selectedUser = user;
		showAssignRoleModal = true;
	}

	function createNewRole() {
		editingRole = null;
		showCreateRoleModal = true;
	}

	function editRole(role: Role) {
		editingRole = role;
		showCreateRoleModal = true;
	}

	function handleRoleCreated(event: CustomEvent) {
		const newRole = event.detail;
		// Add to mock data (in real app, this would be handled by API)
		MOCK_STANDARDIZED_ROLES.push(newRole);
		filteredRoles = [...filteredRoles, newRole];
		showToast('Role created successfully', 'success');
		showCreateRoleModal = false;
	}

	function handleRoleUpdated(event: CustomEvent) {
		const updatedRole = event.detail;
		// Update mock data (in real app, this would be handled by API)
		const index = MOCK_STANDARDIZED_ROLES.findIndex(r => r.id === updatedRole.id);
		if (index !== -1) {
			MOCK_STANDARDIZED_ROLES[index] = updatedRole;
			filteredRoles = filteredRoles.map(r => r.id === updatedRole.id ? updatedRole : r);
		}
		showToast('Role updated successfully', 'success');
		showCreateRoleModal = false;
	}

	function handleModalClose() {
		showCreateRoleModal = false;
		editingRole = null;
	}

	function deleteRole(role: Role) {
		if (role.isSystemRole) {
			showToast('Cannot delete system roles', 'error');
			return;
		}

		if (confirm(`Are you sure you want to delete the role "${role.name}"?`)) {
			// Remove from mock data (in real app, this would be handled by API)
			const index = MOCK_STANDARDIZED_ROLES.findIndex(r => r.id === role.id);
			if (index !== -1) {
				MOCK_STANDARDIZED_ROLES.splice(index, 1);
				filteredRoles = filteredRoles.filter(r => r.id !== role.id);
			}
			showToast('Role deleted successfully', 'success');
		}
	}

	function getRoleColor(role: Role) {
		return role.color || '#6b7280';
	}

	function getRoleIcon(role: Role) {
		const iconMap: { [key: string]: string } = {
			'crown': '👑',
			'server': '🖥️',
			'document-text': '📄',
			'users': '👥',
			'currency-dollar': '💰',
			'shield-check': '🛡️',
			'chart-bar': '📊',
			'pencil': '✏️',
			'check-circle': '✅',
			'video-camera': '📹',
			'eye': '👁️',
			'megaphone': '📢',
			'presentation-chart-line': '📈',
			'calendar': '📅',
			'clipboard-list': '📋',
			'play': '▶️',
			'support': '🎧',
			'academic-cap': '🎓',
			'book-open': '📖'
		};
		return iconMap[role.icon] || '⚙️';
	}

	function formatDate(dateString: string) {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function getStatusBadgeClass(status: string) {
		const statusClasses: { [key: string]: string } = {
			'active': 'var(--status-success)',
			'inactive': 'var(--status-warning)',
			'suspended': 'var(--status-danger)'
		};
		return statusClasses[status] || 'var(--status-neutral)';
	}

	// Reactive statements
	$: if (searchQuery || selectedCategory) {
		handleSearch();
	}

	$: hierarchyRoles = getRoleHierarchy();
	$: systemRoles = getSystemRoles();
	$: customRoles = getCustomRoles();
</script>

<svelte:head>
	<title>Role Management - BOME Admin</title>
</svelte:head>

<div class="admin-page">
	<div class="page-header">
		<div class="header-content">
			<h1 class="page-title">Role Management</h1>
			<p class="page-description">Manage user roles, permissions, and access control</p>
		</div>
		<div class="header-actions">
			<button class="btn btn-primary" on:click={createNewRole}>
				<span class="btn-icon">➕</span>
				Create Role
			</button>
		</div>
	</div>

	<!-- Current User Role Debug Info -->
	{#if $auth.user}
		<div class="debug-section">
			<h3>Current User Role Information</h3>
			<div class="debug-info">
				<p><strong>Email:</strong> {$auth.user.email}</p>
				<p><strong>Name:</strong> {$auth.user.firstName} {$auth.user.lastName}</p>
				<p><strong>Legacy Role:</strong> {$auth.user.role}</p>
				{#if $auth.user.roles && $auth.user.roles.length > 0}
					<p><strong>Assigned Roles:</strong></p>
					<ul>
						{#each $auth.user.roles as role}
							<li>{role.name} (Level {role.level}) - {role.description}</li>
						{/each}
					</ul>
					<p><strong>Total Permissions:</strong> {$auth.user.roles.reduce((total, role) => total + role.permissions.length, 0)}</p>
				{:else}
					<p><strong>Assigned Roles:</strong> None</p>
				{/if}
				<p><strong>Has Super Admin Role:</strong> {hasUserSuperAdminRole($auth.user.email) ? 'Yes ✅' : 'No ❌'}</p>
			</div>
		</div>
	{/if}

	<!-- Loading State -->
	{#if loading}
		<div class="loading-container">
			<div class="loading-spinner"></div>
			<p>Loading role management data...</p>
		</div>
	{:else}
		<!-- Tab Navigation -->
		<div class="tab-navigation">
			<button 
				class="tab-button" 
				class:active={activeTab === 'roles'}
				on:click={() => activeTab = 'roles'}
			>
				<span class="tab-icon">🎭</span>
				Roles ({MOCK_ROLES.length})
			</button>
			<button 
				class="tab-button" 
				class:active={activeTab === 'users'}
				on:click={() => activeTab = 'users'}
			>
				<span class="tab-icon">👥</span>
				Users ({MOCK_USERS_WITH_ROLES.length})
			</button>
			<button 
				class="tab-button" 
				class:active={activeTab === 'permissions'}
				on:click={() => activeTab = 'permissions'}
			>
				<span class="tab-icon">🔐</span>
				Permissions ({MOCK_PERMISSIONS.length})
			</button>
			<button 
				class="tab-button" 
				class:active={activeTab === 'audit'}
				on:click={() => activeTab = 'audit'}
			>
				<span class="tab-icon">📋</span>
				Audit Log
			</button>
			<button 
				class="tab-button" 
				class:active={activeTab === 'analytics'}
				on:click={() => activeTab = 'analytics'}
			>
				<span class="tab-icon">📊</span>
				Analytics
			</button>
		</div>

		<!-- Search and Filter Bar -->
		<div class="filter-bar">
			<div class="search-container">
				<input
					type="text"
					placeholder="Search {activeTab}..."
					bind:value={searchQuery}
					on:input={handleSearch}
					class="search-input"
				/>
				<span class="search-icon">🔍</span>
			</div>

			{#if activeTab === 'roles'}
				<div class="filter-container">
					<select bind:value={selectedCategory} on:change={handleCategoryChange} class="filter-select">
						<option value="all">All Categories</option>
						<option value="core">Core Admin</option>
						<option value="content">Content & Editorial</option>
						<option value="marketing">Marketing & Advertising</option>
						<option value="events">Events & Community</option>
						<option value="technical">Technical & Support</option>
						<option value="academic">Academic</option>
					</select>
				</div>
			{/if}
		</div>

		<!-- Tab Content -->
		<div class="tab-content">
			{#if activeTab === 'roles'}
				<!-- Roles Tab -->
				<div class="roles-grid">
					{#each filteredRoles as role (role.id)}
						<div class="role-card" on:click={() => selectRole(role)}>
							<div class="role-header">
								<div class="role-icon" style="color: {getRoleColor(role)}">
									{getRoleIcon(role)}
								</div>
								<div class="role-info">
									<h3 class="role-name">{role.name}</h3>
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
									<span class="stat-value category-badge" style="background: {getRoleColor(role)}20">
										{role.category}
									</span>
								</div>
							</div>

							<div class="role-footer">
								<span class="role-updated">Updated {formatDate(role.updatedAt)}</span>
								<div class="role-actions">
									<button class="btn btn-sm btn-secondary" on:click={(e) => { e.stopPropagation(); editRole(role); }}>Edit</button>
									{#if !role.isSystemRole}
										<button class="btn btn-sm btn-danger" on:click={(e) => { e.stopPropagation(); deleteRole(role); }}>Delete</button>
									{/if}
								</div>
							</div>
						</div>
					{/each}
				</div>

			{:else if activeTab === 'users'}
				<!-- Users Tab -->
				<div class="users-table">
					<div class="table-header">
						<div class="table-row">
							<div class="table-cell">User</div>
							<div class="table-cell">Roles</div>
							<div class="table-cell">Status</div>
							<div class="table-cell">Last Login</div>
							<div class="table-cell">Actions</div>
						</div>
					</div>
					<div class="table-body">
						{#each filteredUsers as user (user.id)}
							<div class="table-row" on:click={() => selectUser(user)}>
								<div class="table-cell">
									<div class="user-info">
										<div class="user-avatar">
											{user.firstName[0]}{user.lastName[0]}
										</div>
										<div class="user-details">
											<div class="user-name">{user.firstName} {user.lastName}</div>
											<div class="user-email">{user.email}</div>
										</div>
									</div>
								</div>
								<div class="table-cell">
									<div class="user-roles">
										{#each user.roles as role}
											<span class="role-badge" style="background: {getRoleColor(role)}20; color: {getRoleColor(role)}">
												{getRoleIcon(role)} {role.name}
											</span>
										{/each}
									</div>
								</div>
								<div class="table-cell">
									<span class="status-badge" style="background: {getStatusBadgeClass(user.status)}">
										{user.status}
									</span>
								</div>
								<div class="table-cell">
									{user.lastLogin ? formatDate(user.lastLogin) : 'Never'}
								</div>
								<div class="table-cell">
									<div class="table-actions">
										<button class="btn btn-sm btn-secondary">Edit Roles</button>
										<button class="btn btn-sm btn-primary">View Profile</button>
									</div>
								</div>
							</div>
						{/each}
					</div>
				</div>

			{:else if activeTab === 'permissions'}
				<!-- Permissions Tab -->
				<div class="permissions-grid">
					{#each MOCK_PERMISSION_CATEGORIES as category (category.id)}
						<div class="permission-category">
							<div class="category-header">
								<h3 class="category-name">{category.name}</h3>
								<p class="category-description">{category.description}</p>
								<span class="permission-count">{category.permissions.length} permissions</span>
							</div>
							<div class="permissions-list">
								{#each category.permissions as permission (permission.id)}
									<div class="permission-item">
										<div class="permission-info">
											<span class="permission-name">{permission.resource}:{permission.action}</span>
											<span class="permission-description">{permission.description}</span>
										</div>
										<span class="permission-category-badge">{permission.category}</span>
									</div>
								{/each}
							</div>
						</div>
					{/each}
				</div>

			{:else if activeTab === 'audit'}
				<!-- Audit Log Tab -->
				<div class="audit-log">
					{#each MOCK_ROLE_AUDIT_LOGS as log (log.id)}
						<div class="audit-item">
							<div class="audit-header">
								<span class="audit-action">{log.action}</span>
								<span class="audit-entity">{log.entityType}</span>
								<span class="audit-timestamp">{formatDate(log.timestamp)}</span>
							</div>
							<div class="audit-details">
								<p class="audit-description">
									{#if log.action === 'assign'}
										Assigned role to user
									{:else if log.action === 'create'}
										Created new {log.entityType}
									{:else if log.action === 'permission_change'}
										Modified permissions for {log.entityType}
									{/if}
								</p>
								{#if log.reason}
									<p class="audit-reason">Reason: {log.reason}</p>
								{/if}
							</div>
							<div class="audit-metadata">
								<span class="audit-ip">IP: {log.ipAddress}</span>
								<span class="audit-user">User ID: {log.userId}</span>
							</div>
						</div>
					{/each}
				</div>

			{:else if activeTab === 'analytics'}
				<!-- Analytics Tab -->
				<div class="analytics-grid">
					{#each MOCK_ROLE_ANALYTICS as analytic (analytic.roleId)}
						{@const role = MOCK_ROLES.find(r => r.id === analytic.roleId)}
						{#if role}
							<div class="analytics-card">
								<div class="analytics-header">
									<div class="role-info">
										<span class="role-icon">{getRoleIcon(role)}</span>
										<h3 class="role-name">{role.name}</h3>
									</div>
									<span class="active-users">{analytic.activeUsers} active</span>
								</div>
								
								<div class="analytics-stats">
									<div class="stat-row">
										<span class="stat-label">Total Assignments</span>
										<span class="stat-value">{analytic.totalAssignments}</span>
									</div>
									<div class="stat-row">
										<span class="stat-label">Avg Session Duration</span>
										<span class="stat-value">{analytic.averageSessionDuration}min</span>
									</div>
									<div class="stat-row">
										<span class="stat-label">Security Incidents</span>
										<span class="stat-value">{analytic.securityIncidents}</span>
									</div>
								</div>

								<div class="permission-usage">
									<h4>Most Used Permissions</h4>
									<div class="permission-tags">
										{#each analytic.mostUsedPermissions.slice(0, 3) as permission}
											<span class="permission-tag">{permission}</span>
										{/each}
									</div>
								</div>
							</div>
						{/if}
					{/each}
				</div>
			{/if}
		</div>
	{/if}
</div>

<!-- Role Create/Edit Modal -->
<RoleCreateModal 
	bind:isOpen={showCreateRoleModal}
	{editingRole}
	on:roleCreated={handleRoleCreated}
	on:roleUpdated={handleRoleUpdated}
	on:close={handleModalClose}
/>

<style>
	/* Main Layout */
	.admin-page {
		display: flex;
		flex-direction: column;
		gap: var(--space-xl);
		padding: 0; /* Remove padding since admin layout handles it */
	}

	/* Page Header */
	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: var(--space-lg);
		padding-bottom: var(--space-lg);
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
	}

	.header-content h1 {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
	}

	.header-content p {
		color: var(--text-secondary);
		font-size: var(--text-base);
		margin: 0;
		line-height: 1.5;
	}

	.header-actions {
		display: flex;
		gap: var(--space-md);
		flex-shrink: 0;
	}

	/* Debug Section */
	.debug-section {
		background: var(--bg-glass);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		backdrop-filter: blur(10px);
	}

	.debug-section h3 {
		color: var(--primary);
		margin: 0 0 var(--space-lg) 0;
		font-size: var(--text-xl);
		font-weight: 600;
	}

	.debug-info {
		font-size: var(--text-sm);
		line-height: 1.6;
	}

	.debug-info p {
		margin: var(--space-md) 0;
		color: var(--text-primary);
	}

	.debug-info ul {
		margin: var(--space-md) 0;
		padding-left: var(--space-lg);
	}

	.debug-info li {
		color: var(--text-secondary);
		margin: var(--space-sm) 0;
	}

	.debug-info strong {
		color: var(--text-primary);
		font-weight: 600;
	}

	/* Loading State */
	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--space-4xl);
		color: var(--text-secondary);
		gap: var(--space-lg);
	}

	.loading-spinner {
		width: 40px;
		height: 40px;
		border: 3px solid rgba(255, 255, 255, 0.1);
		border-top: 3px solid var(--primary);
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	/* Tab Navigation */
	.tab-navigation {
		display: flex;
		gap: var(--space-xs);
		margin-bottom: var(--space-xl);
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-sm);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.tab-button {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		padding: var(--space-md) var(--space-lg);
		background: transparent;
		border: none;
		color: var(--text-secondary);
		cursor: pointer;
		border-radius: var(--radius-lg);
		transition: all var(--transition-normal);
		font-size: var(--text-sm);
		font-weight: 500;
		white-space: nowrap;
	}

	.tab-button:hover {
		background: rgba(255, 255, 255, 0.05);
		color: var(--text-primary);
	}

	.tab-button.active {
		background: var(--primary-gradient);
		color: var(--white);
		box-shadow: var(--shadow-md);
	}

	.tab-icon {
		font-size: var(--text-base);
	}

	/* Filter Bar */
	.filter-bar {
		display: flex;
		gap: var(--space-lg);
		margin-bottom: var(--space-xl);
		align-items: center;
		flex-wrap: wrap;
	}

	.search-container {
		position: relative;
		flex: 1;
		min-width: 300px;
		max-width: 500px;
	}

	.search-input {
		width: 100%;
		padding: var(--space-md) var(--space-4xl) var(--space-md) var(--space-lg);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-xl);
		background: var(--bg-glass);
		color: var(--text-primary);
		font-size: var(--text-sm);
		transition: all var(--transition-normal);
		backdrop-filter: blur(10px);
	}

	.search-input:focus {
		outline: none;
		border-color: var(--primary);
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	.search-input::placeholder {
		color: var(--text-secondary);
	}

	.search-icon {
		position: absolute;
		right: var(--space-lg);
		top: 50%;
		transform: translateY(-50%);
		color: var(--text-secondary);
		width: 20px;
		height: 20px;
	}

	.filter-select {
		padding: var(--space-md) var(--space-lg);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-xl);
		background: var(--bg-glass);
		color: var(--text-primary);
		font-size: var(--text-sm);
		cursor: pointer;
		transition: all var(--transition-normal);
		backdrop-filter: blur(10px);
		min-width: 150px;
	}

	.filter-select:focus {
		outline: none;
		border-color: var(--primary);
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	option {
		color: var(--gray-900);
	}

	/* Tab Content */
	.tab-content {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		backdrop-filter: blur(10px);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	/* Roles Grid */
	.roles-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
		gap: var(--space-xl);
	}

	.role-card {
		background: var(--bg-glass-dark);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		cursor: pointer;
		transition: all var(--transition-normal);
		backdrop-filter: blur(10px);
	}

	.role-card:hover {
		transform: translateY(-4px);
		box-shadow: var(--shadow-lg);
		border-color: var(--primary);
		background: var(--bg-glass);
	}

	.role-header {
		display: flex;
		align-items: flex-start;
		gap: var(--space-lg);
		margin-bottom: var(--space-lg);
		position: relative;
	}

	.role-icon {
		font-size: var(--text-2xl);
		width: 56px;
		height: 56px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: var(--radius-lg);
		background: var(--primary-gradient);
		color: var(--white);
		flex-shrink: 0;
		box-shadow: var(--shadow-md);
	}

	.role-info {
		flex: 1;
		min-width: 0;
	}

	.role-name {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
		line-height: 1.3;
	}

	.role-description {
		color: var(--text-secondary);
		font-size: var(--text-sm);
		margin: 0;
		line-height: 1.5;
	}

	.system-badge {
		position: absolute;
		top: 0;
		right: 0;
		background: var(--primary-gradient);
		color: var(--white);
		padding: var(--space-sm) var(--space-md);
		border-radius: var(--radius-md);
		font-size: var(--text-xs);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		box-shadow: var(--shadow-sm);
	}

	.role-stats {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: var(--space-lg);
		margin-bottom: var(--space-lg);
		padding: var(--space-lg);
		background: var(--bg-glass);
		border-radius: var(--radius-lg);
		border: 1px solid rgba(255, 255, 255, 0.05);
	}

	.stat {
		text-align: center;
	}

	.stat-label {
		display: block;
		font-size: var(--text-xs);
		color: var(--text-secondary);
		margin-bottom: var(--space-sm);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-weight: 500;
	}

	.stat-value {
		font-size: var(--text-lg);
		font-weight: 700;
		color: var(--text-primary);
	}

	.category-badge {
		padding: var(--space-sm) var(--space-md);
		border-radius: var(--radius-md);
		font-size: var(--text-xs);
		font-weight: 600;
		text-transform: capitalize;
		background: var(--bg-glass);
		color: var(--text-secondary);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.role-footer {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding-top: var(--space-lg);
		border-top: 1px solid rgba(255, 255, 255, 0.1);
	}

	.role-updated {
		font-size: var(--text-xs);
		color: var(--text-secondary);
	}

	.role-actions {
		display: flex;
		gap: var(--space-sm);
	}

	/* Users Table */
	.users-table {
		background: var(--bg-glass-dark);
		border-radius: var(--radius-xl);
		overflow: hidden;
		border: 1px solid rgba(255, 255, 255, 0.1);
		backdrop-filter: blur(10px);
	}

	.table-header {
		background: var(--bg-glass);
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
	}

	.table-row {
		display: grid;
		grid-template-columns: 2fr 2fr 1fr 1.5fr 1.5fr;
		gap: var(--space-lg);
		padding: var(--space-lg);
		align-items: center;
	}

	.table-body .table-row {
		border-bottom: 1px solid rgba(255, 255, 255, 0.05);
		cursor: pointer;
		transition: all var(--transition-normal);
	}

	.table-body .table-row:hover {
		background: var(--bg-glass);
	}

	.table-body .table-row:last-child {
		border-bottom: none;
	}

	.table-cell {
		font-size: var(--text-sm);
		color: var(--text-primary);
	}

	.table-header .table-cell {
		font-weight: 600;
		color: var(--text-primary);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-size: var(--text-xs);
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
		color: var(--white);
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 600;
		font-size: var(--text-sm);
		flex-shrink: 0;
		box-shadow: var(--shadow-sm);
	}

	.user-name {
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-xs);
		font-size: var(--text-sm);
	}

	.user-email {
		color: var(--text-secondary);
		font-size: var(--text-xs);
	}

	.user-roles {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-xs);
	}

	.role-badge {
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-sm);
		font-size: var(--text-xs);
		font-weight: 500;
		display: flex;
		align-items: center;
		gap: var(--space-xs);
		background: var(--primary-gradient);
		color: var(--white);
		box-shadow: var(--shadow-sm);
	}

	.status-badge {
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-sm);
		font-size: var(--text-xs);
		font-weight: 600;
		text-transform: capitalize;
		color: var(--white);
	}

	.status-badge.active {
		background: var(--success);
	}

	.status-badge.inactive {
		background: var(--warning);
	}

	.table-actions {
		display: flex;
		gap: var(--space-sm);
		justify-content: flex-end;
	}

	/* Permissions Grid */
	.permissions-grid {
		display: grid;
		gap: var(--space-xl);
	}

	.permission-category {
		background: var(--bg-glass-dark);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-xl);
		overflow: hidden;
		backdrop-filter: blur(10px);
	}

	.category-header {
		padding: var(--space-xl);
		background: var(--bg-glass);
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
	}

	.category-name {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
	}

	.category-description {
		color: var(--text-secondary);
		margin: 0 0 var(--space-md) 0;
		line-height: 1.5;
	}

	.permission-count {
		background: var(--primary-gradient);
		color: var(--white);
		padding: var(--space-sm) var(--space-md);
		border-radius: var(--radius-md);
		font-size: var(--text-xs);
		font-weight: 600;
		box-shadow: var(--shadow-sm);
	}

	.permissions-list {
		padding: var(--space-lg);
	}

	.permission-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-lg);
		border-bottom: 1px solid rgba(255, 255, 255, 0.05);
		transition: all var(--transition-normal);
	}

	.permission-item:hover {
		background: var(--bg-glass);
	}

	.permission-item:last-child {
		border-bottom: none;
	}

	.permission-name {
		font-weight: 600;
		color: var(--text-primary);
		font-family: var(--font-mono);
		font-size: var(--text-sm);
	}

	.permission-description {
		color: var(--text-secondary);
		font-size: var(--text-sm);
		margin-top: var(--space-xs);
		line-height: 1.4;
	}

	.permission-category-badge {
		background: var(--bg-glass);
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-sm);
		font-size: var(--text-xs);
		color: var(--text-secondary);
		text-transform: capitalize;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	/* Audit Log */
	.audit-log {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.audit-item {
		background: var(--bg-glass-dark);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		backdrop-filter: blur(10px);
	}

	.audit-header {
		display: flex;
		gap: var(--space-lg);
		align-items: center;
		margin-bottom: var(--space-lg);
		flex-wrap: wrap;
	}

	.audit-action {
		background: var(--primary-gradient);
		color: var(--white);
		padding: var(--space-sm) var(--space-md);
		border-radius: var(--radius-md);
		font-size: var(--text-xs);
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		box-shadow: var(--shadow-sm);
	}

	.audit-entity {
		background: var(--bg-glass);
		padding: var(--space-sm) var(--space-md);
		border-radius: var(--radius-md);
		font-size: var(--text-xs);
		color: var(--text-secondary);
		border: 1px solid rgba(255, 255, 255, 0.1);
		font-family: var(--font-mono);
	}

	.audit-timestamp {
		color: var(--text-secondary);
		font-size: var(--text-xs);
		margin-left: auto;
		font-family: var(--font-mono);
	}

	.audit-details {
		margin-bottom: var(--space-lg);
	}

	.audit-description {
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
		font-weight: 500;
	}

	.audit-reason {
		color: var(--text-secondary);
		font-style: italic;
		margin: 0;
		font-size: var(--text-sm);
	}

	.audit-metadata {
		display: flex;
		gap: var(--space-lg);
		font-size: var(--text-xs);
		color: var(--text-secondary);
		padding-top: var(--space-lg);
		border-top: 1px solid rgba(255, 255, 255, 0.1);
		font-family: var(--font-mono);
		flex-wrap: wrap;
	}

	/* Analytics Grid */
	.analytics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
		gap: var(--space-xl);
	}

	.analytics-card {
		background: var(--bg-glass-dark);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		backdrop-filter: blur(10px);
	}

	.analytics-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-lg);
	}

	.analytics-header .role-info {
		display: flex;
		align-items: center;
		gap: var(--space-md);
	}

	.analytics-header .role-name {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.active-users {
		background: var(--success);
		color: var(--white);
		padding: var(--space-sm) var(--space-md);
		border-radius: var(--radius-md);
		font-size: var(--text-xs);
		font-weight: 600;
		box-shadow: var(--shadow-sm);
	}

	.analytics-stats {
		margin-bottom: var(--space-lg);
	}

	.stat-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-md) 0;
		border-bottom: 1px solid rgba(255, 255, 255, 0.05);
		font-size: var(--text-sm);
	}

	.stat-row:last-child {
		border-bottom: none;
	}

	.permission-usage h4 {
		font-size: var(--text-base);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-md) 0;
	}

	.permission-tags {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-sm);
	}

	.permission-tag {
		background: var(--bg-glass);
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-sm);
		font-size: var(--text-xs);
		color: var(--text-secondary);
		font-family: var(--font-mono);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	/* Button Styles */
	.btn {
		padding: var(--space-md) var(--space-lg);
		border: none;
		border-radius: var(--radius-lg);
		font-size: var(--text-sm);
		font-weight: 600;
		cursor: pointer;
		transition: all var(--transition-normal);
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		text-decoration: none;
		justify-content: center;
		white-space: nowrap;
	}

	.btn-sm {
		padding: var(--space-sm) var(--space-md);
		font-size: var(--text-xs);
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
	}

	.btn-danger {
		background: var(--error);
		color: var(--white);
		box-shadow: var(--shadow-md);
	}

	.btn-danger:hover {
		background: #dc2626;
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}

	.btn-icon {
		font-size: var(--text-base);
	}

	/* Responsive Design */
	@media (max-width: 1200px) {
		.roles-grid {
			grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
		}

		.analytics-grid {
			grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		}
	}

	@media (max-width: 768px) {
		.page-header {
			flex-direction: column;
			gap: var(--space-lg);
			align-items: stretch;
		}

		.header-actions {
			justify-content: flex-start;
		}

		.tab-navigation {
			flex-wrap: wrap;
			gap: var(--space-xs);
		}

		.filter-bar {
			flex-direction: column;
			align-items: stretch;
			gap: var(--space-md);
		}

		.search-container {
			max-width: none;
			min-width: 0;
		}

		.roles-grid {
			grid-template-columns: 1fr;
		}

		.table-row {
			grid-template-columns: 1fr;
			gap: var(--space-md);
		}

		.table-header {
			display: none;
		}

		.table-cell {
			display: flex;
			justify-content: space-between;
			align-items: center;
		}

		.table-cell::before {
			content: attr(data-label);
			font-weight: 600;
			color: var(--text-secondary);
			text-transform: uppercase;
			font-size: var(--text-xs);
			letter-spacing: 0.5px;
		}

		.analytics-grid {
			grid-template-columns: 1fr;
		}

		.role-stats {
			grid-template-columns: 1fr;
			gap: var(--space-md);
		}

		.audit-header {
			flex-direction: column;
			align-items: flex-start;
			gap: var(--space-md);
		}

		.audit-timestamp {
			margin-left: 0;
		}

		.audit-metadata {
			flex-direction: column;
			gap: var(--space-sm);
		}
	}

	@media (max-width: 480px) {
		.tab-content {
			padding: var(--space-lg);
		}

		.role-card {
			padding: var(--space-lg);
		}

		.analytics-card {
			padding: var(--space-lg);
		}

		.audit-item {
			padding: var(--space-lg);
		}

		.permission-category .category-header {
			padding: var(--space-lg);
		}

		.permissions-list {
			padding: var(--space-md);
		}

		.permission-item {
			padding: var(--space-md);
			flex-direction: column;
			align-items: flex-start;
			gap: var(--space-sm);
		}
	}
</style> 