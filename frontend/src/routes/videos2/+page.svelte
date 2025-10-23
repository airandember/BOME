<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import Navigation from '$lib/components/Navigation.svelte';
	import Footer from '$lib/components/Footer.svelte';

	let featuredContent = {
		title: "Beauty & Killer",
		genre: "Triller",
		duration: "2h40m",
		year: "2021",
		description: "A screenplay synopsis is a brief summary of a script's plot, characters, and themes. It serves as a marketing tool to attract producers, investors, and talent. A well-written synopsis should be concise, engaging, and capture the essence of the story while highlighting its unique elements and commercial potential.",
		poster: "/placeholder-movie-poster.jpg"
	};

	let contentRows = [
		{
			title: "Trending Now",
			type: "portrait",
			items: Array(8).fill(null).map((_, i) => ({
				id: i + 1,
				title: `Trending Movie ${i + 1}`,
				poster: "/placeholder-poster.jpg"
			}))
		},
		{
			title: "Continue Watching",
			type: "landscape",
			items: Array(6).fill(null).map((_, i) => ({
				id: i + 1,
				title: `Continue Movie ${i + 1}`,
				poster: "/placeholder-poster.jpg",
				progress: i === 0 ? 65 : 0 // First item has progress
			}))
		},
		{
			title: "You'll love to see",
			type: "landscape",
			items: Array(6).fill(null).map((_, i) => ({
				id: i + 1,
				title: `Recommended Movie ${i + 1}`,
				poster: "/placeholder-poster.jpg"
			}))
		},
		{
			title: "New Added",
			type: "portrait",
			items: Array(8).fill(null).map((_, i) => ({
				id: i + 1,
				title: `New Movie ${i + 1}`,
				poster: "/placeholder-poster.jpg"
			}))
		}
	];

	function handleWatchNow() {
		console.log('Watch Now clicked');
		// TODO: Implement watch functionality
	}

	function handleSeeMore() {
		console.log('See More clicked');
		// TODO: Implement see more functionality
	}

	function handleContentClick(item: any) {
		console.log('Content clicked:', item);
		// TODO: Navigate to content detail page
	}
</script>

<svelte:head>
	<title>Videos - BOME</title>
</svelte:head>

<Navigation />

<main class="videos-page">
	<!-- Hero Section -->
	<section class="hero-section">
		<div class="hero-content">
			<div class="hero-text">
				<h1 class="hero-title">{featuredContent.title}</h1>
				<div class="hero-meta">
					<span class="genre">{featuredContent.genre}</span>
					<span class="separator">•</span>
					<span class="duration">{featuredContent.duration}</span>
					<span class="separator">•</span>
					<span class="year">{featuredContent.year}</span>
				</div>
				<p class="hero-description">{featuredContent.description}</p>
				<div class="hero-actions">
					<button class="btn btn-primary" on:click={handleWatchNow}>
						Watch Now
					</button>
					<button class="btn btn-secondary" on:click={handleSeeMore}>
						See More
					</button>
				</div>
			</div>
			<div class="hero-poster">
				<div class="poster-placeholder">
					<svg width="120" height="120" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
						<rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
						<circle cx="8.5" cy="8.5" r="1.5"/>
						<polyline points="21,15 16,10 5,21"/>
					</svg>
				</div>
			</div>
		</div>
	</section>

	<!-- Content Rows -->
	<section class="content-section">
		{#each contentRows as row}
			<div class="content-row">
				<h2 class="row-title">{row.title}</h2>
				<div class="content-carousel">
					{#each row.items as item}
						<div 
							class="content-card {row.type}" 
							on:click={() => handleContentClick(item)}
							role="button"
							tabindex="0"
						>
							<div class="card-poster">
								<div class="poster-placeholder">
									<svg width="60" height="60" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
										<rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
										<circle cx="8.5" cy="8.5" r="1.5"/>
										<polyline points="21,15 16,10 5,21"/>
									</svg>
								</div>
								{#if item.progress && item.progress > 0}
									<div class="progress-bar">
										<div class="progress-fill" style="width: {item.progress}%"></div>
									</div>
								{/if}
							</div>
							<div class="card-title">{item.title}</div>
						</div>
					{/each}
				</div>
			</div>
		{/each}
	</section>
</main>

<Footer />

<style>
	:root {
		--bg-primary: #0f0f0f;
		--bg-secondary: #1a1a1a;
		--bg-card: #2a2a2a;
		--text-primary: #ffffff;
		--text-secondary: #b3b3b3;
		--accent-blue: #0071eb;
		--accent-blue-dark: #0056b3;
		--border-color: #333333;
		--shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
	}

	.videos-page {
		background: var(--bg-primary);
		min-height: 100vh;
		color: var(--text-primary);
	}

	/* Hero Section */
	.hero-section {
		background: linear-gradient(135deg, #1a1a1a 0%, #2a2a2a 100%);
		padding: 4rem 2rem;
		min-height: 60vh;
		display: flex;
		align-items: center;
	}

	.hero-content {
		max-width: 1200px;
		margin: 0 auto;
		display: grid;
		grid-template-columns: 1fr 300px;
		gap: 3rem;
		align-items: center;
	}

	.hero-text {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.hero-title {
		font-size: 3.5rem;
		font-weight: 700;
		color: var(--accent-blue);
		text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.5);
		margin: 0;
		line-height: 1.1;
	}

	.hero-meta {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 1.1rem;
		color: var(--text-secondary);
	}

	.separator {
		color: var(--text-secondary);
		opacity: 0.6;
	}

	.hero-description {
		font-size: 1.1rem;
		line-height: 1.6;
		color: var(--text-secondary);
		max-width: 600px;
		margin: 0;
	}

	.hero-actions {
		display: flex;
		gap: 1rem;
		margin-top: 1rem;
	}

	.btn {
		padding: 0.875rem 2rem;
		border-radius: 8px;
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		border: none;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}

	.btn-primary {
		background: var(--accent-blue);
		color: white;
	}

	.btn-primary:hover {
		background: var(--accent-blue-dark);
		transform: translateY(-2px);
		box-shadow: 0 8px 16px rgba(0, 113, 235, 0.3);
	}

	.btn-secondary {
		background: transparent;
		color: var(--text-primary);
		border: 2px solid var(--border-color);
	}

	.btn-secondary:hover {
		background: var(--bg-card);
		border-color: var(--text-primary);
		transform: translateY(-2px);
	}

	.hero-poster {
		display: flex;
		justify-content: center;
		align-items: center;
	}

	.poster-placeholder {
		width: 100%;
		height: 400px;
		background: var(--bg-card);
		border-radius: 12px;
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--text-secondary);
		border: 2px dashed var(--border-color);
	}

	.poster-placeholder svg {
		opacity: 0.5;
	}

	/* Content Section */
	.content-section {
		padding: 2rem;
		max-width: 1200px;
		margin: 0 auto;
	}

	.content-row {
		margin-bottom: 3rem;
	}

	.row-title {
		font-size: 1.5rem;
		font-weight: 600;
		margin: 0 0 1.5rem 0;
		color: var(--text-primary);
	}

	.content-carousel {
		display: flex;
		gap: 1rem;
		overflow-x: auto;
		padding-bottom: 1rem;
		scrollbar-width: thin;
		scrollbar-color: var(--accent-blue) transparent;
	}

	.content-carousel::-webkit-scrollbar {
		height: 8px;
	}

	.content-carousel::-webkit-scrollbar-track {
		background: transparent;
	}

	.content-carousel::-webkit-scrollbar-thumb {
		background: var(--accent-blue);
		border-radius: 4px;
	}

	.content-carousel::-webkit-scrollbar-thumb:hover {
		background: var(--accent-blue-dark);
	}

	.content-card {
		flex-shrink: 0;
		cursor: pointer;
		transition: transform 0.2s ease;
		background: var(--bg-card);
		border-radius: 8px;
		overflow: hidden;
		box-shadow: var(--shadow);
	}

	.content-card:hover {
		transform: translateY(-4px);
		box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
	}

	.content-card.portrait {
		width: 200px;
		height: 300px;
	}

	.content-card.landscape {
		width: 300px;
		height: 180px;
	}

	.card-poster {
		position: relative;
		width: 100%;
		height: calc(100% - 50px);
		background: var(--bg-secondary);
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.card-poster .poster-placeholder {
		width: 100%;
		height: 100%;
		background: var(--bg-secondary);
		border: none;
		border-radius: 0;
	}

	.card-poster .poster-placeholder svg {
		opacity: 0.3;
	}

	.progress-bar {
		position: absolute;
		bottom: 0;
		left: 0;
		right: 0;
		height: 4px;
		background: rgba(255, 255, 255, 0.2);
	}

	.progress-fill {
		height: 100%;
		background: var(--accent-blue);
		transition: width 0.3s ease;
	}

	.card-title {
		padding: 0.75rem;
		font-size: 0.9rem;
		font-weight: 500;
		color: var(--text-primary);
		text-align: center;
		height: 50px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--bg-card);
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.hero-section {
			padding: 2rem 1rem;
			min-height: 50vh;
		}

		.hero-content {
			grid-template-columns: 1fr;
			gap: 2rem;
			text-align: center;
		}

		.hero-title {
			font-size: 2.5rem;
		}

		.hero-poster {
			order: -1;
		}

		.poster-placeholder {
			height: 250px;
		}

		.content-section {
			padding: 1rem;
		}

		.content-card.portrait {
			width: 150px;
			height: 225px;
		}

		.content-card.landscape {
			width: 250px;
			height: 150px;
		}

		.hero-actions {
			justify-content: center;
		}
	}

	@media (max-width: 480px) {
		.hero-title {
			font-size: 2rem;
		}

		.hero-description {
			font-size: 1rem;
		}

		.btn {
			padding: 0.75rem 1.5rem;
			font-size: 0.9rem;
		}

		.content-card.portrait {
			width: 120px;
			height: 180px;
		}

		.content-card.landscape {
			width: 200px;
			height: 120px;
		}
	}
</style>
