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
		<!-- Footer Bottom -->
		<div class="footer-bottom">
			<div class="footer-bottom-content">
				<p>&copy; {new Date().getFullYear()} BOME - Book of Mormon Evidences. All rights reserved.</p>
				<div class="footer-links">
					<a href="https://bookofmormonevidence.org/privacy-policy/">Privacy</a>
					<a href="https://bookofmormonevidence.org/terms-and-conditions/">Terms</a>
					<!--<a href="/cookies">Cookies</a>-->
					
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
	}

	.footer-container {
		max-width: 100vw;
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

	/* Admin link styling */
	.admin-link {
		display: flex;
		align-items: center;
		gap: 4px;
		opacity: 0.6;
		transition: opacity var(--transition-normal);
	}

	.admin-link:hover {
		opacity: 1;
		color: var(--primary);
	}

	.admin-link svg {
		width: 14px;
		height: 14px;
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
