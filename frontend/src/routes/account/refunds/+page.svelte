<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import { subscriptionService, subscriptionUtils, type Refund } from '$lib/subscription';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let refunds: Refund[] = [];
	let loading = true;
	let error = '';
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

		await loadRefunds();
	});

	const loadRefunds = async () => {
		try {
			loading = true;
			const response = await subscriptionService.getRefunds(50);
			refunds = response.refunds || [];
		} catch (err) {
			error = 'Failed to load refund history';
			console.error('Error loading refunds:', err);
		} finally {
			loading = false;
		}
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
	<title>Refund History - BOME</title>
	<meta name="description" content="View your refund history and status" />
</svelte:head>

<div class="refunds-page">
	<div class="container">
		<header class="page-header">
			<h1>Refund History</h1>
			<p>Track the status of your refund requests</p>
		</header>

		{#if loading}
			<div class="loading-container">
				<LoadingSpinner />
				<p>Loading refund history...</p>
			</div>
		{:else if error}
			<div class="error-container">
				<p class="error-message">{error}</p>
				<button class="btn btn-primary" on:click={() => loadRefunds()}>
					Try Again
				</button>
			</div>
		{:else}
			<div class="refunds-content">
				<div class="summary-card">
					<div class="summary-item">
						<span class="label">Total Refunds</span>
						<span class="value">{refunds.length}</span>
					</div>
					<div class="summary-item">
						<span class="label">Successful Refunds</span>
						<span class="value">{refunds.filter(r => r.status === 'succeeded').length}</span>
					</div>
				</div>

				{#if refunds.length === 0}
					<div class="empty-state">
						<div class="empty-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H5a2 2 0 00-2-2z"></path>
								<path d="M8 21v-4a2 2 0 012-2h4a2 2 0 012 2v4"></path>
								<path d="M9 7V4a2 2 0 012-2h2a2 2 0 012 2v3"></path>
							</svg>
						</div>
						<h3>No refunds found</h3>
						<p>You don't have any refund history yet.</p>
					</div>
				{:else}
					<div class="refunds-table">
						<div class="table-header">
							<div class="header-cell">Refund ID</div>
							<div class="header-cell">Date</div>
							<div class="header-cell">Amount</div>
							<div class="header-cell">Reason</div>
							<div class="header-cell">Status</div>
						</div>

						{#each refunds as refund}
							<div class="table-row">
								<div class="table-cell">
									<div class="refund-info">
										<span class="refund-id">#{refund.id}</span>
										{#if refund.receiptNumber}
											<span class="receipt-number">Receipt: {refund.receiptNumber}</span>
										{/if}
									</div>
								</div>
								<div class="table-cell">
									{subscriptionUtils.formatDate(refund.createdAt)}
								</div>
								<div class="table-cell">
									<span class="amount">
										{subscriptionUtils.formatPrice(refund.amount, refund.currency)}
									</span>
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
							</div>
						{/each}
					</div>
				{/if}
			</div>
		{/if}
	</div>
</div>

<style>
	.refunds-page {
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

	.refunds-content {
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

	.refunds-table {
		margin-bottom: 2rem;
	}

	.table-header {
		display: grid;
		grid-template-columns: 2fr 1fr 1fr 1.5fr 1fr;
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
		grid-template-columns: 2fr 1fr 1fr 1.5fr 1fr;
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

	.refund-info {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.refund-id {
		font-weight: 600;
		color: var(--text-primary);
	}

	.receipt-number {
		font-size: 0.875rem;
		color: var(--text-secondary);
	}

	.amount {
		font-weight: 600;
		color: var(--primary-color);
	}

	.reason {
		font-size: 0.875rem;
		color: var(--text-secondary);
	}

	.status-badge {
		display: inline-block;
		padding: 0.25rem 0.75rem;
		border-radius: 12px;
		font-size: 0.875rem;
		font-weight: 600;
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

	.status-unknown {
		background: var(--bg-secondary);
		color: var(--text-secondary);
	}

	.failure-reason {
		font-size: 0.75rem;
		color: var(--error-text);
		margin-top: 0.25rem;
	}

	@media (max-width: 768px) {
		.page-header h1 {
			font-size: 2rem;
		}

		.refunds-content {
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
	}
</style> 