<script lang="ts">
	import { auth } from '$lib/auth';
	import Navigation from '$lib/components/Navigation.svelte';

	let email = '';
	let loading = false;
	let error = '';
	let success = false;

	async function handleForgotPassword() {
		if (!email) {
			error = 'Please enter your email address';
			return;
		}

		// Basic email validation
		const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
		if (!emailRegex.test(email)) {
			error = 'Please enter a valid email address';
			return;
		}

		loading = true;
		error = '';

		const result = await auth.forgotPassword(email);
		
		if (result.success) {
			success = true;
		} else {
			error = result.error || 'Failed to send reset email. Please try again.';
		}

		loading = false;
	}
</script>

<svelte:head>
	<title>Forgot Password - Book of Mormon Evidences</title>
</svelte:head>

<Navigation />

<div class="auth-container">
	<div class="auth-card">
		{#if !success}
			<div class="auth-header">
				<div class="icon-container">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="12" cy="12" r="10"/>
						<path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/>
						<line x1="12" y1="17" x2="12.01" y2="17"/>
					</svg>
				</div>
				<h1>Forgot Password?</h1>
				<p>Enter your email address and we'll send you a link to reset your password.</p>
			</div>

			<form on:submit|preventDefault={handleForgotPassword} class="auth-form">
				<div class="form-group">
					<label for="email">Email Address</label>
					<input
						type="email"
						id="email"
						bind:value={email}
						placeholder="Enter your email"
						required
						autocomplete="email"
						disabled={loading}
					/>
				</div>

				{#if error}
					<div class="error-message">
						{error}
					</div>
				{/if}

				<button type="submit" class="btn-primary" disabled={loading}>
					{loading ? 'Sending...' : 'Send Reset Link'}
				</button>
			</form>

			<div class="auth-footer">
				<p>
					Remember your password?
					<a href="/auth/login" class="link">Back to Login</a>
				</p>
			</div>
		{:else}
			<!-- Success Message -->
			<div class="success-container">
				<div class="success-icon">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
						<polyline points="22 4 12 14.01 9 11.01"/>
					</svg>
				</div>
				<h1>Check Your Email</h1>
				<p class="success-text">
					If an account exists with <strong>{email}</strong>, you'll receive a password reset link shortly.
				</p>
				<p class="info-text">
					The link will expire in 1 hour for security reasons.
				</p>
				<div class="success-actions">
					<a href="/auth/login" class="btn-primary">Back to Login</a>
					<button 
						class="btn-ghost" 
						on:click={() => { success = false; email = ''; }}
					>
						Send Another Link
					</button>
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
		min-height: 550px;
		padding: 2.5rem;
		background: var(--card-bg);
		border-radius: 18px;
		background: linear-gradient(145deg, var(--primary-gold), var(--primary-gold-dark));
		box-shadow: 20px 20px 60px var(--bg-glass-dark),
					-20px -20px 60px var(--bg-glass);
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
		line-height: 1.5;
		max-width: 400px;
		margin: 0 auto;
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

	.form-group input:disabled {
		opacity: 0.6;
		cursor: not-allowed;
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
		background: var(--success-bg, #4ade80);
		border-radius: 50%;
		box-shadow: 
			4px 4px 12px var(--shadow-dark),
			-2px -2px 8px var(--shadow-light);
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

	.success-text strong {
		color: var(--accent-color);
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

