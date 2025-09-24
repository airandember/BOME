<script lang="ts">
	export let title: string;
	export let class_title: string;
	export let icon: string;
	export let count: number;
	export let isActive: boolean = false;
	export let plans: any[] = [];
	export let onToggle: () => void = () => {};

	function toggleAccordion() {
		onToggle();
	}
</script>

<div class="accordion-container">
	<button 
		class="accordion-header header-{class_title}" 
		on:click={toggleAccordion}
		aria-expanded={isActive}
		aria-controls="accordion-content-{class_title}"
	>
		<div class="accordion-info">
			<h3 class="accordion-title">
				{@html icon}
				{title}
			</h3>
		</div>
		<div class="accordion-stats">
			<span class="plan-count">({count})</span>
		</div>
		<svg class="accordion-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
		</svg>
	</button>
	
	{#if isActive}
		<div 
			id="accordion-content-{class_title}"
			class="accordion-content content-{class_title}" 
			aria-hidden="false"
		>
			{#if plans.length > 0}
				<slot {plans} />
			{:else}
				<div class="empty-state">
					<svg class="empty-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
					</svg>
					<p>No {title.toLowerCase()} plans found</p>
					<span>Create a new plan to get started</span>
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	.accordion-container {
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		overflow: hidden;
		transition: all 0.2s ease;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
	}

	.accordion-container:hover {
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}

	.header-promoted-plans {
		background: #fff5ca !important;
	}

	.header-active-plans {
		background: #dbffee !important;
	}

	.header-inactive-plans {
		background: #ffe5d9 !important;
	}

	.header-active-offers {
		background: #ccf4fd !important;
		color: white;
	}

	.header-inactive-offers {
		background: #ffb0b3 !important;
	}

	.content-promoted-plans {
		background: #fdfded !important;
	}

	.content-active-plans {
		background: #f2fff9 !important;
	}

	.content-active-offers {
		background: #ecfbff !important;
	}

	.content-inactive-offers {
		background: #ffe6e7 !important;
	}

	.accordion-header {
		width: 100%;
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1.5rem;
		background: none;
		border: none;
		cursor: pointer;
		transition: all 0.2s ease;
		color: #111827;
	}

	.accordion-header:hover {
		background: #f9fafb;
	}

	.accordion-header:focus {
		outline: 2px solid #2563eb;
		outline-offset: 2px;
	}

	.accordion-info {
		display: flex;
		align-items: center;
		gap: 1rem;
		flex: 1;
	}

	.accordion-title {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		font-size: 1.125rem;
		font-weight: 600;
		color: #111827;
		margin: 0;
		line-height: 1.4;
	}

	.accordion-title :global(svg) {
		width: 1.5rem;
		height: 1.5rem;
		color: #6b7280;
	}

	.accordion-stats {
		display: flex;
		gap: 1rem;
		margin-right: 1rem;
		font-size: 0.875rem;
	}

	.plan-count {
		color: #111827;
		font-weight: 500;
		background: #f3f4f6;
		padding: 0.25rem 0.5rem;
		border-radius: 0.25rem;
		border: 1px solid #e5e7eb;
	}

	.accordion-icon {
		width: 1.25rem;
		height: 1.25rem;
		color: #6b7280;
		transition: transform 0.2s ease;
	}

	.accordion-header[aria-expanded="true"] .accordion-icon {
		transform: rotate(180deg);
	}

	.accordion-content {
		display: flex;
		flex-direction: row;
		flex-wrap: wrap;
		gap: 1rem;
		justify-content: center;
		padding: 1.5rem 1.5rem 1.5rem 1.5rem;
		border-top: 1px solid #e5e7eb;
	}

	.empty-state {
		text-align: center;
		padding: 3rem 1rem;
		color: #6b7280;
	}

	.empty-icon {
		width: 48px;
		height: 48px;
		margin: 0 auto 1rem;
		opacity: 0.5;
		color: #9ca3af;
	}

	.empty-state p {
		font-size: 1.125rem;
		font-weight: 600;
		margin: 0 0 0.5rem 0;
		color: #374151;
	}

	.empty-state span {
		font-size: 0.875rem;
		opacity: 0.8;
	}

	@media (max-width: 768px) {
		.accordion-header {
			padding: 1rem;
			flex-direction: column;
			align-items: flex-start;
			gap: 1rem;
		}

		.accordion-stats {
			margin-right: 0;
			align-self: flex-end;
		}

		.accordion-content {
			padding: 0 1rem 1rem 1rem;
		}
	}
</style> 
