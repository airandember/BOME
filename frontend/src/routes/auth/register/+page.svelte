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
	
	// Track which flow was used (temp password vs email verification)
	let usedTempPasswordFlow = false;
	let redirectTo = '';
	
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
		
		// Normalize email to lowercase
		const normalizedEmail = email.toLowerCase().trim();

		const result = await auth.register({
			email: normalizedEmail,
			first_name: firstName,
			last_name: lastName
		});
		
		if (result.success) {
			success = true;
			
			// Check if temp password was sent (existing Stripe subscriber)
			if (result.temp_password_sent) {
				usedTempPasswordFlow = true;
				redirectTo = result.redirect_to || '/auth/login';
				// Redirect to login page after showing success
				setTimeout(() => {
					goto(redirectTo);
				}, 4000);
			} else {
				// Standard email verification flow
				usedTempPasswordFlow = false;
				setTimeout(() => {
					const verifyUrl = `/auth/verify-email?email=${encodeURIComponent(normalizedEmail)}&user_id=${result.user_id || ''}`;
					goto(verifyUrl);
				}, 2000);
			}
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
	<div class="auth-card-sales">
		<div class="auth-content">
			<div class="auth-header">
				<h1>Get Started with BOME</h1>
				<p>Join us to explore Book of Mormon evidences - we'll help you set up your account!</p>
			</div>

			<!-- Progress Indicator -->
			<div class="progress-steps">
				<div class="step {success ? 'completed' : 'active'}">
					<span class="step-number">1</span>
					<span class="step-label">Your Info</span>
				</div>
				<div class="step-line {success ? 'completed' : ''}"></div>
				<div class="step {success ? 'active' : ''}">
					<span class="step-number">2</span>
					<span class="step-label">{usedTempPasswordFlow ? 'Check Email' : 'Verify Email'}</span>
				</div>
			</div>

			{#if success}
				{#if usedTempPasswordFlow}
					<!-- Temp Password Flow Success -->
					<div class="success-message temp-password-success">
						<div class="success-icon">🎉</div>
						<h3>Great News - Your Subscription Found!</h3>
						<p>We found your existing subscription and your account is ready to go!</p>
						<div class="success-details">
							<p><strong>Check your email ({email})</strong> for your temporary password.</p>
							<p>Use it to log in, then you can set your own password in your dashboard.</p>
						</div>
						<a href="/auth/login" class="btn-sales">
							Go to Login →
						</a>
					</div>
				{:else}
					<!-- Standard Verification Flow Success -->
					<div class="success-message">
						<div class="success-icon">📧</div>
						<h3>Welcome to BOME!</h3>
						<p>Please check your email at <strong>{email}</strong> to complete your account setup.</p>
						<p class="success-hint">We'll help you create a secure password in the next step!</p>
					</div>
				{/if}
			{:else}
				<!-- OAuth2 Registration Options -->
				<OAuth2Login />
				
				<div class="divider">
					<span>or register with email</span>
				</div>
				
				<form on:submit|preventDefault={handleRegister} class="auth-form">
					<!-- Simple Instructions -->
					<div class="form-instructions">
						<p>📝 <strong>Quick & Easy:</strong> Just enter your name and email below.</p>
					</div>
					
					<div class="form-row">
						<div class="form-group">
							<label for="firstName">
								<span class="label-icon">👤</span>
								First Name
							</label>
							<input
								type="text"
								id="firstName"
								bind:value={firstName}
								placeholder="John"
								required
								autocomplete="given-name"
							/>
						</div>

						<div class="form-group">
							<label for="lastName">
								<span class="label-icon">👤</span>
								Last Name
							</label>
							<input
								type="text"
								id="lastName"
								bind:value={lastName}
								placeholder="Smith"
								required
								autocomplete="family-name"
							/>
						</div>
					</div>

					<div class="form-group">
						<label for="email">
							<span class="label-icon">📧</span>
							Email Address
						</label>
						<input
							type="email"
							id="email"
							bind:value={email}
							placeholder="you@example.com"
							required
							autocomplete="email"
						/>
						<span class="input-hint">We'll send you a link to finish setting up your account</span>
					</div>

					{#if error}
						<div class="error-message">
							<span class="error-icon">⚠️</span>
							{error}
						</div>
					{/if}

					<button type="submit" class="btn-sales btn-large" disabled={loading}>
						{#if loading}
							<span class="spinner"></span>
							Setting up your account...
						{:else}
							Create My Account →
						{/if}
					</button>
				</form>
			{/if}

			<div class="auth-footer">
				<p>
					Already have an account?
					<a href="/auth/login" class="link">Log in here</a>
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

	@media (max-width: 480px) {
		.auth-card-sales .progress-steps {
			gap: 0.25rem;
		}
		
		.auth-card-sales .step-line {
			width: 30px;
		}
		
		.auth-card-sales .step-label {
			font-size: 0.65rem;
		}
	}
</style> 
