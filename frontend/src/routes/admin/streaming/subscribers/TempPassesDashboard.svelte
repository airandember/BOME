<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import { subscriberStore, subscriberStoreActions } from '$lib/stores/subscribers-store';
	import type { EnhancedSubscriber } from '$lib/types/enhanced-subscriber';

	// Types
	interface TempPasswordUser {
		id: number;
		email: string;
		first_name: string;
		last_name: string;
		temp_password: string;
		temp_password_created_at: string;
		created_at: string;
		has_logged_in: boolean;
	}

	interface TempPasswordResult {
		user_id: number;
		email: string;
		temp_password: string;
		email_sent: boolean;
		error?: string;
	}

	// Tab state
	let activeTab: 'eligible' | 'assigned' = $state('eligible');

	// === ELIGIBLE TAB STATE ===
	let eligibleLoading = $state(false);
	let eligibleUsers = $state<EnhancedSubscriber[]>([]);
	let filteredEligibleUsers = $state<EnhancedSubscriber[]>([]);
	let selectedUserIds = $state<Set<number>>(new Set());
	let eligibleSearchTerm = $state('');
	let assigningTempPasswords = $state(false);
	let tempPasswordResults = $state<TempPasswordResult[]>([]);
	let showResultsModal = $state(false);

	// Eligible pagination
	let eligibleCurrentPage = $state(1);
	let eligibleItemsPerPage = $state(50);
	let eligibleTotalPages = $derived(Math.ceil(filteredEligibleUsers.length / eligibleItemsPerPage));
	let paginatedEligibleUsers = $derived(() => {
		const start = (eligibleCurrentPage - 1) * eligibleItemsPerPage;
		return filteredEligibleUsers.slice(start, start + eligibleItemsPerPage);
	});

	// === ASSIGNED TAB STATE ===
	let assignedLoading = $state(false);
	let assignedUsers = $state<TempPasswordUser[]>([]);
	let filteredAssignedUsers = $state<TempPasswordUser[]>([]);
	let assignedSearchTerm = $state('');
	let deactivatingUserId = $state<number | null>(null);

	// Assigned pagination
	let assignedCurrentPage = $state(1);
	let assignedItemsPerPage = $state(50);
	let assignedTotalPages = $derived(Math.ceil(filteredAssignedUsers.length / assignedItemsPerPage));
	let paginatedAssignedUsers = $derived(() => {
		const start = (assignedCurrentPage - 1) * assignedItemsPerPage;
		return filteredAssignedUsers.slice(start, start + assignedItemsPerPage);
	});

	// Set of user IDs that have temp passwords (for quick lookup)
	let tempPasswordUserIds = $derived(() => {
		return new Set(assignedUsers.map(u => u.id));
	});

	// === STATS ===
	let eligibleStats = $derived(() => {
		const total = eligibleUsers.length;
		const withTempPass = eligibleUsers.filter(u => tempPasswordUserIds().has(u.id)).length;
		const needsAssignment = total - withTempPass;
		return { total, withTempPass, needsAssignment };
	});

	let assignedStats = $derived(() => {
		const total = assignedUsers.length;
		const loggedIn = assignedUsers.filter(u => u.has_logged_in).length;
		const pending = total - loggedIn;
		return { total, loggedIn, pending };
	});

	// === SELECTION HELPERS ===
	let allVisibleSelected = $derived(() => {
		const visible = paginatedEligibleUsers();
		return visible.length > 0 && visible.every(u => selectedUserIds.has(u.id));
	});

	let someVisibleSelected = $derived(() => {
		const visible = paginatedEligibleUsers();
		return visible.some(u => selectedUserIds.has(u.id)) && !allVisibleSelected();
	});

	// === LOAD FUNCTIONS ===
	async function loadEligibleUsers() {
		eligibleLoading = true;
		try {
			// Load from subscriber store if not already loaded
			const storeData = $subscriberStore;
			if (storeData.subscribers.length === 0) {
				await subscriberStoreActions.loadSubscribers();
			}
			
			// Filter for eligible users: video access + active plan + not verified
			const allSubscribers = $subscriberStore.subscribers;
			eligibleUsers = allSubscribers.filter(sub => 
				sub.has_video_access && 
				sub.has_active_plan && 
				!sub.email_verified
			);
			
			applyEligibleFilters();
			console.log(`📋 Found ${eligibleUsers.length} eligible users for temp passwords`);
		} catch (error) {
			console.error('Error loading eligible users:', error);
			showToast('Failed to load eligible users', 'error');
			eligibleUsers = [];
			filteredEligibleUsers = [];
		} finally {
			eligibleLoading = false;
		}
	}

	async function loadAssignedUsers() {
		assignedLoading = true;
		try {
			const response = await apiRequest('/admin/users/with-temp-passwords');
			if (response.ok) {
				const data = await response.json();
				assignedUsers = data.users || [];
				applyAssignedFilters();
				console.log(`🔑 Loaded ${assignedUsers.length} users with temp passwords`);
			} else {
				throw new Error('Failed to load users');
			}
		} catch (error) {
			console.error('Error loading assigned users:', error);
			showToast('Failed to load assigned users', 'error');
			assignedUsers = [];
			filteredAssignedUsers = [];
		} finally {
			assignedLoading = false;
		}
	}

	// === FILTER FUNCTIONS ===
	function applyEligibleFilters() {
		let result = [...eligibleUsers];

		if (eligibleSearchTerm.trim()) {
			const term = eligibleSearchTerm.toLowerCase();
			result = result.filter(u => 
				u.email.toLowerCase().includes(term) ||
				(u.first_name || '').toLowerCase().includes(term) ||
				(u.last_name || '').toLowerCase().includes(term) ||
				u.id.toString().includes(term)
			);
		}

		filteredEligibleUsers = result;
		eligibleCurrentPage = 1;
	}

	function applyAssignedFilters() {
		let result = [...assignedUsers];

		if (assignedSearchTerm.trim()) {
			const term = assignedSearchTerm.toLowerCase();
			result = result.filter(u => 
				u.email.toLowerCase().includes(term) ||
				(u.first_name || '').toLowerCase().includes(term) ||
				(u.last_name || '').toLowerCase().includes(term) ||
				u.id.toString().includes(term)
			);
		}

		filteredAssignedUsers = result;
		assignedCurrentPage = 1;
	}

	// Watch for filter changes
	$effect(() => {
		eligibleSearchTerm;
		applyEligibleFilters();
	});

	$effect(() => {
		assignedSearchTerm;
		applyAssignedFilters();
	});

	// === SELECTION FUNCTIONS ===
	function toggleSelectAll() {
		const visible = paginatedEligibleUsers();
		if (allVisibleSelected()) {
			visible.forEach(u => selectedUserIds.delete(u.id));
		} else {
			visible.forEach(u => selectedUserIds.add(u.id));
		}
		selectedUserIds = new Set(selectedUserIds);
	}

	function toggleSelect(userId: number) {
		if (selectedUserIds.has(userId)) {
			selectedUserIds.delete(userId);
		} else {
			selectedUserIds.add(userId);
		}
		selectedUserIds = new Set(selectedUserIds);
	}

	function clearSelection() {
		selectedUserIds = new Set();
	}

	function selectAll() {
		filteredEligibleUsers.forEach(u => selectedUserIds.add(u.id));
		selectedUserIds = new Set(selectedUserIds);
	}

	// === TEMP PASSWORD ASSIGNMENT ===
	async function assignTempPasswords(sendEmail: boolean) {
		if (selectedUserIds.size === 0) {
			showToast('Please select users to assign temp passwords', 'warning');
			return;
		}

		assigningTempPasswords = true;
		try {
			const userIds = Array.from(selectedUserIds);
			console.log('🔑 Assigning temp passwords to', userIds.length, 'users, sendEmail:', sendEmail);

			const response = await apiRequest('/admin/users/bulk-temp-password', {
				method: 'POST',
				body: JSON.stringify({
					user_ids: userIds,
					send_email: sendEmail
				})
			});

			if (response.ok) {
				const data = await response.json();
				tempPasswordResults = data.results || [];

				if (data.assigned_count > 0) {
					showToast(`Assigned temp passwords to ${data.assigned_count} users`, 'success');
				}
				if (data.error_count > 0) {
					showToast(`${data.error_count} users could not be assigned`, 'warning');
				}

				showResultsModal = true;
				clearSelection();
				
				// Refresh both lists
				await Promise.all([loadEligibleUsers(), loadAssignedUsers()]);
			} else {
				const error = await response.json();
				showToast(error.error || 'Failed to assign temp passwords', 'error');
			}
		} catch (error) {
			console.error('Error assigning temp passwords:', error);
			showToast('Failed to assign temp passwords', 'error');
		} finally {
			assigningTempPasswords = false;
		}
	}

	function closeResultsModal() {
		showResultsModal = false;
		tempPasswordResults = [];
	}

	// === HELPERS ===
	function formatDate(dateStr: string | null): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	async function copyToClipboard(text: string) {
		try {
			await navigator.clipboard.writeText(text);
			showToast('Copied to clipboard', 'success');
		} catch {
			showToast('Failed to copy', 'error');
		}
	}

	async function deactivateTempPassword(user: TempPasswordUser) {
		if (!confirm(`Are you sure you want to deactivate the temp password for ${user.email}?`)) {
			return;
		}
		
		deactivatingUserId = user.id;
		try {
			const response = await apiRequest(`/admin/users/${user.id}/temp-password`, {
				method: 'DELETE'
			});
			
			if (response.ok) {
				showToast(`Temp password deactivated for ${user.email}`, 'success');
				// Refresh both lists
				await Promise.all([loadEligibleUsers(), loadAssignedUsers()]);
			} else {
				const error = await response.json();
				showToast(error.error || 'Failed to deactivate temp password', 'error');
			}
		} catch (error) {
			console.error('Error deactivating temp password:', error);
			showToast('Failed to deactivate temp password', 'error');
		} finally {
			deactivatingUserId = null;
		}
	}

	// === LIFECYCLE ===
	onMount(() => {
		loadEligibleUsers();
		loadAssignedUsers();
	});
</script>

<div class="temp-passes-dashboard">
	<!-- Header -->
	<div class="dashboard-header">
		<div class="header-content">
			<h2>Temp Password Management</h2>
			<p>Assign temporary passwords to users with active subscriptions who haven't verified their email.</p>
		</div>
	</div>

	<!-- Tabs -->
	<div class="tabs">
		<button 
			class="tab-btn {activeTab === 'eligible' ? 'active' : ''}"
			onclick={() => activeTab = 'eligible'}
		>
			Eligible Users
			<span class="tab-badge">{eligibleUsers.length}</span>
		</button>
		<button 
			class="tab-btn {activeTab === 'assigned' ? 'active' : ''}"
			onclick={() => activeTab = 'assigned'}
		>
			Assigned
			<span class="tab-badge">{assignedUsers.length}</span>
		</button>
	</div>

	<!-- ELIGIBLE TAB -->
	{#if activeTab === 'eligible'}
		<div class="tab-content">
			<!-- Stats -->
			<div class="stats-grid">
				<div class="stat-card">
					<div class="stat-icon">👥</div>
					<div class="stat-content">
						<div class="stat-value">{eligibleStats().total}</div>
						<div class="stat-label">Total Eligible</div>
					</div>
				</div>
				<div class="stat-card highlight-warning">
					<div class="stat-icon">⚠️</div>
					<div class="stat-content">
						<div class="stat-value">{eligibleStats().needsAssignment}</div>
						<div class="stat-label">Needs Assignment</div>
					</div>
				</div>
				<div class="stat-card">
					<div class="stat-icon">🔑</div>
					<div class="stat-content">
						<div class="stat-value">{eligibleStats().withTempPass}</div>
						<div class="stat-label">Has Temp Pass</div>
					</div>
				</div>
				<div class="stat-card highlight-info">
					<div class="stat-icon">✅</div>
					<div class="stat-content">
						<div class="stat-value">{selectedUserIds.size}</div>
						<div class="stat-label">Selected</div>
					</div>
				</div>
			</div>

			<!-- Criteria Info -->
			<div class="criteria-info">
				<strong>Showing users with:</strong> Video Access + Active Plan + Email Not Verified
			</div>

			<!-- Filters & Actions -->
			<div class="filters-section">
				<div class="filter-row">
					<div class="filter-group search-group">
						<label for="eligible-search">Search</label>
						<input
							id="eligible-search"
							type="text"
							bind:value={eligibleSearchTerm}
							placeholder="Search by email, name, or ID..."
							class="filter-input"
						/>
					</div>

					<div class="bulk-actions">
						<button 
							class="btn btn-outline" 
							onclick={selectAll}
							disabled={filteredEligibleUsers.length === 0}
						>
							Select All ({filteredEligibleUsers.length})
						</button>
						<button 
							class="btn btn-outline" 
							onclick={clearSelection}
							disabled={selectedUserIds.size === 0}
						>
							Clear Selection
						</button>
						<button 
							class="btn btn-primary"
							onclick={() => assignTempPasswords(false)}
							disabled={selectedUserIds.size === 0 || assigningTempPasswords}
						>
							{assigningTempPasswords ? '⏳' : '🔑'} Assign Only ({selectedUserIds.size})
						</button>
						<button 
							class="btn btn-primary"
							onclick={() => assignTempPasswords(true)}
							disabled={selectedUserIds.size === 0 || assigningTempPasswords}
						>
							{assigningTempPasswords ? '⏳' : '📧'} Assign + Email ({selectedUserIds.size})
						</button>
					</div>
				</div>

				<div class="filter-summary">
					<span>Showing {filteredEligibleUsers.length} of {eligibleUsers.length} eligible users</span>
					<button class="btn btn-sm btn-secondary" onclick={loadEligibleUsers} disabled={eligibleLoading}>
						{eligibleLoading ? '⏳' : '🔄'} Refresh
					</button>
				</div>
			</div>

			<!-- Users Table -->
			<div class="table-container">
				{#if eligibleLoading}
					<div class="loading-container">
						<LoadingSpinner />
						<p>Loading eligible users...</p>
					</div>
				{:else if filteredEligibleUsers.length === 0}
					<div class="empty-state">
						<p>No eligible users found.</p>
						<p class="hint">Eligible users have: Video Access + Active Plan + Email Not Verified</p>
					</div>
				{:else}
					<table class="users-table">
						<thead>
							<tr>
								<th class="checkbox-col">
									<input 
										type="checkbox" 
										checked={allVisibleSelected()}
										indeterminate={someVisibleSelected()}
										onchange={toggleSelectAll}
									/>
								</th>
								<th>User</th>
								<th>ID</th>
								<th>Plan</th>
								<th>Video Access</th>
								<th>Active Plan</th>
								<th>Verified</th>
								<th>Temp Pass Active</th>
								<th>Last Login</th>
							</tr>
						</thead>
						<tbody>
							{#each paginatedEligibleUsers() as user}
								<tr class:selected={selectedUserIds.has(user.id)}>
									<td class="checkbox-col">
										<input 
											type="checkbox" 
											checked={selectedUserIds.has(user.id)}
											onchange={() => toggleSelect(user.id)}
										/>
									</td>
									<td class="user-cell">
										<div class="user-email">{user.email}</div>
										{#if user.first_name || user.last_name}
											<div class="user-name">{user.first_name} {user.last_name}</div>
										{/if}
									</td>
									<td class="id-cell">{user.id}</td>
									<td>{user.plan_name || '-'}</td>
									<td>
										<span class="badge badge-success">Yes</span>
									</td>
									<td>
										<span class="badge badge-success">Yes</span>
									</td>
									<td>
										<span class="badge badge-warning">No</span>
									</td>
									<td>
										{#if tempPasswordUserIds().has(user.id)}
											<span class="badge badge-success">Yes</span>
										{:else}
											<span class="badge badge-warning">No</span>
										{/if}
									</td>
									<td class="date-cell">
										{#if user.last_login}
											{formatDate(user.last_login)}
										{:else}
											<span class="badge badge-warning">Never</span>
										{/if}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>

					<!-- Pagination -->
					{#if eligibleTotalPages > 1}
						<div class="pagination">
							<button 
								class="btn btn-sm" 
								disabled={eligibleCurrentPage === 1}
								onclick={() => eligibleCurrentPage = eligibleCurrentPage - 1}
							>
								← Previous
							</button>
							<span class="page-info">Page {eligibleCurrentPage} of {eligibleTotalPages}</span>
							<button 
								class="btn btn-sm" 
								disabled={eligibleCurrentPage === eligibleTotalPages}
								onclick={() => eligibleCurrentPage = eligibleCurrentPage + 1}
							>
								Next →
							</button>
						</div>
					{/if}
				{/if}
			</div>
		</div>
	{/if}

	<!-- ASSIGNED TAB -->
	{#if activeTab === 'assigned'}
		<div class="tab-content">
			<!-- Stats -->
			<div class="stats-grid">
				<div class="stat-card">
					<div class="stat-icon">🔑</div>
					<div class="stat-content">
						<div class="stat-value">{assignedStats().total}</div>
						<div class="stat-label">Total Assigned</div>
					</div>
				</div>
				<div class="stat-card">
					<div class="stat-icon">✅</div>
					<div class="stat-content">
						<div class="stat-value">{assignedStats().loggedIn}</div>
						<div class="stat-label">Have Logged In</div>
					</div>
				</div>
				<div class="stat-card highlight-warning">
					<div class="stat-icon">⏳</div>
					<div class="stat-content">
						<div class="stat-value">{assignedStats().pending}</div>
						<div class="stat-label">Pending Login</div>
					</div>
				</div>
			</div>

			<!-- Filters -->
			<div class="filters-section">
				<div class="filter-row">
					<div class="filter-group search-group">
						<label for="assigned-search">Search</label>
						<input
							id="assigned-search"
							type="text"
							bind:value={assignedSearchTerm}
							placeholder="Search by email, name, or ID..."
							class="filter-input"
						/>
					</div>
				</div>

				<div class="filter-summary">
					<span>Showing {filteredAssignedUsers.length} of {assignedUsers.length} assigned users</span>
					<button class="btn btn-sm btn-secondary" onclick={loadAssignedUsers} disabled={assignedLoading}>
						{assignedLoading ? '⏳' : '🔄'} Refresh
					</button>
				</div>
			</div>

			<!-- Users Table -->
			<div class="table-container">
				{#if assignedLoading}
					<div class="loading-container">
						<LoadingSpinner />
						<p>Loading assigned users...</p>
					</div>
				{:else if filteredAssignedUsers.length === 0}
					<div class="empty-state">
						<p>No users with temp passwords found.</p>
						<p class="hint">Assign temp passwords from the Eligible tab.</p>
					</div>
				{:else}
					<table class="users-table">
						<thead>
							<tr>
								<th>User</th>
								<th>ID</th>
								<th>Temp Password</th>
								<th>Assigned</th>
								<th>Status</th>
								<th>Actions</th>
							</tr>
						</thead>
						<tbody>
							{#each paginatedAssignedUsers() as user}
								<tr class:logged-in={user.has_logged_in}>
									<td class="user-cell">
										<div class="user-email">{user.email}</div>
										{#if user.first_name || user.last_name}
											<div class="user-name">{user.first_name} {user.last_name}</div>
										{/if}
									</td>
									<td class="id-cell">{user.id}</td>
									<td class="temp-pass-cell">
										<span class="badge badge-info">
											{user.temp_password}
										</span>
										<button 
											class="copy-btn" 
											onclick={() => copyToClipboard(user.temp_password)}
											title="Copy password"
										>
											📋
										</button>
									</td>
									<td class="date-cell">{formatDate(user.temp_password_created_at)}</td>
									<td class="status-cell">
										{#if user.has_logged_in}
											<span class="badge badge-success">Has logged in</span>
										{:else}
											<span class="badge badge-warning">Pending login</span>
										{/if}
									</td>
									<td class="actions-cell">
										<button 
											class="btn btn-sm btn-danger"
											onclick={() => deactivateTempPassword(user)}
											disabled={deactivatingUserId === user.id}
											title="Deactivate temp password"
										>
											{deactivatingUserId === user.id ? '⏳' : '❌'}
										</button>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>

					<!-- Pagination -->
					{#if assignedTotalPages > 1}
						<div class="pagination">
							<button 
								class="btn btn-sm" 
								disabled={assignedCurrentPage === 1}
								onclick={() => assignedCurrentPage = assignedCurrentPage - 1}
							>
								← Previous
							</button>
							<span class="page-info">Page {assignedCurrentPage} of {assignedTotalPages}</span>
							<button 
								class="btn btn-sm" 
								disabled={assignedCurrentPage === assignedTotalPages}
								onclick={() => assignedCurrentPage = assignedCurrentPage + 1}
							>
								Next →
							</button>
						</div>
					{/if}
				{/if}
			</div>
		</div>
	{/if}
</div>

<!-- Results Modal -->
{#if showResultsModal}
	<div 
		class="modal-overlay" 
		onclick={closeResultsModal}
		onkeydown={(e) => e.key === 'Escape' && closeResultsModal()}
		role="dialog"
		aria-modal="true"
		aria-labelledby="results-modal-title"
		tabindex="-1"
	>
		<div class="modal-content" role="document" onclick={(e) => e.stopPropagation()}>
			<div class="modal-header">
				<h3 id="results-modal-title">Temp Password Assignment Results</h3>
				<button type="button" class="close-btn" onclick={closeResultsModal}>×</button>
			</div>
			<div class="modal-body">
				{#if tempPasswordResults.length > 0}
					<div class="results-summary">
						<span class="success-count">
							{tempPasswordResults.filter(r => !r.error).length} assigned
						</span>
						<span class="error-count">
							{tempPasswordResults.filter(r => r.error).length} failed
						</span>
					</div>
					<table class="results-table">
						<thead>
							<tr>
								<th>User ID</th>
								<th>Email</th>
								<th>Temp Password</th>
								<th>Email Sent</th>
								<th>Status</th>
							</tr>
						</thead>
						<tbody>
							{#each tempPasswordResults as result}
								<tr class:error={!!result.error}>
									<td>{result.user_id}</td>
									<td>{result.email}</td>
									<td>{result.temp_password || '-'}</td>
									<td>{result.email_sent ? 'Yes' : 'No'}</td>
									<td class:error-text={!!result.error}>
										{result.error || 'Success'}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				{:else}
					<p>No results to display.</p>
				{/if}
			</div>
			<div class="modal-footer">
				<button type="button" class="btn btn-primary" onclick={closeResultsModal}>
					Close
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.temp-passes-dashboard {
		padding: 1.5rem;
	}

	.dashboard-header {
		margin-bottom: 1.5rem;
	}

	.header-content h2 {
		margin: 0 0 0.5rem 0;
		font-size: 1.5rem;
	}

	.header-content p {
		margin: 0;
		color: #6b7280;
		font-size: 0.875rem;
	}

	/* Tabs */
	.tabs {
		display: flex;
		gap: 0;
		border-bottom: 2px solid #e5e7eb;
		margin-bottom: 1.5rem;
	}

	.tab-btn {
		padding: 0.75rem 1.5rem;
		background: none;
		border: none;
		border-bottom: 2px solid transparent;
		margin-bottom: -2px;
		font-size: 0.875rem;
		font-weight: 500;
		color: #6b7280;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		transition: all 0.2s;
	}

	.tab-btn:hover {
		color: #374151;
	}

	.tab-btn.active {
		color: #3b82f6;
		border-bottom-color: #3b82f6;
	}

	.tab-badge {
		background: #e5e7eb;
		color: #374151;
		padding: 0.125rem 0.5rem;
		border-radius: 9999px;
		font-size: 0.75rem;
	}

	.tab-btn.active .tab-badge {
		background: #dbeafe;
		color: #1e40af;
	}

	.tab-content {
		animation: fadeIn 0.2s ease;
	}

	@keyframes fadeIn {
		from { opacity: 0; }
		to { opacity: 1; }
	}

	/* Stats */
	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: 1rem;
		margin-bottom: 1.5rem;
	}

	.stat-card {
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		padding: 1rem;
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.stat-card.highlight-info {
		border-color: #3b82f6;
		background: #eff6ff;
	}

	.stat-card.highlight-warning {
		border-color: #f59e0b;
		background: #fffbeb;
	}

	.stat-card.highlight-warning {
		border-color: #f59e0b;
		background: #fffbeb;
	}

	.stat-icon {
		font-size: 1.5rem;
	}

	.stat-value {
		font-size: 1.5rem;
		font-weight: 700;
		color: #111827;
	}

	.stat-label {
		font-size: 0.75rem;
		color: #6b7280;
		text-transform: uppercase;
	}

	/* Criteria Info */
	.criteria-info {
		background: #f0f9ff;
		border: 1px solid #bae6fd;
		border-radius: 0.5rem;
		padding: 0.75rem 1rem;
		font-size: 0.875rem;
		color: #0369a1;
		margin-bottom: 1rem;
	}

	/* Filters */
	.filters-section {
		background: #f9fafb;
		border-radius: 0.5rem;
		padding: 1rem;
		margin-bottom: 1rem;
	}

	.filter-row {
		display: flex;
		gap: 1rem;
		flex-wrap: wrap;
		align-items: flex-end;
	}

	.filter-group {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.filter-group label {
		font-size: 0.75rem;
		font-weight: 500;
		color: #6b7280;
	}

	.filter-input {
		padding: 0.5rem;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		min-width: 250px;
	}

	.search-group {
		flex: 1;
		min-width: 200px;
	}

	.bulk-actions {
		display: flex;
		gap: 0.5rem;
		flex-wrap: wrap;
	}

	.filter-summary {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-top: 0.75rem;
		font-size: 0.875rem;
		color: #6b7280;
	}

	/* Table */
	.table-container {
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		overflow: hidden;
	}

	.loading-container,
	.empty-state {
		padding: 3rem;
		text-align: center;
		color: #6b7280;
	}

	.empty-state .hint {
		font-size: 0.875rem;
		margin-top: 0.5rem;
	}

	.users-table {
		width: 100%;
		border-collapse: collapse;
	}

	.users-table th,
	.users-table td {
		padding: 0.75rem 1rem;
		text-align: left;
		border-bottom: 1px solid #e5e7eb;
	}

	.users-table th {
		background: #f9fafb;
		font-weight: 600;
		font-size: 0.75rem;
		text-transform: uppercase;
		color: #6b7280;
	}

	.users-table tr.selected {
		background: #eff6ff;
	}

	.users-table tr.logged-in {
		background: #f0fdf4;
	}

	.checkbox-col {
		width: 40px;
		text-align: center;
	}

	.user-cell {
		min-width: 200px;
	}

	.user-email {
		font-weight: 500;
	}

	.user-name {
		font-size: 0.75rem;
		color: #6b7280;
	}

	.id-cell {
		color: #6b7280;
		font-family: monospace;
	}

	.temp-pass-cell {
		white-space: nowrap;
	}

	.date-cell {
		color: #6b7280;
		white-space: nowrap;
		font-size: 0.875rem;
	}

	/* Badges */
	.badge {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		padding: 0.25rem 0.5rem;
		border-radius: 9999px;
		font-size: 0.75rem;
		font-weight: 500;
	}

	.badge-info {
		background: #dbeafe;
		color: #1e40af;
		font-family: monospace;
	}

	.badge-success {
		background: #dcfce7;
		color: #166534;
	}

	.badge-warning {
		background: #fef3c7;
		color: #92400e;
	}

	.copy-btn {
		background: none;
		border: none;
		cursor: pointer;
		padding: 0.125rem;
		opacity: 0.6;
		transition: opacity 0.2s;
	}

	.copy-btn:hover {
		opacity: 1;
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

	.page-info {
		font-size: 0.875rem;
		color: #6b7280;
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
	}

	.btn-primary {
		background: #3b82f6;
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: #2563eb;
	}

	.btn-secondary {
		background: #f3f4f6;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-secondary:hover:not(:disabled) {
		background: #e5e7eb;
	}

	.btn-danger {
		background: #661e1e;
		color: white;
	}

	.btn-danger:hover:not(:disabled) {
		background: #dc2626;
	}

	.btn-outline {
		background: white;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-outline:hover:not(:disabled) {
		background: #f9fafb;
	}

	.btn-sm {
		padding: 0.375rem 0.75rem;
		font-size: 0.75rem;
	}

	.actions-cell {
		text-align: center;
	}

	.btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	/* Modal */
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
		background: white;
		border-radius: 0.5rem;
		box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
		max-width: 800px;
		width: 90%;
		max-height: 80vh;
		display: flex;
		flex-direction: column;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem 1.5rem;
		border-bottom: 1px solid #e5e7eb;
	}

	.modal-header h3 {
		margin: 0;
		font-size: 1.25rem;
		font-weight: 600;
	}

	.close-btn {
		background: none;
		border: none;
		font-size: 1.5rem;
		cursor: pointer;
		color: #6b7280;
		padding: 0.25rem;
		line-height: 1;
	}

	.close-btn:hover {
		color: #111827;
	}

	.modal-body {
		padding: 1.5rem;
		overflow-y: auto;
		flex: 1;
	}

	.modal-footer {
		padding: 1rem 1.5rem;
		border-top: 1px solid #e5e7eb;
		display: flex;
		justify-content: flex-end;
		gap: 0.5rem;
	}

	.results-summary {
		display: flex;
		gap: 1.5rem;
		margin-bottom: 1rem;
		font-weight: 600;
	}

	.success-count {
		color: #16a34a;
	}

	.error-count {
		color: #dc2626;
	}

	.results-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
	}

	.results-table th,
	.results-table td {
		padding: 0.75rem;
		text-align: left;
		border-bottom: 1px solid #e5e7eb;
	}

	.results-table th {
		background: #f9fafb;
		font-weight: 600;
	}

	.results-table tr.error {
		background: #fef2f2;
	}

	.error-text {
		color: #dc2626;
	}
</style>
