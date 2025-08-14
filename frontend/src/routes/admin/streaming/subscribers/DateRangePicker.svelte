<script lang="ts">
	export let label = '';
	export let startDate: string = '';
	export let endDate: string = '';
	export let placeholder = 'Select date range';
	export let onDateRangeChange: (startDate: string, endDate: string) => void = () => {};

	// Track previous values to detect changes
	let previousStartDate = startDate;
	let previousEndDate = endDate;

	// Reactive statement to detect changes and call callback
	$: {
		if (startDate !== previousStartDate || endDate !== previousEndDate) {
			console.log('📅 DateRangePicker: Date changed:', { startDate, endDate, previousStartDate, previousEndDate });
			onDateRangeChange(startDate, endDate);
			previousStartDate = startDate;
			previousEndDate = endDate;
		}
	}

	function clearDates() {
		startDate = '';
		endDate = '';
	}
</script>

<div class="date-range-picker">
	<label class="picker-label">{label}</label>
	<div class="date-inputs">
		<div class="date-input-group">
			<label>From:</label>
			<input 
				type="date" 
				bind:value={startDate}
				class="date-input"
			/>
		</div>
		<div class="date-input-group">
			<label>To:</label>
			<input 
				type="date" 
				bind:value={endDate}
				class="date-input"
			/>
		</div>
		{#if startDate || endDate}
			<button class="clear-button" on:click={clearDates}>
				Clear
			</button>
		{/if}
	</div>
</div>

<style>
	.date-range-picker {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.picker-label {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--text-primary);
	}

	.date-inputs {
		display: flex;
		flex-direction: column;
		align-items: end;
		flex-wrap: wrap;
	}

	.date-input-group {
		display: flex;
		flex-direction: row;
		gap: 0.25rem;
		align-items: center;
	}

	.date-input-group label {
		font-size: 0.75rem;
		color: var(--text-secondary);
	}

	.date-input {
		padding: 0.5rem;
		border: 1px solid var(--border-color);
		border-radius: 4px;
		background: var(--bg-secondary);
		color: var(--text-primary);
		font-size: 0.875rem;
	}

	.clear-button {
		padding: 0.5rem 1rem;
		border: 1px solid var(--border-color);
		border-radius: 4px;
		background: var(--bg-secondary);
		color: var(--text-secondary);
		cursor: pointer;
		font-size: 0.875rem;
		transition: all 0.2s;
	}

	.clear-button:hover {
		background: var(--error-color);
		color: white;
	}
</style> 