<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { browser } from '$app/environment';
	import { auth, apiRequest } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	interface SystemMetrics {
		cpu_usage: number;
		memory_usage: number;
		disk_usage: number;
		network_in: string;
		network_out: string;
		uptime: string;
		load_average: number[];
	}

	interface WebhookEvent {
		id: string;
		timestamp: string;
		event_type: string;
		subsite: string;
		status: 'success' | 'failed' | 'pending';
		response_time: number;
		payload_size: number;
		error_message?: string;
	}

	interface SubsiteHealth {
		streaming: {
			status: 'healthy' | 'warning' | 'critical';
			response_time: number;
			error_rate: number;
			active_connections: number;
			last_check: string;
		};
		articles: {
			status: 'healthy' | 'warning' | 'critical';
			response_time: number;
			error_rate: number;
			active_connections: number;
			last_check: string;
		};
		expo: {
			status: 'healthy' | 'warning' | 'critical';
			response_time: number;
			error_rate: number;
			active_connections: number;
			last_check: string;
		};
	}

	interface Alert {
		id: string;
		severity: 'info' | 'warning' | 'critical';
		title: string;
		message: string;
		timestamp: string;
		subsite?: string;
		acknowledged: boolean;
	}

	let systemMetrics: SystemMetrics = {
		cpu_usage: 0,
		memory_usage: 0,
		disk_usage: 0,
		network_in: '0 MB/s',
		network_out: '0 MB/s',
		uptime: '0 days',
		load_average: [0, 0, 0]
	};

	let webhookEvents: WebhookEvent[] = [];
	let subsiteHealth: SubsiteHealth = {
		streaming: { status: 'healthy', response_time: 0, error_rate: 0, active_connections: 0, last_check: '' },
		articles: { status: 'healthy', response_time: 0, error_rate: 0, active_connections: 0, last_check: '' },
		expo: { status: 'healthy', response_time: 0, error_rate: 0, active_connections: 0, last_check: '' }
	};

	let alerts: Alert[] = [];
	let loading = true;
	let error = '';
	let selectedView = 'overview';
	let refreshInterval: NodeJS.Timeout | undefined;
	let visibilityHandler: (() => void) | null = null;
	let subsiteFilter = 'all';
	let webhookTypeFilter = 'all';

	onMount(async () => {
		if (!$auth.isAuthenticated) {
			return;
		}

		await loadMonitoringData();

		const startPolling = () => {
			if (refreshInterval) clearInterval(refreshInterval);
			refreshInterval = setInterval(async () => {
				await loadMonitoringData();
			}, 10000);
		};

		const stopPolling = () => {
			if (refreshInterval) {
				clearInterval(refreshInterval);
				refreshInterval = undefined;
			}
		};

		visibilityHandler = () => {
			if (document.hidden) {
				stopPolling();
			} else {
				startPolling();
			}
		};

		document.addEventListener('visibilitychange', visibilityHandler);
		startPolling();
	});

	onDestroy(() => {
		if (refreshInterval) {
			clearInterval(refreshInterval);
		}
		if (visibilityHandler) {
			document.removeEventListener('visibilitychange', visibilityHandler);
			visibilityHandler = null;
		}
	});

	async function loadMonitoringData() {
		try {
			loading = true;
			
			// Load system metrics
			const metricsResponse = await apiRequest('/admin/monitoring/system');
			
			if (metricsResponse.ok) {
				const data = await metricsResponse.json();
				// Ensure we have a valid metrics object with all required properties
				if (data && data.metrics) {
					systemMetrics = {
						cpu_usage: data.metrics.cpu_usage ?? 0,
						memory_usage: data.metrics.memory_usage ?? 0,
						disk_usage: data.metrics.disk_usage ?? 0,
						network_in: data.metrics.network_in ?? '0 MB/s',
						network_out: data.metrics.network_out ?? '0 MB/s',
						uptime: data.metrics.uptime ?? '0 days',
						load_average: data.metrics.load_average ?? [0, 0, 0]
					};
				}
			}

			// Load webhook events
			const webhooksResponse = await apiRequest('/admin/monitoring/webhooks');
			
			if (webhooksResponse.ok) {
				const data = await webhooksResponse.json();
				webhookEvents = data.events || [];
			}

			// Load subsite health
			const healthResponse = await apiRequest('/admin/monitoring/health');
			
			if (healthResponse.ok) {
				const data = await healthResponse.json();
				// Ensure we have valid health data for each subsite
				if (data && data.health) {
					subsiteHealth = {
						streaming: {
							status: data.health.streaming?.status ?? 'healthy',
							response_time: data.health.streaming?.response_time ?? 0,
							error_rate: data.health.streaming?.error_rate ?? 0,
							active_connections: data.health.streaming?.active_connections ?? 0,
							last_check: data.health.streaming?.last_check ?? ''
						},
						articles: {
							status: data.health.articles?.status ?? 'healthy',
							response_time: data.health.articles?.response_time ?? 0,
							error_rate: data.health.articles?.error_rate ?? 0,
							active_connections: data.health.articles?.active_connections ?? 0,
							last_check: data.health.articles?.last_check ?? ''
						},
						expo: {
							status: data.health.expo?.status ?? 'healthy',
							response_time: data.health.expo?.response_time ?? 0,
							error_rate: data.health.expo?.error_rate ?? 0,
							active_connections: data.health.expo?.active_connections ?? 0,
							last_check: data.health.expo?.last_check ?? ''
						}
					};
				}
			}

			// Load alerts
			const alertsResponse = await apiRequest('/admin/monitoring/alerts');
			
			if (alertsResponse.ok) {
				const data = await alertsResponse.json();
				alerts = data.alerts || [];
			}

		} catch (err: any) {
			error = 'Failed to load monitoring data';
			console.error('Monitoring load error:', err);
		} finally {
			loading = false;
		}
	}

	// polling handled via onMount visibility-aware logic

	function getStatusColor(status: string) {
		switch (status) {
			case 'healthy': return '#10b981';
			case 'warning': return '#f59e0b';
			case 'critical': return '#ef4444';
			default: return '#6b7280';
		}
	}

	function getSeverityColor(severity: string) {
		switch (severity) {
			case 'info': return '#3b82f6';
			case 'warning': return '#f59e0b';
			case 'critical': return '#ef4444';
			default: return '#6b7280';
		}
	}

	function formatBytes(bytes: number) {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	}

	function formatUptime(uptime: string) {
		return uptime;
	}

	// Safe access functions for system metrics
	function getCpuUsage(): number {
		return systemMetrics?.cpu_usage ?? 0;
	}

	function getMemoryUsage(): number {
		return systemMetrics?.memory_usage ?? 0;
	}

	function getDiskUsage(): number {
		return systemMetrics?.disk_usage ?? 0;
	}

	function getNetworkIn(): string {
		return systemMetrics?.network_in ?? '0 MB/s';
	}

	function getNetworkOut(): string {
		return systemMetrics?.network_out ?? '0 MB/s';
	}

	function getUptime(): string {
		return systemMetrics?.uptime ?? '0 days';
	}

	function getLoadAverage(): number[] {
		return systemMetrics?.load_average ?? [0, 0, 0];
	}

	// Safe access functions for subsite health
	function getHealthErrorRate(health: any): number {
		return health?.error_rate ?? 0;
	}

	function getHealthResponseTime(health: any): number {
		return health?.response_time ?? 0;
	}

	function getHealthActiveConnections(health: any): number {
		return health?.active_connections ?? 0;
	}

	function getHealthLastCheck(health: any): string {
		return health?.last_check ?? '';
	}

	function getHealthStatus(health: any): string {
		return health?.status ?? 'unknown';
	}

	async function acknowledgeAlert(alertId: string) {
		try {
			const response = await apiRequest(`/admin/monitoring/alerts/${alertId}/acknowledge`, {
				method: 'POST'
			});
			
			if (response.ok) {
				alerts = alerts.map(alert => 
					alert.id === alertId ? { ...alert, acknowledged: true } : alert
				);
				showToast('Alert acknowledged', 'success');
			}
		} catch (err) {
			showToast('Failed to acknowledge alert', 'error');
		}
	}

	function getWebhookStatusIcon(status: string) {
		switch (status) {
			case 'success':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<polyline points="20,6 9,17 4,12"></polyline>
				</svg>`;
			case 'failed':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="12" cy="12" r="10"></circle>
					<line x1="15" y1="9" x2="9" y2="15"></line>
					<line x1="9" y1="9" x2="15" y2="15"></line>
				</svg>`;
			case 'pending':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="12" cy="12" r="10"></circle>
					<polyline points="12,6 12,12 16,14"></polyline>
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
	<title>Server Monitoring - BOME Admin</title>
	<meta name="description" content="Real-time server monitoring and system health" />
</svelte:head>

{#if loading}
	<div class="loading-container">
		<LoadingSpinner size="large" />
		<p>Loading monitoring data...</p>
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
		<button class="retry-button" on:click={loadMonitoringData}>Retry</button>
	</div>
{:else}
	<div class="monitoring-dashboard">
		<!-- Header -->
		<header class="dashboard-header">
			<div class="header-content">
				<div class="header-info">
					<h1>Server Monitoring</h1>
					<p>Real-time system health and performance metrics</p>
				</div>
				<div class="header-actions">
					<div class="view-tabs">
						<button class="tab-button" class:active={selectedView === 'overview'} on:click={() => selectedView = 'overview'}>
							Overview
						</button>
						<button class="tab-button" class:active={selectedView === 'webhooks'} on:click={() => selectedView = 'webhooks'}>
							Webhooks
						</button>
						<button class="tab-button" class:active={selectedView === 'alerts'} on:click={() => selectedView = 'alerts'}>
							Alerts
						</button>
					</div>
					<button class="refresh-button" on:click={loadMonitoringData}>
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
		</header>

		{#if selectedView === 'overview'}
			<!-- System Overview -->
			<div class="overview-section">
				<!-- System Metrics -->
				<div class="metrics-grid">
					<div class="metric-card">
						<div class="metric-header">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect>
								<line x1="8" y1="21" x2="16" y2="21"></line>
								<line x1="12" y1="17" x2="12" y2="21"></line>
							</svg>
							<span>CPU Usage</span>
						</div>
						<div class="metric-value">{getCpuUsage().toFixed(1)}%</div>
						<div class="metric-bar">
							<div class="metric-progress" style="width: {getCpuUsage()}%"></div>
						</div>
					</div>

					<div class="metric-card">
						<div class="metric-header">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M22 12h-4l-3 9L9 3l-3 9H2"></path>
							</svg>
							<span>Memory Usage</span>
						</div>
						<div class="metric-value">{getMemoryUsage().toFixed(1)}%</div>
						<div class="metric-bar">
							<div class="metric-progress" style="width: {getMemoryUsage()}%"></div>
						</div>
					</div>

					<div class="metric-card">
						<div class="metric-header">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M22 12h-4l-3 9L9 3l-3 9H2"></path>
							</svg>
							<span>Disk Usage</span>
						</div>
						<div class="metric-value">{getDiskUsage().toFixed(1)}%</div>
						<div class="metric-bar">
							<div class="metric-progress" style="width: {getDiskUsage()}%"></div>
						</div>
					</div>

					<div class="metric-card">
						<div class="metric-header">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<circle cx="12" cy="12" r="10"></circle>
								<polyline points="12,6 12,12 16,14"></polyline>
							</svg>
							<span>Uptime</span>
						</div>
						<div class="metric-value">{getUptime()}</div>
						<div class="metric-subtitle">System running</div>
					</div>
				</div>

				<!-- Subsite Health -->
				<div class="live-indicator">Live</div>
				<div class="health-section prominent">
					<h2>Subsite Health Status</h2>
					<div class="health-grid">
						{#each Object.entries(subsiteHealth) as [subsite, health]}
							<div class="health-card" class:status={getHealthStatus(health)}>
								<div class="health-header">
									<div class="subsite-name">{subsite?.charAt(0)?.toUpperCase() + subsite?.slice(1) || 'Unknown'}</div>
									<div class="status-indicator" style="background-color: {getStatusColor(getHealthStatus(health))}"></div>
								</div>
								<div class="health-metrics">
									<div class="health-metric">
										<span class="metric-label">Response Time:</span>
										<span class="metric-value">{getHealthResponseTime(health)}ms</span>
									</div>
									<div class="health-metric">
										<span class="metric-label">Error Rate:</span>
										<span class="metric-value">{getHealthErrorRate(health).toFixed(2)}%</span>
									</div>
									<div class="health-metric">
										<span class="metric-label">Active Connections:</span>
										<span class="metric-value">{getHealthActiveConnections(health)}</span>
									</div>
									<div class="health-metric">
										<span class="metric-label">Last Check:</span>
										<span class="metric-value">{getHealthLastCheck(health) ? new Date(getHealthLastCheck(health)).toLocaleTimeString() : 'N/A'}</span>
									</div>
								</div>
							</div>
						{/each}
					</div>
				</div>
			</div>
		{:else if selectedView === 'webhooks'}
			<!-- Webhook Monitoring -->
			<div class="webhooks-section">
				<h2>Webhook Events</h2>
				<div class="webhook-filters">
					<label>Subsite:
						<select bind:value={subsiteFilter} on:change={loadMonitoringData}>
							<option value="all">All</option>
							<option value="streaming">Streaming</option>
							<option value="articles">Articles</option>
							<option value="expo">Expo</option>
						</select>
					</label>
					<label>Type:
						<select bind:value={webhookTypeFilter} on:change={loadMonitoringData}>
							<option value="all">All</option>
							{#each Array.from(new Set(webhookEvents.map(e => e.event_type))) as type}
								<option value={type}>{type}</option>
							{/each}
						</select>
					</label>
				</div>
				<div class="webhooks-table">
					<div class="table-header">
						<div class="header-cell">Time</div>
						<div class="header-cell">Event Type</div>
						<div class="header-cell">Subsite</div>
						<div class="header-cell">Status</div>
						<div class="header-cell">Response Time</div>
						<div class="header-cell">Payload Size</div>
					</div>
					{#each webhookEvents.filter(event => 
						(subsiteFilter === 'all' || event.subsite === subsiteFilter) && 
						(webhookTypeFilter === 'all' || event.event_type === webhookTypeFilter)
					) as event}
						<div class="table-row">
							<div class="table-cell">{event.timestamp ? new Date(event.timestamp).toLocaleString() : 'N/A'}</div>
							<div class="table-cell">{event.event_type || 'N/A'}</div>
							<div class="table-cell">{event.subsite || 'N/A'}</div>
							<div class="table-cell">
								<div class="status-badge status-{event.status || 'unknown'}">
									<div class="status-icon" style="display: inline-block">{@html getWebhookStatusIcon(event.status)}</div>
									{event.status || 'unknown'}
								</div>
							</div>
							<div class="table-cell">{event.response_time || 0}ms</div>
							<div class="table-cell">{formatBytes(event.payload_size || 0)}</div>
						</div>
					{/each}
				</div>
			</div>
		{:else if selectedView === 'alerts'}
			<!-- Alerts -->
			<div class="alerts-section">
				<h2>System Alerts</h2>
				<div class="alerts-list">
					{#each alerts as alert}
						<div class="alert-item" class:class-{alert.severity}={true} class:acknowledged={alert.acknowledged}>
							<div class="alert-header">
								<div class="alert-severity" style="background-color: {getSeverityColor(alert.severity)}"></div>
								<div class="alert-title">{alert.title || 'Unknown Alert'}</div>
								<div class="alert-time">{alert.timestamp ? new Date(alert.timestamp).toLocaleString() : 'N/A'}</div>
								{#if !alert.acknowledged}
									<button class="acknowledge-button" on:click={() => acknowledgeAlert(alert.id)}>
										Acknowledge
									</button>
								{/if}
							</div>
							<div class="alert-message">{alert.message || 'No message available'}</div>
							{#if alert.subsite}
								<div class="alert-subsite">Subsite: {alert.subsite}</div>
							{/if}
						</div>
					{/each}
				</div>
			</div>
		{/if}
	</div>
{/if}

<style>
	.monitoring-dashboard {
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
		margin: 0;
	}

	.header-actions {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.view-tabs {
		display: flex;
		gap: 0.5rem;
	}

	.tab-button {
		padding: 0.5rem 1rem;
		border: 1px solid rgba(255, 255, 255, 0.2);
		border-radius: 8px;
		background: var(--bg-glass);
		color: var(--text-secondary);
		cursor: pointer;
		transition: all 0.3s ease;
		font-size: 0.9rem;
	}

	.tab-button.active {
		background: var(--primary);
		color: white;
		border-color: var(--primary);
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

	.overview-section {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.metrics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 1.5rem;
	}

	.metric-card {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: 1.5rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.metric-header {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 1rem;
		color: var(--text-secondary);
		font-size: 0.9rem;
	}

	.metric-header svg {
		width: 20px;
		height: 20px;
	}

	.metric-value {
		font-size: 2rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.metric-subtitle {
		font-size: 0.9rem;
		color: var(--text-secondary);
	}

	.metric-bar {
		width: 100%;
		height: 8px;
		background: rgba(255, 255, 255, 0.1);
		border-radius: 4px;
		overflow: hidden;
	}

	.metric-progress {
		height: 100%;
		background: var(--primary-gradient);
		border-radius: 4px;
		transition: width 0.3s ease;
	}

	.health-section {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 20px;
		padding: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.health-section.prominent {
		border: 2px solid var(--primary);
		box-shadow: 0 0 20px rgba(102, 126, 234, 0.2);
	}

	.live-indicator {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		background: rgba(16, 185, 129, 0.1);
		border: 1px solid #10b981;
		border-radius: 20px;
		color: #10b981;
		font-size: 0.9rem;
		font-weight: 600;
		margin-bottom: 1rem;
	}

	.live-indicator::before {
		content: '';
		width: 8px;
		height: 8px;
		background: #10b981;
		border-radius: 50%;
		animation: pulse 2s infinite;
	}

	@keyframes pulse {
		0% {
			transform: scale(0.95);
			box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7);
		}
		70% {
			transform: scale(1);
			box-shadow: 0 0 0 10px rgba(16, 185, 129, 0);
		}
		100% {
			transform: scale(0.95);
			box-shadow: 0 0 0 0 rgba(16, 185, 129, 0);
		}
	}

	.webhook-filters {
		display: flex;
		gap: 1rem;
		margin-bottom: 1.5rem;
		align-items: center;
	}

	.webhook-filters label {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		font-size: 0.9rem;
		color: var(--text-secondary);
	}

	.webhook-filters select {
		padding: 0.5rem 0.75rem;
		border: 1px solid rgba(255, 255, 255, 0.2);
		border-radius: 8px;
		background: var(--bg-glass);
		color: var(--text-primary);
		font-size: 0.9rem;
		cursor: pointer;
		min-width: 120px;
	}

	.webhook-filters select:focus {
		outline: none;
		border-color: var(--primary);
	}

	.health-section h2 {
		font-size: 1.5rem;
		font-weight: 600;
		margin: 0 0 1.5rem 0;
		color: var(--text-primary);
	}

	.health-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: 1.5rem;
	}

	.health-card {
		background: var(--bg-glass-dark);
		border-radius: 15px;
		padding: 1.5rem;
		border: 2px solid transparent;
		transition: all 0.3s ease;
	}

	.health-card.status-healthy {
		border-color: #10b981;
	}

	.health-card.status-warning {
		border-color: #f59e0b;
	}

	.health-card.status-critical {
		border-color: #ef4444;
	}

	.health-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
	}

	.subsite-name {
		font-weight: 600;
		font-size: 1.1rem;
		color: var(--text-primary);
	}

	.status-indicator {
		width: 12px;
		height: 12px;
		border-radius: 50%;
	}

	.health-metrics {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.health-metric {
		display: flex;
		justify-content: space-between;
		align-items: center;
		font-size: 0.9rem;
	}

	.metric-label {
		color: var(--text-secondary);
	}

	.metric-value {
		font-weight: 600;
		color: var(--text-primary);
	}

	.webhooks-section,
	.alerts-section {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 20px;
		padding: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.webhooks-section h2,
	.alerts-section h2 {
		font-size: 1.5rem;
		font-weight: 600;
		margin: 0 0 1.5rem 0;
		color: var(--text-primary);
	}

	.webhooks-table {
		background: var(--bg-glass-dark);
		border-radius: 12px;
		overflow: hidden;
	}

	.table-header {
		display: grid;
		grid-template-columns: 1fr 1fr 1fr 1fr 1fr 1fr;
		gap: 1rem;
		padding: 1rem;
		background: rgba(255, 255, 255, 0.05);
		font-weight: 600;
		font-size: 0.9rem;
		color: var(--text-secondary);
	}

	.table-row {
		display: grid;
		grid-template-columns: 1fr 1fr 1fr 1fr 1fr 1fr;
		gap: 1rem;
		padding: 1rem;
		border-bottom: 1px solid rgba(255, 255, 255, 0.05);
		transition: background 0.3s ease;
	}

	.table-row:hover {
		background: rgba(255, 255, 255, 0.05);
	}

	.table-cell {
		font-size: 0.9rem;
		color: var(--text-primary);
	}

	.status-badge {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.25rem 0.75rem;
		border-radius: 20px;
		font-size: 0.8rem;
		font-weight: 600;
		width: fit-content;
	}

	.status-badge.status-success {
		background: rgba(16, 185, 129, 0.1);
		color: #10b981;
	}

	.status-badge.status-failed {
		background: rgba(239, 68, 68, 0.1);
		color: #ef4444;
	}

	.status-badge.status-pending {
		background: rgba(245, 158, 11, 0.1);
		color: #f59e0b;
	}

	.status-icon {
		width: 16px;
		height: 16px;
	}

	.alerts-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.alert-item {
		background: var(--bg-glass-dark);
		border-radius: 12px;
		padding: 1.5rem;
		border-left: 4px solid transparent;
		transition: all 0.3s ease;
	}

	.alert-item.severity-info {
		border-left-color: #3b82f6;
	}

	.alert-item.severity-warning {
		border-left-color: #f59e0b;
	}

	.alert-item.severity-critical {
		border-left-color: #ef4444;
	}

	.alert-item.acknowledged {
		opacity: 0.7;
	}

	.alert-header {
		display: flex;
		align-items: center;
		gap: 1rem;
		margin-bottom: 0.5rem;
	}

	.alert-severity {
		width: 8px;
		height: 8px;
		border-radius: 50%;
	}

	.alert-title {
		font-weight: 600;
		color: var(--text-primary);
		flex: 1;
	}

	.alert-time {
		font-size: 0.8rem;
		color: var(--text-secondary);
	}

	.acknowledge-button {
		padding: 0.25rem 0.75rem;
		background: var(--primary);
		color: white;
		border: none;
		border-radius: 6px;
		font-size: 0.8rem;
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.acknowledge-button:hover {
		background: var(--primary-dark);
	}

	.alert-message {
		color: var(--text-secondary);
		margin-bottom: 0.5rem;
	}

	.alert-subsite {
		font-size: 0.8rem;
		color: var(--primary);
		font-weight: 600;
	}

	@media (max-width: 768px) {
		.monitoring-dashboard {
			padding: 1rem;
		}

		.header-content {
			flex-direction: column;
			align-items: stretch;
		}

		.header-actions {
			justify-content: center;
		}

		.metrics-grid {
			grid-template-columns: 1fr;
		}

		.health-grid {
			grid-template-columns: 1fr;
		}

		.table-header,
		.table-row {
			grid-template-columns: 1fr;
			gap: 0.5rem;
		}

		.table-header {
			display: none;
		}

		.table-row {
			padding: 1rem;
			border: 1px solid rgba(255, 255, 255, 0.1);
			border-radius: 8px;
			margin-bottom: 1rem;
		}

		.table-cell {
			display: flex;
			justify-content: space-between;
			align-items: center;
		}

		.table-cell::before {
			content: attr(data-label);
			font-weight: 600;
			color: var(--text-secondary);
		}
	}
</style> 