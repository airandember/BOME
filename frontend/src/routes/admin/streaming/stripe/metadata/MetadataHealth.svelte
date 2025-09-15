<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';
	import { showToast } from '$lib/toast';

	// State variables
	let health = $state<any>(null);
	let loading = $state(true);
	let fixing = $state(false);
	let error = $state('');
	let lastChecked = $state<Date | null>(null);

	// Auto-refresh interval
	let refreshInterval: ReturnType<typeof setInterval> | undefined;

	onMount(() => {
		loadHealthData();
		
		// Auto-refresh every 30 seconds
		refreshInterval = setInterval(loadHealthData, 30000);
		
		return () => {
			if (refreshInterval) clearInterval(refreshInterval);
		};
	});

	async function loadHealthData() {
		try {
			loading = true;
			error = '';
			
			const response = await apiRequest('/admin/streaming/stripe/metadata/health');
			if (!response.ok) {
				throw new Error('Failed to load metadata health');
			}
			
			health = await response.json();
			lastChecked = new Date();
		} catch (err: any) {
			console.error('Error loading metadata health:', err);
			error = err.message || 'Failed to load health data';
			showToast('Failed to load metadata health', 'error');
		} finally {
			loading = false;
		}
	}

	async function previewFix() {
		try {
			fixing = true;
			
			const response = await apiRequest('/admin/streaming/stripe/metadata/fix?dry_run=true', {
				method: 'POST'
			});
			
			if (!response.ok) {
				throw new Error('Failed to preview fix');
			}
			
			const result = await response.json();
			
			showToast(`Preview: Would fix ${result.records_to_fix} records`, 'info');
			
			// Show detailed preview in console for debugging
			console.log('Fix preview:', result);
			
		} catch (err: any) {
			console.error('Error previewing fix:', err);
			showToast('Failed to preview fix', 'error');
		} finally {
			fixing = false;
		}
	}

	async function fixMetadata() {
		if (!confirm('Are you sure you want to fix all metadata corruption? This will update customer records.')) {
			return;
		}
		
		try {
			fixing = true;
			
			const response = await apiRequest('/admin/streaming/stripe/metadata/fix', {
				method: 'POST'
			});
			
			if (!response.ok) {
				throw new Error('Failed to fix metadata');
			}
			
			const result = await response.json();
			
			showToast('Metadata corruption fixed successfully!', 'success');
			
			// Reload health data to show updated status
			await loadHealthData();
			
		} catch (err: any) {
			console.error('Error fixing metadata:', err);
			showToast('Failed to fix metadata', 'error');
		} finally {
			fixing = false;
		}
	}

	function getStatusColor(status: string) {
		switch (status) {
			case 'healthy': return '#10b981'; // green
			case 'warning': return '#f59e0b'; // yellow
			case 'critical': return '#ef4444'; // red
			default: return '#6b7280'; // gray
		}
	}

	function getStatusIcon(status: string) {
		switch (status) {
			case 'healthy': return '✅';
			case 'warning': return '⚠️';
			case 'critical': return '🚨';
			default: return '❓';
		}
	}
</script>

<div class="metadata-health">
	<div class="health-header">
		<h2>Stripe Metadata Health</h2>
		<p>Monitor and fix customer metadata corruption</p>
		
		{#if lastChecked}
			<div class="last-checked">
				Last checked: {lastChecked.toLocaleTimeString()}
			</div>
		{/if}
	</div>

	{#if loading && !health}
		<div class="loading-state">
			<div class="spinner"></div>
			<p>Loading health data...</p>
		</div>
	{:else if error}
		<div class="error-state">
			<div class="error-icon">❌</div>
			<h3>Error Loading Health Data</h3>
			<p>{error}</p>
			<button class="retry-btn" onclick={loadHealthData}>
				Retry
			</button>
		</div>
	{:else if health}
		<!-- Health Overview -->
		<div class="health-overview">
			<div class="health-status" style="border-color: {getStatusColor(health.status)}">
				<div class="status-icon">{getStatusIcon(health.status)}</div>
				<div class="status-info">
					<h3>System Status: {health.status.charAt(0).toUpperCase() + health.status.slice(1)}</h3>
					<div class="health-percentage">
						{health.health.health_percentage.toFixed(1)}% Healthy
					</div>
				</div>
			</div>
		</div>

		<!-- Health Metrics -->
		<div class="health-metrics">
			<div class="metric-card">
				<div class="metric-value">{health.health.total_customers}</div>
				<div class="metric-label">Total Customers</div>
			</div>
			
			<div class="metric-card">
				<div class="metric-value">{health.health.customers_with_users}</div>
				<div class="metric-label">With Users</div>
			</div>
			
			<div class="metric-card success">
				<div class="metric-value">{health.health.correct_metadata}</div>
				<div class="metric-label">Correct Metadata</div>
			</div>
			
			<div class="metric-card warning">
				<div class="metric-value">{health.health.missing_metadata}</div>
				<div class="metric-label">Missing Metadata</div>
			</div>
			
			<div class="metric-card error">
				<div class="metric-value">{health.health.incorrect_metadata}</div>
				<div class="metric-label">Incorrect Metadata</div>
			</div>
			
			<div class="metric-card neutral">
				<div class="metric-value">{health.health.orphaned_customers}</div>
				<div class="metric-label">Orphaned Customers</div>
			</div>
		</div>

		<!-- Recommendations -->
		{#if health.recommendations && health.recommendations.length > 0}
			<div class="recommendations">
				<h3>Recommendations</h3>
				<ul>
					{#each health.recommendations as recommendation}
						<li>{recommendation}</li>
					{/each}
				</ul>
			</div>
		{/if}

		<!-- Action Buttons -->
		<div class="actions">
			<button 
				class="action-btn secondary" 
				onclick={previewFix}
				disabled={fixing || loading}
			>
				{fixing ? 'Previewing...' : 'Preview Fix'}
			</button>
			
			<button 
				class="action-btn primary" 
				onclick={fixMetadata}
				disabled={fixing || loading || health.health.missing_metadata + health.health.incorrect_metadata === 0}
			>
				{fixing ? 'Fixing...' : 'Fix Metadata'}
			</button>
			
			<button 
				class="action-btn secondary" 
				onclick={loadHealthData}
				disabled={loading}
			>
				{loading ? 'Refreshing...' : 'Refresh'}
			</button>
		</div>
	{/if}
</div>

<style>
	.metadata-health {
		padding: 24px;
		max-width: 1200px;
		margin: 0 auto;
	}

	.health-header {
		margin-bottom: 32px;
		text-align: center;
	}

	.health-header h2 {
		font-size: 28px;
		font-weight: 700;
		color: #1f2937;
		margin-bottom: 8px;
	}

	.health-header p {
		color: #6b7280;
		font-size: 16px;
		margin-bottom: 16px;
	}

	.last-checked {
		font-size: 14px;
		color: #9ca3af;
	}

	.loading-state, .error-state {
		text-align: center;
		padding: 48px 24px;
	}

	.spinner {
		width: 40px;
		height: 40px;
		border: 4px solid #f3f4f6;
		border-top: 4px solid #3b82f6;
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin: 0 auto 16px;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.error-icon {
		font-size: 48px;
		margin-bottom: 16px;
	}

	.retry-btn {
		background: #3b82f6;
		color: white;
		border: none;
		padding: 12px 24px;
		border-radius: 8px;
		cursor: pointer;
		font-weight: 500;
		margin-top: 16px;
	}

	.retry-btn:hover {
		background: #2563eb;
	}

	.health-overview {
		margin-bottom: 32px;
	}

	.health-status {
		display: flex;
		align-items: center;
		padding: 24px;
		background: white;
		border-radius: 12px;
		border: 3px solid;
		box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
	}

	.status-icon {
		font-size: 48px;
		margin-right: 24px;
	}

	.status-info h3 {
		font-size: 24px;
		font-weight: 700;
		color: #1f2937;
		margin-bottom: 8px;
	}

	.health-percentage {
		font-size: 20px;
		font-weight: 600;
		color: #6b7280;
	}

	.health-metrics {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: 16px;
		margin-bottom: 32px;
	}

	.metric-card {
		background: white;
		padding: 24px;
		border-radius: 12px;
		text-align: center;
		box-shadow: 0 2px 4px -1px rgba(0, 0, 0, 0.1);
		border: 2px solid #e5e7eb;
	}

	.metric-card.success {
		border-color: #10b981;
		background: #f0fdf4;
	}

	.metric-card.warning {
		border-color: #f59e0b;
		background: #fffbeb;
	}

	.metric-card.error {
		border-color: #ef4444;
		background: #fef2f2;
	}

	.metric-card.neutral {
		border-color: #6b7280;
		background: #f9fafb;
	}

	.metric-value {
		font-size: 32px;
		font-weight: 700;
		color: #1f2937;
		margin-bottom: 8px;
	}

	.metric-label {
		font-size: 14px;
		color: #6b7280;
		font-weight: 500;
	}

	.recommendations {
		background: #f8fafc;
		padding: 24px;
		border-radius: 12px;
		margin-bottom: 32px;
		border: 1px solid #e2e8f0;
	}

	.recommendations h3 {
		font-size: 18px;
		font-weight: 600;
		color: #1f2937;
		margin-bottom: 16px;
	}

	.recommendations ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.recommendations li {
		padding: 8px 0;
		color: #4b5563;
		position: relative;
		padding-left: 24px;
	}

	.recommendations li::before {
		content: '💡';
		position: absolute;
		left: 0;
	}

	.actions {
		display: flex;
		gap: 16px;
		justify-content: center;
		flex-wrap: wrap;
	}

	.action-btn {
		padding: 12px 24px;
		border-radius: 8px;
		font-weight: 600;
		cursor: pointer;
		border: none;
		transition: all 0.2s;
		min-width: 120px;
	}

	.action-btn.primary {
		background: #3b82f6;
		color: white;
	}

	.action-btn.primary:hover:not(:disabled) {
		background: #2563eb;
	}

	.action-btn.secondary {
		background: #f3f4f6;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.action-btn.secondary:hover:not(:disabled) {
		background: #e5e7eb;
	}

	.action-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
</style>
