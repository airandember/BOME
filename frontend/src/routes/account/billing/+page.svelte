<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import { subscriptionService, subscriptionUtils, type Invoice } from '$lib/subscription';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let invoices: Invoice[] = [];
	let loading = true;
	let error = '';
	let currentPage = 1;
	let totalPages = 1;
	let totalInvoices = 0;
	let isAuthenticated = false;

	// Subscribe to auth store
	auth.subscribe(state => {
		isAuthenticated = state.isAuthenticated;
	});

	onMount(async () => {
		if (!isAuthenticated) {
			goto('/login');
			return;
		}

		await loadBillingHistory();
	});

	const loadBillingHistory = async (page = 1) => {
		try {
			loading = true;
			const response = await subscriptionService.getBillingHistory(page, 20);
			invoices = response.invoices || [];
			totalInvoices = response.total || 0;
			totalPages = Math.ceil(totalInvoices / 20);
			currentPage = page;
		} catch (err) {
			error = 'Failed to load billing history';
			console.error('Error loading billing history:', err);
		} finally {
			loading = false;
		}
	};

	const handlePageChange = (page: number) => {
		if (page >= 1 && page <= totalPages) {
			loadBillingHistory(page);
		}
	};

	const handleDownloadInvoice = async (invoiceId: string) => {
		try {
			const response = await subscriptionService.downloadInvoice(invoiceId);
			if (response.downloadUrl) {
				window.open(response.downloadUrl, '_blank');
			} else {
				alert('Failed to download invoice');
			}
		} catch (err) {
			alert('Failed to download invoice');
			console.error('Error downloading invoice:', err);
		}
	};

	const getStatusBadge = (status: string) => {
		const statusConfig = {
			paid: { text: 'Paid', class: 'status-paid' },
			open: { text: 'Open', class: 'status-open' },
			void: { text: 'Void', class: 'status-void' },
			uncollectible: { text: 'Uncollectible', class: 'status-uncollectible' }
		};

		const config = statusConfig[status as keyof typeof statusConfig] || { text: status, class: 'status-unknown' };
		
		return `
			<span class="status-badge ${config.class}">
				${config.text}
			</span>
		`;
	};
</script>

<svelte:head>
	<title>Billing History - BOME</title>
	<meta name="description" content="View your billing history and download invoices" />
</svelte:head>

<div class="billing-page">
	<div class="container">
		<header class="page-header">
			<h1>Billing History</h1>
			<p>View and download your invoices and payment history</p>
		</header>

		{#if loading}
			<div class="loading-container">
				<LoadingSpinner />
				<p>Loading billing history...</p>
			</div>
		{:else if error}
			<div class="error-container">
				<p class="error-message">{error}</p>
				<button class="btn btn-primary" on:click={() => loadBillingHistory()}>
					Try Again
				</button>
			</div>
		{:else}
			<div class="billing-content">
				<div class="summary-card">
					<div class="summary-item">
						<span class="label">Total Invoices</span>
						<span class="value">{totalInvoices}</span>
					</div>
					<div class="summary-item">
						<span class="label">Current Page</span>
						<span class="value">{currentPage} of {totalPages}</span>
					</div>
				</div>

				{#if invoices.length === 0}
					<div class="empty-state">
						<div class="empty-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
								<polyline points="14,2 14,8 20,8"></polyline>
								<line x1="16" y1="13" x2="8" y2="13"></line>
								<line x1="16" y1="17" x2="8" y2="17"></line>
								<polyline points="10,9 9,9 8,9"></polyline>
							</svg>
						</div>
						<h3>No invoices found</h3>
						<p>You don't have any billing history yet.</p>
					</div>
				{:else}
					<div class="invoices-table">
						<div class="table-header">
							<div class="header-cell">Invoice</div>
							<div class="header-cell">Date</div>
							<div class="header-cell">Amount</div>
							<div class="header-cell">Status</div>
							<div class="header-cell">Actions</div>
						</div>

						{#each invoices as invoice}
							<div class="table-row">
								<div class="table-cell">
									<div class="invoice-info">
										<span class="invoice-number">#{invoice.id}</span>
										<span class="invoice-period">
											{subscriptionUtils.formatDate(invoice.periodStart)} - {subscriptionUtils.formatDate(invoice.periodEnd)}
										</span>
									</div>
								</div>
								<div class="table-cell">
									{subscriptionUtils.formatDate(invoice.createdAt)}
								</div>
								<div class="table-cell">
									<span class="amount">
										{subscriptionUtils.formatPrice(invoice.amount, invoice.currency)}
									</span>
								</div>
								<div class="table-cell">
									{@html getStatusBadge(invoice.status)}
								</div>
								<div class="table-cell">
									<div class="actions">
										{#if invoice.status === 'paid' && invoice.downloadUrl}
											<button 
												class="btn btn-small btn-outline"
												on:click={() => handleDownloadInvoice(invoice.id)}
											>
												Download
											</button>
										{/if}
									</div>
								</div>
							</div>
						{/each}
					</div>

					{#if totalPages > 1}
						<div class="pagination">
							<button 
								class="btn btn-outline"
								disabled={currentPage === 1}
								on:click={() => handlePageChange(currentPage - 1)}
							>
								Previous
							</button>
							
							<div class="page-numbers">
								{#each Array.from({ length: totalPages }, (_, i) => i + 1) as page}
									<button 
										class="page-number"
										class:active={page === currentPage}
										on:click={() => handlePageChange(page)}
									>
										{page}
									</button>
								{/each}
							</div>

							<button 
								class="btn btn-outline"
								disabled={currentPage === totalPages}
								on:click={() => handlePageChange(currentPage + 1)}
							>
								Next
							</button>
						</div>
					{/if}
				{/if}
			</div>
		{/if}
	</div>
</div>

<style>
	.billing-page {
		min-height: 100vh;
		padding: 2rem 0;
		background: var(--bg-gradient);
	}

	.container {
		max-width: 1200px;
		margin: 0 auto;
		padding: 0 1rem;
	}

	.page-header {
		text-align: center;
		margin-bottom: 3rem;
	}

	.page-header h1 {
		font-size: 2.5rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.page-header p {
		font-size: 1.1rem;
		color: var(--text-secondary);
	}

	.loading-container,
	.error-container {
		text-align: center;
		padding: 3rem 0;
	}

	.error-message {
		color: var(--error-text);
		margin-bottom: 1rem;
	}

	.billing-content {
		background: var(--card-bg);
		border-radius: 20px;
		padding: 2rem;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
	}

	.summary-card {
		display: flex;
		justify-content: space-around;
		background: var(--bg-secondary);
		border-radius: 16px;
		padding: 1.5rem;
		margin-bottom: 2rem;
		border: 1px solid var(--border-color);
	}

	.summary-item {
		text-align: center;
	}

	.summary-item .label {
		display: block;
		font-size: 0.875rem;
		color: var(--text-secondary);
		margin-bottom: 0.5rem;
	}

	.summary-item .value {
		display: block;
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--text-primary);
	}

	.empty-state {
		text-align: center;
		padding: 3rem 0;
	}

	.empty-icon {
		width: 80px;
		height: 80px;
		margin: 0 auto 1.5rem;
		color: var(--text-secondary);
	}

	.empty-icon svg {
		width: 100%;
		height: 100%;
	}

	.empty-state h3 {
		font-size: 1.5rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.empty-state p {
		color: var(--text-secondary);
	}

	.invoices-table {
		margin-bottom: 2rem;
	}

	.table-header {
		display: grid;
		grid-template-columns: 2fr 1fr 1fr 1fr 1fr;
		gap: 1rem;
		padding: 1rem;
		background: var(--bg-secondary);
		border-radius: 12px;
		margin-bottom: 1rem;
		font-weight: 600;
		color: var(--text-primary);
	}

	.table-row {
		display: grid;
		grid-template-columns: 2fr 1fr 1fr 1fr 1fr;
		gap: 1rem;
		padding: 1rem;
		border-bottom: 1px solid var(--border-color);
		align-items: center;
	}

	.table-row:last-child {
		border-bottom: none;
	}

	.table-cell {
		color: var(--text-primary);
	}

	.invoice-info {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.invoice-number {
		font-weight: 600;
		color: var(--text-primary);
	}

	.invoice-period {
		font-size: 0.875rem;
		color: var(--text-secondary);
	}

	.amount {
		font-weight: 600;
		color: var(--primary-color);
	}

	.status-badge {
		display: inline-block;
		padding: 0.25rem 0.75rem;
		border-radius: 12px;
		font-size: 0.875rem;
		font-weight: 600;
	}

	.status-paid {
		background: var(--success-bg);
		color: var(--success-text);
	}

	.status-open {
		background: var(--warning-bg);
		color: var(--warning-text);
	}

	.status-void {
		background: var(--error-bg);
		color: var(--error-text);
	}

	.status-uncollectible {
		background: var(--error-bg);
		color: var(--error-text);
	}

	.status-unknown {
		background: var(--bg-secondary);
		color: var(--text-secondary);
	}

	.actions {
		display: flex;
		gap: 0.5rem;
	}

	.btn-small {
		padding: 0.5rem 1rem;
		font-size: 0.875rem;
	}

	.pagination {
		display: flex;
		justify-content: center;
		align-items: center;
		gap: 1rem;
		margin-top: 2rem;
	}

	.page-numbers {
		display: flex;
		gap: 0.5rem;
	}

	.page-number {
		width: 40px;
		height: 40px;
		border: 1px solid var(--border-color);
		background: var(--card-bg);
		border-radius: 8px;
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--text-primary);
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.page-number:hover {
		background: var(--bg-secondary);
	}

	.page-number.active {
		background: var(--primary-color);
		color: white;
		border-color: var(--primary-color);
	}

	@media (max-width: 768px) {
		.page-header h1 {
			font-size: 2rem;
		}

		.billing-content {
			padding: 1.5rem;
		}

		.summary-card {
			flex-direction: column;
			gap: 1rem;
		}

		.table-header,
		.table-row {
			grid-template-columns: 1fr;
			gap: 0.5rem;
		}

		.header-cell {
			display: none;
		}

		.table-cell {
			display: flex;
			justify-content: space-between;
			align-items: center;
			padding: 0.5rem 0;
		}

		.table-cell::before {
			content: attr(data-label);
			font-weight: 600;
			color: var(--text-secondary);
		}

		.pagination {
			flex-direction: column;
			gap: 1rem;
		}

		.page-numbers {
			order: -1;
		}
	}
</style> 