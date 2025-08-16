<script lang="ts">
	import { createEventDispatcher, onMount, onDestroy } from 'svelte';
	import { fly, fade, scale } from 'svelte/transition';
	import { quintOut, backOut } from 'svelte/easing';
	import type { YouTubeVideo } from '$lib/types/youtube';

	export let video: YouTubeVideo | null = null;
	export let isOpen = false;

	const dispatch = createEventDispatcher();
	let videoElement: HTMLIFrameElement;
	let modalElement: HTMLDivElement;
	let isClosing = false;

	// Close modal function
	function closeModal() {
		if (isClosing) return;
		isClosing = true;
		
		// Stop the video by reloading the iframe
		if (videoElement) {
			const src = videoElement.src;
			videoElement.src = '';
			setTimeout(() => {
				if (videoElement) videoElement.src = src;
			}, 100);
		}
		
		setTimeout(() => {
			isOpen = false;
			isClosing = false;
			dispatch('close');
		}, 300);
	}

	// Handle escape key
	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' && isOpen) {
			closeModal();
		}
	}

	// Handle click outside
	function handleBackdropClick(event: MouseEvent) {
		if (event.target === modalElement) {
			closeModal();
		}
	}

	// Prevent body scroll when modal is open
	$: if (typeof document !== 'undefined') {
		if (isOpen) {
			document.body.style.overflow = 'hidden';
		} else {
			document.body.style.overflow = '';
		}
	}

	onMount(() => {
		document.addEventListener('keydown', handleKeydown);
		return () => {
			document.removeEventListener('keydown', handleKeydown);
		};
	});

	onDestroy(() => {
		if (typeof document !== 'undefined') {
			document.body.style.overflow = '';
		}
	});

	// Generate embed URL
	$: embedUrl = video ? `https://www.youtube.com/embed/${video.id}?autoplay=1&rel=0&modestbranding=1&fs=1&cc_load_policy=1` : '';
</script>

{#if isOpen && video}
	<!-- Backdrop -->
	<div 
		class="modal-backdrop" 
		bind:this={modalElement}
		on:click={handleBackdropClick}
		on:keydown={(e) => e.key === 'Enter' && closeModal()}
		role="dialog"
		aria-modal="true"
		aria-label="Video player modal"
		tabindex="-1"
		transition:fade={{ duration: 300, easing: quintOut }}
	>
		<!-- Modal Container -->
		<div 
			class="modal-container"
			transition:scale={{ 
				duration: 400, 
				easing: backOut,
				start: 0.8,
				opacity: 0
			}}
		>
			<!-- Close Button -->
			<button 
				class="close-button"
				on:click={closeModal}
				aria-label="Close video player"
				transition:fly={{ y: -20, duration: 300, delay: 200 }}
			>
				<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<line x1="18" y1="6" x2="6" y2="18"></line>
					<line x1="6" y1="6" x2="18" y2="18"></line>
				</svg>
			</button>

			<!-- Video Container -->
			<div 
				class="video-container"
				transition:fly={{ y: 30, duration: 400, delay: 100, easing: quintOut }}
			>
				<iframe
					bind:this={videoElement}
					src={embedUrl}
					title={video.title}
					frameborder="0"
					allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
					allowfullscreen
					class="video-iframe"
				></iframe>
			</div>

			<!-- Video Info -->
			<div 
				class="video-info"
				transition:fly={{ y: 20, duration: 300, delay: 300 }}
			>
				<h2 class="video-title">{video.title}</h2>
				<div class="video-meta">
					<span class="publish-date">
						{new Date(video.published_at).toLocaleDateString('en-US', {
							year: 'numeric',
							month: 'long',
							day: 'numeric'
						})}
					</span>
					{#if video.view_count > 0}
						<span class="view-count">
							{video.view_count.toLocaleString()} views
						</span>
					{/if}
				</div>
				{#if video.description}
					<p class="video-description">{video.description}</p>
				{/if}
			</div>
		</div>
	</div>
{/if}

<style>
	.modal-backdrop {
		position: fixed;
		top: 0;
		left: 0;
		width: 100vw;
		height: 100vh;
		background: rgba(0, 0, 0, 0.95);
		backdrop-filter: blur(10px);
		-webkit-backdrop-filter: blur(10px);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 9999;
		padding: var(--space-xl);
	}

	.modal-container {
		position: relative;
		width: 100%;
		max-width: 1200px;
		max-height: 90vh;
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.close-button {
		position: absolute;
		top: -60px;
		right: 0;
		background: rgba(255, 255, 255, 0.1);
		backdrop-filter: blur(10px);
		-webkit-backdrop-filter: blur(10px);
		border: 2px solid rgba(255, 255, 255, 0.2);
		border-radius: 50%;
		width: 48px;
		height: 48px;
		display: flex;
		align-items: center;
		justify-content: center;
		color: white;
		cursor: pointer;
		transition: all var(--transition-normal);
		z-index: 10;
	}

	.close-button:hover {
		background: rgba(255, 255, 255, 0.2);
		border-color: rgba(255, 255, 255, 0.4);
		transform: scale(1.1);
	}

	.video-container {
		position: relative;
		width: 100%;
		aspect-ratio: 16/9;
		border-radius: var(--radius-2xl);
		overflow: hidden;
		box-shadow: 
			0 25px 50px -12px rgba(0, 0, 0, 0.8),
			0 0 0 1px rgba(255, 255, 255, 0.1);
		background: var(--bg-secondary);
	}

	.video-iframe {
		width: 100%;
		height: 100%;
		border: none;
		border-radius: var(--radius-2xl);
	}

	.video-info {
		background: rgba(255, 255, 255, 0.05);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		color: white;
		max-height: 200px;
		overflow-y: auto;
	}

	.video-title {
		font-size: var(--text-2xl);
		font-weight: 700;
		margin: 0 0 var(--space-md) 0;
		line-height: var(--leading-tight);
		color: white;
	}

	.video-meta {
		display: flex;
		gap: var(--space-lg);
		margin-bottom: var(--space-md);
		font-size: var(--text-sm);
		color: rgba(255, 255, 255, 0.8);
		flex-wrap: wrap;
	}

	.video-description {
		font-size: var(--text-sm);
		line-height: var(--leading-relaxed);
		color: rgba(255, 255, 255, 0.9);
		margin: 0;
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.modal-backdrop {
			padding: var(--space-md);
		}

		.modal-container {
			max-height: 95vh;
		}

		.close-button {
			top: -50px;
			width: 40px;
			height: 40px;
		}

		.video-info {
			padding: var(--space-lg);
			max-height: 150px;
		}

		.video-title {
			font-size: var(--text-xl);
		}
	}

	@media (max-width: 480px) {
		.modal-backdrop {
			padding: var(--space-sm);
		}

		.close-button {
			top: -45px;
			width: 36px;
			height: 36px;
		}

		.video-info {
			padding: var(--space-md);
		}

		.video-title {
			font-size: var(--text-lg);
		}
	}

	/* Custom scrollbar for video info */
	.video-info::-webkit-scrollbar {
		width: 6px;
	}

	.video-info::-webkit-scrollbar-track {
		background: rgba(255, 255, 255, 0.1);
		border-radius: 3px;
	}

	.video-info::-webkit-scrollbar-thumb {
		background: rgba(255, 255, 255, 0.3);
		border-radius: 3px;
	}

	.video-info::-webkit-scrollbar-thumb:hover {
		background: rgba(255, 255, 255, 0.5);
	}
</style>
