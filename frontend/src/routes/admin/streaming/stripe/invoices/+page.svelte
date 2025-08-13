<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';

	let summary: any = null;
	let loading = true;
	let error = '';
	let statusFilter = 'all';
	let showInvoiceModal = false;
	let selectedInvoice: any = null;

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
				error = 'Failed to load invoices';
			}
		} catch (err) {
			error = 'Failed to load invoices';
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
		}).format(amount / 100); // Convert from cents
	}

	function formatDate(dateString: string): string {
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
			case 'paid':
				return 'var(--success)';
			case 'open':
				return 'var(--warning)';
			case 'draft':
				return 'var(--text-muted)';
			case 'void':
				return 'var(--error)';
			case 'uncollectible':
				return 'var(--error)';
			default:
				return 'var(--text-muted)';
		}
	}

	function getStatusText(status: string): string {
		switch (status.toLowerCase()) {
			case 'paid':
				return 'Paid';
			case 'open':
				return 'Open';
			case 'draft':
				return 'Draft';
			case 'void':
				return 'Void';
			case 'uncollectible':
				return 'Uncollectible';
			default:
				return status.charAt(0).toUpperCase() + status.slice(1);
		}
	}

	function getStatusIcon(status: string): string {
		switch (status.toLowerCase()) {
			case 'paid':
				return '✅';
			case 'open':
				return '⏳';
			case 'draft':
				return '📝';
			case 'void':
				return '❌';
			case 'uncollectible':
				return '⚠️';
			default:
				return '📄';
		}
	}

	$: allInvoices = summary?.invoices || [];
	$: invoices = statusFilter === 'all' ? allInvoices : allInvoices.filter((inv: any) => inv.Status === statusFilter);
	$: invoicesCount = summary?.invoices_count || 0;
	$: paidInvoices = allInvoices.filter((inv: any) => inv.Status === 'paid');
	$: openInvoices = allInvoices.filter((inv: any) => inv.Status === 'open');
	$: totalRevenue = paidInvoices.reduce((sum: number, inv: any) => sum + inv.Amount, 0);

	// Invoice modal functions
	function viewInvoice(invoice: any) {
		selectedInvoice = invoice;
		showInvoiceModal = true;
	}

	function closeInvoiceModal() {
		showInvoiceModal = false;
		selectedInvoice = null;
	}

	function handleModalClick(event: MouseEvent) {
		if (event.target === event.currentTarget) {
			closeInvoiceModal();
		}
	}

	function downloadInvoice(invoice: any) {
		// TODO: Implement download functionality
		alert(`Download functionality for invoice ${invoice.ID} will be implemented soon. For now download invoices directly from Stripe.`);
	}

	function sendInvoiceReminder(invoice: any) {
		// TODO: Implement send reminder functionality
		alert(`Send reminder functionality for invoice ${invoice.ID} will be implemented soon`);
	}
</script>

{#if loading}
	<div class="loading">
		<div class="spinner"></div>
		<p>Loading invoices...</p>
	</div>
{:else if error}
	<div class="error-state">
		<h3>Error Loading Invoices</h3>
		<p>{error}</p>
		<button class="btn btn-primary" on:click={fetchSummary}>Retry</button>
	</div>
{:else}
	<div class="invoices-container">
		<div class="invoices-header">
			<div class="header-content">
				<h1>📄 Invoices</h1>
				<p>Manage and track customer invoices and billing</p>
			</div>
			<div class="header-stats">
				<div class="stat-card">
					<span class="stat-value">{invoicesCount}</span>
					<span class="stat-label">Total Invoices</span>
				</div>
				<div class="stat-card paid">
					<span class="stat-value">{paidInvoices.length}</span>
					<span class="stat-label">Paid</span>
				</div>
				<div class="stat-card open">
					<span class="stat-value">{openInvoices.length}</span>
					<span class="stat-label">Open</span>
				</div>
				<div class="stat-card revenue">
					<span class="stat-value">{formatCurrency(totalRevenue)}</span>
					<span class="stat-label">Total Revenue</span>
				</div>
			</div>
			
			<div class="header-actions">
				<button class="btn btn-secondary" on:click={fetchSummary}>
					🔄 Refresh
				</button>
				<!-- <button class="btn btn-primary">
					➕ Create Invoice
				</button> -->
			</div>
		</div>

		{#if invoices.length === 0}
			<div class="empty-state">
				<div class="empty-icon">📄</div>
				<h3>No Invoices Found</h3>
				<p>You haven't created any invoices yet. Invoices will appear here as customers are billed.</p>
			</div>
		{:else}
			<div class="invoices-table-container">
				<div class="table-header">
					<h2>Recent Invoices {statusFilter !== 'all' ? `(${invoices.length} ${statusFilter})` : `(${invoices.length})`}</h2>
					<div class="table-filters">
						<select class="filter-select" bind:value={statusFilter}>
							<option value="all">All Statuses</option>
							<option value="paid">Paid</option>
							<option value="open">Open</option>
							<option value="draft">Draft</option>
							<option value="void">Void</option>
						</select>
					</div>
				</div>

				<div class="table-wrapper">
					<table class="invoices-table">
						<thead>
							<tr>
								<th>Invoice</th>
								<th>Status</th>
								<th>Amount</th>
								<th>Currency</th>
								<th>Created</th>
								<th>Actions</th>
							</tr>
						</thead>
						<tbody>
							{#each invoices as invoice}
								<tr class="invoice-row">
									<td class="invoice-id">
										<div class="invoice-info">
											<span class="invoice-number">#{invoice.ID.slice(-8)}</span>
											<span class="invoice-full-id">{invoice.ID}</span>
										</div>
									</td>
									<td class="invoice-status">
										<div class="status-badge" style="color: {getStatusColor(invoice.Status)}">
											<span class="status-icon">{getStatusIcon(invoice.Status)}</span>
											<span class="status-text">{getStatusText(invoice.Status)}</span>
										</div>
									</td>
									<td class="invoice-amount">
										<span class="amount-value">{formatCurrency(invoice.Amount, invoice.Currency)}</span>
									</td>
									<td class="invoice-currency">
										<span class="currency-code">{invoice.Currency.toUpperCase()}</span>
									</td>
									<td class="invoice-date">
										<span class="date-value">{formatDate(invoice.CreatedAt)}</span>
									</td>
									<td class="invoice-actions">
										<div class="action-buttons">
											<button class="btn btn-sm btn-outline" title="View Invoice" on:click={() => viewInvoice(invoice)}>
												👁️ View
											</button>
											<button class="btn btn-sm btn-outline" title="Download PDF" on:click={() => downloadInvoice(invoice)}>
												📥 Download
											</button>
											{#if invoice.Status === 'open'}
												<button class="btn btn-sm btn-primary" title="Send Reminder" on:click={() => sendInvoiceReminder(invoice)}>
													📧 Send
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

			<!-- Invoice Summary Cards -->
			<div class="invoice-summary">
				<h2>Invoice Summary</h2>
				<div class="summary-grid">
					<div class="summary-card">
						<div class="summary-header">
							<h3>💰 Revenue Breakdown</h3>
						</div>
						<div class="summary-stats">
							<div class="summary-stat">
								<span class="stat-label">Paid Invoices</span>
								<span class="stat-value">{formatCurrency(totalRevenue)}</span>
							</div>
							<div class="summary-stat">
								<span class="stat-label">Outstanding</span>
								<span class="stat-value">{formatCurrency(openInvoices.reduce((sum: number, inv: any) => sum + inv.Amount, 0))}</span>
							</div>
							<div class="summary-stat">
								<span class="stat-label">Average Invoice</span>
								<span class="stat-value">{invoicesCount > 0 ? formatCurrency(allInvoices.reduce((sum: number, inv: any) => sum + inv.Amount, 0) / invoicesCount) : formatCurrency(0)}</span>
							</div>
						</div>
					</div>

					<div class="summary-card">
						<div class="summary-header">
							<h3>📊 Status Distribution</h3>
						</div>
						<div class="status-distribution">
							{#each ['paid', 'open', 'draft', 'void'] as status}
								{@const statusInvoices = allInvoices.filter((inv: any) => inv.Status === status)}
								{#if statusInvoices.length > 0}
									<div class="status-item">
										<div class="status-info">
											<span class="status-icon">{getStatusIcon(status)}</span>
											<span class="status-name">{getStatusText(status)}</span>
										</div>
										<div class="status-count">
											<span class="count">{statusInvoices.length}</span>
											<span class="percentage">({Math.round((statusInvoices.length / invoicesCount) * 100)}%)</span>
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

<!-- Invoice View Modal -->
{#if showInvoiceModal && selectedInvoice}
	<div class="modal-overlay" on:click={handleModalClick} on:keydown={(e) => e.key === 'Escape' && closeInvoiceModal()} role="dialog" aria-modal="true" tabindex="-1">
		<div class="modal-content" role="document">
			<div class="modal-header">
				<h3>📄 Invoice Details</h3>
				<button class="modal-close" on:click={closeInvoiceModal}>&times;</button>
			</div>
			
			<div class="modal-body">
				<div class="invoice-details-grid">
					<div class="detail-section">
						<h4>Invoice Information</h4>
						<div class="detail-rows">
							<div class="detail-row">
								<span class="detail-label">Invoice ID:</span>
								<span class="detail-value">{selectedInvoice.ID}</span>
							</div>
							<div class="detail-row">
								<span class="detail-label">Short ID:</span>
								<span class="detail-value">#{selectedInvoice.ID.slice(-8)}</span>
							</div>
							<div class="detail-row">
								<span class="detail-label">Status:</span>
								<div class="detail-value">
									<div class="status-badge" style="color: {getStatusColor(selectedInvoice.Status)}">
										<span class="status-icon">{getStatusIcon(selectedInvoice.Status)}</span>
										<span class="status-text">{getStatusText(selectedInvoice.Status)}</span>
									</div>
								</div>
							</div>
							<div class="detail-row">
								<span class="detail-label">Created:</span>
								<span class="detail-value">{formatDate(selectedInvoice.CreatedAt)}</span>
							</div>
						</div>
					</div>

					<div class="detail-section">
						<h4>Amount & Currency</h4>
						<div class="detail-rows">
							<div class="detail-row">
								<span class="detail-label">Amount:</span>
								<span class="detail-value amount-highlight">{formatCurrency(selectedInvoice.Amount, selectedInvoice.Currency)}</span>
							</div>
							<div class="detail-row">
								<span class="detail-label">Currency:</span>
								<span class="detail-value">{selectedInvoice.Currency.toUpperCase()}</span>
							</div>
							<div class="detail-row">
								<span class="detail-label">Amount (cents):</span>
								<span class="detail-value">{selectedInvoice.Amount}</span>
							</div>
						</div>
					</div>

					{#if selectedInvoice.Metadata && Object.keys(selectedInvoice.Metadata).length > 0}
						<div class="detail-section full-width">
							<h4>Metadata</h4>
							<div class="metadata-grid">
								{#each Object.entries(selectedInvoice.Metadata) as [key, value]}
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
				<button class="btn btn-secondary" on:click={closeInvoiceModal}>
					Close
				</button>
				<button class="btn btn-outline" on:click={() => downloadInvoice(selectedInvoice)}>
					📥 Download PDF
				</button>
				{#if selectedInvoice.Status === 'open'}
					<button class="btn btn-primary" on:click={() => sendInvoiceReminder(selectedInvoice)}>
						📧 Send Reminder
					</button>
				{/if}
			</div>
		</div>
	</div>
{/if}

<style>
	.invoices-container {
		padding: var(--space-lg);
	}

	.invoices-header {
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

	.stat-card.paid {
		border-color: var(--success);
		background: var(--success-light);
	}

	.stat-card.open {
		border-color: var(--warning);
		background: var(--warning-light);
	}

	.stat-card.revenue {
		border-color: var(--primary);
		background: var(--primary-light);
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

	.invoices-table-container {
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

	.invoices-table {
		width: 100%;
		border-collapse: collapse;
	}

	.invoices-table th {
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

	.invoices-table td {
		padding: var(--space-md) var(--space-lg);
		border-bottom: 1px solid var(--border);
	}

	.invoice-row:hover {
		background: var(--bg-secondary);
	}

	.invoice-info {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.invoice-number {
		font-weight: 600;
		color: var(--text);
		font-size: 1rem;
	}

	.invoice-full-id {
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

	.amount-value {
		font-weight: 600;
		color: var(--text);
		font-size: 1rem;
	}

	.currency-code {
		font-family: monospace;
		font-size: 0.875rem;
		color: var(--text-muted);
		text-transform: uppercase;
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

	.invoice-summary {
		margin-top: var(--space-xl);
	}

	.invoice-summary h2 {
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

	.count {
		font-weight: 600;
		color: var(--text);
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

	@media (max-width: 768px) {
		.invoices-header {
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

		.invoices-table {
			font-size: 0.875rem;
		}

		.invoices-table th,
		.invoices-table td {
			padding: var(--space-sm);
		}

		.action-buttons {
			flex-direction: column;
		}

		.summary-grid {
			grid-template-columns: 1fr;
		}
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

	.invoice-details-grid {
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

	.amount-highlight {
		color: var(--primary) !important;
		font-size: 1.1rem !important;
		font-weight: 700 !important;
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
		.modal-content {
			width: 95%;
			max-height: 95vh;
		}

		.invoice-details-grid {
			grid-template-columns: 1fr;
		}

		.modal-footer {
			flex-direction: column-reverse;
		}
	}
</style>
