<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';

	// Props from parent
	const { summary = null, stripeData = null } = $props<{ 
		summary?: any, 
		stripeData?: any 
	}>();

	// State for analytics data
	let analyticsData = $state<Record<string, any>>({
		balance: null,
		charges: null,
		customers: null,
		subscriptions: null,
		products: null,
		lastUpdated: null
	});

	// Performance tracking
	let performanceMetrics = $state({
		startTime: 0,
		endTime: 0,
		totalDuration: 0,
		endpointTimings: new Map<string, number>(),
		fastestEndpoint: '',
		slowestEndpoint: '',
		successCount: 0,
		errorCount: 0
	});

	// Loading states
	let isLoading = $state(false);
	let isRefreshing = $state(false);
	let error = $state('');

	// Endpoint configuration
	const endpoints = [
		{ 
			id: 'balance', 
			name: 'Account Balance', 
			endpoint: '/admin/streaming/stripe/balance',
			icon: '💰',
			description: 'Account balance and transaction history'
		},
		{ 
			id: 'charges', 
			name: 'Payment Charges', 
			endpoint: '/admin/streaming/stripe/charges',
			icon: '💳',
			description: 'Total charge counts and summaries'
		},
		{ 
			id: 'customers', 
			name: 'Customer Counts', 
			endpoint: '/admin/streaming/stripe/customers',
			icon: '👥',
			description: 'Total customer counts and metrics'
		},
		{ 
			id: 'subscriptions', 
			name: 'Subscription Stats', 
			endpoint: '/admin/streaming/stripe/subscriptions',
			icon: '📋',
			description: 'Active/inactive subscription counts'
		},
		{ 
			id: 'products', 
			name: 'Product Analytics', 
			endpoint: '/admin/streaming/stripe/products',
			icon: '📦',
			description: 'Product counts and revenue metrics'
		}
	];

	onMount(() => {
		loadAnalyticsData();
	});

	// Load all analytics data in parallel
	async function loadAnalyticsData() {
		if (isLoading) return;
		
		try {
			isLoading = true;
			error = '';
			performanceMetrics.startTime = Date.now();
			performanceMetrics.endpointTimings.clear();
			performanceMetrics.successCount = 0;
			performanceMetrics.errorCount = 0;

			console.log('🚀 Starting parallel analytics data fetch...');

			// Create all API calls in parallel
			const apiCalls = endpoints.map(async (endpoint) => {
				const startTime = Date.now();
				const endpointId = endpoint.id;
				
				try {
					console.log(`📡 Fetching ${endpoint.name}...`);
					
					const response = await apiRequest(endpoint.endpoint);
					const duration = Date.now() - startTime;
					
					if (response.ok) {
						const data = await response.json();
						analyticsData[endpointId] = data;
						performanceMetrics.endpointTimings.set(endpointId, duration);
						performanceMetrics.successCount++;
						
						console.log(`✅ ${endpoint.name} loaded in ${duration}ms`);
					} else {
						throw new Error(`HTTP ${response.status}: ${response.statusText}`);
					}
				} catch (err: any) {
					const duration = Date.now() - startTime;
					performanceMetrics.endpointTimings.set(endpointId, duration);
					performanceMetrics.errorCount++;
					analyticsData[endpointId] = { error: err.message || 'Unknown error' };
					
					console.error(`❌ ${endpoint.name} failed in ${duration}ms:`, err);
				}
			});

			// Wait for all calls to complete
			await Promise.allSettled(apiCalls);

			// Calculate performance metrics
			performanceMetrics.endTime = Date.now();
			performanceMetrics.totalDuration = performanceMetrics.endTime - performanceMetrics.startTime;
			
			// Find fastest and slowest endpoints
			let fastestTime = Infinity;
			let slowestTime = 0;
			
			performanceMetrics.endpointTimings.forEach((time, endpointId) => {
				if (time < fastestTime) {
					fastestTime = time;
					performanceMetrics.fastestEndpoint = endpointId;
				}
				if (time > slowestTime) {
					slowestTime = time;
					performanceMetrics.slowestEndpoint = endpointId;
				}
			});

			analyticsData.lastUpdated = new Date();
			
			console.log('🎉 Analytics data load complete!', {
				totalDuration: performanceMetrics.totalDuration,
				successCount: performanceMetrics.successCount,
				errorCount: performanceMetrics.errorCount,
				fastest: performanceMetrics.fastestEndpoint,
				slowest: performanceMetrics.slowestEndpoint
			});

		} catch (err: any) {
			error = 'Failed to load analytics data: ' + (err.message || 'Unknown error');
			console.error('❌ Analytics load failed:', err);
		} finally {
			isLoading = false;
		}
	}

	// Refresh data
	async function refreshData() {
		if (isRefreshing) return;
		
		isRefreshing = true;
		await loadAnalyticsData();
		isRefreshing = false;
	}

	// Format duration for display
	function formatDuration(ms: number): string {
		if (ms < 1000) return `${ms}ms`;
		return `${(ms / 1000).toFixed(1)}s`;
	}

	// Get endpoint status
	function getEndpointStatus(endpointId: string) {
		const data = analyticsData[endpointId];
		const timing = performanceMetrics.endpointTimings.get(endpointId);
		
		if (!data) return 'pending';
		if (data.error) return 'error';
		if (timing !== undefined) return 'success';
		return 'pending';
	}

	// Get endpoint timing
	function getEndpointTiming(endpointId: string): string {
		const timing = performanceMetrics.endpointTimings.get(endpointId);
		return timing ? formatDuration(timing) : 'Pending...';
	}

	// Get endpoint data summary
	function getEndpointSummary(endpointId: string): string {
		const data = analyticsData[endpointId];
		if (!data) return 'No data';
		if (data.error) return `Error: ${data.error}`;
		
		switch (endpointId) {
			case 'balance':
				return data.available ? `$${(data.available[0]?.amount / 100).toFixed(2)} available` : 'No balance data';
			case 'charges':
				return data.total_count ? `${data.total_count.toLocaleString()} charges` : 'No charge data';
			case 'customers':
				return data.total_count ? `${data.total_count.toLocaleString()} customers` : 'No customer data';
			case 'subscriptions':
				return data.total_count ? `${data.total_count.toLocaleString()} subscriptions` : 'No subscription data';
			case 'products':
				return data.total_count ? `${data.total_count.toLocaleString()} products` : 'No product data';
			default:
				return 'Data loaded';
		}
	}
</script>

<div class="analytics-overview">
	<!-- Header -->
	<div class="overview-header">
		<div class="header-content">
			<h1>📊 Stripe Analytics Dashboard</h1>
			<p>Real-time performance metrics and analytics data from Stripe</p>
		</div>
		
		<div class="header-actions">
			<button 
				class="btn btn-primary" 
				onclick={refreshData}
				disabled={isLoading || isRefreshing}
			>
				{isRefreshing ? '🔄 Refreshing...' : '🔄 Refresh Data'}
			</button>
		</div>
	</div>

	<!-- Performance Overview -->
	<div class="performance-overview">
		<div class="performance-header">
			<h2>⚡ Performance Metrics</h2>
			{#if analyticsData.lastUpdated}
				<span class="last-updated">Last updated: {analyticsData.lastUpdated.toLocaleTimeString()}</span>
			{/if}
		</div>
		
		<div class="performance-grid">
			<div class="metric-card">
				<div class="metric-value">{formatDuration(performanceMetrics.totalDuration)}</div>
				<div class="metric-label">Total Load Time</div>
			</div>
			
			<div class="metric-card success">
				<div class="metric-value">{performanceMetrics.successCount}</div>
				<div class="metric-label">Successful Endpoints</div>
			</div>
			
			<div class="metric-card error">
				<div class="metric-value">{performanceMetrics.errorCount}</div>
				<div class="metric-label">Failed Endpoints</div>
			</div>
			
			<div class="metric-card fastest">
				<div class="metric-value">
					{performanceMetrics.fastestEndpoint ? 
						endpoints.find(e => e.id === performanceMetrics.fastestEndpoint)?.icon : '⚡'}
				</div>
				<div class="metric-label">Fastest Endpoint</div>
			</div>
			
			<div class="metric-card slowest">
				<div class="metric-value">
					{performanceMetrics.slowestEndpoint ? 
						endpoints.find(e => e.id === performanceMetrics.slowestEndpoint)?.icon : '🐌'}
				</div>
				<div class="metric-label">Slowest Endpoint</div>
			</div>
		</div>
	</div>

	<!-- Endpoints Grid -->
	<div class="endpoints-grid">
		{#each endpoints as endpoint}
			{@const status = getEndpointStatus(endpoint.id)}
			{@const timing = getEndpointTiming(endpoint.id)}
			{@const summary = getEndpointSummary(endpoint.id)}
			
			<div class="endpoint-card {status}">
				<div class="endpoint-header">
					<div class="endpoint-icon">{endpoint.icon}</div>
					<div class="endpoint-info">
						<h3 class="endpoint-name">{endpoint.name}</h3>
						<p class="endpoint-description">{endpoint.description}</p>
					</div>
					<div class="endpoint-status">
						{#if status === 'pending'}
							<span class="status pending">⏳</span>
						{:else if status === 'success'}
							<span class="status success">✅</span>
						{:else if status === 'error'}
							<span class="status error">❌</span>
						{/if}
					</div>
				</div>
				
				<div class="endpoint-content">
					<div class="endpoint-timing">
						<strong>Response Time:</strong> {timing}
					</div>
					
					<div class="endpoint-summary">
						<strong>Data:</strong> {summary}
					</div>
					
					{#if status === 'error'}
						<div class="endpoint-error">
							<strong>Error:</strong> {analyticsData[endpoint.id]?.error || 'Unknown error'}
						</div>
					{/if}
				</div>
			</div>
		{/each}
	</div>

	<!-- Loading State -->
	{#if isLoading}
		<div class="loading-overlay">
			<div class="loading-content">
				<div class="loading-spinner"></div>
				<h3>Loading Analytics Data...</h3>
				<p>Testing all Stripe endpoints simultaneously</p>
				<div class="loading-progress">
					{performanceMetrics.successCount + performanceMetrics.errorCount} of {endpoints.length} endpoints processed
				</div>
			</div>
		</div>
	{/if}

	<!-- Error State -->
	{#if error}
		<div class="error-banner">
			<div class="error-content">
				<span class="error-icon">⚠️</span>
				<span class="error-message">{error}</span>
				<button class="btn btn-secondary" onclick={refreshData}>Retry</button>
			</div>
		</div>
	{/if}
</div>

<style>
	.analytics-overview {
		padding: var(--space-lg, 1.5rem);
		max-width: 1400px;
		margin: 0 auto;
	}

	.overview-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: var(--space-xl, 2rem);
		flex-wrap: wrap;
		gap: var(--space-lg, 1.5rem);
	}

	.header-content h1 {
		margin: 0 0 var(--space-xs, 0.5rem) 0;
		color: var(--text, #111827);
		font-size: 2rem;
		font-weight: 700;
	}

	.header-content p {
		margin: 0;
		color: var(--text-muted, #6b7280);
		font-size: 1.1rem;
	}

	.header-actions {
		display: flex;
		gap: var(--space-md, 1rem);
	}

	.btn {
		padding: var(--space-sm, 0.75rem) var(--space-lg, 1.5rem);
		border: none;
		border-radius: var(--radius-md, 0.375rem);
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		gap: var(--space-xs, 0.5rem);
	}

	.btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-primary {
		background: #2563eb;
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: #1d4ed8;
		transform: translateY(-1px);
	}

	.btn-secondary {
		background: #f3f4f6;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-secondary:hover:not(:disabled) {
		background: #e5e7eb;
	}

	/* Performance Overview */
	.performance-overview {
		background: var(--surface, white);
		border: 1px solid var(--border, #e5e7eb);
		border-radius: var(--radius-lg, 0.5rem);
		padding: var(--space-lg, 1.5rem);
		margin-bottom: var(--space-xl, 2rem);
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
	}

	.performance-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-lg, 1.5rem);
		flex-wrap: wrap;
		gap: var(--space-md, 1rem);
	}

	.performance-header h2 {
		margin: 0;
		color: var(--text, #111827);
		font-size: 1.5rem;
		font-weight: 600;
	}

	.last-updated {
		color: var(--text-muted, #6b7280);
		font-size: 0.875rem;
	}

	.performance-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: var(--space-md, 1rem);
	}

	.metric-card {
		background: var(--bg-secondary, #f9fafb);
		border: 1px solid var(--border, #e5e7eb);
		border-radius: var(--radius-md, 0.375rem);
		padding: var(--space-md, 1rem);
		text-align: center;
		transition: all 0.2s ease;
	}

	.metric-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}

	.metric-card.success {
		border-color: #059669;
		background: #f0fdf4;
	}

	.metric-card.error {
		border-color: #dc2626;
		background: #fef2f2;
	}

	.metric-card.fastest {
		border-color: #059669;
		background: #f0fdf4;
	}

	.metric-card.slowest {
		border-color: #d97706;
		background: #fffbeb;
	}

	.metric-value {
		font-size: 1.5rem;
		font-weight: bold;
		color: var(--primary, #2563eb);
		margin-bottom: var(--space-xs, 0.5rem);
	}

	.metric-label {
		font-size: 0.875rem;
		color: var(--text-muted, #6b7280);
		font-weight: 500;
	}

	/* Endpoints Grid */
	.endpoints-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
		gap: var(--space-lg, 1.5rem);
	}

	.endpoint-card {
		background: var(--surface, white);
		border: 1px solid var(--border, #e5e7eb);
		border-radius: var(--radius-lg, 0.5rem);
		padding: var(--space-lg, 1.5rem);
		transition: all 0.2s ease;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
	}

	.endpoint-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
	}

	.endpoint-card.success {
		border-left: 4px solid #059669;
	}

	.endpoint-card.error {
		border-left: 4px solid #dc2626;
	}

	.endpoint-card.pending {
		border-left: 4px solid #6b7280;
	}

	.endpoint-header {
		display: flex;
		align-items: flex-start;
		gap: var(--space-md, 1rem);
		margin-bottom: var(--space-md, 1rem);
	}

	.endpoint-icon {
		font-size: 2rem;
		line-height: 1;
	}

	.endpoint-info {
		flex: 1;
	}

	.endpoint-name {
		margin: 0 0 var(--space-xs, 0.5rem) 0;
		color: var(--text, #111827);
		font-size: 1.25rem;
		font-weight: 600;
	}

	.endpoint-description {
		margin: 0;
		color: var(--text-muted, #6b7280);
		font-size: 0.875rem;
		line-height: 1.4;
	}

	.endpoint-status {
		flex-shrink: 0;
	}

	.status {
		font-size: 1.5rem;
		line-height: 1;
	}

	.status.success {
		color: #059669;
	}

	.status.error {
		color: #dc2626;
	}

	.status.pending {
		color: #6b7280;
	}

	.endpoint-content {
		border-top: 1px solid var(--border, #e5e7eb);
		padding-top: var(--space-md, 1rem);
	}

	.endpoint-timing,
	.endpoint-summary {
		margin-bottom: var(--space-sm, 0.5rem);
		font-size: 0.875rem;
	}

	.endpoint-timing strong,
	.endpoint-summary strong {
		color: var(--text, #111827);
	}

	.endpoint-error {
		margin-top: var(--space-sm, 0.5rem);
		padding: var(--space-sm, 0.5rem);
		background: #fef2f2;
		border: 1px solid #fecaca;
		border-radius: var(--radius-sm, 0.25rem);
		font-size: 0.875rem;
		color: #dc2626;
	}

	/* Loading Overlay */
	.loading-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.8);
		display: flex;
		justify-content: center;
		align-items: center;
		z-index: 1000;
	}

	.loading-content {
		background: var(--surface, white);
		border-radius: var(--radius-lg, 0.5rem);
		padding: var(--space-xl, 2rem);
		text-align: center;
		max-width: 400px;
		width: 90%;
	}

	.loading-spinner {
		width: 60px;
		height: 60px;
		border: 4px solid #e5e7eb;
		border-top: 4px solid #2563eb;
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin: 0 auto var(--space-lg, 1.5rem) auto;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.loading-content h3 {
		margin: 0 0 var(--space-sm, 0.5rem) 0;
		color: var(--text, #111827);
		font-size: 1.5rem;
		font-weight: 600;
	}

	.loading-content p {
		margin: 0 0 var(--space-lg, 1.5rem) 0;
		color: var(--text-muted, #6b7280);
	}

	.loading-progress {
		padding: var(--space-sm, 0.5rem) var(--space-md, 1rem);
		background: var(--bg-secondary, #f9fafb);
		border-radius: var(--radius-md, 0.375rem);
		font-size: 0.875rem;
		color: var(--text-muted, #6b7280);
		font-weight: 500;
	}

	/* Error Banner */
	.error-banner {
		background: #fef2f2;
		border: 1px solid #fecaca;
		border-radius: var(--radius-md, 0.375rem);
		padding: var(--space-md, 1rem);
		margin-bottom: var(--space-lg, 1.5rem);
	}

	.error-content {
		display: flex;
		align-items: center;
		gap: var(--space-md, 1rem);
		flex-wrap: wrap;
	}

	.error-icon {
		font-size: 1.25rem;
	}

	.error-message {
		flex: 1;
		color: #dc2626;
		font-weight: 500;
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.analytics-overview {
			padding: var(--space-md, 1rem);
		}

		.overview-header {
			flex-direction: column;
			align-items: stretch;
			text-align: center;
		}

		.header-actions {
			justify-content: center;
		}

		.performance-header {
			flex-direction: column;
			text-align: center;
		}

		.performance-grid {
			grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		}

		.endpoints-grid {
			grid-template-columns: 1fr;
		}

		.endpoint-header {
			flex-direction: column;
			text-align: center;
			gap: var(--space-sm, 0.5rem);
		}

		.endpoint-status {
			align-self: center;
		}
	}
</style>
