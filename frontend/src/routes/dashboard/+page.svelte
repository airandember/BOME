<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { auth } from '$lib/auth';
	import { advertiserStore } from '$lib/stores/advertiser';
	import { videoService } from '$lib/video';
	import { subscriptionService, subscriptionUtils, type Subscription } from '$lib/subscription';
	import { showToast } from '$lib/toast';
	import { MOCK_DASHBOARD_DATA } from '$lib/mockData';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import Navigation from '$lib/components/Navigation.svelte';
	import Footer from '$lib/components/Footer.svelte';

	interface User {
		id: number;
		email: string;
		firstName: string;
		lastName: string;
		role: string;
		emailVerified: boolean;
	}

	interface Video {
		id: number;
		title: string;
		description: string;
		thumbnail: string;
		duration: string;
		views: number;
		likes: number;
		createdAt: string;
		category: string;
	}

	interface Activity {
		type: 'watched' | 'liked' | 'commented' | 'favorited';
		video: Video;
		timestamp: string;
	}

	interface UserStats {
		videosWatched: number;
		totalWatchTime: string;
		favoriteVideos: number;
		commentsPosted: number;
		joinedDate: string;
		subscriptionStatus: string;
	}

	let user: any = null;
	let isAuthenticated = false;
	let loading = true;
	let error = '';
	let dashboardData: any = null;
	let isApprovedAdvertiser = false;

	// Tab state
	let activeTab: 'dashboard' | 'account' | 'advertiser' = 'dashboard';

	// Advertiser state
	let advertiserAccount: any = null;
	let advertiserStatus: 'none' | 'pending' | 'approved' | 'rejected' | 'cancelled' = 'none';

	// Account state
	let subscription: Subscription | null = null;

	// Dashboard data
	let stats = {
		totalWatchTime: 0,
		videosWatched: 0,
		favoriteVideos: 0,
		completedSeries: 0
	};
	let recentActivity: any[] = [];
	let recommendedVideos: any[] = [];
	let favoriteVideos: any[] = [];
	let continueWatching: any[] = [];

	onMount(() => {
		let unsubscribe: any = null;

		async function initializeDashboard() {
			try {
				// Subscribe to auth state changes first
				unsubscribe = auth.subscribe((state) => {
					user = state.user;
					isAuthenticated = state.isAuthenticated;
					
					// If we have auth state, proceed with loading
					if (state.isAuthenticated && state.user) {
						loadDashboardData();
						
						// Check URL parameters for fromAdvertise
						const urlParams = new URLSearchParams($page.url.search);
						const fromParam = urlParams.get('from');
						checkAdvertiserStatus(fromParam === 'advertise');
					} else if (state.isAuthenticated === false) {
						// Only redirect if explicitly not authenticated
						goto('/login');
						return;
					}
				});

				// Initialize auth
				await auth.initialize();
				
				// Check URL parameters for tab and redirect handling
				const urlParams = new URLSearchParams($page.url.search);
				const tabParam = urlParams.get('tab');
				
				if (tabParam === 'advertiser') {
					activeTab = 'advertiser';
				} else if (tabParam === 'account') {
					activeTab = 'account';
				} else {
					activeTab = 'dashboard';
				}
				
				// Clean up URL parameters after processing
				if (urlParams.has('tab') || urlParams.has('from')) {
					const newUrl = new URL($page.url);
					newUrl.search = '';
					window.history.replaceState({}, '', newUrl.pathname);
				}
				
				// Set loading to false after a reasonable timeout
				setTimeout(() => {
					loading = false;
				}, 2000);
				
			} catch (error) {
				console.error('Error loading dashboard:', error);
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

	async function loadDashboardData() {
		try {
			// Load dashboard data - use mock data for development
			stats = MOCK_DASHBOARD_DATA.stats;
			recentActivity = MOCK_DASHBOARD_DATA.recentActivity;
			recommendedVideos = MOCK_DASHBOARD_DATA.recommendedVideos;
			favoriteVideos = MOCK_DASHBOARD_DATA.favoriteVideos;
			continueWatching = MOCK_DASHBOARD_DATA.continueWatching;

			// Load subscription data for account tab
			await loadAccountData();

			console.log('Dashboard data loaded successfully');
		} catch (err) {
			console.error('Error loading dashboard data:', err);
			// Don't throw error, just log it and use empty data
			stats = { totalWatchTime: 0, videosWatched: 0, favoriteVideos: 0, completedSeries: 0 };
			recentActivity = [];
			recommendedVideos = [];
			favoriteVideos = [];
			continueWatching = [];
		}
	}

	async function loadAccountData() {
		try {
			// Load current subscription
			const response = await subscriptionService.getCurrentSubscription();
			subscription = response.subscription;
		} catch (err) {
			// User might not have a subscription, which is okay
			subscription = null;
		}
	}

	async function checkAdvertiserStatus(fromAdvertise = false) {
		try {
			// Get the user's advertiser account from the store
			const account = await advertiserStore.getByUserId(user?.id || 0);
			
			if (account) {
				advertiserAccount = account;
				advertiserStatus = account.status;
				isApprovedAdvertiser = account.status === 'approved';
				
				if (fromAdvertise) {
					// User just completed registration, switch to advertiser tab
					activeTab = 'advertiser';
				}
			} else {
				// No advertiser account found
				advertiserStatus = 'none';
				isApprovedAdvertiser = false;
			}
		} catch (error) {
			console.error('Failed to check advertiser status:', error);
			advertiserStatus = 'none';
			isApprovedAdvertiser = false;
		}
	}

	function formatDate(dateString: string) {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	function formatNumber(num: number) {
		return new Intl.NumberFormat('en-US').format(num);
	}

	function getActivityIcon(type: string) {
		switch (type) {
			case 'watched':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<polygon points="23,7 16,12 23,17 23,7"></polygon>
					<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
				</svg>`;
			case 'liked':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path>
				</svg>`;
			case 'commented':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
				</svg>`;
			case 'favorited':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<polygon points="12,2 15.09,8.26 22,9.27 17,14.14 18.18,21.02 12,17.77 5.82,21.02 7,14.14 2,9.27 8.91,8.26 12,2"></polygon>
				</svg>`;
			default:
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="12" cy="12" r="10"></circle>
					<line x1="12" y1="16" x2="12" y2="12"></line>
					<line x1="12" y1="8" x2="12.01" y2="8"></line>
				</svg>`;
		}
	}

	function getActivityText(activity: Activity) {
		switch (activity.type) {
			case 'watched':
				return `Watched "${activity.video.title}"`;
			case 'liked':
				return `Liked "${activity.video.title}"`;
			case 'commented':
				return `Commented on "${activity.video.title}"`;
			case 'favorited':
				return `Added "${activity.video.title}" to favorites`;
			default:
				return `Interacted with "${activity.video.title}"`;
		}
	}

	function formatCurrency(amount: number) {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(amount);
	}

	function switchTab(tab: 'dashboard' | 'account' | 'advertiser') {
		activeTab = tab;
	}

	function handleAdvertiseClick() {
		goto('/advertise');
	}

	function handleCreateCampaign() {
		goto('/advertise/campaigns');
	}

	function handleManageCampaigns() {
		goto('/advertise/campaigns');
	}

	function handleReapply() {
		goto('/advertise');
	}

	const handleManageSubscription = async () => {
		try {
			const returnUrl = `${window.location.origin}/dashboard?tab=account`;
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
				class="tab-button {activeTab === 'account' ? 'active' : ''}"
				on:click={() => switchTab('account')}
			>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
					<circle cx="12" cy="12" r="3"/>
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
				Advertiser Information
				{#if advertiserStatus === 'pending'}
					<span class="status-indicator pending">Pending</span>
				{:else if advertiserStatus === 'approved'}
					<span class="status-indicator approved">Approved</span>
				{:else if advertiserStatus === 'rejected'}
					<span class="status-indicator rejected">Rejected</span>
				{/if}
			</button>
		</div>

		{#if activeTab === 'dashboard'}
			<!-- Dashboard Tab Content -->
			<div class="tab-content">
				<!-- Welcome Header -->
				<div class="welcome-section glass">
					<div class="welcome-content">
						<h1>Welcome back, {user?.firstName || 'User'}!</h1>
						<p>Continue your journey exploring Book of Mormon evidences</p>
					</div>
					<div class="user-avatar-large">
						{user?.firstName?.charAt(0) || 'U'}{user?.lastName?.charAt(0) || ''}
					</div>
				</div>

				<!-- Stats Overview -->
				<div class="stats-grid">
					<div class="stat-card glass">
						<div class="stat-icon videos">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polygon points="23,7 16,12 23,17 23,7"></polygon>
								<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
							</svg>
						</div>
						<div class="stat-content">
							<div class="stat-value">{stats.videosWatched}</div>
							<div class="stat-label">Videos Watched</div>
						</div>
					</div>

					<div class="stat-card glass">
						<div class="stat-icon time">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<circle cx="12" cy="12" r="10"></circle>
								<polyline points="12,6 12,12 16,14"></polyline>
							</svg>
						</div>
						<div class="stat-content">
							<div class="stat-value">{stats.totalWatchTime}</div>
							<div class="stat-label">Watch Time</div>
						</div>
					</div>

					<div class="stat-card glass">
						<div class="stat-icon favorites">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polygon points="12,2 15.09,8.26 22,9.27 17,14.14 18.18,21.02 12,17.77 5.82,21.02 7,14.14 2,9.27 8.91,8.26 12,2"></polygon>
							</svg>
						</div>
						<div class="stat-content">
							<div class="stat-value">{stats.favoriteVideos}</div>
							<div class="stat-label">Favorites</div>
						</div>
					</div>

					<div class="stat-card glass">
						<div class="stat-icon subscription">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M20 7H4a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2z"></path>
								<polyline points="16,21 12,17 8,21"></polyline>
								<polyline points="12,17 12,3"></polyline>
							</svg>
						</div>
						<div class="stat-content">
							<div class="stat-value">{stats.completedSeries}</div>
							<div class="stat-label">Completed Series</div>
						</div>
					</div>
				</div>

				<!-- Main Content Grid -->
				<div class="content-grid">
					<!-- Continue Watching -->
					{#if continueWatching.length > 0}
						<div class="section-card glass">
							<div class="section-header">
								<h2>Continue Watching</h2>
								<a href="/videos" class="view-all-link">View All</a>
							</div>
							<div class="video-grid">
								{#each continueWatching as video}
									<div class="video-card">
										<div class="video-thumbnail">
											<img src={video.thumbnailUrl} alt={video.title} />
											<div class="video-duration">{video.duration}</div>
											<div class="progress-bar">
												<div class="progress-fill" style="width: 65%"></div>
											</div>
										</div>
										<div class="video-info">
											<h3 class="video-title">{video.title}</h3>
											<p class="video-meta">{formatNumber(video.viewCount)} views • {video.category}</p>
										</div>
									</div>
								{/each}
							</div>
						</div>
					{/if}

					<!-- Recent Activity -->
					<div class="section-card glass">
						<div class="section-header">
							<h2>Recent Activity</h2>
						</div>
						<div class="activity-list">
							{#each recentActivity as activity}
								<div class="activity-item">
									<div class="activity-icon">
										{@html getActivityIcon(activity.type)}
									</div>
									<div class="activity-content">
										<div class="activity-text">{getActivityText(activity)}</div>
										<div class="activity-time">{activity.timestamp}</div>
									</div>
								</div>
							{/each}
						</div>
					</div>

					<!-- Recommended Videos -->
					<div class="section-card glass">
						<div class="section-header">
							<h2>Recommended for You</h2>
							<a href="/videos" class="view-all-link">View All</a>
						</div>
						<div class="video-grid">
							{#each recommendedVideos as video}
								<div class="video-card">
									<div class="video-thumbnail">
										<img src={video.thumbnailUrl} alt={video.title} />
										<div class="video-duration">{video.duration}</div>
									</div>
									<div class="video-info">
										<h3 class="video-title">{video.title}</h3>
										<p class="video-meta">{formatNumber(video.viewCount)} views • {video.category}</p>
									</div>
								</div>
							{/each}
						</div>
					</div>

					<!-- Favorite Videos -->
					<div class="section-card glass">
						<div class="section-header">
							<h2>Your Favorites</h2>
							<a href="/favorites" class="view-all-link">View All</a>
						</div>
						<div class="video-grid">
							{#each favoriteVideos as video}
								<div class="video-card">
									<div class="video-thumbnail">
										<img src={video.thumbnailUrl} alt={video.title} />
										<div class="video-duration">{video.duration}</div>
									</div>
									<div class="video-info">
										<h3 class="video-title">{video.title}</h3>
										<p class="video-meta">{formatNumber(video.viewCount)} views • {video.category}</p>
									</div>
								</div>
							{/each}
						</div>
					</div>
				</div>

				<!-- Quick Actions -->
				<div class="quick-actions glass">
					<h2>Quick Actions</h2>
					<div class="action-buttons">
						<a href="/videos" class="action-btn">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polygon points="23,7 16,12 23,17 23,7"></polygon>
								<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
							</svg>
							Browse Videos
						</a>
						<a href="/articles" class="action-btn">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
								<polyline points="14,2 14,8 20,8"/>
								<line x1="16" y1="13" x2="8" y2="13"/>
								<line x1="16" y1="17" x2="8" y2="17"/>
								<polyline points="10,9 9,9 8,9"/>
							</svg>
							Read Articles
						</a>
						<a href="/favorites" class="action-btn">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polygon points="12,2 15.09,8.26 22,9.27 17,14.14 18.18,21.02 12,17.77 5.82,21.02 7,14.14 2,9.27 8.91,8.26 12,2"></polygon>
							</svg>
							My Favorites
						</a>
					</div>
				</div>
			</div>
		{:else if activeTab === 'account'}
			<!-- Account Tab Content -->
			<div class="tab-content">
				<!-- Profile Section -->
				<div class="account-section glass">
					<div class="section-header">
						<h2>Profile Information</h2>
						<button class="secondary-button" on:click={() => goto('/account/profile')}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
								<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
							</svg>
							Edit Profile
						</button>
					</div>
					<div class="profile-card glass">
						<div class="profile-avatar">
							<div class="avatar-placeholder">
								{user?.firstName?.charAt(0)?.toUpperCase() || user?.name?.charAt(0)?.toUpperCase() || 'U'}
							</div>
						</div>
						<div class="profile-info">
							<h3>{user?.firstName ? `${user.firstName} ${user.lastName || ''}` : user?.name || 'User'}</h3>
							<p class="email">{user?.email || 'email@example.com'}</p>
							<div class="profile-stats">
								<div class="stat">
									<span class="stat-label">Member since</span>
									<span class="stat-value">
										{user?.createdAt ? new Date(user.createdAt).toLocaleDateString() : 'N/A'}
									</span>
								</div>
								<div class="stat">
									<span class="stat-label">Role</span>
									<span class="stat-value">{user?.role || 'Member'}</span>
								</div>
							</div>
						</div>
					</div>
				</div>

				<!-- Subscription Section -->
				<div class="account-section glass">
					<div class="section-header">
						<h2>Subscription</h2>
						{#if subscription}
							<button class="secondary-button" on:click={handleManageSubscription}>
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<circle cx="12" cy="12" r="3"/>
									<path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1 1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
								</svg>
								Manage Subscription
							</button>
						{:else}
							<button class="primary-button" on:click={() => goto('/subscription')}>
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
								</svg>
								Upgrade to Premium
							</button>
						{/if}
					</div>
					<div class="subscription-card glass">
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
										<circle cx="12" cy="12" r="10"/>
										<path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/>
										<line x1="12" y1="17" x2="12.01" y2="17"/>
									</svg>
								</div>
								<h3>No Active Subscription</h3>
								<p>Upgrade to premium to access exclusive content and features</p>
								<button class="primary-button" on:click={() => goto('/subscription')}>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
									</svg>
									View Plans
								</button>
							</div>
						{/if}
					</div>
				</div>

				<!-- Account Actions -->
				<div class="account-actions glass">
					<h2>Account Management</h2>
					<div class="actions-grid">
						<a href="/account/profile" class="action-card">
							<div class="action-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
									<circle cx="12" cy="7" r="4"/>
								</svg>
							</div>
							<div class="action-content">
								<h3>Edit Profile</h3>
								<p>Update your personal information</p>
							</div>
							<div class="action-arrow">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M9 18l6-6-6-6"/>
								</svg>
							</div>
						</a>

						<a href="/account/settings" class="action-card">
							<div class="action-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<circle cx="12" cy="12" r="3"/>
									<path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1 1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
								</svg>
							</div>
							<div class="action-content">
								<h3>Account Settings</h3>
								<p>Manage preferences and privacy</p>
							</div>
							<div class="action-arrow">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M9 18l6-6-6-6"/>
								</svg>
							</div>
						</a>

						<a href="/account/billing" class="action-card">
							<div class="action-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<rect x="1" y="4" width="22" height="16" rx="2" ry="2"/>
									<line x1="1" y1="10" x2="23" y2="10"/>
								</svg>
							</div>
							<div class="action-content">
								<h3>Billing History</h3>
								<p>View invoices and payment methods</p>
							</div>
							<div class="action-arrow">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M9 18l6-6-6-6"/>
								</svg>
							</div>
						</a>

						<a href="/advertise" class="gold-action gold-btn">
							<div class="gold-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
									<circle cx="12" cy="12" r="3"/>
								</svg>
							</div>
							<div class="gold-content">
								<h3>Advertise with BOME</h3>
								<p>Reach engaged viewers with your message</p>
							</div>
							<div class="gold-arrow">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M9 18l6-6-6-6"/>
								</svg>
							</div>
						</a>
					</div>
				</div>
			</div>
		{:else if activeTab === 'advertiser'}
			<!-- Advertiser Tab Content -->
			<div class="tab-content">
				{#if advertiserStatus === 'none'}
					<!-- Not an Advertiser State -->
					<div class="advertiser-state not-advertiser">
						<div class="state-header glass">
							<div class="state-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
									<circle cx="12" cy="12" r="3"/>
								</svg>
							</div>
							<div class="state-content">
								<h1>Advertise on BOME</h1>
								<p>Reach thousands of engaged viewers interested in Book of Mormon evidences and research</p>
							</div>
						</div>

						<div class="benefits-grid">
							<div class="benefit-card glass">
								<div class="benefit-icon">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
										<circle cx="9" cy="7" r="4"/>
										<path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
										<path d="M16 3.13a4 4 0 0 1 0 7.75"/>
									</svg>
								</div>
								<h3>Targeted Audience</h3>
								<p>Connect with viewers genuinely interested in Book of Mormon research and evidences</p>
							</div>

							<div class="benefit-card glass">
								<div class="benefit-icon">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<line x1="12" y1="1" x2="12" y2="23"/>
										<path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/>
									</svg>
								</div>
								<h3>Flexible Pricing</h3>
								<p>Choose from multiple advertising packages that fit your budget and goals</p>
							</div>

							<div class="benefit-card glass">
								<div class="benefit-icon">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M3 3v18h18"/>
										<path d="M18.7 8l-5.1 5.2-2.8-2.7L7 14.3"/>
									</svg>
								</div>
								<h3>Detailed Analytics</h3>
								<p>Track your campaign performance with comprehensive analytics and reporting</p>
							</div>

							<div class="benefit-card glass">
								<div class="benefit-icon">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
									</svg>
								</div>
								<h3>Quality Assurance</h3>
								<p>All advertisements are reviewed to ensure they align with our community standards</p>
							</div>
						</div>

						<div class="cta-section glass">
							<h2>Ready to Get Started?</h2>
							<p>Join other businesses reaching engaged audiences on BOME</p>
							<button class="primary-button" on:click={handleAdvertiseClick}>
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
								</svg>
								Start Advertising
							</button>
						</div>
					</div>
				{:else if advertiserStatus === 'pending'}
					<!-- Pending Approval State -->
					<div class="advertiser-state pending-approval">
						<div class="state-header glass">
							<div class="state-icon pending">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<circle cx="12" cy="12" r="10"/>
									<polyline points="12,6 12,12 16,14"/>
								</svg>
							</div>
							<div class="state-content">
								<h1>Application Under Review</h1>
								<p>Thank you for your interest in advertising with BOME. Your application is being reviewed.</p>
							</div>
						</div>

						<div class="application-summary glass">
							<h2>Application Summary</h2>
							<div class="summary-grid">
								<div class="summary-item">
									<span class="label">Company Name:</span>
									<span class="value">{advertiserAccount?.company_name}</span>
								</div>
								<div class="summary-item">
									<span class="label">Contact Name:</span>
									<span class="value">{advertiserAccount?.contact_name}</span>
								</div>
								<div class="summary-item">
									<span class="label">Business Email:</span>
									<span class="value">{advertiserAccount?.business_email}</span>
								</div>
								<div class="summary-item">
									<span class="label">Selected Package:</span>
									<span class="value">{advertiserAccount?.submitted_package}</span>
								</div>
								<div class="summary-item">
									<span class="label">Submitted:</span>
									<span class="value">{formatDate(advertiserAccount?.created_at)}</span>
								</div>
							</div>
						</div>

						<div class="review-timeline glass">
							<h2>Review Timeline</h2>
							<div class="timeline">
								<div class="timeline-item completed">
									<div class="timeline-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<polyline points="20,6 9,17 4,12"/>
										</svg>
									</div>
									<div class="timeline-content">
										<h3>Application Submitted</h3>
										<p>Your advertiser application has been received</p>
										<span class="timeline-date">{formatDate(advertiserAccount?.created_at)}</span>
									</div>
								</div>
								<div class="timeline-item active">
									<div class="timeline-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<circle cx="12" cy="12" r="10"/>
											<polyline points="12,6 12,12 16,14"/>
										</svg>
									</div>
									<div class="timeline-content">
										<h3>Under Review</h3>
										<p>Our team is reviewing your application and business information</p>
										<span class="timeline-date">In Progress</span>
									</div>
								</div>
								<div class="timeline-item">
									<div class="timeline-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
										</svg>
									</div>
									<div class="timeline-content">
										<h3>Approval Decision</h3>
										<p>You'll receive an email notification with our decision</p>
										<span class="timeline-date">Pending</span>
									</div>
								</div>
							</div>
						</div>

						<div class="next-steps glass">
							<h2>What Happens Next?</h2>
							<ul>
								<li>Our team will review your business information and advertising goals</li>
								<li>We'll verify your business credentials and contact information</li>
								<li>You'll receive an email notification within 2-3 business days</li>
								<li>Once approved, you can start creating and managing campaigns</li>
							</ul>
						</div>
					</div>
				{:else if advertiserStatus === 'approved'}
					<!-- Approved Advertiser State -->
					<div class="advertiser-state approved-advertiser">
						<div class="state-header glass">
							<div class="state-icon approved">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<polyline points="20,6 9,17 4,12"/>
								</svg>
							</div>
							<div class="state-content">
								<h1>Welcome, Approved Advertiser!</h1>
								<p>Your advertising account is active. Start creating campaigns to reach your audience.</p>
							</div>
						</div>

						<div class="account-overview glass">
							<h2>Account Overview</h2>
							<div class="overview-grid">
								<div class="overview-card">
									<div class="overview-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
										</svg>
									</div>
									<div class="overview-content">
										<div class="overview-value">3</div>
										<div class="overview-label">Active Campaigns</div>
									</div>
								</div>
								<div class="overview-card">
									<div class="overview-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<line x1="12" y1="1" x2="12" y2="23"/>
											<path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/>
										</svg>
									</div>
									<div class="overview-content">
										<div class="overview-value">{formatCurrency(1250)}</div>
										<div class="overview-label">Monthly Spend</div>
									</div>
								</div>
								<div class="overview-card">
									<div class="overview-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/>
											<path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/>
										</svg>
									</div>
									<div class="overview-content">
										<div class="overview-value">15,420</div>
										<div class="overview-label">Total Impressions</div>
									</div>
								</div>
								<div class="overview-card">
									<div class="overview-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path d="M3 3v18h18"/>
											<path d="M18.7 8l-5.1 5.2-2.8-2.7L7 14.3"/>
										</svg>
									</div>
									<div class="overview-content">
										<div class="overview-value">2.3%</div>
										<div class="overview-label">Click-Through Rate</div>
									</div>
								</div>
							</div>
						</div>

						<div class="campaign-management glass">
							<div class="section-header">
								<h2>Campaign Management</h2>
								<button class="secondary-button" on:click={handleCreateCampaign}>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<line x1="12" y1="5" x2="12" y2="19"/>
										<line x1="5" y1="12" x2="19" y2="12"/>
									</svg>
									Create Campaign
								</button>
							</div>
							<div class="campaign-list">
								<div class="campaign-item">
									<div class="campaign-info">
										<h3>Book of Mormon Evidence Series</h3>
										<p>Educational content promotion • Active since March 15</p>
									</div>
									<div class="campaign-metrics">
										<span class="metric">
											<strong>5,240</strong> impressions
										</span>
										<span class="metric">
											<strong>127</strong> clicks
										</span>
										<span class="metric">
											<strong>{formatCurrency(456)}</strong> spent
										</span>
									</div>
									<div class="campaign-status active">Active</div>
								</div>
								<div class="campaign-item">
									<div class="campaign-info">
										<h3>Academic Research Resources</h3>
										<p>Resource promotion • Active since March 8</p>
									</div>
									<div class="campaign-metrics">
										<span class="metric">
											<strong>3,180</strong> impressions
										</span>
										<span class="metric">
											<strong>89</strong> clicks
										</span>
										<span class="metric">
											<strong>{formatCurrency(324)}</strong> spent
										</span>
									</div>
									<div class="campaign-status active">Active</div>
								</div>
							</div>
							<button class="text-button" on:click={handleManageCampaigns}>
								View All Campaigns →
							</button>
						</div>

						<div class="quick-actions-advertiser glass">
							<h2>Quick Actions</h2>
							<div class="action-buttons">
								<button class="action-btn primary" on:click={handleCreateCampaign}>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<line x1="12" y1="5" x2="12" y2="19"/>
										<line x1="5" y1="12" x2="19" y2="12"/>
									</svg>
									Create New Campaign
								</button>
								<button class="action-btn" on:click={handleManageCampaigns}>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
									</svg>
									Manage Campaigns
								</button>
								<a href="/advertise/analytics" class="action-btn">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M3 3v18h18"/>
										<path d="M18.7 8l-5.1 5.2-2.8-2.7L7 14.3"/>
									</svg>
									View Analytics
								</a>
							</div>
						</div>
					</div>
				{:else if advertiserStatus === 'rejected'}
					<!-- Rejected Application State -->
					<div class="advertiser-state rejected-application">
						<div class="state-header glass">
							<div class="state-icon rejected">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<circle cx="12" cy="12" r="10"/>
									<line x1="15" y1="9" x2="9" y2="15"/>
									<line x1="9" y1="9" x2="15" y2="15"/>
								</svg>
							</div>
							<div class="state-content">
								<h1>Application Not Approved</h1>
								<p>Unfortunately, we cannot approve your advertiser application at this time.</p>
							</div>
						</div>

						<div class="rejection-details glass">
							<h2>Reason for Rejection</h2>
							<div class="rejection-reason">
								<p>{advertiserAccount?.verification_notes || 'No specific reason provided.'}</p>
							</div>
							<div class="rejection-info">
								<p><strong>Reviewed on:</strong> {formatDate(advertiserAccount?.rejected_at)}</p>
							</div>
						</div>

						<div class="next-steps-rejected glass">
							<h2>What You Can Do Next</h2>
							<div class="steps-list">
								<div class="step-item">
									<div class="step-number">1</div>
									<div class="step-content">
										<h3>Review the Requirements</h3>
										<p>Make sure you understand our advertiser guidelines and requirements</p>
									</div>
								</div>
								<div class="step-item">
									<div class="step-number">2</div>
									<div class="step-content">
										<h3>Address the Issues</h3>
										<p>Take time to address the specific concerns mentioned in the rejection reason</p>
									</div>
								</div>
								<div class="step-item">
									<div class="step-number">3</div>
									<div class="step-content">
										<h3>Reapply When Ready</h3>
										<p>You can submit a new application once you've addressed the issues</p>
									</div>
								</div>
							</div>
						</div>

						<div class="reapply-section glass">
							<h2>Ready to Try Again?</h2>
							<p>If you believe you've addressed the issues mentioned, you can submit a new application.</p>
							<button class="primary-button" on:click={handleReapply}>
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M1 4v6h6"/>
									<path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/>
								</svg>
								Submit New Application
							</button>
						</div>
					</div>
				{/if}
			</div>
		{/if}
	</div>
{/if}

<Footer />

<style>
	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 60vh;
		gap: var(--space-lg);
	}

	.loading-container p {
		font-size: var(--text-lg);
		color: var(--text-secondary);
	}

	.dashboard {
		display: flex;
		flex-direction: column;
		gap: var(--space-2xl);
		padding: var(--space-xl);
		max-width: 1200px;
		margin: 0 auto;
	}

	.welcome-section {
		padding: var(--space-2xl);
		border-radius: var(--radius-xl);
		display: flex;
		align-items: center;
		justify-content: space-between;
		background: var(--primary-gradient);
		color: var(--white);
	}

	.welcome-content h1 {
		font-size: var(--text-3xl);
		font-weight: 700;
		margin-bottom: var(--space-sm);
	}

	.welcome-content p {
		font-size: var(--text-lg);
		opacity: 0.9;
	}

	.user-avatar-large {
		width: 80px;
		height: 80px;
		border-radius: 50%;
		background: rgba(255, 255, 255, 0.2);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: var(--text-2xl);
		font-weight: 700;
		color: var(--white);
		border: 3px solid rgba(255, 255, 255, 0.3);
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: var(--space-xl);
	}

	.stat-card {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		display: flex;
		align-items: center;
		gap: var(--space-lg);
		transition: all var(--transition-normal);
	}

	.stat-card:hover {
		transform: translateY(-4px);
		box-shadow: var(--shadow-lg);
	}

	.stat-icon {
		width: 56px;
		height: 56px;
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: var(--shadow-md);
	}

	.stat-icon.videos {
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
	}

	.stat-icon.time {
		background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
	}

	.stat-icon.favorites {
		background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
	}

	.stat-icon.subscription {
		background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
	}

	.stat-icon svg {
		width: 28px;
		height: 28px;
		color: var(--white);
	}

	.stat-value {
		font-size: var(--text-2xl);
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: var(--space-xs);
	}

	.stat-label {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.content-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-xl);
	}

	.section-card {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
	}

	.section-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: var(--space-xl);
	}

	.section-header h2 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.view-all-link {
		color: var(--primary);
		text-decoration: none;
		font-size: var(--text-sm);
		font-weight: 600;
		transition: all var(--transition-normal);
	}

	.view-all-link:hover {
		color: var(--primary-dark);
		transform: translateX(4px);
	}

	.video-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: var(--space-lg);
	}

	.video-card {
		cursor: pointer;
		transition: all var(--transition-normal);
	}

	.video-card:hover {
		transform: translateY(-4px);
	}

	.video-thumbnail {
		position: relative;
		width: 100%;
		aspect-ratio: 16/9;
		border-radius: var(--radius-lg);
		overflow: hidden;
		margin-bottom: var(--space-md);
	}

	.video-thumbnail img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.video-duration {
		position: absolute;
		bottom: var(--space-sm);
		right: var(--space-sm);
		background: rgba(0, 0, 0, 0.8);
		color: var(--white);
		padding: 2px 6px;
		border-radius: var(--radius-sm);
		font-size: var(--text-xs);
		font-weight: 600;
	}

	.progress-bar {
		position: absolute;
		bottom: 0;
		left: 0;
		width: 100%;
		height: 4px;
		background: rgba(0, 0, 0, 0.3);
	}

	.progress-fill {
		height: 100%;
		background: var(--primary);
		transition: width var(--transition-normal);
	}

	.video-title {
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-xs);
		line-height: 1.4;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	.video-meta {
		font-size: var(--text-xs);
		color: var(--text-secondary);
		margin: 0;
	}

	.activity-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.activity-item {
		display: flex;
		align-items: center;
		gap: var(--space-md);
	}

	.activity-icon {
		width: 40px;
		height: 40px;
		border-radius: var(--radius-lg);
		background: var(--bg-glass);
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.activity-icon svg {
		width: 20px;
		height: 20px;
		color: var(--text-primary);
	}

	.activity-text {
		font-size: var(--text-sm);
		color: var(--text-primary);
		font-weight: 500;
		margin-bottom: var(--space-xs);
	}

	.activity-time {
		font-size: var(--text-xs);
		color: var(--text-secondary);
	}

	.quick-actions {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
	}

	.quick-actions h2 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-xl);
	}

	.action-buttons {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: var(--space-lg);
	}

	.action-btn {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		padding: var(--space-lg);
		background: var(--bg-glass);
		border-radius: var(--radius-lg);
		text-decoration: none;
		color: var(--text-primary);
		font-weight: 500;
		transition: all var(--transition-normal);
	}

	.action-btn:hover {
		background: var(--bg-glass-dark);
		transform: translateY(-2px);
		box-shadow: var(--shadow-md);
	}

	.action-btn svg {
		width: 20px;
		height: 20px;
		color: var(--primary);
	}

	.gold-btn {
		background: linear-gradient(135deg, #FFD700 0%, #FFA500 100%);
		color: #8B4513;
		border: 2px solid transparent;
		position: relative;
		overflow: hidden;
		font-weight: 600;
		box-shadow: 0 4px 15px rgba(255, 215, 0, 0.3);
		display: flex;
		align-items: center;
		gap: var(--space-lg);
		padding: var(--space-lg);
		border-radius: var(--radius-lg);
		text-decoration: none;
		transition: all var(--transition-normal);
		cursor: pointer;
	}

	.gold-btn::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: linear-gradient(135deg, #FFA500 0%, #FFD700 100%);
		opacity: 0;
		transition: opacity var(--transition-normal);
		z-index: 1;
	}

	.gold-btn:hover::before {
		opacity: 1;
	}

	.gold-btn:hover {
		transform: translateY(-4px);
		box-shadow: 0 8px 30px rgba(255, 215, 0, 0.4);
	}

	.gold-btn:hover .gold-content h3,
	.gold-btn:hover .gold-content p {
		color: #654321;
	}

	.gold-btn:hover .gold-arrow {
		color: #654321;
		transform: translateX(4px);
	}

	.gold-btn svg {
		color: #8B4513;
		position: relative;
		z-index: 2;
	}

	.gold-icon {
		width: 48px;
		height: 48px;
		border-radius: var(--radius-lg);
		background: var(--primary-gradient);
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		position: relative;
		z-index: 2;
	}

	.gold-icon svg {
		width: 24px;
		height: 24px;
		color: var(--white);
	}

	.gold-content {
		flex: 1;
		position: relative;
		z-index: 2;
	}

	.gold-content h3 {
		font-size: var(--text-base);
		font-weight: 600;
		color: rgb(15, 11, 32);
		margin-bottom: var(--space-xs);
	}

	.gold-content p {
		font-size: var(--text-sm);
		color: rgb(15, 11, 32);
	}

	.gold-arrow {
		width: 24px;
		height: 24px;
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--text-secondary);
		transition: all var(--transition-normal);
		position: relative;
		z-index: 2;
	}

	.gold-card:hover .action-arrow {
		color: white;
		transform: translateX(4px);
	}

	.gold-arrow svg {
		width: 16px;
		height: 16px;
	}

	.advertise-btn {
		background: linear-gradient(135deg, var(--primary) 0%, var(--secondary) 100%);
		color: var(--white);
		border: 2px solid transparent;
		position: relative;
		overflow: hidden;
	}

	.advertise-btn::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: linear-gradient(135deg, var(--secondary) 0%, var(--primary) 100%);
		opacity: 0;
		transition: opacity var(--transition-normal);
	}

	.advertise-btn:hover::before {
		opacity: 1;
	}

	.advertise-btn:hover {
		transform: translateY(-4px);
		box-shadow: 0 12px 40px rgba(0, 0, 0, 0.15);
	}

	.advertise-btn svg {
		color: var(--white);
		position: relative;
		z-index: 1;
	}

	.advertise-btn span,
	.advertise-btn {
		position: relative;
		z-index: 1;
	}

	.campaign-btn {
		background: linear-gradient(135deg, var(--success) 0%, #38d9a9 100%);
		color: var(--white);
		border: 2px solid transparent;
		position: relative;
		overflow: hidden;
	}

	.campaign-btn::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: linear-gradient(135deg, #38d9a9 0%, var(--success) 100%);
		opacity: 0;
		transition: opacity var(--transition-normal);
	}

	.campaign-btn:hover::before {
		opacity: 1;
	}

	.campaign-btn:hover {
		transform: translateY(-4px);
		box-shadow: 0 12px 40px rgba(0, 0, 0, 0.15);
	}

	.campaign-btn svg {
		color: var(--white);
		position: relative;
		z-index: 1;
	}

	.campaign-btn span,
	.campaign-btn {
		position: relative;
		z-index: 1;
	}

	.advertiser-dashboard-btn {
		background: linear-gradient(135deg, var(--warning) 0%, #ffa726 100%);
		color: var(--white);
		border: 2px solid transparent;
		position: relative;
		overflow: hidden;
	}

	.advertiser-dashboard-btn::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: linear-gradient(135deg, #ffa726 0%, var(--warning) 100%);
		opacity: 0;
		transition: opacity var(--transition-normal);
	}

	.advertiser-dashboard-btn:hover::before {
		opacity: 1;
	}

	.advertiser-dashboard-btn:hover {
		transform: translateY(-4px);
		box-shadow: 0 12px 40px rgba(0, 0, 0, 0.15);
	}

	.advertiser-dashboard-btn svg {
		color: var(--white);
		position: relative;
		z-index: 1;
	}

	.advertiser-dashboard-btn span,
	.advertiser-dashboard-btn {
		position: relative;
		z-index: 1;
	}

	@media (max-width: 1024px) {
		.content-grid {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 768px) {
		.dashboard {
			padding: var(--space-lg);
		}

		.welcome-section {
			flex-direction: column;
			text-align: center;
			gap: var(--space-lg);
		}

		.stats-grid {
			grid-template-columns: repeat(2, 1fr);
		}

		.video-grid {
			grid-template-columns: 1fr;
		}

		.action-buttons {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	@media (max-width: 480px) {
		.stats-grid {
			grid-template-columns: 1fr;
		}

		.action-buttons {
			grid-template-columns: 1fr;
		}

		.stat-card {
			flex-direction: column;
			text-align: center;
			gap: var(--space-md);
		}
	}

	/* Tab Navigation Styles */
	.tab-navigation {
		display: flex;
		gap: var(--space-md);
		padding: var(--space-lg);
		border-radius: var(--radius-xl);
		margin-bottom: var(--space-xl);
		overflow-x: auto;
	}

	.tab-button {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		padding: var(--space-lg) var(--space-xl);
		background: transparent;
		border: 2px solid var(--border-light);
		border-radius: var(--radius-lg);
		color: var(--text-secondary);
		font-weight: 500;
		cursor: pointer;
		transition: all var(--transition-normal);
		white-space: nowrap;
		position: relative;
	}

	.tab-button:hover {
		background: var(--bg-glass);
		border-color: var(--primary-light);
		color: var(--text-primary);
		transform: translateY(-2px);
	}

	.tab-button.active {
		background: var(--primary-gradient);
		border-color: var(--primary);
		color: var(--white);
		box-shadow: var(--shadow-md);
	}

	.tab-button svg {
		width: 20px;
		height: 20px;
	}

	.status-indicator {
		padding: 4px 8px;
		border-radius: var(--radius-sm);
		font-size: var(--text-xs);
		font-weight: 600;
		margin-left: var(--space-sm);
	}

	.status-indicator.pending {
		background: var(--warning-light);
		color: var(--warning-dark);
	}

	.status-indicator.approved {
		background: var(--success-light);
		color: var(--success-dark);
	}

	.status-indicator.rejected {
		background: var(--error-light);
		color: var(--error-dark);
	}

	.tab-content {
		display: flex;
		flex-direction: column;
		gap: var(--space-2xl);
	}

	/* Advertiser State Styles */
	.advertiser-state {
		display: flex;
		flex-direction: column;
		gap: var(--space-2xl);
	}

	.state-header {
		display: flex;
		align-items: center;
		gap: var(--space-xl);
		padding: var(--space-2xl);
		border-radius: var(--radius-xl);
	}

	.state-icon {
		width: 80px;
		height: 80px;
		border-radius: var(--radius-xl);
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.state-icon svg {
		width: 40px;
		height: 40px;
		color: var(--white);
	}

	.state-icon:not(.pending):not(.approved):not(.rejected) {
		background: var(--primary-gradient);
	}

	.state-icon.pending {
		background: linear-gradient(135deg, var(--warning) 0%, #ffa726 100%);
	}

	.state-icon.approved {
		background: linear-gradient(135deg, var(--success) 0%, #38d9a9 100%);
	}

	.state-icon.rejected {
		background: linear-gradient(135deg, var(--error) 0%, #f56565 100%);
	}

	.state-content h1 {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: var(--space-sm);
	}

	.state-content p {
		font-size: var(--text-lg);
		color: var(--text-secondary);
		line-height: 1.6;
	}

	/* Benefits Grid */
	.benefits-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
		gap: var(--space-xl);
	}

	.benefit-card {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		text-align: center;
		transition: all var(--transition-normal);
	}

	.benefit-card:hover {
		transform: translateY(-4px);
		box-shadow: var(--shadow-lg);
	}

	.benefit-icon {
		width: 60px;
		height: 60px;
		border-radius: var(--radius-lg);
		background: var(--primary-gradient);
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 0 auto var(--space-lg);
	}

	.benefit-icon svg {
		width: 28px;
		height: 28px;
		color: var(--white);
	}

	.benefit-card h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-md);
	}

	.benefit-card p {
		color: var(--text-secondary);
		line-height: 1.6;
	}

	/* CTA Section */
	.cta-section {
		padding: var(--space-2xl);
		border-radius: var(--radius-xl);
		text-align: center;
	}

	.cta-section h2 {
		font-size: var(--text-2xl);
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: var(--space-md);
	}

	.cta-section p {
		font-size: var(--text-lg);
		color: var(--text-secondary);
		margin-bottom: var(--space-xl);
	}

	.primary-button {
		display: inline-flex;
		align-items: center;
		gap: var(--space-md);
		padding: var(--space-lg) var(--space-2xl);
		background: var(--primary-gradient);
		color: var(--white);
		border: none;
		border-radius: var(--radius-lg);
		font-weight: 600;
		font-size: var(--text-base);
		cursor: pointer;
		transition: all var(--transition-normal);
		box-shadow: var(--shadow-md);
	}

	.primary-button:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}

	.primary-button svg {
		width: 20px;
		height: 20px;
	}

	.secondary-button {
		display: inline-flex;
		align-items: center;
		gap: var(--space-md);
		padding: var(--space-md) var(--space-lg);
		background: var(--bg-glass);
		color: var(--text-primary);
		border: 2px solid var(--border-light);
		border-radius: var(--radius-lg);
		font-weight: 500;
		cursor: pointer;
		transition: all var(--transition-normal);
	}

	.secondary-button:hover {
		background: var(--bg-glass-dark);
		border-color: var(--primary-light);
		transform: translateY(-2px);
	}

	.secondary-button svg {
		width: 16px;
		height: 16px;
	}

	.text-button {
		background: none;
		border: none;
		color: var(--primary);
		font-weight: 600;
		cursor: pointer;
		transition: all var(--transition-normal);
		padding: var(--space-md) 0;
	}

	.text-button:hover {
		color: var(--primary-dark);
		transform: translateX(4px);
	}

	/* Application Summary */
	.application-summary {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
	}

	.application-summary h2 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-xl);
	}

	.summary-grid {
		display: grid;
		gap: var(--space-lg);
	}

	.summary-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-md) 0;
		border-bottom: 1px solid var(--border-light);
	}

	.summary-item:last-child {
		border-bottom: none;
	}

	.summary-item .label {
		font-weight: 500;
		color: var(--text-secondary);
	}

	.summary-item .value {
		font-weight: 600;
		color: var(--text-primary);
	}

	/* Timeline Styles */
	.review-timeline {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
	}

	.review-timeline h2 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-xl);
	}

	.timeline {
		display: flex;
		flex-direction: column;
		gap: var(--space-xl);
		position: relative;
	}

	.timeline::before {
		content: '';
		position: absolute;
		left: 20px;
		top: 40px;
		bottom: 40px;
		width: 2px;
		background: var(--border-light);
	}

	.timeline-item {
		display: flex;
		gap: var(--space-lg);
		position: relative;
	}

	.timeline-icon {
		width: 40px;
		height: 40px;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		background: var(--bg-secondary);
		border: 2px solid var(--border-light);
		position: relative;
		z-index: 1;
	}

	.timeline-item.completed .timeline-icon {
		background: var(--success);
		border-color: var(--success);
	}

	.timeline-item.active .timeline-icon {
		background: var(--warning);
		border-color: var(--warning);
	}

	.timeline-icon svg {
		width: 20px;
		height: 20px;
		color: var(--white);
	}

	.timeline-content h3 {
		font-size: var(--text-base);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-xs);
	}

	.timeline-content p {
		color: var(--text-secondary);
		margin-bottom: var(--space-xs);
	}

	.timeline-date {
		font-size: var(--text-sm);
		color: var(--text-tertiary);
		font-weight: 500;
	}

	/* Next Steps */
	.next-steps {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
	}

	.next-steps h2 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-lg);
	}

	.next-steps ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.next-steps li {
		padding: var(--space-md) 0;
		color: var(--text-secondary);
		position: relative;
		padding-left: var(--space-xl);
	}

	.next-steps li::before {
		content: '•';
		color: var(--primary);
		font-weight: bold;
		position: absolute;
		left: 0;
	}

	/* Account Overview */
	.account-overview {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
	}

	.account-overview h2 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-xl);
	}

	.overview-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: var(--space-lg);
	}

	.overview-card {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		padding: var(--space-lg);
		background: var(--bg-glass);
		border-radius: var(--radius-lg);
		transition: all var(--transition-normal);
	}

	.overview-card:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-md);
	}

	.overview-icon {
		width: 48px;
		height: 48px;
		border-radius: var(--radius-lg);
		background: var(--primary-gradient);
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.overview-icon svg {
		width: 24px;
		height: 24px;
		color: var(--white);
	}

	.overview-value {
		font-size: var(--text-xl);
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: var(--space-xs);
	}

	.overview-label {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	/* Campaign Management */
	.campaign-management {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
	}

	.campaign-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
		margin-bottom: var(--space-lg);
	}

	.campaign-item {
		display: flex;
		align-items: center;
		gap: var(--space-lg);
		padding: var(--space-lg);
		background: var(--bg-glass);
		border-radius: var(--radius-lg);
		transition: all var(--transition-normal);
	}

	.campaign-item:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-md);
	}

	.campaign-info {
		flex: 1;
	}

	.campaign-info h3 {
		font-size: var(--text-base);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-xs);
	}

	.campaign-info p {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.campaign-metrics {
		display: flex;
		gap: var(--space-lg);
		align-items: center;
	}

	.metric {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.campaign-status {
		padding: var(--space-xs) var(--space-md);
		border-radius: var(--radius-sm);
		font-size: var(--text-xs);
		font-weight: 600;
	}

	.campaign-status.active {
		background: var(--success-light);
		color: var(--success-dark);
	}

	/* Quick Actions Advertiser */
	.quick-actions-advertiser {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
	}

	.quick-actions-advertiser h2 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-xl);
	}

	.quick-actions-advertiser .action-buttons {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: var(--space-lg);
	}

	.quick-actions-advertiser .action-btn.primary {
		background: var(--primary-gradient);
		color: var(--white);
		border: none;
	}

	.quick-actions-advertiser .action-btn.primary:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}

	/* Rejection Details */
	.rejection-details {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
	}

	.rejection-details h2 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-lg);
	}

	.rejection-reason {
		padding: var(--space-lg);
		background: var(--error-light);
		border-left: 4px solid var(--error);
		border-radius: var(--radius-lg);
		margin-bottom: var(--space-lg);
	}

	.rejection-reason p {
		color: var(--error-dark);
		font-weight: 500;
		margin: 0;
	}

	.rejection-info {
		color: var(--text-secondary);
	}

	/* Next Steps Rejected */
	.next-steps-rejected {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
	}

	.next-steps-rejected h2 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-xl);
	}

	.steps-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.step-item {
		display: flex;
		gap: var(--space-lg);
		align-items: flex-start;
	}

	.step-number {
		width: 32px;
		height: 32px;
		border-radius: 50%;
		background: var(--primary-gradient);
		color: var(--white);
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 600;
		flex-shrink: 0;
	}

	.step-content h3 {
		font-size: var(--text-base);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-xs);
	}

	.step-content p {
		color: var(--text-secondary);
		line-height: 1.6;
	}

	/* Reapply Section */
	.reapply-section {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		text-align: center;
	}

	.reapply-section h2 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-md);
	}

	.reapply-section p {
		color: var(--text-secondary);
		margin-bottom: var(--space-xl);
	}

	/* Responsive Design for Tabs */
	@media (max-width: 768px) {
		.tab-navigation {
			padding: var(--space-md);
		}

		.tab-button {
			padding: var(--space-md) var(--space-lg);
			font-size: var(--text-sm);
		}

		.state-header {
			flex-direction: column;
			text-align: center;
			gap: var(--space-lg);
		}

		.state-icon {
			width: 60px;
			height: 60px;
		}

		.state-icon svg {
			width: 30px;
			height: 30px;
		}

		.benefits-grid {
			grid-template-columns: 1fr;
		}

		.overview-grid {
			grid-template-columns: repeat(2, 1fr);
		}

		.campaign-item {
			flex-direction: column;
			align-items: flex-start;
			gap: var(--space-md);
		}

		.campaign-metrics {
			flex-direction: column;
			align-items: flex-start;
			gap: var(--space-sm);
		}

		.summary-item {
			flex-direction: column;
			align-items: flex-start;
			gap: var(--space-xs);
		}

		.actions-grid {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 480px) {
		.overview-grid {
			grid-template-columns: 1fr;
		}

		.quick-actions-advertiser .action-buttons {
			grid-template-columns: 1fr;
		}
	}

	/* Account Tab Styles */
	.account-section {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		margin-bottom: var(--space-xl);
	}

	.profile-card {
		display: flex;
		align-items: center;
		gap: var(--space-xl);
		padding: var(--space-xl);
		border-radius: var(--radius-lg);
	}

	.profile-avatar {
		flex-shrink: 0;
	}

	.avatar-placeholder {
		width: 80px;
		height: 80px;
		border-radius: 50%;
		background: var(--primary-gradient);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: var(--text-2xl);
		font-weight: 700;
		color: var(--white);
		border: 3px solid rgba(255, 255, 255, 0.3);
	}

	.profile-info h3 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-xs);
	}

	.profile-info .email {
		color: var(--text-secondary);
		margin-bottom: var(--space-lg);
	}

	.profile-stats {
		display: flex;
		gap: var(--space-xl);
	}

	.stat {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.stat-label {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.stat-value {
		font-size: var(--text-base);
		font-weight: 600;
		color: var(--text-primary);
	}

	.subscription-card {
		padding: var(--space-xl);
		border-radius: var(--radius-lg);
	}

	.subscription-active {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.subscription-status {
		display: flex;
		align-items: center;
		gap: var(--space-md);
	}

	.status-badge {
		padding: var(--space-xs) var(--space-md);
		border-radius: var(--radius-sm);
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--white);
	}

	.subscription-details {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.detail-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-md) 0;
		border-bottom: 1px solid var(--border-light);
	}

	.detail-row:last-child {
		border-bottom: none;
	}

	.detail-row .label {
		font-weight: 500;
		color: var(--text-secondary);
	}

	.detail-row .value {
		font-weight: 600;
		color: var(--text-primary);
	}

	.detail-row .value.warning {
		color: var(--warning);
	}

	.subscription-inactive {
		text-align: center;
		padding: var(--space-2xl);
	}

	.inactive-icon {
		width: 60px;
		height: 60px;
		border-radius: 50%;
		background: var(--bg-glass);
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 0 auto var(--space-lg);
	}

	.inactive-icon svg {
		width: 28px;
		height: 28px;
		color: var(--text-secondary);
	}

	.subscription-inactive h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-md);
	}

	.subscription-inactive p {
		color: var(--text-secondary);
		margin-bottom: var(--space-xl);
	}

	.account-actions {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
	}

	.account-actions h2 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-xl);
	}

	.actions-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: var(--space-lg);
	}

	.action-card {
		display: flex;
		align-items: center;
		gap: var(--space-lg);
		padding: var(--space-lg);
		background: var(--bg-glass);
		border: 2px solid var(--border-light);
		border-radius: var(--radius-lg);
		text-decoration: none;
		color: var(--text-primary);
		transition: all var(--transition-normal);
		cursor: pointer;
	}

	.action-card:hover {
		background: var(--bg-glass-dark);
		border-color: var(--primary-light);
		transform: translateY(-2px);
		box-shadow: var(--shadow-md);
	}

	.action-icon {
		width: 48px;
		height: 48px;
		border-radius: var(--radius-lg);
		background: var(--primary-gradient);
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.action-icon svg {
		width: 24px;
		height: 24px;
		color: var(--white);
	}

	.action-content {
		flex: 1;
	}

	.action-content h3 {
		font-size: var(--text-base);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-xs);
	}

	.action-content p {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.action-arrow {
		width: 24px;
		height: 24px;
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--text-secondary);
		transition: all var(--transition-normal);
	}

	.action-card:hover .action-arrow {
		color: var(--primary);
		transform: translateX(4px);
	}

	.action-arrow svg {
		width: 16px;
		height: 16px;
	}
</style> 