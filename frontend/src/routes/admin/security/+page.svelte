<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	interface SecurityMetrics {
		threats: {
			blocked_today: number;
			blocked_week: number;
			blocked_month: number;
			active_threats: number;
			threat_level: 'low' | 'medium' | 'high' | 'critical';
		};
		authentication: {
			login_attempts_today: number;
			failed_logins_today: number;
			success_rate: number;
			suspicious_activities: number;
			active_sessions: number;
		};
		system: {
			uptime: number;
			cpu_usage: number;
			memory_usage: number;
			disk_usage: number;
			response_time: number;
			error_rate: number;
		};
		api: {
			total_requests_today: number;
			rate_limited_requests: number;
			api_key_violations: number;
			active_api_keys: number;
		};
		incidents: Array<{
			id: string;
			type: 'security' | 'system' | 'authentication' | 'api';
			severity: 'low' | 'medium' | 'high' | 'critical';
			title: string;
			description: string;
			status: 'open' | 'investigating' | 'resolved';
			created_at: string;
			resolved_at?: string;
		}>;
	}

	let securityData: SecurityMetrics | null = null;
	let loading = true;
	let error = '';
	let refreshInterval: number;

	onMount(() => {
		loadSecurityData();
		
		// Auto-refresh every 30 seconds
		refreshInterval = setInterval(loadSecurityData, 30000);
		
		return () => {
			if (refreshInterval) clearInterval(refreshInterval);
		};
	});

	const loadSecurityData = async () => {
		try {
			if (!securityData) loading = true;
			error = '';
			
			// Mock data - replace with actual API call
			await new Promise(resolve => setTimeout(resolve, 800));
			
			securityData = {
				threats: {
					blocked_today: 23,
					blocked_week: 156,
					blocked_month: 678,
					active_threats: 2,
					threat_level: 'medium'
				},
				authentication: {
					login_attempts_today: 1247,
					failed_logins_today: 34,
					success_rate: 97.3,
					suspicious_activities: 5,
					active_sessions: 892
				},
				system: {
					uptime: 99.98,
					cpu_usage: 23.5,
					memory_usage: 67.2,
					disk_usage: 45.8,
					response_time: 145,
					error_rate: 0.02
				},
				api: {
					total_requests_today: 45678,
					rate_limited_requests: 234,
					api_key_violations: 12,
					active_api_keys: 45
				},
				incidents: [
					{
						id: 'INC-001',
						type: 'security',
						severity: 'high',
						title: 'Suspicious Login Pattern Detected',
						description: 'Multiple failed login attempts from IP 192.168.1.100',
						status: 'investigating',
						created_at: '2024-06-18T10:30:00Z'
					},
					{
						id: 'INC-002',
						type: 'system',
						severity: 'medium',
						title: 'High Memory Usage Alert',
						description: 'Server memory usage exceeded 80% threshold',
						status: 'resolved',
						created_at: '2024-06-18T09:15:00Z',
						resolved_at: '2024-06-18T09:45:00Z'
					},
					{
						id: 'INC-003',
						type: 'api',
						severity: 'low',
						title: 'API Rate Limit Exceeded',
						description: 'Client exceeded rate limit for video streaming API',
						status: 'resolved',
						created_at: '2024-06-18T08:20:00Z',
						resolved_at: '2024-06-18T08:25:00Z'
					}
				]
			};
		} catch (err) {
			error = 'Failed to load security data';
			console.error('Error loading security data:', err);
		} finally {
			loading = false;
		}
	};

	const formatUptime = (uptime: number): string => {
		return `${uptime.toFixed(2)}%`;
	};

	const formatResponseTime = (time: number): string => {
		return `${time}ms`;
	};

	const formatDate = (dateString: string): string => {
		return new Date(dateString).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	};

	const getThreatLevelColor = (level: string) => {
		switch (level) {
			case 'low': return 'success';
			case 'medium': return 'warning';
			case 'high': return 'error';
			case 'critical': return 'critical';
			default: return 'secondary';
		}
	};

	const getSeverityBadge = (severity: string) => {
		const severityConfig = {
			low: { text: 'Low', class: 'severity-low' },
			medium: { text: 'Medium', class: 'severity-medium' },
			high: { text: 'High', class: 'severity-high' },
			critical: { text: 'Critical', class: 'severity-critical' }
		};

		const config = severityConfig[severity as keyof typeof severityConfig] || { text: severity, class: 'severity-unknown' };
		
		return `<span class="severity-badge ${config.class}">${config.text}</span>`;
	};

	const getStatusBadge = (status: string) => {
		const statusConfig = {
			open: { text: 'Open', class: 'status-open' },
			investigating: { text: 'Investigating', class: 'status-investigating' },
			resolved: { text: 'Resolved', class: 'status-resolved' }
		};

		const config = statusConfig[status as keyof typeof statusConfig] || { text: status, class: 'status-unknown' };
		
		return `<span class="status-badge ${config.class}">${config.text}</span>`;
	};

	const handleIncidentClick = (incident: any) => {
		goto(`/admin/security/incidents/${incident.id}`);
	};

	const triggerMaintenanceMode = async () => {
		if (confirm('Are you sure you want to enable maintenance mode? This will temporarily disable access for all users.')) {
			try {
				// Mock API call
				await new Promise(resolve => setTimeout(resolve, 1000));
				showToast('Maintenance mode enabled', 'success');
			} catch (err) {
				showToast('Failed to enable maintenance mode', 'error');
			}
		}
	};

	const runSecurityScan = async () => {
		try {
			loading = true;
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 2000));
			showToast('Security scan completed successfully', 'success');
			await loadSecurityData();
		} catch (err) {
			showToast('Security scan failed', 'error');
		} finally {
			loading = false;
		}
	};
</script>

<svelte:head>
	<title>Security & System Management - Admin Dashboard</title>
</svelte:head>

<div class="security-page">
	<div class="page-header">
		<div class="header-content">
			<div class="header-text">
				<h1>Security & System Management</h1>
				<p>Monitor system security, health, and manage incidents</p>
			</div>
			
			<div class="header-actions">
				<button class="btn btn-outline" on:click={runSecurityScan}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
						<path d="M9 12l2 2 4-4"></path>
					</svg>
					Run Security Scan
				</button>
				<button class="btn btn-error" on:click={triggerMaintenanceMode}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
						<line x1="12" y1="9" x2="12" y2="13"></line>
						<line x1="12" y1="17" x2="12.01" y2="17"></line>
					</svg>
					Maintenance Mode
				</button>
			</div>
		</div>
	</div>

	{#if loading && !securityData}
		<div class="loading-container">
			<LoadingSpinner size="large" color="primary" />
			<p>Loading security data...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<p class="error-message">{error}</p>
			<button class="btn btn-primary" on:click={loadSecurityData}>
				Try Again
			</button>
		</div>
	{:else if securityData}
		<div class="security-dashboard">
			<!-- Security Overview -->
			<div class="section-header">
				<h2>Security Overview</h2>
				<div class="threat-level threat-level-{securityData.threats.threat_level}">
					Threat Level: {securityData.threats.threat_level.toUpperCase()}
				</div>
			</div>
			
			<div class="metrics-grid">
				<div class="metric-card glass">
					<div class="metric-header">
						<div class="metric-icon threats">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
								<path d="M9 12l2 2 4-4"></path>
							</svg>
						</div>
						<h3>Threats Blocked</h3>
					</div>
					<div class="metric-value">{securityData.threats.blocked_today}</div>
					<div class="metric-details">
						<span class="metric-sub">{securityData.threats.blocked_month} this month</span>
						<span class="metric-trend {securityData.threats.active_threats > 0 ? 'negative' : 'positive'}">
							{securityData.threats.active_threats} active
						</span>
					</div>
				</div>

				<div class="metric-card glass">
					<div class="metric-header">
						<div class="metric-icon auth">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
								<circle cx="12" cy="16" r="1"></circle>
								<path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
							</svg>
						</div>
						<h3>Authentication</h3>
					</div>
					<div class="metric-value">{securityData.authentication.success_rate.toFixed(1)}%</div>
					<div class="metric-details">
						<span class="metric-sub">{securityData.authentication.failed_logins_today} failed today</span>
						<span class="metric-trend {securityData.authentication.suspicious_activities > 0 ? 'negative' : 'positive'}">
							{securityData.authentication.suspicious_activities} suspicious
						</span>
					</div>
				</div>

				<div class="metric-card glass">
					<div class="metric-header">
						<div class="metric-icon system">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect>
								<line x1="8" y1="21" x2="16" y2="21"></line>
								<line x1="12" y1="17" x2="12" y2="21"></line>
							</svg>
						</div>
						<h3>System Health</h3>
					</div>
					<div class="metric-value">{formatUptime(securityData.system.uptime)}</div>
					<div class="metric-details">
						<span class="metric-sub">CPU: {securityData.system.cpu_usage}%</span>
						<span class="metric-sub">Memory: {securityData.system.memory_usage}%</span>
					</div>
				</div>

				<div class="metric-card glass">
					<div class="metric-header">
						<div class="metric-icon api">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"></path>
								<rect x="8" y="2" width="8" height="4" rx="1" ry="1"></rect>
							</svg>
						</div>
						<h3>API Security</h3>
					</div>
					<div class="metric-value">{securityData.api.active_api_keys}</div>
					<div class="metric-details">
						<span class="metric-sub">{securityData.api.rate_limited_requests} rate limited</span>
						<span class="metric-trend {securityData.api.api_key_violations > 0 ? 'negative' : 'positive'}">
							{securityData.api.api_key_violations} violations
						</span>
					</div>
				</div>
			</div>

			<!-- System Performance -->
			<div class="section-header">
				<h2>System Performance</h2>
			</div>
			
			<div class="performance-grid">
				<div class="performance-card glass">
					<h3>CPU Usage</h3>
					<div class="progress-ring">
						<div class="progress-value">{securityData.system.cpu_usage}%</div>
						<div class="progress-bar">
							<div class="progress-fill" style="width: {securityData.system.cpu_usage}%"></div>
						</div>
					</div>
				</div>

				<div class="performance-card glass">
					<h3>Memory Usage</h3>
					<div class="progress-ring">
						<div class="progress-value">{securityData.system.memory_usage}%</div>
						<div class="progress-bar">
							<div class="progress-fill" style="width: {securityData.system.memory_usage}%"></div>
						</div>
					</div>
				</div>

				<div class="performance-card glass">
					<h3>Disk Usage</h3>
					<div class="progress-ring">
						<div class="progress-value">{securityData.system.disk_usage}%</div>
						<div class="progress-bar">
							<div class="progress-fill" style="width: {securityData.system.disk_usage}%"></div>
						</div>
					</div>
				</div>

				<div class="performance-card glass">
					<h3>Response Time</h3>
					<div class="progress-ring">
						<div class="progress-value">{formatResponseTime(securityData.system.response_time)}</div>
						<div class="progress-description">Average response time</div>
					</div>
				</div>
			</div>

			<!-- Quick Actions -->
			<div class="section-header">
				<h2>Quick Actions</h2>
			</div>
			
			<div class="actions-grid">
				<button class="action-card glass" on:click={() => goto('/admin/security/audit-logs')}>
					<div class="action-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
							<polyline points="14,2 14,8 20,8"></polyline>
							<line x1="16" y1="13" x2="8" y2="13"></line>
							<line x1="16" y1="17" x2="8" y2="17"></line>
						</svg>
					</div>
					<h3>Audit Logs</h3>
					<p>View system audit logs and user activities</p>
				</button>

				<button class="action-card glass" on:click={() => goto('/admin/security/api-keys')}>
					<div class="action-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"></path>
						</svg>
					</div>
					<h3>API Keys</h3>
					<p>Manage API keys and access permissions</p>
				</button>

				<button class="action-card glass" on:click={() => goto('/admin/security/backup')}>
					<div class="action-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
							<polyline points="7,10 12,15 17,10"></polyline>
							<line x1="12" y1="15" x2="12" y2="3"></line>
						</svg>
					</div>
					<h3>Backup Management</h3>
					<p>Configure and monitor system backups</p>
				</button>

				<button class="action-card glass" on:click={() => goto('/admin/security/incidents')}>
					<div class="action-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
							<line x1="12" y1="9" x2="12" y2="13"></line>
							<line x1="12" y1="17" x2="12.01" y2="17"></line>
						</svg>
					</div>
					<h3>Incident Reports</h3>
					<p>View and manage security incidents</p>
				</button>
			</div>

			<!-- Recent Incidents -->
			<div class="section-header">
				<h2>Recent Incidents</h2>
				<button class="btn btn-ghost" on:click={() => goto('/admin/security/incidents')}>
					View All
				</button>
			</div>
			
			<div class="incidents-table glass">
				<div class="table-header">
					<div class="header-cell">Incident</div>
					<div class="header-cell">Type</div>
					<div class="header-cell">Severity</div>
					<div class="header-cell">Status</div>
					<div class="header-cell">Created</div>
				</div>

				{#each securityData.incidents as incident}
					<div class="table-row" on:click={() => handleIncidentClick(incident)}>
						<div class="table-cell">
							<div class="incident-info">
								<span class="incident-title">{incident.title}</span>
								<span class="incident-id">#{incident.id}</span>
							</div>
						</div>
						<div class="table-cell">
							<span class="incident-type">{incident.type}</span>
						</div>
						<div class="table-cell">
							{@html getSeverityBadge(incident.severity)}
						</div>
						<div class="table-cell">
							{@html getStatusBadge(incident.status)}
						</div>
						<div class="table-cell">
							<span class="incident-date">{formatDate(incident.created_at)}</span>
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>

<style>
	.security-page {
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

	.security-dashboard {
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

	.threat-level {
		padding: var(--space-sm) var(--space-md);
		border-radius: var(--radius-md);
		font-size: var(--text-sm);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.threat-level-low {
		background: var(--success-bg);
		color: var(--success-text);
	}

	.threat-level-medium {
		background: var(--warning-bg);
		color: var(--warning-text);
	}

	.threat-level-high {
		background: var(--error-bg);
		color: var(--error-text);
	}

	.threat-level-critical {
		background: linear-gradient(135deg, #dc2626 0%, #991b1b 100%);
		color: var(--white);
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

	.metric-icon.threats {
		background: var(--error-gradient);
		color: var(--white);
	}

	.metric-icon.auth {
		background: var(--warning-gradient);
		color: var(--white);
	}

	.metric-icon.system {
		background: var(--success-gradient);
		color: var(--white);
	}

	.metric-icon.api {
		background: var(--info-gradient);
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

	/* Performance Grid */
	.performance-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: var(--space-lg);
		margin-bottom: var(--space-2xl);
	}

	.performance-card {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
		text-align: center;
	}

	.performance-card h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-lg);
	}

	.progress-ring {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-md);
	}

	.progress-value {
		font-size: var(--text-2xl);
		font-weight: 700;
		color: var(--primary);
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

	.progress-description {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	/* Actions Grid */
	.actions-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: var(--space-lg);
		margin-bottom: var(--space-2xl);
	}

	.action-card {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
		text-align: center;
		cursor: pointer;
		transition: all var(--transition-normal);
		background: none;
		color: inherit;
	}

	.action-card:hover {
		transform: translateY(-4px);
		box-shadow: var(--shadow-lg);
	}

	.action-icon {
		width: 60px;
		height: 60px;
		background: var(--primary-gradient);
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 0 auto var(--space-lg);
		box-shadow: var(--shadow-md);
	}

	.action-icon svg {
		width: 30px;
		height: 30px;
		color: var(--white);
	}

	.action-card h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-sm);
	}

	.action-card p {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		margin: 0;
	}

	/* Incidents Table */
	.incidents-table {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
	}

	.table-header {
		display: grid;
		grid-template-columns: 2fr 1fr 1fr 1fr 1fr;
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
		grid-template-columns: 2fr 1fr 1fr 1fr 1fr;
		gap: var(--space-lg);
		padding: var(--space-md) 0;
		border-bottom: 1px solid var(--border-color);
		align-items: center;
		cursor: pointer;
		transition: all var(--transition-normal);
		border-radius: var(--radius-md);
	}

	.table-row:hover {
		background: var(--bg-glass);
	}

	.table-row:last-child {
		border-bottom: none;
	}

	.incident-info {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.incident-title {
		font-weight: 600;
		color: var(--text-primary);
	}

	.incident-id {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.incident-type {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		text-transform: capitalize;
	}

	.severity-badge,
	.status-badge {
		display: inline-block;
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-md);
		font-size: var(--text-xs);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.severity-low {
		background: var(--success-bg);
		color: var(--success-text);
	}

	.severity-medium {
		background: var(--warning-bg);
		color: var(--warning-text);
	}

	.severity-high {
		background: var(--error-bg);
		color: var(--error-text);
	}

	.severity-critical {
		background: linear-gradient(135deg, #dc2626 0%, #991b1b 100%);
		color: var(--white);
	}

	.status-open {
		background: var(--error-bg);
		color: var(--error-text);
	}

	.status-investigating {
		background: var(--warning-bg);
		color: var(--warning-text);
	}

	.status-resolved {
		background: var(--success-bg);
		color: var(--success-text);
	}

	.incident-date {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	@media (max-width: 768px) {
		.security-page {
			padding: var(--space-lg);
		}

		.header-content {
			flex-direction: column;
			align-items: stretch;
		}

		.header-actions {
			justify-content: space-between;
		}

		.metrics-grid {
			grid-template-columns: 1fr;
		}

		.performance-grid {
			grid-template-columns: repeat(2, 1fr);
		}

		.actions-grid {
			grid-template-columns: 1fr;
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
	}
</style> 