<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';

	// Get URL parameters
	let userEmail = '';
	let userId = '';
	let success = false;
	let error = '';
	let isResending = false;
	let resendSuccess = false;
	let resendCooldown = 0;
	let autoRedirectTimer = 20; // 20 second countdown
	let showManualRedirect = false;

	onMount(async () => {
		// Get parameters from URL
		const urlParams = $page.url.searchParams;
		userEmail = urlParams.get('email') || '';
		userId = urlParams.get('user_id') || '';
		success = urlParams.get('success') === 'true';
		error = urlParams.get('error') || '';

		// Check for auto-login tokens
		const accessToken = urlParams.get('access_token');
		const refreshToken = urlParams.get('refresh_token');
		const userIdFromToken = urlParams.get('user_id');

		if (success && accessToken && refreshToken && userIdFromToken) {
			// Auto-login the user after successful verification
			try {
				console.log('🔄 Attempting auto-login with tokens...');
				
				// Get user data from backend
				const userResponse = await fetch(`http://localhost:8080/api/v1/users/me`, {
					headers: {
						'Authorization': `Bearer ${accessToken}`,
						'Content-Type': 'application/json'
					}
				});

				if (userResponse.ok) {
					const userData = await userResponse.json();
					console.log('✅ User data fetched successfully:', userData);
					
					// Store auth data
					const tokens = {
						access_token: accessToken,
						refresh_token: refreshToken,
						expires_in: 3600, // 1 hour default
						token_type: 'Bearer'
					};

					// Import storeAuthData function
					const { storeAuthData } = await import('$lib/auth');
					storeAuthData(tokens, userData);
					console.log('✅ Auth data stored successfully');

					// Start countdown timer for redirect
					startAutoRedirectTimer();
					return;
				} else {
					console.error('❌ Failed to fetch user data:', userResponse.status);
				}
			} catch (err) {
				console.error('❌ Auto-login failed:', err);
				// Continue with normal success flow and show manual redirect
				showManualRedirect = true;
			}
		}

		// If we reach here and it's a success, start the timer anyway
		if (success) {
			// If no auto-login tokens, show manual redirect immediately
			if (!accessToken || !refreshToken) {
				showManualRedirect = true;
				autoRedirectTimer = 10; // Shorter timer for manual redirect
			}
			startAutoRedirectTimer();
		}

		// If no email provided, redirect to login
		if (!userEmail && !success && !error) {
			goto('/auth/login');
			return;
		}

		// Start cooldown timer if there was a recent resend
		const lastResend = localStorage.getItem('last_verification_resend');
		if (lastResend) {
			const timeSince = Date.now() - parseInt(lastResend);
			const cooldownTime = 60000; // 1 minute cooldown
			if (timeSince < cooldownTime) {
				resendCooldown = Math.ceil((cooldownTime - timeSince) / 1000);
				startCooldownTimer();
			}
		}
	});

	function startCooldownTimer() {
		if (resendCooldown > 0) {
			const timer = setInterval(() => {
				resendCooldown--;
				if (resendCooldown <= 0) {
					clearInterval(timer);
				}
			}, 1000);
		}
	}

	function startAutoRedirectTimer() {
		console.log('🕐 Starting auto-redirect timer (20 seconds)');
		const timer = setInterval(() => {
			autoRedirectTimer--;
			console.log(`⏰ Auto-redirect in ${autoRedirectTimer} seconds`);
			
			if (autoRedirectTimer <= 0) {
				clearInterval(timer);
				console.log('🚀 Auto-redirecting to home page');
				goto('/');
			} else if (autoRedirectTimer <= 15) {
				// Show manual redirect button after 5 seconds
				showManualRedirect = true;
			}
		}, 1000);
	}

	function goToHomepage() {
		console.log('🏠 Manual redirect to home page');
		goto('/');
	}

	async function resendVerification() {
		if (isResending || resendCooldown > 0) return;

		isResending = true;
		resendSuccess = false;
		error = '';

		try {
			const result = await auth.requestVerification(userEmail);

			if (result.success) {
				resendSuccess = true;
				// Set cooldown
				localStorage.setItem('last_verification_resend', Date.now().toString());
				resendCooldown = 60;
				startCooldownTimer();
			} else {
				error = result.error || 'Failed to resend verification email';
			}
		} catch (err) {
			error = 'Network error. Please try again.';
			console.error('Resend verification error:', err);
		} finally {
			isResending = false;
		}
	}

	function getErrorMessage(errorCode: string): string {
		switch (errorCode) {
			case 'invalid_link':
				return 'The verification link is invalid or malformed.';
			case 'invalid_token':
				return 'The verification token is invalid or has expired.';
			case 'verification_failed':
				return 'Verification failed due to a server error.';
			case 'service_unavailable':
				return 'The verification service is temporarily unavailable.';
			default:
				return 'An unknown error occurred during verification.';
		}
	}
</script>

<svelte:head>
	<title>Email Verification - BOME</title>
	<meta name="description" content="Verify your email address to complete your BOME account setup" />
</svelte:head>

<div class="verify-email-page">
	<div class="container">
		<div class="verification-card">
			{#if success}
				<!-- Success State -->
				<div class="success-state">
					<div class="icon success-icon">✅</div>
					<h1>Email Verified Successfully!</h1>
					<p>Welcome to BOME! Your email has been verified and you're being logged in automatically...</p>
					
					<div class="loading-spinner spinner">
					</div>
					
					<div class="redirect-info">
						{#if $page.url.searchParams.get('access_token')}
							<p>Redirecting to your dashboard in <strong>{autoRedirectTimer}</strong> seconds...</p>
						{:else}
							<p>Redirecting to homepage in <strong>{autoRedirectTimer}</strong> seconds...</p>
						{/if}
						
						{#if showManualRedirect}
							<div class="manual-redirect">
								<p>Taking too long?</p>
								<button class="btn btn-primary" on:click={goToHomepage}>
									🏠 Go to Homepage
								</button>
							</div>
						{/if}
					</div>
				</div>
			{:else if error}
				<!-- Error State -->
				<div class="error-state">
					<div class="icon error-icon">❌</div>
					<h1>Verification Failed</h1>
					<p class="error-message">{getErrorMessage(error)}</p>
					<div class="actions">
						<button 
							class="btn btn-primary" 
							on:click={resendVerification}
							disabled={isResending || resendCooldown > 0}
						>
							{#if isResending}
								Sending...
							{:else if resendCooldown > 0}
								Resend in {resendCooldown}s
							{:else}
								Resend Verification Email
							{/if}
						</button>
						<a href="/login" class="btn btn-secondary">Back to Login</a>
					</div>
				</div>
			{:else}
				<!-- Verification Required State -->
				<div class="verification-required">
					<div class="icon email-icon">📧</div>
					<h1>Email Verification Required</h1>
					<p class="welcome-message">
						Welcome to BOME! To keep your account secure, please verify your email address before logging in.
					</p>
					
					{#if userEmail}
						<div class="email-info">
							<p>We sent a verification link to:</p>
							<div class="email-address">{userEmail}</div>
						</div>
					{/if}

					<div class="instructions">
						<h3>What to do next:</h3>
						<ol>
							<li>Check your email inbox for a message from BOME</li>
							<li>Click the "Verify Email Address" button in the email</li>
							<li>You'll be redirected back here with confirmation</li>
							<li>Then you can log in to your account</li>
						</ol>
					</div>

					{#if resendSuccess}
						<div class="success-message">
							✅ Verification email sent successfully! Please check your inbox.
						</div>
					{/if}

					{#if error}
						<div class="error-message">
							❌ {error}
						</div>
					{/if}

					<div class="actions">
						<button 
							class="btn btn-primary" 
							on:click={resendVerification}
							disabled={isResending || resendCooldown > 0}
						>
							{#if isResending}
								<span class="loading-spinner"></span>
								Sending...
							{:else if resendCooldown > 0}
								Resend in {resendCooldown}s
							{:else}
								Resend Verification Email
							{/if}
						</button>
						<a href="/auth/login" class="btn btn-secondary">Back to Login</a>
					</div>

					<div class="help-section">
						<h4>Need help?</h4>
						<ul>
							<li>Check your spam or junk folder</li>
							<li>Make sure {userEmail} is correct</li>
							<li>Wait a few minutes for the email to arrive</li>
							<li>Contact <a href="mailto:support@bookofmormonevidence.org">support@bookofmormonevidence.org</a> if you continue having issues</li>
						</ul>
					</div>
				</div>
			{/if}
		</div>
	</div>
</div>

<style>
	.verify-email-page {
		min-height: 100vh;
		background: linear-gradient(150deg, var(--primary-gold) 1%, var(--primary-bom-dark) 60%);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 2rem;
	}

	.container {
		width: 100%;
		max-width: 600px;
	}

	.verification-card {
		background: white;
		border-radius: 16px;
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
		padding: 3rem;
		text-align: center;
	}

	.icon {
		font-size: 4rem;
		margin-bottom: 1.5rem;
		display: block;
	}

	.success-icon {
		color: #10b981;
	}

	.error-icon {
		color: #ef4444;
	}

	.email-icon {
		color: #3b82f6;
	}

	h1 {
		color: #1f2937;
		font-size: 2rem;
		font-weight: 700;
		margin-bottom: 1rem;
	}

	.welcome-message {
		color: #6b7280;
		font-size: 1.1rem;
		line-height: 1.6;
		margin-bottom: 2rem;
	}

	.email-info {
		background: #f3f4f6;
		border-radius: 12px;
		padding: 1.5rem;
		margin: 2rem 0;
	}

	.email-info p {
		color: #6b7280;
		margin-bottom: 0.5rem;
		font-size: 0.9rem;
	}

	.email-address {
		color: #1f2937;
		font-weight: 600;
		font-size: 1.1rem;
		word-break: break-all;
	}

	.instructions {
		text-align: left;
		background: #fef3c7;
		border-radius: 12px;
		padding: 1.5rem;
		margin: 2rem 0;
	}

	.instructions h3 {
		color: #92400e;
		font-size: 1.1rem;
		margin-bottom: 1rem;
	}

	.instructions ol {
		color: #78350f;
		padding-left: 1.5rem;
	}

	.instructions li {
		margin-bottom: 0.5rem;
		line-height: 1.5;
	}

	.success-message {
		background: #d1fae5;
		color: #065f46;
		padding: 1rem;
		border-radius: 8px;
		margin: 1rem 0;
		font-weight: 500;
	}

	.error-message {
		background: #fee2e2;
		color: #dc2626;
		padding: 1rem;
		border-radius: 8px;
		margin: 1rem 0;
		font-weight: 500;
	}

	.actions {
		display: flex;
		gap: 1rem;
		justify-content: center;
		margin: 2rem 0;
		flex-wrap: wrap;
	}

	.btn {
		padding: 0.75rem 1.5rem;
		border-radius: 8px;
		font-weight: 600;
		text-decoration: none;
		border: none;
		cursor: pointer;
		transition: all 0.2s;
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		min-width: 140px;
		justify-content: center;
	}

	.btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	/* Loading spinner styles */
	.loading-spinner {
		text-align: center;
		margin: 2rem 0;
	}

	.spinner {
		border: 3px solid #f3f3f3;
		border-top: 3px solid #667eea;
		border-radius: 50%;
		width: 40px;
		height: 40px;
		animation: spin 1s linear infinite;
		margin: 0 auto 1rem;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.loading-spinner p {
		color: #6b7280;
		font-size: 0.9rem;
	}

	/* Redirect info styles */
	.redirect-info {
		margin-top: 1.5rem;
		text-align: center;
	}

	.redirect-info p {
		color: #4b5563;
		font-size: 1rem;
		margin-bottom: 1rem;
	}

	.redirect-info strong {
		color: #667eea;
		font-weight: 600;
	}

	.manual-redirect {
		margin-top: 1.5rem;
		padding: 1rem;
		background: #f8fafc;
		border-radius: 8px;
		border: 1px solid #e2e8f0;
	}

	.manual-redirect p {
		color: #64748b;
		font-size: 0.9rem;
		margin-bottom: 0.75rem;
	}

	.manual-redirect .btn {
		font-size: 0.95rem;
		padding: 0.75rem 1.5rem;
	}

	.btn-primary {
		background: #3b82f6;
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: #2563eb;
		transform: translateY(-1px);
	}

	.btn-secondary {
		background: #f3f4f6;
		color: #6b7280;
	}

	.btn-secondary:hover {
		background: #e5e7eb;
		color: #374151;
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
		to {
			transform: rotate(360deg);
		}
	}

	.help-section {
		text-align: left;
		background: #f9fafb;
		border-radius: 12px;
		padding: 1.5rem;
		margin-top: 2rem;
	}

	.help-section h4 {
		color: #374151;
		font-size: 1rem;
		margin-bottom: 1rem;
	}

	.help-section ul {
		color: #6b7280;
		padding-left: 1.5rem;
	}

	.help-section li {
		margin-bottom: 0.5rem;
		line-height: 1.5;
	}

	.help-section a {
		color: #3b82f6;
		text-decoration: none;
	}

	.help-section a:hover {
		text-decoration: underline;
	}

	/* Mobile responsiveness */
	@media (max-width: 640px) {
		.verify-email-page {
			padding: 1rem;
		}

		.verification-card {
			padding: 2rem;
		}

		h1 {
			font-size: 1.5rem;
		}

		.actions {
			flex-direction: column;
		}

		.btn {
			width: 100%;
		}
	}
</style>
