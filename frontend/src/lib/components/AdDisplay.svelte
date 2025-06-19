<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { Advertisement, AdPlacement, AdServeResponse } from '$lib/types/advertising';
	
	export let placementId: number;
	export let className: string = '';
	export let fallbackContent: string = '';
	
	let ad: Advertisement | null = null;
	let placement: AdPlacement | null = null;
	let trackingData: any = null;
	let loading = true;
	let error: string | null = null;
	let adElement: HTMLElement;
	let impressionTracked = false;
	let viewTimer: number;
	let isVisible = false;
	
	onMount(async () => {
		await loadAd();
		setupViewTracking();
	});
	
	onDestroy(() => {
		if (viewTimer) clearTimeout(viewTimer);
	});
	
	async function loadAd() {
		try {
			const response = await fetch(`/api/v1/ads/serve/${placementId}`, {
				method: 'GET',
				headers: {
					'Content-Type': 'application/json',
					'User-Agent': navigator.userAgent,
					'Referer': window.location.href
				}
			});
			
			if (response.ok) {
				const data: { success: boolean; data: AdServeResponse } = await response.json();
				if (data.success && data.data.ad) {
					ad = data.data.ad;
					placement = data.data.placement;
					trackingData = data.data.tracking_data;
				}
			}
		} catch (err) {
			error = 'Failed to load advertisement';
			console.error('Ad loading error:', err);
		} finally {
			loading = false;
		}
	}
	
	function setupViewTracking() {
		if (!adElement) return;
		
		const observer = new IntersectionObserver(
			(entries) => {
				entries.forEach(entry => {
					if (entry.isIntersecting && !impressionTracked) {
						isVisible = true;
						trackImpression();
						
						// Track view duration after 1 second
						viewTimer = setTimeout(() => {
							trackViewDuration(1000);
						}, 1000);
					} else if (!entry.isIntersecting && isVisible) {
						isVisible = false;
						if (viewTimer) clearTimeout(viewTimer);
					}
				});
			},
			{ threshold: 0.5 }
		);
		
		observer.observe(adElement);
	}
	
	async function trackImpression() {
		if (!ad || impressionTracked) return;
		
		impressionTracked = true;
		
		try {
			await fetch(`/api/v1/ads/impression/${ad.id}`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					view_duration: 0,
					placement_id: placementId,
					user_agent: navigator.userAgent,
					referrer: document.referrer
				})
			});
		} catch (err) {
			console.error('Failed to track impression:', err);
		}
	}
	
	async function trackViewDuration(duration: number) {
		if (!ad) return;
		
		try {
			await fetch(`/api/v1/ads/impression/${ad.id}`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					view_duration: duration,
					placement_id: placementId
				})
			});
		} catch (err) {
			console.error('Failed to track view duration:', err);
		}
	}
	
	async function handleClick() {
		if (!ad) return;
		
		// Track click
		try {
			await fetch(`/api/v1/ads/click/${ad.id}`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					placement_id: placementId,
					user_agent: navigator.userAgent,
					referrer: document.referrer
				})
			});
		} catch (err) {
			console.error('Failed to track click:', err);
		}
		
		// Open ad URL
		window.open(ad.click_url, '_blank', 'noopener,noreferrer');
	}
	
	function getAdStyles() {
		if (!ad || !placement) return '';
		
		return `
			width: ${Math.min(ad.width, placement.max_width)}px;
			height: ${Math.min(ad.height, placement.max_height)}px;
			max-width: 100%;
		`;
	}
</script>

{#if loading}
	<div class="ad-container {className}" bind:this={adElement}>
		<div class="ad-loading">
			<div class="loading-spinner"></div>
			<span>Loading ad...</span>
		</div>
	</div>
{:else if error && fallbackContent}
	<div class="ad-container {className}" bind:this={adElement}>
		<div class="ad-fallback">
			{@html fallbackContent}
		</div>
	</div>
{:else if ad && placement}
	<div class="ad-container {className}" bind:this={adElement}>
		<div class="ad-content" style={getAdStyles()}>
			<div class="ad-label">Advertisement</div>
			
			<button class="ad-clickable" on:click={handleClick} aria-label="View advertisement">
				{#if ad.image_url}
					<img 
						src={ad.image_url} 
						alt={ad.title}
						class="ad-image"
						loading="lazy"
					/>
				{/if}
				
				<div class="ad-text">
					<h3 class="ad-title">{ad.title}</h3>
					{#if ad.content}
						<p class="ad-description">{ad.content}</p>
					{/if}
				</div>
				
				<div class="ad-overlay">
					<span class="ad-cta">Learn More</span>
				</div>
			</button>
		</div>
	</div>
{/if}

<style>
	.ad-container {
		display: flex;
		justify-content: center;
		align-items: center;
		margin: 1rem 0;
		position: relative;
	}
	
	.ad-loading {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.5rem;
		padding: 2rem;
		color: var(--text-secondary);
		font-size: 0.9rem;
	}
	
	.loading-spinner {
		width: 24px;
		height: 24px;
		border: 2px solid var(--border-color);
		border-top: 2px solid var(--accent-color);
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}
	
	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}
	
	.ad-fallback {
		padding: 1rem;
		background: var(--card-bg);
		border-radius: 8px;
		border: 1px solid var(--border-color);
		text-align: center;
		color: var(--text-secondary);
	}
	
	.ad-content {
		position: relative;
		background: var(--card-bg);
		border-radius: 12px;
		overflow: hidden;
		box-shadow: 
			4px 4px 8px var(--shadow-dark),
			-2px -2px 4px var(--shadow-light);
		transition: all 0.3s ease;
		border: 1px solid var(--border-color);
	}
	
	.ad-content:hover {
		transform: translateY(-2px);
		box-shadow: 
			6px 6px 12px var(--shadow-dark),
			-3px -3px 6px var(--shadow-light);
	}
	
	.ad-label {
		position: absolute;
		top: 8px;
		right: 8px;
		background: rgba(0, 0, 0, 0.7);
		color: white;
		padding: 2px 6px;
		border-radius: 4px;
		font-size: 0.7rem;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		z-index: 2;
	}
	
	.ad-clickable {
		display: block;
		width: 100%;
		height: 100%;
		border: none;
		background: none;
		cursor: pointer;
		position: relative;
		overflow: hidden;
	}
	
	.ad-image {
		width: 100%;
		height: auto;
		display: block;
		object-fit: cover;
	}
	
	.ad-text {
		position: absolute;
		bottom: 0;
		left: 0;
		right: 0;
		background: linear-gradient(transparent, rgba(0, 0, 0, 0.8));
		color: white;
		padding: 1rem;
		text-align: left;
	}
	
	.ad-title {
		font-size: 1rem;
		font-weight: 600;
		margin: 0 0 0.25rem 0;
		line-height: 1.3;
	}
	
	.ad-description {
		font-size: 0.8rem;
		margin: 0;
		opacity: 0.9;
		line-height: 1.4;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	
	.ad-overlay {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		opacity: 0;
		transition: opacity 0.3s ease;
	}
	
	.ad-clickable:hover .ad-overlay {
		opacity: 1;
	}
	
	.ad-cta {
		background: var(--accent-color);
		color: white;
		padding: 0.5rem 1rem;
		border-radius: 6px;
		font-weight: 600;
		font-size: 0.9rem;
		transform: translateY(10px);
		transition: transform 0.3s ease;
	}
	
	.ad-clickable:hover .ad-cta {
		transform: translateY(0);
	}
	
	/* Responsive adjustments */
	@media (max-width: 768px) {
		.ad-content {
			max-width: 100%;
		}
		
		.ad-title {
			font-size: 0.9rem;
		}
		
		.ad-description {
			font-size: 0.75rem;
		}
	}
	
	/* Placement-specific styles */
	.ad-container.header {
		margin: 0;
		padding: 0.5rem 0;
	}
	
	.ad-container.sidebar {
		margin: 1rem 0;
		max-width: 300px;
	}
	
	.ad-container.footer {
		margin: 1rem 0 0 0;
	}
	
	.ad-container.video-overlay {
		position: absolute;
		top: 1rem;
		right: 1rem;
		margin: 0;
		z-index: 10;
	}
	
	.ad-container.between-videos {
		margin: 2rem 0;
		padding: 1rem;
		background: var(--bg-secondary);
		border-radius: 12px;
	}
</style> 