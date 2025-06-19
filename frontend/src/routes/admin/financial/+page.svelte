<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	interface FinancialData {
		revenue: {
			today: number;
			week: number;
			month: number;
			year: number;
			mrr: number;
			arr: number;
			growth_rate: number;
		};
		subscriptions: {
			active: number;
			new_today: number;
			new_week: number;
			new_month: number;
			cancelled_today: number;
			cancelled_week: number;
			cancelled_month: number;
			churn_rate: number;
			retention_rate: number;
			ltv: number;
			plans: Array<{
				name: string;
				count: number;
				revenue: number;
				percentage: number;
			}>;
		};
		payments: {
			successful_today: number;
			successful_week: number;
			successful_month: number;
			failed_today: number;
			failed_week: number;
			failed_month: number;
			success_rate: number;
			avg_transaction: number;
			total_volume: number;
		};
		refunds: {
			total_today: number;
			total_week: number;
			total_month: number;
			amount_today: number;
			amount_week: number;
			amount_month: number;
			refund_rate: number;
			pending_count: number;
		};
		top_customers: Array<{
			id: string;
			name: string;
			email: string;
			total_spent: number;
			subscription_plan: string;
			last_payment: string;
		}>;
		recent_transactions: Array<{
			id: string;
			type: 'payment' | 'refund' | 'subscription' | 'cancellation';
			customer: string;
			amount: number;
			status: string;
			date: string;
		}>;
	}

	let financialData: FinancialData | null = null;
	let loading = true;
	let error = '';
	let selectedPeriod = '30d';

	onMount(async () => {
		await loadFinancialData();
	});

	const loadFinancialData = async () => {
		try {
			loading = true;
			error = '';
			
			// Mock data - replace with actual API call
			await new Promise(resolve => setTimeout(resolve, 1000));
			
			financialData = {
				revenue: {
					today: 1247.50,
					week: 8934.25,
					month: 45678.90,
					year: 234567.80,
					mrr: 45678.90,
					arr: 548146.80,
					growth_rate: 0.125
				},
				subscriptions: {
					active: 892,
					new_today: 12,
					new_week: 67,
					new_month: 234,
					cancelled_today: 3,
					cancelled_week: 18,
					cancelled_month: 45,
					churn_rate: 0.032,
					retention_rate: 0.87,
					ltv: 245.67,
					plans: [
						{ name: 'Free', count: 355, revenue: 0, percentage: 39.8 },
						{ name: 'Basic', count: 298, revenue: 2980.00, percentage: 33.4 },
						{ name: 'Premium', count: 239, revenue: 9560.00, percentage: 26.8 }
					]
				},
				payments: {
					successful_today: 45,
					successful_week: 289,
					successful_month: 1234,
					failed_today: 2,
					failed_week: 12,
					failed_month: 67,
					success_rate: 0.948,
					avg_transaction: 29.99,
					total_volume: 45678.90
				},
				refunds: {
					total_today: 2,
					total_week: 8,
					total_month: 23,
					amount_today: 59.98,
					amount_week: 239.92,
					amount_month: 689.77,
					refund_rate: 0.019,
					pending_count: 3
				},
				top_customers: [
					{
						id: 'cus_1',
						name: 'John Smith',
						email: 'john.smith@example.com',
						total_spent: 299.88,
						subscription_plan: 'Premium',
						last_payment: '2024-06-17'
					},
					{
						id: 'cus_2',
						name: 'Sarah Johnson',
						email: 'sarah.j@example.com',
						total_spent: 179.94,
						subscription_plan: 'Basic',
						last_payment: '2024-06-16'
					},
					{
						id: 'cus_3',
						name: 'Mike Davis',
						email: 'mike.davis@example.com',
						total_spent: 239.91,
						subscription_plan: 'Premium',
						last_payment: '2024-06-15'
					}
				],
				recent_transactions: [
					{
						id: 'txn_1',
						type: 'payment',
						customer: 'John Smith',
						amount: 29.99,
						status: 'succeeded',
						date: '2024-06-18T10:30:00Z'
					},
					{
						id: 'txn_2',
						type: 'refund',
						customer: 'Sarah Johnson',
						amount: -19.99,
						status: 'pending',
						date: '2024-06-18T09:15:00Z'
					},
					{
						id: 'txn_3',
						type: 'subscription',
						customer: 'Mike Davis',
						amount: 39.99,
						status: 'succeeded',
						date: '2024-06-18T08:45:00Z'
					}
				]
			};
		} catch (err) {
			error = 'Failed to load financial data';
			console.error('Error loading financial data:', err);
		} finally {
			loading = false;
		}
	};

	const formatCurrency = (amount: number): string => {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(amount);
	};

	const formatNumber = (num: number): string => {
		return new Intl.NumberFormat('en-US').format(num);
	};

	const formatPercentage = (num: number): string => {
		return `${(num * 100).toFixed(1)}%`;
	};

	const formatDate = (dateString: string): string => {
		return new Date(dateString).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	};

	const getTransactionIcon = (type: string) => {
		switch (type) {
			case 'payment':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M20 7H4a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2z"></path>
					<polyline points="16,21 12,17 8,21"></polyline>
					<polyline points="12,17 12,3"></polyline>
				</svg>`;
			case 'refund':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<polyline points="3,9 9,9 9,3"></polyline>
					<path d="M11 18h6a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2H7"></path>
					<polyline points="11,18 5,12 11,6"></polyline>
				</svg>`;
			case 'subscription':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
					<circle cx="9" cy="7" r="4"></circle>
					<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
					<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
				</svg>`;
			case 'cancellation':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="12" cy="12" r="10"></circle>
					<line x1="15" y1="9" x2="9" y2="15"></line>
					<line x1="9" y1="9" x2="15" y2="15"></line>
				</svg>`;
			default:
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="12" cy="12" r="10"></circle>
					<line x1="12" y1="16" x2="12" y2="12"></line>
					<line x1="12" y1="8" x2="12.01" y2="8"></line>
				</svg>`;
		}
	};

	const getStatusBadge = (status: string, type: string) => {
		const statusConfig = {
			succeeded: { text: 'Success', class: 'status-success' },
			pending: { text: 'Pending', class: 'status-pending' },
			failed: { text: 'Failed', class: 'status-failed' },
			cancelled: { text: 'Cancelled', class: 'status-cancelled' }
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
	<title>Financial Management - Admin Dashboard</title>
</svelte:head>

<div class="financial-page">
	<div class="page-header">
		<div class="header-content">
			<div class="header-text">
				<h1>Financial Management</h1>
				<p>Monitor revenue, subscriptions, payments, and refunds</p>
			</div>
			
			<div class="header-controls">
				<div class="period-selector">
					<select bind:value={selectedPeriod} on:change={loadFinancialData} class="period-select">
						<option value="7d">Last 7 Days</option>
						<option value="30d">Last 30 Days</option>
						<option value="90d">Last 90 Days</option>
						<option value="1y">Last Year</option>
					</select>
				</div>
				
				<div class="quick-actions">
					<button class="btn btn-primary" on:click={() => goto('/admin/financial/refunds')}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<polyline points="3,9 9,9 9,3"></polyline>
							<path d="M11 18h6a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2H7"></path>
							<polyline points="11,18 5,12 11,6"></polyline>
						</svg>
						Manage Refunds
					</button>
					<button class="btn btn-outline" on:click={() => goto('/admin/financial/invoices')}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
							<polyline points="14,2 14,8 20,8"></polyline>
							<line x1="16" y1="13" x2="8" y2="13"></line>
							<line x1="16" y1="17" x2="8" y2="17"></line>
						</svg>
						View Invoices
					</button>
				</div>
			</div>
		</div>
	</div>

	{#if loading}
		<div class="loading-container">
			<LoadingSpinner size="large" color="primary" />
			<p>Loading financial data...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<p class="error-message">{error}</p>
			<button class="btn btn-primary" on:click={loadFinancialData}>
				Try Again
			</button>
		</div>
	{:else if financialData}
		<div class="financial-dashboard">
			<!-- Revenue Overview -->
			<div class="section-header">
				<h2>Revenue Overview</h2>
			</div>
			
			<div class="metrics-grid">
				<div class="metric-card glass">
					<div class="metric-header">
						<div class="metric-icon revenue">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<line x1="12" y1="1" x2="12" y2="23"></line>
								<path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"></path>
							</svg>
						</div>
						<h3>Monthly Revenue</h3>
					</div>
					<div class="metric-value">{formatCurrency(financialData.revenue.month)}</div>
					<div class="metric-details">
						<span class="metric-sub">{formatCurrency(financialData.revenue.today)} today</span>
						<span class="metric-trend positive">+{formatPercentage(financialData.revenue.growth_rate)}</span>
					</div>
				</div>

				<div class="metric-card glass">
					<div class="metric-header">
						<div class="metric-icon mrr">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polyline points="22,12 18,12 15,21 9,3 6,12 2,12"></polyline>
							</svg>
						</div>
						<h3>Monthly Recurring Revenue</h3>
					</div>
					<div class="metric-value">{formatCurrency(financialData.revenue.mrr)}</div>
					<div class="metric-details">
						<span class="metric-sub">ARR: {formatCurrency(financialData.revenue.arr)}</span>
					</div>
				</div>

				<div class="metric-card glass">
					<div class="metric-header">
						<div class="metric-icon subscriptions">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
								<circle cx="9" cy="7" r="4"></circle>
								<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
								<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
							</svg>
						</div>
						<h3>Active Subscriptions</h3>
					</div>
					<div class="metric-value">{formatNumber(financialData.subscriptions.active)}</div>
					<div class="metric-details">
						<span class="metric-sub">+{financialData.subscriptions.new_month} this month</span>
						<span class="metric-trend negative">-{financialData.subscriptions.cancelled_month} cancelled</span>
					</div>
				</div>

				<div class="metric-card glass">
					<div class="metric-header">
						<div class="metric-icon payments">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M20 7H4a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2z"></path>
								<polyline points="16,21 12,17 8,21"></polyline>
								<polyline points="12,17 12,3"></polyline>
							</svg>
						</div>
						<h3>Payment Success Rate</h3>
					</div>
					<div class="metric-value">{formatPercentage(financialData.payments.success_rate)}</div>
					<div class="metric-details">
						<span class="metric-sub">{financialData.payments.successful_month} successful</span>
						<span class="metric-sub">{financialData.payments.failed_month} failed</span>
					</div>
				</div>
			</div>

			<!-- Subscription Plans Breakdown -->
			<div class="section-header">
				<h2>Subscription Plans</h2>
			</div>
			
			<div class="plans-grid">
				{#each financialData.subscriptions.plans as plan}
					<div class="plan-card glass">
						<div class="plan-header">
							<h3>{plan.name}</h3>
							<span class="plan-percentage">{plan.percentage}%</span>
						</div>
						<div class="plan-stats">
							<div class="plan-stat">
								<span class="stat-label">Subscribers</span>
								<span class="stat-value">{formatNumber(plan.count)}</span>
							</div>
							<div class="plan-stat">
								<span class="stat-label">Revenue</span>
								<span class="stat-value">{formatCurrency(plan.revenue)}</span>
							</div>
						</div>
						<div class="plan-progress">
							<div class="progress-bar">
								<div class="progress-fill" style="width: {plan.percentage}%"></div>
							</div>
						</div>
					</div>
				{/each}
			</div>

			<!-- Financial Metrics Grid -->
			<div class="section-header">
				<h2>Key Financial Metrics</h2>
			</div>
			
			<div class="financial-metrics">
				<div class="metric-group glass">
					<h3>Customer Lifetime Value</h3>
					<div class="metric-large">{formatCurrency(financialData.subscriptions.ltv)}</div>
					<div class="metric-description">Average revenue per customer</div>
				</div>

				<div class="metric-group glass">
					<h3>Churn Rate</h3>
					<div class="metric-large">{formatPercentage(financialData.subscriptions.churn_rate)}</div>
					<div class="metric-description">Monthly customer churn</div>
				</div>

				<div class="metric-group glass">
					<h3>Refund Rate</h3>
					<div class="metric-large">{formatPercentage(financialData.refunds.refund_rate)}</div>
					<div class="metric-description">
						{formatCurrency(financialData.refunds.amount_month)} this month
						{#if financialData.refunds.pending_count > 0}
							<span class="pending-refunds">({financialData.refunds.pending_count} pending)</span>
						{/if}
					</div>
				</div>

				<div class="metric-group glass">
					<h3>Avg Transaction</h3>
					<div class="metric-large">{formatCurrency(financialData.payments.avg_transaction)}</div>
					<div class="metric-description">Average payment amount</div>
				</div>
			</div>

			<!-- Recent Transactions -->
			<div class="section-header">
				<h2>Recent Transactions</h2>
				<button class="btn btn-ghost" on:click={() => goto('/admin/financial/transactions')}>
					View All
				</button>
			</div>
			
			<div class="transactions-table glass">
				<div class="table-header">
					<div class="header-cell">Transaction</div>
					<div class="header-cell">Customer</div>
					<div class="header-cell">Amount</div>
					<div class="header-cell">Status</div>
					<div class="header-cell">Date</div>
				</div>

				{#each financialData.recent_transactions as transaction}
					<div class="table-row">
						<div class="table-cell">
							<div class="transaction-info">
								<div class="transaction-icon">
									{@html getTransactionIcon(transaction.type)}
								</div>
								<div class="transaction-details">
									<span class="transaction-type">{transaction.type}</span>
									<span class="transaction-id">#{transaction.id}</span>
								</div>
							</div>
						</div>
						<div class="table-cell">
							<span class="customer-name">{transaction.customer}</span>
						</div>
						<div class="table-cell">
							<span class="amount" class:negative={transaction.amount < 0}>
								{formatCurrency(Math.abs(transaction.amount))}
							</span>
						</div>
						<div class="table-cell">
							{@html getStatusBadge(transaction.status, transaction.type)}
						</div>
						<div class="table-cell">
							<span class="transaction-date">{formatDate(transaction.date)}</span>
						</div>
					</div>
				{/each}
			</div>

			<!-- Top Customers -->
			<div class="section-header">
				<h2>Top Customers</h2>
			</div>
			
			<div class="customers-grid">
				{#each financialData.top_customers as customer}
					<div class="customer-card glass">
						<div class="customer-header">
							<div class="customer-avatar">
								{customer.name.split(' ').map(n => n[0]).join('')}
							</div>
							<div class="customer-info">
								<h4>{customer.name}</h4>
								<span class="customer-email">{customer.email}</span>
							</div>
						</div>
						<div class="customer-stats">
							<div class="customer-stat">
								<span class="stat-label">Total Spent</span>
								<span class="stat-value">{formatCurrency(customer.total_spent)}</span>
							</div>
							<div class="customer-stat">
								<span class="stat-label">Plan</span>
								<span class="stat-value">{customer.subscription_plan}</span>
							</div>
							<div class="customer-stat">
								<span class="stat-label">Last Payment</span>
								<span class="stat-value">{new Date(customer.last_payment).toLocaleDateString()}</span>
							</div>
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>

<style>
	.financial-page {
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

	.header-controls {
		display: flex;
		align-items: center;
		gap: var(--space-lg);
		flex-wrap: wrap;
	}

	.period-select {
		padding: var(--space-sm) var(--space-md);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		background: var(--bg-glass);
		color: var(--text-primary);
		font-size: var(--text-sm);
	}

	.quick-actions {
		display: flex;
		gap: var(--space-sm);
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

	.financial-dashboard {
		display: flex;
		flex-direction: column;
		gap: var(--space-2xl);
	}

	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-lg);
	}

	.section-header h2 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	/* Metrics Grid */
	.metrics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
		gap: var(--space-lg);
		margin-bottom: var(--space-2xl);
	}

	.metric-card {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
	}

	.metric-header {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		margin-bottom: var(--space-lg);
	}

	.metric-icon {
		width: 48px;
		height: 48px;
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: var(--shadow-sm);
	}

	.metric-icon.revenue {
		background: var(--success-gradient);
		color: var(--white);
	}

	.metric-icon.mrr {
		background: var(--primary-gradient);
		color: var(--white);
	}

	.metric-icon.subscriptions {
		background: var(--info-gradient);
		color: var(--white);
	}

	.metric-icon.payments {
		background: var(--warning-gradient);
		color: var(--white);
	}

	.metric-icon svg {
		width: 24px;
		height: 24px;
	}

	.metric-header h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.metric-value {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: var(--space-md);
	}

	.metric-details {
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: var(--space-sm);
	}

	.metric-sub {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.metric-trend {
		font-size: var(--text-sm);
		font-weight: 600;
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-md);
	}

	.metric-trend.positive {
		background: var(--success-bg);
		color: var(--success-text);
	}

	.metric-trend.negative {
		background: var(--error-bg);
		color: var(--error-text);
	}

	/* Plans Grid */
	.plans-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: var(--space-lg);
		margin-bottom: var(--space-2xl);
	}

	.plan-card {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
	}

	.plan-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-lg);
	}

	.plan-header h3 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.plan-percentage {
		font-size: var(--text-lg);
		font-weight: 700;
		color: var(--primary);
	}

	.plan-stats {
		display: flex;
		justify-content: space-between;
		margin-bottom: var(--space-lg);
	}

	.plan-stat {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-xs);
	}

	.stat-label {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.stat-value {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
	}

	.plan-progress {
		margin-top: var(--space-lg);
	}

	.progress-bar {
		width: 100%;
		height: 8px;
		background: var(--bg-secondary);
		border-radius: var(--radius-full);
		overflow: hidden;
	}

	.progress-fill {
		height: 100%;
		background: var(--primary-gradient);
		transition: width var(--transition-normal);
	}

	/* Financial Metrics */
	.financial-metrics {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: var(--space-lg);
		margin-bottom: var(--space-2xl);
	}

	.metric-group {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
		text-align: center;
	}

	.metric-group h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-lg);
	}

	.metric-large {
		font-size: var(--text-4xl);
		font-weight: 700;
		color: var(--primary);
		margin-bottom: var(--space-md);
	}

	.metric-description {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.pending-refunds {
		color: var(--warning);
		font-weight: 600;
	}

	/* Transactions Table */
	.transactions-table {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
		margin-bottom: var(--space-2xl);
	}

	.table-header {
		display: grid;
		grid-template-columns: 2fr 1.5fr 1fr 1fr 1fr;
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
		grid-template-columns: 2fr 1.5fr 1fr 1fr 1fr;
		gap: var(--space-lg);
		padding: var(--space-md) 0;
		border-bottom: 1px solid var(--border-color);
		align-items: center;
	}

	.table-row:last-child {
		border-bottom: none;
	}

	.transaction-info {
		display: flex;
		align-items: center;
		gap: var(--space-md);
	}

	.transaction-icon {
		width: 40px;
		height: 40px;
		border-radius: var(--radius-lg);
		background: var(--bg-secondary);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--text-secondary);
	}

	.transaction-icon svg {
		width: 20px;
		height: 20px;
	}

	.transaction-details {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.transaction-type {
		font-weight: 600;
		color: var(--text-primary);
		text-transform: capitalize;
	}

	.transaction-id {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.customer-name {
		font-weight: 500;
		color: var(--text-primary);
	}

	.amount {
		font-weight: 600;
		color: var(--success);
	}

	.amount.negative {
		color: var(--error);
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

	.status-success {
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

	.status-cancelled {
		background: var(--bg-secondary);
		color: var(--text-secondary);
	}

	.transaction-date {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	/* Customers Grid */
	.customers-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
		gap: var(--space-lg);
	}

	.customer-card {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
	}

	.customer-header {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		margin-bottom: var(--space-lg);
	}

	.customer-avatar {
		width: 48px;
		height: 48px;
		border-radius: 50%;
		background: var(--primary-gradient);
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 600;
		color: var(--white);
		font-size: var(--text-sm);
	}

	.customer-info h4 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-xs) 0;
	}

	.customer-email {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.customer-stats {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.customer-stat {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	@media (max-width: 768px) {
		.financial-page {
			padding: var(--space-lg);
		}

		.header-content {
			flex-direction: column;
			align-items: stretch;
		}

		.header-controls {
			justify-content: space-between;
		}

		.metrics-grid {
			grid-template-columns: 1fr;
		}

		.plans-grid {
			grid-template-columns: 1fr;
		}

		.financial-metrics {
			grid-template-columns: repeat(2, 1fr);
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

		.customers-grid {
			grid-template-columns: 1fr;
		}
	}
</style> 