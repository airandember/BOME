<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	interface BackupJob {
		id: string;
		name: string;
		type: 'database' | 'files' | 'full';
		schedule: string;
		status: 'active' | 'inactive' | 'running' | 'failed';
		last_run?: string;
		next_run?: string;
		size?: string;
		retention_days: number;
		destination: string;
		created_at: string;
	}

	interface BackupHistory {
		id: string;
		job_id: string;
		job_name: string;
		started_at: string;
		completed_at?: string;
		status: 'running' | 'completed' | 'failed';
		size?: string;
		duration?: number;
		error_message?: string;
	}

	let backupJobs: BackupJob[] = [];
	let backupHistory: BackupHistory[] = [];
	let loading = true;
	let error = '';
	let showCreateModal = false;
	let showEditModal = false;
	let selectedJob: BackupJob | null = null;

	// Form data
	let formData = {
		name: '',
		type: 'database' as 'database' | 'files' | 'full',
		schedule: 'daily',
		retention_days: 30,
		destination: 'digitalocean'
	};

	const scheduleOptions = [
		{ value: 'hourly', label: 'Every Hour' },
		{ value: 'daily', label: 'Daily at 2 AM' },
		{ value: 'weekly', label: 'Weekly (Sunday 2 AM)' },
		{ value: 'monthly', label: 'Monthly (1st at 2 AM)' }
	];

	const destinationOptions = [
		{ value: 'digitalocean', label: 'Digital Ocean Spaces' },
		{ value: 'aws', label: 'AWS S3' },
		{ value: 'local', label: 'Local Storage' }
	];

	const typeOptions = [
		{ value: 'database', label: 'Database Only' },
		{ value: 'files', label: 'Files Only' },
		{ value: 'full', label: 'Full System' }
	];

	onMount(async () => {
		await loadBackupData();
	});

	const loadBackupData = async () => {
		try {
			loading = true;
			error = '';
			
			// Mock data - replace with actual API call
			await new Promise(resolve => setTimeout(resolve, 1000));
			
			backupJobs = [
				{
					id: 'job_001',
					name: 'Daily Database Backup',
					type: 'database',
					schedule: 'daily',
					status: 'active',
					last_run: '2024-06-18T02:00:00Z',
					next_run: '2024-06-19T02:00:00Z',
					size: '2.5 GB',
					retention_days: 30,
					destination: 'digitalocean',
					created_at: '2024-05-01T10:00:00Z'
				},
				{
					id: 'job_002',
					name: 'Weekly Full Backup',
					type: 'full',
					schedule: 'weekly',
					status: 'active',
					last_run: '2024-06-16T02:00:00Z',
					next_run: '2024-06-23T02:00:00Z',
					size: '15.2 GB',
					retention_days: 90,
					destination: 'aws',
					created_at: '2024-05-01T10:00:00Z'
				},
				{
					id: 'job_003',
					name: 'Video Files Backup',
					type: 'files',
					schedule: 'daily',
					status: 'running',
					last_run: '2024-06-18T02:00:00Z',
					next_run: '2024-06-19T02:00:00Z',
					size: '125.8 GB',
					retention_days: 60,
					destination: 'digitalocean',
					created_at: '2024-05-15T14:30:00Z'
				},
				{
					id: 'job_004',
					name: 'Test Backup',
					type: 'database',
					schedule: 'weekly',
					status: 'failed',
					last_run: '2024-06-17T02:00:00Z',
					retention_days: 7,
					destination: 'local',
					created_at: '2024-06-10T12:00:00Z'
				}
			];

			backupHistory = [
				{
					id: 'hist_001',
					job_id: 'job_001',
					job_name: 'Daily Database Backup',
					started_at: '2024-06-18T02:00:00Z',
					completed_at: '2024-06-18T02:15:00Z',
					status: 'completed',
					size: '2.5 GB',
					duration: 15
				},
				{
					id: 'hist_002',
					job_id: 'job_002',
					job_name: 'Weekly Full Backup',
					started_at: '2024-06-16T02:00:00Z',
					completed_at: '2024-06-16T03:45:00Z',
					status: 'completed',
					size: '15.2 GB',
					duration: 105
				},
				{
					id: 'hist_003',
					job_id: 'job_003',
					job_name: 'Video Files Backup',
					started_at: '2024-06-18T02:00:00Z',
					status: 'running',
					size: '125.8 GB'
				},
				{
					id: 'hist_004',
					job_id: 'job_004',
					job_name: 'Test Backup',
					started_at: '2024-06-17T02:00:00Z',
					completed_at: '2024-06-17T02:05:00Z',
					status: 'failed',
					duration: 5,
					error_message: 'Insufficient disk space on local storage'
				}
			];
		} catch (err) {
			error = 'Failed to load backup data';
			console.error('Error loading backup data:', err);
		} finally {
			loading = false;
		}
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

	const formatDuration = (minutes: number): string => {
		if (minutes < 60) {
			return `${minutes}m`;
		}
		const hours = Math.floor(minutes / 60);
		const mins = minutes % 60;
		return `${hours}h ${mins}m`;
	};

	const getStatusBadge = (status: string) => {
		const statusConfig = {
			active: { text: 'Active', class: 'status-active' },
			inactive: { text: 'Inactive', class: 'status-inactive' },
			running: { text: 'Running', class: 'status-running' },
			failed: { text: 'Failed', class: 'status-failed' },
			completed: { text: 'Completed', class: 'status-completed' }
		};

		const config = statusConfig[status as keyof typeof statusConfig] || { text: status, class: 'status-unknown' };
		
		return `<span class="status-badge ${config.class}">${config.text}</span>`;
	};

	const getTypeIcon = (type: string) => {
		switch (type) {
			case 'database':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<ellipse cx="12" cy="5" rx="9" ry="3"></ellipse>
					<path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"></path>
					<path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"></path>
				</svg>`;
			case 'files':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
					<polyline points="14,2 14,8 20,8"></polyline>
					<line x1="16" y1="13" x2="8" y2="13"></line>
					<line x1="16" y1="17" x2="8" y2="17"></line>
				</svg>`;
			case 'full':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect>
					<line x1="8" y1="21" x2="16" y2="21"></line>
					<line x1="12" y1="17" x2="12" y2="21"></line>
				</svg>`;
			default:
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="12" cy="12" r="10"></circle>
				</svg>`;
		}
	};

	const resetFormData = () => {
		formData = {
			name: '',
			type: 'database',
			schedule: 'daily',
			retention_days: 30,
			destination: 'digitalocean'
		};
	};

	const openCreateModal = () => {
		resetFormData();
		showCreateModal = true;
	};

	const openEditModal = (job: BackupJob) => {
		selectedJob = job;
		formData = {
			name: job.name,
			type: job.type,
			schedule: job.schedule,
			retention_days: job.retention_days,
			destination: job.destination
		};
		showEditModal = true;
	};

	const closeModals = () => {
		showCreateModal = false;
		showEditModal = false;
		selectedJob = null;
		resetFormData();
	};

	const createBackupJob = async () => {
		try {
			if (!formData.name.trim()) {
				showToast('Please enter a name for the backup job', 'error');
				return;
			}

			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 1000));
			
			showToast('Backup job created successfully', 'success');
			closeModals();
			await loadBackupData();
		} catch (err) {
			showToast('Failed to create backup job', 'error');
		}
	};

	const updateBackupJob = async () => {
		try {
			if (!formData.name.trim()) {
				showToast('Please enter a name for the backup job', 'error');
				return;
			}

			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 1000));
			
			showToast('Backup job updated successfully', 'success');
			closeModals();
			await loadBackupData();
		} catch (err) {
			showToast('Failed to update backup job', 'error');
		}
	};

	const runBackupNow = async (job: BackupJob) => {
		if (!confirm(`Are you sure you want to run "${job.name}" backup now?`)) {
			return;
		}

		try {
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 1000));
			
			showToast('Backup job started successfully', 'success');
			await loadBackupData();
		} catch (err) {
			showToast('Failed to start backup job', 'error');
		}
	};

	const toggleJobStatus = async (job: BackupJob) => {
		try {
			const newStatus = job.status === 'active' ? 'inactive' : 'active';
			
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 500));
			
			showToast(`Backup job ${newStatus === 'active' ? 'activated' : 'deactivated'}`, 'success');
			await loadBackupData();
		} catch (err) {
			showToast('Failed to update backup job status', 'error');
		}
	};

	const deleteBackupJob = async (job: BackupJob) => {
		if (!confirm(`Are you sure you want to delete the backup job "${job.name}"? This action cannot be undone.`)) {
			return;
		}

		try {
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 500));
			
			showToast('Backup job deleted successfully', 'success');
			await loadBackupData();
		} catch (err) {
			showToast('Failed to delete backup job', 'error');
		}
	};

	const downloadBackup = async (backup: BackupHistory) => {
		try {
			// Mock download - replace with actual implementation
			showToast('Download started', 'success');
		} catch (err) {
			showToast('Failed to download backup', 'error');
		}
	};

	const restoreBackup = async (backup: BackupHistory) => {
		if (!confirm(`Are you sure you want to restore from this backup? This will overwrite current data.`)) {
			return;
		}

		try {
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 2000));
			
			showToast('Backup restore initiated', 'success');
		} catch (err) {
			showToast('Failed to restore backup', 'error');
		}
	};
</script>

<svelte:head>
	<title>Backup Management - Admin Dashboard</title>
</svelte:head>

<div class="backup-page">
	<div class="page-header">
		<div class="header-content">
			<div class="header-text">
				<h1>Backup Management</h1>
				<p>Configure and monitor system backups</p>
			</div>
			
			<div class="header-actions">
				<button class="btn btn-primary" on:click={openCreateModal}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="12" y1="5" x2="12" y2="19"></line>
						<line x1="5" y1="12" x2="19" y2="12"></line>
					</svg>
					Create Backup Job
				</button>
				<button class="btn btn-outline" on:click={() => goto('/admin/security')}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M19 12H5"></path>
						<path d="M12 19l-7-7 7-7"></path>
					</svg>
					Back to Security
				</button>
			</div>
		</div>
	</div>

	{#if loading}
		<div class="loading-container">
			<LoadingSpinner size="large" color="primary" />
			<p>Loading backup data...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<p class="error-message">{error}</p>
			<button class="btn btn-primary" on:click={loadBackupData}>
				Try Again
			</button>
		</div>
	{:else}
		<!-- Backup Jobs -->
		<div class="section">
			<div class="section-header">
				<h2>Backup Jobs</h2>
			</div>
			
			<div class="jobs-table glass">
				<div class="table-header">
					<div class="header-cell">Job</div>
					<div class="header-cell">Type</div>
					<div class="header-cell">Schedule</div>
					<div class="header-cell">Status</div>
					<div class="header-cell">Last Run</div>
					<div class="header-cell">Next Run</div>
					<div class="header-cell">Size</div>
					<div class="header-cell">Actions</div>
				</div>

				{#each backupJobs as job}
					<div class="table-row">
						<div class="table-cell">
							<div class="job-info">
								<span class="job-name">{job.name}</span>
								<span class="job-destination">{job.destination}</span>
							</div>
						</div>
						<div class="table-cell">
							<div class="type-info">
								<div class="type-icon">
									{@html getTypeIcon(job.type)}
								</div>
								<span class="type-text">{job.type}</span>
							</div>
						</div>
						<div class="table-cell">
							<span class="schedule">{scheduleOptions.find(s => s.value === job.schedule)?.label || job.schedule}</span>
						</div>
						<div class="table-cell">
							{@html getStatusBadge(job.status)}
						</div>
						<div class="table-cell">
							<span class="last-run">
								{job.last_run ? formatDate(job.last_run) : 'Never'}
							</span>
						</div>
						<div class="table-cell">
							<span class="next-run">
								{job.next_run ? formatDate(job.next_run) : 'N/A'}
							</span>
						</div>
						<div class="table-cell">
							<span class="size">{job.size || 'N/A'}</span>
						</div>
						<div class="table-cell">
							<div class="actions-dropdown">
								<button class="btn btn-ghost btn-small">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<circle cx="12" cy="12" r="1"></circle>
										<circle cx="12" cy="5" r="1"></circle>
										<circle cx="12" cy="19" r="1"></circle>
									</svg>
								</button>
								<div class="dropdown-menu">
									<button class="dropdown-item" on:click={() => runBackupNow(job)}>
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<polygon points="5,3 19,12 5,21"></polygon>
										</svg>
										Run Now
									</button>
									<button class="dropdown-item" on:click={() => openEditModal(job)}>
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
											<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
										</svg>
										Edit
									</button>
									<button class="dropdown-item" on:click={() => toggleJobStatus(job)}>
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
											<circle cx="12" cy="16" r="1"></circle>
											<path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
										</svg>
										{job.status === 'active' ? 'Deactivate' : 'Activate'}
									</button>
									<div class="dropdown-divider"></div>
									<button class="dropdown-item danger" on:click={() => deleteBackupJob(job)}>
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<polyline points="3,6 5,6 21,6"></polyline>
											<path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
										</svg>
										Delete
									</button>
								</div>
							</div>
						</div>
					</div>
				{/each}

				{#if backupJobs.length === 0}
					<div class="empty-state">
						<div class="empty-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
								<polyline points="7,10 12,15 17,10"></polyline>
								<line x1="12" y1="15" x2="12" y2="3"></line>
							</svg>
						</div>
						<h3>No backup jobs configured</h3>
						<p>Create your first backup job to start protecting your data.</p>
						<button class="btn btn-primary" on:click={openCreateModal}>
							Create Backup Job
						</button>
					</div>
				{/if}
			</div>
		</div>

		<!-- Backup History -->
		<div class="section">
			<div class="section-header">
				<h2>Recent Backups</h2>
			</div>
			
			<div class="history-table glass">
				<div class="table-header">
					<div class="header-cell">Job</div>
					<div class="header-cell">Started</div>
					<div class="header-cell">Duration</div>
					<div class="header-cell">Size</div>
					<div class="header-cell">Status</div>
					<div class="header-cell">Actions</div>
				</div>

				{#each backupHistory as backup}
					<div class="table-row">
						<div class="table-cell">
							<span class="job-name">{backup.job_name}</span>
						</div>
						<div class="table-cell">
							<span class="started-at">{formatDate(backup.started_at)}</span>
						</div>
						<div class="table-cell">
							<span class="duration">
								{backup.duration ? formatDuration(backup.duration) : 'In Progress'}
							</span>
						</div>
						<div class="table-cell">
							<span class="size">{backup.size || 'N/A'}</span>
						</div>
						<div class="table-cell">
							{@html getStatusBadge(backup.status)}
							{#if backup.error_message}
								<div class="error-message" title={backup.error_message}>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<circle cx="12" cy="12" r="10"></circle>
										<line x1="12" y1="8" x2="12" y2="12"></line>
										<line x1="12" y1="16" x2="12.01" y2="16"></line>
									</svg>
								</div>
							{/if}
						</div>
						<div class="table-cell">
							{#if backup.status === 'completed'}
								<div class="backup-actions">
									<button class="btn btn-ghost btn-small" on:click={() => downloadBackup(backup)}>
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
											<polyline points="7,10 12,15 17,10"></polyline>
											<line x1="12" y1="15" x2="12" y2="3"></line>
										</svg>
									</button>
									<button class="btn btn-ghost btn-small" on:click={() => restoreBackup(backup)}>
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<polyline points="1,4 1,10 7,10"></polyline>
											<polyline points="23,20 23,14 17,14"></polyline>
											<path d="M20.49 9A9 9 0 0 0 5.64 5.64l1.27 1.27m4.18 4.18l1.27 1.27A9 9 0 0 0 18.36 18.36"></path>
										</svg>
									</button>
								</div>
							{:else if backup.status === 'running'}
								<div class="progress-indicator">
									<div class="spinner"></div>
								</div>
							{:else}
								<span class="text-muted">N/A</span>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>

<!-- Create/Edit Modal -->
{#if showCreateModal || showEditModal}
	<div class="modal-overlay" on:click={closeModals}>
		<div class="modal-content" on:click|stopPropagation>
			<div class="modal-header">
				<h2>{showCreateModal ? 'Create Backup Job' : 'Edit Backup Job'}</h2>
				<button class="modal-close" on:click={closeModals}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"></line>
						<line x1="6" y1="6" x2="18" y2="18"></line>
					</svg>
				</button>
			</div>

			<div class="modal-body">
				<div class="form-group">
					<label for="jobName">Name *</label>
					<input
						id="jobName"
						type="text"
						placeholder="Enter backup job name"
						bind:value={formData.name}
						class="form-input"
						required
					/>
				</div>

				<div class="form-row">
					<div class="form-group">
						<label for="jobType">Backup Type *</label>
						<select id="jobType" bind:value={formData.type} class="form-select">
							{#each typeOptions as option}
								<option value={option.value}>{option.label}</option>
							{/each}
						</select>
					</div>

					<div class="form-group">
						<label for="jobSchedule">Schedule *</label>
						<select id="jobSchedule" bind:value={formData.schedule} class="form-select">
							{#each scheduleOptions as option}
								<option value={option.value}>{option.label}</option>
							{/each}
						</select>
					</div>
				</div>

				<div class="form-row">
					<div class="form-group">
						<label for="retention">Retention (days) *</label>
						<input
							id="retention"
							type="number"
							min="1"
							max="365"
							bind:value={formData.retention_days}
							class="form-input"
						/>
					</div>

					<div class="form-group">
						<label for="destination">Destination *</label>
						<select id="destination" bind:value={formData.destination} class="form-select">
							{#each destinationOptions as option}
								<option value={option.value}>{option.label}</option>
							{/each}
						</select>
					</div>
				</div>
			</div>

			<div class="modal-footer">
				<button class="btn btn-outline" on:click={closeModals}>
					Cancel
				</button>
				<button class="btn btn-primary" on:click={showCreateModal ? createBackupJob : updateBackupJob}>
					{showCreateModal ? 'Create Job' : 'Update Job'}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.backup-page {
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

	.header-actions {
		display: flex;
		gap: var(--space-sm);
	}

	.section {
		margin-bottom: var(--space-2xl);
	}

	.section-header {
		margin-bottom: var(--space-lg);
	}

	.section-header h2 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.jobs-table,
	.history-table {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
	}

	.table-header {
		display: grid;
		gap: var(--space-lg);
		padding: var(--space-md) 0;
		border-bottom: 1px solid var(--border-color);
		margin-bottom: var(--space-lg);
	}

	.jobs-table .table-header {
		grid-template-columns: 1.5fr 1fr 1.2fr 1fr 1.2fr 1.2fr 0.8fr 0.8fr;
	}

	.history-table .table-header {
		grid-template-columns: 1.5fr 1.2fr 1fr 0.8fr 1fr 0.8fr;
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
		gap: var(--space-lg);
		padding: var(--space-md) 0;
		border-bottom: 1px solid var(--border-color);
		align-items: center;
	}

	.jobs-table .table-row {
		grid-template-columns: 1.5fr 1fr 1.2fr 1fr 1.2fr 1.2fr 0.8fr 0.8fr;
	}

	.history-table .table-row {
		grid-template-columns: 1.5fr 1.2fr 1fr 0.8fr 1fr 0.8fr;
	}

	.table-row:last-child {
		border-bottom: none;
	}

	.job-info {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.job-name {
		font-weight: 600;
		color: var(--text-primary);
		font-size: var(--text-sm);
	}

	.job-destination {
		font-size: var(--text-xs);
		color: var(--text-secondary);
		text-transform: capitalize;
	}

	.type-info {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
	}

	.type-icon {
		width: 20px;
		height: 20px;
		color: var(--text-secondary);
	}

	.type-icon svg {
		width: 100%;
		height: 100%;
	}

	.type-text {
		font-size: var(--text-sm);
		color: var(--text-primary);
		text-transform: capitalize;
	}

	.schedule,
	.last-run,
	.next-run,
	.size,
	.duration,
	.started-at {
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

	.status-active {
		background: var(--success-bg);
		color: var(--success-text);
	}

	.status-inactive {
		background: var(--warning-bg);
		color: var(--warning-text);
	}

	.status-running {
		background: var(--info-bg);
		color: var(--info-text);
	}

	.status-failed {
		background: var(--error-bg);
		color: var(--error-text);
	}

	.status-completed {
		background: var(--success-bg);
		color: var(--success-text);
	}

	.actions-dropdown {
		position: relative;
	}

	.actions-dropdown:hover .dropdown-menu {
		opacity: 1;
		visibility: visible;
		transform: translateY(0);
	}

	.dropdown-menu {
		position: absolute;
		top: 100%;
		right: 0;
		margin-top: var(--space-sm);
		min-width: 150px;
		padding: var(--space-sm);
		background: var(--bg-glass);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-lg);
		opacity: 0;
		visibility: hidden;
		transform: translateY(-10px);
		transition: all var(--transition-normal);
		z-index: 10;
	}

	.dropdown-item {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		padding: var(--space-sm);
		color: var(--text-primary);
		background: none;
		border: none;
		width: 100%;
		text-align: left;
		cursor: pointer;
		font-size: var(--text-sm);
		border-radius: var(--radius-md);
		transition: background var(--transition-fast);
	}

	.dropdown-item:hover {
		background: var(--bg-hover);
	}

	.dropdown-item.danger {
		color: var(--error);
	}

	.dropdown-item.danger:hover {
		background: var(--error-bg);
	}

	.dropdown-item svg {
		width: 16px;
		height: 16px;
	}

	.dropdown-divider {
		height: 1px;
		background: var(--border-color);
		margin: var(--space-sm) 0;
	}

	.backup-actions {
		display: flex;
		gap: var(--space-sm);
	}

	.btn-small {
		padding: var(--space-xs) var(--space-sm);
		font-size: var(--text-sm);
	}

	.error-message {
		color: var(--error);
		margin-top: var(--space-xs);
		cursor: help;
	}

	.error-message svg {
		width: 16px;
		height: 16px;
	}

	.progress-indicator {
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.spinner {
		width: 20px;
		height: 20px;
		border: 2px solid var(--border-color);
		border-top: 2px solid var(--primary);
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.text-muted {
		color: var(--text-tertiary);
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
		margin-bottom: var(--space-lg);
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
		padding: var(--space-lg);
	}

	.modal-content {
		background: var(--bg-primary);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
		box-shadow: var(--shadow-2xl);
		max-width: 500px;
		width: 100%;
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
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.modal-close {
		background: none;
		border: none;
		color: var(--text-secondary);
		cursor: pointer;
		padding: var(--space-sm);
		border-radius: var(--radius-md);
		transition: all var(--transition-fast);
	}

	.modal-close:hover {
		background: var(--bg-hover);
		color: var(--text-primary);
	}

	.modal-close svg {
		width: 20px;
		height: 20px;
	}

	.modal-body {
		padding: var(--space-xl);
	}

	.form-group {
		margin-bottom: var(--space-lg);
	}

	.form-group label {
		display: block;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-sm);
		font-size: var(--text-sm);
	}

	.form-input,
	.form-select {
		width: 100%;
		padding: var(--space-sm) var(--space-md);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		background: var(--bg-glass);
		color: var(--text-primary);
		font-size: var(--text-sm);
		transition: border-color var(--transition-fast);
	}

	.form-input:focus,
	.form-select:focus {
		outline: none;
		border-color: var(--primary);
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-lg);
	}

	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-sm);
		padding: var(--space-xl);
		border-top: 1px solid var(--border-color);
	}

	@media (max-width: 768px) {
		.backup-page {
			padding: var(--space-lg);
		}

		.header-content {
			flex-direction: column;
			align-items: stretch;
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

		.form-row {
			grid-template-columns: 1fr;
		}
	}
</style> 