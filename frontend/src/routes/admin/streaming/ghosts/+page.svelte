<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/services/api';

	interface GhostCustomer {
		id: number;
		stripe_customer_id: string;
		customer_email: string;
		customer_name: string;
		ghost_type: string;
		ghost_reason: string;
		purge_status: string;
		detection_date: string;
		current_status: string;
		subscription_count: number;
		invoice_count: number;
		notes?: string;
	}

	interface GhostSummary {
		total_ghosts: number;
		hash_id_ghosts: number;
		invalid_format_ghosts: number;
		marked_for_purge: number;
		already_purged: number;
	}

	let ghosts: GhostCustomer[] = [];
	let summary: GhostSummary | null = null;
	let loading = false;
	let error = '';
	let selectedGhosts: Set<string> = new Set();
	let showConfirmDialog = false;
	let actionType: 'mark' | 'purge' | 'bulk' = 'mark';
	let targetCustomerId = '';

	onMount(() => {
		loadGhosts();
		loadSummary();
	});

	async function loadGhosts() {
		loading = true;
		try {
			const response = await apiRequest('/admin/streaming/ghosts/customers');
			ghosts = response.ghosts || [];
		} catch (err) {
			error = 'Failed to load ghost customers';
			console.error(err);
		} finally {
			loading = false;
		}
	}

	async function loadSummary() {
		try {
			summary = await apiRequest('/admin/streaming/ghosts/summary');
		} catch (err) {
			console.error('Failed to load summary:', err);
		}
	}

	async function runDetection() {
		loading = true;
		try {
			const response = await apiRequest('/admin/streaming/ghosts/detect', {
				method: 'POST'
			});
			
			alert(`Ghost detection completed. Found ${response.new_ghosts} new ghosts.`);
			await loadGhosts();
			await loadSummary();
		} catch (err) {
			error = 'Failed to run ghost detection';
			console.error(err);
		} finally {
			loading = false;
		}
	}

	async function markForPurge(customerId: string, reason = 'Admin decision') {
		try {
			await apiRequest(`/admin/streaming/ghosts/customers/${customerId}/mark-purge`, {
				method: 'PUT',
				body: JSON.stringify({ reason })
			});
			
			await loadGhosts();
			alert('Customer marked for purge');
		} catch (err) {
			error = 'Failed to mark customer for purge';
			console.error(err);
		}
	}

	async function purgeCustomer(customerId: string) {
		try {
			await apiRequest(`/admin/streaming/ghosts/customers/${customerId}`, {
				method: 'DELETE',
				body: JSON.stringify({ 
					confirm: true, 
					admin_user: 'admin_dashboard' 
				})
			});
			
			await loadGhosts();
			await loadSummary();
			alert('Customer purged successfully');
		} catch (err) {
			error = 'Failed to purge customer';
			console.error(err);
		}
	}

	async function bulkPurge() {
		if (selectedGhosts.size === 0) {
			alert('Please select customers to purge');
			return;
		}

		try {
			const response = await apiRequest('/admin/streaming/ghosts/bulk-purge', {
				method: 'POST',
				body: JSON.stringify({
					customer_ids: Array.from(selectedGhosts),
					confirm: true,
					admin_user: 'admin_dashboard'
				})
			});
			
			await loadGhosts();
			await loadSummary();
			selectedGhosts.clear();
			alert(`Bulk purge completed. ${response.successful} successful, ${response.failed} failed.`);
		} catch (err) {
			error = 'Failed to bulk purge customers';
			console.error(err);
		}
	}

	function toggleSelection(customerId: string) {
		if (selectedGhosts.has(customerId)) {
			selectedGhosts.delete(customerId);
		} else {
			selectedGhosts.add(customerId);
		}
		selectedGhosts = selectedGhosts; // Trigger reactivity
	}

	function selectAll() {
		selectedGhosts = new Set(ghosts.map(g => g.stripe_customer_id));
	}

	function clearSelection() {
		selectedGhosts.clear();
		selectedGhosts = selectedGhosts;
	}

	function confirmAction(type: 'mark' | 'purge' | 'bulk', customerId = '') {
		actionType = type;
		targetCustomerId = customerId;
		showConfirmDialog = true;
	}

	async function executeAction() {
		showConfirmDialog = false;
		
		switch (actionType) {
			case 'mark':
				await markForPurge(targetCustomerId);
				break;
			case 'purge':
				await purgeCustomer(targetCustomerId);
				break;
			case 'bulk':
				await bulkPurge();
				break;
		}
	}

	function getStatusBadgeClass(status: string) {
		switch (status) {
			case 'detected': return 'badge-warning';
			case 'marked_for_purge': return 'badge-danger';
			case 'purged': return 'badge-success';
			default: return 'badge-secondary';
		}
	}

	function getReasonBadgeClass(reason: string) {
		if (reason.includes('hash_id')) return 'badge-warning';
		if (reason.includes('invalid')) return 'badge-danger';
		if (reason.includes('test')) return 'badge-info';
		return 'badge-secondary';
	}
</script>

<div class="ghost-dashboard">
	<div class="header">
		<h1>👻 Ghost Customer Detection</h1>
		<p>Manage customers that exist locally but not in Stripe</p>
	</div>

	{#if error}
		<div class="alert alert-danger">
			{error}
			<button on:click={() => error = ''}>×</button>
		</div>
	{/if}

	<!-- Summary Cards -->
	{#if summary}
		<div class="summary-cards">
			<div class="card">
				<h3>{summary.total_ghosts}</h3>
				<p>Total Ghosts</p>
			</div>
			<div class="card warning">
				<h3>{summary.hash_id_ghosts}</h3>
				<p>Hash ID Ghosts</p>
			</div>
			<div class="card danger">
				<h3>{summary.invalid_format_ghosts}</h3>
				<p>Invalid Format</p>
			</div>
			<div class="card info">
				<h3>{summary.marked_for_purge}</h3>
				<p>Marked for Purge</p>
			</div>
			<div class="card success">
				<h3>{summary.already_purged}</h3>
				<p>Already Purged</p>
			</div>
		</div>
	{/if}

	<!-- Action Buttons -->
	<div class="actions">
		<button class="btn btn-primary" on:click={runDetection} disabled={loading}>
			🔍 Run Ghost Detection
		</button>
		
		<button class="btn btn-warning" on:click={selectAll} disabled={ghosts.length === 0}>
			Select All
		</button>
		
		<button class="btn btn-secondary" on:click={clearSelection} disabled={selectedGhosts.size === 0}>
			Clear Selection
		</button>
		
		<button 
			class="btn btn-danger" 
			on:click={() => confirmAction('bulk')} 
			disabled={selectedGhosts.size === 0}
		>
			🗑️ Bulk Purge ({selectedGhosts.size})
		</button>
	</div>

	<!-- Ghost Customers Table -->
	<div class="table-container">
		{#if loading}
			<div class="loading">Loading ghost customers...</div>
		{:else if ghosts.length === 0}
			<div class="empty-state">
				<h3>🎉 No Ghost Customers Found!</h3>
				<p>Your database is clean of ghost customers.</p>
			</div>
		{:else}
			<table class="ghosts-table">
				<thead>
					<tr>
						<th>Select</th>
						<th>Stripe ID</th>
						<th>Email</th>
						<th>Name</th>
						<th>Ghost Reason</th>
						<th>Status</th>
						<th>Current</th>
						<th>Related Data</th>
						<th>Detected</th>
						<th>Actions</th>
					</tr>
				</thead>
				<tbody>
					{#each ghosts as ghost}
						<tr class:selected={selectedGhosts.has(ghost.stripe_customer_id)}>
							<td>
								<input 
									type="checkbox" 
									checked={selectedGhosts.has(ghost.stripe_customer_id)}
									on:change={() => toggleSelection(ghost.stripe_customer_id)}
								/>
							</td>
							<td>
								<code class="stripe-id">{ghost.stripe_customer_id}</code>
							</td>
							<td>{ghost.customer_email}</td>
							<td>{ghost.customer_name || 'N/A'}</td>
							<td>
								<span class="badge {getReasonBadgeClass(ghost.ghost_reason)}">
									{ghost.ghost_reason}
								</span>
							</td>
							<td>
								<span class="badge {getStatusBadgeClass(ghost.purge_status)}">
									{ghost.purge_status}
								</span>
							</td>
							<td>
								<span class="badge {ghost.current_status === 'exists' ? 'badge-warning' : 'badge-success'}">
									{ghost.current_status}
								</span>
							</td>
							<td>
								{ghost.subscription_count} subs, {ghost.invoice_count} invoices
							</td>
							<td>
								{new Date(ghost.detection_date).toLocaleDateString()}
							</td>
							<td class="actions-cell">
								{#if ghost.purge_status === 'detected'}
									<button 
										class="btn btn-sm btn-warning"
										on:click={() => confirmAction('mark', ghost.stripe_customer_id)}
									>
										Mark for Purge
									</button>
								{/if}
								
								{#if ghost.current_status === 'exists'}
									<button 
										class="btn btn-sm btn-danger"
										on:click={() => confirmAction('purge', ghost.stripe_customer_id)}
									>
										🗑️ Purge
									</button>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>
</div>

<!-- Confirmation Dialog -->
{#if showConfirmDialog}
	<div class="modal-overlay">
		<div class="modal">
			<h3>Confirm Action</h3>
			<p>
				{#if actionType === 'mark'}
					Mark customer <code>{targetCustomerId}</code> for purging?
				{:else if actionType === 'purge'}
					⚠️ <strong>PERMANENTLY DELETE</strong> customer <code>{targetCustomerId}</code> and all related data?
				{:else if actionType === 'bulk'}
					⚠️ <strong>PERMANENTLY DELETE</strong> {selectedGhosts.size} selected customers and all their related data?
				{/if}
			</p>
			<p class="warning">This action cannot be undone!</p>
			
			<div class="modal-actions">
				<button class="btn btn-secondary" on:click={() => showConfirmDialog = false}>
					Cancel
				</button>
				<button class="btn btn-danger" on:click={executeAction}>
					{actionType === 'mark' ? 'Mark for Purge' : 'DELETE'}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.ghost-dashboard {
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

	.summary-cards {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
		margin-bottom: 2rem;
	}

	.card {
		background: white;
		border: 1px solid #ddd;
		border-radius: 8px;
		padding: 1.5rem;
		text-align: center;
		box-shadow: 0 2px 4px rgba(0,0,0,0.1);
	}

	.card.warning { border-left: 4px solid #ffc107; }
	.card.danger { border-left: 4px solid #dc3545; }
	.card.info { border-left: 4px solid #17a2b8; }
	.card.success { border-left: 4px solid #28a745; }

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

	.actions {
		display: flex;
		gap: 1rem;
		margin-bottom: 2rem;
		flex-wrap: wrap;
	}

	.table-container {
		background: white;
		border-radius: 8px;
		overflow: hidden;
		box-shadow: 0 2px 4px rgba(0,0,0,0.1);
	}

	.ghosts-table {
		width: 100%;
		border-collapse: collapse;
	}

	.ghosts-table th,
	.ghosts-table td {
		padding: 0.75rem;
		text-align: left;
		border-bottom: 1px solid #ddd;
	}

	.ghosts-table th {
		background: #f8f9fa;
		font-weight: 600;
		color: #333;
	}

	.ghosts-table tr:hover {
		background: #f8f9fa;
	}

	.ghosts-table tr.selected {
		background: #e3f2fd;
	}

	.stripe-id {
		background: #f1f3f4;
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-family: monospace;
		font-size: 0.85rem;
	}

	.badge {
		display: inline-block;
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-size: 0.75rem;
		font-weight: 500;
		text-transform: uppercase;
	}

	.badge-warning { background: #fff3cd; color: #856404; }
	.badge-danger { background: #f8d7da; color: #721c24; }
	.badge-success { background: #d4edda; color: #155724; }
	.badge-info { background: #d1ecf1; color: #0c5460; }
	.badge-secondary { background: #e2e3e5; color: #383d41; }

	.actions-cell {
		white-space: nowrap;
	}

	.actions-cell .btn {
		margin-right: 0.5rem;
	}

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

	.btn-sm {
		padding: 0.25rem 0.5rem;
		font-size: 0.75rem;
	}

	.btn-primary { background: #007bff; color: white; }
	.btn-warning { background: #ffc107; color: #212529; }
	.btn-danger { background: #dc3545; color: white; }
	.btn-secondary { background: #6c757d; color: white; }

	.btn:hover:not(:disabled) {
		opacity: 0.9;
		transform: translateY(-1px);
	}

	.loading {
		text-align: center;
		padding: 3rem;
		color: #666;
	}

	.empty-state {
		text-align: center;
		padding: 3rem;
	}

	.empty-state h3 {
		color: #28a745;
		margin-bottom: 0.5rem;
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

	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0,0,0,0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}

	.modal {
		background: white;
		border-radius: 8px;
		padding: 2rem;
		max-width: 500px;
		width: 90%;
		box-shadow: 0 4px 6px rgba(0,0,0,0.1);
	}

	.modal h3 {
		margin-top: 0;
		color: #333;
	}

	.warning {
		color: #dc3545;
		font-weight: 500;
	}

	.modal-actions {
		display: flex;
		gap: 1rem;
		justify-content: flex-end;
		margin-top: 2rem;
	}

	@media (max-width: 768px) {
		.ghost-dashboard {
			padding: 1rem;
		}
		
		.actions {
			flex-direction: column;
		}
		
		.table-container {
			overflow-x: auto;
		}
		
		.ghosts-table {
			min-width: 800px;
		}
	}
</style>
