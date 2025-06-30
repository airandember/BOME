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

	// Tab state
	let activeTab: 'dashboard' | 'advertiser' = 'dashboard';

	onMount(() => {
		let unsubscribe: any = null;

		async function initializeDashboard() {
			try {
				console.log('Dashboard: Starting initialization');
				
				// Test backend connectivity first
				const backendReachable = await testBackendConnectivity();
				console.log('Dashboard: Backend reachable:', backendReachable);
				
				// Initialize auth first
				await initializeAuth();
				console.log('Dashboard: Auth initialization completed');
				
				// Subscribe to auth state changes after initialization
				unsubscribe = auth.subscribe((state) => {
					console.log('Dashboard: Auth state changed:', {
						isAuthenticated: state.isAuthenticated,
						user: state.user ? 'User exists' : 'No user',
						loading: state.loading
					});
					
					user = state.user;
					isAuthenticated = state.isAuthenticated;
					
					// Only proceed if we have a definitive auth state (not loading)
					if (!state.loading) {
						if (state.isAuthenticated && state.user) {
							// User is authenticated, proceed with dashboard setup
							console.log('Dashboard: User authenticated, setting up dashboard');
							
							// Check URL parameters for tab
							const urlParams = new URLSearchParams($page.url.search);
							const tabParam = urlParams.get('tab');
							
							console.log('Dashboard: URL params:', $page.url.search, 'tab param:', tabParam);
							
							if (tabParam === 'advertiser') {
								activeTab = 'advertiser';
							} else {
								activeTab = 'dashboard';
							}
							
							console.log('Dashboard: Active tab set to:', activeTab);
							
							// Clean up URL parameters after a delay to ensure tab is set
							setTimeout(() => {
								if (urlParams.has('tab') || urlParams.has('from')) {
									console.log('Dashboard: Cleaning up URL parameters');
									replaceState($page.url.pathname, {});
								}
							}, 100);
							
							// Set loading to false since we have user data
							loading = false;
						} else if (state.isAuthenticated === false) {
							// User is explicitly not authenticated, redirect to login
							console.log('Dashboard: User not authenticated, redirecting to login');
							goto('/login');
							return;
						}
					}
				});
				
				// Set a timeout to prevent infinite loading
				setTimeout(() => {
					if (loading) {
						console.log('Dashboard: Loading timeout reached, checking auth state');
						loading = false;
					}
				}, 3000);
				
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

	function switchTab(tab: 'dashboard' | 'advertiser') {
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
{:else}
	<div class="dashboard">
		<!-- Tab Navigation -->
		<div class="tab-navigation glass">
			<button 
				class="tab-button {activeTab === 'dashboard' ? 'active' : ''}"
				on:click={() => switchTab('dashboard')}
			>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
					<circle cx="12" cy="7" r="4"/>
				</svg>
				Dashboard
			</button>
			<button 
				class="tab-button"
				on:click={() => goto('/account')}
			>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
					<circle cx="12" cy="7" r="4"/>
				</svg>
				Account
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
	.advertiser-section {
		background: var(--card-bg);
		border-radius: 20px;
		padding: 2rem;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
		text-align: center;
	}

	.welcome-content h1,
	.account-section h1,
	.advertiser-section h1 {
		font-size: 2.5rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.welcome-content p,
	.account-section p,
	.advertiser-section p {
		font-size: 1.1rem;
		color: var(--text-secondary);
		margin-bottom: 2rem;
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