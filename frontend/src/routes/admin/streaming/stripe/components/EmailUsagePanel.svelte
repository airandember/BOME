<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';
	import { showToast } from '$lib/toast';

	interface EmailUsageStats {
		provider: string;
		emails_sent: number;
		emails_failed: number;
		daily_limit: number;
		remaining: number;
		usage_percent: number;
	}

	interface EmailUsageResponse {
		date: string;
		providers: EmailUsageStats[];
		total_sent: number;
		total_failed: number;
		total_remaining: number;
		total_limit: number;
		overall_percent: number;
		monthly_sent: number;
		monthly_limit: number;
		monthly_percent: number;
	}

	let usageData: EmailUsageResponse | null = null;
	let loading = true;
	let error = '';
	let refreshing = false;

	onMount(() => {
		loadEmailUsage();
	});

	async function loadEmailUsage() {
		try {
			loading = true;
			error = '';
			
			const response = await apiRequest('/admin/email/usage');
			if (!response.ok) {
				throw new Error('Failed to load email usage data');
			}
			
			usageData = await response.json();
		} catch (err: any) {
			error = err.message || 'Failed to load email usage data';
			console.error('Error loading email usage:', err);
		} finally {
			loading = false;
		}
	}

	async function refreshUsage() {
		refreshing = true;
		await loadEmailUsage();
		refreshing = false;
		showToast('Email usage data refreshed', 'success');
	}

	function getProviderIcon(provider: string): string {
		switch (provider) {
			case 'resend': return '🚀';
			case 'smtp': return '📮';
			default: return '📨';
		}
	}

	function getUsageColor(percent: number): string {
		if (percent >= 90) return 'text-red-600';
		if (percent >= 75) return 'text-yellow-600';
		if (percent >= 50) return 'text-blue-600';
		return 'text-green-600';
	}

	function getProgressBarColor(percent: number): string {
		if (percent >= 90) return 'bg-red-500';
		if (percent >= 75) return 'bg-yellow-500';
		if (percent >= 50) return 'bg-blue-500';
		return 'bg-green-500';
	}
</script>

<div class="email-usage-panel">
	<div class="panel-header">
		<h3>🚀 Resend Usage Tracking</h3>
		<button 
			class="btn btn-sm btn-outline"
			onclick={refreshUsage}
			disabled={refreshing}
		>
			{#if refreshing}
				<div class="loading-spinner small"></div>
				Refreshing...
			{:else}
				🔄 Refresh
			{/if}
		</button>
	</div>

	{#if loading}
		<div class="loading-container">
			<div class="loading-spinner"></div>
			<p>Loading email usage data...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<p class="error-message">{error}</p>
			<button class="btn btn-primary btn-sm" onclick={loadEmailUsage}>
				Try Again
			</button>
		</div>
	{:else if usageData}
		<!-- Overall Stats -->
		<div class="stats-overview">
			<div class="stat-card">
				<div class="stat-icon">📧</div>
				<div class="stat-content">
					<div class="stat-value">{usageData.total_sent}</div>
					<div class="stat-label">Emails Sent Today</div>
				</div>
			</div>
			
			<div class="stat-card">
				<div class="stat-icon">📈</div>
				<div class="stat-content">
					<div class="stat-value {getUsageColor(usageData.overall_percent)}">{usageData.overall_percent}%</div>
					<div class="stat-label">Daily Usage</div>
				</div>
			</div>
			
			<div class="stat-card">
				<div class="stat-icon">📊</div>
				<div class="stat-content">
					<div class="stat-value {getUsageColor(usageData.monthly_percent || 0)}">{usageData.monthly_percent || 0}%</div>
					<div class="stat-label">Monthly Usage</div>
				</div>
			</div>
			
			<div class="stat-card">
				<div class="stat-icon">📮</div>
				<div class="stat-content">
					<div class="stat-value">{usageData.total_remaining}</div>
					<div class="stat-label">Remaining Today</div>
				</div>
			</div>
		</div>

		<!-- Resend Provider Details -->
		<div class="providers-section">
			<h4>Resend Provider Details</h4>
			<div class="providers-grid">
				{#each usageData.providers.filter(p => p.provider === 'resend') as provider}
					<div class="provider-card">
						<div class="provider-header">
							<span class="provider-icon">{getProviderIcon(provider.provider)}</span>
							<span class="provider-name">RESEND</span>
							<span class="provider-status {getUsageColor(provider.usage_percent)}">
								{provider.usage_percent}%
							</span>
						</div>
						
						<div class="usage-bar">
							<div 
								class="usage-fill {getProgressBarColor(provider.usage_percent)}"
								style="width: {provider.usage_percent}%"
							></div>
						</div>
						
						<div class="provider-stats">
							<div class="provider-stat">
								<span class="stat-label">Sent Today:</span>
								<span class="stat-value">{provider.emails_sent}</span>
							</div>
							<div class="provider-stat">
								<span class="stat-label">Failed:</span>
								<span class="stat-value text-red-600">{provider.emails_failed}</span>
							</div>
							<div class="provider-stat">
								<span class="stat-label">Remaining:</span>
								<span class="stat-value">{provider.remaining}</span>
							</div>
							<div class="provider-stat">
								<span class="stat-label">Daily Limit:</span>
								<span class="stat-value">{provider.daily_limit}</span>
							</div>
						</div>
					</div>
				{/each}
				
				<!-- Monthly Usage Card -->
				<div class="provider-card">
					<div class="provider-header">
						<span class="provider-icon">📊</span>
						<span class="provider-name">MONTHLY USAGE</span>
						<span class="provider-status {getUsageColor(usageData.monthly_percent || 0)}">
							{usageData.monthly_percent || 0}%
						</span>
					</div>
					
					<div class="usage-bar">
						<div 
							class="usage-fill {getProgressBarColor(usageData.monthly_percent || 0)}"
							style="width: {usageData.monthly_percent || 0}%"
						></div>
					</div>
					
					<div class="provider-stats">
						<div class="provider-stat">
							<span class="stat-label">Sent This Month:</span>
							<span class="stat-value">{usageData.monthly_sent || 0}</span>
						</div>
						<div class="provider-stat">
							<span class="stat-label">Monthly Limit:</span>
							<span class="stat-value">{usageData.monthly_limit || 3000}</span>
						</div>
						<div class="provider-stat">
							<span class="stat-label">Remaining:</span>
							<span class="stat-value">{(usageData.monthly_limit || 3000) - (usageData.monthly_sent || 0)}</span>
						</div>
						<div class="provider-stat">
							<span class="stat-label">Resets:</span>
							<span class="stat-value">Monthly</span>
						</div>
					</div>
				</div>
			</div>
		</div>

		<!-- System Status -->
		<div class="system-status">
			<h4>System Status</h4>
			<div class="status-grid">
				<div class="status-item">
					<span class="status-label">Date:</span>
					<span class="status-value">{usageData.date}</span>
				</div>
				<div class="status-item">
					<span class="status-label">Email Provider:</span>
					<span class="status-value text-green-600">🚀 Resend</span>
				</div>
				<div class="status-item">
					<span class="status-label">Service Status:</span>
					<span class="status-value text-green-600">✅ Active</span>
				</div>
				<div class="status-item">
					<span class="status-label">Free Tier:</span>
					<span class="status-value text-blue-600">💎 3K/month</span>
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	.email-usage-panel {
		background: var(--card-bg);
		border-radius: 16px;
		padding: 1.5rem;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
		margin-bottom: 2rem;
	}

	.panel-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1.5rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid var(--border-color);
	}

	.panel-header h3 {
		margin: 0;
		color: var(--text-primary);
		font-size: 1.25rem;
		font-weight: 600;
	}

	.loading-container,
	.error-container {
		text-align: center;
		padding: 2rem 0;
	}

	.error-message {
		color: var(--error-text);
		margin-bottom: 1rem;
	}

	.stats-overview {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
		margin-bottom: 2rem;
	}

	.stat-card {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1rem;
		background: var(--bg-secondary);
		border-radius: 12px;
		border: 1px solid var(--border-color);
	}

	.stat-icon {
		font-size: 2rem;
		opacity: 0.8;
	}

	.stat-content {
		flex: 1;
	}

	.stat-value {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 0.25rem;
	}

	.stat-label {
		font-size: 0.875rem;
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.providers-section {
		margin-bottom: 2rem;
	}

	.providers-section h4 {
		margin: 0 0 1rem 0;
		color: var(--text-primary);
		font-size: 1.1rem;
		font-weight: 600;
	}

	.providers-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: 1rem;
	}

	.provider-card {
		padding: 1rem;
		background: var(--bg-secondary);
		border-radius: 12px;
		border: 1px solid var(--border-color);
	}

	.provider-header {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 1rem;
	}

	.provider-icon {
		font-size: 1.25rem;
	}

	.provider-name {
		flex: 1;
		font-weight: 600;
		color: var(--text-primary);
	}

	.provider-status {
		font-weight: 700;
		font-size: 0.875rem;
	}

	.usage-bar {
		height: 8px;
		background: var(--bg-tertiary);
		border-radius: 4px;
		overflow: hidden;
		margin-bottom: 1rem;
	}

	.usage-fill {
		height: 100%;
		transition: width 0.3s ease;
		border-radius: 4px;
	}

	.provider-stats {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 0.5rem;
	}

	.provider-stat {
		display: flex;
		justify-content: space-between;
		align-items: center;
		font-size: 0.875rem;
	}

	.provider-stat .stat-label {
		color: var(--text-secondary);
		font-size: 0.75rem;
		text-transform: none;
		letter-spacing: normal;
	}

	.provider-stat .stat-value {
		font-weight: 600;
		color: var(--text-primary);
		font-size: 0.875rem;
	}

	.system-status h4 {
		margin: 0 0 1rem 0;
		color: var(--text-primary);
		font-size: 1.1rem;
		font-weight: 600;
	}

	.status-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
	}

	.status-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.75rem;
		background: var(--bg-secondary);
		border-radius: 8px;
		border: 1px solid var(--border-color);
	}

	.status-label {
		font-size: 0.875rem;
		color: var(--text-secondary);
	}

	.status-value {
		font-weight: 600;
		color: var(--text-primary);
	}

	.btn {
		padding: 0.5rem 1rem;
		border: none;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
	}

	.btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-outline {
		background: white;
		color: var(--text-primary);
		border: 1px solid var(--border-color);
	}

	.btn-outline:hover:not(:disabled) {
		background: var(--bg-secondary);
		border-color: var(--primary-color);
	}

	.btn-primary {
		background: var(--primary-color);
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: var(--primary-color-dark);
		transform: translateY(-1px);
	}

	.btn-sm {
		padding: 0.375rem 0.75rem;
		font-size: 0.75rem;
	}

	.loading-spinner {
		width: 40px;
		height: 40px;
		border: 3px solid var(--border-color);
		border-top: 3px solid var(--primary-color);
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin: 0 auto 1rem auto;
	}

	.loading-spinner.small {
		width: 16px;
		height: 16px;
		border-width: 2px;
		margin: 0;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.text-red-600 { color: #dc2626; }
	.text-yellow-600 { color: #d97706; }
	.text-blue-600 { color: #2563eb; }
	.text-green-600 { color: #059669; }

	.bg-red-500 { background-color: #ef4444; }
	.bg-yellow-500 { background-color: #f59e0b; }
	.bg-blue-500 { background-color: #3b82f6; }
	.bg-green-500 { background-color: #10b981; }

	@media (max-width: 768px) {
		.stats-overview {
			grid-template-columns: 1fr;
		}
		
		.providers-grid {
			grid-template-columns: 1fr;
		}
		
		.status-grid {
			grid-template-columns: 1fr;
		}
		
		.provider-stats {
			grid-template-columns: 1fr;
		}
	}
</style>
