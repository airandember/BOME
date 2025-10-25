<script lang="ts">
	import { onMount } from 'svelte';
	import { subscriberElasticService, type DiagnosticData, type UnifiedSubscriber } from '$lib/services/subscriber-elastic-service';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	// State
	let diagnosticData: DiagnosticData | null = $state(null);
	let loading = $state(false);
	let error = $state<string | null>(null);
	let selectedIssue: 'multiple_customers' | 'no_access' | 'manual_access' | null = $state(null);
	let selectedSubscriber: UnifiedSubscriber | null = $state(null);

	// Load diagnostic data on mount
	onMount(async () => {
		await loadDiagnosticData();
	});

	async function loadDiagnosticData() {
		loading = true;
		error = null;
		try {
			diagnosticData = await subscriberElasticService.getDiagnosticData();
			console.log('🔍 Diagnostic data loaded:', diagnosticData);
		} catch (err: any) {
			error = err.message || 'Failed to load diagnostic data';
			console.error('Error loading diagnostic data:', err);
		} finally {
			loading = false;
		}
	}

	function selectIssue(issue: 'multiple_customers' | 'no_access' | 'manual_access') {
		selectedIssue = selectedIssue === issue ? null : issue;
		selectedSubscriber = null;
	}

	function selectSubscriber(subscriber: UnifiedSubscriber) {
		selectedSubscriber = selectedSubscriber?.id === subscriber.id ? null : subscriber;
	}

	async function updateManualAccess(subscriber: UnifiedSubscriber, hasAccess: boolean) {
		try {
			await subscriberElasticService.updateManualVideoAccess(subscriber.id, hasAccess);
			showToast(`Manual access ${hasAccess ? 'granted' : 'revoked'} for ${subscriber.full_name}`, 'success');
			
			// Reload diagnostic data to reflect changes
			await loadDiagnosticData();
		} catch (err: any) {
			showToast(`Failed to update access: ${err.message}`, 'error');
		}
	}

	function formatCurrency(amount: number, currency: string = 'USD'): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: currency.toUpperCase()
		}).format(amount);
	}

	function formatDate(dateString: string | undefined): string {
		if (!dateString) return 'N/A';
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}
</script>

<div class="subscriber-diagnostic-page">
	<div class="page-header">
		<h1>🔍 Subscriber Diagnostic Dashboard</h1>
		<p>Unified view of subscriber data integrity and access issues</p>
		<button class="refresh-btn" onclick={loadDiagnosticData} disabled={loading}>
			{#if loading}
				<LoadingSpinner size="small" />
			{:else}
				🔄
			{/if}
			Refresh
		</button>
	</div>

	{#if loading}
		<div class="loading-container">
			<LoadingSpinner />
			<p>Loading diagnostic data...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<h3>❌ Error Loading Data</h3>
			<p>{error}</p>
			<button class="retry-btn" onclick={loadDiagnosticData}>
				🔄 Retry
			</button>
		</div>
	{:else if diagnosticData}
		<!-- Summary Cards -->
		<div class="summary-grid">
			<div class="summary-card">
				<div class="card-icon">👥</div>
				<div class="card-content">
					<div class="card-title">Total Subscribers</div>
					<div class="card-value">{diagnosticData.summary.total_subscribers.toLocaleString()}</div>
				</div>
			</div>
			
			<div class="summary-card">
				<div class="card-icon">⚠️</div>
				<div class="card-content">
					<div class="card-title">Issues Found</div>
					<div class="card-value">{diagnosticData.summary.issues_found}</div>
				</div>
			</div>
			
			<div class="summary-card">
				<div class="card-icon">🎬</div>
				<div class="card-content">
					<div class="card-title">Manual Overrides</div>
					<div class="card-value">{diagnosticData.summary.manual_overrides}</div>
				</div>
			</div>
			
			<div class="summary-card">
				<div class="card-icon">💰</div>
				<div class="card-content">
					<div class="card-title">Total MRR</div>
					<div class="card-value">{formatCurrency(diagnosticData.statistics.total_mrr)}</div>
				</div>
			</div>
		</div>

		<!-- Issues Section -->
		<div class="issues-section">
			<h2>🚨 Data Integrity Issues</h2>
			
			<!-- Multiple Stripe Customers -->
			<div class="issue-card">
				<div class="issue-header" onclick={() => selectIssue('multiple_customers')}>
					<div class="issue-info">
						<h3>🔗 Multiple Stripe Customers</h3>
						<p>{diagnosticData.issues.multiple_stripe_customers.description}</p>
						<span class="issue-count">{diagnosticData.issues.multiple_stripe_customers.count} subscribers</span>
					</div>
					<div class="issue-toggle">
						{selectedIssue === 'multiple_customers' ? '▼' : '▶'}
					</div>
				</div>
				
				{#if selectedIssue === 'multiple_customers'}
					<div class="issue-content">
						{#each diagnosticData.issues.multiple_stripe_customers.subscribers as subscriber}
							<div class="subscriber-item" onclick={() => selectSubscriber(subscriber)}>
								<div class="subscriber-info">
									<div class="subscriber-name">{subscriber.full_name}</div>
									<div class="subscriber-email">{subscriber.email}</div>
									<div class="subscriber-details">
										<span class="stripe-count">{subscriber.stripe_customer_ids.length} Stripe IDs</span>
										<span class="plan-info">Plan: {subscriber.plan_name || 'None'}</span>
									</div>
								</div>
								<div class="subscriber-actions">
									<button class="btn btn-sm btn-outline" onclick={(e) => { e.stopPropagation(); }}>
										🔍 Investigate
									</button>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Active Plan No Access -->
			<div class="issue-card">
				<div class="issue-header" onclick={() => selectIssue('no_access')}>
					<div class="issue-info">
						<h3>🚫 Active Plan No Access</h3>
						<p>{diagnosticData.issues.active_plan_no_access.description}</p>
						<span class="issue-count">{diagnosticData.issues.active_plan_no_access.count} subscribers</span>
					</div>
					<div class="issue-toggle">
						{selectedIssue === 'no_access' ? '▼' : '▶'}
					</div>
				</div>
				
				{#if selectedIssue === 'no_access'}
					<div class="issue-content">
						{#each diagnosticData.issues.active_plan_no_access.subscribers as subscriber}
							<div class="subscriber-item" onclick={() => selectSubscriber(subscriber)}>
								<div class="subscriber-info">
									<div class="subscriber-name">{subscriber.full_name}</div>
									<div class="subscriber-email">{subscriber.email}</div>
									<div class="subscriber-details">
										<span class="plan-status">Plan: {subscriber.plan_name} ({subscriber.plan_status})</span>
										<span class="access-status">Access: {subscriber.has_video_access ? '✅' : '❌'}</span>
									</div>
								</div>
								<div class="subscriber-actions">
									<button 
										class="btn btn-sm btn-success" 
										onclick={(e) => { e.stopPropagation(); updateManualAccess(subscriber, true); }}
									>
										✅ Grant Access
									</button>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Manual Access -->
			<div class="issue-card">
				<div class="issue-header" onclick={() => selectIssue('manual_access')}>
					<div class="issue-info">
						<h3>🎬 Manual Video Access</h3>
						<p>{diagnosticData.manual_overrides.description}</p>
						<span class="issue-count">{diagnosticData.manual_overrides.count} subscribers</span>
					</div>
					<div class="issue-toggle">
						{selectedIssue === 'manual_access' ? '▼' : '▶'}
					</div>
				</div>
				
				{#if selectedIssue === 'manual_access'}
					<div class="issue-content">
						{#each diagnosticData.manual_overrides.subscribers as subscriber}
							<div class="subscriber-item" onclick={() => selectSubscriber(subscriber)}>
								<div class="subscriber-info">
									<div class="subscriber-name">{subscriber.full_name}</div>
									<div class="subscriber-email">{subscriber.email}</div>
									<div class="subscriber-details">
										<span class="plan-status">Plan: {subscriber.plan_name || 'None'}</span>
										<span class="access-status">Manual Access: ✅</span>
									</div>
								</div>
								<div class="subscriber-actions">
									<button 
										class="btn btn-sm btn-warning" 
										onclick={(e) => { e.stopPropagation(); updateManualAccess(subscriber, false); }}
									>
										🚫 Revoke Access
									</button>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		<!-- Selected Subscriber Details -->
		{#if selectedSubscriber}
			<div class="subscriber-details-modal">
				<div class="modal-header">
					<h3>👤 {selectedSubscriber.full_name}</h3>
					<button class="close-btn" onclick={() => selectedSubscriber = null}>✕</button>
				</div>
				
				<div class="modal-content">
					<div class="details-grid">
						<div class="detail-section">
							<h4>📋 Basic Information</h4>
							<div class="detail-item">
								<span class="label">Email:</span>
								<span class="value">{selectedSubscriber.email}</span>
							</div>
							<div class="detail-item">
								<span class="label">Role:</span>
								<span class="value">{selectedSubscriber.role}</span>
							</div>
							<div class="detail-item">
								<span class="label">Email Verified:</span>
								<span class="value">{selectedSubscriber.email_verified ? '✅' : '❌'}</span>
							</div>
						</div>

						<div class="detail-section">
							<h4>💳 Subscription Information</h4>
							<div class="detail-item">
								<span class="label">Plan Name:</span>
								<span class="value">{selectedSubscriber.plan_name || 'None'}</span>
							</div>
							<div class="detail-item">
								<span class="label">Plan Type:</span>
								<span class="value">{selectedSubscriber.plan_type}</span>
							</div>
							<div class="detail-item">
								<span class="label">Plan Status:</span>
								<span class="value">{selectedSubscriber.plan_status}</span>
							</div>
							<div class="detail-item">
								<span class="label">Plan Price:</span>
								<span class="value">{formatCurrency(selectedSubscriber.plan_price, selectedSubscriber.plan_currency)}</span>
							</div>
						</div>

						<div class="detail-section">
							<h4>🔐 Access Control</h4>
							<div class="detail-item">
								<span class="label">Has Active Plan:</span>
								<span class="value">{selectedSubscriber.has_active_plan ? '✅' : '❌'}</span>
							</div>
							<div class="detail-item">
								<span class="label">Has Video Access:</span>
								<span class="value">{selectedSubscriber.has_video_access ? '✅' : '❌'}</span>
							</div>
							<div class="detail-item">
								<span class="label">Manual Access:</span>
								<span class="value">{selectedSubscriber.manual_access_granted ? '✅' : '❌'}</span>
							</div>
						</div>

						<div class="detail-section">
							<h4>💳 Stripe Information</h4>
							<div class="detail-item">
								<span class="label">Stripe Customer IDs:</span>
								<span class="value">{selectedSubscriber.stripe_customer_ids.join(', ') || 'None'}</span>
							</div>
							<div class="detail-item">
								<span class="label">Subscription ID:</span>
								<span class="value">{selectedSubscriber.subscription_id || 'None'}</span>
							</div>
						</div>

						<div class="detail-section">
							<h4>📊 Business Intelligence</h4>
							<div class="detail-item">
								<span class="label">MRR Contribution:</span>
								<span class="value">{formatCurrency(selectedSubscriber.mrr_contribution)}</span>
							</div>
							<div class="detail-item">
								<span class="label">ARR Contribution:</span>
								<span class="value">{formatCurrency(selectedSubscriber.arr_contribution)}</span>
							</div>
							<div class="detail-item">
								<span class="label">LTV Estimate:</span>
								<span class="value">{formatCurrency(selectedSubscriber.ltv_estimate)}</span>
							</div>
						</div>
					</div>
				</div>
			</div>
		{/if}
	{/if}
</div>

<style>
	.subscriber-diagnostic-page {
		padding: 1.5rem;
		max-width: 1200px;
		margin: 0 auto;
	}

	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 2rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid #e5e7eb;
	}

	.page-header h1 {
		font-size: 1.875rem;
		font-weight: 700;
		color: #111827;
		margin: 0 0 0.5rem 0;
	}

	.page-header p {
		color: #6b7280;
		margin: 0;
	}

	.refresh-btn {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		background: #3b82f6;
		color: white;
		border: none;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: background-color 0.2s ease;
	}

	.refresh-btn:hover:not(:disabled) {
		background: #2563eb;
	}

	.refresh-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.summary-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 1rem;
		margin-bottom: 2rem;
	}

	.summary-card {
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		padding: 1.5rem;
		display: flex;
		align-items: center;
		gap: 1rem;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
	}

	.card-icon {
		font-size: 2rem;
		width: 3rem;
		height: 3rem;
		display: flex;
		align-items: center;
		justify-content: center;
		background: #f3f4f6;
		border-radius: 0.5rem;
	}

	.card-content {
		flex: 1;
	}

	.card-title {
		font-size: 0.875rem;
		font-weight: 500;
		color: #6b7280;
		margin-bottom: 0.25rem;
	}

	.card-value {
		font-size: 1.5rem;
		font-weight: 700;
		color: #111827;
	}

	.issues-section {
		margin-bottom: 2rem;
	}

	.issues-section h2 {
		font-size: 1.5rem;
		font-weight: 600;
		color: #111827;
		margin-bottom: 1rem;
	}

	.issue-card {
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		margin-bottom: 1rem;
		overflow: hidden;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
	}

	.issue-header {
		padding: 1.5rem;
		cursor: pointer;
		display: flex;
		justify-content: space-between;
		align-items: center;
		background: #f9fafb;
		border-bottom: 1px solid #e5e7eb;
		transition: background-color 0.2s ease;
	}

	.issue-header:hover {
		background: #f3f4f6;
	}

	.issue-info h3 {
		font-size: 1.125rem;
		font-weight: 600;
		color: #111827;
		margin: 0 0 0.5rem 0;
	}

	.issue-info p {
		color: #6b7280;
		margin: 0 0 0.5rem 0;
		font-size: 0.875rem;
	}

	.issue-count {
		background: #3b82f6;
		color: white;
		padding: 0.25rem 0.75rem;
		border-radius: 9999px;
		font-size: 0.75rem;
		font-weight: 500;
	}

	.issue-toggle {
		font-size: 1.25rem;
		color: #6b7280;
		transition: transform 0.2s ease;
	}

	.issue-content {
		padding: 1rem;
	}

	.subscriber-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem;
		border: 1px solid #e5e7eb;
		border-radius: 0.375rem;
		margin-bottom: 0.5rem;
		cursor: pointer;
		transition: background-color 0.2s ease;
	}

	.subscriber-item:hover {
		background: #f9fafb;
	}

	.subscriber-info {
		flex: 1;
	}

	.subscriber-name {
		font-weight: 600;
		color: #111827;
		margin-bottom: 0.25rem;
	}

	.subscriber-email {
		color: #6b7280;
		font-size: 0.875rem;
		margin-bottom: 0.5rem;
	}

	.subscriber-details {
		display: flex;
		gap: 1rem;
		font-size: 0.75rem;
	}

	.subscriber-details span {
		background: #f3f4f6;
		padding: 0.25rem 0.5rem;
		border-radius: 0.25rem;
		color: #374151;
	}

	.subscriber-actions {
		display: flex;
		gap: 0.5rem;
	}

	.btn {
		padding: 0.375rem 0.75rem;
		border: none;
		border-radius: 0.375rem;
		font-size: 0.75rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.btn-sm {
		padding: 0.25rem 0.5rem;
		font-size: 0.75rem;
	}

	.btn-outline {
		background: transparent;
		color: #3b82f6;
		border: 1px solid #3b82f6;
	}

	.btn-outline:hover {
		background: #3b82f6;
		color: white;
	}

	.btn-success {
		background: #059669;
		color: white;
	}

	.btn-success:hover {
		background: #047857;
	}

	.btn-warning {
		background: #d97706;
		color: white;
	}

	.btn-warning:hover {
		background: #b45309;
	}

	.subscriber-details-modal {
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

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1.5rem;
		border-bottom: 1px solid #e5e7eb;
		background: #f9fafb;
		border-radius: 0.5rem 0.5rem 0 0;
	}

	.modal-header h3 {
		margin: 0;
		color: #111827;
		font-size: 1.25rem;
		font-weight: 600;
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

	.modal-content {
		background: white;
		border-radius: 0.5rem;
		box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
		max-width: 90vw;
		width: 800px;
		max-height: 90vh;
		overflow-y: auto;
	}

	.details-grid {
		padding: 1.5rem;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: 1.5rem;
	}

	.detail-section h4 {
		font-size: 1rem;
		font-weight: 600;
		color: #111827;
		margin: 0 0 1rem 0;
		padding-bottom: 0.5rem;
		border-bottom: 1px solid #e5e7eb;
	}

	.detail-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.5rem 0;
		border-bottom: 1px solid #f3f4f6;
	}

	.detail-item:last-child {
		border-bottom: none;
	}

	.detail-item .label {
		font-size: 0.875rem;
		font-weight: 500;
		color: #6b7280;
	}

	.detail-item .value {
		font-size: 0.875rem;
		color: #111827;
		font-weight: 500;
		text-align: right;
	}

	.loading-container, .error-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 4rem 2rem;
		text-align: center;
	}

	.error-container h3 {
		color: #dc2626;
		margin: 0 0 0.5rem 0;
	}

	.error-container p {
		color: #6b7280;
		margin: 0 0 1.5rem 0;
	}

	.retry-btn {
		padding: 0.5rem 1rem;
		background: #dc2626;
		color: white;
		border: none;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: background-color 0.2s ease;
	}

	.retry-btn:hover {
		background: #b91c1c;
	}

	@media (max-width: 768px) {
		.subscriber-diagnostic-page {
			padding: 1rem;
		}

		.page-header {
			flex-direction: column;
			gap: 1rem;
			align-items: stretch;
		}

		.summary-grid {
			grid-template-columns: 1fr;
		}

		.details-grid {
			grid-template-columns: 1fr;
		}

		.modal-content {
			width: 95%;
			margin: 1rem;
		}
	}
</style>
