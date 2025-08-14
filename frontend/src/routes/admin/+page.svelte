<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import { showToast } from '$lib/toast';

	let email = '';
	let password = '';
	let isLoading = false;
	let error = '';

	onMount(() => {
		// Redirect if already authenticated as admin
		if ($auth.isAuthenticated && isAdminUser($auth.user)) {
			goto('/admin/dashboard');
		}
	});

	function isAdminUser(user: any): boolean {
		if (!user) return false;
		const adminRoles = [
			'super_admin', 'system_admin', 'content_manager', 
			'articles_manager', 'youtube_manager', 'streaming_manager',
			'events_manager', 'advertisement_manager', 'user_manager',
			'analytics_manager', 'financial_admin', 'admin'
		];
		return adminRoles.includes(user.role);
	}

	async function handleLogin() {
		if (!email || !password) {
			error = 'Please enter both email and password';
			return;
		}

		isLoading = true;
		error = '';

		try {
			const result = await auth.login(email, password);
			
			if (result.success && result.user) {
				// Check if user has admin role using the user data from the login result
				if (isAdminUser(result.user)) {
					showToast('Welcome to the admin panel!', 'success');
					goto('/admin/dashboard');
				} else {
					await auth.logout();
					error = 'Access denied. Admin privileges required.';
					showToast('Access denied. Admin privileges required.', 'error');
				}
			} else {
				error = result.error || 'Login failed. Please try again.';
				showToast(error, 'error');
			}
		} catch (err: any) {
			error = err.message || 'Login failed. Please try again.';
			showToast(error, 'error');
		} finally {
			isLoading = false;
		}
	}

	function handleKeyPress(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			handleLogin();
		}
	}
</script>

<svelte:head>
	<title>Admin Login - BOME</title>
	<meta name="description" content="Administrative access to BOME platform" />
</svelte:head>

<div class="admin-login-page">
	<div class="login-container">
		<div class="login-header">
			<div class="brand-logo">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M12 2L2 7l10 5 10-5-10-5z"></path>
					<path d="M2 17l10 5 10-5"></path>
					<path d="M2 12l10 5 10-5"></path>
				</svg>
			</div>
			<h1>BOME Admin</h1>
			<p>Administrative Access Portal</p>
		</div>

		<form class="login-form" on:submit|preventDefault={handleLogin}>
			{#if error}
				<div class="error-message">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="12" cy="12" r="10"></circle>
						<line x1="15" y1="9" x2="9" y2="15"></line>
						<line x1="9" y1="9" x2="15" y2="15"></line>
					</svg>
					{error}
				</div>
			{/if}

			<div class="form-group">
				<label for="email">Email Address</label>
				<input
					id="email"
					type="email"
					bind:value={email}
					placeholder="admin@bome.com"
					required
					disabled={isLoading}
					on:keypress={handleKeyPress}
				/>
			</div>

			<div class="form-group">
				<label for="password">Password</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					placeholder="Enter your password"
					required
					disabled={isLoading}
					on:keypress={handleKeyPress}
				/>
			</div>

			<button type="submit" class="login-button" disabled={isLoading}>
				{#if isLoading}
					<div class="spinner"></div>
					Signing In...
				{:else}
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"></path>
						<polyline points="10,17 15,12 10,7"></polyline>
						<line x1="15" y1="12" x2="3" y2="12"></line>
					</svg>
					Sign In
				{/if}
			</button>
		</form>

		<div class="login-footer">
			<a href="/" class="back-link">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<line x1="19" y1="12" x2="5" y2="12"></line>
					<polyline points="12,19 5,12 12,5"></polyline>
				</svg>
				Back to Main Site
			</a>
		</div>
	</div>
</div>

<style>
	.admin-login-page {
		min-height: 100vh;
		background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 2rem;
		position: relative;
		overflow: hidden;
	}

	.admin-login-page::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="20" cy="20" r="1" fill="white" opacity="0.1"/><circle cx="80" cy="30" r="0.5" fill="white" opacity="0.1"/><circle cx="40" cy="60" r="1.5" fill="white" opacity="0.1"/></svg>');
		opacity: 0.3;
	}

	.login-container {
		background: rgba(255, 255, 255, 0.1);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		border-radius: 20px;
		padding: 3rem;
		width: 100%;
		max-width: 400px;
		border: 1px solid rgba(255, 255, 255, 0.2);
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
		position: relative;
		z-index: 1;
	}

	.login-header {
		text-align: center;
		margin-bottom: 2rem;
	}

	.brand-logo {
		width: 60px;
		height: 60px;
		background: var(--primary-gradient, linear-gradient(135deg, #667eea 0%, #764ba2 100%));
		border-radius: 15px;
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 0 auto 1rem;
		box-shadow: 0 10px 20px rgba(0, 0, 0, 0.2);
	}

	.brand-logo svg {
		width: 30px;
		height: 30px;
		color: white;
	}

	.login-header h1 {
		color: white;
		font-size: 2rem;
		font-weight: 700;
		margin: 0 0 0.5rem 0;
		font-family: var(--font-display, 'Inter', sans-serif);
	}

	.login-header p {
		color: rgba(255, 255, 255, 0.8);
		margin: 0;
		font-size: 1rem;
	}

	.login-form {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.error-message {
		background: rgba(239, 68, 68, 0.1);
		border: 1px solid rgba(239, 68, 68, 0.3);
		border-radius: 10px;
		padding: 1rem;
		color: #fca5a5;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.9rem;
	}

	.error-message svg {
		width: 16px;
		height: 16px;
		flex-shrink: 0;
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.form-group label {
		color: rgba(255, 255, 255, 0.9);
		font-size: 0.9rem;
		font-weight: 500;
	}

	.form-group input {
		background: rgba(255, 255, 255, 0.1);
		border: 1px solid rgba(255, 255, 255, 0.2);
		border-radius: 10px;
		padding: 0.75rem 1rem;
		color: white;
		font-size: 1rem;
		transition: all 0.3s ease;
	}

	.form-group input::placeholder {
		color: rgba(255, 255, 255, 0.5);
	}

	.form-group input:focus {
		outline: none;
		border-color: var(--primary, #667eea);
		background: rgba(255, 255, 255, 0.15);
		box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
	}

	.form-group input:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.login-button {
		background: var(--primary-gradient, linear-gradient(135deg, #667eea 0%, #764ba2 100%));
		border: none;
		border-radius: 10px;
		padding: 0.75rem 1rem;
		color: white;
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.3s ease;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		box-shadow: 0 10px 20px rgba(0, 0, 0, 0.2);
	}

	.login-button:hover:not(:disabled) {
		transform: translateY(-2px);
		box-shadow: 0 15px 30px rgba(0, 0, 0, 0.3);
	}

	.login-button:disabled {
		opacity: 0.7;
		cursor: not-allowed;
		transform: none;
	}

	.login-button svg {
		width: 18px;
		height: 18px;
	}

	.spinner {
		width: 18px;
		height: 18px;
		border: 2px solid rgba(255, 255, 255, 0.3);
		border-top: 2px solid white;
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.login-footer {
		margin-top: 2rem;
		text-align: center;
	}

	.back-link {
		color: rgba(255, 255, 255, 0.7);
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.9rem;
		transition: color 0.3s ease;
	}

	.back-link:hover {
		color: white;
	}

	.back-link svg {
		width: 16px;
		height: 16px;
	}

	@media (max-width: 480px) {
		.login-container {
			padding: 2rem;
		}

		.login-header h1 {
			font-size: 1.75rem;
		}
	}
</style> 
