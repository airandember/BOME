<script lang="ts">
	import { auth } from '$lib/auth';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import Navigation from '$lib/components/Navigation.svelte';

	let token = '';
	let password = '';
	let confirmPassword = '';
	let loading = false;
	let error = '';
	let success = false;
	let tokenValid = false;
	let checkingToken = true;

	// Password strength indicator
	let passwordStrength: 'weak' | 'medium' | 'strong' = 'weak';

	onMount(() => {
		// Get token from URL query params
		token = $page.url.searchParams.get('token') || '';
		
		if (!token) {
			error = 'Invalid or missing reset token. Please request a new password reset link.';
			checkingToken = false;
		} else {
			tokenValid = true;
			checkingToken = false;
		}
	});

	// Check password strength
	$: {
		if (password.length === 0) {
			passwordStrength = 'weak';
		} else if (password.length < 8) {
			passwordStrength = 'weak';
		} else if (password.length < 12 && /[A-Z]/.test(password) && /[0-9]/.test(password)) {
			passwordStrength = 'medium';
		} else if (password.length >= 12 && /[A-Z]/.test(password) && /[0-9]/.test(password) && /[!@#$%^&*]/.test(password)) {
			passwordStrength = 'strong';
		} else {
			passwordStrength = 'medium';
		}
	}

	// Check if passwords match
	$: passwordsMatch = password === confirmPassword && confirmPassword.length > 0;
	$: passwordsDontMatch = confirmPassword.length > 0 && password !== confirmPassword;

	async function handleResetPassword() {
		// Validation
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

		loading = true;
		error = '';

		const result = await auth.resetPassword(token, password);
		
		if (result.success) {
			success = true;
			// Redirect to login after 3 seconds
			setTimeout(() => {
				goto('/auth/login');
			}, 3000);
		} else {
			error = result.error || 'Failed to reset password. The link may have expired.';
		}

		loading = false;
	}
</script>

<svelte:head>
	<title>Reset Password - Book of Mormon Evidences</title>
</svelte:head>

<Navigation />

<div class="auth-container">
	<div class="auth-card">
		{#if checkingToken}
			<!-- Loading State -->
			<div class="loading-container">
				<div class="spinner"></div>
				<p>Verifying reset link...</p>
			</div>
		{:else if !tokenValid}
			<!-- Invalid Token -->
			<div class="error-container">
				<div class="error-icon">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="12" cy="12" r="10"/>
						<line x1="15" y1="9" x2="9" y2="15"/>
						<line x1="9" y1="9" x2="15" y2="15"/>
					</svg>
				</div>
				<h1>Invalid Reset Link</h1>
				<p class="error-text">
					{error}
				</p>
				<div class="error-actions">
					<a href="/auth/forgot-password" class="btn-primary">Request New Link</a>
					<a href="/auth/login" class="btn-ghost">Back to Login</a>
				</div>
			</div>
		{:else if !success}
			<!-- Reset Password Form -->
			<div class="auth-header">
				<div class="icon-container">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
						<path d="M7 11V7a5 5 0 0 1 10 0v4"/>
					</svg>
				</div>
				<h1>Reset Your Password</h1>
				<p>Choose a strong password for your account</p>
			</div>

			<form on:submit|preventDefault={handleResetPassword} class="auth-form">
				<div class="form-group">
					<label for="password">New Password</label>
					<input
						type="password"
						id="password"
						bind:value={password}
						placeholder="Enter new password"
						required
						autocomplete="new-password"
						disabled={loading}
					/>
					{#if password.length > 0}
						<div class="password-strength">
							<div class="strength-bar">
								<div 
									class="strength-fill strength-{passwordStrength}"
									style="width: {passwordStrength === 'weak' ? '33%' : passwordStrength === 'medium' ? '66%' : '100%'}"
								></div>
							</div>
							<span class="strength-label strength-{passwordStrength}">
								{passwordStrength === 'weak' ? 'Weak' : passwordStrength === 'medium' ? 'Medium' : 'Strong'}
							</span>
						</div>
					{/if}
					<p class="helper-text">
						Use at least 8 characters with a mix of letters, numbers, and symbols
					</p>
				</div>

				<div class="form-group">
					<label for="confirmPassword">Confirm New Password</label>
					<input
						type="password"
						id="confirmPassword"
						bind:value={confirmPassword}
						placeholder="Confirm new password"
						required
						autocomplete="new-password"
						disabled={loading}
						class:valid={passwordsMatch}
						class:invalid={passwordsDontMatch}
					/>
					{#if passwordsMatch}
						<p class="helper-text success">
							✓ Passwords match
						</p>
					{:else if passwordsDontMatch}
						<p class="helper-text error-text">
							✗ Passwords do not match
						</p>
					{/if}
				</div>

				{#if error}
					<div class="error-message">
						{error}
					</div>
				{/if}

				<button type="submit" class="btn-primary" disabled={loading || !passwordsMatch}>
					{loading ? 'Resetting Password...' : 'Reset Password'}
				</button>
			</form>

			<div class="auth-footer">
				<p>
					Remember your password?
					<a href="/auth/login" class="link">Back to Login</a>
				</p>
			</div>
		{:else}
			<!-- Success State -->
			<div class="success-container">
				<div class="success-icon">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
						<polyline points="22 4 12 14.01 9 11.01"/>
					</svg>
				</div>
				<h1>Password Reset Successful!</h1>
				<p class="success-text">
					Your password has been successfully reset.
				</p>
				<p class="info-text">
					Redirecting you to login in 3 seconds...
				</p>
				<div class="success-actions">
					<a href="/auth/login" class="btn-primary">Go to Login</a>
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	.auth-container {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 2rem;
		background: var(--bg-color);
	}

	.auth-card {
		width: 100%;
		max-width: 550px;
		min-height: 600px;
		padding: 2.5rem;
		background: var(--card-bg);
		border-radius: 18px;
		background: linear-gradient(145deg, var(--primary-gold), var(--primary-gold-dark));
		box-shadow: 20px 20px 60px var(--bg-glass-dark),
					-20px -20px 60px var(--bg-glass);
	}

	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 400px;
		gap: 1.5rem;
	}

	.spinner {
		width: 50px;
		height: 50px;
		border: 4px solid var(--bg-glass);
		border-top-color: var(--accent-color);
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	.loading-container p {
		color: var(--text-primary);
		font-size: 1rem;
	}

	.auth-header {
		text-align: center;
		margin-bottom: 2rem;
	}

	.icon-container {
		width: 80px;
		height: 80px;
		margin: 0 auto 1.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--bg-glass);
		border-radius: 50%;
		box-shadow: 
			inset 2px 2px 4px var(--shadow-dark),
			inset -2px -2px 4px var(--shadow-light);
	}

	.icon-container svg {
		width: 40px;
		height: 40px;
		color: var(--text-primary);
	}

	.auth-header h1 {
		font-size: 2rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.auth-header p {
		color: var(--text-inverse);
		font-size: 0.95rem;
	}

	.auth-form {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.form-group label {
		font-weight: 600;
		color: var(--text-primary);
		font-size: 0.9rem;
	}

	.form-group input {
		padding: 0.75rem 1rem;
		min-height: 60px;
		border: none;
		border-radius: 12px;
		background: var(--input-bg);
		color: var(--text-primary);
		font-size: 1rem;
		box-shadow: 
			inset 2px 2px 4px var(--shadow-dark),
			inset -2px -2px 4px var(--shadow-light);
		transition: all 0.2s ease;
	}

	.form-group input:focus {
		outline: none;
		box-shadow: 
			inset 2px 2px 4px var(--shadow-dark),
			inset -2px -2px 4px var(--shadow-light),
			0 0 0 2px var(--accent-color);
	}

	.form-group input.valid {
		box-shadow: 
			inset 2px 2px 4px var(--shadow-dark),
			inset -2px -2px 4px var(--shadow-light),
			0 0 0 2px #4ade80;
	}

	.form-group input.invalid {
		box-shadow: 
			inset 2px 2px 4px var(--shadow-dark),
			inset -2px -2px 4px var(--shadow-light),
			0 0 0 2px #ef4444;
	}

	.form-group input:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.helper-text {
		font-size: 0.85rem;
		color: var(--text-inverse);
		margin: 0;
	}

	.helper-text.success {
		color: #4ade80;
		font-weight: 600;
	}

	.helper-text.error-text {
		color: #ef4444;
		font-weight: 600;
	}

	/* Password Strength Indicator */
	.password-strength {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-top: 0.5rem;
	}

	.strength-bar {
		flex: 1;
		height: 6px;
		background: var(--bg-glass);
		border-radius: 3px;
		overflow: hidden;
	}

	.strength-fill {
		height: 100%;
		transition: width 0.3s ease, background-color 0.3s ease;
	}

	.strength-fill.strength-weak {
		background: #ef4444;
	}

	.strength-fill.strength-medium {
		background: #f59e0b;
	}

	.strength-fill.strength-strong {
		background: #4ade80;
	}

	.strength-label {
		font-size: 0.85rem;
		font-weight: 600;
		min-width: 60px;
	}

	.strength-label.strength-weak {
		color: #ef4444;
	}

	.strength-label.strength-medium {
		color: #f59e0b;
	}

	.strength-label.strength-strong {
		color: #4ade80;
	}

	.btn-primary {
		padding: 0.75rem 1.5rem;
		min-height: 60px;
		border: none;
		border-radius: 12px;
		background: var(--bg-glass-dark);
		color: white;
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		box-shadow: 
			4px 4px 8px var(--shadow-dark),
			-2px -2px 4px var(--shadow-light);
	}

	.btn-primary:hover:not(:disabled) {
		transform: translateY(-2px);
		box-shadow: 
			6px 6px 12px var(--shadow-dark),
			-3px -3px 6px var(--shadow-light);
		background: var(--secondary-gradient);
		color: var(--text-primary);
	}

	.btn-primary:active:not(:disabled) {
		transform: translateY(0);
		box-shadow: 
			2px 2px 4px var(--shadow-dark),
			-1px -1px 2px var(--shadow-light);
	}

	.btn-primary:disabled {
		opacity: 0.6;
		cursor: not-allowed;
		transform: none;
	}

	.btn-ghost {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 12px;
		background: transparent;
		color: var(--text-primary);
		font-size: 0.95rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		display: inline-block;
	}

	.btn-ghost:hover {
		background: var(--bg-glass);
		box-shadow: 
			2px 2px 4px var(--shadow-dark),
			-2px -2px 4px var(--shadow-light);
	}

	.error-message {
		padding: 0.75rem;
		background: var(--error-bg);
		color: var(--error-text);
		border-radius: 8px;
		font-size: 0.9rem;
		text-align: center;
	}

	.auth-footer {
		margin-top: 2rem;
		text-align: center;
	}

	.auth-footer p {
		margin: 0.5rem 0;
		color: var(--text-primary);
		font-size: 0.9rem;
	}

	.link {
		color: var(--accent-color);
		text-decoration: none;
		font-weight: 600;
		transition: color 0.2s ease;
	}

	.link:hover {
		color: var(--accent-hover);
	}

	/* Error State */
	.error-container {
		text-align: center;
		padding: 2rem 0;
	}

	.error-icon {
		width: 100px;
		height: 100px;
		margin: 0 auto 2rem;
		display: flex;
		align-items: center;
		justify-content: center;
		background: #ef4444;
		border-radius: 50%;
		box-shadow: 
			4px 4px 12px var(--shadow-dark),
			-2px -2px 8px var(--shadow-light);
	}

	.error-icon svg {
		width: 50px;
		height: 50px;
		color: white;
		stroke-width: 3;
	}

	.error-container h1 {
		font-size: 2rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.error-text {
		color: var(--text-primary);
		font-size: 1rem;
		line-height: 1.6;
		margin-bottom: 2rem;
	}

	.error-actions {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		margin-top: 2rem;
	}

	/* Success State */
	.success-container {
		text-align: center;
		padding: 2rem 0;
	}

	.success-icon {
		width: 100px;
		height: 100px;
		margin: 0 auto 2rem;
		display: flex;
		align-items: center;
		justify-content: center;
		background: #4ade80;
		border-radius: 50%;
		box-shadow: 
			4px 4px 12px var(--shadow-dark),
			-2px -2px 8px var(--shadow-light);
		animation: successPulse 2s ease-in-out infinite;
	}

	@keyframes successPulse {
		0%, 100% {
			transform: scale(1);
		}
		50% {
			transform: scale(1.05);
		}
	}

	.success-icon svg {
		width: 50px;
		height: 50px;
		color: white;
		stroke-width: 3;
	}

	.success-container h1 {
		font-size: 2rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.success-text {
		color: var(--text-primary);
		font-size: 1rem;
		line-height: 1.6;
		margin-bottom: 1rem;
	}

	.info-text {
		color: var(--text-inverse);
		font-size: 0.9rem;
		margin-bottom: 2rem;
	}

	.success-actions {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		margin-top: 2rem;
	}
</style>

