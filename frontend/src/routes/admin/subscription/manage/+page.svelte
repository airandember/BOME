<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import { subscriptionService } from '$lib/subscription';
	import { showToast } from '$lib/toast';
	import { formatCurrency } from '$lib/utils/currency';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import Navigation from '$lib/components/Navigation.svelte';
	import Footer from '$lib/components/Footer.svelte';

	let loading = true;
	let error = '';
	let subscription: any = null;
	let billingHistory: any[] = [];
	let showCancelModal = false;
	let showUpdateModal = false;
	let isSubmitting = false;

	// Cancel form
	let cancelReason = '';
	let cancelFeedback = '';

	// Update form
	let newPlanId = '';

	onMount(async () => {
		if (!$auth.isAuthenticated) {
			goto('/auth/login');
			return;
		}

		try {
			await loadSubscriptionData();
		} catch (err) {
			error = 'Failed to load subscription data';
			console.error('Error loading subscription:', err);
		} finally {
			loading = false;
		}
	});

	async function loadSubscriptionData() {
		// Load current subscription
		const subResponse = await subscriptionService.getCurrentSubscription();
		subscription = subResponse.subscription;

		// Load billing history
		const historyResponse = await subscriptionService.getBillingHistory();
		billingHistory = historyResponse.invoices || [];
	}

	async function handleCancelSubscription() {
		if (!cancelReason.trim()) {
			showToast('Please select a cancellation reason', 'warning');
			return;
		}

		try {
			isSubmitting = true;
			await subscriptionService.cancelSubscription(cancelReason, cancelFeedback);
			showToast('Subscription cancelled successfully', 'success');
			showCancelModal = false;
			await loadSubscriptionData();
		} catch (err) {
			showToast('Failed to cancel subscription', 'error');
			console.error('Error cancelling subscription:', err);
		} finally {
			isSubmitting = false;
		}
	}

	async function handleUpdateSubscription() {
		if (!newPlanId) {
			showToast('Please select a new plan', 'warning');
			return;
		}

		try {
			isSubmitting = true;
			await subscriptionService.updateSubscription(newPlanId);
			showToast('Subscription updated successfully', 'success');
			showUpdateModal = false;
			await loadSubscriptionData();
		} catch (err) {
			showToast('Failed to update subscription', 'error');
			console.error('Error updating subscription:', err);
		} finally {
			isSubmitting = false;
		}
	}

	async function handleReactivateSubscription() {
		try {
			await subscriptionService.reactivateSubscription();
			showToast('Subscription reactivated successfully', 'success');
			await loadSubscriptionData();
		} catch (err) {
			showToast('Failed to reactivate subscription', 'error');
			console.error('Error reactivating subscription:', err);
		}
	}

	function formatDate(dateString: string) {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	function getStatusBadge(status: string) {
		const statusConfig = {
			active: { text: 'Active', class: 'bg-green-100 text-green-800' },
			cancelled: { text: 'Cancelled', class: 'bg-red-100 text-red-800' },
			past_due: { text: 'Past Due', class: 'bg-yellow-100 text-yellow-800' },
			unpaid: { text: 'Unpaid', class: 'bg-red-100 text-red-800' },
			trialing: { text: 'Trial', class: 'bg-blue-100 text-blue-800' }
		};
		return statusConfig[status] || { text: status, class: 'bg-gray-100 text-gray-800' };
	}

	function getDaysUntilRenewal() {
		if (!subscription?.current_period_end) return null;
		
		const endDate = new Date(subscription.current_period_end);
		const now = new Date();
		const diffTime = endDate.getTime() - now.getTime();
		const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
		
		return diffDays;
	}

	function openStripePortal() {
		// Redirect to Stripe customer portal
		window.open(subscription?.customer_portal_url, '_blank');
	}
</script>

<svelte:head>
	<title>Manage Subscription - BOME</title>
	<meta name="description" content="Manage your BOME subscription, billing, and account settings" />
</svelte:head>

<Navigation />

<div class="subscription-manage-page">
	<div class="container">
		<header class="page-header">
			<h1>Manage Subscription</h1>
			<p>View and manage your subscription details, billing, and account settings</p>
		</header>

		{#if loading}
			<div class="loading-container">
				<LoadingSpinner />
				<p>Loading subscription details...</p>
			</div>
		{:else if error}
			<div class="error-container">
				<p class="error-message">{error}</p>
				<button class="btn btn-primary" on:click={() => window.location.reload()}>
					Try Again
				</button>
			</div>
		{:else if !subscription}
			<div class="no-subscription">
				<div class="no-subscription-content">
					<CreditCardIcon class="no-subscription-icon" />
					<h2>No Active Subscription</h2>
					<p>You don't have an active subscription. Subscribe to access exclusive content.</p>
					<button class="btn btn-primary" on:click={() => goto('/subscription')}>
						View Plans
					</button>
				</div>
			</div>
		{:else}
			<div class="subscription-details" in:fly={{ y: 20, duration: 300 }}>
				<!-- Subscription Overview -->
				<div class="overview-card">
					<div class="overview-header">
						<div class="overview-info">
							<h2>{subscription.plan_name}</h2>
							<div class="status-badge">
								<svelte:component this={getStatusBadge(subscription.status).icon} class="status-icon" />
								<span>{getStatusBadge(subscription.status).text}</span>
							</div>
						</div>
						<div class="overview-price">
							<span class="price-amount">{formatCurrency(subscription.price, subscription.currency)}</span>
							<span class="price-interval">/{subscription.interval}</span>
						</div>
					</div>

					<div class="overview-details">
						<div class="detail-item">
							<CalendarIcon class="detail-icon" />
							<div class="detail-content">
								<label>Next Billing</label>
								<span>{formatDate(subscription.current_period_end)}</span>
								{#if getDaysUntilRenewal() !== null}
									<small class="renewal-countdown">
										{getDaysUntilRenewal()} days until renewal
									</small>
								{/if}
							</div>
						</div>

						<div class="detail-item">
							<CreditCardIcon class="detail-icon" />
							<div class="detail-content">
								<label>Payment Method</label>
								<span>•••• •••• •••• {subscription.payment_method?.last4 || '****'}</span>
								<small>{subscription.payment_method?.brand || 'Card'} ending in {subscription.payment_method?.last4 || '****'}</small>
							</div>
						</div>

						<div class="detail-item">
							<CurrencyDollarIcon class="detail-icon" />
							<div class="detail-content">
								<label>Total Paid</label>
								<span>{formatCurrency(subscription.total_paid || 0, subscription.currency)}</span>
								<small>Since {formatDate(subscription.created_at)}</small>
							</div>
						</div>
					</div>
				</div>

				<!-- Action Buttons -->
				<div class="action-buttons">
					{#if subscription.status === 'active'}
						<button class="btn btn-secondary" on:click={() => showUpdateModal = true}>
							<PencilIcon class="btn-icon" />
							Change Plan
						</button>
						<button class="btn btn-secondary" on:click={openStripePortal}>
							<CreditCardIcon class="btn-icon" />
							Update Payment Method
						</button>
						<button class="btn btn-danger" on:click={() => showCancelModal = true}>
							<TrashIcon class="btn-icon" />
							Cancel Subscription
						</button>
					{:else if subscription.status === 'cancelled'}
						<button class="btn btn-primary" on:click={handleReactivateSubscription}>
							<CheckCircleIcon class="btn-icon" />
							Reactivate Subscription
						</button>
						<button class="btn btn-secondary" on:click={() => goto('/subscription')}>
							View Plans
						</button>
					{:else if subscription.status === 'past_due' || subscription.status === 'unpaid'}
						<button class="btn btn-primary" on:click={openStripePortal}>
							<CreditCardIcon class="btn-icon" />
							Update Payment Method
						</button>
						<button class="btn btn-secondary" on:click={() => goto('/subscription')}>
							View Plans
						</button>
					{/if}
				</div>

				<!-- Billing History -->
				<div class="billing-history">
					<h3>Billing History</h3>
					{#if billingHistory.length > 0}
						<div class="billing-table">
							<div class="table-header">
								<span>Date</span>
								<span>Description</span>
								<span>Amount</span>
								<span>Status</span>
								<span>Actions</span>
							</div>
							{#each billingHistory as invoice}
								<div class="table-row">
									<span>{formatDate(invoice.date)}</span>
									<span>{invoice.description}</span>
									<span>{formatCurrency(invoice.amount, invoice.currency)}</span>
									<span class="invoice-status {invoice.status}">{invoice.status}</span>
									<div class="invoice-actions">
										{#if invoice.pdf_url}
											<a href={invoice.pdf_url} target="_blank" class="btn-link">
												<DocumentTextIcon class="action-icon" />
												Download
											</a>
										{/if}
									</div>
								</div>
							{/each}
						</div>
					{:else}
						<div class="no-billing">
							<p>No billing history available</p>
						</div>
					{/if}
				</div>
			</div>
		{/if}
	</div>
</div>

<!-- Cancel Subscription Modal -->
{#if showCancelModal}
	<div class="modal-overlay" in:fade={{ duration: 200 }}>
		<div class="modal" in:fly={{ y: 20, duration: 200 }}>
			<div class="modal-header">
				<h3>Cancel Subscription</h3>
				<button class="modal-close" on:click={() => showCancelModal = false}>
					<XCircleIcon class="close-icon" />
				</button>
			</div>
			
			<div class="modal-content">
				<p>We're sorry to see you go. Please let us know why you're cancelling so we can improve our service.</p>
				
				<div class="form-group">
					<label for="cancel-reason">Cancellation Reason</label>
					<select id="cancel-reason" bind:value={cancelReason} class="form-select">
						<option value="">Select a reason</option>
						<option value="too_expensive">Too expensive</option>
						<option value="not_using">Not using enough</option>
						<option value="missing_features">Missing features</option>
						<option value="technical_issues">Technical issues</option>
						<option value="switching_service">Switching to another service</option>
						<option value="other">Other</option>
					</select>
				</div>

				<div class="form-group">
					<label for="cancel-feedback">Additional Feedback (Optional)</label>
					<textarea 
						id="cancel-feedback" 
						bind:value={cancelFeedback} 
						class="form-textarea"
						placeholder="Tell us more about your experience..."
						rows="3"
					></textarea>
				</div>

				<div class="modal-actions">
					<button class="btn btn-secondary" on:click={() => showCancelModal = false}>
						Keep Subscription
					</button>
					<button 
						class="btn btn-danger" 
						on:click={handleCancelSubscription}
						disabled={isSubmitting}
					>
						{isSubmitting ? 'Cancelling...' : 'Cancel Subscription'}
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}

<!-- Update Subscription Modal -->
{#if showUpdateModal}
	<div class="modal-overlay" in:fade={{ duration: 200 }}>
		<div class="modal" in:fly={{ y: 20, duration: 200 }}>
			<div class="modal-header">
				<h3>Change Subscription Plan</h3>
				<button class="modal-close" on:click={() => showUpdateModal = false}>
					<XCircleIcon class="close-icon" />
				</button>
			</div>
			
			<div class="modal-content">
				<p>Select a new plan for your subscription. Changes will take effect at your next billing cycle.</p>
				
				<div class="form-group">
					<label for="new-plan">New Plan</label>
					<select id="new-plan" bind:value={newPlanId} class="form-select">
						<option value="">Select a plan</option>
						<option value="monthly">Monthly Plan - $9.99/month</option>
						<option value="annual">Annual Plan - $99.99/year</option>
						<option value="premium">Premium Plan - $19.99/month</option>
					</select>
				</div>

				<div class="modal-actions">
					<button class="btn btn-secondary" on:click={() => showUpdateModal = false}>
						Cancel
					</button>
					<button 
						class="btn btn-primary" 
						on:click={handleUpdateSubscription}
						disabled={isSubmitting}
					>
						{isSubmitting ? 'Updating...' : 'Update Plan'}
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}

<Footer />

<style>
	.subscription-manage-page {
		min-height: 100vh;
		padding: 2rem 0;
		background: var(--bg-gradient);
	}

	.container {
		max-width: 1000px;
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

	.no-subscription {
		text-align: center;
		padding: 4rem 0;
	}

	.no-subscription-content {
		max-width: 400px;
		margin: 0 auto;
	}

	.no-subscription-icon {
		width: 64px;
		height: 64px;
		color: var(--gray-400);
		margin: 0 auto 1rem;
	}

	.no-subscription h2 {
		font-size: 1.5rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.no-subscription p {
		color: var(--text-secondary);
		margin-bottom: 2rem;
	}

	/* Overview Card */
	.overview-card {
		background: var(--card-bg);
		border-radius: 20px;
		padding: 2rem;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
		margin-bottom: 2rem;
	}

	.overview-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 2rem;
		gap: 2rem;
	}

	.overview-info h2 {
		font-size: 1.75rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.status-badge {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		border-radius: 12px;
		font-size: 0.875rem;
		font-weight: 600;
	}

	.status-badge.bg-green-100 {
		background: var(--success-bg);
		color: var(--success-text);
	}

	.status-badge.bg-red-100 {
		background: var(--error-bg);
		color: var(--error-text);
	}

	.status-badge.bg-yellow-100 {
		background: var(--warning-bg);
		color: var(--warning-text);
	}

	.status-badge.bg-blue-100 {
		background: var(--info-bg);
		color: var(--info-text);
	}

	.status-icon {
		width: 16px;
		height: 16px;
	}

	.overview-price {
		text-align: right;
	}

	.price-amount {
		font-size: 2rem;
		font-weight: 700;
		color: var(--primary-color);
	}

	.price-interval {
		font-size: 1rem;
		color: var(--text-secondary);
	}

	.overview-details {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 1.5rem;
	}

	.detail-item {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
	}

	.detail-icon {
		width: 20px;
		height: 20px;
		color: var(--primary-color);
		flex-shrink: 0;
		margin-top: 0.25rem;
	}

	.detail-content {
		flex: 1;
	}

	.detail-content label {
		display: block;
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--text-secondary);
		margin-bottom: 0.25rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.detail-content span {
		display: block;
		font-size: 1rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.25rem;
	}

	.detail-content small {
		font-size: 0.875rem;
		color: var(--text-secondary);
	}

	.renewal-countdown {
		color: var(--warning-text);
		font-weight: 500;
	}

	/* Action Buttons */
	.action-buttons {
		display: flex;
		gap: 1rem;
		margin-bottom: 2rem;
		flex-wrap: wrap;
	}

	.btn {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1.5rem;
		border-radius: 12px;
		font-weight: 600;
		text-decoration: none;
		border: none;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.btn-primary {
		background: var(--primary-color);
		color: white;
	}

	.btn-primary:hover {
		background: var(--primary-hover);
	}

	.btn-secondary {
		background: var(--secondary-bg);
		color: var(--secondary-text);
		border: 1px solid var(--border-color);
	}

	.btn-secondary:hover {
		background: var(--secondary-hover);
	}

	.btn-danger {
		background: var(--error-color);
		color: white;
	}

	.btn-danger:hover {
		background: var(--error-hover);
	}

	.btn-icon {
		width: 16px;
		height: 16px;
	}

	/* Billing History */
	.billing-history {
		background: var(--card-bg);
		border-radius: 16px;
		padding: 1.5rem;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
	}

	.billing-history h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1.5rem;
	}

	.billing-table {
		border: 1px solid var(--border-color);
		border-radius: 12px;
		overflow: hidden;
	}

	.table-header {
		display: grid;
		grid-template-columns: 1fr 2fr 1fr 1fr 1fr;
		gap: 1rem;
		padding: 1rem;
		background: var(--gray-50);
		font-weight: 600;
		color: var(--text-secondary);
		font-size: 0.875rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.table-row {
		display: grid;
		grid-template-columns: 1fr 2fr 1fr 1fr 1fr;
		gap: 1rem;
		padding: 1rem;
		border-top: 1px solid var(--border-color);
		align-items: center;
	}

	.table-row:hover {
		background: var(--gray-50);
	}

	.invoice-status {
		font-size: 0.875rem;
		font-weight: 600;
		padding: 0.25rem 0.75rem;
		border-radius: 8px;
		text-align: center;
	}

	.invoice-status.paid {
		background: var(--success-bg);
		color: var(--success-text);
	}

	.invoice-status.pending {
		background: var(--warning-bg);
		color: var(--warning-text);
	}

	.invoice-status.failed {
		background: var(--error-bg);
		color: var(--error-text);
	}

	.invoice-actions {
		display: flex;
		gap: 0.5rem;
	}

	.btn-link {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		color: var(--primary-color);
		text-decoration: none;
		font-size: 0.875rem;
		font-weight: 500;
	}

	.btn-link:hover {
		text-decoration: underline;
	}

	.action-icon {
		width: 14px;
		height: 14px;
	}

	.no-billing {
		text-align: center;
		padding: 2rem;
		color: var(--text-secondary);
	}

	/* Modal Styles */
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
		padding: 1rem;
	}

	.modal {
		background: var(--card-bg);
		border-radius: 16px;
		max-width: 500px;
		width: 100%;
		max-height: 90vh;
		overflow-y: auto;
		box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1.5rem;
		border-bottom: 1px solid var(--border-color);
	}

	.modal-header h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.modal-close {
		background: none;
		border: none;
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 8px;
		color: var(--text-secondary);
	}

	.modal-close:hover {
		background: var(--gray-100);
	}

	.close-icon {
		width: 20px;
		height: 20px;
	}

	.modal-content {
		padding: 1.5rem;
	}

	.modal-content p {
		color: var(--text-secondary);
		margin-bottom: 1.5rem;
		line-height: 1.6;
	}

	.form-group {
		margin-bottom: 1.5rem;
	}

	.form-group label {
		display: block;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.form-select,
	.form-textarea {
		width: 100%;
		padding: 0.75rem;
		border: 1px solid var(--border-color);
		border-radius: 8px;
		font-size: 1rem;
		background: var(--card-bg);
		color: var(--text-primary);
	}

	.form-select:focus,
	.form-textarea:focus {
		outline: none;
		border-color: var(--primary-color);
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	.modal-actions {
		display: flex;
		gap: 1rem;
		justify-content: flex-end;
		margin-top: 2rem;
	}

	@media (max-width: 768px) {
		.page-header h1 {
			font-size: 2rem;
		}

		.overview-header {
			flex-direction: column;
			align-items: flex-start;
			gap: 1rem;
		}

		.overview-details {
			grid-template-columns: 1fr;
		}

		.action-buttons {
			flex-direction: column;
		}

		.table-header,
		.table-row {
			grid-template-columns: 1fr;
			gap: 0.5rem;
		}

		.table-header {
			display: none;
		}

		.table-row {
			border: 1px solid var(--border-color);
			border-radius: 8px;
			margin-bottom: 1rem;
		}

		.modal-actions {
			flex-direction: column;
		}
	}
</style> 
