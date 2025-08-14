<script lang="ts">
	import { onMount } from 'svelte';
	import { StripeCustomerSyncService, type CustomerSyncResult, type CustomerSyncStats } from '$lib/services/stripe-customer-sync';
	import { StreamingSubscriberService, type Subscriber } from '$lib/services/streaming-subscribers';
	import { showToast } from '$lib/toast';

	export let customerId: number | null = null;
	export let customerEmail: string = '';

	let syncStatus: CustomerSyncResult | null = null;
	let loading = false;
	let syncing = false;
	let lastSyncStats: CustomerSyncStats | null = null;
	
	// Active subscribers sync
	let activeSubscribers: Subscriber[] = [];
	let activeSubscribersLoading = false;
	let activeSubscribersCount = 0;
	let activeSubscribersSyncing = false;
	let activeSubscribersSyncStats: CustomerSyncStats | null = null;

	onMount(async () => {
		if (customerId) {
			await loadSyncStatus();
		}
		// Always load active subscribers for the sync panel
		await loadActiveSubscribers();
	});

	async function loadSyncStatus() {
		if (!customerId) return;

		try {
			loading = true;
			syncStatus = await StripeCustomerSyncService.getSyncStatus(customerId);
		} catch (error) {
			console.error('Failed to load sync status:', error);
			showToast('Failed to load sync status', 'error');
		} finally {
			loading = false;
		}
	}

	async function syncToStripe() {
		if (!customerId) return;

		try {
			syncing = true;
			const result = await StripeCustomerSyncService.syncCustomerToStripe(customerId);
			syncStatus = result;
			showToast(`Customer synced to Stripe: ${result.message}`, 'success');
		} catch (error) {
			console.error('Failed to sync to Stripe:', error);
			showToast('Failed to sync customer to Stripe', 'error');
		} finally {
			syncing = false;
		}
	}

	async function syncFromStripe() {
		if (!syncStatus?.stripe_id) {
			showToast('No Stripe customer ID found', 'warning');
			return;
		}

		try {
			syncing = true;
			const result = await StripeCustomerSyncService.syncCustomerFromStripe(syncStatus.stripe_id);
			syncStatus = result;
			showToast(`Customer synced from Stripe: ${result.message}`, 'success');
		} catch (error) {
			console.error('Failed to sync from Stripe:', error);
			showToast('Failed to sync customer from Stripe', 'error');
		} finally {
			syncing = false;
		}
	}

	async function syncAllCustomers() {
		try {
			syncing = true;
			const stats = await StripeCustomerSyncService.syncAllCustomers();
			lastSyncStats = stats;
			showToast(`Sync completed: ${stats.created} created, ${stats.updated} updated, ${stats.errors} errors`, 'success');
		} catch (error) {
			console.error('Failed to sync all customers:', error);
			showToast('Failed to sync all customers', 'error');
		} finally {
			syncing = false;
		}
	}

	// Load active subscribers from your existing API
	async function loadActiveSubscribers() {
		try {
			activeSubscribersLoading = true;
			const response = await StreamingSubscriberService.getSubscribers({
				limit: 1000, // Get all active subscribers
				offset: 0
			});
			
			// Filter for only active status subscribers
			activeSubscribers = (response.subscribers || []).filter(sub => 
				sub.subscription_status === 'active'
			);
			activeSubscribersCount = activeSubscribers.length;
			
			console.log(`Loaded ${activeSubscribersCount} active subscribers`);
		} catch (error) {
			console.error('Failed to load active subscribers:', error);
			showToast('Failed to load active subscribers', 'error');
		} finally {
			activeSubscribersLoading = false;
		}
	}

	// Sync active subscribers to Stripe
	async function syncActiveSubscribersToStripe() {
		if (activeSubscribersCount === 0) {
			showToast('No active subscribers found', 'warning');
			return;
		}

		try {
			activeSubscribersSyncing = true;
			
			// Extract customer IDs from active subscribers
			const customerIds = activeSubscribers.map(sub => sub.id);
			
			// Bulk sync to Stripe
			const stats = await StripeCustomerSyncService.bulkSyncCustomersToStripe(customerIds);
			activeSubscribersSyncStats = stats;
			
			showToast(`Active subscribers sync completed: ${stats.created} created, ${stats.updated} updated, ${stats.errors} errors`, 'success');
		} catch (error) {
			console.error('Failed to sync active subscribers:', error);
			showToast('Failed to sync active subscribers', 'error');
		} finally {
			activeSubscribersSyncing = false;
		}
	}

	$: actionDescription = syncStatus ? StripeCustomerSyncService.getActionDescription(syncStatus.action) : '';
	$: actionColor = syncStatus ? StripeCustomerSyncService.getActionColor(syncStatus.action) : '';
</script>

<div class="customer-sync-panel">
	<details class="accordion">
		<summary class="accordion-header">
			<h3>🔄 Customer Sync</h3>
			<span class="accordion-icon">▼</span>
		</summary>
		<div class="accordion-content">
			<p>Keep customer data synchronized between your database and Stripe</p>

			{#if customerId}
				<!-- Individual Customer Sync -->
				<div class="sync-section">
					<h4>Sync Customer: {customerEmail}</h4>
					
					{#if loading}
						<div class="loading">Loading sync status...</div>
					{:else if syncStatus}
						<div class="sync-status">
							<div class="status-indicator" style="color: {actionColor}">
								● {actionDescription}
							</div>
							{#if syncStatus.stripe_id}
								<div class="stripe-id">Stripe ID: {syncStatus.stripe_id}</div>
							{/if}
							{#if syncStatus.last_sync_at}
								<div class="last-sync">Last sync: {new Date(syncStatus.last_sync_at).toLocaleString()}</div>
							{/if}
							{#if syncStatus.error}
								<div class="sync-error">Error: {syncStatus.error}</div>
							{/if}
						</div>

						<div class="sync-actions">
							<button 
								class="btn btn-primary" 
								on:click={syncToStripe}
								disabled={syncing}
							>
								{syncing ? '🔄 Syncing...' : '📤 Sync to Stripe'}
							</button>
							
							{#if syncStatus.stripe_id}
								<button 
									class="btn btn-secondary" 
									on:click={syncFromStripe}
									disabled={syncing}
								>
									{syncing ? '🔄 Syncing...' : '📥 Sync from Stripe'}
								</button>
							{/if}
						</div>
					{/if}
				</div>
			{/if}

			<!-- Global Sync -->
			<div class="sync-section">
				<h4>Global Customer Sync</h4>
				<p>Sync all customers between your database and Stripe</p>
				
				<button 
					class="btn btn-warning" 
					on:click={syncAllCustomers}
					disabled={syncing}
				>
					{syncing ? '🔄 Syncing All...' : '🔄 Sync All Customers'}
				</button>

				{#if lastSyncStats}
					<div class="sync-stats">
						<h5>Last Sync Results</h5>
						<div class="stats-grid">
							<div class="stat">
								<span class="stat-value">{lastSyncStats.total_processed}</span>
								<span class="stat-label">Total</span>
							</div>
							<div class="stat">
								<span class="stat-value">{lastSyncStats.created}</span>
								<span class="stat-label">Created</span>
							</div>
							<div class="stat">
								<span class="stat-value">{lastSyncStats.updated}</span>
								<span class="stat-label">Updated</span>
							</div>
							<div class="stat">
								<span class="stat-value">{lastSyncStats.errors}</span>
								<span class="stat-label">Errors</span>
							</div>
						</div>
						<div class="sync-duration">Duration: {lastSyncStats.duration}</div>
					</div>
				{/if}
			</div>

			<!-- Active Subscribers Sync -->
			<div class="sync-section">
				<h4>🟢 Active Subscribers Sync</h4>
				<p>Sync only your active subscribers to Stripe. This ensures your paying customers are always up-to-date.</p>
				
				<div class="active-subscribers-info">
					{#if activeSubscribersLoading}
						<div class="loading">Loading active subscribers...</div>
					{:else if activeSubscribersCount > 0}
						<div class="subscribers-summary">
							<span class="subscribers-count">{activeSubscribersCount} active subscribers found</span>
							<button 
								class="btn btn-success" 
								on:click={syncActiveSubscribersToStripe}
								disabled={activeSubscribersSyncing}
							>
								{activeSubscribersSyncing ? '🔄 Syncing...' : '🟢 Sync Active Subscribers to Stripe'}
							</button>
						</div>
					{:else}
						<div class="no-subscribers">
							<span>No active subscribers loaded</span>
							<button 
								class="btn btn-secondary" 
								on:click={loadActiveSubscribers}
							>
								📊 Load Active Subscribers
							</button>
						</div>
					{/if}
				</div>

				{#if activeSubscribersSyncStats}
					<div class="sync-stats">
						<h5>Active Subscribers Sync Results</h5>
						<div class="stats-grid">
							<div class="stat">
								<span class="stat-value">{activeSubscribersSyncStats.total_processed}</span>
								<span class="stat-label">Total</span>
							</div>
							<div class="stat">
								<span class="stat-value">{activeSubscribersSyncStats.created}</span>
								<span class="stat-label">Created</span>
							</div>
							<div class="stat">
								<span class="stat-value">{activeSubscribersSyncStats.updated}</span>
								<span class="stat-label">Updated</span>
							</div>
							<div class="stat">
								<span class="stat-value">{activeSubscribersSyncStats.errors}</span>
								<span class="stat-label">Errors</span>
							</div>
						</div>
						<div class="sync-duration">Duration: {activeSubscribersSyncStats.duration}</div>
					</div>
				{/if}

				{#if activeSubscribersCount > 0 && !activeSubscribersSyncStats}
					<div class="subscribers-preview">
						<h5>Active Subscribers Preview</h5>
						<div class="subscribers-list">
							{#each activeSubscribers.slice(0, 5) as subscriber}
								<div class="subscriber-item">
									<span class="subscriber-name">{subscriber.first_name} {subscriber.last_name}</span>
									<span class="subscriber-email">{subscriber.email}</span>
									<span class="subscriber-plan">{subscriber.plan_name || 'No Plan'}</span>
								</div>
							{/each}
							{#if activeSubscribersCount > 5}
								<div class="subscriber-item more">
									<span>... and {activeSubscribersCount - 5} more</span>
								</div>
							{/if}
						</div>
					</div>
				{/if}
			</div>

			<!-- Sync Info -->
			<div class="sync-info">
				<h4>How It Works</h4>
				<ul>
					<li><strong>Sync to Stripe:</strong> Creates or updates customer in Stripe with local data</li>
					<li><strong>Sync from Stripe:</strong> Updates local customer with Stripe data</li>
					<li><strong>Global Sync:</strong> Automatically syncs all customers in both directions</li>
					<li><strong>Metadata:</strong> Preserves local customer ID and role in Stripe metadata</li>
				</ul>
			</div>
		</div>
	</details>
</div>

<style>
	.customer-sync-panel {
		margin-bottom: var(--space-lg);
	}

	.accordion {
		background: var(--surface, #ffffff);
		border: 1px solid var(--border, #e5e7eb);
		border-radius: var(--radius-lg, 0.5rem);
		overflow: hidden;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
	}

	.accordion-header {
		padding: var(--space-lg, 1.5rem);
		background: var(--bg-secondary, #f9fafb);
		color: var(--text, #111827);
		font-size: 1.1rem;
		font-weight: 600;
		cursor: pointer;
		display: flex;
		justify-content: space-between;
		align-items: center;
		transition: background-color 0.2s ease;
		list-style: none;
	}

	.accordion-header::-webkit-details-marker {
		display: none;
	}

	.accordion-header:hover {
		background: var(--bg-hover, #f3f4f6);
	}

	.accordion-header:focus {
		outline: none;
		box-shadow: 0 0 0 2px var(--primary, #3b82f6);
	}

	.accordion-header h3 {
		margin: 0;
		color: var(--text, #111827);
		font-size: 1.25rem;
		font-weight: 700;
	}

	.accordion-icon {
		transition: transform 0.3s ease;
		font-size: 1rem;
		color: var(--text-muted, #6b7280);
	}

	.accordion[open] .accordion-icon {
		transform: rotate(180deg);
	}

	.accordion-content {
		padding: var(--space-lg, 1.5rem);
		background: var(--surface, #ffffff);
	}

	.accordion-content > p:first-child {
		margin: 0 0 var(--space-lg, 1.5rem) 0;
		color: var(--text-muted, #6b7280);
		font-size: 0.9rem;
	}

	.sync-section {
		margin: var(--space-lg) 0;
		padding: var(--space-md);
		background: var(--bg-secondary);
		border-radius: var(--radius-md);
	}

	.sync-section h4 {
		margin: 0 0 var(--space-sm) 0;
		color: var(--text);
		font-size: 1.1rem;
	}

	.sync-section p {
		margin: 0 0 var(--space-md) 0;
		color: var(--text-muted);
		font-size: 0.9rem;
	}

	.loading {
		text-align: center;
		color: var(--text-muted);
		font-style: italic;
	}

	.sync-status {
		margin-bottom: var(--space-md);
	}

	.status-indicator {
		font-weight: 600;
		font-size: 1rem;
		margin-bottom: var(--space-xs);
	}

	.stripe-id {
		font-family: var(--font-mono);
		font-size: 0.875rem;
		color: var(--text-muted);
		margin-bottom: var(--space-xs);
	}

	.last-sync {
		font-size: 0.875rem;
		color: var(--text-muted);
		margin-bottom: var(--space-xs);
	}

	.sync-error {
		color: var(--error);
		font-size: 0.875rem;
		font-weight: 500;
	}

	.sync-actions {
		display: flex;
		gap: var(--space-sm);
		flex-wrap: wrap;
	}

	.btn {
		padding: var(--space-sm) var(--space-md);
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
		background: var(--primary-hover);
		transform: translateY(-1px);
	}

	.btn-secondary {
		background: var(--secondary);
		color: white;
	}

	.btn-secondary:hover:not(:disabled) {
		background: var(--secondary-hover);
		transform: translateY(-1px);
	}

	.btn-warning {
		background: var(--warning);
		color: white;
	}

	.btn-warning:hover:not(:disabled) {
		background: var(--warning-hover);
		transform: translateY(-1px);
	}

	.btn-success {
		background: var(--success);
		color: white;
	}

	.btn-success:hover:not(:disabled) {
		background: var(--bg-green-200);
		color: var(--text-primary);
		transform: translateY(-1px);
	}

	.active-subscribers-info {
		margin-bottom: var(--space-md);
	}

	.subscribers-summary {
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: var(--space-md);
	}

	.subscribers-count {
		font-weight: 600;
		color: var(--success);
		font-size: 1.1rem;
	}

	.no-subscribers {
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: var(--space-md);
	}

	.subscribers-preview {
		margin-top: var(--space-md);
		padding: var(--space-md);
		background: var(--surface);
		border-radius: var(--radius-md);
		border: 1px solid var(--border);
	}

	.subscribers-preview h5 {
		margin: 0 0 var(--space-sm) 0;
		color: var(--text);
		font-size: 1rem;
	}

	.subscribers-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.subscriber-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-xs) var(--space-sm);
		background: var(--bg-secondary);
		border-radius: var(--radius-sm);
		font-size: 0.875rem;
	}

	.subscriber-item.more {
		justify-content: center;
		font-style: italic;
		color: var(--text-muted);
	}

	.subscriber-name {
		font-weight: 500;
		color: var(--text);
		min-width: 120px;
	}

	.subscriber-email {
		color: var(--text-muted);
		flex: 1;
		margin: 0 var(--space-sm);
	}

	.subscriber-plan {
		color: var(--primary);
		font-weight: 500;
		min-width: 80px;
		text-align: right;
	}

	.sync-stats {
		margin-top: var(--space-md);
		padding: var(--space-md);
		background: var(--surface);
		border-radius: var(--radius-md);
		border: 1px solid var(--border);
	}

	.sync-stats h5 {
		margin: 0 0 var(--space-sm) 0;
		color: var(--text);
		font-size: 1rem;
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: var(--space-sm);
		margin-bottom: var(--space-sm);
	}

	.stat {
		text-align: center;
	}

	.stat-value {
		display: block;
		font-size: 1.5rem;
		font-weight: bold;
		color: var(--primary);
	}

	.stat-label {
		font-size: 0.75rem;
		color: var(--text-muted);
		text-transform: uppercase;
	}

	.sync-duration {
		text-align: center;
		font-size: 0.875rem;
		color: var(--text-muted);
		font-family: var(--font-mono);
	}

	.sync-info {
		margin-top: var(--space-lg);
		padding: var(--space-md);
		background: var(--bg-secondary);
		border-radius: var(--radius-md);
	}

	.sync-info h4 {
		margin: 0 0 var(--space-sm) 0;
		color: var(--text);
		font-size: 1rem;
	}

	.sync-info ul {
		margin: 0;
		padding-left: var(--space-lg);
		color: var(--text-muted);
		font-size: 0.9rem;
	}

	.sync-info li {
		margin-bottom: var(--space-xs);
	}

	.sync-info strong {
		color: var(--text);
	}

	@media (max-width: 768px) {
		.accordion-header {
			padding: var(--space-md, 1rem);
			flex-direction: column;
			gap: var(--space-sm, 0.5rem);
			text-align: center;
		}

		.accordion-header h3 {
			font-size: 1.1rem;
		}

		.accordion-content {
			padding: var(--space-md, 1rem);
		}

		.sync-actions {
			flex-direction: column;
		}

		.stats-grid {
			grid-template-columns: repeat(2, 1fr);
		}

		.subscribers-summary,
		.no-subscribers {
			flex-direction: column;
			align-items: stretch;
			text-align: center;
		}

		.subscriber-item {
			flex-direction: column;
			align-items: stretch;
			gap: var(--space-xs, 0.25rem);
			text-align: center;
		}

		.subscriber-name,
		.subscriber-email,
		.subscriber-plan {
			min-width: auto;
			text-align: center;
			margin: 0;
		}
	}
</style> 
