<script lang="ts">
	import { onMount } from 'svelte';
	import { goto, replaceState } from '$app/navigation';
	import { page } from '$app/stores';
	import { auth, initializeAuth, testBackendConnectivity } from '$lib/auth';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import Navigation from '$lib/components/Navigation.svelte';
	import SubscriptionManagementCard from '$lib/components/SubscriptionManagementCard.svelte';
	import SubscriptionManagement from '$lib/components/SubscriptionManagement.svelte';

	let user: any = null;
	let isAuthenticated = false;
	let loading = true;
	let error = '';

	// Tab state - now includes profile and subscription
	let activeTab: 'dashboard' | 'profile' | 'subscription' | 'advertiser' = 'dashboard';

	// Change Password state
	let showChangePassword = false;
	let currentPassword = '';
	let newPassword = '';
	let confirmNewPassword = '';
	let passwordChangeLoading = false;
	let passwordChangeError = '';
	let passwordChangeSuccess = false;

	// Password strength indicator
	let passwordStrength: 'weak' | 'medium' | 'strong' = 'weak';

	// Check password strength
	$: {
		if (newPassword.length === 0) {
			passwordStrength = 'weak';
		} else if (newPassword.length < 8) {
			passwordStrength = 'weak';
		} else if (newPassword.length < 12 && /[A-Z]/.test(newPassword) && /[0-9]/.test(newPassword)) {
			passwordStrength = 'medium';
		} else if (newPassword.length >= 12 && /[A-Z]/.test(newPassword) && /[0-9]/.test(newPassword) && /[!@#$%^&*]/.test(newPassword)) {
			passwordStrength = 'strong';
		} else {
			passwordStrength = 'medium';
		}
	}

	// Check if passwords match
	$: passwordsMatch = newPassword === confirmNewPassword && confirmNewPassword.length > 0;
	$: passwordsDontMatch = confirmNewPassword.length > 0 && newPassword !== confirmNewPassword;

	async function handleChangePassword() {
		if (!currentPassword || !newPassword || !confirmNewPassword) {
			passwordChangeError = 'Please fill in all fields';
			return;
		}

		if (newPassword !== confirmNewPassword) {
			passwordChangeError = 'New passwords do not match';
			return;
		}

		if (newPassword.length < 8) {
			passwordChangeError = 'New password must be at least 8 characters';
			return;
		}

		passwordChangeLoading = true;
		passwordChangeError = '';
		passwordChangeSuccess = false;

		const result = await auth.changePassword(currentPassword, newPassword);

		if (result.success) {
			passwordChangeSuccess = true;
			// Reset form
			currentPassword = '';
			newPassword = '';
			confirmNewPassword = '';
			// Hide form after 2 seconds
			setTimeout(() => {
				showChangePassword = false;
				passwordChangeSuccess = false;
			}, 2000);
		} else {
			passwordChangeError = result.error || 'Failed to change password. Please try again.';
		}

		passwordChangeLoading = false;
	}

	function toggleChangePassword() {
		showChangePassword = !showChangePassword;
		// Reset form when toggling
		if (showChangePassword) {
			currentPassword = '';
			newPassword = '';
			confirmNewPassword = '';
			passwordChangeError = '';
			passwordChangeSuccess = false;
		}
	}

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
							} else if (tabParam === 'subscription') {
								activeTab = 'subscription';
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

	function switchTab(tab: 'dashboard' | 'profile' | 'subscription' | 'advertiser') {
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
				<a href="/auth/login" class="btn btn-primary">Log In</a>
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
				class="tab-button {activeTab === 'subscription' ? 'active' : ''}"
				on:click={() => switchTab('subscription')}
			>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
				</svg>
				Subscription
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

		<!-- Temp Password Security Banner -->
		{#if user?.temp_password_active}
		<div class="temp-password-banner">
			<div class="banner-content">
				<div class="banner-icon">🔐</div>
				<div class="banner-text">
					<h3>Secure Your Account</h3>
					<p>You're currently using a temporary password. For your security, please create a personal password.</p>
				</div>
				<button 
					class="btn btn-primary banner-btn"
					on:click={() => { activeTab = 'profile'; showChangePassword = true; }}
				>
					Change Password Now
				</button>
			</div>
		</div>
		{/if}

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
				</div>

				<!-- Change Password Section -->
				<div class="change-password-section profile-info ">
					<div class="section-header">
						<h2>Security</h2>
						<button 
							class="btn btn-outline" 
							on:click={toggleChangePassword}
						>
							{showChangePassword ? 'Cancel' : 'Change Password'}
						</button>
					</div>

					{#if showChangePassword}
						<div class="change-password-form">
							{#if passwordChangeSuccess}
								<div class="success-message">
									<div class="success-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
											<polyline points="22,4 12,14.01 9,11.01"/>
										</svg>
									</div>
									<p>Password changed successfully!</p>
								</div>
							{:else}
								{#if passwordChangeError}
									<div class="error-message">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<circle cx="12" cy="12" r="10"/>
											<line x1="12" y1="8" x2="12" y2="12"/>
											<line x1="12" y1="16" x2="12.01" y2="16"/>
										</svg>
										<span>{passwordChangeError}</span>
									</div>
								{/if}

								<form on:submit|preventDefault={handleChangePassword}>
									<div class="form-group">
										<label for="currentPassword">Current Password</label>
										<input
											id="currentPassword"
											type="password"
											bind:value={currentPassword}
											placeholder="Enter your current password"
											disabled={passwordChangeLoading}
											required
										/>
									</div>

									<div class="form-group">
										<label for="newPassword">New Password</label>
										<input
											id="newPassword"
											type="password"
											bind:value={newPassword}
											placeholder="Enter new password (min 8 characters)"
											disabled={passwordChangeLoading}
											required
											minlength="8"
										/>
										
										<!-- Password Strength Indicator -->
										{#if newPassword.length > 0}
											<div class="password-strength">
												<div class="strength-bars">
													<div class="bar {passwordStrength === 'weak' || passwordStrength === 'medium' || passwordStrength === 'strong' ? 'weak' : ''}"></div>
													<div class="bar {passwordStrength === 'medium' || passwordStrength === 'strong' ? 'medium' : ''}"></div>
													<div class="bar {passwordStrength === 'strong' ? 'strong' : ''}"></div>
												</div>
												<span class="strength-label {passwordStrength}">
													{passwordStrength === 'weak' ? 'Weak' : passwordStrength === 'medium' ? 'Medium' : 'Strong'}
												</span>
											</div>
										{/if}
									</div>

									<div class="form-group">
										<label for="confirmNewPassword">Confirm New Password</label>
										<input
											id="confirmNewPassword"
											type="password"
											bind:value={confirmNewPassword}
											placeholder="Confirm your new password"
											disabled={passwordChangeLoading}
											class:match={passwordsMatch}
											class:no-match={passwordsDontMatch}
											required
										/>
										
										{#if passwordsMatch}
											<div class="validation-message success">
												<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
													<polyline points="20,6 9,17 4,12"/>
												</svg>
												Passwords match
											</div>
										{:else if passwordsDontMatch}
											<div class="validation-message error">
												<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
													<line x1="18" y1="6" x2="6" y2="18"/>
													<line x1="6" y1="6" x2="18" y2="18"/>
												</svg>
												Passwords don't match
											</div>
										{/if}
									</div>

									<button 
										type="submit" 
										class="btn btn-primary"
										disabled={passwordChangeLoading || !passwordsMatch || newPassword.length < 8}
									>
										{#if passwordChangeLoading}
											<span class="spinner"></span>
											Changing Password...
										{:else}
											Change Password
										{/if}
									</button>
								</form>
							{/if}
						</div>
					{/if}
				</div>
			</div>
	{:else if activeTab === 'subscription'}
		<!-- Subscription Tab Content -->
		<div class="tab-content">
			<SubscriptionManagement embedded={true} />
			
			<!-- Stripe Customer Portal Access -->
			<div class="stripe-portal-section glass">
				<div class="portal-content">
					<div class="portal-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<rect x="1" y="4" width="22" height="16" rx="2" ry="2"/>
							<line x1="1" y1="10" x2="23" y2="10"/>
						</svg>
					</div>
					<h3>Manage Billing & Payment Methods</h3>
					<p>Access the Stripe Customer Portal to update your payment methods, view invoices, and manage your billing details.</p>
					<a 
						href="https://billing.stripe.com/p/login/bJe00jcaW9wU3Vf9gU2VG00" 
						target="_blank" 
						rel="noopener noreferrer"
						class="btn btn-stripe"
					>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>
							<polyline points="15 3 21 3 21 9"/>
							<line x1="10" y1="14" x2="21" y2="3"/>
						</svg>
						Open Stripe Portal
					</a>
					<p class="portal-note">Opens in a new tab for security</p>
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
		padding: 6rem 0 0 0;
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
		background: var(--primary-gold-light);
		color: var(--primary-bom-dark);
		cursor: pointer;
		transition: all 0.3s ease;
		font-size: 0.875rem;
		font-weight: 500;
	}

	.tab-button:hover {
		background: var(--primary-color);
		color: var(--primary-bom-dark);
		transform: translateY(-2px);
	}

	.tab-button.active {
		color: var(--primary-bom-dark);
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

	/* Change Password Section */
	.change-password-section {
		background: var(--bg-secondary);
		border-radius: 20px;
		padding: 2rem;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
		margin-top: 2rem;
		max-width: 600px;
		margin-left: auto;
		margin-right: auto;
	}

	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1.5rem;
	}

	.section-header h2 {
		font-size: 1.5rem;
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.change-password-form {
		margin-top: 1.5rem;
		animation: slideDown 0.3s ease-out;
	}

	@keyframes slideDown {
		from {
			opacity: 0;
			transform: translateY(-10px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	.change-password-form .form-group {
		margin-bottom: 1.5rem;
	}

	.change-password-form label {
		display: block;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
		font-size: 0.875rem;
	}

	.change-password-form input {
		width: 100%;
		padding: 0.75rem 1rem;
		border: 2px solid var(--border-color);
		border-radius: 8px;
		font-size: 1rem;
		background: var(--bg-primary);
		color: var(--text-primary);
		transition: all 0.2s ease;
	}

	.change-password-form input:focus {
		outline: none;
		border-color: var(--primary-color);
		box-shadow: 0 0 0 3px rgba(147, 51, 234, 0.1);
	}

	.change-password-form input:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.change-password-form input.match {
		border-color: #10b981;
	}

	.change-password-form input.no-match {
		border-color: #ef4444;
	}

	/* Password Strength Indicator */
	.password-strength {
		margin-top: 0.5rem;
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.strength-bars {
		display: flex;
		gap: 0.25rem;
		flex: 1;
	}

	.strength-bars .bar {
		height: 4px;
		flex: 1;
		background: var(--border-color);
		border-radius: 2px;
		transition: all 0.3s ease;
	}

	.strength-bars .bar.weak {
		background: #ef4444;
	}

	.strength-bars .bar.medium {
		background: #f59e0b;
	}

	.strength-bars .bar.strong {
		background: #10b981;
	}

	.strength-label {
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
	}

	.strength-label.weak {
		color: #ef4444;
	}

	.strength-label.medium {
		color: #f59e0b;
	}

	.strength-label.strong {
		color: #10b981;
	}

	/* Validation Messages */
	.validation-message {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-top: 0.5rem;
		font-size: 0.875rem;
		font-weight: 500;
	}

	.validation-message svg {
		width: 16px;
		height: 16px;
	}

	.validation-message.success {
		color: #10b981;
	}

	.validation-message.error {
		color: #ef4444;
	}

	/* Success Message */
	.success-message {
		text-align: center;
		padding: 2rem;
		animation: fadeIn 0.3s ease-out;
	}

	@keyframes fadeIn {
		from {
			opacity: 0;
			transform: scale(0.9);
		}
		to {
			opacity: 1;
			transform: scale(1);
		}
	}

	.success-icon {
		width: 64px;
		height: 64px;
		margin: 0 auto 1rem;
		color: #10b981;
		animation: pulse 0.6s ease-out;
	}

	@keyframes pulse {
		0%, 100% {
			transform: scale(1);
		}
		50% {
			transform: scale(1.1);
		}
	}

	.success-icon svg {
		width: 100%;
		height: 100%;
	}

	.success-message p {
		font-size: 1.125rem;
		font-weight: 600;
		color: #10b981;
		margin: 0;
	}

	/* Error Message */
	.error-message {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 1rem;
		background: rgba(239, 68, 68, 0.1);
		border: 1px solid #ef4444;
		border-radius: 8px;
		color: #ef4444;
		margin-bottom: 1.5rem;
		font-size: 0.875rem;
	}

	.error-message svg {
		width: 20px;
		height: 20px;
		flex-shrink: 0;
	}

	/* Submit Button Spinner */
	.spinner {
		display: inline-block;
		width: 14px;
		height: 14px;
		border: 2px solid rgba(255, 255, 255, 0.3);
		border-top-color: white;
		border-radius: 50%;
		animation: spin 0.6s linear infinite;
		margin-right: 0.5rem;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	/* Profile Actions */
	.subscription-section {
		margin-top: 2rem;
		max-width: 600px;
		margin-left: auto;
		margin-right: auto;
	}

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
		background: var(--primary-gold-light);
		color: var(--primary-bom-dark);
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
		color: white;
	}

	.btn-primary:hover {
		background: var(--secondary-gradient);
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(var(--primary-rgb), 0.3);
		color: var(--primary-bom-dark);
	}

	/* Stripe Portal Section */
	.stripe-portal-section {
		margin-top: 2rem;
		max-width: 600px;
		margin-left: auto;
		margin-right: auto;
		border-radius: 20px;
		padding: 2rem;
		background: linear-gradient(145deg, rgba(99, 91, 255, 0.1), rgba(99, 91, 255, 0.05));
		border: 1px solid rgba(99, 91, 255, 0.2);
	}

	.portal-content {
		text-align: center;
	}

	.portal-icon {
		width: 64px;
		height: 64px;
		margin: 0 auto 1.5rem;
		background: linear-gradient(135deg, #635bff, #8b7fff);
		border-radius: 16px;
		display: flex;
		align-items: center;
		justify-content: center;
		color: white;
	}

	.portal-icon svg {
		width: 32px;
		height: 32px;
	}

	.portal-content h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.75rem;
	}

	.portal-content p {
		color: var(--text-secondary);
		font-size: 0.9rem;
		line-height: 1.6;
		margin-bottom: 1.5rem;
	}

	.btn-stripe {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.875rem 1.75rem;
		background: linear-gradient(135deg, #635bff, #8b7fff);
		color: white;
		border: none;
		border-radius: 12px;
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.3s ease;
		text-decoration: none;
	}

	.btn-stripe:hover {
		background: linear-gradient(135deg, #5349ff, #7a6aff);
		transform: translateY(-2px);
		box-shadow: 0 8px 20px rgba(99, 91, 255, 0.3);
	}

	.btn-stripe svg {
		width: 18px;
		height: 18px;
	}

	.portal-note {
		font-size: 0.75rem !important;
		color: var(--text-muted) !important;
		margin-top: 1rem !important;
		margin-bottom: 0 !important;
	}

	/* Temp Password Security Banner */
	.temp-password-banner {
		background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
		border: 2px solid #f59e0b;
		border-radius: 16px;
		margin-bottom: 1.5rem;
		padding: 1.5rem;
		animation: pulse-glow 2s ease-in-out infinite;
	}

	@keyframes pulse-glow {
		0%, 100% {
			box-shadow: 0 0 5px rgba(245, 158, 11, 0.3);
		}
		50% {
			box-shadow: 0 0 20px rgba(245, 158, 11, 0.5);
		}
	}

	.temp-password-banner .banner-content {
		display: flex;
		align-items: center;
		gap: 1.5rem;
		flex-wrap: wrap;
	}

	.temp-password-banner .banner-icon {
		font-size: 2.5rem;
		flex-shrink: 0;
	}

	.temp-password-banner .banner-text {
		flex: 1;
		min-width: 200px;
	}

	.temp-password-banner .banner-text h3 {
		color: #92400e;
		font-size: 1.25rem;
		font-weight: 700;
		margin: 0 0 0.25rem 0;
	}

	.temp-password-banner .banner-text p {
		color: #78350f;
		margin: 0;
		font-size: 0.875rem;
	}

	.temp-password-banner .banner-btn {
		background: #f59e0b;
		border: none;
		color: white;
		padding: 0.75rem 1.5rem;
		font-weight: 600;
		border-radius: 8px;
		cursor: pointer;
		transition: all 0.2s;
		flex-shrink: 0;
	}

	.temp-password-banner .banner-btn:hover {
		background: #d97706;
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(245, 158, 11, 0.4);
	}

	@media (max-width: 640px) {
		.temp-password-banner .banner-content {
			flex-direction: column;
			text-align: center;
		}

		.temp-password-banner .banner-btn {
			width: 100%;
		}
	}
</style> 
