<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, apiRequest } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	interface Advertiser {
		id: number;
		user_id: number;
		company_name: string;
		contact_email: string;
		contact_phone: string;
		website: string;
		status: 'pending' | 'approved' | 'rejected' | 'suspended';
		approval_notes: string;
		created_at: string;
		updated_at: string;
		user?: {
			first_name: string;
			last_name: string;
			email: string;
			role: string;
		};
		stats?: {
			total_campaigns: number;
			active_campaigns: number;
			total_spent: number;
			total_impressions: number;
			total_clicks: number;
		};
	}

	let advertisers: Advertiser[] = [];
	let loading = true;
	let error = '';
	let searchTerm = '';
	let statusFilter = 'all';
	let currentPage = 1;
	let totalPages = 1;
	let totalAdvertisers = 0;

	// Action states
	let approvingId: number | null = null;
	let rejectingId: number | null = null;
	let suspendingId: number | null = null;

	onMount(async () => {
		if (!$auth.isAuthenticated) {
			goto('/admin');
			return;
		}

		const user = $auth.user;
		if (!user || !isAdminUser(user)) {
			showToast('Access denied. Admin privileges required.', 'error');
			goto('/admin');
			return;
		}

		await loadAdvertisers();
	});

	function isAdminUser(user: any): boolean {
		if (!user) return false;
		const adminRoles = [
			'super_admin', 'system_admin', 'content_manager', 
			'articles_manager', 'youtube_manager', 'streaming_manager',
			'events_manager', 'advertisement_manager', 'user_manager',
			'analytics_manager', 'financial_admin', 'admin'
		];
		return adminRoles.includes(user.role);
	}

	async function loadAdvertisers() {
		try {
			loading = true;
			error = '';

			// For now, use mock data since the backend endpoint is not fully implemented
			// In production, this would call the actual API
			await new Promise(resolve => setTimeout(resolve, 500)); // Simulate API delay

			// Mock advertiser data
			advertisers = [
				{
					id: 1,
					user_id: 101,
					company_name: 'TechCorp Solutions',
					contact_email: 'ads@techcorp.com',
					contact_phone: '+1-555-0123',
					website: 'https://techcorp.com',
					status: 'approved',
					approval_notes: 'Approved after review of business documentation',
					created_at: '2024-01-15T10:30:00Z',
					updated_at: '2024-01-20T14:22:00Z',
					user: {
						first_name: 'John',
						last_name: 'Smith',
						email: 'john.smith@techcorp.com',
						role: 'advertiser'
					},
					stats: {
						total_campaigns: 5,
						active_campaigns: 3,
						total_spent: 2500.00,
						total_impressions: 150000,
						total_clicks: 2500
					}
				},
				{
					id: 2,
					user_id: 102,
					company_name: 'Global Marketing Inc',
					contact_email: 'marketing@globalmarketing.com',
					contact_phone: '+1-555-0456',
					website: 'https://globalmarketing.com',
					status: 'pending',
					approval_notes: 'Awaiting business verification',
					created_at: '2024-01-18T09:15:00Z',
					updated_at: '2024-01-19T16:20:00Z',
					user: {
						first_name: 'Sarah',
						last_name: 'Johnson',
						email: 'sarah.johnson@globalmarketing.com',
						role: 'advertiser'
					},
					stats: {
						total_campaigns: 0,
						active_campaigns: 0,
						total_spent: 0,
						total_impressions: 0,
						total_clicks: 0
					}
				},
				{
					id: 3,
					user_id: 103,
					company_name: 'Digital Ads Pro',
					contact_email: 'contact@digitaladspro.com',
					contact_phone: '+1-555-0789',
					website: 'https://digitaladspro.com',
					status: 'approved',
					approval_notes: 'Verified business with good track record',
					created_at: '2024-01-10T14:45:00Z',
					updated_at: '2024-01-22T11:30:00Z',
					user: {
						first_name: 'Michael',
						last_name: 'Chen',
						email: 'michael.chen@digitaladspro.com',
						role: 'advertiser'
					},
					stats: {
						total_campaigns: 12,
						active_campaigns: 8,
						total_spent: 8500.00,
						total_impressions: 450000,
						total_clicks: 8900
					}
				},
				{
					id: 4,
					user_id: 104,
					company_name: 'Startup Ventures',
					contact_email: 'ads@startupventures.com',
					contact_phone: '+1-555-0321',
					website: 'https://startupventures.com',
					status: 'rejected',
					approval_notes: 'Business documentation incomplete',
					created_at: '2024-01-20T12:00:00Z',
					updated_at: '2024-01-21T09:45:00Z',
					user: {
						first_name: 'Emily',
						last_name: 'Davis',
						email: 'emily.davis@startupventures.com',
						role: 'advertiser'
					},
					stats: {
						total_campaigns: 0,
						active_campaigns: 0,
						total_spent: 0,
						total_impressions: 0,
						total_clicks: 0
					}
				}
			];

			totalAdvertisers = advertisers.length;
			totalPages = Math.ceil(totalAdvertisers / 10);

		} catch (err: any) {
			error = 'Failed to load advertisers';
			console.error('Advertisers load error:', err);
		} finally {
			loading = false;
		}
	}

	async function approveAdvertiser(id: number) {
		try {
			approvingId = id;
			
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 1000));
			
			// Update local state
			const advertiser = advertisers.find(a => a.id === id);
			if (advertiser) {
				advertiser.status = 'approved';
				advertiser.updated_at = new Date().toISOString();
			}

			showToast('Advertiser approved successfully', 'success');
		} catch (err: any) {
			showToast('Failed to approve advertiser', 'error');
			console.error('Approve error:', err);
		} finally {
			approvingId = null;
		}
	}

	async function rejectAdvertiser(id: number) {
		try {
			rejectingId = id;
			
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 1000));
			
			// Update local state
			const advertiser = advertisers.find(a => a.id === id);
			if (advertiser) {
				advertiser.status = 'rejected';
				advertiser.updated_at = new Date().toISOString();
			}

			showToast('Advertiser rejected successfully', 'success');
		} catch (err: any) {
			showToast('Failed to reject advertiser', 'error');
			console.error('Reject error:', err);
		} finally {
			rejectingId = null;
		}
	}

	async function suspendAdvertiser(id: number) {
		try {
			suspendingId = id;
			
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 1000));
			
			// Update local state
			const advertiser = advertisers.find(a => a.id === id);
			if (advertiser) {
				advertiser.status = 'suspended';
				advertiser.updated_at = new Date().toISOString();
			}

			showToast('Advertiser suspended successfully', 'success');
		} catch (err: any) {
			showToast('Failed to suspend advertiser', 'error');
			console.error('Suspend error:', err);
		} finally {
			suspendingId = null;
		}
	}

	function getStatusColor(status: string) {
		switch (status) {
			case 'approved': return 'success';
			case 'pending': return 'warning';
			case 'rejected': return 'error';
			case 'suspended': return 'error';
			default: return 'neutral';
		}
	}

	function getStatusIcon(status: string) {
		switch (status) {
			case 'approved': return '✓';
			case 'pending': return '⏳';
			case 'rejected': return '✗';
			case 'suspended': return '⏸';
			default: return '?';
		}
	}

	function formatCurrency(amount: number) {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(amount);
	}

	function formatNumber(num: number) {
		return new Intl.NumberFormat('en-US').format(num);
	}

	function formatDate(dateString: string) {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	// Filter advertisers based on search and status
	$: filteredAdvertisers = advertisers.filter(advertiser => {
		const matchesSearch = searchTerm === '' || 
			advertiser.company_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
			advertiser.contact_email.toLowerCase().includes(searchTerm.toLowerCase()) ||
			(advertiser.user?.first_name + ' ' + advertiser.user?.last_name).toLowerCase().includes(searchTerm.toLowerCase());
		
		const matchesStatus = statusFilter === 'all' || advertiser.status === statusFilter;
		
		return matchesSearch && matchesStatus;
	});

	// Pagination
	$: paginatedAdvertisers = filteredAdvertisers.slice((currentPage - 1) * 10, currentPage * 10);
	$: totalPages = Math.ceil(filteredAdvertisers.length / 10);
</script>

<svelte:head>
	<title>Advertiser Management - BOME Admin</title>
	<meta name="description" content="Manage advertisers and their accounts in BOME admin panel" />
</svelte:head>

<div class="advertisers-page">
	<div class="page-header">
		<div class="header-content">
			<div class="header-left">
				<h1>Advertiser Management</h1>
				<p>Manage advertiser accounts, approvals, and performance</p>
			</div>
			<div class="header-actions">
				<button class="refresh-button" on:click={loadAdvertisers}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8"></path>
						<path d="M21 3v5h-5"></path>
						<path d="M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16"></path>
						<path d="M3 21v-5h5"></path>
					</svg>
					Refresh
				</button>
			</div>
		</div>
	</div>

	{#if loading}
		<div class="loading-container">
			<LoadingSpinner size="large" />
			<p>Loading advertisers...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<div class="error-message">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="12" cy="12" r="10"></circle>
					<line x1="15" y1="9" x2="9" y2="15"></line>
					<line x1="9" y1="9" x2="15" y2="15"></line>
				</svg>
				{error}
			</div>
			<button class="retry-button" on:click={loadAdvertisers}>Retry</button>
		</div>
	{:else}
		<!-- Filters and Search -->
		<div class="filters-section">
			<div class="search-box">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="11" cy="11" r="8"></circle>
					<path d="m21 21-4.35-4.35"></path>
				</svg>
				<input 
					type="text" 
					placeholder="Search advertisers..." 
					bind:value={searchTerm}
				/>
			</div>
			
			<div class="filter-controls">
				<select bind:value={statusFilter}>
					<option value="all">All Status</option>
					<option value="pending">Pending</option>
					<option value="approved">Approved</option>
					<option value="rejected">Rejected</option>
					<option value="suspended">Suspended</option>
				</select>
			</div>
		</div>

		<!-- Stats Overview -->
		<div class="stats-overview">
			<div class="stat-card">
				<div class="stat-icon total">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
						<circle cx="9" cy="7" r="4"></circle>
						<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
						<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
					</svg>
				</div>
				<div class="stat-content">
					<div class="stat-value">{formatNumber(totalAdvertisers)}</div>
					<div class="stat-label">Total Advertisers</div>
				</div>
			</div>

			<div class="stat-card">
				<div class="stat-icon pending">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="12" cy="12" r="10"></circle>
						<polyline points="12,6 12,12 16,14"></polyline>
					</svg>
				</div>
				<div class="stat-content">
					<div class="stat-value">{formatNumber(advertisers.filter(a => a.status === 'pending').length)}</div>
					<div class="stat-label">Pending Approval</div>
				</div>
			</div>

			<div class="stat-card">
				<div class="stat-icon active">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M22 12h-4l-3 9L9 3l-3 9H2"></path>
					</svg>
				</div>
				<div class="stat-content">
					<div class="stat-value">{formatNumber(advertisers.filter(a => a.status === 'approved').length)}</div>
					<div class="stat-label">Active Advertisers</div>
				</div>
			</div>

			<div class="stat-card">
				<div class="stat-icon revenue">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="12" y1="1" x2="12" y2="23"></line>
						<path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"></path>
					</svg>
				</div>
				<div class="stat-content">
					<div class="stat-value">{formatCurrency(advertisers.reduce((sum, a) => sum + (a.stats?.total_spent || 0), 0))}</div>
					<div class="stat-label">Total Revenue</div>
				</div>
			</div>
		</div>

		<!-- Advertisers Table -->
		<div class="advertisers-table-container">
			<table class="advertisers-table">
				<thead>
					<tr>
						<th>Company</th>
						<th>Contact</th>
						<th>Status</th>
						<th>Campaigns</th>
						<th>Revenue</th>
						<th>Joined</th>
						<th>Actions</th>
					</tr>
				</thead>
				<tbody>
					{#each paginatedAdvertisers as advertiser}
						<tr>
							<td class="company-cell">
								<div class="company-info">
									<div class="company-name">{advertiser.company_name}</div>
									<div class="company-website">
										<a href={advertiser.website} target="_blank" rel="noopener noreferrer">
											{advertiser.website}
										</a>
									</div>
								</div>
							</td>
							<td class="contact-cell">
								<div class="contact-info">
									<div class="contact-name">
										{advertiser.user?.first_name} {advertiser.user?.last_name}
									</div>
									<div class="contact-email">{advertiser.contact_email}</div>
									<div class="contact-phone">{advertiser.contact_phone}</div>
								</div>
							</td>
							<td class="status-cell">
								<span class="status-badge {getStatusColor(advertiser.status)}">
									{getStatusIcon(advertiser.status)} {advertiser.status}
								</span>
							</td>
							<td class="campaigns-cell">
								<div class="campaigns-info">
									<div class="campaigns-count">
										{advertiser.stats?.active_campaigns || 0} / {advertiser.stats?.total_campaigns || 0}
									</div>
									<div class="campaigns-label">Active / Total</div>
								</div>
							</td>
							<td class="revenue-cell">
								<div class="revenue-info">
									<div class="revenue-amount">{formatCurrency(advertiser.stats?.total_spent || 0)}</div>
									<div class="revenue-stats">
										{formatNumber(advertiser.stats?.total_impressions || 0)} impressions
									</div>
								</div>
							</td>
							<td class="date-cell">
								{formatDate(advertiser.created_at)}
							</td>
							<td class="actions-cell">
								<div class="action-buttons">
									<a href="/admin/advertisers/{advertiser.id}" class="btn btn-secondary btn-sm">
										View Details
									</a>
									
									{#if advertiser.status === 'pending'}
										<button 
											class="btn btn-success btn-sm" 
											on:click={() => approveAdvertiser(advertiser.id)}
											disabled={approvingId === advertiser.id}
										>
											{approvingId === advertiser.id ? 'Approving...' : 'Approve'}
										</button>
										<button 
											class="btn btn-error btn-sm" 
											on:click={() => rejectAdvertiser(advertiser.id)}
											disabled={rejectingId === advertiser.id}
										>
											{rejectingId === advertiser.id ? 'Rejecting...' : 'Reject'}
										</button>
									{:else if advertiser.status === 'approved'}
										<button 
											class="btn btn-warning btn-sm" 
											on:click={() => suspendAdvertiser(advertiser.id)}
											disabled={suspendingId === advertiser.id}
										>
											{suspendingId === advertiser.id ? 'Suspending...' : 'Suspend'}
										</button>
									{:else if advertiser.status === 'rejected'}
										<button 
											class="btn btn-success btn-sm" 
											on:click={() => approveAdvertiser(advertiser.id)}
											disabled={approvingId === advertiser.id}
										>
											{approvingId === advertiser.id ? 'Approving...' : 'Approve'}
										</button>
									{:else if advertiser.status === 'suspended'}
										<button 
											class="btn btn-success btn-sm" 
											on:click={() => approveAdvertiser(advertiser.id)}
											disabled={approvingId === advertiser.id}
										>
											{approvingId === advertiser.id ? 'Approving...' : 'Approve'}
										</button>
									{/if}
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>

			{#if paginatedAdvertisers.length === 0}
				<div class="empty-state">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
						<circle cx="9" cy="7" r="4"></circle>
						<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
						<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
					</svg>
					<h3>No advertisers found</h3>
					<p>Try adjusting your search or filter criteria</p>
				</div>
			{/if}
		</div>

		<!-- Pagination -->
		{#if totalPages > 1}
			<div class="pagination">
				<button 
					class="pagination-btn" 
					disabled={currentPage === 1}
					on:click={() => currentPage--}
				>
					Previous
				</button>
				
				<div class="page-numbers">
					{#each Array.from({length: totalPages}, (_, i) => i + 1) as page}
						<button 
							class="page-btn" 
							class:active={page === currentPage}
							on:click={() => currentPage = page}
						>
							{page}
						</button>
					{/each}
				</div>
				
				<button 
					class="pagination-btn" 
					disabled={currentPage === totalPages}
					on:click={() => currentPage++}
				>
					Next
				</button>
			</div>
		{/if}
	{/if}
</div>

<style>
	.advertisers-page {
		min-height: 100vh;
		background: var(--bg-primary);
		color: var(--text-primary);
		padding: 2rem;
	}

	.page-header {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 20px;
		padding: 2rem;
		margin-bottom: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.header-content {
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: 1rem;
	}

	.header-left h1 {
		font-size: 2rem;
		font-weight: 700;
		margin: 0 0 0.5rem 0;
		color: var(--text-primary);
	}

	.header-left p {
		color: var(--text-secondary);
		margin: 0;
	}

	.header-actions {
		display: flex;
		gap: 1rem;
	}

	.refresh-button {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1rem;
		border: 1px solid rgba(255, 255, 255, 0.2);
		border-radius: 10px;
		background: var(--bg-glass);
		color: var(--text-primary);
		cursor: pointer;
		transition: all 0.3s ease;
		font-size: 0.9rem;
	}

	.refresh-button:hover {
		background: var(--primary);
		color: white;
		border-color: var(--primary);
	}

	.refresh-button svg {
		width: 16px;
		height: 16px;
	}

	.loading-container,
	.error-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 60vh;
		gap: 1rem;
	}

	.error-message {
		background: rgba(239, 68, 68, 0.1);
		border: 1px solid rgba(239, 68, 68, 0.3);
		border-radius: 10px;
		padding: 1rem;
		color: #fca5a5;
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.retry-button {
		background: var(--primary);
		color: white;
		border: none;
		border-radius: 8px;
		padding: 0.5rem 1rem;
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.retry-button:hover {
		background: var(--primary-dark);
	}

	.filters-section {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: 1.5rem;
		margin-bottom: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 1rem;
		flex-wrap: wrap;
	}

	.search-box {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		background: var(--bg-glass-dark);
		border-radius: 10px;
		padding: 0.75rem 1rem;
		flex: 1;
		min-width: 300px;
	}

	.search-box svg {
		width: 18px;
		height: 18px;
		color: var(--text-secondary);
	}

	.search-box input {
		background: transparent;
		border: none;
		color: var(--text-primary);
		font-size: 0.9rem;
		width: 100%;
	}

	.search-box input::placeholder {
		color: var(--text-secondary);
	}

	.filter-controls select {
		background: var(--bg-glass-dark);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 8px;
		padding: 0.75rem 1rem;
		color: var(--text-primary);
		font-size: 0.9rem;
		cursor: pointer;
	}

	.stats-overview {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 1.5rem;
		margin-bottom: 2rem;
	}

	.stat-card {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: 1.5rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
		display: flex;
		align-items: center;
		gap: 1rem;
		transition: all 0.3s ease;
	}

	.stat-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
	}

	.stat-icon {
		width: 60px;
		height: 60px;
		border-radius: 15px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.stat-icon svg {
		width: 28px;
		height: 28px;
		color: white;
	}

	.stat-icon.total {
		background: linear-gradient(135deg, #3b82f6, #1d4ed8);
	}

	.stat-icon.pending {
		background: linear-gradient(135deg, #f59e0b, #d97706);
	}

	.stat-icon.active {
		background: linear-gradient(135deg, #10b981, #047857);
	}

	.stat-icon.revenue {
		background: linear-gradient(135deg, #8b5cf6, #7c3aed);
	}

	.stat-content {
		flex: 1;
	}

	.stat-value {
		font-size: 1.75rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 0.25rem;
	}

	.stat-label {
		font-size: 0.9rem;
		color: var(--text-secondary);
	}

	.advertisers-table-container {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 15px;
		border: 1px solid rgba(255, 255, 255, 0.1);
		overflow: hidden;
		margin-bottom: 2rem;
	}

	.advertisers-table {
		width: 100%;
		border-collapse: collapse;
	}

	.advertisers-table th {
		background: var(--bg-glass-dark);
		padding: 1rem;
		text-align: left;
		font-weight: 600;
		color: var(--text-primary);
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
	}

	.advertisers-table td {
		padding: 1rem;
		border-bottom: 1px solid rgba(255, 255, 255, 0.05);
	}

	.advertisers-table tr:hover {
		background: rgba(255, 255, 255, 0.02);
	}

	.company-info .company-name {
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.25rem;
	}

	.company-info .company-website a {
		color: var(--primary);
		text-decoration: none;
		font-size: 0.85rem;
	}

	.company-info .company-website a:hover {
		text-decoration: underline;
	}

	.contact-info .contact-name {
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.25rem;
	}

	.contact-info .contact-email,
	.contact-info .contact-phone {
		font-size: 0.85rem;
		color: var(--text-secondary);
		margin-bottom: 0.125rem;
	}

	.status-badge {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.25rem 0.75rem;
		border-radius: 20px;
		font-size: 0.8rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.status-badge.success {
		background: rgba(16, 185, 129, 0.1);
		color: #10b981;
		border: 1px solid rgba(16, 185, 129, 0.3);
	}

	.status-badge.warning {
		background: rgba(245, 158, 11, 0.1);
		color: #f59e0b;
		border: 1px solid rgba(245, 158, 11, 0.3);
	}

	.status-badge.error {
		background: rgba(239, 68, 68, 0.1);
		color: #ef4444;
		border: 1px solid rgba(239, 68, 68, 0.3);
	}

	.status-badge.neutral {
		background: rgba(107, 114, 128, 0.1);
		color: #6b7280;
		border: 1px solid rgba(107, 114, 128, 0.3);
	}

	.campaigns-info .campaigns-count {
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.25rem;
	}

	.campaigns-info .campaigns-label {
		font-size: 0.85rem;
		color: var(--text-secondary);
	}

	.revenue-info .revenue-amount {
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.25rem;
	}

	.revenue-info .revenue-stats {
		font-size: 0.85rem;
		color: var(--text-secondary);
	}

	.action-buttons {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.btn {
		padding: 0.5rem 1rem;
		border: none;
		border-radius: 8px;
		font-size: 0.85rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.3s ease;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}

	.btn-sm {
		padding: 0.375rem 0.75rem;
		font-size: 0.8rem;
	}

	.btn-secondary {
		background: var(--bg-glass-dark);
		color: var(--text-primary);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.btn-secondary:hover {
		background: var(--bg-glass);
	}

	.btn-success {
		background: #10b981;
		color: white;
	}

	.btn-success:hover {
		background: #047857;
	}

	.btn-error {
		background: #ef4444;
		color: white;
	}

	.btn-error:hover {
		background: #dc2626;
	}

	.btn-warning {
		background: #f59e0b;
		color: white;
	}

	.btn-warning:hover {
		background: #d97706;
	}

	.btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 4rem 2rem;
		text-align: center;
	}

	.empty-state svg {
		width: 64px;
		height: 64px;
		color: var(--text-secondary);
		margin-bottom: 1rem;
	}

	.empty-state h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 0.5rem 0;
	}

	.empty-state p {
		color: var(--text-secondary);
		margin: 0;
	}

	.pagination {
		display: flex;
		justify-content: center;
		align-items: center;
		gap: 1rem;
		margin-top: 2rem;
	}

	.pagination-btn {
		padding: 0.5rem 1rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
		background: var(--bg-glass);
		color: var(--text-primary);
		border-radius: 8px;
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.pagination-btn:hover:not(:disabled) {
		background: var(--primary);
		color: white;
	}

	.pagination-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.page-numbers {
		display: flex;
		gap: 0.5rem;
	}

	.page-btn {
		width: 40px;
		height: 40px;
		border: 1px solid rgba(255, 255, 255, 0.1);
		background: var(--bg-glass);
		color: var(--text-primary);
		border-radius: 8px;
		cursor: pointer;
		transition: all 0.3s ease;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.page-btn:hover {
		background: var(--primary);
		color: white;
	}

	.page-btn.active {
		background: var(--primary);
		color: white;
	}

	@media (max-width: 768px) {
		.advertisers-page {
			padding: 1rem;
		}

		.header-content {
			flex-direction: column;
			align-items: stretch;
		}

		.filters-section {
			flex-direction: column;
			align-items: stretch;
		}

		.search-box {
			min-width: auto;
		}

		.stats-overview {
			grid-template-columns: 1fr;
		}

		.advertisers-table {
			font-size: 0.85rem;
		}

		.advertisers-table th,
		.advertisers-table td {
			padding: 0.75rem 0.5rem;
		}

		.action-buttons {
			flex-direction: row;
			flex-wrap: wrap;
		}
	}
</style> 
