<script lang="ts">
	// Props
	export let promotionPlans: any[] = [];
	export let period: string = '30d';

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
			if (plan.performanceMetrics?.total_revenue_generated) {
				return total + plan.performanceMetrics.total_revenue_generated;
			}
			return total;
		}, 0);
	}

	// Calculate average conversion rate
	function calculateAverageConversionRate(): number {
		const plansWithMetrics = promotionPlans.filter((plan: any) => 
			plan.performanceMetrics?.average_conversion_rate
		);
		
		if (plansWithMetrics.length === 0) return 0;
		
		const totalRate = plansWithMetrics.reduce((sum: number, plan: any) => 
			sum + plan.performanceMetrics.average_conversion_rate, 0
		);
		
		return totalRate / plansWithMetrics.length;
	}
</script>

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