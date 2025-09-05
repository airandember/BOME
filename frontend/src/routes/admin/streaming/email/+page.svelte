<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import EmailUsagePanel from '../stripe/components/EmailUsagePanel.svelte';

	// Type definitions
	interface EmailHistoryDay {
		date: string;
		resend: number;
		mailgun: number;
		success_rate: number;
	}

	interface EmailStats {
		today: {
			resend: number;
			mailgun: number;
			total: number;
		};
		history: EmailHistoryDay[];
	}

	interface EmailSettings {
		resend_daily_limit: number;
		mailgun_daily_limit: number;
		primary_provider: string;
		failover_enabled: boolean;
	}

	let loading = true;
	let error = '';
	let emailStats: EmailStats = {
		today: { resend: 0, mailgun: 0, total: 0 },
		history: []
	};
	let emailSettings: EmailSettings = {
		resend_daily_limit: 100,
		mailgun_daily_limit: 100,
		primary_provider: 'resend',
		failover_enabled: true
	};

	// Test email form
	let testEmail = '';
	let testSubject = 'Test Email from BOME';
	let testMessage = 'This is a test email to verify your email configuration is working correctly.';
	let sendingTest = false;

	onMount(async () => {
		await loadEmailData();
	});

	async function loadEmailData() {
		try {
			loading = true;
			
			// Load email usage stats and settings in parallel
			const [statsResponse, settingsResponse] = await Promise.all([
				apiRequest('/admin/email/usage'),
				apiRequest('/admin/email/settings')
			]);

			if (statsResponse.ok) {
				const statsData = await statsResponse.json();
				// Transform backend response to match frontend expectations
				const resendProvider = statsData.providers?.find((p: any) => p.provider === 'resend');
				const mailgunProvider = statsData.providers?.find((p: any) => p.provider === 'mailgun');
				
				emailStats = {
					today: {
						resend: resendProvider?.emails_sent || 0,
						mailgun: mailgunProvider?.emails_sent || 0,
						total: statsData.total_sent || 0
					},
					history: [] as EmailHistoryDay[] // TODO: Implement history endpoint
				};
			}

			if (settingsResponse.ok) {
				const settingsData = await settingsResponse.json();
				// Transform backend response to match frontend expectations
				const settings = settingsData.settings || {};
				emailSettings = {
					resend_daily_limit: parseInt(settings.daily_email_limit_resend) || 100,
					mailgun_daily_limit: parseInt(settings.daily_email_limit_mailgun) || 100,
					primary_provider: settings.email_provider_primary || 'resend',
					failover_enabled: settings.auto_failover_enabled === 'true'
				};
			}

		} catch (err: any) {
			error = err.message || 'Failed to load email data';
			console.error('Error loading email data:', err);
		} finally {
			loading = false;
		}
	}

	async function sendTestEmail() {
		if (!testEmail.trim()) {
			showToast('Please enter an email address', 'error');
			return;
		}

		try {
			sendingTest = true;
			
			const response = await apiRequest('/admin/email/test', {
				method: 'POST',
				body: JSON.stringify({
					email: testEmail,
					subject: testSubject,
					body: testMessage
				})
			});

			if (response.ok) {
				showToast('Test email sent successfully!', 'success');
				testEmail = '';
				// Reload stats to show updated usage
				await loadEmailData();
			} else {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Failed to send test email');
			}

		} catch (err: any) {
			showToast(err.message || 'Failed to send test email', 'error');
			console.error('Error sending test email:', err);
		} finally {
			sendingTest = false;
		}
	}

	async function updateSettings() {
		try {
			// Transform frontend settings to backend format
			const backendSettings = {
				daily_email_limit_resend: emailSettings.resend_daily_limit.toString(),
				daily_email_limit_mailgun: emailSettings.mailgun_daily_limit.toString(),
				auto_failover_enabled: emailSettings.failover_enabled.toString(),
				email_enabled: 'true'
			};

			const response = await apiRequest('/admin/email/settings', {
				method: 'PUT',
				body: JSON.stringify(backendSettings)
			});

			if (response.ok) {
				showToast('Email settings updated successfully!', 'success');
				await loadEmailData();
			} else {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Failed to update settings');
			}

		} catch (err: any) {
			showToast(err.message || 'Failed to update settings', 'error');
			console.error('Error updating settings:', err);
		}
	}

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function getUsagePercentage(used: number, limit: number): number {
		return limit > 0 ? Math.round((used / limit) * 100) : 0;
	}

	function getUsageColor(percentage: number): string {
		if (percentage >= 90) return 'text-red-600';
		if (percentage >= 75) return 'text-yellow-600';
		return 'text-green-600';
	}
</script>

<svelte:head>
	<title>Email Usage - Streaming Admin</title>
</svelte:head>

{#if loading}
	<div class="loading-container">
		<div class="spinner"></div>
		<p>Loading email usage data...</p>
	</div>
{:else if error}
	<div class="error-container">
		<div class="error-icon">⚠️</div>
		<h3>Error Loading Email Data</h3>
		<p>{error}</p>
		<button class="btn btn-primary" onclick={loadEmailData}>
			Try Again
		</button>
	</div>
{:else}
	<div class="email-dashboard">
		<!-- Overview Cards -->
		<div class="overview-grid">
			<div class="stat-card">
				<div class="stat-header">
					<h3>Today's Usage</h3>
					<div class="stat-icon">📧</div>
				</div>
				<div class="stat-content">
					<div class="stat-number">{emailStats.today.total}</div>
					<div class="stat-label">Emails Sent</div>
					<div class="stat-breakdown">
						<span>Resend: {emailStats.today.resend}</span>
						<span>Mailgun: {emailStats.today.mailgun}</span>
					</div>
				</div>
			</div>

			<div class="stat-card">
				<div class="stat-header">
					<h3>Resend Usage</h3>
					<div class="stat-icon">🚀</div>
				</div>
				<div class="stat-content">
					<div class="stat-number {getUsageColor(getUsagePercentage(emailStats.today.resend, emailSettings.resend_daily_limit))}">
						{emailStats.today.resend}/{emailSettings.resend_daily_limit}
					</div>
					<div class="stat-label">
						{getUsagePercentage(emailStats.today.resend, emailSettings.resend_daily_limit)}% Used
					</div>
					<div class="progress-bar">
						<div 
							class="progress-fill {getUsageColor(getUsagePercentage(emailStats.today.resend, emailSettings.resend_daily_limit))}"
							style="width: {Math.min(getUsagePercentage(emailStats.today.resend, emailSettings.resend_daily_limit), 100)}%"
						></div>
					</div>
				</div>
			</div>

			<div class="stat-card">
				<div class="stat-header">
					<h3>Mailgun Usage</h3>
					<div class="stat-icon">📮</div>
				</div>
				<div class="stat-content">
					<div class="stat-number {getUsageColor(getUsagePercentage(emailStats.today.mailgun, emailSettings.mailgun_daily_limit))}">
						{emailStats.today.mailgun}/{emailSettings.mailgun_daily_limit}
					</div>
					<div class="stat-label">
						{getUsagePercentage(emailStats.today.mailgun, emailSettings.mailgun_daily_limit)}% Used
					</div>
					<div class="progress-bar">
						<div 
							class="progress-fill {getUsageColor(getUsagePercentage(emailStats.today.mailgun, emailSettings.mailgun_daily_limit))}"
							style="width: {Math.min(getUsagePercentage(emailStats.today.mailgun, emailSettings.mailgun_daily_limit), 100)}%"
						></div>
					</div>
				</div>
			</div>

			<div class="stat-card">
				<div class="stat-header">
					<h3>Provider Status</h3>
					<div class="stat-icon">⚙️</div>
				</div>
				<div class="stat-content">
					<div class="provider-status">
						<div class="provider-item">
							<span class="provider-name">Primary:</span>
							<span class="provider-value">{emailSettings.primary_provider}</span>
						</div>
						<div class="provider-item">
							<span class="provider-name">Failover:</span>
							<span class="provider-value {emailSettings.failover_enabled ? 'text-green-600' : 'text-red-600'}">
								{emailSettings.failover_enabled ? 'Enabled' : 'Disabled'}
							</span>
						</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Usage History -->
		{#if emailStats.history && emailStats.history.length > 0}
			<div class="history-section">
				<h2>Usage History (Last 7 Days)</h2>
				<div class="history-table">
					<table>
						<thead>
							<tr>
								<th>Date</th>
								<th>Resend</th>
								<th>Mailgun</th>
								<th>Total</th>
								<th>Success Rate</th>
							</tr>
						</thead>
						<tbody>
							{#each emailStats.history as day}
								<tr>
									<td>{formatDate(day.date)}</td>
									<td>{day.resend || 0}</td>
									<td>{day.mailgun || 0}</td>
									<td>{(day.resend || 0) + (day.mailgun || 0)}</td>
									<td>
										<span class="success-rate {day.success_rate >= 95 ? 'text-green-600' : day.success_rate >= 90 ? 'text-yellow-600' : 'text-red-600'}">
											{day.success_rate || 100}%
										</span>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		{/if}

		<!-- Test Email Section -->
		<div class="test-section">
			<h2>Send Test Email</h2>
			<div class="test-form">
				<div class="form-row">
					<div class="form-group">
						<label for="test-email">Recipient Email</label>
						<input
							id="test-email"
							type="email"
							bind:value={testEmail}
							placeholder="Enter email address"
							class="form-input"
						/>
					</div>
					<div class="form-group">
						<label for="test-subject">Subject</label>
						<input
							id="test-subject"
							type="text"
							bind:value={testSubject}
							class="form-input"
						/>
					</div>
				</div>
				<div class="form-group">
					<label for="test-message">Message</label>
					<textarea
						id="test-message"
						bind:value={testMessage}
						rows="3"
						class="form-textarea"
					></textarea>
				</div>
				<button 
					class="btn btn-primary"
					onclick={sendTestEmail}
					disabled={sendingTest}
				>
					{#if sendingTest}
						<div class="loading-spinner small"></div>
						Sending...
					{:else}
						Send Test Email
					{/if}
				</button>
			</div>
		</div>

		<!-- Settings Section -->
		<div class="settings-section">
			<h2>Email Settings</h2>
			<div class="settings-form">
				<div class="form-row">
					<div class="form-group">
						<label for="resend-limit">Resend Daily Limit</label>
						<input
							id="resend-limit"
							type="number"
							bind:value={emailSettings.resend_daily_limit}
							min="1"
							max="10000"
							class="form-input"
						/>
					</div>
					<div class="form-group">
						<label for="mailgun-limit">Mailgun Daily Limit</label>
						<input
							id="mailgun-limit"
							type="number"
							bind:value={emailSettings.mailgun_daily_limit}
							min="1"
							max="10000"
							class="form-input"
						/>
					</div>
				</div>
				<div class="form-row">
					<div class="form-group">
						<label for="primary-provider">Primary Provider</label>
						<select
							id="primary-provider"
							bind:value={emailSettings.primary_provider}
							class="form-select"
						>
							<option value="resend">Resend</option>
							<option value="mailgun">Mailgun</option>
						</select>
					</div>
					<div class="form-group">
						<label class="checkbox-label">
							<input
								type="checkbox"
								bind:checked={emailSettings.failover_enabled}
								class="form-checkbox"
							/>
							Enable Automatic Failover
						</label>
					</div>
				</div>
				<button 
					class="btn btn-success"
					onclick={updateSettings}
				>
					Update Settings
				</button>
			</div>
		</div>

		<!-- Include the existing EmailUsagePanel component -->
		<div class="usage-panel-section">
			<h2>Detailed Usage Panel</h2>
			<EmailUsagePanel />
		</div>
	</div>
{/if}

<style>
	.loading-container, .error-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 400px;
		text-align: center;
	}

	.error-icon {
		font-size: 3rem;
		margin-bottom: 1rem;
	}

	.email-dashboard {
		max-width: 1200px;
		margin: 0 auto;
		padding: 0;
	}

	.overview-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
		gap: 1.5rem;
		margin-bottom: 2rem;
	}

	.stat-card {
		background: white;
		border-radius: 0.75rem;
		padding: 1.5rem;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
		border: 1px solid #e5e7eb;
	}

	.stat-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
	}

	.stat-header h3 {
		margin: 0;
		font-size: 1rem;
		font-weight: 600;
		color: #374151;
	}

	.stat-icon {
		font-size: 1.5rem;
	}

	.stat-content {
		text-align: center;
	}

	.stat-number {
		font-size: 2rem;
		font-weight: 700;
		color: #111827;
		margin-bottom: 0.25rem;
	}

	.stat-label {
		font-size: 0.875rem;
		color: #6b7280;
		margin-bottom: 0.5rem;
	}

	.stat-breakdown {
		display: flex;
		justify-content: space-between;
		font-size: 0.75rem;
		color: #9ca3af;
	}

	.progress-bar {
		width: 100%;
		height: 0.5rem;
		background: #f3f4f6;
		border-radius: 0.25rem;
		overflow: hidden;
		margin-top: 0.5rem;
	}

	.progress-fill {
		height: 100%;
		transition: width 0.3s ease;
		border-radius: 0.25rem;
	}

	.progress-fill.text-green-600 {
		background: #10b981;
	}

	.progress-fill.text-yellow-600 {
		background: #f59e0b;
	}

	.progress-fill.text-red-600 {
		background: #ef4444;
	}

	.provider-status {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.provider-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.provider-name {
		font-weight: 500;
		color: #6b7280;
	}

	.provider-value {
		font-weight: 600;
		text-transform: capitalize;
	}

	.history-section, .test-section, .settings-section, .usage-panel-section {
		background: white;
		border-radius: 0.75rem;
		padding: 1.5rem;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
		border: 1px solid #e5e7eb;
		margin-bottom: 2rem;
	}

	.history-section h2, .test-section h2, .settings-section h2, .usage-panel-section h2 {
		margin: 0 0 1.5rem 0;
		font-size: 1.25rem;
		font-weight: 600;
		color: #111827;
	}

	.history-table {
		overflow-x: auto;
	}

	.history-table table {
		width: 100%;
		border-collapse: collapse;
	}

	.history-table th,
	.history-table td {
		padding: 0.75rem;
		text-align: left;
		border-bottom: 1px solid #e5e7eb;
	}

	.history-table th {
		font-weight: 600;
		color: #374151;
		background: #f9fafb;
	}

	.success-rate {
		font-weight: 600;
	}

	.test-form, .settings-form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.form-group label {
		font-weight: 500;
		color: #374151;
		font-size: 0.875rem;
	}

	.checkbox-label {
		display: flex !important;
		flex-direction: row !important;
		align-items: center;
		gap: 0.5rem;
		cursor: pointer;
	}

	.form-input, .form-textarea, .form-select {
		padding: 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		transition: border-color 0.2s ease;
	}

	.form-input:focus, .form-textarea:focus, .form-select:focus {
		outline: none;
		border-color: #3b82f6;
		box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.25);
	}

	.form-checkbox {
		width: 1rem;
		height: 1rem;
		accent-color: #3b82f6;
	}

	.btn {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		text-decoration: none;
	}

	.btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-primary {
		background: #3b82f6;
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: #2563eb;
	}

	.btn-success {
		background: #10b981;
		color: white;
	}

	.btn-success:hover:not(:disabled) {
		background: #059669;
	}

	.loading-spinner {
		width: 20px;
		height: 20px;
		border: 2px solid transparent;
		border-top: 2px solid currentColor;
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	.loading-spinner.small {
		width: 16px;
		height: 16px;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.text-green-600 { color: #10b981; }
	.text-yellow-600 { color: #f59e0b; }
	.text-red-600 { color: #ef4444; }

	.spinner {
		width: 40px;
		height: 40px;
		border: 3px solid #e5e7eb;
		border-top: 3px solid #3b82f6;
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-bottom: 1rem;
	}

	@media (max-width: 768px) {
		.form-row {
			grid-template-columns: 1fr;
		}
		
		.overview-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
