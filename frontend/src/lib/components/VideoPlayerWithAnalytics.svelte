<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { videoAnalytics } from '$lib/services/videoAnalytics';
	import type { WatchHistory } from '$lib/services/videoAnalytics';
	
	// Props
	export let videoId: number;
	export let videoUrl: string;
	export let title: string = '';
	export let autoResume: boolean = true;
	export let showResumePrompt: boolean = true;
	export let onComplete: (() => void) | null = null;
	
	// State
	let player: HTMLVideoElement;
	let isTracking = true;
	let hasResumePoint = false;
	let resumePosition = 0;
	let showResumeDialog = false;
	let isLoading = true;
	
	/**
	 * Initialize video player with analytics
	 */
	onMount(async () => {
		console.log(`🎬 [Video Player] Mounted for video ${videoId}`);
		
		// Check for existing watch history
		if (autoResume) {
			await loadWatchHistory();
		}
		
		isLoading = false;
	});
	
	/**
	 * Load watch history and offer to resume
	 */
	async function loadWatchHistory() {
		try {
			const history: WatchHistory | null = await videoAnalytics.getWatchHistory(videoId);
			
			if (history && history.last_position > 10 && !history.completed) {
				resumePosition = history.last_position;
				hasResumePoint = true;
				
				if (showResumePrompt) {
					showResumeDialog = true;
				} else {
					// Auto-resume without prompt
					resumeFromPosition();
				}
				
				console.log(`📍 [Video Player] Resume point found: ${resumePosition}s`);
			}
		} catch (error) {
			console.error('❌ [Video Player] Failed to load watch history:', error);
		}
	}
	
	/**
	 * Resume from saved position
	 */
	function resumeFromPosition() {
		if (player && resumePosition > 0) {
			player.currentTime = resumePosition;
			console.log(`▶️ [Video Player] Resumed from ${resumePosition}s`);
		}
		showResumeDialog = false;
	}
	
	/**
	 * Start from beginning
	 */
	function startFromBeginning() {
		resumePosition = 0;
		showResumeDialog = false;
		if (player) {
			player.currentTime = 0;
			player.play();
		}
	}
	
	/**
	 * Handle time update (main tracking event)
	 */
	function handleTimeUpdate() {
		if (!player || !isTracking) return;
		
		const { currentTime, duration } = player;
		
		if (duration && duration > 0) {
			videoAnalytics.trackProgress(videoId, currentTime, duration);
		}
	}
	
	/**
	 * Handle video ended
	 */
	function handleEnded() {
		if (!isTracking) return;
		
		console.log(`🏁 [Video Player] Video ${videoId} ended`);
		
		if (player && player.duration) {
			videoAnalytics.markComplete(videoId, player.duration);
		}
		
		// Call optional completion callback
		if (onComplete) {
			onComplete();
		}
	}
	
	/**
	 * Handle play event
	 */
	function handlePlay() {
		isTracking = true;
		console.log(`▶️ [Video Player] Playing video ${videoId}`);
	}
	
	/**
	 * Handle pause event
	 */
	function handlePause() {
		// Continue tracking even when paused (records pause duration)
		console.log(`⏸️ [Video Player] Paused video ${videoId}`);
	}
	
	/**
	 * Handle loaded metadata (video ready)
	 */
	function handleLoadedMetadata() {
		console.log(`📊 [Video Player] Metadata loaded for video ${videoId}`);
		isLoading = false;
	}
	
	/**
	 * Handle error
	 */
	function handleError(e: Event) {
		console.error('❌ [Video Player] Error:', e);
		isLoading = false;
	}
	
	/**
	 * Format seconds to MM:SS
	 */
	function formatTime(seconds: number): string {
		const mins = Math.floor(seconds / 60);
		const secs = Math.floor(seconds % 60);
		return `${mins}:${secs.toString().padStart(2, '0')}`;
	}
	
	/**
	 * Cleanup on destroy
	 */
	onDestroy(() => {
		isTracking = false;
		videoAnalytics.resetTracking(videoId);
		console.log(`🧹 [Video Player] Cleaned up video ${videoId}`);
	});
</script>

<!-- Resume Dialog -->
{#if showResumeDialog}
	<div class="resume-dialog-overlay">
		<div class="resume-dialog">
			<h3>Continue Watching?</h3>
			<p>
				You've watched {Math.floor((resumePosition / player?.duration || 1) * 100)}% of this video.
			</p>
			<p class="resume-time">
				Resume from <strong>{formatTime(resumePosition)}</strong>
			</p>
			<div class="resume-actions">
				<button class="btn-secondary" on:click={startFromBeginning}>
					Start Over
				</button>
				<button class="btn-primary" on:click={resumeFromPosition}>
					Resume
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Video Player -->
<div class="video-player-container">
	{#if isLoading}
		<div class="loading-overlay">
			<div class="loading-spinner"></div>
			<p>Loading video...</p>
		</div>
	{/if}
	
	<video
		bind:this={player}
		src={videoUrl}
		on:timeupdate={handleTimeUpdate}
		on:ended={handleEnded}
		on:play={handlePlay}
		on:pause={handlePause}
		on:loadedmetadata={handleLoadedMetadata}
		on:error={handleError}
		controls
		controlsList="nodownload"
		class="video-player"
		class:hidden={isLoading}
	>
		<track kind="captions" />
		Your browser does not support the video tag.
	</video>
	
	{#if title}
		<div class="video-info">
			<h2 class="video-title">{title}</h2>
		</div>
	{/if}
</div>

<style>
	.video-player-container {
		position: relative;
		width: 100%;
		background: #000;
		border-radius: 8px;
		overflow: hidden;
	}
	
	.video-player {
		width: 100%;
		height: auto;
		display: block;
	}
	
	.video-player.hidden {
		display: none;
	}
	
	.loading-overlay {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		background: #000;
		color: #fff;
		z-index: 10;
	}
	
	.loading-spinner {
		width: 50px;
		height: 50px;
		border: 4px solid rgba(255, 255, 255, 0.3);
		border-top-color: #fff;
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-bottom: 1rem;
	}
	
	@keyframes spin {
		to { transform: rotate(360deg); }
	}
	
	.video-info {
		padding: 1rem;
		background: linear-gradient(to bottom, transparent, rgba(0,0,0,0.8));
		position: absolute;
		bottom: 50px;
		left: 0;
		right: 0;
	}
	
	.video-title {
		margin: 0;
		font-size: 1.25rem;
		font-weight: 600;
		color: #fff;
		text-shadow: 0 2px 4px rgba(0,0,0,0.5);
	}
	
	/* Resume Dialog */
	.resume-dialog-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.8);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
		animation: fadeIn 0.2s ease-out;
	}
	
	@keyframes fadeIn {
		from { opacity: 0; }
		to { opacity: 1; }
	}
	
	.resume-dialog {
		background: #fff;
		border-radius: 12px;
		padding: 2rem;
		max-width: 400px;
		box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
		animation: slideUp 0.3s ease-out;
	}
	
	@keyframes slideUp {
		from {
			opacity: 0;
			transform: translateY(20px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
	
	.resume-dialog h3 {
		margin: 0 0 1rem 0;
		font-size: 1.5rem;
		color: #333;
	}
	
	.resume-dialog p {
		margin: 0.5rem 0;
		color: #666;
		line-height: 1.5;
	}
	
	.resume-time {
		font-size: 1.1rem;
		color: #333;
		margin: 1rem 0 1.5rem 0;
	}
	
	.resume-time strong {
		color: #d4a574;
		font-weight: 600;
	}
	
	.resume-actions {
		display: flex;
		gap: 1rem;
		margin-top: 1.5rem;
	}
	
	.btn-primary,
	.btn-secondary {
		flex: 1;
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 6px;
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s;
	}
	
	.btn-primary {
		background: linear-gradient(135deg, #d4a574 0%, #c89860 100%);
		color: #fff;
	}
	
	.btn-primary:hover {
		background: linear-gradient(135deg, #c89860 0%, #b8874f 100%);
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(212, 165, 116, 0.4);
	}
	
	.btn-secondary {
		background: #f0f0f0;
		color: #333;
	}
	
	.btn-secondary:hover {
		background: #e0e0e0;
	}
	
	/* Responsive */
	@media (max-width: 768px) {
		.resume-dialog {
			margin: 1rem;
			padding: 1.5rem;
		}
		
		.resume-actions {
			flex-direction: column;
		}
		
		.video-title {
			font-size: 1rem;
		}
	}
</style>

