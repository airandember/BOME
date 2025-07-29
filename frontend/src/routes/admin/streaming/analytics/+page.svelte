<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { StreamingSubscriptionService } from '$lib/services/streaming-subscriptions';
	
	// Import analytics components
	import ExecutiveSummary from './components/ExecutiveSummary.svelte';
	import FunnelAnalysis from './components/FunnelAnalysis.svelte';
	import RevenueImpact from './components/RevenueImpact.svelte';
	import CustomerJourney from './components/CustomerJourney.svelte';
	import PromotionAnalytics from './components/PromotionAnalytics.svelte';

	// Reactive variables
	let isLoading = true;
	let analyticsData: any = null;
	let error: string | null = null;
	let selectedPeriod = '30d'; // 7d, 30d, 90d, 1y
	let selectedMetric = 'revenue'; // revenue, subscriptions, churn
	let selectedTab = 'overview'; // overview, executive-summary, funnel-analysis, revenue-impact, customer-journey, promotions, audit, timeline

	// Chart data
	let revenueData: any[] = [];
	let subscriptionData: any[] = [];
	let churnData: any[] = [];

	// Promotion analytics data
	let promotionPlans: any[] = [];
	let promotionHistory: any[] = [];
	let auditLogs: any[] = [];
	let timelineEvents: any[] = [];

	// Business Intelligence data
	let executiveSummaryData: any = null;
	let funnelAnalysisData: any = null;
	let revenueImpactData: any = null;
	let customerJourneyData: any = null;

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
		} catch (err: any) {
			console.error('Error loading analytics data:', err);
			error = err.message;
		} finally {
			isLoading = false;
		}
	}

	// Load executive summary data
	async function loadExecutiveSummaryData() {
		try {
			const params = new URLSearchParams({
				period: selectedPeriod
			});

			const response = await fetch(`/api/admin/streaming/analytics/executive-summary?${params}`);
			
			if (response.ok) {
				executiveSummaryData = await response.json();
			} else {
				// Fallback to mock data for now
				executiveSummaryData = {
					revenue_impact: {
						promotional_revenue: 12450,
						standard_revenue: 45200,
						total_mrr: 57650,
						growth_rate: 15
					},
					customer_impact: {
						new_customers_promos: 234,
						standard_conversions: 156,
						overall_growth: 18
					},
					funnel_performance: {
						promo_conversion: 3.2,
						standard_conversion: 1.8,
						conversion_lift: 78
					}
				};
			}
		} catch (err: any) {
			console.error('Error loading executive summary data:', err);
		}
	}

	// Load funnel analysis data
	async function loadFunnelAnalysisData() {
		try {
			const params = new URLSearchParams({
				period: selectedPeriod
			});

			const response = await fetch(`/api/admin/streaming/analytics/funnel-analysis?${params}`);
			
			if (response.ok) {
				funnelAnalysisData = await response.json();
			} else {
				// Fallback to mock data for now
				funnelAnalysisData = {
					stages: [
						{ name: 'Awareness', standard: 10000, promotional: 15000, lift: 50 },
						{ name: 'Interest', standard: 2500, promotional: 4500, lift: 80 },
						{ name: 'Consideration', standard: 1250, promotional: 2700, lift: 116 },
						{ name: 'Conversion', standard: 180, promotional: 432, lift: 140 },
						{ name: 'Retention', standard: 162, promotional: 389, lift: 140 }
					],
					conversion_rates: {
						standard: 1.8,
						promotional: 2.9,
						lift: 61
					}
				};
			}
		} catch (err: any) {
			console.error('Error loading funnel analysis data:', err);
		}
	}

	// Load revenue impact data
	async function loadRevenueImpactData() {
		try {
			const params = new URLSearchParams({
				period: selectedPeriod
			});

			const response = await fetch(`/api/admin/streaming/analytics/revenue-impact?${params}`);
			
			if (response.ok) {
				revenueImpactData = await response.json();
			} else {
				// Fallback to mock data for now
				revenueImpactData = {
					revenue_breakdown: {
						standard_plans: 45200,
						promotional_plans: 12450,
						total_revenue: 57650
					},
					promotional_performance: [
						{ name: 'Plan Share!', revenue: 8200, percentage: 66 },
						{ name: '3 for 4', revenue: 4250, percentage: 34 }
					],
					baseline_comparison: {
						pre_promo_mrr: 42000,
						current_mrr: 57650,
						promotional_lift: 37
					}
				};
			}
		} catch (err: any) {
			console.error('Error loading revenue impact data:', err);
		}
	}

	// Load customer journey data
	async function loadCustomerJourneyData() {
		try {
			const params = new URLSearchParams({
				period: selectedPeriod
			});

			const response = await fetch(`/api/admin/streaming/analytics/customer-journey?${params}`);
			
			if (response.ok) {
				customerJourneyData = await response.json();
			} else {
				// Fallback to mock data for now
				customerJourneyData = {
					journey_metrics: [
						{ metric: 'Time to Convert', standard: 14, promotional: 7, improvement: 50 },
						{ metric: 'Avg Order Value', standard: 29.99, promotional: 19.99, difference: -33 },
						{ metric: 'Retention Rate', standard: 85, promotional: 78, difference: -7 },
						{ metric: 'Upgrade Rate', standard: 12, promotional: 18, improvement: 50 },
						{ metric: 'LTV', standard: 450, promotional: 380, difference: -16 }
					],
					net_impact: 'positive'
				};
			}
		} catch (err: any) {
			console.error('Error loading customer journey data:', err);
		}
	}

	// Load promotion analytics
	async function loadPromotionAnalytics() {
		try {
			const plans = await StreamingSubscriptionService.getAll();
			promotionPlans = plans.filter((plan: any) => plan.sub_type === 'prmo');
			
			// Process promotion metadata for analytics
			promotionPlans.forEach((plan: any) => {
				if (plan.promotion_metadata) {
					// Extract performance metrics
					const metadata = plan.promotion_metadata;
					if (metadata.promotion_stats) {
						(plan as any).performanceMetrics = metadata.promotion_stats.performance_metrics;
						(plan as any).currentPromotion = metadata.promotion_stats.current_promotion;
					}
				}
			});

			// Generate timeline events from plan change history
			timelineEvents = [];
			plans.forEach((plan: any) => {
				if (plan.plan_change_history && plan.plan_change_history.length > 0) {
					plan.plan_change_history.forEach((event: any) => {
						if (typeof event === 'string') {
							try {
								const parsedEvent = JSON.parse(event);
								timelineEvents.push({
									...parsedEvent,
									planName: plan.name,
									planId: plan.id
								});
							} catch (e) {
								// Handle string events
								timelineEvents.push({
									id: `event_${Date.now()}_${Math.random()}`,
									event_type: 'plan_change',
									timestamp: new Date(),
									description: event,
									planName: plan.name,
									planId: plan.id
								});
							}
						}
					});
				}
			});

			// Sort timeline events by timestamp
			timelineEvents.sort((a: any, b: any) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());

		} catch (err: any) {
			console.error('Error loading promotion analytics:', err);
		}
	}

	// Load audit logs
	async function loadAuditLogs() {
		try {
			const response = await fetch('/api/admin/streaming/analytics/audit-logs');
			if (response.ok) {
				auditLogs = await response.json();
			}
		} catch (err: any) {
			console.error('Error loading audit logs:', err);
		}
	}

	// Process data for charts
	function processChartData() {
		if (!analyticsData) return;

		// Process revenue data
		if (analyticsData.revenue_trend) {
			revenueData = analyticsData.revenue_trend.map((item: any) => ({
				date: new Date(item.date),
				value: item.amount,
				formatted: formatCurrency(item.amount)
			}));
		}

		// Process subscription data
		if (analyticsData.subscription_trend) {
			subscriptionData = analyticsData.subscription_trend.map((item: any) => ({
				date: new Date(item.date),
				value: item.count,
				formatted: item.count.toString()
			}));
		}

		// Process churn data
		if (analyticsData.churn_trend) {
			churnData = analyticsData.churn_trend.map((item: any) => ({
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

	// Handle tab change
	function handleTabChange() {
		if (selectedTab === 'executive-summary') {
			loadExecutiveSummaryData();
		} else if (selectedTab === 'funnel-analysis') {
			loadFunnelAnalysisData();
		} else if (selectedTab === 'revenue-impact') {
			loadRevenueImpactData();
		} else if (selectedTab === 'customer-journey') {
			loadCustomerJourneyData();
		} else if (selectedTab === 'promotions') {
			loadPromotionAnalytics();
		} else if (selectedTab === 'audit') {
			loadAuditLogs();
		}
	}

	// Format currency
	function formatCurrency(amount: number, currency = 'USD'): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: currency
		}).format(amount);
	}

	// Format percentage
	function formatPercentage(value: number): string {
		return `${value.toFixed(2)}%`;
	}

	// Format number
	function formatNumber(value: number): string {
		return new Intl.NumberFormat('en-US').format(value);
	}

	// Format date for timeline
	function formatTimelineDate(date: any): string {
		return new Date(date).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	// Get event type color
	function getEventTypeColor(eventType: string): string {
		const colors: Record<string, string> = {
			'plan_created': 'bg-blue-100 text-blue-800',
			'plan_updated': 'bg-yellow-100 text-yellow-800',
			'promotion_started': 'bg-green-100 text-green-800',
			'promotion_ended': 'bg-red-100 text-red-800',
			'status_toggled': 'bg-purple-100 text-purple-800',
			'plan_change': 'bg-gray-100 text-gray-800'
		};
		return colors[eventType] || 'bg-gray-100 text-gray-800';
	}

	// Get promotion status
	function getPromotionStatus(plan: any): string {
		if (!plan.promotion_start_date || !plan.promotion_end_date) return 'unknown';
		
		const now = new Date();
		const startDate = new Date(plan.promotion_start_date);
		const endDate = new Date(plan.promotion_end_date);
		
		if (now < startDate) return 'upcoming';
		if (now > endDate) return 'expired';
		return 'active';
	}

	// Get status color
	function getStatusColor(status: string): string {
		const colors: Record<string, string> = {
			'active': 'bg-green-100 text-green-800',
			'upcoming': 'bg-blue-100 text-blue-800',
			'expired': 'bg-red-100 text-red-800',
			'unknown': 'bg-gray-100 text-gray-800'
		};
		return colors[status] || 'bg-gray-100 text-gray-800';
	}

	// Calculate total promotion revenue
	function calculateTotalPromotionRevenue(): number {
		return promotionPlans.reduce((total: number, plan: any) => {
			if ((plan as any).performanceMetrics?.total_revenue_generated) {
				return total + (plan as any).performanceMetrics.total_revenue_generated;
			}
			return total;
		}, 0);
	}

	// Calculate average conversion rate
	function calculateAverageConversionRate(): number {
		const plansWithMetrics = promotionPlans.filter((plan: any) => 
			(plan as any).performanceMetrics?.average_conversion_rate
		);
		
		if (plansWithMetrics.length === 0) return 0;
		
		const totalRate = plansWithMetrics.reduce((sum: number, plan: any) => 
			sum + (plan as any).performanceMetrics.average_conversion_rate, 0
		);
		
		return totalRate / plansWithMetrics.length;
	}

	// Get trend icon and color
	function getTrendIcon(value: number, previousValue: number) {
		if (!previousValue) return { icon: '↗️', color: 'text-gray-400' };
		
		const change = value - previousValue;
		if (change > 0) {
			return { icon: '↗️', color: 'text-green-600' };
		} else if (change < 0) {
			return { icon: '↘️', color: 'text-red-600' };
		} else {
			return { icon: '→', color: 'text-gray-400' };
		}
	}

	// Get trend color
	function getTrendColor(value: number, previousValue: number): string {
		if (!previousValue) return 'text-gray-500';
		
		const change = value - previousValue;
		if (change > 0) return 'text-green-600';
		if (change < 0) return 'text-red-600';
		return 'text-gray-500';
	}

	// Calculate percentage change
	function calculateChange(current: number, previous: number): number {
		if (!previous) return 0;
		return ((current - previous) / previous) * 100;
	}

	// Get status distribution data
	function getStatusDistribution(): any[] {
		if (!analyticsData?.subscription_status_distribution) return [];
		
		return Object.entries(analyticsData.subscription_status_distribution).map(([status, count]: [string, any]) => ({
			status,
			count,
			percentage: (count / analyticsData.total_subscriptions) * 100
		}));
	}

	// Get plan distribution data
	function getPlanDistribution(): any[] {
		if (!analyticsData?.plan_distribution) return [];
		
		return Object.entries(analyticsData.plan_distribution).map(([plan, count]: [string, any]) => ({
			plan,
			count,
			percentage: (count / analyticsData.total_subscriptions) * 100
		}));
	}

	// Initialize component
	onMount(() => {
		loadAnalyticsData();
		loadPromotionAnalytics();
	});
</script>

<svelte:head>
	<title>Analytics - Streaming Admin</title>
</svelte:head>

{#if isLoading}
	<div class="flex items-center justify-center py-12">
		<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
	</div>
<!--{:else if error}
	<div class="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
		<div class="flex items-center">
			<span class="text-red-800">⚠️</span>
			<p class="text-red-800 ml-2">{error}</p>
		</div>
	</div>-->
{:else}
	<div class="space-y-6" in:fade={{ duration: 300 }}>
		<!-- Header -->
		<div class="flex justify-between items-center">
			<div>
				<h1 class="text-2xl font-bold text-gray-900">Subscription Analytics</h1>
				<p class="text-gray-600">Revenue, subscriptions, and comprehensive promotion tracking</p>
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

		<!-- Tab Navigation -->
		<div class="border-b border-gray-200">
			<nav class="-mb-px flex space-x-8 overflow-x-auto">
				<button
					class="py-2 px-1 border-b-2 font-medium text-sm whitespace-nowrap {selectedTab === 'overview' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
					on:click={() => selectedTab = 'overview'}
				>
					Overview
				</button>
				<button
					class="py-2 px-1 border-b-2 font-medium text-sm whitespace-nowrap {selectedTab === 'executive-summary' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
					on:click={() => { selectedTab = 'executive-summary'; handleTabChange(); }}
				>
					Executive Summary
				</button>
				<button
					class="py-2 px-1 border-b-2 font-medium text-sm whitespace-nowrap {selectedTab === 'funnel-analysis' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
					on:click={() => { selectedTab = 'funnel-analysis'; handleTabChange(); }}
				>
					Funnel Analysis
				</button>
				<button
					class="py-2 px-1 border-b-2 font-medium text-sm whitespace-nowrap {selectedTab === 'revenue-impact' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
					on:click={() => { selectedTab = 'revenue-impact'; handleTabChange(); }}
				>
					Revenue Impact
				</button>
				<button
					class="py-2 px-1 border-b-2 font-medium text-sm whitespace-nowrap {selectedTab === 'customer-journey' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
					on:click={() => { selectedTab = 'customer-journey'; handleTabChange(); }}
				>
					Customer Journey
				</button>
				<button
					class="py-2 px-1 border-b-2 font-medium text-sm whitespace-nowrap {selectedTab === 'promotions' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
					on:click={() => { selectedTab = 'promotions'; handleTabChange(); }}
				>
					Promotion Analytics
				</button>
				<button
					class="py-2 px-1 border-b-2 font-medium text-sm whitespace-nowrap {selectedTab === 'timeline' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
					on:click={() => selectedTab = 'timeline'}
				>
					Change Timeline
				</button>
				<button
					class="py-2 px-1 border-b-2 font-medium text-sm whitespace-nowrap {selectedTab === 'audit' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
					on:click={() => { selectedTab = 'audit'; handleTabChange(); }}
				>
					Audit Log
				</button>
			</nav>
		</div>

		<!-- Tab Content -->
		{#if selectedTab === 'overview'}
			<!-- Overview Tab -->
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
				<!-- Revenue Card -->
				<div class="bg-white rounded-lg shadow p-6">
					<div class="flex items-center">
						<div class="flex-shrink-0">
							<div class="w-8 h-8 bg-blue-100 rounded-lg flex items-center justify-center">
								<span class="text-blue-600 text-lg">💰</span>
							</div>
						</div>
						<div class="ml-4">
							<p class="text-sm font-medium text-gray-500">Total Revenue</p>
							<p class="text-2xl font-bold text-gray-900">
								{analyticsData?.total_revenue ? formatCurrency(analyticsData.total_revenue) : '$0'}
							</p>
						</div>
					</div>
				</div>

				<!-- Subscriptions Card -->
				<div class="bg-white rounded-lg shadow p-6">
					<div class="flex items-center">
						<div class="flex-shrink-0">
							<div class="w-8 h-8 bg-green-100 rounded-lg flex items-center justify-center">
								<span class="text-green-600 text-lg">👥</span>
							</div>
						</div>
						<div class="ml-4">
							<p class="text-sm font-medium text-gray-500">Active Subscriptions</p>
							<p class="text-2xl font-bold text-gray-900">
								{analyticsData?.total_subscriptions || 0}
							</p>
						</div>
					</div>
				</div>

				<!-- Promotion Revenue Card -->
				<div class="bg-white rounded-lg shadow p-6">
					<div class="flex items-center">
						<div class="flex-shrink-0">
							<div class="w-8 h-8 bg-purple-100 rounded-lg flex items-center justify-center">
								<span class="text-purple-600 text-lg">🎯</span>
							</div>
						</div>
						<div class="ml-4">
							<p class="text-sm font-medium text-gray-500">Promotion Revenue</p>
							<p class="text-2xl font-bold text-gray-900">
								{formatCurrency(calculateTotalPromotionRevenue())}
							</p>
						</div>
					</div>
				</div>

				<!-- Conversion Rate Card -->
				<div class="bg-white rounded-lg shadow p-6">
					<div class="flex items-center">
						<div class="flex-shrink-0">
							<div class="w-8 h-8 bg-yellow-100 rounded-lg flex items-center justify-center">
								<span class="text-yellow-600 text-lg">📈</span>
							</div>
						</div>
						<div class="ml-4">
							<p class="text-sm font-medium text-gray-500">Avg Conversion</p>
							<p class="text-2xl font-bold text-gray-900">
								{formatPercentage(calculateAverageConversionRate())}
							</p>
						</div>
					</div>
				</div>
			</div>

		{:else if selectedTab === 'promotions'}
			<!-- Promotion Analytics Tab -->
			<div class="space-y-6">
				<!-- Promotion Performance Summary -->
				<div class="bg-white rounded-lg shadow p-6">
					<h3 class="text-lg font-semibold text-gray-900 mb-4">Promotion Performance Summary</h3>
					<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
						<div class="text-center">
							<p class="text-2xl font-bold text-gray-900">{promotionPlans.length}</p>
							<p class="text-sm text-gray-500">Active Promotions</p>
						</div>
						<div class="text-center">
							<p class="text-2xl font-bold text-gray-900">{formatCurrency(calculateTotalPromotionRevenue())}</p>
							<p class="text-sm text-gray-500">Total Revenue</p>
						</div>
						<div class="text-center">
							<p class="text-2xl font-bold text-gray-900">{formatPercentage(calculateAverageConversionRate())}</p>
							<p class="text-sm text-gray-500">Avg Conversion</p>
						</div>
					</div>
				</div>

				<!-- Promotion Plans Grid -->
				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
					{#each promotionPlans as plan}
						<div class="bg-white rounded-lg shadow p-6">
							<div class="flex items-center justify-between mb-4">
								<h4 class="text-lg font-semibold text-gray-900">{plan.name}</h4>
								<span class="px-2 py-1 text-xs font-medium rounded-full {getStatusColor(getPromotionStatus(plan))}">
									{getPromotionStatus(plan)}
								</span>
							</div>
							
							<div class="space-y-3">
								<div class="flex justify-between">
									<span class="text-sm text-gray-500">Price:</span>
									<span class="text-sm font-medium">{formatCurrency(plan.price)}</span>
								</div>
								
								{#if plan.promotion_start_date && plan.promotion_end_date}
									<div class="flex justify-between">
										<span class="text-sm text-gray-500">Duration:</span>
										<span class="text-sm font-medium">
											{new Date(plan.promotion_start_date).toLocaleDateString()} - {new Date(plan.promotion_end_date).toLocaleDateString()}
										</span>
									</div>
								{/if}

								{#if plan.performanceMetrics}
									<div class="border-t pt-3">
										<div class="flex justify-between">
											<span class="text-sm text-gray-500">Revenue:</span>
											<span class="text-sm font-medium">{formatCurrency(plan.performanceMetrics.total_revenue_generated || 0)}</span>
										</div>
										<div class="flex justify-between">
											<span class="text-sm text-gray-500">Conversion:</span>
											<span class="text-sm font-medium">{formatPercentage(plan.performanceMetrics.average_conversion_rate || 0)}</span>
										</div>
									</div>
								{/if}
							</div>
						</div>
					{/each}
				</div>
			</div>

		{:else if selectedTab === 'timeline'}
			<!-- Change Timeline Tab -->
			<div class="bg-white rounded-lg shadow">
				<div class="p-6 border-b border-gray-200">
					<h3 class="text-lg font-semibold text-gray-900">Plan Change Timeline</h3>
					<p class="text-sm text-gray-500">Complete audit trail of all plan modifications</p>
				</div>
				
				<div class="p-6">
					{#if timelineEvents.length > 0}
						<div class="space-y-4">
							{#each timelineEvents as event}
								<div class="flex items-start space-x-4 p-4 bg-gray-50 rounded-lg">
									<div class="flex-shrink-0">
										<div class="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center">
											<span class="text-blue-600 text-sm">📝</span>
										</div>
									</div>
									<div class="flex-1 min-w-0">
										<div class="flex items-center space-x-2 mb-2">
											<span class="px-2 py-1 text-xs font-medium rounded-full {getEventTypeColor(event.event_type)}">
												{event.event_type?.replace('_', ' ') || 'plan change'}
											</span>
											<span class="text-sm text-gray-500">{formatTimelineDate(event.timestamp)}</span>
										</div>
										<p class="text-sm font-medium text-gray-900">{event.planName}</p>
										<p class="text-sm text-gray-600">{event.description}</p>
										{#if event.user_id}
											<p class="text-xs text-gray-500 mt-1">By: {event.user_id}</p>
										{/if}
									</div>
								</div>
							{/each}
						</div>
					{:else}
						<div class="text-center py-8">
							<p class="text-gray-500">No timeline events found</p>
						</div>
					{/if}
				</div>
			</div>

		{:else if selectedTab === 'audit'}
			<!-- Audit Log Tab -->
			<div class="bg-white rounded-lg shadow">
				<div class="p-6 border-b border-gray-200">
					<h3 class="text-lg font-semibold text-gray-900">Audit Log</h3>
					<p class="text-sm text-gray-500">Detailed log of all system activities and changes</p>
				</div>
				
				<div class="p-6">
					{#if auditLogs.length > 0}
						<div class="space-y-4">
							{#each auditLogs as log}
								<div class="flex items-start space-x-4 p-4 bg-gray-50 rounded-lg">
									<div class="flex-shrink-0">
										<div class="w-8 h-8 rounded-full bg-green-100 flex items-center justify-center">
											<span class="text-green-600 text-sm">🔍</span>
										</div>
									</div>
									<div class="flex-1 min-w-0">
										<div class="flex items-center space-x-2 mb-2">
											<span class="px-2 py-1 text-xs font-medium rounded-full bg-blue-100 text-blue-800">
												{log.action || 'audit'}
											</span>
											<span class="text-sm text-gray-500">{formatTimelineDate(log.timestamp)}</span>
										</div>
										<p class="text-sm font-medium text-gray-900">{log.description}</p>
										<p class="text-xs text-gray-500 mt-1">User: {log.user_id || 'System'}</p>
									</div>
								</div>
							{/each}
						</div>
					{:else}
						<div class="text-center py-8">
							<p class="text-gray-500">No audit logs found</p>
						</div>
					{/if}
				</div>
			</div>

		{:else if selectedTab === 'executive-summary'}
			<ExecutiveSummary data={executiveSummaryData} period={selectedPeriod} />

		{:else if selectedTab === 'funnel-analysis'}
			<FunnelAnalysis data={funnelAnalysisData} period={selectedPeriod} />

		{:else if selectedTab === 'revenue-impact'}
			<RevenueImpact data={revenueImpactData} period={selectedPeriod} />

		{:else if selectedTab === 'customer-journey'}
			<CustomerJourney data={customerJourneyData} period={selectedPeriod} />

		{:else if selectedTab === 'promotions'}
			<PromotionAnalytics promotionPlans={promotionPlans} period={selectedPeriod} />
		{/if}
	</div>
{/if} 