<script lang="ts">
	import type { SubscriptionOffer } from '$lib/services/subscription-offers';
	import type { SubscriptionPlan } from '$lib/services/streaming-subscriptions';
	import { formatDateForDisplay } from '$lib/utils/date';
	import HistoryFilter from './HistoryFilter.svelte';

	export let offer: SubscriptionOffer | null = null;
	export let subscriptionPlans: SubscriptionPlan[] = [];
	export let isOpen = false;

	// Utility functions with proper plan lookup
	function getPlanName(planId: number): string {
		const plan = subscriptionPlans.find(p => p.id === planId.toString());
		return plan ? plan.name : `Plan ${planId}`;
	}

	function getItemName(itemId: string | null | undefined): string {
		if (!itemId) return 'No specific item';
		
		const itemTypes: Record<string, string> = {
			'ebook': 'eBook',
			'dvd': 'DVD', 
			'expo_ticket': 'Expo Ticket',
			'1': 'eBook',
			'2': 'DVD',
			'3': 'Expo Ticket'
		};
		
		return itemTypes[itemId] || itemId;
	}

	// Format discount display
	$: discountDisplay = offer?.off_discount_type === 'percentage' 
		? `${offer?.off_discount_value || 0}% OFF`
		: `$${offer?.off_discount_value || 0} OFF`;

	// Format usage display
	$: usageDisplay = offer ? `${offer.off_current_uses}/${offer.off_max_uses || '∞'}` : '0/0';

	// Format offer history for display
	$: formattedHistory = offer?.offer_history?.map((event: any) => ({
		...event,
		formattedDate: new Date(event.created_at).toLocaleDateString('en-US', {
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
			formattedDate: new Date(event.created_at).toLocaleDateString('en-US', {
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
			'offer_created': 'bg-blue-100 text-blue-800',
			'offer_updated': 'bg-yellow-100 text-yellow-800',
			'offer_deleted': 'bg-red-100 text-red-800',
			'offer_accepted': 'bg-green-100 text-green-800',
			'offer_applied': 'bg-green-100 text-green-800',
			'offer_viewed': 'bg-purple-100 text-purple-800',
			'offer_expired': 'bg-orange-100 text-orange-800',
			'status_toggled': 'bg-indigo-100 text-indigo-800'
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
		if (offer) {
			console.log('OfferDetailsModal: Offer data:', offer);
			console.log('OfferDetailsModal: offer_history:', offer.offer_history);
			console.log('OfferDetailsModal: formattedHistory:', formattedHistory);
			console.log('OfferDetailsModal: filteredHistory:', filteredHistory);
		}
	}

	// Log when modal opens
	$: if (isOpen && offer) {
		console.log('🔍 OfferDetailsModal: Modal opened for offer:', offer.off_name);
		console.log('🔍 OfferDetailsModal: Test buttons should be visible');
	}
</script>

<!-- Modal Backdrop -->
{#if isOpen && offer}
	<div class="modal-backdrop" on:click={handleBackdropClick}>
		<div class="modal-content" on:click|stopPropagation>
			<!-- Modal Header -->
			<div class="modal-header">
				<div class="modal-title-section">
					<h2 class="modal-title">Offer Details</h2>
					<p class="modal-subtitle">{offer.off_name}</p>
				</div>
				<button type="button" class="modal-close" on:click={closeModal} aria-label="Close">
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
					</svg>
				</button>
			</div>

			<!-- Modal Body -->
			<div class="modal-body">
				<!-- Offer Information -->
				<div class="offer-info-section">
					<div class="info-grid">
						<div class="info-item">
							<span class="info-label">Offer Name:</span>
							<span class="info-value">{offer.off_name}</span>
						</div>
						<div class="info-item">
							<span class="info-label">Plan:</span>
							<span class="info-value">{getPlanName(offer.plan_id)}</span>
						</div>
						<div class="info-item">
							<span class="info-label">Discount:</span>
							<span class="info-value discount">{discountDisplay}</span>
						</div>
						<div class="info-item">
							<span class="info-label">Status:</span>
							<span class="info-value status {offer.is_active ? 'active' : 'inactive'}">
								{offer.is_active ? 'Active' : 'Inactive'}
							</span>
						</div>
						{#if offer.item_id}
							<div class="info-item">
								<span class="info-label">Item:</span>
								<span class="info-value">{getItemName(offer.item_id)}</span>
							</div>
						{/if}
						<div class="info-item">
							<span class="info-label">Usage:</span>
							<span class="info-value">{usageDisplay}</span>
						</div>
						<div class="info-item">
							<span class="info-label">Priority:</span>
							<span class="info-value">Level {offer.off_priority}</span>
						</div>
						<div class="info-item">
							<span class="info-label">Auto-Apply:</span>
							<span class="info-value">{offer.off_auto_apply ? 'Yes' : 'No'}</span>
						</div>
						{#if offer.off_code}
							<div class="info-item">
								<span class="info-label">Code:</span>
								<span class="info-value code">{offer.off_code}</span>
							</div>
						{/if}
						{#if offer.offer_start_date || offer.off_end_date}
							<div class="info-item full-width">
								<span class="info-label">Valid Period:</span>
								<span class="info-value">
									{offer.offer_start_date ? formatDateForDisplay(offer.offer_start_date) : 'No start'}
									{' → '}
									{offer.off_end_date ? formatDateForDisplay(offer.off_end_date) : 'No end'}
								</span>
							</div>
						{/if}
						{#if offer.off_description}
							<div class="info-item full-width">
								<span class="info-label">Description:</span>
								<span class="info-value">{offer.off_description}</span>
							</div>
						{/if}
						{#if offer.off_terms_conditions}
							<div class="info-item full-width">
								<span class="info-label">Terms & Conditions:</span>
								<span class="info-value">{offer.off_terms_conditions}</span>
							</div>
						{/if}
						{#if offer.off_target}
							<div class="info-item">
								<span class="info-label">Target:</span>
								<span class="info-value">{offer.off_target}</span>
							</div>
						{/if}
					</div>
				</div>

				<!-- History Section -->
				<div class="history-section">
					<div class="history-header">
						<h3 class="history-title">Offer History</h3>
						<span class="history-count">{filteredHistory.length} events</span>
					</div>

					<!-- History Filter -->
					<HistoryFilter 
						events={formattedHistory} 
						onFilterChange={handleFilterChange}
					/>

					<!-- History Events -->
					<div class="history-events">
						{#if filteredHistory.length > 0}
							{#each filteredHistory as event (event.id)}
								<div class="history-event">
									<div class="event-header">
										<span class="event-type {getEventTypeColor(event.event_type)}">
											{event.event_type.replace('_', ' ').toUpperCase()}
										</span>
										<span class="event-date">{event.formattedDate}</span>
									</div>
									<div class="event-user">
										{getUserDisplay(event)}
									</div>
									{#if event.description}
										<div class="event-description">
											{event.description}
										</div>
									{/if}
									{#if event.metadata}
										<div class="event-metadata">
											<strong>Metadata:</strong> {JSON.stringify(event.metadata)}
										</div>
									{/if}
								</div>
							{/each}
						{:else}
							<div class="no-history">
								<p>No history events found.</p>
							</div>
						{/if}
					</div>
				</div>
			</div>

			<!-- Modal Footer -->
			<div class="modal-footer">
				<button type="button" class="btn btn-secondary" on:click={closeModal}>
					Close
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.modal-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 50;
		padding: 1rem;
	}

	.modal-content {
		background: white;
		border-radius: 0.5rem;
		border: 1px solid #e5e7eb;
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.25);
		width: 100%;
		max-width: 1500px;
		max-height: 90vh;
		overflow-y: auto;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		padding: 1.5rem 1.5rem 0 1.5rem;
		border-bottom: 1px solid #e5e7eb;
		margin-bottom: 1.5rem;
	}

	.modal-title-section {
		flex: 1;
	}

	.modal-title {
		font-size: 1.5rem;
		font-weight: 700;
		color: #111827;
		margin: 0 0 0.25rem 0;
	}

	.modal-subtitle {
		font-size: 1rem;
		color: #6b7280;
		margin: 0;
	}

	.modal-close {
		background: none;
		border: none;
		color: #6b7280;
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 0.375rem;
		transition: all 0.2s ease;
	}

	.modal-close:hover {
		background: #f3f4f6;
		color: #374151;
	}

	.modal-body {
		padding: 0 1.5rem;
	}

	.offer-info-section {
		margin-bottom: 2rem;
	}

	.info-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 1rem;
	}

	.info-item {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.info-item.full-width {
		grid-column: 1 / -1;
	}

	.info-label {
		font-size: 0.875rem;
		font-weight: 600;
		color: #6b7280;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.info-value {
		font-size: 1rem;
		color: #111827;
		font-weight: 500;
	}

	.info-value.discount {
		color: #059669;
		font-weight: 600;
	}

	.info-value.status.active {
		color: #059669;
		font-weight: 600;
	}

	.info-value.status.inactive {
		color: #dc2626;
		font-weight: 600;
	}

	.info-value.code {
		font-family: monospace;
		background: #f3f4f6;
		padding: 0.25rem 0.5rem;
		border-radius: 0.25rem;
		font-size: 0.875rem;
	}

	.history-section {
		border-top: 1px solid #e5e7eb;
		padding-top: 1.5rem;
	}

	.history-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
	}

	.history-title {
		font-size: 1.25rem;
		font-weight: 600;
		color: #111827;
		margin: 0;
	}

	.history-count {
		font-size: 0.875rem;
		color: #6b7280;
		background: #f3f4f6;
		padding: 0.25rem 0.5rem;
		border-radius: 0.25rem;
	}

	.history-events {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.history-event {
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		padding: 1rem;
		background: #f9fafb;
	}

	.event-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.5rem;
	}

	.event-type {
		padding: 0.25rem 0.5rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.event-date {
		font-size: 0.875rem;
		color: #6b7280;
	}

	.event-user {
		font-size: 0.875rem;
		color: #374151;
		margin-bottom: 0.5rem;
		font-weight: 500;
	}

	.event-description {
		font-size: 0.875rem;
		color: #6b7280;
		margin-bottom: 0.5rem;
	}

	.event-metadata {
		font-size: 0.75rem;
		color: #9ca3af;
		background: #f3f4f6;
		padding: 0.5rem;
		border-radius: 0.25rem;
		font-family: monospace;
	}

	.no-history, .no-offer {
		text-align: center;
		padding: 2rem;
		color: #6b7280;
	}

	.modal-footer {
		display: flex;
		justify-content: flex-end;
		padding: 1.5rem;
		border-top: 1px solid #e5e7eb;
		margin-top: 1.5rem;
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

	.btn-secondary {
		background: #f3f4f6;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-secondary:hover {
		background: #e5e7eb;
		border-color: #9ca3af;
	}

	@media (max-width: 768px) {
		.modal-content {
			margin: 0.5rem;
			max-height: calc(100vh - 1rem);
		}

		.info-grid {
			grid-template-columns: 1fr;
		}

		.event-header {
			flex-direction: column;
			align-items: flex-start;
			gap: 0.5rem;
		}
	}
</style> 
