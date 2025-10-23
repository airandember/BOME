<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { showToast } from '$lib/toast';
	import { apiRequest } from '$lib/auth';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let loading = true;
	let success = false;
	let error = '';
	let message = '';

	// Get token from URL params
	$: token = $page.url.searchParams.get('token');

	onMount(async () => {
		if (!token) {
			error = 'No verification token found';
			loading = false;
			return;
		}

		try {
			await verifyEmail(token);
		} catch (err) {
			console.error('Error verifying email:', err);
			loading = false;
		}
	});

	async function verifyEmail(token: string) {
		try {
			console.log('🔍 Verifying email token:', token.substring(0, 8) + '...');
			
			const response = await apiRequest(`/auth/verify-email/${token}`);
			
			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Failed to verify email');
			}

			const result = await response.json();
			success = true;
			message = result.message || 'Email verified successfully!';
			loading = false;

			showToast('Email verified successfully! You can now sign in.', 'success');
			
			// Redirect to login after 3 seconds
			setTimeout(() => {
				goto('/login');
			}, 3000);
			
		} catch (err: any) {
			console.error('Error verifying email:', err);
			error = err.message || 'Failed to verify email';
			loading = false;
		}
	}

	function goToLogin() {
		goto('/login');
	}

	function goToHome() {
		goto('/');
	}
</script>

<svelte:head>
	<title>Email Verification - BOME</title>
	<meta name="description" content="Verify your email address to complete your account setup" />
</svelte:head>

<div class="verification-page">
	<div class="container">
		{#if loading}
			<div class="loading-container">
				<LoadingSpinner />
				<p>Verifying your email address...</p>
			</div>
		{:else if success}
			<div class="success-container">
				<div class="success-icon">✅</div>
				<h1>Email Verified!</h1>
				<p class="success-message">{message}</p>
				
				<div class="success-info">
					<p>Your email has been successfully verified. You can now:</p>
					<ul>
						<li>Sign in to your account</li>
						<li>Access all premium features</li>
						<li>Receive important notifications</li>
						<li>Subscribe to premium plans</li>
					</ul>
				</div>

				<div class="success-actions">
					<button class="btn btn-primary" on:click={goToLogin}>
						Sign In Now
					</button>
					<button class="btn btn-outline" on:click={goToHome}>
						Back to Home
					</button>
				</div>
				
				<p class="redirect-notice">You'll be redirected to the login page in a few seconds...</p>
			</div>
		{:else if error}
			<div class="error-container">
				<div class="error-icon">❌</div>
				<h1>Verification Failed</h1>
				<p class="error-message">{error}</p>
				
				<div class="error-help">
					<h3>What can you do?</h3>
					<ul>
						<li>Check if you clicked the correct link from your email</li>
						<li>Make sure the verification link hasn't expired (24 hours)</li>
						<li>Try requesting a new verification email</li>
						<li>Contact support if the problem persists</li>
					</ul>
				</div>

				<div class="error-actions">
					<button class="btn btn-primary" on:click={goToLogin}>
						Go to Login
					</button>
					<button class="btn btn-outline" on:click={goToHome}>
						Back to Home
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	.verification-page {
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

	.success-container {
		border-color: var(--success-color);
	}

	.error-container {
		border-color: var(--error-color);
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

	.success-info,
	.error-help {
		text-align: left;
		margin: 2rem 0;
		padding: 2rem;
		background: var(--bg-secondary);
		border-radius: 16px;
		border: 1px solid var(--border-color);
	}

	.success-info h3,
	.error-help h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1rem;
		text-align: center;
	}

	.success-info ul,
	.error-help ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.success-info li,
	.error-help li {
		display: flex;
		align-items: center;
		padding: 0.5rem 0;
		color: var(--text-primary);
		font-size: 1rem;
	}

	.success-info li::before {
		content: '✓';
		color: var(--success-color);
		font-weight: bold;
		margin-right: 0.75rem;
	}

	.error-help li::before {
		content: '•';
		color: var(--primary-color);
		font-weight: bold;
		margin-right: 0.75rem;
	}

	.success-actions,
	.error-actions {
		display: flex;
		justify-content: center;
		gap: 1rem;
		flex-wrap: wrap;
		margin-top: 2rem;
	}

	.redirect-notice {
		margin-top: 2rem;
		font-size: 0.875rem;
		color: var(--text-secondary);
		font-style: italic;
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

		.success-info,
		.error-help {
			padding: 1.5rem;
		}
	}
</style>
