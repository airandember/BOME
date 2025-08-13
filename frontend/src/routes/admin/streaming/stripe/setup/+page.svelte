<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';

	let secret = '';
	let saving = false;
	let error = '';
	let success = '';
	let summary: any = null;
	let loading = true;

	// Modal state for clear key confirmation
	let showClearModal = false;
	let clearConfirmText = '';

	export let data: any = null;

	onMount(async () => {
		if (data) {
			summary = data;
			loading = false;
		} else {
			await fetchSummary();
		}
	});

	async function fetchSummary() {
		try {
			loading = true;
			error = '';
			const res = await apiRequest('/admin/streaming/stripe/summary');
			if (res.ok) {
				const data = await res.json();
				summary = data.summary;
			} else {
				error = 'Failed to load summary';
			}
		} catch (err) {
			error = 'Failed to load summary';
			console.error(err);
		} finally {
			loading = false;
		}
	}

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
				await fetchSummary(); // Refresh the summary
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

	async function clearKey() {
		saving = true;
		error = '';
		success = '';
		
		try {
			const res = await apiRequest('/admin/streaming/stripe/secret', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({ key: 'sk_1337' })
			});
			
			if (res.ok) {
				success = 'Stripe key cleared successfully!';
				summary = { enabled: false };
			} else {
				const errorData = await res.json();
				error = errorData.error || 'Failed to clear key';
			}
		} catch (err) {
			error = 'Failed to clear key';
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
		error = '';
		success = '';
		
		try {
			const res = await apiRequest('/admin/streaming/stripe/secret', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({ key: 'sk_1337' })
			});
			
			if (res.ok) {
				success = 'Stripe key cleared successfully!';
				summary = { enabled: false };
			} else {
				const errorData = await res.json();
				error = errorData.error || 'Failed to clear key';
			}
		} catch (err) {
			error = 'Failed to clear key';
			console.error(err);
		} finally {
			saving = false;
		}
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

		{#if !summary?.enabled}
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
			</div>
		{:else}
			<!-- Already Connected -->
			<div class="connected-section">
				<div class="connected-card">
					<div class="connected-header">
						<div class="connected-icon">✅</div>
						<div class="connected-info">
							<h2>Stripe Connected Successfully</h2>
							<p>Your Stripe account is connected and ready to process payments</p>
							<div class="environment-badge {summary.environment === 'live' ? 'live' : 'test'}">
								{summary.environment === 'live' ? '🔴 LIVE MODE' : '🟡 TEST MODE'}
							</div>
						</div>
					</div>

					<div class="connected-stats">
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
					</div>

					<div class="connected-actions">
						<button class="btn btn-outline" on:click={() => { summary = { enabled: false }; }}>
							🔑 Update Key
						</button>
						<button class="btn btn-secondary" on:click={fetchSummary}>
							🔄 Refresh Connection
						</button>
						<button class="btn btn-danger" on:click={showClearConfirmation}>
							🗑️ Clear Key
						</button>
					</div>
				</div>

				<!-- Next Steps -->
				<div class="next-steps-card">
					<h3>🎯 Next Steps</h3>
					<div class="next-steps">
						<div class="next-step">
							<div class="next-step-icon">📦</div>
							<div class="next-step-content">
								<h4>Create Products</h4>
								<p>Set up your products and services in Stripe</p>
								<a href="/admin/streaming/stripe/products" class="btn btn-sm btn-outline">Manage Products</a>
							</div>
						</div>
						
						<div class="next-step">
							<div class="next-step-icon">💰</div>
							<div class="next-step-content">
								<h4>Configure Pricing</h4>
								<p>Set up pricing plans and subscription tiers</p>
								<a href="/admin/streaming/stripe/products" class="btn btn-sm btn-outline">Set Pricing</a>
							</div>
						</div>
						
						<div class="next-step">
							<div class="next-step-icon">🔗</div>
							<div class="next-step-content">
								<h4>Setup Webhooks</h4>
								<p>Configure webhooks for real-time updates</p>
								<button class="btn btn-sm btn-outline">Configure Webhooks</button>
							</div>
						</div>
					</div>
				</div>
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
	.setup-container {
		padding: var(--space-lg);
		max-width: 800px;
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
		margin: 0 0 var(--space-lg) 0;
		color: var(--text);
		font-size: 1.5rem;
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
</style> 