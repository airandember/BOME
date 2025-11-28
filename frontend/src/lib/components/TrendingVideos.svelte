<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { videoAnalytics } from '$lib/services/videoAnalytics';
	
	// Types
	interface TrendingVideo {
		video_id: number;
		title: string;
		thumbnail_url: string;
		last_24h_views?: number;
		trending_score?: number;
		total_views?: number;
		unique_viewers?: number;
		avg_completion?: number;
		duration?: number;
	}
	
	// Props
	export let limit: number = 100;
	export let showTitle: boolean = true;
	export let autoRefresh: boolean = false;
	export let refreshInterval: number = 60000; // 1 minute
	
	// State
	let videos: TrendingVideo[] = [];
	let isLoading = true;
	let error: string | null = null;
	let lastUpdated: Date | null = null;
	let refreshTimer: NodeJS.Timeout | null = null;
	let viewMode: 'trending' | 'most_watched' = 'trending';
	//let timePeriod: 'week' | 'month' | 'all-time' = 'month'; // For Most Watched mode
	let timePeriod = 'all-time';

	onMount(() => {
		loadVideos();
		
		if (autoRefresh) {
			startAutoRefresh();
		}
		
		return () => {
			if (refreshTimer) {
				clearInterval(refreshTimer);
			}
		};
	});
	
	async function loadVideos() {
		error = null;
		isLoading = true;
		
		try {
			if (viewMode === 'trending') {
				const data = await videoAnalytics.getTrendingVideos(limit);
				videos = data;
				console.log(`🔥 [Trending] Loaded ${videos.length} trending videos`);
			} else {
				// Calculate days based on time period
				let days: number;
				switch (timePeriod) {
					case 'week':
						days = 7;
						break;
					case 'month':
						days = 30;
						break;
					case 'all-time':
						days = 36500; // 100 years (essentially all-time)
						break;
					default:
						days = 30;
				}
				
				const data = await videoAnalytics.getTopVideos(25, days);
				videos = data;
				console.log(`📊 [Most Watched] Loaded ${videos.length} top videos (${timePeriod})`);
			}
			lastUpdated = new Date();
		} catch (err) {
			console.error(`❌ [Videos] Failed to load ${viewMode}:`, err);
			error = `Failed to load ${viewMode === 'trending' ? 'trending' : 'most watched'} videos`;
		} finally {
			isLoading = false;
		}
	}
	
	async function switchViewMode(mode: 'trending' | 'most_watched') {
		viewMode = mode;
		await loadVideos();
	}
	
	async function switchTimePeriod(period: 'week' | 'month' | 'all-time') {
		timePeriod = period;
		await loadVideos();
	}
	
	function getTimePeriodLabel(): string {
		switch (timePeriod) {
			case 'week':
				return 'Top 25 most viewed this week';
			case 'month':
				return 'Top 25 most viewed this month';
			case 'all-time':
				return 'Top 25 most viewed all-time';
			default:
				return 'Top 25 most viewed videos';
		}
	}
	
	function startAutoRefresh() {
		refreshTimer = setInterval(() => {
			console.log('🔄 [Videos] Auto-refreshing...');
			loadVideos();
		}, refreshInterval);
	}
	
	function handleVideoClick(videoId: number) {
		goto(`/videos/${videoId}`);
	}
	
	function getTrendingBadge(score: number): { text: string; class: string } {
		if (score >= 90) return { text: 'ON FIRE', class: 'badge-fire' };
		if (score >= 75) return { text: 'HOT', class: 'badge-hot' };
		if (score >= 60) return { text: 'TRENDING', class: 'badge-trending' };
		return { text: 'RISING', class: 'badge-rising' };
	}
	
	function formatViews(views: number): string {
		if (views >= 1000000) {
			return `${(views / 1000000).toFixed(1)}M`;
		}
		if (views >= 1000) {
			return `${(views / 1000).toFixed(1)}K`;
		}
		return views.toString();
	}
	
	function formatTimeSince(date: Date): string {
		const seconds = Math.floor((new Date().getTime() - date.getTime()) / 1000);
		if (seconds < 60) return 'just now';
		const minutes = Math.floor(seconds / 60);
		if (minutes < 60) return `${minutes}m ago`;
		const hours = Math.floor(minutes / 60);
		return `${hours}h ago`;
	}
	
	function formatDuration(seconds: number): string {
		const hours = Math.floor(seconds / 3600);
		const minutes = Math.floor((seconds % 3600) / 60);
		if (hours > 0) {
			return `${hours}h ${minutes}m`;
		}
		return `${minutes}m`;
	}
</script>

{#if showTitle}
	<div class="section-header">
		<div class="title-row">
			<div class="title-group">
				<h2>
					<span class="fire-icon">{viewMode === 'trending' ? '🔥' : '📊'}</span>
					{viewMode === 'trending' ? 'Trending Now' : 'Most Watched'}
				</h2>
				<p class="subtitle">
					{viewMode === 'trending' ? 'Most watched in the last 7 days' : getTimePeriodLabel()}
				</p>
			</div>
			{#if lastUpdated}
				<div class="last-updated">
					Updated {formatTimeSince(lastUpdated)}
				</div>
			{/if}
		</div>
		
		<!-- View Mode Toggle -->
		<div class="view-toggle">
			<button 
				class="toggle-btn {viewMode === 'trending' ? 'active' : ''}"
				on:click={() => switchViewMode('trending')}
			>
				🔥 Trending
			</button>
			<button 
				class="toggle-btn {viewMode === 'most_watched' ? 'active' : ''}"
				on:click={() => switchViewMode('most_watched')}
			>
				📊 Most Watched
			</button>
		</div>
		
		<!-- Time Period Selector (only for Most Watched) -->
		{#if viewMode === 'most_watched'}
			<div class="time-period-selector">
				<!--<button 
					class="period-btn {timePeriod === 'week' ? 'active' : ''}"
					on:click={() => switchTimePeriod('week')}
				>
					This Week
				</button>
				<button 
					class="period-btn {timePeriod === 'month' ? 'active' : ''}"
					on:click={() => switchTimePeriod('month')}
				>
					This Month
				</button>-->
				<button 
					class="period-btn {timePeriod === 'all-time' ? 'active' : ''}"
					on:click={() => switchTimePeriod('all-time')}
				>
					All-Time
				</button>
			</div>
		{/if}
	</div>
{/if}

<div class="trending-container">
	{#if isLoading}
		<div class="loading">
			<div class="spinner"></div>
			<p>Finding trending videos...</p>
		</div>
	{:else if error}
		<div class="error">
			<p>{error}</p>
			<button on:click={loadVideos} class="retry-btn">Retry</button>
		</div>
	{:else if videos.length === 0}
		<div class="empty-state">
			<div class="empty-icon">📺</div>
			<h3>No trending videos yet</h3>
			<p>Start watching to see what's hot!</p>
		</div>
	{:else}
		<div class="video-grid">
			{#each videos as video, index (video.video_id)}
				<div 
					class="video-card" 
					on:click={() => handleVideoClick(video.video_id)}
					on:keypress={(e) => e.key === 'Enter' && handleVideoClick(video.video_id)}
					role="button"
					tabindex="0"
					style="animation-delay: {index * 0.05}s"
				>
					<div class="rank-badge">#{index + 1}</div>
					
					<div class="thumbnail-container">
						<img 
							src={video.thumbnail_url || '/placeholder-video.png'} 
							alt={video.title}
							class="thumbnail"
						/>
						{#if viewMode === 'trending' && video.trending_score !== undefined}
							<div class="trending-overlay">
								<div class="trending-badge {getTrendingBadge(video.trending_score).class}">
									{getTrendingBadge(video.trending_score).text}
								</div>
							</div>
						{/if}
						{#if video.duration}
							<div class="duration-badge">
								{formatDuration(video.duration)}
							</div>
						{/if}
						<div class="views-badge">
							<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
								<circle cx="12" cy="12" r="3"></circle>
							</svg>
							{#if viewMode === 'trending' && video.last_24h_views !== undefined}
								{formatViews(video.last_24h_views)} (24h)
							{:else if video.total_views !== undefined}
								{formatViews(video.total_views)} total
							{/if}
						</div>
					</div>
					
					<div class="video-info">
						<h3 class="video-title">{video.title}</h3>
						<div class="video-meta">
							{#if viewMode === 'trending' && video.trending_score !== undefined && video.trending_score > 0}
								<div class="score-indicator">
									<!-- Trending score bar is disabled for now, we need to make sure it's a relevant metric-->
									<!--<div class="score-bar">
										<div 
											class="score-fill {getTrendingBadge(video.trending_score).class}"
											style="width: {Math.min(video.trending_score, 100)}%"
										></div>
									</div>
									<span class="score-text">{Math.round(video.trending_score)}/100</span>-->
								</div>
							{:else if viewMode === 'most_watched'}
								<div class="stats-row">
									{#if video.unique_viewers}
										<span class="stat-item">
											👤 {formatViews(video.unique_viewers)} viewers
										</span>
									{/if}
									{#if video.avg_completion !== undefined}
										<span class="stat-item">
											✓ {Math.round(video.avg_completion)}% completion
										</span>
									{/if}
								</div>
							{/if}
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
	.section-header {
		margin-bottom: 1.5rem;
	}
	
	.title-row {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 1rem;
		margin-bottom: 1rem;
	}
	
	.view-toggle {
		display: flex;
		gap: 0.5rem;
		background: #f5f5f5;
		padding: 0.25rem;
		border-radius: 8px;
	}
	
	.toggle-btn {
		padding: 0.5rem 1rem;
		border: none;
		background: transparent;
		border-radius: 6px;
		font-size: 0.9rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s;
		color: #666;
	}
	
	.toggle-btn:hover {
		background: #e0e0e0;
		color: #333;
	}
	
	.toggle-btn.active {
		background: #fff;
		color: #ff6b35;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
	}
	
	.time-period-selector {
		display: flex;
		gap: 0.5rem;
		margin-top: 1rem;
		animation: slideDown 0.3s ease-out;
	}
	
	@keyframes slideDown {
		from {
			opacity: 0;
			transform: translateY(-10px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
	
	.period-btn {
		padding: 0.5rem 1rem;
		border: 2px solid #e0e0e0;
		background: #fff;
		border-radius: 6px;
		font-size: 0.9rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s;
		color: #666;
	}
	
	.period-btn:hover {
		border-color: #ff6b35;
		color: #ff6b35;
		background: #fff5f2;
	}
	
	.period-btn.active {
		background: #ff6b35;
		color: #fff;
		border-color: #ff6b35;
		box-shadow: 0 2px 8px rgba(255, 107, 53, 0.3);
	}
	
	.title-group {
		flex: 1;
	}
	
	.section-header h2 {
		font-size: 1.75rem;
		font-weight: 700;
		color: #1a1a1a;
		margin: 0 0 0.5rem 0;
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	
	.fire-icon {
		animation: pulse 2s ease-in-out infinite;
	}
	
	@keyframes pulse {
		0%, 100% { transform: scale(1); }
		50% { transform: scale(1.1); }
	}
	
	.subtitle {
		color: #666;
		margin: 0;
		font-size: 0.95rem;
	}
	
	.last-updated {
		font-size: 0.85rem;
		color: #999;
		padding: 0.5rem 1rem;
		background: #f5f5f5;
		border-radius: 20px;
		white-space: nowrap;
	}
	
	.trending-container {
		width: 100%;
	}
	
	.loading, .error, .empty-state {
		text-align: center;
		padding: 3rem 1rem;
	}
	
	.spinner {
		width: 40px;
		height: 40px;
		border: 3px solid #f0f0f0;
		border-top-color: #ff6b35;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
		margin: 0 auto 1rem;
	}
	
	@keyframes spin {
		to { transform: rotate(360deg); }
	}
	
	.retry-btn {
		margin-top: 1rem;
		padding: 0.5rem 1.5rem;
		background: #ff6b35;
		color: white;
		border: none;
		border-radius: 6px;
		cursor: pointer;
		font-weight: 600;
		transition: all 0.2s;
	}
	
	.retry-btn:hover {
		background: #ff5722;
		transform: translateY(-1px);
	}
	
	.empty-state {
		background: #f9f9f9;
		border-radius: 12px;
		padding: 3rem 2rem;
	}
	
	.empty-icon {
		font-size: 4rem;
		margin-bottom: 1rem;
	}
	
	.empty-state h3 {
		margin: 0 0 0.5rem 0;
		color: #333;
		font-size: 1.25rem;
	}
	
	.empty-state p {
		color: #666;
		margin: 0;
	}
	
	.video-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 1.5rem;
	}
	
	.video-card {
		background: #fff;
		border-radius: 12px;
		overflow: hidden;
		cursor: pointer;
		transition: all 0.3s ease;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
		position: relative;
		animation: slideUp 0.5s ease-out both;
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
	
	.video-card:hover {
		transform: translateY(-8px);
		box-shadow: 0 12px 32px rgba(255, 107, 53, 0.2);
	}
	
	.rank-badge {
		position: absolute;
		top: 0.75rem;
		left: 0.75rem;
		background: linear-gradient(135deg, #1a1a1a 0%, #333 100%);
		color: #fff;
		width: 36px;
		height: 36px;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 700;
		font-size: 0.9rem;
		z-index: 2;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
	}
	
	.thumbnail-container {
		position: relative;
		padding-top: 56.25%; /* 16:9 */
		background: #000;
		overflow: hidden;
	}
	
	.thumbnail {
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		object-fit: cover;
		transition: transform 0.3s ease;
	}
	
	.video-card:hover .thumbnail {
		transform: scale(1.05);
	}
	
	.trending-overlay {
		position: absolute;
		top: 0;
		right: 0;
		padding: 0.75rem;
	}
	
	.trending-badge {
		padding: 0.35rem 0.75rem;
		border-radius: 6px;
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.5px;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
		animation: glow 2s ease-in-out infinite;
	}
	
	.badge-fire {
		background: linear-gradient(135deg, #ff6b35 0%, #ff5722 100%);
		color: #fff;
	}
	
	.badge-hot {
		background: linear-gradient(135deg, #ff9800 0%, #ff6b35 100%);
		color: #fff;
	}
	
	.badge-trending {
		background: linear-gradient(135deg, #ffc107 0%, #ff9800 100%);
		color: #fff;
	}
	
	.badge-rising {
		background: linear-gradient(135deg, #4caf50 0%, #8bc34a 100%);
		color: #fff;
	}
	
	@keyframes glow {
		0%, 100% { box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3); }
		50% { box-shadow: 0 2px 16px rgba(255, 107, 53, 0.6); }
	}
	
	.views-badge {
		position: absolute;
		bottom: 0.75rem;
		right: 0.75rem;
		background: rgba(0, 0, 0, 0.8);
		color: #fff;
		padding: 0.35rem 0.75rem;
		border-radius: 20px;
		font-size: 0.8rem;
		font-weight: 600;
		display: flex;
		align-items: center;
		gap: 0.35rem;
	}
	
	.duration-badge {
		position: absolute;
		bottom: 0.75rem;
		left: 0.75rem;
		background: rgba(0, 0, 0, 0.8);
		color: #fff;
		padding: 0.35rem 0.75rem;
		border-radius: 20px;
		font-size: 0.8rem;
		font-weight: 600;
	}
	
	.video-info {
		padding: 1rem;
	}
	
	.video-title {
		margin: 0 0 0.75rem 0;
		font-size: 1rem;
		font-weight: 600;
		color: #1a1a1a;
		line-height: 1.4;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
		min-height: 2.8em;
	}
	
	.video-meta {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	
	.score-indicator {
		flex: 1;
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	
	.score-bar {
		flex: 1;
		height: 6px;
		background: #e0e0e0;
		border-radius: 3px;
		overflow: hidden;
	}
	
	.score-fill {
		height: 100%;
		transition: width 0.5s ease;
		border-radius: 3px;
	}
	
	.score-text {
		font-size: 0.85rem;
		font-weight: 600;
		color: #666;
		min-width: 45px;
		text-align: right;
	}
	
	.stats-row {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
		width: 100%;
	}
	
	.stat-item {
		font-size: 0.8rem;
		color: #666;
		font-weight: 500;
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}
	
	/* Top 3 special styling */
	.video-card:nth-child(1) .rank-badge {
		background: linear-gradient(135deg, #ffd700 0%, #ffed4e 100%);
		color: #1a1a1a;
		font-size: 1rem;
		animation: shine 2s ease-in-out infinite;
	}
	
	.video-card:nth-child(2) .rank-badge {
		background: linear-gradient(135deg, #c0c0c0 0%, #e8e8e8 100%);
		color: #1a1a1a;
	}
	
	.video-card:nth-child(3) .rank-badge {
		background: linear-gradient(135deg, #cd7f32 0%, #e8a87c 100%);
		color: #fff;
	}
	
	@keyframes shine {
		0%, 100% { box-shadow: 0 0 10px rgba(255, 215, 0, 0.5); }
		50% { box-shadow: 0 0 20px rgba(255, 215, 0, 0.8); }
	}
	
	/* Responsive */
	@media (max-width: 768px) {
		.video-grid {
			grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
			gap: 1rem;
		}
		
		.section-header h2 {
			font-size: 1.5rem;
		}
		
		.title-row {
			flex-direction: column;
			align-items: flex-start;
		}
		
		.last-updated {
			align-self: stretch;
			text-align: center;
		}
		
		.view-toggle {
			width: 100%;
		}
		
		.toggle-btn {
			flex: 1;
		}
		
		.time-period-selector {
			width: 100%;
			flex-wrap: wrap;
		}
		
		.period-btn {
			flex: 1;
			min-width: 90px;
		}
	}
	
	@media (max-width: 480px) {
		.video-grid {
			grid-template-columns: 1fr;
		}
	}
</style>

