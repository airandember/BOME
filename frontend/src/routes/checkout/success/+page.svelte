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
	let sessionData: any = null;
	let paymentAmount = 0;
	let currency = 'USD';

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
			console.log('🔍 Verifying session:', sessionId);
			
			// Call our new session verification endpoint
			const response = await apiRequest(`/stripe/session/${sessionId}`);
			
			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Failed to verify session');
			}

			const result = await response.json();
			sessionData = result.data;
			
			console.log('✅ Session verification result:', sessionData);

			// Extract session information
			sessionStatus = sessionData.payment_status || sessionData.status;
			customerEmail = sessionData.customer_email || $auth.user?.email || 'your email';
			paymentAmount = sessionData.amount_total ? sessionData.amount_total / 100 : 0; // Convert from cents
			currency = sessionData.currency?.toUpperCase() || 'USD';

			loading = false;

			// Show appropriate toast based on payment status
			if (sessionStatus === 'paid' || sessionStatus === 'complete') {
				showToast('Payment successful! Your subscription is now active.', 'success');
				
				// TODO: Update user subscription in database
				await updateUserSubscription();
			} else if (sessionStatus === 'unpaid' || sessionStatus === 'requires_payment_method') {
				showToast('Payment incomplete. Please try again.', 'warning');
			} else {
				showToast(`Payment status: ${sessionStatus}`, 'info');
			}
		} catch (err: any) {
			console.error('Error checking session status:', err);
			error = err.message || 'Failed to verify payment status';
			loading = false;
			throw err;
		}
	}

	async function updateUserSubscription() {
		try {
			console.log('🔄 Updating user subscription...');
			
			// Call backend endpoint to create subscription record
			const response = await apiRequest('/stripe/create-subscription', {
				method: 'POST',
				body: JSON.stringify({
					session_id: sessionId,
					session_data: sessionData
				})
			});
			
			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Failed to create subscription record');
			}
			
			const result = await response.json();
			console.log('✅ Subscription created successfully:', result);
			
		} catch (err) {
			console.error('❌ Failed to update subscription:', err);
			// Don't throw here - payment was successful even if DB update failed
			// Show a warning toast but don't fail the success page
			showToast('Payment successful, but there was an issue creating your subscription record. Please contact support.', 'warning');
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
					<button class="btn btn-primary" style="color: --var(--text-primary)" on:click={goToSubscriptions}>
						Back to Subscriptions
					</button>
				</div>
			</div>
		{:else if sessionStatus === 'paid' || sessionStatus === 'complete'}
			<div class="success-container">
				<div class="success-icon">🎉</div>
				<h1>Welcome to BOME!</h1>
				<p class="success-message">
					Your subscription has been activated successfully! 
					{#if paymentAmount > 0}
						You've been charged <strong>{new Intl.NumberFormat('en-US', { style: 'currency', currency: currency }).format(paymentAmount)}</strong>.
					{/if}
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
						<!--<li>
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
						</li>-->
						<li>
							<svg class="check-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polyline points="20,6 9,17 4,12"></polyline>
							</svg>
							Over 1600 videos!!
						</li>
					</ul>
				</div>

				<div class="success-actions">
					<button class="btn btn-primary" style="color: --var(--text-primary)" on:click={goToDashboard}>
						Go to Dashboard
					</button>
					<button class="btn btn-outline" on:click={goToSubscriptions}>
						Manage Subscription
					</button>
				</div>
			</div>
		{:else if sessionStatus === 'unpaid' || sessionStatus === 'requires_payment_method'}
			<div class="warning-container">
				<div class="warning-icon">⚠️</div>
				<h1>Payment Incomplete</h1>
				<p class="warning-message">
					Your payment was not completed successfully. 
					{#if sessionStatus === 'requires_payment_method'}
						Please update your payment method and try again.
					{:else}
						Please try again or contact support if the issue persists.
					{/if}
				</p>
				<div class="warning-actions">
					<button class="btn btn-primary" on:click={goToSubscriptions}>
						Try Again
					</button>
				</div>
			</div>
		{:else if sessionStatus === 'processing'}
			<div class="processing-container">
				<div class="processing-icon">⏳</div>
				<h1>Payment Processing</h1>
				<p class="processing-message">
					Your payment is being processed. This may take a few minutes.
					You'll receive an email confirmation once it's complete.
				</p>
				<div class="processing-actions">
					<button class="btn btn-outline" on:click={() => window.location.reload()}>
						Refresh Status
					</button>
					<button class="btn btn-primary" on:click={goToDashboard}>
						Continue to Dashboard
					</button>
				</div>
			</div>
		{:else}
			<div class="error-container">
				<div class="error-icon">❌</div>
				<h1>Payment Status Unknown</h1>
				<p class="error-message">
					We couldn't determine your payment status: <strong>{sessionStatus}</strong>
					<br>Please contact support for assistance.
				</p>
				<div class="error-actions">
					<button class="btn btn-outline" on:click={() => window.location.reload()}>
						Refresh Status
					</button>
					<button class="btn btn-primary" on:click={goToSubscriptions}>
						Back to Subscriptions
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
	.error-container,
	.warning-container,
	.processing-container {
		text-align: center;
		padding: 3rem 2rem;
		background: var(--card-bg);
		border-radius: 20px;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
	}

	.warning-container {
		border-color: #f59e0b;
	}

	.processing-container {
		border-color: #3b82f6;
	}

	.success-icon,
	.error-icon,
	.warning-icon,
	.processing-icon {
		font-size: 4rem;
		margin-bottom: 1.5rem;
	}

	.success-container h1,
	.error-container h1,
	.warning-container h1,
	.processing-container h1 {
		font-size: 2.5rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.success-message,
	.error-message,
	.warning-message,
	.processing-message {
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
	.error-actions,
	.warning-actions,
	.processing-actions {
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
