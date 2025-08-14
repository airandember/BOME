<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';

	let summary: any = null;
	let loading = true;
	let error = '';
	let statusFilter = 'all';
	let showSubscriptionModal = false;
	let selectedSubscription: any = null;

	export let data: any = null;

	onMount(async () => {
		if (data) {
			summary = data;
			loading = false;
		} else {
			await fetchSummary();
		}
	});

	async function fetchSummary() {
		try {
			loading = true;
			error = '';
			const res = await apiRequest('/admin/streaming/stripe/summary');
			if (res.ok) {
				const data = await res.json();
				summary = data.summary;
			} else {
				error = 'Failed to load subscriptions';
			}
		} catch (err) {
			error = 'Failed to load subscriptions';
			console.error(err);
		} finally {
			loading = false;
		}
	}

	// Currency formatting utility
	function formatCurrency(amount: number, currency: string = 'USD'): string {
		return new Intl.NumberFormat('en-US', { 
			style: 'currency', 
			currency: currency.toUpperCase() 
		}).format(amount); // Amount is already in dollars for subscriptions
	}

	function formatDate(dateString: string): string {
		if (!dateString) return 'N/A';
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	function formatDateTime(dateString: string): string {
		if (!dateString) return 'N/A';
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function getStatusColor(status: string): string {
		switch (status.toLowerCase()) {
			case 'active':
				return 'var(--success)';
			case 'trialing':
				return 'var(--primary)';
			case 'past_due':
				return 'var(--warning)';
			case 'canceled':
			case 'cancelled':
				return 'var(--text-muted)';
			case 'unpaid':
				return 'var(--error)';
			case 'incomplete':
			case 'incomplete_expired':
				return 'var(--warning)';
			default:
				return 'var(--text-muted)';
		}
	}

	function getStatusText(status: string): string {
		switch (status.toLowerCase()) {
			case 'active':
				return 'Active';
			case 'trialing':
				return 'Trialing';
			case 'past_due':
				return 'Past Due';
			case 'canceled':
			case 'cancelled':
				return 'Canceled';
			case 'unpaid':
				return 'Unpaid';
			case 'incomplete':
				return 'Incomplete';
			case 'incomplete_expired':
				return 'Incomplete Expired';
			default:
				return status.charAt(0).toUpperCase() + status.slice(1);
		}
	}

	function getStatusIcon(status: string): string {
		switch (status.toLowerCase()) {
			case 'active':
				return '✅';
			case 'trialing':
				return '🆓';
			case 'past_due':
				return '⚠️';
			case 'canceled':
			case 'cancelled':
				return '❌';
			case 'unpaid':
				return '💳';
			case 'incomplete':
				return '⏳';
			case 'incomplete_expired':
				return '🚫';
			default:
				return '🔄';
		}
	}

	function getBillingCycleText(interval: string, intervalCount: number = 1): string {
		if (intervalCount === 1) {
			switch (interval?.toLowerCase()) {
				case 'day':
					return 'Daily';
				case 'week':
					return 'Weekly';
				case 'month':
					return 'Monthly';
				case 'year':
					return 'Yearly';
				default:
					return interval || 'Unknown';
			}
		} else {
			switch (interval?.toLowerCase()) {
				case 'day':
					return `Every ${intervalCount} days`;
				case 'week':
					return `Every ${intervalCount} weeks`;
				case 'month':
					return `Every ${intervalCount} months`;
				case 'year':
					return `Every ${intervalCount} years`;
				default:
					return `Every ${intervalCount} ${interval}s`;
			}
		}
	}

	$: allSubscriptions = summary?.subscriptions || [];
	$: subscriptions = statusFilter === 'all' ? allSubscriptions : allSubscriptions.filter((sub: any) => sub.Status === statusFilter);
	$: subscriptionsCount = summary?.subscriptions_count || 0;
	$: activeSubscriptions = allSubscriptions.filter((sub: any) => sub.Status === 'active');
	$: trialingSubscriptions = allSubscriptions.filter((sub: any) => sub.Status === 'trialing');
	$: pastDueSubscriptions = allSubscriptions.filter((sub: any) => sub.Status === 'past_due');
	$: canceledSubscriptions = allSubscriptions.filter((sub: any) => sub.Status === 'canceled' || sub.Status === 'cancelled');

	// Subscription modal functions
	function viewSubscription(subscription: any) {
		selectedSubscription = subscription;
		showSubscriptionModal = true;
	}

	function closeSubscriptionModal() {
		showSubscriptionModal = false;
		selectedSubscription = null;
	}

	function handleModalClick(event: MouseEvent) {
		if (event.target === event.currentTarget) {
			closeSubscriptionModal();
		}
	}

	function cancelSubscription(subscription: any) {
		// TODO: Implement cancel functionality
		alert(`Cancel functionality for subscription ${subscription.ID} will be implemented soon. For now cancel subscriptions directly from Stripe.`);
	}

	function pauseSubscription(subscription: any) {
		// TODO: Implement pause functionality
		alert(`Pause functionality for subscription ${subscription.ID} will be implemented soon. For now pause subscriptions directly from Stripe.`);
	}

	function reactivateSubscription(subscription: any) {
		// TODO: Implement reactivate functionality
		alert(`Reactivate functionality for subscription ${subscription.ID} will be implemented soon. For now reactivate subscriptions directly from Stripe.`);
	}

	function updateSubscription(subscription: any) {
		// TODO: Implement update functionality
		alert(`Update functionality for subscription ${subscription.ID} will be implemented soon. For now update subscriptions directly from Stripe.`);
	}
</script>

{#if loading}
	<div class="loading">
		<div class="spinner"></div>
		<p>Loading subscriptions...</p>
	</div>
{:else if error}
	<div class="error-state">
		<h3>Error Loading Subscriptions</h3>
		<p>{error}</p>
		<button class="btn btn-primary" on:click={fetchSummary}>Retry</button>
	</div>
{:else}
	<div class="subscriptions-container">
		<div class="subscriptions-header">
			<div class="header-content">
				<h1>🔄 Subscriptions</h1>
				<p>Manage recurring subscriptions and billing cycles</p>
			</div>
			<div class="header-stats">
				<div class="stat-card">
					<span class="stat-value">{subscriptionsCount}</span>
					<span class="stat-label">Total Subscriptions</span>
				</div>
				<div class="stat-card active">
					<span class="stat-value">{activeSubscriptions.length}</span>
					<span class="stat-label">Active</span>
				</div>
				<div class="stat-card trialing">
					<span class="stat-value">{trialingSubscriptions.length}</span>
					<span class="stat-label">Trialing</span>
				</div>
				<div class="stat-card past-due">
					<span class="stat-value">{pastDueSubscriptions.length}</span>
					<span class="stat-label">Past Due</span>
				</div>
			</div>
			
			<div class="header-actions">
				<button class="btn btn-secondary" on:click={fetchSummary}>
					🔄 Refresh
				</button>
				<!-- <button class="btn btn-primary">
					➕ Create Subscription
				</button> -->
			</div>
		</div>

		{#if subscriptions.length === 0}
			<div class="empty-state">
				<div class="empty-icon">🔄</div>
				<h3>No Subscriptions Found</h3>
				<p>You don't have any subscriptions yet. Subscriptions will appear here when customers subscribe to your plans.</p>
			</div>
		{:else}
			<div class="subscriptions-table-container">
				<div class="table-header">
					<h2>Recent Subscriptions {statusFilter !== 'all' ? `(${subscriptions.length} ${statusFilter})` : `(${subscriptions.length})`}</h2>
					<div class="table-filters">
						<select class="filter-select" bind:value={statusFilter}>
							<option value="all">All Statuses</option>
							<option value="active">Active</option>
							<option value="trialing">Trialing</option>
							<option value="past_due">Past Due</option>
							<option value="canceled">Canceled</option>
							<option value="unpaid">Unpaid</option>
							<option value="incomplete">Incomplete</option>
							<option value="incomplete_expired">Incomplete Expired</option>
						</select>
					</div>
				</div>

				<div class="table-wrapper">
					<table class="subscriptions-table">
						<thead>
							<tr>
								<th>Subscription</th>
								<th>Status</th>
								<th>Current Period End</th>
								<th>Cancel At Period End</th>
								<th>Created</th>
								<th>Actions</th>
							</tr>
						</thead>
						<tbody>
							{#each subscriptions as subscription}
								<tr class="subscription-row">
									<td class="subscription-id">
										<div class="subscription-info">
											<span class="subscription-number">#{subscription.ID.slice(-8)}</span>
											<span class="subscription-full-id">{subscription.ID}</span>
										</div>
									</td>
									<td class="subscription-status">
										<div class="status-badge" style="color: {getStatusColor(subscription.Status)}">
											<span class="status-icon">{getStatusIcon(subscription.Status)}</span>
											<span class="status-text">{getStatusText(subscription.Status)}</span>
										</div>
									</td>
									<td class="subscription-period-end">
										<span class="period-end-value">{formatDate(subscription.CurrentPeriodEnd)}</span>
									</td>
									<td class="subscription-cancel-at-end">
										<span class="cancel-indicator" class:will-cancel={subscription.CancelAtPeriodEnd}>
											{subscription.CancelAtPeriodEnd ? '⚠️ Yes' : '✅ No'}
										</span>
									</td>
									<td class="subscription-date">
										<span class="date-value">{formatDate(subscription.CreatedAt)}</span>
									</td>
									<td class="subscription-actions">
										<div class="action-buttons">
											<button class="btn btn-sm btn-outline" title="View Subscription" on:click={() => viewSubscription(subscription)}>
												👁️ View
											</button>
											{#if subscription.Status === 'active' && !subscription.CancelAtPeriodEnd}
												<button class="btn btn-sm btn-secondary" title="Cancel Subscription" on:click={() => cancelSubscription(subscription)}>
													❌ Cancel
												</button>
											{/if}
											{#if subscription.Status === 'active'}
												<button class="btn btn-sm btn-outline" title="Update Subscription" on:click={() => updateSubscription(subscription)}>
													✏️ Update
												</button>
											{/if}
											{#if subscription.Status === 'canceled' || subscription.Status === 'cancelled'}
												<button class="btn btn-sm btn-primary" title="Reactivate Subscription" on:click={() => reactivateSubscription(subscription)}>
													🔄 Reactivate
												</button>
											{/if}
										</div>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>

			<!-- Subscription Summary Cards -->
			<div class="subscription-summary">
				<h2>Subscription Summary</h2>
				<div class="summary-grid">
					<div class="summary-card">
						<div class="summary-header">
							<h3>📊 Status Breakdown</h3>
						</div>
						<div class="summary-stats">
							<div class="summary-stat">
								<span class="stat-label">Active Subscriptions</span>
								<span class="stat-value">{activeSubscriptions.length}</span>
							</div>
							<div class="summary-stat">
								<span class="stat-label">Trialing</span>
								<span class="stat-value">{trialingSubscriptions.length}</span>
							</div>
							<div class="summary-stat">
								<span class="stat-label">Past Due</span>
								<span class="stat-value">{pastDueSubscriptions.length}</span>
							</div>
							<div class="summary-stat">
								<span class="stat-label">Canceled</span>
								<span class="stat-value">{canceledSubscriptions.length}</span>
							</div>
						</div>
					</div>

					<div class="summary-card">
						<div class="summary-header">
							<h3>⚠️ Attention Required</h3>
						</div>
						<div class="attention-items">
							{#if pastDueSubscriptions.length > 0}
								<div class="attention-item">
									<div class="attention-info">
										<span class="attention-icon">⚠️</span>
										<span class="attention-text">Past Due Subscriptions</span>
									</div>
									<div class="attention-count">
										<span class="count">{pastDueSubscriptions.length}</span>
									</div>
								</div>
							{/if}
							{#if allSubscriptions.filter((sub: any) => sub.CancelAtPeriodEnd).length > 0}
								<div class="attention-item">
									<div class="attention-info">
										<span class="attention-icon">📅</span>
										<span class="attention-text">Will Cancel at Period End</span>
									</div>
									<div class="attention-count">
										<span class="count">{allSubscriptions.filter((sub: any) => sub.CancelAtPeriodEnd).length}</span>
									</div>
								</div>
							{/if}
							{#if allSubscriptions.filter((sub: any) => sub.Status === 'incomplete').length > 0}
								<div class="attention-item">
									<div class="attention-info">
										<span class="attention-icon">⏳</span>
										<span class="attention-text">Incomplete Subscriptions</span>
									</div>
									<div class="attention-count">
										<span class="count">{allSubscriptions.filter((sub: any) => sub.Status === 'incomplete').length}</span>
									</div>
								</div>
							{/if}
							{#if pastDueSubscriptions.length === 0 && allSubscriptions.filter((sub: any) => sub.CancelAtPeriodEnd).length === 0 && allSubscriptions.filter((sub: any) => sub.Status === 'incomplete').length === 0}
								<div class="no-attention">
									<span class="no-attention-icon">✅</span>
									<span class="no-attention-text">All subscriptions are healthy</span>
								</div>
							{/if}
						</div>
					</div>

					<div class="summary-card">
						<div class="summary-header">
							<h3>📈 Status Distribution</h3>
						</div>
						<div class="status-distribution">
							{#each ['active', 'trialing', 'past_due', 'canceled', 'unpaid', 'incomplete'] as status}
								{@const statusSubscriptions = allSubscriptions.filter((sub: any) => sub.Status === status)}
								{#if statusSubscriptions.length > 0}
									<div class="status-item">
										<div class="status-info">
											<span class="status-icon">{getStatusIcon(status)}</span>
											<span class="status-name">{getStatusText(status)}</span>
										</div>
										<div class="status-count">
											<span class="count">{statusSubscriptions.length}</span>
											<span class="percentage">({Math.round((statusSubscriptions.length / subscriptionsCount) * 100)}%)</span>
										</div>
									</div>
								{/if}
							{/each}
						</div>
					</div>
				</div>
			</div>
		{/if}
	</div>
{/if}

<!-- Subscription View Modal -->
{#if showSubscriptionModal && selectedSubscription}
	<div class="modal-overlay" on:click={handleModalClick} on:keydown={(e) => e.key === 'Escape' && closeSubscriptionModal()} role="dialog" aria-modal="true" tabindex="-1">
		<div class="modal-content" role="document">
			<div class="modal-header">
				<h3>🔄 Subscription Details</h3>
				<button class="modal-close" on:click={closeSubscriptionModal}>&times;</button>
			</div>
			
			<div class="modal-body">
				<div class="subscription-details-grid">
					<div class="detail-section">
						<h4>Subscription Information</h4>
						<div class="detail-rows">
							<div class="detail-row">
								<span class="detail-label">Subscription ID:</span>
								<span class="detail-value">{selectedSubscription.ID}</span>
							</div>
							<div class="detail-row">
								<span class="detail-label">Short ID:</span>
								<span class="detail-value">#{selectedSubscription.ID.slice(-8)}</span>
							</div>
							<div class="detail-row">
								<span class="detail-label">Status:</span>
								<div class="detail-value">
									<div class="status-badge" style="color: {getStatusColor(selectedSubscription.Status)}">
										<span class="status-icon">{getStatusIcon(selectedSubscription.Status)}</span>
										<span class="status-text">{getStatusText(selectedSubscription.Status)}</span>
									</div>
								</div>
							</div>
							<div class="detail-row">
								<span class="detail-label">Created:</span>
								<span class="detail-value">{formatDateTime(selectedSubscription.CreatedAt)}</span>
							</div>
						</div>
					</div>

					<div class="detail-section">
						<h4>Billing & Schedule</h4>
						<div class="detail-rows">
							<div class="detail-row">
								<span class="detail-label">Current Period End:</span>
								<span class="detail-value">{formatDateTime(selectedSubscription.CurrentPeriodEnd)}</span>
							</div>
							<div class="detail-row">
								<span class="detail-label">Cancel at Period End:</span>
								<span class="detail-value">
									<span class="cancel-indicator" class:will-cancel={selectedSubscription.CancelAtPeriodEnd}>
										{selectedSubscription.CancelAtPeriodEnd ? '⚠️ Yes - Will Cancel' : '✅ No - Will Renew'}
									</span>
								</span>
							</div>
						</div>
					</div>

					{#if selectedSubscription.Metadata && Object.keys(selectedSubscription.Metadata).length > 0}
						<div class="detail-section full-width">
							<h4>Metadata</h4>
							<div class="metadata-grid">
								{#each Object.entries(selectedSubscription.Metadata) as [key, value]}
									<div class="metadata-item">
										<span class="metadata-key">{key}:</span>
										<span class="metadata-value">{value}</span>
									</div>
								{/each}
							</div>
						</div>
					{/if}
				</div>
			</div>
			
			<div class="modal-footer">
				<button class="btn btn-secondary" on:click={closeSubscriptionModal}>
					Close
				</button>
				{#if selectedSubscription.Status === 'active' && !selectedSubscription.CancelAtPeriodEnd}
					<button class="btn btn-secondary" on:click={() => cancelSubscription(selectedSubscription)}>
						❌ Cancel
					</button>
				{/if}
				{#if selectedSubscription.Status === 'active'}
					<button class="btn btn-outline" on:click={() => updateSubscription(selectedSubscription)}>
						✏️ Update
					</button>
				{/if}
				{#if selectedSubscription.Status === 'canceled' || selectedSubscription.Status === 'cancelled'}
					<button class="btn btn-primary" on:click={() => reactivateSubscription(selectedSubscription)}>
						🔄 Reactivate
					</button>
				{/if}
			</div>
		</div>
	</div>
{/if}

<style>
	.subscriptions-container {
		padding: var(--space-lg);
	}

	.subscriptions-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: var(--space-xl);
		flex-wrap: wrap;
		gap: var(--space-lg);
	}

	.header-content h1 {
		margin: 0 0 var(--space-xs) 0;
		color: var(--text);
		font-size: 2rem;
		font-weight: 700;
	}

	.header-content p {
		margin: 0;
		color: var(--text-muted);
		font-size: 1.1rem;
	}

	.header-stats {
		display: flex;
		gap: var(--space-md);
		flex-wrap: wrap;
	}

	.stat-card {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: var(--space-md);
		background: var(--surface);
		border-radius: var(--radius-lg);
		border: 1px solid var(--border);
		min-width: 120px;
	}

	.stat-card.active {
		border-color: var(--success);
		background: var(--success-light);
	}

	.stat-card.trialing {
		border-color: var(--primary);
		background: var(--primary-light);
	}

	.stat-card.past-due {
		border-color: var(--warning);
		background: var(--warning-light);
	}

	.stat-value {
		font-size: 1.5rem;
		font-weight: bold;
		color: var(--primary);
		margin-bottom: var(--space-xs);
	}

	.stat-label {
		font-size: 0.875rem;
		color: var(--text-muted);
		text-align: center;
	}

	.header-actions {
		display: flex;
		gap: var(--space-md);
		align-items: center;
	}

	.btn {
		padding: var(--space-sm) var(--space-lg);
		border: none;
		border-radius: var(--radius-md);
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		gap: var(--space-xs);
	}

	.btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-primary {
		background: var(--primary);
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: var(--primary-dark);
		transform: translateY(-1px);
	}

	.btn-secondary {
		background: var(--surface);
		color: var(--text);
		border: 1px solid var(--border);
	}

	.btn-secondary:hover:not(:disabled) {
		background: var(--surface-hover);
	}

	.btn-outline {
		background: transparent;
		color: var(--text);
		border: 1px solid var(--border);
	}

	.btn-outline:hover:not(:disabled) {
		background: var(--surface-hover);
	}

	.btn-sm {
		padding: var(--space-xs) var(--space-sm);
		font-size: 0.75rem;
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--space-xl);
		text-align: center;
		min-height: 400px;
		background: var(--surface);
		border-radius: var(--radius-lg);
		border: 2px dashed var(--border);
	}

	.empty-icon {
		font-size: 4rem;
		margin-bottom: var(--space-lg);
		opacity: 0.5;
	}

	.empty-state h3 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
		font-size: 1.5rem;
	}

	.empty-state p {
		margin: 0;
		color: var(--text-muted);
		font-size: 1.1rem;
		max-width: 500px;
	}

	.subscriptions-table-container {
		background: var(--surface);
		border-radius: var(--radius-lg);
		border: 1px solid var(--border);
		overflow: hidden;
		margin-bottom: var(--space-xl);
	}

	.table-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-lg);
		border-bottom: 1px solid var(--border);
		background: var(--bg-secondary);
	}

	.table-header h2 {
		margin: 0;
		color: var(--text);
		font-size: 1.25rem;
		font-weight: 600;
	}

	.table-filters {
		display: flex;
		gap: var(--space-md);
		align-items: center;
	}

	.filter-select {
		padding: var(--space-sm) var(--space-md);
		border: 1px solid var(--border);
		border-radius: var(--radius-md);
		background: var(--surface);
		color: var(--text);
		font-size: 0.875rem;
	}

	.table-wrapper {
		overflow-x: auto;
	}

	.subscriptions-table {
		width: 100%;
		border-collapse: collapse;
	}

	.subscriptions-table th {
		padding: var(--space-md) var(--space-lg);
		text-align: left;
		font-weight: 600;
		color: var(--text-muted);
		font-size: 0.875rem;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		border-bottom: 1px solid var(--border);
		background: var(--bg-secondary);
	}

	.subscriptions-table td {
		padding: var(--space-md) var(--space-lg);
		border-bottom: 1px solid var(--border);
	}

	.subscription-row:hover {
		background: var(--bg-secondary);
	}

	.subscription-info {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.subscription-number {
		font-weight: 600;
		color: var(--text);
		font-size: 1rem;
	}

	.subscription-full-id {
		font-family: monospace;
		font-size: 0.75rem;
		color: var(--text-muted);
	}

	.status-badge {
		display: flex;
		align-items: center;
		gap: var(--space-xs);
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-md);
		background: var(--bg-secondary);
		font-weight: 600;
		font-size: 0.875rem;
	}

	.period-end-value {
		font-weight: 600;
		color: var(--text);
		font-size: 0.875rem;
	}

	.cancel-indicator {
		font-weight: 600;
		font-size: 0.875rem;
	}

	.cancel-indicator.will-cancel {
		color: var(--warning);
	}

	.date-value {
		color: var(--text);
		font-size: 0.875rem;
	}

	.action-buttons {
		display: flex;
		gap: var(--space-xs);
		flex-wrap: wrap;
	}

	.subscription-summary {
		margin-top: var(--space-xl);
	}

	.subscription-summary h2 {
		margin: 0 0 var(--space-lg) 0;
		color: var(--text);
		font-size: 1.5rem;
	}

	.summary-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: var(--space-lg);
	}

	.summary-card {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
	}

	.summary-header h3 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
		font-size: 1.1rem;
		font-weight: 600;
	}

	.summary-stats {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.summary-stat {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-sm) 0;
		border-bottom: 1px solid var(--border);
	}

	.summary-stat:last-child {
		border-bottom: none;
	}

	.summary-stat .stat-label {
		color: var(--text-muted);
		font-size: 0.875rem;
	}

	.summary-stat .stat-value {
		color: var(--text);
		font-weight: 600;
		font-size: 1rem;
	}

	.attention-items {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}

	.attention-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-sm);
		border-radius: var(--radius-md);
		background: var(--warning-light);
		border: 1px solid var(--warning);
	}

	.attention-info {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
	}

	.attention-text {
		font-weight: 500;
		color: var(--text);
	}

	.attention-count {
		display: flex;
		align-items: center;
		gap: var(--space-xs);
	}

	.count {
		font-weight: 600;
		color: var(--text);
	}

	.no-attention {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		padding: var(--space-sm);
		border-radius: var(--radius-md);
		background: var(--success-light);
		border: 1px solid var(--success);
	}

	.no-attention-text {
		font-weight: 500;
		color: var(--text);
	}

	.status-distribution {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}

	.status-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-sm);
		border-radius: var(--radius-md);
		background: var(--bg-secondary);
	}

	.status-info {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
	}

	.status-name {
		font-weight: 500;
		color: var(--text);
	}

	.status-count {
		display: flex;
		align-items: center;
		gap: var(--space-xs);
	}

	.percentage {
		font-size: 0.8rem;
		color: var(--text-muted);
	}

	.loading {
		text-align: center;
		padding: var(--space-xl);
	}

	.spinner {
		width: 40px;
		height: 40px;
		border: 4px solid var(--border);
		border-top: 4px solid var(--primary);
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin: 0 auto var(--space-md);
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.error-state {
		text-align: center;
		padding: var(--space-xl);
	}

	.error-state h3 {
		color: var(--error);
		margin-bottom: var(--space-md);
	}

	/* Modal Styles */
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.6);
		display: flex;
		justify-content: center;
		align-items: center;
		z-index: 1000;
		backdrop-filter: blur(4px);
	}

	.modal-content {
		background: var(--surface);
		border-radius: var(--radius-lg);
		box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
		width: 90%;
		max-width: 800px;
		max-height: 90vh;
		display: flex;
		flex-direction: column;
		border: 1px solid var(--border);
		overflow: hidden;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-lg);
		border-bottom: 1px solid var(--border);
		background: var(--bg-secondary);
	}

	.modal-header h3 {
		margin: 0;
		color: var(--text);
		font-size: 1.25rem;
		font-weight: 600;
	}

	.modal-close {
		background: none;
		border: none;
		font-size: 1.5rem;
		cursor: pointer;
		color: var(--text-muted);
		padding: 0;
		width: 32px;
		height: 32px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: var(--radius-md);
		transition: all 0.2s ease;
	}

	.modal-close:hover {
		background: var(--surface-hover);
		color: var(--text);
	}

	.modal-body {
		padding: var(--space-lg);
		overflow-y: auto;
		flex-grow: 1;
	}

	.subscription-details-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-lg);
	}

	.detail-section {
		background: var(--bg-secondary);
		padding: var(--space-lg);
		border-radius: var(--radius-md);
		border: 1px solid var(--border);
	}

	.detail-section.full-width {
		grid-column: 1 / -1;
	}

	.detail-section h4 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
		font-size: 1rem;
		font-weight: 600;
		border-bottom: 1px solid var(--border);
		padding-bottom: var(--space-sm);
	}

	.detail-rows {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}

	.detail-section .detail-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-sm) 0;
	}

	.detail-section .detail-label {
		color: var(--text-muted);
		font-size: 0.875rem;
		font-weight: 500;
	}

	.detail-section .detail-value {
		color: var(--text);
		font-size: 0.875rem;
		font-weight: 600;
		text-align: right;
		word-break: break-all;
	}

	.metadata-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: var(--space-sm);
	}

	.metadata-item {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
		padding: var(--space-sm);
		background: var(--surface);
		border-radius: var(--radius-sm);
		border: 1px solid var(--border);
	}

	.metadata-key {
		color: var(--text-muted);
		font-size: 0.8rem;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.metadata-value {
		color: var(--text);
		font-size: 0.875rem;
		font-family: monospace;
		word-break: break-all;
	}

	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-md);
		padding: var(--space-lg);
		border-top: 1px solid var(--border);
		background: var(--bg-secondary);
	}

	@media (max-width: 768px) {
		.subscriptions-header {
			flex-direction: column;
			align-items: center;
			text-align: center;
		}

		.header-stats {
			justify-content: center;
		}

		.header-actions {
			justify-content: center;
		}

		.table-header {
			flex-direction: column;
			gap: var(--space-md);
			align-items: flex-start;
		}

		.subscriptions-table {
			font-size: 0.875rem;
		}

		.subscriptions-table th,
		.subscriptions-table td {
			padding: var(--space-sm);
		}

		.action-buttons {
			flex-direction: column;
		}

		.summary-grid {
			grid-template-columns: 1fr;
		}

		.modal-content {
			width: 95%;
			max-height: 95vh;
		}

		.subscription-details-grid {
			grid-template-columns: 1fr;
		}

		.modal-footer {
			flex-direction: column-reverse;
		}
	}
</style>
