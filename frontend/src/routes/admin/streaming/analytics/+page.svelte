<script>
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';

	// Reactive variables
	let isLoading = true;
	let analyticsData = null;
	let error = null;
	let selectedPeriod = '30d'; // 7d, 30d, 90d, 1y
	let selectedMetric = 'revenue'; // revenue, subscriptions, churn

	// Chart data
	let revenueData = [];
	let subscriptionData = [];
	let churnData = [];

	// Load analytics data
	async function loadAnalyticsData() {
		try {
			isLoading = true;
			error = null;

			const params = new URLSearchParams({
				period: selectedPeriod,
				metric: selectedMetric
			});

			const response = await fetch(`/api/admin/streaming/analytics/overview?${params}`);
			
			if (response.ok) {
				analyticsData = await response.json();
				processChartData();
			} else {
				throw new Error('Failed to load analytics data');
			}
		} catch (err) {
			console.error('Error loading analytics data:', err);
			error = err.message;
		} finally {
			isLoading = false;
		}
	}

	// Process data for charts
	function processChartData() {
		if (!analyticsData) return;

		// Process revenue data
		if (analyticsData.revenue_trend) {
			revenueData = analyticsData.revenue_trend.map(item => ({
				date: new Date(item.date),
				value: item.amount,
				formatted: formatCurrency(item.amount)
			}));
		}

		// Process subscription data
		if (analyticsData.subscription_trend) {
			subscriptionData = analyticsData.subscription_trend.map(item => ({
				date: new Date(item.date),
				value: item.count,
				formatted: item.count.toString()
			}));
		}

		// Process churn data
		if (analyticsData.churn_trend) {
			churnData = analyticsData.churn_trend.map(item => ({
				date: new Date(item.date),
				value: item.rate,
				formatted: `${item.rate.toFixed(2)}%`
			}));
		}
	}

	// Handle period change
	function handlePeriodChange() {
		loadAnalyticsData();
	}

	// Handle metric change
	function handleMetricChange() {
		loadAnalyticsData();
	}

	// Format currency
	function formatCurrency(amount, currency = 'USD') {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: currency
		}).format(amount);
	}

	// Format percentage
	function formatPercentage(value) {
		return `${value.toFixed(2)}%`;
	}

	// Format number
	function formatNumber(value) {
		return new Intl.NumberFormat('en-US').format(value);
	}

	// Get trend icon and color
	function getTrendIcon(value, previousValue) {
		if (!previousValue) return { icon: TrendingUpIcon, color: 'text-gray-400' };
		
		const change = value - previousValue;
		if (change > 0) {
			return { icon: TrendingUpIcon, color: 'text-green-600' };
		} else if (change < 0) {
			return { icon: TrendingDownIcon, color: 'text-red-600' };
		} else {
			return { icon: TrendingUpIcon, color: 'text-gray-400' };
		}
	}

	// Get trend color
	function getTrendColor(value, previousValue) {
		if (!previousValue) return 'text-gray-500';
		
		const change = value - previousValue;
		if (change > 0) return 'text-green-600';
		if (change < 0) return 'text-red-600';
		return 'text-gray-500';
	}

	// Calculate percentage change
	function calculateChange(current, previous) {
		if (!previous) return 0;
		return ((current - previous) / previous) * 100;
	}

	// Get status distribution data
	function getStatusDistribution() {
		if (!analyticsData?.subscription_status_distribution) return [];
		
		return Object.entries(analyticsData.subscription_status_distribution).map(([status, count]) => ({
			status,
			count,
			percentage: (count / analyticsData.total_subscriptions) * 100
		}));
	}

	// Get plan distribution data
	function getPlanDistribution() {
		if (!analyticsData?.plan_distribution) return [];
		
		return Object.entries(analyticsData.plan_distribution).map(([plan, count]) => ({
			plan,
			count,
			percentage: (count / analyticsData.total_subscriptions) * 100
		}));
	}

	// Initialize component
	onMount(() => {
		loadAnalyticsData();
	});
</script>

<svelte:head>
	<title>Analytics - Streaming Admin</title>
</svelte:head>

{#if isLoading}
	<div class="flex items-center justify-center py-12">
		<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
	</div>
{:else if error}
	<div class="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
		<div class="flex items-center">
			<ExclamationTriangleIcon class="h-5 w-5 text-red-400 mr-2" />
			<p class="text-red-800">{error}</p>
		</div>
	</div>
{:else}
	<div class="space-y-6" in:fade={{ duration: 300 }}>
		<!-- Header -->
		<div class="flex justify-between items-center">
			<div>
				<h1 class="text-2xl font-bold text-gray-900">Subscription Analytics</h1>
				<p class="text-gray-600">Revenue, subscriptions, and performance metrics</p>
			</div>
			<div class="flex items-center space-x-4">
				<!-- Period Selector -->
				<select
					bind:value={selectedPeriod}
					on:change={handlePeriodChange}
					class="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
				>
					<option value="7d">Last 7 days</option>
					<option value="30d">Last 30 days</option>
					<option value="90d">Last 90 days</option>
					<option value="1y">Last year</option>
				</select>

				<!-- Metric Selector -->
				<select
					bind:value={selectedMetric}
					on:change={handleMetricChange}
					class="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
				>
					<option value="revenue">Revenue</option>
					<option value="subscriptions">Subscriptions</option>
					<option value="churn">Churn Rate</option>
				</select>
			</div>
		</div>

		<!-- Key Metrics Grid -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
			<!-- MRR -->
			<div class="bg-white border border-gray-200 rounded-lg p-6" in:fly={{ y: 20, duration: 300 }}>
				<div class="flex items-center justify-between">
					<div>
						<p class="text-sm font-medium text-gray-600">Monthly Recurring Revenue</p>
						<p class="text-2xl font-bold text-gray-900">
							{analyticsData?.mrr ? formatCurrency(analyticsData.mrr) : '$0'}
						</p>
					</div>
					<CurrencyDollarIcon class="h-8 w-8 text-green-600" />
				</div>
				{#if analyticsData?.mrr_change}
					<div class="mt-4 flex items-center">
						<svelte:component this={getTrendIcon(analyticsData.mrr, analyticsData.mrr_previous).icon} class="h-4 w-4 {getTrendIcon(analyticsData.mrr, analyticsData.mrr_previous).color} mr-1" />
						<span class="text-sm {getTrendColor(analyticsData.mrr, analyticsData.mrr_previous)}">
							{formatPercentage(Math.abs(calculateChange(analyticsData.mrr, analyticsData.mrr_previous)))} from last month
						</span>
					</div>
				{/if}
			</div>

			<!-- Active Subscriptions -->
			<div class="bg-white border border-gray-200 rounded-lg p-6" in:fly={{ y: 20, duration: 300 }}>
				<div class="flex items-center justify-between">
					<div>
						<p class="text-sm font-medium text-gray-600">Active Subscriptions</p>
						<p class="text-2xl font-bold text-gray-900">
							{analyticsData?.active_subscriptions ? formatNumber(analyticsData.active_subscriptions) : '0'}
						</p>
					</div>
					<UsersIcon class="h-8 w-8 text-blue-600" />
				</div>
				{#if analyticsData?.subscription_change}
					<div class="mt-4 flex items-center">
						<svelte:component this={getTrendIcon(analyticsData.active_subscriptions, analyticsData.previous_subscriptions).icon} class="h-4 w-4 {getTrendIcon(analyticsData.active_subscriptions, analyticsData.previous_subscriptions).color} mr-1" />
						<span class="text-sm {getTrendColor(analyticsData.active_subscriptions, analyticsData.previous_subscriptions)}">
							{formatPercentage(Math.abs(calculateChange(analyticsData.active_subscriptions, analyticsData.previous_subscriptions)))} from last month
						</span>
					</div>
				{/if}
			</div>

			<!-- Churn Rate -->
			<div class="bg-white border border-gray-200 rounded-lg p-6" in:fly={{ y: 20, duration: 300 }}>
				<div class="flex items-center justify-between">
					<div>
						<p class="text-sm font-medium text-gray-600">Churn Rate</p>
						<p class="text-2xl font-bold text-gray-900">
							{analyticsData?.churn_rate ? formatPercentage(analyticsData.churn_rate) : '0%'}
						</p>
					</div>
					<XCircleIcon class="h-8 w-8 text-red-600" />
				</div>
				{#if analyticsData?.churn_change}
					<div class="mt-4 flex items-center">
						<svelte:component this={getTrendIcon(analyticsData.churn_rate, analyticsData.previous_churn_rate).icon} class="h-4 w-4 {getTrendIcon(analyticsData.churn_rate, analyticsData.previous_churn_rate).color} mr-1" />
						<span class="text-sm {getTrendColor(analyticsData.churn_rate, analyticsData.previous_churn_rate)}">
							{formatPercentage(Math.abs(calculateChange(analyticsData.churn_rate, analyticsData.previous_churn_rate)))} from last month
						</span>
					</div>
				{/if}
			</div>

			<!-- Average Revenue Per User -->
			<div class="bg-white border border-gray-200 rounded-lg p-6" in:fly={{ y: 20, duration: 300 }}>
				<div class="flex items-center justify-between">
					<div>
						<p class="text-sm font-medium text-gray-600">ARPU</p>
						<p class="text-2xl font-bold text-gray-900">
							{analyticsData?.arpu ? formatCurrency(analyticsData.arpu) : '$0'}
						</p>
					</div>
					<ChartBarIcon class="h-8 w-8 text-purple-600" />
				</div>
				{#if analyticsData?.arpu_change}
					<div class="mt-4 flex items-center">
						<svelte:component this={getTrendIcon(analyticsData.arpu, analyticsData.previous_arpu).icon} class="h-4 w-4 {getTrendIcon(analyticsData.arpu, analyticsData.previous_arpu).color} mr-1" />
						<span class="text-sm {getTrendColor(analyticsData.arpu, analyticsData.previous_arpu)}">
							{formatPercentage(Math.abs(calculateChange(analyticsData.arpu, analyticsData.previous_arpu)))} from last month
						</span>
					</div>
				{/if}
			</div>
		</div>

		<!-- Charts Section -->
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
			<!-- Revenue Trend Chart -->
			<div class="bg-white border border-gray-200 rounded-lg p-6">
				<h3 class="text-lg font-semibold text-gray-900 mb-4">Revenue Trend</h3>
				{#if revenueData.length > 0}
					<div class="h-64 flex items-end justify-between space-x-2">
						{#each revenueData as item, index}
							<div class="flex-1 flex flex-col items-center">
								<div 
									class="bg-blue-600 rounded-t w-full"
									style="height: {(item.value / Math.max(...revenueData.map(d => d.value))) * 200}px;"
								></div>
								<p class="text-xs text-gray-500 mt-2">{item.date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}</p>
								<p class="text-xs font-medium text-gray-900">{item.formatted}</p>
							</div>
						{/each}
					</div>
				{:else}
					<div class="h-64 flex items-center justify-center text-gray-500">
						No revenue data available
					</div>
				{/if}
			</div>

			<!-- Subscription Growth Chart -->
			<div class="bg-white border border-gray-200 rounded-lg p-6">
				<h3 class="text-lg font-semibold text-gray-900 mb-4">Subscription Growth</h3>
				{#if subscriptionData.length > 0}
					<div class="h-64 flex items-end justify-between space-x-2">
						{#each subscriptionData as item, index}
							<div class="flex-1 flex flex-col items-center">
								<div 
									class="bg-green-600 rounded-t w-full"
									style="height: {(item.value / Math.max(...subscriptionData.map(d => d.value))) * 200}px;"
								></div>
								<p class="text-xs text-gray-500 mt-2">{item.date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}</p>
								<p class="text-xs font-medium text-gray-900">{item.formatted}</p>
							</div>
						{/each}
					</div>
				{:else}
					<div class="h-64 flex items-center justify-center text-gray-500">
						No subscription data available
					</div>
				{/if}
			</div>
		</div>

		<!-- Distribution Charts -->
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
			<!-- Subscription Status Distribution -->
			<div class="bg-white border border-gray-200 rounded-lg p-6">
				<h3 class="text-lg font-semibold text-gray-900 mb-4">Subscription Status Distribution</h3>
				{#if getStatusDistribution().length > 0}
					<div class="space-y-3">
						{#each getStatusDistribution() as item}
							<div class="flex items-center justify-between">
								<div class="flex items-center">
									<div class="w-3 h-3 rounded-full mr-3 {item.status === 'active' ? 'bg-green-500' : item.status === 'cancelled' ? 'bg-red-500' : item.status === 'past_due' ? 'bg-yellow-500' : 'bg-gray-500'}"></div>
									<span class="text-sm font-medium text-gray-900 capitalize">{item.status.replace('_', ' ')}</span>
								</div>
								<div class="flex items-center space-x-2">
									<div class="w-32 bg-gray-200 rounded-full h-2">
										<div class="bg-blue-600 h-2 rounded-full" style="width: {item.percentage}%"></div>
									</div>
									<span class="text-sm text-gray-600 w-12 text-right">{item.count}</span>
								</div>
							</div>
						{/each}
					</div>
				{:else}
					<div class="text-center py-8 text-gray-500">
						No status distribution data available
					</div>
				{/if}
			</div>

			<!-- Plan Distribution -->
			<div class="bg-white border border-gray-200 rounded-lg p-6">
				<h3 class="text-lg font-semibold text-gray-900 mb-4">Plan Distribution</h3>
				{#if getPlanDistribution().length > 0}
					<div class="space-y-3">
						{#each getPlanDistribution() as item}
							<div class="flex items-center justify-between">
								<div class="flex items-center">
									<div class="w-3 h-3 rounded-full mr-3 {item.plan === 'monthly' ? 'bg-blue-500' : item.plan === 'annual' ? 'bg-green-500' : item.plan === 'premium' ? 'bg-purple-500' : 'bg-gray-500'}"></div>
									<span class="text-sm font-medium text-gray-900 capitalize">{item.plan}</span>
								</div>
								<div class="flex items-center space-x-2">
									<div class="w-32 bg-gray-200 rounded-full h-2">
										<div class="bg-green-600 h-2 rounded-full" style="width: {item.percentage}%"></div>
									</div>
									<span class="text-sm text-gray-600 w-12 text-right">{item.count}</span>
								</div>
							</div>
						{/each}
					</div>
				{:else}
					<div class="text-center py-8 text-gray-500">
						No plan distribution data available
					</div>
				{/if}
			</div>
		</div>

		<!-- Recent Activity -->
		<div class="bg-white border border-gray-200 rounded-lg p-6">
			<h3 class="text-lg font-semibold text-gray-900 mb-4">Recent Activity</h3>
			{#if analyticsData?.recent_activity && analyticsData.recent_activity.length > 0}
				<div class="space-y-4">
					{#each analyticsData.recent_activity.slice(0, 10) as activity}
						<div class="flex items-center space-x-3 p-3 bg-gray-50 rounded-lg">
							<div class="flex-shrink-0">
								{#if activity.type === 'subscription_created'}
									<CheckCircleIcon class="h-5 w-5 text-green-600" />
								{:else if activity.type === 'subscription_cancelled'}
									<XCircleIcon class="h-5 w-5 text-red-600" />
								{:else if activity.type === 'payment_failed'}
									<ExclamationTriangleIcon class="h-5 w-5 text-yellow-600" />
								{:else}
									<ClockIcon class="h-5 w-5 text-gray-600" />
								{/if}
							</div>
							<div class="flex-1">
								<p class="text-sm font-medium text-gray-900">{activity.description}</p>
								<p class="text-xs text-gray-500">{new Date(activity.timestamp).toLocaleString()}</p>
							</div>
							{#if activity.amount}
								<span class="text-sm font-medium text-gray-900">{formatCurrency(activity.amount)}</span>
							{/if}
						</div>
					{/each}
				</div>
			{:else}
				<div class="text-center py-8 text-gray-500">
					No recent activity
				</div>
			{/if}
		</div>

		<!-- Performance Metrics -->
		<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
			<!-- Conversion Rate -->
			<div class="bg-white border border-gray-200 rounded-lg p-6">
				<h4 class="text-sm font-medium text-gray-600 mb-2">Conversion Rate</h4>
				<p class="text-2xl font-bold text-gray-900">
					{analyticsData?.conversion_rate ? formatPercentage(analyticsData.conversion_rate) : '0%'}
				</p>
				<p class="text-sm text-gray-500 mt-1">Trial to paid</p>
			</div>

			<!-- Average Subscription Length -->
			<div class="bg-white border border-gray-200 rounded-lg p-6">
				<h4 class="text-sm font-medium text-gray-600 mb-2">Avg. Subscription Length</h4>
				<p class="text-2xl font-bold text-gray-900">
					{analyticsData?.avg_subscription_length ? `${analyticsData.avg_subscription_length} months` : '0 months'}
				</p>
				<p class="text-sm text-gray-500 mt-1">Active subscriptions</p>
			</div>

			<!-- Customer Lifetime Value -->
			<div class="bg-white border border-gray-200 rounded-lg p-6">
				<h4 class="text-sm font-medium text-gray-600 mb-2">Customer LTV</h4>
				<p class="text-2xl font-bold text-gray-900">
					{analyticsData?.customer_ltv ? formatCurrency(analyticsData.customer_ltv) : '$0'}
				</p>
				<p class="text-sm text-gray-500 mt-1">Lifetime value</p>
			</div>
		</div>
	</div>
{/if} 