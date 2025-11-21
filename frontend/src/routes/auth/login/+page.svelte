<script lang="ts">
	import { auth, isAdmin } from '$lib/auth';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import Navigation from '$lib/components/Navigation.svelte';
	import OAuth2Login from '$lib/components/OAuth2Login.svelte';

	let email = '';
	let password = '';
	let loading = false;
	let error = '';
	let showGmailWarning = false;
	let isGmailAccount = false;
	
	// 🔗 CONTEXT PRESERVATION: Track return URL and plan context
	let returnUrl = '/';
	let planId = '';

	onMount(() => {
		// Check if already logged in, but don't redirect immediately
		auth.subscribe((state) => {
			if (state.isAuthenticated && !state.loading) {
				// Show a message instead of immediate redirect
				console.log('User already authenticated');
			}
		});
		
		// 🔗 Read return URL and plan context from query params
		if (typeof window !== 'undefined') {
			returnUrl = $page.url.searchParams.get('return') || '/';
			planId = $page.url.searchParams.get('plan_id') || '';
		}
	});

	// Check if email is a Gmail account
	function checkEmailProvider() {
		const emailLower = email.toLowerCase().trim();
		isGmailAccount = emailLower.endsWith('@gmail.com') || emailLower.endsWith('@googlemail.com');
		
		if (isGmailAccount) {
			showGmailWarning = true;
		} else {
			showGmailWarning = false;
		}
	}

	async function handleLogin() {
		if (!email || !password) {
			error = 'Please fill in all fields';
			return;
		}

		// Check if trying to login with Gmail
		checkEmailProvider();
		if (isGmailAccount) {
			error = 'Gmail accounts must use "Sign in with Google" above';
			return;
		}

		loading = true;
		error = '';

		const result = await auth.login(email, password);
		
		if (result.success) {
			// 🔗 CONTEXT PRESERVATION: If user was subscribing, redirect back to subscription
			if (planId && returnUrl === '/subscription') {
				goto(`/subscription?auto_checkout=true&plan_id=${planId}`);
			} else if (isAdmin()) {
				goto('/admin');
			} else {
				goto(returnUrl);
			}
		} else {
			error = result.error || 'Login failed';
		}

		loading = false;
	}
</script>

<svelte:head>
	<title>Login - Book of Mormon Evidences</title>
</svelte:head>
<Navigation />
<div class="auth-container">
	<div class="outerNew">
		<div class="dotNew"></div>
		<div class="cardNew">
			<div class="rayNew"></div>
			
			<!-- Auth content inside the card -->
			<div class="auth-content">
				<div class="auth-header">
					<h1>Welcome Back</h1>
					<p>Sign in to your account to continue</p>
				</div>

				<!-- OAuth2 Login Options -->
				<OAuth2Login />
				<hr>
				<br>
				<form on:submit|preventDefault={handleLogin} class="auth-form">
					<div class="form-group">
						<label for="email">Email</label>
						<input
							type="email"
							id="email"
							bind:value={email}
							on:input={checkEmailProvider}
							on:blur={checkEmailProvider}
							placeholder="Enter your email"
							class:gmail-detected={showGmailWarning}
							required
						/>
						{#if showGmailWarning}
							<div class="gmail-warning">
								<span class="gmail-icon">📧</span>
								<p>
									<strong>Gmail account detected!</strong><br>
									Please use the "Sign in with Google" button above for the best experience.
								</p>
							</div>
						{/if}
					</div>

					<div class="form-group">
						<label for="password">Password</label>
						<input
							type="password"
							id="password"
							bind:value={password}
							placeholder="Enter your password"
							autocomplete="current-password"
							required
						/>
					</div>

					{#if error}
						<div class="error-message">
							{error}
						</div>
					{/if}

					<button type="submit" class="btn-primary" disabled={loading}>
						{loading ? 'Signing in...' : 'Sign In'}
					</button>
				</form>

				<div class="auth-footer">
					<p>
						Don't have an account?
						<a href="/auth/register{returnUrl !== '/' || planId ? `?return=${encodeURIComponent(returnUrl)}&plan_id=${planId}` : ''}" class="link">Sign up</a>
					</p>
					<p>
						<a href="/auth/forgot-password" class="link">Forgot your password?</a>
					</p>
				</div>
			</div>
			
			<!-- Decorative lines -->
			<div class="lineNew toplNew"></div>
			<div class="lineNew leftlNew"></div>
			<div class="lineNew bottomlNew"></div>
			<div class="lineNew rightlNew"></div>
		</div>
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

	.auth-header {
		text-align: center;
		margin-bottom: 2rem;
	}

	.auth-header h1 {
		font-size: 2rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.auth-header p {
		color: var(--text-inverse);
		font-size: 0.9rem;
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
		min-height: 70px;
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

	.btn-primary {
		padding: 0.75rem 1.5rem;
		min-height: 70px;
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

	/* Gmail Detection Styles */
	.gmail-detected {
		border: 2px solid #4285f4 !important;
		background: #e8f0fe !important;
	}

	.gmail-warning {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		margin-top: 0.5rem;
		padding: 0.75rem 1rem;
		background: linear-gradient(135deg, #e8f0fe 0%, #d2e3fc 100%);
		border: 1px solid #4285f4;
		border-radius: 8px;
		animation: slideIn 0.3s ease-out;
	}

	.gmail-icon {
		font-size: 1.5rem;
		flex-shrink: 0;
	}

	.gmail-warning p {
		margin: 0;
		font-size: 0.85rem;
		color: #1a73e8;
		line-height: 1.4;
	}

	.gmail-warning strong {
		color: #1967d2;
	}

	@keyframes slideIn {
		from {
			opacity: 0;
			transform: translateY(-10px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	/* Auth content wrapper inside the animated card */
	.auth-content {
		position: relative;
		z-index: 10;
		width: 100%;
		padding: 1.5rem;
		overflow-y: visible;
		/* Counter the card's breathing animation to keep content stable */
		animation: counter-breathe 8s ease-in-out infinite;
	}

	@keyframes counter-breathe {
		0%,
		100% {
			transform: scale(1);
		}
		50% {
			transform: scale(1.0204); /* 1 / 0.98 ≈ 1.0204 */
		}
	}

	/* Adjust text color for better visibility on the gradient card */
	.cardNew .auth-header h1 {
		color: #fff;
		text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
	}

	.cardNew .auth-header p {
		color: #e7cab2;
	}

	.cardNew .form-group label {
		color: #e7cab2;
		font-weight: 600;
	}

	/* Keep form elements readable */
	.cardNew input {
		background: rgba(255, 255, 255, 0.95) !important;
		color: #1a1a1a;
	}

	/* Adjust button styling */
	.cardNew .btn-primary {
		background: linear-gradient(135deg, #ffeb7a 0%, #ffd700 100%);
		color: #1a1a1a;
		font-weight: 700;
		border: 2px solid rgba(255, 255, 255, 0.3);
	}

	.cardNew .btn-primary:hover:not(:disabled) {
		background: linear-gradient(135deg, #ffd700 0%, #ffeb7a 100%);
		transform: translateY(-2px);
		box-shadow: 0 8px 20px rgba(255, 235, 122, 0.4);
	}

	/* Error and success messages */
	.cardNew .error-message {
		background: rgba(255, 0, 0, 0.1);
		border: 1px solid #ff6b6b;
		color: #fff;
	}

	/* Auth footer links */
	.cardNew .auth-footer p {
		color: #e7cab2;
	}

	.cardNew .link {
		color: #ffeb7a;
		font-weight: 700;
	}

	.cardNew .link:hover {
		color: #fff;
		text-shadow: 0 0 8px #ffeb7a;
	}

	/* Adjust hr styling */
	.cardNew hr {
		border: none;
		height: 1px;
		background: linear-gradient(90deg, transparent, #888888, transparent);
		margin: 1.5rem 0;
	}

	/* Responsive sizing for the outer container */
	@media (max-width: 768px) {
		.outerNew {
			width: 95%;
			max-width: 450px;
			min-height: auto;
			height: auto;
		}
		
		.auth-content {
			max-height: none;
		}
	}
</style> 
