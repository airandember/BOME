<script lang="ts">
	// Props
	export let data: any = null;
	export let period: string = '30d';

	// Format currency
	function formatCurrency(amount: number, currency = 'USD'): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: currency
		}).format(amount);
	}
</script>

<div class="space-y-6">
	<!-- Revenue Breakdown -->
	<div class="bg-white rounded-lg shadow p-6">
		<h3 class="text-lg font-semibold text-gray-900 mb-4">Revenue Breakdown</h3>
		<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
			<div class="text-center">
				<p class="text-2xl font-bold text-gray-900">{data?.revenue_breakdown?.standard_plans ? formatCurrency(data.revenue_breakdown.standard_plans) : '$0'}</p>
				<p class="text-sm text-gray-500">Standard Plans</p>
			</div>
			<div class="text-center">
				<p class="text-2xl font-bold text-green-600">{data?.revenue_breakdown?.promotional_plans ? formatCurrency(data.revenue_breakdown.promotional_plans) : '$0'}</p>
				<p class="text-sm text-gray-500">Promotional Plans</p>
			</div>
			<div class="text-center">
				<p class="text-2xl font-bold text-blue-600">{data?.revenue_breakdown?.total_revenue ? formatCurrency(data.revenue_breakdown.total_revenue) : '$0'}</p>
				<p class="text-sm text-gray-500">Total Revenue</p>
			</div>
		</div>
	</div>

	<!-- Promotional Performance -->
	<div class="bg-white rounded-lg shadow p-6">
		<h3 class="text-lg font-semibold text-gray-900 mb-4">Promotional Performance</h3>
		<div class="space-y-4">
			{#each data?.promotional_performance || [] as promo}
				<div class="flex justify-between items-center p-4 bg-gray-50 rounded-lg">
					<div>
						<h4 class="font-medium text-gray-900">{promo.name}</h4>
						<p class="text-sm text-gray-500">{promo.percentage}% of promotional revenue</p>
					</div>
					<div class="text-right">
						<p class="text-lg font-bold text-green-600">{formatCurrency(promo.revenue)}</p>
					</div>
				</div>
			{/each}
		</div>
	</div>

	<!-- Baseline Comparison -->
	<div class="bg-white rounded-lg shadow p-6">
		<h3 class="text-lg font-semibold text-gray-900 mb-4">Baseline Comparison</h3>
		<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
			<div class="text-center">
				<p class="text-2xl font-bold text-gray-900">{data?.baseline_comparison?.pre_promo_mrr ? formatCurrency(data.baseline_comparison.pre_promo_mrr) : '$0'}</p>
				<p class="text-sm text-gray-500">Pre-Promo MRR</p>
			</div>
			<div class="text-center">
				<p class="text-2xl font-bold text-green-600">{data?.baseline_comparison?.current_mrr ? formatCurrency(data.baseline_comparison.current_mrr) : '$0'}</p>
				<p class="text-sm text-gray-500">Current MRR</p>
			</div>
			<div class="text-center">
				<p class="text-2xl font-bold text-blue-600">+{data?.baseline_comparison?.promotional_lift || 0}%</p>
				<p class="text-sm text-gray-500">Promotional Lift</p>
			</div>
		</div>
	</div>
</div> 