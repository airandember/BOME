<script lang="ts">
	import { onMount, createEventDispatcher } from 'svelte';
	import { videoUtils } from '$lib/video';
	import { analytics } from '$lib/services/analytics';

	export let videoUrl: string;
	export let poster: string = '';
	export let autoplay: boolean = false;
	export let controls: boolean = true;
	export let width: string = '100%';
	export let height: string = 'auto';

	const dispatch = createEventDispatcher();

	let videoElement: HTMLVideoElement;
	let isPlaying = false;
	let currentTime = 0;
	let duration = 0;
	let volume = 1;
	let isMuted = false;
	let showControls = true;
	let controlsTimeout: number;
	let videoId: string;
	let quarterWatched = false;
	let halfWatched = false;
	let threeQuartersWatched = false;

	onMount(() => {
		if (videoElement) {
			videoElement.addEventListener('loadedmetadata', handleLoadedMetadata);
			videoElement.addEventListener('timeupdate', handleTimeUpdate);
			videoElement.addEventListener('ended', handleEnded);
			videoElement.addEventListener('play', handlePlay);
			videoElement.addEventListener('pause', handlePause);
		}
	});

	function handleLoadedMetadata() {
		duration = videoElement.duration;
		dispatch('loadedmetadata', { duration });
	}

	function handleTimeUpdate() {
		currentTime = videoElement.currentTime;
		const progress = Math.round((currentTime / duration) * 100);
		
		// Track progress at 25%, 50%, 75%
		if (progress >= 25 && !quarterWatched) {
			quarterWatched = true;
			analytics.trackVideoEvent(videoId, 'progress', {
				milestone: '25%',
				currentTime,
				duration
			});
		} else if (progress >= 50 && !halfWatched) {
			halfWatched = true;
			analytics.trackVideoEvent(videoId, 'progress', {
				milestone: '50%',
				currentTime,
				duration
			});
		} else if (progress >= 75 && !threeQuartersWatched) {
			threeQuartersWatched = true;
			analytics.trackVideoEvent(videoId, 'progress', {
				milestone: '75%',
				currentTime,
				duration
			});
		}
		dispatch('timeupdate', { currentTime, duration });
	}

	function handleEnded() {
		analytics.trackVideoEvent(videoId, 'complete', {
			duration: videoElement.duration || 0
		});
		isPlaying = false;
		dispatch('ended');
	}

	function handlePlay() {
		analytics.trackVideoEvent(videoId, 'play', {
			currentTime: videoElement.currentTime || 0,
			duration: videoElement.duration || 0
		});
		isPlaying = true;
		dispatch('play');
	}

	function handlePause() {
		analytics.trackVideoEvent(videoId, 'pause', {
			currentTime: videoElement.currentTime || 0,
			duration: videoElement.duration || 0
		});
		isPlaying = false;
		dispatch('pause');
	}

	function togglePlay() {
		if (isPlaying) {
			videoElement.pause();
		} else {
			videoElement.play();
		}
	}

	function toggleMute() {
		isMuted = !isMuted;
		videoElement.muted = isMuted;
	}

	function setVolume(value: number) {
		volume = value;
		videoElement.volume = value;
		if (value === 0) {
			isMuted = true;
		} else {
			isMuted = false;
		}
	}

	function seekTo(value: number) {
		videoElement.currentTime = value;
	}

	function toggleFullscreen() {
		if (document.fullscreenElement) {
			document.exitFullscreen();
		} else {
			videoElement.requestFullscreen();
		}
	}

	function showControlsTemporarily() {
		showControls = true;
		clearTimeout(controlsTimeout);
		controlsTimeout = setTimeout(() => {
			if (isPlaying) {
				showControls = false;
			}
		}, 3000);
	}

	function handleMouseMove() {
		if (controls) {
			showControlsTemporarily();
		}
	}

	function handleKeydown(event: KeyboardEvent) {
		switch (event.code) {
			case 'Space':
				event.preventDefault();
				togglePlay();
				break;
			case 'ArrowLeft':
				event.preventDefault();
				seekTo(Math.max(0, currentTime - 10));
				break;
			case 'ArrowRight':
				event.preventDefault();
				seekTo(Math.min(duration, currentTime + 10));
				break;
			case 'ArrowUp':
				event.preventDefault();
				setVolume(Math.min(1, volume + 0.1));
				break;
			case 'ArrowDown':
				event.preventDefault();
				setVolume(Math.max(0, volume - 0.1));
				break;
			case 'KeyM':
				event.preventDefault();
				toggleMute();
				break;
			case 'KeyF':
				event.preventDefault();
				toggleFullscreen();
				break;
		}
	}
</script>

<svelte:window on:keydown={handleKeydown} />

<div 
	class="video-player"
	style="width: {width}; height: {height};"
	on:mousemove={handleMouseMove}
	on:mouseleave={() => showControls = false}
>
	<video
		bind:this={videoElement}
		{poster}
		{autoplay}
		{controls}
		preload="metadata"
		class="video-element"
	>
		<source src={videoUrl} type="video/mp4" />
		Your browser does not support the video tag.
	</video>

	{#if controls && showControls}
		<div class="video-controls">
			<div class="progress-bar">
				<div 
					class="progress-fill"
					style="width: {(currentTime / duration) * 100}%"
				></div>
				<input
					type="range"
					min="0"
					max={duration || 0}
					value={currentTime}
					on:input={(e) => {
						const target = e.target as HTMLInputElement;
						if (target) seekTo(parseFloat(target.value));
					}}
					class="progress-slider"
				/>
			</div>

			<div class="controls-main">
				<div class="controls-left">
					<button 
						class="control-btn"
						on:click={togglePlay}
						aria-label={isPlaying ? 'Pause' : 'Play'}
					>
						{#if isPlaying}
							⏸️
						{:else}
							▶️
						{/if}
					</button>

					<div class="time-display">
						<span>{videoUtils.formatDuration(currentTime)}</span>
						<span>/</span>
						<span>{videoUtils.formatDuration(duration)}</span>
					</div>
				</div>

				<div class="controls-right">
					<div class="volume-control">
						<button 
							class="control-btn"
							on:click={toggleMute}
							aria-label={isMuted ? 'Unmute' : 'Mute'}
						>
							{#if isMuted || volume === 0}
								🔇
							{:else if volume < 0.5}
								🔉
							{:else}
								🔊
							{/if}
						</button>
						<input
							type="range"
							min="0"
							max="1"
							step="0.1"
							value={isMuted ? 0 : volume}
							on:input={(e) => {
								const target = e.target as HTMLInputElement;
								if (target) setVolume(parseFloat(target.value));
							}}
							class="volume-slider"
						/>
					</div>

					<button 
						class="control-btn"
						on:click={toggleFullscreen}
						aria-label="Toggle fullscreen"
					>
						⛶
					</button>
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	.video-player {
		position: relative;
		background: var(--card-bg);
		border-radius: 16px;
		overflow: hidden;
		box-shadow: 
			8px 8px 16px var(--shadow-dark),
			-8px -8px 16px var(--shadow-light);
	}

	.video-element {
		width: 100%;
		height: 100%;
		display: block;
	}

	.video-controls {
		position: absolute;
		bottom: 0;
		left: 0;
		right: 0;
		background: linear-gradient(transparent, rgba(0, 0, 0, 0.7));
		padding: 1rem;
		transition: opacity 0.3s ease;
	}

	.progress-bar {
		position: relative;
		width: 100%;
		height: 4px;
		background: rgba(255, 255, 255, 0.3);
		border-radius: 2px;
		margin-bottom: 1rem;
		cursor: pointer;
	}

	.progress-fill {
		height: 100%;
		background: var(--accent-color);
		border-radius: 2px;
		transition: width 0.1s ease;
	}

	.progress-slider {
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		opacity: 0;
		cursor: pointer;
	}

	.controls-main {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.controls-left,
	.controls-right {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.control-btn {
		background: none;
		border: none;
		color: white;
		font-size: 1.2rem;
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 8px;
		transition: background-color 0.2s ease;
	}

	.control-btn:hover {
		background: rgba(255, 255, 255, 0.1);
	}

	.time-display {
		color: white;
		font-size: 0.9rem;
		font-weight: 500;
	}

	.volume-control {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.volume-slider {
		width: 60px;
		height: 4px;
		background: rgba(255, 255, 255, 0.3);
		border-radius: 2px;
		outline: none;
		cursor: pointer;
	}

	.volume-slider::-webkit-slider-thumb {
		appearance: none;
		width: 12px;
		height: 12px;
		background: white;
		border-radius: 50%;
		cursor: pointer;
	}

	.volume-slider::-moz-range-thumb {
		width: 12px;
		height: 12px;
		background: white;
		border-radius: 50%;
		cursor: pointer;
		border: none;
	}

	@media (max-width: 768px) {
		.controls-main {
			flex-direction: column;
			gap: 1rem;
		}

		.controls-left,
		.controls-right {
			width: 100%;
			justify-content: center;
		}

		.volume-control {
			display: none;
		}
	}
</style> 