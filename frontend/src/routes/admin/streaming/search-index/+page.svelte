<script lang="ts">
	import { onMount } from 'svelte';
	import { toastStore } from '$lib/stores/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	// State management
	let loading = $state(true);
	let schedulerStatus = $state<any>(null);
	let config = $state({
		schedule: '0 0 * * *',
		autoSync: true,
		backupSchedule: '0 6 * * *',
		enableBackup: true,
		timezone: 'America/Denver'
	});
	let stats = $state<any>(null);
	let isGenerating = $state(false);
	let isSaving = $state(false);

	// Helper function to get auth token
	function getAuthToken(): string | null {
		return localStorage.getItem('access_token');
	}

	// Load all data on mount
	onMount(() => {
		loadDashboardData();
	});

	async function loadDashboardData() {
		loading = true;
		try {
			await Promise.all([
				loadSchedulerStatus(),
				loadConfiguration(),
				loadStats()
			]);
		} catch (error) {
			console.error('Failed to load dashboard data:', error);
			toastStore.error('Failed to load search index dashboard data');
		} finally {
			loading = false;
		}
	}

	async function loadSchedulerStatus() {
		try {
			const response = await fetch('/api/v1/admin/search-index/scheduler/status', {
				headers: {
					'Authorization': `Bearer ${getAuthToken()}`
				}
			});

			if (!response.ok) {
				throw new Error(`HTTP ${response.status}`);
			}

			const data = await response.json();
			schedulerStatus = data.status;
			console.log('📊 Scheduler status loaded:', schedulerStatus);
		} catch (error) {
			console.error('Failed to load scheduler status:', error);
			throw error;
		}
	}

	async function loadConfiguration() {
		try {
			const response = await fetch('/api/v1/admin/search-index/config', {
				headers: {
					'Authorization': `Bearer ${getAuthToken()}`
				}
			});

			if (!response.ok) {
				throw new Error(`HTTP ${response.status}`);
			}

			const data = await response.json();
			config = { ...config, ...data.config };
			console.log('⚙️ Configuration loaded:', config);
		} catch (error) {
			console.error('Failed to load configuration:', error);
			throw error;
		}
	}

	async function loadStats() {
		try {
			const response = await fetch('/api/v1/admin/search-index/stats', {
				headers: {
					'Authorization': `Bearer ${getAuthToken()}`
				}
			});

			if (!response.ok) {
				throw new Error(`HTTP ${response.status}`);
			}

			const data = await response.json();
			stats = data.stats;
			console.log('📈 Stats loaded:', stats);
		} catch (error) {
			console.error('Failed to load stats:', error);
			throw error;
		}
	}

	function handleFormSubmit(event: Event) {
		event.preventDefault();
		saveConfiguration();
	}

	async function saveConfiguration() {
		isSaving = true;
		try {
			const response = await fetch('/api/v1/admin/search-index/config', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${getAuthToken()}`
				},
				body: JSON.stringify(config)
			});

			if (!response.ok) {
				throw new Error(`HTTP ${response.status}`);
			}

			toastStore.success('Search index configuration saved successfully');
			
			// Reload status after configuration change
			await loadSchedulerStatus();
		} catch (error) {
			console.error('Failed to save configuration:', error);
			toastStore.error('Failed to save search index configuration');
		} finally {
			isSaving = false;
		}
	}

	async function triggerManualGeneration() {
		isGenerating = true;
		try {
			const response = await fetch('/api/v1/admin/search-index/scheduler/trigger', {
				method: 'POST',
				headers: {
					'Authorization': `Bearer ${getAuthToken()}`
				}
			});

			if (!response.ok) {
				throw new Error(`HTTP ${response.status}`);
			}

			toastStore.success('Search index generation triggered successfully');
			
			// Reload status and stats
			await Promise.all([
				loadSchedulerStatus(),
				loadStats()
			]);
		} catch (error) {
			console.error('Failed to trigger generation:', error);
			toastStore.error('Failed to trigger search index generation');
		} finally {
			isGenerating = false;
		}
	}

	async function downloadSearchIndex() {
		try {
			const response = await fetch('/api/v1/admin/search-index/download', {
				headers: {
					'Authorization': `Bearer ${getAuthToken()}`
				}
			});

			if (!response.ok) {
				throw new Error(`HTTP ${response.status}`);
			}

			// Create download link
			const blob = await response.blob();
			const url = window.URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = 'search-index.json';
			document.body.appendChild(a);
			a.click();
			window.URL.revokeObjectURL(url);
			document.body.removeChild(a);

			toastStore.success('Search index downloaded successfully');
		} catch (error) {
			console.error('Failed to download search index:', error);
			toastStore.error('Failed to download search index');
		}
	}

	// Format file size
	function formatFileSize(bytes: number): string {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	}

	// Format cron schedule for display
	function formatSchedule(schedule: string): string {
		const scheduleMap: Record<string, string> = {
			'0 0 * * *': 'Daily at midnight',
			'0 6 * * *': 'Daily at 6:00 AM',
			'0 12 * * *': 'Daily at noon',
			'0 18 * * *': 'Daily at 6:00 PM',
			'0 0 */2 * *': 'Every 2 days at midnight',
			'0 0 * * 0': 'Weekly on Sunday at midnight'
		};
		return scheduleMap[schedule] || schedule;
	}
</script>

<svelte:head>
	<title>Search Index Management - BOME Admin</title>
</svelte:head>

<div class="search-index-dashboard">
	<div class="dashboard-header">
		<h1>🔍 Search Index Management</h1>
		<p>Manage automated search index generation for lightning-fast video search</p>
	</div>

	{#if loading}
		<div class="loading-container">
			<LoadingSpinner size="large" />
			<p>Loading search index dashboard...</p>
		</div>
	{:else}
		<div class="dashboard-grid">
			<!-- Status Overview -->
			<div class="card status-card">
				<div class="card-header">
					<h2>📊 Status Overview</h2>
				</div>
				<div class="card-content">
					{#if schedulerStatus}
						<div class="status-grid">
							<div class="status-item">
								<div class="status-label">Scheduler Status</div>
								<div class="status-value {schedulerStatus.running ? 'running' : 'stopped'}">
									{schedulerStatus.running ? '🟢 Running' : '🔴 Stopped'}
								</div>
							</div>
							
							{#if schedulerStatus.lastRun}
								<div class="status-item">
									<div class="status-label">Last Generation</div>
									<div class="status-value">{schedulerStatus.lastRun}</div>
								</div>
							{/if}
							
							{#if schedulerStatus.nextRun}
								<div class="status-item">
									<div class="status-label">Next Generation</div>
									<div class="status-value">{schedulerStatus.nextRun}</div>
								</div>
							{/if}
							
							<div class="status-item">
								<div class="status-label">Schedule</div>
								<div class="status-value">{schedulerStatus.schedule || 'Not configured'}</div>
							</div>
						</div>
					{/if}
				</div>
			</div>

			<!-- Statistics -->
			<div class="card stats-card">
				<div class="card-header">
					<h2>📈 Statistics</h2>
				</div>
				<div class="card-content">
					{#if stats}
						<div class="stats-grid">
							<div class="stat-item">
								<div class="stat-number">{stats.totalVideos.toLocaleString()}</div>
								<div class="stat-label">Total Videos</div>
							</div>
							
							{#if stats.searchIndex.exists}
								<div class="stat-item">
									<div class="stat-number">{formatFileSize(stats.searchIndex.size)}</div>
									<div class="stat-label">Index Size</div>
								</div>
								
								<div class="stat-item">
									<div class="stat-number">{stats.searchIndex.age}</div>
									<div class="stat-label">Index Age</div>
								</div>
							{:else}
								<div class="stat-item">
									<div class="stat-number">❌</div>
									<div class="stat-label">No Index File</div>
								</div>
							{/if}
						</div>
					{/if}
				</div>
			</div>

			<!-- Manual Actions -->
			<div class="card actions-card">
				<div class="card-header">
					<h2>🎛️ Manual Actions</h2>
				</div>
				<div class="card-content">
					<div class="action-buttons">
						<button 
							class="btn btn-primary" 
							onclick={triggerManualGeneration}
							disabled={isGenerating}
						>
							{#if isGenerating}
								<LoadingSpinner size="small" />
								Generating...
							{:else}
								🚀 Generate Now
							{/if}
						</button>

						{#if stats?.searchIndex?.exists}
							<button 
								class="btn btn-secondary" 
								onclick={downloadSearchIndex}
							>
								📥 Download Index
							</button>
						{/if}

						<button 
							class="btn btn-outline" 
							onclick={loadDashboardData}
						>
							🔄 Refresh Status
						</button>
					</div>
				</div>
			</div>

			<!-- Configuration -->
			<div class="card config-card full-width">
				<div class="card-header">
					<h2>⚙️ Scheduler Configuration</h2>
				</div>
				<div class="card-content">
					<form onsubmit={handleFormSubmit}>
						<div class="form-grid">
							<div class="form-group">
								<label for="autoSync">
									<input 
										type="checkbox" 
										id="autoSync"
										bind:checked={config.autoSync}
									/>
									Enable Automatic Generation
								</label>
								<small>Automatically generate search index on schedule</small>
							</div>

							<div class="form-group">
								<label for="schedule">Main Schedule (Cron)</label>
								<select id="schedule" bind:value={config.schedule}>
									<option value="0 0 * * *">Daily at midnight</option>
									<option value="0 6 * * *">Daily at 6:00 AM</option>
									<option value="0 12 * * *">Daily at noon</option>
									<option value="0 18 * * *">Daily at 6:00 PM</option>
									<option value="0 0 */2 * *">Every 2 days at midnight</option>
									<option value="0 0 * * 0">Weekly on Sunday at midnight</option>
								</select>
								<small>When to automatically generate the search index</small>
							</div>

							<div class="form-group">
								<label for="enableBackup">
									<input 
										type="checkbox" 
										id="enableBackup"
										bind:checked={config.enableBackup}
									/>
									Enable Backup Schedule
								</label>
								<small>Run a backup generation at a different time</small>
							</div>

							{#if config.enableBackup}
								<div class="form-group">
									<label for="backupSchedule">Backup Schedule (Cron)</label>
									<select id="backupSchedule" bind:value={config.backupSchedule}>
										<option value="0 6 * * *">Daily at 6:00 AM</option>
										<option value="0 12 * * *">Daily at noon</option>
										<option value="0 18 * * *">Daily at 6:00 PM</option>
										<option value="0 3 * * *">Daily at 3:00 AM</option>
									</select>
									<small>Backup generation schedule</small>
								</div>
							{/if}
						</div>

						<div class="form-actions">
							<button 
								type="submit" 
								class="btn btn-primary"
								disabled={isSaving}
							>
								{#if isSaving}
									<LoadingSpinner size="small" />
									Saving...
								{:else}
									💾 Save Configuration
								{/if}
							</button>
						</div>
					</form>
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	.search-index-dashboard {
		padding: 2rem;
		max-width: 1200px;
		margin: 0 auto;
	}

	.dashboard-header {
		margin-bottom: 2rem;
		text-align: center;
	}

	.dashboard-header h1 {
		font-size: 2.5rem;
		margin-bottom: 0.5rem;
		color: var(--color-primary);
	}

	.dashboard-header p {
		font-size: 1.1rem;
		color: var(--color-text-secondary);
		margin: 0;
	}

	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 4rem;
		gap: 1rem;
	}

	.dashboard-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
		gap: 1.5rem;
		margin-bottom: 2rem;
	}

	.card {
		background: var(--color-surface);
		border-radius: 12px;
		border: 1px solid var(--color-border);
		overflow: hidden;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
	}

	.card.full-width {
		grid-column: 1 / -1;
	}

	.card-header {
		background: var(--color-surface-variant);
		padding: 1rem 1.5rem;
		border-bottom: 1px solid var(--color-border);
	}

	.card-header h2 {
		margin: 0;
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--color-text-primary);
	}

	.card-content {
		padding: 1.5rem;
	}

	.status-grid {
		display: grid;
		gap: 1rem;
	}

	.status-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.75rem;
		background: var(--color-background);
		border-radius: 8px;
		border: 1px solid var(--color-border);
	}

	.status-label {
		font-weight: 500;
		color: var(--color-text-secondary);
	}

	.status-value {
		font-weight: 600;
		color: var(--color-text-primary);
	}

	.status-value.running {
		color: var(--color-success);
	}

	.status-value.stopped {
		color: var(--color-error);
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
		gap: 1rem;
	}

	.stat-item {
		text-align: center;
		padding: 1rem;
		background: var(--color-background);
		border-radius: 8px;
		border: 1px solid var(--color-border);
	}

	.stat-number {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--color-primary);
		margin-bottom: 0.25rem;
	}

	.stat-label {
		font-size: 0.875rem;
		color: var(--color-text-secondary);
		font-weight: 500;
	}

	.action-buttons {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.form-grid {
		display: grid;
		gap: 1.5rem;
		margin-bottom: 2rem;
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.form-group label {
		font-weight: 500;
		color: var(--color-text-primary);
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.form-group input[type="checkbox"] {
		margin: 0;
	}

	.form-group select {
		padding: 0.75rem;
		border: 1px solid var(--color-border);
		border-radius: 6px;
		background: var(--color-background);
		color: var(--color-text-primary);
		font-size: 0.875rem;
	}

	.form-group small {
		color: var(--color-text-secondary);
		font-size: 0.8rem;
	}

	.form-actions {
		display: flex;
		justify-content: flex-start;
	}

	.btn {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1.5rem;
		border-radius: 6px;
		font-weight: 500;
		text-decoration: none;
		border: none;
		cursor: pointer;
		transition: all 0.2s ease;
		font-size: 0.875rem;
	}

	.btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-primary {
		background: var(--color-primary);
		color: var(--color-text-primary);
	}

	.btn-primary:hover:not(:disabled) {
		background: var(--color-primary-dark);
	}

	.btn-secondary {
		background: var(--color-secondary);
		color: var(--color-text-primary);
	}

	.btn-secondary:hover:not(:disabled) {
		background: var(--color-secondary-dark);
	}

	.btn-outline {
		background: transparent;
		color: var(--color-primary);
		border: 1px solid var(--color-primary);
	}

	.btn-outline:hover:not(:disabled) {
		background: var(--color-primary);
		color: var(--color-text-primary);
	}

	@media (max-width: 768px) {
		.search-index-dashboard {
			padding: 1rem;
		}

		.dashboard-grid {
			grid-template-columns: 1fr;
		}

		.action-buttons {
			flex-direction: column;
		}
	}
</style>
