<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { showToast } from '$lib/toast';
	import { storeAuthData, apiRequest } from '$lib/auth';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let password = '';
	let confirmPassword = '';
	let loading = false;
	let error = '';
	let success = false;
	let token = '';
	let userId = '';
	let userEmail = '';

	// Get token and user ID from URL params
	onMount(() => {
		const urlParams = $page.url.searchParams;
		token = urlParams.get('token') || '';
		userId = urlParams.get('user_id') || '';
		
		if (!token) {
			error = 'Invalid setup link. Please request a new verification email.';
			return;
		}

		// Extract email from token if possible (for display purposes)
		// This is just for UX - the actual validation happens on the backend
		console.log('🔐 Password setup page loaded with token:', token.substring(0, 8) + '...');
	});

	async function handlePasswordSetup() {
		if (!password || !confirmPassword) {
			error = 'Please fill in all fields';
			return;
		}

		if (password !== confirmPassword) {
			error = 'Passwords do not match';
			return;
		}

		if (password.length < 8) {
			error = 'Password must be at least 8 characters long';
			return;
		}

		// Additional password validation
		const hasUpperCase = /[A-Z]/.test(password);
		const hasLowerCase = /[a-z]/.test(password);
		const hasNumbers = /\d/.test(password);
		const hasSpecialChar = /[!@#$%^&*(),.?":{}|<>]/.test(password);

		if (!hasUpperCase || !hasLowerCase || !hasNumbers || !hasSpecialChar) {
			error = 'Password must contain at least one uppercase letter, one lowercase letter, one number, and one special character';
			return;
		}

		loading = true;
		error = '';

		try {
			console.log('🔐 Setting up password...');
			
			const response = await apiRequest('/auth/setup-password', {
				method: 'POST',
				body: JSON.stringify({
					token: token,
					password: password,
					user_id: userId ? parseInt(userId) : undefined
				})
			});

			console.log('🔍 Response status:', response.status);
			console.log('🔍 Response headers:', Object.fromEntries(response.headers.entries()));

			let data;
			const contentType = response.headers.get('content-type');
			
			if (contentType && contentType.includes('application/json')) {
				data = await response.json();
			} else {
				// Handle HTML/text responses (like nginx error pages)
				const textResponse = await response.text();
				console.log('🔍 Non-JSON response:', textResponse.substring(0, 200) + '...');
				
				if (response.status === 504) {
					data = { 
						error: 'Gateway timeout - the server took too long to respond. Please try again in a moment.',
						debug_info: 'Received HTML response instead of JSON, likely nginx timeout'
					};
				} else {
					data = { 
						error: `Server error (${response.status}). Please try again.`,
						debug_info: `Received non-JSON response: ${textResponse.substring(0, 100)}...`
					};
				}
			}

			if (response.ok) {
				console.log('✅ Password setup successful:', data);
				success = true;
				
				// Store auth data if tokens were returned
				if (data.access_token && data.refresh_token) {
					const tokens = {
						access_token: data.access_token,
						refresh_token: data.refresh_token,
						expires_in: data.expires_in || 3600,
						token_type: data.token_type || 'Bearer'
					};
					
					storeAuthData(tokens, data.user);
					console.log('✅ Auto-login successful after password setup');
					
					showToast('Password setup successful! Welcome to BOME!', 'success');
					
					// Redirect to videos after a short delay
					setTimeout(() => {
						goto('/videos');
					}, 2000);
				} else {
					// Fallback if no tokens returned
					showToast('Password setup successful! Please login with your new password.', 'success');
					setTimeout(() => {
						goto('/login');
					}, 2000);
				}
			} else {
				error = data.error || 'Failed to setup password';
				console.error('❌ Password setup failed:', data);
			}
		} catch (err) {
			console.error('❌ Network error during password setup:', err);
			error = 'Network error. Please try again.';
		}

		loading = false;
	}

	function goToLogin() {
		goto('/login');
	}

	function goToHome() {
		goto('/');
	}
</script>

<svelte:head>
	<title>Setup Password - BOME</title>
	<meta name="description" content="Set up your password to complete your account setup" />
</svelte:head>

<div class="setup-password-page">
	<div class="container">
		{#if success}
			<!-- Success State -->
			<div class="success-container">
				<div class="success-icon">🎉</div>
				<h1>Password Setup Complete!</h1>
				<p class="success-message">
					Your password has been set successfully and you're now logged in!
				</p>
				
				<div class="success-info">
					<p>You can now:</p>
					<ul>
						<li>Access all premium features</li>
						<li>Manage your subscription</li>
						<li>Enjoy unlimited video streaming</li>
						<li>Receive important notifications</li>
					</ul>
				</div>

				<div class="success-actions">
					<button class="btn btn-primary" on:click={() => goto('/dashboard')}>
						Go to Dashboard
					</button>
					<button class="btn btn-outline" on:click={goToHome}>
						Back to Home
					</button>
				</div>
				
				<p class="redirect-notice">You'll be redirected to your dashboard in a few seconds...</p>
			</div>
		{:else if error && !token}
			<!-- Invalid Token State -->
			<div class="error-container">
				<div class="error-icon">❌</div>
				<h1>Invalid Setup Link</h1>
				<p class="error-message">{error}</p>
				
				<div class="error-help">
					<h3>What can you do?</h3>
					<ul>
						<li>Check if you clicked the correct link from your email</li>
						<li>Make sure the setup link hasn't expired (24 hours)</li>
						<li>Try logging in - you may already have a password</li>
						<li>Contact support if the problem persists</li>
					</ul>
				</div>

				<div class="error-actions">
					<button class="btn btn-primary" on:click={goToLogin}>
						Try Login
					</button>
					<button class="btn btn-outline" on:click={goToHome}>
						Back to Home
					</button>
				</div>
			</div>
		{:else}
			<!-- Password Setup Form -->
			<div class="setup-container">
				<div class="setup-header">
					<div class="setup-icon">🔐</div>
					<h1>Set Up Your Password</h1>
					<p>Welcome! To complete your account setup, please create a secure password.</p>
				</div>

				{#if error}
					<div class="error-message">
						{error}
					</div>
				{/if}

				<form on:submit|preventDefault={handlePasswordSetup} class="setup-form">
					<div class="form-group">
						<label for="password">New Password</label>
						<input
							id="password"
							type="password"
							bind:value={password}
							placeholder="Enter your new password"
							disabled={loading}
							required
						/>
						<div class="password-requirements">
							<p>Password must contain:</p>
							<ul>
								<li class:valid={password.length >= 8}>At least 8 characters</li>
								<li class:valid={/[A-Z]/.test(password)}>One uppercase letter</li>
								<li class:valid={/[a-z]/.test(password)}>One lowercase letter</li>
								<li class:valid={/\d/.test(password)}>One number</li>
								<li class:valid={/[!@#$%^&*(),.?":{}|<>]/.test(password)}>One special character</li>
							</ul>
						</div>
					</div>

					<div class="form-group">
						<label for="confirm-password">Confirm Password</label>
						<input
							id="confirm-password"
							type="password"
							bind:value={confirmPassword}
							placeholder="Confirm your new password"
							disabled={loading}
							required
						/>
						{#if confirmPassword && password !== confirmPassword}
							<div class="field-error">Passwords do not match</div>
						{/if}
					</div>

					<button type="submit" class="btn btn-primary btn-full" disabled={loading}>
						{#if loading}
							<LoadingSpinner size="small" />
							Setting up password...
						{:else}
							Set Up Password
						{/if}
					</button>
				</form>

				<div class="setup-footer">
					<p>Already have a password? <button class="link-btn" on:click={goToLogin}>Sign in here</button></p>
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	.setup-password-page {
		min-height: 100vh;
		background: linear-gradient(150deg, var(--primary-gold) 1%, var(--primary-bom-dark) 60%);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 2rem;
	}

	.container {
		max-width: 500px;
		margin: 0 auto;
		padding: 0 1rem;
	}

	.setup-container,
	.success-container,
	.error-container {
		background: white;
		border-radius: 16px;
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
		padding: 3rem 2rem;
		text-align: center;
		width: 100%;
		max-width: 600px;
	}


	.setup-header {
		margin-bottom: 2rem;
	}

	.setup-icon,
	.success-icon,
	.error-icon {
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

	.setup-container h1,
	.success-container h1,
	.error-container h1 {
		color: #1f2937;
		font-size: 2rem;
		font-weight: 700;
		margin-bottom: 1rem;
	}

	.setup-header p,
	.success-message,
	.error-message {
		color: #6b7280;
		font-size: 1.1rem;
		line-height: 1.6;
		margin-bottom: 1rem;
	}

	.setup-form {
		text-align: left;
		margin-bottom: 2rem;
	}

	.form-group {
		margin-bottom: 1.5rem;
	}

	.form-group label {
		display: block;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.form-group input {
		width: 100%;
		padding: 0.875rem 1rem;
		border: 2px solid var(--border-color);
		border-radius: 0.5rem;
		font-size: 1rem;
		background: var(--input-bg);
		color: var(--text-primary);
		transition: all 0.2s ease;
	}

	.form-group input:focus {
		outline: none;
		border-color: var(--primary-color);
		box-shadow: 0 0 0 3px var(--primary-color-alpha);
	}

	.form-group input:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.password-requirements {
		margin-top: 0.75rem;
		padding: 1rem;
		background: var(--bg-secondary);
		border-radius: 0.5rem;
		border: 1px solid var(--border-color);
	}

	.password-requirements p {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.password-requirements ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.password-requirements li {
		display: flex;
		align-items: center;
		font-size: 0.875rem;
		color: var(--text-secondary);
		margin-bottom: 0.25rem;
	}

	.password-requirements li::before {
		content: '✗';
		color: var(--error-color);
		font-weight: bold;
		margin-right: 0.5rem;
		width: 1rem;
	}

	.password-requirements li.valid::before {
		content: '✓';
		color: var(--success-color);
	}

	.password-requirements li.valid {
		color: var(--success-color);
	}

	.field-error {
		color: var(--error-color);
		font-size: 0.875rem;
		margin-top: 0.5rem;
	}

	.error-message {
		background: var(--error-bg);
		color: var(--error-color);
		padding: 1rem;
		border-radius: 0.5rem;
		border: 1px solid var(--error-color);
		margin-bottom: 1.5rem;
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

	.btn {
		padding: 0.875rem 1.5rem;
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
	}

	.btn-primary {
		background: var(--primary-color);
		color: var(--text-primary);
	}

	.btn-primary:hover:not(:disabled) {
		background: var(--primary-color-hover);
		transform: translateY(-1px);
	}

	.btn-primary:disabled {
		opacity: 0.6;
		cursor: not-allowed;
		transform: none;
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

	.btn-full {
		width: 100%;
	}

	.success-actions,
	.error-actions {
		display: flex;
		justify-content: center;
		gap: 1rem;
		flex-wrap: wrap;
		margin-top: 2rem;
	}

	.setup-footer {
		text-align: center;
		padding-top: 1.5rem;
		border-top: 1px solid var(--border-color);
	}

	.setup-footer p {
		color: var(--text-secondary);
		font-size: 0.875rem;
	}

	.link-btn {
		background: none;
		border: none;
		color: var(--primary-color);
		text-decoration: underline;
		cursor: pointer;
		font-size: inherit;
	}

	.link-btn:hover {
		color: var(--primary-color-hover);
	}

	.redirect-notice {
		margin-top: 2rem;
		font-size: 0.875rem;
		color: var(--text-secondary);
		font-style: italic;
	}

	@media (max-width: 768px) {
		.setup-container,
		.success-container,
		.error-container {
			padding: 2rem 1.5rem;
		}

		.setup-container h1,
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
