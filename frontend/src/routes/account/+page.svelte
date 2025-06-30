<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import Navigation from '$lib/components/Navigation.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import { auth, type User } from '$lib/auth';

	let user: User | null = null;
	let loading = true;
	let error: string | null = null;

	onMount(() => {
		console.log('Account page: Loading user data');
		
		// Subscribe to auth store
		const unsubscribe = auth.subscribe((authState) => {
			if (authState.user) {
				user = authState.user;
				console.log('Account page: User loaded:', user);
			} else {
				console.log('Account page: No user found, redirecting to login');
				// Redirect to login if no user
				window.location.href = '/login';
				return;
			}
			loading = false;
		});

		// Return cleanup function
		return () => {
			unsubscribe();
		};
	});
</script>

<svelte:head>
	<title>Account Settings - BOME</title>
</svelte:head>

<Navigation />

{#if loading}
	<div class="loading-container">
		<LoadingSpinner size="large" color="primary" />
		<p>Loading your account...</p>
	</div>
{:else if user}
	<div class="account-page">
		<div class="account-container">
			<div class="account-header glass">
				<h1>Account Settings</h1>
				<p>Manage your account information and preferences</p>
			</div>

			<div class="account-content">
				<div class="account-section glass">
					<h2>Profile Information</h2>
					<div class="profile-info">
						<div class="info-row">
							<label>First Name:</label>
							<span>{user.first_name || 'Not set'}</span>
						</div>
						<div class="info-row">
							<label>Last Name:</label>
							<span>{user.last_name || 'Not set'}</span>
						</div>
						<div class="info-row">
							<label>Email:</label>
							<span>{user.email}</span>
						</div>
						<div class="info-row">
							<label>Role:</label>
							<span class="role-badge">{user.role}</span>
						</div>
						<div class="info-row">
							<label>Email Verified:</label>
							<span>{user.email_verified ? 'Yes' : 'No'}</span>
						</div>
					</div>
				</div>

				<div class="account-section glass">
					<h2>Account Actions</h2>
					<div class="action-buttons">
						<button class="btn btn-primary" on:click={() => window.location.href = '/dashboard?tab=account'}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
								<circle cx="12" cy="7" r="4"/>
							</svg>
							Go to Dashboard
						</button>
						<button class="btn btn-secondary" on:click={() => window.location.href = '/account/profile'}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
								<circle cx="12" cy="7" r="4"/>
							</svg>
							Edit Profile
						</button>
						<button class="btn btn-secondary" on:click={() => window.location.href = '/account/settings'}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<circle cx="12" cy="12" r="3"/>
								<path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1 1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
							</svg>
							Settings
						</button>
					</div>
				</div>
			</div>
		</div>
	</div>
{:else}
	<div class="error-container">
		<p>Unable to load account information. Please try again.</p>
		<button class="btn btn-primary" on:click={() => window.location.reload()}>
			Retry
		</button>
	</div>
{/if}

<style>
	.loading-container,
	.error-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 50vh;
		gap: 1rem;
	}

	.account-page {
		min-height: 100vh;
		padding: 2rem 0;
		background: var(--bg-gradient);
	}

	.account-container {
		max-width: 800px;
		margin: 0 auto;
		padding: 0 1rem;
	}

	.account-header {
		text-align: center;
		padding: 2rem;
		border-radius: 20px;
		margin-bottom: 2rem;
		backdrop-filter: blur(10px);
		border: 1px solid var(--border-color);
	}

	.account-header h1 {
		font-size: 2.5rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.account-header p {
		font-size: 1.1rem;
		color: var(--text-secondary);
	}

	.account-content {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.account-section {
		background: var(--card-bg);
		border-radius: 20px;
		padding: 2rem;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
	}

	.account-section h2 {
		font-size: 1.5rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1.5rem;
	}

	.profile-info {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.info-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem;
		background: var(--bg-secondary);
		border-radius: 12px;
		border: 1px solid var(--border-color);
	}

	.info-row label {
		font-weight: 600;
		color: var(--text-secondary);
		min-width: 120px;
	}

	.info-row span {
		color: var(--text-primary);
		font-weight: 500;
	}

	.role-badge {
		background: var(--primary-color);
		color: white;
		padding: 0.25rem 0.75rem;
		border-radius: 20px;
		font-size: 0.875rem;
		font-weight: 600;
		text-transform: capitalize;
	}

	.action-buttons {
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
	}

	.btn {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 12px;
		cursor: pointer;
		transition: all 0.3s ease;
		font-size: 0.875rem;
		font-weight: 500;
		text-decoration: none;
	}

	.btn svg {
		width: 18px;
		height: 18px;
	}

	.btn-primary {
		background: var(--primary-color);
		color: white;
	}

	.btn-primary:hover {
		background: var(--primary-hover);
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(var(--primary-rgb), 0.3);
	}

	.btn-secondary {
		background: var(--bg-secondary);
		color: var(--text-secondary);
		border: 1px solid var(--border-color);
	}

	.btn-secondary:hover {
		background: var(--primary-color);
		color: white;
		transform: translateY(-2px);
	}

	.glass {
		backdrop-filter: blur(10px);
		background: rgba(255, 255, 255, 0.1);
		border: 1px solid rgba(255, 255, 255, 0.2);
	}

	@media (max-width: 768px) {
		.account-container {
			padding: 0 0.5rem;
		}

		.account-header h1 {
			font-size: 2rem;
		}

		.info-row {
			flex-direction: column;
			align-items: flex-start;
			gap: 0.5rem;
		}

		.info-row label {
			min-width: auto;
		}

		.action-buttons {
			flex-direction: column;
		}

		.btn {
			justify-content: center;
		}
	}
</style> 