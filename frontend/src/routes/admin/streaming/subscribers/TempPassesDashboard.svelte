<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	// Types
	interface EligibleUser {
		id: number;
		email: string;
		first_name: string;
		last_name: string;
		has_temp_password: boolean;
		has_password: boolean;
		temp_password_created_at: string | null;
		linked_customers: number;
		created_at: string;
		selected?: boolean;
	}

	interface TempPasswordResult {
		user_id: number;
		email: string;
		temp_password: string;
		email_sent: boolean;
		error?: string;
	}

	// State
	let loading = $state(false);
	let assigning = $state(false);
	let eligibleUsers = $state<EligibleUser[]>([]);
	let filteredUsers = $state<EligibleUser[]>([]);
	let selectedUserIds = $state<Set<number>>(new Set());
	let tempPasswordResults = $state<TempPasswordResult[]>([]);
	let showResultsModal = $state(false);

	// Search & Filter state
	let searchTerm = $state('');
	let filterHasTempPassword: 'all' | 'yes' | 'no' = $state('all');
	let filterHasStripeCustomer: 'all' | 'yes' | 'no' = $state('all');
	let sortBy: 'email' | 'created_at' | 'id' = $state('email');
	let sortDir: 'asc' | 'desc' = $state('asc');

	// Pagination state
	let currentPage = $state(1);
	let itemsPerPage = $state(50);
	let totalPages = $derived(Math.ceil(filteredUsers.length / itemsPerPage));
	let paginatedUsers = $derived(() => {
		const start = (currentPage - 1) * itemsPerPage;
		return filteredUsers.slice(start, start + itemsPerPage);
	});

	// Stats
	let stats = $derived(() => {
		const total = eligibleUsers.length;
		const withTempPass = eligibleUsers.filter(u => u.has_temp_password).length;
		const withStripe = eligibleUsers.filter(u => u.linked_customers > 0).length;
		// Only count users who need action: have Stripe, no temp password, AND no existing password
		const needsAction = eligibleUsers.filter(u => !u.has_temp_password && u.linked_customers > 0 && !u.has_password).length;
		const withPassword = eligibleUsers.filter(u => u.has_password).length;
		return { total, withTempPass, withStripe, needsAction, withPassword };
	});

	// Selection helpers
	let allVisibleSelected = $derived(() => {
		const visible = paginatedUsers();
		return visible.length > 0 && visible.every(u => selectedUserIds.has(u.id));
	});

	let someVisibleSelected = $derived(() => {
		const visible = paginatedUsers();
		return visible.some(u => selectedUserIds.has(u.id)) && !allVisibleSelected();
	});

	// Load eligible users
	async function loadEligibleUsers() {
		loading = true;
		try {
			const response = await apiRequest('/admin/users/never-logged-in');
			if (response.ok) {
				const data = await response.json();
				eligibleUsers = (data.users || []).map((u: any) => ({
					...u,
					selected: false
				}));
				applyFilters();
				console.log(`📋 Loaded ${eligibleUsers.length} eligible users`);
			} else {
				throw new Error('Failed to load users');
			}
		} catch (error) {
			console.error('Error loading eligible users:', error);
			showToast('Failed to load eligible users', 'error');
			eligibleUsers = [];
			filteredUsers = [];
		} finally {
			loading = false;
		}
	}

	// Apply filters and sorting
	function applyFilters() {
		let result = [...eligibleUsers];

		// Search filter
		if (searchTerm.trim()) {
			const term = searchTerm.toLowerCase();
			result = result.filter(u => 
				u.email.toLowerCase().includes(term) ||
				(u.first_name || '').toLowerCase().includes(term) ||
				(u.last_name || '').toLowerCase().includes(term) ||
				u.id.toString().includes(term)
			);
		}

		// Has temp password filter
		if (filterHasTempPassword === 'yes') {
			result = result.filter(u => u.has_temp_password);
		} else if (filterHasTempPassword === 'no') {
			result = result.filter(u => !u.has_temp_password);
		}

		// Has Stripe customer filter
		if (filterHasStripeCustomer === 'yes') {
			result = result.filter(u => u.linked_customers > 0);
		} else if (filterHasStripeCustomer === 'no') {
			result = result.filter(u => u.linked_customers === 0);
		}

		// Sorting
		result.sort((a, b) => {
			let comparison = 0;
			switch (sortBy) {
				case 'email':
					comparison = a.email.localeCompare(b.email);
					break;
				case 'created_at':
					comparison = new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
					break;
				case 'id':
					comparison = a.id - b.id;
					break;
			}
			return sortDir === 'asc' ? comparison : -comparison;
		});

		filteredUsers = result;
		currentPage = 1; // Reset to first page when filters change
	}

	// Watch for filter changes
	$effect(() => {
		// Trigger on any filter change
		searchTerm;
		filterHasTempPassword;
		filterHasStripeCustomer;
		sortBy;
		sortDir;
		applyFilters();
	});

	// Selection handlers
	function toggleSelectAll() {
		const visible = paginatedUsers();
		if (allVisibleSelected()) {
			// Deselect all visible
			visible.forEach(u => selectedUserIds.delete(u.id));
		} else {
			// Select all visible
			visible.forEach(u => selectedUserIds.add(u.id));
		}
		selectedUserIds = new Set(selectedUserIds);
	}

	function toggleUserSelection(userId: number) {
		if (selectedUserIds.has(userId)) {
			selectedUserIds.delete(userId);
		} else {
			selectedUserIds.add(userId);
		}
		selectedUserIds = new Set(selectedUserIds);
	}

	function selectAllFiltered() {
		filteredUsers.forEach(u => selectedUserIds.add(u.id));
		selectedUserIds = new Set(selectedUserIds);
		showToast(`Selected ${filteredUsers.length} users`, 'success');
	}

	function clearSelection() {
		selectedUserIds = new Set();
	}

	// Quick selection helpers
	function selectNeedingAction() {
		clearSelection();
		eligibleUsers
			.filter(u => !u.has_temp_password && u.linked_customers > 0 && !u.has_password)
			.forEach(u => selectedUserIds.add(u.id));
		selectedUserIds = new Set(selectedUserIds);
		showToast(`Selected ${selectedUserIds.size} users needing action`, 'success');
	}

	function selectWithoutTempPass() {
		clearSelection();
		eligibleUsers
			.filter(u => !u.has_temp_password && !u.has_password)
			.forEach(u => selectedUserIds.add(u.id));
		selectedUserIds = new Set(selectedUserIds);
		showToast(`Selected ${selectedUserIds.size} users without temp passwords`, 'success');
	}

	// Bulk assign temp passwords
	async function handleBulkAssign(sendEmail: boolean) {
		if (selectedUserIds.size === 0) {
			showToast('Please select users to assign temp passwords', 'warning');
			return;
		}

		assigning = true;
		try {
			const response = await apiRequest('/admin/users/bulk-temp-password', {
				method: 'POST',
				body: JSON.stringify({
					user_ids: Array.from(selectedUserIds),
					send_email: sendEmail
				})
			});

			if (response.ok) {
				const data = await response.json();
				tempPasswordResults = data.results || [];
				showResultsModal = true;

				if (data.assigned_count > 0) {
					showToast(`Assigned temp passwords to ${data.assigned_count} users`, 'success');
				}
				if (data.error_count > 0) {
					showToast(`${data.error_count} users could not be assigned`, 'warning');
				}

				// Refresh the list
				clearSelection();
				await loadEligibleUsers();
			} else {
				const error = await response.json();
				showToast(error.error || 'Failed to assign temp passwords', 'error');
			}
		} catch (error) {
			console.error('Error assigning temp passwords:', error);
			showToast('Failed to assign temp passwords', 'error');
		} finally {
			assigning = false;
		}
	}

	// Single user assign
	async function handleSingleAssign(user: EligibleUser, sendEmail: boolean) {
		try {
			const response = await apiRequest('/admin/users/bulk-temp-password', {
				method: 'POST',
				body: JSON.stringify({
					user_ids: [user.id],
					send_email: sendEmail
				})
			});

			if (response.ok) {
				const data = await response.json();
				if (data.assigned_count > 0) {
					showToast(`Temp password assigned to ${user.email}`, 'success');
					await loadEligibleUsers();
				} else if (data.error_count > 0 && data.results?.[0]?.error) {
					// Show specific error from the backend
					showToast(data.results[0].error, 'error');
					await loadEligibleUsers(); // Refresh to update UI state
				} else {
					showToast('Failed to assign temp password', 'error');
				}
			}
		} catch (error) {
			console.error('Error:', error);
			showToast('Failed to assign temp password', 'error');
		}
	}

	// Format date helper
	function formatDate(dateStr: string | null): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	// Copy to clipboard
	async function copyToClipboard(text: string) {
		try {
			await navigator.clipboard.writeText(text);
			showToast('Copied to clipboard', 'success');
		} catch {
			showToast('Failed to copy', 'error');
		}
	}

	// Export results
	function exportResults() {
		if (tempPasswordResults.length === 0) return;

		const csv = [
			['User ID', 'Email', 'Temp Password', 'Email Sent', 'Status'].join(','),
			...tempPasswordResults.map(r => [
				r.user_id,
				r.email,
				r.temp_password || '',
				r.email_sent ? 'Yes' : 'No',
				r.error || 'Success'
			].join(','))
		].join('\n');

		const blob = new Blob([csv], { type: 'text/csv' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `temp_passwords_${new Date().toISOString().split('T')[0]}.csv`;
		a.click();
		URL.revokeObjectURL(url);
	}

	onMount(() => {
		loadEligibleUsers();
	});
</script>

<div class="temp-passes-dashboard">
	<!-- Header -->
	<div class="dashboard-header">
		<div class="header-content">
			<h2>🔑 Temp Password Management</h2>
			<p>Assign temporary passwords (BOME_[user_id]) to users who have never logged in.</p>
		</div>
		<div class="header-actions">
			<button class="btn btn-secondary" onclick={loadEligibleUsers} disabled={loading}>
				{loading ? '⏳' : '🔄'} Refresh
			</button>
		</div>
	</div>

	<!-- Stats Cards -->
	<div class="stats-grid">
		<div class="stat-card">
			<div class="stat-icon">👥</div>
			<div class="stat-content">
				<div class="stat-value">{stats().total}</div>
				<div class="stat-label">Total Eligible Users</div>
			</div>
		</div>
		<div class="stat-card highlight-warning">
			<div class="stat-icon">⚠️</div>
			<div class="stat-content">
				<div class="stat-value">{stats().needsAction}</div>
				<div class="stat-label">Need Action (Stripe + No Pass)</div>
			</div>
		</div>
		<div class="stat-card">
			<div class="stat-icon">🔑</div>
			<div class="stat-content">
				<div class="stat-value">{stats().withTempPass}</div>
				<div class="stat-label">Have Temp Password</div>
			</div>
		</div>
		<div class="stat-card">
			<div class="stat-icon">💳</div>
			<div class="stat-content">
				<div class="stat-value">{stats().withStripe}</div>
				<div class="stat-label">Linked to Stripe</div>
			</div>
		</div>
	</div>

	<!-- Quick Actions -->
	<div class="quick-actions-section">
		<h3>Quick Select</h3>
		<div class="quick-actions">
			<button class="btn btn-warning btn-sm" onclick={selectNeedingAction}>
				⚠️ Select All Needing Action ({stats().needsAction})
			</button>
			<button class="btn btn-secondary btn-sm" onclick={selectWithoutTempPass}>
				🔑 Select All Without Temp Pass
			</button>
			<button class="btn btn-secondary btn-sm" onclick={selectAllFiltered}>
				✅ Select All Filtered ({filteredUsers.length})
			</button>
			{#if selectedUserIds.size > 0}
				<button class="btn btn-ghost btn-sm" onclick={clearSelection}>
					❌ Clear Selection ({selectedUserIds.size})
				</button>
			{/if}
		</div>
	</div>

	<!-- Filters -->
	<div class="filters-section">
		<div class="filter-row">
			<div class="filter-group search-group">
				<label for="search">Search</label>
				<input
					id="search"
					type="text"
					bind:value={searchTerm}
					placeholder="Search by email, name, or ID..."
					class="filter-input"
				/>
			</div>

			<div class="filter-group">
				<label for="hasTempPass">Temp Password</label>
				<select id="hasTempPass" bind:value={filterHasTempPassword} class="filter-select">
					<option value="all">All</option>
					<option value="yes">Has Temp Pass</option>
					<option value="no">No Temp Pass</option>
				</select>
			</div>

			<div class="filter-group">
				<label for="hasStripe">Stripe Customer</label>
				<select id="hasStripe" bind:value={filterHasStripeCustomer} class="filter-select">
					<option value="all">All</option>
					<option value="yes">Has Stripe</option>
					<option value="no">No Stripe</option>
				</select>
			</div>

			<div class="filter-group">
				<label for="sortBy">Sort By</label>
				<select id="sortBy" bind:value={sortBy} class="filter-select">
					<option value="email">Email</option>
					<option value="created_at">Created Date</option>
					<option value="id">User ID</option>
				</select>
			</div>

			<div class="filter-group">
				<label for="sortDir">Order</label>
				<select id="sortDir" bind:value={sortDir} class="filter-select">
					<option value="asc">Ascending</option>
					<option value="desc">Descending</option>
				</select>
			</div>
		</div>

		<div class="filter-summary">
			<span>Showing {filteredUsers.length} of {eligibleUsers.length} users</span>
			{#if selectedUserIds.size > 0}
				<span class="selection-badge">
					{selectedUserIds.size} selected
				</span>
			{/if}
		</div>
	</div>

	<!-- Bulk Actions Bar -->
	{#if selectedUserIds.size > 0}
		<div class="bulk-actions-bar">
			<span class="bulk-count">
				<strong>{selectedUserIds.size}</strong> user{selectedUserIds.size !== 1 ? 's' : ''} selected
			</span>
			<div class="bulk-buttons">
				<button 
					class="btn btn-primary"
					onclick={() => handleBulkAssign(true)}
					disabled={assigning}
				>
					{assigning ? '⏳ Assigning...' : '📧 Assign & Send Email'}
				</button>
				<button 
					class="btn btn-outline"
					onclick={() => handleBulkAssign(false)}
					disabled={assigning}
				>
					{assigning ? '⏳ Assigning...' : '🔑 Assign Only (No Email)'}
				</button>
			</div>
		</div>
	{/if}

	<!-- Users Table -->
	<div class="table-container">
		{#if loading}
			<div class="loading-container">
				<LoadingSpinner />
				<p>Loading eligible users...</p>
			</div>
		{:else if filteredUsers.length === 0}
			<div class="empty-state">
				<p>No users found matching your criteria.</p>
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
								title="Select all visible"
							/>
						</th>
						<th>User</th>
						<th>ID</th>
						<th>Stripe</th>
						<th>Has Password</th>
						<th>Temp Password</th>
						<th>Created</th>
						<th>Actions</th>
					</tr>
				</thead>
				<tbody>
					{#each paginatedUsers() as user}
						<tr class:selected={selectedUserIds.has(user.id)} class:has-temp={user.has_temp_password}>
							<td class="checkbox-col">
								<input
									type="checkbox"
									checked={selectedUserIds.has(user.id)}
									onchange={() => toggleUserSelection(user.id)}
								/>
							</td>
							<td class="user-cell">
								<div class="user-email">{user.email}</div>
								{#if user.first_name || user.last_name}
									<div class="user-name">{user.first_name} {user.last_name}</div>
								{/if}
							</td>
							<td class="id-cell">{user.id}</td>
							<td class="stripe-cell">
								{#if user.linked_customers > 0}
									<span class="badge badge-success">✅ {user.linked_customers}</span>
								{:else}
									<span class="badge badge-muted">❌</span>
								{/if}
							</td>
							<td class="password-cell">
								{#if user.has_password}
									<span class="badge badge-error">🔒 Yes</span>
								{:else}
									<span class="badge badge-muted">No</span>
								{/if}
							</td>
							<td class="temp-pass-cell">
								{#if user.has_temp_password}
									<span class="badge badge-info">
										🔑 BOME_{user.id}
									</span>
									<button 
										class="copy-btn" 
										onclick={() => copyToClipboard(`BOME_${user.id}`)}
										title="Copy password"
									>
										📋
									</button>
								{:else}
									<span class="badge badge-warning">Not set</span>
								{/if}
							</td>
							<td class="date-cell">{formatDate(user.created_at)}</td>
							<td class="actions-cell">
								{#if user.has_password}
									<span class="text-muted" title="User already has a password set">🔒 Has password</span>
								{:else if !user.has_temp_password}
									<button 
										class="btn btn-sm btn-primary"
										onclick={() => handleSingleAssign(user, true)}
										title="Assign and send email"
									>
										📧
									</button>
									<button 
										class="btn btn-sm btn-outline"
										onclick={() => handleSingleAssign(user, false)}
										title="Assign only"
									>
										🔑
									</button>
								{:else}
									<span class="text-muted">Assigned</span>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>

			<!-- Pagination -->
			{#if totalPages > 1}
				<div class="pagination">
					<div class="pagination-info">
						Page {currentPage} of {totalPages}
					</div>
					<div class="pagination-controls">
						<button 
							class="page-btn"
							disabled={currentPage <= 1}
							onclick={() => currentPage = 1}
						>
							First
						</button>
						<button 
							class="page-btn"
							disabled={currentPage <= 1}
							onclick={() => currentPage--}
						>
							← Prev
						</button>
						<span class="page-current">{currentPage}</span>
						<button 
							class="page-btn"
							disabled={currentPage >= totalPages}
							onclick={() => currentPage++}
						>
							Next →
						</button>
						<button 
							class="page-btn"
							disabled={currentPage >= totalPages}
							onclick={() => currentPage = totalPages}
						>
							Last
						</button>
					</div>
					<div class="items-per-page">
						<label for="perPage">Per page:</label>
						<select id="perPage" bind:value={itemsPerPage}>
							<option value={25}>25</option>
							<option value={50}>50</option>
							<option value={100}>100</option>
							<option value={200}>200</option>
						</select>
					</div>
				</div>
			{/if}
		{/if}
	</div>
</div>

<!-- Results Modal -->
{#if showResultsModal}
	<div 
		class="modal-overlay" 
		onclick={() => showResultsModal = false}
		onkeydown={(e) => e.key === 'Escape' && (showResultsModal = false)}
		role="dialog"
		aria-modal="true"
		aria-labelledby="modal-title"
		tabindex="-1"
	>
		<div class="modal-content" onclick={(e) => e.stopPropagation()} role="document">
			<div class="modal-header">
				<h3 id="modal-title">🔑 Assignment Results</h3>
				<button class="close-btn" onclick={() => showResultsModal = false}>✕</button>
			</div>
			<div class="modal-body">
				{#if tempPasswordResults.length > 0}
					<div class="results-summary">
						<span class="success-count">
							✅ {tempPasswordResults.filter(r => !r.error).length} successful
						</span>
						{#if tempPasswordResults.some(r => r.error)}
							<span class="error-count">
								❌ {tempPasswordResults.filter(r => r.error).length} failed
							</span>
						{/if}
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
								<tr class:error={result.error}>
									<td>{result.user_id}</td>
									<td>{result.email || '-'}</td>
									<td class="password-cell">
										{#if result.temp_password}
											<code>{result.temp_password}</code>
											<button 
												class="copy-btn" 
												onclick={() => copyToClipboard(result.temp_password)}
											>
												📋
											</button>
										{:else}
											-
										{/if}
									</td>
									<td>{result.email_sent ? '✅' : '❌'}</td>
									<td class:error-text={result.error}>
										{result.error || '✅ Success'}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				{:else}
					<p>No results to display</p>
				{/if}
			</div>
			<div class="modal-footer">
				{#if tempPasswordResults.length > 0}
					<button class="btn btn-secondary" onclick={exportResults}>
						📥 Export CSV
					</button>
				{/if}
				<button class="btn btn-primary" onclick={() => showResultsModal = false}>
					Close
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.temp-passes-dashboard {
		padding: 0;
	}

	/* Header */
	.dashboard-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 1.5rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid #e5e7eb;
	}

	.dashboard-header h2 {
		margin: 0 0 0.25rem 0;
		font-size: 1.5rem;
		color: #111827;
	}

	.dashboard-header p {
		margin: 0;
		color: #6b7280;
		font-size: 0.875rem;
	}

	.header-actions {
		display: flex;
		gap: 0.5rem;
	}

	/* Stats Grid */
	.stats-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
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

	.stat-card.highlight-warning {
		background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
		border-color: #f59e0b;
	}

	.stat-icon {
		font-size: 1.5rem;
	}

	.stat-content {
		flex: 1;
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
		letter-spacing: 0.5px;
	}

	/* Quick Actions */
	.quick-actions-section {
		background: #f9fafb;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		padding: 1rem;
		margin-bottom: 1.5rem;
	}

	.quick-actions-section h3 {
		margin: 0 0 0.75rem 0;
		font-size: 0.875rem;
		color: #374151;
		font-weight: 600;
	}

	.quick-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	/* Filters */
	.filters-section {
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		padding: 1rem;
		margin-bottom: 1rem;
	}

	.filter-row {
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
		margin-bottom: 0.75rem;
	}

	.filter-group {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.filter-group.search-group {
		flex: 1;
		min-width: 200px;
	}

	.filter-group label {
		font-size: 0.75rem;
		font-weight: 600;
		color: #374151;
	}

	.filter-input, .filter-select {
		padding: 0.5rem;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		font-size: 0.875rem;
	}

	.filter-input:focus, .filter-select:focus {
		outline: none;
		border-color: #2563eb;
		box-shadow: 0 0 0 2px rgba(37, 99, 235, 0.1);
	}

	.filter-summary {
		display: flex;
		justify-content: space-between;
		align-items: center;
		font-size: 0.875rem;
		color: #6b7280;
	}

	.selection-badge {
		background: #2563eb;
		color: white;
		padding: 0.25rem 0.75rem;
		border-radius: 9999px;
		font-size: 0.75rem;
		font-weight: 500;
	}

	/* Bulk Actions Bar */
	.bulk-actions-bar {
		display: flex;
		justify-content: space-between;
		align-items: center;
		background: #dbeafe;
		border: 1px solid #93c5fd;
		border-radius: 0.5rem;
		padding: 1rem;
		margin-bottom: 1rem;
	}

	.bulk-count {
		color: #1e40af;
	}

	.bulk-buttons {
		display: flex;
		gap: 0.5rem;
	}

	/* Table */
	.table-container {
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		overflow: hidden;
	}

	.users-table {
		width: 100%;
		border-collapse: collapse;
	}

	.users-table th {
		background: #f9fafb;
		padding: 0.75rem;
		text-align: left;
		font-size: 0.75rem;
		font-weight: 600;
		color: #374151;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		border-bottom: 1px solid #e5e7eb;
	}

	.users-table td {
		padding: 0.75rem;
		border-bottom: 1px solid #e5e7eb;
		font-size: 0.875rem;
	}

	.users-table tr:hover {
		background: #f9fafb;
	}

	.users-table tr.selected {
		background: #eff6ff;
	}

	.users-table tr.has-temp {
		background: #f0fdf4;
	}

	.users-table tr.has-temp.selected {
		background: #dcfce7;
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
		color: #111827;
	}

	.user-name {
		font-size: 0.75rem;
		color: #6b7280;
	}

	.id-cell {
		font-family: monospace;
		color: #6b7280;
	}

	.date-cell {
		color: #6b7280;
		white-space: nowrap;
	}

	.actions-cell {
		display: flex;
		gap: 0.25rem;
	}

	.temp-pass-cell {
		white-space: nowrap;
	}

	.temp-pass-cell .badge {
		display: inline-flex;
		vertical-align: middle;
	}

	.temp-pass-cell .copy-btn {
		vertical-align: middle;
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

	.badge-success {
		background: #dcfce7;
		color: #166534;
	}

	.badge-warning {
		background: #fef3c7;
		color: #92400e;
	}

	.badge-info {
		background: #dbeafe;
		color: #1e40af;
		font-family: monospace;
	}

	.badge-muted {
		background: #f3f4f6;
		color: #6b7280;
	}

	.badge-error {
		background: #fef2f2;
		color: #dc2626;
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
		justify-content: space-between;
		align-items: center;
		padding: 1rem;
		border-top: 1px solid #e5e7eb;
		background: #f9fafb;
	}

	.pagination-info {
		font-size: 0.875rem;
		color: #6b7280;
	}

	.pagination-controls {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.page-btn {
		padding: 0.375rem 0.75rem;
		border: 1px solid #d1d5db;
		background: white;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		cursor: pointer;
		transition: all 0.2s;
	}

	.page-btn:hover:not(:disabled) {
		background: #f3f4f6;
	}

	.page-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.page-current {
		padding: 0.375rem 0.75rem;
		background: #2563eb;
		color: white;
		border-radius: 0.375rem;
		font-weight: 500;
	}

	.items-per-page {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.875rem;
		color: #6b7280;
	}

	.items-per-page select {
		padding: 0.25rem 0.5rem;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
	}

	/* Loading & Empty States */
	.loading-container, .empty-state {
		padding: 3rem;
		text-align: center;
		color: #6b7280;
	}

	/* Modal */
	.modal-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}

	.modal-content {
		background: white;
		border-radius: 0.5rem;
		width: 90%;
		max-width: 800px;
		max-height: 80vh;
		display: flex;
		flex-direction: column;
		box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
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
	}

	.close-btn {
		background: none;
		border: none;
		font-size: 1.5rem;
		cursor: pointer;
		color: #6b7280;
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
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
		padding: 1rem 1.5rem;
		border-top: 1px solid #e5e7eb;
	}

	.results-summary {
		display: flex;
		gap: 1rem;
		margin-bottom: 1rem;
		font-weight: 500;
	}

	.success-count {
		color: #059669;
	}

	.error-count {
		color: #dc2626;
	}

	.results-table {
		width: 100%;
		border-collapse: collapse;
	}

	.results-table th, .results-table td {
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

	.password-cell {
		white-space: nowrap;
	}

	.error-text {
		color: #dc2626;
	}

	.text-muted {
		color: #9ca3af;
		font-style: italic;
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

	.btn-sm {
		padding: 0.25rem 0.5rem;
		font-size: 0.75rem;
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

	.btn-outline {
		background: transparent;
		color: #2563eb;
		border: 1px solid #2563eb;
	}

	.btn-outline:hover {
		background: #2563eb;
		color: white;
	}

	.btn-warning {
		background: #f59e0b;
		color: white;
	}

	.btn-warning:hover {
		background: #d97706;
	}

	.btn-ghost {
		background: transparent;
		color: #6b7280;
	}

	.btn-ghost:hover {
		background: #f3f4f6;
		color: #374151;
	}

	.btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	/* Responsive */
	@media (max-width: 768px) {
		.stats-grid {
			grid-template-columns: repeat(2, 1fr);
		}

		.filter-row {
			flex-direction: column;
		}

		.bulk-actions-bar {
			flex-direction: column;
			gap: 1rem;
		}

		.bulk-buttons {
			flex-direction: column;
			width: 100%;
		}

		.bulk-buttons .btn {
			width: 100%;
			justify-content: center;
		}
	}
</style>
