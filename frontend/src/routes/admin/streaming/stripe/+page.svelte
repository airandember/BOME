<script lang="ts">
	// @ts-nocheck
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';

	// Import child components
	import Overview from './overview/+page.svelte';
	import Products from './products/+page.svelte';
	import Customers from './customers/+page.svelte';
	import Coupons from './coupons/+page.svelte';
	import Invoices from './invoices/+page.svelte';
	import Payments from './payments/+page.svelte';
	import Subscriptions from './subscriptions/+page.svelte';
	import Setup from './setup/+page.svelte';

	let summary: any = null;
	let loading = true;
	let error = '';
	let activeTab = 'overview';

	// Setup form state (for main page setup)
	let secret = '';
	let saving = false;
	let setupError = '';
	let setupSuccess = '';

	// Modal state for clear key confirmation
	let showClearModal = false;
	let clearConfirmText = '';

	// Customer portal link state
	let portalLink = '';
	let savedPortalLink = ''; // The saved/persisted value
	let savingPortal = false;
	let portalError = '';
	let portalSuccess = '';
	let editingPortal = false; // Whether we're in edit mode

	// Debug logging for data changes
	$: {
		console.log('=== MAIN STRIPE DEBUG ===');
		console.log('Summary data changed:', summary);
		console.log('Summary enabled:', summary?.enabled);
		console.log('Should show setup form:', !summary?.enabled);
		console.log('Coupons count:', summary?.coupons_count);
		console.log('Coupons array length:', summary?.coupons?.length);
		console.log('Active tab:', activeTab);
		console.log('========================');
	}

	// Tab configuration
	const tabs = [
		{ id: 'overview', name: 'Overview', icon: '📊', component: Overview },
		{ id: 'products', name: 'Products', icon: '📦', component: Products },
		{ id: 'customers', name: 'Customers', icon: '👥', component: Customers },
		{ id: 'coupons', name: 'Coupons', icon: '🎟️', component: Coupons },
		{ id: 'invoices', name: 'Invoices', icon: '📄', component: Invoices },
		{ id: 'payments', name: 'Payments', icon: '💳', component: Payments },
		{ id: 'subscriptions', name: 'Subscriptions', icon: '🔄', component: Subscriptions },
		{ id: 'setup', name: 'Setup', icon: '⚙️', component: Setup }
	];

	onMount(async () => {
		await fetchSummary();
		await loadPortalLink();
		
		// If not enabled, default to setup tab
		if (summary && !summary.enabled) {
			activeTab = 'setup';
		}
	});

	async function fetchSummary() {
		try {
			loading = true;
			error = '';
			console.log('🔍 Fetching Stripe summary...');
			const res = await apiRequest('/admin/streaming/stripe/summary');
			if (res.ok) {
				const data = await res.json();
				console.log('📥 Raw API response:', data);
				summary = data.summary;
				console.log('📊 Summary assigned:', summary);
				console.log('🔧 Summary enabled:', summary?.enabled);
			} else {
				error = 'Failed to load Stripe data';
				console.error('❌ API request failed:', res.status, res.statusText);
			}
		} catch (err) {
			error = 'Failed to load Stripe data';
			console.error('❌ Fetch summary error:', err);
		} finally {
			loading = false;
		}
	}

	async function loadPortalLink() {
		try {
			const res = await apiRequest('/admin/streaming/stripe/portal-link');
			if (res.ok) {
				const data = await res.json();
				savedPortalLink = data.portal_url;
				portalLink = savedPortalLink; // Initialize portalLink for editing
			} else {
				console.error('Failed to load portal link:', res.status);
				portalError = 'Failed to load customer portal link';
			}
		} catch (err) {
			console.error('Failed to load portal link:', err);
			portalError = 'Failed to load customer portal link';
		}
	}

	function switchTab(tabId: string) {
		activeTab = tabId;
	}

	// Setup form functions for main page
	async function saveSecret() {
		if (!secret.trim()) return;
		
		saving = true;
		setupError = '';
		setupSuccess = '';
		
		try {
			const res = await apiRequest('/admin/streaming/stripe/secret', {
				method: 'POST',
				body: JSON.stringify({ key: secret })
			});
			
			if (res.ok) {
				setupSuccess = 'Stripe key saved successfully!';
				secret = '';
				await fetchSummary(); // Refresh the summary
			} else {
				const errorData = await res.json();
				setupError = errorData.error || 'Failed to save key';
			}
		} catch (err) {
			setupError = 'Failed to save key';
			console.error(err);
		} finally {
			saving = false;
		}
	}

	// Show the clear confirmation modal
	function showClearConfirmation() {
		showClearModal = true;
		clearConfirmText = '';
	}

	// Close the modal and reset
	function closeClearModal() {
		showClearModal = false;
		clearConfirmText = '';
	}

	// Confirm and execute the clear action
	async function confirmClearKey() {
		if (clearConfirmText !== 'sk_1337') {
			return; // Don't proceed if confirmation text doesn't match
		}

		// Close modal first
		closeClearModal();

		// Execute the clear action
		saving = true;
		setupError = '';
		setupSuccess = '';
		
		try {
			const res = await apiRequest('/admin/streaming/stripe/secret', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({ key: 'sk_1337' })
			});
			
			if (res.ok) {
				setupSuccess = 'Stripe key cleared successfully!';
				console.log('🔄 Clear key successful, fetching fresh summary...');
				
				// Force reactivity by clearing summary first
				summary = null;
				await fetchSummary(); // Refresh the summary
				
				console.log('✅ Fresh summary loaded:', summary);
				console.log('✅ Summary enabled status:', summary?.enabled);
				console.log('✅ Should show setup form:', !summary?.enabled);
			} else {
				const errorData = await res.json();
				setupError = errorData.error || 'Failed to clear key';
			}
		} catch (err) {
			setupError = 'Failed to clear key';
			console.error(err);
		} finally {
			saving = false;
		}
	}

	// Save customer portal link
	async function savePortalLink() {
		if (!portalLink.trim()) return;
		
		savingPortal = true;
		portalError = '';
		portalSuccess = '';
		
		try {
			const res = await apiRequest('/admin/streaming/stripe/portal-link', {
				method: 'POST',
				body: JSON.stringify({ portal_url: portalLink })
			});

			if (res.ok) {
				const data = await res.json();
				savedPortalLink = portalLink; // Use the local value since backend doesn't return it
				editingPortal = false;
				portalSuccess = data.message || 'Customer portal link saved successfully!';
			} else {
				const errorData = await res.json();
				portalError = errorData.error || 'Failed to save portal link';
			}
		} catch (err) {
			portalError = 'Failed to save portal link';
			console.error(err);
		} finally {
			savingPortal = false;
		}
	}

	// Start editing portal link
	function startEditingPortal() {
		editingPortal = true;
		portalLink = savedPortalLink;
		portalError = '';
		portalSuccess = '';
	}

	// Cancel editing portal link
	function cancelEditingPortal() {
		editingPortal = false;
		portalLink = '';
		portalError = '';
		portalSuccess = '';
	}

	// Clear saved portal link
	async function clearPortalLink() {
		try {
			const res = await apiRequest('/admin/streaming/stripe/portal-link', {
				method: 'DELETE'
			});
			
			if (res.ok) {
				const data = await res.json();
				savedPortalLink = '';
				portalLink = '';
				editingPortal = false;
				portalSuccess = data.message || 'Customer portal link cleared successfully!';
				portalError = '';
			} else {
				const errorData = await res.json();
				portalError = errorData.error || 'Failed to clear portal link';
			}
		} catch (err) {
			portalError = 'Failed to clear portal link';
			console.error(err);
		}
	}

	$: activeTabConfig = tabs.find(tab => tab.id === activeTab);
</script>

{#if loading}
	<div class="loading-container">
		<div class="spinner"></div>
		<p>Loading Stripe Dashboard...</p>
	</div>
{:else if error}
	<div class="error-container">
		<h3>Error Loading Stripe Dashboard</h3>
		<p>{error}</p>
		<button class="btn btn-primary" on:click={fetchSummary}>Retry</button>
	</div>
{:else}
	<div class="stripe-dashboard">
		<!-- Header -->
		<div class="dashboard-header">
			<!--<div class="header-content">
				<h1>Stripe Dashboard</h1>
				<p>Manage payments, subscriptions, and customer data</p>
			</div>-->
			
			{#if summary?.enabled}
				<div class="header-status">
					<div class="status-indicator connected">
						<div class="status-dot"></div>
						<span>Connected</span>
					</div>
					<div class="environment-badge {summary.environment === 'live' ? 'live' : 'test'}">
						{summary.environment === 'live' ? '🔴 LIVE' : '🟡 TEST'}
					</div>
				</div>
			{:else}
				<div class="header-status">
					<div class="status-indicator disconnected">
						<div class="status-dot"></div>
						<span>Not Connected</span>
					</div>
				</div>
			{/if}
		</div>

		{#if !summary?.enabled}
			<!-- Setup Container for Main Page -->
			<div class="setup-container">
				<div class="setup-header">
					<h1>Setup Stripe</h1>
					<p>Configure your Stripe integration to start processing payments</p>
				</div>

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
						
						{#if setupError}
							<div class="alert alert-error">
								<div class="alert-icon">❌</div>
								<div class="alert-content">
									<strong>Connection Failed</strong>
									<p>{setupError}</p>
								</div>
							</div>
						{/if}
						
						{#if setupSuccess}
							<div class="alert alert-success">
								<div class="alert-icon">✅</div>
								<div class="alert-content">
									<strong>Success!</strong>
									<p>{setupSuccess}</p>
								</div>
							</div>
						{/if}
					</div>

					<!-- Setup Instructions -->
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

					<!-- Security Notice -->
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

					<!-- Customer Portal Setup -->
					<div class="instructions-card">
						<h3>🔗 Customer Portal Setup</h3>
						<p>Configure your Stripe customer portal link for subscription management</p>
						
						{#if savedPortalLink && !editingPortal}
							<!-- Display saved portal link -->
							<div class="saved-portal">
								<div class="saved-portal-display">
									<div class="saved-portal-label">Current Portal URL:</div>
									<div class="saved-portal-url">
										<a href={savedPortalLink} target="_blank" rel="noopener">
											{savedPortalLink}
										</a>
									</div>
								</div>
								<div class="saved-portal-actions">
									<button class="btn btn-outline" on:click={startEditingPortal}>
										✏️ Update Link
									</button>
									<button class="btn btn-secondary" on:click={clearPortalLink}>
										🗑️ Clear Link
									</button>
								</div>
							</div>
						{:else}
							<!-- Edit/Add form -->
							<form on:submit|preventDefault={savePortalLink} class="portal-form">
								<div class="input-group">
									<label for="main-portal-link" class="input-label">
										Customer Portal URL
									</label>
									<input 
										id="main-portal-link"
										class="input" 
										type="url" 
										placeholder="https://billing.stripe.com/p/login/..." 
										bind:value={portalLink}
									/>
									<div class="input-help">
										Get this URL from your Stripe Dashboard → Settings → Customer Portal
									</div>
								</div>

								<div class="portal-form-actions">
									<button 
										type="submit" 
										class="btn btn-primary" 
										disabled={savingPortal || !portalLink.trim()}
									>
										{savingPortal ? 'Saving...' : (savedPortalLink ? 'Update Portal Link' : 'Save Portal Link')}
									</button>
									{#if editingPortal}
										<button 
											type="button" 
											class="btn btn-secondary" 
											on:click={cancelEditingPortal}
										>
											Cancel
										</button>
									{/if}
								</div>
							</form>
						{/if}
						
						{#if portalError}
							<div class="alert alert-error">
								<div class="alert-icon">❌</div>
								<div class="alert-content">
									<strong>Error</strong>
									<p>{portalError}</p>
								</div>
							</div>
						{/if}
						
						{#if portalSuccess}
							<div class="alert alert-success">
								<div class="alert-icon">✅</div>
								<div class="alert-content">
									<strong>Success!</strong>
									<p>{portalSuccess}</p>
								</div>
							</div>
						{/if}
					</div>
				</div>
			</div>
		{:else}
			<!-- Tab Navigation -->
			<div class="tab-navigation">
				<div class="tab-list">
					{#each tabs as tab}
						<button 
							class="tab-button {activeTab === tab.id ? 'active' : ''}"
							class:disabled={!summary?.enabled && tab.id !== 'setup'}
							on:click={() => switchTab(tab.id)}
							disabled={!summary?.enabled && tab.id !== 'setup'}
						>
							<span class="tab-icon">{tab.icon}</span>
							<span class="tab-name">{tab.name}</span>
							{#if !summary?.enabled && tab.id !== 'setup'}
								<span class="tab-lock">🔒</span>
							{/if}
						</button>
					{/each}
				</div>
			</div>

			<!-- Tab Content -->
			<div class="tab-content">
				{#if activeTab === 'overview' && activeTabConfig?.component}
					<!-- @ts-ignore -->
					<svelte:component this={activeTabConfig.component} data={summary as any} />
				{:else if activeTab === 'products' && activeTabConfig?.component}
					<!-- @ts-ignore -->
					<svelte:component this={activeTabConfig.component} data={summary as any} />
				{:else if activeTab === 'customers' && activeTabConfig?.component}
					<!-- @ts-ignore -->
					<svelte:component this={activeTabConfig.component} data={summary as any} />
				{:else if activeTab === 'coupons' && activeTabConfig?.component}
					<!-- @ts-ignore -->
					<svelte:component this={activeTabConfig.component} data={summary as any} />
				{:else if activeTab === 'invoices' && activeTabConfig?.component}
					<!-- @ts-ignore -->
					<svelte:component this={activeTabConfig.component} data={summary as any} />
				{:else if activeTab === 'payments' && activeTabConfig?.component}
					<!-- @ts-ignore -->
					<svelte:component this={activeTabConfig.component} data={summary as any} />
				{:else if activeTab === 'subscriptions' && activeTabConfig?.component}
					<!-- @ts-ignore -->
					<svelte:component this={activeTabConfig.component} data={summary as any} />
				{:else if activeTab === 'setup' && activeTabConfig?.component}
					<!-- @ts-ignore -->
					<svelte:component this={activeTabConfig.component} data={summary as any} />



				{/if}
			</div>
		{/if}
	</div>
{/if}

<!-- Clear Key Confirmation Modal -->
{#if showClearModal}
	<div class="modal-overlay" on:click={closeClearModal}>
		<div class="modal-content" on:click|stopPropagation>
			<div class="modal-header">
				<h3>⚠️ Clear Stripe Key</h3>
				<button class="modal-close" on:click={closeClearModal}>&times;</button>
			</div>
			
			<div class="modal-body">
				<p><strong>Are you sure you want to clear your Stripe secret key?</strong></p>
				<p>This action will:</p>
				<ul>
					<li>Disable all Stripe payment processing</li>
					<li>Remove your stored secret key</li>
					<li>Return you to the setup screen</li>
				</ul>
				
				<div class="confirmation-input">
					<label for="confirm-text" class="input-label">
						Type <code>sk_1337</code> to confirm:
					</label>
					<input 
						id="confirm-text"
						class="input" 
						type="text" 
						placeholder="sk_1337"
						bind:value={clearConfirmText}
						on:keydown={(e) => e.key === 'Enter' && clearConfirmText === 'sk_1337' && confirmClearKey()}
					/>
				</div>
			</div>
			
			<div class="modal-footer">
				<button class="btn btn-secondary" on:click={closeClearModal}>
					Cancel
				</button>
				<button 
					class="btn btn-danger" 
					disabled={clearConfirmText !== 'sk_1337' || saving}
					on:click={confirmClearKey}
				>
					{#if saving}
						Clearing...
					{:else}
						🗑️ Clear Key
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.stripe-dashboard {
		min-height: 100vh;
		background: var(--bg-primary);
	}

	.loading-container,
	.error-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--space-xl);
		min-height: 50vh;
	}

	.spinner {
		width: 40px;
		height: 40px;
		border: 4px solid var(--border);
		border-top: 4px solid var(--primary);
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-bottom: var(--space-md);
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.error-container h3 {
		color: var(--error);
		margin-bottom: var(--space-md);
	}

	.btn {
		padding: var(--space-md) var(--space-lg);
		border: none;
		border-radius: var(--radius-md);
		font-size: 1rem;
		cursor: pointer;
		transition: all 0.2s ease;
		font-weight: 600;
	}

	.btn-primary {
		background: var(--primary);
		color: white;
	}

	.btn-primary:hover {
		background: var(--primary-dark);
		transform: translateY(-1px);
	}

	.dashboard-header {
		display: flex;
		justify-content: flex-end;
		align-items: center;
		background: var(--surface);
		border-bottom: 1px solid var(--border);
	}

	.header-content h1 {
		margin: 0 0 var(--space-xs) 0;
		color: var(--text);
		font-size: 2rem;
		font-weight: 700;
	}

	.header-content p {
		margin: 0;
		color: var(--text-muted);
		font-size: 1.1rem;
	}

	.header-status {
		display: flex;
		align-items: center;
		gap: var(--space-md);
	}

	.status-indicator {
		display: flex;
		align-items: center;
		gap: var(--space-xs);
		padding: var(--space-xs) var(--space-md);
		border-radius: var(--radius-md);
		font-size: 0.9rem;
		font-weight: 600;
	}

	.status-indicator.connected {
		background: var(--success-light);
		color: var(--success-dark);
	}

	.status-indicator.disconnected {
		background: var(--error-light);
		color: var(--error-dark);
	}

	.status-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: currentColor;
	}

	.environment-badge {
		padding: var(--space-xs) var(--space-md);
		border-radius: var(--radius-md);
		font-size: 0.8rem;
		font-weight: bold;
	}

	.environment-badge.test {
		background: var(--warning);
		color: white;
	}

	.environment-badge.live {
		background: var(--error);
		color: white;
	}

	.tab-navigation {
		background: var(--surface);
		border-bottom: 1px solid var(--border);
		padding: 0 var(--space-lg);
	}

	.tab-list {
		display: flex;
		gap: var(--space-xs);
		overflow-x: auto;
		scrollbar-width: none;
		-ms-overflow-style: none;
	}

	.tab-list::-webkit-scrollbar {
		display: none;
	}

	.tab-button {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		padding: var(--space-md) var(--space-lg);
		border: none;
		background: transparent;
		color: var(--text-muted);
		cursor: pointer;
		border-bottom: 3px solid transparent;
		transition: all 0.2s ease;
		white-space: nowrap;
		font-size: 0.95rem;
		font-weight: 500;
	}

	.tab-button:hover:not(:disabled) {
		color: var(--text);
		background: var(--bg-secondary);
	}

	.tab-button.active {
		color: var(--primary);
		border-bottom-color: var(--primary);
		background: var(--primary-light);
	}

	.tab-button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.tab-icon {
		font-size: 1.1rem;
	}

	.tab-name {
		font-weight: inherit;
	}

	.tab-lock {
		font-size: 0.8rem;
		opacity: 0.7;
	}

	.tab-content {
		flex: 1;
		min-height: calc(100vh - 200px);
		background: var(--bg-primary);
	}

	.coming-soon {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--space-xl);
		text-align: center;
		min-height: 400px;
	}

	.coming-soon-icon {
		font-size: 4rem;
		margin-bottom: var(--space-lg);
		opacity: 0.5;
	}

	.coming-soon h3 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
		font-size: 1.5rem;
	}

	.coming-soon p {
		margin: 0 0 var(--space-lg) 0;
		color: var(--text-muted);
		font-size: 1.1rem;
	}

	.preview-stats {
		display: flex;
		gap: var(--space-lg);
		justify-content: center;
		flex-wrap: wrap;
	}

	.stat {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: var(--space-md);
		background: var(--surface);
		border-radius: var(--radius-lg);
		border: 1px solid var(--border);
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

	/* New styles for setup container */
	.setup-container {
		padding: var(--space-lg);
		background: var(--surface);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-md);
		margin-top: var(--space-lg);
		border: 1px solid var(--border);
	}

	.setup-header {
		text-align: center;
		margin-bottom: var(--space-lg);
		padding-bottom: var(--space-md);
		border-bottom: 1px solid var(--border);
	}

	.setup-header h1 {
		margin: 0 0 var(--space-xs) 0;
		color: var(--text);
		font-size: 2rem;
		font-weight: 700;
	}

	.setup-header p {
		margin: 0;
		color: var(--text-muted);
		font-size: 1.1rem;
	}

	.setup-section {
		display: flex;
		gap: var(--space-lg);
		flex-wrap: wrap;
		justify-content: center;
	}

	.setup-card {
		flex: 1;
		min-width: 350px;
		padding: var(--space-lg);
		background: var(--bg-secondary);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-sm);
		border: 1px solid var(--border);
	}

	.card-header {
		text-align: center;
		margin-bottom: var(--space-lg);
		padding-bottom: var(--space-md);
		border-bottom: 1px solid var(--border);
	}

	.card-header h2 {
		margin: 0 0 var(--space-xs) 0;
		color: var(--text);
		font-size: 1.8rem;
		font-weight: 700;
	}

	.card-header p {
		margin: 0;
		color: var(--text-muted);
		font-size: 1rem;
	}

	.setup-form {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.input-group {
		position: relative;
	}

	.input-label {
		display: flex;
		align-items: center;
		gap: var(--space-xs);
		font-size: 0.9rem;
		font-weight: 500;
		color: var(--text);
		margin-bottom: var(--space-xs);
	}

	.input-label .required {
		color: var(--error);
		font-size: 0.8rem;
	}

	.input {
		padding: var(--space-md) var(--space-lg);
		border: 1px solid var(--border);
		border-radius: var(--radius-md);
		background: var(--bg-input);
		color: var(--text);
		font-size: 1rem;
		transition: all 0.2s ease;
		width: 100%;
	}

	.input:focus {
		outline: none;
		border-color: var(--primary);
		box-shadow: var(--shadow-sm);
	}

	.input-help {
		font-size: 0.8rem;
		color: var(--text-muted);
		margin-top: var(--space-xs);
	}

	.btn-lg {
		padding: var(--space-md) var(--space-lg);
		font-size: 1rem;
	}

	.alert {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		padding: var(--space-md);
		border-radius: var(--radius-md);
		margin-top: var(--space-md);
	}

	.alert-success {
		background-color: var(--success-light);
		color: var(--success-dark);
	}

	.alert-error {
		background-color: var(--error-light);
		color: var(--error-dark);
	}

	.alert-icon {
		font-size: 1.2rem;
	}

	.alert-content {
		flex: 1;
	}

	.alert-content strong {
		font-weight: 600;
	}

	.alert-content p {
		margin: 0;
		font-size: 0.9rem;
	}

	.instructions-card {
		flex: 1;
		min-width: 350px;
		padding: var(--space-lg);
		background: var(--bg-secondary);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-sm);
		border: 1px solid var(--border);
	}

	.instructions-card h3 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
		font-size: 1.5rem;
		font-weight: 700;
	}

	.steps {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.step {
		display: flex;
		align-items: flex-start;
		gap: var(--space-md);
	}

	.step-number {
		font-size: 1.5rem;
		font-weight: bold;
		color: var(--primary);
		min-width: 20px;
		text-align: center;
	}

	.step-content h4 {
		margin: 0 0 var(--space-xs) 0;
		color: var(--text);
		font-size: 1.2rem;
		font-weight: 600;
	}

	.step-content p {
		margin: 0;
		color: var(--text-muted);
		font-size: 1rem;
	}

	.security-notice {
		display: flex;
		align-items: flex-start;
		gap: var(--space-md);
		padding: var(--space-md);
		background: var(--bg-secondary);
		border-radius: var(--radius-lg);
		border: 1px solid var(--border);
		margin-top: var(--space-lg);
	}

	.notice-icon {
		font-size: 1.5rem;
		color: var(--primary);
	}

	.notice-content h4 {
		margin: 0 0 var(--space-xs) 0;
		color: var(--text);
		font-size: 1.2rem;
		font-weight: 600;
	}

	.notice-content ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.notice-content li {
		margin-bottom: var(--space-xs);
		color: var(--text-muted);
		font-size: 0.9rem;
	}

	.portal-form {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
		margin-top: var(--space-lg);
	}

	.portal-form-actions {
		display: flex;
		gap: var(--space-md);
		justify-content: flex-end;
	}

	.saved-portal {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
		padding: var(--space-md);
		background: var(--bg-secondary);
		border-radius: var(--radius-md);
		border: 1px solid var(--border);
	}

	.saved-portal-display {
		display: flex;
		align-items: center;
		gap: var(--space-md);
	}

	.saved-portal-label {
		font-size: 0.9rem;
		color: var(--text-muted);
		font-weight: 500;
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
		word-break: break-all;
	}

	.saved-portal-url a:hover {
		text-decoration: underline;
	}

	.saved-portal-actions {
		display: flex;
		gap: var(--space-md);
		justify-content: flex-end;
	}

	/* Modal Styles */
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.6);
		display: flex;
		justify-content: center;
		align-items: center;
		z-index: 1000;
	}

	.modal-content {
		background: var(--bg-primary);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-xl);
		width: 90%;
		max-width: 600px;
		max-height: 90vh;
		display: flex;
		flex-direction: column;
		border: 1px solid var(--border);
		overflow: hidden;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-md) var(--space-lg);
		border-bottom: 1px solid var(--border);
		background: var(--surface);
	}

	.modal-header h3 {
		margin: 0;
		color: var(--text);
		font-size: 1.8rem;
		font-weight: 700;
	}

	.modal-close {
		background: none;
		border: none;
		font-size: 2rem;
		color: var(--text-muted);
		cursor: pointer;
		padding: var(--space-xs);
		transition: color 0.2s ease;
	}

	.modal-close:hover {
		color: var(--text);
	}

	.modal-body {
		padding: var(--space-lg);
		overflow-y: auto;
		flex-grow: 1;
	}

	.confirmation-input {
		margin-top: var(--space-md);
	}

	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-md);
		padding: var(--space-md) var(--space-lg);
		border-top: 1px solid var(--border);
		background: var(--surface);
	}

	.btn-secondary {
		background: var(--bg-secondary);
		color: var(--text);
		border: 1px solid var(--border);
	}

	.btn-secondary:hover {
		background: var(--bg-hover);
	}

	.btn-danger {
		background: var(--error);
		color: white;
	}

	.btn-danger:hover {
		background: var(--error-dark);
	}

	@media (max-width: 768px) {
		.dashboard-header {
			flex-direction: column;
			gap: var(--space-md);
			text-align: center;
		}

		.header-content h1 {
			font-size: 1.5rem;
		}

		.tab-list {
			justify-content: flex-start;
		}

		.tab-button {
			padding: var(--space-sm) var(--space-md);
			font-size: 0.9rem;
		}

		.setup-section {
			flex-direction: column;
			align-items: center;
		}

		.setup-card,
		.instructions-card,
		.security-notice {
			width: 100%;
			min-width: auto;
		}

		.preview-stats {
			flex-direction: column;
			align-items: center;
		}
	}
</style> 
