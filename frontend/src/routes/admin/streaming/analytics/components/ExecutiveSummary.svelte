<script lang="ts">
	import { onMount } from 'svelte';

	// Props
	export let data: any = null;
	export let period: string = '30d';

	// Local state
	let isLoading = false;
	let error: string | null = null;

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
</script>

<div class="space-y-6">
	<!-- Key Business Metrics -->
	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
		<!-- Revenue Impact Card -->
		<div class="bg-white rounded-lg shadow p-6">
			<div class="flex items-center">
				<div class="flex-shrink-0">
					<div class="w-8 h-8 bg-blue-100 rounded-lg flex items-center justify-center">
						<span class="text-blue-600 text-lg">💰</span>
					</div>
				</div>
				<div class="ml-4">
					<p class="text-sm font-medium text-gray-500">Promotional Revenue</p>
					<p class="text-2xl font-bold text-gray-900">
						{data?.revenue_impact?.promotional_revenue ? formatCurrency(data.revenue_impact.promotional_revenue) : '$0'}
					</p>
					<p class="text-sm text-green-600">+{data?.revenue_impact?.growth_rate || 0}% vs baseline</p>
				</div>
			</div>
		</div>

		<!-- Customer Impact Card -->
		<div class="bg-white rounded-lg shadow p-6">
			<div class="flex items-center">
				<div class="flex-shrink-0">
					<div class="w-8 h-8 bg-green-100 rounded-lg flex items-center justify-center">
						<span class="text-green-600 text-lg">👥</span>
					</div>
				</div>
				<div class="ml-4">
					<p class="text-sm font-medium text-gray-500">New Customers (Promos)</p>
					<p class="text-2xl font-bold text-gray-900">
						{data?.customer_impact?.new_customers_promos || 0}
					</p>
					<p class="text-sm text-green-600">+{data?.customer_impact?.overall_growth || 0}% growth</p>
				</div>
			</div>
		</div>

		<!-- Funnel Performance Card -->
		<div class="bg-white rounded-lg shadow p-6">
			<div class="flex items-center">
				<div class="flex-shrink-0">
					<div class="w-8 h-8 bg-purple-100 rounded-lg flex items-center justify-center">
						<span class="text-purple-600 text-lg">📊</span>
					</div>
				</div>
				<div class="ml-4">
					<p class="text-sm font-medium text-gray-500">Conversion Lift</p>
					<p class="text-2xl font-bold text-gray-900">
						{data?.funnel_performance?.conversion_lift || 0}%
					</p>
					<p class="text-sm text-green-600">vs standard plans</p>
				</div>
			</div>
		</div>

		<!-- Total MRR Card -->
		<div class="bg-white rounded-lg shadow p-6">
			<div class="flex items-center">
				<div class="flex-shrink-0">
					<div class="w-8 h-8 bg-yellow-100 rounded-lg flex items-center justify-center">
						<span class="text-yellow-600 text-lg">📈</span>
					</div>
				</div>
				<div class="ml-4">
					<p class="text-sm font-medium text-gray-500">Total MRR</p>
					<p class="text-2xl font-bold text-gray-900">
						{data?.revenue_impact?.total_mrr ? formatCurrency(data.revenue_impact.total_mrr) : '$0'}
					</p>
					<p class="text-sm text-green-600">Monthly recurring</p>
				</div>
			</div>
		</div>
	</div>

	<!-- Business Intelligence Summary -->
	<div class="bg-white rounded-lg shadow p-6">
		<h3 class="text-lg font-semibold text-gray-900 mb-4">Business Intelligence Summary</h3>
		<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
			<div>
				<h4 class="font-medium text-gray-900 mb-3">Revenue Impact Analysis</h4>
				<div class="space-y-2">
					<div class="flex justify-between">
						<span class="text-sm text-gray-600">Standard Plans Revenue:</span>
						<span class="text-sm font-medium">{data?.revenue_impact?.standard_revenue ? formatCurrency(data.revenue_impact.standard_revenue) : '$0'}</span>
					</div>
					<div class="flex justify-between">
						<span class="text-sm text-gray-600">Promotional Revenue:</span>
						<span class="text-sm font-medium text-green-600">{data?.revenue_impact?.promotional_revenue ? formatCurrency(data.revenue_impact.promotional_revenue) : '$0'}</span>
					</div>
					<div class="flex justify-between">
						<span class="text-sm text-gray-600">Revenue Growth:</span>
						<span class="text-sm font-medium text-green-600">+{data?.revenue_impact?.growth_rate || 0}%</span>
					</div>
				</div>
			</div>
			<div>
				<h4 class="font-medium text-gray-900 mb-3">Customer Impact Analysis</h4>
				<div class="space-y-2">
					<div class="flex justify-between">
						<span class="text-sm text-gray-600">Standard Conversions:</span>
						<span class="text-sm font-medium">{data?.customer_impact?.standard_conversions || 0}</span>
					</div>
					<div class="flex justify-between">
						<span class="text-sm text-gray-600">Promotional Conversions:</span>
						<span class="text-sm font-medium text-green-600">{data?.customer_impact?.new_customers_promos || 0}</span>
					</div>
					<div class="flex justify-between">
						<span class="text-sm text-gray-600">Overall Growth:</span>
						<span class="text-sm font-medium text-green-600">+{data?.customer_impact?.overall_growth || 0}%</span>
					</div>
				</div>
			</div>
		</div>
	</div>
</div> 
