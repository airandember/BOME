<script lang="ts">
	import { auth } from '$lib/auth';
	
	let isAuthenticated = false;
	let user: any = null;
	
	// Subscribe to auth store
	auth.subscribe(state => {
		isAuthenticated = state.isAuthenticated;
		user = state.user;
	});
</script>

<footer class="footer">
	<div class="footer-container">
		<!-- Main Footer Content -->
		<div class="footer-main">
			<!-- Brand Section -->
			<div class="footer-brand">
				<div class="brand-logo">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M12 2L2 7l10 5 10-5-10-5z"></path>
						<path d="M2 17l10 5 10-5"></path>
						<path d="M2 12l10 5 10-5"></path>
					</svg>
				</div>
				<h3>BOME</h3>
				<p>Discover compelling evidence for the Book of Mormon through our modern streaming platform.</p>
				<div class="social-links">
					<a href="#" aria-label="Facebook">
						<svg viewBox="0 0 24 24" fill="currentColor">
							<path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z"/>
						</svg>
					</a>
					<a href="#" aria-label="Twitter">
						<svg viewBox="0 0 24 24" fill="currentColor">
							<path d="M23.953 4.57a10 10 0 01-2.825.775 4.958 4.958 0 002.163-2.723c-.951.555-2.005.959-3.127 1.184a4.92 4.92 0 00-8.384 4.482C7.69 8.095 4.067 6.13 1.64 3.162a4.822 4.822 0 00-.666 2.475c0 1.71.87 3.213 2.188 4.096a4.904 4.904 0 01-2.228-.616v.06a4.923 4.923 0 003.946 4.827 4.996 4.996 0 01-2.212.085 4.936 4.936 0 004.604 3.417 9.867 9.867 0 01-6.102 2.105c-.39 0-.779-.023-1.17-.067a13.995 13.995 0 007.557 2.209c9.053 0 13.998-7.496 13.998-13.985 0-.21 0-.42-.015-.63A9.935 9.935 0 0024 4.59z"/>
						</svg>
					</a>
					<a href="#" aria-label="YouTube">
						<svg viewBox="0 0 24 24" fill="currentColor">
							<path d="M23.498 6.186a3.016 3.016 0 0 0-2.122-2.136C19.505 3.545 12 3.545 12 3.545s-7.505 0-9.377.505A3.017 3.017 0 0 0 .502 6.186C0 8.07 0 12 0 12s0 3.93.502 5.814a3.016 3.016 0 0 0 2.122 2.136c1.871.505 9.376.505 9.376.505s7.505 0 9.377-.505a3.015 3.015 0 0 0 2.122-2.136C24 15.93 24 12 24 12s0-3.93-.502-5.814zM9.545 15.568V8.432L15.818 12l-6.273 3.568z"/>
						</svg>
					</a>
				</div>
			</div>

			<!-- Quick Links -->
			<div class="footer-section">
				<h4>Explore</h4>
				<ul>
					<li><a href="/videos">Videos</a></li>
					<li><a href="/articles">Articles</a></li>
					<li><a href="/events">Events</a></li>
					<li><a href="/about">About</a></li>
					<li><a href="/contact">Contact</a></li>
				</ul>
			</div>

			<!-- Account Links -->
			<div class="footer-section">
				<h4>Account</h4>
				<ul>
					{#if isAuthenticated}
						<li><a href="/dashboard">Dashboard</a></li>
										<li><a href="/dashboard?tab=profile">My Account</a></li>
				<li><a href="/subscription">Subscription</a></li>
				<li><a href="/account/billing">Billing</a></li>
					{:else}
						<li><a href="/login">Login</a></li>
						<li><a href="/register">Sign Up</a></li>
						<li><a href="/subscription">Plans</a></li>
					{/if}
				</ul>
			</div>

			<!-- Support Links -->
			<div class="footer-section">
				<h4>Support</h4>
				<ul>
					<li><a href="/help">Help Center</a></li>
					<li><a href="/faq">FAQ</a></li>
					<li><a href="/advertise" class="advertise-link">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z"></path>
						</svg>
						Advertise with us!
					</a></li>
					<li><a href="/privacy">Privacy Policy</a></li>
					<li><a href="/terms">Terms of Service</a></li>
				</ul>
			</div>
		</div>

		<!-- Footer Bottom -->
		<div class="footer-bottom">
			<div class="footer-bottom-content">
				<p>&copy; {new Date().getFullYear()} BOME - Book of Mormon Evidences. All rights reserved.</p>
				<div class="footer-links">
					<a href="/privacy">Privacy</a>
					<a href="/terms">Terms</a>
					<a href="/cookies">Cookies</a>
				</div>
			</div>
		</div>
	</div>
</footer>

<style>
	.footer {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		border-top: 1px solid rgba(255, 255, 255, 0.1);
		margin-top: var(--space-4xl);
	}

	.footer-container {
		max-width: 1200px;
		margin: 0 auto;
		padding: 0 var(--space-lg);
	}

	.footer-main {
		display: grid;
		grid-template-columns: 2fr 1fr 1fr 1fr;
		gap: var(--space-2xl);
		padding: var(--space-4xl) 0 var(--space-2xl);
	}

	.footer-brand {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.brand-logo {
		width: 48px;
		height: 48px;
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

	.footer-brand h3 {
		font-size: var(--text-2xl);
		font-weight: 700;
		color: var(--text-primary);
		margin: 0;
		font-family: var(--font-display);
	}

	.footer-brand p {
		color: var(--text-secondary);
		line-height: 1.6;
		margin: 0;
	}

	.social-links {
		display: flex;
		gap: var(--space-md);
	}

	.social-links a {
		width: 40px;
		height: 40px;
		border-radius: var(--radius-lg);
		background: var(--bg-glass);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--text-secondary);
		transition: all var(--transition-normal);
	}

	.social-links a:hover {
		background: var(--primary);
		color: var(--white);
		transform: translateY(-2px);
		box-shadow: var(--shadow-md);
	}

	.social-links svg {
		width: 20px;
		height: 20px;
	}

	.footer-section h4 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-lg) 0;
	}

	.footer-section ul {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.footer-section a {
		color: var(--text-secondary);
		text-decoration: none;
		transition: color var(--transition-normal);
		font-size: var(--text-sm);
	}

	.footer-section a:hover {
		color: var(--primary);
	}

	/* Special styling for advertise link */
	.advertise-link {
		display: flex;
		align-items: center;
		gap: var(--space-xs);
		font-weight: 600;
		color: var(--primary) !important;
		transition: all var(--transition-normal);
	}

	.advertise-link svg {
		width: 16px;
		height: 16px;
		transition: transform var(--transition-normal);
	}

	.advertise-link:hover {
		color: var(--primary-dark) !important;
		transform: translateX(2px);
	}

	.advertise-link:hover svg {
		transform: scale(1.1);
	}

	.footer-bottom {
		border-top: 1px solid rgba(255, 255, 255, 0.1);
		padding: var(--space-xl) 0;
	}

	.footer-bottom-content {
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: var(--space-md);
	}

	.footer-bottom p {
		color: var(--text-secondary);
		font-size: var(--text-sm);
		margin: 0;
	}

	.footer-links {
		display: flex;
		gap: var(--space-lg);
	}

	.footer-links a {
		color: var(--text-secondary);
		text-decoration: none;
		font-size: var(--text-sm);
		transition: color var(--transition-normal);
	}

	.footer-links a:hover {
		color: var(--primary);
	}

	/* Responsive Design */
	@media (max-width: 1024px) {
		.footer-main {
			grid-template-columns: 1fr 1fr;
			gap: var(--space-xl);
		}
	}

	@media (max-width: 768px) {
		.footer-main {
			grid-template-columns: 1fr;
			gap: var(--space-xl);
			text-align: center;
		}

		.footer-brand {
			align-items: center;
		}

		.footer-bottom-content {
			flex-direction: column;
			text-align: center;
		}

		.footer-links {
			justify-content: center;
		}
	}

	@media (max-width: 480px) {
		.footer-container {
			padding: 0 var(--space-md);
		}

		.footer-main {
			padding: var(--space-2xl) 0 var(--space-xl);
		}

		.social-links {
			justify-content: center;
		}
	}
</style> 