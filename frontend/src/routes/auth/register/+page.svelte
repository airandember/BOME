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

		const result = await auth.register({
			email,
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
					const verifyUrl = `/auth/verify-email?email=${encodeURIComponent(email)}&user_id=${result.user_id || ''}`;
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
							<a href="/auth/login" class="btn-primary">
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

						<button type="submit" class="btn-primary btn-large" disabled={loading}>
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
		background: var(--bg-large-gradient2);

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

	/* Progress Steps */
	.progress-steps {
		display: flex;
		align-items: center;
		justify-content: center;
		margin-bottom: 2rem;
		gap: 0.5rem;
	}

	.step {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.25rem;
	}

	.step-number {
		width: 32px;
		height: 32px;
		border-radius: 50%;
		background: rgba(255, 255, 255, 0.2);
		color: #e7cab2;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 700;
		font-size: 0.875rem;
		transition: all 0.3s;
	}

	.step.active .step-number {
		background: #ffeb7a;
		color: #1a1a1a;
		box-shadow: 0 0 15px rgba(255, 235, 122, 0.5);
	}

	.step.completed .step-number {
		background: #51cf66;
		color: white;
	}

	.step-label {
		font-size: 0.75rem;
		color: #e7cab2;
		font-weight: 500;
	}

	.step.active .step-label {
		color: #ffeb7a;
	}

	.step-line {
		width: 60px;
		height: 3px;
		background: rgba(255, 255, 255, 0.2);
		border-radius: 2px;
		margin: 0 0.5rem;
		margin-bottom: 1.25rem;
		transition: all 0.3s;
	}

	.step-line.completed {
		background: #51cf66;
	}

	/* Divider with text */
	.divider {
		display: flex;
		align-items: center;
		margin: 1.5rem 0;
		gap: 1rem;
	}

	.divider::before,
	.divider::after {
		content: '';
		flex: 1;
		height: 1px;
		background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.3), transparent);
	}

	.divider span {
		color: #e7cab2;
		font-size: 0.8rem;
		white-space: nowrap;
	}

	/* Form Instructions */
	.form-instructions {
		background: rgba(255, 235, 122, 0.1);
		border: 1px solid rgba(255, 235, 122, 0.3);
		border-radius: 8px;
		padding: 0.75rem 1rem;
		margin-bottom: 1rem;
	}

	.form-instructions p {
		margin: 0;
		color: #ffeb7a;
		font-size: 0.875rem;
	}

	/* Label with icon */
	.label-icon {
		margin-right: 0.25rem;
	}

	/* Input hint text */
	.input-hint {
		font-size: 0.75rem;
		color: #e7cab2;
		margin-top: 0.25rem;
		opacity: 0.8;
	}

	/* Error icon */
	.error-icon {
		margin-right: 0.5rem;
	}

	/* Large button */
	.btn-large {
		min-height: 56px;
		font-size: 1.1rem;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
	}

	/* Loading spinner */
	.spinner {
		width: 20px;
		height: 20px;
		border: 2px solid rgba(26, 26, 26, 0.3);
		border-top-color: #1a1a1a;
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	/* Success message enhancements */
	.success-icon {
		font-size: 3rem;
		margin-bottom: 1rem;
	}

	.success-details {
		background: rgba(255, 255, 255, 0.1);
		border-radius: 8px;
		padding: 1rem;
		margin: 1rem 0;
	}

	.success-details p {
		margin: 0.5rem 0;
		font-size: 0.9rem;
	}

	.success-hint {
		font-size: 0.85rem;
		opacity: 0.9;
		margin-top: 0.75rem;
	}

	/* Temp password success styling */
	.temp-password-success {
		background: linear-gradient(135deg, rgba(255, 235, 122, 0.2), rgba(255, 215, 0, 0.2));
		border: 2px solid #ffeb7a;
	}

	.temp-password-success h3 {
		color: #ffeb7a;
	}

	.temp-password-success .btn-primary {
		margin-top: 1rem;
		display: inline-block;
		text-decoration: none;
	}

	@media (max-width: 480px) {
		.progress-steps {
			gap: 0.25rem;
		}
		
		.step-line {
			width: 30px;
		}
		
		.step-label {
			font-size: 0.65rem;
		}
	}
</style> 
