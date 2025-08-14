<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { browser } from '$app/environment';
	import { auth, apiRequest } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import { goto } from '$app/navigation';
	import { AnalyticsService } from '$lib/services/analytics';
	import type { AnalyticsResponse, RealTimeMetrics, SystemHealth } from '$lib/services/analytics';

	interface DashboardData {
		users: {
			total: number;
			new_today: number;
			active_today: number;
			growth_rate: number;
		};
		content: {
			total_videos: number;
			total_articles: number;
			total_events: number;
			total_views: number;
		};
		revenue: {
			total_monthly: number;
			total_yearly: number;
			mrr: number;
			growth_rate: number;
		};
		system_health: {
			status: 'healthy' | 'warning' | 'critical';
			uptime: string;
			response_time: string;
			error_rate: string;
			active_sessions: number;
		};
		real_time: {
			active_users: number;
			current_streams: number;
			server_load: number;
			bandwidth_usage: string;
		};
		bottlenecks: {
			high_error_rate: boolean;
			slow_response: boolean;
			high_cpu: boolean;
			disk_space: boolean;
		};
	}

	let dashboardData: DashboardData | null = null;
	let loading = true;
	let error: string | null = null;
	let refreshInterval: NodeJS.Timeout | null = null;
	let lastUpdated = new Date();
let visibilityHandler: (() => void) | null = null;

onMount(async () => {
    // Check authentication first
    if (!$auth.isAuthenticated) {
        showToast('Please log in to access analytics', 'error');
        goto('/admin');
        return;
    }

    // Check if user has admin privileges
    const user = $auth.user;
    if (!user || !isAdminUser(user)) {
        showToast('Access denied. Admin privileges required.', 'error');
        goto('/admin');
        return;
    }

    await loadDashboard();

    const startPolling = () => {
        if (refreshInterval) clearInterval(refreshInterval);
        refreshInterval = setInterval(loadDashboard, 15000); // 15s for dashboard
    };
    const stopPolling = () => {
        if (refreshInterval) {
            clearInterval(refreshInterval);
            refreshInterval = null;
        }
    };
    visibilityHandler = () => {
        if (document.hidden) stopPolling();
        else startPolling();
    };
    document.addEventListener('visibilitychange', visibilityHandler);
    startPolling();
});

onDestroy(() => {
    if (refreshInterval) {
        clearInterval(refreshInterval);
        refreshInterval = null;
    }
    if (visibilityHandler) {
        document.removeEventListener('visibilitychange', visibilityHandler);
        visibilityHandler = null;
    }
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

	async function loadDashboard() {
		try {
			loading = true;
			error = null;

			// Check if user is still authenticated
			if (!$auth.isAuthenticated) {
				throw new Error('Authentication required');
			}

			// Fetch analytics data (hub-level)
			const response = await apiRequest('/admin/analytics');

			if (!response.ok) {
				if (response.status === 401) {
					throw new Error('Authentication expired. Please log in again.');
				} else if (response.status === 403) {
					throw new Error('Access denied. Admin privileges required.');
				} else {
					throw new Error(`Failed to load analytics (${response.status})`);
				}
			}

			const data = await response.json();
			
			// Fetch system health data
			let systemHealth = null;
			try {
				const healthResponse = await apiRequest('/admin/analytics/system-health');
				if (healthResponse.ok) {
					const healthData = await healthResponse.json();
					systemHealth = healthData.data;
				}
			} catch (healthError) {
				console.warn('Failed to fetch system health:', healthError);
			}

			// Transform the data into our dashboard format
			dashboardData = {
				users: {
					total: data.data?.users?.total || 0,
					new_today: data.data?.users?.new_today || 0,
					active_today: data.data?.users?.active_today || 0,
					growth_rate: data.data?.users?.growth_rate || 0
				},
				content: {
					total_videos: data.data?.videos?.total || 0,
					total_articles: 0, // Will be populated from articles subsite
					total_events: 0,   // Will be populated from expo subsite
					total_views: data.data?.videos?.total_views || 0
				},
				revenue: {
					total_monthly: data.data?.subscriptions?.revenue_month || 0,
					total_yearly: data.data?.subscriptions?.revenue_year || 0,
					mrr: data.data?.subscriptions?.mrr || 0,
					growth_rate: 0.15 // Mock growth rate
				},
				system_health: {
					status: systemHealth?.status || 'unknown',
					uptime: systemHealth?.uptime || 'Unknown',
					response_time: systemHealth?.response_time || 'Unknown',
					error_rate: systemHealth?.error_rate || 'Unknown',
					active_sessions: systemHealth?.active_sessions || data.data?.real_time?.active_users || 0
				},
				real_time: {
					active_users: data.data?.real_time?.active_users || 0,
					current_streams: data.data?.real_time?.current_streams || 0,
					server_load: data.data?.real_time?.server_load || 0,
					bandwidth_usage: data.data?.real_time?.bandwidth_usage || '0 MB/s'
				},
				bottlenecks: {
					high_error_rate: parseFloat(systemHealth?.error_rate || '0') > 5,
					slow_response: parseInt(systemHealth?.response_time || '0') > 1000,
					high_cpu: (systemHealth?.cpu?.usage || 0) > 80,
					disk_space: (systemHealth?.disk?.percent || 0) > 90
				}
			};

			lastUpdated = new Date();

		} catch (err) {
			error = err instanceof Error ? err.message : 'Unknown error';
			console.error('Dashboard load error:', err);
			
			if (error.includes('Authentication')) {
				showToast(error, 'error');
				goto('/admin');
				return;
			}
			
			showToast('Failed to load dashboard', 'error');
		} finally {
			loading = false;
		}
	}

	function formatNumber(num: number): string {
		return new Intl.NumberFormat().format(num);
	}

	function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(amount);
	}

	function getStatusColor(status: string): string {
		switch (status) {
			case 'healthy': return 'text-green-500';
			case 'warning': return 'text-yellow-500';
			case 'critical': return 'text-red-500';
			default: return 'text-gray-500';
		}
	}

	function getStatusIcon(status: string): string {
		switch (status) {
			case 'healthy': return '✓';
			case 'warning': return '⚠';
			case 'critical': return '✗';
			default: return '?';
		}
	}

	function getBottleneckCount(): number {
		if (!dashboardData || !dashboardData.bottlenecks) return 0;
		return Object.values(dashboardData.bottlenecks).filter(Boolean).length;
	}
</script>

<svelte:head>
	<title>Analytics Dashboard - BOME Admin</title>
</svelte:head>

<div class="dashboard">
	<!-- Header -->
	<div class="header">
		<div class="header-content">
			<h1>Analytics Dashboard</h1>
			<p class="subtitle">High-level metrics and system health overview</p>
		</div>
		<div class="header-actions">
			<div class="last-updated">
				Last updated: {lastUpdated.toLocaleTimeString()}
			</div>
			<button class="refresh-btn" on:click={loadDashboard} disabled={loading}>
				{loading ? 'Refreshing...' : 'Refresh'}
			</button>
		</div>
	</div>

	{#if loading && !dashboardData}
		<div class="loading-container">
			<LoadingSpinner />
			<p>Loading dashboard...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<div class="error-icon">⚠️</div>
			<h3>Failed to Load Dashboard</h3>
			<p class="error-message">{error}</p>
			<button class="retry-btn" on:click={loadDashboard}>Try Again</button>
		</div>
	{:else if dashboardData}
		<!-- System Health Alert -->
		{#if dashboardData && dashboardData.system_health && dashboardData.system_health.status !== 'healthy'}
			<div class="alert {dashboardData.system_health.status || 'warning'}">
				<div class="alert-icon">
					{getStatusIcon(dashboardData.system_health.status || 'warning')}
				</div>
				<div class="alert-content">
					<h3>System Health Alert</h3>
					<p>System status: {(dashboardData.system_health.status || 'unknown').toUpperCase()}</p>
				</div>
				<a href="/admin/monitoring" class="alert-action">View Details</a>
			</div>
		{/if}

		<!-- Bottlenecks Alert -->
		{#if getBottleneckCount() > 0}
			<div class="alert warning">
				<div class="alert-icon">⚠</div>
				<div class="alert-content">
					<h3>Performance Bottlenecks Detected</h3>
					<p>{getBottleneckCount()} performance issue{getBottleneckCount() > 1 ? 's' : ''} detected</p>
				</div>
				<a href="/admin/monitoring" class="alert-action">Investigate</a>
			</div>
		{/if}

		<!-- Key Metrics Grid -->
		<div class="metrics-grid">
			<div class="metric-card primary">
				<div class="metric-header">
					<h3>Total Users</h3>
					<div class="metric-icon">👥</div>
				</div>
				<div class="metric-value">{formatNumber(dashboardData && dashboardData.users ? dashboardData.users.total : 0)}</div>
				<div class="metric-details">
					<span class="metric-change positive">+{formatNumber(dashboardData && dashboardData.users ? dashboardData.users.new_today : 0)} today</span>
					<span class="metric-growth">+{(dashboardData && dashboardData.users ? dashboardData.users.growth_rate : 0).toFixed(1)}%</span>
				</div>
			</div>

			<div class="metric-card">
				<div class="metric-header">
					<h3>Active Users</h3>
					<div class="metric-icon">🟢</div>
				</div>
				<div class="metric-value">{formatNumber(dashboardData && dashboardData.real_time ? dashboardData.real_time.active_users : 0)}</div>
				<div class="metric-details">
					<span class="metric-change">Live</span>
				</div>
			</div>

			<div class="metric-card">
				<div class="metric-header">
					<h3>Content Views</h3>
					<div class="metric-icon">👁️</div>
				</div>
				<div class="metric-value">{formatNumber(dashboardData && dashboardData.content ? dashboardData.content.total_views : 0)}</div>
				<div class="metric-details">
					<span class="metric-change">Total</span>
				</div>
			</div>

			<div class="metric-card revenue">
				<div class="metric-header">
					<h3>Monthly Revenue</h3>
					<div class="metric-icon">💰</div>
				</div>
				<div class="metric-value">{formatCurrency(dashboardData && dashboardData.revenue ? dashboardData.revenue.total_monthly : 0)}</div>
				<div class="metric-details">
					<span class="metric-change positive">+{(dashboardData && dashboardData.revenue ? dashboardData.revenue.growth_rate : 0).toFixed(1)}%</span>
					<span class="metric-mrr">MRR: {formatCurrency(dashboardData && dashboardData.revenue ? dashboardData.revenue.mrr : 0)}</span>
				</div>
			</div>
		</div>

		<!-- System Health Overview -->
		<div class="health-section">
			<div class="section-header">
				<h2>System Health</h2>
				<div class="status-badge {getStatusColor(dashboardData && dashboardData.system_health ? dashboardData.system_health.status : 'unknown')}">
					{getStatusIcon(dashboardData && dashboardData.system_health ? dashboardData.system_health.status : 'unknown')} {(dashboardData && dashboardData.system_health ? dashboardData.system_health.status : 'unknown').toUpperCase()}
				</div>
			</div>
			<div class="health-grid">
				<div class="health-card">
					<div class="health-metric">
						<span class="label">Uptime</span>
						<span class="value">{dashboardData && dashboardData.system_health ? dashboardData.system_health.uptime : 'Unknown'}</span>
					</div>
				</div>
				<div class="health-card">
					<div class="health-metric">
						<span class="label">Response Time</span>
						<span class="value">{dashboardData && dashboardData.system_health ? dashboardData.system_health.response_time : 'Unknown'}</span>
					</div>
				</div>
				<div class="health-card">
					<div class="health-metric">
						<span class="label">Error Rate</span>
						<span class="value {dashboardData && dashboardData.bottlenecks && dashboardData.bottlenecks.high_error_rate ? 'text-red-500' : ''}">{dashboardData && dashboardData.system_health ? dashboardData.system_health.error_rate : 'Unknown'}</span>
					</div>
				</div>
				<div class="health-card">
					<div class="health-metric">
						<span class="label">Active Sessions</span>
						<span class="value">{formatNumber(dashboardData && dashboardData.system_health ? dashboardData.system_health.active_sessions : 0)}</span>
					</div>
				</div>
			</div>
		</div>

		<!-- Quick Actions -->
		<div class="actions-section">
			<h2>Quick Actions</h2>
			<div class="actions-grid">
				<a href="/admin/streaming/analytics" class="action-card">
					<div class="action-icon">🎥</div>
					<h3>Streaming Analytics</h3>
					<p>Video performance, engagement, and user metrics</p>
				</a>
				<a href="/admin/articles/analytics" class="action-card">
					<div class="action-icon">📝</div>
					<h3>Articles Analytics</h3>
					<p>Content performance and reader engagement</p>
				</a>
				<a href="/admin/expo/analytics" class="action-card">
					<div class="action-icon">🎪</div>
					<h3>Expo Analytics</h3>
					<p>Event performance and attendee metrics</p>
				</a>
				<a href="/admin/monitoring" class="action-card">
					<div class="action-icon">🔍</div>
					<h3>System Monitoring</h3>
					<p>Detailed system health and performance</p>
				</a>
				<a href="/admin/analytics/export" class="action-card">
					<div class="action-icon">📊</div>
					<h3>Export Data</h3>
					<p>Download reports and analytics data</p>
				</a>
				<a href="/admin/analytics/webhooks" class="action-card">
					<div class="action-icon">🔗</div>
					<h3>Webhook Analytics</h3>
					<p>Monitor webhook performance</p>
				</a>
			</div>
		</div>
	{/if}
</div>

<style>
	.dashboard {
		padding: var(--space-xl);
		max-width: 1200px;
		margin: 0 auto;
	}

	.header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: var(--space-xl);
		gap: var(--space-lg);
	}

	.header-content {
		flex-grow: 1;
	}

	.header h1 {
		font-size: var(--text-2xl);
		font-weight: 700;
		color: var(--text-primary);
		margin: 0;
	}

	.subtitle {
		color: var(--text-secondary);
		margin: var(--space-xs) 0 0 0;
	}

	.header-actions {
		display: flex;
		gap: var(--space-sm);
		flex-shrink: 0;
	}

	.last-updated {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.refresh-btn {
		padding: var(--space-sm) var(--space-lg);
		background: var(--primary);
		color: white;
		border: none;
		border-radius: var(--radius-lg);
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.refresh-btn:hover:not(:disabled) {
		background: var(--primary-dark);
		transform: translateY(-1px);
	}

	.refresh-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.retry-btn {
		padding: var(--space-sm) var(--space-lg);
		background: var(--primary);
		color: white;
		border: none;
		border-radius: var(--radius-lg);
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.retry-btn:hover:not(:disabled) {
		background: var(--primary-dark);
		transform: translateY(-1px);
	}

	.retry-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.loading-container, .error-container {
		text-align: center;
		padding: var(--space-xxl);
	}

	.error-icon {
		font-size: var(--text-4xl);
		color: var(--error);
		margin-bottom: var(--space-sm);
	}

	.error-message {
		color: var(--error);
		margin-bottom: var(--space-lg);
	}

	/* Alert Styles */
	.alert {
		background-color: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: var(--radius-lg);
		padding: var(--space-md);
		margin-bottom: var(--space-lg);
		border: 1px solid rgba(255, 255, 255, 0.1);
		display: flex;
		align-items: center;
		gap: var(--space-md);
	}

	.alert-icon {
		font-size: var(--text-2xl);
		color: var(--primary);
	}

	.alert-content h3 {
		font-size: var(--text-md);
		font-weight: 600;
		margin: 0 0 var(--space-xs) 0;
		color: var(--text-primary);
	}

	.alert-content p {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		margin: 0;
	}

	.alert-action {
		margin-left: auto;
		padding: var(--space-sm) var(--space-lg);
		background: var(--primary);
		color: white;
		border: none;
		border-radius: var(--radius-md);
		font-weight: 600;
		font-size: var(--text-sm);
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.alert-action:hover {
		background: var(--primary-dark);
		transform: translateY(-1px);
	}

	.alert-action:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	/* Metrics Grid */
	.metrics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: var(--space-lg);
		margin-bottom: var(--space-xl);
	}

	.metric-card {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
		transition: all 0.2s ease;
		display: flex;
		flex-direction: column;
		justify-content: space-between;
	}

	.metric-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
	}

	.metric-card.primary {
		background: linear-gradient(135deg, var(--primary), var(--primary-light));
		color: white;
	}

	.metric-card.primary .metric-header h3 {
		color: white;
	}

	.metric-card.primary .metric-icon {
		color: white;
	}

	.metric-card.revenue {
		background: linear-gradient(135deg, var(--success), var(--success-light));
		color: white;
	}

	.metric-card.revenue .metric-header h3 {
		color: white;
	}

	.metric-card.revenue .metric-icon {
		color: white;
	}

	.metric-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-md);
	}

	.metric-header h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		margin: 0;
		color: var(--text-secondary);
	}

	.metric-icon {
		font-size: var(--text-2xl);
		color: var(--primary);
	}

	.metric-value {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: var(--space-sm);
	}

	.metric-details {
		display: flex;
		justify-content: space-between;
		align-items: center;
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.metric-change {
		font-weight: 600;
	}

	.metric-change.positive {
		color: var(--success);
	}

	.metric-change.negative {
		color: var(--error);
	}

	.metric-growth {
		color: var(--success);
		font-weight: 600;
	}

	.metric-mrr {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		margin-top: var(--space-xs);
	}

	/* Health Section */
	.health-section {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		margin-bottom: var(--space-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-lg);
	}

	.section-header h2 {
		font-size: var(--text-xl);
		font-weight: 600;
		margin: 0;
		color: var(--text-primary);
	}

	.status-badge {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		padding: var(--space-sm) var(--space-md);
		border-radius: var(--radius-lg);
		font-weight: 600;
		font-size: var(--text-sm);
	}

	.health-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		gap: var(--space-lg);
	}

	.health-card {
		text-align: center;
	}

	.health-metric {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.health-metric .label {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.health-metric .value {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
	}

	/* Actions Section */
	.actions-section {
		margin-top: var(--space-xl);
	}

	.actions-section h2 {
		font-size: var(--text-xl);
		font-weight: 600;
		margin: 0 0 var(--space-lg) 0;
		color: var(--text-primary);
	}

	.actions-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: var(--space-lg);
	}

	.action-card {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
		text-decoration: none;
		color: inherit;
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
		transition: all 0.2s ease;
	}

	.action-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
		border-color: var(--primary);
	}

	.action-icon {
		font-size: var(--text-3xl);
		color: var(--primary);
		margin-bottom: var(--space-sm);
	}

	.action-card h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		margin: 0 0 var(--space-sm) 0;
		color: var(--text-primary);
	}

	.action-card p {
		color: var(--text-secondary);
		margin: 0;
		font-size: var(--text-sm);
	}

	/* Status Colors */
	.text-green-500 { color: #10b981; }
	.text-yellow-500 { color: #f59e0b; }
	.text-red-500 { color: #ef4444; }
	.text-gray-500 { color: #6b7280; }

	/* Responsive */
	@media (max-width: 768px) {
		.dashboard {
			padding: var(--space-lg);
		}

		.header {
			flex-direction: column;
			align-items: flex-start;
			gap: var(--space-sm);
		}

		.header-actions {
			width: 100%;
			justify-content: space-between;
		}

		.metrics-grid {
			grid-template-columns: 1fr;
		}

		.health-grid {
			grid-template-columns: 1fr;
		}

		.actions-grid {
			grid-template-columns: 1fr;
		}

		.alert {
			flex-direction: column;
			align-items: flex-start;
			gap: var(--space-sm);
		}

		.alert-action {
			margin-left: 0;
			width: 100%;
			text-align: center;
		}
	}

	@media (max-width: 480px) {
		.dashboard {
			padding: var(--space-md);
		}

		.header h1 {
			font-size: var(--text-xl);
		}

		.metric-value {
			font-size: var(--text-2xl);
		}

		.health-metric .value {
			font-size: var(--text-md);
		}
	}

	/* Dark mode enhancements */
	@media (prefers-color-scheme: dark) {
		.metric-card {
			background: rgba(255, 255, 255, 0.05);
			border-color: rgba(255, 255, 255, 0.1);
		}

		.health-card {
			background: rgba(255, 255, 255, 0.03);
		}

		.action-card {
			background: rgba(255, 255, 255, 0.05);
			border-color: rgba(255, 255, 255, 0.1);
		}
	}

	/* Animation enhancements */
	.metric-card {
		animation: fadeInUp 0.6s ease-out;
	}

	.health-card {
		animation: fadeInUp 0.6s ease-out 0.1s both;
	}

	.action-card {
		animation: fadeInUp 0.6s ease-out 0.2s both;
	}

	@keyframes fadeInUp {
		from {
			opacity: 0;
			transform: translateY(20px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	/* Hover effects */
	.metric-card:hover {
		transform: translateY(-4px);
		box-shadow: 0 12px 30px rgba(0, 0, 0, 0.2);
	}

	.health-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 20px rgba(0, 0, 0, 0.15);
	}

	.action-card:hover {
		transform: translateY(-4px);
		box-shadow: 0 12px 30px rgba(0, 0, 0, 0.2);
	}

	/* Loading animation */
	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 400px;
		gap: var(--space-lg);
	}

	.loading-container p {
		color: var(--text-secondary);
		font-size: var(--text-lg);
	}

	/* Error state */
	.error-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 400px;
		text-align: center;
		gap: var(--space-lg);
	}

	.error-container h3 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	/* Status colors */
	.alert.warning {
		background: linear-gradient(135deg, rgba(245, 158, 11, 0.1), rgba(245, 158, 11, 0.05));
		border-color: rgba(245, 158, 11, 0.3);
	}

	.alert.warning .alert-icon {
		color: #f59e0b;
	}

	.alert.critical {
		background: linear-gradient(135deg, rgba(239, 68, 68, 0.1), rgba(239, 68, 68, 0.05));
		border-color: rgba(239, 68, 68, 0.3);
	}

	.alert.critical .alert-icon {
		color: #ef4444;
	}

	/* Status badge colors */
	.status-badge.text-green-500 {
		background: rgba(16, 185, 129, 0.1);
		color: #10b981;
		border: 1px solid rgba(16, 185, 129, 0.3);
	}

	.status-badge.text-yellow-500 {
		background: rgba(245, 158, 11, 0.1);
		color: #f59e0b;
		border: 1px solid rgba(245, 158, 11, 0.3);
	}

	.status-badge.text-red-500 {
		background: rgba(239, 68, 68, 0.1);
		color: #ef4444;
		border: 1px solid rgba(239, 68, 68, 0.3);
	}

	.status-badge.text-gray-500 {
		background: rgba(107, 114, 128, 0.1);
		color: #6b7280;
		border: 1px solid rgba(107, 114, 128, 0.3);
	}
</style> 