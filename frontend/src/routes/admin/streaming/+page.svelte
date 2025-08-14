<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import { masterVideoService, type MasterVideo } from '$lib/master-video';
	import { stripeFinancialService, type StripeCustomer, type FinancialMetrics, type StripePayment } from '$lib/stripe-financial';
	import { api } from '$lib/api';
	import { StreamingSubscriberService, type Subscriber } from '$lib/services/streaming-subscribers';

	// Tab management
	let activeTab = 'overview'; // 'overview', 'videos', 'subscriptions', 'subscribers', 'financial'

	// Dashboard data
	let dashboardData: any = null;
	let isLoading = true;
	let error = '';

	// Metrics
	let metrics = {
		activeSubscriptions: 0,
		monthlyRevenue: 0,
		churnRate: 0,
		newSubscriptions: 0,
		totalCustomers: 0,
		avgRevenuePerUser: 0
	};

	// Recent events
	let recentEvents: any[] = [];

	// Load dashboard data
	async function loadDashboardData() {
		try {
			isLoading = true;
			error = '';

			// Fetch dashboard data from API
			const response = await api.get('/admin/streaming/dashboard');
			if (response.data) {
				dashboardData = response.data;
				
				// Extract metrics
				if (dashboardData?.dashboard?.metrics) {
					const data = dashboardData.dashboard.metrics;
					metrics = {
						activeSubscriptions: data.active_subscriptions || 0,
						monthlyRevenue: data.revenue_30_days?.total || 0,
						churnRate: data.churn_rate?.rate || 0,
						newSubscriptions: data.new_subscriptions || 0,
						totalCustomers: data.total_customers || 0,
						avgRevenuePerUser: data.avg_revenue_per_user || 0
					};
				}

				// Extract recent events
				recentEvents = dashboardData?.dashboard?.recent_events || [];
			} else {
				throw new Error(response.error || 'Failed to load dashboard data');
			}

			// Fallback or enhancement: derive metrics from subscribers if missing or zeroed
			const needsDerivedMetrics = !metrics || (
				(metrics.activeSubscriptions === 0 && metrics.monthlyRevenue === 0 && metrics.churnRate === 0 && metrics.newSubscriptions === 0)
			);
			if (needsDerivedMetrics) {
				await deriveMetricsFromSubscribers();
			}
		} catch (err: unknown) {
			console.error('Error loading dashboard data:', err);
			error = err instanceof Error ? err.message : 'An unknown error occurred';
			// If API fails, still try to populate metrics from subscribers as a graceful fallback
			try { await deriveMetricsFromSubscribers(); } catch {}
		} finally {
			isLoading = false;
		}
	}

	// Compute metrics directly from subscribers as a fallback
	async function deriveMetricsFromSubscribers() {
		const resp = await StreamingSubscriberService.getSubscribers({ limit: 1000 });
		const subs: Subscriber[] = resp.subscribers || [];

		// Active subs: treat as active when current date is within subscription period
		const now = new Date();
		const isActive = (s: Subscriber) => {
			// Prefer explicit period fields if present
			const start = s.current_period_start ? new Date(s.current_period_start) : StreamingSubscriberService.calculateSubscriptionStartDate(s);
			const end = s.current_period_end ? new Date(s.current_period_end) : StreamingSubscriberService.calculateSubscriptionEndDate(s);
			if (start && start > now) return false; // upcoming period
			if (end && end > now) return true;     // within period
			// Fallback to status-based when dates are missing
			return s.subscription_status === 'active' || s.subscription_status === 'trialing';
		};
		const activeSubs = subs.filter(isActive);
		const activeSubscriptions = activeSubs.length;

		// Monthly revenue (projected): normalize any interval to monthly
		function toMonthly(price: number | undefined, interval?: string, count?: number): number {
			const p = price || 0;
			const c = count && count > 0 ? count : 1;
			const unit = (interval || 'month').toLowerCase();
			switch (unit) {
				case 'year': return (p / 12) * c;
				case 'week': return (p * 4.345) * c; // avg weeks/month
				case 'day': return (p * 30) * c; // approx days/month
				case 'month': default: return p * c;
			}
		}
		const monthlyRevenue = activeSubs.reduce((sum, s) => sum + toMonthly(s.plan_price, s.plan_interval, s.plan_interval_count), 0);

		// New subs this month: created_at within current month
		const monthStart = new Date(now.getFullYear(), now.getMonth(), 1);
		const newSubscriptions = subs.filter(s => {
			if (!s.created_at) return false;
			const d = new Date(s.created_at);
			return d >= monthStart && d <= now;
		}).length;

		// Churn: cancellations in last 30 days over previous active base
		const thirtyDaysAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
		const canceledThisPeriod = subs.filter(s => s.subscription_status === 'canceled' && s.updated_at && new Date(s.updated_at) >= thirtyDaysAgo).length;
		const previousActiveBase = activeSubscriptions + canceledThisPeriod;
		const churnRate = previousActiveBase > 0 ? (canceledThisPeriod / previousActiveBase) * 100 : 0;

		metrics = {
			activeSubscriptions,
			monthlyRevenue: Math.round(monthlyRevenue * 100) / 100,
			churnRate: Math.round(churnRate * 10) / 10,
			newSubscriptions,
			totalCustomers: subs.length,
			avgRevenuePerUser: activeSubscriptions > 0 ? Math.round((monthlyRevenue / activeSubscriptions) * 100) / 100 : 0
		};
	}

	onMount(() => {
		loadDashboardData();
	});
</script>

<svelte:head>
	<title>Streaming Admin Dashboard - BOME</title>
	<meta name="description" content="Manage streaming content, subscribers, and financial data" />
</svelte:head>

{#if isLoading}
	<div class="loading-container">
		<LoadingSpinner />
	</div>
<!--{:else if error}
	<div class="error-container">
		<div class="error-content">
			<h2>Error Loading Dashboard</h2>
			<p>{error}</p>
			<button class="btn btn-primary" on:click={loadDashboardData}>Retry</button>
		</div>
	</div>-->
{:else}
	<div class="dashboard">
		<!-- Overview Stats Grid -->
		<div class="stats-grid">
			<div class="stat-card">
				<div class="stat-header">
					<h3>Active Subscriptions</h3>
					<svg class="stat-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
						<circle cx="9" cy="7" r="4"></circle>
						<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
						<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
					</svg>
				</div>
				<div class="stat-value">{metrics.activeSubscriptions.toLocaleString()}</div>
				<div class="stat-change positive">+12% from last month</div>
			</div>

			<div class="stat-card">
				<div class="stat-header">
					<h3>Projected Monthly Revenue</h3>
					<svg class="stat-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M12 1v22M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"></path>
					</svg>
				</div>
				<div class="stat-value">${metrics.monthlyRevenue.toLocaleString()}</div>
				<div class="stat-change positive">+8% from last month</div>
			</div>

			<div class="stat-card">
				<div class="stat-header">
					<h3>Churn Rate</h3>
					<svg class="stat-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M18 20V10"></path>
						<path d="M12 20V4"></path>
						<path d="M6 20v-6"></path>
					</svg>
				</div>
				<div class="stat-value">{metrics.churnRate.toFixed(1)}%</div>
				<div class="stat-change negative">+0.5% from last month</div>
			</div>

			<div class="stat-card">
				<div class="stat-header">
					<h3>New Subscriptions</h3>
					<svg class="stat-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
						<circle cx="8.5" cy="7" r="4"></circle>
						<line x1="20" y1="8" x2="20" y2="14"></line>
						<line x1="23" y1="11" x2="17" y2="11"></line>
					</svg>
				</div>
				<div class="stat-value">{metrics.newSubscriptions.toLocaleString()}</div>
				<div class="stat-change positive">+15% from last month</div>
			</div>
		</div>

		<!-- Quick Navigation -->
		<div class="quick-nav-section">
			<h2 class="section-title">Quick Navigation</h2>
			<div class="quick-nav-grid">
				<a href="/admin/streaming/videos" class="quick-nav-card">
					<svg class="quick-nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<polygon points="23,7 16,12 23,17 23,7"></polygon>
						<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
					</svg>
					<h3>Video Management</h3>
					<p>Upload and manage video content</p>
				</a>

				<a href="/admin/streaming/subscriptions" class="quick-nav-card">
					<svg class="quick-nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<rect x="1" y="4" width="22" height="16" rx="2" ry="2"></rect>
						<line x1="1" y1="10" x2="23" y2="10"></line>
					</svg>
					<h3>Manage Subscriptions</h3>
					<p>Create and manage subscription plans</p>
				</a>

				<a href="/admin/streaming/customers" class="quick-nav-card">
					<svg class="quick-nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
						<circle cx="9" cy="7" r="4"></circle>
						<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
						<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
					</svg>
					<h3>Customer Management</h3>
					<p>View and manage customer accounts</p>
				</a>

				<a href="/admin/streaming/analytics" class="quick-nav-card">
					<svg class="quick-nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M18 20V10"></path>
						<path d="M12 20V4"></path>
						<path d="M6 20v-6"></path>
					</svg>
					<h3>Analytics & Reports</h3>
					<p>Detailed revenue and subscription analytics</p>
				</a>

				<a href="/admin/streaming/promotions" class="quick-nav-card">
					<svg class="quick-nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"></path>
						<line x1="7" y1="7" x2="7.01" y2="7"></line>
					</svg>
					<h3>Promotions</h3>
					<p>Manage deals and promotional offers</p>
				</a>

				<a href="/admin/streaming/events" class="quick-nav-card">
					<svg class="quick-nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
						<line x1="16" y1="2" x2="16" y2="6"></line>
						<line x1="8" y1="2" x2="8" y2="6"></line>
						<line x1="3" y1="10" x2="21" y2="10"></line>
					</svg>
					<h3>Events</h3>
					<p>Event-based subscription deals</p>
				</a>

				<a href="/admin/streaming/settings" class="quick-nav-card">
					<svg class="quick-nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="12" cy="12" r="3"></circle>
						<path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1.51 1 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
					</svg>
					<h3>Settings</h3>
					<p>Streaming configuration and preferences</p>
				</a>
			</div>
		</div>

		<!-- Recent Activity -->
		<div class="recent-activity-section">
			<h2 class="section-title">Recent Activity</h2>
			<div class="activity-list">
				{#if recentEvents.length > 0}
					{#each recentEvents as event}
						<div class="activity-item">
							<div class="activity-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<circle cx="12" cy="12" r="10"></circle>
									<polyline points="12,6 12,12 16,14"></polyline>
								</svg>
							</div>
							<div class="activity-content">
								<p class="activity-text">{event.description || 'Activity occurred'}</p>
								<p class="activity-time">{event.timestamp || 'Just now'}</p>
							</div>
						</div>
					{/each}
				{:else}
					<div class="no-activity">
						<p>No recent activity to display</p>
					</div>
				{/if}
			</div>
		</div>
	</div>
{/if}

<style>
	/* Loading and Error States */
	.loading-container {
		display: flex;
		justify-content: center;
		align-items: center;
		min-height: 400px;
	}

	.error-container {
		display: flex;
		justify-content: center;
		align-items: center;
		min-height: 400px;
	}

	.error-content {
		text-align: center;
		padding: var(--space-xl);
	}

	.error-content h2 {
		color: var(--error);
		margin-bottom: var(--space-md);
	}

	/* Dashboard Layout */
	.dashboard {
		display: flex;
		flex-direction: column;
		gap: var(--space-xl);
	}

	/* Stats Grid */
	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: var(--space-lg);
	}

	.stat-card {
		background: var(--bg-glass, rgba(255, 255, 255, 0.1));
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: var(--space-lg);
		border: 1px solid rgba(255, 255, 255, 0.1);
		transition: all 0.3s ease;
	}

	.stat-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
	}

	.stat-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-md);
	}

	.stat-header h3 {
		font-size: var(--text-sm);
		font-weight: 500;
		color: var(--text-secondary);
		margin: 0;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-weight: 600;
	}

	.stat-icon {
		width: 24px;
		height: 24px;
		color: var(--primary-color, #3b82f6);
	}

	.stat-value {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: var(--space-sm);
	}

	.stat-change {
		font-size: 0.8rem;
		font-weight: 600;
		color: #10b981;
	}

	.stat-change.positive {
		color: #10b981;
	}

	.stat-change.negative {
		color: #ef4444;
	}

	/* Quick Navigation */
	.quick-nav-section {
		margin-top: var(--space-xl);
	}

	.section-title {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-lg);
	}

	.quick-nav-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: var(--space-lg);
	}

	.quick-nav-card {
		background: var(--bg-glass, rgba(255, 255, 255, 0.1));
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: var(--space-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
		text-decoration: none;
		color: inherit;
		transition: all 0.3s ease;
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
	}

	.quick-nav-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
		background: var(--bg-glass-dark, rgba(255, 255, 255, 0.15));
	}

	.quick-nav-icon {
		width: 48px;
		height: 48px;
		color: var(--primary-color, #3b82f6);
		margin-bottom: var(--space-md);
	}

	.quick-nav-card h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
	}

	.quick-nav-card p {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		margin: 0;
	}

	/* Recent Activity */
	.recent-activity-section {
		margin-top: var(--space-xl);
	}

	.activity-list {
		background: var(--bg-glass, rgba(255, 255, 255, 0.1));
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: var(--space-lg);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.activity-item {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		padding: var(--space-md) 0;
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
		transition: all 0.3s ease;
	}

	.activity-item:hover {
		background: rgba(255, 255, 255, 0.05);
		border-radius: 8px;
		padding: var(--space-md);
		margin: 0 calc(-1 * var(--space-md));
	}

	.activity-item:last-child {
		border-bottom: none;
	}

	.activity-icon {
		width: 40px;
		height: 40px;
		background: var(--primary-gradient, linear-gradient(135deg, #3b82f6, #1d4ed8));
		border-radius: 10px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.activity-icon svg {
		width: 20px;
		height: 20px;
		color: white;
	}

	.activity-content {
		flex: 1;
	}

	.activity-text {
		font-size: var(--text-sm);
		color: var(--text-primary);
		margin: 0 0 var(--space-xs) 0;
		font-weight: 500;
	}

	.activity-time {
		font-size: var(--text-xs);
		color: var(--text-secondary);
		margin: 0;
	}

	.no-activity {
		text-align: center;
		padding: var(--space-xl);
		color: var(--text-secondary);
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.stats-grid {
			grid-template-columns: 1fr;
		}

		.quick-nav-grid {
			grid-template-columns: 1fr;
		}

		.stat-value {
			font-size: var(--text-2xl);
		}

		.quick-nav-card {
			padding: var(--space-lg);
		}

		.quick-nav-icon {
			width: 40px;
			height: 40px;
		}
	}
</style>