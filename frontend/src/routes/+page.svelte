<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/auth';
	import { goto } from '$app/navigation';
	import Navigation from '$lib/components/Navigation.svelte';
	import Footer from '$lib/components/Footer.svelte';

	let isAuthenticated = false;
	let currentSection = 0;
	let isScrolling = false;
	let scrollContainer: HTMLElement;
	let sections: HTMLElement[] = [];

	// Subscribe to auth store
	auth.subscribe(state => {
		isAuthenticated = state.isAuthenticated;
	});

	onMount(() => {
		// Initialize scroll snapping and animation triggers
		setupScrollAnimations();
		
		// Add wheel event listener for smooth scroll snapping
		const handleWheel = (e: WheelEvent) => {
			e.preventDefault();
			if (isScrolling) return;
			
			const direction = e.deltaY > 0 ? 1 : -1;
			const nextSection = currentSection + direction;
			
			if (nextSection >= 0 && nextSection < sections.length) {
				scrollToSection(nextSection);
			}
		};

		// Add touch events for mobile
		let startY = 0;
		const handleTouchStart = (e: TouchEvent) => {
			startY = e.touches[0].clientY;
		};

		const handleTouchEnd = (e: TouchEvent) => {
			if (isScrolling) return;
			const endY = e.changedTouches[0].clientY;
			const diff = startY - endY;
			
			if (Math.abs(diff) > 50) { // Minimum swipe distance
				const direction = diff > 0 ? 1 : -1;
				const nextSection = currentSection + direction;
				
				if (nextSection >= 0 && nextSection < sections.length) {
					scrollToSection(nextSection);
				}
			}
		};

		// Add keyboard navigation
		const handleKeyDown = (e: KeyboardEvent) => {
			if (isScrolling) return;
			
			let direction = 0;
			if (e.key === 'ArrowDown' || e.key === 'PageDown') direction = 1;
			if (e.key === 'ArrowUp' || e.key === 'PageUp') direction = -1;
			
			if (direction !== 0) {
				e.preventDefault();
				const nextSection = currentSection + direction;
				if (nextSection >= 0 && nextSection < sections.length) {
					scrollToSection(nextSection);
				}
			}
		};

		window.addEventListener('wheel', handleWheel, { passive: false });
		window.addEventListener('touchstart', handleTouchStart);
		window.addEventListener('touchend', handleTouchEnd);
		window.addEventListener('keydown', handleKeyDown);

		return () => {
			window.removeEventListener('wheel', handleWheel);
			window.removeEventListener('touchstart', handleTouchStart);
			window.removeEventListener('touchend', handleTouchEnd);
			window.removeEventListener('keydown', handleKeyDown);
		};
	});

	function setupScrollAnimations() {
		sections = Array.from(document.querySelectorAll('.zoom-section'));
		// Start with first section
		updateSectionStates(0);
	}

	function scrollToSection(index: number) {
		if (isScrolling || index === currentSection) return;
		
		isScrolling = true;
		currentSection = index;
		
		// Smooth scroll to section
		sections[index].scrollIntoView({ 
			behavior: 'smooth',
			block: 'start'
		});
		
		// Update animation states
		updateSectionStates(index);
		
		// Reset scrolling flag after animation
		setTimeout(() => {
			isScrolling = false;
		}, 1000);
	}

	function updateSectionStates(activeIndex: number) {
		sections.forEach((section, index) => {
			const isActive = index === activeIndex;
			const isPast = index < activeIndex;
			const isFuture = index > activeIndex;
			
			section.classList.toggle('active', isActive);
			section.classList.toggle('past', isPast);
			section.classList.toggle('future', isFuture);
		});
	}

	const handleGetStarted = () => {
		if (isAuthenticated) {
			goto('/videos');
		} else {
			goto('/register');
		}
	};

	// Navigation dots
	function goToSection(index: number) {
		if (!isScrolling) {
			scrollToSection(index);
		}
	}
</script>

<svelte:head>
	<title>BOME - Book of Mormon Evidence | Modern Streaming Platform</title>
	<meta name="description" content="Discover compelling evidence for the Book of Mormon through our modern streaming platform. High-quality videos, expert analysis, and exclusive content." />
	<meta name="keywords" content="Book of Mormon, evidence, streaming, LDS, archaeology, history" />
</svelte:head>

<Navigation />

<!-- Fixed Header for Slide 1 Only -->
{#if currentSection === 0}
	<div class="home_page_fixed_header">
		<!-- Content for the fixed header can be added here -->
	</div>
{/if}

<div class="scroll-container" bind:this={scrollContainer}>
	<!-- Section 1: Book Close-up -->
	<section class="zoom-section book-section active" data-section="0">
		<div class="zoom-content">
			<div class="book-container">
				<img src="/src/lib/HOMEPAGE_TEST_ASSETS/book_of_mormon_close_up.png" alt="Book of Mormon Close-up" class="book-image" />
				<div class="book-glow"></div>
			</div>
			
			<!-- Main Title Outside Glass Container -->
			<div class="main-title-container">
				<h1 class="hero-title"><span class="hero-title-book">BOOK OF MORMON</span><br> <span class="hero-title-evidence">EVIDENCE</span></h1>
				<blockquote class="hero-quote">
					"The Book of Mormon was the most correct of any book on earth, and the keystone of our religion, and a man would get nearer to God by abiding by its precepts, than by any other book."
					<cite>– Joseph Smith</cite>
				</blockquote>
			</div>

			<!-- Three Navigation Cards -->
			<div class="navigation-cards">
				<a href="/articles" class="nav-card">
					<div class="card-action">READ</div>
					<div class="nav-card-content">
						<div class="card-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
								<polyline points="14,2 14,8 20,8"></polyline>
								<line x1="16" y1="13" x2="8" y2="13"></line>
								<line x1="16" y1="17" x2="8" y2="17"></line>
								<polyline points="10,9 9,9 8,9"></polyline>
							</svg>
						</div>
						<h3>Articles & Research</h3>
						<p>Explore scholarly articles and archaeological evidence</p>
					</div>
				</a>

				<a href="/videos" class="nav-card" on:click|preventDefault={handleGetStarted}>
					<div class="card-action">WATCH</div>
					<div class="nav-card-content">
						<div class="card-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polygon points="23 7 16 12 23 17 23 7"></polygon>
								<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
							</svg>
						</div>
						<h3>Streaming Videos</h3>
						<p>Watch exclusive documentaries and presentations</p>
					</div>
				</a>

				<a href="/events" class="nav-card">
					<div class="card-action">ATTEND</div>
					<div class="nav-card-content">
						<div class="card-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
								<line x1="16" y1="2" x2="16" y2="6"></line>
								<line x1="8" y1="2" x2="8" y2="6"></line>
								<line x1="3" y1="10" x2="21" y2="10"></line>
							</svg>
						</div>
						<h3>Events & Expo</h3>
						<p>Join conferences, seminars, and exhibitions</p>
					</div>
				</a>
			</div>
		</div>
	</section>

	<!-- Section 2: New York State -->
	<section class="zoom-section newyork-section" data-section="1">
		<div class="zoom-content">
			<div class="map-container">
				<img src="/src/lib/HOMEPAGE_TEST_ASSETS/new_york_state_map_1830.jpg" alt="New York State Map 1830" class="map-image" />
				<div class="location-marker hill-cumorah">
					<div class="marker-pulse"></div>
					<div class="marker-dot"></div>
				</div>
			</div>
			<div class="content-overlay">
				<h2 class="section-title">Where It All Began</h2>
				<p class="section-description">Hill Cumorah, New York - The sacred hill where Joseph Smith received the golden plates</p>
				<div class="info-card">
					<h3>Historical Significance</h3>
					<p>The restoration began in upstate New York, where ancient records were preserved for centuries.</p>
				</div>
			</div>
		</div>
	</section>

	<!-- Section 3: United States 1830 -->
	<section class="zoom-section usa1830-section" data-section="2">
		<div class="zoom-content">
			<div class="map-container">
				<img src="/src/lib/HOMEPAGE_TEST_ASSETS/united_states_map_1830.png" alt="United States Map 1830" class="map-image vintage" />
				<div class="location-marker palmyra">
					<div class="marker-pulse"></div>
					<div class="marker-dot"></div>
				</div>
				<div class="location-marker kirtland">
					<div class="marker-pulse"></div>
					<div class="marker-dot"></div>
				</div>
			</div>
			<div class="content-overlay">
				<h2 class="section-title">America in 1830</h2>
				<p class="section-description">The United States during the time of the Book of Mormon's publication</p>
				<div class="timeline-info">
					<div class="timeline-item">
						<span class="year">1830</span>
						<span class="event">Book of Mormon Published</span>
					</div>
					<div class="timeline-item">
						<span class="year">1830</span>
						<span class="event">Church Organized</span>
					</div>
				</div>
			</div>
		</div>
	</section>

	<!-- Section 4: Modern United States -->
	<section class="zoom-section usa-modern-section" data-section="3">
		<div class="zoom-content">
			<div class="map-container">
				<img src="/src/lib/HOMEPAGE_TEST_ASSETS/united_states_map_modern.jpg" alt="Modern United States Map" class="map-image modern" />
				<div class="location-marker utah">
					<div class="marker-pulse"></div>
					<div class="marker-dot"></div>
				</div>
				<div class="location-marker missouri">
					<div class="marker-pulse"></div>
					<div class="marker-dot"></div>
				</div>
			</div>
			<div class="content-overlay">
				<h2 class="section-title">Modern Discoveries</h2>
				<p class="section-description">Archaeological evidence continues to emerge across the Americas</p>
				<div class="stats-grid">
					<div class="stat-item">
						<span class="stat-number">500+</span>
						<span class="stat-label">Archaeological Sites</span>
					</div>
					<div class="stat-item">
						<span class="stat-number">50+</span>
						<span class="stat-label">Expert Researchers</span>
					</div>
				</div>
			</div>
		</div>
	</section>

	<!-- Section 5: World Map -->
	<section class="zoom-section world-section" data-section="4">
		<div class="zoom-content">
			<div class="map-container">
				<img src="/src/lib/HOMEPAGE_TEST_ASSETS/world-physical-maps-international-flat.jpg" alt="World Physical Map" class="map-image world" />
				<div class="location-marker americas">
					<div class="marker-pulse"></div>
					<div class="marker-dot"></div>
				</div>
				<div class="location-marker middle-east">
					<div class="marker-pulse"></div>
					<div class="marker-dot"></div>
				</div>
			</div>
			<div class="content-overlay">
				<h2 class="section-title">Global Connections</h2>
				<p class="section-description">Evidence spans continents, connecting ancient civilizations</p>
				<div class="connection-lines">
					<div class="connection-line"></div>
					<div class="connection-line"></div>
				</div>
			</div>
		</div>
	</section>

	<!-- Section 6: Globe -->
	<section class="zoom-section globe-section" data-section="5">
		<div class="zoom-content">
			<div class="globe-container">
				<div class="globe">
					<img src="/src/lib/HOMEPAGE_TEST_ASSETS/The_Globe_World.png" alt="Earth Globe" class="globe-image" />
					<div class="globe-atmosphere"></div>
					<div class="globe-rotation"></div>
				</div>
			</div>
			<div class="content-overlay">
				<h2 class="section-title">A Global Message</h2>
				<p class="section-description">The Book of Mormon's message reaches every nation, kindred, tongue, and people</p>
				<div class="final-cta">
					<button class="explore-button" on:click={handleGetStarted}>
						Explore the Evidence
					</button>
					<a href="/subscription" class="join-button">Join Our Community</a>
				</div>
			</div>
		</div>
	</section>
</div>

<!-- Navigation Dots -->
<div class="scroll-navigation">
	{#each Array(6) as _, index}
		<button 
			class="nav-dot" 
			class:active={currentSection === index}
			on:click={() => goToSection(index)}
			aria-label="Go to section {index + 1}"
		>
			<span class="dot-inner"></span>
		</button>
	{/each}
</div>

<!-- Scroll Indicator -->
<div class="scroll-indicator" class:hidden={currentSection === 5}>
	<div class="scroll-text">Scroll to explore</div>
	<div class="scroll-arrow">
		<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
			<path d="M7 13l3 3 7-3"></path>
			<path d="M7 6l3 3 7-3"></path>
		</svg>
	</div>
</div>

<Footer />

<style>
	:global(html) {
		scroll-behavior: smooth;
	}

	.scroll-container {
		height: 100vh;
		overflow: hidden;
		position: relative;
	}

	.zoom-section {
		height: 100vh;
		width: 100%;
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
		transition: all 1s cubic-bezier(0.25, 0.46, 0.45, 0.94);
	}

	.zoom-content {
		position: relative;
		width: 100%;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	/* Book Section */
	.book-section {
		background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
	}

	.book-container {
		position: relative;
		width: 100vw;
		height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		transform: scale(1.5);
		transition: transform 1.5s cubic-bezier(0.25, 0.46, 0.45, 0.94);
	}

	.book-section.active .book-container {
		transform: scale(1);
	}

	.book-image {
		width: 100vw;
		height: 100vh;
		object-fit: cover;
		border-radius: 0;
		box-shadow: none;
		filter: drop-shadow(0 0 30px rgba(255, 215, 0, 0.3));
	}

	.book-glow {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: radial-gradient(circle, rgba(255, 215, 0, 0.2) 0%, transparent 70%);
		border-radius: 0;
		animation: glow-pulse 3s ease-in-out infinite;
	}

	@keyframes glow-pulse {
		0%, 100% { opacity: 0.5; transform: scale(1); }
		50% { opacity: 0.8; transform: scale(1.05); }
	}

	/* New York Section */
	.newyork-section {
		background: linear-gradient(135deg, #2c5530 0%, #3d7c47 50%, #4a9960 100%);
	}

	/* USA 1830 Section */
	.usa1830-section {
		background: linear-gradient(135deg, #8b4513 0%, #a0522d 50%, #cd853f 100%);
	}

	/* Modern USA Section */
	.usa-modern-section {
		background: linear-gradient(135deg, #1e3c72 0%, #2a5298 50%, #3b82f6 100%);
	}

	/* World Section */
	.world-section {
		background: linear-gradient(135deg, #134e5e 0%, #71b280 50%, #a8e6cf 100%);
	}

	/* Globe Section */
	.globe-section {
		background-image: url('/src/lib/HOMEPAGE_TEST_ASSETS/The_Globe_World_Background.png');
		background-size: cover;
		background-position: center;
		background-repeat: no-repeat;
		position: relative;
	}

	.globe-section::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="20" cy="20" r="1" fill="white" opacity="0.8"/><circle cx="80" cy="30" r="0.5" fill="white" opacity="0.6"/><circle cx="40" cy="60" r="1.5" fill="white" opacity="0.4"/><circle cx="70" cy="80" r="1" fill="white" opacity="0.7"/><circle cx="10" cy="80" r="0.8" fill="white" opacity="0.5"/></svg>');
		animation: stars-twinkle 10s linear infinite;
	}

	.globe-container {
		position: relative;
		width: 100%;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.globe {
		position: relative;
		width: 600px;
		height: 600px;
		border-radius: 50%;
		overflow: hidden;
	}

	.globe-image {
		width: 100%;
		height: 100%;
		border-radius: 50%;
		animation: globe-rotate 20s linear infinite;
	}

	.globe-atmosphere {
		position: absolute;
		top: -10px;
		left: -10px;
		right: -10px;
		bottom: -10px;
		border-radius: 50%;
		background: radial-gradient(circle, transparent 45%, rgba(135, 206, 235, 0.3) 55%, rgba(135, 206, 235, 0.1) 70%, transparent 80%);
		pointer-events: none;
	}

	@keyframes globe-rotate {
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
	}

	@keyframes stars-twinkle {
		0%, 100% { opacity: 0.8; }
		50% { opacity: 0.3; }
	}

	/* Map and Globe Containers */
	.map-container {
		position: relative;
		width: 100vw;
		height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.map-image {
		width: 100vw;
		height: 100vh;
		object-fit: cover;
		transition: transform 1.5s cubic-bezier(0.25, 0.46, 0.45, 0.94);
	}

	.map-image.vintage {
		filter: sepia(0.3) contrast(1.2);
	}

	/* Location Markers */
	.location-marker {
		position: absolute;
		width: 20px;
		height: 20px;
		transform: translate(-50%, -50%);
	}

	.marker-dot {
		width: 12px;
		height: 12px;
		background: #ff6b6b;
		border-radius: 50%;
		position: relative;
		z-index: 2;
		box-shadow: 0 0 10px rgba(255, 107, 107, 0.8);
	}

	.marker-pulse {
		position: absolute;
		top: -4px;
		left: -4px;
		width: 20px;
		height: 20px;
		border: 2px solid #ff6b6b;
		border-radius: 50%;
		animation: marker-pulse 2s ease-out infinite;
	}

	@keyframes marker-pulse {
		0% { transform: scale(0.5); opacity: 1; }
		100% { transform: scale(2); opacity: 0; }
	}

	/* Marker Positions */
	.hill-cumorah { top: 35%; left: 85%; }
	.palmyra { top: 25%; left: 75%; }
	.kirtland { top: 30%; left: 70%; }
	.utah { top: 45%; left: 25%; }
	.missouri { top: 50%; left: 55%; }
	.americas { top: 60%; left: 30%; }
	.middle-east { top: 40%; left: 75%; }

	/* Content Overlays */
	.content-overlay {
		position: absolute;
		z-index: 10;
		color: white;
		text-align: center;
		max-width: 800px;
		padding: 2.5rem;
		background: rgba(0, 0, 0, 0.4);
		backdrop-filter: blur(15px);
		border-radius: 20px;
		border: 1px solid rgba(255, 255, 255, 0.2);
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
	}

	.book-section .content-overlay {
		top: 10%;
		left: 50%;
		transform: translateX(-50%);
		background: rgba(0, 0, 0, 0.5);
		backdrop-filter: blur(20px);
	}

	.newyork-section .content-overlay,
	.usa1830-section .content-overlay,
	.usa-modern-section .content-overlay,
	.world-section .content-overlay {
		bottom: 10%;
		left: 50%;
		transform: translateX(-50%);
	}

	.globe-section .content-overlay {
		top: 15%;
		left: 50%;
		transform: translateX(-50%);
		background: rgba(0, 20, 40, 0.6);
		backdrop-filter: blur(20px);
		border: 1px solid rgba(135, 206, 235, 0.3);
	}

	/* Typography */
	.main-title {
		font-size: 4rem;
		font-weight: 800;
		margin-bottom: 1.5rem;
		text-shadow: 3px 3px 6px rgba(0, 0, 0, 0.8);
		filter: drop-shadow(0 0 10px rgba(255, 215, 0, 0.3));
	}

	.title-line {
		display: block;
		background: linear-gradient(135deg, #ffd700 0%, #ffed4e 100%);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
		text-shadow: none;
	}

	.section-title {
		font-size: 3rem;
		font-weight: 700;
		margin-bottom: 1.5rem;
		text-shadow: 3px 3px 6px rgba(0, 0, 0, 0.9);
		color: #ffffff;
		filter: drop-shadow(0 0 8px rgba(255, 255, 255, 0.2));
	}

	.subtitle,
	.section-description {
		font-size: 1.5rem;
		margin-bottom: 2rem;
		opacity: 1;
		color: rgba(255, 255, 255, 0.95);
		text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
		line-height: 1.6;
	}

	/* Buttons */
	.cta-button,
	.explore-button {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 1rem 2rem;
		background: linear-gradient(135deg, #ff6b6b 0%, #ee5a24 100%);
		color: white;
		border: none;
		border-radius: 50px;
		font-size: 1.1rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.3s ease;
		box-shadow: 0 10px 30px rgba(255, 107, 107, 0.4);
	}

	.cta-button:hover,
	.explore-button:hover {
		transform: translateY(-3px);
		box-shadow: 0 15px 40px rgba(255, 107, 107, 0.6);
	}

	.join-button {
		display: inline-block;
		padding: 1rem 2rem;
		background: transparent;
		color: white;
		border: 2px solid white;
		border-radius: 50px;
		text-decoration: none;
		font-weight: 600;
		margin-left: 1rem;
		transition: all 0.3s ease;
	}

	.join-button:hover {
		background: white;
		color: #004e92;
	}

	/* Info Cards and Stats */
	.info-card {
		background: rgba(255, 255, 255, 0.15);
		backdrop-filter: blur(20px);
		padding: 2rem;
		border-radius: 20px;
		margin-top: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.3);
		box-shadow: 0 15px 35px rgba(0, 0, 0, 0.2);
	}

	.info-card h3 {
		color: #ffd700;
		font-weight: 700;
		font-size: 1.3rem;
		margin-bottom: 1rem;
		text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.7);
	}

	.info-card p {
		color: rgba(255, 255, 255, 0.95);
		line-height: 1.6;
		text-shadow: 1px 1px 3px rgba(0, 0, 0, 0.8);
	}

	.timeline-info {
		display: flex;
		gap: 2rem;
		justify-content: center;
		margin-top: 2rem;
	}

	.timeline-item {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		text-align: center;
		background: rgba(255, 255, 255, 0.1);
		backdrop-filter: blur(15px);
		padding: 1.5rem;
		border-radius: 15px;
		border: 1px solid rgba(255, 255, 255, 0.2);
		box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
		min-width: 150px;
	}

	.year {
		font-size: 1.5rem;
		font-weight: 700;
		color: #ffd700;
		text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
		filter: drop-shadow(0 0 5px rgba(255, 215, 0, 0.4));
	}

	.event {
		font-size: 1rem;
		opacity: 1;
		color: rgba(255, 255, 255, 0.95);
		text-shadow: 1px 1px 3px rgba(0, 0, 0, 0.8);
		line-height: 1.4;
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 2rem;
		margin-top: 2rem;
	}

	.stat-item {
		text-align: center;
		background: rgba(255, 255, 255, 0.1);
		backdrop-filter: blur(15px);
		padding: 2rem 1.5rem;
		border-radius: 20px;
		border: 1px solid rgba(255, 255, 255, 0.2);
		box-shadow: 0 15px 30px rgba(0, 0, 0, 0.2);
		transition: transform 0.3s ease, box-shadow 0.3s ease;
	}

	.stat-item:hover {
		transform: translateY(-5px);
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
	}

	.stat-number {
		display: block;
		font-size: 2.5rem;
		font-weight: 800;
		color: #ffd700;
		text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
		filter: drop-shadow(0 0 8px rgba(255, 215, 0, 0.4));
		margin-bottom: 0.5rem;
	}

	.stat-label {
		font-size: 1rem;
		opacity: 1;
		color: rgba(255, 255, 255, 0.95);
		text-shadow: 1px 1px 3px rgba(0, 0, 0, 0.8);
		line-height: 1.4;
	}

	/* Navigation Dots */
	.scroll-navigation {
		position: fixed;
		right: 2rem;
		top: 50%;
		transform: translateY(-50%);
		display: flex;
		flex-direction: column;
		gap: 1rem;
		z-index: 100;
		background: rgba(0, 0, 0, 0.2);
		backdrop-filter: blur(10px);
		padding: 1rem 0.5rem;
		border-radius: 25px;
		border: 1px solid rgba(255, 255, 255, 0.2);
	}

	.nav-dot {
		width: 12px;
		height: 12px;
		border: 2px solid rgba(255, 255, 255, 0.7);
		border-radius: 50%;
		background: transparent;
		cursor: pointer;
		transition: all 0.3s ease;
		padding: 0;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
	}

	.nav-dot.active {
		background: white;
		border-color: white;
		transform: scale(1.2);
		box-shadow: 0 0 15px rgba(255, 255, 255, 0.5);
	}

	.nav-dot:hover {
		border-color: white;
		transform: scale(1.1);
		background: rgba(255, 255, 255, 0.3);
	}

	/* Scroll Indicator */
	.scroll-indicator {
		position: fixed;
		bottom: 2rem;
		left: 50%;
		transform: translateX(-50%);
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.5rem;
		color: white;
		z-index: 100;
		transition: opacity 0.3s ease;
		/*background: rgba(0, 0, 0, 0.3);
		backdrop-filter: blur(10px);
		padding: 1rem 1.5rem;*/
		border-radius: 25px;
		border: 1px solid rgba(255, 255, 255, 0.2);
	}

	.scroll-indicator.hidden {
		opacity: 0;
		pointer-events: none;
	}

	.scroll-text {
		font-size: 0.9rem;
		opacity: 1;
		color: rgba(255, 255, 255, 0.95);
		text-shadow: 1px 1px 3px rgba(0, 0, 0, 0.8);
		font-weight: 500;
	}

	.scroll-arrow {
		width: 24px;
		height: 24px;
		animation: bounce 2s ease-in-out infinite;
		filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.5));
	}

	@keyframes bounce {
		0%, 20%, 50%, 80%, 100% { transform: translateY(0); }
		40% { transform: translateY(-10px); }
		60% { transform: translateY(-5px); }
	}

	/* Zoom Animations */
	.zoom-section {
		transform: scale(1.1);
		opacity: 0.7;
	}

	.zoom-section.active {
		transform: scale(1);
		opacity: 1;
	}

	.zoom-section.past {
		transform: scale(0.9);
		opacity: 0.3;
	}

	.zoom-section.future {
		transform: scale(1.2);
		opacity: 0.5;
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.main-title {
			font-size: 2.5rem;
		}

		.section-title {
			font-size: 2rem;
		}

		.subtitle,
		.section-description {
			font-size: 1.2rem;
		}

		.content-overlay {
			padding: 2rem 1.5rem;
			max-width: 90%;
			margin: 0 auto;
		}

		.hero-title {
			font-size: 3rem;
		}

		.hero-quote {
			font-size: 1.2rem;
			padding: 0 1rem;
		}

		.navigation-cards {
			flex-direction: column;
			gap: 1.5rem;
			align-items: center;
			bottom: 5%;
		}

		.nav-card {
			width: 250px;
			min-height: 160px;
			padding: 0;
		}

		.nav-card-content {
			padding: 1rem;
		}

		.card-icon {
			width: 50px;
			height: 50px;
			margin-bottom: 1rem;
		}

		.card-action {
			font-size: 1.8rem;
		}

		.nav-card h3 {
			font-size: 1.2rem;
		}

		.nav-card p {
			font-size: 0.9rem;
		}

		.book-image {
			width: 400px;
			height: 300px;
		}

		.globe {
			width: 400px;
			height: 400px;
		}

		.timeline-item {
			min-width: 200px;
			max-width: 250px;
		}

		.stats-grid {
			grid-template-columns: 1fr;
			gap: 1rem;
			max-width: 300px;
			margin: 2rem auto 0;
		}

		.scroll-navigation {
			right: 1rem;
			padding: 0.8rem 0.4rem;
		}

		.scroll-indicator {
			padding: 0.8rem 1.2rem;
		}

		.final-cta {
			display: flex;
			flex-direction: column;
			gap: 1rem;
			align-items: center;
		}

		.join-button {
			margin-left: 0;
		}
	}

	@media (max-width: 480px) {
		.main-title {
			font-size: 2rem;
		}

		.section-title {
			font-size: 1.5rem;
		}

		.content-overlay {
			padding: 1.5rem 1rem;
			max-width: 95%;
		}

		.hero-title {
			font-size: 2.2rem;
		}

		.hero-quote {
			font-size: 1.1rem;
		}

		.navigation-cards {
			bottom: 3%;
		}

		.nav-card {
			width: 220px;
			min-height: 140px;
			padding: 0;
		}

		.nav-card-content {
			padding: 1rem;
		}

		.card-icon {
			width: 40px;
			height: 40px;
		}

		.card-action {
			font-size: 1.8rem;
		}

		.nav-card h3 {
			font-size: 1.2rem;
		}

		.nav-card p {
			font-size: 0.85rem;
		}

		.book-image {
			width: 300px;
			height: 200px;
		}

		.globe {
			width: 300px;
			height: 300px;
		}

		.info-card {
			padding: 1.5rem;
		}

		.timeline-item {
			padding: 1rem;
			min-width: 180px;
		}

		.stat-item {
			padding: 1.5rem 1rem;
		}

		.scroll-navigation {
			right: 0.5rem;
			padding: 0.6rem 0.3rem;
		}

		.scroll-indicator {
			padding: 0.6rem 1rem;
		}

		.scroll-text {
			font-size: 0.8rem;
		}
	}

	/* Fixed Header for Slide 1 */
	.home_page_fixed_header {
		position: fixed;
		top: 0;
		left: 50%;
		transform: translateX(-50%);
		padding: 2rem 0;
		text-align: center;
		color: white;
		width: 51vw;
		height: 6vh;
		background: rgba(150, 142, 132, 0.5);
		border-radius: 0 0 15px 15px;
		box-shadow: 10px 10px 20px rgba(0, 0, 0, 0.2),
					-10px -10px 20px rgba(13, 3, 41, 0.5),
					inset 0 0 10px rgba(255, 255, 255, 0.5),
					inset 0 0 10px rgba(0, 0, 0, 0.2);
		backdrop-filter: blur(10px);
		pointer-events: auto;
		z-index: 50;
	}

	/* Hero Title and Quote */
	.main-title-container {
		position: absolute;
		top: 15%;
		left: 50%;
		transform: translateX(-50%);
		text-align: center;
		z-index: 15;
		max-width: 2000px;
		background: rgba(14, 18, 70, 0.75);
		border-radius: 25px;
		padding: 1rem 5rem 0 5rem;
	}

	.hero-title {
		font-family: 'Garamond', Times, serif;
		font-weight: 900;
		color: #ffd700;
		text-shadow: 4px 4px 8px rgba(0, 0, 0, 0.9);
		filter: drop-shadow(0 0 15px rgba(255, 215, 0, 0.4));
		margin-bottom: 2rem;
		line-height: 0.75;
		padding: 1rem 1.5rem;
	}

	.hero-title-book {
		font-size: 4.5rem;
	}

	.hero-title-evidence {
		font-size: 13.5rem;
	}

	.hero-quote {
		font-size: 1.4rem;
		color: rgba(255, 255, 255, 0.95);
		text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
		font-style: italic;
		line-height: 1.6;
		margin-bottom: 3rem;
		max-width: 1800px;
		margin-left: auto;
		margin-right: auto;
		
	}

	.hero-quote cite {
		display: block;
		margin-top: 1rem;
		font-size: 1.1rem;
		color: #ffd700;
		font-style: normal;
		font-weight: 600;
	}

	/* Navigation Cards */
	.navigation-cards {
		position: absolute;
		bottom: 11%;
		left: 50%;
		transform: translateX(-50%);
		display: flex;
		gap: 5rem;
		z-index: 15;
	}

	.nav-card {
		background: rgba(150, 142, 132, 0.5);
		backdrop-filter: blur(10px);
		border-radius: 15px;
		padding: 0;
		text-align: center;
		color: white;
		text-decoration: none;
		transition: all 0.3s ease;
		border: 1px solid rgba(255, 255, 255, 0.2);
		box-shadow: 10px 10px 20px rgba(0, 0, 0, 0.2),
					-10px -10px 20px rgba(13, 3, 41, 0.5),
					inset 0 0 10px rgba(255, 255, 255, 0.5),
					inset 0 0 10px rgba(0, 0, 0, 0.2);
		width: 400px;
		min-height: 400px;
		display: flex;
		flex-direction: column;
		justify-content: flex-start;
		align-items: stretch;
		overflow: hidden;
	}

	.nav-card:hover {
		transform: translateY(-5px);
		box-shadow: 15px 15px 30px rgba(0, 0, 0, 0.3),
					-15px -15px 30px rgba(13, 3, 41, 0.6),
					inset 0 0 15px rgba(255, 255, 255, 0.6),
					inset 0 0 15px rgba(0, 0, 0, 0.3);
		background: rgba(150, 142, 132, 0.7);
	}

	.nav-card:hover .card-action {
		background: rgba(255, 215, 0, 0.2);
	}

	.card-action {
		font-size: 3.1rem;
		font-weight: 900;
		color: #ffd700;
		text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.9);
		filter: drop-shadow(0 0 8px rgba(255, 215, 0, 0.5));
		letter-spacing: 2px;
		text-transform: uppercase;
		width: 100%;
		height: 20%;
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 0;
		background: rgba(255, 215, 0, 0.1);
		border-radius: 15px 15px 0 0;
		border-bottom: 2px solid rgba(255, 215, 0, 0.3);
		flex-shrink: 0;
	}

	.nav-card-content {
		flex: 1;
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		padding: 1.5rem;
	}

	.card-icon {
		width: 60px;
		height: 60px;
		margin-bottom: 1rem;
		color: #ffd700;
		filter: drop-shadow(0 0 8px rgba(255, 215, 0, 0.4));
	}

	.nav-card h3 {
		font-size: 1.3rem;
		font-weight: 700;
		margin-bottom: 1rem;
		color: white;
		text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
	}

	.nav-card p {
		font-size: 1rem;
		color: rgba(255, 255, 255, 0.9);
		text-shadow: 1px 1px 3px rgba(0, 0, 0, 0.8);
		line-height: 1.4;
		margin: 0;
	}
</style>
