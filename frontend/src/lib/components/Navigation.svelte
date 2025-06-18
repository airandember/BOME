<script lang="ts">
	import { auth } from '$lib/auth';
	import ThemeToggle from './ThemeToggle.svelte';

	let isMenuOpen = false;
	let isAuthenticated = false;
	let user: any = null;
	let isScrolled = false;

	// Subscribe to auth store
	auth.subscribe(state => {
		isAuthenticated = state.isAuthenticated;
		user = state.user;
	});

	// Handle scroll for navbar background
	if (typeof window !== 'undefined') {
		window.addEventListener('scroll', () => {
			isScrolled = window.scrollY > 50;
		});
	}

	const toggleMenu = () => {
		isMenuOpen = !isMenuOpen;
	};

	const closeMenu = () => {
		isMenuOpen = false;
	};

	const handleLogout = () => {
		auth.logout();
		closeMenu();
	};
</script>

<nav class="navigation" class:scrolled={isScrolled}>
	<div class="nav-container">
		<div class="nav-brand">
			<a href="/" class="brand-link">
				<div class="brand-logo">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M12 2L2 7l10 5 10-5-10-5z"></path>
						<path d="M2 17l10 5 10-5"></path>
						<path d="M2 12l10 5 10-5"></path>
					</svg>
				</div>
				<span class="brand-text">BOME</span>
			</a>
		</div>

		<div class="nav-menu" class:open={isMenuOpen}>
			<a href="/" class="nav-link" on:click={closeMenu}>
				<span>Home</span>
			</a>
			<a href="/videos" class="nav-link" on:click={closeMenu}>
				<span>Videos</span>
			</a>
			<a href="/categories" class="nav-link" on:click={closeMenu}>
				<span>Categories</span>
			</a>
			<a href="/about" class="nav-link" on:click={closeMenu}>
				<span>About</span>
			</a>
			<a href="/contact" class="nav-link" on:click={closeMenu}>
				<span>Contact</span>
			</a>
		</div>

		<div class="nav-actions">
			<ThemeToggle />
			
			{#if isAuthenticated}
				<div class="user-menu">
					<button class="user-button glass" on:click={toggleMenu}>
						<div class="user-avatar">
							{user?.firstName?.charAt(0) || 'U'}
						</div>
						<span class="user-name">{user?.firstName || 'User'}</span>
						<svg class="chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<polyline points="6,9 12,15 18,9"></polyline>
						</svg>
					</button>
					
					<div class="dropdown-menu glass" class:open={isMenuOpen}>
						<a href="/account" class="dropdown-item" on:click={closeMenu}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
								<circle cx="12" cy="7" r="4"></circle>
							</svg>
							<span>Profile</span>
						</a>
						<a href="/subscription" class="dropdown-item" on:click={closeMenu}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M20 7H4a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2z"></path>
								<polyline points="16,21 12,17 8,21"></polyline>
								<polyline points="12,17 12,3"></polyline>
							</svg>
							<span>Subscription</span>
						</a>
						<a href="/account/billing" class="dropdown-item" on:click={closeMenu}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<rect x="1" y="4" width="22" height="16" rx="2" ry="2"></rect>
								<line x1="1" y1="10" x2="23" y2="10"></line>
							</svg>
							<span>Billing</span>
						</a>
						<div class="dropdown-divider"></div>
						<button class="dropdown-item logout" on:click={handleLogout}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
								<polyline points="16,17 21,12 16,7"></polyline>
								<line x1="21" y1="12" x2="9" y2="12"></line>
							</svg>
							<span>Logout</span>
						</button>
					</div>
				</div>
			{:else}
				<div class="auth-buttons">
					<a href="/login" class="btn btn-ghost">Login</a>
					<a href="/register" class="btn btn-primary">Sign Up</a>
				</div>
			{/if}

			<button class="mobile-menu-button" on:click={toggleMenu}>
				<div class="hamburger" class:open={isMenuOpen}>
					<span></span>
					<span></span>
					<span></span>
				</div>
			</button>
		</div>
	</div>
</nav>

<style>
	.navigation {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		z-index: var(--z-fixed);
		transition: all var(--transition-normal);
	}

	.navigation.scrolled {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
		box-shadow: var(--shadow-lg);
	}

	.nav-container {
		max-width: 1200px;
		margin: 0 auto;
		padding: 0 var(--space-lg);
		display: flex;
		align-items: center;
		justify-content: space-between;
		height: 80px;
	}

	.nav-brand {
		flex-shrink: 0;
	}

	.brand-link {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		text-decoration: none;
		color: inherit;
		transition: all var(--transition-normal);
	}

	.brand-link:hover {
		transform: scale(1.05);
	}

	.brand-logo {
		width: 32px;
		height: 32px;
		background: var(--primary-gradient);
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: var(--shadow-md);
	}

	.brand-logo svg {
		width: 20px;
		height: 20px;
		color: var(--white);
	}

	.brand-text {
		font-size: var(--text-xl);
		font-weight: 700;
		color: var(--text-primary);
		font-family: var(--font-display);
	}

	.nav-menu {
		display: flex;
		align-items: center;
		gap: var(--space-2xl);
	}

	.nav-link {
		text-decoration: none;
		color: var(--text-primary);
		font-weight: 500;
		transition: all var(--transition-normal);
		position: relative;
		padding: var(--space-sm) var(--space-md);
		border-radius: var(--radius-lg);
	}

	.nav-link:hover {
		color: var(--primary);
		background: var(--bg-glass);
	}

	.nav-link::after {
		content: '';
		position: absolute;
		bottom: -2px;
		left: 50%;
		width: 0;
		height: 2px;
		background: var(--primary-gradient);
		transition: all var(--transition-normal);
		transform: translateX(-50%);
	}

	.nav-link:hover::after {
		width: 80%;
	}

	.nav-actions {
		display: flex;
		align-items: center;
		gap: var(--space-lg);
	}

	.auth-buttons {
		display: flex;
		gap: var(--space-sm);
	}

	.user-menu {
		position: relative;
	}

	.user-button {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		padding: var(--space-sm) var(--space-md);
		border-radius: var(--radius-xl);
		cursor: pointer;
		transition: all var(--transition-normal);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.user-button:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}

	.user-avatar {
		width: 32px;
		height: 32px;
		background: var(--primary-gradient);
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 600;
		color: var(--white);
		font-size: var(--text-sm);
	}

	.user-name {
		font-weight: 500;
		color: var(--text-primary);
	}

	.chevron {
		width: 16px;
		height: 16px;
		transition: transform var(--transition-normal);
		color: var(--text-secondary);
	}

	.user-menu.open .chevron {
		transform: rotate(180deg);
	}

	.dropdown-menu {
		position: absolute;
		top: 100%;
		right: 0;
		margin-top: var(--space-sm);
		min-width: 220px;
		padding: var(--space-sm);
		border-radius: var(--radius-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
		opacity: 0;
		visibility: hidden;
		transform: translateY(-10px);
		transition: all var(--transition-normal);
	}

	.dropdown-menu.open {
		opacity: 1;
		visibility: visible;
		transform: translateY(0);
	}

	.dropdown-item {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		padding: var(--space-md);
		color: var(--text-primary);
		text-decoration: none;
		transition: all var(--transition-normal);
		border: none;
		background: none;
		width: 100%;
		text-align: left;
		cursor: pointer;
		font-size: var(--text-sm);
		border-radius: var(--radius-lg);
	}

	.dropdown-item:hover {
		background: var(--bg-glass);
		transform: translateX(4px);
	}

	.dropdown-item svg {
		width: 18px;
		height: 18px;
		flex-shrink: 0;
	}

	.dropdown-divider {
		height: 1px;
		background: rgba(255, 255, 255, 0.1);
		margin: var(--space-sm) 0;
	}

	.logout {
		color: var(--error);
	}

	.logout:hover {
		background: var(--error-bg);
	}

	.mobile-menu-button {
		display: none;
		background: none;
		border: none;
		cursor: pointer;
		padding: var(--space-sm);
		border-radius: var(--radius-lg);
		transition: all var(--transition-normal);
	}

	.mobile-menu-button:hover {
		background: var(--bg-glass);
	}

	.hamburger {
		width: 24px;
		height: 20px;
		position: relative;
		display: flex;
		flex-direction: column;
		justify-content: space-between;
	}

	.hamburger span {
		width: 100%;
		height: 2px;
		background: var(--text-primary);
		border-radius: 2px;
		transition: all var(--transition-normal);
		transform-origin: center;
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

	@media (max-width: 768px) {
		.nav-menu {
			position: fixed;
			top: 80px;
			left: 0;
			right: 0;
			background: var(--bg-glass);
			backdrop-filter: blur(20px);
			-webkit-backdrop-filter: blur(20px);
			border-bottom: 1px solid rgba(255, 255, 255, 0.1);
			flex-direction: column;
			padding: var(--space-lg);
			gap: var(--space-md);
			transform: translateY(-100%);
			opacity: 0;
			visibility: hidden;
			transition: all var(--transition-normal);
		}

		.nav-menu.open {
			transform: translateY(0);
			opacity: 1;
			visibility: visible;
		}

		.nav-link {
			width: 100%;
			padding: var(--space-md);
			border-radius: var(--radius-lg);
			text-align: center;
		}

		.nav-link:hover {
			background: var(--bg-glass);
		}

		.mobile-menu-button {
			display: block;
		}

		.auth-buttons {
			display: none;
		}

		.user-menu {
			display: none;
		}

		.nav-container {
			height: 70px;
		}
	}

	@media (max-width: 480px) {
		.nav-container {
			padding: 0 var(--space-md);
		}

		.brand-text {
			font-size: var(--text-lg);
		}

		.brand-logo {
			width: 28px;
			height: 28px;
		}

		.brand-logo svg {
			width: 16px;
			height: 16px;
		}
	}
</style> 