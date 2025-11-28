<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/auth';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	
	// Types
	interface UserWatchStats {
		user_id: number;
		total_watch_time_minutes: number;
		total_watch_time_hours: number;
		total_videos_watched: number;
		videos_completed: number;
		completion_rate: number;
		current_streak: number;
		longest_streak: number;
		total_days_active: number;
		average_session_minutes: number;
		favorite_categories: CategoryStats[];
		recent_activity: DailyActivity[];
		achievements: Achievement[];
		member_since: string;
		last_watched_at?: string;
	}
	
	interface CategoryStats {
		category_id: number;
		category_name: string;
		videos_watched: number;
		watch_time_minutes: number;
		percentage: number;
	}
	
	interface DailyActivity {
		date: string;
		videos_watched: number;
		watch_time_minutes: number;
		videos_completed: number;
	}
	
	interface Achievement {
		id: string;
		name: string;
		description: string;
		icon: string;
		unlocked_at: string;
		progress: number;
		is_unlocked: boolean;
	}
	
	interface TopVideo {
		video_id: number;
		title: string;
		thumbnail_url: string;
		watch_count: number;
		total_watch_minutes: number;
		completion_rate: number;
		last_watched_at: string;
	}
	
	// State
	let isLoading = $state(true);
	let stats = $state<UserWatchStats | null>(null);
	let topVideos = $state<TopVideo[]>([]);
	let error = $state<string | null>(null);
	let selectedTab = $state<'overview' | 'achievements' | 'history'>('overview');
	
	onMount(() => {
		loadStats();
		loadTopVideos();
	});
	
	async function loadStats() {
		isLoading = true;
		error = null;
		
		try {
			const response = await fetch('/api/v1/user/stats');
			if (!response.ok) throw new Error('Failed to fetch statistics');
			
			stats = await response.json();
			console.log('✅ Loaded user stats:', stats);
		} catch (err) {
			console.error('❌ Failed to load stats:', err);
			error = 'Failed to load your watching statistics';
		} finally {
			isLoading = false;
		}
	}
	
	async function loadTopVideos() {
		try {
			const response = await fetch('/api/v1/user/stats/top-videos?limit=5');
			if (!response.ok) throw new Error('Failed to fetch top videos');
			
			const data = await response.json();
			topVideos = data.videos || [];
		} catch (err) {
			console.error('❌ Failed to load top videos:', err);
		}
	}
	
	function formatDuration(minutes: number): string {
		if (minutes < 60) return `${minutes}m`;
		const hours = Math.floor(minutes / 60);
		const mins = minutes % 60;
		return mins > 0 ? `${hours}h ${mins}m` : `${hours}h`;
	}
	
	function formatDate(dateString: string): string {
		const date = new Date(dateString);
		return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
	}
	
	function getStreakColor(streak: number): string {
		if (streak >= 30) return '#ff6b35';
		if (streak >= 7) return '#f59e0b';
		if (streak >= 3) return '#eab308';
		return '#666';
	}
	
	function getMembershipDuration(): string {
		if (!stats) return '';
		const start = new Date(stats.member_since);
		const now = new Date();
		const days = Math.floor((now.getTime() - start.getTime()) / (1000 * 60 * 60 * 24));
		
		if (days < 30) return `${days} days`;
		if (days < 365) return `${Math.floor(days / 30)} months`;
		const years = Math.floor(days / 365);
		const months = Math.floor((days % 365) / 30);
		return months > 0 ? `${years}y ${months}m` : `${years} years`;
	}
	
	const unlockedAchievements = $derived(stats?.achievements.filter(a => a.is_unlocked) || []);
	const lockedAchievements = $derived(stats?.achievements.filter(a => !a.is_unlocked) || []);
</script>

<svelte:head>
	<title>My Watch Statistics</title>
</svelte:head>

<div class="stats-page">
	<!-- Hero Header -->
	<div class="hero-header">
		<div class="hero-content">
			<h1>📊 Your Watch Statistics</h1>
			<p class="hero-subtitle">Track your viewing journey and achievements</p>
		</div>
		{#if stats}
			<div class="hero-badges">
				<div class="hero-badge">
					<div class="badge-icon">⏱️</div>
					<div class="badge-content">
						<div class="badge-value">{Math.round(stats.total_watch_time_hours)}h</div>
						<div class="badge-label">Watch Time</div>
					</div>
				</div>
				<div class="hero-badge">
					<div class="badge-icon">🎬</div>
					<div class="badge-content">
						<div class="badge-value">{stats.total_videos_watched}</div>
						<div class="badge-label">Videos</div>
					</div>
				</div>
				<div class="hero-badge streak" style="--streak-color: {getStreakColor(stats.current_streak)}">
					<div class="badge-icon">🔥</div>
					<div class="badge-content">
						<div class="badge-value">{stats.current_streak}</div>
						<div class="badge-label">Day Streak</div>
					</div>
				</div>
			</div>
		{/if}
	</div>
	
	<!-- Tabs -->
	<div class="tabs">
		<button 
			class="tab {selectedTab === 'overview' ? 'active' : ''}"
			onclick={() => selectedTab = 'overview'}
		>
			📈 Overview
		</button>
		<button 
			class="tab {selectedTab === 'achievements' ? 'active' : ''}"
			onclick={() => selectedTab = 'achievements'}
		>
			🏆 Achievements
		</button>
		<button 
			class="tab {selectedTab === 'history' ? 'active' : ''}"
			onclick={() => selectedTab = 'history'}
		>
			📜 History
		</button>
	</div>
	
	{#if isLoading}
		<div class="loading-container">
			<LoadingSpinner />
			<p>Loading your statistics...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<p class="error-message">{error}</p>
			<button class="btn-primary" onclick={loadStats}>Retry</button>
		</div>
	{:else if stats}
		<!-- Overview Tab -->
		{#if selectedTab === 'overview'}
			<div class="content-grid">
				<!-- Key Stats -->
				<div class="stats-grid">
					<div class="stat-card">
						<div class="stat-icon">✅</div>
						<div class="stat-content">
							<div class="stat-value">{stats.videos_completed}</div>
							<div class="stat-label">Completed</div>
							<div class="stat-subtext">{Math.round(stats.completion_rate)}% rate</div>
						</div>
					</div>
					
					<div class="stat-card">
						<div class="stat-icon">📅</div>
						<div class="stat-content">
							<div class="stat-value">{stats.total_days_active}</div>
							<div class="stat-label">Days Active</div>
							<div class="stat-subtext">Member for {getMembershipDuration()}</div>
						</div>
					</div>
					
					<div class="stat-card">
						<div class="stat-icon">⚡</div>
						<div class="stat-content">
							<div class="stat-value">{Math.round(stats.average_session_minutes)}m</div>
							<div class="stat-label">Avg Session</div>
							<div class="stat-subtext">Per viewing session</div>
						</div>
					</div>
					
					<div class="stat-card highlight">
						<div class="stat-icon">🎯</div>
						<div class="stat-content">
							<div class="stat-value">{stats.longest_streak}</div>
							<div class="stat-label">Longest Streak</div>
							<div class="stat-subtext">Days in a row</div>
						</div>
					</div>
				</div>
				
				<!-- Top Videos -->
				{#if topVideos.length > 0}
					<div class="section-card">
						<h2>🌟 Your Most Watched Videos</h2>
						<div class="top-videos-list">
							{#each topVideos as video, index}
								<div class="top-video-item">
									<div class="video-rank">#{index + 1}</div>
									<img src={video.thumbnail_url || '/placeholder.png'} alt={video.title} class="video-thumb" />
									<div class="video-info">
										<h4>{video.title}</h4>
										<div class="video-meta">
											<span>👁️ {video.watch_count} views</span>
											<span>⏱️ {formatDuration(video.total_watch_minutes)}</span>
											<span>✅ {Math.round(video.completion_rate)}% completed</span>
										</div>
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/if}
				
				<!-- Favorite Categories -->
				{#if stats.favorite_categories.length > 0}
					<div class="section-card">
						<h2>❤️ Your Favorite Categories</h2>
						<div class="categories-list">
							{#each stats.favorite_categories as category}
								<div class="category-item">
									<div class="category-header">
										<span class="category-name">{category.category_name}</span>
										<span class="category-percentage">{Math.round(category.percentage)}%</span>
									</div>
									<div class="category-bar-container">
										<div class="category-bar" style="width: {category.percentage}%"></div>
									</div>
									<div class="category-meta">
										{category.videos_watched} videos • {formatDuration(category.watch_time_minutes)}
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/if}
				
				<!-- Recent Activity Chart -->
				{#if stats.recent_activity.length > 0}
					<div class="section-card full-width">
						<h2>📊 Recent Activity (Last 30 Days)</h2>
						<div class="activity-chart">
							{#each stats.recent_activity.slice(0, 30).reverse() as activity}
								<div class="activity-bar-container">
									<div 
										class="activity-bar" 
										style="height: {Math.min((activity.watch_time_minutes / 120) * 100, 100)}%"
										title="{activity.date}: {activity.videos_watched} videos, {formatDuration(activity.watch_time_minutes)}"
									>
										<div class="activity-tooltip">
											<div class="tooltip-date">{formatDate(activity.date)}</div>
											<div class="tooltip-stat">📺 {activity.videos_watched} videos</div>
											<div class="tooltip-stat">⏱️ {formatDuration(activity.watch_time_minutes)}</div>
										</div>
									</div>
									<div class="activity-date">{new Date(activity.date).getDate()}</div>
								</div>
							{/each}
						</div>
					</div>
				{/if}
			</div>
		{/if}
		
		<!-- Achievements Tab -->
		{#if selectedTab === 'achievements'}
			<div class="achievements-container">
				<!-- Unlocked Achievements -->
				{#if unlockedAchievements.length > 0}
					<div class="achievements-section">
						<h2>🏆 Unlocked Achievements ({unlockedAchievements.length})</h2>
						<div class="achievements-grid">
							{#each unlockedAchievements as achievement}
								<div class="achievement-card unlocked">
									<div class="achievement-icon">{achievement.icon}</div>
									<div class="achievement-content">
										<h3>{achievement.name}</h3>
										<p>{achievement.description}</p>
										<div class="achievement-unlocked">
											✅ Unlocked {formatDate(achievement.unlocked_at)}
										</div>
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/if}
				
				<!-- Locked Achievements -->
				{#if lockedAchievements.length > 0}
					<div class="achievements-section">
						<h2>🔒 Locked Achievements ({lockedAchievements.length})</h2>
						<div class="achievements-grid">
							{#each lockedAchievements as achievement}
								<div class="achievement-card locked">
									<div class="achievement-icon grayscale">{achievement.icon}</div>
									<div class="achievement-content">
										<h3>{achievement.name}</h3>
										<p>{achievement.description}</p>
										<div class="achievement-progress">
											<div class="progress-bar">
												<div class="progress-fill" style="width: {achievement.progress}%"></div>
											</div>
											<span class="progress-text">{Math.round(achievement.progress)}%</span>
										</div>
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/if}
			</div>
		{/if}
		
		<!-- History Tab -->
		{#if selectedTab === 'history'}
			<div class="history-container">
				<div class="section-card">
					<h2>📜 Your Viewing History</h2>
					<p class="section-description">
						Detailed history and statistics will be displayed here
					</p>
					<!-- This can be expanded with ContinueWatching component -->
					<div class="history-placeholder">
						<div class="placeholder-icon">🚧</div>
						<p>Coming soon: Detailed viewing history and recommendations</p>
					</div>
				</div>
			</div>
		{/if}
	{/if}
</div>

<style>
	.stats-page {
		min-height: 100vh;
		background: linear-gradient(to bottom, #f5f7fa 0%, #ffffff 500px);
	}
	
	/* Hero Header */
	.hero-header {
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
		color: white;
		padding: 3rem 2rem;
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: 2rem;
	}
	
	.hero-content h1 {
		font-size: 2.5rem;
		font-weight: 700;
		margin: 0 0 0.5rem 0;
	}
	
	.hero-subtitle {
		font-size: 1.1rem;
		opacity: 0.9;
		margin: 0;
	}
	
	.hero-badges {
		display: flex;
		gap: 1.5rem;
		flex-wrap: wrap;
	}
	
	.hero-badge {
		background: rgba(255, 255, 255, 0.15);
		backdrop-filter: blur(10px);
		border-radius: 16px;
		padding: 1.5rem;
		display: flex;
		align-items: center;
		gap: 1rem;
		min-width: 150px;
		transition: all 0.3s;
	}
	
	.hero-badge:hover {
		background: rgba(255, 255, 255, 0.25);
		transform: translateY(-4px);
	}
	
	.hero-badge.streak {
		background: linear-gradient(135deg, var(--streak-color), rgba(255, 255, 255, 0.2));
		box-shadow: 0 0 20px rgba(255, 107, 53, 0.5);
	}
	
	.badge-icon {
		font-size: 2.5rem;
		line-height: 1;
	}
	
	.badge-value {
		font-size: 2rem;
		font-weight: 700;
		line-height: 1;
		margin-bottom: 0.25rem;
	}
	
	.badge-label {
		font-size: 0.9rem;
		opacity: 0.9;
	}
	
	/* Tabs */
	.tabs {
		display: flex;
		gap: 0.5rem;
		padding: 1rem 2rem 0;
		border-bottom: 2px solid #e0e0e0;
	}
	
	.tab {
		padding: 0.75rem 1.5rem;
		background: none;
		border: none;
		border-bottom: 3px solid transparent;
		cursor: pointer;
		font-weight: 600;
		color: #666;
		transition: all 0.2s;
		margin-bottom: -2px;
	}
	
	.tab:hover {
		color: #333;
		background: #f9f9f9;
	}
	
	.tab.active {
		color: #667eea;
		border-bottom-color: #667eea;
	}
	
	/* Content */
	.content-grid {
		padding: 2rem;
		max-width: 1400px;
		margin: 0 auto;
		display: grid;
		gap: 2rem;
	}
	
	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1.5rem;
	}
	
	.stat-card {
		background: white;
		border-radius: 12px;
		padding: 1.5rem;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
		display: flex;
		align-items: center;
		gap: 1rem;
		transition: all 0.3s;
	}
	
	.stat-card:hover {
		transform: translateY(-4px);
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
	}
	
	.stat-card.highlight {
		background: linear-gradient(135deg, #ff6b35, #ff5722);
		color: white;
	}
	
	.stat-card.highlight .stat-value,
	.stat-card.highlight .stat-label,
	.stat-card.highlight .stat-subtext {
		color: white;
	}
	
	.stat-icon {
		font-size: 2.5rem;
	}
	
	.stat-value {
		font-size: 2rem;
		font-weight: 700;
		color: #1a1a1a;
		line-height: 1;
		margin-bottom: 0.25rem;
	}
	
	.stat-label {
		font-size: 0.9rem;
		color: #666;
		font-weight: 600;
		margin-bottom: 0.25rem;
	}
	
	.stat-subtext {
		font-size: 0.8rem;
		color: #999;
	}
	
	/* Section Card */
	.section-card {
		background: white;
		border-radius: 12px;
		padding: 2rem;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
	}
	
	.section-card.full-width {
		grid-column: 1 / -1;
	}
	
	.section-card h2 {
		font-size: 1.5rem;
		font-weight: 700;
		color: #1a1a1a;
		margin: 0 0 1.5rem 0;
	}
	
	.section-description {
		color: #666;
		margin: 0 0 1rem 0;
	}
	
	/* Top Videos */
	.top-videos-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	
	.top-video-item {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1rem;
		background: #f9f9f9;
		border-radius: 8px;
		transition: all 0.2s;
	}
	
	.top-video-item:hover {
		background: #f0f0f0;
	}
	
	.video-rank {
		font-size: 1.5rem;
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
	
	.video-info {
		flex: 1;
	}
	
	.video-info h4 {
		font-size: 1rem;
		font-weight: 600;
		color: #1a1a1a;
		margin: 0 0 0.5rem 0;
	}
	
	.video-meta {
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
		font-size: 0.85rem;
		color: #666;
	}
	
	/* Categories */
	.categories-list {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}
	
	.category-item {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
	
	.category-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}
	
	.category-name {
		font-weight: 600;
		color: #1a1a1a;
	}
	
	.category-percentage {
		font-weight: 700;
		color: #667eea;
	}
	
	.category-bar-container {
		height: 8px;
		background: #e0e0e0;
		border-radius: 4px;
		overflow: hidden;
	}
	
	.category-bar {
		height: 100%;
		background: linear-gradient(90deg, #667eea, #764ba2);
		transition: width 0.5s ease;
	}
	
	.category-meta {
		font-size: 0.85rem;
		color: #999;
	}
	
	/* Activity Chart */
	.activity-chart {
		display: flex;
		align-items: flex-end;
		gap: 4px;
		height: 200px;
		padding: 1rem 0;
	}
	
	.activity-bar-container {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.5rem;
		position: relative;
	}
	
	.activity-bar {
		width: 100%;
		background: linear-gradient(to top, #667eea, #764ba2);
		border-radius: 4px 4px 0 0;
		min-height: 4px;
		transition: all 0.3s;
		position: relative;
		cursor: pointer;
	}
	
	.activity-bar:hover {
		opacity: 0.8;
	}
	
	.activity-bar:hover .activity-tooltip {
		opacity: 1;
		transform: translateY(-10px);
	}
	
	.activity-tooltip {
		position: absolute;
		bottom: 100%;
		left: 50%;
		transform: translateX(-50%);
		background: #1a1a1a;
		color: white;
		padding: 0.75rem;
		border-radius: 6px;
		font-size: 0.85rem;
		white-space: nowrap;
		opacity: 0;
		pointer-events: none;
		transition: all 0.3s;
		z-index: 10;
	}
	
	.tooltip-date {
		font-weight: 700;
		margin-bottom: 0.25rem;
	}
	
	.tooltip-stat {
		font-size: 0.8rem;
		opacity: 0.9;
	}
	
	.activity-date {
		font-size: 0.75rem;
		color: #999;
	}
	
	/* Achievements */
	.achievements-container {
		padding: 2rem;
		max-width: 1400px;
		margin: 0 auto;
	}
	
	.achievements-section {
		margin-bottom: 3rem;
	}
	
	.achievements-section h2 {
		font-size: 1.75rem;
		font-weight: 700;
		color: #1a1a1a;
		margin: 0 0 1.5rem 0;
	}
	
	.achievements-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: 1.5rem;
	}
	
	.achievement-card {
		background: white;
		border-radius: 12px;
		padding: 1.5rem;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
		display: flex;
		align-items: center;
		gap: 1rem;
		transition: all 0.3s;
	}
	
	.achievement-card:hover {
		transform: translateY(-4px);
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
	}
	
	.achievement-card.unlocked {
		background: linear-gradient(135deg, #fff5e6, white);
		border: 2px solid #ffd700;
	}
	
	.achievement-card.locked {
		opacity: 0.6;
	}
	
	.achievement-icon {
		font-size: 3rem;
		line-height: 1;
	}
	
	.achievement-icon.grayscale {
		filter: grayscale(100%);
	}
	
	.achievement-content {
		flex: 1;
	}
	
	.achievement-content h3 {
		font-size: 1.1rem;
		font-weight: 700;
		color: #1a1a1a;
		margin: 0 0 0.5rem 0;
	}
	
	.achievement-content p {
		color: #666;
		margin: 0 0 0.75rem 0;
		font-size: 0.9rem;
	}
	
	.achievement-unlocked {
		color: #22c55e;
		font-weight: 600;
		font-size: 0.85rem;
	}
	
	.achievement-progress {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}
	
	.progress-bar {
		flex: 1;
		height: 8px;
		background: #e0e0e0;
		border-radius: 4px;
		overflow: hidden;
	}
	
	.progress-fill {
		height: 100%;
		background: linear-gradient(90deg, #667eea, #764ba2);
		transition: width 0.5s ease;
	}
	
	.progress-text {
		font-weight: 700;
		color: #667eea;
		font-size: 0.9rem;
		min-width: 45px;
		text-align: right;
	}
	
	/* History */
	.history-container {
		padding: 2rem;
		max-width: 1400px;
		margin: 0 auto;
	}
	
	.history-placeholder {
		text-align: center;
		padding: 4rem 2rem;
	}
	
	.placeholder-icon {
		font-size: 4rem;
		margin-bottom: 1rem;
		opacity: 0.3;
	}
	
	/* Loading/Error */
	.loading-container,
	.error-container {
		text-align: center;
		padding: 4rem 2rem;
	}
	
	.error-message {
		color: #ef4444;
		margin-bottom: 1rem;
	}
	
	.btn-primary {
		padding: 0.75rem 1.5rem;
		background: #667eea;
		color: white;
		border: none;
		border-radius: 8px;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s;
	}
	
	.btn-primary:hover {
		background: #5568d3;
	}
	
	/* Responsive */
	@media (max-width: 768px) {
		.hero-header {
			flex-direction: column;
			align-items: flex-start;
			padding: 2rem 1rem;
		}
		
		.hero-content h1 {
			font-size: 2rem;
		}
		
		.hero-badges {
			width: 100%;
			justify-content: space-between;
		}
		
		.hero-badge {
			flex: 1;
			min-width: 0;
			padding: 1rem;
		}
		
		.tabs {
			padding: 1rem;
			overflow-x: auto;
		}
		
		.content-grid,
		.achievements-container,
		.history-container {
			padding: 1rem;
		}
		
		.stats-grid {
			grid-template-columns: 1fr;
		}
		
		.achievements-grid {
			grid-template-columns: 1fr;
		}
		
		.video-thumb {
			width: 80px;
			height: 45px;
		}
	}
</style>

