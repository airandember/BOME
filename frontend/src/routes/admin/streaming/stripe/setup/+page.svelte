<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';

	let secret = $state('');
	let saving = $state(false);
	let error = $state('');
	let success = $state('');
	let summary = $state<any>(null);
	let loading = $state(true);

	// Modal state for clear key confirmation
	let showClearModal = $state(false);
	let clearConfirmText = $state('');

	// Public settings state
	let publicSettings = $state<any>(null);
	let savingPublicSettings = $state(false);
	let publicSettingsError = $state('');
	let publicSettingsSuccess = $state('');
	let editingPublishableKey = $state(false);
	let editingPortalUrl = $state(false);
	let editingWebhookSecret = $state(false);
	let publishableKey = $state('');
	let portalUrl = $state('');
	let webhookSecret = $state('');

	// Stripe Data Management state
	let databaseStats = $state<any>(null)
	let dbStatsLoading = $state(false)
	let dbStatsLastUpdated = $state<Date | null>(null)
	let syncInProgress = $state(false)
	let syncStatus = $state('')
	let systemHealth = $state<any>(null)
	let systemStats = $state<any>(null)
	let syncJobs = $state<any[]>([])
	let systemLoading = $state(false)

	// Webhook configuration state
	let webhookEndpointUrl = $state('')
	let webhookStatus = $state<any>(null)
	let webhookLoading = $state(false)

	const { data = null, onClearKey } = $props<{ data?: any; onClearKey: () => void }>();

	onMount(async () => {
		if (data) {
			summary = data;
			loading = false;
		} else {
			//await fetchSummary();
			loading = false;
		}
		await loadPublicSettings();
		await initializeWebhookConfig();
	});

	//async function fetchSummary() {
		//try {
		//	loading = true;
		//	error = '';
		//	const res = await apiRequest('/admin/streaming/stripe/summary');
		//	if (res.ok) {
		//		const data = await res.json();
		//		summary = data.summary;
		//	} else {
		//		error = 'Failed to load summary';
		//	}
		//} catch (err) {
		//	error = 'Failed to load summary';
		//	console.error(err);
		//} finally {
		//	loading = false;
		//}
	//}

	async function saveSecret() {
		if (!secret.trim()) return;
		
		saving = true;
		error = '';
		success = '';
		
		try {
			const res = await apiRequest('/admin/streaming/stripe/secret', {
				method: 'POST',
				body: JSON.stringify({ key: secret })
			});
			
			if (res.ok) {
				success = 'Stripe key saved successfully!';
				secret = '';
				//await fetchSummary(); // Refresh the summary
			} else {
				const errorData = await res.json();
				error = errorData.error || 'Failed to save key';
			}
		} catch (err) {
			error = 'Failed to save key';
			console.error(err);
		} finally {
			saving = false;
		}
	}



	// === PUBLIC SETTINGS FUNCTIONS ===

	// Load public settings
	async function loadPublicSettings() {
		try {
			const res = await apiRequest('/admin/streaming/stripe/public-settings');
			if (res.ok) {
				const data = await res.json();
				publicSettings = data.settings || {};
			}
		} catch (err) {
			console.error('Failed to load public settings:', err);
		}
	}

	// === WEBHOOK CONFIGURATION FUNCTIONS ===

	// Initialize webhook configuration
	async function initializeWebhookConfig() {
		// Generate webhook endpoint URL based on current environment
		let apiBaseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';
		
		// 🔒 PRODUCTION FIX: Ensure we use HTTPS in production
		if (typeof window !== 'undefined') {
			const isProduction = window.location.hostname !== 'localhost' && !window.location.hostname.includes('127.0.0.1');
			if (isProduction) {
				// Force HTTPS for production
				apiBaseUrl = apiBaseUrl.replace('http://', 'https://');
				// Ensure we have the full production URL
				if (!apiBaseUrl.includes('bookofmormonevidence.org')) {
					apiBaseUrl = `https://watch.bookofmormonevidence.org/bome-backend/api/v1`;
				}
			}
		}
		
		// Use the PUBLIC webhook endpoint that Stripe can access (no auth required)
		webhookEndpointUrl = `${apiBaseUrl}/webhooks/stripe`;
		
		// Load webhook status
		await loadWebhookStatus();
	}

	// Load webhook status and recent events
	async function loadWebhookStatus() {
		webhookLoading = true;
		try {
			const res = await apiRequest('/admin/streaming/stripe/webhooks/status');
			if (res.ok) {
				const data = await res.json();
				// The backend returns webhook data nested under "webhook" key
				webhookStatus = data.webhook || data;
			}
		} catch (err) {
			console.error('Failed to load webhook status:', err);
			webhookStatus = { error: 'Failed to load webhook status' };
		} finally {
			webhookLoading = false;
		}
	}

	// Copy webhook URL to clipboard
	async function copyWebhookUrl() {
		try {
			await navigator.clipboard.writeText(webhookEndpointUrl);
			success = 'Webhook URL copied to clipboard!';
			setTimeout(() => success = '', 3000);
		} catch (err) {
			error = 'Failed to copy to clipboard';
			setTimeout(() => error = '', 3000);
		}
	}

	// Ping webhook endpoint to test connectivity
	async function pingWebhook() {
		webhookLoading = true;
		try {
			const res = await apiRequest('/admin/streaming/stripe/webhooks/ping', {
				method: 'POST'
			});
			
			if (res.ok) {
				const result = await res.json();
				success = result.message || 'Webhook ping successful! ✅';
				// Reload webhook status to show the updated activity
				await loadWebhookStatus();
			} else {
				error = 'Webhook ping failed';
			}
		} catch (err) {
			console.error('Failed to ping webhook:', err);
			error = 'Failed to ping webhook';
		} finally {
			webhookLoading = false;
			setTimeout(() => {
				success = '';
				error = '';
			}, 3000);
		}
	}

	// Save publishable key
	async function savePublishableKey() {
		if (!publishableKey.trim()) return;
		
		savingPublicSettings = true;
		publicSettingsError = '';
		publicSettingsSuccess = '';
		
		try {
			const res = await apiRequest('/admin/streaming/stripe/public-settings', {
				method: 'POST',
				body: JSON.stringify({ 
					key: 'stripe_publishable_key', 
					value: publishableKey.trim() 
				})
			});
			
			if (res.ok) {
				publicSettingsSuccess = 'Publishable key saved successfully!';
				publishableKey = '';
				editingPublishableKey = false;
				await loadPublicSettings(); // Refresh settings
			} else {
				const errorData = await res.json();
				publicSettingsError = errorData.error || 'Failed to save publishable key';
			}
		} catch (err) {
			publicSettingsError = 'Failed to save publishable key';
			console.error(err);
		} finally {
			savingPublicSettings = false;
		}
	}

	// Save portal URL
	async function savePortalUrl() {
		if (!portalUrl.trim()) return;
		
		savingPublicSettings = true;
		publicSettingsError = '';
		publicSettingsSuccess = '';
		
		try {
			const res = await apiRequest('/admin/streaming/stripe/public-settings', {
				method: 'POST',
				body: JSON.stringify({ 
					key: 'stripe_portal_url', 
					value: portalUrl.trim() 
				})
			});
			
			if (res.ok) {
				publicSettingsSuccess = 'Portal URL saved successfully!';
				portalUrl = '';
				editingPortalUrl = false;
				await loadPublicSettings(); // Refresh settings
			} else {
				const errorData = await res.json();
				publicSettingsError = errorData.error || 'Failed to save portal URL';
			}
		} catch (err) {
			publicSettingsError = 'Failed to save portal URL';
			console.error(err);
		} finally {
			savingPublicSettings = false;
		}
	}

	// Save webhook secret
	async function saveWebhookSecret() {
		if (!webhookSecret.trim()) return;
		
		savingPublicSettings = true;
		publicSettingsError = '';
		publicSettingsSuccess = '';
		
		try {
			const res = await apiRequest('/admin/streaming/stripe/public-settings', {
				method: 'POST',
				body: JSON.stringify({ 
					key: 'stripe_webhook_secret', 
					value: webhookSecret.trim() 
				})
			});
			
			if (res.ok) {
				publicSettingsSuccess = 'Webhook secret saved successfully!';
				webhookSecret = '';
				editingWebhookSecret = false;
				await loadPublicSettings(); // Refresh settings
			} else {
				const errorData = await res.json();
				publicSettingsError = errorData.error || 'Failed to save webhook secret';
			}
		} catch (err) {
			publicSettingsError = 'Failed to save webhook secret';
			console.error(err);
		} finally {
			savingPublicSettings = false;
		}
	}

	// Delete public setting
	async function deletePublicSetting(key: string) {
		if (!confirm(`Are you sure you want to delete the ${key.replace('stripe_', '')} setting?`)) {
			return;
		}
		
		savingPublicSettings = true;
		publicSettingsError = '';
		publicSettingsSuccess = '';
		
		try {
			const res = await apiRequest(`/admin/streaming/stripe/public-settings/${key}`, {
				method: 'DELETE'
			});
			
			if (res.ok) {
				publicSettingsSuccess = 'Setting deleted successfully!';
				await loadPublicSettings(); // Refresh settings
			} else {
				const errorData = await res.json();
				publicSettingsError = errorData.error || 'Failed to delete setting';
			}
		} catch (err) {
			publicSettingsError = 'Failed to delete setting';
			console.error(err);
		} finally {
			savingPublicSettings = false;
		}
	}

	// === STRIPE DATA MANAGEMENT FUNCTIONS ===

	// Load database stats (real counts from database)
	async function loadDatabaseStats() {
		if (dbStatsLoading) return
		
		dbStatsLoading = true
		console.log("📊 Loading database stats...")
		
		try {
			const response = await apiRequest('/admin/streaming/stripe/database/stats')
			const data = await response.json()
			
			databaseStats = data
			dbStatsLastUpdated = new Date()
			
			console.log("✅ Database stats loaded:", data)
		} catch (error) {
			console.error("❌ Failed to load database stats:", error)
		} finally {
			dbStatsLoading = false
		}
	}

	// Trigger manual sync with enhanced debounce protection
	let lastSyncTime = 0
	let activeRequestIds = new Set()
	
	async function triggerManualSync(syncType = 'customers') {
		console.log(`🔍 [MANUAL-SYNC] ${syncType} sync requested`)
		
		// Generate unique request ID for this specific request
		const requestId = `${syncType}_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
		
		// Prevent double execution within 3 seconds OR if sync is in progress
		const now = Date.now()
		if (syncInProgress || (now - lastSyncTime) < 3000 || activeRequestIds.has(syncType)) {
			console.log(`🚫 Sync blocked: inProgress=${syncInProgress}, timeSince=${now - lastSyncTime}ms, activeRequests=${Array.from(activeRequestIds)}`)
			return
		}
		
		// Mark this request as active
		activeRequestIds.add(syncType)
		lastSyncTime = now
		syncInProgress = true
		syncStatus = `Starting ${syncType} sync...`
		
		try {
			const uniqueRequestId = `frontend_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
			
			console.log(`🔍 [MANUAL-SYNC] Manual sync: ${syncType} (ID: ${uniqueRequestId})`)
			
			const response = await apiRequest(`/admin/streaming/stripe/sync/trigger?type=${syncType}`, {
				method: 'POST',
				headers: {
					'X-Frontend-Request-ID': uniqueRequestId,
					'X-Frontend-Timestamp': Date.now().toString(),
					'X-Frontend-User-Agent': navigator.userAgent
				}
			})
			const data = await response.json()
			console.log(`🔍 [MANUAL-SYNC] Manual sync API response:`, data)
			
			if (response.ok) {
			if (syncType === 'cleanup_orphaned') {
				syncStatus = `🧹 Invalid product subscriptions marked successfully!`
				} else {
					syncStatus = `✅ ${syncType} sync completed successfully!`
				}
				console.log("✅ Manual sync completed:", data)
				
				// Reload database stats to show updated counts
				setTimeout(() => {
					loadDatabaseStats()
				}, 1000)
			} else {
				// Handle rate limiting gracefully - don't show technical details to user
				if (response.status === 429 || data.status === 'rate_limited') {
					syncStatus = `⏳ ${syncType} sync is cooling down. Please wait a few minutes before trying again.`
					console.log("⏳ Sync rate limited:", data)
				} else {
					syncStatus = `❌ ${syncType} sync failed: ${data.error || data.message || 'Unknown error'}`
					console.error("❌ Manual sync failed:", data)
				}
			}
		} catch (error: any) {
			syncStatus = `❌ ${syncType} sync failed: ${error.message || 'Unknown error'}`
			console.error("❌ Manual sync error:", error)
		} finally {
			syncInProgress = false
			// Remove this sync type from active requests
			activeRequestIds.delete(syncType)
			
			// Clear status after different timeouts based on message type
			const clearTimeout = syncStatus.includes('cooling down') ? 10000 : 5000 // 10s for rate limit, 5s for others
			setTimeout(() => {
				syncStatus = ''
			}, clearTimeout)
		}
	}

	async function loadSystemHealth() {
		if (systemLoading) return
		
		systemLoading = true
		try {
			const response = await apiRequest('/admin/streaming/stripe/system/health')
			const data = await response.json()
			systemHealth = data
			console.log("🏥 System health loaded:", data)
		} catch (error) {
			console.error("❌ Failed to load system health:", error)
		} finally {
			systemLoading = false
		}
	}

	async function loadSystemStats() {
		try {
			const response = await apiRequest('/admin/streaming/stripe/system/stats')
			const data = await response.json()
			systemStats = data
			console.log("📊 System stats loaded:", data)
		} catch (error) {
			console.error("❌ Failed to load system stats:", error)
		}
	}

	async function loadSyncJobs() {
		try {
			const response = await apiRequest('/admin/streaming/stripe/system/jobs?limit=10')
			const data = await response.json()
			syncJobs = data.jobs || []
			console.log("📋 Sync jobs loaded:", data)
		} catch (error) {
			console.error("❌ Failed to load sync jobs:", error)
		}
	}

	async function loadSystemData() {
		await Promise.all([
			loadSystemHealth(),
			loadSystemStats(),
			loadSyncJobs()
		])
	}
</script>

{#if loading}
	<div class="loading">
		<div class="spinner"></div>
		<p>Loading setup...</p>
	</div>
{:else}
	<div class="setup-container">
		<div class="setup-header">
			<h1>Setup Stripe</h1>
			<p>Configure your Stripe integration to start processing payments</p>
		</div>

		<!--{#if !summary?.enabled}
			<div class="setup-section">
				<div class="setup-card">
					<div class="card-header">
						<h2>🔑 Connect Your Stripe Account</h2>
						<p>Enter your Stripe secret key to enable payment processing</p>
					</div>
					
					<form on:submit|preventDefault={saveSecret} class="setup-form">
						<div class="input-group">
							<label for="stripe-key" class="input-label">
								Stripe Secret Key
								<span class="required">*</span>
							</label>
							<input 
								id="stripe-key"
								class="input" 
								type="password" 
								placeholder="sk_test_... or sk_live_..." 
								bind:value={secret}
								required
							/>
							<div class="input-help">
								Your secret key will be encrypted and stored securely. It will never be returned or displayed.
							</div>
						</div>

						<button 
							type="submit" 
							class="btn btn-primary btn-lg" 
							disabled={saving || !secret.trim()}
						>
							{saving ? 'Connecting...' : 'Connect Stripe Account'}
						</button>
					</form>
					
					{#if error}
						<div class="alert alert-error">
							<div class="alert-icon">❌</div>
							<div class="alert-content">
								<strong>Connection Failed</strong>
								<p>{error}</p>
							</div>
						</div>
					{/if}
					
					{#if success}
						<div class="alert alert-success">
							<div class="alert-icon">✅</div>
							<div class="alert-content">
								<strong>Success!</strong>
								<p>{success}</p>
							</div>
						</div>
					{/if}
				</div>

				// Setup Instructions //
				<div class="instructions-card">
					<h3>🚀 Getting Started</h3>
					<div class="steps">
						<div class="step">
							<div class="step-number">1</div>
							<div class="step-content">
								<h4>Create a Stripe Account</h4>
								<p>Visit <a href="https://stripe.com" target="_blank" rel="noopener">stripe.com</a> to create your account if you haven't already.</p>
							</div>
						</div>
						
						<div class="step">
							<div class="step-number">2</div>
							<div class="step-content">
								<h4>Get Your API Keys</h4>
								<p>In your Stripe Dashboard, go to <strong>Developers → API Keys</strong> to find your secret key.</p>
							</div>
						</div>
						
						<div class="step">
							<div class="step-number">3</div>
							<div class="step-content">
								<h4>Test vs Live Mode</h4>
								<p>Start with test keys (sk_test_...) for development, then switch to live keys (sk_live_...) for production.</p>
							</div>
						</div>
						
						<div class="step">
							<div class="step-number">4</div>
							<div class="step-content">
								<h4>Enter Your Key</h4>
								<p>Paste your secret key above and click "Connect Stripe Account" to get started.</p>
							</div>
						</div>
					</div>
				</div>

				// Security Notice //
				<div class="security-notice">
					<div class="notice-icon">🔒</div>
					<div class="notice-content">
						<h4>Security & Privacy</h4>
						<ul>
							<li>Your secret key is encrypted using AES-GCM encryption before storage</li>
							<li>Keys are never returned to the frontend or displayed anywhere</li>
							<li>Only you can replace the key by entering a new one</li>
							<li>All communication uses HTTPS encryption</li>
						</ul>
					</div>
				</div>
			</div>
		{:else}-->
			<!-- Already Connected -->
			<div class="connected-section">
				<div class="connected-card">
					<div class="connected-header">
						<div class="connected-icon">✅</div>
						<div class="connected-info">
							<h2>Stripe Connected Successfully</h2>
							<p>Your Stripe account is connected and ready to process payments</p>
							<!--<div class="environment-badge {summary.environment === 'live' ? 'live' : 'test'}">
								{summary.environment === 'live' ? '🔴 LIVE MODE' : '🟡 TEST MODE'}
							</div>-->
						</div>
					</div>

					<!--<div class="connected-stats">
						<div class="stat">
							<span class="stat-value">{summary.products_count || 0}</span>
							<span class="stat-label">Products</span>
						</div>
						<div class="stat">
							<span class="stat-value">{summary.customers_count || 0}</span>
							<span class="stat-label">Customers</span>
						</div>
						<div class="stat">
							<span class="stat-value">{summary.subscriptions_count || 0}</span>
							<span class="stat-label">Subscriptions</span>
						</div>
					</div>-->

					<div class="connected-actions">
						<button class="btn btn-outline" onclick={() => { summary = { enabled: false }; }}>
							🔑 Update Key
						</button>
						<!--<button class="btn btn-secondary" onclick={fetchSummary}>
							🔄 Refresh Connection
						</button>
						 Use the parent's function -->
						<button class="btn btn-danger" onclick={onClearKey}>
							🗑️ Clear Key
						</button>
					</div>
				</div>

				<!-- Public Settings Section -->
				<div class="next-steps-card">
					<h3>🔑 Stripe Configuration</h3>
					<p>Configure your Stripe settings (publishable key, portal URL, webhook secret) for complete integration</p>
					
					<div class="public-settings-grid">
						<!-- Stripe Publishable Key -->
						<div class="setting-item">
							<div class="setting-header">
								<h4>Stripe Publishable Key</h4>
								<span class="setting-type">Public</span>
							</div>
							<p class="setting-description">
								Your Stripe publishable key (pk_test_... or pk_live_...) for frontend checkout forms
							</p>
							
							{#if publicSettings?.stripe_publishable_key && !editingPublishableKey}
								<div class="saved-setting">
									<div class="saved-setting-display">
										<code class="key-display">
											{publicSettings.stripe_publishable_key.substring(0, 12)}...{publicSettings.stripe_publishable_key.slice(-4)}
										</code>
									</div>
									<div class="saved-setting-actions">
										<button class="btn btn-outline btn-sm" onclick={() => editingPublishableKey = true}>
											✏️ Update
										</button>
										<button class="btn btn-secondary btn-sm" onclick={() => deletePublicSetting('stripe_publishable_key')}>
											🗑️ Remove
										</button>
									</div>
								</div>
							{:else}
								<form onsubmit={(e) => { e.preventDefault(); savePublishableKey(); }} class="setting-form">
									<div class="input-group">
										<input 
											class="input" 
											type="text" 
											placeholder="pk_test_..." 
											bind:value={publishableKey}
										/>
									</div>
									<div class="setting-form-actions">
										<button 
											type="submit" 
											class="btn btn-primary btn-sm" 
											disabled={savingPublicSettings || !publishableKey.trim()}
										>
											{savingPublicSettings ? 'Saving...' : 'Save Key'}
										</button>
										{#if editingPublishableKey}
											<button 
												type="button" 
												class="btn btn-secondary btn-sm" 
												onclick={() => { editingPublishableKey = false; publishableKey = ''; }}
											>
												Cancel
											</button>
										{/if}
									</div>
								</form>
							{/if}
						</div>

						<!-- Customer Portal URL -->
						<div class="setting-item">
							<div class="setting-header">
								<h4>Customer Portal URL</h4>
								<span class="setting-type">Public</span>
							</div>
							<p class="setting-description">
								Your Stripe customer portal URL for subscription management
							</p>
							
							{#if publicSettings?.stripe_portal_url && !editingPortalUrl}
								<div class="saved-setting">
									<div class="saved-setting-display">
										<a href={publicSettings.stripe_portal_url} target="_blank" rel="noopener" class="portal-link">
											{publicSettings.stripe_portal_url}
										</a>
									</div>
									<div class="saved-setting-actions">
										<button class="btn btn-outline btn-sm" onclick={() => editingPortalUrl = true}>
											✏️ Update
										</button>
										<button class="btn btn-secondary btn-sm" onclick={() => deletePublicSetting('stripe_portal_url')}>
											🗑️ Remove
										</button>
									</div>
								</div>
							{:else}
								<form onsubmit={(e) => { e.preventDefault(); savePortalUrl(); }} class="setting-form">
									<div class="input-group">
										<input 
											class="input" 
											type="url" 
											placeholder="https://billing.stripe.com/p/login/..." 
											bind:value={portalUrl}
										/>
									</div>
									<div class="setting-form-actions">
										<button 
											type="submit" 
											class="btn btn-primary btn-sm" 
											disabled={savingPublicSettings || !portalUrl.trim()}
										>
											{savingPublicSettings ? 'Saving...' : 'Save URL'}
										</button>
										{#if editingPortalUrl}
											<button 
												type="button" 
												class="btn btn-secondary btn-sm" 
												onclick={() => { editingPortalUrl = false; portalUrl = ''; }}
											>
												Cancel
											</button>
										{/if}
									</div>
								</form>
							{/if}
						</div>

						<!-- Webhook Secret -->
						<div class="setting-item">
							<div class="setting-header">
								<h4>Webhook Secret</h4>
								<span class="setting-type secure">Secure</span>
							</div>
							<p class="setting-description">
								Your Stripe webhook endpoint secret (whsec_...) for validating webhook events. Get this from your Stripe Dashboard → Webhooks → [Your Endpoint] → Signing secret.
							</p>
							
							{#if publicSettings?.stripe_webhook_secret && !editingWebhookSecret}
								<div class="saved-setting">
									<div class="saved-setting-display">
										<code class="key-display">
											whsec_••••••••••••••••••••••••••••••••••••••••••••••••••••
										</code>
									</div>
									<div class="saved-setting-actions">
										<button class="btn btn-outline btn-sm" onclick={() => editingWebhookSecret = true}>
											✏️ Update
										</button>
										<button class="btn btn-secondary btn-sm" onclick={() => deletePublicSetting('stripe_webhook_secret')}>
											🗑️ Remove
										</button>
									</div>
								</div>
							{:else}
								<form onsubmit={(e) => { e.preventDefault(); saveWebhookSecret(); }} class="setting-form">
									<div class="input-group">
										<input 
											class="input" 
											type="password" 
											placeholder="whsec_..." 
											bind:value={webhookSecret}
										/>
									</div>
									<div class="setting-form-actions">
										<button 
											type="submit" 
											class="btn btn-primary btn-sm" 
											disabled={savingPublicSettings || !webhookSecret.trim()}
										>
											{savingPublicSettings ? 'Saving...' : 'Save Secret'}
										</button>
										{#if editingWebhookSecret}
											<button 
												type="button" 
												class="btn btn-secondary btn-sm" 
												onclick={() => { editingWebhookSecret = false; webhookSecret = ''; }}
											>
												Cancel
											</button>
										{/if}
									</div>
								</form>
							{/if}
						</div>
					</div>

					{#if publicSettingsError}
						<div class="alert alert-error">
							<div class="alert-icon">❌</div>
							<div class="alert-content">
								<strong>Error</strong>
								<p>{publicSettingsError}</p>
							</div>
						</div>
					{/if}
					
					{#if publicSettingsSuccess}
						<div class="alert alert-success">
							<div class="alert-icon">✅</div>
							<div class="alert-content">
								<strong>Success!</strong>
								<p>{publicSettingsSuccess}</p>
							</div>
						</div>
					{/if}
				</div>

				<!-- Webhook Configuration Section -->
				<div class="next-steps-card web-hook-whole">
					<h3>🔗 Webhook Configuration</h3>
					<p>Configure Stripe webhooks for real-time payment processing and video access</p>
					
					<div class="webhook-config">
						<!-- Webhook Endpoint URL -->
						<div class="setting-item">
							<div class="setting-header">
								<h4>Webhook Endpoint URL</h4>
								<span class="setting-type">Required</span>
							</div>
							<p class="setting-description">
								Copy this URL to your Stripe Dashboard → Webhooks to enable real-time payment processing.
								When payments succeed, users will automatically get video access. This is a public endpoint that Stripe can access directly.
							</p>
							
							<div class="webhook-url-display">
								<div class="url-container">
									<code class="webhook-url">{webhookEndpointUrl}</code>
									<div class="url-actions">
										<button 
											class="btn btn-outline btn-sm copy-btn" 
											onclick={copyWebhookUrl}
											title="Copy webhook URL"
										>
											📋 Copy
										</button>
										
									</div>
								</div>
							</div>
						</div>

						<!-- Webhook Status -->
						{#if webhookLoading}
							<div class="setting-item">
								<div class="loading-small">
									<div class="spinner-small"></div>
									<span>Loading webhook status...</span>
								</div>
							</div>
						{:else if webhookStatus}
							<div class="setting-item">
								<div class="setting-header">
									<h4>Webhook Status</h4>
									<span class="setting-type {webhookStatus.active ? 'success' : 'warning'}">
										{webhookStatus.active ? 'Active' : 'Inactive'}
									</span>
								</div>
								<p class="setting-description">
									{#if webhookStatus.active}
										✅ Webhooks are working! Last event received: {webhookStatus.lastEvent || 'Never'}
									{:else}
										⚠️ No recent webhook events detected. Make sure to configure webhooks in your Stripe Dashboard.
									{/if}
								</p>
								
								{#if webhookStatus.eventsToday}
									<div class="webhook-stats">
										<div class="stat-item">
											<span class="stat-value">{webhookStatus.eventsToday}</span>
											<span class="stat-label">Events Today</span>
										</div>
										{#if webhookStatus.successRate}
											<div class="stat-item">
												<span class="stat-value">{webhookStatus.successRate}%</span>
												<span class="stat-label">Success Rate</span>
											</div>
										{/if}
									</div>
								{/if}
							</div>
						{/if}

						<!-- Setup Instructions -->
						<div class="setting-item">
							<div class="setting-header">
								<h4>Setup Instructions</h4>
								<span class="setting-type">Guide</span>
							</div>
							<div class="webhook-instructions">
								<div class="instruction-step">
									<div class="step-number">1</div>
									<div class="step-content">
										<strong>Go to Stripe Dashboard</strong>
										<p>Navigate to <a href="https://dashboard.stripe.com/webhooks" target="_blank" rel="noopener">Developers → Webhooks</a></p>
									</div>
								</div>
								<div class="instruction-step">
									<div class="step-number">2</div>
									<div class="step-content">
										<strong>Add Endpoint</strong>
										<p>Click "Add endpoint" and paste the URL above</p>
									</div>
								</div>
								<div class="instruction-step">
									<div class="step-number">3</div>
									<div class="step-content">
										<strong>Select Events</strong>
										<p>Choose these events: <code>customer.*</code>, <code>customer.subscription.*</code>, <code>invoice.payment_succeeded</code>, <code>invoice.payment_failed</code>, <code>product.*</code>, <code>price.*</code></p>
									</div>
								</div>
								<div class="instruction-step">
									<div class="step-number">4</div>
									<div class="step-content">
										<strong>Test Webhook</strong>
										<p>Use Stripe's "Send test webhook" feature to verify the connection</p>
									</div>
								</div>
							</div>
						</div>
					</div>
				</div>

				<!-- Stripe Data Management Section -->
				<div class="data-management-card">
					<h3>🔧 Stripe Data Management</h3>
					<p>Manage and synchronize your Stripe data with the local database</p>
					
					<!-- Sync Status Display -->
					{#if syncStatus}
						<div class="sync-status {syncStatus.includes('✅') ? 'success' : syncStatus.includes('⏳') ? 'warning' : 'error'}">
							<div class="sync-status-message">{syncStatus}</div>
						</div>
					{/if}

					<!-- Database Stats Display -->
					{#if databaseStats}
						<div class="database-stats">
							<h4>📊 Database Statistics</h4>
							<div class="stats-grid">
								<div class="stat-card">
									<div class="stat-value">{databaseStats.customers || 0}</div>
									<div class="stat-label">Customers</div>
								</div>
								<div class="stat-card">
									<div class="stat-value">{databaseStats.subscriptions || 0}</div>
									<div class="stat-label">Active Subscriptions</div>
								</div>
								<div class="stat-card">
									<div class="stat-value">{databaseStats.products || 0}</div>
									<div class="stat-label">Products</div>
								</div>
								<div class="stat-card">
									<div class="stat-value">{databaseStats.coupons || 0}</div>
									<div class="stat-label">Coupons</div>
								</div>
							</div>
							{#if dbStatsLastUpdated}
								<div class="last-updated">
									Last updated: {dbStatsLastUpdated.toLocaleString()}
								</div>
							{/if}
						</div>
					{/if}

					<!-- Management Actions -->
					<div class="management-actions">
						<div class="action-group">
							<h4>📊 Database Operations</h4>
							<div class="button-group">
								<button 
									class="btn btn-secondary" 
									onclick={(e) => { e.preventDefault(); e.stopPropagation(); loadDatabaseStats(); }}
									disabled={dbStatsLoading}
								>
									{dbStatsLoading ? '🔄 Loading...' : '📊 Refresh DB Stats'}
								</button>
							</div>
						</div>

						<div class="action-group">
							<h4>🔄 Data Synchronization</h4>
							<div class="button-group">
								<!--<button 
									class="btn btn-success" 
									onclick={(e) => { e.preventDefault(); e.stopPropagation(); triggerManualSync('customers'); }}
									disabled={syncInProgress}
							>
									{syncInProgress ? '🔄 Syncing...' : '🚀 Sync Customers'}
								</button>-->
								<button 
									class="btn btn-warning" 
									onclick={(e) => { e.preventDefault(); e.stopPropagation(); triggerManualSync('initial'); }}
									disabled={syncInProgress}
								>
									{syncInProgress ? '🔄 Syncing...' : '🔄 Full Sync'}
								</button>
								<!--<button 
									class="btn btn-info" 
									onclick={(e) => { e.preventDefault(); e.stopPropagation(); triggerManualSync('coupons'); }}
									disabled={syncInProgress}
								>
									{syncInProgress ? '🔄 Syncing...' : '🎟️ Sync Coupons'}
								</button>-->
								<button 
									class="btn btn-secondary" 
									onclick={(e) => { e.preventDefault(); e.stopPropagation(); triggerManualSync('monthly_metrics'); }}
									disabled={syncInProgress}
								>
									{syncInProgress ? '🔄 Syncing...' : '📊 Sync Metrics'}
								</button>
								<button 
									class="btn btn-info" 
									onclick={(e) => { e.preventDefault(); e.stopPropagation(); triggerManualSync('products'); }}
									disabled={syncInProgress}
								>
									{syncInProgress ? '🔄 Syncing...' : '📦 Sync Products'}
								</button>
								<button 
									class="btn btn-success" 
									onclick={(e) => { e.preventDefault(); e.stopPropagation(); triggerManualSync('prices'); }}
									disabled={syncInProgress}
								>
									{syncInProgress ? '🔄 Syncing...' : '💰 Sync Prices'}
								</button>
							<button 
								class="btn btn-primary" 
								onclick={(e) => { e.preventDefault(); e.stopPropagation(); triggerManualSync('subscriptions'); }}
								disabled={syncInProgress}
							>
								{syncInProgress ? '🔄 Syncing...' : '💳 Sync Subscriptions'}
							</button>
							<button 
								class="btn btn-warning" 
								onclick={(e) => { e.preventDefault(); e.stopPropagation(); triggerManualSync('cleanup_orphaned'); }}
								disabled={syncInProgress}
								title="Mark subscriptions with invalid product references as 'invalid_product'"
							>
								{syncInProgress ? '🔄 Cleaning...' : '🧹 Ensure Active Plans'}
							</button>
							</div>
						</div>

						<div class="action-group">
							<h4>🏥 System Management</h4>
							<div class="button-group">
								<button 
									class="btn btn-info" 
									onclick={(e) => { e.preventDefault(); e.stopPropagation(); loadSystemData(); }}
									disabled={systemLoading}
								>
									{systemLoading ? '🔄 Loading...' : '🏥 System Health'}
								</button>
							</div>
						</div>
					</div>

					<!-- System Health Panel -->
					{#if systemHealth || systemStats || syncJobs.length > 0}
						<div class="system-panel">
							<h4>🏥 System Status</h4>

							{#if systemHealth}
								<div class="system-health">
									<div class="health-grid">
										<div class="health-card status-{systemHealth.status}">
											<div class="health-icon">
												{systemHealth.status === 'healthy' ? '✅' : systemHealth.status === 'degraded' ? '⚠️' : '❌'}
											</div>
											<div class="health-info">
												<div class="health-label">System Status</div>
												<div class="health-value">{systemHealth.status}</div>
											</div>
										</div>
										<div class="health-card">
											<div class="health-icon">🗄️</div>
											<div class="health-info">
												<div class="health-label">Database</div>
												<div class="health-value">{systemHealth.database_status}</div>
											</div>
										</div>
									</div>
								</div>
							{/if}

							{#if syncJobs.length > 0}
								<div class="sync-jobs">
									<h5>Recent Sync Jobs</h5>
									<div class="jobs-list">
										{#each syncJobs.slice(0, 3) as job}
											<div class="job-card status-{job.status}">
												<div class="job-info">
													<div class="job-type">{job.job_type} - {job.entity_type}</div>
													<div class="job-status">{job.status}</div>
												</div>
												<div class="job-progress">
													{job.processed_items || 0} / {job.total_items || 0}
												</div>
											</div>
										{/each}
									</div>
								</div>
							{/if}
						</div>
					{/if}
				</div>
			</div>
		<!--{/if}-->
	</div>
{/if}

<style>
	.setup-container {
		padding: var(--space-lg);
		max-width: 1800px;
		margin: 0 auto;
	}

	.setup-header {
		text-align: center;
		margin-bottom: var(--space-xl);
	}

	.setup-header h1 {
		margin: 0 0 var(--space-sm) 0;
		color: var(--text);
		font-size: 2.5rem;
	}

	.setup-header p {
		margin: 0;
		color: var(--text-muted);
		font-size: 1.2rem;
	}

	.setup-section {
		display: flex;
		flex-direction: column;
		gap: var(--space-xl);
	}

	.setup-card {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		padding: var(--space-xl);
	}

	.card-header {
		text-align: center;
		margin-bottom: var(--space-xl);
	}

	.card-header h2 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
		font-size: 1.8rem;
	}

	.card-header p {
		margin: 0;
		color: var(--text-muted);
		font-size: 1.1rem;
	}

	.setup-form {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.input-group {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}

	.input-label {
		font-weight: 600;
		color: var(--text);
		font-size: 1rem;
	}

	.required {
		color: var(--error);
	}

	.input {
		padding: var(--space-md);
		border: 2px solid var(--border);
		border-radius: var(--radius-md);
		font-size: 1rem;
		background: var(--surface);
		color: var(--text);
		transition: border-color 0.2s ease;
	}

	.input:focus {
		outline: none;
		border-color: var(--primary);
	}

	.input-help {
		font-size: 0.9rem;
		color: var(--text-muted);
		line-height: 1.4;
	}

	.btn {
		padding: var(--space-md) var(--space-lg);
		border: none;
		border-radius: var(--radius-md);
		font-size: 1rem;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		display: inline-block;
		text-align: center;
		font-weight: 600;
	}

	.btn-lg {
		padding: var(--space-lg) var(--space-xl);
		font-size: 1.1rem;
	}

	.btn-sm {
		padding: var(--space-xs) var(--space-sm);
		font-size: 0.9rem;
	}

	.btn-primary {
		background: var(--primary);
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: var(--primary-dark);
		transform: translateY(-2px);
	}

	.btn-primary:disabled {
		opacity: 0.6;
		cursor: not-allowed;
		transform: none;
	}

	.btn-secondary {
		background: var(--surface);
		color: var(--text);
		border: 1px solid var(--border);
	}

	.btn-secondary:hover {
		background: var(--surface-hover);
	}

	.btn-outline {
		background: transparent;
		color: var(--text);
		border: 1px solid var(--border);
	}

	.btn-outline:hover {
		background: var(--surface-hover);
	}

	.btn-danger {
		background: #dc2626;
		color: white;
	}

	.btn-danger:hover:not(:disabled) {
		background: #b91c1c;
		transform: translateY(-2px);
	}

	.btn-danger:disabled {
		opacity: 0.6;
		cursor: not-allowed;
		transform: none;
	}

	.btn-warning {
		background: #f59e0b;
		color: white;
	}

	.btn-warning:hover:not(:disabled) {
		background: #d97706;
		transform: translateY(-2px);
	}

	.btn-warning:disabled {
		opacity: 0.6;
		cursor: not-allowed;
		transform: none;
	}

	.btn-info {
		background: #0ea5e9;
		color: white;
	}

	.btn-info:hover:not(:disabled) {
		background: #0284c7;
		transform: translateY(-2px);
	}

	.btn-info:disabled {
		opacity: 0.6;
		cursor: not-allowed;
		transform: none;
	}

	.btn-success {
		background: #10b981;
		color: white;
	}

	.btn-success:hover:not(:disabled) {
		background: #059669;
		transform: translateY(-2px);
	}

	.btn-success:disabled {
		opacity: 0.6;
		cursor: not-allowed;
		transform: none;
	}

	.alert {
		display: flex;
		align-items: flex-start;
		gap: var(--space-md);
		padding: var(--space-lg);
		border-radius: var(--radius-md);
		margin-top: var(--space-lg);
	}

	.alert-error {
		background: var(--error-light);
		border: 1px solid var(--error);
	}

	.alert-success {
		background: var(--success-light);
		border: 1px solid var(--success);
	}

	.alert-icon {
		font-size: 1.2rem;
		flex-shrink: 0;
	}

	.alert-content strong {
		display: block;
		margin-bottom: var(--space-xs);
		color: var(--text);
	}

	.alert-content p {
		margin: 0;
		color: var(--text-muted);
	}

	.instructions-card {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		padding: var(--space-xl);
	}

	.instructions-card h3 {
		margin: 0 0 var(--space-lg) 0;
		color: var(--text);
		font-size: 1.5rem;
	}

	.steps {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.step {
		display: flex;
		gap: var(--space-md);
		align-items: flex-start;
	}

	.step-number {
		width: 32px;
		height: 32px;
		border-radius: 50%;
		background: var(--primary);
		color: white;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: bold;
		flex-shrink: 0;
	}

	.step-content h4 {
		margin: 0 0 var(--space-xs) 0;
		color: var(--text);
		font-size: 1.1rem;
	}

	.step-content p {
		margin: 0;
		color: var(--text-muted);
		line-height: 1.5;
	}

	.step-content a {
		color: var(--primary);
		text-decoration: none;
	}

	.step-content a:hover {
		text-decoration: underline;
	}

	.security-notice {
		display: flex;
		gap: var(--space-md);
		padding: var(--space-lg);
		background: var(--bg-secondary);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		align-items: flex-start;
	}

	.notice-icon {
		font-size: 2rem;
		flex-shrink: 0;
	}

	.notice-content h4 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
		font-size: 1.2rem;
	}

	.notice-content ul {
		margin: 0;
		padding-left: var(--space-lg);
		color: var(--text-muted);
		line-height: 1.6;
	}

	.notice-content li {
		margin-bottom: var(--space-xs);
	}

	.connected-section {
		display: flex;
		flex-direction: column;
		gap: var(--space-xl);
	}

	.connected-card {
		background: var(--surface);
		border: 1px solid var(--success);
		border-radius: var(--radius-lg);
		padding: var(--space-xl);
	}

	.connected-header {
		display: flex;
		gap: var(--space-lg);
		align-items: center;
		margin-bottom: var(--space-lg);
	}

	.connected-icon {
		font-size: 3rem;
		flex-shrink: 0;
	}

	.connected-info h2 {
		margin: 0 0 var(--space-sm) 0;
		color: var(--text);
		font-size: 1.8rem;
	}

	.connected-info p {
		margin: 0 0 var(--space-md) 0;
		color: var(--text-muted);
		font-size: 1.1rem;
	}

	.environment-badge {
		display: inline-block;
		padding: var(--space-xs) var(--space-md);
		border-radius: var(--radius-md);
		font-size: 0.9rem;
		font-weight: bold;
	}

	.environment-badge.test {
		background: var(--warning);
		color: white;
	}

	.environment-badge.live {
		background: var(--success);
		color: white;
	}

	.connected-stats {
		display: flex;
		gap: var(--space-lg);
		margin-bottom: var(--space-lg);
		flex-wrap: wrap;
	}

	.stat {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: var(--space-md);
		background: var(--bg-secondary);
		border-radius: var(--radius-lg);
		min-width: 120px;
	}

	.stat-value {
		font-size: 2rem;
		font-weight: bold;
		color: var(--primary);
		margin-bottom: var(--space-xs);
	}

	.stat-label {
		font-size: 0.9rem;
		color: var(--text-muted);
	}

	.connected-actions {
		display: flex;
		gap: var(--space-md);
		flex-wrap: wrap;
	}

	.next-steps-card {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		padding: var(--space-xl);
	}

	.next-steps-card h3 {
		margin: 0 0 var(--space-lg, 1.5rem) 0;
		color: var(--text-primary, #111827);
		font-size: 1.5rem;
	}

	.portal-form {
		display: flex;
		flex-direction: column;
		gap: var(--space-md, 1rem);
		margin-top: var(--space-lg, 1.5rem);
	}

	.portal-form-actions {
		display: flex;
		gap: var(--space-md, 1rem);
		margin-top: var(--space-md, 1rem);
	}

	.next-steps {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.next-step {
		display: flex;
		gap: var(--space-md);
		align-items: center;
		padding: var(--space-md);
		background: var(--bg-secondary);
		border-radius: var(--radius-md);
	}

	.next-step-icon {
		font-size: 2rem;
		flex-shrink: 0;
	}

	.next-step-content {
		flex: 1;
	}

	.next-step-content h4 {
		margin: 0 0 var(--space-xs) 0;
		color: var(--text);
		font-size: 1.1rem;
	}

	.next-step-content p {
		margin: 0 0 var(--space-sm) 0;
		color: var(--text-muted);
		font-size: 0.9rem;
	}

	.loading {
		text-align: center;
		padding: var(--space-xl);
	}

	.spinner {
		width: 40px;
		height: 40px;
		border: 4px solid var(--border);
		border-top: 4px solid var(--primary);
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin: 0 auto var(--space-md);
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
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
		backdrop-filter: blur(4px);
	}

	.modal-content {
		background: var(--surface, white);
		border-radius: var(--radius-lg, 0.5rem);
		box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
		max-width: 500px;
		width: 90%;
		max-height: 90vh;
		overflow-y: auto;
		border: 1px solid var(--border, #e5e7eb);
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-lg, 1.5rem);
		border-bottom: 1px solid var(--border, #e5e7eb);
	}

	.modal-header h3 {
		margin: 0;
		color: var(--text, #111827);
		font-size: 1.25rem;
		font-weight: 600;
	}

	.modal-close {
		background: none;
		border: none;
		font-size: 1.5rem;
		cursor: pointer;
		color: var(--text-muted, #6b7280);
		padding: 0;
		width: 32px;
		height: 32px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: var(--radius-md, 0.375rem);
		transition: all 0.2s ease;
	}

	.modal-close:hover {
		background: var(--surface-hover, #f3f4f6);
		color: var(--text, #111827);
	}

	.modal-body {
		padding: var(--space-lg, 1.5rem);
	}

	.modal-body p {
		margin: 0 0 var(--space-md, 1rem) 0;
		color: var(--text, #111827);
		line-height: 1.5;
	}

	.modal-body ul {
		margin: 0 0 var(--space-lg, 1.5rem) 0;
		padding-left: var(--space-lg, 1.5rem);
		color: var(--text-muted, #6b7280);
	}

	.modal-body li {
		margin-bottom: var(--space-xs, 0.5rem);
	}

	.confirmation-input {
		margin-top: var(--space-lg, 1.5rem);
	}

	.confirmation-input code {
		background: var(--bg-secondary, #f9fafb);
		padding: 0.125rem 0.25rem;
		border-radius: var(--radius-sm, 0.25rem);
		font-family: 'Courier New', monospace;
		font-size: 0.875rem;
		color: var(--primary, #2563eb);
		font-weight: 600;
	}

	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-md, 1rem);
		padding: var(--space-lg, 1.5rem);
		border-top: 1px solid var(--border, #e5e7eb);
		background: var(--bg-secondary, #f9fafb);
		border-radius: 0 0 var(--radius-lg, 0.5rem) var(--radius-lg, 0.5rem);
	}

	.saved-portal {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
		padding: var(--space-md);
		background: var(--bg-secondary);
		border-radius: var(--radius-md);
	}

	.saved-portal-display {
		display: flex;
		align-items: center;
		gap: var(--space-md);
	}

	.saved-portal-label {
		font-weight: 600;
		color: var(--text);
		font-size: 1rem;
	}

	.saved-portal-url {
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.saved-portal-url a {
		color: var(--primary);
		text-decoration: none;
	}

	.saved-portal-url a:hover {
		text-decoration: underline;
	}

	.saved-portal-actions {
		display: flex;
		gap: var(--space-md);
		flex-wrap: wrap;
	}

	@media (max-width: 768px) {
		.setup-container {
			padding: var(--space-md);
		}

		.connected-header {
			flex-direction: column;
			text-align: center;
		}

		.connected-stats {
			justify-content: center;
		}

		.connected-actions {
			justify-content: center;
		}

		.next-step {
			flex-direction: column;
			text-align: center;
		}
	}

	/* Stripe Data Management Styles */
	.data-management-card {
		background: var(--surface, white);
		border: 1px solid var(--border, #e5e7eb);
		border-radius: var(--radius-lg, 0.5rem);
		padding: var(--space-xl, 2rem);
		margin-top: var(--space-xl, 2rem);
	}

	.data-management-card h3 {
		margin: 0 0 var(--space-sm, 0.5rem) 0;
		color: var(--text, #111827);
		font-size: 1.95rem;
		font-weight: 600;
	}

	.data-management-card > p {
		margin: 0 0 var(--space-lg, 1.5rem) 0;
		color: var(--text-muted, #6b7280);
	}

	.sync-status {
		border-radius: var(--radius-md, 0.375rem);
		padding: var(--space-md, 1rem);
		margin-bottom: var(--space-lg, 1.5rem);
		border: 1px solid;
	}

	.sync-status.success {
		background: var(--success-light, #ecfdf5);
		border-color: var(--success, #10b981);
		color: var(--success-dark, #047857);
	}

	.sync-status.warning {
		background: var(--warning-light, #fffbeb);
		border-color: var(--warning, #f59e0b);
		color: var(--warning-dark, #d97706);
	}

	.sync-status.error {
		background: var(--error-light, #fef2f2);
		border-color: var(--error, #ef4444);
		color: var(--error-dark, #dc2626);
	}

	.sync-status-message {
		font-weight: 500;
	}

	.database-stats {
		margin-bottom: var(--space-lg, 1.5rem);
	}

	.database-stats h4 {
		margin: 0 0 var(--space-md, 1rem) 0;
		color: var(--text, #111827);
		font-size: 1.125rem;
		font-weight: 600;
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
		gap: var(--space-md, 1rem);
		margin-bottom: var(--space-md, 1rem);
	}

	.stat-card {
		background: var(--surface-secondary, #f8fafc);
		border: 1px solid var(--border, #e5e7eb);
		border-radius: var(--radius-md, 0.375rem);
		padding: var(--space-md, 1rem);
		text-align: center;
	}

	.stat-value {
		display: block;
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--primary, #3b82f6);
		margin-bottom: var(--space-xs, 0.25rem);
	}

	.stat-label {
		font-size: 0.875rem;
		color: var(--text-muted, #6b7280);
		font-weight: 500;
	}

	.last-updated {
		font-size: 0.875rem;
		color: var(--text-muted, #6b7280);
		font-style: italic;
	}

	.management-actions {
		display: flex;
		flex-direction: row;
		flex-wrap: wrap;
		gap: var(--space-lg, 1.5rem);
		justify-content: space-between;
	}

	.action-group h4 {
		margin: 0 0 var(--space-md, 1rem) 0;
		color: var(--text, #111827);
		font-size: 1.75rem;
		font-weight: 600;
	}

	.button-group {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-sm, 0.5rem);
		justify-content: center;
	}

	.system-panel {
		margin-top: var(--space-lg, 1.5rem);
		padding-top: var(--space-lg, 1.5rem);
		border-top: 1px solid var(--border, #e5e7eb);
	}

	.system-panel h4, .system-panel h5 {
		margin: 0 0 var(--space-md, 1rem) 0;
		color: var(--text, #111827);
		font-weight: 600;
	}

	.system-health .health-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: var(--space-md, 1rem);
		margin-bottom: var(--space-lg, 1.5rem);
	}

	.health-card {
		display: flex;
		align-items: center;
		gap: var(--space-sm, 0.5rem);
		background: var(--surface-secondary, #f8fafc);
		border: 1px solid var(--border, #e5e7eb);
		border-radius: var(--radius-md, 0.375rem);
		padding: var(--space-md, 1rem);
	}

	.health-card.status-healthy {
		border-color: var(--success, #10b981);
		background: var(--success-light, #ecfdf5);
	}

	.health-card.status-degraded {
		border-color: var(--warning, #f59e0b);
		background: var(--warning-light, #fffbeb);
	}

	.health-card.status-unhealthy {
		border-color: var(--error, #ef4444);
		background: var(--error-light, #fef2f2);
	}

	.health-icon {
		font-size: 1.25rem;
	}

	.health-info {
		flex: 1;
	}

	.health-label {
		font-size: 0.875rem;
		color: var(--text-muted, #6b7280);
		margin-bottom: var(--space-xs, 0.25rem);
	}

	.health-value {
		font-weight: 600;
		color: var(--text, #111827);
		text-transform: capitalize;
	}

	.sync-jobs {
		margin-top: var(--space-lg, 1.5rem);
	}

	.jobs-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm, 0.5rem);
	}

	.job-card {
		display: flex;
		justify-content: space-between;
		align-items: center;
		background: var(--surface-secondary, #f8fafc);
		border: 1px solid var(--border, #e5e7eb);
		border-radius: var(--radius-md, 0.375rem);
		padding: var(--space-sm, 0.5rem) var(--space-md, 1rem);
	}

	.job-card.status-completed {
		border-color: var(--success, #10b981);
		background: var(--success-light, #ecfdf5);
	}

	.job-card.status-running {
		border-color: var(--primary, #3b82f6);
		background: var(--primary-light, #eff6ff);
	}

	.job-card.status-failed {
		border-color: var(--error, #ef4444);
		background: var(--error-light, #fef2f2);
	}

	.job-info {
		flex: 1;
	}

	.job-type {
		font-weight: 500;
		color: var(--text, #111827);
		font-size: 0.875rem;
	}

	.job-status {
		font-size: 0.75rem;
		color: var(--text-muted, #6b7280);
		text-transform: capitalize;
	}

	.job-progress {
		font-size: 0.875rem;
		color: var(--text-muted, #6b7280);
		font-weight: 500;
	}

	/* Public Settings Styles */
	.public-settings-grid {
		display: grid;
		gap: 1.5rem;
		margin-top: 1rem;
	}

	.setting-item {
		border: 1px solid var(--border-color);
		border-radius: 12px;
		padding: 1.5rem;
		background: var(--bg-secondary);
	}

	.setting-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.5rem;
	}

	.setting-header h4 {
		margin: 0;
		font-size: 1.1rem;
		font-weight: 600;
		color: var(--text-primary);
	}

	.setting-type {
		background: var(--success-bg);
		color: var(--success-text);
		padding: 0.25rem 0.75rem;
		border-radius: 12px;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
	}

	.setting-description {
		color: var(--text-secondary);
		font-size: 0.9rem;
		margin-bottom: 1rem;
		line-height: 1.4;
	}

	.saved-setting {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 1rem;
	}

	.saved-setting-display {
		flex: 1;
	}

	.key-display {
		background: var(--bg-tertiary);
		padding: 0.5rem 0.75rem;
		border-radius: 6px;
		font-family: 'Courier New', monospace;
		font-size: 0.9rem;
		color: var(--text-primary);
		border: 1px solid var(--border-color);
	}

	.portal-link {
		color: var(--primary-color);
		text-decoration: none;
		font-size: 0.9rem;
	}

	.portal-link:hover {
		text-decoration: underline;
	}

	.saved-setting-actions {
		display: flex;
		gap: 0.5rem;
	}

	.setting-form {
		margin-top: 1rem;
	}

	.setting-form-actions {
		display: flex;
		gap: 0.75rem;
		margin-top: 1rem;
	}

	.btn-sm {
		padding: 0.5rem 1rem;
		font-size: 0.875rem;
	}

	@media (max-width: 768px) {
		.stats-grid {
			grid-template-columns: repeat(2, 1fr);
		}

		.button-group {
			flex-direction: column;
		}

		.system-health .health-grid {
			grid-template-columns: 1fr;
		}
	}

	/* Webhook Configuration Styles */
	.webhook-config {
		display: grid;
		gap: 1.5rem;
		margin-top: 1rem;
	}

	.web-hook-whole {
		margin-top: var(--space-lg);
		border-radius: 50px;
        background: var(--bg-secondary);
        box-shadow:  5px 5px 10px var(--bg-secondary),
             -5px -5px 10px var(--bg-secondary);
	}

	.webhook-url-display {
		margin-top: 0.75rem;
	}

	.url-container {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		background: var(--bg-tertiary);
		border: 1px solid var(--border-color);
		border-radius: 8px;
		padding: 0.75rem;
	}

	.url-actions {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.webhook-url {
		flex: 1;
		font-family: 'Courier New', monospace;
		font-size: 0.875rem;
		color: var(--text-primary);
		background: transparent;
		border: none;
		word-break: break-all;
	}

	.copy-btn {
		flex-shrink: 0;
		white-space: nowrap;
	}

	.webhook-stats {
		display: flex;
		gap: 1.5rem;
		margin-top: 0.75rem;
		padding: 0.75rem;
		background: var(--bg-tertiary);
		border-radius: 8px;
		border: 1px solid var(--border-color);
	}

	.stat-item {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.25rem;
	}

	.stat-value {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--primary);
	}

	.stat-label {
		font-size: 0.75rem;
		color: var(--text-muted);
		text-transform: uppercase;
		font-weight: 500;
	}

	.webhook-instructions {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		margin-top: 0.75rem;
	}

	.instruction-step {
		display: flex;
		gap: 0.75rem;
		align-items: flex-start;
	}

	.step-number {
		flex-shrink: 0;
		width: 1.5rem;
		height: 1.5rem;
		background: var(--primary);
		color: white;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.75rem;
		font-weight: 600;
	}

	.step-content {
		flex: 1;
	}

	.step-content strong {
		display: block;
		margin-bottom: 0.25rem;
		color: var(--text-primary);
	}

	.step-content p {
		margin: 0;
		color: var(--text-secondary);
		font-size: 0.875rem;
		line-height: 1.4;
	}

	.step-content code {
		background: var(--bg-tertiary);
		padding: 0.125rem 0.375rem;
		border-radius: 4px;
		font-size: 0.8rem;
		border: 1px solid var(--border-color);
	}

	.loading-small {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		color: var(--text-muted);
		font-size: 0.875rem;
	}

	.spinner-small {
		width: 1rem;
		height: 1rem;
		border: 2px solid var(--border-color);
		border-top: 2px solid var(--primary);
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	.setting-type.success {
		background: var(--success-bg);
		color: var(--success-text);
	}

	.setting-type.warning {
		background: var(--warning-bg);
		color: var(--warning-text);
	}

	.setting-type.secure {
		background: #1f2937;
		color: #f3f4f6;
	}
</style> 
