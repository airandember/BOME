<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import Navigation from '$lib/components/Navigation.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import type { AdCampaign } from '$lib/types/advertising';

	interface AnalyticsData {
		overview: {
			totalImpressions: number;
			totalClicks: number;
			totalSpent: number;
			averageCTR: number;
			averageCPC: number;
			activeCampaigns: number;
			totalRevenue: number;
			conversionRate: number;
		};
		dailyStats: Array<{
			date: string;
			impressions: number;
			clicks: number;
			spent: number;
			ctr: number;
		}>;
		campaignPerformance: Array<{
			id: number;
			name: string;
			impressions: number;
			clicks: number;
			spent: number;
			ctr: number;
			cpc: number;
			status: string;
		}>;
		topPerforming: {
			campaigns: Array<{ name: string; ctr: number; }>;
			ads: Array<{ title: string; clicks: number; }>;
		};
	}

	let analyticsData: AnalyticsData | null = null;
	let loading = true;
	let error: string | null = null;
	let selectedDateRange = '30d';
	let selectedMetric = 'impressions';

	const dateRanges = [
		{ value: '7d', label: 'Last 7 days' },
		{ value: '30d', label: 'Last 30 days' },
		{ value: '90d', label: 'Last 90 days' },
		{ value: 'custom', label: 'Custom range' }
	];

	const metrics = [
		{ value: 'impressions', label: 'Impressions' },
		{ value: 'clicks', label: 'Clicks' },
		{ value: 'spent', label: 'Spent' },
		{ value: 'ctr', label: 'CTR' }
	];

	onMount(async () => {
		if (!$auth.isAuthenticated) {
			goto('/login');
			return;
		}

		if ($auth.user?.role !== 'advertiser' && $auth.user?.role !== 'admin') {
			goto('/advertise');
			return;
		}

		await loadAnalytics();
	});

	async function loadAnalytics() {
		try {
			loading = true;
			
			// For mock users, create comprehensive mock analytics data
			if ($auth.token?.startsWith('mock-advertiser-token-')) {
				await new Promise(resolve => setTimeout(resolve, 1000)); // Simulate API call
				
				analyticsData = {
					overview: {
						totalImpressions: 156420,
						totalClicks: 4892,
						totalSpent: 2847.50,
						averageCTR: 3.13,
						averageCPC: 0.58,
						activeCampaigns: 3,
						totalRevenue: 8940.25,
						conversionRate: 2.4
					},
					dailyStats: generateDailyStats(),
					campaignPerformance: [
						{
							id: 1,
							name: 'Book of Mormon Research Campaign',
							impressions: 89420,
							clicks: 2847,
							spent: 1650.75,
							ctr: 3.18,
							cpc: 0.58,
							status: 'active'
						},
						{
							id: 2,
							name: 'Historical Evidence Promotion',
							impressions: 45230,
							clicks: 1456,
							spent: 843.25,
							ctr: 3.22,
							cpc: 0.58,
							status: 'active'
						},
						{
							id: 3,
							name: 'Archaeological Findings Campaign',
							impressions: 21770,
							clicks: 589,
							spent: 353.50,
							ctr: 2.71,
							cpc: 0.60,
							status: 'paused'
						}
					],
					topPerforming: {
						campaigns: [
							{ name: 'Historical Evidence Promotion', ctr: 3.22 },
							{ name: 'Book of Mormon Research Campaign', ctr: 3.18 },
							{ name: 'Archaeological Findings Campaign', ctr: 2.71 }
						],
						ads: [
							{ title: 'Discover Archaeological Evidence', clicks: 1247 },
							{ title: 'Book of Mormon Historical Context', clicks: 986 },
							{ title: 'Ancient American Civilizations', clicks: 743 }
						]
					}
				};
			} else {
				// Real API call for production
				const response = await fetch('/api/v1/advertiser/analytics', {
					headers: {
						'Authorization': `Bearer ${$auth.token}`
					}
				});

				if (response.ok) {
					const data = await response.json();
					analyticsData = data.data;
				} else {
					throw new Error('Failed to load analytics data');
				}
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load analytics';
		} finally {
			loading = false;
		}
	}

	function generateDailyStats() {
		const stats = [];
		const today = new Date();
		
		for (let i = 29; i >= 0; i--) {
			const date = new Date(today);
			date.setDate(date.getDate() - i);
			
			const baseImpressions = 3000 + Math.random() * 2000;
			const impressions = Math.floor(baseImpressions);
			const clicks = Math.floor(impressions * (0.025 + Math.random() * 0.015));
			const spent = clicks * (0.45 + Math.random() * 0.25);
			const ctr = (clicks / impressions) * 100;
			
			stats.push({
				date: date.toISOString().split('T')[0],
				impressions,
				clicks,
				spent: Math.round(spent * 100) / 100,
				ctr: Math.round(ctr * 100) / 100
			});
		}
		
		return stats;
	}

	function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(amount);
	}

	function formatNumber(num: number): string {
		return new Intl.NumberFormat('en-US').format(num);
	}

	function formatPercentage(num: number): string {
		return `${num.toFixed(2)}%`;
	}

	function getStatusClass(status: string): string {
		switch (status) {
			case 'active': return 'status-success';
			case 'paused': return 'status-secondary';
			case 'pending': return 'status-pending';
			default: return 'status-secondary';
		}
	}

	async function handleDateRangeChange() {
		await loadAnalytics();
	}

	function exportData() {
		// Mock export functionality
		alert('Analytics data export feature coming soon!');
	}
</script>

<svelte:head>
	<title>Analytics - Advertiser Dashboard - BOME</title>
</svelte:head>

<Navigation />

<div class="page-container">
	<div class="content-wrapper">
		<!-- Page Header -->
		<div class="page-header">
			<div class="header-content">
				<div class="header-text">
					<h1>Analytics Dashboard</h1>
					<p>Track your advertising performance and optimize your campaigns</p>
				</div>
				<div class="header-actions">
					<select bind:value={selectedDateRange} on:change={handleDateRangeChange} class="select-input">
						{#each dateRanges as range}
							<option value={range.value}>{range.label}</option>
						{/each}
					</select>
					<button class="btn btn-secondary" on:click={exportData}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M17 8l-5-5-5 5M12 3v12"/>
						</svg>
						Export
					</button>
				</div>
			</div>
		</div>

		{#if loading}
			<div class="loading-container">
				<LoadingSpinner />
			</div>
		{:else if error}
			<div class="error-card">
				<div class="error-icon">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="12" cy="12" r="10"/>
						<path d="M15 9l-6 6"/>
						<path d="M9 9l6 6"/>
					</svg>
				</div>
				<div class="error-content">
					<h3>Error loading analytics</h3>
					<p>{error}</p>
					<button class="btn btn-primary" on:click={loadAnalytics}>Try Again</button>
				</div>
			</div>
		{:else if analyticsData}
			<!-- Overview Cards -->
			<div class="overview-grid">
				<div class="metric-card">
					<div class="metric-icon impressions">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
							<path d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/>
						</svg>
					</div>
					<div class="metric-content">
						<div class="metric-value">{formatNumber(analyticsData.overview.totalImpressions)}</div>
						<div class="metric-label">Total Impressions</div>
					</div>
				</div>

				<div class="metric-card">
					<div class="metric-icon clicks">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M9 12l2 2 4-4"/>
							<path d="M21 12c-1.274 4.057-5.064 7-9 7s-7.726-2.943-9-7c1.274-4.057 5.064-7 9-7s7.726 2.943 9 7z"/>
						</svg>
					</div>
					<div class="metric-content">
						<div class="metric-value">{formatNumber(analyticsData.overview.totalClicks)}</div>
						<div class="metric-label">Total Clicks</div>
					</div>
				</div>

				<div class="metric-card">
					<div class="metric-icon spent">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<circle cx="12" cy="12" r="10"/>
							<path d="M12 6v6l4 2"/>
						</svg>
					</div>
					<div class="metric-content">
						<div class="metric-value">{formatCurrency(analyticsData.overview.totalSpent)}</div>
						<div class="metric-label">Total Spent</div>
					</div>
				</div>

				<div class="metric-card">
					<div class="metric-icon ctr">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
						</svg>
					</div>
					<div class="metric-content">
						<div class="metric-value">{formatPercentage(analyticsData.overview.averageCTR)}</div>
						<div class="metric-label">Average CTR</div>
					</div>
				</div>

				<div class="metric-card">
					<div class="metric-icon cpc">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<rect x="1" y="4" width="22" height="16" rx="2" ry="2"/>
							<line x1="1" y1="10" x2="23" y2="10"/>
						</svg>
					</div>
					<div class="metric-content">
						<div class="metric-value">{formatCurrency(analyticsData.overview.averageCPC)}</div>
						<div class="metric-label">Average CPC</div>
					</div>
				</div>

				<div class="metric-card">
					<div class="metric-icon revenue">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M12 2L2 7l10 5 10-5-10-5z"/>
							<path d="M2 17l10 5 10-5"/>
							<path d="M2 12l10 5 10-5"/>
						</svg>
					</div>
					<div class="metric-content">
						<div class="metric-value">{formatCurrency(analyticsData.overview.totalRevenue)}</div>
						<div class="metric-label">Total Revenue</div>
					</div>
				</div>
			</div>

			<!-- Charts Section -->
			<div class="charts-section">
				<div class="chart-card">
					<div class="chart-header">
						<h3>Performance Trends</h3>
						<select bind:value={selectedMetric} class="select-input small">
							{#each metrics as metric}
								<option value={metric.value}>{metric.label}</option>
							{/each}
						</select>
					</div>
					<div class="chart-container">
						<div class="chart-placeholder">
							<div class="chart-bars">
								{#each analyticsData.dailyStats.slice(-14) as stat, i}
									<div class="chart-bar" style="height: {(stat[selectedMetric] / Math.max(...analyticsData.dailyStats.map(s => s[selectedMetric]))) * 100}%">
										<div class="bar-tooltip">
											<div class="tooltip-date">{new Date(stat.date).toLocaleDateString()}</div>
											<div class="tooltip-value">
												{#if selectedMetric === 'spent'}
													{formatCurrency(stat[selectedMetric])}
												{:else if selectedMetric === 'ctr'}
													{formatPercentage(stat[selectedMetric])}
												{:else}
													{formatNumber(stat[selectedMetric])}
												{/if}
											</div>
										</div>
									</div>
								{/each}
							</div>
						</div>
					</div>
				</div>
			</div>

			<!-- Campaign Performance Table -->
			<div class="table-section">
				<div class="table-header">
					<h3>Campaign Performance</h3>
					<button class="btn btn-ghost" on:click={() => goto('/advertiser/campaigns')}>
						View All Campaigns
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M5 12h14M12 5l7 7-7 7"/>
						</svg>
					</button>
				</div>
				<div class="table-container">
					<table class="data-table">
						<thead>
							<tr>
								<th>Campaign</th>
								<th>Status</th>
								<th>Impressions</th>
								<th>Clicks</th>
								<th>CTR</th>
								<th>CPC</th>
								<th>Spent</th>
							</tr>
						</thead>
						<tbody>
							{#each analyticsData.campaignPerformance as campaign}
								<tr>
									<td>
										<div class="campaign-info">
											<div class="campaign-name">{campaign.name}</div>
										</div>
									</td>
									<td>
										<span class="status-badge {getStatusClass(campaign.status)}">
											{campaign.status}
										</span>
									</td>
									<td>{formatNumber(campaign.impressions)}</td>
									<td>{formatNumber(campaign.clicks)}</td>
									<td>{formatPercentage(campaign.ctr)}</td>
									<td>{formatCurrency(campaign.cpc)}</td>
									<td>{formatCurrency(campaign.spent)}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>

			<!-- Top Performers -->
			<div class="performers-grid">
				<div class="performer-card">
					<h3>Top Performing Campaigns</h3>
					<div class="performer-list">
						{#each analyticsData.topPerforming.campaigns as campaign, i}
							<div class="performer-item">
								<div class="performer-rank">#{i + 1}</div>
								<div class="performer-info">
									<div class="performer-name">{campaign.name}</div>
									<div class="performer-metric">{formatPercentage(campaign.ctr)} CTR</div>
								</div>
							</div>
						{/each}
					</div>
				</div>

				<div class="performer-card">
					<h3>Top Performing Ads</h3>
					<div class="performer-list">
						{#each analyticsData.topPerforming.ads as ad, i}
							<div class="performer-item">
								<div class="performer-rank">#{i + 1}</div>
								<div class="performer-info">
									<div class="performer-name">{ad.title}</div>
									<div class="performer-metric">{formatNumber(ad.clicks)} clicks</div>
								</div>
							</div>
						{/each}
					</div>
				</div>
			</div>
		{/if}
	</div>
</div>

<Footer />

<style>
	.page-container {
		min-height: 100vh;
		background: var(--bg-primary);
		padding-top: 80px;
	}

	.content-wrapper {
		max-width: 1400px;
		margin: 0 auto;
		padding: var(--space-xl) var(--space-lg);
		display: flex;
		flex-direction: column;
		gap: var(--space-xl);
	}

	.page-header {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
	}

	.header-content {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
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
	}

	.header-actions {
		display: flex;
		gap: var(--space-md);
		align-items: center;
	}

	.select-input {
		background: var(--bg-secondary);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-lg);
		padding: var(--space-md) var(--space-lg);
		color: var(--text-primary);
		font-size: var(--text-sm);
		min-width: 150px;
	}

	.select-input.small {
		padding: var(--space-sm) var(--space-md);
		min-width: 120px;
	}

	.overview-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
		gap: var(--space-lg);
	}

	.metric-card {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		display: flex;
		align-items: center;
		gap: var(--space-lg);
		transition: all var(--transition-normal);
	}

	.metric-card:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}

	.metric-icon {
		width: 60px;
		height: 60px;
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.metric-icon svg {
		width: 28px;
		height: 28px;
		color: var(--white);
	}

	.metric-icon.impressions {
		background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
	}

	.metric-icon.clicks {
		background: linear-gradient(135deg, #10b981 0%, #047857 100%);
	}

	.metric-icon.spent {
		background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
	}

	.metric-icon.ctr {
		background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
	}

	.metric-icon.cpc {
		background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
	}

	.metric-icon.revenue {
		background: linear-gradient(135deg, #06b6d4 0%, #0891b2 100%);
	}

	.metric-content {
		flex: 1;
	}

	.metric-value {
		font-size: var(--text-2xl);
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: var(--space-xs);
	}

	.metric-label {
		color: var(--text-secondary);
		font-size: var(--text-sm);
		font-weight: 500;
	}

	.charts-section {
		display: grid;
		gap: var(--space-lg);
	}

	.chart-card {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
	}

	.chart-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-xl);
	}

	.chart-header h3 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
	}

	.chart-container {
		height: 300px;
		position: relative;
	}

	.chart-placeholder {
		width: 100%;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.chart-bars {
		display: flex;
		align-items: end;
		gap: 4px;
		height: 250px;
		width: 100%;
	}

	.chart-bar {
		flex: 1;
		background: linear-gradient(180deg, var(--primary) 0%, var(--primary-dark) 100%);
		border-radius: 4px 4px 0 0;
		min-height: 10px;
		position: relative;
		transition: all var(--transition-normal);
	}

	.chart-bar:hover {
		opacity: 0.8;
	}

	.chart-bar:hover .bar-tooltip {
		opacity: 1;
		visibility: visible;
		transform: translateY(-10px);
	}

	.bar-tooltip {
		position: absolute;
		bottom: 100%;
		left: 50%;
		transform: translateX(-50%);
		background: var(--bg-secondary);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-md);
		padding: var(--space-sm);
		font-size: var(--text-xs);
		white-space: nowrap;
		opacity: 0;
		visibility: hidden;
		transition: all var(--transition-normal);
		z-index: 10;
	}

	.tooltip-date {
		color: var(--text-secondary);
		margin-bottom: 2px;
	}

	.tooltip-value {
		color: var(--text-primary);
		font-weight: 600;
	}

	.table-section {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
	}

	.table-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-xl);
	}

	.table-header h3 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
	}

	.table-container {
		overflow-x: auto;
	}

	.data-table {
		width: 100%;
		border-collapse: collapse;
	}

	.data-table th {
		text-align: left;
		padding: var(--space-md);
		color: var(--text-secondary);
		font-weight: 600;
		font-size: var(--text-sm);
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
	}

	.data-table td {
		padding: var(--space-md);
		border-bottom: 1px solid rgba(255, 255, 255, 0.05);
	}

	.campaign-info {
		display: flex;
		align-items: center;
		gap: var(--space-md);
	}

	.campaign-name {
		font-weight: 500;
		color: var(--text-primary);
	}

	.status-badge {
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-md);
		font-size: var(--text-xs);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.status-success {
		background: rgba(34, 197, 94, 0.2);
		color: #22c55e;
	}

	.status-secondary {
		background: rgba(156, 163, 175, 0.2);
		color: #9ca3af;
	}

	.status-pending {
		background: rgba(245, 158, 11, 0.2);
		color: #f59e0b;
	}

	.performers-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
		gap: var(--space-lg);
	}

	.performer-card {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
	}

	.performer-card h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-lg);
	}

	.performer-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.performer-item {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		padding: var(--space-md);
		background: rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-lg);
		transition: all var(--transition-normal);
	}

	.performer-item:hover {
		background: rgba(255, 255, 255, 0.1);
	}

	.performer-rank {
		width: 30px;
		height: 30px;
		background: var(--primary-gradient);
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 600;
		font-size: var(--text-sm);
		color: var(--white);
		flex-shrink: 0;
	}

	.performer-info {
		flex: 1;
	}

	.performer-name {
		color: var(--text-primary);
		font-weight: 500;
		margin-bottom: 2px;
	}

	.performer-metric {
		color: var(--text-secondary);
		font-size: var(--text-sm);
	}

	.loading-container {
		display: flex;
		justify-content: center;
		align-items: center;
		min-height: 400px;
	}

	.error-card {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border: 1px solid rgba(239, 68, 68, 0.2);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		text-align: center;
		max-width: 500px;
		margin: 0 auto;
	}

	.error-icon {
		width: 60px;
		height: 60px;
		background: rgba(239, 68, 68, 0.2);
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 0 auto var(--space-lg);
	}

	.error-icon svg {
		width: 24px;
		height: 24px;
		color: #ef4444;
	}

	.error-content h3 {
		color: var(--text-primary);
		font-size: var(--text-lg);
		font-weight: 600;
		margin-bottom: var(--space-sm);
	}

	.error-content p {
		color: var(--text-secondary);
		margin-bottom: var(--space-lg);
	}

	@media (max-width: 768px) {
		.content-wrapper {
			padding: var(--space-lg) var(--space-md);
		}

		.header-content {
			flex-direction: column;
			align-items: stretch;
		}

		.header-actions {
			justify-content: flex-end;
		}

		.overview-grid {
			grid-template-columns: 1fr;
		}

		.performers-grid {
			grid-template-columns: 1fr;
		}

		.table-container {
			font-size: var(--text-sm);
		}

		.chart-bars {
			gap: 2px;
		}
	}
</style> 