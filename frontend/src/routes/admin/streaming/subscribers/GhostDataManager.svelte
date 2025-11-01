<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	interface GhostEntry {
		id: number;
		ghost_type: string;
		stripe_id: string;
		ghost_reason: string;
		referenced_by: any;
		first_detected_at: string;
		last_seen_at: string;
		attempted_syncs: number;
		metadata: any;
		notes: string;
	}

	interface GhostReport {
		total_ghosts: number;
		ghost_products: GhostEntry[];
		ghost_prices: GhostEntry[];
		ghost_subscriptions: GhostEntry[];
		ghost_customers: GhostEntry[];
		last_updated: string;
	}

	interface Props {
		onGhostCountUpdate?: () => Promise<void>;
	}

	let { onGhostCountUpdate }: Props = $props();

	let loading = $state(true);
	let ghostReport: GhostReport | null = $state(null);
	let expandedSection: string | null = $state(null);
	let refreshing = $state(false);

	onMount(async () => {
		await loadGhostData();
	});

	async function loadGhostData() {
		loading = true;
		try {
			const res = await apiRequest('/admin/streaming/ghosts');
			if (res.ok) {
				ghostReport = await res.json();
				console.log('👻 Ghost data loaded:', ghostReport);
			} else {
				showToast('Failed to load ghost data', 'error');
			}
		} catch (err) {
			console.error('Failed to load ghost data:', err);
			showToast('Failed to load ghost data', 'error');
		} finally {
			loading = false;
		}
	}

	async function refresh() {
		refreshing = true;
		await loadGhostData();
		if (onGhostCountUpdate) {
			await onGhostCountUpdate();
		}
		refreshing = false;
		showToast('Ghost data refreshed', 'success');
	}

	function toggleSection(section: string) {
		expandedSection = expandedSection === section ? null : section;
	}

	function getStripeUrl(stripeId: string, type: string): string {
		const baseUrl = 'https://dashboard.stripe.com';
		switch (type) {
			case 'product':
				return `${baseUrl}/products/${stripeId}`;
			case 'price':
				return `${baseUrl}/prices/${stripeId}`;
			case 'subscription':
				return `${baseUrl}/subscriptions/${stripeId}`;
			case 'customer':
				return `${baseUrl}/customers/${stripeId}`;
			default:
				return baseUrl;
		}
	}

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function formatCurrency(amount: number | undefined, currency: string = 'usd'): string {
		if (amount === undefined) return 'N/A';
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: currency.toUpperCase()
		}).format(amount / 100);
	}
</script>

<div class="ghost-manager">
	<div class="ghost-header">
		<div class="ghost-header-content">
			<h2>👻 Ghost Data Detected</h2>
			<p class="ghost-description">
				These items were blocked because they reference deleted or invalid Stripe objects. 
				Fix them in <a href="https://dashboard.stripe.com" target="_blank" rel="noopener noreferrer">Stripe Dashboard</a>, 
				then they'll automatically sync on the next webhook event.
			</p>
		</div>
		<button class="refresh-button" onclick={refresh} disabled={refreshing}>
			{#if refreshing}
				<LoadingSpinner size="sm" />
			{:else}
				🔄
			{/if}
			Refresh
		</button>
	</div>

	{#if loading}
		<div class="loading-container">
			<LoadingSpinner />
			<p>Loading ghost data...</p>
		</div>
	{:else if ghostReport}
		<div class="ghost-summary">
			<div class="summary-card">
				<div class="summary-icon">📦</div>
				<div class="summary-content">
					<div class="summary-label">Ghost Products</div>
					<div class="summary-value">{ghostReport.ghost_products.length}</div>
				</div>
			</div>
			<div class="summary-card">
				<div class="summary-icon">💰</div>
				<div class="summary-content">
					<div class="summary-label">Ghost Prices</div>
					<div class="summary-value">{ghostReport.ghost_prices.length}</div>
				</div>
			</div>
			<div class="summary-card">
				<div class="summary-icon">📋</div>
				<div class="summary-content">
					<div class="summary-label">Ghost Subscriptions</div>
					<div class="summary-value">{ghostReport.ghost_subscriptions.length}</div>
				</div>
			</div>
			<div class="summary-card">
				<div class="summary-icon">👤</div>
				<div class="summary-content">
					<div class="summary-label">Ghost Customers</div>
					<div class="summary-value">{ghostReport.ghost_customers.length}</div>
				</div>
			</div>
		</div>

		<!-- Ghost Products -->
		{#if ghostReport.ghost_products.length > 0}
			<div class="ghost-section">
				<button 
					class="section-header" 
					class:expanded={expandedSection === 'products'}
					onclick={() => toggleSection('products')}
				>
					<span class="section-icon">{expandedSection === 'products' ? '▼' : '▶'}</span>
					<span class="section-title">Ghost Products ({ghostReport.ghost_products.length})</span>
				</button>
				{#if expandedSection === 'products'}
					<div class="section-content">
						{#each ghostReport.ghost_products as ghost}
							<div class="ghost-card">
								<div class="ghost-card-header">
									<span class="ghost-id">{ghost.stripe_id}</span>
									<a href={getStripeUrl(ghost.stripe_id, 'product')} target="_blank" rel="noopener noreferrer" class="stripe-link">
										View in Stripe →
									</a>
								</div>
								<div class="ghost-details">
									<div class="detail-row">
										<span class="detail-label">Name:</span>
										<span class="detail-value">{ghost.metadata?.name || 'N/A'}</span>
									</div>
									<div class="detail-row">
										<span class="detail-label">Reason:</span>
										<span class="detail-value">{ghost.ghost_reason}</span>
									</div>
									<div class="detail-row">
										<span class="detail-label">First seen:</span>
										<span class="detail-value">{formatDate(ghost.first_detected_at)}</span>
									</div>
									<div class="detail-row">
										<span class="detail-label">Last attempt:</span>
										<span class="detail-value">{formatDate(ghost.last_seen_at)} ({ghost.attempted_syncs} attempts)</span>
									</div>
									{#if ghost.notes}
										<div class="detail-row notes">
											<span class="detail-label">Notes:</span>
											<span class="detail-value">{ghost.notes}</span>
										</div>
									{/if}
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		{/if}

		<!-- Ghost Prices -->
		{#if ghostReport.ghost_prices.length > 0}
			<div class="ghost-section">
				<button 
					class="section-header" 
					class:expanded={expandedSection === 'prices'}
					onclick={() => toggleSection('prices')}
				>
					<span class="section-icon">{expandedSection === 'prices' ? '▼' : '▶'}</span>
					<span class="section-title">Ghost Prices ({ghostReport.ghost_prices.length})</span>
				</button>
				{#if expandedSection === 'prices'}
					<div class="section-content">
						{#each ghostReport.ghost_prices as ghost}
							<div class="ghost-card">
								<div class="ghost-card-header">
									<span class="ghost-id">{ghost.stripe_id}</span>
									<a href={getStripeUrl(ghost.stripe_id, 'price')} target="_blank" rel="noopener noreferrer" class="stripe-link">
										View in Stripe →
									</a>
								</div>
								<div class="ghost-details">
									<div class="detail-row">
										<span class="detail-label">Amount:</span>
										<span class="detail-value">{formatCurrency(ghost.metadata?.unit_amount, ghost.metadata?.currency)}</span>
									</div>
									<div class="detail-row">
										<span class="detail-label">Ghost Product:</span>
										<span class="detail-value highlight-error">{ghost.referenced_by?.ghost_product || 'N/A'}</span>
									</div>
									<div class="detail-row">
										<span class="detail-label">Reason:</span>
										<span class="detail-value">{ghost.ghost_reason}</span>
									</div>
									<div class="detail-row">
										<span class="detail-label">Last attempt:</span>
										<span class="detail-value">{formatDate(ghost.last_seen_at)} ({ghost.attempted_syncs} attempts)</span>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		{/if}

		<!-- Ghost Subscriptions -->
		{#if ghostReport.ghost_subscriptions.length > 0}
			<div class="ghost-section">
				<button 
					class="section-header" 
					class:expanded={expandedSection === 'subscriptions'}
					onclick={() => toggleSection('subscriptions')}
				>
					<span class="section-icon">{expandedSection === 'subscriptions' ? '▼' : '▶'}</span>
					<span class="section-title">Ghost Subscriptions ({ghostReport.ghost_subscriptions.length})</span>
				</button>
				{#if expandedSection === 'subscriptions'}
					<div class="section-content">
						{#each ghostReport.ghost_subscriptions as ghost}
							<div class="ghost-card warning">
								<div class="ghost-card-header">
									<span class="ghost-id">{ghost.stripe_id}</span>
									<a href={getStripeUrl(ghost.stripe_id, 'subscription')} target="_blank" rel="noopener noreferrer" class="stripe-link">
										View in Stripe →
									</a>
								</div>
								<div class="ghost-details">
									<div class="detail-row">
										<span class="detail-label">Customer:</span>
										<span class="detail-value">{ghost.referenced_by?.customer_email || 'N/A'}</span>
									</div>
									<div class="detail-row">
										<span class="detail-label">Status:</span>
										<span class="detail-value status-badge" class:active={ghost.metadata?.status === 'active'}>
											{ghost.metadata?.status || 'unknown'}
										</span>
									</div>
									<div class="detail-row">
										<span class="detail-label">Ghost Product:</span>
										<span class="detail-value highlight-error">{ghost.referenced_by?.ghost_product || 'N/A'}</span>
									</div>
									<div class="detail-row">
										<span class="detail-label">Amount:</span>
										<span class="detail-value">{formatCurrency(ghost.metadata?.unit_amount, ghost.metadata?.currency)}</span>
									</div>
									<div class="detail-row">
										<span class="detail-label">Last attempt:</span>
										<span class="detail-value">{formatDate(ghost.last_seen_at)} ({ghost.attempted_syncs} attempts)</span>
									</div>
									{#if ghost.notes}
										<div class="detail-row notes warning-note">
											<span class="detail-label">⚠️  Warning:</span>
											<span class="detail-value">{ghost.notes}</span>
										</div>
									{/if}
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		{/if}

		<!-- Ghost Customers -->
		{#if ghostReport.ghost_customers.length > 0}
			<div class="ghost-section">
				<button 
					class="section-header" 
					class:expanded={expandedSection === 'customers'}
					onclick={() => toggleSection('customers')}
				>
					<span class="section-icon">{expandedSection === 'customers' ? '▼' : '▶'}</span>
					<span class="section-title">Ghost Customers ({ghostReport.ghost_customers.length})</span>
				</button>
				{#if expandedSection === 'customers'}
					<div class="section-content">
						{#each ghostReport.ghost_customers as ghost}
							<div class="ghost-card">
								<div class="ghost-card-header">
									<span class="ghost-id">{ghost.stripe_id}</span>
									<a href={getStripeUrl(ghost.stripe_id, 'customer')} target="_blank" rel="noopener noreferrer" class="stripe-link">
										View in Stripe →
									</a>
								</div>
								<div class="ghost-details">
									<div class="detail-row">
										<span class="detail-label">Reason:</span>
										<span class="detail-value">{ghost.ghost_reason}</span>
									</div>
									<div class="detail-row">
										<span class="detail-label">Last attempt:</span>
										<span class="detail-value">{formatDate(ghost.last_seen_at)} ({ghost.attempted_syncs} attempts)</span>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		{/if}

		{#if ghostReport.total_ghosts === 0}
			<div class="no-ghosts">
				<div class="no-ghosts-icon">✨</div>
				<h3>No Ghost Data!</h3>
				<p>All Stripe data is syncing successfully.</p>
			</div>
		{/if}
	{/if}
</div>

<style>
	.ghost-manager {
		padding: 2rem;
	}

	.ghost-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 2rem;
		gap: 2rem;
	}

	.ghost-header-content h2 {
		font-size: 1.75rem;
		font-weight: 700;
		color: #111827;
		margin: 0 0 0.5rem 0;
	}

	.ghost-description {
		color: #6b7280;
		margin: 0;
		line-height: 1.6;
	}

	.ghost-description a {
		color: #2563eb;
		text-decoration: underline;
	}

	.refresh-button {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1.5rem;
		background: linear-gradient(135deg, #2563eb, #1d4ed8);
		color: white;
		border: none;
		border-radius: 0.5rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s;
		white-space: nowrap;
	}

	.refresh-button:hover:not(:disabled) {
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
	}

	.refresh-button:disabled {
		opacity: 0.7;
		cursor: not-allowed;
	}

	.loading-container {
		text-align: center;
		padding: 4rem 0;
	}

	.ghost-summary {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
		margin-bottom: 2rem;
	}

	.summary-card {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1.5rem;
		background: linear-gradient(135deg, #f9fafb, #ffffff);
		border: 1px solid #e5e7eb;
		border-radius: 0.75rem;
		transition: all 0.2s;
	}

	.summary-card:hover {
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
		transform: translateY(-2px);
	}

	.summary-icon {
		font-size: 2rem;
	}

	.summary-label {
		font-size: 0.875rem;
		color: #6b7280;
		margin-bottom: 0.25rem;
	}

	.summary-value {
		font-size: 1.75rem;
		font-weight: 700;
		color: #111827;
	}

	.ghost-section {
		margin-bottom: 1rem;
		border: 1px solid #e5e7eb;
		border-radius: 0.75rem;
		overflow: hidden;
	}

	.section-header {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		width: 100%;
		padding: 1rem 1.5rem;
		background: #f9fafb;
		border: none;
		cursor: pointer;
		transition: background 0.2s;
		font-size: 1.125rem;
		font-weight: 600;
		color: #111827;
		text-align: left;
	}

	.section-header:hover {
		background: #f3f4f6;
	}

	.section-header.expanded {
		background: #ede9fe;
		color: #9333ea;
	}

	.section-icon {
		font-size: 0.875rem;
		transition: transform 0.2s;
	}

	.section-content {
		padding: 1rem;
		background: white;
	}

	.ghost-card {
		padding: 1.5rem;
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		margin-bottom: 1rem;
		transition: all 0.2s;
	}

	.ghost-card:hover {
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}

	.ghost-card.warning {
		border-left: 4px solid #f59e0b;
		background: #fffbeb;
	}

	.ghost-card-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
	}

	.ghost-id {
		font-family: 'Courier New', monospace;
		font-size: 0.875rem;
		font-weight: 600;
		color: #6b7280;
		background: #f3f4f6;
		padding: 0.25rem 0.75rem;
		border-radius: 0.25rem;
	}

	.stripe-link {
		color: #2563eb;
		text-decoration: none;
		font-weight: 600;
		font-size: 0.875rem;
		transition: color 0.2s;
	}

	.stripe-link:hover {
		color: #1d4ed8;
		text-decoration: underline;
	}

	.ghost-details {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.detail-row {
		display: flex;
		gap: 0.75rem;
	}

	.detail-row.notes {
		flex-direction: column;
		gap: 0.25rem;
		padding: 1rem;
		background: #f9fafb;
		border-radius: 0.375rem;
		border-left: 3px solid #6b7280;
	}

	.detail-row.warning-note {
		background: #fef3c7;
		border-left-color: #f59e0b;
	}

	.detail-label {
		font-weight: 600;
		color: #374151;
		min-width: 140px;
	}

	.detail-value {
		color: #6b7280;
		flex: 1;
	}

	.highlight-error {
		color: #dc2626;
		font-weight: 600;
		font-family: 'Courier New', monospace;
	}

	.status-badge {
		padding: 0.25rem 0.75rem;
		border-radius: 9999px;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		background: #e5e7eb;
		color: #6b7280;
	}

	.status-badge.active {
		background: #d1fae5;
		color: #059669;
	}

	.no-ghosts {
		text-align: center;
		padding: 4rem 2rem;
	}

	.no-ghosts-icon {
		font-size: 4rem;
		margin-bottom: 1rem;
	}

	.no-ghosts h3 {
		font-size: 1.5rem;
		font-weight: 700;
		color: #111827;
		margin: 0 0 0.5rem 0;
	}

	.no-ghosts p {
		color: #6b7280;
		margin: 0;
	}

	@media (max-width: 768px) {
		.ghost-header {
			flex-direction: column;
		}

		.ghost-summary {
			grid-template-columns: 1fr;
		}

		.detail-row {
			flex-direction: column;
			gap: 0.25rem;
		}

		.detail-label {
			min-width: unset;
		}
	}
</style>

