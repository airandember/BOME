<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/auth';
	import { goto } from '$app/navigation';
	import Navigation from '$lib/components/Navigation.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

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
<section id="features" class="features-section">
	<div class="container">
		<div class="section-header text-center fade-in" class:loaded={isLoaded}>
			<h2 class="section-title">
				Why Choose <span class="gradient-text">BOME</span>
			</h2>
			<p class="section-subtitle">
				Experience the future of educational streaming with cutting-edge features designed for discovery and learning.
			</p>
		</div>

		<div class="features-grid grid grid-3">
			<div class="feature-card card glass slide-up" class:loaded={isLoaded}>
				<div class="feature-icon">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M14.5 4h-5L7 7H4a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2h-3l-2.5-3z"></path>
						<circle cx="12" cy="13" r="3"></circle>
					</svg>
				</div>
				<h3>4K Ultra HD</h3>
				<p>Crystal clear video quality that brings ancient artifacts and archaeological sites to life with stunning detail.</p>
			</div>

			<div class="feature-card card glass slide-up" class:loaded={isLoaded} style="animation-delay: 0.1s">
				<div class="feature-icon">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"></path>
					</svg>
				</div>
				<h3>Expert Analysis</h3>
				<p>Content curated and analyzed by leading archaeologists, historians, and Book of Mormon scholars.</p>
			</div>

			<div class="feature-card card glass slide-up" class:loaded={isLoaded} style="animation-delay: 0.2s">
				<div class="feature-icon">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"></path>
					</svg>
				</div>
				<h3>Exclusive Content</h3>
				<p>Access to rare footage, unpublished research, and behind-the-scenes content not available anywhere else.</p>
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
	}

	.section-header {
		margin-bottom: var(--space-4xl);
	}

	.section-title {
		font-size: var(--text-5xl);
		margin-bottom: var(--space-lg);
	}

	.section-subtitle {
		font-size: var(--text-xl);
		color: var(--text-secondary);
		max-width: 600px;
		margin: 0 auto;
	}

	.feature-card {
		padding: var(--space-2xl);
		text-align: center;
		transition: all var(--transition-normal);
	}

	.feature-card:hover {
		transform: translateY(-8px);
	}

	.feature-icon {
		width: 80px;
		height: 80px;
		background: var(--primary-gradient);
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 0 auto var(--space-xl);
	}

	.feature-icon svg {
		width: 32px;
		height: 32px;
		color: var(--white);
	}

	.feature-card h3 {
		font-size: var(--text-2xl);
		margin-bottom: var(--space-md);
	}

	.feature-card p {
		color: var(--text-secondary);
		line-height: 1.6;
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
