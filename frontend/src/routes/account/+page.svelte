<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import { subscriptionService, subscriptionUtils, type Subscription } from '$lib/subscription';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let user: any = null;
	let subscription: Subscription | null = null;
	let loading = true;
	let error = '';
	let isAuthenticated = false;

	// Subscribe to auth store
	auth.subscribe(state => {
		isAuthenticated = state.isAuthenticated;
		user = state.user;
	});

	onMount(async () => {
		if (!isAuthenticated) {
			goto('/login');
			return;
		}

		await loadAccountData();
	});

	const loadAccountData = async () => {
		try {
			loading = true;
			
			// Load current subscription
			try {
				const response = await subscriptionService.getCurrentSubscription();
				subscription = response.subscription;
			} catch (err) {
				// User might not have a subscription, which is okay
				subscription = null;
			}
		} catch (err) {
			error = 'Failed to load account data';
			console.error('Error loading account data:', err);
		} finally {
			loading = false;
		}
	};

	const handleManageSubscription = async () => {
		try {
			const returnUrl = `${window.location.origin}/account`;
			const response = await subscriptionService.createCustomerPortalSession(returnUrl);
			
			if (response.url) {
				window.location.href = response.url;
			} else {
				showToast('Failed to open customer portal', 'error');
			}
		} catch (err) {
			showToast('Failed to open customer portal', 'error');
			console.error('Error opening customer portal:', err);
		}
	};

	const handleUpgrade = () => {
		goto('/subscription');
	};

	const handleViewBilling = () => {
		goto('/account/billing');
	};

	const handleEditProfile = () => {
		goto('/account/profile');
	};

	const handleAccountSettings = () => {
		goto('/account/settings');
	};
</script>

<svelte:head>
	<title>My Account - BOME</title>
	<meta name="description" content="Manage your BOME account, subscription, and billing information" />
</svelte:head>

<div class="account-page">
	<div class="container">
		<header class="page-header">
			<h1>My Account</h1>
			<p>Manage your profile, subscription, and account settings</p>
		</header>

		{#if loading}
			<div class="loading-container">
				<LoadingSpinner />
				<p>Loading your account...</p>
			</div>
		{:else if error}
			<div class="error-container">
				<p class="error-message">{error}</p>
				<button class="btn btn-primary" on:click={loadAccountData}>
					Try Again
				</button>
			</div>
		{:else}
			<div class="account-content">
				<!-- Profile Section -->
				<div class="account-section">
					<div class="section-header">
						<h2>Profile Information</h2>
						<button class="btn btn-outline" on:click={handleEditProfile}>
							Edit Profile
						</button>
					</div>
					<div class="profile-card">
						<div class="profile-avatar">
							<div class="avatar-placeholder">
								{user?.name?.charAt(0)?.toUpperCase() || 'U'}
							</div>
						</div>
						<div class="profile-info">
							<h3>{user?.name || 'User'}</h3>
							<p class="email">{user?.email || 'email@example.com'}</p>
							<div class="profile-stats">
								<div class="stat">
									<span class="stat-label">Member since</span>
									<span class="stat-value">
										{user?.createdAt ? new Date(user.createdAt).toLocaleDateString() : 'N/A'}
									</span>
								</div>
							</div>
						</div>
					</div>
				</div>

				<!-- Subscription Section -->
				<div class="account-section">
					<div class="section-header">
						<h2>Subscription</h2>
						{#if subscription}
							<button class="btn btn-outline" on:click={handleManageSubscription}>
								Manage Subscription
							</button>
						{:else}
							<button class="btn btn-primary" on:click={handleUpgrade}>
								Upgrade to Premium
							</button>
						{/if}
					</div>
					<div class="subscription-card">
						{#if subscription}
							<div class="subscription-active">
								<div class="subscription-status">
									<span class="status-badge" style="background-color: {subscriptionUtils.getStatusColor(subscription.status)}">
										{subscriptionUtils.getStatusText(subscription.status)}
									</span>
								</div>
								<div class="subscription-details">
									<div class="detail-row">
										<span class="label">Status:</span>
										<span class="value">{subscriptionUtils.getStatusText(subscription.status)}</span>
									</div>
									<div class="detail-row">
										<span class="label">Next billing:</span>
										<span class="value">{subscriptionUtils.formatDate(subscription.currentPeriodEnd)}</span>
									</div>
									{#if subscription.cancelAtPeriodEnd}
										<div class="detail-row">
											<span class="label">Cancellation:</span>
											<span class="value warning">Will cancel at period end</span>
										</div>
									{/if}
								</div>
							</div>
						{:else}
							<div class="subscription-inactive">
								<div class="inactive-icon">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<circle cx="12" cy="12" r="10"></circle>
										<path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"></path>
										<line x1="12" y1="17" x2="12.01" y2="17"></line>
									</svg>
								</div>
								<h3>No Active Subscription</h3>
								<p>Upgrade to premium to access exclusive content and features</p>
								<button class="btn btn-primary" on:click={handleUpgrade}>
									View Plans
								</button>
							</div>
						{/if}
					</div>
				</div>

				<!-- Quick Actions -->
				<div class="account-section">
					<div class="section-header">
						<h2>Quick Actions</h2>
					</div>
					<div class="actions-grid">
						<button class="action-card" on:click={handleViewBilling}>
							<div class="action-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect>
									<line x1="8" y1="21" x2="16" y2="21"></line>
									<line x1="12" y1="17" x2="12" y2="21"></line>
								</svg>
							</div>
							<div class="action-content">
								<h3>Billing History</h3>
								<p>View invoices and payment history</p>
							</div>
						</button>

						<button class="action-card" on:click={handleAccountSettings}>
							<div class="action-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<circle cx="12" cy="12" r="3"></circle>
									<path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1 1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
								</svg>
							</div>
							<div class="action-content">
								<h3>Account Settings</h3>
								<p>Manage preferences and security</p>
							</div>
						</button>

						<button class="action-card" on:click={() => goto('/videos')}>
							<div class="action-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<polygon points="23 7 16 12 23 17 23 7"></polygon>
									<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
								</svg>
							</div>
							<div class="action-content">
								<h3>Watch Videos</h3>
								<p>Browse our video library</p>
							</div>
						</button>

						<button class="action-card" on:click={() => goto('/videos/favorites')}>
							<div class="action-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path>
								</svg>
							</div>
							<div class="action-content">
								<h3>Favorites</h3>
								<p>Your saved videos</p>
							</div>
						</button>
					</div>
				</div>

				<!-- Account Actions -->
				<div class="account-actions">
					<button class="action-card" on:click={() => goto('/account/profile')}>
						<div class="action-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
								<circle cx="12" cy="7" r="4"></circle>
							</svg>
						</div>
						<div class="action-content">
							<h3>Edit Profile</h3>
							<p>Update your personal information</p>
						</div>
						<div class="action-arrow">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M9 18l6-6-6-6"></path>
							</svg>
						</div>
					</button>

					<button class="action-card" on:click={() => goto('/account/settings')}>
						<div class="action-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<circle cx="12" cy="12" r="3"></circle>
								<path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1 1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
							</svg>
						</div>
						<div class="action-content">
							<h3>Account Settings</h3>
							<p>Manage preferences and privacy</p>
						</div>
						<div class="action-arrow">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M9 18l6-6-6-6"></path>
							</svg>
						</div>
					</button>

					<button class="action-card" on:click={() => goto('/account/subscription')}>
						<div class="action-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect>
								<line x1="8" y1="21" x2="16" y2="21"></line>
								<line x1="12" y1="17" x2="12" y2="21"></line>
							</svg>
						</div>
						<div class="action-content">
							<h3>Manage Subscription</h3>
							<p>View and update your subscription</p>
						</div>
						<div class="action-arrow">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M9 18l6-6-6-6"></path>
							</svg>
						</div>
					</button>

					<button class="action-card" on:click={() => goto('/account/billing')}>
						<div class="action-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<rect x="1" y="4" width="22" height="16" rx="2" ry="2"></rect>
								<line x1="1" y1="10" x2="23" y2="10"></line>
							</svg>
						</div>
						<div class="action-content">
							<h3>Billing History</h3>
							<p>View invoices and payment methods</p>
						</div>
						<div class="action-arrow">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M9 18l6-6-6-6"></path>
							</svg>
						</div>
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	.account-page {
		min-height: 100vh;
		padding: 2rem 0;
		background: var(--bg-gradient);
	}

	.container {
		max-width: 1200px;
		margin: 0 auto;
		padding: 0 1rem;
	}

	.page-header {
		text-align: center;
		margin-bottom: 3rem;
	}

	.page-header h1 {
		font-size: 2.5rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.page-header p {
		font-size: 1.1rem;
		color: var(--text-secondary);
	}

	.loading-container,
	.error-container {
		text-align: center;
		padding: 3rem 0;
	}

	.error-message {
		color: var(--error-text);
		margin-bottom: 1rem;
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

	.profile-card {
		display: flex;
		align-items: center;
		gap: 1.5rem;
	}

	.profile-avatar {
		flex-shrink: 0;
	}

	.avatar-placeholder {
		width: 80px;
		height: 80px;
		border-radius: 50%;
		background: var(--primary-color);
		color: white;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 2rem;
		font-weight: 700;
	}

	.profile-info h3 {
		font-size: 1.5rem;
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 0.5rem 0;
	}

	.email {
		color: var(--text-secondary);
		margin-bottom: 1rem;
	}

	.profile-stats {
		display: flex;
		gap: 2rem;
	}

	.stat {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.stat-label {
		font-size: 0.875rem;
		color: var(--text-secondary);
	}

	.stat-value {
		font-weight: 600;
		color: var(--text-primary);
	}

	.subscription-card {
		min-height: 150px;
		display: flex;
		align-items: center;
	}

	.subscription-active {
		width: 100%;
	}

	.subscription-status {
		margin-bottom: 1rem;
	}

	.status-badge {
		display: inline-block;
		padding: 0.5rem 1rem;
		border-radius: 20px;
		font-size: 0.875rem;
		font-weight: 600;
		color: white;
	}

	.subscription-details {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.detail-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.detail-row .label {
		color: var(--text-secondary);
	}

	.detail-row .value {
		font-weight: 600;
		color: var(--text-primary);
	}

	.detail-row .value.warning {
		color: var(--warning-text);
	}

	.subscription-inactive {
		text-align: center;
		width: 100%;
		padding: 2rem 0;
	}

	.inactive-icon {
		width: 60px;
		height: 60px;
		margin: 0 auto 1rem;
		color: var(--text-secondary);
	}

	.inactive-icon svg {
		width: 100%;
		height: 100%;
	}

	.subscription-inactive h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.subscription-inactive p {
		color: var(--text-secondary);
		margin-bottom: 1.5rem;
	}

	.actions-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 1rem;
	}

	.action-card {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1.5rem;
		background: var(--bg-secondary);
		border: 1px solid var(--border-color);
		border-radius: 16px;
		transition: all 0.3s ease;
		cursor: pointer;
		text-align: left;
	}

	.action-card:hover {
		transform: translateY(-2px);
		box-shadow: var(--neumorphic-shadow-hover);
	}

	.action-icon {
		width: 40px;
		height: 40px;
		color: var(--primary-color);
		flex-shrink: 0;
	}

	.action-icon svg {
		width: 100%;
		height: 100%;
	}

	.action-content h3 {
		font-size: 1rem;
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 0.25rem 0;
	}

	.action-content p {
		font-size: 0.875rem;
		color: var(--text-secondary);
		margin: 0;
	}

	@media (max-width: 768px) {
		.profile-card {
			flex-direction: column;
			text-align: center;
		}

		.profile-stats {
			justify-content: center;
		}

		.section-header {
			flex-direction: column;
			gap: 1rem;
			align-items: stretch;
		}

		.detail-row {
			flex-direction: column;
			align-items: flex-start;
			gap: 0.25rem;
		}

		.actions-grid {
			grid-template-columns: 1fr;
		}
	}

	.account-actions {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: 1rem;
		margin-bottom: 2rem;
	}

	.action-card {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1.5rem;
		background: var(--bg-secondary);
		border: 1px solid var(--border-color);
		border-radius: 16px;
		transition: all 0.3s ease;
		cursor: pointer;
		text-align: left;
	}

	.action-card:hover {
		transform: translateY(-2px);
		box-shadow: var(--neumorphic-shadow-hover);
		border-color: var(--primary-color);
	}

	.action-icon {
		width: 40px;
		height: 40px;
		color: var(--primary-color);
		flex-shrink: 0;
	}

	.action-icon svg {
		width: 100%;
		height: 100%;
	}

	.action-content {
		flex: 1;
	}

	.action-content h3 {
		font-size: 1rem;
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 0.25rem 0;
	}

	.action-content p {
		font-size: 0.875rem;
		color: var(--text-secondary);
		margin: 0;
	}

	.action-arrow {
		width: 20px;
		height: 20px;
		color: var(--text-secondary);
		flex-shrink: 0;
	}

	.action-arrow svg {
		width: 100%;
		height: 100%;
	}
</style> 