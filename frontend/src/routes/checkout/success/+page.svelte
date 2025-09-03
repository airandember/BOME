<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { auth, apiRequest } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let loading = true;
	let sessionStatus = '';
	let customerEmail = '';
	let error = '';

	// Get session ID from URL params
	$: sessionId = $page.url.searchParams.get('session_id');

	onMount(async () => {
		if (!sessionId) {
			error = 'No session ID found';
			loading = false;
			return;
		}

		try {
			await checkSessionStatus();
		} catch (err) {
			console.error('Error checking session status:', err);
			error = 'Failed to verify payment status';
			loading = false;
		}
	});

	async function checkSessionStatus() {
		try {
			// TODO: Create an endpoint to check session status
			// For now, we'll assume success and show a success message
			sessionStatus = 'complete';
			customerEmail = $auth.user?.email || 'your email';
			loading = false;

			if (sessionStatus === 'complete') {
				showToast('Subscription activated successfully!', 'success');
			}
		} catch (err) {
			console.error('Error checking session status:', err);
			throw err;
		}
	}

	function goToDashboard() {
		goto('/dashboard');
	}

	function goToSubscriptions() {
		goto('/subscription');
	}
</script>

<svelte:head>
	<title>Subscription Success - BOME</title>
	<meta name="description" content="Your subscription has been activated successfully" />
</svelte:head>

<div class="success-page">
	<div class="container">
		{#if loading}
			<div class="loading-container">
				<LoadingSpinner />
				<p>Verifying your payment...</p>
			</div>
		{:else if error}
			<div class="error-container">
				<div class="error-icon">⚠️</div>
				<h1>Payment Verification Failed</h1>
				<p class="error-message">{error}</p>
				<div class="error-actions">
					<button class="btn btn-primary" on:click={goToSubscriptions}>
						Back to Subscriptions
					</button>
				</div>
			</div>
		{:else if sessionStatus === 'complete'}
			<div class="success-container">
				<div class="success-icon">🎉</div>
				<h1>Welcome to BOME!</h1>
				<p class="success-message">
					Your subscription has been activated successfully. 
					A confirmation email has been sent to <strong>{customerEmail}</strong>.
				</p>
				
				<div class="success-features">
					<h2>What's Next?</h2>
					<ul>
						<li>
							<svg class="check-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polyline points="20,6 9,17 4,12"></polyline>
							</svg>
							Access exclusive Book of Mormon evidence content
						</li>
						<li>
							<svg class="check-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polyline points="20,6 9,17 4,12"></polyline>
							</svg>
							Download videos for offline viewing
						</li>
						<li>
							<svg class="check-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polyline points="20,6 9,17 4,12"></polyline>
							</svg>
							Join our community forum
						</li>
						<li>
							<svg class="check-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polyline points="20,6 9,17 4,12"></polyline>
							</svg>
							Get early access to new research
						</li>
					</ul>
				</div>

				<div class="success-actions">
					<button class="btn btn-primary" on:click={goToDashboard}>
						Go to Dashboard
					</button>
					<button class="btn btn-outline" on:click={goToSubscriptions}>
						Manage Subscription
					</button>
				</div>
			</div>
		{:else}
			<div class="error-container">
				<div class="error-icon">❌</div>
				<h1>Payment Incomplete</h1>
				<p class="error-message">
					Your payment was not completed successfully. Please try again.
				</p>
				<div class="error-actions">
					<button class="btn btn-primary" on:click={goToSubscriptions}>
						Try Again
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	.success-page {
		min-height: 100vh;
		background: var(--bg-gradient);
		padding: 2rem 0;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.container {
		max-width: 600px;
		margin: 0 auto;
		padding: 0 1rem;
	}

	.loading-container {
		text-align: center;
		padding: 4rem 2rem;
		background: var(--card-bg);
		border-radius: 20px;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
	}

	.loading-container p {
		margin-top: 1rem;
		color: var(--text-secondary);
		font-size: 1.1rem;
	}

	.success-container,
	.error-container {
		text-align: center;
		padding: 3rem 2rem;
		background: var(--card-bg);
		border-radius: 20px;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
	}

	.success-icon,
	.error-icon {
		font-size: 4rem;
		margin-bottom: 1.5rem;
	}

	.success-container h1,
	.error-container h1 {
		font-size: 2.5rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.success-message,
	.error-message {
		font-size: 1.1rem;
		color: var(--text-secondary);
		margin-bottom: 2rem;
		line-height: 1.6;
	}

	.success-features {
		text-align: left;
		margin: 2rem 0;
		padding: 2rem;
		background: var(--bg-secondary);
		border-radius: 16px;
		border: 1px solid var(--border-color);
	}

	.success-features h2 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1rem;
		text-align: center;
	}

	.success-features ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.success-features li {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem 0;
		color: var(--text-primary);
		font-size: 1rem;
	}

	.check-icon {
		width: 20px;
		height: 20px;
		color: var(--success-color);
		flex-shrink: 0;
	}

	.success-actions,
	.error-actions {
		display: flex;
		justify-content: center;
		gap: 1rem;
		flex-wrap: wrap;
		margin-top: 2rem;
	}

	.btn {
		padding: 0.75rem 1.5rem;
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

	.btn-primary {
		background: var(--primary-color);
		color: white;
	}

	.btn-primary:hover {
		background: var(--primary-color-hover);
		transform: translateY(-1px);
	}

	.btn-outline {
		background: white;
		color: var(--text-primary);
		border: 1px solid var(--border-color);
	}

	.btn-outline:hover {
		background: var(--bg-secondary);
		border-color: var(--primary-color);
	}

	@media (max-width: 768px) {
		.success-container h1,
		.error-container h1 {
			font-size: 2rem;
		}

		.success-actions,
		.error-actions {
			flex-direction: column;
			align-items: center;
		}

		.success-features {
			padding: 1.5rem;
		}
	}
</style>
