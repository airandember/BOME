<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	interface AdminInvoice {
		id: string;
		amount: number;
		currency: string;
		status: 'paid' | 'open' | 'void' | 'uncollectible';
		createdAt: string;
		dueDate: string;
		periodStart: string;
		periodEnd: string;
		downloadUrl?: string;
		customer: {
			id: string;
			name: string;
			email: string;
		};
		subscription: {
			plan: string;
			interval: string;
		};
	}

	let invoices: AdminInvoice[] = [];
	let loading = true;
	let error = '';
	let selectedStatus = 'all';
	let searchQuery = '';
	let currentPage = 1;
	let totalPages = 1;

	const statusOptions = [
		{ value: 'all', label: 'All Invoices' },
		{ value: 'paid', label: 'Paid' },
		{ value: 'open', label: 'Open' },
		{ value: 'void', label: 'Void' },
		{ value: 'uncollectible', label: 'Uncollectible' }
	];

	onMount(async () => {
		await loadInvoices();
	});

	const loadInvoices = async (page = 1) => {
		try {
			loading = true;
			error = '';
			
			// Mock data
			await new Promise(resolve => setTimeout(resolve, 1000));
			
			const mockInvoices: AdminInvoice[] = [
				{
					id: 'in_1234567890',
					amount: 2999,
					currency: 'usd',
					status: 'paid',
					createdAt: '2024-06-18T10:30:00Z',
					dueDate: '2024-07-18T10:30:00Z',
					periodStart: '2024-06-18T10:30:00Z',
					periodEnd: '2024-07-18T10:30:00Z',
					downloadUrl: 'https://example.com/invoice.pdf',
					customer: {
						id: 'cus_1',
						name: 'John Smith',
						email: 'john.smith@example.com'
					},
					subscription: {
						plan: 'Premium',
						interval: 'monthly'
					}
				},
				{
					id: 'in_0987654321',
					amount: 1999,
					currency: 'usd',
					status: 'open',
					createdAt: '2024-06-17T14:20:00Z',
					dueDate: '2024-07-17T14:20:00Z',
					periodStart: '2024-06-17T14:20:00Z',
					periodEnd: '2024-07-17T14:20:00Z',
					customer: {
						id: 'cus_2',
						name: 'Sarah Johnson',
						email: 'sarah.johnson@example.com'
					},
					subscription: {
						plan: 'Basic',
						interval: 'monthly'
					}
				}
			];

			// Filter by status and search
			let filteredInvoices = mockInvoices;
			if (selectedStatus !== 'all') {
				filteredInvoices = filteredInvoices.filter(inv => inv.status === selectedStatus);
			}
			if (searchQuery.trim()) {
				const query = searchQuery.toLowerCase();
				filteredInvoices = filteredInvoices.filter(inv => 
					inv.customer.name.toLowerCase().includes(query) ||
					inv.customer.email.toLowerCase().includes(query) ||
					inv.id.toLowerCase().includes(query)
				);
			}

			invoices = filteredInvoices;
			totalPages = Math.ceil(filteredInvoices.length / 20);
			currentPage = page;
		} catch (err) {
			error = 'Failed to load invoices';
			console.error('Error loading invoices:', err);
		} finally {
			loading = false;
		}
	};

	const formatCurrency = (amount: number, currency: string = 'USD'): string => {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: currency.toUpperCase()
		}).format(amount / 100);
	};

	const formatDate = (dateString: string): string => {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	};

	const getStatusBadge = (status: string) => {
		const statusConfig = {
			paid: { text: 'Paid', class: 'status-paid' },
			open: { text: 'Open', class: 'status-open' },
			void: { text: 'Void', class: 'status-void' },
			uncollectible: { text: 'Uncollectible', class: 'status-uncollectible' }
		};

		const config = statusConfig[status as keyof typeof statusConfig] || { text: status, class: 'status-unknown' };
		
		return `<span class="status-badge ${config.class}">${config.text}</span>`;
	};

	const handleDownload = async (invoice: AdminInvoice) => {
		if (invoice.downloadUrl) {
			window.open(invoice.downloadUrl, '_blank');
		} else {
			showToast('Download not available', 'warning');
		}
	};
</script>

<svelte:head>
	<title>Invoice Management - Admin Dashboard</title>
</svelte:head>

<div class="invoices-admin-page">
	<div class="page-header">
		<div class="header-content">
			<div class="header-text">
				<h1>Invoice Management</h1>
				<p>View and manage customer invoices</p>
			</div>
			
			<div class="header-actions">
				<button class="btn btn-primary" on:click={() => goto('/admin/financial')}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M19 12H5"></path>
						<path d="M12 19l-7-7 7-7"></path>
					</svg>
					Back to Financial
				</button>
			</div>
		</div>
	</div>

	<!-- Filters -->
	<div class="filters-section glass">
		<div class="filters-left">
			<div class="status-filter">
				<label for="status-select">Status:</label>
				<select id="status-select" bind:value={selectedStatus} on:change={() => loadInvoices(1)}>
					{#each statusOptions as option}
						<option value={option.value}>{option.label}</option>
					{/each}
				</select>
			</div>

			<div class="search-filter">
				<input
					type="text"
					placeholder="Search by customer name, email, or invoice ID..."
					bind:value={searchQuery}
					on:input={() => loadInvoices(1)}
					class="search-input"
				/>
			</div>
		</div>
	</div>

	{#if loading}
		<div class="loading-container">
			<LoadingSpinner size="large" color="primary" />
			<p>Loading invoices...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<p class="error-message">{error}</p>
			<button class="btn btn-primary" on:click={() => loadInvoices()}>Try Again</button>
		</div>
	{:else}
		<!-- Invoices Table -->
		<div class="invoices-table glass">
			<div class="table-header">
				<div class="header-cell">Invoice ID</div>
				<div class="header-cell">Customer</div>
				<div class="header-cell">Plan</div>
				<div class="header-cell">Amount</div>
				<div class="header-cell">Status</div>
				<div class="header-cell">Created</div>
				<div class="header-cell">Due Date</div>
				<div class="header-cell">Actions</div>
			</div>

			{#each invoices as invoice}
				<div class="table-row">
					<div class="table-cell">
						<span class="invoice-id">#{invoice.id}</span>
					</div>
					<div class="table-cell">
						<div class="customer-info">
							<span class="customer-name">{invoice.customer.name}</span>
							<span class="customer-email">{invoice.customer.email}</span>
						</div>
					</div>
					<div class="table-cell">
						<div class="plan-info">
							<span class="plan-name">{invoice.subscription.plan}</span>
							<span class="plan-interval">{invoice.subscription.interval}</span>
						</div>
					</div>
					<div class="table-cell">
						<span class="amount">{formatCurrency(invoice.amount, invoice.currency)}</span>
					</div>
					<div class="table-cell">
						{@html getStatusBadge(invoice.status)}
					</div>
					<div class="table-cell">
						<span class="date">{formatDate(invoice.createdAt)}</span>
					</div>
					<div class="table-cell">
						<span class="date">{formatDate(invoice.dueDate)}</span>
					</div>
					<div class="table-cell">
						<div class="actions">
							{#if invoice.downloadUrl}
								<button class="btn btn-ghost btn-small" on:click={() => handleDownload(invoice)}>
									Download
								</button>
							{/if}
						</div>
					</div>
				</div>
			{/each}

			{#if invoices.length === 0}
				<div class="empty-state">
					<div class="empty-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
							<polyline points="14,2 14,8 20,8"></polyline>
							<line x1="16" y1="13" x2="8" y2="13"></line>
							<line x1="16" y1="17" x2="8" y2="17"></line>
						</svg>
					</div>
					<h3>No invoices found</h3>
					<p>No invoices match your current filters.</p>
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	.invoices-admin-page {
		padding: var(--space-xl);
		background: var(--bg-secondary);
		min-height: 100vh;
	}

	.page-header {
		margin-bottom: var(--space-2xl);
	}

	.header-content {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: var(--space-xl);
		flex-wrap: wrap;
	}

	.header-text h1 {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: var(--space-sm);
	}

	.header-text p {
		color: var(--text-secondary);
		font-size: var(--text-lg);
		margin: 0;
	}

	.filters-section {
		padding: var(--space-lg);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
		margin-bottom: var(--space-xl);
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: var(--space-lg);
		flex-wrap: wrap;
	}

	.filters-left {
		display: flex;
		align-items: center;
		gap: var(--space-lg);
		flex-wrap: wrap;
	}

	.status-filter {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
	}

	.status-filter label {
		font-weight: 600;
		color: var(--text-primary);
	}

	.status-filter select {
		padding: var(--space-sm) var(--space-md);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		background: var(--bg-glass);
		color: var(--text-primary);
		font-size: var(--text-sm);
	}

	.search-input {
		padding: var(--space-sm) var(--space-md);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		background: var(--bg-glass);
		color: var(--text-primary);
		font-size: var(--text-sm);
		width: 300px;
	}

	.invoices-table {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
		margin-bottom: var(--space-xl);
	}

	.table-header {
		display: grid;
		grid-template-columns: 2fr 2fr 1fr 1fr 1fr 1fr 1fr 1fr;
		gap: var(--space-lg);
		padding: var(--space-md) 0;
		border-bottom: 1px solid var(--border-color);
		margin-bottom: var(--space-lg);
	}

	.header-cell {
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.table-row {
		display: grid;
		grid-template-columns: 2fr 2fr 1fr 1fr 1fr 1fr 1fr 1fr;
		gap: var(--space-lg);
		padding: var(--space-md) 0;
		border-bottom: 1px solid var(--border-color);
		align-items: center;
	}

	.table-row:last-child {
		border-bottom: none;
	}

	.invoice-id {
		font-weight: 600;
		color: var(--text-primary);
	}

	.customer-info {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.customer-name {
		font-weight: 600;
		color: var(--text-primary);
	}

	.customer-email {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.plan-info {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.plan-name {
		font-weight: 600;
		color: var(--text-primary);
	}

	.plan-interval {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		text-transform: capitalize;
	}

	.amount {
		font-weight: 600;
		color: var(--primary);
	}

	.status-badge {
		display: inline-block;
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-md);
		font-size: var(--text-xs);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
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

	.date {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.actions {
		display: flex;
		gap: var(--space-xs);
	}

	.btn-small {
		padding: var(--space-xs) var(--space-sm);
		font-size: var(--text-sm);
	}

	.empty-state {
		text-align: center;
		padding: var(--space-3xl) 0;
	}

	.empty-icon {
		width: 80px;
		height: 80px;
		margin: 0 auto var(--space-lg);
		color: var(--text-secondary);
	}

	.empty-icon svg {
		width: 100%;
		height: 100%;
	}

	.empty-state h3 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-sm);
	}

	.empty-state p {
		color: var(--text-secondary);
	}

	.loading-container,
	.error-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 400px;
		gap: var(--space-lg);
	}

	.error-message {
		color: var(--error);
		font-size: var(--text-lg);
	}

	@media (max-width: 768px) {
		.invoices-admin-page {
			padding: var(--space-lg);
		}

		.header-content {
			flex-direction: column;
			align-items: stretch;
		}

		.filters-section {
			flex-direction: column;
			align-items: stretch;
		}

		.filters-left {
			flex-direction: column;
			align-items: stretch;
		}

		.search-input {
			width: 100%;
		}

		.table-header,
		.table-row {
			grid-template-columns: 1fr;
			gap: var(--space-sm);
		}

		.header-cell {
			display: none;
		}

		.table-cell {
			display: flex;
			justify-content: space-between;
			align-items: center;
			padding: var(--space-sm) 0;
		}

		.table-cell::before {
			content: attr(data-label);
			font-weight: 600;
			color: var(--text-secondary);
		}
	}
</style> 