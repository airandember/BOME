<script>
	import { onMount, onDestroy } from 'svelte';
	import { apiClient } from '$lib/api';

	let connectionPoolStats = null;
	let databaseHealth = null;
	let loading = true;
	let error = null;
	let refreshInterval = null;

	// Auto-refresh every 10 seconds
	const REFRESH_INTERVAL = 10000;

	onMount(() => {
		loadMonitoringData();
		startAutoRefresh();
	});

	onDestroy(() => {
		if (refreshInterval) {
			clearInterval(refreshInterval);
		}
	});

	function startAutoRefresh() {
		refreshInterval = setInterval(() => {
			loadMonitoringData();
		}, REFRESH_INTERVAL);
	}

	async function loadMonitoringData() {
		try {
			// Load both connection pool stats and full health check
			const [poolResponse, healthResponse] = await Promise.all([
				apiClient.get('/admin/monitoring/db/pool'),
				apiClient.get('/admin/monitoring/db/health')
			]);

			connectionPoolStats = poolResponse.data;
			databaseHealth = healthResponse.data;
			error = null;
		} catch (err) {
			console.error('Failed to load monitoring data:', err);
			error = err.message || 'Failed to load monitoring data';
		} finally {
			loading = false;
		}
	}

	function getHealthStatusColor(status) {
		if (status?.includes('🟢')) return 'text-green-600';
		if (status?.includes('🟡')) return 'text-yellow-600';
		if (status?.includes('🟠')) return 'text-orange-600';
		if (status?.includes('🔴')) return 'text-red-600';
		return 'text-gray-600';
	}

	function getUtilizationColor(percentage) {
		if (percentage > 90) return 'bg-red-500';
		if (percentage > 75) return 'bg-yellow-500';
		if (percentage > 50) return 'bg-orange-500';
		return 'bg-green-500';
	}

	function formatDuration(duration) {
		if (!duration) return 'N/A';
		if (duration < 1000000) return `${Math.round(duration / 1000)}μs`;
		if (duration < 1000000000) return `${Math.round(duration / 1000000)}ms`;
		return `${Math.round(duration / 1000000000)}s`;
	}
</script>

<div class="p-6 max-w-7xl mx-auto">
	<div class="mb-6">
		<h1 class="text-3xl font-bold text-gray-900">Database Monitoring</h1>
		<p class="text-gray-600 mt-2">Real-time connection pool and database health monitoring</p>
	</div>

	{#if loading && !connectionPoolStats}
		<div class="flex justify-center items-center h-64">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
		</div>
	{:else if error}
		<div class="bg-red-50 border border-red-200 rounded-lg p-4">
			<div class="flex">
				<div class="flex-shrink-0">
					<svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
					</svg>
				</div>
				<div class="ml-3">
					<h3 class="text-sm font-medium text-red-800">Error Loading Monitoring Data</h3>
					<p class="text-sm text-red-700 mt-1">{error}</p>
				</div>
			</div>
		</div>
	{:else}
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
			<!-- Connection Pool Stats -->
			{#if connectionPoolStats}
				<div class="bg-white rounded-lg shadow-md p-6">
					<div class="flex items-center justify-between mb-4">
						<h2 class="text-xl font-semibold text-gray-900">Connection Pool</h2>
						<span class="text-sm text-gray-500">
							Updated: {new Date(connectionPoolStats.timestamp).toLocaleTimeString()}
						</span>
					</div>

					<div class="space-y-4">
						<!-- Health Status -->
						<div class="flex items-center justify-between">
							<span class="text-sm font-medium text-gray-700">Health Status</span>
							<span class="text-sm font-medium {getHealthStatusColor(connectionPoolStats.data.health_status)}">
								{connectionPoolStats.data.health_status}
							</span>
						</div>

						<!-- Utilization Bar -->
						<div>
							<div class="flex justify-between text-sm text-gray-700 mb-1">
								<span>Utilization</span>
								<span>{connectionPoolStats.data.utilization_percentage.toFixed(1)}%</span>
							</div>
							<div class="w-full bg-gray-200 rounded-full h-2">
								<div 
									class="h-2 rounded-full {getUtilizationColor(connectionPoolStats.data.utilization_percentage)}"
									style="width: {Math.min(connectionPoolStats.data.utilization_percentage, 100)}%"
								></div>
							</div>
						</div>

						<!-- Connection Stats Grid -->
						<div class="grid grid-cols-2 gap-4">
							<div class="bg-gray-50 rounded-lg p-3">
								<div class="text-2xl font-bold text-blue-600">{connectionPoolStats.data.open_connections}</div>
								<div class="text-sm text-gray-600">Open Connections</div>
								<div class="text-xs text-gray-500">Max: {connectionPoolStats.data.max_open_connections}</div>
							</div>
							<div class="bg-gray-50 rounded-lg p-3">
								<div class="text-2xl font-bold text-green-600">{connectionPoolStats.data.in_use}</div>
								<div class="text-sm text-gray-600">In Use</div>
							</div>
							<div class="bg-gray-50 rounded-lg p-3">
								<div class="text-2xl font-bold text-yellow-600">{connectionPoolStats.data.idle}</div>
								<div class="text-sm text-gray-600">Idle</div>
							</div>
							<div class="bg-gray-50 rounded-lg p-3">
								<div class="text-2xl font-bold text-red-600">{connectionPoolStats.data.wait_count}</div>
								<div class="text-sm text-gray-600">Wait Count</div>
							</div>
						</div>

						<!-- Additional Stats -->
						<div class="border-t pt-4 space-y-2">
							<div class="flex justify-between text-sm">
								<span class="text-gray-600">Wait Duration</span>
								<span class="font-medium">{formatDuration(connectionPoolStats.data.wait_duration)}</span>
							</div>
							<div class="flex justify-between text-sm">
								<span class="text-gray-600">Max Idle Closed</span>
								<span class="font-medium">{connectionPoolStats.data.max_idle_closed}</span>
							</div>
							<div class="flex justify-between text-sm">
								<span class="text-gray-600">Max Lifetime Closed</span>
								<span class="font-medium">{connectionPoolStats.data.max_lifetime_closed}</span>
							</div>
						</div>
					</div>
				</div>
			{/if}

			<!-- Database Health -->
			{#if databaseHealth}
				<div class="bg-white rounded-lg shadow-md p-6">
					<div class="flex items-center justify-between mb-4">
						<h2 class="text-xl font-semibold text-gray-900">PostgreSQL Health</h2>
						<span class="text-sm text-gray-500">
							Response: {formatDuration(databaseHealth.data.postgresql_health.response_time)}
						</span>
					</div>

					<div class="space-y-4">
						<!-- Server Status -->
						<div class="flex items-center justify-between">
							<span class="text-sm font-medium text-gray-700">Server Status</span>
							<span class="text-sm font-medium {databaseHealth.data.postgresql_health.status === 'healthy' ? 'text-green-600' : 'text-red-600'}">
								{databaseHealth.data.postgresql_health.status.toUpperCase()}
							</span>
						</div>

						<!-- Server Stats -->
						{#if databaseHealth.data.postgresql_health.active_connections !== undefined}
							<div class="grid grid-cols-2 gap-4">
								<div class="bg-gray-50 rounded-lg p-3">
									<div class="text-2xl font-bold text-blue-600">{databaseHealth.data.postgresql_health.active_connections}</div>
									<div class="text-sm text-gray-600">Active Connections</div>
								</div>
								{#if databaseHealth.data.postgresql_health.max_connections}
									<div class="bg-gray-50 rounded-lg p-3">
										<div class="text-2xl font-bold text-purple-600">{databaseHealth.data.postgresql_health.max_connections}</div>
										<div class="text-sm text-gray-600">Max Connections</div>
									</div>
								{/if}
							</div>
						{/if}

						{#if databaseHealth.data.postgresql_health.database_size}
							<div class="flex justify-between text-sm">
								<span class="text-gray-600">Database Size</span>
								<span class="font-medium">{databaseHealth.data.postgresql_health.database_size}</span>
							</div>
						{/if}

						<!-- Recommendations -->
						{#if databaseHealth.data.recommendations && databaseHealth.data.recommendations.length > 0}
							<div class="border-t pt-4">
								<h3 class="text-sm font-medium text-gray-900 mb-2">Recommendations</h3>
								<div class="space-y-2">
									{#each databaseHealth.data.recommendations as recommendation}
										<div class="text-sm p-2 rounded {recommendation.includes('🔴') ? 'bg-red-50 text-red-800' : recommendation.includes('🟡') ? 'bg-yellow-50 text-yellow-800' : recommendation.includes('🟠') ? 'bg-orange-50 text-orange-800' : 'bg-green-50 text-green-800'}">
											{recommendation}
										</div>
									{/each}
								</div>
							</div>
						{/if}
					</div>
				</div>
			{/if}
		</div>

		<!-- Refresh Controls -->
		<div class="mt-6 flex justify-between items-center">
			<div class="text-sm text-gray-500">
				Auto-refreshing every {REFRESH_INTERVAL / 1000} seconds
			</div>
			<button
				onclick={loadMonitoringData}
				class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
				disabled={loading}
			>
				{loading ? 'Refreshing...' : 'Refresh Now'}
			</button>
		</div>
	{/if}
</div>

<style>
	/* Add any additional custom styles here */
</style>
