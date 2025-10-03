<script lang="ts">
	import { StripeSyncService, type SyncStatus } from '$lib/services/stripe-sync';
	import { showToast } from '$lib/toast';

	// Props
	const { 
		isOpen = false, 
		onClose = () => {}, 
		onSuccess = () => {} 
	} = $props<{
		isOpen: boolean;
		onClose: () => void;
		onSuccess: () => void;
	}>();

	// State
	let syncing = $state(false);
	let syncStatus = $state<SyncStatus | null>(null);
	let selectedSyncType = $state<'incremental' | 'initial'>('incremental');
	let customSinceDate = $state('');
	let progressMessage = $state('');
	let errorMessage = $state('');

	// Reactive
	const isInitialSync = $derived(selectedSyncType === 'initial');
	const hasActiveSyncJob = $derived(syncStatus?.sync_running || false);
	const currentProgress = $derived(syncStatus?.current_job?.progress_percent || 0);

	// Load sync status when modal opens
	$effect(() => {
		if (isOpen && !syncing) {
			errorMessage = ''; // Clear any previous errors
			loadSyncStatus();
		}
	});

	async function loadSyncStatus() {
		try {
			syncStatus = await StripeSyncService.getSyncStatus();
			errorMessage = ''; // Clear error on success
		} catch (error: any) {
			console.error('Failed to load sync status:', error);
			errorMessage = `🚨 Stripe Sync Service Not Available

Error: ${error.message}

Expected Endpoint: /admin/streaming/stripe/sync/status
Full URL: https://watch.bookofmormonevidence.org/bome-backend/api/v1/admin/streaming/stripe/sync/status

DIAGNOSIS:
The Stripe sync routes are not responding in production. This could mean:

1. ❌ Stripe sync routes not registered in production backend
2. ❌ Database connection issues preventing route registration  
3. ❌ Missing middleware or authentication issues
4. ❌ Backend compilation/deployment issues

IMMEDIATE ACTIONS:
1. Check backend logs for route registration messages
2. Verify database connectivity in production
3. Confirm all Stripe services are initialized
4. Check if /admin/streaming routes are working

TEMPORARY WORKAROUND:
For now, you can sync data manually using the subscription plans interface or direct database queries until the sync service is restored.`;
		}
	}

	async function startSync() {
		if (syncing || hasActiveSyncJob) return;

		try {
			syncing = true;
			progressMessage = 'Starting sync...';

			let response;
			if (selectedSyncType === 'initial') {
				response = await StripeSyncService.triggerInitialSync();
				showToast('Full sync started - this may take several minutes', 'info');
			} else {
				const since = customSinceDate || undefined;
				response = await StripeSyncService.triggerIncrementalSync(since);
				showToast('Incremental sync started', 'success');
			}

			// Start polling for progress
			progressMessage = 'Syncing data...';
			await StripeSyncService.pollSyncStatus(
				(status) => {
					syncStatus = status;
					if (status.current_job) {
						const progress = Math.round(status.current_job.progress_percent);
						progressMessage = `Processing ${status.current_job.entity_type}: ${progress}% (${status.current_job.processed_items}/${status.current_job.total_items})`;
					}
				},
				2000, // Poll every 2 seconds
				300   // 10 minute timeout
			);

			// Sync completed
			progressMessage = 'Sync completed successfully!';
			showToast('Stripe data sync completed successfully', 'success');
			onSuccess();
			
			// Close modal after a brief delay
			setTimeout(() => {
				onClose();
			}, 2000);

		} catch (error: any) {
			console.error('Sync failed:', error);
			progressMessage = '';
			errorMessage = `Sync failed: ${error.message}. Please report this error to technical support.`;
			showToast(`Sync failed: ${error.message}`, 'error');
		} finally {
			syncing = false;
		}
	}

	function handleClose() {
		if (!syncing) {
			onClose();
		}
	}

	// Format date for display
	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleString();
	}
</script>

{#if isOpen}
	<div 
		class="modal-overlay" 
		onclick={handleClose}
		onkeydown={(e) => e.key === 'Escape' && handleClose()}
		role="dialog"
		aria-modal="true"
		tabindex="-1"
	>
		<div 
			class="modal-content" 
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.stopPropagation()}
			role="document"
		>
			<div class="modal-header">
				<h2>🔄 Resync Database</h2>
				<button 
					class="close-btn" 
					onclick={handleClose}
					disabled={syncing}
				>
					×
				</button>
			</div>

			<div class="modal-body">
				{#if errorMessage}
					<!-- Error Display -->
					<div class="error-section">
						<div class="error-header">
							<h3>❌ Error</h3>
						</div>
						<div class="error-content">
							<p>{errorMessage}</p>
							<div class="error-actions">
								<button class="btn btn-secondary" onclick={loadSyncStatus}>
									🔄 Retry Connection
								</button>
								<button class="btn btn-primary" onclick={() => { errorMessage = ''; }}>
									✕ Dismiss
								</button>
							</div>
						</div>
					</div>
				{:else if hasActiveSyncJob && syncStatus?.current_job}
					<!-- Active sync in progress -->
					<div class="sync-active">
						<div class="sync-info">
							<h3>🚀 Sync In Progress</h3>
							<p>Type: <strong>{syncStatus.current_job.type}</strong></p>
							<p>Entity: <strong>{syncStatus.current_job.entity_type}</strong></p>
							<p>Started: <strong>{formatDate(syncStatus.current_job.started_at)}</strong></p>
						</div>

						<div class="progress-container">
							<div class="progress-bar">
								<div 
									class="progress-fill" 
									style="width: {currentProgress}%"
								></div>
							</div>
							<div class="progress-text">
								{Math.round(currentProgress)}% - {syncStatus.current_job.processed_items}/{syncStatus.current_job.total_items} items
							</div>
						</div>

						<div class="sync-actions">
							<button class="btn btn-secondary" onclick={loadSyncStatus}>
								🔄 Refresh Status
							</button>
						</div>
					</div>
				{:else if syncing}
					<!-- Starting sync -->
					<div class="sync-starting">
						<div class="loading-spinner"></div>
						<h3>Starting Sync...</h3>
						<p>{progressMessage}</p>
						
						{#if syncStatus?.current_job}
							<div class="progress-container">
								<div class="progress-bar">
									<div 
										class="progress-fill" 
										style="width: {currentProgress}%"
									></div>
								</div>
								<div class="progress-text">
									{Math.round(currentProgress)}% - {progressMessage}
								</div>
							</div>
						{/if}
					</div>
				{:else}
					<!-- Sync configuration -->
					<div class="sync-config">
						<div class="sync-type-selection">
							<h3>Select Sync Type</h3>
							
							<label class="radio-option">
								<input 
									type="radio" 
									bind:group={selectedSyncType} 
									value="incremental"
								>
								<div class="radio-content">
									<strong>⚡ Incremental Sync (Recommended)</strong>
									<p>Syncs recent changes (last 24 hours by default). Fast and efficient for regular updates.</p>
								</div>
							</label>

							<label class="radio-option">
								<input 
									type="radio" 
									bind:group={selectedSyncType} 
									value="initial"
								>
								<div class="radio-content">
									<strong>🔄 Full Sync</strong>
									<p>Complete sync of all Stripe data (1.5 years). Use for initial setup or major data issues.</p>
								</div>
							</label>
						</div>

						{#if !isInitialSync}
							<div class="date-selection">
								<label for="since-date">
									<strong>Custom Start Date (Optional)</strong>
									<input 
										id="since-date"
										type="date" 
										bind:value={customSinceDate}
										max={new Date().toISOString().split('T')[0]}
									>
									<small>Leave empty to sync last 24 hours</small>
								</label>
							</div>
						{/if}

						<div class="sync-info">
							<h4>What will be synced to your database:</h4>
							<ul>
								<li>✅ Stripe customers → Users table linking</li>
								<li>✅ Subscription data and status updates</li>
								<li>✅ Product and pricing information</li>
								<li>✅ Payment and invoice records</li>
								<li>✅ Database relationships and integrity</li>
							</ul>
						</div>

						{#if syncStatus?.recent_jobs && syncStatus.recent_jobs.length > 0}
							<div class="recent-jobs">
								<h4>Recent Sync Jobs</h4>
								<div class="jobs-list">
									{#each syncStatus.recent_jobs.slice(0, 3) as job}
										<div class="job-item status-{job.status}">
											<div class="job-info">
												<strong>{job.type}</strong> - {job.entity_type}
											</div>
											<div class="job-status">
												{#if job.status === 'completed'}
													✅ Completed
												{:else if job.status === 'running'}
													🔄 Running
												{:else if job.status === 'failed'}
													❌ Failed
												{/if}
											</div>
											<div class="job-date">
												{formatDate(job.started_at)}
											</div>
										</div>
									{/each}
								</div>
							</div>
						{/if}
					</div>
				{/if}
			</div>

			<div class="modal-footer">
				{#if !hasActiveSyncJob && !syncing}
					<button class="btn btn-secondary" onclick={handleClose}>
						Cancel
					</button>
							<button class="btn btn-primary" onclick={startSync}>
								{#if isInitialSync}
									🔄 Start Full Database Sync
								{:else}
									⚡ Start Incremental Database Sync
								{/if}
							</button>
				{:else}
					<button class="btn btn-secondary" onclick={handleClose} disabled={syncing}>
						{syncing ? 'Syncing...' : 'Close'}
					</button>
				{/if}
			</div>
		</div>
	</div>
{/if}

<style>
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
		background: white;
		border-radius: 12px;
		max-width: 600px;
		width: 90%;
		max-height: 80vh;
		overflow-y: auto;
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 24px;
		border-bottom: 1px solid #eee;
	}

	.modal-header h2 {
		margin: 0;
		color: #333;
		font-size: 1.5rem;
	}

	.close-btn {
		background: none;
		border: none;
		font-size: 24px;
		cursor: pointer;
		color: #666;
		padding: 4px;
		border-radius: 4px;
	}

	.close-btn:hover:not(:disabled) {
		background: #f5f5f5;
		color: #333;
	}

	.close-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.modal-body {
		padding: 24px;
	}

	.error-section {
		background: #fef2f2;
		border: 1px solid #fecaca;
		border-radius: 8px;
		padding: 20px;
		margin-bottom: 16px;
	}

	.error-header h3 {
		margin: 0 0 12px 0;
		color: #dc2626;
		font-size: 1.1rem;
	}

	.error-content p {
		margin: 0 0 16px 0;
		color: #7f1d1d;
		line-height: 1.5;
		font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
		background: #fee2e2;
		padding: 12px;
		border-radius: 4px;
		font-size: 0.9rem;
		word-break: break-word;
	}

	.error-actions {
		display: flex;
		gap: 12px;
		flex-wrap: wrap;
	}

	.sync-type-selection {
		margin-bottom: 24px;
	}

	.radio-option {
		display: block;
		margin-bottom: 16px;
		cursor: pointer;
		border: 2px solid #e5e7eb;
		border-radius: 8px;
		padding: 16px;
		transition: all 0.2s;
	}

	.radio-option:hover {
		border-color: #3b82f6;
		background: #f8faff;
	}

	.radio-option:has(input:checked) {
		border-color: #3b82f6;
		background: #f0f7ff;
	}

	.radio-option input {
		margin-right: 12px;
	}

	.radio-content p {
		margin: 4px 0 0 0;
		color: #666;
		font-size: 0.9rem;
	}

	.date-selection {
		margin-bottom: 24px;
		padding: 16px;
		background: #f9fafb;
		border-radius: 8px;
	}

	.date-selection label {
		display: block;
	}

	.date-selection input {
		width: 100%;
		padding: 8px 12px;
		border: 1px solid #d1d5db;
		border-radius: 6px;
		margin-top: 8px;
	}

	.date-selection small {
		color: #666;
		font-size: 0.85rem;
		margin-top: 4px;
		display: block;
	}

	.sync-info {
		background: #f0f9ff;
		padding: 16px;
		border-radius: 8px;
		margin-bottom: 24px;
	}

	.sync-info h4 {
		margin: 0 0 12px 0;
		color: #1e40af;
	}

	.sync-info ul {
		margin: 0;
		padding-left: 20px;
	}

	.sync-info li {
		margin-bottom: 4px;
		color: #374151;
	}

	.recent-jobs {
		margin-top: 24px;
	}

	.recent-jobs h4 {
		margin: 0 0 12px 0;
		color: #374151;
	}

	.jobs-list {
		border: 1px solid #e5e7eb;
		border-radius: 8px;
		overflow: hidden;
	}

	.job-item {
		display: grid;
		grid-template-columns: 1fr auto auto;
		gap: 12px;
		padding: 12px 16px;
		border-bottom: 1px solid #e5e7eb;
		align-items: center;
	}

	.job-item:last-child {
		border-bottom: none;
	}

	.job-item.status-completed {
		background: #f0fdf4;
	}

	.job-item.status-failed {
		background: #fef2f2;
	}

	.job-item.status-running {
		background: #fefce8;
	}

	.job-info {
		font-size: 0.9rem;
	}

	.job-status {
		font-size: 0.85rem;
		white-space: nowrap;
	}

	.job-date {
		font-size: 0.8rem;
		color: #666;
		white-space: nowrap;
	}

	.sync-active,
	.sync-starting {
		text-align: center;
		padding: 24px;
	}

	.sync-active h3,
	.sync-starting h3 {
		margin: 0 0 16px 0;
		color: #1e40af;
	}

	.sync-info p {
		margin: 4px 0;
		color: #374151;
	}

	.progress-container {
		margin: 24px 0;
	}

	.progress-bar {
		width: 100%;
		height: 8px;
		background: #e5e7eb;
		border-radius: 4px;
		overflow: hidden;
		margin-bottom: 8px;
	}

	.progress-fill {
		height: 100%;
		background: linear-gradient(90deg, #3b82f6, #1d4ed8);
		transition: width 0.3s ease;
	}

	.progress-text {
		font-size: 0.9rem;
		color: #374151;
	}

	.loading-spinner {
		width: 40px;
		height: 40px;
		border: 4px solid #e5e7eb;
		border-top: 4px solid #3b82f6;
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin: 0 auto 16px auto;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.sync-actions {
		margin-top: 16px;
	}

	.modal-footer {
		padding: 24px;
		border-top: 1px solid #eee;
		display: flex;
		justify-content: flex-end;
		gap: 12px;
	}

	.btn {
		padding: 10px 20px;
		border-radius: 6px;
		border: none;
		cursor: pointer;
		font-weight: 500;
		transition: all 0.2s;
	}

	.btn-primary {
		background: #3b82f6;
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: #1d4ed8;
	}

	.btn-secondary {
		background: #f3f4f6;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-secondary:hover:not(:disabled) {
		background: #e5e7eb;
	}

	.btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
</style>
