<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import { masterVideoService } from '$lib/master-video';

	let analytics: any = null;
	let recentActivity: Array<{type: string; message: string; time: string; user?: string}> = [];
	let loading = true;
	let error = '';
	let userRole = '';
	let userPermissions: string[] = [];

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

		userRole = user.role;
		userPermissions = getUserPermissions(user.role);
		
		await loadDashboardData();
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

	function getUserPermissions(role: string): string[] {
		const permissionMap: Record<string, string[]> = {
			'super_admin': ['*'], // All permissions
			'system_admin': ['system:read', 'system:update', 'system:manage', 'analytics:read', 'analytics:export'],
			'content_manager': ['content:read', 'content:create', 'content:update', 'content:delete', 'videos:manage', 'analytics:read'],
			'articles_manager': ['articles:read', 'articles:create', 'articles:update', 'articles:delete', 'analytics:read'],
			'youtube_manager': ['videos:read', 'videos:create', 'videos:update', 'videos:delete', 'analytics:read'],
			'streaming_manager': ['videos:read', 'videos:create', 'videos:update', 'videos:delete', 'analytics:read'],
			'events_manager': ['events:read', 'events:create', 'events:update', 'events:delete', 'analytics:read'],
			'advertisement_manager': ['advertisements:read', 'advertisements:create', 'advertisements:update', 'advertisements:delete', 'analytics:read'],
			'user_manager': ['users:read', 'users:create', 'users:update', 'users:delete', 'analytics:read'],
			'analytics_manager': ['analytics:read', 'analytics:export', 'analytics:manage'],
			'financial_admin': ['financial:read', 'financial:manage', 'analytics:read'],
			'admin': ['*'] // Legacy admin - all permissions
		};
		return permissionMap[role] || [];
	}

	function hasPermission(permission: string): boolean {
		return userPermissions.includes('*') || userPermissions.includes(permission);
	}

	async function loadDashboardData() {
		loading = true;
		error = '';

		try {
			// Use the working MasterVideoService for comprehensive analytics
			const analyticsResponse = await masterVideoService.getDashboardAnalytics();

			if (analyticsResponse.success) {
				analytics = analyticsResponse.data;
				recentActivity = analytics.real_time?.recent_activity || [];
			} else {
				throw new Error('Failed to load dashboard analytics');
			}

		} catch (err: any) {
			error = 'Failed to load dashboard data';
			console.error('Dashboard load error:', err);
		} finally {
			loading = false;
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

	function getActivityIcon(type: string) {
		switch (type) {
			case 'user_signup':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
					<circle cx="8.5" cy="7" r="4"></circle>
					<line x1="20" y1="8" x2="20" y2="14"></line>
					<line x1="23" y1="11" x2="17" y2="11"></line>
				</svg>`;
			case 'video_upload':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<polygon points="23,7 16,12 23,17 23,7"></polygon>
					<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
				</svg>`;
			case 'subscription':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M20 7H4a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2z"></path>
					<polyline points="16,21 12,17 8,21"></polyline>
					<polyline points="12,17 12,3"></polyline>
				</svg>`;
			case 'payment':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<line x1="12" y1="1" x2="12" y2="23"></line>
					<path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"></path>
				</svg>`;
			case 'comment':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
				</svg>`;
			default:
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="12" cy="12" r="10"></circle>
					<line x1="12" y1="16" x2="12" y2="12"></line>
					<line x1="12" y1="8" x2="12.01" y2="8"></line>
				</svg>`;
		}
	}

	function setActivityIcon(element: HTMLElement, type: string) {
		element.innerHTML = getActivityIcon(type);
	}

	function activityIconAction(node: HTMLElement, type: string) {
		node.innerHTML = getActivityIcon(type);
		return {
			update(newType: string) {
				node.innerHTML = getActivityIcon(newType);
			}
		};
	}

	async function handleLogout() {
		await auth.logout();
		goto('/admin');
	}
</script>

<svelte:head>
	<title>Admin Dashboard - BOME</title>
	<meta name="description" content="BOME administrative dashboard" />
</svelte:head>

{#if loading}
	<div class="loading-container">
		<LoadingSpinner size="large" />
		<p>Loading dashboard...</p>
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
		<button class="retry-button" on:click={loadDashboardData}>Retry</button>
	</div>
{:else}
	<div class="admin-dashboard">
		<!-- Header -->
		<header class="dashboard-header">
			<div class="header-content">
				<div class="header-info">
					<h1>Admin Dashboard</h1>
					<p>Welcome back, {$auth.user?.first_name || 'Admin'}</p>
					<div class="role-badge">
						<span class="role-name">{userRole.replace('_', ' ').toUpperCase()}</span>
					</div>
				</div>
				<div class="header-actions">
					<button class="refresh-button" on:click={loadDashboardData}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8"></path>
							<path d="M21 3v5h-5"></path>
							<path d="M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16"></path>
							<path d="M3 21v-5h5"></path>
						</svg>
						Refresh
					</button>
					<button class="logout-button" on:click={handleLogout}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
							<polyline points="16,17 21,12 16,7"></polyline>
							<line x1="21" y1="12" x2="9" y2="12"></line>
						</svg>
						Logout
					</button>
				</div>
			</div>
		</header>

		<!-- Hub Overview Section -->
		<div class="hub-overview">
			<h2>Hub Overview</h2>
			<div class="hub-stats-grid">
				<div class="hub-stat-card">
					<div class="hub-stat-label">Total Users</div>
					<div class="hub-stat-value">{formatNumber(analytics?.total_users || 0)}</div>
				</div>
				<div class="hub-stat-card">
					<div class="hub-stat-label">Monthly Revenue</div>
					<div class="hub-stat-value">{formatCurrency(analytics?.subscriber_metrics?.monthly_revenue || 0)}</div>
				</div>
				<div class="hub-stat-card">
					<div class="hub-stat-label">Total Videos</div>
					<div class="hub-stat-value">{formatNumber(analytics?.total_videos || 0)}</div>
				</div>
			</div>
		</div>

		<!-- Subsites Overview Section -->
		<div class="subsites-overview">
			<h2>Subsites Overview</h2>
			<div class="subsites-grid">
				<!-- Streaming Subsite -->
				<div class="subsite-card active">
					<div class="subsite-header">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="32" height="32">
							<polygon points="23,7 16,12 23,17 23,7"></polygon>
							<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
						</svg>
						<div class="subsite-title">Streaming</div>
						<span class="subsite-status active">Active</span>
					</div>
					<div class="subsite-stats">
						<div><strong>{formatNumber(analytics?.total_users || 0)}</strong> users</div>
						<div><strong>{formatNumber(analytics?.total_videos || 0)}</strong> videos</div>
						<div><strong>{formatNumber(analytics?.total_views || 0)}</strong> views</div>
						<div><strong>{formatCurrency(analytics?.subscriber_metrics?.monthly_revenue || 0)}</strong> revenue</div>
					</div>
					<a href="/admin/streaming" class="subsite-link">Go to Streaming Admin</a>
				</div>

				<!-- Articles Subsite -->
				<div class="subsite-card under-construction">
					<div class="subsite-header">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="32" height="32">
							<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
							<polyline points="14,2 14,8 20,8"></polyline>
							<line x1="16" y1="13" x2="8" y2="13"></line>
							<line x1="16" y1="17" x2="8" y2="17"></line>
							<polyline points="10,9 9,9 8,9"></polyline>
						</svg>
						<div class="subsite-title">Articles</div>
						<span class="subsite-status">Under Construction</span>
					</div>
					<div class="subsite-stats">
						<div><strong>0</strong> articles</div>
						<div><strong>0</strong> users</div>
						<div class="subsite-progress">
							<progress value="25" max="100"></progress>
							<span>25% Complete</span>
						</div>
					</div>
					<span class="subsite-link disabled">Coming Soon</span>
				</div>

				<!-- Expo/Tours Subsite -->
				<div class="subsite-card under-construction">
					<div class="subsite-header">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="32" height="32">
							<path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path>
							<circle cx="12" cy="10" r="3"></circle>
						</svg>
						<div class="subsite-title">Expo & Tours</div>
						<span class="subsite-status">Under Construction</span>
					</div>
					<div class="subsite-stats">
						<div><strong>0</strong> events</div>
						<div><strong>0</strong> users</div>
						<div class="subsite-progress">
							<progress value="10" max="100"></progress>
							<span>10% Complete</span>
						</div>
					</div>
					<span class="subsite-link disabled">Coming Soon</span>
				</div>
			</div>
		</div>

		<!-- Stats Grid -->
		<div class="stats-grid">
			{#if hasPermission('analytics:read')}
				<div class="stat-card">
					<div class="stat-icon users">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
							<circle cx="9" cy="7" r="4"></circle>
							<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
							<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
						</svg>
					</div>
					<div class="stat-content">
						<div class="stat-value">{formatNumber(analytics?.total_users || 0)}</div>
						<div class="stat-label">Total Users</div>
						<div class="stat-change positive">+{analytics?.view_analytics?.views_today || 0} today</div>
					</div>
				</div>

				<div class="stat-card">
					<div class="stat-icon videos">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<polygon points="23,7 16,12 23,17 23,7"></polygon>
							<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
						</svg>
					</div>
					<div class="stat-content">
						<div class="stat-value">{formatNumber(analytics?.total_videos || 0)}</div>
						<div class="stat-label">Total Videos</div>
						<div class="stat-change">{formatNumber(analytics?.total_views || 0)} views</div>
					</div>
				</div>

				{#if hasPermission('financial:read')}
					<div class="stat-card">
						<div class="stat-icon revenue">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<line x1="12" y1="1" x2="12" y2="23"></line>
								<path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"></path>
							</svg>
						</div>
						<div class="stat-content">
							<div class="stat-value">{formatCurrency(analytics?.subscriber_metrics?.monthly_revenue || 0)}</div>
							<div class="stat-label">Monthly Revenue</div>
							<div class="stat-change">{formatNumber(analytics?.active_subscriptions || 0)} active</div>
						</div>
					</div>
				{/if}
			{/if}
		</div>

		<!-- Quick Actions -->
		<div class="quick-actions">
			<h2>Quick Actions</h2>
			<div class="actions-grid">
				{#if hasPermission('users:read')}
					<a href="/admin/users" class="action-card">
						<div class="action-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
								<circle cx="9" cy="7" r="4"></circle>
								<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
								<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
							</svg>
						</div>
						<div class="action-content">
							<div class="action-title">User Management</div>
							<div class="action-description">Manage users, roles, and permissions</div>
						</div>
					</a>
				{/if}

				{#if hasPermission('videos:read')}
					<a href="/admin/videos" class="action-card">
						<div class="action-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polygon points="23,7 16,12 23,17 23,7"></polygon>
								<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
							</svg>
						</div>
						<div class="action-content">
							<div class="action-title">Video Management</div>
							<div class="action-description">Upload, edit, and manage video content</div>
						</div>
					</a>
				{/if}

				{#if hasPermission('analytics:read')}
					<a href="/admin/analytics" class="action-card">
						<div class="action-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polyline points="22,12 18,12 15,21 9,3 6,12 2,12"></polyline>
							</svg>
						</div>
						<div class="action-content">
							<div class="action-title">Analytics</div>
							<div class="action-description">View detailed platform analytics</div>
						</div>
					</a>
				{/if}

				{#if hasPermission('system:read')}
					<a href="/admin/system" class="action-card">
						<div class="action-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect>
								<line x1="8" y1="21" x2="16" y2="21"></line>
								<line x1="12" y1="17" x2="12" y2="21"></line>
							</svg>
						</div>
						<div class="action-content">
							<div class="action-title">System Settings</div>
							<div class="action-description">Configure system settings and preferences</div>
						</div>
					</a>
				{/if}
			</div>
		</div>

		<!-- Recent Activity -->
		{#if hasPermission('analytics:read') && recentActivity.length > 0}
			<div class="recent-activity">
				<h2>Recent Activity</h2>
				<div class="activity-list">
					{#each recentActivity.slice(0, 5) as activity}
						<div class="activity-item">
							<div class="activity-icon" use:activityIconAction={activity.type}></div>
							<div class="activity-content">
								<div class="activity-message">{activity.message}</div>
								<div class="activity-meta">
									{#if activity.user}
										<span class="activity-user">{activity.user}</span>
									{/if}
									<span class="activity-time">{activity.time}</span>
								</div>
							</div>
						</div>
					{/each}
				</div>
			</div>
		{/if}
	</div>
{/if}

<style>
	.admin-dashboard {
		min-height: 100vh;
		background: var(--bg-primary);
		color: var(--text-primary);
		padding: 2rem;
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

	.dashboard-header {
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

	.header-info h1 {
		font-size: 2rem;
		font-weight: 700;
		margin: 0 0 0.5rem 0;
		color: var(--text-primary);
	}

	.header-info p {
		color: var(--text-secondary);
		margin: 0 0 1rem 0;
	}

	.role-badge {
		display: inline-block;
	}

	.role-name {
		background: var(--primary-gradient);
		color: white;
		padding: 0.25rem 0.75rem;
		border-radius: 20px;
		font-size: 0.8rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.header-actions {
		display: flex;
		gap: 1rem;
	}

	.refresh-button,
	.logout-button {
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

	.logout-button:hover {
		background: #ef4444;
		color: white;
		border-color: #ef4444;
	}

	.refresh-button svg,
	.logout-button svg {
		width: 16px;
		height: 16px;
	}

	.stats-grid {
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

	.stat-icon.users {
		background: linear-gradient(135deg, #3b82f6, #1d4ed8);
	}

	.stat-icon.videos {
		background: linear-gradient(135deg, #10b981, #047857);
	}

	.stat-icon.revenue {
		background: linear-gradient(135deg, #f59e0b, #d97706);
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
		margin-bottom: 0.5rem;
	}

	.stat-change {
		font-size: 0.8rem;
		font-weight: 600;
	}

	.stat-change.positive {
		color: #10b981;
	}

	.quick-actions {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 20px;
		padding: 2rem;
		margin-bottom: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.quick-actions h2 {
		font-size: 1.5rem;
		font-weight: 600;
		margin: 0 0 1.5rem 0;
		color: var(--text-primary);
	}

	.actions-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 1rem;
	}

	.action-card {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1.5rem;
		background: var(--bg-glass-dark);
		border-radius: 15px;
		text-decoration: none;
		color: var(--text-primary);
		transition: all 0.3s ease;
		border: 1px solid rgba(255, 255, 255, 0.05);
	}

	.action-card:hover {
		background: var(--bg-glass);
		transform: translateY(-2px);
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
	}

	.action-icon {
		width: 50px;
		height: 50px;
		background: var(--primary-gradient);
		border-radius: 12px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.action-icon svg {
		width: 24px;
		height: 24px;
		color: white;
	}

	.action-title {
		font-weight: 600;
		font-size: 1rem;
		margin-bottom: 0.25rem;
	}

	.action-description {
		font-size: 0.85rem;
		color: var(--text-secondary);
		line-height: 1.4;
	}

	.recent-activity {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 20px;
		padding: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.recent-activity h2 {
		font-size: 1.5rem;
		font-weight: 600;
		margin: 0 0 1.5rem 0;
		color: var(--text-primary);
	}

	.activity-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.activity-item {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1rem;
		background: var(--bg-glass-dark);
		border-radius: 12px;
		transition: all 0.3s ease;
	}

	.activity-item:hover {
		background: var(--bg-glass);
	}

	.activity-icon {
		width: 40px;
		height: 40px;
		background: var(--primary-gradient);
		border-radius: 10px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.activity-content {
		flex: 1;
	}

	.activity-message {
		font-weight: 500;
		margin-bottom: 0.25rem;
	}

	.activity-meta {
		display: flex;
		gap: 1rem;
		font-size: 0.8rem;
		color: var(--text-secondary);
	}

	.activity-user {
		font-weight: 600;
		color: var(--primary);
	}

	@media (max-width: 768px) {
		.admin-dashboard {
			padding: 1rem;
		}

		.header-content {
			flex-direction: column;
			align-items: stretch;
		}

		.header-actions {
			justify-content: center;
		}

		.stats-grid {
			grid-template-columns: 1fr;
		}

		.actions-grid {
			grid-template-columns: 1fr;
		}

		.stat-card {
			flex-direction: column;
			text-align: center;
		}

		.action-card {
			flex-direction: column;
			text-align: center;
		}
	}

.hub-overview {
	background: var(--bg-glass);
	backdrop-filter: blur(20px);
	border-radius: 20px;
	padding: 2rem;
	margin-bottom: 2rem;
	border: 1px solid rgba(255, 255, 255, 0.1);
}
.hub-overview h2 {
	font-size: 1.5rem;
	font-weight: 600;
	margin: 0 0 1.5rem 0;
	color: var(--text-primary);
}
.hub-stats-grid {
	display: grid;
	grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
	gap: 1.5rem;
	margin-bottom: 1rem;
}
.hub-stat-card {
	background: var(--bg-glass-dark);
	border-radius: 15px;
	padding: 1.25rem;
	text-align: center;
	box-shadow: 0 2px 8px rgba(0,0,0,0.04);
}
.hub-stat-label {
	font-size: 0.95rem;
	color: var(--text-secondary);
	margin-bottom: 0.5rem;
}
.hub-stat-value {
	font-size: 1.5rem;
	font-weight: 700;
	color: var(--text-primary);
}
.subsites-overview {
	background: var(--bg-glass);
	backdrop-filter: blur(20px);
	border-radius: 20px;
	padding: 2rem;
	margin-bottom: 2rem;
	border: 1px solid rgba(255, 255, 255, 0.1);
}
.subsites-overview h2 {
	font-size: 1.5rem;
	font-weight: 600;
	margin: 0 0 1.5rem 0;
	color: var(--text-primary);
}
.subsites-grid {
	display: grid;
	grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
	gap: 1.5rem;
}
.subsite-card {
	background: var(--bg-glass-dark);
	border-radius: 15px;
	padding: 1.5rem;
	display: flex;
	flex-direction: column;
	align-items: flex-start;
	gap: 1rem;
	box-shadow: 0 2px 8px rgba(0,0,0,0.04);
	position: relative;
}
.subsite-card.active {
	border: 2px solid #10b981;
}
.subsite-card.under-construction {
	opacity: 0.7;
}
.subsite-header {
	display: flex;
	align-items: center;
	gap: 0.75rem;
}
.subsite-title {
	font-weight: 700;
	font-size: 1.1rem;
	color: var(--text-primary);
}
.subsite-status {
	font-size: 0.8rem;
	font-weight: 600;
	margin-left: 0.5rem;
	color: #f59e0b;
}
.subsite-card.active .subsite-status {
	color: #10b981;
}
.subsite-stats {
	display: flex;
	flex-direction: column;
	gap: 0.25rem;
	font-size: 0.95rem;
	color: var(--text-secondary);
}
.subsite-link {
	margin-top: 0.5rem;
	font-size: 0.95rem;
	color: var(--primary);
	text-decoration: underline;
	cursor: pointer;
}
.subsite-link.disabled {
	color: #aaa;
	text-decoration: none;
	cursor: not-allowed;
}
.subsite-progress {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	margin-top: 0.25rem;
}
progress {
	width: 80px;
	height: 8px;
	appearance: none;
	background: #eee;
	border-radius: 4px;
}
progress::-webkit-progress-bar {
	background: #eee;
	border-radius: 4px;
}
progress::-webkit-progress-value {
	background: #f59e0b;
	border-radius: 4px;
}
progress::-moz-progress-bar {
	background: #f59e0b;
	border-radius: 4px;
}
</style> 