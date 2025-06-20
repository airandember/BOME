<script lang="ts">
	import { onMount } from 'svelte';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import type { AdminAnalytics } from '$lib/types/advertising';

	interface Activity {
		type: 'user_signup' | 'video_upload' | 'subscription' | 'payment' | 'comment';
		message: string;
		time: string;
	}

	let stats = {
		totalUsers: 0,
		totalVideos: 0,
		totalRevenue: 0,
		activeSubscriptions: 0,
		recentSignups: 0,
		videoViews: 0
	};

	let recentActivity: Activity[] = [];
	let systemHealth = {
		status: 'healthy',
		uptime: '99.9%',
		lastBackup: '2 hours ago',
		storageUsed: '45%'
	};

	let analytics: AdminAnalytics | null = null;
	let loading = true;
	let error: string | null = null;

	onMount(async () => {
		if (!$auth.isAuthenticated) {
			goto('/login');
			return;
		}

		const user = $auth.user;
		if (!user || user.role !== 'admin') {
			goto('/');
			return;
		}

		await loadAnalytics();
	});

	async function loadAnalytics() {
		loading = true;
		error = null;
		
		try {
			const response = await fetch('/api/v1/admin/analytics', {
				headers: {
					'Authorization': `Bearer ${$auth.token}`
				}
			});

			if (response.ok) {
				const data = await response.json();
				analytics = data.data;
			} else {
				// Mock analytics data
				analytics = {
					total_advertisers: 23,
					active_campaigns: 12,
					total_revenue: 15420.80,
					pending_approvals: 5,
					top_performing_placements: [
						{
							placement_id: 1,
							name: 'Header Banner',
							revenue: 8520.30,
							impressions: 45230
						},
						{
							placement_id: 2,
							name: 'Sidebar Large',
							revenue: 4890.50,
							impressions: 28940
						},
						{
							placement_id: 3,
							name: 'Between Videos',
							revenue: 2010.00,
							impressions: 15680
						}
					],
					revenue_by_month: [
						{ month: 'Jan', revenue: 12450.80, advertisers: 18 },
						{ month: 'Feb', revenue: 13890.20, advertisers: 20 },
						{ month: 'Mar', revenue: 15420.80, advertisers: 23 },
						{ month: 'Apr', revenue: 18230.50, advertisers: 25 },
						{ month: 'May', revenue: 21340.90, advertisers: 28 },
						{ month: 'Jun', revenue: 19850.60, advertisers: 26 }
					]
				};
			}
		} catch (err) {
			error = 'Failed to load analytics';
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
</script>

<svelte:head>
	<title>Admin Dashboard - BOME</title>
</svelte:head>

{#if loading}
	<div class="loading-container">
		<LoadingSpinner size="large" color="primary" />
		<p>Loading dashboard...</p>
	</div>
{:else if error}
	<div class="error-container">
		<div class="alert alert-error">
			<span>{error}</span>
		</div>
	</div>
{:else if analytics}
	<div class="admin-dashboard">
		<div class="dashboard-header">
			<div>
				<h1>Admin Dashboard</h1>
				<p>Overview of platform performance and advertising metrics</p>
			</div>
			<div class="header-actions">
				<button class="btn btn-secondary" on:click={loadAnalytics}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8" />
						<path d="M21 3v5h-5" />
						<path d="M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16" />
						<path d="M3 21v-5h5" />
					</svg>
					Refresh
				</button>
			</div>
		</div>

		<!-- Key Metrics -->
		<div class="metrics-grid">
			<div class="metric-card">
				<div class="metric-header">
					<div class="metric-icon advertisers">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
							<circle cx="9" cy="7" r="4" />
							<path d="M23 21v-2a4 4 0 0 0-3-3.87" />
							<path d="M16 3.13a4 4 0 0 1 0 7.75" />
						</svg>
					</div>
					<div class="metric-value">{analytics.total_advertisers}</div>
				</div>
				<div class="metric-label">Total Advertisers</div>
				<div class="metric-change positive">+3 this month</div>
			</div>

			<div class="metric-card">
				<div class="metric-header">
					<div class="metric-icon campaigns">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<rect x="2" y="3" width="20" height="14" rx="2" ry="2" />
							<line x1="8" y1="21" x2="16" y2="21" />
							<line x1="12" y1="17" x2="12" y2="21" />
						</svg>
					</div>
					<div class="metric-value">{analytics.active_campaigns}</div>
				</div>
				<div class="metric-label">Active Campaigns</div>
				<div class="metric-change positive">+2 this week</div>
			</div>

			<div class="metric-card">
				<div class="metric-header">
					<div class="metric-icon revenue">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<line x1="12" y1="1" x2="12" y2="23" />
							<path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
						</svg>
					</div>
					<div class="metric-value">{formatCurrency(analytics.total_revenue)}</div>
				</div>
				<div class="metric-label">Total Revenue</div>
				<div class="metric-change positive">+15.7% vs last month</div>
			</div>

			<div class="metric-card">
				<div class="metric-header">
					<div class="metric-icon pending">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<circle cx="12" cy="12" r="10" />
							<polyline points="12,6 12,12 16,14" />
						</svg>
					</div>
					<div class="metric-value">{analytics.pending_approvals}</div>
				</div>
				<div class="metric-label">Pending Approvals</div>
				<a href="/admin/advertisements" class="metric-action">Review →</a>
			</div>
		</div>

		<!-- Charts Section -->
		<div class="charts-section">
			<!-- Revenue Chart -->
			<div class="chart-card">
				<div class="chart-header">
					<h3>Revenue Trend</h3>
					<div class="chart-period">Last 6 months</div>
				</div>
				<div class="revenue-chart">
					{#each analytics.revenue_by_month as month}
						<div class="month-bar">
							<div 
								class="bar-fill" 
								style="height: {(month.revenue / 25000) * 100}%"
								title="{month.month}: {formatCurrency(month.revenue)}"
							></div>
							<div class="month-label">{month.month}</div>
						</div>
					{/each}
				</div>
			</div>

			<!-- Top Placements -->
			<div class="chart-card">
				<div class="chart-header">
					<h3>Top Performing Placements</h3>
					<a href="/admin/placements" class="chart-action">View All →</a>
				</div>
				<div class="placements-list">
					{#each analytics.top_performing_placements as placement}
						<div class="placement-item">
							<div class="placement-info">
								<div class="placement-name">{placement.name}</div>
								<div class="placement-stats">
									{formatNumber(placement.impressions)} impressions
								</div>
							</div>
							<div class="placement-revenue">
								{formatCurrency(placement.revenue)}
							</div>
						</div>
					{/each}
				</div>
			</div>
		</div>

		<!-- Quick Actions -->
		<div class="quick-actions">
			<h3>Quick Actions</h3>
			<div class="actions-grid">
				<a href="/admin/advertisements" class="action-card">
					<div class="action-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M9 12l2 2 4-4" />
							<path d="M21 12c.552 0 1-.448 1-1V5c0-.552-.448-1-1-1H3c-.552 0-1 .448-1 1v6c0 .552.448 1 1 1h18z" />
							<path d="M3 12v6c0 .552.448 1 1 1h16c.552 0 1-.448 1-1v-6" />
						</svg>
					</div>
					<div class="action-content">
						<div class="action-title">Review Approvals</div>
						<div class="action-description">Approve pending advertisers and campaigns</div>
					</div>
				</a>

				<a href="/admin/placements" class="action-card">
					<div class="action-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M12 2L2 7l10 5 10-5-10-5z" />
							<path d="M2 17l10 5 10-5" />
							<path d="M2 12l10 5 10-5" />
						</svg>
					</div>
					<div class="action-content">
						<div class="action-title">Manage Placements</div>
						<div class="action-description">Configure ad placement locations and rates</div>
					</div>
				</a>

				<a href="/admin/analytics" class="action-card">
					<div class="action-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<polyline points="22,12 18,12 15,21 9,3 6,12 2,12" />
						</svg>
					</div>
					<div class="action-content">
						<div class="action-title">View Analytics</div>
						<div class="action-description">Detailed performance and revenue analytics</div>
					</div>
				</a>

				<a href="/admin/users" class="action-card">
					<div class="action-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
							<circle cx="9" cy="7" r="4" />
							<path d="M23 21v-2a4 4 0 0 0-3-3.87" />
							<path d="M16 3.13a4 4 0 0 1 0 7.75" />
						</svg>
					</div>
					<div class="action-content">
						<div class="action-title">User Management</div>
						<div class="action-description">Manage users, roles, and permissions</div>
					</div>
				</a>
			</div>
		</div>
	</div>
{/if}

<style>
	.admin-dashboard {
		display: flex;
		flex-direction: column;
		gap: var(--space-xl);
	}

	.dashboard-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: var(--space-lg);
	}

	.dashboard-header h1 {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--text-primary);
		margin: 0;
	}

	.dashboard-header p {
		color: var(--text-secondary);
		margin: var(--space-sm) 0 0 0;
	}

	.header-actions {
		display: flex;
		gap: var(--space-md);
	}

	.loading-container,
	.error-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--space-4xl);
		gap: var(--space-lg);
	}

	.loading-spinner {
		width: 40px;
		height: 40px;
		border: 3px solid var(--border-color);
		border-top: 3px solid var(--primary);
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.metrics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: var(--space-lg);
	}

	.metric-card {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
		transition: all var(--transition-normal);
	}

	.metric-card:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}

	.metric-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: var(--space-md);
	}

	.metric-icon {
		width: 48px;
		height: 48px;
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.metric-icon svg {
		width: 24px;
		height: 24px;
		color: white;
	}

	.metric-icon.advertisers {
		background: linear-gradient(135deg, #3b82f6, #1d4ed8);
	}

	.metric-icon.campaigns {
		background: linear-gradient(135deg, #10b981, #047857);
	}

	.metric-icon.revenue {
		background: linear-gradient(135deg, #f59e0b, #d97706);
	}

	.metric-icon.pending {
		background: linear-gradient(135deg, #ef4444, #dc2626);
	}

	.metric-value {
		font-size: var(--text-2xl);
		font-weight: 700;
		color: var(--text-primary);
	}

	.metric-label {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		margin-bottom: var(--space-xs);
	}

	.metric-change {
		font-size: var(--text-xs);
		font-weight: 600;
	}

	.metric-change.positive {
		color: var(--success);
	}

	.metric-action {
		font-size: var(--text-xs);
		color: var(--primary);
		text-decoration: none;
		font-weight: 600;
	}

	.metric-action:hover {
		text-decoration: underline;
	}

	.charts-section {
		display: grid;
		grid-template-columns: 2fr 1fr;
		gap: var(--space-xl);
	}

	.chart-card {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.chart-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-lg);
	}

	.chart-header h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.chart-period {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.chart-action {
		font-size: var(--text-sm);
		color: var(--primary);
		text-decoration: none;
		font-weight: 600;
	}

	.chart-action:hover {
		text-decoration: underline;
	}

	.revenue-chart {
		display: flex;
		align-items: end;
		gap: var(--space-md);
		height: 200px;
		padding: var(--space-lg) 0;
	}

	.month-bar {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-sm);
	}

	.bar-fill {
		width: 100%;
		background: var(--primary-gradient);
		border-radius: var(--radius-sm);
		min-height: 10px;
		transition: all 0.5s ease;
		cursor: pointer;
	}

	.bar-fill:hover {
		opacity: 0.8;
		transform: scaleY(1.05);
	}

	.month-label {
		font-size: var(--text-xs);
		color: var(--text-secondary);
		font-weight: 600;
	}

	.placements-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.placement-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-md);
		background: var(--bg-glass-dark);
		border-radius: var(--radius-lg);
		transition: all var(--transition-normal);
	}

	.placement-item:hover {
		background: var(--bg-glass);
		transform: translateX(4px);
	}

	.placement-name {
		font-weight: 600;
		color: var(--text-primary);
		font-size: var(--text-sm);
	}

	.placement-stats {
		font-size: var(--text-xs);
		color: var(--text-secondary);
		margin-top: var(--space-xs);
	}

	.placement-revenue {
		font-weight: 700;
		color: var(--primary);
		font-size: var(--text-sm);
	}

	.quick-actions {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.quick-actions h3 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-lg) 0;
	}

	.actions-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: var(--space-lg);
	}

	.action-card {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		padding: var(--space-lg);
		background: var(--bg-glass-dark);
		border-radius: var(--radius-lg);
		text-decoration: none;
		transition: all var(--transition-normal);
		border: 1px solid rgba(255, 255, 255, 0.05);
	}

	.action-card:hover {
		background: var(--bg-glass);
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}

	.action-icon {
		width: 48px;
		height: 48px;
		background: var(--primary-gradient);
		border-radius: var(--radius-lg);
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
		color: var(--text-primary);
		font-size: var(--text-base);
		margin-bottom: var(--space-xs);
	}

	.action-description {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		line-height: 1.4;
	}

	/* Responsive Design */
	@media (max-width: 1200px) {
		.charts-section {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 768px) {
		.dashboard-header {
			flex-direction: column;
			align-items: stretch;
		}

		.metrics-grid {
			grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		}

		.actions-grid {
			grid-template-columns: 1fr;
		}

		.action-card {
			flex-direction: column;
			text-align: center;
		}

		.revenue-chart {
			height: 150px;
		}
	}
</style> 