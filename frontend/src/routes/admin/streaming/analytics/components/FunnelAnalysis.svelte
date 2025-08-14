<script lang="ts">
	// Props
	export let data: any = null;
	export let period: string = '30d';

	// Format number
	function formatNumber(value: number): string {
		return new Intl.NumberFormat('en-US').format(value);
	}
</script>

<div class="space-y-6">
	<!-- Funnel Performance Summary -->
	<div class="bg-white rounded-lg shadow p-6">
		<h3 class="text-lg font-semibold text-gray-900 mb-4">Funnel Performance Summary</h3>
		<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
			<div class="text-center">
				<p class="text-2xl font-bold text-gray-900">{data?.conversion_rates?.standard || 0}%</p>
				<p class="text-sm text-gray-500">Standard Conversion</p>
			</div>
			<div class="text-center">
				<p class="text-2xl font-bold text-green-600">{data?.conversion_rates?.promotional || 0}%</p>
				<p class="text-sm text-gray-500">Promotional Conversion</p>
			</div>
			<div class="text-center">
				<p class="text-2xl font-bold text-blue-600">+{data?.conversion_rates?.lift || 0}%</p>
				<p class="text-sm text-gray-500">Conversion Lift</p>
			</div>
		</div>
	</div>

	<!-- Funnel Stages -->
	<div class="bg-white rounded-lg shadow p-6">
		<h3 class="text-lg font-semibold text-gray-900 mb-4">Funnel Stage Analysis</h3>
		<div class="space-y-4">
			{#each data?.stages || [] as stage}
				<div class="border rounded-lg p-4">
					<div class="flex justify-between items-center mb-3">
						<h4 class="font-medium text-gray-900">{stage.name}</h4>
						<div class="flex space-x-4">
							<span class="text-sm text-gray-500">Standard: {stage.standard?.toLocaleString()}</span>
							<span class="text-sm text-green-600">Promotional: {stage.promotional?.toLocaleString()}</span>
							<span class="text-sm text-blue-600">Lift: +{stage.lift}%</span>
						</div>
					</div>
					<div class="w-full bg-gray-200 rounded-full h-2">
						<div class="bg-blue-600 h-2 rounded-full" style="width: {Math.min((stage.promotional / stage.standard) * 100, 100)}%"></div>
					</div>
				</div>
			{/each}
		</div>
	</div>
</div> 