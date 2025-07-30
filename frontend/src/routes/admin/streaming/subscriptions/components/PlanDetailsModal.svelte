<script lang="ts">
	import type { SubscriptionPlan } from '$lib/services/streaming-subscriptions';
	import { formatDateForDisplay } from '$lib/utils/date';
	import { fade, fly } from 'svelte/transition';
	import HistoryFilter from './HistoryFilter.svelte';

	export let plan: SubscriptionPlan | null = null;
	export let isOpen = false;

	// Format price
	$: formattedPrice = plan ? new Intl.NumberFormat('en-US', {
		style: 'currency',
		currency: plan.currency || 'USD'
	}).format(plan.price) : '$0';

	// Format interval
	$: intervalText = plan?.interval === 'month' ? 'Monthly' : plan?.interval === 'year' ? 'Annual' : plan?.interval || '';

	// Format plan change history for display
	$: formattedHistory = plan?.plan_change_history?.map((event: any) => ({
		...event,
		formattedDate: new Date(event.timestamp).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		})
	})) || [];

	// Filtered history events
	let filteredHistory = formattedHistory;

	// Handle filter changes
	function handleFilterChange(filteredEvents: Record<string, any>[]) {
		filteredHistory = filteredEvents.map((event: any) => ({
			...event,
			formattedDate: new Date(event.timestamp).toLocaleDateString('en-US', {
				year: 'numeric',
				month: 'short',
				day: 'numeric',
				hour: '2-digit',
				minute: '2-digit'
			})
		}));
	}

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

	// Get user display name with fallback to localStorage
	function getUserDisplay(event: any): string {
		console.log('EVENT: ', event);
		if (!event.user_id) {
			console.log('NO USER ID');
			return 'System';
		}

		// Handle JSONB user data format (direct from localStorage)
		if (typeof event.user_id === 'string' && event.user_id.startsWith('{')) {
			try {
				const userData = JSON.parse(event.user_id);
				console.log('Parsed user data:', userData);
				
				// Handle system users
				if (userData.id === 'system' || userData.email === 'system') {
					if (userData.last_name === '(Auto-Expiration)') {
						return 'System (Auto-Expiration)';
					}
					return 'System';
				}

				// Handle regular users
				const firstName = userData.first_name || '';
				const lastName = userData.last_name || '';
				const role = userData.role || 'user';
				
				let displayName = '';
				if (firstName && lastName) {
					displayName = `${firstName} ${lastName}`;
				} else if (firstName) {
					displayName = firstName;
				} else if (userData.email) {
					displayName = userData.email;
				} else {
					displayName = `User ${userData.id}`;
				}

				// Add role badge
				const roleBadges: Record<string, string> = {
					'super_admin': '👑 Super Admin',
					'system_admin': '🔧 System Admin',
					'content_manager': '📝 Content Manager',
					'articles_manager': '📰 Articles Manager',
					'youtube_manager': '📺 YouTube Manager',
					'streaming_manager': '🎥 Streaming Manager',
					'events_manager': '🎪 Events Manager',
					'advertisement_manager': '📢 Ad Manager',
					'user_manager': '👥 User Manager',
					'analytics_manager': '📊 Analytics Manager',
					'financial_admin': '💰 Financial Admin',
					'admin': '⚡ Admin',
					'user': '👤 User',
					'system': '🤖 System',
					'dashboard': '🖥️ Dashboard'
				};

				const roleBadge = roleBadges[role] || `👤 ${role.charAt(0).toUpperCase() + role.slice(1)}`;
				return `${displayName} (${roleBadge})`;
			} catch (error) {
				console.error('Error parsing user data:', error);
				return 'System';
			}
		}

		// Handle legacy string format
		if (event.user_id === 'System' || event.user_id === 'System (Auto-Expiration)') {
			return event.user_id;
		}

		// Fallback
		return 'System';
	}

	// Format currency
	function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(amount);
	}

	// Format percentage
	function formatPercentage(value: number): string {
		return `${value.toFixed(2)}%`;
	}

	// Close modal
	function closeModal() {
		isOpen = false;
	}

	// Handle backdrop click
	function handleBackdropClick(event: MouseEvent) {
		if (event.target === event.currentTarget) {
			closeModal();
		}
	}

	// Debug logging
	$: {
		if (plan) {
			console.log('PlanDetailsModal: Plan data:', plan);
			console.log('PlanDetailsModal: plan_change_history:', plan.plan_change_history);
			console.log('PlanDetailsModal: formattedHistory:', formattedHistory);
			console.log('PlanDetailsModal: filteredHistory:', filteredHistory);
		}
	}

	// Log when modal opens
	$: if (isOpen && plan) {
		console.log('🔍 PlanDetailsModal: Modal opened for plan:', plan.name);
		console.log('🔍 PlanDetailsModal: Test buttons should be visible');
	}

	// Test function to verify localStorage data
	function testLocalStorageData() {
		console.log('=== TESTING LOCALSTORAGE ===');
		const userData = localStorage.getItem('bome_user_data');
		console.log('localStorage bome_user_data:', userData);
		if (userData) {
			try {
				const parsed = JSON.parse(userData);
				console.log('Parsed user data:', parsed);
			} catch (error) {
				console.error('Error parsing user data:', error);
			}
		} else {
			console.log('No user data found in localStorage');
		}
	}

	// Test the complete flow
	function testCompleteFlow() {
		console.log('=== TESTING COMPLETE FLOW ===');
		
		// 1. Test localStorage
		const userData = localStorage.getItem('bome_user_data');
		console.log('1. localStorage data:', userData);
		
		// 2. Test history events
		if (plan && plan.plan_change_history) {
			console.log('2. History events:', plan.plan_change_history);
			plan.plan_change_history.forEach((event, index) => {
				console.log(`   Event ${index}:`, event);
				console.log(`   User ID:`, event.user_id);
				console.log(`   Display:`, getUserDisplay(event));
			});
		} else {
			console.log('2. No history events found');
		}
		
		// 3. Test filtered history
		console.log('3. Filtered history:', filteredHistory);
	}
</script>

{#if isOpen && plan}
	<div class="modal-backdrop" on:click={handleBackdropClick} transition:fade={{ duration: 200 }}>
		<div class="modal-content" transition:fly={{ y: 50, duration: 300 }}>
			<h2>HOWDY BOYS</h2>
			<div class="modal-header">
				<h2 class="modal-title">Plan Details: {plan.name}</h2>
				<div class="header-actions">
					<button class="test-button" on:click={testLocalStorageData} style="background: red; color: white; padding: 8px; margin: 4px;">
						🔍 Test localStorage
					</button>
					<button class="test-button" on:click={testCompleteFlow} style="background: blue; color: white; padding: 8px; margin: 4px;">
						🔍 Test Flow
					</button>
					<button class="close-button" on:click={closeModal}>
						<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
						</svg>
					</button>
				</div>
			</div>

			<div class="modal-body">
				<!-- Debug Test Section -->
				<div class="debug-section" style="background: yellow; padding: 10px; margin-bottom: 20px; border: 2px solid red;">
					<h4 style="color: red; margin: 0 0 10px 0;">🔍 DEBUG TEST SECTION</h4>
					<button 
						on:click={testLocalStorageData}
						style="background: red; color: white; padding: 10px; margin: 5px; border: none; border-radius: 5px; cursor: pointer;"
					>
						🔍 Test localStorage
					</button>
					<button 
						on:click={testCompleteFlow}
						style="background: blue; color: white; padding: 10px; margin: 5px; border: none; border-radius: 5px; cursor: pointer;"
					>
						🔍 Test Complete Flow
					</button>
					<button 
						on:click={() => console.log('Modal is working!')}
						style="background: green; color: white; padding: 10px; margin: 5px; border: none; border-radius: 5px; cursor: pointer;"
					>
						✅ Test Modal
					</button>
					<button 
						on:click={async () => {
							if (!plan) return;
							console.log('Making a test change to the plan...');
							try {
								// Use the existing service directly - we're already authenticated in the dashboard
								const { StreamingSubscriptionService } = await import('$lib/services/streaming-subscriptions');
								const updatedPlan = await StreamingSubscriptionService.update({
									id: plan.id,
									description: plan.description + ' (Test change at ' + new Date().toLocaleTimeString() + ')'
								});
								console.log('Test change result:', updatedPlan);
								// Refresh the plan data
								plan = updatedPlan;
							} catch (error) {
								console.error('Test change failed:', error);
							}
						}}
						style="background: orange; color: white; padding: 10px; margin: 5px; border: none; border-radius: 5px; cursor: pointer;"
					>
						🧪 Test Change
					</button>
				</div>

				<!-- Plan Overview -->
				<div class="section">
					<h3 class="section-title">Plan Overview</h3>
					<div class="overview-grid">
						<div class="overview-item">
							<span class="label">Name</span>
							<span class="value">{plan.name}</span>
						</div>
						<div class="overview-item">
							<span class="label">Description</span>
							<span class="value">{plan.description}</span>
						</div>
						<div class="overview-item">
							<span class="label">Price</span>
							<span class="value">{formattedPrice}/{intervalText}</span>
						</div>
						<div class="overview-item">
							<span class="label">Status</span>
							<span class="value">
								<span class="status-badge {plan.is_active ? 'active' : 'inactive'}">
									{plan.is_active ? 'Active' : 'Inactive'}
								</span>
							</span>
						</div>
						<div class="overview-item">
							<span class="label">Type</span>
							<span class="value">
								<span class="type-badge {plan.sub_type === 'prmo' ? 'promotional' : 'standard'}">
									{plan.sub_type === 'prmo' ? 'Promotional' : 'Standard'}
								</span>
							</span>
						</div>
						<div class="overview-item">
							<span class="label">Created</span>
							<span class="value">{formatDateForDisplay(plan.created_at)}</span>
						</div>
						<div class="overview-item">
							<span class="label">Last Updated</span>
							<span class="value">{formatDateForDisplay(plan.updated_at)}</span>
						</div>
					</div>
				</div>

				<!-- Features -->
				{#if plan.features && plan.features.length > 0}
					<div class="section">
						<h3 class="section-title">Features</h3>
						<div class="features-list">
							{#each plan.features as feature}
								<div class="feature-item">
									<svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" style="max-height: 30px; width: auto">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" ></path>
									</svg>
									<span>{feature}</span>
								</div>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Promotion Details -->
				{#if plan.sub_type === 'prmo'}
					<div class="section">
						<h3 class="section-title">Promotion Details</h3>
						<div class="promotion-details">
							{#if plan.promotion_start_date}
								<div class="promotion-item">
									<span class="label">Start Date</span>
									<span class="value">{formatDateForDisplay(plan.promotion_start_date)}</span>
								</div>
							{/if}
							{#if plan.promotion_end_date}
								<div class="promotion-item">
									<span class="label">End Date</span>
									<span class="value">{formatDateForDisplay(plan.promotion_end_date)}</span>
								</div>
							{/if}
						</div>
					</div>
				{/if}

				<!-- Promotion Analytics -->
				{#if plan.sub_type === 'prmo' && plan.promotion_metadata}
					<div class="section">
						<h3 class="section-title">Promotion Analytics</h3>
						<div class="analytics-grid">
							<div class="analytics-card">
								<span class="analytics-label">Revenue Generated</span>
								<span class="analytics-value">{formatCurrency(plan.promotion_metadata.performance_metrics?.total_revenue_generated || 0)}</span>
							</div>
							<div class="analytics-card">
								<span class="analytics-label">Conversion Rate</span>
								<span class="analytics-value">{formatPercentage(plan.promotion_metadata.performance_metrics?.average_conversion_rate || 0)}</span>
							</div>
							<div class="analytics-card">
								<span class="analytics-label">Customer Acquisition Cost</span>
								<span class="analytics-value">{formatCurrency(plan.promotion_metadata.performance_metrics?.customer_acquisition_cost || 0)}</span>
							</div>
							<div class="analytics-card">
								<span class="analytics-label">Revenue Per Customer</span>
								<span class="analytics-value">{formatCurrency(plan.promotion_metadata.performance_metrics?.revenue_per_customer || 0)}</span>
							</div>
							<div class="analytics-card">
								<span class="analytics-label">Duration</span>
								<span class="analytics-value">{plan.promotion_metadata.current_promotion?.duration_days || 0} days</span>
							</div>
							<div class="analytics-card">
								<span class="analytics-label">Promotions Run</span>
								<span class="analytics-value">{plan.promotion_metadata.historical_data?.total_promotions_run || 1}</span>
							</div>
						</div>
					</div>
				{/if}

				<!-- Plan Change History -->
				{#if formattedHistory.length > 0}
					<div class="section">
						<h3 class="section-title">Change History ({filteredHistory.length} events)</h3>
						<HistoryFilter 
							events={plan?.plan_change_history || []} 
							onFilterChange={handleFilterChange} 
						/>
						<div class="history-timeline">
							{#each filteredHistory as event}
								<div class="history-event">
									<div class="event-header">
										<span class="event-type {getEventTypeColor(event.event_type)}">
											{event.event_type?.replace(/_/g, ' ').replace(/\b\w/g, (l: string) => l.toUpperCase())}
										</span>
										<span class="event-date">{event.formattedDate}</span>
									</div>
									<p class="event-description">{event.description}</p>
									{#if event.user_id}
										<div class="event-user">
											<span class="user-badge">👤 {getUserDisplay(event)}</span>
										</div>
									{/if}
									{#if event.analytics_tag}
										<span class="analytics-tag">Analytics: {event.analytics_tag}</span>
									{/if}
									{#if event.old_values && Object.keys(event.old_values).length > 0}
										<div class="event-changes">
											<div class="changes-header">Changes:</div>
											{#each Object.entries(event.old_values) as [key, oldValue]}
												<div class="change-item">
													<span class="change-label">{key}:</span>
													<span class="change-old">{String(oldValue)}</span>
													<span class="change-arrow">→</span>
													<span class="change-new">{String(event.new_values?.[key] || 'N/A')}</span>
												</div>
											{/each}
										</div>
									{/if}
								</div>
							{/each}
						</div>
					</div>
				{:else}
					<div class="section">
						<h3 class="section-title">Change History</h3>
						<div class="empty-state">
							<p>No change history available for this plan.</p>
						</div>
					</div>
				{/if}
			</div>
		</div>
	</div>
{/if}

<style>
	.modal-backdrop {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
		padding: 1rem;
	}

	.modal-content {
		background: white;
		border-radius: 0.5rem;
		box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
		max-width: 1800px;
		width: 100%;
		max-height: 90vh;
		overflow: hidden;
		display: flex;
		flex-direction: column;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1.5rem;
		border-bottom: 1px solid #e5e7eb;
	}

	.header-actions {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.test-button {
		background: #3b82f6;
		color: white;
		border: none;
		padding: 0.25rem 0.75rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		font-weight: 500;
		cursor: pointer;
		transition: background-color 0.2s;
	}

	.test-button:hover {
		background: #2563eb;
	}

	.modal-title {
		font-size: 1.25rem;
		font-weight: 600;
		color: #111827;
		margin: 0;
	}

	.close-button {
		background: none;
		border: none;
		color: #6b7280;
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 0.375rem;
		transition: all 0.2s ease;
	}

	.close-button:hover {
		background: #f3f4f6;
		color: #374151;
	}

	.modal-body {
		padding: 1.5rem;
		overflow-y: auto;
		flex: 1;
	}

	.section {
		margin-bottom: 2rem;
	}

	.section-title {
		font-size: 1.125rem;
		font-weight: 600;
		color: #111827;
		margin: 0 0 1rem 0;
		border-bottom: 2px solid #e5e7eb;
		padding-bottom: 0.5rem;
	}

	.overview-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 1rem;
	}

	.overview-item {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.label {
		font-size: 0.875rem;
		font-weight: 500;
		color: #6b7280;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.value {
		font-size: 1rem;
		color: #111827;
		font-weight: 500;
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
	}

	.status-badge.inactive {
		background: #f3f4f6;
		color: #374151;
	}

	.type-badge {
		padding: 0.25rem 0.75rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.type-badge.promotional {
		background: #fef3c7;
		color: #92400e;
	}

	.type-badge.standard {
		background: #dbeafe;
		color: #1e40af;
	}

	.features-list {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.feature-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem;
		background: #f8fafc;
		border-radius: 0.375rem;
	}

	.promotion-details {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
	}

	.promotion-item {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.analytics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
	}

	.analytics-card {
		background: #f8fafc;
		border: 1px solid #e2e8f0;
		border-radius: 0.5rem;
		padding: 1rem;
		text-align: center;
	}

	.analytics-label {
		display: block;
		color: #6b7280;
		font-size: 0.875rem;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin-bottom: 0.5rem;
	}

	.analytics-value {
		display: block;
		color: #111827;
		font-size: 1.5rem;
		font-weight: 700;
	}

	.history-timeline {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.history-event {
		background: #f8fafc;
		border: 1px solid #e2e8f0;
		border-radius: 0.5rem;
		padding: 1rem;
		position: relative;
	}

	.event-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.75rem;
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
		font-size: 0.875rem;
		font-weight: 500;
	}

	.event-description {
		color: #374151;
		font-size: 1rem;
		line-height: 1.5;
		margin-bottom: 0.5rem;
	}

	.analytics-tag {
		display: inline-block;
		background: #e0f2fe;
		color: #1e40af;
		padding: 0.25rem 0.75rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin-bottom: 0.5rem;
	}

	.event-user {
		margin-top: 0.5rem;
		padding-top: 0.5rem;
		border-top: 1px solid #e2e8f0;
	}

	.user-badge {
		padding: 0.25rem 0.75rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		background: #e0f2fe;
		color: #1e40af;
	}

	.event-changes {
		margin-top: 0.75rem;
		padding-top: 0.75rem;
		border-top: 1px solid #e2e8f0;
	}

	.changes-header {
		font-size: 0.875rem;
		font-weight: 600;
		color: #374151;
		margin-bottom: 0.5rem;
	}

	.change-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-bottom: 0.25rem;
		font-size: 0.875rem;
	}

	.change-label {
		font-weight: 600;
		color: #6b7280;
		min-width: 80px;
	}

	.change-old {
		color: #dc2626;
		text-decoration: line-through;
	}

	.change-arrow {
		color: #6b7280;
		font-weight: 600;
	}

	.change-new {
		color: #059669;
		font-weight: 600;
	}

	.empty-state {
		text-align: center;
		padding: 2rem;
		color: #6b7280;
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.modal-content {
			margin: 0;
			border-radius: 0;
			max-height: 100vh;
		}

		.overview-grid {
			grid-template-columns: 1fr;
		}

		.analytics-grid {
			grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		}

		.event-header {
			flex-direction: column;
			align-items: flex-start;
			gap: 0.5rem;
		}
	}
</style> 