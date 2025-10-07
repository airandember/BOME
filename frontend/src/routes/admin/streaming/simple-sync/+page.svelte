<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';

	let loading = false;
	let message = '';
	let error = '';
	let polling = false;

	async function checkSyncStatus() {
		try {
			const response = await api.get('/admin/streaming/simple-sync/status');
			if (response.data) {
				const status = response.data;
				if (status.status === 'completed') {
					message = status.message || '✅ Sync completed successfully!';
					polling = false;
					return true;
				} else if (status.status === 'failed') {
					error = status.lastError || 'Sync failed';
					polling = false;
					return true;
				} else if (status.status === 'running') {
					message = '🔄 Sync is still running... Please wait.';
					return false;
				}
			}
		} catch (err) {
			console.error('Error checking sync status:', err);
		}
		return false;
	}

	async function pollForCompletion() {
		if (polling) return;
		
		polling = true;
		message = '🔄 Sync started. Checking for completion...';
		
		const maxAttempts = 60; // 5 minutes with 5-second intervals
		let attempts = 0;
		
		const pollInterval = setInterval(async () => {
			attempts++;
			const completed = await checkSyncStatus();
			
			if (completed || attempts >= maxAttempts) {
				clearInterval(pollInterval);
				polling = false;
				if (attempts >= maxAttempts && !completed) {
					message = '⏰ Polling timeout. Check backend logs for sync status.';
				}
			}
		}, 5000); // Check every 5 seconds
	}

	async function runSimpleSync() {
		loading = true;
		error = '';
		message = '';

		try {
			const response = await api.post('/admin/streaming/simple-sync/all');
			
			if (response.data) {
				message = response.data.message || '🔄 Stripe sync started in background.';
				
				// Start polling for completion
				setTimeout(() => {
					pollForCompletion();
				}, 2000); // Wait 2 seconds before starting to poll
			} else if (response.error) {
				error = response.error;
			}
		} catch (err) {
			// @ts-ignore
			error = `Network error: ${err.message}`;
		} finally {
			loading = false;
		}
	}

	async function linkCustomers() {
		loading = true;
		error = '';
		message = '';

		try {
			const response = await api.post('/admin/streaming/simple-sync/link-customers');
			
			if (response.data) {
				message = response.data.message || '✅ Customer linking completed successfully';
			} else if (response.error) {
				error = response.error;
			}
		} catch (err) {
			// @ts-ignore
			error = `Network error: ${err.message}`;
		} finally {
			loading = false;
		}
	}
</script>

<div class="simple-sync-dashboard">
	<h1>🚀 Simple Stripe Sync</h1>
	<p class="description">
		Clean, straightforward Stripe data synchronization without complex validation or ghost detection.
		This follows Stripe's natural data flow: Products → Prices → Customers → Subscriptions.
	</p>

	<div class="sync-controls">
		<button 
			class="sync-button primary" 
			onclick={runSimpleSync} 
			disabled={loading || polling}
		>
			{#if loading}
				🔄 Starting...
			{:else if polling}
				⏳ Checking Status...
			{:else}
				🚀 Start Background Sync
			{/if}
		</button>

		<button 
			class="sync-button secondary" 
			onclick={linkCustomers} 
			disabled={loading}
		>
			{#if loading}
				🔗 Linking...
			{:else}
				🔗 Link Customers to Users
			{/if}
		</button>
	</div>

	{#if message}
		<div class="message success">
			{@html message.replace(/\n/g, '<br>')}
		</div>
	{/if}

	{#if error}
		<div class="message error">
			❌ {error}
		</div>
	{/if}

	<div class="sync-flow">
		<h3>📊 Simple Sync Flow</h3>
		<div class="flow-steps">
			<div class="step">
				<div class="step-number">1</div>
				<div class="step-content">
					<h4>📦 Products</h4>
					<p>Sync all Stripe products (standalone)</p>
				</div>
			</div>
			<div class="step">
				<div class="step-number">2</div>
				<div class="step-content">
					<h4>💰 Prices</h4>
					<p>Sync all prices (linked to products)</p>
				</div>
			</div>
			<div class="step">
				<div class="step-number">3</div>
				<div class="step-content">
					<h4>👥 Customers</h4>
					<p>Sync all customers (standalone)</p>
				</div>
			</div>
			<div class="step">
				<div class="step-number">4</div>
				<div class="step-content">
					<h4>📋 Subscriptions</h4>
					<p>Sync all subscriptions (linked to customers & prices)</p>
				</div>
			</div>
		</div>
	</div>
</div>

<style>
	.simple-sync-dashboard {
		padding: 2rem;
		max-width: 800px;
		margin: 0 auto;
	}

	.description {
		background: #f8f9fa;
		padding: 1rem;
		border-radius: 8px;
		margin-bottom: 2rem;
		border-left: 4px solid #28a745;
	}

	.sync-controls {
		display: flex;
		gap: 1rem;
		margin-bottom: 2rem;
	}

	.sync-button {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 6px;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s;
	}

	.sync-button.primary {
		background: #007bff;
		color: white;
	}

	.sync-button.primary:hover:not(:disabled) {
		background: #0056b3;
	}

	.sync-button.secondary {
		background: #6c757d;
		color: white;
	}

	.sync-button.secondary:hover:not(:disabled) {
		background: #545b62;
	}

	.sync-button:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.message {
		padding: 1rem;
		border-radius: 6px;
		margin-bottom: 1rem;
	}

	.message.success {
		background: #d4edda;
		color: #155724;
		border: 1px solid #c3e6cb;
	}

	.message.error {
		background: #f8d7da;
		color: #721c24;
		border: 1px solid #f5c6cb;
	}

	.sync-flow {
		background: #f8f9fa;
		padding: 1.5rem;
		border-radius: 8px;
		border: 1px solid #dee2e6;
	}

	.flow-steps {
		display: grid;
		gap: 1rem;
		margin-top: 1rem;
	}

	.step {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1rem;
		background: white;
		border-radius: 6px;
		border: 1px solid #dee2e6;
	}

	.step-number {
		width: 2rem;
		height: 2rem;
		background: #007bff;
		color: white;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: bold;
		flex-shrink: 0;
	}

	.step-content h4 {
		margin: 0 0 0.25rem 0;
		color: #333;
	}

	.step-content p {
		margin: 0;
		color: #666;
		font-size: 0.9rem;
	}
</style>
