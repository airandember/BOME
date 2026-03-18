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
		
		// Normalize email to lowercase
		const normalizedEmail = email.toLowerCase().trim();

		const result = await auth.login(normalizedEmail, password);
		
		if (result.success) {
			// 🔗 CONTEXT PRESERVATION: If user was subscribing, redirect back with auto_checkout
			if (planId && returnUrl === '/subscription') {
				goto(`/subscription?auto_checkout=true&plan_id=${planId}`);
			} else if (planId && returnUrl.startsWith('/secretsub/')) {
				// 🔐 Secret promo subscription flow - redirect back with auto_checkout
				goto(`${returnUrl}?auto_checkout=true`);
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
	<div class="auth-card-sales">
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

				<button type="submit" class="btn-sales" disabled={loading}>
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
	</div>
</div>

<style>
	.auth-container {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 2rem;
		background: var(--sales-off-white);
	}

	.auth-content {
		width: 100%;
	}

	/* Gmail detection - input highlight */
	.gmail-detected {
		border: 2px solid #4285f4 !important;
		background: #e8f0fe !important;
	}

	.gmail-icon {
		font-size: 1.25rem;
		flex-shrink: 0;
	}

	.gmail-warning {
		animation: slideIn 0.3s ease-out;
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
</style> 
