<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let isLoading = true;
	let settings = {
		streaming: {
			quality: '1080p',
			bitrate: '5000',
			enableHLS: true,
			enableDASH: false
		},
		notifications: {
			emailAlerts: true,
			subscriptionUpdates: true,
			technicalIssues: true
		},
		payment: {
			currency: 'USD',
			taxRate: '0.00',
			autoRenewal: true
		}
	};

	onMount(() => {
		console.log('Settings page mounted');
		loadSettings();
	});

	async function loadSettings() {
		try {
			// TODO: Implement settings loading
			await new Promise(resolve => setTimeout(resolve, 1000)); // Simulate API call
		} catch (error) {
			console.error('Error loading settings:', error);
			showToast('Failed to load settings', 'error');
		} finally {
			isLoading = false;
		}
	}

	async function saveSettings() {
		try {
			// TODO: Implement settings saving
			await new Promise(resolve => setTimeout(resolve, 1000)); // Simulate API call
			showToast('Settings saved successfully', 'success');
		} catch (error) {
			console.error('Error saving settings:', error);
			showToast('Failed to save settings', 'error');
		}
	}
</script>

<svelte:head>
	<title>Settings - Streaming Admin</title>
	<meta name="description" content="Configure streaming platform settings" />
</svelte:head>

<div class="settings-page">
	<!-- Header -->
	<header class="page-header">
		<div class="header-content">
			<div class="header-left">
				<button class="btn btn-secondary" on:click={() => goto('/admin')}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="19" y1="12" x2="5" y2="12"></line>
						<polyline points="12,19 5,12 12,5"></polyline>
					</svg>
					Go Back to Main Dashboard
				</button>
				<h1>Streaming Settings</h1>
				<p>Configure your streaming platform preferences and system settings</p>
			</div>
			<div class="header-actions">
				<button class="btn btn-primary" on:click={saveSettings}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"></path>
						<polyline points="17,21 17,13 7,13 7,21"></polyline>
						<polyline points="7,3 7,8 15,8"></polyline>
					</svg>
					Save Settings
				</button>
			</div>
		</div>
	</header>

	<!-- Main Content -->
	<div class="main-content">
		{#if isLoading}
			<div class="loading-container">
				<LoadingSpinner />
				<p>Loading settings...</p>
			</div>
		{:else}
			<div class="settings-sections">
				<!-- Streaming Settings -->
				<div class="settings-section">
					<h2>Streaming Configuration</h2>
					<div class="settings-grid">
						<div class="setting-item">
							<label for="quality">Default Quality</label>
							<select id="quality" bind:value={settings.streaming.quality}>
								<option value="720p">720p</option>
								<option value="1080p">1080p</option>
								<option value="1440p">1440p</option>
								<option value="4K">4K</option>
							</select>
						</div>
						<div class="setting-item">
							<label for="bitrate">Default Bitrate (kbps)</label>
							<input type="number" id="bitrate" bind:value={settings.streaming.bitrate} min="1000" max="10000" />
						</div>
						<div class="setting-item">
							<label class="checkbox-label">
								<input type="checkbox" bind:checked={settings.streaming.enableHLS} />
								<span>Enable HLS Streaming</span>
							</label>
						</div>
						<div class="setting-item">
							<label class="checkbox-label">
								<input type="checkbox" bind:checked={settings.streaming.enableDASH} />
								<span>Enable DASH Streaming</span>
							</label>
						</div>
					</div>
				</div>

				<!-- Notification Settings -->
				<div class="settings-section">
					<h2>Notification Preferences</h2>
					<div class="settings-grid">
						<div class="setting-item">
							<label class="checkbox-label">
								<input type="checkbox" bind:checked={settings.notifications.emailAlerts} />
								<span>Email Alerts</span>
							</label>
						</div>
						<div class="setting-item">
							<label class="checkbox-label">
								<input type="checkbox" bind:checked={settings.notifications.subscriptionUpdates} />
								<span>Subscription Updates</span>
							</label>
						</div>
						<div class="setting-item">
							<label class="checkbox-label">
								<input type="checkbox" bind:checked={settings.notifications.technicalIssues} />
								<span>Technical Issues</span>
							</label>
						</div>
					</div>
				</div>

				<!-- Payment Settings -->
				<div class="settings-section">
					<h2>Payment Configuration</h2>
					<div class="settings-grid">
						<div class="setting-item">
							<label for="currency">Currency</label>
							<select id="currency" bind:value={settings.payment.currency}>
								<option value="USD">USD ($)</option>
								<option value="EUR">EUR (€)</option>
								<option value="GBP">GBP (£)</option>
								<option value="CAD">CAD (C$)</option>
							</select>
						</div>
						<div class="setting-item">
							<label for="taxRate">Tax Rate (%)</label>
							<input type="number" id="taxRate" bind:value={settings.payment.taxRate} min="0" max="50" step="0.01" />
						</div>
						<div class="setting-item">
							<label class="checkbox-label">
								<input type="checkbox" bind:checked={settings.payment.autoRenewal} />
								<span>Enable Auto-Renewal</span>
							</label>
						</div>
					</div>
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	.settings-page {
		padding: 2rem;
		max-width: 1200px;
		margin: 0 auto;
	}

	.page-header {
		margin-bottom: 2rem;
	}

	.header-content {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 2rem;
	}

	.header-left h1 {
		font-size: 2rem;
		font-weight: bold;
		margin: 1rem 0 0.5rem 0;
		color: #1f2937;
	}

	.header-left p {
		color: #6b7280;
		margin: 0;
	}

	.header-actions {
		display: flex;
		gap: 1rem;
	}

	.btn {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1.5rem;
		border-radius: 0.5rem;
		font-weight: 500;
		text-decoration: none;
		border: none;
		cursor: pointer;
		transition: all 0.2s;
	}

	.btn-primary {
		background-color: #2563eb;
		color: white;
	}

	.btn-primary:hover {
		background-color: #1d4ed8;
	}

	.btn-secondary {
		background-color: #f3f4f6;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-secondary:hover {
		background-color: #e5e7eb;
	}

	.btn svg {
		width: 1.25rem;
		height: 1.25rem;
	}

	.main-content {
		background: white;
		border-radius: 0.75rem;
		box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1);
		padding: 2rem;
	}

	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 4rem 2rem;
		color: #6b7280;
	}

	.settings-sections {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.settings-section {
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		padding: 1.5rem;
	}

	.settings-section h2 {
		font-size: 1.25rem;
		font-weight: 600;
		color: #1f2937;
		margin: 0 0 1rem 0;
	}

	.settings-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 1rem;
	}

	.setting-item {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.setting-item label {
		font-weight: 500;
		color: #374151;
	}

	.setting-item input,
	.setting-item select {
		padding: 0.5rem;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		font-size: 0.875rem;
	}

	.setting-item input:focus,
	.setting-item select:focus {
		outline: none;
		border-color: #2563eb;
		box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
	}

	.checkbox-label {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		cursor: pointer;
	}

	.checkbox-label input[type="checkbox"] {
		width: 1rem;
		height: 1rem;
		margin: 0;
	}

	.checkbox-label span {
		font-weight: 500;
		color: #374151;
	}
</style> 
