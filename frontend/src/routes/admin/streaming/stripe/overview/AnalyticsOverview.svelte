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

	// V2 Analytics state
	let v2Analytics = $state<any>(null)
	let v2Loading = $state(false)
	let v2LastUpdated = $state<Date | null>(null)
	let v2FetchTime = $state<string>('')



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
		loadV2Analytics();
	});

	// Load v2 analytics data
	async function loadV2Analytics() {
		if (v2Loading) return
		
		v2Loading = true
		console.log("🚀 Loading v2 analytics...")
		
		try {
			const response = await apiRequest('/admin/streaming/stripe/v2/analytics')
			const data = await response.json()
			
			v2Analytics = data
			v2LastUpdated = new Date()
			v2FetchTime = data.total_fetch_time || '-'
			
			console.log("✅ V2 analytics loaded:", data)
		} catch (error) {
			console.error("❌ Failed to load v2 analytics:", error)
		} finally {
			v2Loading = false
		}
	}

	// Test v1 vs v2 performance
	async function testV1vsV2() {
		console.log("🏁 Testing v1 vs v2 performance...")
		
		try {
			// Test v1 endpoint
			console.log("📈 Testing v1 /stripe/analytics...")
			const startV1 = performance.now()
			const v1Response = await apiRequest('/admin/streaming/stripe/analytics')
			const v1Data = await v1Response.json()
			const v1Duration = performance.now() - startV1
			
			// Test v2 endpoint
			console.log("📈 Testing v2 /stripe/v2/analytics...")
			const startV2 = performance.now()
			const v2Response = await apiRequest('/admin/streaming/stripe/v2/analytics')
			const v2Data = await v2Response.json()
			const v2Duration = performance.now() - startV2
			
			console.log("🏁 Performance Results:")
			console.log(`   v1: ${v1Duration.toFixed(0)}ms - ${v1Data.total_fetch_time}`)
			console.log(`   v2: ${v2Duration.toFixed(0)}ms - ${v2Data.total_fetch_time}`)
			console.log(`   Speedup: ${(v1Duration/v2Duration).toFixed(1)}x faster`)
			
			// 📊 LOG FULL JSON RESPONSES FOR ANALYSIS
			console.log("📊 === V1 FULL JSON RESPONSE ===")
			console.log(JSON.stringify(v1Data, null, 2))
			
			console.log("📊 === V2 FULL JSON RESPONSE ===")
			console.log(JSON.stringify(v2Data, null, 2))
			
		} catch (error) {
			console.error("❌ v1 vs v2 test failed:", error)
		}
	}

	// System Management




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
				onclick={loadV2Analytics}
				disabled={v2Loading}
			>
				{v2Loading ? '🔄 Loading...' : '📈 V2 Analytics'}
			</button>
		</div>
	</div>

	<!-- Performance Overview  -->

	<div class="performance-overview">
		<div class="performance-header">
			<h2>⚡ V2 Analytics Performance</h2>
			{#if v2LastUpdated}
				<span class="last-updated">Last updated: {v2LastUpdated.toLocaleTimeString()}</span>
			{/if}
		</div>
		
		<div class="performance-grid">
			<div class="metric-card">
				<div class="metric-value">{v2FetchTime}</div>
				<div class="metric-label">V2 Load Time</div>
			</div>
			
			<div class="metric-card success">
				<div class="metric-value">{v2Analytics?.enabled ? '8' : '0'}</div>
				<div class="metric-label">V2 Endpoints</div>
			</div>
			
			<div class="metric-card error">
				<div class="metric-value">0</div>
				<div class="metric-label">Failed Endpoints</div>
			</div>
			
			<div class="metric-card fastest">
				<div class="metric-value">⚡</div>
				<div class="metric-label">V2 Analytics</div>
			</div>
			
			<div class="metric-card slowest">
				<div class="metric-value">🚀</div>
				<div class="metric-label">Optimized</div>
			</div>
		</div>
	</div>

	<!-- V2 Analytics Summary Cards -->
	<div class="analytics-summary">
		<div class="summary-grid">
			<!-- Balance Card -->
			<div class="summary-card balance-card">
				<div class="card-icon">💰</div>
				<div class="card-content">
					<div class="card-value">{v2Analytics?.balance?.available_usd || '$0.00'}</div>
					<div class="card-label">Account Balance</div>
				</div>
			</div>

			<!-- Customers Card -->
			<div class="summary-card customer-card">
				<div class="card-icon">👥</div>
				<div class="card-content">
					<div class="card-value">{v2Analytics?.customer_analytics?.total_customers || 0}</div>
					<div class="card-label">Total Customers</div>
					<div class="card-subtitle">{v2Analytics?.customer_analytics?.growth_rate_30d || '0.0%'} growth (30d)</div>
				</div>
			</div>

			<!-- MRR Card -->
			<div class="summary-card mrr-card">
				<div class="card-icon">📈</div>
				<div class="card-content">
					<div class="card-value">{v2Analytics?.mrr_analytics?.actual_mrr || '$0.00'}</div>
					<div class="card-label">Monthly Recurring Revenue</div>
					<div class="card-subtitle">{v2Analytics?.mrr_analytics?.actual_arr || '$0.00'} ARR</div>
				</div>
			</div>

			<!-- Subscriptions Card -->
			<div class="summary-card subscription-card">
				<div class="card-icon">📋</div>
				<div class="card-content">
					<div class="card-value">{v2Analytics?.subscription_health?.active_subscriptions || 0}</div>
					<div class="card-label">Active Subscriptions</div>
					<div class="card-subtitle">{v2Analytics?.subscription_health?.churn_rate || '0.0%'} churn rate</div>
				</div>
			</div>

			<!-- Revenue Card -->
			<div class="summary-card revenue-card">
				<div class="card-icon">💳</div>
				<div class="card-content">
					<div class="card-value">{v2Analytics?.revenue_analytics?.recent_revenue || '$0.00'}</div>
					<div class="card-label">Recent Revenue</div>
					<div class="card-subtitle">{v2Analytics?.payment_analytics?.success_rate || '0.0%'} success rate</div>
				</div>
			</div>

			<!-- Products Card -->
			<div class="summary-card product-card">
				<div class="card-icon">📦</div>
				<div class="card-content">
					<div class="card-value">{v2Analytics?.product_performance?.total_products || 0}</div>
					<div class="card-label">Total Products</div>
					<div class="card-subtitle">{v2Analytics?.product_performance?.top_product || 'N/A'} top seller</div>
				</div>
			</div>
		</div>
	</div>

	<!-- Subscription Diagnostics Section -->
	{#if v2Analytics?.subscription_diagnostics}
		<div class="diagnostics-section">
			<div class="diagnostics-header">
				<h2>🔍 Subscription Diagnostics</h2>
				<p>Investigating the discrepancy between subscription count (496) and unique customers (~70)</p>
			</div>
			
			<div class="diagnostics-grid">
				<div class="diagnostic-card">
					<div class="diagnostic-icon">📊</div>
					<div class="diagnostic-content">
						<div class="diagnostic-value">{v2Analytics.subscription_diagnostics.total_active_subscriptions}</div>
						<div class="diagnostic-label">Total Active Subscriptions</div>
					</div>
				</div>
				
				<div class="diagnostic-card highlight">
					<div class="diagnostic-icon">👥</div>
					<div class="diagnostic-content">
						<div class="diagnostic-value">{v2Analytics.subscription_diagnostics.total_active_customers}</div>
						<div class="diagnostic-label">Active Customers</div>
						<div class="diagnostic-subtitle">This is your real active user count!</div>
					</div>
				</div>
				
				<div class="diagnostic-card">
					<div class="diagnostic-icon">📈</div>
					<div class="diagnostic-content">
						<div class="diagnostic-value">{v2Analytics.subscription_diagnostics.avg_subscriptions_per_customer}</div>
						<div class="diagnostic-label">Avg Subs per Customer</div>
					</div>
				</div>
				
				<div class="diagnostic-card">
					<div class="diagnostic-icon">🆓</div>
					<div class="diagnostic-content">
						<div class="diagnostic-value">{v2Analytics.subscription_diagnostics.trial_subscriptions}</div>
						<div class="diagnostic-label">Trial Subscriptions</div>
					</div>
				</div>
				
				<div class="diagnostic-card">
					<div class="diagnostic-icon">⏰</div>
					<div class="diagnostic-content">
						<div class="diagnostic-value">{v2Analytics.subscription_diagnostics.canceled_but_active_until_period_end}</div>
						<div class="diagnostic-label">Canceled but Active</div>
						<div class="diagnostic-subtitle">Until period end</div>
					</div>
				</div>
				
				<div class="diagnostic-card">
					<div class="diagnostic-icon">❌</div>
					<div class="diagnostic-content">
						<div class="diagnostic-value">{v2Analytics.subscription_diagnostics.recently_canceled_30_days}</div>
						<div class="diagnostic-label">Recently Canceled</div>
						<div class="diagnostic-subtitle">Last 30 days</div>
					</div>
				</div>
			</div>
			
			<div class="diagnostics-explanation">
				<div class="explanation-card">
					<h3>💡 Explanation</h3>
					<p><strong>Why 496 subscriptions ≠ 70 customers:</strong></p>
					<ul>
						<li><strong>Multiple subscriptions per customer:</strong> Some customers have multiple active subscriptions</li>
						<li><strong>Canceled but active:</strong> Subscriptions canceled but still active until billing period ends</li>
						<li><strong>Trial subscriptions:</strong> May be counted as "active" even if not paying yet</li>
						<li><strong>Incomplete subscriptions:</strong> Failed payments but still technically active</li>
					</ul>
					<p><strong>Your real active customer count is: {v2Analytics.subscription_diagnostics.total_active_customers}</strong></p>
				</div>
			</div>
		</div>
	{/if}



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

	.btn-success {
		background: #059669;
		color: white;
	}

	.btn-success:hover:not(:disabled) {
		background: #047857;
		transform: translateY(-1px);
	}

	.btn-warning {
		background: #d97706;
		color: white;
	}

	.btn-warning:hover:not(:disabled) {
		background: #b45309;
		transform: translateY(-1px);
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

	/* V2 Analytics Summary Styles */
	.analytics-summary {
		margin: 2rem 0;
	}

	.summary-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
		gap: 1.5rem;
	}

	.summary-card {
		background: white;
		border-radius: 12px;
		padding: 1.5rem;
		border: 1px solid #e9ecef;
		display: flex;
		align-items: center;
		gap: 1rem;
		transition: transform 0.2s ease, box-shadow 0.2s ease;
	}

	.summary-card:hover {
		transform: translateY(-3px);
		box-shadow: 0 6px 20px rgba(0, 0, 0, 0.1);
	}

	.card-icon {
		font-size: 2.5rem;
		width: 60px;
		height: 60px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 12px;
		background: #f8f9fa;
	}

	.card-content {
		flex: 1;
	}

	.card-value {
		font-size: 1.8rem;
		font-weight: 700;
		color: #495057;
		margin-bottom: 0.25rem;
	}

	.card-label {
		font-size: 1rem;
		color: #495057;
		font-weight: 500;
		margin-bottom: 0.25rem;
	}

	.card-subtitle {
		font-size: 0.85rem;
		color: #6c757d;
	}

	/* Card-specific colors */
	.balance-card .card-icon {
		background: rgba(40, 167, 69, 0.1);
		color: #28a745;
	}

	.customer-card .card-icon {
		background: rgba(23, 162, 184, 0.1);
		color: #17a2b8;
	}

	.mrr-card .card-icon {
		background: rgba(255, 193, 7, 0.1);
		color: #ffc107;
	}

	.subscription-card .card-icon {
		background: rgba(111, 66, 193, 0.1);
		color: #6f42c1;
	}

	.revenue-card .card-icon {
		background: rgba(253, 126, 20, 0.1);
		color: #fd7e14;
	}

	.product-card .card-icon {
		background: rgba(232, 62, 140, 0.1);
		color: #e83e8c;
	}

	/* Diagnostics Section */
	.diagnostics-section {
		background: var(--surface, white);
		border: 1px solid var(--border, #e5e7eb);
		border-radius: var(--radius-lg, 0.5rem);
		padding: var(--space-lg, 1.5rem);
		margin: var(--space-xl, 2rem) 0;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
	}

	.diagnostics-header {
		margin-bottom: var(--space-lg, 1.5rem);
		text-align: center;
	}

	.diagnostics-header h2 {
		margin: 0 0 var(--space-sm, 0.5rem) 0;
		color: var(--text, #111827);
		font-size: 1.5rem;
		font-weight: 600;
	}

	.diagnostics-header p {
		margin: 0;
		color: var(--text-muted, #6b7280);
		font-size: 1rem;
	}

	.diagnostics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: var(--space-md, 1rem);
		margin-bottom: var(--space-lg, 1.5rem);
	}

	.diagnostic-card {
		background: var(--bg-secondary, #f9fafb);
		border: 1px solid var(--border, #e5e7eb);
		border-radius: var(--radius-md, 0.375rem);
		padding: var(--space-md, 1rem);
		text-align: center;
		transition: all 0.2s ease;
	}

	.diagnostic-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}

	.diagnostic-card.highlight {
		border-color: #059669;
		background: #f0fdf4;
		border-width: 2px;
	}

	.diagnostic-icon {
		font-size: 2rem;
		margin-bottom: var(--space-sm, 0.5rem);
	}

	.diagnostic-value {
		font-size: 1.5rem;
		font-weight: bold;
		color: var(--primary, #2563eb);
		margin-bottom: var(--space-xs, 0.25rem);
	}

	.diagnostic-card.highlight .diagnostic-value {
		color: #059669;
	}

	.diagnostic-label {
		font-size: 0.875rem;
		color: var(--text, #111827);
		font-weight: 500;
		margin-bottom: var(--space-xs, 0.25rem);
	}

	.diagnostic-subtitle {
		font-size: 0.75rem;
		color: var(--text-muted, #6b7280);
		font-style: italic;
	}

	.diagnostics-explanation {
		border-top: 1px solid var(--border, #e5e7eb);
		padding-top: var(--space-lg, 1.5rem);
	}

	.explanation-card {
		background: #f8fafc;
		border: 1px solid #e2e8f0;
		border-radius: var(--radius-md, 0.375rem);
		padding: var(--space-lg, 1.5rem);
	}

	.explanation-card h3 {
		margin: 0 0 var(--space-md, 1rem) 0;
		color: var(--text, #111827);
		font-size: 1.125rem;
		font-weight: 600;
	}

	.explanation-card p {
		margin: 0 0 var(--space-sm, 0.5rem) 0;
		color: var(--text, #111827);
		line-height: 1.5;
	}

	.explanation-card ul {
		margin: var(--space-sm, 0.5rem) 0;
		padding-left: var(--space-lg, 1.5rem);
		color: var(--text-muted, #6b7280);
	}

	.explanation-card li {
		margin-bottom: var(--space-xs, 0.25rem);
		line-height: 1.4;
	}

	/* Sync Status Banner */
	.sync-status-banner {
		background: #f0f9ff;
		border: 1px solid #0ea5e9;
		border-radius: var(--radius-md, 0.375rem);
		padding: var(--space-md, 1rem);
		margin: var(--space-lg, 1.5rem) 0;
		animation: slideIn 0.3s ease-out;
	}

	.sync-status-content {
		display: flex;
		align-items: center;
		gap: var(--space-md, 1rem);
	}

	.sync-status-message {
		flex: 1;
		color: #0369a1;
		font-weight: 500;
	}

	.sync-spinner {
		width: 20px;
		height: 20px;
		border: 2px solid #e0f2fe;
		border-top: 2px solid #0ea5e9;
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes slideIn {
		from {
			opacity: 0;
			transform: translateY(-10px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
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

		.summary-grid {
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

