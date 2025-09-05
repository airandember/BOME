<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, apiRequest } from '$lib/auth';
	import { showToast } from '$lib/toast';

	let user: any = null;
	let loading = false;
	let resendCooldown = 0;
	let cooldownInterval: any = null;

	onMount(() => {
		// Get current user
		auth.subscribe((authState: any) => {
			user = authState.user;
			
			// If user is already verified, redirect to dashboard
			if (user && user.email_verified) {
				if (user.role === 'admin' || user.role === 'super_admin') {
					goto('/admin');
				} else {
					goto('/');
				}
			}
		});
	});

	async function resendVerification() {
		if (!user || resendCooldown > 0) return;

		try {
			loading = true;
			
			const response = await apiRequest('/auth/resend-verification', {
				method: 'POST',
				body: JSON.stringify({
					email: user.email
				})
			});

			if (response.ok) {
				showToast('Verification email sent! Please check your inbox.', 'success');
				startCooldown();
			} else {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Failed to send verification email');
			}

		} catch (err: any) {
			showToast(err.message || 'Failed to send verification email', 'error');
			console.error('Error resending verification:', err);
		} finally {
			loading = false;
		}
	}

	function startCooldown() {
		resendCooldown = 60; // 60 seconds cooldown
		cooldownInterval = setInterval(() => {
			resendCooldown--;
			if (resendCooldown <= 0) {
				clearInterval(cooldownInterval);
			}
		}, 1000);
	}

	function logout() {
		auth.logout();
		goto('/login');
	}
</script>

<svelte:head>
	<title>Email Verification Required - BOME</title>
</svelte:head>

<div class="verification-required-page">
	<div class="container">
		<div class="verification-card">
			<div class="icon-section">
				<div class="email-icon">📧</div>
			</div>
			
			<div class="content-section">
				<h1>Email Verification Required</h1>
				
				{#if user}
					<p class="description">
						Hi <strong>{user.first_name}</strong>! To access your dashboard and all features, 
						please verify your email address.
					</p>
					
					<div class="email-info">
						<p>We sent a verification email to:</p>
						<div class="email-address">{user.email}</div>
					</div>
					
					<div class="instructions">
						<h3>What to do next:</h3>
						<ol>
							<li>Check your email inbox (and spam folder)</li>
							<li>Click the verification link in the email</li>
							<li>Return here to access your dashboard</li>
						</ol>
					</div>
					
					<div class="actions">
						<button 
							class="btn btn-primary"
							onclick={resendVerification}
							disabled={loading || resendCooldown > 0}
						>
							{#if loading}
								<div class="loading-spinner small"></div>
								Sending...
							{:else if resendCooldown > 0}
								Resend in {resendCooldown}s
							{:else}
								Resend Verification Email
							{/if}
						</button>
						
						<button 
							class="btn btn-outline"
							onclick={logout}
						>
							Sign Out
						</button>
					</div>
				{:else}
					<p class="description">
						Please sign in to verify your email address.
					</p>
					
					<div class="actions">
						<a href="/login" class="btn btn-primary">
							Sign In
						</a>
					</div>
				{/if}
			</div>
		</div>
		
		<div class="help-section">
			<h3>Need Help?</h3>
			<div class="help-items">
				<div class="help-item">
					<strong>Didn't receive the email?</strong>
					<p>Check your spam folder or try resending the verification email.</p>
				</div>
				<div class="help-item">
					<strong>Wrong email address?</strong>
					<p>Sign out and create a new account with the correct email address.</p>
				</div>
				<div class="help-item">
					<strong>Still having issues?</strong>
					<p>Contact our support team for assistance.</p>
				</div>
			</div>
		</div>
	</div>
</div>

<style>
	.verification-required-page {
		min-height: 100vh;
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 2rem;
	}

	.container {
		max-width: 600px;
		width: 100%;
	}

	.verification-card {
		background: white;
		border-radius: 1rem;
		box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
		padding: 3rem;
		text-align: center;
		margin-bottom: 2rem;
	}

	.icon-section {
		margin-bottom: 2rem;
	}

	.email-icon {
		font-size: 4rem;
		margin-bottom: 1rem;
	}

	.content-section h1 {
		font-size: 2rem;
		font-weight: 700;
		color: #1f2937;
		margin-bottom: 1rem;
	}

	.description {
		font-size: 1.125rem;
		color: #6b7280;
		margin-bottom: 2rem;
		line-height: 1.6;
	}

	.email-info {
		background: #f3f4f6;
		border-radius: 0.5rem;
		padding: 1.5rem;
		margin-bottom: 2rem;
	}

	.email-info p {
		margin: 0 0 0.5rem 0;
		color: #6b7280;
		font-size: 0.875rem;
	}

	.email-address {
		font-size: 1.125rem;
		font-weight: 600;
		color: #1f2937;
		background: white;
		padding: 0.75rem;
		border-radius: 0.375rem;
		border: 1px solid #d1d5db;
	}

	.instructions {
		text-align: left;
		background: #fef3c7;
		border: 1px solid #f59e0b;
		border-radius: 0.5rem;
		padding: 1.5rem;
		margin-bottom: 2rem;
	}

	.instructions h3 {
		margin: 0 0 1rem 0;
		color: #92400e;
		font-size: 1rem;
		font-weight: 600;
	}

	.instructions ol {
		margin: 0;
		padding-left: 1.25rem;
		color: #92400e;
	}

	.instructions li {
		margin-bottom: 0.5rem;
	}

	.actions {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		align-items: center;
	}

	.btn {
		padding: 0.75rem 2rem;
		border: none;
		border-radius: 0.5rem;
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		min-width: 200px;
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
		transform: translateY(-1px);
	}

	.btn-outline {
		background: white;
		color: #374151;
		border: 2px solid #d1d5db;
	}

	.btn-outline:hover {
		background: #f9fafb;
		border-color: #9ca3af;
	}

	.loading-spinner {
		width: 16px;
		height: 16px;
		border: 2px solid transparent;
		border-top: 2px solid currentColor;
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.help-section {
		background: rgba(255, 255, 255, 0.1);
		backdrop-filter: blur(10px);
		border-radius: 1rem;
		padding: 2rem;
		color: white;
	}

	.help-section h3 {
		margin: 0 0 1.5rem 0;
		font-size: 1.25rem;
		font-weight: 600;
		text-align: center;
	}

	.help-items {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.help-item {
		text-align: left;
	}

	.help-item strong {
		display: block;
		margin-bottom: 0.5rem;
		font-weight: 600;
	}

	.help-item p {
		margin: 0;
		opacity: 0.9;
		font-size: 0.875rem;
		line-height: 1.5;
	}

	@media (max-width: 640px) {
		.verification-required-page {
			padding: 1rem;
		}
		
		.verification-card {
			padding: 2rem;
		}
		
		.actions {
			flex-direction: column;
		}
		
		.btn {
			width: 100%;
		}
	}
</style>
