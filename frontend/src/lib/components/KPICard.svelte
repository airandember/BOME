<script lang="ts">
	export let title: string;
	export let value: number | string;
	export let format: 'number' | 'currency' | 'percentage' | 'text' = 'number';
	export let trend: string | null = null;
	export let icon: string = '📊';
	export let color: 'blue' | 'green' | 'yellow' | 'red' | 'purple' | 'gray' = 'blue';
	export let loading = false;
	
	function formatValue(val: number | string, fmt: string): string {
		if (loading) return '...';
		if (val === null || val === undefined) return '-';
		
		switch (fmt) {
			case 'currency':
				return new Intl.NumberFormat('en-US', {
					style: 'currency',
					currency: 'USD',
					minimumFractionDigits: 0,
					maximumFractionDigits: 0
				}).format(Number(val));
			
			case 'percentage':
				return `${Number(val).toFixed(1)}%`;
			
			case 'number':
				return Number(val).toLocaleString();
			
			default:
				return String(val);
		}
	}
	
	function getTrendColor(trendValue: string): string {
		if (trendValue.startsWith('+')) return 'text-green-600';
		if (trendValue.startsWith('-')) return 'text-red-600';
		return 'text-gray-600';
	}
	
	function getColorClasses(colorName: string): { bg: string; text: string; icon: string } {
		const colors: Record<string, { bg: string; text: string; icon: string }> = {
			blue: { bg: 'bg-blue-50', text: 'text-blue-900', icon: 'bg-blue-500' },
			green: { bg: 'bg-green-50', text: 'text-green-900', icon: 'bg-green-500' },
			yellow: { bg: 'bg-yellow-50', text: 'text-yellow-900', icon: 'bg-yellow-500' },
			red: { bg: 'bg-red-50', text: 'text-red-900', icon: 'bg-red-500' },
			purple: { bg: 'bg-purple-50', text: 'text-purple-900', icon: 'bg-purple-500' },
			gray: { bg: 'bg-gray-50', text: 'text-gray-900', icon: 'bg-gray-500' }
		};
		return colors[colorName] || colors.blue;
	}
	
	$: colorClasses = getColorClasses(color);
	$: formattedValue = formatValue(value, format);
</script>

<div class="kpi-card {colorClasses.bg}">
	<div class="kpi-content">
		<div class="kpi-header">
			<div class="kpi-icon {colorClasses.icon}">
				{icon}
			</div>
			<div class="kpi-title">
				{title}
			</div>
		</div>
		
		<div class="kpi-value {colorClasses.text}">
			{#if loading}
				<div class="loading-skeleton"></div>
			{:else}
				{formattedValue}
			{/if}
		</div>
		
		{#if trend && !loading}
			<div class="kpi-trend {getTrendColor(trend)}">
				<span class="trend-icon">
					{trend.startsWith('+') ? '↗️' : trend.startsWith('-') ? '↘️' : '➡️'}
				</span>
				{trend}
			</div>
		{/if}
	</div>
</div>

<style>
	.kpi-card {
		border-radius: 0.75rem;
		padding: 1.5rem;
		border: 1px solid rgba(0, 0, 0, 0.05);
		transition: all 0.2s ease;
		position: relative;
		overflow: hidden;
	}
	
	.kpi-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}
	
	.kpi-content {
		position: relative;
		z-index: 1;
	}
	
	.kpi-header {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 1rem;
	}
	
	.kpi-icon {
		width: 2.5rem;
		height: 2.5rem;
		border-radius: 0.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 1.25rem;
		color: white;
	}
	
	.kpi-title {
		font-size: 0.875rem;
		font-weight: 500;
		color: #6b7280;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	
	.kpi-value {
		font-size: 2rem;
		font-weight: 700;
		line-height: 1;
		margin-bottom: 0.5rem;
	}
	
	.kpi-trend {
		display: flex;
		align-items: center;
		gap: 0.25rem;
		font-size: 0.875rem;
		font-weight: 500;
	}
	
	.trend-icon {
		font-size: 0.75rem;
	}
	
	.loading-skeleton {
		height: 2rem;
		background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
		background-size: 200% 100%;
		animation: loading 1.5s infinite;
		border-radius: 0.25rem;
		width: 60%;
	}
	
	@keyframes loading {
		0% {
			background-position: 200% 0;
		}
		100% {
			background-position: -200% 0;
		}
	}
	
	/* Responsive design */
	@media (max-width: 768px) {
		.kpi-card {
			padding: 1rem;
		}
		
		.kpi-value {
			font-size: 1.5rem;
		}
		
		.kpi-icon {
			width: 2rem;
			height: 2rem;
			font-size: 1rem;
		}
	}
</style>
