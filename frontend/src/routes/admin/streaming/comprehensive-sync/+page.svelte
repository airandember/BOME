<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';

	interface CustomerSyncResult {
		user_id: number;
		email: string;
		old_stripe_id: string;
		new_stripe_id: string;
		sync_status: string;
		plan_name: string;
		subscription_id: string;
		error?: string;
	}

	interface ComprehensiveSyncResult {
		total_users: number;
		ghost_customers: number;
		real_stripe_customers: number;
		newly_linked: number;
		fixed_plans: number;
		errors: number;
		processing_time_ms: number;
		customer_results: CustomerSyncResult[];
	}

	let syncResult: ComprehensiveSyncResult | null = null;
	let isRunning = false;
	let error = '';
	let showDetails = false;
	let filterStatus = 'all';

	// Filtered results based on status
	$: filteredResults = syncResult?.customer_results.filter(result => {
		if (filterStatus === 'all') return true;
		return result.sync_status === filterStatus;
	}) || [];

	// Status counts for filter buttons
	$: statusCounts = syncResult?.customer_results.reduce((counts, result) => {
		counts[result.sync_status] = (counts[result.sync_status] || 0) + 1;
		return counts;
	}, {} as Record<string, number>) || {};

	async function runComprehensiveSync() {
		if (isRunning) return;
		
		isRunning = true;
		error = '';
		syncResult = null;

		try {
			console.log('🚀 Starting comprehensive Stripe sync...');
			
			const response = await apiRequest('/admin/streaming/comprehensive-sync/run', {
				method: 'POST'
			});

			console.log('✅ Comprehensive sync completed:', response);
			syncResult = response.result;
			
		} catch (err) {
			console.error('❌ Comprehensive sync failed:', err);
			error = err instanceof Error ? err.message : 'Unknown error occurred';
		} finally {
			isRunning = false;
		}
	}

	function getStatusBadgeClass(status: string): string {
		switch (status) {
			case 'verified': return 'badge-success';
			case 'newly_linked': return 'badge-info';
			case 'legacy_local_only': return 'badge-warning';
			case 'ghost_for_cleanup': return 'badge-danger';
			case 'error': return 'badge-danger';
			case 'no_change': return 'badge-secondary';
			case 'plan_fixed': return 'badge-success';
			default: return 'badge-secondary';
		}
	}

	function getStatusLabel(status: string): string {
		switch (status) {
			case 'verified': return 'Verified';
			case 'newly_linked': return 'Newly Linked';
			case 'legacy_local_only': return 'Legacy Only';
			case 'ghost_for_cleanup': return 'Ghost (Cleanup)';
			case 'error': return 'Error';
			case 'no_change': return 'No Change';
			case 'plan_fixed': return 'Plan Fixed';
			case 'stripe_customer_not_found': return 'Stripe Not Found';
			case 'invalid_id_cleared': return 'Invalid ID Cleared';
			case 'no_stripe_data': return 'No Stripe Data';
			default: return status;
		}
	}

	function formatProcessingTime(ms: number): string {
		if (ms < 1000) return `${ms}ms`;
		if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
		return `${(ms / 60000).toFixed(1)}m`;
	}
</script>

<div class="comprehensive-sync">
	<div class="header">
		<h1>🔧 Comprehensive Stripe Sync</h1>
		<p>Fix all customer IDs, plans, and Stripe data synchronization</p>
	</div>

	{#if error}
		<div class="alert alert-danger">
			<strong>Error:</strong> {error}
			<button on:click={() => error = ''}>×</button>
		</div>
	{/if}

	<!-- Action Section -->
	<div class="action-section">
		<button 
			class="btn btn-primary btn-lg" 
			on:click={runComprehensiveSync} 
			disabled={isRunning}
		>
			{#if isRunning}
				🔄 Running Comprehensive Sync...
			{:else}
				🚀 Run Comprehensive Sync
			{/if}
		</button>
		
		{#if isRunning}
			<div class="progress-indicator">
				<div class="spinner"></div>
				<p>Processing all users and fixing customer data...</p>
			</div>
		{/if}
	</div>

	<!-- Results Section -->
	{#if syncResult}
		<div class="results-section">
			<h2>🎯 Sync Results</h2>
			
			<!-- Summary Cards -->
			<div class="summary-cards">
				<div class="card">
					<h3>{syncResult.total_users}</h3>
					<p>Total Users</p>
				</div>
				<div class="card ghost">
					<h3>{syncResult.ghost_customers}</h3>
					<p>Ghost Customers</p>
				</div>
				<div class="card success">
					<h3>{syncResult.real_stripe_customers}</h3>
					<p>Real Stripe</p>
				</div>
				<div class="card info">
					<h3>{syncResult.newly_linked}</h3>
					<p>Newly Linked</p>
				</div>
				<div class="card warning">
					<h3>{syncResult.fixed_plans}</h3>
					<p>Plans Fixed</p>
				</div>
				<div class="card danger">
					<h3>{syncResult.errors}</h3>
					<p>Errors</p>
				</div>
				<div class="card time">
					<h3>{formatProcessingTime(syncResult.processing_time_ms)}</h3>
					<p>Processing Time</p>
				</div>
			</div>

			<!-- Details Toggle -->
			<div class="details-toggle">
				<button 
					class="btn btn-secondary" 
					on:click={() => showDetails = !showDetails}
				>
					{showDetails ? '📈 Hide Details' : '📊 Show Details'}
				</button>
			</div>

			<!-- Detailed Results -->
			{#if showDetails && syncResult.customer_results.length > 0}
				<div class="details-section">
					<h3>📋 Customer Sync Details</h3>
					
					<!-- Status Filters -->
					<div class="status-filters">
						<button 
							class="filter-btn {filterStatus === 'all' ? 'active' : ''}"
							on:click={() => filterStatus = 'all'}
						>
							All ({syncResult.customer_results.length})
						</button>
						
						{#each Object.entries(statusCounts) as [status, count]}
							<button 
								class="filter-btn {filterStatus === status ? 'active' : ''}"
								on:click={() => filterStatus = status}
							>
								{getStatusLabel(status)} ({count})
							</button>
						{/each}
					</div>

					<!-- Results Table -->
					<div class="table-container">
						<table class="results-table">
							<thead>
								<tr>
									<th>User ID</th>
									<th>Email</th>
									<th>Old Stripe ID</th>
									<th>New Stripe ID</th>
									<th>Status</th>
									<th>Plan</th>
									<th>Subscription</th>
									<th>Error</th>
								</tr>
							</thead>
							<tbody>
								{#each filteredResults as result}
									<tr>
										<td>{result.user_id}</td>
										<td class="email">{result.email}</td>
										<td>
											{#if result.old_stripe_id}
												<code class="old-id">{result.old_stripe_id}</code>
											{:else}
												<span class="no-data">None</span>
											{/if}
										</td>
										<td>
											{#if result.new_stripe_id}
												<code class="new-id">{result.new_stripe_id}</code>
											{:else}
												<span class="no-data">None</span>
											{/if}
										</td>
										<td>
											<span class="badge {getStatusBadgeClass(result.sync_status)}">
												{getStatusLabel(result.sync_status)}
											</span>
										</td>
										<td>
											{#if result.plan_name}
												<span class="plan-name">{result.plan_name}</span>
											{:else}
												<span class="no-data">No Plan</span>
											{/if}
										</td>
										<td>
											{#if result.subscription_id}
												<code class="subscription-id">{result.subscription_id}</code>
											{:else}
												<span class="no-data">None</span>
											{/if}
										</td>
										<td>
											{#if result.error}
												<span class="error-message" title={result.error}>
													{result.error.length > 50 ? result.error.substring(0, 50) + '...' : result.error}
												</span>
											{:else}
												<span class="no-data">None</span>
											{/if}
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	.comprehensive-sync {
		padding: 2rem;
		max-width: 1400px;
		margin: 0 auto;
	}

	.header {
		text-align: center;
		margin-bottom: 2rem;
	}

	.header h1 {
		color: #333;
		margin-bottom: 0.5rem;
	}

	.action-section {
		text-align: center;
		margin-bottom: 3rem;
		padding: 2rem;
		background: #f8f9fa;
		border-radius: 8px;
	}

	.btn-lg {
		padding: 1rem 2rem;
		font-size: 1.1rem;
		font-weight: 600;
	}

	.progress-indicator {
		margin-top: 2rem;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
	}

	.spinner {
		width: 40px;
		height: 40px;
		border: 4px solid #f3f3f3;
		border-top: 4px solid #007bff;
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.results-section {
		background: white;
		border-radius: 8px;
		padding: 2rem;
		box-shadow: 0 2px 4px rgba(0,0,0,0.1);
	}

	.summary-cards {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		gap: 1rem;
		margin-bottom: 2rem;
	}

	.card {
		background: white;
		border: 1px solid #ddd;
		border-radius: 8px;
		padding: 1.5rem;
		text-align: center;
		box-shadow: 0 2px 4px rgba(0,0,0,0.05);
	}

	.card.ghost { border-left: 4px solid #6f42c1; }
	.card.success { border-left: 4px solid #28a745; }
	.card.info { border-left: 4px solid #17a2b8; }
	.card.warning { border-left: 4px solid #ffc107; }
	.card.danger { border-left: 4px solid #dc3545; }
	.card.time { border-left: 4px solid #6c757d; }

	.card h3 {
		font-size: 2rem;
		margin: 0 0 0.5rem 0;
		color: #333;
	}

	.card p {
		margin: 0;
		color: #666;
		font-size: 0.9rem;
	}

	.details-toggle {
		text-align: center;
		margin-bottom: 2rem;
	}

	.status-filters {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		margin-bottom: 1rem;
	}

	.filter-btn {
		padding: 0.5rem 1rem;
		border: 1px solid #ddd;
		background: white;
		border-radius: 4px;
		cursor: pointer;
		font-size: 0.875rem;
		transition: all 0.2s;
	}

	.filter-btn:hover {
		background: #f8f9fa;
	}

	.filter-btn.active {
		background: #007bff;
		color: white;
		border-color: #007bff;
	}

	.table-container {
		overflow-x: auto;
		border-radius: 8px;
		border: 1px solid #ddd;
	}

	.results-table {
		width: 100%;
		border-collapse: collapse;
		min-width: 1000px;
	}

	.results-table th,
	.results-table td {
		padding: 0.75rem;
		text-align: left;
		border-bottom: 1px solid #ddd;
		font-size: 0.875rem;
	}

	.results-table th {
		background: #f8f9fa;
		font-weight: 600;
		color: #333;
	}

	.results-table tr:hover {
		background: #f8f9fa;
	}

	.email {
		max-width: 200px;
		word-break: break-all;
	}

	.old-id, .new-id, .subscription-id {
		background: #f1f3f4;
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-family: monospace;
		font-size: 0.8rem;
	}

	.old-id {
		color: #dc3545;
	}

	.new-id {
		color: #28a745;
	}

	.subscription-id {
		color: #17a2b8;
	}

	.plan-name {
		font-weight: 500;
		color: #495057;
	}

	.error-message {
		color: #dc3545;
		font-size: 0.8rem;
		cursor: help;
	}

	.no-data {
		color: #6c757d;
		font-style: italic;
		font-size: 0.8rem;
	}

	.badge {
		display: inline-block;
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-size: 0.75rem;
		font-weight: 500;
		text-transform: uppercase;
	}

	.badge-success { background: #d4edda; color: #155724; }
	.badge-info { background: #d1ecf1; color: #0c5460; }
	.badge-warning { background: #fff3cd; color: #856404; }
	.badge-danger { background: #f8d7da; color: #721c24; }
	.badge-secondary { background: #e2e3e5; color: #383d41; }

	.btn {
		padding: 0.5rem 1rem;
		border: none;
		border-radius: 4px;
		cursor: pointer;
		font-size: 0.875rem;
		font-weight: 500;
		text-decoration: none;
		display: inline-block;
		transition: all 0.2s;
	}

	.btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-primary { background: #007bff; color: white; }
	.btn-secondary { background: #6c757d; color: white; }

	.btn:hover:not(:disabled) {
		opacity: 0.9;
		transform: translateY(-1px);
	}

	.alert {
		padding: 1rem;
		border-radius: 4px;
		margin-bottom: 1rem;
		position: relative;
	}

	.alert-danger {
		background: #f8d7da;
		color: #721c24;
		border: 1px solid #f5c6cb;
	}

	.alert button {
		position: absolute;
		right: 1rem;
		top: 0.5rem;
		background: none;
		border: none;
		font-size: 1.5rem;
		cursor: pointer;
		color: inherit;
	}

	@media (max-width: 768px) {
		.comprehensive-sync {
			padding: 1rem;
		}
		
		.summary-cards {
			grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
		}
		
		.status-filters {
			flex-direction: column;
		}
		
		.results-table {
			font-size: 0.8rem;
		}
		
		.results-table th,
		.results-table td {
			padding: 0.5rem;
		}
	}
</style>
