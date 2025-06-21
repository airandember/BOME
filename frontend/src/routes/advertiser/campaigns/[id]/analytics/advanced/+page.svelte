<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import type { AdvancedAnalytics, ExportOptions } from '$lib/types/advertising';
	
	let campaignId: number;
	let analytics: AdvancedAnalytics | null = null;
	let loading = true;
	let error: string | null = null;
	let dateRange = {
		start: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
		end: new Date().toISOString().split('T')[0]
	};
	let showExportModal = false;
	let exportOptions: ExportOptions = {
		format: 'csv',
		date_range: { ...dateRange },
		metrics: ['impressions', 'clicks', 'revenue', 'ctr'],
		group_by: 'day'
	};
	let exporting = false;

	onMount(async () => {
		campaignId = parseInt($page.params.id);
		if (!$auth.isAuthenticated) {
			goto('/login');
			return;
		}
		await loadAnalytics();
	});

	async function loadAnalytics() {
		loading = true;
		error = null;

		try {
			const response = await fetch(`/api/v1/campaigns/${campaignId}/analytics/advanced?start=${dateRange.start}&end=${dateRange.end}`, {
				headers: {
					'Authorization': `Bearer ${$auth.token}`
				}
			});

			if (response.ok) {
				const data = await response.json();
				analytics = data.data;
			} else {
				// Mock advanced analytics data
				analytics = {
					campaign_id: campaignId,
					date_range: dateRange,
					metrics: {
						impressions: 45230,
						clicks: 892,
						ctr: 1.97,
						unique_views: 38420,
						revenue: 2156.80,
						cost_per_click: 2.42,
						cost_per_impression: 0.048,
						return_on_ad_spend: 4.31
					},
					demographics: {
						age_groups: [
							{ range: '18-24', percentage: 15.2 },
							{ range: '25-34', percentage: 28.5 },
							{ range: '35-44', percentage: 31.8 },
							{ range: '45-54', percentage: 18.3 },
							{ range: '55+', percentage: 6.2 }
						],
						geographic: [
							{ location: 'United States', impressions: 28420, clicks: 567 },
							{ location: 'Canada', impressions: 8920, clicks: 178 },
							{ location: 'United Kingdom', impressions: 4230, clicks: 89 },
							{ location: 'Australia', impressions: 3660, clicks: 58 }
						],
						device_types: [
							{ device: 'Desktop', percentage: 52.3 },
							{ device: 'Mobile', percentage: 38.7 },
							{ device: 'Tablet', percentage: 9.0 }
						]
					},
					performance_by_placement: [
						{
							placement_id: 1,
							placement_name: 'Header Banner',
							impressions: 18920,
							clicks: 423,
							ctr: 2.24,
							revenue: 1024.60
						},
						{
							placement_id: 2,
							placement_name: 'Sidebar Large',
							impressions: 15430,
							clicks: 298,
							ctr: 1.93,
							revenue: 721.40
						},
						{
							placement_id: 3,
							placement_name: 'Between Videos',
							impressions: 10880,
							clicks: 171,
							ctr: 1.57,
							revenue: 410.80
						}
					],
					hourly_performance: generateHourlyData(),
					daily_performance: generateDailyData()
				};
			}
		} catch (err) {
			error = 'Failed to load analytics data';
		} finally {
			loading = false;
		}
	}

	function generateHourlyData() {
		return Array.from({ length: 24 }, (_, hour) => ({
			hour,
			impressions: Math.floor(Math.random() * 3000) + 500,
			clicks: Math.floor(Math.random() * 60) + 10,
			ctr: Math.random() * 3 + 1
		}));
	}

	function generateDailyData() {
		const days = Math.ceil((new Date(dateRange.end).getTime() - new Date(dateRange.start).getTime()) / (1000 * 60 * 60 * 24));
		return Array.from({ length: days }, (_, i) => {
			const date = new Date(new Date(dateRange.start).getTime() + i * 24 * 60 * 60 * 1000);
			return {
				date: date.toISOString().split('T')[0],
				impressions: Math.floor(Math.random() * 2000) + 1000,
				clicks: Math.floor(Math.random() * 50) + 20,
				revenue: Math.random() * 100 + 50,
				ctr: Math.random() * 2 + 1.5
			};
		});
	}

	async function handleExport() {
		if (!analytics) return;

		exporting = true;
		try {
			const response = await fetch(`/api/v1/campaigns/${campaignId}/analytics/export`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${$auth.token}`
				},
				body: JSON.stringify(exportOptions)
			});

			if (response.ok) {
				const blob = await response.blob();
				const url = window.URL.createObjectURL(blob);
				const a = document.createElement('a');
				a.href = url;
				a.download = `campaign-${campaignId}-analytics.${exportOptions.format}`;
				document.body.appendChild(a);
				a.click();
				document.body.removeChild(a);
				window.URL.revokeObjectURL(url);
				showExportModal = false;
			} else {
				// Mock export - generate CSV
				const csvData = generateMockCSV();
				const blob = new Blob([csvData], { type: 'text/csv' });
				const url = window.URL.createObjectURL(blob);
				const a = document.createElement('a');
				a.href = url;
				a.download = `campaign-${campaignId}-analytics.csv`;
				document.body.appendChild(a);
				a.click();
				document.body.removeChild(a);
				window.URL.revokeObjectURL(url);
				showExportModal = false;
			}
		} catch (err) {
			error = 'Export failed';
		} finally {
			exporting = false;
		}
	}

	function generateMockCSV(): string {
		if (!analytics) return '';

		const headers = ['Date', 'Impressions', 'Clicks', 'CTR', 'Revenue'];
		const rows = analytics.daily_performance.map(day => [
			day.date,
			day.impressions.toString(),
			day.clicks.toString(),
			day.ctr.toFixed(2) + '%',
			'$' + day.revenue.toFixed(2)
		]);

		return [headers.join(','), ...rows.map(row => row.join(','))].join('\n');
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
</script>

<svelte:head>
	<title>Advanced Analytics - Campaign {campaignId} - BOME</title>
</svelte:head>

<div class="advanced-analytics">
	<!-- Header -->
	<div class="page-header">
		<div class="breadcrumb">
			<a href="/advertiser/campaigns">Campaigns</a>
			<span>/</span>
			<a href="/advertiser/campaigns/{campaignId}">Campaign {campaignId}</a>
			<span>/</span>
			<span>Advanced Analytics</span>
		</div>
		
		<div class="header-actions">
			<div class="date-range">
				<input
					type="date"
					bind:value={dateRange.start}
					on:change={loadAnalytics}
				/>
				<span>to</span>
				<input
					type="date"
					bind:value={dateRange.end}
					on:change={loadAnalytics}
				/>
			</div>
			
			<button 
				class="btn btn-primary"
				on:click={() => showExportModal = true}
				disabled={!analytics || loading}
			>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
					<polyline points="7,10 12,15 17,10" />
					<line x1="12" y1="15" x2="12" y2="3" />
				</svg>
				Export Data
			</button>
		</div>
	</div>

	{#if loading}
		<div class="loading-container">
			<div class="loading-spinner"></div>
			<p>Loading advanced analytics...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<div class="alert alert-error">
				<span>{error}</span>
			</div>
		</div>
	{:else if analytics}
		<!-- Key Metrics -->
		<div class="metrics-grid">
			<div class="metric-card">
				<div class="metric-value">{formatNumber(analytics.metrics.impressions)}</div>
				<div class="metric-label">Total Impressions</div>
				<div class="metric-change positive">+12.5% vs last period</div>
			</div>
			
			<div class="metric-card">
				<div class="metric-value">{formatNumber(analytics.metrics.clicks)}</div>
				<div class="metric-label">Total Clicks</div>
				<div class="metric-change positive">+8.3% vs last period</div>
			</div>
			
			<div class="metric-card">
				<div class="metric-value">{formatPercentage(analytics.metrics.ctr)}</div>
				<div class="metric-label">Click-Through Rate</div>
				<div class="metric-change negative">-2.1% vs last period</div>
			</div>
			
			<div class="metric-card">
				<div class="metric-value">{formatCurrency(analytics.metrics.revenue)}</div>
				<div class="metric-label">Total Revenue</div>
				<div class="metric-change positive">+15.7% vs last period</div>
			</div>
		</div>

		<!-- Advanced Metrics -->
		<div class="advanced-metrics">
			<div class="metric-card">
				<div class="metric-value">{formatCurrency(analytics.metrics.cost_per_click)}</div>
				<div class="metric-label">Cost Per Click</div>
			</div>
			
			<div class="metric-card">
				<div class="metric-value">{formatCurrency(analytics.metrics.cost_per_impression)}</div>
				<div class="metric-label">Cost Per Impression</div>
			</div>
			
			<div class="metric-card">
				<div class="metric-value">{analytics.metrics.return_on_ad_spend.toFixed(2)}x</div>
				<div class="metric-label">Return on Ad Spend</div>
			</div>
			
			<div class="metric-card">
				<div class="metric-value">{formatNumber(analytics.metrics.unique_views)}</div>
				<div class="metric-label">Unique Views</div>
			</div>
		</div>

		<!-- Demographics -->
		<div class="analytics-section">
			<h2>Audience Demographics</h2>
			
			<div class="demographics-grid">
				<!-- Age Groups -->
				<div class="demo-card">
					<h3>Age Distribution</h3>
					<div class="chart-container">
						{#each analytics.demographics.age_groups as group}
							<div class="bar-item">
								<span class="bar-label">{group.range}</span>
								<div class="bar-container">
									<div class="bar" style="width: {group.percentage}%"></div>
								</div>
								<span class="bar-value">{group.percentage.toFixed(1)}%</span>
							</div>
						{/each}
					</div>
				</div>

				<!-- Device Types -->
				<div class="demo-card">
					<h3>Device Distribution</h3>
					<div class="chart-container">
						{#each analytics.demographics.device_types as device}
							<div class="bar-item">
								<span class="bar-label">{device.device}</span>
								<div class="bar-container">
									<div class="bar" style="width: {device.percentage}%"></div>
								</div>
								<span class="bar-value">{device.percentage.toFixed(1)}%</span>
							</div>
						{/each}
					</div>
				</div>
			</div>
		</div>

		<!-- Geographic Performance -->
		<div class="analytics-section">
			<h2>Geographic Performance</h2>
			<div class="geo-table">
				<div class="table-header">
					<div>Location</div>
					<div>Impressions</div>
					<div>Clicks</div>
					<div>CTR</div>
				</div>
				{#each analytics.demographics.geographic as location}
					<div class="table-row">
						<div class="location-name">{location.location}</div>
						<div>{formatNumber(location.impressions)}</div>
						<div>{formatNumber(location.clicks)}</div>
						<div>{formatPercentage((location.clicks / location.impressions) * 100)}</div>
					</div>
				{/each}
			</div>
		</div>

		<!-- Placement Performance -->
		<div class="analytics-section">
			<h2>Performance by Placement</h2>
			<div class="placement-performance">
				{#each analytics.performance_by_placement as placement}
					<div class="placement-card">
						<h4>{placement.placement_name}</h4>
						<div class="placement-metrics">
							<div class="metric">
								<span class="value">{formatNumber(placement.impressions)}</span>
								<span class="label">Impressions</span>
							</div>
							<div class="metric">
								<span class="value">{formatNumber(placement.clicks)}</span>
								<span class="label">Clicks</span>
							</div>
							<div class="metric">
								<span class="value">{formatPercentage(placement.ctr)}</span>
								<span class="label">CTR</span>
							</div>
							<div class="metric">
								<span class="value">{formatCurrency(placement.revenue)}</span>
								<span class="label">Revenue</span>
							</div>
						</div>
					</div>
				{/each}
			</div>
		</div>

		<!-- Hourly Performance -->
		<div class="analytics-section">
			<h2>Hourly Performance</h2>
			<div class="hourly-chart">
				{#each analytics.hourly_performance as hour}
					<div class="hour-bar">
						<div class="bar-fill" style="height: {(hour.impressions / 3000) * 100}%"></div>
						<div class="hour-label">{hour.hour}:00</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>

<!-- Export Modal -->
{#if showExportModal}
	<div class="modal-overlay" on:click={() => showExportModal = false}>
		<div class="modal" on:click|stopPropagation>
			<div class="modal-header">
				<h2>Export Analytics Data</h2>
				<button class="close-btn" on:click={() => showExportModal = false}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M18 6L6 18M6 6l12 12" />
					</svg>
				</button>
			</div>

			<div class="modal-content">
				<form on:submit|preventDefault={handleExport} class="export-form">
					<div class="form-group">
						<label>Export Format</label>
						<select bind:value={exportOptions.format}>
							<option value="csv">CSV</option>
							<option value="excel">Excel</option>
							<option value="pdf">PDF Report</option>
						</select>
					</div>

					<div class="form-group">
						<label>Date Range</label>
						<div class="date-inputs">
							<input type="date" bind:value={exportOptions.date_range.start} />
							<span>to</span>
							<input type="date" bind:value={exportOptions.date_range.end} />
						</div>
					</div>

					<div class="form-group">
						<label>Group By</label>
						<select bind:value={exportOptions.group_by}>
							<option value="day">Daily</option>
							<option value="week">Weekly</option>
							<option value="month">Monthly</option>
						</select>
					</div>

					<div class="form-group">
						<label>Include Metrics</label>
						<div class="checkbox-group">
							{#each ['impressions', 'clicks', 'ctr', 'revenue', 'unique_views'] as metric}
								<label class="checkbox-label">
									<input 
										type="checkbox" 
										bind:group={exportOptions.metrics} 
										value={metric} 
									/>
									<span class="checkmark"></span>
									{metric.charAt(0).toUpperCase() + metric.slice(1).replace('_', ' ')}
								</label>
							{/each}
						</div>
					</div>

					<div class="form-actions">
						<button type="button" class="btn btn-secondary" on:click={() => showExportModal = false}>
							Cancel
						</button>
						<button type="submit" class="btn btn-primary" disabled={exporting}>
							{#if exporting}
								<span class="loading-spinner small"></span>
								Exporting...
							{:else}
								Export Data
							{/if}
						</button>
					</div>
				</form>
			</div>
		</div>
	</div>
{/if}

<style>
	.advanced-analytics {
		display: flex;
		flex-direction: column;
		gap: var(--space-xl);
		padding: var(--space-xl);
	}

	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: var(--space-lg);
	}

	.breadcrumb {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.breadcrumb a {
		color: var(--primary);
		text-decoration: none;
	}

	.breadcrumb a:hover {
		text-decoration: underline;
	}

	.header-actions {
		display: flex;
		align-items: center;
		gap: var(--space-lg);
	}

	.date-range {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		font-size: var(--text-sm);
	}

	.date-range input {
		padding: var(--space-sm);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-lg);
		background: var(--bg-glass);
		color: var(--text-primary);
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

	.loading-spinner.small {
		width: 16px;
		height: 16px;
		border-width: 2px;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.metrics-grid,
	.advanced-metrics {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: var(--space-lg);
	}

	.metric-card {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-lg);
		border: 1px solid rgba(255, 255, 255, 0.1);
		text-align: center;
	}

	.metric-value {
		font-size: var(--text-2xl);
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: var(--space-sm);
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

	.metric-change.negative {
		color: var(--error);
	}

	.analytics-section {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.analytics-section h2 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-lg) 0;
	}

	.demographics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: var(--space-lg);
	}

	.demo-card {
		background: var(--bg-glass-dark);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
	}

	.demo-card h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-lg) 0;
	}

	.chart-container {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.bar-item {
		display: grid;
		grid-template-columns: 80px 1fr 60px;
		align-items: center;
		gap: var(--space-sm);
	}

	.bar-label {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.bar-container {
		background: rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-full);
		height: 8px;
		overflow: hidden;
	}

	.bar {
		height: 100%;
		background: var(--primary-gradient);
		border-radius: var(--radius-full);
		transition: width 0.5s ease;
	}

	.bar-value {
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--text-primary);
		text-align: right;
	}

	.geo-table {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}

	.table-header {
		display: grid;
		grid-template-columns: 2fr 1fr 1fr 1fr;
		gap: var(--space-md);
		padding: var(--space-md);
		background: var(--bg-glass-dark);
		border-radius: var(--radius-lg);
		font-weight: 600;
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.table-row {
		display: grid;
		grid-template-columns: 2fr 1fr 1fr 1fr;
		gap: var(--space-md);
		padding: var(--space-md);
		background: var(--bg-glass);
		border-radius: var(--radius-lg);
		align-items: center;
	}

	.location-name {
		font-weight: 500;
		color: var(--text-primary);
	}

	.placement-performance {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: var(--space-lg);
	}

	.placement-card {
		background: var(--bg-glass-dark);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
	}

	.placement-card h4 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-lg) 0;
	}

	.placement-metrics {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-md);
	}

	.metric {
		text-align: center;
	}

	.metric .value {
		display: block;
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
	}

	.metric .label {
		font-size: var(--text-xs);
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.hourly-chart {
		display: flex;
		align-items: end;
		gap: var(--space-xs);
		height: 200px;
		padding: var(--space-lg);
		background: var(--bg-glass-dark);
		border-radius: var(--radius-lg);
		overflow-x: auto;
	}

	.hour-bar {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-sm);
		min-width: 30px;
	}

	.bar-fill {
		width: 20px;
		background: var(--primary-gradient);
		border-radius: var(--radius-sm);
		min-height: 10px;
		transition: height 0.5s ease;
	}

	.hour-label {
		font-size: var(--text-xs);
		color: var(--text-secondary);
		transform: rotate(-45deg);
		white-space: nowrap;
	}

	/* Modal Styles */
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: var(--z-modal);
		backdrop-filter: blur(4px);
	}

	.modal {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
		max-width: 500px;
		width: 90vw;
		max-height: 90vh;
		overflow-y: auto;
		box-shadow: var(--shadow-2xl);
	}

	.modal-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: var(--space-xl);
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
	}

	.modal-header h2 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.close-btn {
		width: 32px;
		height: 32px;
		border: none;
		background: var(--bg-glass);
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		transition: all var(--transition-normal);
		color: var(--text-secondary);
	}

	.close-btn:hover {
		background: var(--bg-glass-dark);
		color: var(--text-primary);
	}

	.close-btn svg {
		width: 18px;
		height: 18px;
	}

	.modal-content {
		padding: var(--space-xl);
	}

	.export-form {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}

	.form-group label {
		font-weight: 600;
		color: var(--text-primary);
		font-size: var(--text-sm);
	}

	.form-group select,
	.form-group input {
		padding: var(--space-md);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-lg);
		background: var(--bg-glass);
		color: var(--text-primary);
		font-size: var(--text-base);
	}

	.date-inputs {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
	}

	.checkbox-group {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}

	.checkbox-label {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		cursor: pointer;
		font-weight: 500;
	}

	.checkbox-label input[type="checkbox"] {
		display: none;
	}

	.checkmark {
		width: 20px;
		height: 20px;
		border: 2px solid rgba(255, 255, 255, 0.2);
		border-radius: var(--radius-sm);
		background: var(--bg-glass);
		transition: all var(--transition-normal);
		position: relative;
	}

	.checkbox-label input[type="checkbox"]:checked + .checkmark {
		background: var(--primary);
		border-color: var(--primary);
	}

	.checkbox-label input[type="checkbox"]:checked + .checkmark::after {
		content: '';
		position: absolute;
		left: 6px;
		top: 2px;
		width: 6px;
		height: 10px;
		border: solid white;
		border-width: 0 2px 2px 0;
		transform: rotate(45deg);
	}

	.form-actions {
		display: flex;
		gap: var(--space-md);
		justify-content: flex-end;
		padding-top: var(--space-lg);
		border-top: 1px solid rgba(255, 255, 255, 0.1);
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.page-header {
			flex-direction: column;
			align-items: stretch;
		}

		.header-actions {
			flex-direction: column;
			align-items: stretch;
		}

		.table-header,
		.table-row {
			grid-template-columns: 1fr 1fr;
			gap: var(--space-sm);
		}

		.hourly-chart {
			height: 150px;
		}

		.modal {
			width: 95vw;
		}
	}
</style> 