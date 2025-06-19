<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/auth';
	import { goto } from '$app/navigation';
	import Navigation from '$lib/components/Navigation.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import AdDisplay from '$lib/components/AdDisplay.svelte';

	let isAuthenticated = false;
	let scrollY = 0;
	let isLoaded = false;

	// Subscribe to auth store
	auth.subscribe(state => {
		isAuthenticated = state.isAuthenticated;
	});

	onMount(() => {
		// Set loaded state for animations
		setTimeout(() => {
			isLoaded = true;
		}, 100);

		// Parallax scroll effect
		const handleScroll = () => {
			scrollY = window.scrollY;
		};

		window.addEventListener('scroll', handleScroll);
		return () => window.removeEventListener('scroll', handleScroll);
	});

	const handleGetStarted = () => {
		if (isAuthenticated) {
			goto('/videos');
		} else {
			goto('/register');
		}
	};

	const handleLearnMore = () => {
		document.getElementById('features')?.scrollIntoView({ behavior: 'smooth' });
	};
</script>

<svelte:head>
	<title>BOME - Book of Mormon Evidences | Modern Streaming Platform</title>
	<meta name="description" content="Discover compelling evidence for the Book of Mormon through our modern streaming platform. High-quality videos, expert analysis, and exclusive content." />
	<meta name="keywords" content="Book of Mormon, evidence, streaming, LDS, archaeology, history" />
</svelte:head>

<!-- Hero Section with Parallax -->
<section class="hero-section parallax-container">
	<div class="parallax-bg" style="transform: translateY({scrollY * 0.5}px)"></div>
	
	<!-- Floating Elements -->
	<div class="floating-elements">
		<div class="floating-shape shape-1" style="transform: translateY({scrollY * 0.3}px) rotate({scrollY * 0.02}deg)"></div>
		<div class="floating-shape shape-2" style="transform: translateY({scrollY * 0.2}px) rotate({scrollY * -0.01}deg)"></div>
		<div class="floating-shape shape-3" style="transform: translateY({scrollY * 0.4}px) rotate({scrollY * 0.015}deg)"></div>
	</div>

	<Navigation />
	
	<!-- Header Ad Placement -->
	<AdDisplay 
		placementId={1} 
		className="header"
		fallbackContent="<div style='text-align: center; padding: 1rem; color: #666;'>Advertisement Space Available</div>"
	/>
	
	<div class="hero-content container">
		<div class="hero-text fade-in" class:loaded={isLoaded}>
			<h1 class="hero-title">
				<span class="gradient-text">Discover</span> the Evidence
			</h1>
			<p class="hero-subtitle">
				Explore compelling archaeological and historical evidence for the Book of Mormon through our modern streaming platform. 
				High-quality documentaries, expert analysis, and exclusive content await.
			</p>
			<div class="hero-actions">
				<button class="btn btn-primary btn-large" on:click={handleGetStarted}>
					<span>Get Started</span>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M5 12h14"></path>
						<path d="m12 5 7 7-7 7"></path>
					</svg>
				</button>
				<button class="btn btn-ghost btn-large" on:click={handleLearnMore}>
					Learn More
				</button>
			</div>
		</div>
		
		<div class="hero-visual slide-up" class:loaded={isLoaded}>
			<div class="hero-card glass">
				<div class="card-content">
					<div class="feature-badge">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M9 12l2 2 4-4"></path>
							<path d="M21 12c0 4.97-4.03 9-9 9s-9-4.03-9-9 4.03-9 9-9 9 4.03 9 9z"></path>
						</svg>
						<span>4K Quality</span>
					</div>
					<h3>Latest Documentary</h3>
					<p>Ancient Civilizations of the Americas</p>
					<div class="play-button">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<polygon points="5,3 19,12 5,21"></polygon>
						</svg>
					</div>
				</div>
			</div>
		</div>
	</div>

	<!-- Scroll Indicator -->
	<div class="scroll-indicator">
		<div class="scroll-arrow"></div>
	</div>
</section>

<!-- Features Section -->
<section class="features-section" id="features">
	<div class="container">
		<div class="features-content">
			<div class="features-main">
				<div class="section-header text-center fade-in" class:loaded={isLoaded}>
					<h2>Why Choose BOME?</h2>
					<p>Discover what makes our platform the premier destination for Book of Mormon evidence</p>
				</div>

				<div class="features-grid fade-in" class:loaded={isLoaded}>
					<!-- Feature cards content here -->
					<div class="feature-card glass">
						<div class="feature-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"></path>
							</svg>
						</div>
						<h3>Expert Analysis</h3>
						<p>In-depth analysis from leading scholars and researchers in Book of Mormon studies.</p>
					</div>

					<div class="feature-card glass">
						<div class="feature-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<circle cx="12" cy="12" r="10"></circle>
								<polygon points="10,8 16,12 10,16 10,8"></polygon>
							</svg>
						</div>
						<h3>High-Quality Videos</h3>
						<p>Professional documentaries and presentations in stunning HD quality.</p>
					</div>

					<div class="feature-card glass">
						<div class="feature-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"></path>
								<path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"></path>
							</svg>
						</div>
						<h3>Scholarly Research</h3>
						<p>Access to peer-reviewed research and archaeological findings.</p>
					</div>

					<div class="feature-card glass">
						<div class="feature-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
								<circle cx="9" cy="7" r="4"></circle>
								<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
								<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
							</svg>
						</div>
						<h3>Community</h3>
						<p>Connect with fellow seekers and scholars in our vibrant community.</p>
					</div>
				</div>

				<!-- Content Ad Placement -->
				<div class="content-ad-section">
					<AdDisplay 
						placementId={4} 
						className="content"
						fallbackContent="<div style='text-align: center; padding: 2rem; background: rgba(255,255,255,0.05); border-radius: 12px; color: #666;'>Sponsored Content Space</div>"
					/>
				</div>
			</div>

			<!-- Sidebar with Ad -->
			<div class="features-sidebar">
				<AdDisplay 
					placementId={2} 
					className="sidebar"
					fallbackContent="<div style='text-align: center; padding: 2rem; background: rgba(255,255,255,0.05); border-radius: 12px; color: #666;'>Sidebar Advertisement</div>"
				/>
				
				<!-- Additional sidebar content -->
				<div class="sidebar-content">
					<h3>Latest Updates</h3>
					<div class="update-item">
						<h4>New Archaeological Discovery</h4>
						<p>Recent findings in Central America provide new insights...</p>
						<span class="date">2 days ago</span>
					</div>
					<div class="update-item">
						<h4>Expert Interview Series</h4>
						<p>Join us for exclusive interviews with leading scholars...</p>
						<span class="date">1 week ago</span>
					</div>
				</div>
			</div>
		</div>
	</div>
</section>

<!-- Stats Section -->
<section class="stats-section">
	<div class="container">
		<div class="stats-grid grid grid-4">
			<div class="stat-item fade-in" class:loaded={isLoaded}>
				<div class="stat-number gradient-text">500+</div>
				<div class="stat-label">Videos</div>
			</div>
			<div class="stat-item fade-in" class:loaded={isLoaded} style="animation-delay: 0.1s">
				<div class="stat-number gradient-text">50+</div>
				<div class="stat-label">Experts</div>
			</div>
			<div class="stat-item fade-in" class:loaded={isLoaded} style="animation-delay: 0.2s">
				<div class="stat-number gradient-text">10K+</div>
				<div class="stat-label">Members</div>
			</div>
			<div class="stat-item fade-in" class:loaded={isLoaded} style="animation-delay: 0.3s">
				<div class="stat-number gradient-text">24/7</div>
				<div class="stat-label">Access</div>
			</div>
		</div>
	</div>
</section>

<!-- CTA Section -->
<section class="cta-section">
	<div class="container">
		<div class="cta-content text-center fade-in" class:loaded={isLoaded}>
			<h2>Ready to Discover the Evidence?</h2>
			<p>Join thousands of members exploring the fascinating world of Book of Mormon archaeology and history.</p>
			<div class="cta-actions">
				<button class="btn btn-primary btn-large" on:click={handleGetStarted}>
					Start Your Journey
				</button>
				<a href="/subscription" class="btn btn-outline btn-large">View Plans</a>
			</div>
		</div>
	</div>
</section>

<style>
	/* Hero Section */
	.hero-section {
		min-height: 100vh;
		position: relative;
		display: flex;
		align-items: center;
		overflow: hidden;
	}

	.parallax-bg {
		background: var(--primary-gradient);
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 120%;
		z-index: -2;
	}

	.floating-elements {
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		z-index: -1;
		overflow: hidden;
	}

	.floating-shape {
		position: absolute;
		border-radius: 50%;
		background: rgba(255, 255, 255, 0.1);
		backdrop-filter: blur(10px);
	}

	.shape-1 {
		width: 200px;
		height: 200px;
		top: 20%;
		left: 10%;
		animation: float 6s ease-in-out infinite;
	}

	.shape-2 {
		width: 150px;
		height: 150px;
		top: 60%;
		right: 15%;
		animation: float 8s ease-in-out infinite reverse;
	}

	.shape-3 {
		width: 100px;
		height: 100px;
		top: 80%;
		left: 20%;
		animation: float 7s ease-in-out infinite;
	}

	@keyframes float {
		0%, 100% { transform: translateY(0px); }
		50% { transform: translateY(-20px); }
	}

	.hero-content {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-4xl);
		align-items: center;
		position: relative;
		z-index: 1;
	}

	.hero-text {
		opacity: 0;
		transform: translateY(30px);
		transition: all 0.8s ease-out;
	}

	.hero-text.loaded {
		opacity: 1;
		transform: translateY(0);
	}

	.hero-title {
		font-size: var(--text-7xl);
		font-weight: 800;
		line-height: 1.1;
		margin-bottom: var(--space-lg);
		color: var(--white);
	}

	.gradient-text {
		background: linear-gradient(135deg, #fff 0%, #f0f0f0 100%);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}

	.hero-subtitle {
		font-size: var(--text-xl);
		color: rgba(255, 255, 255, 0.9);
		margin-bottom: var(--space-2xl);
		line-height: 1.6;
	}

	.hero-actions {
		display: flex;
		gap: var(--space-lg);
		flex-wrap: wrap;
	}

	.hero-visual {
		opacity: 0;
		transform: translateY(30px);
		transition: all 0.8s ease-out 0.2s;
	}

	.hero-visual.loaded {
		opacity: 1;
		transform: translateY(0);
	}

	.hero-card {
		padding: var(--space-2xl);
		text-align: center;
		position: relative;
	}

	.card-content {
		position: relative;
	}

	.feature-badge {
		display: inline-flex;
		align-items: center;
		gap: var(--space-sm);
		background: rgba(255, 255, 255, 0.1);
		padding: var(--space-sm) var(--space-md);
		border-radius: var(--radius-full);
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--white);
		margin-bottom: var(--space-lg);
	}

	.feature-badge svg {
		width: 16px;
		height: 16px;
	}

	.hero-card h3 {
		font-size: var(--text-2xl);
		color: var(--white);
		margin-bottom: var(--space-sm);
	}

	.hero-card p {
		color: rgba(255, 255, 255, 0.8);
		margin-bottom: var(--space-xl);
	}

	.play-button {
		width: 60px;
		height: 60px;
		background: var(--white);
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 0 auto;
		cursor: pointer;
		transition: all var(--transition-normal);
		box-shadow: var(--shadow-lg);
	}

	.play-button:hover {
		transform: scale(1.1);
		box-shadow: var(--shadow-xl);
	}

	.play-button svg {
		width: 24px;
		height: 24px;
		color: var(--primary);
		margin-left: 2px;
	}

	.scroll-indicator {
		position: absolute;
		bottom: var(--space-2xl);
		left: 50%;
		transform: translateX(-50%);
		z-index: 1;
	}

	.scroll-arrow {
		width: 2px;
		height: 30px;
		background: rgba(255, 255, 255, 0.5);
		position: relative;
		animation: scroll 2s ease-in-out infinite;
	}

	.scroll-arrow::after {
		content: '';
		position: absolute;
		bottom: 0;
		left: -3px;
		width: 8px;
		height: 8px;
		border-right: 2px solid rgba(255, 255, 255, 0.5);
		border-bottom: 2px solid rgba(255, 255, 255, 0.5);
		transform: rotate(45deg);
	}

	@keyframes scroll {
		0%, 100% { transform: translateY(0); }
		50% { transform: translateY(10px); }
	}

	/* Features Section */
	.features-section {
		padding: var(--space-4xl) 0;
		background: var(--bg-secondary);
		position: relative;
	}

	.features-content {
		display: grid;
		grid-template-columns: 1fr 300px;
		gap: var(--space-2xl);
		align-items: start;
	}

	.features-main {
		display: flex;
		flex-direction: column;
		gap: var(--space-2xl);
	}

	.features-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: var(--space-xl);
	}

	.content-ad-section {
		margin: var(--space-2xl) 0;
	}

	.features-sidebar {
		display: flex;
		flex-direction: column;
		gap: var(--space-xl);
		position: sticky;
		top: 100px;
	}

	.sidebar-content {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.sidebar-content h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-lg) 0;
	}

	.update-item {
		padding: var(--space-md) 0;
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
	}

	.update-item:last-child {
		border-bottom: none;
	}

	.update-item h4 {
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-xs) 0;
	}

	.update-item p {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		margin: 0 0 var(--space-xs) 0;
		line-height: 1.4;
	}

	.update-item .date {
		font-size: var(--text-xs);
		color: var(--text-secondary);
		opacity: 0.7;
	}

	/* Stats Section */
	.stats-section {
		padding: var(--space-4xl) 0;
		background: var(--bg-primary);
	}

	.stats-grid {
		text-align: center;
	}

	.stat-item {
		opacity: 0;
		transform: translateY(20px);
		transition: all 0.6s ease-out;
	}

	.stat-item.loaded {
		opacity: 1;
		transform: translateY(0);
	}

	.stat-number {
		font-size: var(--text-6xl);
		font-weight: 800;
		line-height: 1;
		margin-bottom: var(--space-sm);
	}

	.stat-label {
		font-size: var(--text-lg);
		color: var(--text-secondary);
		font-weight: 500;
	}

	/* CTA Section */
	.cta-section {
		padding: var(--space-4xl) 0;
		background: var(--dark-gradient);
		color: var(--white);
	}

	.cta-content {
		opacity: 0;
		transform: translateY(30px);
		transition: all 0.8s ease-out;
	}

	.cta-content.loaded {
		opacity: 1;
		transform: translateY(0);
	}

	.cta-content h2 {
		font-size: var(--text-5xl);
		margin-bottom: var(--space-lg);
	}

	.cta-content p {
		font-size: var(--text-xl);
		color: rgba(255, 255, 255, 0.9);
		margin-bottom: var(--space-2xl);
		max-width: 600px;
		margin-left: auto;
		margin-right: auto;
	}

	.cta-actions {
		display: flex;
		gap: var(--space-lg);
		justify-content: center;
		flex-wrap: wrap;
	}

	/* Responsive Design */
	@media (max-width: 1024px) {
		.hero-content {
			grid-template-columns: 1fr;
			gap: var(--space-3xl);
			text-align: center;
		}

		.hero-title {
			font-size: var(--text-6xl);
		}

		.features-grid {
			grid-template-columns: repeat(2, 1fr);
		}

		.stats-grid {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	@media (max-width: 768px) {
		.hero-title {
			font-size: var(--text-5xl);
		}

		.hero-subtitle {
			font-size: var(--text-lg);
		}

		.hero-actions {
			flex-direction: column;
			align-items: center;
		}

		.features-grid {
			grid-template-columns: 1fr;
		}

		.stats-grid {
			grid-template-columns: 1fr;
		}

		.cta-actions {
			flex-direction: column;
			align-items: center;
		}

		.section-title {
			font-size: var(--text-4xl);
		}

		.stat-number {
			font-size: var(--text-5xl);
		}
	}

	@media (max-width: 480px) {
		.hero-title {
			font-size: var(--text-4xl);
		}

		.hero-card {
			padding: var(--space-xl);
		}

		.feature-card {
			padding: var(--space-xl);
		}
	}
</style>
