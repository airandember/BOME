<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	export let analytics: any;
	export let AbsoluteTotalVideos: number;
	export let checkingConflicts: boolean;

	const dispatch = createEventDispatcher();

	// Computed properties for video stats
	$: syncedCount = analytics?.video_stats?.synced_videos || 0;
	$: needsAttentionCount = analytics?.video_stats?.needs_attention || 0;
	$: conflictCount = analytics?.video_stats?.videos_by_sync_status?.conflict || 0;

	function checkConflicts() {
		dispatch('checkConflicts');
	}

	function showTagAnalytics() {
		dispatch('showTagAnalytics');
	}
</script>

<!-- Stats Summary -->
<div class="stats-grid">
	<div class="stat-card primary">
		<div class="stat-icon">
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<polygon points="23,7 16,12 23,17 23,7"></polygon>
				<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
			</svg>
		</div>
		<div class="stat-content">
			<p class="stat-label">Total Videos</p>
			<p class="stat-value">{AbsoluteTotalVideos}</p>
		</div>
	</div>
	<div class="stat-card success">
		<div class="stat-icon">
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M9 12l2 2 4-4"/>
				<path d="M21 12c0 4.97-4.03 9-9 9s-9-4.03-9-9 4.03-9 9-9 9 4.03 9 9z"/>
			</svg>
		</div>
		<div class="stat-content">
			<p class="stat-label">Synced</p>
			<p class="stat-value">{syncedCount}</p>
		</div>
	</div>
	<div class="stat-card warning">
		<div class="stat-icon">
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
				<line x1="12" y1="9" x2="12" y2="13"/>
				<line x1="12" y1="17" x2="12.01" y2="17"/>
			</svg>
		</div>
		<div class="stat-content">
			<p class="stat-label">Found Conflict</p>
			<p class="stat-value">{conflictCount}</p>
		</div>
	</div>
	<div class="vid_buttons" style="gap: 1rem; display: flex; flex-direction: column;">
		<div class="rabbitButtons" style="display: flex; flex-direction: row; gap: 1rem;">
			<button 
				class="btn btn-secondary" 
				style="background: linear-gradient(135deg, #f59e0b, #d97706); "
				on:click={checkConflicts} 
				disabled={checkingConflicts}
			>
				{#if checkingConflicts}
					<div class="bouncing-rabbit-container">
						<div class="bouncing-rabbit">🐇</div>
					</div>
					<span style="color: white !important; font-size: clamp(10px, 0.8vw, 1.5rem);">Checking...</span>
				{:else}
					<span style="color: white !important; font-size: clamp(12px, 1vw, 2.5rem);">Run Check<br> <span class="bouncing">🐇</span></span>
				{/if}
			</button>
			
			<button class="bt btn-secondary"
				style="background: linear-gradient(135deg, #f59e0b, #d97706); border-radius: 10px; padding: 0.5rem 1rem;"
				on:click={showTagAnalytics}
			>
				<span style="color: white !important; font-size: clamp(12px, 1vw, 2.5rem);">Tags<br>🏷️ </span>
			</button>
		</div>
		<a
			href="/admin/streaming/upload"
			class="btn bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors flex items-center space-x-2"
		>
			<svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
				<polyline points="7,10 12,15 17,10"></polyline>
				<line x1="12" y1="15" x2="12" y2="3"></line>
			</svg>
			<span style="color: white !important; font-size: clamp(12px, 1.5vw, 4rem">Upload</span>
		</a>
	</div>
</div>

<style>
	/* Stats Grid */
	.stats-grid {
		display: flex;
		width: 100%;
		justify-content: space-between;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 1.5rem;
		margin-bottom: 2rem;
	}

	.stat-card {
		background: var(--bg-glass, rgba(255, 255, 255, 0.1));
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: 0.75rem 1.5rem 0 1.5rem ;
		border: 1px solid rgba(255, 255, 255, 0.1);
		display: flex;
		align-items: center;
		gap: 1.5rem;
		transition: all 0.3s ease;
	}

	.stat-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
	}

	.stat-card.primary {
		border-left: 4px solid var(--primary-color, #3b82f6);
	}

	.stat-card.success {
		border-left: 4px solid #10b981;
	}

	.stat-card.warning {
		border-left: 4px solid #f59e0b;
	}

	.stat-card.info {
		border-left: 4px solid #3b82f6;
	}

	.stat-icon {
		width: 60px;
		height: 60px;
		border-radius: 15px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		background: var(--primary-gradient, linear-gradient(135deg, #3b82f6, #1d4ed8));
	}

	.stat-card.success .stat-icon {
		background: linear-gradient(135deg, #10b981, #059669);
	}

	.stat-card.warning .stat-icon {
		background: linear-gradient(135deg, #f59e0b, #d97706);
	}

	.stat-card.info .stat-icon {
		background: linear-gradient(135deg, #3b82f6, #1d4ed8);
	}

	.stat-icon svg {
		width: 28px;
		height: 28px;
		color: white;
	}

	.stat-content {
		flex: 1;
	}

	.rabbitButtons {
		display: flex;
		flex-direction: row;
	}

	.stat-value {
		font-size: 2rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 0.25rem;
	}

	.stat-label {
		font-size: 0.9rem;
		color: var(--text-secondary);
		margin-bottom: 0.5rem;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-weight: 600;
	}

	.stat-change {
		font-size: 0.8rem;
		font-weight: 600;
		color: #10b981;
	}

	/* Bouncing Rabbit Animation */
	.bouncing-rabbit-container {
		width: 60px;
		height: 24px;
		position: relative;
		overflow: hidden;
		margin: 0.25rem 0;
		overflow: visible;
	}

	.bouncing-rabbit {
		translate: scaleX(1);
		position: absolute;
		font-size: 1.5rem;
		animation: rabbitBounce 1.5s ease-in-out infinite;
		filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.3));
	}

	@keyframes rabbitBounce {
		0% {
			left: -30px;
			transform: translateY(0px) rotate(-10deg) scaleX(-1);
		}
		25% {
			left: 15px;
			transform: translateY(-8px) rotate(0deg);
		}
		50% {
			left: 45px;
			transform: translateY(0px) rotate(10deg) scaleX(-1);
		}
		75% {
			left: 75px;
			transform: translateY(-6px) rotate(0deg) scaleX(-1);
		}
		100% {
			left: 190px;
			transform: translateY(0px) rotate(10deg) scaleX(-1);
		}
	}

	/* Button Styles */
	.btn {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 8px;
		font-weight: 600;
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.3s ease;
		text-decoration: none;
	}

	.btn-primary {
		background: var(--primary-gradient, linear-gradient(135deg, #3b82f6, #1d4ed8));
		color: white;
	}

	.btn-primary:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 20px rgba(99, 102, 241, 0.3);
	}

	.btn-secondary {
		background: var(--bg-glass, rgba(255, 255, 255, 0.1));
		color: var(--text-primary);
		border: 1px solid rgba(255, 255, 255, 0.2);
		backdrop-filter: blur(20px);
	}

	.btn-secondary:hover {
		background: rgba(255, 255, 255, 0.2);
		transform: translateY(-2px);
	}

	.bg-blue-600 {
		background: var(--primary-gradient, linear-gradient(135deg, #3b82f6, #1d4ed8)) !important;
	}

	.bg-blue-600:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 20px rgba(59, 130, 246, 0.3);
	}

	.bg-blue-600 svg {
		color: white;
		width: 20px;
		height: 20px;
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.stats-grid {
			grid-template-columns: 1fr;
		}

		.stat-card {
			flex-direction: column;
			text-align: center;
		}
	}

	@media (max-width: 480px) {
		.stat-icon {
			width: 50px;
			height: 50px;
		}

		.stat-icon svg {
			width: 24px;
			height: 24px;
		}

		.stat-value {
			font-size: 1.5rem;
		}
	}
</style>
