<script lang="ts">
	import { onMount } from 'svelte';
	import { creatorPayoutService } from '$lib/services/creator-payout-service';
	import type { PresenterStats, PayoutSummary, Presenter, PresenterPayout, PayoutFormula } from '$lib/types/creatorPayout';
	
	// State
	let activeTab = $state<string>('overview');
	let isLoading = $state<boolean>(true);
	let error = $state<string>('');
	
	// Overview data
	let stats = $state<PresenterStats | null>(null);
	let currentMonthSummary = $state<PayoutSummary | null>(null);
	let recentPayouts = $state<PresenterPayout[]>([]);
	
	// Presenters data
	let presenters = $state<Presenter[]>([]);
	let selectedPresenter = $state<Presenter | null>(null);
	
	// Payouts data
	let payouts = $state<PresenterPayout[]>([]);
	let selectedMonth = $state<string>(getCurrentMonth());
	
	// Settings data
	let formulas = $state<PayoutFormula[]>([]);
	let defaultFormula = $state<PayoutFormula | null>(null);
	
	// UI state
	let showCreatePresenterModal = $state<boolean>(false);
	let showCreateFormulaModal = $state<boolean>(false);
	let showGeneratePayoutsModal = $state<boolean>(false);
	
	function getCurrentMonth(): string {
		const now = new Date();
		return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;
	}
	
	function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(amount);
	}
	
	function formatNumber(num: number): string {
		return new Intl.NumberFormat('en-US').format(num);
	}
	
	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}
	
	async function loadOverviewData() {
		isLoading = true;
		error = '';
		
		try {
			// Load presenter stats
			try {
				stats = await creatorPayoutService.getPresenterStats();
			} catch (err) {
				console.log('No presenter stats available yet:', err);
				stats = {
					total_presenters: 0,
					active_presenters: 0,
					verified_presenters: 0,
					total_videos: 0,
					total_views: 0,
					total_earnings: 0,
					total_paid: 0,
					pending_payouts: 0
				};
			}
			
			// Load current month summary
			try {
				currentMonthSummary = await creatorPayoutService.getPayoutSummary(getCurrentMonth() + '-01');
			} catch (err) {
				console.log('No payouts for current month yet');
				currentMonthSummary = null;
			}
			
			// Load recent payouts (last 2 months)
			const lastMonth = new Date();
			lastMonth.setMonth(lastMonth.getMonth() - 1);
			const lastMonthStr = `${lastMonth.getFullYear()}-${String(lastMonth.getMonth() + 1).padStart(2, '0')}-01`;
			
			try {
				const result = await creatorPayoutService.getPayoutsByMonth(lastMonthStr);
				recentPayouts = result.payouts.slice(0, 10);
			} catch (err) {
				console.log('No recent payouts');
				recentPayouts = [];
			}
			
		} catch (err: any) {
			console.warn('Error loading overview data:', err);
			// Don't set error - let the page load with empty data
		} finally {
			isLoading = false;
		}
	}
	
	async function loadPresenters() {
		isLoading = true;
		try {
			const result = await creatorPayoutService.getPresenters();
			presenters = result.presenters;
		} catch (err: any) {
			error = err.message || 'Failed to load presenters';
			console.error('Error loading presenters:', err);
		} finally {
			isLoading = false;
		}
	}
	
	async function loadPayouts() {
		isLoading = true;
		try {
			const result = await creatorPayoutService.getPayoutsByMonth(selectedMonth + '-01');
			payouts = result.payouts;
			
			// Also load summary
			currentMonthSummary = await creatorPayoutService.getPayoutSummary(selectedMonth + '-01');
		} catch (err: any) {
			error = err.message || 'Failed to load payouts';
			payouts = [];
			currentMonthSummary = null;
			console.error('Error loading payouts:', err);
		} finally {
			isLoading = false;
		}
	}
	
	async function loadFormulas() {
		isLoading = true;
		try {
			const result = await creatorPayoutService.getFormulas(false);
			formulas = result.formulas;
			
			// Load default formula
			try {
				defaultFormula = await creatorPayoutService.getDefaultFormula();
			} catch (err) {
				console.log('No default formula set');
			}
		} catch (err: any) {
			error = err.message || 'Failed to load formulas';
			console.error('Error loading formulas:', err);
		} finally {
			isLoading = false;
		}
	}
	
	async function handleTabChange(tab: string) {
		activeTab = tab;
		error = '';
		
		switch (tab) {
			case 'overview':
				await loadOverviewData();
				break;
			case 'presenters':
				await loadPresenters();
				break;
			case 'payouts':
				await loadPayouts();
				break;
			case 'settings':
				await loadFormulas();
				break;
		}
	}
	
	async function handleGeneratePayouts() {
		if (!confirm(`Generate payouts for ${selectedMonth}? This will calculate earnings for all presenters.`)) {
			return;
		}
		
		isLoading = true;
		try {
			const result = await creatorPayoutService.generateMonthlyPayouts(selectedMonth + '-01');
			alert(`✅ Generated ${result.generated_count} payouts totaling ${formatCurrency(result.total_amount)}`);
			await loadPayouts();
		} catch (err: any) {
			alert(`❌ Error: ${err.message}`);
		} finally {
			isLoading = false;
		}
	}
	
	async function handleApprovePayouts(selectedIds: number[]) {
		if (!selectedIds.length) {
			alert('Please select payouts to approve');
			return;
		}
		
		if (!confirm(`Approve ${selectedIds.length} payouts?`)) {
			return;
		}
		
		isLoading = true;
		try {
			await creatorPayoutService.approvePayouts(selectedIds);
			alert(`✅ Approved ${selectedIds.length} payouts`);
			await loadPayouts();
		} catch (err: any) {
			alert(`❌ Error: ${err.message}`);
		} finally {
			isLoading = false;
		}
	}
	
	onMount(() => {
		loadOverviewData();
	});
</script>

<div class="creator-payouts-dashboard">
	<div class="header">
		<h1>💰 Creator Payouts</h1>
		<p>Manage presenter earnings, formulas, and payments</p>
	</div>
	
	<!-- Tabs -->
	<div class="tabs">
		<button 
			class="tab" 
			class:active={activeTab === 'overview'}
			onclick={() => handleTabChange('overview')}
		>
			📊 Overview
		</button>
		<button 
			class="tab" 
			class:active={activeTab === 'presenters'}
			onclick={() => handleTabChange('presenters')}
		>
			👥 Presenters
		</button>
		<button 
			class="tab" 
			class:active={activeTab === 'payouts'}
			onclick={() => handleTabChange('payouts')}
		>
			💰 Payouts
		</button>
		<button 
			class="tab" 
			class:active={activeTab === 'settings'}
			onclick={() => handleTabChange('settings')}
		>
			⚙️ Settings
		</button>
		<button 
			class="tab" 
			class:active={activeTab === 'reports'}
			onclick={() => handleTabChange('reports')}
		>
			📈 Reports
		</button>
	</div>
	
	<!-- Loading / Error -->
	{#if isLoading}
		<div class="loading">
			<div class="spinner"></div>
			<p>Loading...</p>
		</div>
	{:else if error}
		<div class="error">
			<p>❌ {error}</p>
			<button onclick={() => handleTabChange(activeTab)}>Retry</button>
		</div>
	{/if}
	
	<!-- Tab Content -->
	<div class="tab-content">
		{#if activeTab === 'overview' && !isLoading}
			<div class="overview-tab">
				<!-- Stats Cards -->
				<div class="stats-grid">
					<div class="stat-card">
						<div class="stat-icon">👥</div>
						<div class="stat-info">
							<div class="stat-value">{formatNumber(stats?.total_presenters || 0)}</div>
							<div class="stat-label">Total Presenters</div>
							<div class="stat-subtext">{formatNumber(stats?.active_presenters || 0)} active</div>
						</div>
					</div>
					
					<div class="stat-card">
						<div class="stat-icon">🎥</div>
						<div class="stat-info">
							<div class="stat-value">{formatNumber(stats?.total_videos || 0)}</div>
							<div class="stat-label">Total Videos</div>
							<div class="stat-subtext">{formatNumber(stats?.total_views || 0)} views</div>
						</div>
					</div>
					
					<div class="stat-card success">
						<div class="stat-icon">💵</div>
						<div class="stat-info">
							<div class="stat-value">{formatCurrency(stats?.total_earnings || 0)}</div>
							<div class="stat-label">Total Earnings</div>
							<div class="stat-subtext">{formatCurrency(stats?.total_paid || 0)} paid</div>
						</div>
					</div>
					
					<div class="stat-card warning">
						<div class="stat-icon">⏳</div>
						<div class="stat-info">
							<div class="stat-value">{formatCurrency(stats?.pending_payouts || 0)}</div>
							<div class="stat-label">Pending Payouts</div>
							<div class="stat-subtext">Awaiting payment</div>
						</div>
					</div>
				</div>
				
				<!-- Current Month Summary -->
				{#if currentMonthSummary}
					<div class="section">
						<h2>📅 Current Month ({getCurrentMonth()})</h2>
						<div class="month-summary">
							<div class="summary-item">
								<span class="label">Presenters:</span>
								<span class="value">{formatNumber(currentMonthSummary.total_presenters)}</span>
							</div>
							<div class="summary-item">
								<span class="label">Videos:</span>
								<span class="value">{formatNumber(currentMonthSummary.total_videos)}</span>
							</div>
							<div class="summary-item">
								<span class="label">Views:</span>
								<span class="value">{formatNumber(currentMonthSummary.total_views)}</span>
							</div>
							<div class="summary-item">
								<span class="label">Total Amount:</span>
								<span class="value highlight">{formatCurrency(currentMonthSummary.total_amount)}</span>
							</div>
							<div class="summary-item">
								<span class="label">Pending:</span>
								<span class="value">{formatCurrency(currentMonthSummary.pending_amount)}</span>
							</div>
							<div class="summary-item">
								<span class="label">Paid:</span>
								<span class="value success">{formatCurrency(currentMonthSummary.paid_amount)}</span>
							</div>
						</div>
					</div>
				{/if}
				
				<!-- Recent Payouts -->
				{#if recentPayouts.length > 0}
					<div class="section">
						<h2>📋 Recent Payouts</h2>
						<div class="recent-payouts">
							{#each recentPayouts as payout}
								<div class="payout-item">
									<div class="payout-presenter">{payout.presenter_name}</div>
									<div class="payout-month">{formatDate(payout.payout_month)}</div>
									<div class="payout-amount">{formatCurrency(payout.final_amount)}</div>
									<div class="payout-status status-{payout.status}">{payout.status}</div>
								</div>
							{/each}
						</div>
					</div>
				{/if}
			</div>
		{:else if activeTab === 'presenters' && !isLoading}
			<div class="presenters-tab">
				<div class="section-header">
					<h2>👥 Presenters ({formatNumber(presenters.length)})</h2>
					<button class="btn-primary" onclick={() => showCreatePresenterModal = true}>
						➕ Add Presenter
					</button>
				</div>
				
				<div class="presenters-grid">
					{#each presenters as presenter}
						<div class="presenter-card">
							<div class="presenter-header">
								<div class="presenter-name">{presenter.name}</div>
								{#if presenter.verified}
									<span class="badge verified">✓ Verified</span>
								{/if}
								{#if !presenter.is_active}
									<span class="badge inactive">Inactive</span>
								{/if}
							</div>
							<div class="presenter-stats">
								<div class="stat">
									<span class="label">Videos:</span>
									<span class="value">{formatNumber(presenter.total_videos)}</span>
								</div>
								<div class="stat">
									<span class="label">Views:</span>
									<span class="value">{formatNumber(presenter.total_views)}</span>
								</div>
								<div class="stat">
									<span class="label">Earnings:</span>
									<span class="value">{formatCurrency(presenter.total_earnings)}</span>
								</div>
								<div class="stat">
									<span class="label">Paid:</span>
									<span class="value success">{formatCurrency(presenter.lifetime_paid)}</span>
								</div>
							</div>
							<div class="presenter-actions">
								<button class="btn-sm" onclick={() => selectedPresenter = presenter}>View Details</button>
								{#if !presenter.verified}
									<button class="btn-sm success" onclick={async () => {
										await creatorPayoutService.verifyPresenter(presenter.id);
										await loadPresenters();
									}}>Verify</button>
								{/if}
							</div>
						</div>
					{/each}
				</div>
				
				{#if presenters.length === 0}
					<div class="empty-state">
						<p>📭 No presenters yet</p>
						<button class="btn-primary" onclick={() => showCreatePresenterModal = true}>Add Your First Presenter</button>
					</div>
				{/if}
			</div>
		{:else if activeTab === 'payouts' && !isLoading}
			<div class="payouts-tab">
				<div class="section-header">
					<h2>💰 Payouts</h2>
					<div class="controls">
						<input 
							type="month" 
							bind:value={selectedMonth}
							onchange={() => loadPayouts()}
						/>
						<button class="btn-primary" onclick={handleGeneratePayouts}>
							🚀 Generate Payouts
						</button>
					</div>
				</div>
				
				{#if currentMonthSummary}
					<div class="month-summary-bar">
						<div class="summary-stat">
							<span>{formatNumber(currentMonthSummary.total_presenters)} presenters</span>
						</div>
						<div class="summary-stat">
							<span>{formatNumber(currentMonthSummary.total_videos)} videos</span>
						</div>
						<div class="summary-stat">
							<span>{formatNumber(currentMonthSummary.total_views)} views</span>
						</div>
						<div class="summary-stat highlight">
							<span>{formatCurrency(currentMonthSummary.total_amount)} total</span>
						</div>
					</div>
				{/if}
				
				<div class="payouts-table">
					{#if payouts.length > 0}
						<table>
							<thead>
								<tr>
									<th><input type="checkbox" /></th>
									<th>Presenter</th>
									<th>Videos</th>
									<th>Views</th>
									<th>Base</th>
									<th>Bonus</th>
									<th>Final Amount</th>
									<th>Status</th>
									<th>Actions</th>
								</tr>
							</thead>
							<tbody>
								{#each payouts as payout}
									<tr>
										<td><input type="checkbox" value={payout.id} /></td>
										<td class="presenter-cell">{payout.presenter_name}</td>
										<td>{formatNumber(payout.total_videos)}</td>
										<td>{formatNumber(payout.total_views)}</td>
										<td>{formatCurrency(payout.base_amount)}</td>
										<td class="success">{formatCurrency(payout.bonus_amount)}</td>
										<td class="amount">{formatCurrency(payout.final_amount)}</td>
										<td><span class="status status-{payout.status}">{payout.status}</span></td>
										<td class="actions-cell">
											<button class="btn-xs">View</button>
											{#if payout.status === 'pending'}
												<button class="btn-xs success">Approve</button>
											{/if}
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					{:else}
						<div class="empty-state">
							<p>📭 No payouts for {selectedMonth}</p>
							<button class="btn-primary" onclick={handleGeneratePayouts}>Generate Payouts</button>
						</div>
					{/if}
				</div>
			</div>
		{:else if activeTab === 'settings' && !isLoading}
			<div class="settings-tab">
				<h2>⚙️ Payout Formula Settings</h2>
				
				<div class="formulas-section">
					<div class="section-header">
						<h3>Available Formulas</h3>
						<button class="btn-primary" onclick={() => showCreateFormulaModal = true}>➕ Create Custom Formula</button>
					</div>
					
					<div class="formulas-grid">
						{#each formulas as formula}
							<div class="formula-card" class:default={formula.is_default} class:inactive={!formula.is_active}>
								<div class="formula-header">
									<h4>{formula.name}</h4>
									{#if formula.is_default}
										<span class="badge default">⭐ DEFAULT</span>
									{/if}
									{#if !formula.is_active}
										<span class="badge inactive">Inactive</span>
									{/if}
								</div>
								<p class="formula-description">{formula.description}</p>
								<div class="formula-details">
									<div class="detail">
										<span class="label">Type:</span>
										<span class="value">{formula.formula_type}</span>
									</div>
									<div class="detail">
										<span class="label">Base Rate:</span>
										<span class="value">${formula.base_rate.toFixed(6)}</span>
									</div>
									<div class="detail">
										<span class="label">Min Payout:</span>
										<span class="value">{formatCurrency(formula.min_payout)}</span>
									</div>
									{#if formula.subscriber_multiplier > 1}
										<div class="detail bonus">
											<span class="label">Subscriber Bonus:</span>
											<span class="value">{formula.subscriber_multiplier}x</span>
										</div>
									{/if}
								</div>
								<div class="formula-actions">
									{#if !formula.is_default}
										<button class="btn-sm" onclick={async () => {
											await creatorPayoutService.setDefaultFormula(formula.id);
											await loadFormulas();
										}}>Set as Default</button>
									{/if}
									<button class="btn-sm">Edit</button>
								</div>
							</div>
						{/each}
					</div>
				</div>
			</div>
		{:else if activeTab === 'reports' && !isLoading}
			<div class="reports-tab">
				<h2>📈 Reports & Analytics</h2>
				<p class="coming-soon">📊 Detailed reports and analytics coming soon!</p>
				<p>This section will include:</p>
				<ul>
					<li>Monthly earnings trends</li>
					<li>Presenter performance comparison</li>
					<li>Formula effectiveness analysis</li>
					<li>Payment status tracking</li>
					<li>Export capabilities (CSV/PDF)</li>
				</ul>
			</div>
		{/if}
	</div>
</div>

<style>
	.creator-payouts-dashboard {
		padding: 2rem;
		max-width: 1400px;
		margin: 0 auto;
	}
	
	.header {
		margin-bottom: 2rem;
	}
	
	.header h1 {
		font-size: 2rem;
		margin-bottom: 0.5rem;
		color: #1a1a1a;
	}
	
	.header p {
		color: #666;
	}
	
	/* Tabs */
	.tabs {
		display: flex;
		gap: 0.5rem;
		border-bottom: 2px solid #e0e0e0;
		margin-bottom: 2rem;
	}
	
	.tab {
		padding: 0.75rem 1.5rem;
		background: none;
		border: none;
		border-bottom: 3px solid transparent;
		cursor: pointer;
		font-size: 1rem;
		color: #666;
		transition: all 0.3s;
	}
	
	.tab:hover {
		color: #1a1a1a;
		background: #f5f5f5;
	}
	
	.tab.active {
		color: #4CAF50;
		border-bottom-color: #4CAF50;
		font-weight: 600;
	}
	
	/* Loading/Error */
	.loading, .error {
		text-align: center;
		padding: 3rem;
	}
	
	.spinner {
		border: 3px solid #f3f3f3;
		border-top: 3px solid #4CAF50;
		border-radius: 50%;
		width: 40px;
		height: 40px;
		animation: spin 1s linear infinite;
		margin: 0 auto 1rem;
	}
	
	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}
	
	/* Stats Grid */
	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 1.5rem;
		margin-bottom: 2rem;
	}
	
	.stat-card {
		background: white;
		border-radius: 12px;
		padding: 1.5rem;
		box-shadow: 0 2px 8px rgba(0,0,0,0.1);
		display: flex;
		align-items: center;
		gap: 1rem;
	}
	
	.stat-card.success {
		border-left: 4px solid #4CAF50;
	}
	
	.stat-card.warning {
		border-left: 4px solid #FF9800;
	}
	
	.stat-icon {
		font-size: 2.5rem;
	}
	
	.stat-value {
		font-size: 1.75rem;
		font-weight: 700;
		color: #1a1a1a;
	}
	
	.stat-label {
		font-size: 0.875rem;
		color: #666;
		margin-top: 0.25rem;
	}
	
	.stat-subtext {
		font-size: 0.75rem;
		color: #999;
		margin-top: 0.25rem;
	}
	
	/* Sections */
	.section {
		background: white;
		border-radius: 12px;
		padding: 1.5rem;
		margin-bottom: 1.5rem;
		box-shadow: 0 2px 8px rgba(0,0,0,0.1);
	}
	
	.section h2 {
		font-size: 1.25rem;
		margin-bottom: 1rem;
	}
	
	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1.5rem;
	}
	
	/* Buttons */
	.btn-primary {
		background: #4CAF50;
		color: white;
		border: none;
		padding: 0.75rem 1.5rem;
		border-radius: 8px;
		cursor: pointer;
		font-size: 1rem;
		font-weight: 600;
		transition: all 0.3s;
	}
	
	.btn-primary:hover {
		background: #45a049;
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(76, 175, 80, 0.3);
	}
	
	.btn-sm {
		padding: 0.5rem 1rem;
		font-size: 0.875rem;
		border: 1px solid #ddd;
		background: white;
		border-radius: 6px;
		cursor: pointer;
		transition: all 0.2s;
	}
	
	.btn-sm:hover {
		background: #f5f5f5;
		border-color: #4CAF50;
	}
	
	.btn-sm.success {
		background: #4CAF50;
		color: white;
		border-color: #4CAF50;
	}
	
	.btn-xs {
		padding: 0.25rem 0.5rem;
		font-size: 0.75rem;
		border: 1px solid #ddd;
		background: white;
		border-radius: 4px;
		cursor: pointer;
	}
	
	/* Presenters Grid */
	.presenters-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: 1.5rem;
	}
	
	.presenter-card {
		background: white;
		border-radius: 12px;
		padding: 1.5rem;
		box-shadow: 0 2px 8px rgba(0,0,0,0.1);
		transition: transform 0.3s, box-shadow 0.3s;
	}
	
	.presenter-card:hover {
		transform: translateY(-4px);
		box-shadow: 0 4px 16px rgba(0,0,0,0.15);
	}
	
	.presenter-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
	}
	
	.presenter-name {
		font-size: 1.125rem;
		font-weight: 600;
	}
	
	.presenter-stats {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 0.75rem;
		margin-bottom: 1rem;
	}
	
	.presenter-actions {
		display: flex;
		gap: 0.5rem;
	}
	
	/* Badges */
	.badge {
		padding: 0.25rem 0.75rem;
		border-radius: 12px;
		font-size: 0.75rem;
		font-weight: 600;
	}
	
	.badge.verified {
		background: #4CAF50;
		color: white;
	}
	
	.badge.inactive {
		background: #ddd;
		color: #666;
	}
	
	.badge.default {
		background: #FF9800;
		color: white;
	}
	
	/* Status badges */
	.status {
		padding: 0.25rem 0.75rem;
		border-radius: 12px;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
	}
	
	.status-pending { background: #FFF3E0; color: #F57C00; }
	.status-approved { background: #E8F5E9; color: #388E3C; }
	.status-paid { background: #4CAF50; color: white; }
	.status-failed { background: #FFEBEE; color: #C62828; }
	
	/* Table */
	table {
		width: 100%;
		border-collapse: collapse;
	}
	
	th, td {
		padding: 0.75rem;
		text-align: left;
		border-bottom: 1px solid #eee;
	}
	
	th {
		font-weight: 600;
		background: #f5f5f5;
	}
	
	.amount {
		font-weight: 700;
		color: #4CAF50;
	}
	
	.success {
		color: #4CAF50;
	}
	
	/* Empty state */
	.empty-state {
		text-align: center;
		padding: 3rem;
		color: #666;
	}
	
	/* Formulas */
	.formulas-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
		gap: 1.5rem;
	}
	
	.formula-card {
		background: white;
		border-radius: 12px;
		padding: 1.5rem;
		box-shadow: 0 2px 8px rgba(0,0,0,0.1);
		border: 2px solid transparent;
	}
	
	.formula-card.default {
		border-color: #FF9800;
	}
	
	.formula-card.inactive {
		opacity: 0.6;
	}
	
	.formula-description {
		color: #666;
		font-size: 0.875rem;
		margin: 0.5rem 0 1rem;
	}
	
	.formula-details {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		margin-bottom: 1rem;
	}
	
	.detail {
		display: flex;
		justify-content: space-between;
		font-size: 0.875rem;
	}
	
	.detail.bonus {
		color: #4CAF50;
		font-weight: 600;
	}
	
	.coming-soon {
		font-size: 1.25rem;
		color: #666;
		margin: 2rem 0;
	}
	
	.reports-tab ul {
		list-style: none;
		padding: 0;
	}
	
	.reports-tab li {
		padding: 0.5rem 0;
		color: #666;
	}
	
	.reports-tab li::before {
		content: "📌 ";
	}
</style>

