<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';

	let summary: any = null;
	let loading = true;
	let error = '';
	let statusFilter = 'all';
	let showPaymentModal = false;
	let selectedPayment: any = null;

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
				error = 'Failed to load payments';
			}
		} catch (err) {
			error = 'Failed to load payments';
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
			case 'succeeded':
				return 'var(--success)';
			case 'processing':
				return 'var(--warning)';
			case 'requires_payment_method':
			case 'requires_confirmation':
			case 'requires_action':
				return 'var(--warning)';
			case 'canceled':
				return 'var(--text-muted)';
			case 'requires_capture':
				return 'var(--primary)';
			default:
				return 'var(--text-muted)';
		}
	}

	function getStatusText(status: string): string {
		switch (status.toLowerCase()) {
			case 'succeeded':
				return 'Succeeded';
			case 'processing':
				return 'Processing';
			case 'requires_payment_method':
				return 'Requires Payment Method';
			case 'requires_confirmation':
				return 'Requires Confirmation';
			case 'requires_action':
				return 'Requires Action';
			case 'canceled':
				return 'Canceled';
			case 'requires_capture':
				return 'Requires Capture';
			default:
				return status.charAt(0).toUpperCase() + status.slice(1);
		}
	}

	function getStatusIcon(status: string): string {
		switch (status.toLowerCase()) {
			case 'succeeded':
				return '✅';
			case 'processing':
				return '⏳';
			case 'requires_payment_method':
				return '💳';
			case 'requires_confirmation':
				return '⚠️';
			case 'requires_action':
				return '🔄';
			case 'canceled':
				return '❌';
			case 'requires_capture':
				return '🎯';
			default:
				return '💳';
		}
	}

	$: allPayments = summary?.payment_intents || [];
	$: payments = statusFilter === 'all' ? allPayments : allPayments.filter((payment: any) => payment.Status === statusFilter);
	$: paymentsCount = summary?.payment_intents_count || 0;
	$: succeededPayments = allPayments.filter((payment: any) => payment.Status === 'succeeded');
	$: processingPayments = allPayments.filter((payment: any) => payment.Status === 'processing');
	$: totalRevenue = succeededPayments.reduce((sum: number, payment: any) => sum + payment.Amount, 0);

	// Payment modal functions
	function viewPayment(payment: any) {
		selectedPayment = payment;
		showPaymentModal = true;
	}

	function closePaymentModal() {
		showPaymentModal = false;
		selectedPayment = null;
	}

	function handleModalClick(event: MouseEvent) {
		if (event.target === event.currentTarget) {
			closePaymentModal();
		}
	}

	function refundPayment(payment: any) {
		// TODO: Implement refund functionality
		alert(`Refund functionality for payment ${payment.ID} will be implemented soon. For now process refunds directly from Stripe.`);
	}

	function capturePayment(payment: any) {
		// TODO: Implement capture functionality
		alert(`Capture functionality for payment ${payment.ID} will be implemented soon. For now capture payments directly from Stripe.`);
	}

	function cancelPayment(payment: any) {
		// TODO: Implement cancel functionality
		alert(`Cancel functionality for payment ${payment.ID} will be implemented soon. For now cancel payments directly from Stripe.`);
	}
</script>

{#if loading}
	<div class="loading">
		<div class="spinner"></div>
		<p>Loading payments...</p>
	</div>
{:else if error}
	<div class="error-state">
		<h3>Error Loading Payments</h3>
		<p>{error}</p>
		<button class="btn btn-primary" on:click={fetchSummary}>Retry</button>
	</div>
{:else}
	<div class="payments-container">
		<div class="payments-header">
			<div class="header-content">
				<h1>💳 Payments</h1>
				<p>Manage and track customer payments and transactions</p>
			</div>
			<div class="header-stats">
				<div class="stat-card">
					<span class="stat-value">{paymentsCount}</span>
					<span class="stat-label">Total Payments</span>
				</div>
				<div class="stat-card succeeded">
					<span class="stat-value">{succeededPayments.length}</span>
					<span class="stat-label">Succeeded</span>
				</div>
				<div class="stat-card processing">
					<span class="stat-value">{processingPayments.length}</span>
					<span class="stat-label">Processing</span>
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
					➕ Create Payment
				</button> -->
			</div>
		</div>

		{#if payments.length === 0}
			<div class="empty-state">
				<div class="empty-icon">💳</div>
				<h3>No Payments Found</h3>
				<p>You haven't processed any payments yet. Payments will appear here as customers make purchases.</p>
			</div>
		{:else}
			<div class="payments-table-container">
				<div class="table-header">
					<h2>Recent Payments {statusFilter !== 'all' ? `(${payments.length} ${statusFilter})` : `(${payments.length})`}</h2>
					<div class="table-filters">
						<select class="filter-select" bind:value={statusFilter}>
							<option value="all">All Statuses</option>
							<option value="succeeded">Succeeded</option>
							<option value="processing">Processing</option>
							<option value="requires_payment_method">Requires Payment Method</option>
							<option value="requires_confirmation">Requires Confirmation</option>
							<option value="requires_action">Requires Action</option>
							<option value="requires_capture">Requires Capture</option>
							<option value="canceled">Canceled</option>
						</select>
					</div>
				</div>

				<div class="table-wrapper">
					<table class="payments-table">
						<thead>
							<tr>
								<th>Payment</th>
								<th>Status</th>
								<th>Amount</th>
								<th>Currency</th>
								<th>Created</th>
								<th>Actions</th>
							</tr>
						</thead>
						<tbody>
							{#each payments as payment}
								<tr class="payment-row">
									<td class="payment-id">
										<div class="payment-info">
											<span class="payment-number">#{payment.ID.slice(-8)}</span>
											<span class="payment-full-id">{payment.ID}</span>
										</div>
									</td>
									<td class="payment-status">
										<div class="status-badge" style="color: {getStatusColor(payment.Status)}">
											<span class="status-icon">{getStatusIcon(payment.Status)}</span>
											<span class="status-text">{getStatusText(payment.Status)}</span>
										</div>
									</td>
									<td class="payment-amount">
										<span class="amount-value">{formatCurrency(payment.Amount, payment.Currency)}</span>
									</td>
									<td class="payment-currency">
										<span class="currency-code">{payment.Currency.toUpperCase()}</span>
									</td>
									<td class="payment-date">
										<span class="date-value">{formatDate(payment.CreatedAt)}</span>
									</td>
									<td class="payment-actions">
										<div class="action-buttons">
											<button class="btn btn-sm btn-outline" title="View Payment" on:click={() => viewPayment(payment)}>
												👁️ View
											</button>
											{#if payment.Status === 'succeeded'}
												<button class="btn btn-sm btn-outline" title="Refund Payment" on:click={() => refundPayment(payment)}>
													↩️ Refund
												</button>
											{/if}
											{#if payment.Status === 'requires_capture'}
												<button class="btn btn-sm btn-primary" title="Capture Payment" on:click={() => capturePayment(payment)}>
													🎯 Capture
												</button>
											{/if}
											{#if payment.Status !== 'succeeded' && payment.Status !== 'canceled'}
												<button class="btn btn-sm btn-secondary" title="Cancel Payment" on:click={() => cancelPayment(payment)}>
													❌ Cancel
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

			<!-- Payment Summary Cards -->
			<div class="payment-summary">
				<h2>Payment Summary</h2>
				<div class="summary-grid">
					<div class="summary-card">
						<div class="summary-header">
							<h3>💰 Revenue Breakdown</h3>
						</div>
						<div class="summary-stats">
							<div class="summary-stat">
								<span class="stat-label">Successful Payments</span>
								<span class="stat-value">{formatCurrency(totalRevenue)}</span>
							</div>
							<div class="summary-stat">
								<span class="stat-label">Processing</span>
								<span class="stat-value">{formatCurrency(processingPayments.reduce((sum: number, payment: any) => sum + payment.Amount, 0))}</span>
							</div>
							<div class="summary-stat">
								<span class="stat-label">Average Payment</span>
								<span class="stat-value">{paymentsCount > 0 ? formatCurrency(allPayments.reduce((sum: number, payment: any) => sum + payment.Amount, 0) / paymentsCount) : formatCurrency(0)}</span>
							</div>
						</div>
					</div>

					<div class="summary-card">
						<div class="summary-header">
							<h3>📊 Status Distribution</h3>
						</div>
						<div class="status-distribution">
							{#each ['succeeded', 'processing', 'requires_payment_method', 'requires_confirmation', 'requires_action', 'requires_capture', 'canceled'] as status}
								{@const statusPayments = allPayments.filter((payment: any) => payment.Status === status)}
								{#if statusPayments.length > 0}
									<div class="status-item">
										<div class="status-info">
											<span class="status-icon">{getStatusIcon(status)}</span>
											<span class="status-name">{getStatusText(status)}</span>
										</div>
										<div class="status-count">
											<span class="count">{statusPayments.length}</span>
											<span class="percentage">({Math.round((statusPayments.length / paymentsCount) * 100)}%)</span>
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

<!-- Payment View Modal -->
{#if showPaymentModal && selectedPayment}
	<div class="modal-overlay" on:click={handleModalClick} on:keydown={(e) => e.key === 'Escape' && closePaymentModal()} role="dialog" aria-modal="true" tabindex="-1">
		<div class="modal-content" role="document">
			<div class="modal-header">
				<h3>💳 Payment Details</h3>
				<button class="modal-close" on:click={closePaymentModal}>&times;</button>
			</div>
			
			<div class="modal-body">
				<div class="payment-details-grid">
					<div class="detail-section">
						<h4>Payment Information</h4>
						<div class="detail-rows">
							<div class="detail-row">
								<span class="detail-label">Payment ID:</span>
								<span class="detail-value">{selectedPayment.ID}</span>
							</div>
							<div class="detail-row">
								<span class="detail-label">Short ID:</span>
								<span class="detail-value">#{selectedPayment.ID.slice(-8)}</span>
							</div>
							<div class="detail-row">
								<span class="detail-label">Status:</span>
								<div class="detail-value">
									<div class="status-badge" style="color: {getStatusColor(selectedPayment.Status)}">
										<span class="status-icon">{getStatusIcon(selectedPayment.Status)}</span>
										<span class="status-text">{getStatusText(selectedPayment.Status)}</span>
									</div>
								</div>
							</div>
							<div class="detail-row">
								<span class="detail-label">Created:</span>
								<span class="detail-value">{formatDate(selectedPayment.CreatedAt)}</span>
							</div>
						</div>
					</div>

					<div class="detail-section">
						<h4>Amount & Currency</h4>
						<div class="detail-rows">
							<div class="detail-row">
								<span class="detail-label">Amount:</span>
								<span class="detail-value amount-highlight">{formatCurrency(selectedPayment.Amount, selectedPayment.Currency)}</span>
							</div>
							<div class="detail-row">
								<span class="detail-label">Currency:</span>
								<span class="detail-value">{selectedPayment.Currency.toUpperCase()}</span>
							</div>
							<div class="detail-row">
								<span class="detail-label">Amount (cents):</span>
								<span class="detail-value">{selectedPayment.Amount}</span>
							</div>
						</div>
					</div>

					{#if selectedPayment.Metadata && Object.keys(selectedPayment.Metadata).length > 0}
						<div class="detail-section full-width">
							<h4>Metadata</h4>
							<div class="metadata-grid">
								{#each Object.entries(selectedPayment.Metadata) as [key, value]}
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
				<button class="btn btn-secondary" on:click={closePaymentModal}>
					Close
				</button>
				{#if selectedPayment.Status === 'succeeded'}
					<button class="btn btn-outline" on:click={() => refundPayment(selectedPayment)}>
						↩️ Refund
					</button>
				{/if}
				{#if selectedPayment.Status === 'requires_capture'}
					<button class="btn btn-primary" on:click={() => capturePayment(selectedPayment)}>
						🎯 Capture
					</button>
				{/if}
				{#if selectedPayment.Status !== 'succeeded' && selectedPayment.Status !== 'canceled'}
					<button class="btn btn-secondary" on:click={() => cancelPayment(selectedPayment)}>
						❌ Cancel
					</button>
				{/if}
			</div>
		</div>
	</div>
{/if}

<style>
	.payments-container {
		padding: var(--space-lg);
	}

	.payments-header {
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

	.stat-card.succeeded {
		border-color: var(--success);
		background: var(--success-light);
	}

	.stat-card.processing {
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

	.payments-table-container {
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

	.payments-table {
		width: 100%;
		border-collapse: collapse;
	}

	.payments-table th {
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

	.payments-table td {
		padding: var(--space-md) var(--space-lg);
		border-bottom: 1px solid var(--border);
	}

	.payment-row:hover {
		background: var(--bg-secondary);
	}

	.payment-info {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.payment-number {
		font-weight: 600;
		color: var(--text);
		font-size: 1rem;
	}

	.payment-full-id {
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

	.payment-summary {
		margin-top: var(--space-xl);
	}

	.payment-summary h2 {
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

	.payment-details-grid {
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
		.payments-header {
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

		.payments-table {
			font-size: 0.875rem;
		}

		.payments-table th,
		.payments-table td {
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

		.payment-details-grid {
			grid-template-columns: 1fr;
		}

		.modal-footer {
			flex-direction: column-reverse;
		}
	}
</style>
