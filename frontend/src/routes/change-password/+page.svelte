<script lang="ts">
	import { auth, SecureTokenStorage, isAdmin } from '$lib/auth';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	let currentPassword = '';
	let newPassword = '';
	let confirmPassword = '';
	let loading = false;
	let error = '';
	let success = '';

	onMount(() => {
		// Check if user is authenticated
		auth.subscribe((state) => {
			if (!state.isAuthenticated && !state.loading) {
				goto('/login');
			}
		});
	});

	async function handlePasswordChange() {
		if (!currentPassword || !newPassword || !confirmPassword) {
			error = 'Please fill in all fields';
			return;
		}

		if (newPassword !== confirmPassword) {
			error = 'New passwords do not match';
			return;
		}

		if (newPassword.length < 8) {
			error = 'New password must be at least 8 characters long';
			return;
		}

		loading = true;
		error = '';

		try {
			const response = await fetch('/api/v1/auth/change-password', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${SecureTokenStorage.getAccessToken()}`
				},
				body: JSON.stringify({
					current_password: currentPassword,
					new_password: newPassword
				})
			});

			const data = await response.json();

			if (response.ok) {
				success = 'Password changed successfully! Redirecting...';
				// Clear form
				currentPassword = '';
				newPassword = '';
				confirmPassword = '';
				
				// Redirect after a short delay
				setTimeout(() => {
					if (isAdmin()) {
						goto('/admin');
					} else {
						goto('/');
					}
				}, 2000);
			} else {
				error = data.error || 'Failed to change password';
			}
		} catch (err) {
			error = 'Network error. Please try again.';
		}

		loading = false;
	}
</script>

<svelte:head>
	<title>Change Password - Book of Mormon Evidences</title>
</svelte:head>

<div class="auth-container">
	<div class="auth-card">
		<div class="auth-header">
			<h1>🔐 Change Your Password</h1>
			<p>For security reasons, you must change your password before continuing.</p>
		</div>

		{#if success}
			<div class="success-message">
				{success}
			</div>
		{:else}
			<form on:submit|preventDefault={handlePasswordChange} class="auth-form">
				<div class="form-group">
					<label for="current-password">Current Password</label>
					<input
						type="password"
						id="current-password"
						bind:value={currentPassword}
						placeholder="Enter your current password"
						required
						autocomplete="current-password"
					/>
				</div>

				<div class="form-group">
					<label for="new-password">New Password</label>
					<input
						type="password"
						id="new-password"
						bind:value={newPassword}
						placeholder="Enter your new password"
						required
						autocomplete="new-password"
					/>
					<small class="password-hint">
						Password must be at least 8 characters long
					</small>
				</div>

				<div class="form-group">
					<label for="confirm-password">Confirm New Password</label>
					<input
						type="password"
						id="confirm-password"
						bind:value={confirmPassword}
						placeholder="Confirm your new password"
						required
						autocomplete="new-password"
					/>
				</div>

				{#if error}
					<div class="error-message">
						{error}
					</div>
				{/if}

				<button type="submit" class="btn-primary" disabled={loading}>
					{loading ? 'Changing Password...' : 'Change Password'}
				</button>
			</form>
		{/if}
	</div>
</div>

<style>
	.auth-container {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
		padding: 1rem;
	}

	.auth-card {
		background: white;
		border-radius: 12px;
		box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
		padding: 2rem;
		width: 100%;
		max-width: 400px;
	}

	.auth-header {
		text-align: center;
		margin-bottom: 2rem;
	}

	.auth-header h1 {
		color: #1f2937;
		font-size: 1.875rem;
		font-weight: 700;
		margin-bottom: 0.5rem;
	}

	.auth-header p {
		color: #6b7280;
		font-size: 0.875rem;
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
		color: #374151;
		font-size: 0.875rem;
		font-weight: 500;
	}

	.form-group input {
		padding: 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 0.5rem;
		font-size: 0.875rem;
		transition: border-color 0.2s;
	}

	.form-group input:focus {
		outline: none;
		border-color: #667eea;
		box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
	}

	.password-hint {
		color: #6b7280;
		font-size: 0.75rem;
	}

	.btn-primary {
		background: #667eea;
		color: white;
		padding: 0.75rem;
		border: none;
		border-radius: 0.5rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: background-color 0.2s;
	}

	.btn-primary:hover:not(:disabled) {
		background: #5a67d8;
	}

	.btn-primary:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.error-message {
		background: #fef2f2;
		color: #dc2626;
		padding: 0.75rem;
		border-radius: 0.5rem;
		font-size: 0.875rem;
		border: 1px solid #fecaca;
	}

	.success-message {
		background: #f0fdf4;
		color: #16a34a;
		padding: 0.75rem;
		border-radius: 0.5rem;
		font-size: 0.875rem;
		border: 1px solid #bbf7d0;
		text-align: center;
	}
</style>
