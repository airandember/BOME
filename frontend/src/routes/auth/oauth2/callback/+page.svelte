<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { apiRequest, auth, storeAuthData } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let loading = true;
	let error = '';
	let processing = 'Processing OAuth2 authentication...';

	onMount(async () => {
		await handleOAuth2Callback();
	});

	async function handleOAuth2Callback() {
		try {
			// Get URL parameters
			const urlParams = new URLSearchParams(window.location.search);
			const code = urlParams.get('code');
			const state = urlParams.get('state');
			const errorParam = urlParams.get('error');

			// Check for OAuth2 errors
			if (errorParam) {
				throw new Error(`OAuth2 error: ${errorParam}`);
			}

			if (!code || !state) {
				throw new Error('Missing OAuth2 parameters');
			}

			processing = 'Exchanging authorization code...';

			// Send callback data to backend
			const response = await apiRequest('/oauth2/callback', {
				method: 'POST',
				body: JSON.stringify({
					code: code,
					state: state
				})
			});

			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'OAuth2 authentication failed');
			}

			const data = await response.json();

			processing = 'Completing authentication...';

			// Store authentication data (same as regular login)
			const tokens = {
				access_token: data.access_token,
				refresh_token: data.refresh_token,
				expires_in: data.expires_in,
				token_type: data.token_type
			};

			const user = {
				id: data.user.id,
				email: data.user.email,
				role: data.user.role,
				first_name: data.user.first_name,
				last_name: data.user.last_name,
				email_verified: data.user.email_verified
			};

			// Store authentication data using the same system as regular login
			storeAuthData(tokens, user);

			processing = 'Redirecting...';

			// Show success message
			const providerName = data.provider || 'OAuth2';
			const message = data.is_new_user 
				? `Welcome! Account created with ${providerName}`
				: `Welcome back! Signed in with ${providerName}`;
			
			showToast(message, 'success');

			// Redirect based on user role
			setTimeout(() => {
				if (user.role === 'admin' || user.role === 'super_admin') {
					goto('/admin');
				} else {
					goto('/');
				}
			}, 1000);

		} catch (err: any) {
			console.error('OAuth2 callback error:', err);
			error = err.message || 'OAuth2 authentication failed';
			loading = false;

			// Redirect to login page after showing error
			setTimeout(() => {
				goto('/login');
			}, 3000);
		}
	}
</script>

<svelte:head>
	<title>Authenticating... - BOME</title>
</svelte:head>

<div class="callback-container">
	<div class="callback-card">
		{#if loading && !error}
			<div class="loading-section">
				<LoadingSpinner />
				<h2>Authenticating</h2>
				<p>{processing}</p>
				<div class="progress-dots">
					<span class="dot"></span>
					<span class="dot"></span>
					<span class="dot"></span>
				</div>
			</div>
		{:else if error}
			<div class="error-section">
				<div class="error-icon">⚠️</div>
				<h2>Authentication Failed</h2>
				<p class="error-message">{error}</p>
				<p class="redirect-message">Redirecting to login page...</p>
				<button class="btn btn-primary" onclick={() => goto('/login')}>
					Return to Login
				</button>
			</div>
		{/if}
	</div>
</div>

<style>
	.callback-container {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 2rem;
		background: var(--bg-color);
	}

	.callback-card {
		width: 100%;
		max-width: 500px;
		padding: 3rem 2rem;
		background: var(--card-bg);
		border-radius: 20px;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
		text-align: center;
	}

	.loading-section,
	.error-section {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1.5rem;
	}

	.loading-section h2,
	.error-section h2 {
		margin: 0;
		color: var(--text-primary);
		font-size: 1.5rem;
		font-weight: 600;
	}

	.loading-section p,
	.error-section p {
		margin: 0;
		color: var(--text-secondary);
		font-size: 1rem;
		line-height: 1.5;
	}

	.error-icon {
		font-size: 3rem;
		margin-bottom: 0.5rem;
	}

	.error-message {
		color: var(--error-text) !important;
		font-weight: 500;
	}

	.redirect-message {
		font-size: 0.875rem !important;
		opacity: 0.8;
	}

	.progress-dots {
		display: flex;
		gap: 0.5rem;
		justify-content: center;
		margin-top: 1rem;
	}

	.dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--primary-color);
		animation: pulse 1.5s ease-in-out infinite;
	}

	.dot:nth-child(2) {
		animation-delay: 0.3s;
	}

	.dot:nth-child(3) {
		animation-delay: 0.6s;
	}

	@keyframes pulse {
		0%, 100% {
			opacity: 0.3;
			transform: scale(1);
		}
		50% {
			opacity: 1;
			transform: scale(1.2);
		}
	}

	.btn {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 0.5rem;
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
	}

	.btn-primary {
		background: var(--primary-color);
		color: white;
	}

	.btn-primary:hover {
		background: var(--primary-color-dark);
		transform: translateY(-1px);
	}

	@media (max-width: 640px) {
		.callback-container {
			padding: 1rem;
		}
		
		.callback-card {
			padding: 2rem 1.5rem;
		}
		
		.loading-section h2,
		.error-section h2 {
			font-size: 1.25rem;
		}
	}
</style>
