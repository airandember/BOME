<script lang="ts">
	import { onMount, createEventDispatcher } from 'svelte';
	import { videoUtils } from '$lib/video';
	import { analytics } from '$lib/services/analytics';

	export let videoId: string = '';
	export let title: string = '';
	export let poster: string = '';
	export let autoplay: boolean = false;
	export let controls: boolean = true;
	export let width: string = '100%';
	export let height: string = 'auto';
	export let playbackUrl: string = '';

	const dispatch = createEventDispatcher();

	let videoElement: HTMLVideoElement;
	let isPlaying = false;
	let currentTime = 0;
	let duration = 0;
	let volume = 1;
	let isMuted = false;
	let showControls = true;
	let controlsTimeout: number;
	let quarterWatched = false;
	let halfWatched = false;
	let threeQuartersWatched = false;

	$: videoUrl = videoId ? `https://iframe.mediadelivery.net/embed/${videoId}?autoplay=${autoplay}` : '';

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
	{#if videoId && playbackUrl}
		<video
			bind:this={videoElement}
			id="main-video"
			class="video-element"
			preload="auto"
			crossorigin="anonymous"
			{autoplay}
			{controls}
			{poster}
			playsinline
			data-plyr-config={{
				title,
				controls: ['play-large', 'play', 'progress', 'current-time', 'mute', 'volume', 'captions', 'settings', 'pip', 'airplay', 'fullscreen']
			}}
		>
			<source src={playbackUrl} type="application/x-mpegURL">
			Your browser does not support the video tag.
		</video>
	{:else}
		<div class="video-placeholder">
			<p>Video not available</p>
		</div>
	{/if}

	{#if controls && showControls}
		<div class="video-controls" class:visible={showControls}>
			<div class="progress-bar">
				<input
					type="range"
					min="0"
					max={duration}
					value={currentTime}
					on:input={(e) => {
						const target = e.target as HTMLInputElement;
						if (target) {
							seekTo(parseFloat(target.value));
						}
					}}
				/>
			</div>

			<div class="controls-bottom">
				<button on:click={togglePlay}>
					{isPlaying ? '⏸️' : '▶️'}
				</button>

				<div class="time-display">
					{videoUtils.formatDuration(currentTime)} / {videoUtils.formatDuration(duration)}
				</div>

				<div class="volume-control">
					<button on:click={toggleMute}>
						{isMuted ? '🔇' : '🔊'}
					</button>
					<input
						type="range"
						min="0"
						max="1"
						step="0.1"
						value={volume}
						on:input={(e) => {
							const target = e.target as HTMLInputElement;
							if (target) {
								setVolume(parseFloat(target.value));
							}
						}}
					/>
				</div>

				<button on:click={toggleFullscreen}>
					⛶
				</button>
			</div>
		</div>
	{/if}
</div>

<style lang="postcss">
	.video-player {
		position: relative;
		background: #000;
		overflow: hidden;
		border-radius: 8px;
	}

	.video-element {
		width: 100%;
		height: 100%;
		object-fit: contain;
	}

	.video-placeholder {
		width: 100%;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		background: #1a1a1a;
		color: #fff;
		min-height: 200px;
	}

	.video-controls {
		position: absolute;
		bottom: 0;
		left: 0;
		right: 0;
		background: linear-gradient(transparent, rgba(0, 0, 0, 0.7));
		padding: 10px;
		opacity: 0;
		transition: opacity 0.3s;
	}

	.video-controls.visible {
		opacity: 1;
	}

	.progress-bar {
		width: 100%;
		margin-bottom: 10px;
	}

	.progress-bar input {
		width: 100%;
	}

	.controls-bottom {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.time-display {
		color: white;
		font-size: 14px;
	}

	.volume-control {
		display: flex;
		align-items: center;
		gap: 5px;
	}

	button {
		background: none;
		border: none;
		color: white;
		cursor: pointer;
		padding: 5px;
		font-size: 18px;
	}

	button:hover {
		opacity: 0.8;
	}

	input[type="range"] {
		cursor: pointer;
	}
</style> 