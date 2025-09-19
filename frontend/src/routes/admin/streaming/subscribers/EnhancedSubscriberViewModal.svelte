<script lang="ts">
	import type { EnhancedSubscriber } from '$lib/types/enhanced-subscriber';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	export let isOpen = false;
	export let subscriber: EnhancedSubscriber | null = null;
	export let onClose: () => void = () => {};

	// Format currency
	function formatCurrency(amount: number, currency: string = 'USD'): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: currency.toUpperCase()
		}).format(amount);
	}

	// Format date
	function formatDate(dateString: string | null): string {
		if (!dateString) return 'N/A';
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	// Format boolean as badge
	function formatBoolean(value: boolean): { text: string; class: string } {
		return value 
			? { text: '✅ Yes', class: 'badge-success' }
			: { text: '❌ No', class: 'badge-error' };
	}

	// Close modal on escape key
	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			onClose();
		}
	}

	// Close modal when clicking outside
	function handleBackdropClick(event: MouseEvent) {
		if (event.target === event.currentTarget) {
			onClose();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if isOpen && subscriber}
	<div class="modal-overlay" onclick={handleBackdropClick} role="dialog" aria-modal="true" tabindex="-1">
		<div class="modal-content">
			<!-- Header -->
			<div class="modal-header">
				<div class="header-info">
					<h2>👤 Subscriber Details</h2>
					<p>Comprehensive view of subscriber information</p>
				</div>
				<button 
					type="button" 
					class="close-btn" 
					onclick={onClose}
					aria-label="Close modal"
				>
					✕
				</button>
			</div>

			<!-- Content -->
			<div class="modal-body">
				<!-- Basic Information -->
				<section class="info-section">
					<h3>📋 Basic Information</h3>
					<div class="info-grid">
						<div class="info-item">
							<div class="info-label">Full Name</div>
							<span class="value">{subscriber.full_name || 'N/A'}</span>
						</div>
						<div class="info-item">
							<div class="info-label">Email</div>
							<span class="value email">{subscriber.email}</span>
						</div>
						<div class="info-item">
							<div class="info-label">Role</div>
							<span class="value role-badge role-{subscriber.role}">{subscriber.role}</span>
						</div>
						<div class="info-item">
							<div class="info-label">Email Verified</div>
							<span class="value badge {formatBoolean(subscriber.email_verified).class}">
								{formatBoolean(subscriber.email_verified).text}
							</span>
						</div>
					</div>
				</section>

				<!-- Plan Information -->
				<section class="info-section">
					<h3>💳 Plan Information</h3>
					<div class="info-grid">
						<div class="info-item">
							<div class="info-label">Plan Name</div>
							<span class="value plan-name">{subscriber.plan_name || 'No Plan'}</span>
						</div>
						<div class="info-item">
							<div class="info-label">Plan Type</div>
							<span class="value plan-type plan-type-{subscriber.plan_type}">{subscriber.plan_type}</span>
						</div>
						<div class="info-item">
							<div class="info-label">Legacy Status</div>
							<span class="value legacy-status legacy-{subscriber.plan_legacy_status.toLowerCase()}">
								{subscriber.plan_legacy_status}
							</span>
						</div>
						<div class="info-item">
							<div class="info-label">Plan Status</div>
							<span class="value plan-status status-{subscriber.plan_status}">{subscriber.plan_status}</span>
						</div>
						<div class="info-item">
							<div class="info-label">Plan Price</div>
							<span class="value price">{formatCurrency(subscriber.plan_price, subscriber.plan_currency)}</span>
						</div>
						<div class="info-item">
							<div class="info-label">Currency</div>
							<span class="value currency">{subscriber.plan_currency?.toUpperCase() || 'N/A'}</span>
						</div>
					</div>
				</section>

				<!-- Access & Status -->
				<section class="info-section">
					<h3>🔐 Access & Status</h3>
					<div class="info-grid">
						<div class="info-item">
							<div class="info-label">Active Plan</div>
							<span class="value badge {formatBoolean(subscriber.has_active_plan).class}">
								{formatBoolean(subscriber.has_active_plan).text}
							</span>
						</div>
						<div class="info-item">
							<div class="info-label">Video Access</div>
							<span class="value badge {formatBoolean(subscriber.has_video_access).class}">
								{formatBoolean(subscriber.has_video_access).text}
							</span>
						</div>
						<div class="info-item">
							<div class="info-label">Manual Override</div>
							<span class="value badge {subscriber.manual_access_granted ? 'badge-success' : 'badge-secondary'}">
								{subscriber.manual_access_granted ? '✅ Active' : '❌ Off'}
							</span>
						</div>
						<div class="info-item">
							<div class="info-label">Account Status</div>
							<span class="value account-status status-{subscriber.account_status}">
								{subscriber.account_status}
							</span>
						</div>
					</div>
				</section>

				<!-- Date Information -->
				<section class="info-section">
					<h3>📅 Important Dates</h3>
					<div class="info-grid">
						<div class="info-item">
							<div class="info-label">Plan Start Date</div>
							<span class="value date">{formatDate(subscriber.plan_start_date)}</span>
						</div>
						<div class="info-item">
							<div class="info-label">Billing Period Start</div>
							<span class="value date">{formatDate(subscriber.billing_period_start)}</span>
						</div>
						<div class="info-item">
							<div class="info-label">Billing Period End</div>
							<span class="value date">{formatDate(subscriber.billing_period_end)}</span>
						</div>
						<div class="info-item">
							<div class="info-label">Last Login</div>
							<span class="value date">{formatDate(subscriber.last_login)}</span>
						</div>
						<div class="info-item">
							<div class="info-label">Account Created</div>
							<span class="value date">{formatDate(subscriber.created_at)}</span>
						</div>
						<div class="info-item">
							<div class="info-label">Days Until Expiry</div>
							<span class="value days-left">
								{subscriber.days_until_expiry !== null ? `${subscriber.days_until_expiry} days` : 'N/A'}
							</span>
						</div>
					</div>
				</section>

				<!-- Business Intelligence -->
				<section class="info-section">
					<h3>📊 Business Intelligence</h3>
					<div class="info-grid">
						<div class="info-item">
							<div class="info-label">MRR Contribution</div>
							<span class="value mrr">{formatCurrency(subscriber.mrr_contribution, subscriber.plan_currency)}</span>
						</div>
						<div class="info-item">
							<div class="info-label">LTV Estimate</div>
							<span class="value ltv">{formatCurrency(subscriber.ltv_estimate, subscriber.plan_currency)}</span>
						</div>
						<div class="info-item">
							<div class="info-label">Account Age</div>
							<span class="value age">{subscriber.account_age_days || 0} days</span>
						</div>
						<div class="info-item">
							<div class="info-label">Billing Cycle Length</div>
							<span class="value cycle">{subscriber.billing_cycle_length || 0} days</span>
						</div>
					</div>
				</section>

				<!-- Stripe Information -->
				{#if subscriber.stripe_customer_id || subscriber.stripe_subscription_id}
					<section class="info-section">
						<h3>💳 Stripe Information</h3>
						<div class="info-grid">
							{#if subscriber.stripe_customer_id}
								<div class="info-item">
									<div class="info-label">Stripe Customer ID</div>
									<span class="value stripe-id">{subscriber.stripe_customer_id}</span>
								</div>
							{/if}
							{#if subscriber.stripe_subscription_id}
								<div class="info-item">
									<div class="info-label">Stripe Subscription ID</div>
									<span class="value stripe-id">{subscriber.stripe_subscription_id}</span>
								</div>
							{/if}
						</div>
					</section>
				{/if}
			</div>

			<!-- Footer -->
			<div class="modal-footer">
				<button type="button" class="btn btn-secondary" onclick={onClose}>
					Close
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
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
		backdrop-filter: blur(4px);
	}

	.modal-content {
		background: white;
		border-radius: 1rem;
		box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
		max-width: 85vw;
		width: 90%;
		max-height: 95vh;
		overflow-y: auto;
		border: 1px solid #e5e7eb;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		padding: 1.5rem;
		border-bottom: 1px solid #e5e7eb;
		background: #f9fafb;
		border-radius: 1rem 1rem 0 0;
	}

	.header-info h2 {
		margin: 0 0 0.25rem 0;
		color: #111827;
		font-size: 1.5rem;
		font-weight: 600;
	}

	.header-info p {
		margin: 0;
		color: #6b7280;
		font-size: 0.875rem;
	}

	.close-btn {
		background: none;
		border: none;
		font-size: 1.5rem;
		cursor: pointer;
		color: #6b7280;
		padding: 0;
		width: 32px;
		height: 32px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 0.375rem;
		transition: all 0.2s ease;
	}

	.close-btn:hover {
		background: #e5e7eb;
		color: #111827;
	}

	.modal-body {
		padding: 1.5rem;
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.info-section {
		border: 1px solid #e5e7eb;
		border-radius: 0.75rem;
		overflow: hidden;
	}

	.info-section h3 {
		margin: 0;
		padding: 1rem 1.5rem;
		background: #f9fafb;
		border-bottom: 1px solid #e5e7eb;
		font-size: 1.125rem;
		font-weight: 600;
		color: #111827;
	}

	.info-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 1rem;
		padding: 1.5rem;
	}

	.info-item {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.info-item .info-label {
		font-size: 0.875rem;
		font-weight: 500;
		color: #6b7280;
		text-transform: uppercase;
		letter-spacing: 0.025em;
	}

	.info-item .value {
		font-size: 1rem;
		color: #111827;
		font-weight: 500;
		word-break: break-word;
	}

	/* Value type styles */
	.email {
		color: #2563eb;
		font-family: 'Courier New', monospace;
	}

	.stripe-id {
		font-family: 'Courier New', monospace;
		font-size: 0.875rem;
		color: #6366f1;
	}

	.price, .mrr, .ltv {
		color: #059669;
		font-weight: 600;
	}

	.date {
		color: #7c3aed;
	}

	.days-left {
		color: #dc2626;
		font-weight: 600;
	}

	/* Badge styles */
	.badge {
		display: inline-block;
		padding: 0.25rem 0.75rem;
		border-radius: 9999px;
		font-size: 0.875rem;
		font-weight: 500;
		text-align: center;
	}

	.badge-success {
		background: #dcfce7;
		color: #166534;
	}

	.badge-error {
		background: #fee2e2;
		color: #991b1b;
	}

	.badge-secondary {
		background: #f3f4f6;
		color: #6b7280;
	}

	/* Role badges */
	.role-badge {
		display: inline-block;
		padding: 0.25rem 0.75rem;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		text-transform: capitalize;
	}

	.role-super_admin {
		background: #fef3c7;
		color: #92400e;
	}

	.role-admin {
		background: #dbeafe;
		color: #1e40af;
	}

	.role-user {
		background: #f3f4f6;
		color: #374151;
	}

	/* Plan type styles */
	.plan-type-premium {
		color: #7c3aed;
		font-weight: 600;
	}

	.plan-type-basic {
		color: #059669;
		font-weight: 600;
	}

	.plan-type-none {
		color: #6b7280;
		font-style: italic;
	}

	/* Legacy status */
	.legacy-legacy {
		color: #d97706;
		font-weight: 600;
	}

	.legacy-current {
		color: #059669;
		font-weight: 600;
	}

	.legacy-unknown {
		color: #6b7280;
		font-style: italic;
	}

	/* Status styles */
	.status-active {
		color: #059669;
		font-weight: 600;
	}

	.status-expired, .status-cancelled {
		color: #dc2626;
		font-weight: 600;
	}

	.status-trial {
		color: #d97706;
		font-weight: 600;
	}

	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: 1rem;
		padding: 1.5rem;
		border-top: 1px solid #e5e7eb;
		background: #f9fafb;
		border-radius: 0 0 1rem 1rem;
	}

	.btn {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 0.5rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.btn-secondary {
		background: #f3f4f6;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-secondary:hover {
		background: #e5e7eb;
		border-color: #9ca3af;
	}

	/* Responsive design */
	@media (max-width: 768px) {
		.modal-content {
			width: 95%;
			margin: 1rem;
		}

		.info-grid {
			grid-template-columns: 1fr;
		}

		.modal-header {
			flex-direction: column;
			gap: 1rem;
			align-items: stretch;
		}
	}
</style>
