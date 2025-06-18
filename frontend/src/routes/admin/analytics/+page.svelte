<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { api } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	interface Analytics {
		users: {
			total: number;
			new_today: number;
			new_week: number;
			new_month: number;
			active_today: number;
			growth_rate: number;
			churn_rate: number;
			retention_rate: number;
		};
		videos: {
			total: number;
			published: number;
			pending: number;
			draft: number;
			total_views: number;
			total_likes: number;
			total_comments: number;
			total_shares: number;
			avg_rating: number;
			top_categories: Array<{
				name: string;
				count: number;
				views: number;
			}>;
		};
		subscriptions: {
			active: number;
			new_today: number;
			new_week: number;
			new_month: number;
			cancelled: number;
			revenue_today: number;
			revenue_week: number;
			revenue_month: number;
			revenue_year: number;
			mrr: number;
			arr: number;
			ltv: number;
			plans: Array<{
				name: string;
				count: number;
				revenue: number;
			}>;
		};
		engagement: {
			avg_watch_time: string;
			completion_rate: number;
			like_ratio: number;
			comment_rate: number;
			share_count: number;
			bounce_rate: number;
			session_duration: string;
			pages_per_session: number;
		};
		system: {
			uptime: string;
			response_time: string;
			error_rate: string;
			storage_used: string;
			bandwidth_used: string;
			cdn_hits: string;
			database_size: string;
			active_sessions: number;
		};
		geographic: {
			top_countries: Array<{
				country: string;
				users: number;
				percentage: number;
			}>;
			top_states: Array<{
				state: string;
				users: number;
				percentage: number;
			}>;
		};
		devices: {
			desktop: {
				users: number;
				percentage: number;
				avg_session: string;
			};
			mobile: {
				users: number;
				percentage: number;
				avg_session: string;
			};
			tablet: {
				users: number;
				percentage: number;
				avg_session: string;
			};
			browsers: Array<{
				name: string;
				users: number;
				percentage: number;
			}>;
		};
		time_series: {
			users: Array<{
				date: string;
				new_users: number;
				active_users: number;
				returning_users: number;
			}>;
			revenue: Array<{
				date: string;
				revenue: number;
				subscriptions: number;
				upgrades: number;
			}>;
			engagement: Array<{
				date: string;
				views: number;
				likes: number;
				comments: number;
				shares: number;
			}>;
		};
		conversion: {
			funnel: Array<{
				stage: string;
				count: number;
				conversion: number;
			}>;
			cohort_analysis: Array<{
				cohort: string;
				users: number;
				retention_30d: number;
				retention_90d: number;
			}>;
		};
	}

	interface RealTimeData {
		active_users: number;
		current_streams: number;
		server_load: number;
		bandwidth_usage: string;
		recent_signups: number;
		recent_subscriptions: number;
		error_rate: number;
		response_time: number;
		live_events: Array<{
			time: string;
			event: string;
			details: string;
		}>;
		top_content_now: Array<{
			title: string;
			viewers: number;
		}>;
	}

	let analytics: Analytics | null = null;
	let realTimeData: RealTimeData | null = null;
	let loading = true;
	let selectedPeriod = '7d';
	let selectedView = 'overview';
	let realTimeInterval: number;

	// Chart data
	let chartData: any = null;

	onMount(() => {
		loadAnalytics();
		loadRealTimeData();
		
		// Update real-time data every 30 seconds
		realTimeInterval = setInterval(loadRealTimeData, 30000);
	});

	onDestroy(() => {
		if (realTimeInterval) {
			clearInterval(realTimeInterval);
		}
	});

	async function loadAnalytics() {
		try {
			loading = true;
			const response = await api.get(`/api/v1/admin/analytics/overview?period=${selectedPeriod}`);
			analytics = response.analytics;
			
			// Process data for charts
			if (analytics?.time_series) {
				chartData = {
					userGrowth: analytics.time_series.users,
					revenueGrowth: analytics.time_series.revenue,
					engagement: analytics.time_series.engagement
				};
			}
		} catch (error) {
			showToast('Failed to load analytics', 'error');
			console.error('Error loading analytics:', error);
		} finally {
			loading = false;
		}
	}

	async function loadRealTimeData() {
		try {
			const response = await api.get('/api/v1/admin/analytics/realtime');
			realTimeData = response.real_time;
		} catch (error) {
			console.error('Error loading real-time data:', error);
		}
	}

	async function exportData(format: 'csv' | 'json') {
		try {
			const response = await fetch(`/api/v1/admin/analytics/export?format=${format}&period=${selectedPeriod}`, {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});
			
			if (response.ok) {
				const blob = await response.blob();
				const url = window.URL.createObjectURL(blob);
				const a = document.createElement('a');
				a.href = url;
				a.download = `analytics_export.${format}`;
				document.body.appendChild(a);
				a.click();
				window.URL.revokeObjectURL(url);
				document.body.removeChild(a);
				showToast(`Analytics exported as ${format.toUpperCase()}`, 'success');
			} else {
				showToast('Failed to export analytics', 'error');
			}
		} catch (error) {
			showToast('Failed to export analytics', 'error');
			console.error('Error exporting analytics:', error);
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

	function formatPercentage(num: number) {
		return `${(num * 100).toFixed(1)}%`;
	}

	function getGrowthIndicator(current: number, previous: number) {
		if (previous === 0) return { type: 'neutral', value: '0%' };
		const growth = ((current - previous) / previous) * 100;
		return {
			type: growth > 0 ? 'positive' : growth < 0 ? 'negative' : 'neutral',
			value: `${growth > 0 ? '+' : ''}${growth.toFixed(1)}%`
		};
	}

	function formatTime(timeString: string) {
		return new Date(timeString).toLocaleTimeString();
	}
</script>

<svelte:head>
	<title>Analytics - Admin Dashboard</title>
</svelte:head>

<div class="analytics-page">
	<div class="page-header">
		<div class="header-content">
			<div class="header-text">
				<h1>Analytics Dashboard</h1>
				<p>Monitor your platform's performance and user engagement</p>
			</div>
			
			<div class="header-controls">
				<div class="view-selector">
					<button 
						class="view-btn" 
						class:active={selectedView === 'overview'}
						on:click={() => selectedView = 'overview'}
					>
						Overview
					</button>
					<button 
						class="view-btn" 
						class:active={selectedView === 'realtime'}
						on:click={() => selectedView = 'realtime'}
					>
						Real-time
					</button>
					<button 
						class="view-btn" 
						class:active={selectedView === 'detailed'}
						on:click={() => selectedView = 'detailed'}
					>
						Detailed
					</button>
				</div>
				
				<div class="period-selector">
					<select bind:value={selectedPeriod} on:change={loadAnalytics} class="period-select">
						<option value="24h">Last 24 Hours</option>
						<option value="7d">Last 7 Days</option>
						<option value="30d">Last 30 Days</option>
						<option value="90d">Last 90 Days</option>
					</select>
				</div>
				
				<div class="export-controls">
					<button class="export-btn" on:click={() => exportData('csv')}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
							<polyline points="7,10 12,15 17,10"></polyline>
							<line x1="12" y1="15" x2="12" y2="3"></line>
						</svg>
						CSV
					</button>
					<button class="export-btn" on:click={() => exportData('json')}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
							<polyline points="7,10 12,15 17,10"></polyline>
							<line x1="12" y1="15" x2="12" y2="3"></line>
						</svg>
						JSON
					</button>
				</div>
			</div>
		</div>
	</div>

	{#if loading}
		<div class="loading-container">
			<LoadingSpinner size="large" color="primary" />
			<p>Loading analytics...</p>
		</div>
	{:else if selectedView === 'realtime' && realTimeData}
		<!-- Real-time Analytics View -->
		<div class="realtime-dashboard">
			<div class="realtime-header">
				<h2>Real-time Analytics</h2>
				<div class="live-indicator">
					<div class="pulse"></div>
					<span>Live</span>
				</div>
			</div>
			
			<!-- Real-time Metrics -->
			<div class="realtime-metrics">
				<div class="realtime-card glass">
					<div class="metric-icon active-users">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
							<circle cx="9" cy="7" r="4"></circle>
							<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
							<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
						</svg>
					</div>
					<div class="metric-content">
						<h3>Active Users</h3>
						<div class="metric-value">{formatNumber(realTimeData.active_users)}</div>
					</div>
				</div>
				
				<div class="realtime-card glass">
					<div class="metric-icon streaming">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<polygon points="23,7 16,12 23,17 23,7"></polygon>
							<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
						</svg>
					</div>
					<div class="metric-content">
						<h3>Current Streams</h3>
						<div class="metric-value">{formatNumber(realTimeData.current_streams)}</div>
					</div>
				</div>
				
				<div class="realtime-card glass">
					<div class="metric-icon server-load">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect>
							<line x1="8" y1="21" x2="16" y2="21"></line>
							<line x1="12" y1="17" x2="12" y2="21"></line>
						</svg>
					</div>
					<div class="metric-content">
						<h3>Server Load</h3>
						<div class="metric-value">{formatPercentage(realTimeData.server_load)}</div>
					</div>
				</div>
				
				<div class="realtime-card glass">
					<div class="metric-icon bandwidth">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-2.3M22 12.5a10 10 0 0 1-18.8 2.3"></path>
						</svg>
					</div>
					<div class="metric-content">
						<h3>Bandwidth</h3>
						<div class="metric-value">{realTimeData.bandwidth_usage}</div>
					</div>
				</div>
			</div>
			
			<!-- Live Events Feed -->
			<div class="live-events-section">
				<div class="live-events glass">
					<h3>Live Events</h3>
					<div class="events-list">
						{#each realTimeData.live_events as event}
							<div class="event-item">
								<div class="event-time">{formatTime(event.time)}</div>
								<div class="event-content">
									<div class="event-type">{event.event}</div>
									<div class="event-details">{event.details}</div>
								</div>
							</div>
						{/each}
					</div>
				</div>
				
				<div class="top-content glass">
					<h3>Top Content Right Now</h3>
					<div class="content-list">
						{#each realTimeData.top_content_now as content}
							<div class="content-item">
								<div class="content-title">{content.title}</div>
								<div class="content-viewers">{content.viewers} viewers</div>
							</div>
						{/each}
					</div>
				</div>
			</div>
		</div>
	{:else if selectedView === 'detailed' && analytics}
		<!-- Detailed Analytics View -->
		<div class="detailed-analytics">
			<h2>Detailed Analytics</h2>
			
			<!-- Geographic Analytics -->
			<div class="analytics-section">
				<div class="section-header">
					<h3>Geographic Distribution</h3>
				</div>
				<div class="geographic-grid">
					<div class="geographic-card glass">
						<h4>Top Countries</h4>
						<div class="country-list">
							{#each analytics.geographic.top_countries as country}
								<div class="country-item">
									<div class="country-info">
										<span class="country-name">{country.country}</span>
										<span class="country-percentage">{country.percentage}%</span>
									</div>
									<div class="country-bar">
										<div class="country-fill" style="width: {country.percentage}%"></div>
									</div>
									<span class="country-count">{formatNumber(country.users)} users</span>
								</div>
							{/each}
						</div>
					</div>
					
					<div class="geographic-card glass">
						<h4>Top States (US)</h4>
						<div class="state-list">
							{#each analytics.geographic.top_states as state}
								<div class="state-item">
									<div class="state-info">
										<span class="state-name">{state.state}</span>
										<span class="state-percentage">{state.percentage}%</span>
									</div>
									<div class="state-bar">
										<div class="state-fill" style="width: {state.percentage}%"></div>
									</div>
									<span class="state-count">{formatNumber(state.users)} users</span>
								</div>
							{/each}
						</div>
					</div>
				</div>
			</div>
			
			<!-- Device Analytics -->
			<div class="analytics-section">
				<div class="section-header">
					<h3>Device & Browser Analytics</h3>
				</div>
				<div class="device-grid">
					<div class="device-overview glass">
						<h4>Device Types</h4>
						<div class="device-stats">
							<div class="device-stat">
								<div class="device-icon desktop">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect>
										<line x1="8" y1="21" x2="16" y2="21"></line>
										<line x1="12" y1="17" x2="12" y2="21"></line>
									</svg>
								</div>
								<div class="device-info">
									<div class="device-name">Desktop</div>
									<div class="device-users">{formatNumber(analytics.devices.desktop.users)} users</div>
									<div class="device-percentage">{analytics.devices.desktop.percentage}%</div>
									<div class="device-session">Avg: {analytics.devices.desktop.avg_session}</div>
								</div>
							</div>
							
							<div class="device-stat">
								<div class="device-icon mobile">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<rect x="5" y="2" width="14" height="20" rx="2" ry="2"></rect>
										<line x1="12" y1="18" x2="12.01" y2="18"></line>
									</svg>
								</div>
								<div class="device-info">
									<div class="device-name">Mobile</div>
									<div class="device-users">{formatNumber(analytics.devices.mobile.users)} users</div>
									<div class="device-percentage">{analytics.devices.mobile.percentage}%</div>
									<div class="device-session">Avg: {analytics.devices.mobile.avg_session}</div>
								</div>
							</div>
							
							<div class="device-stat">
								<div class="device-icon tablet">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<rect x="4" y="2" width="16" height="20" rx="2" ry="2"></rect>
										<line x1="12" y1="18" x2="12.01" y2="18"></line>
									</svg>
								</div>
								<div class="device-info">
									<div class="device-name">Tablet</div>
									<div class="device-users">{formatNumber(analytics.devices.tablet.users)} users</div>
									<div class="device-percentage">{analytics.devices.tablet.percentage}%</div>
									<div class="device-session">Avg: {analytics.devices.tablet.avg_session}</div>
								</div>
							</div>
						</div>
					</div>
					
					<div class="browser-stats glass">
						<h4>Browser Distribution</h4>
						<div class="browser-list">
							{#each analytics.devices.browsers as browser}
								<div class="browser-item">
									<div class="browser-info">
										<span class="browser-name">{browser.name}</span>
										<span class="browser-percentage">{browser.percentage}%</span>
									</div>
									<div class="browser-bar">
										<div class="browser-fill" style="width: {browser.percentage}%"></div>
									</div>
									<span class="browser-count">{formatNumber(browser.users)} users</span>
								</div>
							{/each}
						</div>
					</div>
				</div>
			</div>
			
			<!-- Conversion Funnel -->
			<div class="analytics-section">
				<div class="section-header">
					<h3>Conversion Funnel</h3>
				</div>
				<div class="funnel-container glass">
					<div class="funnel-stages">
						{#each analytics.conversion.funnel as stage, index}
							<div class="funnel-stage" style="width: {stage.conversion}%">
								<div class="stage-info">
									<div class="stage-name">{stage.stage}</div>
									<div class="stage-count">{formatNumber(stage.count)}</div>
									<div class="stage-conversion">{stage.conversion}%</div>
								</div>
								{#if index < analytics.conversion.funnel.length - 1}
									<div class="funnel-arrow">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<polyline points="9,18 15,12 9,6"></polyline>
										</svg>
									</div>
								{/if}
							</div>
						{/each}
					</div>
				</div>
			</div>
		</div>
	{:else if analytics}
		<!-- Overview Analytics View (existing content) -->
		<div class="metrics-grid">
			<!-- User Metrics -->
			<div class="metric-card glass">
				<div class="metric-header">
					<div class="metric-icon users">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
							<circle cx="9" cy="7" r="4"></circle>
							<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
							<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
						</svg>
					</div>
					<h3>Total Users</h3>
				</div>
				<div class="metric-value">{formatNumber(analytics.users.total)}</div>
				<div class="metric-details">
					<span class="metric-sub">+{analytics.users.new_today} today</span>
					<span class="metric-trend positive">+{analytics.users.new_week} this week</span>
				</div>
			</div>

			<!-- Video Metrics -->
			<div class="metric-card glass">
				<div class="metric-header">
					<div class="metric-icon videos">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<polygon points="23,7 16,12 23,17 23,7"></polygon>
							<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
						</svg>
					</div>
					<h3>Total Videos</h3>
				</div>
				<div class="metric-value">{formatNumber(analytics.videos.total)}</div>
				<div class="metric-details">
					<span class="metric-sub">{analytics.videos.published} published</span>
					<span class="metric-sub">{analytics.videos.pending} pending</span>
				</div>
			</div>

			<!-- Revenue Metrics -->
			<div class="metric-card glass">
				<div class="metric-header">
					<div class="metric-icon revenue">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<line x1="12" y1="1" x2="12" y2="23"></line>
							<path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"></path>
						</svg>
					</div>
					<h3>Monthly Revenue</h3>
				</div>
				<div class="metric-value">{formatCurrency(analytics.subscriptions.revenue_month)}</div>
				<div class="metric-details">
					<span class="metric-sub">{formatCurrency(analytics.subscriptions.revenue_today)} today</span>
					<span class="metric-trend positive">+{analytics.subscriptions.new_month} subs</span>
				</div>
			</div>

			<!-- Engagement Metrics -->
			<div class="metric-card glass">
				<div class="metric-header">
					<div class="metric-icon engagement">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path>
						</svg>
					</div>
					<h3>Avg. Watch Time</h3>
				</div>
				<div class="metric-value">{analytics.engagement.avg_watch_time}</div>
				<div class="metric-details">
					<span class="metric-sub">{formatPercentage(analytics.engagement.completion_rate)} completion</span>
					<span class="metric-trend positive">{formatPercentage(analytics.engagement.like_ratio)} like ratio</span>
				</div>
			</div>
		</div>

		<!-- Charts and Detailed Analytics -->
		<div class="analytics-grid">
			<!-- User Activity Chart -->
			<div class="chart-card glass">
				<div class="chart-header">
					<h3>User Activity</h3>
					<div class="chart-legend">
						<span class="legend-item">
							<div class="legend-color active"></div>
							Active Users
						</span>
						<span class="legend-item">
							<div class="legend-color new"></div>
							New Users
						</span>
					</div>
				</div>
				<div class="chart-placeholder">
					<div class="chart-bars">
						{#each Array(7) as _, i}
							<div class="chart-bar">
								<div class="bar active" style="height: {Math.random() * 80 + 20}%"></div>
								<div class="bar new" style="height: {Math.random() * 40 + 10}%"></div>
							</div>
						{/each}
					</div>
					<div class="chart-labels">
						<span>Mon</span>
						<span>Tue</span>
						<span>Wed</span>
						<span>Thu</span>
						<span>Fri</span>
						<span>Sat</span>
						<span>Sun</span>
					</div>
				</div>
			</div>

			<!-- Content Performance -->
			<div class="performance-card glass">
				<div class="performance-header">
					<h3>Content Performance</h3>
				</div>
				<div class="performance-stats">
					<div class="stat-item">
						<div class="stat-label">Total Views</div>
						<div class="stat-value">{formatNumber(analytics.videos.total_views)}</div>
						<div class="stat-change positive">+12.5%</div>
					</div>
					<div class="stat-item">
						<div class="stat-label">Total Likes</div>
						<div class="stat-value">{formatNumber(analytics.videos.total_likes)}</div>
						<div class="stat-change positive">+8.3%</div>
					</div>
					<div class="stat-item">
						<div class="stat-label">Comments</div>
						<div class="stat-value">{formatNumber(analytics.videos.total_comments)}</div>
						<div class="stat-change positive">+15.7%</div>
					</div>
					<div class="stat-item">
						<div class="stat-label">Shares</div>
						<div class="stat-value">{formatNumber(analytics.engagement.share_count)}</div>
						<div class="stat-change positive">+22.1%</div>
					</div>
				</div>
			</div>

			<!-- Subscription Analytics -->
			<div class="subscription-card glass">
				<div class="subscription-header">
					<h3>Subscription Analytics</h3>
				</div>
				<div class="subscription-stats">
					<div class="subscription-metric">
						<div class="subscription-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
								<circle cx="8.5" cy="7" r="4"></circle>
								<line x1="20" y1="8" x2="20" y2="14"></line>
								<line x1="23" y1="11" x2="17" y2="11"></line>
							</svg>
						</div>
						<div class="subscription-details">
							<div class="subscription-value">{formatNumber(analytics.subscriptions.active)}</div>
							<div class="subscription-label">Active Subscriptions</div>
						</div>
					</div>
					<div class="subscription-metric">
						<div class="subscription-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<line x1="12" y1="1" x2="12" y2="23"></line>
								<path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"></path>
							</svg>
						</div>
						<div class="subscription-details">
							<div class="subscription-value">{formatCurrency(analytics.subscriptions.revenue_week)}</div>
							<div class="subscription-label">Weekly Revenue</div>
						</div>
					</div>
					<div class="subscription-metric">
						<div class="subscription-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polyline points="23,6 13.5,15.5 8.5,10.5 1,18"></polyline>
								<polyline points="17,6 23,6 23,12"></polyline>
							</svg>
						</div>
						<div class="subscription-details">
							<div class="subscription-value">{analytics.subscriptions.new_week}</div>
							<div class="subscription-label">New This Week</div>
						</div>
					</div>
				</div>
			</div>

			<!-- System Health -->
			<div class="system-card glass">
				<div class="system-header">
					<h3>System Health</h3>
					<div class="system-status healthy">
						<div class="status-indicator"></div>
						Healthy
					</div>
				</div>
				<div class="system-metrics">
					<div class="system-metric">
						<div class="system-label">Uptime</div>
						<div class="system-value">{analytics.system.uptime}</div>
					</div>
					<div class="system-metric">
						<div class="system-label">Response Time</div>
						<div class="system-value">{analytics.system.response_time}</div>
					</div>
					<div class="system-metric">
						<div class="system-label">Error Rate</div>
						<div class="system-value">{analytics.system.error_rate}</div>
					</div>
					<div class="system-metric">
						<div class="system-label">Storage Used</div>
						<div class="system-value">{analytics.system.storage_used}</div>
					</div>
				</div>
			</div>
		</div>
	{:else}
		<div class="error-state glass">
			<div class="error-icon">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="12" cy="12" r="10"></circle>
					<line x1="12" y1="8" x2="12" y2="12"></line>
					<line x1="12" y1="16" x2="12.01" y2="16"></line>
				</svg>
			</div>
			<h3>Failed to Load Analytics</h3>
			<p>There was an error loading the analytics data.</p>
			<button class="btn btn-primary" on:click={loadAnalytics}>
				Try Again
			</button>
		</div>
	{/if}
</div>

<style>
	.analytics-page {
		display: flex;
		flex-direction: column;
		gap: var(--space-2xl);
	}

	.page-header {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.header-content {
		display: flex;
		align-items: center;
		justify-content: space-between;
		flex-wrap: wrap;
		gap: var(--space-lg);
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

	.header-controls {
		display: flex;
		align-items: center;
		gap: var(--space-lg);
		flex-wrap: wrap;
	}

	.view-selector {
		display: flex;
		background: var(--bg-glass);
		border-radius: var(--radius-lg);
		padding: 4px;
		gap: 4px;
	}

	.view-btn {
		padding: var(--space-sm) var(--space-lg);
		border: none;
		border-radius: var(--radius-md);
		background: transparent;
		color: var(--text-secondary);
		font-size: var(--text-sm);
		font-weight: 500;
		cursor: pointer;
		transition: all var(--transition-fast);
	}

	.view-btn.active {
		background: var(--primary);
		color: var(--white);
		box-shadow: var(--shadow-sm);
	}

	.view-btn:hover:not(.active) {
		background: rgba(255, 255, 255, 0.1);
		color: var(--text-primary);
	}

	.export-controls {
		display: flex;
		gap: var(--space-sm);
	}

	.export-btn {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		padding: var(--space-sm) var(--space-md);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		background: var(--bg-glass);
		color: var(--text-primary);
		font-size: var(--text-sm);
		cursor: pointer;
		transition: all var(--transition-fast);
	}

	.export-btn:hover {
		background: var(--bg-hover);
		border-color: var(--primary);
	}

	.export-btn svg {
		width: 16px;
		height: 16px;
	}

	/* Real-time Dashboard Styles */
	.realtime-dashboard {
		display: flex;
		flex-direction: column;
		gap: var(--space-2xl);
	}

	.realtime-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.realtime-header h2 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.live-indicator {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		padding: var(--space-sm) var(--space-md);
		background: rgba(67, 233, 123, 0.1);
		border: 1px solid var(--success);
		border-radius: var(--radius-full);
		color: var(--success);
		font-size: var(--text-sm);
		font-weight: 600;
	}

	.pulse {
		width: 8px;
		height: 8px;
		background: var(--success);
		border-radius: 50%;
		animation: pulse 2s infinite;
	}

	@keyframes pulse {
		0% {
			transform: scale(0.95);
			box-shadow: 0 0 0 0 rgba(67, 233, 123, 0.7);
		}
		70% {
			transform: scale(1);
			box-shadow: 0 0 0 10px rgba(67, 233, 123, 0);
		}
		100% {
			transform: scale(0.95);
			box-shadow: 0 0 0 0 rgba(67, 233, 123, 0);
		}
	}

	.realtime-metrics {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: var(--space-lg);
	}

	.realtime-card {
		display: flex;
		align-items: center;
		gap: var(--space-lg);
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		transition: all var(--transition-normal);
	}

	.realtime-card:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}

	.realtime-card .metric-icon {
		width: 56px;
		height: 56px;
	}

	.realtime-card .metric-icon.active-users {
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
	}

	.realtime-card .metric-icon.streaming {
		background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
	}

	.realtime-card .metric-icon.server-load {
		background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
	}

	.realtime-card .metric-icon.bandwidth {
		background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
	}

	.metric-content h3 {
		font-size: var(--text-md);
		font-weight: 600;
		color: var(--text-secondary);
		margin: 0 0 var(--space-sm) 0;
	}

	.metric-content .metric-value {
		font-size: var(--text-2xl);
		font-weight: 700;
		color: var(--text-primary);
		margin: 0;
	}

	.live-events-section {
		display: grid;
		grid-template-columns: 2fr 1fr;
		gap: var(--space-xl);
	}

	.live-events,
	.top-content {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
	}

	.live-events h3,
	.top-content h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-lg) 0;
	}

	.events-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
		max-height: 400px;
		overflow-y: auto;
	}

	.event-item {
		display: flex;
		gap: var(--space-md);
		padding: var(--space-md);
		background: rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-md);
		transition: all var(--transition-fast);
	}

	.event-item:hover {
		background: rgba(255, 255, 255, 0.1);
	}

	.event-time {
		font-size: var(--text-xs);
		color: var(--text-secondary);
		white-space: nowrap;
		min-width: 60px;
	}

	.event-content {
		flex: 1;
	}

	.event-type {
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 2px;
	}

	.event-details {
		font-size: var(--text-xs);
		color: var(--text-secondary);
	}

	.content-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.content-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-md);
		background: rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-md);
	}

	.content-title {
		font-size: var(--text-sm);
		color: var(--text-primary);
		font-weight: 500;
		flex: 1;
		margin-right: var(--space-md);
	}

	.content-viewers {
		font-size: var(--text-xs);
		color: var(--primary);
		font-weight: 600;
	}

	/* Detailed Analytics Styles */
	.detailed-analytics {
		display: flex;
		flex-direction: column;
		gap: var(--space-2xl);
	}

	.detailed-analytics h2 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.analytics-section {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.section-header h3 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.geographic-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-xl);
	}

	.geographic-card {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
	}

	.geographic-card h4 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-lg) 0;
	}

	.country-list,
	.state-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.country-item,
	.state-item {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		padding: var(--space-md);
		background: rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-md);
	}

	.country-info,
	.state-info {
		display: flex;
		justify-content: space-between;
		align-items: center;
		min-width: 120px;
	}

	.country-name,
	.state-name {
		font-size: var(--text-sm);
		font-weight: 500;
		color: var(--text-primary);
	}

	.country-percentage,
	.state-percentage {
		font-size: var(--text-xs);
		color: var(--text-secondary);
	}

	.country-bar,
	.state-bar {
		flex: 1;
		height: 6px;
		background: rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-full);
		overflow: hidden;
		margin: 0 var(--space-md);
	}

	.country-fill,
	.state-fill {
		height: 100%;
		background: var(--primary);
		border-radius: var(--radius-full);
		transition: width var(--transition-normal);
	}

	.country-count,
	.state-count {
		font-size: var(--text-xs);
		color: var(--text-secondary);
		min-width: 80px;
		text-align: right;
	}

	.device-grid {
		display: grid;
		grid-template-columns: 2fr 1fr;
		gap: var(--space-xl);
	}

	.device-overview,
	.browser-stats {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
	}

	.device-overview h4,
	.browser-stats h4 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-lg) 0;
	}

	.device-stats {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.device-stat {
		display: flex;
		align-items: center;
		gap: var(--space-lg);
		padding: var(--space-lg);
		background: rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-lg);
	}

	.device-icon {
		width: 48px;
		height: 48px;
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.device-icon.desktop {
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
	}

	.device-icon.mobile {
		background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
	}

	.device-icon.tablet {
		background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
	}

	.device-icon svg {
		width: 24px;
		height: 24px;
		color: var(--white);
	}

	.device-info {
		flex: 1;
	}

	.device-name {
		font-size: var(--text-md);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-xs);
	}

	.device-users {
		font-size: var(--text-lg);
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: var(--space-xs);
	}

	.device-percentage {
		font-size: var(--text-sm);
		color: var(--primary);
		font-weight: 600;
		margin-bottom: var(--space-xs);
	}

	.device-session {
		font-size: var(--text-xs);
		color: var(--text-secondary);
	}

	.browser-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.browser-item {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		padding: var(--space-md);
		background: rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-md);
	}

	.browser-info {
		display: flex;
		justify-content: space-between;
		align-items: center;
		min-width: 100px;
	}

	.browser-name {
		font-size: var(--text-sm);
		font-weight: 500;
		color: var(--text-primary);
	}

	.browser-percentage {
		font-size: var(--text-xs);
		color: var(--text-secondary);
	}

	.browser-bar {
		flex: 1;
		height: 6px;
		background: rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-full);
		overflow: hidden;
		margin: 0 var(--space-md);
	}

	.browser-fill {
		height: 100%;
		background: var(--primary);
		border-radius: var(--radius-full);
		transition: width var(--transition-normal);
	}

	.browser-count {
		font-size: var(--text-xs);
		color: var(--text-secondary);
		min-width: 60px;
		text-align: right;
	}

	.funnel-container {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
	}

	.funnel-stages {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		overflow-x: auto;
		padding: var(--space-lg);
	}

	.funnel-stage {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		min-width: 150px;
		padding: var(--space-lg);
		background: linear-gradient(135deg, var(--primary) 0%, rgba(102, 126, 234, 0.8) 100%);
		border-radius: var(--radius-lg);
		color: var(--white);
		text-align: center;
		position: relative;
	}

	.stage-info {
		flex: 1;
	}

	.stage-name {
		font-size: var(--text-sm);
		font-weight: 600;
		margin-bottom: var(--space-xs);
	}

	.stage-count {
		font-size: var(--text-lg);
		font-weight: 700;
		margin-bottom: var(--space-xs);
	}

	.stage-conversion {
		font-size: var(--text-xs);
		opacity: 0.9;
	}

	.funnel-arrow {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 24px;
		height: 24px;
		background: var(--bg-glass);
		border-radius: 50%;
		color: var(--text-primary);
	}

	.funnel-arrow svg {
		width: 14px;
		height: 14px;
	}

	.metrics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
		gap: var(--space-xl);
	}

	.metric-card {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		transition: all var(--transition-normal);
	}

	.metric-card:hover {
		transform: translateY(-4px);
		box-shadow: var(--shadow-lg);
	}

	.metric-header {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		margin-bottom: var(--space-lg);
	}

	.metric-icon {
		width: 48px;
		height: 48px;
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: var(--shadow-md);
	}

	.metric-icon.users {
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
	}

	.metric-icon.videos {
		background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
	}

	.metric-icon.revenue {
		background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
	}

	.metric-icon.engagement {
		background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
	}

	.metric-icon svg {
		width: 24px;
		height: 24px;
		color: var(--white);
	}

	.metric-header h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.metric-value {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: var(--space-md);
	}

	.metric-details {
		display: flex;
		gap: var(--space-md);
		flex-wrap: wrap;
	}

	.metric-sub {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.metric-trend {
		font-size: var(--text-sm);
		font-weight: 600;
		padding: 2px 6px;
		border-radius: var(--radius-sm);
	}

	.metric-trend.positive {
		background: rgba(67, 233, 123, 0.1);
		color: var(--success);
	}

	.metric-trend.negative {
		background: rgba(255, 107, 107, 0.1);
		color: var(--error);
	}

	.analytics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
		gap: var(--space-xl);
	}

	.chart-card,
	.performance-card,
	.subscription-card,
	.system-card {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
	}

	.chart-header,
	.performance-header,
	.subscription-header,
	.system-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: var(--space-xl);
	}

	.chart-header h3,
	.performance-header h3,
	.subscription-header h3,
	.system-header h3 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.chart-legend {
		display: flex;
		gap: var(--space-lg);
	}

	.legend-item {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.legend-color {
		width: 12px;
		height: 12px;
		border-radius: 50%;
	}

	.legend-color.active {
		background: var(--primary);
	}

	.legend-color.new {
		background: var(--secondary);
	}

	.chart-placeholder {
		height: 200px;
		display: flex;
		flex-direction: column;
		justify-content: space-between;
	}

	.chart-bars {
		display: flex;
		align-items: end;
		gap: var(--space-md);
		height: 160px;
		padding: var(--space-md) 0;
	}

	.chart-bar {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 2px;
		height: 100%;
		justify-content: end;
	}

	.bar {
		width: 100%;
		border-radius: var(--radius-sm) var(--radius-sm) 0 0;
		transition: all var(--transition-normal);
	}

	.bar.active {
		background: var(--primary-gradient);
		opacity: 0.8;
	}

	.bar.new {
		background: var(--secondary-gradient);
		opacity: 0.6;
	}

	.chart-labels {
		display: flex;
		justify-content: space-between;
		padding: 0 var(--space-md);
		font-size: var(--text-xs);
		color: var(--text-secondary);
	}

	.performance-stats {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
		gap: var(--space-lg);
	}

	.stat-item {
		text-align: center;
	}

	.stat-label {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		margin-bottom: var(--space-sm);
	}

	.stat-value {
		font-size: var(--text-2xl);
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: var(--space-xs);
	}

	.stat-change {
		font-size: var(--text-sm);
		font-weight: 600;
		padding: 2px 6px;
		border-radius: var(--radius-sm);
	}

	.stat-change.positive {
		background: rgba(67, 233, 123, 0.1);
		color: var(--success);
	}

	.subscription-stats {
		display: flex;
		flex-direction: column;
		gap: var(--space-xl);
	}

	.subscription-metric {
		display: flex;
		align-items: center;
		gap: var(--space-lg);
	}

	.subscription-icon {
		width: 40px;
		height: 40px;
		border-radius: var(--radius-lg);
		background: var(--bg-glass);
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.subscription-icon svg {
		width: 20px;
		height: 20px;
		color: var(--text-primary);
	}

	.subscription-value {
		font-size: var(--text-xl);
		font-weight: 700;
		color: var(--text-primary);
	}

	.subscription-label {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.system-status {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		font-size: var(--text-sm);
		font-weight: 600;
	}

	.system-status.healthy {
		color: var(--success);
	}

	.status-indicator {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--success);
		animation: pulse 2s infinite;
	}

	@keyframes pulse {
		0% {
			box-shadow: 0 0 0 0 rgba(67, 233, 123, 0.7);
		}
		70% {
			box-shadow: 0 0 0 10px rgba(67, 233, 123, 0);
		}
		100% {
			box-shadow: 0 0 0 0 rgba(67, 233, 123, 0);
		}
	}

	.system-metrics {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
		gap: var(--space-lg);
	}

	.system-metric {
		text-align: center;
	}

	.system-label {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		margin-bottom: var(--space-sm);
	}

	.system-value {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
	}

	.btn {
		padding: var(--space-md) var(--space-lg);
		border: none;
		border-radius: var(--radius-lg);
		font-weight: 600;
		cursor: pointer;
		transition: all var(--transition-normal);
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		text-decoration: none;
		font-size: var(--text-base);
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

	.period-selector {
		margin-left: auto;
	}

	.period-select {
		padding: var(--space-md);
		border: none;
		border-radius: var(--radius-lg);
		background: var(--bg-glass);
		color: var(--text-primary);
		font-size: var(--text-sm);
		cursor: pointer;
		min-width: 150px;
	}

	.loading-container,
	.error-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-lg);
		padding: var(--space-3xl);
		text-align: center;
	}

	.error-state {
		border-radius: var(--radius-xl);
	}

	.error-icon {
		width: 64px;
		height: 64px;
		opacity: 0.5;
	}

	.error-icon svg {
		width: 100%;
		height: 100%;
		color: var(--error);
	}

	/* Responsive Design */
	@media (max-width: 1200px) {
		.geographic-grid,
		.device-grid {
			grid-template-columns: 1fr;
		}

		.live-events-section {
			grid-template-columns: 1fr;
		}

		.funnel-stages {
			flex-direction: column;
			align-items: stretch;
		}

		.funnel-stage {
			min-width: unset;
		}

		.funnel-arrow {
			transform: rotate(90deg);
		}
	}

	@media (max-width: 768px) {
		.header-content {
			flex-direction: column;
			align-items: stretch;
		}

		.header-controls {
			flex-direction: column;
			gap: var(--space-md);
		}

		.view-selector {
			justify-content: center;
		}

		.export-controls {
			justify-content: center;
		}

		.realtime-metrics {
			grid-template-columns: 1fr;
		}

		.device-stats {
			gap: var(--space-md);
		}

		.device-stat {
			flex-direction: column;
			text-align: center;
			gap: var(--space-md);
		}
	}

	@media (max-width: 480px) {
		.analytics-page {
			gap: var(--space-lg);
		}

		.realtime-card,
		.geographic-card,
		.device-overview,
		.browser-stats,
		.funnel-container {
			padding: var(--space-lg);
		}

		.view-selector {
			flex-direction: column;
			gap: var(--space-xs);
		}

		.view-btn {
			text-align: center;
		}
	}
</style> 