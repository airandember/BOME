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
	
	// Ghost Subscriptions Analytics
	let selectedPlan = $state<string>('all');
	let selectedStatus = $state<string>('all');
	let searchEmail = $state<string>('');

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

	// Compute Ghost Subscriptions Analytics
	interface PlanAnalytics {
		planId: string;
		planName: string;
		totalCount: number;
		statusBreakdown: Record<string, number>;
		unitAmount: number;
		unitMRR: number;
		unitARR: number;
		currency: string;
		billingType: 'MRR' | 'Quarterly' | 'ARR';
		activeMRR: number;
		unpaidMRR: number;
		pastDueMRR: number;
		totalMRR: number;
		activeARR: number;
		unpaidARR: number;
		pastDueARR: number;
		totalARR: number;
		potentialARR: number;
	}

	// Smart billing detection algorithm
	function detectBillingType(unitAmountInCents: number): 'MRR' | 'Quarterly' | 'ARR' {
		const dollars = unitAmountInCents / 100;
		
		if (dollars <= 40) {
			return 'MRR'; // Monthly recurring revenue
		} else if (dollars === 45) {
			return 'Quarterly'; // $45 every 3 months = $180 ARR
		} else if (dollars > 70) {
			return 'ARR'; // Annual recurring revenue
		}
		
		// Default to MRR for edge cases
		return 'MRR';
	}

	// Convert unit amount to MRR (Monthly Recurring Revenue)
	function toMRR(unitAmountInCents: number, billingType: 'MRR' | 'Quarterly' | 'ARR'): number {
		if (billingType === 'MRR') {
			return unitAmountInCents; // Already monthly
		} else if (billingType === 'Quarterly') {
			return unitAmountInCents / 3; // Quarterly → Monthly ($45 / 3 months = $15/mo)
		} else { // ARR
			return unitAmountInCents / 12; // Annual → Monthly
		}
	}

	// Convert unit amount to ARR (Annual Recurring Revenue)
	function toARR(unitAmountInCents: number, billingType: 'MRR' | 'Quarterly' | 'ARR'): number {
		if (billingType === 'MRR') {
			return unitAmountInCents * 12; // Monthly → Annual
		} else if (billingType === 'Quarterly') {
			return unitAmountInCents * 4; // Quarterly → Annual ($45 × 4 quarters = $180/year)
		} else { // ARR
			return unitAmountInCents; // Already annual
		}
	}

	function computeSubscriptionAnalytics(subscriptions: GhostEntry[]): {
		planAnalytics: PlanAnalytics[];
		totalMRR: number;
		activeMRR: number;
		unpaidMRR: number;
		pastDueMRR: number;
		totalARR: number;
		activeARR: number;
		unpaidARR: number;
		pastDueARR: number;
		potentialARR: number;
		uniquePlans: string[];
		statusCounts: Record<string, number>;
	} {
		const planMap = new Map<string, PlanAnalytics>();
		const statusCounts: Record<string, number> = {};
		
		let totalMRR = 0;
		let activeMRR = 0;
		let unpaidMRR = 0;
		let pastDueMRR = 0;
		let totalARR = 0;
		let activeARR = 0;
		let unpaidARR = 0;
		let pastDueARR = 0;
		let potentialARR = 0;

		subscriptions.forEach(sub => {
			const planId = sub.referenced_by?.ghost_product || 'Unknown Plan';
			const status = sub.metadata?.status || 'unknown';
			const unitAmount = sub.metadata?.unit_amount || 0;
			const currency = sub.metadata?.currency || 'usd';
			const billingType = detectBillingType(unitAmount);

			// Calculate unit MRR and ARR for this plan
			const unitMRR = toMRR(unitAmount, billingType);
			const unitARR = toARR(unitAmount, billingType);

			// Initialize plan analytics if not exists
			if (!planMap.has(planId)) {
				planMap.set(planId, {
					planId,
					planName: planId,
					totalCount: 0,
					statusBreakdown: {},
					unitAmount,
					unitMRR,
					unitARR,
					currency,
					billingType,
					activeMRR: 0,
					unpaidMRR: 0,
					pastDueMRR: 0,
					totalMRR: 0,
					activeARR: 0,
					unpaidARR: 0,
					pastDueARR: 0,
					totalARR: 0,
					potentialARR: 0
				});
			}

			const plan = planMap.get(planId)!;
			plan.totalCount++;
			plan.statusBreakdown[status] = (plan.statusBreakdown[status] || 0) + 1;

			// Count statuses
			statusCounts[status] = (statusCounts[status] || 0) + 1;
		});

		// Second pass: Calculate revenue based on counts
		planMap.forEach(plan => {
			const activeCount = plan.statusBreakdown['active'] || 0;
			const unpaidCount = plan.statusBreakdown['unpaid'] || 0;
			const pastDueCount = plan.statusBreakdown['past_due'] || 0;

			// Active revenue = active subscribers * unit amount
			plan.activeMRR = activeCount * plan.unitMRR;
			plan.activeARR = activeCount * plan.unitARR;

			// Unpaid revenue = unpaid subscribers * unit amount
			plan.unpaidMRR = unpaidCount * plan.unitMRR;
			plan.unpaidARR = unpaidCount * plan.unitARR;

			// Past due revenue = past_due subscribers * unit amount
			plan.pastDueMRR = pastDueCount * plan.unitMRR;
			plan.pastDueARR = pastDueCount * plan.unitARR;

			// Total revenue (all statuses combined)
			plan.totalMRR = plan.totalCount * plan.unitMRR;
			plan.totalARR = plan.totalCount * plan.unitARR;

			// Potential ARR = ALL subscribers * unit ARR (if everyone was paying)
			plan.potentialARR = plan.totalCount * plan.unitARR;

			// Add to overall totals
			activeMRR += plan.activeMRR;
			activeARR += plan.activeARR;
			unpaidMRR += plan.unpaidMRR;
			unpaidARR += plan.unpaidARR;
			pastDueMRR += plan.pastDueMRR;
			pastDueARR += plan.pastDueARR;
			totalMRR += plan.totalMRR;
			totalARR += plan.totalARR;
			potentialARR += plan.potentialARR;
		});

		return {
			planAnalytics: Array.from(planMap.values()).sort((a, b) => b.potentialARR - a.potentialARR),
			totalMRR,
			activeMRR,
			unpaidMRR,
			pastDueMRR,
			totalARR,
			activeARR,
			unpaidARR,
			pastDueARR,
			potentialARR,
			uniquePlans: Array.from(planMap.keys()),
			statusCounts
		};
	}

	// Filtered subscriptions based on selected filters
	$effect(() => {
		if (!ghostReport) return;
	});

	function getFilteredSubscriptions(): GhostEntry[] {
		if (!ghostReport) return [];
		
		return ghostReport.ghost_subscriptions.filter(sub => {
			// Filter by plan
			if (selectedPlan !== 'all') {
				const planId = sub.referenced_by?.ghost_product || 'Unknown Plan';
				if (planId !== selectedPlan) return false;
			}

			// Filter by status
			if (selectedStatus !== 'all') {
				const status = sub.metadata?.status || 'unknown';
				if (status !== selectedStatus) return false;
			}

			// Filter by email search
			if (searchEmail.trim()) {
				const email = sub.referenced_by?.customer_email || '';
				if (!email.toLowerCase().includes(searchEmail.toLowerCase())) return false;
			}

			return true;
		});
	}

	function resetFilters() {
		selectedPlan = 'all';
		selectedStatus = 'all';
		searchEmail = '';
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
				<LoadingSpinner size="small" />
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

		<!-- Ghost Subscriptions (ENHANCED WITH ANALYTICS) -->
		{#if ghostReport.ghost_subscriptions.length > 0}
			{@const analytics = computeSubscriptionAnalytics(ghostReport.ghost_subscriptions)}
			{@const filteredSubs = getFilteredSubscriptions()}
			
			<div class="ghost-section subscriptions-section">
				<button 
					class="section-header critical" 
					class:expanded={expandedSection === 'subscriptions'}
					onclick={() => toggleSection('subscriptions')}
				>
					<span class="section-icon">{expandedSection === 'subscriptions' ? '▼' : '▶'}</span>
					<span class="section-title">
						⚠️ Ghost Subscriptions ({ghostReport.ghost_subscriptions.length}) 
						<span class="revenue-badge">{formatCurrency(analytics.potentialARR)} ARR</span>
					</span>
				</button>
				{#if expandedSection === 'subscriptions'}
					<div class="section-content">
						<!-- CRITICAL REVENUE ALERT -->
						<div class="critical-alert">
							<div class="alert-icon">🚨</div>
							<div class="alert-content">
								<h3>CRITICAL: Ghost Revenue Detected</h3>
								<p>These are <strong>REAL PAYING CUSTOMERS</strong> whose subscriptions aren't syncing due to the WordPress WPV plugin glitch.</p>
							</div>
						</div>

						<!-- REVENUE SUMMARY -->
						<div class="revenue-summary">
							<div class="revenue-card total">
								<div class="revenue-label">Total Annual Revenue (ARR)</div>
								<div class="revenue-value">{formatCurrency(analytics.totalARR)}</div>
								<div class="revenue-count">{ghostReport.ghost_subscriptions.length} subscriptions</div>
								<div class="revenue-mrr">MRR: {formatCurrency(analytics.totalMRR)}</div>
							</div>
							<div class="revenue-card active">
								<div class="revenue-label">Active/Trialing (ARR)</div>
								<div class="revenue-value">{formatCurrency(analytics.activeARR)}</div>
								<div class="revenue-count">{analytics.statusCounts['active'] || 0} active</div>
								<div class="revenue-mrr">MRR: {formatCurrency(analytics.activeMRR)}</div>
							</div>
							<div class="revenue-card unpaid">
								<div class="revenue-label">Unpaid (ARR)</div>
								<div class="revenue-value">{formatCurrency(analytics.unpaidARR)}</div>
								<div class="revenue-count">{analytics.statusCounts['unpaid'] || 0} unpaid</div>
								<div class="revenue-mrr">MRR: {formatCurrency(analytics.unpaidMRR)}</div>
							</div>
							<div class="revenue-card past-due">
								<div class="revenue-label">Past Due (ARR)</div>
								<div class="revenue-value">{formatCurrency(analytics.pastDueARR)}</div>
								<div class="revenue-count">{analytics.statusCounts['past_due'] || 0} past due</div>
								<div class="revenue-mrr">MRR: {formatCurrency(analytics.pastDueMRR)}</div>
							</div>
						</div>

						<!-- PLAN BREAKDOWN TABLE -->
						<div class="plan-analytics-section">
							<h3>📊 Plan Breakdown</h3>
							<div class="analytics-table">
								<table>
									<thead>
										<tr>
											<th>Plan ID</th>
											<th>Count</th>
											<th>Price</th>
											<th>Type</th>
											<th>Active</th>
											<th>Unpaid</th>
											<th>Past Due</th>
											<th>Active MRR</th>
											<th>Active ARR</th>
											<th>At-Risk ARR</th>
											<th>Potential ARR</th>
										</tr>
									</thead>
									<tbody>
										{#each analytics.planAnalytics as plan}
											<tr>
												<td class="plan-id">{plan.planId}</td>
												<td class="count">{plan.totalCount}</td>
												<td class="unit-price">{formatCurrency(plan.unitAmount, plan.currency)}</td>
												<td class="billing-type">
													<span class="type-badge {plan.billingType.toLowerCase()}">{plan.billingType}</span>
												</td>
											<td class="status-count active">{plan.statusBreakdown['active'] || 0}</td>
											<td class="status-count unpaid">{plan.statusBreakdown['unpaid'] || 0}</td>
											<td class="status-count past-due">{plan.statusBreakdown['past_due'] || 0}</td>
											<td class="revenue active-revenue">{formatCurrency(plan.activeMRR)}</td>
											<td class="revenue active-revenue">{formatCurrency(plan.activeARR)}</td>
											<td class="revenue risk-revenue">{formatCurrency(plan.unpaidARR + plan.pastDueARR)}</td>
											<td class="revenue potential-revenue">{formatCurrency(plan.potentialARR)}</td>
											</tr>
										{/each}
										<tr class="totals-row">
											<td colspan="4"><strong>TOTALS</strong></td>
											<td class="status-count active"><strong>{analytics.statusCounts['active'] || 0}</strong></td>
											<td class="status-count unpaid"><strong>{analytics.statusCounts['unpaid'] || 0}</strong></td>
											<td class="status-count past-due"><strong>{analytics.statusCounts['past_due'] || 0}</strong></td>
											<td class="revenue active-revenue"><strong>{formatCurrency(analytics.activeMRR)}</strong></td>
											<td class="revenue active-revenue"><strong>{formatCurrency(analytics.activeARR)}</strong></td>
											<td class="revenue risk-revenue"><strong>{formatCurrency(analytics.unpaidARR + analytics.pastDueARR)}</strong></td>
											<td class="revenue potential-revenue"><strong>{formatCurrency(analytics.potentialARR)}</strong></td>
										</tr>
									</tbody>
								</table>
							</div>
						</div>

						<!-- FILTERS -->
						<div class="filters-section">
							<h3>🔍 Filter Subscriptions</h3>
							<div class="filters">
								<div class="filter-group">
									<label for="plan-filter">Plan:</label>
									<select id="plan-filter" bind:value={selectedPlan}>
										<option value="all">All Plans ({ghostReport.ghost_subscriptions.length})</option>
										{#each analytics.uniquePlans as planId}
											{@const count = analytics.planAnalytics.find(p => p.planId === planId)?.totalCount || 0}
											<option value={planId}>{planId} ({count})</option>
										{/each}
									</select>
								</div>

								<div class="filter-group">
									<label for="status-filter">Status:</label>
									<select id="status-filter" bind:value={selectedStatus}>
										<option value="all">All Statuses</option>
										{#each Object.entries(analytics.statusCounts) as [status, count]}
											<option value={status}>{status} ({count})</option>
										{/each}
									</select>
								</div>

								<div class="filter-group">
									<label for="email-search">Search Email:</label>
									<input 
										id="email-search" 
										type="text" 
										bind:value={searchEmail}
										placeholder="customer@example.com"
									/>
								</div>

								<button class="reset-filters" onclick={resetFilters}>
									Reset Filters
								</button>
							</div>
							<div class="filter-results">
								Showing <strong>{filteredSubs.length}</strong> of <strong>{ghostReport.ghost_subscriptions.length}</strong> subscriptions
							</div>
						</div>

						<!-- FILTERED SUBSCRIPTION CARDS -->
						<div class="subscriptions-grid">
							{#each filteredSubs as ghost}
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
											<span class="detail-value status-badge" 
												class:active={ghost.metadata?.status === 'active'}
												class:unpaid={ghost.metadata?.status === 'unpaid'}
												class:past-due={ghost.metadata?.status === 'past_due'}
											>
												{ghost.metadata?.status || 'unknown'}
											</span>
										</div>
										<div class="detail-row">
											<span class="detail-label">Ghost Product:</span>
											<span class="detail-value highlight-error">{ghost.referenced_by?.ghost_product || 'N/A'}</span>
										</div>
										<div class="detail-row">
											<span class="detail-label">Amount:</span>
											<span class="detail-value revenue-amount">{formatCurrency(ghost.metadata?.unit_amount, ghost.metadata?.currency)}</span>
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

						{#if filteredSubs.length === 0}
							<div class="no-results">
								<p>No subscriptions match the selected filters.</p>
								<button class="btn-reset" onclick={resetFilters}>Reset Filters</button>
							</div>
						{/if}
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

	/* ============================================ */
	/* GHOST SUBSCRIPTIONS ANALYTICS STYLES */
	/* ============================================ */

	.subscriptions-section {
		border: 2px solid #f59e0b;
	}

	.section-header.critical {
		background: linear-gradient(135deg, #fef3c7, #fde68a);
		border-bottom: 2px solid #f59e0b;
	}

	.section-header.critical:hover {
		background: linear-gradient(135deg, #fde68a, #fcd34d);
	}

	.section-title {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		flex: 1;
	}

	.revenue-badge {
		font-size: 0.875rem;
		background: #059669;
		color: white;
		padding: 0.25rem 0.75rem;
		border-radius: 9999px;
		font-weight: 600;
	}

	/* Critical Alert */
	.critical-alert {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
		padding: 1.5rem;
		background: linear-gradient(135deg, #fee2e2, #fecaca);
		border: 2px solid #ef4444;
		border-radius: 0.75rem;
		margin-bottom: 2rem;
	}

	.alert-icon {
		font-size: 2rem;
		animation: pulse 2s ease-in-out infinite;
	}

	@keyframes pulse {
		0%, 100% {
			transform: scale(1);
		}
		50% {
			transform: scale(1.1);
		}
	}

	.alert-content h3 {
		font-size: 1.25rem;
		font-weight: 700;
		color: #991b1b;
		margin: 0 0 0.5rem 0;
	}

	.alert-content p {
		color: #7f1d1d;
		margin: 0;
		line-height: 1.6;
	}

	/* Revenue Summary */
	.revenue-summary {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
		margin-bottom: 2rem;
	}

	.revenue-card {
		padding: 1.5rem;
		border-radius: 0.75rem;
		border: 2px solid;
		transition: all 0.2s;
	}

	.revenue-card:hover {
		transform: translateY(-4px);
		box-shadow: 0 8px 20px rgba(0, 0, 0, 0.15);
	}

	.revenue-card.total {
		background: linear-gradient(135deg, #ede9fe, #ddd6fe);
		border-color: #9333ea;
	}

	.revenue-card.active {
		background: linear-gradient(135deg, #d1fae5, #a7f3d0);
		border-color: #059669;
	}

	.revenue-card.unpaid {
		background: linear-gradient(135deg, #fef3c7, #fde68a);
		border-color: #f59e0b;
	}

	.revenue-card.past-due {
		background: linear-gradient(135deg, #fee2e2, #fecaca);
		border-color: #ef4444;
	}

	.revenue-label {
		font-size: 0.875rem;
		font-weight: 600;
		color: #374151;
		margin-bottom: 0.5rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.revenue-value {
		font-size: 2rem;
		font-weight: 800;
		color: #111827;
		margin-bottom: 0.25rem;
	}

	.revenue-count {
		font-size: 0.875rem;
		color: #6b7280;
	}

	.revenue-mrr {
		font-size: 0.75rem;
		color: #6b7280;
		margin-top: 0.5rem;
		font-weight: 500;
	}

	/* Plan Analytics Table */
	.plan-analytics-section {
		margin: 2rem 0;
		background: white;
		border-radius: 0.75rem;
		padding: 1.5rem;
		border: 1px solid #e5e7eb;
	}

	.plan-analytics-section h3 {
		font-size: 1.25rem;
		font-weight: 700;
		color: #111827;
		margin: 0 0 1rem 0;
	}

	.analytics-table {
		overflow-x: auto;
	}

	.analytics-table table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
	}

	.analytics-table thead {
		background: #f9fafb;
	}

	.analytics-table th {
		padding: 0.75rem 1rem;
		text-align: left;
		font-weight: 600;
		color: #374151;
		border-bottom: 2px solid #e5e7eb;
		white-space: nowrap;
	}

	.analytics-table td {
		padding: 0.75rem 1rem;
		border-bottom: 1px solid #e5e7eb;
	}

	.analytics-table tbody tr:hover {
		background: #f9fafb;
	}

	.analytics-table .plan-id {
		font-family: 'Courier New', monospace;
		font-weight: 600;
		color: #6b7280;
	}

	.analytics-table .count {
		font-weight: 600;
		color: #111827;
		text-align: center;
	}

	.analytics-table .unit-price {
		font-weight: 600;
		color: #374151;
	}

	.analytics-table .billing-type {
		text-align: center;
	}

	.type-badge {
		display: inline-block;
		padding: 0.25rem 0.75rem;
		border-radius: 9999px;
		font-size: 0.75rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.025em;
	}

	.type-badge.mrr {
		background: #dbeafe;
		color: #1e40af;
	}

	.type-badge.quarterly {
		background: #fef3c7;
		color: #92400e;
	}

	.type-badge.arr {
		background: #dcfce7;
		color: #166534;
	}

	.analytics-table .status-count {
		text-align: center;
		font-weight: 600;
	}

	.analytics-table .status-count.active {
		color: #059669;
	}

	.analytics-table .status-count.unpaid {
		color: #f59e0b;
	}

	.analytics-table .status-count.past-due {
		color: #ef4444;
	}

	.analytics-table .revenue {
		font-weight: 600;
		text-align: right;
	}

	.analytics-table .active-revenue {
		color: #059669;
	}

	.analytics-table .risk-revenue {
		color: #ef4444;
	}

	.analytics-table .total-revenue {
		color: #9333ea;
		font-size: 1rem;
	}

	.analytics-table .potential-revenue {
		color: #2563eb;
		font-weight: 700;
		font-size: 1rem;
	}

	.analytics-table .totals-row {
		background: #f3f4f6;
		font-size: 0.9375rem;
	}

	.analytics-table .totals-row td {
		border-top: 2px solid #9333ea;
		border-bottom: 2px solid #9333ea;
		padding: 1rem;
	}

	/* Filters Section */
	.filters-section {
		margin: 2rem 0;
		padding: 1.5rem;
		background: #f9fafb;
		border-radius: 0.75rem;
		border: 1px solid #e5e7eb;
	}

	.filters-section h3 {
		font-size: 1.125rem;
		font-weight: 700;
		color: #111827;
		margin: 0 0 1rem 0;
	}

	.filters {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.filter-group {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.filter-group label {
		font-size: 0.875rem;
		font-weight: 600;
		color: #374151;
	}

	.filter-group select,
	.filter-group input {
		padding: 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 0.5rem;
		font-size: 0.875rem;
		background: white;
		color: #111827;
		transition: all 0.2s;
	}

	.filter-group select:focus,
	.filter-group input:focus {
		outline: none;
		border-color: #9333ea;
		box-shadow: 0 0 0 3px rgba(147, 51, 234, 0.1);
	}

	.reset-filters {
		padding: 0.75rem 1.5rem;
		background: #6b7280;
		color: white;
		border: none;
		border-radius: 0.5rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s;
		align-self: flex-end;
	}

	.reset-filters:hover {
		background: #4b5563;
		transform: translateY(-2px);
	}

	.filter-results {
		font-size: 0.875rem;
		color: #6b7280;
		padding: 0.75rem;
		background: white;
		border-radius: 0.5rem;
		text-align: center;
	}

	.filter-results strong {
		color: #9333ea;
		font-weight: 700;
	}

	/* Subscriptions Grid */
	.subscriptions-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
		gap: 1rem;
		margin-top: 1.5rem;
	}

	.revenue-amount {
		font-weight: 700;
		color: #059669;
		font-size: 1.125rem;
	}

	.status-badge.unpaid {
		background: #fef3c7;
		color: #92400e;
	}

	.status-badge.past-due {
		background: #fee2e2;
		color: #991b1b;
	}

	/* No Results */
	.no-results {
		text-align: center;
		padding: 3rem 2rem;
		background: #f9fafb;
		border-radius: 0.75rem;
		margin-top: 1rem;
	}

	.no-results p {
		color: #6b7280;
		margin: 0 0 1rem 0;
	}

	.btn-reset {
		padding: 0.75rem 1.5rem;
		background: #9333ea;
		color: white;
		border: none;
		border-radius: 0.5rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s;
	}

	.btn-reset:hover {
		background: #7e22ce;
		transform: translateY(-2px);
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

		.revenue-summary {
			grid-template-columns: 1fr;
		}

		.filters {
			grid-template-columns: 1fr;
		}

		.subscriptions-grid {
			grid-template-columns: 1fr;
		}

		.analytics-table {
			font-size: 0.75rem;
		}

		.analytics-table th,
		.analytics-table td {
			padding: 0.5rem;
		}
	}
</style>

