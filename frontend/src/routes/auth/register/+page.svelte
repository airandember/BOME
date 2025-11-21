<script lang="ts">
	import { auth } from '$lib/auth';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import Navigation from '$lib/components/Navigation.svelte';
	import OAuth2Login from '$lib/components/OAuth2Login.svelte';

	let email = '';
	let firstName = '';
	let lastName = '';
	let loading = false;
	let error = '';
	let success = false;
	
	// 🔗 CONTEXT PRESERVATION: Track return URL and plan context
	let returnUrl = '/';
	let planId = '';

	onMount(() => {
		// Redirect if already logged in
		auth.subscribe((state) => {
			if (state.isAuthenticated) {
				goto('/');
			}
		});
		
		// 🔗 Read and store subscription context from query params
		if (typeof window !== 'undefined') {
			returnUrl = $page.url.searchParams.get('return') || '/';
			planId = $page.url.searchParams.get('plan_id') || '';
			
			// Store in sessionStorage for use after email verification
			if (planId) {
				sessionStorage.setItem('selected_plan_id', planId);
				console.log('📋 Subscription context saved:', { planId, returnUrl });
			}
			if (returnUrl && returnUrl !== '/') {
				sessionStorage.setItem('post_verify_return', returnUrl);
			}
		}
	});

	async function handleRegister() {
		if (!email || !firstName || !lastName) {
			error = 'Please fill in all fields';
			return;
		}

		loading = true;
		error = '';

		const result = await auth.register({
			email,
			first_name: firstName,
			last_name: lastName
		});
		
		if (result.success) {
			success = true;
			// Redirect to email verification page instead of login
			setTimeout(() => {
				const verifyUrl = `/auth/verify-email?email=${encodeURIComponent(email)}&user_id=${result.user_id || ''}`;
				goto(verifyUrl);
			}, 2000);
		} else {
			error = result.error || 'Registration failed';
		}

		loading = false;
	}
</script>

<svelte:head>
	<title>Register - Book of Mormon Evidences</title>
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
					<h1>Get Started with BOME</h1>
					<p>Join us to explore Book of Mormon evidences - we'll help you set up your account!</p>
				</div>

				{#if success}
					<div class="success-message">
						<h3>Welcome to BOME!</h3>
						<p>Please check your email to complete your account setup. We'll help you create a secure password in the next step!</p>
					</div>
				{:else}
					<!-- OAuth2 Registration Options -->
					{#if !success}
					<OAuth2Login />
					{/if}
					<hr>
					<form on:submit|preventDefault={handleRegister} class="auth-form">
						<div class="form-row">
							<div class="form-group">
								<label for="firstName">First Name</label>
								<input
									type="text"
									id="firstName"
									bind:value={firstName}
									placeholder="Enter your first name"
									required
								/>
							</div>

							<div class="form-group">
								<label for="lastName">Last Name</label>
								<input
									type="text"
									id="lastName"
									bind:value={lastName}
									placeholder="Enter your last name"
									required
								/>
							</div>
						</div>

						<div class="form-group">
							<label for="email">Email</label>
							<input
								type="email"
								id="email"
								bind:value={email}
								placeholder="Enter your email"
								required
							/>
						</div>

						{#if error}
							<div class="error-message">
								{error}
							</div>
						{/if}

						<button type="submit" class="btn-primary" disabled={loading}>
							{loading ? 'Getting Started...' : 'Get Started'}
						</button>
					</form>
				{/if}

				<div class="auth-footer">
					<p>
						Already have an account?
						<a href="/auth/login" class="link">Log in</a>
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
		color: var(--text-primary);
		font-size: 0.9rem;
	}

	.auth-form {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
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
		color: var(--text-secondary);
		font-size: 1rem;
		box-shadow: 
			inset 2px 2px 4px var(--shadow-dark),
			inset -2px -2px 4px var(--shadow-light);
		transition: all 0.2s ease;
		border-radius: 14px;
        
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
		color: var(--text-primary);
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

	.success-message {
		padding: 1.5rem;
		background: var(--success-bg);
		color: var(--success-text);
		border-radius: 12px;
		text-align: center;
		margin-bottom: 1.5rem;
	}

	.success-message h3 {
		margin-bottom: 0.5rem;
		font-size: 1.2rem;
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

	@media (max-width: 480px) {
		.form-row {
			grid-template-columns: 1fr;
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

	.cardNew .success-message {
		background: rgba(0, 255, 0, 0.1);
		border: 1px solid #51cf66;
		color: #fff;
	}

	.cardNew .success-message h3 {
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
