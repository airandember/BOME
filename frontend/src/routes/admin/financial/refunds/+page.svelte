<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	interface AdminRefund {
		id: string;
		amount: number;
		currency: string;
		status: 'succeeded' | 'pending' | 'failed' | 'canceled';
		reason: 'duplicate' | 'fraudulent' | 'requested_by_customer';
		paymentIntentId: string;
		chargeId?: string;
		createdAt: string;
		receiptNumber?: string;
		failureReason?: string;
		customer: {
			id: string;
			name: string;
			email: string;
		};
		originalPayment: {
			amount: number;
			date: string;
			description: string;
		};
	}

	interface RefundStats {
		total_refunds: number;
		pending_refunds: number;
		successful_refunds: number;
		failed_refunds: number;
		total_amount_refunded: number;
		pending_amount: number;
		refund_rate: number;
		avg_refund_amount: number;
	}

	let refunds: AdminRefund[] = [];
	let refundStats: RefundStats | null = null;
	let loading = true;
	let error = '';
	let selectedStatus = 'all';
	let searchQuery = '';
	let currentPage = 1;
	let totalPages = 1;
	let selectedRefunds: string[] = [];
	let showRefundModal = false;
	let selectedRefund: AdminRefund | null = null;

	const statusOptions = [
		{ value: 'all', label: 'All Refunds' },
		{ value: 'pending', label: 'Pending' },
		{ value: 'succeeded', label: 'Successful' },
		{ value: 'failed', label: 'Failed' },
		{ value: 'canceled', label: 'Canceled' }
	];

	onMount(async () => {
		await loadRefunds();
		await loadRefundStats();
	});

	const loadRefunds = async (page = 1) => {
		try {
			loading = true;
			error = '';
			
			// Mock data - replace with actual API call
			await new Promise(resolve => setTimeout(resolve, 1000));
			
			const mockRefunds: AdminRefund[] = [
				{
					id: 're_1234567890',
					amount: 2999,
					currency: 'usd',
					status: 'pending',
					reason: 'requested_by_customer',
					paymentIntentId: 'pi_1234567890',
					chargeId: 'ch_1234567890',
					createdAt: '2024-06-18T10:30:00Z',
					receiptNumber: 'RCP-001',
					customer: {
						id: 'cus_1',
						name: 'John Smith',
						email: 'john.smith@example.com'
					},
					originalPayment: {
						amount: 2999,
						date: '2024-06-01T10:30:00Z',
						description: 'Premium Subscription - Monthly'
					}
				},
				{
					id: 're_0987654321',
					amount: 1999,
					currency: 'usd',
					status: 'succeeded',
					reason: 'duplicate',
					paymentIntentId: 'pi_0987654321',
					chargeId: 'ch_0987654321',
					createdAt: '2024-06-17T14:20:00Z',
					receiptNumber: 'RCP-002',
					customer: {
						id: 'cus_2',
						name: 'Sarah Johnson',
						email: 'sarah.johnson@example.com'
					},
					originalPayment: {
						amount: 1999,
						date: '2024-05-15T14:20:00Z',
						description: 'Basic Subscription - Monthly'
					}
				},
				{
					id: 're_1122334455',
					amount: 3999,
					currency: 'usd',
					status: 'failed',
					reason: 'fraudulent',
					paymentIntentId: 'pi_1122334455',
					chargeId: 'ch_1122334455',
					createdAt: '2024-06-16T09:15:00Z',
					failureReason: 'Insufficient funds in payment method',
					customer: {
						id: 'cus_3',
						name: 'Mike Davis',
						email: 'mike.davis@example.com'
					},
					originalPayment: {
						amount: 3999,
						date: '2024-05-20T09:15:00Z',
						description: 'Premium Subscription - Yearly'
					}
				}
			];

			// Filter by status
			let filteredRefunds = mockRefunds;
			if (selectedStatus !== 'all') {
				filteredRefunds = mockRefunds.filter(refund => refund.status === selectedStatus);
			}

			// Filter by search query
			if (searchQuery.trim()) {
				const query = searchQuery.toLowerCase();
				filteredRefunds = filteredRefunds.filter(refund => 
					refund.customer.name.toLowerCase().includes(query) ||
					refund.customer.email.toLowerCase().includes(query) ||
					refund.id.toLowerCase().includes(query)
				);
			}

			refunds = filteredRefunds;
			totalPages = Math.ceil(filteredRefunds.length / 20);
			currentPage = page;
		} catch (err) {
			error = 'Failed to load refunds';
			console.error('Error loading refunds:', err);
		} finally {
			loading = false;
		}
	};

	const loadRefundStats = async () => {
		try {
			// Mock data - replace with actual API call
			refundStats = {
				total_refunds: 156,
				pending_refunds: 23,
				successful_refunds: 128,
				failed_refunds: 5,
				total_amount_refunded: 15678.90,
				pending_amount: 2340.50,
				refund_rate: 0.019,
				avg_refund_amount: 100.51
			};
		} catch (err) {
			console.error('Error loading refund stats:', err);
		}
	};

	const handleStatusFilter = async (status: string) => {
		selectedStatus = status;
		currentPage = 1;
		await loadRefunds(1);
	};

	const handleSearch = async () => {
		currentPage = 1;
		await loadRefunds(1);
	};

	const handlePageChange = async (page: number) => {
		if (page >= 1 && page <= totalPages) {
			await loadRefunds(page);
		}
	};

	const handleRefundSelect = (refundId: string) => {
		if (selectedRefunds.includes(refundId)) {
			selectedRefunds = selectedRefunds.filter(id => id !== refundId);
		} else {
			selectedRefunds = [...selectedRefunds, refundId];
		}
	};

	const handleSelectAll = () => {
		if (selectedRefunds.length === refunds.length) {
			selectedRefunds = [];
		} else {
			selectedRefunds = refunds.map(refund => refund.id);
		}
	};

	const handleViewRefund = (refund: AdminRefund) => {
		selectedRefund = refund;
		showRefundModal = true;
	};

	const handleProcessRefund = async (refundId: string, action: 'approve' | 'reject') => {
		try {
			// Mock API call - replace with actual implementation
			await new Promise(resolve => setTimeout(resolve, 500));
			
			showToast(`Refund ${action}d successfully`, 'success');
			await loadRefunds(currentPage);
			await loadRefundStats();
		} catch (err) {
			showToast(`Failed to ${action} refund`, 'error');
			console.error(`Error ${action}ing refund:`, err);
		}
	};

	const handleBulkAction = async (action: 'approve' | 'reject') => {
		if (selectedRefunds.length === 0) {
			showToast('Please select refunds to process', 'warning');
			return;
		}

		try {
			// Mock API call - replace with actual implementation
			await new Promise(resolve => setTimeout(resolve, 1000));
			
			showToast(`${selectedRefunds.length} refunds ${action}d successfully`, 'success');
			selectedRefunds = [];
			await loadRefunds(currentPage);
			await loadRefundStats();
		} catch (err) {
			showToast(`Failed to ${action} selected refunds`, 'error');
			console.error(`Error ${action}ing refunds:`, err);
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
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	};

	const getStatusBadge = (status: string) => {
		const statusConfig = {
			succeeded: { text: 'Succeeded', class: 'status-succeeded' },
			pending: { text: 'Pending', class: 'status-pending' },
			failed: { text: 'Failed', class: 'status-failed' },
			canceled: { text: 'Canceled', class: 'status-canceled' }
		};

		const config = statusConfig[status as keyof typeof statusConfig] || { text: status, class: 'status-unknown' };
		
		return `
			<span class="status-badge ${config.class}">
				${config.text}
			</span>
		`;
	};

	const getReasonText = (reason: string) => {
		const reasonMap = {
			duplicate: 'Duplicate Payment',
			fraudulent: 'Fraudulent Transaction',
			requested_by_customer: 'Customer Request'
		};
		return reasonMap[reason as keyof typeof reasonMap] || reason;
	};
</script>

<svelte:head>
	<title>Refund Management - Admin Dashboard</title>
</svelte:head>

<div class="refunds-admin-page">
	<div class="page-header">
		<div class="header-content">
			<div class="header-text">
				<h1>Refund Management</h1>
				<p>Process and manage customer refund requests</p>
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

	{#if refundStats}
		<!-- Refund Statistics -->
		<div class="stats-grid">
			<div class="stat-card glass">
				<div class="stat-header">
					<div class="stat-icon total">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<polyline points="3,9 9,9 9,3"></polyline>
							<path d="M11 18h6a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2H7"></path>
							<polyline points="11,18 5,12 11,6"></polyline>
						</svg>
					</div>
					<h3>Total Refunds</h3>
				</div>
				<div class="stat-value">{refundStats.total_refunds}</div>
				<div class="stat-details">
					<span class="stat-sub">{formatCurrency(refundStats.total_amount_refunded)}</span>
				</div>
			</div>

			<div class="stat-card glass">
				<div class="stat-header">
					<div class="stat-icon pending">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<circle cx="12" cy="12" r="10"></circle>
							<polyline points="12,6 12,12 16,14"></polyline>
						</svg>
					</div>
					<h3>Pending Refunds</h3>
				</div>
				<div class="stat-value">{refundStats.pending_refunds}</div>
				<div class="stat-details">
					<span class="stat-sub">{formatCurrency(refundStats.pending_amount)}</span>
				</div>
			</div>

			<div class="stat-card glass">
				<div class="stat-header">
					<div class="stat-icon success">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<polyline points="20,6 9,17 4,12"></polyline>
						</svg>
					</div>
					<h3>Successful Refunds</h3>
				</div>
				<div class="stat-value">{refundStats.successful_refunds}</div>
				<div class="stat-details">
					<span class="stat-sub">Rate: {(refundStats.refund_rate * 100).toFixed(1)}%</span>
				</div>
			</div>

			<div class="stat-card glass">
				<div class="stat-header">
					<div class="stat-icon average">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<line x1="12" y1="1" x2="12" y2="23"></line>
							<path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"></path>
						</svg>
					</div>
					<h3>Average Refund</h3>
				</div>
				<div class="stat-value">{formatCurrency(refundStats.avg_refund_amount * 100)}</div>
				<div class="stat-details">
					<span class="stat-sub">{refundStats.failed_refunds} failed</span>
				</div>
			</div>
		</div>
	{/if}

	<!-- Filters and Search -->
	<div class="filters-section glass">
		<div class="filters-left">
			<div class="status-filter">
				<label for="status-select">Status:</label>
				<select id="status-select" bind:value={selectedStatus} on:change={() => handleStatusFilter(selectedStatus)}>
					{#each statusOptions as option}
						<option value={option.value}>{option.label}</option>
					{/each}
				</select>
			</div>

			<div class="search-filter">
				<input
					type="text"
					placeholder="Search by customer name, email, or refund ID..."
					bind:value={searchQuery}
					on:input={handleSearch}
					class="search-input"
				/>
			</div>
		</div>

		{#if selectedRefunds.length > 0}
			<div class="bulk-actions">
				<span class="selected-count">{selectedRefunds.length} selected</span>
				<button class="btn btn-success btn-small" on:click={() => handleBulkAction('approve')}>
					Approve Selected
				</button>
				<button class="btn btn-error btn-small" on:click={() => handleBulkAction('reject')}>
					Reject Selected
				</button>
			</div>
		{/if}
	</div>

	{#if loading}
		<div class="loading-container">
			<LoadingSpinner size="large" color="primary" />
			<p>Loading refunds...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<p class="error-message">{error}</p>
			<button class="btn btn-primary" on:click={() => loadRefunds()}>
				Try Again
			</button>
		</div>
	{:else}
		<!-- Refunds Table -->
		<div class="refunds-table glass">
			<div class="table-header">
				<div class="header-cell">
					<input
						type="checkbox"
						checked={selectedRefunds.length === refunds.length && refunds.length > 0}
						on:change={handleSelectAll}
					/>
				</div>
				<div class="header-cell">Refund ID</div>
				<div class="header-cell">Customer</div>
				<div class="header-cell">Amount</div>
				<div class="header-cell">Reason</div>
				<div class="header-cell">Status</div>
				<div class="header-cell">Date</div>
				<div class="header-cell">Actions</div>
			</div>

			{#each refunds as refund}
				<div class="table-row">
					<div class="table-cell">
						<input
							type="checkbox"
							checked={selectedRefunds.includes(refund.id)}
							on:change={() => handleRefundSelect(refund.id)}
						/>
					</div>
					<div class="table-cell">
						<div class="refund-id">
							<span class="id-text">#{refund.id}</span>
							{#if refund.receiptNumber}
								<span class="receipt-number">Receipt: {refund.receiptNumber}</span>
							{/if}
						</div>
					</div>
					<div class="table-cell">
						<div class="customer-info">
							<span class="customer-name">{refund.customer.name}</span>
							<span class="customer-email">{refund.customer.email}</span>
						</div>
					</div>
					<div class="table-cell">
						<span class="amount">{formatCurrency(refund.amount, refund.currency)}</span>
					</div>
					<div class="table-cell">
						<span class="reason">{getReasonText(refund.reason)}</span>
					</div>
					<div class="table-cell">
						{@html getStatusBadge(refund.status)}
						{#if refund.status === 'failed' && refund.failureReason}
							<div class="failure-reason">{refund.failureReason}</div>
						{/if}
					</div>
					<div class="table-cell">
						<span class="date">{formatDate(refund.createdAt)}</span>
					</div>
					<div class="table-cell">
						<div class="actions">
							<button class="btn btn-ghost btn-small" on:click={() => handleViewRefund(refund)}>
								View
							</button>
							{#if refund.status === 'pending'}
								<button class="btn btn-success btn-small" on:click={() => handleProcessRefund(refund.id, 'approve')}>
									Approve
								</button>
								<button class="btn btn-error btn-small" on:click={() => handleProcessRefund(refund.id, 'reject')}>
									Reject
								</button>
							{/if}
						</div>
					</div>
				</div>
			{/each}

			{#if refunds.length === 0}
				<div class="empty-state">
					<div class="empty-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<polyline points="3,9 9,9 9,3"></polyline>
							<path d="M11 18h6a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2H7"></path>
							<polyline points="11,18 5,12 11,6"></polyline>
						</svg>
					</div>
					<h3>No refunds found</h3>
					<p>No refunds match your current filters.</p>
				</div>
			{/if}
		</div>

		<!-- Pagination -->
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

<!-- Refund Details Modal -->
{#if showRefundModal && selectedRefund}
	<div class="modal-overlay" on:click={() => showRefundModal = false}>
		<div class="modal-content" on:click|stopPropagation>
			<div class="modal-header">
				<h2>Refund Details</h2>
				<button class="modal-close" on:click={() => showRefundModal = false}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"></line>
						<line x1="6" y1="6" x2="18" y2="18"></line>
					</svg>
				</button>
			</div>
			
			<div class="modal-body">
				<div class="refund-details">
					<div class="detail-section">
						<h3>Refund Information</h3>
						<div class="detail-grid">
							<div class="detail-item">
								<span class="detail-label">Refund ID:</span>
								<span class="detail-value">#{selectedRefund.id}</span>
							</div>
							<div class="detail-item">
								<span class="detail-label">Amount:</span>
								<span class="detail-value">{formatCurrency(selectedRefund.amount, selectedRefund.currency)}</span>
							</div>
							<div class="detail-item">
								<span class="detail-label">Status:</span>
								<span class="detail-value">{@html getStatusBadge(selectedRefund.status)}</span>
							</div>
							<div class="detail-item">
								<span class="detail-label">Reason:</span>
								<span class="detail-value">{getReasonText(selectedRefund.reason)}</span>
							</div>
							<div class="detail-item">
								<span class="detail-label">Created:</span>
								<span class="detail-value">{formatDate(selectedRefund.createdAt)}</span>
							</div>
							{#if selectedRefund.receiptNumber}
								<div class="detail-item">
									<span class="detail-label">Receipt:</span>
									<span class="detail-value">{selectedRefund.receiptNumber}</span>
								</div>
							{/if}
						</div>
					</div>

					<div class="detail-section">
						<h3>Customer Information</h3>
						<div class="detail-grid">
							<div class="detail-item">
								<span class="detail-label">Name:</span>
								<span class="detail-value">{selectedRefund.customer.name}</span>
							</div>
							<div class="detail-item">
								<span class="detail-label">Email:</span>
								<span class="detail-value">{selectedRefund.customer.email}</span>
							</div>
							<div class="detail-item">
								<span class="detail-label">Customer ID:</span>
								<span class="detail-value">{selectedRefund.customer.id}</span>
							</div>
						</div>
					</div>

					<div class="detail-section">
						<h3>Original Payment</h3>
						<div class="detail-grid">
							<div class="detail-item">
								<span class="detail-label">Amount:</span>
								<span class="detail-value">{formatCurrency(selectedRefund.originalPayment.amount)}</span>
							</div>
							<div class="detail-item">
								<span class="detail-label">Date:</span>
								<span class="detail-value">{formatDate(selectedRefund.originalPayment.date)}</span>
							</div>
							<div class="detail-item">
								<span class="detail-label">Description:</span>
								<span class="detail-value">{selectedRefund.originalPayment.description}</span>
							</div>
							<div class="detail-item">
								<span class="detail-label">Payment Intent:</span>
								<span class="detail-value">{selectedRefund.paymentIntentId}</span>
							</div>
						</div>
					</div>

					{#if selectedRefund.status === 'failed' && selectedRefund.failureReason}
						<div class="detail-section">
							<h3>Failure Information</h3>
							<div class="failure-info">
								<p>{selectedRefund.failureReason}</p>
							</div>
						</div>
					{/if}
				</div>
			</div>
			
			<div class="modal-footer">
				{#if selectedRefund.status === 'pending'}
					<button class="btn btn-success" on:click={() => handleProcessRefund(selectedRefund.id, 'approve')}>
						Approve Refund
					</button>
					<button class="btn btn-error" on:click={() => handleProcessRefund(selectedRefund.id, 'reject')}>
						Reject Refund
					</button>
				{/if}
				<button class="btn btn-outline" on:click={() => showRefundModal = false}>
					Close
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.refunds-admin-page {
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

	/* Stats Grid */
	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: var(--space-lg);
		margin-bottom: var(--space-2xl);
	}

	.stat-card {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
	}

	.stat-header {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		margin-bottom: var(--space-lg);
	}

	.stat-icon {
		width: 48px;
		height: 48px;
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: var(--shadow-sm);
	}

	.stat-icon.total {
		background: var(--primary-gradient);
		color: var(--white);
	}

	.stat-icon.pending {
		background: var(--warning-gradient);
		color: var(--white);
	}

	.stat-icon.success {
		background: var(--success-gradient);
		color: var(--white);
	}

	.stat-icon.average {
		background: var(--info-gradient);
		color: var(--white);
	}

	.stat-icon svg {
		width: 24px;
		height: 24px;
	}

	.stat-header h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.stat-value {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: var(--space-md);
	}

	.stat-details {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.stat-sub {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	/* Filters Section */
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

	.bulk-actions {
		display: flex;
		align-items: center;
		gap: var(--space-md);
	}

	.selected-count {
		font-weight: 600;
		color: var(--text-primary);
	}

	.btn-small {
		padding: var(--space-xs) var(--space-sm);
		font-size: var(--text-sm);
	}

	/* Refunds Table */
	.refunds-table {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
		margin-bottom: var(--space-xl);
	}

	.table-header {
		display: grid;
		grid-template-columns: auto 2fr 2fr 1fr 1.5fr 1fr 1fr 1.5fr;
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
		grid-template-columns: auto 2fr 2fr 1fr 1.5fr 1fr 1fr 1.5fr;
		gap: var(--space-lg);
		padding: var(--space-md) 0;
		border-bottom: 1px solid var(--border-color);
		align-items: center;
	}

	.table-row:last-child {
		border-bottom: none;
	}

	.refund-id {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.id-text {
		font-weight: 600;
		color: var(--text-primary);
	}

	.receipt-number {
		font-size: var(--text-sm);
		color: var(--text-secondary);
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

	.amount {
		font-weight: 600;
		color: var(--primary);
	}

	.reason {
		font-size: var(--text-sm);
		color: var(--text-secondary);
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

	.status-succeeded {
		background: var(--success-bg);
		color: var(--success-text);
	}

	.status-pending {
		background: var(--warning-bg);
		color: var(--warning-text);
	}

	.status-failed {
		background: var(--error-bg);
		color: var(--error-text);
	}

	.status-canceled {
		background: var(--bg-secondary);
		color: var(--text-secondary);
	}

	.failure-reason {
		font-size: var(--text-xs);
		color: var(--error);
		margin-top: var(--space-xs);
	}

	.date {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.actions {
		display: flex;
		gap: var(--space-xs);
		flex-wrap: wrap;
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

	/* Pagination */
	.pagination {
		display: flex;
		justify-content: center;
		align-items: center;
		gap: var(--space-lg);
		margin-top: var(--space-xl);
	}

	.page-numbers {
		display: flex;
		gap: var(--space-sm);
	}

	.page-number {
		padding: var(--space-sm) var(--space-md);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		background: var(--bg-glass);
		color: var(--text-primary);
		cursor: pointer;
		transition: all var(--transition-fast);
	}

	.page-number:hover {
		background: var(--bg-hover);
	}

	.page-number.active {
		background: var(--primary);
		color: var(--white);
		border-color: var(--primary);
	}

	/* Modal Styles */
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}

	.modal-content {
		background: var(--bg-primary);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
		max-width: 800px;
		width: 90%;
		max-height: 90vh;
		overflow-y: auto;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-xl);
		border-bottom: 1px solid var(--border-color);
	}

	.modal-header h2 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.modal-close {
		width: 32px;
		height: 32px;
		border: none;
		background: var(--bg-glass);
		border-radius: var(--radius-md);
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		color: var(--text-secondary);
	}

	.modal-close:hover {
		background: var(--bg-hover);
		color: var(--text-primary);
	}

	.modal-close svg {
		width: 18px;
		height: 18px;
	}

	.modal-body {
		padding: var(--space-xl);
	}

	.detail-section {
		margin-bottom: var(--space-xl);
	}

	.detail-section h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-lg);
	}

	.detail-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: var(--space-lg);
	}

	.detail-item {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.detail-label {
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.detail-value {
		font-size: var(--text-base);
		color: var(--text-primary);
	}

	.failure-info {
		padding: var(--space-md);
		background: var(--error-bg);
		border-radius: var(--radius-md);
		border-left: 4px solid var(--error);
	}

	.failure-info p {
		color: var(--error-text);
		margin: 0;
	}

	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-md);
		padding: var(--space-xl);
		border-top: 1px solid var(--border-color);
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
		.refunds-admin-page {
			padding: var(--space-lg);
		}

		.header-content {
			flex-direction: column;
			align-items: stretch;
		}

		.stats-grid {
			grid-template-columns: repeat(2, 1fr);
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

		.modal-content {
			width: 95%;
		}

		.detail-grid {
			grid-template-columns: 1fr;
		}
	}
</style> 