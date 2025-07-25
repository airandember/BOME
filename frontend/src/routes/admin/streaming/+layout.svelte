<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import { fade } from 'svelte/transition';

	// Reactive variables
	let currentUser: any = null;
	let isStreamingAdmin = false;
	let isLoading = true;

	// Navigation items for streaming admin
	const navigationItems = [
		{
			name: 'Dashboard',
			href: '/admin/streaming',
			icon: 'home',
			description: 'Overview and key metrics'
		},
		{
			name: 'Videos',
			href: '/admin/streaming/videos',
			icon: 'video-camera',
			description: 'Manage video content and uploads'
		},
		{
			name: 'Subscriptions',
			href: '/admin/streaming/subscriptions',
			icon: 'credit-card',
			description: 'Manage user subscriptions'
		},
		{
			name: 'Subscribers',
			href: '/admin/streaming/subscribers',
			icon: 'users',
			description: 'Subscriber management and support'
		},
		{
			name: 'Analytics',
			href: '/admin/streaming/analytics',
			icon: 'chart-bar',
			description: 'Revenue and subscription analytics'
		},
		
		{
			name: 'Settings',
			href: '/admin/streaming/settings',
			icon: 'cog',
			description: 'Streaming configuration'
		}
	];

	// Check if user has streaming admin permissions
	function checkStreamingAdminPermissions(user: any): boolean {
		if (!user || !user.role) return false;
		
		const streamingAdminRoles = [
			'super_admin',
			'system_admin', 
			'content_manager',
			'streaming_manager'
		];
		
		return streamingAdminRoles.includes(user.role);
	}

	// Handle navigation
	function handleNavigation(href: string): void {
		goto(href);
	}

	// Initialize component
	onMount(() => {
		try {
			// Subscribe to auth store
			const unsubscribe = auth.subscribe((authState: any) => {
				currentUser = authState.user;
				isStreamingAdmin = checkStreamingAdminPermissions(authState.user);
				isLoading = false;
			});

			// Check permissions immediately
			if (currentUser) {
				isStreamingAdmin = checkStreamingAdminPermissions(currentUser);
			}

			// Redirect if not authorized
			if (!isLoading && !isStreamingAdmin) {
				goto('/admin?error=unauthorized');
			}

			return unsubscribe;
		} catch (error) {
			console.error('Error initializing streaming admin layout:', error);
			isLoading = false;
		}
	});

	// Get current navigation item
	$: currentNav = navigationItems.find(item => $page.url.pathname === item.href) || navigationItems[0];
</script>

<svelte:head>
	<title>Streaming Admin - BOME</title>
	<meta name="description" content="Streaming subscription management and analytics" />
</svelte:head>

{#if isLoading}
	<div class="loading-container">
		<div class="spinner"></div>
	</div>
{:else if !isStreamingAdmin}
	<div class="access-denied">
		<div class="text-center">
			<h1 class="text-2xl font-bold m-4">Access Denied</h1>
			<p class="text-secondary m-6">You don't have permission to access the streaming admin area.</p>
			<a href="/admin" class="btn btn-primary">Return to Admin Dashboard</a>
		</div>
	</div>
{:else}
	<div class="streaming-admin-layout">
		<!-- Header -->
		<header class="admin-header">
			<div class="container">
				<div class="header-content">
					<div class="header-left">
						<button 
							class="back-button"
							on:click={() => goto('/admin')}
						>
							<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
							<span>Back to Main Dashboard</span>
						</button>
						<div class="divider"></div>
						<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="streaming-icon"><path d="M13 19v-6a1 1 0 0 0-1-1H6a1 1 0 0 0-1 1v6a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1z"/><path d="M21 13v-6a1 1 0 0 0-1-1H14a1 1 0 0 0-1 1v6a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1z"/><path d="M21 21v-6a1 1 0 0 0-1-1H14a1 1 0 0 0-1 1v6a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1z"/><path d="M13 21v-6a1 1 0 0 0-1-1H6a1 1 0 0 0-1 1v6a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1z"/></svg>
						<div class="header-info">
							<h1 class="text-2xl font-bold">Streaming Admin</h1>
							<p class="text-secondary">Subscription management and analytics</p>
						</div>
					</div>
					
					<div class="header-right">
						<!-- User info -->
						{#if currentUser}
							<div class="user-info">
								<p class="user-email">{currentUser.email}</p>
								<p class="user-role">{currentUser.role?.replace('_', ' ')}</p>
							</div>
							<div class="user-avatar">
								<span>{currentUser.email?.charAt(0).toUpperCase()}</span>
							</div>
						{/if}
					</div>
				</div>
			</div>
		</header>

		<div class="container p-5">
			<div class="layout-content">
				<!-- Sidebar Navigation -->
				<nav class="sidebar">
					<div class="nav-card">
						<h2 class="nav-title">Navigation</h2>
						<ul class="nav-list">
							{#each navigationItems as item}
								<li>
									<button
										class="nav-item {currentNav.name === item.name ? 'active' : ''}"
										on:click={() => handleNavigation(item.href)}
									>
										{#if item.icon === 'home'}
											<svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path><polyline points="9,22 9,12 15,12 15,22"></polyline></svg>
										{:else if item.icon === 'video-camera'}
											<svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 19V5a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2z"/><path d="M17 21l4-4-4-4"/><path d="M9 11h12"/><path d="M9 17h12"/></svg>
										{:else if item.icon === 'credit-card'}
											<svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="1" y="4" width="22" height="16" rx="2" ry="2"></rect><line x1="1" y1="10" x2="23" y2="10"></line></svg>
										{:else if item.icon === 'users'}
											<svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>
										{:else if item.icon === 'chart-bar'}
											<svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 20V10"></path><path d="M12 20V4"></path><path d="M6 20v-6"></path></svg>
										{:else if item.icon === 'tag'}
											<svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"></path><line x1="7" y1="7" x2="7.01" y2="7"></line></svg>
										{:else if item.icon === 'calendar'}
											<svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
										{:else}
											<svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>
										{/if}
										<div class="nav-content">
											<div class="nav-name">{item.name}</div>
											<div class="nav-description">{item.description}</div>
										</div>
									</button>
								</li>
							{/each}
						</ul>
					</div>

					<!-- Quick Stats -->
					<div class="stats-card">
						<h3 class="stats-title">Quick Stats</h3>
						<div class="stats-list">
							<div class="stat-item">
								<span class="stat-label">Active Subscriptions</span>
								<span class="stat-value">--</span>
							</div>
							<div class="stat-item">
								<span class="stat-label">Monthly Revenue</span>
								<span class="stat-value">--</span>
							</div>
							<div class="stat-item">
								<span class="stat-label">Churn Rate</span>
								<span class="stat-value">--</span>
							</div>
						</div>
					</div>
				</nav>

				<!-- Main Content -->
				<main class="main-content" in:fade={{ duration: 200 }}>
					<div class="content-card">
						<!-- Page Header -->
						<div class="page-header">
							<h2 class="page-title">{currentNav.name}</h2>
							<p class="page-description">{currentNav.description}</p>
						</div>

						<!-- Page Content -->
						<div class="page-content">
							<slot />
						</div>
					</div>
				</main>
			</div>
		</div>
	</div>
{/if}

<style>
	/* Loading and Access Denied */
	.loading-container {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.access-denied {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	/* Layout Structure */
	.streaming-admin-layout {
		min-height: 100vh;
		background: var(--bg-secondary);
	}

	/* Header */
	.admin-header {
		background: var(--bg-primary);
		border-bottom: 1px solid var(--gray-200);
		box-shadow: var(--shadow-sm);
	}

	.header-content {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-lg) 0;
	}

	.header-left {
		display: flex;
		align-items: center;
		gap: var(--space-lg);
	}

	.back-button {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		color: var(--text-secondary);
		background: none;
		border: none;
		cursor: pointer;
		padding: var(--space-sm);
		border-radius: var(--radius-md);
		transition: all 0.2s ease;
		font-size: var(--text-sm);
		font-weight: 500;
	}

	.back-button:hover {
		color: var(--text-primary);
		background: var(--bg-secondary);
	}

	.divider {
		width: 1px;
		height: 24px;
		background: var(--gray-300);
	}

	.streaming-icon {
		color: var(--primary);
	}

	.header-info h1 {
		margin: 0;
		color: var(--text-primary);
	}

	.header-info p {
		margin: 0;
		color: var(--text-secondary);
		font-size: var(--text-sm);
	}

	.header-right {
		display: flex;
		align-items: center;
		gap: var(--space-lg);
	}

	.user-info {
		text-align: right;
	}

	.user-email {
		font-size: var(--text-sm);
		font-weight: 500;
		color: var(--text-primary);
		margin: 0;
	}

	.user-role {
		font-size: var(--text-xs);
		color: var(--text-secondary);
		margin: 0;
		text-transform: capitalize;
	}

	.user-avatar {
		width: 32px;
		height: 32px;
		background: var(--primary);
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--white);
		font-size: var(--text-sm);
		font-weight: 500;
	}

	/* Layout Content */
	.layout-content {
		display: flex;
		gap: var(--space-xl);
		margin-top: var(--space-xl);
	}

	/* Sidebar */
	.sidebar {
		width: 280px;
		flex-shrink: 0;
	}

	.nav-card, .stats-card {
		background: var(--bg-primary);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-sm);
		border: 1px solid var(--gray-200);
		padding: var(--space-lg);
		margin-bottom: var(--space-lg);
	}

	.nav-title, .stats-title {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-lg) 0;
	}

	.nav-list {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}

	.nav-item {
		width: 100%;
		display: flex;
		align-items: center;
		gap: var(--space-md);
		padding: var(--space-md);
		text-align: left;
		border-radius: var(--radius-md);
		transition: all 0.2s ease;
		background: none;
		border: none;
		cursor: pointer;
		color: var(--text-secondary);
	}

	.nav-item:hover {
		background: var(--bg-secondary);
		color: var(--text-primary);
	}

	.nav-item.active {
		background: var(--primary);
		color: var(--white);
	}

	.nav-icon {
		width: 20px;
		height: 20px;
		flex-shrink: 0;
	}

	.nav-content {
		flex: 1;
	}

	.nav-name {
		font-weight: 500;
		margin-bottom: 2px;
	}

	.nav-description {
		font-size: var(--text-xs);
		opacity: 0.8;
	}

	/* Stats */
	.stats-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.stat-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.stat-label {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.stat-value {
		font-size: var(--text-sm);
		font-weight: 500;
		color: var(--text-primary);
	}

	/* Main Content */
	.main-content {
		flex: 1;
	}

	.content-card {
		background: var(--bg-primary);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-sm);
		border: 1px solid var(--gray-200);
		padding: var(--space-xl);
	}

	.page-header {
		margin-bottom: var(--space-xl);
	}

	.page-title {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
	}

	.page-description {
		color: var(--text-secondary);
		margin: 0;
	}

	.page-content {
		color: var(--text-primary);
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.layout-content {
			flex-direction: column;
		}
		
		.sidebar {
			width: 100%;
		}
		
		.header-content {
			flex-direction: column;
			gap: var(--space-md);
			text-align: center;
		}
		
		.header-left {
			flex-direction: column;
			gap: var(--space-md);
		}
		
		.header-right {
			flex-direction: column;
			gap: var(--space-md);
		}
	}
</style> 