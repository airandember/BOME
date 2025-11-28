<script lang="ts">
	import { onMount } from 'svelte';
	import { videoAnalytics } from '$lib/services/videoAnalytics';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	
	// Types
	interface VideoStats {
		video_id: number;
		title: string;
		thumbnail_url: string;
		total_views: number;
		unique_viewers: number;
		avg_completion: number;
		total_watch_time: number;
	}
	
	interface TrendingVideo {
		video_id: number;
		title: string;
		thumbnail_url: string;
		last_24h_views: number;
		trending_score: number;
	}
	
	interface SystemStats {
		total_views_today: number;
		total_views_week: number;
		total_watch_time_hours: number;
		avg_engagement_score: number;
		unique_viewers_today: number;
		videos_tracked: number;
	}
	
	// State
	let isLoading = $state(true);
	let error = $state<string | null>(null);
	let timeRange = $state<'24h' | '7d' | '30d' | '90d'>('7d');
	let topVideos = $state<VideoStats[]>([]);
	let trending = $state<TrendingVideo[]>([]);
	let systemStats = $state<SystemStats>({
		total_views_today: 0,
		total_views_week: 0,
		total_watch_time_hours: 0,
		avg_engagement_score: 0,
		unique_viewers_today: 0,
		videos_tracked: 0
	});
	
	// Refresh timer
	let refreshInterval: NodeJS.Timeout | null = null;
	let lastUpdated = $state<Date>(new Date());
	
	onMount(async () => {
		await loadAllData();
		
		// Auto-refresh every 5 minutes
		refreshInterval = setInterval(() => {
			loadAllData();
		}, 300000);
		
		return () => {
			if (refreshInterval) {
				clearInterval(refreshInterval);
			}
		};
	});
	
	async function loadAllData() {
		isLoading = true;
		error = null;
		
		try {
			await Promise.all([
				loadSystemStats(),
				loadTopVideos(),
				loadTrending()
			]);
			
			lastUpdated = new Date();
			console.log('✅ [Admin Analytics] All data loaded');
		} catch (err) {
			console.error('❌ [Admin Analytics] Failed to load data:', err);
			error = 'Failed to load analytics data';
		} finally {
			isLoading = false;
		}
	}

	async function loadSystemStats() {
		// Mock for now - you can implement real API endpoints
		const response = await fetch('/api/v1/analytics/top?limit=100&days=7');
			if (response.ok) {
			const data = await response.json();
			const videos = data.videos || [];
			
			systemStats = {
				total_views_today: videos.reduce((sum: number, v: any) => sum + v.total_views, 0),
				total_views_week: videos.reduce((sum: number, v: any) => sum + v.total_views, 0),
				total_watch_time_hours: videos.reduce((sum: number, v: any) => sum + v.total_watch_time, 0) / 3600,
				avg_engagement_score: videos.length > 0 ? videos.reduce((sum: number, v: any) => sum + (v.avg_completion || 0), 0) / videos.length : 0,
				unique_viewers_today: videos.reduce((sum: number, v: any) => sum + (v.unique_viewers || 0), 0),
				videos_tracked: videos.length
			};
		}
	}
	
	async function loadTopVideos() {
		const days = timeRange === '24h' ? 1 : timeRange === '7d' ? 7 : timeRange === '30d' ? 30 : 90;
		const response = await fetch(`/api/v1/analytics/top?limit=10&days=${days}`);
			
			if (response.ok) {
			const data = await response.json();
			topVideos = data.videos || [];
		}
	}
	
	async function loadTrending() {
		trending = await videoAnalytics.getTrendingVideos(10);
	}
	
	function handleTimeRangeChange(range: typeof timeRange) {
		timeRange = range;
		loadTopVideos();
	}
	
	function formatNumber(num: number): string {
		if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
		if (num >= 1000) return `${(num / 1000).toFixed(1)}K`;
		return num.toString();
	}
	
	function formatDuration(seconds: number): string {
		const hours = Math.floor(seconds / 3600);
		const mins = Math.floor((seconds % 3600) / 60);
		if (hours > 0) return `${hours}h ${mins}m`;
		return `${mins}m`;
	}
	
	function formatTimeSince(date: Date): string {
		const seconds = Math.floor((new Date().getTime() - date.getTime()) / 1000);
		if (seconds < 60) return 'just now';
		const minutes = Math.floor(seconds / 60);
		if (minutes < 60) return `${minutes}m ago`;
		const hours = Math.floor(minutes / 60);
		return `${hours}h ago`;
	}
	
	function getStartDate(): string {
		const date = new Date();
		date.setDate(date.getDate() - 30);
		return date.toISOString().split('T')[0];
	}
	
	function getTodayDate(): string {
		return new Date().toISOString().split('T')[0];
	}
	
	function getEngagementColor(score: number): string {
		if (score >= 75) return '#22c55e';
		if (score >= 50) return '#eab308';
		if (score >= 25) return '#f59e0b';
		return '#ef4444';
	}
	
	function getEngagementLabel(score: number): string {
		if (score >= 75) return 'Excellent';
		if (score >= 50) return 'Good';
		if (score >= 25) return 'Fair';
		return 'Poor';
	}
</script>

<svelte:head>
	<title>Video Analytics - Admin Dashboard</title>
</svelte:head>

<div class="analytics-dashboard">
	<!-- Header -->
	<div class="dashboard-header">
		<div>
			<h1>📊 Video Analytics Dashboard</h1>
			<p class="subtitle">Real-time insights into video performance and user engagement</p>
		</div>
		<div class="header-actions">
			<div class="last-updated">
				Last updated: {formatTimeSince(lastUpdated)}
			</div>
		<button class="btn-refresh" onclick={loadAllData} disabled={isLoading}>
			<span class="refresh-icon" class:spinning={isLoading}>🔄</span>
			Refresh
		</button>
		<a href="/api/v1/exports/video-analytics?start_date={getStartDate()}&end_date={getTodayDate()}" 
		   class="btn-export"
		   download>
			<span class="btn-icon">📥</span>
			Export CSV
		</a>
	</div>
	</div>
	
	{#if isLoading && topVideos.length === 0}
		<div class="loading-container">
			<LoadingSpinner />
			<p>Loading analytics data...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<p class="error-message">{error}</p>
			<button class="btn-primary" onclick={loadAllData}>Retry</button>
		</div>
	{:else}
		<!-- Key Metrics Cards -->
		<div class="metrics-grid">
			<div class="metric-card">
				<div class="metric-icon">👁️</div>
				<div class="metric-content">
					<div class="metric-value">{formatNumber(systemStats.total_views_week)}</div>
					<div class="metric-label">Total Views (7d)</div>
				</div>
			</div>
			
			<div class="metric-card">
				<div class="metric-icon">👥</div>
				<div class="metric-content">
					<div class="metric-value">{formatNumber(systemStats.unique_viewers_today)}</div>
					<div class="metric-label">Unique Viewers</div>
				</div>
			</div>
			
			<div class="metric-card">
				<div class="metric-icon">⏱️</div>
				<div class="metric-content">
					<div class="metric-value">{formatNumber(Math.round(systemStats.total_watch_time_hours))}h</div>
					<div class="metric-label">Watch Time</div>
				</div>
			</div>
			
			<div class="metric-card">
				<div class="metric-icon">📈</div>
				<div class="metric-content">
					<div class="metric-value">{Math.round(systemStats.avg_engagement_score)}%</div>
					<div class="metric-label">Avg Engagement</div>
					<div class="metric-badge" style="background: {getEngagementColor(systemStats.avg_engagement_score)}">
						{getEngagementLabel(systemStats.avg_engagement_score)}
					</div>
				</div>
			</div>
			
			<div class="metric-card">
				<div class="metric-icon">🎬</div>
				<div class="metric-content">
					<div class="metric-value">{systemStats.videos_tracked}</div>
					<div class="metric-label">Videos Tracked</div>
				</div>
			</div>
			
			<div class="metric-card highlight">
				<div class="metric-icon">🔥</div>
				<div class="metric-content">
					<div class="metric-value">{trending.length}</div>
					<div class="metric-label">Trending Now</div>
					<a href="/admin/streaming/analytics?section=trending" class="metric-link">View All →</a>
				</div>
			</div>
		</div>
		
		<!-- Main Content Grid -->
		<div class="content-grid">
			<!-- Top Performing Videos -->
			<div class="section-card">
				<div class="section-header">
					<h2>🏆 Top Performing Videos</h2>
					<div class="time-range-selector">
						<button 
							class="range-btn {timeRange === '24h' ? 'active' : ''}"
							onclick={() => handleTimeRangeChange('24h')}
						>24h</button>
						<button 
							class="range-btn {timeRange === '7d' ? 'active' : ''}"
							onclick={() => handleTimeRangeChange('7d')}
						>7d</button>
						<button 
							class="range-btn {timeRange === '30d' ? 'active' : ''}"
							onclick={() => handleTimeRangeChange('30d')}
						>30d</button>
						<button 
							class="range-btn {timeRange === '90d' ? 'active' : ''}"
							onclick={() => handleTimeRangeChange('90d')}
						>90d</button>
					</div>
				</div>
				
				{#if topVideos.length === 0}
					<div class="empty-state">
						<p>No video data available for this time period</p>
					</div>
				{:else}
					<div class="video-list">
						{#each topVideos as video, index}
							<div class="video-row">
								<div class="video-rank">#{index + 1}</div>
								<img src={video.thumbnail_url || '/placeholder-video.png'} alt={video.title} class="video-thumb" />
								<div class="video-details">
									<h3 class="video-title">{video.title}</h3>
									<div class="video-stats">
										<span class="stat">
											<span class="stat-icon">👁️</span>
											{formatNumber(video.total_views)} views
										</span>
										<span class="stat">
											<span class="stat-icon">👥</span>
											{formatNumber(video.unique_viewers)} unique
										</span>
										<span class="stat">
											<span class="stat-icon">⏱️</span>
											{formatDuration(video.total_watch_time)}
										</span>
										<span class="stat">
											<span class="stat-icon">📊</span>
											{Math.round(video.avg_completion)}% completion
										</span>
									</div>
								</div>
								<div class="video-engagement">
									<div class="engagement-chart">
										<svg width="60" height="60" viewBox="0 0 60 60">
											<circle cx="30" cy="30" r="25" fill="none" stroke="#e0e0e0" stroke-width="5" />
											<circle 
												cx="30" cy="30" r="25" 
												fill="none" 
												stroke={getEngagementColor(video.avg_completion)}
												stroke-width="5"
												stroke-dasharray="{(video.avg_completion / 100) * 157} 157"
												transform="rotate(-90 30 30)"
											/>
											<text x="30" y="35" text-anchor="middle" font-size="14" font-weight="bold" fill="#333">
												{Math.round(video.avg_completion)}%
											</text>
										</svg>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
			
			<!-- Trending Videos -->
			<div class="section-card">
				<div class="section-header">
					<h2>🔥 Trending Videos</h2>
					<span class="live-indicator">● Live</span>
				</div>
				
				{#if trending.length === 0}
					<div class="empty-state">
						<p>No trending videos at the moment</p>
					</div>
				{:else}
					<div class="trending-list">
						{#each trending as video, index}
							<div class="trending-row">
								<div class="trending-rank rank-{index + 1}">#{index + 1}</div>
								<img src={video.thumbnail_url || '/placeholder-video.png'} alt={video.title} class="trending-thumb" />
								<div class="trending-details">
									<h4 class="trending-title">{video.title}</h4>
									<div class="trending-meta">
										<span class="trending-views">
											👁️ {formatNumber(video.last_24h_views)} views (24h)
										</span>
										<div class="trending-score-bar">
											<div 
												class="trending-score-fill"
												style="width: {Math.min(video.trending_score, 100)}%"
											></div>
										</div>
										<span class="trending-score-text">
											{Math.round(video.trending_score)}/100
										</span>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>
		
		<!-- Quick Actions -->
		<div class="quick-actions">
			<h3>📋 Quick Actions</h3>
			<div class="action-buttons">
				<a href="/admin/streaming/videos" class="action-btn">
					<span class="action-icon">🎬</span>
					<span class="action-text">Manage Videos</span>
				</a>
				<a href="/admin/streaming/analytics/reports" class="action-btn">
					<span class="action-icon">📊</span>
					<span class="action-text">Detailed Reports</span>
				</a>
				<a href="/admin/streaming/analytics/export" class="action-btn">
					<span class="action-icon">📥</span>
					<span class="action-text">Export Data</span>
				</a>
				<button class="action-btn" onclick={() => window.print()}>
					<span class="action-icon">🖨️</span>
					<span class="action-text">Print Dashboard</span>
				</button>
			</div>
		</div>
	{/if}
</div>

<style>
	.analytics-dashboard {
		padding: 2rem;
		max-width: 1400px;
		margin: 0 auto;
	}
	
	.dashboard-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 2rem;
		gap: 2rem;
	}
	
	.dashboard-header h1 {
		font-size: 2rem;
		font-weight: 700;
		color: #1a1a1a;
		margin: 0 0 0.5rem 0;
	}
	
	.subtitle {
		color: #666;
		margin: 0;
		font-size: 0.95rem;
	}
	
	.header-actions {
		display: flex;
		align-items: center;
		gap: 1rem;
	}
	
	.last-updated {
		font-size: 0.85rem;
		color: #999;
		padding: 0.5rem 1rem;
		background: #f5f5f5;
		border-radius: 20px;
	}
	
	.btn-refresh {
		padding: 0.5rem 1rem;
		background: #d4a574;
		color: white;
		border: none;
		border-radius: 6px;
		cursor: pointer;
		font-weight: 600;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		transition: all 0.2s;
	}
	
	.btn-refresh:hover:not(:disabled) {
		background: #c89860;
		transform: translateY(-1px);
	}
	
	.btn-refresh:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}
	
	.btn-export {
		padding: 0.5rem 1rem;
		background: #22c55e;
		color: white;
		border: none;
		border-radius: 6px;
		font-weight: 600;
		cursor: pointer;
		text-decoration: none;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		transition: all 0.2s;
	}
	
	.btn-export:hover {
		background: #16a34a;
		transform: translateY(-1px);
	}
	
	.refresh-icon {
		display: inline-block;
		transition: transform 0.3s;
	}
	
	.refresh-icon.spinning {
		animation: spin 1s linear infinite;
	}
	
	@keyframes spin {
		to { transform: rotate(360deg); }
	}
	
	.metrics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 1.5rem;
		margin-bottom: 2rem;
	}
	
	.metric-card {
		background: white;
		border-radius: 12px;
		padding: 1.5rem;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
		display: flex;
		align-items: center;
		gap: 1rem;
		transition: all 0.3s;
	}
	
	.metric-card:hover {
		transform: translateY(-4px);
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
	}
	
	.metric-card.highlight {
		background: linear-gradient(135deg, #ff6b35 0%, #ff5722 100%);
		color: white;
	}
	
	.metric-card.highlight .metric-value,
	.metric-card.highlight .metric-label {
		color: white;
	}
	
	.metric-icon {
		font-size: 2rem;
		line-height: 1;
	}
	
	.metric-content {
		flex: 1;
	}
	
	.metric-value {
		font-size: 1.75rem;
		font-weight: 700;
		color: #1a1a1a;
		line-height: 1;
		margin-bottom: 0.25rem;
	}
	
	.metric-label {
		font-size: 0.85rem;
		color: #666;
		font-weight: 500;
	}
	
	.metric-badge {
		display: inline-block;
		padding: 0.25rem 0.75rem;
		border-radius: 12px;
		font-size: 0.75rem;
		font-weight: 600;
		color: white;
		margin-top: 0.5rem;
	}
	
	.metric-link {
		display: inline-block;
		margin-top: 0.5rem;
		color: white;
		text-decoration: none;
		font-size: 0.85rem;
		font-weight: 600;
	}
	
	.metric-link:hover {
		text-decoration: underline;
	}
	
	.content-grid {
		display: grid;
		grid-template-columns: 2fr 1fr;
		gap: 2rem;
		margin-bottom: 2rem;
	}
	
	.section-card {
		background: white;
		border-radius: 12px;
		padding: 1.5rem;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
	}
	
	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1.5rem;
	}
	
	.section-header h2 {
		font-size: 1.25rem;
		font-weight: 700;
		color: #1a1a1a;
		margin: 0;
	}
	
	.time-range-selector {
		display: flex;
		gap: 0.5rem;
	}
	
	.range-btn {
		padding: 0.35rem 0.75rem;
		background: #f5f5f5;
		border: 1px solid #e0e0e0;
		border-radius: 6px;
		cursor: pointer;
		font-size: 0.85rem;
		font-weight: 600;
		color: #666;
		transition: all 0.2s;
	}
	
	.range-btn:hover {
		background: #e0e0e0;
	}
	
	.range-btn.active {
		background: #d4a574;
		color: white;
		border-color: #d4a574;
	}
	
	.live-indicator {
		color: #ef4444;
		font-size: 0.85rem;
		font-weight: 600;
		animation: pulse 2s ease-in-out infinite;
	}
	
	@keyframes pulse {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.5; }
	}
	
	.video-list, .trending-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	
	.video-row {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1rem;
		border-radius: 8px;
		background: #f9f9f9;
		transition: all 0.2s;
	}
	
	.video-row:hover {
		background: #f0f0f0;
		transform: translateX(4px);
	}
	
	.video-rank {
		font-size: 1.25rem;
		font-weight: 700;
		color: #666;
		min-width: 40px;
		text-align: center;
	}
	
	.video-thumb {
		width: 120px;
		height: 68px;
		object-fit: cover;
		border-radius: 6px;
	}
	
	.video-details {
		flex: 1;
	}
	
	.video-title {
		font-size: 0.95rem;
		font-weight: 600;
		color: #1a1a1a;
		margin: 0 0 0.5rem 0;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	
	.video-stats {
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
		font-size: 0.85rem;
		color: #666;
	}
	
	.stat {
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}
	
	.stat-icon {
		font-size: 0.9rem;
	}
	
	.video-engagement {
		display: flex;
		align-items: center;
	}
	
	.trending-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem;
		border-radius: 8px;
		background: #f9f9f9;
		transition: all 0.2s;
	}
	
	.trending-row:hover {
		background: #f0f0f0;
	}
	
	.trending-rank {
		width: 32px;
		height: 32px;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 700;
		font-size: 0.85rem;
		background: #e0e0e0;
		color: #333;
	}
	
	.trending-rank.rank-1 {
		background: linear-gradient(135deg, #ffd700, #ffed4e);
		color: #1a1a1a;
	}
	
	.trending-rank.rank-2 {
		background: linear-gradient(135deg, #c0c0c0, #e8e8e8);
		color: #1a1a1a;
	}
	
	.trending-rank.rank-3 {
		background: linear-gradient(135deg, #cd7f32, #e8a87c);
		color: white;
	}
	
	.trending-thumb {
		width: 80px;
		height: 45px;
		object-fit: cover;
		border-radius: 4px;
	}
	
	.trending-details {
		flex: 1;
	}
	
	.trending-title {
		font-size: 0.9rem;
		font-weight: 600;
		color: #1a1a1a;
		margin: 0 0 0.5rem 0;
		display: -webkit-box;
		-webkit-line-clamp: 1;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	
	.trending-meta {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.8rem;
	}
	
	.trending-views {
		color: #666;
	}
	
	.trending-score-bar {
		flex: 1;
		height: 4px;
		background: #e0e0e0;
		border-radius: 2px;
		overflow: hidden;
	}
	
	.trending-score-fill {
		height: 100%;
		background: linear-gradient(90deg, #ff6b35, #ff5722);
		transition: width 0.5s ease;
	}
	
	.trending-score-text {
		color: #666;
		font-weight: 600;
		min-width: 45px;
		text-align: right;
	}
	
	.quick-actions {
		background: white;
		border-radius: 12px;
		padding: 1.5rem;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
	}
	
	.quick-actions h3 {
		font-size: 1.1rem;
		font-weight: 700;
		color: #1a1a1a;
		margin: 0 0 1rem 0;
	}
	
	.action-buttons {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
	}
	
	.action-btn {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 1rem;
		background: #f9f9f9;
		border: 1px solid #e0e0e0;
		border-radius: 8px;
		cursor: pointer;
		text-decoration: none;
		color: #333;
		transition: all 0.2s;
	}
	
	.action-btn:hover {
		background: #f0f0f0;
		border-color: #d4a574;
		transform: translateY(-2px);
	}
	
	.action-icon {
		font-size: 1.5rem;
	}
	
	.action-text {
		font-weight: 600;
		font-size: 0.9rem;
	}
	
	.loading-container,
	.error-container {
		text-align: center;
		padding: 4rem 2rem;
	}
	
	.error-message {
		color: #ef4444;
		margin-bottom: 1rem;
	}
	
	.empty-state {
		text-align: center;
		padding: 2rem;
		color: #999;
	}
	
	/* Responsive */
	@media (max-width: 1024px) {
		.content-grid {
			grid-template-columns: 1fr;
		}
	}
	
	@media (max-width: 768px) {
		.analytics-dashboard {
			padding: 1rem;
		}
		
		.dashboard-header {
			flex-direction: column;
			align-items: flex-start;
		}
		
		.metrics-grid {
			grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		}
		
		.video-row {
			flex-direction: column;
			align-items: flex-start;
		}
		
		.video-thumb {
			width: 100%;
			height: auto;
		}
	}
	
	/* Print styles */
	@media print {
		.btn-refresh,
		.quick-actions {
			display: none;
		}
	}
</style>

