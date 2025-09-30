<script lang="ts">
	import { onMount } from 'svelte';
	import { showToast } from '$lib/toast';
	import { apiRequest } from '$lib/auth';


	// State variables
	let isLoading = true;
	let isSaving = false;
	let isSyncing = false;
	let schedulerStatus: any = null;
	let syncStatus: any = null;
	let recentVideos: any[] = [];

	// Configuration form data
	let channelId = '';
	let syncHour = 14; // 2 PM
	let syncMinute = 0;
	let timezone = 'MST';
	let autoSyncEnabled = true;

	// Load initial data
	onMount(async () => {
		await loadDashboardData();
	});

	async function loadDashboardData() {
		isLoading = true;
		try {
			// Load scheduler status
			await loadSchedulerStatus();
			
			// Load sync status
			await loadSyncStatus();
			
			// Load recent videos
			await loadRecentVideos();
			
			// Load current configuration (if stored)
			await loadConfiguration();
			
		} catch (error) {
			console.error('Failed to load dashboard data:', error);
			showToast('Failed to load dashboard data', 'error');
		} finally {
			isLoading = false;
		}
	}

	async function loadSchedulerStatus() {
		try {
			const response = await apiRequest('/youtube/scheduler/status', {
				method: 'GET'
			});
			
			if (response.ok) {
				const data = await response.json();
				schedulerStatus = data.scheduler;
			}
		} catch (error) {
			console.error('Failed to load scheduler status:', error);
		}
	}

	async function loadSyncStatus() {
		try {
			const response = await apiRequest('/youtube/sync/status', {
				method: 'GET'
			});
			
			if (response.ok) {
				syncStatus = await response.json();
			}
		} catch (error) {
			console.error('Failed to load sync status:', error);
		}
	}

	async function loadRecentVideos() {
		try {
			const response = await apiRequest('/youtube/videos/latest?limit=5', {
				method: 'GET'
			});
			
			if (response.ok) {
				const data = await response.json();
				recentVideos = data.videos || [];
			}
		} catch (error) {
			console.error('Failed to load recent videos:', error);
		}
	}

	async function loadConfiguration() {
		try {
			const response = await apiRequest('/youtube/config', {
				method: 'GET'
			});
			
			if (response.ok) {
				const data = await response.json();
				const config = data.config;
				
				channelId = config.channel_id || '';
				syncHour = config.sync_hour || 14;
				syncMinute = config.sync_minute || 0;
				timezone = config.timezone || 'MST';
				autoSyncEnabled = config.auto_sync_enabled !== false;
			}
		} catch (error) {
			console.error('Failed to load configuration:', error);
		}
	}

	async function saveConfiguration() {
		isSaving = true;
		try {
			const response = await apiRequest('/youtube/config', {
				method: 'POST',
				body: JSON.stringify({
					channel_id: channelId,
					sync_hour: syncHour,
					sync_minute: syncMinute,
					timezone: timezone,
					auto_sync_enabled: autoSyncEnabled
				})
			});
			
			if (response.ok) {
				showToast('Configuration saved successfully', 'success');
				// Refresh scheduler status
				await loadSchedulerStatus();
			} else {
				const error = await response.json();
				showToast(`Failed to save configuration: ${error.details || 'Unknown error'}`, 'error');
			}
		} catch (error) {
			console.error('Failed to save configuration:', error);
			showToast('Failed to save configuration', 'error');
		} finally {
			isSaving = false;
		}
	}

	async function triggerManualSync() {
		isSyncing = true;
		try {
			const response = await apiRequest('/youtube/scheduler/trigger', {
				method: 'POST'
			});
			
			if (response.ok) {
				const data = await response.json();
				showToast(`Sync completed: ${data.result.new_videos} new videos, ${data.result.updated_videos} updated`, 'success');
				
				// Refresh data
				await loadDashboardData();
			} else {
				const error = await response.json();
				showToast(`Sync failed: ${error.details || 'Unknown error'}`, 'error');
			}
		} catch (error) {
			console.error('Failed to trigger sync:', error);
			showToast('Failed to trigger sync', 'error');
		} finally {
			isSyncing = false;
		}
	}

	function formatDate(dateString: string) {
		if (!dateString) return 'Never';
		return new Date(dateString).toLocaleString();
	}

	function formatDuration(seconds: number) {
		const hours = Math.floor(seconds / 3600);
		const minutes = Math.floor((seconds % 3600) / 60);
		const secs = seconds % 60;
		
		if (hours > 0) {
			return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
		}
		return `${minutes}:${secs.toString().padStart(2, '0')}`;
	}
</script>

<div class="youtube-dashboard">
	<div class="dashboard-header">
		<h1>YouTube Management</h1>
		<p>Manage YouTube RSS feed synchronization and configuration</p>
	</div>

	{#if isLoading}
		<div class="loading-container">
			<div class="loading-spinner"></div>
			<p>Loading dashboard...</p>
		</div>
	{:else}
		<div class="dashboard-grid">
			<!-- Configuration Panel -->
			<div class="panel configuration-panel">
				<h2>RSS Feed Configuration</h2>
				
				<form on:submit|preventDefault={saveConfiguration}>
					<div class="form-group">
						<label for="channelId">YouTube Channel ID</label>
						<input 
							type="text" 
							id="channelId" 
							bind:value={channelId}
							placeholder="UCYourChannelIdHere"
							required
						/>
						<small>Find your channel ID in YouTube Studio → Settings → Channel</small>
					</div>

					<div class="form-group">
						<label>Sync Schedule</label>
						<div class="time-inputs">
							<select bind:value={syncHour}>
								{#each Array(24) as _, i}
									<option value={i}>{i.toString().padStart(2, '0')}:00</option>
								{/each}
							</select>
							<span>MST (Mountain Standard Time)</span>
						</div>
					</div>

					<div class="form-group">
						<label class="checkbox-label">
							<input 
								type="checkbox" 
								bind:checked={autoSyncEnabled}
							/>
							Enable automatic daily sync
						</label>
					</div>

					<button 
						type="submit" 
						class="btn btn-primary"
						disabled={isSaving}
					>
						{isSaving ? 'Saving...' : 'Save Configuration'}
					</button>
				</form>
			</div>

			<!-- Scheduler Status Panel -->
			<div class="panel status-panel">
				<h2>Scheduler Status</h2>
				
				{#if schedulerStatus}
					<div class="status-grid">
						<div class="status-item">
							<span class="status-label">Status:</span>
							<span class="status-value {schedulerStatus.running ? 'running' : 'stopped'}">
								{schedulerStatus.running ? '🟢 Running' : '🔴 Stopped'}
							</span>
						</div>
						
						<div class="status-item">
							<span class="status-label">Next Sync:</span>
							<span class="status-value">{schedulerStatus.next_sync || 'Not scheduled'}</span>
						</div>
						
						<div class="status-item">
							<span class="status-label">Time Until Next:</span>
							<span class="status-value">{schedulerStatus.next_sync_in || 'N/A'}</span>
						</div>
						
						<div class="status-item">
							<span class="status-label">Sync Schedule:</span>
							<span class="status-value">{schedulerStatus.sync_time || '2:00 PM MST daily'}</span>
						</div>
					</div>
				{:else}
					<p class="no-data">Scheduler status not available</p>
				{/if}

				<div class="panel-actions">
					<button 
						class="btn btn-secondary"
						on:click={triggerManualSync}
						disabled={isSyncing}
					>
						{isSyncing ? 'Syncing...' : '🔄 Manual Sync'}
					</button>
				</div>
			</div>

			<!-- Sync History Panel -->
			<div class="panel sync-panel">
				<h2>Sync Information</h2>
				
				{#if syncStatus}
					<div class="sync-stats">
						<div class="stat-item">
							<span class="stat-number">{syncStatus.total_videos || 0}</span>
							<span class="stat-label">Total Videos</span>
						</div>
						
						<div class="stat-item">
							<span class="stat-number">{formatDate(syncStatus.last_sync)}</span>
							<span class="stat-label">Last Sync</span>
						</div>
						
						<div class="stat-item">
							<span class="stat-number">{syncStatus.sync_enabled ? 'Yes' : 'No'}</span>
							<span class="stat-label">Sync Enabled</span>
						</div>
					</div>
				{:else}
					<p class="no-data">Sync information not available</p>
				{/if}
			</div>

			<!-- Recent Videos Panel -->
			<div class="panel videos-panel">
				<h2>Recent Videos</h2>
				
				{#if recentVideos.length > 0}
					<div class="video-list">
						{#each recentVideos as video}
							<div class="video-item">
								<div class="video-thumbnail">
									<img src={video.thumbnail_url || '/16X10_Placeholder_IMG.png'} alt={video.title} />
								</div>
								<div class="video-info">
									<h4>{video.title}</h4>
									<p class="video-meta">
										{formatDate(video.published)} • {formatDuration(video.duration || 0)}
									</p>
									<p class="video-stats">
										{video.view_count || 0} views • {video.like_count || 0} likes
									</p>
								</div>
							</div>
						{/each}
					</div>
				{:else}
					<p class="no-data">No videos found. Try running a manual sync.</p>
				{/if}
			</div>
		</div>
	{/if}
</div>

<style>
	.youtube-dashboard {
		padding: 2rem;
		max-width: 1400px;
		margin: 0 auto;
	}

	.dashboard-header {
		margin-bottom: 2rem;
	}

	.dashboard-header h1 {
		font-size: 2rem;
		font-weight: 600;
		color: var(--text);
		margin-bottom: 0.5rem;
	}

	.dashboard-header p {
		color: var(--text-muted);
		font-size: 1.1rem;
	}

	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 4rem;
		text-align: center;
	}

	.loading-spinner {
		width: 40px;
		height: 40px;
		border: 4px solid var(--border);
		border-top: 4px solid var(--primary);
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-bottom: 1rem;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.dashboard-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
		gap: 2rem;
	}

	.panel {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		padding: 1.5rem;
		box-shadow: var(--shadow-sm);
	}

	.panel h2 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text);
		margin-bottom: 1rem;
		padding-bottom: 0.5rem;
		border-bottom: 1px solid var(--border);
	}

	.form-group {
		margin-bottom: 1.5rem;
	}

	.form-group label {
		display: block;
		font-weight: 500;
		color: var(--text);
		margin-bottom: 0.5rem;
	}

	.form-group input,
	.form-group select {
		width: 100%;
		padding: 0.75rem;
		border: 1px solid var(--border);
		border-radius: var(--radius-md);
		background: var(--surface-secondary);
		color: var(--text);
		font-size: 1rem;
	}

	.form-group small {
		display: block;
		color: var(--text-muted);
		font-size: 0.875rem;
		margin-top: 0.25rem;
	}

	.time-inputs {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.time-inputs select {
		width: auto;
		min-width: 120px;
	}

	.checkbox-label {
		display: flex !important;
		align-items: center;
		gap: 0.5rem;
		cursor: pointer;
	}

	.checkbox-label input[type="checkbox"] {
		width: auto !important;
		margin: 0;
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
		background: var(--surface-secondary);
		border-radius: var(--radius-md);
	}

	.status-label {
		font-weight: 500;
		color: var(--text-muted);
	}

	.status-value {
		font-weight: 600;
		color: var(--text);
	}

	.status-value.running {
		color: var(--success);
	}

	.status-value.stopped {
		color: var(--error);
	}

	.panel-actions {
		margin-top: 1.5rem;
		padding-top: 1rem;
		border-top: 1px solid var(--border);
	}

	.sync-stats {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
		gap: 1rem;
	}

	.stat-item {
		text-align: center;
		padding: 1rem;
		background: var(--surface-secondary);
		border-radius: var(--radius-md);
	}

	.stat-number {
		display: block;
		font-size: 1.5rem;
		font-weight: 600;
		color: var(--primary);
		margin-bottom: 0.25rem;
	}

	.stat-label {
		font-size: 0.875rem;
		color: var(--text-muted);
		font-weight: 500;
	}

	.videos-panel {
		grid-row: 2;
		grid-column: 1/4;
		display: flex;
		flex-direction: row;
		gap: 1rem;
		flex-wrap: wrap;
		justify-content: space-between;
	}

	.video-list {
		grid-row: 2;
		grid-column: 1/3;
		display: flex;
		flex-direction: row;
		flex-wrap: wrap;
		gap: 1rem;
	}

	.video-item {
		display: flex;
		flex-direction: column;
		max-width: 300px;
		gap: 1rem;
		padding: 1rem;
		background: var(--surface-secondary);
		border-radius: var(--radius-md);
		border: 1px solid var(--border);
	}

	.video-thumbnail {
		flex-shrink: 0;
		width: 220px;
		height: 168px;
		border-radius: var(--radius-sm);
		overflow: hidden;
	}

	.video-thumbnail img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.video-info {
		flex: 1;
		min-width: 0;
	}

	.video-info h4 {
		font-size: 0.9rem;
		font-weight: 600;
		color: var(--text);
		margin-bottom: 0.25rem;
		line-height: 1.3;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	.video-meta,
	.video-stats {
		font-size: 0.8rem;
		color: var(--text-muted);
		margin-bottom: 0.25rem;
	}

	.no-data {
		text-align: center;
		color: var(--text-muted);
		font-style: italic;
		padding: 2rem;
	}

	.btn {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: var(--radius-md);
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
	}

	.btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-primary {
		background: var(--primary);
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: var(--primary-dark);
	}

	.btn-secondary {
		background: var(--surface-secondary);
		color: var(--text);
		border: 1px solid var(--border);
	}

	.btn-secondary:hover:not(:disabled) {
		background: var(--surface-tertiary);
	}

	@media (max-width: 768px) {
		.youtube-dashboard {
			padding: 1rem;
		}

		.dashboard-grid {
			grid-template-columns: 1fr;
		}

		.video-item {
			flex-direction: column;
		}

		.video-thumbnail {
			width: 100%;
			height: 180px;
		}
	}
</style>
