<script lang="ts">
	import { auth } from '$lib/auth';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	let user: any = null;
	let isAuthenticated = false;
	let mobileMenuOpen = false;
	let userMenuOpen = false;

	onMount(() => {
		auth.subscribe((state) => {
			user = state.user;
			isAuthenticated = state.isAuthenticated;
		});
	});

	async function handleLogout() {
		await auth.logout();
		goto('/login');
		mobileMenuOpen = false;
		userMenuOpen = false;
	}

	function toggleMobileMenu() {
		mobileMenuOpen = !mobileMenuOpen;
		if (mobileMenuOpen) {
			userMenuOpen = false;
		}
	}

	function toggleUserMenu() {
		userMenuOpen = !userMenuOpen;
		if (userMenuOpen) {
			mobileMenuOpen = false;
		}
	}

	function closeMenus() {
		mobileMenuOpen = false;
		userMenuOpen = false;
	}

	// Close menus when clicking outside
	function handleClickOutside(event: MouseEvent) {
		const target = event.target as HTMLElement;
		if (!target.closest('.nav-container')) {
			closeMenus();
		}
	}
</script>

<svelte:window on:click={handleClickOutside} />

<header class="header">
	<nav class="nav-container">
		<div class="nav-brand">
			<a href="/" class="brand-link">
				<h1>BOME</h1>
				<span>Book of Mormon Evidences</span>
			</a>
		</div>

		<!-- Desktop Navigation -->
		<div class="nav-menu desktop-menu">
			<a href="/" class="nav-link">Home</a>
			<a href="/videos" class="nav-link">Videos</a>
			<a href="/categories" class="nav-link">Categories</a>
			<a href="/about" class="nav-link">About</a>
		</div>

		<!-- Desktop Auth -->
		<div class="nav-auth desktop-auth">
			{#if isAuthenticated}
				<div class="user-menu-container">
					<button class="user-menu-trigger" on:click={toggleUserMenu}>
						<span class="user-avatar">
							{user?.firstName?.charAt(0) || user?.email?.charAt(0) || 'U'}
						</span>
						<span class="user-name">{user?.firstName || 'User'}</span>
						<span class="dropdown-arrow">▼</span>
					</button>
					
					{#if userMenuOpen}
						<div class="user-dropdown">
							<a href="/profile" class="dropdown-item">Profile</a>
							<a href="/favorites" class="dropdown-item">Favorites</a>
							<a href="/settings" class="dropdown-item">Settings</a>
							<hr class="dropdown-divider" />
							<button class="dropdown-item logout-btn" on:click={handleLogout}>
								Logout
							</button>
						</div>
					{/if}
				</div>
			{:else}
				<a href="/login" class="btn-primary">Login</a>
				<a href="/register" class="btn-secondary">Register</a>
			{/if}
		</div>

		<!-- Mobile Menu Button -->
		<button class="mobile-menu-btn" on:click={toggleMobileMenu} aria-label="Toggle menu">
			<div class="hamburger {mobileMenuOpen ? 'open' : ''}">
				<span></span>
				<span></span>
				<span></span>
			</div>
		</button>
	</nav>

	<!-- Mobile Menu -->
	{#if mobileMenuOpen}
		<div class="mobile-menu">
			<div class="mobile-menu-content">
				<div class="mobile-nav-links">
					<a href="/" class="mobile-nav-link" on:click={closeMenus}>Home</a>
					<a href="/videos" class="mobile-nav-link" on:click={closeMenus}>Videos</a>
					<a href="/categories" class="mobile-nav-link" on:click={closeMenus}>Categories</a>
					<a href="/about" class="mobile-nav-link" on:click={closeMenus}>About</a>
				</div>

				<div class="mobile-auth">
					{#if isAuthenticated}
						<div class="mobile-user-info">
							<div class="mobile-user-avatar">
								{user?.firstName?.charAt(0) || user?.email?.charAt(0) || 'U'}
							</div>
							<div class="mobile-user-details">
								<span class="mobile-user-name">{user?.firstName} {user?.lastName}</span>
								<span class="mobile-user-email">{user?.email}</span>
							</div>
						</div>
						<div class="mobile-user-links">
							<a href="/profile" class="mobile-user-link" on:click={closeMenus}>Profile</a>
							<a href="/favorites" class="mobile-user-link" on:click={closeMenus}>Favorites</a>
							<a href="/settings" class="mobile-user-link" on:click={closeMenus}>Settings</a>
							<button class="mobile-user-link logout-btn" on:click={handleLogout}>Logout</button>
						</div>
					{:else}
						<div class="mobile-auth-buttons">
							<a href="/login" class="btn-primary full-width" on:click={closeMenus}>Login</a>
							<a href="/register" class="btn-secondary full-width" on:click={closeMenus}>Register</a>
						</div>
					{/if}
				</div>
			</div>
		</div>
	{/if}
</header>

<style>
	.header {
		background: var(--card-bg);
		box-shadow: 
			0 4px 8px var(--shadow-dark),
			0 -2px 4px var(--shadow-light);
		position: sticky;
		top: 0;
		z-index: 1000;
	}

	.nav-container {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1rem 2rem;
		max-width: 1200px;
		margin: 0 auto;
		position: relative;
	}

	.nav-brand {
		flex-shrink: 0;
	}

	.brand-link {
		text-decoration: none;
		color: inherit;
		display: flex;
		flex-direction: column;
		align-items: flex-start;
	}

	.nav-brand h1 {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--accent-color);
		margin: 0;
		line-height: 1;
	}

	.nav-brand span {
		font-size: 0.7rem;
		color: var(--text-secondary);
		line-height: 1;
	}

	.nav-menu {
		display: flex;
		gap: 2rem;
		align-items: center;
	}

	.nav-link {
		color: var(--text-primary);
		text-decoration: none;
		font-weight: 500;
		transition: color 0.2s ease;
		padding: 0.5rem 0;
		position: relative;
	}

	.nav-link:hover {
		color: var(--accent-color);
	}

	.nav-link::after {
		content: '';
		position: absolute;
		bottom: 0;
		left: 0;
		width: 0;
		height: 2px;
		background: var(--accent-color);
		transition: width 0.2s ease;
	}

	.nav-link:hover::after {
		width: 100%;
	}

	.nav-auth {
		display: flex;
		gap: 1rem;
		align-items: center;
	}

	.user-menu-container {
		position: relative;
	}

	.user-menu-trigger {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		background: none;
		border: none;
		color: var(--text-primary);
		font-size: 1rem;
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 8px;
		transition: background-color 0.2s ease;
	}

	.user-menu-trigger:hover {
		background: var(--input-bg);
	}

	.user-avatar {
		width: 32px;
		height: 32px;
		background: var(--accent-color);
		color: white;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 600;
		font-size: 0.9rem;
	}

	.user-name {
		font-weight: 500;
	}

	.dropdown-arrow {
		font-size: 0.8rem;
		transition: transform 0.2s ease;
	}

	.user-menu-trigger:hover .dropdown-arrow {
		transform: rotate(180deg);
	}

	.user-dropdown {
		position: absolute;
		top: 100%;
		right: 0;
		background: var(--card-bg);
		border-radius: 12px;
		box-shadow: 
			8px 8px 16px var(--shadow-dark),
			-8px -8px 16px var(--shadow-light);
		padding: 0.5rem;
		min-width: 200px;
		z-index: 1001;
	}

	.dropdown-item {
		display: block;
		width: 100%;
		padding: 0.75rem 1rem;
		text-decoration: none;
		color: var(--text-primary);
		border: none;
		background: none;
		font-size: 1rem;
		text-align: left;
		cursor: pointer;
		border-radius: 8px;
		transition: background-color 0.2s ease;
	}

	.dropdown-item:hover {
		background: var(--input-bg);
	}

	.dropdown-divider {
		border: none;
		height: 1px;
		background: var(--border-color);
		margin: 0.5rem 0;
	}

	.logout-btn {
		color: var(--error-text);
	}

	.mobile-menu-btn {
		display: none;
		background: none;
		border: none;
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 8px;
		transition: background-color 0.2s ease;
	}

	.mobile-menu-btn:hover {
		background: var(--input-bg);
	}

	.hamburger {
		display: flex;
		flex-direction: column;
		gap: 4px;
		width: 24px;
		height: 24px;
		justify-content: center;
	}

	.hamburger span {
		width: 100%;
		height: 2px;
		background: var(--text-primary);
		border-radius: 1px;
		transition: all 0.3s ease;
	}

	.hamburger.open span:nth-child(1) {
		transform: rotate(45deg) translate(6px, 6px);
	}

	.hamburger.open span:nth-child(2) {
		opacity: 0;
	}

	.hamburger.open span:nth-child(3) {
		transform: rotate(-45deg) translate(6px, -6px);
	}

	.mobile-menu {
		position: fixed;
		top: 100%;
		left: 0;
		right: 0;
		background: var(--card-bg);
		box-shadow: 
			0 4px 8px var(--shadow-dark),
			0 -2px 4px var(--shadow-light);
		z-index: 999;
		animation: slideDown 0.3s ease;
	}

	@keyframes slideDown {
		from {
			transform: translateY(-100%);
			opacity: 0;
		}
		to {
			transform: translateY(0);
			opacity: 1;
		}
	}

	.mobile-menu-content {
		padding: 2rem;
	}

	.mobile-nav-links {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		margin-bottom: 2rem;
	}

	.mobile-nav-link {
		color: var(--text-primary);
		text-decoration: none;
		font-size: 1.2rem;
		font-weight: 500;
		padding: 1rem 0;
		border-bottom: 1px solid var(--border-color);
		transition: color 0.2s ease;
	}

	.mobile-nav-link:hover {
		color: var(--accent-color);
	}

	.mobile-user-info {
		display: flex;
		align-items: center;
		gap: 1rem;
		margin-bottom: 1.5rem;
		padding: 1rem;
		background: var(--input-bg);
		border-radius: 12px;
	}

	.mobile-user-avatar {
		width: 48px;
		height: 48px;
		background: var(--accent-color);
		color: white;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 600;
		font-size: 1.2rem;
	}

	.mobile-user-details {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.mobile-user-name {
		font-weight: 600;
		color: var(--text-primary);
	}

	.mobile-user-email {
		font-size: 0.9rem;
		color: var(--text-secondary);
	}

	.mobile-user-links {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.mobile-user-link {
		color: var(--text-primary);
		text-decoration: none;
		font-size: 1rem;
		padding: 0.75rem 1rem;
		border-radius: 8px;
		transition: background-color 0.2s ease;
		background: none;
		border: none;
		cursor: pointer;
		text-align: left;
	}

	.mobile-user-link:hover {
		background: var(--input-bg);
	}

	.mobile-auth-buttons {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.full-width {
		width: 100%;
		text-align: center;
	}

	.btn-primary,
	.btn-secondary {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 12px;
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
		text-decoration: none;
		display: inline-block;
		transition: all 0.2s ease;
	}

	.btn-primary {
		background: var(--accent-color);
		color: white;
		box-shadow: 
			4px 4px 8px var(--shadow-dark),
			-2px -2px 4px var(--shadow-light);
	}

	.btn-primary:hover {
		background: var(--accent-hover);
		transform: translateY(-2px);
		box-shadow: 
			6px 6px 12px var(--shadow-dark),
			-3px -3px 6px var(--shadow-light);
	}

	.btn-secondary {
		background: var(--card-bg);
		color: var(--text-primary);
		box-shadow: 
			4px 4px 8px var(--shadow-dark),
			-2px -2px 4px var(--shadow-light);
	}

	.btn-secondary:hover {
		transform: translateY(-2px);
		box-shadow: 
			6px 6px 12px var(--shadow-dark),
			-3px -3px 6px var(--shadow-light);
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.nav-container {
			padding: 1rem;
		}

		.desktop-menu,
		.desktop-auth {
			display: none;
		}

		.mobile-menu-btn {
			display: block;
		}

		.nav-brand h1 {
			font-size: 1.3rem;
		}

		.nav-brand span {
			font-size: 0.6rem;
		}
	}

	@media (min-width: 769px) {
		.mobile-menu-btn {
			display: none;
		}
	}

	/* Touch-friendly improvements */
	@media (hover: none) and (pointer: coarse) {
		.nav-link,
		.dropdown-item,
		.mobile-nav-link,
		.mobile-user-link {
			min-height: 44px;
			display: flex;
			align-items: center;
		}

		.user-menu-trigger {
			min-height: 44px;
		}

		.mobile-menu-btn {
			min-width: 44px;
			min-height: 44px;
		}
	}
</style> 