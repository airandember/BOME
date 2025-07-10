<!-- Simplified Lazy Loading Image Component -->
<script lang="ts">
	import { onMount } from 'svelte';

	export let src: string;
	export let alt: string = '';
	export let placeholder: string = '';
	export let width: number | string = 'auto';
	export let height: number | string = 'auto';
	export let className: string = '';
	export let loading: 'lazy' | 'eager' = 'lazy';
	export let fallback: string = '/16X10_Placeholder_IMG.png';

	let imgElement: HTMLImageElement;
	let loaded = false;
	let error = false;
	let observer: IntersectionObserver;
	let shouldLoad = false;

	onMount(() => {
		if (loading === 'eager') {
			shouldLoad = true;
		} else if ('IntersectionObserver' in window) {
			observer = new IntersectionObserver(
				(entries) => {
					entries.forEach((entry) => {
						if (entry.isIntersecting) {
							shouldLoad = true;
							observer.unobserve(entry.target);
						}
					});
				},
				{
					rootMargin: '50px',
					threshold: 0.1
				}
			);

			if (imgElement) {
				observer.observe(imgElement);
			}
		} else {
			// Fallback for browsers without IntersectionObserver
			shouldLoad = true;
		}

		return () => {
			if (observer) {
				observer.disconnect();
			}
		};
	});

	function handleLoad() {
		loaded = true;
		error = false;
	}

	function handleError() {
		error = true;
		loaded = false;
	}

	// Determine which source to use
	$: currentSrc = shouldLoad ? (src || fallback) : (placeholder || fallback);
</script>

<div 
	class="lazy-image-container {className}"
	style="width: {typeof width === 'number' ? width + 'px' : width}; height: {typeof height === 'number' ? height + 'px' : height};"
>
	<!-- Show placeholder while not loaded -->
	{#if !loaded && placeholder && !shouldLoad}
		<img 
			src={placeholder} 
			{alt}
			class="placeholder-img"
			loading="eager"
			referrerpolicy="no-referrer"
		/>
	{/if}
	
	<!-- Main image -->
	<img
		bind:this={imgElement}
		src={currentSrc}
		{alt}
		class="main-image"
		class:loaded
		class:error
		loading={loading}
		referrerpolicy="no-referrer"
		on:load={handleLoad}
		on:error={handleError}
	/>
</div>

<style>
	.lazy-image-container {
		position: relative;
		overflow: hidden;
		background-color: #f0f0f0;
	}

	.placeholder-img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		position: absolute;
		top: 0;
		left: 0;
	}

	.main-image {
		width: 100%;
		height: 100%;
		object-fit: cover;
		transition: opacity 0.3s ease-in-out;
		opacity: 1;
	}

	.main-image.error {
		opacity: 0.7;
	}

	.main-image.loaded {
		opacity: 1;
	}
</style> 