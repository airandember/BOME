<script lang="ts">
	import { onMount } from 'svelte';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	interface Activity {
		type: 'user_signup' | 'video_upload' | 'subscription' | 'payment' | 'comment';
		message: string;
		time: string;
	}

	let isLoading = true;
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

	onMount(async () => {
		try {
			// Simulate loading dashboard data
			await new Promise(resolve => setTimeout(resolve, 1000));
			
			// Mock data for now - will be replaced with real API calls
			stats = {
				totalUsers: 1247,
				totalVideos: 342,
				totalRevenue: 45678,
				activeSubscriptions: 892,
				recentSignups: 23,
				videoViews: 15678
			};

			recentActivity = [
				{ type: 'user_signup', message: 'New user registered: john.doe@example.com', time: '2 minutes ago' },
				{ type: 'video_upload', message: 'New video uploaded: "Ancient Civilizations"', time: '15 minutes ago' },
				{ type: 'subscription', message: 'New premium subscription: jane.smith@example.com', time: '1 hour ago' },
				{ type: 'payment', message: 'Payment processed: $29.99', time: '2 hours ago' },
				{ type: 'comment', message: 'New comment on video: "Great content!"', time: '3 hours ago' }
			];

			isLoading = false;
		} catch (error) {
			showToast('Failed to load dashboard data', 'error');
			isLoading = false;
		}
	});

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

{#if isLoading}
	<div class="loading-container">
		<LoadingSpinner size="large" color="primary" />
		<p>Loading dashboard...</p>
	</div>
{:else}
	<div class="dashboard">
		<!-- Welcome Section -->
		<div class="welcome-section glass">
			<div class="welcome-content">
				<h2>Welcome back, Administrator!</h2>
				<p>Here's what's happening with your platform today.</p>
			</div>
			<div class="welcome-actions">
				<button class="btn btn-primary">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M14.5 4h-5L7 7H4a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2h-3l-2.5-3z"></path>
						<circle cx="12" cy="13" r="3"></circle>
					</svg>
					Upload Video
				</button>
				<button class="btn btn-outline">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
						<circle cx="9" cy="7" r="4"></circle>
						<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
						<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
					</svg>
					Manage Users
				</button>
			</div>
		</div>

		<!-- Stats Grid -->
		<div class="stats-grid grid grid-4">
			<div class="stat-card card glass">
				<div class="stat-header">
					<div class="stat-icon users">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
							<circle cx="9" cy="7" r="4"></circle>
							<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
							<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
						</svg>
					</div>
					<div class="stat-trend positive">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<line x1="12" y1="19" x2="12" y2="5"></line>
							<polyline points="5,12 12,5 19,12"></polyline>
						</svg>
						+12%
					</div>
				</div>
				<div class="stat-content">
					<div class="stat-value">{formatNumber(stats.totalUsers)}</div>
					<div class="stat-label">Total Users</div>
				</div>
			</div>

			<div class="stat-card card glass">
				<div class="stat-header">
					<div class="stat-icon videos">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<polygon points="23,7 16,12 23,17 23,7"></polygon>
							<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
						</svg>
					</div>
					<div class="stat-trend positive">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<line x1="12" y1="19" x2="12" y2="5"></line>
							<polyline points="5,12 12,5 19,12"></polyline>
						</svg>
						+8%
					</div>
				</div>
				<div class="stat-content">
					<div class="stat-value">{formatNumber(stats.totalVideos)}</div>
					<div class="stat-label">Total Videos</div>
				</div>
			</div>

			<div class="stat-card card glass">
				<div class="stat-header">
					<div class="stat-icon revenue">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<line x1="12" y1="1" x2="12" y2="23"></line>
							<path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"></path>
						</svg>
					</div>
					<div class="stat-trend positive">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<line x1="12" y1="19" x2="12" y2="5"></line>
							<polyline points="5,12 12,5 19,12"></polyline>
						</svg>
						+15%
					</div>
				</div>
				<div class="stat-content">
					<div class="stat-value">{formatCurrency(stats.totalRevenue)}</div>
					<div class="stat-label">Total Revenue</div>
				</div>
			</div>

			<div class="stat-card card glass">
				<div class="stat-header">
					<div class="stat-icon subscriptions">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M20 7H4a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2z"></path>
							<polyline points="16,21 12,17 8,21"></polyline>
							<polyline points="12,17 12,3"></polyline>
						</svg>
					</div>
					<div class="stat-trend positive">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<line x1="12" y1="19" x2="12" y2="5"></line>
							<polyline points="5,12 12,5 19,12"></polyline>
						</svg>
						+5%
					</div>
				</div>
				<div class="stat-content">
					<div class="stat-value">{formatNumber(stats.activeSubscriptions)}</div>
					<div class="stat-label">Active Subscriptions</div>
				</div>
			</div>
		</div>

		<!-- Content Grid -->
		<div class="content-grid grid grid-2">
			<!-- Recent Activity -->
			<div class="content-card card glass">
				<div class="card-header">
					<h3>Recent Activity</h3>
					<button class="btn btn-ghost btn-small">View All</button>
				</div>
				<div class="activity-list">
					{#each recentActivity as activity}
						<div class="activity-item">
							<div class="activity-icon" class:user_signup={activity.type === 'user_signup'} class:video_upload={activity.type === 'video_upload'} class:subscription={activity.type === 'subscription'} class:payment={activity.type === 'payment'} class:comment={activity.type === 'comment'}>
								{@html getActivityIcon(activity.type)}
							</div>
							<div class="activity-content">
								<div class="activity-message">{activity.message}</div>
								<div class="activity-time">{activity.time}</div>
							</div>
						</div>
					{/each}
				</div>
			</div>

			<!-- System Health -->
			<div class="content-card card glass">
				<div class="card-header">
					<h3>System Health</h3>
					<div class="status-badge" class:healthy={systemHealth.status === 'healthy'} class:warning={systemHealth.status === 'warning'} class:error={systemHealth.status === 'error'}>
						{systemHealth.status}
					</div>
				</div>
				<div class="health-metrics">
					<div class="health-item">
						<div class="health-label">Uptime</div>
						<div class="health-value">{systemHealth.uptime}</div>
					</div>
					<div class="health-item">
						<div class="health-label">Last Backup</div>
						<div class="health-value">{systemHealth.lastBackup}</div>
					</div>
					<div class="health-item">
						<div class="health-label">Storage Used</div>
						<div class="health-value">{systemHealth.storageUsed}</div>
					</div>
				</div>
				<div class="health-actions">
					<button class="btn btn-outline btn-small">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
							<polyline points="7,10 12,15 17,10"></polyline>
							<line x1="12" y1="15" x2="12" y2="3"></line>
						</svg>
						Backup Now
					</button>
					<button class="btn btn-ghost btn-small">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<circle cx="12" cy="12" r="3"></circle>
							<path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1 1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
						</svg>
						Settings
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}

<style>
	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 400px;
		gap: var(--space-lg);
	}

	.loading-container p {
		font-size: var(--text-lg);
		color: var(--text-secondary);
	}

	.dashboard {
		display: flex;
		flex-direction: column;
		gap: var(--space-2xl);
	}

	/* Welcome Section */
	.welcome-section {
		padding: var(--space-2xl);
		display: flex;
		align-items: center;
		justify-content: space-between;
		border-radius: var(--radius-xl);
	}

	.welcome-content h2 {
		font-size: var(--text-3xl);
		margin-bottom: var(--space-sm);
		color: var(--text-primary);
	}

	.welcome-content p {
		font-size: var(--text-lg);
		color: var(--text-secondary);
		margin: 0;
	}

	.welcome-actions {
		display: flex;
		gap: var(--space-md);
	}

	/* Stats Grid */
	.stats-grid {
		gap: var(--space-lg);
	}

	.stat-card {
		padding: var(--space-xl);
		transition: all var(--transition-normal);
	}

	.stat-card:hover {
		transform: translateY(-4px);
	}

	.stat-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: var(--space-lg);
	}

	.stat-icon {
		width: 48px;
		height: 48px;
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: var(--shadow-md);
	}

	.stat-icon.users {
		background: var(--primary-gradient);
	}

	.stat-icon.videos {
		background: var(--secondary-gradient);
	}

	.stat-icon.revenue {
		background: var(--accent-gradient);
	}

	.stat-icon.subscriptions {
		background: var(--dark-gradient);
	}

	.stat-icon svg {
		width: 24px;
		height: 24px;
		color: var(--white);
	}

	.stat-trend {
		display: flex;
		align-items: center;
		gap: var(--space-xs);
		font-size: var(--text-sm);
		font-weight: 600;
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-full);
	}

	.stat-trend.positive {
		background: rgba(0, 212, 170, 0.1);
		color: var(--success);
	}

	.stat-trend.negative {
		background: rgba(255, 107, 107, 0.1);
		color: var(--error);
	}

	.stat-trend svg {
		width: 12px;
		height: 12px;
	}

	.stat-content {
		text-align: left;
	}

	.stat-value {
		font-size: var(--text-4xl);
		font-weight: 800;
		color: var(--text-primary);
		line-height: 1;
		margin-bottom: var(--space-xs);
	}

	.stat-label {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		font-weight: 500;
	}

	/* Content Grid */
	.content-grid {
		gap: var(--space-lg);
	}

	.content-card {
		padding: var(--space-xl);
	}

	.card-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: var(--space-xl);
	}

	.card-header h3 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.status-badge {
		padding: var(--space-xs) var(--space-md);
		border-radius: var(--radius-full);
		font-size: var(--text-xs);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.status-badge.healthy {
		background: rgba(0, 212, 170, 0.1);
		color: var(--success);
	}

	.status-badge.warning {
		background: rgba(255, 167, 38, 0.1);
		color: var(--warning);
	}

	.status-badge.error {
		background: rgba(255, 107, 107, 0.1);
		color: var(--error);
	}

	/* Activity List */
	.activity-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.activity-item {
		display: flex;
		align-items: flex-start;
		gap: var(--space-md);
		padding: var(--space-md);
		border-radius: var(--radius-lg);
		transition: all var(--transition-normal);
	}

	.activity-item:hover {
		background: var(--bg-glass);
	}

	.activity-icon {
		width: 32px;
		height: 32px;
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		background: var(--bg-glass);
	}

	.activity-icon.user_signup {
		background: rgba(102, 126, 234, 0.1);
		color: var(--primary);
	}

	.activity-icon.video_upload {
		background: rgba(240, 147, 251, 0.1);
		color: var(--secondary);
	}

	.activity-icon.subscription {
		background: rgba(79, 172, 254, 0.1);
		color: var(--accent);
	}

	.activity-icon.payment {
		background: rgba(0, 212, 170, 0.1);
		color: var(--success);
	}

	.activity-icon.comment {
		background: rgba(255, 167, 38, 0.1);
		color: var(--warning);
	}

	.activity-icon svg {
		width: 16px;
		height: 16px;
	}

	.activity-content {
		flex: 1;
	}

	.activity-message {
		font-size: var(--text-sm);
		color: var(--text-primary);
		margin-bottom: var(--space-xs);
		line-height: 1.4;
	}

	.activity-time {
		font-size: var(--text-xs);
		color: var(--text-secondary);
	}

	/* Health Metrics */
	.health-metrics {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
		margin-bottom: var(--space-xl);
	}

	.health-item {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: var(--space-md);
		background: var(--bg-glass);
		border-radius: var(--radius-lg);
	}

	.health-label {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		font-weight: 500;
	}

	.health-value {
		font-size: var(--text-sm);
		color: var(--text-primary);
		font-weight: 600;
	}

	.health-actions {
		display: flex;
		gap: var(--space-md);
	}

	/* Responsive */
	@media (max-width: 1024px) {
		.stats-grid {
			grid-template-columns: repeat(2, 1fr);
		}

		.content-grid {
			grid-template-columns: 1fr;
		}

		.welcome-section {
			flex-direction: column;
			gap: var(--space-lg);
			text-align: center;
		}
	}

	@media (max-width: 768px) {
		.stats-grid {
			grid-template-columns: 1fr;
		}

		.welcome-actions {
			flex-direction: column;
			width: 100%;
		}

		.health-actions {
			flex-direction: column;
		}
	}

	@media (max-width: 480px) {
		.stat-card {
			padding: var(--space-lg);
		}

		.content-card {
			padding: var(--space-lg);
		}

		.stat-value {
			font-size: var(--text-3xl);
		}
	}
</style> 