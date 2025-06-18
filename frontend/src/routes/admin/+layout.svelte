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

	// Subscribe to auth store
	auth.subscribe(state => {
		isAuthenticated = state.isAuthenticated;
		user = state.user;
		isAdmin = user?.role === 'admin';
	});

	onMount(() => {
		// Check admin access
		if (!isAuthenticated) {
			goto('/login');
			return;
		}

		if (!isAdmin) {
			goto('/');
			return;
		}

		setTimeout(() => {
			isLoaded = true;
		}, 100);
	});

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
			href: '/admin/content',
			label: 'Content',
			icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
				<polyline points="14,2 14,8 20,8"></polyline>
				<line x1="16" y1="13" x2="8" y2="13"></line>
				<line x1="16" y1="17" x2="8" y2="17"></line>
				<polyline points="10,9 9,9 8,9"></polyline>
			</svg>`
		},
		{
			href: '/admin/settings',
			label: 'Settings',
			icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<circle cx="12" cy="12" r="3"></circle>
				<path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1 1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
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
							<div class="user-role">Administrator</div>
						</div>
					{/if}
				</div>
				<button class="logout-btn" on:click={() => auth.logout()}>
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