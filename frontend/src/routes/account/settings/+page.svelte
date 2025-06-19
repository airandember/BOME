<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let user: any = null;
	let loading = false;
	let saving = false;
	let isAuthenticated = false;

	// Settings data
	let settings = {
		emailNotifications: {
			newVideos: true,
			subscriptionUpdates: true,
			promotions: false,
			systemUpdates: true
		},
		privacy: {
			profileVisibility: 'public',
			showWatchHistory: false,
			allowRecommendations: true
		},
		preferences: {
			autoplay: true,
			quality: 'auto',
			language: 'en',
			theme: 'auto'
		}
	};

	let originalSettings = JSON.parse(JSON.stringify(settings));
	let hasChanges = false;

	// Subscribe to auth store
	auth.subscribe(state => {
		isAuthenticated = state.isAuthenticated;
		user = state.user;
	});

	onMount(async () => {
		if (!isAuthenticated) {
			goto('/login');
			return;
		}

		await loadSettings();
	});

	const loadSettings = async () => {
		try {
			loading = true;
			// Mock loading settings - replace with actual API call
			await new Promise(resolve => setTimeout(resolve, 500));
			
			// For now, use default settings
			originalSettings = JSON.parse(JSON.stringify(settings));
		} catch (err) {
			showToast('Failed to load settings', 'error');
			console.error('Error loading settings:', err);
		} finally {
			loading = false;
		}
	};

	const handleSettingChange = () => {
		hasChanges = JSON.stringify(settings) !== JSON.stringify(originalSettings);
	};

	const handleSave = async () => {
		if (!hasChanges) return;

		try {
			saving = true;
			
			// Mock saving settings - replace with actual API call
			await new Promise(resolve => setTimeout(resolve, 1000));
			
			originalSettings = JSON.parse(JSON.stringify(settings));
			hasChanges = false;
			showToast('Settings saved successfully', 'success');
		} catch (err) {
			showToast('Failed to save settings', 'error');
			console.error('Error saving settings:', err);
		} finally {
			saving = false;
		}
	};

	const handleReset = () => {
		settings = JSON.parse(JSON.stringify(originalSettings));
		hasChanges = false;
	};

	const handleBack = () => {
		if (hasChanges) {
			if (confirm('You have unsaved changes. Are you sure you want to leave?')) {
				goto('/account');
			}
		} else {
			goto('/account');
		}
	};

	const handleDeleteAccount = () => {
		if (confirm('Are you sure you want to delete your account? This action cannot be undone.')) {
			if (confirm('This will permanently delete all your data. Are you absolutely sure?')) {
				showToast('Account deletion is not implemented yet', 'warning');
			}
		}
	};
</script>

<svelte:head>
	<title>Account Settings - BOME</title>
	<meta name="description" content="Manage your BOME account settings and preferences" />
</svelte:head>

<div class="settings-page">
	<div class="container">
		<header class="page-header">
			<button class="back-button" on:click={handleBack}>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M19 12H5"></path>
					<path d="M12 19l-7-7 7-7"></path>
				</svg>
				Back to Account
			</button>
			<h1>Account Settings</h1>
			<p>Manage your preferences and account options</p>
		</header>

		{#if loading}
			<div class="loading-container">
				<LoadingSpinner />
				<p>Loading your settings...</p>
			</div>
		{:else}
			<div class="settings-content">
				<!-- Email Notifications -->
				<div class="settings-section">
					<h2>Email Notifications</h2>
					<p class="section-description">Choose what email notifications you'd like to receive</p>
					
					<div class="settings-group">
						<div class="setting-item">
							<div class="setting-info">
								<h3>New Video Releases</h3>
								<p>Get notified when new videos are published</p>
							</div>
							<label class="toggle">
								<input 
									type="checkbox" 
									bind:checked={settings.emailNotifications.newVideos}
									on:change={handleSettingChange}
								/>
								<span class="slider"></span>
							</label>
						</div>

						<div class="setting-item">
							<div class="setting-info">
								<h3>Subscription Updates</h3>
								<p>Billing reminders and subscription changes</p>
							</div>
							<label class="toggle">
								<input 
									type="checkbox" 
									bind:checked={settings.emailNotifications.subscriptionUpdates}
									on:change={handleSettingChange}
								/>
								<span class="slider"></span>
							</label>
						</div>

						<div class="setting-item">
							<div class="setting-info">
								<h3>Promotions & Offers</h3>
								<p>Special offers and promotional content</p>
							</div>
							<label class="toggle">
								<input 
									type="checkbox" 
									bind:checked={settings.emailNotifications.promotions}
									on:change={handleSettingChange}
								/>
								<span class="slider"></span>
							</label>
						</div>

						<div class="setting-item">
							<div class="setting-info">
								<h3>System Updates</h3>
								<p>Important system notifications and updates</p>
							</div>
							<label class="toggle">
								<input 
									type="checkbox" 
									bind:checked={settings.emailNotifications.systemUpdates}
									on:change={handleSettingChange}
								/>
								<span class="slider"></span>
							</label>
						</div>
					</div>
				</div>

				<!-- Privacy Settings -->
				<div class="settings-section">
					<h2>Privacy & Data</h2>
					<p class="section-description">Control your privacy and data sharing preferences</p>
					
					<div class="settings-group">
						<div class="setting-item">
							<div class="setting-info">
								<h3>Profile Visibility</h3>
								<p>Who can see your profile information</p>
							</div>
							<select 
								bind:value={settings.privacy.profileVisibility}
								on:change={handleSettingChange}
								class="setting-select"
							>
								<option value="public">Public</option>
								<option value="members">Members Only</option>
								<option value="private">Private</option>
							</select>
						</div>

						<div class="setting-item">
							<div class="setting-info">
								<h3>Show Watch History</h3>
								<p>Allow others to see what videos you've watched</p>
							</div>
							<label class="toggle">
								<input 
									type="checkbox" 
									bind:checked={settings.privacy.showWatchHistory}
									on:change={handleSettingChange}
								/>
								<span class="slider"></span>
							</label>
						</div>

						<div class="setting-item">
							<div class="setting-info">
								<h3>Personalized Recommendations</h3>
								<p>Use your viewing history to suggest relevant content</p>
							</div>
							<label class="toggle">
								<input 
									type="checkbox" 
									bind:checked={settings.privacy.allowRecommendations}
									on:change={handleSettingChange}
								/>
								<span class="slider"></span>
							</label>
						</div>
					</div>
				</div>

				<!-- Viewing Preferences -->
				<div class="settings-section">
					<h2>Viewing Preferences</h2>
					<p class="section-description">Customize your video viewing experience</p>
					
					<div class="settings-group">
						<div class="setting-item">
							<div class="setting-info">
								<h3>Autoplay Videos</h3>
								<p>Automatically play the next video in a series</p>
							</div>
							<label class="toggle">
								<input 
									type="checkbox" 
									bind:checked={settings.preferences.autoplay}
									on:change={handleSettingChange}
								/>
								<span class="slider"></span>
							</label>
						</div>

						<div class="setting-item">
							<div class="setting-info">
								<h3>Default Video Quality</h3>
								<p>Choose your preferred video quality</p>
							</div>
							<select 
								bind:value={settings.preferences.quality}
								on:change={handleSettingChange}
								class="setting-select"
							>
								<option value="auto">Auto (Recommended)</option>
								<option value="1080p">1080p HD</option>
								<option value="720p">720p HD</option>
								<option value="480p">480p</option>
								<option value="360p">360p</option>
							</select>
						</div>

						<div class="setting-item">
							<div class="setting-info">
								<h3>Language</h3>
								<p>Choose your preferred language</p>
							</div>
							<select 
								bind:value={settings.preferences.language}
								on:change={handleSettingChange}
								class="setting-select"
							>
								<option value="en">English</option>
								<option value="es">Spanish</option>
								<option value="fr">French</option>
								<option value="pt">Portuguese</option>
							</select>
						</div>

						<div class="setting-item">
							<div class="setting-info">
								<h3>Theme</h3>
								<p>Choose your preferred theme</p>
							</div>
							<select 
								bind:value={settings.preferences.theme}
								on:change={handleSettingChange}
								class="setting-select"
							>
								<option value="auto">Auto (System)</option>
								<option value="light">Light</option>
								<option value="dark">Dark</option>
							</select>
						</div>
					</div>
				</div>

				<!-- Account Actions -->
				<div class="settings-section danger-section">
					<h2>Account Actions</h2>
					<p class="section-description">Manage your account data and settings</p>
					
					<div class="settings-group">
						<div class="action-item">
							<div class="action-info">
								<h3>Export Account Data</h3>
								<p>Download a copy of your account data and activity</p>
							</div>
							<button class="btn btn-outline" disabled>
								Export Data
							</button>
						</div>

						<div class="action-item">
							<div class="action-info">
								<h3>Delete Account</h3>
								<p class="danger-text">Permanently delete your account and all associated data</p>
							</div>
							<button class="btn btn-danger" on:click={handleDeleteAccount}>
								Delete Account
							</button>
						</div>
					</div>
				</div>

				<!-- Save Actions -->
				<div class="save-actions">
					<button 
						type="button" 
						class="btn btn-outline" 
						on:click={handleReset}
						disabled={!hasChanges}
					>
						Reset Changes
					</button>
					<button 
						type="button" 
						class="btn btn-primary"
						on:click={handleSave}
						disabled={!hasChanges || saving}
					>
						{#if saving}
							<LoadingSpinner size="small" />
							Saving...
						{:else}
							Save Settings
						{/if}
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	.settings-page {
		min-height: 100vh;
		padding: 2rem 0;
		background: var(--bg-gradient);
	}

	.container {
		max-width: 800px;
		margin: 0 auto;
		padding: 0 1rem;
	}

	.page-header {
		text-align: center;
		margin-bottom: 3rem;
		position: relative;
	}

	.back-button {
		position: absolute;
		left: 0;
		top: 0;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		background: none;
		border: none;
		color: var(--text-secondary);
		cursor: pointer;
		font-size: 0.875rem;
		transition: color 0.3s ease;
	}

	.back-button:hover {
		color: var(--primary-color);
	}

	.back-button svg {
		width: 16px;
		height: 16px;
	}

	.page-header h1 {
		font-size: 2.5rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.page-header p {
		font-size: 1.1rem;
		color: var(--text-secondary);
	}

	.loading-container {
		text-align: center;
		padding: 3rem 0;
	}

	.settings-content {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.settings-section {
		background: var(--card-bg);
		border-radius: 20px;
		padding: 2rem;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
	}

	.settings-section h2 {
		font-size: 1.5rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.section-description {
		color: var(--text-secondary);
		margin-bottom: 2rem;
	}

	.settings-group {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.setting-item,
	.action-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem 0;
		border-bottom: 1px solid var(--border-color);
	}

	.setting-item:last-child,
	.action-item:last-child {
		border-bottom: none;
	}

	.setting-info,
	.action-info {
		flex: 1;
	}

	.setting-info h3,
	.action-info h3 {
		font-size: 1rem;
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 0.25rem 0;
	}

	.setting-info p,
	.action-info p {
		font-size: 0.875rem;
		color: var(--text-secondary);
		margin: 0;
	}

	.danger-text {
		color: var(--error-text) !important;
	}

	.setting-select {
		padding: 0.5rem 1rem;
		border: 1px solid var(--border-color);
		border-radius: 8px;
		background: var(--bg-secondary);
		color: var(--text-primary);
		font-size: 0.875rem;
		min-width: 120px;
	}

	.setting-select:focus {
		outline: none;
		border-color: var(--primary-color);
	}

	/* Toggle Switch Styles */
	.toggle {
		position: relative;
		display: inline-block;
		width: 50px;
		height: 24px;
	}

	.toggle input {
		opacity: 0;
		width: 0;
		height: 0;
	}

	.slider {
		position: absolute;
		cursor: pointer;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background-color: var(--bg-secondary);
		border: 1px solid var(--border-color);
		transition: 0.3s;
		border-radius: 24px;
	}

	.slider:before {
		position: absolute;
		content: "";
		height: 18px;
		width: 18px;
		left: 2px;
		bottom: 2px;
		background-color: white;
		transition: 0.3s;
		border-radius: 50%;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
	}

	input:checked + .slider {
		background-color: var(--primary-color);
		border-color: var(--primary-color);
	}

	input:checked + .slider:before {
		transform: translateX(26px);
	}

	.danger-section {
		border-color: var(--error-border);
	}

	.danger-section h2 {
		color: var(--error-text);
	}

	.btn-danger {
		background: var(--error-bg);
		color: var(--error-text);
		border: 1px solid var(--error-border);
	}

	.btn-danger:hover {
		background: var(--error-hover);
	}

	.save-actions {
		display: flex;
		justify-content: flex-end;
		gap: 1rem;
		padding: 2rem;
		background: var(--card-bg);
		border-radius: 20px;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
	}

	@media (max-width: 768px) {
		.back-button {
			position: static;
			justify-content: center;
			margin-bottom: 1rem;
		}

		.setting-item,
		.action-item {
			flex-direction: column;
			align-items: flex-start;
			gap: 1rem;
		}

		.save-actions {
			flex-direction: column;
		}
	}
</style> 