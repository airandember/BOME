<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	interface AuditLog {
		id: string;
		timestamp: string;
		user_id?: string;
		user_email?: string;
		user_name?: string;
		action: string;
		resource: string;
		resource_id?: string;
		ip_address: string;
		user_agent: string;
		status: 'success' | 'failed' | 'warning';
		details?: string;
		metadata?: Record<string, any>;
	}

	let auditLogs: AuditLog[] = [];
	let loading = true;
	let error = '';
	let searchQuery = '';
	let selectedAction = 'all';
	let selectedStatus = 'all';
	let selectedDateRange = '7d';
	let currentPage = 1;
	let totalPages = 1;

	const actionOptions = [
		{ value: 'all', label: 'All Actions' },
		{ value: 'login', label: 'Login' },
		{ value: 'logout', label: 'Logout' },
		{ value: 'create', label: 'Create' },
		{ value: 'update', label: 'Update' },
		{ value: 'delete', label: 'Delete' },
		{ value: 'view', label: 'View' },
		{ value: 'upload', label: 'Upload' },
		{ value: 'download', label: 'Download' },
		{ value: 'payment', label: 'Payment' },
		{ value: 'subscription', label: 'Subscription' }
	];

	const statusOptions = [
		{ value: 'all', label: 'All Status' },
		{ value: 'success', label: 'Success' },
		{ value: 'failed', label: 'Failed' },
		{ value: 'warning', label: 'Warning' }
	];

	const dateRangeOptions = [
		{ value: '1d', label: 'Last 24 Hours' },
		{ value: '7d', label: 'Last 7 Days' },
		{ value: '30d', label: 'Last 30 Days' },
		{ value: '90d', label: 'Last 90 Days' }
	];

	onMount(async () => {
		await loadAuditLogs();
	});

	const loadAuditLogs = async (page = 1) => {
		try {
			loading = true;
			error = '';
			
			// Mock data - replace with actual API call
			await new Promise(resolve => setTimeout(resolve, 1000));
			
			const mockLogs: AuditLog[] = [
				{
					id: 'log_001',
					timestamp: '2024-06-18T10:30:00Z',
					user_id: 'user_123',
					user_email: 'john.smith@example.com',
					user_name: 'John Smith',
					action: 'login',
					resource: 'authentication',
					ip_address: '192.168.1.100',
					user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
					status: 'success',
					details: 'Successful login from Chrome browser'
				},
				{
					id: 'log_002',
					timestamp: '2024-06-18T10:25:00Z',
					user_id: 'user_456',
					user_email: 'admin@bome.com',
					user_name: 'Admin User',
					action: 'update',
					resource: 'user',
					resource_id: 'user_789',
					ip_address: '192.168.1.50',
					user_agent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
					status: 'success',
					details: 'Updated user profile information',
					metadata: { fields_updated: ['email', 'firstName'] }
				},
				{
					id: 'log_003',
					timestamp: '2024-06-18T10:20:00Z',
					user_id: 'user_789',
					user_email: 'sarah.johnson@example.com',
					user_name: 'Sarah Johnson',
					action: 'upload',
					resource: 'video',
					resource_id: 'video_123',
					ip_address: '192.168.1.200',
					user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
					status: 'success',
					details: 'Uploaded new video: "Ancient Civilizations"',
					metadata: { file_size: '2.5GB', duration: '45:30' }
				},
				{
					id: 'log_004',
					timestamp: '2024-06-18T10:15:00Z',
					user_id: 'user_999',
					user_email: 'hacker@malicious.com',
					user_name: 'Unknown User',
					action: 'login',
					resource: 'authentication',
					ip_address: '203.0.113.1',
					user_agent: 'curl/7.68.0',
					status: 'failed',
					details: 'Failed login attempt - invalid credentials',
					metadata: { attempts: 5, blocked: true }
				},
				{
					id: 'log_005',
					timestamp: '2024-06-18T10:10:00Z',
					user_id: 'user_123',
					user_email: 'john.smith@example.com',
					user_name: 'John Smith',
					action: 'payment',
					resource: 'subscription',
					resource_id: 'sub_456',
					ip_address: '192.168.1.100',
					user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
					status: 'success',
					details: 'Payment processed for Premium subscription',
					metadata: { amount: 29.99, currency: 'USD', payment_method: 'card' }
				},
				{
					id: 'log_006',
					timestamp: '2024-06-18T10:05:00Z',
					action: 'system',
					resource: 'backup',
					ip_address: '127.0.0.1',
					user_agent: 'System/1.0',
					status: 'warning',
					details: 'Backup process completed with warnings',
					metadata: { backup_size: '15.2GB', warnings: 2 }
				}
			];

			// Apply filters
			let filteredLogs = mockLogs;
			
			if (selectedAction !== 'all') {
				filteredLogs = filteredLogs.filter(log => log.action === selectedAction);
			}
			
			if (selectedStatus !== 'all') {
				filteredLogs = filteredLogs.filter(log => log.status === selectedStatus);
			}
			
			if (searchQuery.trim()) {
				const query = searchQuery.toLowerCase();
				filteredLogs = filteredLogs.filter(log => 
					log.user_email?.toLowerCase().includes(query) ||
					log.user_name?.toLowerCase().includes(query) ||
					log.action.toLowerCase().includes(query) ||
					log.resource.toLowerCase().includes(query) ||
					log.details?.toLowerCase().includes(query) ||
					log.ip_address.includes(query)
				);
			}

			auditLogs = filteredLogs;
			totalPages = Math.ceil(filteredLogs.length / 20);
			currentPage = page;
		} catch (err) {
			error = 'Failed to load audit logs';
			console.error('Error loading audit logs:', err);
		} finally {
			loading = false;
		}
	};

	const handleFilterChange = async () => {
		currentPage = 1;
		await loadAuditLogs(1);
	};

	const handlePageChange = async (page: number) => {
		if (page >= 1 && page <= totalPages) {
			await loadAuditLogs(page);
		}
	};

	const formatDate = (dateString: string): string => {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
			second: '2-digit'
		});
	};

	const getStatusBadge = (status: string) => {
		const statusConfig = {
			success: { text: 'Success', class: 'status-success' },
			failed: { text: 'Failed', class: 'status-failed' },
			warning: { text: 'Warning', class: 'status-warning' }
		};

		const config = statusConfig[status as keyof typeof statusConfig] || { text: status, class: 'status-unknown' };
		
		return `<span class="status-badge ${config.class}">${config.text}</span>`;
	};

	const getActionIcon = (action: string) => {
		switch (action) {
			case 'login':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"></path>
					<polyline points="10,17 15,12 10,7"></polyline>
					<line x1="15" y1="12" x2="3" y2="12"></line>
				</svg>`;
			case 'logout':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
					<polyline points="16,17 21,12 16,7"></polyline>
					<line x1="21" y1="12" x2="9" y2="12"></line>
				</svg>`;
			case 'create':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<line x1="12" y1="5" x2="12" y2="19"></line>
					<line x1="5" y1="12" x2="19" y2="12"></line>
				</svg>`;
			case 'update':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
					<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
				</svg>`;
			case 'delete':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<polyline points="3,6 5,6 21,6"></polyline>
					<path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
				</svg>`;
			case 'upload':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
					<polyline points="17,8 12,3 7,8"></polyline>
					<line x1="12" y1="3" x2="12" y2="15"></line>
				</svg>`;
			case 'download':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
					<polyline points="7,10 12,15 17,10"></polyline>
					<line x1="12" y1="15" x2="12" y2="3"></line>
				</svg>`;
			case 'payment':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<rect x="1" y="4" width="22" height="16" rx="2" ry="2"></rect>
					<line x1="1" y1="10" x2="23" y2="10"></line>
				</svg>`;
			case 'system':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect>
					<line x1="8" y1="21" x2="16" y2="21"></line>
					<line x1="12" y1="17" x2="12" y2="21"></line>
				</svg>`;
			default:
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="12" cy="12" r="10"></circle>
					<line x1="12" y1="16" x2="12" y2="12"></line>
					<line x1="12" y1="8" x2="12.01" y2="8"></line>
				</svg>`;
		}
	};

	const exportLogs = async () => {
		try {
			// Mock export - replace with actual implementation
			showToast('Audit logs exported successfully', 'success');
		} catch (err) {
			showToast('Failed to export audit logs', 'error');
		}
	};

	const viewLogDetails = (log: AuditLog) => {
		// Show detailed view or navigate to detail page
		console.log('View log details:', log);
	};
</script>

<svelte:head>
	<title>Audit Logs - Admin Dashboard</title>
</svelte:head>

<div class="audit-logs-page">
	<div class="page-header">
		<div class="header-content">
			<div class="header-text">
				<h1>Audit Logs</h1>
				<p>Track user activities and system events</p>
			</div>
			
			<div class="header-actions">
				<button class="btn btn-outline" on:click={exportLogs}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
						<polyline points="7,10 12,15 17,10"></polyline>
						<line x1="12" y1="15" x2="12" y2="3"></line>
					</svg>
					Export Logs
				</button>
				<button class="btn btn-primary" on:click={() => goto('/admin/security')}>
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
					placeholder="Search by user, action, resource, or IP..."
					bind:value={searchQuery}
					on:input={handleFilterChange}
					class="filter-input"
				/>
			</div>

			<div class="filter-group">
				<label for="action">Action</label>
				<select id="action" bind:value={selectedAction} on:change={handleFilterChange} class="filter-select">
					{#each actionOptions as option}
						<option value={option.value}>{option.label}</option>
					{/each}
				</select>
			</div>

			<div class="filter-group">
				<label for="status">Status</label>
				<select id="status" bind:value={selectedStatus} on:change={handleFilterChange} class="filter-select">
					{#each statusOptions as option}
						<option value={option.value}>{option.label}</option>
					{/each}
				</select>
			</div>

			<div class="filter-group">
				<label for="dateRange">Date Range</label>
				<select id="dateRange" bind:value={selectedDateRange} on:change={handleFilterChange} class="filter-select">
					{#each dateRangeOptions as option}
						<option value={option.value}>{option.label}</option>
					{/each}
				</select>
			</div>
		</div>
	</div>

	{#if loading}
		<div class="loading-container">
			<LoadingSpinner size="large" color="primary" />
			<p>Loading audit logs...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<p class="error-message">{error}</p>
			<button class="btn btn-primary" on:click={() => loadAuditLogs()}>
				Try Again
			</button>
		</div>
	{:else}
		<!-- Audit Logs Table -->
		<div class="logs-table glass">
			<div class="table-header">
				<div class="header-cell">Timestamp</div>
				<div class="header-cell">User</div>
				<div class="header-cell">Action</div>
				<div class="header-cell">Resource</div>
				<div class="header-cell">IP Address</div>
				<div class="header-cell">Status</div>
				<div class="header-cell">Actions</div>
			</div>

			{#each auditLogs as log}
				<div class="table-row">
					<div class="table-cell">
						<span class="timestamp">{formatDate(log.timestamp)}</span>
					</div>
					<div class="table-cell">
						<div class="user-info">
							{#if log.user_name}
								<span class="user-name">{log.user_name}</span>
								<span class="user-email">{log.user_email}</span>
							{:else}
								<span class="system-user">System</span>
							{/if}
						</div>
					</div>
					<div class="table-cell">
						<div class="action-info">
							<div class="action-icon">
								{@html getActionIcon(log.action)}
							</div>
							<span class="action-text">{log.action}</span>
						</div>
					</div>
					<div class="table-cell">
						<div class="resource-info">
							<span class="resource-type">{log.resource}</span>
							{#if log.resource_id}
								<span class="resource-id">#{log.resource_id}</span>
							{/if}
						</div>
					</div>
					<div class="table-cell">
						<span class="ip-address">{log.ip_address}</span>
					</div>
					<div class="table-cell">
						{@html getStatusBadge(log.status)}
					</div>
					<div class="table-cell">
						<button class="btn btn-ghost btn-small" on:click={() => viewLogDetails(log)}>
							View Details
						</button>
					</div>
				</div>
			{/each}

			{#if auditLogs.length === 0}
				<div class="empty-state">
					<div class="empty-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
							<polyline points="14,2 14,8 20,8"></polyline>
							<line x1="16" y1="13" x2="8" y2="13"></line>
							<line x1="16" y1="17" x2="8" y2="17"></line>
						</svg>
					</div>
					<h3>No audit logs found</h3>
					<p>No logs match your current filters.</p>
				</div>
			{/if}
		</div>

		<!-- Pagination -->
		{#if totalPages > 1}
			<div class="pagination">
				<button 
					class="btn btn-outline"
					disabled={currentPage === 1}
					on:click={() => handlePageChange(currentPage - 1)}
				>
					Previous
				</button>
				
				<div class="page-numbers">
					{#each Array.from({ length: totalPages }, (_, i) => i + 1) as page}
						<button 
							class="page-number"
							class:active={page === currentPage}
							on:click={() => handlePageChange(page)}
						>
							{page}
						</button>
					{/each}
				</div>
				
				<button 
					class="btn btn-outline"
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
	.audit-logs-page {
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
		grid-template-columns: 2fr 1fr 1fr 1fr;
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

	.logs-table {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
		margin-bottom: var(--space-xl);
	}

	.table-header {
		display: grid;
		grid-template-columns: 1.5fr 1.5fr 1fr 1fr 1fr 0.8fr 0.8fr;
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
		grid-template-columns: 1.5fr 1.5fr 1fr 1fr 1fr 0.8fr 0.8fr;
		gap: var(--space-lg);
		padding: var(--space-md) 0;
		border-bottom: 1px solid var(--border-color);
		align-items: center;
	}

	.table-row:last-child {
		border-bottom: none;
	}

	.timestamp {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.user-info {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.user-name {
		font-weight: 600;
		color: var(--text-primary);
		font-size: var(--text-sm);
	}

	.user-email {
		font-size: var(--text-xs);
		color: var(--text-secondary);
	}

	.system-user {
		font-weight: 600;
		color: var(--warning);
		font-size: var(--text-sm);
	}

	.action-info {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
	}

	.action-icon {
		width: 20px;
		height: 20px;
		color: var(--text-secondary);
	}

	.action-icon svg {
		width: 100%;
		height: 100%;
	}

	.action-text {
		font-size: var(--text-sm);
		color: var(--text-primary);
		text-transform: capitalize;
	}

	.resource-info {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.resource-type {
		font-size: var(--text-sm);
		color: var(--text-primary);
		text-transform: capitalize;
	}

	.resource-id {
		font-size: var(--text-xs);
		color: var(--text-secondary);
	}

	.ip-address {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		font-family: monospace;
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

	.status-success {
		background: var(--success-bg);
		color: var(--success-text);
	}

	.status-failed {
		background: var(--error-bg);
		color: var(--error-text);
	}

	.status-warning {
		background: var(--warning-bg);
		color: var(--warning-text);
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
	}

	.pagination {
		display: flex;
		justify-content: center;
		align-items: center;
		gap: var(--space-lg);
		margin-top: var(--space-xl);
	}

	.page-numbers {
		display: flex;
		gap: var(--space-sm);
	}

	.page-number {
		padding: var(--space-sm) var(--space-md);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		background: var(--bg-glass);
		color: var(--text-primary);
		cursor: pointer;
		transition: all var(--transition-fast);
	}

	.page-number:hover {
		background: var(--bg-hover);
	}

	.page-number.active {
		background: var(--primary);
		color: var(--white);
		border-color: var(--primary);
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

	@media (max-width: 768px) {
		.audit-logs-page {
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
	}
</style> 