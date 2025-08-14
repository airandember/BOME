<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import { page } from '$app/stores';
	import { theme } from '$lib/theme';

	let isAdmin = false;
	let userRole = '';
	let isSuperAdmin = false;
	let isDark = false;

	// Subscribe to theme changes
	theme.subscribe(state => {
		isDark = state.isDark;
	});

	// Subsites configuration
	const subsites = [
		{
			id: 'streaming',
			name: 'Streaming',
			icon: 'play',
			path: '/admin/streaming',
			description: 'Video streaming platform',
			status: 'active'
		},
		{
			id: 'articles',
			name: 'Articles',
			icon: 'file-text',
			path: '/admin/articles',
			description: 'Blog and articles platform',
			status: 'under_construction'
		},
		{
			id: 'expo',
			name: 'Expo & Tours',
			icon: 'map-pin',
			path: '/admin/expo',
			description: 'Events and tours platform',
			status: 'under_construction'
		}
	];

	onMount(() => {
		// Initial setup will be handled by reactive statements
	});

	// Reactive statements to handle auth state changes
	$: if ($auth.isAuthenticated && $auth.user) {
		const adminRoles = [
			'super_admin', 'system_admin', 'content_manager', 
			'articles_manager', 'youtube_manager', 'streaming_manager',
			'events_manager', 'advertisement_manager', 'user_manager',
			'analytics_manager', 'financial_admin', 'admin'
		];
		isAdmin = adminRoles.includes($auth.user.role);
		userRole = $auth.user.role;
		isSuperAdmin = $auth.user.role === 'super_admin';
	} else {
		isAdmin = false;
		userRole = '';
		isSuperAdmin = false;
	}

	// Reactive redirect logic
	$: if ($page.url.pathname !== '/admin' && !isAdmin && $auth.isAuthenticated) {
		// User is authenticated but not admin, redirect to login
		goto('/admin');
	} else if ($page.url.pathname !== '/admin' && !$auth.isAuthenticated) {
		// User is not authenticated, redirect to login
		goto('/admin');
	}

	function isAdminUser(user: any): boolean {
		if (!user) return false;
		const adminRoles = [
			'super_admin', 'system_admin', 'content_manager', 
			'articles_manager', 'youtube_manager', 'streaming_manager',
			'events_manager', 'advertisement_manager', 'user_manager',
			'analytics_manager', 'financial_admin', 'admin'
		];
		return adminRoles.includes(user.role);
	}

	function getUserPermissions(role: string): string[] {
		const permissionMap: Record<string, string[]> = {
			'super_admin': ['*'], // All permissions
			'system_admin': ['system:read', 'system:update', 'system:manage', 'analytics:read', 'analytics:export', 'monitoring:read'],
			'content_manager': ['content:read', 'content:create', 'content:update', 'content:delete', 'videos:manage', 'analytics:read'],
			'articles_manager': ['articles:read', 'articles:create', 'articles:update', 'articles:delete', 'analytics:read'],
			'youtube_manager': ['videos:read', 'videos:create', 'videos:update', 'videos:delete', 'analytics:read'],
			'streaming_manager': ['videos:read', 'videos:create', 'videos:update', 'videos:delete', 'analytics:read'],
			'events_manager': ['events:read', 'events:create', 'events:update', 'events:delete', 'analytics:read'],
			'advertisement_manager': ['advertisements:read', 'advertisements:create', 'advertisements:update', 'advertisements:delete', 'analytics:read'],
			'user_manager': ['users:read', 'users:create', 'users:update', 'users:delete', 'analytics:read'],
			'analytics_manager': ['analytics:read', 'analytics:export', 'analytics:manage', 'monitoring:read'],
			'financial_admin': ['financial:read', 'financial:manage', 'analytics:read'],
			'admin': ['*'] // Legacy admin - all permissions
		};
		return permissionMap[role] || [];
	}

	function hasPermission(permission: string): boolean {
		const permissions = getUserPermissions(userRole);
		return permissions.includes('*') || permissions.includes(permission);
	}

	function getSubsiteIcon(iconName: string) {
		switch (iconName) {
			case 'play':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<polygon points="23,7 16,12 23,17 23,7"></polygon>
					<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
				</svg>`;
			case 'file-text':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
					<polyline points="14,2 14,8 20,8"></polyline>
					<line x1="16" y1="13" x2="8" y2="13"></line>
					<line x1="16" y1="17" x2="8" y2="17"></line>
					<polyline points="10,9 9,9 8,9"></polyline>
				</svg>`;
			case 'map-pin':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path>
					<circle cx="12" cy="10" r="3"></circle>
				</svg>`;
			default:
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="12" cy="12" r="10"></circle>
					<line x1="12" y1="16" x2="12" y2="12"></line>
					<line x1="12" y1="8" x2="12.01" y2="8"></line>
				</svg>`;
		}
	}

	const toggleTheme = () => {
		theme.toggle();
	};

	async function handleLogout() {
		await auth.logout();
		goto('/admin');
	}
</script>

<svelte:head>
	<title>BOME Admin</title>
	<meta name="description" content="BOME administrative interface" />
</svelte:head>

{#if $page.url.pathname === '/admin'}
	<!-- Login page - no navigation -->
	<slot />
{:else if isAdmin}
	<!-- Admin pages with navigation -->
	<div class="admin-layout">
		<!-- Sidebar Navigation -->
		<nav class="admin-sidebar">
			<div class="sidebar-header">
				<div class="brand-logo">
					<a href="/" aria-label="BOME Admin">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M12 2L2 7l10 5 10-5-10-5z"></path>
							<path d="M2 17l10 5 10-5"></path>
							<path d="M2 12l10 5 10-5"></path>
						</svg>
					</a>
				</div>
				<div class="brand-text">
					<h2>BOME Admin</h2>
					
					{#if isSuperAdmin}
						<span class="super-admin-badge">SUPER ADMIN</span>
						{:else}<span class="role-badge">{userRole.replace('_', ' ').toUpperCase()}</span>
					{/if}
				</div>
				<button class="theme-toggle glass" on:click={toggleTheme} aria-label="Toggle theme">
					<div class="toggle-container">
						<svg class="sun-icon" class:active={!isDark} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<circle cx="12" cy="12" r="5"></circle>
							<line x1="12" y1="1" x2="12" y2="3"></line>
							<line x1="12" y1="21" x2="12" y2="23"></line>
							<line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line>
							<line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line>
							<line x1="1" y1="12" x2="3" y2="12"></line>
							<line x1="21" y1="12" x2="23" y2="12"></line>
							<line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line>
							<line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>
						</svg>
						
						<svg class="moon-icon" class:active={isDark} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path>
						</svg>
						
						<div class="toggle-slider" class:dark={isDark}></div>
					</div>
				</button>
			</div>

			<div class="sidebar-nav">
				<!-- Main Dashboard -->
				<a href="/admin/dashboard" class="nav-item" class:active={$page.url.pathname === '/admin/dashboard'}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<rect x="3" y="3" width="7" height="7"></rect>
						<rect x="14" y="3" width="7" height="7"></rect>
						<rect x="14" y="14" width="7" height="7"></rect>
						<rect x="3" y="14" width="7" height="7"></rect>
					</svg>
					Hub Dashboard
				</a>

				<!-- Subsites Section -->
				<div class="nav-section">
					<div class="section-header">Subsites</div>
					
					{#each subsites as subsite}
						<a href={subsite.path} class="nav-item subsite-item" class:active={$page.url.pathname.startsWith(subsite.path)} class:under-construction={subsite.status === 'under_construction'}>
							<div class="subsite-icon" style="innerHTML={getSubsiteIcon(subsite.icon)}"></div>
							<div class="subsite-info">
								<div class="subsite-name">{subsite.name}</div>
								<div class="subsite-status {subsite.status}">
									{#if subsite.status === 'under_construction'}
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
											<line x1="12" y1="9" x2="12" y2="13"></line>
											<line x1="12" y1="17" x2="12.01" y2="17"></line>
										</svg>
										Under Construction
									{:else}
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<polyline points="20,6 9,17 4,12"></polyline>
										</svg>
										Active
									{/if}
								</div>
							</div>
						</a>
					{/each}
				</div>

				<!-- Analytics & Monitoring -->
				{#if hasPermission('analytics:read')}
					<div class="nav-section">
						<div class="section-header">Analytics & Monitoring</div>
						
						<a href="/admin/analytics" class="nav-item" class:active={$page.url.pathname.startsWith('/admin/analytics')}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polyline points="22,12 18,12 15,21 9,3 6,12 2,12"></polyline>
							</svg>
							Analytics
						</a>

						{#if hasPermission('monitoring:read')}
							<a href="/admin/monitoring" class="nav-item" class:active={$page.url.pathname.startsWith('/admin/monitoring')}>
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect>
									<line x1="8" y1="21" x2="16" y2="21"></line>
									<line x1="12" y1="17" x2="12" y2="21"></line>
								</svg>
								Server Monitoring
							</a>
						{/if}
					</div>
				{/if}

				<!-- Management Section -->
				<div class="nav-section">
					<div class="section-header">Management</div>
					
					{#if hasPermission('users:read')}
						<a href="/admin/users" class="nav-item" class:active={$page.url.pathname.startsWith('/admin/users')}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
								<circle cx="9" cy="7" r="4"></circle>
								<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
								<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
							</svg>
							Users
						</a>
					{/if}

					{#if hasPermission('advertisements:read')}
						<a href="/admin/advertisers" class="nav-item" class:active={$page.url.pathname.startsWith('/admin/advertisers')}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
								<circle cx="9" cy="7" r="4"></circle>
								<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
								<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
							</svg>
							Advertisers
						</a>
					{/if}

					{#if hasPermission('advertisements:read')}
						<a href="/admin/advertisements" class="nav-item" class:active={$page.url.pathname.startsWith('/admin/advertisements')}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
								<circle cx="8.5" cy="8.5" r="1.5"></circle>
								<polyline points="21,15 16,10 5,21"></polyline>
							</svg>
							Advertisements
						</a>
					{/if}

					{#if hasPermission('system:read')}
						<a href="/admin/system" class="nav-item" class:active={$page.url.pathname.startsWith('/admin/system')}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<circle cx="12" cy="12" r="3"></circle>
								<path d="M12 1v6m0 6v6"></path>
								<path d="M21 12h-6m-6 0H3"></path>
							</svg>
							System Settings
						</a>
					{/if}
				</div>
			</div>

			<div class="sidebar-footer">
				<div class="user-info">
					<div class="user-avatar">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
							<circle cx="12" cy="7" r="4"></circle>
						</svg>
					</div>
					<div class="user-details">
						<div class="user-name">{$auth.user?.first_name || 'Admin'}</div>
						<div class="user-email">{$auth.user?.email || ''}</div>
					</div>
				</div>
				<button class="logout-button" on:click={handleLogout}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
						<polyline points="16,17 21,12 16,7"></polyline>
						<line x1="21" y1="12" x2="9" y2="12"></line>
					</svg>
					Logout
				</button>
			</div>
		</nav>

		<!-- Main Content -->
		<main class="admin-main">
			<slot />
		</main>
	</div>
{:else}
	<!-- Not admin - redirect to login -->
	<div class="redirect-message">
		<p>Redirecting to admin login...</p>
	</div>
{/if}

<style>
	.admin-layout {
		display: flex;
		min-height: 100vh;
		background: var(--bg-primary);
	}

	.admin-sidebar {
		width: 320px;
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-right: 1px solid rgba(255, 255, 255, 0.1);
		display: flex;
		flex-direction: column;
		position: fixed;
		height: 100vh;
		z-index: 100;
		overflow-y: auto;
	}

	.sidebar-header {
		padding: 2rem 1.5rem;
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.brand-logo {
		width: 40px;
		height: 40px;
		background: var(--primary-gradient);
		border-radius: 10px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.brand-logo svg {
		width: 20px;
		height: 20px;
		color: white;
	}

	.brand-text {
		flex: 1;
		min-width: 0;
	}

	.brand-text h2 {
		font-size: 1.25rem;
		font-weight: 700;
		color: var(--text-primary);
		margin: 0 0 0.25rem 0;
	}

	.role-badge {
		background: var(--primary-gradient);
		color: white;
		padding: 0.125rem 0.5rem;
		border-radius: 12px;
		font-size: 0.7rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		display: block;
	}

	.super-admin-badge {
		background: linear-gradient(135deg, #ff6b6b, #ee5a24);
		color: white;
		padding: 0.125rem 0.5rem;
		border-radius: 12px;
		font-size: 0.7rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		display: block;
	}

	/* Theme Toggle Styles */
	.theme-toggle {
		width: 48px;
		height: 48px;
		border-radius: var(--radius-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
		cursor: pointer;
		transition: all var(--transition-normal);
		position: relative;
		overflow: hidden;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--bg-glass);
	}

	.theme-toggle:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}

	.theme-toggle:active {
		transform: translateY(0);
	}

	.toggle-container {
		position: relative;
		width: 32px;
		height: 32px;
		border-radius: var(--radius-lg);
		background: var(--bg-glass);
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
	}

	.sun-icon,
	.moon-icon {
		width: 18px;
		height: 18px;
		position: absolute;
		transition: all var(--transition-bounce);
		color: var(--text-primary);
	}

	.sun-icon {
		opacity: 1;
		transform: scale(1) rotate(0deg);
	}

	.sun-icon.active {
		opacity: 1;
		transform: scale(1) rotate(0deg);
		color: var(--warning);
	}

	.sun-icon:not(.active) {
		opacity: 0.5;
		transform: scale(0.8) rotate(-90deg);
	}

	.moon-icon {
		opacity: 0.5;
		transform: scale(0.8) rotate(90deg);
	}

	.moon-icon.active {
		opacity: 1;
		transform: scale(1) rotate(0deg);
		color: var(--accent);
	}

	.moon-icon:not(.active) {
		opacity: 0.5;
		transform: scale(0.8) rotate(90deg);
	}

	.toggle-slider {
		position: absolute;
		top: 2px;
		left: 2px;
		width: 12px;
		height: 12px;
		background: var(--primary-gradient);
		border-radius: 50%;
		transition: all var(--transition-bounce);
		box-shadow: var(--shadow-sm);
	}

	.toggle-slider.dark {
		transform: translateX(16px);
		background: var(--accent-gradient);
	}

	/* Hover effects */
	.theme-toggle:hover .toggle-slider {
		transform: scale(1.1);
	}

	.theme-toggle:hover .toggle-slider.dark {
		transform: translateX(16px) scale(1.1);
	}

	/* Focus styles */
	.theme-toggle:focus {
		outline: none;
		box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.2);
	}

	/* Responsive */
	@media (max-width: 768px) {
		.theme-toggle {
			width: 44px;
			height: 44px;
		}

		.toggle-container {
			width: 28px;
			height: 28px;
		}

		.sun-icon,
		.moon-icon {
			width: 16px;
			height: 16px;
		}

		.toggle-slider {
			width: 10px;
			height: 10px;
		}

		.toggle-slider.dark {
			transform: translateX(14px);
		}

		.theme-toggle:hover .toggle-slider.dark {
			transform: translateX(14px) scale(1.1);
		}
	}

	@media (max-width: 480px) {
		.theme-toggle {
			width: 40px;
			height: 40px;
		}

		.toggle-container {
			width: 24px;
			height: 24px;
		}

		.sun-icon,
		.moon-icon {
			width: 14px;
			height: 14px;
		}

		.toggle-slider {
			width: 8px;
			height: 8px;
		}

		.toggle-slider.dark {
			transform: translateX(12px);
		}

		.theme-toggle:hover .toggle-slider.dark {
			transform: translateX(12px) scale(1.1);
		}
	}

	.sidebar-nav {
		flex: 1;
		padding: 1rem 0;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.nav-section {
		margin-bottom: 1rem;
	}

	.section-header {
		padding: 0.5rem 1.5rem;
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		margin-bottom: 0.5rem;
	}

	.nav-item {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem 1.5rem;
		color: var(--text-secondary);
		text-decoration: none;
		transition: all 0.3s ease;
		border-radius: 0 8px 8px 0;
		margin-right: 1rem;
	}

	.nav-item:hover {
		background: rgba(255, 255, 255, 0.1);
		color: var(--text-primary);
	}

	.nav-item.active {
		background: var(--primary);
		color: white;
	}

	.nav-item svg {
		width: 18px;
		height: 18px;
		flex-shrink: 0;
	}

	/* Subsites styling */
	.subsite-item {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 1rem 1.5rem;
	}

	.subsite-icon {
		width: 32px;
		height: 32px;
		background: var(--bg-glass-dark);
		border-radius: 8px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.subsite-icon svg {
		width: 16px;
		height: 16px;
		color: var(--text-secondary);
	}

	.subsite-info {
		flex: 1;
		min-width: 0;
	}

	.subsite-name {
		font-weight: 600;
		color: var(--text-primary);
		font-size: 0.9rem;
		margin-bottom: 0.125rem;
	}

	.subsite-status {
		display: flex;
		align-items: center;
		gap: 0.25rem;
		font-size: 0.7rem;
		font-weight: 500;
	}

	.subsite-status.active {
		color: #10b981;
	}

	.subsite-status.under_construction {
		color: #f59e0b;
	}

	.subsite-status svg {
		width: 12px;
		height: 12px;
	}

	.subsite-item.under-construction {
		opacity: 0.7;
	}

	.subsite-item.under-construction:hover {
		opacity: 1;
	}

	.sidebar-footer {
		padding: 1.5rem;
		border-top: 1px solid rgba(255, 255, 255, 0.1);
	}

	.user-info {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 1rem;
	}

	.user-avatar {
		width: 40px;
		height: 40px;
		background: var(--bg-glass-dark);
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.user-avatar svg {
		width: 20px;
		height: 20px;
		color: var(--text-secondary);
	}

	.user-details {
		flex: 1;
		min-width: 0;
	}

	.user-name {
		font-weight: 600;
		color: var(--text-primary);
		font-size: 0.9rem;
		margin-bottom: 0.125rem;
	}

	.user-email {
		font-size: 0.8rem;
		color: var(--text-secondary);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.logout-button {
		width: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		padding: 0.75rem;
		background: rgba(239, 68, 68, 0.1);
		border: 1px solid rgba(239, 68, 68, 0.3);
		border-radius: 8px;
		color: #fca5a5;
		cursor: pointer;
		transition: all 0.3s ease;
		font-size: 0.9rem;
	}

	.logout-button:hover {
		background: #ef4444;
		color: white;
		border-color: #ef4444;
	}

	.logout-button svg {
		width: 16px;
		height: 16px;
	}

	.admin-main {
		flex: 1;
		margin-left: 320px;
		overflow-y: auto;
	}

	.redirect-message {
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 100vh;
		background: var(--bg-primary);
		color: var(--text-primary);
	}

	@media (max-width: 768px) {
		.admin-sidebar {
			transform: translateX(-100%);
			transition: transform 0.3s ease;
		}

		.admin-sidebar.open {
			transform: translateX(0);
		}

		.admin-main {
			margin-left: 0;
		}
	}
</style> 
