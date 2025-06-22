<!-- Lazy Loading Image Component with Progressive Enhancement -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { imageCache } from '$lib/utils/cache';

	export let src: string;
	export let alt: string = '';
	export let placeholder: string = '';
	export let width: number | string = 'auto';
	export let height: number | string = 'auto';
	export let className: string = '';
	export let loading: 'lazy' | 'eager' = 'lazy';
	export let quality: 'low' | 'medium' | 'high' = 'medium';
	export let fallback: string = '/images/placeholder.svg';
	export let progressive: boolean = true;
	export let cacheKey: string = '';

	let imgElement: HTMLImageElement;
	let loaded = false;
	let error = false;
	let inView = false;
	let observer: IntersectionObserver;
	
	// Generate different quality versions
	$: lowQualitySrc = progressive ? generateQualityUrl(src, 'low') : src;
	$: mediumQualitySrc = progressive ? generateQualityUrl(src, 'medium') : src;
	$: highQualitySrc = progressive ? generateQualityUrl(src, 'high') : src;
	
	// Determine which source to use
	$: currentSrc = getCurrentSrc();
	
	function generateQualityUrl(url: string, targetQuality: 'low' | 'medium' | 'high'): string {
		if (!url || !url.includes('bunny')) return url;
		
		// For Bunny.net URLs, add quality parameters
		const qualityParams = {
			low: '?width=400&quality=60',
			medium: '?width=800&quality=80',
			high: '?width=1200&quality=95'
		};
		
		return url + (qualityParams[targetQuality] || '');
	}
	
	function getCurrentSrc(): string {
		if (!inView && loading === 'lazy') return placeholder || '';
		
		if (progressive) {
			// Progressive loading: start with low quality, upgrade based on viewport
			if (window.innerWidth < 768) return lowQualitySrc;
			if (window.innerWidth < 1200) return mediumQualitySrc;
			return highQualitySrc;
		}
		
		return quality === 'low' ? lowQualitySrc : 
		       quality === 'medium' ? mediumQualitySrc : 
		       highQualitySrc;
	}

	onMount(() => {
		// Check cache first
		const cached = cacheKey ? imageCache.get(cacheKey) : null;
		if (cached) {
			loaded = true;
			return;
		}

		// Set up intersection observer for lazy loading
		if (loading === 'lazy' && 'IntersectionObserver' in window) {
			observer = new IntersectionObserver(
				(entries) => {
					entries.forEach((entry) => {
						if (entry.isIntersecting) {
							inView = true;
							observer.unobserve(entry.target);
							loadImage();
						}
					});
				},
				{
					rootMargin: '50px', // Start loading 50px before entering viewport
					threshold: 0.1
				}
			);

			if (imgElement) {
				observer.observe(imgElement);
			}
		} else {
			// Fallback for browsers without IntersectionObserver or eager loading
			inView = true;
			loadImage();
		}

		return () => {
			if (observer) {
				observer.disconnect();
			}
		};
	});

	async function loadImage() {
		if (loaded || error) return;

		try {
			// Preload the image
			const img = new Image();
			img.crossOrigin = 'anonymous';
			
			img.onload = () => {
				loaded = true;
				error = false;
				
				// Cache the successful load
				if (cacheKey) {
					imageCache.set(cacheKey, { 
						src: currentSrc, 
						loaded: true, 
						timestamp: Date.now() 
					});
				}
			};
			
			img.onerror = () => {
				error = true;
				console.warn(`Failed to load image: ${currentSrc}`);
			};
			
			img.src = currentSrc;
		} catch (err) {
			error = true;
			console.error('Image loading error:', err);
		}
	}

	function handleLoad() {
		loaded = true;
		error = false;
	}

	function handleError() {
		error = true;
		console.warn(`Image failed to load: ${currentSrc}`);
	}

	function handleRetry() {
		error = false;
		loaded = false;
		loadImage();
	}
</script>

<div 
	class="lazy-image-container {className}"
	style="width: {typeof width === 'number' ? width + 'px' : width}; height: {typeof height === 'number' ? height + 'px' : height};"
>
	{#if error}
		<!-- Error state with retry option -->
		<div class="image-error">
			<div class="error-content">
				<div class="error-icon">⚠️</div>
				<p class="error-message">Failed to load image</p>
				<button class="retry-button" on:click={handleRetry}>
					Retry
				</button>
			</div>
		</div>
	{:else if !loaded && (loading === 'lazy' && !inView)}
		<!-- Placeholder state -->
		<div class="image-placeholder">
			{#if placeholder}
				<img 
					src={placeholder} 
					{alt}
					class="placeholder-img"
					loading="eager"
				/>
			{:else}
				<div class="placeholder-skeleton">
					<div class="skeleton-shimmer"></div>
				</div>
			{/if}
		</div>
	{:else}
		<!-- Main image with progressive enhancement -->
		<img
			bind:this={imgElement}
			src={currentSrc || fallback}
			{alt}
			class="lazy-image"
			class:loaded
			class:loading={!loaded}
			on:load={handleLoad}
			on:error={handleError}
			loading={loading}
			decoding="async"
		/>
		
		{#if !loaded}
			<div class="loading-overlay">
				<div class="loading-spinner">
					<div class="spinner"></div>
				</div>
			</div>
		{/if}
	{/if}
</div>

<style>
	.lazy-image-container {
		position: relative;
		overflow: hidden;
		background: var(--bg-secondary);
		border-radius: 8px;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.lazy-image {
		width: 100%;
		height: 100%;
		object-fit: cover;
		transition: opacity 0.3s ease, filter 0.3s ease;
		opacity: 0;
		filter: blur(5px);
	}

	.lazy-image.loaded {
		opacity: 1;
		filter: blur(0);
	}

	.lazy-image.loading {
		opacity: 0.7;
	}

	.image-placeholder {
		width: 100%;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--bg-secondary);
	}

	.placeholder-img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		opacity: 0.6;
	}

	.placeholder-skeleton {
		width: 100%;
		height: 100%;
		background: linear-gradient(90deg, 
			var(--bg-secondary) 25%, 
			var(--bg-tertiary) 50%, 
			var(--bg-secondary) 75%
		);
		background-size: 200% 100%;
		position: relative;
		overflow: hidden;
	}

	.skeleton-shimmer {
		position: absolute;
		top: 0;
		left: -100%;
		width: 100%;
		height: 100%;
		background: linear-gradient(90deg, 
			transparent, 
			rgba(255, 255, 255, 0.1), 
			transparent
		);
		animation: shimmer 2s infinite;
	}

	@keyframes shimmer {
		0% { left: -100%; }
		100% { left: 100%; }
	}

	.loading-overlay {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		background: rgba(0, 0, 0, 0.1);
		backdrop-filter: blur(2px);
	}

	.loading-spinner {
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.spinner {
		width: 24px;
		height: 24px;
		border: 2px solid var(--border-color);
		border-top: 2px solid var(--primary-color);
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.image-error {
		width: 100%;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--error-bg);
		color: var(--error-text);
	}

	.error-content {
		text-align: center;
		padding: 1rem;
	}

	.error-icon {
		font-size: 2rem;
		margin-bottom: 0.5rem;
	}

	.error-message {
		font-size: 0.9rem;
		margin-bottom: 1rem;
		color: var(--text-secondary);
	}

	.retry-button {
		background: var(--primary-color);
		color: white;
		border: none;
		padding: 0.5rem 1rem;
		border-radius: 6px;
		cursor: pointer;
		font-size: 0.8rem;
		transition: background 0.2s ease;
	}

	.retry-button:hover {
		background: var(--primary-hover);
	}

	/* Responsive optimizations */
	@media (max-width: 768px) {
		.lazy-image-container {
			border-radius: 6px;
		}
		
		.error-icon {
			font-size: 1.5rem;
		}
		
		.error-message {
			font-size: 0.8rem;
		}
	}

	/* Reduced motion support */
	@media (prefers-reduced-motion: reduce) {
		.lazy-image,
		.spinner,
		.skeleton-shimmer {
			animation: none;
			transition: none;
		}
		
		.lazy-image.loaded {
			opacity: 1;
			filter: none;
		}
	}

	/* High contrast mode support */
	@media (prefers-contrast: high) {
		.lazy-image-container {
			border: 1px solid currentColor;
		}
		
		.placeholder-skeleton {
			background: var(--text-secondary);
		}
	}
</style> 