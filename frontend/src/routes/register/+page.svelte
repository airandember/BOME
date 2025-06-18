<script lang="ts">
	import { auth } from '$lib/auth';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	let email = '';
	let password = '';
	let confirmPassword = '';
	let firstName = '';
	let lastName = '';
	let loading = false;
	let error = '';
	let success = false;

	onMount(() => {
		// Redirect if already logged in
		auth.subscribe((state) => {
			if (state.isAuthenticated) {
				goto('/');
			}
		});
	});

	async function handleRegister() {
		if (!email || !password || !confirmPassword || !firstName || !lastName) {
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

		const result = await auth.register(email, password, firstName, lastName);
		
		if (result.success) {
			success = true;
			setTimeout(() => {
				goto('/login');
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

<div class="auth-container">
	<div class="auth-card">
		<div class="auth-header">
			<h1>Create Account</h1>
			<p>Join us to explore Book of Mormon evidences</p>
		</div>

		{#if success}
			<div class="success-message">
				<h3>Registration Successful!</h3>
				<p>Your account has been created. Redirecting to login...</p>
			</div>
		{:else}
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

				<div class="form-group">
					<label for="password">Password</label>
					<input
						type="password"
						id="password"
						bind:value={password}
						placeholder="Enter your password"
						required
					/>
				</div>

				<div class="form-group">
					<label for="confirmPassword">Confirm Password</label>
					<input
						type="password"
						id="confirmPassword"
						bind:value={confirmPassword}
						placeholder="Confirm your password"
						required
					/>
				</div>

				{#if error}
					<div class="error-message">
						{error}
					</div>
				{/if}

				<button type="submit" class="btn-primary" disabled={loading}>
					{loading ? 'Creating Account...' : 'Create Account'}
				</button>
			</form>
		{/if}

		<div class="auth-footer">
			<p>
				Already have an account?
				<a href="/login" class="link">Sign in</a>
			</p>
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

	.auth-card {
		width: 100%;
		max-width: 450px;
		padding: 2.5rem;
		background: var(--card-bg);
		border-radius: 20px;
		box-shadow: 
			8px 8px 16px var(--shadow-dark),
			-8px -8px 16px var(--shadow-light);
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
		color: var(--text-secondary);
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
		border: none;
		border-radius: 12px;
		background: var(--accent-color);
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
		color: var(--text-secondary);
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
</style> 