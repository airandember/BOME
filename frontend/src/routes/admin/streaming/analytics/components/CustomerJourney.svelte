<script lang="ts">
	// Props
	export let data: any = null;
	export let period: string = '30d';
</script>

<div class="space-y-6">
	<!-- Journey Metrics -->
	<div class="bg-white rounded-lg shadow p-6">
		<h3 class="text-lg font-semibold text-gray-900 mb-4">Customer Journey Metrics</h3>
		<div class="space-y-4">
			{#each data?.journey_metrics || [] as metric}
				<div class="flex justify-between items-center p-4 bg-gray-50 rounded-lg">
					<div>
						<h4 class="font-medium text-gray-900">{metric.metric}</h4>
						<div class="flex space-x-4 mt-1">
							<span class="text-sm text-gray-500">Standard: {metric.standard}</span>
							<span class="text-sm text-green-600">Promotional: {metric.promotional}</span>
						</div>
					</div>
					<div class="text-right">
						{#if metric.improvement}
							<p class="text-sm font-medium text-green-600">+{metric.improvement}%</p>
						{:else if metric.difference}
							<p class="text-sm font-medium {metric.difference > 0 ? 'text-green-600' : 'text-red-600'}">{metric.difference > 0 ? '+' : ''}{metric.difference}%</p>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	</div>

	<!-- Net Impact Summary -->
	<div class="bg-white rounded-lg shadow p-6">
		<h3 class="text-lg font-semibold text-gray-900 mb-4">Net Business Impact</h3>
		<div class="text-center">
			<div class="inline-flex items-center px-4 py-2 rounded-full {data?.net_impact === 'positive' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}">
				<span class="text-lg mr-2">{data?.net_impact === 'positive' ? '📈' : '📉'}</span>
				<span class="font-medium">{data?.net_impact === 'positive' ? 'Positive Impact' : 'Negative Impact'}</span>
			</div>
			<p class="text-sm text-gray-500 mt-2">
				Promotional strategies are {data?.net_impact === 'positive' ? 'driving growth' : 'impacting performance'} 
				across customer journey metrics
			</p>
		</div>
	</div>
</div> 
