<script lang="ts">
	import { onMount } from 'svelte';
	import { goto, replaceState } from '$app/navigation';
	import { page } from '$app/stores';
	import { auth, initializeAuth, testBackendConnectivity } from '$lib/auth';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import Navigation from '$lib/components/Navigation.svelte';

	let user: any = null;
	let isAuthenticated = false;
	let loading = true;
	let error = '';

	// Tab state - now includes profile
	let activeTab: 'dashboard' | 'profile' | 'advertiser' = 'dashboard';

	onMount(() => {
		let unsubscribe: any = null;

		async function initializeDashboard() {
			try {
				console.log('Dashboard: Starting initialization');
				
				// Initialize auth first and wait for it to complete
				await initializeAuth();
				console.log('Dashboard: Auth initialization completed');
				
				// Subscribe to auth state changes after initialization
				unsubscribe = auth.subscribe((state) => {
					console.log('Dashboard: Auth state changed:', {
						isAuthenticated: state.isAuthenticated,
						user: state.user ? 'User exists' : 'No user',
						loading: state.loading,
						token: state.token ? 'Token exists' : 'No token'
					});
					
					user = state.user;
					isAuthenticated = state.isAuthenticated;
					
					// Only proceed if we have a definitive auth state (not loading)
					if (!state.loading) {
						console.log('Dashboard: Auth state not loading, checking authentication:', {
							isAuthenticated: state.isAuthenticated,
							hasUser: !!state.user,
							user: state.user
						});
						
						if (state.isAuthenticated && state.user) {
							// User is authenticated, proceed with dashboard setup
							console.log('Dashboard: User authenticated, setting up dashboard');
							
							// Check URL parameters for tab
							const urlParams = new URLSearchParams($page.url.search);
							const tabParam = urlParams.get('tab');
							
							console.log('Dashboard: URL params:', $page.url.search, 'tab param:', tabParam);
							
							if (tabParam === 'profile') {
								activeTab = 'profile';
							} else if (tabParam === 'advertiser') {
								activeTab = 'advertiser';
							} else {
								activeTab = 'dashboard';
							}
							
							console.log('Dashboard: Active tab set to:', activeTab);
							
							// Set loading to false since we have user data
							loading = false;
						} else if (state.isAuthenticated === false && !state.loading) {
							// User is explicitly not authenticated and not loading
							// Instead of immediate redirect, show a message and provide login option
							console.log('Dashboard: User not authenticated, showing login prompt');
							loading = false;
							error = 'Please log in to access your dashboard';
						}
					}
				});
				
				// Set a longer timeout to prevent infinite loading (increased from 3s to 10s)
				setTimeout(() => {
					if (loading) {
						console.log('Dashboard: Loading timeout reached, checking auth state');
						loading = false;
						error = 'Loading took longer than expected. Please refresh the page.';
					}
				}, 10000);
				
			} catch (err) {
				console.error('Error loading dashboard:', err);
				error = 'Some features may not be available';
				loading = false;
			}
		}

		// Start the async initialization
		initializeDashboard();

		// Return cleanup function
		return () => {
			if (unsubscribe) {
				unsubscribe();
			}
		};
	});

	function switchTab(tab: 'dashboard' | 'profile' | 'advertiser') {
		activeTab = tab;
	}
</script>

<svelte:head>
	<title>Dashboard - BOME</title>
</svelte:head>

<Navigation />

{#if loading}
	<div class="loading-container">
		<LoadingSpinner size="large" color="primary" />
		<p>Loading your dashboard...</p>
	</div>
{:else if error}
	<div class="error-container">
		<div class="error-content glass">
			<div class="error-icon">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="12" cy="12" r="10"></circle>
					<line x1="15" y1="9" x2="9" y2="15"></line>
					<line x1="9" y1="9" x2="15" y2="15"></line>
				</svg>
			</div>
			<h2>Access Required</h2>
			<p>{error}</p>
			<div class="error-actions">
				<a href="/login" class="btn btn-primary">Log In</a>
				<a href="/register" class="btn btn-ghost">Create Account</a>
				<button class="btn btn-outline" on:click={() => window.location.reload()}>Try Again</button>
			</div>
		</div>
	</div>
{:else}
	<div class="dashboard">
		<!-- Tab Navigation -->
		<div class="tab-navigation glass">
			<button 
				class="tab-button {activeTab === 'dashboard' ? 'active' : ''}"
				on:click={() => switchTab('dashboard')}
			>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
					<polyline points="9,22 9,12 15,12 15,22"/>
				</svg>
				Dashboard
			</button>
			<button 
				class="tab-button {activeTab === 'profile' ? 'active' : ''}"
				on:click={() => switchTab('profile')}
			>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
					<circle cx="12" cy="7" r="4"/>
				</svg>
				Profile
			</button>
			<button 
				class="tab-button {activeTab === 'advertiser' ? 'active' : ''}"
				on:click={() => switchTab('advertiser')}
			>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
					<circle cx="12" cy="12" r="3"/>
				</svg>
				Advertiser
			</button>
		</div>

		{#if activeTab === 'dashboard'}
			<!-- Dashboard Tab Content -->
			<div class="tab-content">
				<div class="welcome-section glass">
					<div class="welcome-content">
						<h1>Welcome back, {user?.first_name || 'User'}!</h1>
						<p>Continue your journey exploring Book of Mormon evidences</p>
					</div>
					
					<!-- Quick Profile Summary -->
					<div class="profile-summary">
						<div class="profile-card">
							<div class="profile-avatar">
								<span>{user?.first_name?.[0] || 'U'}</span>
							</div>
							<div class="profile-details">
								<h3>{user?.first_name} {user?.last_name}</h3>
								<p>{user?.email}</p>
								<span class="role-badge">{user?.role}</span>
							</div>
						</div>
					</div>
				</div>
			</div>
		{:else if activeTab === 'profile'}
			<!-- Profile Tab Content -->
			<div class="tab-content">
				<div class="profile-section glass">
					<h1>Your Profile</h1>
					<p>Manage your account information and settings</p>
					
					<div class="profile-info">
						<div class="info-row">
							<label>First Name:</label>
							<span>{user?.first_name || 'Not set'}</span>
						</div>
						<div class="info-row">
							<label>Last Name:</label>
							<span>{user?.last_name || 'Not set'}</span>
						</div>
						<div class="info-row">
							<label>Email:</label>
							<span>{user?.email}</span>
						</div>
						<div class="info-row">
							<label>Role:</label>
							<span class="role-badge">{user?.role}</span>
						</div>
						<div class="info-row">
							<label>Email Verified:</label>
							<span class="verification-status {user?.email_verified ? 'verified' : 'unverified'}">
								{user?.email_verified ? '✓ Verified' : '✗ Unverified'}
							</span>
						</div>
					</div>
					
					<div class="profile-actions">
						<button class="btn btn-primary" on:click={() => goto('/subscription')}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
							</svg>
							Manage Subscription
						</button>
						<button class="btn btn-secondary" on:click={() => goto('/dashboard?tab=advertiser')}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
								<circle cx="12" cy="12" r="3"/>
							</svg>
							Advertiser Portal
						</button>
					</div>
				</div>
			</div>
		{:else if activeTab === 'advertiser'}
			<!-- Advertiser Tab Content -->
			<div class="tab-content">
				<div class="advertiser-section glass">
					<h1>Advertiser Dashboard</h1>
					<p>Manage your advertising campaigns and analytics</p>
					<div class="advertiser-actions">
						<button class="btn btn-primary" on:click={() => goto('/advertiser')}>
							Go to Advertiser Portal
						</button>
					</div>
				</div>
			</div>
		{/if}
	</div>
{/if}

<style>
	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 50vh;
		gap: 1rem;
	}

	.error-container {
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 50vh;
		padding: 2rem;
	}

	.error-content {
		text-align: center;
		max-width: 500px;
		padding: 3rem 2rem;
		border-radius: 20px;
		backdrop-filter: blur(10px);
		border: 1px solid var(--border-color);
		box-shadow: var(--shadow-lg);
	}

	.error-icon {
		width: 64px;
		height: 64px;
		margin: 0 auto 1.5rem;
		color: var(--error);
	}

	.error-icon svg {
		width: 100%;
		height: 100%;
	}

	.error-content h2 {
		font-size: 1.5rem;
		font-weight: 600;
		margin-bottom: 1rem;
		color: var(--text-primary);
	}

	.error-content p {
		color: var(--text-secondary);
		margin-bottom: 2rem;
		line-height: 1.6;
	}

	.error-actions {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		align-items: center;
	}

	.error-actions .btn {
		min-width: 150px;
	}

	@media (min-width: 640px) {
		.error-actions {
			flex-direction: row;
			justify-content: center;
		}
	}

	.dashboard {
		min-height: 100vh;
		padding: 2rem 0;
		background: var(--bg-gradient);
	}

	.tab-navigation {
		display: flex;
		justify-content: center;
		gap: 1rem;
		margin-bottom: 2rem;
		padding: 1rem;
		border-radius: 20px;
		backdrop-filter: blur(10px);
		border: 1px solid var(--border-color);
	}

	.tab-button {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 12px;
		background: var(--bg-secondary);
		color: var(--text-secondary);
		cursor: pointer;
		transition: all 0.3s ease;
		font-size: 0.875rem;
		font-weight: 500;
	}

	.tab-button:hover {
		background: var(--primary-color);
		color: white;
		transform: translateY(-2px);
	}

	.tab-button.active {
		background: var(--primary-color);
		color: white;
		box-shadow: 0 4px 12px rgba(var(--primary-rgb), 0.3);
	}

	.tab-button svg {
		width: 18px;
		height: 18px;
	}

	.tab-content {
		max-width: 1200px;
		margin: 0 auto;
		padding: 0 1rem;
	}

	.welcome-section,
	.account-section,
	.advertiser-section,
	.profile-section {
		background: var(--card-bg);
		border-radius: 20px;
		padding: 2rem;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
		text-align: center;
	}

	.welcome-content h1,
	.account-section h1,
	.advertiser-section h1,
	.profile-section h1 {
		font-size: 2.5rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.welcome-content p,
	.account-section p,
	.advertiser-section p,
	.profile-section p {
		font-size: 1.1rem;
		color: var(--text-secondary);
		margin-bottom: 2rem;
	}

	/* Profile Summary Styles */
	.profile-summary {
		margin-top: 2rem;
		display: flex;
		justify-content: center;
	}

	.profile-card {
		display: flex;
		align-items: center;
		gap: 1rem;
		background: var(--bg-secondary);
		padding: 1.5rem;
		border-radius: 16px;
		border: 1px solid var(--border-color);
		max-width: 400px;
		width: 100%;
	}

	.profile-avatar {
		width: 60px;
		height: 60px;
		border-radius: 50%;
		background: var(--primary-color);
		display: flex;
		align-items: center;
		justify-content: center;
		color: white;
		font-size: 1.5rem;
		font-weight: 700;
		flex-shrink: 0;
	}

	.profile-details {
		flex: 1;
		text-align: left;
	}

	.profile-details h3 {
		margin: 0 0 0.25rem 0;
		color: var(--text-primary);
		font-size: 1.25rem;
		font-weight: 600;
	}

	.profile-details p {
		margin: 0 0 0.5rem 0;
		color: var(--text-secondary);
		font-size: 0.875rem;
	}

	/* Profile Info Styles */
	.profile-info {
		text-align: left;
		max-width: 500px;
		margin: 0 auto 2rem auto;
		background: var(--bg-secondary);
		padding: 1.5rem;
		border-radius: 12px;
		border: 1px solid var(--border-color);
	}

	.info-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.75rem 0;
		border-bottom: 1px solid var(--border-color);
	}

	.info-row:last-child {
		border-bottom: none;
	}

	.info-row label {
		font-weight: 600;
		color: var(--text-primary);
		font-size: 0.875rem;
	}

	.info-row span {
		color: var(--text-secondary);
		font-size: 0.875rem;
	}

	.role-badge {
		background: var(--primary-color);
		color: white;
		padding: 0.25rem 0.75rem;
		border-radius: 20px;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
	}

	.verification-status {
		padding: 0.25rem 0.75rem;
		border-radius: 20px;
		font-size: 0.75rem;
		font-weight: 600;
	}

	.verification-status.verified {
		background: #10b981;
		color: white;
	}

	.verification-status.unverified {
		background: #ef4444;
		color: white;
	}

	/* Profile Actions */
	.profile-actions {
		display: flex;
		justify-content: center;
		gap: 1rem;
		margin-top: 2rem;
		flex-wrap: wrap;
	}

	.btn-secondary {
		background: var(--bg-secondary);
		color: var(--text-primary);
		border: 1px solid var(--border-color);
	}

	.btn-secondary:hover {
		background: var(--border-color);
		transform: translateY(-2px);
	}

	.account-info {
		text-align: left;
		max-width: 400px;
		margin: 0 auto;
		background: var(--bg-secondary);
		padding: 1.5rem;
		border-radius: 12px;
	}

	.account-info p {
		margin: 0.5rem 0;
		color: var(--text-primary);
	}

	.glass {
		backdrop-filter: blur(10px);
		background: rgba(255, 255, 255, 0.1);
		border: 1px solid rgba(255, 255, 255, 0.2);
	}

	.advertiser-actions {
		display: flex;
		justify-content: center;
		gap: 1rem;
		margin-top: 2rem;
	}

	.btn {
		display: inline-flex;
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

	.btn-primary {
		background: var(--primary-color);
		color: white;
	}

	.btn-primary:hover {
		background: var(--primary-hover);
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(var(--primary-rgb), 0.3);
	}
</style> 