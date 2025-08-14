<script lang="ts">
	import type { SubscriptionPlan } from '$lib/services/streaming-subscriptions';
	import { formatDateForDisplay } from '$lib/utils/date';
	import StripeIntegrationStatus from './StripeIntegrationStatus.svelte'; // Add Stripe status component

	export let plan: SubscriptionPlan;
	export let isOptimisticallyUpdating: (planId: string | number) => boolean;
	export let showStripeStatus: boolean = true; // Add prop to control Stripe status display
	
	// Callback props for Svelte 5 compatibility
	export let onEdit: (plan: SubscriptionPlan) => void;
	export let onToggleStatus: (plan: SubscriptionPlan) => void;
	export let onTogglePromotion: (plan: SubscriptionPlan) => void;
	export let onViewDetails: (plan: SubscriptionPlan) => void;

	// Format price
	$: formattedPrice = new Intl.NumberFormat('en-US', {
		style: 'currency',
		currency: plan.currency || 'USD'
	}).format(plan.price);

	// Format interval
	$: intervalText = plan.interval === 'month' ? 'Monthly' : plan.interval === 'year' ? 'Annual' : plan.interval;

	// Promotion status logic
	$: promotionStatus = (() => {
		if (plan.sub_type !== 'prmo') return null;
		
		const now = new Date();
		const startDate = plan.promotion_start_date ? new Date(plan.promotion_start_date) : null;
		const endDate = plan.promotion_end_date ? new Date(plan.promotion_end_date) : null;
		
		if (!startDate || !endDate) return null;
		
		if (now < startDate) {
			return { status: 'upcoming', label: 'Upcoming', color: 'blue' };
		} else if (now >= startDate && now <= endDate) {
			return { status: 'promoted', label: 'Promoted', color: 'yellow' };
		} else {
			return { status: 'expired', label: 'Expired', color: 'red' };
		}
	})();

	// Check if promotion is expired and should be deactivated
	$: isExpired = promotionStatus?.status === 'expired';

	// Format plan change history for display
	$: formattedHistory = plan.plan_change_history?.map((event: any) => ({
		...event,
		formattedDate: new Date(event.timestamp).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		})
	})) || [];

	// Get event type color
	function getEventTypeColor(eventType: string): string {
		const colors: Record<string, string> = {
			'plan_created': 'bg-blue-100 text-blue-800',
			'plan_updated': 'bg-yellow-100 text-yellow-800',
			'promotion_started': 'bg-green-100 text-green-800',
			'promotion_ended': 'bg-red-100 text-red-800',
			'status_activated': 'bg-green-100 text-green-800',
			'status_deactivated': 'bg-red-100 text-red-800',
			'price_changed': 'bg-purple-100 text-purple-800',
			'type_changed': 'bg-indigo-100 text-indigo-800',
			'promotion_expired': 'bg-orange-100 text-orange-800',
			'plan_deleted': 'bg-gray-100 text-gray-800'
		};
		return colors[eventType] || 'bg-gray-100 text-gray-800';
	}
</script>

<div class="plan-card" class:updating={isOptimisticallyUpdating(plan.id)} class:expired={isExpired}>
	{#if isOptimisticallyUpdating(plan.id)}
		<div class="optimistic-overlay">
			<div class="loading-spinner"></div>
			<span>Updating...</span>
		</div>
	{/if}

	<div class="plan-header">
		<div class="plan-info">
			<h3 class="plan-name">{plan.name}</h3>
			<p class="plan-description">{plan.description}</p>
		</div>
		<div class="plan-status">
			{#if plan.sub_type === "prmo" && promotionStatus}
				<span class="status-badge {promotionStatus.status}" class:upcoming={promotionStatus.status === 'upcoming'} class:promoted={promotionStatus.status === 'promoted'} class:expired={promotionStatus.status === 'expired'}>
					{promotionStatus.label}
				</span>
			{/if}
			<!-- Only show Active/Inactive badge for standard plans or currently promoted plans -->
			{#if plan.sub_type === "stnd" || (plan.sub_type === "prmo" && promotionStatus?.status === 'promoted')}
				<span class="status-badge" class:active={plan.is_active} class:inactive={!plan.is_active}>
					{plan.is_active ? 'Active' : 'Inactive'}
				</span>
			{/if}
		</div>
	</div>

	<!-- Promotion dates for promotional plans -->
	{#if plan.sub_type === "prmo" && (plan.promotion_start_date || plan.promotion_end_date)}
		<div class="promotion-dates">
			<h4>Promotion Period:</h4>
			<div class="date-range">
				{#if plan.promotion_start_date}
					<span class="date-label">Start:</span>
					<span class="date-value">{formatDateForDisplay(plan.promotion_start_date)}</span>
				{/if}
				{#if plan.promotion_end_date}
					<span class="date-label">End:</span>
					<span class="date-value">{formatDateForDisplay(plan.promotion_end_date)}</span>
				{/if}
			</div>
		</div>
	{/if}

	<div class="plan-details">
		<div class="plan-price">
			<span class="price-amount">{formattedPrice}</span>
			<span class="price-interval">/{intervalText}</span>
		</div>
		
		{#if plan.features && plan.features.length > 0}
			<div class="plan-features">
				<h4>Features:</h4>
				<ul>
					{#each plan.features as feature}
						<li>{feature}</li>
					{/each}
				</ul>
			</div>
		{/if}
	</div>

	<!-- Stripe Integration Status -->
	{#if showStripeStatus}
		<StripeIntegrationStatus planId={plan.id} planName={plan.name} />
	{/if}

	<div class="plan-footer">
		<div class="plan-actions">
			<button class="btn btn-secondary" on:click={() => onViewDetails(plan)} disabled={isOptimisticallyUpdating(plan.id)}>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
				</svg>
				Details
			</button>
			
			<button class="btn btn-secondary" on:click={() => onEdit(plan)} disabled={isOptimisticallyUpdating(plan.id)}>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
				</svg>
				Edit
			</button>
			
			<!-- Show Deactivate/Activate button only for standard plans (sub_type = stnd) -->
			{#if plan.sub_type === "stnd" || !plan.sub_type}
				<button 
					class="btn" 
					class:btn-secondary={plan.is_active} 
					class:btn-primary={!plan.is_active}
					on:click={() => onToggleStatus(plan)} 
					disabled={isOptimisticallyUpdating(plan.id)}
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
					</svg>
					{plan.is_active ? 'Deactivate' : 'Activate'}
				</button>
			{/if}
			
			<!-- Show promotion button for promotional plans (sub_type = prmo) -->
			{#if plan.sub_type === "prmo"}
				<button
					class="btn {plan.is_active ? 'btn-secondary' : 'btn-primary'}"
					on:click={() => onToggleStatus(plan)}
					disabled={isOptimisticallyUpdating(plan.id)}
				>
					{#if plan.is_active}
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
						</svg>
						Deactivate
					{:else}
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"></path>
						</svg>
						Activate
					{/if}
				</button>
			{/if}
			
		</div>
	</div>
</div>

<style>
	.plan-card {
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		padding: 1.25rem;
		transition: all 0.2s ease;
		position: relative;
		overflow: hidden;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
		max-width: 450px;
	}

	.plan-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
		border-color: #d1d5db;
	}

	.plan-card.updating {
		position: relative;
		opacity: 0.7;
		pointer-events: none;
	}

	.plan-card.expired {
		opacity: 0.6;
		border-color: #fca5a5;
		background: #fef2f2;
	}

	.plan-card.updating::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(37, 99, 235, 0.1);
		border-radius: 0.5rem;
		z-index: 1;
	}

	.optimistic-overlay {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(37, 99, 235, 0.1);
		backdrop-filter: blur(4px);
		border-radius: 0.5rem;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		z-index: 2;
		color: #2563eb;
		font-weight: 500;
	}

	.loading-spinner {
		width: 20px;
		height: 20px;
		border: 2px solid rgba(37, 99, 235, 0.3);
		border-top: 2px solid #2563eb;
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-bottom: 0.5rem;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.plan-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 1rem;
	}

	.plan-name {
		color: #111827;
		font-size: 1.25rem;
		font-weight: 600;
		margin: 0 0 0.5rem 0;
	}

	.plan-description {
		color: #6b7280;
		margin: 0;
		font-size: 0.875rem;
		line-height: 1.4;
	}

	.plan-status {
		display: flex;
		gap: 0.5rem;
		flex-wrap: wrap;
	}

	.status-badge {
		padding: 0.25rem 0.75rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.status-badge.active {
		background: #dcfce7;
		color: #166534;
		border: 1px solid #bbf7d0;
	}

	.status-badge.inactive {
		background: #f3f4f6;
		color: #374151;
		border: 1px solid #e5e7eb;
	}

	.status-badge.upcoming {
		background: #dbeafe;
		color: #1e40af;
		border: 1px solid #93c5fd;
	}

	.status-badge.promoted {
		background: #fef3c7;
		color: #92400e;
		border: 1px solid #fde68a;
	}

	.status-badge.expired {
		background: #fee2e2;
		color: #991b1b;
		border: 1px solid #fca5a5;
	}

	.promotion-dates {
		margin-bottom: 1rem;
		padding: 0.75rem;
		background: #f8fafc;
		border-radius: 0.375rem;
		border: 1px solid #e2e8f0;
	}

	.promotion-dates h4 {
		color: #374151;
		font-size: 0.875rem;
		font-weight: 600;
		margin: 0 0 0.5rem 0;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.date-range {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.date-label {
		color: #6b7280;
		font-size: 0.75rem;
		font-weight: 500;
		min-width: 40px;
	}

	.date-value {
		color: #111827;
		font-size: 0.875rem;
		font-weight: 600;
	}

	.plan-details {
		margin-bottom: 1.5rem;
	}

	.plan-price {
		margin-bottom: 1rem;
	}

	.price-amount {
		color: #111827;
		font-size: 1.5rem;
		font-weight: 700;
	}

	.price-interval {
		color: #6b7280;
		font-size: 1rem;
		font-weight: 500;
	}

	.plan-features {
		border: 1px solid #e5e7eb;
		padding: 1rem;
		border-radius: 0.5rem;
		background-color: #f8fafc;
	}


	.plan-features h4 {
		color: #111827;
		font-size: 0.875rem;
		font-weight: 600;
		margin: 0 0 0.5rem 0;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.plan-features ul {
		display: flex;
		flex-direction: row;
		flex-wrap: wrap;
		justify-content: center;
		gap: 2rem;
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.plan-features li {
		color: #6b7280;
		font-size: 0.875rem;
		padding: 0.25rem 0;
		position: relative;
		padding-left: 1rem;
	}

	.plan-features li::before {
		content: '✓';
		position: absolute;
		left: 0;
		color: #10b981;
		font-weight: bold;
	}

	.plan-history {
		margin-top: 1.5rem;
		padding-top: 1rem;
		border-top: 1px solid #e5e7eb;
	}

	.plan-history h4 {
		color: #111827;
		font-size: 0.875rem;
		font-weight: 600;
		margin: 0 0 0.5rem 0;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.history-timeline {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.history-event {
		background: #f8fafc;
		border: 1px solid #e2e8f0;
		border-radius: 0.375rem;
		padding: 0.75rem;
		position: relative;
	}

	.history-event::before {
		content: '';
		position: absolute;
		top: 0;
		left: -10px; /* Adjust as needed for the line */
		width: 2px;
		height: 100%;
		background: #d1d5db;
		z-index: -1;
	}

	.event-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.5rem;
	}

	.event-type {
		padding: 0.25rem 0.75rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.event-date {
		color: #6b7280;
		font-size: 0.75rem;
		font-weight: 500;
	}

	.event-description {
		color: #374151;
		font-size: 0.875rem;
		line-height: 1.4;
		margin-bottom: 0.25rem;
	}

	.analytics-tag {
		background: #e0f2fe;
		color: #1e40af;
		padding: 0.25rem 0.75rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.history-more {
		text-align: center;
		padding-top: 0.5rem;
	}

	.promotion-analytics {
		margin-top: 1.5rem;
		padding-top: 1rem;
		border-top: 1px solid #e5e7eb;
	}

	.promotion-analytics h4 {
		color: #111827;
		font-size: 0.875rem;
		font-weight: 600;
		margin: 0 0 0.5rem 0;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.analytics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		gap: 0.75rem;
	}

	.analytics-item {
		background: #f8fafc;
		border: 1px solid #e2e8f0;
		border-radius: 0.375rem;
		padding: 0.75rem;
		text-align: center;
	}

	.analytics-label {
		color: #6b7280;
		font-size: 0.75rem;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin-bottom: 0.25rem;
	}

	.analytics-value {
		color: #111827;
		font-size: 1.25rem;
		font-weight: 700;
	}

	.plan-footer {
		border-top: 1px solid #e5e7eb;
		padding-top: 1rem;
	}

	.plan-actions {
		display: flex;
		gap: 0.5rem;
		flex-wrap: wrap;
	}

	.plan-actions .btn {
		flex: 1;
		min-width: 100px;
	}

	.btn {
		padding: 0.5rem 1rem;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		border: none;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
	}

	.btn:hover {
		transform: translateY(-1px);
	}

	.btn:focus {
		outline: 2px solid #2563eb;
		outline-offset: 2px;
	}

	.btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
		transform: none !important;
	}

	.btn:disabled:hover {
		transform: none !important;
		box-shadow: none !important;
	}

	.btn-primary {
		background: #2563eb;
		color: white;
	}

	.btn-primary:hover {
		background: #1d4ed8;
		box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
	}

	.btn-secondary {
		background: #f3f4f6;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-secondary:hover {
		background: #e5e7eb;
		border-color: #9ca3af;
	}

	.btn-danger {
		background: #dc2626;
		color: white;
	}

	.btn-danger:hover {
		background: #b91c1c;
		box-shadow: 0 4px 12px rgba(220, 38, 38, 0.3);
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.plan-header {
			flex-direction: column;
			align-items: flex-start;
			gap: 1rem;
		}

		.plan-status {
			align-self: flex-end;
		}

		.plan-actions {
			flex-direction: column;
		}

		.plan-actions .btn {
			flex: none;
		}
	}
</style> 