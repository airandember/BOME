<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';
	import { toastStore } from '$lib/stores/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	// State variables
	let webhookStatus = $state<any>(null);
	let webhookLogs = $state<any[]>([]);
	let loading = $state(true);
	let logsLoading = $state(false);
	let error = $state('');
	let pinging = $state(false);
	let refreshing = $state(false);
	let showLogs = $state(true);
	let retryingEvents = $state(new Set<number>()); // Track which events are being retried
	let expandedRows = $state(new Set<number>()); // Track which rows are expanded
	
	// Pagination and filtering
	let currentPage = $state(1);
	let totalPages = $state(1);
	let totalEvents = $state(0);
	let hasMore = $state(false);
	let eventTypeFilter = $state('');
	let statusFilter = $state('');

	onMount(async () => {
		await loadWebhookStatus();
		// Auto-load webhook logs on mount
		if (showLogs) {
			await loadWebhookLogs(1);
		}
	});

	async function loadWebhookStatus() {
		try {
			refreshing = true;
			error = '';
			
			const response = await apiRequest('/admin/streaming/stripe/webhooks/status');
			
			if (response.ok) {
				const data = await response.json();
				console.log('🔍🔍🔍🔍🔍 Webhook status data:', data);
				webhookStatus = data.webhook || null;
				console.log('✅ Webhook status loaded:', webhookStatus);
			} else {
				throw new Error(`Failed to load webhook status: ${response.status}`);
			}
		} catch (err: any) {
			console.error('❌ Failed to load webhook status:', err);
			error = err.message || 'Failed to load webhook status';
		} finally {
			loading = false;
			refreshing = false;
		}
	}

	async function loadWebhookLogs(page = 1) {
		try {
			logsLoading = true;
			
			// Build query parameters
			const params = new URLSearchParams({
				page: page.toString(),
				limit: '20'
			});
			
			if (eventTypeFilter) {
				params.append('event_type', eventTypeFilter);
			}
			
			if (statusFilter) {
				params.append('status', statusFilter);
			}
			
			const response = await apiRequest(`/admin/streaming/stripe/webhooks/logs?${params}`);
			
			if (response.ok) {
				const data = await response.json();
				webhookLogs = data.events || [];
				currentPage = data.page || 1;
				totalPages = data.total_pages || 1;
				totalEvents = data.total || 0;
				hasMore = data.has_more || false;
				
				console.log('✅ Webhook logs loaded:', {
					events: webhookLogs.length,
					page: currentPage,
					total: totalEvents
				});
			} else {
				throw new Error(`Failed to load webhook logs: ${response.status}`);
			}
		} catch (err: any) {
			console.error('❌ Failed to load webhook logs:', err);
			error = err.message || 'Failed to load webhook logs';
			toastStore.error('Failed to load webhook logs', { duration: 3000 });
		} finally {
			logsLoading = false;
		}
	}

	async function toggleLogs() {
		showLogs = !showLogs;
		if (showLogs && webhookLogs.length === 0) {
			await loadWebhookLogs();
		}
	}

	async function applyFilters() {
		currentPage = 1;
		await loadWebhookLogs(1);
	}

	async function clearFilters() {
		eventTypeFilter = '';
		statusFilter = '';
		currentPage = 1;
		await loadWebhookLogs(1);
	}

	function formatTimestamp(timestamp: string) {
		return new Date(timestamp).toLocaleString();
	}

	function getStatusClass(status: string) {
		switch (status.toLowerCase()) {
			case 'success':
				return 'status-success';
			case 'failed':
			case 'error':
				return 'status-error';
			default:
				return 'status-pending';
		}
	}

	function formatDuration(ms: number) {
		if (ms < 1000) {
			return `${ms}ms`;
		}
		return `${(ms / 1000).toFixed(2)}s`;
	}

	function formatSize(bytes: number) {
		if (bytes < 1024) {
			return `${bytes}B`;
		} else if (bytes < 1024 * 1024) {
			return `${(bytes / 1024).toFixed(1)}KB`;
		}
		return `${(bytes / (1024 * 1024)).toFixed(1)}MB`;
	}

	function getHealthStatusClass(healthStatus: string): string {
		switch (healthStatus) {
			case 'healthy':
				return 'status-healthy';
			case 'degraded':
				return 'status-degraded';
			case 'unhealthy':
				return 'status-unhealthy';
			case 'monitoring':
				return 'status-monitoring';
			case 'inactive':
				return 'status-inactive';
			case 'no_activity':
				return 'status-no-activity';
			default:
				// Fallback to inactive for professional appearance
				return 'status-inactive';
		}
	}

	function getHealthStatusLabel(healthStatus: string): string {
		switch (healthStatus) {
			case 'healthy':
				return '🟢 Healthy';
			case 'degraded':
				return '🟡 Degraded';
			case 'unhealthy':
				return '🔴 Unhealthy';
			case 'monitoring':
				return '🔵 Monitoring';
			case 'inactive':
				return '⚫ Inactive';
			case 'no_activity':
				return '🟣 No Recent Activity';
			default:
				// Fallback to inactive for professional appearance
				return '⚫ Inactive';
		}
	}

	async function pingWebhook() {
		try {
			pinging = true;
			error = '';
			
			const response = await apiRequest('/admin/streaming/stripe/webhooks/ping', {
				method: 'POST'
			});
			
			if (response.ok) {
				const data = await response.json();
				console.log('🏓 Ping response:', data);
				
				toastStore.success('Webhook ping sent successfully! Check your webhook logs.', {
					duration: 5000
				});
				
				// Refresh status after ping and reload logs if visible
				setTimeout(() => {
					loadWebhookStatus();
					if (showLogs) {
						loadWebhookLogs(currentPage);
					}
				}, 1000);
			} else {
				const errorData = await response.json().catch(() => ({}));
				throw new Error(errorData.error || `Ping failed: ${response.status}`);
			}
		} catch (err: any) {
			console.error('❌ Webhook ping failed:', err);
			error = err.message || 'Failed to ping webhook';
			toastStore.error(`Webhook ping failed: ${err.message}`, {
				duration: 5000
			});
		} finally {
			pinging = false;
		}
	}

	async function retryWebhookEvent(eventId: number) {
		try {
			retryingEvents.add(eventId);
			
			const response = await apiRequest(`/admin/streaming/stripe/webhooks/retry/${eventId}`, {
				method: 'POST'
			});
			
			if (response.ok) {
				const data = await response.json();
				console.log('🔄 Retry response:', data);
				
				toastStore.success(`Webhook event retried successfully! Attempt ${data.retry_count}`, {
					duration: 4000
				});
				
				// Refresh logs to show updated status
				setTimeout(() => {
					loadWebhookLogs(currentPage);
				}, 500);
			} else {
				const errorData = await response.json().catch(() => ({}));
				throw new Error(errorData.error || `Retry failed: ${response.status}`);
			}
		} catch (err: any) {
			console.error('❌ Webhook retry failed:', err);
			toastStore.error(`Webhook retry failed: ${err.message}`, {
				duration: 5000
			});
		} finally {
			retryingEvents.delete(eventId);
		}
	}

	function toggleRowExpansion(eventId: number) {
		if (expandedRows.has(eventId)) {
			expandedRows.delete(eventId);
		} else {
			expandedRows.add(eventId);
		}
		expandedRows = new Set(expandedRows); // Trigger reactivity
	}
</script>

<div class="webhook-dashboard">
	<!-- Header -->
	<div class="dashboard-header">
		<div class="header-content">
			<h1>🔗 Webhook Configuration</h1>
			<p>Monitor and manage your Stripe webhook endpoints</p>
		</div>
		
		<div class="header-actions">
			<button 
				class="btn btn-secondary"
				onclick={loadWebhookStatus}
				disabled={refreshing}
				title="Refresh webhook status"
			>
				{#if refreshing}
					<span class="spinner-small"></span>
				{:else}
					🔄
				{/if}
				Refresh
			</button>
		</div>
	</div>

	{#if loading}
		<div class="loading-container">
			<div class="spinner"></div>
			<p>Loading webhook configuration...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<div class="error-icon">⚠️</div>
			<h3>Error Loading Webhook Status</h3>
			<p class="error-message">{error}</p>
			<button class="btn btn-primary" onclick={loadWebhookStatus}>
				🔄 Retry
			</button>
		</div>
	{:else}
		<div class="webhook-content">
			<!-- Webhook Status Card -->
			<div class="status-card">
				<div class="card-header">
					<h2>📊 Webhook Status</h2>
					<div class="status-indicator {getHealthStatusClass(webhookStatus?.health_status)}">
						<div class="status-dot"></div>
						<span>{getHealthStatusLabel(webhookStatus?.health_status)}</span>
					</div>
				</div>

				{#if webhookStatus}
					<div class="status-details">
						<div class="detail-grid">
							<div class="detail-item">
								<span class="detail-label">Endpoint URL:</span>
								<span class="detail-value">
									{webhookStatus.endpoint || 'Not configured'}
								</span>
							</div>
							
							<div class="detail-item">
								<span class="detail-label">Last Event:</span>
								<span class="detail-value">
									{webhookStatus.lastEvent || 'Never'}
								</span>
							</div>
							
							<div class="detail-item">
								<span class="detail-label">Events Today:</span>
								<span class="detail-value">
									{webhookStatus.eventsToday || 0} events
								</span>
							</div>
							
							<div class="detail-item">
								<span class="detail-label">Total Events:</span>
								<span class="detail-value">
									{webhookStatus.totalEvents || 0} events
								</span>
							</div>
							
							<div class="detail-item">
								<span class="detail-label">Success Rate:</span>
								<span class="detail-value">
									{webhookStatus.successRate !== undefined 
										? `${webhookStatus.successRate}%`
										: '100%'
									}
								</span>
							</div>

							{#if webhookStatus.recentFailureRate > 0}
							<div class="detail-item">
								<span class="detail-label">Recent Failures:</span>
								<span class="detail-value error-rate">
									{webhookStatus.recentFailureRate}%
								</span>
							</div>
							{/if}
						</div>
					</div>
				{:else}
					<div class="no-webhook">
						<div class="no-webhook-icon">🔌</div>
						<h3>No Webhook Configuration Found</h3>
						<p>Set up your Stripe webhook to start receiving real-time events.</p>
					</div>
				{/if}
			</div>

			<!-- Actions Card -->
			<div class="actions-card">
				<div class="card-header">
					<h2>🛠️ Webhook Actions</h2>
				</div>

				<div class="actions-grid">
					<div class="action-item">
						<div class="action-icon">🏓</div>
						<div class="action-content">
							<h3>Test Webhook</h3>
							<p>Send a ping event to test your webhook endpoint</p>
							<button 
								class="btn btn-primary"
								onclick={pingWebhook}
								disabled={pinging}
							>
								{#if pinging}
									<span class="spinner-small"></span>
									Sending Ping...
								{:else}
									🏓 Send Ping
								{/if}
							</button>
						</div>
					</div>

					<div class="action-item">
						<div class="action-icon">📋</div>
						<div class="action-content">
							<h3>Webhook Logs</h3>
							<p>View detailed logs of webhook events and responses</p>
							<button class="btn btn-secondary" onclick={toggleLogs}>
								{#if showLogs}
									📋 Hide Logs
								{:else}
									📋 View Logs
								{/if}
							</button>
						</div>
					</div>

					<!--<div class="action-item">
						<div class="action-icon">⚙️</div>
						<div class="action-content">
							<h3>Configuration</h3>
							<p>Manage webhook endpoint settings and event types</p>
							<button class="btn btn-secondary" disabled>
								⚙️ Configure
								<span class="coming-soon">(Coming Soon)</span>
							</button>
						</div>
					</div>-->
				</div>
			</div>

			<!-- Webhook Logs Section -->
			{#if showLogs}
				<div class="logs-card">
					<div class="card-header">
						<h2>📋 Webhook Event Logs</h2>
						<div class="logs-summary">
							<span class="logs-count">{totalEvents} total events</span>
							{#if totalPages > 1}
								<span class="logs-pagination">Page {currentPage} of {totalPages}</span>
							{/if}
						</div>
					</div>

					<!-- Filters -->
					<div class="logs-filters">
						<div class="filter-group">
							<label for="event-type-filter">Event Type:</label>
							<select id="event-type-filter" bind:value={eventTypeFilter} onchange={applyFilters}>
								<option value="">All Events</option>
								<option value="v2.core.event_destination.ping">Ping Events</option>
								<option value="customer.created">Customer Created</option>
								<option value="customer.updated">Customer Updated</option>
								<option value="customer.subscription.created">Subscription Created</option>
								<option value="customer.subscription.updated">Subscription Updated</option>
								<option value="invoice.payment_succeeded">Payment Succeeded</option>
								<option value="invoice.payment_failed">Payment Failed</option>
							</select>
						</div>

						<div class="filter-group">
							<label for="status-filter">Status:</label>
							<select id="status-filter" bind:value={statusFilter} onchange={applyFilters}>
								<option value="">All Statuses</option>
								<option value="success">Success</option>
								<option value="failed">Failed</option>
							</select>
						</div>

						<div class="filter-actions">
							<button class="btn btn-secondary" onclick={clearFilters} disabled={logsLoading}>
								Clear Filters
							</button>
							<button class="btn btn-primary" onclick={() => loadWebhookLogs(1)} disabled={logsLoading}>
								{#if logsLoading}
									<span class="spinner-small"></span>
								{:else}
									🔄
								{/if}
								Refresh
							</button>
						</div>
					</div>

					<!-- Logs Table -->
					{#if logsLoading}
						<div class="logs-loading">
							<div class="spinner"></div>
							<p>Loading webhook logs...</p>
						</div>
					{:else if webhookLogs.length === 0}
						<div class="no-logs">
							<div class="no-logs-icon">📝</div>
							<h3>No Webhook Logs Found</h3>
							<p>No webhook events match your current filters. Try adjusting the filters or send a test ping.</p>
						</div>
					{:else}
						<div class="logs-table-container">
							<table class="logs-table">
								<thead>
									<tr>
										<th>Timestamp</th>
										<th>User</th>
										<th>Event Type</th>
										<th>Status</th>
										<th>Actions</th>
									</tr>
								</thead>
								<tbody>
									{#each webhookLogs as log (log.id)}
										<!-- Main row -->
										<tr class="log-row" class:expanded={expandedRows.has(log.id)}>
											<td class="timestamp">
												{formatTimestamp(log.created_at)}
											</td>
											<td class="user-cell">
												{#if log.user_email}
													<div class="user-info">
														<span class="user-email">{log.user_email}</span>
														{#if log.user_id}
															<span class="user-id">ID: {log.user_id}</span>
														{/if}
													</div>
												{:else}
													<span class="no-user">System Event</span>
												{/if}
											</td>
											<td class="event-type">
												<code>{log.event_type}</code>
												{#if log.description}
													<div class="event-description">{log.description}</div>
												{/if}
											</td>
											<td class="status">
												<span class="status-badge {getStatusClass(log.status)}">
													{log.status}
												</span>
											</td>
											<td class="actions">
												<button 
													class="btn-details" 
													onclick={() => toggleRowExpansion(log.id)}
													title="View event details"
												>
													{#if expandedRows.has(log.id)}
														▼ Hide
													{:else}
														▶ Details
													{/if}
												</button>
												{#if log.status === 'failed' && log.retry_count < 5}
													<button 
														class="btn-retry" 
														onclick={() => retryWebhookEvent(log.id)}
														disabled={retryingEvents.has(log.id)}
														title="Retry failed webhook event"
													>
														{#if retryingEvents.has(log.id)}
															<LoadingSpinner size="small" />
														{:else}
															🔄
														{/if}
													</button>
												{/if}
											</td>
										</tr>
										
										<!-- Expanded details row -->
										{#if expandedRows.has(log.id)}
											<tr class="details-row">
												<td colspan="5">
													<div class="details-content">
														<div class="details-grid">
															<!-- Technical Details -->
															<div class="detail-section">
																<h4>📊 Technical Details</h4>
																<div class="detail-items">
																	<div class="detail-row">
																		<span class="label">Response Time:</span>
																		<span class="value">{formatDuration(log.response_time)}</span>
																	</div>
																	<div class="detail-row">
																		<span class="label">Payload Size:</span>
																		<span class="value">{formatSize(log.payload_size)}</span>
																	</div>
																	<div class="detail-row">
																		<span class="label">Status Code:</span>
																		<span class="value">
																			<span class="status-code-badge status-code-{Math.floor(log.status_code / 100)}xx">
																				{log.status_code}
																			</span>
																		</span>
																	</div>
																	{#if log.retry_count > 0}
																		<div class="detail-row">
																			<span class="label">Retry Count:</span>
																			<span class="value">{log.retry_count}</span>
																		</div>
																	{/if}
																</div>
															</div>

															<!-- Stripe Details (if available) -->
															{#if log.stripe_event_id || log.stripe_object_id}
																<div class="detail-section">
																	<h4>🔗 Stripe Information</h4>
																	<div class="detail-items">
																		{#if log.stripe_event_id}
																			<div class="detail-row">
																				<span class="label">Event ID:</span>
																				<span class="value"><code>{log.stripe_event_id}</code></span>
																			</div>
																		{/if}
																		{#if log.stripe_object_id}
																			<div class="detail-row">
																				<span class="label">Object ID:</span>
																				<span class="value"><code>{log.stripe_object_id}</code></span>
																			</div>
																		{/if}
																		{#if log.stripe_object_type}
																			<div class="detail-row">
																				<span class="label">Object Type:</span>
																				<span class="value">{log.stripe_object_type}</span>
																			</div>
																		{/if}
																		{#if log.api_version}
																			<div class="detail-row">
																				<span class="label">API Version:</span>
																				<span class="value">{log.api_version}</span>
																			</div>
																		{/if}
																	</div>
																</div>
															{/if}

															<!-- Subscription/Payment Details -->
															{#if log.subscription_id || log.invoice_id || log.amount_cents}
																<div class="detail-section">
																	<h4>💳 Subscription & Payment</h4>
																	<div class="detail-items">
																		{#if log.subscription_id}
																			<div class="detail-row">
																				<span class="label">Subscription:</span>
																				<span class="value"><code>{log.subscription_id}</code></span>
																			</div>
																		{/if}
																		{#if log.subscription_status}
																			<div class="detail-row">
																				<span class="label">Status:</span>
																				<span class="value">
																					<span class="status-badge status-{log.subscription_status}">
																						{log.subscription_status}
																					</span>
																				</span>
																			</div>
																		{/if}
																		{#if log.invoice_id}
																			<div class="detail-row">
																				<span class="label">Invoice:</span>
																				<span class="value"><code>{log.invoice_id}</code></span>
																			</div>
																		{/if}
																		{#if log.amount_cents}
																			<div class="detail-row">
																				<span class="label">Amount:</span>
																				<span class="value">
																					${(log.amount_cents / 100).toFixed(2)} {log.currency?.toUpperCase() || 'USD'}
																				</span>
																			</div>
																		{/if}
																	</div>
																</div>
															{/if}
														</div>

														<!-- Error Message (if any) -->
														{#if log.error_message}
															<div class="error-section">
																<h4>⚠️ Error Details</h4>
																<pre class="error-pre">{log.error_message}</pre>
															</div>
														{/if}

														<!-- Full Event Data (if available) -->
														{#if log.event_data}
															<div class="json-section">
																<h4>📄 Full Event Data</h4>
																<pre class="json-pre">{JSON.stringify(JSON.parse(log.event_data), null, 2)}</pre>
															</div>
														{/if}
													</div>
												</td>
											</tr>
										{/if}
									{/each}
								</tbody>
							</table>
						</div>

						<!-- Pagination -->
						{#if totalPages > 1}
							<div class="logs-pagination">
								<button 
									class="btn btn-secondary" 
									onclick={() => loadWebhookLogs(currentPage - 1)}
									disabled={currentPage <= 1 || logsLoading}
								>
									← Previous
								</button>
								
								<span class="pagination-info">
									Page {currentPage} of {totalPages}
								</span>
								
								<button 
									class="btn btn-secondary" 
									onclick={() => loadWebhookLogs(currentPage + 1)}
									disabled={currentPage >= totalPages || logsLoading}
								>
									Next →
								</button>
							</div>
						{/if}
					{/if}
				</div>
			{/if}

			<!-- Setup Instructions -->
			<div class="instructions-card">
				<div class="card-header">
					<h2>📚 Webhook Setup Guide</h2>
				</div>

				<div class="instructions-content">
					<div class="step">
						<div class="step-number">1</div>
						<div class="step-content">
							<h4>Create Webhook Endpoint</h4>
							<p>In your Stripe Dashboard, go to <strong>Developers → Webhooks</strong> and create a new endpoint.</p>
						</div>
					</div>

					<div class="step">
						<div class="step-number">2</div>
						<div class="step-content">
							<h4>Set Endpoint URL</h4>
							<p>Use your backend URL with the webhook path:</p>
							<code class="endpoint-url">https://your-domain.com/api/v1/stripe/webhooks</code>
						</div>
					</div>

					<div class="step">
						<div class="step-number">3</div>
						<div class="step-content">
							<h4>Configure Events</h4>
							<p>Select the events you want to receive. Common events include:</p>
							<ul class="event-list">
								<li><code>customer.created</code></li>
								<li><code>customer.updated</code></li>
								<li><code>invoice.payment_succeeded</code></li>
								<li><code>invoice.payment_failed</code></li>
								<li><code>customer.subscription.created</code></li>
								<li><code>customer.subscription.updated</code></li>
								<li><code>customer.subscription.deleted</code></li>
							</ul>
						</div>
					</div>

					<div class="step">
						<div class="step-number">4</div>
						<div class="step-content">
							<h4>Test Your Webhook</h4>
							<p>Use the "Send Ping" button above to test your webhook endpoint is working correctly.</p>
						</div>
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	.webhook-dashboard {
		padding: var(--space-lg);
		background: var(--bg-primary);
		min-height: 100vh;
	}

	.dashboard-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-xl);
		padding-bottom: var(--space-lg);
		border-bottom: 1px solid var(--border);
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

	.header-actions {
		display: flex;
		gap: var(--space-md);
	}

	.loading-container,
	.error-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--space-xl);
		text-align: center;
		min-height: 50vh;
	}

	.spinner {
		width: 40px;
		height: 40px;
		border: 4px solid var(--border);
		border-top: 4px solid var(--primary);
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-bottom: var(--space-md);
	}

	.spinner-small {
		display: inline-block;
		width: 16px;
		height: 16px;
		border: 2px solid rgba(255, 255, 255, 0.3);
		border-top: 2px solid currentColor;
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-right: var(--space-sm);
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.error-container {
		color: var(--error);
	}

	.error-icon {
		font-size: 3rem;
		margin-bottom: var(--space-md);
	}

	.error-container h3 {
		margin: 0 0 var(--space-md) 0;
		font-size: 1.5rem;
	}

	.error-message {
		margin-bottom: var(--space-lg);
		padding: var(--space-md);
		background: var(--error-light);
		border-radius: var(--radius-md);
		border-left: 4px solid var(--error);
	}

	.webhook-content {
		display: grid;
		gap: var(--space-xl);
		max-width: 1200px;
		margin: 0 auto;
	}

	.status-card,
	.actions-card,
	.instructions-card,
	.logs-card {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-sm);
		overflow: hidden;
	}

	.card-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-lg);
		background: var(--bg-secondary);
		border-bottom: 1px solid var(--border);
	}

	.card-header h2 {
		margin: 0;
		color: var(--text);
		font-size: 1.5rem;
		font-weight: 600;
	}

	.status-indicator {
		display: flex;
		align-items: center;
		gap: var(--space-xs);
		padding: var(--space-xs) var(--space-md);
		border-radius: var(--radius-md);
		font-size: 0.9rem;
		font-weight: 600;
	}

	.status-indicator.active {
		background: var(--success-light);
		color: var(--success-dark);
	}

	.status-indicator.inactive {
		background: var(--error-light);
		color: var(--error-dark);
	}

	.status-healthy {
		background: var(--success-light);
		color: var(--success-dark);
	}

	.status-degraded {
		background: var(--warning-light);
		color: var(--warning-dark);
	}

	.status-unhealthy {
		background: var(--error-light);
		color: var(--error-dark);
	}

	.status-monitoring {
		background: var(--primary-light);
		color: var(--primary-dark);
	}

	.status-inactive {
		background: #6b7280;
		color: #f9fafb;
	}

	.status-no-activity {
		background: #f3e8ff;
		color: #7c3aed;
		border: 1px solid #c4b5fd;
	}

	.status-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: currentColor;
	}

	.status-details {
		padding: var(--space-lg);
	}

	.detail-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: var(--space-lg);
	}

	.detail-item {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.detail-label {
		font-size: 0.9rem;
		color: var(--text-muted);
		font-weight: 500;
	}

	.detail-value {
		font-size: 1rem;
		color: var(--text);
		font-weight: 600;
		word-break: break-all;
	}

	.error-rate {
		color: var(--error) !important;
		font-weight: 700;
	}

	.no-webhook {
		padding: var(--space-xl);
		text-align: center;
	}

	.no-webhook-icon {
		font-size: 3rem;
		margin-bottom: var(--space-lg);
		opacity: 0.7;
	}

	.no-webhook h3 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
		font-size: 1.5rem;
	}

	.no-webhook p {
		margin: 0;
		color: var(--text-muted);
		font-size: 1.1rem;
	}

	.actions-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: var(--space-lg);
		padding: var(--space-lg);
	}

	.action-item {
		display: flex;
		gap: var(--space-md);
		padding: var(--space-lg);
		background: var(--bg-secondary);
		border: 1px solid var(--border);
		border-radius: var(--radius-md);
	}

	.action-icon {
		font-size: 2rem;
		color: var(--primary);
		flex-shrink: 0;
	}

	.action-content {
		flex: 1;
	}

	.action-content h3 {
		margin: 0 0 var(--space-xs) 0;
		color: var(--text);
		font-size: 1.2rem;
		font-weight: 600;
	}

	.action-content p {
		margin: 0 0 var(--space-md) 0;
		color: var(--text-muted);
		font-size: 0.9rem;
		line-height: 1.4;
	}

	.coming-soon {
		font-size: 0.8rem;
		color: var(--text-muted);
		font-style: italic;
		margin-left: var(--space-xs);
	}

	.instructions-content {
		padding: var(--space-lg);
	}

	.step {
		display: flex;
		gap: var(--space-md);
		margin-bottom: var(--space-lg);
		padding-bottom: var(--space-lg);
		border-bottom: 1px solid var(--border);
	}

	.step:last-child {
		margin-bottom: 0;
		padding-bottom: 0;
		border-bottom: none;
	}

	.step-number {
		flex-shrink: 0;
		width: 32px;
		height: 32px;
		background: var(--primary);
		color: white;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 600;
		font-size: 1.1rem;
	}

	.step-content h4 {
		margin: 0 0 var(--space-xs) 0;
		color: var(--text);
		font-size: 1.2rem;
		font-weight: 600;
	}

	.step-content p {
		margin: 0 0 var(--space-sm) 0;
		color: var(--text-muted);
		line-height: 1.5;
	}

	.endpoint-url {
		display: inline-block;
		padding: var(--space-xs) var(--space-sm);
		background: var(--bg-secondary);
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		font-family: monospace;
		font-size: 0.9rem;
		color: var(--text);
		margin: var(--space-xs) 0;
	}

	.event-list {
		margin: var(--space-sm) 0 0 0;
		padding-left: var(--space-lg);
	}

	.event-list li {
		margin-bottom: var(--space-xs);
		color: var(--text-muted);
	}

	.event-list code {
		background: var(--bg-secondary);
		padding: 2px var(--space-xs);
		border-radius: var(--radius-sm);
		font-size: 0.85rem;
		color: var(--text);
	}

	.btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: var(--space-sm) var(--space-lg);
		border: none;
		border-radius: var(--radius-md);
		font-size: 0.95rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		gap: var(--space-xs);
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
		background: var(--bg-secondary);
		color: var(--text);
		border: 1px solid var(--border);
	}

	.btn-secondary:hover:not(:disabled) {
		background: var(--surface-hover);
	}

	.btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
		transform: none !important;
	}

	@media (max-width: 768px) {
		.webhook-dashboard {
			padding: var(--space-md);
		}

		.dashboard-header {
			flex-direction: column;
			gap: var(--space-md);
			text-align: center;
		}

		.header-content h1 {
			font-size: 1.8rem;
		}

		.detail-grid {
			grid-template-columns: 1fr;
		}

		.actions-grid {
			grid-template-columns: 1fr;
		}

		.action-item {
			flex-direction: column;
			text-align: center;
		}

		.step {
			flex-direction: column;
			text-align: center;
		}

		.step-number {
			align-self: center;
		}
	}

	/* Webhook Logs Styles */
	.logs-card {
		margin-top: var(--space-xl);
	}

	.logs-summary {
		display: flex;
		gap: var(--space-md);
		align-items: center;
		font-size: 0.9rem;
		color: var(--text-muted);
	}

	.logs-count {
		font-weight: 600;
	}

	.logs-filters {
		display: flex;
		gap: var(--space-lg);
		align-items: flex-end;
		padding: var(--space-lg);
		border-bottom: 1px solid var(--border);
		flex-wrap: wrap;
	}

	.filter-group {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
		min-width: 150px;
	}

	.filter-group label {
		font-size: 0.9rem;
		font-weight: 500;
		color: var(--text);
	}

	.filter-group select {
		padding: var(--space-sm);
		border: 1px solid var(--border);
		border-radius: var(--radius-md);
		background: var(--bg-secondary);
		color: var(--text);
		font-size: 0.9rem;
	}

	.filter-actions {
		display: flex;
		gap: var(--space-sm);
		margin-left: auto;
	}

	.logs-loading {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--space-xl);
		text-align: center;
	}

	.no-logs {
		padding: var(--space-xl);
		text-align: center;
	}

	.no-logs-icon {
		font-size: 3rem;
		margin-bottom: var(--space-lg);
		opacity: 0.7;
	}

	.no-logs h3 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
		font-size: 1.5rem;
	}

	.no-logs p {
		margin: 0;
		color: var(--text-muted);
		font-size: 1.1rem;
	}

	.logs-table-container {
		overflow-x: auto;
	}

	.logs-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.9rem;
	}

	.logs-table th {
		background: var(--bg-secondary);
		padding: var(--space-md);
		text-align: left;
		font-weight: 600;
		color: var(--text);
		border-bottom: 1px solid var(--border);
		white-space: nowrap;
	}

	.logs-table td {
		padding: var(--space-md);
		border-bottom: 1px solid var(--border);
		vertical-align: top;
	}

	.log-row:hover {
		background: var(--bg-secondary);
	}

	.log-row.expanded {
		background: var(--bg-secondary);
		border-left: 3px solid var(--primary);
	}

	.timestamp {
		font-family: monospace;
		font-size: 0.85rem;
		color: var(--text-muted);
		white-space: nowrap;
	}

	.user-cell {
		min-width: 180px;
	}

	.user-info {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.user-email {
		font-weight: 500;
		color: var(--text);
	}

	.user-id {
		font-size: 0.75rem;
		color: var(--text-muted);
		font-family: monospace;
	}

	.no-user {
		color: var(--text-muted);
		font-style: italic;
		font-size: 0.85rem;
	}

	.event-description {
		font-size: 0.75rem;
		color: var(--text-muted);
		margin-top: 0.25rem;
		font-style: italic;
	}

	.event-type code {
		background: var(--bg-secondary);
		padding: 2px var(--space-xs);
		border-radius: var(--radius-sm);
		font-size: 0.8rem;
		color: var(--text);
		font-family: monospace;
	}

	.status-badge {
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-md);
		font-size: 0.8rem;
		font-weight: 600;
		text-transform: uppercase;
	}

	.status-success {
		background: var(--success-light);
		color: var(--success-dark);
	}

	.status-error {
		background: var(--error-light);
		color: var(--error-dark);
	}

	.status-pending {
		background: var(--warning-light);
		color: var(--warning-dark);
	}

	.response-time {
		font-family: monospace;
		font-size: 0.85rem;
		color: var(--text-muted);
	}

	.payload-size {
		font-family: monospace;
		font-size: 0.85rem;
		color: var(--text-muted);
	}

	.status-code-badge {
		padding: 2px var(--space-xs);
		border-radius: var(--radius-sm);
		font-size: 0.8rem;
		font-weight: 600;
		font-family: monospace;
	}

	.status-code-2xx {
		background: var(--success-light);
		color: var(--success-dark);
	}

	.status-code-4xx {
		background: var(--warning-light);
		color: var(--warning-dark);
	}

	.status-code-5xx {
		background: var(--error-light);
		color: var(--error-dark);
	}

	.error-message {
		position: relative;
	}

	.error-message.status-2xx {
		border-left: 4px solid var(--success);
		background: var(--success-light);
	}

	.error-message.status-3xx {
		border-left: 4px solid var(--primary);
		background: var(--primary-light);
	}

	.error-message.status-4xx {
		border-left: 4px solid var(--warning);
		background: var(--warning-light);
	}

	.error-message.status-5xx {
		border-left: 4px solid var(--error);
		background: var(--error-light);
	}

	.error {
		color: var(--error);
		font-size: 0.85rem;
		cursor: help;
	}

	.no-error {
		color: var(--text-muted);
		font-style: italic;
	}

	.actions {
		text-align: center;
		min-width: 150px;
		display: flex;
		gap: 0.5rem;
		justify-content: center;
		align-items: center;
	}

	.btn-details {
		background: var(--bg-secondary);
		color: var(--text);
		border: 1px solid var(--border);
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-size: 0.75rem;
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		transition: all 0.2s ease;
		white-space: nowrap;
	}

	.btn-details:hover {
		background: var(--primary);
		color: white;
		border-color: var(--primary);
	}

	.btn-retry {
		background: var(--primary);
		color: white;
		border: none;
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-size: 0.75rem;
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		transition: all 0.2s ease;
	}

	.btn-retry:hover:not(:disabled) {
		background: var(--primary-dark);
		transform: translateY(-1px);
	}

	.btn-retry:disabled {
		opacity: 0.6;
		cursor: not-allowed;
		transform: none;
	}

	/* Details accordion */
	.details-row {
		background: var(--surface);
		border-left: 3px solid var(--primary);
	}

	.details-row td {
		padding: 0 !important;
	}

	.details-content {
		padding: var(--space-lg);
		animation: slideDown 0.3s ease-out;
	}

	@keyframes slideDown {
		from {
			opacity: 0;
			transform: translateY(-10px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	.details-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: var(--space-lg);
		margin-bottom: var(--space-lg);
	}

	.detail-section h4 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
		font-size: 1rem;
		font-weight: 600;
		border-bottom: 2px solid var(--border);
		padding-bottom: var(--space-xs);
	}

	.detail-items {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}

	.detail-row {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: var(--space-sm);
		padding: var(--space-xs) 0;
		border-bottom: 1px solid var(--border);
	}

	.detail-row:last-child {
		border-bottom: none;
	}

	.detail-row .label {
		font-size: 0.85rem;
		color: var(--text-muted);
		font-weight: 500;
		white-space: nowrap;
	}

	.detail-row .value {
		font-size: 0.85rem;
		color: var(--text);
		font-weight: 600;
		text-align: right;
		word-break: break-all;
	}

	.detail-row .value code {
		background: var(--bg-secondary);
		padding: 2px var(--space-xs);
		border-radius: var(--radius-sm);
		font-size: 0.75rem;
		font-family: monospace;
	}

	.error-section,
	.json-section {
		margin-top: var(--space-lg);
		padding: var(--space-md);
		background: var(--bg-secondary);
		border-radius: var(--radius-md);
		border: 1px solid var(--border);
	}

	.error-section h4,
	.json-section h4 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
		font-size: 1rem;
		font-weight: 600;
	}

	.error-pre,
	.json-pre {
		margin: 0;
		padding: var(--space-md);
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		font-family: monospace;
		font-size: 0.8rem;
		color: var(--text);
		overflow-x: auto;
		white-space: pre-wrap;
		word-break: break-word;
	}

	.error-pre {
		color: var(--error);
	}

	.status-active {
		background: var(--success-light);
		color: var(--success-dark);
	}

	.status-unpaid {
		background: var(--warning-light);
		color: var(--warning-dark);
	}

	.status-canceled {
		background: var(--error-light);
		color: var(--error-dark);
	}

	.max-retries {
		color: var(--warning);
		font-size: 0.75rem;
		font-weight: 500;
	}

	.no-action {
		color: var(--text-muted);
		font-style: italic;
		font-size: 0.75rem;
	}

	.logs-pagination {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-lg);
		border-top: 1px solid var(--border);
	}

	.pagination-info {
		font-size: 0.9rem;
		color: var(--text-muted);
		font-weight: 500;
	}

	@media (max-width: 768px) {
		.logs-filters {
			flex-direction: column;
			align-items: stretch;
		}

		.filter-actions {
			margin-left: 0;
			justify-content: stretch;
		}

		.filter-actions .btn {
			flex: 1;
		}

		.logs-table {
			font-size: 0.8rem;
		}

		.logs-table th,
		.logs-table td {
			padding: var(--space-sm);
		}

		.logs-pagination {
			flex-direction: column;
			gap: var(--space-md);
		}
	}
</style>
