<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { auth } from '$lib/auth';
	import { goto } from '$app/navigation';
	import { theme } from '$lib/theme';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import ToastContainer from '$lib/components/ToastContainer.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let isAuthenticated = false;
	let isAdmin = false;
	let user: any = null;
	let isLoaded = false;
	let isSidebarOpen = true;
	let authChecked = false;

	// Subscribe to auth store
	const unsubscribe = auth.subscribe(state => {
		isAuthenticated = state.isAuthenticated;
		user = state.user;
		isAdmin = user?.role === 'admin';
		
		// Only check auth after we have a definitive state
		if (!authChecked) {
			checkAuth();
		}
	});

	onMount(() => {
		// Small delay to ensure auth store is initialized
		setTimeout(() => {
			checkAuth();
		}, 50);

		return () => {
			unsubscribe();
		};
	});

	function checkAuth() {
		authChecked = true;
		
		// Check admin access
		if (!isAuthenticated) {
			goto('/login');
			return;
		}

		if (!isAdmin) {
			goto('/');
			return;
		}

		// All checks passed, show the admin interface
		setTimeout(() => {
			isLoaded = true;
		}, 100);
	}

	const toggleSidebar = () => {
		isSidebarOpen = !isSidebarOpen;
	};

	const navItems = [
		{
			href: '/admin',
			label: 'Dashboard',
			icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<rect x="3" y="3" width="7" height="9"></rect>
				<rect x="14" y="3" width="7" height="5"></rect>
				<rect x="14" y="12" width="7" height="9"></rect>
				<rect x="3" y="16" width="7" height="5"></rect>
			</svg>`
		},
		{
			href: '/admin/analytics',
			label: 'Analytics',
			icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<polyline points="22,12 18,12 15,21 9,3 6,12 2,12"></polyline>
			</svg>`
		},
		{
			href: '/admin/users',
			label: 'Users',
			icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
				<circle cx="9" cy="7" r="4"></circle>
				<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
				<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
			</svg>`
		},
		{
			href: '/admin/videos',
			label: 'Videos',
			icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<polygon points="23,7 16,12 23,17 23,7"></polygon>
				<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
			</svg>`
		},
		{
			href: '/admin/financial',
			label: 'Financial',
			icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<line x1="12" y1="1" x2="12" y2="23"></line>
				<path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"></path>
			</svg>`
		},
		{
			href: '/admin/security',
			label: 'Security',
			icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
				<path d="M9 12l2 2 4-4"></path>
			</svg>`
		},
		{
			href: '/admin/roles',
			label: 'Role Management',
			icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M16 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
				<circle cx="12" cy="7" r="4"></circle>
				<path d="M22 21v-2a4 4 0 0 0-3-3.87"></path>
				<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
				<circle cx="16" cy="7" r="4"></circle>
				<path d="M20 11v2a4 4 0 0 1-4 4H8a4 4 0 0 1-4-4v-2"></path>
			</svg>`
		},
		{
			href: '/admin/advertisements',
			label: 'Advertisements',
			icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect>
				<line x1="8" y1="21" x2="16" y2="21"></line>
				<line x1="12" y1="17" x2="12" y2="21"></line>
			</svg>`
		},
		{
			href: '/admin/placements',
			label: 'Ad Placements',
			icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path>
				<polyline points="3.27,6.96 12,12.01 20.73,6.96"></polyline>
				<line x1="12" y1="22.08" x2="12" y2="12"></line>
			</svg>`
		},
		{
			href: '/admin/settings',
			label: 'Settings',
			icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<circle cx="12" cy="12" r="3"></circle>
				<path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1 1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
			</svg>`
		},
		{
			href: '/admin/design-system',
			label: 'Design System',
			icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
				<circle cx="9" cy="9" r="2"></circle>
				<path d="M21 15l-3.086-3.086a2 2 0 0 0-2.828 0L6 21"></path>
				<path d="M14 14l-3-3"></path>
			</svg>`
		}
	];
</script>

<svelte:head>
	<title>Admin Dashboard - BOME</title>
	<meta name="description" content="BOME Admin Dashboard" />
</svelte:head>

{#if !isLoaded}
	<div class="loading-screen">
		<LoadingSpinner size="large" color="primary" />
		<p>Loading Admin Dashboard...</p>
	</div>
{:else}
	<div class="admin-layout">
		<!-- Sidebar -->
		<aside class="sidebar glass" class:collapsed={!isSidebarOpen}>
			<div class="sidebar-header">
				<div class="brand">
					<div class="brand-logo">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M12 2L2 7l10 5 10-5-10-5z"></path>
							<path d="M2 17l10 5 10-5"></path>
							<path d="M2 12l10 5 10-5"></path>
						</svg>
					</div>
					{#if isSidebarOpen}
						<span class="brand-text">BOME Admin</span>
					{/if}
				</div>
				<button class="sidebar-toggle" on:click={toggleSidebar}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M3 12h18"></path>
						<path d="M3 6h18"></path>
						<path d="M3 18h18"></path>
					</svg>
				</button>
			</div>

			<nav class="sidebar-nav">
				{#each navItems as item}
					<a href={item.href} class="nav-item" class:active={$page.url.pathname === item.href}>
						<div class="nav-icon" class:active={$page.url.pathname === item.href}>
							{@html item.icon}
						</div>
						{#if isSidebarOpen}
							<span class="nav-label">{item.label}</span>
						{/if}
					</a>
				{/each}
			</nav>

			<div class="sidebar-footer">
				<div class="user-info">
					<div class="user-avatar">
						{user?.firstName?.charAt(0) || 'A'}
					</div>
					{#if isSidebarOpen}
						<div class="user-details">
							<div class="user-name">{user?.firstName} {user?.lastName}</div>
							<div class="user-role">
								{#if user?.roles && user.roles.length > 0}
									{user.roles[0].name}
								{:else}
									Administrator
								{/if}
							</div>
							{#if user?.roles && user.roles.some((r: any) => r.id === 'super-administrator')}
								<div class="super-admin-badge">Super Admin</div>
							{/if}
						</div>
					{/if}
				</div>
				<button class="logout-btn" on:click={async () => { await auth.logout(); goto('/login'); }}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
						<polyline points="16,17 21,12 16,7"></polyline>
						<line x1="21" y1="12" x2="9" y2="12"></line>
					</svg>
					{#if isSidebarOpen}
						<span>Logout</span>
					{/if}
				</button>
			</div>
		</aside>

		<!-- Main Content -->
		<main class="main-content" class:sidebar-open={isSidebarOpen}>
			<header class="content-header glass">
				<div class="header-left">
					<button class="mobile-toggle" on:click={toggleSidebar}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M3 12h18"></path>
							<path d="M3 6h18"></path>
							<path d="M3 18h18"></path>
						</svg>
					</button>
					<h1 class="page-title">Admin Dashboard</h1>
				</div>
				<div class="header-right">
					<a href="/" class="home-btn" title="Go to Home">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path>
							<polyline points="9,22 9,12 15,12 15,22"></polyline>
						</svg>
					</a>
					<ThemeToggle />
				</div>
			</header>

			<div class="content-area">
				<slot />
			</div>
		</main>
	</div>

	<ToastContainer />
{/if}

<style>
	.loading-screen {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		background: var(--primary-gradient);
		color: var(--white);
		gap: var(--space-lg);
	}

	.loading-screen p {
		font-size: var(--text-lg);
		opacity: 0.8;
	}

	.admin-layout {
		display: flex;
		min-height: 100vh;
		background: var(--bg-secondary);
	}

	/* Sidebar */
	.sidebar {
		width: 280px;
		height: 100vh;
		position: fixed;
		left: 0;
		top: 0;
		z-index: var(--z-fixed);
		display: flex;
		flex-direction: column;
		transition: all var(--transition-normal);
		border-right: 1px solid rgba(255, 255, 255, 0.1);
	}

	.sidebar.collapsed {
		width: 80px;
	}

	.sidebar-header {
		padding: var(--space-xl);
		display: flex;
		align-items: center;
		justify-content: space-between;
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
	}

	.brand {
		display: flex;
		align-items: center;
		gap: var(--space-md);
	}

	.brand-logo {
		width: 40px;
		height: 40px;
		background: var(--primary-gradient);
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: var(--shadow-md);
	}

	.brand-logo svg {
		width: 24px;
		height: 24px;
		color: var(--white);
	}

	.brand-text {
		font-size: var(--text-xl);
		font-weight: 700;
		color: var(--text-primary);
		font-family: var(--font-display);
	}

	.sidebar-toggle {
		width: 32px;
		height: 32px;
		border: none;
		background: var(--bg-glass);
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		transition: all var(--transition-normal);
		color: var(--text-primary);
	}

	.sidebar-toggle:hover {
		background: var(--bg-glass-dark);
		transform: scale(1.05);
	}

	.sidebar-toggle svg {
		width: 18px;
		height: 18px;
	}

	.sidebar-nav {
		flex: 1;
		padding: var(--space-lg);
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}

	.nav-item {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		padding: var(--space-md);
		border-radius: var(--radius-lg);
		text-decoration: none;
		color: var(--text-secondary);
		transition: all var(--transition-normal);
		cursor: pointer;
	}

	.nav-item:hover {
		background: var(--bg-glass);
		color: var(--text-primary);
		transform: translateX(4px);
	}

	.nav-item.active {
		background: var(--primary-gradient);
		color: var(--white);
		box-shadow: var(--shadow-md);
	}

	.nav-icon {
		width: 24px;
		height: 24px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.nav-icon svg {
		width: 20px;
		height: 20px;
	}

	.nav-label {
		font-weight: 500;
		white-space: nowrap;
	}

	.sidebar-footer {
		padding: var(--space-lg);
		border-top: 1px solid rgba(255, 255, 255, 0.1);
	}

	.user-info {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		margin-bottom: var(--space-lg);
	}

	.user-avatar {
		width: 40px;
		height: 40px;
		background: var(--primary-gradient);
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 600;
		color: var(--white);
		font-size: var(--text-sm);
	}

	.user-details {
		flex: 1;
	}

	.user-name {
		font-weight: 600;
		color: var(--text-primary);
		font-size: var(--text-sm);
	}

	.user-role {
		font-size: var(--text-xs);
		color: var(--text-secondary);
	}

	.super-admin-badge {
		background: linear-gradient(135deg, #dc2626, #991b1b);
		color: white;
		padding: var(--space-1) var(--space-2);
		border-radius: var(--radius-sm);
		font-size: var(--text-xs);
		font-weight: 600;
		margin-top: var(--space-1);
		text-align: center;
		box-shadow: 0 2px 4px rgba(220, 38, 38, 0.2);
		animation: glow 2s ease-in-out infinite alternate;
	}

	@keyframes glow {
		from {
			box-shadow: 0 2px 4px rgba(220, 38, 38, 0.2);
		}
		to {
			box-shadow: 0 2px 8px rgba(220, 38, 38, 0.4), 0 0 16px rgba(220, 38, 38, 0.1);
		}
	}

	.logout-btn {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		width: 100%;
		padding: var(--space-md);
		border: none;
		background: var(--bg-glass);
		border-radius: var(--radius-lg);
		color: var(--text-secondary);
		cursor: pointer;
		transition: all var(--transition-normal);
	}

	.logout-btn:hover {
		background: var(--error);
		color: var(--white);
		transform: translateY(-2px);
	}

	.logout-btn svg {
		width: 18px;
		height: 18px;
	}

	/* Main Content */
	.main-content {
		flex: 1;
		margin-left: 280px;
		transition: all var(--transition-normal);
	}

	.main-content.sidebar-open {
		margin-left: 280px;
	}

	.main-content:not(.sidebar-open) {
		margin-left: 80px;
	}

	.content-header {
		position: sticky;
		top: 0;
		z-index: var(--z-sticky);
		padding: var(--space-xl);
		display: flex;
		align-items: center;
		justify-content: space-between;
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
	}

	.header-left {
		display: flex;
		align-items: center;
		gap: var(--space-lg);
	}

	.mobile-toggle {
		display: none;
		width: 40px;
		height: 40px;
		border: none;
		background: var(--bg-glass);
		border-radius: var(--radius-lg);
		align-items: center;
		justify-content: center;
		cursor: pointer;
		transition: all var(--transition-normal);
		color: var(--text-primary);
	}

	.mobile-toggle:hover {
		background: var(--bg-glass-dark);
		transform: scale(1.05);
	}

	.mobile-toggle svg {
		width: 20px;
		height: 20px;
	}

	.page-title {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--text-primary);
		margin: 0;
	}

	.header-right {
		display: flex;
		align-items: center;
		gap: var(--space-md);
	}

	.home-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 40px;
		height: 40px;
		border-radius: var(--radius-lg);
		background: var(--bg-glass);
		color: var(--text-primary);
		text-decoration: none;
		transition: all var(--transition-normal);
		cursor: pointer;
	}

	.home-btn:hover {
		background: var(--bg-glass-dark);
		transform: translateY(-2px);
		box-shadow: var(--shadow-md);
	}

	.home-btn svg {
		width: 20px;
		height: 20px;
	}

	.content-area {
		padding: var(--space-2xl);
	}

	/* Responsive */
	@media (max-width: 1024px) {
		.sidebar {
			transform: translateX(-100%);
		}

		.sidebar.open {
			transform: translateX(0);
		}

		.main-content {
			margin-left: 0;
		}

		.mobile-toggle {
			display: flex;
		}

		.sidebar-toggle {
			display: none;
		}
	}

	@media (max-width: 768px) {
		.content-header {
			padding: var(--space-lg);
		}

		.content-area {
			padding: var(--space-lg);
		}

		.page-title {
			font-size: var(--text-2xl);
		}
	}

	@media (max-width: 480px) {
		.content-header {
			padding: var(--space-md);
		}

		.content-area {
			padding: var(--space-md);
		}

		.page-title {
			font-size: var(--text-xl);
		}
	}
</style> 